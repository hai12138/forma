# ADR-019: Evidence as Immutable Provenance

## Status

Accepted (FORMA S3)

## Decision

`BusinessEvidence` rows are immutable. Content changes require a new Evidence record. Assertions link to Evidence via `forma_assertion_evidence_ref`.

## Consequences

- Interview turns always produce Evidence before Assertions.
- Confirmed assertions without Evidence are rejected server-side.
