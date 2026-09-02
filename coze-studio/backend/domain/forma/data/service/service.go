/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

type DataService interface {
	AnalyzeDataRequirements(ctx context.Context, in *AnalyzeInput) (*AnalyzeResult, error)
	GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error)
	RetryFailedAnalysis(ctx context.Context, tenantID, analysisRunID, actorID string) (*AnalyzeResult, error)

	ListRequirements(ctx context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error)
	GetRequirement(ctx context.Context, tenantID, requirementID string) (*entity.DataRequirement, error)
	CreateManualRequirement(ctx context.Context, in *ManualCreateInput) (*entity.DataRequirement, error)

	ConfirmRequirement(ctx context.Context, tenantID, requirementID, actorID, reason string) (*entity.DataRequirement, *entity.DataRequirementDecision, error)
	RejectRequirement(ctx context.Context, tenantID, requirementID, actorID, reason string) (*entity.DataRequirement, *entity.DataRequirementDecision, error)
	EditConfirmRequirement(ctx context.Context, in *EditConfirmInput) (*entity.DataRequirement, *entity.DataRequirement, *entity.DataRequirementDecision, error)

	ListDecisions(ctx context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error)
}

type Components struct {
	Repo        repository.DataRepository
	BusinessSVC businesssvc.BusinessService
	Model       FormaDataModel
}

type dataServiceImpl struct {
	repo        repository.DataRepository
	businessSVC businesssvc.BusinessService
	model       FormaDataModel
}

func NewDataService(c *Components) DataService {
	return &dataServiceImpl{repo: c.Repo, businessSVC: c.BusinessSVC, model: c.Model}
}

type AnalyzeInput struct {
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	ClientRequestID       string
	ActorID               string
}

type AnalyzeResult struct {
	Run          *entity.DataRequirementAnalysisRun
	Requirements []*entity.DataRequirement
	OwnedExecute bool
}

type ManualCreateInput struct {
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	ActorID               string
	RequirementKind       entity.RequirementKind
	SemanticName          string
	Description           string
	BusinessElementRefs   []string
	Requiredness          string
	FreshnessRequirement  string
	AccessNeed            string
}

type EditConfirmInput struct {
	TenantID             string
	SourceRequirementID  string
	ActorID              string
	Reason               string
	RequirementKind      entity.RequirementKind
	SemanticName         string
	Description          string
	BusinessElementRefs  []string
	Requiredness         string
	FreshnessRequirement string
	AccessNeed           string
}

func newID(prefix string) string {
	return prefix + "_" + uuid.NewString()
}

func requestDigest(in *AnalyzeInput) string {
	return DigestJSON(in)
}

func sanitizeError(msg string) string {
	lower := strings.ToLower(msg)
	forbidden := []string{"authorization", "cookie", "api_key", "apikey", "password", "secret", "bearer "}
	for _, f := range forbidden {
		if strings.Contains(lower, f) {
			return "sanitized error"
		}
	}
	if len(msg) > 1024 {
		return msg[:1024]
	}
	return msg
}

func collectElementIDs(sm *businessentity.SemanticModel) map[string]struct{} {
	ids := map[string]struct{}{}
	if sm == nil {
		return ids
	}
	for _, n := range sm.Nodes {
		ids[n.ID] = struct{}{}
	}
	for _, e := range sm.Edges {
		ids[e.ID] = struct{}{}
	}
	for _, r := range sm.Rules {
		ids[r.ID] = struct{}{}
	}
	for _, s := range sm.States {
		ids[s.ID] = struct{}{}
	}
	return ids
}

func validKind(k entity.RequirementKind) bool {
	switch k {
	case entity.KindEntity, entity.KindAttribute, entity.KindRelation, entity.KindEvent,
		entity.KindMetric, entity.KindState, entity.KindTimeSeries, entity.KindDocument,
		entity.KindLookup, entity.KindHistory:
		return true
	default:
		return false
	}
}

