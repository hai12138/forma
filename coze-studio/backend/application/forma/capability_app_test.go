/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	tenantentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

type capturingAuditRepo struct {
	mu   sync.Mutex
	rows []*tenantentity.AuditEvent
}

func (r *capturingAuditRepo) Create(_ context.Context, e *tenantentity.AuditEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *e
	r.rows = append(r.rows, &cp)
	return nil
}

func (r *capturingAuditRepo) snapshot() []*tenantentity.AuditEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*tenantentity.AuditEvent, len(r.rows))
	copy(out, r.rows)
	return out
}

type capabilityBusinessStub struct {
	mu   sync.Mutex
	byID map[string]*businessentity.BusinessModel // tenant|business
}

func newCapabilityBusinessStub() *capabilityBusinessStub {
	return &capabilityBusinessStub{byID: map[string]*businessentity.BusinessModel{}}
}

func (s *capabilityBusinessStub) put(tenantID, businessID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.byID[tenantID+"|"+businessID] = &businessentity.BusinessModel{
		TenantID: tenantID, BusinessID: businessID, AssetID: businessID,
		CurrentRevision: 1, SchemaVersion: "1.0", CreatedAt: now, UpdatedAt: now,
	}
}

func (s *capabilityBusinessStub) Get(_ context.Context, tenantID, businessID string) (*businessentity.BusinessModel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.byID[tenantID+"|"+businessID]
	if !ok {
		return nil, businessentity.ErrNotFound
	}
	cp := *v
	return &cp, nil
}
func (s *capabilityBusinessStub) List(context.Context, string) ([]*businessentity.BusinessModel, error) {
	return nil, nil
}
func (s *capabilityBusinessStub) InitBusiness(context.Context, string, string, string, string, *businessentity.SemanticModel, string) (*businessentity.BusinessModel, *businessentity.BusinessModelRevision, *businessentity.BusinessModelLayout, error) {
	return nil, nil, nil, businessentity.ErrInvalidModel
}
func (s *capabilityBusinessStub) GetModel(context.Context, string, string) (*businessentity.BusinessModel, *businessentity.SemanticModel, *businessentity.BusinessModelRevision, error) {
	return nil, nil, nil, businessentity.ErrNotFound
}
func (s *capabilityBusinessStub) SaveModel(context.Context, string, string, string, int32, *businessentity.SemanticModel, string) (*businessentity.BusinessModelRevision, bool, error) {
	return nil, false, businessentity.ErrInvalidModel
}
func (s *capabilityBusinessStub) ListRevisions(context.Context, string, string) ([]*businessentity.BusinessModelRevision, error) {
	return nil, nil
}
func (s *capabilityBusinessStub) GetRevision(context.Context, string, string, int32) (*businessentity.BusinessModelRevision, *businessentity.SemanticModel, error) {
	return nil, nil, businessentity.ErrRevisionNotFound
}
func (s *capabilityBusinessStub) Diff(context.Context, string, string, int32, int32) (*businessentity.BusinessModelDiff, *businessentity.BusinessImpactSummary, error) {
	return nil, nil, businessentity.ErrNotFound
}
func (s *capabilityBusinessStub) GetLayout(context.Context, string, string) (*businessentity.BusinessModelLayout, *businessentity.ViewLayout, error) {
	return nil, nil, businessentity.ErrNotFound
}
func (s *capabilityBusinessStub) SaveLayout(context.Context, string, string, string, int32, int32, *businessentity.ViewLayout) (*businessentity.BusinessModelLayout, error) {
	return nil, businessentity.ErrInvalidModel
}

var _ businesssvc.BusinessService = (*capabilityBusinessStub)(nil)

type capabilityAppHarness struct {
	app      *formaapp.ApplicationService
	audit    *capturingAuditRepo
	biz      *capabilityBusinessStub
	bm       *capsvc.FakeBusinessModelPort
	contract *capsvc.FakeContractPort
	uow      *capsvc.MemoryUnitOfWork
}

