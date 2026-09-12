# FORMA-S5 — Business Capability
# STAGE CONTRACT (Architecture Freeze)

**Gate:** S5-G1 — Business Capability Architecture & Stage Contract Freeze
**Amendment:** S5-G1-F7 — Architecture Consistency Fix (2026-09-12) — supersedes F1–F6 wording where amended
**Prior amendments:** S5-G1-F1 … F6 (2026-09-10 / 2026-09-11 / 2026-09-12)
**Status:** CONTRACT_READY (awaiting human architecture review)
**S4 Baseline:** `forma-s4-frozen-r2` → `7c05fc5da16e0f3c256ad06aaa5d2c76b9ebc7ae`
**S5-G0 Baseline:** Platform Admin / User Management Foundation — PASS (`S5_G1_READY = YES`)
**Scope of this document:** Architecture invariants, domain model, boundaries, hard gates, and G2–G6 roadmap for Business Capability.
**Code change rule for G1 / G1-F1 … G1-F7:** Documentation and static verification only. No domain code, migrations, adapters, APIs, UI modules, or model calls.

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
MANUAL_CREATE → BusinessCapabilityRevision(DRAFT)      [manual create]
    or
EDIT/DERIVE Decision → BusinessCapabilityRevision(DRAFT, source=DERIVED_EDIT)
    ↓
Human CONFIRM / EDIT_CONFIRM / REJECT  (AI Proposal path)
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

### 3.4 Capability field Source of Truth (LOCKED)

| Concern | SoT | Notes |
|---------|-----|-------|
| Stable identity | `capability_id == asset_id` | Single identity namespace (§4.9) |
| Tenant / Business ownership | **BusinessCapability** aggregate | `tenant_id`, `business_id` |
| Active pointer | **BusinessCapability** aggregate | `active_revision_id` only |
| Versioned semantics (`name`, `description`, `capability_kind`, schemas, bindings, preconditions, effects, business_model pin, …) | **BusinessCapabilityRevision** | Aggregate MUST NOT duplicate these as mutable SoT fields |
| List / search / display header | **AssetRef** (`forma_asset_ref`) | **Projection only** — must not become a second semantic SoT |
| Revision lifecycle status | **BusinessCapabilityRevision.status** | `DRAFT` / `VALIDATED` / `ACTIVE` / `STALE` / `DEPRECATED` |
| Asset Registry lifecycle status | **AssetRef.Status** | Asset Registry enum — **distinct** from Revision status |

**Real AssetRef fields (aligned to `asset_registry/entity.AssetRef`):**
`Name`, `SemanticVersion`, `Revision` (int32), `SchemaVersion`, `ContentDigest`, `Status`, `Kind`, `OwnerID`, `CreatedBy`, timestamps.
**There is no `Description` column on AssetRef.** Capability `description` lives only on **BusinessCapabilityRevision**.
**Forbidden:** projecting `description` onto AssetRef unless a future gate explicitly authorizes an Asset Registry schema/migration.

#### 3.4.1 `ProjectCapabilityAssetRef` — complete deterministic function (LOCKED)

V1 uses a **single stable AssetRef header row** per Capability (`capability_id == asset_id`).
**Forbidden:** inserting a new AssetRef header row on each Capability Activate or each new DRAFT.
**Forbidden:** treating `AssetRef.Revision` as an activation counter.
**Forbidden:** deciding AssetRef projection from “the last lifecycle action” alone.

**Unique function:**

```
ProjectCapabilityAssetRef(capability, revisions)
  → { Name, SemanticVersion, ContentDigest, Status, … }
  | stable consistency error  (no AssetRef write)
```

Inputs: the Capability aggregate (`capability_id`, `tenant_id`, `active_revision_id`, …) and the **full** set of `BusinessCapabilityRevision` rows belonging to that aggregate.

**Preconditions / consistency errors (evaluate before selection; on error do NOT write AssetRef; abort the enclosing lifecycle transaction with full rollback):**

| Condition | Result |
|-----------|--------|
| `revisions` is empty | Stable **consistency error** |
| Any duplicate `version` within the Capability | Stable **consistency error** |
| **ACTIVE pointer / ACTIVE rows invariant broken** (see below) | Stable **consistency error** |
| `active_revision_id` is non-null **and** the pointed revision is missing, belongs to another aggregate / tenant / capability, or does **not** have status `ACTIVE` | Stable **consistency error** (covered by ACTIVE invariant) |

**ACTIVE pointer ↔ ACTIVE rows invariant (LOCKED — indivisible aggregate consistency):**

Let `ActiveSet` = `{ r ∈ revisions | r.status == ACTIVE }`.

1. If `active_revision_id == null`: `ActiveSet` **MUST be empty**.
2. If `active_revision_id != null`:
   - `ActiveSet` **MUST contain exactly one** revision;
   - that revision **MUST** belong to the same `tenant_id` / `capability_id` aggregate;
   - that revision’s `revision_id` **MUST equal** `active_revision_id`.
3. Violations (orphan ACTIVE with null pointer; multiple ACTIVE; pointer points to A while the sole/other ACTIVE is B; pointer non-null but ActiveSet empty; etc.) → stable **consistency error**, **no AssetRef write**, **entire lifecycle transaction rollback**.

**Database uniqueness (LOCKED for G2 schema):**

```
UNIQUE (tenant_id, capability_id, version)
```

Selection of “maximum version” MUST use the numeric `version` field only — **never** database return order / insertion order / `created_at`.

