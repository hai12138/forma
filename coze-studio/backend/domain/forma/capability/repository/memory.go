/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

// NewMemoryCapabilityRepository returns a concurrency-safe in-memory repository.
func NewMemoryCapabilityRepository() CapabilityRepository {
	return &memRepo{
		caps:              make(map[string]*entity.BusinessCapability),
		revs:              make(map[string]*entity.BusinessCapabilityRevision),
		proposals:         make(map[string]*entity.CapabilityProposal),
		decisions:         make(map[string]*entity.CapabilityDecision),
		runs:              make(map[string]*entity.CapabilityAnalysisRun),
		runKeys:           make(map[string]string),
		deriveKeys:        make(map[string]string),
		proposalDecisions: make(map[string]string),
	}
}

type memRepo struct {
	mu                sync.Mutex
	caps              map[string]*entity.BusinessCapability
	revs              map[string]*entity.BusinessCapabilityRevision
	proposals         map[string]*entity.CapabilityProposal
	decisions         map[string]*entity.CapabilityDecision
	runs              map[string]*entity.CapabilityAnalysisRun
	runKeys           map[string]string
	deriveKeys        map[string]string
	proposalDecisions map[string]string
}

type memTx struct{ r *memRepo }

type memSnap struct {
	caps              map[string]*entity.BusinessCapability
	revs              map[string]*entity.BusinessCapabilityRevision
	proposals         map[string]*entity.CapabilityProposal
	decisions         map[string]*entity.CapabilityDecision
	runs              map[string]*entity.CapabilityAnalysisRun
	runKeys           map[string]string
	deriveKeys        map[string]string
	proposalDecisions map[string]string
}

func k(parts ...string) string {
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += "\x00" + parts[i]
	}
	return out
}

func (r *memRepo) snapshot() memSnap {
	return memSnap{
		caps: cloneCapMap(r.caps), revs: cloneRevMap(r.revs), proposals: clonePropMap(r.proposals),
		decisions: cloneDecMap(r.decisions), runs: cloneRunMap(r.runs),
		runKeys: cloneStrMap(r.runKeys), deriveKeys: cloneStrMap(r.deriveKeys),
		proposalDecisions: cloneStrMap(r.proposalDecisions),
	}
}

func (r *memRepo) restore(s memSnap) {
	r.caps, r.revs, r.proposals, r.decisions, r.runs = s.caps, s.revs, s.proposals, s.decisions, s.runs
	r.runKeys, r.deriveKeys, r.proposalDecisions = s.runKeys, s.deriveKeys, s.proposalDecisions
}

func (r *memRepo) Transaction(_ context.Context, fn func(txRepo CapabilityRepository) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	snap := r.snapshot()
	if err := fn(&memTx{r: r}); err != nil {
		r.restore(snap)
		return err
	}
	return nil
}
func (t *memTx) Transaction(ctx context.Context, fn func(txRepo CapabilityRepository) error) error {
	return fn(t)
}

func (r *memRepo) CreateCapability(ctx context.Context, cap *entity.BusinessCapability) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createCapability(ctx, cap)
}
func (t *memTx) CreateCapability(ctx context.Context, cap *entity.BusinessCapability) error {
	return t.r.createCapability(ctx, cap)
}
func (r *memRepo) createCapability(_ context.Context, cap *entity.BusinessCapability) error {
	id := k(cap.TenantID, cap.CapabilityID)
	if _, ok := r.caps[id]; ok {
		return entity.ErrConflict
	}
	cp := *cap
	// AggregateGeneration starts at 0 on create unless explicitly set.
	r.caps[id] = &cp
	return nil
}

