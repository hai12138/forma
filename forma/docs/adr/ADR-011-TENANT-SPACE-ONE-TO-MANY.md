# ADR-011: Tenant → Space One-to-Many

## Context

Enterprises need multiple Coze Spaces (dev/knowledge/delivery) under one customer boundary.

## Decision

`forma_tenant_space_ref` maps Tenant 1:N Space with extensible `purpose` strings. An active Coze Space binds to at most one Forma Tenant. No DB FK to Coze tables.

## Consequences

- Space sharing across tenants requires a future Shared Resource model (not S1)
- Space access validation goes through CrossDomain ACL (`GetUserSpaceList`)

## Rejected Alternatives

- 1:1 Tenant↔Space
- Embedding Forma fields into Coze space tables
- Direct imports of Coze user DAL from Forma tenancy domain
