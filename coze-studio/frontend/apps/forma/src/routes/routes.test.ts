import { afterEach, describe, expect, it, vi } from 'vitest';
import { createElement } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { act } from 'react-dom/test-utils';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import { DataPlaneApp } from '@forma/data';
import type { FormaApiClient, FormaTenant } from '@forma/api-client';

const ownerTenant: FormaTenant = {
  tenant_id: 't1',
  tenant_key: 't1',
  name: 'T1',
  display_name: 'T1',
  status: 'ACTIVE',
  revision: 1,
  role: 'OWNER',
};

function mockClient(): FormaApiClient {
  return new Proxy(
    {},
    {
      get: (_target, prop) => {
        if (prop === 'then') return undefined;
        if (prop === 'getDataSource') {
          return vi.fn().mockResolvedValue({
            data: {
              source_id: 'src_lab',
              name: 'Lab source',
              source_type: 'EXTERNAL_SQL',
              status: 'ACTIVE',
              created_by: 'p',
              created_at: '',
              updated_at: '',
            },
          });
        }
        return vi.fn().mockResolvedValue({ data: [] });
      },
    },
  ) as unknown as FormaApiClient;
}

vi.mock('@/hooks/use-forma-session', () => {
  const tenant = {
    tenant_id: 't1',
    tenant_key: 't1',
    name: 'T1',
    display_name: 'T1',
    status: 'ACTIVE',
    revision: 1,
    role: 'OWNER',
  };
  const client = new Proxy(
    {},
    {
      get: (_target, prop) => {
        if (prop === 'then') return undefined;
        if (prop === 'getDataSource') {
          return vi.fn().mockResolvedValue({
            data: {
              source_id: 'src_lab',
              name: 'Lab source',
              source_type: 'EXTERNAL_SQL',
              status: 'ACTIVE',
              created_by: 'p',
              created_at: '',
              updated_at: '',
            },
          });
        }
        return vi.fn().mockResolvedValue({ data: [] });
      },
    },
  );
  return {
    useFormaSession: () => ({
      state: 'ready',
      error: null,
      me: {
        principal: {
          principal_id: 'p1',
          principal_type: 'user',
          display_name: 'Owner',
          coze_user_id: '1',
          status: 'ACTIVE',
        },
        tenants: [tenant],
        current_tenant: tenant,
      },
      tenants: [tenant],
      currentTenant: tenant,
      assetCounts: { business: 0, capability: 0, agent: 0, application: 0 },
      client,
      switchTenant: vi.fn(),
      refresh: vi.fn(),
      bootstrap: vi.fn(),
    }),
  };
});

import { AppRouter } from './index';
import { navigation, routeIds } from '../lib/navigation';

const trees: Array<{ root: Root; node: HTMLElement }> = [];

async function renderAt(ui: ReturnType<typeof createElement>, path: string) {
  const node = document.createElement('div');
  document.body.appendChild(node);
  const root = createRoot(node);
  trees.push({ root, node });
  await act(async () => {
    root.render(
      createElement(
        MemoryRouter,
        {
          initialEntries: [path],
          future: { v7_startTransition: true, v7_relativeSplatPath: true },
        },
        ui,
      ),
    );
  });
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
  return node;
}

afterEach(() => {
  while (trees.length) {
    const t = trees.pop();
    if (!t) continue;
    act(() => t.root.unmount());
    t.node.remove();
  }
});

describe('forma routes', () => {
  it('keeps product navigation including business, analyst, and data', () => {
    const paths = navigation.flatMap(g => g.items.map(i => i.path));
    expect(paths).toContain('/');
    expect(paths).toContain('/design');
    expect(paths).toContain('/analyst');
    expect(paths).toContain('/business');
    expect(paths).toContain('/data');
    expect(paths).toContain('/delivery');
    expect(routeIds.length).toBeGreaterThanOrEqual(16);
  });

  it('renders /business and /analyst through AppRouter', async () => {
    const business = await renderAt(createElement(AppRouter), '/business');
    expect(business.querySelector('.forma-shell')).not.toBeNull();
    expect(business.textContent).toMatch(/业务/);

    const analyst = await renderAt(createElement(AppRouter), '/analyst');
    expect(analyst.querySelector('.forma-shell')).not.toBeNull();
    expect(analyst.textContent).toMatch(/分析/);
  });
});

describe('data plane nested routes', () => {
  const cases = [
    '/data',
    '/data/requirements',
    '/data/sources',
    '/data/sources/src_lab',
    '/data/mappings',
    '/data/contracts',
    '/data/contracts/ctr_lab',
    '/data/health',
  ];

  it.each(cases)('AppRouter reaches %s', async path => {
    const node = await renderAt(createElement(AppRouter), path);
    expect(node.querySelector('[data-testid="data-plane-shell"]')).not.toBeNull();
  });

  it.each(cases)('DataPlaneApp MemoryRouter reaches %s', async path => {
    const node = await renderAt(
      createElement(
        Routes,
        null,
        createElement(Route, {
          path: '/data/*',
          element: createElement(DataPlaneApp, {
            client: mockClient(),
            currentTenant: ownerTenant,
          }),
        }),
      ),
      path,
    );
    expect(node.querySelector('[data-testid="data-plane-shell"]')).not.toBeNull();
  });
});
