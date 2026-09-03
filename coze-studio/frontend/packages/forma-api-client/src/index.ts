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

export interface FormaAnalystSession {
  session_id: string;
  business_id: string;
  status: string;
  title: string;
  confirmation_policy: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface FormaAnalystTurn {
  turn_id: string;
  sequence: number;
  speaker: string;
  content: string;
  content_type: string;
  analysis_status: string;
  client_request_id?: string;
  created_at: string;
}

export interface FormaEvidence {
  evidence_id: string;
  session_id: string;
  turn_id: string;
  source_type: string;
  quote: string;
  created_by: string;
  created_at: string;
}

export interface FormaAssertion {
  assertion_id: string;
  session_id: string;
  assertion_type: string;
  subject_ref: string;
  predicate: string;
  object_value: string;
  confidence: number;
  status: string;
  source_marker: FormaSourceMarker;
  evidence_ids: string[];
  structured_value?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface FormaAssertionEdit {
  assertion_type: string;
  subject_ref: string;
  predicate: string;
  object_value: string;
}

export interface FormaConflict {
  conflict_id: string;
  session_id: string;
  assertion_id_a: string;
  assertion_id_b: string;
  subject_ref: string;
  predicate: string;
  status: string;
}

export interface FormaGap {
  gap_id: string;
  session_id: string;
  gap_type: string;
  question: string;
  status: string;
  related_assertion_ids?: string[];
}

export interface FormaPatchOperation {
  op: string;
  target_id?: string;
  node?: FormaSemanticNode;
  edge?: FormaSemanticEdge;
  state?: FormaBusinessState;
  rule?: FormaBusinessRule;
  source_assertion_ids: string[];
}

export interface FormaSemanticModelPatch {
  operations: FormaPatchOperation[];
}

export interface FormaProposal {
  proposal_id: string;
  business_id: string;
  session_id: string;
  base_revision: number;
  assertion_ids: string[];
  patch: FormaSemanticModelPatch;
  status: string;
  content_digest: string;
  created_at: string;
}

export interface FormaTurnSubmission {
  user_turn: FormaAnalystTurn;
  analyst_turn?: FormaAnalystTurn;
  evidence?: FormaEvidence;
  assertions?: FormaAssertion[];
  conflicts?: FormaConflict[];
  gaps?: FormaGap[];
  model_failed?: boolean;
  model_error?: string;
  next_question?: { question: string; goal: string; priority: number };
}

export interface FormaApplyProposalResult {
  revision_no: number;
  proposal_id: string;
}

export interface FormaProposalPreview {
  proposal: FormaProposal;
  current_revision: number;
  validation_valid: boolean;
  validation_error?: string;
  assertion_count: number;
  proposed_model?: FormaSemanticModel;
  diff?: FormaDiffResponse['diff'];
  impact?: FormaDiffResponse['impact'];
}

/* ---- S4 Data Plane DTOs (snake_case matching backend app DTOs) ---- */

export interface FormaDataRequirement {
  requirement_id: string;
  business_id: string;
  business_model_revision: number;
  requirement_kind: string;
  semantic_name: string;
  description: string;
  business_element_refs: string[];
  requiredness: string;
  freshness_requirement: string;
  access_need: string;
  status: string;
  source: string;
  derived_from_requirement_id?: string;
  analysis_run_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface FormaDataRequirementDecision {
  decision_id: string;
  business_id: string;
  source_requirement_id: string;
  target_requirement_id?: string;
  action: string;
  actor_principal_id: string;
  reason: string;
  business_model_revision: number;
  created_at: string;
}

export interface FormaDataAnalysisRun {
  analysis_run_id: string;
  business_id: string;
  business_model_revision: number;
  client_request_id: string;
  status: string;
  model_ref?: string;
  error_key?: string;
  error_message_sanitized?: string;
  retry_count: number;
  last_retry_by?: string;
  last_retry_at?: string;
  created_by: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface FormaAnalyzeDataRequirementsResponse {
  analysis_run: FormaDataAnalysisRun;
  requirements: FormaDataRequirement[];
  owned_execute: boolean;
}

export interface FormaCreateManualDataRequirementInput {
  business_model_revision: number;
  requirement_kind: string;
  semantic_name: string;
  description: string;
  business_element_refs?: string[];
  requiredness: string;
  freshness_requirement: string;
  access_need: string;
}

export interface FormaEditConfirmDataRequirementInput {
  reason?: string;
  requirement_kind: string;
  semantic_name: string;
  description: string;
  business_element_refs?: string[];
  requiredness: string;
  freshness_requirement: string;
  access_need: string;
}

export interface FormaDataSource {
  source_id: string;
  name: string;
  source_type: string;
  status: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface FormaDataConnection {
  connection_id: string;
  source_id: string;
  name: string;
  environment: string;
  adapter_type: string;
  public_config: Record<string, unknown> | string;
  credential_ref_id?: string;
  status: string;
  last_test_status?: string;
  last_test_at?: string;
  last_test_error_key?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface FormaDataAsset {
  asset_id: string;
  source_id: string;
  connection_id: string;
  asset_type: string;
  name: string;
  physical_locator: Record<string, unknown> | string;
  locator_digest: string;
  created_at: string;
  updated_at: string;
}

export interface FormaPhysicalField {
  name: string;
  data_type: string;
  native_type?: string;
  nullable: boolean;
  primary_key: boolean;
  description?: string;
  path?: string;
  ordinal: number;
  metadata?: Record<string, unknown>;
}

export interface FormaPhysicalRelationship {
  name: string;
  from_fields: string[];
  to_schema: string;
  to_fields: string[];
  relationship_type: string;
}

export interface FormaPhysicalSchema {
  name: string;
  fields: FormaPhysicalField[];
  relationships: FormaPhysicalRelationship[];
}

export interface FormaSchemaSnapshot {
  snapshot_id: string;
  source_id: string;
  connection_id: string;
  asset_id: string;
  schema: FormaPhysicalSchema | Record<string, unknown>;
  fingerprint: string;
  created_by: string;
  created_at: string;
}

/** Credential response — MUST NOT include password|secret|token|api_key|authorization. */
export interface FormaCredentialRef {
  credential_ref_id: string;
  status: string;
  provider: string;
  created_at: string;
  rotated_at?: string;
  last_rotated_at?: string;
}

/** Create body may carry ephemeral secret fields; never echoed in responses. */
export interface FormaCreateCredentialInput {
  secret_type: string;
  secret: Record<string, unknown> | string;
}

export interface FormaSemanticMapping {
  mapping_id: string;
  business_id: string;
  business_model_revision: number;
  requirement_id: string;
  source_id: string;
  connection_id: string;
  asset_id: string;
  schema_snapshot_id: string;
  target_field_paths: string[];
  mapping_type: string;
  transform_spec: Record<string, unknown> | string;
  status: string;
  source: string;
  confidence: number;
  reason: string;
  derived_from_mapping_id?: string;
  analysis_run_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface FormaMappingCoverage {
  total_confirmed_requirements: number;
  confirmed_mappings: number;
  unmapped_requirement_ids: string[];
  coverage: number;
}

export interface FormaMappingAnalysisRun {
  analysis_run_id: string;
  business_id: string;
  business_model_revision: number;
  client_request_id: string;
  status: string;
  model_ref?: string;
  error_key?: string;
  error_message_sanitized?: string;
  retry_count: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface FormaAnalyzeSemanticMappingsResponse {
  analysis_run: FormaMappingAnalysisRun;
  mappings: FormaSemanticMapping[];
  owned_execute: boolean;
}

export interface FormaCreateSemanticMappingInput {
  business_model_revision: number;
  requirement_id: string;
  source_id: string;
  connection_id: string;
  asset_id: string;
  schema_snapshot_id: string;
  target_field_paths: string[];
  mapping_type: string;
  transform_spec?: Record<string, unknown>;
  confidence?: number;
  reason?: string;
}

export interface FormaEditConfirmSemanticMappingInput {
  source_id: string;
  connection_id: string;
  asset_id: string;
  schema_snapshot_id: string;
  target_field_paths: string[];
  mapping_type: string;
  transform_spec?: Record<string, unknown>;
  confidence?: number;
  reason?: string;
}

export interface FormaLogicalField {
  logical_key: string;
  semantic_name: string;
  logical_type: string;
  description: string;
  requirement_id: string;
  nullable: boolean;
  classification: string;
}

export interface FormaContractLogicalSchema {
  fields: FormaLogicalField[];
}

export interface FormaContractBinding {
  requirement_id: string;
  mapping_id: string;
  source_id: string;
  connection_id: string;
  asset_id: string;
  schema_snapshot_id: string;
}

export interface FormaDataContract {
  contract_id: string;
  business_id: string;
  active_revision_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface FormaDataContractRevision {
  revision_id: string;
  business_id: string;
  contract_id: string;
  version: number;
  status: string;
  business_model_revision: number;
  name: string;
  description: string;
  requirement_ids: string[];
  logical_schema: FormaContractLogicalSchema;
  query_capabilities: string[];
  filter_schema: { fields: Array<{ logical_key: string; operators: string[] }> };
  sort_schema: { fields: Array<{ logical_key: string; directions: string[] }> };
  pagination_policy: { default_limit: number; max_limit: number };
  freshness_policy: string;
  classification_policy: Record<string, string>;
  /** Omitted for MEMBER/VIEWER via member-safe revision projection. */
  binding_refs?: FormaContractBinding[];
  access_policy_ref?: string;
  derived_from_revision_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

/** Consumer-facing — no binding_refs / physical source fields. */
export interface FormaDataContractDescriptor {
  contract_id: string;
  revision_id: string;
  version: number;
  business_model_revision: number;
  logical_schema: FormaContractLogicalSchema;
  query_capabilities: string[];
  filter_schema: { fields: Array<{ logical_key: string; operators: string[] }> };
  sort_schema: { fields: Array<{ logical_key: string; directions: string[] }> };
  pagination_policy: { default_limit: number; max_limit: number };
  freshness_policy: string;
  classification: Record<string, string>;
  access_policy_ref?: string;
  status: string;
}

export interface FormaCreateDataContractInput {
  business_model_revision: number;
  name: string;
  description?: string;
  requirement_ids: string[];
  logical_schema: FormaContractLogicalSchema;
  query_capabilities?: string[];
  filter_schema?: FormaDataContractRevision['filter_schema'];
  sort_schema?: FormaDataContractRevision['sort_schema'];
  pagination_policy?: FormaDataContractRevision['pagination_policy'];
  freshness_policy?: string;
  classification_policy?: Record<string, string>;
  mapping_ids: string[];
  access_policy_ref?: string;
}

export interface FormaCreateDataContractRevisionInput extends FormaCreateDataContractInput {
  base_revision_id: string;
}

export interface FormaCreateDataContractResponse {
  contract: FormaDataContract;
  revision: FormaDataContractRevision;
}

/** Wire shape from domain entity without json tags → Go PascalCase. */
export interface FormaValidationResult {
  ValidationID: string;
  TenantID: string;
  BusinessID: string;
  ContractID: string;
  RevisionID: string;
  Version: number;
  Status: string;
  Errors: Array<{ code: string; message: string }>;
  Warnings: Array<{ code: string; message: string }>;
  SnapshotFingerprints: Record<string, string>;
  ValidatedBy: string;
  ValidatedAt: string;
  CreatedAt: string;
}

export interface FormaDriftResult {
  DriftResultID: string;
  TenantID: string;
  BusinessID: string;
  ContractID: string;
  RevisionID: string;
  Version: number;
  Severity: string;
  Findings: Array<{
    code: string;
    message: string;
    binding_mapping_id: string;
    field_path: string;
  }>;
  ComparedSnapshotIDs: Record<string, string>;
  EvaluatedBy: string;
  EvaluatedAt: string;
  CreatedAt: string;
}

export interface FormaGapResult {
  GapResultID: string;
  TenantID: string;
  BusinessID: string;
  ContractID: string;
  RevisionID: string;
  Version: number;
  FromBusinessRevision: number;
  CurrentBusinessRevision: number;
  NewConfirmedRequirementIDs: string[];
  UnmappedRequirementIDs: string[];
  GapStatus: string;
  EvaluatedBy: string;
  EvaluatedAt: string;
  CreatedAt: string;
}

export interface FormaLifecycleEvent {
  EventID: string;
  TenantID: string;
  BusinessID: string;
  ContractID: string;
  RevisionID: string;
  Version: number;
  Action: string;
  ActorPrincipalID: string;
  Reason: string;
  CreatedAt: string;
}

export interface FormaApiClientOptions {
  baseUrl?: string;
  fetchImpl?: typeof fetch;
  getRequestId?: () => string;
  getTenantId?: () => string | undefined;
  /** Fired once per 401 — Forma AuthGuard uses this for session expiry UX. */
  onUnauthorized?: () => void;
}

export class FormaApiClient {
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;
  private readonly getRequestId: () => string;
  private readonly getTenantId?: () => string | undefined;
  private readonly onUnauthorized?: () => void;

  constructor(options: FormaApiClientOptions = {}) {
    this.baseUrl = options.baseUrl ?? '';
    this.fetchImpl = options.fetchImpl ?? fetch.bind(globalThis);
    this.getRequestId =
      options.getRequestId ??
      (() => `forma-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`);
    this.getTenantId = options.getTenantId;
    this.onUnauthorized = options.onUnauthorized;
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

  async createAnalystSession(
    businessId: string,
    body: { title?: string; confirmation_policy?: string },
  ): Promise<FormaApiEnvelope<FormaAnalystSession>> {
    return this.request<FormaAnalystSession>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/analyst/sessions`,
      body,
    );
  }

  async listAnalystSessions(businessId: string): Promise<FormaApiEnvelope<FormaAnalystSession[]>> {
    return this.request<FormaAnalystSession[]>(
      'GET',
      `/api/forma/v1/businesses/${businessId}/analyst/sessions`,
    );
  }

  async getAnalystSession(
    businessId: string,
    sessionId: string,
  ): Promise<FormaApiEnvelope<FormaAnalystSession>> {
    return this.request<FormaAnalystSession>(
      'GET',
      `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}`,
    );
  }

  async submitAnalystTurn(
    businessId: string,
    sessionId: string,
    body: { content: string; client_request_id?: string },
  ): Promise<FormaApiEnvelope<FormaTurnSubmission>> {
    return this.request<FormaTurnSubmission>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}/turns`,
      body,
    );
  }

  async listAnalystTurns(
    businessId: string,
    sessionId: string,
  ): Promise<FormaApiEnvelope<FormaAnalystTurn[]>> {
    return this.request<FormaAnalystTurn[]>(
      'GET',
      `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}/turns`,
    );
  }

  async askAnalystGap(
    businessId: string,
    sessionId: string,
    gapId: string,
  ): Promise<FormaApiEnvelope<{ analyst_turn: FormaAnalystTurn; gap: FormaGap }>> {
    return this.request<{ analyst_turn: FormaAnalystTurn; gap: FormaGap }>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}/gaps/${gapId}/ask`,
      {},
    );
  }

  async listAssertions(businessId: string): Promise<FormaApiEnvelope<FormaAssertion[]>> {
    return this.request<FormaAssertion[]>('GET', `/api/forma/v1/businesses/${businessId}/assertions`);
  }

  async listEvidence(businessId: string): Promise<FormaApiEnvelope<FormaEvidence[]>> {
    return this.request<FormaEvidence[]>('GET', `/api/forma/v1/businesses/${businessId}/evidence`);
  }

  async confirmAssertion(
    businessId: string,
    assertionId: string,
    body?: { comment?: string; edit?: FormaAssertionEdit },
  ): Promise<FormaApiEnvelope<FormaAssertion>> {
    return this.request<FormaAssertion>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/assertions/${assertionId}/confirm`,
      body ?? {},
    );
  }

  async rejectAssertion(
    businessId: string,
    assertionId: string,
    body?: { comment?: string },
  ): Promise<FormaApiEnvelope<FormaAssertion>> {
    return this.request<FormaAssertion>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/assertions/${assertionId}/reject`,
      body ?? {},
    );
  }

  async listConflicts(businessId: string): Promise<FormaApiEnvelope<FormaConflict[]>> {
    return this.request<FormaConflict[]>('GET', `/api/forma/v1/businesses/${businessId}/conflicts`);
  }

  async listGaps(businessId: string): Promise<FormaApiEnvelope<FormaGap[]>> {
    return this.request<FormaGap[]>('GET', `/api/forma/v1/businesses/${businessId}/gaps`);
  }

  async createProposal(
    businessId: string,
    body: { session_id?: string; assertion_ids?: string[] },
  ): Promise<FormaApiEnvelope<FormaProposal>> {
    return this.request<FormaProposal>('POST', `/api/forma/v1/businesses/${businessId}/proposals`, body);
  }

  async getProposal(
    businessId: string,
    proposalId: string,
  ): Promise<FormaApiEnvelope<FormaProposal>> {
    return this.request<FormaProposal>(
      'GET',
      `/api/forma/v1/businesses/${businessId}/proposals/${proposalId}`,
    );
  }

  async applyProposal(
    businessId: string,
    proposalId: string,
  ): Promise<FormaApiEnvelope<FormaApplyProposalResult>> {
    return this.request<FormaApplyProposalResult>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/proposals/${proposalId}/apply`,
    );
  }

  async retryAnalystTurnAnalysis(
    businessId: string,
    sessionId: string,
    turnId: string,
  ): Promise<FormaApiEnvelope<FormaTurnSubmission>> {
    return this.request<FormaTurnSubmission>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}/turns/${turnId}/retry-analysis`,
    );
  }

  async getProposalPreview(
    businessId: string,
    proposalId: string,
  ): Promise<FormaApiEnvelope<FormaProposalPreview>> {
    return this.request<FormaProposalPreview>(
      'GET',
      `/api/forma/v1/businesses/${businessId}/proposals/${proposalId}/preview`,
    );
  }

  /* ---- S4 Data Plane ---- */

  async listDataRequirements(
    businessId: string,
    opts?: { revision?: number; business_model_revision?: number; status?: string },
  ): Promise<FormaApiEnvelope<FormaDataRequirement[]>> {
    const q = new URLSearchParams();
    const rev = opts?.business_model_revision ?? opts?.revision;
    if (rev != null) q.set('business_model_revision', String(rev));
    if (opts?.status) q.set('status', opts.status);
    const qs = q.toString();
    return this.request<FormaDataRequirement[]>(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-requirements${qs ? `?${qs}` : ''}`,
    );
  }

