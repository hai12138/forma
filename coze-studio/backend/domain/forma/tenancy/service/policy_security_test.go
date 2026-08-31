/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

func TestAdminCannotPromoteSelfToOwner(t *testing.T) {
	policy := service.MembershipPolicy{}
	err := policy.CanChangeRole(
		entity.RoleAdmin,
		"admin-1", "admin-1",
		entity.RoleAdmin, entity.RoleOwner,
		"owner-1",
	)
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.CodeMembershipForbidden, fe.Code)
}

func TestAdminCannotPromoteMemberToOwner(t *testing.T) {
	policy := service.MembershipPolicy{}
	err := policy.CanChangeRole(
		entity.RoleAdmin,
		"admin-1", "member-1",
		entity.RoleMember, entity.RoleOwner,
		"owner-1",
	)
	require.Error(t, err)
	_, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
}

func TestAdminCannotAddOwner(t *testing.T) {
	policy := service.MembershipPolicy{}
	err := policy.CanAddMember(entity.RoleAdmin, entity.RoleOwner)
	require.Error(t, err)
	_, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
}

func TestAdminCannotModifyOwner(t *testing.T) {
	policy := service.MembershipPolicy{}
	err := policy.CanChangeRole(
		entity.RoleAdmin,
		"admin-1", "owner-1",
		entity.RoleOwner, entity.RoleAdmin,
		"owner-1",
	)
	require.Error(t, err)

	err = policy.CanRemoveMember(
		entity.RoleAdmin,
		"admin-1", "owner-1",
		entity.RoleOwner,
		"owner-1",
		2,
	)
	require.Error(t, err)
}

func TestMemberCannotManageMembership(t *testing.T) {
	policy := service.MembershipPolicy{}
	require.Error(t, policy.CanAddMember(entity.RoleMember, entity.RoleViewer))
	require.Error(t, policy.CanChangeRole(
		entity.RoleMember, "m1", "m2",
		entity.RoleViewer, entity.RoleMember, "owner-1",
	))
	require.Error(t, policy.CanRemoveMember(
		entity.RoleMember, "m1", "m2",
		entity.RoleViewer, "owner-1", 1,
	))
}

func TestViewerCannotManageMembership(t *testing.T) {
	policy := service.MembershipPolicy{}
	require.Error(t, policy.CanAddMember(entity.RoleViewer, entity.RoleMember))
	require.Error(t, policy.CanChangeRole(
		entity.RoleViewer, "v1", "m1",
		entity.RoleMember, entity.RoleViewer, "owner-1",
	))
	require.Error(t, policy.CanRemoveMember(
		entity.RoleViewer, "v1", "m1",
		entity.RoleMember, "owner-1", 1,
	))
}

func TestLastOwnerCannotBeDemoted(t *testing.T) {
	svc, _ := newTestServiceWithAudit()
	ctx := context.Background()

	primary, err := svc.ResolveOrCreatePrincipal(ctx, 10, "Primary")
	require.NoError(t, err)
	soleOwner, err := svc.ResolveOrCreatePrincipal(ctx, 11, "SoleOwner")
	require.NoError(t, err)

	// Primary owner id is recorded on the tenant, but the only ACTIVE OWNER membership
	// belongs to soleOwner — exercises the last-owner guard independently of primary immutability.
	tenant, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name:             "Solo",
		OwnerPrincipalID: primary.PrincipalID,
	})
	require.NoError(t, err)
	_, err = svc.AddMember(ctx, &service.AddMemberRequest{
		TenantID:    tenant.TenantID,
		PrincipalID: soleOwner.PrincipalID,
		Role:        entity.RoleOwner,
		CreatedBy:   primary.PrincipalID,
	})
	require.NoError(t, err)

	_, err = svc.UpdateMemberRole(ctx, tenant.TenantID, soleOwner.PrincipalID, entity.RoleAdmin, 1, soleOwner.PrincipalID, "req-last")
	require.ErrorIs(t, err, entity.ErrLastOwner)

	err = svc.RemoveMember(ctx, tenant.TenantID, soleOwner.PrincipalID, soleOwner.PrincipalID, "req-last")
	require.ErrorIs(t, err, entity.ErrLastOwner)
}

