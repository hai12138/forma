# FORMA S0-A COZE BASELINE REVIEW

## Status

**PASS_WITH_GATES**

本轮已完成 Coze Studio 源码级基线分析，并建立 Forma ↔ Coze 映射。但当前 workspace **不存在 Git 元数据**（无 `.git`、无 branch/commit/remote），且多项 Forma 企业能力在 Coze 开源版中缺失或为 Partial。在 Git 基线恢复、上游策略确认、以及本报告 Blocking Questions 人工答复前，**不得进入 S0-B**。

---

## Executive Summary

Forma 的技术路线成立：**Coze Studio 提供成熟的 Agent Engineering Foundation；Forma 在其上构建 Business-to-Agent Product Layer**。当前 `coze-studio/` 源码证实：

- **Backend**：Go 1.24 + CloudWeGo Hertz，严格 DDD 分层（`api` → `application` → `crossdomain` → `domain` → `infra`），Agent 与 Workflow **运行时均基于 CloudWeGo Eino**（`github.com/cloudwego/eino v0.4.8`），**未发现 LangGraph**。
- **Frontend**：Rush + pnpm monorepo，React 18.2 + TypeScript 5.8 + rsbuild/rspack + react-router-dom + zustand + Semi Design（`@coze-arch/coze-design`）。导航以 **Space / Develop / Library / Explore** 为中心，**与 Forma v1.2 IA 完全不同**。
- **Coze 已成熟可复用**：Single Agent CRUD/IDE、Prompt、Model 管理、Knowledge、Workflow 编辑器与执行、Plugin/Tool、Database/Variables/Memory、Conversation/Playground、Publish、App、Connector、OpenAPI/PAT、基础 Permission。
- **Forma 必须新建**：Business Asset、Visual Business Model Editor 后端、Evidence/Confirmation、Business Data Contract、Capability Contract/Gateway、Human Task Center、Forma 级 Evaluation/Release Gate、Channel Gateway（9 类）、Governance/Delivery/Observability 企业层、Multi-Agent Application 编排语义（Router/Supervisor/Pipeline 等）。
- **Forma v1.2 前端**（React 19 + Vinext + shadcn + Tailwind 4）是**最终产品 Shell 与视觉基线**，不得被 Coze 原生 UI 替换，也不应 iframe 整个 v1.2。

推荐扩展边界：Forma 域通过 `backend/crossdomain/forma/` + `backend/application/forma/` 建立 **Anti-Corruption Layer**，经 `forma_asset_ref ↔ coze_resource_ref` 映射 Coze Agent/Workflow/Knowledge/Plugin，**禁止 Forma Business Domain 直接依赖 Coze Agent Repository**。

---

## Coze Version Baseline

| 项 | 值 |
|---|---|
| **Repository** | `coze-studio/`（位于 `forma-workspace/coze-studio/`） |
| **Commit** | **不可用** — workspace 与 `coze-studio/` 均无 `.git` 目录 |
| **Branch** | **不可用** |
| **Remote** | **不可用** — README 指向 upstream `https://github.com/coze-dev/coze-studio.git` |
| **License** | Apache License 2.0 — `coze-studio/LICENSE-APACHE` |

### Backend Stack

| 组件 | 版本/技术 | 证据路径 |
|---|---|---|
| Go | 1.24.0 | `backend/go.mod` |
| HTTP | CloudWeGo Hertz v0.10.2 | `backend/go.mod`, `backend/main.go` |
| Agent/Workflow Runtime | CloudWeGo Eino v0.4.8 + eino-ext | `backend/go.mod`, `backend/domain/agent/singleagent/internal/agentflow/` |
| ORM | GORM + MySQL driver | `backend/go.mod`, `backend/domain/*/internal/dal/` |
| Cache | Redis (go-redis) | `backend/infra/cache/`, `docker/docker-compose.yml` |
| Vector | Milvus client v2 | `backend/go.mod`, `docker/docker-compose.yml` |
| Search | Elasticsearch v7/v8 | `backend/infra/es/` |
| Object Storage | MinIO / TOS / S3 | `backend/infra/storage/` |
| Message Queue | NSQ / RocketMQ / Kafka (IBM sarama) | `backend/infra/eventbus/` |
| IDL | Apache Thrift | `idl/`, `backend/api/model/` |
| Migration | Atlas (HCL + SQL migrations) | `docker/atlas/`, `docker/docker-compose.yml` |
| 构建/启动 | `Makefile`, `scripts/setup/server.sh`, Docker | `Makefile`, `docker/docker-compose.yml` |

### Frontend Stack

| 组件 | 版本/技术 | 证据路径 |
|---|---|---|
| Monorepo | Rush 5.147.1（259 projects） | `rush.json` |
| Package Manager | pnpm 8.15.8 | `rush.json` |
| Node | >= 21（`.nvmrc`: lts/iron） | `rush.json`, `.nvmrc` |
| React | ~18.2.0 | `frontend/apps/coze-studio/package.json` |
| TypeScript | ~5.8.2 | `frontend/apps/coze-studio/package.json` |
| Router | react-router-dom ^6.11.1 | `frontend/apps/coze-studio/package.json` |
| State | zustand ^4.4.7 + immer ^10.0.3 | `frontend/apps/coze-studio/package.json`, store packages |
| Build | rsbuild ~1.1.0 + rspack | `frontend/apps/coze-studio/package.json` |
| UI | `@coze-arch/coze-design` 0.0.6-alpha + Semi UI ~2.72.3（`@coze-arch/bot-semi`） | `frontend/apps/coze-studio/package.json` |
| CSS | Tailwind ~3.3.3 + Less | `tailwind.config.ts`, `*.less` |
| Test | vitest ~3.0.5 | `frontend/apps/coze-studio/package.json` |

### Runtime

- **Agent Runtime**：Eino ReAct Agent Flow（`eino/flow/agent/react` + `eino/compose`）
- **Workflow Runtime**：Eino Compose 图执行（`domain/workflow/internal/compose/`）
- **Checkpoint/Resume**：Eino CheckPointStore + `infra/checkpoint/`（Redis/Mem）
- **DeepSeek**：仅作为 **Model Provider**（`eino-ext/components/model/deepseek`），**不是 Harness**
- **LangGraph**：源码中 **Not Found**

### Docker 部署

- 生产式：`make web` → `docker compose -f docker/docker-compose.yml up`
- 开发调试：`make debug` → `docker/docker-compose-debug.yml`
- 镜像：`cozedev/coze-studio-server:latest`, `cozedev/coze-studio-web:latest`
- 中间件：MySQL 8.4、Redis 8、ES 8.18、MinIO、Milvus 2.5、etcd、NSQ
- 入口：`http://localhost:8888`（nginx → coze-web → coze-server:8888）

---

## Repository Architecture

```
coze-studio/
├── backend/          # Go DDD 后端（api, application, crossdomain, domain, infra, pkg, bizpkg, types, conf）
├── frontend/         # Rush monorepo（apps/coze-studio + packages/*）
├── idl/              # Thrift IDL（api, app, workflow, plugin, conversation, permission...）
├── docker/           # Docker Compose, Atlas migrations, nginx, volumes
├── common/           # Rush autoinstallers, templates
├── scripts/          # build_fe.sh, setup/server.sh, db migrate
├── helm/             # K8s charts
├── docs/             # 项目文档
├── rush.json         # Monorepo 配置
├── Makefile          # debug/fe/web/sync_db
└── LICENSE-APACHE    # Apache 2.0
```

**IDL → API Model → Handler → Application → CrossDomain → Domain** 是标准请求链路。Thrift 定义在 `idl/`，生成代码在 `backend/api/model/`。

---

## Backend Architecture

### 1. 总体分层

