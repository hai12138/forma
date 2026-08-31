export type FormaApiErrorCode =
  | 'NETWORK_ERROR'
  | 'UNAUTHORIZED'
  | 'HTTP_ERROR'
  | 'PARSE_ERROR';

export class FormaApiError extends Error {
  readonly code: FormaApiErrorCode;
  readonly status?: number;
  readonly requestId?: string;

  constructor(
    code: FormaApiErrorCode,
    message: string,
    options?: { status?: number; requestId?: string },
  ) {
    super(message);
    this.name = 'FormaApiError';
    this.code = code;
    this.status = options?.status;
    this.requestId = options?.requestId;
  }
}

export interface FormaApiEnvelope<T> {
  code: number;
  msg: string;
  request_id: string;
  data: T;
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

export interface FormaApiClientOptions {
  baseUrl?: string;
  fetchImpl?: typeof fetch;
  getRequestId?: () => string;
}

export class FormaApiClient {
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;
  private readonly getRequestId: () => string;

  constructor(options: FormaApiClientOptions = {}) {
    this.baseUrl = options.baseUrl ?? '';
    this.fetchImpl = options.fetchImpl ?? fetch.bind(globalThis);
    this.getRequestId =
      options.getRequestId ??
      (() => `forma-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`);
  }

  async health(): Promise<FormaApiEnvelope<FormaHealthData>> {
    return this.get<FormaHealthData>('/api/forma/v1/health');
  }

  async version(): Promise<FormaApiEnvelope<FormaVersionData>> {
    return this.get<FormaVersionData>('/api/forma/v1/version');
  }

  async baseline(): Promise<FormaApiEnvelope<FormaBaselineData>> {
    return this.get<FormaBaselineData>('/api/forma/v1/meta/baseline');
  }

  private async get<T>(path: string): Promise<FormaApiEnvelope<T>> {
    const requestId = this.getRequestId();
    let response: Response;
    try {
      response = await this.fetchImpl(`${this.baseUrl}${path}`, {
        method: 'GET',
        headers: {
          Accept: 'application/json',
          'X-Request-ID': requestId,
        },
        credentials: 'include',
      });
    } catch (error) {
      throw new FormaApiError(
        'NETWORK_ERROR',
        error instanceof Error ? error.message : 'Network request failed',
        { requestId },
      );
    }

    if (response.status === 401) {
      throw new FormaApiError('UNAUTHORIZED', 'Unauthorized', {
        status: 401,
        requestId,
      });
    }

    if (!response.ok) {
      throw new FormaApiError('HTTP_ERROR', `HTTP ${response.status}`, {
        status: response.status,
        requestId,
      });
    }

    try {
      const body = (await response.json()) as FormaApiEnvelope<T>;
      return { ...body, request_id: body.request_id || requestId };
    } catch {
      throw new FormaApiError('PARSE_ERROR', 'Invalid JSON response', {
        status: response.status,
        requestId,
      });
    }
  }
}

export function createFormaApiClient(options?: FormaApiClientOptions) {
  return new FormaApiClient(options);
}
