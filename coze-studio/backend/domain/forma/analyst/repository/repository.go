/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/internal/dal"
)

type AnalystRepository interface {
	CreateSession(ctx context.Context, s *entity.AnalystSession) error
	GetSession(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error)
	ListSessions(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystSession, error)
	UpdateSession(ctx context.Context, s *entity.AnalystSession) error

	CreateTurn(ctx context.Context, t *entity.AnalystTurn) error
	GetTurnByClientRequestID(ctx context.Context, tenantID, sessionID, clientRequestID string) (*entity.AnalystTurn, error)
	GetTurn(ctx context.Context, tenantID, turnID string) (*entity.AnalystTurn, error)
	ListTurns(ctx context.Context, tenantID, sessionID string) ([]*entity.AnalystTurn, error)
	MaxTurnSequence(ctx context.Context, tenantID, sessionID string) (int32, error)
	UpdateTurnAnalysis(ctx context.Context, tenantID, turnID string, status entity.AnalysisStatus, modelRequestID string) error

	CreateEvidence(ctx context.Context, e *entity.BusinessEvidence) error
	ListEvidence(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessEvidence, error)
	GetEvidence(ctx context.Context, tenantID, evidenceID string) (*entity.BusinessEvidence, error)

	CreateAssertion(ctx context.Context, a *entity.BusinessAssertion) error
	UpdateAssertion(ctx context.Context, a *entity.BusinessAssertion) error
	GetAssertion(ctx context.Context, tenantID, assertionID string) (*entity.BusinessAssertion, error)
	ListAssertions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessAssertion, error)
	CreateAssertionEvidenceRef(ctx context.Context, tenantID, assertionID, evidenceID string, at time.Time) error
	ListEvidenceIDsForAssertion(ctx context.Context, tenantID, assertionID string) ([]string, error)
	ListAssertionIDsForEvidence(ctx context.Context, tenantID, evidenceID string) ([]string, error)

	CreateConfirmation(ctx context.Context, c *entity.BusinessConfirmation) error
	ListConfirmationsForAssertion(ctx context.Context, tenantID, assertionID string) ([]*entity.BusinessConfirmation, error)

	CreateConflict(ctx context.Context, c *entity.AssertionConflict) error
	GetConflictByPair(ctx context.Context, tenantID, businessID, assertionIDA, assertionIDB string) (*entity.AssertionConflict, error)
	UpdateConflictStatus(ctx context.Context, tenantID, conflictID string, status entity.ConflictStatus, at time.Time) error
	ListConflicts(ctx context.Context, tenantID, businessID string) ([]*entity.AssertionConflict, error)

	GetSessionForUpdate(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error)

	CreateGap(ctx context.Context, g *entity.AnalystGap) error
	ListGaps(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystGap, error)
	UpdateGapStatus(ctx context.Context, tenantID, gapID string, status entity.GapStatus, at time.Time) error

	CreateProposal(ctx context.Context, p *entity.BusinessModelProposal) error
	GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error)
	UpdateProposalStatus(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, at time.Time) error

	CreateProvenance(ctx context.Context, p *entity.RevisionProvenance) error
	GetProvenance(ctx context.Context, tenantID, businessID string, revisionNo int32) (*entity.RevisionProvenance, error)

	CreateModelCall(ctx context.Context, r *entity.ModelCallRecord) error

	Transaction(ctx context.Context, fn func(txRepo AnalystRepository) error) error
}

type gormAnalystRepo struct {
	dao *dal.AnalystDAO
	db  *gorm.DB
}

func NewAnalystRepository(db *gorm.DB) AnalystRepository {
	return &gormAnalystRepo{dao: dal.NewAnalystDAO(db), db: db}
}

func (r *gormAnalystRepo) Transaction(ctx context.Context, fn func(txRepo AnalystRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormAnalystRepo{dao: dal.NewAnalystDAO(tx), db: tx})
	})
}

