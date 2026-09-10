# FORMA-S5 — Business Capability
# STAGE CONTRACT (Architecture Freeze)

**Gate:** S5-G1 — Business Capability Architecture & Stage Contract Freeze
**Amendment:** S5-G1-F1 — Architecture Consistency Fix (2026-09-10)
**Status:** CONTRACT_READY (awaiting human architecture review)
**S4 Baseline:** `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae`
**S5-G0 Baseline:** Platform Admin / User Management Foundation — PASS (`S5_G1_READY = YES`)
**Scope of this document:** Architecture invariants, domain model, boundaries, hard gates, and G2–G6 roadmap for Business Capability.
**Code change rule for G1 / G1-F1:** Documentation and static verification only. No domain code, migrations, adapters, APIs, UI modules, or model calls.

---

## 1. Product Definition

### 1.1 Official Name

**S5 — Business Capability**

(S5-G0 Platform Admin is a completed foundation gate under the S5 stage series; this Stage Contract freezes the **Business Capability** product domain.)

### 1.2 Responsibility

A **Business Capability** is:

> Based on a confirmed **Business Model** and an activated **Data Contract**, provide a stable, authorizable, auditable, versionable **business capability** to downstream consumers.

### 1.3 Core Chain

```
Business Model (confirmed revision)
    ↓
Active Data Contract (logical descriptor)
    ↓
CapabilityAnalysisRun → CapabilityProposal(PROPOSED)   [AI path]
    or
MANUAL_CREATE → BusinessCapabilityRevision(DRAFT)      [manual path]
    ↓
Human CONFIRM / EDIT_CONFIRM / REJECT
    ↓
(on CONFIRM / EDIT_CONFIRM) materialize immutable BusinessCapabilityRevision(DRAFT)
    ↓
Validate → Activate
    ↓
Downstream Agent / Workflow / Application (later stages)
```

### 1.4 S5 Business Capability Is Not

- A replacement for Business Model
- A replacement for Data Contract
- An arbitrary SQL / API / script executor
- An Agent, Workflow, or Application
- A UI button definition
- A mirror of Coze Bot / Workflow
- An unaudited business write entry point

---

## 2. Baseline & Preservation

| Stage | Status | Tag / Target |
|-------|--------|--------------|
| S0–S3 | PASS / FROZEN | respective `forma-s*-frozen` tags |
| S4 | PASS / FROZEN | `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae` |
| S5-G0 | PASS | Platform Admin / User Management Foundation |

**Forbidden in S5-G1 and later Business Capability gates unless explicitly authorized:**

- Modify S0–S4 frozen architecture semantics
- Move or recreate `forma-s4-frozen-r2`
- Create `forma-s5-frozen` before S5-G6 + human freeze
- Enter S5-G2 until this Stage Contract is human-approved
- Implement Agent, Workflow, Application, Runtime, or Query Engine

---

## 3. Source of Truth Boundaries (LOCKED)

### 3.1 BUSINESS_MODEL_SOT

**Business Model remains the Canonical Business Semantic Source of Truth** (ADR-013).

- Capability **refers to** an explicit Business Model revision.
- Capability **must not replace** or redefine Business Model semantics.
- Capability **must not** invent industry business objects outside the referenced model.

### 3.2 CONTRACT_LOGICAL_ONLY

**Active Data Contract logical descriptor** is the only data-read interface for QUERY capabilities.

S5 **must not** read, store, or expose:

- table / column / endpoint / JSON path
- connection / credential
- schema snapshot / physical locator
- any S4 physical binding material

Capability may only reference:

- Data Contract ID
- Data Contract version / revision ID
- Logical schema / query capabilities / filter·sort·pagination policies from the **consumer-facing logical descriptor**

### 3.3 No Silent Upstream Mutation

When Business Model or Data Contract changes:

- ACTIVE Capability **must not** be silently rewritten.
- Impact / gap / invalidation is detected and recorded.
- Humans create a **new Capability revision** (or Deprecate / mark STALE per §8).

---

## 4. Architecture Invariants

### 4.1 DOMAIN_AGNOSTIC

