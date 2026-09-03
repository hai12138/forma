import { Navigate } from 'react-router-dom';

import { useFormaSession } from '@/hooks/use-forma-session';

import './auth.css';

export function OnboardingPage() {
  const { state, bootstrap, refresh, error } = useFormaSession();

  if (state === 'unauthenticated') {
    return <Navigate to="/login" replace />;
  }
  if (state === 'ready') {
    return <Navigate to="/" replace />;
  }
  if (state === 'loading') {
    return (
      <div className="forma-auth-page" data-testid="forma-loading">
        <div className="forma-auth-card">加载中…</div>
      </div>
    );
  }

  return (
    <div className="forma-auth-page" data-testid="forma-onboarding-page">
      <div className="forma-auth-card">
        <div className="forma-auth-brand">
          <span className="forma-brand-mark">F</span>
          <div>
            <strong>Forma</strong>
            <div className="forma-brand-sub">Business-to-Agent</div>
          </div>
        </div>
        <h1>欢迎使用 Forma</h1>
        <p className="forma-muted">创建第一个 Workspace，开始构建业务到 Agent 的数据与契约。</p>
        {error ? (
          <div className="forma-auth-error" role="alert">
            无法创建 Workspace，请重试。
          </div>
        ) : null}
        <button
          className="forma-btn forma-btn-primary"
          type="button"
          data-testid="onboarding-bootstrap"
          onClick={() => {
            void (async () => {
              await bootstrap();
              await refresh();
            })();
          }}
        >
          创建第一个 Workspace
        </button>
      </div>
    </div>
  );
}
