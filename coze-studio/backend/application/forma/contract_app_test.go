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

type contractAppStub struct {
	contracts map[string]*dataentity.DataContract
	revisions map[string]*dataentity.DataContractRevision
}

func (s *contractAppStub) CreateContract(_ context.Context, in *datasvc.CreateContractInput) (*dataentity.DataContract, *dataentity.DataContractRevision, error) {
	now := time.Now().UTC()
	c := &dataentity.DataContract{ContractID: "ctr1", TenantID: in.TenantID, BusinessID: in.BusinessID, CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now}
	r := &dataentity.DataContractRevision{RevisionID: "crev1", TenantID: in.TenantID, BusinessID: in.BusinessID, ContractID: c.ContractID, Version: 1, Status: dataentity.ContractStatusDraft, CreatedAt: now, UpdatedAt: now}
	s.contracts[in.TenantID+"/"+c.ContractID] = c
	s.revisions[in.TenantID+"/"+r.RevisionID] = r
	return c, r, nil
}
func (s *contractAppStub) GetContract(_ context.Context, tenantID, id string) (*dataentity.DataContract, error) {
	v, ok := s.contracts[tenantID+"/"+id]
	if !ok {
		return nil, dataentity.ErrContractNotFound
	}
	return v, nil
}
func (s *contractAppStub) ListContracts(_ context.Context, tenantID, businessID string) ([]*dataentity.DataContract, error) {
	out := []*dataentity.DataContract{}
	for _, v := range s.contracts {
		if v.TenantID == tenantID && v.BusinessID == businessID {
			out = append(out, v)
		}
	}
	return out, nil
}
func (s *contractAppStub) GetRevision(_ context.Context, tenantID, id string) (*dataentity.DataContractRevision, error) {
	v, ok := s.revisions[tenantID+"/"+id]
	if !ok {
		return nil, dataentity.ErrContractRevisionNotFound
	}
	return v, nil
}
func (s *contractAppStub) ListRevisions(context.Context, string, string) ([]*dataentity.DataContractRevision, error) {
	return nil, nil
}
func (*contractAppStub) CreateRevision(context.Context, *datasvc.CreateRevisionInput) (*dataentity.DataContractRevision, error) {
	return nil, errors.New("unexpected mutation")
}
func (*contractAppStub) ValidateRevision(context.Context, string, string, string) (*dataentity.DataContractRevision, *dataentity.DataValidationResult, error) {
	return nil, nil, errors.New("unexpected mutation")
}
func (*contractAppStub) ActivateRevision(context.Context, string, string, string, string) (*dataentity.DataContractRevision, error) {
	return nil, errors.New("unexpected mutation")
}
func (*contractAppStub) DeprecateRevision(context.Context, string, string, string, string) (*dataentity.DataContractRevision, error) {
	return nil, errors.New("unexpected mutation")
}
func (*contractAppStub) ListValidationResults(context.Context, string, string) ([]*dataentity.DataValidationResult, error) {
	return nil, nil
}
func (*contractAppStub) EvaluateDrift(context.Context, *datasvc.EvaluateDriftInput) (*dataentity.DataDriftResult, *dataentity.DataContractRevision, error) {
	return nil, nil, errors.New("unexpected mutation")
}
func (*contractAppStub) ListDriftResults(context.Context, string, string) ([]*dataentity.DataDriftResult, error) {
	return nil, nil
}
func (*contractAppStub) EvaluateBusinessGap(context.Context, *datasvc.EvaluateGapInput) (*dataentity.DataContractGapResult, error) {
	return nil, errors.New("unexpected mutation")
}
func (*contractAppStub) ListGapResults(context.Context, string, string) ([]*dataentity.DataContractGapResult, error) {
	return nil, nil
}
func (*contractAppStub) ListLifecycleEvents(context.Context, string, string) ([]*dataentity.DataContractLifecycleEvent, error) {
	return nil, nil
}
func (*contractAppStub) BuildContractDescriptor(*dataentity.DataContractRevision) *dataentity.DataContractDescriptor {
	return nil
}
func (*contractAppStub) GetActiveContractDescriptor(context.Context, string, string) (*dataentity.DataContractDescriptor, error) {
	return nil, dataentity.ErrContractNotActive
}

