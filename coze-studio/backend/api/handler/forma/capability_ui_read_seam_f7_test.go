/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"

	formaRouter "github.com/coze-dev/coze-studio/backend/api/router/forma"
	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	userentity "github.com/coze-dev/coze-studio/backend/domain/user/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

const (
	f7ProposalRoute = "GET /api/forma/v1/businesses/:id/capability-proposals/:proposalId"
	f7AnalysisProps = "GET /api/forma/v1/businesses/:id/capability-analyses/:analysisRunId/proposals"
)

func TestF7_Handler_ProposalReadRoutesRegistered(t *testing.T) {
	formaapp.ApplicationSVC = &formaapp.ApplicationService{}
	h := server.Default()
	formaRouter.Register(h)

	cases := []struct {
		method string
		path   string
	}{
		{"GET", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1"},
		{"GET", "/api/forma/v1/businesses/biz_test/capability-analyses/run_1/proposals"},
	}
	for _, tc := range cases {
		w := ut.PerformRequest(h.Engine, tc.method, tc.path, nil)
		if w.Code == http.StatusNotFound {
			t.Fatalf("F7 route not registered: %s %s — %s (expected %s and %s)",
				tc.method, tc.path, w.Body.String(), f7ProposalRoute, f7AnalysisProps)
		}
	}
}

type f7ProposalReadSVC struct {
	countingCapabilitySVC
	proposal *capentity.CapabilityProposal
	list     []*capentity.CapabilityProposal
	gotProp  string
	gotRun   string
}

func (s *f7ProposalReadSVC) GetProposal(_ context.Context, _, proposalID string) (*capentity.CapabilityProposal, error) {
	s.gotProp = proposalID
	if s.proposal != nil && s.proposal.ProposalID == proposalID {
		cp := *s.proposal
		return &cp, nil
	}
	return nil, capentity.ErrProposalNotFound
}

func (s *f7ProposalReadSVC) ListProposalsByAnalysisRun(_ context.Context, _, analysisRunID string) ([]*capentity.CapabilityProposal, error) {
	s.gotRun = analysisRunID
	out := make([]*capentity.CapabilityProposal, 0, len(s.list))
	for _, p := range s.list {
		if p.AnalysisRunID == analysisRunID {
			cp := *p
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *f7ProposalReadSVC) GetAnalysisRun(_ context.Context, tenantID, analysisRunID string) (*capentity.CapabilityAnalysisRun, error) {
	s.gotRun = analysisRunID
	for _, p := range s.list {
		if p.AnalysisRunID == analysisRunID {
			return &capentity.CapabilityAnalysisRun{
				AnalysisRunID: analysisRunID, TenantID: tenantID, BusinessID: p.BusinessID,
				Status: capentity.AnalysisSucceeded, Attempt: 1,
			}, nil
		}
	}
	if analysisRunID == "run_exact" {
		return &capentity.CapabilityAnalysisRun{
			AnalysisRunID: analysisRunID, TenantID: tenantID, BusinessID: "biz_exact",
			Status: capentity.AnalysisSucceeded, Attempt: 1,
		}, nil
	}
	return nil, capentity.ErrAnalysisNotFound
}

var _ capsvc.CapabilityService = (*f7ProposalReadSVC)(nil)

func newF7HandlerEnv(t *testing.T, cap capsvc.CapabilityService) (*server.Hertz, string) {
	t.Helper()
	const (
		userID      int64 = 9700
		tenantID          = "ten_f7_handler"
		principalID       = "prin_f7_handler"
	)
	tenancy := &capabilityHandlerTenancyStub{
		principal: &tenantentity.Principal{
			PrincipalID: principalID, CozeUserID: userID, Status: tenantentity.PrincipalStatusActive,
		},
		tenant: &tenantentity.Tenant{
			TenantID: tenantID, Status: tenantentity.TenantStatusActive, OwnerPrincipalID: principalID,
		},
		member: &tenantentity.Membership{
			TenantID: tenantID, PrincipalID: principalID, Role: tenantentity.RoleOwner, Status: tenantentity.MembershipActive,
		},
	}
	formaapp.ApplicationSVC = &formaapp.ApplicationService{
		TenancySVC:    tenancy,
		CapabilitySVC: cap,
	}
	h := server.Default()
	h.Use(func(c context.Context, ctx *app.RequestContext) {
		c = ctxcache.Init(c)
		ctxcache.Store(c, consts.SessionDataKeyInCtx, &userentity.Session{
			UserID: userID, UserEmail: "f7-handler@example.com",
		})
		ctx.Next(c)
	})
	formaRouter.Register(h)
	return h, tenantID
}

func performF7GET(t *testing.T, h *server.Hertz, path, tenantID, requestID string) (int, map[string]any) {
	t.Helper()
	headers := []ut.Header{
		{Key: "X-Forma-Tenant", Value: tenantID},
		{Key: "X-Request-ID", Value: requestID},
	}
	w := ut.PerformRequest(h.Engine, http.MethodGet, path, nil, headers...)
	var env map[string]any
	body := w.Body.Bytes()
	if len(body) > 0 && body[0] == '{' {
		if err := json.Unmarshal(body, &env); err != nil {
			t.Fatalf("unmarshal envelope: %v body=%s", err, w.Body.String())
		}
	} else if len(body) > 0 {
		env = map[string]any{"raw": w.Body.String()}
	}
	return w.Code, env
}

func f7FixtureProposal(id, runID, biz string, status capentity.ProposalStatus) *capentity.CapabilityProposal {
	return &capentity.CapabilityProposal{
		ProposalID: id, TenantID: "ten_f7_handler", BusinessID: biz,
		AnalysisRunID: runID, Status: status,
		Payload: capentity.SemanticPayload{
			Name: "F7ReadOnly", Description: "logical only", CapabilityKind: capentity.KindCommand,
			BusinessModelRevision: 1,
		},
		CreatedAt: time.Unix(1_700_000_000, 0).UTC(),
	}
}

func TestF7_Handler_GetProposal_EnvelopeAndDTOShape(t *testing.T) {
	prop := f7FixtureProposal("prop_f7_1", "run_f7_1", "biz_test", capentity.ProposalProposed)
	cap := &f7ProposalReadSVC{proposal: prop, list: []*capentity.CapabilityProposal{prop}}
	h, tenantID := newF7HandlerEnv(t, cap)

	rid := "req-f7-get-prop"
	status, env := performF7GET(t, h, "/api/forma/v1/businesses/biz_test/capability-proposals/prop_f7_1", tenantID, rid)
	if status == http.StatusNotFound {
		t.Fatalf("route missing (%s): env=%v", f7ProposalRoute, env)
	}
	if status != http.StatusOK {
		t.Fatalf("status=%d env=%v", status, env)
	}
	if env["request_id"] != rid {
		t.Fatalf("request_id=%v want %s", env["request_id"], rid)
	}
	data, ok := env["data"].(map[string]any)
	if !ok || data == nil {
		t.Fatalf("expected data object, got %T %v", env["data"], env["data"])
	}
	for _, key := range []string{"proposal_id", "business_id", "analysis_run_id", "status", "payload", "created_at"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("missing DTO field %q in %v", key, data)
		}
	}
	raw, _ := json.Marshal(env)
	body := strings.ToLower(string(raw))
	for _, forbidden := range []string{"password", "connection_id", "physical_locator", "snapshot_id", "credential", "secret_ref"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("envelope must not echo %q: %s", forbidden, body)
		}
	}
}

func TestF7_Handler_ListAnalysisProposals_EnvelopeAndDTOShape(t *testing.T) {
	prop := f7FixtureProposal("prop_f7_2", "run_f7_2", "biz_test", capentity.ProposalRejected)
	cap := &f7ProposalReadSVC{proposal: prop, list: []*capentity.CapabilityProposal{prop}}
	h, tenantID := newF7HandlerEnv(t, cap)

	rid := "req-f7-list-props"
	status, env := performF7GET(t, h,
		"/api/forma/v1/businesses/biz_test/capability-analyses/run_f7_2/proposals", tenantID, rid)
	if status == http.StatusNotFound {
		t.Fatalf("route missing (%s): env=%v", f7AnalysisProps, env)
	}
	if status != http.StatusOK {
		t.Fatalf("status=%d env=%v", status, env)
	}
	if env["request_id"] != rid {
		t.Fatalf("request_id=%v want %s", env["request_id"], rid)
	}
	data, ok := env["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T %v", env["data"], env["data"])
	}
	if len(data) != 1 {
		t.Fatalf("len(data)=%d want 1", len(data))
	}
	row, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("row type %T", data[0])
	}
	if row["proposal_id"] != "prop_f7_2" {
		t.Fatalf("proposal_id=%v", row["proposal_id"])
	}
	if row["status"] != string(capentity.ProposalRejected) {
		t.Fatalf("status=%v", row["status"])
	}
}

