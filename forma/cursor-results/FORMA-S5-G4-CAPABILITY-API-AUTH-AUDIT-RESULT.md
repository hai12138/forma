# FORMA-S5-G4 — Capability API / Authorization / Audit / Concurrency Consistency
# RESULT

**Gate:** S5-G4  
**Date:** 2026-09-14  
**Status:** **PASS (local)** — Forma CI pending / updating after green

---

## 1. Baseline

| Item | Value |
|------|-------|
| APPROVED_BASELINE (main tip before G4) | `5004e4b1b4d4d76c8d5a346e44b60d8980581a33` |
| S5-G3-F3 implementation | `36c0078b922d628c583a02d6d14594e5ab76c5ea` |
| Scope | Capability API + Auth matrix §11.2 + Audit + concurrency/idempotency surface |
| PRODUCT_CODE_CHANGE | **YES** (API / application / error map / ListCapabilities domain) |
| MIGRATION_CHANGE | **NONE** |
| REAL_MODEL_CALLS | **0** (`DeterministicFakeGenerator` in `InitService`) |
| Capability runtime invoke | **NONE** |
| G5 UI / G6 / Agent / Workflow / Application / Runtime | **NONE** |

---

## 2. Implementation commit

| Field | Value |
|-------|-------|
| COMMIT_SHA (implementation) | `5c3d257ec88e0d0f6dd8467824cb511322449bd7` |
| Message | `feat(s5-g4): Capability API, authorization, audit, concurrency` |

---

## 3. API / DTO / route inventory

Thin handlers → `ApplicationService` under `/api/forma/v1` (+ `FormaTenantMW`):

| Method | Path | App method | Auth |
|--------|------|------------|------|
| GET | `/businesses/:id/capabilities` | `ListCapabilities` | Read |
| POST | `/businesses/:id/capabilities` | `CreateCapability` | Admin |
| POST | `/businesses/:id/capabilities/analyze` | `StartCapabilityAnalysis` | Admin |
| GET | `/businesses/:id/capabilities/:capabilityId` | `GetCapability` | Read |
| GET | `.../revisions` | `ListCapabilityRevisions` | Read |
| GET | `.../revisions/:revisionId` | `GetCapabilityRevision` | Read |
| POST | `.../derive` | `DeriveCapability` | Admin |
| POST | `.../edit` | `EditCapability` | Admin |
| GET | `.../decisions` | `ListCapabilityDecisions` | Read |
| GET | `/capability-analyses/:analysisRunId` | `GetCapabilityAnalysis` | Read |
| POST | `/capability-proposals/:proposalId/confirm` | `ConfirmCapabilityProposal` | Admin |
| POST | `.../edit-confirm` | `EditConfirmCapabilityProposal` | Admin |
| POST | `.../reject` | `RejectCapabilityProposal` | Admin |
| POST | `/capability-revisions/:revisionId/validate` | `ValidateCapabilityRevision` | Admin |
| POST | `.../activate` | `ActivateCapabilityRevision` | **Owner** |
| POST | `.../deprecate` | `DeprecateCapabilityRevision` | **Owner** |
| GET | `.../validations` | `ListCapabilityValidations` | Read |

**DTOs:** `CapabilityDTO`, `CapabilityRevisionDTO`, `CapabilitySemanticPayloadDTO`, `CapabilityDataContractBindingDTO` (logical bindings only: `data_contract_id` / revision / `logical_field_mappings`), proposal/analysis/decision/validation DTOs. No physical binding, JDBC, credentials, or secret fields.

**Domain extension (List API):** `ListCapabilitiesByBusiness` on DAO/repo/memory + `CapabilityService.ListCapabilities`.

**Error codes:** `FORMA_CAPABILITY_*` in `*80` range (40380/40480…/50080) — `*60` reserved by Data Contract.

---

## 4. Role matrix evidence (§11.2)

Enforcement: `loadCapabilityTenant` re-reads **ACTIVE** membership from Tenancy SoT; forged roles ignored.

| Action class | Allowed | Helper |
|--------------|---------|--------|
| List/Get (+ revisions/decisions/validations/analysis get) | OWNER, ADMIN, MEMBER, VIEWER | `requireCapabilityRead` |
| Create / Derive / Edit / Analyze / Confirm / EditConfirm / Reject / Validate | OWNER, ADMIN | `requireCapabilityAdmin` (`roleAtLeastAdmin`) |
| Activate / Deprecate | **OWNER only** | `requireCapabilityOwner` (`roleIsOwner`) |

