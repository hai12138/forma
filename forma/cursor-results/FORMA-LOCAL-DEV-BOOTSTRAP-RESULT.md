# FORMA-LOCAL-DEV-BOOTSTRAP-RESULT

**Date:** 2026-09-03  
**Scope:** DevEx only — local one-command launcher + docs  
**DO NOT START:** S4-G6-F2 / S5  
**Tag:** `forma-s4-frozen` untouched

## Verdict

| Gate | Result |
|---|---|
| FORMA_LOCAL_DEV_BOOTSTRAP | **PASS** |
| PRODUCT_DOMAIN_CHANGE | **NONE** |
| REAL_MODEL_CALLS | **0** |

## FILES_CREATED

| Path | Role |
|---|---|
| `forma/docs/LOCAL-DEVELOPMENT.md` | Canonical Chinese local-dev guide |
| `coze-studio/scripts/forma/local/forma-local.mjs` | Shared launcher core |
| `coze-studio/scripts/forma/local/forma-local.ps1` | Windows entry |
| `coze-studio/scripts/forma/local/forma-local.sh` | macOS/Linux entry |
| `coze-studio/scripts/forma/local/.forma-local.env.example` | Forma-only optional overrides |
| `coze-studio/scripts/forma/local/docker-compose.ports.override.yml` | Host MySQL 3306 conflict → publish `13306` via `!override` |
| `coze-studio/scripts/forma/local/tests/forma-local-smoke.test.mjs` | Deterministic smoke tests |
| `forma/cursor-results/FORMA-LOCAL-DEV-BOOTSTRAP-RESULT.md` | This file |

## FILES_UPDATED

| Path | Change |
|---|---|
| `coze-studio/.gitignore` | ignore `.forma-local/` |
| `.gitignore` | ignore `coze-studio/.forma-local/` |
| `coze-studio/Makefile` | thin `forma-local*` wrappers |

## WINDOWS_SCRIPT

`PASS` — PowerShell parse OK; `forma-local.ps1` → node core.

## BASH_SCRIPT

`PASS` — `bash -n` OK.

## DOCUMENTATION

`PASS` — `forma/docs/LOCAL-DEVELOPMENT.md` (modes A/B, exact commands, pages, auth, AI, secret, FAQ, Windows/macOS, reset, known G6-F2 UI gaps).

## DOCTOR

`PASS` — toolchain / ports / files / secret WARN without mutating system.

## MIDDLEWARE_START

`PASS` — `docker compose -f docker-compose-debug.yml --profile middleware` (reused). Host 3306 busy → override publish `13306`.

## MIGRATION

`PASS` — `atlas migrate apply --allow-dirty` on existing Forma migrations only (idempotent). No schema generate / hash rewrite.

## BACKEND_START

`PASS` — real `backend/main.go` → `bin/opencoze-linux` (Docker golang build; reuse unless `FORMA_LOCAL_REBUILD_BACKEND=1`). Run as `forma-local-backend` on compose network. **Not** `forma-live-harness`.

## FRONTEND_START

`PASS` — Rush `install-run-rushx.js dev` for `@forma/app` on `:3001`, proxy `/api/forma` → `:8888`.

## STATUS / LOGS / STOP

`PASS` — status health lines; logs backend/frontend; stop leaves middleware; `--all` available.

## PAGE_SMOKE

`PASS` — HTTP 200 HTML for `/`, `/business`, `/analyst`, `/data`, `/data/requirements`, `/data/sources`, `/data/mappings`, `/data/contracts`, `/data/health`. Backend + FE proxy `/api/forma/v1/health` → 200.

## PORT_CONFLICT

`PASS` — occupied `:3001` → fail fast with useful error (no silent 3002).

## PID_SAFETY

`PASS` — only recorded frontend PID + `forma-local-backend` container; stale PID detection in tests.

## SECRET_SAFETY

`PASS` — logs/status redact; only `configured = YES|NO` (no master key value).

## WINDOWS_SUPPORT / MAC_LINUX_SUPPORT

`PASS` — PS1 first-class; SH for Unix; shared mjs; no WSL requirement. Native `server.sh` preferred when Go+bash available; else Docker opencoze path.

## Manual smoke (this machine)

```text
doctor → PASS
start  → READY (FE :3001, BE :8888)
status → Backend/Forma API/Frontend OK
pages  → all listed routes HTTP 200
stop   → app DOWN, middleware left up
```

## KNOWN_PRODUCT_ISSUES (not startup failures)

- Mapping EditConfirm UI pending G6-F2
- Drift Snapshot Picker pending G6-F2

## Delivery

| Item | Value |
|---|---|
| COMMIT_SHA | `c3feae76ac53a4e2915cd23b7e620d7171cb0199` |
| CI_RUN | [33727360670](https://github.com/hai12138/forma/actions/runs/33727360670) |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |

## Commands

**Windows**

```powershell
cd <repo>\coze-studio
.\scripts\forma\local\forma-local.ps1 doctor
.\scripts\forma\local\forma-local.ps1 start
```

**macOS / Linux**

```bash
cd <repo>/coze-studio
./scripts/forma/local/forma-local.sh doctor
./scripts/forma/local/forma-local.sh start
```

FRONTEND_URL = http://localhost:3001  
BACKEND_URL = http://localhost:8888

## STOP

Do **not** start S4-G6-F2. Do **not** start S5. Do **not** move `forma-s4-frozen`.
Await human local start check.
