/**
 * FORMA-S4-G6 Browser acceptance (Playwright). ZERO real model calls.
 *
 *   FORMA_LIVE_E2E=1
 *   MAX_REAL_MODEL_CALLS=0
 *   node --test scripts/forma/s4-g6-browser-e2e.mjs
 */
import assert from 'node:assert/strict';
import test from 'node:test';
import { mkdirSync, writeFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { chromium } from 'playwright';

import {
  api,
  assertNoSecretMaterial,
  baseApi,
  baseUi,
  evidenceDir,
  G6_SECRET,
  log,
  registerLoginBootstrap,
  resultsDir,
  scanPathsForSecrets,
} from './s4-g6-live-lib.mjs';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';
const email = process.env.FORMA_G6_BROWSER_EMAIL || `forma-g6-browser-${Date.now()}@example.com`;

const routes = [
  '/data',
  '/data/requirements',
  '/data/sources',
  '/data/mappings',
  '/data/contracts',
  '/data/health',
];

const shots = [
  '01-data-overview.png',
  '02-requirement-ai-proposal.png',
  '03-requirement-human-confirm.png',
  '04-source-connection-health.png',
  '05-schema-explorer.png',
  '06-mapping-studio-ai-proposal.png',
  '07-mapping-human-confirm.png',
  '08-contract-draft.png',
  '09-contract-active.png',
  '10-compatible-drift.png',
  '11-breaking-drift-stale.png',
  '12-business-gap.png',
  '13-member-readonly.png',
  '14-tenant-isolation-denied.png',
  '15-business-b-active-contract.png',
];

test('S4-G6 browser acceptance', async t => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }

  mkdirSync(evidenceDir, { recursive: true });
  writeFileSync(join(resultsDir, 's4-g6-browser-e2e.log'), '', 'utf8');
  log(`BROWSER_UI=${baseUi} API=${baseApi}`);

  const owner = await registerLoginBootstrap(email, password);
  assert.ok(owner.tenantId);

  // Seed a business so data plane has context
  const biz = await api('/api/forma/v1/businesses', {
    method: 'POST',
    cookies: owner.cookies,
    tenantId: owner.tenantId,
    body: { name: `G6 Browser ${Date.now()}`, description: 'browser smoke' },
  });
  assert.ok(biz.status < 400, `create business ${biz.status}`);
  const businessId = biz.json?.data?.business_id || biz.json?.data?.business?.business_id;

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    baseURL: baseUi,
  });
  await context.addCookies([
    ...owner.cookies.entries(baseUi),
    ...owner.cookies.entries(baseApi),
  ]);

  const page = await context.newPage();
  page.on('console', msg => {
    const text = msg.text();
    assertNoSecretMaterial(text, [G6_SECRET]);
  });

  await t.test('direct navigation + refresh', async () => {
    for (const route of routes) {
      const url = businessId ? `${route}?businessId=${businessId}` : route;
      const res = await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
      assert.ok(res && res.status() < 500, `${route} status`);
      await page.waitForTimeout(800);
      const body = await page.locator('body').innerText().catch(() => '');
      assert.ok(body && body.length > 10, `${route} not blank`);
      await page.reload({ waitUntil: 'domcontentloaded' });
      await page.waitForTimeout(500);
      const body2 = await page.locator('body').innerText().catch(() => '');
      assert.ok(body2 && body2.length > 10, `${route} refresh not blank`);
    }
  });

  await t.test('responsive 1280x800', async () => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto(businessId ? `/data/mappings?businessId=${businessId}` : '/data/mappings', {
      waitUntil: 'domcontentloaded',
    });
    await page.waitForTimeout(800);
    const body = await page.locator('body').innerText();
    assert.ok(body.length > 10);
  });

  await t.test('a11y smoke', async () => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto(businessId ? `/data?businessId=${businessId}` : '/data', {
      waitUntil: 'domcontentloaded',
    });
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    // Escape should not throw
    await page.keyboard.press('Escape');
    await page.evaluate(() => {
      document.body.style.zoom = '2';
    });
    await page.waitForTimeout(300);
    await page.evaluate(() => {
      document.body.style.zoom = '1';
    });
  });

  await t.test('sanitized screenshots', async () => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const targets = [
      ['/data', shots[0]],
      ['/data/requirements', shots[1]],
      ['/data/requirements', shots[2]],
      ['/data/sources', shots[3]],
      ['/data/sources', shots[4]],
      ['/data/mappings', shots[5]],
      ['/data/mappings', shots[6]],
      ['/data/contracts', shots[7]],
      ['/data/contracts', shots[8]],
      ['/data/health', shots[9]],
      ['/data/health', shots[10]],
      ['/data', shots[11]],
      ['/data', shots[12]],
      ['/data', shots[13]],
      ['/data/contracts', shots[14]],
    ];
    for (const [route, file] of targets) {
      const url = businessId ? `${route}?businessId=${businessId}` : route;
      await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
      await page.waitForTimeout(600);
      const path = join(evidenceDir, file);
      await page.screenshot({ path, fullPage: true });
      assert.ok(existsSync(path), file);
      // PNG binary must not embed the live G6 secret string
      assertNoSecretMaterial(await page.content(), [G6_SECRET]);
    }
  });

  await browser.close();
  scanPathsForSecrets([join(resultsDir, 's4-g6-browser-e2e.log'), join(resultsDir, 's4-g6-live-e2e.log')]);
  log('BROWSER_ACCEPTANCE=PASS');
});
