/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/datasource/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/datasource/repository"
)

const testSecretPrefix = "FORMA_TEST_SUPER_SECRET_"

// SetTestMasterKeyForTests is test-only by construction: this file is not
// included in production builds.
func SetTestMasterKeyForTests(t *testing.T) {
	t.Helper()
	t.Setenv("FORMA_SECRET_MASTER_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
}

type fakeAdapter struct {
	assets  []DiscoveredAsset
	schema  *entity.PhysicalSchema
	testErr error
}

func (f *fakeAdapter) TestConnection(context.Context, *AdapterRequest) error { return f.testErr }
func (f *fakeAdapter) DiscoverAssets(context.Context, *AdapterRequest) ([]DiscoveredAsset, error) {
	return f.assets, nil
}
func (f *fakeAdapter) GetSchema(context.Context, *AdapterRequest, map[string]any) (*entity.PhysicalSchema, error) {
	return f.schema, nil
}
func (f *fakeAdapter) Preview(context.Context, *PreviewRequest) ([]map[string]any, error) {
	return nil, nil
}

func testService(t *testing.T, adapter DataSourceAdapter) DataSourceService {
	t.Helper()
	SetTestMasterKeyForTests(t)
	secrets, err := NewLocalSecretProviderFromEnv()
	require.NoError(t, err)
	return NewDataSourceService(&Components{
		Repo:     repository.NewMemoryDataSourceRepository(),
		Secrets:  secrets,
		Adapters: NewAdapterRegistry(map[entity.AdapterType]DataSourceAdapter{entity.AdapterMySQL: adapter}),
	})
}

func TestLocalSecretProviderRoundTripAndAADIsolation(t *testing.T) {
	SetTestMasterKeyForTests(t)
	p, err := NewLocalSecretProviderFromEnv()
	require.NoError(t, err)
	secret := []byte(testSecretPrefix + "password")
	encrypted, err := p.StoreSecret(context.Background(), "tenant-a", "credential-a", 1, secret)
	require.NoError(t, err)
	require.NotContains(t, string(encrypted), string(secret))
	plain, err := p.ResolveSecret(context.Background(), "tenant-a", "credential-a", 1, encrypted)
	require.NoError(t, err)
	require.Equal(t, secret, plain)
	_, err = p.ResolveSecret(context.Background(), "tenant-b", "credential-a", 1, encrypted)
	require.ErrorIs(t, err, entity.ErrSecretInvalid)
}

func TestLocalSecretProviderFailsClosed(t *testing.T) {
	t.Setenv("FORMA_SECRET_MASTER_KEY", "")
	_, err := NewLocalSecretProviderFromEnv()
	require.ErrorIs(t, err, entity.ErrSecretProviderUnavailable)
	t.Setenv("FORMA_SECRET_MASTER_KEY", "not-a-32-byte-key")
	_, err = NewLocalSecretProviderFromEnv()
	require.ErrorIs(t, err, entity.ErrSecretProviderUnavailable)
}

func TestPublicConfigRejectsSecretsRecursively(t *testing.T) {
	for _, raw := range []string{
		`{"password":"x"}`, `{"nested":{"API-Key":"x"}}`, `{"items":[{"authorization":"x"}]}`,
		`{"clientSecret":"x"}`, `{"refresh_token":"x"}`,
	} {
		require.ErrorIs(t, ValidatePublicConfig(raw), entity.ErrPublicConfigInvalid, raw)
	}
	require.NoError(t, ValidatePublicConfig(`{"host":"db","port":3306,"options":{"ssl":true}}`))
}

func TestDiscoveryIsTenantScopedAndIdempotent(t *testing.T) {
	a := &fakeAdapter{assets: []DiscoveredAsset{{AssetType: entity.AssetDataset, Name: "orders", PhysicalLocator: map[string]any{"table": "orders", "schema": "public"}}}}
	s := testService(t, a)
	ctx := context.Background()
	source, err := s.CreateSource(ctx, &CreateSourceInput{TenantID: "tenant-a", Name: "warehouse", SourceType: entity.SourceTypeDatabase, ActorID: "admin"})
	require.NoError(t, err)
	conn, err := s.CreateConnection(ctx, &CreateConnectionInput{TenantID: "tenant-a", SourceID: source.SourceID, Name: "dev", Environment: entity.EnvironmentDev, AdapterType: entity.AdapterMySQL, PublicConfigJSON: `{"host":"db"}`, ActorID: "admin"})
	require.NoError(t, err)
	first, err := s.Discover(ctx, "tenant-a", source.SourceID, conn.ConnectionID)
	require.NoError(t, err)
	second, err := s.Discover(ctx, "tenant-a", source.SourceID, conn.ConnectionID)
	require.NoError(t, err)
	require.Equal(t, first[0].AssetID, second[0].AssetID)
	require.Len(t, second, 1)
	_, err = s.GetAsset(ctx, "tenant-b", first[0].AssetID)
	require.ErrorIs(t, err, entity.ErrDataAssetNotFound)
}

func TestSchemaFingerprintStableAndSnapshotsImmutable(t *testing.T) {
	a := &fakeAdapter{
		assets: []DiscoveredAsset{{AssetType: entity.AssetDataset, Name: "orders", PhysicalLocator: map[string]any{"schema": "public", "table": "orders"}}},
		schema: &entity.PhysicalSchema{Name: "public.orders", Fields: []entity.PhysicalField{{Name: "id", DataType: "bigint", PrimaryKey: true, Ordinal: 1}}, Relationships: []entity.PhysicalRelationship{}},
	}
	s := testService(t, a)
	ctx := context.Background()
	source, _ := s.CreateSource(ctx, &CreateSourceInput{TenantID: "t", Name: "db", SourceType: entity.SourceTypeDatabase})
	conn, _ := s.CreateConnection(ctx, &CreateConnectionInput{TenantID: "t", SourceID: source.SourceID, Name: "dev", Environment: entity.EnvironmentDev, AdapterType: entity.AdapterMySQL, PublicConfigJSON: `{"host":"db"}`})
	assets, err := s.Discover(ctx, "t", source.SourceID, conn.ConnectionID)
	require.NoError(t, err)
	one, err := s.CaptureSchema(ctx, "t", source.SourceID, conn.ConnectionID, assets[0].AssetID, "admin")
	require.NoError(t, err)
	two, err := s.CaptureSchema(ctx, "t", source.SourceID, conn.ConnectionID, assets[0].AssetID, "admin")
	require.NoError(t, err)
	require.NotEqual(t, one.SnapshotID, two.SnapshotID)
	require.Equal(t, one.Fingerprint, two.Fingerprint)
	one.SchemaJSON = "mutated"
	stored, err := s.GetSnapshot(ctx, "t", one.SnapshotID)
	require.NoError(t, err)
	require.NotEqual(t, "mutated", stored.SchemaJSON)
}

func TestCredentialErrorsAndJSONNeverContainSecret(t *testing.T) {
	s := testService(t, &fakeAdapter{})
	secret := []byte(`{"token":"` + testSecretPrefix + `api-token"}`)
	ref, err := s.CreateCredential(context.Background(), &CreateCredentialInput{TenantID: "t", SecretType: "PASSWORD", ActorID: "admin", Secret: append([]byte(nil), secret...)})
	require.NoError(t, err)
	encoded, err := json.Marshal(ref)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), string(secret))
	_, err = s.RotateCredential(context.Background(), "other", ref.CredentialRefID, append([]byte(nil), secret...))
	require.Error(t, err)
	require.False(t, strings.Contains(err.Error(), string(secret)))
	require.True(t, errors.Is(err, entity.ErrDataCredentialNotFound))
}