func newCapabilityAppHarness() *capabilityAppHarness {
	audit := &capturingAuditRepo{}
	biz := newCapabilityBusinessStub()
	uow := capsvc.NewMemoryUnitOfWork()
	bm := capsvc.NewFakeBusinessModelPort()
	contracts := capsvc.NewFakeContractPort()
	app := &formaapp.ApplicationService{
		TenancySVC: tenancysvc.NewTenancyService(&tenancysvc.Components{
			PrincipalRepo:  newMemPrincipalRepo(),
			TenantRepo:     newMemTenantRepo(),
			MembershipRepo: newMemMembershipRepo(),
			SpaceRefRepo:   newMemSpaceRefRepo(),
			AuditRepo:      audit,
		}),
		BusinessSVC: biz,
		CapabilitySVC: capsvc.NewCapabilityService(&capsvc.Components{
			UoW:       uow,
			Generator: &capsvc.DeterministicFakeGenerator{Proposals: []capentity.SemanticPayload{fixture.LaboratoryCommandCapability()}},
			Business:  bm,
			Contract:  contracts,
		}),
	}
	return &capabilityAppHarness{app: app, audit: audit, biz: biz, bm: bm, contract: contracts, uow: uow}
}

func (h *capabilityAppHarness) seedBiz(tenantID, businessID string) {
	h.biz.put(tenantID, businessID)
}

func (h *capabilityAppHarness) seedPorts(tenantID, businessID string, rev int32, payload capentity.SemanticPayload) {
	h.bm.Put(&capsvc.BusinessModelRevisionEvidence{
		TenantID: tenantID, BusinessID: businessID, Revision: rev, ContentDigest: "bm-digest",
	})
	h.bm.SetCurrentRevision(tenantID, businessID, rev)
	for _, b := range payload.DataContractBindings {
		fields := make([]capsvc.ContractLogicalField, 0)
		filters := make([]capsvc.ContractFilterFieldSpec, 0)
		seen := map[string]struct{}{}
		for _, m := range b.LogicalFieldMappings {
			if _, ok := seen[m.ContractLogicalKey]; ok {
				continue
			}
			seen[m.ContractLogicalKey] = struct{}{}
			lt := "STRING"
			for _, f := range append(append([]capentity.LogicalField{}, payload.InputSchema.Fields...), payload.OutputSchema.Fields...) {
				if f.LogicalKey == m.CapabilityLogicalKey {
					lt = f.LogicalType
					break
				}
			}
			fields = append(fields, capsvc.ContractLogicalField{LogicalKey: m.ContractLogicalKey, LogicalType: lt})
			filters = append(filters, capsvc.ContractFilterFieldSpec{
				LogicalKey: m.ContractLogicalKey, Operators: []string{"EQ", "NE", "GT", "GTE", "LT", "LTE", "IN"},
			})
		}
		h.contract.Put(&capsvc.ContractLogicalDescriptor{
			TenantID: tenantID, BusinessID: businessID,
			ContractID: b.DataContractID, RevisionID: b.DataContractRevisionID, Version: b.DataContractVersion,
			BusinessModelRevision: rev, Status: "ACTIVE",
			LogicalSchema: fields, QueryCapabilities: []string{"READ", "LIST", "FILTER"}, FilterSchema: filters,
			PaginationPolicy: capsvc.ContractPaginationPolicy{DefaultLimit: 20, MaxLimit: 100},
		})
	}
}

func ctxCapability(session context.Context, tenantID, principalID string, role tenantentity.MembershipRole, cozeUserID int64) context.Context {
	return tenantctx.WithTenantContext(session, &tenantctx.TenantContext{
		TenantID: tenantID, PrincipalID: principalID, MembershipRole: role,
		CozeUserID: cozeUserID, TenantStatus: tenantentity.TenantStatusActive,
	})
}

