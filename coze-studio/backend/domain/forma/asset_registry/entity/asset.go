/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "time"

type AssetKind string

const (
	AssetKindBusiness     AssetKind = "BUSINESS"
	AssetKindCapability   AssetKind = "CAPABILITY"
	AssetKindAgent        AssetKind = "AGENT"
	AssetKindApplication  AssetKind = "APPLICATION"
)

type AssetStatus string

const (
	AssetStatusDraft      AssetStatus = "DRAFT"
	AssetStatusInReview   AssetStatus = "IN_REVIEW"
	AssetStatusVerified   AssetStatus = "VERIFIED"
	AssetStatusFrozen     AssetStatus = "FROZEN"
	AssetStatusReleased   AssetStatus = "RELEASED"
	AssetStatusDeprecated AssetStatus = "DEPRECATED"
	AssetStatusArchived   AssetStatus = "ARCHIVED"
)

type CozeResourceType string

const (
	CozeResourceTypeAgent     CozeResourceType = "AGENT"
	CozeResourceTypeWorkflow  CozeResourceType = "WORKFLOW"
	CozeResourceTypePlugin    CozeResourceType = "PLUGIN"
	CozeResourceTypeKnowledge CozeResourceType = "KNOWLEDGE"
	CozeResourceTypeApp       CozeResourceType = "APP"
	CozeResourceTypeDatabase  CozeResourceType = "DATABASE"
)

// AssetRef is the Forma asset header stored in forma_asset_ref.
type AssetRef struct {
	ID              int64
	TenantID        string
	AssetID         string
	Kind            AssetKind
	Name            string
	SemanticVersion string
	Revision        int32
	SchemaVersion   string
	Status          AssetStatus
	OwnerID         int64
	CreatedBy       int64
	ContentDigest   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// CozeResourceRef maps a Forma asset revision to a Coze resource by stable ID reference.
type CozeResourceRef struct {
	ID               int64
	TenantID         string
	AssetID          string
	AssetRevision    int32
	CozeResourceType CozeResourceType
	CozeResourceID   int64
	CozeSpaceID      *int64
	CozeVersion      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
