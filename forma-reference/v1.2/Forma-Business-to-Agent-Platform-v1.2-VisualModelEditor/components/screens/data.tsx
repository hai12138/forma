import { BusinessCanvas } from '../business-canvas';
import { dataGraph } from '@/lib/business-canvas';
import { useState } from 'react';
import { Database, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Heading, Panel, Tabs, Badge, Notice, Rows } from '../shared';
import { useStore } from '@/lib/store';
export default function DataPlane() {
  const { state, update, notify } = useStore();
  const [tab, setTab] = useState('业务数据');
  const [mapping, setMapping] = useState(state.mapping);
  const [valid, setValid] = useState(false);
  return (
    <>
      <Heading
        eyebrow="DATA PLANE / CANONICAL BUSINESS DATA"
        title="同一份业务语义，不同的数据来源。"
        description="交易数据、知识与临时运行数据分离。业务 Agent 只依赖统一的数据契约。"
      >
        <Badge>{state.dataMode} mode</Badge>
      </Heading>
      <Tabs
        items={[
          '业务数据',
          'Data Contract',
          '知识存储',
          '临时运行数据',
          '隔离与迁移',
        ]}
        value={tab}
        onChange={setTab}
      />
      {tab === '业务数据' && (
        <>
          <div className="three-col">
            {[
              [
                'External',
                '连接现有业务系统',
                '客户系统为事实来源；平台仅保存引用与执行记录。',
              ],
              [
                'Managed',
                '托管最小业务能力',
                '根据业务模型生成对象、状态机和受控 CRUD。',
              ],
              [
                'Hybrid',
                '按业务对象组合',
                '工单由平台托管，设备从客户资产系统读取。',
              ],
            ].map((x) => (
              <button
                key={x[0]}
                className={
                  'mode-card ' + (state.dataMode === x[0] ? 'selected' : '')
                }
                onClick={() =>
                  update({ dataMode: x[0] }, '数据模式已改为 ' + x[0], true)
                }
              >
                <span className="icon-tile blue">
                  <Database size={22} />
                </span>
                <h2>
                  {x[0]} {state.dataMode === x[0] && <Check size={16} />}
                </h2>
                <h3>{x[1]}</h3>
                <p>{x[2]}</p>
              </button>
            ))}
          </div>
          <BusinessCanvas
            graph={dataGraph(state.dataMode)}
            title={`业务对象与数据归属 · ${state.dataMode}`}
            compact
          />
          <Notice>
            Managed Business Runtime 是轻量业务执行服务，不替代完整
            ERP：必须校验状态迁移、权限、幂等键和事务边界。
          </Notice>
        </>
      )}
      {tab === 'Data Contract' && (
        <div className="split">
          <Panel title="WorkOrder · Canonical Schema">
            <Rows
              headers={['标准字段', '类型 / 约束', '客户字段']}
              rows={Object.entries(mapping).map(([k, v]) => [
                <code key="cell-3435">{k}</code>,
                k === 'priority'
                  ? 'enum: normal | urgent'
                  : 'string · required',
                <Input
                  key="cell-3587"
                  aria-label={k + ' mapping'}
                  value={v}
                  onChange={(e) => {
                    setValid(false);
                    setMapping({ ...mapping, [k]: e.target.value });
                  }}
                />,
              ])}
            />
            <div className="panel-actions">
              <Button
                variant="outline"
                onClick={() => {
                  const values = Object.values(mapping);
                  setValid(
                    values.every((x) => /^[a-zA-Z_][\w.]*$/.test(x)) &&
                      new Set(values).size === values.length,
                  );
                  notify(
                    values.every((x) => /^[a-zA-Z_][\w.]*$/.test(x)) &&
                      new Set(values).size === values.length
                      ? '映射校验通过（本地样例）'
                      : '映射必须非空、合法且不重复',
                  );
                }}
              >
                校验映射
              </Button>
              <Button
                disabled={!valid}
                onClick={() =>
                  update({ mapping }, 'Data Contract 映射已保存', true)
                }
              >
                保存契约
              </Button>
            </div>
          </Panel>
          <Panel title="标准化后的业务数据">
            <pre className="code-block">
              {JSON.stringify(
                {
                  id: 'WO-1024',
                  priority: 'urgent',
                  assignee: 'engineer-008',
                  status: 'assigned',
                  tenant_id: 'northstar',
                },
                null,
                2,
              )}
            </pre>
            <Notice>
              字段重命名由 Adapter 处理，状态枚举必须做显式转换；未知枚举应返回
              CONTRACT_VALIDATION_FAILED。
            </Notice>
          </Panel>
        </div>
      )}
      {tab === '知识存储' && (
        <Panel
          title="Knowledge Storage"
          aside={<Badge>与交易数据物理分离</Badge>}
        >
          <Rows
            headers={['知识集合', '来源 / 版本', '有效期', '访问范围', '状态']}
            rows={[
              [
                '园区服务手册',
                '运营知识库 · 2026.08',
                '2026-12-31',
                '客服 / 工单 Agent',
                <Badge key="cell-5852" tone="success">
                  已审核
                </Badge>,
              ],
              [
                '设备维保标准',
                '设备供应商 · 2.1',
                '2026-11-30',
                '设备 Agent',
                <Badge key="cell-6051" tone="success">
                  可信来源
                </Badge>,
              ],
              [
                '过期招商政策',
                '手工导入 · 2025.09',
                '2026-08-01',
                '禁止检索',
                <Badge key="cell-6250" tone="danger">
                  已过期
                </Badge>,
              ],
            ]}
          />
          <Notice>
            检索时同时过滤
            tenant、ACL、版本和有效期；保留来源片段，不能把检索文本当作执行指令。
          </Notice>
        </Panel>
      )}
      {tab === '临时运行数据' && (
        <Panel title="Temporary Runtime Data">
          <Rows
            headers={['数据类型', '作用域', 'TTL', '到期行为']}
            rows={[
              [
                'Conversation Context',
                'tenant / application / session',
                '24 小时',
                '清除上下文与缓存',
              ],
              ['工具调用中间结果', 'run / step', '1 小时', '删除敏感结果'],
              [
                'Human Task checkpoint',
                '任务绑定',
                '审批完成 + 7 天',
                '清除恢复负载，保留审计摘要',
              ],
              [
                '幂等键',
                'tenant / capability',
                '72 小时',
                '过期后拒绝旧请求重放',
              ],
            ]}
          />
          <Notice>
            示例 TTL 可配置；遇到 Legal Hold 时暂停删除并审计，TTL
            清理不能破坏待审批任务。
          </Notice>
        </Panel>
      )}
      {tab === '隔离与迁移' && (
        <>
          <div className="two-col">
            <Panel title="Tenant 隔离">
              <div className="definition-row">
                Standard → 独立 Tenant Schema
              </div>
              <div className="definition-row">
                Enterprise → Dedicated Database
              </div>
              <div className="definition-row">
                驻留区域 → 客户选择的合规 Region
              </div>
              <div className="definition-row">
                缓存 / 向量索引 / 队列 → 同步隔离
              </div>
            </Panel>
            <Panel title="Managed → External 迁移">
              <ol className="steps">
                <li>冻结写入窗口，导出带版本的数据</li>
                <li>校验目标 Schema、字段映射与权限</li>
                <li>影子读对比、数据条数与摘要校验</li>
                <li>切换 Adapter，保留回滚窗口</li>
              </ol>
              <Button
                variant="outline"
                onClick={() =>
                  notify(
                    '迁移演练通过：4 个对象，0 个 Schema 差异（模拟，未迁移数据）',
                  )
                }
              >
                运行迁移演练
              </Button>
            </Panel>
          </div>
          <Notice tone="warning">
            切换 Adapter
            不代表迁移完成。真实上线需验证写入路由、数据一致性、事件游标、停机窗口和恢复策略。
          </Notice>
        </>
      )}
    </>
  );
}
