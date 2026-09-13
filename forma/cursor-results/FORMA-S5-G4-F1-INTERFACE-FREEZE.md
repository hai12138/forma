# S5-G4-F1 — Frozen Public Interface & File Ownership

**Baseline main:** `dc1a90b887d5d1dac74e77f8b1aa36677ac6e203`  
**G4 impl:** `5c3d257ec88e0d0f6dd8467824cb511322449bd7`  
**Status:** FROZEN for parallel agents — do not declare PASS; submit candidate commits only.

---

## Identity contract (LOCKED)

| Field | Source | Notes |
|-------|--------|-------|
| `ActorPrincipalID` / Decision / AnalysisAttempt / Revision.`CreatedBy` / Validation.`ValidatedBy` | `TenantContext.PrincipalID` | Opaque principal string. **Never** CozeUserID. |
| AssetRef `OwnerID` / AssetRef `CreatedBy` | Explicit `OwnerID int64` (= `TenantContext.CozeUserID`) | Must be passed explicitly on ManualCreate / Confirm / EditConfirm when creating Capability. |
| App layer | Pass `ActorID: tc.PrincipalID`, `OwnerID: tc.CozeUserID` | Remove `capabilityActorID` numeric preference. |
| Domain | Remove `parseOwnerID(ActorID)` for ManualCreate/Confirm new-cap path | Require `OwnerID > 0` when creating new Capability/AssetRef; ActorID is PrincipalID only. |

### Domain input deltas

```go
// ManualCreateInput — ActorID = PrincipalID; OwnerID required (>0) for AssetRef
type ManualCreateInput struct {
    TenantID, BusinessID, CapabilityID, ActorID string
    OwnerID int64 // REQUIRED when creating capability (explicit CozeUserID)
    Payload entity.SemanticPayload
}

// ConfirmInput / EditConfirmInput — add OwnerID for first-create Capability path
type ConfirmInput struct {
    TenantID, ProposalID, ActorID, Reason, ClientRequestID, CapabilityID string
    OwnerID int64 // REQUIRED when materializing creates new Capability
}
// EditConfirmInput same OwnerID rule
```

`DeriveInput` / `StartAnalysisInput` / Reject / Validate / Activate / Deprecate: ActorID = PrincipalID only (no OwnerID needed unless creating AssetRef).

---

## Audit metadata validation (LOCKED)

Validate **before** any transaction / persistence:

- Fields: `reason`, `client_request_id` (where present on the op)
- Rules: max length (reason ≤ 1024, client_request_id ≤ 128), valid UTF-8, no control chars (except optional `\n`/`\t` in reason only — prefer reject all controls for both), reject credential-shaped substrings (case-insensitive): `password`, `token`, `cookie`, `authorization`, `bearer`, `jwt`, `api_key`, `api-key`, `private_key`, `private-key`, `secret`, `-----begin`
- On failure: `entity.ErrInvalidPayload` → map to `FORMA_CAPABILITY_INVALID_PAYLOAD`
- **Reject** (do not redact/star). No Decision/Run/Attempt/Revision/status writes.

Apply on: ManualCreate (if reason ever added), Derive/Edit, StartAnalysis, Confirm, EditConfirm, Reject, Activate, Deprecate (reason + client_request_id where applicable).

---

## Analysis FAILED + Retry (LOCKED)

### Error mapping
- `ErrAnalysisFailed` → **new** `CodeCapabilityAnalysisFailed` / `KeyCapabilityAnalysisFailed` = `FORMA_CAPABILITY_ANALYSIS_FAILED` (HTTP 409)
- Must **NOT** map to `FORMA_CAPABILITY_VALIDATION_FAILED`