func validateProposals(proposals []entity.DataRequirementProposal, ids map[string]struct{}) error {
	for i, p := range proposals {
		if !validKind(p.RequirementKind) {
			return fmt.Errorf("%w: proposal[%d] invalid kind", entity.ErrInvalidProposal, i)
		}
		if strings.TrimSpace(p.SemanticName) == "" {
			return fmt.Errorf("%w: proposal[%d] empty semantic_name", entity.ErrInvalidProposal, i)
		}
		if p.Status != "" && p.Status != entity.StatusProposed {
			return fmt.Errorf("%w: proposal[%d] forbidden status %s", entity.ErrInvalidProposal, i, p.Status)
		}
		if p.Source != "" && p.Source != entity.SourceAIGenerated {
			return fmt.Errorf("%w: proposal[%d] forbidden source %s", entity.ErrInvalidProposal, i, p.Source)
		}
		if len(p.BusinessElementRefs) == 0 {
			return fmt.Errorf("%w: proposal[%d] missing business_element_refs", entity.ErrInvalidProposal, i)
		}
		for _, ref := range p.BusinessElementRefs {
			if _, ok := ids[ref]; !ok {
				return fmt.Errorf("%w: %s", entity.ErrBusinessElementRefInvalid, ref)
			}
		}
	}
	return nil
}

func (s *dataServiceImpl) loadPinnedSemantic(ctx context.Context, tenantID, businessID string, revision int32) (*businessentity.SemanticModel, error) {
	if s.businessSVC == nil {
		return nil, entity.ErrNotConfigured
	}
	_, sm, err := s.businessSVC.GetRevision(ctx, tenantID, businessID, revision)
	if err != nil {
		return nil, entity.ErrBusinessRevisionNotFound
	}
	return sm, nil
}

