/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

// --- Malicious wrappers: embed CapabilityService; override only poisoned methods. ---

type f8PoisonGetProposalSVC struct {
	capsvc.CapabilityService
	mutate func(*capentity.CapabilityProposal) *capentity.CapabilityProposal
}

func (s *f8PoisonGetProposalSVC) GetProposal(ctx context.Context, tenantID, proposalID string) (*capentity.CapabilityProposal, error) {
	prop, err := s.CapabilityService.GetProposal(ctx, tenantID, proposalID)
	if err != nil || prop == nil || s.mutate == nil {
		return prop, err
	}
	cp := *prop
	return s.mutate(&cp), nil
}

type f8PoisonGetAnalysisSVC struct {
	capsvc.CapabilityService
	mutate func(*capentity.CapabilityAnalysisRun) *capentity.CapabilityAnalysisRun
}

func (s *f8PoisonGetAnalysisSVC) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*capentity.CapabilityAnalysisRun, error) {
	run, err := s.CapabilityService.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil || run == nil || s.mutate == nil {
		return run, err
	}
	cp := *run
	return s.mutate(&cp), nil
}

type f8PoisonListProposalsSVC struct {
	capsvc.CapabilityService
	mutate func([]*capentity.CapabilityProposal) []*capentity.CapabilityProposal
}

func (s *f8PoisonListProposalsSVC) ListProposalsByAnalysisRun(ctx context.Context, tenantID, analysisRunID string) ([]*capentity.CapabilityProposal, error) {
	props, err := s.CapabilityService.ListProposalsByAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil || s.mutate == nil {
		return props, err
	}
	return s.mutate(props), nil
}

var (
	_ capsvc.CapabilityService = (*f8PoisonGetProposalSVC)(nil)
	_ capsvc.CapabilityService = (*f8PoisonGetAnalysisSVC)(nil)
	_ capsvc.CapabilityService = (*f8PoisonListProposalsSVC)(nil)
)

func f8BootstrapOwner(t *testing.T, uid int64, email string) (*capabilityAppHarness, context.Context, string) {
	t.Helper()
	h := newCapabilityAppHarness()
	ownerSession := withSession(uid, email)
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, uid, email, 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, uid)
	return h, ownerCtx, tenantID
}

func f8StartAnalysis(t *testing.T, h *capabilityAppHarness, ownerCtx context.Context, clientReqID string) (propID, runID string) {
	t.Helper()
	analysis, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: clientReqID,
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, analysis.AnalysisRun)
	require.NotEmpty(t, analysis.Proposals)
	return analysis.Proposals[0].ProposalID, analysis.AnalysisRun.AnalysisRunID
}

func assertCapabilityProposalNotFound(t *testing.T, err error) {
	t.Helper()
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok, "want FormaError, got %v", err)
	require.Equal(t, formaerrors.CodeCapabilityProposalNotFound, fe.Code)
}

func assertCapabilityAnalysisNotFound(t *testing.T, err error) {
	t.Helper()
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok, "want FormaError, got %v", err)
	require.Equal(t, formaerrors.CodeCapabilityAnalysisNotFound, fe.Code)
}

func assertCapabilityConflict(t *testing.T, err error) {
	t.Helper()
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok, "want FormaError, got %v", err)
	require.Equal(t, formaerrors.CodeCapabilityConflict, fe.Code)
}

// --- GetCapabilityProposal identity (fail closed → CapabilityProposalNotFound) ---

func TestF8_GetCapabilityProposal_Identity_WrongTenantID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8801, "f8-prop-tenant@example.com")
	propID, _ := f8StartAnalysis(t, h, ownerCtx, "f8-prop-tenant")

	h.app.CapabilitySVC = &f8PoisonGetProposalSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(p *capentity.CapabilityProposal) *capentity.CapabilityProposal {
			p.TenantID = "foreign-tenant"
			return p
		},
	}
	dto, err := h.app.GetCapabilityProposal(ownerCtx, "lab", propID)
	require.Nil(t, dto, "must not leak foreign entity as DTO")
	assertCapabilityProposalNotFound(t, err)
}

func TestF8_GetCapabilityProposal_Identity_WrongProposalID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8802, "f8-prop-id@example.com")
	propID, _ := f8StartAnalysis(t, h, ownerCtx, "f8-prop-id")

	h.app.CapabilitySVC = &f8PoisonGetProposalSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(p *capentity.CapabilityProposal) *capentity.CapabilityProposal {
			p.ProposalID = "cprop_foreign_leaked"
			return p
		},
	}
	dto, err := h.app.GetCapabilityProposal(ownerCtx, "lab", propID)
	require.Nil(t, dto, "must not leak foreign proposal_id")
	assertCapabilityProposalNotFound(t, err)
}

