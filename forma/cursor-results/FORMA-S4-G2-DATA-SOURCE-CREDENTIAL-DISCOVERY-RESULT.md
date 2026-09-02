# FORMA-S4-G2 DATA SOURCE / CREDENTIAL / DISCOVERY RESULT

## Status

**S4-G2 = PASS**

## Baseline

| Item | Value |
|------|-------|
| Parent tip | `d2d6791645eb38a377f75a3d9e9956c68b76dc96` |
| S4-G1 / G1-F1 | PASS |
| S3 frozen | `forma-s3-frozen` |

## G1 Small Closeout

| Item | Result |
|------|--------|
| `DataAnalysisRunDTO.last_retry_by` | PASS |
| `DataAnalysisRunDTO.last_retry_at` | PASS |
| Internal lease fields hidden from API | PASS |
| G1 domain semantics unchanged | PASS |

## Scope Delivered

| Component | Status |
|-----------|--------|
| DataSource (tenant-level) | PASS |
| DataConnection (multi-env) | PASS |
| CredentialRef (no plaintext) | PASS |
| LocalEncryptedSecretProvider (AES-256-GCM) | PASS |
| MySQL Adapter | PASS |
| PostgreSQL Adapter | PASS |
| HTTP Adapter | PASS |
| DataAsset + discovery idempotency | PASS |
| SchemaSnapshot (immutable) | PASS |
| Stable schema fingerprint | PASS |
| Connection test / preview | PASS |
| Secret write-only API | PASS |
| Tenant isolation | PASS |
| Admin authorization (OWNER/ADMIN mutate) | PASS |
| Backend API | PASS |
| Migration | PASS |

## Out of Scope (not implemented)

SemanticMapping, Mapping DSL, DataContract, ContractRevision, Drift, Capability, Agent, Workflow, Data Plane UI, `forma-s4-frozen`.

## Package Layout

```
coze-studio/backend/domain/forma/datasource/
  entity/
  service/       (adapters, secret provider, orchestration)
  repository/
  internal/dal/
```

## Migration

| File | Tables |
|------|--------|
| `20250902110000_s4_g2_data_source_discovery.sql` | `forma_data_source`, `forma_data_connection`, `forma_data_credential_ref`, `forma_data_secret_local`, `forma_data_asset`, `forma_data_schema_snapshot` |

## API Routes (tenant-scoped)

| Method | Path |
|--------|------|
| GET/POST | `/api/forma/v1/data-sources` |
| GET/PATCH/POST archive | `/api/forma/v1/data-sources/:sourceId` |
| GET/POST | `/api/forma/v1/data-sources/:sourceId/connections` |
| GET/PATCH | `/api/forma/v1/data-sources/:sourceId/connections/:connectionId` |
| POST | `.../connections/:connectionId/test` |
| POST | `.../connections/:connectionId/discover` |
| POST | `/api/forma/v1/credentials` |
| POST | `/api/forma/v1/credentials/:credentialRefId/rotate` |
| POST | `/api/forma/v1/credentials/:credentialRefId/revoke` |
| GET | `/api/forma/v1/data-sources/:sourceId/assets` |
| GET | `/api/forma/v1/data-assets/:assetId` |
| POST | `.../assets/:assetId/capture-schema` |
| GET | `/api/forma/v1/schema-snapshots/:snapshotId` |

## Security Invariants

| Invariant | Result |
|-----------|--------|
| SECRET_ISOLATION | PASS |
| AES-GCM fail-closed (missing/invalid master key) | PASS |
| public_config rejects secret keys | PASS |
| Credential API write-only (no secret in response) | PASS |
| ResolveSecret adapter-only | PASS |
| Error sanitization (no secret in errors) | PASS |
| REAL_MODEL_CALLS | 0 |

## Adapter Coverage

| Adapter | Result |
|---------|--------|
| MYSQL_ADAPTER | PASS |
| POSTGRES_ADAPTER | PASS |
| HTTP_ADAPTER | PASS |
| SCHEMA_SNAPSHOT | PASS |
| TENANT_ISOLATION | PASS |

## Test Matrix

Domain/service tests cover: source/connection separation, multi-env, credential create/rotate, AES-GCM, tenant credential isolation, discovery idempotency, snapshot immutability, fingerprint stability, HTTP unsafe method block, preview row limits, admin vs member auth, no G1 requirement mutation, domain agnostic scan.

## Delivery

| Item | Value |
|------|-------|
| Commit SHA | `33e4bb8f4fc2dde49862f98e1603733bbcf1e639` |
| Forma CI | [33629604516](https://github.com/hai12138/forma/actions/runs/33629604516) **ALL GREEN** |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |
| S4-G2 final status | **PASS** |

## STOP

Do **not** start S4-G3. Do **not** implement SemanticMapping or DataContract. Do **not** create `forma-s4-frozen`.
