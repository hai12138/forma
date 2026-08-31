> 历史文档：以下记录 v1.1 的只读 Canvas。v1.2 的 /business 主视图已升级为可手动编辑的 VisualBusinessModelEditor，下文“不支持拖动、增删边”等限制仅适用于保留的只读 BusinessCanvas 组件。当前能力与验证以 [06-visual-model-editor.md](06-visual-model-editor.md) 和 [04-verification.md](04-verification.md) 为准。

# Business Canvas 增量交接 · v1.1

基线：用户上传的 Forma-Business-to-Agent-Platform-v1.0.zip。只修改业务图展示，不重建信息架构，不更换状态存储和发布规则。本次交付为本地 ZIP，未更新原线上站点。

## 设计

白色大画布、20px 点阵背景、低对比细边框、系统字体、轻量阴影。节点用颜色、图标及文字共同表达语义：角色蓝、实体琥珀、流程粉、状态青、规则紫、外部系统灰；额外支持 Agent 绿和 Application 靛蓝，保持四层资产可辨。

连线有箭头和业务标签，候选规则约束使用虚线。顶部保留标题、计数、导出和全屏；左侧选择/平移/适配工具；底部图例和缩放。节点文字过长省略，选中后在画布下方显示完整内容。所有样式限定于 BusinessCanvas，不覆盖原 Shell 或其他卡片。

## 组件和模型

- `components/business-canvas.tsx`：共享交互与 SVG 渲染。
- `components/business-canvas.css`：画布局部样式。
- `lib/business-canvas.ts`：BusinessNodeType、BusinessNode、BusinessEdge、BusinessGraph，以及各视图的图数据适配器。
- `tests/business-canvas.test.mjs`：状态同步、数据归属、编排拓扑、布局边界测试。

节点字段：`id`（图内唯一）、`type`、`label`、可选 `description`、可选 `position:{x,y}`、可选 `route`。边字段：`id`（图内唯一）、`source`、`target`、`label`、可选 `dashed`。边必须引用存在的节点。

`BusinessCanvas` 接收 `graph`、`title`、`compact`、`onSelect(node)`、`onNavigate(route)` 和 `actions`。节点类型标记和业务数据不依赖页面组件。缺省坐标采用确定性的拓扑分层布局；循环图放置到兜底列，不会阻塞渲染。主业务图使用展示坐标，未来可替换为布局引擎；不是 AI 自动布局或 BPMN 编辑器。

## 覆盖

| 位置 | 增量变化 | 保留内容 |
|---|---|---|
| /business 业务画布 | 角色、对象、规则、流程、状态和外部系统的统一关系图 | 标题、版本标记、证据入口 |
| /business 变更影响 | 四层资产依赖图，选中后可打开资产 | 变更计划表、评测入口 |
| /analyst | 添加工单访谈候选模型精简图，规则跟随 state.rule | 模式库、聊天、事实提取；仍为本地模拟 |
| /data 业务数据 | 对象至 External / Managed 的带标签归属图；Hybrid 按对象分流 | 模式选择、Data Contract 编辑和其余页签 |
| /capabilities 依赖图 | 当前所选能力版本的共享 Canvas | Registry、契约、绑定、OpenAPI、SDK |
| /applications 协作画布 | 共享 Canvas；随 Agent 组合与协作模式生成关系 | 管理 Agent 弹窗、侧栏、共享资源、冲突处理、配置 |

证据、版本、字段映射和变更动作表仍为表格，它们是编辑/审阅数据，不用表格代替业务关系图。渠道身份处理、设计工具交接等非业务模型流程不在本次范围，维持原布局。

## 交互与边界

节点支持鼠标或 Tab + Enter/Space 选中；平移工具支持拖拽；缩放按钮支持放大/缩小；适配会重置平移和缩放。聚焦图面时方向键平移、0 适配、Escape 取消选中。全屏使用浏览器 Fullscreen API，不支持时显示提示。导出仅为明确标注的 PNG / SVG / JSON 入口占位，不伪造下载成功。

应用 Pipeline/Handoff 为顺序边；Router 为择一分支；Parallel 为并行分支；Supervisor 为协调分支；Human-in-the-loop 为审批后继续。这些是前端拓扑说明，未实现真实执行引擎。未选择 Agent 时保留“发布受阻”的图示，原发布 Gate 不变。

当前不支持拖动节点、增删边、连线避障、多人协作、图片导出、真实 LLM 模型生成。无坐标图已支持基础分层，不保证复杂图的最优排版。极小视口可以缩放/平移查看，尚未完成浏览器视觉和响应式验收。

## 手工验收建议（本轮未执行）

1. /business 选中规则，确认完整候选文本；缩放、平移后点击适配；全屏并退出；导出显示占位提示。
2. 在证据页修改规则并保存，检查候选文字更新且双人确认失效。
3. /data 切换三种模式，确认 Asset 在 Hybrid 下指向客户系统，其余对象指向 Managed。
4. /capabilities 切换 Registry 条目，检查依赖图包含当前能力及版本。
5. /applications 添加/移除 Agent，切换 Pipeline、Parallel、Router；点击结果节点仍进入冲突与降级。
6. 完整走原有审批 → Agent → 应用 → 评测 → 冻结 → 发布流程。

## 安装兼容

当前环境使用 pnpm 11.19.0。原锁文件中的 ast-types 0.16.3、webpack 5.110.2 被最低发布年龄策略拒绝；保持直接依赖不变，显式固定 ast-types 0.16.1、webpack 5.108.0，按安全策略重新解析锁文件。`pnpm-workspace.yaml` 仅允许 esbuild、sharp、workerd 的标准构建初始化，没有降低发布年龄策略。ZIP 预览不需要安装依赖。
