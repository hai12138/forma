# Forma API (S0-B)

Independent Forma API namespace. Coze API remains under `/api/*` generated from Thrift IDL.

## Namespace

`/api/forma/v1/`

## S0-B Endpoints

| Method | Path | Auth |
|---|---|---|
| GET | `/api/forma/v1/health` | Public |
| GET | `/api/forma/v1/version` | Public |
| GET | `/api/forma/v1/meta/baseline` | Public |

## Router

Manual registration in `backend/api/router/forma/` (not hz-generated).

## IDL

Future business APIs will live under `idl/forma/`. S0-B uses hand-written handlers only.
