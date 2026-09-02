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

type MappingRepository interface {
	CreateMapping(context.Context, *entity.SemanticMapping) error
	CreateMappingsBatch(context.Context, []*entity.SemanticMapping) error
	GetMapping(context.Context, string, string) (*entity.SemanticMapping, error)
	ListMappings(context.Context, string, string, int32, entity.MappingStatus) ([]*entity.SemanticMapping, error)
	UpdateMappingStatusCAS(context.Context, string, string, entity.MappingStatus, entity.MappingStatus) (bool, error)
	GetConfirmedMappingByRequirement(context.Context, string, string, int32, string) (*entity.SemanticMapping, error)

	CreateOrClaimMappingAnalysisRun(context.Context, *entity.SemanticMappingAnalysisRun) (*entity.SemanticMappingAnalysisRun, bool, error)
	GetMappingAnalysisRun(context.Context, string, string) (*entity.SemanticMappingAnalysisRun, error)
	MarkMappingAnalysisSucceeded(context.Context, string, string, string, int32) error
	MarkMappingAnalysisFailed(context.Context, string, string, string, string, int32) error
	ClaimMappingAnalysisRetry(context.Context, string, string, string) (bool, int32, error)
	ClaimExpiredMappingAnalysis(context.Context, string, string, int32, time.Time) (*entity.SemanticMappingAnalysisRun, bool, error)

	CreateMappingDecision(context.Context, *entity.SemanticMappingDecision) error
	ListMappingDecisions(context.Context, string, string) ([]*entity.SemanticMappingDecision, error)
	Transaction(context.Context, func(MappingRepository) error) error
}

type gormMappingRepo struct {
	dao *dal.MappingDAO
	db  *gorm.DB
}

func NewMappingRepository(db *gorm.DB) MappingRepository {
	return &gormMappingRepo{dao: dal.NewMappingDAO(db), db: db}
}
func (r *gormMappingRepo) CreateMapping(c context.Context, v *entity.SemanticMapping) error {
	return r.dao.CreateMapping(c, v)
}
func (r *gormMappingRepo) CreateMappingsBatch(c context.Context, v []*entity.SemanticMapping) error {
	return r.dao.CreateMappingsBatch(c, v)
}
func (r *gormMappingRepo) GetMapping(c context.Context, t, id string) (*entity.SemanticMapping, error) {
	return r.dao.GetMapping(c, t, id)
}
func (r *gormMappingRepo) ListMappings(c context.Context, t, b string, rev int32, s entity.MappingStatus) ([]*entity.SemanticMapping, error) {
	return r.dao.ListMappings(c, t, b, rev, s)
}
func (r *gormMappingRepo) UpdateMappingStatusCAS(c context.Context, t, id string, f, to entity.MappingStatus) (bool, error) {
	return r.dao.UpdateMappingStatusCAS(c, t, id, f, to)
}
func (r *gormMappingRepo) GetConfirmedMappingByRequirement(c context.Context, t, b string, rev int32, req string) (*entity.SemanticMapping, error) {
	return r.dao.GetConfirmedMappingByRequirement(c, t, b, rev, req)
}
func (r *gormMappingRepo) CreateOrClaimMappingAnalysisRun(c context.Context, v *entity.SemanticMappingAnalysisRun) (*entity.SemanticMappingAnalysisRun, bool, error) {
	return r.dao.CreateOrClaimAnalysisRun(c, v)
}
func (r *gormMappingRepo) GetMappingAnalysisRun(c context.Context, t, id string) (*entity.SemanticMappingAnalysisRun, error) {
	return r.dao.GetAnalysisRun(c, t, id)
}
func (r *gormMappingRepo) MarkMappingAnalysisSucceeded(c context.Context, t, id, m string, g int32) error {
	return r.dao.MarkAnalysisSucceeded(c, t, id, m, g)
}
func (r *gormMappingRepo) MarkMappingAnalysisFailed(c context.Context, t, id, k, msg string, g int32) error {
	return r.dao.MarkAnalysisFailed(c, t, id, k, msg, g)
}
func (r *gormMappingRepo) ClaimMappingAnalysisRetry(c context.Context, t, id, a string) (bool, int32, error) {
	return r.dao.ClaimAnalysisRetry(c, t, id, a)
}
func (r *gormMappingRepo) ClaimExpiredMappingAnalysis(c context.Context, t, id string, g int32, now time.Time) (*entity.SemanticMappingAnalysisRun, bool, error) {
	return r.dao.ClaimExpiredAnalysis(c, t, id, g, now)
}
func (r *gormMappingRepo) CreateMappingDecision(c context.Context, v *entity.SemanticMappingDecision) error {
	return r.dao.CreateDecision(c, v)
}
func (r *gormMappingRepo) ListMappingDecisions(c context.Context, t, id string) ([]*entity.SemanticMappingDecision, error) {
	return r.dao.ListDecisions(c, t, id)
}
func (r *gormMappingRepo) Transaction(c context.Context, fn func(MappingRepository) error) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error { return fn(&gormMappingRepo{dao: dal.NewMappingDAO(tx), db: tx}) })
}
