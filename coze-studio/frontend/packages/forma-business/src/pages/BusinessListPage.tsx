import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type { FormaApiClient, FormaBusiness, FormaTenant } from '@forma/api-client';
import { FormaApiError } from '@forma/api-client';

import { createBusinessSubmitHandler, formatUpdatedAt } from '../create-handlers';
import '../styles/editor.css';

export interface BusinessListPageProps {
  client: FormaApiClient;
  /** From useFormaSession().currentTenant — refetch when tenant changes. */
  currentTenant: FormaTenant | null;
}

export { createBusinessSubmitHandler, formatUpdatedAt } from '../create-handlers';

export function BusinessListPage({ client, currentTenant }: BusinessListPageProps) {
  const navigate = useNavigate();
  const [items, setItems] = useState<FormaBusiness[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState('');
  const [seedWorkOrder, setSeedWorkOrder] = useState(false);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await client.listBusinesses();
      setItems(resp.data ?? []);
    } catch (err) {
      if (err instanceof FormaApiError && err.code === 'UNAUTHORIZED') {
        setError('未登录');
      } else {
        setError(err instanceof Error ? err.message : '加载失败');
      }
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [client, currentTenant?.tenant_id]);

  useEffect(() => {
    void load();
  }, [load]);

  const onCreate = async () => {
    setBusy(true);
    setError(null);
    await createBusinessSubmitHandler({
      client,
      name,
      seedWorkOrder,
      onCreated: b => {
        setShowCreate(false);
        setName('');
        setSeedWorkOrder(false);
        navigate(`/business/${b.business_id}`);
      },
      onError: msg => setError(msg),
    })();
    setBusy(false);
  };

  const onArchive = async (id: string) => {
    if (!window.confirm('确认归档该业务资产？')) return;
    try {
      await client.archiveBusiness(id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : '归档失败');
    }
  };

  return (
    <div className="forma-panel">
      <div className="forma-biz-list-header">
        <div>
          <h1 style={{ margin: 0 }}>业务资产</h1>
          <p className="forma-placeholder" style={{ margin: '4px 0 0' }}>
            Business Model · Source of Truth
            {currentTenant ? ` · ${currentTenant.display_name}` : ''}
          </p>
        </div>
        <button
          type="button"
          className="forma-vme-btn primary"
          onClick={() => setShowCreate(true)}
        >
          新建业务
        </button>
      </div>

      {error && <p className="forma-error">{error}</p>}
      {loading && <p className="forma-placeholder">加载中…</p>}

      {!loading && items.length === 0 && (
        <div className="forma-biz-empty" data-testid="business-empty-state">
          <strong>暂无业务资产</strong>
          <p>创建第一个 Business Model，开始可视化建模。</p>
          <button
            type="button"
            className="forma-vme-btn primary"
            onClick={() => setShowCreate(true)}
          >
            新建业务
          </button>
        </div>
      )}

      {!loading && items.length > 0 && (
        <table className="forma-biz-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>状态</th>
              <th>修订</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {items.map(b => (
              <tr key={b.business_id}>
                <td>
                  <button
                    type="button"
                    className="forma-vme-btn"
                    style={{ border: 'none', padding: 0, color: 'var(--forma-primary)' }}
                    onClick={() => navigate(`/business/${b.business_id}`)}
                  >
                    {b.name}
                  </button>
                </td>
                <td>
                  <span
                    className={`forma-biz-status${
                      b.status === 'ARCHIVED' ? ' archived' : ''
                    }`}
                  >
                    {b.status}
                  </span>
                </td>
                <td>r{b.current_revision}</td>
                <td>{formatUpdatedAt(b.updated_at)}</td>
                <td style={{ display: 'flex', gap: 6 }}>
                  <button
                    type="button"
                    className="forma-vme-btn"
                    onClick={() => navigate(`/business/${b.business_id}`)}
                  >
                    打开
                  </button>
                  {b.status !== 'ARCHIVED' && (
                    <button
                      type="button"
                      className="forma-vme-btn danger"
                      onClick={() => void onArchive(b.business_id)}
                    >
                      归档
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showCreate && (
        <div className="forma-biz-modal-backdrop" role="dialog" aria-modal="true">
          <div className="forma-biz-modal">
            <h2>新建业务资产</h2>
            <label>
              名称
              <input
                type="text"
                value={name}
                onChange={e => {
                  setName(e.target.value);
                  if (e.target.value.trim() === '维修工单') setSeedWorkOrder(true);
                }}
                placeholder="例如：维修工单"
                autoFocus
              />
            </label>
            <label style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
              <input
                type="checkbox"
                checked={seedWorkOrder}
                onChange={e => setSeedWorkOrder(e.target.checked)}
              />
              使用「维修工单」参考语义模型初始化
            </label>
            <div className="forma-biz-modal-actions">
              <button
                type="button"
                className="forma-vme-btn"
                onClick={() => setShowCreate(false)}
                disabled={busy}
              >
                取消
              </button>
              <button
                type="button"
                className="forma-vme-btn primary"
                onClick={() => void onCreate()}
                disabled={busy}
              >
                {busy ? '创建中…' : '创建'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
