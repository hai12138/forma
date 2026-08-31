# ADR-008: Forma Tenant Model

## Context

Forma is a multi-customer Business-to-Agent platform. Coze Space is a technical workspace, not an enterprise customer boundary.

## Decision

Introduce first-class `Forma Tenant` with soft lifecycle (`ACTIVE` / `SUSPENDED` / `ARCHIVED`), membership roles (`OWNER` / `ADMIN` / `MEMBER` / `VIEWER`), and optimistic `revision` concurrency.

## Consequences

- All Forma assets are tenant-scoped.
- Suspended tenants deny business APIs.
- Physical delete of production tenants is forbidden.

## Rejected Alternatives

- Equating Tenant with Coze Space
- Hard-deleting tenants
- Trusting client-supplied tenant identity without membership checks
