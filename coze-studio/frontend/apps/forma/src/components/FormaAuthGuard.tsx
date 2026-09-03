import { Navigate, useLocation } from 'react-router-dom';

import { useFormaSession } from '@/hooks/use-forma-session';
import { encodeReturnTo } from '@/lib/return-to';

import '../pages/auth.css';

/**
 * FormaAuthGuard — product auth UX over Coze SessionAuth SoT.
 * Never renders AppShell / protected children while unauthenticated.
 */
export function FormaAuthGuard({ children }: { children: React.ReactNode }) {
  const { state } = useFormaSession();
  const location = useLocation();

  if (state === 'loading') {
    return (
      <div className="forma-auth-page" data-testid="forma-auth-loading">
        <div className="forma-auth-card">加载中…</div>
      </div>
    );
  }

  if (state === 'unauthenticated') {
    const returnTo = encodeReturnTo(`${location.pathname}${location.search}`);
    return <Navigate to={`/login?returnTo=${returnTo}`} replace />;
  }

  if (state === 'empty' || state === 'authenticated_no_tenant') {
    return <Navigate to="/onboarding" replace />;
  }

  if (state === 'suspended') {
    return (
      <div className="forma-auth-page" data-testid="forma-auth-suspended">
        <div className="forma-auth-card">
          <h1>Workspace 已暂停</h1>
          <p className="forma-muted">当前 Workspace 不可用，请联系管理员。</p>
        </div>
      </div>
    );
  }

  if (state === 'forbidden') {
    return (
      <div className="forma-auth-page" data-testid="forma-auth-forbidden">
        <div className="forma-auth-card">
          <h1>无权访问</h1>
          <p className="forma-muted">当前身份无权访问所选 Workspace。</p>
        </div>
      </div>
    );
  }

  if (state === 'network_error') {
    return (
      <div className="forma-auth-page" data-testid="forma-auth-network">
        <div className="forma-auth-card">
          <h1>网络异常</h1>
          <p className="forma-muted">无法连接 Forma 服务，请稍后重试。</p>
        </div>
      </div>
    );
  }

  // ready
  return <>{children}</>;
}