Forma Core **must not** hard-code industry-specific business objects.

Forbidden as platform-level concepts (non-exhaustive):

`work_order`, `repair`, `refund`, `energy`, `device`, 工单, 退款, 设备, 能源, 审批, 医疗, 采购, 订单

These may appear only in fixtures, demos, tests, or industry templates **outside** Core.

**Generality standard:** onboarding a never-before-supported industry requires:

```
Forma Core Code Change = 0
```

### 4.2 AI_NO_SILENT_MUTATION / HUMAN_DECISION_PROVENANCE

| Actor | Allowed | Forbidden |
|-------|---------|-----------|
| AI | Create `CapabilityAnalysisRun` and `CapabilityProposal(PROPOSED)` | Confirm, EditConfirm, Reject, Validate, Activate, Deprecate, create Revision, alter ACTIVE |
| Human | CONFIRM / EDIT_CONFIRM / REJECT / MANUAL_CREATE / Validate / Activate / Deprecate | — |

Every human CONFIRM / EDIT_CONFIRM / REJECT / CREATE / Activate / Deprecate MUST produce an immutable `CapabilityDecision` (or lifecycle) audit record.

### 4.3 IMMUTABLE_REVISION

**All persisted** `BusinessCapabilityRevision` semantic content is immutable, **including DRAFT**.

- Edits to any revision create a **new** DRAFT with `derived_from_revision_id`.
- Status transitions, active pointer updates, and audit/decision events are **not** semantic content mutation.
- **Forbidden:** UPDATE that overwrites `input_schema`, `output_schema`, bindings, preconditions, effects, or other semantic fields on an existing revision row.

### 4.4 COMMAND_EXECUTION_BOUNDARY

COMMAND capabilities express business action **intent** only. They:

- MUST NOT execute writes through S4 Data Contract (S4 remains read-only)
- MUST NOT carry SQL / Shell / JavaScript / Python / arbitrary executable expressions
- MUST NOT bypass a future Action Adapter / Execution boundary
- Are **not** executed in G1–G5 product code (execution belongs to later Runtime / Action gates)

### 4.5 TENANT_ISOLATION / ROLE_AUTHORIZATION

All Capability resources are tenant-scoped. Cross-tenant Business / Contract / Capability references are rejected.

Authorization reuses S1 Tenancy + S5-G0 Platform roles.
**SUPER_ADMIN does not automatically receive tenant business-data permissions.**

Exact role matrix: §11.2.

### 4.6 SECRET_ISOLATION

SECRET values must never enter LLM prompts/context, logs, audit payloads, error responses, browser storage, Git, or screenshots.
PII preview/sample defaults to masked. Classification reuse S4: `PUBLIC` · `INTERNAL` · `CONFIDENTIAL` · `PII` · `SECRET`.

### 4.7 CONSUMER_VERSION_PINNING

Agents / Workflows / Applications (future) may depend only on:

```
ACTIVE Capability ID + explicit Capability version + public interface
```

Capability updates must not silently change consumers pinned to a prior version.

### 4.8 NO_S4_REGRESSION

S4 Data Plane / Data Contract semantics, read-only boundary, credential isolation, and freeze tag remain unchanged.

### 4.9 CAPABILITY_ASSET_IDENTITY

```
capability_id == asset_id
```

- ID is allocated by Asset Registry or a unified create flow — **no second identity namespace**.
- Asset Registry owns **header lifecycle**.
- Capability aggregate **does not** persist a duplicate generic `status` field.
- Capability Domain owns **revision lifecycle** and **`active_revision_id`** only.
- Create failure MUST NOT leave an orphan Asset or Capability; G2 MUST use a transactional Unit of Work or equivalent idempotent compensation.

---

## 5. Repository Conventions (Future Layout — G1 does not create code)

Follow existing Forma conventions (do not invent a parallel layout):

