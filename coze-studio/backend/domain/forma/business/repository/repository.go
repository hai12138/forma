/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/business/internal/dal"
)

type BusinessRepository interface {
	CreateMaster(ctx context.Context, m *entity.BusinessModel) error
	GetMaster(ctx context.Context, tenantID, businessID string) (*entity.BusinessModel, error)
	ListMasters(ctx context.Context, tenantID string) ([]*entity.BusinessModel, error)
	CASBumpRevision(ctx context.Context, tenantID, businessID string, expected, next int32) (bool, error)
	TouchUpdatedAt(ctx context.Context, tenantID, businessID string) error
	CreateRevision(ctx context.Context, r *entity.BusinessModelRevision) error
	GetRevision(ctx context.Context, tenantID, businessID string, rev int32) (*entity.BusinessModelRevision, error)
	ListRevisions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessModelRevision, error)
	UpsertLayout(ctx context.Context, l *entity.BusinessModelLayout) error
	GetLayout(ctx context.Context, tenantID, businessID string) (*entity.BusinessModelLayout, error)
	CASBumpLayout(ctx context.Context, tenantID, businessID string, expected, next int32, basedOn int32, layoutJSON, updatedBy string) (bool, error)
	Transaction(ctx context.Context, fn func(txRepo BusinessRepository) error) error
}

type gormBusinessRepo struct {
	dao *dal.BusinessDAO
	db  *gorm.DB
}

func NewBusinessRepository(db *gorm.DB) BusinessRepository {
	return &gormBusinessRepo{dao: dal.NewBusinessDAO(db), db: db}
}

func (r *gormBusinessRepo) CreateMaster(ctx context.Context, m *entity.BusinessModel) error {
	return r.dao.CreateMaster(ctx, m)
}
func (r *gormBusinessRepo) GetMaster(ctx context.Context, tenantID, businessID string) (*entity.BusinessModel, error) {
	return r.dao.GetMaster(ctx, tenantID, businessID)
}
func (r *gormBusinessRepo) ListMasters(ctx context.Context, tenantID string) ([]*entity.BusinessModel, error) {
	return r.dao.ListMasters(ctx, tenantID)
}
func (r *gormBusinessRepo) CASBumpRevision(ctx context.Context, tenantID, businessID string, expected, next int32) (bool, error) {
	return r.dao.CASBumpRevision(ctx, tenantID, businessID, expected, next)
}
func (r *gormBusinessRepo) TouchUpdatedAt(ctx context.Context, tenantID, businessID string) error {
	return r.dao.TouchUpdatedAt(ctx, tenantID, businessID)
}
func (r *gormBusinessRepo) CreateRevision(ctx context.Context, rrev *entity.BusinessModelRevision) error {
	return r.dao.CreateRevision(ctx, rrev)
}
func (r *gormBusinessRepo) GetRevision(ctx context.Context, tenantID, businessID string, rev int32) (*entity.BusinessModelRevision, error) {
	return r.dao.GetRevision(ctx, tenantID, businessID, rev)
}
func (r *gormBusinessRepo) ListRevisions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessModelRevision, error) {
	return r.dao.ListRevisions(ctx, tenantID, businessID)
}
func (r *gormBusinessRepo) UpsertLayout(ctx context.Context, l *entity.BusinessModelLayout) error {
	return r.dao.UpsertLayout(ctx, l)
}
func (r *gormBusinessRepo) GetLayout(ctx context.Context, tenantID, businessID string) (*entity.BusinessModelLayout, error) {
	return r.dao.GetLayout(ctx, tenantID, businessID)
}
func (r *gormBusinessRepo) CASBumpLayout(ctx context.Context, tenantID, businessID string, expected, next int32, basedOn int32, layoutJSON, updatedBy string) (bool, error) {
	return r.dao.CASBumpLayout(ctx, tenantID, businessID, expected, next, basedOn, layoutJSON, updatedBy)
}

func (r *gormBusinessRepo) Transaction(ctx context.Context, fn func(txRepo BusinessRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormBusinessRepo{dao: dal.NewBusinessDAO(tx), db: tx})
	})
}