func TestCapabilityApp_RoleMatrixAndIsolation(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8100, "cap-owner@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8100, "cap-owner@example.com", 0)
	require.NoError(t, err)
	admin, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 8101, "cap-admin@example.com")
	require.NoError(t, err)
	member, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 8102, "cap-member@example.com")
	require.NoError(t, err)
	viewer, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 8103, "cap-viewer@example.com")
	require.NoError(t, err)
	for _, item := range []struct {
		principal string
		role      tenantentity.MembershipRole
	}{
		{admin.PrincipalID, tenantentity.RoleAdmin},
		{member.PrincipalID, tenantentity.RoleMember},
		{viewer.PrincipalID, tenantentity.RoleViewer},
	} {
		_, err = h.app.TenancySVC.AddMember(ownerSession, &tenancysvc.AddMemberRequest{
			TenantID: boot.Tenant.TenantID, PrincipalID: item.principal, Role: item.role,
			CreatedBy: boot.Principal.PrincipalID,
		})
		require.NoError(t, err)
	}

	tenantID := boot.Tenant.TenantID
	bizID := "lab"
	h.seedBiz(tenantID, bizID)

	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8100)
	adminCtx := ctxCapability(ownerSession, tenantID, admin.PrincipalID, tenantentity.RoleAdmin, 8101)
	memberCtx := ctxCapability(ownerSession, tenantID, member.PrincipalID, tenantentity.RoleMember, 8102)
	viewerCtx := ctxCapability(ownerSession, tenantID, viewer.PrincipalID, tenantentity.RoleViewer, 8103)

	cmd := fixture.LaboratoryCommandCapability()
	createIn := &formaapp.CreateCapabilityInput{Payload: cmd}

	// OWNER / ADMIN can create
	created, err := h.app.CreateCapability(ownerCtx, bizID, createIn)
	require.NoError(t, err)
	require.NotNil(t, created.Capability)
	capID := created.Capability.CapabilityID
	revID := created.Revision.RevisionID

	_, err = h.app.CreateCapability(adminCtx, bizID, &formaapp.CreateCapabilityInput{
		CapabilityID: "cap-admin-2", Payload: cmd,
	})
	require.NoError(t, err)

	// MEMBER / VIEWER can list/get
	listed, err := h.app.ListCapabilities(memberCtx, bizID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listed), 1)
	_, err = h.app.GetCapability(viewerCtx, bizID, capID)
	require.NoError(t, err)
	_, err = h.app.ListCapabilityRevisions(memberCtx, bizID, capID)
	require.NoError(t, err)
	_, err = h.app.GetCapabilityRevision(viewerCtx, bizID, capID, revID)
	require.NoError(t, err)
	_, err = h.app.ListCapabilityDecisions(memberCtx, bizID, capID)
	require.NoError(t, err)

	// MEMBER / VIEWER denied on mutations
	denyMutations := []func() error{
		func() error { _, e := h.app.CreateCapability(memberCtx, bizID, createIn); return e },
		func() error {
			_, e := h.app.DeriveCapability(memberCtx, bizID, capID, &formaapp.DeriveCapabilityInput{
				SourceRevisionID: revID, ClientRequestID: "d1", Payload: cmd,
			})
			return e
		},
		func() error {
			_, e := h.app.EditCapability(viewerCtx, bizID, capID, &formaapp.EditCapabilityInput{
				SourceRevisionID: revID, ClientRequestID: "e1", Payload: cmd,
			})
			return e
		},
		func() error {
			_, e := h.app.StartCapabilityAnalysis(memberCtx, bizID, &formaapp.StartCapabilityAnalysisInput{
				BusinessModelRevision: 1, ClientRequestID: "a1",
				Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
			})
			return e
		},
		func() error { _, e := h.app.ValidateCapabilityRevision(memberCtx, bizID, revID); return e },
		func() error {
			_, e := h.app.ConfirmCapabilityProposal(memberCtx, bizID, "missing-prop", nil)
			return e
		},
		func() error {
			_, e := h.app.EditConfirmCapabilityProposal(viewerCtx, bizID, "missing-prop", &formaapp.EditConfirmCapabilityProposalInput{
				EffectivePayload: cmd,
			})
			return e
		},
		func() error {
			_, e := h.app.RejectCapabilityProposal(memberCtx, bizID, "missing-prop", nil)
			return e
		},
		func() error {
			_, e := h.app.ActivateCapabilityRevision(memberCtx, bizID, revID, nil)
			return e
		},
		func() error {
			_, e := h.app.ActivateCapabilityRevision(adminCtx, bizID, revID, nil)
			return e
		},
		func() error {
			_, e := h.app.DeprecateCapabilityRevision(adminCtx, bizID, revID, nil)
			return e
		},
	}
	for _, fn := range denyMutations {
		fe, ok := formaerrors.AsFormaError(fn())
		require.True(t, ok)
		require.Equal(t, formaerrors.CodeCapabilityForbidden, fe.Code)
	}

	// ADMIN can validate (after port seed); OWNER can activate
	h.seedPorts(tenantID, bizID, 1, cmd)
	validated, err := h.app.ValidateCapabilityRevision(adminCtx, bizID, revID)
	require.NoError(t, err)
	require.Equal(t, string(capentity.RevisionValidated), validated.Revision.Status)

	activated, err := h.app.ActivateCapabilityRevision(ownerCtx, bizID, revID, &formaapp.CapabilityReasonInput{Reason: "go-live"})
	require.NoError(t, err)
	require.Equal(t, string(capentity.RevisionActive), activated.Status)
	got, err := h.app.GetCapability(ownerCtx, bizID, capID)
	require.NoError(t, err)
	require.Equal(t, revID, got.ActiveRevisionID)

	// Cross-tenant: wrong tenant id in context vs foreign capability
	foreignCtx := ctxCapability(ownerSession, "other-tenant", boot.Principal.PrincipalID, tenantentity.RoleOwner, 8100)
	_, err = h.app.GetCapability(foreignCtx, bizID, capID)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.True(t, fe.Code == formaerrors.CodeTenantForbidden || fe.Code == formaerrors.CodeCapabilityNotFound)

	// Business path mismatch fail-closed
	_, err = h.app.GetCapability(ownerCtx, "other-biz", capID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.True(t, fe.Code == formaerrors.CodeBusinessNotFound || fe.Code == formaerrors.CodeCapabilityNotFound)

	h.seedBiz(tenantID, "other-biz")
	_, err = h.app.GetCapability(ownerCtx, "other-biz", capID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityNotFound, fe.Code)
}

