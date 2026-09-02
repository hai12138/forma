import { useCallback, useEffect, useState } from 'react';

import type { FormaDriftResult, FormaGapResult, FormaLifecycleEvent, FormaValidationResult } from '@forma/api-client';

import { EmptyState } from '../components/EmptyState';
import { isEditor } from '../utils/roles';
import { useDataPlaneContext } from './useDataPlaneContext';

export function DataHealthPage() {
  const { client, currentTenant, businessId } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [contractId, setContractId] = useState('');
  const [revisionId, setRevisionId] = useState('');
  const [validations, setValidations] = useState<FormaValidationResult[]>([]);
  const [drifts, setDrifts] = useState<FormaDriftResult[]>([]);
  const [gaps, setGaps] = useState<FormaGapResult[]>([]);
  const [events, setEvents] = useState<FormaLifecycleEvent[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!businessId || !contractId) return;
    setError(null);
    try {
      const life = await client.listDataContractLifecycleEvents(businessId, contractId);
      setEvents(life.data ?? []);
      if (revisionId) {
        const [v, d, g] = await Promise.all([
          client.listDataContractValidationResults(businessId, contractId, revisionId),
          client.listDataContractDriftResults(businessId, contractId, revisionId),
          client.listDataContractGapResults(businessId, contractId, revisionId),
        ]);
        setValidations(v.data ?? []);
        setDrifts(d.data ?? []);
        setGaps(g.data ?? []);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    }
  }, [client, businessId, contractId, revisionId]);

  useEffect(() => {
    void load();
  }, [load]);

  if (!businessId) {
    return <EmptyState title="请选择业务资产" hint="查看契约校验、漂移、缺口与生命周期事件。" />;
  }

  return (
    <div data-testid="data-health-page">
      <h2 style={{ marginTop: 0 }}>数据健康</h2>
      <div className="forma-panel" style={{ marginBottom: 12 }}>
        <div className="forma-form-row">
          <label>契约 ID</label>
          <input value={contractId} onChange={e => setContractId(e.target.value)} />
        </div>
        <div className="forma-form-row">
          <label>修订 ID</label>
          <input value={revisionId} onChange={e => setRevisionId(e.target.value)} />
        </div>
        <button className="forma-btn" type="button" onClick={() => void load()}>
          刷新
        </button>
        {canEdit && contractId && revisionId ? (
          <>
            <button
              className="forma-btn"
              type="button"
              onClick={() =>
                void client
                  .evaluateDataContractDrift(businessId, contractId, revisionId, {
                    new_snapshot_ids: {},
                  })
                  .then(() => {
                    setMessage('漂移评估已提交');
                    return load();
                  })
              }
            >
              评估漂移
            </button>
            <button
              className="forma-btn"
              type="button"
              onClick={() =>
                void client.evaluateDataContractGap(businessId, contractId, revisionId).then(() => {
                  setMessage('缺口评估已提交');
                  return load();
                })
              }
            >
              评估缺口
            </button>
          </>
        ) : null}
      </div>
      {message ? <div className="forma-banner">{message}</div> : null}
      {error ? <div className="forma-error">{error}</div> : null}
      {!contractId ? (
        <EmptyState title="输入契约 ID 以查看健康信号" />
      ) : (
        <>
          <section className="forma-panel">
            <h3>校验结果</h3>
            {validations.length === 0 ? (
              <EmptyState title="无校验结果" />
            ) : (
              validations.map(v => (
                <div key={v.ValidationID} className="forma-card">
                  {v.Status} · {v.ValidationID}
                </div>
              ))
            )}
          </section>
          <section className="forma-panel">
            <h3>漂移</h3>
            {drifts.length === 0 ? (
              <EmptyState title="无漂移结果" />
            ) : (
              drifts.map(d => (
                <div key={d.DriftResultID} className="forma-card">
                  {d.Severity} · findings {d.Findings?.length ?? 0}
                </div>
              ))
            )}
          </section>
          <section className="forma-panel">
            <h3>缺口</h3>
            {gaps.length === 0 ? (
              <EmptyState title="无缺口结果" />
            ) : (
              gaps.map(g => (
                <div key={g.GapResultID} className="forma-card">
                  {g.GapStatus} · unmapped {g.UnmappedRequirementIDs?.length ?? 0}
                </div>
              ))
            )}
          </section>
          <section className="forma-panel">
            <h3>生命周期</h3>
            {events.length === 0 ? (
              <EmptyState title="无生命周期事件" />
            ) : (
              events.map(e => (
                <div key={e.EventID} className="forma-card">
                  {e.Action} · rev {e.RevisionID}
                </div>
              ))
            )}
          </section>
        </>
      )}
    </div>
  );
}
