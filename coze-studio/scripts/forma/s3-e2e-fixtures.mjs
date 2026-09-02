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
  const cmd = `docker exec -i ${container} mysql -u${mysqlUser} -p${mysqlPass} ${mysqlDb}`;
  return execSync(cmd, { input: sql, encoding: 'utf8' });
}

export function queryMysql(sql) {
  const container = mysqlContainer();
  if (!container) {
    throw new Error('coze-mysql container not running');
  }
  const cmd = `docker exec -i ${container} mysql -u${mysqlUser} -p${mysqlPass} ${mysqlDb} -N`;
  return execSync(cmd, { input: sql, encoding: 'utf8' }).trim();
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
  const suffix = businessId.replace(/^biz_/, '').replace(/-/g, '').slice(0, 24);
  const conflictId = `conf_e2e_${suffix}`;
  const assertA = `assert_e2e_a_${suffix}`;
  const assertB = `assert_e2e_b_${suffix}`;
  const evA = `evid_e2e_a_${suffix}`;
  const evB = `evid_e2e_b_${suffix}`;
  const existing = queryMysql(
    `SELECT status FROM forma_assertion_conflict WHERE tenant_id='${tenantId}' AND business_id='${businessId}' AND conflict_id='${conflictId}' LIMIT 1`,
  );
  if (existing === 'OPEN') return conflictId;

  execMysql(`
    DELETE FROM forma_assertion_conflict WHERE tenant_id='${tenantId}' AND conflict_id='${conflictId}';
    DELETE FROM forma_assertion_evidence_ref WHERE tenant_id='${tenantId}' AND assertion_id IN ('${assertA}', '${assertB}');
    DELETE FROM forma_business_assertion WHERE tenant_id='${tenantId}' AND assertion_id IN ('${assertA}', '${assertB}');
    DELETE FROM forma_business_evidence WHERE tenant_id='${tenantId}' AND evidence_id IN ('${evA}', '${evB}');
  `);

  const ts = nowSql();
  const turnA = userTurnId || `turn_e2e_a_${randomUUID().slice(0, 8)}`;
  const turnB = `turn_e2e_b_${randomUUID().slice(0, 8)}`;
  const subject = 'actor:管理员';
  const predicate = 'has_permission';

  execMysql(`
    INSERT INTO forma_business_evidence
      (evidence_id, tenant_id, business_id, session_id, turn_id, source_type, source_ref, quote, content_digest, created_by, created_at)
    VALUES
      ('${evA}', '${tenantId}', '${businessId}', '${sessionId}', '${turnA}', 'INTERVIEW_TURN', '${turnA}', 'E2E conflict fixture A', 'dig_a', '${principalId}', '${ts}'),
      ('${evB}', '${tenantId}', '${businessId}', '${sessionId}', '${turnB}', 'INTERVIEW_TURN', '${turnB}', 'E2E conflict fixture B', 'dig_b', '${principalId}', '${ts}')
  `);
  execMysql(`
    INSERT INTO forma_business_assertion
      (assertion_id, tenant_id, business_id, session_id, assertion_type, subject_ref, predicate, object_value, structured_value_json, confidence, status, source_marker, derived_from_assertion_id, created_by, created_at, updated_at)
    VALUES
      ('${assertA}', '${tenantId}', '${businessId}', '${sessionId}', 'PROPERTY', '${subject}', '${predicate}', 'E2E_CONFLICT_A_关闭工单', '{}', 0.9, 'PROPOSED', 'AI_EXTRACTED', '', '${principalId}', '${ts}', '${ts}'),
      ('${assertB}', '${tenantId}', '${businessId}', '${sessionId}', 'PROPERTY', '${subject}', '${predicate}', 'E2E_CONFLICT_B_维修班长关闭', '{}', 0.9, 'PROPOSED', 'AI_EXTRACTED', '', '${principalId}', '${ts}', '${ts}')
  `);
  execMysql(`
    INSERT INTO forma_assertion_evidence_ref (tenant_id, assertion_id, evidence_id, created_at)
    VALUES
      ('${tenantId}', '${assertA}', '${evA}', '${ts}'),
      ('${tenantId}', '${assertB}', '${evB}', '${ts}')
  `);
  execMysql(`
    INSERT INTO forma_assertion_conflict
      (conflict_id, tenant_id, business_id, session_id, assertion_id_a, assertion_id_b, subject_ref, predicate, status, created_at)
    VALUES
      ('${conflictId}', '${tenantId}', '${businessId}', '${sessionId}', '${assertA}', '${assertB}', '${subject}', '${predicate}', 'OPEN', '${ts}')
  `);

  const seeded = queryMysql(
    `SELECT status FROM forma_assertion_conflict WHERE tenant_id='${tenantId}' AND conflict_id='${conflictId}' LIMIT 1`,
  );
  if (seeded !== 'OPEN') {
    throw new Error(`conflict fixture insert failed for ${conflictId}: ${seeded || 'missing'}`);
  }
  return conflictId;
}

