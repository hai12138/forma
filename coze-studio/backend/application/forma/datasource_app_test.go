/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	datasourcerepo "github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
	datasourcesvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

func TestDatasourceApplicationRequiresAdminForMutationAndAllowsMemberRead(t *testing.T) {
	app := newAppService()
	app.DatasourceSVC = datasourcesvc.NewDataSourceService(&datasourcesvc.SourceComponents{
		Repo:     datasourcerepo.NewMemoryDataSourceRepository(),
		Adapters: datasourcesvc.NewDefaultAdapterRegistry(),
	})
	ownerCtx := withSession(7100, "datasource-owner@example.com")
	boot, err := app.TenancySVC.Bootstrap(ownerCtx, 7100, "datasource-owner@example.com", 0)
	require.NoError(t, err)
	member, err := app.TenancySVC.ResolveOrCreatePrincipal(ownerCtx, 7101, "datasource-member@example.com")
	require.NoError(t, err)
	_, err = app.TenancySVC.AddMember(ownerCtx, &tenancysvc.AddMemberRequest{TenantID: boot.Tenant.TenantID, PrincipalID: member.PrincipalID, Role: entity.RoleMember, CreatedBy: boot.Principal.PrincipalID})
	require.NoError(t, err)

	ownerCtx = tenantctx.WithTenantContext(ownerCtx, &tenantctx.TenantContext{TenantID: boot.Tenant.TenantID, PrincipalID: boot.Principal.PrincipalID, MembershipRole: entity.RoleOwner, TenantStatus: entity.TenantStatusActive})
	created, err := app.CreateDataSource(ownerCtx, &formaapp.CreateDataSourceInput{Name: "warehouse", SourceType: "RELATIONAL_DATABASE"})
	require.NoError(t, err)

	memberCtx := tenantctx.WithTenantContext(ownerCtx, &tenantctx.TenantContext{TenantID: boot.Tenant.TenantID, PrincipalID: member.PrincipalID, MembershipRole: entity.RoleMember, TenantStatus: entity.TenantStatusActive})
	read, err := app.GetDataSource(memberCtx, created.SourceID)
	require.NoError(t, err)
	require.Equal(t, created.SourceID, read.SourceID)
	_, err = app.CreateDataSource(memberCtx, &formaapp.CreateDataSourceInput{Name: "forbidden", SourceType: "RELATIONAL_DATABASE"})
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeDataForbidden, fe.Code)
}
