import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const appRoot = join(repoRoot, 'frontend', 'apps', 'forma', 'src');
const dataApp = join(
  repoRoot,
  'frontend',
  'packages',
  'forma-data',
  'src',
  'pages',
  'DataPlaneApp.tsx',
);

test('navigation defines 16 routes', () => {
  const navSrc = readFileSync(join(appRoot, 'lib', 'navigation.ts'), 'utf8');
  const matches = navSrc.match(/path:\s*'\/[^']*'/g) ?? [];
  assert.ok(matches.length >= 16, `expected >=16 routes, got ${matches.length}`);
  assert.ok(navSrc.includes("path: '/'"));
  assert.ok(navSrc.includes("path: '/design'"));
  assert.ok(navSrc.includes("path: '/analyst'"));
  assert.ok(navSrc.includes("path: '/business'"));
  assert.ok(navSrc.includes("path: '/data'"));
  assert.ok(navSrc.includes("path: '/delivery'"));
});

test('design tokens css exists', () => {
  const tokens = readFileSync(join(appRoot, 'styles', 'tokens.css'), 'utf8');
  assert.ok(tokens.includes('--forma-primary'));
  assert.ok(tokens.includes('--forma-background'));
});

test('AppRouter mounts DataPlaneApp at /data/* and keeps business/analyst', () => {
  const routesSrc = readFileSync(join(appRoot, 'routes', 'index.tsx'), 'utf8');
  assert.ok(routesSrc.includes('@forma/data'), 'expected @forma/data import');
  assert.ok(routesSrc.includes('path="/data/*"') || routesSrc.includes("path='/data/*'"));
  assert.ok(routesSrc.includes('path="/business"'));
  assert.ok(routesSrc.includes('path="/analyst"'));
  assert.ok(!routesSrc.includes('PlaceholderPage title="数据'));
});

test('DataPlaneApp declares nested product routes', () => {
  const src = readFileSync(dataApp, 'utf8');
  assert.ok(src.includes('DataOverviewPage'));
  assert.ok(src.includes('path="requirements"'));
  assert.ok(src.includes('path="sources"'));
  assert.ok(src.includes('path="sources/:sourceId"'));
  assert.ok(src.includes('path="mappings"'));
  assert.ok(src.includes('path="contracts"'));
  assert.ok(src.includes('path="contracts/:contractId"'));
  assert.ok(src.includes('path="health"'));
});