| 层 | 路径 | 职责 |
|---|---|---|
| **API** | `backend/api/` | Hertz HTTP handler、router、middleware、Thrift 生成 model |
| **Application** | `backend/application/` | 用例编排、DTO 转换、事务边界、注入 Domain/CrossDomain |
| **CrossDomain** | `backend/crossdomain/` | 跨域接口 + 默认实现（ACL/防腐层入口） |
| **Domain** | `backend/domain/` | 实体、领域服务、Repository 接口、internal/dal 实现 |
| **Infra** | `backend/infra/` | DB、Cache、ES、Storage、EventBus、Checkpoint、Embedding 等 |
| **Pkg** | `backend/pkg/` | 通用工具（errorx, logs, sonic, safego...） |
| **Bizpkg** | `backend/bizpkg/` | 横切业务包（config/modelmgr, llm/modelbuilder） |
| **Types** | `backend/types/` | errno, consts, DDL gen 工具 |
| **Conf** | `backend/conf/` | model/plugin/prompt/workflow YAML 配置 |

启动入口：`backend/main.go` → `application.Init()` → `api/router/register.go`

### 2. Domain 组织

`backend/domain/` 下 17 个域：

| Domain | 路径 | 说明 |
|---|---|---|
| agent | `domain/agent/singleagent/` | Single Agent（Bot） |
| app | `domain/app/` | Application（Coze App） |
| connector | `domain/connector/` | 渠道连接器 |
| conversation | `domain/conversation/` | 会话、消息、Agent Run |
| knowledge | `domain/knowledge/` | 知识库 |
| memory | `domain/memory/` | Database + Variables |
| plugin | `domain/plugin/` | Plugin/Tool |
| workflow | `domain/workflow/` | Workflow CRUD + 执行 |
| user | `domain/user/` | User + Space |
| permission | `domain/permission/` | RBAC 资源权限 |
| prompt | `domain/prompt/` | Prompt 资源 |
| search | `domain/search/` | ES 资源/项目搜索 |
| openauth | `domain/openauth/` | PAT/OpenAPI Auth |
| upload | `domain/upload/` | 文件上传 |
| template | `domain/template/` | 模板 |
| shortcutcmd | `domain/shortcutcmd/` | 快捷指令 |
| datacopy | `domain/datacopy/` | 数据复制任务 |

### 3. Application 如何调用 Domain

模式：**Application Service 持有 `DomainSVC` 接口，通过 init 注入**。

示例 — Single Agent：

- Application：`backend/application/singleagent/single_agent.go` — `SingleAgentApplicationService.DomainSVC singleagent.SingleAgent`
- Domain Service：`backend/domain/agent/singleagent/service/single_agent_impl.go`
- 初始化：`backend/application/singleagent/init.go`

示例 — Workflow：

- Application：`backend/application/workflow/workflow.go`
- Domain：`backend/domain/workflow/service/service_impl.go`

Application 层 **不直接访问 dal**，通过 domain service 或 crossdomain 默认服务。

**启动编排**（`backend/application/application.go`）：`initBasicServices` → `initPrimaryServices` → `initComplexServices`，完成后注册全部 CrossDomain `SetDefaultSVC()`。

### 4. CrossDomain 跨域调用

`backend/crossdomain/` 每个子域包含：

- `contract.go` — 接口定义（如 `crossdomain/agent/contract.go` 的 `SingleAgent`）
- `impl/` — 默认实现，委托给 domain service
- `model/` — 跨域 DTO（禁止引用 domain entity）
- `*mock/` — mock 生成

**注册模式**（`backend/application/application.go`）：

```go
crossagent.SetDefaultSVC(singleagentImpl.InitService(...))
crossworkflow.SetDefaultSVC(workflowImpl.InitService(...))
// ... 共 15+ crossdomain 服务
```

Domain 内跨域调用通过 `crossdomain/xxx.DefaultSVC()`，避免 domain 间直接 import。

### 5. Repository Interface 位置

| Domain | Repository 接口路径 |
|---|---|
| Agent | `backend/domain/agent/singleagent/repository/repository.go` |
| App | `backend/domain/app/repository/app.go` |
| Workflow | `backend/domain/workflow/interface.go` (`Repository`, `Service`) |
| Plugin | `backend/domain/plugin/repository/plugin_repository.go`, `tool_repository.go` |
| Knowledge | `backend/domain/knowledge/repository/repository.go` |
| User/Space | `backend/domain/user/repository/repository.go` |
| Conversation | `backend/domain/conversation/*/repository/repository.go` |
| Memory | `backend/domain/memory/database/repository/`, `variables/repository/` |
| Prompt | `backend/domain/prompt/repository/repository.go` |

### 6. Repository Implementation 位置

实现位于各 domain 的 `internal/dal/`：

- 例：`backend/domain/agent/singleagent/internal/dal/` — GORM gen 生成 query/model
- 例：`backend/domain/workflow/internal/repo/repository.go`
- 例：`backend/domain/user/internal/dal/space.go`, `space_user.go`

ORM 代码生成工具：`backend/types/ddl/gen_orm_query.go`

### 7. 数据库模型定义

- **GORM gen 模型**：`backend/domain/*/internal/dal/model/*.gen.go`
- **Domain Entity**：`backend/domain/*/entity/`
- **CrossDomain Model**：`backend/crossdomain/*/model/`
- **初始 Schema**：`docker/volumes/mysql/schema.sql`
- **Atlas HCL 声明式 Schema**：`docker/atlas/opencoze_latest_schema.hcl`

### 8. Migration 管理

- **Atlas migrations**：`docker/atlas/migrations/`（20250703–20251028 共 14+ 文件）
- **Docker 启动时自动 apply**：`docker/docker-compose.yml` mysql entrypoint 调用 `atlas schema apply`
- **开发者流程**：`docker/atlas/README.md` — `atlas migrate diff` → review SQL → `atlas migrate hash`
- **Makefile**：`make sync_db` / `make dump_db`

### 9. 用户体系

- Domain：`backend/domain/user/`
- Entity：`backend/domain/user/entity/`
- Service：`backend/domain/user/service/user_impl.go`
- Application：`backend/application/user/user.go`
- CrossDomain：`backend/crossdomain/user/`
- API：`api/handler/coze/passport_service.go`；IDL — `idl/passport/`
- 鉴权：Session-key（`user.session_key`），中间件 `api/middleware/session.go`
- 前端：`frontend/packages/foundation/account-*`

### 10. Workspace / Space / Project / Tenant

| 概念 | Coze 实现 | 路径 |
|---|---|---|
| **Space** | 工作空间（Personal/Team） | `backend/domain/user/entity/space.go`, `internal/dal/model/space.gen.go` |
| **Project** | Project IDE（App 开发容器） | `frontend/packages/project-ide/`, search entity `project_doc.go` |
| **App** | Coze Application | `backend/domain/app/` |
| **Agent/Bot** | Single Agent | `backend/domain/agent/singleagent/` |
| **Tenant** | **Not Found** 作为一等概念 — 开源版以 Space 隔离，无 Forma 级 tenant_id 语义 | — |

前端 Space 路由：`/space/:space_id/develop|library|bot/:bot_id|project-ide/:project_id|knowledge|database|plugin`

### 11. 权限

- Domain：`backend/domain/permission/permission.go` — `CheckAuthz(ctx, *CheckAuthzData)`
- 资源类型：`backend/domain/permission/consts.go` — Workspace, App, Agent, Plugin, Workflow, Knowledge, Connector, Project, EvaluationDataset...
- Application：`backend/application/permission/`
- CrossDomain：`backend/crossdomain/permission/`
- Middleware：`backend/api/router/coze/middleware.go`

