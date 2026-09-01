/**
 * FORMA-S2-G1 MANUAL_LIVE_BROWSER_GATE
 *
 * Real Playwright browser interactions against live Forma frontend + backend.
 * Assertions verify revision numbers — screenshots are evidence only.
 *
 *   FORMA_LIVE_E2E=1
 *   FORMA_LIVE_BASE_URL=http://127.0.0.1:8888
 *   FORMA_UI_BASE_URL=http://127.0.0.1:3001
 *   node --test scripts/forma/s2-g1-browser-gates.mjs
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
const email = process.env.FORMA_LIVE_EMAIL || `forma-s2g1-${Date.now()}@example.com`;
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const outDir = join(root, 'forma', 'cursor-results', 's2-g1-ui');

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
      return [...cookies.entries()].map(([name, value]) => ({
        name,
        value,
        url,
      }));
    },
  };
}

async function api(path, { method = 'GET', body, tenantId, cookies } = {}) {
  const headers = { Accept: 'application/json', 'X-Request-ID': `s2g1-${Date.now()}` };
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

const seed = {
  schema_version: '2.0',
  nodes: [
    { id: 'n1', type: 'ACTOR', name: '报修人', source_marker: 'MANUAL_MODIFIED' },
    { id: 'n2', type: 'BUSINESS_OBJECT', name: '维修工单', source_marker: 'MANUAL_MODIFIED' },
    { id: 'n3', type: 'PROCESS', name: '受理', source_marker: 'MANUAL_MODIFIED' },
  ],
  edges: [
    {
      id: 'e1',
      source: 'n1',
      target: 'n2',
      type: 'CREATES',
      label: '创建',
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
  rules: [
    {
      id: 'rule1',
      name: '关闭权限',
      applies_to: ['n2'],
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
  states: [
    {
      id: 's1',
      object_ref: 'n2',
      name: '待受理',
      initial: true,
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
};

test('S2-G1 real browser gates', async (t) => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1 — MANUAL_LIVE_BROWSER_GATE');
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

  const boot = await api('/api/forma/v1/bootstrap', {
    method: 'POST',
    body: {},
    cookies,
  });
  assert.equal(boot.res.status, 200, JSON.stringify(boot.json));
  const tenantId = boot.json.data.tenant.tenant_id;

  const created = await api('/api/forma/v1/businesses', {
    method: 'POST',
    tenantId,
    cookies,
    body: { name: '维修工单-S2G1', semantic_model: seed, change_summary: 's2-g1 seed' },
  });
  assert.equal(created.res.status, 200, JSON.stringify(created.json));
  const businessId = created.json.data.business_id;

  let modelRev = created.json.data.current_revision;
  let layoutGet = await api(`/api/forma/v1/businesses/${businessId}/layout`, {
    tenantId,
    cookies,
  });
  let layoutRev = layoutGet.json.data.layout_revision;

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await context.addCookies([
    ...cookies.entries(baseUi),
    ...cookies.entries(baseApi),
  ]);
  await context.addInitScript((tid) => {
    try {
      sessionStorage.setItem('forma.selectedTenantId', tid);
    } catch {
      /* ignore */
    }
  }, tenantId);
  const page = await context.newPage();

  await page.goto(`${baseUi}/business/${businessId}`, { waitUntil: 'networkidle' });
  await page.waitForSelector('.forma-vme', { timeout: 60000 });
  await page.screenshot({ path: join(outDir, '01-editor-current.png'), fullPage: true });

  const readAttrs = async () =>
    page.locator('.forma-vme').evaluate((el) => ({
      modelRevision: Number(el.getAttribute('data-model-revision')),
      layoutRevision: Number(el.getAttribute('data-layout-revision')),
      semanticDirty: el.getAttribute('data-semantic-dirty') === 'true',
      layoutDirty: el.getAttribute('data-layout-dirty') === 'true',
      layoutMode: el.getAttribute('data-layout-mode'),
      readonly: el.getAttribute('data-readonly') === 'true',
    }));

  // --- STEP 11: real browser drag ---
  await t.test('real browser drag → layout dirty → save layout', async () => {
    const before = await readAttrs();
    assert.equal(before.modelRevision, modelRev);
    assert.equal(before.layoutRevision, layoutRev);

    const node = page.locator('[data-node="n1"]');
    const box = await node.boundingBox();
    assert.ok(box, 'node n1 visible');
    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
    await page.mouse.down();
    await page.mouse.move(box.x + 120, box.y + 80, { steps: 12 });
    await page.mouse.up();

    await page.waitForTimeout(200);
    const afterDrag = await readAttrs();
    assert.equal(afterDrag.layoutDirty, true, 'Layout Dirty after real drag');
    assert.equal(afterDrag.semanticDirty, false, 'Semantic Dirty must stay false');
    await page.screenshot({
      path: join(outDir, '02-node-dragged-layout-dirty.png'),
      fullPage: true,
    });

    await page.getByTestId('save-layout').click();
    await page.waitForFunction(
      (prev) => {
        const el = document.querySelector('.forma-vme');
        return (
          el &&
          Number(el.getAttribute('data-layout-revision')) > prev &&
          el.getAttribute('data-layout-dirty') === 'false'
        );
      },
      layoutRev,
      { timeout: 30000 },
    );
    const afterSave = await readAttrs();
    assert.equal(afterSave.modelRevision, modelRev, 'semantic revision unchanged');
    assert.ok(afterSave.layoutRevision > layoutRev, 'layout_revision +1');
    layoutRev = afterSave.layoutRevision;
    await page.screenshot({
      path: join(outDir, '03-layout-saved-semantic-revision-unchanged.png'),
      fullPage: true,
    });

    // Confirm via API
    const m = await api(`/api/forma/v1/businesses/${businessId}/model`, { tenantId, cookies });
    assert.equal(m.json.data.current_revision, modelRev);
    const l = await api(`/api/forma/v1/businesses/${businessId}/layout`, { tenantId, cookies });
    assert.equal(l.json.data.layout_revision, layoutRev);
  });

  // --- STEP 12: semantic save ---
  await t.test('real browser double-click rename → save model', async () => {
    const node = page.locator('[data-node="n1"]');
    await node.dblclick();
    const input = page.getByTestId('node-name-input');
    await input.waitFor({ state: 'visible' });
    await input.fill('报修客户-浏览器');
    await input.blur();
    await page.waitForTimeout(150);
    const dirty = await readAttrs();
    assert.equal(dirty.semanticDirty, true);
    await page.screenshot({ path: join(outDir, '04-semantic-edited.png'), fullPage: true });

    await page.getByTestId('save-model').click();
    await page.waitForFunction(
      (prev) => {
        const el = document.querySelector('.forma-vme');
        return (
          el &&
          Number(el.getAttribute('data-model-revision')) > prev &&
          el.getAttribute('data-semantic-dirty') === 'false'
        );
      },
      modelRev,
      { timeout: 30000 },
    );
    const after = await readAttrs();
    assert.equal(after.modelRevision, modelRev + 1);
    assert.equal(after.layoutRevision, layoutRev, 'layout not bumped by semantic save');
    modelRev = after.modelRevision;
    await page.screenshot({
      path: join(outDir, '05-semantic-saved-revision-incremented.png'),
      fullPage: true,
    });

    await page.getByRole('button', { name: 'Revisions' }).click();
    await page.waitForSelector('[data-testid="revisions-panel"]');
    await page.getByTestId(`revision-r${modelRev - 1}`).click();
    await page.waitForSelector('[data-testid="revision-readonly-banner"]');
    const hist = await readAttrs();
    assert.equal(hist.readonly, true);
    await page.screenshot({ path: join(outDir, '07-history-readonly.png'), fullPage: true });
    await page.getByRole('button', { name: 'Back to Current' }).click();
    await page.waitForFunction(() => {
      const el = document.querySelector('.forma-vme');
      return el && el.getAttribute('data-readonly') === 'false';
    });
  });

  // --- STEP 13: auto layout ---
  await t.test('auto layout → save layout', async () => {
    await page.getByTestId('auto-layout').click();
    await page.waitForTimeout(150);
    const dirty = await readAttrs();
    assert.equal(dirty.layoutDirty, true);
    assert.equal(dirty.semanticDirty, false);
    assert.equal(dirty.layoutMode, 'auto');
    await page.screenshot({ path: join(outDir, '06-auto-layout.png'), fullPage: true });
    await page.getByTestId('save-layout').click();
    await page.waitForFunction(
      (prev) => {
        const el = document.querySelector('.forma-vme');
        return el && Number(el.getAttribute('data-layout-revision')) > prev;
      },
      layoutRev,
      { timeout: 30000 },
    );
    const after = await readAttrs();
    assert.equal(after.modelRevision, modelRev);
    assert.ok(after.layoutRevision > layoutRev);
    layoutRev = after.layoutRevision;
  });

  // --- STEP 14: relationship contract ---
  await t.test('rule has no connection handle; edge type is dropdown', async () => {
    const ruleHandle = page.locator('[data-node="rule1"] [data-handle]');
    assert.equal(await ruleHandle.count(), 0, 'Rule must not show connection handle');
    await page.locator('[data-edge="e1"]').click();
    await page.getByTestId('edge-type-select').waitFor({ state: 'visible' });
    const options = await page.getByTestId('edge-type-select').locator('option').allTextContents();
    assert.ok(options.includes('CREATES'));
    assert.ok(options.includes('RELATES_TO'));
  });

  // --- STEP 15: dependency delete ---
  await t.test('dependency-aware delete stays valid', async () => {
    page.once('dialog', (d) => d.accept());
    await page.locator('[data-node="n2"]').click();
    await page.getByRole('button', { name: '删除' }).click();
    await page.waitForTimeout(200);
    const dirty = await readAttrs();
    assert.equal(dirty.semanticDirty, true);
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
    await page.screenshot({ path: join(outDir, '09-delete-dependency.png'), fullPage: true });

    const m = await api(`/api/forma/v1/businesses/${businessId}/model`, { tenantId, cookies });
    assert.equal(m.res.status, 200);
    assert.ok(!m.json.data.semantic_model.nodes.find((n) => n.id === 'n2'));
    assert.ok(!m.json.data.semantic_model.states.find((s) => s.object_ref === 'n2'));
  });

  // Diff panel screenshot
  await page.getByRole('button', { name: 'Diff' }).click();
  await page.waitForSelector('[data-testid="diff-panel"]');
  await page.screenshot({ path: join(outDir, '08-diff.png'), fullPage: true });

  await browser.close();
});
