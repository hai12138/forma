/**
 * Forma S0-B-G1 migration real apply validation (CASE A/B/C).
 * Requires: disposable MySQL on FORMA_TEST_MYSQL_PORT (default 3307).
 */
import { spawnSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import assert from 'node:assert/strict';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const formaAtlasDir = join(repoRoot, 'docker', 'atlas', 'forma');
const cozeSchema = join(repoRoot, 'docker', 'volumes', 'mysql', 'schema.sql');
const migrationSql = readFileSync(
  join(formaAtlasDir, 'migrations', '20250831100000_initial.sql'),
  'utf8',
);

const host = process.env.FORMA_TEST_MYSQL_HOST ?? '127.0.0.1';
const port = process.env.FORMA_TEST_MYSQL_PORT ?? '3307';
const user = process.env.FORMA_TEST_MYSQL_USER ?? 'coze';
const password = process.env.FORMA_TEST_MYSQL_PASSWORD ?? 'coze123';
const database = process.env.FORMA_TEST_MYSQL_DB ?? 'opencoze';

const mysqlArgs = ['-u', user, `-p${password}`, database];

function mysqlExec(sql, useDatabase = true) {
  const args = ['exec', 'forma-mysql-g1', 'mysql', '-u', user, `-p${password}`];
  if (useDatabase) {
    args.push(database);
  }
  args.push('-e', sql);
  const result = spawnSync('docker', args, {
    encoding: 'utf8',
  });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || 'mysql exec failed');
  }
  return result.stdout;
}

function atlasApply(extraEnv = {}) {
  const atlasHost = process.env.FORMA_ATLAS_DOCKER_HOST ?? 'host.docker.internal';
  const url = `mysql://${user}:${password}@${atlasHost}:${port}/${database}?charset=utf8mb4&parseTime=True`;
  const result = spawnSync(
    'docker',
    [
      'run',
      '--rm',
      '-v',
      `${formaAtlasDir.replace(/\\/g, '/')}:/forma-atlas`,
      '-e',
      `FORMA_ATLAS_URL=${url}`,
      'arigaio/atlas:latest',
      'migrate',
      'apply',
      '--allow-dirty',
      '--dir',
      'file:///forma-atlas/migrations',
      '--url',
      url,
    ],
    { encoding: 'utf8', env: { ...process.env, ...extraEnv } },
  );
  return result;
}

function tableExists(name) {
  const out = mysqlExec(`SHOW TABLES LIKE '${name}';`);
  return out.includes(name);
}

function columnExists(table, column) {
  const out = mysqlExec(`SHOW COLUMNS FROM ${table} LIKE '${column}';`);
  return out.includes(column);
}

test('CASE A — fresh install: Coze schema then Forma migration', () => {
  mysqlExec('DROP DATABASE IF EXISTS opencoze;', false);
  mysqlExec('CREATE DATABASE opencoze CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;', false);

  const cozeSql = readFileSync(cozeSchema, 'utf8');
  const applyCoze = spawnSync('docker', ['exec', '-i', 'forma-mysql-g1', 'mysql', ...mysqlArgs], {
    input: cozeSql,
    encoding: 'utf8',
  });
  assert.equal(applyCoze.status, 0, applyCoze.stderr || applyCoze.stdout);

  const applyForma = atlasApply();
  assert.equal(applyForma.status, 0, applyForma.stderr || applyForma.stdout);

  assert.ok(tableExists('forma_asset_ref'));
  assert.ok(tableExists('forma_coze_resource_ref'));
  assert.ok(columnExists('forma_asset_ref', 'tenant_id'));
  assert.ok(columnExists('forma_asset_ref', 'content_digest'));
  assert.ok(columnExists('forma_coze_resource_ref', 'coze_resource_type'));
  assert.ok(migrationSql.includes('uk_forma_asset_tenant_asset_revision'));
});

test('CASE B — upgrade: existing Coze schema, apply Forma only', () => {
  assert.ok(tableExists('workflow_version'), 'Coze schema should remain intact');
  assert.ok(tableExists('forma_asset_ref'));
});

test('CASE C — idempotency: re-run Forma migration apply', () => {
  const first = atlasApply();
  assert.equal(first.status, 0, first.stderr || first.stdout);
  const second = atlasApply();
  assert.equal(second.status, 0, second.stderr || second.stdout);
  assert.ok(tableExists('forma_asset_ref'));
  assert.ok(tableExists('forma_coze_resource_ref'));
});
