/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
)

type AdapterRequest struct {
	PublicConfigJSON string
	Secret           []byte
}

type DiscoveredAsset struct {
	AssetType       entity.AssetType
	Name            string
	PhysicalLocator map[string]any
}

type PreviewRequest struct {
	AdapterRequest
	PhysicalLocator map[string]any
	Limit           int
	Method          string
}

type DataSourceAdapter interface {
	TestConnection(context.Context, *AdapterRequest) error
	DiscoverAssets(context.Context, *AdapterRequest) ([]DiscoveredAsset, error)
	GetSchema(context.Context, *AdapterRequest, map[string]any) (*entity.PhysicalSchema, error)
	Preview(context.Context, *PreviewRequest) ([]map[string]any, error)
	ValidateContract(context.Context, *AdapterRequest, map[string]any) error
}

type AdapterRegistry struct {
	adapters map[entity.AdapterType]DataSourceAdapter
}

func NewAdapterRegistry(adapters map[entity.AdapterType]DataSourceAdapter) *AdapterRegistry {
	copyMap := make(map[entity.AdapterType]DataSourceAdapter, len(adapters))
	for k, v := range adapters {
		copyMap[k] = v
	}
	return &AdapterRegistry{adapters: copyMap}
}

func NewDefaultAdapterRegistry() *AdapterRegistry {
	return NewAdapterRegistry(map[entity.AdapterType]DataSourceAdapter{
		entity.AdapterMySQL:      NewMySQLAdapter(nil),
		entity.AdapterPostgreSQL: NewPostgreSQLAdapter(nil),
		entity.AdapterHTTP:       NewHTTPAdapter(nil),
	})
}

func (r *AdapterRegistry) Get(t entity.AdapterType) (DataSourceAdapter, error) {
	if r == nil {
		return nil, entity.ErrDataAdapterNotSupported
	}
	a, ok := r.adapters[t]
	if !ok {
		return nil, entity.ErrDataAdapterNotSupported
	}
	return a, nil
}

type DBOpener func(driverName, dsn string) (*sql.DB, error)

type sqlAdapter struct {
	driver string
	opener DBOpener
}

type sqlPublicConfig struct {
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	Database   string            `json:"database"`
	Username   string            `json:"username"`
	Parameters map[string]string `json:"parameters"`
}

type sqlSecret struct {
	Password string `json:"password"`
}

func NewMySQLAdapter(opener DBOpener) DataSourceAdapter {
	if opener == nil {
		opener = sql.Open
	}
	return &sqlAdapter{driver: "mysql", opener: opener}
}
func NewPostgreSQLAdapter(opener DBOpener) DataSourceAdapter {
	if opener == nil {
		opener = sql.Open
	}
	return &sqlAdapter{driver: "pgx", opener: opener}
}

func (a *sqlAdapter) open(ctx context.Context, req *AdapterRequest) (*sql.DB, error) {
	var cfg sqlPublicConfig
	var secret sqlSecret
	if req == nil || json.Unmarshal([]byte(req.PublicConfigJSON), &cfg) != nil ||
		json.Unmarshal(req.Secret, &secret) != nil || cfg.Host == "" || cfg.Database == "" || cfg.Username == "" {
		return nil, entity.ErrPublicConfigInvalid
	}
	var dsn string
	if a.driver == "mysql" {
		if cfg.Port == 0 {
			cfg.Port = 3306
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", cfg.Username, secret.Password, cfg.Host, cfg.Port, cfg.Database)
	} else {
		if cfg.Port == 0 {
			cfg.Port = 5432
		}
		q := url.Values{"sslmode": []string{"disable"}}
		for k, v := range cfg.Parameters {
			q.Set(k, v)
		}
		dsn = (&url.URL{Scheme: "postgres", User: url.UserPassword(cfg.Username, secret.Password), Host: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), Path: cfg.Database, RawQuery: q.Encode()}).String()
	}
	db, err := a.opener(a.driver, dsn)
	if err != nil {
		return nil, entity.ErrDataConnectionFailed
	}
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 10*time.Second {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, entity.ErrDataConnectionFailed
	}
	return db, nil
}

func (a *sqlAdapter) TestConnection(ctx context.Context, req *AdapterRequest) error {
	db, err := a.open(ctx, req)
	if err != nil {
		return err
	}
	defer db.Close()
	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil || one != 1 {
		return entity.ErrDataConnectionFailed
	}
	return nil
}

