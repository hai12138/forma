# S5-G4-F8-F2 — Go Vet Copylocks + Race CI Closure INTERFACE FREEZE

**Baseline main:** `1081a2f509203a238abc4125332fc0022bc147e6`
**Prior gate:** S5-G4-F8-F1 = BLOCKED_BY_EXISTING_GO_VET
**Status:** FROZEN — main integrator only; do not start S5-G5; no freeze tag.

---

## Scope (do not expand)

Close the F8-F1 `go vet` copylocks block and obtain durable CI evidence that:

1. `Forma capability static analysis` (`go vet`) PASS
2. `Forma capability race tests` (`go test -race`) actually runs and PASS
3. Existing backend tests + `Forma pre-commit CWD check` still execute and PASS

**Out of scope:** production capability code, `FakeContractPort` redesign, mutex changes, migration, frontend, Runtime, G5, real model calls, rewriting F8 / F8-F1 history.

---

## Exact file allowlist

```text
coze-studio/backend/domain/forma/capability/service/validate_g3_test.go
.github/workflows/forma-ci.yml
forma/cursor-results/FORMA-S5-G4-F8-F1-EVIDENCE-CORRECTION-RESULT.md
forma/cursor-results/FORMA-S5-G4-F8-F2-GO-VET-RACE-CLOSURE-FREEZE.md
forma/cursor-results/FORMA-S5-G4-F8-F2-GO-VET-RACE-CLOSURE-RESULT.md
docs/agents/issue-tracker.md
```

Any other path required → STOP; revise this freeze first.

---

## Task contracts

### 1. Vet fix (test-only, minimal)

In `validate_g3_test.go`, **only delete**:

```go
bad := *contracts
_ = bad
```

Do **not** modify `FakeContractPort`, mutex fields, production code, or other test assertions / behavior.

### 2. CI step order

Keep step commands unchanged:

```yaml
- name: Forma capability static analysis
  run: go vet ./application/forma/... ./domain/forma/capability/...

- name: Forma capability race tests
  run: go test -race ./application/forma/... ./domain/forma/capability/... -count=1
```

Move `Forma pre-commit CWD check` **before** the new vet/race steps so a new-check failure cannot skip the pre-existing check.

### 3. Evidence rules

- No local Go path guessing / local Go install claims.
- If race exposes a new defect → STOP; do not expand allowlist.
- Preserve F8-F1 BLOCKED evidence; append superseded-by-F8-F2 only.
- `IMPLEMENTATION_SHA` pairs with its CI run; RESULT tip SHA/run reported separately.

---

## Reviewers (after GREEN tip CI for implementation)

| Reviewer | Scope |
|----------|-------|
| S | mutex-copy deleted; test semantics unchanged; vet/race really executed |
| Q | file whitelist; CI step order; SHA/run pairing; audit wording; no scope creep |

Reviewers are read-only; findings return to main integrator.

---

## Stop

`S5_G5_READY = NO`. Do not create `forma-s5-frozen`.
