/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

type CreateDataContractInput struct {
	BusinessModelRevision int32                          `json:"business_model_revision"`
	Name                  string                         `json:"name"`
	Description           string                         `json:"description"`
	RequirementIDs        []string                       `json:"requirement_ids"`
	LogicalSchema         entity.ContractLogicalSchema   `json:"logical_schema"`
	QueryCapabilities     []entity.QueryCapability       `json:"query_capabilities"`
	FilterSchema          entity.FilterSchema            `json:"filter_schema"`
	SortSchema            entity.SortSchema              `json:"sort_schema"`
	PaginationPolicy      entity.PaginationPolicy        `json:"pagination_policy"`
	FreshnessPolicy       entity.FreshnessPolicy         `json:"freshness_policy"`
	ClassificationPolicy  map[string]entity.DataClassification `json:"classification_policy"`
	MappingIDs            []string                       `json:"mapping_ids"`
	AccessPolicyRef       string                         `json:"access_policy_ref"`
}

type CreateDataContractRevisionInput struct {
	BaseRevisionID        string                         `json:"base_revision_id"`
	BusinessModelRevision int32                          `json:"business_model_revision"`
	Name                  string                         `json:"name"`
	Description           string                         `json:"description"`
	RequirementIDs        []string                       `json:"requirement_ids"`
	LogicalSchema         entity.ContractLogicalSchema   `json:"logical_schema"`
	QueryCapabilities     []entity.QueryCapability       `json:"query_capabilities"`
	FilterSchema          entity.FilterSchema            `json:"filter_schema"`
	SortSchema            entity.SortSchema              `json:"sort_schema"`
	PaginationPolicy      entity.PaginationPolicy        `json:"pagination_policy"`
	FreshnessPolicy       entity.FreshnessPolicy         `json:"freshness_policy"`
	ClassificationPolicy  map[string]entity.DataClassification `json:"classification_policy"`
	MappingIDs            []string                       `json:"mapping_ids"`
	AccessPolicyRef       string                         `json:"access_policy_ref"`
}

type ContractReasonInput struct {
	Reason string `json:"reason"`
}

type EvaluateDriftAppInput struct {
	NewSnapshotIDs map[string]string `json:"new_snapshot_ids"`
}

