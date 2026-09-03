#!/usr/bin/env node
/**
 * Forma local development launcher — shared core.
 *
 * Topology (canonical):
 *   Browser → Forma FE :3001 → proxy /api/forma → Coze/Forma Backend :8888 → middleware
 *
 * Does NOT wrap forma-live-harness / G6 disposable DBs / Playwright.
 */

import { spawn, spawnSync } from 'node:child_process';
import { createConnection } from 'node:net';
import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
  appendFileSync,
  unlinkSync,
  copyFileSync,
  rmSync,
  openSync,
  statSync,
} from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { randomBytes } from 'node:crypto';
import { setTimeout as delay } from 'node:timers/promises';

const __dirname = dirname(fileURLToPath(import.meta.url));
const COZE_STUDIO_ROOT = resolve(__dirname, '..', '..', '..');
const REPO_ROOT = resolve(COZE_STUDIO_ROOT, '..');

const FRONTEND_PORT = 3001;
const BACKEND_PORT = 8888;
const MIDDLEWARE_PORTS = [
  { name: 'MySQL', port: 3306 },
  { name: 'Redis', port: 6379 },
  { name: 'Elasticsearch', port: 9200 },
  { name: 'MinIO', port: 9000 },
  { name: 'Milvus', port: 19530 },
  { name: 'NSQd', port: 4150 },
  { name: 'etcd', port: 2379 },
];

const RUNTIME_DIR = join(COZE_STUDIO_ROOT, '.forma-local');
const LOG_DIR = join(RUNTIME_DIR, 'logs');
const STATE_DIR = join(RUNTIME_DIR, 'runtime');
const LOCAL_ENV = join(RUNTIME_DIR, '.env');
const ENV_DEBUG = join(COZE_STUDIO_ROOT, 'docker', '.env.debug');
const ENV_DEBUG_EXAMPLE = join(COZE_STUDIO_ROOT, 'docker', '.env.debug.example');
const LOCAL_ENV_EXAMPLE = join(__dirname, '.forma-local.env.example');
const COMPOSE_FILE = join(COZE_STUDIO_ROOT, 'docker', 'docker-compose-debug.yml');
const FORMA_ATLAS_DIR = join(COZE_STUDIO_ROOT, 'docker', 'atlas', 'forma');
const BACKEND_DIR = join(COZE_STUDIO_ROOT, 'backend');
const BIN_DIR = join(COZE_STUDIO_ROOT, 'bin');
const FORMA_APP_DIR = join(COZE_STUDIO_ROOT, 'frontend', 'apps', 'forma');
const SERVER_SH = join(COZE_STUDIO_ROOT, 'scripts', 'setup', 'server.sh');
const RUSH_JS = join(COZE_STUDIO_ROOT, 'common', 'scripts', 'install-run-rush.js');
const RUSHX_JS = join(COZE_STUDIO_ROOT, 'common', 'scripts', 'install-run-rushx.js');
const GO_MOD = join(BACKEND_DIR, 'go.mod');
const RUSH_JSON = join(COZE_STUDIO_ROOT, 'rush.json');

const PORTS_OVERRIDE = join(__dirname, 'docker-compose.ports.override.yml');
const MYSQL_HOST_PORT_DEFAULT = 3306;
const MYSQL_HOST_PORT_FALLBACK = 13306;
const ATLAS_IMAGE = process.env.FORMA_LOCAL_ATLAS_IMAGE || 'arigaio/atlas:0.35.0-community-alpine';
const GOLANG_IMAGE = process.env.FORMA_LOCAL_GOLANG_IMAGE || 'golang:1.24-bookworm';
const BACKEND_RUNTIME_IMAGE =
  process.env.FORMA_LOCAL_BACKEND_IMAGE || 'golang:1.24-bookworm'; // has ca-certificates; avoids apt on every start
const COMPOSE_PROJECT = 'coze-studio-debug';
const COMPOSE_NETWORK = `${COMPOSE_PROJECT}_coze-network`;
const BACKEND_CONTAINER = 'forma-local-backend';
const LINUX_BIN = join(BIN_DIR, 'opencoze-linux');

const PAGES = [
  '/',
  '/business',
  '/analyst',
  '/data',
  '/data/requirements',
  '/data/sources',
  '/data/mappings',
  '/data/contracts',
  '/data/health',
];

