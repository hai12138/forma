/**
 * Deterministic smoke checks for Forma local launcher (no full Docker stack).
 */
import test from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { createServer } from 'node:net';
import { writeFileSync, mkdirSync, rmSync } from 'node:fs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const localDir = resolve(__dirname, '..');
const core = join(localDir, 'forma-local.mjs');
const cozeRoot = resolve(localDir, '..', '..', '..');

function runCore(args, opts = {}) {
  return spawnSync(process.execPath, [core, ...args], {
    encoding: 'utf8',
    cwd: opts.cwd || process.cwd(),
    env: { ...process.env, ...(opts.env || {}) },
  });
}

test('help exits 0 and lists commands', () => {
  const r = runCore(['help']);
  assert.equal(r.status, 0, r.stderr);
  assert.match(r.stdout, /doctor/);
  assert.match(r.stdout, /start/);
  assert.match(r.stdout, /status/);
});

test('invalid command exits 2', () => {
  const r = runCore(['not-a-real-command']);
  assert.equal(r.status, 2);
  assert.match(r.stderr + r.stdout, /Unknown command/);
});

test('repo-root resolution is independent of cwd', async () => {
  const mod = await import(pathToFileURL(core).href);
  assert.equal(mod.__test.COZE_STUDIO_ROOT, cozeRoot);
  assert.equal(mod.__test.resolveRootFromScript(), cozeRoot);

  const r = runCore(['help'], { cwd: join(cozeRoot, '..') });
  assert.equal(r.status, 0);
  assert.match(r.stdout, /coze-studio/);
});

test('secret env values are redacted', async () => {
  const mod = await import(pathToFileURL(core).href);
  const sample =
    'FORMA_SECRET_MASTER_KEY=abc123secret password=hunter2 Authorization: Bearer tok123 api_key=zzz';
  const out = mod.__test.redact(sample);
  assert.doesNotMatch(out, /abc123secret/);
  assert.doesNotMatch(out, /hunter2/);
  assert.doesNotMatch(out, /tok123/);
  assert.doesNotMatch(out, /zzz/);
  assert.match(out, /FORMA_SECRET_MASTER_KEY=<redacted>/);
});

test('pid stale handling: dead pid is not alive', async () => {
  const mod = await import(pathToFileURL(core).href);
  assert.equal(mod.__test.pidAlive(99999999), false);
  assert.equal(mod.__test.pidAlive(process.pid), true);
});

test('port conflict detection', async () => {
  const mod = await import(pathToFileURL(core).href);
  const server = createServer();
  await new Promise((resolvePromise) => server.listen(0, '127.0.0.1', resolvePromise));
  const { port } = server.address();
  assert.equal(await mod.__test.portListening(port), true);
  await new Promise((resolvePromise) => server.close(resolvePromise));
  assert.equal(await mod.__test.portListening(port), false);
});

test('doctor runs without modifying system (exit 0 or 1)', () => {
  const r = runCore(['doctor']);
  assert.ok(r.status === 0 || r.status === 1, `unexpected status ${r.status}\n${r.stdout}\n${r.stderr}`);
  assert.match(r.stdout, /DOCTOR = (PASS|FAIL)/);
  assert.doesNotMatch(r.stdout, /FORMA_SECRET_MASTER_KEY=[A-Za-z0-9+/=]{20,}/);
});

test('parse helpers', async () => {
  const mod = await import(pathToFileURL(core).href);
  assert.equal(mod.__test.parseGoVersion('module x\n\ngo 1.24.0\n'), '1.24.0');
  assert.equal(mod.__test.parseNodeRange('"nodeSupportedVersionRange": ">=21"'), '>=21');
  assert.equal(mod.__test.nodeSatisfies('>=21', '22.22.0'), true);
  assert.equal(mod.__test.nodeSatisfies('>=21', '18.0.0'), false);
  assert.equal(mod.__test.dockerPath('D:\\a\\b'), '/d/a/b');
});

test('start fails fast when frontend port is occupied', async () => {
  const holder = createServer();
  try {
    await new Promise((resolvePromise, reject) => {
      holder.once('error', reject);
      holder.listen(3001, '127.0.0.1', resolvePromise);
    });
  } catch {
    return; // environment already has 3001 busy; unit covered elsewhere
  }
  const r = runCore(['start']);
  holder.close();
  assert.equal(r.status, 1);
  assert.match(`${r.stdout}${r.stderr}`, /port 3001/);
});

test('bash -n syntax check when bash available', () => {
  const sh = join(localDir, 'forma-local.sh');
  const bash =
    process.platform === 'win32'
      ? 'C:\\Program Files\\Git\\bin\\bash.exe'
      : 'bash';
  const r = spawnSync(bash, ['-n', sh], { encoding: 'utf8' });
  if (r.error && r.error.code === 'ENOENT') {
    return; // skip
  }
  assert.equal(r.status, 0, r.stderr);
});
