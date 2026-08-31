/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

func Health(ctx context.Context, c *app.RequestContext) {
	resp := formaapp.ApplicationSVC.Health(ctx)
	c.JSON(http.StatusOK, envelope(ctx, c, resp))
}

func Version(ctx context.Context, c *app.RequestContext) {
	resp := formaapp.ApplicationSVC.Version(ctx)
	c.JSON(http.StatusOK, envelope(ctx, c, resp))
}

func Baseline(ctx context.Context, c *app.RequestContext) {
	resp := formaapp.ApplicationSVC.Baseline(ctx)
	c.JSON(http.StatusOK, envelope(ctx, c, resp))
}

type apiEnvelope struct {
	Code      int32  `json:"code"`
	Msg       string `json:"msg"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
	ErrorKey  string `json:"error_key,omitempty"`
}

func envelope(ctx context.Context, c *app.RequestContext, data any) apiEnvelope {
	return apiEnvelope{
		Code:      formaerrors.CodeOK,
		Msg:       "ok",
		RequestID: requestID(c),
		Data:      data,
	}
}

func errorEnvelope(c *app.RequestContext, err error) (int, apiEnvelope) {
	fe, ok := formaerrors.AsFormaError(err)
	if !ok {
		fe = formaerrors.MapDomainError(err)
	}
	msg := fe.Msg
	if fe.Key != "" {
		msg = fe.Key + ": " + fe.Msg
	}
	return fe.HTTPStatus, apiEnvelope{
		Code:      fe.Code,
		Msg:       msg,
		RequestID: requestID(c),
		Data:      nil,
		ErrorKey:  fe.Key,
	}
}

func writeError(ctx context.Context, c *app.RequestContext, err error) {
	_ = ctx
	status, body := errorEnvelope(c, err)
	c.JSON(status, body)
}

func writeOK(ctx context.Context, c *app.RequestContext, data any) {
	c.JSON(http.StatusOK, envelope(ctx, c, data))
}

func requestID(c *app.RequestContext) string {
	rid := string(c.GetHeader("X-Request-ID"))
	if rid == "" {
		rid = c.Response.Header.Get("X-Request-ID")
	}
	if rid == "" {
		rid = "forma-" + c.GetString("request_id")
	}
	return rid
}
