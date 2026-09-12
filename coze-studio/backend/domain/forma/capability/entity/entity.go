/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "time"

// RevisionStatus — BusinessCapabilityRevision lifecycle (§8.1).
type RevisionStatus string

const (
	RevisionDraft      RevisionStatus = "DRAFT"
	RevisionValidated  RevisionStatus = "VALIDATED"
	RevisionActive     RevisionStatus = "ACTIVE"
	RevisionStale      RevisionStatus = "STALE"
	RevisionDeprecated RevisionStatus = "DEPRECATED"
)

// CapabilityKind — domain-agnostic kinds only (§7.1).
type CapabilityKind string

const (
	KindQuery   CapabilityKind = "QUERY"
	KindCommand CapabilityKind = "COMMAND"
)

// Source — revision provenance (§6.4).
type Source string

const (
	SourceManualCreated Source = "MANUAL_CREATED"
	SourceAIProposal    Source = "AI_PROPOSAL"
	SourceDerivedEdit   Source = "DERIVED_EDIT"
)

// ProposalStatus — CapabilityProposal workflow (§9).
type ProposalStatus string

const (
	ProposalProposed      ProposalStatus = "PROPOSED"
	ProposalRejected      ProposalStatus = "REJECTED"
	ProposalConfirmed     ProposalStatus = "CONFIRMED"
	ProposalEditConfirmed ProposalStatus = "EDIT_CONFIRMED"
)

// DecisionAction — immutable human decision (§6.4 / §8.5 / §9).
type DecisionAction string

const (
	DecisionCreate      DecisionAction = "CREATE"
	DecisionConfirm     DecisionAction = "CONFIRM"
	DecisionEditConfirm DecisionAction = "EDIT_CONFIRM"
	DecisionReject      DecisionAction = "REJECT"
	DecisionEdit        DecisionAction = "EDIT"
	DecisionDerive      DecisionAction = "DERIVE"
	DecisionActivate    DecisionAction = "ACTIVATE"
	DecisionDeprecate   DecisionAction = "DEPRECATE"
)

// AnalysisStatus — CapabilityAnalysisRun (§9.5).
type AnalysisStatus string

const (
	AnalysisPending   AnalysisStatus = "PENDING"
	AnalysisSucceeded AnalysisStatus = "SUCCEEDED"
	AnalysisFailed    AnalysisStatus = "FAILED"
)

// QueryOperation — V1 QUERY ops (§10.1.2); LOOKUP/AGGREGATE deferred.
type QueryOperation string

const (
	QueryOpRead   QueryOperation = "READ"
	QueryOpList   QueryOperation = "LIST"
	QueryOpFilter QueryOperation = "FILTER"
)

// OutputCardinality — frozen V1 enum (§10.1.4).
type OutputCardinality string

const (
	CardinalityOne  OutputCardinality = "ONE"
	CardinalityMany OutputCardinality = "MANY"
)

// LogicalField is a capability input/output logical field (no physical paths).
type LogicalField struct {
	LogicalKey  string `json:"logical_key"`
	LogicalType string `json:"logical_type"`
	Required    bool   `json:"required,omitempty"`
	Nullable    bool   `json:"nullable,omitempty"`
	Description string `json:"description,omitempty"`
}

// LogicalSchema is a JSON-friendly logical schema document.
type LogicalSchema struct {
	Fields []LogicalField `json:"fields"`
}

// PredicateKind — frozen V1 precondition predicates (§10.3). Sole enum; no free-form Operator.
type PredicateKind string

const (
	PredicateExists   PredicateKind = "EXISTS"
	PredicateEQ       PredicateKind = "EQ"
	PredicateNEQ      PredicateKind = "NEQ"
	PredicateIN       PredicateKind = "IN"
	PredicateNotIn    PredicateKind = "NOT_IN"
	PredicateGT       PredicateKind = "GT"
	PredicateGTE      PredicateKind = "GTE"
	PredicateLT       PredicateKind = "LT"
	PredicateLTE      PredicateKind = "LTE"
	PredicateEmpty    PredicateKind = "EMPTY"
	PredicateNotEmpty PredicateKind = "NOT_EMPTY"
)

// EffectKind — frozen V1 effect kinds (§10.3).
type EffectKind string

const (
	EffectReadOnly    EffectKind = "READ_ONLY"
	EffectIntent      EffectKind = "INTENT"
	EffectStateChange EffectKind = "STATE_CHANGE"
	EffectNotify      EffectKind = "NOTIFY"
)

