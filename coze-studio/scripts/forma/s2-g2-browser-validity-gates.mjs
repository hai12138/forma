/**
 * FORMA-S2-G2 MANUAL_LIVE_BROWSER_VALIDITY_GATE
 *
 * Real Playwright browser checks that Editor UI cannot produce
 * Backend-INVALID models through normal interactions.
 *
 *   FORMA_LIVE_E2E=1
 *   FORMA_LIVE_BASE_URL=http://127.0.0.1:8888
 *   FORMA_UI_BASE_URL=http://127.0.0.1:3001
 *   node --test scripts/forma/s2-g2-browser-validity-gates.mjs
 */
import assert from 'node:assert/strict';
import { mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import { chromium } from 'playwright';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const baseApi = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const baseUi = (process.env.FORMA_UI_BASE_URL || 'http://127.0.0.1:3001').replace(/\/$/, '');
const email = process.env.FORMA_LIVE_EMAIL || `forma-s2g2-${Date.now()}@example.com`;
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const outDir = join(root, 'forma', 'cursor-results', 's2-g2-ui');

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
    entries(url) {
      return [...cookies.entries()].map(([name, value]) => ({ name, value, url }));
    },
  };
}

async function api(path, { method = 'GET', body, tenantId, cookies } = {}) {
  const headers = { Accept: 'application/json', 'X-Request-ID': `s2g2-${Date.now()}` };
  if (cookies) headers.Cookie = cookies.header();
  if (tenantId) headers['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  const res = await fetch(`${baseApi}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (cookies) cookies.store(res);
  const json = await res.json().catch(() => ({}));
  return { res, json };
}

test('S2-G2 browser validity gates', async (t) => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1 — MANUAL_LIVE_BROWSER_VALIDITY_GATE');
    return;
  }

  mkdirSync(outDir, { recursive: true });
  const cookies = jar();

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

  const boot = await api('/api/forma/v1/bootstrap', { method: 'POST', body: {}, cookies });
  assert.equal(boot.res.status, 200, JSON.stringify(boot.json));
  const tenantId = boot.json.data.tenant.tenant_id;

  // Empty business (no seed model beyond default empty)
  const created = await api('/api/forma/v1/businesses', {
    method: 'POST',
    tenantId,
    cookies,
    body: {
      name: '空业务-S2G2',
      semantic_model: {
        schema_version: '2.0',
        nodes: [],
        edges: [],
        rules: [],
        states: [],
      },
      change_summary: 'empty',
    },
  });
  assert.equal(created.res.status, 200, JSON.stringify(created.json));
  const businessId = created.json.data.business_id;
  let modelRev = created.json.data.current_revision;

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await context.addCookies([...cookies.entries(baseUi), ...cookies.entries(baseApi)]);
  await context.addInitScript((tid) => {
    try {
      sessionStorage.setItem('forma.selectedTenantId', tid);
    } catch {
      /* ignore */
    }
  }, tenantId);
  const page = await context.newPage();

  const invalidPuts = [];
  page.on('response', async (resp) => {
    if (resp.request().method() === 'PUT' && resp.url().includes('/model')) {
      if (resp.status() >= 400) {
        const body = await resp.json().catch(() => ({}));
        invalidPuts.push({ status: resp.status(), body });
      }
    }
  });

  await page.goto(`${baseUi}/business/${businessId}`, { waitUntil: 'networkidle' });
  await page.waitForSelector('.forma-vme', { timeout: 60000 });

  const readAttrs = async () =>
    page.locator('.forma-vme').evaluate((el) => ({
      modelRevision: Number(el.getAttribute('data-model-revision')),
      semanticDirty: el.getAttribute('data-semantic-dirty') === 'true',
    }));

  await t.test('CASE A: empty model — add state disabled', async () => {
    const btn = page.getByTestId('add-state');
    await btn.waitFor({ state: 'visible' });
    assert.equal(await btn.isDisabled(), true);
    await page.screenshot({ path: join(outDir, '01-empty-add-state-disabled.png') });
  });

  await t.test('CASE B: add node then state with object_ref select', async () => {
    await page.getByRole('button', { name: '＋参与者' }).click();
    await page.waitForTimeout(150);
    assert.equal(await page.getByTestId('add-state').isDisabled(), false);
    await page.getByTestId('add-state').click();
    await page.getByTestId('state-object-ref-select').waitFor({ state: 'visible' });
    const options = await page
      .getByTestId('state-object-ref-select')
      .locator('option')
      .allTextContents();
    assert.ok(options.length >= 1);
    assert.ok(options.every((o) => o.trim().length > 0));
    // no raw free-text object_ref input
    assert.equal(await page.locator('input[name="object_ref"]').count(), 0);
    await page.getByTestId('save-model').click();
    await page.waitForFunction(
      (prev) => {
        const el = document.querySelector('.forma-vme');
        return el && Number(el.getAttribute('data-model-revision')) > prev;
      },
      modelRev,
      { timeout: 30000 },
    );
    const after = await readAttrs();
    assert.ok(after.modelRevision > modelRev);
    modelRev = after.modelRevision;
    await page.screenshot({ path: join(outDir, '02-state-object-ref-select.png') });
  });

  await t.test('CASE C: rule applies_to selector', async () => {
    await page.getByTestId('add-rule').click();
    await page.getByTestId('rule-applies-to').waitFor({ state: 'visible' });
    assert.equal(await page.locator('input[name="applies_to"][type="text"]').count(), 0);
    assert.ok((await page.locator('input[name="applies_to"][type="checkbox"]').count()) >= 1);
    await page.getByTestId('save-model').click();
    await page.waitForFunction(
      (prev) => {
        const el = document.querySelector('.forma-vme');
        return el && Number(el.getAttribute('data-model-revision')) > prev;
      },
      modelRev,
      { timeout: 30000 },
    );
    modelRev = (await readAttrs()).modelRevision;
    await page.screenshot({ path: join(outDir, '03-rule-applies-to.png') });
  });

  await t.test('CASE D: blank node name blocked — no invalid PUT', async () => {
    const before = await readAttrs();
    const node = page.locator('[data-kind="node"]').first();
    await node.dblclick();
    const input = page.getByTestId('node-name-input');
    await input.waitFor({ state: 'visible' });
    await input.fill('   ');
    await input.blur();
    await page.waitForTimeout(200);
    // Either not dirty with blank, or dirty blocked from save — revision must not bump via invalid API
    const save = page.getByTestId('save-model');
    if (await save.isEnabled()) {
      await save.click();
      await page.waitForTimeout(500);
    }
    const after = await readAttrs();
    assert.equal(after.modelRevision, before.modelRevision);
    assert.equal(
      invalidPuts.filter((p) => p.body?.error_key === 'FORMA_BUSINESS_INVALID_MODEL').length,
      0,
    );
    // restore name
    await input.fill('报修人');
    await input.blur();
    await page.screenshot({ path: join(outDir, '04-blank-name-blocked.png') });
  });

  await t.test('CASE E: final save model returns 200', async () => {
    const before = await readAttrs();
    // ensure a real semantic change
    const input = page.getByTestId('node-name-input');
    if (await input.count()) {
      await input.fill(`报修人-G2-${Date.now().toString(36)}`);
      await input.blur();
    } else {
      await page.locator('[data-kind="node"]').first().dblclick();
      await page.getByTestId('node-name-input').fill(`报修人-G2-${Date.now().toString(36)}`);
      await page.getByTestId('node-name-input').blur();
    }
    await page.waitForTimeout(150);
    await page.getByTestId('save-model').click();
    await page.waitForFunction(
      (prev) => {
        const el = document.querySelector('.forma-vme');
        return el && Number(el.getAttribute('data-model-revision')) > prev;
      },
      before.modelRevision,
      { timeout: 30000 },
    );
    const m = await api(`/api/forma/v1/businesses/${businessId}/model`, { tenantId, cookies });
    assert.equal(m.res.status, 200);
    assert.notEqual(m.json.error_key, 'FORMA_BUSINESS_INVALID_MODEL');
    await page.screenshot({ path: join(outDir, '05-final-save-ok.png') });
  });

  await browser.close();
});
