# S5-G4-F8 — Read Seam Fail-Closed INTERFACE FREEZE

**Baseline main:** `9cc0f4f0911105f630a090f2d4ea87f0be1f5813`
**Prior gate:** S5-G4-F7 PASS / CI ALL GREEN
**Status:** FROZEN — candidates only; agents MUST NOT push main or declare PASS.

---

## Scope (do not expand)

F8 closes F7 UI Read Seam gaps only:

1. Tenant / requested-ID identity checks on Proposal and Analysis lookups.
2. Proposal list totality (nil / blank id / mismatch / duplicate → fail closed, no partial list).
3. Matching adversarial tests + audit docs.

**Out of scope:** migration, frontend, Runtime, G5, real model calls, DTO/route/domain interface/repository/DB changes, rewriting F7 Freeze history.

---

## F7 freeze deviation record (do not rewrite F7 Freeze)

F7 Agent B allowlist omitted files that were later required for compile stubs / red-team harden:

| Deviation | Evidence |
|-----------|----------|
| `capability_auth.go` edited on main after F7 freeze (integrator red-team harden `546ba5a8`) | Not in F7 Agent B production allowlist |
| Compile stubs outside allowlist on candidate `b8b0d84b` | `capability_test.go`, `service_test.go`, `persisted_failed_f3_test.go` |
| Candidate commit type `feat(...)` on Agent B branch | Rewritten to `fix(...)` on main cherry-pick |

F8 does **not** rewrite F7 Freeze. Deviations are recorded here for audit.

---

## Frozen fail-closed contracts

### `requireCapabilityProposal` (capability_auth.go)

Before success return, ALL must hold:

```text
prop != nil
prop.TenantID == tenantID
prop.BusinessID == businessID
prop.ProposalID == proposalID
```

Any failure → `MapDomainError(ErrProposalNotFound)` (CapabilityProposalNotFound). Never leak foreign entity fields.

### `requireCapabilityAnalysis` (capability_auth.go)

Before success return, ALL must hold:

```text
run != nil
run.TenantID == tenantID
run.BusinessID == businessID
run.AnalysisRunID == analysisRunID
```

Any failure → `MapDomainError(ErrAnalysisNotFound)` (CapabilityAnalysisNotFound).

### `ListCapabilityProposalsByAnalysis` (capability_analysis_app.go)

Reject (no silent skip, no partial list) when any row is:

```text
nil proposal
tenant / business / analysis_run_id mismatch
blank proposal_id
duplicate proposal_id
```

All map to `ErrConsistency` → CapabilityConflict. Legal lists remain stable-sorted by `proposal_id`; terminal statuses readable; reads do not mutate.

---

## Agent ownership (Wave 1)

### Agent A — Tests ONLY

```text
coze-studio/backend/application/forma/capability_read_seam_f8_test.go  (new only)
```

- Embed `capsvc.CapabilityService` in malicious return wrappers (do not enlarge old test stubs for the full interface).
- Must RED on pre-F8 code for identity + list totality cases.

Branch: `forma/s5-g4-f8-agent-a-tests`

### Agent B — Production ONLY

```text
coze-studio/backend/application/forma/capability_auth.go
coze-studio/backend/application/forma/capability_analysis_app.go
```

Branch: `forma/s5-g4-f8-agent-b-implementation`

If additional production files are required → STOP; main integrator revises this freeze.

---

## Main integrator

1. Cherry-pick A → record RED.
2. Cherry-pick B → GREEN.
3. Full backend + race + static + `git diff --check 9cc0f4f0...HEAD`.
4. Dual reviewers (S = Stage/isolation/totality; Q = Freeze whitelist/TDD/scope/CI pairing).
5. Append F7 RESULT supersession note; write F8 RESULT; update `docs/agents/issue-tracker.md`.
6. `IMPLEMENTATION_SHA` and its CI run must pair; RESULT tip SHA/CI reported separately.
7. `S5_G5_READY = NO`. No `forma-s5-frozen`. No G5.
