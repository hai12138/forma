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

// AdminListUsers GET /api/forma/v1/admin/users
func AdminListUsers(ctx context.Context, c *app.RequestContext) {
	users, err := formaapp.ApplicationSVC.AdminListUsers(ctx)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, users)
}

// AdminCreateUser POST /api/forma/v1/admin/users
func AdminCreateUser(ctx context.Context, c *app.RequestContext) {
	var in formaapp.AdminCreateUserInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request body"))
		return
	}
	resp, err := formaapp.ApplicationSVC.AdminCreateUser(ctx, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

// AdminGetUser GET /api/forma/v1/admin/users/:principalId
func AdminGetUser(ctx context.Context, c *app.RequestContext) {
	principalID := c.Param("principalId")
	if principalID == "" {
		writeError(ctx, c, formaerrors.BadRequest("principalId is required"))
		return
	}
	user, err := formaapp.ApplicationSVC.AdminGetUser(ctx, principalID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, user)
}

// AdminDisableUser POST /api/forma/v1/admin/users/:principalId/disable
func AdminDisableUser(ctx context.Context, c *app.RequestContext) {
	principalID := c.Param("principalId")
	if principalID == "" {
		writeError(ctx, c, formaerrors.BadRequest("principalId is required"))
		return
	}
	if err := formaapp.ApplicationSVC.AdminDisableUser(ctx, principalID); err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, map[string]any{"disabled": true})
}

// AdminEnableUser POST /api/forma/v1/admin/users/:principalId/enable
func AdminEnableUser(ctx context.Context, c *app.RequestContext) {
	principalID := c.Param("principalId")
	if principalID == "" {
		writeError(ctx, c, formaerrors.BadRequest("principalId is required"))
		return
	}
	if err := formaapp.ApplicationSVC.AdminEnableUser(ctx, principalID); err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, map[string]any{"enabled": true})
}

// AdminResetPassword POST /api/forma/v1/admin/users/:principalId/reset-password
func AdminResetPassword(ctx context.Context, c *app.RequestContext) {
	principalID := c.Param("principalId")
	if principalID == "" {
		writeError(ctx, c, formaerrors.BadRequest("principalId is required"))
		return
	}
	var in formaapp.AdminResetPasswordInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request body"))
		return
	}
	if err := formaapp.ApplicationSVC.AdminResetPassword(ctx, principalID, &in); err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, map[string]any{"password_reset": true})
}