const SECRET_SCAN =
  /(api[_-]?key|authorization|password|cookie|session_key|FORMA_SECRET_MASTER_KEY)(\s*[=:]\s*|\s+)(Bearer\s+)?['"]?[^\s'"]+/gi;

const args = process.argv.slice(2);
const command = (args[0] || 'help').toLowerCase();
const rest = args.slice(1);

function say(msg = '') {
  process.stdout.write(`${redact(String(msg))}\n`);
}

function err(msg) {
  process.stderr.write(`${redact(String(msg))}\n`);
}

function redact(text) {
  return String(text).replace(SECRET_SCAN, (m) => {
    const lower = m.toLowerCase();
    if (lower.includes('forma_secret_master_key')) {
      return 'FORMA_SECRET_MASTER_KEY=<redacted>';
    }
    const key = m.match(/^[^=:\s]+/i);
    const name = key ? key[0] : 'secret';
    return `${name}=<redacted>`;
  });
}

function ensureDirs() {
  mkdirSync(LOG_DIR, { recursive: true });
  mkdirSync(STATE_DIR, { recursive: true });
  mkdirSync(BIN_DIR, { recursive: true });
}

function statePath(name) {
  return join(STATE_DIR, name);
}

function logPath(name) {
  return join(LOG_DIR, `${name}.log`);
}

function writeState(name, value) {
  ensureDirs();
  writeFileSync(statePath(name), String(value), 'utf8');
}

function readState(name) {
  const p = statePath(name);
  if (!existsSync(p)) return null;
  return readFileSync(p, 'utf8').trim();
}

function clearState(name) {
  const p = statePath(name);
  if (existsSync(p)) unlinkSync(p);
}

function which(cmd) {
  const r = spawnSync(process.platform === 'win32' ? 'where' : 'which', [cmd], {
    encoding: 'utf8',
    shell: false,
  });
  if (r.status !== 0) return null;
  const line = (r.stdout || '').split(/\r?\n/).map((s) => s.trim()).find(Boolean);
  return line || null;
}

function run(cmd, cmdArgs, opts = {}) {
  const r = spawnSync(cmd, cmdArgs, {
    encoding: 'utf8',
    maxBuffer: 20 * 1024 * 1024,
    ...opts,
  });
  return r;
}

function mustOk(r, label) {
  if (r.status !== 0) {
    const detail = [r.stderr, r.stdout].filter(Boolean).join('\n').trim();
    throw new Error(`${label} failed (exit ${r.status})${detail ? `:\n${detail}` : ''}`);
  }
  return r;
}

function dockerPath(hostPath) {
  const normalized = hostPath.replace(/\\/g, '/');
  if (process.platform === 'win32' && /^[A-Za-z]:/.test(normalized)) {
    return `/${normalized[0].toLowerCase()}${normalized.slice(2)}`;
  }
  return normalized;
}

function parseEnvFile(filePath) {
  const out = {};
  if (!existsSync(filePath)) return out;
  const text = readFileSync(filePath, 'utf8');
  for (const raw of text.split(/\r?\n/)) {
    let line = raw.trim();
    if (!line || line.startsWith('#')) continue;
    if (line.startsWith('export ')) line = line.slice(7).trim();
    const eq = line.indexOf('=');
    if (eq <= 0) continue;
    const key = line.slice(0, eq).trim();
    let val = line.slice(eq + 1).trim();
    // Strip unquoted inline comments.
    if (!val.startsWith('"') && !val.startsWith("'")) {
      const hash = val.indexOf(' #');
      if (hash >= 0) val = val.slice(0, hash).trim();
      else if (val.includes('#') && !val.includes('://')) {
        // keep URLs; otherwise cut at bare #
        const bare = val.indexOf('#');
        if (bare > 0) val = val.slice(0, bare).trim();
      }
    } else {
      // Quoted value: find matching closing quote, drop trailing comment.
      const q = val[0];
      const end = val.indexOf(q, 1);
      if (end > 0) val = val.slice(0, end + 1);
    }
    if (
      (val.startsWith('"') && val.endsWith('"')) ||
      (val.startsWith("'") && val.endsWith("'"))
    ) {
      val = val.slice(1, -1);
    }
    out[key] = val;
  }
  return expandEnvMap(out);
}

function expandEnvMap(map) {
  const out = { ...map };
  for (let pass = 0; pass < 5; pass++) {
    let changed = false;
    for (const [k, v] of Object.entries(out)) {
      const next = String(v).replace(/\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)/g, (_, a, b) => {
        const name = a || b;
        if (Object.prototype.hasOwnProperty.call(out, name)) {
          changed = true;
          return out[name];
        }
        return '';
      });
      if (next !== v) {
        out[k] = next;
        changed = true;
      }
    }
    if (!changed) break;
  }
  return out;
}

function loadMergedEnv() {
  const base = parseEnvFile(ENV_DEBUG);
  const local = parseEnvFile(LOCAL_ENV);
  return { ...base, ...local };
}

function secretConfigured(env) {
  const v = env.FORMA_SECRET_MASTER_KEY || '';
  if (!v) return false;
  try {
    const b64 = Buffer.from(v, 'base64');
    if (b64.length === 32) return true;
  } catch {
    /* ignore */
  }
  if (/^[0-9a-fA-F]{64}$/.test(v)) return true;
  return false;
}

function portListening(port, host = '127.0.0.1') {
  return new Promise((resolvePromise) => {
    const socket = createConnection({ port, host }, () => {
      socket.destroy();
      resolvePromise(true);
    });
    socket.on('error', () => resolvePromise(false));
    socket.setTimeout(800, () => {
      socket.destroy();
      resolvePromise(false);
    });
  });
}

async function assertPortFree(port, label) {
  if (await portListening(port)) {
    throw new Error(
      `BLOCKED: port ${port} (${label}) is already in use.\n` +
        `Free the port, or stop the foreign process, then retry.\n` +
        `Canonical Forma frontend port is ${FRONTEND_PORT} (strict — will not silently use another port).`,
    );
  }
}

function parseGoVersion(goModText) {
  const m = goModText.match(/^go\s+(\d+\.\d+(?:\.\d+)?)/m);
  return m ? m[1] : null;
}

function parseNodeRange(rushText) {
  const m = rushText.match(/"nodeSupportedVersionRange"\s*:\s*"([^"]+)"/);
  return m ? m[1] : '>=21';
}

function versionGte(actual, required) {
  const a = actual.replace(/^v/, '').split('.').map(Number);
  const r = required.split('.').map(Number);
  for (let i = 0; i < Math.max(a.length, r.length); i++) {
    const av = a[i] || 0;
    const rv = r[i] || 0;
    if (av > rv) return true;
    if (av < rv) return false;
  }
  return true;
}

function nodeSatisfies(range, version) {
  // Support simple ">=N" / ">=N.M" used by this repo.
  const m = range.match(/^>=\s*(\d+(?:\.\d+)?(?:\.\d+)?)/);
  if (!m) return true;
  return versionGte(version, m[1]);
}

