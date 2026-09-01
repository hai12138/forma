/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	analystrepo "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businessrepo "github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/stretchr/testify/require"
)

// memAnalystRepo is an in-memory AnalystRepository for domain integration tests.
type memAnalystRepo struct {
	mu                   sync.Mutex
	txnMu                sync.Mutex
	sessions             map[string]*entity.AnalystSession
	turns                map[string]*entity.AnalystTurn
	evidence             map[string]*entity.BusinessEvidence
	assertions           map[string]*entity.BusinessAssertion
	confirmations        map[string]*entity.BusinessConfirmation
	conflicts            map[string]*entity.AssertionConflict
	gaps                 map[string]*entity.AnalystGap
	proposals            map[string]*entity.BusinessModelProposal
	provenance           map[string]*entity.RevisionProvenance
	modelCalls           []*entity.ModelCallRecord
	evRefs               map[string]map[string]bool // assertionID -> evidenceIDs
	failOn               string
	failAfterAssertion   int
	createAssertionCalls int
}

func newMemAnalystRepo() *memAnalystRepo {
	return &memAnalystRepo{
		sessions:      map[string]*entity.AnalystSession{},
		turns:         map[string]*entity.AnalystTurn{},
		evidence:      map[string]*entity.BusinessEvidence{},
		assertions:    map[string]*entity.BusinessAssertion{},
		confirmations: map[string]*entity.BusinessConfirmation{},
		conflicts:     map[string]*entity.AssertionConflict{},
		gaps:          map[string]*entity.AnalystGap{},
		proposals:     map[string]*entity.BusinessModelProposal{},
		provenance:    map[string]*entity.RevisionProvenance{},
		evRefs:        map[string]map[string]bool{},
	}
}

func (m *memAnalystRepo) cloneLocked() *memAnalystRepo {
	out := newMemAnalystRepo()
	out.failOn = m.failOn
	out.failAfterAssertion = m.failAfterAssertion
	out.createAssertionCalls = m.createAssertionCalls
	for k, v := range m.sessions {
		cp := *v
		out.sessions[k] = &cp
	}
	for k, v := range m.turns {
		cp := *v
		out.turns[k] = &cp
	}
	for k, v := range m.evidence {
		cp := *v
		out.evidence[k] = &cp
	}
	for k, v := range m.assertions {
		cp := *v
		out.assertions[k] = &cp
	}
	for k, v := range m.confirmations {
		cp := *v
		out.confirmations[k] = &cp
	}
	for k, v := range m.conflicts {
		cp := *v
		out.conflicts[k] = &cp
	}
	for k, v := range m.gaps {
		cp := *v
		out.gaps[k] = &cp
	}
	for k, v := range m.proposals {
		cp := *v
		out.proposals[k] = &cp
	}
	for k, v := range m.provenance {
		cp := *v
		out.provenance[k] = &cp
	}
	for k, v := range m.evRefs {
		out.evRefs[k] = map[string]bool{}
		for eid := range v {
			out.evRefs[k][eid] = true
		}
	}
	for _, c := range m.modelCalls {
		cp := *c
		out.modelCalls = append(out.modelCalls, &cp)
	}
	return out
}

func (m *memAnalystRepo) applyFrom(src *memAnalystRepo) {
	m.sessions = src.sessions
	m.turns = src.turns
	m.evidence = src.evidence
	m.assertions = src.assertions
	m.confirmations = src.confirmations
	m.conflicts = src.conflicts
	m.gaps = src.gaps
	m.proposals = src.proposals
	m.provenance = src.provenance
	m.evRefs = src.evRefs
	m.modelCalls = src.modelCalls
	m.failOn = src.failOn
	m.failAfterAssertion = src.failAfterAssertion
	m.createAssertionCalls = src.createAssertionCalls
}

func (m *memAnalystRepo) Transaction(_ context.Context, fn func(txRepo analystrepo.AnalystRepository) error) error {
	m.txnMu.Lock()
	defer m.txnMu.Unlock()
	m.mu.Lock()
	snap := m.cloneLocked()
	m.mu.Unlock()
	if err := fn(snap); err != nil {
		return err
	}
	m.mu.Lock()
	m.applyFrom(snap)
	m.mu.Unlock()
	return nil
}

