/**
 * FORMA-S4-G6-F3 — Auth boundary cleanup + exact-candidate browser acceptance.
 * REAL_MODEL_CALLS = 0. No mock login. No DOM hash forgery.
 *
 *   FORMA_LIVE_E2E=1
 *   FORMA_EXPECTED_CANDIDATE_SHA=<exact SHA>
 *   MAX_REAL_MODEL_CALLS=0
 *   node --test scripts/forma/s4-g6-f3-browser-e2e.mjs
 *
 * FIXTURE_SETUP_ONLY: SQL may seed PROPOSED + provenance refs; browser performs transitions.
 */
import assert from 'node:assert/strict';
import test from 'node:test';
import { createHash } from 'node:crypto';
import { mkdirSync, writeFileSync, readFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { chromium } from 'playwright';

import {
  api,
  assertNoSecretMaterial,
  baseApi,
  baseUi,
  ensureLabSchema,
  G6_SECRET,
  loginExisting,
  labBusinessModel,
  log,
  mysqlExec,
  procurementBusinessModel,
  putModel,
  registerLoginBootstrap,
  resultsDir,
  scanPathsForSecrets,
} from './s4-g6-live-lib.mjs';

const enabled = process.env.FORMA_LIVE_E2E === '1';
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';
const expectedSha = (process.env.FORMA_EXPECTED_CANDIDATE_SHA || '').trim();
const f3Evidence = join(resultsDir, 's4-g6-f3-ui');
const REAL_MODEL_CALLS = 0;
const repoRoot = join(resultsDir, '..', '..');

const summary = {
  CANDIDATE_SHA: '',
  GIT_WORKTREE_CLEAN: false,
  AUTH_BROWSER_FLOW: 'PENDING',
  REQUIREMENT_CONFIRM: 'PENDING',
  REQUIREMENT_REJECT: 'PENDING',
  REQUIREMENT_EDIT_CONFIRM: 'PENDING',
  MAPPING_CONFIRM: 'PENDING',
  MAPPING_REJECT: 'PENDING',
  MAPPING_EDIT_CONFIRM: 'PENDING',
  CONTRACT_VALIDATE: 'PENDING',
  CONTRACT_ACTIVATE: 'PENDING',
  DRIFT_COMPATIBLE: 'PENDING',
  DRIFT_BREAKING_TO_STALE: 'PENDING',
  GAP_BROWSER: 'PENDING',
  MEMBER_BROWSER: 'PENDING',
  TENANT_BROWSER: 'PENDING',
  BUSINESS_B_ACTIVE_CONTRACT: 'PENDING',
  SECRET_SCAN: 'PENDING',
  REAL_MODEL_CALLS: 0,
  SCREENSHOT_LIMITATION: null,
  FIXTURE_SETUP_ONLY: true,
  COZE_AUTH_CORE_CHANGE: 'NONE',
};

function dataOk(r, label) {
  assert.ok(r.status < 400, `${label} status=${r.status} body=${JSON.stringify(r.json)}`);
  return r.json?.data;
}

function fileHash(path) {
  return createHash('sha256').update(readFileSync(path)).digest('hex');
}

function assertExactCandidate() {
  const status = spawnSync('git', ['status', '--porcelain'], {
    encoding: 'utf8',
    cwd: repoRoot,
  });
  assert.equal(status.status, 0, 'git status failed');
  const porcelain = (status.stdout || '').trim();
  assert.equal(porcelain, '', `GIT_WORKTREE_CLEAN required before F3 browser; dirty:\n${porcelain}`);

  const head = spawnSync('git', ['rev-parse', 'HEAD'], { encoding: 'utf8', cwd: repoRoot }).stdout.trim();
  assert.ok(expectedSha, 'FORMA_EXPECTED_CANDIDATE_SHA is required');
  assert.equal(head, expectedSha, `HEAD ${head} != EXPECTED ${expectedSha}`);

  const cozeDiff = spawnSync(
    'git',
    [
      'diff',
      'forma-s4-frozen^{}',
      '--',
      'coze-studio/backend/api/handler/coze/',
      'coze-studio/backend/api/middleware/session.go',
    ],
    { encoding: 'utf8', cwd: repoRoot },
  );
  assert.equal((cozeDiff.stdout || '').trim(), '', 'COZE_AUTH_CORE_CHANGE must be NONE vs forma-s4-frozen');

  summary.CANDIDATE_SHA = head;
  summary.GIT_WORKTREE_CLEAN = true;
  return head;
}

async function shot(page, name) {
  mkdirSync(f3Evidence, { recursive: true });
  const path = join(f3Evidence, name);
  await page.screenshot({ path, fullPage: true });
  assert.ok(existsSync(path), name);
  assertNoSecretMaterial(await page.content(), [G6_SECRET, password]);
  return path;
}

function noteScreenshotLimitation(a, b, label) {
  if (fileHash(a) === fileHash(b)) {
    const msg = `${label}: screenshot hashes identical despite state change (report limitation; no DOM forgery)`;
    log(msg);
    summary.SCREENSHOT_LIMITATION = summary.SCREENSHOT_LIMITATION
      ? `${summary.SCREENSHOT_LIMITATION}; ${msg}`
      : msg;
  }
}

async function createManualRequirement(owner, businessId, modelRevision, semanticName, description) {
  const req = dataOk(
    await api(`/api/forma/v1/businesses/${businessId}/data-requirements`, {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: {
        business_model_revision: modelRevision,
        requirement_kind: 'ENTITY',
        semantic_name: semanticName,
        description,
        business_element_refs: [],
        requiredness: 'REQUIRED',
        freshness_requirement: 'DAILY',
        access_need: 'READ',
      },
    }),
    `req ${semanticName}`,
  );
  return req.requirement_id || req.requirement?.requirement_id;
}

async function confirmRequirement(owner, businessId, requirementId, reason = 'f3-seed-confirm') {
  dataOk(
    await api(`/api/forma/v1/businesses/${businessId}/data-requirements/${requirementId}/confirm`, {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: { reason },
    }),
    `confirm ${requirementId}`,
  );
}

/** FIXTURE_SETUP_ONLY: demote CONFIRMED→PROPOSED with analysis provenance (no transition under test). */
function fixtureProposeRequirement(requirementId, tenantId, businessId, modelRevision, actor) {
  const analysisRunId = `ar_f3_${requirementId.slice(-12)}`;
  mysqlExec(
    `INSERT INTO forma_data_requirement_analysis_run
      (analysis_run_id, tenant_id, business_id, business_model_revision, client_request_id, request_digest,
       status, model_ref, error_key, error_message_sanitized, retry_count, created_by, created_at, updated_at)
     VALUES
      ('${analysisRunId}', '${tenantId}', '${businessId}', ${Number(modelRevision)},
       'f3-fixture-${requirementId}', 'f3-digest-${requirementId}',
       'SUCCEEDED', 'fixture://none', '', '', 0, '${actor}', UTC_TIMESTAMP(3), UTC_TIMESTAMP(3))
     ON DUPLICATE KEY UPDATE status='SUCCEEDED'`,
  );
  mysqlExec(
    `UPDATE forma_data_requirement
     SET status='PROPOSED', source='AI_GENERATED', analysis_run_id='${analysisRunId}'
     WHERE requirement_id='${requirementId}'`,
  );
  const after = mysqlExec(
    `SELECT requirement_id,status,source,analysis_run_id FROM forma_data_requirement WHERE requirement_id='${requirementId}'`,
  );
  assert.match(after, /PROPOSED/, `fixture req ${requirementId}`);
  return analysisRunId;
}

async function createManualMapping(owner, businessId, modelRevision, body) {
  const mapCreate = await api(`/api/forma/v1/businesses/${businessId}/semantic-mappings`, {
    method: 'POST',
    cookies: owner.cookies,
    tenantId: owner.tenantId,
    body: {
      business_model_revision: modelRevision,
      mapping_type: 'DIRECT',
      transform_spec: { type: 'DIRECT' },
      confidence: 0.9,
      reason: 'f3-seed',
      ...body,
    },
  });
  assert.ok(
    mapCreate.status < 400,
    `mapping create HTTP ${mapCreate.status} ${JSON.stringify(mapCreate.json)}`,
  );
  return mapCreate.json?.data?.mapping_id || mapCreate.json?.data?.mapping?.mapping_id;
}

async function confirmMapping(owner, businessId, mappingId, reason = 'f3-seed-confirm-map') {
  dataOk(
    await api(`/api/forma/v1/businesses/${businessId}/semantic-mappings/${mappingId}/confirm`, {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: { reason },
    }),
    `confirm map ${mappingId}`,
  );
}

function fixtureProposeMapping(mappingId, tenantId, businessId, modelRevision, actor) {
  const analysisRunId = `mar_f3_${mappingId.slice(-12)}`;
  mysqlExec(
    `INSERT INTO forma_data_semantic_mapping_analysis_run
      (analysis_run_id, tenant_id, business_id, business_model_revision, client_request_id, request_digest,
       request_json, status, model_ref, error_key, error_message_sanitized, retry_count,
       execution_generation, created_by, created_at, updated_at)
     VALUES
      ('${analysisRunId}', '${tenantId}', '${businessId}', ${Number(modelRevision)},
       'f3-map-fixture-${mappingId}', 'f3-map-digest-${mappingId}',
       '{}', 'SUCCEEDED', 'fixture://none', '', '', 0,
       1, '${actor}', UTC_TIMESTAMP(3), UTC_TIMESTAMP(3))
     ON DUPLICATE KEY UPDATE status='SUCCEEDED'`,
  );
  mysqlExec(
    `UPDATE forma_data_semantic_mapping
     SET status='PROPOSED', source='AI_GENERATED', analysis_run_id='${analysisRunId}'
     WHERE mapping_id='${mappingId}'`,
  );
  const after = mysqlExec(
    `SELECT mapping_id,status,source FROM forma_data_semantic_mapping WHERE mapping_id='${mappingId}'`,
  );
  assert.match(after, /PROPOSED/, `fixture map ${mappingId}`);
  return analysisRunId;
}

async function getRequirement(owner, businessId, modelRevision, requirementId) {
  const list = dataOk(
    await api(
      `/api/forma/v1/businesses/${businessId}/data-requirements?business_model_revision=${modelRevision}`,
      {
        cookies: owner.cookies,
        tenantId: owner.tenantId,
      },
    ),
    'list reqs',
  );
  const hit = (list || []).find(r => r.requirement_id === requirementId);
  assert.ok(hit, `requirement ${requirementId} missing`);
  return hit;
}

async function getMapping(owner, businessId, modelRevision, mappingId) {
  const list = dataOk(
    await api(
      `/api/forma/v1/businesses/${businessId}/semantic-mappings?business_model_revision=${modelRevision}`,
      { cookies: owner.cookies, tenantId: owner.tenantId },
    ),
    'list maps',
  );
  const hit = (list || []).find(m => m.mapping_id === mappingId);
  assert.ok(hit, `mapping ${mappingId} missing`);
  return hit;
}

async function seedLabPlane(owner, label) {
  ensureLabSchema();
  // Reset lab table to known schema for drift tests.
  mysqlExec(`DROP TABLE IF EXISTS assay_result; DROP TABLE IF EXISTS sample;`, 'forma_g6_lab');
  ensureLabSchema();

  const bizResp = await api('/api/forma/v1/businesses', {
    method: 'POST',
    cookies: owner.cookies,
    tenantId: owner.tenantId,
    body: { name: `${label} ${Date.now()}`, description: 'f3 browser' },
  });
  const biz = dataOk(bizResp, 'create business');
  const businessId = biz?.business_id || biz?.business?.business_id;
  const expectedRev = biz?.current_revision ?? biz?.business?.current_revision ?? 0;
  const putResp = await putModel(
    owner.cookies,
    owner.tenantId,
    businessId,
    labBusinessModel,
    expectedRev,
  );
  dataOk(putResp, 'put model');
  const modelRevision =
    putResp.json?.data?.current_revision ??
    putResp.json?.data?.business?.current_revision ??
    expectedRev + 1;

  const cred = dataOk(
    await api('/api/forma/v1/credentials', {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: {
        secret_type: 'password',
        secret: { password: process.env.FORMA_MYSQL_PASSWORD || 'coze123' },
      },
    }),
    'credential',
  );
  const source = dataOk(
    await api('/api/forma/v1/data-sources', {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: { name: 'lab-mysql', source_type: 'RELATIONAL_DATABASE' },
    }),
    'source',
  );
  const sourceId = source.source_id;
  const conn = dataOk(
    await api(`/api/forma/v1/data-sources/${sourceId}/connections`, {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: {
        name: 'lab-dev',
        environment: 'DEV',
        adapter_type: 'MYSQL',
        credential_ref_id: cred.credential_ref_id,
        public_config: {
          host: process.env.FORMA_MYSQL_HOST_FROM_BACKEND || 'mysql',
          port: 3306,
          database: 'forma_g6_lab',
          username: 'coze',
          ssl_mode: 'DISABLE',
        },
      },
    }),
    'connection',
  );
  const connectionId = conn.connection_id;
  dataOk(
    await api(`/api/forma/v1/data-sources/${sourceId}/connections/${connectionId}/discover`, {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: {},
    }),
    'discover',
  );
  const assets = dataOk(
    await api(`/api/forma/v1/data-sources/${sourceId}/assets`, {
      cookies: owner.cookies,
      tenantId: owner.tenantId,
    }),
    'assets',
  );
  const asset = (assets || []).find(a => (a.name || '').includes('sample')) || assets?.[0];
  assert.ok(asset?.asset_id, 'asset');
  const snap1 = dataOk(
    await api(
      `/api/forma/v1/data-sources/${sourceId}/connections/${connectionId}/assets/${asset.asset_id}/capture-schema`,
      { method: 'POST', cookies: owner.cookies, tenantId: owner.tenantId, body: {} },
    ),
    'snap1',
  );
  const fields = snap1.schema?.fields || snap1.Schema?.Fields || [];
  const byName = name =>
    fields.find(f => (f.name || '') === name || (f.path || '').endsWith(`.${name}`));
  const tempField = byName('temperature_c') || fields.find(f => (f.path || f.name || '').includes('temperature'));
  const fieldPath = tempField?.path || tempField?.name;
  assert.ok(fieldPath, `temperature field missing in snapshot: ${JSON.stringify(fields).slice(0, 500)}`);
  const pathFor = name => {
    const f = byName(name);
    assert.ok(f?.path || f?.name, `field ${name} missing`);
    return f.path || f.name;
  };
  log(`F3 seed fieldPath=${fieldPath}`);

  return {
    businessId,
    businessName: biz?.name || biz?.business?.name || label,
    sourceId,
    connectionId,
    assetId: asset.asset_id,
    snapPinned: snap1.snapshot_id,
    modelRevision,
    fieldPath,
    pathSampleId: pathFor('sample_id'),
    pathBatchId: pathFor('batch_id'),
    pathStatus: pathFor('status'),
    columnTemperature: 'temperature_c',
  };
}

test('S4-G6-F3 exact-candidate browser acceptance', async t => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }

  mkdirSync(f3Evidence, { recursive: true });
  mkdirSync(resultsDir, { recursive: true });

  // Exact-candidate gate MUST run before any evidence/log writes that dirty the tree.
  const candidateSha = assertExactCandidate();
  writeFileSync(join(resultsDir, 's4-g6-f3-browser-e2e.log'), '', 'utf8');
  log(`F3_BROWSER UI=${baseUi} API=${baseApi} CANDIDATE_SHA=${candidateSha}`);

  const email = process.env.FORMA_G6_F3_EMAIL || `forma-g6-f3-${Date.now()}@example.com`;
  const emailB = process.env.FORMA_G6_F3_EMAIL_B || `forma-g6-f3-b-${Date.now()}@example.com`;
  const emailMember =
    process.env.FORMA_G6_F3_EMAIL_M || `forma-g6-f3-m-${Date.now()}@example.com`;

  const owner = await registerLoginBootstrap(email, password);
  const ownerB = await registerLoginBootstrap(emailB, password);
  const memberUser = await registerLoginBootstrap(emailMember, password);
  dataOk(
    await api(`/api/forma/v1/tenants/${owner.tenantId}/members`, {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: { principal_id: memberUser.principalId, role: 'MEMBER' },
    }),
    'add member',
  );

  const browser = await chromium.launch({ headless: true });

  await t.test('AUTH_BROWSER_FLOW', async () => {
    const fresh = await browser.newContext({
      viewport: { width: 1440, height: 900 },
      baseURL: baseUi,
    });
    const page = await fresh.newPage();
    page.on('console', msg => assertNoSecretMaterial(msg.text(), [G6_SECRET, password]));

    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 60000 });
    await page.waitForSelector('[data-testid="forma-login-page"]', { timeout: 30000 });
    assert.equal(await page.locator('[data-testid="forma-sidebar"]').count(), 0);
    assert.equal(await page.locator('[data-testid="forma-app-shell"]').count(), 0);
    await shot(page, 'auth-01-login.png');

    await page.goto('/data', { waitUntil: 'domcontentloaded' });
    await page.waitForURL(/\/login/, { timeout: 15000 });

    await page.fill('[data-testid="login-email"]', email);
    await page.fill('[data-testid="login-password"]', password);
    await page.click('[data-testid="login-submit"]');
    await page.waitForSelector('[data-testid="forma-app-shell"]', { timeout: 45000 });
    await shot(page, 'auth-02-home.png');

    await page.goto('/data', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('[data-testid="data-plane-shell"]', { timeout: 30000 });
    await shot(page, 'auth-03-data.png');

    // Logout via Forma-owned adapter (network assert).
    await page.click('[data-testid="user-menu-trigger"]');
    await page.waitForSelector('[data-testid="logout-button"]', { timeout: 10000 });
    const [logoutResp] = await Promise.all([
      page.waitForResponse(
        r => r.url().includes('/api/forma/v1/auth/logout') && r.request().method() === 'POST',
        { timeout: 20000 },
      ),
      page.waitForSelector('[data-testid="forma-login-page"]', { timeout: 30000 }),
      page.click('[data-testid="logout-button"]'),
    ]);
    assert.ok(logoutResp.ok(), `forma logout HTTP ${logoutResp.status()}`);
    assert.equal(await page.locator('[data-testid="forma-sidebar"]').count(), 0);
    await shot(page, 'auth-04-logout.png');

    await page.goto('/data/requirements', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('[data-testid="forma-login-page"]', { timeout: 20000 });
    assert.ok(page.url().includes('/login'));
    await fresh.close();
    summary.AUTH_BROWSER_FLOW = 'PASS';
  });

  const ownerFresh = await registerLoginBootstrap(email, password);
  const seed = await seedLabPlane(ownerFresh, 'F3-Lab');
  const actor = ownerFresh.principalId || 'f3-actor';

  // Contract seed: confirmed req+mapping targeting temperature_c
  const contractReqId = await createManualRequirement(
    ownerFresh,
    seed.businessId,
    seed.modelRevision,
    'sample_temperature',
    'Temperature reading',
  );
  await confirmRequirement(ownerFresh, seed.businessId, contractReqId);
  const contractMapId = await createManualMapping(ownerFresh, seed.businessId, seed.modelRevision, {
    requirement_id: contractReqId,
    source_id: seed.sourceId,
    connection_id: seed.connectionId,
    asset_id: seed.assetId,
    schema_snapshot_id: seed.snapPinned,
    target_field_paths: [seed.fieldPath],
  });
  assert.ok(contractMapId, 'contract mapping');
  await confirmMapping(ownerFresh, seed.businessId, contractMapId);

  // Three independent PROPOSED requirements (FIXTURE_SETUP_ONLY)
  // Manual create already yields PROPOSED — enrich provenance via fixture.
  const reqA = await createManualRequirement(
    ownerFresh,
    seed.businessId,
    seed.modelRevision,
    'f3_req_confirm',
    'req A confirm',
  );
  const reqB = await createManualRequirement(
    ownerFresh,
    seed.businessId,
    seed.modelRevision,
    'f3_req_reject',
    'req B reject',
  );
  const reqC = await createManualRequirement(
    ownerFresh,
    seed.businessId,
    seed.modelRevision,
    'f3_req_edit',
    'req C edit',
  );
  fixtureProposeRequirement(reqA, ownerFresh.tenantId, seed.businessId, seed.modelRevision, actor);
  fixtureProposeRequirement(reqB, ownerFresh.tenantId, seed.businessId, seed.modelRevision, actor);
  fixtureProposeRequirement(reqC, ownerFresh.tenantId, seed.businessId, seed.modelRevision, actor);

  // Three independent PROPOSED mappings (each needs own confirmed requirement)
  const mapReqA = await createManualRequirement(
    ownerFresh,
    seed.businessId,
    seed.modelRevision,
    'f3_map_req_a',
    'map req A',
  );
  await confirmRequirement(ownerFresh, seed.businessId, mapReqA);
  const mapReqB = await createManualRequirement(
    ownerFresh,
    seed.businessId,
    seed.modelRevision,
    'f3_map_req_b',
    'map req B',
  );
  await confirmRequirement(ownerFresh, seed.businessId, mapReqB);
  const mapReqC = await createManualRequirement(
    ownerFresh,
    seed.businessId,
    seed.modelRevision,
    'f3_map_req_c',
    'map req C',
  );
  await confirmRequirement(ownerFresh, seed.businessId, mapReqC);
  const mapA = await createManualMapping(ownerFresh, seed.businessId, seed.modelRevision, {
    requirement_id: mapReqA,
    source_id: seed.sourceId,
    connection_id: seed.connectionId,
    asset_id: seed.assetId,
    schema_snapshot_id: seed.snapPinned,
    target_field_paths: [seed.pathSampleId],
  });
  const mapB = await createManualMapping(ownerFresh, seed.businessId, seed.modelRevision, {
    requirement_id: mapReqB,
    source_id: seed.sourceId,
    connection_id: seed.connectionId,
    asset_id: seed.assetId,
    schema_snapshot_id: seed.snapPinned,
    target_field_paths: [seed.pathBatchId],
  });
  const mapC = await createManualMapping(ownerFresh, seed.businessId, seed.modelRevision, {
    requirement_id: mapReqC,
    source_id: seed.sourceId,
    connection_id: seed.connectionId,
    asset_id: seed.assetId,
    schema_snapshot_id: seed.snapPinned,
    target_field_paths: [seed.pathStatus],
  });
  fixtureProposeMapping(mapA, ownerFresh.tenantId, seed.businessId, seed.modelRevision, actor);
  fixtureProposeMapping(mapB, ownerFresh.tenantId, seed.businessId, seed.modelRevision, actor);
  fixtureProposeMapping(mapC, ownerFresh.tenantId, seed.businessId, seed.modelRevision, actor);

  const ctx = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    baseURL: baseUi,
  });
  await ctx.addCookies([
    ...ownerFresh.cookies.entries(baseUi),
    ...ownerFresh.cookies.entries(baseApi),
  ]);
  const page = await ctx.newPage();

  await t.test('REQUIREMENT_BROWSER_FLOW', async () => {
    await page.goto(`/data/requirements?businessId=${seed.businessId}`, {
      waitUntil: 'networkidle',
      timeout: 60000,
    });
    await page.waitForSelector('[data-testid="data-requirements-page"]', { timeout: 30000 });
    await page.waitForTimeout(1200);

    const confirmBtn = page.locator(
      `[data-testid="requirement-card"]:has-text("f3_req_confirm") [data-testid="confirm-requirement"]`,
    );
    assert.equal(await confirmBtn.count(), 1, 'REQUIREMENT_CONFIRM button missing → FAIL');
    const propShot = await shot(page, 'req-ai-proposal.png');
    await confirmBtn.click();
    await page.waitForTimeout(1200);
    const afterA = await getRequirement(ownerFresh, seed.businessId, seed.modelRevision, reqA);
    assert.equal(afterA.status, 'CONFIRMED');
    summary.REQUIREMENT_CONFIRM = 'PASS';
    const confShot = await shot(page, 'req-human-confirm.png');
    noteScreenshotLimitation(propShot, confShot, 'requirement confirm');

    const rejectBtn = page.locator(
      `[data-testid="requirement-card"]:has-text("f3_req_reject") [data-testid="reject-requirement"]`,
    );
    assert.equal(await rejectBtn.count(), 1, 'REQUIREMENT_REJECT button missing → FAIL');
    await rejectBtn.click();
    await page.waitForTimeout(1200);
    const afterB = await getRequirement(ownerFresh, seed.businessId, seed.modelRevision, reqB);
    assert.equal(afterB.status, 'REJECTED');
    summary.REQUIREMENT_REJECT = 'PASS';

    const editBtn = page.locator(
      `[data-testid="requirement-card"]:has-text("f3_req_edit") [data-testid="edit-confirm-requirement"]`,
    );
    assert.equal(await editBtn.count(), 1, 'REQUIREMENT_EDIT_CONFIRM button missing → FAIL');
    await editBtn.click();
    await page.fill(
      `[data-testid="requirement-card"]:has-text("f3_req_edit") input`,
      'f3_req_edit_modified',
    );
    await page.locator('[data-testid="submit-edit-confirm"]').click();
    await page.waitForTimeout(1500);
    const afterCOrig = await getRequirement(ownerFresh, seed.businessId, seed.modelRevision, reqC);
    assert.equal(afterCOrig.status, 'SUPERSEDED');
    const listed = dataOk(
      await api(
        `/api/forma/v1/businesses/${seed.businessId}/data-requirements?business_model_revision=${seed.modelRevision}`,
        {
          cookies: ownerFresh.cookies,
          tenantId: ownerFresh.tenantId,
        },
      ),
      'reqs after edit',
    );
    const replacement = (listed || []).find(
      r =>
        r.semantic_name === 'f3_req_edit_modified' &&
        r.derived_from_requirement_id === reqC &&
        r.status === 'CONFIRMED',
    );
    assert.ok(replacement, 'edit-confirm replacement missing');
    assert.equal(replacement.source, 'MANUAL_MODIFIED');
    const decisions = dataOk(
      await api(`/api/forma/v1/businesses/${seed.businessId}/data-requirements/${reqC}/decisions`, {
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
      }),
      'req decisions',
    );
    assert.ok(
      (decisions || []).some(d => d.action === 'EDIT_CONFIRM' || d.Action === 'EDIT_CONFIRM'),
      `EDIT_CONFIRM decision missing: ${JSON.stringify(decisions)}`,
    );
    summary.REQUIREMENT_EDIT_CONFIRM = 'PASS';
  });

  await t.test('MAPPING_BROWSER_FLOW', async () => {
    await page.goto(`/data/mappings?businessId=${seed.businessId}`, {
      waitUntil: 'networkidle',
      timeout: 60000,
    });
    await page.waitForSelector('[data-testid="mapping-studio"]', { timeout: 30000 });
    await page.waitForTimeout(1200);

    // Select requirement then act on mapping cards — open map A via left list if needed.
    async function openRequirement(reqId) {
      const loc = page.locator(`[data-testid="mapping-requirement-${reqId}"]`);
      if ((await loc.count()) > 0) {
        await loc.click();
        await page.waitForTimeout(500);
      }
    }

    await openRequirement(mapReqA);
    const confirmMap = page.locator('[data-testid="confirm-mapping"]');
    assert.ok((await confirmMap.count()) >= 1, 'MAPPING_CONFIRM button missing → FAIL');
    const mapAi = await shot(page, 'map-ai-proposal.png');
    await confirmMap.first().click();
    await page.waitForTimeout(1200);
    const afterMapA = await getMapping(ownerFresh, seed.businessId, seed.modelRevision, mapA);
    assert.equal(afterMapA.status, 'CONFIRMED');
    summary.MAPPING_CONFIRM = 'PASS';

    await openRequirement(mapReqB);
    const rejectMap = page.locator('[data-testid="reject-mapping"]');
    assert.ok((await rejectMap.count()) >= 1, 'MAPPING_REJECT button missing → FAIL');
    await rejectMap.first().click();
    await page.waitForTimeout(1200);
    const afterMapB = await getMapping(ownerFresh, seed.businessId, seed.modelRevision, mapB);
    assert.equal(afterMapB.status, 'REJECTED');
    summary.MAPPING_REJECT = 'PASS';

    await openRequirement(mapReqC);
    const editMap = page.locator('[data-testid="edit-confirm-mapping"]');
    assert.ok((await editMap.count()) >= 1, 'MAPPING_EDIT_CONFIRM button missing → FAIL');
    await editMap.first().click();
    await page.waitForSelector('[data-testid="edit-confirm-mapping-panel"]', { timeout: 10000 });
    // Keep controlled DIRECT DSL (avoid CAST form gaps); submit EditConfirm as-is.
    const [editResp] = await Promise.all([
      page.waitForResponse(
        r => r.url().includes('/edit-confirm') && r.request().method() === 'POST',
        { timeout: 20000 },
      ),
      page.click('[data-testid="submit-edit-confirm-mapping"]'),
    ]);
    assert.ok(editResp.ok(), `edit-confirm mapping HTTP ${editResp.status()}`);
    await page.waitForTimeout(1200);
    const afterMapC = await getMapping(ownerFresh, seed.businessId, seed.modelRevision, mapC);
    assert.equal(afterMapC.status, 'SUPERSEDED');
    const maps = dataOk(
      await api(
        `/api/forma/v1/businesses/${seed.businessId}/semantic-mappings?business_model_revision=${seed.modelRevision}`,
        { cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId },
      ),
      'maps after edit',
    );
    const replacement = (maps || []).find(
      m => m.derived_from_mapping_id === mapC && m.status === 'CONFIRMED',
    );
    assert.ok(replacement, 'mapping edit-confirm replacement missing');
    assert.equal(replacement.source, 'MANUAL_MODIFIED');
    const mapDecisions = dataOk(
      await api(`/api/forma/v1/businesses/${seed.businessId}/semantic-mappings/${mapC}/decisions`, {
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
      }),
      'map decisions',
    );
    assert.ok(
      (mapDecisions || []).some(d => d.action === 'EDIT_CONFIRM' || d.Action === 'EDIT_CONFIRM'),
      `mapping EDIT_CONFIRM missing: ${JSON.stringify(mapDecisions)}`,
    );
    const mapHuman = await shot(page, 'map-human-confirm.png');
    noteScreenshotLimitation(mapAi, mapHuman, 'mapping edit-confirm');
    summary.MAPPING_EDIT_CONFIRM = 'PASS';
  });

  let contractId;
  let revisionId;
  let activeRevisionId;

  await t.test('CONTRACT_BROWSER_FLOW', async () => {
    const created = dataOk(
      await api(`/api/forma/v1/businesses/${seed.businessId}/data-contracts`, {
        method: 'POST',
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
        body: {
          business_model_revision: seed.modelRevision,
          name: 'F3 Contract Lab',
          description: 'f3',
          requirement_ids: [contractReqId],
          mapping_ids: [contractMapId],
          logical_schema: {
            fields: [
              {
                logical_key: 'f3_temp',
                semantic_name: 'sample_temperature',
                logical_type: 'DECIMAL',
                description: 'f3',
                requirement_id: contractReqId,
                nullable: true,
                classification: 'INTERNAL',
              },
            ],
          },
          query_capabilities: ['READ', 'FILTER'],
          filter_schema: { fields: [] },
          sort_schema: { fields: [] },
          pagination_policy: { default_limit: 50, max_limit: 100 },
          freshness_policy: 'DAILY',
          classification_policy: {},
        },
      }),
      'create contract',
    );
    contractId = created.contract?.contract_id || created.contract_id;
    revisionId = created.revision?.revision_id || created.revision_id;

    await page.goto(`/data/contracts/${contractId}?businessId=${seed.businessId}`, {
      waitUntil: 'domcontentloaded',
    });
    await page.waitForSelector('[data-testid="revisions-tab"], [data-testid="status-badge"]', {
      timeout: 30000,
    });
    if (await page.locator('[data-testid="revisions-tab"]').count()) {
      await page.click('[data-testid="revisions-tab"]');
    }
    const draftShot = await shot(page, 'contract-draft.png');
    await page.waitForSelector('[data-testid="validate-revision"]', { timeout: 20000 });
    const [valResp] = await Promise.all([
      page.waitForResponse(
        r => r.url().includes('/validate') && r.request().method() === 'POST',
        { timeout: 30000 },
      ),
      page.click('[data-testid="validate-revision"]'),
    ]);
    const valText = await valResp.text();
    assert.ok(valResp.ok(), `browser validate HTTP ${valResp.status()} ${valText}`);
    let valJson;
    try {
      valJson = JSON.parse(valText);
    } catch {
      valJson = {};
    }
    const revValidated = dataOk(
      await api(
        `/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}/revisions/${revisionId}`,
        { cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId },
      ),
      'rev validated',
    );
    assert.equal(
      revValidated.status || revValidated.revision?.status,
      'VALIDATED',
      `validate did not persist VALIDATED; resp=${valText} rev=${JSON.stringify(revValidated)}`,
    );
    summary.CONTRACT_VALIDATE = 'PASS';
    await page.reload({ waitUntil: 'networkidle' });
    if (await page.locator('[data-testid="revisions-tab"]').count()) {
      await page.click('[data-testid="revisions-tab"]');
    }
    await page.waitForFunction(
      () => !document.body.innerText.includes('加载中'),
      { timeout: 30000 },
    );
    await page.waitForSelector(
      '[data-testid="revision-card-VALIDATED"], [data-testid="status-badge"][data-status="VALIDATED"]',
      { timeout: 30000 },
    );
    const validatedDom =
      (await page.locator('[data-testid="revision-card-VALIDATED"]').count()) +
      (await page.locator('[data-testid="status-badge"][data-status="VALIDATED"]').count()) +
      (await page.getByText('已验证').count());
    assert.ok(
      validatedDom > 0,
      `DOM missing VALIDATED; body=${(await page.locator('body').innerText()).slice(0, 800)}`,
    );

    await page.waitForSelector('[data-testid="activate-revision"]', { timeout: 20000 });
    await Promise.all([
      page.waitForResponse(
        r => r.url().includes('/activate') && r.request().method() === 'POST',
        { timeout: 30000 },
      ),
      (async () => {
        await page.click('[data-testid="activate-revision"]');
        if (await page.locator('[data-testid="activate-confirm"]').count()) {
          await page.click('[data-testid="activate-confirm"]');
        }
      })(),
    ]);
    await page.reload({ waitUntil: 'domcontentloaded' });
    if (await page.locator('[data-testid="revisions-tab"]').count()) {
      await page.click('[data-testid="revisions-tab"]');
    }
    await page.waitForSelector(
      '[data-testid="revision-card-ACTIVE"], [data-testid="status-badge"][data-status="ACTIVE"]',
      { timeout: 20000 },
    );
    const revActive = dataOk(
      await api(
        `/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}/revisions/${revisionId}`,
        { cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId },
      ),
      'rev active',
    );
    assert.equal(revActive.status || revActive.revision?.status, 'ACTIVE');
    const ctr = dataOk(
      await api(`/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}`, {
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
      }),
      'contract active',
    );
    activeRevisionId = ctr.active_revision_id;
    assert.equal(activeRevisionId, revisionId);
    const activeShot = await shot(page, 'contract-active.png');
    noteScreenshotLimitation(draftShot, activeShot, 'contract activate');
    summary.CONTRACT_ACTIVATE = 'PASS';
  });

  await t.test('DRIFT_AND_GAP_BROWSER', async () => {
    const snapFresh = dataOk(
      await api(
        `/api/forma/v1/data-sources/${seed.sourceId}/connections/${seed.connectionId}/assets/${seed.assetId}/capture-schema`,
        { method: 'POST', cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId, body: {} },
      ),
      'snap fresh compatible',
    );

    await page.goto(`/data/health?businessId=${seed.businessId}`, {
      waitUntil: 'domcontentloaded',
    });
    await page.fill('[data-testid="health-contract-id"]', contractId);
    await page.fill('[data-testid="health-revision-id"]', activeRevisionId);
    await page.waitForSelector('[data-testid="drift-snapshot-picker"]', { timeout: 20000 });
    await page
      .locator(`[data-testid="fresh-snapshot-select-${seed.snapPinned}"]`)
      .selectOption(snapFresh.snapshot_id);
    await page.click('[data-testid="evaluate-drift"]');
    await page.waitForSelector('[data-testid="drift-severity-banner"]', { timeout: 20000 });
    const banner = await page.locator('[data-testid="drift-severity-banner"]').innerText();
    assert.match(banner, /COMPATIBLE|NO_CHANGE/);
    const ctrStill = dataOk(
      await api(`/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}`, {
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
      }),
      'contract after compatible',
    );
    assert.equal(ctrStill.active_revision_id, activeRevisionId);
    const compatShot = await shot(page, 'drift-compatible.png');
    summary.DRIFT_COMPATIBLE = 'PASS';

    // BREAKING: drop the real mapped field (temperature_c)
    const confirmedMap = await getMapping(
      ownerFresh,
      seed.businessId,
      seed.modelRevision,
      contractMapId,
    );
    const paths = confirmedMap.target_field_paths || confirmedMap.TargetFieldPaths || [seed.fieldPath];
    assert.ok(paths.includes(seed.fieldPath), `expected mapped path ${seed.fieldPath}`);
    mysqlExec(`ALTER TABLE sample DROP COLUMN ${seed.columnTemperature};`, 'forma_g6_lab');
    const snapBreak = dataOk(
      await api(
        `/api/forma/v1/data-sources/${seed.sourceId}/connections/${seed.connectionId}/assets/${seed.assetId}/capture-schema`,
        { method: 'POST', cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId, body: {} },
      ),
      'snap breaking',
    );
    await page.reload({ waitUntil: 'domcontentloaded' });
    await page.fill('[data-testid="health-contract-id"]', contractId);
    await page.fill('[data-testid="health-revision-id"]', activeRevisionId);
    await page.waitForSelector(`[data-testid="fresh-snapshot-select-${seed.snapPinned}"]`, {
      timeout: 20000,
    });
    await page
      .locator(`[data-testid="fresh-snapshot-select-${seed.snapPinned}"]`)
      .selectOption(snapBreak.snapshot_id);
    await page.click('[data-testid="evaluate-drift"]');
    await page.waitForTimeout(1500);
    await page.waitForFunction(
      () => {
        const el = document.querySelector('[data-testid="drift-severity-banner"]');
        return el && /BREAKING/.test(el.textContent || '');
      },
      { timeout: 20000 },
    );
    const breakShot = await shot(page, 'drift-breaking.png');
    noteScreenshotLimitation(compatShot, breakShot, 'drift breaking');

    const ctrAfter = dataOk(
      await api(`/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}`, {
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
      }),
      'contract after breaking',
    );
    assert.equal(ctrAfter.active_revision_id || '', '');
    const revStale = dataOk(
      await api(
        `/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}/revisions/${revisionId}`,
        { cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId },
      ),
      'rev stale',
    );
    assert.equal(revStale.status || revStale.revision?.status, 'STALE');
    const desc = await api(
      `/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}/active-descriptor`,
      { cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId },
    );
    assert.ok(desc.status >= 400, 'active descriptor must fail');
    const body = JSON.stringify(desc.json || {});
    assert.match(body, /CONTRACT_NOT_ACTIVE|FORMA_DATA_CONTRACT_NOT_ACTIVE|not active/i);
    summary.DRIFT_BREAKING_TO_STALE = 'PASS';

    // GAP against STALE revision still evaluates gap payload
    await page.click('[data-testid="evaluate-gap"]');
    await page.waitForSelector('[data-testid="gap-result-card"]', { timeout: 20000 });
    const gaps = dataOk(
      await api(
        `/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}/revisions/${revisionId}/gap-results`,
        { cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId },
      ),
      'gap results',
    );
    assert.ok((gaps || []).length > 0, 'gap results empty');
    const g0 = gaps[0];
    assert.ok(
      g0.FromBusinessRevision != null || g0.from_business_revision != null,
      'FromBusinessRevision',
    );
    assert.ok(
      g0.CurrentBusinessRevision != null || g0.current_business_revision != null,
      'CurrentBusinessRevision',
    );
    assert.ok(
      Array.isArray(g0.NewConfirmedRequirementIDs || g0.new_confirmed_requirement_ids),
      'NewConfirmedRequirementIDs',
    );
    assert.ok(
      Array.isArray(g0.UnmappedRequirementIDs || g0.unmapped_requirement_ids),
      'UnmappedRequirementIDs',
    );
    await shot(page, 'gap-result.png');
    summary.GAP_BROWSER = 'PASS';
  });

  await t.test('MEMBER_BROWSER', async () => {
    const memberCtx = await browser.newContext({
      viewport: { width: 1440, height: 900 },
      baseURL: baseUi,
    });
    const mpage = await memberCtx.newPage();
    await mpage.goto('/login', { waitUntil: 'domcontentloaded' });
    await mpage.fill('[data-testid="login-email"]', emailMember);
    await mpage.fill('[data-testid="login-password"]', password);
    await mpage.click('[data-testid="login-submit"]');
    await mpage.waitForSelector('[data-testid="forma-app-shell"]', { timeout: 45000 });

    // Ensure active tenant is Tenant A (owner), not bootstrap tenant.
    const select = mpage.locator('#forma-tenant-select');
    if ((await select.count()) > 0) {
      await select.selectOption(ownerFresh.tenantId);
      await mpage.waitForTimeout(1000);
    } else {
      await mpage.evaluate(tid => sessionStorage.setItem('forma.selectedTenantId', tid), ownerFresh.tenantId);
      await mpage.reload({ waitUntil: 'domcontentloaded' });
      await mpage.waitForSelector('[data-testid="forma-app-shell"]', { timeout: 30000 });
    }

    await mpage.goto(`/data/mappings?businessId=${seed.businessId}`, {
      waitUntil: 'domcontentloaded',
    });
    await mpage.waitForSelector('[data-testid="mapping-studio"]', { timeout: 30000 });
    assert.equal(await mpage.locator('[data-testid="confirm-mapping"]').count(), 0);
    assert.equal(await mpage.locator('[data-testid="edit-confirm-mapping"]').count(), 0);
    await shot(mpage, 'member-readonly.png');

    // Re-login AFTER browser session so API jar is not invalidated by UI login.
    const memberApi = await loginExisting(emailMember, password);
    const memberMe = await api('/api/forma/v1/me', { cookies: memberApi.cookies });
    const memberTenants = memberMe.json?.data?.tenants || [];
    const memberOnA = memberTenants.find(t => t.tenant_id === ownerFresh.tenantId);
    assert.ok(memberOnA, `member missing tenant A: ${JSON.stringify(memberTenants)}`);
    assert.equal(
      memberOnA.role,
      'MEMBER',
      `expected MEMBER on tenant A, got ${memberOnA.role}`,
    );
    const denied = await api(
      `/api/forma/v1/businesses/${seed.businessId}/data-requirements/${contractReqId}/confirm`,
      {
        method: 'POST',
        cookies: memberApi.cookies,
        tenantId: ownerFresh.tenantId,
        body: { reason: 'member-should-fail' },
      },
    );
    assert.ok(
      denied.status === 403 ||
        denied.status === 401 ||
        String(denied.json?.error_key || '').includes('FORBIDDEN') ||
        String(denied.json?.msg || '').toLowerCase().includes('forbidden'),
      `member mutation ${denied.status} role=${memberOnA.role} ${JSON.stringify(denied.json)}`,
    );
    await memberCtx.close();
    summary.MEMBER_BROWSER = 'PASS';
  });

  await t.test('TENANT_BROWSER', async () => {
    const bCtx = await browser.newContext({
      viewport: { width: 1440, height: 900 },
      baseURL: baseUi,
    });
    await bCtx.addCookies([
      ...ownerB.cookies.entries(baseUi),
      ...ownerB.cookies.entries(baseApi),
    ]);
    const bpage = await bCtx.newPage();
    const cross = await api(
      `/api/forma/v1/businesses/${seed.businessId}/data-contracts/${contractId}`,
      { cookies: ownerB.cookies, tenantId: ownerB.tenantId },
    );
    assert.ok(
      cross.status === 403 || cross.status === 404 || cross.json?.data == null,
      `tenant isolation API ${cross.status}`,
    );

    await bpage.goto(`/data/contracts/${contractId}?businessId=${seed.businessId}`, {
      waitUntil: 'domcontentloaded',
    });
    await bpage.waitForTimeout(1200);
    const body = await bpage.locator('body').innerText();
    assert.ok(!body.includes('F3 Contract Lab'), 'tenant B must not see contract name');
    assert.ok(!body.includes(G6_SECRET));
    // Deny leak of Tenant A revision lifecycle labels in isolation page when access fails.
    const leakedActiveCard = await bpage.locator('[data-testid="revision-card-ACTIVE"]').count();
    assert.equal(leakedActiveCard, 0, 'tenant B must not see Tenant A ACTIVE revision card');
    await shot(bpage, 'tenant-isolation.png');
    await bCtx.close();
    summary.TENANT_BROWSER = 'PASS';
  });

  await t.test('BUSINESS_B_ACTIVE_CONTRACT', async () => {
    const bizB = dataOk(
      await api('/api/forma/v1/businesses', {
        method: 'POST',
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
        body: { name: `采购合同审批 ${Date.now()}`, description: 'procurement' },
      }),
      'biz B',
    );
    const businessBId = bizB.business_id || bizB.business?.business_id;
    const expectedRevB = bizB.current_revision ?? bizB.business?.current_revision ?? 0;
    const putB = await putModel(
      ownerFresh.cookies,
      ownerFresh.tenantId,
      businessBId,
      procurementBusinessModel,
      expectedRevB,
    );
    dataOk(putB, 'put B model');
    const revB =
      putB.json?.data?.current_revision ?? putB.json?.data?.business?.current_revision ?? 1;

    // Deterministic fixture — no model calls
    const reqBId = await createManualRequirement(
      ownerFresh,
      businessBId,
      revB,
      'procurement_amount',
      '采购金额',
    );
    await confirmRequirement(ownerFresh, businessBId, reqBId);
    // Reuse lab asset mapping is ok for activation path (fixture)
    const mapBId = await createManualMapping(ownerFresh, businessBId, revB, {
      requirement_id: reqBId,
      source_id: seed.sourceId,
      connection_id: seed.connectionId,
      asset_id: seed.assetId,
      schema_snapshot_id: seed.snapPinned,
      target_field_paths: [seed.pathSampleId],
    });
    await confirmMapping(ownerFresh, businessBId, mapBId);
    // Recapture schema after DROP may miss temperature; sample_id still exists.
    // If snapPinned schema still lists temperature but asset changed — binding uses snap id as pinned.
    const createdB = dataOk(
      await api(`/api/forma/v1/businesses/${businessBId}/data-contracts`, {
        method: 'POST',
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
        body: {
          business_model_revision: revB,
          name: '采购合同契约',
          description: 'biz-b',
          requirement_ids: [reqBId],
          mapping_ids: [mapBId],
          logical_schema: {
            fields: [
              {
                logical_key: 'amount',
                semantic_name: 'procurement_amount',
                logical_type: 'string',
                description: 'b',
                requirement_id: reqBId,
                nullable: true,
                classification: 'INTERNAL',
              },
            ],
          },
          query_capabilities: ['READ', 'FILTER'],
          filter_schema: { fields: [] },
          sort_schema: { fields: [] },
          pagination_policy: { default_limit: 50, max_limit: 100 },
          freshness_policy: 'DAILY',
          classification_policy: {},
        },
      }),
      'contract B',
    );
    const contractBId = createdB.contract?.contract_id || createdB.contract_id;
    const revisionBId = createdB.revision?.revision_id || createdB.revision_id;

    await page.goto(`/data/contracts/${contractBId}?businessId=${businessBId}`, {
      waitUntil: 'domcontentloaded',
    });
    await page.waitForSelector('[data-testid="contract-detail-page"]', { timeout: 30000 });
    if (await page.locator('[data-testid="revisions-tab"]').count()) {
      await page.click('[data-testid="revisions-tab"]');
    }
    await page.waitForSelector('[data-testid="validate-revision"]', { timeout: 20000 });
    const [valB] = await Promise.all([
      page.waitForResponse(
        r => r.url().includes('/validate') && r.request().method() === 'POST',
        { timeout: 30000 },
      ),
      page.click('[data-testid="validate-revision"]'),
    ]);
    assert.ok(valB.ok(), `biz B validate HTTP ${valB.status()} ${await valB.text()}`);
    const revBCheck = dataOk(
      await api(
        `/api/forma/v1/businesses/${businessBId}/data-contracts/${contractBId}/revisions/${revisionBId}`,
        { cookies: ownerFresh.cookies, tenantId: ownerFresh.tenantId },
      ),
      'biz B rev validated',
    );
    assert.equal(revBCheck.status || revBCheck.revision?.status, 'VALIDATED');
    await page.reload({ waitUntil: 'networkidle' });
    if (await page.locator('[data-testid="revisions-tab"]').count()) {
      await page.click('[data-testid="revisions-tab"]');
    }
    await page.waitForFunction(
      () => !document.body.innerText.includes('加载中'),
      { timeout: 30000 },
    );
    await page.waitForSelector(
      '[data-testid="revision-card-VALIDATED"], [data-testid="status-badge"][data-status="VALIDATED"]',
      { timeout: 30000 },
    );
    assert.ok(
      (await page.getByText('已验证').count()) +
        (await page.locator('[data-testid="revision-card-VALIDATED"]').count()) >
        0,
      'biz B DOM missing VALIDATED',
    );
    await page.waitForSelector('[data-testid="activate-revision"]', { timeout: 20000 });
    await page.click('[data-testid="activate-revision"]');
    if (await page.locator('[data-testid="activate-confirm"]').count()) {
      await page.click('[data-testid="activate-confirm"]');
    }
    await page.reload({ waitUntil: 'domcontentloaded' });
    if (await page.locator('[data-testid="revisions-tab"]').count()) {
      await page.click('[data-testid="revisions-tab"]');
    }
    await page.waitForSelector(
      '[data-testid="revision-card-ACTIVE"], [data-testid="status-badge"][data-status="ACTIVE"]',
      { timeout: 20000 },
    );
    const ctrB = dataOk(
      await api(`/api/forma/v1/businesses/${businessBId}/data-contracts/${contractBId}`, {
        cookies: ownerFresh.cookies,
        tenantId: ownerFresh.tenantId,
      }),
      'contract B active',
    );
    assert.equal(ctrB.active_revision_id, revisionBId);
    await shot(page, 'business-b-active-contract.png');
    summary.BUSINESS_B_ACTIVE_CONTRACT = 'PASS';
  });

  await ctx.close();
  await browser.close();

  scanPathsForSecrets([join(resultsDir, 's4-g6-f3-browser-e2e.log')]);
  for (const f of ['auth-01-login.png', 'req-ai-proposal.png', 'map-human-confirm.png']) {
    const p = join(f3Evidence, f);
    if (existsSync(p)) assertNoSecretMaterial(readFileSync(p), [G6_SECRET]);
  }
  summary.SECRET_SCAN = 'PASS';
  summary.REAL_MODEL_CALLS = REAL_MODEL_CALLS;

  const summaryPath = join(resultsDir, 's4-g6-f3-browser-summary.json');
  writeFileSync(summaryPath, `${JSON.stringify(summary, null, 2)}\n`, 'utf8');
  const browserSummarySha = createHash('sha256').update(readFileSync(summaryPath)).digest('hex');
  writeFileSync(
    join(resultsDir, 's4-g6-f3-browser-summary.sha256'),
    `${browserSummarySha}\n`,
    'utf8',
  );
  log(`F3_BROWSER_ACCEPTANCE=PASS candidate=${candidateSha} summary_sha=${browserSummarySha}`);
});
