import type { FormaApiClient, FormaBusiness, FormaSemanticModel } from '@forma/api-client';

import { workOrderSeed } from './work-order-seed';

/** Pure handler factory — tested without mounting the full page. */
export function createBusinessSubmitHandler(opts: {
  client: Pick<FormaApiClient, 'createBusiness'>;
  name: string;
  seedWorkOrder: boolean;
  onCreated: (b: FormaBusiness) => void;
  onError: (msg: string) => void;
}) {
  return async () => {
    const trimmed = opts.name.trim();
    if (!trimmed) {
      opts.onError('请输入业务名称');
      return;
    }
    try {
      const body: {
        name: string;
        semantic_model?: FormaSemanticModel;
        change_summary?: string;
      } = { name: trimmed };
      if (opts.seedWorkOrder || trimmed === '维修工单') {
        body.semantic_model = workOrderSeed();
        body.change_summary = '初始化维修工单语义模型';
      }
      const resp = await opts.client.createBusiness(body);
      opts.onCreated(resp.data);
    } catch (err) {
      opts.onError(err instanceof Error ? err.message : '创建失败');
    }
  };
}

export function formatUpdatedAt(iso: string): string {
  if (!iso) return '—';
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}
