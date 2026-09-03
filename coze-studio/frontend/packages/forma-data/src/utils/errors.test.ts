import { describe, expect, it } from 'vitest';

import { FormaApiError } from '@forma/api-client';

import { sanitizedErrorMessage, validationIssueLabel } from './errors';

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

  it('never echoes unknown FormaApiError or plain Error messages', () => {
    const leak =
      'postgres://user:password@db/prod Authorization: Bearer tok_abc token=xyz';
    expect(sanitizedErrorMessage(new Error(leak))).toBe('操作失败');
    expect(
      sanitizedErrorMessage(new FormaApiError('HTTP_ERROR', leak, { errorKey: 'UNKNOWN_KEY' })),
    ).toBe('操作失败');
    expect(sanitizedErrorMessage(new FormaApiError('SOME_CODE', leak))).toBe('操作失败');
  });

  it('redact param never causes message echo', () => {
    expect(
      sanitizedErrorMessage(new Error('failed FORMA_G5_TEST_SUPER_SECRET_XYZ'), [
        'FORMA_G5_TEST_SUPER_SECRET_XYZ',
      ]),
    ).toBe('操作失败');
  });
});

describe('validationIssueLabel', () => {
  it('maps allowlisted codes and never returns raw backend message', () => {
    expect(validationIssueLabel('BINDING_LINEAGE', 'binding "x" lineage mismatch')).toBe(
      '绑定血缘不一致。',
    );
    const leak = 'dsn=postgres://u:password@h/db token=abc Authorization: Bearer z';
    expect(validationIssueLabel('UNKNOWN_CODE', leak)).toBe('校验未通过，请检查契约配置。');
    expect(validationIssueLabel('UNKNOWN_CODE', leak)).not.toContain('password');
  });
});
