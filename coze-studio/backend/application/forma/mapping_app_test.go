/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datarepo "github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	tenantentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

type mappingAppStub struct {
	rows map[string]*dataentity.SemanticMapping
}

func (s *mappingAppStub) AnalyzeSemanticMappings(_ context.Context, in *datasvc.AnalyzeSemanticMappingsInput) (*datasvc.AnalyzeSemanticMappingsResult, error) {
	now := time.Now().UTC()
	return &datasvc.AnalyzeSemanticMappingsResult{Run: &dataentity.SemanticMappingAnalysisRun{
		AnalysisRunID: "run", TenantID: in.TenantID, BusinessID: in.BusinessID,
		BusinessModelRevision: in.BusinessModelRevision, Status: dataentity.AnalysisSucceeded,
		CreatedAt: now, UpdatedAt: now,
	}}, nil
}
func (*mappingAppStub) GetMappingAnalysisRun(_ context.Context, tenantID, id string) (*dataentity.SemanticMappingAnalysisRun, error) {
	if id != "run" {
		return nil, dataentity.ErrMappingAnalysisNotFound
	}
	now := time.Now().UTC()
	return &dataentity.SemanticMappingAnalysisRun{
		AnalysisRunID: id, TenantID: tenantID, BusinessID: "lab", BusinessModelRevision: 7,
		Status: dataentity.AnalysisSucceeded, CreatedAt: now, UpdatedAt: now,
	}, nil
}
func (*mappingAppStub) RetryFailedMappingAnalysis(context.Context, string, string, string) (*datasvc.AnalyzeSemanticMappingsResult, error) {
	return nil, dataentity.ErrMappingAnalysisNotFound
}
func (s *mappingAppStub) ListMappings(_ context.Context, tenantID, businessID string, revision int32, _ dataentity.MappingStatus) ([]*dataentity.SemanticMapping, error) {
	out := []*dataentity.SemanticMapping{}
	for _, row := range s.rows {
		if row.TenantID == tenantID && row.BusinessID == businessID && row.BusinessModelRevision == revision {
			out = append(out, row)
		}
	}
	return out, nil
}
func (s *mappingAppStub) GetMapping(_ context.Context, tenantID, id string) (*dataentity.SemanticMapping, error) {
	row, ok := s.rows[tenantID+"/"+id]
	if !ok {
		return nil, dataentity.ErrMappingNotFound
	}
	return row, nil
}
func (*mappingAppStub) CreateManualMapping(context.Context, *datasvc.ManualMappingInput) (*dataentity.SemanticMapping, error) {
	return nil, errors.New("unexpected mutation")
}
func (*mappingAppStub) ConfirmMapping(context.Context, string, string, string, string) (*dataentity.SemanticMapping, *dataentity.SemanticMappingDecision, error) {
	return nil, nil, errors.New("unexpected mutation")
}
func (*mappingAppStub) RejectMapping(context.Context, string, string, string, string) (*dataentity.SemanticMapping, *dataentity.SemanticMappingDecision, error) {
	return nil, nil, errors.New("unexpected mutation")
}
func (*mappingAppStub) EditConfirmMapping(context.Context, *datasvc.EditConfirmMappingInput) (*dataentity.SemanticMapping, *dataentity.SemanticMapping, *dataentity.SemanticMappingDecision, error) {
	return nil, nil, nil, errors.New("unexpected mutation")
}
func (*mappingAppStub) ListMappingDecisions(context.Context, string, string) ([]*dataentity.SemanticMappingDecision, error) {
	return nil, nil
}
func (*mappingAppStub) GetMappingCoverage(context.Context, string, string, int32) (*datasvc.MappingCoverage, error) {
	return &datasvc.MappingCoverage{}, nil
}

