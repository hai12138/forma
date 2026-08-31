import type { Agent } from './domain';

export type BusinessNodeType =
  | 'role'
  | 'entity'
  | 'process'
  | 'state'
  | 'rule'
  | 'external'
  | 'agent'
  | 'application';
export type BusinessNode = {
  id: string;
  type: BusinessNodeType;
  label: string;
  description?: string;
  position?: { x: number; y: number };
  route?: string;
};
export type BusinessEdge = {
  id: string;
  source: string;
  target: string;
  label: string;
  dashed?: boolean;
};
export type BusinessGraph = { nodes: BusinessNode[]; edges: BusinessEdge[] };
export const nodeTypes: Record<
  BusinessNodeType,
  { label: string; color: string; background: string }
> = {
  role: { label: '角色', color: '#2563eb', background: '#f0f5ff' },
  entity: { label: '实体', color: '#c47712', background: '#fff8eb' },
  process: { label: '流程', color: '#db477e', background: '#fff2f7' },
  state: { label: '状态', color: '#0891a8', background: '#effbfd' },
  rule: { label: '规则', color: '#8a55d6', background: '#f7f2ff' },
  external: { label: '外部系统', color: '#596579', background: '#f3f5f8' },
  agent: { label: 'Agent', color: '#188362', background: '#effaf5' },
  application: { label: '应用', color: '#5160d8', background: '#f2f3ff' },
};
export const edge = (
  source: string,
  target: string,
  label: string,
  dashed = false,
): BusinessEdge => ({
  id: `${source}-${target}`,
  source,
  target,
  label,
  dashed,
});
export function chain(
  nodes: BusinessNode[],
  labels: string[] = [],
): BusinessGraph {
  return {
    nodes,
    edges: nodes
      .slice(1)
      .map((n, i) => edge(nodes[i].id, n.id, labels[i] || '依赖')),
  };
}
// Optional positions are presentation hints. JSON without coordinates receives a deterministic layered layout.
export function layoutGraph(graph: BusinessGraph) {
  const ids = new Set(graph.nodes.map((n) => n.id));
  const depths = new Map<string, number>();
  const pending = new Set(ids);
  for (let pass = 0; pass < graph.nodes.length && pending.size; pass++) {
    for (const id of pending) {
      const parents = graph.edges
        .filter((e) => e.target === id && ids.has(e.source))
        .map((e) => e.source);
      if (parents.every((p) => depths.has(p))) {
        depths.set(id, Math.max(-1, ...parents.map((p) => depths.get(p)!)) + 1);
        pending.delete(id);
      }
    }
  }
  // Cycles remain visible in a final column instead of hanging the renderer.
  const fallback = Math.max(0, ...depths.values()) + 1;
  const rows = new Map<number, number>();
  return graph.nodes.map((n) => {
    const col = depths.get(n.id) ?? fallback;
    const row = rows.get(col) ?? 0;
    rows.set(col, row + 1);
    return { ...n, position: n.position ?? { x: col * 290, y: row * 150 } };
  });
}
export function workOrderGraph(
  rule: string,
  confirmed: boolean,
): BusinessGraph {
  const n = (
    id: string,
    type: BusinessNodeType,
    label: string,
    description: string,
    x: number,
    y: number,
  ): BusinessNode => ({ id, type, label, description, position: { x, y } });
  return {
    nodes: [
      n('reporter', 'role', '报修人 / 客户', '发起报修 · 确认结果', 0, 180),
      n('service', 'role', '客服专员', '登记与受理业务请求', 0, 0),
      n('order', 'entity', '工单 · WorkOrder', '报修内容与处理责任', 310, 180),
      n('asset', 'entity', '设备 · Asset', '关联待维修设备', 310, 0),
      n('supervisor', 'role', '物业主管', '普通工单分派 / 人工兜底', 620, 0),
      n(
        'dispatch',
        'process',
        '派单 · Assignment',
        '普通人工 / 紧急自动',
        620,
        180,
      ),
      n(
        'rule',
        'rule',
        confirmed ? '派单规则 · 已确认' : '派单规则 · 候选',
        rule,
        620,
        390,
      ),
      n('engineer', 'role', '当班工程师', '接单并执行维修任务', 930, 180),
      n(
        'approval',
        'process',
        '验收 · Approval',
        '客户验收，高风险人工确认',
        930,
        390,
      ),
      n(
        'status',
        'state',
        '工单生命周期',
        '已创建 → 待分派 → 处理中 → 待验收 → 已完成 → 已关闭',
        310,
        390,
      ),
      n('system', 'external', '客户设备系统', '按契约读取设备信息', 620, -180),
    ],
    edges: [
      edge('reporter', 'order', '提交报修'),
      edge('service', 'order', '登记'),
      edge('order', 'asset', '关联设备'),
      edge('asset', 'system', '读取'),
      edge('order', 'dispatch', '请求分派'),
      edge('supervisor', 'dispatch', '普通 / 兜底'),
      edge('rule', 'dispatch', '约束', true),
      edge('dispatch', 'engineer', '分配责任'),
      edge('engineer', 'approval', '提交结果'),
      edge('approval', 'status', '验收后迁移'),
      edge('order', 'status', '记录状态'),
    ],
  };
}
export function impactGraph(
  capability = 'work_order.auto_assign',
): BusinessGraph {
  return chain(
    [
      {
        id: 'rule',
        type: 'rule',
        label: 'BR-021',
        description: '业务规则 / Business Model',
        route: 'business',
      },
      {
        id: 'capability',
        type: 'process',
        label: capability,
        description: 'Capability Asset',
        route: 'capabilities',
      },
      {
        id: 'agent',
        type: 'agent',
        label: '工单 Agent',
        description: 'Business Agent Asset',
        route: 'agents',
      },
      {
        id: 'app',
        type: 'application',
        label: '园区智能助手',
        description: 'Application Asset',
        route: 'applications',
      },
    ],
    ['约束能力', '调用能力', '组合交付'],
  );
}
export function workOrderSummary(
  rule: string,
  confirmed: boolean,
): BusinessGraph {
  const graph = workOrderGraph(rule, confirmed);
  const ids = ['reporter', 'order', 'dispatch', 'rule'];
  return {
    nodes: graph.nodes
      .filter((n) => ids.includes(n.id))
      .map((n) => ({
        ...n,
        position:
          n.id === 'rule' ? { x: 620, y: 160 } : { x: n.position!.x, y: 0 },
      })),
    edges: graph.edges.filter(
      (e) => ids.includes(e.source) && ids.includes(e.target),
    ),
  };
}
export function dataGraph(mode: string): BusinessGraph {
  const objects = ['WorkOrder', 'Asset', 'Assignment', 'Approval'];
  const nodes: BusinessNode[] = objects.map((id) => ({
    id,
    type: 'entity',
    label: id,
    description: 'Canonical Object · v1.2',
  }));
  const edges = objects.map((id) =>
    edge(
      id,
      mode === 'External' || (mode === 'Hybrid' && id === 'Asset')
        ? 'external'
        : 'managed',
      '存储归属',
    ),
  );
  if (edges.some((e) => e.target === 'external'))
    nodes.push({
      id: 'external',
      type: 'external',
      label: '客户业务系统',
      description: 'External · 数据留在原系统',
    });
  if (edges.some((e) => e.target === 'managed'))
    nodes.push({
      id: 'managed',
      type: 'process',
      label: 'Managed Runtime',
      description: '平台托管 · 独立租户 Schema',
    });
  return { nodes, edges };
}
export function applicationGraph(agents: Agent[], mode: string): BusinessGraph {
  const nodes: BusinessNode[] = [
    {
      id: 'gateway',
      type: 'external',
      label: 'Channel Gateway',
      description: '统一身份与 Behavior Policy',
    },
    {
      id: 'Supervisor',
      type: 'process',
      label: mode,
      description: '应用编排 · 共享受控 Context',
    },
    ...agents.map((a) => ({
      id: a.id,
      type: 'agent' as const,
      label: a.name,
      description: `v${a.version} · ${a.capabilities.length} 项能力`,
    })),
    {
      id: 'output',
      type: 'state',
      label: '结果校验 / 人工兜底',
      description: '查看冲突与降级策略',
    },
  ];
  const sequential = mode === 'Pipeline' || mode === 'Handoff';
  const edges = [edge('gateway', 'Supervisor', '授权请求')];
  agents.forEach((a, i) => {
    edges.push(
      edge(
        sequential && i > 0 ? agents[i - 1].id : 'Supervisor',
        a.id,
        sequential
          ? mode === 'Handoff'
            ? '移交上下文'
            : '顺序执行'
          : mode === 'Router'
            ? '按意图择一'
            : mode === 'Parallel'
              ? '并行调用'
              : mode === 'Human-in-the-loop'
                ? '审批后继续'
                : '协调任务',
      ),
    );
    if (!sequential || i === agents.length - 1)
      edges.push(edge(a.id, 'output', '返回结果'));
  });
  if (!agents.length)
    edges.push(edge('Supervisor', 'output', '未选择 Agent · 发布受阻', true));
  return { nodes, edges };
}
