/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "time"

type RequirementKind string

const (
	KindEntity     RequirementKind = "ENTITY"
	KindAttribute  RequirementKind = "ATTRIBUTE"
	KindRelation   RequirementKind = "RELATION"
	KindEvent      RequirementKind = "EVENT"
	KindMetric     RequirementKind = "METRIC"
	KindState      RequirementKind = "STATE"
	KindTimeSeries RequirementKind = "TIME_SERIES"
	KindDocument   RequirementKind = "DOCUMENT"
	KindLookup     RequirementKind = "LOOKUP"
	KindHistory    RequirementKind = "HISTORY"
)

type RequirementStatus string

const (
	StatusProposed   RequirementStatus = "PROPOSED"
	StatusConfirmed  RequirementStatus = "CONFIRMED"
	StatusRejected   RequirementStatus = "REJECTED"
	StatusSuperseded RequirementStatus = "SUPERSEDED"
)

type RequirementSource string

const (
	SourceAIGenerated    RequirementSource = "AI_GENERATED"
	SourceManualCreated  RequirementSource = "MANUAL_CREATED"
	SourceManualModified RequirementSource = "MANUAL_MODIFIED"
)

type AnalysisRunStatus string

const (
	AnalysisPending   AnalysisRunStatus = "PENDING"
	AnalysisSucceeded AnalysisRunStatus = "SUCCEEDED"
	AnalysisFailed    AnalysisRunStatus = "FAILED"
)

type DecisionAction string

const (
	DecisionConfirm     DecisionAction = "CONFIRM"
	DecisionReject      DecisionAction = "REJECT"
	DecisionEditConfirm DecisionAction = "EDIT_CONFIRM"
)

// DataRequirement is a domain-agnostic data need derived from Business Model.
type DataRequirement struct {
	RequirementID             string
	TenantID                  string
	BusinessID                string
	BusinessModelRevision     int32
	RequirementKind           RequirementKind
	SemanticName              string
	Description               string
	BusinessElementRefs       []string
	Requiredness              string
	FreshnessRequirement      string
	AccessNeed                string
	Status                    RequirementStatus
	Source                    RequirementSource
	DerivedFromRequirementID  string
	AnalysisRunID             string
	CreatedBy                 string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// DataRequirementAnalysisRun binds one AnalyzeDataRequirements execution.
type DataRequirementAnalysisRun struct {
	AnalysisRunID           string
	TenantID                string
	BusinessID              string
	BusinessModelRevision   int32
	ClientRequestID         string
	RequestDigest           string
	Status                  AnalysisRunStatus
	ModelRef                string
	ErrorKey                string
	ErrorMessageSanitized   string
	RetryCount              int32
	LastRetryBy             string
	LastRetryAt             *time.Time
	ExecutionGeneration     int32
	ExecutionClaimedAt      *time.Time
	LeaseExpiresAt          *time.Time
	CreatedBy               string
	StartedAt               *time.Time
	CompletedAt             *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// DataRequirementDecision is an immutable human audit event.
type DataRequirementDecision struct {
	DecisionID            string
	TenantID              string
	BusinessID            string
	SourceRequirementID   string
	TargetRequirementID   string
	Action                DecisionAction
	ActorPrincipalID      string
	Reason                string
	BusinessModelRevision int32
	CreatedAt             time.Time
}

// DataRequirementProposal is AI output prior to domain persistence.
type DataRequirementProposal struct {
	RequirementKind      RequirementKind
	SemanticName         string
	Description          string
	BusinessElementRefs  []string
	Requiredness         string
	FreshnessRequirement string
	AccessNeed           string
	// Forbidden if set by model — domain rejects.
	Status RequirementStatus
	Source RequirementSource
}
