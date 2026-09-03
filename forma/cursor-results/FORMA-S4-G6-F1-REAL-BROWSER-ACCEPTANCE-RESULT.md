# FORMA-S4-G6-F1 REAL BROWSER WORKFLOW ACCEPTANCE — RESULT

## Status

```
S4_G6_F1_STATUS = FAIL / PRODUCT_FIX_REQUIRED
HUMAN_FREEZE_READY = NO
PRODUCT_UNDER_TEST_SHA = 3f45d8bc31862da7304bc5d99a858f41ff3e300e
PRODUCT_CODE_CHANGE = NONE
REAL_MODEL_CALLS = 0
FREEZE_TAG = forma-s4-frozen (UNCHANGED)
FREEZE_TAG_OBJECT_SHA = fb8749540e994623f65e451adb472be75b1c4f06
FREEZE_TARGET = 3f45d8bc31862da7304bc5d99a858f41ff3e300e
TAG_CI = 33711584069 PASS
```

**DO NOT move / rewrite / recreate `forma-s4-frozen`.**  
**DO NOT start S5.**  
Await human decision on a new Freeze Candidate after product UI fixes.

---

## 0. Scope

This gate corrects the G6 Browser Evidence gap: prior `s4-g6-browser-e2e.mjs` only navigated / refreshed / screenshot-duplicated routes and did **not** exercise:

Browser click → API → Backend → Persisted state → UI refresh.

F1 requires those workflows on the **exact frozen product** at `3f45d8bc…`.

---

## 1. Tag CI correction (docs)

Prior G6 Result incorrectly recorded `TAG_CI = NOT_CONFIGURED`.

Fact:

| Item | Value |
|------|-------|
| Workflow | Forma CI listens `tags: ['forma-*']` |
| Run | [33711584069](https://github.com/hai12138/forma/actions/runs/33711584069) |
| Head | `forma-s4-frozen` |
| Target commit | `3f45d8bc31862da7304bc5d99a858f41ff3e300e` |
| Result | **ALL GREEN** |

`TAG_CI = 33711584069 PASS`

---

## 2. Product inspection (frozen SHA)

Inspected UI at `3f45d8bc` under `coze-studio/frontend/packages/forma-data`:

| Area | Finding |
|------|---------|
| Requirements | Confirm / Reject / EditConfirm **present** (`confirm-requirement`, `reject-requirement`, `edit-confirm-requirement`, `submit-edit-confirm`) |
| Mappings | Confirm / Reject **present**; **EditConfirm absent** (API client has `editConfirmSemanticMapping`, page never wires it) |
| Contract detail | Validate / Activate **present** (`validate-revision`, `activate-revision`, `activate-confirm`) |
| Health Drift | Button `evaluate-drift` **hardcodes** `new_snapshot_ids: {}` — **no snapshot picker** |
| Health Gap | `evaluate-gap` present |
| MEMBER | Mutation controls **hidden** when `!isEditor` |

Backend `EvaluateDrift` (frozen) **requires** every pinned `SchemaSnapshotID` to appear in `NewSnapshotIDs` or returns:

`FORMA_DATA_CONTRACT_DRIFT_INVALID: missing new snapshot for pinned "…"`

Therefore the Health UI button cannot complete a real compatible/breaking drift evaluation against an ACTIVE contract.

---

## 3. Blocking gates (product)

| Gate | Required | Frozen product | Result |
|------|----------|----------------|--------|
| MAPPING_BROWSER_FLOW | Browser **Edit & Confirm** | No EditConfirm control | **BLOCKED** |
| DRIFT_BROWSER | Browser Evaluate Drift with compatible + breaking fresh snapshots; UI refresh to STALE | UI always posts `{}` → backend DENIED / cannot select snapshots | **BLOCKED** |

These are **not** harness-only gaps. Fixing them requires product UI changes on a **new** candidate (not this freeze tag).

### Minimal product fixes (for a future candidate — NOT applied)

1. **Mapping Studio** — add EditConfirm path for `PROPOSED` mappings (reuse `editConfirmSemanticMapping`), with DOM status + lineage after submit.
2. **Data Health** — accept `new_snapshot_ids` (pinned → fresh) via inputs or binding-aware picker; wire `evaluate-drift` to that map instead of `{}`.

---

## 4. What was NOT done (by rule)

- No product code change
- No freeze tag move / recreate
- No real model calls
- No superseding freeze
- No S5 start
- Did **not** claim PASS by API-mutating then screenshotting

---

## 5. Gate matrix (F1)

| Gate | Result | Notes |
|------|--------|-------|
| PRODUCT_UNDER_TEST_SHA | PASS (recorded) | `3f45d8bc31862da7304bc5d99a858f41ff3e300e` |
| REAL_MODEL_CALLS | PASS | `0` |
| REQUIREMENT_BROWSER_FLOW | NOT_RUN | Blocked by overall PRODUCT_FIX_REQUIRED stop; UI appears capable |
| MAPPING_BROWSER_FLOW | **FAIL / PRODUCT_FIX_REQUIRED** | Missing EditConfirm UI |
| CONTRACT_VALIDATE_BROWSER | NOT_RUN | UI appears capable |
| CONTRACT_ACTIVATE_BROWSER | NOT_RUN | UI appears capable |
| DRIFT_BROWSER | **FAIL / PRODUCT_FIX_REQUIRED** | Empty `new_snapshot_ids` |
| GAP_BROWSER | NOT_RUN | UI appears capable (no snapshot map needed) |
| MEMBER_BROWSER | NOT_RUN | UI hides mutations for MEMBER |
| TENANT_BROWSER | NOT_RUN | — |
| BUSINESS_B_BROWSER | NOT_RUN | — |
| SCREENSHOT_STATE_DISTINCTNESS | FAIL | Prior G6 screenshots invalid; new distinct evidence not produced (stop) |
| SECRET_SCAN | N/A (no new live browser run) | — |
| TAG_CI | **PASS** | `33711584069` |
| PRODUCT_CODE_CHANGE | **NONE** | As required on FAIL path |

---

## 6. Root cause (human review finding)

G6 browser harness only exercised navigation/a11y/screenshot. Multiple “state” PNGs shared identical content hashes — invalid as Human Confirm / Active / Stale / Member / Tenant evidence.

F1 attempted real workflow acceptance against frozen product and discovered **UI contract incompleteness** that prevents hard gates without product change.

---

## 7. Recommended next human decision

1. Open a **new** S4 freeze-candidate branch from `3f45d8bc` (or main tip).
2. Apply the two minimal UI fixes above (+ tests).
3. Re-run full F1 real-browser workflow on that candidate.
4. Only then consider a **new** annotated freeze process — **do not** rewrite `forma-s4-frozen`.

---

## 8. Final

```
S4_G6_F1_STATUS = FAIL / PRODUCT_FIX_REQUIRED
HUMAN_FREEZE_READY = NO
BLOCKING =
  - MAPPING_EDIT_CONFIRM_UI_MISSING
  - HEALTH_DRIFT_SNAPSHOT_MAP_HARDCODED_EMPTY
PRODUCT_CODE_CHANGE = NONE
forma-s4-frozen = UNCHANGED → 3f45d8bc31862da7304bc5d99a858f41ff3e300e
TAG_CI = 33711584069 PASS
REAL_MODEL_CALLS = 0
DO NOT START S5
```
