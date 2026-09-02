import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import type {
  FormaDataContract,
  FormaDataContractDescriptor,
  FormaDataContractRevision,
} from '@forma/api-client';

import { ContractBindingDetail } from '../components/ContractBindingDetail';
import { ContractLogicalInterface } from '../components/ContractLogicalInterface';
import { EmptyState } from '../components/EmptyState';
import { StatusBadge } from '../components/StatusBadge';
import { isEditor } from '../utils/roles';
import { useDataPlaneContext } from './useDataPlaneContext';

export function DataContractsPage() {
  const { client, currentTenant, businessId } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [items, setItems] = useState<FormaDataContract[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState('');

  const load = useCallback(async () => {
    if (!businessId) return;
    setLoading(true);
    setError(null);
    try {
      const resp = await client.listDataContracts(businessId);
      setItems(resp.data ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [client, businessId]);

  useEffect(() => {
    void load();
  }, [load]);

  const create = async () => {
    if (!name.trim()) return;
    await client.createDataContract(businessId, {
      business_model_revision: 1,
      name: name.trim(),
      description: name.trim(),
      requirement_ids: [],
      logical_schema: { fields: [] },
      mapping_ids: [],
    });
    setName('');
    await load();
  };

  const q = businessId ? `?businessId=${encodeURIComponent(businessId)}` : '';

  if (!businessId) {
    return <EmptyState title="请选择业务资产" hint="数据契约提供消费者逻辑接口。" />;
  }

  return (
    <div data-testid="data-contracts-page">
      <div className="forma-data-toolbar" style={{ marginBottom: 12 }}>
        <h2 style={{ margin: 0 }}>数据契约</h2>
      </div>
      {error ? <div className="forma-error">{error}</div> : null}
      {canEdit ? (
        <div className="forma-panel" style={{ marginBottom: 12 }}>
          <div className="forma-form-row">
            <label>契约名称</label>
            <input value={name} onChange={e => setName(e.target.value)} />
          </div>
          <button className="forma-btn forma-btn-primary" type="button" onClick={() => void create()}>
            创建契约
          </button>
        </div>
      ) : null}
      {loading ? <div className="forma-muted">加载中…</div> : null}
      {!loading && items.length === 0 ? (
        <EmptyState title="暂无数据契约" />
      ) : (
        items.map(c => (
          <div className="forma-card" key={c.contract_id}>
            <Link to={`../contracts/${c.contract_id}${q}`}>{c.contract_id}</Link>
            <span className="forma-muted" style={{ marginLeft: 8 }}>
              active: {c.active_revision_id ?? '—'}
            </span>
          </div>
        ))
      )}
    </div>
  );
}

export function ContractDetailPage() {
  const { contractId = '' } = useParams();
  const { client, currentTenant, businessId } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [tab, setTab] = useState<'logical' | 'binding' | 'revisions'>('logical');
  const [descriptor, setDescriptor] = useState<FormaDataContractDescriptor | null>(null);
  const [revisions, setRevisions] = useState<FormaDataContractRevision[]>([]);
  const [selected, setSelected] = useState<FormaDataContractRevision | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!businessId || !contractId) return;
    setError(null);
    try {
      const [revs, desc] = await Promise.all([
        client.listDataContractRevisions(businessId, contractId),
        client.getActiveDataContractDescriptor(businessId, contractId).catch(() => null),
      ]);
      setRevisions(revs.data ?? []);
      setDescriptor(desc?.data ?? null);
      setSelected(revs.data?.[0] ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    }
  }, [client, businessId, contractId]);

  useEffect(() => {
    void load();
  }, [load]);

  const deprecate = async (rev: FormaDataContractRevision) => {
    await client.deprecateDataContractRevision(businessId, contractId, rev.revision_id, {
      reason: 'ui-deprecate-stale',
    });
    setMessage(`已弃用修订 ${rev.revision_id}（状态 ${rev.status}）`);
    await load();
  };

  const q = businessId ? `?businessId=${encodeURIComponent(businessId)}` : '';

  if (!businessId || !contractId) {
    return <EmptyState title="缺少契约上下文" />;
  }

  const hasStale = revisions.some(r => r.status === 'STALE');
  const hasActive = revisions.some(r => r.status === 'ACTIVE');

  return (
    <div data-testid="contract-detail-page">
      <div className="forma-data-toolbar">
        <Link to={`../contracts${q}`}>← 返回</Link>
        <h2 style={{ margin: 0 }}>{contractId}</h2>
      </div>
      {hasStale && hasActive ? (
        <div className="forma-banner forma-banner-warn" data-testid="stale-warning">
          存在 STALE 历史修订，同时有更新的 ACTIVE 修订。可对历史 STALE 执行弃用。
        </div>
      ) : null}
      {message ? (
        <div className="forma-banner" data-testid="deprecate-success">
          {message}
        </div>
      ) : null}
      {error ? <div className="forma-error">{error}</div> : null}
      <div className="forma-tabs">
        <button
          type="button"
          className={`forma-tab${tab === 'logical' ? ' active' : ''}`}
          onClick={() => setTab('logical')}
        >
          逻辑接口
        </button>
        <button
          type="button"
          className={`forma-tab${tab === 'binding' ? ' active' : ''}`}
          onClick={() => setTab('binding')}
        >
          物理绑定
        </button>
        <button
          type="button"
          className={`forma-tab${tab === 'revisions' ? ' active' : ''}`}
          onClick={() => setTab('revisions')}
        >
          修订
        </button>
      </div>
      {tab === 'logical' ? (
        descriptor ? (
          <ContractLogicalInterface descriptor={descriptor} />
        ) : (
          <EmptyState title="无生效逻辑接口" hint="需先激活契约修订。" />
        )
      ) : null}
      {tab === 'binding' ? (
        <ContractBindingDetail bindings={selected?.binding_refs ?? []} />
      ) : null}
      {tab === 'revisions' ? (
        <div>
          {revisions.map(r => (
            <div className="forma-card" key={r.revision_id}>
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <strong>
                  v{r.version} {r.name}
                </strong>
                <StatusBadge status={r.status} />
              </div>
              <div className="forma-card-actions">
                <button className="forma-btn" type="button" onClick={() => setSelected(r)}>
                  查看绑定
                </button>
                {canEdit && r.status === 'DRAFT' ? (
                  <>
                    <button
                      className="forma-btn"
                      type="button"
                      onClick={() =>
                        void client.validateDataContractRevision(businessId, contractId, r.revision_id)
                      }
                    >
                      校验
                    </button>
                    <button
                      className="forma-btn forma-btn-primary"
                      type="button"
                      onClick={() =>
                        void client
                          .activateDataContractRevision(businessId, contractId, r.revision_id)
                          .then(load)
                      }
                    >
                      激活
                    </button>
                  </>
                ) : null}
                {canEdit && (r.status === 'STALE' || r.status === 'ACTIVE' || r.status === 'DRAFT') ? (
                  <button
                    className="forma-btn forma-btn-danger"
                    type="button"
                    data-testid="deprecate-revision"
                    onClick={() => void deprecate(r)}
                  >
                    弃用
                  </button>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
