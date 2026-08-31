# Coze Core Patches (Forma)

Forma minimizes invasive changes to Coze core. All patches use `FORMA-BEGIN` / `FORMA-END` markers for upstream merge review.

## Modified Files

| File | Change | Stage | Risk |
|---|---|---|---|
| `backend/application/application.go` | Forma app init + crossforma registration | S0/S1 | Medium |
| `backend/api/router/register.go` | `formaRouter.Register(r)` | S0 | Low |
| `backend/api/middleware/session.go` | Public Forma meta paths | S0 | Low |
| `rush.json` | `@forma/app`, `@forma/api-client`, `channel-forma` | S0 | Low |
| `common/git-hooks/pre-commit` | Resolve CWD to coze-studio root | S1 | Low |

## New Files (Zero Coze Domain Modification)

- `backend/domain/forma/**` (asset_registry, tenancy, errors, meta)
- `backend/application/forma/**`
- `backend/crossdomain/forma/**`
- `backend/api/handler/forma/**`
- `backend/api/router/forma/**`
- `backend/api/middleware/forma_tenant.go`
- `docker/atlas/forma/**`
- `frontend/apps/forma/**`
- `frontend/packages/forma-api-client/**`

## Rules

1. Never modify `backend/domain/agent`, `workflow`, `plugin`, `knowledge` internals.
2. Never add Forma columns to Coze core tables.
3. Space ACL via `crossdomain/user`, never `domain/user/internal/dal`.
4. Expand Forma via new files; shrink FORMA patch blocks over time.
