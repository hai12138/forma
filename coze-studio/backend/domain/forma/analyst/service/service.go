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

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
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

	CreateProposal(ctx context.Context, tenantID, businessID, sessionID, actorID string, assertionIDs []string) (*entity.BusinessModelProposal, error)
	GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error)
	ApplyProposal(ctx context.Context, tenantID, businessID, proposalID, actorID string) (*entity.BusinessModelRevision, error)
	GetProvenance(ctx context.Context, tenantID, businessID string, revisionNo int32) (*entity.RevisionProvenance, error)
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
	Model        FormaAnalystModel
	AuditHook    AuditHook
}

type AuditHook interface {
	RecordAnalystAudit(ctx context.Context, tenantID, actorID, action, resourceID, requestID string) error
}

type noopAudit struct{}

func (noopAudit) RecordAnalystAudit(_ context.Context, _, _, _, _, _ string) error { return nil }

type analystServiceImpl struct {
	repo        repository.AnalystRepository
	businessSVC businesssvc.BusinessService
	model       FormaAnalystModel
	audit       AuditHook
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
		model = NewDeterministicFakeModel()
	}
	return &analystServiceImpl{
		repo:        c.Repo,
		businessSVC: c.BusinessSVC,
		model:       model,
		audit:       audit,
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

	now := time.Now().UTC()
	seq, err := s.repo.MaxTurnSequence(ctx, tenantID, sessionID)
	if err != nil {
		return nil, err
	}
	turnID := newID("turn")
	clientReq := clientRequestID
	if clientReq == "" {
		clientReq = turnID
	}
	userTurn := &entity.AnalystTurn{
		TurnID:          turnID,
		TenantID:        tenantID,
		SessionID:       sessionID,
		Sequence:        seq + 1,
		Speaker:         entity.SpeakerUser,
		Content:         content,
		ContentType:     entity.ContentText,
		ClientRequestID: clientReq,
		AnalysisStatus:  entity.AnalysisPending,
		CreatedAt:       now,
	}
	if err := s.repo.CreateTurn(ctx, userTurn); err != nil {
		return nil, err
	}

	evidence := &entity.BusinessEvidence{
		EvidenceID:    newID("evid"),
		TenantID:      tenantID,
		BusinessID:    businessID,
		SessionID:     sessionID,
		TurnID:        userTurn.TurnID,
		SourceType:    entity.EvidenceInterviewTurn,
		SourceRef:     userTurn.TurnID,
		Quote:         content,
		ContentDigest: digestString(content),
		CreatedBy:     actorID,
		CreatedAt:     now,
	}
	if err := s.repo.CreateEvidence(ctx, evidence); err != nil {
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ANALYST_TURN_SUBMITTED", userTurn.TurnID, clientRequestID)

	ctxInput, manifest, err := s.buildContextInput(ctx, tenantID, businessID, sessionID)
	if err != nil {
		return nil, err
	}
	_, contextText := BuildContext(ctxInput)
	manifest.IncludedItems = append(manifest.IncludedItems, contextText != "")

	modelRequestID := newID("mreq")
	start := time.Now()
	extractReq := &ExtractionRequest{
		RequestID:       modelRequestID,
		TenantID:        tenantID,
		BusinessID:      businessID,
		SessionID:       sessionID,
		ContextManifest: manifest,
		SystemPolicy:    analystSystemPolicy,
		UserTurnContent: content,
		UserTurnID:      userTurn.TurnID,
	}
	extraction, modelErr := s.model.ExtractAssertions(ctx, extractReq)
	latency := int32(time.Since(start).Milliseconds())

	result := &entity.TurnSubmissionResult{
		UserTurn:        userTurn,
		Evidence:        evidence,
		ContextManifest: manifest,
	}

	if modelErr != nil {
		_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisFailed, modelRequestID)
		_ = s.repo.CreateModelCall(ctx, &entity.ModelCallRecord{
			RequestID:    modelRequestID,
			TenantID:     tenantID,
			BusinessID:   businessID,
			SessionID:    sessionID,
			Operation:    "ExtractAssertions",
			LatencyMs:    latency,
			Success:      false,
			ErrorMessage: modelErr.Error(),
			CreatedAt:    now,
		})
		result.ModelFailed = true
		result.ModelError = modelErr.Error()
		userTurn.AnalysisStatus = entity.AnalysisFailed
		return result, nil
	}

	assertions, conflicts, gaps, err := s.persistExtraction(ctx, tenantID, businessID, sessionID, actorID, evidence, extraction, now)
	if err != nil {
		_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisFailed, modelRequestID)
		result.ModelFailed = true
		result.ModelError = err.Error()
		return result, nil
	}
	_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisCompleted, modelRequestID)
	userTurn.AnalysisStatus = entity.AnalysisCompleted

	plan := PlanNextQuestion(ctxInput)
	planReq := &InterviewTurnRequest{
		RequestID:       newID("mreq"),
		TenantID:        tenantID,
		BusinessID:      businessID,
		SessionID:       sessionID,
		ContextManifest: manifest,
		SystemPolicy:    analystSystemPolicy,
		UserMessage:     content,
		NextQuestion:    plan,
	}
	turnResp, turnErr := s.model.GenerateInterviewTurn(ctx, planReq)
	analystContent := "我已分析您的描述并提取了业务事实，请在右侧查看并确认。"
	if turnErr == nil && turnResp != nil && turnResp.Content != "" {
		analystContent = turnResp.Content
	}

	analystTurn := &entity.AnalystTurn{
		TurnID:         newID("turn"),
		TenantID:       tenantID,
		SessionID:      sessionID,
		Sequence:       userTurn.Sequence + 1,
		Speaker:        entity.SpeakerAnalyst,
		Content:        analystContent,
		ContentType:    entity.ContentText,
		ModelRequestID: modelRequestID,
		AnalysisStatus: entity.AnalysisCompleted,
		CreatedAt:      now,
	}
	if err := s.repo.CreateTurn(ctx, analystTurn); err != nil {
		return nil, err
	}

	_ = s.repo.CreateModelCall(ctx, &entity.ModelCallRecord{
		RequestID:    modelRequestID,
		TenantID:     tenantID,
		BusinessID:   businessID,
		SessionID:    sessionID,
		Operation:    "ExtractAssertions",
		LatencyMs:    latency,
		Success:      true,
		CreatedAt:    now,
	})

	result.Assertions = assertions
	result.Conflicts = conflicts
	result.Gaps = gaps
	result.AnalystTurn = analystTurn
	result.NextQuestion = plan
	return result, nil
}

