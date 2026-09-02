/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/internal/dal"
	"gorm.io/gorm"
)

type DataRepository interface {
	CreateRequirement(ctx context.Context, req *entity.DataRequirement) error
	GetRequirement(ctx context.Context, tenantID, requirementID string) (*entity.DataRequirement, error)
	ListRequirementsByRevision(ctx context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error)
	UpdateRequirementStatusCAS(ctx context.Context, tenantID, requirementID string, from, to entity.RequirementStatus) (bool, error)

	// CreateOrClaimAnalysisRun inserts PENDING run. created=true means caller owns execution.
	// If key exists with same digest, returns existing and created=false.
	// If key exists with different digest, returns ErrAnalysisIdempotencyConflict.
	CreateOrClaimAnalysisRun(ctx context.Context, run *entity.DataRequirementAnalysisRun) (existing *entity.DataRequirementAnalysisRun, created bool, err error)
	GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error)
	GetAnalysisRunByIdempotencyKey(ctx context.Context, tenantID, businessID string, revision int32, clientRequestID string) (*entity.DataRequirementAnalysisRun, error)
	MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string, expectedGeneration int32) error
	MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorKey, sanitizedMsg string, expectedGeneration int32) error
	ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID, actorID string) (claimed bool, newGeneration int32, err error)
	// ClaimExpiredPendingExecution CAS-takes an abandoned PENDING run after lease expiry.
	ClaimExpiredPendingExecution(ctx context.Context, tenantID, analysisRunID string, expectedGeneration int32, now time.Time) (*entity.DataRequirementAnalysisRun, bool, error)

	CreateDecision(ctx context.Context, d *entity.DataRequirementDecision) error
	GetDecisionBySource(ctx context.Context, tenantID, sourceRequirementID string) (*entity.DataRequirementDecision, error)
	ListDecisionsByRequirement(ctx context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error)

	CreateRequirementsBatch(ctx context.Context, reqs []*entity.DataRequirement) error

	Transaction(ctx context.Context, fn func(txRepo DataRepository) error) error
}

type gormDataRepo struct {
	dao *dal.DataDAO
	db  *gorm.DB
}

func NewDataRepository(db *gorm.DB) DataRepository {
	return &gormDataRepo{dao: dal.NewDataDAO(db), db: db}
}

func (r *gormDataRepo) CreateRequirement(ctx context.Context, req *entity.DataRequirement) error {
	return r.dao.CreateRequirement(ctx, req)
}

func (r *gormDataRepo) GetRequirement(ctx context.Context, tenantID, requirementID string) (*entity.DataRequirement, error) {
	return r.dao.GetRequirement(ctx, tenantID, requirementID)
}

func (r *gormDataRepo) ListRequirementsByRevision(ctx context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error) {
	return r.dao.ListRequirementsByRevision(ctx, tenantID, businessID, revision)
}

func (r *gormDataRepo) UpdateRequirementStatusCAS(ctx context.Context, tenantID, requirementID string, from, to entity.RequirementStatus) (bool, error) {
	return r.dao.UpdateRequirementStatusCAS(ctx, tenantID, requirementID, from, to)
}

func (r *gormDataRepo) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.DataRequirementAnalysisRun) (*entity.DataRequirementAnalysisRun, bool, error) {
	return r.dao.CreateOrClaimAnalysisRun(ctx, run)
}

func (r *gormDataRepo) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error) {
	return r.dao.GetAnalysisRun(ctx, tenantID, analysisRunID)
}

func (r *gormDataRepo) GetAnalysisRunByIdempotencyKey(ctx context.Context, tenantID, businessID string, revision int32, clientRequestID string) (*entity.DataRequirementAnalysisRun, error) {
	return r.dao.GetAnalysisRunByIdempotencyKey(ctx, tenantID, businessID, revision, clientRequestID)
}

func (r *gormDataRepo) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string, expectedGeneration int32) error {
	return r.dao.MarkAnalysisSucceeded(ctx, tenantID, analysisRunID, modelRef, expectedGeneration)
}

func (r *gormDataRepo) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorKey, sanitizedMsg string, expectedGeneration int32) error {
	return r.dao.MarkAnalysisFailed(ctx, tenantID, analysisRunID, errorKey, sanitizedMsg, expectedGeneration)
}

func (r *gormDataRepo) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID, actorID string) (bool, int32, error) {
	return r.dao.ClaimAnalysisRetry(ctx, tenantID, analysisRunID, actorID)
}

func (r *gormDataRepo) ClaimExpiredPendingExecution(ctx context.Context, tenantID, analysisRunID string, expectedGeneration int32, now time.Time) (*entity.DataRequirementAnalysisRun, bool, error) {
	return r.dao.ClaimExpiredPendingExecution(ctx, tenantID, analysisRunID, expectedGeneration, now)
}

func (r *gormDataRepo) CreateDecision(ctx context.Context, d *entity.DataRequirementDecision) error {
	return r.dao.CreateDecision(ctx, d)
}

func (r *gormDataRepo) GetDecisionBySource(ctx context.Context, tenantID, sourceRequirementID string) (*entity.DataRequirementDecision, error) {
	return r.dao.GetDecisionBySource(ctx, tenantID, sourceRequirementID)
}

func (r *gormDataRepo) ListDecisionsByRequirement(ctx context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error) {
	return r.dao.ListDecisionsByRequirement(ctx, tenantID, requirementID)
}

func (r *gormDataRepo) CreateRequirementsBatch(ctx context.Context, reqs []*entity.DataRequirement) error {
	return r.dao.CreateRequirementsBatch(ctx, reqs)
}

func (r *gormDataRepo) Transaction(ctx context.Context, fn func(txRepo DataRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormDataRepo{dao: dal.NewDataDAO(tx), db: tx})
	})
}
