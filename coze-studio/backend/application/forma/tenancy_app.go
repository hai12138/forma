/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	crossforma "github.com/coze-dev/coze-studio/backend/crossdomain/forma"
	"github.com/coze-dev/coze-studio/backend/crossdomain/forma/integration"
	crossuser "github.com/coze-dev/coze-studio/backend/crossdomain/user"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

type MeResponse struct {
	Principal     *PrincipalDTO    `json:"principal"`
	CurrentTenant *TenantDTO       `json:"current_tenant"`
	Memberships   []*MembershipDTO `json:"memberships"`
	Tenants       []*TenantDTO     `json:"tenants"`
	CozeUserID    int64            `json:"coze_user_id"`
}

type PrincipalDTO struct {
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	DisplayName   string `json:"display_name"`
	CozeUserID    int64  `json:"coze_user_id"`
	Status        string `json:"status"`
}

type TenantDTO struct {
	TenantID    string `json:"tenant_id"`
	TenantKey   string `json:"tenant_key"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	Revision    int32  `json:"revision"`
	Role        string `json:"role,omitempty"`
}

type MembershipDTO struct {
	TenantID    string `json:"tenant_id"`
	PrincipalID string `json:"principal_id"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	Revision    int32  `json:"revision"`
}

type CreateTenantInput struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	TenantKey   string `json:"tenant_key"`
}

type PatchTenantInput struct {
	Name             *string                     `json:"name"`
	DisplayName      *string                     `json:"display_name"`
	Status           *tenancyentity.TenantStatus `json:"status"`
	ExpectedRevision int32                       `json:"expected_revision"`
}

type AddMemberInput struct {
	PrincipalID string                       `json:"principal_id"`
	Role        tenancyentity.MembershipRole `json:"role"`
}

type PatchMemberInput struct {
	Role             tenancyentity.MembershipRole `json:"role"`
	ExpectedRevision int32                        `json:"expected_revision"`
}

type BindSpaceInput struct {
	CozeSpaceID int64                      `json:"coze_space_id"`
	Purpose     tenancyentity.SpacePurpose `json:"purpose"`
}

type BootstrapInput struct {
	DefaultSpaceID int64 `json:"default_space_id"`
}

type AssetCountsResponse struct {
	Business    int `json:"business"`
	Capability  int `json:"capability"`
	Agent       int `json:"agent"`
	Application int `json:"application"`
}

func (s *ApplicationService) currentPrincipal(ctx context.Context) (*tenancyentity.Principal, error) {
	session := ctxutil.GetUserSessionFromCtx(ctx)
	if session == nil {
		return nil, formaerrors.Unauthenticated("session required")
	}
	if s.TenancySVC == nil {
		return nil, formaerrors.Internal("tenancy service not initialized")
	}
	p, err := s.TenancySVC.ResolveOrCreatePrincipal(ctx, session.UserID, session.UserEmail)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return p, nil
}

