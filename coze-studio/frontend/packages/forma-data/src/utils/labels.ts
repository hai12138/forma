/** Status / readiness labels — domain-agnostic Chinese UI (no i18n lib). */

const STATUS_LABELS: Record<string, string> = {
  PROPOSED: '待确认',
  CONFIRMED: '已确认',
  REJECTED: '已拒绝',
  SUPERSEDED: '已替代',
  ACTIVE: '生效中',
  DRAFT: '草稿',
  VALIDATED: '已验证',
  STALE: '已过期',
  DEPRECATED: '已弃用',
  ARCHIVED: '已归档',
  PENDING: '处理中',
  FAILED: '失败',
  SUCCEEDED: '成功',
  PASSED: '通过',
  VALID: '有效',
  INVALID: '无效',
};

export function statusLabel(status: string | undefined | null): string {
  if (!status) return '未知';
  return STATUS_LABELS[status] ?? status;
}

export function readinessLabel(opts: {
  confirmedRequirements: number;
  coverage: number;
  activeContracts: number;
  staleContracts: number;
}): string {
  if (opts.confirmedRequirements === 0) {
    return '尚未确认数据需求 — 可从业务模型分析或手动新增';
  }
  if (opts.coverage < 1) {
    return `映射覆盖率 ${(opts.coverage * 100).toFixed(0)}% — 仍有未映射的已确认需求`;
  }
  if (opts.activeContracts === 0) {
    return '映射已就绪 — 可创建数据契约';
  }
  if (opts.staleContracts > 0) {
    return `有 ${opts.staleContracts} 个过期契约修订 — 请评估漂移或发布新修订`;
  }
  return '数据平面就绪 — 需求、映射与契约均可用';
}

export function confidenceDisclaimer(): string {
  return '置信度不代表已确认';
}

/** Product-facing provenance label — never expose runtime vendor names. */
export function mappingSourceLabel(source: string | undefined | null): string {
  switch (source) {
    case 'AI_GENERATED':
      return 'AI 建议';
    case 'MANUAL':
      return '人工创建';
    case 'MANUAL_MODIFIED':
      return '人工修改并确认';
    default:
      return source || '未知来源';
  }
}
