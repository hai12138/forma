# FORMA S3 AI BUSINESS ANALYST RESULT

## Status

**IN_PROGRESS — S3-G2-F1 idempotency + retry claim fixed; Live Model / Browser gates BLOCKED on runtime**

Do **not** mark FROZEN. Do **not** start S4.

| Gate | Status |
|------|--------|
| Analysis idempotency (`createdNew`) | **PASS** (local 100/100 + race) |
| Retry claim CAS | **PASS** (unit/integration) |
| PENDING frontend UX | **PASS** (analysis-pending + retry-analysis) |
| Forma CI (implementation) | **PENDING** — awaiting push + green run on new HEAD |
| Live Model E2E | **BLOCKED_ON_LIVE_MODEL** |
| Browser E2E + UI evidence | **BLOCKED** (runtime) |
| PASS_WITH_REVIEW_GATE | **NOT YET** |

## Frozen Baseline

| Milestone | SHA |
|-----------|-----|
| S2 Frozen Tag `forma-s2-frozen` | `413c3bcc148dfe518b31d6267e1a0c72fc2f0645` |
| S3 initial | `49821f80ad427ab51ea53b984cdbb5482a57be2c` |
| S3-G1 | `e58d97d6d2b2d2b25b25c9d0a46ce01b1903b292` |
| S3-G1-F1 | `f25634efd5f6f8cd334843435412bbec2b8ad169` |
| S3-G1-F2 | `f52e355c99c178c3ff4d4647620a5aef88046dfe` |
| S3-G2 | `7ff6e5830d0494edd5f2cc9f98d98f7324f30e49` |
| S3-G2-F1 (this round) | pending push |

## CI History

| Run | Result | Notes |
|-----|--------|-------|
| `33469903703` | **FAIL** | backend compile, atlas checksum, rush, hook |
| `33471829970` | **FAIL** | proposal syntax; MySQL migration |
| `33474851052` | **FAIL** | migration PASS, frontend PASS, backend FAIL |
| `33475985604` | **PASS** | `f52e355c` — all Forma jobs green |
| `33476030584` | **PASS** | docs RESULT |
| `33476235800` | **PASS** | docs CI |
| `33480332921` | **PASS** | `7ff6e583` — **S3-G2 migration + backend + frontend ALL GREEN** |
| `33484440794` | **FAIL** | docs-only on `2a36bd52` — `TestConcurrentSameClientRequestIdempotent` expected 1 analyst turn, actual 2 |

**Current Forma CI:** NOT GREEN on HEAD until S3-G2-F1 implementation CI passes.

## S3-G2-F1 — Analysis Idempotency / Live Gates (this round)

### BLOCKER-01 — Same `client_request_id` must not re-run analysis

**Root cause:** concurrent duplicate requests both passed external idempotency pre-check; transaction path returned existing User Turn but still invoked `runAnalysisForUserTurn` → duplicate extraction / analyst turns.

**Fix:**

- `createUserTurnWithEvidence(...)` returns `(userTurn, evidence, createdNew bool, err)`
- `SubmitTurn`: `createdNew == false` → `buildIdempotentResult(...)` only; no analysis
- `TestConcurrentSameClientRequestIdempotent`: 10 goroutines, 1 user turn / 1 evidence / 1 extract call / 1 analyst turn / `NextTurnSequence=3`

### STEP 4 — Retry claim CAS

- `ClaimTurnForRetry(tenantID, turnID, expectedStatuses, claimToken)` — DB conditional update
- `RetryTurnAnalysis`: only claimed request runs analysis; others return current persisted state

### STEP 3 — PENDING recovery contract (frontend)

- User turn `PENDING` without analyst reply → `data-testid="analysis-pending"` + Retry Analysis
- `data-testid="start-session"` on 开始访谈 button

### Local test evidence

```text
go test ./domain/forma/analyst/service -run TestConcurrentSameClientRequestIdempotent -count=100  → 100/100 PASS
go test -race ./domain/forma/analyst/service -run "Concurrent|Idempotent|Retry" -count=10         → PASS
go test ./domain/forma/... ./crossdomain/forma/... ./application/forma/... -count=1              → PASS (docker golang:1.24)
```

### Browser harness (STEP 7–22)

- `scripts/forma/s3-browser-e2e.mjs` rewritten:
  - Playwright `BrowserContext.request` shares auth cookies with page
  - Correct selectors: `start-session`, `analyst-input`, `analyst-submit`, `turn-user`, `turn-analyst`
  - Full S3 hard gate structure with asserts + 11 screenshots + log outputs

### Live runtime blockers

Local probe against `127.0.0.1:8888` (2026-09-01):

- `/api/forma/v1/businesses/:id/analyst/sessions` → **404 Not Found**
- `/api/forma/v1/businesses/:id/assertions` → **404 Not Found**

Running backend predates S3 analyst route registration. Additionally `FORMA_ANALYST` builtin chat model must be configured (`modelbuilder.GetBuiltinChatModel`).

**Do not fabricate** `s3-live-model-e2e.log` / screenshots without real model execution.

**Status = `BLOCKED_ON_LIVE_MODEL`** until:

1. Coze backend rebuilt/restarted with current S3 code
2. `FORMA_ANALYST_*` model env configured
3. `FORMA_LIVE_E2E=1 node --test scripts/forma/s3-browser-e2e.mjs` produces logs under `forma/cursor-results/`

## Preserved S3-G2 gates (do not redo)

ContextText → Model, Context Budget, Gap Ask → ANALYST Turn, Production Model Fail Closed, Extraction/Confirmation/Proposal Apply transactions, STALE CAS, Migration, Frontend architecture.

**DO NOT START S4.** No `forma-s3-frozen`.
