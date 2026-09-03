import { describe, expect, it } from 'vitest';

import { buildTransformSpec, MAPPING_TYPES } from '../utils/mapping-dsl';

describe('mapping DSL alignment', () => {
  it('uses TIME_NORMALIZE and FIELD_PATH, not TIMEZONE or JSON_PATH', () => {
    expect(MAPPING_TYPES).toContain('TIME_NORMALIZE');
    expect(MAPPING_TYPES).toContain('FIELD_PATH');
    expect(MAPPING_TYPES).not.toContain('TIMEZONE');
    expect(MAPPING_TYPES).not.toContain('JSON_PATH');
    expect(MAPPING_TYPES.join(',')).not.toMatch(/SQL|JAVASCRIPT|PYTHON|script/i);
  });

  it('keeps transform_spec.type equal to mapping_type for every controlled type', () => {
    for (const type of MAPPING_TYPES) {
      const spec = buildTransformSpec(type, {
        from_type: 'string',
        to_type: 'number',
        pairs: 'A=B',
        from_unit: 'celsius',
        to_unit: 'kelvin',
        factor: '1',
        offset: '273.15',
        source_timezone: 'Asia/Shanghai',
        target_timezone: 'UTC',
        format: 'RFC3339',
        path: 'sensor.temperature',
        relationship: 'sample_has_reading',
        from_fields: 'id',
        to_schema: 'readings',
        to_fields: 'sample_id',
      });
      expect(spec.type).toBe(type);
    }
  });

  it('builds TIME_NORMALIZE with source_timezone, target_timezone, format', () => {
    const spec = buildTransformSpec('TIME_NORMALIZE', {
      source_timezone: 'Asia/Shanghai',
      target_timezone: 'UTC',
      format: 'RFC3339',
    });
    expect(spec).toEqual({
      type: 'TIME_NORMALIZE',
      source_timezone: 'Asia/Shanghai',
      target_timezone: 'UTC',
      format: 'RFC3339',
    });
  });

  it('builds FIELD_PATH with path', () => {
    const spec = buildTransformSpec('FIELD_PATH', { path: 'sensor.temperature' });
    expect(spec).toEqual({ type: 'FIELD_PATH', path: 'sensor.temperature' });
  });
});
