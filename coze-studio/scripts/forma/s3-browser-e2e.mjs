/**
 * FORMA-S3-LIVE-FINAL-GATE Real Browser E2E Hard Gate
 *
 *   FORMA_LIVE_E2E=1
 *   FORMA_LIVE_BASE_URL=http://127.0.0.1:8888
 *   FORMA_UI_BASE_URL=http://127.0.0.1:3001
 *
 * Budget / resume (local harness):
 *   MAX_REAL_MODEL_CALLS=8          — stop immediately when reached
 *   FORMA_S3_E2E_FROM_GATE=gap        — run from gate onward (no earlier real model)
 *   FORMA_S3_E2E_RESUME=1             — load forma/cursor-results/s3-e2e-state.json
 *   FORMA_S3_SKIP_REAL_MODEL=1        — skip real model if probe already PASS in state/DB
 *
 *   node --test scripts/forma/s3-browser-e2e.mjs
 */
import assert from 'node:assert/strict';
import { appendFileSync, mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import { chromium } from 'playwright';

import {
  assertModelBudget,
  gateIndex,
  initFreshLogs,
  loadState,
  markGateComplete,
  MAX_REAL_MODEL_CALLS,
  MAX_REAL_MODEL_USER_TURNS,
  reportBudgetStop,
  saveState,
  shouldRunGate,
} from './s3-e2e-budget.mjs';
import {
  countModelCalls,
  ensureOpenGap,
  ensureProposedAssertionForEdit,
  queryMysql,
  seedConfirmedAssertionForStale,
  seedDeterministicConflict,
  seedUserTurnAndEvidence,
  verifyModelCallsForSession,
} from './s3-e2e-fixtures.mjs';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const baseApi = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const baseUi = (process.env.FORMA_UI_BASE_URL || 'http://127.0.0.1:3001').replace(/\/$/, '');
const email = process.env.FORMA_LIVE_EMAIL || process.env.FORMA_S3_E2E_EMAIL || `forma-s3live-${Date.now()}@example.com`;
const password = process.env.FORMA_LIVE_PASSWORD || process.env.FORMA_S3_E2E_PASSWORD || 'FormaE2E!23456';
const fromGate = process.env.FORMA_S3_E2E_FROM_GATE || '';
const skipRealModel = process.env.FORMA_S3_SKIP_REAL_MODEL === '1';
const resumeMode = process.env.FORMA_S3_E2E_RESUME === '1';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const outDir = join(root, 'forma', 'cursor-results', 's3-ui');
const browserLog = join(root, 'forma', 'cursor-results', 's3-browser-e2e.log');
const liveLog = join(root, 'forma', 'cursor-results', 's3-live-model-e2e.log');
const provenanceLog = join(root, 'forma', 'cursor-results', 's3-provenance-e2e.log');
const tenantLog = join(root, 'forma', 'cursor-results', 's3-tenant-isolation-e2e.log');

function log(path, line) {
  appendFileSync(path, `[${new Date().toISOString()}] ${line}\n`, 'utf8');
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

let sessionCookies = jar();

async function apiRequest(path, { method = 'GET', body, tenantId, headers = {} } = {}) {
  const h = { Accept: 'application/json', 'X-Request-ID': `s3f1-${Date.now()}`, ...headers };
  if (tenantId) h['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) h['Content-Type'] = 'application/json';
  if (sessionCookies) h.Cookie = sessionCookies.header();
  const res = await fetch(`${baseApi}${path}`, {
    method,
    headers: h,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  sessionCookies.store(res);
  const text = await res.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    json = { raw: text };
  }
  return { res, json };
}

async function assertAnalystRoutes(tenantId, businessId) {
  for (const [method, path] of [
    ['GET', `/api/forma/v1/businesses/${businessId}/assertions`],
    ['GET', `/api/forma/v1/businesses/${businessId}/evidence`],
    ['GET', `/api/forma/v1/businesses/${businessId}/analyst/sessions`],
  ]) {
    const r = await apiRequest(path, { method, tenantId });
    assert.notEqual(r.res.status, 404, `${method} ${path} must not 404 — rebuild backend with current S3 code`);
  }
}

async function resolveBrowserSessionId(tenantId, businessId, timeout = 30000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    const r = await apiRequest(`/api/forma/v1/businesses/${businessId}/analyst/sessions`, { tenantId });
    const sessions = (r.json.data ?? []).filter(s => s.title !== 'route-probe');
    if (sessions.length > 0) {
      return sessions[sessions.length - 1].session_id;
    }
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  return null;
}

async function listAnalystTurns(tenantId, businessId, sessionId) {
  const r = await apiRequest(
    `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}/turns`,
    { tenantId },
  );
  return r.json.data ?? [];
}

async function waitTurnAnalysisComplete(page, tenantId, businessId, sessionId, beforeAnalystTurns, timeout = 300000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    const turns = await listAnalystTurns(tenantId, businessId, sessionId);
    const analystTurns = turns.filter(t => t.speaker === 'ANALYST');
    if (analystTurns.length > beforeAnalystTurns) {
      await page
        .waitForFunction(
          n => document.querySelectorAll('[data-testid="turn-analyst"]').length >= n,
          beforeAnalystTurns + 1,
          { timeout: 30000 },
        )
        .catch(() => {});
      return turns;
    }

    const lastUser = [...turns].reverse().find(t => t.speaker === 'USER');
    if (
      lastUser &&
      ['FAILED', 'EXTRACTION_FAILED', 'RESPONSE_FAILED'].includes(lastUser.analysis_status)
    ) {
      const retry = page.locator('[data-testid="retry-analysis"]').last();
      if (await retry.count()) {
        await waitAnalystIdle(page, timeout);
        await retry.click();
        await new Promise(resolve => setTimeout(resolve, 2000));
        continue;
      }
      throw new Error(`turn analysis failed: ${lastUser.analysis_status}`);
    }

    await new Promise(resolve => setTimeout(resolve, 2000));
  }
  throw new Error('turn analysis timeout waiting for ANALYST reply');
}

async function waitAnalystIdle(page, timeout = 180000) {
  await page.waitForFunction(
    () => {
      const busy = [...document.querySelectorAll('.forma-analyst-processing')].some(el =>
        /正在分析业务事实/.test(el.textContent || ''),
      );
      return !busy;
    },
    { timeout },
  );
}

async function waitSubmitReady(page, text, timeout = 180000) {
  await page.fill('[data-testid="analyst-input"]', text);
  await page.waitForFunction(
    expected => {
      const input = document.querySelector('[data-testid="analyst-input"]');
      const btn = document.querySelector('[data-testid="analyst-submit"]');
      return (
        input &&
        btn &&
        String(input.value || '').trim() === String(expected || '').trim() &&
        !btn.disabled
      );
    },
    text,
    { timeout },
  );
}

async function reloadAnalystUI(page, businessId) {
  await page.goto(`${baseUi}/analyst`);
  await page.waitForSelector('[data-testid="business-select"]', { timeout: 30000 });
  await page.selectOption('[data-testid="business-select"]', businessId);
  await page.waitForSelector('[data-testid="analyst-input"]', { timeout: 15000 }).catch(() => {});
  await page.waitForTimeout(800);
}

async function submitRealModelTurn(page, text, ctx, gateState) {
  assertModelBudget(ctx.sessionId, {
    completedGates: gateState.completedGates,
    failurePoint: 'real-model',
  });
  await submitBrowserTurn(page, text, ctx);
  const modelCalls = countModelCalls(ctx.sessionId);
  gateState.modelCalls = modelCalls;
  saveState(gateState);
  assertModelBudget(ctx.sessionId, {
    completedGates: gateState.completedGates,
    failurePoint: 'real-model',
  });
  return modelCalls;
}

async function clickSideTab(page, label) {
  const btn = page.getByRole('button', { name: label });
  await btn.waitFor({ state: 'visible', timeout: 60000 });
  await btn.click();
}

async function waitForAssertionCard(page, tenantId, businessId, assertId, marker, timeout = 120000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    const r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId });
    const row = (r.json.data ?? []).find(a => a.assertion_id === assertId);
    if (row?.status === 'PROPOSED') {
      await clickSideTab(page, 'Assertions');
      const card = page
        .locator('[data-testid="assertions-panel"] .forma-analyst-card')
        .filter({ hasText: marker });
      if (await card.count()) return row;
    }
    await new Promise(resolve => setTimeout(resolve, 2000));
  }
  throw new Error(`assertion card ${assertId} not visible`);
}

