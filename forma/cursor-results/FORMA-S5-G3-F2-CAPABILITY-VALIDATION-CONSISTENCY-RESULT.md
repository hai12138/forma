# FORMA-S5-G3-F2 — Capability Validation Consistency (Current BM / Strict Provenance / Evidence Fence)
# RESULT

**Gate:** S5-G3-F2
**Date:** 2026-09-13
**Status:** **PENDING_CI** (implementation committed; awaiting Forma CI)

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main) | `4e5b5d4e8d269bb2c164be9de679edf87c85dfaa` |
| S5-G3-F1 implementation | `f315792f4fb8b63d1ac475943ffc5f0b685a1263` |
| Scope | Minimal Validate/Activate consistency fixes |
| PRODUCT_CODE_CHANGE | CAPABILITY_DOMAIN + validation adapter |
| MIGRATION_CHANGE | **NONE** |
| REAL_MODEL_CALLS | 0 |
| HTTP / UI / Impact / S5-G4 | NONE |

---

## 2. Delivered

| Fix | Result |
|-----|--------|
| BM adapter reads Business master; requires `CurrentRevision == requested`; then loads revision | PASS |
| Tenant/business/current mismatch fail-closed | PASS |
| BM CurrentRevision advance → old pin Validate FAIL (`IssueBMNotFound`) | PASS |
| Validate PASS then BM advance → Activate `ErrMissingValidationEvidence` | PASS |
| AI Proposal: same tenant/business; terminal status; non-empty Cap/MatRev/AnalysisRun strict eq | PASS |
| AI Decision CONFIRM/EDIT_CONFIRM: tenant/business/proposal/capability/target strict eq (no empty skip) | PASS |
| AnalysisRun exists; tenant/business/BM revision match; status `SUCCEEDED` | PASS |
| DERIVED: Decision strict; load `derived_from_revision_id`; source same tenant/business/capability | PASS |
| Final evidence fence helper shared by Validate + Activate before writes | PASS |
| Drift between 1st/2nd evidence read aborts txn; no bad evidence; state unchanged | PASS |
| Historical non-current BM pin rejected (G3 regression updated) | PASS |

Local verification:
- `go test ./domain/forma/capability/... -count=1` → PASS
- `go test ./domain/forma/... -count=1` → PASS
- `node coze-studio/scripts/forma/migration-validate.mjs` → 19/19 PASS
- `git diff --check` → PASS

---

## 3. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `74db871d4a944a2bcd0957374c5d39a91381b319` |
| CI_RUN | _(pending push)_ |
| forma-backend | _(pending)_ |
| forma-migration-apply | _(pending)_ |
| forma-frontend | _(pending)_ |
| CI | PENDING |

---

## 4. Gate Summary

```text
S5_G3_F2_STATUS = PENDING_CI
BM_CURRENT_ONLY = PASS
AI_PROVENANCE_STRICT = PASS
DERIVED_SOURCE_REVISION = PASS
FINAL_EVIDENCE_FENCE = PASS
PRODUCT_CODE_CHANGE = CAPABILITY_DOMAIN_PLUS_ADAPTER
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = PENDING
S5_G4_READY = NO
```

**Stop:** Do not start S5-G4. Await Forma CI + human review.
