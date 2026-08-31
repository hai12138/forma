# ADR-010: Server-Side Tenant Context

## Context

Frontend can send `tenant_id` headers or body fields. Relying on them enables cross-tenant attacks.

## Decision

`TenantContext` is always constructed server-side from authenticated session + membership + optional selection header `X-Forma-Tenant`. Asset APIs require this context.

## Consequences

- Forged tenant headers → 403
- UI filtering is never the security boundary
- Middleware `FormaTenantMW` centralizes enforcement for protected routes

## Rejected Alternatives

- Trusting JSON body `tenant_id`
- Client-only workspace switching without server checks
- Repository `GetByID(asset_id)` without tenant predicate
