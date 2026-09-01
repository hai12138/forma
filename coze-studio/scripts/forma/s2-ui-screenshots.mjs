/**
 * Capture S2 UI screenshots against local Forma app + live harness.
 * Usage: node scripts/forma/s2-ui-screenshots.mjs
 */
import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const outDir = join(root, 'forma', 'cursor-results', 's2-ui');
mkdirSync(outDir, { recursive: true });

const baseApi = process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888';
const baseUi = process.env.FORMA_UI_BASE_URL || 'http://127.0.0.1:3001';
const email = `forma-s2-ui-${Date.now()}@example.com`;
const password = 'FormaE2E!23456';

async function registerAndLogin() {
  const jar = [];
  const store = (res) => {
    const raw = res.headers.getSetCookie?.() || [];
    for (const c of raw) jar.push(c.split(';')[0]);
  };
  let res = await fetch(`${baseApi}/api/passport/web/email/register/v2/`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  store(res);
  if (!res.ok) {
    res = await fetch(`${baseApi}/api/passport/web/email/login/`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    store(res);
  }
  const boot = await fetch(`${baseApi}/api/forma/v1/bootstrap`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Cookie: jar.join('; '),
    },
    body: '{}',
  });
  store(boot);
  const bootJson = await boot.json();
  const tenantId = bootJson.data?.tenant?.tenant_id;
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
    rules: [],
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
  const create = await fetch(`${baseApi}/api/forma/v1/businesses`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Cookie: jar.join('; '),
      'X-Forma-Tenant': tenantId,
    },
    body: JSON.stringify({
      name: '维修工单',
      semantic_model: seed,
      change_summary: 'ui seed',
    }),
  });
  const created = await create.json();
  return {
    cookies: jar,
    tenantId,
    businessId: created.data?.business_id,
  };
}

const { cookies, tenantId, businessId } = await registerAndLogin();
writeFileSync(
  join(outDir, 'session.json'),
  JSON.stringify({ email, tenantId, businessId }, null, 2),
);

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
for (const pair of cookies) {
  const [name, ...rest] = pair.split('=');
  await context.addCookies([
    { name, value: rest.join('='), url: baseUi },
    { name, value: rest.join('='), url: baseApi },
  ]);
}
await context.addInitScript((tid) => {
  try {
    sessionStorage.setItem('forma.selectedTenantId', tid);
  } catch {
    /* ignore */
  }
}, tenantId);

const page = await context.newPage();
await page.goto(`${baseUi}/business`, { waitUntil: 'networkidle' });
await page.waitForTimeout(1500);
await page.screenshot({ path: join(outDir, '01-business-list.png'), fullPage: true });

if (businessId) {
  await page.goto(`${baseUi}/business/${businessId}`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);
  await page.screenshot({ path: join(outDir, '02-visual-editor.png'), fullPage: true });

  const node = page.locator('.forma-vm-node, [data-node-id], .vm-node').first();
  if ((await node.count()) > 0) {
    await node.click();
    await page.waitForTimeout(500);
    await page.screenshot({ path: join(outDir, '03-node-selected.png'), fullPage: true });
  } else {
    // click first foreignObject / canvas item by text
    const byText = page.getByText('报修人').first();
    if (await byText.count()) {
      await byText.click();
      await page.waitForTimeout(500);
    }
    await page.screenshot({ path: join(outDir, '03-node-selected.png'), fullPage: true });
  }

  const revBtn = page.getByRole('button', { name: /Revision|历史|版本/i }).first();
  if (await revBtn.count()) {
    await revBtn.click();
    await page.waitForTimeout(800);
    await page.screenshot({ path: join(outDir, '04-revision-history.png'), fullPage: true });
  }

  const diffBtn = page.getByRole('button', { name: /Diff|对比/i }).first();
  if (await diffBtn.count()) {
    await diffBtn.click();
    await page.waitForTimeout(800);
    await page.screenshot({ path: join(outDir, '05-diff-view.png'), fullPage: true });
  }

  const fitBtn = page.getByRole('button', { name: /Fit|适配/i }).first();
  if (await fitBtn.count()) {
    await fitBtn.click();
    await page.waitForTimeout(400);
    await page.screenshot({ path: join(outDir, '06-fit-view.png'), fullPage: true });
  }
}

await browser.close();
console.log('screenshots written to', outDir);
