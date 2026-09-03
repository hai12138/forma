import { FormaApiError } from '@forma/api-client';

const ERROR_LABELS: Record<string, string> = {
  FORMA_DATA_CONTRACT_NOT_ACTIVE: '当前没有可用的活动数据契约。',
  FORMA_DATA_SEMANTIC_MAPPING_ALREADY_CONFIRMED: '该数据需求已经有正式映射。',
  FORMA_DATA_SEMANTIC_MAPPING_ALREADY_DECIDED: '该映射已经完成人工决策。',
  UNAUTHORIZED: '未登录或会话已失效。',
  FORBIDDEN: '没有权限执行该操作。',
  NOT_FOUND: '资源不存在或无权访问。',
};

/** Map stable Forma errors to user copy. Never dump stacks or driver internals. */
export function sanitizedErrorMessage(err: unknown, redact: string[] = []): string {
  let message = '操作失败';
  if (err instanceof FormaApiError) {
    if (err.errorKey && ERROR_LABELS[err.errorKey]) {
      message = ERROR_LABELS[err.errorKey];
    } else if (ERROR_LABELS[err.code]) {
      message = ERROR_LABELS[err.code];
    } else {
      message = err.message || '操作失败';
    }
  } else if (err instanceof Error) {
    message = err.message || '操作失败';
  }
  for (const secret of redact) {
    if (secret && message.includes(secret)) {
      message = '操作失败';
      break;
    }
  }
  return message;
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
