/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/internal/dal"
	"gorm.io/gorm"
)

// CapabilityRepository is the Capability domain persistence boundary.
type CapabilityRepository interface {
	CreateCapability(ctx context.Context, cap *entity.BusinessCapability) error
	GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error)
	GetCapabilityForUpdate(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error)
	UpdateActiveRevisionID(ctx context.Context, tenantID, capabilityID, activeRevisionID string) error
	CASBumpAggregateGeneration(ctx context.Context, tenantID, capabilityID string, expectedGen int64) (bool, error)

	CreateRevision(ctx context.Context, rev *entity.BusinessCapabilityRevision) error
	GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error)
	GetRevisionForUpdate(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error)
	ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error)
	UpdateRevisionStatus(ctx context.Context, tenantID, revisionID string, from, to entity.RevisionStatus) (bool, error)
	MaxVersion(ctx context.Context, tenantID, capabilityID string) (int32, error)

	CreateProposal(ctx context.Context, p *entity.CapabilityProposal) error
	GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error)
	GetProposalForUpdate(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error)
	ListProposalsByAnalysisRun(ctx context.Context, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error)
	TerminalizeProposal(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, materializedRevisionID, capabilityID string) error

	CreateDecision(ctx context.Context, d *entity.CapabilityDecision) error
	GetDecision(ctx context.Context, tenantID, decisionID string) (*entity.CapabilityDecision, error)
	GetDecisionByProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityDecision, error)
	GetDecisionByDeriveKey(ctx context.Context, tenantID, capabilityID, sourceRevisionID, clientRequestID string) (*entity.CapabilityDecision, error)
	ListDecisionsByCapability(ctx context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error)

	CreateOrClaimAnalysisRun(ctx context.Context, run *entity.CapabilityAnalysisRun) (existing *entity.CapabilityAnalysisRun, created bool, err error)
	GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error)
	MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string, expectedAttempt int32) error
	MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorCode string, expectedAttempt int32) error
	ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID, actorID string) (claimed bool, newAttempt int32, err error)
	ClaimExpiredPendingExecution(ctx context.Context, tenantID, analysisRunID string, expectedAttempt int32, now time.Time) (*entity.CapabilityAnalysisRun, bool, error)

	Transaction(ctx context.Context, fn func(txRepo CapabilityRepository) error) error
}

type gormCapabilityRepo struct {
	dao *dal.CapabilityDAO
	db  *gorm.DB
}

func NewCapabilityRepository(db *gorm.DB) CapabilityRepository {
	return &gormCapabilityRepo{dao: dal.NewCapabilityDAO(db), db: db}
}

