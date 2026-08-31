import { readFile } from 'node:fs/promises';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import ts from 'typescript';
const source = await readFile(
  new URL('../lib/domain.ts', import.meta.url),
  'utf8',
);
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2022,
  },
}).outputText;
const { initialState, gateReasons, invalidate, validateImport } = await import(
  'data:text/javascript;base64,' + Buffer.from(compiled).toString('base64')
);
const fixture = () => structuredClone(initialState);
function verified() {
  const s = fixture();
  s.approvals = s.approvals.map((a) => ({ ...a, confirmed: true }));
  s.evaluation = {
    revision: s.revision,
    passed: true,
    total: 24,
    failed: 0,
    at: '2026-08-31',
  };
  return s;
}
test('unconfirmed business rules block release even if regression passed', () => {
  const s = verified();
  s.approvals[1].confirmed = false;
  assert.ok(gateReasons(s).some((r) => r.includes('双人确认')));
});
test('all confirmed and current regression passes gate', () =>
  assert.deepEqual(gateReasons(verified()), []));
test('stale regression cannot authorize new asset snapshot', () => {
  const s = verified();
  s.revision++;
  assert.ok(gateReasons(s).some((r) => r.includes('回归')));
});
test('failed regression blocks release', () => {
  const s = verified();
  s.evaluation.passed = false;
  assert.ok(gateReasons(s).length > 0);
});
test('configuration edits clear freeze and regression but preserve audit history', () => {
  const s = verified();
  s.frozen = s.revision;
  s.stage = 'Prod';
  s.released = true;
  const next = invalidate(s, { runtime: 'Eino' });
  assert.equal(next.revision, s.revision + 1);
  assert.equal(next.evaluation, null);
  assert.equal(next.frozen, null);
  assert.equal(next.released, false);
  assert.equal(next.stage, 'Dev');
  assert.deepEqual(next.audit, s.audit);
  assert.equal(s.runtime, 'LangGraph');
});
test('empty application cannot pass gate', () => {
  const s = verified();
  s.application.selected = [];
  assert.ok(gateReasons(s).some((r) => r.includes('至少')));
});
test('dangling application dependency blocks gate', () => {
  const s = verified();
  s.application.selected = ['missing'];
  assert.ok(gateReasons(s).some((r) => r.includes('失效')));
});
test('soft deleted application dependency blocks gate', () => {
  const s = verified();
  s.agents[0].deleted = true;
  assert.ok(gateReasons(s).some((r) => r.includes('失效')));
});
test('import creates fresh unprivileged draft IDs', () => {
  const original = fixture().agents[0];
  const [a] = validateImport(original);
  assert.notEqual(a.id, original.id);
  assert.equal(a.status, '草稿');
  assert.equal(a.version, '0.1.0');
  assert.equal(a.permission, 'tenant:current · 最小权限');
  assert.equal(a.deleted, false);
});
test('unknown capability import rejected', () =>
  assert.throws(() =>
    validateImport({
      name: 'X',
      role: 'Y',
      capabilities: ['dangerous.unknown'],
    }),
  ));
test('missing name import rejected', () =>
  assert.throws(() =>
    validateImport({ name: ' ', role: 'Y', capabilities: [] }),
  ));
test('oversized agent collections rejected', () =>
  assert.throws(() => validateImport(Array(101).fill(fixture().agents[0]))));
