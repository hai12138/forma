# FORMA-S5-G1-F3 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F3
**Date:** 2026-09-11
**Status:** PENDING_CI (docs committed; awaiting Forma CI ALL GREEN + human architecture review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `8fcbb9b7e045cfdabc8b36f615aeaa1e739da46c` |
| S4 freeze | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| Scope | Docs only |
| PRODUCT_CODE_CHANGE | NONE |
| MIGRATION_CHANGE | NONE |
| REAL_MODEL_CALLS | 0 |

---

## 2. Deliverables

| Artifact | Action |
|----------|--------|
| `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md` | AMENDED (G1-F3) |
| `forma/cursor-results/FORMA-S5-G1-F3-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED |

---

## 3. Fixes Applied

| Item | Result |
|------|--------|
| ConfirmProposal replay | PASS — lock first; PROPOSED materializes; matching terminal replays revision; action mismatch → conflict; missing revision/decision → consistency error; no duplicates |
| DeriveRevision atomic | PASS — lock Capability+source; atomic DRAFT+Decision; rollback on failure; idempotent key replay; no concurrent duplicate derive |
| RejectProposal atomic | PASS — lock Proposal; atomic REJECT Decision+REJECTED; replay returns Decision; Confirm terminal → conflict; no REJECTED without Decision |
| AssetRef alignment | PASS — no description projection; exact Name/SemanticVersion/Revision/ContentDigest sources; frozen status map; RELEASED requires active_revision_id |
| QUERY input compatibility | PASS — caller required/optional only; bind existing logical fields; FilterSchema subset; LOOKUP deferred (no LookupSchema in S4 descriptor) |
| AnalysisRun retry | PASS — same analysis_run_id only; FAILED→PENDING + attempt; ordinary replay returns current status; SUCCEEDED unique Proposal set; linked retry row removed |

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
S5_G1_F3_STATUS = PENDING_CI
CONFIRM_REPLAY = PASS
DERIVE_ATOMICITY = PASS
REJECT_ATOMICITY = PASS
ASSETREF_ALIGNMENT = PASS
QUERY_INPUT_COMPAT = PASS
ANALYSIS_RETRY = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = pending
S5_G2_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §4/§5, then **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`.
