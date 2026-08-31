import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const root = join(repoRoot, 'docker', 'atlas', 'forma');

test('initial migration defines forma tables', () => {
  const sql = readFileSync(join(root, 'migrations', '20250831100000_initial.sql'), 'utf8');
  assert.ok(sql.includes('forma_asset_ref'));
  assert.ok(sql.includes('forma_coze_resource_ref'));
  assert.ok(sql.includes('CREATE TABLE IF NOT EXISTS'));
});

test('S1 tenancy migration defines identity tables', () => {
  const sql = readFileSync(join(root, 'migrations', '20250831120000_s1_tenancy.sql'), 'utf8');
  assert.ok(sql.includes('forma_principal'));
  assert.ok(sql.includes('forma_tenant'));
  assert.ok(sql.includes('forma_tenant_membership'));
  assert.ok(sql.includes('forma_tenant_space_ref'));
  assert.ok(sql.includes('forma_audit_event'));
  assert.ok(!sql.includes('FOREIGN KEY'));
});

test('atlas sum references migrations', () => {
  const sum = readFileSync(join(root, 'migrations', 'atlas.sum'), 'utf8');
  assert.ok(sum.includes('20250831100000_initial.sql'));
  assert.ok(sum.includes('20250831120000_s1_tenancy.sql'));
});