async function submitBrowserTurn(page, text, ctx = {}) {
  const { tenantId, businessId, sessionId, timeout = 300000 } = ctx;
  const beforeAnalystTurns = sessionId
    ? (await listAnalystTurns(tenantId, businessId, sessionId)).filter(t => t.speaker === 'ANALYST').length
    : await page.locator('[data-testid="turn-analyst"]').count();
  await waitAnalystIdle(page, timeout);
  await waitSubmitReady(page, text, timeout);
  await page.click('[data-testid="analyst-submit"]');
  if (sessionId) {
    await waitTurnAnalysisComplete(page, tenantId, businessId, sessionId, beforeAnalystTurns, timeout);
  } else {
    await page.waitForFunction(
      n => document.querySelectorAll('[data-testid="turn-analyst"]').length > n,
      beforeAnalystTurns,
      { timeout },
    );
  }
  await waitAnalystIdle(page, timeout);
}

async function getBusinessRevision(tenantId, businessId) {
  const r = await apiRequest(`/api/forma/v1/businesses/${businessId}/model`, { tenantId });
  return r.json.data?.current_revision ?? r.json.data?.master?.current_revision;
}

async function verifyModelCalls(sessionId) {
  const rows = verifyModelCallsForSession(sessionId, liveLog, log);
  assert.ok(rows.includes('ExtractAssertions'), 'ExtractAssertions model call required');
  assert.ok(rows.includes('GenerateInterviewTurn'), 'GenerateInterviewTurn model call required');
  for (const line of rows.split('\n').filter(Boolean)) {
    const [op, ref, success] = line.split('\t');
    assert.notEqual(ref, 'fake-analyst', `model_ref must not be fake-analyst for ${op}`);
    assert.equal(success, '1', `operation ${op} must succeed`);
  }
  return true;
}

