/** Aligns with backend entity.SourceType / AdapterType (source.go). */
export const SOURCE_TYPES = ['RELATIONAL_DATABASE', 'HTTP_API'] as const;
export type FormaUiSourceType = (typeof SOURCE_TYPES)[number];

export const RELATIONAL_ADAPTERS = ['MYSQL', 'POSTGRESQL'] as const;
export const HTTP_ADAPTERS = ['HTTP'] as const;

export function isValidSourceType(v: string): v is FormaUiSourceType {
  return (SOURCE_TYPES as readonly string[]).includes(v);
}

export function adaptersForSourceType(sourceType: string): readonly string[] {
  if (sourceType === 'HTTP_API') return HTTP_ADAPTERS;
  if (sourceType === 'RELATIONAL_DATABASE') return RELATIONAL_ADAPTERS;
  return [];
}

export function defaultAdapterForSourceType(sourceType: string): string {
  const adapters = adaptersForSourceType(sourceType);
  return adapters[0] ?? '';
}

export function defaultPortForAdapter(adapterType: string): number {
  if (adapterType === 'POSTGRESQL') return 5432;
  if (adapterType === 'MYSQL') return 3306;
  return 0;
}

export type ConnectionFormFields = {
  host?: string;
  port?: string | number;
  database?: string;
  username?: string;
  base_url?: string;
  openapi_url?: string;
};

/**
 * Build public_config for CreateConnection.
 * HTTP_API → base_url / openapi_url; RELATIONAL_DATABASE → host/port/database/username.
 */
export function buildConnectionPublicConfig(
  sourceType: string,
  adapterType: string,
  form: ConnectionFormFields,
): Record<string, unknown> {
  if (sourceType === 'HTTP_API' || adapterType === 'HTTP') {
    const cfg: Record<string, unknown> = {
      base_url: String(form.base_url ?? '').trim(),
    };
    const openapi = String(form.openapi_url ?? '').trim();
    if (openapi) {
      cfg.openapi_url = openapi;
    }
    return cfg;
  }
  const portRaw = form.port;
  const port =
    typeof portRaw === 'number'
      ? portRaw
      : Number(portRaw) || defaultPortForAdapter(adapterType);
  return {
    host: String(form.host ?? '').trim(),
    port,
    database: String(form.database ?? '').trim(),
    username: String(form.username ?? '').trim(),
  };
}

export function buildCreateConnectionBody(input: {
  name: string;
  environment: string;
  sourceType: string;
  adapterType: string;
  form: ConnectionFormFields;
  credential_ref_id?: string;
}): {
  name: string;
  environment: string;
  adapter_type: string;
  public_config: Record<string, unknown>;
  credential_ref_id?: string;
} {
  const allowed = adaptersForSourceType(input.sourceType);
  const adapterType = allowed.includes(input.adapterType)
    ? input.adapterType
    : defaultAdapterForSourceType(input.sourceType);
  const body: {
    name: string;
    environment: string;
    adapter_type: string;
    public_config: Record<string, unknown>;
    credential_ref_id?: string;
  } = {
    name: input.name,
    environment: input.environment,
    adapter_type: adapterType,
    public_config: buildConnectionPublicConfig(input.sourceType, adapterType, input.form),
  };
  if (input.credential_ref_id) {
    body.credential_ref_id = input.credential_ref_id;
  }
  return body;
}