| Concern | Path |
|---------|------|
| Domain | `coze-studio/backend/domain/forma/capability/` with `entity/`, `service/`, `repository/`, `internal/dal/` |
| Application | `coze-studio/backend/application/forma/` |
| API handlers | `coze-studio/backend/api/handler/forma/` |
| Router | `coze-studio/backend/api/router/forma/` |
| Cross-domain ACL | `coze-studio/backend/crossdomain/forma/` (+ `integration/` when needed) |
| Migrations | `coze-studio/docker/atlas/forma/migrations/` (Forma-owned `forma_*` tables only) |
| Frontend package | `@forma/capability` under `frontend/packages/forma-capability/` |
| Shell route | Replace `/capabilities` placeholder (“能力资产”) |

**Referenced, not rebuilt:**

`Tenant`, `Principal`, Membership Role (`OWNER`/`ADMIN`/`MEMBER`/`VIEWER`), Platform Role (`SUPER_ADMIN`/`USER`), `Business`, `BusinessModelRevision`, `DataContract`, `DataContractRevision`, Asset Registry

**Database ownership (LOCKED, ADR-006):**

- No Coze core table modification
- No Forma columns on Coze tables
- No hard FK into Coze tables
- Integrate via ref / mapping / ACL only

---

## 6. Domain Model

### 6.1 Canonical Business Semantic Assets

| # | Asset | Role |
|---|-------|------|
| 1 | **BusinessCapability** | Canonical CAPABILITY asset aggregate: stable identity (`capability_id == asset_id`), `active_revision_id`, tenant ownership |
| 2 | **BusinessCapabilityRevision** | Immutable semantic version of a capability (public interface + bindings + preconditions/effects) |

**Asset Registry note:** `AssetKindCapability = "CAPABILITY"` already exists in S0 Asset Registry taxonomy.
`capability_id == asset_id` is **LOCKED** (§4.9).
Asset Registry owns asset header lifecycle metadata; Capability Domain owns capability semantic revisions and the active pointer — **not** a duplicate aggregate `status`.

### 6.2 Structural / Binding Components (not separate Canonical Assets)

These are part of a Capability Revision payload or first-class revision sub-resources — **not** new Canonical Asset kinds:

| Component | Role |
|-----------|------|
| `CapabilityInputSchema` | Declared input logical schema for invocation |
| `CapabilityOutputSchema` | Declared output logical schema for invocation |
| `CapabilityPrecondition` | Domain-agnostic preconditions (typed predicates over logical inputs / model state refs) |
| `CapabilityEffect` | Declared business effects / postconditions (intent description; not executable code) |
| `CapabilityDataContractBinding` | Pin to Data Contract ID + version + logical field/query mapping (logical only) |

### 6.3 Operational / Workflow / Audit Records (not Semantic SoT / not Canonical Assets)

| Record | Role |
|--------|------|
| `CapabilityAnalysisRun` | AI analysis execution binding; idempotency key + `request_digest` |
| `CapabilityProposal` | Workflow record for AI (or review) proposals; status `PROPOSED` / `REJECTED` / materialization linkage; **not** a Canonical Asset |
| `CapabilityDecision` | Immutable human decision event (CONFIRM / EDIT_CONFIRM / REJECT / CREATE / Activate / Deprecate, …) |
| `CapabilityValidationResult` | Technical validation outcome for a revision |
| `CapabilityImpactResult` | Impact / gap / invalidation outcome when upstream Business Model or Contract changes |

These must not redefine Business Model or Data Contract semantics.

### 6.4 Suggested Identity Fields (informative for G2+)

**BusinessCapability:**
`capability_id` (== `asset_id`), `tenant_id`, `business_id`, `name`, `description`, `capability_kind`, `active_revision_id` (nullable), `created_by`, `created_at`, `updated_at`
(**No** duplicate generic `status` column on the Capability aggregate.)

**BusinessCapabilityRevision:**
`revision_id`, `capability_id`, `tenant_id`, `business_id`, `version`, `status`, `business_model_revision`, `capability_kind`, `input_schema`, `output_schema`, `preconditions[]`, `effects[]`, `data_contract_bindings[]`, `derived_from_revision_id` (nullable), `analysis_run_id` (nullable), `proposal_id` (nullable), `source` (`AI_PROPOSAL` / `MANUAL_CREATED` / `DERIVED_EDIT`), `created_by`, `created_at`