func (r *gormCapabilityRepo) CreateCapability(ctx context.Context, cap *entity.BusinessCapability) error {
	return r.dao.CreateCapability(ctx, cap)
}
func (r *gormCapabilityRepo) GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	return r.dao.GetCapability(ctx, tenantID, capabilityID)
}
func (r *gormCapabilityRepo) GetCapabilityForUpdate(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	return r.dao.GetCapabilityForUpdate(ctx, tenantID, capabilityID)
}
func (r *gormCapabilityRepo) UpdateActiveRevisionID(ctx context.Context, tenantID, capabilityID, activeRevisionID string) error {
	return r.dao.UpdateActiveRevisionID(ctx, tenantID, capabilityID, activeRevisionID)
}
func (r *gormCapabilityRepo) CASBumpAggregateGeneration(ctx context.Context, tenantID, capabilityID string, expectedGen int64) (bool, error) {
	return r.dao.CASBumpAggregateGeneration(ctx, tenantID, capabilityID, expectedGen)
}
func (r *gormCapabilityRepo) CreateRevision(ctx context.Context, rev *entity.BusinessCapabilityRevision) error {
	return r.dao.CreateRevision(ctx, rev)
}
func (r *gormCapabilityRepo) GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	return r.dao.GetRevision(ctx, tenantID, revisionID)
}
func (r *gormCapabilityRepo) GetRevisionForUpdate(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	return r.dao.GetRevisionForUpdate(ctx, tenantID, revisionID)
}
func (r *gormCapabilityRepo) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	return r.dao.ListRevisions(ctx, tenantID, capabilityID)
}
func (r *gormCapabilityRepo) UpdateRevisionStatus(ctx context.Context, tenantID, revisionID string, from, to entity.RevisionStatus) (bool, error) {
	return r.dao.UpdateRevisionStatus(ctx, tenantID, revisionID, from, to)
}
func (r *gormCapabilityRepo) MaxVersion(ctx context.Context, tenantID, capabilityID string) (int32, error) {
	return r.dao.MaxVersion(ctx, tenantID, capabilityID)
}
func (r *gormCapabilityRepo) CreateProposal(ctx context.Context, p *entity.CapabilityProposal) error {
	return r.dao.CreateProposal(ctx, p)
}
func (r *gormCapabilityRepo) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	return r.dao.GetProposal(ctx, tenantID, proposalID)
}
func (r *gormCapabilityRepo) GetProposalForUpdate(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	return r.dao.GetProposalForUpdate(ctx, tenantID, proposalID)
}
func (r *gormCapabilityRepo) ListProposalsByAnalysisRun(ctx context.Context, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error) {
	return r.dao.ListProposalsByAnalysisRun(ctx, tenantID, analysisRunID)
}
func (r *gormCapabilityRepo) TerminalizeProposal(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, materializedRevisionID, capabilityID string) error {
	return r.dao.TerminalizeProposal(ctx, tenantID, proposalID, status, materializedRevisionID, capabilityID)
}
func (r *gormCapabilityRepo) CreateDecision(ctx context.Context, d *entity.CapabilityDecision) error {
	return r.dao.CreateDecision(ctx, d)
}
func (r *gormCapabilityRepo) GetDecision(ctx context.Context, tenantID, decisionID string) (*entity.CapabilityDecision, error) {
	return r.dao.GetDecision(ctx, tenantID, decisionID)
}
func (r *gormCapabilityRepo) GetDecisionByProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityDecision, error) {
	return r.dao.GetDecisionByProposal(ctx, tenantID, proposalID)
}
func (r *gormCapabilityRepo) GetDecisionByDeriveKey(ctx context.Context, tenantID, capabilityID, sourceRevisionID, clientRequestID string) (*entity.CapabilityDecision, error) {
	return r.dao.GetDecisionByDeriveKey(ctx, tenantID, capabilityID, sourceRevisionID, clientRequestID)
}
func (r *gormCapabilityRepo) ListDecisionsByCapability(ctx context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error) {
	return r.dao.ListDecisionsByCapability(ctx, tenantID, capabilityID)
}
func (r *gormCapabilityRepo) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.CapabilityAnalysisRun) (*entity.CapabilityAnalysisRun, bool, error) {
	return r.dao.CreateOrClaimAnalysisRun(ctx, run)
}
func (r *gormCapabilityRepo) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	return r.dao.GetAnalysisRun(ctx, tenantID, analysisRunID)
}
func (r *gormCapabilityRepo) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string, expectedAttempt int32) error {
	return r.dao.MarkAnalysisSucceeded(ctx, tenantID, analysisRunID, modelRef, expectedAttempt)
}
func (r *gormCapabilityRepo) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorCode string, expectedAttempt int32) error {
	return r.dao.MarkAnalysisFailed(ctx, tenantID, analysisRunID, errorCode, expectedAttempt)
}
func (r *gormCapabilityRepo) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID, actorID string) (bool, int32, error) {
	return r.dao.ClaimAnalysisRetry(ctx, tenantID, analysisRunID, actorID)
}
func (r *gormCapabilityRepo) ClaimExpiredPendingExecution(ctx context.Context, tenantID, analysisRunID string, expectedAttempt int32, now time.Time) (*entity.CapabilityAnalysisRun, bool, error) {
	return r.dao.ClaimExpiredPendingExecution(ctx, tenantID, analysisRunID, expectedAttempt, now)
}
func (r *gormCapabilityRepo) Transaction(ctx context.Context, fn func(txRepo CapabilityRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormCapabilityRepo{dao: dal.NewCapabilityDAO(tx), db: tx})
	})
}
