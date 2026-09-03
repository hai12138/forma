import { useState, type FormEvent } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';

import { useFormaSession } from '@/hooks/use-forma-session';
import { changePassword } from '@/lib/passport';

import './auth.css';

export function ChangePasswordPage() {
  const { state, refresh } = useFormaSession();
  const navigate = useNavigate();

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  if (state === 'unauthenticated') {
    return <Navigate to="/login" replace />;
  }

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (busy) return;

    if (newPassword !== confirmPassword) {
      setError('两次输入的新密码不一致');
      return;
    }
    if (newPassword.length < 8) {
      setError('新密码至少需要 8 个字符');
      return;
    }
    if (newPassword === 'admin123') {
      setError('不能使用初始密码 admin123 作为新密码');
      return;
    }

    setBusy(true);
    setError(null);
    try {
      const result = await changePassword(currentPassword, newPassword);
      if (!result.ok) {
        setError(result.message || '修改密码失败');
        return;
      }
      const next = await refresh();
      if (next === 'empty' || next === 'authenticated_no_tenant') {
        navigate('/onboarding', { replace: true });
      } else {
        navigate('/', { replace: true });
      }
    } catch {
      setError('修改密码失败，请重试');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="forma-auth-page" data-testid="forma-change-password-page">
      <div className="forma-auth-card">
        <div className="forma-auth-brand">
          <span className="forma-brand-mark">F</span>
          <div>
            <strong>Forma</strong>
            <div className="forma-brand-sub">Business-to-Agent</div>
          </div>
        </div>
        <h1>修改密码</h1>
        <p className="forma-muted">首次登录需要修改初始密码后才能继续使用。</p>
        {error ? (
          <div className="forma-auth-error" role="alert" data-testid="change-password-error">
            {error}
          </div>
        ) : null}
        <form onSubmit={e => void onSubmit(e)} data-testid="change-password-form">
          <label htmlFor="forma-current-password">当前密码</label>
          <input
            id="forma-current-password"
            type="password"
            autoComplete="current-password"
            required
            value={currentPassword}
            onChange={e => setCurrentPassword(e.target.value)}
            data-testid="current-password"
          />
          <label htmlFor="forma-new-password">新密码</label>
          <input
            id="forma-new-password"
            type="password"
            autoComplete="new-password"
            required
            minLength={8}
            placeholder="至少 8 个字符"
            value={newPassword}
            onChange={e => setNewPassword(e.target.value)}
            data-testid="new-password"
          />
          <label htmlFor="forma-confirm-password">确认新密码</label>
          <input
            id="forma-confirm-password"
            type="password"
            autoComplete="new-password"
            required
            value={confirmPassword}
            onChange={e => setConfirmPassword(e.target.value)}
            data-testid="confirm-password"
          />
          <button
            className="forma-btn forma-btn-primary"
            type="submit"
            disabled={busy}
            data-testid="change-password-submit"
          >
            {busy ? '修改中…' : '修改密码'}
          </button>
        </form>
      </div>
    </div>
  );
}