**CapabilityProposal:**
`proposal_id`, `tenant_id`, `business_id`, `analysis_run_id`, `capability_id` (nullable until bound), `status` (`PROPOSED` / `REJECTED` / `CONFIRMED` / `EDIT_CONFIRMED`), `payload` (proposed semantic content), `materialized_revision_id` (nullable), `created_at`

**CapabilityDecision:**
`decision_id`, `tenant_id`, `business_id`, `capability_id`, `proposal_id` (nullable), `source_revision_id` (nullable), `target_revision_id` (nullable), `action`, `actor_principal_id`, `reason` (nullable), `created_at`
Decisions are immutable after create. AI must not create human decisions.

**CapabilityAnalysisRun:**
`analysis_run_id`, `tenant_id`, `business_id`, `business_model_revision`, `client_request_id`, `request_digest`, `status`, `created_at`, …

---

## 7. Capability Kinds & Boundaries

### 7.1 Domain-agnostic kinds only

| Kind | Meaning | V1 |
|------|---------|----|
| **QUERY** | Read via S4 logical Data Contract (`READ` / `LOOKUP` / `LIST` / `FILTER` / `AGGREGATE`) | **In scope** |
| **COMMAND** | Express business action intent (approve / submit / close / create / … as **model-derived verbs**, not platform enums) | **In scope (definition only; no execution)** |
| **COMPOSITE** | Orchestrate multiple capabilities | **Deferred** — insufficient V1 evidence; do not pre-build empty abstractions |

No industry capability-type enums.

### 7.2 QUERY rules

- May only invoke Active Data Contract **logical** operations.
- Input/output schemas must be compatible with the bound contract logical schema (§10).
- Must not expose physical bindings to consumers.

### 7.3 COMMAND rules (LOCKED)

COMMAND describes:

- Business action intent
- Input schema
- Preconditions
- Declared effects
- Authorization requirements

COMMAND **must not**:

- Write through S4 Data Contract
- Embed SQL / Shell / JavaScript / Python / arbitrary executable expressions
- Bypass future Action Adapter / Execution boundary
- Be treated as an executed write in G1–G5

**G1 does not implement real write execution.**
Execution adapters, side-effect runtimes, and compensating transactions are **out of scope** for this Stage Contract’s implementation gates until a later stage explicitly opens them.

---

## 8. State Machine & Immutable Revision (LOCKED)

### 8.1 Unique legal revision status machine

```
DRAFT → VALIDATED → ACTIVE
ACTIVE → DEPRECATED
ACTIVE → STALE
STALE → DEPRECATED
```

| Status | Meaning |
|--------|---------|
| `DRAFT` | Immutable persisted revision awaiting Validate; semantic fields never updated in place |
| `VALIDATED` | Passed technical validation against pinned Business Model revision + Data Contract logical bindings; human Confirm prerequisite satisfied (§9.6) |
| `ACTIVE` | Consumers may pin this revision; at most one ACTIVE revision per Capability aggregate |
| `STALE` | Upstream pin invalidated per §8.3; not auto-repaired |
| `DEPRECATED` | Retired (human Deprecate, or superseded ACTIVE on successful Activate) |

**Illegal:** `VALIDATED → DEPRECATED` without passing ACTIVE; `STALE → ACTIVE` in place; `DRAFT → ACTIVE` skipping VALIDATED.

### 8.2 Transition rules

| From | To | Actor | Notes |
|------|----|-------|-------|
| — | DRAFT | Human CONFIRM / EDIT_CONFIRM (materialize from Proposal) or Human MANUAL_CREATE | See §9; AI never creates Revision |
| DRAFT | VALIDATED | Human-triggered Validate | Requires §9.6 Confirm prerequisite + successful `CapabilityValidationResult` |
| VALIDATED | ACTIVE | Human Activate | Single transaction per §8.4 |
| ACTIVE | DEPRECATED | Human Deprecate **or** Activate supersession | Auditable |
| ACTIVE | STALE | System/human impact evaluation | Only via §8.3 triggers + recorded `CapabilityImpactResult` |
| STALE | DEPRECATED | Human | Retirement after invalidation |

