/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
)

func AssetCounts(ctx context.Context, c *app.RequestContext) {
	tenantID := ""
	if tc, ok := tenantctx.FromContext(ctx); ok && tc != nil {
		tenantID = tc.TenantID
	}
	if tenantID == "" {
		tenantID = string(c.GetHeader(formaapp.HeaderFormaTenant))
	}
	resp, err := formaapp.ApplicationSVC.AssetCounts(ctx, tenantID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}
