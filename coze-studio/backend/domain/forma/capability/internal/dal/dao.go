/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const analysisExecutionLease = 5 * time.Minute

type CapabilityDAO struct {
	db *gorm.DB
}

func NewCapabilityDAO(db *gorm.DB) *CapabilityDAO {
	return &CapabilityDAO{db: db}
}

func leaseExpiryFrom(now time.Time) time.Time {
	return now.Add(analysisExecutionLease)
}

func isDup(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (d *CapabilityDAO) CreateCapability(ctx context.Context, cap *entity.BusinessCapability) error {
	row := &capabilityRow{
		CapabilityID: cap.CapabilityID,
		TenantID:     cap.TenantID,
		BusinessID:   cap.BusinessID,
		CreatedBy:    cap.CreatedBy,
		CreatedAt:    cap.CreatedAt,
		UpdatedAt:    cap.UpdatedAt,
	}
	if cap.ActiveRevisionID != "" {
		row.ActiveRevisionID = strPtr(cap.ActiveRevisionID)
	}
	err := d.db.WithContext(ctx).Create(row).Error
	if isDup(err) {
		return entity.ErrConflict
	}
	return err
}

func (d *CapabilityDAO) GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	var row capabilityRow
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND capability_id = ?", tenantID, capabilityID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toCapability(&row), nil
}

func (d *CapabilityDAO) GetCapabilityForUpdate(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	var row capabilityRow
	err := d.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND capability_id = ?", tenantID, capabilityID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toCapability(&row), nil
}

func (d *CapabilityDAO) UpdateActiveRevisionID(ctx context.Context, tenantID, capabilityID, activeRevisionID string) error {
	now := time.Now().UTC()
	updates := map[string]any{"updated_at": now}
	if activeRevisionID == "" {
		updates["active_revision_id"] = nil
	} else {
		updates["active_revision_id"] = activeRevisionID
	}
	res := d.db.WithContext(ctx).Model(&capabilityRow{}).
		Where("tenant_id = ? AND capability_id = ?", tenantID, capabilityID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrNotFound
	}
	return nil
}

func (d *CapabilityDAO) CreateRevision(ctx context.Context, rev *entity.BusinessCapabilityRevision) error {
	row, err := revisionFrom(rev)
	if err != nil {
		return err
	}
	err = d.db.WithContext(ctx).Create(row).Error
	if isDup(err) {
		return entity.ErrConflict
	}
	return err
}

func (d *CapabilityDAO) GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	var row revisionRow
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND revision_id = ?", tenantID, revisionID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrRevisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return toRevision(&row)
}

func (d *CapabilityDAO) GetRevisionForUpdate(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	var row revisionRow
	err := d.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND revision_id = ?", tenantID, revisionID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrRevisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return toRevision(&row)
}