Repair path after STALE/DEPRECATED: human creates a **new** DRAFT revision (`derived_from_revision_id` optional); never revive STALE content in place as ACTIVE.

**Illegal transitions** return stable Forma error keys (to be enumerated in G4 API gate). Examples of direction (names reserved):

`FORMA_CAPABILITY_INVALID_TRANSITION`, `FORMA_CAPABILITY_REVISION_IMMUTABLE`, `FORMA_CAPABILITY_ACTIVE_CONFLICT`, `FORMA_CAPABILITY_CONFIRM_REQUIRED`

### 8.3 STALE triggers (precise — not a copy of Data Contract drift)

Capability **never** observes physical schema. Therefore STALE is **not** caused by table/column drift.

STALE may be set on an ACTIVE revision **only when** all of the following hold:

1. An impact evaluation runs against the ACTIVE revision’s pins; and
2. At least one of:
   - Pinned Data Contract revision is no longer `ACTIVE` (became `STALE` or `DEPRECATED` / unbound); or
   - Pinned Business Model revision is superseded **and** impact analysis records an **incompatible** gap that invalidates the capability’s declared interface or bindings; and
3. A `CapabilityImpactResult` is persisted; and
4. Transition is auditable.

**Forbidden:**

- AI unilaterally marking STALE
- Silent rewrite of ACTIVE content to “fix” upstream changes
- Physical Discover / schema capture triggering Capability STALE

### 8.4 Activate / active pointer policy (LOCKED)

Activating a new VALIDATED revision **MUST** occur in **one transaction**:

1. Lock the Capability aggregate
2. Current old ACTIVE revision (if any) → `DEPRECATED`
3. New revision → `ACTIVE`
4. `active_revision_id` → new revision
5. Write lifecycle / decision audit

When the current ACTIVE becomes `STALE` or `DEPRECATED`:

- Clear `active_revision_id` **only if** the active pointer currently references that revision
- Handling historical STALE / former ACTIVE revisions **MUST NOT** clear a newer ACTIVE pointer

**Concurrency:** concurrent Activate of two VALIDATED revisions → **at most one succeeds**; loser receives conflict error.
Optimistic concurrency (expected revision / generation) or equivalent transactional guard is required for Activate / Deprecate / STALE.

### 8.5 Edit lineage & semantic immutability

Editing any persisted revision’s semantics (including DRAFT):

1. Create **new** DRAFT revision (`derived_from_revision_id` = source)
2. **Never** UPDATE/overwrite source revision semantic fields
3. Human Confirm (if from Proposal path) / Validate / Activate remain separate steps with `CapabilityDecision` records

Status transition, active pointer change, and audit events are allowed without rewriting semantic content.

---

## 9. Proposal / Revision Seam & AI Propose / Human Confirm (LOCKED)

### 9.1 Unique Proposal → Revision model

```
CapabilityAnalysisRun
  → CapabilityProposal(PROPOSED)
  → Human CONFIRM / EDIT_CONFIRM
  → materialize immutable BusinessCapabilityRevision(DRAFT)

Human REJECT
  → CapabilityProposal(REJECTED)
  → no Revision created
```

- `CapabilityProposal` is a **workflow record**, not a Canonical Asset.
- AI may **only** create Proposals (via AnalysisRun).
- AI **must not** directly create a Revision that can be Validated or Activated.
- CONFIRM, EDIT_CONFIRM, and REJECT **must** generate an immutable `CapabilityDecision`.
- `MANUAL_CREATE` may produce a DRAFT directly, but **must** record a human CREATE decision and `source=MANUAL_CREATED`.

**Forbidden wording / behavior:** AI proposal materialized directly as DRAFT without CONFIRM / EDIT_CONFIRM.

### 9.2 AI may

- Create `CapabilityAnalysisRun` and `CapabilityProposal(PROPOSED)` from Business Model revision + Active Data Contract logical descriptors
- Suggest input/output schemas, preconditions, effects, and contract bindings inside Proposal payload
- Produce analysis metadata (confidence, reason) — **confidence ≠ confirmation**

