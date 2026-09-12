# FORMA-S5-G2-F1 — Capability Domain Architecture Consistency Fix
# RESULT

**Gate:** S5-G2-F1
**Date:** 2026-09-13
**Status:** **PASS** (Capability consistency fixes; Forma CI ALL GREEN; await human review before S5-G3)

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

`git diff --check`: PASS

---

## 3. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `e8f6f38846f33f81dd3f4816508689f49ec8ff17` |
| CI_RUN | `34705402311` (tip `077c3f30…`) |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |
| CI | **ALL GREEN** |
| CI URL | https://github.com/hai12138/forma/actions/runs/34705402311 |

---

## 4. Gate Summary

```text
S5_G2_F1_STATUS = PASS
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
COMMIT_SHA = e8f6f38846f33f81dd3f4816508689f49ec8ff17
CI_RUN = 34705402311
CI = ALL GREEN
S5_G3_READY = NO
```

**Stop:** **STOP**. Do not start S5-G3. Do not create `forma-s5-frozen`. Await human review.
