# ADR-021: Proposal Before Business Model Revision

## Status

Accepted (FORMA S3)

## Decision

Semantic changes flow through `BusinessModelProposal` with `SemanticModelPatch` operations (not raw JSON Patch as product semantics). Apply validates full model via S2 validator and checks `base_revision` staleness.

## Consequences

- Proposal preview shows operations and base revision.
- Stale proposals cannot overwrite newer revisions.