### 9.3 AI must not

- CONFIRM / EDIT_CONFIRM / REJECT
- Validate / Activate / Deprecate
- Create `BusinessCapabilityRevision`
- Silently mutate ACTIVE Capability
- Create `CapabilityDecision` human records
- Re-call models for the same idempotent analysis key + digest (§9.5)

### 9.4 Provenance

AI-originated DRAFT revisions (after human CONFIRM / EDIT_CONFIRM) MUST trace to:

- `CapabilityAnalysisRun`
- `CapabilityProposal`
- Explicit `business_model_revision`
- Input sources (contract ID/version pins and/or requirement refs as applicable)
- Human `CapabilityDecision` (CONFIRM or EDIT_CONFIRM)

### 9.5 Idempotency (LOCKED)

Minimum logical key:

```
tenant_id + business_id + business_model_revision + client_request_id
```

Persist `request_digest` covering at least:

- `business_model_revision`
- Sorted Data Contract ID/version pins
- Input requirement refs
- Analysis options

Rules:

| Case | Behavior |
|------|----------|
| Same key + same digest | Return original `CapabilityAnalysisRun` and original Proposal set; **no** new model call |
| Same key + different digest | Return stable **idempotency conflict** error |
| Concurrent identical requests | Exactly one model execution acquires the run |
| Proposal materialization for one AnalysisRun | Must be idempotent |
| Retries | Must not create duplicate Proposals or Revisions |

### 9.6 Human Confirm prerequisite for Validate (LOCKED)

A DRAFT may be Validated **only if**:

- It was materialized via human CONFIRM or EDIT_CONFIRM from a Proposal; **or**
- It is `source=MANUAL_CREATED` with a recorded human CREATE decision

**Forbidden:**

- Validating an AI Proposal directly (Proposal is not a Revision)
- Validate or Activate implicitly substituting for Confirm
- Skipping Confirm on the AI path

---

### 9.7 Model adapter path (future gates)

```
Forma Capability Domain
  → Forma CrossDomain / ACL
  → Coze/Eino Model Manager
```

**Forbidden in Capability Domain:** provider-specific SDKs (OpenAI / DeepSeek / Qwen / …).

**S5-G1 / S5-G1-F1: REAL_MODEL_CALLS = 0**

---

## 10. Schema Compatibility & Controlled Expressions

### 10.1 V1 QUERY schema compatibility (LOCKED)

Capability `input_schema` / `output_schema` are **logical** schemas.

For QUERY capabilities:

- Logical types **MUST** use **S4 normalized logical types**
- V1 **disallows** implicit type conversion
- Every Capability **required** output field must map to a Contract field that **exists** and is **non-null** in the logical schema
- Capability **nullable** output may accept a non-null Contract field
- Requested QUERY operation MUST be in Contract `query_capabilities`
- Filter field/operator MUST be a **subset** of Contract `filter_schema`
- Sort field/direction MUST be a **subset** of Contract `sort_schema`
- Pagination limit MUST NOT exceed Contract max limit
- `LOOKUP` / `LIST` / `AGGREGATE` output **cardinality** MUST be explicit
- Missing required input/output, type mismatch, or weakened guarantees → **Validation FAIL**

COMMAND schemas describe intent payloads only; they still MUST NOT reference physical fields. COMMAND↔Contract write compatibility is out of scope (no S4 writes).

### 10.2 Forbidden

- Physical field names / paths in Capability public interface
- Arbitrary code expressions (JS / Python / SQL / Shell)
- Undefined dynamic types without schema
- SECRET in LLM, logs, audit, or error responses
- Unmasked PII as default preview

### 10.3 Controlled expression DSL

If preconditions/effects need computable predicates:

- Only a **controlled, parseable, validatable, auditable** DSL may be introduced in a later gate
- G1 evidence is insufficient to freeze a full language surface

**Decision (G1):** precondition/effect **DSL language freeze is DEFERRED** to a later gate (no later than Capability Validation / API gates when needed).
Until then, preconditions/effects are structured typed records (enums + logical field refs + literal comparands), not free-form code.

