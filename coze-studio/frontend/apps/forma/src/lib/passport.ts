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
    return '邮箱或密码不正确，请重试。';
  }
  if (status >= 500) {
    return '登录服务暂时不可用，请稍后重试。';
  }
  // Never echo raw body (may contain internals).
  void bodyText;
  return '登录失败，请检查账号信息后重试。';
}

export async function passportLogin(email: string, password: string): Promise<PassportResult> {
  const res = await fetch('/api/passport/web/email/login/', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  const text = await res.text();
  if (!res.ok) {
    return { ok: false, message: sanitizedAuthError(res.status, text) };
  }
  return { ok: true };
}

export async function passportLogout(): Promise<void> {
  try {
    await fetch('/api/passport/web/logout/', {
      method: 'GET',
      credentials: 'include',
      headers: { Accept: 'application/json' },
      signal: AbortSignal.timeout(8000),
    });
  } catch {
    // Still clear local non-secret state; session may already be gone.
  }
}
