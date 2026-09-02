/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/coze-dev/coze-studio/backend/domain/forma/datasource/entity"
)

type memoryRepository struct {
	mu          sync.RWMutex
	sources     map[string]*entity.DataSource
	connections map[string]*entity.DataConnection
	credentials map[string]*entity.CredentialRef
	secrets     map[string][]byte
	assets      map[string]*entity.DataAsset
	assetKeys   map[string]string
	snapshots   map[string]*entity.SchemaSnapshot
}

func NewMemoryDataSourceRepository() DataSourceRepository {
	return &memoryRepository{
		sources: map[string]*entity.DataSource{}, connections: map[string]*entity.DataConnection{},
		credentials: map[string]*entity.CredentialRef{}, secrets: map[string][]byte{},
		assets: map[string]*entity.DataAsset{}, assetKeys: map[string]string{}, snapshots: map[string]*entity.SchemaSnapshot{},
	}
}

func key(t, id string) string { return t + "\x00" + id }
func cloneSource(v *entity.DataSource) *entity.DataSource {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}
func cloneConnection(v *entity.DataConnection) *entity.DataConnection {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}
func cloneCredential(v *entity.CredentialRef) *entity.CredentialRef {
	if v == nil {
		return nil
	}
	x := *v
	if v.RotatedAt != nil {
		y := *v.RotatedAt
		x.RotatedAt = &y
	}
	if v.RevokedAt != nil {
		y := *v.RevokedAt
		x.RevokedAt = &y
	}
	return &x
}
func cloneAsset(v *entity.DataAsset) *entity.DataAsset {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}
func cloneSnapshot(v *entity.SchemaSnapshot) *entity.SchemaSnapshot {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}

func (r *memoryRepository) CreateSource(_ context.Context, v *entity.DataSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[key(v.TenantID, v.SourceID)] = cloneSource(v)
	return nil
}
func (r *memoryRepository) GetSource(_ context.Context, t, id string) (*entity.DataSource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.sources[key(t, id)]
	if !ok {
		return nil, entity.ErrDataSourceNotFound
	}
	return cloneSource(v), nil
}
func (r *memoryRepository) ListSources(_ context.Context, t string) ([]*entity.DataSource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []*entity.DataSource{}
	for _, v := range r.sources {
		if v.TenantID == t {
			out = append(out, cloneSource(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (r *memoryRepository) UpdateSource(_ context.Context, v *entity.DataSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(v.TenantID, v.SourceID)
	if _, ok := r.sources[k]; !ok {
		return entity.ErrDataSourceNotFound
	}
	r.sources[k] = cloneSource(v)
	return nil
}
func (r *memoryRepository) CreateConnection(_ context.Context, v *entity.DataConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[key(v.TenantID, v.ConnectionID)] = cloneConnection(v)
	return nil
}
func (r *memoryRepository) GetConnection(_ context.Context, t, id string) (*entity.DataConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.connections[key(t, id)]
	if !ok {
		return nil, entity.ErrDataConnectionNotFound
	}
	return cloneConnection(v), nil
}
func (r *memoryRepository) ListConnections(_ context.Context, t, s string) ([]*entity.DataConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []*entity.DataConnection{}
	for _, v := range r.connections {
		if v.TenantID == t && v.SourceID == s {
			out = append(out, cloneConnection(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (r *memoryRepository) UpdateConnection(_ context.Context, v *entity.DataConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(v.TenantID, v.ConnectionID)
	if _, ok := r.connections[k]; !ok {
		return entity.ErrDataConnectionNotFound
	}
	r.connections[k] = cloneConnection(v)
	return nil
}
func (r *memoryRepository) CreateCredential(_ context.Context, v *entity.CredentialRef, s []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(v.TenantID, v.CredentialRefID)
	r.credentials[k] = cloneCredential(v)
	r.secrets[k] = append([]byte(nil), s...)
	return nil
}
func (r *memoryRepository) GetCredential(_ context.Context, t, id string) (*entity.CredentialRef, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.credentials[key(t, id)]
	if !ok {
		return nil, entity.ErrDataCredentialNotFound
	}
	return cloneCredential(v), nil
}
func (r *memoryRepository) UpdateCredential(_ context.Context, v *entity.CredentialRef, s []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(v.TenantID, v.CredentialRefID)
	if _, ok := r.credentials[k]; !ok {
		return entity.ErrDataCredentialNotFound
	}
	r.credentials[k] = cloneCredential(v)
	if s != nil {
		r.secrets[k] = append([]byte(nil), s...)
	}
	return nil
}
func (r *memoryRepository) GetSecret(_ context.Context, t, id string) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.secrets[key(t, id)]
	if !ok {
		return nil, entity.ErrDataCredentialNotFound
	}
	return append([]byte(nil), v...), nil
}
func (r *memoryRepository) DeleteSecret(_ context.Context, t, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.secrets, key(t, id))
	return nil
}
func (r *memoryRepository) UpsertAsset(_ context.Context, v *entity.DataAsset) (*entity.DataAsset, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ik := fmt.Sprintf("%s\x00%s\x00%s\x00%s", v.TenantID, v.SourceID, v.ConnectionID, v.LocatorDigest)
	if id, ok := r.assetKeys[ik]; ok {
		return cloneAsset(r.assets[key(v.TenantID, id)]), false, nil
	}
	r.assets[key(v.TenantID, v.AssetID)] = cloneAsset(v)
	r.assetKeys[ik] = v.AssetID
	return cloneAsset(v), true, nil
}
func (r *memoryRepository) GetAsset(_ context.Context, t, id string) (*entity.DataAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.assets[key(t, id)]
	if !ok {
		return nil, entity.ErrDataAssetNotFound
	}
	return cloneAsset(v), nil
}
func (r *memoryRepository) ListAssets(_ context.Context, t, s string) ([]*entity.DataAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []*entity.DataAsset{}
	for _, v := range r.assets {
		if v.TenantID == t && v.SourceID == s {
			out = append(out, cloneAsset(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (r *memoryRepository) CreateSnapshot(_ context.Context, v *entity.SchemaSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(v.TenantID, v.SnapshotID)
	if _, ok := r.snapshots[k]; ok {
		return fmt.Errorf("duplicate snapshot")
	}
	r.snapshots[k] = cloneSnapshot(v)
	return nil
}
func (r *memoryRepository) GetSnapshot(_ context.Context, t, id string) (*entity.SchemaSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.snapshots[key(t, id)]
	if !ok {
		return nil, entity.ErrDataSchemaSnapshotNotFound
	}
	return cloneSnapshot(v), nil
}
