/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"gorm.io/gorm"
)

type MappingDAO struct{ db *gorm.DB }

func NewMappingDAO(db *gorm.DB) *MappingDAO { return &MappingDAO{db: db} }

type mappingRow struct {
	ID                                                                                                                                                                                                                              int64 `gorm:"column:id;primaryKey"`
	MappingID, TenantID, BusinessID, RequirementID, SourceID, ConnectionID, AssetID, SchemaSnapshotID, TargetFieldPathsJSON, MappingType, TransformSpecJSON, Status, Source, Reason, DerivedFromMappingID, AnalysisRunID, CreatedBy string
	BusinessModelRevision                                                                                                                                                                                                           int32
	Confidence                                                                                                                                                                                                                      float64
	CreatedAt, UpdatedAt                                                                                                                                                                                                            time.Time
}

func (mappingRow) TableName() string { return "forma_data_semantic_mapping" }

type mappingAnalysisRow struct {
	ID                                                                                                                                                          int64 `gorm:"column:id;primaryKey"`
	AnalysisRunID, TenantID, BusinessID, ClientRequestID, RequestDigest, RequestJSON, Status, ModelRef, ErrorKey, ErrorMessageSanitized, LastRetryBy, CreatedBy string
	BusinessModelRevision, RetryCount, ExecutionGeneration                                                                                                      int32
	LastRetryAt, ExecutionClaimedAt, LeaseExpiresAt, StartedAt, CompletedAt                                                                                     *time.Time
	CreatedAt, UpdatedAt                                                                                                                                        time.Time
}

func (mappingAnalysisRow) TableName() string { return "forma_data_semantic_mapping_analysis_run" }

type mappingDecisionRow struct {
	ID                                                                                                   int64 `gorm:"column:id;primaryKey"`
	DecisionID, TenantID, BusinessID, SourceMappingID, TargetMappingID, Action, ActorPrincipalID, Reason string
	BusinessModelRevision                                                                                int32
	CreatedAt                                                                                            time.Time
}

func (mappingDecisionRow) TableName() string { return "forma_data_semantic_mapping_decision" }

