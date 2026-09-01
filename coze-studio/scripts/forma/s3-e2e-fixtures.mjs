/**
 * FORMA S3 Live E2E deterministic DB fixtures (test harness only).
 * Used by scripts/forma/s3-browser-e2e.mjs — not production code.
 */
import { execSync } from 'node:child_process';
import { randomUUID } from 'node:crypto';

const mysqlUser = process.env.FORMA_MYSQL_USER || 'coze';
const mysqlPass = process.env.FORMA_MYSQL_PASSWORD || 'coze123';
const mysqlDb = process.env.FORMA_MYSQL_DATABASE || 'opencoze';

function mysqlContainer() {
  const name = process.env.FORMA_MYSQL_CONTAINER || 'forma-live-mysql';
  return execSync(`docker ps -qf "name=${name}"`, { encoding: 'utf8' }).trim().split('\n')[0];
}

export function execMysql(sql) {
  const container = mysqlContainer();
  if (!container) {
    throw new Error('coze-mysql container not running');
  }
  const escaped = sql.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
  const cmd = `docker exec ${container} mysql -u${mysqlUser} -p${mysqlPass} ${mysqlDb} -e "${escaped}"`;
  return execSync(cmd, { encoding: 'utf8' });
}

export function queryMysql(sql) {
  const container = mysqlContainer();
  if (!container) {
    throw new Error('coze-mysql container not running');
  }
  const escaped = sql.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
  const cmd = `docker exec ${container} mysql -u${mysqlUser} -p${mysqlPass} ${mysqlDb} -N -e "${escaped}"`;
  return execSync(cmd, { encoding: 'utf8' }).trim();
}

function nowSql() {
  return new Date().toISOString().slice(0, 23).replace('T', ' ');
}

export function ensureOpenGap({ tenantId, businessId, sessionId }) {
  const existing = queryMysql(
    `SELECT gap_id FROM forma_analyst_gap WHERE tenant_id='${tenantId}' AND business_id='${businessId}' AND status='OPEN' LIMIT 1`,
  );
  if (existing) return existing.split('\n')[0];

  const gapId = `gap_e2e_${randomUUID().slice(0, 8)}`;
  const ts = nowSql();
  execMysql(`
    INSERT INTO forma_analyst_gap
      (gap_id, tenant_id, business_id, session_id, gap_type, question, related_assertion_ids_json, status, created_at, updated_at)
    VALUES
      ('${gapId}', '${tenantId}', '${businessId}', '${sessionId}', 'INFORMATION',
       '工单关闭后是否需要额外审批？', '[]', 'OPEN', '${ts}', '${ts}')
  `);
  return gapId;
}

