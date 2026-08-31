# ADR-001: Coze Studio as Agent Engineering Foundation

## Status

Accepted (2026-08-31)

## Context

Forma V1 must ship enterprise Business-to-Agent capabilities without rebuilding mature Agent engineering primitives (Agent CRUD, Workflow, Plugin, Knowledge, Model, Runtime).

Coze Studio open source provides a DDD Go backend, Eino-based runtime, and a large frontend monorepo for agent IDE and workflow editing.

## Decision

Forma V1 is developed as a **fork/extension** of Coze Studio. Coze remains the **Agent Engineering Foundation**; Forma is the **Business-to-Agent Product Layer** on top.

## Consequences

- Faster time-to-market for Agent execution, workflow, and plugin tooling.
- Upstream merge discipline required (`upstream` = coze-dev/coze-studio).
- Forma-specific domains live in isolated `backend/domain/forma/*` packages.

## Rejected Alternatives

- **Greenfield Agent platform** — Duplicates Eino runtime, workflow engine, and IDE; high cost and risk.
- **Coze as external API only** — Loses deep integration and self-hosted deployment cohesion.
