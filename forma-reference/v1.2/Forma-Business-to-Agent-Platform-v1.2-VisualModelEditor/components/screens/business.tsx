import { BusinessCanvas } from '../business-canvas';
import { VisualBusinessModelEditor } from '../visual-business-model-editor';
import { createVisualModel, canvasGraph } from '@/lib/visual-model';
import { workOrderSummary, impactGraph } from '@/lib/business-canvas';
import { useState } from 'react';
import { Sparkles, ArrowRight, Check, GitBranch, Network } from 'lucide-react';
import { Button } from '@/components/ui/button';

import { Textarea } from '@/components/ui/textarea';
import { Heading, Panel, Tabs, Badge, Notice, Rows, Field } from '../shared';
import { useStore } from '@/lib/store';
export function Analyst({ go }: { go: (r: string) => void }) {
  const { state, update } = useStore();
  const [pattern, setPattern] = useState('Ticket');
  const [text, setText] = useState(
    '紧急工单直接自动派给当班工程师，普通工单仍由主管分派。',
  );
  const [busy, setBusy] = useState(false);
  const patterns = [
    'Ticket',
    'Approval',
    'Inspection',
    'Incident',
    'Fulfillment',
    'Reservation',
    'Order',
    'Case',
    'Asset Lifecycle',
    'Customer Service',
  ];
  return (
    <>
      <Heading
        eyebrow="DISCOVER / AI BUSINESS ANALYST"
        title="先理解业务，再构建 Agent。"
        description="以业务模式驱动访谈，让每一个业务结论都有来源。"
      >
        <Badge tone="warning">AI 提取 · 人工确认</Badge>
      </Heading>
      <div className="split analyst-grid">
        <Panel title="业务模式库" aside={<Badge>10 Patterns</Badge>}>
          <p className="muted">
            选择通用业务模式，生成对应的角色、状态、规则和异常问题。
          </p>
          <div className="pattern-list">
            {patterns.map((p, i) => (
              <button
                className={pattern === p ? 'pattern active' : 'pattern'}
                onClick={() => setPattern(p)}
                key={p}
              >
                <span className="icon-tile blue">
                  <Network size={17} />
                </span>
                <span>
                  <b>{p}</b>
                  <small>
                    {
                      [
                        '工单受理与履约',
                        '审批与授权',
                        '巡检与异常',
                        '事件响应',
                        '服务交付',
                        '资源预约',
                        '订单生命周期',
                        '个案管理',
                        '设备全生命周期',
                        '客户服务',
                      ][i]
                    }
                  </small>
                </span>
                {pattern === p && <Check size={15} />}
              </button>
            ))}
          </div>
        </Panel>
        <Panel
          title="园区物业 · 业务发现会话"
          aside={<Badge tone="success">访谈进行中</Badge>}
        >
          <div className="chat-thread">
            <div className="chat assistant">
              <span className="ai-orb">
                <Sparkles size={17} />
              </span>
              <div>
                <b>AI 业务分析师</b>
                <p>
                  已选择 {pattern}{' '}
                  模式。我们先明确三个问题：谁发起业务、何时完成、哪些情况需要人工介入？
                </p>
                <div className="chips">
                  <span>角色</span>
                  <span>业务对象</span>
                  <span>状态机</span>
                  <span>异常规则</span>
                </div>
              </div>
            </div>
            <div className="chat user">
              <div>
                <b>李经理 · 物业运营</b>
                <p>
                  现在由客服登记工单，主管派单，工程师完成后由客户确认。紧急情况希望更快处理。
                </p>
              </div>
              <span className="avatar">李</span>
            </div>
            <div className="chat assistant">
              <span className="ai-orb">
                <Sparkles size={17} />
              </span>
              <div>
                <b>需要澄清的决策</b>
                <p>
                  {pattern === 'Ticket'
                    ? '紧急工单是否可以跳过主管？自动派单失败时，是否转交人工？'
                    : '请描述 ' +
                      pattern +
                      ' 的完成条件、负责人以及超时后的处理方式。'}
                </p>
                <small>来源：本轮访谈 · 尚未视为已确认事实</small>
              </div>
            </div>
            {state.discovery.map((m, i) => (
              <div className="chat user" key={i}>
                <div>
                  <b>补充业务说明</b>
                  <p>{m}</p>
                </div>
              </div>
            ))}
            {state.discovery.length > 0 && (
              <Notice>
                已保存访谈证据。下一步在业务资产中审阅候选规则，并由业务与 IT
                负责人分别确认。
                <Button variant="link" onClick={() => go('business')}>
                  审阅候选模型 <ArrowRight size={14} />
                </Button>
              </Notice>
            )}
          </div>
          <div className="composer">
            <Textarea
              aria-label="业务说明"
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="描述你的业务规则、异常情况和约束…"
            />
            <div className="composer-footer">
              <small>演示提取器 · 不连接真实语言模型</small>
              <Button
                disabled={!text.trim() || busy}
                onClick={() => {
                  setBusy(true);
                  setTimeout(() => {
                    update(
                      {
                        discovery: [...state.discovery, text],
                        rule: text,
                        approvals: state.approvals.map((a) => ({
                          ...a,
                          confirmed: false,
                        })),
                      },
                      '业务说明已转为待确认候选规则',
                      true,
                    );
                    setText('');
                    setBusy(false);
                  }, 650);
                }}
              >
                <Sparkles size={14} />
                {busy ? '提取候选事实…' : '提取业务事实'}
              </Button>
            </div>
          </div>
        </Panel>
      </div>
      <BusinessCanvas
        graph={
          state.visualModel
            ? canvasGraph(state.visualModel)
            : workOrderSummary(
                state.rule,
                state.approvals.every((a) => a.confirmed),
              )
        }
        title="候选业务模型 · 当前工单访谈草稿"
        compact
      />
    </>
  );
}
export function Business({ go }: { go: (r: string) => void }) {
  const { state, update } = useStore();
  const [tab, setTab] = useState('业务画布');
  const [ruleDraft, setRuleDraft] = useState({
    base: state.rule,
    value: state.rule,
  });
  const rule = ruleDraft.base === state.rule ? ruleDraft.value : state.rule;
  const setRule = (value: string) => setRuleDraft({ base: state.rule, value });
  const confirmed = state.approvals.every((a) => a.confirmed);
  return (
    <>
      <Heading
        eyebrow="BUSINESS ASSET / WORK ORDER"
        title="园区工单业务模型"
        description="Business Model · BM-001 · 负责人：李经理 · 由 Ticket Pattern 构建"
      >
        <Badge tone={confirmed ? 'success' : 'warning'}>
          {confirmed ? '已达成业务共识' : '1 项规则待确认'}
        </Badge>
      </Heading>
      <Tabs
        items={['业务画布', '证据与确认', '版本与 Diff', '变更影响']}
        value={tab}
        onChange={setTab}
      />
      {tab === '业务画布' && (
        <>
          <div className="model-summary">
            <span className="icon-tile blue">
              <Network />
            </span>
            <div>
              <h2>一次报修，完整的服务闭环</h2>
              <p className="muted">
                从报修受理到客户验收 · 可编辑结构化业务模型 · 语义与布局独立保存
              </p>
            </div>
            <Badge>草稿快照 r{state.revision}</Badge>
          </div>
          <VisualBusinessModelEditor
            value={state.visualModel ?? createVisualModel(state.rule)}
            onChange={(visualModel, action) => update({ visualModel }, action)}
            onImpact={() => setTab('变更影响')}
          />
          <Notice tone="warning">
            紧急路径依赖候选规则：{state.rule}{' '}
            <Button variant="link" onClick={() => setTab('证据与确认')}>
              查看证据 →
            </Button>
          </Notice>
        </>
      )}
      {tab === '证据与确认' && (
        <div className="split">
          <Panel
            title="候选规则 #BR-021"
            aside={
              <Badge tone={confirmed ? 'success' : 'warning'}>
                {confirmed ? '已确认' : '待确认'}
              </Badge>
            }
          >
            <Field label="业务结论">
              <Textarea
                value={rule}
                onChange={(e) => setRule(e.target.value)}
              />
            </Field>
            <Button
              variant="outline"
              disabled={!rule.trim() || rule === state.rule}
              onClick={() =>
                update(
                  {
                    rule,
                    approvals: state.approvals.map((a) => ({
                      ...a,
                      confirmed: false,
                    })),
                  },
                  '规则已更新，原有确认失效',
                  true,
                )
              }
            >
              保存候选规则
            </Button>
            <div className="evidence-card">
              <div className="section-line">
                <h3>Evidence #EV-021</h3>
                <Badge tone="success">Confidence · 0.92</Badge>
              </div>
              <blockquote>
                “紧急的不用再等我来派，让当班师傅直接接。”
              </blockquote>
              <p className="muted">李经理 / 物业运营负责人 · 2026-08-31 访谈</p>
            </div>
            <div className="evidence-card conflict">
              <h3>冲突证据 #EV-018</h3>
              <blockquote>“所有工单必须经主管派单。”</blockquote>
              <p className="muted">王工 / 现行业务流程说明 v1.3</p>
            </div>
            <Notice>
              置信度不是审批。以下双人确认表示同意用新规则替代旧规则；修改规则后必须重新确认。
            </Notice>
          </Panel>
          <Panel
            title="多人确认"
            aside={
              <Badge>
                {state.approvals.filter((a) => a.confirmed).length} / 2
              </Badge>
            }
          >
            <p className="muted">
              原型可模拟不同审批人；生产系统必须通过登录身份与职责分离校验。
            </p>
            {state.approvals.map((a, i) => (
              <div className="approval-row" key={a.role}>
                <span className="avatar">{i ? '王' : '李'}</span>
                <div>
                  <b>{a.role}</b>
                  <small>
                    {a.confirmed ? '已确认当前规则' : '等待审阅来源及冲突'}
                  </small>
                </div>
                <Button
                  variant={a.confirmed ? 'outline' : 'default'}
                  disabled={a.confirmed || rule !== state.rule}
                  onClick={() =>
                    update(
                      {
                        approvals: state.approvals.map((p, j) =>
                          j === i ? { ...p, confirmed: true } : p,
                        ),
                      },
                      a.role + ' 已确认当前业务规则',
                      true,
                    )
                  }
                >
                  {a.confirmed ? '已确认' : '模拟确认'}
                </Button>
              </div>
            ))}
            <div className="divider" />
            <h3>确认后的下一步</h3>
            <p className="muted paragraph">
              追踪受影响的能力和
              Agent，对新的资产快照生成回归测试，旧生产版本保持不变。
            </p>
            <Button onClick={() => setTab('变更影响')}>
              查看变更影响 <ArrowRight size={15} />
            </Button>
          </Panel>
        </div>
      )}
      {tab === '版本与 Diff' && (
        <Panel
          title="生产 v1.3 → 当前候选版本"
          aside={<Badge>语义 Diff</Badge>}
        >
          {state.visualModel?.revisions
            .slice()
            .reverse()
            .map((r) => (
              <p key={r.revision}>
                Business Model r{r.revision} · {r.action} ·{' '}
                {r.semantic_model.nodes.length} 节点 /{' '}
                {r.semantic_model.edges.length} 关系
              </p>
            ))}
          <div className="diff">
            <div>
              <small>− v1.3 / 生产冻结版本</small>
              <p>所有工单必须经主管派单。</p>
              <p>自动派单：关闭</p>
            </div>
            <div>
              <small>+ 当前草稿 / r{state.revision}</small>
              <p>{state.rule}</p>
              <p>自动派单：仅紧急工单；失败转人工</p>
            </div>
          </div>
          <Rows
            headers={['版本', '内容', '状态']}
            rows={[
              [
                `r${state.revision}`,
                '当前资产快照，包含最新规则与审批',
                <Badge key="cell-12998" tone="warning">
                  草稿
                </Badge>,
              ],
              [
                'v1.3',
                '主管统一派单 · 2026-08-20',
                <Badge key="cell-13144" tone="success">
                  生产冻结
                </Badge>,
              ],
              ['v1.2', '新增客户验收节点', '归档'],
            ]}
          />
        </Panel>
      )}
      {tab === '变更影响' && (
        <>
          {state.visualModel?.impact && (
            <Notice tone="warning">
              Business Model r{state.visualModel.impact.revision}：变更对象{' '}
              {state.visualModel.impact.changed.join('、')}
              。当前确认、评测、冻结已失效；请审阅依赖并重新确认与回归。
            </Notice>
          )}
          <Notice>
            影响分析示例来自显式依赖关系。生产版本 v1.3
            不受影响；修改后评测与冻结凭证自动失效。
          </Notice>
          <BusinessCanvas
            graph={impactGraph()}
            title="业务变更 · 四层资产影响"
            compact
            onNavigate={go}
          />
          <Panel title="变更计划">
            <Rows
              headers={['受影响资产', '需要执行的动作', '影响类型']}
              rows={[
                ['work_order.assign', '保留普通工单路径', '行为兼容'],
                [
                  'work_order.auto_assign',
                  '增加值班与超时前置条件',
                  '新增能力',
                ],
                ['工单 Agent', '引用新能力版本并重跑回归', '依赖升级'],
                ['园区智能助手', '重新锁定依赖、验证、冻结', '新发布候选'],
              ]}
            />
            <div className="panel-actions">
              <Button onClick={() => go('evaluation')}>
                <GitBranch size={15} />
                进入回归评测
              </Button>
            </div>
          </Panel>
        </>
      )}
    </>
  );
}
