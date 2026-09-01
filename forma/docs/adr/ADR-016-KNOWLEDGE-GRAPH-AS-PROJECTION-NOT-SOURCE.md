# ADR-016: Knowledge Graph as Projection, Not Source

## Context

Graph databases are useful for query/visualization, but treating Neo4j (or Coze Knowledge) as authoritative Business Model recreates dual-write and SoT drift.

## Decision

Knowledge Graph (when introduced) is a **projection** of Business Model revisions — never the write path for business semantics.

Canonical storage remains Forma MySQL immutable semantic revisions.

## Consequences

- Sync/rebuild from revision snapshots.
- S2 does not introduce Neo4j for Business Model.

## Rejected Alternatives

- Neo4j as SoT; Coze Knowledge as SoT; writing editor changes directly to KG.
