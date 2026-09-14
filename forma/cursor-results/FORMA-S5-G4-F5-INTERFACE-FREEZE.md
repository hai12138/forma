# S5-G4-F5 — Frozen Ownership & Contracts

**Baseline main:** `353db3daf9883c366e0de540e498fbbc46ff1fa6`
**Status:** FROZEN — candidates only; do NOT declare PASS; agents do NOT push main.

---

## Agent A — Adversarial Tests (ONLY)

```
domain/forma/capability/service/bearer_secret_f5_test.go  (new only)
```

Must assert REJECT / ALLOW matrix for safe-phrase tightening.
No production code changes.
Branch: `forma/s5-g4-f5-agent-a-tests`

## Agent B — Safe-Phrase Implementation (ONLY)

```
domain/forma/capability/service/validator.go
```

1. Keep `Bearer + any non-empty non-whitespace` detection.
2. Remove open `bearer of <anything>` exemption.
3. Exact allowlist only: `bearer of responsibility` (case-insensitive).
4. Safe boundaries: phrase + extra token / second Bearer / credential → still reject.
5. No bare-keyword bans; no FAILED attempt changes.

Branch: `forma/s5-g4-f5-agent-b-implementation`

## Main integrator (TDD)

1. Integrate A first — confirm F5 tests fail on old impl.
2. Integrate B — confirm all green (adjust F4 open-phrase allow cases if they conflict with F5 policy).
3. `git diff --check 353db3da...HEAD` exit 0.
4. RESULT + CI; `S5_G5_READY = NO`.
5. No G5 / Runtime / Freeze Tag.