func (s *dataServiceImpl) AnalyzeDataRequirements(ctx context.Context, in *AnalyzeInput) (*AnalyzeResult, error) {
	if s.repo == nil || s.model == nil || s.businessSVC == nil {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || in.TenantID == "" || in.BusinessID == "" || in.ClientRequestID == "" || in.BusinessModelRevision <= 0 {
		return nil, entity.ErrInvalidProposal
	}

	sm, err := s.loadPinnedSemantic(ctx, in.TenantID, in.BusinessID, in.BusinessModelRevision)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	digest := requestDigest(in)
	run := &entity.DataRequirementAnalysisRun{
		AnalysisRunID:         newID("arun"),
		TenantID:              in.TenantID,
		BusinessID:            in.BusinessID,
		BusinessModelRevision: in.BusinessModelRevision,
		ClientRequestID:       in.ClientRequestID,
		RequestDigest:         digest,
		Status:                entity.AnalysisPending,
		CreatedBy:             in.ActorID,
		StartedAt:             &now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	seedExecutionLease(run, now)

	existing, created, err := s.repo.CreateOrClaimAnalysisRun(ctx, run)
	if err != nil {
		return nil, err
	}
	if !created {
		return s.handleExistingAnalysisRun(ctx, in, existing, sm)
	}

	return s.executeAnalysis(ctx, existing, sm, in.ActorID, existing.ExecutionGeneration)
}

func (s *dataServiceImpl) handleExistingAnalysisRun(ctx context.Context, in *AnalyzeInput, existing *entity.DataRequirementAnalysisRun, sm *businessentity.SemanticModel) (*AnalyzeResult, error) {
	reqs, lerr := s.repo.ListRequirementsByRevision(ctx, in.TenantID, in.BusinessID, in.BusinessModelRevision)
	if lerr != nil {
		return nil, lerr
	}
	filtered := filterByAnalysisRun(reqs, existing.AnalysisRunID)
	now := time.Now().UTC()

	switch existing.Status {
	case entity.AnalysisSucceeded, entity.AnalysisFailed:
		return &AnalyzeResult{Run: existing, Requirements: filtered, OwnedExecute: false}, nil
	case entity.AnalysisPending:
		if hasActiveAnalysisLease(existing, now) {
			return &AnalyzeResult{Run: existing, Requirements: filtered, OwnedExecute: false}, nil
		}
		if !analysisLeaseExpired(existing, now) {
			return &AnalyzeResult{Run: existing, Requirements: filtered, OwnedExecute: false}, nil
		}
		run, claimed, err := s.repo.ClaimExpiredPendingExecution(ctx, in.TenantID, existing.AnalysisRunID, existing.ExecutionGeneration, now)
		if err != nil {
			return nil, err
		}
		if !claimed {
			latest, gerr := s.repo.GetAnalysisRun(ctx, in.TenantID, existing.AnalysisRunID)
			if gerr != nil {
				return nil, gerr
			}
			reqs, lerr = s.repo.ListRequirementsByRevision(ctx, in.TenantID, in.BusinessID, in.BusinessModelRevision)
			if lerr != nil {
				return nil, lerr
			}
			return &AnalyzeResult{Run: latest, Requirements: filterByAnalysisRun(reqs, latest.AnalysisRunID), OwnedExecute: false}, nil
		}
		return s.executeAnalysis(ctx, run, sm, in.ActorID, run.ExecutionGeneration)
	default:
		return &AnalyzeResult{Run: existing, Requirements: filtered, OwnedExecute: false}, nil
	}
}

func filterByAnalysisRun(reqs []*entity.DataRequirement, runID string) []*entity.DataRequirement {
	out := make([]*entity.DataRequirement, 0)
	for _, r := range reqs {
		if r.AnalysisRunID == runID {
			out = append(out, r)
		}
	}
	return out
}

func (s *dataServiceImpl) executeAnalysis(ctx context.Context, run *entity.DataRequirementAnalysisRun, sm *businessentity.SemanticModel, actorID string, generation int32) (*AnalyzeResult, error) {
	resp, err := s.model.AnalyzeDataRequirements(ctx, &AnalyzeDataRequirementsRequest{
		RequestID:             run.AnalysisRunID,
		TenantID:              run.TenantID,
		BusinessID:            run.BusinessID,
		BusinessModelRevision: run.BusinessModelRevision,
		SemanticModel:         sm,
	})
	if err != nil {
		_ = s.repo.MarkAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, "FORMA_DATA_MODEL_FAILED", sanitizeError(err.Error()), generation)
		failed, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		return &AnalyzeResult{Run: failed, Requirements: nil, OwnedExecute: true}, entity.ErrModelFailed
	}
	if resp == nil {
		_ = s.repo.MarkAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, "FORMA_DATA_MODEL_FAILED", "empty model response", generation)
		failed, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		return &AnalyzeResult{Run: failed, Requirements: nil, OwnedExecute: true}, entity.ErrModelFailed
	}

	ids := collectElementIDs(sm)
	if err := validateProposals(resp.Proposals, ids); err != nil {
		key := "FORMA_DATA_REQUIREMENT_INVALID"
		if strings.Contains(err.Error(), "element ref") {
			key = "FORMA_DATA_BUSINESS_ELEMENT_REF_INVALID"
		}
		_ = s.repo.MarkAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, key, sanitizeError(err.Error()), generation)
		failed, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		return &AnalyzeResult{Run: failed, Requirements: nil, OwnedExecute: true}, err
	}

	now := time.Now().UTC()
	reqs := make([]*entity.DataRequirement, 0, len(resp.Proposals))
	for _, p := range resp.Proposals {
		reqs = append(reqs, &entity.DataRequirement{
			RequirementID:            newID("req"),
			TenantID:                 run.TenantID,
			BusinessID:               run.BusinessID,
			BusinessModelRevision:    run.BusinessModelRevision,
			RequirementKind:          p.RequirementKind,
			SemanticName:             p.SemanticName,
			Description:              p.Description,
			BusinessElementRefs:      append([]string{}, p.BusinessElementRefs...),
			Requiredness:             p.Requiredness,
			FreshnessRequirement:     p.FreshnessRequirement,
			AccessNeed:               p.AccessNeed,
			Status:                   entity.StatusProposed,
			Source:                   entity.SourceAIGenerated,
			DerivedFromRequirementID: "",
			AnalysisRunID:            run.AnalysisRunID,
			CreatedBy:                run.CreatedBy,
			CreatedAt:                now,
			UpdatedAt:                now,
		})
	}

	err = s.repo.Transaction(ctx, func(tx repository.DataRepository) error {
		if err := tx.CreateRequirementsBatch(ctx, reqs); err != nil {
			return err
		}
		modelRef := resp.ModelRef
		if modelRef == "" {
			modelRef = "unknown"
		}
		return tx.MarkAnalysisSucceeded(ctx, run.TenantID, run.AnalysisRunID, modelRef, generation)
	})
	if err != nil {
		_ = s.repo.MarkAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, "FORMA_INTERNAL", sanitizeError(err.Error()), generation)
		failed, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		return &AnalyzeResult{Run: failed, Requirements: nil, OwnedExecute: true}, err
	}

	final, err := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
	if err != nil {
		return nil, err
	}
	return &AnalyzeResult{Run: final, Requirements: reqs, OwnedExecute: true}, nil
}