func TestCapabilityApp_InactiveAndSuperAdminDenied(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8200, "cap-owner2@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8200, "cap-owner2@example.com", 0)
	require.NoError(t, err)
	member, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 8201, "cap-member2@example.com")
	require.NoError(t, err)
	_, err = h.app.TenancySVC.AddMember(ownerSession, &tenancysvc.AddMemberRequest{
		TenantID: boot.Tenant.TenantID, PrincipalID: member.PrincipalID, Role: tenantentity.RoleMember,
		CreatedBy: boot.Principal.PrincipalID,
	})
	require.NoError(t, err)
	h.seedBiz(boot.Tenant.TenantID, "lab")

	// Remove membership → inactive denied even with forged ACTIVE role
	require.NoError(t, h.app.TenancySVC.RemoveMember(ownerSession, boot.Tenant.TenantID, member.PrincipalID, boot.Principal.PrincipalID, "rm1"))
	inactiveCtx := ctxCapability(ownerSession, boot.Tenant.TenantID, member.PrincipalID, tenantentity.RoleMember, 8201)
	_, err = h.app.ListCapabilities(inactiveCtx, "lab")
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeTenantForbidden, fe.Code)

	// SUPER_ADMIN-shaped principal without tenant membership
	sa, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 8299, "super@example.com")
	require.NoError(t, err)
	saCtx := ctxCapability(ownerSession, boot.Tenant.TenantID, sa.PrincipalID, tenantentity.RoleOwner, 8299)
	_, err = h.app.ListCapabilities(saCtx, "lab")
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeTenantForbidden, fe.Code)
}

