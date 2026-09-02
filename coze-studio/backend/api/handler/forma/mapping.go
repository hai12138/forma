/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

func mappingRevision(ctx context.Context, c *app.RequestContext) (int32, bool) {
	v, err := strconv.ParseInt(string(c.Query("business_model_revision")), 10, 32)
	if err != nil || v <= 0 {
		writeError(ctx, c, formaerrors.DataRequirementInvalid("business_model_revision query parameter required"))
		return 0, false
	}
	return int32(v), true
}
func AnalyzeSemanticMappings(ctx context.Context, c *app.RequestContext) {
	var in formaapp.AnalyzeSemanticMappingsInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.AnalyzeSemanticMappings(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func GetMappingAnalysisRun(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetMappingAnalysisRun(ctx, c.Param("id"), c.Param("analysisRunId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func RetryFailedMappingAnalysis(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.RetryFailedMappingAnalysis(ctx, c.Param("id"), c.Param("analysisRunId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func ListSemanticMappings(ctx context.Context, c *app.RequestContext) {
	rev, ok := mappingRevision(ctx, c)
	if !ok {
		return
	}
	v, err := formaapp.ApplicationSVC.ListSemanticMappings(ctx, c.Param("id"), rev, string(c.Query("status")))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func CreateManualSemanticMapping(ctx context.Context, c *app.RequestContext) {
	var in formaapp.SemanticMappingInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.CreateManualSemanticMapping(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func ConfirmSemanticMapping(ctx context.Context, c *app.RequestContext) {
	decideSemanticMapping(ctx, c, "confirm")
}
func RejectSemanticMapping(ctx context.Context, c *app.RequestContext) {
	decideSemanticMapping(ctx, c, "reject")
}
func decideSemanticMapping(ctx context.Context, c *app.RequestContext, action string) {
	var in formaapp.DecideDataRequirementInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.DecideSemanticMapping(ctx, c.Param("id"), c.Param("mappingId"), action, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func EditConfirmSemanticMapping(ctx context.Context, c *app.RequestContext) {
	var in formaapp.EditConfirmSemanticMappingInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.EditConfirmSemanticMapping(ctx, c.Param("id"), c.Param("mappingId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func ListSemanticMappingDecisions(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListSemanticMappingDecisions(ctx, c.Param("id"), c.Param("mappingId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
func GetSemanticMappingCoverage(ctx context.Context, c *app.RequestContext) {
	rev, ok := mappingRevision(ctx, c)
	if !ok {
		return
	}
	v, err := formaapp.ApplicationSVC.GetSemanticMappingCoverage(ctx, c.Param("id"), rev)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