func (d *CapabilityDAO) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	var rows []revisionRow
	if err := d.db.WithContext(ctx).Where("tenant_id = ? AND capability_id = ?", tenantID, capabilityID).
		Order("version ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.BusinessCapabilityRevision, 0, len(rows))
	for i := range rows {
		r, err := toRevision(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (d *CapabilityDAO) UpdateRevisionStatus(ctx context.Context, tenantID, revisionID string, from, to entity.RevisionStatus) (bool, error) {
	res := d.db.WithContext(ctx).Model(&revisionRow{}).
		Where("tenant_id = ? AND revision_id = ? AND status = ?", tenantID, revisionID, string(from)).
		Update("status", string(to))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (d *CapabilityDAO) MaxVersion(ctx context.Context, tenantID, capabilityID string) (int32, error) {
	var max *int32
	err := d.db.WithContext(ctx).Model(&revisionRow{}).
		Where("tenant_id = ? AND capability_id = ?", tenantID, capabilityID).
		Select("MAX(version)").Scan(&max).Error
	if err != nil {
		return 0, err
	}
	if max == nil {
		return 0, nil
	}
	return *max, nil
}

func (d *CapabilityDAO) CreateProposal(ctx context.Context, p *entity.CapabilityProposal) error {
	payload, err := json.Marshal(p.Payload)
	if err != nil {
		return err
	}
	row := &proposalRow{
		ProposalID:    p.ProposalID,
		TenantID:      p.TenantID,
		BusinessID:    p.BusinessID,
		AnalysisRunID: p.AnalysisRunID,
		CapabilityID:  strPtr(p.CapabilityID),
		Status:        string(p.Status),
		PayloadJSON:   string(payload),
		CreatedAt:     p.CreatedAt,
	}
	if p.MaterializedRevisionID != "" {
		row.MaterializedRevisionID = strPtr(p.MaterializedRevisionID)
	}
	err = d.db.WithContext(ctx).Create(row).Error
	if isDup(err) {
		return entity.ErrConflict
	}
	return err
}

func (d *CapabilityDAO) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	var row proposalRow
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND proposal_id = ?", tenantID, proposalID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrProposalNotFound
	}
	if err != nil {
		return nil, err
	}
	return toProposal(&row)
}

func (d *CapabilityDAO) GetProposalForUpdate(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	var row proposalRow
	err := d.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND proposal_id = ?", tenantID, proposalID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrProposalNotFound
	}
	if err != nil {
		return nil, err
	}
	return toProposal(&row)
}

func (d *CapabilityDAO) ListProposalsByAnalysisRun(ctx context.Context, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error) {
	var rows []proposalRow
	if err := d.db.WithContext(ctx).Where("tenant_id = ? AND analysis_run_id = ?", tenantID, analysisRunID).
		Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.CapabilityProposal, 0, len(rows))
	for i := range rows {
		p, err := toProposal(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (d *CapabilityDAO) TerminalizeProposal(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, materializedRevisionID, capabilityID string) error {
	updates := map[string]any{"status": string(status)}
	if materializedRevisionID != "" {
		updates["materialized_revision_id"] = materializedRevisionID
	}
	if capabilityID != "" {
		updates["capability_id"] = capabilityID
	}
	res := d.db.WithContext(ctx).Model(&proposalRow{}).
		Where("tenant_id = ? AND proposal_id = ? AND status = ?", tenantID, proposalID, string(entity.ProposalProposed)).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrConflict
	}
	return nil
}

func (d *CapabilityDAO) CreateDecision(ctx context.Context, dec *entity.CapabilityDecision) error {
	row := &decisionRow{
		DecisionID:       dec.DecisionID,
		TenantID:         dec.TenantID,
		BusinessID:       dec.BusinessID,
		CapabilityID:     strPtr(dec.CapabilityID),
		ProposalID:       strPtr(dec.ProposalID),
		SourceRevisionID: strPtr(dec.SourceRevisionID),
		TargetRevisionID: strPtr(dec.TargetRevisionID),
		Action:           string(dec.Action),
		PayloadDigest:    dec.PayloadDigest,
		ClientRequestID:  strPtr(dec.ClientRequestID),
		ActorPrincipalID: dec.ActorPrincipalID,
		Reason:           dec.Reason,
		CreatedAt:        dec.CreatedAt,
	}
	err := d.db.WithContext(ctx).Create(row).Error
	if isDup(err) {
		return entity.ErrConflict
	}
	return err
}

func (d *CapabilityDAO) GetDecision(ctx context.Context, tenantID, decisionID string) (*entity.CapabilityDecision, error) {
	var row decisionRow
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND decision_id = ?", tenantID, decisionID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDecisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDecision(&row), nil
}

func (d *CapabilityDAO) GetDecisionByProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityDecision, error) {
	var row decisionRow
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND proposal_id = ?", tenantID, proposalID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDecisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDecision(&row), nil
}

func (d *CapabilityDAO) GetDecisionByDeriveKey(ctx context.Context, tenantID, capabilityID, sourceRevisionID, clientRequestID string) (*entity.CapabilityDecision, error) {
	var row decisionRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND capability_id = ? AND source_revision_id = ? AND client_request_id = ?",
			tenantID, capabilityID, sourceRevisionID, clientRequestID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrDecisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDecision(&row), nil
}

func (d *CapabilityDAO) ListDecisionsByCapability(ctx context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error) {
	var rows []decisionRow
	if err := d.db.WithContext(ctx).Where("tenant_id = ? AND capability_id = ?", tenantID, capabilityID).
		Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.CapabilityDecision, 0, len(rows))
	for i := range rows {
		out = append(out, toDecision(&rows[i]))
	}
	return out, nil
}

