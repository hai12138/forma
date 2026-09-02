export function EmptyState({ title, hint }: { title: string; hint?: string }) {
  return (
    <div className="forma-empty-state" data-testid="empty-state">
      <div>{title}</div>
      {hint ? <div className="forma-muted" style={{ marginTop: 8 }}>{hint}</div> : null}
    </div>
  );
}