func (m *memAnalystRepo) CreateSession(_ context.Context, s *entity.AnalystSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *s
	if cp.NextTurnSequence <= 0 {
		cp.NextTurnSequence = 1
	}
	m.sessions[s.SessionID] = &cp
	return nil
}
func (m *memAnalystRepo) GetSession(_ context.Context, tenantID, sessionID string) (*entity.AnalystSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[sessionID]
	if s == nil || s.TenantID != tenantID {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}
func (m *memAnalystRepo) ListSessions(_ context.Context, tenantID, businessID string) ([]*entity.AnalystSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.AnalystSession
	for _, s := range m.sessions {
		if s.TenantID == tenantID && s.BusinessID == businessID {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memAnalystRepo) UpdateSession(_ context.Context, s *entity.AnalystSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *s
	m.sessions[s.SessionID] = &cp
	return nil
}
func (m *memAnalystRepo) GetSessionForUpdate(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error) {
	return m.GetSession(ctx, tenantID, sessionID)
}
func (m *memAnalystRepo) CreateTurn(_ context.Context, t *entity.AnalystTurn) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.ClientRequestID != "" {
		for _, ex := range m.turns {
			if ex.TenantID == t.TenantID && ex.SessionID == t.SessionID && ex.ClientRequestID == t.ClientRequestID {
				return fmt.Errorf("duplicate entry 1062 client_request_id")
			}
		}
	}
	cp := *t
	m.turns[t.TurnID] = &cp
	return nil
}
func (m *memAnalystRepo) GetTurnByClientRequestID(_ context.Context, tenantID, sessionID, clientRequestID string) (*entity.AnalystTurn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.turns {
		if t.TenantID == tenantID && t.SessionID == sessionID && t.ClientRequestID == clientRequestID {
			cp := *t
			return &cp, nil
		}
	}
	return nil, nil
}
func (m *memAnalystRepo) GetTurn(_ context.Context, tenantID, turnID string) (*entity.AnalystTurn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.turns[turnID]
	if t == nil || t.TenantID != tenantID {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}
func (m *memAnalystRepo) GetTurnByReplyTo(_ context.Context, tenantID, sessionID, replyToTurnID string) (*entity.AnalystTurn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.turns {
		if t.TenantID == tenantID && t.SessionID == sessionID &&
			t.ReplyToTurnID == replyToTurnID && t.Speaker == entity.SpeakerAnalyst {
			cp := *t
			return &cp, nil
		}
	}
	return nil, nil
}
func (m *memAnalystRepo) ListTurns(_ context.Context, tenantID, sessionID string) ([]*entity.AnalystTurn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.AnalystTurn
	for _, t := range m.turns {
		if t.TenantID == tenantID && t.SessionID == sessionID {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memAnalystRepo) MaxTurnSequence(_ context.Context, tenantID, sessionID string) (int32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var max int32
	for _, t := range m.turns {
		if t.TenantID == tenantID && t.SessionID == sessionID && t.Sequence > max {
			max = t.Sequence
		}
	}
	return max, nil
}
func (m *memAnalystRepo) UpdateTurnAnalysis(_ context.Context, tenantID, turnID string, status entity.AnalysisStatus, modelRequestID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.turns[turnID]
	if t == nil || t.TenantID != tenantID {
		return entity.ErrInvalidTurn
	}
	t.AnalysisStatus = status
	t.ModelRequestID = modelRequestID
	return nil
}
func (m *memAnalystRepo) ClaimTurnForRetry(_ context.Context, tenantID, turnID string, expectedStatuses []entity.AnalysisStatus, claimToken string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if claimToken == "" {
		return false, nil
	}
	t := m.turns[turnID]
	if t == nil || t.TenantID != tenantID {
		return false, entity.ErrInvalidTurn
	}
	statusOK := false
	for _, s := range expectedStatuses {
		if t.AnalysisStatus == s {
			statusOK = true
			break
		}
	}
	if !statusOK {
		return false, nil
	}
	if t.ModelRequestID != "" && t.ModelRequestID != claimToken && strings.HasPrefix(t.ModelRequestID, "retry_claim:") {
		return false, nil
	}
	t.ModelRequestID = claimToken
	return true, nil
}
func (m *memAnalystRepo) CreateEvidence(_ context.Context, e *entity.BusinessEvidence) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *e
	m.evidence[e.EvidenceID] = &cp
	return nil
}
func (m *memAnalystRepo) ListEvidence(_ context.Context, tenantID, businessID string) ([]*entity.BusinessEvidence, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.BusinessEvidence
	for _, e := range m.evidence {
		if e.TenantID == tenantID && e.BusinessID == businessID {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memAnalystRepo) GetEvidence(_ context.Context, tenantID, evidenceID string) (*entity.BusinessEvidence, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := m.evidence[evidenceID]
	if e == nil || e.TenantID != tenantID {
		return nil, nil
	}
	cp := *e
	return &cp, nil
}
func (m *memAnalystRepo) CreateAssertion(_ context.Context, a *entity.BusinessAssertion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createAssertionCalls++
	if m.failAfterAssertion > 0 && m.createAssertionCalls >= m.failAfterAssertion {
		return fmt.Errorf("injected CreateAssertion failure")
	}
	cp := *a
	m.assertions[a.AssertionID] = &cp
	return nil
}
func (m *memAnalystRepo) UpdateAssertion(_ context.Context, a *entity.BusinessAssertion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOn == "UpdateAssertion" {
		return fmt.Errorf("injected UpdateAssertion failure")
	}
	cp := *a
	m.assertions[a.AssertionID] = &cp
	return nil
}
func (m *memAnalystRepo) GetAssertion(_ context.Context, tenantID, assertionID string) (*entity.BusinessAssertion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.assertions[assertionID]
	if a == nil || a.TenantID != tenantID {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}
func (m *memAnalystRepo) ListAssertions(_ context.Context, tenantID, businessID string) ([]*entity.BusinessAssertion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.BusinessAssertion
	for _, a := range m.assertions {
		if a.TenantID == tenantID && a.BusinessID == businessID {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memAnalystRepo) CreateAssertionEvidenceRef(_ context.Context, tenantID, assertionID, evidenceID string, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.evRefs[assertionID] == nil {
		m.evRefs[assertionID] = map[string]bool{}
	}
	m.evRefs[assertionID][evidenceID] = true
	return nil
}
func (m *memAnalystRepo) ListEvidenceIDsForAssertion(_ context.Context, tenantID, assertionID string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.assertions[assertionID]
	if a == nil || a.TenantID != tenantID {
		return nil, nil
	}
	var ids []string
	for id := range m.evRefs[assertionID] {
		ids = append(ids, id)
	}
	return ids, nil
}
func (m *memAnalystRepo) ListAssertionIDsForEvidence(_ context.Context, tenantID, evidenceID string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var ids []string
	for aid, refs := range m.evRefs {
		if refs[evidenceID] {
			if a := m.assertions[aid]; a != nil && a.TenantID == tenantID {
				ids = append(ids, aid)
			}
		}
	}
	return ids, nil
}
func (m *memAnalystRepo) CreateConfirmation(_ context.Context, c *entity.BusinessConfirmation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOn == "CreateConfirmation" {
		return fmt.Errorf("injected CreateConfirmation failure")
	}
	cp := *c
	m.confirmations[c.ConfirmationID] = &cp
	return nil
}
func (m *memAnalystRepo) ListConfirmationsForAssertion(_ context.Context, tenantID, assertionID string) ([]*entity.BusinessConfirmation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.BusinessConfirmation
	for _, c := range m.confirmations {
		if c.TenantID == tenantID && c.AssertionID == assertionID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memAnalystRepo) CreateConflict(_ context.Context, c *entity.AssertionConflict) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *c
	m.conflicts[c.ConflictID] = &cp
	return nil
}
func (m *memAnalystRepo) GetConflictByPair(_ context.Context, tenantID, businessID, assertionIDA, assertionIDB string) (*entity.AssertionConflict, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idA, idB := canonicalConflictPair(assertionIDA, assertionIDB)
	for _, c := range m.conflicts {
		if c.TenantID == tenantID && c.BusinessID == businessID && c.AssertionIDA == idA && c.AssertionIDB == idB {
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}
func (m *memAnalystRepo) UpdateConflictStatus(_ context.Context, tenantID, conflictID string, status entity.ConflictStatus, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.conflicts[conflictID]
	if c == nil || c.TenantID != tenantID {
		return entity.ErrNotFound
	}
	c.Status = status
	return nil
}
func (m *memAnalystRepo) ListConflicts(_ context.Context, tenantID, businessID string) ([]*entity.AssertionConflict, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.AssertionConflict
	for _, c := range m.conflicts {
		if c.TenantID == tenantID && c.BusinessID == businessID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memAnalystRepo) CreateGap(_ context.Context, g *entity.AnalystGap) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *g
	m.gaps[g.GapID] = &cp
	return nil
}
func (m *memAnalystRepo) ListGaps(_ context.Context, tenantID, businessID string) ([]*entity.AnalystGap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.AnalystGap
	for _, g := range m.gaps {
		if g.TenantID == tenantID && g.BusinessID == businessID {
			cp := *g
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memAnalystRepo) UpdateGapStatus(_ context.Context, tenantID, gapID string, status entity.GapStatus, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	g := m.gaps[gapID]
	if g == nil || g.TenantID != tenantID {
		return entity.ErrNotFound
	}
	g.Status = status
	return nil
}
func (m *memAnalystRepo) CreateProposal(_ context.Context, p *entity.BusinessModelProposal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.proposals[p.ProposalID] = &cp
	return nil
}
func (m *memAnalystRepo) GetProposal(_ context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.proposals[proposalID]
	if p == nil || p.TenantID != tenantID {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}
func (m *memAnalystRepo) GetProposalForUpdate(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error) {
	return m.GetProposal(ctx, tenantID, proposalID)
}
func (m *memAnalystRepo) UpdateProposalStatus(_ context.Context, tenantID, proposalID string, status entity.ProposalStatus, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOn == "UpdateProposalStatus" {
		return fmt.Errorf("injected UpdateProposalStatus failure")
	}
	p := m.proposals[proposalID]
	if p == nil || p.TenantID != tenantID {
		return entity.ErrProposalNotFound
	}
	p.Status = status
	return nil
}
func (m *memAnalystRepo) MarkProposalStaleIfReady(_ context.Context, tenantID, proposalID string, at time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.proposals[proposalID]
	if p == nil || p.TenantID != tenantID {
		return false, entity.ErrProposalNotFound
	}
	if p.Status != entity.ProposalReadyForReview {
		return false, nil
	}
	p.Status = entity.ProposalStale
	p.UpdatedAt = at
	return true, nil
}
func (m *memAnalystRepo) CreateProvenance(_ context.Context, p *entity.RevisionProvenance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOn == "CreateProvenance" {
		return fmt.Errorf("injected CreateProvenance failure")
	}
	key := fmt.Sprintf("%s|%s|%d", p.TenantID, p.BusinessID, p.RevisionNo)
	cp := *p
	m.provenance[key] = &cp
	return nil
}
func (m *memAnalystRepo) GetProvenance(_ context.Context, tenantID, businessID string, revisionNo int32) (*entity.RevisionProvenance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s|%s|%d", tenantID, businessID, revisionNo)
	p := m.provenance[key]
	if p == nil {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}
func (m *memAnalystRepo) CreateModelCall(_ context.Context, r *entity.ModelCallRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *r
	m.modelCalls = append(m.modelCalls, &cp)
	return nil
}

var _ analystrepo.AnalystRepository = (*memAnalystRepo)(nil)

// testBusinessMem mirrors business/service memRepo for analyst integration tests.
type testBusinessMem struct {
	mu        sync.Mutex
	masters   map[string]*businessentity.BusinessModel
	revisions map[string]map[int32]*businessentity.BusinessModelRevision
	layouts   map[string]*businessentity.BusinessModelLayout
}

func newBusinessTestMem() *testBusinessMem {
	return &testBusinessMem{
		masters:   map[string]*businessentity.BusinessModel{},
		revisions: map[string]map[int32]*businessentity.BusinessModelRevision{},
		layouts:   map[string]*businessentity.BusinessModelLayout{},
	}
}

func bizKey(tenant, biz string) string { return tenant + "|" + biz }

func (m *testBusinessMem) CreateMaster(_ context.Context, model *businessentity.BusinessModel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *model
	m.masters[bizKey(model.TenantID, model.BusinessID)] = &cp
	return nil
}
func (m *testBusinessMem) GetMaster(_ context.Context, tenantID, businessID string) (*businessentity.BusinessModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.masters[bizKey(tenantID, businessID)]
	if v == nil {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}
func (m *testBusinessMem) ListMasters(_ context.Context, tenantID string) ([]*businessentity.BusinessModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*businessentity.BusinessModel
	for _, v := range m.masters {
		if v.TenantID == tenantID {
			cp := *v
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *testBusinessMem) CASBumpRevision(_ context.Context, tenantID, businessID string, expected, next int32) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.masters[bizKey(tenantID, businessID)]
	if v == nil || v.CurrentRevision != expected {
		return false, nil
	}
	v.CurrentRevision = next
	v.UpdatedAt = time.Now().UTC()
	return true, nil
}
func (m *testBusinessMem) TouchUpdatedAt(_ context.Context, tenantID, businessID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v := m.masters[bizKey(tenantID, businessID)]; v != nil {
		v.UpdatedAt = time.Now().UTC()
	}
	return nil
}
func (m *testBusinessMem) CreateRevision(_ context.Context, r *businessentity.BusinessModelRevision) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := bizKey(r.TenantID, r.BusinessID)
	if m.revisions[k] == nil {
		m.revisions[k] = map[int32]*businessentity.BusinessModelRevision{}
	}
	cp := *r
	m.revisions[k][r.RevisionNo] = &cp
	return nil
}
func (m *testBusinessMem) GetRevision(_ context.Context, tenantID, businessID string, rev int32) (*businessentity.BusinessModelRevision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.revisions[bizKey(tenantID, businessID)][rev]
	if v == nil {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}
func (m *testBusinessMem) ListRevisions(_ context.Context, tenantID, businessID string) ([]*businessentity.BusinessModelRevision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*businessentity.BusinessModelRevision
	for _, v := range m.revisions[bizKey(tenantID, businessID)] {
		cp := *v
		out = append(out, &cp)
	}
	return out, nil
}
func (m *testBusinessMem) UpsertLayout(_ context.Context, l *businessentity.BusinessModelLayout) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *l
	m.layouts[bizKey(l.TenantID, l.BusinessID)] = &cp
	return nil
}
func (m *testBusinessMem) GetLayout(_ context.Context, tenantID, businessID string) (*businessentity.BusinessModelLayout, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.layouts[bizKey(tenantID, businessID)]
	if v == nil {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}
func (m *testBusinessMem) CASBumpLayout(_ context.Context, tenantID, businessID string, expected, next int32, basedOn int32, layoutJSON, updatedBy string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.layouts[bizKey(tenantID, businessID)]
	if v == nil || v.LayoutRevision != expected {
		return false, nil
	}
	v.LayoutRevision = next
	v.BasedOnModelRevision = basedOn
	v.LayoutJSON = layoutJSON
	v.UpdatedBy = updatedBy
	v.UpdatedAt = time.Now().UTC()
	return true, nil
}
func (m *testBusinessMem) Transaction(ctx context.Context, fn func(txRepo businessrepo.BusinessRepository) error) error {
	return fn(m)
}

var _ businessrepo.BusinessRepository = (*testBusinessMem)(nil)

func newTestAnalystSvc(ar *memAnalystRepo, br *testBusinessMem) AnalystService {
	biz := businesssvc.NewBusinessService(&businesssvc.Components{Repo: br})
	return NewAnalystService(&Components{
		Repo:         ar,
		BusinessSVC:  biz,
		BusinessRepo: br,
		Model:        NewDeterministicFakeModel(),
	})
}

func TestProposalDigestDeterministic(t *testing.T) {
	patch := &entity.SemanticModelPatch{
		Operations: []entity.PatchOperation{{Op: entity.PatchAddNode}},
	}
	d1 := ProposalDigest(patch, 3, []string{"b", "a", "c"})
	d2 := ProposalDigest(patch, 3, []string{"a", "c", "b"})
	require.Equal(t, d1, d2)
}

func TestAIAssertionAlwaysProposedOnSubmit(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")

	sess, err := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)
	require.NoError(t, err)

	res, err := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "员工发现设备故障后提交报修，维修人员接单处理", "cr1", "p1")
	require.NoError(t, err)
	require.NotNil(t, res.Evidence)
	require.NotEmpty(t, res.Assertions)
	for _, a := range res.Assertions {
		require.Equal(t, entity.AssertionProposed, a.Status)
	}
	master, _ := br.GetMaster(ctx, "t1", "b1")
	require.Equal(t, int32(1), master.CurrentRevision)
}

func TestDuplicateClientRequestReturnsSameTurn(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)

	r1, err := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "员工报修", "dup_cr", "p1")
	require.NoError(t, err)
	r2, err := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "员工报修", "dup_cr", "p1")
	require.NoError(t, err)
	require.Equal(t, r1.UserTurn.TurnID, r2.UserTurn.TurnID)
}

func TestCrossSessionAssertionRejectedFromProposal(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	s1, _ := svc.CreateSession(ctx, "t1", "b1", "s1", "p1", entity.PolicyDevelopment)
	s2, _ := svc.CreateSession(ctx, "t1", "b1", "s2", "p1", entity.PolicyDevelopment)

	now := time.Now().UTC()
	aOther := &entity.BusinessAssertion{
		AssertionID:   "assert_other",
		TenantID:      "t1",
		BusinessID:    "b1",
		SessionID:     s2.SessionID,
		AssertionType: entity.AssertionActorExists,
		SubjectRef:    "actor:admin",
		Predicate:     "role",
		ObjectValue:   "管理员",
		Status:        entity.AssertionConfirmed,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, ar.CreateAssertion(ctx, aOther))
	require.NoError(t, ar.CreateAssertionEvidenceRef(ctx, "t1", aOther.AssertionID, "ev1", now))

	_, err := svc.CreateProposal(ctx, "t1", "b1", s1.SessionID, "p1", []string{aOther.AssertionID})
	require.Error(t, err)
}

func TestEditBeforeConfirmCreatesDerivedAssertion(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)
	res, _ := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "员工报修", "cr_edit", "p1")
	orig := res.Assertions[0]

	confirmed, err := svc.ConfirmAssertion(ctx, "t1", "b1", orig.AssertionID, "p1", "edit", &AssertionEdit{
		AssertionType: orig.AssertionType,
		SubjectRef:    orig.SubjectRef,
		Predicate:     "edited",
		ObjectValue:   "编辑后的员工",
	})
	require.NoError(t, err)
	require.NotEqual(t, orig.AssertionID, confirmed.AssertionID)
	require.Equal(t, entity.AssertionConfirmed, confirmed.Status)

	origReload, _ := ar.GetAssertion(ctx, "t1", orig.AssertionID)
	require.Equal(t, entity.AssertionSuperseded, origReload.Status)
	require.Equal(t, orig.ObjectValue, origReload.ObjectValue)
}

func TestConfirmationAtomicRollback(t *testing.T) {
	ar := newMemAnalystRepo()
	ar.failOn = "CreateConfirmation"
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)
	res, _ := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "员工报修", "cr_rb", "p1")
	a := res.Assertions[0]

	_, err := svc.ConfirmAssertion(ctx, "t1", "b1", a.AssertionID, "p1", "", nil)
	require.Error(t, err)
	reload, _ := ar.GetAssertion(ctx, "t1", a.AssertionID)
	require.Equal(t, entity.AssertionProposed, reload.Status)
}

func TestConflictDedup(t *testing.T) {
	ar := newMemAnalystRepo()
	svc := &analystServiceImpl{repo: ar}
	ctx := context.Background()
	now := time.Now().UTC()
	c1, err := svc.upsertConflict(ctx, ar, "t1", "b1", "s1", "a2", "a1", "rule:x", "perm", now)
	require.NoError(t, err)
	c2, err := svc.upsertConflict(ctx, ar, "t1", "b1", "s1", "a1", "a2", "rule:x", "perm", now)
	require.NoError(t, err)
	require.Equal(t, c1.ConflictID, c2.ConflictID)
}

func seedBusiness(ctx context.Context, br *testBusinessMem, tenantID, businessID string) {
	_ = br.CreateMaster(ctx, &businessentity.BusinessModel{
		TenantID:        tenantID,
		BusinessID:      businessID,
		CurrentRevision: 1,
		SchemaVersion:   businessentity.SemanticSchemaVersion,
	})
	_ = br.CreateRevision(ctx, &businessentity.BusinessModelRevision{
		TenantID:          tenantID,
		BusinessID:        businessID,
		RevisionNo:        1,
		BaseRevisionNo:    0,
		SchemaVersion:     businessentity.SemanticSchemaVersion,
		SemanticModelJSON: `{"schema_version":"1.0","nodes":[],"edges":[],"rules":[],"states":[]}`,
		ContentDigest:     "d",
		ChangeSummary:     "seed",
		CreatedBy:         "p1",
		CreatedAt:         time.Now().UTC(),
	})
}

type genFailFakeModel struct {
	DeterministicFakeModel
}

func (genFailFakeModel) GenerateInterviewTurn(_ context.Context, _ *InterviewTurnRequest) (*InterviewTurnResponse, error) {
	return nil, entity.ErrModelFailed
}

func TestConcurrentTurnSequencesUnique(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, err := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)
	require.NoError(t, err)

	const n = 10
	results := make([]*entity.TurnSubmissionResult, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			cr := fmt.Sprintf("cr_concurrent_%d", idx)
			results[idx], errs[idx] = svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "并发员工报修", cr, "p1")
		}(i)
	}
	wg.Wait()
	seen := map[int32]bool{}
	for i := 0; i < n; i++ {
		require.NoError(t, errs[i])
		require.NotNil(t, results[i].UserTurn)
		require.False(t, seen[results[i].UserTurn.Sequence], "duplicate sequence %d", results[i].UserTurn.Sequence)
		seen[results[i].UserTurn.Sequence] = true
		require.Greater(t, results[i].UserTurn.ReservedReplySequence, results[i].UserTurn.Sequence)
		if results[i].AnalystTurn != nil {
			require.Equal(t, results[i].UserTurn.ReservedReplySequence, results[i].AnalystTurn.Sequence)
			require.Equal(t, results[i].UserTurn.TurnID, results[i].AnalystTurn.ReplyToTurnID)
		}
	}
	require.Len(t, seen, n)
}

func TestConcurrentSameClientRequestIdempotent(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	model := &recordingAnalystModel{}
	svc := NewAnalystService(&Components{
		Repo:         ar,
		BusinessSVC:  businesssvc.NewBusinessService(&businesssvc.Components{Repo: br}),
		BusinessRepo: br,
		Model:        model,
	})
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)

	const n = 10
	results := make([]*entity.TurnSubmissionResult, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "相同请求", "same_cr", "p1")
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		require.NoError(t, errs[i])
		require.NotNil(t, results[i])
		require.NotNil(t, results[i].UserTurn)
	}
	first := results[0].UserTurn.TurnID
	for i := 0; i < n; i++ {
		require.Equal(t, first, results[i].UserTurn.TurnID)
	}

	turns, _ := ar.ListTurns(ctx, "t1", sess.SessionID)
	userTurns := 0
	for _, tr := range turns {
		if tr.Speaker == entity.SpeakerUser && tr.ClientRequestID == "same_cr" {
			userTurns++
		}
	}
	require.Equal(t, 1, userTurns)

	session, err := ar.GetSession(ctx, "t1", sess.SessionID)
	require.NoError(t, err)
	require.Equal(t, int32(3), session.NextTurnSequence)

	analystTurns := 0
	for _, tr := range turns {
		if tr.Speaker == entity.SpeakerAnalyst && tr.ReplyToTurnID == first {
			analystTurns++
		}
	}
	require.Equal(t, 1, analystTurns)

	evCount := 0
	for _, e := range ar.evidence {
		if e.BusinessID == "b1" && e.SessionID == sess.SessionID {
			evCount++
		}
	}
	require.Equal(t, 1, evCount)

	model.mu.Lock()
	extractCalls := len(model.extractCalls)
	model.mu.Unlock()
	require.Equal(t, 1, extractCalls, "duplicate client_request_id must not re-run extraction")
}

func TestResponseFailedRetryNoDuplicateAssertions(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := NewAnalystService(&Components{
		Repo:         ar,
		BusinessSVC:  businesssvc.NewBusinessService(&businesssvc.Components{Repo: br}),
		BusinessRepo: br,
		Model:        &genFailFakeModel{},
	})
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)

	res, err := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "员工发现设备故障后提交报修，维修人员接单处理", "cr_genfail", "p1")
	require.NoError(t, err)
	require.True(t, res.ModelFailed)
	require.Equal(t, entity.AnalysisResponseFailed, res.UserTurn.AnalysisStatus)
	require.NotEmpty(t, res.Assertions)

	beforeCount := len(res.Assertions)
	evCount := 0
	for _, e := range ar.evidence {
		if e.BusinessID == "b1" {
			evCount++
		}
	}
	require.Equal(t, 1, evCount)

	// Retry with working fake for generation path only
	svcImpl := svc.(*analystServiceImpl)
	svcImpl.model = NewDeterministicFakeModel()
	retryRes, err := svc.RetryTurnAnalysis(ctx, "t1", "b1", sess.SessionID, res.UserTurn.TurnID, "p1")
	require.NoError(t, err)
	require.False(t, retryRes.ModelFailed)
	require.NotNil(t, retryRes.AnalystTurn)
	require.Equal(t, res.UserTurn.TurnID, retryRes.AnalystTurn.ReplyToTurnID)

	allAssertions, _ := ar.ListAssertions(ctx, "t1", "b1")
	require.Equal(t, beforeCount, len(allAssertions))
}

func TestProposalStalePersisted(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	biz := businesssvc.NewBusinessService(&businesssvc.Components{Repo: br})
	svc := NewAnalystService(&Components{
		Repo:         ar,
		BusinessSVC:  biz,
		BusinessRepo: br,
		Model:        NewDeterministicFakeModel(),
	})
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	// bump revision
	m, _ := br.GetMaster(ctx, "t1", "b1")
	m.CurrentRevision = 2
	_ = br.CreateRevision(ctx, &businessentity.BusinessModelRevision{
		TenantID:          "t1",
		BusinessID:        "b1",
		RevisionNo:        2,
		BaseRevisionNo:    1,
		SchemaVersion:     businessentity.SemanticSchemaVersion,
		SemanticModelJSON: `{"schema_version":"1.0","nodes":[],"edges":[],"rules":[],"states":[]}`,
		ContentDigest:     "d2",
		ChangeSummary:     "bump",
		CreatedBy:         "p1",
		CreatedAt:         time.Now().UTC(),
	})
	_ = br.CreateMaster(ctx, m)

	now := time.Now().UTC()
	prop := &entity.BusinessModelProposal{
		ProposalID:    "prop_stale",
		TenantID:      "t1",
		BusinessID:    "b1",
		SessionID:     "s1",
		BaseRevision:  1,
		AssertionIDs:  []string{"a1"},
		Patch:         &entity.SemanticModelPatch{},
		Status:        entity.ProposalReadyForReview,
		ContentDigest: "dig",
		CreatedBy:     "p1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, ar.CreateProposal(ctx, prop))

	_, err := svc.ApplyProposal(ctx, "t1", "b1", "prop_stale", "p1")
	require.ErrorIs(t, err, entity.ErrProposalStale)
	updated, _ := ar.GetProposal(ctx, "t1", "prop_stale")
	require.Equal(t, entity.ProposalStale, updated.Status)
}

func TestProposalStaleDetectedInTransaction(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	svcImpl := svc.(*analystServiceImpl)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")

	m, _ := br.GetMaster(ctx, "t1", "b1")
	m.CurrentRevision = 2
	_ = br.CreateRevision(ctx, &businessentity.BusinessModelRevision{
		TenantID:          "t1",
		BusinessID:        "b1",
		RevisionNo:        2,
		BaseRevisionNo:    1,
		SchemaVersion:     businessentity.SemanticSchemaVersion,
		SemanticModelJSON: `{"schema_version":"1.0","nodes":[],"edges":[],"rules":[],"states":[]}`,
		ContentDigest:     "d2",
		ChangeSummary:     "bump",
		CreatedBy:         "p1",
		CreatedAt:         time.Now().UTC(),
	})
	_ = br.CreateMaster(ctx, m)

	now := time.Now().UTC()
	prop := &entity.BusinessModelProposal{
		ProposalID:    "prop_tx_stale",
		TenantID:      "t1",
		BusinessID:    "b1",
		SessionID:     "s1",
		BaseRevision:  1,
		AssertionIDs:  []string{"a1"},
		Patch:         &entity.SemanticModelPatch{},
		Status:        entity.ProposalReadyForReview,
		ContentDigest: "dig",
		CreatedBy:     "p1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, ar.CreateProposal(ctx, prop))

	_, err := svcImpl.applyProposalWithRepos(ctx, ar, br, "t1", "b1", "prop_tx_stale", "p1")
	require.ErrorIs(t, err, entity.ErrProposalStale)

	ok, err := ar.MarkProposalStaleIfReady(ctx, "t1", "prop_tx_stale", time.Now().UTC())
	require.NoError(t, err)
	require.True(t, ok)

	updated, _ := ar.GetProposal(ctx, "t1", "prop_tx_stale")
	require.Equal(t, entity.ProposalStale, updated.Status)

	master, _ := br.GetMaster(ctx, "t1", "b1")
	require.Equal(t, int32(2), master.CurrentRevision)
}

func TestMarkProposalStaleDoesNotOverwriteApplied(t *testing.T) {
	ar := newMemAnalystRepo()
	ctx := context.Background()
	now := time.Now().UTC()
	prop := &entity.BusinessModelProposal{
		ProposalID:    "prop_applied",
		TenantID:      "t1",
		BusinessID:    "b1",
		SessionID:     "s1",
		BaseRevision:  1,
		AssertionIDs:  []string{"a1"},
		Patch:         &entity.SemanticModelPatch{},
		Status:        entity.ProposalApplied,
		ContentDigest: "dig",
		CreatedBy:     "p1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, ar.CreateProposal(ctx, prop))

	ok, err := ar.MarkProposalStaleIfReady(ctx, "t1", "prop_applied", time.Now().UTC())
	require.NoError(t, err)
	require.False(t, ok)

	updated, _ := ar.GetProposal(ctx, "t1", "prop_applied")
	require.Equal(t, entity.ProposalApplied, updated.Status)
}

func TestExtractionPersistenceRollback(t *testing.T) {
	ar := newMemAnalystRepo()
	ar.failAfterAssertion = 2
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)

	content := "员工发现设备故障后提交报修，维修人员接单处理，完成后由管理员关闭。"
	res, err := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, content, "cr_extract_fail", "p1")
	require.NoError(t, err)
	require.Equal(t, entity.AnalysisExtractionFailed, res.UserTurn.AnalysisStatus)

	allAssertions, _ := ar.ListAssertions(ctx, "t1", "b1")
	require.Len(t, allAssertions, 0)
	gaps, _ := ar.ListGaps(ctx, "t1", "b1")
	require.Len(t, gaps, 0)
	conflicts, _ := ar.ListConflicts(ctx, "t1", "b1")
	require.Len(t, conflicts, 0)
	for _, refs := range ar.evRefs {
		require.Len(t, refs, 0)
	}

	evCount := 0
	for _, e := range ar.evidence {
		if e.BusinessID == "b1" {
			evCount++
		}
	}
	require.Equal(t, 1, evCount)

	ar.failAfterAssertion = 0
	ar.createAssertionCalls = 0
	retryRes, err := svc.RetryTurnAnalysis(ctx, "t1", "b1", sess.SessionID, res.UserTurn.TurnID, "p1")
	require.NoError(t, err)
	require.False(t, retryRes.ModelFailed)
	require.NotNil(t, retryRes.AnalystTurn)
	require.Greater(t, len(retryRes.Assertions), 0)

	allAssertions, _ = ar.ListAssertions(ctx, "t1", "b1")
	require.Greater(t, len(allAssertions), 0)
	require.Equal(t, 1, evCount)
}
