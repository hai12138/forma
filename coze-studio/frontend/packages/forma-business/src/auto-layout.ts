import type { FormaSemanticModel, FormaViewLayout } from '@forma/api-client';

const COL_GAP = 290;
const ROW_GAP = 150;

export type LayoutGraphNode = { id: string };
export type LayoutGraphEdge = { source: string; target: string };
export type LayoutGraph = { nodes: LayoutGraphNode[]; edges: LayoutGraphEdge[] };

/**
 * Deterministic layered layout (v1.2 Golden Reference layoutGraph).
 * Same input → identical coordinates. Does not mutate semantic model.
 */
export function layoutGraph(graph: LayoutGraph): Array<{ id: string; position: { x: number; y: number } }> {
  const ids = new Set(graph.nodes.map(n => n.id));
  const depths = new Map<string, number>();
  const pending = new Set(ids);
  for (let pass = 0; pass < graph.nodes.length && pending.size; pass++) {
    for (const id of [...pending]) {
      const parents = graph.edges
        .filter(e => e.target === id && ids.has(e.source))
        .map(e => e.source);
      if (parents.every(p => depths.has(p))) {
        depths.set(id, Math.max(-1, ...parents.map(p => depths.get(p)!)) + 1);
        pending.delete(id);
      }
    }
  }
  const fallback = Math.max(0, ...depths.values(), -1) + 1;
  const rows = new Map<number, number>();
  // Stable order by input node order for determinism within a column.
  return graph.nodes.map(n => {
    const col = depths.get(n.id) ?? fallback;
    const row = rows.get(col) ?? 0;
    rows.set(col, row + 1);
    return { id: n.id, position: { x: col * COL_GAP, y: row * ROW_GAP } };
  });
}

/** Build auto layout ViewLayout from semantic model + current canvas items. */
export function computeAutoLayout(
  model: FormaSemanticModel,
  current: FormaViewLayout,
  canvasIds: string[],
): FormaViewLayout {
  const idSet = new Set(canvasIds);
  const nodes = canvasIds.map(id => ({ id }));
  const edges = model.edges
    .filter(e => idSet.has(e.source) && idSet.has(e.target))
    .map(e => ({ source: e.source, target: e.target }));
  const laid = layoutGraph({ nodes, edges });
  const node_positions: Record<string, { x: number; y: number }> = {
    ...current.node_positions,
  };
  for (const n of laid) {
    node_positions[n.id] = n.position;
  }
  return {
    ...current,
    mode: 'auto',
    node_positions,
  };
}
