# FORMA-S5 — Business Capability
# STAGE CONTRACT (Architecture Freeze)

**Gate:** S5-G1 — Business Capability Architecture & Stage Contract Freeze  
**Status:** CONTRACT_READY (awaiting human architecture review)  
**S4 Baseline:** `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae`  
**S5-G0 Baseline:** Platform Admin / User Management Foundation — PASS (`S5_G1_READY = YES`)  
**Scope of this document:** Architecture invariants, domain model, boundaries, hard gates, and G2–G6 roadmap for Business Capability.  
**Code change rule for G1:** Documentation and static verification only. No domain code, migrations, adapters, APIs, UI modules, or model calls.

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
Business Capability Proposal (AI or manual)
    ↓
Human Review / Confirm / Edit / Reject
    ↓
Capability Revision
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
| AI | Propose Capability drafts / analyses | Confirm, Validate, Activate, Deprecate, alter ACTIVE |
| Human | Confirm / Edit / Reject / Validate / Activate / Deprecate | — |

Every human Confirm / Edit / Reject / Activate / Deprecate MUST produce an immutable decision (or lifecycle) audit record.

### 4.3 IMMUTABLE_REVISION

ACTIVE Capability revision content is immutable. Edits create a new revision. Active pointer updates are concurrency-safe and auditable.

### 4.4 COMMAND_EXECUTION_BOUNDARY

COMMAND capabilities express business action **intent** only. They:

- MUST NOT execute writes through S4 Data Contract (S4 remains read-only)
- MUST NOT carry SQL / Shell / JavaScript / Python / arbitrary executable expressions
- MUST NOT bypass a future Action Adapter / Execution boundary
- Are **not** executed in G1–G5 product code (execution belongs to later Runtime / Action gates)

### 4.5 TENANT_ISOLATION / ROLE_AUTHORIZATION

All Capability resources are tenant-scoped. Cross-tenant Business / Contract / Capability references are rejected.

Authorization reuses S1 Tenancy + S5-G0 Platform roles.  
**SUPER_ADMIN is not automatic ownership of all tenant business data** unless an existing stage contract explicitly grants a specific action.

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
| 1 | **BusinessCapability** | Canonical CAPABILITY asset aggregate: stable identity, active pointer, tenant ownership |
| 2 | **BusinessCapabilityRevision** | Immutable semantic version of a capability (public interface + bindings + preconditions/effects) |

**Asset Registry note:** `AssetKindCapability = "CAPABILITY"` already exists in S0 Asset Registry taxonomy.  
`capability_id` SHOULD align with `asset_id` for the CAPABILITY kind (same pattern as Business: `business_id == asset_id`), subject to G2 implementation confirmation.  
Asset Registry owns asset header lifecycle metadata; Capability Domain owns capability semantic revisions and activation pointer.

### 6.2 Structural / Binding Components (not separate Canonical Assets)

These are part of a Capability Revision payload or first-class revision sub-resources — **not** new Canonical Asset kinds:

| Component | Role |
|-----------|------|
| `CapabilityInputSchema` | Declared input logical schema for invocation |
| `CapabilityOutputSchema` | Declared output logical schema for invocation |
| `CapabilityPrecondition` | Domain-agnostic preconditions (typed predicates over logical inputs / model state refs) |
| `CapabilityEffect` | Declared business effects / postconditions (intent description; not executable code) |
| `CapabilityDataContractBinding` | Pin to Data Contract ID + version + logical field/query mapping (logical only) |

### 6.3 Operational / Workflow / Audit Records (not Semantic SoT)

| Record | Role |
|--------|------|
| `CapabilityAnalysisRun` | AI analysis execution binding (optional; when AI proposes) |
| `CapabilityDecision` | Immutable human decision event (Confirm / Reject / EditConfirm / Activate / Deprecate, …) |
| `CapabilityValidationResult` | Technical validation outcome for a revision |
| `CapabilityImpactResult` | Impact / gap / invalidation outcome when upstream Business Model or Contract changes |

These must not redefine Business Model or Data Contract semantics.

### 6.4 Suggested Identity Fields (informative for G2+)

**BusinessCapability:**  
`capability_id`, `tenant_id`, `business_id`, `name`, `description`, `capability_kind`, `active_revision_id` (nullable), `status` (aggregate pointer status), `created_by`, `created_at`, `updated_at`

**BusinessCapabilityRevision:**  
`revision_id`, `capability_id`, `tenant_id`, `business_id`, `version`, `status`, `business_model_revision`, `capability_kind`, `input_schema`, `output_schema`, `preconditions[]`, `effects[]`, `data_contract_bindings[]`, `derived_from_revision_id` (nullable), `analysis_run_id` (nullable), `source`, `created_by`, `created_at`