func TestF8_GetCapabilityProposal_Identity_WrongBusinessID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8803, "f8-prop-biz@example.com")
	propID, _ := f8StartAnalysis(t, h, ownerCtx, "f8-prop-biz")

	h.app.CapabilitySVC = &f8PoisonGetProposalSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(p *capentity.CapabilityProposal) *capentity.CapabilityProposal {
			p.BusinessID = "other-lab"
			return p
		},
	}
	dto, err := h.app.GetCapabilityProposal(ownerCtx, "lab", propID)
	require.Nil(t, dto, "must not leak foreign business entity")
	assertCapabilityProposalNotFound(t, err)
}

// --- ListCapabilityProposalsByAnalysis — analysis lookup identity ---

func TestF8_ListByAnalysis_AnalysisIdentity_WrongTenantID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8811, "f8-an-tenant@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-an-tenant")

	h.app.CapabilitySVC = &f8PoisonGetAnalysisSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(r *capentity.CapabilityAnalysisRun) *capentity.CapabilityAnalysisRun {
			r.TenantID = "foreign-tenant"
			return r
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed)
	assertCapabilityAnalysisNotFound(t, err)
}

func TestF8_ListByAnalysis_AnalysisIdentity_WrongAnalysisRunID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8812, "f8-an-id@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-an-id")

	h.app.CapabilitySVC = &f8PoisonGetAnalysisSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(r *capentity.CapabilityAnalysisRun) *capentity.CapabilityAnalysisRun {
			r.AnalysisRunID = "crun_foreign_leaked"
			return r
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed)
	assertCapabilityAnalysisNotFound(t, err)
}

func TestF8_ListByAnalysis_AnalysisIdentity_WrongBusinessID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8813, "f8-an-biz@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-an-biz")

	h.app.CapabilitySVC = &f8PoisonGetAnalysisSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(r *capentity.CapabilityAnalysisRun) *capentity.CapabilityAnalysisRun {
			r.BusinessID = "other-lab"
			return r
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed)
	assertCapabilityAnalysisNotFound(t, err)
}

// --- ListCapabilityProposalsByAnalysis — list totality (ErrConsistency → CapabilityConflict) ---

func TestF8_ListByAnalysis_Totality_NilEntry(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8821, "f8-list-nil@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-list-nil")

	h.app.CapabilitySVC = &f8PoisonListProposalsSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(props []*capentity.CapabilityProposal) []*capentity.CapabilityProposal {
			out := make([]*capentity.CapabilityProposal, 0, len(props)+1)
			out = append(out, nil)
			out = append(out, props...)
			return out
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed, "nil entry must not yield partial list")
	assertCapabilityConflict(t, err)
}

func TestF8_ListByAnalysis_Totality_TenantMismatch(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8822, "f8-list-tenant@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-list-tenant")

	h.app.CapabilitySVC = &f8PoisonListProposalsSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(props []*capentity.CapabilityProposal) []*capentity.CapabilityProposal {
			require.NotEmpty(t, props)
			cp := *props[0]
			cp.TenantID = "foreign-tenant"
			return []*capentity.CapabilityProposal{&cp}
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed)
	assertCapabilityConflict(t, err)
}

func TestF8_ListByAnalysis_Totality_BusinessMismatch(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8823, "f8-list-biz@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-list-biz")

	h.app.CapabilitySVC = &f8PoisonListProposalsSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(props []*capentity.CapabilityProposal) []*capentity.CapabilityProposal {
			require.NotEmpty(t, props)
			cp := *props[0]
			cp.BusinessID = "other-lab"
			return []*capentity.CapabilityProposal{&cp}
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed)
	assertCapabilityConflict(t, err)
}

func TestF8_ListByAnalysis_Totality_AnalysisRunIDMismatch(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8824, "f8-list-run@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-list-run")

	h.app.CapabilitySVC = &f8PoisonListProposalsSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(props []*capentity.CapabilityProposal) []*capentity.CapabilityProposal {
			require.NotEmpty(t, props)
			cp := *props[0]
			cp.AnalysisRunID = "crun_other"
			return []*capentity.CapabilityProposal{&cp}
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed)
	assertCapabilityConflict(t, err)
}

