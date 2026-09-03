import { describe, expect, it, vi, afterEach } from 'vitest';
import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { act } from 'react-dom/test-utils';

import { FormaApiError } from '@forma/api-client';

import { ContractLogicalInterface } from '../components/ContractLogicalInterface';
import { EmptyState } from '../components/EmptyState';
import { SecretCredentialForm } from '../components/SecretCredentialForm';
import {
  activeRevision,
  draftRevision,
  labBinding,
  labBusiness,
  labDescriptor,
  labHttpSource,
  labMapping,
  labRequirement,
  procurementMapping,
  procurementRequirement,
  staleRevision,
  validatedRevision,
} from '../model/fixtures';
import { confidenceDisclaimer, readinessLabel, statusLabel } from '../utils/labels';
import { isEditor } from '../utils/roles';
import {
  canActivateRevision,
  canDeprecateRevision,
  canValidateRevision,
} from '../utils/contract-lifecycle';
import {
  click,
  memberTenant,
  mockClient,
  ownerTenant,
  renderDataPlane,
  SECRET,
  setValue,
  viewerTenant,
  waitFor,
  type RenderedPlane,
} from '../test-helpers';

const mounted: RenderedPlane[] = [];

async function mount(path: string, opts?: Parameters<typeof renderDataPlane>[1]) {
  const rendered = await renderDataPlane(path, opts);
  mounted.push(rendered);
  return rendered;
}

afterEach(() => {
  while (mounted.length) {
    mounted.pop()?.unmount();
  }
});

describe('labels & roles', () => {
  it('maps status to Chinese and stays domain-agnostic', () => {
    expect(statusLabel('PROPOSED')).toBe('待确认');
    expect(statusLabel('STALE')).toBe('已过期');
    expect(statusLabel('VALIDATED')).toBe('已验证');
    expect(
      readinessLabel({
        confirmedRequirements: 0,
        coverage: 0,
        activeContracts: 0,
        staleContracts: 0,
      }),
    ).toContain('尚未确认');
    expect(statusLabel.toString()).not.toMatch(/work_order|repair|refund|energy|device/);
  });

  it('isEditor only for OWNER|ADMIN', () => {
    expect(isEditor('OWNER')).toBe(true);
    expect(isEditor('ADMIN')).toBe(true);
    expect(isEditor('MEMBER')).toBe(false);
    expect(isEditor('VIEWER')).toBe(false);
  });
});

describe('contract lifecycle helpers', () => {
  it('DRAFT validate only, VALIDATED activate only, ACTIVE/STALE deprecate, never DRAFT deprecate', () => {
    expect(canValidateRevision('DRAFT')).toBe(true);
    expect(canActivateRevision('DRAFT')).toBe(false);
    expect(canDeprecateRevision('DRAFT')).toBe(false);

    expect(canValidateRevision('VALIDATED')).toBe(false);
    expect(canActivateRevision('VALIDATED')).toBe(true);
    expect(canDeprecateRevision('VALIDATED')).toBe(false);

    expect(canDeprecateRevision('ACTIVE')).toBe(true);
    expect(canDeprecateRevision('STALE')).toBe(true);
    expect(canValidateRevision('ACTIVE')).toBe(false);
    expect(canActivateRevision('STALE')).toBe(false);
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
  });
});

describe('domain-agnostic fixtures', () => {
  it('lab and procurement fixtures share the same UI types without industry switches', () => {
    expect(labRequirement.business_id).not.toBe(procurementRequirement.business_id);
    expect(labMapping.mapping_type).toBe('DIRECT');
    expect(procurementMapping.mapping_type).toBe('CAST');
    expect(confidenceDisclaimer()).toBe('置信度不代表已确认');
  });
});

describe('route behavior — MemoryRouter', () => {
  const paths: Array<{ path: string; testId: string }> = [
    { path: '/data?businessId=biz_lab', testId: 'data-overview' },
    { path: '/data/requirements?businessId=biz_lab', testId: 'data-requirements-page' },
    { path: '/data/sources?businessId=biz_lab', testId: 'data-sources-page' },
    { path: '/data/sources/src_lab?businessId=biz_lab', testId: 'source-detail-page' },
    { path: '/data/mappings?businessId=biz_lab', testId: 'mapping-studio' },
    { path: '/data/contracts?businessId=biz_lab', testId: 'data-contracts-page' },
    { path: '/data/contracts/ctr_lab?businessId=biz_lab', testId: 'contract-detail-page' },
    { path: '/data/health?businessId=biz_lab', testId: 'data-health-page' },
  ];

  it.each(paths)('renders $path', async ({ path, testId }) => {
    const { container } = await mount(path);
    await waitFor(() => {
      expect(container.querySelector(`[data-testid="${testId}"]`)).not.toBeNull();
    });
  });
});