**CapabilityDecision:**  
`decision_id`, `tenant_id`, `business_id`, `capability_id`, `source_revision_id`, `target_revision_id` (nullable), `action`, `actor_principal_id`, `reason` (nullable), `created_at`  
Decisions are immutable after create. AI must not create human decisions.

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
DRAFT → VALIDATED → ACTIVE → DEPRECATED
                 ↘
                  STALE
```

| Status | Meaning |
|--------|---------|
| `DRAFT` | Editable working revision (new revision only; never mutate ACTIVE content) |
| `VALIDATED` | Passed technical validation against pinned Business Model revision + Data Contract logical bindings |
| `ACTIVE` | Consumers may pin this revision; at most one ACTIVE revision per Capability aggregate |
| `STALE` | Upstream pin invalidated per §8.3; not auto-repaired |
| `DEPRECATED` | Human-retired |

### 8.2 Transition rules

| From | To | Actor | Notes |
|------|----|-------|-------|
| — | DRAFT | Human (or AI proposal materialized as DRAFT) | Create revision |
| DRAFT | VALIDATED | Human-triggered Validate | Requires successful `CapabilityValidationResult` |
| VALIDATED | ACTIVE | Human Activate | Updates aggregate `active_revision_id` atomically; previous ACTIVE → DEPRECATED or remains history per activate policy |
| ACTIVE | DEPRECATED | Human Deprecate | Auditable |
| ACTIVE | STALE | System/human impact evaluation | Only via §8.3 triggers + recorded `CapabilityImpactResult` |
| STALE | DEPRECATED | Human | Retirement after invalidation |
| STALE | — | Human creates **new** DRAFT revision | Repair path; no in-place revive of STALE content as ACTIVE |

**Illegal transitions** return stable Forma error keys (to be enumerated in G4 API gate). Examples of direction (names reserved):

`FORMA_CAPABILITY_INVALID_TRANSITION`, `FORMA_CAPABILITY_REVISION_IMMUTABLE`, `FORMA_CAPABILITY_ACTIVE_CONFLICT`

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

### 8.4 Concurrency

- Exactly one ACTIVE revision pointer per Capability aggregate.
- Activate / Deprecate / STALE transitions MUST use optimistic concurrency (expected revision / generation) or equivalent transactional guard.
- Concurrent Activate of two VALIDATED revisions: exactly one wins; loser receives conflict error.

### 8.5 Edit lineage

Editing an existing revision’s semantics:

1. Create new DRAFT revision (`derived_from_revision_id` = source)
2. Do not overwrite source revision content
3. Human Confirm / Validate / Activate as separate steps with `CapabilityDecision` records

---

## 9. AI Propose / Human Confirm (LOCKED)

### 9.1 AI may

- Propose Capability drafts from Business Model revision + Active Data Contract logical descriptors
- Suggest input/output schemas, preconditions, effects, and contract bindings
- Produce analysis metadata (confidence, reason) — **confidence ≠ confirmation**

### 9.2 AI must not

- Confirm / Validate / Activate / Deprecate
- Silently mutate ACTIVE Capability
- Create `CapabilityDecision` human records
- Call models again for the same idempotent analysis key

### 9.3 Provenance

AI-generated capability revisions MUST trace to:

- `CapabilityAnalysisRun` (when AI-produced)
- Explicit `business_model_revision`
- Input sources (contract IDs/versions and/or requirement refs as applicable)

### 9.4 Idempotency (LOCKED)

Logical key (minimum):

`tenant_id` + `business_id` + `business_model_revision` + `client_request_id`

→ exactly one logical `CapabilityAnalysisRun`.

- Duplicate requests return the existing result
- Must not re-invoke the model
- Concurrent identical `client_request_id`: only one execution acquires the run

### 9.5 Model adapter path (future gates)

```
Forma Capability Domain
  → Forma CrossDomain / ACL
  → Coze/Eino Model Manager
