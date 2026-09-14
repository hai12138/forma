/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

func assertCapabilityProjectionFields(t *testing.T, dto *formaapp.CapabilityDTO, wantName, wantSemVer, wantStatus, wantDigest string) {
	t.Helper()
	require.NotNil(t, dto)
	require.Equal(t, wantName, dto.Name)
	require.Equal(t, wantSemVer, dto.SemanticVersion)
	require.Equal(t, wantStatus, dto.AssetStatus)
	require.Equal(t, wantDigest, dto.ContentDigest)
	require.NotEmpty(t, dto.Name)
	require.NotEmpty(t, dto.SemanticVersion)
	require.NotEmpty(t, dto.AssetStatus)
	require.NotEmpty(t, dto.ContentDigest)
}

func assertDTOJSONSecretFree(t *testing.T, v any) {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	body := strings.ToLower(string(raw))
	for _, forbidden := range []string{
		"password", "authorization", "bearer", "api_key", "connection_id",
		"physical_locator", "snapshot_id", "schema_snapshot_id", "credential",
		"secret_ref", "jdbc",
	} {
		require.NotContains(t, body, forbidden, "DTO JSON must not contain %q", forbidden)
	}
}

func TestF7_App_ListGetCreate_ProjectionFields(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8700, "f7-proj@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8700, "f7-proj@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8700)

	cmd := fixture.LaboratoryCommandCapability()
	created, err := h.app.CreateCapability(ownerCtx, "lab", &formaapp.CreateCapabilityInput{
		CapabilityID: "cap-f7-proj", Payload: cmd,
	})
	require.NoError(t, err)
	require.NotNil(t, created.Capability)
	require.NotNil(t, created.Revision)

	asset, err := h.uow.AssetsView().GetCapabilityAsset(context.Background(), tenantID, "cap-f7-proj")
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetKindCapability, asset.Kind)

	// Create fills projection via ProjectCapabilityAssetRef — DTO must match persisted AssetRef
	// and pure projection (same digest/status/name/semver).
	want, err := capsvc.ProjectCapabilityAssetRef(
		&capentity.BusinessCapability{
			CapabilityID: "cap-f7-proj", TenantID: tenantID, BusinessID: "lab",
		},
		[]*capentity.BusinessCapabilityRevision{{
			RevisionID: created.Revision.RevisionID, CapabilityID: "cap-f7-proj",
			TenantID: tenantID, BusinessID: "lab", Version: 1,
			Status: capentity.RevisionDraft, Name: cmd.Name, Description: cmd.Description,
			CapabilityKind: cmd.CapabilityKind, BusinessModelRevision: cmd.BusinessModelRevision,
			InputSchema: cmd.InputSchema, OutputSchema: cmd.OutputSchema,
			Preconditions: cmd.Preconditions, Effects: cmd.Effects,
			DataContractBindings: cmd.DataContractBindings,
			QueryOperation:       cmd.QueryOperation, OutputCardinality: cmd.OutputCardinality,
			Source: capentity.SourceManualCreated,
		}},
	)
	require.NoError(t, err)
	require.Equal(t, want.Name, asset.Name)
	require.Equal(t, want.SemanticVersion, asset.SemanticVersion)
	require.Equal(t, want.ContentDigest, asset.ContentDigest)
	require.Equal(t, want.Status, asset.Status)

	assertCapabilityProjectionFields(t, created.Capability, want.Name, want.SemanticVersion, string(want.Status), want.ContentDigest)
	assertDTOJSONSecretFree(t, created)

	got, err := h.app.GetCapability(ownerCtx, "lab", "cap-f7-proj")
	require.NoError(t, err)
	assertCapabilityProjectionFields(t, got, want.Name, want.SemanticVersion, string(want.Status), want.ContentDigest)
	assertDTOJSONSecretFree(t, got)

	listed, err := h.app.ListCapabilities(ownerCtx, "lab")
	require.NoError(t, err)
	require.NotEmpty(t, listed)
	var found *formaapp.CapabilityDTO
	for _, row := range listed {
		if row.CapabilityID == "cap-f7-proj" {
			found = row
			break
		}
	}
	require.NotNil(t, found)
	assertCapabilityProjectionFields(t, found, want.Name, want.SemanticVersion, string(want.Status), want.ContentDigest)
	assertDTOJSONSecretFree(t, listed)
}

