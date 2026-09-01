# FORMA S3 AI BUSINESS ANALYST RESULT

## Status

PASS_WITH_GATES

S3 initial implementation: `49821f80` — PASS_WITH_GATES  
S3-G1 closure (this round): **FAIL / G1 IN PROGRESS** — CI blockers addressed locally; Live Model + Browser E2E still pending runtime.

## Frozen Baseline

- S2 Frozen Tag: `forma-s2-frozen` → `413c3bcc148dfe518b31d6267e1a0c72fc2f0645`
- Post-Freeze main: `65b165dcd84d1b9f9c083e87d65bbf16f9061ad1`
- S3 implementation: `49821f80ad427ab51ea53b984cdbb5482a57be2c`

## Initial S3 CI

- Run: `33469903703`
- Result: **FAIL** (`forma-backend`, `forma-migration-apply`, `forma-frontend`)
- Root causes: backend compile (`CreateModelCall` receiver shadow), Atlas checksum mismatch, Rush shrinkwrap stale, `post-rush-install.sh` mode `100644`

## S3-G1 Closure (this round)

### CI / Infra fixes

| Item | Fix |
|------|-----|
| BLOCKER-01 Backend compile | `CreateModelCall(ctx, record *entity.ModelCallRecord)` |
| BLOCKER-02 Atlas checksum | `arigaio/atlas:0.32.1 migrate hash`; `atlas.sum` includes G1 migration |
| BLOCKER-03 Rush shrinkwrap | `rush update` — `pnpm-lock.yaml` includes `@forma/analyst` |
| BLOCKER-04 Hook executable | `git update-index --chmod=+x scripts/hooks/post-rush-install.sh` → `100755` |
| STEP 3 G1 migration | `20250902010000_s3_g1_integrity.sql` (turn sequence + conflict pair unique indexes) |

### Production model path (fail-closed)

- `CozeEinoAnalystModel`: **no** `DeterministicFakeModel` fallback on extract/generate failure
- `NewAnalystService`: `Model == nil` → `NewUnavailableAnalystModel()` (not fake)
- `GenerateInterviewTurn`: always invokes Eino; planner question passed as model context
- Model call observability: `ExtractAssertions` + `GenerateInterviewTurn` recorded

### Transactional integrity

- `ConfirmAssertion` / `RejectAssertion`: single repo transaction; confirmation event atomic
- Edit-before-confirm: original AI assertion → `SUPERSEDED`; new `MANUAL_MODIFIED` derived assertion + confirmation
- `ApplyProposal`: cross-domain transaction via `gorm.DB` + analyst/business repos; no ignored provenance/status errors

### API / UX

- `POST .../turns/:turnId/retry-analysis`
- `GET .../proposals/:proposalId/preview` (S2 semantic diff)
- Frontend: tenant hard reset, Retry Analysis, Edit & Confirm, evidence↔assertion links, conflict detail, proposal semantic diff

### Tests added

- Backend integration: digest deterministic, AI assertions PROPOSED, duplicate client_request_id, cross-session proposal rejection, edit-before-confirm provenance, confirmation rollback, conflict dedup
- Frontend: analyst workspace smoke / tenant switch test scaffold
- Migration validate: S3-G1 migration + atlas.sum entry

### Not completed (gates remain)

| Gate | Status |
|------|--------|
| G16 Live Model E2E | **BLOCKED/Pending** — requires configured Coze/Eino builtin chat model in CI/runtime |
| G17 Browser E2E | **Pending** — Playwright full Interview→Confirm→Proposal→Apply not executed locally |
| G18 UI screenshots | **Pending** — `forma/cursor-results/s3-ui/` |
| Forma CI ALL GREEN | **Pending push verification** |

## Files Changed (S3-G1)

### Backend
- `backend/domain/forma/analyst/service/` — analysis, confirm/apply/retry, integrity, unavailable model
- `backend/crossdomain/forma/integration/analyst_model_adapter.go`
- `backend/application/forma/analyst_app.go`, `forma.go`
- `backend/api/handler/forma/analyst.go`, `api/router/forma/api.go`
- `docker/atlas/forma/migrations/20250902010000_s3_g1_integrity.sql`, `atlas.sum`

### Frontend
- `frontend/packages/forma-analyst/` — workspace UX, tests
- `frontend/packages/forma-api-client/` — retry + proposal preview APIs
- `common/config/subspaces/default/pnpm-lock.yaml`

### Docs / CI
- `scripts/forma/migration-validate.mjs`
- This RESULT file

## Remaining Mock (allowed scope)

- `DeterministicFakeModel` **only** via explicit test DI (`integration_test.go`, unit tests)
- **Not** used in production wiring (`forma.go` uses `CozeEinoAnalystModel`)

## Gate Evidence (updated)

| Gate | Status |
|------|--------|
| G01–G15 | Implemented + tests (partial expanded) |
| G16 Live model E2E | **Pending / BLOCKED without real model** |
| G17 Browser E2E | **Pending** |
| G18 Work order interview | Heuristic in **tests only** |
| G19 S2→S3 migration | SQL + atlas.sum (incl. G1) |
| G20 S0/S1/S2 regression | CI retains prior gates |
| G21 Forma CI | **Pending green run after push** |

## S4 Preconditions

S3 freeze + human review. Browser/live E2E gates GREEN. **Do not start S4.**

---

**DO NOT START S4.** Awaiting human review. **No `forma-s3-frozen` tag.**
