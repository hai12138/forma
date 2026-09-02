/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import (
	"encoding/json"
	"time"
)

type MappingStatus string

const (
	MappingStatusProposed   MappingStatus = "PROPOSED"
	MappingStatusConfirmed  MappingStatus = "CONFIRMED"
	MappingStatusRejected   MappingStatus = "REJECTED"
	MappingStatusSuperseded MappingStatus = "SUPERSEDED"
)

type MappingSource string

const (
	MappingSourceAIGenerated    MappingSource = "AI_GENERATED"
	MappingSourceManualCreated  MappingSource = "MANUAL_CREATED"
	MappingSourceManualModified MappingSource = "MANUAL_MODIFIED"
)

type MappingType string

const (
	MappingTypeDirect        MappingType = "DIRECT"
	MappingTypeCast          MappingType = "CAST"
	MappingTypeEnumMap       MappingType = "ENUM_MAP"
	MappingTypeUnitConvert   MappingType = "UNIT_CONVERT"
	MappingTypeTimeNormalize MappingType = "TIME_NORMALIZE"
	MappingTypeFieldPath     MappingType = "FIELD_PATH"
	MappingTypeJoinRef       MappingType = "JOIN_REF"
)

// Transform specifications are declarative JSON data. They never contain executable code.
type DirectTransformSpec struct {
	Type MappingType `json:"type"`
}
type CastTransformSpec struct {
	Type MappingType `json:"type"`
	To   string      `json:"to"`
}
type EnumMapTransformSpec struct {
	Type    MappingType       `json:"type"`
	Values  map[string]string `json:"values"`
	Default *string           `json:"default,omitempty"`
}
type UnitConvertTransformSpec struct {
	Type   MappingType `json:"type"`
	From   string      `json:"from"`
	To     string      `json:"to"`
	Factor float64     `json:"factor"`
	Offset float64     `json:"offset,omitempty"`
}
type TimeNormalizeTransformSpec struct {
	Type         MappingType `json:"type"`
	InputFormat  string      `json:"input_format"`
	OutputFormat string      `json:"output_format"`
	Timezone     string      `json:"timezone,omitempty"`
}
type FieldPathTransformSpec struct {
	Type MappingType `json:"type"`
	Path string      `json:"path"`
}
type JoinRefTransformSpec struct {
	Type         MappingType `json:"type"`
	Relationship string      `json:"relationship"`
	FromFields   []string    `json:"from_fields"`
	ToSchema     string      `json:"to_schema"`
	ToFields     []string    `json:"to_fields"`
}

type SemanticMapping struct {
	MappingID             string
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	RequirementID         string
	SourceID              string
	ConnectionID          string
	AssetID               string
	SchemaSnapshotID      string
	TargetFieldPaths      []string
	MappingType           MappingType
	TransformSpec         json.RawMessage
	Status                MappingStatus
	Source                MappingSource
	Confidence            float64
	Reason                string
	DerivedFromMappingID  string
	AnalysisRunID         string
	CreatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type SemanticMappingAnalysisRun struct {
	AnalysisRunID         string
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	ClientRequestID       string
	RequestDigest         string
	RequestJSON           string
	Status                AnalysisRunStatus
	ModelRef              string
	ErrorKey              string
	ErrorMessageSanitized string
	RetryCount            int32
	LastRetryBy           string
	LastRetryAt           *time.Time
	ExecutionGeneration   int32
	ExecutionClaimedAt    *time.Time
	LeaseExpiresAt        *time.Time
	CreatedBy             string
	StartedAt             *time.Time
	CompletedAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type SemanticMappingDecision struct {
	DecisionID            string
	TenantID              string
	BusinessID            string
	SourceMappingID       string
	TargetMappingID       string
	Action                DecisionAction
	ActorPrincipalID      string
	Reason                string
	BusinessModelRevision int32
	CreatedAt             time.Time
}

type SemanticMappingProposal struct {
	RequirementID    string          `json:"requirement_id"`
	SourceID         string          `json:"source_id"`
	ConnectionID     string          `json:"connection_id"`
	AssetID          string          `json:"asset_id"`
	SchemaSnapshotID string          `json:"schema_snapshot_id"`
	TargetFieldPaths []string        `json:"target_field_paths"`
	MappingType      MappingType     `json:"mapping_type"`
	TransformSpec    json.RawMessage `json:"transform_spec"`
	Confidence       float64         `json:"confidence"`
	Reason           string          `json:"reason"`
	Status           MappingStatus   `json:"status,omitempty"`
	Source           MappingSource   `json:"source,omitempty"`
}