func (r *memRepo) GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getCapability(ctx, tenantID, capabilityID)
}
func (t *memTx) GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	return t.r.getCapability(ctx, tenantID, capabilityID)
}
func (r *memRepo) getCapability(_ context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	c, ok := r.caps[k(tenantID, capabilityID)]
	if !ok {
		return nil, entity.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *memRepo) GetCapabilityForUpdate(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getCapability(ctx, tenantID, capabilityID)
}
func (t *memTx) GetCapabilityForUpdate(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	return t.r.getCapability(ctx, tenantID, capabilityID)
}

func (r *memRepo) UpdateActiveRevisionID(ctx context.Context, tenantID, capabilityID, activeRevisionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.updateActiveRevisionID(ctx, tenantID, capabilityID, activeRevisionID)
}
func (t *memTx) UpdateActiveRevisionID(ctx context.Context, tenantID, capabilityID, activeRevisionID string) error {
	return t.r.updateActiveRevisionID(ctx, tenantID, capabilityID, activeRevisionID)
}
func (r *memRepo) updateActiveRevisionID(_ context.Context, tenantID, capabilityID, activeRevisionID string) error {
	c, ok := r.caps[k(tenantID, capabilityID)]
	if !ok {
		return entity.ErrNotFound
	}
	c.ActiveRevisionID = activeRevisionID
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memRepo) CASBumpAggregateGeneration(ctx context.Context, tenantID, capabilityID string, expectedGen int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.casBumpAggregateGeneration(ctx, tenantID, capabilityID, expectedGen)
}
func (t *memTx) CASBumpAggregateGeneration(ctx context.Context, tenantID, capabilityID string, expectedGen int64) (bool, error) {
	return t.r.casBumpAggregateGeneration(ctx, tenantID, capabilityID, expectedGen)
}
func (r *memRepo) casBumpAggregateGeneration(_ context.Context, tenantID, capabilityID string, expectedGen int64) (bool, error) {
	c, ok := r.caps[k(tenantID, capabilityID)]
	if !ok {
		return false, entity.ErrNotFound
	}
	if c.AggregateGeneration != expectedGen {
		return false, nil
	}
	c.AggregateGeneration = expectedGen + 1
	c.UpdatedAt = time.Now().UTC()
	return true, nil
}

func (r *memRepo) CreateRevision(ctx context.Context, rev *entity.BusinessCapabilityRevision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createRevision(ctx, rev)
}
func (t *memTx) CreateRevision(ctx context.Context, rev *entity.BusinessCapabilityRevision) error {
	return t.r.createRevision(ctx, rev)
}
func (r *memRepo) createRevision(_ context.Context, rev *entity.BusinessCapabilityRevision) error {
	rk := k(rev.TenantID, rev.RevisionID)
	if _, ok := r.revs[rk]; ok {
		return entity.ErrConflict
	}
	for _, existing := range r.revs {
		if existing.TenantID == rev.TenantID && existing.CapabilityID == rev.CapabilityID && existing.Version == rev.Version {
			return entity.ErrConflict
		}
	}
	r.revs[rk] = cloneRevision(rev)
	return nil
}

func (r *memRepo) GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getRevision(ctx, tenantID, revisionID)
}
func (t *memTx) GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	return t.r.getRevision(ctx, tenantID, revisionID)
}
func (r *memRepo) getRevision(_ context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	rev, ok := r.revs[k(tenantID, revisionID)]
	if !ok {
		return nil, entity.ErrRevisionNotFound
	}
	return cloneRevision(rev), nil
}

func (r *memRepo) GetRevisionForUpdate(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getRevision(ctx, tenantID, revisionID)
}
func (t *memTx) GetRevisionForUpdate(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	return t.r.getRevision(ctx, tenantID, revisionID)
}

