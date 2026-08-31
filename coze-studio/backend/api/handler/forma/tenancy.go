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

func Me(ctx context.Context, c *app.RequestContext) {
	resp, err := formaapp.ApplicationSVC.Me(ctx)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListTenants(ctx context.Context, c *app.RequestContext) {
	resp, err := formaapp.ApplicationSVC.ListTenants(ctx)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func CreateTenant(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateTenantInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.CreateTenant(ctx, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetTenant(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.GetTenant(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func PatchTenant(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var in formaapp.PatchTenantInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.PatchTenant(ctx, id, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListMembers(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListMembers(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func AddMember(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var in formaapp.AddMemberInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.AddMember(ctx, id, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func PatchMember(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	principalID := c.Param("principalId")
	var in formaapp.PatchMemberInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.PatchMember(ctx, id, principalID, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListSpaces(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListSpaces(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func BindSpace(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var in formaapp.BindSpaceInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.BindSpace(ctx, id, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func Bootstrap(ctx context.Context, c *app.RequestContext) {
	var in formaapp.BootstrapInput
	_ = c.BindAndValidate(&in)
	resp, err := formaapp.ApplicationSVC.Bootstrap(ctx, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}
