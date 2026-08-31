import {
  LayoutDashboard,
  ScanLine,
  Network,
  Database,
  Blocks,
  Bot,
  Layers,
  ListChecks,
  FlaskConical,
  Rocket,
  Radio,
  Cpu,
  Activity,
  ShieldCheck,
  Package,
  Palette,
} from 'lucide-react';
export const navigation = [
  {
    group: '工作空间',
    items: [
      { id: 'overview', label: '总览', icon: LayoutDashboard },
      { id: 'analyst', label: 'AI 业务分析师', icon: ScanLine },
    ],
  },
  {
    group: '构建 · BUILD',
    items: [
      { id: 'business', label: '业务资产', icon: Network },
      { id: 'data', label: '数据平面', icon: Database },
      { id: 'capabilities', label: '能力资产', icon: Blocks },
      { id: 'agents', label: '业务 Agent', icon: Bot },
      { id: 'applications', label: '应用构建器', icon: Layers },
    ],
  },
  {
    group: '交付 · SHIP',
    items: [
      { id: 'human', label: '人工任务', icon: ListChecks },
      { id: 'evaluation', label: '测试与评测', icon: FlaskConical },
      { id: 'releases', label: '版本与发布', icon: Rocket },
      { id: 'channels', label: '渠道网关', icon: Radio },
    ],
  },
  {
    group: '运行 · OPERATE',
    items: [
      { id: 'runtime', label: '运行时与 Kernel', icon: Cpu },
      { id: 'observability', label: '可观测性', icon: Activity },
      { id: 'governance', label: '安全与治理', icon: ShieldCheck },
      { id: 'delivery', label: '商业交付', icon: Package },
      { id: 'design', label: '设计系统', icon: Palette },
    ],
  },
];
