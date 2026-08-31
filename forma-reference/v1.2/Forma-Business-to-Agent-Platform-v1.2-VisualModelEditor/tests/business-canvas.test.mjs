import { readFile } from 'node:fs/promises';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import ts from 'typescript';
const code = ts.transpileModule(
  await readFile(new URL('../lib/business-canvas.ts', import.meta.url), 'utf8'),
  {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
  },
).outputText;
const {
  workOrderGraph,
  impactGraph,
  dataGraph,
  applicationGraph,
  layoutGraph,
} = await import(
  'data:text/javascript;base64,' + Buffer.from(code).toString('base64')
);
function valid(graph) {
  const ids = new Set(graph.nodes.map((n) => n.id));
  assert.equal(ids.size, graph.nodes.length);
  assert.equal(new Set(graph.edges.map((e) => e.id)).size, graph.edges.length);
  for (const e of graph.edges) {
    assert.ok(ids.has(e.source));
    assert.ok(ids.has(e.target));
    assert.ok(e.label);
  }
}
test('work order graph preserves candidate rule text and confirmation state', () => {
  for (const confirmed of [false, true]) {
    const graph = workOrderGraph('最新业务约束', confirmed);
    valid(graph);
    const rule = graph.nodes.find((n) => n.id === 'rule');
    assert.equal(rule.description, '最新业务约束');
    assert.ok(rule.label.includes(confirmed ? '已确认' : '候选'));
    for (const type of [
      'role',
      'entity',
      'process',
      'state',
      'rule',
      'external',
    ])
      assert.ok(graph.nodes.some((n) => n.type === type));
  }
});
test('data ownership follows External, Managed and Hybrid without dangling edges', () => {
  for (const mode of ['External', 'Managed', 'Hybrid']) {
    const graph = dataGraph(mode);
    valid(graph);
    assert.equal(
      graph.edges.find((e) => e.source === 'Asset').target,
      mode === 'Managed' ? 'managed' : 'external',
    );
    assert.equal(
      graph.edges.find((e) => e.source === 'WorkOrder').target,
      mode === 'External' ? 'external' : 'managed',
    );
  }
});
test('application topology distinguishes sequential and fan-out orchestration, including zero agents', () => {
  const agents = ['a', 'b', 'c'].map((id) => ({
    id,
    name: id,
    version: '1',
    capabilities: [],
  }));
  for (const mode of [
    'Pipeline',
    'Handoff',
    'Router',
    'Parallel',
    'Supervisor',
    'Human-in-the-loop',
  ]) {
    const graph = applicationGraph(agents, mode);
    valid(graph);
    assert.equal(
      graph.edges.some((e) => e.source === 'a' && e.target === 'b'),
      mode === 'Pipeline' || mode === 'Handoff',
    );
    valid(applicationGraph([], mode));
  }
});
test('automatic layout accepts missing coordinates, cycles, disconnected nodes and empty graphs', () => {
  assert.deepEqual(layoutGraph({ nodes: [], edges: [] }), []);
  const graph = impactGraph();
  valid(graph);
  const layout = layoutGraph(graph);
  assert.ok(layout[1].position.x > layout[0].position.x);
  graph.edges.push({
    id: 'cycle',
    source: 'app',
    target: 'rule',
    label: 'cycle',
  });
  for (const n of layoutGraph(graph))
    assert.ok(Number.isFinite(n.position.x) && Number.isFinite(n.position.y));
});
