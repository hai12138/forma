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
  /** Coze snowflake ID — always string (never JS number). */
  coze_user_id: string;
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

/** Tenant ↔ Coze Space mapping; coze_space_id is always a decimal string. */
export interface FormaTenantSpace {
  tenant_id: string;
  coze_space_id: string;
  purpose: string;
  status: string;
}

export interface FormaMeData {
  principal: FormaPrincipal;
  current_tenant: FormaTenant | null;
  memberships: FormaMembership[];
  tenants: FormaTenant[];
  coze_user_id?: string;
}

export interface FormaAssetCounts {
  business: number;
  capability: number;
  agent: number;
  application: number;
}

export type FormaSourceMarker = 'AI_GENERATED' | 'MANUAL_MODIFIED';

export interface FormaSemanticNode {
  id: string;
  type: string;
  name: string;
  description?: string;
  properties?: Record<string, unknown>;
  source_marker: FormaSourceMarker;
}

export interface FormaSemanticEdge {
  id: string;
  source: string;
  target: string;
  type: string;
  label?: string;
  description?: string;
  properties?: Record<string, unknown>;
  source_marker: FormaSourceMarker;
}

export interface FormaBusinessRule {
  id: string;
  name: string;
  description?: string;
  expression?: string;
  applies_to?: string[];
  severity?: string;
  source_marker: FormaSourceMarker;
  properties?: Record<string, unknown>;
}

export interface FormaBusinessState {
  id: string;
  object_ref: string;
  name: string;
  description?: string;
  initial?: boolean;
  terminal?: boolean;
  source_marker: FormaSourceMarker;
  properties?: Record<string, unknown>;
}

export interface FormaSemanticModel {
  schema_version: string;
  nodes: FormaSemanticNode[];
  edges: FormaSemanticEdge[];
  rules: FormaBusinessRule[];
  states: FormaBusinessState[];
  evidence_refs?: string[];
  assertion_refs?: string[];
}

export interface FormaViewLayout {
  node_positions: Record<string, { x: number; y: number }>;
  zoom: number;
  viewport: { x: number; y: number };
  mode?: string;
  groups?: string[][];
  collapsed?: Record<string, boolean>;
  canvas?: Record<string, unknown>;
}

export interface FormaBusiness {
  business_id: string;
  asset_id: string;
  name: string;
  status: string;
  current_revision: number;
  schema_version: string;
  updated_at: string;
  created_at: string;
}

export interface FormaModelResponse {
  business_id: string;
  current_revision: number;
  content_digest: string;
  change_summary: string;
  no_change?: boolean;
  semantic_model: FormaSemanticModel;
}

export interface FormaBusinessRevision {
  revision_no: number;
  base_revision_no: number;
  schema_version: string;
  content_digest: string;
  change_summary: string;
  created_by: string;
  created_at: string;
}

export interface FormaRevisionDetail {
  revision: FormaBusinessRevision;
  semantic_model: FormaSemanticModel;
  read_only: boolean;
}

export interface FormaElementDiff {
  added: string[];
  removed: string[];
  modified: string[];
}

export interface FormaDiffResponse {
  diff: {
    from_revision: number;
    to_revision: number;
    nodes: FormaElementDiff;
    edges: FormaElementDiff;
    rules: FormaElementDiff;
    states: FormaElementDiff;
  };
  impact: {
    semantic_changed: boolean;
    affected_node_ids: string[];
    affected_rule_ids: string[];
    affected_state_ids: string[];
  };
}

export interface FormaLayoutResponse {
  business_id: string;
  layout_revision: number;
  based_on_model_revision: number;
  layout: FormaViewLayout;
}

