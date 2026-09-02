# FORMA-S4 — Data Plane / Data Contract
# STAGE CONTRACT (Architecture Freeze)

**Gate:** S4-G0 — Architecture & Stage Contract Freeze  
**Amendment:** S4-G0-F1 — Data Requirement Audit / Idempotency Contract  
**Status:** CONTRACT_READY (awaiting final human architecture review after G0-F1)  
**Baseline:** `forma-s3-frozen` → `a13fbfc606f873be8e0f88e30baa1709ff32c9dd`  
**G0 contract commit:** `c6057ce059ecd3fe4fdf9940889916a76864302a`  
**Scope of this document:** Architecture invariants, domain model, boundaries, gates, and G1–G6 roadmap.  
**Code change rule for G0 / G0-F1:** Documentation only. No domain code, migrations, adapters, APIs, or model calls.

---

## 1. Product Definition

### 1.1 Official Name

**S4 — Data Plane / Data Contract**

### 1.2 Responsibility

Convert a confirmed **Business Model** into a stable, secure, versioned, auditable, **domain-agnostic** data-access contract.

### 1.3 Core Chain

```
Business Model
    ↓
Data Requirement Analysis
    ↓
Data Requirement Proposal
    ↓
Human Review / Confirm / Edit / Reject
    ↓
Confirmed Data Requirements
    ↓
Data Source
    ↓
Data Asset / Schema Discovery
    ↓
Semantic Mapping Proposal
    ↓
Human Review / Confirm
    ↓
Data Contract
    ↓
Data Contract Revision
    ↓
Future S5 Business Capability
```

### 1.4 S4 Is Not

- Database administration platform
- ETL platform
- BI / analytics platform
- Data warehouse
- Agent Runtime
- Workflow Builder

---

## 2. Baseline & Preservation

| Stage | Status | Tag / Target |
|-------|--------|--------------|
| S0 | PASS / FROZEN | `forma-s0-frozen` |
| S1 | PASS / FROZEN | `forma-s1-frozen` |
| S2 | PASS / FROZEN | `forma-s2-frozen` |
| S3 | PASS / FROZEN | `forma-s3-frozen` → `a13fbfc606f873be8e0f88e30baa1709ff32c9dd` |

S4 must build on the above frozen baseline.

**Forbidden in S4:**

- Modify S0–S3 frozen architecture semantics
- Create `forma-s4-frozen` before G6 + human freeze
- Enter S4-G1 until this Stage Contract is human-approved

---

## 3. Architecture Invariants

### 3.1 FORMA-DOMAIN-AGNOSTIC-INVARIANT

Forma Core **must not** hard-code industry-specific business objects.

**Forbidden as platform-level concepts** (non-exhaustive):

`work_order`, `repair`, `refund`, `energy`, `device`, 招商, 客服, 维修, 工单, 审批单, 订单, 园区

These may appear **only** in:

- E2E fixtures
- Demo / example data
- Test data
- Industry templates (outside Core)

They **must not** enter:

- Forma Core Domain
- Generic table schemas
- Generic API contracts
- Generic capabilities
- Platform enums

Business entities, processes, states, roles, rules, and events originate from:

```
User Description → AI Proposal → Human Confirmation → Business Model
```

**Generality standard:** Onboarding a never-before-supported industry business requires:

```
Forma Core Code Change = 0
```

### 3.2 Business Model Ownership (SoT)

**Business Model remains the Canonical Business Semantic Source of Truth.**

S4 must **REFER TO** Business Model. S4 must **never REPLACE** Business Model.

Example (conceptual only):

- Business Model: “order amount above threshold requires manager approval”
- S4 may derive requirements for amount / approver / approval status
- Data Contract is **not** the SoT for that business rule

### 3.3 AI Propose / Human Confirm

Same principle as S3:

| Actor | Allowed | Forbidden |
|-------|---------|-----------|
| AI | Propose requirements / mappings | Confirm, activate, alter ACTIVE contract |
| Human | Confirm / Edit / Reject / Activate | — |

### 3.4 Credential Isolation

Secret values must never enter Git, DB plaintext, Frontend, browser storage, URL, LLM prompt/context, application/audit logs, error responses, test evidence, or screenshots.

### 3.5 Read-Only Boundary (S4 V1)

Data Contracts support **READ / LOOKUP / LIST / FILTER / AGGREGATE** only.  
Business write/actions (approve, close, submit, create, execute) belong to **S5 Business Capability**.

---

## 4. Repository Conventions (Actual Layout)

