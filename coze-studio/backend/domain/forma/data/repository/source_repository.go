/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/internal/dal"
	"gorm.io/gorm"
)

type DataSourceRepository interface {
	CreateSource(context.Context, *entity.DataSource) error
	GetSource(context.Context, string, string) (*entity.DataSource, error)
	ListSources(context.Context, string) ([]*entity.DataSource, error)
	UpdateSource(context.Context, *entity.DataSource) error

	CreateConnection(context.Context, *entity.DataConnection) error
	GetConnection(context.Context, string, string) (*entity.DataConnection, error)
	ListConnections(context.Context, string, string) ([]*entity.DataConnection, error)
	UpdateConnection(context.Context, *entity.DataConnection) error

	CreateCredential(context.Context, *entity.CredentialRef, []byte) error
	GetCredential(context.Context, string, string) (*entity.CredentialRef, error)
	UpdateCredential(context.Context, *entity.CredentialRef, []byte) error
	GetSecret(context.Context, string, string) ([]byte, error)
	DeleteSecret(context.Context, string, string) error

	UpsertAsset(context.Context, *entity.DataAsset) (*entity.DataAsset, bool, error)
	GetAsset(context.Context, string, string) (*entity.DataAsset, error)
	ListAssets(context.Context, string, string) ([]*entity.DataAsset, error)

	CreateSnapshot(context.Context, *entity.SchemaSnapshot) error
	GetSnapshot(context.Context, string, string) (*entity.SchemaSnapshot, error)
	ListSnapshotsByAsset(ctx context.Context, tenantID, sourceID, connectionID, assetID string) ([]*entity.SchemaSnapshot, error)
}

type gormRepository struct{ dao *dal.DataSourceDAO }

func NewDataSourceRepository(db *gorm.DB) DataSourceRepository {
	return &gormRepository{dao: dal.NewDataSourceDAO(db)}
}

func (r *gormRepository) CreateSource(c context.Context, v *entity.DataSource) error {
	return r.dao.CreateSource(c, v)
}
func (r *gormRepository) GetSource(c context.Context, t, id string) (*entity.DataSource, error) {
	return r.dao.GetSource(c, t, id)
}
func (r *gormRepository) ListSources(c context.Context, t string) ([]*entity.DataSource, error) {
	return r.dao.ListSources(c, t)
}
func (r *gormRepository) UpdateSource(c context.Context, v *entity.DataSource) error {
	return r.dao.UpdateSource(c, v)
}
func (r *gormRepository) CreateConnection(c context.Context, v *entity.DataConnection) error {
	return r.dao.CreateConnection(c, v)
}
func (r *gormRepository) GetConnection(c context.Context, t, id string) (*entity.DataConnection, error) {
	return r.dao.GetConnection(c, t, id)
}
func (r *gormRepository) ListConnections(c context.Context, t, s string) ([]*entity.DataConnection, error) {
	return r.dao.ListConnections(c, t, s)
}
func (r *gormRepository) UpdateConnection(c context.Context, v *entity.DataConnection) error {
	return r.dao.UpdateConnection(c, v)
}
func (r *gormRepository) CreateCredential(c context.Context, v *entity.CredentialRef, secret []byte) error {
	return r.dao.CreateCredential(c, v, secret)
}
func (r *gormRepository) GetCredential(c context.Context, t, id string) (*entity.CredentialRef, error) {
	return r.dao.GetCredential(c, t, id)
}
func (r *gormRepository) UpdateCredential(c context.Context, v *entity.CredentialRef, secret []byte) error {
	return r.dao.UpdateCredential(c, v, secret)
}
func (r *gormRepository) GetSecret(c context.Context, t, id string) ([]byte, error) {
	return r.dao.GetSecret(c, t, id)
}
func (r *gormRepository) DeleteSecret(c context.Context, t, id string) error {
	return r.dao.DeleteSecret(c, t, id)
}
func (r *gormRepository) UpsertAsset(c context.Context, v *entity.DataAsset) (*entity.DataAsset, bool, error) {
	return r.dao.UpsertAsset(c, v)
}
func (r *gormRepository) GetAsset(c context.Context, t, id string) (*entity.DataAsset, error) {
	return r.dao.GetAsset(c, t, id)
}
func (r *gormRepository) ListAssets(c context.Context, t, s string) ([]*entity.DataAsset, error) {
	return r.dao.ListAssets(c, t, s)
}
func (r *gormRepository) CreateSnapshot(c context.Context, v *entity.SchemaSnapshot) error {
	return r.dao.CreateSnapshot(c, v)
}
func (r *gormRepository) GetSnapshot(c context.Context, t, id string) (*entity.SchemaSnapshot, error) {
	return r.dao.GetSnapshot(c, t, id)
}
func (r *gormRepository) ListSnapshotsByAsset(c context.Context, tenantID, sourceID, connectionID, assetID string) ([]*entity.SchemaSnapshot, error) {
	return r.dao.ListSnapshotsByAsset(c, tenantID, sourceID, connectionID, assetID)
}
