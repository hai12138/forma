# FORMA S3 AI BUSINESS ANALYST RESULT

## Status

**IN_PROGRESS — S3 Live Final Gate: engineering closure complete; Live Model E2E BLOCKED on Coze Model config**

Do **not** mark FROZEN. Do **not** start S4.

| Gate | Status |
|------|--------|
| CI `33489002205` (HEAD `65f3d610`) | **PASS** — forma-backend / migration / frontend |
| Analysis idempotency (`createdNew`) | **PASS** |
| PENDING retry lease (`analysisPendingLease`) | **PASS** (tests) |
| Browser harness (no SKIP + fixtures) | **PASS** (script ready) |
| Runtime rebuild + S3 routes | **PASS** (`s3-route-probe.log`) |
| S3 migrations on live MySQL | **PASS** |
| Real Model + Browser E2E | **BLOCKED_ON_LIVE_MODEL** |
| **PASS_WITH_REVIEW_GATE** | **NOT YET** |

## Frozen Baseline

| Milestone | SHA |
|-----------|-----|
| S2 Frozen `forma-s2-frozen` | `413c3bcc148dfe518b31d6267e1a0c72fc2f0645` |
| S3-G2-F1 | `65f3d6106609840998c40a1dc8f41b1d16f84bc5` |
| S3 Live Final (this round) | pending push — freeze candidate |

## CI History

| Run | Result | Notes |
|-----|--------|-------|
| `33484440794` | **FAIL** | `TestConcurrentSameClientRequestIdempotent` expected 1 / actual 2 |
| `33489002205` | **PASS** | `65f3d610` — **ALL GREEN** after idempotency fix |

**Current Forma CI:** **PASS** on `33489002205`.

## S3 Live Final Gate (this round)

### FIX-01 — PENDING Retry Lease

- Constant `analysisPendingLease = 5 * time.Minute`
- Initial analysis sets `model_request_id = analysis_claim:<request_id>`
- PENDING retry blocked while active lease; abandoned after lease expiry
- `ClaimTurnForRetry` CAS respects expired `analysis_claim:*`
- Tests: `TestActivePendingCannotRetry`, `TestAbandonedPendingCanRetryAfterLease`, `TestConcurrentRetryOnlyOneClaim`

### FIX-02 — Browser Hard Gate (no SKIP)

- `scripts/forma/s3-e2e-fixtures.mjs` — deterministic Gap / Conflict / Edit fixtures via test DB only
- `scripts/forma/s3-browser-e2e.mjs` — all mandatory gates assert PASS or FAIL (no SKIPPED)
- Playwright `BrowserContext.request` shared auth; correct UI selectors

### Runtime rebuild

- Rebuilt `forma-live-harness` from current repo (`InitLiveHarness` + ModelConf)
- Applied Forma migrations to `forma-live-mysql` (S0–S3-G2)
- Route probe: `node scripts/forma/s3-live-route-probe.mjs` → **PASS**
  - Evidence: `forma/cursor-results/s3-route-probe.log`

### Live Model blocker

Coze builtin model via `modelbuilder.GetBuiltinChatModel` requires **Coze Model Manager config**:

- `BUILTIN_CM_TYPE` + matching `BUILTIN_CM_*_API_KEY` / `BUILTIN_CM_*_MODEL` on `forma-live-harness`, **or**
- `model_instance` rows in live MySQL

Local harness has **no configured chat model** (empty `BUILTIN_CM_*`, `model_instance` count = 0, no Ollama on `:11434`).

**Do not** use `DeterministicFakeModel` or fabricate `s3-live-model-e2e.log`.

To unblock:

```bash
# Example: DeepSeek via Coze BUILTIN_CM (not FORMA_ANALYST_*)
docker stop forma-live-harness
docker run -d --name forma-live-harness \
  -p 127.0.0.1:8888:8888 \
  -e MYSQL_DSN='coze:coze123@tcp(forma-live-mysql:3306)/opencoze?charset=utf8mb4&parseTime=True' \
  -e REDIS_ADDR='forma-live-redis:6379' \
  -e BUILTIN_CM_TYPE=deepseek \
  -e BUILTIN_CM_DEEPSEEK_API_KEY='***' \
  -e BUILTIN_CM_DEEPSEEK_MODEL='deepseek-chat' \
  -v .../coze-studio/bin/forma-live-harness:/app/forma-live-harness \
  --network ... debian:bookworm-slim /app/forma-live-harness

FORMA_LIVE_E2E=1 node --test scripts/forma/s3-browser-e2e.mjs
```

## Preserved (do not redo)

Context Builder, Evidence, Assertion, Confirmation, Proposal, Transaction, Gap Ask, Tenant, Business Model architecture.

**DO NOT START S4.** No `forma-s3-frozen` tag until human review after live gates pass.