export interface FormaBootstrapData {
  principal: FormaPrincipal;
  tenant: FormaTenant;
  membership?: FormaMembership;
  space?: FormaTenantSpace;
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
    default_space_id?: string;
  }): Promise<FormaApiEnvelope<FormaBootstrapData>> {
    return this.request<FormaBootstrapData>('POST', '/api/forma/v1/bootstrap', body ?? {});
  }

  async listSpaces(tenantId: string): Promise<FormaApiEnvelope<FormaTenantSpace[]>> {
    return this.request<FormaTenantSpace[]>('GET', `/api/forma/v1/tenants/${tenantId}/spaces`);
  }

  async bindSpace(
    tenantId: string,
    body: { coze_space_id: string; purpose?: string },
  ): Promise<FormaApiEnvelope<FormaTenantSpace>> {
    return this.request<FormaTenantSpace>('POST', `/api/forma/v1/tenants/${tenantId}/spaces`, body);
  }

  async assetCounts(): Promise<FormaApiEnvelope<FormaAssetCounts>> {
    return this.request<FormaAssetCounts>('GET', '/api/forma/v1/assets/counts');
  }

  async listBusinesses(): Promise<FormaApiEnvelope<FormaBusiness[]>> {
    return this.request<FormaBusiness[]>('GET', '/api/forma/v1/businesses');
  }

  async createBusiness(body: {
    name: string;
    semantic_model?: FormaSemanticModel;
    change_summary?: string;
  }): Promise<FormaApiEnvelope<FormaBusiness>> {
    return this.request<FormaBusiness>('POST', '/api/forma/v1/businesses', body);
  }

  async getBusiness(id: string): Promise<FormaApiEnvelope<FormaBusiness>> {
    return this.request<FormaBusiness>('GET', `/api/forma/v1/businesses/${id}`);
  }

  async patchBusiness(
    id: string,
    body: { name: string },
  ): Promise<FormaApiEnvelope<FormaBusiness>> {
    return this.request<FormaBusiness>('PATCH', `/api/forma/v1/businesses/${id}`, body);
  }

  async archiveBusiness(id: string): Promise<FormaApiEnvelope<FormaBusiness>> {
    return this.request<FormaBusiness>('POST', `/api/forma/v1/businesses/${id}/archive`);
  }

  async getBusinessModel(id: string): Promise<FormaApiEnvelope<FormaModelResponse>> {
    return this.request<FormaModelResponse>('GET', `/api/forma/v1/businesses/${id}/model`);
  }

  async putBusinessModel(
    id: string,
    body: {
      expected_revision: number;
      semantic_model: FormaSemanticModel;
      change_summary?: string;
    },
  ): Promise<FormaApiEnvelope<FormaModelResponse>> {
    return this.request<FormaModelResponse>('PUT', `/api/forma/v1/businesses/${id}/model`, body);
  }

  async listBusinessRevisions(
    id: string,
  ): Promise<FormaApiEnvelope<FormaBusinessRevision[]>> {
    return this.request<FormaBusinessRevision[]>(
      'GET',
      `/api/forma/v1/businesses/${id}/revisions`,
    );
  }

  async getBusinessRevision(
    id: string,
    revision: number,
  ): Promise<FormaApiEnvelope<FormaRevisionDetail>> {
    return this.request<FormaRevisionDetail>(
      'GET',
      `/api/forma/v1/businesses/${id}/revisions/${revision}`,
    );
  }

  async diffBusiness(
    id: string,
    from: number,
    to: number,
  ): Promise<FormaApiEnvelope<FormaDiffResponse>> {
    return this.request<FormaDiffResponse>(
      'GET',
      `/api/forma/v1/businesses/${id}/diff?from=${from}&to=${to}`,
    );
  }

  async getBusinessLayout(id: string): Promise<FormaApiEnvelope<FormaLayoutResponse>> {
    return this.request<FormaLayoutResponse>('GET', `/api/forma/v1/businesses/${id}/layout`);
  }

  async putBusinessLayout(
    id: string,
    body: {
      expected_layout_revision: number;
      based_on_model_revision: number;
      layout: FormaViewLayout;
    },
  ): Promise<FormaApiEnvelope<FormaLayoutResponse>> {
    return this.request<FormaLayoutResponse>('PUT', `/api/forma/v1/businesses/${id}/layout`, body);
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