func TestCapabilityApp_AnalyzeConfirmIdempotencyAndAuditIsolation(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8300, "cap-audit@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8300, "cap-audit@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8300)

	analysis1, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "idem-an-1",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.NotEmpty(t, analysis1.Proposals)
	propID := analysis1.Proposals[0].ProposalID

	analysis2, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "idem-an-1",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Equal(t, analysis1.AnalysisRun.AnalysisRunID, analysis2.AnalysisRun.AnalysisRunID)

	gotRun, err := h.app.GetCapabilityAnalysis(ownerCtx, "lab", analysis1.AnalysisRun.AnalysisRunID)
	require.NoError(t, err)
	require.Equal(t, analysis1.AnalysisRun.AnalysisRunID, gotRun.AnalysisRunID)

	rev, err := h.app.ConfirmCapabilityProposal(ownerCtx, "lab", propID, &formaapp.ConfirmCapabilityProposalInput{
		ClientRequestID: "confirm-1", CapabilityID: "cap-ai-1",
	})
	require.NoError(t, err)
	require.NotNil(t, rev)

	rev2, err := h.app.ConfirmCapabilityProposal(ownerCtx, "lab", propID, &formaapp.ConfirmCapabilityProposalInput{
		ClientRequestID: "confirm-1", CapabilityID: "cap-ai-1",
	})
	require.NoError(t, err)
	require.Equal(t, rev.RevisionID, rev2.RevisionID)

	// Derive idempotency
	cmd := fixture.LaboratoryCommandCapability()
	created, err := h.app.CreateCapability(ownerCtx, "lab", &formaapp.CreateCapabilityInput{Payload: cmd})
	require.NoError(t, err)
	d1, err := h.app.DeriveCapability(ownerCtx, "lab", created.Capability.CapabilityID, &formaapp.DeriveCapabilityInput{
		SourceRevisionID: created.Revision.RevisionID, ClientRequestID: "derive-idem", Payload: cmd,
	})
	require.NoError(t, err)
	d2, err := h.app.DeriveCapability(ownerCtx, "lab", created.Capability.CapabilityID, &formaapp.DeriveCapabilityInput{
		SourceRevisionID: created.Revision.RevisionID, ClientRequestID: "derive-idem", Payload: cmd,
	})
	require.NoError(t, err)
	require.Equal(t, d1.Revision.RevisionID, d2.Revision.RevisionID)

	for _, ev := range h.audit.snapshot() {
		blob := strings.ToLower(ev.Action + " " + ev.Resource + " " + ev.RequestID + " " + ev.PrincipalID)
		for _, forbidden := range []string{"password", "token", "authorization", "cookie", "secret", "api_key", "bearer"} {
			require.NotContains(t, blob, forbidden)
		}
		if strings.HasPrefix(ev.Action, "capability.") {
			require.NotContains(t, ev.Resource, " ")
			require.NotEmpty(t, ev.Resource)
		}
	}

	// Malformed / not-found mapping
	_, err = h.app.GetCapability(ownerCtx, "lab", "missing-cap")
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityNotFound, fe.Code)

	_, err = h.app.CreateCapability(ownerCtx, "lab", nil)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityInvalidPayload, fe.Code)
}

func TestCapabilityApp_ActivateDeprecateActivePointer(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8400, "cap-act@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8400, "cap-act@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8400)

	cmd := fixture.LaboratoryCommandCapability()
	c1, err := h.app.CreateCapability(ownerCtx, "lab", &formaapp.CreateCapabilityInput{CapabilityID: "cap-ptr", Payload: cmd})
	require.NoError(t, err)
	h.seedPorts(tenantID, "lab", 1, cmd)
	_, err = h.app.ValidateCapabilityRevision(ownerCtx, "lab", c1.Revision.RevisionID)
	require.NoError(t, err)
	_, err = h.app.ActivateCapabilityRevision(ownerCtx, "lab", c1.Revision.RevisionID, &formaapp.CapabilityReasonInput{Reason: "v1"})
	require.NoError(t, err)

	d, err := h.app.DeriveCapability(ownerCtx, "lab", "cap-ptr", &formaapp.DeriveCapabilityInput{
		SourceRevisionID: c1.Revision.RevisionID, ClientRequestID: "d-ptr", Payload: cmd,
	})
	require.NoError(t, err)
	h.seedPorts(tenantID, "lab", 1, cmd)
	_, err = h.app.ValidateCapabilityRevision(ownerCtx, "lab", d.Revision.RevisionID)
	require.NoError(t, err)
	_, err = h.app.ActivateCapabilityRevision(ownerCtx, "lab", d.Revision.RevisionID, &formaapp.CapabilityReasonInput{Reason: "v2"})
	require.NoError(t, err)

	got, err := h.app.GetCapability(ownerCtx, "lab", "cap-ptr")
	require.NoError(t, err)
	require.Equal(t, d.Revision.RevisionID, got.ActiveRevisionID)

	_, err = h.app.DeprecateCapabilityRevision(ownerCtx, "lab", d.Revision.RevisionID, &formaapp.CapabilityReasonInput{Reason: "retire"})
	require.NoError(t, err)
	got, err = h.app.GetCapability(ownerCtx, "lab", "cap-ptr")
	require.NoError(t, err)
	require.Empty(t, got.ActiveRevisionID)

	vals, err := h.app.ListCapabilityValidations(ownerCtx, "lab", d.Revision.RevisionID)
	require.NoError(t, err)
	require.NotEmpty(t, vals)
}

