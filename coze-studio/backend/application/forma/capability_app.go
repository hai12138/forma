/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"strings"

	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

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
		ActorID:      capabilityPrincipalID(tc),
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
		ActorID:          capabilityPrincipalID(tc),
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
		ActorID:          capabilityPrincipalID(tc),
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
		TenantID: tc.TenantID, ProposalID: proposalID, ActorID: capabilityPrincipalID(tc),
		OwnerID: tc.CozeUserID, Reason: reason, ClientRequestID: clientReq, CapabilityID: capID,
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
		TenantID: tc.TenantID, ProposalID: proposalID, ActorID: capabilityPrincipalID(tc),
		OwnerID: tc.CozeUserID, Reason: in.Reason, ClientRequestID: in.ClientRequestID, CapabilityID: in.CapabilityID,
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
		TenantID: tc.TenantID, ProposalID: proposalID, ActorID: capabilityPrincipalID(tc),
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
	rev, result, err := s.CapabilitySVC.Validate(ctx, tc.TenantID, revisionID, capabilityPrincipalID(tc))
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
	rev, err := s.CapabilitySVC.Activate(ctx, tc.TenantID, revisionID, capabilityPrincipalID(tc), reason)
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
	rev, err := s.CapabilitySVC.Deprecate(ctx, tc.TenantID, revisionID, capabilityPrincipalID(tc), reason)
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
