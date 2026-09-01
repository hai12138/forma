# FORMA S3 AI BUSINESS ANALYST RESULT

## Status

PASS_WITH_GATES

## Frozen Baseline

- S2 Frozen Tag: `forma-s2-frozen` → `413c3bcc148dfe518b31d6267e1a0c72fc2f0645`
- Post-Freeze main: `65b165dcd84d1b9f9c083e87d65bbf16f9061ad1`
- S3 work continues on `main` after frozen baseline verification

## Files Changed

### Backend
- `backend/domain/forma/analyst/` — entity, repository, service, DAL
- `backend/application/forma/analyst_app.go`, `analyst_audit.go`, `forma.go`
- `backend/api/handler/forma/analyst.go`, `api/router/forma/api.go`
- `backend/domain/forma/errors/codes.go`
- `backend/crossdomain/forma/integration/analyst_model_adapter.go`
- `backend/domain/forma/meta/version.go` → `0.4.0-s3`
- `docker/atlas/forma/migrations/20250902000000_s3_analyst.sql`

### Frontend
- `frontend/packages/forma-analyst/` — Analyst workspace
- `frontend/packages/forma-api-client/src/index.ts` — analyst API methods
- `frontend/apps/forma` — route wiring

### Docs / CI
- `forma/docs/architecture/AI-BUSINESS-ANALYST-ARCHITECTURE.md`
- `forma/docs/adr/ADR-018` … `ADR-022`
- `.github/workflows/forma-ci.yml`
- `scripts/forma/migration-validate.mjs`

## Analyst Domain

Implemented `backend/domain/forma/analyst/` mirroring business DDD layout. LLM calls via `FormaAnalystModel` ACL only.

## Interview Session

`AnalystSession` with statuses DRAFT/ACTIVE/REVIEWING/COMPLETED/CANCELLED. Multiple sessions per business. `ConfirmationPolicy` extension point (DEVELOPMENT/PRODUCTION).

## Evidence

Immutable `BusinessEvidence` from interview turns (`INTERVIEW_TURN`). Content digest stored.

## Assertions

Structured extraction contract with schema validation. AI assertions default `PROPOSED` + `AI_GENERATED`. Evidence link required for Confirm.

## Confirmation

Immutable `BusinessConfirmation` events. OWNER/ADMIN confirm/reject enforced in application layer.

## Conflicts

Deterministic detection: same `subject_ref` + `predicate` with differing `object_value`. `AssertionConflict` persisted and listed.

## Gaps

`AnalystGap` from extraction and planner drives `NextQuestionPlan`.

## Context Builder

`BuildContext` with budget manifest (included/excluded items, token estimate).

## Model Integration

`CozeEinoAnalystModel` in integration layer; `DeterministicFakeModel` for tests with work-order heuristic extraction.

## Structured Extraction

`ValidateExtractionResult` rejects invalid types/out-of-range confidence. Invalid extraction does not persist canonical assertions.

## Proposal

`SemanticModelPatch` operations (ADD_NODE, ADD_EDGE, ADD_STATE, ADD_RULE, …) with `source_assertion_ids`. Builder uses CONFIRMED assertions only.

## Business Model Apply

`ApplyProposal` → S2 `BusinessService.SaveModel` with CAS `base_revision`. No direct revision INSERT.

## Provenance

`forma_revision_provenance` + semantic model `evidence_refs` / `assertion_refs` on apply.

## Security

System policy boundary in context builder. Confirm/Apply are server-side only (not model tool-call).

## Tenant Isolation

All analyst entities tenant-scoped via repository queries and API tenant middleware.

## Migration

`20250902000000_s3_analyst.sql` — 12 analyst tables. `atlas.sum` updated.

## Frontend

`/analyst` → `AnalystWorkspacePage`: interview thread, assertions/evidence/conflicts/gaps tabs, proposal preview + apply.

## Live Model E2E

**GATE**: Requires CI/runtime with configured builtin chat model. Heuristic fallback ensures deterministic extraction when model unavailable.

## Browser E2E

**GATE**: Playwright browser gates for full Interview→Confirm→Apply flow not executed in this session (no local Go/Playwright run). UI exposes `data-testid` hooks for automation.

## UI Evidence

**GATE**: Screenshots under `forma/cursor-results/s3-ui/` pending browser gate run.

## CI

`forma-ci.yml` extended: analyst domain tests, S3 migration check, `@forma/analyst` build/test.

## Coze Core Files Modified

None (Forma-only paths per S3 scope).

## Remaining Mock

- `DeterministicFakeModel` for unit/integration stability
- Live model path falls back to heuristic when builtin model not configured

## Known Limitations

- No full DOCUMENT_REF file parsing
- No voice/multimodal turns
- PRODUCTION two-person confirmation policy field only (not enforced)
- Proposal diff preview in UI lists operations; full S2 semantic diff component not embedded in analyst panel yet
- Browser E2E + live model E2E gates require CI/runtime execution

## Gate Evidence

| Gate | Status |
|------|--------|
| G01 Session create | Implemented + API |
| G02 Turn → Evidence | Implemented |
| G03 Assertion ↔ Evidence | Implemented |
| G04 No silent revision change | Service design + test |
| G05 Confirm policy | App layer |
| G06 Conflicts | Implemented |
| G07 Gaps | Implemented |
| G08 Invalid extraction blocked | ValidateExtractionResult |
| G09 Confirmed → Proposal | Implemented |
| G10 S2 Validator on apply | ApplyPatch + SaveModel |
| G11 Apply via S2 service | Implemented |
| G12 Revision +1 | SaveModel CAS |
| G13 Provenance chain | revision_provenance table |
| G14 Stale proposal | ErrProposalStale |
| G15 Tenant isolation | tenant_id on all tables |
| G16 Live model E2E | **Pending CI** |
| G17 Browser E2E | **Pending** |
| G18 Work order interview | Heuristic fake + UI |
| G19 S2→S3 migration | SQL + atlas.sum |
| G20 S0/S1/S2 regression | CI retains prior gates |
| G21 Forma CI | **Pending push** |

## Risks

- Atlas checksum must match in strict Atlas CI environments
- Empty `client_request_id` uses turn_id for idempotency uniqueness
- Rush monorepo requires `rush update` after new `@forma/analyst` package

## S4 Preconditions

S3 freeze + human review. Browser/live E2E gates GREEN. Do not start S4 until approved.

---

**DO NOT START S4.** Awaiting human review.
