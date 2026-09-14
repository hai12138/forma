/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"

	formaRouter "github.com/coze-dev/coze-studio/backend/api/router/forma"
	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	bizentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
	userentity "github.com/coze-dev/coze-studio/backend/domain/user/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

func TestCapabilityRoutesAreRegistered(t *testing.T) {
	formaapp.ApplicationSVC = &formaapp.ApplicationService{}
	h := server.Default()
	formaRouter.Register(h)

	cases := []struct {
		method string
		path   string
		body   []byte
	}{
		{"GET", "/api/forma/v1/businesses/biz_test/capabilities", nil},
		{"POST", "/api/forma/v1/businesses/biz_test/capabilities", []byte(`{"payload":{"name":"x","capability_kind":"COMMAND","business_model_revision":1}}`)},
		{"POST", "/api/forma/v1/businesses/biz_test/capabilities/analyze", []byte(`{"business_model_revision":1,"client_request_id":"r1","analysis":{"business_model_revision":1}}`)},
		{"GET", "/api/forma/v1/businesses/biz_test/capabilities/cap_1", nil},
		{"GET", "/api/forma/v1/businesses/biz_test/capabilities/cap_1/revisions", nil},
		{"GET", "/api/forma/v1/businesses/biz_test/capabilities/cap_1/revisions/rev_1", nil},
		{"POST", "/api/forma/v1/businesses/biz_test/capabilities/cap_1/derive", []byte(`{"source_revision_id":"rev_1","client_request_id":"d1","payload":{"name":"x","capability_kind":"COMMAND","business_model_revision":1}}`)},
		{"POST", "/api/forma/v1/businesses/biz_test/capabilities/cap_1/edit", []byte(`{"source_revision_id":"rev_1","client_request_id":"e1","payload":{"name":"x","capability_kind":"COMMAND","business_model_revision":1}}`)},
		{"GET", "/api/forma/v1/businesses/biz_test/capabilities/cap_1/decisions", nil},
		{"GET", "/api/forma/v1/businesses/biz_test/capability-analyses/run_1", nil},
		{"POST", "/api/forma/v1/businesses/biz_test/capability-analyses/run_1/retry", []byte(`{}`)},
		{"POST", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/confirm", []byte(`{}`)},
		{"POST", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/edit-confirm", []byte(`{"effective_payload":{"name":"x","capability_kind":"COMMAND","business_model_revision":1}}`)},
		{"POST", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/reject", []byte(`{}`)},
		{"POST", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/validate", nil},
		{"POST", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/activate", []byte(`{}`)},
		{"POST", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/deprecate", []byte(`{}`)},
		{"GET", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/validations", nil},
	}

	for _, tc := range cases {
		var body *ut.Body
		var headers []ut.Header
		if tc.body != nil {
			body = &ut.Body{Body: bytes.NewBuffer(tc.body), Len: len(tc.body)}
			headers = append(headers, ut.Header{Key: "Content-Type", Value: "application/json"})
		}
		w := ut.PerformRequest(h.Engine, tc.method, tc.path, body, headers...)
		if w.Code == 404 {
			t.Fatalf("capability route not registered: %s %s — %s", tc.method, tc.path, w.Body.String())
		}
	}
}

// --- Handler envelope / malformed JSON (through Register + FormaTenantMW) ---

type capabilityHandlerTenancyStub struct {
	principal *tenantentity.Principal
	tenant    *tenantentity.Tenant
	member    *tenantentity.Membership
}

func (s *capabilityHandlerTenancyStub) ResolveOrCreatePrincipal(context.Context, int64, string) (*tenantentity.Principal, error) {
	return s.principal, nil
}
func (s *capabilityHandlerTenancyStub) CreateTenant(context.Context, *tenancysvc.CreateTenantRequest) (*tenantentity.Tenant, error) {
	return nil, tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) GetTenant(_ context.Context, tenantID string) (*tenantentity.Tenant, error) {
	if s.tenant != nil && s.tenant.TenantID == tenantID {
		cp := *s.tenant
		return &cp, nil
	}
	return nil, tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) UpdateTenant(context.Context, *tenantentity.Tenant, int32, string, string) (*tenantentity.Tenant, error) {
	return nil, tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) ListTenantsForPrincipal(context.Context, string) ([]*tenantentity.Tenant, error) {
	if s.tenant == nil {
		return nil, nil
	}
	cp := *s.tenant
	return []*tenantentity.Tenant{&cp}, nil
}
func (s *capabilityHandlerTenancyStub) AddMember(context.Context, *tenancysvc.AddMemberRequest) (*tenantentity.Membership, error) {
	return nil, tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) UpdateMemberRole(context.Context, string, string, tenantentity.MembershipRole, int32, string, string) (*tenantentity.Membership, error) {
	return nil, tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) RemoveMember(context.Context, string, string, string, string) error {
	return tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) ListMembers(context.Context, string) ([]*tenantentity.Membership, error) {
	return nil, nil
}
func (s *capabilityHandlerTenancyStub) GetMembership(_ context.Context, tenantID, principalID string) (*tenantentity.Membership, error) {
	if s.member != nil && s.member.TenantID == tenantID && s.member.PrincipalID == principalID {
		cp := *s.member
		return &cp, nil
	}
	return nil, nil
}
func (s *capabilityHandlerTenancyStub) ListMembershipsForPrincipal(context.Context, string) ([]*tenantentity.Membership, error) {
	if s.member == nil {
		return nil, nil
	}
	cp := *s.member
	return []*tenantentity.Membership{&cp}, nil
}
func (s *capabilityHandlerTenancyStub) BindSpace(context.Context, *tenancysvc.BindSpaceRequest) (*tenantentity.TenantSpaceRef, error) {
	return nil, tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) UnbindSpace(context.Context, string, int64, string, string) error {
	return tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) ListSpaces(context.Context, string) ([]*tenantentity.TenantSpaceRef, error) {
	return nil, nil
}
func (s *capabilityHandlerTenancyStub) Bootstrap(context.Context, int64, string, int64) (*tenancysvc.BootstrapResult, error) {
	return nil, tenantentity.ErrNotFound
}
func (s *capabilityHandlerTenancyStub) RecordAudit(context.Context, *tenantentity.AuditEvent) error {
	return nil
}
func (s *capabilityHandlerTenancyStub) GetPlatformRole(context.Context, string) (*tenantentity.FormaPlatformRoleAssignment, error) {
	return nil, nil
}
func (s *capabilityHandlerTenancyStub) SetPlatformRole(context.Context, string, tenantentity.PlatformRole) error {
	return nil
}
func (s *capabilityHandlerTenancyStub) SetPasswordChangeRequired(context.Context, string, bool) error {
	return nil
}
func (s *capabilityHandlerTenancyStub) ListPlatformRoles(context.Context) ([]*tenantentity.FormaPlatformRoleAssignment, error) {
	return nil, nil
}
func (s *capabilityHandlerTenancyStub) CountActiveSuperAdmins(context.Context) (int, error) {
	return 0, nil
}
func (s *capabilityHandlerTenancyStub) SuspendPrincipal(context.Context, string) error {
	return nil
}
func (s *capabilityHandlerTenancyStub) ActivatePrincipal(context.Context, string) error {
	return nil
}
func (s *capabilityHandlerTenancyStub) GetPrincipalByID(context.Context, string) (*tenantentity.Principal, error) {
	return s.principal, nil
}
func (s *capabilityHandlerTenancyStub) ListAllPrincipals(context.Context) ([]*tenantentity.Principal, error) {
	if s.principal == nil {
		return nil, nil
	}
	cp := *s.principal
	return []*tenantentity.Principal{&cp}, nil
}

var _ tenancysvc.TenancyService = (*capabilityHandlerTenancyStub)(nil)

// countingCapabilitySVC counts domain mutation entry points to prove malformed JSON never writes.
type countingCapabilitySVC struct {
	mu     sync.Mutex
	writes int
}

func (s *countingCapabilitySVC) bump() {
	s.mu.Lock()
	s.writes++
	s.mu.Unlock()
}

func (s *countingCapabilitySVC) writeCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writes
}

func (s *countingCapabilitySVC) ManualCreate(context.Context, *capsvc.ManualCreateInput) (*capentity.BusinessCapability, *capentity.BusinessCapabilityRevision, error) {
	s.bump()
	return nil, nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) DeriveRevision(context.Context, *capsvc.DeriveInput) (*capentity.BusinessCapabilityRevision, *capentity.CapabilityDecision, error) {
	s.bump()
	return nil, nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) StartAnalysis(context.Context, *capsvc.StartAnalysisInput) (*capsvc.AnalysisResult, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) GetAnalysisRun(context.Context, string, string) (*capentity.CapabilityAnalysisRun, error) {
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) RetryFailedAnalysis(context.Context, string, string, string) (*capsvc.AnalysisResult, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) ConfirmProposal(context.Context, *capsvc.ConfirmInput) (*capentity.BusinessCapabilityRevision, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) EditConfirmProposal(context.Context, *capsvc.EditConfirmInput) (*capentity.BusinessCapabilityRevision, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) RejectProposal(context.Context, *capsvc.RejectInput) (*capentity.CapabilityDecision, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) GetCapability(context.Context, string, string) (*capentity.BusinessCapability, error) {
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) ListCapabilities(context.Context, string, string) ([]*capentity.BusinessCapability, error) {
	return nil, nil
}
func (s *countingCapabilitySVC) GetRevision(context.Context, string, string) (*capentity.BusinessCapabilityRevision, error) {
	return nil, capentity.ErrRevisionNotFound
}
func (s *countingCapabilitySVC) ListRevisions(context.Context, string, string) ([]*capentity.BusinessCapabilityRevision, error) {
	return nil, nil
}
func (s *countingCapabilitySVC) GetProposal(context.Context, string, string) (*capentity.CapabilityProposal, error) {
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) ListValidations(context.Context, string, string) ([]*capentity.CapabilityValidationResult, error) {
	return nil, nil
}
func (s *countingCapabilitySVC) ListDecisions(context.Context, string, string) ([]*capentity.CapabilityDecision, error) {
	return nil, nil
}
func (s *countingCapabilitySVC) Validate(context.Context, string, string, string) (*capentity.BusinessCapabilityRevision, *capentity.CapabilityValidationResult, error) {
	s.bump()
	return nil, nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) Activate(context.Context, string, string, string, string) (*capentity.BusinessCapabilityRevision, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) MarkStale(context.Context, string, string, string, string) (*capentity.BusinessCapabilityRevision, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}
func (s *countingCapabilitySVC) Deprecate(context.Context, string, string, string, string) (*capentity.BusinessCapabilityRevision, error) {
	s.bump()
	return nil, capentity.ErrNotFound
}

var _ capsvc.CapabilityService = (*countingCapabilitySVC)(nil)

type capabilityEnvelope struct {
	Code      int32  `json:"code"`
	Msg       string `json:"msg"`
	RequestID string `json:"request_id"`
	ErrorKey  string `json:"error_key"`
}

func newCapabilityHandlerEnv(t *testing.T) (*server.Hertz, *countingCapabilitySVC, string) {
	t.Helper()
	const (
		userID      int64 = 9100
		tenantID          = "ten_cap_handler"
		principalID       = "prin_cap_handler"
	)
	cap := &countingCapabilitySVC{}
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
			UserID:    userID,
			UserEmail: "cap-handler@example.com",
		})
		ctx.Next(c)
	})
	formaRouter.Register(h)
	return h, cap, tenantID
}

func performCapabilityJSON(t *testing.T, h *server.Hertz, method, path, tenantID, requestID string, raw []byte) (int, capabilityEnvelope) {
	t.Helper()
	headers := []ut.Header{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "X-Forma-Tenant", Value: tenantID},
		{Key: "X-Request-ID", Value: requestID},
	}
	var body *ut.Body
	if raw != nil {
		body = &ut.Body{Body: bytes.NewBuffer(raw), Len: len(raw)}
	}
	w := ut.PerformRequest(h.Engine, method, path, body, headers...)
	var env capabilityEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v body=%s", err, w.Body.String())
	}
	return w.Code, env
}

