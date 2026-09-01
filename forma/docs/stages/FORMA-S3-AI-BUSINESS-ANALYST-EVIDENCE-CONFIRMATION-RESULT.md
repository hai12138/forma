# FORMA S3 AI BUSINESS ANALYST RESULT

## Status

**BLOCKED_ON_LIVE_MODEL — engineering + core cleanup complete; Live Model E2E requires human Coze chat model credential**

Do **not** mark FROZEN. Do **not** start S4.

| Gate | Status |
|------|--------|
| S3 Live Engineering commit `ec79261c` | **PASS** |
| Engineering CI `33493675844` | **PASS** — forma-backend / migration / frontend |
| Core cleanup commit `e3cac805` | **PUSHED** — removed `InitLiveHarness` Coze core delta |
| Core cleanup CI | **PENDING** — awaiting run on `e3cac805` |
| `go test` domain/crossdomain/application forma | **PASS** |
| Live harness build | **PASS** |
| Runtime route probe | **PASS** — `forma/cursor-results/s3-route-probe.log` |
| Real Model probe | **BLOCKED** — no chat model configured |
| Browser E2E | **BLOCKED** — depends on real model |
| No Silent Mutation / Provenance / Tenant Isolation | **NOT RUN** — blocked on live model |
| UI Evidence (`s3-ui/*.png`) | **NOT RUN** — blocked on live model |
| **PASS_WITH_REVIEW_GATE** | **NOT YET** |

## HUMAN_ACTION_REQUIRED

```
REAL COZE CHAT MODEL CREDENTIAL
```

Live runtime check (presence only, no secrets logged):

| Check | Result |
|-------|--------|
| `BUILTIN_CM_TYPE` on `forma-live-harness` | **UNSET** |
| `BUILTIN_CM_DEEPSEEK_*` | **UNSET** |
| `BUILTIN_CM_QWEN_*` | **UNSET** |
| `model_instance` rows in `forma-live-mysql` | **0** |

**Do not** use `DeterministicFakeModel`, `fake-analyst`, mock provider, or fabricated E2E logs.

To unblock (local / Docker env only — never commit secrets):

```bash
# Example: DeepSeek via Coze BUILTIN_CM contract
docker stop forma-live-harness
docker run -d --name forma-live-harness \
  -p 127.0.0.1:8888:8888 \
  -e MYSQL_DSN='coze:coze123@tcp(forma-live-mysql:3306)/opencoze?charset=utf8mb4&parseTime=True' \
  -e REDIS_ADDR='forma-live-redis:6379' \
  -e BUILTIN_CM_TYPE=deepseek \
  -e BUILTIN_CM_DEEPSEEK_API_KEY='<LOCAL_SECRET>' \
  -e BUILTIN_CM_DEEPSEEK_MODEL='deepseek-chat' \
  -v .../coze-studio/bin/forma-live-harness:/app/forma-live-harness \
  --network ... debian:bookworm-slim /app/forma-live-harness

FORMA_LIVE_E2E=1 \
FORMA_LIVE_BASE_URL=http://127.0.0.1:8888 \
FORMA_UI_BASE_URL=http://127.0.0.1:5173 \
node --test scripts/forma/s3-browser-e2e.mjs
```

## Frozen Baseline

| Milestone | SHA |
|-----------|-----|
| S2 Frozen `forma-s2-frozen` | `413c3bcc148dfe518b31d6267e1a0c72fc2f0645` |
| S3-G2-F1 | `65f3d6106609840998c40a1dc8f41b1d16f84bc5` |
| S3 Live Engineering Candidate | `ec79261cb74a8627f82c84bc420f3879410038b4` |
| S3 Core Cleanup | `e3cac805` — `fix(forma): remove S3 live harness Coze core delta` |
| S3 Freeze Candidate | **NOT YET** — blocked on live model |

## CI History

| Run | Result | Notes |
|-----|--------|-------|
| `33484440794` | **FAIL** | `TestConcurrentSameClientRequestIdempotent` expected 1 / actual 2 |
| `33489002205` | **PASS** | `65f3d610` — ALL GREEN after idempotency fix |
| `33493675844` | **PASS** | `ec79261c` — S3 Live Engineering Candidate ALL GREEN |
| (pending) | — | `e3cac805` — core cleanup |

## S3 Live Final Gate

### Core Cleanup (this round)

Removed unnecessary Coze Runtime Core delta:

- **Deleted** `InitLiveHarness(...)` from `backend/bizpkg/config/config.go`
- **Kept** historical `InitBaseForLiveHarness(...)`
- **Changed** `forma-live-harness` to use existing `bizconfig.Init(ctx, db, oss)` with error handling
- S3 frozen delta no longer includes new Coze core helper API

### FIX-01 — PENDING Retry Lease

- Constant `analysisPendingLease = 5 * time.Minute`
- Initial analysis sets `model_request_id = analysis_claim:<request_id>`
- PENDING retry blocked while active lease; abandoned after lease expiry
- `ClaimTurnForRetry` CAS respects expired `analysis_claim:*`

### FIX-02 — Browser Hard Gate (no SKIP)

- `scripts/forma/s3-e2e-fixtures.mjs` — deterministic Gap / Conflict / Edit fixtures via test DB only
- `scripts/forma/s3-browser-e2e.mjs` — all mandatory gates assert PASS or FAIL (no SKIPPED)
- Playwright `BrowserContext.request` shared auth; correct UI selectors

### Runtime

- Rebuilt `forma-live-harness` from `e3cac805` (uses `bizconfig.Init`, not `InitLiveHarness`)
- Applied Forma migrations to `forma-live-mysql` (S0–S3-G2)
- Route probe after restart: **PASS**
  - Evidence: `forma/cursor-results/s3-route-probe.log`

### Live Model blocker

Coze builtin model via `modelbuilder.GetBuiltinChatModel` requires **Coze Model Manager config**:

- `BUILTIN_CM_TYPE` + matching `BUILTIN_CM_*_API_KEY` / `BUILTIN_CM_*_MODEL` on `forma-live-harness`, **or**
- `model_instance` rows in live MySQL

Neither is present in current live stack.

## Required Evidence (not yet produced)

| Artifact | Status |
|----------|--------|
| `s3-live-model-e2e.log` | **MISSING** |
| `s3-browser-e2e.log` | **INCOMPLETE** (partial run only) |
| `s3-provenance-e2e.log` | **MISSING** |
| `s3-tenant-isolation-e2e.log` | **MISSING** |
| `s3-ui/01-session-started.png` … `11-tenant-switch.png` | **MISSING** |

## Preserved (do not redo)

Context Builder, Evidence, Assertion, Confirmation, Proposal, Transaction, Gap Ask, Tenant, Business Model architecture.

**DO NOT START S4.** No `forma-s3-frozen` tag until live gates pass and human review completes.
