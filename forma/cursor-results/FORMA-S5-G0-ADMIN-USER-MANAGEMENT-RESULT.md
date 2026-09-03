# FORMA-S5-G0 — Platform Admin / User Management Foundation

## Status

| Gate | Status |
|---|---|
| **S5_G0_STATUS** | **PENDING_CI** |
| **S5_G0_F1_STATUS** | **PENDING_CI** |
| DEFAULT_ADMIN | LOCAL_PASS |
| ADMIN_USER_MANAGEMENT | LOCAL_PASS |
| COZE_AUTH_CORE_CHANGE | NONE |
| S4_REGRESSION | LOCAL_PASS |
| REAL_MODEL_CALLS | 0 |
| CI | PENDING (awaiting GitHub Forma CI on tip) |
| S5_G1_READY | NO |

> Truthfulness rule: `S5_G0_STATUS = PASS` / `S5_G1_READY = YES` only after tip commit CI is ALL GREEN.

## Commits

| Key | Value |
|---|---|
| S4_BASELINE | `forma-s4-frozen-r2` |
| IMPLEMENTATION_COMMIT | `cb6500b02072e056ca450efec5a8f6b92431e58b` |
| F1_FIX_COMMIT | `da20a7b83e094fa382b12a2aa81a218b86651b0e` |
| F1_TEST_ALIGN_COMMIT | `b6950e01fc47f89d53302c4649cd2da8c0776981` |
| TIP_COMMIT | _(filled after this push)_ |

## Login Account Contract (LOCKED)

**Forma Local Account Alias**

| Input | Canonical Email | Product Account |
|---|---|---|
| `admin` | `admin@forma.local` | `admin` |
| `user01` | `user01@forma.local` | `user01` |
| `user01@forma.local` | `user01@forma.local` | `user01` |
| `email@example.com` | `email@example.com` | `email@example.com` |

Rules:

- Local aliases and `@forma.local` emails are the **same identity**.
- Reject whitespace, control characters, path characters (`/`, `\`), invalid local formats.
- Not “bootstrap alias only”.

## F1 Fixes

1. Backend test doubles: `UpdateStatus` + `ListAll` + `memPlatformRoleRepo`
2. Frontend shell TS2367: shared `NavItem`/`NavGroup` renderer
3. Atlas checksum regenerated for `20250903000000_s5_g0_platform_admin.sql`
4. Locked Forma Local Account Alias via `NormalizeAccount`
5. Admin security tests + passport unit test alignment
6. Disabled-user login denied after session creation

## Local regression (pre-CI)

| Check | Result |
|---|---|
| `go test ./domain/forma/tenancy/... ./application/forma/...` | PASS |
| `node scripts/forma/typecheck.mjs` | PASS |
| `node scripts/forma/migration-apply-test.mjs` (CASE A/B/C) | PASS |
| `@forma/app` vitest | PASS |
| `COZE_AUTH_CORE_CHANGE` vs `forma-s4-frozen-r2` | NONE |

## Security Invariants

- `admin` / `admin123` = INITIAL PASSWORD ONLY
- First login → `PASSWORD_CHANGE_REQUIRED`
- Bootstrap = `CREATE_IF_ABSENT`
- SUPER_ADMIN backend authorization
- Last SUPER_ADMIN protection
- Disabled user denied
- `COZE_AUTH_CORE_CHANGE = NONE`

## Final Gate (after CI ALL GREEN)

```
S5_G0_F1_STATUS = PASS
S5_G0_STATUS = PASS
DEFAULT_ADMIN = PASS
ADMIN_USER_MANAGEMENT = PASS
COZE_AUTH_CORE_CHANGE = NONE
S4_REGRESSION = PASS
REAL_MODEL_CALLS = 0
COMMIT_SHA = ...
CI_RUN = ...
CI = ALL GREEN
S5_G1_READY = YES
```
