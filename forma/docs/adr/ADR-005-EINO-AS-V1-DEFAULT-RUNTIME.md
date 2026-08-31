# ADR-005: Eino as V1 Default Runtime

## Status

Accepted (2026-08-31)

## Context

Coze Studio backend uses CloudWeGo **Eino** for Agent ReAct flows and Workflow compose execution. Forma needs a default runtime for V1 without multiplying execution stacks.

## Decision

**V1 default runtime = Coze / Eino Runtime.**

Do not introduce LangGraph as default. Do not implement DeepSeek Harness in V1. DeepSeek **Model API** may be used via Coze Model Manager. Reserve **Runtime Adapter** interface for future runtimes.

## Consequences

- Forma focuses on business layer, not execution engine rewrite.
- Runtime compatibility matrix documented at release time.
- Future alternate runtimes require adapter certification.

## Rejected Alternatives

- **LangGraph default** — Not present in Coze codebase; adds stack complexity.
- **DeepSeek Harness as V1** — No concrete harness spec in scope; conflates model API with execution framework.