func (r *memRepo) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listRevisions(ctx, tenantID, capabilityID)
}
func (t *memTx) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	return t.r.listRevisions(ctx, tenantID, capabilityID)
}
func (r *memRepo) listRevisions(_ context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	out := make([]*entity.BusinessCapabilityRevision, 0)
	for _, rev := range r.revs {
		if rev.TenantID == tenantID && rev.CapabilityID == capabilityID {
			out = append(out, cloneRevision(rev))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

func (r *memRepo) UpdateRevisionStatus(ctx context.Context, tenantID, revisionID string, from, to entity.RevisionStatus) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.updateRevisionStatus(ctx, tenantID, revisionID, from, to)
}
func (t *memTx) UpdateRevisionStatus(ctx context.Context, tenantID, revisionID string, from, to entity.RevisionStatus) (bool, error) {
	return t.r.updateRevisionStatus(ctx, tenantID, revisionID, from, to)
}
func (r *memRepo) updateRevisionStatus(_ context.Context, tenantID, revisionID string, from, to entity.RevisionStatus) (bool, error) {
	rev, ok := r.revs[k(tenantID, revisionID)]
	if !ok || rev.Status != from {
		return false, nil
	}
	rev.Status = to
	return true, nil
}

func (r *memRepo) MaxVersion(ctx context.Context, tenantID, capabilityID string) (int32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.maxVersion(ctx, tenantID, capabilityID)
}
func (t *memTx) MaxVersion(ctx context.Context, tenantID, capabilityID string) (int32, error) {
	return t.r.maxVersion(ctx, tenantID, capabilityID)
}
func (r *memRepo) maxVersion(_ context.Context, tenantID, capabilityID string) (int32, error) {
	var max int32
	for _, rev := range r.revs {
		if rev.TenantID == tenantID && rev.CapabilityID == capabilityID && rev.Version > max {
			max = rev.Version
		}
	}
	return max, nil
}

func (r *memRepo) CreateProposal(ctx context.Context, p *entity.CapabilityProposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createProposal(ctx, p)
}
func (t *memTx) CreateProposal(ctx context.Context, p *entity.CapabilityProposal) error {
	return t.r.createProposal(ctx, p)
}
func (r *memRepo) createProposal(_ context.Context, p *entity.CapabilityProposal) error {
	id := k(p.TenantID, p.ProposalID)
	if _, ok := r.proposals[id]; ok {
		return entity.ErrConflict
	}
	r.proposals[id] = cloneProposal(p)
	return nil
}

func (r *memRepo) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getProposal(ctx, tenantID, proposalID)
}
func (t *memTx) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	return t.r.getProposal(ctx, tenantID, proposalID)
}
func (r *memRepo) getProposal(_ context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	p, ok := r.proposals[k(tenantID, proposalID)]
	if !ok {
		return nil, entity.ErrProposalNotFound
	}
	return cloneProposal(p), nil
}

func (r *memRepo) GetProposalForUpdate(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getProposal(ctx, tenantID, proposalID)
}
func (t *memTx) GetProposalForUpdate(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	return t.r.getProposal(ctx, tenantID, proposalID)
}

func (r *memRepo) ListProposalsByAnalysisRun(ctx context.Context, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listProposalsByAnalysisRun(ctx, tenantID, analysisRunID)
}
func (t *memTx) ListProposalsByAnalysisRun(ctx context.Context, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error) {
	return t.r.listProposalsByAnalysisRun(ctx, tenantID, analysisRunID)
}
func (r *memRepo) listProposalsByAnalysisRun(_ context.Context, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error) {
	out := make([]*entity.CapabilityProposal, 0)
	for _, p := range r.proposals {
		if p.TenantID == tenantID && p.AnalysisRunID == analysisRunID {
			out = append(out, cloneProposal(p))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *memRepo) TerminalizeProposal(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, materializedRevisionID, capabilityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.terminalizeProposal(ctx, tenantID, proposalID, status, materializedRevisionID, capabilityID)
}
func (t *memTx) TerminalizeProposal(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, materializedRevisionID, capabilityID string) error {
	return t.r.terminalizeProposal(ctx, tenantID, proposalID, status, materializedRevisionID, capabilityID)
}
func (r *memRepo) terminalizeProposal(_ context.Context, tenantID, proposalID string, status entity.ProposalStatus, materializedRevisionID, capabilityID string) error {
	p, ok := r.proposals[k(tenantID, proposalID)]
	if !ok || p.Status != entity.ProposalProposed {
		return entity.ErrConflict
	}
	p.Status = status
	if materializedRevisionID != "" {
		p.MaterializedRevisionID = materializedRevisionID
	}
	if capabilityID != "" {
		p.CapabilityID = capabilityID
	}
	return nil
}

func (r *memRepo) CreateDecision(ctx context.Context, d *entity.CapabilityDecision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createDecision(ctx, d)
}
func (t *memTx) CreateDecision(ctx context.Context, d *entity.CapabilityDecision) error {
	return t.r.createDecision(ctx, d)
}
func (r *memRepo) createDecision(_ context.Context, d *entity.CapabilityDecision) error {
	dk := k(d.TenantID, d.DecisionID)
	if _, ok := r.decisions[dk]; ok {
		return entity.ErrConflict
	}
	if d.ProposalID != "" {
		pk := k(d.TenantID, d.ProposalID)
		if _, ok := r.proposalDecisions[pk]; ok {
			return entity.ErrConflict
		}
		r.proposalDecisions[pk] = d.DecisionID
	}
	if d.ClientRequestID != "" && d.SourceRevisionID != "" && d.CapabilityID != "" {
		derive := k(d.TenantID, d.CapabilityID, d.SourceRevisionID, d.ClientRequestID)
		if _, ok := r.deriveKeys[derive]; ok {
			return entity.ErrConflict
		}
		r.deriveKeys[derive] = d.DecisionID
	}
	cp := *d
	r.decisions[dk] = &cp
	return nil
}

func (r *memRepo) GetDecision(ctx context.Context, tenantID, decisionID string) (*entity.CapabilityDecision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getDecision(ctx, tenantID, decisionID)
}
func (t *memTx) GetDecision(ctx context.Context, tenantID, decisionID string) (*entity.CapabilityDecision, error) {
	return t.r.getDecision(ctx, tenantID, decisionID)
}
func (r *memRepo) getDecision(_ context.Context, tenantID, decisionID string) (*entity.CapabilityDecision, error) {
	d, ok := r.decisions[k(tenantID, decisionID)]
	if !ok {
		return nil, entity.ErrDecisionNotFound
	}
	cp := *d
	return &cp, nil
}

func (r *memRepo) GetDecisionByProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityDecision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getDecisionByProposal(ctx, tenantID, proposalID)
}
func (t *memTx) GetDecisionByProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityDecision, error) {
	return t.r.getDecisionByProposal(ctx, tenantID, proposalID)
}
func (r *memRepo) getDecisionByProposal(_ context.Context, tenantID, proposalID string) (*entity.CapabilityDecision, error) {
	id, ok := r.proposalDecisions[k(tenantID, proposalID)]
	if !ok {
		return nil, entity.ErrDecisionNotFound
	}
	return r.getDecision(context.Background(), tenantID, id)
}

