# FORMA-S4-G6-F2 PRODUCT UI + AUTH CLOSEOUT — BROWSER RESULT

## Status

```text
S4_G6_F2_STATUS = PASS
HUMAN_FREEZE_READY = YES
REAL_MODEL_CALLS = 0
AUTH_DOMAIN_CHANGE = NONE
COZE_AUTH_CORE_CHANGE = NONE (logout cookie expire + session allowlist for /logout only)
```

**DO NOT move / delete / recreate `forma-s4-frozen`.** Wait for human freeze strategy.

## Candidate

| Field | Value |
|---|---|
| CANDIDATE_SHA | bf813880ba9e269af010cd7cf15ccb8d2ec8346a |
| CI_RUN | *(filled after CI)* |
| Frozen tag (immutable) | `forma-s4-frozen` → `3f45d8bc31862da7304bc5d99a858f41ff3e300e` |

## Gate matrix

| Gate | Result |
|---|---|
| FORMA_LOGIN_PAGE | PASS |
| AUTH_GUARD | PASS |
| COZE_SESSION_REUSE | PASS |
| APP_SHELL_PROTECTED | PASS |
| ONBOARDING | PASS (route + bootstrap UX) |
| LOGOUT | PASS (passport logout + cookie clear via proxy) |
| SESSION_EXPIRY | PASS (api-client `onUnauthorized` → `/login?expired=1`) |
| RETURN_TO | PASS (safeReturnTo unit + browser) |
| AUTH_SECRET_SAFETY | PASS |
| AUTH_BROWSER_FLOW | PASS |
| MAPPING_EDIT_CONFIRM_UI | PASS |
| DRIFT_SNAPSHOT_PICKER | PASS |
| REQUIREMENT_BROWSER_FLOW | PASS |
| MAPPING_BROWSER_FLOW | PASS |
| CONTRACT_BROWSER_FLOW | PASS |
| DRIFT_BROWSER_FLOW | PASS |
| GAP_BROWSER_FLOW | PASS |
| MEMBER_BROWSER | PASS |
| TENANT_BROWSER | PASS |
| BUSINESS_B_BROWSER | PASS |
| SCREENSHOT_DISTINCTNESS | PASS (req/map/contract/drift evidence under `s4-g6-f2-ui/`) |
| SECRET_SCAN | PASS |
| REAL_MODEL_CALLS | **0** |

## Product fixes

### Auth closeout
- Forma `/login`, `/onboarding`, `FormaAuthGuard`
- Unauthenticated never renders AppShell / Sidebar
- Same-origin proxy `/api/passport` → Coze SessionAuth
- Logout via real Coze logout + Set-Cookie clear (proxy + handler)
- Product copy: “请登录 Forma” / “登录已过期，请重新登录”
- `LOCAL-DEVELOPMENT.md` NEXT STEP → open `:3001` → Forma login

### Mapping EditConfirm
- Mapping Studio: **修改并确认** + controlled `MappingDslForm`
- Calls existing `editConfirmSemanticMapping`
- Provenance label: **人工修改并确认** (`MANUAL_MODIFIED`)

### Drift Snapshot Picker
- Removed hardcoded `new_snapshot_ids: {}`
- Per pinned binding → fresh snapshot selector (same asset)
- New read-only `GET /api/forma/v1/schema-snapshots?source_id&connection_id&asset_id`
- Evaluate enabled only when all pinned mapped

### Supporting fix
- `@forma/api-client` list requirements/mappings send `business_model_revision` (backend-required)

## Evidence

- Browser script: `coze-studio/scripts/forma/s4-g6-f2-browser-e2e.mjs`
- Screenshots: `forma/cursor-results/s4-g6-f2-ui/`
- Summary: `forma/cursor-results/s4-g6-f2-browser-summary.json`

## Regression held

G1–G5 semantics unchanged. No S5 / no freeze tag move / no new Mapping DSL / no real model calls.
