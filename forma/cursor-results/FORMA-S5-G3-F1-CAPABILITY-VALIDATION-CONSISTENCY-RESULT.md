# FORMA-S5-G3-F1 — Capability Validation Consistency
# RESULT

**Gate:** S5-G3-F1
**Date:** 2026-09-13
**Status:** **PENDING_CI** (local tests green; await Forma CI)

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE | `4c0687305ae6bf126e669aa07b15d6f21a0b9639` |
| S5-G3 implementation | `2aa3f0ee711f550dfe27db27bb47d6813b6e99ef` |
| Scope | Minimal Validate/Activate consistency fixes |
| PRODUCT_CODE_CHANGE | CAPABILITY_DOMAIN_ONLY |
| MIGRATION_CHANGE | **NONE** |
| REAL_MODEL_CALLS | 0 |
| HTTP / UI / Impact / S5-G4 | NONE |

---

## 2. Delivered

| Fix | Result |
|-----|--------|
| FAIL ValidationResult commits; `ErrValidationFailed` only after UoW | PASS |
| FAIL List/Get readable; revision stays DRAFT | PASS |
| FAIL replay / concurrent FAIL → same ValidationID | PASS |
| Aggregate-wide Cap-key union completeness; cross-binding duplicate FAIL | PASS |
| Multi-contract field split PASS | PASS |
| AI / DERIVED provenance tightened (target/source/proposal/capability/analysis) | PASS |
| `buildValidationEvidence` under lock for Validate + Activate | PASS |
| Descriptor / SortSchema / Classification drift fail-closed | PASS |
| FILTER: optional inputs still operator-checked | PASS |
| `issue_codes_json` unmarshal → `ErrConsistency` | PASS |
| `ContractEvidenceDigest` includes SortSchema + Classification | PASS |
| MemoryUoW `FailCreateValidation` inject rolls back | PASS |

Local verification:
- `go test ./domain/forma/capability/... -count=1` → PASS
- `go test ./domain/forma/... -count=1` → PASS
- `node scripts/forma/migration-validate.mjs` → 19/19 PASS
- `git diff --check` → PASS

---

## 3. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `f315792f4fb8b63d1ac475943ffc5f0b685a1263` |
| CI_RUN | PENDING |
| forma-backend | PENDING |
| forma-migration-apply | PENDING |
| forma-frontend | PENDING |
| CI | **PENDING_CI** |

---

## 4. Gate Summary

```text
S5_G3_F1_STATUS = PENDING_CI
FAIL_PERSISTENCE = PASS
AGGREGATE_MAPPING = PASS
PROVENANCE = PASS
EVIDENCE_UNDER_LOCK = PASS
FILTER_OPTIONAL_OPS = PASS
ISSUE_CODES_FAIL_CLOSED = PASS
DIGEST_SORT_CLASSIFICATION = PASS
FAIL_CREATE_VALIDATION_ROLLBACK = PASS
PRODUCT_CODE_CHANGE = CAPABILITY_DOMAIN_ONLY
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = PENDING_CI
S5_G4_READY = NO
```

**Stop:** **STOP**. Do not start S5-G4. Await CI + human review.
