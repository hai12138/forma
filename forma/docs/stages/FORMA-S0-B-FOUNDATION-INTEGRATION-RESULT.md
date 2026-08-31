# FORMA S0-B FOUNDATION INTEGRATION RESULT

## Status

**PASS_WITH_GATES**

S0-B foundation scaffolding is complete. Local environment gates (Go toolchain absent, Rush/NPM bootstrap failure) prevent full on-machine verification; CI workflow and node-based smoke scripts provide automated gates on push/PR.

**DO NOT START S1.** Await human review.

---

## Git Baseline

| Check | Result | Evidence |
|---|---|---|
| Branch | `main` | `git branch` |
| HEAD | `98c7aca26d64ac602dc7c0227e2bce38d89666a8` | `git rev-parse HEAD` |
| `origin` → Forma repo | PASS | `https://github.com/hai12138/forma.git` |
| `upstream` → coze-dev/coze-studio | PASS | `https://github.com/coze-dev/coze-studio.git` |
| Tag `forma-baseline-0` | PASS | Points to `98c7aca` (workspace initial commit) |
| STEP 0 gate | PASS | All baseline conditions met |

**Working tree:** Uncommitted S0-B changes present (expected — not committed in this stage).

---

## Coze Baseline

| Item | Value |
|---|---|
| COZE_BASELINE_COMMIT | `fefb05ff27be1da939612fbf9faf5db62583b8ae` (`upstream/main`) |
| Workspace baseline tag | `forma-baseline-0` → `98c7aca26d64ac602dc7c0227e2bce38d89666a8` |
| Runtime foundation | Eino (Coze default) |
| Recorded in code | `backend/domain/forma/meta/version.go` |

---

## Forma Directory Structure

```
coze-studio/
  backend/
    domain/forma/
      asset_registry/          # Entity, DAL, repository, service, tests
      meta/                    # Version/baseline constants
      arch_test.go             # ACL dependency guard
    application/forma/         # Application bootstrap + meta API logic
    crossdomain/forma/
      integration/             # FormaCozeIntegration + CozeAgentAdapter
      impl/                    # CrossDomain default service
    api/
      handler/forma/           # health, version, baseline handlers
      router/forma/            # /api/forma/v1/* registration
  docker/atlas/forma/          # Independent Forma migrations
  frontend/
    apps/forma/                # Independent Forma product shell
    packages/forma-api-client/ # Typed client (health/version/baseline)
  idl/forma/                   # IDL namespace placeholder
  scripts/forma/               # Migration + route smoke validators

forma/
  docs/
    adr/                       # ADR-001 … ADR-007
    architecture/              # COZE-CORE-PATCHES, UPSTREAM-STRATEGY
    stages/                    # This report
  cursor-results/              # Copy of this report
```

**Not created (by design):** Empty placeholder domains beyond `asset_registry`. No Business Model, AI Analyst, Capability Gateway, MCP, Human Task, Evaluation, Release, or Channel implementation.

---

## Backend Foundation

| Component | Status | Notes |
|---|---|---|
| DDD skeleton | DONE | Coze-style layering under `domain/forma`, `application/forma`, `crossdomain/forma` |
| Application init | DONE | Thin `FORMA-BEGIN/END` patches in `application.go` |
| CrossDomain registration | DONE | `crossforma.SetDefaultSVC` with integration adapter |
| Coze core domains untouched | PASS | No edits to `agent`, `workflow`, `plugin`, `knowledge` internals |

---

## Asset Registry

| Item | Status |
|---|---|
| Entity (`AssetRef`, kinds, lifecycle) | DONE |
| Repository contract + GORM impl | DONE |
| Service (`CreateAsset`, constants) | DONE |
| Domain tests (sqlmock) | DONE (source present; local run blocked — no Go) |
| Business/Capability content | NOT IN SCOPE |

**Asset kinds:** BUSINESS, CAPABILITY, AGENT, APPLICATION

**Lifecycle:** DRAFT, IN_REVIEW, VERIFIED, FROZEN, RELEASED, DEPRECATED, ARCHIVED

**Fields:** tenant_id, asset_id, kind, name, semantic_version, revision, schema_version, status, owner_id, created_by, created_at, updated_at, content_digest, deleted_at

---

