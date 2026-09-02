/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

type MappingService interface {
	AnalyzeSemanticMappings(context.Context, *AnalyzeSemanticMappingsInput) (*AnalyzeSemanticMappingsResult, error)
	GetMappingAnalysisRun(context.Context, string, string) (*entity.SemanticMappingAnalysisRun, error)
	RetryFailedMappingAnalysis(context.Context, string, string, string) (*AnalyzeSemanticMappingsResult, error)
	ListMappings(context.Context, string, string, int32, entity.MappingStatus) ([]*entity.SemanticMapping, error)
	GetMapping(context.Context, string, string) (*entity.SemanticMapping, error)
	CreateManualMapping(context.Context, *ManualMappingInput) (*entity.SemanticMapping, error)
	ConfirmMapping(context.Context, string, string, string, string) (*entity.SemanticMapping, *entity.SemanticMappingDecision, error)
	RejectMapping(context.Context, string, string, string, string) (*entity.SemanticMapping, *entity.SemanticMappingDecision, error)
	EditConfirmMapping(context.Context, *EditConfirmMappingInput) (*entity.SemanticMapping, *entity.SemanticMapping, *entity.SemanticMappingDecision, error)
	ListMappingDecisions(context.Context, string, string) ([]*entity.SemanticMappingDecision, error)
	GetMappingCoverage(context.Context, string, string, int32) (*MappingCoverage, error)
}

type MappingComponents struct {
	MappingRepo    repository.MappingRepository
	DataRepo       repository.DataRepository
	DataSourceRepo repository.DataSourceRepository
	BusinessSVC    businesssvc.BusinessService
	Model          FormaDataModel
}
type mappingService struct {
	mappings repository.MappingRepository
	data     repository.DataRepository
	sources  repository.DataSourceRepository
	business businesssvc.BusinessService
	model    FormaDataModel
}

func NewMappingService(c *MappingComponents) MappingService {
	if c == nil {
		return &mappingService{}
	}
	return &mappingService{mappings: c.MappingRepo, data: c.DataRepo, sources: c.DataSourceRepo, business: c.BusinessSVC, model: c.Model}
}

type AnalyzeSemanticMappingsInput struct {
	TenantID              string   `json:"tenant_id"`
	BusinessID            string   `json:"business_id"`
	BusinessModelRevision int32    `json:"business_model_revision"`
	RequirementIDs        []string `json:"requirement_ids"`
	SchemaSnapshotIDs     []string `json:"schema_snapshot_ids"`
	ClientRequestID       string   `json:"client_request_id"`
	ActorID               string   `json:"actor_id"`
}
type AnalyzeSemanticMappingsResult struct {
	Run          *entity.SemanticMappingAnalysisRun
	Mappings     []*entity.SemanticMapping
	OwnedExecute bool
}
type ManualMappingInput struct {
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	RequirementID         string
	SourceID              string
	ConnectionID          string
	AssetID               string
	SchemaSnapshotID      string
	TargetFieldPaths      []string
	MappingType           entity.MappingType
	TransformSpec         json.RawMessage
	Confidence            float64
	Reason                string
	ActorID               string
}
type EditConfirmMappingInput struct {
	TenantID         string
	SourceMappingID  string
	SourceID         string
	ConnectionID     string
	AssetID          string
	SchemaSnapshotID string
	TargetFieldPaths []string
	MappingType      entity.MappingType
	TransformSpec    json.RawMessage
	Confidence       float64
	Reason           string
	ActorID          string
}
type MappingCoverage struct {
	TotalConfirmedRequirements int      `json:"total_confirmed_requirements"`
	ConfirmedMappings          int      `json:"confirmed_mappings"`
	UnmappedRequirementIDs     []string `json:"unmapped_requirement_ids"`
	Coverage                   float64  `json:"coverage"`
}