func (d *CapabilityDAO) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.CapabilityAnalysisRun) (*entity.CapabilityAnalysisRun, bool, error) {
	existing, err := d.getAnalysisByKey(ctx, run.TenantID, run.BusinessID, run.BusinessModelRevision, run.ClientRequestID)
	if err == nil {
		if existing.RequestDigest != run.RequestDigest {
			return nil, false, entity.ErrIdempotencyConflict
		}
		return existing, false, nil
	}
	if !errors.Is(err, entity.ErrAnalysisNotFound) {
		return nil, false, err
	}
	row := analysisFrom(run)
	if row.Attempt == 0 {
		row.Attempt = 1
	}
	if row.ExecutionClaimedAt == nil {
		now := time.Now().UTC()
		row.ExecutionClaimedAt = &now
	}
	if row.LeaseExpiresAt == nil {
		exp := leaseExpiryFrom(*row.ExecutionClaimedAt)
		row.LeaseExpiresAt = &exp
	}
	err = d.db.WithContext(ctx).Create(row).Error
	if err != nil {
		if isDup(err) {
			existing, gerr := d.getAnalysisByKey(ctx, run.TenantID, run.BusinessID, run.BusinessModelRevision, run.ClientRequestID)
			if gerr != nil {
				return nil, false, gerr
			}
			if existing.RequestDigest != run.RequestDigest {
				return nil, false, entity.ErrIdempotencyConflict
			}
			return existing, false, nil
		}
		return nil, false, err
	}
	return toAnalysis(row), true, nil
}

func (d *CapabilityDAO) getAnalysisByKey(ctx context.Context, tenantID, businessID string, revision int32, clientRequestID string) (*entity.CapabilityAnalysisRun, error) {
	var row analysisRunRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ? AND business_model_revision = ? AND client_request_id = ?",
			tenantID, businessID, revision, clientRequestID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrAnalysisNotFound
	}
	if err != nil {
		return nil, err
	}
	return toAnalysis(&row), nil
}

func (d *CapabilityDAO) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	var row analysisRunRow
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND analysis_run_id = ?", tenantID, analysisRunID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrAnalysisNotFound
	}
	if err != nil {
		return nil, err
	}
	return toAnalysis(&row), nil
}

func (d *CapabilityDAO) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string, expectedAttempt int32) error {
	now := time.Now().UTC()
	res := d.db.WithContext(ctx).Model(&analysisRunRow{}).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ? AND attempt = ?",
			tenantID, analysisRunID, string(entity.AnalysisPending), expectedAttempt).
		Updates(map[string]any{
			"status":     string(entity.AnalysisSucceeded),
			"model_ref":  modelRef,
			"error_code": "",
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrStaleGeneration
	}
	return nil
}

func (d *CapabilityDAO) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorCode string, expectedAttempt int32) error {
	now := time.Now().UTC()
	res := d.db.WithContext(ctx).Model(&analysisRunRow{}).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ? AND attempt = ?",
			tenantID, analysisRunID, string(entity.AnalysisPending), expectedAttempt).
		Updates(map[string]any{
			"status":     string(entity.AnalysisFailed),
			"error_code": errorCode,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrStaleGeneration
	}
	return nil
}

func (d *CapabilityDAO) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID, actorID string) (bool, int32, error) {
	_ = actorID
	var row analysisRunRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ?", tenantID, analysisRunID, string(entity.AnalysisFailed)).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	now := time.Now().UTC()
	exp := leaseExpiryFrom(now)
	newAttempt := row.Attempt + 1
	res := d.db.WithContext(ctx).Model(&analysisRunRow{}).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ? AND attempt = ?",
			tenantID, analysisRunID, string(entity.AnalysisFailed), row.Attempt).
		Updates(map[string]any{
			"status":               string(entity.AnalysisPending),
			"attempt":              newAttempt,
			"execution_claimed_at": now,
			"lease_expires_at":     exp,
			"error_code":           "",
			"updated_at":           now,
		})
	if res.Error != nil {
		return false, 0, res.Error
	}
	if res.RowsAffected != 1 {
		return false, 0, nil
	}
	return true, newAttempt, nil
}

