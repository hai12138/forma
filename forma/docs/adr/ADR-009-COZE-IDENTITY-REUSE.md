# ADR-009: Coze Identity Reuse

## Context

Building a separate username/password system would duplicate Coze Passport and fragment sessions.

## Decision

V1 reuses Coze User / Session authentication. Forma maps each Coze user to a stable `FormaPrincipal` (`provider=coze`, `external_subject=coze_user_id`).

## Consequences

- No Forma login page credentials flow.
- Unauthenticated users are redirected/blocked to Coze session login.
- Principal resolution is idempotent.

## Rejected Alternatives

- Re-implementing password auth in Forma
- Copying Coze user rows into Forma tables as source of truth
- Using email alone as principal key
