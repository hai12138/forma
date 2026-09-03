/**
 * FORMA-S4-G6 PostgreSQL adapter smoke (no model calls).
 * Requires disposable Postgres reachable from forma-live-harness (forma-live-net).
 */
import assert from 'node:assert/strict';
import test from 'node:test';
import {
  api,
  registerLoginBootstrap,
  log,
  G6_SECRET,
  assertNoSecretMaterial,
} from './s4-g6-live-lib.mjs';

function dataOk(r, label) {
  assert.ok(r.status < 400, `${label} status=${r.status} body=${JSON.stringify(r.json)}`);
  assert.ok(r.json?.code === 0 || r.json?.code === undefined, `${label} code`);
  return r.json?.data;
}

const enabled = process.env.FORMA_LIVE_E2E === '1';
const password = process.env.FORMA_LIVE_PASSWORD || 'FormaE2E!23456';
const email = process.env.FORMA_G6_PG_EMAIL || `forma-g6-pg-${Date.now()}@example.com`;
const pgHost = process.env.FORMA_PG_HOST || 'docker-db_postgres-1';
const pgPass = process.env.FORMA_PG_PASSWORD || 'difyai123456';
const pgDb = process.env.FORMA_PG_DATABASE || 'forma_g6_pg';
const pgUser = process.env.FORMA_PG_USER || 'postgres';

test('S4-G6 postgres smoke', async t => {
  if (!enabled) {
    t.skip('FORMA_LIVE_E2E!=1');
    return;
  }
  const owner = await registerLoginBootstrap(email, password);
  const cred = dataOk(
    await api('/api/forma/v1/credentials', {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: { secret_type: 'password', secret: { password: pgPass } },
    }),
    'pg credential',
  );
  assert.ok(!JSON.stringify(cred).includes(pgPass));
  const src = dataOk(
    await api('/api/forma/v1/data-sources', {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: { name: 'G6 PG', source_type: 'RELATIONAL_DATABASE' },
    }),
    'pg source',
  );
  const conn = dataOk(
    await api(`/api/forma/v1/data-sources/${src.source_id}/connections`, {
      method: 'POST',
      cookies: owner.cookies,
      tenantId: owner.tenantId,
      body: {
        name: 'pg-dev',
        environment: 'DEV',
        adapter_type: 'POSTGRESQL',
        public_config: {
          host: pgHost,
          port: 5432,
          database: pgDb,
          username: pgUser,
        },
        credential_ref_id: cred.credential_ref_id,
      },
    }),
    'pg connection',
  );
  dataOk(
    await api(
      `/api/forma/v1/data-sources/${src.source_id}/connections/${conn.connection_id}/test`,
      { method: 'POST', cookies: owner.cookies, tenantId: owner.tenantId, body: {} },
    ),
    'pg test',
  );
  const discovered = dataOk(
    await api(
      `/api/forma/v1/data-sources/${src.source_id}/connections/${conn.connection_id}/discover`,
      { method: 'POST', cookies: owner.cookies, tenantId: owner.tenantId, body: {} },
    ),
    'pg discover',
  );
  const assets = Array.isArray(discovered) ? discovered : discovered.assets || [];
  assert.ok(assets.some(a => /sample/i.test(a.name || '')));
  assert.ok(assets.some(a => /v_sample|VIEW/i.test(`${a.name} ${a.asset_type}`)));
  const table = assets.find(a => a.name === 'sample') || assets[0];
  const snap = dataOk(
    await api(
      `/api/forma/v1/data-sources/${src.source_id}/connections/${conn.connection_id}/assets/${table.asset_id}/capture-schema`,
      { method: 'POST', cookies: owner.cookies, tenantId: owner.tenantId, body: {} },
    ),
    'pg capture',
  );
  assert.ok(snap.fingerprint && snap.schema);
  assertNoSecretMaterial(JSON.stringify(snap), [G6_SECRET, pgPass]);
  log('POSTGRES_SMOKE=PASS');
});