export function ensureProposedAssertionForEdit({ tenantId, businessId, sessionId, principalId }) {
  const suffix = businessId.replace(/^biz_/, '').replace(/-/g, '').slice(0, 24);
  const assertId = `assert_e2e_edit_${suffix}`;
  const evId = `evid_e2e_edit_${suffix}`;
  const existing = queryMysql(
    `SELECT status FROM forma_business_assertion WHERE tenant_id='${tenantId}' AND business_id='${businessId}' AND assertion_id='${assertId}' LIMIT 1`,
  );
  if (existing === 'PROPOSED') return assertId;

  execMysql(`
    DELETE FROM forma_assertion_evidence_ref WHERE tenant_id='${tenantId}' AND assertion_id='${assertId}';
    DELETE FROM forma_business_assertion WHERE tenant_id='${tenantId}' AND assertion_id='${assertId}';
    DELETE FROM forma_business_evidence WHERE tenant_id='${tenantId}' AND evidence_id='${evId}';
  `);

  const ts = nowSql();
  const turnId = `turn_e2e_edit_${randomUUID().slice(0, 8)}`;

  execMysql(`
    INSERT INTO forma_business_evidence
      (evidence_id, tenant_id, business_id, session_id, turn_id, source_type, source_ref, quote, content_digest, created_by, created_at)
    VALUES
      ('${evId}', '${tenantId}', '${businessId}', '${sessionId}', '${turnId}', 'INTERVIEW_TURN', '${turnId}', 'E2E edit fixture evidence', 'dig_edit', '${principalId}', '${ts}')
  `);
  execMysql(`
    INSERT INTO forma_business_assertion
      (assertion_id, tenant_id, business_id, session_id, assertion_type, subject_ref, predicate, object_value, structured_value_json, confidence, status, source_marker, derived_from_assertion_id, created_by, created_at, updated_at)
    VALUES
      ('${assertId}', '${tenantId}', '${businessId}', '${sessionId}', 'ACTOR_EXISTS', 'actor:E2E_EDIT_TARGET', 'exists', 'E2E_EDIT_TARGET_维修人员', '{}', 0.85, 'PROPOSED', 'AI_EXTRACTED', '', '${principalId}', '${ts}', '${ts}')
  `);
  execMysql(`
    INSERT INTO forma_assertion_evidence_ref (tenant_id, assertion_id, evidence_id, created_at)
    VALUES ('${tenantId}', '${assertId}', '${evId}', '${ts}')
  `);

  const seeded = queryMysql(
    `SELECT status FROM forma_business_assertion WHERE tenant_id='${tenantId}' AND business_id='${businessId}' AND assertion_id='${assertId}' LIMIT 1`,
  );
  if (seeded !== 'PROPOSED') {
    throw new Error(`edit fixture insert failed for ${assertId}: ${seeded || 'missing'}`);
  }
  return assertId;
}

export function countModelCalls(sessionId) {
  if (!sessionId) return 0;
  const n = queryMysql(
    `SELECT COUNT(*) FROM forma_analyst_model_call WHERE session_id='${sessionId}'`,
  );
  return parseInt(n, 10) || 0;
}

export function verifyModelCallsForSession(sessionId, liveLogPath, appendLog) {
  const rows = queryMysql(
    `SELECT operation, model_ref, success, latency_ms FROM forma_analyst_model_call WHERE session_id='${sessionId}' ORDER BY id`,
  );
  appendLog(liveLogPath, `session_id=${sessionId}`);
  appendLog(liveLogPath, `model_call_count=${countModelCalls(sessionId)}`);
  for (const line of rows.split('\n').filter(Boolean)) {
    appendLog(liveLogPath, line.replace(/\t/g, ' '));
  }
  return rows;
}

