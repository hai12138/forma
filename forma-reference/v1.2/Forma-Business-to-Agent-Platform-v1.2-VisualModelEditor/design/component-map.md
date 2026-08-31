# Figma → Code 映射登记

当前没有关联 Figma 文件，node URL 留空。建立正式文件后补齐，并记录 design version / commit / reviewer。

| Design component | Code path | Node URL | 状态 |
|---|---|---|---|
| Action/Button | components/ui/button.tsx | 待连接 | 代码已提供 |
| Form/Input | components/ui/input.tsx | 待连接 | 代码已提供 |
| Overlay/EditorDialog | components/shared.tsx → Modal | 待连接 | 代码已提供 |
| Feedback/StatusBadge | components/shared.tsx → Badge | 待连接 | 代码已提供 |
| Surface/Panel | components/shared.tsx → Panel | 待连接 | 代码已提供 |
| Asset/AgentCard | components/screens/agents.tsx | 待连接 | 领域内复合结构 |
| Business/EvidenceCard | components/screens/business.tsx | 待连接 | 领域内复合结构 |
| Business/BusinessCanvas | components/business-canvas.tsx | 待连接 | 主图 / 精简视图共享 |
| Business/SemanticNode | components/business-canvas.tsx | 待连接 | 8 种语义类型，含原四层资产 |
| Business/GraphModel | lib/business-canvas.ts | 不适用 | Nodes / Edges + 基础自动布局 |
| Release/Gate | components/screens/ship.tsx | 待连接 | 领域内复合结构 |

不要将“待连接”替换为虚构文件 ID。Code Connect 应在实际组件节点和权限可用后配置。

| Business/VisualBusinessModelEditor | components/visual-business-model-editor.tsx | 待连接 | v1.2 手动模型编辑器 |
| Business/SemanticModel | lib/visual-model.ts | 不适用 | 语义、布局、历史与治理适配 |
