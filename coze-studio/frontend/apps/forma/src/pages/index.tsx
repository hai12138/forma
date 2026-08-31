import { useFormaBaseline } from '@/hooks/use-forma-baseline';

export function OverviewPage() {
  const { loading, error, baseline } = useFormaBaseline();

  return (
    <div className="forma-panel">
      <h1 style={{ marginTop: 0 }}>Forma 总览</h1>
      <p>Business-to-Agent 产品 Shell（S0-B Foundation）。</p>
      <div className="forma-grid" style={{ marginTop: 16 }}>
        <div className="forma-card">
          <strong>产品基线</strong>
          <div>Forma v1.2 Visual Model Editor IA</div>
        </div>
        <div className="forma-card">
          <strong>Runtime Foundation</strong>
          <div>Coze / Eino（V1 默认）</div>
        </div>
        <div className="forma-card">
          <strong>Reference Business</strong>
          <div>维修工单（平台不硬编码）</div>
        </div>
      </div>
      <div className="forma-panel" style={{ marginTop: 16 }}>
        <strong>Platform Baseline</strong>
        {loading && <p>Loading…</p>}
        {error && <p className="forma-error">{error}</p>}
        {baseline && (
          <ul>
            <li>Forma: {baseline.forma_version}</li>
            <li>Schema: {baseline.forma_schema_version}</li>
            <li>Tag: {baseline.forma_baseline_tag}</li>
            <li>Coze commit: {baseline.coze_baseline_commit.slice(0, 12)}…</li>
            <li>Runtime: {baseline.runtime_foundation}</li>
          </ul>
        )}
      </div>
    </div>
  );
}

export function PlaceholderPage({ title }: { title: string }) {
  return (
    <div className="forma-panel">
      <h1 style={{ marginTop: 0 }}>{title}</h1>
      <p className="forma-placeholder">Forma module not connected yet</p>
      <p className="forma-placeholder">
        S0-B 仅建立 Shell 与工程基础；业务模块将在后续阶段接入。
      </p>
    </div>
  );
}

export function DesignPage() {
  const swatches = [
    ['background', 'var(--forma-background)'],
    ['surface', 'var(--forma-surface)'],
    ['primary', 'var(--forma-primary)'],
    ['foreground', 'var(--forma-foreground)'],
    ['border', 'var(--forma-border)'],
  ] as const;

  return (
    <div className="forma-panel">
      <h1 style={{ marginTop: 0 }}>Forma 设计系统</h1>
      <p>Apple-like Enterprise · AI Native · v1.2 Tokens</p>
      <div className="forma-grid">
        {swatches.map(([name, value]) => (
          <div key={name} className="forma-card">
            <div className="forma-token-swatch" style={{ background: value }} />
            <div style={{ marginTop: 8 }}>
              <strong>{name}</strong>
            </div>
            <code>{value}</code>
          </div>
        ))}
      </div>
    </div>
  );
}