**Deterministic selection (after preconditions pass; evaluate top-down; first match wins):**

| Priority | Aggregate condition | Projected revision `R` | AssetRef.Status |
|----------|---------------------|------------------------|-----------------|
| 1 | Legal ACTIVE pointer (`active_revision_id` non-null and ActiveSet is exactly that ACTIVE revision) | that ACTIVE revision | `RELEASED` |
| 2 | No legal ACTIVE; at least one `VALIDATED` exists | `VALIDATED` with **maximum** `version` | `VERIFIED` |
| 3 | No ACTIVE / VALIDATED; at least one `DRAFT` exists | `DRAFT` with **maximum** `version` | `DRAFT` |
| 4 | No ACTIVE / VALIDATED / DRAFT; at least one `STALE` exists | `STALE` with **maximum** `version` | `IN_REVIEW` |
| 5 | All remaining revisions are `DEPRECATED` | `DEPRECATED` with **maximum** `version` | `DEPRECATED` |

**STALE + DEPRECATED mixed sets:** Priority 4 applies whenever any `STALE` exists and there is no ACTIVE / VALIDATED / DRAFT — **DEPRECATED rows may coexist** with STALE; they do not block STALE selection.

**Hard invariant:** `AssetRef.Status == RELEASED` **requires** a legal ACTIVE pointer.
**Forbidden:** `RELEASED` + empty/null `active_revision_id`.

**Field sources from chosen `R`:**

| AssetRef field | Exact source |
|----------------|--------------|
| `Name` | `R.name` |
| `SemanticVersion` | SemVer string `0.{N}.0` where `N = R.version` (example: version `3` → `"0.3.0"`). Display projection only — **not** revision identity |
| `Revision` | Always `1` after Asset create; **never** modified by lifecycle updates |
| `SchemaVersion` | Fixed `"1.0"` at Asset create; **unchanged** for all V1 lifecycle projection updates |
| `ContentDigest` | Hex-encoded SHA-256 of the **canonical serialization** of `R`’s **behavioral semantic** payload (§3.4.2) |
| `Kind` | always `CAPABILITY` |
| `Status` | from the selection table above |

**Capability revision identity / version SoT:** only `BusinessCapabilityRevision` (`revision_id`, `version`, …).
**CozeResourceRef.AssetRevision:** continues to reference the stable `AssetRef.Revision = 1`.

#### 3.4.2 ContentDigest semantic coverage (LOCKED)

`ContentDigest` digests **only** the projected Revision’s **Capability behavioral semantics**.

**Included (must all be covered):**

- `name`, `description`, `capability_kind`
- `business_model_revision`
- `input_schema`, `output_schema`
- `preconditions[]`, `effects[]`
- `data_contract_bindings[]`
- `query_operation`, `output_cardinality` (QUERY)

**Excluded (must NOT affect digest) — provenance / identity / audit:**

- `source` (`MANUAL_CREATED` / `AI_PROPOSAL` / `DERIVED_EDIT`)
- `derived_from_revision_id`, `analysis_run_id`, `proposal_id`
- `revision_id`, `version`, `status`
- `capability_id`, `tenant_id`, `business_id` (aggregate identity)
- `created_by`, `created_at`, and any other audit / clock fields
- AssetRef fields themselves

**Provenance neutrality (LOCKED):** identical Capability behavioral semantics MUST produce the **same** `ContentDigest`, whether the Revision originated as `MANUAL_CREATED`, `AI_PROPOSAL`, or `DERIVED_EDIT`.

**Canonicalization:** exact byte serialization / field ordering / digest versioning is **fixed in G2** and MUST align with existing Forma digest conventions if already present in the platform.
**Determinism:** identical Included semantic payloads MUST produce identical `ContentDigest` values.

#### 3.4.3 Capability aggregate lock & projection transaction (LOCKED)

All operations that allocate a Capability-unique `version` and/or execute `ProjectCapabilityAssetRef` MUST:

1. Hold the **Capability aggregate lock** (see first-create vs existing below)
2. Read the current revision set **under that lock**
3. Allocate next unique `version` (enforced also by `UNIQUE (tenant_id, capability_id, version)`)
4. Persist Revision / pointer mutations
5. Run `ProjectCapabilityAssetRef`; on consistency error → abort the transaction (**no** AssetRef write; full rollback)
6. Commit in the **same** transaction

Applies to: Confirm / EDIT_CONFIRM (after target Capability is known), DeriveRevision, Validate, Activate, STALE, Deprecate, MANUAL_CREATE.

**Confirm / EDIT_CONFIRM ordering:**

1. Lock **Proposal** first (unchanged)
2. Resolve target Capability (first-create vs existing)
3. **Existing Capability:** `Proposal lock → Capability aggregate FOR UPDATE →` read revisions → allocate version → create Revision / Decision → terminalize Proposal → `ProjectCapabilityAssetRef` → commit
4. **First-create (LOCKED — single implementable seam; no vague “new aggregate lock path”):**

```
BEGIN
  1. INSERT BusinessCapability with unique PK (capability_id == asset_id)
     — success means THIS transaction owns the new aggregate
  2. INSERT the single AssetRef (Revision=1, SchemaVersion="1.0")
  3. INSERT the initial DRAFT BusinessCapabilityRevision
  4. INSERT CapabilityDecision; terminalize Proposal
  5. ProjectCapabilityAssetRef(capability, revisions)
  6. COMMIT
```