  async analyzeDataRequirements(
    businessId: string,
    body: { business_model_revision: number; client_request_id?: string },
  ): Promise<FormaApiEnvelope<FormaAnalyzeDataRequirementsResponse>> {
    return this.request<FormaAnalyzeDataRequirementsResponse>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-requirements/analyze`,
      body,
    );
  }

  async getDataAnalysisRun(
    businessId: string,
    analysisRunId: string,
  ): Promise<FormaApiEnvelope<FormaDataAnalysisRun>> {
    return this.request<FormaDataAnalysisRun>(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-analyses/${analysisRunId}`,
    );
  }

  async createManualDataRequirement(
    businessId: string,
    body: FormaCreateManualDataRequirementInput,
  ): Promise<FormaApiEnvelope<FormaDataRequirement>> {
    return this.request<FormaDataRequirement>(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-requirements`,
      body,
    );
  }

  async confirmDataRequirement(
    businessId: string,
    requirementId: string,
    body?: { reason?: string },
  ): Promise<FormaApiEnvelope<{ requirement: FormaDataRequirement; decision: FormaDataRequirementDecision }>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-requirements/${requirementId}/confirm`,
      body ?? {},
    );
  }

  async rejectDataRequirement(
    businessId: string,
    requirementId: string,
    body?: { reason?: string },
  ): Promise<FormaApiEnvelope<{ requirement: FormaDataRequirement; decision: FormaDataRequirementDecision }>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-requirements/${requirementId}/reject`,
      body ?? {},
    );
  }

  async editConfirmDataRequirement(
    businessId: string,
    requirementId: string,
    body: FormaEditConfirmDataRequirementInput,
  ): Promise<
    FormaApiEnvelope<{
      original: FormaDataRequirement;
      replacement: FormaDataRequirement;
      decision: FormaDataRequirementDecision;
    }>
  > {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-requirements/${requirementId}/edit-confirm`,
      body,
    );
  }

