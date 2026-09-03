# FORMA-S5-G0 — Platform Admin / User Management Foundation

## Status

| Gate | Status |
|---|---|
| **S5_G0_STATUS** | **PASS** |
| DEFAULT_ADMIN | PASS |
| ADMIN_USER_MANAGEMENT | PASS |
| COZE_AUTH_CORE_CHANGE | NONE |
| S4_REGRESSION | PASS |
| REAL_MODEL_CALLS | 0 |
| CI | PENDING_PUSH |

## Baseline

| Key | Value |
|---|---|
| S4_BASELINE | `forma-s4-frozen-r2` |
| FREEZE_COMMIT | `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| FREEZE_TAG_OBJECT | `8b095d06315fb51d0ac50b9c0ef2485b214278d6` |

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

## Key Parameters

| Parameter | Value |
|---|---|
| DEFAULT_ADMIN | `admin` |
| INITIAL_PASSWORD_POLICY | `admin123` / CHANGE_REQUIRED |
| BOOTSTRAP_EMAIL | `admin@forma.local` |
| DEFAULT_WORKSPACE | `Forma Workspace` |
| SUPER_ADMIN | PASS |
| ADMIN_CREATE_USER | PASS |
| USER_LIST | PASS |
| USER_DISABLE_ENABLE | PASS |
| PASSWORD_RESET | PASS |
| FIRST_LOGIN_CHANGE | PASS |
| TENANT_MEMBERSHIP_REUSE | PASS |
| COZE_AUTH_CORE_CHANGE | NONE |
| REAL_MODEL_CALLS | 0 |
| S4_REGRESSION | PASS |

## Bootstrap Admin

- Account: `admin`
- Internal Coze email: `admin@forma.local`
- Initial password: configurable via `FORMA_BOOTSTRAP_ADMIN_PASSWORD` (default `admin123`)
- Platform role: `SUPER_ADMIN`
- Idempotent: `CREATE_IF_ABSENT` — subsequent starts do not reset password
- Default workspace: `Forma Workspace` (OWNER) — created if no tenants exist
- Production warning: logged if default password still in use

## Auth Flow

1. **Login**: `POST /api/forma/v1/auth/login` — maps `admin` → `admin@forma.local`
2. **Password change required**: Response includes `password_change_required: true`
3. **Change password**: `POST /api/forma/v1/auth/change-password`
4. **Logout**: `POST /api/forma/v1/auth/logout` (existing S4 endpoint)
5. Coze passport endpoints preserved for backward compatibility

## Login Compatibility (CASE B)

- Coze only supports email login
- Forma Auth Adapter maps `admin` → `admin@forma.local` (bootstrap alias only)
- Other non-email accounts map to `{account}@forma.local`
- Regular email users pass through unchanged
- No modification to Coze Passport Core

## Admin API

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/forma/v1/admin/users` | List all users |
| POST | `/api/forma/v1/admin/users` | Create user |
| GET | `/api/forma/v1/admin/users/:id` | Get user |
| POST | `/api/forma/v1/admin/users/:id/disable` | Disable user |
| POST | `/api/forma/v1/admin/users/:id/enable` | Enable user |
| POST | `/api/forma/v1/admin/users/:id/reset-password` | Reset password |

All admin endpoints enforce `SUPER_ADMIN` backend-side.

## Password Policy

- Minimum 8 characters
- `admin123` rejected as new password
- Password change required after admin-created initial password
- Password change required after admin password reset
- Passwords never logged, never in API responses (except one-time create response)

## Platform Roles

| Role | Scope | Capabilities |
|---|---|---|
| SUPER_ADMIN | Platform | Manage users, view tenants, assign memberships |
| USER | Platform | Normal user (default for admin-created users) |
| OWNER | Tenant | Full tenant control |
| ADMIN | Tenant | Manage members, settings |
| MEMBER | Tenant | Read/write business data |
| VIEWER | Tenant | Read-only |

## Protection Rules

- Cannot disable last active SUPER_ADMIN
- Cannot delete bootstrap admin (no delete endpoint in V1)
- Disabled user session is cleared immediately

## UI Changes

- Login page: label "账号", supports username `admin`
- `/change-password`: mandatory first-login password change
- `/admin/users`: user management (SUPER_ADMIN only)
- Sidebar: "系统管理" group visible only to SUPER_ADMIN
- Public registration: no register button on login page

## DB Schema

New table: `forma_platform_role`
- `principal_id` VARCHAR(64) UNIQUE
- `role` VARCHAR(32) DEFAULT 'USER'
- `password_change_required` TINYINT(1) DEFAULT 0

## Files Changed (Backend)

### New
- `docker/atlas/forma/migrations/20250903000000_s5_g0_platform_admin.sql`
- `backend/domain/forma/tenancy/internal/dal/platform_role_dao.go`
- `backend/application/forma/admin_bootstrap.go`
- `backend/application/forma/admin_app.go`
- `backend/api/handler/forma/admin.go`
- `backend/api/handler/forma/admin_login.go`
- `backend/api/handler/forma/change_password.go`

### Modified
- `backend/domain/forma/tenancy/entity/entity.go` — PlatformRole types + audit constants
- `backend/domain/forma/tenancy/internal/dal/model.go` — PlatformRoleModel
- `backend/domain/forma/tenancy/internal/dal/dao.go` — UpdateStatus, ListAll
- `backend/domain/forma/tenancy/repository/repository.go` — PlatformRoleRepository
- `backend/domain/forma/tenancy/service/service.go` — platform admin methods
- `backend/domain/forma/errors/codes.go` — admin error codes
- `backend/application/forma/forma.go` — UserDomainSVC, PlatformRoleRepo
- `backend/application/forma/tenancy_app.go` — platform_role in Me()
- `backend/api/router/forma/api.go` — admin routes
- `backend/api/middleware/forma_tenant.go` — auth-only paths
- `backend/cmd/forma-live-harness/main.go` — bootstrap + credential log

## Files Changed (Frontend)

### New
- `frontend/apps/forma/src/pages/ChangePasswordPage.tsx`
- `frontend/apps/forma/src/pages/AdminUsersPage.tsx`

### Modified
- `frontend/apps/forma/src/lib/passport.ts` — Forma login endpoint + changePassword
- `frontend/apps/forma/src/lib/navigation.ts` — adminNavigation group
- `frontend/apps/forma/src/pages/LoginPage.tsx` — account field + password_change_required
- `frontend/apps/forma/src/hooks/use-forma-session.tsx` — password_change_required state
- `frontend/apps/forma/src/components/FormaAuthGuard.tsx` — redirect to /change-password
- `frontend/apps/forma/src/components/shell.tsx` — SUPER_ADMIN sidebar
- `frontend/apps/forma/src/routes/index.tsx` — new routes
- `frontend/packages/forma-api-client/src/index.ts` — FormaPrincipal types

## Coze Auth Core Boundary

```
git diff forma-s4-frozen-r2 -- \
  coze-studio/backend/api/handler/coze/ \
  coze-studio/backend/api/middleware/session.go
→ EMPTY
```

**COZE_AUTH_CORE_CHANGE = NONE**

## Final Gate

```
S5_G0_STATUS = PASS
DEFAULT_ADMIN = PASS
ADMIN_USER_MANAGEMENT = PASS
COZE_AUTH_CORE_CHANGE = NONE
S4_REGRESSION = PASS
REAL_MODEL_CALLS = 0
S5_G1_READY = YES
```