- Any step failure → **full rollback** (no orphan Capability, AssetRef, Revision, Decision, or terminal Proposal)
- Identity / AssetRef unique-key conflict → stable **conflict**; **no** orphan leftovers
- Concurrent first-creates racing the same intended `capability_id` / `asset_id`: at most one INSERT wins; losers receive stable conflict

**Concurrent Confirms** from two different Proposals targeting the **same** existing Capability: both may succeed only as distinct DRAFT versions under the aggregate lock; each recomputes projection; if an ACTIVE already exists, RELEASED Name / SemVer / ContentDigest remain those of the ACTIVE.

#### 3.4.4 When to recompute projection (LOCKED)

After **every** successful Confirm / Derive / Validate / Activate / STALE / Deprecate (and MANUAL_CREATE), within the **same transaction and Capability aggregate lock**, recompute `ProjectCapabilityAssetRef` and write Name / SemanticVersion / ContentDigest / Status (`Revision` / `SchemaVersion` stay `1` / `"1.0"`).

**Existing Capability + new DRAFT:** do not create a second AssetRef; recompute projection; ACTIVE RELEASED projection must not change.

**First-time create only:** follow the first-create INSERT seam in §3.4.3, then project.

#### 3.4.5 Static verification matrix (architecture acceptance — G2 tests MUST cover)

| Case | Expected |
|------|----------|
| STALE + DEPRECATED mixed (no ACTIVE/VALIDATED/DRAFT) | Project max-version STALE; Status=`IN_REVIEW` |
| Invalid / dangling / cross-aggregate / non-ACTIVE `active_revision_id` | Consistency error; AssetRef unchanged; txn rollback |
| Empty `revisions` | Consistency error; AssetRef unchanged; txn rollback |
| Duplicate `version` values | Consistency error; AssetRef unchanged; txn rollback |
| `active_revision_id == null` + orphan ACTIVE row | Consistency error; AssetRef unchanged; txn rollback |
| Legal pointer to A + second ACTIVE row | Consistency error; AssetRef unchanged; txn rollback |
| Pointer points to A while ActiveSet is B (or includes B ≠ A) | Consistency error; AssetRef unchanged; txn rollback |
| Two Proposals concurrently Confirm to one existing Capability | Both create distinct DRAFTs under aggregate lock + UNIQUE version; no duplicate AssetRef |
| ACTIVE present while concurrent DRAFTs are created | Projection remains ACTIVE / `RELEASED` Name·SemVer·ContentDigest unchanged |
| Concurrent first-create race on same capability_id/asset_id | One wins; loser stable conflict; no orphans |

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
| AI | Create `CapabilityAnalysisRun` and `CapabilityProposal(PROPOSED)` | Confirm, EditConfirm, Reject, Edit/Derive, Validate, Activate, Deprecate, create Revision, alter ACTIVE |
| Human | CONFIRM / EDIT_CONFIRM / REJECT / MANUAL_CREATE / EDIT (DERIVE) / Validate / Activate / Deprecate | — |

Every human CONFIRM / EDIT_CONFIRM / REJECT / CREATE / EDIT (DERIVE) / Activate / Deprecate MUST produce an immutable `CapabilityDecision` (or lifecycle) audit record.

### 4.3 IMMUTABLE_REVISION

**All persisted** `BusinessCapabilityRevision` semantic content is immutable, **including DRAFT**.

- Edits to any revision create a **new** DRAFT with `derived_from_revision_id` via the **DERIVED_EDIT** path (§8.5 / §9.1).
- Status transitions, active pointer updates, Asset header projection refresh, and audit/decision events are **not** semantic content mutation.
- **Forbidden:** UPDATE that overwrites `input_schema`, `output_schema`, bindings, preconditions, effects, name, description, kind, or other semantic fields on an existing revision row.

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
- Asset Registry owns **header lifecycle** (projection + Asset status).
- **BusinessCapability** aggregate persists only: identity, tenant/business ownership, `active_revision_id`, audit timestamps / created_by — **not** versioned semantic fields and **not** a duplicate generic Capability `status`.
- Capability Domain owns **revision lifecycle** and **`active_revision_id`**.
- Create / Confirm failure MUST NOT leave an orphan Asset or Capability; G2 MUST use a transactional Unit of Work or equivalent idempotent compensation (§9.8).

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
| 1 | **BusinessCapability** | Canonical CAPABILITY aggregate: `capability_id == asset_id`, tenant/business ownership, `active_revision_id` |
| 2 | **BusinessCapabilityRevision** | **SoT** for versioned capability semantics (name/description/kind/schemas/bindings/…) |

**Asset Registry note:** `AssetKindCapability = "CAPABILITY"` already exists in S0 Asset Registry taxonomy.
`capability_id == asset_id` is **LOCKED** (§4.9).
Asset Registry header is a **retrieval/display projection**, not a second semantic SoT (§3.4).

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
| `CapabilityAnalysisRun` | AI analysis execution binding; idempotency key + `request_digest`; status `PENDING` / `SUCCEEDED` / `FAILED` (§9.5) |
| `CapabilityProposal` | Workflow record for AI proposals; **not** a Canonical Asset |
| `CapabilityDecision` | Immutable human decision event |
| `CapabilityValidationResult` | Technical validation outcome for a revision |
| `CapabilityImpactResult` | Impact / gap / invalidation outcome when upstream Business Model or Contract changes |

These must not redefine Business Model or Data Contract semantics.

### 6.4 Suggested Identity Fields (informative for G2+)

