/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

const HeaderFormaTenant = "X-Forma-Tenant"

// ResolveTenantContext builds a TenantContext from the session + optional tenant header.
// Flow: session → ResolveOrCreatePrincipal → if header empty pick first ACTIVE membership
// tenant; else verify membership → reject suspended → load allowed_space_ids from space refs.
func (s *ApplicationService) ResolveTenantContext(ctx context.Context, selectedTenantHeader string) (*tenantctx.TenantContext, error) {
	if s == nil || s.TenancySVC == nil {
		return nil, formaerrors.Internal("tenancy service not initialized")
	}

	session := ctxutil.GetUserSessionFromCtx(ctx)
	if session == nil {
		return nil, formaerrors.Unauthenticated("session required")
	}

	principal, err := s.TenancySVC.ResolveOrCreatePrincipal(ctx, session.UserID, session.UserEmail)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if principal == nil {
		return nil, formaerrors.Unauthenticated("principal not resolved")
	}

	var (
		tenantID   = selectedTenantHeader
		membership *entity.Membership
		tenant     *entity.Tenant
	)

	if tenantID == "" {
		memberships, listErr := s.TenancySVC.ListMembershipsForPrincipal(ctx, principal.PrincipalID)
		if listErr != nil {
			return nil, formaerrors.MapDomainError(listErr)
		}
		for _, m := range memberships {
			if m == nil || m.Status != entity.MembershipActive {
				continue
			}
			t, tErr := s.TenancySVC.GetTenant(ctx, m.TenantID)
			if tErr != nil {
				return nil, formaerrors.MapDomainError(tErr)
			}
			if t == nil || t.Status != entity.TenantStatusActive {
				continue
			}
			tenantID = t.TenantID
			membership = m
			tenant = t
			break
		}
		if tenantID == "" {
			return nil, formaerrors.MembershipRequired("no active tenant membership")
		}
	} else {
		m, mErr := s.TenancySVC.GetMembership(ctx, tenantID, principal.PrincipalID)
		if mErr != nil {
			return nil, formaerrors.MapDomainError(mErr)
		}
		if m == nil || m.Status != entity.MembershipActive {
			return nil, formaerrors.TenantForbidden("not a member of tenant")
		}
		t, tErr := s.TenancySVC.GetTenant(ctx, tenantID)
		if tErr != nil {
			return nil, formaerrors.MapDomainError(tErr)
		}
		if t == nil {
			return nil, formaerrors.TenantNotFound("tenant not found")
		}
		membership = m
		tenant = t
	}

	if tenant.Status == entity.TenantStatusSuspended {
		return nil, formaerrors.TenantSuspended("tenant is suspended")
	}
	if tenant.Status == entity.TenantStatusArchived {
		return nil, formaerrors.TenantForbidden("tenant is archived")
	}

	spaces, err := s.TenancySVC.ListSpaces(ctx, tenant.TenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	allowed := make([]int64, 0, len(spaces))
	for _, sp := range spaces {
		if sp != nil && sp.Status == entity.SpaceRefActive {
			allowed = append(allowed, sp.CozeSpaceID)
		}
	}

	return &tenantctx.TenantContext{
		TenantID:        tenant.TenantID,
		PrincipalID:     principal.PrincipalID,
		MembershipRole:  membership.Role,
		CozeUserID:      session.UserID,
		AllowedSpaceIDs: allowed,
		TenantStatus:    tenant.Status,
	}, nil
}

func roleAtLeastAdmin(role entity.MembershipRole) bool {
	return role == entity.RoleOwner || role == entity.RoleAdmin
}
