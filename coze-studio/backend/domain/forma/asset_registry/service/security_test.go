/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/service"
)

type memAssetRepo struct {
	mu   sync.Mutex
	rows []*entity.AssetRef
	seq  int64
}

func (r *memAssetRepo) Create(_ context.Context, asset *entity.AssetRef) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	cp := *asset
	cp.ID = r.seq
	r.rows = append(r.rows, &cp)
	asset.ID = cp.ID
	return nil
}

func (r *memAssetRepo) GetByTenantAssetRevision(_ context.Context, tenantID, assetID string, revision int32) (*entity.AssetRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.rows {
		if a.TenantID == tenantID && a.AssetID == assetID && a.Revision == revision && a.DeletedAt == nil {
			cp := *a
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memAssetRepo) ListByTenant(_ context.Context, tenantID string) ([]*entity.AssetRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*entity.AssetRef, 0)
	for _, a := range r.rows {
		if a.TenantID == tenantID && a.DeletedAt == nil {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memAssetRepo) UpdateName(_ context.Context, tenantID, assetID string, revision int32, name string) (*entity.AssetRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.rows {
		if a.TenantID == tenantID && a.AssetID == assetID && a.Revision == revision && a.DeletedAt == nil {
			a.Name = name
			a.UpdatedAt = time.Now().UTC()
			cp := *a
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memAssetRepo) Archive(_ context.Context, tenantID, assetID string, revision int32) (*entity.AssetRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	for _, a := range r.rows {
		if a.TenantID == tenantID && a.AssetID == assetID && a.Revision == revision && a.DeletedAt == nil {
			a.Status = entity.AssetStatusArchived
			a.DeletedAt = &now
			a.UpdatedAt = now
			cp := *a
			return &cp, nil
		}
	}
	return nil, nil
}

type memCozeRepo struct{}

func (r *memCozeRepo) Create(_ context.Context, _ *entity.CozeResourceRef) error { return nil }
func (r *memCozeRepo) ListByAsset(_ context.Context, _, _ string, _ int32) ([]*entity.CozeResourceRef, error) {
	return nil, nil
}

func TestAssetRegistry_TenantIsolation_Get(t *testing.T) {
	repo := &memAssetRepo{}
	reg := service.NewAssetRegistry(&service.Components{
		AssetRepo: repo,
		CozeRepo:  &memCozeRepo{},
	})
	ctx := context.Background()

	asset, err := reg.CreateAsset(ctx, &service.CreateAssetRequest{
		TenantID:  "tenant-a",
		AssetID:   "asset-1",
		Kind:      entity.AssetKindBusiness,
		Name:      "A only",
		OwnerID:   1,
		CreatedBy: 1,
	})
	require.NoError(t, err)

	got, err := reg.GetAsset(ctx, "tenant-a", asset.AssetID, asset.Revision)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "tenant-a", got.TenantID)

	// Forged tenant header simulation: Tenant B must not see Tenant A asset.
	cross, err := reg.GetAsset(ctx, "tenant-b", asset.AssetID, asset.Revision)
	require.NoError(t, err)
	assert.Nil(t, cross)
}

func TestAssetRegistry_UpdateAndArchiveRequireTenant(t *testing.T) {
	repo := &memAssetRepo{}
	reg := service.NewAssetRegistry(&service.Components{
		AssetRepo: repo,
		CozeRepo:  &memCozeRepo{},
	})
	ctx := context.Background()

	asset, err := reg.CreateAsset(ctx, &service.CreateAssetRequest{
		TenantID:  "tenant-a",
		AssetID:   "asset-2",
		Kind:      entity.AssetKindCapability,
		Name:      "Cap",
		OwnerID:   1,
		CreatedBy: 1,
	})
	require.NoError(t, err)

	_, err = reg.UpdateAssetName(ctx, "", asset.AssetID, asset.Revision, "x")
	require.Error(t, err)

	updated, err := reg.UpdateAssetName(ctx, "tenant-a", asset.AssetID, asset.Revision, "Cap2")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Cap2", updated.Name)

	// Wrong tenant cannot update.
	miss, err := reg.UpdateAssetName(ctx, "tenant-b", asset.AssetID, asset.Revision, "hack")
	require.NoError(t, err)
	assert.Nil(t, miss)

	archived, err := reg.ArchiveAsset(ctx, "tenant-a", asset.AssetID, asset.Revision)
	require.NoError(t, err)
	require.NotNil(t, archived)
	assert.Equal(t, entity.AssetStatusArchived, archived.Status)

	// After archive, tenant-scoped get (deleted_at set) returns nil.
	got, err := reg.GetAsset(ctx, "tenant-a", asset.AssetID, asset.Revision)
	require.NoError(t, err)
	assert.Nil(t, got)
}
