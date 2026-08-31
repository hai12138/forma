import {
  layoutGraph,
  workOrderGraph,
  type BusinessGraph,
  type BusinessNode,
  type BusinessEdge,
} from './business-canvas';
import type { State } from './domain';
export type Origin = 'AI_generated' | 'manual_modified';
export type SemanticNode = Omit<BusinessNode, 'position'> & { source: Origin };
// Legacy canvas edge.source is its endpoint; semantic edge.source is provenance.
export type SemanticEdge = Omit<BusinessEdge, 'source'> & {
  from: string;
  source: Origin;
};
export type SemanticModel = {
  nodes: SemanticNode[];
  edges: SemanticEdge[];
  rules: string[];
  states: string[];
};
export type ViewLayout = {
  node_positions: Record<string, { x: number; y: number }>;
  zoom: number;
  viewport: { x: number; y: number };
  mode: 'auto' | 'manual';
  groups: string[][];
};
export type Snapshot = {
  semantic_model: SemanticModel;
  view_layout: ViewLayout;
};
export type VisualModel = Snapshot & {
  revision: number;
  past: Snapshot[];
  future: Snapshot[];
  revisions: {
    revision: number;
    action: string;
    semantic_model: SemanticModel;
  }[];
  impact: { revision: number; status: 'pending'; changed: string[] } | null;
  saved_layout?: ViewLayout;
};
const same = (a: unknown, b: unknown) =>
  JSON.stringify(a) === JSON.stringify(b);
export const semanticChanged = (a: Snapshot, b: Snapshot) =>
  !same(a.semantic_model, b.semantic_model);
