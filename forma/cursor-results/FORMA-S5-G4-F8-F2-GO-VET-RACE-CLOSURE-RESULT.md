# FORMA-S5-G4-F8-F2 — Go Vet Copylocks + Race CI Closure
# RESULT

**Gate:** S5-G4-F8-F2
**Date:** 2026-09-15
**Status:** **PASS / CI ALL GREEN** (implementation SHA paired)

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main) | `1081a2f509203a238abc4125332fc0022bc147e6` |
| Freeze tip | `008488a72faf2e546e2895042e83806022edca35` |
| PRODUCT_CODE_CHANGE | **NONE** |
| TEST_CODE_CHANGE | **MINIMAL** (delete dead mutex-copy lines only) |
| CI_WORKFLOW_CHANGE | **ORDER_ONLY** (pre-commit before vet/race; commands unchanged) |
| MIGRATION_CHANGE | **NONE** |
| FRONTEND_CHANGE | **NONE** |
| RUNTIME_CHANGE | **NONE** |
| REAL_MODEL_CALLS | **0** |
| G5_CHANGE | **NONE** |

---

## 2. Changes

### Vet fix

In `validate_g3_test.go`, deleted only:

```go
bad := *contracts
_ = bad
```

`FakeContractPort`, mutex fields, production code, and remaining test assertions unchanged.

### CI order

`Forma pre-commit CWD check` now runs **before**:

- `Forma capability static analysis` (`go vet …`)
- `Forma capability race tests` (`go test -race …`)

Step `run:` bodies unchanged from F8-F1.

---

## 3. Implementation CI (paired)

| Field | Value |
|-------|-------|
| IMPLEMENTATION_SHA | `f073b24dd7bb967bb0cf23c06bfe4af937c55be4` |
| IMPLEMENTATION_CI_RUN | https://github.com/hai12138/forma/actions/runs/34937595390 |

| Step | Result |
|------|--------|
| Forma domain tests (+ other backend tests) | success |
| Forma pre-commit CWD check | success |
| Forma capability static analysis (`go vet`) | success |
| Forma capability race tests (`go test -race`) | success (executed, not skipped) |
| forma-backend | success |
| forma-migration-apply | success |
| forma-frontend | success |

---

## 4. Preserved prior evidence

| Item | Value |
|------|-------|
| F8 product fail-closed | PASS (`27059ef3` / CI `34912401820`) |
| F8 durable RED | `aa36bbab` / CI `34911998223` Forma domain tests FAILURE |
| F8-F1 BLOCKED record | preserved in F8-F1 RESULT (not deleted) |

---

## 5. Gate checklist

```text
S5_G4_F8_F2_STATUS = PASS
COPYLOCK_VET_FIX = PASS
TEST_SEMANTICS_UNCHANGED = PASS
PRECOMMIT_CHECK_EXECUTED = PASS
CI_GO_VET = PASS
CI_RACE = PASS
RED_EVIDENCE_PRESERVED = PASS
F8_PRODUCT_CODE_STATUS = PASS
PRODUCT_CODE_CHANGE = NONE
TEST_CODE_CHANGE = MINIMAL
CI_WORKFLOW_CHANGE = ORDER_ONLY
MIGRATION_CHANGE = NONE
FRONTEND_CHANGE = NONE
RUNTIME_CHANGE = NONE
REAL_MODEL_CALLS = 0
IMPLEMENTATION_SHA = f073b24dd7bb967bb0cf23c06bfe4af937c55be4
IMPLEMENTATION_CI_RUN = https://github.com/hai12138/forma/actions/runs/34937595390
S5_G5_READY = NO
```

RESULT tip SHA / tip CI are reported separately in the Cursor final return (not chased inside this document).

**Stop:** No `forma-s5-frozen`. No S5-G5.
