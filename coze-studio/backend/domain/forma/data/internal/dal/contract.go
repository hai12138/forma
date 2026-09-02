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
	"gorm.io/gorm/clause"
)

type ContractDAO struct{ db *gorm.DB }

func NewContractDAO(db *gorm.DB) *ContractDAO { return &ContractDAO{db: db} }

type contractRow struct {
	ID                                                 int64 `gorm:"column:id;primaryKey"`
	ContractID, TenantID, BusinessID, ActiveRevisionID string
	CreatedBy                                          string
	CreatedAt, UpdatedAt                               time.Time
}

func (contractRow) TableName() string { return "forma_data_contract" }

type contractRevisionRow struct {
	ID                                                                                                                                                                                                                                                       int64 `gorm:"column:id;primaryKey"`
	RevisionID, TenantID, BusinessID, ContractID, Status, Name, Description, RequirementIDsJSON, LogicalSchemaJSON, QueryCapabilitiesJSON, FilterSchemaJSON, SortSchemaJSON, PaginationPolicyJSON, FreshnessPolicy, ClassificationPolicyJSON, BindingRefsJSON string
	AccessPolicyRef, DerivedFromRevisionID, CreatedBy                                                                                                                                                                                                        string
	Version, BusinessModelRevision                                                                                                                                                                                                                           int32
	CreatedAt, UpdatedAt                                                                                                                                                                                                                                     time.Time
}

func (contractRevisionRow) TableName() string { return "forma_data_contract_revision" }

type contractValidationRow struct {
	ID                                                                                                                         int64 `gorm:"column:id;primaryKey"`
	ValidationID, TenantID, BusinessID, ContractID, RevisionID, Status, ErrorsJSON, WarningsJSON, SnapshotFingerprintsJSON, ValidatedBy string
	Version                                                                                                                    int32
	ValidatedAt, CreatedAt                                                                                                     time.Time
}

func (contractValidationRow) TableName() string { return "forma_data_contract_validation_result" }

type contractLifecycleRow struct {
	ID                                                                                 int64 `gorm:"column:id;primaryKey"`
	EventID, TenantID, BusinessID, ContractID, RevisionID, Action, ActorPrincipalID, Reason string
	Version                                                                            int32
	CreatedAt                                                                          time.Time
}

func (contractLifecycleRow) TableName() string { return "forma_data_contract_lifecycle_event" }

type contractDriftRow struct {
	ID                                                                                                           int64 `gorm:"column:id;primaryKey"`
	DriftResultID, TenantID, BusinessID, ContractID, RevisionID, Severity, FindingsJSON, ComparedSnapshotIDsJSON, EvaluatedBy string
	Version                                                                                                      int32
	EvaluatedAt, CreatedAt                                                                                       time.Time
}

func (contractDriftRow) TableName() string { return "forma_data_contract_drift_result" }

type contractGapRow struct {
	ID                                                                                                                                               int64 `gorm:"column:id;primaryKey"`
	GapResultID, TenantID, BusinessID, ContractID, RevisionID, NewConfirmedRequirementIDsJSON, UnmappedRequirementIDsJSON, GapStatus, EvaluatedBy string
	Version, FromBusinessRevision, CurrentBusinessRevision                                                                                           int32
	EvaluatedAt, CreatedAt                                                                                                                           time.Time
}

func (contractGapRow) TableName() string { return "forma_data_contract_gap_result" }

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	if len(b) == 0 {
		return "null"
	}
	return string(b)
}