func (r *memRepo) GetDecisionByDeriveKey(ctx context.Context, tenantID, capabilityID, sourceRevisionID, clientRequestID string) (*entity.CapabilityDecision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getDecisionByDeriveKey(ctx, tenantID, capabilityID, sourceRevisionID, clientRequestID)
}
func (t *memTx) GetDecisionByDeriveKey(ctx context.Context, tenantID, capabilityID, sourceRevisionID, clientRequestID string) (*entity.CapabilityDecision, error) {
	return t.r.getDecisionByDeriveKey(ctx, tenantID, capabilityID, sourceRevisionID, clientRequestID)
}
func (r *memRepo) getDecisionByDeriveKey(_ context.Context, tenantID, capabilityID, sourceRevisionID, clientRequestID string) (*entity.CapabilityDecision, error) {
	id, ok := r.deriveKeys[k(tenantID, capabilityID, sourceRevisionID, clientRequestID)]
	if !ok {
		return nil, entity.ErrDecisionNotFound
	}
	return r.getDecision(context.Background(), tenantID, id)
}

func (r *memRepo) ListDecisionsByCapability(ctx context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listDecisionsByCapability(ctx, tenantID, capabilityID)
}
func (t *memTx) ListDecisionsByCapability(ctx context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error) {
	return t.r.listDecisionsByCapability(ctx, tenantID, capabilityID)
}
func (r *memRepo) listDecisionsByCapability(_ context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error) {
	out := make([]*entity.CapabilityDecision, 0)
	for _, d := range r.decisions {
		if d.TenantID == tenantID && d.CapabilityID == capabilityID {
			cp := *d
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memRepo) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.CapabilityAnalysisRun) (*entity.CapabilityAnalysisRun, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createOrClaimAnalysisRun(ctx, run)
}
func (t *memTx) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.CapabilityAnalysisRun) (*entity.CapabilityAnalysisRun, bool, error) {
	return t.r.createOrClaimAnalysisRun(ctx, run)
}
func (r *memRepo) createOrClaimAnalysisRun(_ context.Context, run *entity.CapabilityAnalysisRun) (*entity.CapabilityAnalysisRun, bool, error) {
	ik := k(run.TenantID, run.BusinessID, fmt.Sprintf("%d", run.BusinessModelRevision), run.ClientRequestID)
	if id, ok := r.runKeys[ik]; ok {
		existing := r.runs[k(run.TenantID, id)]
		if existing.RequestDigest != run.RequestDigest {
			return nil, false, entity.ErrIdempotencyConflict
		}
		return cloneAnalysisRun(existing), false, nil
	}
	cp := *run
	if cp.Attempt == 0 {
		cp.Attempt = 1
	}
	now := time.Now().UTC()
	if cp.ExecutionClaimedAt == nil {
		cp.ExecutionClaimedAt = &now
	} else {
		t := *cp.ExecutionClaimedAt
		cp.ExecutionClaimedAt = &t
	}
	if cp.LeaseExpiresAt == nil {
		exp := now.Add(5 * time.Minute)
		cp.LeaseExpiresAt = &exp
	} else {
		t := *cp.LeaseExpiresAt
		cp.LeaseExpiresAt = &t
	}
	stored := cp
	r.runs[k(run.TenantID, run.AnalysisRunID)] = &stored
	r.runKeys[ik] = run.AnalysisRunID
	return cloneAnalysisRun(&stored), true, nil
}