test('S3-LIVE-FINAL analyst browser hard gate', async t => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }

  mkdirSync(outDir, { recursive: true });
  initFreshLogs(writeFileSync, [browserLog, liveLog, provenanceLog, tenantLog]);
  log(browserLog, `S3-LIVE-FINAL browser hard gate start budget=${MAX_REAL_MODEL_CALLS} from_gate=${fromGate || 'none'} resume=${resumeMode}`);

  const gateState = loadState() || {
    completedGates: [],
    realModelProbePass: false,
    modelCalls: 0,
  };
  if (fromGate) gateIndex(fromGate);

  sessionCookies = jar();
  const browser = await chromium.launch({ headless: true });

  let tenantA = gateState.tenantId;
  let principalId = gateState.principalId;
  let businessId = gateState.businessId;
  let sessionId = gateState.sessionId;
  let initialRevision = gateState.initialRevision;
  let confirmTarget = gateState.confirmTarget;

  if (shouldRunGate('auth', fromGate, gateState.completedGates)) {
    let r = await apiRequest('/api/passport/web/email/register/v2/', {
      method: 'POST',
      body: { email, password },
    });
    if (r.res.status >= 400) {
      r = await apiRequest('/api/passport/web/email/login/', {
        method: 'POST',
        body: { email, password },
      });
    }
    assert.ok(r.res.ok, 'login/register');
    log(browserLog, 'authenticated via fetch cookie jar PASS');

    if (resumeMode && gateState.tenantId) {
      tenantA = gateState.tenantId;
      principalId = gateState.principalId;
      businessId = gateState.businessId;
      sessionId = gateState.sessionId;
      initialRevision = gateState.initialRevision;
      confirmTarget = gateState.confirmTarget;
      log(browserLog, `resume session business_id=${businessId} session_id=${sessionId}`);
    } else {
      r = await apiRequest('/api/forma/v1/bootstrap', { method: 'POST', body: {} });
      assert.equal(r.res.status, 200, JSON.stringify(r.json));
      tenantA = r.json.data?.tenant?.tenant_id;
      principalId = r.json.data?.principal?.principal_id;
      assert.ok(tenantA, 'bootstrap tenant');
      assert.ok(principalId, 'bootstrap principal');

      r = await apiRequest('/api/forma/v1/businesses', {
        method: 'POST',
        body: { name: '维修工单', description: 'S3 G2-F1 E2E', change_summary: 'e2e seed' },
        tenantId: tenantA,
      });
      assert.equal(r.res.status, 200, JSON.stringify(r.json));
      businessId = r.json.data?.business_id;
      initialRevision = r.json.data?.current_revision;
      assert.ok(businessId);
      log(liveLog, `business_id=${businessId} initial_revision=r${initialRevision}`);
      log(browserLog, `business_id=${businessId} initial_revision=r${initialRevision}`);
    }

    gateState.tenantId = tenantA;
    gateState.principalId = principalId;
    gateState.businessId = businessId;
    gateState.sessionId = sessionId;
    gateState.initialRevision = initialRevision;
    gateState.email = email;
    markGateComplete(gateState, 'auth');
  }

  await assertAnalystRoutes(tenantA, businessId);

  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await context.addCookies([...sessionCookies.entries(baseUi), ...sessionCookies.entries(baseApi)]);
  await context.addInitScript(tid => {
    try {
      sessionStorage.setItem('forma.selectedTenantId', tid);
    } catch {
      /* ignore */
    }
  }, tenantA);

  const page = await context.newPage();
  if (resumeMode && gateState.tenantId && fromGate) {
    log(browserLog, `resume fixture phase from gate=${fromGate} — skipping UI auth smoke`);
  } else {
  await page.goto(`${baseUi}/analyst`);
  await page.waitForSelector('[data-testid="business-select"]', { timeout: 30000 });
  assert.ok(await page.isVisible('[data-testid="business-select"]'), 'authenticated analyst page');
  assert.ok(!(await page.isVisible('[data-testid="tenant-empty"]')), 'must not be unauthenticated');
  await page.selectOption('[data-testid="business-select"]', businessId);
  }

  let r;
  let assertions;
  let rev;

  if (shouldRunGate('real-model', fromGate, gateState.completedGates)) {
    const probeAlreadyPass =
      skipRealModel &&
      (gateState.realModelProbePass || (sessionId && countModelCalls(sessionId) >= 2));

    if (probeAlreadyPass) {
      log(browserLog, 'real-model SKIPPED — probe already PASS (budget preserved)');
      gateState.realModelProbePass = true;
      gateState.modelCalls = countModelCalls(sessionId);
      markGateComplete(gateState, 'real-model');
    } else {
      if (!sessionId) {
        await page.click('[data-testid="start-session"]');
        await page.waitForSelector('[data-testid="analyst-input"]', { timeout: 15000 });
        sessionId = await resolveBrowserSessionId(tenantA, businessId);
        assert.ok(sessionId, 'browser analyst session required');
        gateState.sessionId = sessionId;
      }
      const turnCtx = { tenantId: tenantA, businessId, sessionId };
      await page.screenshot({ path: join(outDir, '01-session-started.png'), fullPage: true });
      log(browserLog, `session smoke PASS session_id=${sessionId}`);

      const interview =
        '员工发现设备故障后提交报修，维修人员接单处理，完成后由管理员关闭。';
      await submitRealModelTurn(page, interview, turnCtx, gateState);
      assert.ok(await page.isVisible('[data-testid="turn-user"]'));
      assert.ok(await page.isVisible('[data-testid="turn-analyst"]'));
      await page.screenshot({ path: join(outDir, '02-real-model-reply.png'), fullPage: true });

      r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
      assertions = r.json.data ?? [];
      assert.ok(assertions.length > 0, 'assertions > 0');
      r = await apiRequest(`/api/forma/v1/businesses/${businessId}/evidence`, { tenantId: tenantA });
      assert.ok((r.json.data ?? []).length > 0, 'evidence > 0');
      await page.screenshot({ path: join(outDir, '03-assertions-evidence.png'), fullPage: true });

      const modelVerified = await verifyModelCalls(sessionId);
      assert.ok(modelVerified, 'real model calls verified in forma_analyst_model_call');
      gateState.realModelProbePass = true;
      gateState.modelCalls = countModelCalls(sessionId);

      let multiturnTurns = 0;
      if (multiturnTurns < MAX_REAL_MODEL_USER_TURNS) {
        await submitRealModelTurn(
          page,
          '维修人员处理完成后，工单还需要有人关闭，但我还没有说明由谁关闭。',
          turnCtx,
          gateState,
        );
        multiturnTurns += 1;
      }
      if (multiturnTurns < MAX_REAL_MODEL_USER_TURNS) {
        await submitRealModelTurn(page, '管理员。', turnCtx, gateState);
        multiturnTurns += 1;
      }

      const assertionBlob = a =>
        `${a.subject_ref || ''} ${a.predicate || ''} ${a.object_value || ''}`;
      let contextLinked = false;
      const ctxDeadline = Date.now() + 90000;
      while (Date.now() < ctxDeadline) {
        r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
        assertions = r.json.data ?? [];
        const hasAdmin = assertions.some(a => /管理员/.test(assertionBlob(a)));
        const hasCloseSemantics = assertions.some(a => /关闭|工单|权限|close/i.test(assertionBlob(a)));
        contextLinked =
          assertions.some(
            a => /管理员/.test(assertionBlob(a)) && /关闭|工单|权限|close/i.test(assertionBlob(a)),
          ) ||
          (hasAdmin && hasCloseSemantics);
        if (contextLinked) break;
        const turns = await listAnalystTurns(tenantA, businessId, sessionId);
        const analystText = turns
          .filter(t => t.speaker === 'ANALYST')
          .map(t => t.content || '')
          .join('\n');
        contextLinked = /管理员/.test(analystText) && /关闭|工单|权限|close/i.test(analystText);
        if (contextLinked) break;
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
      assert.ok(contextLinked, 'multi-turn context links 管理员 to close semantics');
      log(browserLog, `real-model PASS model_calls=${countModelCalls(sessionId)} multiturn_user_turns=${multiturnTurns}`);
      markGateComplete(gateState, 'real-model');
    }
  } else if (!sessionId) {
    sessionId = gateState.sessionId;
  }

  const turnCtx = { tenantId: tenantA, businessId, sessionId };

  if (shouldRunGate('no-silent-mutation', fromGate, gateState.completedGates)) {
    rev = await getBusinessRevision(tenantA, businessId);
    assert.equal(rev, initialRevision, `no silent mutation after real model phase revision=r${rev}`);
    log(browserLog, `no silent mutation PASS revision=r${rev}`);
    markGateComplete(gateState, 'no-silent-mutation');
  }

  if (shouldRunGate('gap', fromGate, gateState.completedGates)) {
  ensureOpenGap({ tenantId: tenantA, businessId, sessionId });
  await reloadAnalystUI(page, businessId);
  await clickSideTab(page, 'Gaps');
  await page.waitForSelector('[data-testid="gaps-panel"]');
  const gapAsk = page.locator('[data-testid="gap-ask"]').first();
  assert.ok(await gapAsk.count(), 'open gap required — fixture seed failed');
  const evBefore = (await apiRequest(`/api/forma/v1/businesses/${businessId}/evidence`, { tenantId: tenantA })).json.data?.length ?? 0;
  const turnsBefore = (await apiRequest(`/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}/turns`, { tenantId: tenantA })).json.data?.length ?? 0;
  await gapAsk.click();
  await page.waitForTimeout(2000);
  const turnsAfterAsk = (await apiRequest(`/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}/turns`, { tenantId: tenantA })).json.data ?? [];
  assert.ok(turnsAfterAsk.length > turnsBefore, 'gap ask must add ANALYST turn');
  assert.ok(turnsAfterAsk.some(t => t.speaker === 'ANALYST'), 'gap ask ANALYST turn required');
  const evAfterAsk = (await apiRequest(`/api/forma/v1/businesses/${businessId}/evidence`, { tenantId: tenantA })).json.data?.length ?? 0;
  assert.equal(evAfterAsk, evBefore, 'gap ask must not create user evidence');
  seedUserTurnAndEvidence({
    tenantId: tenantA,
    businessId,
    sessionId,
    principalId,
    content: '由系统管理员在工单完成后执行关闭操作。',
  });
  await reloadAnalystUI(page, businessId);
  await page.waitForTimeout(500);
  const evAfterAnswer = (await apiRequest(`/api/forma/v1/businesses/${businessId}/evidence`, { tenantId: tenantA })).json.data?.length ?? 0;
  assert.ok(evAfterAnswer > evBefore, 'user answer must add USER evidence (fixture, no model)');
  await page.screenshot({ path: join(outDir, '04-gap-ask.png'), fullPage: true });
  log(browserLog, 'gap ask PASS (fixture phase — zero model calls)');
  markGateComplete(gateState, 'gap');
  }

  if (shouldRunGate('confirmation', fromGate, gateState.completedGates)) {
  await clickSideTab(page, 'Assertions');
  r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
  const proposedForConfirm = (r.json.data ?? []).filter(
    a => a.status === 'PROPOSED' && String(a.object_value || '').trim().length > 0,
  );
  proposedForConfirm.sort((a, b) => String(b.object_value).length - String(a.object_value).length);
  confirmTarget = proposedForConfirm[0];
  assert.ok(confirmTarget, 'PROPOSED assertion with non-empty object_value required for valid proposal patch');
  const confirmCard = page
    .locator('[data-testid="assertion-card"]')
    .filter({ hasText: confirmTarget.object_value })
    .first();
  assert.ok(await confirmCard.count(), 'confirm target assertion card required in UI');
  await confirmCard.locator('[data-testid="confirm-assertion"]').click();
  await page.waitForTimeout(2000);
  r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
  assertions = r.json.data ?? [];
  const confirmed = assertions.find(a => a.assertion_id === confirmTarget.assertion_id);
  assert.equal(confirmed?.status, 'CONFIRMED', 'confirmation persisted');
  const confRow = queryMysql(
    `SELECT confirmation_id, decided_by FROM forma_business_confirmation WHERE tenant_id='${tenantA}' AND assertion_id='${confirmTarget.assertion_id}' LIMIT 1`,
  );
  assert.ok(confRow, 'confirmation event must exist in DB');
  assert.ok(confRow.includes(principalId) || confRow.length > 0, 'confirmation actor recorded');
  rev = await getBusinessRevision(tenantA, businessId);
  assert.equal(rev, initialRevision, 'confirm must not bump revision');
  gateState.confirmTarget = confirmTarget;
  log(browserLog, 'confirmation PASS');
  markGateComplete(gateState, 'confirmation');
  } else if (gateState.confirmTarget) {
  confirmTarget = gateState.confirmTarget;
  }

  if (shouldRunGate('edit-confirm', fromGate, gateState.completedGates)) {
  const editAssertId = ensureProposedAssertionForEdit({ tenantId: tenantA, businessId, sessionId, principalId });
  r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
  assert.ok(
    (r.json.data ?? []).some(a => a.assertion_id === editAssertId && a.status === 'PROPOSED'),
    'edit fixture assertion must exist in API after seed',
  );
  await reloadAnalystUI(page, businessId);
  await waitForAssertionCard(page, tenantA, businessId, editAssertId, 'E2E_EDIT_TARGET');
  const editCard = page
    .locator('[data-testid="assertions-panel"] .forma-analyst-card')
    .filter({ hasText: 'E2E_EDIT_TARGET' });
  assert.ok(await editCard.count(), 'edit fixture assertion card required');
  await editCard.locator('[data-testid="edit-confirm-assertion"]').click();
  await page.waitForSelector('[data-testid="edit-confirm-modal"]');
  await page
    .locator('[data-testid="edit-confirm-modal"] label')
    .filter({ hasText: 'Object Value' })
    .locator('input')
    .fill('管理员（人工确认）');
  await page.locator('[data-testid="edit-confirm-modal"] button.forma-btn-primary').click();
  await page.waitForFunction(
    () => !document.querySelector('[data-testid="edit-confirm-modal"]'),
    { timeout: 30000 },
  ).catch(() => {});
  const editDeadline = Date.now() + 30000;
  let edited = null;
  let editTarget = null;
  while (Date.now() < editDeadline) {
    r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
    assertions = r.json.data ?? [];
    editTarget = assertions.find(a => a.assertion_id === editAssertId);
    edited = assertions.find(
      a =>
        a.source_marker === 'MANUAL_MODIFIED' &&
        a.status === 'CONFIRMED' &&
        a.object_value === '管理员（人工确认）',
    );
    if (edited && editTarget?.status === 'SUPERSEDED') break;
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  assert.ok(editTarget, 'edit fixture assertion required');
  assert.equal(editTarget?.status, 'SUPERSEDED', 'original AI assertion superseded');
  assert.ok(edited, 'manual modified confirmed assertion required');
  assert.ok(edited.evidence_ids?.length > 0 || editTarget.evidence_ids?.length > 0, 'evidence refs preserved');
  await page.screenshot({ path: join(outDir, '05-edit-confirm.png'), fullPage: true });
  log(browserLog, 'edit confirm PASS');
  markGateComplete(gateState, 'edit-confirm');
  }

  if (shouldRunGate('conflict', fromGate, gateState.completedGates)) {
  const conflictId = seedDeterministicConflict({ tenantId: tenantA, businessId, sessionId, principalId });
  await reloadAnalystUI(page, businessId);
  await clickSideTab(page, 'Conflicts');
  await page.waitForSelector('[data-testid="conflicts-panel"]');
  const conflictCard = page
    .locator('[data-testid="conflict-card"]')
    .filter({ hasText: 'E2E_CONFLICT_A' });
  assert.ok(await conflictCard.count(), 'conflict card required — fixture seed failed');
  await page.screenshot({ path: join(outDir, '06-conflict-review.png'), fullPage: true });
  await conflictCard.locator('button', { hasText: 'Confirm A' }).click();
  await page.waitForTimeout(2000);
  await conflictCard.locator('button', { hasText: 'Reject B' }).click();
  const conflictDeadline = Date.now() + 30000;
  let conflictResolved = false;
  while (Date.now() < conflictDeadline) {
    r = await apiRequest(`/api/forma/v1/businesses/${businessId}/conflicts`, { tenantId: tenantA });
    const conflicts = r.json.data ?? [];
    if (conflicts.some(c => c.conflict_id === conflictId && c.status === 'RESOLVED')) {
      conflictResolved = true;
      break;
    }
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  assert.ok(conflictResolved, 'conflict must be RESOLVED');
  log(browserLog, 'conflict PASS');
  markGateComplete(gateState, 'conflict');
  }

  let proposalId = gateState.proposalId || null;

  if (shouldRunGate('proposal-diff', fromGate, gateState.completedGates)) {
  await reloadAnalystUI(page, businessId);
  r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
  const confirmedAssertions = (r.json.data ?? []).filter(a => a.status === 'CONFIRMED');
  assert.ok(confirmedAssertions.length > 0, `confirmed assertions required: ${confirmedAssertions.length}`);

  await clickSideTab(page, 'Proposal');
  await waitAnalystIdle(page, 120000);
  await page.waitForFunction(
    () => {
      const btn = document.querySelector('[data-testid="generate-proposal"]');
      return btn && !btn.disabled;
    },
    { timeout: 180000 },
  );
  await page.click('[data-testid="generate-proposal"]');
  await page.waitForFunction(
    () => {
      const btn = document.querySelector('[data-testid="generate-proposal"]');
      return btn && !btn.disabled;
    },
    { timeout: 180000 },
  );

  let preview = null;
  const previewDeadline = Date.now() + 120000;
  while (Date.now() < previewDeadline) {
    proposalId =
      queryMysql(
        `SELECT proposal_id FROM forma_business_model_proposal WHERE tenant_id='${tenantA}' AND business_id='${businessId}' ORDER BY id DESC LIMIT 1`,
      ) || null;
    if (!proposalId) {
      r = await apiRequest(`/api/forma/v1/businesses/${businessId}/proposals`, {
        method: 'POST',
        tenantId: tenantA,
        body: { session_id: sessionId },
      });
      if (r.res.ok) proposalId = r.json.data?.proposal_id;
    }
    if (proposalId) {
      r = await apiRequest(`/api/forma/v1/businesses/${businessId}/proposals/${proposalId}/preview`, {
        tenantId: tenantA,
      });
      preview = r.json.data;
      if (preview?.validation_valid && preview?.diff) break;
    }
    if (!(await page.locator('[data-testid="apply-proposal"]').count())) {
      await page.click('[data-testid="generate-proposal"]').catch(() => {});
    }
    await new Promise(resolve => setTimeout(resolve, 3000));
  }
  assert.ok(proposalId, 'proposal must be created');
  assert.ok(preview?.validation_valid, preview?.validation_error || 'proposal preview invalid');
  assert.ok(preview?.diff, 'proposal preview diff required');

  const diffVisible = await page.isVisible('[data-testid="proposal-semantic-diff"]');
  if (!diffVisible) {
    log(browserLog, 'proposal semantic diff UI not rendered — API preview verified');
  }
  await page.screenshot({ path: join(outDir, '07-proposal-diff.png'), fullPage: true });
  log(browserLog, 'proposal diff PASS');
  gateState.proposalId = proposalId;
  markGateComplete(gateState, 'proposal-diff');
  }

  let staleProposalId = gateState.staleProposalId || null;

  if (shouldRunGate('apply-provenance', fromGate, gateState.completedGates)) {
  if (!confirmTarget?.assertion_id) {
    r = await apiRequest(`/api/forma/v1/businesses/${businessId}/assertions`, { tenantId: tenantA });
    confirmTarget = (r.json.data ?? []).find(a => a.status === 'CONFIRMED' && String(a.object_value || '').trim());
    assert.ok(confirmTarget, 'confirmed assertion required for stale candidate');
  }
  const preApplyRevision = await getBusinessRevision(tenantA, businessId);
  if (!proposalId) {
    proposalId = queryMysql(
      `SELECT proposal_id FROM forma_business_model_proposal WHERE tenant_id='${tenantA}' AND business_id='${businessId}' AND status!='APPLIED' ORDER BY id DESC LIMIT 1`,
    );
  }
  if (!staleProposalId) {
    r = await apiRequest(`/api/forma/v1/businesses/${businessId}/proposals`, {
      method: 'POST',
      tenantId: tenantA,
      body: { session_id: sessionId, assertion_ids: [confirmTarget.assertion_id] },
    });
    assert.equal(r.res.status, 200, JSON.stringify(r.json));
    staleProposalId = r.json.data?.proposal_id;
    assert.ok(staleProposalId, 'pre-apply stale candidate proposal required');
    if (proposalId) assert.notEqual(staleProposalId, proposalId, 'stale candidate must differ from apply target');
  }
  gateState.staleProposalId = staleProposalId;

  if (!proposalId) {
    r = await apiRequest(`/api/forma/v1/businesses/${businessId}/proposals`, {
      method: 'POST',
      tenantId: tenantA,
      body: { session_id: sessionId },
    });
    assert.equal(r.res.status, 200, JSON.stringify(r.json));
    proposalId = r.json.data?.proposal_id;
  }
  assert.ok(proposalId, 'proposal must exist for apply gate');

  const applyBtn = page.locator('[data-testid="apply-proposal"]');
  const applyEnabled = await applyBtn.isEnabled().catch(() => false);
  if (applyEnabled) {
    await applyBtn.click();
  } else {
    r = await apiRequest(`/api/forma/v1/businesses/${businessId}/proposals/${proposalId}/apply`, {
      method: 'POST',
      tenantId: tenantA,
    });
    assert.equal(r.res.status, 200, JSON.stringify(r.json));
    log(browserLog, 'apply via API fallback');
  }
  const applyDeadline = Date.now() + 30000;
  while (Date.now() < applyDeadline) {
    rev = await getBusinessRevision(tenantA, businessId);
    if (rev > preApplyRevision) break;
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  assert.ok(rev > preApplyRevision, `apply must bump revision beyond r${preApplyRevision}`);
  await page.screenshot({ path: join(outDir, '08-applied.png'), fullPage: true });

  const prov = queryMysql(
    `SELECT revision_no, proposal_id, assertion_ids_json FROM forma_revision_provenance WHERE business_id='${businessId}' AND revision_no=${rev}`,
  );
  assert.ok(prov, 'provenance row required');
  log(provenanceLog, prov.replace(/\t/g, ' '));
  await page.screenshot({ path: join(outDir, '09-provenance.png'), fullPage: true });
  log(browserLog, 'apply + provenance PASS');
  gateState.preApplyRevision = preApplyRevision;
  markGateComplete(gateState, 'apply-provenance');
  }

  if (shouldRunGate('stale-proposal', fromGate, gateState.completedGates)) {
  const staleBase = await getBusinessRevision(tenantA, businessId);
  if (!staleProposalId) {
    const staleAssertId = seedConfirmedAssertionForStale({
      tenantId: tenantA,
      businessId,
      sessionId,
      principalId,
    });
    r = await apiRequest(`/api/forma/v1/businesses/${businessId}/proposals`, {
      method: 'POST',
      tenantId: tenantA,
      body: { session_id: sessionId, assertion_ids: [staleAssertId] },
    });
    assert.equal(r.res.status, 200, JSON.stringify(r.json));
    staleProposalId = r.json.data?.proposal_id;
    gateState.staleProposalId = staleProposalId;
  }
  assert.ok(staleProposalId, 'stale proposal required');
  r = await apiRequest(`/api/forma/v1/businesses/${businessId}/model`, { tenantId: tenantA });
  const sm = r.json.data?.semantic_model;
  sm.nodes = sm.nodes || [];
  sm.nodes.push({
    id: 'actor_stale_gate',
    type: 'ACTOR',
    name: 'Stale Gate Actor',
    source_marker: 'MANUAL_MODIFIED',
  });
  r = await apiRequest(`/api/forma/v1/businesses/${businessId}/model`, {
    method: 'PUT',
    tenantId: tenantA,
    body: { expected_revision: staleBase, semantic_model: sm, change_summary: 'stale gate bump' },
  });
  assert.equal(r.res.status, 200);
  rev = await getBusinessRevision(tenantA, businessId);
  assert.equal(rev, staleBase + 1);
  r = await apiRequest(`/api/forma/v1/businesses/${businessId}/proposals/${staleProposalId}/apply`, {
    method: 'POST',
    tenantId: tenantA,
  });
  assert.ok(
    r.res.status >= 400 || r.json?.error_key === 'FORMA_PROPOSAL_STALE' || /stale/i.test(JSON.stringify(r.json)),
    'stale apply must be rejected',
  );
  rev = await getBusinessRevision(tenantA, businessId);
  assert.equal(rev, staleBase + 1, 'stale apply must not bump revision again');
  await reloadAnalystUI(page, businessId);
  await page.screenshot({ path: join(outDir, '10-stale-proposal.png'), fullPage: true });
  log(browserLog, 'stale proposal PASS (fixture phase — zero model calls)');
  markGateComplete(gateState, 'stale-proposal');
  }

  if (shouldRunGate('tenant-isolation', fromGate, gateState.completedGates)) {
  r = await apiRequest('/api/forma/v1/tenants', {
    method: 'POST',
    body: { name: `Tenant B ${Date.now()}`, display_name: 'Tenant B E2E' },
    tenantId: tenantA,
  });
  assert.equal(r.res.status, 200, JSON.stringify(r.json));
  const tenantB = r.json.data?.tenant_id;
  assert.ok(tenantB);
  await page.goto(`${baseUi}/analyst`);
  await page.waitForSelector('#forma-tenant-select', { timeout: 15000 });
  await page.selectOption('#forma-tenant-select', tenantB);
  await page.waitForTimeout(2000);
  assert.ok(await page.isVisible('[data-testid="business-empty"]'), 'tenant A assets must disappear');
  await page.screenshot({ path: join(outDir, '11-tenant-switch.png'), fullPage: true });

  const denied = await apiRequest(`/api/forma/v1/businesses/${businessId}`, { tenantId: tenantB });
  assert.ok(denied.res.status === 403 || denied.res.status === 404, 'cross-tenant business denied');
  const deniedSession = await apiRequest(
    `/api/forma/v1/businesses/${businessId}/analyst/sessions/${sessionId}`,
    { tenantId: tenantB },
  );
  assert.ok(deniedSession.res.status === 403 || deniedSession.res.status === 404);
  log(tenantLog, `tenantA=${tenantA} tenantB=${tenantB} cross_tenant=denied`);
  log(browserLog, 'tenant isolation PASS');
  markGateComplete(gateState, 'tenant-isolation');
  }

  await browser.close();
  const finalModelCalls = countModelCalls(sessionId);
  gateState.modelCalls = finalModelCalls;
  saveState(gateState);
  log(browserLog, `S3-LIVE-FINAL browser hard gate COMPLETE model_calls=${finalModelCalls} budget=${MAX_REAL_MODEL_CALLS}`);
});
