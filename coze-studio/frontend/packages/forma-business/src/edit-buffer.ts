import type { FormaSemanticModel, FormaViewLayout } from '@forma/api-client';

export type EditSnapshot = {
  semantic_model: FormaSemanticModel;
  layout: FormaViewLayout;
};

export type EditBuffer = {
  baselineSemantic: FormaSemanticModel;
  baselineLayout: FormaViewLayout;
  current: EditSnapshot;
  past: EditSnapshot[];
  future: EditSnapshot[];
  modelRevision: number;
  layoutRevision: number;
};

const same = (a: unknown, b: unknown) => JSON.stringify(a) === JSON.stringify(b);

export function cloneSnapshot(s: EditSnapshot): EditSnapshot {
  return structuredClone(s);
}

export function createEditBuffer(input: {
  semantic_model: FormaSemanticModel;
  layout: FormaViewLayout;
  modelRevision: number;
  layoutRevision: number;
}): EditBuffer {
  const snap: EditSnapshot = {
    semantic_model: structuredClone(input.semantic_model),
    layout: structuredClone(input.layout),
  };
  return {
    baselineSemantic: structuredClone(input.semantic_model),
    baselineLayout: structuredClone(input.layout),
    current: snap,
    past: [],
    future: [],
    modelRevision: input.modelRevision,
    layoutRevision: input.layoutRevision,
  };
}

/** Semantic dirty: name/type/properties/edges/rules/states changed vs last Save Model baseline. */
export function isSemanticDirty(buffer: EditBuffer): boolean {
  return !same(buffer.current.semantic_model, buffer.baselineSemantic);
}

/** Layout-only dirty: positions/zoom/viewport differ; must NOT bump semantic revision. */
export function isLayoutDirty(buffer: EditBuffer): boolean {
  return !same(buffer.current.layout, buffer.baselineLayout);
}

export function pushSnapshot(buffer: EditBuffer, next: EditSnapshot): EditBuffer {
  if (same(buffer.current, next)) return buffer;
  return {
    ...buffer,
    past: [...buffer.past, cloneSnapshot(buffer.current)].slice(-100),
    future: [],
    current: cloneSnapshot(next),
  };
}

/** Layout-only edit — updates layout; still recorded in undo buffer. */
export function applyLayoutChange(
  buffer: EditBuffer,
  layout: FormaViewLayout,
): EditBuffer {
  return pushSnapshot(buffer, {
    semantic_model: buffer.current.semantic_model,
    layout,
  });
}

/** Semantic edit — marks MANUAL_MODIFIED on changed elements. */
export function applySemanticChange(
  buffer: EditBuffer,
  semantic_model: FormaSemanticModel,
  layout?: FormaViewLayout,
): EditBuffer {
  const marked = markManualModified(buffer.current.semantic_model, semantic_model);
  const nextLayout = layout ?? ensurePositions(marked, buffer.current.layout);
  return pushSnapshot(buffer, { semantic_model: marked, layout: nextLayout });
}

export function undo(buffer: EditBuffer): EditBuffer {
  const target = buffer.past.at(-1);
  if (!target) return buffer;
  return {
    ...buffer,
    past: buffer.past.slice(0, -1),
    future: [...buffer.future, cloneSnapshot(buffer.current)],
    current: cloneSnapshot(target),
  };
}

export function redo(buffer: EditBuffer): EditBuffer {
  const target = buffer.future.at(-1);
  if (!target) return buffer;
  return {
    ...buffer,
    future: buffer.future.slice(0, -1),
    past: [...buffer.past, cloneSnapshot(buffer.current)],
    current: cloneSnapshot(target),
  };
}

/** After successful Save Model — reset semantic baseline; keep layout baseline as-is. */
export function resetSemanticBaseline(
  buffer: EditBuffer,
  modelRevision: number,
  semantic_model?: FormaSemanticModel,
): EditBuffer {
  const model = semantic_model ?? buffer.current.semantic_model;
  return {
    ...buffer,
    modelRevision,
    baselineSemantic: structuredClone(model),
    current: {
      ...buffer.current,
      semantic_model: structuredClone(model),
    },
    past: [],
    future: [],
  };
}

/** After successful Save Layout — reset layout baseline. */
export function resetLayoutBaseline(
  buffer: EditBuffer,
  layoutRevision: number,
  layout?: FormaViewLayout,
): EditBuffer {
  const lay = layout ?? buffer.current.layout;
  return {
    ...buffer,
    layoutRevision,
    baselineLayout: structuredClone(lay),
    current: {
      ...buffer.current,
      layout: structuredClone(lay),
    },
  };
}

