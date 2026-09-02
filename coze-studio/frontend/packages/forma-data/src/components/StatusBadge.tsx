import { statusLabel } from '../utils/labels';

export function StatusBadge({ status }: { status: string }) {
  const cls = `forma-badge forma-badge-${status.toLowerCase()}`;
  return (
    <span className={cls} data-testid="status-badge" data-status={status}>
      {statusLabel(status)}
    </span>
  );
}