func TestCapabilityOptionalBody_MalformedJSONFailClosed(t *testing.T) {
	h, cap, tenantID := newCapabilityHandlerEnv(t)
	malformed := []byte(`{"reason":`)
	cases := []struct {
		name string
		path string
	}{
		{"confirm", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/confirm"},
		{"reject", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/reject"},
		{"activate", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/activate"},
		{"deprecate", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/deprecate"},
		{"edit-confirm", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/edit-confirm"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := cap.writeCount()
			rid := "req-malformed-" + tc.name
			status, env := performCapabilityJSON(t, h, http.MethodPost, tc.path, tenantID, rid, malformed)
			if status != http.StatusBadRequest {
				t.Fatalf("status=%d body=%+v", status, env)
			}
			if env.Code != formaerrors.CodeAdminBadRequest {
				t.Fatalf("code=%d want %d", env.Code, formaerrors.CodeAdminBadRequest)
			}
			if env.ErrorKey != formaerrors.KeyAdminBadRequest {
				t.Fatalf("error_key=%q want %q", env.ErrorKey, formaerrors.KeyAdminBadRequest)
			}
			if env.RequestID != rid {
				t.Fatalf("request_id=%q want %q", env.RequestID, rid)
			}
			if cap.writeCount() != before {
				t.Fatalf("expected no capability writes on malformed JSON; writes %d -> %d", before, cap.writeCount())
			}
		})
	}
}

func TestCapabilityOptionalBody_EmptyBodySkipsBind(t *testing.T) {
	h, _, tenantID := newCapabilityHandlerEnv(t)
	cases := []struct {
		name string
		path string
		body []byte
	}{
		{"confirm-nil", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/confirm", nil},
		{"confirm-empty", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/confirm", []byte{}},
		{"confirm-ws", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/confirm", []byte("   \n")},
		{"reject-empty", "/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/reject", []byte{}},
		{"activate-empty", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/activate", []byte{}},
		{"deprecate-empty", "/api/forma/v1/businesses/biz_test/capability-revisions/rev_1/deprecate", []byte{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rid := "req-empty-" + tc.name
			status, env := performCapabilityJSON(t, h, http.MethodPost, tc.path, tenantID, rid, tc.body)
			if env.ErrorKey == formaerrors.KeyAdminBadRequest {
				t.Fatalf("empty optional body must not map to BadRequest; status=%d env=%+v", status, env)
			}
			if env.RequestID != rid {
				t.Fatalf("request_id=%q want %q", env.RequestID, rid)
			}
			// Past bind: stub GetProposal/GetRevision → not found / mapped capability error.
			if status == http.StatusNotFound || status == http.StatusBadRequest || status == http.StatusConflict || status == http.StatusForbidden {
				// ok — application path reached
			} else if status == http.StatusOK {
				t.Fatalf("unexpected OK with stub capability svc: %+v", env)
			}
			if env.ErrorKey == "" && status != http.StatusOK {
				t.Fatalf("expected stable error_key for app-layer failure; status=%d env=%+v", status, env)
			}
		})
	}
}

func TestCapabilityHandler_WriteErrorEnvelopeMapping(t *testing.T) {
	h, _, tenantID := newCapabilityHandlerEnv(t)
	// Nil CapabilitySVC forces CapabilityNotConfigured through writeError.
	formaapp.ApplicationSVC.CapabilitySVC = nil

	rid := "req-mapping-confirm"
	status, env := performCapabilityJSON(t, h, http.MethodPost,
		"/api/forma/v1/businesses/biz_test/capability-proposals/prop_1/confirm",
		tenantID, rid, []byte(`{}`))
	if status != http.StatusInternalServerError {
		t.Fatalf("status=%d want 500 env=%+v", status, env)
	}
	if env.Code != formaerrors.CodeCapabilityNotConfigured {
		t.Fatalf("code=%d want %d", env.Code, formaerrors.CodeCapabilityNotConfigured)
	}
	if env.ErrorKey != formaerrors.KeyCapabilityNotConfigured {
		t.Fatalf("error_key=%q want %q", env.ErrorKey, formaerrors.KeyCapabilityNotConfigured)
	}
	if env.RequestID != rid {
		t.Fatalf("request_id=%q want %q", env.RequestID, rid)
	}
}

// panicOnInvokeCapabilitySVC fails the test if any domain capability method is reached.
// Used to prove malformed Retry never enters ApplicationSVC.RetryCapabilityAnalysis.
type panicOnInvokeCapabilitySVC struct{}

func (panicOnInvokeCapabilitySVC) fail(op string) {
	panic("ApplicationSVC must not be invoked for malformed JSON; reached CapabilitySVC." + op)
}

func (s panicOnInvokeCapabilitySVC) ManualCreate(context.Context, *capsvc.ManualCreateInput) (*capentity.BusinessCapability, *capentity.BusinessCapabilityRevision, error) {
	s.fail("ManualCreate")
	return nil, nil, nil
}
func (s panicOnInvokeCapabilitySVC) DeriveRevision(context.Context, *capsvc.DeriveInput) (*capentity.BusinessCapabilityRevision, *capentity.CapabilityDecision, error) {
	s.fail("DeriveRevision")
	return nil, nil, nil
}
func (s panicOnInvokeCapabilitySVC) StartAnalysis(context.Context, *capsvc.StartAnalysisInput) (*capsvc.AnalysisResult, error) {
	s.fail("StartAnalysis")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) GetAnalysisRun(context.Context, string, string) (*capentity.CapabilityAnalysisRun, error) {
	s.fail("GetAnalysisRun")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) RetryFailedAnalysis(context.Context, string, string, string) (*capsvc.AnalysisResult, error) {
	s.fail("RetryFailedAnalysis")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) ConfirmProposal(context.Context, *capsvc.ConfirmInput) (*capentity.BusinessCapabilityRevision, error) {
	s.fail("ConfirmProposal")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) EditConfirmProposal(context.Context, *capsvc.EditConfirmInput) (*capentity.BusinessCapabilityRevision, error) {
	s.fail("EditConfirmProposal")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) RejectProposal(context.Context, *capsvc.RejectInput) (*capentity.CapabilityDecision, error) {
	s.fail("RejectProposal")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) GetCapability(context.Context, string, string) (*capentity.BusinessCapability, error) {
	s.fail("GetCapability")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) ListCapabilities(context.Context, string, string) ([]*capentity.BusinessCapability, error) {
	s.fail("ListCapabilities")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) GetRevision(context.Context, string, string) (*capentity.BusinessCapabilityRevision, error) {
	s.fail("GetRevision")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) ListRevisions(context.Context, string, string) ([]*capentity.BusinessCapabilityRevision, error) {
	s.fail("ListRevisions")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) GetProposal(context.Context, string, string) (*capentity.CapabilityProposal, error) {
	s.fail("GetProposal")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) ListValidations(context.Context, string, string) ([]*capentity.CapabilityValidationResult, error) {
	s.fail("ListValidations")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) ListDecisions(context.Context, string, string) ([]*capentity.CapabilityDecision, error) {
	s.fail("ListDecisions")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) Validate(context.Context, string, string, string) (*capentity.BusinessCapabilityRevision, *capentity.CapabilityValidationResult, error) {
	s.fail("Validate")
	return nil, nil, nil
}
func (s panicOnInvokeCapabilitySVC) Activate(context.Context, string, string, string, string) (*capentity.BusinessCapabilityRevision, error) {
	s.fail("Activate")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) MarkStale(context.Context, string, string, string, string) (*capentity.BusinessCapabilityRevision, error) {
	s.fail("MarkStale")
	return nil, nil
}
func (s panicOnInvokeCapabilitySVC) Deprecate(context.Context, string, string, string, string) (*capentity.BusinessCapabilityRevision, error) {
	s.fail("Deprecate")
	return nil, nil
}

var _ capsvc.CapabilityService = panicOnInvokeCapabilitySVC{}

// panicOnInvokeBusinessSVC fails if ApplicationSVC reaches business resolution (Retry/Create path).
type panicOnInvokeBusinessSVC struct{}

func (panicOnInvokeBusinessSVC) fail(op string) {
	panic("ApplicationSVC must not be invoked for malformed JSON; reached BusinessSVC." + op)
}

func (s panicOnInvokeBusinessSVC) InitBusiness(context.Context, string, string, string, string, *bizentity.SemanticModel, string) (*bizentity.BusinessModel, *bizentity.BusinessModelRevision, *bizentity.BusinessModelLayout, error) {
	s.fail("InitBusiness")
	return nil, nil, nil, nil
}
func (s panicOnInvokeBusinessSVC) Get(context.Context, string, string) (*bizentity.BusinessModel, error) {
	s.fail("Get")
	return nil, nil
}
func (s panicOnInvokeBusinessSVC) List(context.Context, string) ([]*bizentity.BusinessModel, error) {
	s.fail("List")
	return nil, nil
}
func (s panicOnInvokeBusinessSVC) GetModel(context.Context, string, string) (*bizentity.BusinessModel, *bizentity.SemanticModel, *bizentity.BusinessModelRevision, error) {
	s.fail("GetModel")
	return nil, nil, nil, nil
}
func (s panicOnInvokeBusinessSVC) SaveModel(context.Context, string, string, string, int32, *bizentity.SemanticModel, string) (*bizentity.BusinessModelRevision, bool, error) {
	s.fail("SaveModel")
	return nil, false, nil
}
func (s panicOnInvokeBusinessSVC) ListRevisions(context.Context, string, string) ([]*bizentity.BusinessModelRevision, error) {
	s.fail("ListRevisions")
	return nil, nil
}
func (s panicOnInvokeBusinessSVC) GetRevision(context.Context, string, string, int32) (*bizentity.BusinessModelRevision, *bizentity.SemanticModel, error) {
	s.fail("GetRevision")
	return nil, nil, nil
}
func (s panicOnInvokeBusinessSVC) Diff(context.Context, string, string, int32, int32) (*bizentity.BusinessModelDiff, *bizentity.BusinessImpactSummary, error) {
	s.fail("Diff")
	return nil, nil, nil
}
func (s panicOnInvokeBusinessSVC) GetLayout(context.Context, string, string) (*bizentity.BusinessModelLayout, *bizentity.ViewLayout, error) {
	s.fail("GetLayout")
	return nil, nil, nil
}
func (s panicOnInvokeBusinessSVC) SaveLayout(context.Context, string, string, string, int32, int32, *bizentity.ViewLayout) (*bizentity.BusinessModelLayout, error) {
	s.fail("SaveLayout")
	return nil, nil
}

var _ businesssvc.BusinessService = panicOnInvokeBusinessSVC{}

func assertStableBadRequestEnvelope(t *testing.T, status int, env capabilityEnvelope, wantRequestID string) {
	t.Helper()
	if status != http.StatusBadRequest {
		t.Fatalf("status=%d want 400 env=%+v", status, env)
	}
	if env.Code != formaerrors.CodeAdminBadRequest {
		t.Fatalf("code=%d want %d", env.Code, formaerrors.CodeAdminBadRequest)
	}
	if env.ErrorKey != formaerrors.KeyAdminBadRequest {
		t.Fatalf("error_key=%q want %q", env.ErrorKey, formaerrors.KeyAdminBadRequest)
	}
	if env.ErrorKey == "" {
		t.Fatalf("error_key must be present")
	}
	if env.RequestID != wantRequestID {
		t.Fatalf("request_id=%q want %q", env.RequestID, wantRequestID)
	}
	if env.RequestID == "" {
		t.Fatalf("request_id must be present")
	}
}

func TestCapabilityRetry_MalformedJSONDoesNotInvokeApplicationSVC(t *testing.T) {
	h, _, tenantID := newCapabilityHandlerEnv(t)
	// Panic stubs: if bind fails closed, ApplicationSVC.RetryCapabilityAnalysis is never entered
	// (BusinessSVC.Get / CapabilitySVC.* would panic if the app layer ran).
	formaapp.ApplicationSVC.CapabilitySVC = panicOnInvokeCapabilitySVC{}
	formaapp.ApplicationSVC.BusinessSVC = panicOnInvokeBusinessSVC{}

	rid := "req-malformed-retry"
	status, env := performCapabilityJSON(t, h, http.MethodPost,
		"/api/forma/v1/businesses/biz_test/capability-analyses/run_1/retry",
		tenantID, rid, []byte(`{"reason":`))
	assertStableBadRequestEnvelope(t, status, env, rid)
}

func TestCapabilityCreate_MalformedJSONStableBadRequest(t *testing.T) {
	h, _, tenantID := newCapabilityHandlerEnv(t)
	formaapp.ApplicationSVC.CapabilitySVC = panicOnInvokeCapabilitySVC{}
	formaapp.ApplicationSVC.BusinessSVC = panicOnInvokeBusinessSVC{}

	rid := "req-malformed-create"
	status, env := performCapabilityJSON(t, h, http.MethodPost,
		"/api/forma/v1/businesses/biz_test/capabilities",
		tenantID, rid, []byte(`{"payload":`))
	assertStableBadRequestEnvelope(t, status, env, rid)
}

func TestCapabilityRetry_EmptyBodyOKPathMapping(t *testing.T) {
	h, _, tenantID := newCapabilityHandlerEnv(t)
	cases := []struct {
		name string
		body []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"ws", []byte("   \n")},
		{"empty-object", []byte(`{}`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rid := "req-retry-empty-" + tc.name
			status, env := performCapabilityJSON(t, h, http.MethodPost,
				"/api/forma/v1/businesses/biz_test/capability-analyses/run_1/retry",
				tenantID, rid, tc.body)
			if env.ErrorKey == formaerrors.KeyAdminBadRequest {
				t.Fatalf("empty optional retry body must not map to BadRequest; status=%d env=%+v", status, env)
			}
			if env.RequestID != rid {
				t.Fatalf("request_id=%q want %q", env.RequestID, rid)
			}
			if env.RequestID == "" {
				t.Fatalf("request_id must be present")
			}
			// Past bind: stub path → business not configured / analysis not found / mapped error.
			if status == http.StatusOK {
				t.Fatalf("unexpected OK with stub svc: %+v", env)
			}
			if env.ErrorKey == "" {
				t.Fatalf("expected stable error_key for app-layer failure; status=%d env=%+v", status, env)
			}
		})
	}
}
