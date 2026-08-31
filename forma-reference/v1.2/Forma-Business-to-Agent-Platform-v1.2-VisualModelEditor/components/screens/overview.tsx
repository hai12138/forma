import {
  ArrowUpRight,
  ArrowRight,
  Plus,
  Sparkles,
  Check,
  Layers,
  Network,
  Blocks,
  Bot,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
export default function Overview({ go }: { go: (route: string) => void }) {
  return (
    <>
      <div className="page-heading">
        <div>
          <p className="eyebrow">BUSINESS-TO-AGENT PLATFORM</p>
          <h1>让业务，成为能力。</h1>
          <p className="muted">
            从业务事实到可信赖的 Agent 应用。所有资产，在这里连接。
          </p>
        </div>
        <Button onClick={() => go('analyst')}>
          <Plus size={16} />
          开始业务发现
        </Button>
      </div>
      <div className="overview-layout">
        <div>
          <section className="asset-flow">
            <div className="section-line">
              <h2>业务到应用，一条完整链路</h2>
              <span className="muted">ASSET LIFECYCLE</span>
            </div>
            <div className="asset-grid">
              {[
                {
                  icon: Network,
                  name: '业务资产',
                  en: 'Business',
                  n: '12',
                  sub: '已确认的业务模型',
                  color: 'blue',
                },
                {
                  icon: Blocks,
                  name: '能力资产',
                  en: 'Capability',
                  n: '36',
                  sub: '可复用业务能力',
                  color: 'violet',
                },
                {
                  icon: Bot,
                  name: 'Agent 资产',
                  en: 'Agent',
                  n: '8',
                  sub: '专注业务的智能体',
                  color: 'teal',
                },
                {
                  icon: Layers,
                  name: '应用资产',
                  en: 'Application',
                  n: '3',
                  sub: '可交付的解决方案',
                  color: 'orange',
                },
              ].map((a, i) => (
                <button
                  key={a.en}
                  onClick={() =>
                    go(
                      ['business', 'capabilities', 'agents', 'applications'][i],
                    )
                  }
                  className="asset-card"
                >
                  <span className={'icon-tile ' + a.color}>
                    <a.icon size={22} />
                  </span>
                  <small>{a.en.toUpperCase()} ASSET</small>
                  <h3>{a.name}</h3>
                  <strong>
                    {a.n}
                    <ArrowUpRight size={18} />
                  </strong>
                  <p>{a.sub}</p>
                </button>
              ))}
            </div>
            <div className="flow-footer">
              <span>
                <Check size={14} />
                版本化资产
              </span>
              <span>
                <Check size={14} />
                全链路依赖
              </span>
              <span>
                <Check size={14} />
                发布前验证
              </span>
              <span className="push">业务事实 → 生产应用</span>
            </div>
          </section>
          <section className="panel">
            <div className="section-line">
              <h2>
                正在构建的应用 <span className="number">03</span>
              </h2>
              <button
                className="text-button"
                onClick={() => go('applications')}
              >
                全部应用 <ArrowRight size={15} />
              </button>
            </div>
            {[
              {
                name: '园区智能助手',
                description: '物业服务 · 工单 · 设备 · 客服',
                tag: '待发布',
                color: 'blue',
                n: 4,
              },
              {
                name: '企业服务助手',
                description: '企业服务 · 招商 · 访客预约',
                tag: '开发中',
                color: 'violet',
                n: 3,
              },
              {
                name: '设施运维 Copilot',
                description: '巡检 · 事件响应 · 资产生命周期',
                tag: '已发布',
                color: 'teal',
                n: 2,
              },
            ].map((a, i) => (
              <button
                className="application-row"
                key={a.name}
                onClick={() => go('applications')}
              >
                <span className={'app-icon ' + a.color}>
                  <Layers size={23} />
                </span>
                <span className="row-primary">
                  <b>{a.name}</b>
                  <small>{a.description}</small>
                </span>
                <span className="agent-dots">
                  <span>工</span>
                  <span>设</span>
                  <span>+{a.n - 2}</span>
                </span>
                <span
                  className={
                    'badge ' +
                    (i === 2 ? 'success' : i === 0 ? 'warning' : 'neutral')
                  }
                >
                  {a.tag}
                </span>
                <ArrowUpRight size={18} />
              </button>
            ))}
          </section>
          <section className="panel">
            <div className="section-line">
              <h2>工作空间动态</h2>
              <span className="muted">今天 · 8 月 31 日</span>
            </div>
            {[
              '工单 Agent v1.4 已通过回归评测',
              '紧急工单自动派单规则等待业务负责人确认',
              '园区智能助手 v1.3 已冻结，生产环境运行正常',
            ].map((a, i) => (
              <div className="timeline-row" key={a}>
                <span className={'timeline-dot t' + i} />
                <span>
                  {a}
                  <small>
                    {
                      [
                        '测试与评测 · 24 条业务用例',
                        '业务资产 · Evidence #EV-021',
                        '版本与发布 · Production',
                      ][i]
                    }
                  </small>
                </span>
                <time>{['10:42', '10:28', '09:15'][i]}</time>
              </div>
            ))}
          </section>
        </div>
        <aside className="context-column">
          <section className="insight-card">
            <span className="ai-label">
              <Sparkles size={17} />
              AI 工作伙伴
            </span>
            <h2>
              下一步，让业务
              <br />
              共识变得清晰。
            </h2>
            <p>
              紧急工单的派单规则存在分歧。确认业务事实后，我会为你追踪影响到的能力、Agent
              和测试。
            </p>
            <button onClick={() => go('business')}>
              审阅业务证据 <ArrowRight size={16} />
            </button>
            <div className="insight-foot">
              <span className="pulse" /> 1 项冲突 · 影响 4 个资产
            </div>
          </section>
          <section className="panel compact">
            <div className="section-line">
              <h2>需要你的关注</h2>
              <span className="badge warning">3</span>
            </div>
            {[
              '确认紧急工单处理规则',
              '审批高风险能力调用',
              '检查发布 Gate 阻断项',
            ].map((s, i) => (
              <button
                className="attention"
                key={s}
                onClick={() => go(['business', 'human', 'releases'][i])}
              >
                <span className={'attention-index idx' + i}>{i + 1}</span>
                <span>
                  {s}
                  <small>
                    {
                      [
                        '业务建模 · 待多人确认',
                        '人工任务 · 18 分钟后超时',
                        '园区智能助手 · Staging',
                      ][i]
                    }
                  </small>
                </span>
                <ArrowUpRight size={14} />
              </button>
            ))}
          </section>
          <section className="panel compact health">
            <div className="section-line">
              <h2>运行状态</h2>
              <span className="badge success">正常</span>
            </div>
            <div>
              <span>业务任务成功率</span>
              <strong>
                98.6<small>%</small>
              </strong>
            </div>
            <div className="mini-bars">
              {[
                35, 48, 41, 62, 50, 65, 70, 58, 74, 66, 83, 78, 88, 72, 91, 83,
                90, 98,
              ].map((h, i) => (
                <span key={i} style={{ height: h + '%' }} />
              ))}
            </div>
            <p className="muted">最近 24 小时 · 示例数据</p>
          </section>
        </aside>
      </div>
    </>
  );
}