func TestContractApplicationAuthorizationAndTenantIsolation(t *testing.T) {
	app := newAppService()
	app.DataSVC = datasvc.NewDataService(&datasvc.Components{Repo: datarepo.NewMemoryDataRepository()})
	stub := &contractAppStub{contracts: map[string]*dataentity.DataContract{}, revisions: map[string]*dataentity.DataContractRevision{}}
	app.ContractSVC = stub

	ownerSession := withSession(7300, "contract-owner@example.com")
	boot, err := app.TenancySVC.Bootstrap(ownerSession, 7300, "contract-owner@example.com", 0)
	require.NoError(t, err)
	admin, err := app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 7301, "contract-admin@example.com")
	require.NoError(t, err)
	member, err := app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 7302, "contract-member@example.com")
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
	createIn := &formaapp.CreateDataContractInput{
		BusinessModelRevision: 7, Name: "c", RequirementIDs: []string{"req"}, MappingIDs: []string{"map"},
	}
	for _, ctx := range []context.Context{
		ctxFor(boot.Principal.PrincipalID, tenantentity.RoleOwner),
		ctxFor(admin.PrincipalID, tenantentity.RoleAdmin),
	} {
		_, err = app.CreateDataContract(ctx, "lab", createIn)
		require.NoError(t, err)
	}

	now := time.Now().UTC()
	stub.contracts[boot.Tenant.TenantID+"/ctr1"] = &dataentity.DataContract{
		ContractID: "ctr1", TenantID: boot.Tenant.TenantID, BusinessID: "lab", CreatedAt: now, UpdatedAt: now,
	}
	stub.revisions[boot.Tenant.TenantID+"/crev1"] = &dataentity.DataContractRevision{
		RevisionID: "crev1", TenantID: boot.Tenant.TenantID, BusinessID: "lab", ContractID: "ctr1", CreatedAt: now, UpdatedAt: now,
	}
	stub.contracts["other-tenant/foreign"] = &dataentity.DataContract{ContractID: "foreign", TenantID: "other-tenant", BusinessID: "lab"}

	memberCtx := ctxFor(member.PrincipalID, tenantentity.RoleMember)
	_, err = app.ListDataContracts(memberCtx, "lab")
	require.NoError(t, err)
	_, err = app.GetDataContract(memberCtx, "lab", "ctr1")
	require.NoError(t, err)
	_, err = app.ListDataContractLifecycleEvents(memberCtx, "lab", "ctr1")
	require.NoError(t, err)

	mutations := []func() error{
		func() error { _, e := app.CreateDataContract(memberCtx, "lab", createIn); return e },
		func() error {
			_, e := app.CreateDataContractRevision(memberCtx, "lab", "ctr1", &formaapp.CreateDataContractRevisionInput{})
			return e
		},
		func() error { _, e := app.ValidateDataContractRevision(memberCtx, "lab", "ctr1", "crev1"); return e },
		func() error {
			_, e := app.ActivateDataContractRevision(memberCtx, "lab", "ctr1", "crev1", nil)
			return e
		},
		func() error {
			_, e := app.DeprecateDataContractRevision(memberCtx, "lab", "ctr1", "crev1", nil)
			return e
		},
		func() error {
			_, e := app.EvaluateDataContractDrift(memberCtx, "lab", "ctr1", "crev1", &formaapp.EvaluateDriftAppInput{})
			return e
		},
		func() error { _, e := app.EvaluateDataContractGap(memberCtx, "lab", "ctr1", "crev1"); return e },
	}
	for _, mutate := range mutations {
		fe, ok := formaerrors.AsFormaError(mutate())
		require.True(t, ok)
		require.Equal(t, formaerrors.CodeDataForbidden, fe.Code)
	}

	_, err = app.GetDataContract(memberCtx, "lab", "foreign")
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeDataContractNotFound, fe.Code)
}
