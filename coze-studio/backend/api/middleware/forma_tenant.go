/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package middleware

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
)

// FormaTenantMW resolves X-Forma-Tenant into TenantContext for protected Forma routes.
// Public meta routes should not use this middleware.
func FormaTenantMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		path := string(ctx.GetRequest().URI().Path())
		if isFormaPublicPath(path) || isFormaAuthOnlyPath(path, string(ctx.Method())) {
			ctx.Next(c)
			return
		}

		header := string(ctx.GetHeader(formaapp.HeaderFormaTenant))
		tc, err := formaapp.ApplicationSVC.ResolveTenantContext(c, header)
		if err != nil {
			writeFormaError(ctx, err)
			return
		}
		rid := string(ctx.GetHeader("X-Request-ID"))
		if rid != "" {
			tc.RequestID = rid
		}
		c = tenantctx.WithTenantContext(c, tc)
		ctx.Next(c)
	}
}

func isFormaPublicPath(path string) bool {
	switch path {
	case "/api/forma/v1/health", "/api/forma/v1/version", "/api/forma/v1/meta/baseline":
		return true
	default:
		return false
	}
}

// Auth-only routes need a session but do not require a resolved tenant context.
func isFormaAuthOnlyPath(path, method string) bool {
	switch path {
	case "/api/forma/v1/me", "/api/forma/v1/bootstrap":
		return true
	case "/api/forma/v1/tenants":
		return method == http.MethodGet || method == http.MethodPost
	default:
		return false
	}
}

func writeFormaError(ctx *app.RequestContext, err error) {
	fe, ok := formaerrors.AsFormaError(err)
	if !ok {
		fe = formaerrors.MapDomainError(err)
	}
	rid := string(ctx.GetHeader("X-Request-ID"))
	if rid == "" {
		rid = ctx.Response.Header.Get("X-Request-ID")
	}
	msg := fe.Msg
	if fe.Key != "" {
		msg = fe.Key + ": " + fe.Msg
	}
	ctx.JSON(fe.HTTPStatus, map[string]any{
		"code":       fe.Code,
		"msg":        msg,
		"request_id": rid,
		"data":       nil,
		"error_key":  fe.Key,
	})
	ctx.Abort()
}