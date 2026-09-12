# FORMA-S5-G2-F3 — Capability UoW / Analysis Safety Closure
# RESULT

**Gate:** S5-G2-F3
**Date:** 2026-09-13
**Status:** **PENDING_CI** (local tests green; await Forma CI)

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
| CI_RUN | PENDING |
| forma-backend | PENDING |
| forma-migration-apply | PENDING |
| forma-frontend | PENDING |
| CI | **PENDING_CI** |

---

## 4. Gate Summary

```text
S5_G2_F3_STATUS = PENDING_CI
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
CI = PENDING_CI
S5_G3_READY = NO
```

**Stop:** **STOP**. Do not start S5-G3. Do not create `forma-s5-frozen`. Await CI + human review.
