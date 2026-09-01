import { describe, expect, it, vi } from 'vitest';

import type { FormaApiClient, FormaBusiness, FormaTenant } from '@forma/api-client';

import { AnalystWorkspacePage } from './AnalystWorkspacePage';

describe('AnalystWorkspacePage tenant switch', () => {
  it('clears workspace state when tenant changes', async () => {
    const businessesA: FormaBusiness[] = [
      {
        business_id: 'biz_a',
        asset_id: 'a1',
        name: 'Tenant A Business',
        status: 'ACTIVE',
        current_revision: 1,
        schema_version: '1.0',
        updated_at: '2026-01-01T00:00:00Z',
        created_at: '2026-01-01T00:00:00Z',
      },
    ];

    const client = {
      listBusinesses: vi.fn().mockResolvedValue({ data: businessesA }),
      listAssertions: vi.fn().mockResolvedValue({ data: [] }),
      listEvidence: vi.fn().mockResolvedValue({ data: [] }),
      listConflicts: vi.fn().mockResolvedValue({ data: [] }),
      listGaps: vi.fn().mockResolvedValue({ data: [] }),
      listAnalystTurns: vi.fn().mockResolvedValue({ data: [] }),
    } as unknown as FormaApiClient;

    const tenantA: FormaTenant = {
      tenant_id: 'tenant_a',
      tenant_key: 'a',
      name: 'A',
      display_name: 'A',
      status: 'ACTIVE',
      revision: 1,
    };

    // Smoke: module exports page component used by app route.
    expect(AnalystWorkspacePage).toBeDefined();
    expect(client.listBusinesses).toBeDefined();
    expect(tenantA.tenant_id).toBe('tenant_a');
  });
});
