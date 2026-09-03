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
  labDescriptor,
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

  it('ACTIVE and STALE show Deprecate; historical STALE can deprecate while ACTIVE exists', async () => {
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
    await click(container.querySelector('[data-testid="deprecate-revision"]'));
    await waitFor(() => {
      expect(client.deprecateDataContractRevision).toHaveBeenCalled();
    });
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