func TestF7_App_AssetStatusMatrix_DraftVerifiedReleasedDeprecated(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8701, "f7-status@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8701, "f7-status@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8701)
	cmd := fixture.LaboratoryCommandCapability()

	created, err := h.app.CreateCapability(ownerCtx, "lab", &formaapp.CreateCapabilityInput{
		CapabilityID: "cap-f7-st", Payload: cmd,
	})
	require.NoError(t, err)
	require.Equal(t, "DRAFT", created.Capability.AssetStatus)

	h.seedPorts(tenantID, "lab", 1, cmd)
	validated, err := h.app.ValidateCapabilityRevision(ownerCtx, "lab", created.Revision.RevisionID)
	require.NoError(t, err)
	got, err := h.app.GetCapability(ownerCtx, "lab", "cap-f7-st")
	require.NoError(t, err)
	require.Equal(t, "VERIFIED", got.AssetStatus)
	require.Equal(t, string(capentity.RevisionValidated), validated.Revision.Status)

	_, err = h.app.ActivateCapabilityRevision(ownerCtx, "lab", created.Revision.RevisionID, &formaapp.CapabilityReasonInput{Reason: "live"})
	require.NoError(t, err)
	got, err = h.app.GetCapability(ownerCtx, "lab", "cap-f7-st")
	require.NoError(t, err)
	require.Equal(t, "RELEASED", got.AssetStatus)

	_, err = h.app.DeprecateCapabilityRevision(ownerCtx, "lab", created.Revision.RevisionID, &formaapp.CapabilityReasonInput{Reason: "retire"})
	require.NoError(t, err)
	got, err = h.app.GetCapability(ownerCtx, "lab", "cap-f7-st")
	require.NoError(t, err)
	require.Equal(t, "DEPRECATED", got.AssetStatus)
}

func TestF7_App_ListGet_FailClosedMissingAssetRefMappedConflict(t *testing.T) {
	// Domain F7 covers Kind/AssetID/duplicate corruption. App asserts projection fail-closed
	// surfaces Cap conflict (409) via MapDomainError — never invent name from capability_id.
	h := newCapabilityAppHarness()
	ownerSession := withSession(8702, "f7-fc@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8702, "f7-fc@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8702)
	cmd := fixture.LaboratoryCommandCapability()

	_, err = h.app.CreateCapability(ownerCtx, "lab", &formaapp.CreateCapabilityInput{
		CapabilityID: "cap-f7-fc", Payload: cmd,
	})
	require.NoError(t, err)

	h.app.CapabilitySVC = &f7ConsistencyCapSVC{inner: h.app.CapabilitySVC}
	_, err = h.app.GetCapability(ownerCtx, "lab", "cap-f7-fc")
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityConflict, fe.Code)
	_, err = h.app.ListCapabilities(ownerCtx, "lab")
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityConflict, fe.Code)
}

// f7ConsistencyCapSVC forces Get/List to surface ErrConsistency (missing / invalid AssetRef).
type f7ConsistencyCapSVC struct {
	inner capsvc.CapabilityService
}

