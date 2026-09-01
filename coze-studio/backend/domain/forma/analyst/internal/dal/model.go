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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
)

type SessionModel struct {
	ID                     int64      `gorm:"column:id;primaryKey;autoIncrement"`
	SessionID              string     `gorm:"column:session_id"`
	TenantID               string     `gorm:"column:tenant_id"`
	BusinessID             string     `gorm:"column:business_id"`
	Status                 string     `gorm:"column:status"`
	Title                  string     `gorm:"column:title"`
	RuntimeConversationRef string     `gorm:"column:runtime_conversation_ref"`
	ConfirmationPolicy     string     `gorm:"column:confirmation_policy"`
	NextTurnSequence       int32      `gorm:"column:next_turn_sequence"`
	CreatedBy              string     `gorm:"column:created_by"`
	CreatedAt              time.Time  `gorm:"column:created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at"`
	ClosedAt               *time.Time `gorm:"column:closed_at"`
}

func (SessionModel) TableName() string { return "forma_analyst_session" }

type TurnModel struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TurnID          string    `gorm:"column:turn_id"`
	TenantID        string    `gorm:"column:tenant_id"`
	SessionID       string    `gorm:"column:session_id"`
	Sequence        int32     `gorm:"column:sequence"`
	Speaker         string    `gorm:"column:speaker"`
	Content         string    `gorm:"column:content"`
	ContentType     string    `gorm:"column:content_type"`
	ClientRequestID       string    `gorm:"column:client_request_id"`
	ModelRequestID      string    `gorm:"column:model_request_id"`
	ReplyToTurnID         string    `gorm:"column:reply_to_turn_id"`
	ReservedReplySequence int32     `gorm:"column:reserved_reply_sequence"`
	AnalysisStatus        string    `gorm:"column:analysis_status"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (TurnModel) TableName() string { return "forma_analyst_turn" }

type EvidenceModel struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	EvidenceID    string    `gorm:"column:evidence_id"`
	TenantID      string    `gorm:"column:tenant_id"`
	BusinessID    string    `gorm:"column:business_id"`
	SessionID     string    `gorm:"column:session_id"`
	TurnID        string    `gorm:"column:turn_id"`
	SourceType    string    `gorm:"column:source_type"`
	SourceRef     string    `gorm:"column:source_ref"`
	Quote         string    `gorm:"column:quote"`
	ContentDigest string    `gorm:"column:content_digest"`
	CreatedBy     string    `gorm:"column:created_by"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (EvidenceModel) TableName() string { return "forma_business_evidence" }

type AssertionModel struct {
	ID                     int64     `gorm:"column:id;primaryKey;autoIncrement"`
	AssertionID            string    `gorm:"column:assertion_id"`
	TenantID               string    `gorm:"column:tenant_id"`
	BusinessID             string    `gorm:"column:business_id"`
	SessionID              string    `gorm:"column:session_id"`
	AssertionType          string    `gorm:"column:assertion_type"`
	SubjectRef             string    `gorm:"column:subject_ref"`
	Predicate              string    `gorm:"column:predicate"`
	ObjectValue            string    `gorm:"column:object_value"`
	StructuredValueJSON    string    `gorm:"column:structured_value_json"`
	Confidence             float64   `gorm:"column:confidence"`
	Status                 string    `gorm:"column:status"`
	SourceMarker           string    `gorm:"column:source_marker"`
	DerivedFromAssertionID string    `gorm:"column:derived_from_assertion_id"`
	CreatedBy              string    `gorm:"column:created_by"`
	CreatedAt              time.Time `gorm:"column:created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at"`
}

func (AssertionModel) TableName() string { return "forma_business_assertion" }

type AssertionEvidenceRefModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID    string    `gorm:"column:tenant_id"`
	AssertionID string    `gorm:"column:assertion_id"`
	EvidenceID  string    `gorm:"column:evidence_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (AssertionEvidenceRefModel) TableName() string { return "forma_assertion_evidence_ref" }

type ConfirmationModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ConfirmationID string    `gorm:"column:confirmation_id"`
	TenantID       string    `gorm:"column:tenant_id"`
	BusinessID     string    `gorm:"column:business_id"`
	AssertionID    string    `gorm:"column:assertion_id"`
	Decision       string    `gorm:"column:decision"`
	Comment        string    `gorm:"column:comment"`
	DecidedBy      string    `gorm:"column:decided_by"`
	DecidedAt      time.Time `gorm:"column:decided_at"`
}

func (ConfirmationModel) TableName() string { return "forma_business_confirmation" }

type ConflictModel struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ConflictID   string    `gorm:"column:conflict_id"`
	TenantID     string    `gorm:"column:tenant_id"`
	BusinessID   string    `gorm:"column:business_id"`
	SessionID    string    `gorm:"column:session_id"`
	AssertionIDA string    `gorm:"column:assertion_id_a"`
	AssertionIDB string    `gorm:"column:assertion_id_b"`
	SubjectRef   string    `gorm:"column:subject_ref"`
	Predicate    string    `gorm:"column:predicate"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (ConflictModel) TableName() string { return "forma_assertion_conflict" }

