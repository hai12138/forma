/** Parse a controlled transform_spec back into MappingDslForm string fields. */
export function parseTransformSpecToForm(
  spec: Record<string, unknown> | string | undefined | null,
): Record<string, string> {
  if (!spec) return {};
  let obj: Record<string, unknown>;
  if (typeof spec === 'string') {
    try {
      obj = JSON.parse(spec) as Record<string, unknown>;
    } catch {
      return {};
    }
  } else {
    obj = spec;
  }
  const type = String(obj.type || '');
  switch (type) {
    case 'CAST':
      return {
        from_type: String(obj.from_type ?? ''),
        to_type: String(obj.to_type ?? ''),
      };
    case 'ENUM_MAP': {
      const pairs = obj.pairs;
      if (pairs && typeof pairs === 'object' && !Array.isArray(pairs)) {
        return {
          pairs: Object.entries(pairs as Record<string, string>)
            .map(([k, v]) => `${k}=${v}`)
            .join('\n'),
        };
      }
      return { pairs: '' };
    }
    case 'UNIT_CONVERT':
      return {
        from_unit: String(obj.from_unit ?? ''),
        to_unit: String(obj.to_unit ?? ''),
        factor: String(obj.factor ?? '1'),
        offset: String(obj.offset ?? '0'),
      };
    case 'TIME_NORMALIZE':
      return {
        source_timezone: String(obj.source_timezone ?? 'UTC'),
        target_timezone: String(obj.target_timezone ?? 'UTC'),
        format: String(obj.format ?? ''),
      };
    case 'FIELD_PATH':
      return { path: String(obj.path ?? '') };
    case 'JOIN_REF':
      return {
        relationship: String(obj.relationship ?? ''),
        from_fields: Array.isArray(obj.from_fields)
          ? (obj.from_fields as string[]).join(',')
          : '',
        to_schema: String(obj.to_schema ?? ''),
        to_fields: Array.isArray(obj.to_fields) ? (obj.to_fields as string[]).join(',') : '',
      };
    default:
      return {};
  }
}
