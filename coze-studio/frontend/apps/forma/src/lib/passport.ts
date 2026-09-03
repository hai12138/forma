/**
 * Coze Passport helpers used by Forma Login / Logout.
 * Session SoT remains Coze SessionAuth (HttpOnly session_key cookie).
 * Never log passwords or store tokens in localStorage/sessionStorage.
 */

export type PassportResult = {
  ok: boolean;
  message?: string;
};

function sanitizedAuthError(status: number, bodyText: string): string {
  if (status === 401 || status === 403) {
    return '账号或密码不正确，请重试。';
  }
  if (status >= 500) {
    return '登录服务暂时不可用，请稍后重试。';
  }
  // Never echo raw body (may contain internals).
  void bodyText;
  return '登录失败，请检查账号信息后重试。';
}

export async function passportLogin(
  account: string,
  password: string,
): Promise<PassportResult & { password_change_required?: boolean }> {
  const res = await fetch('/api/forma/v1/auth/login', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify({ account, password }),
  });
  const text = await res.text();
  if (!res.ok) {
    return { ok: false, message: sanitizedAuthError(res.status, text) };
  }
  try {
    const data = JSON.parse(text);
    return { ok: true, password_change_required: data?.data?.password_change_required };
  } catch {
    return { ok: true };
  }
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<PassportResult> {
  const res = await fetch('/api/forma/v1/auth/change-password', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
  const text = await res.text();
  if (!res.ok) {
    try {
      const data = JSON.parse(text);
      return { ok: false, message: data?.msg || '修改密码失败' };
    } catch {
      return { ok: false, message: '修改密码失败' };
    }
  }
  return { ok: true };
}

/**
 * Forma-owned logout facade → POST /api/forma/v1/auth/logout
 * (revokes Coze session + expires HttpOnly cookie; does not call Coze passport UI).
 */
export async function passportLogout(): Promise<void> {
  try {
    await fetch('/api/forma/v1/auth/logout', {
      method: 'POST',
      credentials: 'include',
      headers: { Accept: 'application/json' },
      signal: AbortSignal.timeout(8000),
    });
  } catch {
    // Still clear local non-secret state; session may already be gone.
  }
}
