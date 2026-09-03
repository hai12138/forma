import { useState } from 'react';
import { NavLink } from 'react-router-dom';

import { adminNavigation, navigation } from '@/lib/navigation';
import { passportLogout } from '@/lib/passport';
import { useFormaSession } from '@/hooks/use-forma-session';

import './shell.css';
import '../pages/auth.css';

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

/** Non-auth banners only — unauthenticated / empty are AuthGuard routes. */
function SessionBanner() {
  const { state, error, refresh } = useFormaSession();

  if (state === 'ready' || state === 'loading' || state === 'unauthenticated' || state === 'empty') {
    return null;
  }

  if (state !== 'network_error' && state !== 'suspended' && state !== 'forbidden') {
    return null;
  }

  const messages: Record<string, string> = {
    forbidden: '当前身份无权访问所选 Workspace。',
    suspended: 'Workspace 已暂停，业务 API 已拒绝。',
    network_error: '网络异常，请检查连接后重试。',
  };

  return (
    <div className="forma-banner" role="status" data-testid="session-banner">
      <span>{messages[state] || error}</span>
      <div className="forma-banner-actions">
        <button type="button" onClick={() => void refresh()}>
          重试
        </button>
      </div>
    </div>
  );
}

function UserMenu() {
  const { me, clearLocalSession } = useFormaSession();
  const [open, setOpen] = useState(false);
  const label = me?.principal?.display_name || me?.principal?.principal_id || '用户';

  return (
    <div className="forma-user-menu" data-testid="user-menu">
      <button
        type="button"
        className="forma-user-menu-btn"
        data-testid="user-menu-trigger"
        onClick={() => setOpen(v => !v)}
      >
        {label}
      </button>
      {open ? (
        <div className="forma-user-menu-panel" data-testid="user-menu-panel">
          <button
            type="button"
            data-testid="logout-button"
            onClick={() => {
              void (async () => {
                setOpen(false);
                await passportLogout();
                clearLocalSession();
                window.location.assign('/login');
              })();
            }}
          >
            退出登录
          </button>
        </div>
      ) : null}
    </div>
  );
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const { me, currentTenant } = useFormaSession();

  return (
    <div className="forma-shell" data-testid="forma-app-shell">
      <aside className="forma-sidebar" data-testid="forma-sidebar">
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
        {me?.principal?.platform_role === 'SUPER_ADMIN'
          ? adminNavigation.map(group => (
              <div key={group.group} className="forma-nav-group">
                <div className="forma-nav-group-title">{group.group}</div>
                {group.items.map(item => (
                  <NavLink
                    key={item.id}
                    to={item.path}
                    end={item.path === '/'}
                    className={({ isActive }) => `forma-nav-item${isActive ? ' active' : ''}`}
                  >
                    {item.label}
                  </NavLink>
                ))}
              </div>
            ))
          : null}
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
            {me?.principal ? <UserMenu /> : null}
          </div>
        </header>
        <SessionBanner />
        <div className="forma-content">{children}</div>
      </main>
    </div>
  );
}
