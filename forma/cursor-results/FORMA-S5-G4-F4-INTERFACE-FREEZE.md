# S5-G4-F4 — Frozen Ownership & Contracts

**Baseline main:** `9472cae79cb291fff85ef2b045c302e22b62e0c6`
**Status:** FROZEN — candidates only; do NOT declare PASS; do NOT push main from agents.

---

## Agent A — Bearer Secret Totality (ONLY)

```
domain/forma/capability/service/validator.go
domain/forma/capability/service/audit_metadata.go
domain/forma/capability/service/*_f4_test.go  (new)
```

1. Reject any non-empty, non-whitespace value after `Bearer` — NOT limited to `[A-Za-z0-9._-]`.
2. Must reject at least: `Bearer /abc`, `Bearer +abc`, `Bearer ~abc`, `Bearer "abc"`, `Bearer of`.
3. Continue allowing full business phrase `bearer of responsibility`.
4. If the same string has both a safe business phrase AND another Bearer credential, still reject.
5. Do NOT restore bare-keyword bans for `secret` / `token` / `cookie` / `password`.
6. Tests: `containsSecret`, `ValidateAuditMetadata`, illegal metadata → no partial write.

Branch: `forma/s5-g4-f4-agent-a-bearer`

## Agent B — FAILED Attempt Attribution (ONLY)

```
domain/forma/capability/service/service.go
domain/forma/capability/service/*_f4_test.go  (new)
```

1. Extend `returnPersistedFailedAnalysis` to take `expectedAttempt`.
2. Return FAILED DTO only when refetch OK AND `Status == FAILED` AND `Attempt == expectedAttempt`.
3. Status/attempt mismatch → fail closed (`nil` result + stable sanitized consistency error); never forge DTO.
4. Pass the caller's owned attempt on all paths: generator error, invalid proposal/persist, Retry load failure, lease takeover load failure.
5. Tests: exact equality (not `Attempt >= expected`); interleaved N mark-fail then refetch sees N+1 → old request must NOT return N+1 DTO.
6. Same `analysis_run_id`, attempt monotonic, no duplicate Proposal.
7. Replace inaccurate "when mark stuck" comments with accurate persisted mark-failed wording.

Branch: `forma/s5-g4-f4-agent-b-attempt`

## Main integrator

- Integrate A then B after both candidates pass their tests (no `merge(...)` commit type).
- Do not rewrite pushed F3 history.
- `git diff --check 9472cae7...HEAD` exit 0.
- Full tests + RESULT + CI; `S5_G5_READY = NO`.
- MIGRATION_CHANGE=NONE; REAL_MODEL_CALLS=0; G5_CHANGE=NONE.
- No S5-G5 / Runtime / Agent-Workflow projection / Freeze Tag.
