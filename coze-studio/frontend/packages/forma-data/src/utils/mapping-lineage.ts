import type { FormaSchemaSnapshot } from '@forma/api-client';

export type MappingLineage = {
  source_id: string;
  connection_id: string;
  asset_id: string;
  schema_snapshot_id: string;
};

/** Extract consistent physical lineage from a SchemaSnapshot DTO. */
export function mappingLineageFromSnapshot(
  snap: Partial<FormaSchemaSnapshot> | null | undefined,
): MappingLineage | null {
  if (!snap) return null;
  const source_id = String(snap.source_id ?? '').trim();
  const connection_id = String(snap.connection_id ?? '').trim();
  const asset_id = String(snap.asset_id ?? '').trim();
  const schema_snapshot_id = String(snap.snapshot_id ?? '').trim();
  if (!source_id || !connection_id || !asset_id || !schema_snapshot_id) {
    return null;
  }
  return { source_id, connection_id, asset_id, schema_snapshot_id };
}

export function lineageMatchesSnapshot(
  lineage: MappingLineage,
  snap: Partial<FormaSchemaSnapshot>,
): boolean {
  return (
    lineage.source_id === snap.source_id &&
    lineage.connection_id === snap.connection_id &&
    lineage.asset_id === snap.asset_id &&
    lineage.schema_snapshot_id === snap.snapshot_id
  );
}
