import { NavLink } from 'react-router-dom';

import { navigation } from '@/lib/navigation';
import { useFormaSession } from '@/hooks/use-forma-session';

import './shell.css';

function TenantSwitcher() {
  const { state, tenants, currentTenant, switchTenant } = useFormaSession();

  if (state === 'loading' || !currentTenant) {
    return <div className="forma-tenant-switcher forma-muted">Tenant…</div>;
  }

  if (tenants.length <= 1) {
    return (
      <div className="forma-tenant-switcher">
        <span className="forma-muted">Workspace</span>
        <strong>{currentTenant.display_name || currentTenant.name}</strong>
      </div>
    );
  }

  return (
    <div className="forma-tenant-switcher">
      <label className="forma-muted" htmlFor="forma-tenant-select">
        Workspace
      </label>
      <select
        id="forma-tenant-select"
        value={currentTenant.tenant_id}
        onChange={e => {
          void switchTenant(e.target.value);
        }}
      >
        {tenants.map(t => (
          <option key={t.tenant_id} value={t.tenant_id}>
            {t.display_name || t.name}
          </option>
        ))}
      </select>
    </div>
  );
}

function SessionBanner() {
  const { state, error, bootstrap, refresh } = useFormaSession();

  if (state === 'ready' || state === 'loading') {
    return null;
  }

  const messages: Record<string, string> = {
    unauthenticated: '未登录。请先完成 Coze Session 认证，再进入 Forma。',
    forbidden: '当前身份无权访问所选 Tenant。',
    suspended: 'Tenant 已暂停，业务 API 已拒绝。',
    empty: '尚未加入任何 Tenant。可执行 Bootstrap 创建默认 Workspace。',
    network_error: `网络错误：${error || 'unknown'}`,
  };

  return (
    <div className="forma-banner" role="status">
      <span>{messages[state] || error}</span>
      <div className="forma-banner-actions">
        {state === 'empty' && (
          <button type="button" onClick={() => void bootstrap()}>
            Bootstrap
          </button>
        )}
        <button type="button" onClick={() => void refresh()}>
          Retry
        </button>
      </div>
    </div>
  );
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const { state, me, currentTenant } = useFormaSession();

  return (
    <div className="forma-shell">
      <aside className="forma-sidebar">
        <div className="forma-brand">
          <span className="forma-brand-mark">F</span>
          <div>
            <strong>Forma</strong>
            <div className="forma-brand-sub">Business-to-Agent</div>
          </div>
        </div>
        <TenantSwitcher />
        {navigation.map(group => (
          <div key={group.group} className="forma-nav-group">
            <div className="forma-nav-group-title">{group.group}</div>
            {group.items.map(item => (
              <NavLink
                key={item.id}
                to={item.path}
                end={item.path === '/'}
                className={({ isActive }) =>
                  `forma-nav-item${isActive ? ' active' : ''}`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </div>
        ))}
      </aside>
      <main className="forma-main">
        <header className="forma-topbar">
          <div>
            <div className="forma-eyebrow">Reference Business · 维修工单</div>
            <div className="forma-env">
              {currentTenant
                ? `${currentTenant.display_name} · ${currentTenant.status}`
                : 'Forma Product Shell'}
            </div>
          </div>
          <div className="forma-topbar-meta">
            {state === 'loading' && <span>Loading identity…</span>}
            {me?.principal && (
              <span>
                {me.principal.display_name || me.principal.principal_id}
              </span>
            )}
          </div>
        </header>
        <SessionBanner />
        <div className="forma-content">{children}</div>
      </main>
    </div>
  );
}