func (d *CapabilityDAO) ClaimExpiredPendingExecution(ctx context.Context, tenantID, analysisRunID string, expectedAttempt int32, now time.Time) (*entity.CapabilityAnalysisRun, bool, error) {
	exp := leaseExpiryFrom(now)
	newAttempt := expectedAttempt + 1
	res := d.db.WithContext(ctx).Model(&analysisRunRow{}).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ? AND attempt = ? AND lease_expires_at <= ?",
			tenantID, analysisRunID, string(entity.AnalysisPending), expectedAttempt, now).
		Updates(map[string]any{
			"attempt":              newAttempt,
			"execution_claimed_at": now,
			"lease_expires_at":     exp,
			"updated_at":           now,
		})
	if res.Error != nil {
		return nil, false, res.Error
	}
	if res.RowsAffected != 1 {
		latest, err := d.GetAnalysisRun(ctx, tenantID, analysisRunID)
		return latest, false, err
	}
	run, err := d.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		return nil, false, err
	}
	return run, true, nil
}

func toCapability(row *capabilityRow) *entity.BusinessCapability {
	return &entity.BusinessCapability{
		CapabilityID:     row.CapabilityID,
		TenantID:         row.TenantID,
		BusinessID:       row.BusinessID,
		ActiveRevisionID: deref(row.ActiveRevisionID),
		CreatedBy:        row.CreatedBy,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func revisionFrom(rev *entity.BusinessCapabilityRevision) (*revisionRow, error) {
	inJSON, err := json.Marshal(rev.InputSchema)
	if err != nil {
		return nil, err
	}
	outJSON, err := json.Marshal(rev.OutputSchema)
	if err != nil {
		return nil, err
	}
	preJSON, err := json.Marshal(rev.Preconditions)
	if err != nil {
		return nil, err
	}
	if rev.Preconditions == nil {
		preJSON = []byte("[]")
	}
	effJSON, err := json.Marshal(rev.Effects)
	if err != nil {
		return nil, err
	}
	if rev.Effects == nil {
		effJSON = []byte("[]")
	}
	bindJSON, err := json.Marshal(rev.DataContractBindings)
	if err != nil {
		return nil, err
	}
	if rev.DataContractBindings == nil {
		bindJSON = []byte("[]")
	}
	return &revisionRow{
		RevisionID:               rev.RevisionID,
		CapabilityID:             rev.CapabilityID,
		TenantID:                 rev.TenantID,
		BusinessID:               rev.BusinessID,
		Version:                  rev.Version,
		Status:                   string(rev.Status),
		Name:                     rev.Name,
		Description:              rev.Description,
		BusinessModelRevision:    rev.BusinessModelRevision,
		CapabilityKind:           string(rev.CapabilityKind),
		InputSchemaJSON:          string(inJSON),
		OutputSchemaJSON:         string(outJSON),
		PreconditionsJSON:        string(preJSON),
		EffectsJSON:              string(effJSON),
		DataContractBindingsJSON: string(bindJSON),
		QueryOperation:           string(rev.QueryOperation),
		OutputCardinality:        string(rev.OutputCardinality),
		DerivedFromRevisionID:    rev.DerivedFromRevisionID,
		AnalysisRunID:            rev.AnalysisRunID,
		ProposalID:               rev.ProposalID,
		Source:                   string(rev.Source),
		CreatedBy:                rev.CreatedBy,
		CreatedAt:                rev.CreatedAt,
	}, nil
}

func toRevision(row *revisionRow) (*entity.BusinessCapabilityRevision, error) {
	rev := &entity.BusinessCapabilityRevision{
		RevisionID:            row.RevisionID,
		CapabilityID:          row.CapabilityID,
		TenantID:              row.TenantID,
		BusinessID:            row.BusinessID,
		Version:               row.Version,
		Status:                entity.RevisionStatus(row.Status),
		Name:                  row.Name,
		Description:           row.Description,
		BusinessModelRevision: row.BusinessModelRevision,
		CapabilityKind:        entity.CapabilityKind(row.CapabilityKind),
		QueryOperation:        entity.QueryOperation(row.QueryOperation),
		OutputCardinality:     entity.OutputCardinality(row.OutputCardinality),
		DerivedFromRevisionID: row.DerivedFromRevisionID,
		AnalysisRunID:         row.AnalysisRunID,
		ProposalID:            row.ProposalID,
		Source:                entity.Source(row.Source),
		CreatedBy:             row.CreatedBy,
		CreatedAt:             row.CreatedAt,
	}
	_ = json.Unmarshal([]byte(row.InputSchemaJSON), &rev.InputSchema)
	_ = json.Unmarshal([]byte(row.OutputSchemaJSON), &rev.OutputSchema)
	_ = json.Unmarshal([]byte(row.PreconditionsJSON), &rev.Preconditions)
	_ = json.Unmarshal([]byte(row.EffectsJSON), &rev.Effects)
	_ = json.Unmarshal([]byte(row.DataContractBindingsJSON), &rev.DataContractBindings)
	return rev, nil
}

func toProposal(row *proposalRow) (*entity.CapabilityProposal, error) {
	p := &entity.CapabilityProposal{
		ProposalID:             row.ProposalID,
		TenantID:               row.TenantID,
		BusinessID:             row.BusinessID,
		AnalysisRunID:          row.AnalysisRunID,
		CapabilityID:           deref(row.CapabilityID),
		Status:                 entity.ProposalStatus(row.Status),
		MaterializedRevisionID: deref(row.MaterializedRevisionID),
		CreatedAt:              row.CreatedAt,
	}
	if err := json.Unmarshal([]byte(row.PayloadJSON), &p.Payload); err != nil {
		return nil, err
	}
	return p, nil
}

func toDecision(row *decisionRow) *entity.CapabilityDecision {
	return &entity.CapabilityDecision{
		DecisionID:       row.DecisionID,
		TenantID:         row.TenantID,
		BusinessID:       row.BusinessID,
		CapabilityID:     deref(row.CapabilityID),
		ProposalID:       deref(row.ProposalID),
		SourceRevisionID: deref(row.SourceRevisionID),
		TargetRevisionID: deref(row.TargetRevisionID),
		Action:           entity.DecisionAction(row.Action),
		PayloadDigest:    row.PayloadDigest,
		ClientRequestID:  deref(row.ClientRequestID),
		ActorPrincipalID: row.ActorPrincipalID,
		Reason:           row.Reason,
		CreatedAt:        row.CreatedAt,
	}
}

func analysisFrom(run *entity.CapabilityAnalysisRun) *analysisRunRow {
	return &analysisRunRow{
		AnalysisRunID:         run.AnalysisRunID,
		TenantID:              run.TenantID,
		BusinessID:            run.BusinessID,
		BusinessModelRevision: run.BusinessModelRevision,
		ClientRequestID:       run.ClientRequestID,
		RequestDigest:         run.RequestDigest,
		Status:                string(run.Status),
		Attempt:               run.Attempt,
		ErrorCode:             run.ErrorCode,
		ModelRef:              run.ModelRef,
		RequestJSON:           run.RequestJSON,
		ExecutionClaimedAt:    run.ExecutionClaimedAt,
		LeaseExpiresAt:        run.LeaseExpiresAt,
		CreatedBy:             run.CreatedBy,
		CreatedAt:             run.CreatedAt,
		UpdatedAt:             run.UpdatedAt,
	}
}

func toAnalysis(row *analysisRunRow) *entity.CapabilityAnalysisRun {
	return &entity.CapabilityAnalysisRun{
		AnalysisRunID:         row.AnalysisRunID,
		TenantID:              row.TenantID,
		BusinessID:            row.BusinessID,
		BusinessModelRevision: row.BusinessModelRevision,
		ClientRequestID:       row.ClientRequestID,
		RequestDigest:         row.RequestDigest,
		Status:                entity.AnalysisStatus(row.Status),
		Attempt:               row.Attempt,
		ErrorCode:             row.ErrorCode,
		ModelRef:              row.ModelRef,
		RequestJSON:           row.RequestJSON,
		ExecutionClaimedAt:    row.ExecutionClaimedAt,
		LeaseExpiresAt:        row.LeaseExpiresAt,
		CreatedBy:             row.CreatedBy,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}
