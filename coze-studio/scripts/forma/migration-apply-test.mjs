/**
 * Forma S0-B-G2 — portable migration integration test (CASE A/B/C).
 *
 * Topology (no host ports, no host.docker.internal):
 *   Docker network forma-mig-net-<id>
 *     ├── MySQL (alias: mysql)  image: mysql:8.4.5 (Coze baseline)
 *     └── Atlas (ephemeral)     connects to mysql:3306
 *
 * Creates users with MySQL 8.4 default auth (caching_sha2_password).
 * Does NOT enable mysql_native_password.
 */
import { spawnSync } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { setTimeout as delay } from 'node:timers/promises';
import test from 'node:test';
import assert from 'node:assert/strict';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const formaAtlasDir = join(repoRoot, 'docker', 'atlas', 'forma');
const cozeSchemaPath = join(repoRoot, 'docker', 'volumes', 'mysql', 'schema.sql');
const migrationSql = readFileSync(
  join(formaAtlasDir, 'migrations', '20250831100000_initial.sql'),
  'utf8',
);

/** Coze docker-compose.yml baseline */
const MYSQL_IMAGE = process.env.FORMA_TEST_MYSQL_IMAGE ?? 'mysql:8.4.5';
const ATLAS_IMAGE = process.env.FORMA_TEST_ATLAS_IMAGE ?? 'arigaio/atlas:0.32.1';

const RUN_ID = `${Date.now()}-${process.pid}-${randomBytes(3).toString('hex')}`;
const NETWORK = `forma-mig-net-${RUN_ID}`;
const MYSQL_CONTAINER = `forma-mig-mysql-${RUN_ID}`;
const DATABASE = `opencoze_${RUN_ID.replace(/-/g, '_').slice(0, 40)}`;
const APP_USER = 'coze';
const APP_PASSWORD = 'coze123';
const ROOT_PASSWORD = 'root';
const MYSQL_ALIAS = 'mysql';

const READY_TIMEOUT_MS = 120_000;
const READY_POLL_MS = 2_000;

function dockerPath(hostPath) {
  const normalized = hostPath.replace(/\\/g, '/');
  if (process.platform === 'win32' && /^[A-Za-z]:/.test(normalized)) {
    return `/${normalized[0].toLowerCase()}${normalized.slice(2)}`;
  }
  return normalized;
}

function run(cmd, args, options = {}) {
  const result = spawnSync(cmd, args, {
    encoding: 'utf8',
    maxBuffer: 20 * 1024 * 1024,
    ...options,
  });
  return result;
}

function mustOk(result, label) {
  if (result.status !== 0) {
    const detail = [result.stderr, result.stdout].filter(Boolean).join('\n');
    throw new Error(`${label} failed (exit ${result.status}):\n${detail}`);
  }
  return result;
}

function docker(...args) {
  return run('docker', args);
}

function mysqlLogs() {
  const logs = docker('logs', '--tail', '80', MYSQL_CONTAINER);
  return logs.stdout || logs.stderr || '(no logs)';
}

async function waitForMysql() {
  const deadline = Date.now() + READY_TIMEOUT_MS;
  let lastErr = '';
  while (Date.now() < deadline) {
    const ping = docker(
      'exec',
      MYSQL_CONTAINER,
      'mysqladmin',
      'ping',
      '-h',
      '127.0.0.1',
      '-uroot',
      `-p${ROOT_PASSWORD}`,
      '--silent',
    );
    if (ping.status === 0) {
      const sel = docker(
        'exec',
        MYSQL_CONTAINER,
        'mysql',
        '-uroot',
        `-p${ROOT_PASSWORD}`,
        '-e',
        'SELECT 1',
      );
      if (sel.status === 0) {
        return;
      }
      lastErr = sel.stderr || sel.stdout;
    } else {
      lastErr = ping.stderr || ping.stdout;
    }
    await delay(READY_POLL_MS);
  }
  throw new Error(
    `MySQL not ready after ${READY_TIMEOUT_MS}ms.\nLast error: ${lastErr}\nLogs:\n${mysqlLogs()}`,
  );
}

function rootExec(sql, database) {
  const args = ['exec', MYSQL_CONTAINER, 'mysql', '-uroot', `-p${ROOT_PASSWORD}`, '-N', '-s'];
  if (database) {
    args.push(database);
  }
  args.push('-e', sql);
  const result = docker(...args);
  if (result.status !== 0) {
    throw new Error(
      `mysql exec failed:\n${result.stderr || result.stdout}\nSQL: ${sql}\nLogs:\n${mysqlLogs()}`,
    );
  }
  return (result.stdout || '').trim();
}

