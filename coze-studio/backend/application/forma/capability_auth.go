/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"errors"
	"time"

	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

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

// capabilityPrincipalID returns the opaque tenancy PrincipalID for Decision / attempt / CreatedBy / ValidatedBy.
// AssetRef OwnerID must use tc.CozeUserID explicitly — never derive OwnerID from PrincipalID.
func capabilityPrincipalID(tc *tenantctx.TenantContext) string {
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

// --- Scoped lookups (fail closed on business mismatch; preserve non-NotFound domain errors) ---

func (s *ApplicationService) requireCapability(ctx context.Context, tenantID, businessID, capabilityID string) (*capentity.BusinessCapability, error) {
	cap, err := s.CapabilitySVC.GetCapability(ctx, tenantID, capabilityID)
	if err != nil {
		if errors.Is(err, capentity.ErrNotFound) {
			return nil, formaerrors.MapDomainError(capentity.ErrNotFound)
		}
		return nil, formaerrors.MapDomainError(err)
	}
	if cap == nil || cap.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrNotFound)
	}
	return cap, nil
}

func (s *ApplicationService) requireCapabilityRevision(ctx context.Context, tenantID, businessID, revisionID string) (*capentity.BusinessCapabilityRevision, error) {
	rev, err := s.CapabilitySVC.GetRevision(ctx, tenantID, revisionID)
	if err != nil {
		if errors.Is(err, capentity.ErrRevisionNotFound) || errors.Is(err, capentity.ErrNotFound) {
			return nil, formaerrors.MapDomainError(capentity.ErrRevisionNotFound)
		}
		return nil, formaerrors.MapDomainError(err)
	}
	if rev == nil || rev.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrRevisionNotFound)
	}
	return rev, nil
}

func (s *ApplicationService) requireCapabilityProposal(ctx context.Context, tenantID, businessID, proposalID string) (*capentity.CapabilityProposal, error) {
	prop, err := s.CapabilitySVC.GetProposal(ctx, tenantID, proposalID)
	if err != nil {
		if errors.Is(err, capentity.ErrProposalNotFound) || errors.Is(err, capentity.ErrNotFound) {
			return nil, formaerrors.MapDomainError(capentity.ErrProposalNotFound)
		}
		return nil, formaerrors.MapDomainError(err)
	}
	if prop == nil || prop.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrProposalNotFound)
	}
	return prop, nil
}

func (s *ApplicationService) requireCapabilityAnalysis(ctx context.Context, tenantID, businessID, analysisRunID string) (*capentity.CapabilityAnalysisRun, error) {
	run, err := s.CapabilitySVC.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		if errors.Is(err, capentity.ErrAnalysisNotFound) || errors.Is(err, capentity.ErrNotFound) {
			return nil, formaerrors.MapDomainError(capentity.ErrAnalysisNotFound)
		}
		return nil, formaerrors.MapDomainError(err)
	}
	if run == nil || run.BusinessID != businessID {
		return nil, formaerrors.MapDomainError(capentity.ErrAnalysisNotFound)
	}
	return run, nil
}