func (s *analystServiceImpl) buildIdempotentResult(ctx context.Context, tenantID, businessID, sessionID string, userTurn *entity.AnalystTurn) (*entity.TurnSubmissionResult, error) {
	turns, err := s.repo.ListTurns(ctx, tenantID, sessionID)
	if err != nil {
		return nil, err
	}
	var analystTurn *entity.AnalystTurn
	for _, t := range turns {
		if t.Sequence == userTurn.Sequence+1 && t.Speaker == entity.SpeakerAnalyst {
			analystTurn = t
			break
		}
	}
	evidenceList, err := s.repo.ListEvidence(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	var evidence *entity.BusinessEvidence
	for _, e := range evidenceList {
		if e.TurnID == userTurn.TurnID {
			evidence = e
			break
		}
	}
	assertions, err := s.loadAssertionsWithEvidence(ctx, tenantID, businessID)
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

func (s *analystServiceImpl) persistExtraction(
	ctx context.Context,
	tenantID, businessID, sessionID, actorID string,
	evidence *entity.BusinessEvidence,
	extraction *entity.ExtractionResult,
	now time.Time,
) ([]*entity.BusinessAssertion, []*entity.AssertionConflict, []*entity.AnalystGap, error) {
	if extraction == nil {
		return nil, nil, nil, nil
	}
	var assertions []*entity.BusinessAssertion
	turnEvidenceMap := map[string]string{}
	if evidence != nil {
		turnEvidenceMap[evidence.TurnID] = evidence.EvidenceID
	}

	for _, ea := range extraction.Assertions {
		a := &entity.BusinessAssertion{
			AssertionID:   newID("assert"),
			TenantID:      tenantID,
			BusinessID:    businessID,
			SessionID:     sessionID,
			AssertionType: ea.AssertionType,
			SubjectRef:    ea.SubjectRef,
			Predicate:     ea.Predicate,
			ObjectValue:   ea.ObjectValue,
			StructuredValue: ea.StructuredValue,
			Confidence:    ea.Confidence,
			Status:        entity.AssertionProposed,
			SourceMarker:  businessentity.SourceAIGenerated,
			CreatedBy:     actorID,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.repo.CreateAssertion(ctx, a); err != nil {
			return nil, nil, nil, err
		}
		linked := false
		for _, tid := range ea.EvidenceTurnIDs {
			eid := turnEvidenceMap[tid]
			if eid == "" && evidence != nil && evidence.TurnID == tid {
				eid = evidence.EvidenceID
			}
			if eid != "" {
				if err := s.repo.CreateAssertionEvidenceRef(ctx, tenantID, a.AssertionID, eid, now); err != nil {
					return nil, nil, nil, err
				}
				linked = true
			}
		}
		if !linked && evidence != nil {
			if err := s.repo.CreateAssertionEvidenceRef(ctx, tenantID, a.AssertionID, evidence.EvidenceID, now); err != nil {
				return nil, nil, nil, err
			}
		}
		a.EvidenceIDs, _ = s.repo.ListEvidenceIDsForAssertion(ctx, tenantID, a.AssertionID)
		assertions = append(assertions, a)
		_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ASSERTION_CREATED", a.AssertionID, "")
	}

	for _, link := range extraction.EvidenceLinks {
		if link.AssertionIndex < 0 || link.AssertionIndex >= len(assertions) {
			continue
		}
		a := assertions[link.AssertionIndex]
		eid := turnEvidenceMap[link.EvidenceTurnID]
		if eid != "" {
			_ = s.repo.CreateAssertionEvidenceRef(ctx, tenantID, a.AssertionID, eid, now)
		}
	}

	var gaps []*entity.AnalystGap
	for _, g := range extraction.Gaps {
		gap := &entity.AnalystGap{
			GapID:      newID("gap"),
			TenantID:   tenantID,
			BusinessID: businessID,
			SessionID:  sessionID,
			GapType:    g.GapType,
			Question:   g.Question,
			Status:     entity.GapOpen,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.repo.CreateGap(ctx, gap); err != nil {
			return nil, nil, nil, err
		}
		gaps = append(gaps, gap)
	}

	var conflicts []*entity.AssertionConflict
	// Deterministic conflict detection from persisted assertions
	allProposed := append([]*entity.BusinessAssertion{}, assertions...)
	existing, _ := s.loadAssertionsWithEvidence(ctx, tenantID, businessID)
	for _, e := range existing {
		if e.Status == entity.AssertionProposed {
			allProposed = append(allProposed, e)
		}
	}
	conflicts = detectConflicts(allProposed, tenantID, businessID, sessionID, now)
	for _, c := range conflicts {
		if err := s.repo.CreateConflict(ctx, c); err != nil {
			return nil, nil, nil, err
		}
	}

	// Model-reported conflicts
	for _, mc := range extraction.Conflicts {
		if mc.AssertionIndexA < len(assertions) && mc.AssertionIndexB < len(assertions) {
			c := &entity.AssertionConflict{
				ConflictID:   newID("conf"),
				TenantID:     tenantID,
				BusinessID:   businessID,
				SessionID:    sessionID,
				AssertionIDA: assertions[mc.AssertionIndexA].AssertionID,
				AssertionIDB: assertions[mc.AssertionIndexB].AssertionID,
				SubjectRef:   assertions[mc.AssertionIndexA].SubjectRef,
				Predicate:    assertions[mc.AssertionIndexA].Predicate,
				Status:       entity.ConflictOpen,
				CreatedAt:    now,
			}
			if err := s.repo.CreateConflict(ctx, c); err != nil {
				return nil, nil, nil, err
			}
			conflicts = append(conflicts, c)
		}
	}

	return assertions, conflicts, gaps, nil
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

func (s *analystServiceImpl) ConfirmAssertion(ctx context.Context, tenantID, businessID, assertionID, actorID, comment string, edit *AssertionEdit) (*entity.BusinessAssertion, error) {
	a, err := s.getAssertionForBusiness(ctx, tenantID, businessID, assertionID)
	if err != nil {
		return nil, err
	}
	if a.Status == entity.AssertionConfirmed || a.Status == entity.AssertionRejected {
		return nil, entity.ErrAssertionAlreadyDecided
	}
	evIDs, _ := s.repo.ListEvidenceIDsForAssertion(ctx, tenantID, assertionID)
	if len(evIDs) == 0 {
		return nil, entity.ErrAssertionEvidenceRequired
	}
	now := time.Now().UTC()
	if edit != nil {
		a.AssertionType = edit.AssertionType
		a.SubjectRef = edit.SubjectRef
		a.Predicate = edit.Predicate
		a.ObjectValue = edit.ObjectValue
		a.SourceMarker = businessentity.SourceManualModified
	}
	a.Status = entity.AssertionConfirmed
	a.UpdatedAt = now
	if err := s.repo.UpdateAssertion(ctx, a); err != nil {
		return nil, err
	}
	conf := &entity.BusinessConfirmation{
		ConfirmationID: newID("confm"),
		TenantID:       tenantID,
		BusinessID:     businessID,
		AssertionID:    assertionID,
		Decision:       entity.DecisionConfirm,
		Comment:        comment,
		DecidedBy:      actorID,
		DecidedAt:      now,
	}
	if err := s.repo.CreateConfirmation(ctx, conf); err != nil {
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ASSERTION_CONFIRMED", assertionID, "")
	a.EvidenceIDs = evIDs
	return a, nil
}

func (s *analystServiceImpl) RejectAssertion(ctx context.Context, tenantID, businessID, assertionID, actorID, comment string) (*entity.BusinessAssertion, error) {
	a, err := s.getAssertionForBusiness(ctx, tenantID, businessID, assertionID)
	if err != nil {
		return nil, err
	}
	if a.Status == entity.AssertionConfirmed || a.Status == entity.AssertionRejected {
		return nil, entity.ErrAssertionAlreadyDecided
	}
	now := time.Now().UTC()
	a.Status = entity.AssertionRejected
	a.UpdatedAt = now
	if err := s.repo.UpdateAssertion(ctx, a); err != nil {
		return nil, err
	}
	conf := &entity.BusinessConfirmation{
		ConfirmationID: newID("confm"),
		TenantID:       tenantID,
		BusinessID:     businessID,
		AssertionID:    assertionID,
		Decision:       entity.DecisionReject,
		Comment:        comment,
		DecidedBy:      actorID,
		DecidedAt:      now,
	}
	if err := s.repo.CreateConfirmation(ctx, conf); err != nil {
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ASSERTION_REJECTED", assertionID, "")
	a.EvidenceIDs, _ = s.repo.ListEvidenceIDsForAssertion(ctx, tenantID, assertionID)
	return a, nil
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
		all, err := s.loadAssertionsWithEvidence(ctx, tenantID, businessID)
		if err != nil {
			return nil, err
		}
		for _, a := range all {
			if a.Status == entity.AssertionConfirmed {
				confirmed = append(confirmed, a)
				assertionIDs = append(assertionIDs, a.AssertionID)
			}
		}
	} else {
		for _, id := range assertionIDs {
			a, err := s.getAssertionForBusiness(ctx, tenantID, businessID, id)
			if err != nil {
				return nil, err
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

func (s *analystServiceImpl) ApplyProposal(ctx context.Context, tenantID, businessID, proposalID, actorID string) (*businessentity.BusinessModelRevision, error) {
	proposal, err := s.GetProposal(ctx, tenantID, proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.BusinessID != businessID {
		return nil, entity.ErrProposalNotFound
	}
	if proposal.Status == entity.ProposalApplied {
		return nil, entity.ErrProposalAlreadyApplied
	}
	if proposal.Status == entity.ProposalRejected {
		return nil, entity.ErrProposalInvalid
	}

	master, model, _, err := s.businessSVC.GetModel(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	if master.CurrentRevision != proposal.BaseRevision {
		_ = s.repo.UpdateProposalStatus(ctx, tenantID, proposalID, entity.ProposalStale, time.Now().UTC())
		return nil, entity.ErrProposalStale
	}

	newModel, err := ApplyPatch(model, proposal.Patch)
	if err != nil {
		return nil, err
	}

	// Collect evidence/assertion refs for provenance on semantic model
	newModel.EvidenceRefs = collectEvidenceRefs(ctx, s.repo, tenantID, businessID, proposal.AssertionIDs)
	newModel.AssertionRefs = append([]string(nil), proposal.AssertionIDs...)

	summary := fmt.Sprintf("Apply analyst proposal %s", proposal.ProposalID)
	rev, noChange, err := s.businessSVC.SaveModel(ctx, tenantID, businessID, actorID, proposal.BaseRevision, newModel, summary)
	if err != nil {
		return nil, err
	}
	if noChange {
		return nil, entity.ErrProposalInvalid
	}

	now := time.Now().UTC()
	_ = s.repo.UpdateProposalStatus(ctx, tenantID, proposalID, entity.ProposalApplied, now)
	_ = s.repo.CreateProvenance(ctx, &entity.RevisionProvenance{
		TenantID:     tenantID,
		BusinessID:   businessID,
		RevisionNo:   rev.RevisionNo,
		ProposalID:   proposal.ProposalID,
		AssertionIDs: proposal.AssertionIDs,
		CreatedAt:    now,
	})
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "PROPOSAL_APPLIED", proposalID, "")

	return rev, nil
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

func (s *analystServiceImpl) buildContextInput(ctx context.Context, tenantID, businessID, sessionID string) (*ContextInput, *entity.ContextManifest, error) {
	input := &ContextInput{EvidenceByTurn: map[string]*entity.BusinessEvidence{}}
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
	manifest, _ := BuildContext(input)
	return input, manifest, nil
}

func newID(prefix string) string {
	return prefix + "_" + uuid.NewString()
}

func digestString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// BusinessModelRevision alias for ApplyProposal return
type BusinessModelRevision = businessentity.BusinessModelRevision