func TestF7_Handler_MalformedEmptyPath_FailClosedNoSecretEcho(t *testing.T) {
	cap := &f7ProposalReadSVC{}
	h, tenantID := newF7HandlerEnv(t, cap)

	cases := []struct {
		name string
		path string
	}{
		{"empty-proposal-trailing", "/api/forma/v1/businesses/biz_test/capability-proposals/"},
		{"empty-analysis-segment", "/api/forma/v1/businesses/biz_test/capability-analyses//proposals"},
		{"missing-proposal", "/api/forma/v1/businesses/biz_test/capability-proposals/missing_prop"},
		{"missing-analysis-list", "/api/forma/v1/businesses/biz_test/capability-analyses/missing_run/proposals"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rid := "req-f7-fc-" + tc.name
			status, env := performF7GET(t, h, tc.path, tenantID, rid)
			if status == http.StatusOK {
				t.Fatalf("expected fail-closed, got OK env=%v", env)
			}
			raw := ""
			if b, err := json.Marshal(env); err == nil {
				raw = strings.ToLower(string(b))
			}
			for _, forbidden := range []string{"password=hunter2", "connection_id", "secret_ref", "physical_locator"} {
				if strings.Contains(raw, forbidden) {
					t.Fatalf("must not echo %q: %s", forbidden, raw)
				}
			}
			if msg, _ := env["msg"].(string); strings.Contains(strings.ToLower(msg), "password") {
				t.Fatalf("msg must not contain password: %s", msg)
			}
			_ = formaerrors.CodeOK
		})
	}
}

