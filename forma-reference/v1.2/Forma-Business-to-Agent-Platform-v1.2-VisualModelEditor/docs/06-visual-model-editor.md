# Visual Business Model Editor · v1.2

基线为用户提供的 Forma-Business-to-Agent-Platform-v1.1-BusinessCanvas.zip。仅增量修改业务编辑器及必要状态联动，不重做其他已确认页面。本轮交付本地 ZIP，不部署原在线站点。

## 需求覆盖

| # | 能力 | 实现与边界 |
|---|---|---|
| 1 | 节点拖拽 | SVG Pointer Events；按照 zoom 换算坐标；指针释放时合并成一次布局历史；取消不提交 |
| 2 | 编辑节点属性 | 双击聚焦名称，侧栏编辑名称、描述、类型，明确点击保存才提交 |
| 3 | 新增节点 | 顶部新增角色、实体、流程、状态、规则、外部系统；UUID 标识 |
| 4 | 删除节点 | 同时删除所有关联关系与布局位置；可撤销 |
| 5 | 创建关系 | 右侧连接点拖至目标；或侧栏起点 + 点击/Enter 目标；禁止自连和悬空边 |
| 6 | 编辑关系 | 选择边，在属性面板修改语义/标签；不允许空标签 |
| 7 | 删除关系 | 侧栏删除或 Delete；可撤销 |
| 8 | Undo/Redo | 统一历史栈，上限 100 步；新操作清空 redo；键盘与按钮；语义撤销仍生成新 revision |
| 9 | 自动/手动布局 | 确定性拓扑布局；手动切换保持当前位置；拖动自动转手动。auto 表示最后一次布局方式，不是实时后台布局服务 |
| 10 | 保存布局 | 本机自动保存；按钮创建独立 saved_layout 恢复点；恢复时只保留当前仍存在节点，不复活已删除业务对象 |
| 11 | Fit/Zoom/Fullscreen | 根据画布大小适配；5%–300% 缩放；浏览器全屏，不支持则明确提示 |
| 12 | 语义/布局分离 | 独立 semantic_model / view_layout；语义提交创建快照、使当前确认/评测/冻结失效并标记 Impact pending；布局保持凭证 |
| 13 | 来源 | 节点与语义关系 source=AI_generated/manual_modified；只标记实际语义变动对象。撤销恢复被撤销前的来源快照，版本记录保留撤销操作 |
| 14 | AI 重新布局 | 布局专用 API；本地模拟明确提示，不能接收或覆盖业务语义 |
| 15 | 可复用与数据驱动 | VisualBusinessModelEditor(value,onChange,onImpact)，通过 canvasGraph 兼容旧 nodes/edges；原只读 Canvas 保持不变 |
| 16 | 文档 | blueprint、design handoff、coverage、verification 与此文档同步更新 |
| 17 | 验证与打包 | 严格类型检查、两类生产构建、原有与新增领域测试、实际浏览器交互；新的 v1.2 ZIP |

## 数据契约

VisualModel 包含 semantic_model、view_layout、revision、past、future、revisions、impact 与可选 saved_layout。旧 localStorage 保持原键 forma-prototype-v1；没有 visualModel 的旧记录按当时规则惰性生成模型，不破坏原记录。

semantic_model.nodes 不允许保存 position。节点包含 id、type、label、description、source。semantic_model.edges 使用 id、from、target、label、source、可选 dashed。原 BusinessGraph.edges.source 仍然是起点，只有适配器转换该字段，避免将来源字符串误当成节点 ID。

rules/states 是 rule/state 节点 ID 的规范化索引，其业务定义在节点 description 中。类型切换、删除会同步维护索引；本轮不支持独立规则 AST、可执行状态迁移约束或 BPMN 协议。

view_layout 包含 node_positions、zoom、viewport、mode、groups。groups 预留但没有分组编辑 UI。saved_layout 是布局恢复点，不属于语义模型。拖动过程中只更新组件临时状态，指针释放才入历史与本地持久化。

## 治理与 Impact Analysis

语义差异按完整结构比较，no-op 不创建历史或版本。节点/边增删改创建新的 Business Model revision，revisions 保存提交后的完整语义快照及动作。Impact 记录最新 revision 和变动对象 ID；版本页展示真实编辑历史，原生产 v1.3 对照仍为预置演示。

Store 的 applyVisualModel 统一清空当前模型 approvals、evaluation、frozen，并将候选发布状态回到 Dev。原型只有一个业务模型，相关资产按整套工单候选保守失效，Impact 页保留已有四层显式依赖示例，不宣称精准推导任意业务图到能力、Agent、应用的依赖。pending 表示需要人工审阅；尚未实现自动完成影响任务的服务。

平台 state.revision 是原有所有资产的快照号；visualModel.revision 是业务模型语义号。确认和 Agent/Application 编辑沿用原平台快照行为，不增加模型语义版本。纯布局修改不改变任何一个版本，不使确认、评测或冻结失效。语义 Undo/Redo 增加新版本，决不能恢复旧评测或冻结。

证据页与 Analyst 的候选规则修改同步进入同一模型提交路径；画布编辑规则后证据输入框同步。/analyst 展示现有编辑模型。/data、/capabilities、/applications 仍是各自契约/依赖/编排图，未被替换成业务编辑器。已发布生产历史为原型静态示例，本次不会连接或修改真实生产。

## 存储与生产边界

本轮是可运行的前端原型，不是后端版本管理服务。所有数据仅保存在当前浏览器与当前源的 localStorage。自动保存与显式布局保存共享原 Store 的错误提示；配额/权限失败时仅保留会话内数据，不能保证跨设备持久化。历史栈最多 100 步，语义快照仍随编辑增长；长期生产使用应迁移到服务端版本库。

没有真实 LLM 布局、多用户协作、并发冲突合并、服务端权限与审计、复杂连线避障、边折点编辑、图片导出或自动业务语义修复。自动布局处理环、断开图与空图，但不保证复杂图最佳视觉排布。只读 Canvas 原图片导出仍为占位；本轮主编辑器不提供图片导出入口。

生产接入须将语义修改、revision 创建、凭证失效和影响任务写入同一服务端事务；布局采用独立接口及权限，AI 布局服务仅返回坐标白名单；审批/评测/冻结绑定模型版本或语义哈希。恢复历史语义应创建新提交，禁止恢复历史凭证。
