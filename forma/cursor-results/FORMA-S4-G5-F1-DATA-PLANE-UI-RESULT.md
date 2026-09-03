## Status

**S4-G5-F1 = PASS**

## Baseline

| Item | Value |
|------|-------|
| Parent tip | `001576525036041b14e6a5f55e31bbd08f506345` |
| S4-G5 | PASS (`7c952201` / CI `33666159055`) |

## Pass Gate

| Gate | Result |
|------|--------|
| CONTRACT_LIFECYCLE_UX | PASS |
| MAPPING_DSL_ALIGNMENT | PASS |
| PHYSICAL_BINDING_ROLE_ISOLATION | PASS |
| ROLE_BEHAVIOR_TESTS | PASS |
| ROUTE_BEHAVIOR_TESTS | PASS |
| CREDENTIAL_WRITE_ONLY | PASS |
| CONTRACT_LOGICAL_PHYSICAL_SEPARATION | PASS |
| DOMAIN_AGNOSTIC | PASS |
| REAL_TYPECHECK | PASS |
| FRONTEND_BUILD | PASS (CI) |
| REAL_MODEL_CALLS | 0 |
| NO_MIGRATION | PASS |
| NO_BUSINESS_MODEL_MUTATION | PASS |

## Fixes

### Contract lifecycle UX

- DRAFT: Validate only
- VALIDATED: Activate only (confirm dialog)
- ACTIVE / STALE: Deprecate only
- DRAFT cannot Deprecate
- Validate / Activate / Deprecate reload revision state
- Mutations use sanitized errors; no unhandled rejection

### Mapping DSL

- `TIMEZONE` → `TIME_NORMALIZE` (`source_timezone`, `target_timezone`, `format`)
- `JSON_PATH` → `FIELD_PATH` (`path`)
- `transform_spec.type` always equals `mapping_type`
- No SQL / JS / Python / arbitrary expression editors

### Role isolation

- Physical binding tab and binding payload: OWNER/ADMIN only
- MEMBER/VIEWER: logical interface + public lifecycle read
- Credential create and all mutations: OWNER/ADMIN only
- Backend permission remains the security boundary

### Tests

- MemoryRouter/Routes render of `/data`, `/data/requirements`, `/data/sources`, `/data/sources/:sourceId`, `/data/mappings`, `/data/contracts`, `/data/contracts/:contractId`, `/data/health`
- Removed string-array / comment-hit fake route tests
- OWNER vs MEMBER (and VIEWER) rendering
- DRAFT→Validate, VALIDATED→Activate, ACTIVE/STALE→Deprecate
- TIME_NORMALIZE and FIELD_PATH request payloads
- Credential secret `FORMA_G5_TEST_SUPER_SECRET_...` never echoed; input cleared after submit

### Typecheck

CI runs `coze-studio/scripts/forma/typecheck.mjs`, which invokes real `tsc --noEmit` for:

- `@forma/api-client`
- `@forma/data`
- `@forma/app` (`-p tsconfig.build.json`)

`build: exit 0` is no longer used as a typecheck substitute.

## Local Verification

| Check | Result |
|-------|--------|
| `rush test --only @forma/data` | 35 PASS |
| `rush test --only @forma/api-client` | 1 PASS |
| `rush test --only @forma/app` | 18 PASS |
| `rush test --only @forma/business` | 18 PASS |
| `rush test --only @forma/analyst` | 1 PASS |
| `node scripts/forma/typecheck.mjs` | 3/3 PASS |
| `node scripts/forma/routes-smoke.mjs` | 4/4 PASS |
| Migration | NONE |
| REAL_MODEL_CALLS | 0 |

## Out of Scope

S4-G6 Live E2E / Security / Freeze. Do **not** create `forma-s4-frozen`.

## Delivery

| Item | Value |
|------|-------|
| Commit SHA | `a3c14697a04954e0fdba92c8912dcd772f658d7d` |
| Forma CI | [33700145745](https://github.com/hai12138/forma/actions/runs/33700145745) **ALL GREEN** |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |

## STOP

Do **not** start S4-G6. Do **not** create `forma-s4-frozen`.
