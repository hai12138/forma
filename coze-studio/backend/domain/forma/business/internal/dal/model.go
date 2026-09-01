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

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

type BusinessModelModel struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID        string    `gorm:"column:tenant_id"`
	BusinessID      string    `gorm:"column:business_id"`
	AssetID         string    `gorm:"column:asset_id"`
	CurrentRevision int32     `gorm:"column:current_revision"`
	SchemaVersion   string    `gorm:"column:schema_version"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (BusinessModelModel) TableName() string { return "forma_business_model" }

type BusinessRevisionModel struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID          string    `gorm:"column:tenant_id"`
	BusinessID        string    `gorm:"column:business_id"`
	RevisionNo        int32     `gorm:"column:revision_no"`
	BaseRevisionNo    int32     `gorm:"column:base_revision_no"`
	SchemaVersion     string    `gorm:"column:schema_version"`
	SemanticModelJSON string    `gorm:"column:semantic_model_json"`
	ContentDigest     string    `gorm:"column:content_digest"`
	ChangeSummary     string    `gorm:"column:change_summary"`
	CreatedBy         string    `gorm:"column:created_by"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

func (BusinessRevisionModel) TableName() string { return "forma_business_model_revision" }

type BusinessLayoutModel struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID             string    `gorm:"column:tenant_id"`
	BusinessID           string    `gorm:"column:business_id"`
	LayoutRevision       int32     `gorm:"column:layout_revision"`
	BasedOnModelRevision int32     `gorm:"column:based_on_model_revision"`
	LayoutJSON           string    `gorm:"column:layout_json"`
	UpdatedBy            string    `gorm:"column:updated_by"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

func (BusinessLayoutModel) TableName() string { return "forma_business_model_layout" }

type BusinessDAO struct {
	db *gorm.DB
}

func NewBusinessDAO(db *gorm.DB) *BusinessDAO {
	return &BusinessDAO{db: db}
}

func (d *BusinessDAO) WithDB(db *gorm.DB) *BusinessDAO {
	return &BusinessDAO{db: db}
}

func (d *BusinessDAO) CreateMaster(ctx context.Context, m *entity.BusinessModel) error {
	return d.db.WithContext(ctx).Create(&BusinessModelModel{
		TenantID:        m.TenantID,
		BusinessID:      m.BusinessID,
		AssetID:         m.AssetID,
		CurrentRevision: m.CurrentRevision,
		SchemaVersion:   m.SchemaVersion,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}).Error
}

func (d *BusinessDAO) GetMaster(ctx context.Context, tenantID, businessID string) (*entity.BusinessModel, error) {
	var row BusinessModelModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMaster(&row), nil
}

func (d *BusinessDAO) ListMasters(ctx context.Context, tenantID string) ([]*entity.BusinessModel, error) {
	var rows []BusinessModelModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("updated_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.BusinessModel, 0, len(rows))
	for i := range rows {
		out = append(out, toMaster(&rows[i]))
	}
	return out, nil
}

func (d *BusinessDAO) CASBumpRevision(ctx context.Context, tenantID, businessID string, expected, next int32) (bool, error) {
	now := time.Now().UTC()
	res := d.db.WithContext(ctx).Model(&BusinessModelModel{}).
		Where("tenant_id = ? AND business_id = ? AND current_revision = ?", tenantID, businessID, expected).
		Updates(map[string]any{
			"current_revision": next,
			"updated_at":       now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (d *BusinessDAO) TouchUpdatedAt(ctx context.Context, tenantID, businessID string) error {
	return d.db.WithContext(ctx).Model(&BusinessModelModel{}).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		Update("updated_at", time.Now().UTC()).Error
}

func (d *BusinessDAO) CreateRevision(ctx context.Context, r *entity.BusinessModelRevision) error {
	return d.db.WithContext(ctx).Create(&BusinessRevisionModel{
		TenantID:          r.TenantID,
		BusinessID:        r.BusinessID,
		RevisionNo:        r.RevisionNo,
		BaseRevisionNo:    r.BaseRevisionNo,
		SchemaVersion:     r.SchemaVersion,
		SemanticModelJSON: r.SemanticModelJSON,
		ContentDigest:     r.ContentDigest,
		ChangeSummary:     r.ChangeSummary,
		CreatedBy:         r.CreatedBy,
		CreatedAt:         r.CreatedAt,
	}).Error
}

func (d *BusinessDAO) GetRevision(ctx context.Context, tenantID, businessID string, rev int32) (*entity.BusinessModelRevision, error) {
	var row BusinessRevisionModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ? AND revision_no = ?", tenantID, businessID, rev).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toRevision(&row), nil
}

func (d *BusinessDAO) ListRevisions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessModelRevision, error) {
	var rows []BusinessRevisionModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		Order("revision_no DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.BusinessModelRevision, 0, len(rows))
	for i := range rows {
		out = append(out, toRevision(&rows[i]))
	}
	return out, nil
}

func (d *BusinessDAO) UpsertLayout(ctx context.Context, l *entity.BusinessModelLayout) error {
	var existing BusinessLayoutModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", l.TenantID, l.BusinessID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return d.db.WithContext(ctx).Create(&BusinessLayoutModel{
			TenantID:             l.TenantID,
			BusinessID:           l.BusinessID,
			LayoutRevision:       l.LayoutRevision,
			BasedOnModelRevision: l.BasedOnModelRevision,
			LayoutJSON:           l.LayoutJSON,
			UpdatedBy:            l.UpdatedBy,
			UpdatedAt:            l.UpdatedAt,
		}).Error
	}
	if err != nil {
		return err
	}
	return d.db.WithContext(ctx).Model(&BusinessLayoutModel{}).
		Where("id = ?", existing.ID).
		Updates(map[string]any{
			"layout_revision":         l.LayoutRevision,
			"based_on_model_revision": l.BasedOnModelRevision,
			"layout_json":             l.LayoutJSON,
			"updated_by":              l.UpdatedBy,
			"updated_at":              l.UpdatedAt,
		}).Error
}

func (d *BusinessDAO) GetLayout(ctx context.Context, tenantID, businessID string) (*entity.BusinessModelLayout, error) {
	var row BusinessLayoutModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toLayout(&row), nil
}

func (d *BusinessDAO) CASBumpLayout(ctx context.Context, tenantID, businessID string, expected, next int32, basedOn int32, layoutJSON, updatedBy string) (bool, error) {
	now := time.Now().UTC()
	res := d.db.WithContext(ctx).Model(&BusinessLayoutModel{}).
		Where("tenant_id = ? AND business_id = ? AND layout_revision = ?", tenantID, businessID, expected).
		Updates(map[string]any{
			"layout_revision":         next,
			"based_on_model_revision": basedOn,
			"layout_json":             layoutJSON,
			"updated_by":              updatedBy,
			"updated_at":              now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func toMaster(r *BusinessModelModel) *entity.BusinessModel {
	return &entity.BusinessModel{
		ID:              r.ID,
		TenantID:        r.TenantID,
		BusinessID:      r.BusinessID,
		AssetID:         r.AssetID,
		CurrentRevision: r.CurrentRevision,
		SchemaVersion:   r.SchemaVersion,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func toRevision(r *BusinessRevisionModel) *entity.BusinessModelRevision {
	return &entity.BusinessModelRevision{
		ID:                r.ID,
		TenantID:          r.TenantID,
		BusinessID:        r.BusinessID,
		RevisionNo:        r.RevisionNo,
		BaseRevisionNo:    r.BaseRevisionNo,
		SchemaVersion:     r.SchemaVersion,
		SemanticModelJSON: r.SemanticModelJSON,
		ContentDigest:     r.ContentDigest,
		ChangeSummary:     r.ChangeSummary,
		CreatedBy:         r.CreatedBy,
		CreatedAt:         r.CreatedAt,
	}
}

func toLayout(r *BusinessLayoutModel) *entity.BusinessModelLayout {
	return &entity.BusinessModelLayout{
		ID:                   r.ID,
		TenantID:             r.TenantID,
		BusinessID:           r.BusinessID,
		LayoutRevision:       r.LayoutRevision,
		BasedOnModelRevision: r.BasedOnModelRevision,
		LayoutJSON:           r.LayoutJSON,
		UpdatedBy:            r.UpdatedBy,
		UpdatedAt:            r.UpdatedAt,
	}
}
