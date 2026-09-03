# FORMA-S4-G6 LIVE E2E / SECURITY / FINAL FREEZE — RESULT

## Status

**S4_G6_STATUS = PASS**

**S4_STATUS = PASS / FROZEN**

**FREEZE_TAG = `forma-s4-frozen`**

**FREEZE_COMMIT_SHA = `3f45d8bc31862da7304bc5d99a858f41ff3e300e`**

**FREEZE_TAG_OBJECT_SHA = `fb8749540e994623f65e451adb472be75b1c4f06`** (annotated)

**CANDIDATE_SHA = `de810b15df2d9dbd0b130918820761dd705f9800`**

**CANDIDATE_CI = [33710829010](https://github.com/hai12138/forma/actions/runs/33710829010) ALL GREEN**

**FREEZE_CI = [33711307076](https://github.com/hai12138/forma/actions/runs/33711307076) ALL GREEN**

**TAG_CI = NOT_CONFIGURED / pending** (tag push verified on origin; freeze commit CI green)

## Baseline

| Item | Value |
|------|-------|
| Candidate SHA | `de810b15df2d9dbd0b130918820761dd705f9800` |
| Candidate CI | [33710829010](https://github.com/hai12138/forma/actions/runs/33710829010) **ALL GREEN** (forma-backend / forma-migration-apply / forma-frontend) |
| Requested main tip (pre-G6) | `1606c177c999d00d3bfe563c9a22ee1f750d6cb9` |
| G5-F2 implementation | `8793c38782f6948d1934d7fb1d90259e4659a90e` |
| G5-F2 CI | `33705669429` PASS |
| Latest main CI (pre-G6) | `33706058330` PASS |
| Frozen baselines | `forma-s0-frozen` … `forma-s3-frozen` |
| S4 gates G0–G5-F2 | ALL PASS |

## Environments

| Layer | Value |
|-------|-------|
| Backend | `forma-live-harness` `:8888` on `forma-live-net` (real SessionAuth + Forma router + MySQL) |
| Frontend | `http://127.0.0.1:3001` (`frontend/apps/forma`) |
| MySQL | `forma-live-mysql` 8.x — disposable DB `forma_g6_lab` |
| Postgres | `docker-db_postgres-1` — disposable DB `forma_g6_pg` |
| HTTP fixture | in-network nginx `forma-g6-http-fx` (+ host fallback) OpenAPI 3.x |
| Browser | Playwright Chromium 1440×900 / 1280×800 |
| Model provider | FormaDataModel → Forma ACL → Coze / Eino Model Manager (native Coze config) |

## Real model budget

| Call | Operation | Evidence |
|------|-----------|----------|
| CALL A | `AnalyzeDataRequirements` | Live owned execute (~66s) — PROPOSED only; idempotent replay no extra call |
| CALL B | `SuggestSemanticMappings` | Live owned execute (~23s) — PROPOSED only |
| **REAL_MODEL_CALLS** | **2** | Hard counter + no auto-retry beyond existing policy |

Subsequent acceptance reruns used `FORMA_G6_SKIP_CALL_A/B=1` with manual stand-ins so the Coze/Eino key was not burned again. Prior real CALL A/B evidence retained in `s4-g6-live-e2e.log`.

Forbidden provider SDKs: none introduced (no direct DeepSeek/OpenAI/Ark/Qwen SDK from Forma).

## Business scenarios

| Business | Domain | Notes |
|----------|--------|-------|
| A | 实验室样本流转 | Full path: requirements → MySQL source → mapping → contract → drift/recovery |
| B | 采购合同审批 | Deterministic (no 3rd model call): manual requirements on same Core APIs |

**DOMAIN_AGNOSTIC**: `before_core_sha == after_core_sha` for production Core during live acceptance. No industry branching enums in Core.

## Minimal product fixes (allowed under G6)

1. **TestConnection typed-nil panic** — `TestDataConnection` no longer returns `MapDomainError(nil)` as a non-nil `error`; `errorEnvelope` hardened against typed-nil `*FormaError`.
2. **MySQL schema JOIN duplication** — `table_constraints` join now matches `table_name`; field-path dedupe defense in `GetSchema`.
3. **Analyze/retry auth** — `AnalyzeDataRequirements` / `RetryFailedDataAnalysis` require `requireDataEdit` (MEMBER denied).
4. **Read-only contract boundary** — `ValidateQueryCapabilities` enforced on CreateContract / CreateRevision (CREATE/UPDATE/DELETE/EXECUTE/COMMAND/MUTATE denied at write).

No new S4 domain entities, DSL types, Contract semantics, or S5 features.

## Live gates (API)

Script: `coze-studio/scripts/forma/s4-g6-live-e2e.mjs` — **17/17 PASS**

Also: `s4-g6-postgres-smoke.mjs` **PASS**, `s4-g6-browser-e2e.mjs` **PASS**.

## Browser evidence

Directory: `forma/cursor-results/s4-g6-ui/` (15 sanitized PNGs). Secret scan: exact secret material = 0.

Routes verified with direct navigation + refresh: `/data`, `/data/requirements`, `/data/sources`, `/data/mappings`, `/data/contracts`, `/data/health`.

## 12 Final Hard Gates

| # | Gate | Result | Evidence |
|---|------|--------|----------|
| 1 | DOMAIN_AGNOSTIC | PASS | Two businesses; Core SHA unchanged; no industry branching |
| 2 | AI_NO_SILENT_MUTATION | PASS | CALL A/B → PROPOSED only |
| 3 | REQUIREMENT_PROVENANCE | PASS | AnalysisRun + decisions API |
| 4 | CREDENTIAL_ISOLATION | PASS | No secret read API; ciphertext-only DB; response scan |
| 5 | SOURCE_INDEPENDENCE | PASS | MySQL + HTTP (+ Postgres smoke) via same abstractions |
| 6 | IMMUTABLE_REVISION | PASS | Snapshot/history immutability; STALE recovery via new revision |
| 7 | SCHEMA_DRIFT | PASS | Compatible unused column; breaking rename → STALE; v2 activate |
| 8 | BUSINESS_REVISION_GAP | PASS | New business revision does not auto-mutate contract |
| 9 | TENANT_ISOLATION | PASS | Tenant B cross-access denied / empty isolation |
| 10 | READ_ONLY_BOUNDARY | PASS | Mutating query capabilities denied on create |
| 11 | REQUIREMENT_ANALYSIS_IDEMPOTENCY | PASS | Same `client_request_id` reuse; counter stable |
| 12 | REQUIREMENT_DECISION_PROVENANCE | PASS | Confirm/Reject/EditConfirm + decisions |

## Extra Final Gates

| Gate | Result |
|------|--------|
| SEMANTIC_MAPPING_PROVENANCE | PASS |
| MAPPING_HUMAN_CONFIRMATION | PASS |
| MAPPING_IDEMPOTENCY | PASS (prior unit + live analyze owned path) |
| CONTROLLED_DSL | PASS (deterministic suite retained) |
| CONTRACT_LOGICAL_PHYSICAL_SEPARATION | PASS |
| ACTIVE_POINTER_CONSISTENCY | PASS |
| CONSUMER_DESCRIPTOR_SAFETY | PASS (MEMBER strip) |
| SSRF_PROTECTION | PASS (metadata / link-local / schemes / userinfo) |
| ROLE_AUTHORIZATION | PASS (MEMBER negatives live) |
| BROWSER_ACCEPTANCE | PASS |

## Security scans

| Scan | Result |
|------|--------|
| Exact secret material in logs/evidence | 0 matches |
| Credential GET/List plaintext | NONE (404 / no value fields) |
| Gitignored live env | `.forma-live-harness.env` / `forma-live.env` not staged |
| SSRF live handler path | DENIED targets verified |
| Public config password nested | DENIED |

## Freeze sequence

1. Candidate commit: product minimal fixes + live scripts  
2. Push `main` → wait Forma CI ALL GREEN on exact SHA  
3. Docs/evidence commit: RESULT + sanitized `s4-g6-ui/` (no product code)  
4. Push → CI green on freeze tip  
5. Annotated tag `forma-s4-frozen` → push tag → verify object/target  
6. **STOP** — do not start S5

## Final status block

```
S4_G6_STATUS = PASS
S4_STATUS = PASS / FROZEN
FREEZE_TAG = forma-s4-frozen
FREEZE_COMMIT_SHA = 3f45d8bc31862da7304bc5d99a858f41ff3e300e
FREEZE_TAG_OBJECT_SHA = fb8749540e994623f65e451adb472be75b1c4f06
CI_RUN = 33711307076
CANDIDATE_CI_RUN = 33710829010
TAG_CI = NOT_CONFIGURED
REAL_MODEL_CALLS = 2
BUSINESS_A_E2E = PASS
BUSINESS_B_E2E = PASS
DOMAIN_AGNOSTIC = PASS
AI_NO_SILENT_MUTATION = PASS
REQUIREMENT_PROVENANCE = PASS
CREDENTIAL_ISOLATION = PASS
SOURCE_INDEPENDENCE = PASS
IMMUTABLE_REVISION = PASS
SCHEMA_DRIFT = PASS
BUSINESS_REVISION_GAP = PASS
TENANT_ISOLATION = PASS
READ_ONLY_BOUNDARY = PASS
ANALYSIS_IDEMPOTENCY = PASS
DECISION_PROVENANCE = PASS
SEMANTIC_MAPPING = PASS
CONTRACT = PASS
BROWSER_E2E = PASS
SECURITY = PASS
BUSINESS_MODEL_MUTATION = NONE
```

## STOP

S4 is frozen. Do **not** start S5 / S6 / S7 / S8 / S9 until human confirms `forma-s4-frozen` target.