func (s *mappingService) configured() bool {
	return s.mappings != nil && s.data != nil && s.sources != nil && s.model != nil
}
func mappingRequestDigest(in *AnalyzeSemanticMappingsInput) string {
	type digestInput struct {
		TenantID, BusinessID              string
		BusinessModelRevision             int32
		RequirementIDs, SchemaSnapshotIDs []string
		ClientRequestID                   string
	}
	return DigestJSON(digestInput{in.TenantID, in.BusinessID, in.BusinessModelRevision, in.RequirementIDs, in.SchemaSnapshotIDs, in.ClientRequestID})
}
func (s *mappingService) AnalyzeSemanticMappings(ctx context.Context, in *AnalyzeSemanticMappingsInput) (*AnalyzeSemanticMappingsResult, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || in.TenantID == "" || in.BusinessID == "" || in.BusinessModelRevision <= 0 || in.ClientRequestID == "" || len(in.RequirementIDs) == 0 || len(in.SchemaSnapshotIDs) == 0 {
		return nil, entity.ErrMappingTargetInvalid
	}
	raw, _ := json.Marshal(in)
	now := time.Now().UTC()
	exp := now.Add(5 * time.Minute)
	run := &entity.SemanticMappingAnalysisRun{AnalysisRunID: newID("marun"), TenantID: in.TenantID, BusinessID: in.BusinessID, BusinessModelRevision: in.BusinessModelRevision, ClientRequestID: in.ClientRequestID, RequestDigest: mappingRequestDigest(in), RequestJSON: string(raw), Status: entity.AnalysisPending, ExecutionGeneration: 1, ExecutionClaimedAt: &now, LeaseExpiresAt: &exp, CreatedBy: in.ActorID, StartedAt: &now, CreatedAt: now, UpdatedAt: now}
	existing, created, err := s.mappings.CreateOrClaimMappingAnalysisRun(ctx, run)
	if err != nil {
		return nil, err
	}
	if !created {
		maps, _ := s.mappings.ListMappings(ctx, in.TenantID, in.BusinessID, in.BusinessModelRevision, "")
		out := filterMappingsByRun(maps, existing.AnalysisRunID)
		if existing.Status == entity.AnalysisPending && existing.LeaseExpiresAt != nil && !time.Now().UTC().Before(*existing.LeaseExpiresAt) {
			claimed, owned, e := s.mappings.ClaimExpiredMappingAnalysis(ctx, in.TenantID, existing.AnalysisRunID, existing.ExecutionGeneration, time.Now().UTC())
			if e != nil {
				return nil, e
			}
			if owned {
				return s.executeMappingAnalysis(ctx, claimed, in)
			}
		}
		return &AnalyzeSemanticMappingsResult{Run: existing, Mappings: out}, nil
	}
	return s.executeMappingAnalysis(ctx, existing, in)
}
func filterMappingsByRun(in []*entity.SemanticMapping, id string) []*entity.SemanticMapping {
	out := []*entity.SemanticMapping{}
	for _, v := range in {
		if v.AnalysisRunID == id {
			out = append(out, v)
		}
	}
	return out
}
func (s *mappingService) loadMappingInputs(ctx context.Context, in *AnalyzeSemanticMappingsInput) ([]MappingRequirementMetadata, []NormalizedSchemaSnapshot, error) {
	reqs := make([]MappingRequirementMetadata, 0, len(in.RequirementIDs))
	seen := map[string]bool{}
	for _, id := range in.RequirementIDs {
		if seen[id] {
			return nil, nil, entity.ErrMappingTargetInvalid
		}
		seen[id] = true
		r, err := s.data.GetRequirement(ctx, in.TenantID, id)
		if err != nil || r.Status != entity.StatusConfirmed || r.BusinessID != in.BusinessID || r.BusinessModelRevision != in.BusinessModelRevision {
			return nil, nil, entity.ErrMappingRequirementNotConfirmed
		}
		reqs = append(reqs, MappingRequirementMetadata{r.RequirementID, r.RequirementKind, r.SemanticName, r.Description, r.Requiredness, r.AccessNeed})
	}
	snaps := make([]NormalizedSchemaSnapshot, 0, len(in.SchemaSnapshotIDs))
	for _, id := range in.SchemaSnapshotIDs {
		snap, err := s.sources.GetSnapshot(ctx, in.TenantID, id)
		if err != nil {
			return nil, nil, entity.ErrMappingLineageInvalid
		}
		asset, err := s.sources.GetAsset(ctx, in.TenantID, snap.AssetID)
		if err != nil {
			return nil, nil, entity.ErrMappingLineageInvalid
		}
		conn, err := s.sources.GetConnection(ctx, in.TenantID, snap.ConnectionID)
		if err != nil {
			return nil, nil, entity.ErrMappingLineageInvalid
		}
		source, err := s.sources.GetSource(ctx, in.TenantID, snap.SourceID)
		if err != nil {
			return nil, nil, entity.ErrMappingLineageInvalid
		}
		if asset.SourceID != source.SourceID || asset.ConnectionID != conn.ConnectionID || conn.SourceID != source.SourceID || snap.SourceID != source.SourceID || snap.ConnectionID != conn.ConnectionID {
			return nil, nil, entity.ErrMappingLineageInvalid
		}
		var schema entity.PhysicalSchema
		if json.Unmarshal([]byte(snap.SchemaJSON), &schema) != nil {
			return nil, nil, entity.ErrMappingTargetInvalid
		}
		snaps = append(snaps, NormalizedSchemaSnapshot{snap.SnapshotID, snap.SourceID, snap.ConnectionID, snap.AssetID, schema})
	}
	return reqs, snaps, nil
}
func (s *mappingService) executeMappingAnalysis(ctx context.Context, run *entity.SemanticMappingAnalysisRun, in *AnalyzeSemanticMappingsInput) (*AnalyzeSemanticMappingsResult, error) {
	reqs, snaps, err := s.loadMappingInputs(ctx, in)
	if err != nil {
		return s.failMappingRun(ctx, run, err)
	}
	resp, err := s.model.SuggestSemanticMappings(ctx, &SuggestSemanticMappingsRequest{RequestID: run.AnalysisRunID, TenantID: run.TenantID, BusinessID: run.BusinessID, BusinessModelRevision: run.BusinessModelRevision, RequirementIDs: append([]string(nil), in.RequirementIDs...), Requirements: reqs, SchemaSnapshots: snaps})
	if err != nil || resp == nil {
		if err == nil {
			err = errors.New("empty model response")
		}
		return s.failMappingRun(ctx, run, fmt.Errorf("%w: %v", entity.ErrModelFailed, err))
	}
	reqSet := map[string]bool{}
	for _, v := range in.RequirementIDs {
		reqSet[v] = true
	}
	snapSet := map[string]NormalizedSchemaSnapshot{}
	for _, v := range snaps {
		snapSet[v.SchemaSnapshotID] = v
	}
	now := time.Now().UTC()
	maps := make([]*entity.SemanticMapping, 0, len(resp.Proposals))
	for i, p := range resp.Proposals {
		snap, ok := snapSet[p.SchemaSnapshotID]
		if !ok || !reqSet[p.RequirementID] || p.SourceID != snap.SourceID || p.ConnectionID != snap.ConnectionID || p.AssetID != snap.AssetID {
			return s.failMappingRun(ctx, run, fmt.Errorf("%w: proposal[%d]", entity.ErrMappingLineageInvalid, i))
		}
		if p.Status != "" && p.Status != entity.MappingStatusProposed {
			return s.failMappingRun(ctx, run, entity.ErrMappingInvalidState)
		}
		if p.Source != "" && p.Source != entity.MappingSourceAIGenerated {
			return s.failMappingRun(ctx, run, entity.ErrMappingInvalidState)
		}
		if err := ValidateMappingTarget(p.TargetFieldPaths, &snap.Schema); err != nil {
			return s.failMappingRun(ctx, run, err)
		}
		if err := ValidateTransformSpec(p.MappingType, p.TransformSpec); err != nil {
			return s.failMappingRun(ctx, run, err)
		}
		if p.MappingType == entity.MappingTypeJoinRef {
			if err := ValidateJoinRef(p.TransformSpec, &snap.Schema); err != nil {
				return s.failMappingRun(ctx, run, err)
			}
		}
		maps = append(maps, &entity.SemanticMapping{MappingID: newID("map"), TenantID: run.TenantID, BusinessID: run.BusinessID, BusinessModelRevision: run.BusinessModelRevision, RequirementID: p.RequirementID, SourceID: p.SourceID, ConnectionID: p.ConnectionID, AssetID: p.AssetID, SchemaSnapshotID: p.SchemaSnapshotID, TargetFieldPaths: append([]string(nil), p.TargetFieldPaths...), MappingType: p.MappingType, TransformSpec: append([]byte(nil), p.TransformSpec...), Status: entity.MappingStatusProposed, Source: entity.MappingSourceAIGenerated, Confidence: p.Confidence, Reason: p.Reason, AnalysisRunID: run.AnalysisRunID, CreatedBy: run.CreatedBy, CreatedAt: now, UpdatedAt: now})
	}
	modelRef := resp.ModelRef
	if modelRef == "" {
		modelRef = "unknown"
	}
	err = s.mappings.Transaction(ctx, func(tx repository.MappingRepository) error {
		if err := tx.CreateMappingsBatch(ctx, maps); err != nil {
			return err
		}
		return tx.MarkMappingAnalysisSucceeded(ctx, run.TenantID, run.AnalysisRunID, modelRef, run.ExecutionGeneration)
	})
	if err != nil {
		return s.failMappingRun(ctx, run, err)
	}
	final, _ := s.mappings.GetMappingAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
	return &AnalyzeSemanticMappingsResult{Run: final, Mappings: maps, OwnedExecute: true}, nil
}
func (s *mappingService) failMappingRun(ctx context.Context, run *entity.SemanticMappingAnalysisRun, err error) (*AnalyzeSemanticMappingsResult, error) {
	key := "FORMA_DATA_SEMANTIC_MAPPING_INVALID"
	if errors.Is(err, entity.ErrModelFailed) {
		key = "FORMA_DATA_MODEL_FAILED"
	}
	_ = s.mappings.MarkMappingAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, key, sanitizeError(err.Error()), run.ExecutionGeneration)
	failed, _ := s.mappings.GetMappingAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
	return &AnalyzeSemanticMappingsResult{Run: failed, OwnedExecute: true}, err
}
func (s *mappingService) GetMappingAnalysisRun(ctx context.Context, t, id string) (*entity.SemanticMappingAnalysisRun, error) {
	return s.mappings.GetMappingAnalysisRun(ctx, t, id)
}
func (s *mappingService) RetryFailedMappingAnalysis(ctx context.Context, t, id, actor string) (*AnalyzeSemanticMappingsResult, error) {
	run, err := s.mappings.GetMappingAnalysisRun(ctx, t, id)
	if err != nil {
		return nil, err
	}
	if run.Status != entity.AnalysisFailed {
		return nil, entity.ErrMappingAnalysisNotFailed
	}
	ok, g, err := s.mappings.ClaimMappingAnalysisRetry(ctx, t, id, actor)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, entity.ErrMappingAnalysisNotFailed
	}
	run, _ = s.mappings.GetMappingAnalysisRun(ctx, t, id)
	run.ExecutionGeneration = g
	var in AnalyzeSemanticMappingsInput
	if json.Unmarshal([]byte(run.RequestJSON), &in) != nil {
		return s.failMappingRun(ctx, run, entity.ErrMappingTargetInvalid)
	}
	return s.executeMappingAnalysis(ctx, run, &in)
}
func (s *mappingService) ListMappings(ctx context.Context, t, b string, r int32, st entity.MappingStatus) ([]*entity.SemanticMapping, error) {
	return s.mappings.ListMappings(ctx, t, b, r, st)
}
func (s *mappingService) GetMapping(ctx context.Context, t, id string) (*entity.SemanticMapping, error) {
	return s.mappings.GetMapping(ctx, t, id)
}
func (s *mappingService) validateManual(ctx context.Context, in *ManualMappingInput) (*entity.PhysicalSchema, error) {
	req, err := s.data.GetRequirement(ctx, in.TenantID, in.RequirementID)
	if err != nil || req.Status != entity.StatusConfirmed || req.BusinessID != in.BusinessID || req.BusinessModelRevision != in.BusinessModelRevision {
		return nil, entity.ErrMappingRequirementNotConfirmed
	}
	snap, err := s.sources.GetSnapshot(ctx, in.TenantID, in.SchemaSnapshotID)
	if err != nil {
		return nil, entity.ErrMappingLineageInvalid
	}
	if snap.SourceID != in.SourceID || snap.ConnectionID != in.ConnectionID || snap.AssetID != in.AssetID {
		return nil, entity.ErrMappingLineageInvalid
	}
	asset, ae := s.sources.GetAsset(ctx, in.TenantID, in.AssetID)
	conn, ce := s.sources.GetConnection(ctx, in.TenantID, in.ConnectionID)
	source, se := s.sources.GetSource(ctx, in.TenantID, in.SourceID)
	if ae != nil || ce != nil || se != nil || asset.SourceID != source.SourceID || asset.ConnectionID != conn.ConnectionID || conn.SourceID != source.SourceID {
		return nil, entity.ErrMappingLineageInvalid
	}
	var schema entity.PhysicalSchema
	if json.Unmarshal([]byte(snap.SchemaJSON), &schema) != nil {
		return nil, entity.ErrMappingTargetInvalid
	}
	if err := ValidateMappingTarget(in.TargetFieldPaths, &schema); err != nil {
		return nil, err
	}
	if err := ValidateTransformSpec(in.MappingType, in.TransformSpec); err != nil {
		return nil, err
	}
	if in.MappingType == entity.MappingTypeJoinRef {
		if err := ValidateJoinRef(in.TransformSpec, &schema); err != nil {
			return nil, err
		}
	}
	return &schema, nil
}
func (s *mappingService) CreateManualMapping(ctx context.Context, in *ManualMappingInput) (*entity.SemanticMapping, error) {
	if in == nil {
		return nil, entity.ErrMappingTargetInvalid
	}
	if _, err := s.validateManual(ctx, in); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	m := &entity.SemanticMapping{MappingID: newID("map"), TenantID: in.TenantID, BusinessID: in.BusinessID, BusinessModelRevision: in.BusinessModelRevision, RequirementID: in.RequirementID, SourceID: in.SourceID, ConnectionID: in.ConnectionID, AssetID: in.AssetID, SchemaSnapshotID: in.SchemaSnapshotID, TargetFieldPaths: append([]string(nil), in.TargetFieldPaths...), MappingType: in.MappingType, TransformSpec: append([]byte(nil), in.TransformSpec...), Status: entity.MappingStatusProposed, Source: entity.MappingSourceManualCreated, Confidence: in.Confidence, Reason: in.Reason, CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now}
	if err := s.mappings.CreateMapping(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}
func (s *mappingService) decide(ctx context.Context, t, id, actor, reason string, action DecisionActionAlias) (*entity.SemanticMapping, *entity.SemanticMappingDecision, error) {
	var out *entity.SemanticMapping
	var dec *entity.SemanticMappingDecision
	err := s.mappings.Transaction(ctx, func(tx repository.MappingRepository) error {
		m, err := tx.GetMapping(ctx, t, id)
		if err != nil {
			return err
		}
		if m.Status != entity.MappingStatusProposed {
			return entity.ErrMappingInvalidState
		}
		if action == DecisionActionAlias(entity.DecisionConfirm) {
			if _, err := tx.GetConfirmedMappingByRequirement(ctx, t, m.BusinessID, m.BusinessModelRevision, m.RequirementID); err == nil {
				return entity.ErrMappingAlreadyConfirmed
			} else if !errors.Is(err, entity.ErrMappingNotFound) {
				return err
			}
		}
		to := entity.MappingStatusRejected
		if action == DecisionActionAlias(entity.DecisionConfirm) {
			to = entity.MappingStatusConfirmed
		}
		ok, err := tx.UpdateMappingStatusCAS(ctx, t, id, entity.MappingStatusProposed, to)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrMappingAlreadyDecided
		}
		now := time.Now().UTC()
		d := &entity.SemanticMappingDecision{DecisionID: newID("mdec"), TenantID: t, BusinessID: m.BusinessID, SourceMappingID: id, Action: entity.DecisionAction(action), ActorPrincipalID: actor, Reason: reason, BusinessModelRevision: m.BusinessModelRevision, CreatedAt: now}
		if err := tx.CreateMappingDecision(ctx, d); err != nil {
			return err
		}
		m.Status = to
		m.UpdatedAt = now
		out = m
		dec = d
		return nil
	})
	return out, dec, err
}

