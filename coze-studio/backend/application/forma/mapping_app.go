/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

type AnalyzeSemanticMappingsInput struct {
	BusinessModelRevision int32    `json:"business_model_revision"`
	RequirementIDs        []string `json:"requirement_ids"`
	SchemaSnapshotIDs     []string `json:"schema_snapshot_ids"`
	ClientRequestID       string   `json:"client_request_id"`
}
type SemanticMappingInput struct {
	BusinessModelRevision int32           `json:"business_model_revision"`
	RequirementID         string          `json:"requirement_id"`
	SourceID              string          `json:"source_id"`
	ConnectionID          string          `json:"connection_id"`
	AssetID               string          `json:"asset_id"`
	SchemaSnapshotID      string          `json:"schema_snapshot_id"`
	TargetFieldPaths      []string        `json:"target_field_paths"`
	MappingType           string          `json:"mapping_type"`
	TransformSpec         json.RawMessage `json:"transform_spec"`
	Confidence            float64         `json:"confidence"`
	Reason                string          `json:"reason"`
}
type EditConfirmSemanticMappingInput struct {
	SourceID         string          `json:"source_id"`
	ConnectionID     string          `json:"connection_id"`
	AssetID          string          `json:"asset_id"`
	SchemaSnapshotID string          `json:"schema_snapshot_id"`
	TargetFieldPaths []string        `json:"target_field_paths"`
	MappingType      string          `json:"mapping_type"`
	TransformSpec    json.RawMessage `json:"transform_spec"`
	Confidence       float64         `json:"confidence"`
	Reason           string          `json:"reason"`
}
type SemanticMappingDTO struct {
	MappingID             string          `json:"mapping_id"`
	BusinessID            string          `json:"business_id"`
	BusinessModelRevision int32           `json:"business_model_revision"`
	RequirementID         string          `json:"requirement_id"`
	SourceID              string          `json:"source_id"`
	ConnectionID          string          `json:"connection_id"`
	AssetID               string          `json:"asset_id"`
	SchemaSnapshotID      string          `json:"schema_snapshot_id"`
	TargetFieldPaths      []string        `json:"target_field_paths"`
	MappingType           string          `json:"mapping_type"`
	TransformSpec         json.RawMessage `json:"transform_spec"`
	Status                string          `json:"status"`
	Source                string          `json:"source"`
	Confidence            float64         `json:"confidence"`
	Reason                string          `json:"reason"`
	DerivedFromMappingID  string          `json:"derived_from_mapping_id,omitempty"`
	AnalysisRunID         string          `json:"analysis_run_id,omitempty"`
	CreatedBy             string          `json:"created_by"`
	CreatedAt             string          `json:"created_at"`
	UpdatedAt             string          `json:"updated_at"`
}
type MappingAnalysisRunDTO struct {
	AnalysisRunID         string `json:"analysis_run_id"`
	BusinessID            string `json:"business_id"`
	BusinessModelRevision int32  `json:"business_model_revision"`
	ClientRequestID       string `json:"client_request_id"`
	Status                string `json:"status"`
	ModelRef              string `json:"model_ref,omitempty"`
	ErrorKey              string `json:"error_key,omitempty"`
	ErrorMessage          string `json:"error_message_sanitized,omitempty"`
	RetryCount            int32  `json:"retry_count"`
	CreatedBy             string `json:"created_by"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}
type MappingDecisionDTO struct {
	DecisionID            string `json:"decision_id"`
	BusinessID            string `json:"business_id"`
	SourceMappingID       string `json:"source_mapping_id"`
	TargetMappingID       string `json:"target_mapping_id,omitempty"`
	Action                string `json:"action"`
	ActorPrincipalID      string `json:"actor_principal_id"`
	Reason                string `json:"reason"`
	BusinessModelRevision int32  `json:"business_model_revision"`
	CreatedAt             string `json:"created_at"`
}
type AnalyzeSemanticMappingsResponse struct {
	AnalysisRun  *MappingAnalysisRunDTO `json:"analysis_run"`
	Mappings     []*SemanticMappingDTO  `json:"mappings"`
	OwnedExecute bool                   `json:"owned_execute"`
}
type SemanticMappingDecisionResponse struct {
	Mapping  *SemanticMappingDTO `json:"mapping"`
	Decision *MappingDecisionDTO `json:"decision"`
}
type EditConfirmSemanticMappingResponse struct {
	Original    *SemanticMappingDTO `json:"original"`
	Replacement *SemanticMappingDTO `json:"replacement"`
	Decision    *MappingDecisionDTO `json:"decision"`
}

func (s *ApplicationService) AnalyzeSemanticMappings(ctx context.Context, b string, in *AnalyzeSemanticMappingsInput) (*AnalyzeSemanticMappingsResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if s.MappingSVC == nil || in == nil {
		return nil, formaerrors.DataNotConfigured("mapping service not initialized")
	}
	r, err := s.MappingSVC.AnalyzeSemanticMappings(ctx, &datasvc.AnalyzeSemanticMappingsInput{TenantID: tc.TenantID, BusinessID: b, BusinessModelRevision: in.BusinessModelRevision, RequirementIDs: in.RequirementIDs, SchemaSnapshotIDs: in.SchemaSnapshotIDs, ClientRequestID: in.ClientRequestID, ActorID: tc.PrincipalID})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return mappingAnalysisResponse(r), nil
}
func (s *ApplicationService) GetMappingAnalysisRun(ctx context.Context, b, id string) (*MappingAnalysisRunDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	r, err := s.MappingSVC.GetMappingAnalysisRun(ctx, tc.TenantID, id)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if r.BusinessID != b {
		return nil, formaerrors.MapDomainError(entity.ErrMappingAnalysisNotFound)
	}
	return mappingRunDTO(r), nil
}
func (s *ApplicationService) RetryFailedMappingAnalysis(ctx context.Context, b, id string) (*AnalyzeSemanticMappingsResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	run, err := s.MappingSVC.GetMappingAnalysisRun(ctx, tc.TenantID, id)
	if err != nil || run.BusinessID != b {
		return nil, formaerrors.MapDomainError(entity.ErrMappingAnalysisNotFound)
	}
	r, err := s.MappingSVC.RetryFailedMappingAnalysis(ctx, tc.TenantID, id, tc.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return mappingAnalysisResponse(r), nil
}
func (s *ApplicationService) ListSemanticMappings(ctx context.Context, b string, rev int32, status string) ([]*SemanticMappingDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	v, err := s.MappingSVC.ListMappings(ctx, tc.TenantID, b, rev, entity.MappingStatus(status))
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return mappingDTOs(v), nil
}
func (s *ApplicationService) CreateManualSemanticMapping(ctx context.Context, b string, in *SemanticMappingInput) (*SemanticMappingDTO, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.MapDomainError(entity.ErrMappingTargetInvalid)
	}
	v, err := s.MappingSVC.CreateManualMapping(ctx, &datasvc.ManualMappingInput{TenantID: tc.TenantID, BusinessID: b, BusinessModelRevision: in.BusinessModelRevision, RequirementID: in.RequirementID, SourceID: in.SourceID, ConnectionID: in.ConnectionID, AssetID: in.AssetID, SchemaSnapshotID: in.SchemaSnapshotID, TargetFieldPaths: in.TargetFieldPaths, MappingType: entity.MappingType(in.MappingType), TransformSpec: in.TransformSpec, Confidence: in.Confidence, Reason: in.Reason, ActorID: tc.PrincipalID})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return mappingDTO(v), nil
}
func (s *ApplicationService) requireMapping(ctx context.Context, t, b, id string) (*entity.SemanticMapping, error) {
	m, err := s.MappingSVC.GetMapping(ctx, t, id)
	if err != nil || m.BusinessID != b {
		return nil, formaerrors.MapDomainError(entity.ErrMappingNotFound)
	}
	return m, nil
}
func (s *ApplicationService) DecideSemanticMapping(ctx context.Context, b, id, action string, in *DecideDataRequirementInput) (*SemanticMappingDecisionResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireMapping(ctx, tc.TenantID, b, id); err != nil {
		return nil, err
	}
	reason := ""
	if in != nil {
		reason = in.Reason
	}
	var m *entity.SemanticMapping
	var d *entity.SemanticMappingDecision
	if action == "confirm" {
		m, d, err = s.MappingSVC.ConfirmMapping(ctx, tc.TenantID, id, tc.PrincipalID, reason)
	} else {
		m, d, err = s.MappingSVC.RejectMapping(ctx, tc.TenantID, id, tc.PrincipalID, reason)
	}
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &SemanticMappingDecisionResponse{mappingDTO(m), mappingDecisionDTO(d)}, nil
}
func (s *ApplicationService) EditConfirmSemanticMapping(ctx context.Context, b, id string, in *EditConfirmSemanticMappingInput) (*EditConfirmSemanticMappingResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.MapDomainError(entity.ErrMappingTargetInvalid)
	}
	if _, err = s.requireMapping(ctx, tc.TenantID, b, id); err != nil {
		return nil, err
	}
	o, r, d, err := s.MappingSVC.EditConfirmMapping(ctx, &datasvc.EditConfirmMappingInput{TenantID: tc.TenantID, SourceMappingID: id, SourceID: in.SourceID, ConnectionID: in.ConnectionID, AssetID: in.AssetID, SchemaSnapshotID: in.SchemaSnapshotID, TargetFieldPaths: in.TargetFieldPaths, MappingType: entity.MappingType(in.MappingType), TransformSpec: in.TransformSpec, Confidence: in.Confidence, Reason: in.Reason, ActorID: tc.PrincipalID})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &EditConfirmSemanticMappingResponse{mappingDTO(o), mappingDTO(r), mappingDecisionDTO(d)}, nil
}
func (s *ApplicationService) ListSemanticMappingDecisions(ctx context.Context, b, id string) ([]*MappingDecisionDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireMapping(ctx, tc.TenantID, b, id); err != nil {
		return nil, err
	}
	v, err := s.MappingSVC.ListMappingDecisions(ctx, tc.TenantID, id)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*MappingDecisionDTO, 0, len(v))
	for _, d := range v {
		out = append(out, mappingDecisionDTO(d))
	}
	return out, nil
}
func (s *ApplicationService) GetSemanticMappingCoverage(ctx context.Context, b string, rev int32) (*datasvc.MappingCoverage, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	v, err := s.MappingSVC.GetMappingCoverage(ctx, tc.TenantID, b, rev)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return v, nil
}
func mappingDTO(v *entity.SemanticMapping) *SemanticMappingDTO {
	if v == nil {
		return nil
	}
	return &SemanticMappingDTO{v.MappingID, v.BusinessID, v.BusinessModelRevision, v.RequirementID, v.SourceID, v.ConnectionID, v.AssetID, v.SchemaSnapshotID, v.TargetFieldPaths, string(v.MappingType), v.TransformSpec, string(v.Status), string(v.Source), v.Confidence, v.Reason, v.DerivedFromMappingID, v.AnalysisRunID, v.CreatedBy, v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func mappingDTOs(v []*entity.SemanticMapping) []*SemanticMappingDTO {
	out := make([]*SemanticMappingDTO, 0, len(v))
	for _, m := range v {
		out = append(out, mappingDTO(m))
	}
	return out
}
func mappingRunDTO(v *entity.SemanticMappingAnalysisRun) *MappingAnalysisRunDTO {
	if v == nil {
		return nil
	}
	return &MappingAnalysisRunDTO{v.AnalysisRunID, v.BusinessID, v.BusinessModelRevision, v.ClientRequestID, string(v.Status), v.ModelRef, v.ErrorKey, v.ErrorMessageSanitized, v.RetryCount, v.CreatedBy, v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func mappingDecisionDTO(v *entity.SemanticMappingDecision) *MappingDecisionDTO {
	if v == nil {
		return nil
	}
	return &MappingDecisionDTO{v.DecisionID, v.BusinessID, v.SourceMappingID, v.TargetMappingID, string(v.Action), v.ActorPrincipalID, v.Reason, v.BusinessModelRevision, v.CreatedAt.UTC().Format(time.RFC3339Nano)}
}
func mappingAnalysisResponse(v *datasvc.AnalyzeSemanticMappingsResult) *AnalyzeSemanticMappingsResponse {
	if v == nil {
		return nil
	}
	return &AnalyzeSemanticMappingsResponse{mappingRunDTO(v.Run), mappingDTOs(v.Mappings), v.OwnedExecute}
}