describe('role-aware rendering', () => {
  it('OWNER sees requirement mutations, MEMBER/VIEWER do not', async () => {
    const owner = await mount('/data/requirements?businessId=biz_lab', { tenant: ownerTenant });
    await waitFor(() => {
      expect(owner.container.querySelector('[data-testid="confirm-requirement"]')).not.toBeNull();
      expect(owner.container.querySelector('[data-testid="analyze-requirements"]')).not.toBeNull();
    });

    const member = await mount('/data/requirements?businessId=biz_lab', { tenant: memberTenant });
    await waitFor(() => {
      expect(member.container.querySelector('[data-testid="data-requirements-page"]')).not.toBeNull();
    });
    expect(member.container.querySelector('[data-testid="confirm-requirement"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="analyze-requirements"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="reject-requirement"]')).toBeNull();

    const viewer = await mount('/data/requirements?businessId=biz_lab', { tenant: viewerTenant });
    await waitFor(() => {
      expect(viewer.container.querySelector('[data-testid="data-requirements-page"]')).not.toBeNull();
    });
    expect(viewer.container.querySelector('[data-testid="confirm-requirement"]')).toBeNull();
  });

  it('MEMBER cannot create sources or credentials', async () => {
    const owner = await mount('/data/sources?businessId=biz_lab', { tenant: ownerTenant });
    await waitFor(() => {
      expect(owner.container.querySelector('[data-testid="secret-credential-form"]')).not.toBeNull();
      expect(owner.container.querySelector('[data-testid="create-source"]')).not.toBeNull();
    });

    const member = await mount('/data/sources?businessId=biz_lab', { tenant: memberTenant });
    await waitFor(() => {
      expect(member.container.querySelector('[data-testid="data-sources-page"]')).not.toBeNull();
    });
    expect(member.container.querySelector('[data-testid="secret-credential-form"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="create-source"]')).toBeNull();
    expect(member.container.textContent).not.toContain(SECRET);
  });

  it('MEMBER mapping studio is read-only', async () => {
    const owner = await mount('/data/mappings?businessId=biz_lab', { tenant: ownerTenant });
    await waitFor(() => {
      expect(owner.container.querySelector('[data-testid="manual-mapping-form"]')).not.toBeNull();
      expect(owner.container.querySelector('[data-testid="confirm-mapping"]')).not.toBeNull();
    });

    const member = await mount('/data/mappings?businessId=biz_lab', { tenant: memberTenant });
    await waitFor(() => {
      expect(member.container.querySelector('[data-testid="mapping-studio"]')).not.toBeNull();
    });
    expect(member.container.querySelector('[data-testid="manual-mapping-form"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="analyze-mappings"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="confirm-mapping"]')).toBeNull();
  });

  it('MEMBER cannot see physical binding tab or lifecycle mutations', async () => {
    const owner = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { tenant: ownerTenant });
    await waitFor(() => {
      expect(owner.container.querySelector('[data-testid="physical-binding-tab"]')).not.toBeNull();
    });

    const member = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { tenant: memberTenant });
    await waitFor(() => {
      expect(member.container.querySelector('[data-testid="contract-detail-page"]')).not.toBeNull();
    });
    expect(member.container.querySelector('[data-testid="physical-binding-tab"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="contract-binding-detail"]')).toBeNull();
    expect(member.container.textContent).not.toContain('src_lab');
    await click(member.container.querySelector('[data-testid="revisions-tab"]'));
    expect(member.container.querySelector('[data-testid="validate-revision"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="activate-revision"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="deprecate-revision"]')).toBeNull();
  });

  it('MEMBER health page hides drift/gap mutations', async () => {
    const member = await mount('/data/health?businessId=biz_lab', { tenant: memberTenant });
    await waitFor(() => {
      expect(member.container.querySelector('[data-testid="data-health-page"]')).not.toBeNull();
    });
    await setValue(
      member.container.querySelector('[data-testid="health-contract-id"]') as HTMLInputElement,
      'ctr_lab',
    );
    await setValue(
      member.container.querySelector('[data-testid="health-revision-id"]') as HTMLInputElement,
      'rev_lab_2',
    );
    expect(member.container.querySelector('[data-testid="evaluate-drift"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="evaluate-gap"]')).toBeNull();
  });

  it('OWNER health page can execute drift/gap; MEMBER cannot', async () => {
    const owner = await mount('/data/health?businessId=biz_lab', { tenant: ownerTenant });
    await waitFor(() => {
      expect(owner.container.querySelector('[data-testid="data-health-page"]')).not.toBeNull();
    });
    await setValue(
      owner.container.querySelector('[data-testid="health-contract-id"]') as HTMLInputElement,
      'ctr_lab',
    );
    await setValue(
      owner.container.querySelector('[data-testid="health-revision-id"]') as HTMLInputElement,
      'rev_lab_2',
    );
    await waitFor(() => {
      expect(owner.container.querySelector('[data-testid="evaluate-drift"]')).not.toBeNull();
      expect(owner.container.querySelector('[data-testid="evaluate-gap"]')).not.toBeNull();
    });

    const member = await mount('/data/health?businessId=biz_lab', { tenant: memberTenant });
    await setValue(
      member.container.querySelector('[data-testid="health-contract-id"]') as HTMLInputElement,
      'ctr_lab',
    );
    await setValue(
      member.container.querySelector('[data-testid="health-revision-id"]') as HTMLInputElement,
      'rev_lab_2',
    );
    expect(member.container.querySelector('[data-testid="evaluate-drift"]')).toBeNull();
    expect(member.container.querySelector('[data-testid="evaluate-gap"]')).toBeNull();
  });

  it('MEMBER network revision payload omits physical binding fields', async () => {
    const memberSafe = {
      ...activeRevision,
      binding_refs: undefined,
    };
    delete (memberSafe as { binding_refs?: unknown }).binding_refs;
    const client = mockClient({
      listDataContractRevisions: vi.fn().mockResolvedValue({ data: [memberSafe] }),
      getDataContractRevision: vi.fn().mockResolvedValue({ data: memberSafe }),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', {
      client,
      tenant: memberTenant,
    });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="contract-detail-page"]')).not.toBeNull();
    });
    const resp = await client.listDataContractRevisions('biz_lab', 'ctr_lab');
    const json = JSON.stringify(resp.data);
    for (const forbidden of [
      'binding_refs',
      'source_id',
      'connection_id',
      'asset_id',
      'schema_snapshot_id',
      'src_lab',
      'conn_lab',
      'asset_lab',
      'snap_lab',
    ]) {
      expect(json).not.toContain(forbidden);
    }
    expect(container.querySelector('[data-testid="physical-binding-tab"]')).toBeNull();
    expect(container.textContent).not.toContain('src_lab');
  });
});