**Partial**：有资源级 RBAC 框架（`domain/permission/authz_checker.go`），但开源版实际为 **Creator-ownership 模型**（`operatorID == resourceInfo.CreatorID` 才 Allow），**无 Forma 级 ABAC、字段级权限、双人确认、Secret Vault**。

### 12. Agent / Bot Domain

- 路径：`backend/domain/agent/singleagent/`
- Entity：`entity/single_agent.go`, `entity/publish.go`
- Service：`service/single_agent_impl.go`
- **Runtime**：`internal/agentflow/agent_flow_builder.go` — Eino ReAct
- Application：`backend/application/singleagent/`（create, get, publish, duplicate）
- CrossDomain：`backend/crossdomain/agent/contract.go`
- 前端 Agent IDE：`frontend/packages/agent-ide/`

### 13. Application / App

- Domain：`backend/domain/app/`
- Service：`service/service_impl.go`, `service/publish_app.go`
- Repository：`repository/app.go`, `app_impl.go`
- CrossDomain：`backend/crossdomain/app/`
- Application：`backend/application/app/`
- 前端：`frontend/packages/project-ide/`, `@coze-studio/project-publish`

Coze App = 含 Workflow 的应用容器，**≠ Forma Application Asset（多 Agent 编排 + 交付包）**。

### 14. Workflow

- Domain：`backend/domain/workflow/`（2000+ 行 service_impl.go）
- 执行：`service/executable_impl.go` — Eino compose SyncExecute/StreamExecute/StreamResume
- 内部：`internal/compose/`, `internal/nodes/`, `internal/canvas/`
- CrossDomain：`backend/crossdomain/workflow/contract.go`
- Application：`backend/application/workflow/`
- 前端：`frontend/packages/workflow/`
- 路由：`/work_flow`（独立全屏）

### 15. Plugin / Tool

- Domain：`backend/domain/plugin/`
- Tool 调用：`service/tool/invocation_http.go`, `invocation_mcp.go`（**MCP stub**）, `invocation_saas.go`
- Repository：`repository/plugin_impl.go`, `tool_impl.go`
- CrossDomain：`backend/crossdomain/plugin/`
- Application：`backend/application/plugin/`
- 配置：`backend/conf/plugin/pluginproduct/*.yaml`
- 前端：`frontend/packages/agent-ide/bot-plugin/`, `frontend/apps/coze-studio/src/pages/plugin/`

### 16. Knowledge

- Domain：`backend/domain/knowledge/`
- CrossDomain：`backend/crossdomain/knowledge/`
- Application：`backend/application/knowledge/`
- Infra：Embedding（Eino-ext）、Milvus、ES
- 前端：`frontend/packages/data/knowledge/`

### 17. Model

- 配置管理：`backend/bizpkg/config/modelmgr/`
- LLM 构建：`backend/bizpkg/llm/modelbuilder/` — 基于 Eino-ext 适配 OpenAI/Claude/Gemini/Ark/DeepSeek/Ollama/Qwen
- 模板：`backend/conf/model/template/*.yaml`
- Application：`backend/application/modelmgr/`
- 前端：`frontend/packages/agent-ide/model-manager/`

### 18. Publish

- Agent Publish：`backend/application/singleagent/publish.go`, `domain/agent/singleagent/entity/publish.go`
- App Publish：`backend/domain/app/service/publish_app.go`
- Workflow Release：`crossdomain/workflow/contract.go` — `ReleaseApplicationWorkflows`
- Connector 绑定发布：`domain/app/internal/dal/app_connector_release_ref.go`
- 前端：`frontend/packages/agent-ide/agent-publish/`, `frontend/packages/studio/workspace/project-publish/`

### 19. Connector / Channel

- Domain：`backend/domain/connector/`
- Entity：`entity/connector.go`
- CrossDomain：`backend/crossdomain/connector/`
- Application：`backend/application/connector/`
- App-Connector 关联：`domain/app/entity/connector.go`

**Partial**：Coze Connector 是发布渠道抽象（API/Web SDK 等），**不是 Forma 9 类 Channel Gateway + 身份链**。

### 20. Runtime 代码

| Runtime 组件 | 路径 |
|---|---|
| Agent Flow Builder | `backend/domain/agent/singleagent/internal/agentflow/agent_flow_builder.go` |
| Agent Flow Runner | `backend/domain/agent/singleagent/internal/agentflow/agent_flow_runner.go` |
| Agent Stream Execute | `backend/domain/conversation/agentrun/internal/singleagent_run.go` |
| Workflow Compose | `backend/domain/workflow/internal/compose/` |
| Workflow Execute | `backend/domain/workflow/service/executable_impl.go` |
| Interrupt/Resume | `agentflow/callback_reply_chunk.go`, `crossdomain/workflow` `StreamResume` |
| Checkpoint Store | `backend/infra/checkpoint/redis.go`, `mem.go` |
| Code Node Runner | `backend/domain/workflow/internal/nodes/code/code.go` |
| Code Runner (Python/JS) | `backend/infra/coderunner/` — Deno sandbox 或 direct 模式，由 `application/workflow/init.go` 注入 |
| Conversation AgentRun | `backend/domain/conversation/agentrun/` |

---

## Frontend Architecture

### Monorepo 结构

Rush 共 **259 个项目**（`rush.json`）；workspace 由 Rush + PNPM subspaces 管理，**无** 根级 `pnpm-workspace.yaml` / `turbo.json`。

```
frontend/
├── apps/
│   └── coze-studio/          # 主应用 shell
├── packages/
│   ├── agent-ide/            # Agent 编辑器（50+ 子包）
│   ├── workflow/             # Workflow 编辑器与 test-run
│   ├── data/                 # knowledge, memory/database, variables
│   ├── foundation/           # account, space-ui, layout, global
│   ├── project-ide/          # App/Project IDE
│   ├── studio/               # workspace, project-publish, stores
│   ├── arch/                 # i18n, logger, coze-design, web-context
│   ├── common/               # chat-area, prompt-kit, auth, assets
│   ├── community/            # explore (plugin store, template)
│   ├── components/           # virtual-list 等
│   └── devops/               # testset-manage
└── config/                   # eslint, ts-config, postcss
```

### Coze Studio App

- 路径：`frontend/apps/coze-studio/`
- 入口：`src/index.tsx` → `src/app.tsx`
- 路由：`src/routes/index.tsx`
- 布局：`src/layout.tsx`
- 构建：`rsbuild.config.ts`

### 关键模块与包

| 能力 | 前端包 | 产品入口 |
|---|---|---|
| Agent IDE | `packages/agent-ide/entry`, `layout`, `prompt`, `tool` | `/space/:id/bot/:bot_id` |
| Workflow | `packages/workflow/*`, `agent-ide/workflow` | `/work_flow` |
| Knowledge | `packages/data/knowledge/*` | `/space/:id/knowledge/:dataset_id` |
| Plugin | `packages/agent-ide/bot-plugin/*`, `apps/.../pages/plugin` | `/space/:id/plugin/:plugin_id` |
| Model | `packages/agent-ide/model-manager` | Agent IDE 内 |
| Publish | `packages/agent-ide/agent-publish`, `studio/project-publish` | `/bot/:id/publish`, `project-ide/.../publish` |
| Resource/Library | `apps/coze-studio/src/pages/library.tsx` | `/space/:id/library` |
| Workspace | `packages/foundation/space-ui-*`, `studio/workspace/*` | `/space/:id/develop` |
| Navigation | `packages/foundation/layout`, `space-ui-base/components/workspace-sub-menu` | Space 左侧子菜单 |
| Playground/Debug | `packages/agent-ide/chat-debug-area`, `chat-area-plugin-debug-common` | Agent IDE 调试面板 |

### 导航 / 路由 / 状态

