import { afterEach, describe, expect, it, vi } from 'vitest';
import { createElement } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { act } from 'react-dom/test-utils';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

const sessionState = vi.hoisted(() => ({
  state: 'unauthenticated' as string,
}));

vi.mock('@/hooks/use-forma-session', () => ({
  useFormaSession: () => ({
    state: sessionState.state,
    error: null,
    me: null,
    tenants: [],
    currentTenant: null,
    assetCounts: null,
    client: {},
    switchTenant: vi.fn(),
    refresh: vi.fn(),
    bootstrap: vi.fn(),
    clearLocalSession: vi.fn(),
  }),
}));

import { FormaAuthGuard } from '../components/FormaAuthGuard';
import { AppRouter } from '../routes';

const trees: Array<{ root: Root; node: HTMLElement }> = [];

async function renderAt(path: string, ui: ReturnType<typeof createElement>) {
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
  });
  return node;
}

afterEach(() => {
  sessionState.state = 'unauthenticated';
  while (trees.length) {
    const t = trees.pop();
    if (!t) continue;
    act(() => t.root.unmount());
    t.node.remove();
  }
});

describe('FormaAuthGuard', () => {
  it('unauthenticated does not render AppShell children', async () => {
    sessionState.state = 'unauthenticated';
    const node = await renderAt(
      '/data',
      createElement(
        Routes,
        null,
        createElement(Route, {
          path: '/login',
          element: createElement('div', { 'data-testid': 'login-stub' }, 'login'),
        }),
        createElement(Route, {
          path: '/*',
          element: createElement(
            FormaAuthGuard,
            null,
            createElement('div', { 'data-testid': 'forma-app-shell' }, 'shell'),
          ),
        }),
      ),
    );
    expect(node.querySelector('[data-testid="forma-app-shell"]')).toBeNull();
    expect(node.querySelector('[data-testid="login-stub"]')).not.toBeNull();
  });

  it('ready renders protected children', async () => {
    sessionState.state = 'ready';
    const node = await renderAt(
      '/',
      createElement(
        FormaAuthGuard,
        null,
        createElement('div', { 'data-testid': 'forma-app-shell' }, 'shell'),
      ),
    );
    expect(node.querySelector('[data-testid="forma-app-shell"]')).not.toBeNull();
  });

  it('empty redirects to onboarding', async () => {
    sessionState.state = 'empty';
    const node = await renderAt(
      '/',
      createElement(
        Routes,
        null,
        createElement(Route, {
          path: '/onboarding',
          element: createElement('div', { 'data-testid': 'onboarding-stub' }, 'onboarding'),
        }),
        createElement(Route, {
          path: '/*',
          element: createElement(FormaAuthGuard, null, createElement('div', null, 'x')),
        }),
      ),
    );
    expect(node.querySelector('[data-testid="onboarding-stub"]')).not.toBeNull();
  });
});

describe('AppRouter auth routes', () => {
  it('exposes /login without AppShell when unauthenticated', async () => {
    sessionState.state = 'unauthenticated';
    const node = await renderAt('/login', createElement(AppRouter));
    expect(node.querySelector('[data-testid="forma-login-page"]')).not.toBeNull();
    expect(node.querySelector('[data-testid="forma-app-shell"]')).toBeNull();
    expect(node.querySelector('[data-testid="forma-sidebar"]')).toBeNull();
    expect(node.textContent).not.toContain('Coze Session');
  });
});
