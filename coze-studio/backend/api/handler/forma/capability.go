/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"bytes"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

// ListCapabilities lists capabilities for a business.
func ListCapabilities(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilities(ctx, c.Param("id"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// CreateCapability creates a capability from a required JSON body.
// Malformed JSON maps to stable BadRequest; raw bind errors are never returned.
func CreateCapability(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateCapabilityInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request body"))
		return
	}
	v, err := formaapp.ApplicationSVC.CreateCapability(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// GetCapability returns a capability by id.
func GetCapability(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetCapability(ctx, c.Param("id"), c.Param("capabilityId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// ListCapabilityRevisions lists revisions for a capability.
func ListCapabilityRevisions(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilityRevisions(ctx, c.Param("id"), c.Param("capabilityId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// GetCapabilityRevision returns a capability revision by id.
func GetCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetCapabilityRevision(ctx, c.Param("id"), c.Param("capabilityId"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// DeriveCapability derives a new revision from a required JSON body.
// Malformed JSON maps to stable BadRequest; raw bind errors are never returned.
func DeriveCapability(ctx context.Context, c *app.RequestContext) {
	var in formaapp.DeriveCapabilityInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request body"))
		return
	}
	v, err := formaapp.ApplicationSVC.DeriveCapability(ctx, c.Param("id"), c.Param("capabilityId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// EditCapability edits a capability via a required JSON body.
// Malformed JSON maps to stable BadRequest; raw bind errors are never returned.
func EditCapability(ctx context.Context, c *app.RequestContext) {
	var in formaapp.EditCapabilityInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request body"))
		return
	}
	v, err := formaapp.ApplicationSVC.EditCapability(ctx, c.Param("id"), c.Param("capabilityId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// StartCapabilityAnalysis starts analysis from a required JSON body.
// Malformed JSON maps to stable BadRequest; raw bind errors are never returned.
func StartCapabilityAnalysis(ctx context.Context, c *app.RequestContext) {
	var in formaapp.StartCapabilityAnalysisInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request body"))
		return
	}
	v, err := formaapp.ApplicationSVC.StartCapabilityAnalysis(ctx, c.Param("id"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// GetCapabilityAnalysis returns an analysis run by id.
func GetCapabilityAnalysis(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetCapabilityAnalysis(ctx, c.Param("id"), c.Param("analysisRunId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// GetCapabilityProposal returns a capability proposal by id.
func GetCapabilityProposal(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.GetCapabilityProposal(ctx, c.Param("id"), c.Param("proposalId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// ConfirmCapabilityProposal confirms a capability proposal.
// Empty body is treated as empty optional input; non-empty malformed JSON fails closed with HTTP 400.
func ConfirmCapabilityProposal(ctx context.Context, c *app.RequestContext) {
	var in formaapp.ConfirmCapabilityProposalInput
	if err := bindOptionalCapabilityJSON(c, &in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.ConfirmCapabilityProposal(ctx, c.Param("id"), c.Param("proposalId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// EditConfirmCapabilityProposal edit-confirms a capability proposal with an effective payload.
func EditConfirmCapabilityProposal(ctx context.Context, c *app.RequestContext) {
	var in formaapp.EditConfirmCapabilityProposalInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, formaerrors.BadRequest("invalid request body"))
		return
	}
	v, err := formaapp.ApplicationSVC.EditConfirmCapabilityProposal(ctx, c.Param("id"), c.Param("proposalId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// RejectCapabilityProposal rejects a capability proposal.
// Empty body is treated as empty optional input; non-empty malformed JSON fails closed with HTTP 400.
func RejectCapabilityProposal(ctx context.Context, c *app.RequestContext) {
	var in formaapp.RejectCapabilityProposalInput
	if err := bindOptionalCapabilityJSON(c, &in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.RejectCapabilityProposal(ctx, c.Param("id"), c.Param("proposalId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// ValidateCapabilityRevision validates a capability revision.
func ValidateCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ValidateCapabilityRevision(ctx, c.Param("id"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// ActivateCapabilityRevision activates a capability revision.
// Empty body is treated as empty optional input; non-empty malformed JSON fails closed with HTTP 400.
func ActivateCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CapabilityReasonInput
	if err := bindOptionalCapabilityJSON(c, &in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.ActivateCapabilityRevision(ctx, c.Param("id"), c.Param("revisionId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// DeprecateCapabilityRevision deprecates a capability revision.
// Empty body is treated as empty optional input; non-empty malformed JSON fails closed with HTTP 400.
func DeprecateCapabilityRevision(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CapabilityReasonInput
	if err := bindOptionalCapabilityJSON(c, &in); err != nil {
		writeError(ctx, c, err)
		return
	}
	v, err := formaapp.ApplicationSVC.DeprecateCapabilityRevision(ctx, c.Param("id"), c.Param("revisionId"), &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// ListCapabilityValidations lists validation results for a revision.
func ListCapabilityValidations(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilityValidations(ctx, c.Param("id"), c.Param("revisionId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// ListCapabilityDecisions lists decisions for a capability.
func ListCapabilityDecisions(ctx context.Context, c *app.RequestContext) {
	v, err := formaapp.ApplicationSVC.ListCapabilityDecisions(ctx, c.Param("id"), c.Param("capabilityId"))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, v)
}

// bindOptionalCapabilityJSON binds an optional JSON body.
// Empty / whitespace-only bodies skip bind (zero-value input). Non-empty malformed JSON returns BadRequest.
func bindOptionalCapabilityJSON(c *app.RequestContext, in any) error {
	body := c.Request.Body()
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := c.BindAndValidate(in); err != nil {
		return formaerrors.BadRequest("invalid request body")
	}
	return nil
}
