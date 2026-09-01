/**
 * FORMA-S3-G2 Real Browser E2E Gate
 *
 *   FORMA_LIVE_E2E=1
 *   FORMA_LIVE_BASE_URL=http://127.0.0.1:8888
 *   FORMA_UI_BASE_URL=http://127.0.0.1:3001
 *   node --test scripts/forma/s3-browser-e2e.mjs
 */
import assert from 'node:assert/strict';
import { appendFileSync, mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import { chromium } from 'playwright';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const baseApi = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const baseUi = (process.env.FORMA_UI_BASE_URL || 'http://127.0.0.1:3001').replace(/\/$/, '');
const email = process.env.FORMA_LIVE_EMAIL || `forma-s3g2-${Date.now()}@example.com`;
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const outDir = join(root, 'forma', 'cursor-results', 's3-ui');
const logPath = join(root, 'forma', 'cursor-results', 's3-browser-e2e.log');

function log(line) {
  appendFileSync(logPath, `[${new Date().toISOString()}] ${line}\n`, 'utf8');
}

function jar() {
  const cookies = new Map();
  return {
    store(res) {
      for (const c of res.headers.getSetCookie?.() || []) {
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
  const headers = { Accept: 'application/json', 'X-Request-ID': `s3g2-${Date.now()}` };
  if (cookies) headers.Cookie = cookies.header();
  if (tenantId) headers['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  const res = await fetch(`${baseApi}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (cookies) cookies.store(res);
  return { res, json: await res.json().catch(() => ({})) };
}

test('S3-G2 analyst browser gates', async t => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }

  mkdirSync(outDir, { recursive: true });
  writeFileSync(logPath, '', 'utf8');
  log('start');

  const cookies = jar();
  let r = await api('/api/passport/web/email/register/v2/', {
    method: 'POST',
    body: { email, password },
    cookies,
  });
  if (r.res.status >= 400) {
    r = await api('/api/passport/web/email/login/', { method: 'POST', body: { email, password }, cookies });
  }
  assert.ok(r.res.ok, 'login');

  r = await api('/api/forma/v1/tenants', { method: 'POST', body: { name: `S3G2 ${Date.now()}` }, cookies });
  const tenantId = r.json.data?.tenant_id;
  assert.ok(tenantId);

  r = await api('/api/forma/v1/businesses', {
    method: 'POST',
    body: { name: '维修工单', description: 'S3 G2' },
    tenantId,
    cookies,
  });
  const businessId = r.json.data?.business_id;
  assert.ok(businessId);

  const browser = await chromium.launch({ headless: true });
  const page = await (await browser.newContext()).newPage();
  await page.goto(`${baseUi}/forma/analyst`);
  await page.waitForSelector('[data-testid="business-select"]', { timeout: 30000 });
  await page.selectOption('[data-testid="business-select"]', businessId);
  await page.click('[data-testid="start-session"]');
  await page.waitForSelector('[data-testid="interview-input"]', { timeout: 15000 });
  await page.screenshot({ path: join(outDir, '01-session-started.png'), fullPage: true });

  await page.fill('[data-testid="interview-input"]', '员工发现设备故障后提交报修，维修人员接单处理，完成后由管理员关闭。');
  await page.click('[data-testid="submit-turn"]');
  await page.waitForSelector('[data-testid="analyst-reply"]', { timeout: 120000 });
  await page.screenshot({ path: join(outDir, '02-real-model-reply.png'), fullPage: true });
  log('model reply visible');

  await browser.close();
});
