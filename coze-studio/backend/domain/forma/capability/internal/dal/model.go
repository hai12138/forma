/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"time"
)

type capabilityRow struct {
	ID                  int64     `gorm:"column:id;primaryKey"`
	CapabilityID        string    `gorm:"column:capability_id"`
	TenantID            string    `gorm:"column:tenant_id"`
	BusinessID          string    `gorm:"column:business_id"`
	ActiveRevisionID    *string   `gorm:"column:active_revision_id"`
	AggregateGeneration int64     `gorm:"column:aggregate_generation"`
	CreatedBy           string    `gorm:"column:created_by"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (capabilityRow) TableName() string { return "forma_business_capability" }

type revisionRow struct {
	ID                        int64     `gorm:"column:id;primaryKey"`
	RevisionID                string    `gorm:"column:revision_id"`
	CapabilityID              string    `gorm:"column:capability_id"`
	TenantID                  string    `gorm:"column:tenant_id"`
	BusinessID                string    `gorm:"column:business_id"`
	Version                   int32     `gorm:"column:version"`
	Status                    string    `gorm:"column:status"`
	Name                      string    `gorm:"column:name"`
	Description               string    `gorm:"column:description"`
	BusinessModelRevision     int32     `gorm:"column:business_model_revision"`
	CapabilityKind            string    `gorm:"column:capability_kind"`
	InputSchemaJSON           string    `gorm:"column:input_schema_json"`
	OutputSchemaJSON          string    `gorm:"column:output_schema_json"`
	PreconditionsJSON         string    `gorm:"column:preconditions_json"`
	EffectsJSON               string    `gorm:"column:effects_json"`
	DataContractBindingsJSON  string    `gorm:"column:data_contract_bindings_json"`
	QueryOperation            string    `gorm:"column:query_operation"`
	OutputCardinality         string    `gorm:"column:output_cardinality"`
	DerivedFromRevisionID     string    `gorm:"column:derived_from_revision_id"`
	AnalysisRunID             string    `gorm:"column:analysis_run_id"`
	ProposalID                string    `gorm:"column:proposal_id"`
	Source                    string    `gorm:"column:source"`
	CreatedBy                 string    `gorm:"column:created_by"`
	CreatedAt                 time.Time `gorm:"column:created_at"`
}

func (revisionRow) TableName() string { return "forma_business_capability_revision" }

type analysisRunRow struct {
	ID                    int64      `gorm:"column:id;primaryKey"`
	AnalysisRunID         string     `gorm:"column:analysis_run_id"`
	TenantID              string     `gorm:"column:tenant_id"`
	BusinessID            string     `gorm:"column:business_id"`
	BusinessModelRevision int32      `gorm:"column:business_model_revision"`
	ClientRequestID       string     `gorm:"column:client_request_id"`
	RequestDigest         string     `gorm:"column:request_digest"`
	Status                string     `gorm:"column:status"`
	Attempt               int32      `gorm:"column:attempt"`
	ErrorCode             string     `gorm:"column:error_code"`
	ModelRef              string     `gorm:"column:model_ref"`
	RequestJSON           string     `gorm:"column:request_json"`
	ExecutionClaimedAt    *time.Time `gorm:"column:execution_claimed_at"`
	LeaseExpiresAt        *time.Time `gorm:"column:lease_expires_at"`
	CreatedBy             string     `gorm:"column:created_by"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (analysisRunRow) TableName() string { return "forma_capability_analysis_run" }

type proposalRow struct {
	ID                     int64     `gorm:"column:id;primaryKey"`
	ProposalID             string    `gorm:"column:proposal_id"`
	TenantID               string    `gorm:"column:tenant_id"`
	BusinessID             string    `gorm:"column:business_id"`
	AnalysisRunID          string    `gorm:"column:analysis_run_id"`
	CapabilityID           *string   `gorm:"column:capability_id"`
	Status                 string    `gorm:"column:status"`
	PayloadJSON            string    `gorm:"column:payload_json"`
	MaterializedRevisionID *string   `gorm:"column:materialized_revision_id"`
	CreatedAt              time.Time `gorm:"column:created_at"`
}

func (proposalRow) TableName() string { return "forma_capability_proposal" }

type decisionRow struct {
	ID               int64     `gorm:"column:id;primaryKey"`
	DecisionID       string    `gorm:"column:decision_id"`
	TenantID         string    `gorm:"column:tenant_id"`
	BusinessID       string    `gorm:"column:business_id"`
	CapabilityID     *string   `gorm:"column:capability_id"`
	ProposalID       *string   `gorm:"column:proposal_id"`
	SourceRevisionID *string   `gorm:"column:source_revision_id"`
	TargetRevisionID *string   `gorm:"column:target_revision_id"`
	Action           string    `gorm:"column:action"`
	PayloadDigest    string    `gorm:"column:payload_digest"`
	ClientRequestID  *string   `gorm:"column:client_request_id"`
	ActorPrincipalID string    `gorm:"column:actor_principal_id"`
	Reason           string    `gorm:"column:reason"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (decisionRow) TableName() string { return "forma_capability_decision" }