function applySqlFile(sqlPath, database) {
  const sql = readFileSync(sqlPath, 'utf8');
  const result = run(
    'docker',
    ['exec', '-i', MYSQL_CONTAINER, 'mysql', '-uroot', `-p${ROOT_PASSWORD}`, database],
    { input: sql },
  );
  mustOk(result, `apply SQL ${sqlPath}`);
}

function atlasUrl() {
  return `mysql://${APP_USER}:${APP_PASSWORD}@${MYSQL_ALIAS}:3306/${DATABASE}?charset=utf8mb4&parseTime=True`;
}

function atlasMigrateApply() {
  const url = atlasUrl();
  const mount = `${dockerPath(formaAtlasDir)}:/forma-atlas`;
  const result = docker(
    'run',
    '--rm',
    '--network',
    NETWORK,
    '-v',
    mount,
    ATLAS_IMAGE,
    'migrate',
    'apply',
    '--allow-dirty',
    '--dir',
    'file:///forma-atlas/migrations',
    '--url',
    url,
  );
  return result;
}

function atlasMigrateStatus() {
  const url = atlasUrl();
  const mount = `${dockerPath(formaAtlasDir)}:/forma-atlas`;
  return docker(
    'run',
    '--rm',
    '--network',
    NETWORK,
    '-v',
    mount,
    ATLAS_IMAGE,
    'migrate',
    'status',
    '--dir',
    'file:///forma-atlas/migrations',
    '--url',
    url,
  );
}

function tableExists(name) {
  const out = rootExec(
    `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${DATABASE}' AND table_name='${name}'`,
  );
  return out === '1';
}

function columnExists(table, column) {
  const out = rootExec(
    `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='${DATABASE}' AND table_name='${table}' AND column_name='${column}'`,
  );
  return out === '1';
}

function indexExists(table, indexName) {
  const out = rootExec(
    `SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema='${DATABASE}' AND table_name='${table}' AND index_name='${indexName}'`,
  );
  return Number(out) > 0;
}

function resetDatabase() {
  rootExec(`DROP DATABASE IF EXISTS \`${DATABASE}\``);
  rootExec(
    `CREATE DATABASE \`${DATABASE}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`,
  );
}

function applyCozeSchema() {
  // schema.sql begins with CREATE DATABASE IF NOT EXISTS opencoze — rewrite to our DB.
  const raw = readFileSync(cozeSchemaPath, 'utf8');
  const rewritten = raw
    .replace(/CREATE DATABASE IF NOT EXISTS opencoze[^;]*;/i, `USE \`${DATABASE}\`;`)
    .replace(/\bUSE\s+opencoze\b/gi, `USE \`${DATABASE}\``);
  const result = run(
    'docker',
    ['exec', '-i', MYSQL_CONTAINER, 'mysql', '-uroot', `-p${ROOT_PASSWORD}`, DATABASE],
    { input: rewritten },
  );
  mustOk(result, 'apply Coze schema');
}

function provisionAppUser() {
  // Default auth plugin on MySQL 8.4 = caching_sha2_password (no WITH clause).
  rootExec(`CREATE USER IF NOT EXISTS '${APP_USER}'@'%' IDENTIFIED BY '${APP_PASSWORD}'`);
  rootExec(`GRANT ALL PRIVILEGES ON \`${DATABASE}\`.* TO '${APP_USER}'@'%'`);
  rootExec('FLUSH PRIVILEGES');
  const plugin = rootExec(
    `SELECT plugin FROM mysql.user WHERE user='${APP_USER}' AND host='%' LIMIT 1`,
  );
  assert.notEqual(
    plugin,
    'mysql_native_password',
    'app user must not use mysql_native_password',
  );
  assert.ok(plugin.length > 0, 'app user plugin must be set');
}

async function setupStack() {
  mustOk(docker('network', 'create', NETWORK), 'docker network create');

  const start = docker(
    'run',
    '-d',
    '--name',
    MYSQL_CONTAINER,
    '--network',
    NETWORK,
    '--network-alias',
    MYSQL_ALIAS,
    '-e',
    `MYSQL_ROOT_PASSWORD=${ROOT_PASSWORD}`,
    MYSQL_IMAGE,
    '--character-set-server=utf8mb4',
    '--collation-server=utf8mb4_unicode_ci',
  );
  mustOk(start, 'docker run mysql');

  await waitForMysql();
  resetDatabase();
  provisionAppUser();
}