function pidAlive(pid) {
  if (!pid || Number.isNaN(Number(pid))) return false;
  try {
    process.kill(Number(pid), 0);
    return true;
  } catch {
    return false;
  }
}

function stopPid(pid, label) {
  if (!pidAlive(pid)) return false;
  try {
    if (process.platform === 'win32') {
      spawnSync('taskkill', ['/PID', String(pid), '/T', '/F'], { encoding: 'utf8' });
    } else {
      try {
        process.kill(Number(pid), 'SIGTERM');
      } catch {
        /* ignore */
      }
    }
  } catch (e) {
    err(`WARN: failed to stop ${label} pid=${pid}: ${e.message}`);
  }
  return true;
}

function spawnDetached(cmd, cmdArgs, { cwd, logName, env }) {
  ensureDirs();
  const file = logPath(logName);
  appendFileSync(
    file,
    `\n==== ${new Date().toISOString()} start ${cmd} ${cmdArgs.join(' ')} ====\n`,
  );
  const fd = openSync(file, 'a');
  const child = spawn(cmd, cmdArgs, {
    cwd,
    env: { ...process.env, ...env },
    detached: true,
    stdio: ['ignore', fd, fd],
    windowsHide: true,
    shell: false,
  });
  child.unref();
  return child.pid;
}

function findBash() {
  if (process.platform !== 'win32') {
    return which('bash');
  }
  const candidates = [
    'C:\\Program Files\\Git\\bin\\bash.exe',
    'C:\\Program Files\\Git\\usr\\bin\\bash.exe',
    which('bash'),
  ].filter(Boolean);
  for (const c of candidates) {
    if (existsSync(c)) return c;
  }
  return null;
}

function preferNativeBackend() {
  if (process.platform === 'win32') return false;
  return Boolean(which('go') && findBash() && existsSync(SERVER_SH));
}

async function fetchWithTimeout(url, ms = 2000) {
  const ac = new AbortController();
  const t = setTimeout(() => ac.abort(), ms);
  try {
    return await fetch(url, { redirect: 'manual', signal: ac.signal });
  } finally {
    clearTimeout(t);
  }
}

async function waitForHttp(url, { timeoutMs = 120_000, ok = (res) => res.status > 0 } = {}) {
  const deadline = Date.now() + timeoutMs;
  let last = '';
  while (Date.now() < deadline) {
    try {
      const res = await fetchWithTimeout(url, 2500);
      last = `HTTP ${res.status}`;
      if (ok(res)) return res;
    } catch (e) {
      last = e.message;
    }
    await delay(1500);
  }
  throw new Error(`Timeout waiting for ${url} (${last})`);
}

function ensureEnvFiles() {
  ensureDirs();
  if (!existsSync(ENV_DEBUG)) {
    if (!existsSync(ENV_DEBUG_EXAMPLE)) {
      throw new Error(`BLOCKED: missing ${ENV_DEBUG_EXAMPLE}`);
    }
    copyFileSync(ENV_DEBUG_EXAMPLE, ENV_DEBUG);
    say(`Created ${rel(ENV_DEBUG)} from example.`);
  }
  if (!existsSync(LOCAL_ENV) && existsSync(LOCAL_ENV_EXAMPLE)) {
    copyFileSync(LOCAL_ENV_EXAMPLE, LOCAL_ENV);
    say(`Created ${rel(LOCAL_ENV)} from example (optional Forma overrides).`);
  }
}

function rel(p) {
  return p.startsWith(COZE_STUDIO_ROOT)
    ? p.slice(COZE_STUDIO_ROOT.length + 1).replace(/\\/g, '/')
    : p;
}

function mysqlHostPort() {
  const forced = process.env.FORMA_LOCAL_MYSQL_HOST_PORT;
  if (forced) return Number(forced);
  const st = readState('mysql.host_port');
  if (st) return Number(st);
  return MYSQL_HOST_PORT_DEFAULT;
}

async function decideMysqlHostPort() {
  const forced = process.env.FORMA_LOCAL_MYSQL_HOST_PORT;
  if (forced) {
    writeState('mysql.host_port', String(forced));
    return Number(forced);
  }
  // If coze-mysql already running with published ports, prefer that mapping.
  const inspect = run(
    'docker',
    ['inspect', '-f', '{{(index (index .NetworkSettings.Ports "3306/tcp") 0).HostPort}}', 'coze-mysql'],
    { encoding: 'utf8' },
  );
  if (inspect.status === 0) {
    const hp = String(inspect.stdout || '').trim();
    if (/^\d+$/.test(hp)) {
      writeState('mysql.host_port', hp);
      return Number(hp);
    }
  }
  if (await portListening(MYSQL_HOST_PORT_DEFAULT)) {
    // Host MySQL (or other) owns 3306 — publish Coze MySQL on fallback.
    writeState('mysql.host_port', String(MYSQL_HOST_PORT_FALLBACK));
    say(
      `WARN: host port ${MYSQL_HOST_PORT_DEFAULT} busy — using compose ports override → 127.0.0.1:${MYSQL_HOST_PORT_FALLBACK}`,
    );
    return MYSQL_HOST_PORT_FALLBACK;
  }
  writeState('mysql.host_port', String(MYSQL_HOST_PORT_DEFAULT));
  return MYSQL_HOST_PORT_DEFAULT;
}

function composeFilesArgs() {
  const files = ['-f', COMPOSE_FILE];
  if (mysqlHostPort() === MYSQL_HOST_PORT_FALLBACK || process.env.FORMA_LOCAL_MYSQL_HOST_PORT === String(MYSQL_HOST_PORT_FALLBACK)) {
    files.push('-f', PORTS_OVERRIDE);
  }
  return files;
}

