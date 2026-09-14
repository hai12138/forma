# S5-G4-F6 — Frozen Ownership & Contracts

**Baseline main:** `937b0a47a02c0a1ac55e5696cf16626a7ab43881`
**Status:** FROZEN — candidates only; agents do NOT push main.

---

## Agent A — Boundary/Unicode Tests (ONLY)

```
domain/forma/capability/service/bearer_secret_f6_test.go  (new only)
```

ALLOW: exact `bearer of responsibility` (+ case + surrounding TrimSpace whitespace).
REJECT: prefix/suffix/repetition/punctuation attachments; Unicode whitespace after Bearer.
No production code. Branch: `forma/s5-g4-f6-agent-a-tests`

## Agent B — Unicode-Aware Implementation (ONLY)

```
domain/forma/capability/service/validator.go
```

1. Allow ONLY when `strings.EqualFold(strings.TrimSpace(s), "bearer of responsibility")`.
2. No unanchored regex deletion of safe phrase from mid-input.
3. Elsewhere: standalone `Bearer` + Unicode whitespace(s) + non-whitespace → credential (`unicode.IsSpace`).
4. No bare-keyword bans; no FAILED/API/Migration/G5/Runtime changes.

Branch: `forma/s5-g4-f6-agent-b-implementation`

## Main integrator (TDD)

1. Integrate A — prove F6 red on old impl.
2. Integrate B — all green (adjust prior allow tests only if they conflict with whole-input policy).
3. `git diff --check 937b0a47...HEAD` exit 0; RESULT + CI; `S5_G5_READY = NO`.
4. No G5 / Runtime / Freeze Tag.
