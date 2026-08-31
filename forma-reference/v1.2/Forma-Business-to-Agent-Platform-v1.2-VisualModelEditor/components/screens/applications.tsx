import { BusinessCanvas } from '../business-canvas';
import { applicationGraph } from '@/lib/business-canvas';
import { useState } from 'react';
import { Plus, Play, Package, ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { Heading, Panel, Tabs, Badge, Notice, Field, Modal } from '../shared';
import { useStore } from '@/lib/store';

export default function Applications({ go }: { go: (r: string) => void }) {
  const { state, update } = useStore();
  const app = state.application;
  const [tab, setTab] = useState('协作画布');
  const [pick, setPick] = useState(false);
  const [focus, setFocus] = useState('Supervisor');
  const [draft, setDraft] = useState(app);
  const modes = [
    'Router',
    'Supervisor',
    'Pipeline',
    'Parallel',
    'Handoff',
    'Human-in-the-loop',
  ];
  const selected = state.agents.filter(
    (a) => app.selected.includes(a.id) && !a.deleted,
  );
  return (
    <>
      <Heading
        eyebrow="APPLICATION ASSET / SOLUTION BUILDER"
        title={app.name}
        description="一个交付应用，组合多个业务 Agent。锁定依赖、编排协作、共享上下文。"
      >
        <Button variant="outline" onClick={() => go('evaluation')}>
          <Play size={14} />
          应用评测
        </Button>
        <Button onClick={() => go('delivery')}>
          <Package size={14} />
          打包与交付
        </Button>
      </Heading>
      <Tabs
        items={['协作画布', '共享资源', '冲突与降级', '应用配置']}
        value={tab}
        onChange={setTab}
      />
      {tab === '协作画布' && (
        <div className="builder-layout">
          <BusinessCanvas
            graph={applicationGraph(selected, app.mode)}
            title="应用协作 · Business Canvas"
            onSelect={(node) => {
              if (node.id === 'output') setTab('冲突与降级');
              else setFocus(node.id);
            }}
            actions={
              <Button size="sm" variant="outline" onClick={() => setPick(true)}>
                <Plus size={14} />
                管理 Agent
              </Button>
            }
          />
          <Panel
            title={
              focus === 'Supervisor'
                ? '协作配置'
                : state.agents.find((a) => a.id === focus)?.name || '节点配置'
            }
          >
            <Field label="协作模式">
              <div className="mode-options">
                {modes.map((m) => (
                  <button
                    className={app.mode === m ? 'selected' : ''}
                    key={m}
                    onClick={() =>
                      update(
                        { application: { ...app, mode: m } },
                        '应用协作方式改为 ' + m,
                        true,
                      )
                    }
                  >
                    {m}
                  </button>
                ))}
              </div>
            </Field>
            <div className="divider" />
            <h3>执行约束</h3>
            <div className="property-list">
              <p>
                <span>最大协作深度</span>
                <b>6 层</b>
              </p>
              <p>
                <span>总执行预算</span>
                <b>¥ 0.50 / 次</b>
              </p>
              <p>
                <span>最大运行时长</span>
                <b>60 秒</b>
              </p>
              <p>
                <span>Context 作用域</span>
                <b>Application run</b>
              </p>
            </div>
            <Notice>
              业务 Agent 只处理授权的业务范围，不能绕过平台确认策略。
            </Notice>
            <Button variant="outline" onClick={() => go('agents')}>
              查看业务 Agent <ArrowRight size={14} />
            </Button>
          </Panel>
        </div>
      )}
      {tab === '共享资源' && (
        <div className="two-col">
          <Panel title="Shared Context">
            <Field label="允许共享的字段">
              <Textarea
                value={draft.context}
                onChange={(e) =>
                  setDraft({ ...draft, context: e.target.value })
                }
              />
            </Field>
            <Notice>
              共享采用字段白名单。身份与租户信息由 Gateway
              注入，不接受模型改写。
            </Notice>
            <Button
              onClick={() =>
                update(
                  { application: { ...app, context: draft.context } },
                  '共享 Context 已保存',
                  true,
                )
              }
            >
              保存 Context
            </Button>
          </Panel>
          <Panel title="Shared Knowledge">
            <Field label="知识集合">
              <Input
                value={draft.knowledge}
                onChange={(e) =>
                  setDraft({ ...draft, knowledge: e.target.value })
                }
              />
            </Field>
            <div className="definition-row">继承知识文档 ACL 与有效期</div>
            <div className="definition-row">检索响应带来源与版本</div>
            <div className="definition-row">禁止跨 Agent 传播未授权结果</div>
            <Button
              onClick={() =>
                update(
                  { application: { ...app, knowledge: draft.knowledge } },
                  '共享 Knowledge 已保存',
                  true,
                )
              }
            >
              保存 Knowledge
            </Button>
          </Panel>
        </div>
      )}
      {tab === '冲突与降级' && (
        <Panel title="Conflict & Fallback Policy">
          <div className="three-col">
            {[
              [
                '并行写冲突',
                '同一业务对象使用乐观锁；冲突时重新读取，不覆盖他人结果。',
              ],
              [
                'Agent 结论冲突',
                '根据证据与业务权限裁决；无法确定时创建人工任务。',
              ],
              [
                '超时或能力不可用',
                '只读调用可重试；写入结果未知时先 reconcile，再决定补偿。',
              ],
            ].map((x) => (
              <div className="policy-card" key={x[0]}>
                <h3>{x[0]}</h3>
                <p>{x[1]}</p>
              </div>
            ))}
          </div>
          <Field label="降级目标">
            <select
              value={app.fallback}
              onChange={(e) =>
                update(
                  { application: { ...app, fallback: e.target.value } },
                  '降级目标已更新',
                  true,
                )
              }
            >
              <option>转人工任务中心</option>
              <option>只读模式 + 人工通知</option>
              <option>安全终止并说明原因</option>
            </select>
          </Field>
          <Button
            variant="outline"
            onClick={() => {
              update(
                { human: { ...state.human, 'HT-1026': '待处理' } },
                '已模拟工具超时，转交人工任务 HT-1026',
              );
              go('human');
            }}
          >
            模拟超时与人工接管
          </Button>
        </Panel>
      )}
      {tab === '应用配置' && (
        <Panel title="Application Manifest">
          <Field label="应用名称">
            <Input
              value={draft.name}
              onChange={(e) => setDraft({ ...draft, name: e.target.value })}
            />
          </Field>
          <div className="two-col">
            <pre className="code-block">
              {JSON.stringify(
                {
                  name: app.name,
                  version: '1.4.0',
                  agents: selected.map((a) => ({
                    id: a.id,
                    version: a.version,
                  })),
                  orchestration: app.mode,
                  runtime: state.runtime,
                  policy: 'enterprise-safe-v1',
                },
                null,
                2,
              )}
            </pre>
            <div>
              <Notice>
                交付版本锁定 Agent、能力、Schema、知识版本、Behavior
                Policy、Runtime 和评测快照。
              </Notice>
              <Button
                disabled={!draft.name.trim()}
                onClick={() =>
                  update(
                    { application: { ...app, name: draft.name.trim() } },
                    '应用名称已更新',
                    true,
                  )
                }
              >
                保存应用配置
              </Button>
            </div>
          </div>
        </Panel>
      )}
      <Modal
        title="选择要组合的业务 Agent"
        open={pick}
        onClose={() => setPick(false)}
        description="选择结果即时保存在本地草稿，改变组合后必须重新评测。"
      >
        <div className="cap-checks">
          {state.agents
            .filter((a) => !a.deleted)
            .map((a) => (
              <label key={a.id}>
                <Checkbox
                  checked={app.selected.includes(a.id)}
                  onCheckedChange={(checked) =>
                    update(
                      {
                        application: {
                          ...app,
                          selected: checked
                            ? [...app.selected, a.id]
                            : app.selected.filter((id) => id !== a.id),
                        },
                      },
                      '应用 Agent 组合已更新',
                      true,
                    )
                  }
                />
                <span>
                  {a.name}
                  <small>
                    v{a.version} · {a.role}
                  </small>
                </span>
                <Badge>{a.status}</Badge>
              </label>
            ))}
        </div>
        {app.selected.length === 0 && (
          <Notice tone="warning">
            应用至少需要一个 Agent 才能通过发布 Gate。
          </Notice>
        )}
        <Button onClick={() => setPick(false)}>
          完成 · 已选择 {app.selected.length} 个 Agent
        </Button>
      </Modal>
    </>
  );
}
