# FORMA-S5-G1-F3 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F3
**Date:** 2026-09-11
**Status:** **PASS** (docs-only; Forma CI ALL GREEN; await human architecture review before S5-G2)

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
| `forma/cursor-results/FORMA-S5-G1-F3-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED / FINALIZED |

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

`git diff --check`: PASS (clean)

---

## 4. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `385150d4dea8704f9d1f36c64b587830e0b89b1e` |
| CI_RUN | `34556254534` |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |
| CI | **ALL GREEN** |
| CI URL | https://github.com/hai12138/forma/actions/runs/34556254534 |

---

## 5. Gate Summary

```text
S5_G1_F3_STATUS = PASS
CONFIRM_REPLAY = PASS
DERIVE_ATOMICITY = PASS
REJECT_ATOMICITY = PASS
ASSETREF_ALIGNMENT = PASS
QUERY_INPUT_COMPAT = PASS
ANALYSIS_RETRY = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = ALL GREEN
S5_G2_READY = NO
```

**Stop:** **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`. Await human architecture review.