func TestPrimaryOwnerInvariant(t *testing.T) {
	svc, _ := newTestServiceWithAudit()
	ctx := context.Background()

	primary, err := svc.ResolveOrCreatePrincipal(ctx, 11, "Primary")
	require.NoError(t, err)
	coOwner, err := svc.ResolveOrCreatePrincipal(ctx, 12, "CoOwner")
	require.NoError(t, err)

	tenant, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name:             "Dual",
		OwnerPrincipalID: primary.PrincipalID,
	})
	require.NoError(t, err)
	_, err = svc.AddMember(ctx, &service.AddMemberRequest{
		TenantID: tenant.TenantID, PrincipalID: primary.PrincipalID,
		Role: entity.RoleOwner, CreatedBy: primary.PrincipalID,
	})
	require.NoError(t, err)
	_, err = svc.AddMember(ctx, &service.AddMemberRequest{
		TenantID: tenant.TenantID, PrincipalID: coOwner.PrincipalID,
		Role: entity.RoleOwner, CreatedBy: primary.PrincipalID,
	})
	require.NoError(t, err)

	policy := service.MembershipPolicy{}
	err = policy.CanChangeRole(
		entity.RoleOwner, coOwner.PrincipalID, primary.PrincipalID,
		entity.RoleOwner, entity.RoleAdmin, tenant.OwnerPrincipalID,
	)
	require.ErrorIs(t, err, entity.ErrPrimaryOwnerImmutable)

	_, err = svc.UpdateMemberRole(ctx, tenant.TenantID, primary.PrincipalID, entity.RoleAdmin, 1, coOwner.PrincipalID, "req-po")
	require.ErrorIs(t, err, entity.ErrPrimaryOwnerImmutable)

	err = svc.RemoveMember(ctx, tenant.TenantID, primary.PrincipalID, coOwner.PrincipalID, "req-po")
	require.ErrorIs(t, err, entity.ErrPrimaryOwnerImmutable)
}

func TestAuditActorIsActorNotTarget_MemberRoleChanged(t *testing.T) {
	svc, audit := newTestServiceWithAudit()
	ctx := context.Background()

	owner, err := svc.ResolveOrCreatePrincipal(ctx, 21, "Owner")
	require.NoError(t, err)
	member, err := svc.ResolveOrCreatePrincipal(ctx, 22, "Member")
	require.NoError(t, err)

	tenant, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name:             "Audit",
		OwnerPrincipalID: owner.PrincipalID,
	})
	require.NoError(t, err)
	_, err = svc.AddMember(ctx, &service.AddMemberRequest{
		TenantID: tenant.TenantID, PrincipalID: owner.PrincipalID,
		Role: entity.RoleOwner, CreatedBy: owner.PrincipalID,
	})
	require.NoError(t, err)
	m, err := svc.AddMember(ctx, &service.AddMemberRequest{
		TenantID: tenant.TenantID, PrincipalID: member.PrincipalID,
		Role: entity.RoleMember, CreatedBy: owner.PrincipalID,
	})
	require.NoError(t, err)

	_, err = svc.UpdateMemberRole(ctx, tenant.TenantID, member.PrincipalID, entity.RoleAdmin, m.Revision, owner.PrincipalID, "req-audit")
	require.NoError(t, err)

	events := audit.byAction(entity.AuditMemberRoleChanged)
	require.Len(t, events, 1)
	assert.Equal(t, owner.PrincipalID, events[0].PrincipalID, "audit PrincipalID must be actor")
	assert.Equal(t, member.PrincipalID, events[0].Resource, "audit Resource must be target")
	assert.Equal(t, "req-audit", events[0].RequestID)
}

func newTestServiceWithAudit() (service.TenancyService, *memAuditRepo) {
	audit := newMemAuditRepo()
	svc := service.NewTenancyService(&service.Components{
		PrincipalRepo:  newMemPrincipalRepo(),
		TenantRepo:     newMemTenantRepo(),
		MembershipRepo: newMemMembershipRepo(),
		SpaceRefRepo:   newMemSpaceRefRepo(),
		AuditRepo:      audit,
	})
	return svc, audit
}

func (r *memAuditRepo) byAction(action string) []*entity.AuditEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entity.AuditEvent
	for _, e := range r.rows {
		if e.Action == action {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out
}
