/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"

	"gorm.io/gorm"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	assetrepo "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/repository"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

// GormAssetProjection adapts Asset Registry DAO onto AssetProjection for shared-tx UoW.
type GormAssetProjection struct {
	repo assetrepo.AssetRefRepository
}

func NewGormAssetProjection(db *gorm.DB) *GormAssetProjection {
	return &GormAssetProjection{repo: assetrepo.NewAssetRefRepository(db)}
}

func (g *GormAssetProjection) CreateCapabilityAsset(ctx context.Context, asset *assetentity.AssetRef) error {
	if asset == nil {
		return entity.ErrInvalidPayload
	}
	if asset.Kind == "" {
		asset.Kind = assetentity.AssetKindCapability
	}
	if asset.Revision == 0 {
		asset.Revision = 1
	}
	if asset.SchemaVersion == "" {
		asset.SchemaVersion = "1.0"
	}
	return g.repo.Create(ctx, asset)
}

func (g *GormAssetProjection) UpdateCapabilityProjection(ctx context.Context, tenantID, assetID, name, semanticVersion, contentDigest string, status assetentity.AssetStatus) error {
	updated, err := g.repo.UpdateCapabilityProjection(ctx, tenantID, assetID, name, semanticVersion, contentDigest, status)
	if err != nil {
		return err
	}
	if updated == nil {
		return entity.ErrNotFound
	}
	return nil
}

func (g *GormAssetProjection) GetCapabilityAsset(ctx context.Context, tenantID, assetID string) (*assetentity.AssetRef, error) {
	a, err := g.repo.GetByTenantAssetRevision(ctx, tenantID, assetID, 1)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, entity.ErrNotFound
	}
	return a, nil
}

var _ AssetProjection = (*GormAssetProjection)(nil)
