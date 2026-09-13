/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"time"

	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
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
