# Forma Database Migrations

Forma uses an **independent migration namespace** under `docker/atlas/forma/`. Coze core migrations remain in `docker/atlas/migrations/`.

## Tables

| Table | Purpose |
|---|---|
| `forma_asset_ref` | Forma asset header registry |
| `forma_coze_resource_ref` | Forma asset ↔ Coze resource mapping (implements `coze_resource_ref` contract) |

No foreign keys to Coze core tables. Mapping uses stable numeric Coze resource IDs only.

## Startup Order

1. MySQL starts and applies Coze schema (`schema.sql` + Coze Atlas migrations)
2. Apply Forma migrations:

```bash
export FORMA_ATLAS_URL="mysql://coze:coze123@localhost:3306/opencoze?charset=utf8mb4&parseTime=True"
cd docker/atlas/forma
atlas migrate apply --dir "file://migrations" --url "$FORMA_ATLAS_URL"
```

## Developer Workflow

```bash
# After schema change
atlas migrate diff update --env forma-local --to "$FORMA_ATLAS_URL"
atlas migrate hash
atlas migrate apply --dir "file://migrations" --url "$FORMA_ATLAS_URL"
```

## Safety

- Migrations use `CREATE TABLE IF NOT EXISTS` for idempotent fresh installs
- Re-running `atlas migrate apply` on an applied revision is a no-op
- Never add Forma columns to Coze core tables
