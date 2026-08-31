# Upstream Strategy

## Remotes

| Remote | URL | Purpose |
|---|---|---|
| `origin` | `https://github.com/hai12138/forma.git` | Forma product repository |
| `upstream` | `https://github.com/coze-dev/coze-studio.git` | Coze Studio open source |

## Baseline

| Item | Value |
|---|---|
| Tag | `forma-baseline-0` |
| Workspace commit | `98c7aca26d64ac602dc7c0227e2bce38d89666a8` |
| Upstream reference (at S0-B) | `fefb05ff27be1da939612fbf9faf5db62583b8ae` (`upstream/main`) |
| COZE_BASELINE_COMMIT | Recorded in `backend/domain/forma/meta/version.go` |

## Merge Process (not executed in S0-B)

1. `git fetch upstream`
2. `git checkout main`
3. `git merge upstream/main` (prefer merge over rebase for audit trail)
4. Run compatibility gates (below)
5. Review all files touched in `COZE-CORE-PATCHES.md` FORMA blocks
6. Human sign-off before release

## Compatibility Gates

### Backend

- `go test ./domain/forma/...`
- `go test ./crossdomain/forma/...`
- Forma ACL arch test (`domain/forma/arch_test.go`)
- Apply Forma migrations on fresh DB

### Frontend

- `rush build --to @forma/app`
- `@forma/app` typecheck + route smoke tests
- `@coze-studio/app` build (no regression)

### Integration

- `GET /api/forma/v1/health`
- `GET /api/forma/v1/version`
- `GET /api/forma/v1/meta/baseline`
- Coze `/api/passport/*` still reachable

## High-Risk Upstream Conflict Files

- `backend/application/application.go`
- `backend/api/router/register.go`
- `backend/api/middleware/session.go`
- `rush.json`
- `docker/atlas/migrations/*` (Coze only — do not merge Forma SQL here)

## S0-B Note

No upstream merge performed in this stage. Mechanism and baseline tags only.
