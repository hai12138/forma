# ADR-002: Forma v1.2 as Product UI Baseline

## Status

Accepted (2026-08-31)

## Context

Forma requires a distinct enterprise IA: Business Asset, Capability, Agent, Application, Data Plane, Governance, Delivery — not Coze's Space/Develop/Library model.

Forma v1.2 Visual Model Editor prototype defines approved UI, UX, IA, and Design System.

## Decision

**Forma v1.2 remains the sole product UI baseline.** Production shell is `frontend/apps/forma`. Coze Studio UI is not the Forma product shell.

## Consequences

- Separate frontend app and design tokens from Coze Semi Design.
- Coze UI components may appear only inside Forma wrappers/embeds (workflow editor, playground).
- Higher frontend integration effort vs reusing Coze shell wholesale.

## Rejected Alternatives

- **Replace Forma shell with Coze Studio navigation** — Violates product IA and enterprise positioning.
- **iframe entire v1.2** — Poor integration, auth, and long-term maintainability.
