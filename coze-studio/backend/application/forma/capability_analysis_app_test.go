/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

type analysisAppHarness struct {
	app *formaapp.ApplicationService
	biz *capabilityBusinessStub
	gen *capsvc.DeterministicFakeGenerator
}

func newAnalysisAppHarness(gen *capsvc.DeterministicFakeGenerator) *analysisAppHarness {
	if gen == nil {
		gen = &capsvc.DeterministicFakeGenerator{Proposals: []capentity.SemanticPayload{fixture.LaboratoryCommandCapability()}}
	}
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
			Generator: gen,
			Business:  bm,
			Contract:  contracts,
		}),
	}
	return &analysisAppHarness{app: app, biz: biz, gen: gen}
}

func (h *analysisAppHarness) seedBiz(tenantID, businessID string) {
	h.biz.put(tenantID, businessID)
}

func TestCapabilityAnalysisApp_FailedFirstAttemptVisible(t *testing.T) {
	gen := &capsvc.DeterministicFakeGenerator{Err: errors.New("boom secret password=xyz")}
	h := newAnalysisAppHarness(gen)
	ownerSession := withSession(9100, "cap-fail@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 9100, "cap-fail@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 9100)

	resp, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "fail-first",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err, "FAILED first attempt must surface as usable DTO (HTTP 200 path)")
	require.NotNil(t, resp)
	require.NotNil(t, resp.AnalysisRun)
	require.Equal(t, string(capentity.AnalysisFailed), resp.AnalysisRun.Status)
	require.Equal(t, "FORMA_CAPABILITY_ANALYSIS_FAILED", resp.AnalysisRun.ErrorCode)
	require.NotEmpty(t, resp.AnalysisRun.AnalysisRunID)
	require.Equal(t, int32(1), resp.AnalysisRun.Attempt)
	require.Equal(t, boot.Principal.PrincipalID, resp.AnalysisRun.CreatedBy)

	got, err := h.app.GetCapabilityAnalysis(ownerCtx, "lab", resp.AnalysisRun.AnalysisRunID)
	require.NoError(t, err)
	require.Equal(t, resp.AnalysisRun.AnalysisRunID, got.AnalysisRunID)
	require.Equal(t, string(capentity.AnalysisFailed), got.Status)
}

func TestCapabilityAnalysisApp_ExplicitRetryIncrementsAttempt(t *testing.T) {
	gen := &capsvc.DeterministicFakeGenerator{Err: errors.New("boom")}
	h := newAnalysisAppHarness(gen)
	ownerSession := withSession(9110, "cap-retry@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 9110, "cap-retry@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 9110)

	failed, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "retry-me",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Equal(t, string(capentity.AnalysisFailed), failed.AnalysisRun.Status)
	runID := failed.AnalysisRun.AnalysisRunID

	gen.Err = nil
	gen.Proposals = []capentity.SemanticPayload{fixture.LaboratoryCommandCapability()}
	retried, err := h.app.RetryCapabilityAnalysis(ownerCtx, "lab", runID, &formaapp.RetryCapabilityAnalysisInput{
		Reason: "operator requested retry",
	})
	require.NoError(t, err)
	require.Equal(t, runID, retried.AnalysisRun.AnalysisRunID)
	require.Equal(t, string(capentity.AnalysisSucceeded), retried.AnalysisRun.Status)
	require.GreaterOrEqual(t, retried.AnalysisRun.Attempt, int32(2))
	require.NotEmpty(t, retried.Proposals)
}

