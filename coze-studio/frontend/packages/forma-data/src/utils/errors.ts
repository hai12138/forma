import { FormaApiError } from '@forma/api-client';

const ERROR_LABELS: Record<string, string> = {
  FORMA_DATA_CONTRACT_NOT_ACTIVE: '当前没有可用的活动数据契约。',
  FORMA_DATA_SEMANTIC_MAPPING_ALREADY_CONFIRMED: '该数据需求已经有正式映射。',
  FORMA_DATA_SEMANTIC_MAPPING_ALREADY_DECIDED: '该映射已经完成人工决策。',
  UNAUTHORIZED: '未登录或会话已失效。',
  FORBIDDEN: '没有权限执行该操作。',
  NOT_FOUND: '资源不存在或无权访问。',
};

/** Allowlisted contract validation issue codes → product copy. Never echo backend message. */
const VALIDATION_ISSUE_LABELS: Record<string, string> = {
  REQUIREMENT_IDS_DUPLICATE: '需求 ID 重复。',
  BUSINESS_REVISION: '业务模型修订不存在。',
  REQUIREMENT: '需求未在钉住的业务修订中确认。',
  BINDING_MAPPING: '绑定映射无效或未确认。',
  BINDING_LINEAGE: '绑定血缘不一致。',
  SNAPSHOT: '结构快照不存在。',
  SCHEMA_JSON: '结构快照无效。',
  SCHEMA_PATHS: '结构路径校验失败。',
  MAPPING_TARGET: '映射目标字段无效。',
  TRANSFORM: '转换规格无效。',
  BINDING_COUNT: '每个需求必须恰好有一个绑定。',
  BINDING_ORPHAN: '存在未声明需求的绑定。',
  LOGICAL_SCHEMA: '逻辑 Schema 无效。',
  LOGICAL_COVERAGE: '逻辑字段覆盖不完整。',
  QUERY_CAPABILITIES: '查询能力无效。',
  FILTER_SCHEMA: '过滤 Schema 无效。',
  SORT_SCHEMA: '排序 Schema 无效。',
  PAGINATION: '分页策略无效。',
  FRESHNESS: '新鲜度策略无效。',
  CLASSIFICATION: '分级策略无效。',
  LOGICAL_TYPE_MISMATCH: '逻辑类型不匹配。',
  NULLABILITY_GUARANTEE_LOST: '空值保证丢失。',
};

const GENERIC_FAILURE = '操作失败';
const GENERIC_VALIDATION_FAILURE = '校验未通过，请检查契约配置。';

/**
 * Map stable Forma error_key/code to product copy.
 * Unknown FormaApiError and plain Error always use a generic message — never echo err.message.
 */
export function sanitizedErrorMessage(err: unknown, _redact: string[] = []): string {
  if (err instanceof FormaApiError) {
    if (err.errorKey && ERROR_LABELS[err.errorKey]) {
      return ERROR_LABELS[err.errorKey];
    }
    if (ERROR_LABELS[err.code]) {
      return ERROR_LABELS[err.code];
    }
  }
  return GENERIC_FAILURE;
}

/** Map allowlisted validation issue code to product copy; never surface raw backend message. */
export function validationIssueLabel(code: string, _message?: string): string {
  if (code && VALIDATION_ISSUE_LABELS[code]) {
    return VALIDATION_ISSUE_LABELS[code];
  }
  return GENERIC_VALIDATION_FAILURE;
}

export async function safeMutate(
  action: () => Promise<void>,
  onError: (message: string) => void,
  redact: string[] = [],
): Promise<boolean> {
  try {
    await action();
    return true;
  } catch (err) {
    onError(sanitizedErrorMessage(err, redact));
    return false;
  }
}
