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
});