  async listDataRequirementDecisions(
    businessId: string,
    requirementId: string,
  ): Promise<FormaApiEnvelope<FormaDataRequirementDecision[]>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-requirements/${requirementId}/decisions`,
    );
  }

  async analyzeSemanticMappings(
    businessId: string,
    body: {
      business_model_revision: number;
      requirement_ids: string[];
      schema_snapshot_ids: string[];
      client_request_id?: string;
    },
  ): Promise<FormaApiEnvelope<FormaAnalyzeSemanticMappingsResponse>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/semantic-mappings/analyze`,
      body,
    );
  }

  async getMappingAnalysisRun(
    businessId: string,
    analysisRunId: string,
  ): Promise<FormaApiEnvelope<FormaMappingAnalysisRun>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/semantic-mapping-analysis/${analysisRunId}`,
    );
  }

  async listSemanticMappings(
    businessId: string,
    opts?: { business_model_revision?: number; status?: string },
  ): Promise<FormaApiEnvelope<FormaSemanticMapping[]>> {
    const q = new URLSearchParams();
    if (opts?.business_model_revision != null) {
      q.set('business_model_revision', String(opts.business_model_revision));
    }
    if (opts?.status) q.set('status', opts.status);
    const qs = q.toString();
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/semantic-mappings${qs ? `?${qs}` : ''}`,
    );
  }

  async createManualSemanticMapping(
    businessId: string,
    body: FormaCreateSemanticMappingInput,
  ): Promise<FormaApiEnvelope<FormaSemanticMapping>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/semantic-mappings`,
      body,
    );
  }

  async confirmSemanticMapping(
    businessId: string,
    mappingId: string,
    body?: { reason?: string },
  ): Promise<FormaApiEnvelope<{ mapping: FormaSemanticMapping; decision: unknown }>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/semantic-mappings/${mappingId}/confirm`,
      body ?? {},
    );
  }

  async rejectSemanticMapping(
    businessId: string,
    mappingId: string,
    body?: { reason?: string },
  ): Promise<FormaApiEnvelope<{ mapping: FormaSemanticMapping; decision: unknown }>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/semantic-mappings/${mappingId}/reject`,
      body ?? {},
    );
  }

  async editConfirmSemanticMapping(
    businessId: string,
    mappingId: string,
    body: FormaEditConfirmSemanticMappingInput,
  ): Promise<
    FormaApiEnvelope<{
      original: FormaSemanticMapping;
      replacement: FormaSemanticMapping;
      decision: unknown;
    }>
  > {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/semantic-mappings/${mappingId}/edit-confirm`,
      body,
    );
  }

  async getSemanticMappingCoverage(
    businessId: string,
  ): Promise<FormaApiEnvelope<FormaMappingCoverage>> {
    return this.request('GET', `/api/forma/v1/businesses/${businessId}/semantic-mapping-coverage`);
  }

  async listDataContracts(businessId: string): Promise<FormaApiEnvelope<FormaDataContract[]>> {
    return this.request('GET', `/api/forma/v1/businesses/${businessId}/data-contracts`);
  }

  async createDataContract(
    businessId: string,
    body: FormaCreateDataContractInput,
  ): Promise<FormaApiEnvelope<FormaCreateDataContractResponse>> {
    return this.request('POST', `/api/forma/v1/businesses/${businessId}/data-contracts`, body);
  }

  async getDataContract(
    businessId: string,
    contractId: string,
  ): Promise<FormaApiEnvelope<FormaDataContract>> {
    return this.request('GET', `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}`);
  }

  async getActiveDataContractDescriptor(
    businessId: string,
    contractId: string,
  ): Promise<FormaApiEnvelope<FormaDataContractDescriptor>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/active-descriptor`,
    );
  }

  async listDataContractRevisions(
    businessId: string,
    contractId: string,
  ): Promise<FormaApiEnvelope<FormaDataContractRevision[]>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions`,
    );
  }

  async createDataContractRevision(
    businessId: string,
    contractId: string,
    body: FormaCreateDataContractRevisionInput,
  ): Promise<FormaApiEnvelope<FormaDataContractRevision>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions`,
      body,
    );
  }

  async getDataContractRevision(
    businessId: string,
    contractId: string,
    revisionId: string,
  ): Promise<FormaApiEnvelope<FormaDataContractRevision>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}`,
    );
  }

  async validateDataContractRevision(
    businessId: string,
    contractId: string,
    revisionId: string,
  ): Promise<FormaApiEnvelope<{ revision: FormaDataContractRevision; result: FormaValidationResult }>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/validate`,
      {},
    );
  }

  async activateDataContractRevision(
    businessId: string,
    contractId: string,
    revisionId: string,
    body?: { reason?: string },
  ): Promise<FormaApiEnvelope<FormaDataContractRevision>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/activate`,
      body ?? {},
    );
  }

  async deprecateDataContractRevision(
    businessId: string,
    contractId: string,
    revisionId: string,
    body?: { reason?: string },
  ): Promise<FormaApiEnvelope<FormaDataContractRevision>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/deprecate`,
      body ?? {},
    );
  }

  async listDataContractValidationResults(
    businessId: string,
    contractId: string,
    revisionId: string,
  ): Promise<FormaApiEnvelope<FormaValidationResult[]>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/validation-results`,
    );
  }

  async evaluateDataContractDrift(
    businessId: string,
    contractId: string,
    revisionId: string,
    body: { new_snapshot_ids: Record<string, string> },
  ): Promise<FormaApiEnvelope<{ result: FormaDriftResult; revision: FormaDataContractRevision }>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/evaluate-drift`,
      body,
    );
  }

  async listDataContractDriftResults(
    businessId: string,
    contractId: string,
    revisionId: string,
  ): Promise<FormaApiEnvelope<FormaDriftResult[]>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/drift-results`,
    );
  }

  async evaluateDataContractGap(
    businessId: string,
    contractId: string,
    revisionId: string,
  ): Promise<FormaApiEnvelope<FormaGapResult>> {
    return this.request(
      'POST',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/evaluate-gap`,
      {},
    );
  }

  async listDataContractGapResults(
    businessId: string,
    contractId: string,
    revisionId: string,
  ): Promise<FormaApiEnvelope<FormaGapResult[]>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/revisions/${revisionId}/gap-results`,
    );
  }

  async listDataContractLifecycleEvents(
    businessId: string,
    contractId: string,
  ): Promise<FormaApiEnvelope<FormaLifecycleEvent[]>> {
    return this.request(
      'GET',
      `/api/forma/v1/businesses/${businessId}/data-contracts/${contractId}/lifecycle-events`,
    );
  }

  async listDataSources(): Promise<FormaApiEnvelope<FormaDataSource[]>> {
    return this.request('GET', '/api/forma/v1/data-sources');
  }

  async createDataSource(body: {
    name: string;
    source_type: string;
  }): Promise<FormaApiEnvelope<FormaDataSource>> {
    return this.request('POST', '/api/forma/v1/data-sources', body);
  }

  async getDataSource(sourceId: string): Promise<FormaApiEnvelope<FormaDataSource>> {
    return this.request('GET', `/api/forma/v1/data-sources/${sourceId}`);
  }

  async archiveDataSource(sourceId: string): Promise<FormaApiEnvelope<FormaDataSource>> {
    return this.request('POST', `/api/forma/v1/data-sources/${sourceId}/archive`);
  }

  async listDataConnections(sourceId: string): Promise<FormaApiEnvelope<FormaDataConnection[]>> {
    return this.request('GET', `/api/forma/v1/data-sources/${sourceId}/connections`);
  }

  async createDataConnection(
    sourceId: string,
    body: {
      name: string;
      environment: string;
      adapter_type: string;
      public_config: Record<string, unknown>;
      credential_ref_id?: string;
    },
  ): Promise<FormaApiEnvelope<FormaDataConnection>> {
    return this.request('POST', `/api/forma/v1/data-sources/${sourceId}/connections`, body);
  }

  async getDataConnection(
    sourceId: string,
    connectionId: string,
  ): Promise<FormaApiEnvelope<FormaDataConnection>> {
    return this.request(
      'GET',
      `/api/forma/v1/data-sources/${sourceId}/connections/${connectionId}`,
    );
  }

  async testDataConnection(
    sourceId: string,
    connectionId: string,
  ): Promise<FormaApiEnvelope<FormaDataConnection>> {
    return this.request(
      'POST',
      `/api/forma/v1/data-sources/${sourceId}/connections/${connectionId}/test`,
      {},
    );
  }

  async discoverDataAssets(
    sourceId: string,
    connectionId: string,
  ): Promise<FormaApiEnvelope<FormaDataAsset[]>> {
    return this.request(
      'POST',
      `/api/forma/v1/data-sources/${sourceId}/connections/${connectionId}/discover`,
      {},
    );
  }

  async createDataCredential(
    body: FormaCreateCredentialInput,
  ): Promise<FormaApiEnvelope<FormaCredentialRef>> {
    return this.request('POST', '/api/forma/v1/credentials', body);
  }

  async listDataAssets(sourceId: string): Promise<FormaApiEnvelope<FormaDataAsset[]>> {
    return this.request('GET', `/api/forma/v1/data-sources/${sourceId}/assets`);
  }

  async getDataAsset(assetId: string): Promise<FormaApiEnvelope<FormaDataAsset>> {
    return this.request('GET', `/api/forma/v1/data-assets/${assetId}`);
  }

  async captureDataSchema(
    sourceId: string,
    connectionId: string,
    assetId: string,
  ): Promise<FormaApiEnvelope<FormaSchemaSnapshot>> {
    return this.request(
      'POST',
      `/api/forma/v1/data-sources/${sourceId}/connections/${connectionId}/assets/${assetId}/capture-schema`,
      {},
    );
  }

  async getSchemaSnapshot(snapshotId: string): Promise<FormaApiEnvelope<FormaSchemaSnapshot>> {
    return this.request('GET', `/api/forma/v1/schema-snapshots/${snapshotId}`);
  }

  /** Tenant-scoped list of schema snapshots for one asset (MEMBER readable). */
  async listSchemaSnapshots(params: {
    sourceId: string;
    connectionId: string;
    assetId: string;
  }): Promise<FormaApiEnvelope<FormaSchemaSnapshot[]>> {
    const q = new URLSearchParams({
      source_id: params.sourceId,
      connection_id: params.connectionId,
      asset_id: params.assetId,
    });
    return this.request('GET', `/api/forma/v1/schema-snapshots?${q.toString()}`);
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
      try {
        this.onUnauthorized?.();
      } catch {
        // never let auth side-effects break error path
      }
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
