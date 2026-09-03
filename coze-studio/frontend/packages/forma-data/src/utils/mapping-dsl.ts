export const MAPPING_TYPES = [
  'DIRECT',
  'CAST',
  'ENUM_MAP',
  'UNIT_CONVERT',
  'TIME_NORMALIZE',
  'FIELD_PATH',
  'JOIN_REF',
] as const;

export type MappingType = (typeof MAPPING_TYPES)[number];

function parseEnumPairs(raw: string): Record<string, string> {
  const pairs: Record<string, string> = {};
  for (const line of raw.split(/[\n,]/)) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const eq = trimmed.indexOf('=');
    if (eq <= 0) continue;
    pairs[trimmed.slice(0, eq).trim()] = trimmed.slice(eq + 1).trim();
  }
  return pairs;
}

function splitList(raw: string): string[] {
  return raw
    .split(',')
    .map(s => s.trim())
    .filter(Boolean);
}

/**
 * Controlled transform_spec builder.
 * `type` always equals `mappingType`. No SQL / JS / Python / arbitrary expressions.
 */
export function buildTransformSpec(
  mappingType: string,
  form: Record<string, string>,
): Record<string, unknown> {
  switch (mappingType) {
    case 'CAST':
      return {
        type: 'CAST',
        from_type: form.from_type || 'string',
        to_type: form.to_type || 'string',
      };
    case 'ENUM_MAP':
      return { type: 'ENUM_MAP', pairs: parseEnumPairs(form.pairs || '') };
    case 'UNIT_CONVERT':
      return {
        type: 'UNIT_CONVERT',
        from_unit: form.from_unit || '',
        to_unit: form.to_unit || '',
        factor: Number(form.factor || 1),
        offset: Number(form.offset || 0),
      };
    case 'TIME_NORMALIZE':
      return {
        type: 'TIME_NORMALIZE',
        source_timezone: form.source_timezone || 'UTC',
        target_timezone: form.target_timezone || 'UTC',
        format: form.format || '',
      };
    case 'FIELD_PATH':
      return { type: 'FIELD_PATH', path: form.path || '' };
    case 'JOIN_REF':
      return {
        type: 'JOIN_REF',
        relationship: form.relationship || '',
        from_fields: splitList(form.from_fields || ''),
        to_schema: form.to_schema || '',
        to_fields: splitList(form.to_fields || ''),
      };
    case 'DIRECT':
      return { type: 'DIRECT' };
    default:
      return { type: mappingType };
  }
}