function dockerCompose(args, opts = {}) {
  return run(
    'docker',
    [
      'compose',
      '-p',
      COMPOSE_PROJECT,
      ...composeFilesArgs(),
      '--env-file',
      ENV_DEBUG,
      ...args,
    ],
    { cwd: join(COZE_STUDIO_ROOT, 'docker'), ...opts },
  );
}

async function cmdDoctor() {
  const checks = [];
  const push = (name, status, detail) => checks.push({ name, status, detail });

  // Git
  const git = which('git');
  push('Git', git ? 'PASS' : 'FAIL', git || 'git not found');

  // Docker
  const dockerBin = which('docker');
  if (!dockerBin) {
    push('Docker', 'FAIL', 'docker not found');
  } else {
    const info = run('docker', ['info'], { encoding: 'utf8' });
    push(
      'Docker daemon',
      info.status === 0 ? 'PASS' : 'FAIL',
      info.status === 0 ? 'reachable' : (info.stderr || 'daemon not reachable').split('\n')[0],
    );
    const dc = run('docker', ['compose', 'version'], { encoding: 'utf8' });
    push(
      'docker compose',
      dc.status === 0 ? 'PASS' : 'FAIL',
      (dc.stdout || dc.stderr || '').trim().split('\n')[0] || 'missing',
    );
  }

  // Node
  const nodeVer = process.versions.node;
  const rushText = existsSync(RUSH_JSON) ? readFileSync(RUSH_JSON, 'utf8') : '';
  const nodeRange = parseNodeRange(rushText);
  push(
    'Node.js',
    nodeSatisfies(nodeRange, nodeVer) ? 'PASS' : 'FAIL',
    `v${nodeVer} (required ${nodeRange})`,
  );

  // Go
  const goBin = which('go');
  const requiredGo = existsSync(GO_MOD)
    ? parseGoVersion(readFileSync(GO_MOD, 'utf8'))
    : null;
  if (preferNativeBackend()) {
    if (!goBin) {
      push('Go', 'FAIL', `required for native backend (go ${requiredGo})`);
    } else {
      const gv = run('go', ['version'], { encoding: 'utf8' });
      const m = (gv.stdout || '').match(/go(\d+\.\d+(?:\.\d+)?)/);
      const ok = m && requiredGo ? versionGte(m[1], requiredGo) : gv.status === 0;
      push('Go', ok ? 'PASS' : 'FAIL', (gv.stdout || '').trim() || 'go missing');
    }
  } else {
    push(
      'Go',
      dockerBin ? 'PASS' : 'WARN',
      process.platform === 'win32'
        ? 'Windows uses Docker to build real opencoze (native Go+milvus unsupported)'
        : goBin
          ? (run('go', ['version'], { encoding: 'utf8' }).stdout || '').trim()
          : 'Go missing — backend will use Docker build of real main.go',
    );
  }

  // Rush deps
  const pnpmLock = join(COZE_STUDIO_ROOT, 'common', 'config', 'subspaces', 'default', 'pnpm-lock.yaml');
  const rushInstalled =
    existsSync(join(COZE_STUDIO_ROOT, 'common', 'temp', 'default')) ||
    existsSync(join(FORMA_APP_DIR, 'node_modules'));
  push(
    'Rush workspace',
    existsSync(pnpmLock) ? (rushInstalled ? 'PASS' : 'WARN') : 'FAIL',
    rushInstalled
      ? 'lockfile + install state present'
      : existsSync(pnpmLock)
        ? 'lockfile present; run start (will rush install) or: node common/scripts/install-run-rush.js install'
        : 'missing pnpm-lock.yaml',
  );

  // Required files
  const requiredFiles = [
    COMPOSE_FILE,
    ENV_DEBUG_EXAMPLE,
    SERVER_SH,
    RUSH_JS,
    RUSHX_JS,
    join(FORMA_APP_DIR, 'package.json'),
    join(FORMA_APP_DIR, 'rsbuild.config.ts'),
    join(FORMA_ATLAS_DIR, 'migrations'),
    join(BACKEND_DIR, 'main.go'),
  ];
  const missing = requiredFiles.filter((f) => !existsSync(f));
  push(
    'Required files',
    missing.length === 0 ? 'PASS' : 'FAIL',
    missing.length === 0 ? 'ok' : `missing: ${missing.map(rel).join(', ')}`,
  );

  // Ports
  for (const p of [
    { name: 'Frontend', port: FRONTEND_PORT },
    { name: 'Backend', port: BACKEND_PORT },
    ...MIDDLEWARE_PORTS,
  ]) {
    const busy = await portListening(p.port);
    const owned =
      (p.port === FRONTEND_PORT && pidAlive(readState('frontend.pid'))) ||
      (p.port === BACKEND_PORT &&
        (pidAlive(readState('backend.pid')) || (await dockerBackendRunning())));
    push(
      `Port ${p.port} (${p.name})`,
      !busy || owned ? 'PASS' : 'WARN',
      busy ? (owned ? 'in use by forma-local' : 'OCCUPIED') : 'free',
    );
  }

  // Disk
  try {
    const st = statSync(COZE_STUDIO_ROOT);
    push('Disk (workspace)', st ? 'PASS' : 'WARN', COZE_STUDIO_ROOT);
  } catch (e) {
    push('Disk (workspace)', 'FAIL', e.message);
  }

  // Secret
  ensureEnvFiles();
  const env = loadMergedEnv();
  push(
    'FORMA_SECRET_MASTER_KEY',
    secretConfigured(env) ? 'PASS' : 'WARN',
    secretConfigured(env)
      ? 'configured = YES'
      : 'not set — Preview OK; DataSource Credential will be unavailable',
  );

  say('Forma Local Doctor');
  say(`COZE_STUDIO_ROOT=${COZE_STUDIO_ROOT}`);
  say('');
  let fails = 0;
  for (const c of checks) {
    say(`${c.status.padEnd(4)}  ${c.name.padEnd(28)}  ${c.detail}`);
    if (c.status === 'FAIL') fails += 1;
  }
  say('');
  if (fails > 0) {
    say(`DOCTOR = FAIL (${fails} failing check(s))`);
    process.exitCode = 1;
  } else {
    say('DOCTOR = PASS');
  }
}

