/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import (
	"time"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

type SessionStatus string

const (
	SessionDraft      SessionStatus = "DRAFT"
	SessionActive     SessionStatus = "ACTIVE"
	SessionReviewing  SessionStatus = "REVIEWING"
	SessionCompleted  SessionStatus = "COMPLETED"
	SessionCancelled  SessionStatus = "CANCELLED"
)

type Speaker string

const (
	SpeakerUser    Speaker = "USER"
	SpeakerAnalyst Speaker = "ANALYST"
)

type ContentType string

const (
	ContentText ContentType = "TEXT"
)

type EvidenceSourceType string

const (
	EvidenceInterviewTurn EvidenceSourceType = "INTERVIEW_TURN"
	EvidenceManualNote    EvidenceSourceType = "MANUAL_NOTE"
	EvidenceDocumentRef   EvidenceSourceType = "DOCUMENT_REF"
)

type AssertionType string

const (
	AssertionActorExists          AssertionType = "ACTOR_EXISTS"
	AssertionBusinessObjectExists AssertionType = "BUSINESS_OBJECT_EXISTS"
	AssertionProcessExists        AssertionType = "PROCESS_EXISTS"
	AssertionEventExists          AssertionType = "EVENT_EXISTS"
	AssertionSystemExists         AssertionType = "SYSTEM_EXISTS"
	AssertionPolicyExists         AssertionType = "POLICY_EXISTS"
	AssertionRelationExists         AssertionType = "RELATION_EXISTS"
	AssertionStateExists            AssertionType = "STATE_EXISTS"
	AssertionStateTransition        AssertionType = "STATE_TRANSITION"
	AssertionBusinessRule           AssertionType = "BUSINESS_RULE"
	AssertionProperty               AssertionType = "PROPERTY"
)

type AssertionStatus string

const (
	AssertionProposed   AssertionStatus = "PROPOSED"
	AssertionConfirmed  AssertionStatus = "CONFIRMED"
	AssertionRejected   AssertionStatus = "REJECTED"
	AssertionSuperseded AssertionStatus = "SUPERSEDED"
)

type ConfirmationDecision string

const (
	DecisionConfirm ConfirmationDecision = "CONFIRM"
	DecisionReject  ConfirmationDecision = "REJECT"
)

type ConfirmationPolicy string

const (
	PolicyDevelopment ConfirmationPolicy = "DEVELOPMENT"
	PolicyProduction  ConfirmationPolicy = "PRODUCTION"
)

type GapStatus string

const (
	GapOpen      GapStatus = "OPEN"
	GapResolved  GapStatus = "RESOLVED"
	GapDismissed GapStatus = "DISMISSED"
)

type ConflictStatus string

const (
	ConflictOpen     ConflictStatus = "OPEN"
	ConflictResolved ConflictStatus = "RESOLVED"
)

type ProposalStatus string

const (
	ProposalDraft          ProposalStatus = "DRAFT"
	ProposalReadyForReview ProposalStatus = "READY_FOR_REVIEW"
	ProposalApplied        ProposalStatus = "APPLIED"
	ProposalRejected       ProposalStatus = "REJECTED"
	ProposalStale          ProposalStatus = "STALE"
)

type AnalysisStatus string

const (
	AnalysisNone              AnalysisStatus = "NONE"
	AnalysisPending           AnalysisStatus = "PENDING"
	AnalysisCompleted         AnalysisStatus = "COMPLETED"
	AnalysisFailed            AnalysisStatus = "FAILED" // legacy; prefer EXTRACTION_FAILED
	AnalysisExtractionFailed  AnalysisStatus = "EXTRACTION_FAILED"
	AnalysisResponseFailed    AnalysisStatus = "RESPONSE_FAILED"
)

// AnalystSession is the canonical Forma interview session (not a Coze conversation alias).
type AnalystSession struct {
	SessionID              string
	TenantID               string
	BusinessID             string
	Status                 SessionStatus
	Title                  string
	RuntimeConversationRef string
	ConfirmationPolicy     ConfirmationPolicy
	NextTurnSequence       int32
	CreatedBy              string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ClosedAt               *time.Time
}

type AnalystTurn struct {
	TurnID                string
	TenantID              string
	SessionID             string
	Sequence              int32
	Speaker               Speaker
	Content               string
	ContentType           ContentType
	ClientRequestID       string
	ModelRequestID        string
	ReplyToTurnID         string
	ReservedReplySequence int32
	AnalysisStatus        AnalysisStatus
	CreatedAt             time.Time
}

type BusinessEvidence struct {
	EvidenceID    string
	TenantID      string
	BusinessID    string
	SessionID     string
	TurnID        string
	SourceType    EvidenceSourceType
	SourceRef     string
	Quote         string
	ContentDigest string
	CreatedBy     string
	CreatedAt     time.Time
}

type BusinessAssertion struct {
	AssertionID            string
	TenantID               string
	BusinessID             string
	SessionID              string
	AssertionType          AssertionType
	SubjectRef             string
	Predicate              string
	ObjectValue            string
	StructuredValue        map[string]any
	Confidence             float64
	Status                 AssertionStatus
	SourceMarker           businessentity.SourceMarker
	DerivedFromAssertionID string
	CreatedBy              string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	EvidenceIDs            []string
}

type BusinessConfirmation struct {
	ConfirmationID string
	TenantID       string
	BusinessID     string
	AssertionID    string
	Decision       ConfirmationDecision
	Comment        string
	DecidedBy      string
	DecidedAt      time.Time
}

type AssertionConflict struct {
	ConflictID   string
	TenantID     string
	BusinessID   string
	SessionID    string
	AssertionIDA string
	AssertionIDB string
	SubjectRef   string
	Predicate    string
	Status       ConflictStatus
	CreatedAt    time.Time
}

type AnalystGap struct {
	GapID                string
	TenantID             string
	BusinessID           string
	SessionID            string
	GapType              string
	Question             string
	RelatedAssertionIDs  []string
	Status               GapStatus
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// PatchOpType semantic patch operations (not raw JSON Patch).
type PatchOpType string

const (
	PatchAddNode     PatchOpType = "ADD_NODE"
	PatchUpdateNode  PatchOpType = "UPDATE_NODE"
	PatchRemoveNode  PatchOpType = "REMOVE_NODE"
	PatchAddEdge     PatchOpType = "ADD_EDGE"
	PatchUpdateEdge  PatchOpType = "UPDATE_EDGE"
	PatchRemoveEdge  PatchOpType = "REMOVE_EDGE"
	PatchAddState    PatchOpType = "ADD_STATE"
	PatchUpdateState PatchOpType = "UPDATE_STATE"
	PatchRemoveState PatchOpType = "REMOVE_STATE"
	PatchAddRule     PatchOpType = "ADD_RULE"
	PatchUpdateRule  PatchOpType = "UPDATE_RULE"
	PatchRemoveRule  PatchOpType = "REMOVE_RULE"
)

type PatchOperation struct {
	Op                 PatchOpType                    `json:"op"`
	TargetID           string                         `json:"target_id,omitempty"`
	Node               *businessentity.SemanticNode   `json:"node,omitempty"`
	Edge               *businessentity.SemanticEdge   `json:"edge,omitempty"`
	State              *businessentity.BusinessState  `json:"state,omitempty"`
	Rule               *businessentity.BusinessRule   `json:"rule,omitempty"`
	SourceAssertionIDs []string                       `json:"source_assertion_ids"`
}

type SemanticModelPatch struct {
	Operations []PatchOperation `json:"operations"`
}

type BusinessModelProposal struct {
	ProposalID     string
	TenantID       string
	BusinessID     string
	SessionID      string
	BaseRevision   int32
	AssertionIDs   []string
	Patch          *SemanticModelPatch
	Status         ProposalStatus
	ContentDigest  string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RevisionProvenance struct {
	TenantID     string
	BusinessID   string
	RevisionNo   int32
	ProposalID   string
	AssertionIDs []string
	CreatedAt    time.Time
}

type ModelCallRecord struct {
	RequestID    string
	TenantID     string
	BusinessID   string
	SessionID    string
	Operation    string
	ModelRef     string
	LatencyMs    int32
	Success      bool
	InputTokens  int32
	OutputTokens int32
	ErrorMessage string
	CreatedAt    time.Time
}

// NextQuestionPlan output from interview planner.
type NextQuestionPlan struct {
	Question         string   `json:"question"`
	Goal             string   `json:"goal"`
	RelatedElements  []string `json:"related_elements"`
	Priority         int      `json:"priority"`
}

// ContextManifest for debugging context budgeting.
type ContextManifest struct {
	IncludedItems []string `json:"included_items"`
	ExcludedItems []string `json:"excluded_items"`
	TokenEstimate int      `json:"token_estimate"`
}

// Structured extraction contract from model.
type ExtractionAssertion struct {
	AssertionType   AssertionType   `json:"assertion_type"`
	SubjectRef      string          `json:"subject_ref"`
	Predicate       string          `json:"predicate"`
	ObjectValue     string          `json:"object_value"`
	StructuredValue map[string]any  `json:"structured_value,omitempty"`
	Confidence      float64         `json:"confidence"`
	EvidenceTurnIDs []string        `json:"evidence_turn_ids"`
}

type ExtractionEvidenceLink struct {
	AssertionIndex int    `json:"assertion_index"`
	EvidenceTurnID string `json:"evidence_turn_id"`
}

type ExtractionGap struct {
	GapType    string `json:"gap_type"`
	Question   string `json:"question"`
}

type ExtractionConflict struct {
	AssertionIndexA int `json:"assertion_index_a"`
	AssertionIndexB int `json:"assertion_index_b"`
}

type ExtractionResult struct {
	Assertions    []ExtractionAssertion    `json:"assertions"`
	EvidenceLinks []ExtractionEvidenceLink `json:"evidence_links"`
	Gaps          []ExtractionGap        `json:"gaps"`
	Conflicts     []ExtractionConflict   `json:"conflicts"`
}

type TurnSubmissionResult struct {
	UserTurn       *AnalystTurn
	AnalystTurn    *AnalystTurn
	Evidence       *BusinessEvidence
	Assertions     []*BusinessAssertion
	Conflicts      []*AssertionConflict
	Gaps           []*AnalystGap
	NextQuestion   *NextQuestionPlan
	ModelFailed    bool
	ModelError     string
	ContextManifest *ContextManifest
}