**BusinessCapability (aggregate — non-semantic):**
`capability_id` (== `asset_id`), `tenant_id`, `business_id`, `active_revision_id` (nullable), `created_by`, `created_at`, `updated_at`
(**No** `name` / `description` / `capability_kind` / duplicate `status` on the aggregate.)

**BusinessCapabilityRevision (semantic SoT):**
`revision_id`, `capability_id`, `tenant_id`, `business_id`, `version`, `status`, `name`, `description`, `business_model_revision`, `capability_kind`, `input_schema`, `output_schema`, `preconditions[]`, `effects[]`, `data_contract_bindings[]`, `query_operation` (QUERY), `output_cardinality` (QUERY), `derived_from_revision_id` (nullable), `analysis_run_id` (nullable), `proposal_id` (nullable), `source` (`AI_PROPOSAL` / `MANUAL_CREATED` / `DERIVED_EDIT`), `created_by`, `created_at`

**CapabilityProposal:**
`proposal_id`, `tenant_id`, `business_id`, `analysis_run_id`, `capability_id` (nullable until CONFIRM/EDIT_CONFIRM binds), `status` (`PROPOSED` / `REJECTED` / `CONFIRMED` / `EDIT_CONFIRMED`), `payload` (**immutable** proposed semantic content; never updated after create), `materialized_revision_id` (nullable), `created_at`

**CapabilityDecision:**
`decision_id`, `tenant_id`, `business_id`, `capability_id` (nullable **only** for unbound-Proposal REJECT — §9.9), `proposal_id` (required when `capability_id` is null; required for CONFIRM/EDIT_CONFIRM), `source_revision_id` (required for EDIT/DERIVE), `target_revision_id` (required for EDIT/DERIVE; set for CONFIRM/EDIT_CONFIRM materialization), `action`, `payload_digest` (required for CONFIRM / EDIT_CONFIRM / DERIVE — see §8.5 / §9.8), `client_request_id` (request tracing only for Confirm; part of Derive logical key), `actor_principal_id`, `reason` (nullable), `created_at`
Decisions are immutable after create. AI must not create human decisions.

**`payload_digest` meaning (LOCKED — single field):**

| Action | `payload_digest` value |
|--------|------------------------|
| CONFIRM / EDIT_CONFIRM | Digest of the **effective payload** used to materialize the Revision |
| EDIT / DERIVE | Digest of the **target semantic payload** / derive request body that created the target DRAFT |

**Forbidden:** storing a derive `request_digest` into a differently named `effective_payload_digest` column, or maintaining two conflicting digest field names for the same Decision row.

**CapabilityAnalysisRun:**
`analysis_run_id`, `tenant_id`, `business_id`, `business_model_revision`, `client_request_id`, `request_digest`, `status` (`PENDING` / `SUCCEEDED` / `FAILED`), `attempt` (monotonic), `error_code` (nullable, sanitized), `created_at`, `updated_at`, …

---

## 7. Capability Kinds & Boundaries

### 7.1 Domain-agnostic kinds only

| Kind | Meaning | V1 |
|------|---------|----|
| **QUERY** | Read via S4 logical Data Contract (`READ` / `LIST` / `FILTER`; `LOOKUP` / `AGGREGATE` deferred — §10.1.3) | **In scope** |
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
| `VALIDATED` | Passed technical validation against pinned Business Model revision + Data Contract logical bindings; Validate prerequisite satisfied (§9.6) |
| `ACTIVE` | Consumers may pin this revision; at most one ACTIVE revision per Capability aggregate |
| `STALE` | Upstream pin invalidated per §8.3; not auto-repaired |
| `DEPRECATED` | Retired (human Deprecate, or superseded ACTIVE on successful Activate) |

**Illegal:** `VALIDATED → DEPRECATED` without passing ACTIVE; `STALE → ACTIVE` in place; `DRAFT → ACTIVE` skipping VALIDATED.

### 8.2 Transition rules

| From | To | Actor | Notes |
|------|----|-------|-------|
| — | DRAFT | Human CONFIRM / EDIT_CONFIRM (Proposal) · MANUAL_CREATE · EDIT/DERIVE | See §9; AI never creates Revision |
| DRAFT | VALIDATED | Human-triggered Validate | Requires §9.6 prerequisite + successful `CapabilityValidationResult`; under Capability aggregate lock run `ProjectCapabilityAssetRef` in **same** transaction (§3.4) |
| VALIDATED | ACTIVE | Human Activate | Single transaction per §8.4 (+ Asset header projection §3.4) |
| ACTIVE | DEPRECATED | Human Deprecate **or** Activate supersession | Auditable |
| ACTIVE | STALE | System/human impact evaluation | Only via §8.3 triggers + recorded `CapabilityImpactResult` |
| STALE | DEPRECATED | Human | Retirement after invalidation |

Repair path after STALE/DEPRECATED: human creates a **new** DRAFT via EDIT/DERIVE (`derived_from_revision_id` set); never revive STALE content in place as ACTIVE.

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

Activating a new VALIDATED revision **MUST** occur in **one transaction** under the Capability aggregate lock (§3.4.3):

1. Lock the Capability aggregate
2. Current old ACTIVE revision (if any) → `DEPRECATED`
3. New revision → `ACTIVE`
4. `active_revision_id` → new revision
5. Recompute `ProjectCapabilityAssetRef(capability, revisions)` and write AssetRef projection (§3.4); on consistency error → abort; `Revision` remains `1`, `SchemaVersion` remains `"1.0"`
6. Write lifecycle / decision audit

When the current ACTIVE becomes `STALE` or `DEPRECATED`:

