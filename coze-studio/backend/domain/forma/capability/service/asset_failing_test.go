/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"errors"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
)

// FailingAssetProjection wraps an AssetProjection and can fail Create/Update for rollback tests.
type FailingAssetProjection struct {
	Inner      AssetProjection
	FailCreate bool
	FailUpdate bool
	CreateErr  error
	UpdateErr  error
}

func (f *FailingAssetProjection) CreateCapabilityAsset(ctx context.Context, asset *assetentity.AssetRef) error {
	if f.FailCreate {
		if f.CreateErr != nil {
			return f.CreateErr
		}
		return errors.New("injected asset create failure")
	}
	return f.Inner.CreateCapabilityAsset(ctx, asset)
}

func (f *FailingAssetProjection) UpdateCapabilityProjection(ctx context.Context, tenantID, assetID, name, semanticVersion, contentDigest string, status assetentity.AssetStatus) error {
	if f.FailUpdate {
		if f.UpdateErr != nil {
			return f.UpdateErr
		}
		return errors.New("injected asset update failure")
	}
	return f.Inner.UpdateCapabilityProjection(ctx, tenantID, assetID, name, semanticVersion, contentDigest, status)
}

func (f *FailingAssetProjection) GetCapabilityAsset(ctx context.Context, tenantID, assetID string) (*assetentity.AssetRef, error) {
	return f.Inner.GetCapabilityAsset(ctx, tenantID, assetID)
}

var _ AssetProjection = (*FailingAssetProjection)(nil)
