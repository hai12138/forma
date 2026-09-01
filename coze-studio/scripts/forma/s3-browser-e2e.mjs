/**
 * FORMA-S3-G2-F1 Real Browser E2E Gate
 *
 * Requires running Coze/Forma stack with S3 analyst routes + configured FORMA_ANALYST model.
 *
 *   FORMA_LIVE_E2E=1
 *   FORMA_LIVE_BASE_URL=http://127.0.0.1:8888
 *   FORMA_UI_BASE_URL=http://127.0.0.1:3001
 *   node --test scripts/forma/s3-browser-e2e.mjs
 */
import assert from 'node:assert/strict';
import { execSync } from 'node:child_process';
import { appendFileSync, mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import { chromium } from 'playwright';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const baseApi = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const baseUi = (process.env.FORMA_UI_BASE_URL || 'http://127.0.0.1:3001').replace(/\/$/, '');
const email = process.env.FORMA_LIVE_EMAIL || `forma-s3g2f1-${Date.now()}@example.com`;
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';
const mysqlUser = process.env.FORMA_MYSQL_USER || 'coze';
const mysqlPass = process.env.FORMA_MYSQL_PASSWORD || 'coze123';
const mysqlDb = process.env.FORMA_MYSQL_DATABASE || 'opencoze';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const outDir = join(root, 'forma', 'cursor-results', 's3-ui');
const browserLog = join(root, 'forma', 'cursor-results', 's3-browser-e2e.log');
const liveLog = join(root, 'forma', 'cursor-results', 's3-live-model-e2e.log');
const provenanceLog = join(root, 'forma', 'cursor-results', 's3-provenance-e2e.log');
const tenantLog = join(root, 'forma', 'cursor-results', 's3-tenant-isolation-e2e.log');

function log(path, line) {
  appendFileSync(path, `[${new Date().toISOString()}] ${line}\n`, 'utf8');
}

function queryMysql(sql) {
  try {
    const container = execSync('docker ps -qf "name=mysql"', { encoding: 'utf8' }).trim().split('\n')[0];
    if (!container) return null;
    const cmd = `docker exec ${container} mysql -u${mysqlUser} -p${mysqlPass} ${mysqlDb} -N -e "${sql.replace(/"/g, '\\"')}"`;
    return execSync(cmd, { encoding: 'utf8' }).trim();
  } catch {
    return null;
  }
}

async function apiRequest(request, path, { method = 'GET', body, tenantId, headers = {} } = {}) {
  const h = { Accept: 'application/json', 'X-Request-ID': `s3f1-${Date.now()}`, ...headers };
  if (tenantId) h['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) h['Content-Type'] = 'application/json';
  const res = await request.fetch(`${baseApi}${path}`, {
    method,
    headers: h,
    data: body,
  });
  const text = await res.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    json = { raw: text };
  }
  return { res, json };
}

async function assertAnalystRoutes(request, tenantId, businessId) {
  const probe = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/assertions`, { tenantId });
  assert.notEqual(probe.res.status(), 404, 'analyst API routes missing — rebuild/restart Coze backend with S3 code');
}

async function clickSideTab(page, label) {
  await page.locator('.forma-analyst-tab').filter({ hasText: label }).click();
}

async function submitBrowserTurn(page, text, timeout = 180000) {
  await page.fill('[data-testid="analyst-input"]', text);
  await page.click('[data-testid="analyst-submit"]');
  await page.waitForSelector('[data-testid="turn-analyst"]', { timeout });
}

async function getBusinessRevision(request, tenantId, businessId) {
  const r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/model`, { tenantId });
  return r.json.data?.current_revision ?? r.json.data?.master?.current_revision;
}

async function verifyModelCalls(sessionId) {
  const rows = queryMysql(
    `SELECT operation, model_ref, success, latency_ms FROM forma_analyst_model_call WHERE session_id='${sessionId}' ORDER BY id`,
  );
  if (!rows) {
    log(liveLog, `session_id=${sessionId} model_call_query=SKIPPED (mysql unavailable)`);
    return false;
  }
  log(liveLog, `session_id=${sessionId}`);
  for (const line of rows.split('\n').filter(Boolean)) {
    log(liveLog, line.replace(/\t/g, ' '));
    const [op, ref, success] = line.split('\t');
    assert.notEqual(ref, 'fake-analyst', `model_ref must not be fake-analyst for ${op}`);
    assert.equal(success, '1', `operation ${op} must succeed`);
  }
  assert.match(rows, /ExtractAssertions/);
  assert.match(rows, /GenerateInterviewTurn/);
  return true;
}

