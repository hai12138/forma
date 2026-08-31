# Forma v1.2 · Visual Model Editor

基于 v1.1 原工程增量升级，其他已确认页面保持不变。打开 /business 使用手动业务模型编辑器。包含更新后的源码、交接文档、测试与 preview-dist，无需安装依赖即可通过 start-preview.cmd 或 node preview.mjs 预览。

支持六类节点与关系编辑、拖拽、Undo/Redo、布局保存/恢复、Fit/Zoom/Fullscreen；语义修改生成新模型版本并使当前确认/评测/冻结失效，纯布局操作不触发失效。AI 重新布局为本地模拟，持久化为浏览器本机存储。

[v1.2 编辑器说明](docs/06-visual-model-editor.md) · [本轮验证](docs/04-verification.md)

以下保留原工程运行说明与历史更新。

# Forma · Business-to-Agent Platform

基于 v1.0 增量更新 Business Canvas 的企业工程平台高保真前端原型，包含 16 个主路由、业务资产治理、Agent CRUD、多 Agent 应用构建、模拟评测与发布闭环。

原 v1.0 私有在线预览（本次未更新）：[https://forma-business-agent-studio.cocoa-maple-0505.chatgpt.site](https://forma-business-agent-studio.cocoa-maple-0505.chatgpt.site)。仅站点所有者可访问；本地 ZIP 不依赖在线预览。

**所有数据、模型、审批、Runtime、渠道与发布均为本地模拟。** 没有连接生产系统、真实 Secret 或 Figma。设计与生产实现边界见下列文档。

## v1.1 更新

业务主图、变更影响、访谈摘要、数据归属、能力依赖与应用协作统一复用 BusinessCanvas。支持节点选中、缩放、平移、适配及全屏；导出为占位入口。其他页面、导航、四层资产与本地状态逻辑保留。

详细设计与手工验收建议见 `docs/05-business-canvas.md`。默认 4173 端口若已被旧原型占用，可在 PowerShell 中先设置 `$env:PORT=4183` 再运行 `node preview.mjs`。

## 文档

- `docs/01-product-blueprint.md`：完整产品方案、四层资产、IA、流程与架构。
- `docs/02-design-system-and-handoff.md`：Tokens、组件、Figma / Dev Mode / MCP / Cursor 交接。
- `docs/03-prototype-coverage.md`：逐项覆盖、三条演示路线、待实现边界。
- `docs/04-verification.md`：工程验证。
- `design/tokens.json`、`design/component-map.md`：设计数据与节点映射登记。

## 直接预览，无需安装项目依赖

下载 ZIP 并解压，电脑需要 Node.js 22.13 或以上版本。Windows 可运行 `start-preview.cmd`，或在工程目录运行：

```sh
node preview.mjs
```

打开 `http://127.0.0.1:4173`；Ctrl+C 停止。服务器只监听本机。ZIP 中的 `preview-dist` 已构建完整前端，不要直接双击 index.html，模块与路由需要 HTTP 服务。

## Cursor 开发

打开整个工程文件夹，保留 pnpm 锁文件。本轮使用 pnpm 11.19.0，保留 pnpm-workspace.yaml 中的兼容固定项：

```sh
pnpm install --frozen-lockfile
pnpm dev
pnpm typecheck
pnpm test
pnpm build
pnpm build:portable
```

`dev` 启动 Sites/Vinext；`build` 生成 Worker 兼容网站；`build:portable` 生成本地便携前端；`preview` 启动预构建版本。两种入口共用相同 React 页面。

依赖已安装但包管理器启动器损坏时，可直接运行 `node node_modules/vinext/dist/cli.js dev`。此命令不能替代安装依赖，也不需要修改全局安全配置。

## 工程结构

```text
app/                        路由与主题样式
components/platform.tsx     Shell、导航与搜索
components/screens/         按领域拆分的页面
components/shared.tsx       业务基础组件
components/ui/              shadcn / Base UI 基础库
lib/domain.ts               领域类型、Gate 与导入校验
lib/store.tsx                本地模拟状态
lib/navigation.ts           导航定义
design/                     Tokens 与 Figma 映射登记
docs/                       产品、覆盖与交接文档
tests/                      关键领域规则测试
portable/                   便携入口
preview-dist/               ZIP 内预构建版本
preview.mjs                 无额外依赖的本地服务
```

## 演示与边界

业务规则双人模拟确认 → 编辑 Agent → 组合应用 → 生成用例并回归 → 冻结 → Test → Staging → Prod → 导出 Manifest。配置变化会使评测与冻结失效。详细脚本见覆盖文档。

localStorage 键：`forma-prototype-v1`。不同域名、端口和浏览器不共享状态。安全与治理页可重置演示数据。请勿输入真实个人信息或 Secret。

前端 Gate 不是生产授权边界。后续须实现服务端资产、身份、事务与发布服务。基础依赖保留各自许可证；本项目不授予第三方服务、商标或订阅权限。