func (s *dataServiceImpl) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error) {
	return s.repo.GetAnalysisRun(ctx, tenantID, analysisRunID)
}

func (s *dataServiceImpl) RetryFailedAnalysis(ctx context.Context, tenantID, analysisRunID, actorID string) (*AnalyzeResult, error) {
	run, err := s.repo.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		return nil, err
	}
	if run.Status != entity.AnalysisFailed {
		return nil, entity.ErrAnalysisNotFailed
	}
	claimed, generation, err := s.repo.ClaimAnalysisRetry(ctx, tenantID, analysisRunID, actorID)
	if err != nil {
		return nil, err
	}
	if !claimed {
		latest, gerr := s.repo.GetAnalysisRun(ctx, tenantID, analysisRunID)
		if gerr != nil {
			return nil, gerr
		}
		if latest.Status == entity.AnalysisPending || latest.Status == entity.AnalysisSucceeded {
			reqs, _ := s.repo.ListRequirementsByRevision(ctx, tenantID, latest.BusinessID, latest.BusinessModelRevision)
			return &AnalyzeResult{Run: latest, Requirements: filterByAnalysisRun(reqs, latest.AnalysisRunID), OwnedExecute: false}, nil
		}
		return nil, entity.ErrAnalysisNotFailed
	}
	run, err = s.repo.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		return nil, err
	}
	sm, err := s.loadPinnedSemantic(ctx, run.TenantID, run.BusinessID, run.BusinessModelRevision)
	if err != nil {
		_ = s.repo.MarkAnalysisFailed(ctx, tenantID, analysisRunID, "FORMA_DATA_BUSINESS_REVISION_NOT_FOUND", sanitizeError(err.Error()), generation)
		return nil, err
	}
	return s.executeAnalysis(ctx, run, sm, run.CreatedBy, generation)
}

func (s *dataServiceImpl) ListRequirements(ctx context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error) {
	return s.repo.ListRequirementsByRevision(ctx, tenantID, businessID, revision)
}

func (s *dataServiceImpl) GetRequirement(ctx context.Context, tenantID, requirementID string) (*entity.DataRequirement, error) {
	return s.repo.GetRequirement(ctx, tenantID, requirementID)
}

