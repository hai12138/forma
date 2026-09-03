/**
 * FORMA S4-G6 — shared live helpers (auth, API, model budget, secret scan).
 * Secrets must never be logged or written to evidence.
 */
import { createServer } from 'node:http';
import { appendFileSync, mkdirSync, writeFileSync, existsSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import { randomBytes } from 'node:crypto';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
export const resultsDir = join(root, 'forma', 'cursor-results');
export const evidenceDir = join(resultsDir, 's4-g6-ui');
export const statePath = join(resultsDir, 's4-g6-e2e-state.json');
export const logPath = join(resultsDir, 's4-g6-live-e2e.log');

export const MAX_REAL_MODEL_CALLS = parseInt(process.env.MAX_REAL_MODEL_CALLS || '2', 10);
export const baseApi = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
export const baseUi = (process.env.FORMA_UI_BASE_URL || 'http://127.0.0.1:3001').replace(/\/$/, '');
export const mysqlContainer = process.env.FORMA_MYSQL_CONTAINER || 'forma-live-mysql';
export const mysqlUser = process.env.FORMA_MYSQL_USER || 'coze';
export const mysqlPassword = process.env.FORMA_MYSQL_PASSWORD || 'coze123';
export const mysqlDatabase = process.env.FORMA_MYSQL_DATABASE || 'opencoze';

export const G6_SECRET = `FORMA_G6_SUPER_SECRET_${randomBytes(8).toString('hex')}`;

let modelCalls = 0;
const modelCallLog = [];

export function getModelCalls() {
  return { count: modelCalls, calls: [...modelCallLog] };
}

export function assertModelBudget(operation) {
  if (modelCalls >= MAX_REAL_MODEL_CALLS) {
    throw new Error(
      `HARD STOP: real model budget exhausted (${modelCalls}/${MAX_REAL_MODEL_CALLS}) before ${operation}`,
    );
  }
}

export function recordModelCall(operation) {
  assertModelBudget(operation);
  modelCalls += 1;
  modelCallLog.push({ call_index: modelCalls, operation, at: new Date().toISOString() });
  log(`REAL_MODEL_CALL call_index=${modelCalls} operation=${operation}`);
  if (modelCalls > MAX_REAL_MODEL_CALLS) {
    throw new Error(`HARD STOP: model calls exceeded budget (${modelCalls})`);
  }
}

export function log(line) {
  mkdirSync(resultsDir, { recursive: true });
  appendFileSync(logPath, `[${new Date().toISOString()}] ${line}\n`, 'utf8');
}

export function saveState(state) {
  writeFileSync(statePath, `${JSON.stringify(state, null, 2)}\n`, 'utf8');
}

export function loadState() {
  if (!existsSync(statePath)) return null;
  try {
    return JSON.parse(readFileSync(statePath, 'utf8'));
  } catch {
    return null;
  }
}

export function jar() {
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

export async function api(path, { method = 'GET', body, tenantId, cookies, headers = {} } = {}) {
  const h = {
    Accept: 'application/json',
    'X-Request-ID': `s4g6-${Date.now()}-${randomBytes(2).toString('hex')}`,
    ...headers,
  };
  if (cookies) h.Cookie = cookies.header();
  if (tenantId) h['X-Forma-Tenant'] = tenantId;
  if (body !== undefined) h['Content-Type'] = 'application/json';
  const res = await fetch(`${baseApi}${path}`, {
    method,
    headers: h,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (cookies) cookies.store(res);
  const text = await res.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    json = { raw: text };
  }
  return { status: res.status, json, text, headers: res.headers };
}

export async function registerLoginBootstrap(email, password) {
  const cookies = jar();
  let r = await api('/api/passport/web/email/register/v2/', {
    method: 'POST',
    body: { email, password },
    cookies,
  });
  if (!cookies.header().includes('session_key')) {
    r = await api('/api/passport/web/email/login/', {
      method: 'POST',
      body: { email, password },
      cookies,
    });
  }
  if (!cookies.header().includes('session_key')) {
    throw new Error(`auth failed: ${JSON.stringify(r.json)}`);
  }
  const boot = await api('/api/forma/v1/bootstrap', { method: 'POST', body: {}, cookies });
  if (boot.status >= 400) throw new Error(`bootstrap failed: ${JSON.stringify(boot.json)}`);
  const me = await api('/api/forma/v1/me', { cookies });
  const tenantId = me.json?.data?.tenants?.[0]?.tenant_id || boot.json?.data?.tenant?.tenant_id;
  if (!tenantId) throw new Error(`no tenant: ${JSON.stringify(me.json)}`);
  return {
    cookies,
    tenantId,
    principalId: me.json?.data?.principal?.principal_id,
    role: me.json?.data?.tenants?.[0]?.role,
  };
}

export async function loginExisting(email, password) {
  const cookies = jar();
  const r = await api('/api/passport/web/email/login/', {
    method: 'POST',
    body: { email, password },
    cookies,
  });
  if (!cookies.header().includes('session_key')) {
    throw new Error(`login failed: ${JSON.stringify(r.json)}`);
  }
  const me = await api('/api/forma/v1/me', { cookies });
  const tenants = me.json?.data?.tenants || [];
  const tenantId = tenants[0]?.tenant_id;
  return {
    cookies,
    tenantId,
    principalId: me.json?.data?.principal?.principal_id,
    role: tenants[0]?.role,
    tenants,
  };
}

export function mysqlExec(sql, database = mysqlDatabase) {
  const result = spawnSync(
    'docker',
    ['exec', mysqlContainer, 'mysql', '-u', mysqlUser, `-p${mysqlPassword}`, database, '-e', sql],
    { encoding: 'utf8', maxBuffer: 10 << 20 },
  );
  if (result.status !== 0) {
    throw new Error(`mysql failed: ${result.stderr || result.stdout}`);
  }
  return result.stdout;
}

export async function putModel(cookies, tenantId, businessId, semanticModel, expectedRevision = 0) {
  return api(`/api/forma/v1/businesses/${businessId}/model`, {
    method: 'PUT',
    cookies,
    tenantId,
    body: { semantic_model: semanticModel, expected_revision: expectedRevision },
  });
}

export function ensureLabSchema() {
  // Prefer root for CREATE DATABASE + GRANT (coze may lack CREATE privilege).
  const rootPass = process.env.FORMA_MYSQL_ROOT_PASSWORD || 'root';
  const root = spawnSync(
    'docker',
    [
      'exec',
      mysqlContainer,
      'mysql',
      '-u',
      'root',
      `-p${rootPass}`,
      '-e',
      `CREATE DATABASE IF NOT EXISTS forma_g6_lab CHARACTER SET utf8mb4; GRANT ALL ON forma_g6_lab.* TO 'coze'@'%'; FLUSH PRIVILEGES;`,
    ],
    { encoding: 'utf8', maxBuffer: 10 << 20 },
  );
  if (root.status !== 0) {
    // Fallback: try as coze if DB already exists
    log(`root grant warn: ${root.stderr || root.stdout}`);
  }
  mysqlExec(
    `
CREATE TABLE IF NOT EXISTS sample (
  sample_id VARCHAR(64) PRIMARY KEY,
  batch_id VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  collected_at DATETIME NULL COMMENT 'collection timestamp',
  temperature_c DECIMAL(10,2) NULL
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS assay_result (
  result_id VARCHAR(64) PRIMARY KEY,
  sample_id VARCHAR(64) NOT NULL,
  assay_code VARCHAR(64) NOT NULL,
  value_num DECIMAL(18,6) NULL,
  CONSTRAINT fk_assay_sample FOREIGN KEY (sample_id) REFERENCES sample(sample_id)
) ENGINE=InnoDB;
CREATE OR REPLACE VIEW v_sample_summary AS
SELECT s.sample_id, s.batch_id, s.status, COUNT(a.result_id) AS result_count
FROM sample s LEFT JOIN assay_result a ON a.sample_id = s.sample_id
GROUP BY s.sample_id, s.batch_id, s.status;
`,
    'forma_g6_lab',
  );
}

export const labBusinessModel = {
  schema_version: '2.0',
  nodes: [
    { id: 'obj_sample', type: 'BUSINESS_OBJECT', name: '实验室样本', source_marker: 'MANUAL_MODIFIED' },
    { id: 'obj_batch', type: 'BUSINESS_OBJECT', name: '样本批次', source_marker: 'MANUAL_MODIFIED' },
    { id: 'obj_assay', type: 'BUSINESS_OBJECT', name: '检测项目', source_marker: 'MANUAL_MODIFIED' },
    { id: 'obj_result', type: 'BUSINESS_OBJECT', name: '检测结果', source_marker: 'MANUAL_MODIFIED' },
    { id: 'actor_lab', type: 'ACTOR', name: '实验员', source_marker: 'MANUAL_MODIFIED' },
  ],
  edges: [
    {
      id: 'e_batch_sample',
      source: 'obj_batch',
      target: 'obj_sample',
      type: 'CREATES',
      label: '产生样本',
      source_marker: 'MANUAL_MODIFIED',
    },
    {
      id: 'e_sample_result',
      source: 'obj_sample',
      target: 'obj_result',
      type: 'CREATES',
      label: '产生结果',
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
  rules: [],
  states: [
    {
      id: 'st_collected',
      object_ref: 'obj_sample',
      name: '已采集',
      initial: true,
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
};

export const procurementBusinessModel = {
  schema_version: '2.0',
  nodes: [
    { id: 'obj_contract', type: 'BUSINESS_OBJECT', name: '采购合同', source_marker: 'MANUAL_MODIFIED' },
    { id: 'obj_vendor', type: 'BUSINESS_OBJECT', name: '供应商', source_marker: 'MANUAL_MODIFIED' },
    { id: 'obj_amount', type: 'BUSINESS_OBJECT', name: '合同金额', source_marker: 'MANUAL_MODIFIED' },
    { id: 'actor_buyer', type: 'ACTOR', name: '采购负责人', source_marker: 'MANUAL_MODIFIED' },
  ],
  edges: [
    {
      id: 'e_vendor_contract',
      source: 'obj_vendor',
      target: 'obj_contract',
      type: 'CREATES',
      label: '签署合同',
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
  rules: [],
  states: [
    {
      id: 'st_draft',
      object_ref: 'obj_contract',
      name: '草稿',
      initial: true,
      source_marker: 'MANUAL_MODIFIED',
    },
  ],
};

export function startHttpFixture(port = 18080) {
  const openapi = {
    openapi: '3.0.3',
    info: { title: 'G6 Fixture API', version: '1.0.0' },
    paths: {
      '/items': {
        get: {
          operationId: 'listItems',
          responses: {
            '200': {
              description: 'ok',
              content: {
                'application/json': {
                  schema: {
                    type: 'object',
                    properties: {
                      items: {
                        type: 'array',
                        items: { $ref: '#/components/schemas/Item' },
                      },
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
    components: {
      schemas: {
        Item: {
          type: 'object',
          properties: {
            id: { type: 'string' },
            nested: {
              type: 'object',
              properties: { code: { type: 'string' }, qty: { type: 'integer' } },
            },
          },
        },
      },
    },
  };

  const fixtureRoot = join(resultsDir, '_g6-http-fixture');
  mkdirSync(fixtureRoot, { recursive: true });
  writeFileSync(join(fixtureRoot, 'openapi.json'), JSON.stringify(openapi));
  writeFileSync(
    join(fixtureRoot, 'items'),
    JSON.stringify({ items: [{ id: '1', nested: { code: 'A', qty: 2 } }] }),
  );
  writeFileSync(
    join(fixtureRoot, 'index.html'),
    JSON.stringify({ ok: true, service: 'forma-g6-http-fixture' }),
  );

  // Prefer in-network docker fixture so harness can reach it without host firewall.
  const container = process.env.FORMA_HTTP_FIXTURE_CONTAINER || 'forma-g6-http-fx';
  spawnSync('docker', ['rm', '-f', container], { encoding: 'utf8' });
  const run = spawnSync(
    'docker',
    [
      'run',
      '-d',
      '--name',
      container,
      '--network',
      'forma-live-net',
      '-v',
      `${fixtureRoot.replace(/\\/g, '/')}:/usr/share/nginx/html:ro`,
      'nginx:1.27-alpine',
    ],
    { encoding: 'utf8' },
  );
  if (run.status === 0 && run.stdout.trim()) {
    const host = process.env.FORMA_HTTP_FIXTURE_HOST || container;
    return Promise.resolve({
      server: null,
      port: 80,
      container,
      baseUrl: `http://${host}`,
      openapiUrl: `http://${host}/openapi.json`,
      close: async () => {
        spawnSync('docker', ['rm', '-f', container], { encoding: 'utf8' });
      },
    });
  }

  // Fallback: host listener + docker gateway / host.docker.internal
  const server = createServer((req, res) => {
    if (req.url === '/openapi.json') {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(openapi));
      return;
    }
    if (req.url === '/items' || req.url === '/') {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ items: [{ id: '1', nested: { code: 'A', qty: 2 } }] }));
      return;
    }
    res.writeHead(404);
    res.end('not found');
  });
  return new Promise(resolve => {
    server.listen(port, '0.0.0.0', () => {
      const gateway =
        process.env.FORMA_HTTP_FIXTURE_HOST ||
        spawnSync(
          'docker',
          [
            'network',
            'inspect',
            'forma-live-net',
            '--format',
            '{{range .IPAM.Config}}{{.Gateway}}{{end}}',
          ],
          { encoding: 'utf8' },
        ).stdout.trim() ||
        'host.docker.internal';
      resolve({
        server,
        port,
        baseUrl: `http://${gateway}:${port}`,
        openapiUrl: `http://${gateway}:${port}/openapi.json`,
        close: () =>
          new Promise((r, j) => server.close(err => (err ? j(err) : r()))),
      });
    });
  });
}

/** Scan text/files for exact secret material. Keywords alone are OK. */
export function assertNoSecretMaterial(text, secrets = [G6_SECRET]) {
  const body = String(text ?? '');
  for (const s of secrets) {
    if (s && body.includes(s)) {
      throw new Error('SECRET_LEAK: exact secret material found in evidence/output');
    }
  }
  const banned = [
    /Authorization:\s*Bearer\s+[A-Za-z0-9\-._~+/]+=*/i,
    /password\s*=\s*[^\s"']{8,}/i,
    /api[_-]?key\s*=\s*[^\s"']{8,}/i,
  ];
  for (const re of banned) {
    if (re.test(body)) {
      throw new Error(`SECRET_LEAK: pattern ${re} matched in evidence/output`);
    }
  }
}

export function scanPathsForSecrets(paths, secrets = [G6_SECRET]) {
  for (const p of paths) {
    if (!existsSync(p)) continue;
    assertNoSecretMaterial(readFileSync(p, 'utf8'), secrets);
  }
}