type GapModel struct {
	ID                       int64     `gorm:"column:id;primaryKey;autoIncrement"`
	GapID                    string    `gorm:"column:gap_id"`
	TenantID                 string    `gorm:"column:tenant_id"`
	BusinessID               string    `gorm:"column:business_id"`
	SessionID                string    `gorm:"column:session_id"`
	GapType                  string    `gorm:"column:gap_type"`
	Question                 string    `gorm:"column:question"`
	RelatedAssertionIDsJSON  string    `gorm:"column:related_assertion_ids_json"`
	Status                   string    `gorm:"column:status"`
	CreatedAt                time.Time `gorm:"column:created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at"`
}

func (GapModel) TableName() string { return "forma_analyst_gap" }

type ProposalModel struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ProposalID       string    `gorm:"column:proposal_id"`
	TenantID         string    `gorm:"column:tenant_id"`
	BusinessID       string    `gorm:"column:business_id"`
	SessionID        string    `gorm:"column:session_id"`
	BaseRevision     int32     `gorm:"column:base_revision"`
	AssertionIDsJSON string    `gorm:"column:assertion_ids_json"`
	PatchJSON        string    `gorm:"column:patch_json"`
	Status           string    `gorm:"column:status"`
	ContentDigest    string    `gorm:"column:content_digest"`
	CreatedBy        string    `gorm:"column:created_by"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (ProposalModel) TableName() string { return "forma_business_model_proposal" }

type ProvenanceModel struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID         string    `gorm:"column:tenant_id"`
	BusinessID       string    `gorm:"column:business_id"`
	RevisionNo       int32     `gorm:"column:revision_no"`
	ProposalID       string    `gorm:"column:proposal_id"`
	AssertionIDsJSON string    `gorm:"column:assertion_ids_json"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (ProvenanceModel) TableName() string { return "forma_revision_provenance" }

