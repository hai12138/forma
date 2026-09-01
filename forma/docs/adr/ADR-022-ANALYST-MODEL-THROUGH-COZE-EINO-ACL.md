# ADR-022: Analyst Model Through Coze Eino ACL

## Status

Accepted (FORMA S3)

## Decision

`FormaAnalystModel` interface lives in analyst domain. Implementation `CozeEinoAnalystModel` in `crossdomain/forma/integration` uses Coze Model Manager builtin chat model. Domain packages do not import provider SDKs.

## Consequences

- Unit tests use `DeterministicFakeModel`.
- Live E2E uses configured builtin model with heuristic fallback.