- Clear `active_revision_id` **only if** the active pointer currently references that revision
- Handling historical STALE / former ACTIVE revisions **MUST NOT** clear a newer ACTIVE pointer

**Concurrency:** concurrent Activate of two VALIDATED revisions → **at most one succeeds**; loser receives conflict error.
Optimistic concurrency (expected revision / generation) or equivalent transactional guard is required for Activate / Deprecate / STALE.

### 8.5 DeriveRevision atomic interface (LOCKED — single edit model)

**Unique model for editing an existing persisted Revision** (including DRAFT / VALIDATED / ACTIVE / STALE / DEPRECATED sources):

```
OWNER/ADMIN DeriveRevision
  → lock Capability aggregate + source Revision
  → atomically create target DRAFT (source=DERIVED_EDIT)
       + CapabilityDecision(EDIT|DERIVE)
         with source_revision_id + target_revision_id + payload_digest
  → ProjectCapabilityAssetRef(capability, revisions)
```

**Unit of Work (mandatory):**

1. Lock Capability aggregate and the source `BusinessCapabilityRevision` (§3.4.3)
2. Under that lock: read revisions; allocate Capability-unique next `version` (`UNIQUE (tenant_id, capability_id, version)`)
3. Create immutable target `BusinessCapabilityRevision(DRAFT, source=DERIVED_EDIT, derived_from_revision_id=source)`
4. Create immutable `CapabilityDecision(action=EDIT|DERIVE)` with `source_revision_id`, `target_revision_id`, and `payload_digest`
5. Recompute `ProjectCapabilityAssetRef` (§3.4). Consistency error → abort. If an ACTIVE already exists, RELEASED Name/SemVer/ContentDigest **must not** change
6. Commit

**Failure:** any step fails → full rollback; no orphan DRAFT or Decision.

**Idempotency (LOCKED — key and digest separated):**

- **Logical key:** `tenant_id + capability_id + source_revision_id + client_request_id`
- **`payload_digest`:** persisted on the Decision; covers the target semantic payload / derive request body

| Case | Behavior |
|------|----------|
| Same key + same digest | Return the original target Revision (and original Decision); **no** duplicate |
| Same key + different digest | Stable **idempotency conflict** |
| Concurrent same key | At most one create; others observe winner or conflict |

**Concurrency:** concurrent DeriveRevision with distinct keys may create distinct derived DRAFTs.

Rules:

- Only `OWNER` or `ADMIN` (ACTIVE membership) may EDIT/DERIVE.
- New DRAFT semantic content is written **once** at create; never updated in place.
- **Do not** route ordinary human edits of existing revisions through Proposal + EDIT_CONFIRM.
- Proposal + EDIT_CONFIRM remains **only** for AI Proposal materialization (§9.1).
- AI must never create DERIVED_EDIT revisions.

This is the **only** allowed non-Proposal edit model. Do not introduce a second parallel edit seam.

---

## 9. Proposal / Revision Seam & AI Propose / Human Confirm (LOCKED)

### 9.1 Unique Proposal → Revision model (AI path) + Manual / Derive paths

**AI Proposal path:**

```
CapabilityAnalysisRun
  → CapabilityProposal(PROPOSED)
  → Human CONFIRM / EDIT_CONFIRM
  → materialize immutable BusinessCapabilityRevision(DRAFT, source=AI_PROPOSAL)

Human REJECT
  → CapabilityProposal(REJECTED)
  → no Revision created
  → no Asset / Capability shell created (§9.9)
```

**Manual create path:**

```
Human MANUAL_CREATE + CapabilityDecision(CREATE)
  → BusinessCapabilityRevision(DRAFT, source=MANUAL_CREATED)
  (Asset + Capability bound in the same create Unit of Work)
```

**Derived edit path:** §8.5 (`source=DERIVED_EDIT`).

- `CapabilityProposal` is a **workflow record**, not a Canonical Asset.
- AI may **only** create Proposals (via AnalysisRun).
- AI **must not** directly create a Revision that can be Validated or Activated.
- CONFIRM, EDIT_CONFIRM, REJECT, CREATE, EDIT/DERIVE **must** generate an immutable `CapabilityDecision`.

**Forbidden:** AI proposal materialized directly as DRAFT without CONFIRM / EDIT_CONFIRM.

### 9.2 AI may

- Create `CapabilityAnalysisRun` and `CapabilityProposal(PROPOSED)` from Business Model revision + Active Data Contract logical descriptors
- Suggest input/output schemas, preconditions, effects, and contract bindings inside Proposal payload
- Produce analysis metadata (confidence, reason) — **confidence ≠ confirmation**

### 9.3 AI must not

- CONFIRM / EDIT_CONFIRM / REJECT / EDIT / DERIVE
- Validate / Activate / Deprecate
- Create `BusinessCapabilityRevision`
- Silently mutate ACTIVE Capability
- Create `CapabilityDecision` human records
- Re-call models except via explicit FAILED retry rules (§9.5)

### 9.4 Provenance

AI-originated DRAFT revisions (after human CONFIRM / EDIT_CONFIRM) MUST trace to:

- `CapabilityAnalysisRun`
- `CapabilityProposal`
- Explicit `business_model_revision`
- Input sources (contract ID/version pins and/or requirement refs when present on the AnalysisRun)
- Human `CapabilityDecision` (CONFIRM or EDIT_CONFIRM)

DERIVED_EDIT DRAFT revisions MUST trace to EDIT/DERIVE Decision with `source_revision_id` + `target_revision_id`.

