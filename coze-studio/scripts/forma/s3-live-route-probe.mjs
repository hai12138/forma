/**
 * FORMA S3 live route probe — verifies current harness exposes S3 analyst APIs.
 *   node scripts/forma/s3-live-route-probe.mjs
 */
import assert from 'node:assert/strict';
import { appendFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const base = (process.env.FORMA_LIVE_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const logPath = join(root, 'forma', 'cursor-results', 's3-route-probe.log');

function log(line) {
  appendFileSync(logPath, `[${new Date().toISOString()}] ${line}\n`, 'utf8');
  console.log(line);
}

const cookies = new Map();
async function req(method, path, body, tenant) {
  const h = { Accept: 'application/json', Cookie: [...cookies].map(([k, v]) => `${k}=${v}`).join('; ') };
  if (body !== undefined) h['Content-Type'] = 'application/json';
  if (tenant) h['X-Forma-Tenant'] = tenant;
  const r = await fetch(`${base}${path}`, {
    method,
    headers: h,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  for (const c of r.headers.getSetCookie?.() || []) {
    const [pair] = c.split(';');
    const i = pair.indexOf('=');
    if (i > 0) cookies.set(pair.slice(0, i), pair.slice(i + 1));
  }
  const text = await r.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    json = { raw: text };
  }
  return { status: r.status, json };
}

writeFileSync(logPath, '', 'utf8');
log(`probe base=${base}`);

const email = `route-probe-${Date.now()}@example.com`;
await req('POST', '/api/passport/web/email/register/v2/', { email, password: 'FormaE2E!23456' });
const boot = await req('POST', '/api/forma/v1/bootstrap', {});
assert.equal(boot.status, 200);
const tenant = boot.json.data?.tenant?.tenant_id;
assert.ok(tenant);
const biz = await req('POST', '/api/forma/v1/businesses', { name: 'route-probe' }, tenant);
assert.equal(biz.status, 200);
const bid = biz.json.data?.business_id;
assert.ok(bid);

for (const [method, path, body] of [
  ['GET', `/api/forma/v1/businesses/${bid}/assertions`],
  ['GET', `/api/forma/v1/businesses/${bid}/evidence`],
  ['POST', `/api/forma/v1/businesses/${bid}/analyst/sessions`, { title: 'probe', confirmation_policy: 'DEVELOPMENT' }],
]) {
  const r = await req(method, path, body, tenant);
  log(`${method} ${path} -> ${r.status}`);
  assert.notEqual(r.status, 404, `${method} ${path} must not 404`);
}

log('S3 route probe PASS');
