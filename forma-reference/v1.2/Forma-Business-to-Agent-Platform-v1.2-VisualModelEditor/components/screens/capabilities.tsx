import { BusinessCanvas } from '../business-canvas';
import { impactGraph } from '@/lib/business-canvas';
import { useState } from 'react';
import { Blocks } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Heading, Panel, Tabs, Badge, Notice, Rows, Field } from '../shared';
import { capabilities } from '@/lib/domain';
import { useStore } from '@/lib/store';
export default function Capabilities() {
  const { update } = useStore();
  const [selected, setSelected] = useState(capabilities[2]);
  const [tab, setTab] = useState('能力契约');
  const [impl, setImpl] = useState('Workflow');
  const [spec, setSpec] = useState(
    '{"openapi":"3.0.3","paths":{"/tickets":{"post":{"operationId":"createTicket"}}}}',
  );
  const [parsed, setParsed] = useState<string[]>([]);
  const [error, setError] = useState('');
  return (
    <>
      <Heading
        eyebrow="CAPABILITY ASSET / REGISTRY"
        title="把系统能力，转化为业务契约。"
        description="统一语义、权限和错误边界。实现可以替换，业务契约保持稳定。"
      >
        <Badge>{capabilities.length} 个示例能力</Badge>
      </Heading>
      <div className="split catalog-layout">
        <Panel title="Capability Registry">
          <div className="registry-list">
            {capabilities.map((c) => (
              <button
                key={c.id}
                onClick={() => {
                  setSelected(c);
                  setImpl(c.impl);
                }}
                className={selected.id === c.id ? 'selected' : ''}
              >
                <Blocks size={18} />
                <span>
                  <b>{c.name}</b>
                  <small>{c.id}</small>
                </span>
              </button>
            ))}
          </div>
        </Panel>
        <div>
          <div className="detail-heading">
            <div>
              <h2>{selected.name}</h2>
              <p className="muted">
                <code>{selected.id}</code> · v{selected.version}
              </p>
            </div>
            <Badge tone={selected.risk === '高风险' ? 'warning' : 'neutral'}>
              {selected.risk}
            </Badge>
          </div>
          <Tabs
            items={[
              '能力契约',
              '实现绑定',
              '依赖图',
              'OpenAPI Mapping',
              'Adapter SDK',
            ]}
            value={tab}
            onChange={setTab}
          />
          {tab === '能力契约' && (
            <Panel title="Business Capability Contract">
              <div className="two-col">
                <div>
                  <Field label="输入 Schema">
                    <pre className="code-block">
                      {JSON.stringify(
                        {
                          work_order_id: 'string · required',
                          tenant_id: 'from identity',
                          idempotency_key: 'string · required',
                        },
                        null,
                        2,
                      )}
                    </pre>
                  </Field>
                  <Field label="输出 Schema">
                    <pre className="code-block">
                      {JSON.stringify(
                        {
                          status: 'success | pending',
                          assignment_id: 'string',
                          audit_id: 'string',
                        },
                        null,
                        2,
                      )}
                    </pre>
                  </Field>
                </div>
                <div>
                  <Rows
                    headers={['约束', '声明']}
                    rows={[
                      ['Permission', 'work_order:write + tenant scope'],
                      ['Side effect', '业务写入 / 外部调用'],
                      ['Confirmation', '高风险必须确认'],
                      ['Timeout / SLA', '10s / 99.9% 目标'],
                      ['Preconditions', '工单存在，当前状态允许派单'],
                      ['Postconditions', '责任人确定，审计事件落库'],
                      ['Compatibility', 'SemVer · 破坏性变更新 Major'],
                    ]}
                  />
                </div>
              </div>
              <Notice>
                错误契约：PERMISSION_DENIED / STATE_CONFLICT / TIMEOUT /
                ADAPTER_UNAVAILABLE。写操作超时返回结果未知，不允许盲目重试。
              </Notice>
            </Panel>
          )}
          {tab === '实现绑定' && (
            <Panel title="Implementation Binding">
              <div className="implementation-grid">
                {[
                  'REST API',
                  'MCP',
                  'Database',
                  'Workflow',
                  'Managed Runtime',
                  'Human Task',
                ].map((x) => (
                  <button
                    className={impl === x ? 'selected' : ''}
                    onClick={() => setImpl(x)}
                    key={x}
                  >
                    {x}
                  </button>
                ))}
              </div>
              <Field label="实现入口（示例）">
                <Input
                  key={impl}
                  defaultValue={
                    {
                      'REST API': 'adapter://property-api/v2',
                      MCP: 'mcp://knowledge/search',
                      Database: 'query://asset-by-id',
                      Workflow: 'workflow://urgent-dispatch',
                      'Managed Runtime': 'managed://work-order/assign',
                      'Human Task': 'human://supervisor-approval',
                    }[impl]
                  }
                />
              </Field>
              <Notice>
                切换实现前必须通过同一契约测试集。此处保存为模拟审计记录，不建立真实连接。
              </Notice>
              <Button
                onClick={() =>
                  update({}, '已模拟绑定 ' + selected.id + ' → ' + impl, true)
                }
              >
                保存实现绑定
              </Button>
            </Panel>
          )}
          {tab === '依赖图' && (
            <Panel title="显式依赖 · 影响可追踪">
              <BusinessCanvas
                graph={impactGraph(selected.id + '@' + selected.version)}
                title="能力依赖关系"
                compact
              />
              <Notice>
                版本引用固定到资产快照，不在运行时解析
                latest；循环依赖在构建阶段拒绝。
              </Notice>
            </Panel>
          )}
          {tab === 'OpenAPI Mapping' && (
            <Panel title="从 OpenAPI 推荐能力映射">
              <Field label="粘贴 OpenAPI JSON（本地解析器）">
                <Textarea
                  rows={7}
                  value={spec}
                  onChange={(e) => setSpec(e.target.value)}
                />
              </Field>
              <Button
                onClick={() => {
                  try {
                    const s = JSON.parse(spec);
                    if (!s.openapi || !s.paths)
                      throw Error('需要 openapi 和 paths 字段');
                    const ops: string[] = [];
                    Object.entries(s.paths).forEach(([p, methods]) =>
                      Object.entries(
                        methods as Record<string, unknown>,
                      ).forEach(([method, value]) => {
                        if (
                          ['get', 'post', 'put', 'patch', 'delete'].includes(
                            method,
                          )
                        )
                          ops.push(
                            method.toUpperCase() +
                              ' ' +
                              p +
                              ' → ' +
                              ((value as { operationId?: string })
                                .operationId || 'unmapped'),
                          );
                      }),
                    );
                    if (!ops.length) throw Error('未发现支持的 HTTP 操作');
                    setParsed(ops);
                    setError('');
                  } catch (e) {
                    setError((e as Error).message);
                    setParsed([]);
                  }
                }}
              >
                分析接口
              </Button>
              {error && <Notice tone="warning">{error}</Notice>}
              {parsed.map((x) => (
                <div className="mapping-result" key={x}>
                  <code>{x}</code>
                  <Badge tone="warning">待人工确认</Badge>
                </div>
              ))}
              <p className="muted paragraph">
                当前只做结构解析，不推断权限、副作用或业务语义。生产 AI
                推荐必须由能力负责人批准。
              </p>
            </Panel>
          )}
          {tab === 'Adapter SDK' && (
            <Panel title="稳定契约，替换实现">
              <pre className="code-block">{`interface CapabilityAdapter {
  describe(): Contract;
  validate(input: unknown): Validation;
  execute(ctx: TenantContext, input: unknown): Promise<Result>;
  health(): Promise<Health>;
  reconcile(idempotencyKey: string): Promise<Result>;
}`}</pre>
              <Rows
                headers={['交付物', '验证要求']}
                rows={[
                  ['manifest + capabilities', '版本、权限与错误枚举'],
                  ['Contract fixtures', '成功 / 拒绝 / 超时 / 重放'],
                  ['Trace bridge', 'trace_id / span_id / audit_id'],
                  ['Secret reference', '只使用 Vault 引用，禁止明文打包'],
                ]}
              />
            </Panel>
          )}
        </div>
      </div>
    </>
  );
}