type ModelCallModel struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	RequestID    string    `gorm:"column:request_id"`
	TenantID     string    `gorm:"column:tenant_id"`
	BusinessID   string    `gorm:"column:business_id"`
	SessionID    string    `gorm:"column:session_id"`
	Operation    string    `gorm:"column:operation"`
	ModelRef     string    `gorm:"column:model_ref"`
	LatencyMs    int32     `gorm:"column:latency_ms"`
	Success      bool      `gorm:"column:success"`
	InputTokens  int32     `gorm:"column:input_tokens"`
	OutputTokens int32     `gorm:"column:output_tokens"`
	ErrorMessage string    `gorm:"column:error_message"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (ModelCallModel) TableName() string { return "forma_analyst_model_call" }

type AnalystDAO struct {
	db *gorm.DB
}

func NewAnalystDAO(db *gorm.DB) *AnalystDAO {
	return &AnalystDAO{db: db}
}

func (d *AnalystDAO) WithDB(db *gorm.DB) *AnalystDAO {
	return &AnalystDAO{db: db}
}

func marshalJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func unmarshalJSONStringSlice(s string) []string {
	if s == "" || s == "null" {
		return nil
	}
	var out []string
	_ = json.Unmarshal([]byte(s), &out)
	return out
}

func unmarshalStructured(s string) map[string]any {
	if s == "" || s == "null" {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal([]byte(s), &out)
	return out
}

func toSession(m *SessionModel) *entity.AnalystSession {
	return &entity.AnalystSession{
		SessionID:              m.SessionID,
		TenantID:               m.TenantID,
		BusinessID:             m.BusinessID,
		Status:                 entity.SessionStatus(m.Status),
		Title:                  m.Title,
		RuntimeConversationRef: m.RuntimeConversationRef,
		ConfirmationPolicy:     entity.ConfirmationPolicy(m.ConfirmationPolicy),
		NextTurnSequence:       m.NextTurnSequence,
		CreatedBy:              m.CreatedBy,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		ClosedAt:               m.ClosedAt,
	}
}

func toTurn(m *TurnModel) *entity.AnalystTurn {
	return &entity.AnalystTurn{
		TurnID:          m.TurnID,
		TenantID:        m.TenantID,
		SessionID:       m.SessionID,
		Sequence:        m.Sequence,
		Speaker:         entity.Speaker(m.Speaker),
		Content:         m.Content,
		ContentType:     entity.ContentType(m.ContentType),
		ClientRequestID:       m.ClientRequestID,
		ModelRequestID:        m.ModelRequestID,
		ReplyToTurnID:         m.ReplyToTurnID,
		ReservedReplySequence: m.ReservedReplySequence,
		AnalysisStatus:        entity.AnalysisStatus(m.AnalysisStatus),
		CreatedAt:       m.CreatedAt,
	}
}

func toEvidence(m *EvidenceModel) *entity.BusinessEvidence {
	return &entity.BusinessEvidence{
		EvidenceID:    m.EvidenceID,
		TenantID:      m.TenantID,
		BusinessID:    m.BusinessID,
		SessionID:     m.SessionID,
		TurnID:        m.TurnID,
		SourceType:    entity.EvidenceSourceType(m.SourceType),
		SourceRef:     m.SourceRef,
		Quote:         m.Quote,
		ContentDigest: m.ContentDigest,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
	}
}

func toAssertion(m *AssertionModel) *entity.BusinessAssertion {
	return &entity.BusinessAssertion{
		AssertionID:            m.AssertionID,
		TenantID:               m.TenantID,
		BusinessID:             m.BusinessID,
		SessionID:              m.SessionID,
		AssertionType:          entity.AssertionType(m.AssertionType),
		SubjectRef:             m.SubjectRef,
		Predicate:              m.Predicate,
		ObjectValue:            m.ObjectValue,
		StructuredValue:        unmarshalStructured(m.StructuredValueJSON),
		Confidence:             m.Confidence,
		Status:                 entity.AssertionStatus(m.Status),
		SourceMarker:           businessentity.SourceMarker(m.SourceMarker),
		DerivedFromAssertionID: m.DerivedFromAssertionID,
		CreatedBy:              m.CreatedBy,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

func (d *AnalystDAO) CreateSession(ctx context.Context, s *entity.AnalystSession) error {
	return d.db.WithContext(ctx).Create(&SessionModel{
		SessionID:              s.SessionID,
		TenantID:               s.TenantID,
		BusinessID:             s.BusinessID,
		Status:                 string(s.Status),
		Title:                  s.Title,
		RuntimeConversationRef: s.RuntimeConversationRef,
		ConfirmationPolicy:     string(s.ConfirmationPolicy),
		NextTurnSequence:       s.NextTurnSequence,
		CreatedBy:              s.CreatedBy,
		CreatedAt:              s.CreatedAt,
		UpdatedAt:              s.UpdatedAt,
		ClosedAt:               s.ClosedAt,
	}).Error
}

func (d *AnalystDAO) GetSession(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error) {
	var row SessionModel
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND session_id = ?", tenantID, sessionID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toSession(&row), nil
}

func (d *AnalystDAO) ListSessions(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystSession, error) {
	var rows []SessionModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.AnalystSession, 0, len(rows))
	for i := range rows {
		out = append(out, toSession(&rows[i]))
	}
	return out, nil
}

func (d *AnalystDAO) UpdateSession(ctx context.Context, s *entity.AnalystSession) error {
	return d.db.WithContext(ctx).Model(&SessionModel{}).
		Where("tenant_id = ? AND session_id = ?", s.TenantID, s.SessionID).
		Updates(map[string]any{
			"status":             string(s.Status),
			"title":              s.Title,
			"next_turn_sequence": s.NextTurnSequence,
			"updated_at":         s.UpdatedAt,
			"closed_at":          s.ClosedAt,
		}).Error
}

func (d *AnalystDAO) CreateTurn(ctx context.Context, t *entity.AnalystTurn) error {
	return d.db.WithContext(ctx).Create(&TurnModel{
		TurnID:          t.TurnID,
		TenantID:        t.TenantID,
		SessionID:       t.SessionID,
		Sequence:        t.Sequence,
		Speaker:         string(t.Speaker),
		Content:         t.Content,
		ContentType:     string(t.ContentType),
		ClientRequestID:       t.ClientRequestID,
		ModelRequestID:        t.ModelRequestID,
		ReplyToTurnID:         t.ReplyToTurnID,
		ReservedReplySequence: t.ReservedReplySequence,
		AnalysisStatus:        string(t.AnalysisStatus),
		CreatedAt:       t.CreatedAt,
	}).Error
}

func (d *AnalystDAO) GetTurnByClientRequestID(ctx context.Context, tenantID, sessionID, clientRequestID string) (*entity.AnalystTurn, error) {
	if clientRequestID == "" {
		return nil, nil
	}
	var row TurnModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND session_id = ? AND client_request_id = ?", tenantID, sessionID, clientRequestID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toTurn(&row), nil
}

func (d *AnalystDAO) GetTurn(ctx context.Context, tenantID, turnID string) (*entity.AnalystTurn, error) {
	var row TurnModel
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND turn_id = ?", tenantID, turnID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toTurn(&row), nil
}

func (d *AnalystDAO) GetTurnByReplyTo(ctx context.Context, tenantID, sessionID, replyToTurnID string) (*entity.AnalystTurn, error) {
	if replyToTurnID == "" {
		return nil, nil
	}
	var row TurnModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND session_id = ? AND reply_to_turn_id = ? AND speaker = ?",
			tenantID, sessionID, replyToTurnID, string(entity.SpeakerAnalyst)).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toTurn(&row), nil
}

func (d *AnalystDAO) ListTurns(ctx context.Context, tenantID, sessionID string) ([]*entity.AnalystTurn, error) {
	var rows []TurnModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND session_id = ?", tenantID, sessionID).
		Order("sequence ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.AnalystTurn, 0, len(rows))
	for i := range rows {
		out = append(out, toTurn(&rows[i]))
	}
	return out, nil
}

func (d *AnalystDAO) MaxTurnSequence(ctx context.Context, tenantID, sessionID string) (int32, error) {
	var maxSeq int32
	err := d.db.WithContext(ctx).Model(&TurnModel{}).
		Where("tenant_id = ? AND session_id = ?", tenantID, sessionID).
		Select("COALESCE(MAX(sequence), 0)").
		Scan(&maxSeq).Error
	return maxSeq, err
}

func (d *AnalystDAO) UpdateTurnAnalysis(ctx context.Context, tenantID, turnID string, status entity.AnalysisStatus, modelRequestID string) error {
	return d.db.WithContext(ctx).Model(&TurnModel{}).
		Where("tenant_id = ? AND turn_id = ?", tenantID, turnID).
		Updates(map[string]any{
			"analysis_status": string(status),
			"model_request_id": modelRequestID,
		}).Error
}

func (d *AnalystDAO) CreateEvidence(ctx context.Context, e *entity.BusinessEvidence) error {
	return d.db.WithContext(ctx).Create(&EvidenceModel{
		EvidenceID:    e.EvidenceID,
		TenantID:      e.TenantID,
		BusinessID:    e.BusinessID,
		SessionID:     e.SessionID,
		TurnID:        e.TurnID,
		SourceType:    string(e.SourceType),
		SourceRef:     e.SourceRef,
		Quote:         e.Quote,
		ContentDigest: e.ContentDigest,
		CreatedBy:     e.CreatedBy,
		CreatedAt:     e.CreatedAt,
	}).Error
}

func (d *AnalystDAO) ListEvidence(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessEvidence, error) {
	var rows []EvidenceModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.BusinessEvidence, 0, len(rows))
	for i := range rows {
		out = append(out, toEvidence(&rows[i]))
	}
	return out, nil
}

func (d *AnalystDAO) GetEvidence(ctx context.Context, tenantID, evidenceID string) (*entity.BusinessEvidence, error) {
	var row EvidenceModel
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND evidence_id = ?", tenantID, evidenceID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toEvidence(&row), nil
}

func (d *AnalystDAO) CreateAssertion(ctx context.Context, a *entity.BusinessAssertion) error {
	return d.db.WithContext(ctx).Create(&AssertionModel{
		AssertionID:            a.AssertionID,
		TenantID:               a.TenantID,
		BusinessID:             a.BusinessID,
		SessionID:              a.SessionID,
		AssertionType:          string(a.AssertionType),
		SubjectRef:             a.SubjectRef,
		Predicate:              a.Predicate,
		ObjectValue:            a.ObjectValue,
		StructuredValueJSON:    marshalJSON(a.StructuredValue),
		Confidence:             a.Confidence,
		Status:                 string(a.Status),
		SourceMarker:           string(a.SourceMarker),
		DerivedFromAssertionID: a.DerivedFromAssertionID,
		CreatedBy:              a.CreatedBy,
		CreatedAt:              a.CreatedAt,
		UpdatedAt:              a.UpdatedAt,
	}).Error
}

func (d *AnalystDAO) UpdateAssertion(ctx context.Context, a *entity.BusinessAssertion) error {
	return d.db.WithContext(ctx).Model(&AssertionModel{}).
		Where("tenant_id = ? AND assertion_id = ?", a.TenantID, a.AssertionID).
		Updates(map[string]any{
			"assertion_type":           string(a.AssertionType),
			"subject_ref":              a.SubjectRef,
			"predicate":                a.Predicate,
			"object_value":             a.ObjectValue,
			"structured_value_json":    marshalJSON(a.StructuredValue),
			"confidence":               a.Confidence,
			"status":                   string(a.Status),
			"source_marker":            string(a.SourceMarker),
			"derived_from_assertion_id": a.DerivedFromAssertionID,
			"updated_at":               a.UpdatedAt,
		}).Error
}

func (d *AnalystDAO) GetAssertion(ctx context.Context, tenantID, assertionID string) (*entity.BusinessAssertion, error) {
	var row AssertionModel
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND assertion_id = ?", tenantID, assertionID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toAssertion(&row), nil
}

func (d *AnalystDAO) ListAssertions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessAssertion, error) {
	var rows []AssertionModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.BusinessAssertion, 0, len(rows))
	for i := range rows {
		out = append(out, toAssertion(&rows[i]))
	}
	return out, nil
}

func (d *AnalystDAO) CreateAssertionEvidenceRef(ctx context.Context, tenantID, assertionID, evidenceID string, at time.Time) error {
	return d.db.WithContext(ctx).Create(&AssertionEvidenceRefModel{
		TenantID:    tenantID,
		AssertionID: assertionID,
		EvidenceID:  evidenceID,
		CreatedAt:   at,
	}).Error
}

func (d *AnalystDAO) ListEvidenceIDsForAssertion(ctx context.Context, tenantID, assertionID string) ([]string, error) {
	var rows []AssertionEvidenceRefModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND assertion_id = ?", tenantID, assertionID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.EvidenceID)
	}
	return out, nil
}

func (d *AnalystDAO) ListAssertionIDsForEvidence(ctx context.Context, tenantID, evidenceID string) ([]string, error) {
	var rows []AssertionEvidenceRefModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND evidence_id = ?", tenantID, evidenceID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.AssertionID)
	}
	return out, nil
}

func (d *AnalystDAO) CreateConfirmation(ctx context.Context, c *entity.BusinessConfirmation) error {
	return d.db.WithContext(ctx).Create(&ConfirmationModel{
		ConfirmationID: c.ConfirmationID,
		TenantID:       c.TenantID,
		BusinessID:     c.BusinessID,
		AssertionID:    c.AssertionID,
		Decision:       string(c.Decision),
		Comment:        c.Comment,
		DecidedBy:      c.DecidedBy,
		DecidedAt:      c.DecidedAt,
	}).Error
}

func (d *AnalystDAO) ListConfirmationsForAssertion(ctx context.Context, tenantID, assertionID string) ([]*entity.BusinessConfirmation, error) {
	var rows []ConfirmationModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND assertion_id = ?", tenantID, assertionID).
		Order("decided_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.BusinessConfirmation, 0, len(rows))
	for i := range rows {
		out = append(out, &entity.BusinessConfirmation{
			ConfirmationID: rows[i].ConfirmationID,
			TenantID:       rows[i].TenantID,
			BusinessID:     rows[i].BusinessID,
			AssertionID:    rows[i].AssertionID,
			Decision:       entity.ConfirmationDecision(rows[i].Decision),
			Comment:        rows[i].Comment,
			DecidedBy:      rows[i].DecidedBy,
			DecidedAt:      rows[i].DecidedAt,
		})
	}
	return out, nil
}

func (d *AnalystDAO) GetSessionForUpdate(ctx context.Context, tenantID, sessionID string) (*entity.AnalystSession, error) {
	var row SessionModel
	err := d.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND session_id = ?", tenantID, sessionID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toSession(&row), nil
}

func (d *AnalystDAO) GetConflictByPair(ctx context.Context, tenantID, businessID, assertionIDA, assertionIDB string) (*entity.AssertionConflict, error) {
	var row ConflictModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ? AND assertion_id_a = ? AND assertion_id_b = ?",
			tenantID, businessID, assertionIDA, assertionIDB).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.AssertionConflict{
		ConflictID:   row.ConflictID,
		TenantID:     row.TenantID,
		BusinessID:   row.BusinessID,
		SessionID:    row.SessionID,
		AssertionIDA: row.AssertionIDA,
		AssertionIDB: row.AssertionIDB,
		SubjectRef:   row.SubjectRef,
		Predicate:    row.Predicate,
		Status:       entity.ConflictStatus(row.Status),
		CreatedAt:    row.CreatedAt,
	}, nil
}

func (d *AnalystDAO) UpdateConflictStatus(ctx context.Context, tenantID, conflictID string, status entity.ConflictStatus, at time.Time) error {
	return d.db.WithContext(ctx).Model(&ConflictModel{}).
		Where("tenant_id = ? AND conflict_id = ?", tenantID, conflictID).
		Update("status", string(status)).Error
}

func (d *AnalystDAO) CreateConflict(ctx context.Context, c *entity.AssertionConflict) error {
	return d.db.WithContext(ctx).Create(&ConflictModel{
		ConflictID:   c.ConflictID,
		TenantID:     c.TenantID,
		BusinessID:   c.BusinessID,
		SessionID:    c.SessionID,
		AssertionIDA: c.AssertionIDA,
		AssertionIDB: c.AssertionIDB,
		SubjectRef:   c.SubjectRef,
		Predicate:    c.Predicate,
		Status:       string(c.Status),
		CreatedAt:    c.CreatedAt,
	}).Error
}

func (d *AnalystDAO) ListConflicts(ctx context.Context, tenantID, businessID string) ([]*entity.AssertionConflict, error) {
	var rows []ConflictModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.AssertionConflict, 0, len(rows))
	for i := range rows {
		out = append(out, &entity.AssertionConflict{
			ConflictID:   rows[i].ConflictID,
			TenantID:     rows[i].TenantID,
			BusinessID:   rows[i].BusinessID,
			SessionID:    rows[i].SessionID,
			AssertionIDA: rows[i].AssertionIDA,
			AssertionIDB: rows[i].AssertionIDB,
			SubjectRef:   rows[i].SubjectRef,
			Predicate:    rows[i].Predicate,
			Status:       entity.ConflictStatus(rows[i].Status),
			CreatedAt:    rows[i].CreatedAt,
		})
	}
	return out, nil
}

func (d *AnalystDAO) CreateGap(ctx context.Context, g *entity.AnalystGap) error {
	return d.db.WithContext(ctx).Create(&GapModel{
		GapID:                   g.GapID,
		TenantID:                g.TenantID,
		BusinessID:              g.BusinessID,
		SessionID:               g.SessionID,
		GapType:                 g.GapType,
		Question:                g.Question,
		RelatedAssertionIDsJSON: marshalJSON(g.RelatedAssertionIDs),
		Status:                  string(g.Status),
		CreatedAt:               g.CreatedAt,
		UpdatedAt:               g.UpdatedAt,
	}).Error
}

func (d *AnalystDAO) ListGaps(ctx context.Context, tenantID, businessID string) ([]*entity.AnalystGap, error) {
	var rows []GapModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ?", tenantID, businessID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.AnalystGap, 0, len(rows))
	for i := range rows {
		out = append(out, &entity.AnalystGap{
			GapID:               rows[i].GapID,
			TenantID:            rows[i].TenantID,
			BusinessID:          rows[i].BusinessID,
			SessionID:           rows[i].SessionID,
			GapType:             rows[i].GapType,
			Question:            rows[i].Question,
			RelatedAssertionIDs: unmarshalJSONStringSlice(rows[i].RelatedAssertionIDsJSON),
			Status:              entity.GapStatus(rows[i].Status),
			CreatedAt:           rows[i].CreatedAt,
			UpdatedAt:           rows[i].UpdatedAt,
		})
	}
	return out, nil
}

func (d *AnalystDAO) UpdateGapStatus(ctx context.Context, tenantID, gapID string, status entity.GapStatus, at time.Time) error {
	return d.db.WithContext(ctx).Model(&GapModel{}).
		Where("tenant_id = ? AND gap_id = ?", tenantID, gapID).
		Updates(map[string]any{
			"status":     string(status),
			"updated_at": at,
		}).Error
}

func (d *AnalystDAO) CreateProposal(ctx context.Context, p *entity.BusinessModelProposal) error {
	patchJSON := ""
	if p.Patch != nil {
		patchJSON = marshalJSON(p.Patch)
	}
	return d.db.WithContext(ctx).Create(&ProposalModel{
		ProposalID:       p.ProposalID,
		TenantID:         p.TenantID,
		BusinessID:       p.BusinessID,
		SessionID:        p.SessionID,
		BaseRevision:     p.BaseRevision,
		AssertionIDsJSON: marshalJSON(p.AssertionIDs),
		PatchJSON:        patchJSON,
		Status:           string(p.Status),
		ContentDigest:    p.ContentDigest,
		CreatedBy:        p.CreatedBy,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}).Error
}

func (d *AnalystDAO) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error) {
	var row ProposalModel
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND proposal_id = ?", tenantID, proposalID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toProposal(&row), nil
}

func (d *AnalystDAO) GetProposalForUpdate(ctx context.Context, tenantID, proposalID string) (*entity.BusinessModelProposal, error) {
	var row ProposalModel
	err := d.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND proposal_id = ?", tenantID, proposalID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toProposal(&row), nil
}

func (d *AnalystDAO) UpdateProposalStatus(ctx context.Context, tenantID, proposalID string, status entity.ProposalStatus, at time.Time) error {
	return d.db.WithContext(ctx).Model(&ProposalModel{}).
		Where("tenant_id = ? AND proposal_id = ?", tenantID, proposalID).
		Updates(map[string]any{
			"status":     string(status),
			"updated_at": at,
		}).Error
}

func toProposal(m *ProposalModel) *entity.BusinessModelProposal {
	var patch entity.SemanticModelPatch
	if m.PatchJSON != "" {
		_ = json.Unmarshal([]byte(m.PatchJSON), &patch)
	}
	return &entity.BusinessModelProposal{
		ProposalID:    m.ProposalID,
		TenantID:      m.TenantID,
		BusinessID:    m.BusinessID,
		SessionID:     m.SessionID,
		BaseRevision:  m.BaseRevision,
		AssertionIDs:  unmarshalJSONStringSlice(m.AssertionIDsJSON),
		Patch:         &patch,
		Status:        entity.ProposalStatus(m.Status),
		ContentDigest: m.ContentDigest,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func (d *AnalystDAO) CreateProvenance(ctx context.Context, p *entity.RevisionProvenance) error {
	return d.db.WithContext(ctx).Create(&ProvenanceModel{
		TenantID:         p.TenantID,
		BusinessID:       p.BusinessID,
		RevisionNo:       p.RevisionNo,
		ProposalID:       p.ProposalID,
		AssertionIDsJSON: marshalJSON(p.AssertionIDs),
		CreatedAt:        p.CreatedAt,
	}).Error
}

func (d *AnalystDAO) GetProvenance(ctx context.Context, tenantID, businessID string, revisionNo int32) (*entity.RevisionProvenance, error) {
	var row ProvenanceModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND business_id = ? AND revision_no = ?", tenantID, businessID, revisionNo).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.RevisionProvenance{
		TenantID:     row.TenantID,
		BusinessID:   row.BusinessID,
		RevisionNo:   row.RevisionNo,
		ProposalID:   row.ProposalID,
		AssertionIDs: unmarshalJSONStringSlice(row.AssertionIDsJSON),
		CreatedAt:    row.CreatedAt,
	}, nil
}

func (d *AnalystDAO) CreateModelCall(ctx context.Context, r *entity.ModelCallRecord) error {
	return d.db.WithContext(ctx).Create(&ModelCallModel{
		RequestID:    r.RequestID,
		TenantID:     r.TenantID,
		BusinessID:   r.BusinessID,
		SessionID:    r.SessionID,
		Operation:    r.Operation,
		ModelRef:     r.ModelRef,
		LatencyMs:    r.LatencyMs,
		Success:      r.Success,
		InputTokens:  r.InputTokens,
		OutputTokens: r.OutputTokens,
		ErrorMessage: r.ErrorMessage,
		CreatedAt:    r.CreatedAt,
	}).Error
}
