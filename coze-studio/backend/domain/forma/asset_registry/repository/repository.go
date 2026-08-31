/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/internal/dal"
)

type AssetRefRepository interface {
	Create(ctx context.Context, asset *entity.AssetRef) error
	GetByTenantAssetRevision(ctx context.Context, tenantID, assetID string, revision int32) (*entity.AssetRef, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*entity.AssetRef, error)
}

type CozeResourceRefRepository interface {
	Create(ctx context.Context, ref *entity.CozeResourceRef) error
	ListByAsset(ctx context.Context, tenantID, assetID string, revision int32) ([]*entity.CozeResourceRef, error)
}

func NewAssetRefRepository(db *gorm.DB) AssetRefRepository {
	return dal.NewAssetRefDAO(db)
}

func NewCozeResourceRefRepository(db *gorm.DB) CozeResourceRefRepository {
	return dal.NewCozeResourceRefDAO(db)
}
