# FORMA S2 BUSINESS ASSET / BUSINESS MODEL / VISUAL EDITOR RESULT

## Status

**PASS_WITH_FINAL_REVIEW_GATE**

S2 + **S2-G1 Canonicality & Browser Closure** + **S2-G2 Editor Validity Closure** complete.
**DO NOT START S3.** Do **not** create `forma-s2-frozen` until human final review.

## Frozen Baseline

| Item | Value |
|------|-------|
| S0 | `forma-s0-frozen` → `d68a49bf1ae780f71d6aecd3ff6d3eb3a1c7a3e6` |
| S1 | `forma-s1-frozen` → `601857c49a167c40c97849f4af543b95cc76fdcb` |
| S1 Tag CI | [33406827644](https://github.com/hai12138/forma/actions/runs/33406827644) ALL GREEN |
| S2 tip (pre-G1) | `30112654dd8c349505c07c930ee18986023c1344` |
| S2 Forma CI (pre-G1) | [33453550169](https://github.com/hai12138/forma/actions/runs/33453550169) ALL GREEN |
| S2-G1 tip | `f6d24f9a907aad8f671fbe373d4cef8ca3ae9d12` |
| S2-G1 Forma CI | [33460385171](https://github.com/hai12138/forma/actions/runs/33460385171) ALL GREEN |
| Coze upstream | `fefb05ff27be1da939612fbf9faf5db62583b8ae` |
| Runtime | Coze / Eino |
| STEP 0 | PASS — `forma-s1-frozen` present; main contains S1 |

## Files Changed

### Backend
- `backend/domain/forma/business/` — entity, validator, digest, diff, service, repository, DAL, fixture (维修工单)
- Canonical NodeType freeze; reject `agent`/`application`/`state`/`rule` as NodeType; alias canonicalize
- SourceMarker contract; edge label required; layout `based_on_model_revision` integrity
- Deterministic Diff/Impact sorting
- `backend/domain/forma/errors/codes.go` — `FORMA_BUSINESS_LAYOUT_MODEL_REVISION_NOT_FOUND`
- `backend/application/forma/business_app.go` + `forma.go` InitService wiring
- `backend/api/handler/forma/business.go` + router `/businesses*`
- `backend/domain/forma/meta/version.go` — `0.3.0-s2` / schema `2.0`
- `docker/atlas/forma/migrations/20250901000000_s2_business_model.sql` + `atlas.sum`
- `scripts/forma/live-business-e2e.mjs` (tenant isolation expansion)
- `scripts/forma/s2-g1-browser-gates.mjs` (**MANUAL_LIVE_BROWSER_GATE**)
- `scripts/forma/s2-g2-browser-validity-gates.mjs` (**MANUAL_LIVE_BROWSER_VALIDITY_GATE**)

### Frontend
- `frontend/packages/forma-business/` — Visual Model Editor
  - Auto/Manual layout (`auto-layout.ts`)
  - Canonical adapter (`canonical.ts`)
  - Dependency-aware delete; Rule not edge endpoint
  - Edge type dropdown + required label
  - Double-click name edit; historical revision read-only view
  - S2-G2: State/Rule selectors, required-name guard, `semantic-validator.ts` preflight
- `frontend/packages/forma-api-client` — Business API types/methods
- `frontend/apps/forma` — routes `/business`, `/business/:businessId`

### Docs
- `forma/docs/architecture/BUSINESS-MODEL-ARCHITECTURE.md`
- ADR-013 … ADR-017
- This result (+ `forma/cursor-results/` copy)

### CI
- `.github/workflows/forma-ci.yml` — business domain tests, S2 migration file, `@forma/business` build/test

## Business Domain

DDD package `domain/forma/business` with `entity` / `service` / `repository` / `internal/dal`. No Capability/Agent/Workflow domain deps.

## Business Asset

- Reuses `forma_asset_ref` with `kind=BUSINESS`
- `business_id == asset_id` (ADR-013)
- Create uses DB transaction: Asset + Model master + revision 1 + layout 1
- Lifecycle remains Asset Registry (DRAFT/ARCHIVED); no duplicate lifecycle on model

## Semantic Model

`schema_version`, `nodes[]`, `edges[]`, `rules[]`, `states[]`, reserved `evidence_refs` / `assertion_refs`.

**Canonical NodeTypes only:** `ACTOR` `BUSINESS_OBJECT` `PROCESS` `EVENT` `DECISION` `SYSTEM` `POLICY`.

State/Rule remain first-class `states[]`/`rules[]` (not NodeTypes).

## Revision Model

Immutable `forma_business_model_revision`; CAS via `expected_revision`; identical digest → `no_change` (no spurious revision).

## Layout Model

`forma_business_model_layout` with independent `layout_revision` / `expected_layout_revision`. Drag/zoom/fit/fullscreen/Auto Layout do not create semantic revisions.

`based_on_model_revision` must exist for same tenant+business (historical allowed; nonexistent/other-business rejected).

## Validator

`ValidateSemanticModel` — unique IDs, dangling endpoints (nodes or states), required names, canonical types, SourceMarker enum, non-empty edge labels, reject agent/application/state/rule NodeTypes; canonicalize role/entity/process/external.

## Digest

Canonical serialize (sorted IDs/keys) + SHA-256. Layout excluded. Empty slices normalized so API never returns null arrays that break UI.

## Diff Engine

Element-level Nodes/Edges/Rules/States Added|Removed|Modified — **sorted** before return. Impact IDs sorted + deduped. Deterministic across repeated Diff.

## API

`/api/forma/v1/businesses` CRUD/archive + `/model` + `/revisions` + `/diff` + `/layout`. TenantContext required; body `tenant_id` not trusted.

## Tenant Isolation

Lookups always scoped by TenantContext tenant. Cross-tenant Get/Model/Revision/Layout/Diff/Archive/PATCH/PUT Model/PUT Layout → NOT_FOUND / DENIED. Live E2E expanded.

## Migration

S2 migration only (S0/S1/S1-G1 untouched). CASE A/B/C `migration-apply-test.mjs` PASS. Live MySQL upgraded S1-G1 → S2 PASS.

## Visual Editor

`@forma/business` VisualModelEditor: dotted white canvas, colored nodes, edges, left tools, top Save Model/Layout + Auto Layout/Manual + Undo/Redo + Zoom/Fit/Fullscreen + Revisions/Diff, right property panel + source marker, legend. Semantic dirty vs layout dirty separated. Historical revision read-only view with Back to Current.

## v1.2 Migration

Golden Reference: `forma-reference/v1.2/Forma-Business-to-Agent-Platform-v1.2-VisualModelEditor/` (read-only). Production uses real API — not mock store. `layoutGraph` ported for deterministic Auto Layout.

## Reference Business

维修工单 seed: backend `fixture.WorkOrderRepairSemanticModel` + frontend `workOrderSeed()`. Fixture-only — not platform core enums.

## Browser E2E

- Live API: `FORMA_LIVE_E2E=1 node --test scripts/forma/live-business-e2e.mjs` — **PASS** (create, layout isolation, semantic +1, diff, tenant isolation expansion, layout revision integrity)
- Real browser: `FORMA_LIVE_E2E=1 node --test scripts/forma/s2-g1-browser-gates.mjs` — **MANUAL_LIVE_BROWSER_GATE PASS** (assertions on revision numbers; not screenshot-only)

## UI Evidence (S2-G1)

| File | Content |
|------|---------|
| `s2-g1-ui/01-editor-current.png` | Editor current |
| `s2-g1-ui/02-node-dragged-layout-dirty.png` | Real drag → layout dirty |
| `s2-g1-ui/03-layout-saved-semantic-revision-unchanged.png` | Save Layout; semantic rev unchanged |
| `s2-g1-ui/04-semantic-edited.png` | Double-click rename |
| `s2-g1-ui/05-semantic-saved-revision-incremented.png` | Save Model +1 |
| `s2-g1-ui/06-auto-layout.png` | Auto Layout |
| `s2-g1-ui/07-history-readonly.png` | Historical revision read-only |
| `s2-g1-ui/08-diff.png` | Diff panel |
| `s2-g1-ui/09-delete-dependency.png` | Dependency-aware delete |

## S2-G1 Canonicality & Browser Closure

| Item | Result |
|------|--------|
| Canonical Node Model | PASS — only 7 formal types persisted |
| Source Marker Contract | PASS — enum only; empty→MANUAL_MODIFIED |
| Editor/Backend Contract | PASS — Rule no handle; edge dropdown; label required |
| Dependency Delete | PASS — cascade states; strip rule applies_to |
| Auto/Manual Layout | PASS — deterministic; layout dirty only |
| Layout Referential Integrity | PASS — `FORMA_BUSINESS_LAYOUT_MODEL_REVISION_NOT_FOUND` |
| Deterministic Diff | PASS — sorted Added/Removed/Modified + Impact |
| Historical Read-only View | PASS — GET revision; Back to Current |
| Real Browser Drag Gate | PASS — Playwright mouse drag |
| Real Browser Semantic Gate | PASS — dblclick → Save Model +1 |
| Real Browser Auto Layout Gate | PASS |
| Tenant Isolation Expansion | PASS — all business handlers |
| CI | S0/S1/S2 jobs retained; G1 unit tests in domain + `@forma/business` |

## S2-G2 Editor Validity Closure

| Item | Result |
|------|--------|
| State creation validity | PASS — empty model disables ＋状态 |
| State object_ref selector | PASS — select from model.nodes only |
| Rule applies_to selector | PASS — checkbox Node/State candidates |
| Required name validation | PASS — blank Node/State/Rule names blocked in UI |
| Frontend preflight | PASS — `validateEditorSemanticModel` before Save Model API |
| Semantic ID uniqueness | PASS — Nodes/States/Rules/Edges global unique (BE+FE) |
| Browser validity gate | PASS — `s2-g2-browser-validity-gates.mjs` CASE A–E |
| CI | Domain + `@forma/business` tests extended; await G2 tip CI |

## CI

S2 pre-G1 tip CI: [33453550169](https://github.com/hai12138/forma/actions/runs/33453550169)

S2-G1 Forma CI: [33460385171](https://github.com/hai12138/forma/actions/runs/33460385171)

- forma-backend = PASS
- forma-migration-apply = PASS
- forma-frontend = PASS

→ **GATE-18 = PASS** (S2-G1 confirmed ALL GREEN)

S2-G2 tip Forma CI will re-run on push; await ALL GREEN + human final review before `forma-s2-frozen`.

Local verification (S2-G2):
- `go test ./domain/forma/business/...` PASS
- `@forma/business` vitest 18/18 PASS
- Live business E2E PASS (no regression)
- S2-G1 MANUAL_LIVE_BROWSER_GATE PASS (6/6)
- S2-G2 MANUAL_LIVE_BROWSER_VALIDITY_GATE PASS (6/6)

## Coze Core Files Modified

**None** (Forma packages / atlas / Forma CI only). Harness binary rebuild mounts Forma code; no Coze Studio shell swap.

## Remaining Mock

- Placeholder pages for Analyst / Data / Capability / Agent / etc. (intentional non-S2)
- Overview KPI cards still show zero Capability/Agent/Application (real counts)

## Known Limitations

- Full interactive browser harness remains **MANUAL_LIVE_BROWSER_GATE** (not GitHub Hosted Runner) — must be run with live harness + FE; assertions are automated
- AI Analyst / Evidence extraction not implemented (skeleton refs only)
- Knowledge Graph projection not implemented (ADR-016)

## Gate Evidence

| Gate | Result | Evidence |
|------|--------|----------|
| GATE-01 BUSINESS Asset in registry | PASS | CreateBusiness txn + list; live E2E |
| GATE-02 Immutable semantic revision | PASS | domain service + DB table; no UPDATE path |
| GATE-03 Semantic edit → new revision | PASS | live E2E + service_test + browser gate |
| GATE-04 Layout ≠ semantic revision | PASS | live E2E + browser drag/save layout |
| GATE-05 Deterministic digest | PASS | digest tests |
| GATE-06 Invalid graph rejected | PASS | validator tests (canonical + source_marker + label) |
| GATE-07 Version conflict | PASS | service_test + error key |
| GATE-08 Layout conflict independent | PASS | service_test |
| GATE-09 Semantic Diff | PASS | diff sorted tests + live E2E |
| GATE-10 Tenant isolation | PASS | expanded live E2E |
| GATE-11 维修工单 reference | PASS | fixture + seed + UI |
| GATE-12 v1.2 editor in Forma app | PASS | `@forma/business` + Auto Layout |
| GATE-13 Browser drag ≠ semantic | PASS | **real Playwright drag** + revision asserts |
| GATE-14 Browser Save Model +1 | PASS | **real Playwright** semantic save |
| GATE-15 Revision + Diff | PASS | history read-only + Diff panel |
| GATE-16 S1→S2 migration | PASS | atlas apply live + CASE A/B/C |
| GATE-17 S0/S1 CI preserved | PASS | workflow keeps prior jobs/paths |
| GATE-18 S2 Forma CI ALL GREEN | **PASS** | [33453550169](https://github.com/hai12138/forma/actions/runs/33453550169) three jobs ALL GREEN |

## Risks

- Empty JSON slices historically marshaled as `null` — mitigated by normalize on read + FE guards
- Live harness is manually rebuilt binary mount — rebuild before browser gates
- Editor undo buffers client-only (by design)

## S3 Preconditions

1. Human review of this S2 + S2-G1 result
2. S2-G1 tip Forma CI ALL GREEN
3. Only then create `forma-s2-frozen` and verify Tag CI
4. Capability / Data Contract stages remain blocked until Business Model SoT accepted

---

**Status = PASS_WITH_FINAL_REVIEW_GATE**

**DO NOT START S3.**

**DO NOT create `forma-s2-frozen` this round.**

等待人工最终源码审核。
