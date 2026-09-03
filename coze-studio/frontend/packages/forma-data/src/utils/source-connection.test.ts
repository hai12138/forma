import { describe, expect, it } from 'vitest';

import {
  adaptersForSourceType,
  buildConnectionPublicConfig,
  buildCreateConnectionBody,
  defaultAdapterForSourceType,
  isValidSourceType,
  SOURCE_TYPES,
} from './source-connection';

describe('source-connection alignment', () => {
  it('only allows backend source_type enums', () => {
    expect(SOURCE_TYPES).toEqual(['RELATIONAL_DATABASE', 'HTTP_API']);
    expect(isValidSourceType('RELATIONAL_DATABASE')).toBe(true);
    expect(isValidSourceType('HTTP_API')).toBe(true);
    expect(isValidSourceType('EXTERNAL_SQL')).toBe(false);
    expect(isValidSourceType('EXTERNAL_HTTP')).toBe(false);
    expect(isValidSourceType('MYSQL')).toBe(false);
  });

  it('maps adapters per source type (never MYSQL for HTTP)', () => {
    expect(adaptersForSourceType('RELATIONAL_DATABASE')).toEqual(['MYSQL', 'POSTGRESQL']);
    expect(adaptersForSourceType('HTTP_API')).toEqual(['HTTP']);
    expect(defaultAdapterForSourceType('HTTP_API')).toBe('HTTP');
    expect(adaptersForSourceType('HTTP_API')).not.toContain('MYSQL');
  });

  it('builds relational public_config', () => {
    expect(
      buildConnectionPublicConfig('RELATIONAL_DATABASE', 'POSTGRESQL', {
        host: 'db.example',
        port: '',
        database: 'lab',
        username: 'u',
      }),
    ).toEqual({ host: 'db.example', port: 5432, database: 'lab', username: 'u' });
  });

  it('builds HTTP public_config and rejects SQL shape for HTTP body', () => {
    const body = buildCreateConnectionBody({
      name: 'api',
      environment: 'DEV',
      sourceType: 'HTTP_API',
      adapterType: 'MYSQL',
      form: {
        host: 'should-not-appear',
        base_url: 'https://api.example/v1',
        openapi_url: 'https://api.example/openapi.json',
      },
    });
    expect(body.adapter_type).toBe('HTTP');
    expect(body.public_config).toEqual({
      base_url: 'https://api.example/v1',
      openapi_url: 'https://api.example/openapi.json',
    });
    expect(body.public_config).not.toHaveProperty('host');
    expect(body.public_config).not.toHaveProperty('database');
  });
});
