import { useCallback, useEffect, useState } from 'react';

import type {
  FormaBusiness,
  FormaDataRequirement,
  FormaPhysicalField,
  FormaPhysicalSchema,
  FormaSemanticMapping,
} from '@forma/api-client';

import { EmptyState } from '../components/EmptyState';
import { StatusBadge } from '../components/StatusBadge';
import { confidenceDisclaimer } from '../utils/labels';
import { isEditor } from '../utils/roles';
import { useDataPlaneContext } from './useDataPlaneContext';

const MAPPING_TYPES = ['DIRECT', 'CAST', 'ENUM_MAP', 'UNIT_CONVERT', 'TIMEZONE', 'JSON_PATH', 'JOIN_REF'] as const;

function buildTransformSpec(
  mappingType: string,
  form: Record<string, string>,
): Record<string, unknown> {
  switch (mappingType) {
    case 'CAST':
      return { type: 'CAST', from_type: form.from_type || 'string', to_type: form.to_type || 'string' };
    case 'ENUM_MAP':
      return { type: 'ENUM_MAP', pairs: form.pairs ? JSON.parse(form.pairs) : [] };
    case 'UNIT_CONVERT':
      return {
        type: 'UNIT_CONVERT',
        from_unit: form.from_unit || '',
        to_unit: form.to_unit || '',
        factor: Number(form.factor || 1),
        offset: Number(form.offset || 0),
      };
    case 'TIMEZONE':
      return {
        type: 'TIMEZONE',
        source_timezone: form.source_timezone || 'UTC',
        target_timezone: form.target_timezone || 'UTC',
        format: form.format || '',
      };
    case 'JSON_PATH':
      return { type: 'JSON_PATH', path: form.path || '' };
    case 'JOIN_REF':
      return {
        type: 'JOIN_REF',
        relationship: form.relationship || '',
        from_fields: (form.from_fields || '').split(',').map(s => s.trim()).filter(Boolean),
        to_schema: form.to_schema || '',
        to_fields: (form.to_fields || '').split(',').map(s => s.trim()).filter(Boolean),
      };
    default:
      return { type: 'DIRECT' };
  }
}