### Domain/App
- Prefer existing `ClaimAnalysisRetry` / retry path in domain; expose app method `RetryCapabilityAnalysis`
- First StartAnalysis failure: return FAILED run DTO + same analysis_run_id + sanitized `error_code` (HTTP envelope may still be error OR OK-with-FAILED — **prefer**: application returns `(*StartCapabilityAnalysisResponse, error)` where on `ErrAnalysisFailed` the response is non-nil with FAILED run and error is mapped AnalysisFailed; handler returns that data with error envelope OR 200 with failed status — **LOCKED choice**: Handler returns **200 OK** with `data.analysis_run.status=FAILED` and `error_code` when run persisted, OR returns error envelope **with** `data` populated. Match existing patterns if any; else: return FormaError AnalysisFailed **and** include run in a dedicated result type that handler serializes as:
  - Prefer: `writeOK` with FAILED run when OwnedExecute completed to FAILED (so client always sees run).  
  - **LOCKED:** On StartAnalysis / Retry when domain returns `(result, ErrAnalysisFailed)` with non-nil result: **HTTP 200** + DTO (status FAILED). On hard failures without run: map error normally.
- Identical client_request_id replay of FAILED run: return existing FAILED run, **do not** auto-retry.
- Explicit retry only via new route.

### Route
```
POST /api/forma/v1/businesses/:id/capability-analyses/:analysisRunId/retry
```
Auth: OWNER|ADMIN + ACTIVE membership. Body optional `{ "reason"?: string }` — validate secret rules on reason if present.

---

## Handler malformed JSON (LOCKED)

For Confirm, Reject, Activate, Deprecate (and EditConfirm / Retry if body present):
- Empty body → treat as empty optional input (OK)
- Non-empty malformed JSON → HTTP 400, stable bad-request / invalid payload key, **no writes**

---

## Tenancy request audit (LOCKED)

- Domain UoW decision/attempt audit remains authoritative.
- `recordCapabilityAudit` (TenancySVC.RecordAudit): **best-effort**; on failure log internal warning **without** raw error text; never fail the successful business mutation.

---

## File ownership (parallel)

### Agent A ONLY (domain identity + secret validation)
- `domain/forma/capability/service/service.go` (identity + metadata validate + ManualCreate/Confirm OwnerID)
- `domain/forma/capability/service/*` helpers for metadata validation (may add `audit_metadata.go`)
- `domain/forma/capability/service/*_test.go` for identity/secret/no-partial-write
- May touch `entity/errors.go` only if needed (prefer ErrInvalidPayload)
- **DO NOT** touch: router, handlers, application/forma/capability_*, errors/codes.go mapping for AnalysisFailed (Agent B), RESULT

### Agent B ONLY (retry API + analysis mapping)
- `application/forma/capability_analysis_app.go` (**create** — analysis Start/Get/Retry methods; may copy from app)
- `domain/forma/errors/codes.go` — AnalysisFailed mapping only (+ constructor)
- `api/router/forma/api.go` — retry route only
- `api/handler/forma/capability.go` — Retry handler only (or new capability_analysis.go)
- Tests for retry / FAILED visibility under `application/forma/capability_analysis_*_test.go`
- **DO NOT** change PrincipalID/OwnerID semantics beyond consuming frozen contract (ActorID=PrincipalID, OwnerID=CozeUserID)
- **DO NOT** touch domain service identity/parseOwnerID (Agent A)

### Agent C ONLY (handler standards + docs)
- `api/handler/forma/capability.go` / `capability_test.go` — malformed JSON, real middleware/handler envelope tests
- `application/forma/capability_app_test.go` — remove dead `coze` field; role/tenant regressions if needed
- Trailing whitespace fix on `forma/cursor-results/FORMA-S5-G4-CAPABILITY-API-AUTH-AUDIT-RESULT.md`
- Go doc comments on new exported types/methods **in files C owns**
- Best-effort tenancy audit helper if in app (coordinate: prefer `capability_auth.go` after split — C may patch `recordCapabilityAudit` behavior)
- **DO NOT** change domain OwnerID/parseOwnerID or retry domain logic

### Main agent ONLY (after A→B→C)
- Split `capability_app.go` → `capability_dto.go`, `capability_auth.go`, `capability_app.go`, `capability_analysis_app.go`
- Merge conflicts, wire InitService if needed
- Full tests, RESULT `FORMA-S5-G4-F1-...`, commit, push, CI
- `S5_G5_READY = NO`

---

## Branch naming

```
forma/s5-g4-f1-agent-a-identity
forma/s5-g4-f1-agent-b-retry
forma/s5-g4-f1-agent-c-handler
```

Base: `dc1a90b887d5d1dac74e77f8b1aa36677ac6e203`

---

## REAL_MODEL_CALLS = 0

Keep `DeterministicFakeGenerator`. No live model.