export function createVisualModel(rule: string): VisualModel {
  const graph = workOrderGraph(rule, false);
  return {
    semantic_model: {
      nodes: graph.nodes.map(({ position: _position, ...n }) => ({
        ...n,
        source: 'AI_generated',
      })),
      edges: graph.edges.map(({ source, ...e }) => ({
        ...e,
        from: source,
        source: 'AI_generated',
      })),
      rules: graph.nodes.filter((n) => n.type === 'rule').map((n) => n.id),
      states: graph.nodes.filter((n) => n.type === 'state').map((n) => n.id),
    },
    view_layout: {
      node_positions: Object.fromEntries(
        layoutGraph(graph).map((n) => [n.id, n.position]),
      ),
      zoom: 0.65,
      viewport: { x: 65, y: 145 },
      mode: 'manual',
      groups: [],
    },
    revision: 1,
    past: [],
    future: [],
    revisions: [],
    impact: null,
  };
}
export function canvasGraph(model: Snapshot): BusinessGraph {
  return {
    nodes: model.semantic_model.nodes.map((n) => ({
      ...n,
      position: model.view_layout.node_positions[n.id],
    })),
    edges: model.semantic_model.edges.map(
      ({ source: _origin, from, ...e }) => ({ ...e, source: from }),
    ),
  };
}
function snapshot(m: Snapshot): Snapshot {
  return structuredClone({
    semantic_model: m.semantic_model,
    view_layout: m.view_layout,
  });
}
export function validateSemantic(s: SemanticModel) {
  const ids = new Set(s.nodes.map((n) => n.id));
  if (ids.size !== s.nodes.length || s.nodes.some((n) => !n.label.trim()))
    throw Error('节点名称不能为空，ID 必须唯一');
  if (
    new Set(s.edges.map((e) => e.id)).size !== s.edges.length ||
    s.edges.some(
      (e) =>
        !ids.has(e.from) ||
        !ids.has(e.target) ||
        !e.label.trim() ||
        e.from === e.target,
    )
  )
    throw Error('关系需要有效的两个不同节点与非空标签');
  if (
    s.rules.some(
      (id) => !s.nodes.some((n) => n.id === id && n.type === 'rule'),
    ) ||
    s.states.some(
      (id) => !s.nodes.some((n) => n.id === id && n.type === 'state'),
    )
  )
    throw Error('规则与状态引用无效');
}
function apply(m: VisualModel, next: Snapshot, action: string): VisualModel {
  validateSemantic(next.semantic_model);
  const changed = semanticChanged(m, next);
  const revision = m.revision + (changed ? 1 : 0);
  const before = [...m.semantic_model.nodes, ...m.semantic_model.edges];
  const after = [...next.semantic_model.nodes, ...next.semantic_model.edges];
  const changedIds = [
    ...new Set([...before, ...after].map((n) => n.id)),
  ].filter(
    (id) =>
      !same(
        before.find((n) => n.id === id),
        after.find((n) => n.id === id),
      ),
  );
  return {
    ...m,
    ...snapshot(next),
    revision,
    revisions: changed
      ? [
          ...m.revisions,
          {
            revision,
            action,
            semantic_model: structuredClone(next.semantic_model),
          },
        ]
      : m.revisions,
    impact: changed
      ? {
          revision,
          status: 'pending',
          changed: changedIds.length ? changedIds : ['rules/states'],
        }
      : m.impact,
  };
}
export function commitModel(
  m: VisualModel,
  next: Snapshot,
  action: string,
): VisualModel {
  if (same(snapshot(m), snapshot(next))) return m;
  return {
    ...apply(m, next, action),
    past: [...m.past, snapshot(m)].slice(-100),
    future: [],
  };
}
export function historyModel(
  m: VisualModel,
  direction: 'undo' | 'redo',
): VisualModel {
  const items = direction === 'undo' ? m.past : m.future;
  const target = items.at(-1);
  if (!target) return m;
  return {
    ...apply(m, target, direction === 'undo' ? '撤销语义修改' : '重做语义修改'),
    past: direction === 'undo' ? m.past.slice(0, -1) : [...m.past, snapshot(m)],
    future:
      direction === 'undo' ? [...m.future, snapshot(m)] : m.future.slice(0, -1),
  };
}
export function editSemantic(m: VisualModel, s: SemanticModel, action: string) {
  const nodes = s.nodes.map((n) =>
    same(
      n,
      m.semantic_model.nodes.find((old) => old.id === n.id),
    )
      ? n
      : { ...n, source: 'manual_modified' as const },
  );
  const edges = s.edges.map((e) =>
    same(
      e,
      m.semantic_model.edges.find((old) => old.id === e.id),
    )
      ? e
      : { ...e, source: 'manual_modified' as const },
  );
  const semantic_model = {
    ...s,
    nodes,
    edges,
    rules: nodes.filter((n) => n.type === 'rule').map((n) => n.id),
    states: nodes.filter((n) => n.type === 'state').map((n) => n.id),
  };
  const node_positions = Object.fromEntries(
    nodes.map((n, i) => [
      n.id,
      m.view_layout.node_positions[n.id] ?? { x: i * 40, y: i * 40 },
    ]),
  );
  return commitModel(
    m,
    { semantic_model, view_layout: { ...m.view_layout, node_positions } },
    action,
  );
}
export function deleteNode(m: VisualModel, id: string) {
  return editSemantic(
    m,
    {
      ...m.semantic_model,
      nodes: m.semantic_model.nodes.filter((n) => n.id !== id),
      edges: m.semantic_model.edges.filter(
        (e) => e.from !== id && e.target !== id,
      ),
    },
    '删除节点及关联关系',
  );
}
// Layout-only API accepts no semantic payload, including simulated AI layout.
export function arrangeModel(m: VisualModel): VisualModel {
  const graph = canvasGraph(m);
  const node_positions = Object.fromEntries(
    layoutGraph({
      ...graph,
      nodes: graph.nodes.map(({ position: _p, ...n }) => n),
    }).map((n) => [n.id, n.position]),
  );
  return commitModel(
    m,
    { ...m, view_layout: { ...m.view_layout, node_positions, mode: 'auto' } },
    '自动布局',
  );
}
export function applyVisualModel(s: State, model: VisualModel): State {
  const previous = s.visualModel ?? createVisualModel(s.rule);
  if (!semanticChanged(previous, model)) return { ...s, visualModel: model };
  return {
    ...s,
    visualModel: model,
    rule:
      model.semantic_model.nodes.find((n) => n.id === 'rule')?.description ??
      '',
    approvals: s.approvals.map((a) => ({ ...a, confirmed: false })),
    revision: s.revision + 1,
    evaluation: null,
    frozen: null,
    stage: 'Dev',
    released: false,
  };
}