- **Router**：react-router-dom v6，`createBrowserRouter` — `frontend/apps/coze-studio/src/routes/index.tsx`
- **主导航**：Space（develop/library/bot/project-ide/knowledge/database/plugin）+ Explore（plugin/template）+ Workflow 独立路由
- **State**：zustand + immer（`@coze-arch/bot-studio-store`, `@coze-studio/bot-detail-store`, `@coze-foundation/space-store` 等）
- **Design System**：`@coze-arch/coze-design`（npm alpha）+ Semi UI ~2.72.3 + Tailwind + Less/Sass
- **i18n**：`@coze-arch/i18n`

**与 Forma v1.2 对比**：Coze 无 `/business`, `/capabilities`, `/analyst`, `/releases`, `/governance` 等路由；IA 完全不同。

---

## Existing Agent Engineering Capability Inventory

### COZE-EXISTING-CAPABILITY-INVENTORY

| 能力 | 状态 | 代码位置 | 产品入口 | 可复用程度 | Forma 是否还需实现 |
|---|---|---|---|---|---|
| **Agent CRUD** | Implemented | `backend/application/singleagent/`, `domain/agent/singleagent/` | `/space/:id/develop`, `/bot/:bot_id` | 高 | 需 WRAP — Forma Business Agent 映射 Coze Agent |
| **Agent Editor** | Implemented | `frontend/packages/agent-ide/` | Agent IDE | 高（引擎）/ 低（UI） | UI 用 v1.2；引擎 REUSE |
| **Prompt** | Implemented | `domain/prompt/`, `agent-ide/prompt` | Agent IDE Prompt 面板 | 高 | WRAP — 业务 Prompt ≠ 企业 Business Rule |
| **Model** | Implemented | `bizpkg/config/modelmgr/`, `bizpkg/llm/modelbuilder/` | Admin model-management, Agent IDE | 高 | REUSE |
| **Knowledge** | Implemented | `domain/knowledge/`, `packages/data/knowledge/` | Library / Knowledge 页 | 高 | REUSE + EXTEND（ACL/版本绑定 Forma Asset） |
| **Workflow** | Implemented | `domain/workflow/`, `packages/workflow/` | `/work_flow`, Agent 内绑定 | 高 | REUSE — Capability Workflow 实现 |
| **Plugin** | Implemented | `domain/plugin/`, `agent-ide/bot-plugin/` | Library Plugin 页 | 高 | REUSE — Capability REST 实现 |
| **Tool** | Implemented | `domain/plugin/service/tool/` | Plugin Tool 页 | 高 | REUSE |
| **Database** | Implemented | `domain/memory/database/`, `crossdomain/database/` | `/space/:id/database/:table_id` | 高 | REUSE — Managed Data Plane 部分 |
| **Variables** | Implemented | `domain/memory/variables/` | Agent/Workflow 变量 | 高 | REUSE |
| **Memory** | Implemented | `domain/memory/`, `agent-ide/memory-tool-pane` | Agent IDE | 高 | REUSE（会话/变量级） |
| **Conversation** | Implemented | `domain/conversation/`, `packages/common/chat-area/` | Playground, OpenAPI chat | 高 | REUSE |
| **Playground** | Implemented | `idl/playground/`, Agent IDE debug | Agent IDE 调试 | 高 | WRAP — Forma Agent Test 需业务断言 |
| **Debug** | Implemented | `agent-ide/chat-area-plugin-debug-common/` | Agent IDE | 高 | PARTIAL REUSE |
| **Publish** | Implemented | `singleagent/publish.go`, `app/publish_app.go` | Publish 页 | 高（Coze 语义） | EXTEND — Forma Release Gate 是上层 |
| **Application** | Implemented | `domain/app/`, `project-ide/` | Project IDE | 中 | WRAP — Coze App ≠ Forma Application Asset |
| **API / SDK** | Implemented | `application/openauth/`, `conversation/openapi_*`, Chat SDK 文档 | OpenAPI + PAT | 高 | REUSE + EXTEND Forma Gateway |
| **Connector** | Partial | `domain/connector/`, `crossdomain/connector/` | Publish 渠道选择 | 中 | EXTEND — Forma 9 类 Channel |
| **Multi-Agent** | Partial | API model 有 `MultiAgent` 枚举（`api/model/app/bot_common/`） | 商业版特性，开源 Partial | 低-中 | NEW — Forma Supervisor/Router 等 |
| **Human-in-the-loop** | Partial | Workflow/Agent Eino Interrupt+Resume（OAuth/Tool） | Workflow 中断恢复 | 低 | NEW — Forma Human Task Center |
| **Evaluation** | Partial | Permission 类型有 EvaluationDataset/Task；`devops/testset-manage`；evaluation panel store 仅 UI 开关 | Agent IDE evaluation panel | 低 | NEW — Forma 业务驱动 Evaluation |
| **Observability** | Partial | Workflow trace API（`api/model/workflow/trace.go`）；Slardar 前端 adapter | Workflow test-run trace | 低 | NEW/EXTEND — Forma 全链路 Trace/KPI |
| **Permission** | Partial | `domain/permission/` RBAC | API middleware | 中 | EXTEND — Forma RBAC+ABAC+Governance |
| **Channel** | Partial | Connector 抽象 | Publish | 低 | NEW — Forma Channel Gateway |

---

## Runtime Architecture

### COZE-RUNTIME-ARCHITECTURE

```
┌─────────────────────────────────────────────────────────────┐
│                    Product Layer (Coze UI/API)               │
│  Agent IDE │ Workflow Editor │ Playground │ OpenAPI Chat     │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│              Application + CrossDomain Layer                 │
│  conversation/agent_run │ singleagent │ workflow app svc    │
└──────────────────────────┬──────────────────────────────────┘
                           │
         ┌─────────────────┴─────────────────┐
         ▼                                   ▼
┌─────────────────────┐           ┌─────────────────────┐
│   Agent Runtime      │           │  Workflow Runtime    │
│  Eino ReAct Agent    │           │  Eino Compose Graph  │
│  agentflow/          │           │  workflow/compose/   │
└─────────┬───────────┘           └─────────┬───────────┘
          │                                  │
          └──────────┬───────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    Eino Infrastructure                       │
│  compose │ schema │ components/tool │ components/model      │
│  flow/agent/react │ callbacks │ CheckPointStore              │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  Infra: LLM Providers (OpenAI/Claude/Gemini/Ark/DeepSeek/   │
│  Ollama/Qwen) │ Redis Checkpoint │ Milvus │ MySQL │ Plugin  │
│  HTTP/MCP(stub) │ Knowledge Retriever │ Code Runner           │
└─────────────────────────────────────────────────────────────┘
```

**Agent Runtime 是什么？**

- 基于 Eino 的 **ReAct Agent Graph**：Persona → Knowledge Retriever → Prompt Template → ReAct Agent + Tools
- 构建：`backend/domain/agent/singleagent/internal/agentflow/agent_flow_builder.go`
- 执行：`agent_flow_runner.go` → 流式事件 → `conversation/agentrun`

**Workflow Runtime 是什么？**

- 基于 Eino Compose 的 **DAG/Canvas 图执行引擎**
- Canvas JSON → Schema → Compose Graph → Sync/Stream Execute
- 支持 Tool Interrupt + StreamResume（`executable_impl.go`, `crossdomain/workflow/contract.go`）

**Eino 用在哪里？**

| 用途 | 包 | 路径 |
|---|---|---|
| Agent ReAct 循环 | `eino/flow/agent/react` | `agentflow/agent_flow_builder.go` |
| Graph 编排 | `eino/compose` | `workflow/internal/compose/`, `agentflow/` |
| 消息 Schema | `eino/schema` | 全局 message/stream |
| LLM 组件 | `eino-ext/components/model/*` | `bizpkg/llm/modelbuilder/` |
| Embedding | `eino-ext/components/embedding/*` | knowledge pipeline |
| Tool 组件 | `eino/components/tool` | agent tools |
| Checkpoint | `compose.CheckPointStore` | `infra/checkpoint/` |