export function MappingStudioPage() {
  const { client, currentTenant, businessId, businesses } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const business: FormaBusiness | undefined = businesses.find(b => b.business_id === businessId);

  const [requirements, setRequirements] = useState<FormaDataRequirement[]>([]);
  const [mappings, setMappings] = useState<FormaSemanticMapping[]>([]);
  const [selectedReqId, setSelectedReqId] = useState('');
  const [selectedMapping, setSelectedMapping] = useState<FormaSemanticMapping | null>(null);
  const [fields, setFields] = useState<FormaPhysicalField[]>([]);
  const [snapshotIds, setSnapshotIds] = useState('');
  const [mappingType, setMappingType] = useState<string>('DIRECT');
  const [targetPath, setTargetPath] = useState('');
  const [dslForm, setDslForm] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    if (!businessId) return;
    setLoading(true);
    setError(null);
    try {
      const [reqs, maps] = await Promise.all([
        client.listDataRequirements(businessId, { status: 'CONFIRMED' }),
        client.listSemanticMappings(businessId),
      ]);
      setRequirements(reqs.data ?? []);
      setMappings(maps.data ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [client, businessId]);

  useEffect(() => {
    void load();
  }, [load]);

  const loadSchemaFields = async (snapshotId: string) => {
    if (!snapshotId) return;
    try {
      const resp = await client.getSchemaSnapshot(snapshotId);
      const schema = resp.data.schema as FormaPhysicalSchema;
      setFields(schema?.fields ?? []);
    } catch {
      setFields([]);
    }
  };

  useEffect(() => {
    if (selectedMapping?.schema_snapshot_id) {
      void loadSchemaFields(selectedMapping.schema_snapshot_id);
      setMappingType(selectedMapping.mapping_type);
      setTargetPath(selectedMapping.target_field_paths?.[0] ?? '');
    }
  }, [selectedMapping?.mapping_id]);

  const analyze = async () => {
    if (!business) return;
    const ids = snapshotIds
      .split(',')
      .map(s => s.trim())
      .filter(Boolean);
    await client.analyzeSemanticMappings(businessId, {
      business_model_revision: business.current_revision,
      requirement_ids: requirements.map(r => r.requirement_id),
      schema_snapshot_ids: ids,
      client_request_id: `map-${Date.now()}`,
    });
    await load();
  };

  const createManual = async () => {
    if (!business || !selectedReqId || !targetPath) return;
    const firstSnap = snapshotIds.split(',')[0]?.trim();
    if (!firstSnap) {
      setError('请填写 schema_snapshot_ids');
      return;
    }
    await client.createManualSemanticMapping(businessId, {
      business_model_revision: business.current_revision,
      requirement_id: selectedReqId,
      source_id: 'src_manual',
      connection_id: 'conn_manual',
      asset_id: 'asset_manual',
      schema_snapshot_id: firstSnap,
      target_field_paths: [targetPath],
      mapping_type: mappingType,
      transform_spec: buildTransformSpec(mappingType, dslForm),
      confidence: 1,
      reason: 'manual',
    });
    await load();
  };

  if (!businessId) {
    return <EmptyState title="请选择业务资产" hint="映射工作室连接已确认需求与物理字段。" />;
  }

  const reqMappings = mappings.filter(m => !selectedReqId || m.requirement_id === selectedReqId);

  return (
    <div data-testid="mapping-studio">
      <div className="forma-data-toolbar" style={{ marginBottom: 12 }}>
        <h2 style={{ margin: 0 }}>映射工作室</h2>
        {canEdit ? (
          <>
            <input
              placeholder="schema_snapshot_ids（逗号分隔）"
              value={snapshotIds}
              onChange={e => setSnapshotIds(e.target.value)}
              style={{ minWidth: 240 }}
            />
            <button className="forma-btn forma-btn-primary" type="button" onClick={() => void analyze()}>
              分析映射
            </button>
          </>
        ) : null}
      </div>
      {error ? <div className="forma-error">{error}</div> : null}
      {loading ? <div className="forma-muted">加载中…</div> : null}
      <div className="forma-mapping-studio">
        <div className="forma-mapping-col" data-testid="mapping-left">
          <h3>已确认需求</h3>
          {requirements.length === 0 ? (
            <EmptyState title="无已确认需求" />
          ) : (
            requirements.map(r => (
              <button
                key={r.requirement_id}
                type="button"
                className="forma-card"
                style={{
                  textAlign: 'left',
                  cursor: 'pointer',
                  outline: selectedReqId === r.requirement_id ? '2px solid #1a56db' : undefined,
                }}
                onClick={() => setSelectedReqId(r.requirement_id)}
              >
                {r.semantic_name}
              </button>
            ))
          )}
        </div>
        <div className="forma-mapping-col" data-testid="mapping-center">
          <h3>映射详情</h3>
          {reqMappings.map(m => (
            <div
              className="forma-card"
              key={m.mapping_id}
              data-testid="mapping-card"
              onClick={() => setSelectedMapping(m)}
              onKeyDown={() => setSelectedMapping(m)}
              role="button"
              tabIndex={0}
            >
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <StatusBadge status={m.status} />
                <span>{m.mapping_type}</span>
              </div>
              <div>
                置信度：{(m.confidence * 100).toFixed(0)}%
                <div className="forma-disclaimer" data-testid="confidence-disclaimer">
                  {confidenceDisclaimer()}
                </div>
              </div>
              {canEdit && m.status === 'PROPOSED' ? (
                <div className="forma-card-actions">
                  <button
                    className="forma-btn forma-btn-primary"
                    type="button"
                    onClick={e => {
                      e.stopPropagation();
                      void client.confirmSemanticMapping(businessId, m.mapping_id).then(load);
                    }}
                  >
                    确认
                  </button>
                  <button
                    className="forma-btn forma-btn-danger"
                    type="button"
                    onClick={e => {
                      e.stopPropagation();
                      void client.rejectSemanticMapping(businessId, m.mapping_id).then(load);
                    }}
                  >
                    拒绝
                  </button>
                </div>
              ) : null}
            </div>
          ))}
          {canEdit ? (
            <div className="forma-panel">
              <div className="forma-form-row">
                <label>映射类型</label>
                <select value={mappingType} onChange={e => setMappingType(e.target.value)}>
                  {MAPPING_TYPES.map(t => (
                    <option key={t} value={t}>
                      {t}
                    </option>
                  ))}
                </select>
              </div>
              <div className="forma-form-row">
                <label>目标字段路径</label>
                <input value={targetPath} onChange={e => setTargetPath(e.target.value)} />
              </div>
              {mappingType === 'CAST' ? (
                <>
                  <div className="forma-form-row">
                    <label>from_type</label>
                    <input
                      value={dslForm.from_type ?? ''}
                      onChange={e => setDslForm(f => ({ ...f, from_type: e.target.value }))}
                    />
                  </div>
                  <div className="forma-form-row">
                    <label>to_type</label>
                    <input
                      value={dslForm.to_type ?? ''}
                      onChange={e => setDslForm(f => ({ ...f, to_type: e.target.value }))}
                    />
                  </div>
                </>
              ) : null}
              {mappingType === 'JSON_PATH' ? (
                <div className="forma-form-row">
                  <label>path</label>
                  <input
                    value={dslForm.path ?? ''}
                    onChange={e => setDslForm(f => ({ ...f, path: e.target.value }))}
                  />
                </div>
              ) : null}
              <button className="forma-btn forma-btn-primary" type="button" onClick={() => void createManual()}>
                手动创建映射
              </button>
            </div>
          ) : null}
        </div>
        <div className="forma-mapping-col" data-testid="mapping-right">
          <h3>Schema 字段</h3>
          {fields.length === 0 ? (
            <EmptyState title="无字段" hint="选择映射或加载快照以查看物理字段。" />
          ) : (
            fields.map(f => (
              <button
                key={f.path || f.name}
                type="button"
                className="forma-card"
                onClick={() => setTargetPath(f.path || f.name)}
              >
                {f.path || f.name} · {f.data_type}
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