function teardownStack() {
  docker('rm', '-f', MYSQL_CONTAINER);
  docker('network', 'rm', NETWORK);
}

test('Forma migration integration (CASE A/B/C)', async (t) => {
  t.before(async () => {
    await setupStack();
  });

  t.after(() => {
    teardownStack();
  });

  await t.test('CASE A — fresh install: Coze schema then Forma migration', () => {
    resetDatabase();
    provisionAppUser();
    applyCozeSchema();

    assert.ok(tableExists('workflow_version'), 'Coze workflow_version must exist');
    assert.equal(tableExists('forma_asset_ref'), false);

    const apply = atlasMigrateApply();
    assert.equal(apply.status, 0, apply.stderr || apply.stdout);

    assert.ok(tableExists('forma_asset_ref'));
    assert.ok(tableExists('forma_coze_resource_ref'));
    assert.ok(tableExists('forma_principal'));
    assert.ok(tableExists('forma_tenant'));
    assert.ok(tableExists('forma_tenant_membership'));
    assert.ok(tableExists('forma_tenant_space_ref'));
    assert.ok(tableExists('forma_audit_event'));
    assert.ok(columnExists('forma_asset_ref', 'tenant_id'));
    assert.ok(columnExists('forma_asset_ref', 'content_digest'));
    assert.ok(columnExists('forma_asset_ref', 'deleted_at'));
    assert.ok(columnExists('forma_tenant', 'revision'));
    assert.ok(columnExists('forma_coze_resource_ref', 'coze_resource_type'));
    assert.ok(columnExists('forma_coze_resource_ref', 'coze_resource_id'));
    assert.ok(indexExists('forma_asset_ref', 'uk_forma_asset_tenant_asset_revision'));
    assert.ok(indexExists('forma_coze_resource_ref', 'uk_forma_coze_resource_binding'));
    assert.ok(indexExists('forma_principal', 'uk_forma_principal_coze_user'));
    assert.ok(indexExists('forma_tenant_space_ref', 'uk_forma_space_id'), 'S1-G1 unique coze_space_id');
    assert.equal(
      indexExists('forma_tenant_space_ref', 'uk_forma_space_active'),
      false,
      'legacy (coze_space_id, status) unique index must be dropped',
    );
    assert.ok(migrationSql.includes('CREATE TABLE IF NOT EXISTS `forma_asset_ref`'));
  });

  await t.test('CASE B — upgrade: Coze-only DB then Forma migration', () => {
    resetDatabase();
    provisionAppUser();
    applyCozeSchema();

    assert.ok(tableExists('workflow_version'), 'Coze schema present');
    assert.ok(tableExists('agent_to_database'), 'Coze agent_to_database present');
    assert.equal(tableExists('forma_asset_ref'), false, 'Forma tables must not exist yet');
    assert.equal(tableExists('forma_coze_resource_ref'), false);

    const apply = atlasMigrateApply();
    assert.equal(apply.status, 0, apply.stderr || apply.stdout);

    assert.ok(tableExists('workflow_version'), 'Coze schema must remain');
    assert.ok(tableExists('agent_to_database'), 'Coze schema must remain');
    assert.ok(tableExists('forma_asset_ref'), 'Forma asset table created');
    assert.ok(tableExists('forma_coze_resource_ref'), 'Forma mapping table created');
  });

  await t.test('CASE C — idempotency: re-run Atlas apply/status on same network', () => {
    assert.ok(tableExists('forma_asset_ref'), 'CASE C requires CASE B migrated DB');

    const first = atlasMigrateApply();
    assert.equal(first.status, 0, first.stderr || first.stdout);

    const second = atlasMigrateApply();
    assert.equal(second.status, 0, second.stderr || second.stdout);

    const status = atlasMigrateStatus();
    assert.equal(status.status, 0, status.stderr || status.stdout);
    const statusOut = `${status.stdout}\n${status.stderr}`;
    assert.match(statusOut, /Migration Status|OK|current/i);

    assert.ok(tableExists('forma_asset_ref'));
    assert.ok(tableExists('forma_coze_resource_ref'));
    assert.ok(tableExists('workflow_version'));
  });
});
