# Business Model Architecture (S2)

## Overview

```
Forma Asset Registry
        │  kind = BUSINESS
        ▼
BUSINESS Asset  (lifecycle: DRAFT…ARCHIVED)
        │  business_id == asset_id
        ▼
Business Model aggregate
        ├── Semantic Revisions (immutable SoT)
        └── View Layout (presentation only)
```

Business Model is Forma’s **Business-to-Agent Source of Truth**. Prompt, Coze Agent, Workflow, and Knowledge Graph are **not** SoT.

## Semantic Model

`schema_version`, `nodes[]`, `edges[]`, `rules[]`, `states[]`, plus reserved `evidence_refs[]` / `assertion_refs[]` (skeleton only in S2).

Node types (formal): `ACTOR`, `BUSINESS_OBJECT`, `PROCESS`, `EVENT`, `DECISION`, `SYSTEM`, `POLICY` (+ v1.2 UI aliases accepted by validator).

Edge types (extensible): `PERFORMS`, `CREATES`, `UPDATES`, `TRIGGERS`, `REQUIRES`, `DEPENDS_ON`, `TRANSITIONS_TO`, `RELATES_TO`.

`source_marker`: `AI_GENERATED` | `MANUAL_MODIFIED` (S2 writes MANUAL_MODIFIED; AI Analyst later).

## Revision & Digest

Canonical serialize (sorted element IDs / keys) → SHA-256 `content_digest`. Layout excluded.

Optimistic concurrency: `expected_revision` / `expected_layout_revision`.

## Diff & Impact

Element-level diff (nodes/edges/rules/states Added|Removed|Modified).

`BusinessImpactSummary` (S2): semantic_changed + affected ids — no Capability/Agent impact yet.

## Storage

| Table | Role |
|-------|------|
| `forma_business_model` | Master pointer (`current_revision`) |
| `forma_business_model_revision` | Immutable semantic snapshots |
| `forma_business_model_layout` | Layout with own revision |

MySQL only. No Neo4j / Coze Knowledge / Workflow as canonical store.

## API

`/api/forma/v1/businesses` (+ model, revisions, diff, layout). All TenantContext-scoped; body `tenant_id` ignored as security source.

## Future Projections

Capability Gateway, Data Contracts, Agent Composer, Evaluation, and Knowledge Graph consume Business Model revisions — they do not replace them.
