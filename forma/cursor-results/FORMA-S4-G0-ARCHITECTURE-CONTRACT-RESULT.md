# FORMA-S4-G0 ARCHITECTURE CONTRACT RESULT

## Status

**PASS / CONTRACT_READY**

Awaiting **human architecture review**.  
**DO NOT START S4-G1.**  
**DO NOT create `forma-s4-frozen`.**

---

## Baseline

| Item | Value |
|------|-------|
| S0 | PASS / FROZEN |
| S1 | PASS / FROZEN |
| S2 | PASS / FROZEN |
| S3 | PASS / FROZEN |
| S3 immutable tag | `forma-s3-frozen` |
| S3 freeze target commit | `a13fbfc606f873be8e0f88e30baa1709ff32c9dd` |
| Main tip at G0 authoring | `50a77cd4391542502064d03e9585cc19951aca01` (docs-only freeze record after S3 freeze) |

S4 is defined strictly on the S3 frozen baseline. No S0–S3 architecture mutation.

---

## S4 Scope (locked)

**S4 — Data Plane / Data Contract**

Turn confirmed Business Model into stable, secure, versioned, auditable, domain-agnostic data access contracts.

Core chain:

Business Model → Data Requirement Analysis → Proposal → Human Confirm → Confirmed Requirements → Data Source → Discovery → Semantic Mapping → Human Confirm → Data Contract → Immutable Revision → (future) S5 Capability

---

## Architecture Decisions

| Decision | Result |
|----------|--------|
| Domain-agnostic invariant | **LOCKED** — no industry objects in Core |
| Business Model SoT | **PRESERVED** — S4 refers; never replaces |
| AI propose / human confirm | **LOCKED** — same as S3 |
| Planned domain package | `backend/domain/forma/data/` (entity/service/repository/internal/dal) |
| CrossDomain / ACL | Reuse `crossdomain/forma` + `integration` pattern |
| Model boundary | `FormaDataModel` via CrossDomain → Coze/Eino (no provider SDK in domain) |
| Migrations ownership | Forma Atlas path only (`docker/atlas/forma/migrations/`) — none in G0 |
| Frontend IA | Module name **Data**; existing `/data` placeholder reserved for G5 |
| DataSource vs Connection | Separated (1:N); env-scoped connections |
| Secrets | `CredentialRef` + `SecretProvider`; plaintext forbidden everywhere listed in contract |
| SchemaSnapshot | Required; ACTIVE contract must not silently rediscover |
| Mapping DSL | Controlled types only; no arbitrary executable expressions |
| Contract revisions | **IMMUTABLE** |
| S4 V1 operations | READ / LOOKUP / LIST / FILTER / AGGREGATE only |
| S4↔S5 boundary | Capabilities depend on Contract ID + Version + Logical Schema |
| Cost policy | ≥95% fixtures; G0 real model calls = **0** |
| Gate roadmap | G0→G1→G2→G3→G4→G5→G6; one gate at a time |

---

## Canonical Assets Defined

- DataRequirement
- DataSource
- DataConnection
- CredentialRef
- DataAsset
- SchemaSnapshot
- SemanticMapping
- DataContract
- DataContractRevision
- DataValidationResult
- DataDriftResult

---

## Forbidden in G0 (and verified)

| Forbidden | G0 status |
|-----------|-----------|
| S4 business/domain Go code | **NOT ADDED** |
| Database migration | **NOT ADDED** |
| Real data source adapter | **NOT ADDED** |
| Secret provider implementation | **NOT ADDED** |
| Real LLM / model calls | **0** |
| Modify S0–S3 frozen architecture | **NOT DONE** |
| Create `forma-s4-frozen` | **NOT DONE** |
| Enter S4-G1 | **BLOCKED** |

---

## Repo Review Notes (no refactor)

Observed conventions preserved for later gates:

1. Domain layout matches analyst/business/tenancy packages.
2. `arch_test.go` forbids domain → Coze agent/user repository imports.
3. Analyst model ACL (`FormaAnalystModel` + `CozeEinoAnalystModel`) is the template for `FormaDataModel`.
4. Industry fixture `business/fixture/work_order.go` remains fixture-only (allowed under domain-agnostic invariant).
5. Frontend already has `/data` PlaceholderPage titled “数据平面”.
6. Migrations live under Forma Atlas; Coze core tables untouched.

---

## Hard Gates (preview locked)

1. Domain Agnostic  
2. AI No Silent Mutation  
3. Requirement Provenance  
4. Credential Isolation  
5. Source Independence  
6. Immutable Revision  
7. Schema Drift → STALE  
8. Business Revision Gap  
9. Tenant Isolation  
10. Read-only Boundary  

---

## Artifacts

| File | Role |
|------|------|
| `forma/docs/stages/FORMA-S4-DATA-PLANE-DATA-CONTRACT-STAGE-CONTRACT.md` | Formal Stage Contract |
| `forma/cursor-results/FORMA-S4-G0-ARCHITECTURE-CONTRACT-RESULT.md` | This G0 result |

---

## Verification

| Check | Result |
|-------|--------|
| Documentation only | **YES** |
| REAL_MODEL_CALLS | **0** |
| S3 Live E2E re-run | **NOT RUN** (forbidden) |
| Forma CI (this commit) | *(recorded after push)* |

---

## Final Output Checklist

```
S4_G0_STATUS = PASS / CONTRACT_READY
BASELINE_S3_TAG = forma-s3-frozen
S4_DOMAIN_AGNOSTIC = LOCKED
BUSINESS_MODEL_SOT = PRESERVED
AI_PROPOSE_HUMAN_CONFIRM = LOCKED
DATA_REQUIREMENT = DEFINED
DATA_SOURCE = DEFINED
DATA_ASSET = DEFINED
SCHEMA_SNAPSHOT = DEFINED
SEMANTIC_MAPPING = DEFINED
DATA_CONTRACT = DEFINED
CONTRACT_REVISION = IMMUTABLE
CREDENTIAL_ISOLATION = LOCKED
S4_READ_ONLY = LOCKED
S4_S5_BOUNDARY = LOCKED
REAL_MODEL_CALLS = 0
```

**STOP after CI green. Wait for human architecture review before S4-G1.**
