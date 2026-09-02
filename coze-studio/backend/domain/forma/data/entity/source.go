/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "time"

type SourceType string

const (
	SourceTypeRelationalDatabase SourceType = "RELATIONAL_DATABASE"
	SourceTypeHTTPAPI            SourceType = "HTTP_API"
	SourceTypeFileStorage        SourceType = "FILE_STORAGE"
	SourceTypeObjectStorage      SourceType = "OBJECT_STORAGE"
	SourceTypeManagedData        SourceType = "MANAGED_DATA"
	SourceTypeEventStream        SourceType = "EVENT_STREAM"
	SourceTypeCustomAdapter      SourceType = "CUSTOM_ADAPTER"
)

type DataSourceStatus string

const (
	DataSourceActive   DataSourceStatus = "ACTIVE"
	DataSourceDisabled DataSourceStatus = "DISABLED"
	DataSourceArchived DataSourceStatus = "ARCHIVED"
)

type Environment string

const (
	EnvironmentDev  Environment = "DEV"
	EnvironmentTest Environment = "TEST"
	EnvironmentProd Environment = "PROD"
)

type AdapterType string

const (
	AdapterMySQL      AdapterType = "MYSQL"
	AdapterPostgreSQL AdapterType = "POSTGRESQL"
	AdapterHTTP       AdapterType = "HTTP"
)

type CredentialProvider string

const ProviderLocalEncrypted CredentialProvider = "LOCAL_ENCRYPTED"

type CredentialStatus string

const (
	CredentialActive  CredentialStatus = "ACTIVE"
	CredentialRevoked CredentialStatus = "REVOKED"
)

type AssetType string

const (
	AssetDataset   AssetType = "DATASET"
	AssetEntitySet AssetType = "ENTITY_SET"
	AssetEndpoint  AssetType = "ENDPOINT"
)

type DataSource struct {
	SourceID   string
	TenantID   string
	Name       string
	SourceType SourceType
	Status     DataSourceStatus
	CreatedBy  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DataConnection struct {
	ConnectionID     string
	TenantID         string
	SourceID         string
	Name             string
	Environment      Environment
	AdapterType      AdapterType
	PublicConfigJSON string
	CredentialRefID  string
	Status           DataConnectionStatus
	LastTestStatus   ConnectionTestStatus
	LastTestAt       *time.Time
	LastTestErrorKey string
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type DataConnectionStatus string

const (
	DataConnectionActive   DataConnectionStatus = "ACTIVE"
	DataConnectionInactive DataConnectionStatus = "INACTIVE"
)

type ConnectionTestStatus string

const (
	ConnectionTestHealthy ConnectionTestStatus = "HEALTHY"
	ConnectionTestFailed  ConnectionTestStatus = "FAILED"
)

// CredentialRef is metadata only. Secret material is deliberately stored
// separately and is never represented by this entity.
type CredentialRef struct {
	CredentialRefID string
	TenantID        string
	Provider        CredentialProvider
	SecretType      string
	KeyVersion      int32
	Status          CredentialStatus
	CreatedBy       string
	CreatedAt       time.Time
	RotatedAt       *time.Time
	RevokedAt       *time.Time
}

type DataAsset struct {
	AssetID             string
	TenantID            string
	SourceID            string
	ConnectionID        string
	AssetType           AssetType
	Name                string
	PhysicalLocatorJSON string
	LocatorDigest       string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SchemaSnapshot is append-only. Repositories expose Create and Get only.
type SchemaSnapshot struct {
	SnapshotID   string
	TenantID     string
	SourceID     string
	ConnectionID string
	AssetID      string
	SchemaJSON   string
	Fingerprint  string
	CreatedBy    string
	CreatedAt    time.Time
}
