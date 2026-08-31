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
}

func envelope(ctx context.Context, c *app.RequestContext, data any) apiEnvelope {
	rid := string(c.GetHeader("X-Request-ID"))
	if rid == "" {
		rid = c.Response.Header.Get("X-Request-ID")
	}
	if rid == "" {
		rid = "forma-" + c.GetString("request_id")
	}
	_ = ctx
	return apiEnvelope{
		Code:      0,
		Msg:       "ok",
		RequestID: rid,
		Data:      data,
	}
}
