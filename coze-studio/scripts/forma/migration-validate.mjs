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
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('S1-G1 space mapping migration present', () => {
  const sql = readFileSync(join(root, 'migrations', '20250831140000_s1_g1_space_mapping.sql'), 'utf8');
  assert.ok(sql.includes('forma_tenant_space_ref'));
  assert.ok(sql.includes('uk_forma_space_id') || sql.includes('coze_space_id'));
});

test('S2 business model migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250901000000_s2_business_model.sql'),
    'utf8',
  );
  assert.ok(sql.includes('forma_business_model'));
  assert.ok(sql.includes('forma_business_model_revision'));
  assert.ok(sql.includes('forma_business_model_layout'));
  assert.ok(sql.includes('semantic_model_json'));
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('S3 analyst migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902000000_s3_analyst.sql'),
    'utf8',
  );
  assert.ok(sql.includes('forma_analyst_session'));
  assert.ok(sql.includes('forma_business_evidence'));
  assert.ok(sql.includes('forma_business_assertion'));
  assert.ok(sql.includes('forma_business_model_proposal'));
  assert.ok(sql.includes('forma_revision_provenance'));
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('S3-G1 integrity migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902010000_s3_g1_integrity.sql'),
    'utf8',
  );
  assert.ok(sql.includes('next_turn_sequence'));
  assert.ok(sql.includes('reply_to_turn_id'));
  assert.ok(sql.includes('uk_forma_analyst_turn_sequence'));
  assert.ok(sql.includes('uk_forma_assertion_conflict_pair'));
  assert.ok(sql.includes('ALTER TABLE'));
  assert.ok(!sql.includes('CREATE UNIQUE INDEX IF NOT EXISTS'));
});

test('S4-G1 data requirement migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902100000_s4_g1_data_requirement.sql'),
    'utf8',
  );
  assert.ok(sql.includes('forma_data_requirement'));
  assert.ok(sql.includes('forma_data_requirement_analysis_run'));
  assert.ok(sql.includes('forma_data_requirement_decision'));
  assert.ok(sql.includes('uk_forma_data_analysis_idempotency'));
  assert.ok(sql.includes('uk_forma_data_req_decision_source'));
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('S4-G1-F1 analysis lease audit migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902103000_s4_g1_analysis_lease_audit.sql'),
    'utf8',
  );
  assert.ok(sql.includes('execution_generation'));
  assert.ok(sql.includes('lease_expires_at'));
  assert.ok(sql.includes('last_retry_by'));
  assert.ok(sql.includes('last_retry_at'));
  assert.ok(sql.includes('ALTER TABLE'));
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('S4-G2 data source discovery migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902110000_s4_g2_data_source_discovery.sql'),
    'utf8',
  );
  for (const table of [
    'forma_data_source',
    'forma_data_connection',
    'forma_data_credential_ref',
    'forma_data_secret_local',
    'forma_data_asset',
    'forma_data_schema_snapshot',
  ]) {
    assert.ok(sql.includes(table));
  }
  assert.ok(sql.includes('uk_forma_data_asset_locator'));
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('S4-G2-F1 contract alignment migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902113000_s4_g2_contract_alignment.sql'),
    'utf8',
  );
  assert.ok(sql.includes("source_type` = 'RELATIONAL_DATABASE'"));
  assert.ok(sql.includes("source_type` = 'HTTP_API'"));
  assert.ok(sql.includes('last_test_status'));
  assert.ok(sql.includes('last_test_at'));
  assert.ok(sql.includes('last_test_error_key'));
});

test('S4-G3 semantic mapping migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902120000_s4_g3_semantic_mapping.sql'),
    'utf8',
  );
  assert.ok(sql.includes('forma_data_semantic_mapping'));
  assert.ok(sql.includes('forma_data_semantic_mapping_analysis_run'));
  assert.ok(sql.includes('forma_data_semantic_mapping_decision'));
  assert.ok(sql.includes('uk_forma_data_mapping_analysis_idempotency'));
  assert.ok(sql.includes('uk_forma_data_mapping_decision_source'));
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('S4-G4 data contract migration present', () => {
  const sql = readFileSync(
    join(root, 'migrations', '20250902130000_s4_g4_data_contract.sql'),
    'utf8',
  );
  for (const table of [
    'forma_data_contract',
    'forma_data_contract_revision',
    'forma_data_contract_validation_result',
    'forma_data_contract_lifecycle_event',
    'forma_data_contract_drift_result',
    'forma_data_contract_gap_result',
  ]) {
    assert.ok(sql.includes(table));
  }
  assert.ok(sql.includes('uk_forma_data_contract'));
  assert.ok(sql.includes('uk_forma_data_contract_revision_version'));
  assert.ok(sql.includes('uk_forma_data_contract_validation'));
  assert.ok(sql.includes('uk_forma_data_contract_lifecycle_event'));
  assert.ok(sql.includes('uk_forma_data_contract_drift'));
  assert.ok(sql.includes('uk_forma_data_contract_gap'));
  assert.ok(!/FOREIGN\s+KEY\s*\(/i.test(sql));
});

test('atlas sum references migrations', () => {
  const sum = readFileSync(join(root, 'migrations', 'atlas.sum'), 'utf8');
  assert.ok(sum.includes('20250831100000_initial.sql'));
  assert.ok(sum.includes('20250831120000_s1_tenancy.sql'));
  assert.ok(sum.includes('20250831140000_s1_g1_space_mapping.sql'));
  assert.ok(sum.includes('20250901000000_s2_business_model.sql'));
  assert.ok(sum.includes('20250902000000_s3_analyst.sql'));
  assert.ok(sum.includes('20250902010000_s3_g1_integrity.sql'));
  assert.ok(sum.includes('20250902100000_s4_g1_data_requirement.sql'));
  assert.ok(sum.includes('20250902103000_s4_g1_analysis_lease_audit.sql'));
  assert.ok(sum.includes('20250902110000_s4_g2_data_source_discovery.sql'));
  assert.ok(sum.includes('20250902113000_s4_g2_contract_alignment.sql'));
  assert.ok(sum.includes('20250902120000_s4_g3_semantic_mapping.sql'));
  assert.ok(sum.includes('20250902130000_s4_g4_data_contract.sql'));
});
