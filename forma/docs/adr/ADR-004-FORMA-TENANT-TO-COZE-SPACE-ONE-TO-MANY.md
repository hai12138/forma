# ADR-004: Forma Tenant to Coze Space (1:N)

## Status

Accepted (2026-08-31)

## Context

Coze open source uses **Space** as the primary isolation unit. Forma targets enterprise customers with **Tenant** as the commercial and governance boundary.

## Decision

**Forma Tenant : Coze Space = 1 : N**

- Forma Tenant = enterprise customer boundary (contracts, billing, governance).
- Coze Space = technical workspace inside a tenant (teams, dev/prod separation).

Forma `tenant_id` is carried on all Forma assets; Coze `space_id` appears on Coze resource refs only.

## Consequences

- Forma must implement tenant registry and mapping (future S1+).
- Cannot assume `space_id == tenant_id`.
- Multi-space customers supported without duplicating business semantics.

## Rejected Alternatives

- **Space equals Tenant** — Too coarse for enterprise multi-workspace patterns.
- **Ignore Coze Space** — Breaks Coze Agent/Workflow scoping.
