/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/crossdomain/forma/integration"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/repository"
	assetsvc "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/service"
	analystrepo "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
	analystsvc "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/service"
	businessrepo "github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/meta"
	tenancyrepo "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/repository"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

type ServiceComponents struct {
	DB    *gorm.DB
	IDGen idgen.IDGenerator
}

type ApplicationService struct {
	DB          *gorm.DB
	IDGen       idgen.IDGenerator
	DomainSVC   assetsvc.AssetRegistry
	TenancySVC  tenancysvc.TenancyService
	BusinessSVC businesssvc.BusinessService
	AnalystSVC  analystsvc.AnalystService
}

var ApplicationSVC = &ApplicationService{}

func InitService(_ context.Context, components *ServiceComponents) *ApplicationService {
	ApplicationSVC.DB = components.DB
	ApplicationSVC.IDGen = components.IDGen
	ApplicationSVC.DomainSVC = assetsvc.NewAssetRegistry(&assetsvc.Components{
		AssetRepo: repository.NewAssetRefRepository(components.DB),
		CozeRepo:  repository.NewCozeResourceRefRepository(components.DB),
		IDGen:     components.IDGen,
	})
	ApplicationSVC.TenancySVC = tenancysvc.NewTenancyService(&tenancysvc.Components{
		PrincipalRepo:  tenancyrepo.NewPrincipalRepository(components.DB),
		TenantRepo:     tenancyrepo.NewTenantRepository(components.DB),
		MembershipRepo: tenancyrepo.NewMembershipRepository(components.DB),
		SpaceRefRepo:   tenancyrepo.NewSpaceRefRepository(components.DB),
		AuditRepo:      tenancyrepo.NewAuditRepository(components.DB),
	})
	ApplicationSVC.BusinessSVC = businesssvc.NewBusinessService(&businesssvc.Components{
		Repo: businessrepo.NewBusinessRepository(components.DB),
	})
	auditHook := NewTenancyAuditHook(ApplicationSVC.TenancySVC)
	ApplicationSVC.AnalystSVC = analystsvc.NewAnalystService(&analystsvc.Components{
		Repo:        analystrepo.NewAnalystRepository(components.DB),
		BusinessSVC: ApplicationSVC.BusinessSVC,
		Model:       integration.NewCozeEinoAnalystModel("FORMA_ANALYST"),
		AuditHook:   auditHook,
	})
	return ApplicationSVC
}

type HealthResponse struct {
	Status string `json:"status"`
}

type VersionResponse struct {
	FormaVersion       string `json:"forma_version"`
	FormaSchemaVersion string `json:"forma_schema_version"`
}

type BaselineResponse struct {
	FormaVersion            string `json:"forma_version"`
	FormaSchemaVersion      string `json:"forma_schema_version"`
	FormaBaselineTag        string `json:"forma_baseline_tag"`
	CozeBaselineCommit      string `json:"coze_baseline_commit"`
	WorkspaceBaselineCommit string `json:"workspace_baseline_commit"`
	RuntimeFoundation       string `json:"runtime_foundation"`
}

func (s *ApplicationService) Health(_ context.Context) *HealthResponse {
	return &HealthResponse{Status: "ok"}
}

func (s *ApplicationService) Version(_ context.Context) *VersionResponse {
	return &VersionResponse{
		FormaVersion:       meta.FormaVersion,
		FormaSchemaVersion: meta.FormaSchemaVersion,
	}
}

func (s *ApplicationService) Baseline(_ context.Context) *BaselineResponse {
	return &BaselineResponse{
		FormaVersion:            meta.FormaVersion,
		FormaSchemaVersion:      meta.FormaSchemaVersion,
		FormaBaselineTag:        meta.FormaBaselineTag,
		CozeBaselineCommit:      meta.CozeBaselineCommit,
		WorkspaceBaselineCommit: meta.WorkspaceBaselineCommit,
		RuntimeFoundation:       meta.RuntimeFoundation,
	}
}
