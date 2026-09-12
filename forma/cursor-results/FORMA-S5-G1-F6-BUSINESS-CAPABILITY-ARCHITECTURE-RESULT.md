# FORMA-S5-G1-F6 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F6
**Date:** 2026-09-12
**Status:** PENDING_CI (docs committed; awaiting Forma CI ALL GREEN + human architecture review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `bfd1137469df9c71ad9e1030c430efac7c602531` |
| S4 freeze | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| Scope | Docs only |
| PRODUCT_CODE_CHANGE | NONE |
| MIGRATION_CHANGE | NONE |
| REAL_MODEL_CALLS | 0 |

---

## 2. Deliverables

| Artifact | Action |
|----------|--------|
| `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md` | AMENDED (G1-F6) |
| `forma/cursor-results/FORMA-S5-G1-F6-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED |

---

## 3. Fixes Applied

| Item | Result |
|------|--------|
| ProjectCapabilityAssetRef completeness | PASS — dangling/cross/non-ACTIVE pointer → consistency error (no AssetRef write); STALE may coexist with DEPRECATED; empty revisions / duplicate version → consistency error |
| Capability aggregate concurrency | PASS — Proposal lock then Capability lock; version allocate + project under lock; UNIQUE (tenant_id, capability_id, version); all projecting lifecycle ops same lock+txn |
| ContentDigest definition | PASS — explicit include/exclude semantic fields; G2 canonicalization; identical payload → identical digest |
| Static verification matrix | PASS — six required cases documented in §3.4.5 |

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
S5_G1_F6_STATUS = PENDING_CI
PROJECTION_TOTALITY = PASS
AGGREGATE_LOCK = PASS
CONTENT_DIGEST = PASS
STATIC_MATRIX = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = pending
S5_G2_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §4/§5, then **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`.
