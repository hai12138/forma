import { Route, Routes } from 'react-router-dom';

import { BusinessEditorPage, BusinessListPage } from '@forma/business';

import { AppShell } from '@/components/shell';
import { useFormaSession } from '@/hooks/use-forma-session';
import { DesignPage, OverviewPage, PlaceholderPage } from '@/pages';

function BusinessListRoute() {
  const { client, currentTenant } = useFormaSession();
  return <BusinessListPage client={client} currentTenant={currentTenant} />;
}

function BusinessEditorRoute() {
  const { client, currentTenant } = useFormaSession();
  return <BusinessEditorPage client={client} currentTenant={currentTenant} />;
}

export function AppRouter() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<OverviewPage />} />
        <Route path="/design" element={<DesignPage />} />
        <Route path="/analyst" element={<PlaceholderPage title="AI 业务分析师" />} />
        <Route path="/business" element={<BusinessListRoute />} />
        <Route path="/business/:businessId" element={<BusinessEditorRoute />} />
        <Route path="/data" element={<PlaceholderPage title="数据平面" />} />
        <Route path="/capabilities" element={<PlaceholderPage title="能力资产" />} />
        <Route path="/agents" element={<PlaceholderPage title="业务 Agent" />} />
        <Route path="/applications" element={<PlaceholderPage title="应用构建器" />} />
        <Route path="/human" element={<PlaceholderPage title="人工任务" />} />
        <Route path="/evaluation" element={<PlaceholderPage title="测试与评测" />} />
        <Route path="/releases" element={<PlaceholderPage title="版本与发布" />} />
        <Route path="/channels" element={<PlaceholderPage title="渠道网关" />} />
        <Route path="/runtime" element={<PlaceholderPage title="运行时与 Kernel" />} />
        <Route path="/observability" element={<PlaceholderPage title="可观测性" />} />
        <Route path="/governance" element={<PlaceholderPage title="安全与治理" />} />
        <Route path="/delivery" element={<PlaceholderPage title="商业交付" />} />
      </Routes>
    </AppShell>
  );
}
