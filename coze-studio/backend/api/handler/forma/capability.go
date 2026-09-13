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

func ListCapabilities(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilities(ctx, c.Param("id"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func CreateCapability(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateCapabilityInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.CreateCapability(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func GetCapability(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetCapability(ctx, c.Param("id"), c.Param("capabilityId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListCapabilityRevisions(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilityRevisions(ctx, c.Param("id"), c.Param("capabilityId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func GetCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetCapabilityRevision(ctx, c.Param("id"), c.Param("capabilityId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func DeriveCapability(ctx context.Context, c *app.RequestContext) {
	var in formaapp.DeriveCapabilityInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.DeriveCapability(ctx, c.Param("id"), c.Param("capabilityId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func EditCapability(ctx context.Context, c *app.RequestContext) {
	var in formaapp.EditCapabilityInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.EditCapability(ctx, c.Param("id"), c.Param("capabilityId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func StartCapabilityAnalysis(ctx context.Context, c *app.RequestContext) {
	var in formaapp.StartCapabilityAnalysisInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.StartCapabilityAnalysis(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func GetCapabilityAnalysis(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetCapabilityAnalysis(ctx, c.Param("id"), c.Param("analysisRunId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ConfirmCapabilityProposal(ctx context.Context, c *app.RequestContext) {
	var in formaapp.ConfirmCapabilityProposalInput
	_ = c.BindAndValidate(&in)
	v, err := formaapp.ApplicationSVC.ConfirmCapabilityProposal(ctx, c.Param("id"), c.Param("proposalId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func EditConfirmCapabilityProposal(ctx context.Context, c *app.RequestContext) {
	var in formaapp.EditConfirmCapabilityProposalInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.EditConfirmCapabilityProposal(ctx, c.Param("id"), c.Param("proposalId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func RejectCapabilityProposal(ctx context.Context, c *app.RequestContext) {
	var in formaapp.RejectCapabilityProposalInput
	_ = c.BindAndValidate(&in)
	v, err := formaapp.ApplicationSVC.RejectCapabilityProposal(ctx, c.Param("id"), c.Param("proposalId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ValidateCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ValidateCapabilityRevision(ctx, c.Param("id"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ActivateCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CapabilityReasonInput
	_ = c.BindAndValidate(&in)
	v, err := formaapp.ApplicationSVC.ActivateCapabilityRevision(ctx, c.Param("id"), c.Param("revisionId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func DeprecateCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CapabilityReasonInput
	_ = c.BindAndValidate(&in)
	v, err := formaapp.ApplicationSVC.DeprecateCapabilityRevision(ctx, c.Param("id"), c.Param("revisionId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListCapabilityValidations(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilityValidations(ctx, c.Param("id"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

func ListCapabilityDecisions(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilityDecisions(ctx, c.Param("id"), c.Param("capabilityId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}
