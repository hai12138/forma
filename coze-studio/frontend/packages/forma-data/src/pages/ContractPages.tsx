import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import type {
  FormaDataContract,
  FormaDataContractDescriptor,
  FormaDataContractRevision,
  FormaValidationResult,
} from '@forma/api-client';

import { ContractBindingDetail } from '../components/ContractBindingDetail';
import { ContractLogicalInterface } from '../components/ContractLogicalInterface';
import { EmptyState } from '../components/EmptyState';
import { StatusBadge } from '../components/StatusBadge';
import {
  canActivateRevision,
  canDeprecateRevision,
  canValidateRevision,
} from '../utils/contract-lifecycle';
import { safeMutate } from '../utils/errors';
import { isEditor } from '../utils/roles';
import { useDataPlaneContext } from './useDataPlaneContext';

export function DataContractsPage() {
  const { client, currentTenant, businessId } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [items, setItems] = useState<FormaDataContract[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [busy, setBusy] = useState(false);

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

  const create = () => {
    if (!name.trim() || busy) return;
    setBusy(true);
    void safeMutate(async () => {
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
    }, setError).finally(() => setBusy(false));
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
          <button
            className="forma-btn forma-btn-primary"
            type="button"
            data-testid="create-contract"
            disabled={busy}
            onClick={create}
          >
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
  const [validation, setValidation] = useState<FormaValidationResult | null>(null);
  const [busy, setBusy] = useState(false);
  const [pendingActivateId, setPendingActivateId] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!businessId || !contractId) return;
    setError(null);
    try {
      const [revs, desc] = await Promise.all([
        client.listDataContractRevisions(businessId, contractId),
        client.getActiveDataContractDescriptor(businessId, contractId).catch(() => null),
      ]);
      const list = revs.data ?? [];
      setRevisions(list);
      setDescriptor(desc?.data ?? null);
      setSelected(prev => list.find(r => r.revision_id === prev?.revision_id) ?? list[0] ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    }
  }, [client, businessId, contractId]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!canEdit && tab === 'binding') {
      setTab('logical');
    }
  }, [canEdit, tab]);

  const runRevisionAction = (fn: () => Promise<void>) => {
    if (busy) return;
    setBusy(true);
    setError(null);
    void safeMutate(fn, setError).finally(() => setBusy(false));
  };

  const validate = (rev: FormaDataContractRevision) => {
    runRevisionAction(async () => {
      const resp = await client.validateDataContractRevision(businessId, contractId, rev.revision_id);
      setValidation(resp.data.result ?? null);
      setMessage(resp.data.result?.Status === 'FAIL' ? '校验未通过，修订仍为草稿。' : '校验通过。');
      await load();
    });
  };

  const activate = (revisionId: string) => {
    setPendingActivateId(null);
    runRevisionAction(async () => {
      await client.activateDataContractRevision(businessId, contractId, revisionId, {
        reason: 'ui-activate',
      });
      setMessage('已启用新版本。');
      await load();
    });
  };

  const deprecate = (rev: FormaDataContractRevision) => {
    runRevisionAction(async () => {
      await client.deprecateDataContractRevision(businessId, contractId, rev.revision_id, {
        reason: 'ui-deprecate',
      });
      setMessage(`已停用修订 ${rev.revision_id}`);
      await load();
    });
  };

  const q = businessId ? `?businessId=${encodeURIComponent(businessId)}` : '';

  if (!businessId || !contractId) {
    return <EmptyState title="缺少契约上下文" />;
  }

  const hasStale = revisions.some(r => r.status === 'STALE');
  const hasActive = revisions.some(r => r.status === 'ACTIVE');
  const selectedIsStale = selected?.status === 'STALE';

  return (
    <div data-testid="contract-detail-page">
      <div className="forma-data-toolbar">
        <Link to={`../contracts${q}`}>← 返回</Link>
        <h2 style={{ margin: 0 }}>{contractId}</h2>
      </div>
      {selectedIsStale ? (
        <div className="forma-banner forma-banner-warn" data-testid="stale-blocking-warning">
          底层数据结构已发生破坏性变化，该契约当前不可作为活动接口使用。
        </div>
      ) : null}
      {hasStale && hasActive ? (
        <div className="forma-banner forma-banner-warn" data-testid="stale-warning">
          存在 STALE 历史修订，同时有更新的 ACTIVE 修订。可对历史 STALE 执行停用。
        </div>
      ) : null}
      {message ? (
        <div className="forma-banner" data-testid="lifecycle-success">
          {message}
        </div>
      ) : null}
      {error ? (
        <div className="forma-error" data-testid="contract-error">
          {error}
        </div>
      ) : null}
      {validation ? (
        <div className="forma-panel" data-testid="validation-result">
          校验结果：{validation.Status}
          {(validation.Errors ?? []).length > 0 ? (
            <ul>
              {validation.Errors.map((e, i) => (
                <li key={`${e.code}-${i}`}>
                  {e.code}: {e.message}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      ) : null}
      {pendingActivateId ? (
        <div className="forma-panel" data-testid="activate-confirm-dialog">
          <p>启用新版本后，旧活动版本将停止作为默认版本。</p>
          <div className="forma-card-actions">
            <button
              className="forma-btn forma-btn-primary"
              type="button"
              data-testid="activate-confirm"
              disabled={busy}
              onClick={() => activate(pendingActivateId)}
            >
              确认启用
            </button>
            <button className="forma-btn" type="button" onClick={() => setPendingActivateId(null)}>
              取消
            </button>
          </div>
        </div>
      ) : null}
      <div className="forma-tabs">
        <button
          type="button"
          className={`forma-tab${tab === 'logical' ? ' active' : ''}`}
          onClick={() => setTab('logical')}
        >
          逻辑接口
        </button>
        {canEdit ? (
          <button
            type="button"
            className={`forma-tab${tab === 'binding' ? ' active' : ''}`}
            data-testid="physical-binding-tab"
            onClick={() => setTab('binding')}
          >
            物理绑定
          </button>
        ) : null}
        <button
          type="button"
          className={`forma-tab${tab === 'revisions' ? ' active' : ''}`}
          data-testid="revisions-tab"
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
      {tab === 'binding' && canEdit ? (
        <div>
          <p className="forma-muted" data-testid="binding-admin-notice">
            物理绑定不是上层业务能力接口的一部分。仅管理员可查看。
          </p>
          <ContractBindingDetail bindings={selected?.binding_refs ?? []} />
        </div>
      ) : null}
      {tab === 'revisions' ? (
        <div>
          {revisions.map(r => (
            <div className="forma-card" key={r.revision_id} data-testid={`revision-card-${r.status}`}>
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <strong>
                  v{r.version} {r.name}
                </strong>
                <StatusBadge status={r.status} />
              </div>
              <div className="forma-card-actions">
                {canEdit ? (
                  <button className="forma-btn" type="button" onClick={() => setSelected(r)}>
                    查看绑定
                  </button>
                ) : null}
                {canEdit && canValidateRevision(r.status) ? (
                  <button
                    className="forma-btn"
                    type="button"
                    data-testid="validate-revision"
                    disabled={busy}
                    onClick={() => validate(r)}
                  >
                    验证契约
                  </button>
                ) : null}
                {canEdit && canActivateRevision(r.status) ? (
                  <button
                    className="forma-btn forma-btn-primary"
                    type="button"
                    data-testid="activate-revision"
                    disabled={busy}
                    onClick={() => setPendingActivateId(r.revision_id)}
                  >
                    启用
                  </button>
                ) : null}
                {canEdit && canDeprecateRevision(r.status) ? (
                  <button
                    className="forma-btn forma-btn-danger"
                    type="button"
                    data-testid="deprecate-revision"
                    disabled={busy}
                    onClick={() => deprecate(r)}
                  >
                    停用
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