// OpaqueID is a constrained identifier (no spaces / SQL / path injection).
type OpaqueID string

// Precondition is a structured typed record (not free-form code) (§10.3).
// Predicate is the sole comparison enum — Operator field removed (S5-G2-F2).
type Precondition struct {
	ID          string        `json:"id"`
	Predicate   PredicateKind `json:"predicate"`
	LogicalKey  string        `json:"logical_key,omitempty"`
	Comparand   any           `json:"comparand,omitempty"`
	Description string        `json:"description,omitempty"`
}

// Effect declares business postcondition intent (not executable).
type Effect struct {
	ID          string     `json:"id"`
	Kind        EffectKind `json:"kind"`
	LogicalKey  string     `json:"logical_key,omitempty"`
	Description string     `json:"description,omitempty"`
}

// LogicalFieldMapping maps capability logical keys to contract logical keys only.
type LogicalFieldMapping struct {
	CapabilityLogicalKey string `json:"capability_logical_key"`
	ContractLogicalKey   string `json:"contract_logical_key"`
}

// DataContractBinding pins a Data Contract by ID/version (logical only).
type DataContractBinding struct {
	DataContractID         string                `json:"data_contract_id"`
	DataContractRevisionID string                `json:"data_contract_revision_id"`
	DataContractVersion    int32                 `json:"data_contract_version,omitempty"`
	LogicalFieldMappings   []LogicalFieldMapping `json:"logical_field_mappings,omitempty"`
}

// SemanticPayload holds behavioral fields used for ContentDigest / DecisionPayloadDigest.
// Excludes provenance, identity, status, version, and audit fields (§3.4.2).
type SemanticPayload struct {
	Name                  string                `json:"name"`
	Description           string                `json:"description"`
	CapabilityKind        CapabilityKind        `json:"capability_kind"`
	BusinessModelRevision int32                 `json:"business_model_revision"`
	InputSchema           LogicalSchema         `json:"input_schema"`
	OutputSchema          LogicalSchema         `json:"output_schema"`
	Preconditions         []Precondition        `json:"preconditions"`
	Effects               []Effect              `json:"effects"`
	DataContractBindings  []DataContractBinding `json:"data_contract_bindings"`
	QueryOperation        QueryOperation        `json:"query_operation,omitempty"`
	OutputCardinality     OutputCardinality     `json:"output_cardinality,omitempty"`
}