export function seedDeterministicConflict({ tenantId, businessId, sessionId, principalId, userTurnId }) {
  const existing = queryMysql(
    `SELECT conflict_id FROM forma_assertion_conflict WHERE tenant_id='${tenantId}' AND business_id='${businessId}' AND status='OPEN' LIMIT 1`,
  );
  if (existing) return existing.split('\n')[0];

  const ts = nowSql();
  const assertA = `assert_e2e_a_${randomUUID().slice(0, 8)}`;
  const assertB = `assert_e2e_b_${randomUUID().slice(0, 8)}`;
  const evA = `evid_e2e_a_${randomUUID().slice(0, 8)}`;
  const evB = `evid_e2e_b_${randomUUID().slice(0, 8)}`;
  const conflictId = `conf_e2e_${randomUUID().slice(0, 8)}`;
  const turnA = userTurnId || `turn_e2e_a_${randomUUID().slice(0, 8)}`;
  const turnB = `turn_e2e_b_${randomUUID().slice(0, 8)}`;
  const subject = 'actor:管理员';
  const predicate = 'has_permission';

  execMysql(`
    INSERT INTO forma_business_evidence
      (evidence_id, tenant_id, business_id, session_id, turn_id, source_type, source_ref, quote, content_digest, created_by, created_at)
    VALUES
      ('${evA}', '${tenantId}', '${businessId}', '${sessionId}', '${turnA}', 'INTERVIEW_TURN', '${turnA}', '工单由管理员关闭', 'dig_a', '${principalId}', '${ts}'),
      ('${evB}', '${tenantId}', '${businessId}', '${sessionId}', '${turnB}', 'INTERVIEW_TURN', '${turnB}', '工单由维修班长关闭', 'dig_b', '${principalId}', '${ts}');
    INSERT INTO forma_business_assertion
      (assertion_id, tenant_id, business_id, session_id, assertion_type, subject_ref, predicate, object_value, structured_value_json, confidence, status, source_marker, derived_from_assertion_id, created_by, created_at, updated_at)
    VALUES
      ('${assertA}', '${tenantId}', '${businessId}', '${sessionId}', 'PROPERTY', '${subject}', '${predicate}', '关闭工单', '{}', 0.9, 'PROPOSED', 'AI_EXTRACTED', '', '${principalId}', '${ts}', '${ts}'),
      ('${assertB}', '${tenantId}', '${businessId}', '${sessionId}', 'PROPERTY', '${subject}', '${predicate}', '关闭工单(维修班长)', '{}', 0.9, 'PROPOSED', 'AI_EXTRACTED', '', '${principalId}', '${ts}', '${ts}');
    INSERT INTO forma_assertion_evidence_ref (tenant_id, assertion_id, evidence_id, created_at)
    VALUES
      ('${tenantId}', '${assertA}', '${evA}', '${ts}'),
      ('${tenantId}', '${assertB}', '${evB}', '${ts}');
    INSERT INTO forma_assertion_conflict
      (conflict_id, tenant_id, business_id, session_id, assertion_id_a, assertion_id_b, subject_ref, predicate, status, created_at)
    VALUES
      ('${conflictId}', '${tenantId}', '${businessId}', '${sessionId}', '${assertA}', '${assertB}', '${subject}', '${predicate}', 'OPEN', '${ts}')
  `);
  return conflictId;
}

export function ensureProposedAssertionForEdit({ tenantId, businessId, sessionId, principalId }) {
  const count = queryMysql(
    `SELECT COUNT(*) FROM forma_business_assertion WHERE tenant_id='${tenantId}' AND business_id='${businessId}' AND status='PROPOSED'`,
  );
  if (Number(count) >= 1) return;

  const ts = nowSql();
  const assertId = `assert_e2e_edit_${randomUUID().slice(0, 8)}`;
  const evId = `evid_e2e_edit_${randomUUID().slice(0, 8)}`;
  const turnId = `turn_e2e_edit_${randomUUID().slice(0, 8)}`;

  execMysql(`
    INSERT INTO forma_business_evidence
      (evidence_id, tenant_id, business_id, session_id, turn_id, source_type, source_ref, quote, content_digest, created_by, created_at)
    VALUES
      ('${evId}', '${tenantId}', '${businessId}', '${sessionId}', '${turnId}', 'INTERVIEW_TURN', '${turnId}', '维修人员处理完成后通知管理员', 'dig_edit', '${principalId}', '${ts}');
    INSERT INTO forma_business_assertion
      (assertion_id, tenant_id, business_id, session_id, assertion_type, subject_ref, predicate, object_value, structured_value_json, confidence, status, source_marker, derived_from_assertion_id, created_by, created_at, updated_at)
    VALUES
      ('${assertId}', '${tenantId}', '${businessId}', '${sessionId}', 'ACTOR_EXISTS', 'actor:维修人员', 'exists', '维修人员', '{}', 0.85, 'PROPOSED', 'AI_EXTRACTED', '', '${principalId}', '${ts}', '${ts}');
    INSERT INTO forma_assertion_evidence_ref (tenant_id, assertion_id, evidence_id, created_at)
    VALUES ('${tenantId}', '${assertId}', '${evId}', '${ts}')
  `);
}

export function verifyModelCallsForSession(sessionId, liveLogPath, appendLog) {
  const rows = queryMysql(
    `SELECT operation, model_ref, success, latency_ms FROM forma_analyst_model_call WHERE session_id='${sessionId}' ORDER BY id`,
  );
  appendLog(liveLogPath, `session_id=${sessionId}`);
  for (const line of rows.split('\n').filter(Boolean)) {
    appendLog(liveLogPath, line.replace(/\t/g, ' '));
  }
  return rows;
}
