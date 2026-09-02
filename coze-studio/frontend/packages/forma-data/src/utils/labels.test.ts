import { describe, expect, it } from 'vitest';

import { readinessLabel, statusLabel } from '../utils/labels';
import { isEditor } from '../utils/roles';
import * as labelsMod from '../utils/labels';

describe('production utils domain-agnostic', () => {
  it('labels module has no work_order switch', () => {
    const src = `${statusLabel.toString()}${readinessLabel.toString()}${Object.keys(labelsMod).join(',')}`;
    expect(src).not.toMatch(/work_order/);
  });

  it('roles only check OWNER|ADMIN', () => {
    expect(isEditor('OWNER')).toBe(true);
    expect(isEditor('ADMIN')).toBe(true);
    expect(isEditor('MEMBER')).toBe(false);
    expect(isEditor('VIEWER')).toBe(false);
  });
});