function markManualModified(
  prev: FormaSemanticModel,
  next: FormaSemanticModel,
): FormaSemanticModel {
  const nodes = next.nodes.map(n => {
    const old = prev.nodes.find(o => o.id === n.id);
    return same(n, old) ? n : { ...n, source_marker: 'MANUAL_MODIFIED' as const };
  });
  const edges = next.edges.map(e => {
    const old = prev.edges.find(o => o.id === e.id);
    return same(e, old) ? e : { ...e, source_marker: 'MANUAL_MODIFIED' as const };
  });
  const rules = next.rules.map(r => {
    const old = prev.rules.find(o => o.id === r.id);
    return same(r, old) ? r : { ...r, source_marker: 'MANUAL_MODIFIED' as const };
  });
  const states = next.states.map(s => {
    const old = prev.states.find(o => o.id === s.id);
    return same(s, old) ? s : { ...s, source_marker: 'MANUAL_MODIFIED' as const };
  });
  return { ...next, nodes, edges, rules, states };
}

function ensurePositions(
  model: FormaSemanticModel,
  layout: FormaViewLayout,
): FormaViewLayout {
  const positions = { ...layout.node_positions };
  let i = 0;
  const ids = [
    ...model.nodes.map(n => n.id),
    ...model.states.map(s => s.id),
    ...model.rules.map(r => r.id),
  ];
  for (const id of ids) {
    if (!positions[id]) {
      positions[id] = { x: 40 + (i % 4) * 220, y: 40 + Math.floor(i / 4) * 120 };
    }
    i += 1;
  }
  return { ...layout, node_positions: positions };
}

export type CanvasItem =
  | { kind: 'node'; id: string; type: string; name: string; description?: string; source_marker: string }
  | { kind: 'state'; id: string; type: 'STATE'; name: string; description?: string; source_marker: string; object_ref: string }
  | { kind: 'rule'; id: string; type: 'RULE'; name: string; description?: string; source_marker: string };

export function collectCanvasItems(model: FormaSemanticModel): CanvasItem[] {
  const nodes: CanvasItem[] = model.nodes.map(n => ({
    kind: 'node' as const,
    id: n.id,
    type: n.type,
    name: n.name,
    description: n.description,
    source_marker: n.source_marker,
  }));
  const states: CanvasItem[] = model.states.map(s => ({
    kind: 'state' as const,
    id: s.id,
    type: 'STATE' as const,
    name: s.name,
    description: s.description,
    source_marker: s.source_marker,
    object_ref: s.object_ref,
  }));
  const rules: CanvasItem[] = model.rules.map(r => ({
    kind: 'rule' as const,
    id: r.id,
    type: 'RULE' as const,
    name: r.name,
    description: r.description,
    source_marker: r.source_marker,
  }));
  return [...nodes, ...states, ...rules];
}

/** Rule is NOT a relationship endpoint — only node/state are. */
export function isEdgeEndpoint(model: FormaSemanticModel, id: string): boolean {
  return model.nodes.some(n => n.id === id) || model.states.some(s => s.id === id);
}

export type NodeDeleteImpact = {
  edgeCount: number;
  stateCount: number;
  ruleRefCount: number;
  dependentStateIds: string[];
  dependentEdgeIds: string[];
  ruleIdsWithRef: string[];
};

export function analyzeNodeDeleteImpact(
  model: FormaSemanticModel,
  nodeId: string,
): NodeDeleteImpact {
  const dependentStateIds = model.states
    .filter(s => s.object_ref === nodeId)
    .map(s => s.id);
  const stateSet = new Set(dependentStateIds);
  const dependentEdgeIds = model.edges
    .filter(
      e =>
        e.source === nodeId ||
        e.target === nodeId ||
        stateSet.has(e.source) ||
        stateSet.has(e.target),
    )
    .map(e => e.id);
  const ruleIdsWithRef = model.rules
    .filter(r => (r.applies_to ?? []).includes(nodeId))
    .map(r => r.id);
  return {
    edgeCount: dependentEdgeIds.length,
    stateCount: dependentStateIds.length,
    ruleRefCount: ruleIdsWithRef.length,
    dependentStateIds,
    dependentEdgeIds,
    ruleIdsWithRef,
  };
}

/**
 * Cascade delete node:
 * - remove direct + dependent-state edges
 * - delete states with object_ref == node
 * - strip node from rule.applies_to (keep rule)
 */
export function deleteNodeWithDependencies(
  model: FormaSemanticModel,
  nodeId: string,
): FormaSemanticModel {
  const impact = analyzeNodeDeleteImpact(model, nodeId);
  const dropStates = new Set(impact.dependentStateIds);
  const dropEdges = new Set(impact.dependentEdgeIds);
  return {
    ...model,
    nodes: model.nodes.filter(n => n.id !== nodeId),
    states: model.states.filter(s => !dropStates.has(s.id)),
    edges: model.edges.filter(e => !dropEdges.has(e.id)),
    rules: model.rules.map(r => ({
      ...r,
      applies_to: (r.applies_to ?? []).filter(ref => ref !== nodeId),
    })),
  };
}
