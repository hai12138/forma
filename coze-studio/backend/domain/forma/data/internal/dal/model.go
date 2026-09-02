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

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DataDAO struct {
	db *gorm.DB
}

func NewDataDAO(db *gorm.DB) *DataDAO {
	return &DataDAO{db: db}
}

type requirementRow struct {
	ID                       int64     `gorm:"column:id;primaryKey"`
	RequirementID            string    `gorm:"column:requirement_id"`
	TenantID                 string    `gorm:"column:tenant_id"`
	BusinessID               string    `gorm:"column:business_id"`
	BusinessModelRevision    int32     `gorm:"column:business_model_revision"`
	RequirementKind          string    `gorm:"column:requirement_kind"`
	SemanticName             string    `gorm:"column:semantic_name"`
	Description              string    `gorm:"column:description"`
	BusinessElementRefsJSON  string    `gorm:"column:business_element_refs_json"`
	Requiredness             string    `gorm:"column:requiredness"`
	FreshnessRequirement     string    `gorm:"column:freshness_requirement"`
	AccessNeed               string    `gorm:"column:access_need"`
	Status                   string    `gorm:"column:status"`
	Source                   string    `gorm:"column:source"`
	DerivedFromRequirementID string    `gorm:"column:derived_from_requirement_id"`
	AnalysisRunID            string    `gorm:"column:analysis_run_id"`
	CreatedBy                string    `gorm:"column:created_by"`
	CreatedAt                time.Time `gorm:"column:created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at"`
}

func (requirementRow) TableName() string { return "forma_data_requirement" }

