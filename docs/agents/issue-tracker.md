# Issue tracker: GitHub

Issues and specs for this repo live as **GitHub Issues**. Use the `gh` CLI for all issue-tracker operations.

Infer the repository from `git remote` (typically `origin`). `gh` does this automatically when run inside a clone of `hai12138/forma`.

**This document is configuration only.** Creating this file does **not** authorize creating, closing, labeling, or commenting on any GitHub Issue unless a later task explicitly requires it.

## Conventions

- **Create an issue:** `gh issue create --title "..." --body "..."`. Prefer a heredoc for multi-line bodies.
- **Read an issue:** `gh issue view <n> --comments` (filter comments/labels with `jq` when needed).
- **List issues:**
  `gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'`
  Add `--label` / `--state` filters as appropriate.
- **Comment:** `gh issue comment <n> --body "..."`
- **Apply / remove labels:** `gh issue edit <n> --add-label "..."` / `--remove-label "..."`
- **Close:** `gh issue close <n> --comment "..."`

## Pull requests as a triage surface

**PRs as a request surface: no.**

Do not treat external pull requests as the primary feature-request / triage intake for this repo unless this flag is later changed to `yes`.

GitHub shares one number space across issues and PRs. A bare `#N` may be either:

1. First try `gh pr view N`
2. If that fails, try `gh issue view N`

## When a skill says "publish to the issue tracker"

Create a GitHub Issue with `gh issue create` (only when the skill/task explicitly requires publishing).

## When a skill says "fetch the relevant ticket"

Run `gh issue view <n> --comments`.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single GitHub Issue; **child** Issues are decision tickets.

- **Map:** one Issue labelled `wayfinder:map`, holding Notes / Decisions-so-far / Fog. Create with `gh issue create --label wayfinder:map` (only when explicitly starting wayfinding).
- **Child ticket:** an Issue linked to the map as a GitHub **sub-issue** (`gh api` on the sub-issues endpoint). Where sub-issues are unavailable, add the child to a task list in the map body and put `Part of #<map>` at the top of the child body. Labels: `wayfinder:<kind>` where `<kind>` ∈ `research` / `prototype` / `grilling` / `task`. Once claimed, assign the ticket to the driving developer.
- **Blocking:** prefer GitHub native issue dependencies (UI-visible). Add an edge with:
  `gh api --method POST repos/<owner>/<repo>/issues/<blocked_number>/dependencies/blocked_by -F issue_id=<blocker_database_id>`
  where `<blocker_database_id>` is from `gh api repos/<owner>/<repo>/issues/<blocker_number> --jq .id` (not `#number` / `node_id`).
  GitHub reports `issue_dependencies_summary.blocked_by` (open blockers only — the live gate).
  Fallback when dependencies are unavailable: a `Blocked by: #a, #b` line at the top of the child body.
  A ticket is **unblocked** when every blocker is closed.
- **Frontier query:** list the map’s open children (`gh issue list --state open`, scoped to the map’s sub-issues / task list); drop any with an open blocker (`issue_dependencies_summary.blocked_by > 0`, or an open issue in the `Blocked by` line) or an assignee; the **first in map order** wins.
- **Claim:** `gh issue edit <n> --add-assignee @me` — the session’s first write on that ticket.
- **Resolve:** `gh issue comment <n> --body "<decision>"`, then `gh issue close <n>`, then append a context pointer (short gist + link) to the map’s Decisions-so-far.

## Forma stage gate notes (audit pointer only)

This section does **not** create, close, label, or comment on any GitHub Issue.

| Gate | Status | Pointer |
|------|--------|---------|
| S5-G4-F7 Capability UI Read Seam | PASS (superseded gaps → F8) | `forma/cursor-results/FORMA-S5-G4-F7-CAPABILITY-UI-READ-SEAM-RESULT.md` |
| S5-G4-F8 Read Seam Fail-Closed | PASS / CI ALL GREEN (evidence corrected by F8-F1) | `forma/cursor-results/FORMA-S5-G4-F8-READ-SEAM-FAIL-CLOSED-RESULT.md` |
| S5-G4-F8-F1 Evidence Correction | BLOCKED_BY_EXISTING_GO_VET (history preserved; closed by F8-F2) | `forma/cursor-results/FORMA-S5-G4-F8-F1-EVIDENCE-CORRECTION-RESULT.md` |
| S5-G4-F8-F2 Go Vet / Race Closure | PASS / CI ALL GREEN | `forma/cursor-results/FORMA-S5-G4-F8-F2-GO-VET-RACE-CLOSURE-RESULT.md` |
| IMPLEMENTATION_SHA (F8) | `27059ef3a884e941ab526e59dbe1e4c44fa9226b` | CI https://github.com/hai12138/forma/actions/runs/34912401820 |
| IMPLEMENTATION_SHA (F8-F2) | `f073b24dd7bb967bb0cf23c06bfe4af937c55be4` | CI https://github.com/hai12138/forma/actions/runs/34937595390 |
| RED_SHA (F8 Agent A) | `aa36bbab08c1ad290a07db80f65bde1c74457d39` | CI https://github.com/hai12138/forma/actions/runs/34911998223 FAILURE |
| S5-G5 | **NOT STARTED** | `S5_G5_READY = NO`; do not create `forma-s5-frozen` |
