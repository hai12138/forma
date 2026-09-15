# FORMA-S5-G4-F8-F1 — Evidence Correction (CI vet/race + durable RED)
# RESULT

**Gate:** S5-G4-F8-F1
**Date:** 2026-09-15
**Status:** **BLOCKED_BY_EXISTING_GO_VET** — stop; no product-code expansion

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main) | `0e2ed41c5e88a897f882bda71831979bcbf54852` |
| Scope | CI workflow + F8 RESULT evidence correction + this RESULT + issue-tracker |
| PRODUCT_CODE_CHANGE | **NONE** |
| MIGRATION_CHANGE | **NONE** |
| FRONTEND_CHANGE | **NONE** |
| RUNTIME_CHANGE | **NONE** |
| REAL_MODEL_CALLS | **0** |
| G5_CHANGE | **NONE** |

Allowed paths only:

```text
.github/workflows/forma-ci.yml
forma/cursor-results/FORMA-S5-G4-F8-READ-SEAM-FAIL-CLOSED-RESULT.md
forma/cursor-results/FORMA-S5-G4-F8-F1-EVIDENCE-CORRECTION-RESULT.md
docs/agents/issue-tracker.md
```

---

## 2. Why F8-F1

F8 product fail-closed outcome stands. Audit gaps:

1. Formal RED relied on non-durable `.git/F8_RED.txt` instead of Agent A CI.
2. Local Go was unavailable; RESULT incorrectly claimed local backend / vet PASS and race deferred.
3. Forma CI lacked persistent `go vet` / `-race` steps for capability packages.

---

## 3. Durable RED (unchanged product; corrected citation)

```text
RED_CI_RUN = https://github.com/hai12138/forma/actions/runs/34911998223
RED_SHA = aa36bbab08c1ad290a07db80f65bde1c74457d39
RED_BACKEND_STEP = Forma domain tests / FAILURE
RED_TREE_MATCHES_INTEGRATED_TEST_COMMIT = PASS
```

```text
.git/F8_RED.txt = NON_DURABLE / NOT USED AS FORMAL EVIDENCE
LOCAL_GO_AVAILABLE = NO
LOCAL_BACKEND_TESTS = NOT_RUN
LOCAL_GO_VET = NOT_RUN
LOCAL_RACE = NOT_RUN
```

F8 GREEN (unchanged):

```text
IMPLEMENTATION_GREEN_CI = https://github.com/hai12138/forma/actions/runs/34912401820
RESULT_TIP_GREEN_CI = https://github.com/hai12138/forma/actions/runs/34912997302
```

---

## 4. CI workflow change

In `forma-backend` job, after existing Forma tests, added (Go 1.24 via `actions/setup-go@v5`):

```yaml
- name: Forma capability static analysis
  run: go vet ./application/forma/... ./domain/forma/capability/...

- name: Forma capability race tests
  run: go test -race ./application/forma/... ./domain/forma/capability/... -count=1
```

No local Go install/path guessing. Per gate rule: if vet/race expose pre-existing defects → **STOP** (no skip / no silent product expansion).

---

## 5. Tip CI evidence — STOP

| Field | Value |
|-------|-------|
| RESULT_TIP_SHA (workflow+docs) | `437fdab6e816fa4ceedbdf4c18ca952e50809cc0` |
| RESULT_TIP_CI_RUN | https://github.com/hai12138/forma/actions/runs/34925437921 |
| CI_GO_VET (Forma capability static analysis) | **FAIL** |
| CI_RACE (Forma capability race tests) | **NOT_RUN** (skipped after vet failure) |
| forma-backend | **failure** |
| forma-migration-apply | success |
| forma-frontend | success |

### go vet findings (pre-existing; outside F8-F1 allowlist)

Annotations from run `34925437921` / job forma-backend:

```text
assignment copies lock value to _: github.com/coze-dev/coze-studio/backend/domain/forma/capability/service.FakeContractPort contains sync.Mutex
assignment copies lock value to bad: github.com/coze-dev/coze-studio/backend/domain/forma/capability/service.FakeContractPort contains sync.Mutex
```

This is **existing** capability domain test helper code (`FakeContractPort` + `sync.Mutex` copy). Fixing it requires editing production/test Go under `domain/forma/capability/…`, which is **outside** the F8-F1 file allowlist. Gate forbids skip, mask, or silent scope expansion → **main integrator STOP**.

Race step did not execute after vet failure.

---

## 6. Gate checklist

```text
S5_G4_F8_F1_STATUS = BLOCKED_BY_EXISTING_GO_VET
RED_CI_EVIDENCE = PASS
RED_SHA = aa36bbab08c1ad290a07db80f65bde1c74457d39
RED_CI_RUN = https://github.com/hai12138/forma/actions/runs/34911998223
IMPLEMENTATION_GREEN_CI = https://github.com/hai12138/forma/actions/runs/34912401820
LOCAL_GO_AVAILABLE = NO
LOCAL_CLAIMS_CORRECTED = PASS
CI_GO_VET = FAIL
CI_RACE = NOT_RUN
PRODUCT_CODE_CHANGE = NONE
MIGRATION_CHANGE = NONE
FRONTEND_CHANGE = NONE
RUNTIME_CHANGE = NONE
REAL_MODEL_CALLS = 0
S5_G5_READY = NO
```

**Stop:** No product-code fix in this gate. No skip of vet/race. No `forma-s5-frozen`. No S5-G5. Human authorization required before any allowlist expansion to fix `FakeContractPort` vet.

---

## Superseded by F8-F2 (history preserved)

**Date:** 2026-09-15
**Note:** This F8-F1 RESULT remains the durable record of `BLOCKED_BY_EXISTING_GO_VET` and tip CI failures `34925437921` / `34925707174`. The BLOCKED status is **not** rewritten.

**Closure:** S5-G4-F8-F2 deleted only the dead `bad := *contracts` / `_ = bad` lines in `validate_g3_test.go`, reordered pre-commit before vet/race, and obtained durable GREEN CI for both steps.

| Item | Value |
|------|-------|
| F8-F2 RESULT | `forma/cursor-results/FORMA-S5-G4-F8-F2-GO-VET-RACE-CLOSURE-RESULT.md` |
| F8-F2 IMPLEMENTATION_SHA | `f073b24dd7bb967bb0cf23c06bfe4af937c55be4` |
| F8-F2 IMPLEMENTATION_CI | https://github.com/hai12138/forma/actions/runs/34937595390 |
