/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "time"

type ContractStatus string

const (
	ContractStatusDraft      ContractStatus = "DRAFT"
	ContractStatusValidated  ContractStatus = "VALIDATED"
	ContractStatusActive     ContractStatus = "ACTIVE"
	ContractStatusStale      ContractStatus = "STALE"
	ContractStatusDeprecated ContractStatus = "DEPRECATED"
)

type DataClassification string

const (
	DataClassificationPublic       DataClassification = "PUBLIC"
	DataClassificationInternal     DataClassification = "INTERNAL"
	DataClassificationConfidential DataClassification = "CONFIDENTIAL"
	DataClassificationPII          DataClassification = "PII"
	DataClassificationSecret       DataClassification = "SECRET"
)

type QueryCapability string

const (
	QueryCapabilityRead      QueryCapability = "READ"
	QueryCapabilityLookup    QueryCapability = "LOOKUP"
	QueryCapabilityList      QueryCapability = "LIST"
	QueryCapabilityFilter    QueryCapability = "FILTER"
	QueryCapabilityAggregate QueryCapability = "AGGREGATE"
)

type FilterOperator string

const (
	FilterOperatorEQ       FilterOperator = "EQ"
	FilterOperatorNE       FilterOperator = "NE"
	FilterOperatorGT       FilterOperator = "GT"
	FilterOperatorGTE      FilterOperator = "GTE"
	FilterOperatorLT       FilterOperator = "LT"
	FilterOperatorLTE      FilterOperator = "LTE"
	FilterOperatorIN       FilterOperator = "IN"
	FilterOperatorBetween  FilterOperator = "BETWEEN"
	FilterOperatorContains FilterOperator = "CONTAINS"
)

type SortDirection string

const (
	SortDirectionASC  SortDirection = "ASC"
	SortDirectionDESC SortDirection = "DESC"
)

type FreshnessPolicy string

const (
	FreshnessPolicyRealtime     FreshnessPolicy = "REALTIME"
	FreshnessPolicyNearRealtime FreshnessPolicy = "NEAR_REALTIME"
	FreshnessPolicyHourly       FreshnessPolicy = "HOURLY"
	FreshnessPolicyDaily        FreshnessPolicy = "DAILY"
	FreshnessPolicyOnDemand     FreshnessPolicy = "ON_DEMAND"
)

type ValidationStatus string

const (
	ValidationStatusPass ValidationStatus = "PASS"
	ValidationStatusFail ValidationStatus = "FAIL"
)

type DriftSeverity string

const (
	DriftSeverityNoChange   DriftSeverity = "NO_CHANGE"
	DriftSeverityCompatible DriftSeverity = "COMPATIBLE"
	DriftSeverityBreaking   DriftSeverity = "BREAKING"
)

type LifecycleAction string

const (
	LifecycleActionValidatePass LifecycleAction = "VALIDATE_PASS"
	LifecycleActionValidateFail LifecycleAction = "VALIDATE_FAIL"
	LifecycleActionActivate     LifecycleAction = "ACTIVATE"
	LifecycleActionMarkStale    LifecycleAction = "MARK_STALE"
	LifecycleActionDeprecate    LifecycleAction = "DEPRECATE"
)

type DataContract struct {
	ContractID       string
	TenantID         string
	BusinessID       string
	ActiveRevisionID string
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type LogicalField struct {
	LogicalKey     string             `json:"logical_key"`
	SemanticName   string             `json:"semantic_name"`
	LogicalType    string             `json:"logical_type"`
	Description    string             `json:"description"`
	RequirementID  string             `json:"requirement_id"`
	Nullable       bool               `json:"nullable"`
	Classification DataClassification `json:"classification"`
}

type ContractLogicalSchema struct {
	Fields []LogicalField `json:"fields"`
}

type ContractBinding struct {
	RequirementID    string `json:"requirement_id"`
	MappingID        string `json:"mapping_id"`
	SourceID         string `json:"source_id"`
	ConnectionID     string `json:"connection_id"`
	AssetID          string `json:"asset_id"`
	SchemaSnapshotID string `json:"schema_snapshot_id"`
}

type FilterFieldSpec struct {
	LogicalKey string           `json:"logical_key"`
	Operators  []FilterOperator `json:"operators"`
}

type FilterSchema struct {
	Fields []FilterFieldSpec `json:"fields"`
}

type SortFieldSpec struct {
	LogicalKey string          `json:"logical_key"`
	Directions []SortDirection `json:"directions"`
}

type SortSchema struct {
	Fields []SortFieldSpec `json:"fields"`
}

type PaginationPolicy struct {
	DefaultLimit int32 `json:"default_limit"`
	MaxLimit     int32 `json:"max_limit"`
}

type DataContractRevision struct {
	RevisionID             string
	TenantID               string
	BusinessID             string
	ContractID             string
	Version                int32
	Status                 ContractStatus
	BusinessModelRevision  int32
	Name                   string
	Description            string
	RequirementIDs         []string
	LogicalSchema          ContractLogicalSchema
	QueryCapabilities      []QueryCapability
	FilterSchema           FilterSchema
	SortSchema             SortSchema
	PaginationPolicy       PaginationPolicy
	FreshnessPolicy        FreshnessPolicy
	ClassificationPolicy   map[string]DataClassification
	BindingRefs            []ContractBinding
	AccessPolicyRef        string
	DerivedFromRevisionID  string
	CreatedBy              string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type ValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DataValidationResult struct {
	ValidationID         string
	TenantID             string
	BusinessID           string
	ContractID           string
	RevisionID           string
	Version              int32
	Status               ValidationStatus
	Errors               []ValidationIssue
	Warnings             []ValidationIssue
	SnapshotFingerprints map[string]string
	ValidatedBy          string
	ValidatedAt          time.Time
	CreatedAt            time.Time
}

type DataContractLifecycleEvent struct {
	EventID          string
	TenantID         string
	BusinessID       string
	ContractID       string
	RevisionID       string
	Version          int32
	Action           LifecycleAction
	ActorPrincipalID string
	Reason           string
	CreatedAt        time.Time
}

type DriftFinding struct {
	Code             string `json:"code"`
	Message          string `json:"message"`
	BindingMappingID string `json:"binding_mapping_id"`
	FieldPath        string `json:"field_path"`
}

type DataDriftResult struct {
	DriftResultID        string
	TenantID             string
	BusinessID           string
	ContractID           string
	RevisionID           string
	Version              int32
	Severity             DriftSeverity
	Findings             []DriftFinding
	ComparedSnapshotIDs  map[string]string
	EvaluatedBy          string
	EvaluatedAt          time.Time
	CreatedAt            time.Time
}

type DataContractGapResult struct {
	GapResultID               string
	TenantID                  string
	BusinessID                string
	ContractID                string
	RevisionID                string
	Version                   int32
	FromBusinessRevision      int32
	CurrentBusinessRevision   int32
	NewConfirmedRequirementIDs []string
	UnmappedRequirementIDs    []string
	GapStatus                 string
	EvaluatedBy               string
	EvaluatedAt               time.Time
	CreatedAt                 time.Time
}
