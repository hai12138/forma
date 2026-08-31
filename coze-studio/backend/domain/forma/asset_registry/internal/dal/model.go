/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import "time"

type AssetRefModel struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID        string     `gorm:"column:tenant_id;size:64;not null"`
	AssetID         string     `gorm:"column:asset_id;size:64;not null"`
	Kind            string     `gorm:"column:kind;size:32;not null"`
	Name            string     `gorm:"column:name;size:255;not null"`
	SemanticVersion string     `gorm:"column:semantic_version;size:64;not null"`
	Revision        int32      `gorm:"column:revision;not null"`
	SchemaVersion   string     `gorm:"column:schema_version;size:32;not null"`
	Status          string     `gorm:"column:status;size:32;not null"`
	OwnerID         int64      `gorm:"column:owner_id;not null"`
	CreatedBy       int64      `gorm:"column:created_by;not null"`
	ContentDigest   *string    `gorm:"column:content_digest;size:128"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (AssetRefModel) TableName() string { return "forma_asset_ref" }

type CozeResourceRefModel struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID         string    `gorm:"column:tenant_id;size:64;not null"`
	AssetID          string    `gorm:"column:asset_id;size:64;not null"`
	AssetRevision    int32     `gorm:"column:asset_revision;not null"`
	CozeResourceType string    `gorm:"column:coze_resource_type;size:32;not null"`
	CozeResourceID   int64     `gorm:"column:coze_resource_id;not null"`
	CozeSpaceID      *int64    `gorm:"column:coze_space_id"`
	CozeVersion      *string   `gorm:"column:coze_version;size:64"`
	CreatedAt        time.Time `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time `gorm:"column:updated_at;not null"`
}

func (CozeResourceRefModel) TableName() string { return "forma_coze_resource_ref" }
