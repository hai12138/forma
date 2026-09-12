# FORMA-S5-G2-F3 — Capability UoW / Analysis Safety Closure
# RESULT

**Gate:** S5-G2-F3
**Date:** 2026-09-13
**Status:** **PASS** (Forma CI ALL GREEN)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `6fb8c2e6e81bbe0407e83f479cc2d87260cfe0cf` |
| Scope | Capability UoW ownership + analysis safety only |
| PRODUCT_CODE_CHANGE | CAPABILITY_DOMAIN_ONLY |
| MIGRATION_CHANGE | S5_G2_F3_ONLY |
| REAL_MODEL_CALLS | 0 |
| HTTP / freeze / S5-G3 | NONE |

---

## 2. Fixes

| Area | Result |
|------|--------|
| Memory UoW ownership | PASS — owns memRepo + memAssets; `NewMemoryUnitOfWorkWith` removed |
| UoW failure seams | PASS — `MemoryUoWOptions` FailAssetCreate/Update/Commit + callbacks |
| Analysis secret isolation | PASS — `ValidateAnalysisRequest` + `containsCredentialShape` (no bare `\bsecret\b` on IDs) |
| Persisted request replay | PASS — `loadValidatedPersistedAnalysisRequest` (validate + digest match) |
| Attempt lifecycle | PASS — `SUPERSEDED` + `completed_at`; lease takeover supersedes prior |
| Structure validation | PASS — exact Predicate/Effect; comparand rules; bindings/mappings opaque |
| MapRepoError | PASS — bare sentinels; no driver text |
| ActorID required | PASS — StartAnalysis / RetryFailedAnalysis |

Local verification:
- `go test ./domain/forma/capability/... -count=1` → PASS
- `go test ./domain/forma/... -count=1` → PASS
- `node scripts/forma/migration-validate.mjs` → 18/18 PASS
- `git diff --check` → PASS

---

## 3. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `c2caf92e51ad3fb87ae23012eae1a992fbcea8f3` |
| CI_RUN | [34709533051](https://github.com/hai12138/forma/actions/runs/34709533051) |
| forma-backend | **PASS** |
| forma-migration-apply | **PASS** |
| forma-frontend | **PASS** |
| CI | **ALL GREEN** |

---

## 4. Gate Summary

```text
S5_G2_F3_STATUS = PASS
UOW_OWNERSHIP = PASS
UOW_COMMIT_ROLLBACK = PASS
ANALYSIS_SECRET_ISOLATION = PASS
LOAD_VALIDATED_PERSISTED = PASS
ATTEMPT_SUPERSEDE = PASS
STRUCTURE_VALIDATION = PASS
MAP_REPO_ERROR_BARE = PASS
PRODUCT_CODE_CHANGE = CAPABILITY_DOMAIN_ONLY
MIGRATION_CHANGE = S5_G2_F3_ONLY
REAL_MODEL_CALLS = 0
CI = ALL_GREEN
S5_G3_READY = NO
```

**Stop:** **STOP**. Do not start S5-G3. Do not create `forma-s5-frozen`. Await human review.