func (r *memRepo) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getAnalysisRun(ctx, tenantID, analysisRunID)
}
func (t *memTx) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	return t.r.getAnalysisRun(ctx, tenantID, analysisRunID)
}
func (r *memRepo) getAnalysisRun(_ context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	run, ok := r.runs[k(tenantID, analysisRunID)]
	if !ok {
		return nil, entity.ErrAnalysisNotFound
	}
	return cloneAnalysisRun(run), nil
}

func (r *memRepo) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string, expectedAttempt int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.markAnalysisSucceeded(ctx, tenantID, analysisRunID, modelRef, expectedAttempt)
}
func (t *memTx) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string, expectedAttempt int32) error {
	return t.r.markAnalysisSucceeded(ctx, tenantID, analysisRunID, modelRef, expectedAttempt)
}
func (r *memRepo) markAnalysisSucceeded(_ context.Context, tenantID, analysisRunID, modelRef string, expectedAttempt int32) error {
	run, ok := r.runs[k(tenantID, analysisRunID)]
	if !ok || run.Status != entity.AnalysisPending || run.Attempt != expectedAttempt {
		return entity.ErrStaleGeneration
	}
	run.Status = entity.AnalysisSucceeded
	run.ModelRef = modelRef
	run.ErrorCode = ""
	run.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memRepo) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorCode string, expectedAttempt int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.markAnalysisFailed(ctx, tenantID, analysisRunID, errorCode, expectedAttempt)
}
func (t *memTx) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorCode string, expectedAttempt int32) error {
	return t.r.markAnalysisFailed(ctx, tenantID, analysisRunID, errorCode, expectedAttempt)
}
func (r *memRepo) markAnalysisFailed(_ context.Context, tenantID, analysisRunID, errorCode string, expectedAttempt int32) error {
	run, ok := r.runs[k(tenantID, analysisRunID)]
	if !ok || run.Status != entity.AnalysisPending || run.Attempt != expectedAttempt {
		return entity.ErrStaleGeneration
	}
	run.Status = entity.AnalysisFailed
	run.ErrorCode = errorCode
	run.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memRepo) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID, actorID string) (bool, int32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.claimAnalysisRetry(ctx, tenantID, analysisRunID, actorID)
}
func (t *memTx) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID, actorID string) (bool, int32, error) {
	return t.r.claimAnalysisRetry(ctx, tenantID, analysisRunID, actorID)
}
func (r *memRepo) claimAnalysisRetry(_ context.Context, tenantID, analysisRunID, _ string) (bool, int32, error) {
	run, ok := r.runs[k(tenantID, analysisRunID)]
	if !ok || run.Status != entity.AnalysisFailed {
		return false, 0, nil
	}
	now := time.Now().UTC()
	exp := now.Add(5 * time.Minute)
	run.Status = entity.AnalysisPending
	run.Attempt++
	run.ExecutionClaimedAt = &now
	run.LeaseExpiresAt = &exp
	run.ErrorCode = ""
	run.UpdatedAt = now
	return true, run.Attempt, nil
}

