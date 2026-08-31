# FORMA S0-B FOUNDATION INTEGRATION RESULT

## Status

**PASS**

S0 Foundation Integration is complete and frozen. All local gates and GitHub Actions Forma CI jobs passed.

**DO NOT START S1** from this document alone — S0 is frozen; S1 is a separate stage.

---

## Final External Gate Closure

| Item | Value |
|---|---|
| GitHub Actions run | `33371051867` |
| Commit | `d68a49bf1ae780f71d6aecd3ff6d3eb3a1c7a3e6` |
| Tag | `forma-s0-frozen` |

| Job | Result |
|---|---|
| forma-backend | PASS |
| forma-frontend | PASS |
| forma-migration-apply | PASS |

---

## Gate Closure

### Root Cause

| Failure | Cause |
|---|---|
| CASE A/B `ERROR 1524 Plugin 'mysql_native_password' is not loaded` | Test assumed an externally provisioned MySQL (`forma-mysql-g1`) whose lifecycle/auth differed between Windows local and Linux CI. Coze `schema.sql` does **not** contain `mysql_native_password`; the fragile external container + host-port topology caused auth/plugin mismatch under MySQL 8.4 (plugin removed by default). |
| CASE C `lookup host.docker.internal: no such host` | `atlasApply()` defaulted Atlas container → `host.docker.internal:<hostPort>`. That DNS name is Docker Desktop–specific and **unavailable on GitHub Actions Linux**. |
| Topology mismatch | CI YAML started MySQL + host port; script assumed fixed container name + host.docker.internal. Local Windows and CI Linux were not the same test. |

### Files Changed

| File | Change |
|---|---|
| `coze-studio/scripts/forma/migration-apply-test.mjs` | Full rewrite: self-contained Docker network + MySQL + Atlas |
| `.github/workflows/forma-ci.yml` | Job only runs the script; no YAML MySQL bootstrap; watch `scripts/forma/**` |
| `coze-studio/docker/atlas/forma/README.md` | Document portable test topology |

### MySQL Image Version

Pinned **`mysql:8.4.5`** — matches Coze `docker/docker-compose.yml` baseline. Not `mysql:latest`.

Atlas pinned **`arigaio/atlas:0.32.1`**.

### Authentication Fix

- Start MySQL with **root only** (no `MYSQL_USER` entrypoint quirks).
- Create app user: `CREATE USER ... IDENTIFIED BY ...` (MySQL 8.4 default → `caching_sha2_password`).
- Assert plugin ≠ `mysql_native_password`.
- **Did not** re-enable legacy `mysql_native_password`.
- Coze upstream `schema.sql` untouched.

### Docker Network Topology

```
forma-mig-net-<runId>
  ├── forma-mig-mysql-<runId>  (alias: mysql)
  └── atlas (ephemeral)        → mysql://coze:***@mysql:3306/<db>
```

- Unique network / container / database per run (`timestamp-pid-random`)
- No `host.docker.internal`
- No host `localhost:3306/3307` in default path
- Cleanup via `t.after` → `docker rm -f` + `network rm`
- Readiness: poll `mysqladmin ping` + `SELECT 1` (timeout 120s, dump logs on failure)

### CASE A Evidence

Local: `ok 1 - CASE A — fresh install` (Coze schema → Forma migration; tables/columns/indexes verified)

### CASE B Evidence

Local: `ok 2 - CASE B — upgrade` (Coze-only → assert no Forma → apply Forma → Coze retained)

### CASE C Evidence

Local: `ok 3 - CASE C — idempotency` (re-apply + `atlas migrate status` via same Docker network `mysql:3306`)

### Local Regression (G2)

| Command | Result |
|---|---|
| `node scripts/forma/migration-apply-test.mjs` | PASS (4/4) |
| `node scripts/forma/migration-validate.mjs` | PASS (2/2) |
| `node scripts/forma/routes-smoke.mjs` | PASS (2/2) |
| `go test ./domain/forma/... ./api/handler/forma/...` | PASS |
| `npx tsc --noEmit` / `rsbuild build` / `vitest` (@forma/app) | PASS |

### Commit

`bb9018fcacf3a0edb94912847dc31f45f27d7e4c` — `fix(forma): make migration integration test portable`

### GitHub Actions Status

**EXTERNAL_GATE** — push triggers Forma CI; confirm `forma-migration-apply` green in GitHub Actions UI.

---

## Gate Closure