S4 packages **follow existing Forma conventions** (do not invent a parallel layout):

| Concern | Path |
|---------|------|
| Domain | `coze-studio/backend/domain/forma/data/` with `entity/`, `service/`, `repository/`, `internal/dal/` |
| Cross-domain ACL | `coze-studio/backend/crossdomain/forma/` + `integration/` |
| Model adapter pattern | Domain interface (e.g. `FormaDataModel`) implemented under `crossdomain/forma/integration` (same pattern as `FormaAnalystModel`) |
| Migrations | `coze-studio/docker/atlas/forma/migrations/` (Forma-owned tables only) |
| Frontend shell route | Existing `/data` placeholder (“数据平面”) — S4-G5 replaces placeholder with real module |
| Architecture guard | `domain/forma/arch_test.go` — domain must not import Coze agent/user repositories |

**Referenced but not rebuilt in S4:**

`Tenant`, `Principal`, `Role`, `Permission`, `Business`, `BusinessModelRevision`

---

## 5. Core Domain Model (Canonical Assets)

| # | Asset | Role |
|---|-------|------|
| 1 | DataRequirement | What the business needs to run |
| 2 | DataSource | Logical data origin |
| 3 | DataConnection | Environment-scoped connection to a source |
| 4 | CredentialRef | Opaque reference to secrets (not the secret) |
| 5 | DataAsset | Discoverable physical object abstracted as Forma asset |
| 6 | SchemaSnapshot | Immutable captured physical schema |
| 7 | SemanticMapping | Requirement/Business semantic ↔ physical schema |
| 8 | DataContract | Stable logical data interface for future capabilities |
| 9 | DataContractRevision | Immutable version of a contract |
| 10 | DataValidationResult | Technical validation outcome |
| 11 | DataDriftResult | Drift detection outcome |

**G0-F1 clarification:** The 11 assets above remain the only S4 **Business / Data Semantic Canonical Assets**.

The following are **Operational / Workflow / Audit Records** (not new semantic SoT assets):

- `DataRequirementAnalysisRun`
- `DataRequirementDecision`

They must not redefine Business Model semantics.

---

## 6. DataRequirement

Expresses **“what data the business needs”**, not “what columns currently exist”.

### 6.1 `requirement_kind` (generic only)

`ENTITY` · `ATTRIBUTE` · `RELATION` · `EVENT` · `METRIC` · `STATE` · `TIME_SERIES` · `DOCUMENT` · `LOOKUP` · `HISTORY`

No industry enums.

### 6.2 Suggested fields

`requirement_id`, `tenant_id`, `business_id`, `business_model_revision`, `requirement_kind`, `semantic_name`, `description`, `business_element_refs[]`, `requiredness`, `freshness_requirement`, `access_need`, `status`, `source`, `derived_from_requirement_id`, `analysis_run_id` (nullable), `created_by`, `created_at`, `updated_at`

- `AI_GENERATED` requirements **must** reference `analysis_run_id`
- `MANUAL_CREATED` requirements may have `analysis_run_id = null`

### 6.3 Status / Source

| status | meaning |
|--------|---------|
| PROPOSED | AI or draft |
| CONFIRMED | Human confirmed |
| REJECTED | Human rejected |
| SUPERSEDED | Replaced by a later requirement |

| source | meaning |
|--------|---------|
| AI_GENERATED | Produced by AI |
| MANUAL_CREATED | Human created |
| MANUAL_MODIFIED | Human edited replacement (Edit & Confirm lineage) |

**AI must never produce `CONFIRMED`.**

### 6.4 Semantic immutability (G0-F1)

After a DataRequirement is created, these core semantic fields **must not be overwritten in place**:

`business_model_revision`, `requirement_kind`, `semantic_name`, `description`, `business_element_refs`, `source`, `derived_from_requirement_id`, `analysis_run_id`

Human edit requires a **replacement** requirement (see §6.7).  
`status` may change only via the legal state machine / decision actions.

### 6.5 Provenance

Every requirement must answer: **“Why is this data needed?”**

```
BusinessModelRevision
  → DataRequirementAnalysisRun
  → AI_GENERATED DataRequirement
  → DataRequirementDecision
  → CONFIRMED / MANUAL_MODIFIED / REJECTED DataRequirement
```

S4 does **not** duplicate S3 Evidence. Upstream Business Model provenance (Proposal → Assertion → Evidence → User Turn) remains in S2/S3.

### 6.6 DataRequirementAnalysisRun (workflow / audit)

Not a Business Semantic SoT. Binds one analysis execution:

