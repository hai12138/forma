# FORMA-S5-G1-F2 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F2
**Date:** 2026-09-11
**Status:** PENDING_CI (docs committed; awaiting Forma CI ALL GREEN + human architecture review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `392530956eee40ed5fda74d4027d48f5a99fabf9` |
| S4 freeze | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| Scope | Docs only |
| PRODUCT_CODE_CHANGE | NONE |
| MIGRATION_CHANGE | NONE |
| REAL_MODEL_CALLS | 0 |

---

## 2. Deliverables

| Artifact | Action |
|----------|--------|
| `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md` | AMENDED (G1-F2) |
| `forma/cursor-results/FORMA-S5-G1-F2-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED |

---

## 3. Fixes Applied

| Item | Result |
|------|--------|
| DERIVED_EDIT lifecycle | PASS — single edit model via OWNER/ADMIN EDIT/DERIVE Decision with source+target revision ids; added to §9.6 Validate prerequisites; Proposal+EDIT_CONFIRM reserved for AI path only |
| Unbound Proposal REJECT audit | PASS — `capability_id` null only on unbound REJECT; `proposal_id` required then; CONFIRM/EDIT_CONFIRM must bind `capability_id == asset_id`; no shell Asset/Capability on REJECT |
| Field SoT | PASS — aggregate = identity/ownership/active pointer; Revision = versioned semantics SoT; Asset header = projection; Activate atomic projection + Asset status ≠ Revision status |
| ConfirmProposal atomic boundary | PASS — lock Proposal; create/bind Asset+Capability; Revision+Decision; terminal Proposal update; no partial state; idempotent concurrency/retry |
| CapabilityAnalysisRun | PASS — PENDING/SUCCEEDED/FAILED; same key+digest return matrix; FAILED explicit retry + attempt/audit; ≤1 concurrent model executor |
| V1 QUERY compatibility | PASS — required/optional input binding direction; cardinality enum ONE/MANY/AGGREGATE mapped to READ/LOOKUP/LIST/FILTER/AGGREGATE; incompat → Validation FAIL |

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
S5_G1_F2_STATUS = PENDING_CI
DERIVED_EDIT_LIFECYCLE = PASS
REJECT_AUDIT = PASS
FIELD_SOT = PASS
CONFIRM_ATOMICITY = PASS
ANALYSIS_RUN_LIFECYCLE = PASS
QUERY_COMPATIBILITY = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = pending
S5_G2_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §4/§5, then **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`.
