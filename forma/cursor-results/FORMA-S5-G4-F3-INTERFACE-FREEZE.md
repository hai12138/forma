# S5-G4-F3 — Frozen Ownership & Contracts

**Baseline main:** `61657777005d84f0b7a91e99a57209769e18fa56`  
**F2 implementation:** `37c7c1fc36e32e12c4b4cc7d261ee78144d2aab4`  
**Status:** FROZEN — candidates only; do NOT declare PASS; do NOT push main.

---

## Agent A — Secret Assignment (ONLY)

```
domain/forma/capability/service/validator.go
domain/forma/capability/service/audit_metadata.go
domain/forma/capability/service/*_f3_test.go  (new)
```

1. Keep credential-shape; do NOT restore bare keyword bans.
2. Extend `freeTextAssignmentPatterns` (and Bearer shape) to reject:
   - `token=<non-empty>`, `cookie=<non-empty>`, `secret=<non-empty>`, `session=<non-empty>`
   - `client_secret=`, `access_token=`, `refresh_token=`
   - `authorization:` / `authorization=`
   - `Bearer <non-empty>` (any non-empty token after Bearer, not only ≥8)
3. Continue JWT / ghp_ / github_pat_ / sk- / sk-proj- / xox* / PEM / long random / long base64.
4. Allow: `review trade secret policy`, `tokenization completed`, `cookie policy approved`, `password reset workflow approved`, `bearer of responsibility`.
5. Reason vs ClientRequestID semantics unchanged (audit_metadata already splits); ensure assignment forms apply to reason via `containsSecret`.
6. No-partial-write tests for illegal metadata.

Branch: `forma/s5-g4-f3-agent-a-secret`

## Agent B — Persisted FAILED Totality (ONLY)

```
domain/forma/capability/service/service.go
application/forma/capability_analysis_app.go
domain and/or application *_f3_test.go (new; prefer domain for service paths, app for HTTP-200 DTO if needed)
```

1. Shared helper e.g. `failedAnalysisResultAfterMark(ctx, tenantID, analysisRunID, attempt, domainErr) (*AnalysisResult, error)`:
   - After mark-failed path succeeds, **re-fetch** run; only if `Status == FAILED` return `&AnalysisResult{Run: failed, OwnedExecute: true}, sanitizedErr`.
   - If mark or refetch fails → return nil result + error (fail closed; no forged DTO).
2. Wire all paths:
   - `executeAnalysis` generator error / invalid proposal (already similar — consolidate)
   - `RetryFailedAnalysis` when `loadValidatedPersistedAnalysisRequest` fails
   - expired PENDING lease takeover when load fails
3. App keeps `keepFailedAnalysisDTO` / `capabilityAnalysisResultDTO` (status-based HTTP 200).
4. Retry: same analysis_run_id, attempt monotonic, no second run, no duplicate proposals.
5. Tests listed in user brief.

**Do NOT** edit validator.go / audit_metadata.go (Agent A).

Branch: `forma/s5-g4-f3-agent-b-failed`

## Main integrator

- Merge A→B; ensure one secret detection path and one FAILED-after-persist helper.
- `git diff --check 61657777...HEAD` exit 0.
- Full tests + RESULT + CI; `S5_G5_READY = NO`.
- MIGRATION_CHANGE=NONE; REAL_MODEL_CALLS=0; G5_CHANGE=NONE.
