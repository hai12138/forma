import type { FormaContractBinding } from '@forma/api-client';

export interface ContractBindingDetailProps {
  bindings: FormaContractBinding[];
}

/** Admin-only physical bindings (separate from logical interface). */
export function ContractBindingDetail({ bindings }: ContractBindingDetailProps) {
  if (!bindings.length) {
    return (
      <div className="forma-empty-state" data-testid="binding-empty">
        暂无物理绑定
      </div>
    );
  }

  return (
    <div className="forma-panel" data-testid="contract-binding-detail">
      <strong>物理绑定（管理）</strong>
      <ul>
        {bindings.map(b => (
          <li key={`${b.mapping_id}-${b.requirement_id}`} className="forma-card" style={{ marginTop: 8 }}>
            <div>需求：{b.requirement_id}</div>
            <div>映射：{b.mapping_id}</div>
            <div>源：{b.source_id}</div>
            <div>连接：{b.connection_id}</div>
            <div>资产：{b.asset_id}</div>
            <div>快照：{b.schema_snapshot_id}</div>
          </li>
        ))}
      </ul>
    </div>
  );
}
