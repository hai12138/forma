# FORMA-S5-G1-F5 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F5
**Date:** 2026-09-11
**Status:** **PASS** (docs-only; Forma CI ALL GREEN; await human architecture review before S5-G2)

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
| `forma/cursor-results/FORMA-S5-G1-F5-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED / FINALIZED |

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

`git diff --check`: PASS (clean)

---

## 4. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `49e8dcedd61a25ea3885423a68cde87339585675` |
| CI_RUN | `34588551428` |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |
| CI | **ALL GREEN** |
| CI URL | https://github.com/hai12138/forma/actions/runs/34588551428 |

---

## 5. Gate Summary

```text
S5_G1_F5_STATUS = PASS
AGGREGATE_PROJECTION = PASS
EXISTING_CAPABILITY_DRAFT = PASS
PAYLOAD_DIGEST = PASS
CONFIRM_IDEMPOTENCY = PASS
QUERY_V1_SCOPE = PASS
SCHEMA_VERSION = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = ALL GREEN
S5_G2_READY = NO
```

**Stop:** **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`. Await human architecture review.
