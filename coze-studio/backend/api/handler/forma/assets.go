/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
)

func AssetCounts(ctx context.Context, c *app.RequestContext) {
	resp, err := formaapp.ApplicationSVC.AssetCounts(ctx)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}
