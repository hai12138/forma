/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"encoding/json"
	"time"

	datasourceentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datasourcesvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
)

type CreateDataSourceInput struct {
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
}
type PatchDataSourceInput struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
type CreateDataConnectionInput struct {
	Name            string          `json:"name"`
	Environment     string          `json:"environment"`
	AdapterType     string          `json:"adapter_type"`
	PublicConfig    json.RawMessage `json:"public_config"`
	CredentialRefID string          `json:"credential_ref_id"`
}
type PatchDataConnectionInput struct {
	Name            string          `json:"name"`
	Environment     string          `json:"environment"`
	PublicConfig    json.RawMessage `json:"public_config"`
	CredentialRefID string          `json:"credential_ref_id"`
}
type CreateCredentialInput struct {
	SecretType string          `json:"secret_type"`
	Secret     json.RawMessage `json:"secret"`
}
type RotateCredentialInput struct {
	Secret json.RawMessage `json:"secret"`
}

type DataSourceDTO struct {
	SourceID   string `json:"source_id"`
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
	Status     string `json:"status"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
type DataConnectionDTO struct {
	ConnectionID     string          `json:"connection_id"`
	SourceID         string          `json:"source_id"`
	Name             string          `json:"name"`
	Environment      string          `json:"environment"`
	AdapterType      string          `json:"adapter_type"`
	PublicConfig     json.RawMessage `json:"public_config"`
	CredentialRefID  string          `json:"credential_ref_id,omitempty"`
	Status           string          `json:"status"`
	LastTestStatus   string          `json:"last_test_status,omitempty"`
	LastTestAt       string          `json:"last_test_at,omitempty"`
	LastTestErrorKey string          `json:"last_test_error_key,omitempty"`
	CreatedBy        string          `json:"created_by"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}
