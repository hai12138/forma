# S5-G4-F7 — Capability UI Read Seam INTERFACE FREEZE

**Baseline:** `5dfc715105a927243311ef8f04aa0a5bfb28c730`
**Scouts:** Wire Contract + Auth/Security (read-only)
**Status:** FROZEN — candidates only; agents MUST NOT push main or declare PASS.

---

## Resolved conflicts (do not re-guess)

1. **Batch AssetRef:** Extend `AssetProjection` with `ListCapabilityAssetsByTenant(ctx, tenantID) ([]*assetentity.AssetRef, error)` (CAPABILITY kind only). Wire via `CapabilityUnitOfWork.Assets()` non-tx root view (`gorm`: `NewGormAssetProjection(db)`; memory: owned `MemoryAssetProjection`). ListCapabilities uses **exactly one** tenant batch call — never per-row `GetCapabilityAsset`.
2. **CreateCapabilityResponse projection:** After successful `ManualCreate`, fill summary via **`ProjectCapabilityAssetRef(cap, []*Revision{rev})`** using the returned revision — deterministic, no post-commit AssetRef read that could fail after write.
3. **Get/List projection:** Read persisted AssetRef headers (`Name`, `SemanticVersion`, `Status` → `asset_status`, `ContentDigest`). Fail closed on missing / wrong `Kind` / `AssetID != CapabilityID` / tenant mismatch / duplicate asset_id in batch map → `ErrConsistency`/`ErrConflict` via existing MapDomainError (409). Never invent name from capability_id.
4. **No new Migration / frontend / Runtime / G5 / real model.**

---

## Frozen API

### CapabilityDTO (extend; keep existing fields)

```go
Name            string `json:"name"`
SemanticVersion string `json:"semantic_version"`
AssetStatus     string `json:"asset_status"`
ContentDigest   string `json:"content_digest"`
```

- `capability_id == asset_id`; Kind must be `CAPABILITY`.
- AssetRef is projection only — Revision semantic payload remains SoT.

### GET Proposal

```text
GET /api/forma/v1/businesses/:id/capability-proposals/:proposalId
→ Forma envelope<CapabilityProposalDTO>
```

App: `GetCapabilityProposal(ctx, businessID, proposalID) (*CapabilityProposalDTO, error)`
- `requireCapabilityRead` + existing `requireCapabilityProposal`
- ACTIVE OWNER/ADMIN/MEMBER/VIEWER readable
- SUPER_ADMIN without ACTIVE membership → TenantForbidden
- All proposal statuses readable; DTO logical-only / secret-free

### GET Analysis proposals

```text
GET /api/forma/v1/businesses/:id/capability-analyses/:analysisRunId/proposals
→ Forma envelope<CapabilityProposalDTO[]>
```

Domain façade (add to `CapabilityService`):
```go
ListProposalsByAnalysisRun(ctx, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error)
```

App must: require read → `requireCapabilityAnalysis` → list → defensive check each proposal tenant/business/analysis_run_id → stable order → no mutations. Existing GET Analysis response unchanged.

---

## Agent ownership (Wave 1)

### Agent A — Tests ONLY (exact new files)

```text
coze-studio/backend/domain/forma/capability/service/ui_read_seam_f7_test.go
coze-studio/backend/application/forma/capability_ui_read_seam_f7_test.go
coze-studio/backend/api/handler/forma/capability_ui_read_seam_f7_test.go
```

Branch: `forma/s5-g4-f7-agent-a-tests`

### Agent B — Production ONLY (exact allowlist)

```text
coze-studio/backend/domain/forma/capability/service/service.go
coze-studio/backend/domain/forma/capability/service/uow.go
coze-studio/backend/domain/forma/capability/service/asset_memory.go
coze-studio/backend/domain/forma/capability/service/asset_gorm.go
coze-studio/backend/domain/forma/capability/service/asset_failing_test.go
coze-studio/backend/application/forma/capability_dto.go
coze-studio/backend/application/forma/capability_app.go
coze-studio/backend/application/forma/capability_analysis_app.go
coze-studio/backend/api/handler/forma/capability.go
coze-studio/backend/api/handler/forma/capability_analysis.go
coze-studio/backend/api/router/forma/api.go
```

Branch: `forma/s5-g4-f7-agent-b-implementation`

If additional production files are required → STOP; main integrator revises this freeze.

---

## Main integrator

TDD: cherry-pick A (RED) → B (GREEN) → full gates → dual reviewers → RESULT → CI.
`S5_G5_READY = NO`. No `forma-s5-frozen`. No G5.