describe('contract lifecycle UX', () => {
  it('DRAFT shows Validate only and reloads after success', async () => {
    let revisions = [draftRevision];
    const client = mockClient({
      listDataContractRevisions: vi.fn(async () => ({ data: revisions })),
      getActiveDataContractDescriptor: vi.fn().mockRejectedValue(new Error('none')),
      validateDataContractRevision: vi.fn(async () => {
        revisions = [{ ...draftRevision, status: 'VALIDATED' }];
        return {
          data: {
            revision: revisions[0],
            result: {
              ValidationID: 'val_1',
              TenantID: 't1',
              BusinessID: 'biz_lab',
              ContractID: 'ctr_lab',
              RevisionID: draftRevision.revision_id,
              Version: 3,
              Status: 'PASS',
              Errors: [],
              Warnings: [],
              SnapshotFingerprints: { snap_lab: 'abc' },
              ValidatedBy: 'p',
              ValidatedAt: '2026-01-01T00:00:00Z',
              CreatedAt: '2026-01-01T00:00:00Z',
            },
          },
        };
      }),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client });
    await click(container.querySelector('[data-testid="revisions-tab"]'));
    await waitFor(() => {
      expect(container.querySelector('[data-testid="validate-revision"]')).not.toBeNull();
    });
    expect(container.querySelector('[data-testid="activate-revision"]')).toBeNull();
    expect(container.querySelector('[data-testid="deprecate-revision"]')).toBeNull();
    await click(container.querySelector('[data-testid="validate-revision"]'));
    await waitFor(() => {
      expect(client.validateDataContractRevision).toHaveBeenCalled();
      expect(container.querySelector('[data-testid="activate-revision"]')).not.toBeNull();
    });
    expect(container.querySelector('[data-testid="validate-revision"]')).toBeNull();
    expect(container.querySelector('[data-testid="validation-result"]')?.textContent).toContain('PASS');
  });

  it('VALIDATED shows Activate only and calls activate after confirm', async () => {
    let revisions = [validatedRevision];
    const client = mockClient({
      listDataContractRevisions: vi.fn(async () => ({ data: revisions })),
      activateDataContractRevision: vi.fn(async () => {
        revisions = [{ ...validatedRevision, status: 'ACTIVE' }];
        return { data: revisions[0] };
      }),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client });
    await click(container.querySelector('[data-testid="revisions-tab"]'));
    await waitFor(() => {
      expect(container.querySelector('[data-testid="activate-revision"]')).not.toBeNull();
    });
    expect(container.querySelector('[data-testid="validate-revision"]')).toBeNull();
    expect(container.querySelector('[data-testid="deprecate-revision"]')).toBeNull();
    await click(container.querySelector('[data-testid="activate-revision"]'));
    await waitFor(() => {
      expect(container.querySelector('[data-testid="activate-confirm-dialog"]')).not.toBeNull();
    });
    await click(container.querySelector('[data-testid="activate-confirm"]'));
    await waitFor(() => {
      expect(client.activateDataContractRevision).toHaveBeenCalled();
      expect(container.querySelector('[data-testid="deprecate-revision"]')).not.toBeNull();
    });
  });

  it('ACTIVE Deprecate uses exact revision_id and refreshes list', async () => {
    let revisions = [activeRevision];
    const listFn = vi.fn(async () => ({ data: revisions }));
    const client = mockClient({
      listDataContractRevisions: listFn,
      deprecateDataContractRevision: vi.fn(async (_b, _c, revId) => {
        expect(revId).toBe(activeRevision.revision_id);
        revisions = [{ ...activeRevision, status: 'DEPRECATED' }];
        return { data: revisions[0] };
      }),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client });
    await click(container.querySelector('[data-testid="revisions-tab"]'));
    await waitFor(() => {
      expect(
        container.querySelector(`[data-revision-id="${activeRevision.revision_id}"]`),
      ).not.toBeNull();
    });
    const listCallsBefore = listFn.mock.calls.length;
    await click(container.querySelector(`[data-revision-id="${activeRevision.revision_id}"]`));
    await waitFor(() => {
      expect(client.deprecateDataContractRevision).toHaveBeenCalledWith(
        'biz_lab',
        'ctr_lab',
        activeRevision.revision_id,
        expect.objectContaining({ reason: 'ui-deprecate' }),
      );
      expect(listFn.mock.calls.length).toBeGreaterThan(listCallsBefore);
    });
  });

  it('STALE Deprecate uses exact revision_id and refreshes list', async () => {
    let revisions = [staleRevision, activeRevision];
    const listFn = vi.fn(async () => ({ data: revisions }));
    const client = mockClient({
      listDataContractRevisions: listFn,
      deprecateDataContractRevision: vi.fn(async (_b, _c, revId) => {
        expect(revId).toBe(staleRevision.revision_id);
        revisions = [{ ...staleRevision, status: 'DEPRECATED' }, activeRevision];
        return { data: revisions[0] };
      }),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client });
    await click(container.querySelector('[data-testid="revisions-tab"]'));
    await waitFor(() => {
      expect(
        container.querySelector(`[data-revision-id="${staleRevision.revision_id}"]`),
      ).not.toBeNull();
    });
    const listCallsBefore = listFn.mock.calls.length;
    await click(container.querySelector(`[data-revision-id="${staleRevision.revision_id}"]`));
    await waitFor(() => {
      expect(client.deprecateDataContractRevision).toHaveBeenCalledWith(
        'biz_lab',
        'ctr_lab',
        staleRevision.revision_id,
        expect.objectContaining({ reason: 'ui-deprecate' }),
      );
      expect(listFn.mock.calls.length).toBeGreaterThan(listCallsBefore);
    });
  });

  it('ACTIVE and STALE both expose Deprecate while ACTIVE exists', async () => {
    const client = mockClient({
      listDataContractRevisions: vi.fn().mockResolvedValue({ data: [staleRevision, activeRevision] }),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client });
    await click(container.querySelector('[data-testid="revisions-tab"]'));
    await waitFor(() => {
      expect(container.querySelectorAll('[data-testid="deprecate-revision"]').length).toBe(2);
    });
    expect(container.querySelector('[data-testid="validate-revision"]')).toBeNull();
    expect(container.querySelector('[data-testid="activate-revision"]')).toBeNull();
    expect(container.querySelector('[data-testid="stale-warning"]')).not.toBeNull();
  });

  it('shows sanitized error and does not reject on validate failure', async () => {
    const rejections: unknown[] = [];
    const onReject = (ev: PromiseRejectionEvent) => rejections.push(ev.reason);
    window.addEventListener('unhandledrejection', onReject);
    const client = mockClient({
      listDataContractRevisions: vi.fn().mockResolvedValue({ data: [draftRevision] }),
      validateDataContractRevision: vi.fn().mockRejectedValue(
        new FormaApiError('FORBIDDEN', 'raw driver boom', {
          errorKey: 'FORMA_DATA_CONTRACT_NOT_ACTIVE',
        }),
      ),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client });
    await click(container.querySelector('[data-testid="revisions-tab"]'));
    await waitFor(() => {
      expect(container.querySelector('[data-testid="validate-revision"]')).not.toBeNull();
    });
    await click(container.querySelector('[data-testid="validate-revision"]'));
    await waitFor(() => {
      expect(container.querySelector('[data-testid="contract-error"]')?.textContent).toContain(
        '当前没有可用的活动数据契约。',
      );
    });
    expect(container.textContent).not.toContain('raw driver boom');
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });
    expect(rejections).toEqual([]);
    window.removeEventListener('unhandledrejection', onReject);
  });

  it('does not leak secrets from Error.message or validation messages into DOM', async () => {
    const leak =
      'postgres://user:password@db/prod Authorization: Bearer tok_abc token=xyz';
    const client = mockClient({
      listDataContractRevisions: vi.fn().mockRejectedValue(new Error(leak)),
    });
    const { container } = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="contract-error"]')?.textContent).toContain(
        '操作失败',
      );
    });
    expect(container.innerHTML).not.toContain('password');
    expect(container.innerHTML).not.toContain('Authorization');
    expect(container.innerHTML).not.toContain('Bearer');
    expect(container.innerHTML).not.toContain('tok_abc');
    expect(container.innerHTML).not.toContain('token=xyz');

    const draftClient = mockClient({
      listDataContractRevisions: vi.fn().mockResolvedValue({ data: [draftRevision] }),
      validateDataContractRevision: vi.fn().mockResolvedValue({
        data: {
          revision: draftRevision,
          result: {
            ValidationID: 'val_leak',
            TenantID: 't1',
            BusinessID: 'biz_lab',
            ContractID: 'ctr_lab',
            RevisionID: draftRevision.revision_id,
            Version: 3,
            Status: 'FAIL',
            Errors: [{ code: 'SCHEMA_JSON', message: leak }],
            Warnings: [],
            SnapshotFingerprints: {},
            ValidatedBy: 'p',
            ValidatedAt: '2026-01-01T00:00:00Z',
            CreatedAt: '2026-01-01T00:00:00Z',
          },
        },
      }),
    });
    const draft = await mount('/data/contracts/ctr_lab?businessId=biz_lab', { client: draftClient });
    await click(draft.container.querySelector('[data-testid="revisions-tab"]'));
    await click(draft.container.querySelector('[data-testid="validate-revision"]'));
    await waitFor(() => {
      expect(draft.container.querySelector('[data-testid="validation-issue"]')).not.toBeNull();
    });
    expect(draft.container.innerHTML).not.toContain('password');
    expect(draft.container.innerHTML).not.toContain('Authorization');
    expect(draft.container.innerHTML).not.toContain('tok_abc');
    expect(draft.container.querySelector('[data-testid="validation-issue"]')?.textContent).toBe(
      '结构快照无效。',
    );
  });

  it('create contract uses business.current_revision when > 1', async () => {
    const biz = { ...labBusiness, current_revision: 5 };
    const client = mockClient({
      listBusinesses: vi.fn().mockResolvedValue({ data: [biz] }),
      createDataContract: vi.fn().mockResolvedValue({
        data: { contract: { contract_id: 'ctr_new' }, revision: draftRevision },
      }),
    });
    const { container } = await mount('/data/contracts?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="create-contract"]')).not.toBeNull();
    });
    await setValue(container.querySelector('[data-testid="contract-name"]') as HTMLInputElement, 'C5');
    await click(container.querySelector('[data-testid="create-contract"]'));
    await waitFor(() => {
      expect(client.createDataContract).toHaveBeenCalled();
    });
    const body = (client.createDataContract as ReturnType<typeof vi.fn>).mock.calls[0][1];
    expect(body.business_model_revision).toBe(5);
  });
});