type analysisRunRow struct {
	ID                    int64      `gorm:"column:id;primaryKey"`
	AnalysisRunID         string     `gorm:"column:analysis_run_id"`
	TenantID              string     `gorm:"column:tenant_id"`
	BusinessID            string     `gorm:"column:business_id"`
	BusinessModelRevision int32      `gorm:"column:business_model_revision"`
	ClientRequestID       string     `gorm:"column:client_request_id"`
	RequestDigest         string     `gorm:"column:request_digest"`
	Status                string     `gorm:"column:status"`
	ModelRef              string     `gorm:"column:model_ref"`
	ErrorKey              string     `gorm:"column:error_key"`
	ErrorMessageSanitized string     `gorm:"column:error_message_sanitized"`
	RetryCount            int32      `gorm:"column:retry_count"`
	CreatedBy             string     `gorm:"column:created_by"`
	StartedAt             *time.Time `gorm:"column:started_at"`
	CompletedAt           *time.Time `gorm:"column:completed_at"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (analysisRunRow) TableName() string { return "forma_data_requirement_analysis_run" }

type decisionRow struct {
	ID                    int64     `gorm:"column:id;primaryKey"`
	DecisionID            string    `gorm:"column:decision_id"`
	TenantID              string    `gorm:"column:tenant_id"`
	BusinessID            string    `gorm:"column:business_id"`
	SourceRequirementID   string    `gorm:"column:source_requirement_id"`
	TargetRequirementID   string    `gorm:"column:target_requirement_id"`
	Action                string    `gorm:"column:action"`
	ActorPrincipalID      string    `gorm:"column:actor_principal_id"`
	Reason                string    `gorm:"column:reason"`
	BusinessModelRevision int32     `gorm:"column:business_model_revision"`
	CreatedAt             time.Time `gorm:"column:created_at"`
}

func (decisionRow) TableName() string { return "forma_data_requirement_decision" }

func encodeRefs(refs []string) string {
	if refs == nil {
		refs = []string{}
	}
	b, _ := json.Marshal(refs)
	return string(b)
}

func decodeRefs(s string) []string {
	if s == "" {
		return []string{}
	}
	var out []string
	_ = json.Unmarshal([]byte(s), &out)
	if out == nil {
		return []string{}
	}
	return out
}

func toRequirement(r *requirementRow) *entity.DataRequirement {
	return &entity.DataRequirement{
		RequirementID:            r.RequirementID,
		TenantID:                 r.TenantID,
		BusinessID:               r.BusinessID,
		BusinessModelRevision:    r.BusinessModelRevision,
		RequirementKind:          entity.RequirementKind(r.RequirementKind),
		SemanticName:             r.SemanticName,
		Description:              r.Description,
		BusinessElementRefs:      decodeRefs(r.BusinessElementRefsJSON),
		Requiredness:             r.Requiredness,
		FreshnessRequirement:     r.FreshnessRequirement,
		AccessNeed:               r.AccessNeed,
		Status:                   entity.RequirementStatus(r.Status),
		Source:                   entity.RequirementSource(r.Source),
		DerivedFromRequirementID: r.DerivedFromRequirementID,
		AnalysisRunID:            r.AnalysisRunID,
		CreatedBy:                r.CreatedBy,
		CreatedAt:                r.CreatedAt,
		UpdatedAt:                r.UpdatedAt,
	}
}

func fromRequirement(req *entity.DataRequirement) *requirementRow {
	return &requirementRow{
		RequirementID:            req.RequirementID,
		TenantID:                 req.TenantID,
		BusinessID:               req.BusinessID,
		BusinessModelRevision:    req.BusinessModelRevision,
		RequirementKind:          string(req.RequirementKind),
		SemanticName:             req.SemanticName,
		Description:              req.Description,
		BusinessElementRefsJSON:  encodeRefs(req.BusinessElementRefs),
		Requiredness:             req.Requiredness,
		FreshnessRequirement:     req.FreshnessRequirement,
		AccessNeed:               req.AccessNeed,
		Status:                   string(req.Status),
		Source:                   string(req.Source),
		DerivedFromRequirementID: req.DerivedFromRequirementID,
		AnalysisRunID:            req.AnalysisRunID,
		CreatedBy:                req.CreatedBy,
		CreatedAt:                req.CreatedAt,
		UpdatedAt:                req.UpdatedAt,
	}
}

func toAnalysisRun(r *analysisRunRow) *entity.DataRequirementAnalysisRun {
	return &entity.DataRequirementAnalysisRun{
		AnalysisRunID:         r.AnalysisRunID,
		TenantID:              r.TenantID,
		BusinessID:            r.BusinessID,
		BusinessModelRevision: r.BusinessModelRevision,
		ClientRequestID:       r.ClientRequestID,
		RequestDigest:         r.RequestDigest,
		Status:                entity.AnalysisRunStatus(r.Status),
		ModelRef:              r.ModelRef,
		ErrorKey:              r.ErrorKey,
		ErrorMessageSanitized: r.ErrorMessageSanitized,
		RetryCount:            r.RetryCount,
		CreatedBy:             r.CreatedBy,
		StartedAt:             r.StartedAt,
		CompletedAt:           r.CompletedAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

func isDup(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func (d *DataDAO) CreateRequirement(ctx context.Context, req *entity.DataRequirement) error {
	return d.db.WithContext(ctx).Create(fromRequirement(req)).Error
}

func (d *DataDAO) CreateRequirementsBatch(ctx context.Context, reqs []*entity.DataRequirement) error {
	if len(reqs) == 0 {
		return nil
	}
	rows := make([]*requirementRow, 0, len(reqs))
	for _, r := range reqs {
		rows = append(rows, fromRequirement(r))
	}
	return d.db.WithContext(ctx).Create(&rows).Error
}

func (d *DataDAO) GetRequirement(ctx context.Context, tenantID, requirementID string) (*entity.DataRequirement, error) {
	var row requirementRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND requirement_id = ?", tenantID, requirementID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrRequirementNotFound
	}
	if err != nil {
		return nil, err
	}
	return toRequirement(&row), nil
}

func (d *DataDAO) ListRequirementsByRevision(ctx context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error) {
	var rows []requirementRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ? AND business_model_revision = ?", tenantID, businessID, revision).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.DataRequirement, 0, len(rows))
	for i := range rows {
		out = append(out, toRequirement(&rows[i]))
	}
	return out, nil
}

func (d *DataDAO) UpdateRequirementStatusCAS(ctx context.Context, tenantID, requirementID string, from, to entity.RequirementStatus) (bool, error) {
	now := time.Now().UTC()
	res := d.db.WithContext(ctx).Model(&requirementRow{}).
		Where("tenant_id = ? AND requirement_id = ? AND status = ?", tenantID, requirementID, string(from)).
		Updates(map[string]any{"status": string(to), "updated_at": now})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (d *DataDAO) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.DataRequirementAnalysisRun) (*entity.DataRequirementAnalysisRun, bool, error) {
	existing, err := d.GetAnalysisRunByIdempotencyKey(ctx, run.TenantID, run.BusinessID, run.BusinessModelRevision, run.ClientRequestID)
	if err == nil {
		if existing.RequestDigest != run.RequestDigest {
			return nil, false, entity.ErrAnalysisIdempotencyConflict
		}
		return existing, false, nil
	}
	if !errors.Is(err, entity.ErrAnalysisNotFound) {
		return nil, false, err
	}

	row := &analysisRunRow{
		AnalysisRunID:         run.AnalysisRunID,
		TenantID:              run.TenantID,
		BusinessID:            run.BusinessID,
		BusinessModelRevision: run.BusinessModelRevision,
		ClientRequestID:       run.ClientRequestID,
		RequestDigest:         run.RequestDigest,
		Status:                string(run.Status),
		ModelRef:              run.ModelRef,
		CreatedBy:             run.CreatedBy,
		StartedAt:             run.StartedAt,
		CreatedAt:             run.CreatedAt,
		UpdatedAt:             run.UpdatedAt,
	}
	err = d.db.WithContext(ctx).Create(row).Error
	if err != nil {
		if isDup(err) {
			existing, gerr := d.GetAnalysisRunByIdempotencyKey(ctx, run.TenantID, run.BusinessID, run.BusinessModelRevision, run.ClientRequestID)
			if gerr != nil {
				return nil, false, gerr
			}
			if existing.RequestDigest != run.RequestDigest {
				return nil, false, entity.ErrAnalysisIdempotencyConflict
			}
			return existing, false, nil
		}
		return nil, false, err
	}
	return toAnalysisRun(row), true, nil
}

func (d *DataDAO) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error) {
	var row analysisRunRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND analysis_run_id = ?", tenantID, analysisRunID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrAnalysisNotFound
	}
	if err != nil {
		return nil, err
	}
	return toAnalysisRun(&row), nil
}

func (d *DataDAO) GetAnalysisRunByIdempotencyKey(ctx context.Context, tenantID, businessID string, revision int32, clientRequestID string) (*entity.DataRequirementAnalysisRun, error) {
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
	return toAnalysisRun(&row), nil
}

func (d *DataDAO) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string) error {
	now := time.Now().UTC()
	res := d.db.WithContext(ctx).Model(&analysisRunRow{}).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ?", tenantID, analysisRunID, string(entity.AnalysisPending)).
		Updates(map[string]any{
			"status":       string(entity.AnalysisSucceeded),
			"model_ref":    modelRef,
			"completed_at": now,
			"updated_at":   now,
			"error_key":    "",
			"error_message_sanitized": "",
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrRequirementInvalidState
	}
	return nil
}

func (d *DataDAO) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorKey, sanitizedMsg string) error {
	now := time.Now().UTC()
	if len(sanitizedMsg) > 1024 {
		sanitizedMsg = sanitizedMsg[:1024]
	}
	res := d.db.WithContext(ctx).Model(&analysisRunRow{}).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ?", tenantID, analysisRunID, string(entity.AnalysisPending)).
		Updates(map[string]any{
			"status":                  string(entity.AnalysisFailed),
			"error_key":               errorKey,
			"error_message_sanitized": sanitizedMsg,
			"completed_at":            now,
			"updated_at":              now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrRequirementInvalidState
	}
	return nil
}

func (d *DataDAO) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID string) (bool, error) {
	now := time.Now().UTC()
	res := d.db.WithContext(ctx).Model(&analysisRunRow{}).
		Where("tenant_id = ? AND analysis_run_id = ? AND status = ?", tenantID, analysisRunID, string(entity.AnalysisFailed)).
		Updates(map[string]any{
			"status":                  string(entity.AnalysisPending),
			"retry_count":             gorm.Expr("retry_count + 1"),
			"started_at":              now,
			"completed_at":            nil,
			"error_key":               "",
			"error_message_sanitized": "",
			"updated_at":              now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (d *DataDAO) CreateDecision(ctx context.Context, dec *entity.DataRequirementDecision) error {
	row := &decisionRow{
		DecisionID:            dec.DecisionID,
		TenantID:              dec.TenantID,
		BusinessID:            dec.BusinessID,
		SourceRequirementID:   dec.SourceRequirementID,
		TargetRequirementID:   dec.TargetRequirementID,
		Action:                string(dec.Action),
		ActorPrincipalID:      dec.ActorPrincipalID,
		Reason:                dec.Reason,
		BusinessModelRevision: dec.BusinessModelRevision,
		CreatedAt:             dec.CreatedAt,
	}
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: false}).Create(row).Error
	if err != nil {
		if isDup(err) {
			return entity.ErrRequirementAlreadyDecided
		}
		return err
	}
	return nil
}

func (d *DataDAO) GetDecisionBySource(ctx context.Context, tenantID, sourceRequirementID string) (*entity.DataRequirementDecision, error) {
	var row decisionRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND source_requirement_id = ?", tenantID, sourceRequirementID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrRequirementNotFound
	}
	if err != nil {
		return nil, err
	}
	return &entity.DataRequirementDecision{
		DecisionID:            row.DecisionID,
		TenantID:              row.TenantID,
		BusinessID:            row.BusinessID,
		SourceRequirementID:   row.SourceRequirementID,
		TargetRequirementID:   row.TargetRequirementID,
		Action:                entity.DecisionAction(row.Action),
		ActorPrincipalID:      row.ActorPrincipalID,
		Reason:                row.Reason,
		BusinessModelRevision: row.BusinessModelRevision,
		CreatedAt:             row.CreatedAt,
	}, nil
}

func (d *DataDAO) ListDecisionsByRequirement(ctx context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error) {
	var rows []decisionRow
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND (source_requirement_id = ? OR target_requirement_id = ?)", tenantID, requirementID, requirementID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.DataRequirementDecision, 0, len(rows))
	for _, row := range rows {
		out = append(out, &entity.DataRequirementDecision{
			DecisionID:            row.DecisionID,
			TenantID:              row.TenantID,
			BusinessID:            row.BusinessID,
			SourceRequirementID:   row.SourceRequirementID,
			TargetRequirementID:   row.TargetRequirementID,
			Action:                entity.DecisionAction(row.Action),
			ActorPrincipalID:      row.ActorPrincipalID,
			Reason:                row.Reason,
			BusinessModelRevision: row.BusinessModelRevision,
			CreatedAt:             row.CreatedAt,
		})
	}
	return out, nil
}
