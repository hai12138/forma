/**
 * Forma S2 live API gates (Business Model) against forma-live-harness.
 *
 *   FORMA_LIVE_E2E=1
 *   FORMA_LIVE_BASE_URL=http://127.0.0.1:8888
 *   node --test scripts/forma/live-business-e2e.mjs
 */
import assert from 'node:assert/strict';
import test from 'node:test';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const baseUrl = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(
  /\/$/,
  '',
);
const email =
  process.env.FORMA_LIVE_EMAIL ||
  `forma-s2-${Date.now()}@example.com`;
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';

function jar() {
  const cookies = new Map();
  return {
    store(res) {
      const raw = res.headers.getSetCookie?.() || [];
      for (const c of raw) {
        const [pair] = c.split(';');
        const i = pair.indexOf('=');
        if (i > 0) cookies.set(pair.slice(0, i), pair.slice(i + 1));
      }
    },
    header() {
      return [...cookies.entries()].map(([k, v]) => `${k}=${v}`).join('; ');
    },
  };
}

async function api(path, { method = 'GET', body, tenantId, cookies } = {}) {
  const headers = { Accept: 'application/json', 'X-Request-ID': `s2-${Date.now()}` };
  if (cookies) headers.Cookie = cookies.header();
  if (tenantId) headers['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  const res = await fetch(`${baseUrl}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (cookies) cookies.store(res);
  const json = await res.json().catch(() => ({}));
  return { res, json };
}

const workOrderSeed = {
  schema_version: '2.0',
  nodes: [
    {
      id: 'actor_reporter',
      type: 'ACTOR',
      name: '报修人',
      source_marker: 'MANUAL_MODIFIED',
    },
    {
      id: 'obj_work_order',
      type: 'BUSINESS_OBJECT',
      name: '维修工单',
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
  edges: [
    {
      id: 'e1',
      source: 'actor_reporter',
      target: 'obj_work_order',
      type: 'CREATES',
      label: '创建',
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
  rules: [],
  states: [
    {
      id: 'st_pending',
      object_ref: 'obj_work_order',
      name: '待受理',
      initial: true,
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
};

test('S2 live Business Model gates', async (t) => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }

  const cookies = jar();

  await t.test('register/login', async () => {
    let r = await api('/api/passport/web/email/register/v2/', {
      method: 'POST',
      body: { email, password },
      cookies,
    });
    if (r.res.status >= 400) {
      r = await api('/api/passport/web/email/login/', {
        method: 'POST',
        body: { email, password },
        cookies,
      });
    }
    assert.ok(r.res.ok || r.json?.code === 0, JSON.stringify(r.json));
  });

  let tenantId;
  await t.test('bootstrap tenant', async () => {
    const r = await api('/api/forma/v1/bootstrap', { method: 'POST', body: {}, cookies });
    assert.equal(r.res.status, 200, JSON.stringify(r.json));
    tenantId = r.json.data.tenant.tenant_id;
    assert.ok(tenantId);
  });

  let businessId;
  let revision = 1;
  await t.test('create business 维修工单', async () => {
    const r = await api('/api/forma/v1/businesses', {
      method: 'POST',
      tenantId,
      cookies,
      body: {
        name: '维修工单',
        semantic_model: workOrderSeed,
        change_summary: 'seed work order',
      },
    });
    assert.equal(r.res.status, 200, JSON.stringify(r.json));
    businessId = r.json.data.business_id;
    revision = r.json.data.current_revision;
    assert.equal(revision, 1);
  });

  let layoutRevision = 1;
  await t.test('layout save does not bump semantic revision', async () => {
    const layoutGet = await api(`/api/forma/v1/businesses/${businessId}/layout`, {
      tenantId,
      cookies,
    });
    assert.equal(layoutGet.res.status, 200, JSON.stringify(layoutGet.json));
    layoutRevision = layoutGet.json.data.layout_revision;
    const layout = {
      ...layoutGet.json.data.layout,
      node_positions: {
        ...(layoutGet.json.data.layout.node_positions || {}),
        actor_reporter: { x: 120, y: 80 },
      },
      zoom: 1.1,
    };
    const put = await api(`/api/forma/v1/businesses/${businessId}/layout`, {
      method: 'PUT',
      tenantId,
      cookies,
      body: {
        expected_layout_revision: layoutRevision,
        based_on_model_revision: revision,
        layout,
      },
    });
    assert.equal(put.res.status, 200, JSON.stringify(put.json));
    layoutRevision = put.json.data.layout_revision;

    const model = await api(`/api/forma/v1/businesses/${businessId}/model`, {
      tenantId,
      cookies,
    });
    assert.equal(model.json.data.current_revision, revision);
  });

  await t.test('semantic save bumps revision', async () => {
    const model = await api(`/api/forma/v1/businesses/${businessId}/model`, {
      tenantId,
      cookies,
    });
    const sm = model.json.data.semantic_model;
    sm.nodes[0].name = '报修客户';
    const put = await api(`/api/forma/v1/businesses/${businessId}/model`, {
      method: 'PUT',
      tenantId,
      cookies,
      body: {
        expected_revision: revision,
        semantic_model: sm,
        change_summary: 'rename reporter',
      },
    });
    assert.equal(put.res.status, 200, JSON.stringify(put.json));
    assert.equal(put.json.data.current_revision, revision + 1);
    assert.equal(put.json.data.no_change || false, false);
    revision = put.json.data.current_revision;
  });

  await t.test('diff revisions', async () => {
    const r = await api(`/api/forma/v1/businesses/${businessId}/diff?from=1&to=${revision}`, {
      tenantId,
      cookies,
    });
    assert.equal(r.res.status, 200, JSON.stringify(r.json));
    assert.ok(r.json.data.diff.nodes.modified.includes('actor_reporter'));
  });

  await t.test('tenant isolation', async () => {
    const cookiesB = jar();
    const emailB = `forma-s2b-${Date.now()}@example.com`;
    await api('/api/passport/web/email/register/v2/', {
      method: 'POST',
      body: { email: emailB, password },
      cookies: cookiesB,
    });
    const boot = await api('/api/forma/v1/bootstrap', {
      method: 'POST',
      body: {},
      cookies: cookiesB,
    });
    const tenantB = boot.json.data.tenant.tenant_id;
    const denied = await api(`/api/forma/v1/businesses/${businessId}`, {
      tenantId: tenantB,
      cookies: cookiesB,
    });
    assert.ok(denied.res.status === 404 || denied.json.error_key === 'FORMA_BUSINESS_NOT_FOUND');
  });
});
