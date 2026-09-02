import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const root = join(repoRoot, 'frontend', 'apps', 'forma', 'src');

test('navigation defines 16 routes', () => {
  const navSrc = readFileSync(join(root, 'lib', 'navigation.ts'), 'utf8');
  const matches = navSrc.match(/path:\s*'\/[^']*'/g) ?? [];
  assert.ok(matches.length >= 16, `expected >=16 routes, got ${matches.length}`);
  assert.ok(navSrc.includes("path: '/'"));
  assert.ok(navSrc.includes("path: '/design'"));
  assert.ok(navSrc.includes("path: '/analyst'"));
  assert.ok(navSrc.includes("path: '/delivery'"));
});

test('design tokens css exists', () => {
  const tokens = readFileSync(join(root, 'styles', 'tokens.css'), 'utf8');
  assert.ok(tokens.includes('--forma-primary'));
  assert.ok(tokens.includes('--forma-background'));
});

test('data plane routes wired in AppRouter', () => {
  const routesSrc = readFileSync(join(root, 'routes', 'index.tsx'), 'utf8');
  assert.ok(routesSrc.includes('@forma/data'), 'expected @forma/data import');
  assert.ok(routesSrc.includes('/data/*') || routesSrc.includes("path=\"/data/*\""));
  for (const p of [
    '/data',
    '/data/requirements',
    '/data/sources',
    '/data/mappings',
    '/data/contracts',
    '/data/health',
  ]) {
    assert.ok(routesSrc.includes(p), `expected ${p} in routes/index.tsx`);
  }
});
