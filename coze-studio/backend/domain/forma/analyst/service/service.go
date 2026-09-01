/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businessrepo "github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
)

type AnalystService interface {
	CreateSession(ctx context.Context, tenantID, businessID, title, createdBy string, policy entity.ConfirmationPolicy) (*entity.AnalystSession, error)
	ListSessions(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystSession, error)
	GetSession(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error)

	SubmitTurn(ctx context.Context, tenantID, businessID, sessionID, content, clientRequestID, actorID string) (*entity.TurnSubmissionResult, error)
	ListTurns(ctx context.Context, tenantID, sessionID string) ([]*entity.AnalystTurn, error)

	ListAssertions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessAssertion, error)
	ListEvidence(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessEvidence, error)

	ConfirmAssertion(ctx context.Context, tenantID, businessID, assertionID, actorID, comment string, edit *AssertionEdit) (*entity.BusinessAssertion, error)
	RejectAssertion(ctx context.Context, tenantID, businessID, assertionID, actorID, comment string) (*entity.BusinessAssertion, error)

	ListConflicts(ctx context.Context, tenantID, businessID string) ([]*entity.AssertionConflict, error)
	ListGaps(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystGap, error)
	AskGap(ctx context.Context, tenantID, businessID, sessionID, gapID, actorID string) (*entity.GapAskResult, error)

	CreateProposal(ctx context.Context, tenantID, businessID, sessionID, actorID string, assertionIDs []string) (*entity.BusinessModelProposal, error)
	GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error)
	ApplyProposal(ctx context.Context, tenantID, businessID, proposalID, actorID string) (*businessentity.BusinessModelRevision, error)
	GetProvenance(ctx context.Context, tenantID, businessID string, revisionNo int32) (*entity.RevisionProvenance, error)
	RetryTurnAnalysis(ctx context.Context, tenantID, businessID, sessionID, turnID, actorID string) (*entity.TurnSubmissionResult, error)
	GetProposalPreview(ctx context.Context, tenantID, proposalID string) (*ProposalPreviewResult, error)
}

type AssertionEdit struct {
	AssertionType entity.AssertionType
	SubjectRef    string
	Predicate     string
	ObjectValue   string
}

type Components struct {
	Repo         repository.AnalystRepository
	BusinessSVC  businesssvc.BusinessService
	BusinessRepo businessrepo.BusinessRepository
	DB           *gorm.DB
	Model        FormaAnalystModel
	AuditHook    AuditHook
}

type AuditHook interface {
	RecordAnalystAudit(ctx context.Context, tenantID, actorID, action, resourceID, requestID string) error
}

type noopAudit struct{}

func (noopAudit) RecordAnalystAudit(_ context.Context, _, _, _, _, _ string) error { return nil }

type analystServiceImpl struct {
	repo         repository.AnalystRepository
	businessSVC  businesssvc.BusinessService
	businessRepo businessrepo.BusinessRepository
	db           *gorm.DB
	model        FormaAnalystModel
	audit        AuditHook
}

func NewAnalystService(c *Components) AnalystService {
	if c == nil {
		return &analystServiceImpl{}
	}
	audit := c.AuditHook
	if audit == nil {
		audit = noopAudit{}
	}
	model := c.Model
	if model == nil {
		model = NewUnavailableAnalystModel()
	}
	return &analystServiceImpl{
		repo:         c.Repo,
		businessSVC:  c.BusinessSVC,
		businessRepo: c.BusinessRepo,
		db:           c.DB,
		model:        model,
		audit:        audit,
	}
}

