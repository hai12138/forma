# FORMA-S5-G4-F1 — Capability API Identity / Retry / Secret / Handler Consistency
# RESULT

**Gate:** S5-G4-F1  
**Date:** 2026-09-14  
**Status:** **PASS (local)** — Forma CI pending / updating after green

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main) | `dc1a90b887d5d1dac74e77f8b1aa36677ac6e203` |
| G4 implementation | `5c3d257ec88e0d0f6dd8467824cb511322449bd7` |
| Freeze tip | `0bdb40ce198cdc6bd7d2a12490279d6da7086543` |
| PRODUCT_CODE_CHANGE | **YES** |
| MIGRATION_CHANGE | **NONE** |
| REAL_MODEL_CALLS | **0** |
| G5 / forma-s5-frozen | **NONE** |

---

## 2. Parallel candidates (A→B→C)

| Agent | Branch | Candidate SHA |
|-------|--------|---------------|
| A identity/secret | `forma/s5-g4-f1-agent-a-identity` | `198650a98b059f5cc4a8e333ddae82c505d780c4` |
| B FAILED/retry | `forma/s5-g4-f1-agent-b-retry` | `4b37dca8989302d84d96b4f3b1864f23476c7522` |
| C handler/docs | `forma/s5-g4-f1-agent-c-handler` | `fe36614707aed4f1dcbac02c84fcc132881b7539` |

Freeze: `forma/cursor-results/FORMA-S5-G4-F1-INTERFACE-FREEZE.md`

---

## 3. Integrator deliverables

| Fix | Result |
|-----|--------|
| PrincipalID vs Asset OwnerID separation | PASS |
| Explicit OwnerID on ManualCreate / Confirm / EditConfirm | PASS |
| Audit metadata reject-before-write (`ValidateAuditMetadata`) | PASS |
| FAILED first-attempt returns same run DTO (HTTP 200 + status FAILED) | PASS |
| Explicit Retry API + attempt increment + concurrent claim | PASS |
| `ErrAnalysisFailed` → `FORMA_CAPABILITY_ANALYSIS_FAILED` (not ValidationFailed) | PASS |
| Malformed JSON fail-closed on Confirm/Reject/Activate/Deprecate | PASS |
| Real handler envelope tests | PASS |
| Tenancy request audit best-effort (no raw err text) | PASS |
| App file split: `capability_dto.go` / `capability_auth.go` / `capability_app.go` / `capability_analysis_app.go` | PASS |

---

## 4. Implementation commit

| Field | Value |
|-------|-------|
| COMMIT_SHA (integrator product) | 14ab0af8f23d7b8974fce36cce12beaa35762e4a |
| Merges | `0dd1fdab` (A), `9fad0c30` (B), `edcb59a5` (C) + integrator identity/split |

---

## 5. Gate checklist

```text
S5_G4_F1_STATUS = PASS_LOCAL
PRINCIPAL_IDENTITY = PASS
OWNER_ID_SEPARATION = PASS
FAILED_RUN_VISIBILITY = PASS
EXPLICIT_RETRY_API = PASS
RETRY_CONCURRENCY = PASS
AUDIT_METADATA_SECRET_ISOLATION = PASS
MALFORMED_JSON_FAIL_CLOSED = PASS
REAL_HANDLER_TESTS = PASS
PRODUCT_CODE_CHANGE = YES
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
COMMIT_SHA = 14ab0af8f23d7b8974fce36cce12beaa35762e4a
CI_RUN = pending
CI = pending
S5_G5_READY = NO
```

---

## 6. Local verification

| Gate | Result |
|------|--------|
| Principal / OwnerID / secret / no-partial-write domain tests | PASS |
| Analysis FAILED / retry / concurrency app tests | PASS |
| Handler malformed JSON + envelope tests | PASS |
| `go test ./domain/forma/... ./application/forma/... ./api/handler/forma/... ./crossdomain/forma/...` | PASS |
| `migration-validate.mjs` | 19/19 PASS |
| CASE A/B/C apply | deferred to CI (local Docker daemon unavailable historically) |
| `typecheck.mjs` + `routes-smoke.mjs` | PASS |
| Rush full FE build | deferred to CI (Windows toolchain) |
| `git diff --check` | PASS |

---

## 7. CI

| Field | Value |
|-------|-------|
| CI_RUN | _pending after push_ |
| forma-backend | _pending_ |
| forma-migration-apply | _pending_ |
| forma-frontend | _pending_ |
| CI | _pending_ |

**Stop:** Do not create `forma-s5-frozen`. Do not start S5-G5. Await human review after CI ALL GREEN.