---

## 11. Authorization, Tenancy & Audit

### 11.1 Identity reuse

- Tenant / Principal / Membership Role: S1
- Platform Role (`SUPER_ADMIN` / `USER`): S5-G0
- Session SoT: Coze SessionAuth (unchanged)

Membership for Capability actions MUST be **ACTIVE**.

### 11.2 Permission matrix (LOCKED for G4 enforcement)

| Action | Allowed roles (tenant-scoped; ACTIVE membership) |
|--------|--------------------------------------------------|
| List / Get Capability | `OWNER`, `ADMIN`, `MEMBER`, `VIEWER` |
| Create (MANUAL_CREATE) | `OWNER`, `ADMIN` |
| Confirm / EditConfirm / Reject | `OWNER`, `ADMIN` |
| Validate | `OWNER`, `ADMIN` |
| Activate / Deprecate | `OWNER` |
| Invoke QUERY (future runtime) | Deferred to Runtime gate; must still require ACTIVE membership |
| Invoke COMMAND (future runtime) | Deferred to Runtime gate |
| Platform-wide admin of all tenants’ capabilities | **Not** implied by `SUPER_ADMIN` |

**SUPER_ADMIN does not automatically receive tenant business-data permissions.**

Cross-tenant references: **deny**.

### 11.3 Audit

Audit / decision / validation / impact records MUST NOT store credential, token, cookie, Authorization header, or unsanitized errors.

---

## 12. Consumer Boundary (LOCKED)

Future consumers (Agent / Workflow / Application):

- Depend only on `capability_id` + **explicit version** + public interface
- MUST NOT bypass Capability to depend on Data Contract physical bindings
- MUST NOT assume auto-upgrade when a new Capability revision activates
- Projection into Coze Agent/Workflow/App is **out of scope** for S5-G1–G5

This stage does **not** implement:

- Agent projection
- Coze Workflow generation
- Runtime execution / Query Engine
- Application assembly

---

## 13. UI Information Architecture (future G5)

Primary module name: **能力资产 / Capabilities** (existing shell placeholder `/capabilities`).

Planned pages (informative):

- Capability list (per Business / tenant)
- Capability detail / revisions
- Propose & review (AI suggestions visually distinct from confirmed)
- Validate / Activate / Deprecate actions
- Impact / STALE insights

AI suggestions must never look like confirmed configuration.

---

## 14. Testing Strategy & Cost Policy

### FORMA-AI-TEST-COST-POLICY (inherited)

- ≥95% tests use deterministic fixtures
- Real model only when a later gate must prove AI proposal quality
- Confirm / Validate / Activate / Auth / Tenant / UI / Impact **must not** depend on real model for correctness

**S5-G1 / S5-G1-F1: REAL_MODEL_CALLS = 0**

### Generality acceptance (later E2E gates)

At least two dissimilar businesses. Domain implementation must not change per business type.

---

## 15. Non-Goals (S5 Business Capability stage series through G5)

- Agent Composer / Coze Agent generation
- Workflow generation
- Application assembly / Release / Evaluation
- MCP / Knowledge Graph
- ETL / Data Warehouse / BI
- Real COMMAND write execution / Action Adapter runtime
- Query Engine productization beyond S4 contract read semantics
- Moving S4 freeze tags / creating `forma-s5-frozen` before G6 + human freeze

---

## 16. Final Hard Gates (locked in Stage Contract)

