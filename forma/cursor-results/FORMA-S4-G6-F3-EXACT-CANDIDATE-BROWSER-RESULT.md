# FORMA-S4-G6-F3 EXACT-CANDIDATE BROWSER RESULT

## Status

```text
S4_G6_F3_STATUS = PASS
HUMAN_FREEZE_READY = YES
CANDIDATE_SHA = c213a8a4ac906ccaf4383dc4c551eba4d6abc1f9
BROWSER_SUMMARY_SHA = 9db03944455eef26bd15706ea71962e3d42dfcbd5e07fad9a31a165b0a058c23
COZE_AUTH_CORE_CHANGE = NONE
REAL_MODEL_CALLS = 0
PRODUCT_DOMAIN_CHANGE = NONE
AUTH_DOMAIN_CHANGE = NONE
```

**DO NOT move / delete / recreate `forma-s4-frozen`.** Wait for human freeze strategy.

## Candidate

| Field | Value |
|---|---|
| CANDIDATE_SHA | `c213a8a4ac906ccaf4383dc4c551eba4d6abc1f9` |
| BROWSER_SUMMARY_SHA | `9db03944455eef26bd15706ea71962e3d42dfcbd5e07fad9a31a165b0a058c23` |
| GIT_WORKTREE_CLEAN at browser start | `true` |
| Frozen tag (immutable) | `forma-s4-frozen` → `3f45d8bc31862da7304bc5d99a858f41ff3e300e` |
| CI_RUN | https://github.com/hai12138/forma/actions/runs/33760794665 (tip `d6e9161e` includes candidate; forma-backend / forma-migration-apply / forma-frontend ALL GREEN) |

## Auth boundary

| Gate | Result |
|---|---|
| FORMA_LOGIN_PAGE | PASS |
| AUTH_GUARD | PASS (preserved from F2) |
| Forma-owned `POST /api/forma/v1/auth/logout` | PASS |
| COZE_AUTH_CORE_CHANGE | **NONE** — `git diff forma-s4-frozen^{} -- coze-studio/backend/api/handler/coze/ coze-studio/backend/api/middleware/session.go` is empty |
| Login via proxied Coze passport | PASS (unchanged Coze login implementation) |
| Session middleware allowlist | unchanged vs frozen (no `/logout` allowlist patch) |

## Browser machine summary

Source: `forma/cursor-results/s4-g6-f3-browser-summary.json`

| Gate | Result |
|---|---|
| AUTH_BROWSER_FLOW | PASS |
| REQUIREMENT_CONFIRM | PASS |
| REQUIREMENT_REJECT | PASS |
| REQUIREMENT_EDIT_CONFIRM | PASS |
| MAPPING_CONFIRM | PASS |
| MAPPING_REJECT | PASS |
| MAPPING_EDIT_CONFIRM | PASS |
| CONTRACT_VALIDATE | PASS |
| CONTRACT_ACTIVATE | PASS |
| DRIFT_COMPATIBLE | PASS |
| DRIFT_BREAKING_TO_STALE | PASS |
| GAP_BROWSER | PASS |
| MEMBER_BROWSER | PASS |
| TENANT_BROWSER | PASS |
| BUSINESS_B_ACTIVE_CONTRACT | PASS |
| SECRET_SCAN | PASS |
| REAL_MODEL_CALLS | 0 |
| FIXTURE_SETUP_ONLY | true |
| SCREENSHOT_LIMITATION | null (no DOM forgery) |

## Evidence

Screenshots: `forma/cursor-results/s4-g6-f3-ui/`

## Stop

- Do **not** move `forma-s4-frozen`
- Do **not** create a new freeze tag
- Do **not** start S5
