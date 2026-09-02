import { useEffect, useState } from 'react';
import { NavLink, Outlet, useSearchParams } from 'react-router-dom';

import type { FormaApiClient, FormaBusiness, FormaTenant } from '@forma/api-client';
import { FormaApiError } from '@forma/api-client';

import '../styles/data.css';

export interface DataPlaneShellProps {
  client: FormaApiClient;
  currentTenant: FormaTenant | null;
}

const NAV = [
  { to: '.', end: true, label: '概览' },
  { to: 'requirements', label: '数据需求' },
  { to: 'sources', label: '数据源' },
  { to: 'mappings', label: '映射工作室' },
  { to: 'contracts', label: '数据契约' },
  { to: 'health', label: '健康' },
];

export function DataPlaneShell({ client, currentTenant }: DataPlaneShellProps) {
  const tenantId = currentTenant?.tenant_id ?? '';
  const [searchParams, setSearchParams] = useSearchParams();
  const [businesses, setBusinesses] = useState<FormaBusiness[]>([]);
  const [error, setError] = useState<string | null>(null);
  const businessId = searchParams.get('businessId') ?? '';

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setError(null);
      if (!tenantId) {
        setBusinesses([]);
        if (businessId) {
          const next = new URLSearchParams(searchParams);
          next.delete('businessId');
          setSearchParams(next, { replace: true });
        }
        return;
      }
      try {
        const resp = await client.listBusinesses();
        if (cancelled) return;
        const list = resp.data ?? [];
        setBusinesses(list);
        if (businessId && !list.some(b => b.business_id === businessId)) {
          const next = new URLSearchParams(searchParams);
          next.delete('businessId');
          setSearchParams(next, { replace: true });
        }
      } catch (err) {
        if (cancelled) return;
        setBusinesses([]);
        if (err instanceof FormaApiError && err.code === 'UNAUTHORIZED') {
          setError('未登录');
        } else {
          setError(err instanceof Error ? err.message : '加载业务失败');
        }
      }
    })();
    return () => {
      cancelled = true;
    };
    // Clear business on tenant change; re-sync business list.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client, tenantId]);

  const onBusinessChange = (id: string) => {
    const next = new URLSearchParams(searchParams);
    if (id) next.set('businessId', id);
    else next.delete('businessId');
    setSearchParams(next, { replace: true });
  };

  return (
    <div className="forma-data-shell" data-testid="data-plane-shell">
      <div className="forma-data-toolbar">
        <strong>数据平面</strong>
        <label>
          业务资产{' '}
          <select
            value={businessId}
            onChange={e => onBusinessChange(e.target.value)}
            data-testid="business-selector"
          >
            <option value="">选择业务…</option>
            {businesses.map(b => (
              <option key={b.business_id} value={b.business_id}>
                {b.name}
              </option>
            ))}
          </select>
        </label>
        <span className="forma-muted">租户角色：{currentTenant?.role ?? '—'}</span>
      </div>
      <nav className="forma-data-plane-nav" aria-label="数据平面导航">
        {NAV.map(item => (
          <NavLink
            key={item.to}
            to={item.to === '.' ? { pathname: '.', search: searchParams.toString() } : { pathname: item.to, search: searchParams.toString() }}
            end={item.end}
            className={({ isActive }) => (isActive ? 'active' : undefined)}
          >
            {item.label}
          </NavLink>
        ))}
      </nav>
      {error ? <div className="forma-error">{error}</div> : null}
      <Outlet context={{ client, currentTenant, businessId, businesses }} />
    </div>
  );
}

export interface DataPlaneOutletContext {
  client: FormaApiClient;
  currentTenant: FormaTenant | null;
  businessId: string;
  businesses: FormaBusiness[];
}
