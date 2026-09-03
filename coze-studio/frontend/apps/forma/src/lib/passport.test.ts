import { afterEach, describe, expect, it, vi } from 'vitest';

import { changePassword, passportLogin } from './passport';

describe('passportLogin', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('posts account/password to Forma login and never returns password', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => '{"code":0,"data":{"password_change_required":false}}',
    });
    vi.stubGlobal('fetch', fetchMock);
    const result = await passportLogin('admin', 'SecretPass1!');
    expect(result.ok).toBe(true);
    expect(result.password_change_required).toBe(false);
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/forma/v1/auth/login',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
      }),
    );
    const body = JSON.parse(fetchMock.mock.calls[0][1].body as string);
    expect(body.account).toBe('admin');
    expect(body.password).toBe('SecretPass1!');
    expect(JSON.stringify(result)).not.toContain('SecretPass1!');
  });

  it('surfaces password_change_required from login response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        text: async () => '{"code":0,"data":{"password_change_required":true}}',
      }),
    );
    const result = await passportLogin('admin', 'admin123');
    expect(result.ok).toBe(true);
    expect(result.password_change_required).toBe(true);
    expect(JSON.stringify(result)).not.toContain('admin123');
  });

  it('returns sanitized error on failure', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        text: async () => 'raw driver stack password=leak',
      }),
    );
    const result = await passportLogin('a@b.com', 'x');
    expect(result.ok).toBe(false);
    expect(result.message).toContain('账号或密码不正确');
    expect(result.message).not.toContain('password');
    expect(result.message).not.toContain('leak');
  });
});

describe('changePassword', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('posts current/new password to Forma change-password endpoint', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => '{"code":0,"data":{"password_changed":true}}',
    });
    vi.stubGlobal('fetch', fetchMock);
    const result = await changePassword('OldPass123!', 'NewPass123!');
    expect(result.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/forma/v1/auth/change-password',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
      }),
    );
    const body = JSON.parse(fetchMock.mock.calls[0][1].body as string);
    expect(body.current_password).toBe('OldPass123!');
    expect(body.new_password).toBe('NewPass123!');
    expect(JSON.stringify(result)).not.toContain('OldPass123!');
    expect(JSON.stringify(result)).not.toContain('NewPass123!');
  });
});
