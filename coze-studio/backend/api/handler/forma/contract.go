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

func CreateDataContract(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateDataContractInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.CreateDataContract(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListDataContracts(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListDataContracts(ctx, c.Param("id"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func GetDataContract(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetDataContract(ctx, c.Param("id"), c.Param("contractId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListDataContractRevisions(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListDataContractRevisions(ctx, c.Param("id"), c.Param("contractId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func CreateDataContractRevision(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateDataContractRevisionInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.CreateDataContractRevision(ctx, c.Param("id"), c.Param("contractId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func GetDataContractRevision(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetDataContractRevision(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ValidateDataContractRevision(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ValidateDataContractRevision(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ActivateDataContractRevision(ctx context.Context, c *app.RequestContext) {
	var in formaapp.ContractReasonInput
	_ = c.BindAndValidate(&in)
	v, err := formaapp.ApplicationSVC.ActivateDataContractRevision(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func DeprecateDataContractRevision(ctx context.Context, c *app.RequestContext) {
	var in formaapp.ContractReasonInput
	_ = c.BindAndValidate(&in)
	v, err := formaapp.ApplicationSVC.DeprecateDataContractRevision(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListDataContractValidationResults(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListDataContractValidationResults(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func EvaluateDataContractDrift(ctx context.Context, c *app.RequestContext) {
	var in formaapp.EvaluateDriftAppInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.EvaluateDataContractDrift(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListDataContractDriftResults(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListDataContractDriftResults(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func EvaluateDataContractGap(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.EvaluateDataContractGap(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListDataContractGapResults(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListDataContractGapResults(ctx, c.Param("id"), c.Param("contractId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListDataContractLifecycleEvents(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListDataContractLifecycleEvents(ctx, c.Param("id"), c.Param("contractId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
