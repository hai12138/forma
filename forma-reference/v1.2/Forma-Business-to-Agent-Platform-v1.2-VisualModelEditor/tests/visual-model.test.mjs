import { readFile } from 'node:fs/promises';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import ts from 'typescript';
const compile = async (name) =>
  ts.transpileModule(
    await readFile(new URL(`../lib/${name}.ts`, import.meta.url), 'utf8'),
    {
      compilerOptions: {
        module: ts.ModuleKind.ESNext,
        target: ts.ScriptTarget.ES2022,
      },
    },
  ).outputText;
const url = (code) =>
  'data:text/javascript;base64,' + Buffer.from(code).toString('base64');
const canvas = url(await compile('business-canvas'));
const api = await import(
  url(
    (await compile('visual-model')).replace(
      "'./business-canvas'",
      JSON.stringify(canvas),
    ),
  )
);
const { initialState } = await import(url(await compile('domain')));
const {
  createVisualModel,
  editSemantic,
  commitModel,
  historyModel,
  arrangeModel,
  deleteNode,
  applyVisualModel,
  canvasGraph,
  validateSemantic,
} = api;
const seed = () => createVisualModel('紧急自动派单');
const approved = (m) => ({
  ...structuredClone(initialState),
  visualModel: m,
  approvals: initialState.approvals.map((a) => ({ ...a, confirmed: true })),
  evaluation: { revision: 1, passed: true },
  frozen: 1,
  stage: 'Prod',
  released: true,
});
const rename = (m) =>
  editSemantic(
    m,
    {
      ...m.semantic_model,
      nodes: m.semantic_model.nodes.map((n) =>
        n.id === 'order' ? { ...n, label: '服务工单' } : n,
      ),
    },
    '重命名',
  );
