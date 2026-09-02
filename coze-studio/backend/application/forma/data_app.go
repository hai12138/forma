/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"time"

	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
)

type AnalyzeDataRequirementsInput struct {
	BusinessModelRevision int32  `json:"business_model_revision"`
	ClientRequestID       string `json:"client_request_id"`
}

type ManualDataRequirementInput struct {
	BusinessModelRevision int32    `json:"business_model_revision"`
	RequirementKind       string   `json:"requirement_kind"`
	SemanticName          string   `json:"semantic_name"`
	Description           string   `json:"description"`
	BusinessElementRefs   []string `json:"business_element_refs"`
	Requiredness          string   `json:"requiredness"`
	FreshnessRequirement  string   `json:"freshness_requirement"`
	AccessNeed            string   `json:"access_need"`
}

type DecideDataRequirementInput struct {
	Reason string `json:"reason"`
}

type EditConfirmDataRequirementInput struct {
	Reason               string   `json:"reason"`
	RequirementKind      string   `json:"requirement_kind"`
	SemanticName         string   `json:"semantic_name"`
	Description          string   `json:"description"`
	BusinessElementRefs  []string `json:"business_element_refs"`
	Requiredness         string   `json:"requiredness"`
	FreshnessRequirement string   `json:"freshness_requirement"`
	AccessNeed           string   `json:"access_need"`
}

type DataRequirementDTO struct {
	RequirementID            string   `json:"requirement_id"`
	BusinessID               string   `json:"business_id"`
	BusinessModelRevision    int32    `json:"business_model_revision"`
	RequirementKind          string   `json:"requirement_kind"`
	SemanticName             string   `json:"semantic_name"`
	Description              string   `json:"description"`
	BusinessElementRefs      []string `json:"business_element_refs"`
	Requiredness             string   `json:"requiredness"`
	FreshnessRequirement     string   `json:"freshness_requirement"`
	AccessNeed               string   `json:"access_need"`
	Status                   string   `json:"status"`
	Source                   string   `json:"source"`
	DerivedFromRequirementID string   `json:"derived_from_requirement_id,omitempty"`
	AnalysisRunID            string   `json:"analysis_run_id,omitempty"`
	CreatedBy                string   `json:"created_by"`
	CreatedAt                string   `json:"created_at"`
	UpdatedAt                string   `json:"updated_at"`
}