**Runtime 与产品层边界**

- **产品层**（Application/API/Frontend）：配置 Agent/Workflow、发起 Run、展示 SSE 流
- **Runtime 层**（Domain agentflow/workflow compose）：图构建、LLM 调用、Tool 执行、Interrupt
- **Infra 层**：模型连接、存储、消息队列
- Forma Kernel/Behavior Policy 应插入 **Application/CrossDomain 边界**，不在 Eino Graph 内部硬编码业务规则

**未发现**：LangGraph、DeepSeek Harness（仅有 DeepSeek 模型 API 适配）

---

## Forma v1.2 Product Baseline Summary

### 来源

- 原型：`forma-reference/v1.2/Forma-Business-to-Agent-Platform-v1.2-VisualModelEditor/`
- 文档：`forma-docs/01-product-blueprint.md` 等 + v1.2 增量 `docs/06-visual-model-editor.md`

### 四层资产

| 资产 | Forma 语义 | v1.2 原型 |
|---|---|---|
| Business Asset | Role/Object/Process/State/Rule/Evidence/Confirmation | `/business` — Visual Business Model Editor |
| Capability Asset | Contract/Implementation/Dependency | `/capabilities` |
| Agent Asset | 业务 Role + Capability refs | `/agents` |
| Application Asset | 多 Agent 编排 + 交付 | `/applications` |

### v1.2 关键增量

- **Visual Business Model Editor**：`components/visual-business-model-editor.tsx`
- **semantic_model / view_layout 分离**：`lib/visual-model.ts`
- 语义提交 → revision 递增 → 确认/评测失效 → Impact Analysis
- 布局操作不增加 revision

### Forma 导航（最终 IA — 不可被 Coze 替换）

| 路由 | 页面 |
|---|---|
| `/` | 总览 |
| `/analyst` | AI 业务分析师 |
| `/business` | 业务资产 / Visual Model Editor |
| `/data` | 数据平面 |
| `/capabilities` | 能力资产 |
| `/agents` | 业务 Agent |
| `/applications` | 应用构建器 |
| `/human` | 人工任务 |
| `/evaluation` | 测试与评测 |
| `/releases` | 版本与发布 |
| `/channels` | 渠道网关 |
| `/runtime` | Runtime 与 Kernel |
| `/observability` | 可观测性 |
| `/governance` | 安全与治理 |
| `/delivery` | 商业交付 |
| `/design` | 设计系统 |

### v1.2 前端技术栈（与 Coze 完全不同）

| 项 | v1.2 | Coze |
|---|---|---|
| React | 19.2.6 | 18.2.0 |
| Build | Vinext/Vite 8 | rsbuild/rspack |
| UI | shadcn/Base UI | Semi Design |
| CSS | Tailwind 4 + 语义 Tokens | Tailwind 3 + Less |
| State | React Context (`lib/store.tsx`) | zustand |
| 路由 | pathname + pushState | react-router-dom |

**结论**：v1.2 是 **独立产品 Shell**，迁移时保留 UI/UX/IA，底层接 Forma API + Coze 引擎。

---

## Forma ↔ Coze Capability Map

### FORMA-COZE-CAPABILITY-MAP

| Forma Feature | Coze Existing | Decision | Coze Code Location | Forma Integration Point | Data Ownership | Risk |
|---|---|---|---|---|---|---|
| **AI Business Analyst** | Not Found | NEW | — | `backend/application/forma/analyst/` | Forma | 高 — 核心差异化 |
| **Business Asset** | Not Found | NEW | — | `backend/domain/forma/business/` | Forma | 高 |
| **Visual Business Model Editor** | Not Found（Coze Workflow Canvas 不同语义） | NEW UI + NEW Backend | Coze workflow canvas 仅参考 | `frontend/packages/forma-business-editor/` | Forma | 中 — 勿混用 Workflow UI |
| **Evidence / Confirmation** | Not Found | NEW | — | `domain/forma/evidence/` | Forma | 高 |
| **Data Plane** | Partial（Database, Knowledge, External via Plugin） | EXTEND | `domain/memory/`, `domain/knowledge/` | `crossdomain/forma/dataplane/` | 混合 | 中 |
| **Business Data Contract** | Not Found | NEW | — | `domain/forma/contract/` | Forma | 高 |
| **Capability** | Partial（Plugin+Workflow+Database 可映射） | WRAP | `domain/plugin/`, `domain/workflow/` | `crossdomain/forma/capability/` | Forma 契约 + Coze 实现 | 中 |
| **Capability REST** | Implemented（Plugin HTTP） | REUSE | `plugin/service/tool/invocation_http.go` | Capability Gateway | Coze 执行 | 低 |
| **Capability MCP** | Partial（stub） | EXTEND | `plugin/service/tool/invocation_mcp.go` | Forma MCP Gateway | Coze | 高 — 当前 not implemented |
| **Capability Workflow** | Implemented | REUSE | `domain/workflow/` | 绑定 Capability impl ref | Coze | 低 |
| **Capability Knowledge** | Implemented | REUSE | `domain/knowledge/` | 绑定 + ACL | Coze | 低 |
| **Business Agent** | Implemented（Single Agent） | WRAP | `domain/agent/singleagent/` | `forma_agent_binding` → coze agent_id | 双层 | 中 |
| **Agent Editor** | Implemented（Coze IDE） | WRAP | `packages/agent-ide/` | v1.2 `/agents` UI + Coze API | Coze 执行配置 | 中 — UI 必须 v1.2 |
| **Agent Test** | Partial（Playground） | EXTEND | `agent-ide/chat-debug-area/` | Forma Evaluation 子集 | 混合 | 中 |
| **Agent Application** | Partial（Coze App） | WRAP | `domain/app/`, `project-ide/` | Forma Application Asset | Forma 编排 + Coze 资源 | 高 |
| **Multi-Agent** | Partial（API 枚举） | NEW | API models only | `domain/forma/application/orchestration/` | Forma | 高 |
| **Supervisor** | Not Found | NEW | — | Forma orchestration | Forma | 高 |
| **Router** | Not Found | NEW | — | Forma orchestration | Forma | 高 |
| **Human Task** | Partial（Interrupt only） | NEW | `agentflow` interrupt, `workflow` resume | `domain/forma/human_task/` | Forma | 高 |
| **Evaluation** | Partial | NEW | `permission/consts` types, `devops/testset-manage` | `domain/forma/evaluation/` | Forma | 高 |
| **Release** | Partial（Publish） | EXTEND | `app/publish_app.go`, `singleagent/publish.go` | `domain/forma/release/` Gate/Freeze/Canary | Forma 快照 | 高 |
| **Channel Gateway** | Partial（Connector） | EXTEND | `domain/connector/` | `domain/forma/channel/` | 混合 | 高 |
| **Runtime** | Implemented（Eino） | REUSE | `agentflow/`, `workflow/compose/` | Forma Runtime Adapter 层 | Coze 引擎 | 中 |
| **Observability** | Partial | EXTEND | `api/model/workflow/trace.go`, Slardar | `domain/forma/observability/` | 混合 | 中 |
| **Governance** | Partial（Permission） | NEW | `domain/permission/` | `domain/forma/governance/` | Forma | 高 |
| **Application Package** | Not Found | NEW | — | `domain/forma/delivery/` | Forma | 高 |

---

## Reuse List

可直接复用或薄包装（REUSE / WRAP）的 Coze 能力：

