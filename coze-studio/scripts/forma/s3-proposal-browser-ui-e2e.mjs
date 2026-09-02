/**
 * FORMA-S3-BROWSER-UI-FINAL-CLOSURE
 * Proposal semantic diff + Apply browser gates only. ZERO real model calls.
 *
 *   FORMA_LIVE_E2E=1
 *   MAX_REAL_MODEL_CALLS=0
 *   FORMA_S3_E2E_RESUME=1
 *   node --test scripts/forma/s3-proposal-browser-ui-e2e.mjs
 */
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { appendFileSync, mkdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import { chromium } from 'playwright';

import { loadState, statePath } from './s3-e2e-budget.mjs';
import {
  countModelCalls,
  countModelCallsForBusiness,
  queryMysql,
  seedConfirmedAssertionForProposalUI,
} from './s3-e2e-fixtures.mjs';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const baseApi = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const baseUi = (process.env.FORMA_UI_BASE_URL || 'http://127.0.0.1:3001').replace(/\/$/, '');
const email =
  process.env.FORMA_LIVE_EMAIL ||
  process.env.FORMA_S3_E2E_EMAIL ||
  loadState()?.authEmail ||
  'forma-s3live@example.com';
const password = process.env.FORMA_LIVE_PASSWORD || process.env.FORMA_S3_E2E_PASSWORD || 'FormaE2E!23456';
const maxModelCalls = parseInt(process.env.MAX_REAL_MODEL_CALLS || '0', 10);

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const outDir = join(root, 'forma', 'cursor-results', 's3-ui');
const browserLog = join(root, 'forma', 'cursor-results', 's3-browser-e2e.log');
const provenanceLog = join(root, 'forma', 'cursor-results', 's3-provenance-e2e.log');

function log(path, line) {
  appendFileSync(path, `[${new Date().toISOString()}] ${line}\n`, 'utf8');
}

function sha256File(path) {
  return createHash('sha256').update(readFileSync(path)).digest('hex');
}

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

async function apiRequest(cookies, path, { method = 'GET', body, tenantId, headers = {} } = {}) {
  const h = { Accept: 'application/json', 'X-Request-ID': `s3pui-${Date.now()}`, ...headers };
  if (tenantId) h['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) h['Content-Type'] = 'application/json';
  if (cookies) h.Cookie = cookies.header();
  const res = await fetch(`${baseApi}${path}`, {
    method,
    headers: h,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  cookies.store(res);
  const text = await res.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    json = { raw: text };
  }
  return { res, json };
}

async function getBusinessRevision(cookies, tenantId, businessId) {
  const r = await apiRequest(cookies, `/api/forma/v1/businesses/${businessId}/model`, { tenantId });
  return r.json.data?.current_revision ?? r.json.data?.master?.current_revision;
}

async function clickSideTab(page, label) {
  const btn = page.getByRole('button', { name: label });
  await btn.waitFor({ state: 'visible', timeout: 60000 });
  await btn.click();
}

async function waitAnalystIdle(page, timeout = 120000) {
  await page.waitForFunction(
    () => {
      const busy = [...document.querySelectorAll('.forma-analyst-processing')].some(el =>
        /正在分析业务事实/.test(el.textContent || ''),
      );
      return !busy;
    },
    { timeout },
  ).catch(() => {});
}

test('S3 proposal browser UI final closure', async t => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }

  const state = loadState();
  assert.ok(state?.businessId && state?.sessionId && state?.tenantId, `missing ${statePath}`);

  const { tenantId: tenantA, businessId, sessionId, principalId } = state;
  const modelCallsBefore = countModelCallsForBusiness(businessId);

  mkdirSync(outDir, { recursive: true });
  log(browserLog, 'S3-BROWSER-UI-FINAL proposal closure start MAX_REAL_MODEL_CALLS=0');

  const cookies = jar();
  let r = await apiRequest(cookies, '/api/passport/web/email/login/', {
    method: 'POST',
    body: { email, password },
  });
  if (!r.res.ok) {
    r = await apiRequest(cookies, '/api/passport/web/email/register/v2/', {
      method: 'POST',
      body: { email, password },
    });
  }
  assert.ok(r.res.ok, 'login/register');

  r = await apiRequest(cookies, `/api/forma/v1/businesses/${businessId}/analyst/sessions`, {
    method: 'POST',
    tenantId: tenantA,
    body: { title: 'S3 proposal UI closure', confirmation_policy: 'DEVELOPMENT' },
  });
  assert.equal(r.res.status, 200, JSON.stringify(r.json));
  const uiSessionId = r.json.data?.session_id;
  assert.ok(uiSessionId, 'fixture analyst session required');

  seedConfirmedAssertionForProposalUI({
    tenantId: tenantA,
    businessId,
    sessionId: uiSessionId,
    principalId,
  });

  const revBefore = await getBusinessRevision(cookies, tenantA, businessId);

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await context.addCookies([...cookies.entries(baseUi), ...cookies.entries(baseApi)]);
  await context.addInitScript(tid => {
    try {
      sessionStorage.setItem('forma.selectedTenantId', tid);
    } catch {
      /* ignore */
    }
  }, tenantA);

  const page = await context.newPage();
  await page.goto(`${baseUi}/analyst?session_id=${encodeURIComponent(uiSessionId)}`);
  await page.waitForSelector('[data-testid="business-select"]', { timeout: 30000 });
  await page.selectOption('[data-testid="business-select"]', businessId);
  await page.waitForFunction(
    sid => document.body.textContent?.includes(String(sid).slice(0, 12)),
    uiSessionId,
    { timeout: 30000 },
  );

  await clickSideTab(page, 'Proposal');
  await waitAnalystIdle(page);
  await page.waitForFunction(
    () => {
      const btn = document.querySelector('[data-testid="generate-proposal"]');
      return btn && !btn.disabled;
    },
    { timeout: 120000 },
  );
  await page.click('[data-testid="generate-proposal"]');
  await page.waitForSelector('[data-testid="proposal-semantic-diff"]', { timeout: 120000 });
  await page.waitForSelector('[data-testid="proposal-validation"]', { timeout: 30000 });
  const validationText = await page.locator('[data-testid="proposal-validation"]').innerText();
  assert.match(validationText, /VALID/i, validationText);

  for (const section of ['nodes', 'edges', 'states', 'rules']) {
    await page.waitForSelector(`[data-testid="proposal-diff-${section}"]`, { timeout: 30000 });
  }
  const summary = await page.locator('[data-testid="proposal-diff-summary"]').innerText();
  assert.match(summary, /Current r\d+ vs Proposed r\d+/i, summary);

  await page.screenshot({ path: join(outDir, '07-proposal-diff.png'), fullPage: true });
  log(browserLog, 'proposal semantic diff UI PASS');

  const applyBtn = page.locator('[data-testid="apply-proposal"]');
  await applyBtn.waitFor({ state: 'visible', timeout: 30000 });
  await page.waitForFunction(
    () => {
      const btn = document.querySelector('[data-testid="apply-proposal"]');
      return btn && !btn.disabled;
    },
    { timeout: 120000 },
  );
  await applyBtn.click();
  const applied = page.locator('[data-testid="proposal-applied"]');
  await applied.waitFor({ state: 'visible', timeout: 60000 });
  const appliedText = await applied.innerText();
  assert.match(appliedText, /APPLIED/i, appliedText);
  await applied.scrollIntoViewIfNeeded();
  await applied.screenshot({ path: join(outDir, '08-applied.png') });
  log(browserLog, 'apply via browser UI PASS');

  const provenance = page.locator('[data-testid="proposal-provenance"]');
  await provenance.waitFor({ state: 'visible', timeout: 30000 });
  await provenance.scrollIntoViewIfNeeded();
  await provenance.screenshot({ path: join(outDir, '09-provenance.png') });
  log(browserLog, 'proposal APPLIED UI PASS');

  const revAfter = await getBusinessRevision(cookies, tenantA, businessId);
  assert.equal(revAfter, revBefore + 1, 'apply must bump revision by 1');

  const prov = queryMysql(
    `SELECT revision_no, proposal_id, assertion_ids_json FROM forma_revision_provenance WHERE business_id='${businessId}' AND revision_no=${revAfter}`,
  );
  assert.ok(prov, 'provenance row required');
  log(provenanceLog, prov.replace(/\t/g, ' '));

  const modelCallsAfter = countModelCallsForBusiness(businessId);
  assert.equal(modelCallsAfter, modelCallsBefore, 'zero new real model calls');

  const shas = ['07-proposal-diff.png', '08-applied.png', '09-provenance.png'].map(f =>
    sha256File(join(outDir, f)),
  );
  assert.equal(new Set(shas).size, 3, `screenshot SHA must differ: ${shas.join(', ')}`);

  await browser.close();
  log(browserLog, `S3-BROWSER-UI-FINAL proposal closure COMPLETE model_calls_unchanged=${modelCallsBefore}`);
});
