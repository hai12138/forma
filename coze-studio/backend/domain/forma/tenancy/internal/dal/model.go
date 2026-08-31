/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import "time"

type PrincipalModel struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PrincipalID     string    `gorm:"column:principal_id;size:64;not null"`
	PrincipalType   string    `gorm:"column:principal_type;size:32;not null"`
	Provider        string    `gorm:"column:provider;size:64;not null;default:coze"`
	ExternalSubject string    `gorm:"column:external_subject;size:255;not null"`
	CozeUserID      int64     `gorm:"column:coze_user_id;not null"`
	DisplayName     string    `gorm:"column:display_name;size:255;not null;default:''"`
	Status          string    `gorm:"column:status;size:32;not null;default:ACTIVE"`
	CreatedAt       time.Time `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null"`
}

func (PrincipalModel) TableName() string { return "forma_principal" }

type TenantModel struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID         string     `gorm:"column:tenant_id;size:64;not null"`
	TenantKey        string     `gorm:"column:tenant_key;size:128;not null"`
	Name             string     `gorm:"column:name;size:255;not null"`
	DisplayName      string     `gorm:"column:display_name;size:255;not null"`
	Status           string     `gorm:"column:status;size:32;not null;default:ACTIVE"`
	OwnerPrincipalID string     `gorm:"column:owner_principal_id;size:64;not null"`
	Revision         int32      `gorm:"column:revision;not null;default:1"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (TenantModel) TableName() string { return "forma_tenant" }

type MembershipModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID    string    `gorm:"column:tenant_id;size:64;not null"`
	PrincipalID string    `gorm:"column:principal_id;size:64;not null"`
	Role        string    `gorm:"column:role;size:32;not null"`
	Status      string    `gorm:"column:status;size:32;not null;default:ACTIVE"`
	Revision    int32     `gorm:"column:revision;not null;default:1"`
	JoinedAt    time.Time `gorm:"column:joined_at;not null"`
	CreatedBy   string    `gorm:"column:created_by;size:64;not null;default:''"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

func (MembershipModel) TableName() string { return "forma_tenant_membership" }

type TenantSpaceRefModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID    string    `gorm:"column:tenant_id;size:64;not null"`
	CozeSpaceID int64     `gorm:"column:coze_space_id;not null"`
	Purpose     string    `gorm:"column:purpose;size:64;not null;default:DEFAULT"`
	Status      string    `gorm:"column:status;size:32;not null;default:ACTIVE"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

func (TenantSpaceRefModel) TableName() string { return "forma_tenant_space_ref" }

type AuditEventModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID    string    `gorm:"column:tenant_id;size:64;not null;default:''"`
	PrincipalID string    `gorm:"column:principal_id;size:64;not null;default:''"`
	Action      string    `gorm:"column:action;size:64;not null"`
	Resource    string    `gorm:"column:resource;size:255;not null;default:''"`
	RequestID   string    `gorm:"column:request_id;size:128;not null;default:''"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
}

func (AuditEventModel) TableName() string { return "forma_audit_event" }