```
Business Model Revision
  → AnalyzeDataRequirements
  → N DataRequirements
```

Suggested fields:

`analysis_run_id`, `tenant_id`, `business_id`, `business_model_revision`, `client_request_id`, `request_digest`, `status`, `model_ref`, `error_key`, `error_message_sanitized`, `created_by`, `started_at`, `completed_at`, `created_at`

Status (G0-F1 minimum): `PENDING` · `SUCCEEDED` · `FAILED`  
Do not invent extra statuses unless a later gate proves need.

#### Idempotency (LOCKED)

Same logical key:

`tenant_id` + `business_id` + `business_model_revision` + `client_request_id`

→ **exactly one** logical Analysis Run.

- Duplicate requests **must return the existing result**
- Must **not** re-invoke the model
- Concurrent identical `client_request_id`: **only one** execution may acquire the run

Must prevent double-click / network retry / browser retry / concurrent request from causing duplicate model calls or duplicate DataRequirements.

#### Retry contract

`FAILED` AnalysisRun may be **explicitly** retried (auditable). Implementation deferred to G1.

**Forbidden:** HTTP/client retry → unconditional new model call / silent second logical request.

#### Security

AnalysisRun must not store raw credentials, API keys, secrets, Authorization headers, or cookies.  
`error_message_sanitized` only. Credential secret values must never enter AI input.

#### Domain-agnostic

AnalysisRun / Decision records must not embed industry fields/enums (工单 / 退款 / 能源 / 设备 / 审批 / 订单, …). Only generic revision / requirement / decision / principal / run metadata.

### 6.7 DataRequirementDecision (immutable audit event)

Not a Business Semantic SoT. Records **human** decisions on requirements.

Actions: `CONFIRM` · `REJECT` · `EDIT_CONFIRM`

Suggested fields:

`decision_id`, `tenant_id`, `business_id`, `source_requirement_id`, `target_requirement_id` (nullable), `action`, `actor_principal_id`, `reason` (nullable), `business_model_revision`, `created_at`

Decisions are **immutable** after create.  
AI / Model / Tool **must not** create Human Decisions.  
Permissions reuse Forma identity (S1).

| Action | Effect |
|--------|--------|
| **CONFIRM** | `PROPOSED` → `CONFIRMED`; create Decision(`CONFIRM`) |
| **REJECT** | `PROPOSED` → `REJECTED`; create Decision(`REJECT`); reason optional. **Forbidden:** physical delete of AI proposal as Reject |
| **EDIT_CONFIRM** | Original → `SUPERSEDED`; new requirement `source=MANUAL_MODIFIED`, `derived_from_requirement_id=original`, `status=CONFIRMED`; Decision(`EDIT_CONFIRM`) with `source_requirement_id=original`, `target_requirement_id=new`. Preserve original AI content, Business Model refs, and AnalysisRun provenance |

**Forbidden:** overwrite AI_GENERATED semantic content in place and keep the same record as “edited”.

---

## 7. AI Data Analyst Contract

Abstract interface (name reserved for G1+):

**`FormaDataModel`**

Responsibilities only:

- `AnalyzeDataRequirements`
- `SuggestSemanticMappings`

Call path:

```
Forma Data Domain
  → Forma CrossDomain / ACL
  → Coze/Eino Model Manager
```

**Forbidden in Forma Data Domain:** OpenAI / DeepSeek / Qwen / any provider-specific SDK.

---

## 8. DataSource / DataConnection / Credential

### 8.1 DataSource

Logical origin. Architecture-level `source_type`:

`RELATIONAL_DATABASE` · `HTTP_API` · `FILE_STORAGE` · `OBJECT_STORAGE` · `MANAGED_DATA` · `EVENT_STREAM` · `CUSTOM_ADAPTER`

**S4 V1 planned implementations:** `RELATIONAL_DATABASE` (MySQL, PostgreSQL) and `HTTP_API` (REST/JSON; OpenAPI discovery when available). Others are extension points only.

### 8.2 Separation

```
DataSource (1) ──< (N) DataConnection
```

Connections carry environment (DEV / TEST / PROD). Contracts must not depend on raw host strings.

### 8.3 CredentialRef + SecretProvider

Domain stores only `CredentialRef`.

```
DataConnection → CredentialRef → SecretProvider
```

V1 planned: `LocalEncryptedSecretProvider` (AES-GCM; master key `FORMA_SECRET_MASTER_KEY` in runtime env only).  
Must be replaceable by Vault / Cloud KMS / Secrets Manager **without changing Data Domain**.