func (s *dataServiceImpl) CreateManualRequirement(ctx context.Context, in *ManualCreateInput) (*entity.DataRequirement, error) {
	if in == nil || in.TenantID == "" || in.BusinessID == "" {
		return nil, entity.ErrInvalidProposal
	}
	sm, err := s.loadPinnedSemantic(ctx, in.TenantID, in.BusinessID, in.BusinessModelRevision)
	if err != nil {
		return nil, err
	}
	ids := collectElementIDs(sm)
	for _, ref := range in.BusinessElementRefs {
		if _, ok := ids[ref]; !ok {
			return nil, entity.ErrBusinessElementRefInvalid
		}
	}
	if !validKind(in.RequirementKind) || strings.TrimSpace(in.SemanticName) == "" {
		return nil, entity.ErrInvalidProposal
	}
	now := time.Now().UTC()
	req := &entity.DataRequirement{
		RequirementID:            newID("req"),
		TenantID:                 in.TenantID,
		BusinessID:               in.BusinessID,
		BusinessModelRevision:    in.BusinessModelRevision,
		RequirementKind:          in.RequirementKind,
		SemanticName:             in.SemanticName,
		Description:              in.Description,
		BusinessElementRefs:      append([]string{}, in.BusinessElementRefs...),
		Requiredness:             in.Requiredness,
		FreshnessRequirement:     in.FreshnessRequirement,
		AccessNeed:               in.AccessNeed,
		Status:                   entity.StatusProposed,
		Source:                   entity.SourceManualCreated,
		DerivedFromRequirementID: "",
		AnalysisRunID:            "",
		CreatedBy:                in.ActorID,
		CreatedAt:                now,
		UpdatedAt:                now,
	}
	if err := s.repo.CreateRequirement(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *dataServiceImpl) ConfirmRequirement(ctx context.Context, tenantID, requirementID, actorID, reason string) (*entity.DataRequirement, *entity.DataRequirementDecision, error) {
	var outReq *entity.DataRequirement
	var outDec *entity.DataRequirementDecision
	err := s.repo.Transaction(ctx, func(tx repository.DataRepository) error {
		req, err := tx.GetRequirement(ctx, tenantID, requirementID)
		if err != nil {
			return err
		}
		if req.Status != entity.StatusProposed {
			return entity.ErrRequirementInvalidState
		}
		ok, err := tx.UpdateRequirementStatusCAS(ctx, tenantID, requirementID, entity.StatusProposed, entity.StatusConfirmed)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrRequirementAlreadyDecided
		}
		now := time.Now().UTC()
		dec := &entity.DataRequirementDecision{
			DecisionID:            newID("ddec"),
			TenantID:              tenantID,
			BusinessID:            req.BusinessID,
			SourceRequirementID:   requirementID,
			TargetRequirementID:   "",
			Action:                entity.DecisionConfirm,
			ActorPrincipalID:      actorID,
			Reason:                reason,
			BusinessModelRevision: req.BusinessModelRevision,
			CreatedAt:             now,
		}
		if err := tx.CreateDecision(ctx, dec); err != nil {
			return err
		}
		req.Status = entity.StatusConfirmed
		req.UpdatedAt = now
		outReq = req
		outDec = dec
		return nil
	})
	return outReq, outDec, err
}

func (s *dataServiceImpl) RejectRequirement(ctx context.Context, tenantID, requirementID, actorID, reason string) (*entity.DataRequirement, *entity.DataRequirementDecision, error) {
	var outReq *entity.DataRequirement
	var outDec *entity.DataRequirementDecision
	err := s.repo.Transaction(ctx, func(tx repository.DataRepository) error {
		req, err := tx.GetRequirement(ctx, tenantID, requirementID)
		if err != nil {
			return err
		}
		if req.Status != entity.StatusProposed {
			return entity.ErrRequirementInvalidState
		}
		ok, err := tx.UpdateRequirementStatusCAS(ctx, tenantID, requirementID, entity.StatusProposed, entity.StatusRejected)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrRequirementAlreadyDecided
		}
		now := time.Now().UTC()
		dec := &entity.DataRequirementDecision{
			DecisionID:            newID("ddec"),
			TenantID:              tenantID,
			BusinessID:            req.BusinessID,
			SourceRequirementID:   requirementID,
			Action:                entity.DecisionReject,
			ActorPrincipalID:      actorID,
			Reason:                reason,
			BusinessModelRevision: req.BusinessModelRevision,
			CreatedAt:             now,
		}
		if err := tx.CreateDecision(ctx, dec); err != nil {
			return err
		}
		req.Status = entity.StatusRejected
		req.UpdatedAt = now
		outReq = req
		outDec = dec
		return nil
	})
	return outReq, outDec, err
}