## Resource Mapping

| Table | Role |
|---|---|
| `forma_asset_ref` | Forma asset header |
| `forma_coze_resource_ref` | Forma asset ↔ Coze resource mapping (`coze_resource_ref` contract) |

**Coze resource types:** AGENT, WORKFLOW, PLUGIN, KNOWLEDGE, APP, DATABASE

**Constraints satisfied:**
- No FK to Coze internal tables
- Stable numeric Coze resource ID references only
- No Forma columns added to Coze core tables

---

## Migration

| Item | Status |
|---|---|
| Independent namespace | `docker/atlas/forma/` |
| Initial migration | `20250831100000_initial.sql` |
| Atlas config | `atlas.hcl`, `migrations/atlas.sum` |
| Documentation | `docker/atlas/forma/README.md` (startup order, Coze vs Forma) |
| Idempotent fresh install | `CREATE TABLE IF NOT EXISTS` |
| Local smoke validation | PASS (see Tests) |
| Live `atlas migrate apply` | NOT RUN locally (no MySQL/Atlas CLI in session) |

---

## API

| Endpoint | Status |
|---|---|
| `GET /api/forma/v1/health` | Implemented |
| `GET /api/forma/v1/version` | Implemented |
| `GET /api/forma/v1/meta/baseline` | Implemented |

**Response envelope:** `{ code, msg, request_id, data }`

**Session middleware:** Public paths whitelisted via `FORMA-BEGIN/END` in `session.go`

**Live HTTP verification:** NOT RUN (backend not started in session)

---

## CrossDomain ACL

| Item | Status |
|---|---|
| `FormaCozeIntegration` interface | DONE |
| `CozeAgentAdapter` (`Describe`, `Health`) | DONE — thin impl via `crossdomain/agent`, not `domain/agent/repository` |
| Architecture test | `domain/forma/arch_test.go` |
| Forma domain → Coze agent repo import | Forbidden by test |
| Integration → crossdomain/agent | Required by test |

---

## Frontend Forma App

| Item | Status |
|---|---|
| App location | `frontend/apps/forma/` (independent from `@coze-studio/app`) |
| Shell + navigation | v1.2 IA (4 groups, 16 routes) |
| Overview (`/`) | Baseline card + `@forma/api-client` hook |
| Design (`/design`) | Token swatches |
| Other 14 routes | Placeholder: "Forma module not connected yet" |
| Mock business store | NOT ADDED |
| localStorage production state | NOT ADDED |

**Routes:** `/`, `/analyst`, `/business`, `/data`, `/capabilities`, `/agents`, `/applications`, `/human`, `/evaluation`, `/releases`, `/channels`, `/runtime`, `/observability`, `/governance`, `/delivery`, `/design`

---

## Design System

| Item | Status |
|---|---|
| Tokens | `frontend/apps/forma/src/styles/tokens.css` (v1.2 Apple-like Enterprise / AI Native) |
| Global Semi/Coze CSS | NOT used as Forma global DS |
| Coze components | Reserved for future Forma Wrapper embed only |

---

## CI

Workflow: `.github/workflows/forma-ci.yml`

| Gate | Trigger |
|---|---|
| Forma Go tests | `go test ./domain/forma/... ./crossdomain/forma/... ./application/forma/...` |
| Migration file present | File existence check |
| Migration smoke (node) | `scripts/forma/migration-validate.mjs` |
| Rush build `@forma/api-client` | typecheck via build |
| Rush build `@forma/app` | production build |
| Vitest route tests | `@forma/app` test script |
| Route + token smoke (node) | `scripts/forma/routes-smoke.mjs` |

**Coze core regression:** Not explicitly gated in Forma CI (unchanged Coze paths); no Coze core logic modified beyond thin FORMA patches.

---

## Tests

| Command | Result | Evidence |
|---|---|---|
| `git status` / baseline verification | PASS | upstream, origin, tag confirmed |
| `node scripts/forma/migration-validate.mjs` | PASS | 2/2 tests ok |
| `node scripts/forma/routes-smoke.mjs` | PASS | 16 routes + design tokens ok |
| `go test ./domain/forma/... ./crossdomain/forma/... ./application/forma/...` | BLOCKED | Go not in PATH on local Windows host |
| `rush install` / `rush build --to @forma/app` | BLOCKED | Rush: "NPM executable does not exist" |
| `GET /api/forma/v1/health` (live) | NOT RUN | Backend not started |
| `GET /api/forma/v1/version` (live) | NOT RUN | Backend not started |
| `GET /api/forma/v1/meta/baseline` (live) | NOT RUN | Backend not started |
| Coze integration health (live) | NOT RUN | Backend not started |
| Git diff review | PASS | 4 core files + new Forma tree only |

