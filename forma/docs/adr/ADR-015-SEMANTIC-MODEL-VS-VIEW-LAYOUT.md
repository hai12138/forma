# ADR-015: Semantic Model vs View Layout

## Context

Canvas interaction (drag, zoom, fit, fullscreen) is presentation. Mixing it into semantic revisions floods history and breaks “business meaning” audit.

## Decision

Strictly separate:

| Concern | Storage | Concurrency |
|---------|---------|-------------|
| Semantic Model | `forma_business_model_revision` | `expected_revision` |
| View Layout | `forma_business_model_layout` | `expected_layout_revision` |

Layout includes node positions, viewport, zoom, groups, collapsed, canvas settings. Layout never participates in semantic `content_digest`. AI relayout (future) also must not create semantic revisions.

## Consequences

- UX: drag → layout save; rename/type/edge/rule/state → Save Model → new semantic revision.
- Independent CAS conflicts (`FORMA_BUSINESS_LAYOUT_CONFLICT` vs `FORMA_BUSINESS_MODEL_CONFLICT`).

## Rejected Alternatives

- Single document with positions inside semantic JSON; one shared revision counter for both.
