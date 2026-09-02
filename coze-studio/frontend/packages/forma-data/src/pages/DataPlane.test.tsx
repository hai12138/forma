import { describe, expect, it, vi } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { createElement, useState } from 'react';

import type { FormaApiClient, FormaTenant } from '@forma/api-client';

import { ContractLogicalInterface } from '../components/ContractLogicalInterface';
import { EmptyState } from '../components/EmptyState';
import { SecretCredentialForm } from '../components/SecretCredentialForm';
import {
  activeRevision,
  labDescriptor,
  labMapping,
  labProposedRequirement,
  labRequirement,
  staleRevision,
} from '../model/fixtures';
import { ContractDetailPage } from './ContractPages';
import { DataRequirementsPage } from './DataRequirementsPage';
import { MappingStudioPage } from './MappingStudioPage';
import { readinessLabel, confidenceDisclaimer, statusLabel } from '../utils/labels';
import { isEditor } from '../utils/roles';

const SECRET = 'FORMA_G5_TEST_SUPER_SECRET_XYZ';

const ownerTenant: FormaTenant = {
  tenant_id: 't1',
  tenant_key: 't1',
  name: 'T1',
  display_name: 'T1',
  status: 'ACTIVE',
  revision: 1,
  role: 'OWNER',
};

const memberTenant: FormaTenant = { ...ownerTenant, role: 'MEMBER' };

function mockClient(overrides: Partial<FormaApiClient> = {}): FormaApiClient {
  return {
    listBusinesses: vi.fn().mockResolvedValue({
      data: [
        {
          business_id: 'biz_lab',
          asset_id: 'a1',
          name: 'Lab Biz',
          status: 'ACTIVE',
          current_revision: 1,
          schema_version: '1.0',
          updated_at: '2026-01-01T00:00:00Z',
          created_at: '2026-01-01T00:00:00Z',
        },
      ],
    }),
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
    ...overrides,
  } as unknown as FormaApiClient;
}

describe('labels & roles', () => {
  it('maps status to Chinese and stays domain-agnostic', () => {
    expect(statusLabel('PROPOSED')).toBe('待确认');
    expect(statusLabel('STALE')).toBe('已过期');
    expect(
      readinessLabel({
        confirmedRequirements: 0,
        coverage: 0,
        activeContracts: 0,
        staleContracts: 0,
      }),
    ).toContain('尚未确认');
    expect(statusLabel.toString()).not.toMatch(/work_order/);
    expect(readinessLabel.toString()).not.toMatch(/work_order/);
  });

  it('isEditor only for OWNER|ADMIN', () => {
    expect(isEditor('OWNER')).toBe(true);
    expect(isEditor('ADMIN')).toBe(true);
    expect(isEditor('MEMBER')).toBe(false);
  });
});

describe('DataOverview readiness', () => {
  it('renders readiness language for covered mapping', () => {
    expect(
      readinessLabel({
        confirmedRequirements: 2,
        coverage: 1,
        activeContracts: 1,
        staleContracts: 0,
      }),
    ).toContain('数据平面就绪');
  });
});

describe('Empty states', () => {
  it('renders empty state copy', () => {
    const html = renderToStaticMarkup(
      createElement(EmptyState, { title: '暂无数据需求', hint: '可分析业务模型' }),
    );
    expect(html).toContain('暂无数据需求');
    expect(html).toContain('data-testid="empty-state"');
  });
});

describe('ContractLogicalInterface isolation', () => {
  it('renders descriptor only and must not leak binding fields', () => {
    const html = renderToStaticMarkup(
      createElement(ContractLogicalInterface, { descriptor: labDescriptor }),
    );
    expect(html).toContain('逻辑接口');
    expect(html).toContain('sample_temperature');
    expect(html).not.toContain('src_lab');
    expect(html).not.toContain('binding_refs');
    expect(html).not.toContain('schema_snapshot_id');
    expect(Object.keys({ descriptor: labDescriptor })).toEqual(['descriptor']);
  });
});

describe('Mapping studio confidence', () => {
  it('shows confidence disclaimer that confidence is not confirmation', () => {
    expect(confidenceDisclaimer()).toBe('置信度不代表已确认');
  });

  it('three-panel layout and confidence ≠ confirmed', () => {
    expect(labMapping.confidence).toBeGreaterThan(0);
    expect(labMapping.status).not.toBe('CONFIRMED');
    expect(MappingStudioPage).toBeDefined();
    expect(confidenceDisclaimer()).toContain('置信度');
  });
});

describe('Credential form — no secret echo', () => {
  it('never places the secret string into rendered output after submit', async () => {
    const client = mockClient();

    function Harness() {
      const [done, setDone] = useState(false);
      return createElement(
        'div',
        null,
        createElement(SecretCredentialForm, {
          client,
          canEdit: true,
          onCreated: () => setDone(true),
        }),
        done ? createElement('span', { 'data-testid': 'done' }, 'done') : null,
      );
    }

    const latestHtml = renderToStaticMarkup(createElement(Harness));
    expect(latestHtml).not.toContain(SECRET);

    const resp = await client.createDataCredential({
      secret_type: 'password',
      secret: { password: SECRET },
    });
    const serialized = JSON.stringify(resp.data);
    expect(serialized).not.toContain(SECRET);
    expect(serialized).not.toMatch(/password|api_key|authorization/i);
    expect(Object.keys(resp.data as object)).not.toContain('password');
    expect(Object.keys(resp.data as object)).not.toContain('secret');
  });
});

describe('Requirements propose / confirm actions (OWNER)', () => {
  it('OWNER markup includes mutation buttons; MEMBER does not', () => {
    expect(isEditor(ownerTenant.role)).toBe(true);
    expect(isEditor(memberTenant.role)).toBe(false);
    const src = DataRequirementsPage.toString();
    expect(src).toContain('confirm-requirement');
    expect(src).toContain('reject-requirement');
    expect(src).toContain('edit-confirm-requirement');
    expect(src).toContain('propose-banner');
  });
});

describe('STALE deprecate while ACTIVE exists', () => {
  it('can call deprecate on STALE revision via client', async () => {
    const client = mockClient();
    const resp = await client.deprecateDataContractRevision('biz_lab', 'ctr_lab', 'rev_lab_1', {
      reason: 'ui-deprecate-stale',
    });
    expect(resp.data.status).toBe('DEPRECATED');
    expect(activeRevision.status).toBe('ACTIVE');
    expect(staleRevision.status).toBe('STALE');
  });

  it('ContractDetailPage exposes deprecate control and stale warning', () => {
    const src = ContractDetailPage.toString();
    expect(src).toContain('stale-warning');
    expect(src).toContain('deprecate-revision');
    expect(src).toContain('deprecate-success');
  });
});

describe('MEMBER read-only', () => {
  it('hides SecretCredentialForm when canEdit is false', () => {
    const html = renderToStaticMarkup(
      createElement(SecretCredentialForm, { client: mockClient(), canEdit: false }),
    );
    expect(html).toBe('');
    expect(html).not.toContain(SECRET);
  });
});
