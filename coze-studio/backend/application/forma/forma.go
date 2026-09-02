/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/crossdomain/forma/integration"
	analystrepo "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
	analystsvc "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/repository"
	assetsvc "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/service"
	businessrepo "github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	datarepo "github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
	datasourcerepo "github.com/coze-dev/coze-studio/backend/domain/forma/datasource/repository"
	datasourcesvc "github.com/coze-dev/coze-studio/backend/domain/forma/datasource/service"
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
	DB            *gorm.DB
	IDGen         idgen.IDGenerator
	DomainSVC     assetsvc.AssetRegistry
	TenancySVC    tenancysvc.TenancyService
	BusinessSVC   businesssvc.BusinessService
	AnalystSVC    analystsvc.AnalystService
	DataSVC       datasvc.DataService
	DatasourceSVC datasourcesvc.DataSourceService
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
		Repo:         analystrepo.NewAnalystRepository(components.DB),
		BusinessSVC:  ApplicationSVC.BusinessSVC,
		BusinessRepo: businessrepo.NewBusinessRepository(components.DB),
		DB:           components.DB,
		Model:        integration.NewCozeEinoAnalystModel("FORMA_ANALYST"),
		AuditHook:    auditHook,
	})
	ApplicationSVC.DataSVC = datasvc.NewDataService(&datasvc.Components{
		Repo:        datarepo.NewDataRepository(components.DB),
		BusinessSVC: ApplicationSVC.BusinessSVC,
		Model:       integration.NewCozeEinoDataModel("FORMA_DATA"),
	})
	secretProvider, _ := datasourcesvc.NewLocalSecretProviderFromEnv()
	ApplicationSVC.DatasourceSVC = datasourcesvc.NewDataSourceService(&datasourcesvc.Components{
		Repo:     datasourcerepo.NewDataSourceRepository(components.DB),
		Secrets:  secretProvider,
		Adapters: datasourcesvc.NewDefaultAdapterRegistry(),
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
