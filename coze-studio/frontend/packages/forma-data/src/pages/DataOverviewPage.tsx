import { useEffect, useState } from 'react';

import type { FormaMappingCoverage } from '@forma/api-client';

import { EmptyState } from '../components/EmptyState';
import { readinessLabel } from '../utils/labels';
import { useDataPlaneContext } from './useDataPlaneContext';

export function DataOverviewPage() {
  const { client, businessId, businesses } = useDataPlaneContext();
  const business = businesses.find(b => b.business_id === businessId);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirmed, setConfirmed] = useState(0);
  const [coverage, setCoverage] = useState<FormaMappingCoverage | null>(null);
  const [activeContracts, setActiveContracts] = useState(0);
  const [staleContracts, setStaleContracts] = useState(0);

  useEffect(() => {
    if (!businessId) return;
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const [reqs, cov, contracts] = await Promise.all([
          client.listDataRequirements(businessId, {
            status: 'CONFIRMED',
            business_model_revision: business?.current_revision,
          }),
          client.getSemanticMappingCoverage(businessId),
          client.listDataContracts(businessId),
        ]);
        if (cancelled) return;
        setConfirmed((reqs.data ?? []).length);
        setCoverage(cov.data);
        let active = 0;
        let stale = 0;
        for (const c of contracts.data ?? []) {
          if (!c.active_revision_id) continue;
          try {
            const rev = await client.getDataContractRevision(
              businessId,
              c.contract_id,
              c.active_revision_id,
            );
            if (rev.data?.status === 'ACTIVE') active += 1;
            if (rev.data?.status === 'STALE') stale += 1;
          } catch {
            /* ignore per-contract errors for overview */
          }
        }
        if (!cancelled) {
          setActiveContracts(active);
          setStaleContracts(stale);
        }
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : '加载失败');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [client, businessId, business?.current_revision]);

  if (!businessId) {
    return <EmptyState title="请选择业务资产" hint="数据平面按业务展示需求、映射与契约就绪状态。" />;
  }
  if (loading) return <div className="forma-muted">加载中…</div>;
  if (error) return <div className="forma-error">{error}</div>;

  const text = readinessLabel({
    confirmedRequirements: confirmed,
    coverage: coverage?.coverage ?? 0,
    activeContracts,
    staleContracts,
  });

  return (
    <div className="forma-panel" data-testid="data-overview">
      <h2 style={{ marginTop: 0 }}>就绪状态</h2>
      <p data-testid="readiness-label">{text}</p>
      <ul className="forma-muted">
        <li>已确认需求：{confirmed}</li>
        <li>
          映射覆盖：{coverage ? `${(coverage.coverage * 100).toFixed(0)}%` : '—'}（
          {coverage?.confirmed_mappings ?? 0}/{coverage?.total_confirmed_requirements ?? 0}）
        </li>
        <li>生效契约：{activeContracts}</li>
        <li>过期契约：{staleContracts}</li>
      </ul>
    </div>
  );
}
