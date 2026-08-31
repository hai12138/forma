import { useState } from 'react';
import { Package, Download, ExternalLink } from 'lucide-react';
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
  Empty,
} from '../shared';
import { gateReasons, download, capabilities } from '@/lib/domain';
import { useStore } from '@/lib/store';
export function Delivery({ go }: { go: (r: string) => void }) {
  const { state, notify, update } = useStore();
  const [tab, setTab] = useState('交付包');
  const [customer, setCustomer] = useState('Northstar 园区');
  const [license, setLicense] = useState('企业年度授权');
  const ready =
    gateReasons(state).length === 0 && state.frozen === state.revision;
  const manifest = {
    format: 'forma.application/v1',
    simulated: true,
    name: state.application.name,
    version: '1.4.0',
    snapshot: state.revision,
    customer,
    license: { type: license, scope: 'northstar', expires: '2027-08-31' },
    agents: state.agents
      .filter((a) => state.application.selected.includes(a.id))
      .map((a) => ({
        id: a.id,
        version: a.version,
        role: a.role,
        capabilities: a.capabilities,
      })),
    capabilities,
    orchestration: state.application.mode,
    context: state.application.context,
    knowledge: state.application.knowledge,
    behaviorPolicy: 'enterprise-safe-v1',
    runtime: { adapter: state.runtime, certified: false },
    data: { mode: state.dataMode, mapping: state.mapping },
    channels: state.channels,
    evaluation: state.evaluation,
    secrets: ['vault://northstar/model-gateway'],
    signature: null,
  };
  return (
    <>
      <Heading
        eyebrow="COMMERCIAL / APPLICATION DELIVERY"
        title="交付的是解决方案，不只是一段对话。"
        description="Agent Application Package · 授权、部署、升级与回滚边界统一管理。"
      >
        <Button
          disabled={!ready}
          onClick={() => {
            download('application-package-manifest.json', manifest);
            update({}, '已导出冻结快照的应用 Manifest（模拟包，未签名）');
          }}
        >
          <Download size={14} />
          导出 Package Manifest
        </Button>
      </Heading>
      <Tabs
        items={['交付包', '客户授权', '部署与升级']}
        value={tab}
        onChange={setTab}
      />
      {!ready && (
        <Notice tone="warning">
          交付包导出需要业务确认、当前快照回归和版本冻结。
          <Button variant="link" onClick={() => go('releases')}>
            处理发布 Gate →
          </Button>
        </Notice>
      )}
      {tab === '交付包' && (
        <div className="split">
          <Panel
            title="Package Composition"
            aside={
              <Badge tone={ready ? 'success' : 'warning'}>
                {ready ? '可导出' : '待冻结'}
              </Badge>
            }
          >
            <div className="package-heading">
              <span className="icon-tile blue">
                <Package size={30} />
              </span>
              <div>
                <h2>{state.application.name}</h2>
                <p className="muted">
                  v1.4.0 · r{state.revision} ·{' '}
                  {state.application.selected.length} 个业务 Agent
                </p>
              </div>
            </div>
            <Rows
              headers={['包内容', '说明']}
              rows={[
                ['manifest.json', '版本、目标与依赖锁定'],
                ['business_agents / capabilities', '业务定义与标准契约'],
                ['schemas / mappings', '业务数据契约'],
                ['knowledge references', '不可变知识版本与 ACL'],
                ['behavior / permissions', '统一行为策略与权限'],
                ['runtime / channels', '环境无关的适配配置'],
                ['evaluation_suite / report', '评测依据与结果摘要'],
                ['migrations / rollback', 'Schema 迁移与兼容性说明'],
              ]}
            />
            <Notice>
              当前导出为可审阅 JSON
              Manifest，不是已签名的生产部署包；真实系统需生成校验摘要、SBOM、签名与部署材料。
            </Notice>
          </Panel>
          <Panel title="Manifest Preview">
            <pre className="code-block long-code">
              {JSON.stringify(manifest, null, 2)}
            </pre>
          </Panel>
        </div>
      )}
      {tab === '客户授权' && (
        <Panel title="License & Entitlements">
          <div className="two-col">
            <Field label="客户 / Tenant">
              <Input
                value={customer}
                onChange={(e) => setCustomer(e.target.value)}
              />
            </Field>
            <Field label="授权类型">
              <select
                value={license}
                onChange={(e) => setLicense(e.target.value)}
              >
                <option>企业年度授权</option>
                <option>私有部署永久授权</option>
                <option>试用授权</option>
              </select>
            </Field>
          </div>
          <Rows
            headers={['授权维度', '示例范围']}
            rows={[
              ['Tenant', 'northstar'],
              ['应用 / Agent 数', '3 个应用 / 20 个 Agent'],
              ['运行额度', '100,000 次 / 月'],
              ['渠道', 'Web、企业微信、API'],
              ['到期', '2027-08-31'],
              ['离线授权', '带签名租约与可控宽限期'],
            ]}
          />
          <Notice>原型只展示授权配置，不实施计费或许可证校验。</Notice>
        </Panel>
      )}
      {tab === '部署与升级' && (
        <>
          <div className="three-col">
            {[
              ['托管部署', '平台控制面与托管运行面；独立租户边界。'],
              ['客户专属部署', '专属数据与执行环境；明确模型出口。'],
              ['离线交付', '依赖镜像、离线许可证与离线评测工具。'],
            ].map((x) => (
              <div className="mode-card" key={x[0]}>
                <h2>{x[0]}</h2>
                <p>{x[1]}</p>
              </div>
            ))}
          </div>
          <Panel title="升级计划 v1.3 → v1.4">
            <ol className="steps">
              <li>校验许可证、签名与依赖兼容性</li>
              <li>执行备份、Schema 迁移预演及恢复演练</li>
              <li>在 Staging 运行客户回归测试与渠道验收</li>
              <li>冻结升级包，灰度放量并观察业务 KPI</li>
              <li>异常回滚应用；涉及数据迁移时执行专门恢复流程</li>
            </ol>
            <Button
              variant="outline"
              onClick={() =>
                notify(
                  '升级预检模拟完成：需要客户环境、签名包和真实回归才能部署。',
                )
              }
            >
              模拟升级预检
            </Button>
          </Panel>
        </>
      )}
    </>
  );
}
export function DesignSystem() {
  const [tab, setTab] = useState('Design Tokens');
  const [status, setStatus] = useState('正常');
  const tokens = [
    ['Background', '#f7f8fa'],
    ['Surface', '#ffffff'],
    ['Primary', '#2860dd'],
    ['Text', '#242d3d'],
    ['Border', '#e6e9ef'],
    ['Success', '#4c9773'],
    ['Warning', '#bc8a38'],
    ['Danger', '#c35454'],
  ];
  return (
    <>
      <Heading
        eyebrow="FOUNDATIONS / DESIGN SYSTEM"
        title="克制、清晰，始终如一。"
        description="Apple-like Enterprise + AI Native · 从 Figma Variables 到 React 组件的共同语言。"
      >
        <Badge>Forma UI · 1.0</Badge>
      </Heading>
      <Tabs
        items={['Design Tokens', '组件规范', '交互状态', 'Figma → Cursor']}
        value={tab}
        onChange={setTab}
      />
      {tab === 'Design Tokens' && (
        <>
          <Panel title="Semantic Color Tokens">
            <div className="token-grid">
              {tokens.map((t) => (
                <div className="token" key={t[0]}>
                  <span style={{ background: t[1] }} />
                  <b>{t[0]}</b>
                  <code>{t[1]}</code>
                </div>
              ))}
            </div>
          </Panel>
          <div className="two-col">
            <Panel title="Typography">
              <p className="type-display">让业务，成为能力。</p>
              <h2>Section Heading · 15 / 650</h2>
              <p className="paragraph">
                Body · 13 / 400 · 系统无衬线字体，中文优先清晰可读。
              </p>
              <code>work_order.assign · Monospace / 12</code>
            </Panel>
            <Panel title="Geometry & Motion">
              <Rows
                headers={['Token', '数值']}
                rows={[
                  ['Spacing', '4 / 8 / 12 / 16 / 24 / 32'],
                  ['Radius', '4 / 8 / 10 / 12'],
                  ['Sidebar', '224px，移动端 64px'],
                  ['Motion', '150ms ease-out'],
                  ['Focus', '3px #a9c4ff，offset 3px'],
                ]}
              />
            </Panel>
          </div>
        </>
      )}
      {tab === '组件规范' && (
        <>
          <Panel title="Button / Badge / Input">
            <div className="component-demo">
              <Button>主要操作</Button>
              <Button variant="outline">次要操作</Button>
              <Button variant="ghost">轻量操作</Button>
              <Button variant="destructive">危险操作</Button>
              <Button disabled>不可操作</Button>
            </div>
            <div className="component-demo">
              <Badge tone="success">已验证</Badge>
              <Badge tone="warning">待确认</Badge>
              <Badge tone="danger">已阻断</Badge>
              <Badge>草稿</Badge>
            </div>
            <Field label="表单标签">
              <Input placeholder="标签独立于 placeholder，错误要有明确文本" />
            </Field>
          </Panel>
          <Panel title="产品复合组件">
            <Rows
              headers={['组件', '用途', '关键状态']}
              rows={[
                [
                  'AssetCard / AgentCard',
                  '资产入口、版本与操作',
                  '默认 / 悬停 / 已删除',
                ],
                [
                  'EvidenceCard / ApprovalRow',
                  '来源、冲突与确认',
                  '候选 / 冲突 / 已确认',
                ],
                [
                  'OrchestrationCanvas',
                  '节点、编排与共享配置',
                  '选择 / 编辑 / 失效',
                ],
                [
                  'ReleaseGate / VersionFreeze',
                  '发布条件和冻结证据',
                  '阻断 / 通过 / 过期',
                ],
                [
                  'HumanTaskDetail',
                  '审批与异常接管',
                  '等待 / 批准 / 拒绝 / 升级',
                ],
              ]}
            />
          </Panel>
        </>
      )}
      {tab === '交互状态' && (
        <Panel title="状态展示">
          <Tabs
            items={['正常', '加载', '空数据', '错误', '无权限', '成功']}
            value={status}
            onChange={setStatus}
          />
          {status === '正常' && <Notice>内容已加载，可以查看或操作。</Notice>}
          {status === '加载' && (
            <div className="skeleton-demo" aria-busy="true">
              <span />
              <span />
              <span />
              <p>正在读取资产…</p>
            </div>
          )}
          {status === '空数据' && (
            <Empty
              title="还没有业务资产"
              description="从一次业务访谈开始，建立你的第一个模型。"
            />
          )}
          {status === '错误' && (
            <Notice tone="warning">
              加载失败：示例 Adapter 不可用。请检查连接后重试。
            </Notice>
          )}
          {status === '无权限' && (
            <Notice tone="warning">
              你没有 work_order:approve 权限。请联系工作空间管理员。
            </Notice>
          )}
          {status === '成功' && (
            <Notice>保存成功。已生成新的草稿快照，旧生产版本保持不变。</Notice>
          )}
        </Panel>
      )}
      {tab === 'Figma → Cursor' && (
        <>
          <Panel title="设计与开发衔接">
            <div className="process-flow">
              {[
                'Tokens → Variables',
                '组件 → Figma Library',
                'Dev Mode → MCP',
                'Cursor → React',
              ].map((x, i) => (
                <div className="process-item" key={x}>
                  <span>0{i + 1}</span>
                  <h3>{x}</h3>
                </div>
              ))}
            </div>
            <Rows
              headers={['步骤', '约束与交付物']}
              rows={[
                [
                  '建立 Figma 文件',
                  'Foundations / Components / Product Flows / Handoff',
                ],
                [
                  '映射组件',
                  'Variants 与 Props 一一对应，维护 component-map.md',
                ],
                ['连接 MCP', '按官方安装方式授权，读取指定节点上下文'],
                ['Cursor 实施', '引用实际组件，不把截图当唯一规格'],
                ['变更校验', 'Tokens、交互与代码同版本验收'],
              ]}
            />
            <Notice>
              本次未创建 Figma 文件，也未连接账号；代码工程、Tokens
              和映射规范均在交付包中。MCP 连接不自动保证代码与设计双向同步。
            </Notice>
          </Panel>
          <Panel title="官方资料">
            <div className="resource-links">
              <a
                href="https://developers.figma.com/docs/figma-mcp-server/"
                target="_blank"
                rel="noreferrer"
              >
                Figma MCP 官方文档 <ExternalLink size={14} />
              </a>
              <a
                href="https://developers.figma.com/docs/figma-mcp-server/local-server-installation/"
                target="_blank"
                rel="noreferrer"
              >
                Dev Mode / 桌面 MCP 设置 <ExternalLink size={14} />
              </a>
              <a
                href="https://developers.figma.com/docs/code-connect/"
                target="_blank"
                rel="noreferrer"
              >
                Code Connect 组件映射 <ExternalLink size={14} />
              </a>
            </div>
          </Panel>
        </>
      )}
    </>
  );
}
