/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"context"
	"errors"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/datasource/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DataSourceDAO struct{ db *gorm.DB }

func NewDataSourceDAO(db *gorm.DB) *DataSourceDAO { return &DataSourceDAO{db: db} }

type sourceRow struct {
	ID                                                      int64 `gorm:"primaryKey"`
	SourceID, TenantID, Name, SourceType, Status, CreatedBy string
	CreatedAt, UpdatedAt                                    time.Time
}

func (sourceRow) TableName() string { return "forma_data_source" }

type connectionRow struct {
	ID                                                                                                             int64 `gorm:"primaryKey"`
	ConnectionID, TenantID, SourceID, Name, Environment, AdapterType, PublicConfigJSON, CredentialRefID, CreatedBy string
	CreatedAt, UpdatedAt                                                                                           time.Time
}

func (connectionRow) TableName() string { return "forma_data_connection" }

type credentialRow struct {
	ID                                                                 int64 `gorm:"primaryKey"`
	CredentialRefID, TenantID, Provider, SecretType, Status, CreatedBy string
	KeyVersion                                                         int32
	CreatedAt                                                          time.Time
	RotatedAt, RevokedAt                                               *time.Time
}

func (credentialRow) TableName() string { return "forma_data_credential_ref" }

type secretRow struct {
	ID                        int64 `gorm:"primaryKey"`
	CredentialRefID, TenantID string
	Ciphertext                []byte
	KeyVersion                int32
	CreatedAt, UpdatedAt      time.Time
}

func (secretRow) TableName() string { return "forma_data_secret_local" }

type assetRow struct {
	ID                                                                                             int64 `gorm:"primaryKey"`
	AssetID, TenantID, SourceID, ConnectionID, AssetType, Name, PhysicalLocatorJSON, LocatorDigest string
	CreatedAt, UpdatedAt                                                                           time.Time
}

func (assetRow) TableName() string { return "forma_data_asset" }

type snapshotRow struct {
	ID                                                                                        int64 `gorm:"primaryKey"`
	SnapshotID, TenantID, SourceID, ConnectionID, AssetID, SchemaJSON, Fingerprint, CreatedBy string
	CreatedAt                                                                                 time.Time
}

func (snapshotRow) TableName() string { return "forma_data_schema_snapshot" }

