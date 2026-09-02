# FORMA-S4-G3 SEMANTIC MAPPING RESULT

## Status

**S4-G3 = PASS**  
**FORMA-S4-G3-F1 = PASS**

## Baseline

| Item | Value |
|------|-------|
| Parent tip | `22bb090392604f88990acd28364a9ce8f2e1aab6` |
| S4-G2-F1 | PASS (`356b0c15`) |
| G3 implementation | `7402f98ce43e53c17afee12831c6b4ed64708472` |
| G3 CI | [33644969294](https://github.com/hai12138/forma/actions/runs/33644969294) PASS |
| Pre-F1 main tip | `40fb2bdc0c7b499e31e44a14398f6249d9a0bb93` |

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

G3-F1 migration: **NONE** (no change to G3 migration).

## Controlled DSL

DIRECT, CAST, ENUM_MAP, UNIT_CONVERT, TIME_NORMALIZE, FIELD_PATH, JOIN_REF — validate + serialize only; no SQL/JS/eval.

## Invariants (G3)

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

## G3 Delivery

| Item | Value |
|------|-------|
| Commit SHA | `7402f98ce43e53c17afee12831c6b4ed64708472` |
| Forma CI | [33644969294](https://github.com/hai12138/forma/actions/runs/33644969294) **ALL GREEN** |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |
| S4-G3 final status | **PASS** |

## G3-F1 Contract Hardening

**FORMA-S4-G3-F1 = PASS**

| Invariant | Result |
|-----------|--------|
| BUSINESS_MODEL_CONTEXT | PASS |
| PINNED_SEMANTIC_MODEL | PASS |
| FIELD_PATH_NO_GUESS | PASS |
| DUPLICATE_PATH_REJECTED | PASS |
| DIRECT_CARDINALITY | PASS |
| FIELD_PATH_CONSISTENCY | PASS |
| CAST_DSL | PASS |
| ENUM_MAP_DSL | PASS |
| UNIT_CONVERT_DSL | PASS |
| TIME_NORMALIZE_DSL | PASS |
| JOIN_REF_DSL | PASS |
| CONFIDENCE_VALIDATION | PASS |
| APPLICATION_AUTH | PASS |
| TENANT_ISOLATION | PASS |
| BUSINESS_MODEL_MUTATION | NONE |
| G1_REGRESSION | PASS |
| G2_REGRESSION | PASS |
| REAL_MODEL_CALLS | 0 |

### Hardening notes

- Canonical `PhysicalField.Path` is mandatory and unique; field-name fallback is denied.
- Mapping targets are non-empty, exact, and duplicate-free.
- DIRECT, FIELD_PATH, CAST, ENUM_MAP, UNIT_CONVERT, TIME_NORMALIZE, and JOIN_REF contracts are fully validated.
- AI and manual confidence values are bounded to `[0,1]`; confidence never authorizes confirmation.
- Mapping analysis loads the pinned Business Model revision without mutating the current revision.
- Suggest payload includes `semantic_model`, requirements, schema snapshots, business ID, and revision.
- Credential / Secret / PublicConfig remain excluded from model requests.
- OWNER/ADMIN mutation, MEMBER read-only access, and tenant isolation covered by `mapping_app_test.go`.
- Migration: NONE.
- REAL_MODEL_CALLS: 0.

| Item | Value |
|------|-------|
| Commit SHA | `983d2796ddf9b29ddf3b048f486ce59ea2b2efe1` |
| Forma CI | [33650518397](https://github.com/hai12138/forma/actions/runs/33650518397) **ALL GREEN** |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |

## STOP

Do **not** start S4-G4. Do **not** implement DataContract / Drift / Capability. Do **not** create `forma-s4-frozen`.