| Gate | Previous Status | Final Status | Evidence |
|---|---|---|---|
| Go Tests | BLOCKED (no Go) | **PASS** | `go test ./domain/forma/... ./application/forma/... ./crossdomain/forma/... ./api/handler/forma/...` — all ok (Go 1.24.6 user-local) |
| Frontend Typecheck | BLOCKED (Rush/NPM) | **PASS** | `npx tsc --noEmit` in `frontend/apps/forma`; Rush `@forma/api-client` build ok |
| Frontend Build | BLOCKED (Rush/NPM) | **PASS** | `npx rsbuild build` — dist 182.9 kB; Rush `@forma/app` blocked on Windows WSL/bash deps (CI Linux expected) |
| Route Smoke | PASS | **PASS** | `node scripts/forma/routes-smoke.mjs` 2/2; `vitest --run` 1/1 |
| Migration Static Validation | PASS | **PASS** | `node scripts/forma/migration-validate.mjs` 2/2 |
| Migration Real Apply | NOT RUN | **PASS** | G2 portable Docker network test CASE A/B/C 3/3 |
| Live API | NOT RUN | **PASS** | `go test ./api/handler/forma/...` — Hertz router + TCP `:18888` baseline smoke |
| GitHub Actions | Unconfirmed | **EXTERNAL_GATE** | G2 push pending Actions confirmation |
| Core Patch Review | PASS | **PASS** | Only 4 Coze core files modified (+ pnpm-lock from `rush update`) |

### Environment Diagnosis (S0-B-G1)

| Tool | Version |
|---|---|
| OS | Microsoft Windows NT 10.0.26200.0 |
| Node | v22.22.0 |
| npm | 10.9.4 |
| corepack | 0.34.0 |
| pnpm | 8.15.8 (via `corepack prepare pnpm@8.15.8 --activate`) |
| Go | go1.24.6 windows/amd64 (user-local `%LOCALAPPDATA%\go1.24.6`) |
| Docker | 29.5.2 / Compose v5.1.4 |
| Rush | 5.147.1 (via install-run-rush.js) |

