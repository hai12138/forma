/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	"github.com/coze-dev/coze-studio/backend/application/user"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
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
// Maps the bootstrap admin alias "admin" to "admin@forma.local",
// then delegates to Coze User domain for authentication.
func FormaLogin(ctx context.Context, c *app.RequestContext) {
	var req loginRequest
	if err := c.BindAndValidate(&req); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid login request"))
		return
	}

	account := strings.TrimSpace(req.Account)
	password := req.Password

	if account == "" || password == "" {
		writeError(ctx, c, formaerrors.BadRequest("account and password required"))
		return
	}

	// Map bootstrap admin alias to internal email
	email := account
	if account == formaapp.DefaultBootstrapAdminUsername {
		email = formaapp.DefaultBootstrapAdminEmail
	} else if !strings.Contains(account, "@") {
		// Non-email usernames are mapped to @forma.local
		email = account + "@forma.local"
	}

	// Delegate to Coze user domain
	cozeUser, err := user.UserApplicationSVC.DomainSVC.Login(ctx, email, password)
	if err != nil {
		writeError(ctx, c, formaerrors.Unauthenticated("invalid credentials"))
		return
	}

	// Set session cookie (same pattern as Coze passport handler)
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

	// Check if password change is required
	principal, _ := formaapp.ApplicationSVC.TenancySVC.ResolveOrCreatePrincipal(ctx, cozeUser.UserID, cozeUser.Name)
	passwordChangeRequired := false
	principalID := ""
	if principal != nil {
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
