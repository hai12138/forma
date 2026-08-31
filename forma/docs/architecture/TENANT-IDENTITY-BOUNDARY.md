# Tenant / Identity Boundary

## Principle

```
Coze User  ≠  Forma Tenant
Coze Space ≠  Forma Tenant
```

Frozen relation:

```
Forma Tenant  1:N  Coze Space
```

## Relationship Diagram

```
Coze Authenticated Session
          │
          ▼
   FormaPrincipal  (provider=coze, coze_user_id)
          │
          │  N:M via forma_tenant_membership
          ▼
     Forma Tenant
          │
          ├── Forma Assets (tenant_id scoped)
          │
          └── forma_tenant_space_ref (1:N)
                    │
                    ▼
              Coze Space
                    │
        Agent / Workflow / Plugin / Knowledge / App
```

## Trust Boundary

| Input | Trusted? |
|---|---|
| Coze session cookie | Yes (after SessionAuthMW) |
| `X-Forma-Tenant` header | Selection only — must pass membership check |
| JSON `tenant_id` / `role` / `principal_id` | **Never** as identity proof |

## TenantContext (server-generated)

```
tenant_id
principal_id
membership_role
coze_user_id
allowed_space_ids
request_id
tenant_status
```

Built by: Session → Principal Resolver → Membership → Selected Tenant → Space refs.

## Space ACL

Forma uses `crossdomain/user.GetUserSpaceList` via `CozeSpaceAdapter`.  
Forma tenancy domain must **not** import `domain/user/internal/dal`.