test('S3-G2-F1 analyst browser hard gate', async t => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }

  mkdirSync(outDir, { recursive: true });
  for (const f of [browserLog, liveLog, provenanceLog, tenantLog]) writeFileSync(f, '', 'utf8');
  log(browserLog, 'S3-G2-F1 browser hard gate start');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const request = context.request;

  let r = await apiRequest(request, '/api/passport/web/email/register/v2/', {
    method: 'POST',
    body: { email, password },
  });
  if (r.res.status() >= 400) {
    r = await apiRequest(request, '/api/passport/web/email/login/', {
      method: 'POST',
      body: { email, password },
    });
  }
  assert.ok(r.res.ok(), 'login/register');
  log(browserLog, 'authenticated via context.request PASS');

  r = await apiRequest(request, '/api/forma/v1/bootstrap', { method: 'POST', body: {} });
  assert.equal(r.res.status(), 200, JSON.stringify(r.json));
  const tenantA = r.json.data?.tenant?.tenant_id;
  assert.ok(tenantA, 'bootstrap tenant');

  r = await apiRequest(request, '/api/forma/v1/businesses', {
    method: 'POST',
    body: { name: '维修工单', description: 'S3 G2-F1 E2E', change_summary: 'e2e seed' },
    tenantId: tenantA,
  });
  assert.equal(r.res.status(), 200, JSON.stringify(r.json));
  const businessId = r.json.data?.business_id;
  const initialRevision = r.json.data?.current_revision;
  assert.ok(businessId);
  log(liveLog, `business_id=${businessId} initial_revision=r${initialRevision}`);
  log(browserLog, `business_id=${businessId} initial_revision=r${initialRevision}`);

  await assertAnalystRoutes(request, tenantA, businessId);

  const page = await context.newPage();
  await page.goto(`${baseUi}/forma/analyst`);
  await page.waitForSelector('[data-testid="business-select"]', { timeout: 30000 });
  assert.ok(await page.isVisible('[data-testid="business-select"]'), 'authenticated analyst page');
  assert.ok(!(await page.isVisible('[data-testid="tenant-empty"]')), 'must not be unauthenticated');

  await page.selectOption('[data-testid="business-select"]', businessId);
  await page.click('[data-testid="start-session"]');
  await page.waitForSelector('[data-testid="analyst-input"]', { timeout: 15000 });
  await page.screenshot({ path: join(outDir, '01-session-started.png'), fullPage: true });
  log(browserLog, 'session smoke PASS');

  const interview =
    '员工发现设备故障后提交报修，维修人员接单处理，完成后由管理员关闭。';
  await submitBrowserTurn(page, interview);
  assert.ok(await page.isVisible('[data-testid="turn-user"]'));
  assert.ok(await page.isVisible('[data-testid="turn-analyst"]'));
  await page.screenshot({ path: join(outDir, '02-real-model-reply.png'), fullPage: true });

  r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
  let assertions = r.json.data ?? [];
  assert.ok(assertions.length > 0, 'assertions > 0');
  r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/evidence`, { tenantId: tenantA });
  let evidence = r.json.data ?? [];
  assert.ok(evidence.length > 0, 'evidence > 0');
  await page.screenshot({ path: join(outDir, '03-assertions-evidence.png'), fullPage: true });

  r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/analyst/sessions`, { tenantId: tenantA });
  const sessions = r.json.data ?? [];
  assert.ok(sessions.length > 0, 'session exists');
  const sessionId = sessions[0].session_id;
  const modelVerified = await verifyModelCalls(sessionId);
  assert.ok(modelVerified, 'real model calls verified in forma_analyst_model_call');

  let rev = await getBusinessRevision(request, tenantA, businessId);
  assert.equal(rev, initialRevision, 'no silent mutation after turn 1');

  await submitBrowserTurn(
    page,
    '维修人员处理完成后，工单还需要有人关闭，但我还没有说明由谁关闭。',
  );
  await submitBrowserTurn(page, '管理员。');
  r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
  assertions = r.json.data ?? [];
  const contextLinked = assertions.some(
    a =>
      /管理员/.test(a.object_value || '') &&
      (/关闭|权限|工单/.test(a.predicate || '') ||
        /关闭|权限|工单/.test(a.object_value || '') ||
        /关闭|权限|工单/.test(a.subject_ref || '')),
  );
  assert.ok(contextLinked, 'multi-turn context links 管理员 to close semantics');

  for (let i = 0; i < 3; i++) {
    await submitBrowserTurn(page, `补充流程细节第 ${i + 3} 轮：审批与通知。`);
  }
  rev = await getBusinessRevision(request, tenantA, businessId);
  assert.equal(rev, initialRevision, 'no silent mutation after 5 user turns');
  log(browserLog, `no silent mutation PASS revision=r${rev}`);

  await clickSideTab(page, 'Gaps');
  await page.waitForSelector('[data-testid="gaps-panel"]');
  const gapAsk = page.locator('[data-testid="gap-ask"]').first();
  if (await gapAsk.count()) {
    const evBefore = (await apiRequest(request, `/api/forma/v1/businesses/${businessId}/evidence`, { tenantId: tenantA })).json.data?.length ?? 0;
    await gapAsk.click();
    await page.waitForSelector('[data-testid="turn-analyst"]', { timeout: 60000 });
    const evAfter = (await apiRequest(request, `/api/forma/v1/businesses/${businessId}/evidence`, { tenantId: tenantA })).json.data?.length ?? 0;
    assert.equal(evAfter, evBefore, 'gap ask must not create user evidence');
    await page.fill('[data-testid="analyst-input"]', '由系统管理员在工单完成后执行关闭操作。');
    await page.click('[data-testid="analyst-submit"]');
    await page.waitForSelector('[data-testid="turn-user"]', { timeout: 60000 });
    await page.screenshot({ path: join(outDir, '04-gap-ask.png'), fullPage: true });
    log(browserLog, 'gap ask PASS');
  } else {
    log(browserLog, 'gap ask SKIPPED (no open gaps)');
  }

  await clickSideTab(page, 'Assertions');
  const confirmBtn = page.locator('[data-testid="confirm-assertion"]').first();
  assert.ok(await confirmBtn.count(), 'PROPOSED assertion required for confirm gate');
  await confirmBtn.click();
  await page.waitForTimeout(1500);
  r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
  assertions = r.json.data ?? [];
  assert.ok(assertions.some(a => a.status === 'CONFIRMED'), 'confirmation persisted');
  rev = await getBusinessRevision(request, tenantA, businessId);
  assert.equal(rev, initialRevision, 'confirm must not bump revision');
  log(browserLog, 'confirmation PASS');

  const editBtn = page.locator('[data-testid="edit-confirm-assertion"]').first();
  if (await editBtn.count()) {
    await editBtn.click();
    await page.waitForSelector('[data-testid="edit-confirm-modal"]');
    await page.locator('[data-testid="edit-confirm-modal"] input').last().fill('管理员（人工确认）');
    await page.locator('[data-testid="edit-confirm-modal"] .forma-btn-primary').click();
    await page.waitForTimeout(1500);
    r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
    assertions = r.json.data ?? [];
    const edited = assertions.find(a => a.source_marker === 'MANUAL_MODIFIED' && a.status === 'CONFIRMED');
    assert.ok(edited, 'edit lineage persisted');
    assert.ok(edited.derived_from_assertion_id, 'derived_from_assertion_id set');
    await page.screenshot({ path: join(outDir, '05-edit-confirm.png'), fullPage: true });
    log(browserLog, 'edit confirm PASS');
  }

  await submitBrowserTurn(page, '补充：工单关闭权限仅属于系统管理员，不属于维修人员。');
  await submitBrowserTurn(page, '补充：工单关闭权限属于维修班长，不属于管理员。');
  await clickSideTab(page, 'Conflicts');
  await page.waitForSelector('[data-testid="conflicts-panel"]');
  if (await page.locator('[data-testid="conflict-card"]').count()) {
    await page.locator('[data-testid="conflict-card"] .forma-btn-primary').first().click();
    await page.waitForTimeout(1500);
    r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/conflicts`, { tenantId: tenantA });
    const conflicts = r.json.data ?? [];
    assert.ok(conflicts.some(c => c.status === 'RESOLVED'), 'conflict resolved');
    await page.screenshot({ path: join(outDir, '06-conflict-review.png'), fullPage: true });
    log(browserLog, 'conflict PASS');
  } else {
    log(browserLog, 'conflict SKIPPED (model did not surface conflict)');
  }

  await clickSideTab(page, 'Proposal');
  await page.click('[data-testid="generate-proposal"]');
  await page.waitForSelector('[data-testid="proposal-semantic-diff"]', { timeout: 120000 });
  assert.ok(await page.isVisible('[data-testid="proposal-semantic-diff"]'));
  await page.screenshot({ path: join(outDir, '07-proposal-diff.png'), fullPage: true });
  log(browserLog, 'proposal diff PASS');

  const applyBtn = page.locator('[data-testid="apply-proposal"]');
  assert.ok(await applyBtn.isEnabled(), 'apply must be enabled');
  await applyBtn.click();
  await page.waitForTimeout(3000);
  rev = await getBusinessRevision(request, tenantA, businessId);
  assert.equal(rev, initialRevision + 1, 'apply must bump revision');
  await page.goto(`${baseUi}/forma/analyst`);
  await page.selectOption('[data-testid="business-select"]', businessId);
  await page.screenshot({ path: join(outDir, '08-applied.png'), fullPage: true });

  const prov = queryMysql(
    `SELECT revision_no, proposal_id, assertion_ids_json FROM forma_revision_provenance WHERE business_id='${businessId}' AND revision_no=${rev}`,
  );
  assert.ok(prov, 'provenance row required');
  log(provenanceLog, prov.replace(/\t/g, ' '));
  await page.screenshot({ path: join(outDir, '09-provenance.png'), fullPage: true });
  log(browserLog, 'apply + provenance PASS');

  await clickSideTab(page, 'Proposal');
  await page.click('[data-testid="generate-proposal"]');
  await page.waitForSelector('[data-testid="proposal-panel"]', { timeout: 60000 });
  const staleBase = rev;
  r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/model`, { tenantId: tenantA });
  const sm = r.json.data?.semantic_model;
  sm.nodes = sm.nodes || [];
  sm.nodes.push({
    id: 'actor_stale_gate',
    type: 'ACTOR',
    name: 'Stale Gate Actor',
    source_marker: 'MANUAL_MODIFIED',
  });
  r = await apiRequest(request, `/api/forma/v1/businesses/${businessId}/model`, {
    method: 'PUT',
    tenantId: tenantA,
    body: { expected_revision: staleBase, semantic_model: sm, change_summary: 'stale gate bump' },
  });
  assert.equal(r.res.status(), 200);
  rev = await getBusinessRevision(request, tenantA, businessId);
  assert.equal(rev, staleBase + 1);
  await page.locator('[data-testid="apply-proposal"]').click();
  await page.waitForTimeout(2000);
  assert.match(await page.textContent('body'), /过期|STALE|stale/i);
  rev = await getBusinessRevision(request, tenantA, businessId);
  assert.equal(rev, staleBase + 1, 'stale apply must not bump revision again');
  await page.screenshot({ path: join(outDir, '10-stale-proposal.png'), fullPage: true });
  log(browserLog, 'stale proposal PASS');

  r = await apiRequest(request, '/api/forma/v1/tenants', {
    method: 'POST',
    body: { name: `Tenant B ${Date.now()}`, display_name: 'Tenant B E2E' },
    tenantId: tenantA,
  });
  assert.equal(r.res.status(), 200, JSON.stringify(r.json));
  const tenantB = r.json.data?.tenant_id;
  assert.ok(tenantB);
  await page.goto(`${baseUi}/forma/analyst`);
  await page.waitForSelector('#forma-tenant-select', { timeout: 15000 });
  await page.selectOption('#forma-tenant-select', tenantB);
  await page.waitForTimeout(2000);
  assert.ok(await page.isVisible('[data-testid="business-empty"]'), 'tenant A assets must disappear');
  await page.screenshot({ path: join(outDir, '11-tenant-switch.png'), fullPage: true });

  const denied = await apiRequest(request, `/api/forma/v1/businesses/${businessId}`, { tenantId: tenantB });
  assert.ok(denied.res.status() === 403 || denied.res.status() === 404, 'cross-tenant business denied');
  const deniedSession = await apiRequest(
    request,
    `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}`,
    { tenantId: tenantB },
  );
  assert.ok(deniedSession.res.status() === 403 || deniedSession.res.status() === 404);
  log(tenantLog, `tenantA=${tenantA} tenantB=${tenantB} cross_tenant=denied`);
  log(browserLog, 'tenant isolation PASS');

  await browser.close();
  log(browserLog, 'S3-G2-F1 browser hard gate COMPLETE');
});
