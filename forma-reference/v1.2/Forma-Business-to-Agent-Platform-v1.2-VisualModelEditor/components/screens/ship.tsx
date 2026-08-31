import { useState } from 'react';
import {
  FlaskConical,
  Play,
  Check,
  LockKeyhole,
  Rocket,
  RotateCcw,
  ArrowRight,
  Download,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';

import {
  Heading,
  Panel,
  Tabs,
  Badge,
  Notice,
  Rows,
  Field,
  Modal,
  Stat,
} from '../shared';
import { gateReasons, download } from '@/lib/domain';
import { useStore } from '@/lib/store';
const cases = [
  ['业务正常路径', '普通工单 → 主管派单 → 客户验收', '业务状态转换符合规则'],
  ['紧急工单', '紧急工单 → 自动分派当班工程师', '规则双人确认 + 自动分派'],
  ['权限隔离', '租户 A 请求租户 B 工单', '拒绝跨租户读取'],
  ['重复请求', '相同幂等键两次创建工单', '仅创建一次'],
  ['人工审批', '高风险派单 → 暂停 → 审批 → 恢复', '不可绕过审批'],
  ['能力不可用', 'Adapter 超时 / 结果未知', '先核对结果，禁止盲重试'],
];
export function Evaluation({ go }: { go: (r: string) => void }) {
  const { state, update } = useStore();
  const [generated, setGenerated] = useState(false);
  const [busy, setBusy] = useState(false);
  const [mock, setMock] = useState('正常响应');
  const [tab, setTab] = useState('业务测试集');
  const ok =
    state.evaluation?.passed && state.evaluation.revision === state.revision;
  return (
    <>
      <Heading
        eyebrow="QUALITY / MODEL-DRIVEN EVALUATION"
        title="用业务事实，验证 Agent 行为。"
        description="从角色、状态、规则和异常路径生成用例，将评测结果绑定到资产快照。"
      >
        <Button
          variant="outline"
          onClick={() => {
            setGenerated(true);
            update({}, '已根据当前 Business Model 生成 24 条示例用例');
          }}
        >
          <FlaskConical size={14} />
          生成业务测试
        </Button>
        <Button
          disabled={busy || !generated}
          onClick={() => {
            const revision = state.revision;
            setBusy(true);
            setTimeout(() => {
              const passed =
                state.approvals.every((a) => a.confirmed) &&
                state.application.selected.length > 0 &&
                mock !== '注入契约破坏';
              update(
                {
                  evaluation: {
                    revision,
                    passed,
                    total: 24,
                    failed: passed ? 0 : mock === '注入契约破坏' ? 4 : 2,
                    at: new Date().toISOString(),
                  },
                },
                passed
                  ? '当前快照 24 / 24 回归通过（模拟）'
                  : '回归失败：查看规则确认或 Mock 契约错误',
              );
              setBusy(false);
            }, 1000);
          }}
        >
          <Play size={14} />
          {busy ? '正在运行回归…' : '运行回归'}
        </Button>
      </Heading>
      <div className="stats-row">
        <Stat
          label="资产快照"
          value={'r' + state.revision}
          sub="配置变化后自动失效"
        />
        <Stat
          label="测试用例"
          value={generated || state.evaluation ? '24' : '—'}
          sub="6 类场景 × 4 个边界变体"
        />
        <Stat
          label="通过率"
          value={
            state.evaluation
              ? (((24 - state.evaluation.failed) / 24) * 100).toFixed(1) + '%'
              : '未运行'
          }
          sub="确定性模拟，不代表真实 LLM 质量"
        />
        <Stat
          label="Release Gate"
          value={ok ? '通过' : '阻断'}
          sub={ok ? '可进入版本冻结' : '待完成确认与当前快照评测'}
        />
      </div>
      <Tabs
        items={['业务测试集', 'Capability Mock', '回归报告']}
        value={tab}
        onChange={setTab}
      />
      {tab === '业务测试集' && (
        <Panel
          title="从 Business Model 派生的测试"
          aside={<Badge>{generated ? '已生成' : '待生成'}</Badge>}
        >
          <Rows
            headers={['场景', '业务路径', '断言', '结果']}
            rows={cases.map((c, i) => [
              ...c,
              <Badge
                key="cell-3755"
                tone={
                  state.evaluation
                    ? i === 1 && !ok
                      ? 'danger'
                      : 'success'
                    : 'neutral'
                }
              >
                {state.evaluation
                  ? i === 1 && !ok
                    ? '存在失败'
                    : '通过'
                  : '待运行'}
              </Badge>,
            ])}
          />
          <Notice>
            生成器为预置场景示例。生产实现需保存 Given / When /
            Then、模型版本、评审状态和 Golden assertions。
          </Notice>
        </Panel>
      )}
      {tab === 'Capability Mock' && (
        <Panel title="故障注入与 Mock">
          <Field label="响应策略">
            <select value={mock} onChange={(e) => setMock(e.target.value)}>
              <option>正常响应</option>
              <option>超时后转人工</option>
              <option>注入契约破坏</option>
            </select>
          </Field>
          <pre className="code-block">
            {JSON.stringify(
              {
                capability: 'work_order.auto_assign',
                fixture: mock,
                expected:
                  mock === '注入契约破坏'
                    ? 'Contract validation must fail'
                    : mock === '超时后转人工'
                      ? 'Human task created'
                      : 'Assigned to on-call engineer',
                noExternalCalls: true,
              },
              null,
              2,
            )}
          </pre>
          <Notice>
            选择“注入契约破坏”可演示失败结果与发布阻断；切回正常响应并重新运行可恢复。
          </Notice>
        </Panel>
      )}
      {tab === '回归报告' && (
        <Panel
          title="Regression Report"
          aside={
            <Button
              variant="outline"
              disabled={!state.evaluation}
              onClick={() =>
                download('evaluation-report.json', {
                  evaluation: state.evaluation,
                  cases,
                  revision: state.revision,
                  mock,
                  simulated: true,
                })
              }
            >
              <Download size={14} />
              导出报告
            </Button>
          }
        >
          {state.evaluation ? (
            <>
              <Rows
                headers={['指标', '结果']}
                rows={[
                  ['快照', `r${state.evaluation.revision}`],
                  ['用例数', 24],
                  ['失败数', state.evaluation.failed],
                  ['记录时间', state.evaluation.at],
                  ['Gate', ok ? '通过' : '阻断'],
                ]}
              />
              {!ok && (
                <Notice tone="warning">
                  {!state.approvals.every((a) => a.confirmed)
                    ? 'BR-021 尚未完成双人确认，请返回业务证据页处理。'
                    : '存在契约错误或配置问题，请修复后重跑。'}
                </Notice>
              )}
              <Button onClick={() => go(ok ? 'releases' : 'business')}>
                {ok ? '进入版本发布' : '处理业务规则'}
                <ArrowRight size={14} />
              </Button>
            </>
          ) : (
            <Notice>尚未运行回归。先生成业务测试集，再运行测试。</Notice>
          )}
        </Panel>
      )}
    </>
  );
}
export function Releases({ go }: { go: (r: string) => void }) {
  const { state, update } = useStore();
  const [confirm, setConfirm] = useState('');
  const reasons = gateReasons(state);
  const gate = reasons.length === 0;
  const stages = ['Dev', 'Test', 'Staging', 'Prod'];
  const index = stages.indexOf(state.stage);
  return (
    <>
      <Heading
        eyebrow="DELIVERY / RELEASE CONTROL"
        title="验证之后，才是发布。"
        description="候选版本 v1.4.0 · 依赖锁定 → Release Gate → Version Freeze → Canary"
      >
        <Badge tone={state.released ? 'success' : gate ? 'success' : 'warning'}>
          {state.released ? '生产灰度中' : gate ? 'Gate 已通过' : '发布被阻断'}
        </Badge>
      </Heading>
      <div className="release-pipeline">
        {stages.map((s, i) => (
          <div key={s} className={i <= index ? 'complete' : ''}>
            <span>{i < index ? <Check size={18} /> : i + 1}</span>
            <h2>{s}</h2>
            <p>{['开发草稿', '回归测试', '发布演练', '生产灰度'][i]}</p>
            <Badge tone={s === state.stage ? 'success' : 'neutral'}>
              {s === state.stage
                ? '当前环境'
                : i < index
                  ? '已通过'
                  : '等待晋级'}
            </Badge>
          </div>
        ))}
      </div>
      <div className="two-col">
        <Panel title="Release Gate 检查">
          <Rows
            headers={['检查项', '状态']}
            rows={[
              [
                '业务规则双人确认',
                <Badge
                  key="cell-8399"
                  tone={
                    state.approvals.every((a) => a.confirmed)
                      ? 'success'
                      : 'warning'
                  }
                >
                  {state.approvals.every((a) => a.confirmed)
                    ? '通过'
                    : '待确认'}
                </Badge>,
              ],
              [
                '当前快照回归',
                <Badge
                  key="cell-8817"
                  tone={state.evaluation?.passed ? 'success' : 'warning'}
                >
                  {state.evaluation?.passed ? '通过' : '待运行 / 失败'}
                </Badge>,
              ],
              [
                'Agent 依赖完整',
                state.application.selected.length + ' 个 Agent',
              ],
              [
                'Version Freeze',
                state.frozen === state.revision
                  ? 'r' + state.frozen + ' 已冻结'
                  : '未冻结',
              ],
            ]}
          />
          {reasons.map((r) => (
            <Notice key={r} tone="warning">
              {r}
            </Notice>
          ))}
          <div className="panel-actions">
            <Button
              variant="outline"
              onClick={() =>
                go(
                  !state.approvals.every((a) => a.confirmed)
                    ? 'business'
                    : 'evaluation',
                )
              }
            >
              处理 Gate
            </Button>
            <Button
              disabled={!gate || state.frozen === state.revision}
              onClick={() =>
                update(
                  { frozen: state.revision },
                  '当前快照已冻结为 r' + state.revision,
                )
              }
            >
              <LockKeyhole size={14} />
              冻结版本
            </Button>
          </div>
        </Panel>
        <Panel title="发布策略">
          <Field label="Canary 流量比例">
            <input
              type="range"
              min="1"
              max="100"
              value={state.canary}
              onChange={(e) =>
                update(
                  { canary: Number(e.target.value) },
                  '灰度比例调整为 ' + e.target.value + '%',
                )
              }
            />
            <b>{state.canary}%</b>
          </Field>
          <div className="property-list">
            <p>
              <span>错误率阈值</span>
              <b>≤ 1%</b>
            </p>
            <p>
              <span>关键业务失败</span>
              <b>任何失败即阻断</b>
            </p>
            <p>
              <span>观察窗口</span>
              <b>30 分钟</b>
            </p>
            <p>
              <span>上一生产版本</span>
              <b>v1.3.0</b>
            </p>
          </div>
          <Notice>
            所有发布动作均为本地演示。回滚仅切换应用版本，不自动逆转已发生的业务写入。
          </Notice>
          <div className="actions">
            <Button
              disabled={
                !gate ||
                state.frozen !== state.revision ||
                state.stage === 'Prod'
              }
              onClick={() => {
                if (state.stage === 'Staging') setConfirm('publish');
                else
                  update(
                    { stage: stages[index + 1] },
                    '候选版本已晋级 ' + stages[index + 1],
                  );
              }}
            >
              <Rocket size={14} />
              {state.stage === 'Staging'
                ? '发布至 Prod'
                : state.stage === 'Prod'
                  ? '已发布'
                  : '晋级至 ' + stages[index + 1]}
            </Button>
            <Button
              variant="outline"
              disabled={!state.released}
              onClick={() => setConfirm('rollback')}
            >
              <RotateCcw size={14} />
              回滚
            </Button>
          </div>
        </Panel>
      </div>
      <Panel title="版本记录">
        <Rows
          headers={['版本', '环境', '流量', '状态']}
          rows={[
            [
              'v1.4.0',
              state.stage,
              state.released ? state.canary + '%' : '0%',
              <Badge
                key="cell-12491"
                tone={state.released ? 'success' : 'neutral'}
              >
                {state.released ? '灰度发布' : '候选版本'}
              </Badge>,
            ],
            [
              'v1.3.0',
              'Prod',
              state.released ? 100 - state.canary + '%' : '100%',
              <Badge key="cell-12775" tone="success">
                稳定版本
              </Badge>,
            ],
            ['v1.2.0', 'Archive', '0%', '历史归档'],
          ]}
        />
      </Panel>
      <Modal
        title={confirm === 'publish' ? '确认模拟生产发布' : '确认模拟回滚'}
        open={!!confirm}
        onClose={() => setConfirm('')}
      >
        <p>
          {confirm === 'publish'
            ? `将冻结快照 r${state.frozen} 以 ${state.canary}% 流量发布到模拟 Prod。`
            : '将模拟生产流量恢复到 v1.3.0。业务数据不会被逆向迁移。'}
        </p>
        <Button
          onClick={() => {
            if (confirm === 'publish') {
              if (gate && state.frozen === state.revision)
                update(
                  { stage: 'Prod', released: true },
                  'v1.4.0 已模拟灰度发布至 Prod',
                );
            } else
              update(
                { stage: 'Staging', released: false },
                '已回滚至稳定版本 v1.3.0（模拟）',
              );
            setConfirm('');
          }}
        >
          确认{confirm === 'publish' ? '发布' : '回滚'}
        </Button>
      </Modal>
    </>
  );
}
export function HumanTasks() {
  const { state, update } = useStore();
  const [selected, setSelected] = useState('HT-1024');
  const [reason, setReason] = useState('已核对业务范围，允许本次操作。');
  const [filter, setFilter] = useState('待处理');
  const tasks = [
    {
      id: 'HT-1024',
      title: '紧急工单跨班组派单',
      source: '工单 Agent / work_order.assign',
      risk: '高风险',
      due: '18 分钟',
      detail:
        '工单 WO-1024，A 座电梯故障。建议由夜班工程师接管，涉及跨班组权限。',
    },
    {
      id: 'HT-1025',
      title: '业务规则例外审批',
      source: '园区智能助手 / exception',
      risk: '业务例外',
      due: '42 分钟',
      detail: '客户希望跳过普通验收流程，需要主管根据服务条款裁决。',
    },
    {
      id: 'HT-1026',
      title: 'Adapter 超时人工接管',
      source: '工单 Agent / fallback',
      risk: '执行结果未知',
      due: '60 分钟',
      detail: '外部派单接口未及时返回，请核对客户系统中的派单记录后再处理。',
    },
  ];
  const item = tasks.find((x) => x.id === selected)!;
  return (
    <>
      <Heading
        eyebrow="OPERATIONS / HUMAN TASK CENTER"
        title="在人与 Agent 之间，保留判断。"
        description="统一审批、异常接管与超时升级。审批结果绑定具体执行快照。"
      >
        <Badge tone="warning">
          {Object.values(state.human).filter((s) => s === '待处理').length}{' '}
          项待办
        </Badge>
      </Heading>
      <Tabs
        items={['待处理', '全部任务']}
        value={filter}
        onChange={setFilter}
      />
      <div className="split">
        <Panel title="任务队列">
          {tasks
            .filter(
              (t) => filter === '全部任务' || state.human[t.id] === '待处理',
            )
            .map((t) => (
              <button
                className={'task-item ' + (selected === t.id ? 'selected' : '')}
                key={t.id}
                onClick={() => setSelected(t.id)}
              >
                <span>
                  <b>{t.title}</b>
                  <small>
                    {t.id} · {t.source}
                  </small>
                </span>
                <Badge
                  tone={state.human[t.id] === '待处理' ? 'warning' : 'success'}
                >
                  {state.human[t.id]}
                </Badge>
              </button>
            ))}
          {tasks.every((t) => state.human[t.id] !== '待处理') &&
            filter === '待处理' && (
              <Notice>当前队列已处理完毕，可切换“全部任务”查看记录。</Notice>
            )}
        </Panel>
        <Panel
          title={item.title}
          aside={<Badge tone="warning">{item.risk}</Badge>}
        >
          <p className="paragraph">{item.detail}</p>
          <div className="property-list">
            <p>
              <span>业务对象</span>
              <b>WO-1024 / northstar</b>
            </p>
            <p>
              <span>审批角色</span>
              <b>物业主管</b>
            </p>
            <p>
              <span>到期</span>
              <b>{item.due}（示例）</b>
            </p>
            <p>
              <span>当前状态</span>
              <b>{state.human[selected]}</b>
            </p>
          </div>
          <Field label="决策理由（必填）">
            <Textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </Field>
          <div className="actions">
            <Button
              disabled={!reason.trim() || state.human[selected] !== '待处理'}
              onClick={() =>
                update(
                  { human: { ...state.human, [selected]: '已批准' } },
                  selected + ' 批准并恢复模拟执行：' + reason,
                )
              }
            >
              批准并恢复
            </Button>
            <Button
              variant="destructive"
              disabled={!reason.trim() || state.human[selected] !== '待处理'}
              onClick={() =>
                update(
                  { human: { ...state.human, [selected]: '已拒绝' } },
                  selected + ' 已拒绝：' + reason,
                )
              }
            >
              拒绝
            </Button>
            <Button
              variant="outline"
              disabled={state.human[selected] !== '待处理'}
              onClick={() =>
                update(
                  { human: { ...state.human, [selected]: '已升级' } },
                  selected + ' 已转交上级主管',
                )
              }
            >
              升级处理
            </Button>
          </div>
          <Notice>
            生产审批必须校验审批人权限、截止时间、请求摘要与资源版本；重复审批不可重复执行。
          </Notice>
        </Panel>
      </div>
    </>
  );
}
