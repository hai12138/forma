# FORMA-S5-G2-F1 — Capability Domain Architecture Consistency Fix
# RESULT

**Gate:** S5-G2-F1
**Date:** 2026-09-13
**Status:** PENDING_CI (implementation ready; awaiting Forma CI ALL GREEN + human review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `771afa8d324f57324772ecbd06d0fc1d6343b1ae` |
| Scope | Capability domain consistency fixes only |
| PRODUCT_CODE_CHANGE | CAPABILITY_DOMAIN_ONLY |
| MIGRATION_CHANGE | S5_G2_F1_ONLY |
| REAL_MODEL_CALLS | 0 |
| HTTP / router / frontend | NONE |
| forma-s5-frozen | NOT CREATED |
| S5-G3 | NOT STARTED |

---

## 2. Fixes

| Area | Result |
|------|--------|
| Shared UoW fail-closed | PASS — removed non-atomic fallback; `ErrUoWNotConfigured`; non-memory assets require `DB` |
| AssetProjection fault rollback | PASS — `FailingAssetProjection` + full Cap/Rev/Decision/Asset rollback |
| Activate concurrency | PASS — `aggregate_generation` CAS; concurrent Activate exactly-one |
| Lifecycle evidence | PASS — Validate/MarkStale always fail closed until G3 evidence |
| Lifecycle audit | PASS — Activate/Deprecate write immutable Decision with actor/reason |
| Proposal target binding | PASS — bound CapabilityID mismatch → conflict, no writes |
| Materialization validation | PASS — unified `ValidateMaterializationPayload` |
| Secret isolation | PASS — Options map removed; secret patterns rejected; generator errors sanitized |
| Analysis error handling | PASS — no swallowed List/Mark/Get errors; no nil Run+nil err |
| JSON fail-closed | PASS — corrupt payload JSON → `ErrConsistency` |
| Memory lease isolation | PASS — deep-copy `ExecutionClaimedAt` / `LeaseExpiresAt` |

---

## 3. Local verification

| Check | Result |
|-------|--------|
| `go test ./domain/forma/capability/... -count=1` | PASS |
| `go test ./domain/forma/... -count=1` | PASS |
| `node scripts/forma/migration-validate.mjs` | PASS (16/16) |
| `git diff --check` | _pending commit_ |
| migration CASE A/B/C | Deferred to Forma CI `forma-migration-apply` |

---

## 4. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `e8f6f38846f33f81dd3f4816508689f49ec8ff17` |
| CI_RUN | _pending_ |
| forma-backend | _pending_ |
| forma-migration-apply | _pending_ |
| forma-frontend | _pending_ |
| CI | _pending_ |

---

## 5. Gate Summary (pre-CI)

```text
S5_G2_F1_STATUS = PENDING_CI
UOW_ATOMICITY = PASS
ACTIVATE_CONCURRENCY = PASS
LIFECYCLE_EVIDENCE = PASS
LIFECYCLE_AUDIT = PASS
PROPOSAL_TARGET_BINDING = PASS
MATERIALIZATION_VALIDATION = PASS
SECRET_ISOLATION = PASS
ANALYSIS_ERROR_HANDLING = PASS
JSON_FAIL_CLOSED = PASS
MEMORY_LEASE_ISOLATION = PASS
PRODUCT_CODE_CHANGE = CAPABILITY_DOMAIN_ONLY
MIGRATION_CHANGE = S5_G2_F1_ONLY
REAL_MODEL_CALLS = 0
CI = pending
S5_G3_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §4/§5, then **STOP**. Do not start S5-G3. Do not create `forma-s5-frozen`.
