/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"bytes"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"

	formaRouter "github.com/coze-dev/coze-studio/backend/api/router/forma"
	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
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
