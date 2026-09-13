/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"strconv"
	"strings"
	"time"

	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

// --- Request DTOs ---

type CreateCapabilityInput struct {
	CapabilityID string                    `json:"capability_id,omitempty"`
	Payload      capentity.SemanticPayload `json:"payload"`
}

type DeriveCapabilityInput struct {
	SourceRevisionID string                    `json:"source_revision_id"`
	ClientRequestID  string                    `json:"client_request_id"`
	Reason           string                    `json:"reason,omitempty"`
	Payload          capentity.SemanticPayload `json:"payload"`
}

type EditCapabilityInput struct {
	SourceRevisionID string                    `json:"source_revision_id"`
	ClientRequestID  string                    `json:"client_request_id"`
	Reason           string                    `json:"reason,omitempty"`
	Payload          capentity.SemanticPayload `json:"payload"`
}

type StartCapabilityAnalysisInput struct {
	BusinessModelRevision int32                     `json:"business_model_revision"`
	ClientRequestID       string                    `json:"client_request_id"`
	Analysis              capentity.AnalysisRequest `json:"analysis"`
}

type ConfirmCapabilityProposalInput struct {
	Reason          string `json:"reason,omitempty"`
	ClientRequestID string `json:"client_request_id,omitempty"`
	CapabilityID    string `json:"capability_id,omitempty"`
}

type EditConfirmCapabilityProposalInput struct {
	Reason           string                    `json:"reason,omitempty"`
	ClientRequestID  string                    `json:"client_request_id,omitempty"`
	CapabilityID     string                    `json:"capability_id,omitempty"`
	EffectivePayload capentity.SemanticPayload `json:"effective_payload"`
}

type RejectCapabilityProposalInput struct {
	Reason          string `json:"reason,omitempty"`
	ClientRequestID string `json:"client_request_id,omitempty"`
}

type CapabilityReasonInput struct {
	Reason string `json:"reason,omitempty"`
}

// --- Response DTOs (logical bindings only; no physical/secret fields) ---

