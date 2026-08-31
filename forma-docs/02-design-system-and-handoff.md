# Forma UI · Design System 与 Figma / Cursor 交接

## 设计方向

Apple-like Enterprise 的重点是秩序和可读性：浅灰工作区、纯白内容面板、细边框、有限阴影、系统字体、稳定间距、较少装饰。AI Native 体现为有上下文的工作伙伴、可审核的生成物和明确的下一步，不是每页放一个聊天框。

四层资产使用蓝 / 紫 / 青绿 / 琥珀作为辅助识别色。状态颜色独立于资产颜色，必须配合文字，不依赖红绿识别。破坏性动作必须解释影响与恢复方式。

## Tokens

机器可读源：`design/tokens.json`。当前主题：Light。正式暗色主题未设计，不应只用反色冒充适配。

| 分类 | Token | 值 / 语义 |
|---|---|---|
| 颜色 | background | #F7F8FA，页面背景 |
| 颜色 | surface | #FFFFFF，内容面板 |
| 颜色 | primary | #2860DD，主要操作 |
| 颜色 | foreground | #242D3D，主内容 |
| 颜色 | border | #E6E9EF，结构边界 |
| 状态 | success / warning / danger | #4C9773 / #BC8A38 / #C35454 |
| 间距 | base / scale | 4px；4/8/12/16/24/32 |
| 圆角 | small / control / panel | 4 / 8 / 12px |
| 字体 | font-family | 系统字体栈，中文 Microsoft YaHei / 系统回退 |
| 字号 | display / section / body / metadata | 29 / 15 / 13 / 10–11px |
| 动效 | duration | 150ms；尊重 prefers-reduced-motion |

小字号只用于辅助信息；生产须在真实中文显示器、200% 缩放和 WCAG 对比度审计后调整。当前未进行浏览器可访问性认证，不宣称符合全部 WCAG 条款。

## 组件规范

基础交互组件复用 `components/ui` 的 Button、Input、Textarea、Checkbox、Dialog、Table，保留键盘和焦点行为。业务复合组件位于 `components/shared.tsx` 与各领域 screen 文件。

| Figma Component | 代码组件 / 结构 | Props / Variants | 核心约束 |
|---|---|---|---|
| Action/Button | `Button` | default / outline / ghost / destructive；disabled；size | 标签说明动作，图标按钮有 aria-label |
| Form/Input | `Input` / `Field` | value、invalid、readOnly | 必须有标签，不只依赖 placeholder |
| Feedback/StatusBadge | `Badge` | success / warning / danger / neutral | 状态必须有文字 |
| Surface/Panel | `Panel` | title、aside、children | Header 与 Body 的间距统一 |
| Navigation/ViewTabs | `Tabs` | items、value、onChange | 使用可聚焦按钮与 aria-pressed |
| Data/RecordTable | `Rows` + `Table` | headers、rows | 小屏横向滚动，表头不可缺失 |
| Overlay/EditorDialog | `Modal` + `Dialog` | title、description、open | 焦点约束、Escape 关闭、取消动作 |
| Feedback/Notice | `Notice` | info / warning | 给出阻断原因与下一步 |
| Feedback/EmptyState | `Empty` | title、description、action | 空数据不伪装为失败 |
| Asset/AgentCard | Agent screen | status、version、capabilities | 编辑、复制、导出、删除明确分开 |
| Business/EvidenceCard | Business screen | confidence、source、conflict | Confidence 与 Confirmation 分离 |
| Application/CanvasNode | Application screen | selected、mode、agent ref | 构建模型与视觉坐标分离 |
| Release/Gate | Release screen | blocked / passed / frozen | 必须绑定当前资产快照 |

画布不使用截图作为实现；资产表格不在每页独立重造基础输入控件。设计可以把已稳定的领域复合结构提升为正式组件，不能仅为了组件数量而抽象。

## 页面与状态规范

页面框架：Breadcrumb → Eyebrow / Title / Description / Actions → View Tabs → 主工作面 → Context / Detail → Footer。总览用资产链路与待办；编辑页用画布或表单；治理页用清晰的规则表。避免把所有页面都做成同一种 KPI 卡片墙。