async function dockerBackendRunning() {
  const r = run('docker', ['inspect', '-f', '{{.State.Running}}', BACKEND_CONTAINER], {
    encoding: 'utf8',
  });
  return r.status === 0 && String(r.stdout).trim() === 'true';
}

async function startMiddleware() {
  await decideMysqlHostPort();
  say('→ Starting middleware (docker compose --profile middleware)...');
  const r = dockerCompose(['--profile', 'middleware', 'up', '-d', '--wait']);
  if (r.status !== 0) {
    appendFileSync(logPath('middleware'), `${r.stdout || ''}\n${r.stderr || ''}\n`);
    throw new Error(
      `BLOCKED: middleware start failed.\nSee ${rel(logPath('middleware'))}\n${(r.stderr || r.stdout || '').slice(0, 2000)}`,
    );
  }
  appendFileSync(logPath('middleware'), `${r.stdout || ''}\n`);
  say(`  Middleware UP (MySQL host port ${mysqlHostPort()})`);
}

async function applyFormaMigrations() {
  say('→ Applying Forma Atlas migrations (existing only)...');
  const env = loadMergedEnv();
  const user = env.MYSQL_USER || 'coze';
  const pass = env.MYSQL_PASSWORD || 'coze123';
  const db = env.MYSQL_DATABASE || 'opencoze';

  // Prefer compose network + mysql service DNS.
  const url = `mysql://${user}:${pass}@mysql:3306/${db}?charset=utf8mb4&parseTime=True`;
  const mount = `${dockerPath(FORMA_ATLAS_DIR)}:/forma-atlas`;

  // Ensure network exists
  const net = run('docker', ['network', 'inspect', COMPOSE_NETWORK], { encoding: 'utf8' });
  if (net.status !== 0) {
    throw new Error(`BLOCKED: compose network missing: ${COMPOSE_NETWORK}`);
  }

  const apply = run(
    'docker',
    [
      'run',
      '--rm',
      '--network',
      COMPOSE_NETWORK,
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
    ],
    { encoding: 'utf8' },
  );
  appendFileSync(
    logPath('migration'),
    `\n==== ${new Date().toISOString()} ====\n${apply.stdout || ''}\n${apply.stderr || ''}\n`,
  );
  if (apply.status !== 0) {
    throw new Error(
      `BLOCKED: Forma migration apply failed.\n${(apply.stderr || apply.stdout || '').slice(0, 2000)}`,
    );
  }
  say('  Forma migrations applied (idempotent)');
}

