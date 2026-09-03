import { Route, Routes } from 'react-router-dom';

import { BusinessEditorPage, BusinessListPage } from '@forma/business';
import { AnalystWorkspacePage } from '@forma/analyst';
import { DataPlaneApp } from '@forma/data';

import { FormaAuthGuard } from '@/components/FormaAuthGuard';
import { AppShell } from '@/components/shell';
import { useFormaSession } from '@/hooks/use-forma-session';
import { AdminUsersPage } from '@/pages/AdminUsersPage';
import { ChangePasswordPage } from '@/pages/ChangePasswordPage';
import { DesignPage, OverviewPage, PlaceholderPage } from '@/pages';
import { LoginPage } from '@/pages/LoginPage';
import { OnboardingPage } from '@/pages/OnboardingPage';

function BusinessListRoute() {
  const { client, currentTenant } = useFormaSession();
  return <BusinessListPage client={client} currentTenant={currentTenant} />;
}

function BusinessEditorRoute() {
  const { client, currentTenant } = useFormaSession();
  return <BusinessEditorPage client={client} currentTenant={currentTenant} />;
}

function AnalystRoute() {
  const { client, currentTenant } = useFormaSession();
  return <AnalystWorkspacePage client={client} currentTenant={currentTenant} />;
}

function DataPlaneRoute() {
  const { client, currentTenant } = useFormaSession();
  return <DataPlaneApp client={client} currentTenant={currentTenant} />;
}

function ProtectedApp() {
  return (
    <FormaAuthGuard>
      <AppShell>
        <Routes>
          <Route path="/" element={<OverviewPage />} />
          <Route path="/design" element={<DesignPage />} />
          <Route path="/analyst" element={<AnalystRoute />} />
          <Route path="/business" element={<BusinessListRoute />} />
          <Route path="/business/:businessId" element={<BusinessEditorRoute />} />
          <Route path="/data/*" element={<DataPlaneRoute />} />
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
          <Route path="/admin/users" element={<AdminUsersPage />} />
        </Routes>
      </AppShell>
    </FormaAuthGuard>
  );
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/change-password" element={<ChangePasswordPage />} />
      <Route path="/onboarding" element={<OnboardingPage />} />
      <Route path="/*" element={<ProtectedApp />} />
    </Routes>
  );
}