### 9.5 CapabilityAnalysisRun lifecycle & idempotency (LOCKED)

**Statuses (frozen):** `PENDING` → `SUCCEEDED` | `FAILED`

Minimum logical key:

```
tenant_id + business_id + business_model_revision + client_request_id
```

Persist `request_digest` covering at least:

- `business_model_revision`
- Sorted Data Contract ID/version pins
- Input requirement refs
- Analysis options

**Ordinary replay (same key + same digest):** always return **this** `analysis_run_id`’s **current** status and payload. Never create a second AnalysisRun for the same key+digest.

| Existing status | Ordinary replay behavior |
|-----------------|--------------------------|
| `PENDING` | Return the same run; **do not** start a second model call; waiter observes completion |
| `SUCCEEDED` | Return the same run + **that run’s unique Proposal set**; **no** new model call |
| `FAILED` | Return the FAILED run; **do not** auto-retry |

**Same key + different digest:** return stable **idempotency conflict** error (all statuses).

**FAILED explicit retry (V1 — single run only):**

- Requires an explicit retry API/action (not a silent duplicate submit).
- **Must** reuse the **same** `analysis_run_id` (no linked retry row).
- Atomically: `FAILED` → `PENDING` and increment `attempt` (+ attempt/audit record).
- Then at most one model executor runs; concurrent retries wait or receive conflict.
- On success: `PENDING` → `SUCCEEDED` with the unique Proposal set for this run.
- On failure: `PENDING` → `FAILED`.

**Forbidden:** creating a second AnalysisRun / “linked retry row” for the same key+digest.

**Concurrency:** concurrent identical first requests → exactly one model execution acquires `PENDING`.

**Proposal materialization** for one AnalysisRun must be idempotent; retries must not create duplicate Proposals or Revisions.

### 9.6 Validate prerequisites (LOCKED)

A DRAFT may be Validated **only if** one of:

1. Materialized via human CONFIRM or EDIT_CONFIRM from a Proposal (`source=AI_PROPOSAL`); **or**
2. `source=MANUAL_CREATED` with a recorded human CREATE decision; **or**
3. `source=DERIVED_EDIT` with a recorded human EDIT/DERIVE decision that includes `source_revision_id` and `target_revision_id`

**Forbidden:**

- Validating an AI Proposal directly (Proposal is not a Revision)
- Validate or Activate implicitly substituting for Confirm / Edit / Create
- Skipping Confirm on the AI path
- Validating a DERIVED_EDIT DRAFT without EDIT/DERIVE Decision provenance

### 9.7 Model adapter path (future gates)

```
Forma Capability Domain
  → Forma CrossDomain / ACL
  → Coze/Eino Model Manager
```

**Forbidden in Capability Domain:** provider-specific SDKs (OpenAI / DeepSeek / Qwen / …).

**S5-G1 / S5-G1-F1 … G1-F7: REAL_MODEL_CALLS = 0**

### 9.8 ConfirmProposal atomic boundary & replay (LOCKED)

CONFIRM / EDIT_CONFIRM **MUST** execute as **one Unit of Work**.

**Payload rules (LOCKED):**

- `CapabilityProposal.payload` is **immutable** after Proposal create (AI or system must not rewrite it).
- **CONFIRM** effective payload = `Proposal.payload` (caller must not substitute a different body).
- **EDIT_CONFIRM** MUST carry a **complete** `effective_payload` (full semantic document, not a patch). The materialized Revision is created from that `effective_payload`.
- Persist **`payload_digest`** = digest of the effective payload used for materialization on the `CapabilityDecision` (same field name as Derive — §6.4 / §8.5).

**Idempotency identity (LOCKED):**

- A Proposal may terminalize **once**. Materialization identity is the Proposal itself.
- **Logical key:** `tenant_id + proposal_id`
- **`client_request_id`:** request tracing only — **does not** change materialization identity. Different `client_request_id` values still obey the rules below.
- **Action:** `CONFIRM` or `EDIT_CONFIRM`
- **Digest:** `payload_digest`

| Case | Behavior |
|------|----------|
| Terminal + same action + same digest | Return the original `materialized_revision_id` / Revision; **no** new Asset/Capability/Revision/Decision |
| Terminal + different action and/or different digest | Stable **conflict** / **idempotency conflict** |
| Non-terminal (`PROPOSED`) first success | First materialization (steps below) |

**Lock rule:** begin transaction and **lock the Proposal row first**. Do **not** require the pre-lock observed status to be `PROPOSED` before locking; decide behavior **after** the lock using the locked row.

**Post-lock behavior:**

| Locked Proposal status | Requested action | Behavior |
|------------------------|------------------|----------|
| `PROPOSED` | CONFIRM or EDIT_CONFIRM | **First materialization** (steps below) |
| `CONFIRMED` | CONFIRM | Same digest → replay original Revision; different digest → idempotency conflict |
| `EDIT_CONFIRMED` | EDIT_CONFIRM | Same digest → replay; different digest → idempotency conflict |
| `CONFIRMED` | EDIT_CONFIRM (or vice versa) | Stable **conflict** (terminal status ≠ requested action) |
| `REJECTED` | CONFIRM or EDIT_CONFIRM | Stable **conflict** |
| `CONFIRMED` / `EDIT_CONFIRMED` | matching action | If `materialized_revision_id` **or** the corresponding Decision is **missing** → stable **consistency error** |
| `REJECTED` | REJECT (see §9.9) | If REJECT Decision is **missing** → stable **consistency error**; `materialized_revision_id` MUST remain null |