1. **Single Agent 引擎** — CRUD、Prompt、Model、Tool 绑定、流式执行
2. **Workflow 引擎** — 可视化编辑、执行、作为 Tool、Interrupt/Resume
3. **Plugin/Tool** — HTTP 插件、OpenAPI 导入、OAuth
4. **Knowledge** — 文档上传、切片、向量检索
5. **Database/Variables** — Agent/Workflow 结构化记忆
6. **Conversation/Playground** — 调试会话
7. **Model Manager** — 多供应商模型配置（含 DeepSeek API）
8. **Publish** — Agent/App 发布流程（作为 Forma Release 子步骤）
9. **Connector** — 基础渠道发布（需扩展为 Forma Channel Gateway）
10. **OpenAPI/PAT** — 外部 API 集成
11. **Permission RBAC 框架** — 资源级权限检查
12. **Eino Runtime** — Agent ReAct + Workflow Compose

---

## Forma Do Not Reimplement

### FORMA-DO-NOT-REIMPLEMENT

以下能力 Coze 源码已成熟，Forma **不应重写**：

| 能力 | 证据 | Forma 做法 |
|---|---|---|
| **Agent 执行循环（ReAct）** | `agentflow/agent_flow_builder.go` | 经 CrossDomain 调用 |
| **Workflow DAG 执行** | `workflow/service/executable_impl.go` | Capability Workflow 绑定 |
| **LLM 多模型适配** | `bizpkg/llm/modelbuilder/` | 直接使用 Model Manager |
| **Plugin HTTP 调用** | `plugin/service/tool/invocation_http.go` | Capability REST 实现 |
| **Knowledge RAG 管道** | `domain/knowledge/` + Eino retriever | Agent Knowledge 引用 |
| **Database CRUD** | `domain/memory/database/` | Managed Data 子集 |
| **Variables/Memory** | `domain/memory/variables/` | Agent Context |
| **Conversation SSE 流** | `conversation/agentrun/` | Agent Test / Runtime |
| **Workflow 可视化编辑器** | `frontend/packages/workflow/` | 仅 Capability 配置入口，非 Forma 主 UI |
| **Agent IDE 调试能力** | `agent-ide/chat-debug-area/` | 嵌入 v1.2 Agent Test（可选 iframe 单面板，非整页） |
| **OpenAPI Chat/Workflow API** | `application/conversation/openapi_*` | Channel API/Webhook |
| **Checkpoint/Interrupt 机制** | Eino compose + `infra/checkpoint/` | Human Task 恢复底层 |

---

## Required New Forma Domains

Forma 必须新建的 Domain（后端）：

| Domain | 职责 |
|---|---|
| `forma/business` | Business Model, semantic_model, view_layout, Pattern |
| `forma/evidence` | Evidence, Assertion, Confirmation, 双人审批 |
| `forma/contract` | Business Data Contract, Canonical Schema, Mapping |
| `forma/capability` | Capability Contract, Implementation binding, Dependency Graph |
| `forma/agent_binding` | Business Agent ↔ Coze Agent 映射 |
| `forma/application` | 多 Agent 编排（Router/Supervisor/Pipeline...） |
| `forma/human_task` | 审批队列、checkpoint 关联、SLA |
| `forma/evaluation` | 业务测试套件、Gate 绑定 |
| `forma/release` | Freeze, Canary, Rollback, 环境晋级 |
| `forma/channel` | 9 类渠道、身份链 |
| `forma/governance` | Secret, Audit, Retention, Policy |
| `forma/delivery` | Application Package, License, Upgrade |
| `forma/observability` | 业务 Trace, KPI, Cost |
| `forma/asset_registry` | 统一 asset_id, revision, dependency_refs |
| `forma/integration` | Anti-Corruption Layer → Coze CrossDomain |

---

## Forma Extension Boundary

### 1. 是否符合 Coze 代码习惯？

**是。** Coze 已采用：

- `backend/domain/<name>/` — entity + service + repository + internal/dal
- `backend/application/<name>/` — 应用服务 + init.go
- `backend/crossdomain/<name>/` — contract + impl + model
- `backend/infra/<capability>/` — 基础设施
- `idl/<name>/` — Thrift API 定义
- `frontend/packages/<team>/<name>/` — Rush 包

建议的 `forma_*` / `forma-*` 命名 **完全一致**。

### 2. 更低侵入的扩展方式？

| 方式 | 适用 | 侵入性 |
|---|---|---|
| **纯新增 Forma 包/域** | 业务建模、Evidence、Gate | 最低 |
| **CrossDomain Adapter** | 调用 Coze Agent/Workflow | 低 |
| **API Router 挂载** | `/api/forma/*` 新路由 | 低 |
| **Frontend 新 App Shell** | Forma v1.2 作为主应用 | 中（并存 Coze app） |
| **修改 application.go Init** | 注册 Forma 服务 | 中 |
| **修改 Coze domain 表** | ❌ 禁止 — 用 mapping 表 | 高 |

### 3. 必须修改 Coze Core 的位置

| 文件 | 原因 | 风险 |
|---|---|---|
| `backend/application/application.go` | 注册 Forma init | 中 |
| `backend/api/router/register.go` 或新增 `router/forma/` | Forma API 路由 | 中 |
| `idl/` 新增 `forma/` | Forma API 契约 | 低 |
| `backend/main.go` | 通常不需要改 | 低 |
| `frontend/apps/coze-studio/src/routes/index.tsx` | **不应改** — Forma 独立 Shell | — |

### 4. 可完全新增

- 全部 `backend/domain/forma_*`
- 全部 `backend/application/forma/`
- 全部 `backend/crossdomain/forma/`
- 全部 `backend/infra/.../forma/`（若需独立表）
- 全部 `frontend/packages/forma-*`
- 新 app：`frontend/apps/forma/`（推荐，而非嵌入 coze-studio）
- 全部 `idl/forma/`

### 5. Upstream Merge 冲突最高风险

| 路径 | 原因 |
|---|---|
| `backend/application/application.go` | 每次 upstream 新增服务都改 Init |
| `backend/api/router/coze/api.go` | Handler 注册集中 |
| `backend/api/router/coze/middleware.go` | 鉴权中间件 |
| `rush.json` | 新增 project 条目 |
| `docker/atlas/migrations/*` | 若 Forma 表混入 Coze migration |
| `backend/go.mod` | 依赖变更 |
| `frontend/apps/coze-studio/package.json` | 依赖变更 |

**低风险策略**：Forma DB migration 独立目录；Forma 路由独立文件；Init 用 `// FORMA-BEGIN/END` 标记块。

### 6. Anti-Corruption Layer 设计

```
Forma Domain (business, capability, application)
        │
        ▼
Forma CrossDomain Integration Interface
  ├── CozeAgentAdapter      → crossdomain/agent.DefaultSVC()
  ├── CozeWorkflowAdapter   → crossdomain/workflow.DefaultSVC()
  ├── CozeKnowledgeAdapter  → crossdomain/knowledge.DefaultSVC()
  ├── CozePluginAdapter     → crossdomain/plugin.DefaultSVC()
  └── CozeConnectorAdapter  → crossdomain/connector.DefaultSVC()
        │
        ▼
Coze Domain (禁止 Forma 直接 import)
```

**映射表**（Forma 自有 DB）：

```sql
-- 概念示意，S0-B 再正式建模
forma_asset_ref (
  tenant_id, asset_id, kind, revision, content_digest
)
coze_resource_ref (
  asset_id, coze_resource_type, coze_resource_id, coze_version, space_id
)
```

---

## Proposed Directory Layout

