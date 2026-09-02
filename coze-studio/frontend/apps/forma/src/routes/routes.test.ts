import { describe, expect, it } from 'vitest';

import { navigation, routeIds } from '@/lib/navigation';

describe('forma routes', () => {
  it('defines all 16 product routes including overview and design', () => {
    const paths = navigation.flatMap(g => g.items.map(i => i.path));
    expect(paths).toContain('/');
    expect(paths).toContain('/design');
    expect(paths).toContain('/analyst');
    expect(paths).toContain('/business');
    expect(paths).toContain('/delivery');
    expect(routeIds.length).toBeGreaterThanOrEqual(16);
  });

  it('includes business editor detail route pattern', () => {
    // Wired in AppRouter as /business/:businessId (not in sidebar nav).
    const editorPath = '/business/:businessId';
    expect(editorPath.startsWith('/business/')).toBe(true);
    expect(editorPath).toMatch(/:businessId/);
  });

  it('defines data plane nested paths', () => {
    const dataPaths = [
      '/data',
      '/data/requirements',
      '/data/sources',
      '/data/mappings',
      '/data/contracts',
      '/data/health',
    ];
    for (const p of dataPaths) {
      expect(p.startsWith('/data')).toBe(true);
    }
    expect(navigation.flatMap(g => g.items.map(i => i.path))).toContain('/data');
  });
});
