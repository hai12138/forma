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

		v1.POST("/businesses", formahandler.CreateBusiness)
		v1.GET("/businesses", formahandler.ListBusinesses)
		v1.GET("/businesses/:id", formahandler.GetBusiness)
		v1.PATCH("/businesses/:id", formahandler.PatchBusiness)
		v1.POST("/businesses/:id/archive", formahandler.ArchiveBusiness)
		v1.GET("/businesses/:id/model", formahandler.GetBusinessModel)
		v1.PUT("/businesses/:id/model", formahandler.PutBusinessModel)
		v1.GET("/businesses/:id/revisions", formahandler.ListBusinessRevisions)
		v1.GET("/businesses/:id/revisions/:revision", formahandler.GetBusinessRevision)
		v1.GET("/businesses/:id/diff", formahandler.DiffBusiness)
		v1.GET("/businesses/:id/layout", formahandler.GetBusinessLayout)
		v1.PUT("/businesses/:id/layout", formahandler.PutBusinessLayout)

		v1.POST("/businesses/:id/analyst/sessions", formahandler.CreateAnalystSession)
		v1.GET("/businesses/:id/analyst/sessions", formahandler.ListAnalystSessions)
		v1.GET("/businesses/:id/analyst/sessions/:sessionId", formahandler.GetAnalystSession)
		v1.POST("/businesses/:id/analyst/sessions/:sessionId/turns", formahandler.SubmitAnalystTurn)
		v1.GET("/businesses/:id/analyst/sessions/:sessionId/turns", formahandler.ListAnalystTurns)
		v1.GET("/businesses/:id/assertions", formahandler.ListAssertions)
		v1.GET("/businesses/:id/evidence", formahandler.ListBusinessEvidence)
		v1.POST("/businesses/:id/assertions/:assertionId/confirm", formahandler.ConfirmAssertion)
		v1.POST("/businesses/:id/assertions/:assertionId/reject", formahandler.RejectAssertion)
		v1.GET("/businesses/:id/conflicts", formahandler.ListConflicts)
		v1.GET("/businesses/:id/gaps", formahandler.ListGaps)
		v1.POST("/businesses/:id/proposals", formahandler.CreateProposal)
		v1.GET("/businesses/:id/proposals/:proposalId", formahandler.GetProposal)
		v1.POST("/businesses/:id/proposals/:proposalId/apply", formahandler.ApplyProposal)
	}
}
