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

func AnalyzeDataRequirements(ctx context.Context, c *app.RequestContext) {
	var in formaapp.AnalyzeDataRequirementsInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.AnalyzeDataRequirements(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetDataAnalysisRun(ctx context.Context, c *app.RequestContext) {
	resp, err := formaapp.ApplicationSVC.GetDataAnalysisRun(ctx, c.Param("id"), c.Param("analysisRunId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func RetryFailedDataAnalysis(ctx context.Context, c *app.RequestContext) {
	resp, err := formaapp.ApplicationSVC.RetryFailedDataAnalysis(ctx, c.Param("id"), c.Param("analysisRunId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListDataRequirements(ctx context.Context, c *app.RequestContext) {
	revision, err := strconv.ParseInt(string(c.Query("business_model_revision")), 10, 32)
	if err != nil || revision <= 0 {
		writeError(ctx, c, formaerrors.DataRequirementInvalid("business_model_revision query parameter required"))
		return
	}
	resp, err := formaapp.ApplicationSVC.ListDataRequirements(ctx, c.Param("id"), int32(revision))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func CreateManualDataRequirement(ctx context.Context, c *app.RequestContext) {
	var in formaapp.ManualDataRequirementInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.CreateManualDataRequirement(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ConfirmDataRequirement(ctx context.Context, c *app.RequestContext) {
	var in formaapp.DecideDataRequirementInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.ConfirmDataRequirement(ctx, c.Param("id"), c.Param("requirementId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func RejectDataRequirement(ctx context.Context, c *app.RequestContext) {
	var in formaapp.DecideDataRequirementInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.RejectDataRequirement(ctx, c.Param("id"), c.Param("requirementId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func EditConfirmDataRequirement(ctx context.Context, c *app.RequestContext) {
	var in formaapp.EditConfirmDataRequirementInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.EditConfirmDataRequirement(ctx, c.Param("id"), c.Param("requirementId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListDataRequirementDecisions(ctx context.Context, c *app.RequestContext) {
	resp, err := formaapp.ApplicationSVC.ListDataRequirementDecisions(ctx, c.Param("id"), c.Param("requirementId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}
