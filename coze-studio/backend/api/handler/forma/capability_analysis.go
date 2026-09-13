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

// RetryCapabilityAnalysis retries a FAILED capability analysis run.
// Empty body is treated as empty optional input; non-empty malformed JSON is handled by BindAndValidate.
func RetryCapabilityAnalysis(ctx context.Context, c *app.RequestContext) {
	var in formaapp.RetryCapabilityAnalysisInput
	_ = c.BindAndValidate(&in)
	v, err := formaapp.ApplicationSVC.RetryCapabilityAnalysis(ctx, c.Param("id"), c.Param("analysisRunId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
