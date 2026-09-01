# ADR-017: Business Model Opaque String IDs

## Context

ADR-012 established public Coze IDs as strings. Business Model elements (business, asset, node, edge, rule, state) also cross API and UI boundaries.

## Decision

All Business Model public IDs are **opaque strings**:

- `tenant_id`, `business_id`, `asset_id`
- `node_id`, `edge_id`, `rule_id`, `state_id`

No platform-core enums for domain-specific labels (e.g. 维修工单 states). Reference businesses are fixtures/seeds only.

## Consequences

- JSON APIs never expose numeric snowflakes for these IDs.
- Validators treat IDs as opaque uniqueness keys.

## Rejected Alternatives

- Auto-increment public IDs; hard-coded domain enums in platform core; JS number IDs.
