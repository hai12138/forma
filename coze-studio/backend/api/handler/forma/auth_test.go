/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"

	formaRouter "github.com/coze-dev/coze-studio/backend/api/router/forma"
	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
)

func TestFormaLogoutRouteIsRegistered(t *testing.T) {
	formaapp.ApplicationSVC = &formaapp.ApplicationService{}
	h := server.Default()
	formaRouter.Register(h)
	w := ut.PerformRequest(h.Engine, "POST", "/api/forma/v1/auth/logout", nil)
	// Unauthenticated → handler/session rejects; route must exist (not 404).
	if w.Code == 404 {
		t.Fatalf("forma logout route was not registered: %s", w.Body.String())
	}
}
