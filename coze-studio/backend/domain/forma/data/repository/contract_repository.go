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

type ContractRepository interface {
	Transaction(context.Context, func(ContractRepository) error) error

	CreateContract(context.Context, *entity.DataContract) error
	GetContract(context.Context, string, string) (*entity.DataContract, error)
	GetContractForUpdate(context.Context, string, string) (*entity.DataContract, error)
	ListContracts(context.Context, string, string) ([]*entity.DataContract, error)
	UpdateContractActiveRevision(context.Context, string, string, string) error
	ClearContractActiveRevisionIfMatch(context.Context, string, string, string) error

	CreateRevision(context.Context, *entity.DataContractRevision) error
	GetRevision(context.Context, string, string) (*entity.DataContractRevision, error)
	ListRevisions(context.Context, string, string) ([]*entity.DataContractRevision, error)
	GetRevisionByVersion(context.Context, string, string, int32) (*entity.DataContractRevision, error)
	AllocateNextVersion(context.Context, string, string) (int32, error)
	UpdateRevisionStatus(context.Context, string, string, entity.ContractStatus, entity.ContractStatus) error

	CreateValidationResult(context.Context, *entity.DataValidationResult) error
	ListValidationResults(context.Context, string, string) ([]*entity.DataValidationResult, error)
	GetValidationResult(context.Context, string, string) (*entity.DataValidationResult, error)

	CreateLifecycleEvent(context.Context, *entity.DataContractLifecycleEvent) error
	ListLifecycleEvents(context.Context, string, string) ([]*entity.DataContractLifecycleEvent, error)

	CreateDriftResult(context.Context, *entity.DataDriftResult) error
	ListDriftResults(context.Context, string, string) ([]*entity.DataDriftResult, error)

	CreateGapResult(context.Context, *entity.DataContractGapResult) error
	ListGapResults(context.Context, string, string) ([]*entity.DataContractGapResult, error)
}

type gormContractRepo struct {
	dao *dal.ContractDAO
	db  *gorm.DB
}

func NewContractRepository(db *gorm.DB) ContractRepository {
	return &gormContractRepo{dao: dal.NewContractDAO(db), db: db}
}

func (r *gormContractRepo) Transaction(c context.Context, fn func(ContractRepository) error) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		return fn(&gormContractRepo{dao: dal.NewContractDAO(tx), db: tx})
	})
}

func (r *gormContractRepo) CreateContract(c context.Context, v *entity.DataContract) error {
	return r.dao.CreateContract(c, v)
}
func (r *gormContractRepo) GetContract(c context.Context, t, id string) (*entity.DataContract, error) {
	return r.dao.GetContract(c, t, id)
}
func (r *gormContractRepo) GetContractForUpdate(c context.Context, t, id string) (*entity.DataContract, error) {
	return r.dao.GetContractForUpdate(c, t, id)
}
func (r *gormContractRepo) ListContracts(c context.Context, t, b string) ([]*entity.DataContract, error) {
	return r.dao.ListContracts(c, t, b)
}
func (r *gormContractRepo) UpdateContractActiveRevision(c context.Context, t, contractID, revisionID string) error {
	return r.dao.UpdateContractActiveRevision(c, t, contractID, revisionID)
}
func (r *gormContractRepo) ClearContractActiveRevisionIfMatch(c context.Context, t, contractID, expectedRevisionID string) error {
	return r.dao.ClearContractActiveRevisionIfMatch(c, t, contractID, expectedRevisionID)
}
func (r *gormContractRepo) CreateRevision(c context.Context, v *entity.DataContractRevision) error {
	return r.dao.CreateRevision(c, v)
}
func (r *gormContractRepo) GetRevision(c context.Context, t, id string) (*entity.DataContractRevision, error) {
	return r.dao.GetRevision(c, t, id)
}
func (r *gormContractRepo) ListRevisions(c context.Context, t, contractID string) ([]*entity.DataContractRevision, error) {
	return r.dao.ListRevisions(c, t, contractID)
}
func (r *gormContractRepo) GetRevisionByVersion(c context.Context, t, contractID string, version int32) (*entity.DataContractRevision, error) {
	return r.dao.GetRevisionByVersion(c, t, contractID, version)
}
func (r *gormContractRepo) AllocateNextVersion(c context.Context, t, contractID string) (int32, error) {
	return r.dao.AllocateNextVersion(c, t, contractID)
}
func (r *gormContractRepo) UpdateRevisionStatus(c context.Context, t, revisionID string, from, to entity.ContractStatus) error {
	return r.dao.UpdateRevisionStatus(c, t, revisionID, from, to)
}
func (r *gormContractRepo) CreateValidationResult(c context.Context, v *entity.DataValidationResult) error {
	return r.dao.CreateValidationResult(c, v)
}
func (r *gormContractRepo) ListValidationResults(c context.Context, t, revisionID string) ([]*entity.DataValidationResult, error) {
	return r.dao.ListValidationResults(c, t, revisionID)
}
func (r *gormContractRepo) GetValidationResult(c context.Context, t, id string) (*entity.DataValidationResult, error) {
	return r.dao.GetValidationResult(c, t, id)
}
func (r *gormContractRepo) CreateLifecycleEvent(c context.Context, v *entity.DataContractLifecycleEvent) error {
	return r.dao.CreateLifecycleEvent(c, v)
}
func (r *gormContractRepo) ListLifecycleEvents(c context.Context, t, contractID string) ([]*entity.DataContractLifecycleEvent, error) {
	return r.dao.ListLifecycleEvents(c, t, contractID)
}
func (r *gormContractRepo) CreateDriftResult(c context.Context, v *entity.DataDriftResult) error {
	return r.dao.CreateDriftResult(c, v)
}
func (r *gormContractRepo) ListDriftResults(c context.Context, t, revisionID string) ([]*entity.DataDriftResult, error) {
	return r.dao.ListDriftResults(c, t, revisionID)
}
func (r *gormContractRepo) CreateGapResult(c context.Context, v *entity.DataContractGapResult) error {
	return r.dao.CreateGapResult(c, v)
}
func (r *gormContractRepo) ListGapResults(c context.Context, t, revisionID string) ([]*entity.DataContractGapResult, error) {
	return r.dao.ListGapResults(c, t, revisionID)
}
