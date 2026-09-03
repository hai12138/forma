## Status

**S4-G5-F2 = PASS** (awaiting Forma CI ALL GREEN + human review)

## Baseline

| Item | Value |
|------|-------|
| Parent tip (requested) | `bcaeb44d59d50ac4a364fdff8f6e8f0fbdd603f8` |
| S4-G5-F1 | PASS (`a3c14697` / docs tip `bcaeb44d`) |
| Scope | Minimal F2 fixes only — **no S4-G6**, **no `forma-s4-frozen`**, no S4 backend domain-semantic / migration changes |

## Pass Gate

| Gate | Result |
|------|--------|
| PHYSICAL_BINDING_PAYLOAD_ISOLATION | PASS |
| SANITIZED_ERROR_ALLOWLIST | PASS |
| SOURCE_TYPE_ALIGNMENT | PASS |
| CONNECTION_ADAPTER_ALIGNMENT | PASS |
| MAPPING_LINEAGE | PASS |
| CURRENT_BUSINESS_REVISION | PASS |
| ROLE_TESTS | PASS |
| ROUTE_TESTS | PASS (all 8 `/data*` MemoryRouter routes) |
| REAL_TYPECHECK | PASS |
| FRONTEND_BUILD | PASS (local typecheck + package tests; CI pending) |
| REAL_MODEL_CALLS | 0 |
| NO_MIGRATION | PASS |

## Fixes

### 1. Physical binding payload isolation

- Application DTO `DataContractRevisionDTO.binding_refs` uses `omitempty`.
- `ListDataContractRevisions` / `GetDataContractRevision` project bindings only when `roleAtLeastAdmin` (OWNER/ADMIN).
- MEMBER/VIEWER responses omit `binding_refs` and therefore never serialize `source_id` / `connection_id` / `asset_id` / `schema_snapshot_id` from bindings.
- Go test `TestContractRevisionPhysicalBindingPayloadIsolation` asserts JSON for MEMBER/VIEWER has no physical binding keys; OWNER/ADMIN retain them.
- Frontend test asserts MEMBER network revision objects lack physical binding fields (not React-tab-only).

### 2. Error output safety

- `sanitizedErrorMessage` maps only allowlisted `error_key` / `code`; unknown `FormaApiError` and plain `Error` → `操作失败` (never `err.message`).
- Contract validation issues use `validationIssueLabel(code)` allowlist; raw backend `message` is never rendered.
- DOM regression covers DB connection strings, Authorization, password, token in load/validation errors.

### 3. Data Source UI ↔ backend enums

- `source_type` submit: `RELATIONAL_DATABASE` | `HTTP_API` only.
- Relational connections: `MYSQL` / `POSTGRESQL` + `host`/`port`/`database`/`username`.
- HTTP_API connections: `HTTP` adapter only + `base_url` / `openapi_url`.
- HTTP sources never hardcode `MYSQL`.
- Real request payload tests + unit helpers aligned to backend `source.go` / `source_service.go`.

### 4. Mapping lineage

- Removed `src_manual` / `conn_manual` / `asset_manual`.
- Manual mapping loads SchemaSnapshot DTO and submits consistent `source_id` / `connection_id` / `asset_id` / `schema_snapshot_id`.
- Incomplete snapshot → UI error, no create call.
- Mismatch negative + valid snapshot positive tests.

### 5. Contract revision

- Create Contract uses `business.current_revision` (not fixed `1`).
- Request test with `current_revision: 5`.

### 6. F1 test evidence

- Health: OWNER can evaluate drift/gap; MEMBER cannot (separate renders).
- ACTIVE and STALE Deprecate each assert exact `revision_id` and list refresh.
- All eight route behavior tests retained.

## Local Verification

| Check | Result |
|-------|--------|
| `rush test --only @forma/data` | 53 PASS |
| `scripts/forma/typecheck.mjs` | PASS (`api-client`, `data`, `app`) |
| `scripts/forma/routes-smoke.mjs` | PASS |
| `go test ./application/forma/ -run ContractRevisionPhysicalBinding\|ContractApplication` | PASS |
| REAL_MODEL_CALLS | 0 |
| Migrations touched | none |

## Out of scope (explicit)

- S4-G6
- `forma-s4-frozen` branch/tag
- Backend domain semantic changes / new migrations
- Consumer Contract Interface (S5)

## Stop condition

Wait for latest `main` Forma CI jobs **forma-backend**, **forma-migration-apply**, **forma-frontend** ALL GREEN, then **stop for human review**.