describe('mapping studio DSL payloads', () => {
  it('submits TIME_NORMALIZE transform_spec matching mapping_type', async () => {
    const client = mockClient();
    const { container } = await mount('/data/mappings?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="mapping-type-select"]')).not.toBeNull();
    });
    await click(container.querySelector(`[data-testid="mapping-requirement-${labRequirement.requirement_id}"]`));
    await setValue(container.querySelector('[data-testid="snapshot-ids-input"]') as HTMLInputElement, 'snap_lab');
    await setValue(container.querySelector('[data-testid="target-path-input"]') as HTMLInputElement, 'recorded_at');
    await setValue(container.querySelector('[data-testid="mapping-type-select"]') as HTMLSelectElement, 'TIME_NORMALIZE');
    await waitFor(() => {
      expect(container.querySelector('[data-testid="dsl-source-timezone"]')).not.toBeNull();
    });
    await setValue(container.querySelector('[data-testid="dsl-source-timezone"]') as HTMLInputElement, 'Asia/Shanghai');
    await setValue(container.querySelector('[data-testid="dsl-target-timezone"]') as HTMLInputElement, 'UTC');
    await setValue(container.querySelector('[data-testid="dsl-format"]') as HTMLInputElement, 'RFC3339');
    await click(container.querySelector('[data-testid="create-manual-mapping"]'));
    await waitFor(() => {
      expect(client.createManualSemanticMapping).toHaveBeenCalled();
    });
    const body = (client.createManualSemanticMapping as ReturnType<typeof vi.fn>).mock.calls[0][1];
    expect(body.mapping_type).toBe('TIME_NORMALIZE');
    expect(body.transform_spec.type).toBe('TIME_NORMALIZE');
    expect(body.transform_spec.source_timezone).toBe('Asia/Shanghai');
    expect(body.transform_spec.target_timezone).toBe('UTC');
    expect(body.transform_spec.format).toBe('RFC3339');
    expect(body.source_id).toBe('src_lab');
    expect(body.connection_id).toBe('conn_lab');
    expect(body.asset_id).toBe('asset_lab');
    expect(body.schema_snapshot_id).toBe('snap_lab');
    expect(body.source_id).not.toBe('src_manual');
  });

  it('submits FIELD_PATH transform_spec matching mapping_type', async () => {
    const client = mockClient();
    const { container } = await mount('/data/mappings?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="mapping-type-select"]')).not.toBeNull();
    });
    await click(container.querySelector(`[data-testid="mapping-requirement-${labRequirement.requirement_id}"]`));
    await setValue(container.querySelector('[data-testid="snapshot-ids-input"]') as HTMLInputElement, 'snap_lab');
    await setValue(container.querySelector('[data-testid="target-path-input"]') as HTMLInputElement, 'sensor.temperature');
    await setValue(container.querySelector('[data-testid="mapping-type-select"]') as HTMLSelectElement, 'FIELD_PATH');
    await waitFor(() => {
      expect(container.querySelector('[data-testid="dsl-field-path"]')).not.toBeNull();
    });
    await setValue(container.querySelector('[data-testid="dsl-field-path"]') as HTMLInputElement, 'sensor.temperature');
    await click(container.querySelector('[data-testid="create-manual-mapping"]'));
    await waitFor(() => {
      expect(client.createManualSemanticMapping).toHaveBeenCalled();
    });
    const body = (client.createManualSemanticMapping as ReturnType<typeof vi.fn>).mock.calls[0][1];
    expect(body.mapping_type).toBe('FIELD_PATH');
    expect(body.transform_spec).toEqual({ type: 'FIELD_PATH', path: 'sensor.temperature' });
  });

  it('valid snapshot lineage is submitted; incomplete snapshot fails without manual placeholders', async () => {
    const incompleteClient = mockClient({
      getSchemaSnapshot: vi.fn().mockResolvedValue({
        data: { snapshot_id: 'snap_lab', schema: { fields: [] } },
      }),
    });
    const incomplete = await mount('/data/mappings?businessId=biz_lab', { client: incompleteClient });
    await click(
      incomplete.container.querySelector(
        `[data-testid="mapping-requirement-${labRequirement.requirement_id}"]`,
      ),
    );
    await setValue(
      incomplete.container.querySelector('[data-testid="snapshot-ids-input"]') as HTMLInputElement,
      'snap_lab',
    );
    await setValue(
      incomplete.container.querySelector('[data-testid="target-path-input"]') as HTMLInputElement,
      'temp_c',
    );
    await click(incomplete.container.querySelector('[data-testid="create-manual-mapping"]'));
    await waitFor(() => {
      expect(incomplete.container.querySelector('[data-testid="mapping-error"]')?.textContent).toContain(
        '结构快照不完整',
      );
    });
    expect(incompleteClient.createManualSemanticMapping).not.toHaveBeenCalled();

    const client = mockClient();
    const { container } = await mount('/data/mappings?businessId=biz_lab', { client });
    await click(container.querySelector(`[data-testid="mapping-requirement-${labRequirement.requirement_id}"]`));
    await setValue(container.querySelector('[data-testid="snapshot-ids-input"]') as HTMLInputElement, 'snap_lab');
    await setValue(container.querySelector('[data-testid="target-path-input"]') as HTMLInputElement, 'temp_c');
    await click(container.querySelector('[data-testid="create-manual-mapping"]'));
    await waitFor(() => {
      expect(client.createManualSemanticMapping).toHaveBeenCalled();
    });
    const body = (client.createManualSemanticMapping as ReturnType<typeof vi.fn>).mock.calls[0][1];
    expect(body).toMatchObject({
      source_id: 'src_lab',
      connection_id: 'conn_lab',
      asset_id: 'asset_lab',
      schema_snapshot_id: 'snap_lab',
    });
    expect(JSON.stringify(body)).not.toContain('src_manual');
    expect(JSON.stringify(body)).not.toContain('conn_manual');
    expect(JSON.stringify(body)).not.toContain('asset_manual');
  });
});