---

## 9. DataAsset / Schema Discovery / SchemaSnapshot

### 9.1 DataAsset

Unified abstractions: `DATASET` · `ENTITY_SET` · `ENDPOINT` · `DOCUMENT_SET` · `STREAM`

Domain must **not** define `MySQLTable` / `PostgresTable` / `RestEndpoint` as core types — those belong to adapters.

### 9.2 DataSourceAdapter (interface)

`TestConnection` · `DiscoverAssets` · `GetSchema` · `Preview` · `ValidateContract`

Implementations (later gates): `MySQLAdapter`, `PostgresAdapter`, `HTTPAdapter`.  
Future adapters (Oracle, SQL Server, Mongo, Kafka, …) require **Forma Data Domain code change = 0**.

### 9.3 SchemaSnapshot

Contracts must not permanently depend on live rediscovery.

`DataContractRevision` references a fixed `SchemaSnapshot` (`fingerprint`, `schema_json`, `captured_at`, …).

**Forbidden:** rediscover → silently mutate ACTIVE contract.

---

## 10. Semantic Mapping & Controlled DSL

Maps business/data requirement semantics to physical schema.

AI may propose mappings with `confidence`, `reason`, evidence.  
**`confidence ≠ confirmation`.** No auto-confirm.

### Controlled Mapping DSL (S4 V1)

Allowed: `DIRECT` · `CAST` · `ENUM_MAP` · `UNIT_CONVERT` · `TIME_NORMALIZE` · `FIELD_PATH` · `JOIN_REF`

**Forbidden:** arbitrary JavaScript / Python / SQL fragment / Shell / executable expressions.

Transforms must be typed, validated, serializable, auditable.

---

## 11. DataContract

Stable logical data interface for future Business Capabilities.

S5/S6/S7 **must not** depend on table/column/endpoint/JSON field names.  
They depend on: **DataContract ID + Version + Logical Schema**.

### 11.1 Suggested structure

`contract_id`, `tenant_id`, `business_id`, `name`, `description`, `business_model_revision`, `requirement_ids[]`, `logical_schema`, `query_capabilities`, `filter_schema`, `sort_schema`, `pagination_policy`, `freshness_policy`, `classification_policy`, `access_policy_ref`, `source_binding`, `mapping_revision`, `schema_snapshot_id`, `status`, `version`, `created_by`, `created_at`

### 11.2 Lifecycle

| status | meaning |
|--------|---------|
| DRAFT | Editable |
| VALIDATED | Structure + mapping technically validated |
| ACTIVE | Capabilities may depend |
| STALE | Physical change broke guarantees |
| DEPRECATED | Human-retired |

**ACTIVE contracts must not be auto-modified by Discover or AI.**

### 11.3 DataContractRevision

**Immutable.** No in-place update of an existing version.  
Capabilities pin explicit versions (e.g. Contract X v2).

### 11.4 Schema Drift

Breaking changes (required field deleted, incompatible type, missing path) → contract **STALE**. No auto-repair of ACTIVE.

Compatible changes may record warnings.

### 11.5 Business Model Revision Gap

Business Model `rN → rN+1` **must not** silently mutate contracts.

Flow:

```
Business Revision Analysis
  → New Data Requirement Proposal / Gap
  → Human Review
  → Contract Revision Proposal
  → New Contract Version
```

---

## 12. Access Control & Classification

### 12.1 Identity

Reuse S1: Tenant, Principal, Role, Permission.  
All requests are tenant-scoped.

Data Contract access policy direction: tenant / role / principal / operation / field / row scope.

### 12.2 Classification

`PUBLIC` · `INTERNAL` · `CONFIDENTIAL` · `PII` · `SECRET`

- **SECRET:** never enters LLM
- **PII:** preview/sample default masking

---

## 13. Database Ownership

Unchanged from S0–S3:

- Forma-owned tables via Forma Atlas migrations
- No Coze core table modification
- No Forma columns on Coze tables
- No hard FK into Coze tables
- Coze resources linked only via mapping/ref (`forma_coze_resource_ref` pattern)

---

## 14. UI Information Architecture

Primary module name: **Data** (not “Database”).

Planned pages:

- Data Overview
- Data Requirements
- Data Sources
- Schema Explorer
- Mapping Studio
- Data Contracts
- Data Drift

Product flow:

```
Business Model → Analyze Data Needs → Review Requirements → Connect Source
  → Discover → Review Semantic Mapping → Preview Contract → Validate → Activate
```

### Mapping Studio (three columns)

