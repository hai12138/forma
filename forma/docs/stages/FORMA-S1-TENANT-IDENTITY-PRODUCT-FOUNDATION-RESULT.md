# FORMA S1 TENANT / IDENTITY / PRODUCT FOUNDATION RESULT

## Status

**PASS_WITH_GATES**

S1 tenancy / identity / product foundation is implemented. Local Go tests, migration A/B/C (incl. S1 tables), frontend typecheck/build/route smoke PASS. GitHub Actions Forma CI not verified from this environment after push (**EXTERNAL_GATE** for GATE-10).

**DO NOT START S2.** Await human review.

---

## S0 Frozen Baseline

| Item | Value |
|---|---|
| Tag | `forma-s0-frozen` |
| Commit | `d68a49bf1ae780f71d6aecd3ff6d3eb3a1c7a3e6` |
| Coze upstream | `fefb05ff27be1da939612fbf9faf5db62583b8ae` |
| Runtime | Eino |
| CI run (S0) | `33371051867` |

---

## Files Changed

Key additions:

- `backend/domain/forma/tenancy/**` — entity, DAL, repo, service, TenantContext
- `backend/domain/forma/errors/codes.go`
- `backend/application/forma/tenancy_app.go`, `tenant_context.go`
- `backend/api/middleware/forma_tenant.go`
- `backend/api/handler/forma/tenancy.go`, `assets.go`
- `backend/crossdomain/forma/integration/coze_space_adapter.go`
- `docker/atlas/forma/migrations/20250831120000_s1_tenancy.sql`
- `frontend/apps/forma` session provider, tenant switcher, overview counts
- `frontend/packages/forma-api-client` tenancy client
- ADRs 008–011, `TENANT-IDENTITY-BOUNDARY.md`
- Pre-commit CWD fix in `common/git-hooks/pre-commit` + install script

---

## Tenancy Domain

Tenant lifecycle: ACTIVE / SUSPENDED / ARCHIVED. Soft delete only. Optimistic `revision` on PATCH.

---

## Principal Model

`FormaPrincipal` with stable `provider=coze` + `external_subject=coze_user_id` mapping. Types: USER / SERVICE.

---

## Membership

Roles: OWNER / ADMIN / MEMBER / VIEWER. Statuses: ACTIVE / INVITED / SUSPENDED / REMOVED. Server verifies membership for TenantContext.

---

## Tenant → Coze Space Mapping

`forma_tenant_space_ref` — 1:N, purpose extensible. Active space unique to one tenant. No FK to Coze tables.

---

## TenantContext

Built server-side: Session → Principal → Membership → `X-Forma-Tenant` selection → AllowedSpaceIDs. Header is selection, not proof.

---

## Asset Registry Enforcement

All Get/List/Update/Archive require `tenant_id`. Cross-tenant Get denied in tests.

---

## API

| Method | Path |
|---|---|
| GET | `/api/forma/v1/me` |
| GET/POST | `/api/forma/v1/tenants` |
| GET/PATCH | `/api/forma/v1/tenants/:id` |
| GET/POST | `/api/forma/v1/tenants/:id/members` |
| PATCH | `/api/forma/v1/tenants/:id/members/:principalId` |
| GET/POST | `/api/forma/v1/tenants/:id/spaces` |
| POST | `/api/forma/v1/bootstrap` |
| GET | `/api/forma/v1/assets/counts` |

Error keys: FORMA_UNAUTHENTICATED, FORMA_TENANT_*, FORMA_MEMBERSHIP_*, FORMA_SPACE_*, FORMA_VERSION_CONFLICT.

---

## Frontend

Forma Shell: real tenant switcher (sessionStorage selection key only — not production asset store), identity display, banners for unauthenticated/forbidden/suspended/empty/network. Overview shows real asset counts / empty state. Navigation IA unchanged.

---

## Migration

`20250831120000_s1_tenancy.sql` — principal, tenant, membership, space_ref, audit_event. Local portable apply CASE A/B/C PASS (S0→S1 tables present).

---

## Audit

Skeleton events: TENANT_CREATED/UPDATED, MEMBER_*, SPACE_BOUND/UNBOUND via `forma_audit_event`.

---

## Security Tests

| Test | Result |
|---|---|
| Bootstrap idempotent | PASS |
| Tenant/Membership revision conflict | PASS |
| Asset cross-tenant Get denied | PASS |
| Forged `X-Forma-Tenant` forbidden | PASS |
| Unauthenticated resolve | PASS |

---

## CI

Extended Forma CI: S1 tenancy tests, S1 migration file check, pre-commit CWD check. S0 gates retained.

---

## Coze Core Files Modified

| File | Change |
|---|---|
| `application/application.go` | Forma tenancy init (FORMA blocks) |
| `api/router/register.go` | Forma router (S0) |
| `api/middleware/session.go` | Public meta paths (S0) |
| `common/git-hooks/pre-commit` | CWD → coze-studio root |
| `rush.json` | Forma projects (S0) |

New middleware file: `api/middleware/forma_tenant.go` (Forma-owned).

---

## Remaining Mock

Placeholder module pages still show “not connected yet”. No mock KPI on overview.

---

## Known Limitations

1. Full live Coze session E2E against running server not exercised in this agent session (handler/unit/integration tests cover contracts).
2. GitHub Actions GATE-10 requires human confirmation after push.
3. `.git/hooks/pre-commit` may need `scripts/forma/install-precommit.sh` on developer machines (tracked source in `common/git-hooks` is CWD-safe).
4. Bootstrap without available Coze Space creates Tenant without space binding (space bindable later).

---

## Gate Evidence

| Gate | Result | Evidence |
|---|---|---|
| GATE-01 Principal mapping | PASS | `ResolveOrCreatePrincipal` + Bootstrap tests |
| GATE-02 Multi-tenant membership | PASS | ListTenantsForPrincipal / memberships model |
| GATE-03 Multi-space binding | PASS | BindSpace + space_ref table/tests |
| GATE-04 Forged tenant header | PASS | `TestResolveTenantContext_ForgedHeaderForbidden` |
| GATE-05 Cross-tenant asset | PASS | `TestAssetRegistry_TenantIsolation_Get` |
| GATE-06 Suspended denial | PASS | TenantContext + FormaTenantMW error mapping |
| GATE-07 Bootstrap idempotent | PASS | `TestBootstrap_Idempotent` |
| GATE-08 Frontend identity | PASS | Session provider + shell switcher + build |
| GATE-09 S0→S1 migration | PASS | migration-apply-test CASE A/B/C + S1 tables |
| GATE-10 Forma CI | EXTERNAL_GATE | Confirm after push |

---

## Risks

| Risk | Mitigation |
|---|---|
| Space uniqueness index `(coze_space_id, status)` allows re-bind after INACTIVE | Documented; Shared Resource deferred |
| SessionStorage for selected tenant | Selection UX only; server enforces membership |

---

## S2 Preconditions

1. Human sign-off on this S1 report
2. Forma CI green on S1 commit
3. Live smoke: login → bootstrap → /me → switch tenant → asset counts
4. Do not start Business Model / Analyst until S1 PASS

---

**Stage:** FORMA-S1  
**DO NOT START S2.**