列表状态：正常、搜索无结果、无资产、加载、无权限、失败。编辑状态：新建、已修改、保存中、校验错误、保存成功、历史只读、被引用不可删除。流程状态：待确认、待回归、通过、失效、已冻结、发布中、灰度、回滚。所有真实网络交互上线后应接入统一 error boundary 与可重试机制。

原型已有：搜索无结果、创建必填校验、导入错误、模拟运行加载、审批结果、Gate 阻断/通过、软删除/恢复、成功通知，以及 `/design` 中的状态展示。`/design` 的加载/错误/无权限是组件示例，不意味着全部页面已有真实服务器权限错误处理。

## Figma 文件建议

新建文件 **Forma / Enterprise Platform / v1**，分为：

1. `00 Foundations`：语义颜色、字号、间距、圆角、图标原则、Grid。
2. `01 Components`：基础组件、Variants、交互状态、无障碍注释。
3. `02 Product Flows`：发现→确认、Agent CRUD、应用→发布、人工恢复四条流程。
4. `03 Screens`：16 个主页面，Desktop 1440 / Tablet 1024 / Mobile 390。
5. `04 Handoff`：组件代码映射、API 状态、权限说明、验收与变更记录。

Tokens 导入为 Figma Variables；CSS 中的值与 JSON 保持一致。Figma 节点 URL、设计版本、对应代码提交与负责人填入 `design/component-map.md`。本次没有可用的用户 Figma 文件/节点 ID，因此映射表留待连接后补齐，不制造虚假链接。

## Dev Mode / MCP / Cursor

截至本次核对，Figma 官方 MCP 可向支持的客户端提供 Variables、组件与布局上下文，也支持特定写入/Code-to-Canvas 工作流；当前官方推荐远程 MCP。具体功能受客户端、授权与可用性限制，不能承诺任意 Cursor 会话自动支持所有写入功能。[Figma MCP](https://developers.figma.com/docs/figma-mcp-server/)。

连接建议：在 Cursor 中使用 Figma 官方安装流程完成 OAuth，授权指定文件；以明确节点 URL 获取设计上下文。若组织要求桌面模式，需在 Figma 桌面应用开启对应服务，再按官方配置连接。[远程安装](https://developers.figma.com/docs/figma-mcp-server/remote-server-installation/)、[桌面安装](https://developers.figma.com/docs/figma-mcp-server/local-server-installation/)。

Code Connect 将 Figma 组件关联到真实代码，可增强 MCP 的实现上下文。官方当前介绍了 UI 和 CLI 两种方式；采用前应核对计划、席位与权限要求，本交付不假设账号已具备。[Code Connect 官方文档](https://developers.figma.com/docs/code-connect/)。

衔接步骤：

1. 先在 Cursor 打开整个工程，运行本地原型，了解真实状态变化。
2. 在 Figma 建立 Variables 和基础组件，不从截图手工猜色值。
3. 可以按当前受支持的 Code-to-Canvas 流程导入参考布局；之后人工检查 Auto Layout、组件复用和语义命名。
4. 将关键组件映射到实际代码，并给 Cursor 指定节点 URL 与目标页面。
5. Cursor 先阅读 Blueprint、路由说明和组件映射，复用现有组件，再实现差异。
6. 验收可交互行为、响应式布局、权限边界与 Tokens，不能只对比一张静态截图。
7. 将 Figma 版本、工程提交和验收结果记录到同一变更条目。

MCP 提供上下文和可用工具，不保证代码与设计自动双向同步。代码生成或画布导入均需审阅。不要把生产 Secret、客户个人信息或未脱敏 Trace 带进 Figma。

## 可直接交给 Cursor 的实施任务模板

> 阅读 docs/01-product-blueprint.md、docs/02-design-system-and-handoff.md 和 docs/03-prototype-coverage.md。以现有四层资产与导航为基线，完成 [目标模块] 的真实服务接入。设计上下文为 [Figma 节点 URL，未提供则不要虚构]。复用 components/ui 和 shared 组件，保持 design/tokens.json 的语义。把 lib/store.tsx 的对应模拟行为替换为经过鉴权的 API；保留错误、权限、加载与审计状态。不得把前端 Gate 作为发布授权依据。交付前验证业务验收用例、类型检查与构建，并说明仍然模拟的部分。

以上模板是交接材料，不会自动创建 Cursor 任务或修改用户的编辑器设置。
