import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import type {
  FormaContractBinding,
  FormaDataContract,
  FormaDataContractRevision,
  FormaDriftResult,
  FormaGapResult,
  FormaLifecycleEvent,
  FormaSchemaSnapshot,
  FormaValidationResult,
} from '@forma/api-client';

import { EmptyState } from '../components/EmptyState';
import { safeMutate, sanitizedErrorMessage } from '../utils/errors';
import { isEditor } from '../utils/roles';
import { useDataPlaneContext } from './useDataPlaneContext';

type FreshMap = Record<string, string>;

function uniqueBindings(bindings: FormaContractBinding[]): FormaContractBinding[] {
  const seen = new Set<string>();
  const out: FormaContractBinding[] = [];
  for (const b of bindings) {
    const key = b.schema_snapshot_id;
    if (!key || seen.has(key)) continue;
    seen.add(key);
    out.push(b);
  }
  return out;
}

export function DataHealthPage() {
  const { client, currentTenant, businessId } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [contractId, setContractId] = useState('');
  const [revisionId, setRevisionId] = useState('');
  const [revision, setRevision] = useState<FormaDataContractRevision | null>(null);
  const [contract, setContract] = useState<FormaDataContract | null>(null);
  const [freshByPinned, setFreshByPinned] = useState<FreshMap>({});
  const [snapshotsByPinned, setSnapshotsByPinned] = useState<Record<string, FormaSchemaSnapshot[]>>(
    {},
  );
  const [validations, setValidations] = useState<FormaValidationResult[]>([]);
  const [drifts, setDrifts] = useState<FormaDriftResult[]>([]);
  const [gaps, setGaps] = useState<FormaGapResult[]>([]);
  const [events, setEvents] = useState<FormaLifecycleEvent[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [lastDriftSeverity, setLastDriftSeverity] = useState<string | null>(null);
  const loadGenRef = useRef(0);

  const pinnedBindings = useMemo(
    () => uniqueBindings(revision?.binding_refs ?? []),
    [revision?.binding_refs],
  );

  const allFreshSelected =
    pinnedBindings.length > 0 &&
    pinnedBindings.every(b => Boolean(freshByPinned[b.schema_snapshot_id]));

  const loadRevisionContext = useCallback(async () => {
    if (!businessId || !contractId || !revisionId) {
      return { cleared: true as const };
    }
    setError(null);
    try {
      const [revResp, ctrResp] = await Promise.all([
        client.getDataContractRevision(businessId, contractId, revisionId),
        client.getDataContract(businessId, contractId),
      ]);
      const bindings = uniqueBindings(revResp.data.binding_refs ?? []);
      const nextSnaps: Record<string, FormaSchemaSnapshot[]> = {};
      await Promise.all(
        bindings.map(async b => {
          try {
            const list = await client.listSchemaSnapshots({
              sourceId: b.source_id,
              connectionId: b.connection_id,
              assetId: b.asset_id,
            });
            nextSnaps[b.schema_snapshot_id] = (list.data ?? []).filter(
              s =>
                s.source_id === b.source_id &&
                s.connection_id === b.connection_id &&
                s.asset_id === b.asset_id,
            );
          } catch {
            nextSnaps[b.schema_snapshot_id] = [];
          }
        }),
      );
      return {
        cleared: false as const,
        revision: revResp.data,
        contract: ctrResp.data,
        snapshots: nextSnaps,
      };
    } catch (err) {
      setError(sanitizedErrorMessage(err));
      return { cleared: true as const };
    }
  }, [client, businessId, contractId, revisionId]);

  const load = useCallback(async () => {
    if (!businessId || !contractId) return;
    const gen = ++loadGenRef.current;
    setError(null);
    try {
      const life = await client.listDataContractLifecycleEvents(businessId, contractId);
      if (gen !== loadGenRef.current) return;
      setEvents(life.data ?? []);
      if (revisionId) {
        const [v, d, g] = await Promise.all([
          client.listDataContractValidationResults(businessId, contractId, revisionId),
          client.listDataContractDriftResults(businessId, contractId, revisionId),
          client.listDataContractGapResults(businessId, contractId, revisionId),
        ]);
        if (gen !== loadGenRef.current) return;
        setValidations(v.data ?? []);
        setDrifts(d.data ?? []);
        setGaps(g.data ?? []);
      }
      const ctx = await loadRevisionContext();
      if (gen !== loadGenRef.current) return;
      if (ctx.cleared) {
        if (!revisionId) {
          setRevision(null);
          setContract(null);
          setSnapshotsByPinned({});
          setFreshByPinned({});
        }
        return;
      }
      setRevision(ctx.revision);
      setContract(ctx.contract);
      setSnapshotsByPinned(ctx.snapshots);
    } catch (err) {
      if (gen !== loadGenRef.current) return;
      setError(sanitizedErrorMessage(err));
    }
  }, [client, businessId, contractId, revisionId, loadRevisionContext]);

  useEffect(() => {
    setFreshByPinned({});
    setLastDriftSeverity(null);
  }, [contractId, revisionId]);

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
          <input
            value={contractId}
            onChange={e => setContractId(e.target.value)}
            data-testid="health-contract-id"
          />
        </div>
        <div className="forma-form-row">
          <label>修订 ID</label>
          <input
            value={revisionId}
            onChange={e => setRevisionId(e.target.value)}
            data-testid="health-revision-id"
          />
        </div>
        <button className="forma-btn" type="button" onClick={() => void load()}>
          刷新
        </button>
        {contract ? (
          <div className="forma-muted" data-testid="health-contract-status">
            契约状态指针：active_revision_id=
            {contract.active_revision_id || '（空）'} · revision=
            {revision?.status || '—'}
          </div>
        ) : null}
        {canEdit && contractId && revisionId ? (
          <>
            <div
              className="forma-panel"
              style={{ marginTop: 12 }}
              data-testid="drift-snapshot-picker"
            >
              <h3 style={{ marginTop: 0 }}>漂移快照映射</h3>
              {pinnedBindings.length === 0 ? (
                <EmptyState title="当前修订无 pinned SchemaSnapshot" hint="加载含 binding_refs 的修订后可评估漂移。" />
              ) : (
                pinnedBindings.map(b => (
                  <div
                    key={b.schema_snapshot_id}
                    className="forma-card"
                    data-testid={`pinned-snapshot-${b.schema_snapshot_id}`}
                  >
                    <div>
                      <strong>Pinned Snapshot</strong>：{b.schema_snapshot_id}
                    </div>
                    <div className="forma-muted">
                      asset={b.asset_id} · source={b.source_id} · connection={b.connection_id}
                    </div>
                    <div className="forma-form-row">
                      <label htmlFor={`fresh-${b.schema_snapshot_id}`}>Fresh Snapshot</label>
                      <select
                        id={`fresh-${b.schema_snapshot_id}`}
                        data-testid={`fresh-snapshot-select-${b.schema_snapshot_id}`}
                        value={freshByPinned[b.schema_snapshot_id] || ''}
                        onChange={e =>
                          setFreshByPinned(prev => ({
                            ...prev,
                            [b.schema_snapshot_id]: e.target.value,
                          }))
                        }
                      >
                        <option value="">选择同 Asset 的 fresh snapshot…</option>
                        {(snapshotsByPinned[b.schema_snapshot_id] || []).map(s => (
                          <option key={s.snapshot_id} value={s.snapshot_id}>
                            {s.snapshot_id} · {s.fingerprint?.slice(0, 8) || 'fp'} ·{' '}
                            {s.created_at}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                ))
              )}
            </div>
            <button
              className="forma-btn"
              type="button"
              data-testid="evaluate-drift"
              disabled={!allFreshSelected}
              onClick={() =>
                void safeMutate(async () => {
                  const resp = await client.evaluateDataContractDrift(
                    businessId,
                    contractId,
                    revisionId,
                    { new_snapshot_ids: { ...freshByPinned } },
                  );
                  const severity = resp.data?.result?.Severity || '';
                  setLastDriftSeverity(severity);
                  setMessage(`漂移评估完成：${severity || 'OK'}`);
                  if (resp.data?.revision) {
                    setRevision(resp.data.revision);
                  }
                  await load();
                }, setError)
              }
            >
              评估漂移
            </button>
            <button
              className="forma-btn"
              type="button"
              data-testid="evaluate-gap"
              onClick={() =>
                void safeMutate(async () => {
                  await client.evaluateDataContractGap(businessId, contractId, revisionId);
                  setMessage('缺口评估已提交');
                  await load();
                }, setError)
              }
            >
              评估缺口
            </button>
          </>
        ) : null}
      </div>
      {message ? (
        <div className="forma-banner" data-testid="health-message">
          {message}
        </div>
      ) : null}
      {lastDriftSeverity ? (
        <div className="forma-banner" data-testid="drift-severity-banner">
          最近漂移结果：{lastDriftSeverity}
        </div>
      ) : null}
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
                <div
                  key={d.DriftResultID}
                  className="forma-card"
                  data-testid="drift-result-card"
                  data-severity={d.Severity}
                >
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
                <div key={g.GapResultID} className="forma-card" data-testid="gap-result-card">
                  {g.GapStatus} · pinned revision {g.FromBusinessRevision} · current revision{' '}
                  {g.CurrentBusinessRevision} · new confirmed{' '}
                  {g.NewConfirmedRequirementIDs?.length ?? 0} · unmapped{' '}
                  {g.UnmappedRequirementIDs?.length ?? 0}
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