func (r *gormAnalystRepo) CreateSession(ctx context.Context, s *entity.AnalystSession) error {
	return r.dao.CreateSession(ctx, s)
}
func (r *gormAnalystRepo) GetSession(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error) {
	return r.dao.GetSession(ctx, tenantID, sessionID)
}
func (r *gormAnalystRepo) ListSessions(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystSession, error) {
	return r.dao.ListSessions(ctx, tenantID, businessID)
}
func (r *gormAnalystRepo) UpdateSession(ctx context.Context, s *entity.AnalystSession) error {
	return r.dao.UpdateSession(ctx, s)
}
func (r *gormAnalystRepo) CreateTurn(ctx context.Context, t *entity.AnalystTurn) error {
	return r.dao.CreateTurn(ctx, t)
}
func (r *gormAnalystRepo) GetTurnByClientRequestID(ctx context.Context, tenantID, sessionID, clientRequestID string) (*entity.AnalystTurn, error) {
	return r.dao.GetTurnByClientRequestID(ctx, tenantID, sessionID, clientRequestID)
}
func (r *gormAnalystRepo) GetTurn(ctx context.Context, tenantID, turnID string) (*entity.AnalystTurn, error) {
	return r.dao.GetTurn(ctx, tenantID, turnID)
}
func (r *gormAnalystRepo) ListTurns(ctx context.Context, tenantID, sessionID string) ([]*entity.AnalystTurn, error) {
	return r.dao.ListTurns(ctx, tenantID, sessionID)
}
func (r *gormAnalystRepo) MaxTurnSequence(ctx context.Context, tenantID, sessionID string) (int32, error) {
	return r.dao.MaxTurnSequence(ctx, tenantID, sessionID)
}
func (r *gormAnalystRepo) UpdateTurnAnalysis(ctx context.Context, tenantID, turnID string, status entity.AnalysisStatus, modelRequestID string) error {
	return r.dao.UpdateTurnAnalysis(ctx, tenantID, turnID, status, modelRequestID)
}
func (r *gormAnalystRepo) CreateEvidence(ctx context.Context, e *entity.BusinessEvidence) error {
	return r.dao.CreateEvidence(ctx, e)
}
func (r *gormAnalystRepo) ListEvidence(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessEvidence, error) {
	return r.dao.ListEvidence(ctx, tenantID, businessID)
}
func (r *gormAnalystRepo) GetEvidence(ctx context.Context, tenantID, evidenceID string) (*entity.BusinessEvidence, error) {
	return r.dao.GetEvidence(ctx, tenantID, evidenceID)
}
func (r *gormAnalystRepo) CreateAssertion(ctx context.Context, a *entity.BusinessAssertion) error {
	return r.dao.CreateAssertion(ctx, a)
}
func (r *gormAnalystRepo) UpdateAssertion(ctx context.Context, a *entity.BusinessAssertion) error {
	return r.dao.UpdateAssertion(ctx, a)
}
func (r *gormAnalystRepo) GetAssertion(ctx context.Context, tenantID, assertionID string) (*entity.BusinessAssertion, error) {
	return r.dao.GetAssertion(ctx, tenantID, assertionID)
}
func (r *gormAnalystRepo) ListAssertions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessAssertion, error) {
	return r.dao.ListAssertions(ctx, tenantID, businessID)
}
func (r *gormAnalystRepo) CreateAssertionEvidenceRef(ctx context.Context, tenantID, assertionID, evidenceID string, at time.Time) error {
	return r.dao.CreateAssertionEvidenceRef(ctx, tenantID, assertionID, evidenceID, at)
}
func (r *gormAnalystRepo) ListEvidenceIDsForAssertion(ctx context.Context, tenantID, assertionID string) ([]string, error) {
	return r.dao.ListEvidenceIDsForAssertion(ctx, tenantID, assertionID)
}
func (r *gormAnalystRepo) ListAssertionIDsForEvidence(ctx context.Context, tenantID, evidenceID string) ([]string, error) {
	return r.dao.ListAssertionIDsForEvidence(ctx, tenantID, evidenceID)
}
func (r *gormAnalystRepo) CreateConfirmation(ctx context.Context, c *entity.BusinessConfirmation) error {
	return r.dao.CreateConfirmation(ctx, c)
}
func (r *gormAnalystRepo) ListConfirmationsForAssertion(ctx context.Context, tenantID, assertionID string) ([]*entity.BusinessConfirmation, error) {
	return r.dao.ListConfirmationsForAssertion(ctx, tenantID, assertionID)
}
func (r *gormAnalystRepo) GetSessionForUpdate(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error) {
	return r.dao.GetSessionForUpdate(ctx, tenantID, sessionID)
}
func (r *gormAnalystRepo) CreateConflict(ctx context.Context, c *entity.AssertionConflict) error {
	return r.dao.CreateConflict(ctx, c)
}
func (r *gormAnalystRepo) GetConflictByPair(ctx context.Context, tenantID, businessID, assertionIDA, assertionIDB string) (*entity.AssertionConflict, error) {
	return r.dao.GetConflictByPair(ctx, tenantID, businessID, assertionIDA, assertionIDB)
}
func (r *gormAnalystRepo) UpdateConflictStatus(ctx context.Context, tenantID, conflictID string, status entity.ConflictStatus, at time.Time) error {
	return r.dao.UpdateConflictStatus(ctx, tenantID, conflictID, status, at)
}
func (r *gormAnalystRepo) ListConflicts(ctx context.Context, tenantID, businessID string) ([]*entity.AssertionConflict, error) {
	return r.dao.ListConflicts(ctx, tenantID, businessID)
}
func (r *gormAnalystRepo) CreateGap(ctx context.Context, g *entity.AnalystGap) error {
	return r.dao.CreateGap(ctx, g)
}
func (r *gormAnalystRepo) ListGaps(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystGap, error) {
	return r.dao.ListGaps(ctx, tenantID, businessID)
}
func (r *gormAnalystRepo) UpdateGapStatus(ctx context.Context, tenantID, gapID string, status entity.GapStatus, at time.Time) error {
	return r.dao.UpdateGapStatus(ctx, tenantID, gapID, status, at)
}
func (r *gormAnalystRepo) CreateProposal(ctx context.Context, p *entity.BusinessModelProposal) error {
	return r.dao.CreateProposal(ctx, p)
}
func (r *gormAnalystRepo) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error) {
	return r.dao.GetProposal(ctx, tenantID, proposalID)
}
func (r *gormAnalystRepo) UpdateProposalStatus(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, at time.Time) error {
	return r.dao.UpdateProposalStatus(ctx, tenantID, proposalID, status, at)
}
func (r *gormAnalystRepo) CreateProvenance(ctx context.Context, p *entity.RevisionProvenance) error {
	return r.dao.CreateProvenance(ctx, p)
}
func (r *gormAnalystRepo) GetProvenance(ctx context.Context, tenantID, businessID string, revisionNo int32) (*entity.RevisionProvenance, error) {
	return r.dao.GetProvenance(ctx, tenantID, businessID, revisionNo)
}
func (r *gormAnalystRepo) CreateModelCall(ctx context.Context, record *entity.ModelCallRecord) error {
	return r.dao.CreateModelCall(ctx, record)
}
