/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "time"

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "ACTIVE"
	TenantStatusSuspended TenantStatus = "SUSPENDED"
	TenantStatusArchived  TenantStatus = "ARCHIVED"
)

type PrincipalType string

const (
	PrincipalTypeUser    PrincipalType = "USER"
	PrincipalTypeService PrincipalType = "SERVICE"
)

type PrincipalStatus string

const (
	PrincipalStatusActive    PrincipalStatus = "ACTIVE"
	PrincipalStatusSuspended PrincipalStatus = "SUSPENDED"
)

type MembershipRole string

const (
	RoleOwner  MembershipRole = "OWNER"
	RoleAdmin  MembershipRole = "ADMIN"
	RoleMember MembershipRole = "MEMBER"
	RoleViewer MembershipRole = "VIEWER"
)

type MembershipStatus string

const (
	MembershipActive    MembershipStatus = "ACTIVE"
	MembershipInvited   MembershipStatus = "INVITED"
	MembershipSuspended MembershipStatus = "SUSPENDED"
	MembershipRemoved   MembershipStatus = "REMOVED"
)

type SpacePurpose string

const (
	SpacePurposeDefault      SpacePurpose = "DEFAULT"
	SpacePurposeDevelopment  SpacePurpose = "DEVELOPMENT"
	SpacePurposeKnowledge    SpacePurpose = "KNOWLEDGE"
	SpacePurposeDelivery     SpacePurpose = "DELIVERY"
	SpacePurposeCustom       SpacePurpose = "CUSTOM"
)

type SpaceRefStatus string

const (
	SpaceRefActive   SpaceRefStatus = "ACTIVE"
	SpaceRefInactive SpaceRefStatus = "INACTIVE"
)

type Tenant struct {
	ID               int64        `json:"id"`
	TenantID         string       `json:"tenant_id"`
	TenantKey        string       `json:"tenant_key"`
	Name             string       `json:"name"`
	DisplayName      string       `json:"display_name"`
	Status           TenantStatus `json:"status"`
	OwnerPrincipalID string       `json:"owner_principal_id"`
	Revision         int32        `json:"revision"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	DeletedAt        *time.Time   `json:"deleted_at,omitempty"`
}

type Principal struct {
	ID              int64           `json:"id"`
	PrincipalID     string          `json:"principal_id"`
	PrincipalType   PrincipalType   `json:"principal_type"`
	Provider        string          `json:"provider"`
	ExternalSubject string          `json:"external_subject"`
	CozeUserID      int64           `json:"coze_user_id"`
	DisplayName     string          `json:"display_name"`
	Status          PrincipalStatus `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type Membership struct {
	ID          int64            `json:"id"`
	TenantID    string           `json:"tenant_id"`
	PrincipalID string           `json:"principal_id"`
	Role        MembershipRole   `json:"role"`
	Status      MembershipStatus `json:"status"`
	Revision    int32            `json:"revision"`
	JoinedAt    time.Time        `json:"joined_at"`
	CreatedBy   string           `json:"created_by"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type TenantSpaceRef struct {
	ID          int64          `json:"id"`
	TenantID    string         `json:"tenant_id"`
	CozeSpaceID int64          `json:"coze_space_id"`
	Purpose     SpacePurpose   `json:"purpose"`
	Status      SpaceRefStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AuditEvent records a tenancy mutation.
// PrincipalID is always the actor (who performed the action).
// Resource is the target (e.g. target principal_id, tenant_id, or coze_space_id).
// RequestID correlates with the inbound request when available.
type AuditEvent struct {
	ID          int64
	TenantID    string
	PrincipalID string // actor; not an alias for the target
	Action      string
	Resource    string // target of the action
	RequestID   string
	CreatedAt   time.Time
}

const (
	AuditTenantCreated      = "TENANT_CREATED"
	AuditTenantUpdated      = "TENANT_UPDATED"
	AuditMemberAdded        = "MEMBER_ADDED"
	AuditMemberRoleChanged  = "MEMBER_ROLE_CHANGED"
	AuditMemberRemoved      = "MEMBER_REMOVED"
	AuditSpaceBound          = "SPACE_BOUND"
	AuditSpaceUnbound       = "SPACE_UNBOUND"
)
