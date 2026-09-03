import { createElement } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { act } from 'react-dom/test-utils';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { vi } from 'vitest';

import type { FormaApiClient, FormaTenant } from '@forma/api-client';

import { DataPlaneApp } from './pages/DataPlaneApp';
import {
  activeRevision,
  labBusiness,
  labDescriptor,
  labMapping,
  labProposedRequirement,
  labRequirement,
  labSource,
  staleRevision,
} from './model/fixtures';

export const SECRET = 'FORMA_G5_TEST_SUPER_SECRET_XYZ';

export const ownerTenant: FormaTenant = {
  tenant_id: 't1',
  tenant_key: 't1',
  name: 'T1',
  display_name: 'T1',
  status: 'ACTIVE',
  revision: 1,
  role: 'OWNER',
};

export const memberTenant: FormaTenant = { ...ownerTenant, role: 'MEMBER' };
export const viewerTenant: FormaTenant = { ...ownerTenant, role: 'VIEWER' };

export function mockClient(overrides: Partial<FormaApiClient> = {}): FormaApiClient {
  const base = {
    listBusinesses: vi.fn().mockResolvedValue({ data: [labBusiness] }),
    listDataRequirements: vi.fn().mockResolvedValue({ data: [labRequirement, labProposedRequirement] }),
    getSemanticMappingCoverage: vi.fn().mockResolvedValue({
      data: {
        total_confirmed_requirements: 1,
        confirmed_mappings: 1,
        unmapped_requirement_ids: [],
        coverage: 1,
      },
    }),
    listDataContracts: vi.fn().mockResolvedValue({
      data: [
        {
          contract_id: 'ctr_lab',
          business_id: 'biz_lab',
          active_revision_id: 'rev_lab_2',
          created_by: 'p',
          created_at: '',
          updated_at: '',
        },
      ],
    }),
    getDataContractRevision: vi.fn().mockResolvedValue({ data: activeRevision }),
    listSemanticMappings: vi.fn().mockResolvedValue({ data: [labMapping] }),
    listDataContractRevisions: vi.fn().mockResolvedValue({ data: [staleRevision, activeRevision] }),
    getActiveDataContractDescriptor: vi.fn().mockResolvedValue({ data: labDescriptor }),
    deprecateDataContractRevision: vi.fn().mockResolvedValue({
      data: { ...staleRevision, status: 'DEPRECATED' },
    }),
    validateDataContractRevision: vi.fn().mockResolvedValue({
      data: {
        revision: { ...activeRevision, status: 'VALIDATED' },
        result: {
          ValidationID: 'val_1',
          TenantID: 't1',
          BusinessID: 'biz_lab',
          ContractID: 'ctr_lab',
          RevisionID: 'rev_lab_draft',
          Version: 3,
          Status: 'PASS',
          Errors: [],
          Warnings: [],
          SnapshotFingerprints: {},
          ValidatedBy: 'p',
          ValidatedAt: '2026-01-01T00:00:00Z',
          CreatedAt: '2026-01-01T00:00:00Z',
        },
      },
    }),
    activateDataContractRevision: vi.fn().mockResolvedValue({
      data: { ...activeRevision, status: 'ACTIVE' },
    }),
    createDataCredential: vi.fn().mockResolvedValue({
      data: {
        credential_ref_id: 'cred_ok',
        status: 'ACTIVE',
        provider: 'sql',
        created_at: '2026-01-01T00:00:00Z',
      },
    }),
    confirmDataRequirement: vi.fn().mockResolvedValue({ data: { requirement: labRequirement, decision: {} } }),
    rejectDataRequirement: vi.fn().mockResolvedValue({
      data: { requirement: labProposedRequirement, decision: {} },
    }),
    editConfirmDataRequirement: vi.fn().mockResolvedValue({
      data: { original: labProposedRequirement, replacement: labRequirement, decision: {} },
    }),
    createManualDataRequirement: vi.fn().mockResolvedValue({ data: labRequirement }),
    analyzeDataRequirements: vi.fn().mockResolvedValue({
      data: {
        analysis_run: { analysis_run_id: 'ar1' },
        requirements: [labProposedRequirement],
        owned_execute: true,
      },
    }),
    listDataSources: vi.fn().mockResolvedValue({ data: [labSource] }),
    getDataSource: vi.fn().mockResolvedValue({ data: labSource }),
    listDataConnections: vi.fn().mockResolvedValue({ data: [] }),
    listDataAssets: vi.fn().mockResolvedValue({ data: [] }),
    createDataSource: vi.fn().mockResolvedValue({ data: labSource }),
    createManualSemanticMapping: vi.fn().mockResolvedValue({ data: labMapping }),
    analyzeSemanticMappings: vi.fn().mockResolvedValue({ data: { mappings: [], owned_execute: true } }),
    confirmSemanticMapping: vi.fn().mockResolvedValue({ data: { mapping: labMapping, decision: {} } }),
    rejectSemanticMapping: vi.fn().mockResolvedValue({ data: { mapping: labMapping, decision: {} } }),
    createDataContract: vi.fn().mockResolvedValue({ data: { contract: { contract_id: 'ctr_new' } } }),
    listDataContractLifecycleEvents: vi.fn().mockResolvedValue({ data: [] }),
    listDataContractValidationResults: vi.fn().mockResolvedValue({ data: [] }),
    listDataContractDriftResults: vi.fn().mockResolvedValue({ data: [] }),
    listDataContractGapResults: vi.fn().mockResolvedValue({ data: [] }),
    evaluateDataContractDrift: vi.fn().mockResolvedValue({ data: { result: {}, revision: activeRevision } }),
    evaluateDataContractGap: vi.fn().mockResolvedValue({ data: {} }),
    getSchemaSnapshot: vi.fn().mockResolvedValue({ data: { snapshot_id: 'snap_lab', schema: { fields: [] } } }),
    ...overrides,
  };
  return new Proxy(base, {
    get(target, prop) {
      if (prop in target) return target[prop as keyof typeof target];
      return vi.fn().mockResolvedValue({ data: [] });
    },
  }) as unknown as FormaApiClient;
}