```

**Forbidden in Capability Domain:** provider-specific SDKs (OpenAI / DeepSeek / Qwen / …).

**S5-G1: REAL_MODEL_CALLS = 0**

---

## 10. Schema Compatibility & Controlled Expressions

### 10.1 Compatibility rules

- Capability `input_schema` / `output_schema` are **logical** schemas.
- QUERY bindings: every required logical input/output used by the capability MUST be satisfiable by the pinned Data Contract logical schema and allowed query capabilities.
- COMMAND schemas describe intent payloads only; they still MUST NOT reference physical fields.

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

### 11.2 Minimum permission matrix (to be enforced in G4)

| Action | Minimum (tenant-scoped unless noted) |
|--------|--------------------------------------|
| List / Get Capability | MEMBER+ (tenant member) |
| Create draft / Propose accept into draft | ADMIN+ (or OWNER) |
| Confirm / Reject / EditConfirm | ADMIN+ |
| Validate | ADMIN+ |
| Activate / Deprecate | OWNER (or ADMIN if later policy explicitly allows; default OWNER) |
| Invoke QUERY (future runtime) | per Capability ACL + membership; default MEMBER+ |
| Invoke COMMAND (future runtime) | stricter; default ADMIN+ / OWNER — exact matrix deferred to Runtime gate |
| Platform-wide admin of all tenants’ capabilities | **Not** implied by SUPER_ADMIN |

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

**S5-G1: REAL_MODEL_CALLS = 0**

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
| 4 | AI_NO_SILENT_MUTATION | AI cannot Confirm / Validate / Activate / alter ACTIVE |
| 5 | HUMAN_DECISION_PROVENANCE | Human lifecycle actions produce immutable decisions |
| 6 | IMMUTABLE_REVISION | No in-place mutation of ACTIVE revision content |
| 7 | TENANT_ISOLATION | Tenant A/B fully isolated; cross-tenant refs denied |
| 8 | ROLE_AUTHORIZATION | Actions enforced server-side via membership / platform roles |
| 9 | COMMAND_EXECUTION_BOUNDARY | COMMAND is intent only; no arbitrary write/exec in S5 Capability gates |
| 10 | NO_ARBITRARY_CODE | No JS/Python/SQL/Shell expressions in Capability definitions |
| 11 | IDEMPOTENCY | Same `client_request_id` must not re-call model or duplicate analysis runs |
| 12 | AUDITABILITY | Activate / Deprecate / STALE / decisions are auditable |
| 13 | SECRET_ISOLATION | No secret leakage into LLM/logs/audit/errors |
| 14 | CONSUMER_VERSION_PINNING | Consumers pin explicit Capability versions |
| 15 | NO_S4_REGRESSION | S4 freeze semantics and read-only contract boundary preserved |

---

## 17. Implementation Roadmap (locked)

| Gate | Focus |
|------|-------|
| **S5-G0** | Platform Admin / User Management Foundation — **PASS** (prerequisite; not redefined here) |
| **S5-G1** | Architecture & Stage Contract Freeze (**this document**) |
| **S5-G2** | Capability Domain, Revision, state machine, human decisions, idempotent analysis runs |
| **S5-G3** | Business Model / Data Contract binding & validation |
| **S5-G4** | Capability API, authorization, audit, concurrency consistency |
| **S5-G5** | Capability UI (`@forma/capability`, `/capabilities`) |
| **S5-G6** | Live E2E, Security, Browser Acceptance, Freeze |

Each gate: Implement → Test → Review → PASS.  
**Forbidden:** implement entire S5 Business Capability in one pass.

**Roadmap note:** No separate formal S5 Business Capability roadmap document existed in-repo before G1. This section is the first locked roadmap for the Capability domain and aligns with the authorized S5-G1 brief. S5-G0 remains the platform-admin foundation gate already completed.

---

## 18. G1 Exit Criteria

- Stage Contract published at `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md`
- G1 Result published at `forma/cursor-results/FORMA-S5-G1-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md`
- Docs-only commit; Forma CI ALL GREEN
- No Capability domain / migration / adapter / UI implementation
- `REAL_MODEL_CALLS = 0`
- `PRODUCT_CODE_CHANGE = NONE`
- `MIGRATION_CHANGE = NONE`
- `S4_FROZEN_BASELINE_UNCHANGED = PASS`
- Human architecture review required before S5-G2
- **DO NOT START S5-G2** until review PASS
- **DO NOT** create `forma-s5-frozen`

---

## 19. Deferred Decisions (explicit — not pretend-implemented)

| Item | Status |
|------|--------|
| COMPOSITE capability kind | Deferred |
| Full precondition/effect DSL language | Deferred |
| COMMAND runtime / Action Adapter | Deferred to later stage |
| Exact Activate permission (OWNER-only vs ADMIN+) | Default OWNER in §11.2; may be refined in G4 with evidence |
| `capability_id == asset_id` transactional create details | Affirmed as intent; G2 implements |
| Consumer invoke ACL fine-grain field/row policies | Deferred to Runtime / later ACL gate |

---

## Document Control

| Field | Value |
|-------|-------|
| Document | FORMA-S5 Business Capability Stage Contract |
| Gate | S5-G1 |
| Baseline tags | `forma-s4-frozen-r2` |
| Related ADRs | ADR-002, ADR-006, ADR-013 |
| Related contracts | FORMA-S4 Data Plane / Data Contract Stage Contract |
| Related results | FORMA-S4-FINAL-FREEZE-R2, FORMA-S5-G0-ADMIN-USER-MANAGEMENT |
