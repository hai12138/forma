import { describe, expect, it } from 'vitest';

import { FormaApiError } from '@forma/api-client';

import { sanitizedErrorMessage } from './errors';

describe('sanitizedErrorMessage', () => {
  it('maps stable contract/mapping keys to product copy', () => {
    expect(
      sanitizedErrorMessage(
        new FormaApiError('HTTP_ERROR', 'hidden internals', {
          errorKey: 'FORMA_DATA_CONTRACT_NOT_ACTIVE',
        }),
      ),
    ).toBe('当前没有可用的活动数据契约。');
    expect(
      sanitizedErrorMessage(
        new FormaApiError('HTTP_ERROR', 'hidden internals', {
          errorKey: 'FORMA_DATA_SEMANTIC_MAPPING_ALREADY_CONFIRMED',
        }),
      ),
    ).toBe('该数据需求已经有正式映射。');
  });

  it('redacts secrets from error text', () => {
    expect(sanitizedErrorMessage(new Error('failed FORMA_G5_TEST_SUPER_SECRET_XYZ'), ['FORMA_G5_TEST_SUPER_SECRET_XYZ'])).toBe(
      '操作失败',
    );
  });
});
