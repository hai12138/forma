/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
)

type AssetRefDAO struct {
	db *gorm.DB
}

func NewAssetRefDAO(db *gorm.DB) *AssetRefDAO {
	return &AssetRefDAO{db: db}
}

func (d *AssetRefDAO) Create(ctx context.Context, asset *entity.AssetRef) error {
	model := toAssetModel(asset)
	return d.db.WithContext(ctx).Create(model).Error
}

func (d *AssetRefDAO) GetByTenantAssetRevision(ctx context.Context, tenantID, assetID string, revision int32) (*entity.AssetRef, error) {
	var model AssetRefModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND asset_id = ? AND revision = ? AND deleted_at IS NULL", tenantID, assetID, revision).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAssetEntity(&model), nil
}

func (d *AssetRefDAO) ListByTenant(ctx context.Context, tenantID string) ([]*entity.AssetRef, error) {
	var models []AssetRefModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("updated_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.AssetRef, 0, len(models))
	for i := range models {
		out = append(out, toAssetEntity(&models[i]))
	}
	return out, nil
}

func toAssetModel(a *entity.AssetRef) *AssetRefModel {
	var digest *string
	if a.ContentDigest != "" {
		digest = &a.ContentDigest
	}
	return &AssetRefModel{
		ID:              a.ID,
		TenantID:        a.TenantID,
		AssetID:         a.AssetID,
		Kind:            string(a.Kind),
		Name:            a.Name,
		SemanticVersion: a.SemanticVersion,
		Revision:        a.Revision,
		SchemaVersion:   a.SchemaVersion,
		Status:          string(a.Status),
		OwnerID:         a.OwnerID,
		CreatedBy:       a.CreatedBy,
		ContentDigest:   digest,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
		DeletedAt:       a.DeletedAt,
	}
}

func toAssetEntity(m *AssetRefModel) *entity.AssetRef {
	digest := ""
	if m.ContentDigest != nil {
		digest = *m.ContentDigest
	}
	return &entity.AssetRef{
		ID:              m.ID,
		TenantID:        m.TenantID,
		AssetID:         m.AssetID,
		Kind:            entity.AssetKind(m.Kind),
		Name:            m.Name,
		SemanticVersion: m.SemanticVersion,
		Revision:        m.Revision,
		SchemaVersion:   m.SchemaVersion,
		Status:          entity.AssetStatus(m.Status),
		OwnerID:         m.OwnerID,
		CreatedBy:       m.CreatedBy,
		ContentDigest:   digest,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		DeletedAt:       m.DeletedAt,
	}
}

type CozeResourceRefDAO struct {
	db *gorm.DB
}

func NewCozeResourceRefDAO(db *gorm.DB) *CozeResourceRefDAO {
	return &CozeResourceRefDAO{db: db}
}

func (d *CozeResourceRefDAO) Create(ctx context.Context, ref *entity.CozeResourceRef) error {
	model := toCozeResourceModel(ref)
	return d.db.WithContext(ctx).Create(model).Error
}

func (d *CozeResourceRefDAO) ListByAsset(ctx context.Context, tenantID, assetID string, revision int32) ([]*entity.CozeResourceRef, error) {
	var models []CozeResourceRefModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND asset_id = ? AND asset_revision = ?", tenantID, assetID, revision).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.CozeResourceRef, 0, len(models))
	for i := range models {
		out = append(out, toCozeResourceEntity(&models[i]))
	}
	return out, nil
}

func toCozeResourceModel(r *entity.CozeResourceRef) *CozeResourceRefModel {
	var version *string
	if r.CozeVersion != "" {
		version = &r.CozeVersion
	}
	return &CozeResourceRefModel{
		ID:               r.ID,
		TenantID:         r.TenantID,
		AssetID:          r.AssetID,
		AssetRevision:    r.AssetRevision,
		CozeResourceType: string(r.CozeResourceType),
		CozeResourceID:   r.CozeResourceID,
		CozeSpaceID:      r.CozeSpaceID,
		CozeVersion:      version,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func toCozeResourceEntity(m *CozeResourceRefModel) *entity.CozeResourceRef {
	version := ""
	if m.CozeVersion != nil {
		version = *m.CozeVersion
	}
	return &entity.CozeResourceRef{
		ID:               m.ID,
		TenantID:         m.TenantID,
		AssetID:          m.AssetID,
		AssetRevision:    m.AssetRevision,
		CozeResourceType: entity.CozeResourceType(m.CozeResourceType),
		CozeResourceID:   m.CozeResourceID,
		CozeSpaceID:      m.CozeSpaceID,
		CozeVersion:      version,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func NowUTC() time.Time { return time.Now().UTC() }
