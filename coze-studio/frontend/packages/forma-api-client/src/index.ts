export type FormaApiErrorCode =
  | 'NETWORK_ERROR'
  | 'UNAUTHORIZED'
  | 'HTTP_ERROR'
  | 'PARSE_ERROR'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'FORMA_ERROR';

export class FormaApiError extends Error {
  readonly code: FormaApiErrorCode;
  readonly status?: number;
  readonly requestId?: string;
  readonly errorKey?: string;

  constructor(
    code: FormaApiErrorCode,
    message: string,
    options?: { status?: number; requestId?: string; errorKey?: string },
  ) {
    super(message);
    this.name = 'FormaApiError';
    this.code = code;
    this.status = options?.status;
    this.requestId = options?.requestId;
    this.errorKey = options?.errorKey;
  }
}

export interface FormaApiEnvelope<T> {
  code: number;
  msg: string;
  request_id: string;
  data: T;
  error_key?: string;
}

export interface FormaHealthData {
  status: string;
}

export interface FormaVersionData {
  forma_version: string;
  forma_schema_version: string;
}

export interface FormaBaselineData {
  forma_version: string;
  forma_schema_version: string;
  forma_baseline_tag: string;
  coze_baseline_commit: string;
  workspace_baseline_commit: string;
  runtime_foundation: string;
}

export interface FormaPrincipal {
  principal_id: string;
  principal_type: string;
  display_name: string;
  coze_user_id: number;
  status: string;
}

export interface FormaTenant {
  tenant_id: string;
  tenant_key: string;
  name: string;
  display_name: string;
  status: string;
  revision: number;
  role?: string;
}

export interface FormaMembership {
  tenant_id: string;
  principal_id: string;
  role: string;
  status: string;
  revision: number;
}

export interface FormaMeData {
  principal: FormaPrincipal;
  current_tenant: FormaTenant | null;
  memberships: FormaMembership[];
  tenants: FormaTenant[];
}

export interface FormaAssetCounts {
  business: number;
  capability: number;
  agent: number;
  application: number;
}

export interface FormaBootstrapData {
  principal: FormaPrincipal;
  tenant: FormaTenant;
  created: boolean;
}

export interface FormaApiClientOptions {
  baseUrl?: string;
  fetchImpl?: typeof fetch;
  getRequestId?: () => string;
  getTenantId?: () => string | undefined;
}

export class FormaApiClient {
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;
  private readonly getRequestId: () => string;
  private readonly getTenantId?: () => string | undefined;

  constructor(options: FormaApiClientOptions = {}) {
    this.baseUrl = options.baseUrl ?? '';
    this.fetchImpl = options.fetchImpl ?? fetch.bind(globalThis);
    this.getRequestId =
      options.getRequestId ??
      (() => `forma-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`);
    this.getTenantId = options.getTenantId;
  }

  async health(): Promise<FormaApiEnvelope<FormaHealthData>> {
    return this.request<FormaHealthData>('GET', '/api/forma/v1/health');
  }

  async version(): Promise<FormaApiEnvelope<FormaVersionData>> {
    return this.request<FormaVersionData>('GET', '/api/forma/v1/version');
  }

  async baseline(): Promise<FormaApiEnvelope<FormaBaselineData>> {
    return this.request<FormaBaselineData>('GET', '/api/forma/v1/meta/baseline');
  }

  async me(): Promise<FormaApiEnvelope<FormaMeData>> {
    return this.request<FormaMeData>('GET', '/api/forma/v1/me');
  }

  async listTenants(): Promise<FormaApiEnvelope<FormaTenant[]>> {
    return this.request<FormaTenant[]>('GET', '/api/forma/v1/tenants');
  }

  async bootstrap(body?: {
    display_name?: string;
    default_space_id?: number;
  }): Promise<FormaApiEnvelope<FormaBootstrapData>> {
    return this.request<FormaBootstrapData>('POST', '/api/forma/v1/bootstrap', body ?? {});
  }

  async assetCounts(): Promise<FormaApiEnvelope<FormaAssetCounts>> {
    return this.request<FormaAssetCounts>('GET', '/api/forma/v1/assets/counts');
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<FormaApiEnvelope<T>> {
    const requestId = this.getRequestId();
    const headers: Record<string, string> = {
      Accept: 'application/json',
      'X-Request-ID': requestId,
    };
    const tenantId = this.getTenantId?.();
    if (tenantId) {
      headers['X-Forma-Tenant'] = tenantId;
    }
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json';
    }

    let response: Response;
    try {
      response = await this.fetchImpl(`${this.baseUrl}${path}`, {
        method,
        headers,
        credentials: 'include',
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (error) {
      throw new FormaApiError(
        'NETWORK_ERROR',
        error instanceof Error ? error.message : 'Network request failed',
        { requestId },
      );
    }

    let envelope: FormaApiEnvelope<T> | undefined;
    try {
      envelope = (await response.json()) as FormaApiEnvelope<T>;
    } catch {
      envelope = undefined;
    }

    if (response.status === 401) {
      throw new FormaApiError('UNAUTHORIZED', envelope?.msg || 'Unauthorized', {
        status: 401,
        requestId: envelope?.request_id || requestId,
        errorKey: envelope?.error_key,
      });
    }

    if (response.status === 403) {
      throw new FormaApiError('FORBIDDEN', envelope?.msg || 'Forbidden', {
        status: 403,
        requestId: envelope?.request_id || requestId,
        errorKey: envelope?.error_key,
      });
    }

    if (response.status === 404) {
      throw new FormaApiError('NOT_FOUND', envelope?.msg || 'Not found', {
        status: 404,
        requestId: envelope?.request_id || requestId,
        errorKey: envelope?.error_key,
      });
    }

    if (!response.ok) {
      throw new FormaApiError('HTTP_ERROR', envelope?.msg || `HTTP ${response.status}`, {
        status: response.status,
        requestId: envelope?.request_id || requestId,
        errorKey: envelope?.error_key,
      });
    }

    if (!envelope) {
      throw new FormaApiError('PARSE_ERROR', 'Invalid JSON response', {
        status: response.status,
        requestId,
      });
    }

    return { ...envelope, request_id: envelope.request_id || requestId };
  }
}

export function createFormaApiClient(options?: FormaApiClientOptions) {
  return new FormaApiClient(options);
}