// BusinessCapability is the CAPABILITY aggregate (non-semantic SoT) (§6.4).
// capability_id == asset_id.
type BusinessCapability struct {
	CapabilityID         string
	TenantID             string
	BusinessID           string
	ActiveRevisionID     string // empty when no ACTIVE
	AggregateGeneration  int64  // bumped on successful Activate (CAS fence)
	CreatedBy            string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// BusinessCapabilityRevision is the versioned semantic SoT (§6.4). Immutable after create.
type BusinessCapabilityRevision struct {
	RevisionID            string
	CapabilityID          string
	TenantID              string
	BusinessID            string
	Version               int32
	Status                RevisionStatus
	Name                  string
	Description           string
	BusinessModelRevision int32
	CapabilityKind        CapabilityKind
	InputSchema           LogicalSchema
	OutputSchema          LogicalSchema
	Preconditions         []Precondition
	Effects               []Effect
	DataContractBindings  []DataContractBinding
	QueryOperation        QueryOperation
	OutputCardinality     OutputCardinality
	DerivedFromRevisionID string
	AnalysisRunID         string
	ProposalID            string
	Source                Source
	CreatedBy             string
	CreatedAt             time.Time
}

// ToSemanticPayload extracts behavioral fields for digests.
func (r *BusinessCapabilityRevision) ToSemanticPayload() SemanticPayload {
	if r == nil {
		return SemanticPayload{}
	}
	return SemanticPayload{
		Name:                  r.Name,
		Description:           r.Description,
		CapabilityKind:        r.CapabilityKind,
		BusinessModelRevision: r.BusinessModelRevision,
		InputSchema:           cloneLogicalSchema(r.InputSchema),
		OutputSchema:          cloneLogicalSchema(r.OutputSchema),
		Preconditions:         clonePreconditions(r.Preconditions),
		Effects:               cloneEffects(r.Effects),
		DataContractBindings:  cloneBindings(r.DataContractBindings),
		QueryOperation:        r.QueryOperation,
		OutputCardinality:     r.OutputCardinality,
	}
}

// CapabilityProposal is an AI workflow record (not a Canonical Asset).
type CapabilityProposal struct {
	ProposalID             string
	TenantID               string
	BusinessID             string
	AnalysisRunID          string
	CapabilityID           string // nullable until CONFIRM/EDIT_CONFIRM
	Status                 ProposalStatus
	Payload                SemanticPayload // immutable after create
	MaterializedRevisionID string
	CreatedAt              time.Time
}

// CapabilityDecision is an immutable human decision event.
type CapabilityDecision struct {
	DecisionID       string
	TenantID         string
	BusinessID       string
	CapabilityID     string // null only for unbound Proposal REJECT
	ProposalID       string
	SourceRevisionID string
	TargetRevisionID string
	Action           DecisionAction
	PayloadDigest    string
	ClientRequestID  string
	ActorPrincipalID string
	Reason           string
	CreatedAt        time.Time
}

// CapabilityAnalysisRun binds one AI analysis execution (§9.5).
type CapabilityAnalysisRun struct {
	AnalysisRunID         string
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	ClientRequestID       string
	RequestDigest         string
	Status                AnalysisStatus
	Attempt               int32 // monotonic generation fence
	ErrorCode             string
	ModelRef              string
	RequestJSON           string // sanitized request snapshot for retry
	ExecutionClaimedAt    *time.Time
	LeaseExpiresAt        *time.Time
	CreatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// AnalysisRequest covers fields hashed into AnalysisRequestDigest.
// No free-form Options map — secrets must not ride along on analysis requests.
type AnalysisRequest struct {
	BusinessModelRevision int32             `json:"business_model_revision"`
	DataContractPins      []DataContractPin `json:"data_contract_pins"`
	RequirementRefs       []string          `json:"requirement_refs"`
}

// DataContractPin is a sorted pin used in analysis request digests.
type DataContractPin struct {
	DataContractID      string `json:"data_contract_id"`
	DataContractVersion int32  `json:"data_contract_version"`
}

// AnalysisAttemptTrigger — why an analysis attempt row was created.
type AnalysisAttemptTrigger string

const (
	AttemptTriggerFirst         AnalysisAttemptTrigger = "FIRST"
	AttemptTriggerRetry         AnalysisAttemptTrigger = "RETRY"
	AttemptTriggerLeaseTakeover AnalysisAttemptTrigger = "LEASE_TAKEOVER"
)

// AnalysisAttemptResult — attempt lifecycle result.
type AnalysisAttemptResult string

const (
	AttemptResultPending    AnalysisAttemptResult = "PENDING"
	AttemptResultSucceeded  AnalysisAttemptResult = "SUCCEEDED"
	AttemptResultFailed     AnalysisAttemptResult = "FAILED"
	AttemptResultSuperseded AnalysisAttemptResult = "SUPERSEDED"
)

// CapabilityAnalysisAttempt is an audit row for each claim / retry / lease takeover.
type CapabilityAnalysisAttempt struct {
	AttemptID        string
	AnalysisRunID    string
	TenantID         string
	Attempt          int32
	ActorPrincipalID string
	TriggerKind      AnalysisAttemptTrigger
	ResultStatus     AnalysisAttemptResult
	ErrorCode        string
	CreatedAt        time.Time
	CompletedAt      *time.Time
}

func cloneLogicalSchema(in LogicalSchema) LogicalSchema {
	out := LogicalSchema{Fields: make([]LogicalField, len(in.Fields))}
	copy(out.Fields, in.Fields)
	return out
}

func clonePreconditions(in []Precondition) []Precondition {
	if in == nil {
		return nil
	}
	out := make([]Precondition, len(in))
	copy(out, in)
	return out
}

func cloneEffects(in []Effect) []Effect {
	if in == nil {
		return nil
	}
	out := make([]Effect, len(in))
	copy(out, in)
	return out
}

func cloneBindings(in []DataContractBinding) []DataContractBinding {
	if in == nil {
		return nil
	}
	out := make([]DataContractBinding, len(in))
	for i := range in {
		out[i] = in[i]
		if in[i].LogicalFieldMappings != nil {
			out[i].LogicalFieldMappings = append([]LogicalFieldMapping(nil), in[i].LogicalFieldMappings...)
		}
	}
	return out
}