func TestF8_ListByAnalysis_Totality_EmptyProposalID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8825, "f8-list-blank@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-list-blank")

	h.app.CapabilitySVC = &f8PoisonListProposalsSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(props []*capentity.CapabilityProposal) []*capentity.CapabilityProposal {
			require.NotEmpty(t, props)
			cp := *props[0]
			cp.ProposalID = ""
			return []*capentity.CapabilityProposal{&cp}
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed, "blank proposal_id must not yield partial/success list")
	assertCapabilityConflict(t, err)
}

func TestF8_ListByAnalysis_Totality_DuplicateProposalID(t *testing.T) {
	h, ownerCtx, _ := f8BootstrapOwner(t, 8826, "f8-list-dup@example.com")
	_, runID := f8StartAnalysis(t, h, ownerCtx, "f8-list-dup")

	h.app.CapabilitySVC = &f8PoisonListProposalsSVC{
		CapabilityService: h.app.CapabilitySVC,
		mutate: func(props []*capentity.CapabilityProposal) []*capentity.CapabilityProposal {
			require.NotEmpty(t, props)
			a := *props[0]
			b := *props[0]
			return []*capentity.CapabilityProposal{&a, &b}
		},
	}
	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.Nil(t, listed, "duplicate proposal_id must not yield partial list")
	assertCapabilityConflict(t, err)
}

// --- Legal happy path (still assert) ---

func TestF8_ListByAnalysis_HappyPath_StableSortTerminalNoMutation(t *testing.T) {
	h := newCapabilityAppHarness()
	// Two payloads → two proposal_ids so sort order is observable.
	h.app.CapabilitySVC = capsvc.NewCapabilityService(&capsvc.Components{
		UoW: h.uow,
		Generator: &capsvc.DeterministicFakeGenerator{
			Proposals: []capentity.SemanticPayload{
				fixture.LaboratoryCommandCapability(),
				fixture.LaboratoryFlowCapability(),
			},
		},
		Business: h.bm,
		Contract: h.contract,
	})

	ownerSession := withSession(8830, "f8-happy@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 8830, "f8-happy@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 8830)

	cmd := fixture.LaboratoryCommandCapability()
	h.seedPorts(tenantID, "lab", 1, cmd)
	flow := fixture.LaboratoryFlowCapability()
	h.seedPorts(tenantID, "lab", 1, flow)

	analysis, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "f8-happy",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(analysis.Proposals), 2)
	runID := analysis.AnalysisRun.AnalysisRunID

	listed, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listed), 2)
	for i := 1; i < len(listed); i++ {
		require.LessOrEqual(t, listed[i-1].ProposalID, listed[i].ProposalID, "stable sort by proposal_id")
	}
	for _, p := range listed {
		require.Equal(t, string(capentity.ProposalProposed), p.Status)
		require.Equal(t, runID, p.AnalysisRunID)
		require.Equal(t, "lab", p.BusinessID)
	}

	before, err := h.app.GetCapabilityAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	_, err = h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	after, err := h.app.GetCapabilityAnalysis(ownerCtx, "lab", runID)
	require.NoError(t, err)
	require.Equal(t, before.Status, after.Status)
	require.Equal(t, before.Attempt, after.Attempt)

	propID := listed[0].ProposalID
	_, err = h.app.ConfirmCapabilityProposal(ownerCtx, "lab", propID, &formaapp.ConfirmCapabilityProposalInput{
		ClientRequestID: "f8-confirm", CapabilityID: "cap-f8-confirm",
	})
	require.NoError(t, err)
	got, err := h.app.GetCapabilityProposal(ownerCtx, "lab", propID)
	require.NoError(t, err)
	require.Equal(t, string(capentity.ProposalConfirmed), got.Status)

	// Fresh analysis for REJECT terminal readability (confirm consumes one proposal).
	anR, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "f8-happy-rej",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rejID := anR.Proposals[0].ProposalID
	_, err = h.app.RejectCapabilityProposal(ownerCtx, "lab", rejID, &formaapp.RejectCapabilityProposalInput{
		ClientRequestID: "f8-rej",
	})
	require.NoError(t, err)
	got, err = h.app.GetCapabilityProposal(ownerCtx, "lab", rejID)
	require.NoError(t, err)
	require.Equal(t, string(capentity.ProposalRejected), got.Status)

	listedR, err := h.app.ListCapabilityProposalsByAnalysis(ownerCtx, "lab", anR.AnalysisRun.AnalysisRunID)
	require.NoError(t, err)
	foundRejected := false
	for _, p := range listedR {
		if p.ProposalID == rejID {
			require.Equal(t, string(capentity.ProposalRejected), p.Status)
			foundRejected = true
		}
	}
	require.True(t, foundRejected)
}
