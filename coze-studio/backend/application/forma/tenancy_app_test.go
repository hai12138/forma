/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

func TestPatchMember_ExpectedRevisionRequired(t *testing.T) {
	app := newAppService()
	ctx := withSession(100, "owner@example.com")

	boot, err := app.TenancySVC.Bootstrap(ctx, 100, "owner@example.com", 0)
	require.NoError(t, err)

	member, err := app.TenancySVC.ResolveOrCreatePrincipal(ctx, 101, "member@example.com")
	require.NoError(t, err)
	_, err = app.TenancySVC.AddMember(ctx, &tenancysvc.AddMemberRequest{
		TenantID:    boot.Tenant.TenantID,
		PrincipalID: member.PrincipalID,
		Role:        entity.RoleMember,
		CreatedBy:   boot.Principal.PrincipalID,
	})
	require.NoError(t, err)

	tc := &tenantctx.TenantContext{
		TenantID:       boot.Tenant.TenantID,
		PrincipalID:    boot.Principal.PrincipalID,
		MembershipRole: entity.RoleOwner,
		CozeUserID:     100,
		TenantStatus:   entity.TenantStatusActive,
	}
	ctx = tenantctx.WithTenantContext(ctx, tc)

	_, err = app.PatchMember(ctx, boot.Tenant.TenantID, member.PrincipalID, &formaapp.PatchMemberInput{
		Role:             entity.RoleAdmin,
		ExpectedRevision: 0,
	})
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.CodeTenantRequired, fe.Code)
	assert.Contains(t, fe.Msg, "expected_revision")
}

func TestPatchTenant_ExpectedRevisionRequired(t *testing.T) {
	app := newAppService()
	ctx := withSession(200, "owner2@example.com")

	boot, err := app.TenancySVC.Bootstrap(ctx, 200, "owner2@example.com", 0)
	require.NoError(t, err)

	tc := &tenantctx.TenantContext{
		TenantID:       boot.Tenant.TenantID,
		PrincipalID:    boot.Principal.PrincipalID,
		MembershipRole: entity.RoleOwner,
		CozeUserID:     200,
		TenantStatus:   entity.TenantStatusActive,
	}
	ctx = tenantctx.WithTenantContext(ctx, tc)

	name := "Renamed"
	_, err = app.PatchTenant(ctx, boot.Tenant.TenantID, &formaapp.PatchTenantInput{
		Name:             &name,
		ExpectedRevision: 0,
	})
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.CodeTenantRequired, fe.Code)
}

func TestAddMember_AdminCannotAddOwner(t *testing.T) {
	app := newAppService()
	ctx := withSession(300, "owner3@example.com")

	boot, err := app.TenancySVC.Bootstrap(ctx, 300, "owner3@example.com", 0)
	require.NoError(t, err)

	admin, err := app.TenancySVC.ResolveOrCreatePrincipal(ctx, 301, "admin@example.com")
	require.NoError(t, err)
	_, err = app.TenancySVC.AddMember(ctx, &tenancysvc.AddMemberRequest{
		TenantID: boot.Tenant.TenantID, PrincipalID: admin.PrincipalID,
		Role: entity.RoleAdmin, CreatedBy: boot.Principal.PrincipalID,
	})
	require.NoError(t, err)

	target, err := app.TenancySVC.ResolveOrCreatePrincipal(ctx, 302, "target@example.com")
	require.NoError(t, err)

	tc := &tenantctx.TenantContext{
		TenantID:       boot.Tenant.TenantID,
		PrincipalID:    admin.PrincipalID,
		MembershipRole: entity.RoleAdmin,
		CozeUserID:     301,
		TenantStatus:   entity.TenantStatusActive,
	}
	ctx = tenantctx.WithTenantContext(withSession(301, "admin@example.com"), tc)

	_, err = app.AddMember(ctx, boot.Tenant.TenantID, &formaapp.AddMemberInput{
		PrincipalID: target.PrincipalID,
		Role:        entity.RoleOwner,
	})
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.CodeMembershipForbidden, fe.Code)
}