type DecisionActionAlias entity.DecisionAction

func (s *mappingService) ConfirmMapping(ctx context.Context, t, id, a, r string) (*entity.SemanticMapping, *entity.SemanticMappingDecision, error) {
	return s.decide(ctx, t, id, a, r, DecisionActionAlias(entity.DecisionConfirm))
}
func (s *mappingService) RejectMapping(ctx context.Context, t, id, a, r string) (*entity.SemanticMapping, *entity.SemanticMappingDecision, error) {
	return s.decide(ctx, t, id, a, r, DecisionActionAlias(entity.DecisionReject))
}
func (s *mappingService) EditConfirmMapping(ctx context.Context, in *EditConfirmMappingInput) (*entity.SemanticMapping, *entity.SemanticMapping, *entity.SemanticMappingDecision, error) {
	if in == nil {
		return nil, nil, nil, entity.ErrMappingTargetInvalid
	}
	var src, rep *entity.SemanticMapping
	var dec *entity.SemanticMappingDecision
	err := s.mappings.Transaction(ctx, func(tx repository.MappingRepository) error {
		old, err := tx.GetMapping(ctx, in.TenantID, in.SourceMappingID)
		if err != nil {
			return err
		}
		if old.Status != entity.MappingStatusProposed {
			return entity.ErrMappingInvalidState
		}
		if _, err := tx.GetConfirmedMappingByRequirement(ctx, old.TenantID, old.BusinessID, old.BusinessModelRevision, old.RequirementID); err == nil {
			return entity.ErrMappingAlreadyConfirmed
		} else if !errors.Is(err, entity.ErrMappingNotFound) {
			return err
		}
		manual := &ManualMappingInput{TenantID: old.TenantID, BusinessID: old.BusinessID, BusinessModelRevision: old.BusinessModelRevision, RequirementID: old.RequirementID, SourceID: firstNonEmpty(in.SourceID, old.SourceID), ConnectionID: firstNonEmpty(in.ConnectionID, old.ConnectionID), AssetID: firstNonEmpty(in.AssetID, old.AssetID), SchemaSnapshotID: firstNonEmpty(in.SchemaSnapshotID, old.SchemaSnapshotID), TargetFieldPaths: in.TargetFieldPaths, MappingType: in.MappingType, TransformSpec: in.TransformSpec}
		if len(manual.TargetFieldPaths) == 0 {
			manual.TargetFieldPaths = old.TargetFieldPaths
		}
		if manual.MappingType == "" {
			manual.MappingType = old.MappingType
		}
		if len(manual.TransformSpec) == 0 {
			manual.TransformSpec = old.TransformSpec
		}
		if _, err := s.validateManual(ctx, manual); err != nil {
			return err
		}
		ok, err := tx.UpdateMappingStatusCAS(ctx, in.TenantID, old.MappingID, entity.MappingStatusProposed, entity.MappingStatusSuperseded)
		if err != nil || !ok {
			if err != nil {
				return err
			}
			return entity.ErrMappingAlreadyDecided
		}
		now := time.Now().UTC()
		replacement := &entity.SemanticMapping{MappingID: newID("map"), TenantID: old.TenantID, BusinessID: old.BusinessID, BusinessModelRevision: old.BusinessModelRevision, RequirementID: old.RequirementID, SourceID: manual.SourceID, ConnectionID: manual.ConnectionID, AssetID: manual.AssetID, SchemaSnapshotID: manual.SchemaSnapshotID, TargetFieldPaths: append([]string(nil), manual.TargetFieldPaths...), MappingType: manual.MappingType, TransformSpec: append([]byte(nil), manual.TransformSpec...), Status: entity.MappingStatusConfirmed, Source: entity.MappingSourceManualModified, Confidence: in.Confidence, Reason: in.Reason, DerivedFromMappingID: old.MappingID, AnalysisRunID: old.AnalysisRunID, CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now}
		if err := tx.CreateMapping(ctx, replacement); err != nil {
			return err
		}
		d := &entity.SemanticMappingDecision{DecisionID: newID("mdec"), TenantID: old.TenantID, BusinessID: old.BusinessID, SourceMappingID: old.MappingID, TargetMappingID: replacement.MappingID, Action: entity.DecisionEditConfirm, ActorPrincipalID: in.ActorID, Reason: in.Reason, BusinessModelRevision: old.BusinessModelRevision, CreatedAt: now}
		if err := tx.CreateMappingDecision(ctx, d); err != nil {
			return err
		}
		old.Status = entity.MappingStatusSuperseded
		src = old
		rep = replacement
		dec = d
		return nil
	})
	return src, rep, dec, err
}
func (s *mappingService) ListMappingDecisions(ctx context.Context, t, id string) ([]*entity.SemanticMappingDecision, error) {
	return s.mappings.ListMappingDecisions(ctx, t, id)
}
func (s *mappingService) GetMappingCoverage(ctx context.Context, t, b string, rev int32) (*MappingCoverage, error) {
	reqs, err := s.data.ListRequirementsByRevision(ctx, t, b, rev)
	if err != nil {
		return nil, err
	}
	maps, err := s.mappings.ListMappings(ctx, t, b, rev, entity.MappingStatusConfirmed)
	if err != nil {
		return nil, err
	}
	confirmed := map[string]bool{}
	for _, m := range maps {
		confirmed[m.RequirementID] = true
	}
	out := &MappingCoverage{}
	for _, r := range reqs {
		if r.Status == entity.StatusConfirmed {
			out.TotalConfirmedRequirements++
			if confirmed[r.RequirementID] {
				out.ConfirmedMappings++
			} else {
				out.UnmappedRequirementIDs = append(out.UnmappedRequirementIDs, r.RequirementID)
			}
		}
	}
	if out.TotalConfirmedRequirements > 0 {
		out.Coverage = float64(out.ConfirmedMappings) / float64(out.TotalConfirmedRequirements)
	}
	return out, nil
}

var _ = strings.TrimSpace
