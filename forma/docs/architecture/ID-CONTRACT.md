# Forma External Resource ID Contract

**Status:** Frozen (S1-G2)  
**Applies to:** Backend public API, frontend packages, JSON channels, MCP / external system boundaries

## Rules

### Internal (domain / repository / DAL)

- Coze / Snowflake / opaque native IDs may remain `int64` where the Coze foundation already uses int64.
- Do not treat these values as safe JavaScript numbers.

### External / public

All Coze resource IDs crossing Backend API ↔ Frontend ↔ JSON ↔ Package boundaries **MUST** be serialized as **decimal strings**.

Includes (non-exhaustive):

| Field | Public type |
|---|---|
| `coze_user_id` | `string` |
| `coze_space_id` | `string` |
| `coze_agent_id` | `string` (future) |
| `coze_workflow_id` | `string` (future) |
| `coze_plugin_id` | `string` (future) |
| `coze_knowledge_id` | `string` (future) |
| `coze_app_id` | `string` (future) |
| `coze_database_id` | `string` (future) |

Same default for MCP and any other external opaque IDs unless the value is a true mathematical quantity.

### Mapping

```
Domain int64
  → FormatCozeID → API DTO string → JSON string → Frontend string

Frontend string
  → API string → ParseCozeID → Domain int64
```

Helpers: `backend/domain/forma/idcontract` (`FormatCozeID` / `ParseCozeID`).

### Forbidden

- `int64` → JSON number → JavaScript `number` → `int64`
- Accepting JSON numbers for Coze ID fields
- Floats, scientific notation, leading zeros, zero, negative
- Dual types `number | string` for compatibility

### Validation (`ParseCozeID`)

- Non-empty after trim
- **ASCII** digits only (`'0'`–`'9'`) — Unicode digits (Arabic-Indic, full-width, etc.) rejected
- Form `[1-9][0-9]*` (no leading zeros)
- `> 0`
- Fits in signed int64 (overflow rejected)
- No `.` / `e` / `E` / `+` / `-`
