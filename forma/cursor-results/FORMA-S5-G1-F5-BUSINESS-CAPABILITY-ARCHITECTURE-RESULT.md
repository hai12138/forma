# FORMA-S5-G1-F5 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F5
**Date:** 2026-09-11
**Status:** PENDING_CI (docs committed; awaiting Forma CI ALL GREEN + human architecture review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `d69db33503a697f70d4e3e665646596cccfef1d6` |
| S4 freeze | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| Scope | Docs only |
| PRODUCT_CODE_CHANGE | NONE |
| MIGRATION_CHANGE | NONE |
| REAL_MODEL_CALLS | 0 |

---

## 2. Deliverables

| Artifact | Action |
|----------|--------|
| `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md` | AMENDED (G1-F5) |
| `forma/cursor-results/FORMA-S5-G1-F5-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED |

---

## 3. Fixes Applied

| Item | Result |
|------|--------|
| ProjectCapabilityAssetRef | PASS — pure aggregate projection with deterministic ACTIVE→VALIDATED→DRAFT→STALE→DEPRECATED priority; unique version; recompute in same txn after Confirm/Derive/Validate/Activate/STALE/Deprecate |
| Existing Capability Confirm/Derive | PASS — AssetRef created only on first create; new DRAFT uses aggregate projection; ACTIVE RELEASED Name/SemVer/ContentDigest protected |
| Decision payload_digest | PASS — single field for CONFIRM/EDIT_CONFIRM/DERIVE; field table + Confirm + Derive aligned |
| Confirm idempotency | PASS — logical key = tenant_id + proposal_id; client_request_id tracing only |
| QUERY V1 | PASS — AGGREGATE deferred with LOOKUP; cardinality table READ/LIST/FILTER only; FILTER requires required input + required predicate |
| AssetRef.SchemaVersion | PASS — fixed "1.0" at create; unchanged in V1 lifecycle |

---

## 4. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | _pending push_ |
| CI_RUN | _pending_ |
| forma-backend | _pending_ |
| forma-migration-apply | _pending_ |
| forma-frontend | _pending_ |
| CI | _pending_ |

---

## 5. Gate Summary (pre-CI)

```text
S5_G1_F5_STATUS = PENDING_CI
AGGREGATE_PROJECTION = PASS
EXISTING_CAPABILITY_DRAFT = PASS
PAYLOAD_DIGEST = PASS
CONFIRM_IDEMPOTENCY = PASS
QUERY_V1_SCOPE = PASS
SCHEMA_VERSION = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = pending
S5_G2_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §4/§5, then **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`.