func (s *analystServiceImpl) CreateSession(ctx context.Context, tenantID, businessID, title, createdBy string, policy entity.ConfirmationPolicy) (*entity.AnalystSession, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("analyst repo not initialized")
	}
	if policy == "" {
		policy = entity.PolicyDevelopment
	}
	now := time.Now().UTC()
	session := &entity.AnalystSession{
		SessionID:          newID("asess"),
		TenantID:           tenantID,
		BusinessID:         businessID,
		Status:             entity.SessionActive,
		Title:              title,
		ConfirmationPolicy: policy,
		NextTurnSequence:   1,
		CreatedBy:          createdBy,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, createdBy, "ANALYST_SESSION_CREATED", session.SessionID, "")
	return session, nil
}

func (s *analystServiceImpl) ListSessions(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystSession, error) {
	return s.repo.ListSessions(ctx, tenantID, businessID)
}

func (s *analystServiceImpl) GetSession(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error) {
	session, err := s.repo.GetSession(ctx, tenantID, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, entity.ErrSessionNotFound
	}
	return session, nil
}

func sessionClosed(status entity.SessionStatus) bool {
	return status == entity.SessionCompleted || status == entity.SessionCancelled
}

func (s *analystServiceImpl) SubmitTurn(ctx context.Context, tenantID, businessID, sessionID, content, clientRequestID, actorID string) (*entity.TurnSubmissionResult, error) {
	session, err := s.GetSession(ctx, tenantID, sessionID)
	if err != nil {
		return nil, err
	}
	if session.BusinessID != businessID {
		return nil, entity.ErrSessionNotFound
	}
	if sessionClosed(session.Status) {
		return nil, entity.ErrSessionClosed
	}
	if strings.TrimSpace(content) == "" {
		return nil, entity.ErrInvalidTurn
	}

	// Idempotency
	if clientRequestID != "" {
		existing, err := s.repo.GetTurnByClientRequestID(ctx, tenantID, sessionID, clientRequestID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return s.buildIdempotentResult(ctx, tenantID, businessID, sessionID, existing)
		}
	}

	userTurn, evidence, err := s.createUserTurnWithEvidence(ctx, tenantID, businessID, sessionID, content, clientRequestID, actorID)
	if err != nil {
		if clientRequestID != "" && isDuplicateKeyErr(err) {
			existing, gErr := s.repo.GetTurnByClientRequestID(ctx, tenantID, sessionID, clientRequestID)
			if gErr == nil && existing != nil {
				return s.buildIdempotentResult(ctx, tenantID, businessID, sessionID, existing)
			}
		}
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ANALYST_TURN_SUBMITTED", userTurn.TurnID, clientRequestID)
	return s.runAnalysisForUserTurn(ctx, tenantID, businessID, sessionID, userTurn, evidence, actorID)
}

func (s *analystServiceImpl) buildIdempotentResult(ctx context.Context, tenantID, businessID, sessionID string, userTurn *entity.AnalystTurn) (*entity.TurnSubmissionResult, error) {
	analystTurn, err := s.repo.GetTurnByReplyTo(ctx, tenantID, sessionID, userTurn.TurnID)
	if err != nil {
		return nil, err
	}
	evidenceList, err := s.repo.ListEvidence(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	var evidence *entity.BusinessEvidence
	for _, e := range evidenceList {
		if e != nil && e.TurnID == userTurn.TurnID {
			evidence = e
			break
		}
	}
	assertions, err := s.loadAssertionsForEvidence(ctx, tenantID, businessID, evidence)
	if err != nil {
		return nil, err
	}
	return &entity.TurnSubmissionResult{
		UserTurn:    userTurn,
		AnalystTurn: analystTurn,
		Evidence:    evidence,
		Assertions:  assertions,
	}, nil
}

func detectConflicts(assertions []*entity.BusinessAssertion, tenantID, businessID, sessionID string, now time.Time) []*entity.AssertionConflict {
	type key struct{ subject, predicate string }
	byKey := map[key][]*entity.BusinessAssertion{}
	for _, a := range assertions {
		if a == nil {
			continue
		}
		k := key{a.SubjectRef, a.Predicate}
		byKey[k] = append(byKey[k], a)
	}
	var out []*entity.AssertionConflict
	for k, group := range byKey {
		if len(group) < 2 {
			continue
		}
		values := map[string]bool{}
		for _, a := range group {
			values[a.ObjectValue] = true
		}
		if len(values) > 1 {
			c := &entity.AssertionConflict{
				ConflictID:   newID("conf"),
				TenantID:     tenantID,
				BusinessID:   businessID,
				SessionID:    sessionID,
				AssertionIDA: group[0].AssertionID,
				AssertionIDB: group[1].AssertionID,
				SubjectRef:   k.subject,
				Predicate:    k.predicate,
				Status:       entity.ConflictOpen,
				CreatedAt:    now,
			}
			out = append(out, c)
		}
	}
	return out
}

func (s *analystServiceImpl) ListTurns(ctx context.Context, tenantID, sessionID string) ([]*entity.AnalystTurn, error) {
	return s.repo.ListTurns(ctx, tenantID, sessionID)
}

func (s *analystServiceImpl) loadAssertionsWithEvidence(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessAssertion, error) {
	list, err := s.repo.ListAssertions(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	for _, a := range list {
		ids, _ := s.repo.ListEvidenceIDsForAssertion(ctx, tenantID, a.AssertionID)
		a.EvidenceIDs = ids
	}
	return list, nil
}

func (s *analystServiceImpl) ListAssertions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessAssertion, error) {
	return s.loadAssertionsWithEvidence(ctx, tenantID, businessID)
}

func (s *analystServiceImpl) ListEvidence(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessEvidence, error) {
	return s.repo.ListEvidence(ctx, tenantID, businessID)
}

func (s *analystServiceImpl) getAssertionForBusiness(ctx context.Context, tenantID, businessID, assertionID string) (*entity.BusinessAssertion, error) {
	a, err := s.repo.GetAssertion(ctx, tenantID, assertionID)
	if err != nil {
		return nil, err
	}
	if a == nil || a.BusinessID != businessID {
		return nil, entity.ErrAssertionNotFound
	}
	return a, nil
}

func (s *analystServiceImpl) ListConflicts(ctx context.Context, tenantID, businessID string) ([]*entity.AssertionConflict, error) {
	return s.repo.ListConflicts(ctx, tenantID, businessID)
}

func (s *analystServiceImpl) ListGaps(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystGap, error) {
	return s.repo.ListGaps(ctx, tenantID, businessID)
}

func (s *analystServiceImpl) CreateProposal(ctx context.Context, tenantID, businessID, sessionID, actorID string, assertionIDs []string) (*entity.BusinessModelProposal, error) {
	if s.businessSVC == nil {
		return nil, fmt.Errorf("business service required")
	}
	master, model, _, err := s.businessSVC.GetModel(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	if master == nil {
		return nil, entity.ErrNotFound
	}

	var confirmed []*entity.BusinessAssertion
	if len(assertionIDs) == 0 {
		if sessionID == "" {
			return nil, fmt.Errorf("%w: session_id required", entity.ErrProposalInvalid)
		}
		all, err := s.loadAssertionsWithEvidence(ctx, tenantID, businessID)
		if err != nil {
			return nil, err
		}
		confirmed = filterAssertionsBySession(all, sessionID, true)
		for _, a := range confirmed {
			assertionIDs = append(assertionIDs, a.AssertionID)
		}
	} else {
		for _, id := range assertionIDs {
			a, err := s.getAssertionForBusiness(ctx, tenantID, businessID, id)
			if err != nil {
				return nil, err
			}
			if sessionID != "" && a.SessionID != sessionID {
				return nil, fmt.Errorf("%w: assertion %s not in session", entity.ErrProposalInvalid, id)
			}
			if a.Status != entity.AssertionConfirmed {
				return nil, fmt.Errorf("%w: assertion %s not confirmed", entity.ErrProposalInvalid, id)
			}
			confirmed = append(confirmed, a)
		}
	}
	if len(confirmed) == 0 {
		return nil, fmt.Errorf("%w: no confirmed assertions", entity.ErrProposalInvalid)
	}

	patchReq := &ProposalRequest{
		RequestID:    newID("mreq"),
		TenantID:     tenantID,
		BusinessID:   businessID,
		SessionID:    sessionID,
		Assertions:   confirmed,
		BaseRevision: master.CurrentRevision,
	}
	patch, err := s.model.ProposeModelPatch(ctx, patchReq)
	if err != nil {
		return nil, err
	}
	if patch == nil || len(patch.Operations) == 0 {
		patch = BuildProposalPatch(confirmed)
	}
	_, err = ApplyPatch(model, patch)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	proposal := &entity.BusinessModelProposal{
		ProposalID:    newID("prop"),
		TenantID:      tenantID,
		BusinessID:    businessID,
		SessionID:     sessionID,
		BaseRevision:  master.CurrentRevision,
		AssertionIDs:  assertionIDs,
		Patch:         patch,
		Status:        entity.ProposalReadyForReview,
		ContentDigest: ProposalDigest(patch, master.CurrentRevision, assertionIDs),
		CreatedBy:     actorID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateProposal(ctx, proposal); err != nil {
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "PROPOSAL_CREATED", proposal.ProposalID, "")
	return proposal, nil
}

func (s *analystServiceImpl) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error) {
	p, err := s.repo.GetProposal(ctx, tenantID, proposalID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, entity.ErrProposalNotFound
	}
	return p, nil
}

func collectEvidenceRefs(ctx context.Context, repo repository.AnalystRepository, tenantID, businessID string, assertionIDs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, aid := range assertionIDs {
		ids, _ := repo.ListEvidenceIDsForAssertion(ctx, tenantID, aid)
		for _, eid := range ids {
			if !seen[eid] {
				seen[eid] = true
				out = append(out, eid)
			}
		}
	}
	return out
}

func (s *analystServiceImpl) GetProvenance(ctx context.Context, tenantID, businessID string, revisionNo int32) (*entity.RevisionProvenance, error) {
	return s.repo.GetProvenance(ctx, tenantID, businessID, revisionNo)
}

func (s *analystServiceImpl) buildContextInput(ctx context.Context, tenantID, businessID, sessionID string) (*ContextInput, *entity.ContextManifest, string, error) {
	input := &ContextInput{EvidenceByTurn: map[string]*entity.BusinessEvidence{}}
	if sess, _ := s.repo.GetSession(ctx, tenantID, sessionID); sess != nil {
		input.FocusGapID = sess.FocusGapID
	}
	if s.businessSVC != nil {
		_, model, _, err := s.businessSVC.GetModel(ctx, tenantID, businessID)
		if err == nil {
			input.CurrentModel = model
		}
	}
	all, _ := s.loadAssertionsWithEvidence(ctx, tenantID, businessID)
	for _, a := range all {
		if a.SessionID != sessionID {
			continue
		}
		switch a.Status {
		case entity.AssertionConfirmed:
			input.Confirmed = append(input.Confirmed, a)
		case entity.AssertionProposed:
			input.Proposed = append(input.Proposed, a)
		}
	}
	input.OpenConflicts, _ = s.repo.ListConflicts(ctx, tenantID, businessID)
	input.OpenGaps, _ = s.repo.ListGaps(ctx, tenantID, businessID)
	input.RecentTurns, _ = s.repo.ListTurns(ctx, tenantID, sessionID)
	evList, _ := s.repo.ListEvidence(ctx, tenantID, businessID)
	for _, e := range evList {
		if e.SessionID == sessionID && e.TurnID != "" {
			input.EvidenceByTurn[e.TurnID] = e
		}
	}
	manifest, contextText := BuildContext(input)
	return input, manifest, contextText, nil
}

func newID(prefix string) string {
	return prefix + "_" + uuid.NewString()
}

func digestString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
