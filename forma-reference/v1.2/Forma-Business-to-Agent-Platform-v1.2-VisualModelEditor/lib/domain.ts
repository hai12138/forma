export type Agent = {
  id: string;
  name: string;
  description: string;
  role: string;
  version: string;
  status: '草稿' | '已验证' | '已发布';
  capabilities: string[];
  knowledge: string;
  permission: string;
  interaction: string;
  deleted: boolean;
  history: { version: string; note: string }[];
};
export type Approval = { role: string; confirmed: boolean };
export type Evaluation = {
  revision: number;
  passed: boolean;
  total: number;
  failed: number;
  at: string;
} | null;
export type State = {
  visualModel?: import('./visual-model').VisualModel;
  agents: Agent[];
  approvals: Approval[];
  rule: string;
  revision: number;
  dataMode: string;
  mapping: Record<string, string>;
  application: {
    name: string;
    selected: string[];
    mode: string;
    context: string;
    knowledge: string;
    fallback: string;
  };
  evaluation: Evaluation;
  frozen: number | null;
  stage: string;
  canary: number;
  released: boolean;
  human: Record<string, string>;
  channels: Record<string, string>;
  runtime: string;
  audit: { at: string; action: string }[];
  discovery: string[];
};
export const capabilities = [
  {
    id: 'work_order.create',
    name: '创建工单',
    impl: 'Managed Runtime',
    risk: '写入',
    version: '1.2.0',
  },
  {
    id: 'work_order.assign',
    name: '分配工单',
    impl: 'REST API',
    risk: '高风险',
    version: '1.3.0',
  },
  {
    id: 'work_order.auto_assign',
    name: '紧急工单自动派单',
    impl: 'Workflow',
    risk: '高风险',
    version: '1.0.0',
  },
  {
    id: 'asset.lookup',
    name: '查询设备档案',
    impl: 'Database',
    risk: '只读',
    version: '2.1.0',
  },
  {
    id: 'knowledge.search',
    name: '检索服务知识',
    impl: 'MCP',
    risk: '只读',
    version: '1.0.0',
  },
  {
    id: 'work_order.approve',
    name: '主管确认',
    impl: 'Human Task',
    risk: '人工审批',
    version: '1.0.0',
  },
];
const specs = [
  ['work-order', '工单 Agent', '受理、分派与跟进园区工单', '物业服务专员'],
  ['equipment', '设备 Agent', '查询设备、诊断故障与关联维保', '设施运维专家'],
  ['energy', '能耗 Agent', '分析用量趋势与异常耗能', '能源分析师'],
  ['service', '客服 Agent', '回答园区服务问题与知识检索', '客户服务顾问'],
  ['investment', '招商 Agent', '商机识别与企业意向跟进', '招商顾问'],
  ['visitor', '访客 Agent', '访客预约、核验与通行协同', '访客服务专员'],
  ['inspection', '巡检 Agent', '生成巡检任务与异常上报', '设施巡检员'],
  ['enterprise', '企业服务 Agent', '企业诉求分类与政策服务', '企业服务顾问'],
];
export const initialState: State = {
  agents: specs.map((a, i) => ({
    id: a[0],
    name: a[1],
    description: a[2],
    role: a[3],
    version: '1.' + (i % 3) + '.0',
    status: i < 2 ? '已发布' : '草稿',
    capabilities:
      i === 0
        ? ['work_order.create', 'work_order.assign', 'work_order.auto_assign']
        : i === 1
          ? ['asset.lookup']
          : ['knowledge.search'],
    knowledge: '园区服务手册 · 2026.08',
    permission: 'tenant:current · 最小权限',
    interaction: '先澄清必要信息；执行写操作前确认；禁止猜测执行结果。',
    deleted: false,
    history: [{ version: '1.0.0', note: '初始业务定义' }],
  })),
  approvals: [
    { role: '业务负责人 · 李经理', confirmed: false },
    { role: 'IT 负责人 · 王工', confirmed: false },
  ],
  rule: '紧急工单自动派单；普通工单由主管分派。',
  revision: 1,
  dataMode: 'Hybrid',
  mapping: {
    id: 'ticket_id',
    priority: 'urgency',
    assignee: 'owner_id',
    status: 'state',
  },
  application: {
    name: '园区智能助手',
    selected: ['work-order', 'equipment', 'energy', 'service'],
    mode: 'Supervisor',
    context: 'tenant_id, actor_id, location, work_order_id',
    knowledge: '园区服务手册 · 已审核知识',
    fallback: '转人工任务中心',
  },
  evaluation: null,
  frozen: null,
  stage: 'Dev',
  canary: 10,
  released: false,
  human: { 'HT-1024': '待处理', 'HT-1025': '待处理', 'HT-1026': '待处理' },
  channels: { 'Web / H5': '已配置', 企业微信: '已配置' },
  runtime: 'LangGraph',
  audit: [{ at: '09:15', action: '生产版本 v1.3 冻结并部署（示例历史）' }],
  discovery: [],
};
export function invalidate(s: State, patch: Partial<State>): State {
  return {
    ...s,
    ...patch,
    revision: s.revision + 1,
    evaluation: null,
    frozen: null,
    stage: 'Dev',
    released: false,
  };
}
export function gateReasons(s: State): string[] {
  return [
    !s.approvals.every((a) => a.confirmed) && '业务规则尚未完成双人确认',
    (!s.evaluation?.passed || s.evaluation.revision !== s.revision) &&
      '当前资产快照尚未通过回归',
    !s.application.selected.length && '应用至少需要一个 Agent',
    s.application.selected.some(
      (id) => !s.agents.some((a) => a.id === id && !a.deleted),
    ) && '应用引用了失效 Agent',
  ].filter((v): v is string => typeof v === 'string');
}
export function validateImport(raw: unknown): Agent[] {
  const xs = Array.isArray(raw) ? raw : [raw];
  if (!xs.length || xs.length > 100) throw Error('一次导入 1–100 个 Agent');
  return xs.map((x) => {
    if (!x || typeof x !== 'object') throw Error('Agent 格式错误');
    const a = x as Record<string, unknown>;
    if (
      typeof a.name !== 'string' ||
      !a.name.trim() ||
      a.name.length > 60 ||
      typeof a.role !== 'string' ||
      !Array.isArray(a.capabilities) ||
      !a.capabilities.every(
        (c) => typeof c === 'string' && capabilities.some((k) => k.id === c),
      )
    )
      throw Error('名称、Role 或 Capability 引用无效');
    return {
      id: crypto.randomUUID(),
      name: a.name.trim(),
      role: a.role,
      description:
        typeof a.description === 'string' ? a.description : '导入的业务 Agent',
      version: '0.1.0',
      status: '草稿',
      capabilities: a.capabilities as string[],
      knowledge: typeof a.knowledge === 'string' ? a.knowledge : '',
      permission: 'tenant:current · 最小权限',
      interaction:
        typeof a.interaction === 'string' ? a.interaction : '执行前确认',
      deleted: false,
      history: [],
    };
  });
}
export function download(name: string, data: unknown) {
  const url = URL.createObjectURL(
    new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }),
  );
  const a = document.createElement('a');
  a.href = url;
  a.download = name;
  a.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