func (s *f7ConsistencyCapSVC) ManualCreate(ctx context.Context, in *capsvc.ManualCreateInput) (*capentity.BusinessCapability, *capentity.BusinessCapabilityRevision, error) {
	return s.inner.ManualCreate(ctx, in)
}
func (s *f7ConsistencyCapSVC) DeriveRevision(ctx context.Context, in *capsvc.DeriveInput) (*capentity.BusinessCapabilityRevision, *capentity.CapabilityDecision, error) {
	return s.inner.DeriveRevision(ctx, in)
}
func (s *f7ConsistencyCapSVC) StartAnalysis(ctx context.Context, in *capsvc.StartAnalysisInput) (*capsvc.AnalysisResult, error) {
	return s.inner.StartAnalysis(ctx, in)
}
func (s *f7ConsistencyCapSVC) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*capentity.CapabilityAnalysisRun, error) {
	return s.inner.GetAnalysisRun(ctx, tenantID, analysisRunID)
}
func (s *f7ConsistencyCapSVC) RetryFailedAnalysis(ctx context.Context, tenantID, analysisRunID, actorID string) (*capsvc.AnalysisResult, error) {
	return s.inner.RetryFailedAnalysis(ctx, tenantID, analysisRunID, actorID)
}
func (s *f7ConsistencyCapSVC) ConfirmProposal(ctx context.Context, in *capsvc.ConfirmInput) (*capentity.BusinessCapabilityRevision, error) {
	return s.inner.ConfirmProposal(ctx, in)
}
func (s *f7ConsistencyCapSVC) EditConfirmProposal(ctx context.Context, in *capsvc.EditConfirmInput) (*capentity.BusinessCapabilityRevision, error) {
	return s.inner.EditConfirmProposal(ctx, in)
}
func (s *f7ConsistencyCapSVC) RejectProposal(ctx context.Context, in *capsvc.RejectInput) (*capentity.CapabilityDecision, error) {
	return s.inner.RejectProposal(ctx, in)
}
func (s *f7ConsistencyCapSVC) GetCapability(context.Context, string, string) (*capentity.BusinessCapability, error) {
	return nil, capentity.ErrConsistency
}
func (s *f7ConsistencyCapSVC) ListCapabilities(context.Context, string, string) ([]*capentity.BusinessCapability, error) {
	return nil, capentity.ErrConsistency
}
func (s *f7ConsistencyCapSVC) GetRevision(ctx context.Context, tenantID, revisionID string) (*capentity.BusinessCapabilityRevision, error) {
	return s.inner.GetRevision(ctx, tenantID, revisionID)
}
func (s *f7ConsistencyCapSVC) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*capentity.BusinessCapabilityRevision, error) {
	return s.inner.ListRevisions(ctx, tenantID, capabilityID)
}
func (s *f7ConsistencyCapSVC) GetProposal(ctx context.Context, tenantID, proposalID string) (*capentity.CapabilityProposal, error) {
	return s.inner.GetProposal(ctx, tenantID, proposalID)
}
func (s *f7ConsistencyCapSVC) ListValidations(ctx context.Context, tenantID, revisionID string) ([]*capentity.CapabilityValidationResult, error) {
	return s.inner.ListValidations(ctx, tenantID, revisionID)
}
func (s *f7ConsistencyCapSVC) ListDecisions(ctx context.Context, tenantID, capabilityID string) ([]*capentity.CapabilityDecision, error) {
	return s.inner.ListDecisions(ctx, tenantID, capabilityID)
}
func (s *f7ConsistencyCapSVC) Validate(ctx context.Context, tenantID, revisionID, actorID string) (*capentity.BusinessCapabilityRevision, *capentity.CapabilityValidationResult, error) {
	return s.inner.Validate(ctx, tenantID, revisionID, actorID)
}
func (s *f7ConsistencyCapSVC) Activate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*capentity.BusinessCapabilityRevision, error) {
	return s.inner.Activate(ctx, tenantID, revisionID, actorID, reason)
}
func (s *f7ConsistencyCapSVC) MarkStale(ctx context.Context, tenantID, revisionID, actorID, reason string) (*capentity.BusinessCapabilityRevision, error) {
	return s.inner.MarkStale(ctx, tenantID, revisionID, actorID, reason)
}
func (s *f7ConsistencyCapSVC) Deprecate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*capentity.BusinessCapabilityRevision, error) {
	return s.inner.Deprecate(ctx, tenantID, revisionID, actorID, reason)
}

var _ capsvc.CapabilityService = (*f7ConsistencyCapSVC)(nil)

func seedF7Roles(t *testing.T, h *capabilityAppHarness, ownerSession context.Context, boot *tenancysvc.BootstrapResult, baseUID int64) (
	ownerCtx, adminCtx, memberCtx, viewerCtx context.Context, member *tenantentity.Principal,
) {
	t.Helper()
	admin, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, baseUID+1, "f7-admin@example.com")
	require.NoError(t, err)
	member, err = h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, baseUID+2, "f7-member@example.com")
	require.NoError(t, err)
	viewer, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, baseUID+3, "f7-viewer@example.com")
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
	ownerCtx = ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, baseUID)
	adminCtx = ctxCapability(ownerSession, tenantID, admin.PrincipalID, tenantentity.RoleAdmin, baseUID+1)
	memberCtx = ctxCapability(ownerSession, tenantID, member.PrincipalID, tenantentity.RoleMember, baseUID+2)
	viewerCtx = ctxCapability(ownerSession, tenantID, viewer.PrincipalID, tenantentity.RoleViewer, baseUID+3)
	return ownerCtx, adminCtx, memberCtx, viewerCtx, member
}