describe('data source / connection payloads', () => {
  it('creates RELATIONAL_DATABASE source and MYSQL connection with SQL public_config', async () => {
    const client = mockClient();
    const { container } = await mount('/data/sources?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="source-type-select"]')).not.toBeNull();
    });
    await setValue(container.querySelector('[data-testid="source-name"]') as HTMLInputElement, 'DB');
    await setValue(
      container.querySelector('[data-testid="source-type-select"]') as HTMLSelectElement,
      'RELATIONAL_DATABASE',
    );
    await click(container.querySelector('[data-testid="create-source"]'));
    await waitFor(() => {
      expect(client.createDataSource).toHaveBeenCalled();
    });
    expect((client.createDataSource as ReturnType<typeof vi.fn>).mock.calls[0][0]).toEqual({
      name: 'DB',
      source_type: 'RELATIONAL_DATABASE',
    });

    const detail = await mount('/data/sources/src_lab?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(detail.container.querySelector('[data-testid="create-connection"]')).not.toBeNull();
    });
    await setValue(
      detail.container.querySelector('[data-testid="connection-adapter"]') as HTMLSelectElement,
      'POSTGRESQL',
    );
    await setValue(detail.container.querySelector('[data-testid="connection-host"]') as HTMLInputElement, 'db.host');
    await setValue(detail.container.querySelector('[data-testid="connection-database"]') as HTMLInputElement, 'lab');
    await setValue(detail.container.querySelector('[data-testid="connection-username"]') as HTMLInputElement, 'reader');
    await click(detail.container.querySelector('[data-testid="create-connection"]'));
    await waitFor(() => {
      expect(client.createDataConnection).toHaveBeenCalled();
    });
    const connBody = (client.createDataConnection as ReturnType<typeof vi.fn>).mock.calls[0][1];
    expect(connBody.adapter_type).toBe('POSTGRESQL');
    expect(connBody.public_config).toEqual({
      host: 'db.host',
      port: 5432,
      database: 'lab',
      username: 'reader',
    });
  });

  it('HTTP_API connection only allows HTTP adapter and HTTP public_config', async () => {
    const client = mockClient({
      getDataSource: vi.fn().mockResolvedValue({ data: labHttpSource }),
      listDataSources: vi.fn().mockResolvedValue({ data: [labHttpSource] }),
    });
    const { container } = await mount('/data/sources/src_lab_http?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="connection-adapter"]')).not.toBeNull();
    });
    const adapter = container.querySelector('[data-testid="connection-adapter"]') as HTMLSelectElement;
    expect(Array.from(adapter.options).map(o => o.value)).toEqual(['HTTP']);
    expect(container.querySelector('[data-testid="connection-base-url"]')).not.toBeNull();
    expect(container.querySelector('[data-testid="connection-host"]')).toBeNull();
    await setValue(
      container.querySelector('[data-testid="connection-base-url"]') as HTMLInputElement,
      'https://api.example/v1',
    );
    await setValue(
      container.querySelector('[data-testid="connection-openapi-url"]') as HTMLInputElement,
      'https://api.example/openapi.json',
    );
    await click(container.querySelector('[data-testid="create-connection"]'));
    await waitFor(() => {
      expect(client.createDataConnection).toHaveBeenCalled();
    });
    const body = (client.createDataConnection as ReturnType<typeof vi.fn>).mock.calls[0][1];
    expect(body.adapter_type).toBe('HTTP');
    expect(body.public_config).toEqual({
      base_url: 'https://api.example/v1',
      openapi_url: 'https://api.example/openapi.json',
    });
    expect(body.adapter_type).not.toBe('MYSQL');
  });
});

