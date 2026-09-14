# S5-G4-F2 — Frozen Interface & File Ownership

**Baseline main:** `042f1520ef307ae2f71ca5f2d8352f44185b641f`  
**F1 integrated:** `14ab0af8f23d7b8974fce36cce12beaa35762e4a`  
**Status:** FROZEN — candidates only; do NOT declare PASS; do NOT push main.

---

## Agent A — HTTP Boundary (ONLY)

```
api/handler/forma/capability.go
api/handler/forma/capability_analysis.go
api/handler/forma/capability_test.go
application/forma/capability_dto.go
```

1. Retry MUST use `bindOptionalCapabilityJSON` (same as Confirm/Reject/Activate/Deprecate).
2. Empty body OK; non-empty malformed → stable `formaerrors.BadRequest("invalid request body")` HTTP 400.
3. Malformed Retry: no ApplicationSVC call, no attempt bump, no generator.
4. Required JSON handlers (Create, Derive, Edit, StartAnalysis, EditConfirm, and any other BindAndValidate required bodies): map bind errors to BadRequest — never raw/500.
5. Go doc comments on all exported DTOs in capability_dto.go and exported handlers touched.
6. Extend real Register→middleware→handler→envelope tests (malformed retry, malformed create, stable error_key, request_id).

Branch: `forma/s5-g4-f2-agent-a-http`

## Agent B — FAILED Run Visibility (ONLY)

```
application/forma/capability_analysis_app.go
application/forma/capability_analysis_app_test.go
```

1. FAILED DTO keep condition:
```
res != nil && res.Run != nil && res.Run.Status == capentity.AnalysisFailed
```
Do NOT require `errors.Is(err, ErrAnalysisFailed)`.
2. Cover generator direct error AND generator returns invalid proposal → FAILED visibility.
3. First fail + retry fail: return DTO (nil err) → HTTP 200 path; same analysis_run_id; status FAILED; attempt correct; sanitized error_code.
4. Hard failure without persisted run → normal error envelope.
5. Real second-tenant retry isolation test.
6. Concurrent retry: assert generator call count == 1 (single executor).

Branch: `forma/s5-g4-f2-agent-b-failed`

## Agent C — Secret Shape & Identity Evidence (ONLY)

```
domain/forma/capability/service/audit_metadata.go
domain/forma/capability/service/identity_audit_test.go
(+ new independent domain test file under service/ if needed)
```

May READ `validator.go` patterns (`containsCredentialShape`, `freeTextAssignmentPatterns`, `ValidateOpaqueID`, `containsSecret`) and call them from audit_metadata — prefer reuse, do not duplicate bare keyword lists. If exporting a helper from validator is needed for reuse, a minimal unexported→shared helper change in validator.go is allowed ONLY for credential detection reuse (keep scope tiny).

1. No bare keyword substring-only detection.
2. Reuse credential-shape: Bearer, JWT, ghp_, github_pat_, sk-/sk-proj-, xox*, PEM, assignment forms, long random/base64.
3. Allow legitimate business text: `review trade secret policy`, `tokenization completed`, `cookie policy approved`, `password reset workflow approved`.
4. `client_request_id`: ValidateOpaqueID + credential-shape + length ≤128; empty OK where optional.
5. Reject before txn; assert no Run/Attempt/Decision/Revision/lifecycle writes.
6. Confirm + EditConfirm first-create identity: Decision/Revision/CreatedBy = PrincipalID; AssetRef OwnerID/CreatedBy = CozeUserID.

Branch: `forma/s5-g4-f2-agent-c-secret`

## Main integrator (after A→B→C)

- Extract shared private FAILED-result helper for Start/Retry.
- `git diff --check <baseline>...HEAD` exit 0 (not warnings-as-PASS).
- Full tests + RESULT + CI; `S5_G5_READY = NO`.
- MIGRATION_CHANGE=NONE; REAL_MODEL_CALLS=0; G5_CHANGE=NONE.