func contractFrom(v *entity.DataContract) *contractRow {
	return &contractRow{
		ContractID: v.ContractID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ActiveRevisionID: v.ActiveRevisionID, CreatedBy: v.CreatedBy,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}
func contractTo(v *contractRow) *entity.DataContract {
	return &entity.DataContract{
		ContractID: v.ContractID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ActiveRevisionID: v.ActiveRevisionID, CreatedBy: v.CreatedBy,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}

func revisionFrom(v *entity.DataContractRevision) *contractRevisionRow {
	classPolicy := v.ClassificationPolicy
	if classPolicy == nil {
		classPolicy = map[string]entity.DataClassification{}
	}
	return &contractRevisionRow{
		RevisionID: v.RevisionID, TenantID: v.TenantID, BusinessID: v.BusinessID, ContractID: v.ContractID,
		Version: v.Version, Status: string(v.Status), BusinessModelRevision: v.BusinessModelRevision,
		Name: v.Name, Description: v.Description,
		RequirementIDsJSON: mustJSON(v.RequirementIDs), LogicalSchemaJSON: mustJSON(v.LogicalSchema),
		QueryCapabilitiesJSON: mustJSON(v.QueryCapabilities), FilterSchemaJSON: mustJSON(v.FilterSchema),
		SortSchemaJSON: mustJSON(v.SortSchema), PaginationPolicyJSON: mustJSON(v.PaginationPolicy),
		FreshnessPolicy: string(v.FreshnessPolicy), ClassificationPolicyJSON: mustJSON(classPolicy),
		BindingRefsJSON: mustJSON(v.BindingRefs), AccessPolicyRef: v.AccessPolicyRef,
		DerivedFromRevisionID: v.DerivedFromRevisionID, CreatedBy: v.CreatedBy,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}
func revisionTo(v *contractRevisionRow) *entity.DataContractRevision {
	out := &entity.DataContractRevision{
		RevisionID: v.RevisionID, TenantID: v.TenantID, BusinessID: v.BusinessID, ContractID: v.ContractID,
		Version: v.Version, Status: entity.ContractStatus(v.Status), BusinessModelRevision: v.BusinessModelRevision,
		Name: v.Name, Description: v.Description, FreshnessPolicy: entity.FreshnessPolicy(v.FreshnessPolicy),
		AccessPolicyRef: v.AccessPolicyRef, DerivedFromRevisionID: v.DerivedFromRevisionID, CreatedBy: v.CreatedBy,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		ClassificationPolicy: map[string]entity.DataClassification{},
	}
	_ = json.Unmarshal([]byte(v.RequirementIDsJSON), &out.RequirementIDs)
	_ = json.Unmarshal([]byte(v.LogicalSchemaJSON), &out.LogicalSchema)
	_ = json.Unmarshal([]byte(v.QueryCapabilitiesJSON), &out.QueryCapabilities)
	_ = json.Unmarshal([]byte(v.FilterSchemaJSON), &out.FilterSchema)
	_ = json.Unmarshal([]byte(v.SortSchemaJSON), &out.SortSchema)
	_ = json.Unmarshal([]byte(v.PaginationPolicyJSON), &out.PaginationPolicy)
	_ = json.Unmarshal([]byte(v.ClassificationPolicyJSON), &out.ClassificationPolicy)
	_ = json.Unmarshal([]byte(v.BindingRefsJSON), &out.BindingRefs)
	if out.ClassificationPolicy == nil {
		out.ClassificationPolicy = map[string]entity.DataClassification{}
	}
	return out
}

func validationFrom(v *entity.DataValidationResult) *contractValidationRow {
	return &contractValidationRow{
		ValidationID: v.ValidationID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ContractID: v.ContractID, RevisionID: v.RevisionID, Version: v.Version, Status: string(v.Status),
		ErrorsJSON: mustJSON(v.Errors), WarningsJSON: mustJSON(v.Warnings),
		SnapshotFingerprintsJSON: mustJSON(v.SnapshotFingerprints), ValidatedBy: v.ValidatedBy,
		ValidatedAt: v.ValidatedAt, CreatedAt: v.CreatedAt,
	}
}
func validationTo(v *contractValidationRow) *entity.DataValidationResult {
	out := &entity.DataValidationResult{
		ValidationID: v.ValidationID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ContractID: v.ContractID, RevisionID: v.RevisionID, Version: v.Version,
		Status: entity.ValidationStatus(v.Status), ValidatedBy: v.ValidatedBy,
		ValidatedAt: v.ValidatedAt, CreatedAt: v.CreatedAt,
		SnapshotFingerprints: map[string]string{},
	}
	_ = json.Unmarshal([]byte(v.ErrorsJSON), &out.Errors)
	_ = json.Unmarshal([]byte(v.WarningsJSON), &out.Warnings)
	_ = json.Unmarshal([]byte(v.SnapshotFingerprintsJSON), &out.SnapshotFingerprints)
	if out.SnapshotFingerprints == nil {
		out.SnapshotFingerprints = map[string]string{}
	}
	return out
}

func lifecycleFrom(v *entity.DataContractLifecycleEvent) *contractLifecycleRow {
	return &contractLifecycleRow{
		EventID: v.EventID, TenantID: v.TenantID, BusinessID: v.BusinessID, ContractID: v.ContractID,
		RevisionID: v.RevisionID, Version: v.Version, Action: string(v.Action),
		ActorPrincipalID: v.ActorPrincipalID, Reason: v.Reason, CreatedAt: v.CreatedAt,
	}
}
func lifecycleTo(v *contractLifecycleRow) *entity.DataContractLifecycleEvent {
	return &entity.DataContractLifecycleEvent{
		EventID: v.EventID, TenantID: v.TenantID, BusinessID: v.BusinessID, ContractID: v.ContractID,
		RevisionID: v.RevisionID, Version: v.Version, Action: entity.LifecycleAction(v.Action),
		ActorPrincipalID: v.ActorPrincipalID, Reason: v.Reason, CreatedAt: v.CreatedAt,
	}
}

func driftFrom(v *entity.DataDriftResult) *contractDriftRow {
	return &contractDriftRow{
		DriftResultID: v.DriftResultID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ContractID: v.ContractID, RevisionID: v.RevisionID, Version: v.Version, Severity: string(v.Severity),
		FindingsJSON: mustJSON(v.Findings), ComparedSnapshotIDsJSON: mustJSON(v.ComparedSnapshotIDs),
		EvaluatedBy: v.EvaluatedBy, EvaluatedAt: v.EvaluatedAt, CreatedAt: v.CreatedAt,
	}
}
func driftTo(v *contractDriftRow) *entity.DataDriftResult {
	out := &entity.DataDriftResult{
		DriftResultID: v.DriftResultID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ContractID: v.ContractID, RevisionID: v.RevisionID, Version: v.Version,
		Severity: entity.DriftSeverity(v.Severity), EvaluatedBy: v.EvaluatedBy,
		EvaluatedAt: v.EvaluatedAt, CreatedAt: v.CreatedAt, ComparedSnapshotIDs: map[string]string{},
	}
	_ = json.Unmarshal([]byte(v.FindingsJSON), &out.Findings)
	_ = json.Unmarshal([]byte(v.ComparedSnapshotIDsJSON), &out.ComparedSnapshotIDs)
	if out.ComparedSnapshotIDs == nil {
		out.ComparedSnapshotIDs = map[string]string{}
	}
	return out
}

func gapFrom(v *entity.DataContractGapResult) *contractGapRow {
	return &contractGapRow{
		GapResultID: v.GapResultID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ContractID: v.ContractID, RevisionID: v.RevisionID, Version: v.Version,
		FromBusinessRevision: v.FromBusinessRevision, CurrentBusinessRevision: v.CurrentBusinessRevision,
		NewConfirmedRequirementIDsJSON: mustJSON(v.NewConfirmedRequirementIDs),
		UnmappedRequirementIDsJSON:     mustJSON(v.UnmappedRequirementIDs),
		GapStatus: v.GapStatus, EvaluatedBy: v.EvaluatedBy, EvaluatedAt: v.EvaluatedAt, CreatedAt: v.CreatedAt,
	}
}
func gapTo(v *contractGapRow) *entity.DataContractGapResult {
	out := &entity.DataContractGapResult{
		GapResultID: v.GapResultID, TenantID: v.TenantID, BusinessID: v.BusinessID,
		ContractID: v.ContractID, RevisionID: v.RevisionID, Version: v.Version,
		FromBusinessRevision: v.FromBusinessRevision, CurrentBusinessRevision: v.CurrentBusinessRevision,
		GapStatus: v.GapStatus, EvaluatedBy: v.EvaluatedBy, EvaluatedAt: v.EvaluatedAt, CreatedAt: v.CreatedAt,
	}
	_ = json.Unmarshal([]byte(v.NewConfirmedRequirementIDsJSON), &out.NewConfirmedRequirementIDs)
	_ = json.Unmarshal([]byte(v.UnmappedRequirementIDsJSON), &out.UnmappedRequirementIDs)
	return out
}

func (d *ContractDAO) CreateContract(c context.Context, v *entity.DataContract) error {
	return d.db.WithContext(c).Create(contractFrom(v)).Error
}
func (d *ContractDAO) GetContract(c context.Context, t, id string) (*entity.DataContract, error) {
	var row contractRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND contract_id = ?", t, id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrContractNotFound
	}
	if err != nil {
		return nil, err
	}
	return contractTo(&row), nil
}
func (d *ContractDAO) ListContracts(c context.Context, t, b string) ([]*entity.DataContract, error) {
	var rows []contractRow
	if err := d.db.WithContext(c).Where("tenant_id = ? AND business_id = ?", t, b).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.DataContract, 0, len(rows))
	for i := range rows {
		out = append(out, contractTo(&rows[i]))
	}
	return out, nil
}
func (d *ContractDAO) UpdateContractActiveRevision(c context.Context, t, contractID, revisionID string) error {
	res := d.db.WithContext(c).Model(&contractRow{}).
		Where("tenant_id = ? AND contract_id = ?", t, contractID).
		Updates(map[string]any{"active_revision_id": revisionID, "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrContractNotFound
	}
	return nil
}

func (d *ContractDAO) CreateRevision(c context.Context, v *entity.DataContractRevision) error {
	if err := d.db.WithContext(c).Create(revisionFrom(v)).Error; err != nil {
		if isDup(err) {
			return entity.ErrContractVersionConflict
		}
		return err
	}
	return nil
}
func (d *ContractDAO) GetRevision(c context.Context, t, id string) (*entity.DataContractRevision, error) {
	var row contractRevisionRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND revision_id = ?", t, id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrContractRevisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return revisionTo(&row), nil
}
func (d *ContractDAO) ListRevisions(c context.Context, t, contractID string) ([]*entity.DataContractRevision, error) {
	var rows []contractRevisionRow
	if err := d.db.WithContext(c).Where("tenant_id = ? AND contract_id = ?", t, contractID).Order("version ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.DataContractRevision, 0, len(rows))
	for i := range rows {
		out = append(out, revisionTo(&rows[i]))
	}
	return out, nil
}
func (d *ContractDAO) GetRevisionByVersion(c context.Context, t, contractID string, version int32) (*entity.DataContractRevision, error) {
	var row contractRevisionRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND contract_id = ? AND version = ?", t, contractID, version).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrContractRevisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return revisionTo(&row), nil
}
func (d *ContractDAO) AllocateNextVersion(c context.Context, t, contractID string) (int32, error) {
	var row contractRow
	err := d.db.WithContext(c).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND contract_id = ?", t, contractID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, entity.ErrContractNotFound
	}
	if err != nil {
		return 0, err
	}
	var maxVersion int32
	if err := d.db.WithContext(c).Model(&contractRevisionRow{}).
		Where("tenant_id = ? AND contract_id = ?", t, contractID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error; err != nil {
		return 0, err
	}
	return maxVersion + 1, nil
}
func (d *ContractDAO) UpdateRevisionStatus(c context.Context, t, revisionID string, from, to entity.ContractStatus) error {
	res := d.db.WithContext(c).Model(&contractRevisionRow{}).
		Where("tenant_id = ? AND revision_id = ? AND status = ?", t, revisionID, string(from)).
		Updates(map[string]any{"status": string(to), "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return entity.ErrContractInvalidState
	}
	return nil
}

func (d *ContractDAO) CreateValidationResult(c context.Context, v *entity.DataValidationResult) error {
	return d.db.WithContext(c).Create(validationFrom(v)).Error
}
func (d *ContractDAO) ListValidationResults(c context.Context, t, revisionID string) ([]*entity.DataValidationResult, error) {
	var rows []contractValidationRow
	if err := d.db.WithContext(c).Where("tenant_id = ? AND revision_id = ?", t, revisionID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.DataValidationResult, 0, len(rows))
	for i := range rows {
		out = append(out, validationTo(&rows[i]))
	}
	return out, nil
}
func (d *ContractDAO) GetValidationResult(c context.Context, t, id string) (*entity.DataValidationResult, error) {
	var row contractValidationRow
	err := d.db.WithContext(c).Where("tenant_id = ? AND validation_id = ?", t, id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entity.ErrContractValidationFailed
	}
	if err != nil {
		return nil, err
	}
	return validationTo(&row), nil
}

func (d *ContractDAO) CreateLifecycleEvent(c context.Context, v *entity.DataContractLifecycleEvent) error {
	return d.db.WithContext(c).Create(lifecycleFrom(v)).Error
}
func (d *ContractDAO) ListLifecycleEvents(c context.Context, t, contractID string) ([]*entity.DataContractLifecycleEvent, error) {
	var rows []contractLifecycleRow
	if err := d.db.WithContext(c).Where("tenant_id = ? AND contract_id = ?", t, contractID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.DataContractLifecycleEvent, 0, len(rows))
	for i := range rows {
		out = append(out, lifecycleTo(&rows[i]))
	}
	return out, nil
}

func (d *ContractDAO) CreateDriftResult(c context.Context, v *entity.DataDriftResult) error {
	return d.db.WithContext(c).Create(driftFrom(v)).Error
}
func (d *ContractDAO) ListDriftResults(c context.Context, t, revisionID string) ([]*entity.DataDriftResult, error) {
	var rows []contractDriftRow
	if err := d.db.WithContext(c).Where("tenant_id = ? AND revision_id = ?", t, revisionID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.DataDriftResult, 0, len(rows))
	for i := range rows {
		out = append(out, driftTo(&rows[i]))
	}
	return out, nil
}

func (d *ContractDAO) CreateGapResult(c context.Context, v *entity.DataContractGapResult) error {
	return d.db.WithContext(c).Create(gapFrom(v)).Error
}
func (d *ContractDAO) ListGapResults(c context.Context, t, revisionID string) ([]*entity.DataContractGapResult, error) {
	var rows []contractGapRow
	if err := d.db.WithContext(c).Where("tenant_id = ? AND revision_id = ?", t, revisionID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.DataContractGapResult, 0, len(rows))
	for i := range rows {
		out = append(out, gapTo(&rows[i]))
	}
	return out, nil
}
