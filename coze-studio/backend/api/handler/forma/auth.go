/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/coze-dev/coze-studio/backend/api/model/passport"
	"github.com/coze-dev/coze-studio/backend/application/user"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	"github.com/coze-dev/coze-studio/backend/domain/user/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/hertzutil/domain"
)

// Logout is a Forma-owned auth facade over Coze SessionAuth.
// It revokes the Coze session via the existing user application service,
// then expires the HttpOnly session cookie. It does not reimplement login.
func Logout(ctx context.Context, c *app.RequestContext) {
	_, err := user.UserApplicationSVC.PassportWebLogoutGet(ctx, &passport.PassportWebLogoutGetRequest{})
	if err != nil {
		writeError(ctx, c, formaerrors.MapDomainError(err))
		return
	}

	// Expire HttpOnly session cookie for the browser origin. Cookie secret is never read by Forma JS.
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

	writeOK(ctx, c, map[string]any{
		"logged_out": true,
	})
}
