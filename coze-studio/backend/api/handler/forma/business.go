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
)

func CreateBusiness(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateBusinessInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.CreateBusiness(ctx, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListBusinesses(ctx context.Context, c *app.RequestContext) {
	resp, err := formaapp.ApplicationSVC.ListBusinesses(ctx)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetBusiness(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.GetBusiness(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func PatchBusiness(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var in formaapp.PatchBusinessInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.PatchBusiness(ctx, id, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ArchiveBusiness(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ArchiveBusiness(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetBusinessModel(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.GetBusinessModel(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func PutBusinessModel(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var in formaapp.PutModelInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.PutBusinessModel(ctx, id, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func ListBusinessRevisions(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.ListBusinessRevisions(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetBusinessRevision(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	revStr := c.Param("revision")
	rev, err := strconv.ParseInt(revStr, 10, 32)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.GetBusinessRevision(ctx, id, int32(rev))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func DiffBusiness(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	from, _ := strconv.ParseInt(string(c.Query("from")), 10, 32)
	to, _ := strconv.ParseInt(string(c.Query("to")), 10, 32)
	resp, err := formaapp.ApplicationSVC.DiffBusiness(ctx, id, int32(from), int32(to))
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func GetBusinessLayout(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	resp, err := formaapp.ApplicationSVC.GetBusinessLayout(ctx, id)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}

func PutBusinessLayout(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var in formaapp.PutLayoutInput
	if err := c.BindAndValidate(&in); err != nil {
		writeError(ctx, c, err)
		return
	}
	resp, err := formaapp.ApplicationSVC.PutBusinessLayout(ctx, id, &in)
	if err != nil {
		writeError(ctx, c, err)
		return
	}
	writeOK(ctx, c, resp)
}
