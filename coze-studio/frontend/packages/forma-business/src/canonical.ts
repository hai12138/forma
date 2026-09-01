/** Canonical NodeTypes for Business Model SoT. */
export const CANONICAL_NODE_TYPES = [
  'ACTOR',
  'BUSINESS_OBJECT',
  'PROCESS',
  'EVENT',
  'DECISION',
  'SYSTEM',
  'POLICY',
] as const;

export type CanonicalNodeType = (typeof CANONICAL_NODE_TYPES)[number];

/** Formal relationship types (UI dropdown + backend contract). */
export const EDGE_TYPES = [
  'PERFORMS',
  'CREATES',
  'UPDATES',
  'TRIGGERS',
  'REQUIRES',
  'DEPENDS_ON',
  'TRANSITIONS_TO',
  'RELATES_TO',
] as const;

export type FormalEdgeType = (typeof EDGE_TYPES)[number];

/** v1.2 import / FE compatibility aliases → canonical. */
const ALIAS_TO_CANONICAL: Record<string, CanonicalNodeType> = {
  role: 'ACTOR',
  entity: 'BUSINESS_OBJECT',
  process: 'PROCESS',
  external: 'SYSTEM',
};

const REJECTED_NODE_TYPES = new Set(['agent', 'application', 'state', 'rule']);

export function canonicalizeNodeType(type: string): CanonicalNodeType | null {
  if ((CANONICAL_NODE_TYPES as readonly string[]).includes(type)) {
    return type as CanonicalNodeType;
  }
  if (ALIAS_TO_CANONICAL[type]) return ALIAS_TO_CANONICAL[type];
  return null;
}

export function isRejectedNodeType(type: string): boolean {
  return REJECTED_NODE_TYPES.has(type);
}

/** Convert alias types in a model before persistence. Rejects agent/application/state/rule. */
export function adaptModelForPersistence<T extends {
  nodes: Array<{ type: string }>;
}>(model: T): T {
  return {
    ...model,
    nodes: model.nodes.map(n => {
      if (isRejectedNodeType(n.type)) {
        throw new Error(`NodeType ${n.type} is not allowed in Business Model SoT`);
      }
      const canon = canonicalizeNodeType(n.type);
      if (!canon) {
        throw new Error(`Unsupported NodeType ${n.type}`);
      }
      return { ...n, type: canon };
    }),
  };
}

export const SOURCE_MARKERS = ['AI_GENERATED', 'MANUAL_MODIFIED'] as const;