| # | Gate | Rule |
|---|------|------|
| 1 | DOMAIN_AGNOSTIC | Unknown business → Forma Core change = 0 |
| 2 | BUSINESS_MODEL_SOT | Capability refers to Business Model; never replaces it |
| 3 | CONTRACT_LOGICAL_ONLY | No physical bindings in Capability public surface |
| 4 | AI_NO_SILENT_MUTATION | AI cannot Confirm / Validate / Activate / create Revision / alter ACTIVE |
| 5 | HUMAN_DECISION_PROVENANCE | Human lifecycle actions produce immutable decisions; Confirm required before Validate on AI path |
| 6 | IMMUTABLE_REVISION | No in-place mutation of **any** persisted revision semantic content (including DRAFT) |
| 7 | TENANT_ISOLATION | Tenant A/B fully isolated; cross-tenant refs denied |
| 8 | ROLE_AUTHORIZATION | Actions enforced server-side via exact role matrix (§11.2) |
| 9 | COMMAND_EXECUTION_BOUNDARY | COMMAND is intent only; no arbitrary write/exec in S5 Capability gates |
| 10 | NO_ARBITRARY_CODE | No JS/Python/SQL/Shell expressions in Capability definitions |
| 11 | IDEMPOTENCY | Same analysis key + digest must not re-call model or duplicate proposals/revisions |
| 12 | AUDITABILITY | Activate / Deprecate / STALE / decisions are auditable |
| 13 | SECRET_ISOLATION | No secret leakage into LLM/logs/audit/errors |
| 14 | CONSUMER_VERSION_PINNING | Consumers pin explicit Capability versions |
| 15 | NO_S4_REGRESSION | S4 freeze semantics and read-only contract boundary preserved |

---

## 17. Implementation Roadmap (locked)

| Gate | Focus |
|------|-------|
| **S5-G0** | Platform Admin / User Management Foundation — **PASS** (prerequisite; not redefined here) |
| **S5-G1** | Architecture & Stage Contract Freeze |
| **S5-G1-F1** | Architecture Consistency Fix (**this amendment**) |
| **S5-G2** | Capability Domain, Revision, state machine, human decisions, idempotent analysis runs |
| **S5-G3** | Business Model / Data Contract binding & validation |
| **S5-G4** | Capability API, authorization, audit, concurrency consistency |
| **S5-G5** | Capability UI (`@forma/capability`, `/capabilities`) |
| **S5-G6** | Live E2E, Security, Browser Acceptance, Freeze |

Each gate: Implement → Test → Review → PASS.
**Forbidden:** implement entire S5 Business Capability in one pass.

**Roadmap note:** No separate formal S5 Business Capability roadmap document existed in-repo before G1. This section is the first locked roadmap for the Capability domain and aligns with the authorized S5-G1 brief. S5-G0 remains the platform-admin foundation gate already completed.

---

## 18. G1 / G1-F1 Exit Criteria

- Stage Contract published at `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md`
- G1 Result at `forma/cursor-results/FORMA-S5-G1-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md`
- G1-F1 Result at `forma/cursor-results/FORMA-S5-G1-F1-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md`
- Docs-only commit; Forma CI ALL GREEN
- No Capability domain / migration / adapter / UI implementation
- `REAL_MODEL_CALLS = 0`
- `PRODUCT_CODE_CHANGE = NONE`
- `MIGRATION_CHANGE = NONE`
- `S4_FROZEN_BASELINE_UNCHANGED = PASS`
- Human architecture review required before S5-G2
- **DO NOT START S5-G2** until review PASS
- **DO NOT** create `forma-s5-frozen`
- **S5_G2_READY = NO** until human review after G1-F1

---

## 19. Deferred Decisions (explicit — not pretend-implemented)

| Item | Status |
|------|--------|
| COMPOSITE capability kind | Deferred |
| Full precondition/effect DSL language | Deferred |
| COMMAND runtime / Action Adapter | Deferred to later stage |
| Invoke QUERY/COMMAND runtime ACL | Deferred to Runtime gate |
| Consumer invoke ACL fine-grain field/row policies | Deferred to Runtime / later ACL gate |

---

## Document Control

| Field | Value |
|-------|-------|
| Document | FORMA-S5 Business Capability Stage Contract |
| Gate | S5-G1 / S5-G1-F1 |
| Baseline tags | `forma-s4-frozen-r2` |
| Related ADRs | ADR-002, ADR-006, ADR-013 |
| Related contracts | FORMA-S4 Data Plane / Data Contract Stage Contract |
| Related results | FORMA-S4-FINAL-FREEZE-R2, FORMA-S5-G0-ADMIN-USER-MANAGEMENT, FORMA-S5-G1, FORMA-S5-G1-F1 |
