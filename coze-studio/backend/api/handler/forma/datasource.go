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

func ListDataSources(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.ListDataSources(ctx)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func CreateDataSource(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateDataSourceInput
	if e := c.BindAndValidate(&in); e != nil {
		writeError(ctx, c, e)
		return
	}
	v, e := formaapp.ApplicationSVC.CreateDataSource(ctx, &in)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func GetDataSource(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.GetDataSource(ctx, c.Param("sourceId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func PatchDataSource(ctx context.Context, c *app.RequestContext) {
	var in formaapp.PatchDataSourceInput
	if e := c.BindAndValidate(&in); e != nil {
		writeError(ctx, c, e)
		return
	}
	v, e := formaapp.ApplicationSVC.PatchDataSource(ctx, c.Param("sourceId"), &in)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func ArchiveDataSource(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.ArchiveDataSource(ctx, c.Param("sourceId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func ListDataConnections(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.ListDataConnections(ctx, c.Param("sourceId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func CreateDataConnection(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateDataConnectionInput
	if e := c.BindAndValidate(&in); e != nil {
		writeError(ctx, c, e)
		return
	}
	v, e := formaapp.ApplicationSVC.CreateDataConnection(ctx, c.Param("sourceId"), &in)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func GetDataConnection(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.GetDataConnection(ctx, c.Param("sourceId"), c.Param("connectionId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func PatchDataConnection(ctx context.Context, c *app.RequestContext) {
	var in formaapp.PatchDataConnectionInput
	if e := c.BindAndValidate(&in); e != nil {
		writeError(ctx, c, e)
		return
	}
	v, e := formaapp.ApplicationSVC.PatchDataConnection(ctx, c.Param("sourceId"), c.Param("connectionId"), &in)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func TestDataConnection(ctx context.Context, c *app.RequestContext) {
	if e := formaapp.ApplicationSVC.TestDataConnection(ctx, c.Param("sourceId"), c.Param("connectionId")); e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, map[string]string{"status": "ok"})
}
func DiscoverDataAssets(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.DiscoverDataAssets(ctx, c.Param("sourceId"), c.Param("connectionId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func CreateDataCredential(ctx context.Context, c *app.RequestContext) {
	var in formaapp.CreateCredentialInput
	if e := c.BindAndValidate(&in); e != nil {
		writeError(ctx, c, e)
		return
	}
	v, e := formaapp.ApplicationSVC.CreateDataCredential(ctx, &in)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func RotateDataCredential(ctx context.Context, c *app.RequestContext) {
	var in formaapp.RotateCredentialInput
	if e := c.BindAndValidate(&in); e != nil {
		writeError(ctx, c, e)
		return
	}
	v, e := formaapp.ApplicationSVC.RotateDataCredential(ctx, c.Param("credentialRefId"), &in)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func RevokeDataCredential(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.RevokeDataCredential(ctx, c.Param("credentialRefId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func ListDataAssets(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.ListDataAssets(ctx, c.Param("sourceId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func GetDataAsset(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.GetDataAsset(ctx, c.Param("assetId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func CaptureDataSchema(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.CaptureDataSchema(ctx, c.Param("sourceId"), c.Param("connectionId"), c.Param("assetId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func GetDataSchemaSnapshot(ctx context.Context, c *app.RequestContext) {
	v, e := formaapp.ApplicationSVC.GetDataSchemaSnapshot(ctx, c.Param("snapshotId"))
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, v)
}
func ListDataSchemaSnapshots(ctx context.Context, c *app.RequestContext) {
	vs, e := formaapp.ApplicationSVC.ListDataSchemaSnapshots(
		ctx,
		string(c.Query("source_id")),
		string(c.Query("connection_id")),
		string(c.Query("asset_id")),
	)
	if e != nil {
		writeError(ctx, c, e)
		return
	}
	writeOK(ctx, c, vs)
}
