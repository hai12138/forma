import { afterEach, describe, expect, it, vi } from 'vitest';

import { passportLogin } from './passport';

describe('passportLogin', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('posts email/password to Coze passport and never returns password', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => '{"code":0}',
    });
    vi.stubGlobal('fetch', fetchMock);
    const result = await passportLogin('a@b.com', 'SecretPass1!');
    expect(result.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/passport/web/email/login/',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
      }),
    );
    const body = JSON.parse(fetchMock.mock.calls[0][1].body as string);
    expect(body.email).toBe('a@b.com');
    expect(body.password).toBe('SecretPass1!');
    expect(JSON.stringify(result)).not.toContain('SecretPass1!');
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
    expect(result.message).toContain('邮箱或密码不正确');
    expect(result.message).not.toContain('password');
    expect(result.message).not.toContain('leak');
  });
});
