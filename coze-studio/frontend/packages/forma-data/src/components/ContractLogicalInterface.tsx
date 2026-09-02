import type { FormaDataContractDescriptor } from '@forma/api-client';

import { StatusBadge } from './StatusBadge';

export interface ContractLogicalInterfaceProps {
  descriptor: FormaDataContractDescriptor;
}

/**
 * Consumer-facing logical interface. Accepts ONLY FormaDataContractDescriptor.
 * Must not read binding_refs / source_id / connection_id / asset_id.
 */
export function ContractLogicalInterface({ descriptor }: ContractLogicalInterfaceProps) {
  return (
    <div className="forma-panel" data-testid="contract-logical-interface">
      <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 8 }}>
        <strong>逻辑接口</strong>
        <StatusBadge status={descriptor.status} />
        <span className="forma-muted">v{descriptor.version}</span>
      </div>
      <p className="forma-muted">
        修订 {descriptor.revision_id} · 业务模型修订 {descriptor.business_model_revision}
      </p>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
        <thead>
          <tr>
            <th align="left">逻辑键</th>
            <th align="left">语义名</th>
            <th align="left">类型</th>
            <th align="left">分级</th>
          </tr>
        </thead>
        <tbody>
          {(descriptor.logical_schema?.fields ?? []).map(f => (
            <tr key={f.logical_key}>
              <td>{f.logical_key}</td>
              <td>{f.semantic_name}</td>
              <td>{f.logical_type}</td>
              <td>{f.classification}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="forma-muted" style={{ marginTop: 8 }}>
        查询能力：{(descriptor.query_capabilities ?? []).join(', ') || '无'}
      </div>
    </div>
  );
}