/** USER turn + evidence without triggering model analysis (fixture phase). */
export function seedUserTurnAndEvidence({ tenantId, businessId, sessionId, principalId, content }) {
  const turnId = `turn_e2e_user_${randomUUID().slice(0, 8)}`;
  const evId = `evid_e2e_user_${randomUUID().slice(0, 8)}`;
  const ts = nowSql();
  const maxSeq = parseInt(
    queryMysql(`SELECT COALESCE(MAX(\`sequence\`), 0) FROM forma_analyst_turn WHERE session_id='${sessionId}'`),
    10,
  );
  const nextSeq = maxSeq + 1;
  const clientReq = `e2e_fixture_${randomUUID().slice(0, 12)}`;

  execMysql(`
    INSERT INTO forma_analyst_turn
      (turn_id, tenant_id, session_id, \`sequence\`, speaker, content, content_type, client_request_id, analysis_status, created_at)
    VALUES
      ('${turnId}', '${tenantId}', '${sessionId}', ${nextSeq}, 'USER', '${content.replace(/'/g, "''")}', 'TEXT', '${clientReq}', 'COMPLETED', '${ts}')
  `);
  execMysql(`
    UPDATE forma_analyst_session
    SET next_turn_sequence=${nextSeq + 1}, updated_at='${ts}'
    WHERE session_id='${sessionId}' AND tenant_id='${tenantId}'
  `);
  execMysql(`
    INSERT INTO forma_business_evidence
      (evidence_id, tenant_id, business_id, session_id, turn_id, source_type, source_ref, quote, content_digest, created_by, created_at)
    VALUES
      ('${evId}', '${tenantId}', '${businessId}', '${sessionId}', '${turnId}', 'INTERVIEW_TURN', '${turnId}', '${content.replace(/'/g, "''")}', 'dig_fixture', '${principalId}', '${ts}')
  `);
  return { turnId, evidenceId: evId };
}

/** Confirmed assertion for stale-proposal gate — no model. */
export function seedConfirmedAssertionForStale({ tenantId, businessId, sessionId, principalId }) {
  const suffix = businessId.replace(/^biz_/, '').replace(/-/g, '').slice(0, 20);
  const assertId = `assert_e2e_stale_${suffix}`;
  const evId = `evid_e2e_stale_${suffix}`;
  const existing = queryMysql(
    `SELECT status FROM forma_business_assertion WHERE tenant_id='${tenantId}' AND assertion_id='${assertId}' LIMIT 1`,
  );
  if (existing === 'CONFIRMED') return assertId;

  execMysql(`
    DELETE FROM forma_assertion_evidence_ref WHERE tenant_id='${tenantId}' AND assertion_id='${assertId}';
    DELETE FROM forma_business_assertion WHERE tenant_id='${tenantId}' AND assertion_id='${assertId}';
    DELETE FROM forma_business_evidence WHERE tenant_id='${tenantId}' AND evidence_id='${evId}';
  `);

  const ts = nowSql();
  const turnId = `turn_e2e_stale_${randomUUID().slice(0, 8)}`;
  execMysql(`
    INSERT INTO forma_business_evidence
      (evidence_id, tenant_id, business_id, session_id, turn_id, source_type, source_ref, quote, content_digest, created_by, created_at)
    VALUES
      ('${evId}', '${tenantId}', '${businessId}', '${sessionId}', '${turnId}', 'INTERVIEW_TURN', '${turnId}', 'E2E stale fixture', 'dig_stale', '${principalId}', '${ts}')
  `);
  execMysql(`
    INSERT INTO forma_business_assertion
      (assertion_id, tenant_id, business_id, session_id, assertion_type, subject_ref, predicate, object_value, structured_value_json, confidence, status, source_marker, derived_from_assertion_id, created_by, created_at, updated_at)
    VALUES
      ('${assertId}', '${tenantId}', '${businessId}', '${sessionId}', 'ACTOR_EXISTS', 'actor:E2E_STALE_GATE', 'exists', 'E2E_STALE_GATE_审批员', '{}', 0.9, 'CONFIRMED', 'AI_EXTRACTED', '', '${principalId}', '${ts}', '${ts}')
  `);
  execMysql(`
    INSERT INTO forma_assertion_evidence_ref (tenant_id, assertion_id, evidence_id, created_at)
    VALUES ('${tenantId}', '${assertId}', '${evId}', '${ts}')
  `);
  return assertId;
}
