/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import "context"

// BusinessModelValidationPort loads BM pin evidence (fingerprint only).
// Capability must not import business repository/DAL — adapters live in crossdomain.
type BusinessModelValidationPort interface {
	GetBusinessModelRevision(ctx context.Context, tenantID, businessID string, revision int32) (*BusinessModelRevisionEvidence, error)
}

// ContractValidationPort loads Active contract logical descriptors for validation.
type ContractValidationPort interface {
	GetActiveContractLogicalDescriptor(ctx context.Context, tenantID, businessID, contractID string) (*ContractLogicalDescriptor, error)
}

// BusinessModelRevisionEvidence is capability-owned BM pin evidence (no Semantic Model JSON).
type BusinessModelRevisionEvidence struct {
	TenantID      string
	BusinessID    string
	Revision      int32
	ContentDigest string
}

// ContractLogicalField is a consumer-facing logical field (no physical/source/credential).
type ContractLogicalField struct {
	LogicalKey  string
	LogicalType string
	Nullable    bool
}

// ContractFilterFieldSpec is a filter allowlist entry for one logical key.
type ContractFilterFieldSpec struct {
	LogicalKey string
	Operators  []string
}

// ContractSortFieldSpec is a sort allowlist entry.
type ContractSortFieldSpec struct {
	LogicalKey string
	Directions []string
}

// ContractPaginationPolicy mirrors S4 consumer pagination limits.
type ContractPaginationPolicy struct {
	DefaultLimit int32
	MaxLimit     int32
}

// ContractLogicalDescriptor is capability-owned Active contract evidence (logical only).
// TenantID + BusinessID MUST be filled by the adapter from the contract root for isolation checks.
type ContractLogicalDescriptor struct {
	TenantID              string
	BusinessID            string
	ContractID            string
	RevisionID            string
	Version               int32
	BusinessModelRevision int32
	Status                string
	LogicalSchema         []ContractLogicalField
	QueryCapabilities     []string
	FilterSchema          []ContractFilterFieldSpec
	SortSchema            []ContractSortFieldSpec
	PaginationPolicy      ContractPaginationPolicy
	FreshnessPolicy       string
	Classification        map[string]string
	AccessPolicyRef       string
}
