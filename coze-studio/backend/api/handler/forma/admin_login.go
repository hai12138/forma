/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	"github.com/coze-dev/coze-studio/backend/application/user"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/domain/user/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/hertzutil/domain"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type loginResponse struct {
	PasswordChangeRequired bool   `json:"password_change_required"`
	PrincipalID            string `json:"principal_id,omitempty"`
}

// FormaLogin is the Forma-owned login endpoint.
// Applies Forma Local Account Alias normalization, then delegates to Coze User domain.
func FormaLogin(ctx context.Context, c *app.RequestContext) {
	var req loginRequest
	if err := c.BindAndValidate(&req); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid login request"))
		return
	}

	if req.Account == "" || req.Password == "" {
		writeError(ctx, c, formaerrors.BadRequest("account and password required"))
		return
	}

	normalized, err := formaapp.NormalizeAccount(req.Account)
	if err != nil {
		writeError(ctx, c, formaerrors.Unauthenticated("invalid credentials"))
		return
	}

	cozeUser, err := user.UserApplicationSVC.DomainSVC.Login(ctx, normalized.Email, req.Password)
	if err != nil {
		writeError(ctx, c, formaerrors.Unauthenticated("invalid credentials"))
		return
	}

	c.SetCookie(
		entity.SessionKey,
		cozeUser.SessionKey,
		consts.SessionMaxAgeSecond,
		"/",
		domain.GetOriginHost(c),
		protocol.CookieSameSiteDefaultMode,
		false,
		true,
	)

	principal, _ := formaapp.ApplicationSVC.TenancySVC.ResolveOrCreatePrincipal(ctx, cozeUser.UserID, cozeUser.Name)
	passwordChangeRequired := false
	principalID := ""
	if principal != nil {
		if principal.Status == tenancyentity.PrincipalStatusSuspended {
			_ = user.UserApplicationSVC.DomainSVC.Logout(ctx, cozeUser.UserID)
			c.SetCookie(
				entity.SessionKey,
				"",
				-1,
				"/",
				domain.GetOriginHost(c),
				protocol.CookieSameSiteDefaultMode,
				false,
				true,
			)
			writeError(ctx, c, formaerrors.Unauthenticated("account disabled"))
			return
		}
		principalID = principal.PrincipalID
		role, _ := formaapp.ApplicationSVC.TenancySVC.GetPlatformRole(ctx, principal.PrincipalID)
		if role != nil && role.PasswordChangeRequired {
			passwordChangeRequired = true
		}
	}

	writeOK(ctx, c, &loginResponse{
		PasswordChangeRequired: passwordChangeRequired,
		PrincipalID:            principalID,
	})
}