```
backend/
├── domain/
│   ├── forma/
│   │   ├── business/
│   │   ├── evidence/
│   │   ├── contract/
│   │   ├── capability/
│   │   ├── application/
│   │   ├── human_task/
│   │   ├── evaluation/
│   │   ├── release/
│   │   ├── channel/
│   │   ├── governance/
│   │   ├── delivery/
│   │   └── asset_registry/
│   └── ... (coze core 不动)
├── application/
│   └── forma/
│       ├── init.go
│       ├── business/
│       ├── analyst/
│       └── ...
├── crossdomain/
│   └── forma/
│       ├── integration/          # ACL adapters to coze
│       │   ├── agent_adapter.go
│       │   ├── workflow_adapter.go
│       │   └── ...
│       └── contract.go
├── infra/
│   └── orm/forma/                # Forma 独立表 migration
idl/
└── forma/
    ├── business.thrift
    ├── capability.thrift
    └── ...

frontend/
├── apps/
│   └── forma/                    # v1.2 迁移目标 app（主 Shell）
├── packages/
│   ├── forma-shell/              # 导航、布局、设计 tokens
│   ├── forma-business-editor/    # Visual Business Model Editor
│   ├── forma-capability/
│   ├── forma-agent/
│   ├── forma-application/
│   ├── forma-ship/               # evaluation, release, human
│   ├── forma-operate/              # runtime, observability, governance
│   ├── forma-delivery/
│   ├── forma-api-client/         # Forma + Coze API 客户端
│   └── forma-coze-embed/           # 可选：嵌入 Coze Workflow/Agent 面板
```

---

## Data Ownership Boundary

### Coze 拥有（执行引擎数据）

| 数据 | 存储 |
|---|---|
| Agent Draft/Version | `single_agent_draft`, `single_agent_version` |
| Workflow Canvas/Version | `workflow_*` 表 |
| Plugin/Tool | plugin 表 |
| Knowledge Doc/Slice | knowledge 表 + Milvus |
| Database/Variables | memory 表 |
| Model Instance | `model_instance` |
| App Draft/Release | `app_draft`, `app_release_record` |
| Conversation/Message | conversation 表 |
| Connector Release | `app_connector_release_ref` |
| User/Space | `user`, `space`, `space_user` |

### Forma 拥有（业务语义数据）

| 数据 | 说明 |
|---|---|
| Business Model | semantic_model + view_layout |
| Evidence / Assertion / Confirmation | 证据链 |
| Business Graph / Pattern | 行业模式 |
| Data Contract / Mapping | Canonical Schema |
| Capability Contract | 语义 ID、副作用、确认策略 |
| Capability Dependency | 依赖图 |
| Agent Binding | Forma Business Agent ↔ Coze Agent |
| Application Binding | 编排 + 共享策略 |
| Human Task | 审批队列 |
| Evaluation Suite / Report | 业务测试 |
| Release Snapshot | Gate 绑定快照 |
| Delivery Package | Manifest + 签名 |
| forma_asset_ref / coze_resource_ref | 映射层 |

**禁止**：在 `single_agent_draft` 等 Coze 表加 Forma 业务字段。所有关联走 **mapping 表 + asset_id**。

---

## Forma v1.2 UI Migration Map

### FORMA-V1.2-UI-MIGRATION-MAP

| v1.2 Route | v1.2 Component | Target Coze Area | Recommended Forma Package | Reusable Coze Component | Keep v1.2 UI? | Difficulty | Risk |
|---|---|---|---|---|---|---|---|
| `/` | `screens/overview.tsx` | 无对应 | `forma-shell` | 无 | **是** | M | 低 |
| `/analyst` | `screens/business.tsx` (Analyst) | 无 | `forma-business-editor` | 无 | **是** | H | 中 |
| `/business` | `visual-business-model-editor.tsx` | 无（≠ Workflow） | `forma-business-editor` | 无 | **是** | H | 中 |
| `/data` | `screens/data.tsx` | `/space/:id/database`, knowledge | `forma-dataplane` | Database UI, Knowledge UI（逻辑分离） | **是** | H | 中 |
| `/capabilities` | `screens/capabilities.tsx` | `/space/:id/plugin`, workflow | `forma-capability` | Plugin/Workflow 配置面板（可选 embed） | **是** | H | 中 |
| `/agents` | `screens/agents.tsx` | `/space/:id/bot/:bot_id` | `forma-agent` | Agent 引擎 API | **是** | M | 中 |
| `/applications` | `screens/applications.tsx` | `/project-ide/:id` | `forma-application` | 无（编排语义不同） | **是** | H | 高 |
| `/human` | `screens/ship.tsx` (HumanTasks) | 无 | `forma-ship` | 无 | **是** | H | 中 |
| `/evaluation` | `screens/ship.tsx` (Evaluation) | testset-manage（Partial） | `forma-ship` | 无 | **是** | H | 中 |
| `/releases` | `screens/ship.tsx` (Releases) | publish pages | `forma-ship` | Publish 流程 API | **是** | H | 高 |
| `/channels` | `screens/operate.tsx` (Channels) | Connector publish | `forma-operate` | Connector 选择器 | **是** | H | 高 |
| `/runtime` | `screens/operate.tsx` (Runtime) | Agent/Workflow runtime | `forma-operate` | 无 | **是** | M | 中 |
| `/observability` | `screens/operate.tsx` | Workflow trace | `forma-operate` | Trace 组件（Partial） | **是** | H | 中 |
| `/governance` | `screens/operate.tsx` | Permission admin | `forma-operate` | 无 | **是** | H | 高 |
| `/delivery` | `screens/delivery-design.tsx` | 无 | `forma-delivery` | 无 | **是** | H | 中 |
| `/design` | `screens/delivery-design.tsx` (DesignSystem) | 无 | `forma-shell` | 无 | **是** | L | 低 |
| `components/ui/*` | shadcn | Semi Design | `forma-ui` | 无 — 保持 shadcn | **是** | M | 低 |
| `lib/store.tsx` | 本地 mock | 全部 Coze/Forma API | `forma-api-client` | 无 | 替换 | H | 高 |
| `lib/domain.ts` | 领域类型 | 后端 types | `forma-types` | 无 | 迁移 | M | 低 |
| `lib/visual-model.ts` | 编辑器逻辑 | 新后端 | `forma-business-editor` | 无 | **是** | M | 低 |
| `components/platform.tsx` | App Shell | 无 | `forma-shell` | 无 | **是** | M | 低 |
| `lib/navigation.ts` | 导航定义 | Coze space menu | `forma-shell` | 无 | **是** | L | 低 |

**硬性约束**：

- ❌ 不得用 Coze Space 导航作为 Forma 最终导航
- ❌ 不得 iframe 整个 v1.2
- ❌ 不得用 Coze UI 覆盖 Forma Design System
- ✅ 可选：在 Forma 页面内 embed Coze Workflow Editor / Agent Debug 面板（局部）

---

## Coze Core Modification Risk Map

| 文件/区域 | 修改频率（upstream） | Forma 触碰概率 | 风险等级 | 缓解 |
|---|---|---|---|---|
| `backend/application/application.go` | 高 | 必须 | 🔴 高 | FORMA 标记块 + 薄 wrapper init |
| `backend/api/router/coze/*` | 高 | 可能 | 🔴 高 | 独立 `router/forma/` |
| `backend/domain/agent/*` | 中 | 不应 | 🟢 低 | 仅 CrossDomain 调用 |
| `backend/domain/workflow/*` | 中 | 不应 | 🟢 低 | ACL |
| `docker/atlas/migrations/*` | 高 | 若混用 | 🔴 高 | Forma 独立 migration 目录 |
| `rush.json` | 中 | 必须 | 🟡 中 | 追加 projects，不删改 |
| `frontend/apps/coze-studio/*` | 中 | 不应 | 🟢 低 | 独立 forma app |
| `backend/go.mod` | 中 | 可能 | 🟡 中 | 最小新增依赖 |
| `idl/**/*.thrift` | 中 | 新增 forma | 🟡 中 | 独立 idl/forma |

