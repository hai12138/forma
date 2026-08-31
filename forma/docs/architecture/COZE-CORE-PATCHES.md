# Coze Core Patches (Forma S0-B)

Forma minimizes invasive changes to Coze core. All patches use `FORMA-BEGIN` / `FORMA-END` markers for upstream merge review.

## Modified Files

| File | Change | Risk |
|---|---|---|
| `backend/application/application.go` | Import Forma app; init `formaSVC` in `initBasicServices`; register `crossforma.SetDefaultSVC` | Medium |
| `backend/api/router/register.go` | Call `formaRouter.Register(r)` at INSERT_POINT | Low |
| `backend/api/middleware/session.go` | Public paths for `/api/forma/v1/health`, `version`, `meta/baseline` | Low |
| `rush.json` | Add `@forma/app`, `@forma/api-client` projects | Low |

## New Files (Zero Coze Core Modification)

- `backend/domain/forma/**`
- `backend/application/forma/**`
- `backend/crossdomain/forma/**`
- `backend/api/handler/forma/**`
- `backend/api/router/forma/**`
- `docker/atlas/forma/**`
- `frontend/apps/forma/**`
- `frontend/packages/forma-api-client/**`
- `idl/forma/README.md`

## Rules

1. Never modify `backend/domain/agent`, `workflow`, `plugin`, `knowledge` internals.
2. Never add Forma columns to Coze core tables.
3. Expand Forma via new files; shrink FORMA patch blocks over time if upstream adds extension hooks.
