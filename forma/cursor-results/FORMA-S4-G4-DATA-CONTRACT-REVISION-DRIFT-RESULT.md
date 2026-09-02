# FORMA-S4-G4 DATA CONTRACT / REVISION / DRIFT RESULT

## Status

**S4-G4 = PASS**

## Baseline

| Item | Value |
|------|-------|
| Parent tip | `37937031533e26b6838b3f6d7cf6042c191168d0` |
| G3-F1 | PASS (`983d2796` / CI `33650518397`) |
| Latest main CI before G4 | `33651137166` PASS |

## Scope Delivered

| Component | Status |
|-----------|--------|
| DataContract identity | PASS |
| DataContractRevision (immutable payload) | PASS |
| ContractLogicalSchema + LogicalField | PASS |
| Contract Binding (materialized from mapping_ids) | PASS |
| Contract Validation + ValidationResult | PASS |
| Lifecycle + LifecycleEvent audit | PASS |
| Schema Drift Evaluation | PASS |
| Breaking Drift → STALE | PASS |
| Business Revision Gap | PASS |
| Source Independence (Logical vs Binding) | PASS |
| Tenant isolation | PASS |
| Authorization (OWNER/ADMIN mutate, MEMBER read) | PASS |
| Backend API | PASS |
| Migration `20250902130000_s4_g4_data_contract.sql` | PASS |
| G3 DSL strict unknown-field closeout | PASS |

## Out of Scope

Data Plane UI, Business Capability, Capability Binding, Agent Composer, Workflow, Application Assembly, Contract Runtime Query Engine, G5+, `forma-s4-frozen`.

## Domain Location

All production code under `domain/forma/data/` (single S4 bounded context).

## Migration

| File | Tables |
|------|--------|
| `20250902130000_s4_g4_data_contract.sql` | `forma_data_contract`, `forma_data_contract_revision`, `forma_data_contract_validation_result`, `forma_data_contract_lifecycle_event`, `forma_data_contract_drift_result`, `forma_data_contract_gap_result` |

No FOREIGN KEY. G1–G3 migrations untouched.

## Invariants

| Invariant | Result |
|-----------|--------|
| DATA_CONTRACT | PASS |
| LOGICAL_SCHEMA | PASS |
| SOURCE_INDEPENDENCE | PASS |
| REVISION_IMMUTABLE | PASS |
| VALIDATION | PASS |
| LIFECYCLE | PASS |
| READ_ONLY | PASS |
| SCHEMA_DRIFT | PASS |
| BREAKING_TO_STALE | PASS |
| NO_AUTO_REPAIR | PASS |
| BUSINESS_REVISION_GAP | PASS |
| TENANT_ISOLATION | PASS |
| AUTHORIZATION | PASS |
| DOMAIN_AGNOSTIC | PASS |
| G1_REGRESSION | PASS |
| G2_REGRESSION | PASS |
| G3_REGRESSION | PASS |
| BUSINESS_MODEL_MUTATION | NONE |
| REAL_MODEL_CALLS | 0 |

## Delivery

