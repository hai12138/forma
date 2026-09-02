/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/datasource/entity"
)

func TestHTTPAdapterPreviewAndDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/openapi.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"paths":{"/customers":{"get":{}},"/unsafe":{"post":{}}}}`))
		default:
			if r.Method == http.MethodHead {
				return
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))
	defer server.Close()
	adapter := NewHTTPAdapter(nil)
	req := &AdapterRequest{PublicConfigJSON: `{"base_url":"` + server.URL + `","openapi_url":"` + server.URL + `/openapi.json"}`}
	require.NoError(t, adapter.TestConnection(context.Background(), req))
	assets, err := adapter.DiscoverAssets(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, assets, 1)
	rows, err := adapter.Preview(context.Background(), &PreviewRequest{PhysicalLocator: assets[0].PhysicalLocator, Method: "GET"})
	require.NoError(t, err)
	require.Equal(t, true, rows[0]["body"].(map[string]any)["ok"])
}

func TestHTTPAdapterRejectsCredentialsUnsafeMethodsAndOversize(t *testing.T) {
	adapter := NewHTTPAdapter(nil)
	err := adapter.TestConnection(context.Background(), &AdapterRequest{PublicConfigJSON: `{"base_url":"https://user:pass@example.com"}`})
	require.ErrorIs(t, err, entity.ErrPublicConfigInvalid)
	_, err = adapter.Preview(context.Background(), &PreviewRequest{PhysicalLocator: map[string]any{"url": "https://example.com"}, Method: "POST"})
	require.ErrorIs(t, err, entity.ErrPublicConfigInvalid)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(strings.Repeat("x", (2<<20)+1))) }))
	defer server.Close()
	_, err = adapter.Preview(context.Background(), &PreviewRequest{PhysicalLocator: map[string]any{"url": server.URL}, Method: "GET"})
	require.ErrorIs(t, err, entity.ErrDataConnectionFailed)
}

func TestAdapterRegistryAndIdentifierSafety(t *testing.T) {
	registry := NewDefaultAdapterRegistry()
	for _, kind := range []entity.AdapterType{entity.AdapterMySQL, entity.AdapterPostgreSQL, entity.AdapterHTTP} {
		_, err := registry.Get(kind)
		require.NoError(t, err)
	}
	_, err := registry.Get("ORACLE")
	require.ErrorIs(t, err, entity.ErrDataAdapterNotSupported)
	_, err = quoteIdentifier("mysql", "orders; DROP TABLE users")
	require.ErrorIs(t, err, entity.ErrPublicConfigInvalid)
	require.Equal(t, `"orders"`, func() string { v, _ := quoteIdentifier("pgx", "orders"); return v }())
}

func TestMySQLAdapterReadOnlyConnectionContract(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	mock.ExpectPing()
	mock.ExpectQuery(`SELECT 1`).WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	opener := func(driver, dsn string) (*sql.DB, error) {
		require.Equal(t, "mysql", driver)
		require.Contains(t, dsn, "@tcp(db.example:3306)/warehouse")
		return db, nil
	}
	adapter := NewMySQLAdapter(opener)
	err = adapter.TestConnection(context.Background(), &AdapterRequest{
		PublicConfigJSON: `{"host":"db.example","port":3306,"database":"warehouse","username":"reader"}`,
		Secret:           []byte(`{"password":"` + testSecretPrefix + `db"}`),
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMySQLPreviewCapsLimitAtOneHundred(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	mock.ExpectPing()
	mock.ExpectQuery("SELECT \\* FROM `public`.`orders` LIMIT 100").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	adapter := NewMySQLAdapter(func(string, string) (*sql.DB, error) { return db, nil })
	rows, err := adapter.Preview(context.Background(), &PreviewRequest{
		AdapterRequest: AdapterRequest{
			PublicConfigJSON: `{"host":"db","database":"warehouse","username":"reader"}`,
			Secret:           []byte(`{"password":"x"}`),
		},
		PhysicalLocator: map[string]any{"schema": "public", "table": "orders"},
		Limit:           1000,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLSchemaIsNormalized(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	mock.ExpectPing()
	mock.ExpectQuery("FROM information_schema.columns").
		WithArgs("public", "customers").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "ordinal_position", "is_primary"}).
			AddRow("id", "bigint", "NO", 1, 1).
			AddRow("account_id", "bigint", "YES", 2, 0))
	mock.ExpectQuery("FROM information_schema.table_constraints").
		WithArgs("public", "customers").
		WillReturnRows(sqlmock.NewRows([]string{"constraint_name", "column_name", "table_schema", "table_name", "column_name"}).
			AddRow("fk_customer_account", "account_id", "public", "accounts", "id"))
	adapter := NewPostgreSQLAdapter(func(driver string, _ string) (*sql.DB, error) {
		require.Equal(t, "pgx", driver)
		return db, nil
	})
	schema, err := adapter.GetSchema(context.Background(), &AdapterRequest{
		PublicConfigJSON: `{"host":"db","database":"warehouse","username":"reader"}`,
		Secret:           []byte(`{"password":"x"}`),
	}, map[string]any{"schema": "public", "table": "customers"})
	require.NoError(t, err)
	require.Equal(t, "public.customers", schema.Name)
	require.True(t, schema.Fields[0].PrimaryKey)
	require.Equal(t, "public.accounts", schema.Relationships[0].ToSchema)
	require.Equal(t, []string{"account_id"}, schema.Relationships[0].FromFields)
	require.NoError(t, mock.ExpectationsWereMet())
}