---

## Upstream Merge Strategy

### COZE-UPSTREAM-STRATEGY

**目标配置**（本轮不修改 remote）：

```
upstream = https://github.com/coze-dev/coze-studio.git
origin   = Forma 私有仓库（待创建）
```

**当前阻塞**：workspace 无 Git，无法执行任何 git 操作。S0-B 前必须：

1. 在 `coze-studio/` 或 workspace 根初始化 Git
2. 添加 `upstream` remote 指向 `coze-dev/coze-studio`
3. 添加 `origin` 指向 Forma 仓库
4. 记录 baseline commit hash

**Upstream Sync 流程（设计）**：

```
1. git fetch upstream
2. git checkout forma-main
3. git merge upstream/main   # 或 rebase（团队选定一种，建议 merge 保历史）
4. Forma compatibility test:
   a. backend: go test ./...
   b. frontend: rush build --to forma（Forma 包）
   c. frontend: rush build --to @coze-studio/app（Coze 核心）
   d. migration test: atlas migrate status + forma migrations
   e. UI regression: forma-shell 视觉快照
   f. Forma extension conflict review:
      - git diff --name-only upstream/main | grep -E 'application.go|router|atlas'
5. 人工 sign-off
```

**Forma 专属 CI Gate**：

- Forma 包 lint/test 独立于 Coze
- Mapping 表 migration 不回写 Coze schema
- ACL 接口兼容性契约测试

---

## License Notes

- **类型**：Apache License 2.0
- **文件**：`coze-studio/LICENSE-APACHE`
- **Forma fork/修改/商业部署需保留**：
  - 原始版权声明与 LICENSE 全文
  - 修改文件中的变更说明（NOTICE 文件，如适用）
  - 贡献者归属（`AUTHORS` 参考）
- **允许**：商业使用、修改、分发、专利授权（受 Apache 2.0 条款约束）
- **注意**：Forma 新增代码应使用独立版权声明；Coze 衍生部分继续 Apache 2.0
- **第三方**：Eino、Hertz、Semi Design 等各有许可证，需在 SBOM 中追踪（Forma Delivery 要求）

---

## Architecture Risks

| # | 风险 | 严重度 | 缓解 |
|---|---|---|---|
| R1 | 无 Git 基线，无法追踪 upstream | 🔴 | 立即初始化 Git + 记录 commit |
| R2 | Forma v1.2 与 Coze 前端栈完全不同（React 19 vs 18） | 🟡 | 独立 `apps/forma`，不混 monorepo 依赖 |
| R3 | MCP Tool 在 Coze 中为 stub | 🟡 | Forma Capability MCP 自行实现 Gateway |
| R4 | Coze 无 Tenant 概念 | 🟡 | Forma 层引入 tenant_id，Space 映射 |
| R5 | Multi-Agent 编排 Coze 开源 Partial | 🔴 | Forma Application Domain 完整新建 |
| R6 | Human Task 仅 Workflow Interrupt | 🔴 | Forma Human Task 独立队列 + 恢复协议 |
| R7 | Evaluation 无后端 | 🔴 | Forma Evaluation Domain 新建 |
| R8 | 直接改 Coze Agent 表导致 upstream 冲突 | 🔴 | 强制 mapping 表 |
| R9 | DeepSeek 仅 Model API，非 Harness | 🟡 | Forma Runtime Adapter 独立设计 |
| R10 | Coze 安全警告（公开部署风险） | 🟡 | Forma 生产需安全评估 + 网络隔离 |

---

## Blocking Questions

1. **Git 基线**：当前 `coze-studio/` 从何渠道获取？对应 upstream 哪个 commit/tag？需人工提供 hash。
2. **Forma 仓库结构**：`forma-workspace` 是 mono-repo（coze-studio + forma）还是 fork 单 repo？
3. **Tenant 模型**：Forma tenant 与 Coze Space 是 1:1 还是 1:N？
4. **首个 Runtime Adapter**：是否确认以 Coze Eino Runtime 为 V1 默认（非 LangGraph）？
5. **DeepSeek Harness**：是否有具体仓库/协议？若无，V1 是否仅 DeepSeek Model API？
6. **MCP 优先级**：Capability MCP 是否 V1 必须，还是 HTTP Workflow 先行？
7. **Coze UI 嵌入深度**：Agent Test 是否允许 embed Coze Playground 面板，还是完全自研？
8. **Database 策略**：Forma 表与 Coze 表同 MySQL instance 不同 schema，还是完全独立 DB？
9. **商业部署**：Forma 是否 AGPL/Apache 双许可，还是全 Apache？
10. **Reference Business**：维修工单是否确认为首个 E2E 验证场景？

---

## S0-B Preconditions

在进入 S0-B 前，必须满足：

- [ ] Git 初始化完成，upstream/origin remote 配置（可执行 fetch）
- [ ] Baseline commit hash 记录在本报告
- [ ] 本报告 Blocking Questions 至少 #1–#5 人工确认
- [ ] Forma 目录布局方案批准（本文 Proposed Directory Layout）
- [ ] 数据所有权边界批准（mapping 表方案）
- [ ] v1.2 UI Migration Map 批准（保持 v1.2 Shell）
- [ ] 明确 S0-B 范围边界（见下节）
- [ ] 开发环境可启动：`make web` 或 `make debug` 验证通过

---

## Recommended S0-B Scope

建议 S0-B（**仅基础设施，不含业务功能**）：

1. **Git & Upstream Setup** — init repo, remotes, baseline tag `forma-baseline-0`
2. **Forma 空壳目录创建** — `backend/domain/forma/asset_registry/`, `application/forma/init.go`, `crossdomain/forma/integration/`, `idl/forma/`
3. **Mapping 表 Migration** — `forma_asset_ref`, `coze_resource_ref` Atlas migration（独立目录）
4. **Forma API Router 挂载** — `/api/forma/health`, `/api/forma/version`（不含业务）
5. **Frontend App 脚手架** — `frontend/apps/forma/` 从 v1.2 复制 Shell（platform.tsx + navigation），不接后端
6. **ACL 接口定义** — `CozeAgentAdapter` interface + mock 实现
7. **CI 骨架** — backend test + forma frontend typecheck

**明确不在 S0-B**：

- Business Model 后端
- AI Analyst
- Capability Gateway 实现
- Human Task / Evaluation / Release
- Coze UI 迁移
- FastAPI 主后台

---

## DO NOT START S0-B

**等待人工审核。**

本报告基于 `coze-studio/` 当前源码与 `forma-reference/v1.2`、`forma-docs/` 产品基线生成。任何能力标记均来自代码路径实证，不含未验证假设。

### Subagent Validation Notes（2026-08-31）

独立探索代理 [Explore Coze backend architecture](e69d4e07-497f-448e-aa1a-c3bbb8e7c6c7) 与 [Explore Coze frontend architecture](01a315dc-d027-4523-904f-18504ca0e0ba) 的结论与主报告一致，并补充确认：

- Backend 权限为 **Creator-ownership**，非完整企业 RBAC/ABAC
- Workflow **Code Node** 经 `infra/coderunner` 执行 Python/JS（Deno sandbox）
- Eino 引用覆盖 **200+ 后端文件**
- Frontend **259** Rush 项目；状态管理为 **zustand + immer**；Semi UI **~2.72.3**

上述要点已合并入正文；**不改变** PASS_WITH_GATES 结论与 S0-B 前置条件。

---

*Generated: 2026-08-31 · Task: FORMA-S0-A · Analyst: Cursor Agent*
