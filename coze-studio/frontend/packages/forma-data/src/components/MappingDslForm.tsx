import { MAPPING_TYPES } from '../utils/mapping-dsl';

export interface MappingDslFormProps {
  mappingType: string;
  form: Record<string, string>;
  onChange: (next: Record<string, string>) => void;
}

function Field({
  id,
  label,
  value,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="forma-form-row">
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        data-testid={id}
        value={value}
        onChange={e => onChange(e.target.value)}
      />
    </div>
  );
}

/** Controlled mapping DSL fields. No SQL / script / arbitrary expression editors. */
export function MappingDslForm({ mappingType, form, onChange }: MappingDslFormProps) {
  const set = (key: string, value: string) => onChange({ ...form, [key]: value });

  if (!MAPPING_TYPES.includes(mappingType as (typeof MAPPING_TYPES)[number])) {
    return null;
  }

  if (mappingType === 'DIRECT') {
    return null;
  }

  if (mappingType === 'CAST') {
    return (
      <>
        <Field id="dsl-from-type" label="来源类型" value={form.from_type ?? ''} onChange={v => set('from_type', v)} />
        <Field id="dsl-to-type" label="目标逻辑类型" value={form.to_type ?? ''} onChange={v => set('to_type', v)} />
      </>
    );
  }

  if (mappingType === 'ENUM_MAP') {
    return (
      <Field
        id="dsl-enum-pairs"
        label="枚举对应（key=value，逗号或换行分隔）"
        value={form.pairs ?? ''}
        onChange={v => set('pairs', v)}
      />
    );
  }

  if (mappingType === 'UNIT_CONVERT') {
    return (
      <>
        <Field id="dsl-from-unit" label="源单位" value={form.from_unit ?? ''} onChange={v => set('from_unit', v)} />
        <Field id="dsl-to-unit" label="目标单位" value={form.to_unit ?? ''} onChange={v => set('to_unit', v)} />
        <Field id="dsl-factor" label="factor" value={form.factor ?? '1'} onChange={v => set('factor', v)} />
        <Field id="dsl-offset" label="offset" value={form.offset ?? '0'} onChange={v => set('offset', v)} />
      </>
    );
  }

  if (mappingType === 'TIME_NORMALIZE') {
    return (
      <>
        <Field
          id="dsl-source-timezone"
          label="source_timezone"
          value={form.source_timezone ?? ''}
          onChange={v => set('source_timezone', v)}
        />
        <Field
          id="dsl-target-timezone"
          label="target_timezone"
          value={form.target_timezone ?? ''}
          onChange={v => set('target_timezone', v)}
        />
        <Field id="dsl-format" label="format" value={form.format ?? ''} onChange={v => set('format', v)} />
      </>
    );
  }

  if (mappingType === 'FIELD_PATH') {
    return (
      <Field id="dsl-field-path" label="path" value={form.path ?? ''} onChange={v => set('path', v)} />
    );
  }

  if (mappingType === 'JOIN_REF') {
    return (
      <>
        <Field
          id="dsl-relationship"
          label="relationship"
          value={form.relationship ?? ''}
          onChange={v => set('relationship', v)}
        />
        <Field
          id="dsl-from-fields"
          label="from_fields"
          value={form.from_fields ?? ''}
          onChange={v => set('from_fields', v)}
        />
        <Field id="dsl-to-schema" label="to_schema" value={form.to_schema ?? ''} onChange={v => set('to_schema', v)} />
        <Field id="dsl-to-fields" label="to_fields" value={form.to_fields ?? ''} onChange={v => set('to_fields', v)} />
      </>
    );
  }

  return null;
}