func (s *dataServiceImpl) EditConfirmRequirement(ctx context.Context, in *EditConfirmInput) (*entity.DataRequirement, *entity.DataRequirement, *entity.DataRequirementDecision, error) {
	if in == nil {
		return nil, nil, nil, entity.ErrInvalidProposal
	}
	var original, replacement *entity.DataRequirement
	var decision *entity.DataRequirementDecision
	err := s.repo.Transaction(ctx, func(tx repository.DataRepository) error {
		src, err := tx.GetRequirement(ctx, in.TenantID, in.SourceRequirementID)
		if err != nil {
			return err
		}
		if src.Status != entity.StatusProposed {
			return entity.ErrRequirementInvalidState
		}
		sm, err := s.loadPinnedSemantic(ctx, src.TenantID, src.BusinessID, src.BusinessModelRevision)
		if err != nil {
			return err
		}
		ids := collectElementIDs(sm)
		kind := in.RequirementKind
		if kind == "" {
			kind = src.RequirementKind
		}
		if !validKind(kind) {
			return entity.ErrInvalidProposal
		}
		name := in.SemanticName
		if name == "" {
			name = src.SemanticName
		}
		desc := in.Description
		if desc == "" {
			desc = src.Description
		}
		refs := in.BusinessElementRefs
		if len(refs) == 0 {
			refs = append([]string{}, src.BusinessElementRefs...)
		}
		for _, ref := range refs {
			if _, ok := ids[ref]; !ok {
				return entity.ErrBusinessElementRefInvalid
			}
		}

		ok, err := tx.UpdateRequirementStatusCAS(ctx, in.TenantID, in.SourceRequirementID, entity.StatusProposed, entity.StatusSuperseded)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrRequirementAlreadyDecided
		}

		now := time.Now().UTC()
		rep := &entity.DataRequirement{
			RequirementID:            newID("req"),
			TenantID:                 src.TenantID,
			BusinessID:               src.BusinessID,
			BusinessModelRevision:    src.BusinessModelRevision,
			RequirementKind:          kind,
			SemanticName:             name,
			Description:              desc,
			BusinessElementRefs:      refs,
			Requiredness:             firstNonEmpty(in.Requiredness, src.Requiredness),
			FreshnessRequirement:     firstNonEmpty(in.FreshnessRequirement, src.FreshnessRequirement),
			AccessNeed:               firstNonEmpty(in.AccessNeed, src.AccessNeed),
			Status:                   entity.StatusConfirmed,
			Source:                   entity.SourceManualModified,
			DerivedFromRequirementID: src.RequirementID,
			AnalysisRunID:            src.AnalysisRunID,
			CreatedBy:                in.ActorID,
			CreatedAt:                now,
			UpdatedAt:                now,
		}
		if err := tx.CreateRequirement(ctx, rep); err != nil {
			return err
		}
		dec := &entity.DataRequirementDecision{
			DecisionID:            newID("ddec"),
			TenantID:              src.TenantID,
			BusinessID:            src.BusinessID,
			SourceRequirementID:   src.RequirementID,
			TargetRequirementID:   rep.RequirementID,
			Action:                entity.DecisionEditConfirm,
			ActorPrincipalID:      in.ActorID,
			Reason:                in.Reason,
			BusinessModelRevision: src.BusinessModelRevision,
			CreatedAt:             now,
		}
		if err := tx.CreateDecision(ctx, dec); err != nil {
			return err
		}
		src.Status = entity.StatusSuperseded
		src.UpdatedAt = now
		original = src
		replacement = rep
		decision = dec
		return nil
	})
	return original, replacement, decision, err
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func (s *dataServiceImpl) ListDecisions(ctx context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error) {
	return s.repo.ListDecisionsByRequirement(ctx, tenantID, requirementID)
}

// DigestJSON is exported for tests needing payload digests.
func DigestJSON(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