func TestCapabilityAnalysisApp_ConcurrentRetryAtMostOneSuccess(t *testing.T) {
	gen := &capsvc.DeterministicFakeGenerator{Err: errors.New("boom")}
	h := newAnalysisAppHarness(gen)
	ownerSession := withSession(9120, "cap-conc@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 9120, "cap-conc@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 9120)

	failed, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "conc-retry",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	runID := failed.AnalysisRun.AnalysisRunID

	gen.Err = nil
	gen.GenerateFn = func(context.Context, capsvc.GenerateRequest) (*capsvc.GenerateResult, error) {
		time.Sleep(40 * time.Millisecond)
		return &capsvc.GenerateResult{
			ModelRef:  "fake-model",
			Proposals: []capentity.SemanticPayload{fixture.LaboratoryCommandCapability()},
		}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := h.app.RetryCapabilityAnalysis(ownerCtx, "lab", runID, nil)
			if err == nil && res != nil && res.AnalysisRun != nil && res.AnalysisRun.Status == string(capentity.AnalysisSucceeded) {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	require.Equal(t, 1, successes, "ClaimAnalysisRetry must admit at most one executor success path")
}

func TestCapabilityAnalysisApp_ReplayFailedDoesNotAutoRetry(t *testing.T) {
	gen := &capsvc.DeterministicFakeGenerator{Err: errors.New("boom")}
	h := newAnalysisAppHarness(gen)
	ownerSession := withSession(9130, "cap-replay@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 9130, "cap-replay@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 9130)

	first, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "replay-fail",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Equal(t, string(capentity.AnalysisFailed), first.AnalysisRun.Status)
	callsAfterFail := gen.CallCount()

	gen.Err = nil
	gen.Proposals = []capentity.SemanticPayload{fixture.LaboratoryCommandCapability()}
	replay, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "replay-fail",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Equal(t, first.AnalysisRun.AnalysisRunID, replay.AnalysisRun.AnalysisRunID)
	require.Equal(t, string(capentity.AnalysisFailed), replay.AnalysisRun.Status)
	require.Equal(t, first.AnalysisRun.Attempt, replay.AnalysisRun.Attempt)
	require.False(t, replay.OwnedExecute)
	require.Equal(t, callsAfterFail, gen.CallCount(), "identical client_request_id must not auto-retry FAILED run")
}

func TestMapDomainError_AnalysisFailedNotValidationFailed(t *testing.T) {
	fe := formaerrors.MapDomainError(capentity.ErrAnalysisFailed)
	require.Equal(t, formaerrors.CodeCapabilityAnalysisFailed, fe.Code)
	require.Equal(t, formaerrors.KeyCapabilityAnalysisFailed, fe.Key)
	require.NotEqual(t, formaerrors.KeyCapabilityValidationFailed, fe.Key)

	feVal := formaerrors.MapDomainError(capentity.ErrValidationFailed)
	require.Equal(t, formaerrors.KeyCapabilityValidationFailed, feVal.Key)
}

func TestCapabilityAnalysisApp_RetryAuthAndIsolation(t *testing.T) {
	gen := &capsvc.DeterministicFakeGenerator{Err: errors.New("boom")}
	h := newAnalysisAppHarness(gen)
	ownerSession := withSession(9140, "cap-auth@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 9140, "cap-auth@example.com", 0)
	require.NoError(t, err)
	member, err := h.app.TenancySVC.ResolveOrCreatePrincipal(ownerSession, 9141, "cap-auth-member@example.com")
	require.NoError(t, err)
	_, err = h.app.TenancySVC.AddMember(ownerSession, &tenancysvc.AddMemberRequest{
		TenantID: boot.Tenant.TenantID, PrincipalID: member.PrincipalID, Role: tenantentity.RoleMember,
		CreatedBy: boot.Principal.PrincipalID,
	})
	require.NoError(t, err)

	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	h.seedBiz(tenantID, "other")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 9140)
	memberCtx := ctxCapability(ownerSession, tenantID, member.PrincipalID, tenantentity.RoleMember, 9141)

	failed, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "auth-retry",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	runID := failed.AnalysisRun.AnalysisRunID

	gen.Err = nil
	gen.Proposals = []capentity.SemanticPayload{fixture.LaboratoryCommandCapability()}

	_, err = h.app.RetryCapabilityAnalysis(memberCtx, "lab", runID, nil)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityForbidden, fe.Code)

	_, err = h.app.RetryCapabilityAnalysis(ownerCtx, "other", runID, nil)
	fe, ok = formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityAnalysisNotFound, fe.Code)
}

func TestCapabilityAnalysisApp_RetryNonFailedConflict(t *testing.T) {
	h := newAnalysisAppHarness(nil)
	ownerSession := withSession(9150, "cap-notfail@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 9150, "cap-notfail@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 9150)

	okRun, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "succeeded-run",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Equal(t, string(capentity.AnalysisSucceeded), okRun.AnalysisRun.Status)

	_, err = h.app.RetryCapabilityAnalysis(ownerCtx, "lab", okRun.AnalysisRun.AnalysisRunID, nil)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityIllegalTransition, fe.Code)
}

func TestCapabilityAnalysisApp_RetryReasonSecretRejected(t *testing.T) {
	gen := &capsvc.DeterministicFakeGenerator{Err: errors.New("boom")}
	h := newAnalysisAppHarness(gen)
	ownerSession := withSession(9160, "cap-reason@example.com")
	boot, err := h.app.TenancySVC.Bootstrap(ownerSession, 9160, "cap-reason@example.com", 0)
	require.NoError(t, err)
	tenantID := boot.Tenant.TenantID
	h.seedBiz(tenantID, "lab")
	ownerCtx := ctxCapability(ownerSession, tenantID, boot.Principal.PrincipalID, tenantentity.RoleOwner, 9160)

	failed, err := h.app.StartCapabilityAnalysis(ownerCtx, "lab", &formaapp.StartCapabilityAnalysisInput{
		BusinessModelRevision: 1, ClientRequestID: "reason-reject-case",
		Analysis: capentity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)

	gen.Err = nil
	gen.Proposals = []capentity.SemanticPayload{fixture.LaboratoryCommandCapability()}
	_, err = h.app.RetryCapabilityAnalysis(ownerCtx, "lab", failed.AnalysisRun.AnalysisRunID, &formaapp.RetryCapabilityAnalysisInput{
		Reason: "contains api_key material",
	})
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	require.Equal(t, formaerrors.CodeCapabilityInvalidPayload, fe.Code)
	require.Equal(t, string(capentity.AnalysisFailed), failed.AnalysisRun.Status)
}