| Item | Value |
|------|-------|
| Commit SHA | `4cbfd737e8119560b8106cbd64dfb107ae6f00d0` |
| Forma CI | [33654908875](https://github.com/hai12138/forma/actions/runs/33654908875) **ALL GREEN** |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |

## STOP

Do **not** start S4-G5. Do **not** implement Data Plane UI / Business Capability. Do **not** create `forma-s4-frozen`.

---

# FORMA-S4-G4-F1 DATA CONTRACT CONSISTENCY / GUARANTEE / DRIFT HARDENING

## Status

**S4-G4-F1 = PASS**

## Scope Delivered

| Component | Status |
|-----------|--------|
| Contract row lock (`GetContractForUpdate` FOR UPDATE) | PASS |
| Active pointer CAS clear (`ClearContractActiveRevisionIfMatch`) | PASS |
| Activate serializes on contract lock | PASS |
| Deprecate / BREAKING→STALE clears active pointer | PASS |
| Drift snapshot lineage (source/connection/asset) | PASS |
| `ResolveMappingOutputContractType` + physical type normalize | PASS |
| Validate type / nullability guarantees | PASS |
| Drift TYPE_GUARANTEE_LOST / NULLABILITY_GUARANTEE_LOST | PASS |
| Gap Unmapped = current confirmed without confirmed mapping | PASS |
| Classification policy logical-key only | PASS |
| Consumer `DataContractDescriptor` + active-descriptor API | PASS |
| Duplicate `requirement_ids` denied | PASS |
| Migration | NONE (prefer no migration) |
| REAL_MODEL_CALLS | 0 |

## Invariants (G4-F1)

| Invariant | Result |
|-----------|--------|
| ACTIVE_POINTER_CONSISTENCY | PASS |
| ACTIVATION_SERIALIZATION | PASS |
| DRIFT_LINEAGE | PASS |
| CONTRACT_TYPE_GUARANTEE | PASS |
| NULLABILITY_GUARANTEE | PASS |
| GAP_SEMANTICS | PASS |
| REQUIREMENT_ID_UNIQUENESS | PASS |
| CLASSIFICATION_POLICY | PASS |
| CONSUMER_DESCRIPTOR | PASS |
| SOURCE_INDEPENDENCE | PASS |
| G1_REGRESSION | PASS |
| G2_REGRESSION | PASS |
| G3_REGRESSION | PASS |
| BUSINESS_MODEL_MUTATION | NONE |
| REAL_MODEL_CALLS | 0 |
| NO_MIGRATION | PASS |
| G4_ARCHITECTURE_PRESERVED | PASS |

## Verification

```text
docker run --rm -v "d:/product/Forma/forma-workspace/coze-studio/backend:/src" -w /src golang:1.24 \
  go test ./domain/forma/data/... ./application/forma/ ./api/handler/forma/ -count=1
```

All packages green.

## Delivery

| Item | Value |
|------|-------|
| Commit SHA | `cfb21cb4a9f8a62913ed1c7ac39e4cd7d42bdacc` |
| Forma CI | [33658216479](https://github.com/hai12138/forma/actions/runs/33658216479) **ALL GREEN** |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |

## STOP

Do **not** start S4-G5. Do **not** implement Data Plane UI / Business Capability. Do **not** create `forma-s4-frozen`.

---

# FORMA-S4-G4-F2 STALE REVISION LIFECYCLE / ACTIVE POINTER FINAL CONSISTENCY

## Status

**S4-G4-F2 = PASS**

## Bug Fixed

Deprecate of historical STALE revision no longer calls ClearIfMatch when `active_revision_id` points at a newer ACTIVE revision.

Chosen STALE behavior:
- CASE A empty pointer → STALE→DEPRECATED, no clear
- CASE B pointer → other ACTIVE → STALE→DEPRECATED, leave pointer
- CASE C legacy pointer == this STALE → clear safely then deprecate

ACTIVE→DEPRECATED still requires pointer match and clears via CAS.

## Invariants (G4-F2)

| Invariant | Result |
|-----------|--------|
| STALE_DEPRECATE_WITH_NEW_ACTIVE | PASS |
| ACTIVE_POINTER_PRESERVATION | PASS |
| ACTIVE_DEPRECATE_CLEAR | PASS |
| BREAKING_STALE_CLEAR | PASS |
| ACTIVE_DESCRIPTOR | PASS |
| ACTIVATION_LOCK_CONTRACT | PASS |
| NO_MIGRATION | PASS |
| G1_REGRESSION | PASS |
| G2_REGRESSION | PASS |
| G3_REGRESSION | PASS |
| G4_REGRESSION | PASS |
| BUSINESS_MODEL_MUTATION | NONE |
| REAL_MODEL_CALLS | 0 |

## Delivery

| Item | Value |
|------|-------|
| Commit SHA | `09ac3d37a1bef6b9dca3f6f89f2020f712efbf91` |
| Forma CI | [33662545669](https://github.com/hai12138/forma/actions/runs/33662545669) **ALL GREEN** |
| forma-backend | PASS |
| forma-migration-apply | PASS |
| forma-frontend | PASS |

## STOP

Do **not** start S4-G5. Do **not** implement Data Plane UI / Business Capability. Do **not** create `forma-s4-frozen`.
