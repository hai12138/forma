# FORMA-S5-G4-F5 — Bearer Safe-Phrase Exact Allowlist
# RESULT

**Gate:** S5-G4-F5
**Date:** 2026-09-14
**Status:** **PASS (local)** — Forma CI pending

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main) | `353db3daf9883c366e0de540e498fbbc46ff1fa6` |
| Freeze tip | `362a8f02eb82991acfdde303c1cd6bd419295e99` |
| PRODUCT_CODE_CHANGE | **YES** |
| MIGRATION_CHANGE | **NONE** |
| REAL_MODEL_CALLS | **0** |
| G5_CHANGE | **NONE** |

---

## 2. Parallel candidates (TDD)

| Agent | Branch | SHA |
|-------|--------|-----|
| A Tests | `forma/s5-g4-f5-agent-a-tests` | `ffe4a055e74fc474de7c811dabd25d9c9a3461db` |
| B Impl | `forma/s5-g4-f5-agent-b-implementation` | `1b8576356ac20eed877c5a2c1a05c519cb8e3e23` |

Integration order: cherry-pick A (F5 red on old impl) → cherry-pick B (green) → align F4 open-phrase allow cases to exact allowlist.

---

## 3. Delivered

| Item | Result |
|------|--------|
| Exact allowlist only `bearer of responsibility` (case-insensitive) | PASS |
| `Bearer of /abc` `+abc` `duty` `trust` rejected | PASS |
| Mixed phrase + credential rejected | PASS |
| Audit illegal reason → no Revision/Decision / lifecycle change | PASS |
| FAILED attempt paths unchanged (F4 regression suite) | PASS |
| `git diff --check 353db3da...HEAD` exit 0 | PASS |

---

## 4. Gate checklist

```text
S5_G4_F5_STATUS = PASS
BEARER_SAFE_PHRASE_EXACT = PASS
BEARER_OF_CREDENTIAL_REJECTION = PASS
MIXED_PHRASE_CREDENTIAL_REJECTION = PASS
AUDIT_NO_PARTIAL_WRITE = PASS
FAILED_ATTEMPT_REGRESSION = PASS
GIT_DIFF_CHECK = PASS
PRODUCT_CODE_CHANGE = YES
MIGRATION_CHANGE = NONE
G5_CHANGE = NONE
REAL_MODEL_CALLS = 0
COMMIT_SHA = c34c3e99506546e9a4a190546ebceb66adf904e7
CI_RUN = pending
CI = pending
S5_G5_READY = NO
```

---

## 5. Local verification

| Gate | Result |
|------|--------|
| Capability domain / application / handler | PASS |
| Full Forma backend packages | PASS |
| migration-validate 19/19 | PASS |
| CASE A/B/C | deferred to CI |
| typecheck + routes-smoke | PASS |
| Rush `@forma/app` build | deferred to CI (Windows rtsc.sh) |

---

## 6. CI

| Field | Value |
|-------|-------|
| CI_RUN | _pending_ |
| forma-backend | _pending_ |
| forma-migration-apply | _pending_ |
| forma-frontend | _pending_ |

**Stop:** No `forma-s5-frozen`. No G5 / Runtime / Agent-Workflow projection. `S5_G5_READY = NO`. Await human review after CI ALL GREEN.