func TestF7_Handler_ExactRouteParamsReachDomain(t *testing.T) {
	cap := &f7ProposalReadSVC{
		proposal: f7FixtureProposal("prop_exact", "run_exact", "biz_exact", capentity.ProposalProposed),
		list:     []*capentity.CapabilityProposal{f7FixtureProposal("prop_exact", "run_exact", "biz_exact", capentity.ProposalProposed)},
	}
	h, tenantID := newF7HandlerEnv(t, cap)

	status, _ := performF7GET(t, h, "/api/forma/v1/businesses/biz_exact/capability-proposals/prop_exact", tenantID, "rid-get")
	if status == http.StatusNotFound {
		t.Fatalf("Get proposal route not registered (%s)", f7ProposalRoute)
	}
	if cap.gotProp != "prop_exact" {
		t.Fatalf("expected proposalId param prop_exact, got %q (handler must bind :proposalId)", cap.gotProp)
	}

	cap.gotRun = ""
	status, _ = performF7GET(t, h, "/api/forma/v1/businesses/biz_exact/capability-analyses/run_exact/proposals", tenantID, "rid-list")
	if status == http.StatusNotFound {
		t.Fatalf("List analysis proposals route not registered (%s)", f7AnalysisProps)
	}
	if cap.gotRun != "run_exact" {
		t.Fatalf("expected analysisRunId param run_exact, got %q (handler must bind :analysisRunId)", cap.gotRun)
	}
}
