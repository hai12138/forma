# FORMA S3 AI BUSINESS ANALYST RESULT

## Status

**IN_PROGRESS — S3-G2 context + gap semantics implemented; Live Model / Browser gates pending runtime**

Do **not** mark FROZEN. Do **not** start S4.

## Frozen Baseline

| Milestone | SHA |
|-----------|-----|
| S2 Frozen Tag `forma-s2-frozen` | `413c3bcc148dfe518b31d6267e1a0c72fc2f0645` |
| S3 initial | `49821f80ad427ab51ea53b984cdbb5482a57be2c` |
| S3-G1 | `e58d97d6d2b2d2b25b25c9d0a46ce01b1903b292` |
| S3-G1-F1 | `f25634efd5f6f8cd334843435412bbec2b8ad169` |
| S3-G1-F2 | `f52e355c99c178c3ff4d4647620a5aef88046dfe` |
| **S3-G2** | **`7ff6e583` — `fix(forma): close S3 context and live acceptance gates`** |

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

## S3-G2 Context & Live Acceptance (this round)

### BLOCKER-01 — Context reaches model

- `ExtractionRequest.ContextText` / `InterviewTurnRequest.ContextText` added
- `buildContextInput` returns rendered business context (no system policy in text)
- `CozeEinoAnalystModel` formats `<context>` + `<current_user_turn>` prompts
- `EnrichedSystemPolicy` — context is UNTRUSTED BUSINESS DATA boundary

### BLOCKER-02 — Gap Ask semantics

- Migration `20250902020000_s3_g2_context_gap.sql`: `focus_gap_id` on session
- `POST .../sessions/:sessionId/gaps/:gapId/ask` → ANALYST turn (no USER evidence)
- `PlanNextQuestion` prioritizes focused gap
- Confirm assertion resolves focused gap (`OPEN` → `RESOLVED`)
- Frontend `Ask This` calls API (no longer prefills user message)

### Tests (local PASS)

- `TestMultiTurnContextTextReachesModel` — spy model receives prior turn context
- `TestBuildContextBudgetTruncation` — token budget + `context_truncated`
- `TestGapAskCreatesAnalystTurnWithoutEvidence`
- `TestGapResolvedOnConfirmWhenFocused`

### Browser / Live (pending)

- `scripts/forma/s3-browser-e2e.mjs` — Playwright gate (`FORMA_LIVE_E2E=1`)
- Live model E2E logs — requires configured Coze/Eino runtime

## Gate Evidence

| Gate | Status |
|------|--------|
| Context reaches model | **PASS** (unit/integration) |
| Gap Ask semantics | **PASS** (unit/integration + API) |
| Context budget | **PASS** (tests) |
| S3-G1 engineering gates | **PASS** (preserved) |
| Forma CI | **PASS** (`33480332921` on `7ff6e583`) |
| Live Model E2E | **PENDING** |
| Browser E2E + UI evidence | **PENDING** |
| PASS_WITH_REVIEW_GATE | **NOT YET** |

**DO NOT START S4.** No `forma-s3-frozen`.
