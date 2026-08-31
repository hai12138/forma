# ADR-006: Forma Data Ownership Separation

## Status

Accepted (2026-08-31)

## Context

Forma and Coze must evolve on different lifecycles. Mixing business semantics into Coze core tables causes upstream merge failures and unclear ownership.

## Decision

- **Same MySQL instance**, **logical separation**:
  - Coze tables: unchanged schema ownership.
  - Forma tables: `forma_*` prefix, independent migrations under `docker/atlas/forma/`.
- Link via `forma_asset_ref` + `forma_coze_resource_ref` (no FK to Coze tables).
- **Never** add Forma business columns to Coze Agent/Workflow/etc.

## Consequences

- Joins across boundaries go through mapping layer, not SQL FK.
- Independent Forma migration pipeline.
- Clear SBOM and license separation (Apache Coze vs Forma proprietary).

## Rejected Alternatives

- **Single unified schema in Coze migrations** — High upstream conflict.
- **Separate database instance (V1)** — Valid for later scale; not required for S0-B on same instance.
