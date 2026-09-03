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
| CI | PENDING (awaiting GitHub Forma CI after F1 push) |
| S5_G1_READY | NO |

> Truthfulness rule: `S5_G0_STATUS = PASS` / `S5_G1_READY = YES` only after new exact commit CI is ALL GREEN.

## Baseline

| Key | Value |
|---|---|
| S4_BASELINE | `forma-s4-frozen-r2` |
| FREEZE_COMMIT | `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| IMPLEMENTATION_COMMIT (S5-G0) | `cb6500b02072e056ca450efec5a8f6b92431e58b` |
| F1_FIX_COMMIT | _(filled after push)_ |

## Login Account Contract (LOCKED)

**Forma Local Account Alias**

| Input | Canonical Email | Product Account |
|---|---|---|
| `admin` | `admin@forma.local` | `admin` |
| `user01` | `user01@forma.local` | `user01` |
| `user01@forma.local` | `user01@forma.local` | `user01` |
| `email@example.com` | `email@example.com` | `email@example.com` |

Rules:

- Local aliases and `@forma.local` emails are the **same identity** (no duplicate principals).
- Reject whitespace, control characters, path characters (`/`, `\`), invalid local formats.
- Not “bootstrap alias only” — product supports local accounts for admin-created users.

## F1 Fixes

1. **Backend test contract** — `memPrincipalRepo` implements `UpdateStatus` + `ListAll`; `memPlatformRoleRepo` added.
2. **Frontend typecheck** — shell nav rendering uses shared `NavItem`/`NavGroup` helpers (fixes TS2367).
3. **Atlas checksum** — regenerated via `atlas migrate hash` for `20250903000000_s5_g0_platform_admin.sql`.
4. **Alias contract** — centralized `NormalizeAccount` + docs/result alignment.
5. **Security tests** — bootstrap, password-change-required, create/disable/enable/reset, last SUPER_ADMIN protection.

## Architecture

```
Forma UI → Forma Auth Adapter (login / change-password / logout)
    ↓
Coze User Domain (email/password, Argon2id, session)
    ↓
Forma Principal + PlatformRole (SUPER_ADMIN / USER)
    ↓
Forma Tenant Membership (OWNER / ADMIN / MEMBER / VIEWER)
```

## Security Invariants

- `admin` / `admin123` = INITIAL PASSWORD ONLY
- First login → `PASSWORD_CHANGE_REQUIRED`
- Bootstrap = `CREATE_IF_ABSENT` (restart must not reset changed password)
- SUPER_ADMIN enforced backend-side
- Last SUPER_ADMIN cannot be disabled/demoted
- Disabled user login/access denied
- `COZE_AUTH_CORE_CHANGE = NONE`

## Final Gate (after CI)

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
