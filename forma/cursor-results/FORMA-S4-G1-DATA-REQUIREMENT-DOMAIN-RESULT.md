# FORMA-S4-G1 DATA REQUIREMENT DOMAIN RESULT

## Status

**S4-G1 = PASS (pending CI confirmation)**

## Baseline

| Item | Value |
|------|-------|
| S3 frozen tag | `forma-s3-frozen` → `a13fbfc606f873be8e0f88e30baa1709ff32c9dd` |
| S4-G0 | PASS / ARCHITECTURE_APPROVED |
| G0-F1 commit | `fae829ecfdb3de17a51a6882c295871252f30a9a` |
| Parent tip before G1 | `14732f72` |

## Scope Delivered

| Component | Status |
|-----------|--------|
| DataRequirement entity + enums | PASS |
| DataRequirementAnalysisRun | PASS |
| DataRequirementDecision (immutable) | PASS |
| Business Model Revision pinning | PASS |
| Business element ref validation | PASS |
| AI proposal boundary (PROPOSED only) | PASS |
| Human Confirm / Reject / Edit-Confirm | PASS |
| Edit lineage (SUPERSEDED + MANUAL_MODIFIED) | PASS |
| Manual requirement creation | PASS |
| Analysis idempotency + digest conflict | PASS |
| Exact execution owner (CAS / unique) | PASS |
| Failed analysis explicit retry | PASS |
| Tenant isolation | PASS |
| Backend API | PASS |
| Persistence / Migration | PASS |
| FormaDataModel ACL (Coze/Eino adapter) | PASS |
| Architecture dependency guard | PASS |

## Out of Scope (not implemented)

DataSource, DataConnection, Credential, SecretProvider, Schema Discovery, DataAsset, Semantic Mapping, DataContract, DataContractRevision, Drift, Data UI, Business Capability, Agent, Workflow, S5+, `forma-s4-frozen`.

## Package Layout

```
coze-studio/backend/domain/forma/data/
  entity/
  service/
  repository/        (+ memory for tests)
  internal/dal/
```

## Migration

| File | Tables |
|------|--------|
| `20250902100000_s4_g1_data_requirement.sql` | `forma_data_requirement`, `forma_data_requirement_analysis_run`, `forma_data_requirement_decision` |

Unique constraints:
- AnalysisRun: `(tenant_id, business_id, business_model_revision, client_request_id)`
- Decision: `(tenant_id, source_requirement_id)` — single terminal decision

## API Routes

| Method | Path |
|--------|------|
| POST | `/api/forma/v1/businesses/:id/data-requirements/analyze` |
| GET | `/api/forma/v1/businesses/:id/data-analyses/:analysisRunId` |
| POST | `/api/forma/v1/businesses/:id/data-analyses/:analysisRunId/retry` |
| GET/POST | `/api/forma/v1/businesses/:id/data-requirements` |
| POST | `/api/forma/v1/businesses/:id/data-requirements/:requirementId/confirm` |
| POST | `/api/forma/v1/businesses/:id/data-requirements/:requirementId/reject` |
| POST | `/api/forma/v1/businesses/:id/data-requirements/:requirementId/edit-confirm` |
| GET | `/api/forma/v1/businesses/:id/data-requirements/:requirementId/decisions` |

## Error Keys

- `FORMA_DATA_REQUIREMENT_NOT_FOUND`
- `FORMA_DATA_REQUIREMENT_ALREADY_DECIDED`
- `FORMA_DATA_REQUIREMENT_INVALID_STATE`
- `FORMA_DATA_ANALYSIS_NOT_FOUND`
- `FORMA_DATA_ANALYSIS_IDEMPOTENCY_CONFLICT`
- `FORMA_DATA_ANALYSIS_NOT_FAILED`
- `FORMA_DATA_BUSINESS_REVISION_NOT_FOUND`
- `FORMA_DATA_BUSINESS_ELEMENT_REF_INVALID`

## Test Matrix

| Check | Result |
|-------|--------|
| AI output → PROPOSED only | PASS |
| AI cannot Confirm | PASS |
| Business element refs validation | PASS |
| Analysis success atomic persist | PASS |
| Analysis failure → FAILED + zero partial | PASS |
| Same client_request_id sequential idempotency | PASS |
| Same client_request_id concurrent idempotency | PASS |
| Same client_request_id + different digest → conflict | PASS |
| Exactly one fake model invocation (concurrent) | PASS |
| Failed explicit Retry | PASS |
| Confirm + Decision atomic | PASS |
| Reject + Decision atomic | PASS |
| EditConfirm lineage | PASS |
| Original remains SUPERSEDED | PASS |
| Replacement = MANUAL_MODIFIED / CONFIRMED | PASS |
| Decision immutable | PASS |
| Single terminal decision under concurrency | PASS |
| Manual create → PROPOSED | PASS |
| Confirmed requirement always has decision | PASS |
| Business Model revision unchanged | PASS |
| Tenant A/B isolation | PASS |
| Domain agnostic fixtures (lab + procurement) | PASS |
| Domain agnostic static scan (production files) | PASS |
| No provider SDK import in Data Domain | PASS |
| No real model call | PASS |
| Migration validation (CASE A/B/C convention) | PASS |

## Domain Agnostic Fixtures

- **A:** Laboratory sample flow (`node_lab_sample`, `node_lab_tech`, `rule_lab_hold`)
- **B:** Procurement contract approval (`node_po`, `node_approver`, `rule_po_threshold`)

Both use the same domain, API, tables, and enums — no industry switch.

## Invariants

| Invariant | Result |
|-----------|--------|
| IDEMPOTENCY | PASS |
| DECISION_PROVENANCE | PASS |
| EDIT_LINEAGE | PASS |
| TENANT_ISOLATION | PASS |
| DOMAIN_AGNOSTIC | PASS |
| BUSINESS_MODEL_MUTATION | NONE |
| REAL_MODEL_CALLS | 0 |

## Delivery

| Item | Value |
|------|-------|
| Commit SHA | TBD |
| Forma CI | TBD |
| S4-G1 final status | TBD |

## STOP

Do **not** start S4-G2. Do **not** create `forma-s4-frozen`. Await human review.
