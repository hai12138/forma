# FORMA-S5-G1-F4 — Business Capability Architecture Consistency Fix
# RESULT

**Gate:** S5-G1-F4
**Date:** 2026-09-11
**Status:** PENDING_CI (docs committed; awaiting Forma CI ALL GREEN + human architecture review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `73735f410a9a3fc7d11052c904952dd9ea6f8716` |
| S4 freeze | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| Scope | Docs only |
| PRODUCT_CODE_CHANGE | NONE |
| MIGRATION_CHANGE | NONE |
| REAL_MODEL_CALLS | 0 |

---

## 2. Deliverables

| Artifact | Action |
|----------|--------|
| `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md` | AMENDED (G1-F4) |
| `forma/cursor-results/FORMA-S5-G1-F4-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md` | CREATED |

---

## 3. Fixes Applied

| Item | Result |
|------|--------|
| ConfirmProposal / EDIT_CONFIRM | PASS — immutable Proposal.payload; CONFIRM uses Proposal.payload; EDIT_CONFIRM full effective_payload; effective_payload_digest on Decision; key≠digest; same key+action+digest replays; digest mismatch → conflict |
| DeriveRevision idempotency | PASS — fixed logical key; independent request_digest; same/different digest rules; Decision recovers target_revision_id |
| AssetRef.Revision | PASS — single header row; Revision=1 forever; Capability version only on BusinessCapabilityRevision; CozeResourceRef stays on Revision=1; no Activate counter / duplicate headers |
| SemanticVersion | PASS — SemVer `0.N.0` display projection only |
| AssetRef status + transactions | PASS — DRAFT/VERIFIED/RELEASED/DEPRECATED map complete; Validate→VERIFIED same txn; all pointer/projection updates atomic |
| QUERY | PASS — removed non-computable underlying-parameter wording; FILTER ≥1 required input + FilterSchema; READ/LIST/AGGREGATE rules frozen; LOOKUP deferred |

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
S5_G1_F4_STATUS = PENDING_CI
CONFIRM_EDIT_PAYLOAD = PASS
DERIVE_IDEMPOTENCY = PASS
ASSETREF_REVISION = PASS
SEMVER_PROJECTION = PASS
ASSETREF_STATUS_TXN = PASS
QUERY_OP_RULES = PASS
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = pending
S5_G2_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §4/§5, then **STOP**. Do not start S5-G2. Do not create `forma-s5-frozen`.
