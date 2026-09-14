/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"sync"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

// MemoryAssetProjection is an in-memory AssetProjection for tests.
type MemoryAssetProjection struct {
	mu     sync.Mutex
	assets map[string]*assetentity.AssetRef
}

func NewMemoryAssetProjection() *MemoryAssetProjection {
	return &MemoryAssetProjection{assets: make(map[string]*assetentity.AssetRef)}
}

func assetKey(tenantID, assetID string) string { return tenantID + "\x00" + assetID }

func (m *MemoryAssetProjection) CreateCapabilityAsset(_ context.Context, asset *assetentity.AssetRef) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := assetKey(asset.TenantID, asset.AssetID)
	if _, ok := m.assets[k]; ok {
		return entity.ErrConflict
	}
	cp := *asset
	if cp.Kind == "" {
		cp.Kind = assetentity.AssetKindCapability
	}
	m.assets[k] = &cp
	return nil
}

func (m *MemoryAssetProjection) UpdateCapabilityProjection(_ context.Context, tenantID, assetID, name, semanticVersion, contentDigest string, status assetentity.AssetStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.assets[assetKey(tenantID, assetID)]
	if !ok {
		return entity.ErrNotFound
	}
	a.Name = name
	a.SemanticVersion = semanticVersion
	a.ContentDigest = contentDigest
	a.Status = status
	// Revision and SchemaVersion must remain unchanged.
	return nil
}

func (m *MemoryAssetProjection) GetCapabilityAsset(_ context.Context, tenantID, assetID string) (*assetentity.AssetRef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.assets[assetKey(tenantID, assetID)]
	if !ok {
		return nil, entity.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *MemoryAssetProjection) ListCapabilityAssetsByTenant(_ context.Context, tenantID string) ([]*assetentity.AssetRef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*assetentity.AssetRef, 0)
	for _, a := range m.assets {
		if a == nil || a.TenantID != tenantID || a.Kind != assetentity.AssetKindCapability {
			continue
		}
		cp := *a
		out = append(out, &cp)
	}
	return out, nil
}

type assetSnap map[string]*assetentity.AssetRef

func (m *MemoryAssetProjection) snapshot() assetSnap {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(assetSnap, len(m.assets))
	for k, v := range m.assets {
		cp := *v
		out[k] = &cp
	}
	return out
}

func (m *MemoryAssetProjection) restore(s assetSnap) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.assets = make(map[string]*assetentity.AssetRef, len(s))
	for k, v := range s {
		cp := *v
		m.assets[k] = &cp
	}
}

var _ AssetProjection = (*MemoryAssetProjection)(nil)
