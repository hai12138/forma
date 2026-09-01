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

func CreateAnalystSession(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	var in formaapp.CreateAnalystSessionInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.CreateAnalystSession(ctx, businessID, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListAnalystSessions(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListAnalystSessions(ctx, businessID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetAnalystSession(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	sessionID := c.Param("sessionId")
	resp, err := formaapp.ApplicationSVC.GetAnalystSession(ctx, businessID, sessionID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func SubmitAnalystTurn(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	sessionID := c.Param("sessionId")
	var in formaapp.SubmitTurnInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.SubmitAnalystTurn(ctx, businessID, sessionID, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListAnalystTurns(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	sessionID := c.Param("sessionId")
	resp, err := formaapp.ApplicationSVC.ListAnalystTurns(ctx, businessID, sessionID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListAssertions(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListAssertions(ctx, businessID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListBusinessEvidence(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListEvidence(ctx, businessID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ConfirmAssertion(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	assertionID := c.Param("assertionId")
	var in formaapp.ConfirmAssertionInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.ConfirmAssertion(ctx, businessID, assertionID, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func RejectAssertion(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	assertionID := c.Param("assertionId")
	var in formaapp.ConfirmAssertionInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.RejectAssertion(ctx, businessID, assertionID, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListConflicts(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListConflicts(ctx, businessID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListGaps(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListGaps(ctx, businessID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func CreateProposal(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	var in formaapp.CreateProposalInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.CreateProposal(ctx, businessID, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetProposal(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	proposalID := c.Param("proposalId")
	resp, err := formaapp.ApplicationSVC.GetProposal(ctx, businessID, proposalID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ApplyProposal(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	proposalID := c.Param("proposalId")
	resp, err := formaapp.ApplicationSVC.ApplyProposal(ctx, businessID, proposalID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func RetryAnalystTurnAnalysis(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	sessionID := c.Param("sessionId")
	turnID := c.Param("turnId")
	resp, err := formaapp.ApplicationSVC.RetryAnalystTurnAnalysis(ctx, businessID, sessionID, turnID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetProposalPreview(ctx context.Context, c *app.RequestContext) {
	businessID := c.Param("id")
	proposalID := c.Param("proposalId")
	resp, err := formaapp.ApplicationSVC.GetProposalPreview(ctx, businessID, proposalID)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}
