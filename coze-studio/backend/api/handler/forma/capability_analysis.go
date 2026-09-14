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
// Empty body is treated as empty optional input; non-empty malformed JSON fails closed with HTTP 400.
func RetryCapabilityAnalysis(ctx context.Context, c *app.RequestContext) {
	var in formaapp.RetryCapabilityAnalysisInput
	if err := bindOptionalCapabilityJSON(c, &in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.RetryCapabilityAnalysis(ctx, c.Param("id"), c.Param("analysisRunId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// ListCapabilityProposalsByAnalysis lists proposals for a capability analysis run.
func ListCapabilityProposalsByAnalysis(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilityProposalsByAnalysis(ctx, c.Param("id"), c.Param("analysisRunId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
