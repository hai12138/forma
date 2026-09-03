/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

type DataSourceService interface {
	CreateSource(context.Context, *CreateSourceInput) (*entity.DataSource, error)
	GetSource(context.Context, string, string) (*entity.DataSource, error)
	ListSources(context.Context, string) ([]*entity.DataSource, error)
	PatchSource(context.Context, *PatchSourceInput) (*entity.DataSource, error)
	ArchiveSource(context.Context, string, string) (*entity.DataSource, error)
	CreateConnection(context.Context, *CreateConnectionInput) (*entity.DataConnection, error)
	GetConnection(context.Context, string, string, string) (*entity.DataConnection, error)
	ListConnections(context.Context, string, string) ([]*entity.DataConnection, error)
	PatchConnection(context.Context, *PatchConnectionInput) (*entity.DataConnection, error)
	CreateCredential(context.Context, *CreateCredentialInput) (*entity.CredentialRef, error)
	RotateCredential(context.Context, string, string, []byte) (*entity.CredentialRef, error)
	RevokeCredential(context.Context, string, string) (*entity.CredentialRef, error)
	TestConnection(context.Context, string, string, string) error
	Discover(context.Context, string, string, string) ([]*entity.DataAsset, error)
	GetAsset(context.Context, string, string) (*entity.DataAsset, error)
	ListAssets(context.Context, string, string) ([]*entity.DataAsset, error)
	CaptureSchema(context.Context, string, string, string, string, string) (*entity.SchemaSnapshot, error)
	GetSnapshot(context.Context, string, string) (*entity.SchemaSnapshot, error)
	ListSnapshotsByAsset(ctx context.Context, tenantID, sourceID, connectionID, assetID string) ([]*entity.SchemaSnapshot, error)
}

type SourceComponents struct {
	Repo     repository.DataSourceRepository
	Secrets  SecretProvider
	Adapters *AdapterRegistry
}
type dataSourceService struct {
	repo     repository.DataSourceRepository
	secrets  SecretProvider
	adapters *AdapterRegistry
}

func NewDataSourceService(c *SourceComponents) DataSourceService {
	if c == nil {
		return &dataSourceService{}
	}
	return &dataSourceService{repo: c.Repo, secrets: c.Secrets, adapters: c.Adapters}
}

type CreateSourceInput struct {
	TenantID, Name, ActorID string
	SourceType              entity.SourceType
}
type PatchSourceInput struct {
	TenantID, SourceID, Name string
	Status                   entity.DataSourceStatus
}
type CreateConnectionInput struct {
	TenantID, SourceID, Name, PublicConfigJSON, CredentialRefID, ActorID string
	Environment                                                          entity.Environment
	AdapterType                                                          entity.AdapterType
}
type PatchConnectionInput struct {
	TenantID, SourceID, ConnectionID, Name, PublicConfigJSON, CredentialRefID string
	Environment                                                               entity.Environment
}
type CreateCredentialInput struct {
	TenantID, SecretType, ActorID string
	Secret                        []byte
}

func (s *dataSourceService) ready() bool { return s != nil && s.repo != nil }
func validSourceType(v entity.SourceType) bool {
	return v == entity.SourceTypeRelationalDatabase || v == entity.SourceTypeHTTPAPI
}
func validEnvironment(v entity.Environment) bool {
	return v == entity.EnvironmentDev || v == entity.EnvironmentTest || v == entity.EnvironmentProd
}
func validSecret(secret []byte) bool {
	if len(secret) == 0 || !json.Valid(secret) {
		return false
	}
	var object map[string]any
	return json.Unmarshal(secret, &object) == nil && len(object) > 0
}

