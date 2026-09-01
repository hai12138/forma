# ADR-013: Business Model as Source of Truth

## Context

Forma maps business semantics to agents and capabilities. Without a single Source of Truth, Prompt, Coze Agent, Workflow, or Knowledge Graph tend to become de-facto models and diverge.

## Decision

**Business Model** (semantic graph: nodes, edges, rules, states) under a **BUSINESS** Asset is the Source of Truth.

- `business_id == asset_id` (one BUSINESS asset ↔ one Business Model aggregate).
- Capability / Agent / Application / Evaluation must derive from or reference Business Model later — not the reverse.
- Asset Registry owns lifecycle (DRAFT/ARCHIVED…); Business Model owns semantic revisions and layout.

## Consequences

- Editing business meaning creates immutable semantic revisions.
- Downstream projections (KG, workflows) are non-canonical.
- Create Business must transactionally create Asset + Model master + revision 1 + layout 1.

## Rejected Alternatives

- Prompt-as-SoT; Coze Agent/Workflow-as-SoT; Neo4j/KG-as-SoT; separate Business ID namespace unrelated to Asset Registry.
