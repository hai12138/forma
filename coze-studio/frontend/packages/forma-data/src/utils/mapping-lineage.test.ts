import { describe, expect, it } from 'vitest';

import { lineageMatchesSnapshot, mappingLineageFromSnapshot } from './mapping-lineage';

const validSnap = {
  snapshot_id: 'snap_lab',
  source_id: 'src_lab',
  connection_id: 'conn_lab',
  asset_id: 'asset_lab',
  schema: { fields: [] },
  fingerprint: 'fp',
  created_by: 'p',
  created_at: '2026-01-01T00:00:00Z',
};

describe('mapping-lineage', () => {
  it('builds complete lineage from a valid snapshot DTO', () => {
    expect(mappingLineageFromSnapshot(validSnap)).toEqual({
      source_id: 'src_lab',
      connection_id: 'conn_lab',
      asset_id: 'asset_lab',
      schema_snapshot_id: 'snap_lab',
    });
  });

  it('rejects incomplete snapshot (no manual placeholders)', () => {
    expect(mappingLineageFromSnapshot({ snapshot_id: 'snap_lab' })).toBeNull();
    expect(
      mappingLineageFromSnapshot({
        snapshot_id: 'snap_x',
        source_id: 'src_manual',
        connection_id: '',
        asset_id: 'asset_manual',
      }),
    ).toBeNull();
  });

  it('detects lineage mismatch against snapshot', () => {
    const lineage = mappingLineageFromSnapshot(validSnap)!;
    expect(lineageMatchesSnapshot(lineage, validSnap)).toBe(true);
    expect(
      lineageMatchesSnapshot(
        { ...lineage, source_id: 'src_manual', connection_id: 'conn_manual', asset_id: 'asset_manual' },
        validSnap,
      ),
    ).toBe(false);
  });
});
