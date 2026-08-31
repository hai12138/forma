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
