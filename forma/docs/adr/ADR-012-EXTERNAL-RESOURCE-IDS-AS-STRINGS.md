# ADR-012: External Resource IDs as Strings

## Context

Coze resource identifiers (user, space, agent, workflow, …) are int64 Snowflake values. JavaScript `Number` is IEEE-754 float64 and only safely represents integers up to `Number.MAX_SAFE_INTEGER` (`9007199254740991`). Serializing Coze IDs as JSON numbers causes silent precision loss in browsers and Node when IDs exceed that limit.

S1-G1 live E2E already observed this risk when round-tripping personal space IDs through JS.

## Decision

From S1-G2 onward, all **public** Forma API / frontend / package / channel contracts expose Coze (and other opaque external) resource IDs as **decimal strings**.

Domain and repository layers may continue to use `int64` internally. Conversion is explicit via `idcontract.FormatCozeID` / `ParseCozeID`.

Future Coze/MCP/external opaque IDs default to string unless the value is a true mathematical quantity.

## Consequences

- Public DTOs (`PrincipalDTO`, `TenantSpaceDTO`, bootstrap/bind inputs) use `string`
- JSON number for `coze_*_id` is rejected (Go unmarshal type error or validation)
- Frontend `@forma/api-client` types use `string` only — no `number | string`
- Precision regression tests use values above JS safe integer

## Rejected Alternatives

1. **Keep JSON numbers** — unacceptable silent corruption above safe integer  
2. **Dual `number | string`** — invites accidental number paths; product not yet released  
3. **Always use string in domain** — fights Coze foundation int64 and SQL types  
4. **Base64 / ULID rewrite of Coze IDs** — unnecessary; decimal string preserves interoperability