func (a *sqlAdapter) DiscoverAssets(ctx context.Context, req *AdapterRequest) ([]DiscoveredAsset, error) {
	db, err := a.open(ctx, req)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	query := `SELECT table_schema, table_name, CASE WHEN table_type = 'VIEW' THEN 'VIEW' ELSE 'TABLE' END AS object_type FROM information_schema.tables WHERE table_type IN ('BASE TABLE','VIEW') AND table_schema NOT IN ('information_schema','mysql','performance_schema','sys','pg_catalog') ORDER BY table_schema, table_name`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	defer rows.Close()
	out := []DiscoveredAsset{}
	for rows.Next() {
		var schema, table, objectType string
		if rows.Scan(&schema, &table, &objectType) != nil {
			return nil, entity.ErrDataDiscoveryFailed
		}
		out = append(out, DiscoveredAsset{AssetType: entity.AssetDataset, Name: table, PhysicalLocator: map[string]any{"schema": schema, "table": table, "object_type": objectType}})
	}
	if rows.Err() != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	return out, nil
}

var safeIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$]*$`)

func quoteIdentifier(driver, value string) (string, error) {
	if !safeIdentifier.MatchString(value) {
		return "", entity.ErrPublicConfigInvalid
	}
	if driver == "mysql" {
		return "`" + value + "`", nil
	}
	return `"` + value + `"`, nil
}

func (a *sqlAdapter) GetSchema(ctx context.Context, req *AdapterRequest, locator map[string]any) (*entity.PhysicalSchema, error) {
	schema, _ := locator["schema"].(string)
	table, _ := locator["table"].(string)
	if _, err := quoteIdentifier(a.driver, schema); err != nil {
		return nil, err
	}
	if _, err := quoteIdentifier(a.driver, table); err != nil {
		return nil, err
	}
	db, err := a.open(ctx, req)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	query := `SELECT c.column_name, c.data_type, c.column_type, c.is_nullable, c.ordinal_position,
		c.column_comment,
		CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN 1 ELSE 0 END
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu ON c.table_schema=kcu.table_schema AND c.table_name=kcu.table_name AND c.column_name=kcu.column_name
		LEFT JOIN information_schema.table_constraints tc ON kcu.constraint_name=tc.constraint_name AND kcu.table_schema=tc.table_schema AND kcu.table_name=tc.table_name
		WHERE c.table_schema=? AND c.table_name=? ORDER BY c.ordinal_position`
	if a.driver != "mysql" {
		query = `SELECT c.column_name, c.data_type, c.udt_name, c.is_nullable, c.ordinal_position,
			COALESCE(pg_catalog.col_description(pg_catalog.to_regclass(c.table_schema || '.' || c.table_name), c.ordinal_position), ''),
			CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN 1 ELSE 0 END
			FROM information_schema.columns c
			LEFT JOIN information_schema.key_column_usage kcu ON c.table_schema=kcu.table_schema AND c.table_name=kcu.table_name AND c.column_name=kcu.column_name
			LEFT JOIN information_schema.table_constraints tc ON kcu.constraint_name=tc.constraint_name AND kcu.table_schema=tc.table_schema AND kcu.table_name=tc.table_name
			WHERE c.table_schema=? AND c.table_name=? ORDER BY c.ordinal_position`
		query = strings.Replace(query, "?", "$1", 1)
		query = strings.Replace(query, "?", "$2", 1)
	}
	rows, err := db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	defer rows.Close()
	out := &entity.PhysicalSchema{Name: schema + "." + table, Fields: []entity.PhysicalField{}, Relationships: []entity.PhysicalRelationship{}}
	seenPaths := map[string]int{}
	for rows.Next() {
		var f entity.PhysicalField
		var nullable string
		var pk int
		if rows.Scan(&f.Name, &f.DataType, &f.NativeType, &nullable, &f.Ordinal, &f.Description, &pk) != nil {
			return nil, entity.ErrDataDiscoveryFailed
		}
		f.Nullable = strings.EqualFold(nullable, "YES")
		f.PrimaryKey = pk == 1
		f.Path = schema + "." + table + "." + f.Name
		if idx, ok := seenPaths[f.Path]; ok {
			if f.PrimaryKey {
				out.Fields[idx].PrimaryKey = true
			}
			continue
		}
		seenPaths[f.Path] = len(out.Fields)
		out.Fields = append(out.Fields, f)
	}
	if rows.Err() != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	_ = rows.Close()
	relationshipQuery := `SELECT rc.constraint_name, kcu.column_name, kcu.referenced_table_schema,
		kcu.referenced_table_name, kcu.referenced_column_name
		FROM information_schema.referential_constraints rc
		JOIN information_schema.key_column_usage kcu
		  ON rc.constraint_schema=kcu.constraint_schema AND rc.constraint_name=kcu.constraint_name
		WHERE rc.constraint_schema=? AND rc.table_name=?
		ORDER BY rc.constraint_name, kcu.ordinal_position`
	if a.driver != "mysql" {
		relationshipQuery = `SELECT tc.constraint_name, kcu.column_name, ccu.table_schema,
			ccu.table_name, ccu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
			  ON tc.constraint_name=kcu.constraint_name AND tc.table_schema=kcu.table_schema
			JOIN information_schema.constraint_column_usage ccu
			  ON tc.constraint_name=ccu.constraint_name AND tc.table_schema=ccu.table_schema
			WHERE tc.constraint_type='FOREIGN KEY' AND tc.table_schema=$1 AND tc.table_name=$2
			ORDER BY tc.constraint_name, kcu.ordinal_position`
	}
	relationshipRows, err := db.QueryContext(ctx, relationshipQuery, schema, table)
	if err != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	defer relationshipRows.Close()
	byName := map[string]int{}
	for relationshipRows.Next() {
		var name, from, targetSchema, targetTable, targetField string
		if relationshipRows.Scan(&name, &from, &targetSchema, &targetTable, &targetField) != nil {
			return nil, entity.ErrDataDiscoveryFailed
		}
		if index, exists := byName[name]; exists {
			out.Relationships[index].FromFields = append(out.Relationships[index].FromFields, from)
			out.Relationships[index].ToFields = append(out.Relationships[index].ToFields, targetField)
			continue
		}
		byName[name] = len(out.Relationships)
		out.Relationships = append(out.Relationships, entity.PhysicalRelationship{
			Name: name, FromFields: []string{from}, ToSchema: targetSchema + "." + targetTable,
			ToFields: []string{targetField}, RelationshipType: "MANY_TO_ONE",
		})
	}
	if relationshipRows.Err() != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	return out, nil
}

func (a *sqlAdapter) Preview(ctx context.Context, req *PreviewRequest) ([]map[string]any, error) {
	if req == nil {
		return nil, entity.ErrPublicConfigInvalid
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	schema, _ := req.PhysicalLocator["schema"].(string)
	table, _ := req.PhysicalLocator["table"].(string)
	qs, e := quoteIdentifier(a.driver, schema)
	if e != nil {
		return nil, e
	}
	qt, e := quoteIdentifier(a.driver, table)
	if e != nil {
		return nil, e
	}
	db, err := a.open(ctx, &req.AdapterRequest)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.QueryContext(ctx, "SELECT * FROM "+qs+"."+qt+" LIMIT "+strconv.Itoa(limit))
	if err != nil {
		return nil, entity.ErrDataConnectionFailed
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	out := []map[string]any{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if rows.Scan(ptrs...) != nil {
			return nil, entity.ErrDataConnectionFailed
		}
		m := map[string]any{}
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				m[c] = string(b)
			} else {
				m[c] = vals[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func (a *sqlAdapter) ValidateContract(context.Context, *AdapterRequest, map[string]any) error {
	return entity.ErrDataContractNotAvailable
}

type HTTPAdapter struct {
	client   *http.Client
	maxBytes int64
	policy   OutboundNetworkPolicy
}

func NewHTTPAdapter(client *http.Client) DataSourceAdapter {
	return NewHTTPAdapterWithPolicy(client, nil)
}

func NewHTTPAdapterWithPolicy(client *http.Client, policy OutboundNetworkPolicy) DataSourceAdapter {
	if policy == nil {
		policy = NewDefaultOutboundNetworkPolicy(nil)
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	cloned := *client
	cloned.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("redirect limit")
		}
		if err := policy.ValidateURL(next.Context(), next.URL); err != nil {
			return err
		}
		return nil
	}
	return &HTTPAdapter{client: &cloned, maxBytes: 2 << 20, policy: policy}
}

type httpConfig struct {
	BaseURL    string            `json:"base_url"`
	OpenAPIURL string            `json:"openapi_url"`
	Headers    map[string]string `json:"headers"`
}

func parseHTTPConfig(raw string) (*httpConfig, error) {
	var c httpConfig
	if json.Unmarshal([]byte(raw), &c) != nil {
		return nil, entity.ErrPublicConfigInvalid
	}
	u, e := url.Parse(c.BaseURL)
	if e != nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, entity.ErrPublicConfigInvalid
	}
	return &c, nil
}
func (a *HTTPAdapter) request(ctx context.Context, method, target string, headers map[string]string) ([]byte, error) {
	u, e := url.Parse(target)
	if e != nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, entity.ErrPublicConfigInvalid
	}
	if e = a.policy.ValidateURL(ctx, u); e != nil {
		return nil, e
	}
	r, e := http.NewRequestWithContext(ctx, method, target, nil)
	if e != nil {
		return nil, entity.ErrPublicConfigInvalid
	}
	for name, value := range headers {
		r.Header.Set(name, value)
	}
	resp, e := a.client.Do(r)
	if e != nil {
		return nil, entity.ErrDataConnectionFailed
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, entity.ErrDataConnectionFailed
	}
	b, e := io.ReadAll(io.LimitReader(resp.Body, a.maxBytes+1))
	if e != nil || int64(len(b)) > a.maxBytes {
		return nil, entity.ErrDataConnectionFailed
	}
	return b, nil
}
func httpRequestHeaders(req *AdapterRequest, config *httpConfig) map[string]string {
	headers := make(map[string]string, len(config.Headers))
	for k, v := range config.Headers {
		headers[k] = v
	}
	if req != nil && len(req.Secret) > 0 {
		var secret struct {
			Headers map[string]string `json:"headers"`
		}
		if json.Unmarshal(req.Secret, &secret) == nil {
			for k, v := range secret.Headers {
				headers[k] = v
			}
		}
	}
	return headers
}
func (a *HTTPAdapter) TestConnection(ctx context.Context, req *AdapterRequest) error {
	if req == nil {
		return entity.ErrPublicConfigInvalid
	}
	c, e := parseHTTPConfig(req.PublicConfigJSON)
	if e != nil {
		return e
	}
	headers := httpRequestHeaders(req, c)
	u, _ := url.Parse(c.BaseURL)
	if e = a.policy.ValidateURL(ctx, u); e != nil {
		return e
	}
	request, _ := http.NewRequestWithContext(ctx, http.MethodHead, c.BaseURL, nil)
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, e := a.client.Do(request)
	if e != nil {
		return entity.ErrDataConnectionFailed
	}
	_ = response.Body.Close()
	if response.StatusCode == http.StatusMethodNotAllowed {
		_, e = a.request(ctx, http.MethodGet, c.BaseURL, headers)
		return e
	}
	if response.StatusCode >= 400 {
		return entity.ErrDataConnectionFailed
	}
	return nil
}
func (a *HTTPAdapter) DiscoverAssets(ctx context.Context, req *AdapterRequest) ([]DiscoveredAsset, error) {
	if req == nil {
		return nil, entity.ErrPublicConfigInvalid
	}
	c, e := parseHTTPConfig(req.PublicConfigJSON)
	if e != nil {
		return nil, e
	}
	if c.OpenAPIURL == "" {
		return []DiscoveredAsset{{AssetType: entity.AssetEndpoint, Name: c.BaseURL, PhysicalLocator: map[string]any{"url": c.BaseURL, "method": "GET"}}}, nil
	}
	b, e := a.request(ctx, http.MethodGet, c.OpenAPIURL, httpRequestHeaders(req, c))
	if e != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if json.Unmarshal(b, &doc) != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	out := []DiscoveredAsset{}
	for p, methods := range doc.Paths {
		for m := range methods {
			m = strings.ToUpper(m)
			if m == "GET" {
				out = append(out, DiscoveredAsset{AssetType: entity.AssetEndpoint, Name: m + " " + p, PhysicalLocator: map[string]any{"url": strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(p, "/"), "method": m, "path": p}})
			}
		}
	}
	return out, nil
}
func (a *HTTPAdapter) GetSchema(ctx context.Context, req *AdapterRequest, locator map[string]any) (*entity.PhysicalSchema, error) {
	if req == nil {
		return nil, entity.ErrPublicConfigInvalid
	}
	config, err := parseHTTPConfig(req.PublicConfigJSON)
	if err != nil || config.OpenAPIURL == "" {
		return nil, entity.ErrDataDiscoveryFailed
	}
	document, err := a.request(ctx, http.MethodGet, config.OpenAPIURL, httpRequestHeaders(req, config))
	if err != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	path, _ := locator["path"].(string)
	return ParseOpenAPISchema(document, path)
}

func (a *HTTPAdapter) ValidateContract(context.Context, *AdapterRequest, map[string]any) error {
	return entity.ErrDataContractNotAvailable
}
func (a *HTTPAdapter) Preview(ctx context.Context, req *PreviewRequest) ([]map[string]any, error) {
	method := strings.ToUpper(req.Method)
	if method == "" {
		method, _ = req.PhysicalLocator["method"].(string)
	}
	if method == "" {
		method = "GET"
	}
	if method != "GET" && method != "HEAD" {
		return nil, entity.ErrPublicConfigInvalid
	}
	target, _ := req.PhysicalLocator["url"].(string)
	headers := map[string]string{}
	if req.PublicConfigJSON != "" {
		config, configErr := parseHTTPConfig(req.PublicConfigJSON)
		if configErr != nil {
			return nil, configErr
		}
		headers = httpRequestHeaders(&req.AdapterRequest, config)
	}
	b, e := a.request(ctx, method, target, headers)
	if e != nil {
		return nil, e
	}
	var value any
	if len(b) > 0 && json.Unmarshal(b, &value) == nil {
		return []map[string]any{{"body": value}}, nil
	}
	return []map[string]any{{"body": string(b)}}, nil
}
