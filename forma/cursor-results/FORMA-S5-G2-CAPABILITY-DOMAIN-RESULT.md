# FORMA-S5-G2 — Business Capability Domain Implementation
# RESULT

**Gate:** S5-G2
**Date:** 2026-09-12
**Status:** PENDING_CI (implementation committed; awaiting Forma CI ALL GREEN + human review)

---

## 1. Baseline

| Item | Value |
|------|-------|
| main tip at start | `ee2672db9b0298af8ba91aaffe7b430b0c82ce10` |
| S4 freeze | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| Scope | Capability domain + S5-G2 migration only |
| PRODUCT_CODE_CHANGE | CAPABILITY_DOMAIN_ONLY |
| MIGRATION_CHANGE | S5_G2_ONLY |
| REAL_MODEL_CALLS | 0 |

---

## 2. Deliverables

| Artifact | Action |
|----------|--------|
| `coze-studio/backend/domain/forma/capability/**` | CREATED (entity/service/repository/internal/dal/fixture) |
| Asset Registry `UpdateCapabilityProjection` | EXTENDED |
| `20250903010000_s5_g2_business_capability.sql` | CREATED |
| `atlas.sum` / forma-ci migration-present / migration-validate | UPDATED |
| `forma/cursor-results/FORMA-S5-G2-CAPABILITY-DOMAIN-RESULT.md` | CREATED |

No HTTP/router/frontend. No provider SDK. No `forma-s5-frozen`.

---

## 3. ContentDigest canonicalization (G2 frozen)

1. **CapabilityContentDigest** — behavioral fields only (`name`, `description`, `capability_kind`, `business_model_revision`, schemas, preconditions, effects, bindings, query_operation, output_cardinality). Excludes source/provenance/ids/status/version/audit. SHA-256 lowercase hex.
2. **DecisionPayloadDigest** — effective materialization/derive semantic body; excludes actor/`client_request_id`/timestamps.
3. **AnalysisRequestDigest** — model pin + sorted contract pins + sorted requirement refs + options.

Rules: JSON object keys rebuilt with sorted maps; set-like collections sorted by stable semantic key; ordered arrays preserved; `nil`/`[]`/`{}` normalized; identical behavior ⇒ identical ContentDigest across MANUAL/AI/DERIVED.

---

## 4. Local verification

| Check | Result |
|-------|--------|
| `go test ./domain/forma/capability/... -count=1` | PASS |
| `go test ./domain/forma/... -count=1` | PASS |
| `node scripts/forma/migration-validate.mjs` | PASS (15/15) |
| `git diff --check` | PASS |
| `migration-apply-test.mjs` CASE A/B/C | Deferred to Forma CI (`forma-migration-apply`) — local Docker daemon unavailable |
| REAL_MODEL_CALLS | 0 (DeterministicFakeGenerator only) |

---

## 5. Commit & CI

| Field | Value |
|-------|-------|
| COMMIT_SHA | `8ace5d9d77fac59acb964189bcb5bdffe64e9709` |
| CI_RUN | _pending_ |
| forma-backend | _pending_ |
| forma-migration-apply | _pending_ |
| forma-frontend | _pending_ |
| CI | _pending_ |

---

## 6. Gate Summary (pre-CI)

```text
S5_G2_STATUS = PENDING_CI
DOMAIN_MODEL = PASS
IMMUTABLE_REVISION = PASS
STATE_MACHINE = PASS
ASSETREF_PROJECTION = PASS
CONTENT_DIGEST = PASS
MANUAL_CREATE_ATOMICITY = PASS
DERIVE_IDEMPOTENCY = PASS
CONFIRM_IDEMPOTENCY = PASS
REJECT_ATOMICITY = PASS
ANALYSIS_RUN_IDEMPOTENCY = PASS
EXECUTION_FENCING = PASS
TENANT_ISOLATION = PASS
DOMAIN_AGNOSTIC = PASS
SECRET_ISOLATION = PASS
MIGRATION_CASE_A_B_C = PENDING_CI
PRODUCT_CODE_CHANGE = CAPABILITY_DOMAIN_ONLY
MIGRATION_CHANGE = S5_G2_ONLY
REAL_MODEL_CALLS = 0
CI = pending
S5_G3_READY = NO
```

**Stop:** After CI ALL GREEN, finalize §5/§6, then **STOP**. Do not start S5-G3. Do not create `forma-s5-frozen`.
