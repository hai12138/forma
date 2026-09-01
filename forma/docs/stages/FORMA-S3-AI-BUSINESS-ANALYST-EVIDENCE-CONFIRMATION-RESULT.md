# FORMA S3 AI BUSINESS ANALYST RESULT

## Status

**IN_PROGRESS — S3-G1-F2 backend green; awaiting CI + Live Model + Browser gates**

Do **not** mark FROZEN. Do **not** start S4.

## Frozen Baseline

| Milestone | SHA |
|-----------|-----|
| S2 Frozen Tag `forma-s2-frozen` | `413c3bcc148dfe518b31d6267e1a0c72fc2f0645` |
| Post-Freeze main | `65b165dcd84d1b9f9c083e87d65bbf16f9061ad1` |
| S3 initial | `49821f80ad427ab51ea53b984cdbb5482a57be2c` |
| S3-G1 | `e58d97d6d2b2d2b25b25c9d0a46ce01b1903b292` |
| S3-G1-F1 | `f25634efd5f6f8cd334843435412bbec2b8ad169` |
| **S3-G1-F2** | **`f52e355c` — `fix(forma): close S3 G1 final backend gates`** |

## CI History

| Run | Result | Notes |
|-----|--------|-------|
| `33469903703` | **FAIL** | backend compile, atlas checksum, rush shrinkwrap, hook mode |
| `33471829970` | **FAIL** | `proposal.go` syntax `]}`; MySQL `CREATE UNIQUE INDEX IF NOT EXISTS`; backend tests blocked |
| `33474851052` | **FAIL** | **migration PASS, frontend PASS, backend FAIL** — see root cause below |
| `33475985604` | **PASS** | `f52e355c` — **migration PASS, frontend PASS, backend PASS** |
| `33476030584` | **PASS** | `6a11fa89` — docs-only RESULT update |

### Run `33474851052` — Backend root cause

1. Wrong `BusinessModelRevision` package on `ApplyProposal` interface (`entity` vs `businessentity`)
2. Unused `conflicts` variable in `persistExtraction`
3. `bool` appended to `IncludedItems` (`[]string`) in `analysis.go`

## S3-G1-F2 Closure (this round)

### STEP 1 — Backend compile fixes

- `ApplyProposal` returns `*businessentity.BusinessModelRevision` (no analyst-local alias)
- Removed unused `conflicts` var; extraction persist moved to `extraction_persist.go`
- Removed bool append to `IncludedItems`; optional `"rendered_context"` string when context non-empty

### STEP 3 — Idempotency before sequence reservation

- `createUserTurnWithEvidence`: `GetTurnByClientRequestID` **before** `next_turn_sequence += 2`
- Duplicate `client_request_id` returns existing turn + evidence without consuming sequence

### STEP 5 — Extraction persistence atomicity

- `persistExtraction` wraps `repo.Transaction`
- `persistExtractionWithRepo` uses `txRepo` for assertions, evidence refs, gaps, conflicts

### STEP 7–8 — Proposal STALE race + CAS

- Transaction内 stale → return `ErrProposalStale` (no in-tx status write → no rollback)
- Post-tx: `MarkProposalStaleIfReady` (CAS `READY_FOR_REVIEW` → `STALE`)
- `APPLIED` never overwritten by stale handler

### STEP 10 — Transaction test infrastructure

- `memAnalystRepo.txnMu` serializes snapshot transactions for concurrent tests

### STEP 11–13 — Contracts preserved

- Turn lineage: `next_turn_sequence` / `reserved_reply_sequence` / `reply_to_turn_id`
- Retry state machine: `EXTRACTION_FAILED` → full; `RESPONSE_FAILED` → generation only
- Production fail-closed: no `DeterministicFakeModel` in wiring
- Analyst turns: unique `client_request_id` (turn id) to satisfy session idempotency unique key

### Local compile gate (Docker Go 1.24)

```
go test ./domain/forma/... ./crossdomain/forma/... ./application/forma/... -count=1  → PASS
go test ./domain/forma/analyst/... -count=1                                         → PASS
```

### New / expanded tests

- `TestConcurrentSameClientRequestIdempotent` — 10 concurrent same CR → 1 turn, `next_turn_sequence = 3`
- `TestExtractionPersistenceRollback` — fail on 2nd assertion → 0 assertions; retry → single complete set
- `TestProposalStaleDetectedInTransaction` — in-tx stale, CAS mark, no r3
- `TestMarkProposalStaleDoesNotOverwriteApplied` — APPLIED not reverted to STALE

## Gate Evidence

| Gate | Status |
|------|--------|
| Backend compile | **PASS** (local) |
| Idempotency / sequence | **PASS** (tests) |
| Extraction atomicity / rollback | **PASS** (tests) |
| Stale race + CAS | **PASS** (tests) |
| Transaction concurrency | **PASS** (tests) |
| Retry state machine | **PASS** (tests) |
| Forma CI all green | **PASS** (`33475985604` on `f52e355c`) |
| G16 Live Model E2E | **BLOCKED** — requires real Coze/Eino model |
| G17 Browser E2E + `forma/cursor-results/s3-ui/` | **PENDING** |
| No silent mutation | **PENDING** (browser) |
| Provenance hard gate | **PENDING** (browser) |
| Tenant isolation | **PENDING** (browser) |

## Target final status

When **CI ALL GREEN + Live Model PASS + Browser PASS + all hard gates PASS**:

→ `PASS_WITH_REVIEW_GATE` (not FROZEN)

**DO NOT START S4.** No `forma-s3-frozen`.