**First materialization steps (only when locked status is `PROPOSED`):**

1. Resolve effective payload (CONFIRM → Proposal.payload; EDIT_CONFIRM → request `effective_payload`) and compute `payload_digest`
2. Resolve target Capability (first-create vs existing bind declared by payload)
3. **Existing Capability:** `Capability aggregate FOR UPDATE` (§3.4.3). Under that lock: read revisions; allocate next `version`; create DRAFT Revision only (**no** second AssetRef); create Decision; terminalize Proposal; `ProjectCapabilityAssetRef` (ACTIVE RELEASED projection must not change); consistency error → full rollback
4. **First-create:** follow the INSERT seam in §3.4.3 (unique BusinessCapability PK insert owns the aggregate → AssetRef → initial Revision → Decision → terminalize Proposal → project). Conflict / any failure → full rollback; no orphans
5. Commit

**Failure:** any step fails → **no** partial Asset, Capability, Revision, Decision, or Proposal terminal update remains (rollback / idempotent compensation).

**Forbidden:** duplicate Asset, Capability, Revision, or Decision on concurrency or replay.

### 9.9 RejectProposal atomic interface (LOCKED)

RejectProposal **MUST** execute as **one Unit of Work**:

1. Lock the `CapabilityProposal` row
2. Post-lock:
   - If status is `PROPOSED`: write immutable `CapabilityDecision(REJECT)` and set Proposal → `REJECTED` atomically
   - If status is already `REJECTED`: idempotent replay → return the original REJECT Decision; **no** second Decision
   - If status is `CONFIRMED` or `EDIT_CONFIRMED`: return stable **conflict**
3. Commit

**Invariants:**

- **No** Revision is created on Reject.
- **Forbidden:** creating an empty shell Asset or Capability solely for Reject audit.
- **Forbidden:** Proposal status `REJECTED` without a corresponding REJECT Decision.
- `capability_id` may be **null only** on a REJECT Decision for a Proposal that was never bound to a Capability.
- When `capability_id` is null on a Decision: `proposal_id` is **required**.
- CONFIRM / EDIT_CONFIRM Decisions always have non-null `capability_id` (§9.8).

---

## 10. Schema Compatibility & Controlled Expressions

### 10.1 V1 QUERY schema compatibility (LOCKED)

Capability `input_schema` / `output_schema` are **logical** schemas.
Compatibility direction: **Capability requirements ⊆ Contract guarantees** expressed by the real S4 **DataContractDescriptor** only:

`logical_schema`, `query_capabilities`, `filter_schema`, `sort_schema`, `pagination_policy`
(**No** `LookupSchema`. **No** Contract-level required/optional input flags beyond `LogicalField.nullable` and FilterSchema membership.)

#### 10.1.1 Caller obligations vs Contract fields

- Capability `required` / `optional` on **input** define **caller obligations only** (what the invoker must/may supply). They are **not** copies of S4 descriptor required/optional metadata (which does not exist as such).
- Every Capability input binding MUST reference an existing `LogicalSchema.fields[].logical_key` on the pinned Contract.
- Every Capability output binding MUST reference an existing `logical_key` on the pinned Contract.
- Logical types **MUST** use **S4 normalized logical types**
- V1 **disallows** implicit type conversion (exact normalized type match required)
- Capability **required** output → Contract field MUST exist and `nullable=false`
- Capability **nullable** output → may bind to nullable or non-null Contract field
- Missing binding, unknown `logical_key`, type mismatch, or weakened nullability/type guarantees → **Validation FAIL**

#### 10.1.2 Operations, filters, sort, pagination (LOCKED per operation)

V1 Capability QUERY supports only operations that are **mechanically verifiable** from the S4 descriptor: **`READ`**, **`LIST`**, **`FILTER`**.

- Requested QUERY operation MUST be listed in Contract `query_capabilities`
- Sort field/direction, when used, MUST be a **subset** of Contract `sort_schema`
- Pagination limit, when used, MUST NOT exceed Contract `pagination_policy.max_limit`

**FILTER**

- MUST declare **at least one required** Capability input
- MUST declare **at least one actual filter predicate** whose value is supplied by a **required** Capability input
- Every filter field/operator used by the Capability MUST belong to Contract `filter_schema` (field present; operator allowed for that field)
- Optional Capability inputs may bind additional FilterSchema fields but cannot replace the required-input + required-predicate rules above

**READ** (`output_cardinality=ONE`)

- MUST NOT require FilterSchema membership for success when no filters are declared
- If any filter predicates are declared, each field/operator MUST belong to Contract `filter_schema`
- Optional inputs, if present, MUST bind to existing `logical_schema` fields
- Pagination: if present, `limit` MUST be `1` and MUST NOT exceed Contract max limit; otherwise omit pagination

**LIST** (`output_cardinality=MANY`)

- Filter predicates are optional; if present, each field/operator MUST belong to Contract `filter_schema`
- Optional/required inputs that act as filters MUST bind to FilterSchema fields; other inputs MUST still bind to `logical_schema` fields
- Pagination MAY be declared; `limit` MUST be within Contract `pagination_policy` (default may follow Contract default_limit)
- Sort, if declared, MUST be a subset of Contract `sort_schema`

#### 10.1.3 LOOKUP / AGGREGATE (V1 deferred)

`DataContractDescriptor` has **no** lookup-key / LookupSchema surface and **no** aggregate function / grouping / aggregate output-type contract.