func (r *memRepo) ClaimExpiredPendingExecution(ctx context.Context, tenantID, analysisRunID string, expectedAttempt int32, now time.Time) (*entity.CapabilityAnalysisRun, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.claimExpiredPendingExecution(ctx, tenantID, analysisRunID, expectedAttempt, now)
}
func (t *memTx) ClaimExpiredPendingExecution(ctx context.Context, tenantID, analysisRunID string, expectedAttempt int32, now time.Time) (*entity.CapabilityAnalysisRun, bool, error) {
	return t.r.claimExpiredPendingExecution(ctx, tenantID, analysisRunID, expectedAttempt, now)
}
func (r *memRepo) claimExpiredPendingExecution(_ context.Context, tenantID, analysisRunID string, expectedAttempt int32, now time.Time) (*entity.CapabilityAnalysisRun, bool, error) {
	run, ok := r.runs[k(tenantID, analysisRunID)]
	if !ok {
		return nil, false, entity.ErrAnalysisNotFound
	}
	if run.Status != entity.AnalysisPending || run.Attempt != expectedAttempt || run.LeaseExpiresAt == nil || now.Before(*run.LeaseExpiresAt) {
		return cloneAnalysisRun(run), false, nil
	}
	exp := now.Add(5 * time.Minute)
	run.Attempt++
	claimed := now
	run.ExecutionClaimedAt = &claimed
	run.LeaseExpiresAt = &exp
	run.UpdatedAt = now
	return cloneAnalysisRun(run), true, nil
}

func cloneRevision(in *entity.BusinessCapabilityRevision) *entity.BusinessCapabilityRevision {
	if in == nil {
		return nil
	}
	b, _ := json.Marshal(in)
	var out entity.BusinessCapabilityRevision
	_ = json.Unmarshal(b, &out)
	return &out
}

func cloneProposal(in *entity.CapabilityProposal) *entity.CapabilityProposal {
	if in == nil {
		return nil
	}
	b, _ := json.Marshal(in)
	var out entity.CapabilityProposal
	_ = json.Unmarshal(b, &out)
	return &out
}

func cloneCapMap(in map[string]*entity.BusinessCapability) map[string]*entity.BusinessCapability {
	out := make(map[string]*entity.BusinessCapability, len(in))
	for key, v := range in {
		cp := *v
		out[key] = &cp
	}
	return out
}
func cloneRevMap(in map[string]*entity.BusinessCapabilityRevision) map[string]*entity.BusinessCapabilityRevision {
	out := make(map[string]*entity.BusinessCapabilityRevision, len(in))
	for key, v := range in {
		out[key] = cloneRevision(v)
	}
	return out
}
func clonePropMap(in map[string]*entity.CapabilityProposal) map[string]*entity.CapabilityProposal {
	out := make(map[string]*entity.CapabilityProposal, len(in))
	for key, v := range in {
		out[key] = cloneProposal(v)
	}
	return out
}
func cloneDecMap(in map[string]*entity.CapabilityDecision) map[string]*entity.CapabilityDecision {
	out := make(map[string]*entity.CapabilityDecision, len(in))
	for key, v := range in {
		cp := *v
		out[key] = &cp
	}
	return out
}
func cloneRunMap(in map[string]*entity.CapabilityAnalysisRun) map[string]*entity.CapabilityAnalysisRun {
	out := make(map[string]*entity.CapabilityAnalysisRun, len(in))
	for key, v := range in {
		out[key] = cloneAnalysisRun(v)
	}
	return out
}

func cloneAnalysisRun(in *entity.CapabilityAnalysisRun) *entity.CapabilityAnalysisRun {
	if in == nil {
		return nil
	}
	cp := *in
	if in.ExecutionClaimedAt != nil {
		t := *in.ExecutionClaimedAt
		cp.ExecutionClaimedAt = &t
	}
	if in.LeaseExpiresAt != nil {
		t := *in.LeaseExpiresAt
		cp.LeaseExpiresAt = &t
	}
	return &cp
}
func cloneStrMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, v := range in {
		out[key] = v
	}
	return out
}

// SetLeaseExpiresAtForTest adjusts lease expiry (tests only).
func SetLeaseExpiresAtForTest(repo CapabilityRepository, _ context.Context, tenantID, analysisRunID string, expiresAt time.Time) error {
	r, ok := repo.(*memRepo)
	if !ok {
		return errors.New("not memory repo")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.runs[k(tenantID, analysisRunID)]
	if !ok {
		return entity.ErrAnalysisNotFound
	}
	run.LeaseExpiresAt = &expiresAt
	return nil
}
