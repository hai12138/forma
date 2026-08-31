import { useFormaBaseline } from '@/hooks/use-forma-baseline';
import { useFormaSession } from '@/hooks/use-forma-session';

export function OverviewPage() {
  const { loading, error, baseline } = useFormaBaseline();
  const { state, assetCounts, currentTenant } = useFormaSession();

  const counts = assetCounts ?? {
    business: 0,
    capability: 0,
    agent: 0,
    application: 0,
  };
  const total =
    counts.business + counts.capability + counts.agent + counts.application;

  return (
    <div className="forma-panel">
      <h1 style={{ marginTop: 0 }}>Forma 总览</h1>
      <p>
        Tenant 安全边界已启用
        {currentTenant ? ` · ${currentTenant.display_name}` : ''}。
      </p>

      <div className="forma-grid" style={{ marginTop: 16 }}>
        <div className="forma-card">
          <strong>Business</strong>
          <div className="forma-kpi">{counts.business}</div>
        </div>
        <div className="forma-card">
          <strong>Capability</strong>
          <div className="forma-kpi">{counts.capability}</div>
        </div>
        <div className="forma-card">
          <strong>Agent</strong>
          <div className="forma-kpi">{counts.agent}</div>
        </div>
        <div className="forma-card">
          <strong>Application</strong>
          <div className="forma-kpi">{counts.application}</div>
        </div>
      </div>

      {state === 'ready' && total === 0 && (
        <div className="forma-panel" style={{ marginTop: 16 }}>
          <strong>Empty State</strong>
          <p className="forma-placeholder">
            当前 Tenant 尚无资产。Business / Capability / Agent / Application
            将在后续阶段接入；此处为真实 Asset Registry 计数，非 Mock KPI。
          </p>
        </div>
      )}

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
        S1 仅建立 Tenant / Identity 产品基础；业务模块将在后续阶段接入。
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
