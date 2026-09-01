# ADR-014: Immutable Business Model Revision

## Context

Semantic edits need auditability, optimistic concurrency, and safe collaboration across clients.

## Decision

Store each semantic snapshot in `forma_business_model_revision` as **immutable**.

- Changes create `revision_no = current + 1` with `base_revision_no`, `semantic_model_json`, `content_digest`, `change_summary`.
- CAS on `forma_business_model.current_revision` via `expected_revision`.
- Identical digest → no new revision (`no_change`).
- Never UPDATE existing revision content.

## Consequences

- Diff and history are first-class.
- Conflicts return `FORMA_BUSINESS_MODEL_CONFLICT`.
- Storage grows with semantic edits (acceptable for V1).

## Rejected Alternatives

- Mutable single-row semantic JSON; soft-delete revisions; shared revision counter with layout.