func TestCapabilityApp_DTOExcludesSecrets(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8500, "cap-dto@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8500, "cap-dto@example.com", 0)
	require.NoError(t, err)
	h.seedBiz(boot.Tenant.TenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, boot.Tenant.TenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8500)

	payload := fixture.LaboratoryFlowCapability()
	created, err := h.app.CreateCapability(ownerCtx, "lab", &formaapp.CreateCapabilityInput{Payload: payload})
	require.NoError(t, err)
	raw, err := json.Marshal(created)
	require.NoError(t, err)
	body := string(raw)
	for _, forbidden := range []string{
		"password", "Authorization", "Bearer", "api_key", "connection_id", "schema_snapshot_id",
		"credential", "secret_ref", "jdbc",
	} {
		require.NotContains(t, body, forbidden)
	}
	require.Contains(t, body, "data_contract_id")
	require.Contains(t, body, "logical_field_mappings")
}

func TestMapDomainError_CapabilityCodes(t *testing.T) {
	cases := []struct {
		err  error
		code int32
		key  string
	}{
		{capentity.ErrNotFound, formaerrors.CodeCapabilityNotFound, formaerrors.KeyCapabilityNotFound},
		{capentity.ErrRevisionNotFound, formaerrors.CodeCapabilityRevisionNotFound, formaerrors.KeyCapabilityRevisionNotFound},
		{capentity.ErrProposalNotFound, formaerrors.CodeCapabilityProposalNotFound, formaerrors.KeyCapabilityProposalNotFound},
		{capentity.ErrCrossTenant, formaerrors.CodeCapabilityForbidden, formaerrors.KeyCapabilityForbidden},
		{capentity.ErrConflict, formaerrors.CodeCapabilityConflict, formaerrors.KeyCapabilityConflict},
		{capentity.ErrConsistency, formaerrors.CodeCapabilityConflict, formaerrors.KeyCapabilityConflict},
		{capentity.ErrIdempotencyConflict, formaerrors.CodeCapabilityIdempotencyConflict, formaerrors.KeyCapabilityIdempotencyConflict},
		{capentity.ErrActiveConflict, formaerrors.CodeCapabilityActiveConflict, formaerrors.KeyCapabilityActiveConflict},
		{capentity.ErrInvalidPayload, formaerrors.CodeCapabilityInvalidPayload, formaerrors.KeyCapabilityInvalidPayload},
		{capentity.ErrNotConfigured, formaerrors.CodeCapabilityNotConfigured, formaerrors.KeyCapabilityNotConfigured},
	}
	for _, tc := range cases {
		fe := formaerrors.MapDomainError(tc.err)
		require.Equal(t, tc.code, fe.Code, tc.err.Error())
		require.Equal(t, tc.key, fe.Key, tc.err.Error())
		require.NotContains(t, strings.ToLower(fe.Msg), "password")
		require.NotContains(t, strings.ToLower(fe.Msg), "token")
	}
}