func (s *dataSourceService) CreateSource(ctx context.Context, in *CreateSourceInput) (*entity.DataSource, error) {
	if !s.ready() || in == nil || in.TenantID == "" || strings.TrimSpace(in.Name) == "" {
		return nil, entity.ErrPublicConfigInvalid
	}
	if !validSourceType(in.SourceType) {
		return nil, entity.ErrSourceTypeNotSupported
	}
	now := time.Now().UTC()
	v := &entity.DataSource{SourceID: uuid.NewString(), TenantID: in.TenantID, Name: strings.TrimSpace(in.Name), SourceType: in.SourceType, Status: entity.DataSourceActive, CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateSource(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (s *dataSourceService) GetSource(ctx context.Context, t, id string) (*entity.DataSource, error) {
	return s.repo.GetSource(ctx, t, id)
}
func (s *dataSourceService) ListSources(ctx context.Context, t string) ([]*entity.DataSource, error) {
	return s.repo.ListSources(ctx, t)
}
func (s *dataSourceService) PatchSource(ctx context.Context, in *PatchSourceInput) (*entity.DataSource, error) {
	if in == nil {
		return nil, entity.ErrPublicConfigInvalid
	}
	v, e := s.repo.GetSource(ctx, in.TenantID, in.SourceID)
	if e != nil {
		return nil, e
	}
	if strings.TrimSpace(in.Name) != "" {
		v.Name = strings.TrimSpace(in.Name)
	}
	if in.Status != "" {
		if in.Status != entity.DataSourceActive && in.Status != entity.DataSourceDisabled {
			return nil, entity.ErrPublicConfigInvalid
		}
		v.Status = in.Status
	}
	v.UpdatedAt = time.Now().UTC()
	if e = s.repo.UpdateSource(ctx, v); e != nil {
		return nil, e
	}
	return v, nil
}
func (s *dataSourceService) ArchiveSource(ctx context.Context, t, id string) (*entity.DataSource, error) {
	v, e := s.repo.GetSource(ctx, t, id)
	if e != nil {
		return nil, e
	}
	v.Status = entity.DataSourceArchived
	v.UpdatedAt = time.Now().UTC()
	if e = s.repo.UpdateSource(ctx, v); e != nil {
		return nil, e
	}
	return v, nil
}

func (s *dataSourceService) CreateConnection(ctx context.Context, in *CreateConnectionInput) (*entity.DataConnection, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !validEnvironment(in.Environment) {
		return nil, entity.ErrPublicConfigInvalid
	}
	src, e := s.repo.GetSource(ctx, in.TenantID, in.SourceID)
	if e != nil {
		return nil, e
	}
	if src.Status == entity.DataSourceArchived {
		return nil, entity.ErrDataSourceNotFound
	}
	if e = ValidatePublicConfig(in.PublicConfigJSON); e != nil {
		return nil, e
	}
	if (src.SourceType == entity.SourceTypeHTTPAPI && in.AdapterType != entity.AdapterHTTP) ||
		(src.SourceType == entity.SourceTypeRelationalDatabase && in.AdapterType == entity.AdapterHTTP) {
		return nil, entity.ErrDataAdapterNotSupported
	}
	if _, e = s.adapters.Get(in.AdapterType); e != nil {
		return nil, e
	}
	if in.CredentialRefID != "" {
		cred, ce := s.repo.GetCredential(ctx, in.TenantID, in.CredentialRefID)
		if ce != nil {
			return nil, ce
		}
		if cred.Status != entity.CredentialActive {
			return nil, entity.ErrDataCredentialNotFound
		}
	}
	now := time.Now().UTC()
	v := &entity.DataConnection{ConnectionID: uuid.NewString(), TenantID: in.TenantID, SourceID: in.SourceID, Name: strings.TrimSpace(in.Name), Environment: in.Environment, AdapterType: in.AdapterType, PublicConfigJSON: in.PublicConfigJSON, CredentialRefID: in.CredentialRefID, Status: entity.DataConnectionInactive, CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now}
	if e = s.repo.CreateConnection(ctx, v); e != nil {
		return nil, e
	}
	return v, nil
}
func (s *dataSourceService) GetConnection(ctx context.Context, t, sourceID, id string) (*entity.DataConnection, error) {
	v, e := s.repo.GetConnection(ctx, t, id)
	if e != nil {
		return nil, e
	}
	if v.SourceID != sourceID {
		return nil, entity.ErrDataConnectionNotFound
	}
	return v, nil
}
func (s *dataSourceService) ListConnections(ctx context.Context, t, sourceID string) ([]*entity.DataConnection, error) {
	if _, e := s.repo.GetSource(ctx, t, sourceID); e != nil {
		return nil, e
	}
	return s.repo.ListConnections(ctx, t, sourceID)
}
func (s *dataSourceService) PatchConnection(ctx context.Context, in *PatchConnectionInput) (*entity.DataConnection, error) {
	if in == nil {
		return nil, entity.ErrPublicConfigInvalid
	}
	v, e := s.GetConnection(ctx, in.TenantID, in.SourceID, in.ConnectionID)
	if e != nil {
		return nil, e
	}
	if strings.TrimSpace(in.Name) != "" {
		v.Name = strings.TrimSpace(in.Name)
	}
	if in.Environment != "" {
		if !validEnvironment(in.Environment) {
			return nil, entity.ErrPublicConfigInvalid
		}
		v.Environment = in.Environment
	}
	if in.PublicConfigJSON != "" {
		if e = ValidatePublicConfig(in.PublicConfigJSON); e != nil {
			return nil, e
		}
		v.PublicConfigJSON = in.PublicConfigJSON
	}
	if in.CredentialRefID != "" {
		cred, ce := s.repo.GetCredential(ctx, in.TenantID, in.CredentialRefID)
		if ce != nil || cred.Status != entity.CredentialActive {
			return nil, entity.ErrDataCredentialNotFound
		}
		v.CredentialRefID = in.CredentialRefID
	}
	v.UpdatedAt = time.Now().UTC()
	if e = s.repo.UpdateConnection(ctx, v); e != nil {
		return nil, e
	}
	return v, nil
}

func (s *dataSourceService) CreateCredential(ctx context.Context, in *CreateCredentialInput) (*entity.CredentialRef, error) {
	if s.secrets == nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	if in == nil || in.TenantID == "" || strings.TrimSpace(in.SecretType) == "" || !validSecret(in.Secret) {
		return nil, entity.ErrSecretInvalid
	}
	now := time.Now().UTC()
	v := &entity.CredentialRef{CredentialRefID: uuid.NewString(), TenantID: in.TenantID, Provider: entity.ProviderLocalEncrypted, SecretType: in.SecretType, KeyVersion: 1, Status: entity.CredentialActive, CreatedBy: in.ActorID, CreatedAt: now}
	encrypted, e := s.secrets.StoreSecret(ctx, v.TenantID, v.CredentialRefID, v.KeyVersion, in.Secret)
	if e != nil {
		return nil, e
	}
	defer wipe(in.Secret)
	if e = s.repo.CreateCredential(ctx, v, encrypted); e != nil {
		return nil, e
	}
	return v, nil
}
func (s *dataSourceService) RotateCredential(ctx context.Context, t, id string, secret []byte) (*entity.CredentialRef, error) {
	if s.secrets == nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	v, e := s.repo.GetCredential(ctx, t, id)
	if e != nil {
		return nil, e
	}
	if v.Status != entity.CredentialActive || !validSecret(secret) {
		return nil, entity.ErrSecretInvalid
	}
	v.KeyVersion++
	now := time.Now().UTC()
	v.RotatedAt = &now
	encrypted, e := s.secrets.RotateSecret(ctx, t, id, v.KeyVersion, secret)
	defer wipe(secret)
	if e != nil {
		return nil, e
	}
	if e = s.repo.UpdateCredential(ctx, v, encrypted); e != nil {
		return nil, e
	}
	return v, nil
}
func (s *dataSourceService) RevokeCredential(ctx context.Context, t, id string) (*entity.CredentialRef, error) {
	if s.secrets == nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	v, e := s.repo.GetCredential(ctx, t, id)
	if e != nil {
		return nil, e
	}
	encrypted, _ := s.repo.GetSecret(ctx, t, id)
	_ = s.secrets.DeleteSecret(ctx, t, id, v.KeyVersion, encrypted)
	if e = s.repo.DeleteSecret(ctx, t, id); e != nil {
		return nil, e
	}
	now := time.Now().UTC()
	v.Status = entity.CredentialRevoked
	v.RevokedAt = &now
	if e = s.repo.UpdateCredential(ctx, v, nil); e != nil {
		return nil, e
	}
	return v, nil
}
func wipe(v []byte) {
	for i := range v {
		v[i] = 0
	}
}

func (s *dataSourceService) adapterRequest(ctx context.Context, t, sourceID, connectionID string) (DataSourceAdapter, *AdapterRequest, *entity.DataConnection, error) {
	conn, e := s.GetConnection(ctx, t, sourceID, connectionID)
	if e != nil {
		return nil, nil, nil, e
	}
	adapter, e := s.adapters.Get(conn.AdapterType)
	if e != nil {
		return nil, nil, nil, e
	}
	req := &AdapterRequest{PublicConfigJSON: conn.PublicConfigJSON}
	if conn.CredentialRefID != "" {
		if s.secrets == nil {
			return nil, nil, nil, entity.ErrSecretProviderUnavailable
		}
		cred, ce := s.repo.GetCredential(ctx, t, conn.CredentialRefID)
		if ce != nil || cred.Status != entity.CredentialActive {
			return nil, nil, nil, entity.ErrDataCredentialNotFound
		}
		encrypted, ce := s.repo.GetSecret(ctx, t, cred.CredentialRefID)
		if ce != nil {
			return nil, nil, nil, ce
		}
		req.Secret, ce = s.secrets.ResolveSecret(ctx, t, cred.CredentialRefID, cred.KeyVersion, encrypted)
		if ce != nil {
			return nil, nil, nil, ce
		}
	}
	return adapter, req, conn, nil
}
func (s *dataSourceService) TestConnection(ctx context.Context, t, sourceID, connectionID string) error {
	a, r, conn, e := s.adapterRequest(ctx, t, sourceID, connectionID)
	if e != nil {
		return e
	}
	defer wipe(r.Secret)
	now := time.Now().UTC()
	if e = a.TestConnection(ctx, r); e != nil {
		conn.LastTestStatus = entity.ConnectionTestFailed
		conn.LastTestAt = &now
		conn.LastTestErrorKey = "FORMA_DATA_CONNECTION_FAILED"
		conn.UpdatedAt = now
		_ = s.repo.UpdateConnection(ctx, conn)
		return entity.ErrDataConnectionFailed
	}
	conn.Status = entity.DataConnectionActive
	conn.LastTestStatus = entity.ConnectionTestHealthy
	conn.LastTestAt = &now
	conn.LastTestErrorKey = ""
	conn.UpdatedAt = now
	if e = s.repo.UpdateConnection(ctx, conn); e != nil {
		return e
	}
	return nil
}
func canonicalJSON(v any) (string, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return "", entity.ErrPublicConfigInvalid
	}
	var normalized any
	if e = json.Unmarshal(b, &normalized); e != nil {
		return "", entity.ErrPublicConfigInvalid
	}
	b, e = json.Marshal(normalized)
	return string(b), e
}
func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func (s *dataSourceService) Discover(ctx context.Context, t, sourceID, connectionID string) ([]*entity.DataAsset, error) {
	a, r, _, e := s.adapterRequest(ctx, t, sourceID, connectionID)
	if e != nil {
		return nil, e
	}
	defer wipe(r.Secret)
	found, e := a.DiscoverAssets(ctx, r)
	if e != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	out := make([]*entity.DataAsset, 0, len(found))
	for _, d := range found {
		locator, ce := canonicalJSON(d.PhysicalLocator)
		if ce != nil {
			return nil, entity.ErrDataDiscoveryFailed
		}
		now := time.Now().UTC()
		asset := &entity.DataAsset{AssetID: uuid.NewString(), TenantID: t, SourceID: sourceID, ConnectionID: connectionID, AssetType: d.AssetType, Name: d.Name, PhysicalLocatorJSON: locator, LocatorDigest: digest(locator), CreatedAt: now, UpdatedAt: now}
		stored, _, ue := s.repo.UpsertAsset(ctx, asset)
		if ue != nil {
			return nil, ue
		}
		out = append(out, stored)
	}
	return out, nil
}
func (s *dataSourceService) GetAsset(ctx context.Context, t, id string) (*entity.DataAsset, error) {
	return s.repo.GetAsset(ctx, t, id)
}
func (s *dataSourceService) ListAssets(ctx context.Context, t, sourceID string) ([]*entity.DataAsset, error) {
	if _, e := s.repo.GetSource(ctx, t, sourceID); e != nil {
		return nil, e
	}
	return s.repo.ListAssets(ctx, t, sourceID)
}
func (s *dataSourceService) CaptureSchema(ctx context.Context, t, sourceID, connectionID, assetID, actorID string) (*entity.SchemaSnapshot, error) {
	asset, e := s.repo.GetAsset(ctx, t, assetID)
	if e != nil {
		return nil, e
	}
	if asset.SourceID != sourceID || asset.ConnectionID != connectionID {
		return nil, entity.ErrDataAssetNotFound
	}
	var locator map[string]any
	if json.Unmarshal([]byte(asset.PhysicalLocatorJSON), &locator) != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	a, r, _, e := s.adapterRequest(ctx, t, sourceID, connectionID)
	if e != nil {
		return nil, e
	}
	defer wipe(r.Secret)
	schema, e := a.GetSchema(ctx, r, locator)
	if e != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	schemaJSON, e := canonicalJSON(schema)
	if e != nil {
		return nil, e
	}
	v := &entity.SchemaSnapshot{SnapshotID: uuid.NewString(), TenantID: t, SourceID: sourceID, ConnectionID: connectionID, AssetID: assetID, SchemaJSON: schemaJSON, Fingerprint: digest(schemaJSON), CreatedBy: actorID, CreatedAt: time.Now().UTC()}
	if e = s.repo.CreateSnapshot(ctx, v); e != nil {
		return nil, e
	}
	return v, nil
}
func (s *dataSourceService) GetSnapshot(ctx context.Context, t, id string) (*entity.SchemaSnapshot, error) {
	return s.repo.GetSnapshot(ctx, t, id)
}
func (s *dataSourceService) ListSnapshotsByAsset(ctx context.Context, tenantID, sourceID, connectionID, assetID string) ([]*entity.SchemaSnapshot, error) {
	if tenantID == "" || sourceID == "" || connectionID == "" || assetID == "" {
		return nil, entity.ErrDataAssetNotFound
	}
	asset, err := s.repo.GetAsset(ctx, tenantID, assetID)
	if err != nil {
		return nil, err
	}
	if asset.SourceID != sourceID || asset.ConnectionID != connectionID {
		return nil, entity.ErrDataAssetNotFound
	}
	return s.repo.ListSnapshotsByAsset(ctx, tenantID, sourceID, connectionID, assetID)
}
