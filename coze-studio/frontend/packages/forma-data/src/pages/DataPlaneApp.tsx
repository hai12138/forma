import { Route, Routes } from 'react-router-dom';

import type { FormaApiClient, FormaTenant } from '@forma/api-client';

import { DataPlaneShell } from '../components/DataPlaneShell';
import { ContractDetailPage, DataContractsPage } from './ContractPages';
import { DataHealthPage } from './DataHealthPage';
import { DataOverviewPage } from './DataOverviewPage';
import { DataRequirementsPage } from './DataRequirementsPage';
import { DataSourcesPage, SourceDetailPage } from './DataSourcesPage';
import { MappingStudioPage } from './MappingStudioPage';

export interface DataPlaneAppProps {
  client: FormaApiClient;
  currentTenant: FormaTenant | null;
}

export function DataPlaneApp({ client, currentTenant }: DataPlaneAppProps) {
  return (
    <Routes>
      <Route element={<DataPlaneShell client={client} currentTenant={currentTenant} />}>
        <Route index element={<DataOverviewPage />} />
        <Route path="requirements" element={<DataRequirementsPage />} />
        <Route path="sources" element={<DataSourcesPage />} />
        <Route path="sources/:sourceId" element={<SourceDetailPage />} />
        <Route path="mappings" element={<MappingStudioPage />} />
        <Route path="contracts" element={<DataContractsPage />} />
        <Route path="contracts/:contractId" element={<ContractDetailPage />} />
        <Route path="health" element={<DataHealthPage />} />
      </Route>
    </Routes>
  );
}