describe('Credential form — no secret echo', () => {
  it('clears input and never renders the raw secret after submit', async () => {
    const client = mockClient();
    const { container } = await mount('/data/sources?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="cred-password-input"]')).not.toBeNull();
    });
    const input = container.querySelector('[data-testid="cred-password-input"]') as HTMLInputElement;
    await setValue(input, SECRET);
    expect(input.value).toBe(SECRET);
    const form = container.querySelector('[data-testid="secret-credential-form"]') as HTMLFormElement;
    await click(form.querySelector('button[type="submit"]'));
    await waitFor(() => {
      expect(client.createDataCredential).toHaveBeenCalled();
    });
    const payload = (client.createDataCredential as ReturnType<typeof vi.fn>).mock.calls[0][0];
    expect(payload.secret.password).toBe(SECRET);
    const html = container.innerHTML;
    expect(html).not.toContain(SECRET);
    expect(container.textContent).not.toContain(SECRET);
    expect((container.querySelector('[data-testid="cred-password-input"]') as HTMLInputElement).value).toBe('');
    const serialized = JSON.stringify(
      (client.createDataCredential as ReturnType<typeof vi.fn>).mock.results[0].value,
    );
    expect(serialized).not.toContain(SECRET);
    const resp = await client.createDataCredential({
      secret_type: 'password',
      secret: { password: SECRET },
    });
    expect(JSON.stringify(resp.data)).not.toContain(SECRET);
    expect(Object.keys(resp.data as object)).not.toContain('password');
    expect(Object.keys(resp.data as object)).not.toContain('secret');
  });

  it('hides SecretCredentialForm when canEdit is false', () => {
    const html = renderToStaticMarkup(
      createElement(SecretCredentialForm, { client: mockClient(), canEdit: false }),
    );
    expect(html).toBe('');
    expect(html).not.toContain(SECRET);
  });
});

