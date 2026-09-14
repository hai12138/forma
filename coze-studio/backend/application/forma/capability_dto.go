/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"time"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
)

// --- Request DTOs ---

// CreateCapabilityInput is the required JSON body for creating a capability.
type CreateCapabilityInput struct {
	CapabilityID string                    `json:"capability_id,omitempty"`
	Payload      capentity.SemanticPayload `json:"payload"`
}

// DeriveCapabilityInput is the required JSON body for deriving a capability revision.
type DeriveCapabilityInput struct {
	SourceRevisionID string                    `json:"source_revision_id"`
	ClientRequestID  string                    `json:"client_request_id"`
	Reason           string                    `json:"reason,omitempty"`
	Payload          capentity.SemanticPayload `json:"payload"`
}

// EditCapabilityInput is the required JSON body for editing a capability.
type EditCapabilityInput struct {
	SourceRevisionID string                    `json:"source_revision_id"`
	ClientRequestID  string                    `json:"client_request_id"`
	Reason           string                    `json:"reason,omitempty"`
	Payload          capentity.SemanticPayload `json:"payload"`
}

// StartCapabilityAnalysisInput is the required JSON body for starting capability analysis.
type StartCapabilityAnalysisInput struct {
	BusinessModelRevision int32                     `json:"business_model_revision"`
	ClientRequestID       string                    `json:"client_request_id"`
	Analysis              capentity.AnalysisRequest `json:"analysis"`
}

// ConfirmCapabilityProposalInput is the optional JSON body for confirming a proposal.
type ConfirmCapabilityProposalInput struct {
	Reason          string `json:"reason,omitempty"`
	ClientRequestID string `json:"client_request_id,omitempty"`
	CapabilityID    string `json:"capability_id,omitempty"`
}

// EditConfirmCapabilityProposalInput is the required JSON body for edit-confirming a proposal.
type EditConfirmCapabilityProposalInput struct {
	Reason           string                    `json:"reason,omitempty"`
	ClientRequestID  string                    `json:"client_request_id,omitempty"`
	CapabilityID     string                    `json:"capability_id,omitempty"`
	EffectivePayload capentity.SemanticPayload `json:"effective_payload"`
}

// RejectCapabilityProposalInput is the optional JSON body for rejecting a proposal.
type RejectCapabilityProposalInput struct {
	Reason          string `json:"reason,omitempty"`
	ClientRequestID string `json:"client_request_id,omitempty"`
}

// CapabilityReasonInput is the optional JSON body for activate/deprecate reason.
type CapabilityReasonInput struct {
	Reason string `json:"reason,omitempty"`
}

// --- Response DTOs (logical bindings only; no physical/secret fields) ---

// CapabilityDTO is the API representation of a business capability.
type CapabilityDTO struct {
	CapabilityID     string `json:"capability_id"`
	BusinessID       string `json:"business_id"`
	Name             string `json:"name"`
	SemanticVersion  string `json:"semantic_version"`
	AssetStatus      string `json:"asset_status"`
	ContentDigest    string `json:"content_digest"`
	ActiveRevisionID string `json:"active_revision_id,omitempty"`
	CreatedBy        string `json:"created_by"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// CapabilityDataContractBindingDTO is a logical data-contract binding on a capability payload.
type CapabilityDataContractBindingDTO struct {
	DataContractID         string                          `json:"data_contract_id"`
	DataContractRevisionID string                          `json:"data_contract_revision_id"`
	DataContractVersion    int32                           `json:"data_contract_version,omitempty"`
	LogicalFieldMappings   []capentity.LogicalFieldMapping `json:"logical_field_mappings,omitempty"`
}

// CapabilitySemanticPayloadDTO is the API representation of a capability semantic payload.
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

// CapabilityRevisionDTO is the API representation of a capability revision.
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

// CreateCapabilityResponse is returned after creating a capability and its first revision.
type CreateCapabilityResponse struct {
	Capability *CapabilityDTO         `json:"capability"`
	Revision   *CapabilityRevisionDTO `json:"revision"`
}

// DeriveCapabilityResponse is returned after deriving a new capability revision.
type DeriveCapabilityResponse struct {
	Revision *CapabilityRevisionDTO `json:"revision"`
	Decision *CapabilityDecisionDTO `json:"decision"`
}

// CapabilityProposalDTO is the API representation of a capability proposal.
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

// CapabilityAnalysisRunDTO is the API representation of a capability analysis run.
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

// StartCapabilityAnalysisResponse is returned after starting or retrying capability analysis.
type StartCapabilityAnalysisResponse struct {
	AnalysisRun  *CapabilityAnalysisRunDTO `json:"analysis_run"`
	Proposals    []*CapabilityProposalDTO  `json:"proposals"`
	OwnedExecute bool                      `json:"owned_execute"`
}

// CapabilityDecisionDTO is the API representation of a capability decision.
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

// CapabilityValidationResultDTO is the API representation of a capability validation result.
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

// ValidateCapabilityResponse is returned after validating a capability revision.
type ValidateCapabilityResponse struct {
	Revision *CapabilityRevisionDTO         `json:"revision"`
	Result   *CapabilityValidationResultDTO `json:"result"`
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

// attachCapabilityAssetRef fills AssetRef projection fields onto a CapabilityDTO.
// Fail closed on missing / wrong kind / asset_id mismatch / tenant mismatch.
func attachCapabilityAssetRef(dto *CapabilityDTO, cap *capentity.BusinessCapability, asset *assetentity.AssetRef, tenantID string) error {
	if dto == nil {
		return capentity.ErrConsistency
	}
	if err := capsvc.ValidateCapabilityAssetProjection(tenantID, cap, asset); err != nil {
		return err
	}
	dto.Name = asset.Name
	dto.SemanticVersion = asset.SemanticVersion
	dto.AssetStatus = string(asset.Status)
	dto.ContentDigest = asset.ContentDigest
	return nil
}

// attachCapabilityProjectionResult fills summary fields from ProjectCapabilityAssetRef (Create path).
func attachCapabilityProjectionResult(dto *CapabilityDTO, proj *capsvc.ProjectionResult) error {
	if dto == nil || proj == nil {
		return capentity.ErrConsistency
	}
	dto.Name = proj.Name
	dto.SemanticVersion = proj.SemanticVersion
	dto.AssetStatus = string(proj.Status)
	dto.ContentDigest = proj.ContentDigest
	return nil
}

// capabilityAssetMapByID indexes CAPABILITY assets by asset_id; duplicate asset_id → ErrConflict.
func capabilityAssetMapByID(assets []*assetentity.AssetRef) (map[string]*assetentity.AssetRef, error) {
	return capsvc.IndexCapabilityAssetsByID(assets)
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