func mappingFrom(v *entity.SemanticMapping) *mappingRow {
	paths, _ := json.Marshal(v.TargetFieldPaths)
	return &mappingRow{MappingID: v.MappingID, TenantID: v.TenantID, BusinessID: v.BusinessID, BusinessModelRevision: v.BusinessModelRevision, RequirementID: v.RequirementID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetID: v.AssetID, SchemaSnapshotID: v.SchemaSnapshotID, TargetFieldPathsJSON: string(paths), MappingType: string(v.MappingType), TransformSpecJSON: string(v.TransformSpec), Status: string(v.Status), Source: string(v.Source), Confidence: v.Confidence, Reason: v.Reason, DerivedFromMappingID: v.DerivedFromMappingID, AnalysisRunID: v.AnalysisRunID, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func mappingTo(v *mappingRow) *entity.SemanticMapping {
	var paths []string
	_ = json.Unmarshal([]byte(v.TargetFieldPathsJSON), &paths)
	return &entity.SemanticMapping{MappingID: v.MappingID, TenantID: v.TenantID, BusinessID: v.BusinessID, BusinessModelRevision: v.BusinessModelRevision, RequirementID: v.RequirementID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetID: v.AssetID, SchemaSnapshotID: v.SchemaSnapshotID, TargetFieldPaths: paths, MappingType: entity.MappingType(v.MappingType), TransformSpec: json.RawMessage(v.TransformSpecJSON), Status: entity.MappingStatus(v.Status), Source: entity.MappingSource(v.Source), Confidence: v.Confidence, Reason: v.Reason, DerivedFromMappingID: v.DerivedFromMappingID, AnalysisRunID: v.AnalysisRunID, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func analysisFrom(v *entity.SemanticMappingAnalysisRun) *mappingAnalysisRow {
	return &mappingAnalysisRow{AnalysisRunID: v.AnalysisRunID, TenantID: v.TenantID, BusinessID: v.BusinessID, BusinessModelRevision: v.BusinessModelRevision, ClientRequestID: v.ClientRequestID, RequestDigest: v.RequestDigest, RequestJSON: v.RequestJSON, Status: string(v.Status), ModelRef: v.ModelRef, ErrorKey: v.ErrorKey, ErrorMessageSanitized: v.ErrorMessageSanitized, RetryCount: v.RetryCount, LastRetryBy: v.LastRetryBy, LastRetryAt: v.LastRetryAt, ExecutionGeneration: v.ExecutionGeneration, ExecutionClaimedAt: v.ExecutionClaimedAt, LeaseExpiresAt: v.LeaseExpiresAt, CreatedBy: v.CreatedBy, StartedAt: v.StartedAt, CompletedAt: v.CompletedAt, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func analysisTo(v *mappingAnalysisRow) *entity.SemanticMappingAnalysisRun {
	return &entity.SemanticMappingAnalysisRun{AnalysisRunID: v.AnalysisRunID, TenantID: v.TenantID, BusinessID: v.BusinessID, BusinessModelRevision: v.BusinessModelRevision, ClientRequestID: v.ClientRequestID, RequestDigest: v.RequestDigest, RequestJSON: v.RequestJSON, Status: entity.AnalysisRunStatus(v.Status), ModelRef: v.ModelRef, ErrorKey: v.ErrorKey, ErrorMessageSanitized: v.ErrorMessageSanitized, RetryCount: v.RetryCount, LastRetryBy: v.LastRetryBy, LastRetryAt: v.LastRetryAt, ExecutionGeneration: v.ExecutionGeneration, ExecutionClaimedAt: v.ExecutionClaimedAt, LeaseExpiresAt: v.LeaseExpiresAt, CreatedBy: v.CreatedBy, StartedAt: v.StartedAt, CompletedAt: v.CompletedAt, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func (d *MappingDAO) CreateMapping(c context.Context, v *entity.SemanticMapping) error {
	return d.db.WithContext(c).Create(mappingFrom(v)).Error
}
func (d *MappingDAO) CreateMappingsBatch(c context.Context, vs []*entity.SemanticMapping) error {
	if len(vs) == 0 {
		return nil
	}
	rows := make([]*mappingRow, 0, len(vs))
	for _, v := range vs {
		rows = append(rows, mappingFrom(v))
	}
	return d.db.WithContext(c).Create(&rows).Error
}
func (d *MappingDAO) GetMapping(c context.Context, t, id string) (*entity.SemanticMapping, error) {
	var row mappingRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND mapping_id = ?", t, id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrMappingNotFound
	}
	if err != nil {
		return nil, err
	}
	return mappingTo(&row), nil
}
func (d *MappingDAO) ListMappings(c context.Context, t, b string, rev int32, s entity.MappingStatus) ([]*entity.SemanticMapping, error) {
	var rows []mappingRow
	q := d.db.WithContext(c).Where("tenant_id = ? AND business_id = ? AND business_model_revision = ?", t, b, rev)
	if s != "" {
		q = q.Where("status = ?", string(s))
	}
	if err := q.Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.SemanticMapping, 0, len(rows))
	for i := range rows {
		out = append(out, mappingTo(&rows[i]))
	}
	return out, nil
}
func (d *MappingDAO) UpdateMappingStatusCAS(c context.Context, t, id string, f, to entity.MappingStatus) (bool, error) {
	res := d.db.WithContext(c).Model(&mappingRow{}).Where("tenant_id = ? AND mapping_id = ? AND status = ?", t, id, string(f)).Updates(map[string]any{"status": string(to), "updated_at": time.Now().UTC()})
	if to == entity.MappingStatusConfirmed && isDup(res.Error) {
		return false, entity.ErrMappingAlreadyConfirmed
	}
	return res.RowsAffected == 1, res.Error
}
func (d *MappingDAO) GetConfirmedMappingByRequirement(c context.Context, t, b string, rev int32, req string) (*entity.SemanticMapping, error) {
	var row mappingRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND business_id = ? AND business_model_revision = ? AND requirement_id = ? AND status = ?", t, b, rev, req, string(entity.MappingStatusConfirmed)).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrMappingNotFound
	}
	if err != nil {
		return nil, err
	}
	return mappingTo(&row), nil
}
func (d *MappingDAO) CreateOrClaimAnalysisRun(c context.Context, v *entity.SemanticMappingAnalysisRun) (*entity.SemanticMappingAnalysisRun, bool, error) {
	var row mappingAnalysisRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND business_id = ? AND business_model_revision = ? AND client_request_id = ?", v.TenantID, v.BusinessID, v.BusinessModelRevision, v.ClientRequestID).First(&row).Error
	if err == nil {
		if row.RequestDigest != v.RequestDigest {
			return nil, false, entity.ErrMappingAnalysisIdempotencyConflict
		}
		return analysisTo(&row), false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	row = *analysisFrom(v)
	if row.ExecutionGeneration == 0 {
		row.ExecutionGeneration = 1
	}
	if err = d.db.WithContext(c).Create(&row).Error; err != nil {
		if isDup(err) {
			return d.CreateOrClaimAnalysisRun(c, v)
		}
		return nil, false, err
	}
	return analysisTo(&row), true, nil
}
func (d *MappingDAO) GetAnalysisRun(c context.Context, t, id string) (*entity.SemanticMappingAnalysisRun, error) {
	var row mappingAnalysisRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND analysis_run_id = ?", t, id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrMappingAnalysisNotFound
	}
	if err != nil {
		return nil, err
	}
	return analysisTo(&row), nil
}
func (d *MappingDAO) mark(c context.Context, t, id string, status entity.AnalysisRunStatus, values map[string]any, g int32) error {
	values["status"] = string(status)
	values["completed_at"] = time.Now().UTC()
	values["updated_at"] = time.Now().UTC()
	res := d.db.WithContext(c).Model(&mappingAnalysisRow{}).Where("tenant_id = ? AND analysis_run_id = ? AND status = ? AND execution_generation = ?", t, id, string(entity.AnalysisPending), g).Updates(values)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrMappingInvalidState
	}
	return nil
}
func (d *MappingDAO) MarkAnalysisSucceeded(c context.Context, t, id, m string, g int32) error {
	return d.mark(c, t, id, entity.AnalysisSucceeded, map[string]any{"model_ref": m, "error_key": "", "error_message_sanitized": ""}, g)
}
func (d *MappingDAO) MarkAnalysisFailed(c context.Context, t, id, k, msg string, g int32) error {
	return d.mark(c, t, id, entity.AnalysisFailed, map[string]any{"error_key": k, "error_message_sanitized": msg}, g)
}
func (d *MappingDAO) ClaimAnalysisRetry(c context.Context, t, id, a string) (bool, int32, error) {
	run, err := d.GetAnalysisRun(c, t, id)
	if err != nil {
		return false, 0, err
	}
	if run.Status != entity.AnalysisFailed {
		return false, 0, nil
	}
	now := time.Now().UTC()
	exp := now.Add(5 * time.Minute)
	g := run.ExecutionGeneration + 1
	res := d.db.WithContext(c).Model(&mappingAnalysisRow{}).Where("tenant_id = ? AND analysis_run_id = ? AND status = ? AND execution_generation = ?", t, id, string(entity.AnalysisFailed), run.ExecutionGeneration).Updates(map[string]any{"status": string(entity.AnalysisPending), "retry_count": gorm.Expr("retry_count + 1"), "last_retry_by": a, "last_retry_at": now, "execution_generation": g, "execution_claimed_at": now, "lease_expires_at": exp, "started_at": now, "completed_at": nil, "updated_at": now})
	return res.RowsAffected == 1, g, res.Error
}
func (d *MappingDAO) ClaimExpiredAnalysis(c context.Context, t, id string, g int32, now time.Time) (*entity.SemanticMappingAnalysisRun, bool, error) {
	exp := now.Add(5 * time.Minute)
	res := d.db.WithContext(c).Model(&mappingAnalysisRow{}).Where("tenant_id = ? AND analysis_run_id = ? AND status = ? AND execution_generation = ? AND lease_expires_at <= ?", t, id, string(entity.AnalysisPending), g, now).Updates(map[string]any{"execution_generation": g + 1, "execution_claimed_at": now, "lease_expires_at": exp, "updated_at": now})
	run, err := d.GetAnalysisRun(c, t, id)
	return run, res.RowsAffected == 1, err
}
func (d *MappingDAO) CreateDecision(c context.Context, v *entity.SemanticMappingDecision) error {
	row := &mappingDecisionRow{DecisionID: v.DecisionID, TenantID: v.TenantID, BusinessID: v.BusinessID, SourceMappingID: v.SourceMappingID, TargetMappingID: v.TargetMappingID, Action: string(v.Action), ActorPrincipalID: v.ActorPrincipalID, Reason: v.Reason, BusinessModelRevision: v.BusinessModelRevision, CreatedAt: v.CreatedAt}
	if err := d.db.WithContext(c).Create(row).Error; err != nil {
		if isDup(err) {
			return entity.ErrMappingAlreadyDecided
		}
		return err
	}
	return nil
}
func (d *MappingDAO) ListDecisions(c context.Context, t, id string) ([]*entity.SemanticMappingDecision, error) {
	var rows []mappingDecisionRow
	if err := d.db.WithContext(c).Where("tenant_id = ? AND (source_mapping_id = ? OR target_mapping_id = ?)", t, id, id).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.SemanticMappingDecision, 0, len(rows))
	for _, v := range rows {
		out = append(out, &entity.SemanticMappingDecision{DecisionID: v.DecisionID, TenantID: v.TenantID, BusinessID: v.BusinessID, SourceMappingID: v.SourceMappingID, TargetMappingID: v.TargetMappingID, Action: entity.DecisionAction(v.Action), ActorPrincipalID: v.ActorPrincipalID, Reason: v.Reason, BusinessModelRevision: v.BusinessModelRevision, CreatedAt: v.CreatedAt})
	}
	return out, nil
}