**V1 decision:** Capability `query_operation` values **`LOOKUP`** and **`AGGREGATE`** are **DEFERRED**. Declaring either → **Validation FAIL** until a later gate extends the S4 descriptor with explicit contracts (not invented in S5).

#### 10.1.4 Output cardinality (frozen enum — V1)

| `output_cardinality` | Allowed V1 `query_operation` values |
|----------------------|-------------------------------------|
| `ONE` | `READ` |
| `MANY` | `LIST`, `FILTER` |

Rules:

- Every QUERY revision MUST declare exactly one `query_operation` and exactly one `output_cardinality`.
- Pairing outside the table above (including `LOOKUP` / `AGGREGATE`) → **Validation FAIL**.
- `ONE` means at most one logical result entity; `MANY` means zero-or-more.

Any incompatibility above → **Validation FAIL** (no warnings-as-pass).

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
| EDIT / DERIVE (DERIVED_EDIT) | `OWNER`, `ADMIN` |
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

**S5-G1 / S5-G1-F1 … G1-F7: REAL_MODEL_CALLS = 0**

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
| 5 | HUMAN_DECISION_PROVENANCE | Human lifecycle actions produce immutable decisions; Confirm / Create / Derive required before Validate as applicable |
| 6 | IMMUTABLE_REVISION | No in-place mutation of **any** persisted revision semantic content (including DRAFT) |
| 7 | TENANT_ISOLATION | Tenant A/B fully isolated; cross-tenant refs denied |
| 8 | ROLE_AUTHORIZATION | Actions enforced server-side via exact role matrix (§11.2) |
| 9 | COMMAND_EXECUTION_BOUNDARY | COMMAND is intent only; no arbitrary write/exec in S5 Capability gates |
| 10 | NO_ARBITRARY_CODE | No JS/Python/SQL/Shell expressions in Capability definitions |
| 11 | IDEMPOTENCY | Same analysis / Confirm / Derive key + digest must not duplicate model work / proposals / revisions |
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
| **S5-G1-F1** … **S5-G1-F6** | Architecture Consistency Fixes |
| **S5-G1-F7** | Architecture Consistency Fix (**this amendment**) |
| **S5-G2** | Capability Domain, Revision, state machine, human decisions, idempotent analysis runs |
| **S5-G3** | Business Model / Data Contract binding & validation |
| **S5-G4** | Capability API, authorization, audit, concurrency consistency |
| **S5-G5** | Capability UI (`@forma/capability`, `/capabilities`) |
| **S5-G6** | Live E2E, Security, Browser Acceptance, Freeze |

Each gate: Implement → Test → Review → PASS.
**Forbidden:** implement entire S5 Business Capability in one pass.

**Roadmap note:** No separate formal S5 Business Capability roadmap document existed in-repo before G1. This section is the first locked roadmap for the Capability domain and aligns with the authorized S5-G1 brief. S5-G0 remains the platform-admin foundation gate already completed.

---

## 18. G1 / G1-F1 … G1-F7 Exit Criteria

- Stage Contract published at `forma/docs/stages/FORMA-S5-BUSINESS-CAPABILITY-STAGE-CONTRACT.md`
- G1 … G1-F6 Results under `forma/cursor-results/`
- G1-F7 Result at `forma/cursor-results/FORMA-S5-G1-F7-BUSINESS-CAPABILITY-ARCHITECTURE-RESULT.md`
- Repo issue tracker config at `docs/agents/issue-tracker.md`
- Docs-only commit; Forma CI ALL GREEN
- No Capability domain / migration / adapter / UI implementation
- `REAL_MODEL_CALLS = 0`
- `PRODUCT_CODE_CHANGE = NONE`
- `MIGRATION_CHANGE = NONE`
- `S4_FROZEN_BASELINE_UNCHANGED = PASS`
- Human architecture review required before S5-G2
- **DO NOT START S5-G2** until review PASS
- **DO NOT** create `forma-s5-frozen`
- **S5_G2_READY = NO** until human review after G1-F7

---

## 19. Deferred Decisions (explicit — not pretend-implemented)

| Item | Status |
|------|--------|
| COMPOSITE capability kind | Deferred |
| Full precondition/effect DSL language | Deferred |
| COMMAND runtime / Action Adapter | Deferred to later stage |
| Invoke QUERY/COMMAND runtime ACL | Deferred to Runtime gate |
| Consumer invoke ACL fine-grain field/row policies | Deferred to Runtime / later ACL gate |
| Capability `query_operation=LOOKUP` | Deferred until S4 descriptor exposes explicit lookup-key metadata |
| Capability `query_operation=AGGREGATE` | Deferred until S4 descriptor exposes aggregate function/grouping/output-type contract |
| AssetRef `Description` column | Deferred — requires explicit Asset Registry schema/migration authorization |
| ContentDigest canonical byte layout / digest-schema version id | Deferred to G2 implementation freeze under §3.4.2 rules |

---

## Document Control

| Field | Value |
|-------|-------|
| Document | FORMA-S5 Business Capability Stage Contract |
| Gate | S5-G1 / S5-G1-F1 … S5-G1-F7 |
| Baseline tags | `forma-s4-frozen-r2` |
| Related ADRs | ADR-002, ADR-006, ADR-013 |
| Related contracts | FORMA-S4 Data Plane / Data Contract Stage Contract |
| Related results | FORMA-S4-FINAL-FREEZE-R2, FORMA-S5-G0, FORMA-S5-G1 … F7 |
