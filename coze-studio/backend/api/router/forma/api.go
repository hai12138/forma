/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	formahandler "github.com/coze-dev/coze-studio/backend/api/handler/forma"
	"github.com/coze-dev/coze-studio/backend/api/middleware"
)

func Register(r *server.Hertz) {
	root := r.Group("/api")
	v1 := root.Group("/forma/v1", middleware.FormaTenantMW())
	{
		v1.GET("/health", formahandler.Health)
		v1.GET("/version", formahandler.Version)
		v1.GET("/meta/baseline", formahandler.Baseline)

		v1.GET("/me", formahandler.Me)
		v1.GET("/tenants", formahandler.ListTenants)
		v1.POST("/tenants", formahandler.CreateTenant)
		v1.GET("/tenants/:id", formahandler.GetTenant)
		v1.PATCH("/tenants/:id", formahandler.PatchTenant)
		v1.GET("/tenants/:id/members", formahandler.ListMembers)
		v1.POST("/tenants/:id/members", formahandler.AddMember)
		v1.PATCH("/tenants/:id/members/:principalId", formahandler.PatchMember)
		v1.GET("/tenants/:id/spaces", formahandler.ListSpaces)
		v1.POST("/tenants/:id/spaces", formahandler.BindSpace)
		v1.POST("/bootstrap", formahandler.Bootstrap)
		v1.GET("/assets/counts", formahandler.AssetCounts)
	}
}
