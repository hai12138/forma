import { useMemo, useState, type FormEvent } from 'react';
import { Navigate, useNavigate, useSearchParams } from 'react-router-dom';

import { useFormaSession } from '@/hooks/use-forma-session';
import { passportLogin } from '@/lib/passport';
import { safeReturnTo } from '@/lib/return-to';

import './auth.css';

export function LoginPage() {
  const { state, refresh } = useFormaSession();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const returnTo = useMemo(() => safeReturnTo(params.get('returnTo')), [params]);
  const expired = params.get('expired') === '1';

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(
    expired ? '登录已过期，请重新登录。' : null,
  );
  const [busy, setBusy] = useState(false);

  if (state === 'ready') {
    return <Navigate to={returnTo} replace />;
  }
  if (state === 'empty' || state === 'authenticated_no_tenant') {
    return <Navigate to="/onboarding" replace />;
  }

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (busy) return;
    setBusy(true);
    setError(null);
    try {
      const result = await passportLogin(email.trim(), password);
      setPassword('');
      if (!result.ok) {
        setError(result.message || '登录失败，请重试。');
        return;
      }
      const next = await refresh();
      if (next === 'empty' || next === 'authenticated_no_tenant') {
        navigate('/onboarding', { replace: true });
        return;
      }
      if (next === 'ready') {
        navigate(returnTo, { replace: true });
        return;
      }
      if (next === 'unauthenticated') {
        setError('登录失败，请重试。');
        return;
      }
      navigate(returnTo, { replace: true });
    } catch {
      setPassword('');
      setError('登录失败，请重试。');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="forma-auth-page" data-testid="forma-login-page">
      <div className="forma-auth-card">
        <div className="forma-auth-brand">
          <span className="forma-brand-mark">F</span>
          <div>
            <strong>Forma</strong>
            <div className="forma-brand-sub">Business-to-Agent</div>
          </div>
        </div>
        <h1>请登录 Forma</h1>
        <p className="forma-muted">使用工作账号进入业务到 Agent 工作台。</p>
        {error ? (
          <div className="forma-auth-error" role="alert" data-testid="login-error">
            {error}
          </div>
        ) : null}
        <form onSubmit={e => void onSubmit(e)} data-testid="login-form">
          <label htmlFor="forma-login-email">邮箱</label>
          <input
            id="forma-login-email"
            name="email"
            type="email"
            autoComplete="username"
            required
            value={email}
            onChange={e => setEmail(e.target.value)}
            data-testid="login-email"
          />
          <label htmlFor="forma-login-password">密码</label>
          <input
            id="forma-login-password"
            name="password"
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={e => setPassword(e.target.value)}
            data-testid="login-password"
          />
          <button
            className="forma-btn forma-btn-primary"
            type="submit"
            disabled={busy}
            data-testid="login-submit"
          >
            {busy ? '登录中…' : '登录'}
          </button>
        </form>
      </div>
    </div>
  );
}