func TestMappingApplicationAuthorizationAndTenantIsolation(t *testing.T) {
	app := newAppService()
	app.DataSVC = datasvc.NewDataService(&datasvc.Components{Repo: datarepo.NewMemoryDataRepository()})
	stub := &mappingAppStub{rows: map[string]*dataentity.SemanticMapping{}}
	app.MappingSVC = stub

	ownerSession := withSession(7200, "mapping-owner@example.com")
	boot, err := app.TenancySVC.Bootstrap(ownerSession, 7200, "mapping-owner@example.com", 0)
	require.NoError(t, err)
	admin, err := app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 7201, "mapping-admin@example.com")
	require.NoError(t, err)
	member, err := app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 7202, "mapping-member@example.com")
	require.NoError(t, err)
	for _, item := range []struct {
		principal string
		role      tenantentity.MembershipRole
	}{{admin.PrincipalID, tenantentity.RoleAdmin}, {member.PrincipalID, tenantentity.RoleMember}} {
		_, err = app.TenancySVC.AddMember(ownerSession, &tenancysvc.AddMemberRequest{
			TenantID: boot.Tenant.TenantID, PrincipalID: item.principal, Role: item.role,
			CreatedBy: boot.Principal.PrincipalID,
		})
		require.NoError(t, err)
	}

	ctxFor := func(principal string, role tenantentity.MembershipRole) context.Context {
		return tenantctx.WithTenantContext(ownerSession, &tenantctx.TenantContext{
			TenantID: boot.Tenant.TenantID, PrincipalID: principal,
			MembershipRole: role, TenantStatus: tenantentity.TenantStatusActive,
		})
	}
	analyze := &formaapp.AnalyzeSemanticMappingsInput{
		BusinessModelRevision: 7, RequirementIDs: []string{"req"},
		SchemaSnapshotIDs: []string{"snap"}, ClientRequestID: "request",
	}
	for _, ctx := range []context.Context{
		ctxFor(boot.Principal.PrincipalID, tenantentity.RoleOwner),
		ctxFor(admin.PrincipalID, tenantentity.RoleAdmin),
	} {
		_, err = app.AnalyzeSemanticMappings(ctx, "lab", analyze)
		require.NoError(t, err)
	}

	now := time.Now().UTC()
	row := &dataentity.SemanticMapping{
		MappingID: "map", TenantID: boot.Tenant.TenantID, BusinessID: "lab",
		BusinessModelRevision: 7, CreatedAt: now, UpdatedAt: now,
	}
	stub.rows[boot.Tenant.TenantID+"/map"] = row
	stub.rows["other-tenant/foreign"] = &dataentity.SemanticMapping{MappingID: "foreign", TenantID: "other-tenant", BusinessID: "lab"}
	memberCtx := ctxFor(member.PrincipalID, tenantentity.RoleMember)
	_, err = app.ListSemanticMappings(memberCtx, "lab", 7, "")
	require.NoError(t, err)
	_, err = app.GetMappingAnalysisRun(memberCtx, "lab", "run")
	require.NoError(t, err)
	_, err = app.GetSemanticMappingCoverage(memberCtx, "lab", 7)
	require.NoError(t, err)

	mutations := []func() error{
		func() error { _, e := app.AnalyzeSemanticMappings(memberCtx, "lab", analyze); return e },
		func() error {
			_, e := app.CreateManualSemanticMapping(memberCtx, "lab", &formaapp.SemanticMappingInput{})
			return e
		},
		func() error { _, e := app.DecideSemanticMapping(memberCtx, "lab", "map", "confirm", nil); return e },
		func() error { _, e := app.DecideSemanticMapping(memberCtx, "lab", "map", "reject", nil); return e },
		func() error {
			_, e := app.EditConfirmSemanticMapping(memberCtx, "lab", "map", &formaapp.EditConfirmSemanticMappingInput{})
			return e
		},
	}
	for _, mutate := range mutations {
		fe, ok := formaerrors.AsFormaError(mutate())
		require.True(t, ok)
		require.Equal(t, formaerrors.CodeDataForbidden, fe.Code)
	}

	_, err = app.ListSemanticMappingDecisions(memberCtx, "lab", "foreign")
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeDataSemanticMappingNotFound, fe.Code)
}
