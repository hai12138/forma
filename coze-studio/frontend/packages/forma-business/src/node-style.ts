/** Semantic colors for formal node types (+ STATE/RULE visuals). */

export type VisualNodeKind =
  | 'ACTOR'
  | 'BUSINESS_OBJECT'
  | 'PROCESS'
  | 'EVENT'
  | 'DECISION'
  | 'SYSTEM'
  | 'POLICY'
  | 'STATE'
  | 'RULE'
  | string;

export const NODE_STYLES: Record<
  string,
  { label: string; color: string; background: string }
> = {
  ACTOR: { label: '参与者', color: '#2563eb', background: '#f0f5ff' },
  BUSINESS_OBJECT: { label: '业务对象', color: '#c47712', background: '#fff8eb' },
  PROCESS: { label: '流程', color: '#db477e', background: '#fff2f7' },
  EVENT: { label: '事件', color: '#0891a8', background: '#effbfd' },
  DECISION: { label: '决策', color: '#bc8a38', background: '#fff8eb' },
  SYSTEM: { label: '系统', color: '#596579', background: '#f3f5f8' },
  POLICY: { label: '策略', color: '#8a55d6', background: '#f7f2ff' },
  STATE: { label: '状态', color: '#0891a8', background: '#effbfd' },
  RULE: { label: '规则', color: '#8a55d6', background: '#f7f2ff' },
};

export const ADDABLE_NODE_TYPES = [
  'ACTOR',
  'BUSINESS_OBJECT',
  'PROCESS',
  'EVENT',
  'DECISION',
  'SYSTEM',
  'POLICY',
] as const;

export function styleFor(type: string) {
  return NODE_STYLES[type] ?? { label: type, color: '#596579', background: '#f3f5f8' };
}
