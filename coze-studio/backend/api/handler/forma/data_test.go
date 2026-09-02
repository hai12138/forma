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

func TestDataRequirementAnalyzeRouteIsRegistered(t *testing.T) {
	formaapp.ApplicationSVC = &formaapp.ApplicationService{}
	h := server.Default()
	formaRouter.Register(h)
	body := []byte(`{"business_model_revision":1,"client_request_id":"request-1"}`)

	w := ut.PerformRequest(
		h.Engine,
		"POST",
		"/api/forma/v1/businesses/biz_test/data-requirements/analyze",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "Content-Type", Value: "application/json"},
	)
	if w.Code == 404 {
		t.Fatalf("data requirement analyze route was not registered: %s", w.Body.String())
	}
}