type CapabilityDTO struct {
	CapabilityID     string `json:"capability_id"`
	BusinessID       string `json:"business_id"`
	ActiveRevisionID string `json:"active_revision_id,omitempty"`
	CreatedBy        string `json:"created_by"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type CapabilityDataContractBindingDTO struct {
	DataContractID         string                          `json:"data_contract_id"`
	DataContractRevisionID string                          `json:"data_contract_revision_id"`
	DataContractVersion    int32                           `json:"data_contract_version,omitempty"`
	LogicalFieldMappings   []capentity.LogicalFieldMapping `json:"logical_field_mappings,omitempty"`
}

type CapabilitySemanticPayloadDTO struct {
	Name                  string                             `json:"name"`
	Description           string                             `json:"description"`
	CapabilityKind        string                             `json:"capability_kind"`
	BusinessModelRevision int32                              `json:"business_model_revision"`
	InputSchema           capentity.LogicalSchema            `json:"input_schema"`
	OutputSchema          capentity.LogicalSchema            `json:"output_schema"`
	Preconditions         []capentity.Precondition           `json:"preconditions"`
	Effects               []capentity.Effect                 `json:"effects"`
	DataContractBindings  []CapabilityDataContractBindingDTO `json:"data_contract_bindings"`
	QueryOperation        string                             `json:"query_operation,omitempty"`
	OutputCardinality     string                             `json:"output_cardinality,omitempty"`
}

type CapabilityRevisionDTO struct {
	RevisionID            string                       `json:"revision_id"`
	CapabilityID          string                       `json:"capability_id"`
	BusinessID            string                       `json:"business_id"`
	Version               int32                        `json:"version"`
	Status                string                       `json:"status"`
	Source                string                       `json:"source"`
	DerivedFromRevisionID string                       `json:"derived_from_revision_id,omitempty"`
	AnalysisRunID         string                       `json:"analysis_run_id,omitempty"`
	ProposalID            string                       `json:"proposal_id,omitempty"`
	Payload               CapabilitySemanticPayloadDTO `json:"payload"`
	CreatedBy             string                       `json:"created_by"`
	CreatedAt             string                       `json:"created_at"`
}

type CreateCapabilityResponse struct {
	Capability *CapabilityDTO         `json:"capability"`
	Revision   *CapabilityRevisionDTO `json:"revision"`
}

type DeriveCapabilityResponse struct {
	Revision *CapabilityRevisionDTO `json:"revision"`
	Decision *CapabilityDecisionDTO `json:"decision"`
}

type CapabilityProposalDTO struct {
	ProposalID             string                       `json:"proposal_id"`
	BusinessID             string                       `json:"business_id"`
	AnalysisRunID          string                       `json:"analysis_run_id"`
	CapabilityID           string                       `json:"capability_id,omitempty"`
	Status                 string                       `json:"status"`
	Payload                CapabilitySemanticPayloadDTO `json:"payload"`
	MaterializedRevisionID string                       `json:"materialized_revision_id,omitempty"`
	CreatedAt              string                       `json:"created_at"`
}

type CapabilityAnalysisRunDTO struct {
	AnalysisRunID         string `json:"analysis_run_id"`
	BusinessID            string `json:"business_id"`
	BusinessModelRevision int32  `json:"business_model_revision"`
	ClientRequestID       string `json:"client_request_id"`
	Status                string `json:"status"`
	Attempt               int32  `json:"attempt"`
	ErrorCode             string `json:"error_code,omitempty"`
	ModelRef              string `json:"model_ref,omitempty"`
	CreatedBy             string `json:"created_by"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

type StartCapabilityAnalysisResponse struct {
	AnalysisRun  *CapabilityAnalysisRunDTO `json:"analysis_run"`
	Proposals    []*CapabilityProposalDTO  `json:"proposals"`
	OwnedExecute bool                      `json:"owned_execute"`
}

type CapabilityDecisionDTO struct {
	DecisionID       string `json:"decision_id"`
	BusinessID       string `json:"business_id"`
	CapabilityID     string `json:"capability_id,omitempty"`
	ProposalID       string `json:"proposal_id,omitempty"`
	SourceRevisionID string `json:"source_revision_id,omitempty"`
	TargetRevisionID string `json:"target_revision_id,omitempty"`
	Action           string `json:"action"`
	PayloadDigest    string `json:"payload_digest,omitempty"`
	ClientRequestID  string `json:"client_request_id,omitempty"`
	ActorPrincipalID string `json:"actor_principal_id"`
	Reason           string `json:"reason,omitempty"`
	CreatedAt        string `json:"created_at"`
}

type CapabilityValidationResultDTO struct {
	ValidationID               string   `json:"validation_id"`
	BusinessID                 string   `json:"business_id"`
	CapabilityID               string   `json:"capability_id"`
	RevisionID                 string   `json:"revision_id"`
	Status                     string   `json:"status"`
	IssueCodes                 []string `json:"issue_codes"`
	RevisionContentDigest      string   `json:"revision_content_digest,omitempty"`
	BusinessModelRevision      int32    `json:"business_model_revision,omitempty"`
	BusinessModelContentDigest string   `json:"business_model_content_digest,omitempty"`
	ContractEvidenceDigest     string   `json:"contract_evidence_digest,omitempty"`
	EvidenceDigest             string   `json:"evidence_digest,omitempty"`
	ValidatedBy                string   `json:"validated_by"`
	ValidatedAt                string   `json:"validated_at"`
}

type ValidateCapabilityResponse struct {
	Revision *CapabilityRevisionDTO         `json:"revision"`
	Result   *CapabilityValidationResultDTO `json:"result"`
}

// --- Auth helpers (ACTIVE membership from SoT; SUPER_ADMIN alone is insufficient) ---

func (s *ApplicationService) loadCapabilityTenant(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, ok := tenantctx.FromContext(ctx)
	if !ok || tc == nil || tc.TenantID == "" || tc.PrincipalID == "" {
		return nil, formaerrors.TenantRequired("tenant context required")
	}
	if s.TenancySVC == nil {
		return nil, formaerrors.Internal("tenancy service not initialized")
	}
	m, err := s.TenancySVC.GetMembership(ctx, tc.TenantID, tc.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if m == nil || m.Status != tenancyentity.MembershipActive {
		return nil, formaerrors.TenantForbidden("not an active member of tenant")
	}
	out := *tc
	out.MembershipRole = m.Role
	return &out, nil
}

func (s *ApplicationService) requireCapabilityRead(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, err := s.loadCapabilityTenant(ctx)
	if err != nil {
		return nil, err
	}
	if s.CapabilitySVC == nil {
		return nil, formaerrors.CapabilityNotConfigured("capability service not initialized")
	}
	return tc, nil
}

func (s *ApplicationService) requireCapabilityAdmin(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if !roleAtLeastAdmin(tc.MembershipRole) {
		return nil, formaerrors.CapabilityForbidden("capability mutation requires OWNER or ADMIN")
	}
	return tc, nil
}

func (s *ApplicationService) requireCapabilityOwner(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if !roleIsOwner(tc.MembershipRole) {
		return nil, formaerrors.CapabilityForbidden("capability activate/deprecate requires OWNER")
	}
	return tc, nil
}

func (s *ApplicationService) requireCapabilityBusiness(ctx context.Context, tenantID, businessID string) error {
	if s.BusinessSVC == nil {
		return formaerrors.Internal("business service not initialized")
	}
	if _, err := s.BusinessSVC.Get(ctx, tenantID, businessID); err != nil {
		return formaerrors.MapDomainError(err)
	}
	return nil
}

func capabilityActorID(tc *tenantctx.TenantContext) string {
	// Domain Asset OwnerID requires a numeric actor; CozeUserID matches domain test convention.
	if tc.CozeUserID != 0 {
		return strconv.FormatInt(tc.CozeUserID, 10)
	}
	return tc.PrincipalID
}

func (s *ApplicationService) recordCapabilityAudit(ctx context.Context, tc *tenantctx.TenantContext, action, resource string) {
	if s.TenancySVC == nil || tc == nil {
		return
	}
	if err := s.TenancySVC.RecordAudit(ctx, &tenancyentity.AuditEvent{
		TenantID:    tc.TenantID,
		PrincipalID: tc.PrincipalID,
		Action:      action,
		Resource:    resource,
		RequestID:   tc.RequestID,
		CreatedAt:   time.Now().UTC(),
	}); err != nil {
		// Best-effort tenancy audit: never fail the successful business mutation.
		// Do not log raw err text (may contain sensitive internals).
		logs.Warnf("capability tenancy audit failed action=%s", action)
	}
}

// --- Scoped lookups (fail closed on business mismatch) ---

func (s *ApplicationService) requireCapability(ctx context.Context, tenantID, businessID, capabilityID string) (*capentity.BusinessCapability, error) {
	cap, err := s.CapabilitySVC.GetCapability(ctx, tenantID, capabilityID)
	if err != nil || cap == nil || cap.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrNotFound)
	}
	return cap, nil
}

