# FORMA-S5-G3-F3 — Activate Final Evidence Fence Window Closure
# RESULT

**Gate:** S5-G3-F3
**Date:** 2026-09-13
**Status:** **PENDING_CI** (implementation committed; awaiting Forma CI)

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main) | `ab53543a1ab4eb06cbb2438db1ded204f8d2e21c` |
| S5-G3-F2 implementation | `74db871d4a944a2bcd0957374c5d39a91381b319` |
| Scope | Close Activate final-fence → first status-write window |
| PRODUCT_CODE_CHANGE | **CAPABILITY_DOMAIN_ONLY** |
| MIGRATION_CHANGE | **NONE** |
| REAL_MODEL_CALLS | 0 |
| HTTP / UI / Impact / S5-G4 | NONE |

---

## 2. Delivered

| Fix | Result |
|-----|--------|
| `Activate`: `ListRevisions` moved before `requireFinalEvidenceFence` | PASS |
| Order: initial evidence → PASS check → list revs → final fence → digest recheck → first `UpdateRevisionStatus` | PASS |
| No port/repo/DAO between final fence success and first status write (memory loop only) | PASS |
| Deterministic race: `ListRevisions` hook flips Contract evidence; fence aborts Activate | PASS |
| ACTIVE / VALIDATED / active pointer / generation / decisions / AssetRef unchanged on race | PASS |
| Removed unused `flipAfterReadsContractPort.driftKey` | PASS |
| S5-G3 / F1 / F2 regressions retained | PASS |

Local verification:
- `go test ./domain/forma/capability/... -count=1` → PASS
- `go test ./domain/forma/... -count=1` → PASS
- `node coze-studio/scripts/forma/migration-validate.mjs` → 19/19 PASS
- `git diff --check` → PASS

---

## 3. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `36c0078b922d628c583a02d6d14594e5ab76c5ea` |
| CI_RUN | _(pending push)_ |
| forma-backend | _(pending)_ |
| forma-migration-apply | _(pending)_ |
| forma-frontend | _(pending)_ |
| CI | PENDING |

---

## 4. Gate Summary

```text
S5_G3_F3_STATUS = PENDING_CI
ACTIVATE_FENCE_WINDOW_CLOSED = PASS
LIST_REVISIONS_BEFORE_FENCE = PASS
RACE_HOOK_TEST = PASS
PRODUCT_CODE_CHANGE = CAPABILITY_DOMAIN_ONLY
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
CI = PENDING
S5_G4_READY = NO
```

**Stop:** Do not start S5-G4. Await Forma CI + human review.
