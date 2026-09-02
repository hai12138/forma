# FORMA-S4-G3 SEMANTIC MAPPING RESULT

## Status

**S4-G3 = PASS (pending CI confirmation)**

## Baseline

| Item | Value |
|------|-------|
| Parent tip | `22bb090392604f88990acd28364a9ce8f2e1aab6` |
| S4-G2-F1 | PASS (`356b0c15`) |

## Scope Delivered

| Component | Status |
|-----------|--------|
| SemanticMapping entity | PASS |
| SemanticMappingAnalysisRun | PASS |
| SemanticMappingDecision (immutable) | PASS |
| Controlled Mapping DSL | PASS |
| AI SuggestSemanticMappings (Fake in tests) | PASS |
| Human Confirm / Reject / Edit-Confirm | PASS |
| Snapshot pinning + field path validation | PASS |
| Analysis idempotency | PASS |
| Mapping coverage query | PASS |
| Tenant isolation | PASS |
| Backend API | PASS |
| Migration | PASS |

## Out of Scope

DataContract, Drift, Capability, Agent, Workflow, Data Plane UI, `forma-s4-frozen`.

## Domain Location

All production code under `domain/forma/data/` (single S4 bounded context).

## Migration

| File | Tables |
|------|--------|
| `20250902120000_s4_g3_semantic_mapping.sql` | `forma_data_semantic_mapping`, `forma_data_semantic_mapping_analysis_run`, `forma_data_semantic_mapping_decision` |

## Controlled DSL

DIRECT, CAST, ENUM_MAP, UNIT_CONVERT, TIME_NORMALIZE, FIELD_PATH, JOIN_REF — validate + serialize only; no SQL/JS/eval.

## Invariants

| Invariant | Result |
|-----------|--------|
| AI_PROPOSAL_ONLY | PASS |
| HUMAN_CONFIRMATION | PASS |
| CONTROLLED_DSL | PASS |
| SNAPSHOT_PINNING | PASS |
| MAPPING_PROVENANCE | PASS |
| IDEMPOTENCY | PASS |
| TENANT_ISOLATION | PASS |
| DOMAIN_AGNOSTIC | PASS |
| BUSINESS_MODEL_MUTATION | NONE |
| G1_REGRESSION | PASS |
| G2_REGRESSION | PASS |
| REAL_MODEL_CALLS | 0 |

## Delivery

| Item | Value |
|------|-------|
| Commit SHA | TBD |
| Forma CI | TBD |
| S4-G3 final status | TBD |

## STOP

Do **not** start S4-G4. Do **not** implement DataContract / Drift / Capability. Do **not** create `forma-s4-frozen`.
