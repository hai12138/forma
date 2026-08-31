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
	ID                int64
	TenantID          string
	TenantKey         string
	Name              string
	DisplayName       string
	Status            TenantStatus
	OwnerPrincipalID  string
	Revision          int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type Principal struct {
	ID              int64
	PrincipalID     string
	PrincipalType   PrincipalType
	Provider        string
	ExternalSubject string
	CozeUserID      int64
	DisplayName     string
	Status          PrincipalStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Membership struct {
	ID          int64
	TenantID    string
	PrincipalID string
	Role        MembershipRole
	Status      MembershipStatus
	Revision    int32
	JoinedAt    time.Time
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TenantSpaceRef struct {
	ID          int64
	TenantID    string
	CozeSpaceID int64
	Purpose     SpacePurpose
	Status      SpaceRefStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AuditEvent struct {
	ID          int64
	TenantID    string
	PrincipalID string
	Action      string
	Resource    string
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
