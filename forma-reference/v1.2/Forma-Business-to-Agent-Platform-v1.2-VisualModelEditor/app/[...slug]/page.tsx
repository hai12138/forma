import Platform from '@/components/platform';
import type { Metadata } from 'next';
const titles: Record<string, string> = {
  analyst: 'AI 业务分析师',
  business: '园区工单业务模型',
  data: '数据平面',
  capabilities: '业务能力资产',
  agents: '业务 Agent 中心',
  applications: '园区智能助手 · 应用构建器',
  human: '人工任务中心',
  evaluation: '测试与评测',
  releases: '版本与发布',
  channels: '渠道网关',
  runtime: 'Runtime 与 Platform Kernel',
  observability: '可观测性',
  governance: '安全与治理',
  delivery: '商业交付',
  design: 'Forma 设计系统',
};
export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string[] }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const title = (titles[slug[0]] || '页面未找到') + ' · Forma';
  const description =
    (titles[slug[0]] || '工作空间') +
    '：Business-to-Agent Platform 交互原型。所有外部执行均为模拟。';
  return {
    title,
    description,
    openGraph: { title, description, images: [] },
    twitter: { card: 'summary', title, description, images: [] },
  };
}
export default function Screen() {
  return <Platform />;
}