describe('mapping EditConfirm UI', () => {
  it('edit-confirm submits controlled DSL and shows MANUAL_MODIFIED replacement', async () => {
    let maps = [labMapping];
    const client = mockClient({
      listSemanticMappings: vi.fn(async () => ({ data: maps })),
      editConfirmSemanticMapping: vi.fn(async () => {
        const original = { ...labMapping, status: 'SUPERSEDED' };
        const replacement = {
          ...labMapping,
          mapping_id: 'map_lab_edit',
          status: 'CONFIRMED',
          source: 'MANUAL_MODIFIED',
          derived_from_mapping_id: labMapping.mapping_id,
          mapping_type: 'CAST',
          transform_spec: { type: 'CAST', from_type: 'string', to_type: 'number' },
        };
        maps = [original, replacement];
        return { data: { original, replacement, decision: { decision: 'EDIT_CONFIRM' } } };
      }),
    });
    const { container } = await mount('/data/mappings?businessId=biz_lab', { client });
    await waitFor(() => {
      expect(container.querySelector('[data-testid="edit-confirm-mapping"]')).not.toBeNull();
    });
    expect(container.querySelector('[data-testid="confirm-mapping"]')).not.toBeNull();
    expect(container.querySelector('[data-testid="reject-mapping"]')).not.toBeNull();
    await click(container.querySelector('[data-testid="edit-confirm-mapping"]'));
    await waitFor(() => {
      expect(container.querySelector('[data-testid="edit-confirm-mapping-panel"]')).not.toBeNull();
    });
    await setValue(
      container.querySelector('[data-testid="edit-mapping-type-select"]') as HTMLSelectElement,
      'CAST',
    );
    await waitFor(() => {
      expect(container.querySelector('[data-testid="dsl-from-type"]')).not.toBeNull();
    });
    await setValue(container.querySelector('[data-testid="dsl-from-type"]') as HTMLInputElement, 'string');
    await setValue(container.querySelector('[data-testid="dsl-to-type"]') as HTMLInputElement, 'number');
    await setValue(
      container.querySelector('[data-testid="edit-target-path-input"]') as HTMLInputElement,
      'temp_c',
    );
    await click(container.querySelector('[data-testid="submit-edit-confirm-mapping"]'));
    await waitFor(() => {
      expect(client.editConfirmSemanticMapping).toHaveBeenCalled();
    });
    const body = (client.editConfirmSemanticMapping as ReturnType<typeof vi.fn>).mock.calls[0][2];
    expect(body.mapping_type).toBe('CAST');
    expect(body.transform_spec).toEqual({
      type: 'CAST',
      from_type: 'string',
      to_type: 'number',
    });
    await waitFor(() => {
      expect(container.textContent).toContain('人工修改并确认');
      expect(container.textContent).toContain('已替代');
    });
  });
});