| LEFT | CENTER | RIGHT |
|------|--------|-------|
| Business / Data Requirements | Semantic Mapping | Physical Schema |

UI must distinguish: AI Suggested · Human Confirmed · Manual Modified · Unmapped.  
AI suggestions must never look like confirmed configuration.

---

## 15. Cross-Stage Boundary (S4 ↔ S5)

| S4 outputs | S5 inputs |
|------------|-----------|
| Confirmed Data Requirements | Business Model |
| Active Data Contracts | Active Data Contracts |

S5 must **not** read MySQL tables / Postgres columns / HTTP endpoints directly.  
S5 knows only **contract logical semantics**.

---

## 16. Testing Strategy & Cost Policy

### FORMA-AI-TEST-COST-POLICY

- **≥95%** of tests use deterministic fixtures
- Real model only to prove:
  - **A.** Business Model → Data Requirement Proposal
  - **B.** Schema Metadata → Semantic Mapping Proposal
- Confirm / Edit / Reject / Source / Discovery / Contract / Drift / Revision / Tenant / Permission / UI **must not** depend on real model

**S4-G0: REAL MODEL CALLS = 0**

### Generality Acceptance

At least **two dissimilar businesses** in final E2E (examples: lab sample flow vs procurement contract approval).  
Must not rely solely on a repair/work-order fixture.  
Domain implementation must not change per business type.

---

## 17. Non-Goals (S4)

Business Capability · Capability Binding · Agent Composer / Projection · Coze Agent generation · Workflow generation · Human Task · Application · Release · Evaluation · MCP · Knowledge Graph · ETL · Data Warehouse · BI · Write Command Plane

---

## 18. Final Hard Gates (locked in Stage Contract)

| Gate | Name | Rule |
|------|------|------|
| 1 | Domain Agnostic | Unknown business → Forma Core change = 0 |
| 2 | AI No Silent Mutation | AI cannot Confirm / Activate |
| 3 | Requirement Provenance | Requirement traces to Business Model Revision (+ AnalysisRun when AI-generated) |
| 4 | Credential Isolation | No secret leakage |
| 5 | Source Independence | Logical contract independent of physical naming |
| 6 | Immutable Revision | Contract revision not updated in place |
| 7 | Schema Drift | Breaking drift → STALE |
| 8 | Business Revision Gap | Model revision does not silently mutate contract |
| 9 | Tenant Isolation | Tenant A/B fully isolated |
| 10 | Read-only Boundary | Contract does not execute business writes |
| 11 | Requirement Analysis Idempotency | Same `client_request_id` must not re-call model or duplicate requirements |
| 12 | Requirement Decision Provenance | Any CONFIRMED / REJECTED / MANUAL_MODIFIED requirement must trace to a human Decision |

---

## 19. Implementation Roadmap (locked)

| Gate | Focus |
|------|-------|
| **S4-G0** | Architecture / Contract Freeze (+ G0-F1 audit/idempotency amendment) |
| **S4-G1** | Data Requirement Domain — includes `DataRequirement`, `DataRequirementAnalysisRun`, `DataRequirementDecision`, requirement provenance, AI proposal boundary, human confirmation, edit lineage, idempotency, tenant isolation (**not implemented in G0-F1**) |
| **S4-G2** | Data Source / Credential / Discovery |
| **S4-G3** | Semantic Mapping |
| **S4-G4** | Data Contract / Revision / Drift |
| **S4-G5** | Data Plane UI |
| **S4-G6** | Live E2E / Security / Freeze |

Each gate: Implement → Test → Review → PASS.  
**Forbidden:** implement entire S4 in one pass.

---

## 20. G0 / G0-F1 Exit Criteria

- Stage Contract published (including G0-F1 audit / idempotency amendment)
- G0 Result published and updated for G0-F1
- Docs-only commit; Forma CI ALL GREEN
- No S4 domain/migration/adapter/UI implementation
- `REAL_MODEL_CALLS = 0`
- Human architecture review: G0 core PASS; G0-F1 awaiting final PASS
- **DO NOT START S4-G1** until final review PASS

---

## Document Control

| Item | Value |
|------|-------|
| Contract file | `forma/docs/stages/FORMA-S4-DATA-PLANE-DATA-CONTRACT-STAGE-CONTRACT.md` |
| Result file | `forma/cursor-results/FORMA-S4-G0-ARCHITECTURE-CONTRACT-RESULT.md` |
| Owner stage | S4-G0 / S4-G0-F1 |
| Next stage | S4-G1 (blocked until final human architecture review) |
