import type { FormaSemanticModel } from '@forma/api-client';

import {
  CANONICAL_NODE_TYPES,
  EDGE_TYPES,
  SOURCE_MARKERS,
  canonicalizeNodeType,
  isRejectedNodeType,
} from './canonical';

export type SemanticValidationIssue = {
  code: string;
  message: string;
  elementId?: string;
};

export type SemanticValidationResult = {
  ok: boolean;
  issues: SemanticValidationIssue[];
};

const SOURCE_OK = new Set<string>(SOURCE_MARKERS);
const EDGE_OK = new Set<string>(EDGE_TYPES);

/**
 * Lightweight editor preflight — mirrors editor-editable invariants.
 * Backend ValidateSemanticModel remains authoritative.
 */
export function validateEditorSemanticModel(
  model: FormaSemanticModel | null | undefined,
): SemanticValidationResult {
  const issues: SemanticValidationIssue[] = [];
  if (!model) {
    return { ok: false, issues: [{ code: 'MODEL_NIL', message: 'semantic model is required' }] };
  }

  const nodes = model.nodes ?? [];
  const edges = model.edges ?? [];
  const rules = model.rules ?? [];
  const states = model.states ?? [];

  const globalIDs = new Map<string, string>(); // id → kind
  const claim = (id: string, kind: string) => {
    const trimmed = id.trim();
    if (!trimmed) {
      issues.push({ code: 'ID_EMPTY', message: `${kind} id empty` });
      return false;
    }
    const prev = globalIDs.get(trimmed);
    if (prev) {
      issues.push({
        code: 'ID_COLLISION',
        message: `${kind} id ${trimmed} collides with ${prev}`,
        elementId: trimmed,
      });
      return false;
    }
    globalIDs.set(trimmed, kind);
    return true;
  };

  const nodeIDs = new Set<string>();
  for (const n of nodes) {
    if (!claim(n.id, 'node')) continue;
    nodeIDs.add(n.id);
    if (!n.name?.trim()) {
      issues.push({
        code: 'NODE_NAME_EMPTY',
        message: `节点名称不能为空 (${n.id})`,
        elementId: n.id,
      });
    }
    if (isRejectedNodeType(n.type)) {
      issues.push({
        code: 'NODE_TYPE_REJECTED',
        message: `NodeType ${n.type} 不允许`,
        elementId: n.id,
      });
    } else if (!canonicalizeNodeType(n.type) || !(CANONICAL_NODE_TYPES as readonly string[]).includes(canonicalizeNodeType(n.type)!)) {
      // After canonicalize, must be canonical (editor should only emit canonical).
      const canon = canonicalizeNodeType(n.type);
      if (!canon) {
        issues.push({
          code: 'NODE_TYPE_INVALID',
          message: `不支持的 NodeType ${n.type}`,
          elementId: n.id,
        });
      }
    }
    if (n.source_marker && !SOURCE_OK.has(n.source_marker)) {
      issues.push({
        code: 'SOURCE_MARKER_INVALID',
        message: `非法 source_marker ${n.source_marker}`,
        elementId: n.id,
      });
    }
  }

  const stateIDs = new Set<string>();
  for (const s of states) {
    if (!claim(s.id, 'state')) continue;
    stateIDs.add(s.id);
    if (!s.name?.trim()) {
      issues.push({
        code: 'STATE_NAME_EMPTY',
        message: `状态名称不能为空 (${s.id})`,
        elementId: s.id,
      });
    }
    if (!s.object_ref || !nodeIDs.has(s.object_ref)) {
      issues.push({
        code: 'STATE_OBJECT_REF_INVALID',
        message: `状态 ${s.id} 的 object_ref 无效`,
        elementId: s.id,
      });
    }
    if (s.source_marker && !SOURCE_OK.has(s.source_marker)) {
      issues.push({
        code: 'SOURCE_MARKER_INVALID',
        message: `非法 source_marker ${s.source_marker}`,
        elementId: s.id,
      });
    }
  }

  const endpointOK = (id: string) => nodeIDs.has(id) || stateIDs.has(id);

  for (const e of edges) {
    if (!claim(e.id, 'edge')) continue;
    if (!endpointOK(e.source) || !endpointOK(e.target)) {
      issues.push({
        code: 'EDGE_ENDPOINT_INVALID',
        message: `关系 ${e.id} 端点无效`,
        elementId: e.id,
      });
    }
    if (stateIDs.has(e.source) === false && nodeIDs.has(e.source) === false) {
      /* already covered */
    }
    // Rule must never be edge endpoint — rules are not in endpointOK.
    if (!e.label?.trim()) {
      issues.push({
        code: 'EDGE_LABEL_EMPTY',
        message: `关系标签不能为空 (${e.id})`,
        elementId: e.id,
      });
    }
    if (e.type && !EDGE_OK.has(e.type)) {
      issues.push({
        code: 'EDGE_TYPE_INVALID',
        message: `不支持的关系类型 ${e.type}`,
        elementId: e.id,
      });
    }
    if (e.source_marker && !SOURCE_OK.has(e.source_marker)) {
      issues.push({
        code: 'SOURCE_MARKER_INVALID',
        message: `非法 source_marker ${e.source_marker}`,
        elementId: e.id,
      });
    }
  }

  for (const r of rules) {
    if (!claim(r.id, 'rule')) continue;
    if (!r.name?.trim()) {
      issues.push({
        code: 'RULE_NAME_EMPTY',
        message: `规则名称不能为空 (${r.id})`,
        elementId: r.id,
      });
    }
    for (const ref of r.applies_to ?? []) {
      if (ref && !endpointOK(ref)) {
        issues.push({
          code: 'RULE_APPLIES_TO_INVALID',
          message: `规则 ${r.id} applies_to 引用无效: ${ref}`,
          elementId: r.id,
        });
      }
    }
    if (r.source_marker && !SOURCE_OK.has(r.source_marker)) {
      issues.push({
        code: 'SOURCE_MARKER_INVALID',
        message: `非法 source_marker ${r.source_marker}`,
        elementId: r.id,
      });
    }
  }

  return { ok: issues.length === 0, issues };
}

export function formatValidationIssues(result: SemanticValidationResult): string {
  if (result.ok) return '';
  return result.issues.map(i => i.message).join('；');
}
