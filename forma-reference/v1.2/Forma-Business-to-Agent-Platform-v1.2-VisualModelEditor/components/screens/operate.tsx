import { useState } from 'react';
import { Radio, Check, ShieldCheck, Cpu, Download } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

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
import { useStore } from '@/lib/store';
import { download } from '@/lib/domain';
export function Channels() {
  const { state, update } = useStore();
  const [channel, setChannel] = useState('');
  const [actor, setActor] = useState('channel_user_id → enterprise_actor_id');
  const [vault, setVault] = useState('vault://northstar/channel/credential');
  const [error, setError] = useState('');
  const all = [
    'Web / H5',
    '微信服务号',
    '企业微信',
    '飞书',
    '钉钉',
    'Slack',
    'Teams',
    'App SDK',
    'API / Webhook',
  ];
  return (
    <>
      <Heading
        eyebrow="CHANNEL GATEWAY / OMNICHANNEL"
        title="一个应用，触达每个业务现场。"
        description="渠道只负责消息适配与身份入口，业务规则和权限由平台统一执行。"
      >
        <Badge>{Object.keys(state.channels).length} / 9 个渠道已配置</Badge>
      </Heading>
      <Notice>
        下列渠道均为配置原型，没有建立真实连接。正式接入须完成回调签名校验、平台审核、权限授权和身份绑定。
      </Notice>
      <div className="agent-grid">
        {all.map((c, i) => (
          <article key={c} className="channel-card">
            <span className={'icon-tile ' + ['blue', 'teal', 'violet'][i % 3]}>
              <Radio size={23} />
            </span>
            <div className="section-line">
              <h2>{c}</h2>
              <Badge tone={state.channels[c] ? 'success' : 'neutral'}>
                {state.channels[c] ? '已配置' : '未配置'}
              </Badge>
            </div>
            <p>
              {
                [
                  '嵌入网站与移动端业务入口',
                  '服务号会话与模板消息适配',
                  '组织身份与群聊协作',
                  '组织成员与审批消息',
                  '工作通知与会话交互',
                  'Workspace 与用户映射',
                  'Entra 身份与会话适配',
                  '集成到客户原生 App',
                  '标准化服务端调用',
                ][i]
              }
            </p>
            <div className="channel-bottom">
              <span>身份映射 · Policy 继承</span>
              <Button
                variant="outline"
                onClick={() => {
                  setChannel(c);
                  setError('');
                }}
              >
                配置渠道
              </Button>
            </div>
          </article>
        ))}
      </div>
      <Panel title="Channel → Identity → Policy → Application">
        <div className="process-flow">
          {[
            '验证渠道签名',
            '映射租户与身份',
            '应用权限与限流',
            '执行版本化应用',
          ].map((x, i) => (
            <div className="process-item" key={x}>
              <span>{i + 1}</span>
              <h3>{x}</h3>
            </div>
          ))}
        </div>
        <Notice>
          未知渠道身份默认拒绝写操作，并引导企业身份绑定；不得仅凭昵称或手机号自动授予业务权限。
        </Notice>
      </Panel>
      <Modal
        title={channel + ' · 渠道配置'}
        open={!!channel}
        onClose={() => setChannel('')}
      >
        <Field label="目标应用">
          <Input readOnly value={state.application.name} />
        </Field>
        <Field label="身份映射">
          <Input value={actor} onChange={(e) => setActor(e.target.value)} />
        </Field>
        <Field label="Credential 引用（不接受真实 Secret）">
          <Input
            value={vault}
            onChange={(e) => setVault(e.target.value)}
            placeholder="vault://tenant/channel/key"
          />
        </Field>
        <Field label="回调校验">
          <Input readOnly value="签名验证 + 重放保护 + tenant binding" />
        </Field>
        {error && <Notice tone="warning">{error}</Notice>}
        <Button
          onClick={() => {
            if (!vault.startsWith('vault://') || !actor.trim()) {
              setError('请输入 Vault 引用与身份映射；不要输入真实密钥。');
              return;
            }
            update(
              { channels: { ...state.channels, [channel]: actor } },
              channel + ' 配置已保存（未连接真实服务）',
              true,
            );
            setChannel('');
          }}
        >
          保存渠道配置
        </Button>
        {state.channels[channel] && (
          <Button
            variant="destructive"
            onClick={() => {
              const next = { ...state.channels };
              delete next[channel];
              update({ channels: next }, channel + ' 已移除配置', true);
              setChannel('');
            }}
          >
            移除配置
          </Button>
        )}
      </Modal>
    </>
  );
}
export function Runtime() {
  const { state, update } = useStore();
  const [tab, setTab] = useState('Runtime Adapter');
  return (
    <>
      <Heading
        eyebrow="PLATFORM / RUNTIME & AGENT KERNEL"
        title="统一业务行为，开放执行底座。"
        description="Conversation、Context、Memory、Tool Calling、Streaming 和 Agent Loop 复用 Runtime。"
      >
        <Badge>当前候选：{state.runtime}</Badge>
      </Heading>
      <Tabs
        items={['Runtime Adapter', 'Platform Kernel', 'Behavior Policy']}
        value={tab}
        onChange={setTab}
      />
      {tab === 'Runtime Adapter' && (
        <>
          <div className="three-col">
            {[
              ['LangGraph', '图编排与 checkpoint', '候选适配 · 未完成认证'],
              [
                'Coze / Eino',
                '应用平台 / Go 编排框架',
                '两个独立适配器，分别评估',
              ],
              [
                'DeepSeek Harness',
                '用户指定的执行底座候选',
                '具体仓库、协议与版本待明确',
              ],
            ].map((x) => (
              <button
                className={
                  'mode-card ' + (state.runtime === x[0] ? 'selected' : '')
                }
                onClick={() =>
                  update(
                    { runtime: x[0] },
                    'Runtime 候选已改为 ' + x[0] + '，需重新评测',
                    true,
                  )
                }
                key={x[0]}
              >
                <span className="icon-tile violet">
                  <Cpu size={22} />
                </span>
                <h2>
                  {x[0]}
                  {state.runtime === x[0] && <Check size={15} />}
                </h2>
                <h3>{x[1]}</h3>
                <p>{x[2]}</p>
              </button>
            ))}
          </div>
          <Panel title="平台适配契约 · 验收矩阵">
            <Rows
              headers={['能力', '平台要求', '适配验证']}
              rows={[
                ['run / stream / cancel', '统一事件与取消语义', '待端到端验证'],
                [
                  'checkpoint / resume',
                  '审批后恢复，同一业务执行 ID',
                  '需持久化与幂等验证',
                ],
                [
                  'tool invocation',
                  '贯穿身份、权限与 trace',
                  '必须经过 Capability Gateway',
                ],
                [
                  'memory scopes',
                  'tenant / agent / application',
                  '禁止跨租户串读',
                ],
                [
                  'version & replay',
                  '锁定模型、Prompt 与 Runtime',
                  '回放禁止重复外部写入',
                ],
                ['fallback', '能力缺失时拒绝或明确降级', '禁止假装支持'],
              ]}
            />
          </Panel>
          <Notice tone="warning">
            这里是可兼容架构与配置交互，并非已经实现上述框架的连接。DeepSeek
            模型 API 与“DeepSeek Harness”不能等同；需先明确后者的实际实现。
          </Notice>
        </>
      )}
      {tab === 'Platform Kernel' && (
        <>
          <div className="kernel-stack">
            <div>
              <small>APPLICATION LAYER</small>
              <h2>
                业务 Role · Context · Capability · Rule · Permission ·
                Interaction
              </h2>
            </div>
            <div className="kernel-policy">
              <small>PLATFORM AGENT KERNEL</small>
              <h2>Behavior Policy / 身份 / 预算 / 审批 / 审计 / Trace</h2>
            </div>
            <div>
              <small>OPEN RUNTIME</small>
              <h2>
                Conversation · Context · Memory · Tool Calling · Streaming ·
                Agent Loop
              </h2>
            </div>
            <div>
              <small>INFRASTRUCTURE</small>
              <h2>Model Gateway · Data Plane · Secret Vault · Task Queue</h2>
            </div>
          </div>
          <Notice>
            平台仍需调试并评测通用行为。复用开源 Runtime
            不等于无需治理、无需安全验证或无需优化交互。
          </Notice>
        </>
      )}
      {tab === 'Behavior Policy' && (
        <Panel title="Enterprise Safe · v1">
          <Rows
            headers={['策略层', '行为要求', '可覆盖性']}
            rows={[
              [
                '平台强制',
                '租户隔离、敏感信息、审计、工具调用安全',
                '不可覆盖',
              ],
              ['组织策略', '数据驻留、保留期、审批阈值、成本预算', '只能收紧'],
              ['应用策略', '协作范围、共享字段、失败处理', '必须符合上层'],
              ['Agent 交互', '业务澄清、话术、引导与回复结构', '允许业务定制'],
            ]}
          />
          <div className="three-col">
            {['不编造工具结果', '先澄清后执行', '高风险操作先确认'].map((x) => (
              <div className="policy-card" key={x}>
                <ShieldCheck size={22} />
                <h3>{x}</h3>
                <p>由 Kernel 与 Gateway 强制执行，并作为评测用例。</p>
              </div>
            ))}
          </div>
        </Panel>
      )}
    </>
  );
}
export function Observability() {
  const [tab, setTab] = useState('执行追踪');
  const [trace, setTrace] = useState('TR-2108');
  const [query, setQuery] = useState('');
  const traces = [
    ['TR-2108', '紧急工单自动派单', '工单 Agent', '1.82s', '¥0.038', '成功'],
    ['TR-2107', '设备档案查询', '设备 Agent', '0.64s', '¥0.012', '成功'],
    ['TR-2106', '高风险派单审批', '工单 Agent', '等待人工', '¥0.021', '暂停'],
    ['TR-2105', '外部接口请求超时', '工单 Agent', '10.0s', '¥0.017', '已降级'],
  ];
  return (
    <>
      <Heading
        eyebrow="OPERATIONS / OBSERVABILITY"
        title="从一次调用，看见业务结果。"
        description="业务 KPI、Agent Trace、Tool Trace 与成本使用统一 Run ID 关联。"
      >
        <Badge>最近 24 小时 · 示例数据</Badge>
      </Heading>
      <div className="stats-row">
        <Stat
          label="业务任务成功率"
          value="98.6%"
          sub="已完成 2,486 / 2,521 次"
        />
        <Stat
          label="平均响应时间"
          value="1.82s"
          sub="P95 4.6s · 不含人工等待"
        />
        <Stat label="当日成本" value="¥ 86.42" sub="模型 + 工具 + 执行资源" />
        <Stat
          label="人工接管率"
          value="3.2%"
          sub="81 次 · 其中安全审批 52 次"
        />
      </div>
      <Tabs
        items={['执行追踪', '成本分析', '业务 KPI']}
        value={tab}
        onChange={setTab}
      />
      {tab === '执行追踪' && (
        <>
          <Panel
            title="Agent Runs"
            aside={
              <Input
                aria-label="搜索 Trace"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="搜索 ID 或业务任务"
                className="trace-search"
              />
            }
          >
            <Rows
              headers={[
                'Trace ID',
                '任务',
                'Agent',
                '时延',
                '成本',
                '状态',
                '操作',
              ]}
              rows={traces
                .filter((t) => t.join(' ').includes(query))
                .map((t) => [
                  ...t.slice(0, 5),
                  <Badge
                    key="cell-11463"
                    tone={t[5] === '成功' ? 'success' : 'warning'}
                  >
                    {t[5]}
                  </Badge>,
                  <Button
                    key="cell-11589"
                    variant="link"
                    onClick={() => setTrace(t[0])}
                  >
                    查看 Trace
                  </Button>,
                ])}
            />
          </Panel>
          <Panel
            title={trace + ' · 执行瀑布'}
            aside={
              <Badge>
                {trace === 'TR-2106'
                  ? '等待审批'
                  : trace === 'TR-2105'
                    ? '超时降级'
                    : '执行完成'}
              </Badge>
            }
          >
            <div className="trace-waterfall">
              {[
                'Channel Gateway · identity.verify',
                'Kernel · behavior_policy.check',
                'Agent · business_reasoning',
                'Capability · contract.validate',
                'Adapter · execute',
                'Audit · outcome.record',
              ].map((x, i) => (
                <div key={x}>
                  <span>{x}</span>
                  <div>
                    <i
                      style={{
                        marginLeft: i * 9 + '%',
                        width: [12, 16, 37, 12, 24, 9][i] + '%',
                      }}
                    />
                  </div>
                  <small>{[42, 61, 830, 31, 810, 46][i]}ms</small>
                </div>
              ))}
            </div>
            <Notice>
              {trace === 'TR-2105'
                ? '失败根因：外部适配器超时；执行结果未知，已转人工核对。'
                : trace === 'TR-2106'
                  ? '当前检查点：高风险写操作等待物业主管确认，尚未执行工具。'
                  : '策略校验与业务执行成功，输出包含 audit_id。'}{' '}
              Trace 中的客户手机号与敏感内容应脱敏。
            </Notice>
          </Panel>
        </>
      )}
      {tab === '成本分析' && (
        <Panel title="按 Agent 归集成本">
          <Rows
            headers={['Agent', '模型成本', '工具成本', '总额', '预算状态']}
            rows={[
              [
                '工单 Agent',
                '¥32.18',
                '¥6.02',
                '¥38.20',
                <Badge key="cell-13543" tone="success">
                  预算内
                </Badge>,
              ],
              [
                '设备 Agent',
                '¥20.41',
                '¥2.55',
                '¥22.96',
                <Badge key="cell-13732" tone="success">
                  预算内
                </Badge>,
              ],
              [
                '客服 Agent',
                '¥19.08',
                '¥1.11',
                '¥20.19',
                <Badge key="cell-13921" tone="success">
                  预算内
                </Badge>,
              ],
              ['其他', '¥4.07', '¥1.00', '¥5.07', '预算内'],
            ]}
          />
          <Notice>
            成本与业务结果关联，避免仅优化 Token 数而降低业务完成率。
          </Notice>
        </Panel>
      )}
      {tab === '业务 KPI' && (
        <Panel title="业务指标 · 不只看调用次数">
          <Rows
            headers={['指标', '当前值', '目标', '定义']}
            rows={[
              ['紧急工单派单时长', '2.4 分钟', '≤ 5 分钟', '创建到责任人确定'],
              ['一次解决率', '86.3%', '≥ 85%', '无需二次派单的已完成工单'],
              ['SLA 达成率', '97.1%', '≥ 95%', '承诺时限内完成'],
              ['客户满意度', '4.7 / 5', '≥ 4.5', '验收后回访评分'],
            ]}
          />
        </Panel>
      )}
    </>
  );
}
export function Governance() {
  const { state, update, reset } = useStore();
  const [tab, setTab] = useState('权限策略');
  const [region, setRegion] = useState('中国大陆 · 专属部署');
  const [days, setDays] = useState('90');
  const [resetOpen, setResetOpen] = useState(false);
  return (
    <>
      <Heading
        eyebrow="ENTERPRISE / SECURITY & GOVERNANCE"
        title="可信赖，来自每一层边界。"
        description="Tenant、业务对象、字段和 Capability 的权限共同生效，默认拒绝。"
      >
        <Badge tone="success">安全策略 · 演示配置</Badge>
      </Heading>
      <Tabs
        items={['权限策略', 'Secret Vault', '数据驻留与保留', '审计日志']}
        value={tab}
        onChange={setTab}
      />
      {tab === '权限策略' && (
        <Panel title="RBAC + ABAC 权限矩阵">
          <Rows
            headers={['角色', '业务对象', '字段权限', 'Capability', '审批权限']}
            rows={[
              [
                '报修人',
                '本人提交的工单',
                '手机号仅本人可见',
                'create / get-own',
                '无',
              ],
              [
                '工程师',
                '本班组已分派工单',
                '客户手机号脱敏',
                'update / complete',
                '无',
              ],
              [
                '物业主管',
                '当前园区工单',
                '按职责授权',
                'assign / approve',
                '本园区高风险操作',
              ],
              [
                'IT 管理员',
                '配置与连接器',
                '不能直接读取业务数据',
                'adapter / schema',
                '技术发布审批',
              ],
            ]}
          />
          <Notice>
            原型中的身份切换与审批按钮不代表真实权限执行；生产需在服务端实现鉴权与职责分离。
          </Notice>
        </Panel>
      )}
      {tab === 'Secret Vault' && (
        <Panel title="只暴露引用，不暴露 Secret">
          <Rows
            headers={['引用', '用途', '轮换', '状态']}
            rows={[
              [
                'vault://northstar/property-api',
                '物业系统适配器',
                '每 30 天',
                '仅演示引用',
              ],
              [
                'vault://northstar/channel/wecom',
                '企业微信身份入口',
                '每 90 天',
                '仅演示引用',
              ],
              [
                'vault://northstar/model-gateway',
                '模型调用',
                '每 30 天',
                '仅演示引用',
              ],
            ]}
          />
          <Notice>
            浏览器、日志、导出包和 Figma 文档均不得包含真实
            Secret。部署时使用目标环境的 Vault 注入。
          </Notice>
          <Button
            variant="outline"
            onClick={() =>
              update({}, '已生成 Secret 轮换演练记录（未操作真实密钥）')
            }
          >
            模拟轮换演练
          </Button>
        </Panel>
      )}
      {tab === '数据驻留与保留' && (
        <Panel title="Data Residency / Retention">
          <div className="two-col">
            <Field label="部署与存储区域">
              <select
                value={region}
                onChange={(e) => setRegion(e.target.value)}
              >
                <option>中国大陆 · 专属部署</option>
                <option>新加坡 · 专属区域</option>
                <option>欧洲 · 专属区域</option>
              </select>
            </Field>
            <Field label="Trace 保留天数（7–365）">
              <Input
                type="number"
                min={7}
                max={365}
                value={days}
                onChange={(e) => setDays(e.target.value)}
              />
            </Field>
          </div>
          <Rows
            headers={['类别', '保留策略', '例外']}
            rows={[
              ['业务交易数据', '按客户合同与法定要求', '不随会话 TTL 删除'],
              ['知识文档', '到期禁用检索；依策略清理', '冻结包保留来源版本'],
              ['临时运行数据', '1–24 小时', '待审批 checkpoint 延长'],
              ['审计', '追加写、归档与完整性验证', 'Legal Hold 阻止清除'],
            ]}
          />
          <Button
            disabled={Number(days) < 7 || Number(days) > 365}
            onClick={() =>
              update(
                {},
                '已记录治理配置演练：' +
                  region +
                  '，Trace 保留 ' +
                  days +
                  ' 天',
              )
            }
          >
            保存配置演练
          </Button>
          <Notice>
            区域选择只是产品配置示例，不会迁移当前网站或数据。真实部署需覆盖模型供应商、备份、日志与跨境链路。
          </Notice>
        </Panel>
      )}
      {tab === '审计日志' && (
        <Panel
          title="Audit · 本地操作记录"
          aside={
            <Button
              variant="outline"
              onClick={() =>
                download('audit-log.json', {
                  simulated: true,
                  events: state.audit,
                })
              }
            >
              <Download size={14} />
              导出
            </Button>
          }
        >
          <Rows
            headers={['时间', '操作', '来源']}
            rows={state.audit.map((e) => [
              e.at,
              e.action,
              '原型用户 / northstar',
            ])}
          />
        </Panel>
      )}
      <div className="demo-controls">
        <p>演示数据仅保存在当前浏览器，可恢复初始场景。</p>
        <Button variant="outline" onClick={() => setResetOpen(true)}>
          重置演示数据
        </Button>
      </div>
      <Modal
        title="重置本地演示数据"
        open={resetOpen}
        onClose={() => setResetOpen(false)}
      >
        <p>
          这会清除当前浏览器中的 Agent
          编辑、审批与发布演练记录。已下载的文件不受影响。
        </p>
        <Button
          variant="destructive"
          onClick={() => {
            reset();
            setResetOpen(false);
          }}
        >
          确认重置
        </Button>
      </Modal>
    </>
  );
}