func (s *ApplicationService) Me(ctx context.Context) (*MeResponse, error) {
	session := ctxutil.GetUserSessionFromCtx(ctx)
	if session == nil {
		return nil, formaerrors.Unauthenticated("session required")
	}
	p, err := s.currentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	tenants, err := s.TenancySVC.ListTenantsForPrincipal(ctx, p.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	memberships, err := s.TenancySVC.ListMembershipsForPrincipal(ctx, p.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}

	roleByTenant := map[string]string{}
	memDTOs := make([]*MembershipDTO, 0, len(memberships))
	for _, m := range memberships {
		if m == nil || m.Status != tenancyentity.MembershipActive {
			continue
		}
		roleByTenant[m.TenantID] = string(m.Role)
		memDTOs = append(memDTOs, &MembershipDTO{
			TenantID:    m.TenantID,
			PrincipalID: m.PrincipalID,
			Role:        string(m.Role),
			Status:      string(m.Status),
			Revision:    m.Revision,
		})
	}

	tenantDTOs := make([]*TenantDTO, 0, len(tenants))
	var current *TenantDTO
	header := ""
	if tc, ok := tenantctx.FromContext(ctx); ok && tc != nil {
		header = tc.TenantID
	}
	for _, t := range tenants {
		if t == nil {
			continue
		}
		dto := &TenantDTO{
			TenantID:    t.TenantID,
			TenantKey:   t.TenantKey,
			Name:        t.Name,
			DisplayName: t.DisplayName,
			Status:      string(t.Status),
			Revision:    t.Revision,
			Role:        roleByTenant[t.TenantID],
		}
		tenantDTOs = append(tenantDTOs, dto)
		if header != "" && t.TenantID == header {
			current = dto
		}
	}
	if current == nil && len(tenantDTOs) > 0 {
		current = tenantDTOs[0]
	}

	return &MeResponse{
		Principal: &PrincipalDTO{
			PrincipalID:   p.PrincipalID,
			PrincipalType: string(p.PrincipalType),
			DisplayName:   p.DisplayName,
			CozeUserID:    p.CozeUserID,
			Status:        string(p.Status),
		},
		CurrentTenant: current,
		Memberships:   memDTOs,
		Tenants:       tenantDTOs,
		CozeUserID:    session.UserID,
	}, nil
}

func (s *ApplicationService) ListTenants(ctx context.Context) ([]*tenancyentity.Tenant, error) {
	p, err := s.currentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	tenants, err := s.TenancySVC.ListTenantsForPrincipal(ctx, p.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return tenants, nil
}

func (s *ApplicationService) CreateTenant(ctx context.Context, in *CreateTenantInput) (*tenancyentity.Tenant, error) {
	if in == nil || in.Name == "" {
		return nil, formaerrors.TenantRequired("name is required")
	}
	p, err := s.currentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	tenant, err := s.TenancySVC.CreateTenant(ctx, &tenancysvc.CreateTenantRequest{
		TenantKey:        in.TenantKey,
		Name:             in.Name,
		DisplayName:      in.DisplayName,
		OwnerPrincipalID: p.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	_, err = s.TenancySVC.AddMember(ctx, &tenancysvc.AddMemberRequest{
		TenantID:    tenant.TenantID,
		PrincipalID: p.PrincipalID,
		Role:        tenancyentity.RoleOwner,
		CreatedBy:   p.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return tenant, nil
}

func (s *ApplicationService) GetTenant(ctx context.Context, tenantID string) (*tenancyentity.Tenant, error) {
	if _, err := s.requireMemberOfAllowSuspended(ctx, tenantID); err != nil {
		return nil, err
	}
	tenant, err := s.TenancySVC.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if tenant == nil {
		return nil, formaerrors.TenantNotFound("tenant not found")
	}
	return tenant, nil
}

func (s *ApplicationService) PatchTenant(ctx context.Context, tenantID string, in *PatchTenantInput) (*tenancyentity.Tenant, error) {
	if in == nil {
		return nil, formaerrors.TenantRequired("patch body required")
	}
	if in.ExpectedRevision <= 0 {
		return nil, formaerrors.TenantRequired("expected_revision is required")
	}
	tc, err := s.requireMemberOfAllowSuspended(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !roleAtLeastAdmin(tc.MembershipRole) {
		return nil, formaerrors.MembershipForbidden("OWNER or ADMIN required")
	}
	tenant, err := s.TenancySVC.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if tenant == nil {
		return nil, formaerrors.TenantNotFound("tenant not found")
	}
	if in.Name != nil {
		tenant.Name = *in.Name
	}
	if in.DisplayName != nil {
		tenant.DisplayName = *in.DisplayName
	}
	if in.Status != nil {
		tenant.Status = *in.Status
	}
	updated, err := s.TenancySVC.UpdateTenant(ctx, tenant, in.ExpectedRevision, tc.PrincipalID, tc.RequestID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return updated, nil
}

func (s *ApplicationService) ListMembers(ctx context.Context, tenantID string) ([]*tenancyentity.Membership, error) {
	if _, err := s.requireMemberOf(ctx, tenantID); err != nil {
		return nil, err
	}
	members, err := s.TenancySVC.ListMembers(ctx, tenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return members, nil
}

func (s *ApplicationService) AddMember(ctx context.Context, tenantID string, in *AddMemberInput) (*tenancyentity.Membership, error) {
	if in == nil || in.PrincipalID == "" {
		return nil, formaerrors.MembershipRequired("principal_id is required")
	}
	tc, err := s.requireMemberOf(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	role := in.Role
	if role == "" {
		role = tenancyentity.RoleMember
	}
	if !tenancysvc.ValidRole(role) {
		return nil, formaerrors.MapDomainError(tenancyentity.ErrInvalidRole)
	}
	policy := tenancysvc.MembershipPolicy{}
	if err := policy.CanAddMember(tc.MembershipRole, role); err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	m, err := s.TenancySVC.AddMember(ctx, &tenancysvc.AddMemberRequest{
		TenantID:    tenantID,
		PrincipalID: in.PrincipalID,
		Role:        role,
		CreatedBy:   tc.PrincipalID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return m, nil
}

func (s *ApplicationService) PatchMember(ctx context.Context, tenantID, principalID string, in *PatchMemberInput) (*tenancyentity.Membership, error) {
	if in == nil || in.Role == "" {
		return nil, formaerrors.MembershipRequired("role is required")
	}
	if in.ExpectedRevision <= 0 {
		return nil, formaerrors.TenantRequired("expected_revision is required")
	}
	tc, err := s.requireMemberOf(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !tenancysvc.ValidRole(in.Role) {
		return nil, formaerrors.MapDomainError(tenancyentity.ErrInvalidRole)
	}
	target, gErr := s.TenancySVC.GetMembership(ctx, tenantID, principalID)
	if gErr != nil {
		return nil, formaerrors.MapDomainError(gErr)
	}
	if target == nil || target.Status != tenancyentity.MembershipActive {
		return nil, formaerrors.MembershipRequired("membership not found")
	}
	tenant, tErr := s.TenancySVC.GetTenant(ctx, tenantID)
	if tErr != nil {
		return nil, formaerrors.MapDomainError(tErr)
	}
	if tenant == nil {
		return nil, formaerrors.TenantNotFound("tenant not found")
	}
	policy := tenancysvc.MembershipPolicy{}
	if err := policy.CanChangeRole(
		tc.MembershipRole,
		tc.PrincipalID,
		principalID,
		target.Role,
		in.Role,
		tenant.OwnerPrincipalID,
	); err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	m, err := s.TenancySVC.UpdateMemberRole(ctx, tenantID, principalID, in.Role, in.ExpectedRevision, tc.PrincipalID, tc.RequestID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return m, nil
}

func (s *ApplicationService) ListSpaces(ctx context.Context, tenantID string) ([]*tenancyentity.TenantSpaceRef, error) {
	if _, err := s.requireMemberOf(ctx, tenantID); err != nil {
		return nil, err
	}
	spaces, err := s.TenancySVC.ListSpaces(ctx, tenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return spaces, nil
}

func (s *ApplicationService) BindSpace(ctx context.Context, tenantID string, in *BindSpaceInput) (*tenancyentity.TenantSpaceRef, error) {
	if in == nil || in.CozeSpaceID == 0 {
		return nil, formaerrors.SpaceNotFound("coze_space_id is required")
	}
	tc, err := s.requireMemberOf(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !roleAtLeastAdmin(tc.MembershipRole) {
		return nil, formaerrors.MembershipForbidden("OWNER or ADMIN required")
	}

	if adapter := spaceAdapter(); adapter != nil {
		if vErr := adapter.ValidateSpaceAccess(ctx, tc.CozeUserID, in.CozeSpaceID); vErr != nil {
			if fe, ok := formaerrors.AsFormaError(vErr); ok {
				return nil, fe
			}
			return nil, formaerrors.MapDomainError(vErr)
		}
	}

	ref, err := s.TenancySVC.BindSpace(ctx, &tenancysvc.BindSpaceRequest{
		TenantID:         tenantID,
		CozeSpaceID:      in.CozeSpaceID,
		Purpose:          in.Purpose,
		ActorPrincipalID: tc.PrincipalID,
		RequestID:        tc.RequestID,
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return ref, nil
}

func (s *ApplicationService) Bootstrap(ctx context.Context, in *BootstrapInput) (*tenancysvc.BootstrapResult, error) {
	session := ctxutil.GetUserSessionFromCtx(ctx)
	if session == nil {
		return nil, formaerrors.Unauthenticated("session required")
	}
	spaceID := int64(0)
	if in != nil {
		spaceID = in.DefaultSpaceID
	}
	if spaceID == 0 {
		resolved, err := firstUserSpaceID(ctx, session.UserID)
		if err != nil {
			return nil, err
		}
		spaceID = resolved
	}
	if spaceID > 0 {
		if adapter := spaceAdapter(); adapter != nil {
			if err := adapter.ValidateSpaceAccess(ctx, session.UserID, spaceID); err != nil {
				if fe, ok := formaerrors.AsFormaError(err); ok {
					return nil, fe
				}
				return nil, formaerrors.MapDomainError(err)
			}
		}
	}
	result, err := s.TenancySVC.Bootstrap(ctx, session.UserID, session.UserEmail, spaceID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return result, nil
}

func (s *ApplicationService) AssetCounts(ctx context.Context) (*AssetCountsResponse, error) {
	tc, ok := tenantctx.FromContext(ctx)
	if !ok || tc == nil || tc.TenantID == "" {
		return nil, formaerrors.TenantRequired("tenant context required")
	}
	if _, err := s.requireMemberOf(ctx, tc.TenantID); err != nil {
		return nil, err
	}
	if s.DomainSVC == nil {
		return nil, formaerrors.Internal("asset registry not initialized")
	}
	assets, err := s.DomainSVC.ListAssetsByTenant(ctx, tc.TenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	resp := &AssetCountsResponse{}
	for _, a := range assets {
		if a == nil {
			continue
		}
		switch a.Kind {
		case "BUSINESS":
			resp.Business++
		case "CAPABILITY":
			resp.Capability++
		case "AGENT":
			resp.Agent++
		case "APPLICATION":
			resp.Application++
		}
	}
	return resp, nil
}

func (s *ApplicationService) requireMemberOf(ctx context.Context, tenantID string) (*tenantctx.TenantContext, error) {
	return s.requireMemberOfOpt(ctx, tenantID, false)
}

func (s *ApplicationService) requireMemberOfAllowSuspended(ctx context.Context, tenantID string) (*tenantctx.TenantContext, error) {
	return s.requireMemberOfOpt(ctx, tenantID, true)
}

func (s *ApplicationService) requireMemberOfOpt(ctx context.Context, tenantID string, allowSuspended bool) (*tenantctx.TenantContext, error) {
	if tenantID == "" {
		return nil, formaerrors.TenantRequired("tenant_id is required")
	}
	if tc, ok := tenantctx.FromContext(ctx); ok && tc != nil && tc.TenantID == tenantID {
		if tc.TenantStatus == tenancyentity.TenantStatusSuspended && !allowSuspended {
			return nil, formaerrors.TenantSuspended("tenant is suspended")
		}
		return tc, nil
	}
	return s.resolveTenantContext(ctx, tenantID, allowSuspended)
}

func spaceAdapter() integration.CozeSpaceAdapter {
	svc := crossforma.DefaultSVC()
	if svc == nil || svc.Integration() == nil {
		return nil
	}
	return svc.Integration().Space()
}

func firstUserSpaceID(ctx context.Context, cozeUserID int64) (int64, error) {
	svc := crossuser.DefaultSVC()
	if svc == nil {
		return 0, nil
	}
	spaces, err := svc.GetUserSpaceList(ctx, cozeUserID)
	if err != nil {
		return 0, formaerrors.MapDomainError(err)
	}
	for _, sp := range spaces {
		if sp != nil && sp.ID != 0 {
			return sp.ID, nil
		}
	}
	return 0, nil
}