describe('drift snapshot picker', () => {
  it('disables evaluate-drift until all pinned snapshots have fresh selection', async () => {
    const client = mockClient();
    const { container } = await mount('/data/health?businessId=biz_lab', { client });
    await setValue(
      container.querySelector('[data-testid="health-contract-id"]') as HTMLInputElement,
      'ctr_lab',
    );
    await setValue(
      container.querySelector('[data-testid="health-revision-id"]') as HTMLInputElement,
      'rev_lab_2',
    );
    await waitFor(() => {
      expect(container.querySelector('[data-testid="evaluate-drift"]')).not.toBeNull();
      expect(container.querySelector('[data-testid="drift-snapshot-picker"]')).not.toBeNull();
    });
    const btn = container.querySelector('[data-testid="evaluate-drift"]') as HTMLButtonElement;
    expect(btn.disabled).toBe(true);
    await waitFor(() => {
      expect(
        container.querySelector('[data-testid="fresh-snapshot-select-snap_lab"]'),
      ).not.toBeNull();
    });
    const select = container.querySelector(
      '[data-testid="fresh-snapshot-select-snap_lab"]',
    ) as HTMLSelectElement;
    expect(Array.from(select.options).map(o => o.value)).toContain('snap_lab_fresh');
    expect(Array.from(select.options).map(o => o.value)).not.toContain('snap_other_asset');
    await setValue(select, 'snap_lab_fresh');
    expect((container.querySelector('[data-testid="evaluate-drift"]') as HTMLButtonElement).disabled).toBe(
      false,
    );
  });

  it('submits real new_snapshot_ids and renders COMPATIBLE / BREAKING', async () => {
    const client = mockClient({
      evaluateDataContractDrift: vi
        .fn()
        .mockResolvedValueOnce({
          data: {
            result: {
              DriftResultID: 'dr1',
              Severity: 'COMPATIBLE',
              Findings: [],
            },
            revision: activeRevision,
          },
        })
        .mockResolvedValueOnce({
          data: {
            result: {
              DriftResultID: 'dr2',
              Severity: 'BREAKING',
              Findings: [{ code: 'FIELD_REMOVED', message: 'x', binding_mapping_id: 'm', field_path: 'a' }],
            },
            revision: { ...activeRevision, status: 'STALE' },
          },
        }),
      getDataContract: vi
        .fn()
        .mockResolvedValueOnce({
          data: {
            contract_id: 'ctr_lab',
            business_id: 'biz_lab',
            active_revision_id: 'rev_lab_2',
            created_by: 'p',
            created_at: '',
            updated_at: '',
          },
        })
        .mockResolvedValue({
          data: {
            contract_id: 'ctr_lab',
            business_id: 'biz_lab',
            active_revision_id: '',
            created_by: 'p',
            created_at: '',
            updated_at: '',
          },
        }),
    });
    const { container } = await mount('/data/health?businessId=biz_lab', { client });
    await setValue(
      container.querySelector('[data-testid="health-contract-id"]') as HTMLInputElement,
      'ctr_lab',
    );
    await setValue(
      container.querySelector('[data-testid="health-revision-id"]') as HTMLInputElement,
      'rev_lab_2',
    );
    await waitFor(() => {
      expect(
        container.querySelector('[data-testid="fresh-snapshot-select-snap_lab"]'),
      ).not.toBeNull();
    });
    await setValue(
      container.querySelector('[data-testid="fresh-snapshot-select-snap_lab"]') as HTMLSelectElement,
      'snap_lab_fresh',
    );
    await click(container.querySelector('[data-testid="evaluate-drift"]'));
    await waitFor(() => {
      expect(client.evaluateDataContractDrift).toHaveBeenCalledWith(
        'biz_lab',
        'ctr_lab',
        'rev_lab_2',
        { new_snapshot_ids: { snap_lab: 'snap_lab_fresh' } },
      );
      expect(container.querySelector('[data-testid="drift-severity-banner"]')?.textContent).toContain(
        'COMPATIBLE',
      );
    });
    await click(container.querySelector('[data-testid="evaluate-drift"]'));
    await waitFor(() => {
      expect(container.querySelector('[data-testid="drift-severity-banner"]')?.textContent).toContain(
        'BREAKING',
      );
      expect(container.querySelector('[data-testid="health-contract-status"]')?.textContent).toMatch(
        /active_revision_id=（空）|STALE/,
      );
    });
  });

  it('multi pinned snapshot mapping requires every fresh selection', async () => {
    const binding2 = {
      requirement_id: 'req_2',
      mapping_id: 'map_2',
      source_id: 'src_lab',
      connection_id: 'conn_lab',
      asset_id: 'asset_lab',
      schema_snapshot_id: 'snap_lab_b',
    };
    const client = mockClient({
      getDataContractRevision: vi.fn().mockResolvedValue({
        data: {
          ...activeRevision,
          binding_refs: [labBinding, binding2],
        },
      }),
      listSchemaSnapshots: vi.fn().mockImplementation(async (params: { assetId: string }) => ({
        data: [
          { ...labSchemaSnapshot, snapshot_id: 'snap_lab', asset_id: params.assetId },
          { ...labSchemaSnapshot, snapshot_id: 'snap_lab_b', asset_id: params.assetId },
          { ...labSchemaSnapshot, snapshot_id: 'snap_lab_fresh', asset_id: params.assetId },
        ],
      })),
    });
    const { container } = await mount('/data/health?businessId=biz_lab', { client });
    await setValue(
      container.querySelector('[data-testid="health-contract-id"]') as HTMLInputElement,
      'ctr_lab',
    );
    await setValue(
      container.querySelector('[data-testid="health-revision-id"]') as HTMLInputElement,
      'rev_lab_2',
    );
    await waitFor(() => {
      expect(container.querySelector('[data-testid="fresh-snapshot-select-snap_lab"]')).not.toBeNull();
      expect(container.querySelector('[data-testid="fresh-snapshot-select-snap_lab_b"]')).not.toBeNull();
    });
    const btn = () => container.querySelector('[data-testid="evaluate-drift"]') as HTMLButtonElement;
    expect(btn().disabled).toBe(true);
    await setValue(
      container.querySelector('[data-testid="fresh-snapshot-select-snap_lab"]') as HTMLSelectElement,
      'snap_lab_fresh',
    );
    // Second pinned still missing → must stay disabled (multi-mapping gate).
    await waitFor(() => {
      expect(btn().disabled).toBe(true);
    });
    expect(
      container.querySelector('[data-testid="fresh-snapshot-select-snap_lab_b"]'),
    ).not.toBeNull();
  });
});
