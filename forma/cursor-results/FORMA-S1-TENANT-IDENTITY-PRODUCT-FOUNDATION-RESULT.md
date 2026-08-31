# FORMA S1 TENANT / IDENTITY / PRODUCT FOUNDATION RESULT

## Status

**PASS — S1 FROZEN**

S1 tenancy / identity / product foundation complete through S1-G2 Coze ID contract hardening.

| Item | Value |
|---|---|
| S1 feature | `a10d24035e36591ef37bcb4cfb28b73aaa786198` |
| **S1-G1** | `f3745fe7a7cbc9119dfee73fd535d50e90e3c443` |
| Forma CI (S1-G1) | [33378641009](https://github.com/hai12138/forma/actions/runs/33378641009) |
| **S1-G2** | `0f87d2ee6db26f0acb31244af6f5799306d9c965` |
| Forma CI (S1-G2) | [33380202709](https://github.com/hai12138/forma/actions/runs/33380202709) — backend / migration-apply / frontend **success** |
| S0 Frozen | `forma-s0-frozen` → `d68a49bf1ae780f71d6aecd3ff6d3eb3a1c7a3e6` |
| Coze Upstream Baseline | `fefb05ff27be1da939612fbf9faf5db62583b8ae` |

**S1 FROZEN.**  
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

## Files Changed (S1)

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

Tenant lifecycle: ACTIVE / SUSPENDED / ARCHIVED. Soft delete only. Optimistic `revision` on PATCH (required; no auto-fill).

---

## Principal Model

`FormaPrincipal` with stable `provider=coze` + `external_subject=coze_user_id` mapping. Types: USER / SERVICE.

---

## Membership

Roles: OWNER / ADMIN / MEMBER / VIEWER. Statuses: ACTIVE / INVITED / SUSPENDED / REMOVED. Server verifies membership for TenantContext.

### S1 Membership Role Policy (frozen)

| Actor | Can add/change |
|---|---|
| OWNER | OWNER, ADMIN, MEMBER, VIEWER; may transfer/grant OWNER |
| ADMIN | ADMIN, MEMBER, VIEWER only — **cannot** create/promote/modify OWNER |
| MEMBER / VIEWER | no membership management |

Last-owner + Primary Owner (`owner_principal_id`) invariants enforced in domain. V1 defers TransferOwnership API; Primary Owner OWNER role cannot be demoted.

---

## Tenant → Coze Space Mapping

`forma_tenant_space_ref` — **one row per `coze_space_id`** (S1-G1). Unbind → INACTIVE on same row; rebind → reactivate/update. History via audit. At most one ACTIVE owner tenant per space.

---

## TenantContext

Built server-side: Session → Principal → Membership → `X-Forma-Tenant` selection → AllowedSpaceIDs. Header is selection, not proof. AssetCounts fail-closed if TenantContext missing (no raw header fallback).

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

Forma Shell: real tenant switcher (sessionStorage selection key only — not production asset store), identity display, banners for unauthenticated/forbidden/suspended/empty/network. Overview shows real asset counts / empty state. Navigation IA unchanged (Forma v1.2 baseline).

---

## Migration

- `20250831120000_s1_tenancy.sql` — principal, tenant, membership, space_ref, audit_event
- `20250831140000_s1_g1_space_mapping.sql` — UNIQUE(`coze_space_id`); drop `(coze_space_id, status)` unique

Local portable apply CASE A/B/C: S0 → S1 → S1-G1.

---

## Audit

Skeleton events: TENANT_CREATED/UPDATED, MEMBER_*, SPACE_BOUND/UNBOUND via `forma_audit_event`.  
`principal_id` = **actor** (not target). RequestID from TenantContext when present.

---

## Security Tests

| Test | Result |
|---|---|
| Bootstrap idempotent | PASS |
| Tenant/Membership revision conflict | PASS |
| expected_revision required | PASS |
| Asset cross-tenant Get denied | PASS |
| Forged `X-Forma-Tenant` forbidden | PASS |
| Unauthenticated resolve | PASS |
| Admin cannot promote/add/modify OWNER | PASS |
| Member/Viewer cannot manage membership | PASS |
| Last owner cannot be demoted | PASS |
| Primary owner invariant | PASS |
| Space bind/unbind/rebind cycle | PASS |
| Audit actor correctness | PASS |

---

## CI

Extended Forma CI: S1 + S1-G1 tenancy tests, S1-G1 migration file check, pre-commit CWD check. S0 gates retained.

| Job | Prior (human-confirmed) | S1-G1 |
|---|---|---|
| forma-backend | PASS | PASS — run `33378641009` |
| forma-frontend | PASS | PASS — run `33378641009` |
| forma-migration-apply | PASS | PASS — run `33378641009` |

---

## Coze Core Files Modified

| File | Change |
|---|---|
| `application/application.go` | Forma tenancy init (FORMA blocks) |
| `api/router/register.go` | Forma router (S0) |
| `api/middleware/session.go` | Public meta + health paths |
| `bizpkg/config/config.go` | `InitBaseForLiveHarness` (FORMA) |
| `common/git-hooks/pre-commit` | CWD → coze-studio root |
| `rush.json` | Forma projects (S0) |

New middleware file: `api/middleware/forma_tenant.go` (Forma-owned).

---

## Remaining Mock

Placeholder module pages still show “not connected yet”. No mock KPI on overview.

---

## Known Limitations

1. TransferOwnership API deferred; Primary Owner demotion blocked instead.
2. UnbindSpace HTTP API not exposed in S1 (domain UnbindSpace + lifecycle tests cover mapping).
3. `.git/hooks/pre-commit` may need `scripts/forma/install-precommit.sh` on developer machines.
4. Large Coze space IDs exceed JS `Number.MAX_SAFE_INTEGER` when round-tripped via JSON number (domain uses int64).

---

## Gate Evidence

| Gate | Result | Evidence |
|---|---|---|
| GATE-01 Principal mapping | PASS | `ResolveOrCreatePrincipal` + Bootstrap tests |
| GATE-02 Multi-tenant membership | PASS | ListTenantsForPrincipal / memberships model |
| GATE-03 Multi-space binding | PASS | BindSpace + space_ref table/tests |
| GATE-04 Forged tenant header | PASS | unit + live E2E |
| GATE-05 Cross-tenant asset | PASS | `TestAssetRegistry_TenantIsolation_Get` |
| GATE-06 Suspended denial | PASS | unit + live E2E FORMA_TENANT_SUSPENDED |
| GATE-07 Bootstrap idempotent | PASS | `TestBootstrap_Idempotent` |
| GATE-08 Frontend identity | PASS | Session provider + shell switcher + build |
| GATE-09 S0→S1 migration | PASS | migration-apply-test CASE A/B/C + S1 tables |
| GATE-10 Forma CI | **PASS** | Human-confirmed: S1 `a10d240…` + docs `2de31f5…` — forma-backend / frontend / migration-apply all PASS |

---

## S1-G1 Identity Security Closure

| Item | Result | Notes |
|---|---|---|
| Owner/Admin Policy | PASS | `MembershipPolicy` CanAdd/CanChange/CanRemove |
| Last Owner Protection | PASS | domain invariant + tests |
| Primary Owner Consistency | PASS | demote primary OWNER blocked; TransferOwnership deferred |
| Space Lifecycle Fix | PASS | `uk_forma_space_id` + UpsertBind; cycle tests |
| Audit Actor Fix | PASS | actor ≠ target |
| Revision Enforcement | PASS | `expected_revision > 0` required; no auto-fill |
| TenantContext Fail-Closed | PASS | AssetCounts no raw `X-Forma-Tenant` fallback |
| Live Coze Session E2E | PASS | `scripts/forma/live-e2e.mjs` + `forma-live-harness` (real SessionAuthMW, passport register/login cookie) |
| Live Tenant Switch | PASS | tenant A/B → `/assets/counts` |
| Live Forged Tenant | PASS | 403 FORMA_* |
| Live Suspended Tenant | PASS | 403 FORMA_TENANT_SUSPENDED; OWNER can reactivate via PATCH |
| Live Space Validation | PASS | bootstrap binds personal space; inaccessible space denied |
| Browser Smoke | PASS | Edge headless: Forma Product Shell (not Coze); title Forma; Tenant switcher; identity loading/unauth path; `/api/forma` proxy → harness health 200. Screenshot: `forma/cursor-results/forma-shell-smoke.png` |
| CI (post-push) | **PASS** | [33378641009](https://github.com/hai12138/forma/actions/runs/33378641009) all jobs success |

### Live E2E harness

- `backend/cmd/forma-live-harness` — minimal real Coze passport + SessionAuthMW + Forma router
- Env: Docker `forma-live-mysql:3308`, `forma-live-redis:6380`, harness `:8888`
- Command: `FORMA_LIVE_E2E=1 FORMA_LIVE_BASE_URL=http://127.0.0.1:8888 node --test scripts/forma/live-e2e.mjs`

---

## S1-G2 Coze ID Contract Hardening

| Item | Result | Notes |
|---|---|---|
| Backend DTO | PASS | `coze_user_id` / `coze_space_id` / `default_space_id` → JSON **string** |
| TenantSpaceDTO / BootstrapResponse | PASS | Entities hide Coze IDs from JSON (`json:"-"`); public DTO only |
| `idcontract.Format/ParseCozeID` | PASS | Digits-only, >0, int64 overflow; reject float/sci/neg |
| Frontend `@forma/api-client` | PASS | `string` only; `FormaTenantSpace`; no dual number\|string |
| Precision tests (> JS safe int) | PASS | `9007199254740993`, `7563957783431741441` |
| BindSpace contract tests | PASS | string OK; JSON number / invalid strings rejected |
| Live E2E | PASS | `typeof coze_*_id === "string"`; BigInt round-trip; number body rejected |
| Frontend typecheck/build/vitest | PASS | `@forma/app` routes smoke |
| Architecture | PASS | `ID-CONTRACT.md` + `ADR-012` |
| CI | **PASS** | [33380202709](https://github.com/hai12138/forma/actions/runs/33380202709) |

---

## Risks

| Risk | Mitigation |
|---|---|
| JS number precision for Coze snowflake IDs | **Closed in S1-G2** — public IDs are strings (`ID-CONTRACT.md`) |
| SessionStorage for selected tenant | Selection UX only; server enforces membership |

---

## S1 Freeze

| Item | Value |
|---|---|
| Tag | `forma-s1-frozen` |
| S1 Frozen Commit | `4e923ba9e5f91673f91889c71c7687f76aa354ad` (`forma-s1-frozen`) |
| S1-G2 code | `0f87d2ee6db26f0acb31244af6f5799306d9c965` |
| S0 Frozen Commit | `d68a49bf1ae780f71d6aecd3ff6d3eb3a1c7a3e6` |
| Coze Upstream Baseline | `fefb05ff27be1da939612fbf9faf5db62583b8ae` |
| Forma CI (freeze baseline) | [33380202709](https://github.com/hai12138/forma/actions/runs/33380202709) |

---

## S2 Preconditions

1. Human sign-off on S1 freeze (`forma-s1-frozen`)
2. Forma CI green on frozen commit
3. Do not start Business Model / Analyst / Capability until explicitly authorized

---

**Stage:** FORMA-S1 / S1-G2  
**S1 FROZEN.**  
**DO NOT START S2.**
