# AI Business Analyst Architecture (FORMA S3)

## Overview

Forma S3 establishes the **AI Business Analyst** pipeline. AI never mutates the Business Model Source of Truth directly.

```
Business Interview (AnalystSession / AnalystTurn)
        ↓
Evidence (immutable BusinessEvidence)
        ↓
Assertion (BusinessAssertion, PROPOSED by default)
        ↓
Conflict / Gap detection
        ↓
Confirmation (BusinessConfirmation — immutable events)
        ↓
Proposed Business Model Patch (SemanticModelPatch)
        ↓
Human Apply
        ↓
New Business Model Revision (via S2 BusinessService)
```

## Domain boundaries

| Layer | Responsibility |
|-------|----------------|
| `analyst/entity` | Session, Turn, Evidence, Assertion, Confirmation, Conflict, Gap, Proposal |
| `analyst/service` | Turn flow, extraction validation, conflict/gap, proposal builder, apply orchestration |
| `business/service` | Semantic model validation, revision CAS, digest |
| `crossdomain/forma/integration` | `FormaAnalystModel` ACL → Coze/Eino builtin chat model |

## Key principles

1. **Business Model remains SoT** — only `ApplyProposal` calls `BusinessService.SaveModel`.
2. **Evidence is immutable** — new source → new Evidence row.
3. **Assertions default to PROPOSED** — confidence is metadata, not authorization.
4. **Confirmation is an immutable event** — status on Assertion is a projection.
5. **Proposals reference assertions** — each patch operation carries `source_assertion_ids`.
6. **Stale proposals** — `base_revision` mismatch → `STALE`, cannot apply.
7. **Tenant isolation** — all analyst tables scoped by `tenant_id`.

## Context budgeting

`AnalystContextBuilder` prioritizes: current model snapshot → open conflicts/gaps → recent turns → confirmed assertions → evidence excerpts. Manifest records included/excluded items for debugging.

## Provenance

`forma_revision_provenance` links `revision_no` → `proposal_id` → `assertion_ids`. Evidence links via `forma_assertion_evidence_ref`. Turns via Evidence `turn_id`.

## Authorization

- **OWNER / ADMIN**: Confirm, Reject, Create/Apply Proposal
- **MEMBER**: Interview, create manual notes (future), propose assertions
- **VIEWER**: Read-only

`ConfirmationPolicy` field on session: `DEVELOPMENT` (single approver) vs `PRODUCTION` (future two-person rule).

## LLM boundary

Domain depends on `FormaAnalystModel` interface only. Production wiring uses `CozeEinoAnalystModel` in integration layer. Tests use `DeterministicFakeModel`.
