import type {
  FormaContractBinding,
  FormaDataContractDescriptor,
  FormaDataContractRevision,
  FormaDataRequirement,
  FormaPhysicalSchema,
  FormaSemanticMapping,
} from '@forma/api-client';

/** Lab sample flow — domain-agnostic fixtures for tests only. */
export const labRequirement: FormaDataRequirement = {
  requirement_id: 'req_lab_temp',
  business_id: 'biz_lab',
  business_model_revision: 1,
  requirement_kind: 'ENTITY',
  semantic_name: 'sample_temperature',
  description: 'Lab sample temperature reading',
  business_element_refs: ['node_sample'],
  requiredness: 'REQUIRED',
  freshness_requirement: 'NEAR_REALTIME',
  access_need: 'READ',
  status: 'CONFIRMED',
  source: 'MANUAL',
  created_by: 'principal_1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

export const labProposedRequirement: FormaDataRequirement = {
  ...labRequirement,
  requirement_id: 'req_lab_proposed',
  semantic_name: 'sample_batch_id',
  status: 'PROPOSED',
  source: 'AI_GENERATED',
};

export const labSchema: FormaPhysicalSchema = {
  name: 'readings',
  fields: [
    {
      name: 'temp_c',
      data_type: 'number',
      nullable: false,
      primary_key: false,
      path: 'temp_c',
      ordinal: 0,
    },
    {
      name: 'batch_id',
      data_type: 'string',
      nullable: false,
      primary_key: true,
      path: 'batch_id',
      ordinal: 1,
    },
  ],
  relationships: [],
};

export const labMapping: FormaSemanticMapping = {
  mapping_id: 'map_lab_1',
  business_id: 'biz_lab',
  business_model_revision: 1,
  requirement_id: labRequirement.requirement_id,
  source_id: 'src_lab',
  connection_id: 'conn_lab',
  asset_id: 'asset_lab',
  schema_snapshot_id: 'snap_lab',
  target_field_paths: ['temp_c'],
  mapping_type: 'DIRECT',
  transform_spec: { type: 'DIRECT' },
  status: 'PROPOSED',
  source: 'AI_GENERATED',
  confidence: 0.91,
  reason: 'name similarity',
  created_by: 'principal_1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

/** Procurement sample — second domain-agnostic fixture set. */
export const procurementRequirement: FormaDataRequirement = {
  requirement_id: 'req_proc_po',
  business_id: 'biz_proc',
  business_model_revision: 2,
  requirement_kind: 'ENTITY',
  semantic_name: 'purchase_order_id',
  description: 'Purchase order identifier',
  business_element_refs: ['node_po'],
  requiredness: 'REQUIRED',
  freshness_requirement: 'DAILY',
  access_need: 'READ',
  status: 'CONFIRMED',
  source: 'MANUAL',
  created_by: 'principal_2',
  created_at: '2026-02-01T00:00:00Z',
  updated_at: '2026-02-01T00:00:00Z',
};

export const procurementMapping: FormaSemanticMapping = {
  mapping_id: 'map_proc_1',
  business_id: 'biz_proc',
  business_model_revision: 2,
  requirement_id: procurementRequirement.requirement_id,
  source_id: 'src_proc',
  connection_id: 'conn_proc',
  asset_id: 'asset_proc',
  schema_snapshot_id: 'snap_proc',
  target_field_paths: ['po_id'],
  mapping_type: 'CAST',
  transform_spec: { type: 'CAST', from_type: 'string', to_type: 'string' },
  status: 'CONFIRMED',
  source: 'MANUAL',
  confidence: 1,
  reason: 'manual',
  created_by: 'principal_2',
  created_at: '2026-02-01T00:00:00Z',
  updated_at: '2026-02-01T00:00:00Z',
};

export const labDescriptor: FormaDataContractDescriptor = {
  contract_id: 'ctr_lab',
  revision_id: 'rev_lab_2',
  version: 2,
  business_model_revision: 1,
  logical_schema: {
    fields: [
      {
        logical_key: 'temperature',
        semantic_name: 'sample_temperature',
        logical_type: 'number',
        description: 'Temperature',
        requirement_id: labRequirement.requirement_id,
        nullable: false,
        classification: 'INTERNAL',
      },
    ],
  },
  query_capabilities: ['FILTER', 'SORT'],
  filter_schema: { fields: [{ logical_key: 'temperature', operators: ['EQ', 'GT'] }] },
  sort_schema: { fields: [{ logical_key: 'temperature', directions: ['ASC', 'DESC'] }] },
  pagination_policy: { default_limit: 50, max_limit: 200 },
  freshness_policy: 'NEAR_REALTIME',
  classification: { temperature: 'INTERNAL' },
  status: 'ACTIVE',
};

export const labBinding: FormaContractBinding = {
  requirement_id: labRequirement.requirement_id,
  mapping_id: labMapping.mapping_id,
  source_id: 'src_lab',
  connection_id: 'conn_lab',
  asset_id: 'asset_lab',
  schema_snapshot_id: 'snap_lab',
};

export const staleRevision: FormaDataContractRevision = {
  revision_id: 'rev_lab_1',
  business_id: 'biz_lab',
  contract_id: 'ctr_lab',
  version: 1,
  status: 'STALE',
  business_model_revision: 1,
  name: 'Lab Contract v1',
  description: 'Historical stale revision',
  requirement_ids: [labRequirement.requirement_id],
  logical_schema: labDescriptor.logical_schema,
  query_capabilities: ['FILTER'],
  filter_schema: labDescriptor.filter_schema,
  sort_schema: labDescriptor.sort_schema,
  pagination_policy: labDescriptor.pagination_policy,
  freshness_policy: 'NEAR_REALTIME',
  classification_policy: { temperature: 'INTERNAL' },
  binding_refs: [labBinding],
  created_by: 'principal_1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-02T00:00:00Z',
};

export const activeRevision: FormaDataContractRevision = {
  ...staleRevision,
  revision_id: 'rev_lab_2',
  version: 2,
  status: 'ACTIVE',
  name: 'Lab Contract v2',
};