func sourceFrom(v *entity.DataSource) *sourceRow {
	return &sourceRow{SourceID: v.SourceID, TenantID: v.TenantID, Name: v.Name, SourceType: string(v.SourceType), Status: string(v.Status), CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func sourceTo(v *sourceRow) *entity.DataSource {
	return &entity.DataSource{SourceID: v.SourceID, TenantID: v.TenantID, Name: v.Name, SourceType: entity.SourceType(v.SourceType), Status: entity.DataSourceStatus(v.Status), CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func connectionFrom(v *entity.DataConnection) *connectionRow {
	return &connectionRow{ConnectionID: v.ConnectionID, TenantID: v.TenantID, SourceID: v.SourceID, Name: v.Name, Environment: string(v.Environment), AdapterType: string(v.AdapterType), PublicConfigJSON: v.PublicConfigJSON, CredentialRefID: v.CredentialRefID, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func connectionTo(v *connectionRow) *entity.DataConnection {
	return &entity.DataConnection{ConnectionID: v.ConnectionID, TenantID: v.TenantID, SourceID: v.SourceID, Name: v.Name, Environment: entity.Environment(v.Environment), AdapterType: entity.AdapterType(v.AdapterType), PublicConfigJSON: v.PublicConfigJSON, CredentialRefID: v.CredentialRefID, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func credentialFrom(v *entity.CredentialRef) *credentialRow {
	return &credentialRow{CredentialRefID: v.CredentialRefID, TenantID: v.TenantID, Provider: string(v.Provider), SecretType: v.SecretType, KeyVersion: v.KeyVersion, Status: string(v.Status), CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, RotatedAt: v.RotatedAt, RevokedAt: v.RevokedAt}
}
func credentialTo(v *credentialRow) *entity.CredentialRef {
	return &entity.CredentialRef{CredentialRefID: v.CredentialRefID, TenantID: v.TenantID, Provider: entity.CredentialProvider(v.Provider), SecretType: v.SecretType, KeyVersion: v.KeyVersion, Status: entity.CredentialStatus(v.Status), CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, RotatedAt: v.RotatedAt, RevokedAt: v.RevokedAt}
}
func assetFrom(v *entity.DataAsset) *assetRow {
	return &assetRow{AssetID: v.AssetID, TenantID: v.TenantID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetType: string(v.AssetType), Name: v.Name, PhysicalLocatorJSON: v.PhysicalLocatorJSON, LocatorDigest: v.LocatorDigest, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func assetTo(v *assetRow) *entity.DataAsset {
	return &entity.DataAsset{AssetID: v.AssetID, TenantID: v.TenantID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetType: entity.AssetType(v.AssetType), Name: v.Name, PhysicalLocatorJSON: v.PhysicalLocatorJSON, LocatorDigest: v.LocatorDigest, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func snapshotFrom(v *entity.SchemaSnapshot) *snapshotRow {
	return &snapshotRow{SnapshotID: v.SnapshotID, TenantID: v.TenantID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetID: v.AssetID, SchemaJSON: v.SchemaJSON, Fingerprint: v.Fingerprint, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt}
}
func snapshotTo(v *snapshotRow) *entity.SchemaSnapshot {
	return &entity.SchemaSnapshot{SnapshotID: v.SnapshotID, TenantID: v.TenantID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetID: v.AssetID, SchemaJSON: v.SchemaJSON, Fingerprint: v.Fingerprint, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt}
}

func (d *DataSourceDAO) CreateSource(c context.Context, v *entity.DataSource) error {
	return d.db.WithContext(c).Create(sourceFrom(v)).Error
}
func (d *DataSourceDAO) GetSource(c context.Context, t, id string) (*entity.DataSource, error) {
	var r sourceRow
	e := d.db.WithContext(c).Where("tenant_id=? AND source_id=?", t, id).First(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDataSourceNotFound
	}
	return sourceTo(&r), e
}
func (d *DataSourceDAO) ListSources(c context.Context, t string) ([]*entity.DataSource, error) {
	var rs []sourceRow
	e := d.db.WithContext(c).Where("tenant_id=?", t).Order("created_at").Find(&rs).Error
	out := make([]*entity.DataSource, 0, len(rs))
	for i := range rs {
		out = append(out, sourceTo(&rs[i]))
	}
	return out, e
}
func (d *DataSourceDAO) UpdateSource(c context.Context, v *entity.DataSource) error {
	r := sourceFrom(v)
	res := d.db.WithContext(c).Model(&sourceRow{}).Where("tenant_id=? AND source_id=?", v.TenantID, v.SourceID).Updates(map[string]any{"name": r.Name, "status": r.Status, "updated_at": r.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return entity.ErrDataSourceNotFound
	}
	return nil
}
func (d *DataSourceDAO) CreateConnection(c context.Context, v *entity.DataConnection) error {
	return d.db.WithContext(c).Create(connectionFrom(v)).Error
}
func (d *DataSourceDAO) GetConnection(c context.Context, t, id string) (*entity.DataConnection, error) {
	var r connectionRow
	e := d.db.WithContext(c).Where("tenant_id=? AND connection_id=?", t, id).First(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDataConnectionNotFound
	}
	return connectionTo(&r), e
}
func (d *DataSourceDAO) ListConnections(c context.Context, t, s string) ([]*entity.DataConnection, error) {
	var rs []connectionRow
	e := d.db.WithContext(c).Where("tenant_id=? AND source_id=?", t, s).Order("created_at").Find(&rs).Error
	out := make([]*entity.DataConnection, 0, len(rs))
	for i := range rs {
		out = append(out, connectionTo(&rs[i]))
	}
	return out, e
}
func (d *DataSourceDAO) UpdateConnection(c context.Context, v *entity.DataConnection) error {
	r := connectionFrom(v)
	res := d.db.WithContext(c).Model(&connectionRow{}).Where("tenant_id=? AND connection_id=?", v.TenantID, v.ConnectionID).Updates(map[string]any{"name": r.Name, "environment": r.Environment, "public_config_json": r.PublicConfigJSON, "credential_ref_id": r.CredentialRefID, "updated_at": r.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return entity.ErrDataConnectionNotFound
	}
	return nil
}
func (d *DataSourceDAO) CreateCredential(c context.Context, v *entity.CredentialRef, s []byte) error {
	return d.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(credentialFrom(v)).Error; e != nil {
			return e
		}
		return tx.Create(&secretRow{CredentialRefID: v.CredentialRefID, TenantID: v.TenantID, Ciphertext: s, KeyVersion: v.KeyVersion, CreatedAt: v.CreatedAt, UpdatedAt: v.CreatedAt}).Error
	})
}
func (d *DataSourceDAO) GetCredential(c context.Context, t, id string) (*entity.CredentialRef, error) {
	var r credentialRow
	e := d.db.WithContext(c).Where("tenant_id=? AND credential_ref_id=?", t, id).First(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDataCredentialNotFound
	}
	return credentialTo(&r), e
}
func (d *DataSourceDAO) UpdateCredential(c context.Context, v *entity.CredentialRef, s []byte) error {
	return d.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&credentialRow{}).Where("tenant_id=? AND credential_ref_id=?", v.TenantID, v.CredentialRefID).Updates(credentialFrom(v))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return entity.ErrDataCredentialNotFound
		}
		if s != nil {
			return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tenant_id"}, {Name: "credential_ref_id"}}, DoUpdates: clause.AssignmentColumns([]string{"ciphertext", "key_version", "updated_at"})}).Create(&secretRow{CredentialRefID: v.CredentialRefID, TenantID: v.TenantID, Ciphertext: s, KeyVersion: v.KeyVersion, UpdatedAt: time.Now().UTC()}).Error
		}
		return nil
	})
}
func (d *DataSourceDAO) GetSecret(c context.Context, t, id string) ([]byte, error) {
	var r secretRow
	e := d.db.WithContext(c).Where("tenant_id=? AND credential_ref_id=?", t, id).First(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDataCredentialNotFound
	}
	return append([]byte(nil), r.Ciphertext...), e
}
func (d *DataSourceDAO) DeleteSecret(c context.Context, t, id string) error {
	return d.db.WithContext(c).Where("tenant_id=? AND credential_ref_id=?", t, id).Delete(&secretRow{}).Error
}
func (d *DataSourceDAO) UpsertAsset(c context.Context, v *entity.DataAsset) (*entity.DataAsset, bool, error) {
	var r assetRow
	e := d.db.WithContext(c).Where("tenant_id=? AND source_id=? AND connection_id=? AND locator_digest=?", v.TenantID, v.SourceID, v.ConnectionID, v.LocatorDigest).First(&r).Error
	if e == nil {
		return assetTo(&r), false, nil
	}
	if !errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, false, e
	}
	e = d.db.WithContext(c).Create(assetFrom(v)).Error
	if e != nil {
		var x assetRow
		if xerr := d.db.WithContext(c).Where("tenant_id=? AND source_id=? AND connection_id=? AND locator_digest=?", v.TenantID, v.SourceID, v.ConnectionID, v.LocatorDigest).First(&x).Error; xerr == nil {
			return assetTo(&x), false, nil
		}
		return nil, false, e
	}
	return v, true, nil
}
func (d *DataSourceDAO) GetAsset(c context.Context, t, id string) (*entity.DataAsset, error) {
	var r assetRow
	e := d.db.WithContext(c).Where("tenant_id=? AND asset_id=?", t, id).First(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDataAssetNotFound
	}
	return assetTo(&r), e
}
func (d *DataSourceDAO) ListAssets(c context.Context, t, s string) ([]*entity.DataAsset, error) {
	var rs []assetRow
	e := d.db.WithContext(c).Where("tenant_id=? AND source_id=?", t, s).Order("created_at").Find(&rs).Error
	out := make([]*entity.DataAsset, 0, len(rs))
	for i := range rs {
		out = append(out, assetTo(&rs[i]))
	}
	return out, e
}
func (d *DataSourceDAO) CreateSnapshot(c context.Context, v *entity.SchemaSnapshot) error {
	return d.db.WithContext(c).Create(snapshotFrom(v)).Error
}
func (d *DataSourceDAO) GetSnapshot(c context.Context, t, id string) (*entity.SchemaSnapshot, error) {
	var r snapshotRow
	e := d.db.WithContext(c).Where("tenant_id=? AND snapshot_id=?", t, id).First(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDataSchemaSnapshotNotFound
	}
	return snapshotTo(&r), e
}