func (s *ApplicationService) requireCapabilityRevision(ctx context.Context, tenantID, businessID, revisionID string) (*capentity.BusinessCapabilityRevision, error) {
	rev, err := s.CapabilitySVC.GetRevision(ctx, tenantID, revisionID)
	if err != nil || rev == nil || rev.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrRevisionNotFound)
	}
	return rev, nil
}

func (s *ApplicationService) requireCapabilityProposal(ctx context.Context, tenantID, businessID, proposalID string) (*capentity.CapabilityProposal, error) {
	prop, err := s.CapabilitySVC.GetProposal(ctx, tenantID, proposalID)
	if err != nil || prop == nil || prop.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrProposalNotFound)
	}
	return prop, nil
}

func (s *ApplicationService) requireCapabilityAnalysis(ctx context.Context, tenantID, businessID, analysisRunID string) (*capentity.CapabilityAnalysisRun, error) {
	run, err := s.CapabilitySVC.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil || run == nil || run.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrAnalysisNotFound)
	}
	return run, nil
}

// --- API methods ---

func (s *ApplicationService) ListCapabilities(ctx context.Context, businessID string) ([]*CapabilityDTO, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireCapabilityBusiness(ctx, tc.TenantID, businessID); err != nil {
		return nil, err
	}
	rows, err := s.CapabilitySVC.ListCapabilities(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*CapabilityDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, capabilityDTO(r))
	}
	return out, nil
}

