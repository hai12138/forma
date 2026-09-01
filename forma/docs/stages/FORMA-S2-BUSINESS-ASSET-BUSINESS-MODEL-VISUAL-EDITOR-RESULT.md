# FORMA S2 BUSINESS ASSET / BUSINESS MODEL / VISUAL EDITOR RESULT

## Status

**PASS_WITH_GATES**

Local domain/API/migration/live E2E/UI evidence PASS. Forma CI on GitHub not yet re-run on this tip (await push + review). **DO NOT START S3.**

## Frozen Baseline

| Item | Value |
|------|-------|
| S0 | `forma-s0-frozen` → `d68a49bf1ae780f71d6aecd3ff6d3eb3a1c7a3e6` |
| S1 | `forma-s1-frozen` → `601857c49a167c40c97849f4af543b95cc76fdcb` |
| S1 Tag CI | [33406827644](https://github.com/hai12138/forma/actions/runs/33406827644) ALL GREEN |
| Coze upstream | `fefb05ff27be1da939612fbf9faf5db62583b8ae` |
| Runtime | Coze / Eino |
| STEP 0 | PASS — `forma-s1-frozen` present; main contains S1 |

## Files Changed

### Backend
- `backend/domain/forma/business/` — entity, validator, digest, diff, service, repository, DAL, fixture (维修工单)
- `backend/application/forma/business_app.go` + `forma.go` InitService wiring
- `backend/api/handler/forma/business.go` + router `/businesses*`
- `backend/domain/forma/errors/codes.go` — `FORMA_BUSINESS_*`
- `backend/domain/forma/meta/version.go` — `0.3.0-s2` / schema `2.0`
- `docker/atlas/forma/migrations/20250901000000_s2_business_model.sql` + `atlas.sum`
- `scripts/forma/live-business-e2e.mjs`, `s2-ui-screenshots.mjs`, migration-validate S2 asserts

### Frontend
- `frontend/packages/forma-business/` — list + Visual Model Editor (v1.2 IA)
- `frontend/packages/forma-api-client` — Business API types/methods
- `frontend/apps/forma` — routes `/business`, `/business/:businessId`
- `rush.json` — `@forma/business`

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

Formal node types + v1.2 aliases; extensible edge types; `source_marker` AI_GENERATED | MANUAL_MODIFIED.

## Revision Model

Immutable `forma_business_model_revision`; CAS via `expected_revision`; identical digest → `no_change` (no spurious revision).

## Layout Model

`forma_business_model_layout` with independent `layout_revision` / `expected_layout_revision`. Drag/zoom/fit/fullscreen do not create semantic revisions.

## Validator

`ValidateSemanticModel` — unique IDs, dangling endpoints (nodes or states), required names, unsupported types, self-edge policy.

## Digest

Canonical serialize (sorted IDs/keys) + SHA-256. Layout excluded. Empty slices normalized so API never returns null arrays that break UI.

## Diff Engine

Element-level Nodes/Edges/Rules/States Added|Removed|Modified + `BusinessImpactSummary` (no Capability/Agent impact pretence).

## API

`/api/forma/v1/businesses` CRUD/archive + `/model` + `/revisions` + `/diff` + `/layout`. TenantContext required; body `tenant_id` not trusted.

## Tenant Isolation

Lookups always scoped by TenantContext tenant. Cross-tenant Get/Model/Layout/Diff/Archive → NOT_FOUND / DENIED. Live E2E covered.

## Migration

S2 migration only (S0/S1/S1-G1 untouched). CASE A/B/C `migration-apply-test.mjs` PASS. Live MySQL upgraded S1-G1 → S2 PASS.

## Visual Editor

`@forma/business` VisualModelEditor: dotted white canvas, colored nodes, edges, left tools, top Save Model/Layout + Undo/Redo + Zoom/Fit/Fullscreen + Revisions/Diff, right property panel + source marker, legend. Semantic dirty vs layout dirty separated.

## v1.2 Migration

Golden Reference: `forma-reference/v1.2/Forma-Business-to-Agent-Platform-v1.2-VisualModelEditor/` (read-only). Production uses real API — not mock store.

## Reference Business

维修工单 seed: backend `fixture.WorkOrderRepairSemanticModel` + frontend `workOrderSeed()`. Fixture-only — not platform core enums.

## Browser E2E

- Live API: `FORMA_LIVE_E2E=1 node --test scripts/forma/live-business-e2e.mjs` — **8/8 PASS** (create, layout isolation, semantic +1, diff, tenant isolation)
- UI Playwright screenshots under `forma/cursor-results/s2-ui/`

## UI Evidence

| File | Content |
|------|---------|
| `s2-ui/01-business-list-empty.png` | Empty state (unauth) |
| `s2-ui/01-business-list.png` | List with 维修工单 r1 |
| `s2-ui/02-visual-editor.png` | Visual canvas + toolbar |
| `s2-ui/03-node-selected.png` | Node / property panel |
| `s2-ui/04-revision-history.png` | Revisions panel |
| `s2-ui/05-diff-view.png` | Diff panel |
| `s2-ui/06-fit-view.png` | Fit view |

## CI

Extended Forma CI locally configured. Remote ALL GREEN pending push of this tip.

Local verification:
- `go test ./domain/forma/business/...` PASS
- `go test ./application/forma/...` PASS
- `go test ./api/handler/forma/...` PASS
- `node scripts/forma/migration-validate.mjs` PASS
- `node scripts/forma/migration-apply-test.mjs` PASS (A/B/C)
- `@forma/business` vitest PASS (per FE agent)
- Live business E2E PASS

## Coze Core Files Modified

**None** (Forma packages / atlas / Forma CI only). Harness binary rebuild mounts Forma code; no Coze Studio shell swap.

## Remaining Mock

- Placeholder pages for Analyst / Data / Capability / Agent / etc. (intentional non-S2)
- Overview KPI cards still show zero Capability/Agent/Application (real counts)

## Known Limitations

- Full interactive browser drag→layout / Save Model revision gates evidenced primarily via **live API E2E**; UI screenshots show editor + panels (Playwright)
- AI Analyst / Evidence extraction not implemented (skeleton refs only)
- Knowledge Graph projection not implemented (ADR-016)
- Remote Forma CI not yet executed on this commit tip

## Gate Evidence

| Gate | Result | Evidence |
|------|--------|----------|
| GATE-01 BUSINESS Asset in registry | PASS | CreateBusiness txn + list; live E2E |
| GATE-02 Immutable semantic revision | PASS | domain service + DB table; no UPDATE path |
| GATE-03 Semantic edit → new revision | PASS | live E2E + service_test |
| GATE-04 Layout ≠ semantic revision | PASS | live E2e layout save; service_test |
| GATE-05 Deterministic digest | PASS | digest tests |
| GATE-06 Invalid graph rejected | PASS | validator tests |
| GATE-07 Version conflict | PASS | service_test + error key |
| GATE-08 Layout conflict independent | PASS | service_test |
| GATE-09 Semantic Diff | PASS | diff tests + live E2E |
| GATE-10 Tenant isolation | PASS | live E2E tenant B |
| GATE-11 维修工单 reference | PASS | fixture + seed + UI list/editor |
| GATE-12 v1.2 editor in Forma app | PASS | `@forma/business` + screenshots |
| GATE-13 Browser drag ≠ semantic | PASS | live API layout path (UI drag→layout save wired) |
| GATE-14 Browser Save Model +1 | PASS | live E2E semantic save |
| GATE-15 Revision + Diff | PASS | live E2E + UI panels |
| GATE-16 S1→S2 migration | PASS | atlas apply live + CASE A/B/C |
| GATE-17 S0/S1 CI preserved | PASS | workflow keeps prior jobs/paths |
| GATE-18 S2 Forma CI ALL GREEN | **PENDING** | needs push + Actions run |

## Risks

- Empty JSON slices historically marshaled as `null` — mitigated by normalize on read + FE guards
- Live harness is manually rebuilt binary mount — document rebuild in ops notes
- Editor undo buffers client-only (by design)

## S3 Preconditions

1. Human review of this S2 result + ADRs
2. Push tip; Forma CI ALL GREEN; optional freeze tag
3. Capability / Data Contract stages remain blocked until Business Model SoT accepted

---

**DO NOT START S3.**

等待人工审核。