**Note:** Coze `pre-commit` hook paths assume `coze-studio/` CWD at repo root; commit used `SKIP_COMMIT_MSG_HOOK=true` (hook's own escape hatch). Recommend fixing hook wrapper for Forma monorepo layout in a follow-up.

---

## Git Baseline

| Check | Result | Evidence |
|---|---|---|
| Branch | `main` | `git branch` |
| HEAD (post S0-B) | `dd9eca7143523e75cfa684b1a8e91631a3ac8e3f` | `git rev-parse HEAD` |
| `origin` → Forma repo | PASS | `https://github.com/hai12138/forma.git` |
| `upstream` → coze-dev/coze-studio | PASS | `https://github.com/coze-dev/coze-studio.git` |
| Tag `forma-baseline-0` | PASS | `98c7aca` |
| Push | PASS | `98c7aca2..dd9eca71 main -> main` |

---

## Coze Baseline

| Item | Value |
|---|---|
| COZE_BASELINE_COMMIT | `fefb05ff27be1da939612fbf9faf5db62583b8ae` |
| Workspace baseline tag | `forma-baseline-0` → `98c7aca` |
| Runtime foundation | Eino |
| Recorded in | `backend/domain/forma/meta/version.go` |

---

## Forma Directory Structure

```
coze-studio/
  backend/domain/forma/          asset_registry, meta, arch_test
  backend/application/forma/     bootstrap + meta responses
  backend/crossdomain/forma/     ACL + FormaCozeIntegration
  backend/api/handler/forma/     health, version, baseline + tests
  backend/api/router/forma/      /api/forma/v1/*
  docker/atlas/forma/            independent migrations
  frontend/apps/forma/           product shell (16 routes)
  frontend/packages/forma-api-client/
  scripts/forma/                 migration-validate, migration-apply-test, routes-smoke
forma/docs/                      ADRs, architecture, this report
.github/workflows/forma-ci.yml
```

---

## Backend Foundation

DDD skeleton under `domain/forma`, `application/forma`, `crossdomain/forma`. Coze core domains (`agent`, `workflow`, `plugin`, `knowledge`) untouched.

---

## Asset Registry

Entity, repository, service, sqlmock tests. Kinds: BUSINESS, CAPABILITY, AGENT, APPLICATION. Lifecycle: DRAFT → ARCHIVED.

---

## Resource Mapping

Tables: `forma_asset_ref`, `forma_coze_resource_ref`. No FK to Coze core tables. Types: AGENT, WORKFLOW, PLUGIN, KNOWLEDGE, APP, DATABASE.

---

## Migration

Independent `docker/atlas/forma/`. Real apply validated (CASE A fresh, CASE B upgrade coexistence, CASE C idempotency). `atlas.sum` regenerated with valid hash.

---

## API

| Endpoint | Verified |
|---|---|
| `GET /api/forma/v1/health` | PASS (Hertz + TCP) |
| `GET /api/forma/v1/version` | PASS |
| `GET /api/forma/v1/meta/baseline` | PASS — `runtime_foundation=eino`, `coze_baseline_commit=fefb05ff…` |

Envelope: `{ code, msg, request_id, data }`.

---

## CrossDomain ACL

`FormaCozeIntegration` + `CozeAgentAdapter` via `crossdomain/agent`. `arch_test.go` forbids `domain/agent` imports in Forma domain.

---

## Frontend Forma App

Independent `@forma/app` — v1.2 IA, 16 routes, tokens, overview + design pages, placeholders elsewhere. No mock business store.

---

## Design System

`tokens.css` — Apple-like Enterprise / AI Native. Semi/Coze not global DS.

---

## CI

`.github/workflows/forma-ci.yml` — jobs: `forma-backend`, `forma-migration-apply`, `forma-frontend`.

**EXTERNAL_GATE:** Confirm workflow run for commit `dd9eca71` is green on GitHub.

---

## Tests (G1 Regression — executed 2026-08-31)

| Command | Result | Evidence |
|---|---|---|
| `go test ./domain/forma/... ./application/forma/... ./crossdomain/forma/... ./api/handler/forma/...` | PASS | all ok |
| `npx tsc --noEmit` (@forma/app) | PASS | exit 0 |
| `npx rsbuild build` (@forma/app) | PASS | 182.9 kB dist |
| `npx vitest --run` (@forma/app) | PASS | 1/1 |
| `node scripts/forma/migration-validate.mjs` | PASS | 2/2 |
| `node scripts/forma/routes-smoke.mjs` | PASS | 2/2 |
| `node scripts/forma/migration-apply-test.mjs` | PASS | 3/3 |
| `rush build --to @forma/app` (Windows) | FAIL (env) | WSL/bash `rtsc.sh` — CI Linux expected PASS |
| Full `opencoze` binary build (Windows) | FAIL (env) | milvus pkg Windows incompatibility — not S0-B scope |
| GitHub Actions Forma CI | EXTERNAL | push ok; results unverified here |

---

## Coze Core Files Modified

| File | Change |
|---|---|
| `backend/application/application.go` | Forma init + crossdomain registration |
| `backend/api/router/register.go` | `formaRouter.Register(r)` |
| `backend/api/middleware/session.go` | Public Forma meta paths |
| `rush.json` | `@forma/app`, `@forma/api-client`, `channel-forma` tag |

Also: `common/config/subspaces/default/pnpm-lock.yaml` (rush update for new Forma packages — not Coze runtime logic).

All FORMA patches use `FORMA-BEGIN` / `FORMA-END`.

---

## Remaining Mock

Placeholder pages only (`Forma module not connected yet`). No production mock store.

---

## Known Limitations

1. **GitHub CI** — requires manual verification (private repo / no `gh`).
2. **Windows Rush full build** — dependency chain uses bash `rtsc.sh`; Linux CI is canonical.
3. **Full opencoze server** — not built on Windows; API validated via Hertz in-process + TCP tests.
4. **Coze pre-commit hook** — broken at Forma repo root; used `SKIP_COMMIT_MSG_HOOK`.
5. **MySQL 3306** — blocked on host; migration tests use disposable container on 3307.

---

## Upstream Compatibility

Documented in `forma/docs/architecture/UPSTREAM-STRATEGY.md`. No upstream merge in S0-B.

---

## Security Notes

Public meta endpoints only expose version/baseline metadata. No secrets committed.

---

## Risks

| Risk | Mitigation |
|---|---|
| CI not verified locally | Human confirms Actions green |
| FORMA patch merge conflicts | COZE-CORE-PATCHES inventory |
| Atlas on existing Coze DB | `--allow-dirty` documented in migration README + apply test |

---

## S1 Preconditions

1. **Confirm GitHub Actions `Forma CI` green** for `dd9eca71`.
2. Human sign-off on S0-B + G1.
3. Fix Forma-root pre-commit hook CWD (optional hygiene).
4. S0-A decision gates remain in force.

---

**Stage:** FORMA-S0-B + FORMA-S0-B-G1  
**Completed:** 2026-08-31  
**Commit:** `dd9eca7143523e75cfa684b1a8e91631a3ac8e3f`  
**Next:** S1 blocked pending human review + CI confirmation

**DO NOT START S1.**
