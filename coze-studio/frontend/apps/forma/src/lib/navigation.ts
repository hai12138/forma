export const navigation = [
  {
    group: '工作空间',
    items: [
      { id: 'overview', label: '总览', path: '/' },
      { id: 'analyst', label: 'AI 业务分析师', path: '/analyst' },
    ],
  },
  {
    group: '构建 · BUILD',
    items: [
      { id: 'business', label: '业务资产', path: '/business' },
      { id: 'data', label: '数据平面', path: '/data' },
      { id: 'capabilities', label: '能力资产', path: '/capabilities' },
      { id: 'agents', label: '业务 Agent', path: '/agents' },
      { id: 'applications', label: '应用构建器', path: '/applications' },
    ],
  },
  {
    group: '交付 · SHIP',
    items: [
      { id: 'human', label: '人工任务', path: '/human' },
      { id: 'evaluation', label: '测试与评测', path: '/evaluation' },
      { id: 'releases', label: '版本与发布', path: '/releases' },
      { id: 'channels', label: '渠道网关', path: '/channels' },
    ],
  },
  {
    group: '运行 · OPERATE',
    items: [
      { id: 'runtime', label: '运行时与 Kernel', path: '/runtime' },
      { id: 'observability', label: '可观测性', path: '/observability' },
      { id: 'governance', label: '安全与治理', path: '/governance' },
      { id: 'delivery', label: '商业交付', path: '/delivery' },
      { id: 'design', label: '设计系统', path: '/design' },
    ],
  },
] as const;

export const routeIds = navigation.flatMap(g => g.items.map(i => i.id));

export const adminNavigation = [
  {
    group: '系统管理',
    items: [{ id: 'admin-users', label: '用户管理', path: '/admin/users' }],
  },
] as const;
