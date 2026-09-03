import { statusLabel } from '../utils/labels';

export function StatusBadge({ status }: { status: string }) {
  const safe = status || 'UNKNOWN';
  const cls = `forma-badge forma-badge-${safe.toLowerCase()}`;
  return (
    <span className={cls} data-testid="status-badge" data-status={safe}>
      {statusLabel(safe)}
    </span>
  );
}
