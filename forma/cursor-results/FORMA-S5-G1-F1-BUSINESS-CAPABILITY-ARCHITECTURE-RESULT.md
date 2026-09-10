# FORMA-S5-G1-F1 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F1
**Date:** 2026-09-10
**Status:** PENDING_CI (docs committed; awaiting Forma CI ALL GREEN + human architecture review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `ceb24fbbfa888f3c6227d724836babe0711aa0dd` |
| Original S5-G1 architecture commit | `82ccd90f6c4e1d12987e2d9e219a4174dbaf299c` |
| S4 freeze | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| Scope | Docs only |
| PRODUCT_CODE_CHANGE | NONE |
| MIGRATION_CHANGE | NONE |
| REAL_MODEL_CALLS | 0 |

---

## 2. Deliverables

| Artifact | Action |
|----------|--------|
| `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md` | AMENDED (G1-F1 consistency locks) |
| `forma/cursor-results/FORMA-S5-G1-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | DATE fixed to `2026-09-10` |
| `forma/cursor-results/FORMA-S5-G1-F1-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED |

---

## 3. Consistency Fixes Applied

| Lock | Result |
|------|--------|
| PROPOSAL_REVISION_SEAM | PASS — AnalysisRun → Proposal → CONFIRM/EDIT_CONFIRM → DRAFT Revision; REJECT creates no Revision |
| HUMAN_CONFIRM_INVARIANT | PASS — Validate requires Confirm path or MANUAL_CREATED + CREATE decision; Validate/Activate cannot substitute Confirm |
| REVISION_IMMUTABILITY | PASS — all persisted revisions including DRAFT are semantically immutable |
| ACTIVE_POINTER_POLICY | PASS — single-transaction Activate; pointer clear only when matching; concurrent Activate ≤1 success |
| ANALYSIS_IDEMPOTENCY | PASS — key + `request_digest`; same/different digest rules; no duplicate Proposal/Revision on retry |
| SCHEMA_COMPATIBILITY | PASS — V1 QUERY rules locked (S4 types, no implicit cast, subset filter/sort, cardinality, FAIL on weaken) |
| CAPABILITY_ASSET_IDENTITY | PASS — `capability_id == asset_id`; no duplicate aggregate status; transactional create / compensation required in G2 |
| ROLE_MATRIX | PASS — exact OWNER/ADMIN/MEMBER/VIEWER; Activate/Deprecate = OWNER; SUPER_ADMIN not automatic tenant data access |

### Removed / corrected language

- Removed “AI proposal materialized directly as DRAFT”
- Removed “or remains history per activate policy”
- Removed fuzzy `MEMBER+` / `ADMIN+`
- Soft `SHOULD align` identity → locked `capability_id == asset_id`

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
S5_G1_F1_STATUS = PENDING_CI
PROPOSAL_REVISION_SEAM = PASS
HUMAN_CONFIRM_INVARIANT = PASS
REVISION_IMMUTABILITY = PASS
ACTIVE_POINTER_POLICY = PASS
ANALYSIS_IDEMPOTENCY = PASS
SCHEMA_COMPATIBILITY = PASS
CAPABILITY_ASSET_IDENTITY = PASS
ROLE_MATRIX = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = pending
S5_G2_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §4/§5, then **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`.
