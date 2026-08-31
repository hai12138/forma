# ADR-003: Forma–Coze Anti-Corruption Layer

## Status

Accepted (2026-08-31)

## Context

Forma Business Domain must not depend on Coze internal repositories (e.g. `domain/agent/singleagent/repository`). Direct coupling would block upstream merges and blur data ownership.

## Decision

Introduce **Forma CrossDomain Integration** (`backend/crossdomain/forma/integration/`) with adapters such as `CozeAgentAdapter` that call **Coze CrossDomain facades** (`crossdomain/agent.DefaultSVC()`), not domain repositories.

Mapping tables `forma_asset_ref` and `forma_coze_resource_ref` link Forma assets to Coze resources by stable ID.

## Consequences

- Forma domain stays testable with mocks.
- Adapter layer must evolve as Coze crossdomain APIs change.
- Automated arch tests enforce import boundaries.

## Rejected Alternatives

- **Direct repository injection into Forma** — Tight coupling, upstream conflict, violates bounded context.
- **Store Forma fields on Coze Agent rows** — Forbidden by data ownership policy (ADR-006).
