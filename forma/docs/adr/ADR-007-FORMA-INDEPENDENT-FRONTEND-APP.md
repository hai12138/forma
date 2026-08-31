# ADR-007: Forma Independent Frontend App

## Status

Accepted (2026-08-31)

## Context

Forma product shell (v1.2 IA) differs fundamentally from Coze Studio shell. Embedding Forma pages inside `frontend/apps/coze-studio` would expose Coze navigation and conflate product identities.

## Decision

Create **`frontend/apps/forma`** as the standalone Forma Product App in the Rush monorepo.

Coze packages remain engineering capability sources (workflow editor, agent debug) consumed via **Forma wrappers/embeds** only.

## Consequences

- Separate build/deploy artifact for Forma UI (`@forma/app`).
- Shared monorepo tooling (Rush/pnpm/rsbuild) but isolated routes and design tokens.
- Two apps may run on different ports in dev (Forma `:3001`, Coze `:8888`).

## Rejected Alternatives

- **Forma routes inside coze-studio app** — Violates DECISION-003 and IA baseline.
- **Standalone repo outside monorepo** — Loses workspace package sharing for `@forma/api-client` and future embeds.
