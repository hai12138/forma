/**
 * Forma S1-G1 live E2E against a running Coze/Forma backend.
 *
 * Requires:
 *   FORMA_LIVE_E2E=1
 *   FORMA_LIVE_BASE_URL (default http://127.0.0.1:8888)
 *   FORMA_LIVE_EMAIL / FORMA_LIVE_PASSWORD (optional; auto-register if missing)
 *
 * Exercises real SessionAuthMW + Forma tenant APIs (no mock auth).
 */
import assert from 'node:assert/strict';
import test from 'node:test';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const baseUrl = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const email =
  process.env.FORMA_LIVE_EMAIL ||
  `forma-e2e-${Date.now()}@example.com`;
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';

function skipIfDisabled() {
  if (!enabled) {
    // Soft-skip: document as EXTERNAL when server unavailable
    return true;
  }
  return false;
}

async function request(method, path, { body, cookie, tenantId } = {}) {
  const headers = { Accept: 'application/json', 'X-Request-ID': `live-${Date.now()}` };
  if (cookie) headers.Cookie = cookie;
  if (tenantId) headers['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  const res = await fetch(`${baseUrl}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await res.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    json = { raw: text };
  }
  const setCookie = res.headers.getSetCookie?.() || [];
  return { status: res.status, json, setCookie, headers: res.headers };
}

function extractSessionCookie(setCookie) {
  for (const c of setCookie) {
    const m = /session_key=([^;]+)/.exec(c);
    if (m) return `session_key=${m[1]}`;
  }
  return '';
}