func TestF7_App_ProposalReads_RoleMatrixAndIsolation(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8710, "f7-prop-owner@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8710, "f7-prop-owner@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx, adminCtx, memberCtx, viewerCtx, member := seedF7Roles(t, h, ownerSession, boot, 8710)

	analysis, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "f7-prop-an",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.NotEmpty(t, analysis.Proposals)
	propID := analysis.Proposals[0].ProposalID
	runID := analysis.AnalysisRun.AnalysisRunID
	require.Equal(t, string(capentity.ProposalProposed), analysis.Proposals[0].Status)

	for _, ctx := range []context.Context{ownerCtx, adminCtx, memberCtx, viewerCtx} {
		got, err := h.app.GetCapabilityProposal(ctx, "lab", propID)
		require.NoError(t, err)
		require.Equal(t, propID, got.ProposalID)
		require.Equal(t, runID, got.AnalysisRunID)
		require.Equal(t, "lab", got.BusinessID)
		assertDTOJSONSecretFree(t, got)

		listed, err := h.app.ListCapabilityProposalsByAnalysis(ctx, "lab", runID)
		require.NoError(t, err)
		require.NotEmpty(t, listed)
		require.Equal(t, propID, listed[0].ProposalID)
		assertDTOJSONSecretFree(t, listed)
	}

	require.NoError(t, h.app.TenancySVC.RemoveMember(ownerSession, tenantID, member.PrincipalID, boot.Principal.PrincipalID, "rm-f7"))
	inactiveCtx := ctxCapability(ownerSession, tenantID, member.PrincipalID, tenantentity.RoleMember, 8712)
	_, err = h.app.GetCapabilityProposal(inactiveCtx, "lab", propID)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeTenantForbidden, fe.Code)
	_, err = h.app.ListCapabilityProposalsByAnalysis(inactiveCtx, "lab", runID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeTenantForbidden, fe.Code)

	sa, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 8799, "f7-super@example.com")
	require.NoError(t, err)
	saCtx := ctxCapability(ownerSession, tenantID, sa.PrincipalID, tenantentity.RoleOwner, 8799)
	_, err = h.app.GetCapabilityProposal(saCtx, "lab", propID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeTenantForbidden, fe.Code)
	_, err = h.app.ListCapabilityProposalsByAnalysis(saCtx, "lab", runID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeTenantForbidden, fe.Code)

	foreignCtx := ctxCapability(ownerSession, "other-tenant", boot.Principal.PrincipalID, tenantentity.RoleOwner, 8710)
	_, err = h.app.GetCapabilityProposal(foreignCtx, "lab", propID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.True(t, fe.Code == formaerrors.CodeTenantForbidden || fe.Code == formaerrors.CodeCapabilityProposalNotFound)

	h.seedBiz(tenantID, "other-lab")
	_, err = h.app.GetCapabilityProposal(ownerCtx, "other-lab", propID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityProposalNotFound, fe.Code)
	_, err = h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "other-lab", runID)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.True(t, fe.Code == formaerrors.CodeCapabilityAnalysisNotFound || fe.Code == formaerrors.CodeCapabilityProposalNotFound)
}

func TestF7_App_ProposalReads_TerminalAndListScopeStable(t *testing.T) {
	h := newCapabilityAppHarness()
	ownerSession := withSession(8720, "f7-term@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8720, "f7-term@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8720)

	an1, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "f7-term-1",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := an1.Proposals[0].ProposalID
	runID := an1.AnalysisRun.AnalysisRunID

	an2, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "f7-term-2",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.NotEqual(t, runID, an2.AnalysisRun.AnalysisRunID)

	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	for _, p := range listed {
		require.Equal(t, runID, p.AnalysisRunID)
		require.Equal(t, "lab", p.BusinessID)
	}
	listed2, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	require.Equal(t, len(listed), len(listed2))
	for i := range listed {
		require.Equal(t, listed[i].ProposalID, listed2[i].ProposalID)
		require.Equal(t, listed[i].Status, listed2[i].Status)
	}

	before, err := h.app.GetCapabilityAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	_, err = h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	after, err := h.app.GetCapabilityAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	require.Equal(t, before.Status, after.Status)
	require.Equal(t, before.Attempt, after.Attempt)

	got, err := h.app.GetCapabilityProposal(ownerCtx, "lab", propID)
	require.NoError(t, err)
	require.Equal(t, string(capentity.ProposalProposed), got.Status)

	_, err = h.app.ConfirmCapabilityProposal(ownerCtx, "lab", propID, &formaapp.ConfirmCapabilityProposalInput{
		ClientRequestID: "f7-confirm", CapabilityID: "cap-f7-confirm",
	})
	require.NoError(t, err)
	got, err = h.app.GetCapabilityProposal(ownerCtx, "lab", propID)
	require.NoError(t, err)
	require.Equal(t, string(capentity.ProposalConfirmed), got.Status)
	assertDTOJSONSecretFree(t, got)

	anR, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "f7-term-rej",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rejID := anR.Proposals[0].ProposalID
	_, err = h.app.RejectCapabilityProposal(ownerCtx, "lab", rejID, &formaapp.RejectCapabilityProposalInput{
		ClientRequestID: "f7-rej",
	})
	require.NoError(t, err)
	got, err = h.app.GetCapabilityProposal(ownerCtx, "lab", rejID)
	require.NoError(t, err)
	require.Equal(t, string(capentity.ProposalRejected), got.Status)

	listedR, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", anR.AnalysisRun.AnalysisRunID)
	require.NoError(t, err)
	require.Equal(t, string(capentity.ProposalRejected), listedR[0].Status)
}
