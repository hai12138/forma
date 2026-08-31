import { NavLink } from 'react-router-dom';

import { navigation } from '@/lib/navigation';
import { useFormaBaseline } from '@/hooks/use-forma-baseline';

import './shell.css';

export function AppShell({ children }: { children: React.ReactNode }) {
  const { loading, error, baseline } = useFormaBaseline();

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
            <div className="forma-env">Dev · Forma Product Shell</div>
          </div>
          <div className="forma-topbar-meta">
            {loading && <span>Loading baseline…</span>}
            {!loading && error && <span className="forma-error">{error}</span>}
            {!loading && baseline && (
              <span>
                {baseline.forma_version} · {baseline.runtime_foundation}
              </span>
            )}
          </div>
        </header>
        <div className="forma-content">{children}</div>
      </main>
    </div>
  );
}