function writeDockerBackendEnv(env) {
  ensureDirs();
  const dockerEnv = {
    ...env,
    MYSQL_HOST: 'mysql',
    MYSQL_PORT: '3306',
    MYSQL_DSN: `${env.MYSQL_USER || 'coze'}:${env.MYSQL_PASSWORD || 'coze123'}@tcp(mysql:3306)/${env.MYSQL_DATABASE || 'opencoze'}?charset=utf8mb4&parseTime=True`,
    ATLAS_URL: `mysql://${env.MYSQL_USER || 'coze'}:${env.MYSQL_PASSWORD || 'coze123'}@mysql:3306/${env.MYSQL_DATABASE || 'opencoze'}?charset=utf8mb4&parseTime=True`,
    REDIS_ADDR: 'redis:6379',
    ES_ADDR: 'http://elasticsearch:9200',
    MINIO_ENDPOINT: 'minio:9000',
    MINIO_API_HOST: 'http://minio:9000',
    MILVUS_ADDR: 'milvus:19530',
    MQ_NAME_SERVER: 'nsqd:4150',
    LISTEN_ADDR: ':8888',
    SERVER_HOST: 'http://localhost:8888',
    WEB_LISTEN_ADDR: '0.0.0.0:8888',
  };
  const lines = Object.entries(dockerEnv)
    .filter(([, v]) => v !== undefined && v !== null)
    .map(([k, v]) => {
      // docker --env-file / godotenv: avoid comments and bare spaces; quote if needed
      const s = String(v);
      if (/[\s#"']/.test(s) || s === '') {
        return `${k}="${s.replace(/\\/g, '\\\\').replace(/"/g, '\\"')}"`;
      }
      return `${k}=${s}`;
    });
  const path = join(STATE_DIR, 'backend.docker.env');
  writeFileSync(path, `${lines.join('\n')}\n`, 'utf8');
  return path;
}

async function buildLinuxOpencoze() {
  const force = process.env.FORMA_LOCAL_REBUILD_BACKEND === '1';
  if (!force && existsSync(LINUX_BIN) && statSync(LINUX_BIN).size > 1_000_000) {
    say(`  Reusing existing ${rel(LINUX_BIN)} (set FORMA_LOCAL_REBUILD_BACKEND=1 to rebuild)`);
    return;
  }
  say('→ Building real Coze backend (opencoze main.go) via Docker golang...');
  mkdirSync(BIN_DIR, { recursive: true });
  const outMount = `${dockerPath(BIN_DIR)}:/out`;
  const srcMount = `${dockerPath(BACKEND_DIR)}:/src`;
  const r = run(
    'docker',
    [
      'run',
      '--rm',
      '-v',
      srcMount,
      '-v',
      outMount,
      '-w',
      '/src',
      GOLANG_IMAGE,
      'go',
      'build',
      '-ldflags=-s -w',
      '-o',
      '/out/opencoze-linux',
      'main.go',
    ],
    { encoding: 'utf8' },
  );
  appendFileSync(
    logPath('backend-build'),
    `\n==== ${new Date().toISOString()} ====\n${r.stdout || ''}\n${r.stderr || ''}\n`,
  );
  if (r.status !== 0 || !existsSync(LINUX_BIN)) {
    throw new Error(
      `BLOCKED: opencoze Docker build failed.\nSee ${rel(logPath('backend-build'))}\n${(r.stderr || r.stdout || '').slice(0, 2000)}`,
    );
  }
  say(`  Built ${rel(LINUX_BIN)}`);
}

async function startBackendDocker() {
  const env = loadMergedEnv();
  const envFile = writeDockerBackendEnv(env);
  await buildLinuxOpencoze();

  // Remove stale container
  run('docker', ['rm', '-f', BACKEND_CONTAINER], { encoding: 'utf8' });

  say('→ Starting backend container (real /app/opencoze)...');
  const confMount = `${dockerPath(join(BACKEND_DIR, 'conf'))}:/app/resources/conf:ro`;
  const binMount = `${dockerPath(LINUX_BIN)}:/app/opencoze:ro`;
  const envMount = `${dockerPath(envFile)}:/app/.env:ro`;
  const r = run(
    'docker',
    [
      'run',
      '-d',
      '--name',
      BACKEND_CONTAINER,
      '--network',
      COMPOSE_NETWORK,
      '-p',
      `127.0.0.1:${BACKEND_PORT}:8888`,
      '--env-file',
      envFile,
      '-v',
      binMount,
      '-v',
      confMount,
      '-v',
      envMount,
      '-w',
      '/app',
      BACKEND_RUNTIME_IMAGE,
      '/app/opencoze',
    ],
    { encoding: 'utf8' },
  );
  if (r.status !== 0) {
    throw new Error(`BLOCKED: backend container start failed:\n${r.stderr || r.stdout}`);
  }
  const cid = String(r.stdout).trim();
  writeState('backend.mode', 'docker');
  writeState('backend.cid', cid);
  clearState('backend.pid');
  say(`  Backend container ${cid.slice(0, 12)}`);
}

async function startBackendNative() {
  say('→ Starting backend via scripts/setup/server.sh (canonical)...');
  const bash = findBash();
  if (!bash) throw new Error('BLOCKED: bash required for native backend start');
  const env = loadMergedEnv();
  const pid = spawnDetached(bash, [SERVER_SH, '-start'], {
    cwd: COZE_STUDIO_ROOT,
    logName: 'backend',
    env: {
      APP_ENV: 'debug',
      ...env,
      MYSQL_PORT: String(mysqlHostPort()),
      MYSQL_HOST: '127.0.0.1',
      MYSQL_DSN: `${env.MYSQL_USER || 'coze'}:${env.MYSQL_PASSWORD || 'coze123'}@tcp(127.0.0.1:${mysqlHostPort()})/${env.MYSQL_DATABASE || 'opencoze'}?charset=utf8mb4&parseTime=True`,
    },
  });
  writeState('backend.mode', 'native');
  writeState('backend.pid', String(pid));
  clearState('backend.cid');
  say(`  Backend pid=${pid}`);
}

async function ensureRushInstall() {
  const marker = join(FORMA_APP_DIR, 'node_modules', '@rsbuild', 'core');
  if (existsSync(marker)) {
    say('  Rush deps already present for @forma/app');
    return;
  }
  say('→ Rush install (node common/scripts/install-run-rush.js install)...');
  const r = run('node', [RUSH_JS, 'install'], {
    cwd: COZE_STUDIO_ROOT,
    encoding: 'utf8',
  });
  appendFileSync(
    logPath('rush-install'),
    `\n==== ${new Date().toISOString()} ====\n${r.stdout || ''}\n${r.stderr || ''}\n`,
  );
  if (r.status !== 0) {
    throw new Error(
      `BLOCKED: rush install failed.\nSee ${rel(logPath('rush-install'))}\n${(r.stderr || r.stdout || '').slice(0, 2000)}`,
    );
  }
}

async function startFrontend() {
  await ensureRushInstall();
  say('→ Starting Forma frontend (@forma/app rsbuild dev) on :3001...');
  const pid = spawnDetached(process.execPath, [RUSHX_JS, 'dev'], {
    cwd: FORMA_APP_DIR,
    logName: 'frontend',
    env: {
      // Force rsbuild onto canonical port; wrapper already asserted free.
      PORT: String(FRONTEND_PORT),
    },
  });
  writeState('frontend.pid', String(pid));
  say(`  Frontend pid=${pid}`);
}

async function waitReady() {
  say('→ Health checks...');
  await waitForHttp(`http://127.0.0.1:${BACKEND_PORT}/api/forma/v1/health`, {
    timeoutMs: 180_000,
    ok: (res) => res.status === 200,
  });
  say('  Backend /api/forma/v1/health OK');
  await waitForHttp(`http://127.0.0.1:${FRONTEND_PORT}/`, {
    timeoutMs: 180_000,
    ok: (res) => res.status === 200,
  });
  say('  Frontend / OK');
}

function printReadyBanner() {
  const env = loadMergedEnv();
  const secret = secretConfigured(env);
  say('');
  say('========================================');
  say('Forma Local Development READY');
  say('========================================');
  say('');
  say('Frontend:');
  say(`  http://localhost:${FRONTEND_PORT}`);
  say('');
  say('Backend:');
  say(`  http://localhost:${BACKEND_PORT}`);
  say('');
  say('Pages:');
  for (const p of PAGES) {
    say(`  http://localhost:${FRONTEND_PORT}${p}`);
  }
  say('');
  say(`Secret master key configured = ${secret ? 'YES' : 'NO'}`);
  if (!secret) {
    say('WARN: DataSource Credential features need a local master key.');
    say('      Generate: node scripts/forma/local/forma-local.mjs gen-secret');
  }
  say('AI: Preview mode does not require LLM keys.');
  say('    Full AI: configure Coze model YAML under backend/conf/model/');
  say('');
  say('Known product issues (not startup failures):');
  say('  - Mapping EditConfirm UI pending G6-F2');
  say('  - Drift Snapshot Picker pending G6-F2');
  say('');
}

async function cmdStart() {
  ensureDirs();
  ensureEnvFiles();

  // Fail fast on canonical ports before any long-running work / HTTP probes.
  const feOwned = pidAlive(readState('frontend.pid'));
  const beOwned =
    pidAlive(readState('backend.pid')) || (await dockerBackendRunning());
  if (!feOwned && (await portListening(FRONTEND_PORT))) {
    // Distinguish our healthy FE vs foreign occupant.
    let ours = false;
    try {
      const res = await fetchWithTimeout(`http://127.0.0.1:${FRONTEND_PORT}/`, 1500);
      ours = res.status === 200 && feOwned;
    } catch {
      ours = false;
    }
    if (!ours) {
      await assertPortFree(FRONTEND_PORT, 'Forma frontend');
    }
  }
  if (!beOwned && (await portListening(BACKEND_PORT))) {
    await assertPortFree(BACKEND_PORT, 'Forma/Coze backend');
  }

  // If already managed and healthy, report READY
  const status = await collectStatus();
  if (status.frontend.ok && status.backend.ok && status.formaApi.ok) {
    say('Already running.');
    printReadyBanner();
    return;
  }

  if (!feOwned) {
    await assertPortFree(FRONTEND_PORT, 'Forma frontend');
  }
  if (!beOwned) {
    await assertPortFree(BACKEND_PORT, 'Forma/Coze backend');
  }

  await startMiddleware();
  await applyFormaMigrations();

  if (preferNativeBackend()) {
    await startBackendNative();
  } else {
    await startBackendDocker();
  }

  await startFrontend();
  await waitReady();
  printReadyBanner();
}

async function stopBackend() {
  const mode = readState('backend.mode');
  const pid = readState('backend.pid');
  const cid = readState('backend.cid');
  if (mode === 'docker' || cid || (await dockerBackendRunning())) {
    run('docker', ['rm', '-f', BACKEND_CONTAINER], { encoding: 'utf8' });
    say('  Backend container stopped');
  }
  if (pid) {
    stopPid(pid, 'backend');
    say(`  Backend pid ${pid} stop attempted`);
  }
  clearState('backend.pid');
  clearState('backend.cid');
  clearState('backend.mode');
}

async function stopFrontend() {
  const pid = readState('frontend.pid');
  if (pid) {
    stopPid(pid, 'frontend');
    say(`  Frontend pid ${pid} stop attempted`);
  }
  clearState('frontend.pid');
}

async function cmdStop() {
  const all = rest.includes('--all');
  say('→ Stopping Forma local app processes...');
  await stopFrontend();
  await stopBackend();
  if (all) {
    say('→ Stopping middleware (compose down)...');
    const r = dockerCompose(['--profile', '*', 'down']);
    appendFileSync(logPath('middleware'), `${r.stdout || ''}\n${r.stderr || ''}\n`);
    say('  Middleware down');
  } else {
    say('  Middleware left running (use: stop --all)');
  }
  say('STOP = OK');
}

async function cmdRestart() {
  say('→ Restart (no DB wipe)...');
  await stopFrontend();
  await stopBackend();
  await cmdStart();
}

async function collectStatus() {
  const dockerOk = run('docker', ['info'], { encoding: 'utf8' }).status === 0;
  const mysql = await portListening(mysqlHostPort());
  const redis = await portListening(6379);
  const backendHttp = await portListening(BACKEND_PORT);
  let formaApi = false;
  let formaDetail = 'n/a';
  if (backendHttp) {
    try {
      const res = await fetchWithTimeout(`http://127.0.0.1:${BACKEND_PORT}/api/forma/v1/health`, 2000);
      formaApi = res.status === 200;
      formaDetail = `HTTP ${res.status}`;
    } catch (e) {
      formaDetail = e.message;
    }
  }
  let frontendOk = false;
  let frontendDetail = 'n/a';
  if (await portListening(FRONTEND_PORT)) {
    try {
      const res = await fetchWithTimeout(`http://127.0.0.1:${FRONTEND_PORT}/`, 2000);
      frontendOk = res.status === 200;
      frontendDetail = `HTTP ${res.status}`;
    } catch (e) {
      frontendDetail = e.message;
    }
  }
  return {
    docker: dockerOk,
    mysql,
    redis,
    backend: { ok: backendHttp && formaApi, listening: backendHttp },
    formaApi: { ok: formaApi, detail: formaDetail },
    frontend: { ok: frontendOk, detail: frontendDetail },
  };
}

async function cmdStatus() {
  const s = await collectStatus();
  const line = (name, ok, detail) =>
    say(`${name.padEnd(12)} ${ok ? 'OK' : 'DOWN'}  ${detail || ''}`);

  say('Forma Local Status');
  say(`Docker       ${s.docker ? 'RUNNING' : 'DOWN'}`);
  say(`MySQL        ${s.mysql ? 'HEALTHY' : 'DOWN'}  :${mysqlHostPort()}`);
  say(`Redis        ${s.redis ? 'HEALTHY' : 'DOWN'}  :6379`);
  line('Backend', s.backend.ok, `http://localhost:${BACKEND_PORT}`);
  line('Forma API', s.formaApi.ok, s.formaApi.detail);
  line('Frontend', s.frontend.ok, `http://localhost:${FRONTEND_PORT}`);
  say('');
  say(`backend.mode = ${readState('backend.mode') || '(none)'}`);
  say(`frontend.pid = ${readState('frontend.pid') || '(none)'} alive=${pidAlive(readState('frontend.pid'))}`);
  const env = loadMergedEnv();
  say(`Secret master key configured = ${secretConfigured(env) ? 'YES' : 'NO'}`);
}

function tailFile(file, lines = 80) {
  if (!existsSync(file)) {
    say(`(no log) ${rel(file)}`);
    return;
  }
  const text = readFileSync(file, 'utf8');
  const arr = text.split(/\r?\n/);
  const slice = arr.slice(Math.max(0, arr.length - lines)).join('\n');
  say(`----- ${rel(file)} -----`);
  say(redact(slice));
}

async function cmdLogs() {
  const target = (rest[0] || 'all').toLowerCase();
  if (target === 'backend' || target === 'all') {
    if ((await dockerBackendRunning()) || readState('backend.mode') === 'docker') {
      const r = run('docker', ['logs', '--tail', '80', BACKEND_CONTAINER], { encoding: 'utf8' });
      say('----- docker:forma-local-backend -----');
      say(redact(`${r.stdout || ''}${r.stderr || ''}`));
    }
    tailFile(logPath('backend'));
  }
  if (target === 'frontend' || target === 'all') {
    tailFile(logPath('frontend'));
  }
  if (target === 'middleware') {
    tailFile(logPath('middleware'));
  }
  if (!['backend', 'frontend', 'middleware', 'all'].includes(target)) {
    err(`Unknown logs target: ${target}`);
    process.exitCode = 2;
  }
}

async function cmdReset() {
  const wipe = rest.includes('--data');
  if (!wipe) {
    say('reset without --data only stops app processes.');
    say('To wipe docker volumes: reset --data  (requires typing YES)');
    await cmdStop();
    return;
  }
  const confirm = rest.find((a) => a === 'YES' || a === '--yes');
  if (!confirm) {
    err('BLOCKED: destructive reset requires explicit YES');
    err('Example: forma-local.mjs reset --data YES');
    process.exitCode = 2;
    return;
  }
  await cmdStop();
  say('→ compose down + removing docker/data (DESTRUCTIVE)...');
  dockerCompose(['--profile', '*', 'down', '-v']);
  const dataDir = join(COZE_STUDIO_ROOT, 'docker', 'data');
  if (existsSync(dataDir)) {
    rmSync(dataDir, { recursive: true, force: true });
  }
  say('RESET --data complete');
}

function cmdGenSecret() {
  const key = randomBytes(32).toString('base64');
  say('Generated local FORMA_SECRET_MASTER_KEY (32-byte base64).');
  say('Write it into gitignored coze-studio/.forma-local/.env as:');
  say('');
  // Print key alone on its line for copy — still a secret; user asked to generate.
  // We allow printing HERE only for gen-secret command output (not logs/status).
  process.stdout.write(`FORMA_SECRET_MASTER_KEY=${key}\n`);
  say('');
  say('Do not commit this value. Do not paste into chat logs if avoidable.');
}

function cmdHelp() {
  say(`Forma local development launcher

Usage:
  forma-local.ps1|forma-local.sh <command> [args]

Commands:
  doctor              Check toolchain / ports / files (no changes)
  start               Middleware → Forma migration → backend → frontend
  stop [--all]        Stop app processes; --all also compose down
  restart             Stop app + start again (no DB wipe)
  status              Show component health
  logs [backend|frontend|middleware|all]
  gen-secret          Generate FORMA_SECRET_MASTER_KEY value
  reset [--data YES]  Stop; optional destructive data wipe
  help                This message

Canonical URLs:
  Frontend  http://localhost:${FRONTEND_PORT}
  Backend   http://localhost:${BACKEND_PORT}

COZE_STUDIO_ROOT resolves from script location:
  ${COZE_STUDIO_ROOT}
`);
}

async function main() {
  try {
    switch (command) {
      case 'doctor':
        await cmdDoctor();
        break;
      case 'start':
        await cmdStart();
        break;
      case 'stop':
        await cmdStop();
        break;
      case 'restart':
        await cmdRestart();
        break;
      case 'status':
        await cmdStatus();
        break;
      case 'logs':
        await cmdLogs();
        break;
      case 'gen-secret':
        cmdGenSecret();
        break;
      case 'reset':
        await cmdReset();
        break;
      case 'help':
      case '--help':
      case '-h':
        cmdHelp();
        break;
      default:
        err(`Unknown command: ${command}`);
        cmdHelp();
        process.exitCode = 2;
    }
  } catch (e) {
    err(e && e.stack ? e.stack : String(e));
    process.exitCode = 1;
  }
}

// Exported for tests
export const __test = {
  COZE_STUDIO_ROOT,
  FRONTEND_PORT,
  BACKEND_PORT,
  redact,
  parseEnvFile,
  secretConfigured,
  portListening,
  pidAlive,
  parseGoVersion,
  parseNodeRange,
  nodeSatisfies,
  dockerPath,
  resolveRootFromScript: () => COZE_STUDIO_ROOT,
};

const isDirect =
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url);

if (isDirect) {
  await main();
}
