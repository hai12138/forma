/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// FormaChangePassword handles POST /api/forma/v1/auth/change-password.
// Requires authentication. Changes the current user's password.
func FormaChangePassword(ctx context.Context, c *app.RequestContext) {
	var req changePasswordRequest
	if err := c.BindAndValidate(&req); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request"))
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(ctx, c, formaerrors.BadRequest("current_password and new_password required"))
		return
	}

	if err := formaapp.ApplicationSVC.AdminChangePassword(ctx, req.CurrentPassword, req.NewPassword); err != nil {
		writeError(ctx, c, err)
		return
	}

	writeOK(ctx, c, map[string]any{"password_changed": true})
}