**Tests:** `TestCapabilityApp_RoleMatrixAndIsolation` — OWNER/ADMIN create+validate; MEMBER/VIEWER read OK; MEMBER/VIEWER mutations + Confirm/EditConfirm/Reject denied; ADMIN Activate/Deprecate denied; OWNER Activate OK.

**Inactive membership:** `TestCapabilityApp_InactiveAndSuperAdminDenied` → `CodeTenantForbidden`.

**SUPER_ADMIN without tenant membership:** same test → `CodeTenantForbidden` (no auto tenant data access).

---

## 5. Tenant isolation evidence

- Business path: `requireCapabilityBusiness` via `BusinessSVC.Get(tenant, business)`.
- Capability / Revision / Proposal / Analysis scoped with `BusinessID == path business` else NotFound fail-closed.
- Cross-tenant Get → TenantForbidden or CapabilityNotFound (`TestCapabilityApp_RoleMatrixAndIsolation`).
- Domain cross-tenant refs remain fail-closed (G2/G3 regressions retained).

---

## 6. Audit / secret isolation evidence

- Mutating ops: `RecordAudit` actions `capability.create|derive|edit|analyze|confirm|edit_confirm|reject|validate|activate|deprecate`.
- Resource = safe id only; PrincipalID from tenancy; RequestID from tenant context.
- `TestCapabilityApp_AnalyzeConfirmIdempotencyAndAuditIsolation` asserts audit blob has no `password|token|authorization|cookie|secret|api_key|bearer`.
- `TestCapabilityApp_DTOExcludesSecrets` marshaled create response excludes credential/physical markers; includes logical `data_contract_id` / `logical_field_mappings`.
- Analysis DTO exposes stable `error_code` only (not raw underlying errors).

---

## 7. Concurrency / idempotency evidence

- Domain UoW / aggregate lock / active pointer / revision immutability unchanged; app is thin wrapper.
- Analyze + Confirm + Derive idempotent replay covered in `TestCapabilityApp_AnalyzeConfirmIdempotencyAndAuditIsolation`.
- Activate supersession + Deprecate clears active pointer: `TestCapabilityApp_ActivateDeprecateActivePointer`.
- Prior G2/G3 domain concurrency / evidence fence tests still pass under `./domain/forma/capability/...`.

---

## 8. Local verification

| Gate | Result |
|------|--------|
| `go test ./application/forma/ ./domain/forma/capability/... ./api/handler/forma/ ./domain/forma/errors/` | **PASS** |
| `go test ./domain/forma/... ./application/forma/... ./api/handler/forma/... ./crossdomain/forma/...` | **PASS** (middleware path N/A) |
| `node scripts/forma/migration-validate.mjs` | **19/19 PASS** |
| `migration-apply-test.mjs` CASE A/B/C | **SKIPPED locally** (Docker Desktop daemon not running) — deferred to CI `forma-migration-apply` |
| `node scripts/forma/typecheck.mjs` | **PASS** |
| `node scripts/forma/routes-smoke.mjs` | **PASS** |
| Rush `@forma/app` build/tests | **SKIPPED locally** (Windows bash/`rtsc.sh` toolchain) — deferred to CI `forma-frontend` |
| `git diff --check` | **PASS** (whitespace warnings only) |

---

## 9. CI

| Field | Value |
|-------|-------|
| CI_RUN | _pending after push_ |
| forma-backend | _pending_ |
| forma-migration-apply | _pending_ |
| forma-frontend | _pending_ |
| CI | _pending_ |

---

## 10. Gate summary

```text
S5_G4_STATUS = PASS_LOCAL
PRODUCT_CODE_CHANGE = YES
MIGRATION_CHANGE = NONE
REAL_MODEL_CALLS = 0
ROLE_MATRIX_ENFORCED = PASS
TENANT_ISOLATION = PASS
AUDIT_SECRET_ISOLATION = PASS
CONCURRENCY_IDEMPOTENCY = PASS
RUNTIME_INVOKE = NONE
S5_G5_READY = NO
```

**Stop:** Do **not** create `forma-s5-frozen`. Do **not** announce G5. Await human review after CI ALL GREEN.
