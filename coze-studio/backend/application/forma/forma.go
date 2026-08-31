/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/repository"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/meta"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

type ServiceComponents struct {
	DB    *gorm.DB
	IDGen idgen.IDGenerator
}

type ApplicationService struct {
	DomainSVC service.AssetRegistry
}

var ApplicationSVC = &ApplicationService{}

func InitService(_ context.Context, components *ServiceComponents) *ApplicationService {
	ApplicationSVC.DomainSVC = service.NewAssetRegistry(&service.Components{
		AssetRepo: repository.NewAssetRefRepository(components.DB),
		CozeRepo:  repository.NewCozeResourceRefRepository(components.DB),
		IDGen:     components.IDGen,
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
	FormaVersion           string `json:"forma_version"`
	FormaSchemaVersion     string `json:"forma_schema_version"`
	FormaBaselineTag       string `json:"forma_baseline_tag"`
	CozeBaselineCommit     string `json:"coze_baseline_commit"`
	WorkspaceBaselineCommit string `json:"workspace_baseline_commit"`
	RuntimeFoundation      string `json:"runtime_foundation"`
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
