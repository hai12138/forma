import { useCallback, useEffect, useState } from 'react';

import { useFormaSession } from '@/hooks/use-forma-session';

import './auth.css';

interface AdminUser {
  principal_id: string;
  account: string;
  display_name: string;
  status: string;
  platform_role: string;
  workspaces: Array<{ tenant_id: string; tenant_name: string; role: string }>;
  password_change_required: boolean;
  created_at: string;
}

interface CreateUserForm {
  account: string;
  display_name: string;
  password: string;
  tenant_id: string;
  tenant_role: string;
}

export function AdminUsersPage() {
  const { currentTenant } = useFormaSession();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState<CreateUserForm>({
    account: '',
    display_name: '',
    password: '',
    tenant_id: '',
    tenant_role: 'MEMBER',
  });
  const [createError, setCreateError] = useState<string | null>(null);
  const [createBusy, setCreateBusy] = useState(false);
  const [initialPassword, setInitialPassword] = useState<string | null>(null);
  const [resetForm, setResetForm] = useState<{ principalId: string; password: string } | null>(null);
  const [actionBusy, setActionBusy] = useState<string | null>(null);

  const loadUsers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch('/api/forma/v1/admin/users', {
        credentials: 'include',
        headers: { Accept: 'application/json' },
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data?.msg || 'Failed to load users');
      setUsers(data.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadUsers();
  }, [loadUsers]);

  const handleCreate = async () => {
    setCreateBusy(true);
    setCreateError(null);
    try {
      const body: Record<string, string> = {
        account: createForm.account,
        display_name: createForm.display_name || createForm.account,
        password: createForm.password,
      };
      if (createForm.tenant_id) {
        body.tenant_id = createForm.tenant_id;
        body.tenant_role = createForm.tenant_role;
      }
      const res = await fetch('/api/forma/v1/admin/users', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify(body),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data?.msg || 'Failed to create user');
      setInitialPassword(data.data?.initial_password || null);
      setShowCreate(false);
      setCreateForm({ account: '', display_name: '', password: '', tenant_id: '', tenant_role: 'MEMBER' });
      await loadUsers();
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create user');
    } finally {
      setCreateBusy(false);
    }
  };

  const handleDisable = async (principalId: string) => {
    setActionBusy(principalId);
    try {
      await fetch(`/api/forma/v1/admin/users/${principalId}/disable`, {
        method: 'POST',
        credentials: 'include',
        headers: { Accept: 'application/json' },
      });
      await loadUsers();
    } finally {
      setActionBusy(null);
    }
  };

  const handleEnable = async (principalId: string) => {
    setActionBusy(principalId);
    try {
      await fetch(`/api/forma/v1/admin/users/${principalId}/enable`, {
        method: 'POST',
        credentials: 'include',
        headers: { Accept: 'application/json' },
      });
      await loadUsers();
    } finally {
      setActionBusy(null);
    }
  };

  const handleResetPassword = async () => {
    if (!resetForm) return;
    setActionBusy(resetForm.principalId);
    try {
      const res = await fetch(`/api/forma/v1/admin/users/${resetForm.principalId}/reset-password`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify({ new_password: resetForm.password }),
      });
      if (!res.ok) {
        const data = await res.json();
        alert(data?.msg || 'Reset failed');
      } else {
        setResetForm(null);
        await loadUsers();
      }
    } finally {
      setActionBusy(null);
    }
  };

  return (
    <div data-testid="admin-users-page" style={{ padding: '24px' }}>
      <div
        style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}
      >
        <h2 style={{ margin: 0 }}>用户管理</h2>
        <button
          className="forma-btn forma-btn-primary"
          onClick={() => setShowCreate(true)}
          data-testid="create-user-btn"
        >
          创建用户
        </button>
      </div>

      {initialPassword ? (
        <div
          style={{ background: '#e8f5e9', padding: '12px 16px', borderRadius: '6px', marginBottom: '16px' }}
          data-testid="initial-password-display"
        >
          <strong>用户创建成功</strong>
          <span style={{ marginLeft: '12px' }}>
            初始密码：<code data-testid="initial-password-value">{initialPassword}</code>
          </span>
          <button style={{ marginLeft: '12px' }} onClick={() => setInitialPassword(null)}>
            关闭
          </button>
        </div>
      ) : null}

      {error ? <div className="forma-auth-error">{error}</div> : null}

      {showCreate ? (
        <div
          style={{ background: '#f5f5f5', padding: '16px', borderRadius: '8px', marginBottom: '16px' }}
          data-testid="create-user-dialog"
        >
          <h3 style={{ marginTop: 0 }}>创建用户</h3>
          {createError ? <div className="forma-auth-error">{createError}</div> : null}
          <div style={{ display: 'grid', gap: '8px', maxWidth: '400px' }}>
            <label>
              账号{' '}
              <input
                type="text"
                value={createForm.account}
                onChange={e => setCreateForm(f => ({ ...f, account: e.target.value }))}
                data-testid="create-user-account"
                placeholder="用户名或邮箱"
              />
            </label>
            <label>
              显示名称{' '}
              <input
                type="text"
                value={createForm.display_name}
                onChange={e => setCreateForm(f => ({ ...f, display_name: e.target.value }))}
                data-testid="create-user-display-name"
              />
            </label>
            <label>
              初始密码{' '}
              <input
                type="password"
                value={createForm.password}
                onChange={e => setCreateForm(f => ({ ...f, password: e.target.value }))}
                data-testid="create-user-password"
                placeholder="至少 8 个字符"
              />
            </label>
            <label>
              Workspace
              <select
                value={createForm.tenant_id}
                onChange={e => setCreateForm(f => ({ ...f, tenant_id: e.target.value }))}
                data-testid="create-user-tenant"
              >
                <option value="">不分配</option>
                {currentTenant ? (
                  <option value={currentTenant.tenant_id}>
                    {currentTenant.display_name || currentTenant.name}
                  </option>
                ) : null}
              </select>
            </label>
            <label>
              角色
              <select
                value={createForm.tenant_role}
                onChange={e => setCreateForm(f => ({ ...f, tenant_role: e.target.value }))}
                data-testid="create-user-role"
              >
                <option value="OWNER">OWNER</option>
                <option value="ADMIN">ADMIN</option>
                <option value="MEMBER">MEMBER</option>
              </select>
            </label>
            <div style={{ display: 'flex', gap: '8px', marginTop: '8px' }}>
              <button
                className="forma-btn forma-btn-primary"
                onClick={() => void handleCreate()}
                disabled={createBusy || !createForm.account || !createForm.password}
                data-testid="create-user-submit"
              >
                {createBusy ? '创建中…' : '创建'}
              </button>
              <button
                className="forma-btn"
                onClick={() => {
                  setShowCreate(false);
                  setCreateError(null);
                }}
              >
                取消
              </button>
            </div>
          </div>
        </div>
      ) : null}

      {resetForm ? (
        <div
          style={{ background: '#fff3e0', padding: '16px', borderRadius: '8px', marginBottom: '16px' }}
          data-testid="reset-password-dialog"
        >
          <h3 style={{ marginTop: 0 }}>重置密码</h3>
          <label>
            新密码{' '}
            <input
              type="password"
              value={resetForm.password}
              onChange={e => setResetForm(f => (f ? { ...f, password: e.target.value } : null))}
              data-testid="reset-password-input"
              placeholder="至少 8 个字符"
            />
          </label>
          <div style={{ display: 'flex', gap: '8px', marginTop: '8px' }}>
            <button
              className="forma-btn forma-btn-primary"
              onClick={() => void handleResetPassword()}
              disabled={!resetForm.password || resetForm.password.length < 8}
              data-testid="reset-password-submit"
            >
              重置
            </button>
            <button className="forma-btn" onClick={() => setResetForm(null)}>
              取消
            </button>
          </div>
        </div>
      ) : null}

      {loading ? (
        <div>加载中…</div>
      ) : (
        <table style={{ width: '100%', borderCollapse: 'collapse' }} data-testid="users-table">
          <thead>
            <tr style={{ borderBottom: '2px solid #e0e0e0', textAlign: 'left' }}>
              <th style={{ padding: '8px' }}>账号</th>
              <th style={{ padding: '8px' }}>显示名称</th>
              <th style={{ padding: '8px' }}>状态</th>
              <th style={{ padding: '8px' }}>平台角色</th>
              <th style={{ padding: '8px' }}>所属 Workspace</th>
              <th style={{ padding: '8px' }}>创建时间</th>
              <th style={{ padding: '8px' }}>操作</th>
            </tr>
          </thead>
          <tbody>
            {users.map(u => (
              <tr key={u.principal_id} style={{ borderBottom: '1px solid #eee' }} data-testid={`user-row-${u.account}`}>
                <td style={{ padding: '8px' }} data-testid="user-account">
                  {u.account}
                </td>
                <td style={{ padding: '8px' }}>{u.display_name}</td>
                <td style={{ padding: '8px' }}>
                  <span
                    data-testid="user-status"
                    style={{
                      color: u.status === 'ACTIVE' ? '#2e7d32' : '#c62828',
                      fontWeight: 600,
                    }}
                  >
                    {u.status}
                  </span>
                  {u.password_change_required ? (
                    <span style={{ marginLeft: '4px', fontSize: '12px', color: '#ef6c00' }}>需改密</span>
                  ) : null}
                </td>
                <td style={{ padding: '8px' }}>
                  <span data-testid="user-platform-role">{u.platform_role}</span>
                </td>
                <td style={{ padding: '8px' }}>
                  {u.workspaces.map(w => (
                    <div key={w.tenant_id}>
                      {w.tenant_name} ({w.role})
                    </div>
                  ))}
                </td>
                <td style={{ padding: '8px' }}>{new Date(u.created_at).toLocaleString()}</td>
                <td style={{ padding: '8px' }}>
                  <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
                    {u.status === 'ACTIVE' ? (
                      <button
                        className="forma-btn"
                        style={{ fontSize: '12px', padding: '2px 8px' }}
                        disabled={actionBusy === u.principal_id}
                        onClick={() => void handleDisable(u.principal_id)}
                        data-testid="disable-user-btn"
                      >
                        禁用
                      </button>
                    ) : (
                      <button
                        className="forma-btn"
                        style={{ fontSize: '12px', padding: '2px 8px' }}
                        disabled={actionBusy === u.principal_id}
                        onClick={() => void handleEnable(u.principal_id)}
                        data-testid="enable-user-btn"
                      >
                        启用
                      </button>
                    )}
                    <button
                      className="forma-btn"
                      style={{ fontSize: '12px', padding: '2px 8px' }}
                      onClick={() => setResetForm({ principalId: u.principal_id, password: '' })}
                      data-testid="reset-password-btn"
                    >
                      重置密码
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