test('legacy data migrates without coordinates in semantic nodes; edge provenance does not overwrite endpoints', () => {
  const m = seed();
  assert.ok(
    m.semantic_model.nodes.every(
      (n) => !('position' in n) && n.source === 'AI_generated',
    ),
  );
  assert.equal(m.semantic_model.edges[0].source, 'AI_generated');
  assert.equal(canvasGraph(m).edges[0].source, 'reporter');
});
test('drag, viewport, zoom, save and AI layout preserve every governance credential', () => {
  const m = seed(),
    s = approved(m);
  const moved = commitModel(
    m,
    {
      ...m,
      view_layout: {
        ...m.view_layout,
        zoom: 1.1,
        viewport: { x: 40, y: 60 },
        node_positions: {
          ...m.view_layout.node_positions,
          order: { x: 99, y: 88 },
        },
      },
    },
    '移动',
  );
  for (const next of [
    moved,
    arrangeModel(moved),
    { ...moved, saved_layout: moved.view_layout },
  ]) {
    const result = applyVisualModel(s, next);
    assert.equal(result.revision, s.revision);
    assert.deepEqual(result.evaluation, s.evaluation);
    assert.equal(result.frozen, 1);
    assert.equal(result.released, true);
    assert.deepEqual(result.approvals, s.approvals);
    assert.deepEqual(next.semantic_model, m.semantic_model);
    assert.equal(next.revision, m.revision);
  }
});
test('semantic change creates revision, snapshot, impact and invalidates all candidate credentials', () => {
  const m = seed(),
    next = rename(m),
    s = applyVisualModel(approved(m), next);
  assert.equal(next.revision, 2);
  assert.equal(next.revisions.length, 1);
  assert.ok(next.impact.changed.includes('order'));
  assert.equal(s.revision, 2);
  assert.equal(s.evaluation, null);
  assert.equal(s.frozen, null);
  assert.ok(s.approvals.every((a) => !a.confirmed));
  assert.equal(s.stage, 'Dev');
  assert.equal(s.released, false);
  assert.equal(
    next.semantic_model.nodes.find((n) => n.id === 'order').source,
    'manual_modified',
  );
});
test('all six node types can be added, retyped and removed with rule/state indexes synchronized', () => {
  for (const type of [
    'role',
    'entity',
    'process',
    'state',
    'rule',
    'external',
  ]) {
    let m = seed();
    m = editSemantic(
      m,
      {
        ...m.semantic_model,
        nodes: [
          ...m.semantic_model.nodes,
          { id: 'new', label: '新增', type, source: 'manual_modified' },
        ],
      },
      '新增',
    );
    assert.ok(m.semantic_model.nodes.some((n) => n.id === 'new'));
    if (type === 'rule') assert.ok(m.semantic_model.rules.includes('new'));
    if (type === 'state') assert.ok(m.semantic_model.states.includes('new'));
    m = editSemantic(
      m,
      {
        ...m.semantic_model,
        nodes: m.semantic_model.nodes.map((n) =>
          n.id === 'new' ? { ...n, type: 'entity' } : n,
        ),
      },
      '改类型',
    );
    assert.ok(!m.semantic_model.rules.includes('new'));
    assert.ok(!m.semantic_model.states.includes('new'));
    m = deleteNode(m, 'new');
    assert.ok(!m.semantic_model.nodes.some((n) => n.id === 'new'));
    assert.ok(!m.view_layout.node_positions.new);
  }
});
test('node deletion removes incident edges; undo restores them without restoring old approvals', () => {
  const m = seed(),
    deleted = deleteNode(m, 'order');
  assert.ok(
    deleted.semantic_model.edges.every(
      (e) => e.from !== 'order' && e.target !== 'order',
    ),
  );
  const undo = historyModel(deleted, 'undo');
  assert.deepEqual(undo.semantic_model, m.semantic_model);
  assert.equal(undo.revision, 3);
  const state = applyVisualModel(approved(deleted), undo);
  assert.equal(state.frozen, null);
  assert.ok(state.approvals.every((a) => !a.confirmed));
});
test('relationship create, label edit, delete and undo are semantic revisions', () => {
  let m = seed();
  m = editSemantic(
    m,
    {
      ...m.semantic_model,
      edges: [
        ...m.semantic_model.edges,
        {
          id: 'new-edge',
          from: 'reporter',
          target: 'engineer',
          label: '联系',
          source: 'manual_modified',
        },
      ],
    },
    '连线',
  );
  m = editSemantic(
    m,
    {
      ...m.semantic_model,
      edges: m.semantic_model.edges.map((e) =>
        e.id === 'new-edge' ? { ...e, label: '咨询' } : e,
      ),
    },
    '语义',
  );
  assert.equal(m.revision, 3);
  m = editSemantic(
    m,
    {
      ...m.semantic_model,
      edges: m.semantic_model.edges.filter((e) => e.id !== 'new-edge'),
    },
    '删除',
  );
  assert.equal(m.revision, 4);
  assert.ok(
    historyModel(m, 'undo').semantic_model.edges.some(
      (e) => e.label === '咨询',
    ),
  );
});
test('layout undo/redo leaves semantic revisions unchanged; semantic undo/redo is monotonic', () => {
  const m = seed(),
    auto = arrangeModel(m);
  assert.deepEqual(historyModel(auto, 'undo').view_layout, m.view_layout);
  assert.equal(historyModel(auto, 'undo').revision, 1);
  const edited = rename(m),
    undone = historyModel(edited, 'undo'),
    redone = historyModel(undone, 'redo');
  assert.equal(redone.revision, 4);
  assert.deepEqual(redone.semantic_model, edited.semantic_model);
  assert.equal(redone.revisions.length, 3);
});
test('new edit after undo clears redo; no-op does not create a revision or history', () => {
  const m = seed();
  assert.equal(commitModel(m, m, 'noop'), m);
  assert.equal(editSemantic(m, m.semantic_model, 'noop'), m);
  const undone = historyModel(rename(m), 'undo');
  assert.equal(arrangeModel(undone).future.length, 0);
});
test('invalid dangling, self and blank relationships are rejected atomically', () => {
  for (const edge of [
    { from: 'missing', target: 'order', label: 'test' },
    { from: 'order', target: 'order', label: 'test' },
    { from: 'reporter', target: 'order', label: ' ' },
  ]) {
    const m = seed();
    assert.throws(() =>
      validateSemantic({
        ...m.semantic_model,
        edges: [
          ...m.semantic_model.edges,
          { id: 'bad', source: 'manual_modified', ...edge },
        ],
      }),
    );
    assert.equal(m.revision, 1);
  }
});
test('persisted model roundtrip preserves edits, layout, history, source and saved checkpoint', () => {
  const m = rename(seed());
  m.saved_layout = structuredClone(m.view_layout);
  const restored = JSON.parse(JSON.stringify(m));
  assert.deepEqual(restored, m);
  assert.deepEqual(
    historyModel(restored, 'undo').semantic_model,
    seed().semantic_model,
  );
});