---

## Coze Core Files Modified

| File | Change |
|---|---|
| `coze-studio/backend/application/application.go` | Import Forma app; init `formaSVC`; register `crossforma.SetDefaultSVC` |
| `coze-studio/backend/api/router/register.go` | Call `formaRouter.Register(r)` |
| `coze-studio/backend/api/middleware/session.go` | Public paths for Forma meta API |
| `coze-studio/rush.json` | Register `@forma/app`, `@forma/api-client` |

All patches marked with `FORMA-BEGIN` / `FORMA-END`. Documented in `forma/docs/architecture/COZE-CORE-PATCHES.md`.

**Not modified:** `backend/domain/agent`, `workflow`, `plugin`, `knowledge`; Eino runtime; Coze workflow/plugin/knowledge code.

---

## Remaining Mock

| Item | Location | Notes |
|---|---|---|
| Placeholder pages | `frontend/apps/forma/src/pages/index.tsx` | Intentional S0-B stubs |
| CozeAgentAdapter when agent SVC nil | `crossdomain/forma/integration/coze_agent_adapter.go` | Returns unavailable — not a business mock |
| Overview baseline fetch | `use-forma-baseline.ts` | Real API client; fails gracefully when backend down |

No production mock business state or localStorage store.

---

## Known Limitations

1. **Local toolchain:** Go and Rush/pnpm bootstrap unavailable on review host; rely on GitHub Actions for full backend/frontend gates.
2. **Migration apply:** SQL validated statically; live `atlas migrate apply` upgrade path not exercised locally.
3. **API smoke:** Handlers implemented but not hit against running Hertz server in this session.
4. **Forma DB wiring:** Migrations defined; automatic apply on Coze docker-compose startup not yet integrated (manual step documented).
5. **IDL:** S0-B uses hand-written handlers; Thrift IDL generation deferred.
6. **Coze embed wrappers:** Workflow editor / playground embed not started (S1+).

---

## Upstream Compatibility

| Mechanism | Status |
|---|---|
| `forma/docs/architecture/UPSTREAM-STRATEGY.md` | DONE |
| Baseline tag + commit recorded | DONE |
| Core patch inventory | DONE |
| Merge process documented | DONE (not executed) |
| Compatibility gates listed | DONE |

No upstream merge performed in S0-B.

---

## Security Notes

- Forma meta endpoints (`health`, `version`, `baseline`) are intentionally public for platform probes; no sensitive data exposed.
- Forma proprietary code marked UNLICENSED in package metadata; Coze Apache 2.0 attribution preserved in fork.
- Session bypass limited to three explicit paths; no wildcard.
- No secrets committed.

---

## Risks

| Risk | Mitigation |
|---|---|
| FORMA patch blocks in `application.go` conflict on upstream merge | Documented in COZE-CORE-PATCHES + UPSTREAM-STRATEGY |
| Forma migrations drift from Coze docker startup | README documents manual apply order; S1 may automate |
| CI not yet run on remote after S0-B commit | Push + verify Forma CI green before S1 |
| ACL test is import-string based | Extend with `go:generate` or depguard in S1 if needed |

---

## S1 Preconditions

1. Human sign-off on this S0-B report.
2. Forma CI green on `main` after S0-B merge/commit.
3. Live verification: API health/version/baseline + Coze passport still reachable.
4. Forma migration apply integrated or scripted in dev/prod bootstrap.
5. Go + Rush toolchain confirmed on developer machines.
6. Decision gates from S0-A remain in force (Forma shell, Tenant 1:N Space, Eino default, no LangGraph/DeepSeek Harness in V1).

---

**Stage:** FORMA-S0-B FOUNDATION INTEGRATION  
**Completed:** 2026-08-31  
**Next stage:** S1 — **blocked pending human review**