func (s *ApplicationService) GetCapability(ctx context.Context, businessID, capabilityID string) (*CapabilityDTO, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireCapabilityBusiness(ctx, tc.TenantID, businessID); err != nil {
		return nil, err
	}
	cap, err := s.requireCapability(ctx, tc.TenantID, businessID, capabilityID)
	if err != nil {
		return nil, err
	}
	return capabilityDTO(cap), nil
}

func (s *ApplicationService) CreateCapability(ctx context.Context, businessID string, in *CreateCapabilityInput) (*CreateCapabilityResponse, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.CapabilityInvalidPayload("payload required")
	}
	if err := s.requireCapabilityBusiness(ctx, tc.TenantID, businessID); err != nil {
		return nil, err
	}
	cap, rev, err := s.CapabilitySVC.ManualCreate(ctx, &capsvc.ManualCreateInput{
		TenantID:     tc.TenantID,
		BusinessID:   businessID,
		CapabilityID: strings.TrimSpace(in.CapabilityID),
		ActorID:      capabilityActorID(tc),
		OwnerID:      tc.CozeUserID,
		Payload:      in.Payload,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.create", cap.CapabilityID)
	return &CreateCapabilityResponse{Capability: capabilityDTO(cap), Revision: capabilityRevisionDTO(rev)}, nil
}

func (s *ApplicationService) ListCapabilityRevisions(ctx context.Context, businessID, capabilityID string) ([]*CapabilityRevisionDTO, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapability(ctx, tc.TenantID, businessID, capabilityID); err != nil {
		return nil, err
	}
	rows, err := s.CapabilitySVC.ListRevisions(ctx, tc.TenantID, capabilityID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*CapabilityRevisionDTO, 0, len(rows))
	for _, r := range rows {
		if r.BusinessID != businessID {
			continue
		}
		out = append(out, capabilityRevisionDTO(r))
	}
	return out, nil
}

func (s *ApplicationService) GetCapabilityRevision(ctx context.Context, businessID, capabilityID, revisionID string) (*CapabilityRevisionDTO, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapability(ctx, tc.TenantID, businessID, capabilityID); err != nil {
		return nil, err
	}
	rev, err := s.requireCapabilityRevision(ctx, tc.TenantID, businessID, revisionID)
	if err != nil {
		return nil, err
	}
	if rev.CapabilityID != capabilityID {
		return nil, formaerrors.CapabilityRevisionNotFound("")
	}
	return capabilityRevisionDTO(rev), nil
}

func (s *ApplicationService) DeriveCapability(ctx context.Context, businessID, capabilityID string, in *DeriveCapabilityInput) (*DeriveCapabilityResponse, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.CapabilityInvalidPayload("payload required")
	}
	if _, err := s.requireCapability(ctx, tc.TenantID, businessID, capabilityID); err != nil {
		return nil, err
	}
	rev, dec, err := s.CapabilitySVC.DeriveRevision(ctx, &capsvc.DeriveInput{
		TenantID:         tc.TenantID,
		CapabilityID:     capabilityID,
		SourceRevisionID: in.SourceRevisionID,
		ClientRequestID:  in.ClientRequestID,
		ActorID:          capabilityActorID(tc),
		Reason:           in.Reason,
		Action:           capentity.DecisionDerive,
		Payload:          in.Payload,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.derive", rev.RevisionID)
	return &DeriveCapabilityResponse{Revision: capabilityRevisionDTO(rev), Decision: capabilityDecisionDTO(dec)}, nil
}

func (s *ApplicationService) EditCapability(ctx context.Context, businessID, capabilityID string, in *EditCapabilityInput) (*DeriveCapabilityResponse, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.CapabilityInvalidPayload("payload required")
	}
	if _, err := s.requireCapability(ctx, tc.TenantID, businessID, capabilityID); err != nil {
		return nil, err
	}
	rev, dec, err := s.CapabilitySVC.DeriveRevision(ctx, &capsvc.DeriveInput{
		TenantID:         tc.TenantID,
		CapabilityID:     capabilityID,
		SourceRevisionID: in.SourceRevisionID,
		ClientRequestID:  in.ClientRequestID,
		ActorID:          capabilityActorID(tc),
		Reason:           in.Reason,
		Action:           capentity.DecisionEdit,
		Payload:          in.Payload,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.edit", rev.RevisionID)
	return &DeriveCapabilityResponse{Revision: capabilityRevisionDTO(rev), Decision: capabilityDecisionDTO(dec)}, nil
}

func (s *ApplicationService) ConfirmCapabilityProposal(ctx context.Context, businessID, proposalID string, in *ConfirmCapabilityProposalInput) (*CapabilityRevisionDTO, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapabilityProposal(ctx, tc.TenantID, businessID, proposalID); err != nil {
		return nil, err
	}
	reason, clientReq, capID := "", "", ""
	if in != nil {
		reason, clientReq, capID = in.Reason, in.ClientRequestID, in.CapabilityID
	}
	rev, err := s.CapabilitySVC.ConfirmProposal(ctx, &capsvc.ConfirmInput{
		TenantID: tc.TenantID, ProposalID: proposalID, ActorID: capabilityActorID(tc),
		Reason: reason, ClientRequestID: clientReq, CapabilityID: capID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.confirm", rev.RevisionID)
	return capabilityRevisionDTO(rev), nil
}

func (s *ApplicationService) EditConfirmCapabilityProposal(ctx context.Context, businessID, proposalID string, in *EditConfirmCapabilityProposalInput) (*CapabilityRevisionDTO, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.CapabilityInvalidPayload("effective_payload required")
	}
	if _, err := s.requireCapabilityProposal(ctx, tc.TenantID, businessID, proposalID); err != nil {
		return nil, err
	}
	rev, err := s.CapabilitySVC.EditConfirmProposal(ctx, &capsvc.EditConfirmInput{
		TenantID: tc.TenantID, ProposalID: proposalID, ActorID: capabilityActorID(tc),
		Reason: in.Reason, ClientRequestID: in.ClientRequestID, CapabilityID: in.CapabilityID,
		EffectivePayload: in.EffectivePayload,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.edit_confirm", rev.RevisionID)
	return capabilityRevisionDTO(rev), nil
}

func (s *ApplicationService) RejectCapabilityProposal(ctx context.Context, businessID, proposalID string, in *RejectCapabilityProposalInput) (*CapabilityDecisionDTO, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapabilityProposal(ctx, tc.TenantID, businessID, proposalID); err != nil {
		return nil, err
	}
	reason, clientReq := "", ""
	if in != nil {
		reason, clientReq = in.Reason, in.ClientRequestID
	}
	dec, err := s.CapabilitySVC.RejectProposal(ctx, &capsvc.RejectInput{
		TenantID: tc.TenantID, ProposalID: proposalID, ActorID: capabilityActorID(tc),
		Reason: reason, ClientRequestID: clientReq,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.reject", proposalID)
	return capabilityDecisionDTO(dec), nil
}

func (s *ApplicationService) ValidateCapabilityRevision(ctx context.Context, businessID, revisionID string) (*ValidateCapabilityResponse, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapabilityRevision(ctx, tc.TenantID, businessID, revisionID); err != nil {
		return nil, err
	}
	rev, result, err := s.CapabilitySVC.Validate(ctx, tc.TenantID, revisionID, capabilityActorID(tc))
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.validate", revisionID)
	return &ValidateCapabilityResponse{
		Revision: capabilityRevisionDTO(rev),
		Result:   capabilityValidationDTO(result),
	}, nil
}

func (s *ApplicationService) ActivateCapabilityRevision(ctx context.Context, businessID, revisionID string, in *CapabilityReasonInput) (*CapabilityRevisionDTO, error) {
	tc, err := s.requireCapabilityOwner(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapabilityRevision(ctx, tc.TenantID, businessID, revisionID); err != nil {
		return nil, err
	}
	reason := ""
	if in != nil {
		reason = in.Reason
	}
	rev, err := s.CapabilitySVC.Activate(ctx, tc.TenantID, revisionID, capabilityActorID(tc), reason)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.activate", revisionID)
	return capabilityRevisionDTO(rev), nil
}

func (s *ApplicationService) DeprecateCapabilityRevision(ctx context.Context, businessID, revisionID string, in *CapabilityReasonInput) (*CapabilityRevisionDTO, error) {
	tc, err := s.requireCapabilityOwner(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapabilityRevision(ctx, tc.TenantID, businessID, revisionID); err != nil {
		return nil, err
	}
	reason := ""
	if in != nil {
		reason = in.Reason
	}
	rev, err := s.CapabilitySVC.Deprecate(ctx, tc.TenantID, revisionID, capabilityActorID(tc), reason)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.deprecate", revisionID)
	return capabilityRevisionDTO(rev), nil
}

func (s *ApplicationService) ListCapabilityValidations(ctx context.Context, businessID, revisionID string) ([]*CapabilityValidationResultDTO, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapabilityRevision(ctx, tc.TenantID, businessID, revisionID); err != nil {
		return nil, err
	}
	rows, err := s.CapabilitySVC.ListValidations(ctx, tc.TenantID, revisionID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*CapabilityValidationResultDTO, 0, len(rows))
	for _, r := range rows {
		if r.BusinessID != businessID {
			continue
		}
		out = append(out, capabilityValidationDTO(r))
	}
	return out, nil
}

func (s *ApplicationService) ListCapabilityDecisions(ctx context.Context, businessID, capabilityID string) ([]*CapabilityDecisionDTO, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireCapability(ctx, tc.TenantID, businessID, capabilityID); err != nil {
		return nil, err
	}
	rows, err := s.CapabilitySVC.ListDecisions(ctx, tc.TenantID, capabilityID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*CapabilityDecisionDTO, 0, len(rows))
	for _, r := range rows {
		if r.BusinessID != businessID {
			continue
		}
		out = append(out, capabilityDecisionDTO(r))
	}
	return out, nil
}

// --- DTO mappers ---

func capabilityDTO(v *capentity.BusinessCapability) *CapabilityDTO {
	if v == nil {
		return nil
	}
	return &CapabilityDTO{
		CapabilityID: v.CapabilityID, BusinessID: v.BusinessID, ActiveRevisionID: v.ActiveRevisionID,
		CreatedBy: v.CreatedBy,
		CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func capabilitySemanticDTO(p capentity.SemanticPayload) CapabilitySemanticPayloadDTO {
	bindings := make([]CapabilityDataContractBindingDTO, 0, len(p.DataContractBindings))
	for _, b := range p.DataContractBindings {
		bindings = append(bindings, CapabilityDataContractBindingDTO{
			DataContractID:         b.DataContractID,
			DataContractRevisionID: b.DataContractRevisionID,
			DataContractVersion:    b.DataContractVersion,
			LogicalFieldMappings:   b.LogicalFieldMappings,
		})
	}
	return CapabilitySemanticPayloadDTO{
		Name: p.Name, Description: p.Description, CapabilityKind: string(p.CapabilityKind),
		BusinessModelRevision: p.BusinessModelRevision,
		InputSchema:           p.InputSchema, OutputSchema: p.OutputSchema,
		Preconditions: p.Preconditions, Effects: p.Effects,
		DataContractBindings: bindings,
		QueryOperation:       string(p.QueryOperation),
		OutputCardinality:    string(p.OutputCardinality),
	}
}

func capabilityRevisionDTO(v *capentity.BusinessCapabilityRevision) *CapabilityRevisionDTO {
	if v == nil {
		return nil
	}
	return &CapabilityRevisionDTO{
		RevisionID: v.RevisionID, CapabilityID: v.CapabilityID, BusinessID: v.BusinessID,
		Version: v.Version, Status: string(v.Status), Source: string(v.Source),
		DerivedFromRevisionID: v.DerivedFromRevisionID, AnalysisRunID: v.AnalysisRunID,
		ProposalID: v.ProposalID, Payload: capabilitySemanticDTO(v.ToSemanticPayload()),
		CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func capabilityProposalDTO(v *capentity.CapabilityProposal) *CapabilityProposalDTO {
	if v == nil {
		return nil
	}
	return &CapabilityProposalDTO{
		ProposalID: v.ProposalID, BusinessID: v.BusinessID, AnalysisRunID: v.AnalysisRunID,
		CapabilityID: v.CapabilityID, Status: string(v.Status),
		Payload: capabilitySemanticDTO(v.Payload), MaterializedRevisionID: v.MaterializedRevisionID,
		CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func capabilityAnalysisDTO(v *capentity.CapabilityAnalysisRun) *CapabilityAnalysisRunDTO {
	if v == nil {
		return nil
	}
	return &CapabilityAnalysisRunDTO{
		AnalysisRunID: v.AnalysisRunID, BusinessID: v.BusinessID,
		BusinessModelRevision: v.BusinessModelRevision, ClientRequestID: v.ClientRequestID,
		Status: string(v.Status), Attempt: v.Attempt, ErrorCode: v.ErrorCode, ModelRef: v.ModelRef,
		CreatedBy: v.CreatedBy,
		CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func capabilityDecisionDTO(v *capentity.CapabilityDecision) *CapabilityDecisionDTO {
	if v == nil {
		return nil
	}
	return &CapabilityDecisionDTO{
		DecisionID: v.DecisionID, BusinessID: v.BusinessID, CapabilityID: v.CapabilityID,
		ProposalID: v.ProposalID, SourceRevisionID: v.SourceRevisionID, TargetRevisionID: v.TargetRevisionID,
		Action: string(v.Action), PayloadDigest: v.PayloadDigest, ClientRequestID: v.ClientRequestID,
		ActorPrincipalID: v.ActorPrincipalID, Reason: v.Reason,
		CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func capabilityValidationDTO(v *capentity.CapabilityValidationResult) *CapabilityValidationResultDTO {
	if v == nil {
		return nil
	}
	return &CapabilityValidationResultDTO{
		ValidationID: v.ValidationID, BusinessID: v.BusinessID, CapabilityID: v.CapabilityID,
		RevisionID: v.RevisionID, Status: string(v.Status), IssueCodes: v.IssueCodes,
		RevisionContentDigest: v.RevisionContentDigest, BusinessModelRevision: v.BusinessModelRevision,
		BusinessModelContentDigest: v.BusinessModelContentDigest, ContractEvidenceDigest: v.ContractEvidenceDigest,
		EvidenceDigest: v.EvidenceDigest, ValidatedBy: v.ValidatedBy,
		ValidatedAt: v.ValidatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toStartAnalysisResponse(res *capsvc.AnalysisResult) *StartCapabilityAnalysisResponse {
	if res == nil {
		return nil
	}
	props := make([]*CapabilityProposalDTO, 0, len(res.Proposals))
	for _, p := range res.Proposals {
		props = append(props, capabilityProposalDTO(p))
	}
	return &StartCapabilityAnalysisResponse{
		AnalysisRun: capabilityAnalysisDTO(res.Run), Proposals: props, OwnedExecute: res.OwnedExecute,
	}
}