type DataContractDTO struct {
	ContractID       string `json:"contract_id"`
	BusinessID       string `json:"business_id"`
	ActiveRevisionID string `json:"active_revision_id,omitempty"`
	CreatedBy        string `json:"created_by"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type DataContractRevisionDTO struct {
	RevisionID            string                         `json:"revision_id"`
	BusinessID            string                         `json:"business_id"`
	ContractID            string                         `json:"contract_id"`
	Version               int32                          `json:"version"`
	Status                string                         `json:"status"`
	BusinessModelRevision int32                          `json:"business_model_revision"`
	Name                  string                         `json:"name"`
	Description           string                         `json:"description"`
	RequirementIDs        []string                       `json:"requirement_ids"`
	LogicalSchema         entity.ContractLogicalSchema   `json:"logical_schema"`
	QueryCapabilities     []entity.QueryCapability       `json:"query_capabilities"`
	FilterSchema          entity.FilterSchema            `json:"filter_schema"`
	SortSchema            entity.SortSchema              `json:"sort_schema"`
	PaginationPolicy      entity.PaginationPolicy        `json:"pagination_policy"`
	FreshnessPolicy       entity.FreshnessPolicy         `json:"freshness_policy"`
	ClassificationPolicy  map[string]entity.DataClassification `json:"classification_policy"`
	BindingRefs           []entity.ContractBinding       `json:"binding_refs"`
	AccessPolicyRef       string                         `json:"access_policy_ref,omitempty"`
	DerivedFromRevisionID string                         `json:"derived_from_revision_id,omitempty"`
	CreatedBy             string                         `json:"created_by"`
	CreatedAt             string                         `json:"created_at"`
	UpdatedAt             string                         `json:"updated_at"`
}

type CreateDataContractResponse struct {
	Contract *DataContractDTO         `json:"contract"`
	Revision *DataContractRevisionDTO `json:"revision"`
}

type ValidateRevisionResponse struct {
	Revision *DataContractRevisionDTO  `json:"revision"`
	Result   *entity.DataValidationResult `json:"result"`
}

type EvaluateDriftResponse struct {
	Result   *entity.DataDriftResult   `json:"result"`
	Revision *DataContractRevisionDTO  `json:"revision"`
}

func (s *ApplicationService) CreateDataContract(ctx context.Context, businessID string, in *CreateDataContractInput) (*CreateDataContractResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if s.ContractSVC == nil || in == nil {
		return nil, formaerrors.DataNotConfigured("contract service not initialized")
	}
	c, rev, err := s.ContractSVC.CreateContract(ctx, &datasvc.CreateContractInput{
		TenantID: tc.TenantID, BusinessID: businessID, BusinessModelRevision: in.BusinessModelRevision,
		Name: in.Name, Description: in.Description, RequirementIDs: in.RequirementIDs,
		LogicalSchema: in.LogicalSchema, QueryCapabilities: in.QueryCapabilities,
		FilterSchema: in.FilterSchema, SortSchema: in.SortSchema, PaginationPolicy: in.PaginationPolicy,
		FreshnessPolicy: in.FreshnessPolicy, ClassificationPolicy: in.ClassificationPolicy,
		MappingIDs: in.MappingIDs, AccessPolicyRef: in.AccessPolicyRef, ActorID: tc.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &CreateDataContractResponse{contractDTO(c), contractRevisionDTO(rev)}, nil
}

func (s *ApplicationService) ListDataContracts(ctx context.Context, businessID string) ([]*DataContractDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if s.ContractSVC == nil {
		return nil, formaerrors.DataNotConfigured("contract service not initialized")
	}
	rows, err := s.ContractSVC.ListContracts(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*DataContractDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, contractDTO(r))
	}
	return out, nil
}

func (s *ApplicationService) GetDataContract(ctx context.Context, businessID, contractID string) (*DataContractDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	c, err := s.requireContract(ctx, tc.TenantID, businessID, contractID)
	if err != nil {
		return nil, err
	}
	return contractDTO(c), nil
}

func (s *ApplicationService) requireContract(ctx context.Context, tenantID, businessID, contractID string) (*entity.DataContract, error) {
	if s.ContractSVC == nil {
		return nil, formaerrors.DataNotConfigured("contract service not initialized")
	}
	c, err := s.ContractSVC.GetContract(ctx, tenantID, contractID)
	if err != nil || c.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(entity.ErrContractNotFound)
	}
	return c, nil
}

func (s *ApplicationService) requireContractRevision(ctx context.Context, tenantID, businessID, contractID, revisionID string) (*entity.DataContractRevision, error) {
	if s.ContractSVC == nil {
		return nil, formaerrors.DataNotConfigured("contract service not initialized")
	}
	rev, err := s.ContractSVC.GetRevision(ctx, tenantID, revisionID)
	if err != nil || rev.BusinessID != businessID || rev.ContractID != contractID {
		return nil, formaerrors.MapDomainError(entity.ErrContractRevisionNotFound)
	}
	return rev, nil
}

func (s *ApplicationService) ListDataContractRevisions(ctx context.Context, businessID, contractID string) ([]*DataContractRevisionDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContract(ctx, tc.TenantID, businessID, contractID); err != nil {
		return nil, err
	}
	rows, err := s.ContractSVC.ListRevisions(ctx, tc.TenantID, contractID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*DataContractRevisionDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, contractRevisionDTO(r))
	}
	return out, nil
}

func (s *ApplicationService) CreateDataContractRevision(ctx context.Context, businessID, contractID string, in *CreateDataContractRevisionInput) (*DataContractRevisionDTO, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.MapDomainError(entity.ErrContractInvalidPayload)
	}
	if _, err = s.requireContract(ctx, tc.TenantID, businessID, contractID); err != nil {
		return nil, err
	}
	rev, err := s.ContractSVC.CreateRevision(ctx, &datasvc.CreateRevisionInput{
		TenantID: tc.TenantID, ContractID: contractID, BaseRevisionID: in.BaseRevisionID,
		BusinessModelRevision: in.BusinessModelRevision, Name: in.Name, Description: in.Description,
		RequirementIDs: in.RequirementIDs, LogicalSchema: in.LogicalSchema, QueryCapabilities: in.QueryCapabilities,
		FilterSchema: in.FilterSchema, SortSchema: in.SortSchema, PaginationPolicy: in.PaginationPolicy,
		FreshnessPolicy: in.FreshnessPolicy, ClassificationPolicy: in.ClassificationPolicy,
		MappingIDs: in.MappingIDs, AccessPolicyRef: in.AccessPolicyRef, ActorID: tc.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return contractRevisionDTO(rev), nil
}

func (s *ApplicationService) GetDataContractRevision(ctx context.Context, businessID, contractID, revisionID string) (*DataContractRevisionDTO, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	rev, err := s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID)
	if err != nil {
		return nil, err
	}
	return contractRevisionDTO(rev), nil
}

func (s *ApplicationService) ValidateDataContractRevision(ctx context.Context, businessID, contractID, revisionID string) (*ValidateRevisionResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	rev, result, err := s.ContractSVC.ValidateRevision(ctx, tc.TenantID, revisionID, tc.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &ValidateRevisionResponse{contractRevisionDTO(rev), result}, nil
}

func (s *ApplicationService) ActivateDataContractRevision(ctx context.Context, businessID, contractID, revisionID string, in *ContractReasonInput) (*DataContractRevisionDTO, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	reason := ""
	if in != nil {
		reason = in.Reason
	}
	rev, err := s.ContractSVC.ActivateRevision(ctx, tc.TenantID, revisionID, tc.PrincipalID, reason)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return contractRevisionDTO(rev), nil
}

func (s *ApplicationService) DeprecateDataContractRevision(ctx context.Context, businessID, contractID, revisionID string, in *ContractReasonInput) (*DataContractRevisionDTO, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	reason := ""
	if in != nil {
		reason = in.Reason
	}
	rev, err := s.ContractSVC.DeprecateRevision(ctx, tc.TenantID, revisionID, tc.PrincipalID, reason)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return contractRevisionDTO(rev), nil
}

func (s *ApplicationService) ListDataContractValidationResults(ctx context.Context, businessID, contractID, revisionID string) ([]*entity.DataValidationResult, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	rows, err := s.ContractSVC.ListValidationResults(ctx, tc.TenantID, revisionID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return rows, nil
}

func (s *ApplicationService) EvaluateDataContractDrift(ctx context.Context, businessID, contractID, revisionID string, in *EvaluateDriftAppInput) (*EvaluateDriftResponse, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.MapDomainError(entity.ErrContractDriftInvalid)
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	result, rev, err := s.ContractSVC.EvaluateDrift(ctx, &datasvc.EvaluateDriftInput{
		TenantID: tc.TenantID, RevisionID: revisionID, NewSnapshotIDs: in.NewSnapshotIDs, ActorID: tc.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &EvaluateDriftResponse{result, contractRevisionDTO(rev)}, nil
}

func (s *ApplicationService) ListDataContractDriftResults(ctx context.Context, businessID, contractID, revisionID string) ([]*entity.DataDriftResult, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	rows, err := s.ContractSVC.ListDriftResults(ctx, tc.TenantID, revisionID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return rows, nil
}

func (s *ApplicationService) EvaluateDataContractGap(ctx context.Context, businessID, contractID, revisionID string) (*entity.DataContractGapResult, error) {
	tc, err := s.requireDataEdit(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	result, err := s.ContractSVC.EvaluateBusinessGap(ctx, &datasvc.EvaluateGapInput{
		TenantID: tc.TenantID, RevisionID: revisionID, ActorID: tc.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return result, nil
}

func (s *ApplicationService) ListDataContractGapResults(ctx context.Context, businessID, contractID, revisionID string) ([]*entity.DataContractGapResult, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContractRevision(ctx, tc.TenantID, businessID, contractID, revisionID); err != nil {
		return nil, err
	}
	rows, err := s.ContractSVC.ListGapResults(ctx, tc.TenantID, revisionID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return rows, nil
}

func (s *ApplicationService) ListDataContractLifecycleEvents(ctx context.Context, businessID, contractID string) ([]*entity.DataContractLifecycleEvent, error) {
	tc, err := s.requireDataTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireContract(ctx, tc.TenantID, businessID, contractID); err != nil {
		return nil, err
	}
	rows, err := s.ContractSVC.ListLifecycleEvents(ctx, tc.TenantID, contractID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return rows, nil
}

func contractDTO(v *entity.DataContract) *DataContractDTO {
	if v == nil {
		return nil
	}
	return &DataContractDTO{
		ContractID: v.ContractID, BusinessID: v.BusinessID, ActiveRevisionID: v.ActiveRevisionID,
		CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func contractRevisionDTO(v *entity.DataContractRevision) *DataContractRevisionDTO {
	if v == nil {
		return nil
	}
	return &DataContractRevisionDTO{
		RevisionID: v.RevisionID, BusinessID: v.BusinessID, ContractID: v.ContractID, Version: v.Version,
		Status: string(v.Status), BusinessModelRevision: v.BusinessModelRevision, Name: v.Name, Description: v.Description,
		RequirementIDs: v.RequirementIDs, LogicalSchema: v.LogicalSchema, QueryCapabilities: v.QueryCapabilities,
		FilterSchema: v.FilterSchema, SortSchema: v.SortSchema, PaginationPolicy: v.PaginationPolicy,
		FreshnessPolicy: v.FreshnessPolicy, ClassificationPolicy: v.ClassificationPolicy, BindingRefs: v.BindingRefs,
		AccessPolicyRef: v.AccessPolicyRef, DerivedFromRevisionID: v.DerivedFromRevisionID, CreatedBy: v.CreatedBy,
		CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
