# FORMA-S5-G4-F7 — Capability UI Read Seam
# RESULT

**Gate:** S5-G4-F7
**Date:** 2026-09-14
**Status:** **PASS / CI ALL GREEN**

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main) | `5dfc715105a927243311ef8f04aa0a5bfb28c730` |
| Freeze tip | `1a8896082ae93fc924cf81bec084c24a1ca42b5e` |
| PRODUCT_CODE_CHANGE | **CAPABILITY_READ_SEAM_ONLY** |
| MIGRATION_CHANGE | **NONE** |
| G5_CHANGE | **NONE** |
| RUNTIME_CHANGE | **NONE** |
| REAL_MODEL_CALLS | **0** |

---

## 2. Parallel candidates (TDD)

| Agent | Branch | SHA |
|-------|--------|-----|
| A Tests | `forma/s5-g4-f7-agent-a-tests` | `cd51c35256db20e06a84bb6f733bc06102efbef9` |
| B Impl | `forma/s5-g4-f7-agent-b-implementation` | `b8b0d84b187d246a5d7d42b7e0e7d1a2d90fdcf3` |

Integration: cherry-pick A (RED) → B as `fix(...)` → integrator align + red-team harden.

---

## 3. Delivered

| Item | Result |
|------|--------|
| CapabilityDTO AssetRef projection fields | PASS |
| ListCapabilities one tenant AssetRef batch | PASS |
| Get/Create projection (Create via ProjectCapabilityAssetRef) | PASS |
| GET capability-proposals/:proposalId | PASS |
| GET capability-analyses/:analysisRunId/proposals | PASS |
| ACTIVE membership read; SUPER_ADMIN no implicit tenant | PASS |
| Tenant/business isolation; logical-only; secret-free | PASS |
| Read no mutation; Analysis GET backward compatible | PASS |
| Red-team P1 remediated (missing→409; MapRepoError; list fail-closed) | PASS |

---

## 4. Gate checklist

```text
S5_G4_F7_STATUS = PASS
CAPABILITY_ASSET_PROJECTION = PASS
CAPABILITY_LIST_NO_N_PLUS_ONE = PASS
CAPABILITY_GET_PROJECTION = PASS
CAPABILITY_CREATE_PROJECTION = PASS
PROPOSAL_GET = PASS
ANALYSIS_PROPOSAL_LIST = PASS
ACTIVE_MEMBERSHIP_READ = PASS
SUPER_ADMIN_NO_IMPLICIT_TENANT_ACCESS = PASS
TENANT_BUSINESS_ISOLATION = PASS
LOGICAL_ONLY = PASS
SECRET_ISOLATION = PASS
READ_NO_MUTATION = PASS
BACKWARD_COMPATIBILITY = PASS
RED_THEN_GREEN = PASS
ADVERSARIAL_REVIEW = PASS
GIT_DIFF_CHECK = PASS
PRODUCT_CODE_CHANGE = CAPABILITY_READ_SEAM_ONLY
MIGRATION_CHANGE = NONE
G5_CHANGE = NONE
RUNTIME_CHANGE = NONE
REAL_MODEL_CALLS = 0
COMMIT_SHA = 546ba5a8e419581aae32d305307ce2c2c227dd91
LATEST_TIP_SHA = 42ee6dce3d1b6a54f249b617a8d02acb941992d1
CI_RUN = https://github.com/hai12138/forma/actions/runs/34864115804
CI = ALL GREEN
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
| CI_RUN | https://github.com/hai12138/forma/actions/runs/34864115804 |
| tip SHA checked | `2ba01c94793e2b9daece3bd7dc49e28f6943f51d` |
| forma-backend | success |
| forma-migration-apply | success |
| forma-frontend | success |

**Stop:** No `forma-s5-frozen`. No G5 / Runtime / `@forma/capability` / frontend route changes. Await human review. `S5_G5_READY = NO`.

---

## Human review finding / superseded by F8

**Date:** 2026-09-15
**Finding:** Human review identified remaining Read Seam gaps after F7 PASS:

1. Proposal / Analysis lookups did not fail-closed on tenant / requested-ID identity mismatches (could leak foreign entity DTOs).
2. `ListCapabilityProposalsByAnalysis` lacked full totality for nil entry, blank `proposal_id`, and duplicate `proposal_id` (partial list risk).
3. F7 Freeze file-whitelist deviations needed audit record without rewriting F7 Freeze history.

**Superseded by:** S5-G4-F8 — `forma/cursor-results/FORMA-S5-G4-F8-READ-SEAM-FAIL-CLOSED-RESULT.md`
**F8 IMPLEMENTATION_SHA:** `27059ef3a884e941ab526e59dbe1e4c44fa9226b`
**F8 IMPLEMENTATION_CI:** https://github.com/hai12138/forma/actions/runs/34912401820

This appendix does **not** delete or rewrite prior F7 RESULT content. F7 delivered the UI Read Seam surfaces; F8 closes identity + list totality fail-closed hardening only.