test('LIVE: Coze register/login → Forma principal bootstrap → tenant switch → forged/suspended', async (t) => {
  if (skipIfDisabled()) {
    t.skip('FORMA_LIVE_E2E!=1 — set env and start Coze/Forma backend to run live gates');
    return;
  }

  // Health must be reachable
  const health = await request('GET', '/api/forma/v1/health');
  assert.equal(health.status, 200, `backend not reachable: ${JSON.stringify(health.json)}`);

  // Register (real Coze passport) — cookie may already be set; login as fallback.
  const reg = await request('POST', '/api/passport/web/email/register/v2/', {
    body: { email, password },
  });
  let cookie = extractSessionCookie(reg.setCookie);
  if (!cookie) {
    const login = await request('POST', '/api/passport/web/email/login/', {
      body: { email, password },
    });
    assert.ok(
      login.status < 400 && (login.json?.code === 0 || login.json?.code === undefined),
      `login failed: ${login.status} ${JSON.stringify(login.json)}`,
    );
    cookie = extractSessionCookie(login.setCookie);
  }
  assert.ok(cookie, `session_key cookie missing; reg=${JSON.stringify(reg.json)}`);

  // /me before bootstrap
  let me = await request('GET', '/api/forma/v1/me', { cookie });
  assert.equal(me.status, 200, JSON.stringify(me.json));
  assert.ok(me.json?.data?.principal?.principal_id, 'principal missing');

  // bootstrap
  const boot = await request('POST', '/api/forma/v1/bootstrap', {
    cookie,
    body: {},
  });
  assert.ok(boot.status < 400, `bootstrap failed: ${JSON.stringify(boot.json)}`);

  me = await request('GET', '/api/forma/v1/me', { cookie });
  assert.equal(me.status, 200);
  const principalId = me.json.data.principal.principal_id;
  assert.ok(principalId);
  const tenants = me.json.data.tenants || [];
  assert.ok(tenants.length >= 1, 'expected at least one tenant after bootstrap');
  const tenantA = tenants[0].tenant_id;

  // create second tenant
  const created = await request('POST', '/api/forma/v1/tenants', {
    cookie,
    body: { name: `E2E Tenant B ${Date.now()}`, display_name: 'E2E Tenant B' },
  });
  assert.ok(created.status < 400, JSON.stringify(created.json));
  const tenantB = created.json.data.tenant_id || created.json.data.TenantID;
  assert.ok(tenantB);

  // switch A → counts
  const countsA = await request('GET', '/api/forma/v1/assets/counts', {
    cookie,
    tenantId: tenantA,
  });
  assert.equal(countsA.status, 200, JSON.stringify(countsA.json));

  // switch B → counts
  const countsB = await request('GET', '/api/forma/v1/assets/counts', {
    cookie,
    tenantId: tenantB,
  });
  assert.equal(countsB.status, 200, JSON.stringify(countsB.json));

  // forged tenant
  const forged = await request('GET', '/api/forma/v1/assets/counts', {
    cookie,
    tenantId: 'ten_forged_does_not_exist',
  });
  assert.equal(forged.status, 403);
  const forgedKey = forged.json.error_key || forged.json.msg || '';
  assert.ok(
    String(forgedKey).includes('FORMA_TENANT_FORBIDDEN') ||
      String(forgedKey).includes('FORMA_TENANT_NOT_FOUND') ||
      String(forged.json.msg || '').includes('FORMA_'),
    `unexpected forged response: ${JSON.stringify(forged.json)}`,
  );

  // suspend tenant B then hit counts
  const getB = await request('GET', `/api/forma/v1/tenants/${tenantB}`, {
    cookie,
    tenantId: tenantB,
  });
  assert.equal(getB.status, 200);
  const rev = getB.json.data.revision ?? getB.json.data.Revision;
  assert.ok(rev > 0, `revision missing: ${JSON.stringify(getB.json)}`);
  const suspend = await request('PATCH', `/api/forma/v1/tenants/${tenantB}`, {
    cookie,
    tenantId: tenantB,
    body: { status: 'SUSPENDED', expected_revision: rev },
  });
  assert.ok(suspend.status < 400, JSON.stringify(suspend.json));

  const suspendedCounts = await request('GET', '/api/forma/v1/assets/counts', {
    cookie,
    tenantId: tenantB,
  });
  assert.equal(suspendedCounts.status, 403);
  assert.ok(
    String(suspendedCounts.json.error_key || suspendedCounts.json.msg || '').includes(
      'FORMA_TENANT_SUSPENDED',
    ),
    JSON.stringify(suspendedCounts.json),
  );

  // restore
  const restoreRev =
    suspend.json.data?.revision ?? suspend.json.data?.Revision ?? rev + 1;
  const restore = await request('PATCH', `/api/forma/v1/tenants/${tenantB}`, {
    cookie,
    tenantId: tenantB,
    body: { status: 'ACTIVE', expected_revision: restoreRev },
  });
  assert.ok(restore.status < 400, JSON.stringify(restore.json));

  // space bind: bootstrap should have bound the user's real personal Coze Space
  // (via Coze Session + GetUserSpaceList — not a raw forma_tenant_space_ref insert).
  const spaces = await request('GET', `/api/forma/v1/tenants/${tenantA}/spaces`, {
    cookie,
    tenantId: tenantA,
  });
  assert.equal(spaces.status, 200, JSON.stringify(spaces.json));
  const spaceList = Array.isArray(spaces.json.data) ? spaces.json.data : [];
  assert.ok(
    spaceList.length >= 1,
    `expected bootstrap-bound personal space: ${JSON.stringify(spaces.json)}`,
  );
  const spaceId = spaceList[0].coze_space_id;
  assert.equal(typeof spaceId, 'string', `coze_space_id must be string: ${JSON.stringify(spaceList[0])}`);
  assert.match(String(spaceId), /^[1-9][0-9]*$/);
  // Precision: string must round-trip without JS number coercion
  assert.equal(String(BigInt(spaceId)), spaceId, 'coze_space_id lost precision');

  const meCoze = me.json.data.coze_user_id ?? me.json.data.principal?.coze_user_id;
  assert.equal(typeof meCoze, 'string', `coze_user_id must be string: ${meCoze}`);
  assert.equal(String(BigInt(meCoze)), meCoze);

  // inaccessible / forged Coze Space must deny (string contract)
  const badSpace = await request('POST', `/api/forma/v1/tenants/${tenantA}/spaces`, {
    cookie,
    tenantId: tenantA,
    body: { coze_space_id: '999999999999', purpose: 'DEFAULT' },
  });
  assert.ok(badSpace.status === 403 || badSpace.status === 404, JSON.stringify(badSpace.json));

  // JSON number must be rejected by BindSpace contract
  const numberRejected = await request('POST', `/api/forma/v1/tenants/${tenantA}/spaces`, {
    cookie,
    tenantId: tenantA,
    body: undefined,
  });
  // send raw number body via fetch
  {
    const res = await fetch(`${baseUrl}/api/forma/v1/tenants/${tenantA}/spaces`, {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        Cookie: cookie,
        'X-Forma-Tenant': tenantA,
        'X-Request-ID': `live-num-${Date.now()}`,
      },
      body: JSON.stringify({ coze_space_id: 7563957783431741441, purpose: 'DEFAULT' }),
    });
    assert.ok(res.status >= 400, `JSON number coze_space_id must fail, got ${res.status}`);
  }
  void numberRejected;
});