type CredentialRefDTO struct {
	CredentialRefID string `json:"credential_ref_id"`
	Status          string `json:"status"`
	Provider        string `json:"provider"`
	CreatedAt       string `json:"created_at"`
	RotatedAt       string `json:"rotated_at,omitempty"`
}
type DataAssetDTO struct {
	AssetID         string          `json:"asset_id"`
	SourceID        string          `json:"source_id"`
	ConnectionID    string          `json:"connection_id"`
	AssetType       string          `json:"asset_type"`
	Name            string          `json:"name"`
	PhysicalLocator json.RawMessage `json:"physical_locator"`
	LocatorDigest   string          `json:"locator_digest"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}
type SchemaSnapshotDTO struct {
	SnapshotID   string          `json:"snapshot_id"`
	SourceID     string          `json:"source_id"`
	ConnectionID string          `json:"connection_id"`
	AssetID      string          `json:"asset_id"`
	Schema       json.RawMessage `json:"schema"`
	Fingerprint  string          `json:"fingerprint"`
	CreatedBy    string          `json:"created_by"`
	CreatedAt    string          `json:"created_at"`
}

func (s *ApplicationService) requireDatasourceTenant(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, ok := tenantctx.FromContext(ctx)
	if !ok || tc == nil || tc.TenantID == "" {
		return nil, formaerrors.TenantRequired("tenant context required")
	}
	if _, e := s.requireMemberOf(ctx, tc.TenantID); e != nil {
		return nil, e
	}
	if s.DatasourceSVC == nil {
		return nil, formaerrors.Internal("datasource service not initialized")
	}
	return tc, nil
}
func (s *ApplicationService) requireDatasourceAdmin(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	if !roleAtLeastAdmin(tc.MembershipRole) {
		return nil, formaerrors.DataForbidden("data source administration requires OWNER or ADMIN")
	}
	return tc, nil
}
func raw(v json.RawMessage) string {
	if len(v) == 0 {
		return "{}"
	}
	return string(v)
}

func (s *ApplicationService) CreateDataSource(ctx context.Context, in *CreateDataSourceInput) (*DataSourceDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.CreateSource(ctx, &datasourcesvc.CreateSourceInput{TenantID: tc.TenantID, Name: in.Name, ActorID: tc.PrincipalID, SourceType: datasourceentity.SourceType(in.SourceType)})
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataSourceDTO(v), nil
}
func (s *ApplicationService) GetDataSource(ctx context.Context, id string) (*DataSourceDTO, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.GetSource(ctx, tc.TenantID, id)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataSourceDTO(v), nil
}
func (s *ApplicationService) ListDataSources(ctx context.Context) ([]*DataSourceDTO, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	vs, e := s.DatasourceSVC.ListSources(ctx, tc.TenantID)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	out := make([]*DataSourceDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, dataSourceDTO(v))
	}
	return out, nil
}
func (s *ApplicationService) PatchDataSource(ctx context.Context, id string, in *PatchDataSourceInput) (*DataSourceDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.PatchSource(ctx, &datasourcesvc.PatchSourceInput{TenantID: tc.TenantID, SourceID: id, Name: in.Name, Status: datasourceentity.DataSourceStatus(in.Status)})
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataSourceDTO(v), nil
}
func (s *ApplicationService) ArchiveDataSource(ctx context.Context, id string) (*DataSourceDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.ArchiveSource(ctx, tc.TenantID, id)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataSourceDTO(v), nil
}
func (s *ApplicationService) CreateDataConnection(ctx context.Context, sourceID string, in *CreateDataConnectionInput) (*DataConnectionDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.CreateConnection(ctx, &datasourcesvc.CreateConnectionInput{TenantID: tc.TenantID, SourceID: sourceID, Name: in.Name, PublicConfigJSON: raw(in.PublicConfig), CredentialRefID: in.CredentialRefID, ActorID: tc.PrincipalID, Environment: datasourceentity.Environment(in.Environment), AdapterType: datasourceentity.AdapterType(in.AdapterType)})
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataConnectionDTO(v), nil
}
func (s *ApplicationService) GetDataConnection(ctx context.Context, sourceID, id string) (*DataConnectionDTO, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.GetConnection(ctx, tc.TenantID, sourceID, id)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataConnectionDTO(v), nil
}
func (s *ApplicationService) ListDataConnections(ctx context.Context, sourceID string) ([]*DataConnectionDTO, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	vs, e := s.DatasourceSVC.ListConnections(ctx, tc.TenantID, sourceID)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	out := make([]*DataConnectionDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, dataConnectionDTO(v))
	}
	return out, nil
}
func (s *ApplicationService) PatchDataConnection(ctx context.Context, sourceID, id string, in *PatchDataConnectionInput) (*DataConnectionDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.PatchConnection(ctx, &datasourcesvc.PatchConnectionInput{TenantID: tc.TenantID, SourceID: sourceID, ConnectionID: id, Name: in.Name, PublicConfigJSON: func() string {
		if len(in.PublicConfig) == 0 {
			return ""
		}
		return string(in.PublicConfig)
	}(), CredentialRefID: in.CredentialRefID, Environment: datasourceentity.Environment(in.Environment)})
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataConnectionDTO(v), nil
}
func (s *ApplicationService) CreateDataCredential(ctx context.Context, in *CreateCredentialInput) (*CredentialRefDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	if in == nil {
		return nil, formaerrors.DataSecretInvalid("request required")
	}
	v, e := s.DatasourceSVC.CreateCredential(ctx, &datasourcesvc.CreateCredentialInput{TenantID: tc.TenantID, SecretType: in.SecretType, ActorID: tc.PrincipalID, Secret: append([]byte(nil), in.Secret...)})
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return credentialDTO(v), nil
}
func (s *ApplicationService) RotateDataCredential(ctx context.Context, id string, in *RotateCredentialInput) (*CredentialRefDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	if in == nil {
		return nil, formaerrors.DataSecretInvalid("request required")
	}
	v, e := s.DatasourceSVC.RotateCredential(ctx, tc.TenantID, id, append([]byte(nil), in.Secret...))
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return credentialDTO(v), nil
}
func (s *ApplicationService) RevokeDataCredential(ctx context.Context, id string) (*CredentialRefDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.RevokeCredential(ctx, tc.TenantID, id)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return credentialDTO(v), nil
}
func (s *ApplicationService) TestDataConnection(ctx context.Context, sourceID, id string) error {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return e
	}
	return formaerrors.MapDomainError(s.DatasourceSVC.TestConnection(ctx, tc.TenantID, sourceID, id))
}
func (s *ApplicationService) DiscoverDataAssets(ctx context.Context, sourceID, id string) ([]*DataAssetDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	vs, e := s.DatasourceSVC.Discover(ctx, tc.TenantID, sourceID, id)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataAssetDTOs(vs), nil
}
func (s *ApplicationService) ListDataAssets(ctx context.Context, sourceID string) ([]*DataAssetDTO, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	vs, e := s.DatasourceSVC.ListAssets(ctx, tc.TenantID, sourceID)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataAssetDTOs(vs), nil
}
func (s *ApplicationService) GetDataAsset(ctx context.Context, id string) (*DataAssetDTO, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.GetAsset(ctx, tc.TenantID, id)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return dataAssetDTO(v), nil
}
func (s *ApplicationService) CaptureDataSchema(ctx context.Context, sourceID, connectionID, assetID string) (*SchemaSnapshotDTO, error) {
	tc, e := s.requireDatasourceAdmin(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.CaptureSchema(ctx, tc.TenantID, sourceID, connectionID, assetID, tc.PrincipalID)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return schemaSnapshotDTO(v), nil
}
func (s *ApplicationService) GetDataSchemaSnapshot(ctx context.Context, id string) (*SchemaSnapshotDTO, error) {
	tc, e := s.requireDatasourceTenant(ctx)
	if e != nil {
		return nil, e
	}
	v, e := s.DatasourceSVC.GetSnapshot(ctx, tc.TenantID, id)
	if e != nil {
		return nil, formaerrors.MapDomainError(e)
	}
	return schemaSnapshotDTO(v), nil
}

func dataSourceDTO(v *datasourceentity.DataSource) *DataSourceDTO {
	return &DataSourceDTO{SourceID: v.SourceID, Name: v.Name, SourceType: string(v.SourceType), Status: string(v.Status), CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func dataConnectionDTO(v *datasourceentity.DataConnection) *DataConnectionDTO {
	return &DataConnectionDTO{ConnectionID: v.ConnectionID, SourceID: v.SourceID, Name: v.Name, Environment: string(v.Environment), AdapterType: string(v.AdapterType), PublicConfig: json.RawMessage(v.PublicConfigJSON), CredentialRefID: v.CredentialRefID, Status: string(v.Status), LastTestStatus: string(v.LastTestStatus), LastTestAt: formatOptionalTime(v.LastTestAt), LastTestErrorKey: v.LastTestErrorKey, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func credentialDTO(v *datasourceentity.CredentialRef) *CredentialRefDTO {
	return &CredentialRefDTO{CredentialRefID: v.CredentialRefID, Status: string(v.Status), Provider: string(v.Provider), CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), RotatedAt: formatOptionalTime(v.RotatedAt)}
}
func dataAssetDTO(v *datasourceentity.DataAsset) *DataAssetDTO {
	return &DataAssetDTO{AssetID: v.AssetID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetType: string(v.AssetType), Name: v.Name, PhysicalLocator: json.RawMessage(v.PhysicalLocatorJSON), LocatorDigest: v.LocatorDigest, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func dataAssetDTOs(vs []*datasourceentity.DataAsset) []*DataAssetDTO {
	out := make([]*DataAssetDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, dataAssetDTO(v))
	}
	return out
}
func schemaSnapshotDTO(v *datasourceentity.SchemaSnapshot) *SchemaSnapshotDTO {
	return &SchemaSnapshotDTO{SnapshotID: v.SnapshotID, SourceID: v.SourceID, ConnectionID: v.ConnectionID, AssetID: v.AssetID, Schema: json.RawMessage(v.SchemaJSON), Fingerprint: v.Fingerprint, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano)}
}
