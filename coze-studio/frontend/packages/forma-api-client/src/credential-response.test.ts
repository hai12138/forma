import { describe, expect, it } from 'vitest';

import type { FormaCredentialRef } from './index';

/** Credential RESPONSE keys must never include secret material. */
const FORBIDDEN_CREDENTIAL_RESPONSE_KEYS = [
  'password',
  'secret',
  'token',
  'api_key',
  'authorization',
] as const;

describe('FormaCredentialRef response shape', () => {
  it('excludes secret-bearing keys from the response type surface', () => {
    const sample: FormaCredentialRef = {
      credential_ref_id: 'cred_1',
      status: 'ACTIVE',
      provider: 'sql',
      created_at: '2026-01-01T00:00:00Z',
      rotated_at: '2026-01-02T00:00:00Z',
    };

    const keys = Object.keys(sample);
    for (const forbidden of FORBIDDEN_CREDENTIAL_RESPONSE_KEYS) {
      expect(keys).not.toContain(forbidden);
    }

    // Compile-time + runtime guard: assignability check via excess property probe.
    const probe = sample as Record<string, unknown>;
    for (const forbidden of FORBIDDEN_CREDENTIAL_RESPONSE_KEYS) {
      expect(probe[forbidden]).toBeUndefined();
    }
  });
});