type DataAnalysisRunDTO struct {
	AnalysisRunID         string `json:"analysis_run_id"`
	BusinessID            string `json:"business_id"`
	BusinessModelRevision int32  `json:"business_model_revision"`
	ClientRequestID       string `json:"client_request_id"`
	Status                string `json:"status"`
	ModelRef              string `json:"model_ref,omitempty"`
	ErrorKey              string `json:"error_key,omitempty"`
	ErrorMessage          string `json:"error_message_sanitized,omitempty"`
	RetryCount            int32  `json:"retry_count"`
	LastRetryBy           string `json:"last_retry_by,omitempty"`
	LastRetryAt           string `json:"last_retry_at,omitempty"`
	CreatedBy             string `json:"created_by"`
	StartedAt             string `json:"started_at,omitempty"`
	CompletedAt           string `json:"completed_at,omitempty"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

type DataRequirementDecisionDTO struct {
	DecisionID            string `json:"decision_id"`
	BusinessID            string `json:"business_id"`
	SourceRequirementID   string `json:"source_requirement_id"`
	TargetRequirementID   string `json:"target_requirement_id,omitempty"`
	Action                string `json:"action"`
	ActorPrincipalID      string `json:"actor_principal_id"`
	Reason                string `json:"reason"`
	BusinessModelRevision int32  `json:"business_model_revision"`
	CreatedAt             string `json:"created_at"`
}

type AnalyzeDataRequirementsResponse struct {
	AnalysisRun  *DataAnalysisRunDTO   `json:"analysis_run"`
	Requirements []*DataRequirementDTO `json:"requirements"`
	OwnedExecute bool                  `json:"owned_execute"`
}

type DataRequirementDecisionResponse struct {
	Requirement *DataRequirementDTO         `json:"requirement"`
	Decision    *DataRequirementDecisionDTO `json:"decision"`
}

type EditConfirmDataRequirementResponse struct {
	Original    *DataRequirementDTO         `json:"original"`
	Replacement *DataRequirementDTO         `json:"replacement"`
	Decision    *DataRequirementDecisionDTO `json:"decision"`
}

func (s *ApplicationService) AnalyzeDataRequirements(ctx context.Context, businessID string, in *AnalyzeDataRequirementsInput) (*AnalyzeDataRequirementsResponse, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.DataRequirementInvalid("request required")
	}
	result, err := s.DataSVC.AnalyzeDataRequirements(ctx, &datasvc.AnalyzeInput{
		TenantID:              tc.TenantID,
		BusinessID:            businessID,
		BusinessModelRevision: in.BusinessModelRevision,
		ClientRequestID:       in.ClientRequestID,
		ActorID:               tc.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toAnalyzeDataRequirementsResponse(result), nil
}

func (s *ApplicationService) GetDataAnalysisRun(ctx context.Context, businessID, analysisRunID string) (*DataAnalysisRunDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	run, err := s.DataSVC.GetAnalysisRun(ctx, tc.TenantID, analysisRunID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if run.BusinessID != businessID {
		return nil, formaerrors.DataAnalysisNotFound("data analysis run not found")
	}
	return toDataAnalysisRunDTO(run), nil
}

func (s *ApplicationService) RetryFailedDataAnalysis(ctx context.Context, businessID, analysisRunID string) (*AnalyzeDataRequirementsResponse, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	run, err := s.DataSVC.GetAnalysisRun(ctx, tc.TenantID, analysisRunID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if run.BusinessID != businessID {
		return nil, formaerrors.DataAnalysisNotFound("data analysis run not found")
	}
	result, err := s.DataSVC.RetryFailedAnalysis(ctx, tc.TenantID, analysisRunID, tc.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toAnalyzeDataRequirementsResponse(result), nil
}

func (s *ApplicationService) ListDataRequirements(ctx context.Context, businessID string, revision int32) ([]*DataRequirementDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if revision <= 0 {
		return nil, formaerrors.DataRequirementInvalid("business_model_revision required")
	}
	requirements, err := s.DataSVC.ListRequirements(ctx, tc.TenantID, businessID, revision)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toDataRequirementDTOs(requirements), nil
}

func (s *ApplicationService) CreateManualDataRequirement(ctx context.Context, businessID string, in *ManualDataRequirementInput) (*DataRequirementDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.DataRequirementInvalid("request required")
	}
	requirement, err := s.DataSVC.CreateManualRequirement(ctx, &datasvc.ManualCreateInput{
		TenantID:              tc.TenantID,
		BusinessID:            businessID,
		BusinessModelRevision: in.BusinessModelRevision,
		ActorID:               tc.PrincipalID,
		RequirementKind:       dataentity.RequirementKind(in.RequirementKind),
		SemanticName:          in.SemanticName,
		Description:           in.Description,
		BusinessElementRefs:   in.BusinessElementRefs,
		Requiredness:          in.Requiredness,
		FreshnessRequirement:  in.FreshnessRequirement,
		AccessNeed:            in.AccessNeed,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toDataRequirementDTO(requirement), nil
}

func (s *ApplicationService) ConfirmDataRequirement(ctx context.Context, businessID, requirementID string, in *DecideDataRequirementInput) (*DataRequirementDecisionResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	requirement, err := s.requireDataRequirement(ctx, tc.TenantID, businessID, requirementID)
	if err != nil {
		return nil, err
	}
	reason := ""
	if in != nil {
		reason = in.Reason
	}
	requirement, decision, err := s.DataSVC.ConfirmRequirement(ctx, tc.TenantID, requirement.RequirementID, tc.PrincipalID, reason)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &DataRequirementDecisionResponse{Requirement: toDataRequirementDTO(requirement), Decision: toDataRequirementDecisionDTO(decision)}, nil
}

func (s *ApplicationService) RejectDataRequirement(ctx context.Context, businessID, requirementID string, in *DecideDataRequirementInput) (*DataRequirementDecisionResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	requirement, err := s.requireDataRequirement(ctx, tc.TenantID, businessID, requirementID)
	if err != nil {
		return nil, err
	}
	reason := ""
	if in != nil {
		reason = in.Reason
	}
	requirement, decision, err := s.DataSVC.RejectRequirement(ctx, tc.TenantID, requirement.RequirementID, tc.PrincipalID, reason)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &DataRequirementDecisionResponse{Requirement: toDataRequirementDTO(requirement), Decision: toDataRequirementDecisionDTO(decision)}, nil
}

func (s *ApplicationService) EditConfirmDataRequirement(ctx context.Context, businessID, requirementID string, in *EditConfirmDataRequirementInput) (*EditConfirmDataRequirementResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.DataRequirementInvalid("request required")
	}
	if _, err := s.requireDataRequirement(ctx, tc.TenantID, businessID, requirementID); err != nil {
		return nil, err
	}
	original, replacement, decision, err := s.DataSVC.EditConfirmRequirement(ctx, &datasvc.EditConfirmInput{
		TenantID:             tc.TenantID,
		SourceRequirementID:  requirementID,
		ActorID:              tc.PrincipalID,
		Reason:               in.Reason,
		RequirementKind:      dataentity.RequirementKind(in.RequirementKind),
		SemanticName:         in.SemanticName,
		Description:          in.Description,
		BusinessElementRefs:  in.BusinessElementRefs,
		Requiredness:         in.Requiredness,
		FreshnessRequirement: in.FreshnessRequirement,
		AccessNeed:           in.AccessNeed,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &EditConfirmDataRequirementResponse{
		Original:    toDataRequirementDTO(original),
		Replacement: toDataRequirementDTO(replacement),
		Decision:    toDataRequirementDecisionDTO(decision),
	}, nil
}

func (s *ApplicationService) ListDataRequirementDecisions(ctx context.Context, businessID, requirementID string) ([]*DataRequirementDecisionDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireDataRequirement(ctx, tc.TenantID, businessID, requirementID); err != nil {
		return nil, err
	}
	decisions, err := s.DataSVC.ListDecisions(ctx, tc.TenantID, requirementID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*DataRequirementDecisionDTO, 0, len(decisions))
	for _, decision := range decisions {
		out = append(out, toDataRequirementDecisionDTO(decision))
	}
	return out, nil
}

func (s *ApplicationService) requireDataTenant(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, ok := tenantctx.FromContext(ctx)
	if !ok || tc == nil || tc.TenantID == "" {
		return nil, formaerrors.TenantRequired("tenant context required")
	}
	if _, err := s.requireMemberOf(ctx, tc.TenantID); err != nil {
		return nil, err
	}
	if s.DataSVC == nil {
		return nil, formaerrors.DataNotConfigured("data service not initialized")
	}
	return tc, nil
}

func (s *ApplicationService) requireDataEdit(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if !roleAtLeastAdmin(tc.MembershipRole) {
		return nil, formaerrors.DataForbidden("data requirement decisions require OWNER or ADMIN")
	}
	return tc, nil
}

func (s *ApplicationService) requireDataRequirement(ctx context.Context, tenantID, businessID, requirementID string) (*dataentity.DataRequirement, error) {
	requirement, err := s.DataSVC.GetRequirement(ctx, tenantID, requirementID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if requirement.BusinessID != businessID {
		return nil, formaerrors.DataRequirementNotFound("data requirement not found")
	}
	return requirement, nil
}

func toAnalyzeDataRequirementsResponse(result *datasvc.AnalyzeResult) *AnalyzeDataRequirementsResponse {
	if result == nil {
		return nil
	}
	return &AnalyzeDataRequirementsResponse{
		AnalysisRun:  toDataAnalysisRunDTO(result.Run),
		Requirements: toDataRequirementDTOs(result.Requirements),
		OwnedExecute: result.OwnedExecute,
	}
}

func toDataRequirementDTOs(requirements []*dataentity.DataRequirement) []*DataRequirementDTO {
	out := make([]*DataRequirementDTO, 0, len(requirements))
	for _, requirement := range requirements {
		out = append(out, toDataRequirementDTO(requirement))
	}
	return out
}

func toDataRequirementDTO(requirement *dataentity.DataRequirement) *DataRequirementDTO {
	if requirement == nil {
		return nil
	}
	return &DataRequirementDTO{
		RequirementID:            requirement.RequirementID,
		BusinessID:               requirement.BusinessID,
		BusinessModelRevision:    requirement.BusinessModelRevision,
		RequirementKind:          string(requirement.RequirementKind),
		SemanticName:             requirement.SemanticName,
		Description:              requirement.Description,
		BusinessElementRefs:      requirement.BusinessElementRefs,
		Requiredness:             requirement.Requiredness,
		FreshnessRequirement:     requirement.FreshnessRequirement,
		AccessNeed:               requirement.AccessNeed,
		Status:                   string(requirement.Status),
		Source:                   string(requirement.Source),
		DerivedFromRequirementID: requirement.DerivedFromRequirementID,
		AnalysisRunID:            requirement.AnalysisRunID,
		CreatedBy:                requirement.CreatedBy,
		CreatedAt:                requirement.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:                requirement.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toDataAnalysisRunDTO(run *dataentity.DataRequirementAnalysisRun) *DataAnalysisRunDTO {
	if run == nil {
		return nil
	}
	return &DataAnalysisRunDTO{
		AnalysisRunID:         run.AnalysisRunID,
		BusinessID:            run.BusinessID,
		BusinessModelRevision: run.BusinessModelRevision,
		ClientRequestID:       run.ClientRequestID,
		Status:                string(run.Status),
		ModelRef:              run.ModelRef,
		ErrorKey:              run.ErrorKey,
		ErrorMessage:          run.ErrorMessageSanitized,
		RetryCount:            run.RetryCount,
		LastRetryBy:           run.LastRetryBy,
		LastRetryAt:           formatOptionalTime(run.LastRetryAt),
		CreatedBy:             run.CreatedBy,
		StartedAt:             formatOptionalTime(run.StartedAt),
		CompletedAt:           formatOptionalTime(run.CompletedAt),
		CreatedAt:             run.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:             run.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toDataRequirementDecisionDTO(decision *dataentity.DataRequirementDecision) *DataRequirementDecisionDTO {
	if decision == nil {
		return nil
	}
	return &DataRequirementDecisionDTO{
		DecisionID:            decision.DecisionID,
		BusinessID:            decision.BusinessID,
		SourceRequirementID:   decision.SourceRequirementID,
		TargetRequirementID:   decision.TargetRequirementID,
		Action:                string(decision.Action),
		ActorPrincipalID:      decision.ActorPrincipalID,
		Reason:                decision.Reason,
		BusinessModelRevision: decision.BusinessModelRevision,
		CreatedAt:             decision.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
