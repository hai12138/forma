import { useCallback, useEffect, useState } from 'react';

import type { FormaBusiness, FormaDataRequirement } from '@forma/api-client';

import { EmptyState } from '../components/EmptyState';
import { StatusBadge } from '../components/StatusBadge';
import { isEditor } from '../utils/roles';
import { useDataPlaneContext } from './useDataPlaneContext';

export function DataRequirementsPage() {
  const { client, currentTenant, businessId, businesses } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [items, setItems] = useState<FormaDataRequirement[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [showManual, setShowManual] = useState(false);
  const [manualName, setManualName] = useState('');
  const [editId, setEditId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');

  const business: FormaBusiness | undefined = businesses.find(b => b.business_id === businessId);

  const load = useCallback(async () => {
    if (!businessId) return;
    setLoading(true);
    setError(null);
    try {
      const resp = await client.listDataRequirements(businessId);
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

  const proposed = items.filter(i => i.status === 'PROPOSED');

  const analyze = async () => {
    if (!business) return;
    setMessage(null);
    try {
      const resp = await client.analyzeDataRequirements(businessId, {
        business_model_revision: business.current_revision,
        client_request_id: `ui-${Date.now()}`,
      });
      setMessage(
        resp.data.owned_execute
          ? `分析完成，提出 ${(resp.data.requirements ?? []).length} 条需求`
          : '已有进行中的分析任务',
      );
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : '分析失败');
    }
  };

  const confirm = async (id: string) => {
    await client.confirmDataRequirement(businessId, id, { reason: 'ui-confirm' });
    await load();
  };
  const reject = async (id: string) => {
    await client.rejectDataRequirement(businessId, id, { reason: 'ui-reject' });
    await load();
  };
  const editConfirm = async (req: FormaDataRequirement) => {
    await client.editConfirmDataRequirement(businessId, req.requirement_id, {
      reason: 'ui-edit-confirm',
      requirement_kind: req.requirement_kind,
      semantic_name: editName || req.semantic_name,
      description: req.description,
      business_element_refs: req.business_element_refs,
      requiredness: req.requiredness,
      freshness_requirement: req.freshness_requirement,
      access_need: req.access_need,
    });
    setEditId(null);
    await load();
  };

  const createManual = async () => {
    if (!business || !manualName.trim()) return;
    await client.createManualDataRequirement(businessId, {
      business_model_revision: business.current_revision,
      requirement_kind: 'ENTITY',
      semantic_name: manualName.trim(),
      description: manualName.trim(),
      business_element_refs: [],
      requiredness: 'REQUIRED',
      freshness_requirement: 'DAILY',
      access_need: 'READ',
    });
    setManualName('');
    setShowManual(false);
    await load();
  };

  if (!businessId) {
    return <EmptyState title="请选择业务资产" hint="确认后的数据需求将用于映射与契约。" />;
  }

  return (
    <div data-testid="data-requirements-page">
      <div className="forma-data-toolbar" style={{ marginBottom: 12 }}>
        <h2 style={{ margin: 0 }}>数据需求</h2>
        {canEdit ? (
          <>
            <button className="forma-btn forma-btn-primary" type="button" onClick={() => void analyze()}>
              从业务模型分析
            </button>
            <button className="forma-btn" type="button" onClick={() => setShowManual(v => !v)}>
              手动新增
            </button>
          </>
        ) : null}
      </div>
      {message ? <div className="forma-banner">{message}</div> : null}
      {proposed.length > 0 ? (
        <div className="forma-banner" data-testid="propose-banner">
          有 {proposed.length} 条待确认的数据需求提案
        </div>
      ) : null}
      {error ? <div className="forma-error">{error}</div> : null}
      {showManual && canEdit ? (
        <div className="forma-panel" style={{ marginBottom: 12 }}>
          <div className="forma-form-row">
            <label>语义名称</label>
            <input value={manualName} onChange={e => setManualName(e.target.value)} />
          </div>
          <button className="forma-btn forma-btn-primary" type="button" onClick={() => void createManual()}>
            创建
          </button>
        </div>
      ) : null}
      {loading ? <div className="forma-muted">加载中…</div> : null}
      {!loading && items.length === 0 ? (
        <EmptyState title="暂无数据需求" hint="可分析业务模型或手动新增。" />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {items.map(req => (
            <div className="forma-card" key={req.requirement_id} data-testid="requirement-card">
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <strong>{req.semantic_name}</strong>
                <StatusBadge status={req.status} />
                <span className="forma-muted">{req.source}</span>
              </div>
              <div className="forma-muted">{req.description}</div>
              {canEdit && req.status === 'PROPOSED' ? (
                <div className="forma-card-actions">
                  <button
                    className="forma-btn forma-btn-primary"
                    type="button"
                    data-testid="confirm-requirement"
                    onClick={() => void confirm(req.requirement_id)}
                  >
                    确认
                  </button>
                  <button
                    className="forma-btn forma-btn-danger"
                    type="button"
                    data-testid="reject-requirement"
                    onClick={() => void reject(req.requirement_id)}
                  >
                    拒绝
                  </button>
                  <button
                    className="forma-btn"
                    type="button"
                    data-testid="edit-confirm-requirement"
                    onClick={() => {
                      setEditId(req.requirement_id);
                      setEditName(req.semantic_name);
                    }}
                  >
                    编辑并确认
                  </button>
                </div>
              ) : null}
              {editId === req.requirement_id && canEdit ? (
                <div className="forma-form-row" style={{ marginTop: 8 }}>
                  <input value={editName} onChange={e => setEditName(e.target.value)} />
                  <button className="forma-btn forma-btn-primary" type="button" onClick={() => void editConfirm(req)}>
                    提交编辑确认
                  </button>
                </div>
              ) : null}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