export interface RenderedPlane {
  container: HTMLElement;
  unmount: () => void;
}

export async function renderDataPlane(
  path: string,
  opts: { client?: FormaApiClient; tenant?: FormaTenant } = {},
): Promise<RenderedPlane> {
  const client = opts.client ?? mockClient();
  const tenant = opts.tenant ?? ownerTenant;
  const container = document.createElement('div');
  document.body.appendChild(container);
  const root: Root = createRoot(container);
  await act(async () => {
    root.render(
      createElement(
        MemoryRouter,
        {
          initialEntries: [path],
          future: { v7_startTransition: true, v7_relativeSplatPath: true },
        },
        createElement(
          Routes,
          null,
          createElement(Route, {
            path: '/data/*',
            element: createElement(DataPlaneApp, { client, currentTenant: tenant }),
          }),
        ),
      ),
    );
  });
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
  return {
    container,
    unmount: () => {
      act(() => {
        root.unmount();
      });
      container.remove();
    },
  };
}

export async function waitFor(assertFn: () => void, timeout = 2500): Promise<void> {
  const start = Date.now();
  let lastErr: unknown;
  while (Date.now() - start < timeout) {
    try {
      assertFn();
      return;
    } catch (err) {
      lastErr = err;
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 20));
      });
    }
  }
  throw lastErr;
}

export async function click(el: Element | null): Promise<void> {
  if (!el) throw new Error('click target missing');
  await act(async () => {
    (el as HTMLElement).click();
  });
}

export async function setValue(
  el: HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement | null,
  value: string,
): Promise<void> {
  if (!el) throw new Error('input missing');
  await act(async () => {
    const proto =
      el instanceof HTMLSelectElement ? HTMLSelectElement.prototype : HTMLInputElement.prototype;
    Object.getOwnPropertyDescriptor(proto, 'value')?.set?.call(el, value);
    const tracker = (el as unknown as { _valueTracker?: { setValue: (v: string) => void } })
      ._valueTracker;
    tracker?.setValue('');
    el.dispatchEvent(new Event('input', { bubbles: true }));
    el.dispatchEvent(new Event('change', { bubbles: true }));
  });
}
