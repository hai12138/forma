# Forma V1 Cursor 全量开发实施方案

> 基线：`Forma-Business-to-Agent-Platform-v1.2-VisualModelEditor`  
> 文档定位：从高保真前端原型进入真实生产系统开发的唯一实施基线  
> 推荐实施方式：Cursor 分阶段执行，S0 → S13，逐阶段验收，不允许一次性“把整个平台做完”  
> Reference Business：维修工单（仅用于验证平台通用能力，不把 Forma 产品绑定为物业系统）

---

## 0. 文档结论与实施总原则

Forma V1 的产品目标不是“再做一个 Agent Builder”，而是：

> **将企业业务知识，工程化为可验证、可治理、可交付的 Agent 应用。**

v1.2 原型已经确定以下产品基线，生产开发不得随意推翻：

1. 四层核心资产：
   - Business Asset
   - Capability Asset
   - Agent Asset
   - Application Asset
2. 数据平面：
   - External
   - Managed
   - Hybrid
3. Business Model：
   - Evidence / Assertion / Confirmation
   - Version / Diff / Impact
   - `semantic_model` 与 `view_layout` 分离
   - Visual Business Model Editor 可手动编辑
4. Capability：
   - Contract
   - Implementation
   - Mapping
   - Dependency
5. Business Agent：
   - 只关注业务 Role / Context / Capability / Rule / Permission / Interaction
   - 不重复实现通用 Conversation / Memory / Streaming / Agent Loop
6. Agent Application：
   - 可组合多个 Business Agent
   - 支持 Router / Supervisor / Pipeline / Parallel / Handoff / Human-in-the-loop
7. Evaluation：
   - 从 Business Model / Capability / Agent / Application 派生测试
   - Gate 必须绑定当前资产快照
8. 发布：
   - Dev → Test → Staging → Prod
   - Freeze / Canary / Rollback
9. Channel Gateway：
   - Web/H5
   - 微信服务号
   - 企业微信
   - 飞书
   - 钉钉
   - Slack
   - Microsoft Teams
   - App SDK
   - API/Webhook
10. Runtime：
   - 必须 Adapter 化
   - 业务资产不得绑定单一 Runtime
11. Human Task：
   - 是正式 Capability Implementation，不是临时补丁
12. Observability / Governance / Delivery：
   - Trace / Audit / Secret / Policy / Package / License / Upgrade

### 0.1 生产开发必须遵守的 10 条原则

**P01. 原型不是后端真相。**  
当前 `lib/store.tsx`、localStorage、模拟审批、模拟 Evaluation、模拟 Runtime、模拟渠道全部必须逐阶段替换为真实服务端事实。

**P02. 业务模型是 Source of Truth。**  
Prompt 不是企业业务的最终真相；Business Model + Evidence + Confirmation 才是。

**P03. Agent 不直接访问客户系统。**  
Agent → Capability Gateway → Adapter / Managed Runtime / Human Task。

**P04. Agent 不直接信任模型输入中的 tenant / actor / permission。**  
身份由可信网关注入。

**P05. Layout 与 Semantic 必须永久分离。**  
移动画布节点不能导致业务版本变化；业务语义修改必须创建新 revision。

**P06. 发布资产必须不可变。**  
Production 不解析 `latest`；冻结后引用明确版本与 digest。

**P07. 高风险写操作必须具备幂等、确认、审计和未知结果恢复。**

**P08. 先模块化单体，再按边界拆服务。**  
不得按左侧菜单一页一个微服务。

**P09. 每阶段都必须有自动化 Gate。**  
Cursor 返回“代码已写完”不算完成。

**P10. 所有新增功能必须保留 v1.2 的 Apple-like Enterprise UI 语言。**  
不得为了快速实现回退成传统后台模板。

---

# 1. v1.2 原型现状与生产差距

## 1.1 原型已经真实存在的前端能力

v1.2 已有：

- React + TypeScript 高保真前端
- 16 个一级路由
- Business Canvas
- Visual Business Model Editor
- 节点拖拽、增删改
- 关系增删改
- Undo / Redo
- 自动 / 手动布局
- 保存 / 恢复布局
- Fit / Zoom / Fullscreen
- 语义 revision 与布局分离
- Agent CRUD / 复制 / 软删除 / 恢复
- 多 Agent Application 组合
- 6 种 Application 协作模式的产品表达
- Human Task 产品交互
- Evaluation / Gate 产品交互
- Release / Freeze / Canary / Rollback 产品交互
- 9 类 Channel 产品表达
- Runtime Adapter 产品表达
- Observability / Governance / Delivery 产品表达
- Design Tokens
- Figma / Cursor handoff 规范
- TypeScript / 构建 / 领域规则测试

## 1.2 当前仍然是模拟的部分

以下全部视为“未实现生产能力”：

- 真正多租户后端
- 用户体系 / 企业身份
- 真实 Business Asset Registry
- 真实 Evidence Storage
- LLM Business Analyst
- 多人真实确认
- 服务端 Visual Model revision
- 真实 Diff / Impact Engine
- External Data 连接
- Managed Business Runtime
- Business Data Contract 执行
- Capability Gateway
- REST / MCP / DB / Workflow Adapter
- Adapter SDK
- Agent Runtime
- 多 Agent Runtime
- Durable Human Task
- Evaluation Engine
- Release Engine
- Channel OAuth / 回调 / 消息
- Secret Vault
- Trace / Metrics / Cost
- 不可篡改 Audit
- Application Package 签名
- License / Upgrade
- 真实 Figma 连接

---

# 2. 推荐生产技术架构

> 本节是实施建议，不是 v1.2 原型已经存在的能力。

## 2.1 技术栈建议

### Frontend

继续沿用现有原型设计资产：

- React 19
- TypeScript
- 现有 `components/ui`
- 现有 Design Tokens
- Lucide
- 当前路由和 Screen 结构

生产开发中逐步增加：

- TanStack Query：服务端状态
- Zod：前端契约校验
- Playwright：浏览器 E2E
- React Hook Form（需要复杂表单时）
- 前端不持久化企业资产真相

> 不要求立即更换现有 Vinext/Vite 结构。S0 先验证其生产可维护性；如果决定切换框架，应在 S0 一次完成，S1 后禁止大规模重构前端基础框架。

### Backend

推荐首版：

- Python 3.12+
- FastAPI
- Pydantic v2
- SQLAlchemy 2.x
- Alembic
- asyncpg
- httpx

理由：

- Business Modeling / LLM / Agent Runtime / Adapter 开发生态更适合 Python
- 和 JSON Schema / OpenAPI / Pydantic 契约结合自然
- 方便后续 LangGraph 等 Runtime Adapter

### Persistence

- PostgreSQL 16：平台 Control Plane 主库
- PostgreSQL Schema / Dedicated DB：Managed Business Runtime
- Redis：短缓存、分布式锁、rate limit、短期 runtime state
- S3 / MinIO：Knowledge 原文、附件、Package、Evaluation Artifact
- pgvector：第一阶段 Knowledge Embedding 可选实现
- 不在 V1 引入独立图数据库作为硬依赖

Business Graph 第一版使用：

- PostgreSQL 表存语义对象与关系
- JSONB 保存不可变 Snapshot
- 应用层构建关系图和 Impact Graph

只有出现明确图查询规模瓶颈后再评估 Neo4j / AGE。

### Async / Worker

V1 推荐：

- PostgreSQL Outbox
- 独立 Worker Process
- Redis 用于轻任务调度 / 锁
- Durable Human Task 状态由数据库持久化

不要一开始引入过重工作流基础设施。

### Runtime

首个生产 Runtime 推荐：

- **LangGraph Adapter**

原因：
- 支持可恢复的状态执行思路
- Human-in-the-loop / checkpoint 比较契合 Forma
- Python 后端接入成本低

但是必须遵守：

> LangGraph 只是第一个 Runtime Adapter，不得成为 Forma Business Asset / Capability Asset 的数据模型。

后续 Runtime：

- DeepSeek Harness Adapter
- Eino Adapter
- Coze / 第三方 Runtime Adapter
- Custom Runtime Adapter

### Model Provider

统一 `ModelProvider`：

- OpenAI-compatible Chat API
- Embedding Provider
- Tool-call capability matrix
- Structured Output capability matrix

禁止业务代码直接依赖某一家模型 SDK。

### Observability

- OpenTelemetry
- structured JSON logging
- Prometheus metrics
- Grafana
- Trace backend 可从 Tempo / OTLP compatible 方案中选一个

### Local Development

Docker Compose 至少包含：

- postgres
- redis
- minio
- api
- worker
- web
- runtime-worker
- optional observability

---

# 3. 推荐 Repo 结构

```text
forma/
├─ apps/
│  └─ web/                     # 当前 v1.2 React 原型演进而来
│
├─ services/
│  ├─ api/                     # FastAPI 模块化单体
│  ├─ worker/                  # async jobs / outbox / evaluation
│  └─ runtime-worker/          # Runtime Adapter execution
│
├─ packages/
│  ├─ contracts/               # JSON Schema / OpenAPI / shared generated types
│  ├─ ui/                      # 稳定后从现有组件抽出
│  ├─ adapter-sdk-python/
│  ├─ runtime-sdk-python/
│  └─ channel-sdk/
│
├─ infra/
│  ├─ docker-compose.yml
│  ├─ docker/
│  ├─ migrations/
│  ├─ nginx/
│  └─ observability/
│
├─ tests/
│  ├─ e2e/
│  ├─ contract/
│  ├─ integration/
│  └─ security/
│
├─ docs/
│  ├─ architecture/
│  ├─ adr/
│  ├─ api/
│  └─ delivery/
│
├─ scripts/
├─ Makefile / Taskfile
└─ .env.example
```

## 3.1 FastAPI 内部模块边界

第一阶段保持一个 API 服务，内部按领域模块拆：

```text
app/modules/
├─ identity/
├─ tenancy/
├─ asset_registry/
├─ business_modeling/
├─ data_plane/
├─ capability/
├─ agent/
├─ application/
├─ human_task/
├─ evaluation/
├─ release/
├─ runtime/
├─ channel/
├─ knowledge/
├─ observability/
├─ governance/
└─ delivery/
```

禁止：

```text
business-service/
agent-service/
capability-service/
...
```

在第一天就全部拆成微服务。

---

# 4. 核心数据模型

## 4.1 Asset Header

所有四层资产统一：

```text
tenant_id
asset_id
kind
name
semantic_version
revision
schema_version
status
owner_id
created_by
created_at
updated_at
content_digest
deleted_at
```

依赖独立表：

```text
asset_dependency
- tenant_id
- source_asset_id
- source_revision/version
- target_asset_id
- target_revision/version
- dependency_type
```

## 4.2 Business Asset

```text
business_model
business_model_revision
business_node
business_edge
business_evidence
business_assertion
business_confirmation
business_pattern
business_impact_task
business_view_layout
```

关键要求：

- semantic snapshot 不可变
- layout 独立存储
- layout API 不产生 semantic revision
- semantic update 必须事务化：
  1. 写 revision
  2. 写 snapshot
  3. 使 confirmation 失效
  4. 使 evaluation 失效
  5. 使 freeze 失效
  6. 创建 impact task
  7. 写 audit/outbox

## 4.3 Data Plane

```text
data_contract
data_contract_version
canonical_object
canonical_field
data_source
data_mapping
managed_schema
managed_object
knowledge_source
knowledge_document
retention_policy
```

## 4.4 Capability

```text
capability
capability_version
capability_implementation
capability_dependency
capability_test_suite
adapter_registration
```

## 4.5 Agent

```text
agent
agent_revision
agent_capability_ref
agent_rule_ref
agent_knowledge_ref
agent_permission_ref
agent_history
```

软删除：

- 不物理删除发布历史
- 被 active Application 引用时禁止删除
- 删除后进入 Trash
- restore 后重验依赖

## 4.6 Application

```text
application
application_revision
application_agent_ref
application_graph
application_shared_context
application_shared_knowledge
application_policy
application_channel_ref
application_runtime_ref
```

## 4.7 Human / Eval / Release

```text
human_task
human_task_decision
execution_checkpoint

evaluation_suite
evaluation_case
evaluation_run
evaluation_result
evaluation_artifact

release_snapshot
release_gate_result
environment_deployment
canary_policy
rollback_record
```

## 4.8 Run / Audit

```text
run
run_step
tool_invocation
runtime_event
trace_link
audit_event
cost_record
```

---

# 5. API 统一规范

统一：

```text
/api/v1/...
```

## 5.1 API 规则

所有写请求：

- tenant 从 AuthContext 获取
- 资源版本采用 optimistic concurrency
- 支持 idempotency key 的高风险接口必须校验
- 返回 `request_id`
- 写操作生成 audit
- 领域错误返回结构化 error code

错误示例：

```json
{
  "error": {
    "code": "BUSINESS_MODEL_VERSION_CONFLICT",
    "message": "业务模型已被其他版本更新",
    "request_id": "...",
    "details": {}
  }
}
```

## 5.2 并发控制

关键资产更新：

```text
If-Match / revision
```

冲突：

```text
409 VERSION_CONFLICT
```

客户端不得静默覆盖。

---

# 6. Reference Business：维修工单

V1 真实闭环使用维修工单验证，但所有领域代码必须通过 Canonical Contract 实现。

## 6.1 Reference Model

角色：

- 报修人
- 客服
- 派单负责人
- 维修人员
- 审核负责人

对象：

- WorkOrder
- Equipment
- Location
- Employee

状态示例：

```text
Draft
PendingAssignment
Assigned
InProgress
PendingApproval
Completed
Rejected
Cancelled
```

Capability：

```text
work_order.query
work_order.get
work_order.create
work_order.assign
work_order.update
work_order.submit
work_order.approve
work_order.reject
work_order.close
asset.lookup
employee.available_workers
notification.send
```

> 所有 `work_order.*` 都只是 Reference Capability。平台框架不得把 `work_order` 写死到通用模块。

---

# 7. S0–S13 总路线

| Stage | 名称 | 核心结果 |
|---|---|---|
| S0 | 工程与架构基线 | 可持续开发的生产 Repo / CI / Compose |
| S1 | Identity / Tenant / Asset Registry | 真实多租户与四层资产基座 |
| S2 | Business Asset + Visual Editor Backend | v1.2 画布从 localStorage 变为服务端资产 |
| S3 | AI Business Analyst | 真实访谈→候选事实→Evidence→确认 |
| S4 | Data Plane | External / Managed / Hybrid 与 Data Contract |
| S5 | Capability Engine | Capability Contract / Gateway / Adapter SDK |
| S6 | Business Agent | 真实 Agent CRUD / Version / Test |
| S7 | Application + Human Task | 多 Agent 应用与 Durable HITL |
| S8 | Evaluation | 自动测试、Regression、Gate |
| S9 | Runtime Adapter | 首个真实 Runtime 完整运行 |
| S10 | Channel Gateway | Web/H5 优先，再微信/企业微信/飞书 |
| S11 | Release Engineering | Env / Freeze / Canary / Rollback |
| S12 | Observability & Governance | Trace / Cost / Policy / Secret / Audit |
| S13 | Package & Delivery | 可交付 Application Package / License / Upgrade |

---

# 8. S0：工程与架构基线

## 目标

将 v1.2 从“可运行原型”升级成“可持续生产开发工程”。

## 范围

### Frontend

- 将当前 v1.2 移入 `apps/web`
- 保留现有 UI / 路由 / Tokens
- `lib/store.tsx` 暂时保留作为 mock fallback
- 建立 API Client 层
- 建立 Error Boundary
- 建立 auth loading / unauthorized / forbidden 状态
- 增加 Playwright

### Backend

创建 FastAPI：

- `/health/live`
- `/health/ready`
- `/api/v1/meta/version`

基础：

- settings
- structured logging
- request_id
- exception mapping
- DB session
- migration
- transaction helper

### DB

建立：

- tenant
- principal
- audit_event
- schema_migration_info

### Infra

Docker Compose：

- postgres
- redis
- minio
- api
- web
- worker

### CI

至少：

Frontend:
```text
lint
typecheck
test
build
playwright smoke
```

Backend:
```text
ruff
mypy/pyright
pytest
alembic check
```

## S0 Gate

必须满足：

- 单命令本地启动
- 新电脑按 README 可启动
- `/` 正常
- `/health/ready` 正常
- PostgreSQL migration 可正向执行
- CI 全绿
- v1.2 页面视觉无结构性回退
- localStorage 模拟仍能演示，但标记为 mock

## S0 产出

```text
S0-RESULT.md
docs/architecture/system-context.md
docs/adr/ADR-001-modular-monolith.md
docker-compose.yml
.env.example
CI config
```

## Cursor 执行提示词

```text
你正在实施 Forma V1 S0：工程与架构基线。

首先阅读：
- README.md
- docs/01-product-blueprint.md
- docs/02-design-system-and-handoff.md
- docs/03-prototype-coverage.md
- docs/06-visual-model-editor.md

目标：
1. 保持 v1.2 UI、IA、Design Tokens 和交互不回退。
2. 建立 production-oriented monorepo。
3. 将现有前端移动/整理到 apps/web，但不要重写页面。
4. 创建 FastAPI 基础服务、PostgreSQL、Redis、MinIO、Worker 和 Docker Compose。
5. 建立 lint/typecheck/test/build/e2e smoke CI。
6. 建立 request_id、结构化日志、统一 error response 和 Alembic。
7. 当前不实现业务模块。

严格禁止：
- 不得修改 Business Asset / Capability / Agent / Application 产品语义。
- 不得用前端 localStorage 冒充生产资产数据库。
- 不得一次拆成十几个微服务。
- 不得连接真实客户 Secret。
- 不得把 Runtime 集成提前塞进 S0。

完成后生成 S0-RESULT.md：
Status / Implemented / Files / Commands / Test Evidence / Remaining Mock / Risks / Next Gate。
```

---

# 9. S1：Identity / Tenant / Asset Registry

## 目标

建立所有后续模块必须依赖的企业级身份和四层资产注册基座。

## 后端

### Identity

先实现：

- Local development identity
- OIDC Provider abstraction
- Access Token 验证
- `AuthContext`
- tenant membership
- role / attribute context

生产第一版可以接 Keycloak 或企业 OIDC，但平台代码只依赖 OIDC abstraction。

### Tenant

实现：

- Tenant CRUD（管理员）
- Tenant status
- Membership
- Environment entitlement

### Asset Registry

统一四层资产：

```text
BUSINESS
CAPABILITY
AGENT
APPLICATION
```

支持：

- Draft
- In Review
- Verified
- Frozen
- Released
- Deprecated
- Archived
- soft delete

## API

```text
GET    /api/v1/me
GET    /api/v1/tenants/current
GET    /api/v1/assets
GET    /api/v1/assets/{id}
POST   /api/v1/assets
PATCH  /api/v1/assets/{id}
DELETE /api/v1/assets/{id}
POST   /api/v1/assets/{id}/restore
GET    /api/v1/assets/{id}/dependencies
```

## Frontend

替换：

- workspace header
- Agent / Business / Capability / Application 的假 tenant 信息

新增：

- Loading
- Unauthorized
- Forbidden
- Tenant disabled

## 安全

后端必须验证 tenant。

测试：

```text
Tenant A 用户请求 Tenant B asset => 403 / 404
```

不能只依赖查询参数 `tenant_id`。

## S1 Gate

- 两个 tenant 自动化隔离测试
- asset soft delete
- immutable released snapshot
- dependency ref 基础验证
- audit 写入
- 前端所有请求使用真实 token context

## Cursor Prompt

```text
实施 Forma V1 S1：Identity / Tenant / Asset Registry。

保持 v1.2 UI 不变，优先替换假 Workspace 与资产元数据。

必须：
- 服务端强制 tenant 隔离
- Asset Header 统一
- 四类资产共用 Registry
- Released/Frozen 资产不可直接编辑
- soft delete 不破坏发布历史
- 写操作生成 audit

测试必须包含跨租户拒绝。
不得开始实现 AI Analyst、Runtime 或 Channel。
输出 S1-RESULT.md。
```

---

# 10. S2：Business Asset + Visual Business Model Editor 后端化

## 目标

把 v1.2 最核心的 Visual Business Model Editor 从 localStorage 迁移为真实 Business Asset 服务。

## 数据模型

```text
business_model
business_model_revision
business_model_snapshot
business_node
business_edge
business_view_layout
business_impact_task
```

### Semantic

```text
nodes
edges
rules
states
```

### Layout

```text
node_positions
viewport
zoom
mode
groups
```

必须不同 API。

## API

```text
POST /api/v1/business-models
GET  /api/v1/business-models/{id}

GET  /api/v1/business-models/{id}/semantic
PUT  /api/v1/business-models/{id}/semantic

GET  /api/v1/business-models/{id}/layout
PUT  /api/v1/business-models/{id}/layout

GET  /api/v1/business-models/{id}/revisions
GET  /api/v1/business-models/{id}/revisions/{revision}
GET  /api/v1/business-models/{id}/diff?from=&to=

GET  /api/v1/business-models/{id}/impact
```

## 事务规则

Semantic update：

```text
BEGIN
validate revision
validate graph
create semantic revision
create immutable snapshot
invalidate confirmations
invalidate evaluations
invalidate freezes
create impact task
audit
outbox
COMMIT
```

Layout update：

```text
validate nodes exist
save layout
audit(optional)
NO semantic revision
NO evaluation invalidation
```

## Frontend

现有：

```text
VisualBusinessModelEditor
```

保持 UI。

替换：

```text
Store local visualModel
→ BusinessModel API
```

Undo/Redo：

- 当前客户端可保留本地交互历史
- 最终 Semantic Undo 必须调用后端形成新 revision
- 不能“回到旧 revision 并恢复旧 confirmation”

## Tests

- node add/edit/delete
- edge add/edit/delete
- invalid edge
- duplicate ID
- revision conflict
- layout does not invalidate
- semantic does invalidate
- semantic undo creates newer revision
- Released revision immutable

## S2 Gate

使用真实 PostgreSQL：

1. 新增节点 → revision +1
2. 拖节点 → revision 不变
3. Semantic 修改后 Evaluation / Freeze 失效
4. 刷新浏览器数据保持
5. 两浏览器并发修改产生 409 而不是覆盖

## Cursor Prompt

```text
实施 S2：Business Asset + Visual Model 服务端化。

以 v1.2 `lib/visual-model.ts` 与 `components/visual-business-model-editor.tsx`
的行为作为 UI/领域验收参考，但不要把 localStorage 实现照搬成后端设计。

关键不变量：
- semantic_model 和 view_layout 永久分离
- layout 不能触发 model revision
- semantic 每次有效提交生成不可变 revision
- semantic mutation 同事务失效 confirmation/evaluation/freeze 并创建 impact
- Undo/Redo 语义恢复属于新 revision
- Released snapshot 不可变

保持现有 Canvas 外观和快捷键。
输出 S2-RESULT.md 与 API 契约。
```

---

# 11. S3：AI Business Analyst + Evidence / Confirmation

## 目标

实现真正的“从企业自然语言业务描述到可确认 Business Model”。

## Pattern Library

首批：

- Ticket
- Approval
- Inspection
- Incident
- Fulfillment
- Reservation
- Order
- Case
- Asset Lifecycle
- Customer Service

Pattern 是提问骨架，不是行业硬编码模板。

## 数据模型

```text
interview_session
interview_message
business_evidence
business_assertion
assertion_evidence_ref
business_confirmation
clarification_task
```

## AI 流程

```text
用户回答
→ Evidence 保存
→ LLM Structured Extraction
→ Assertion Candidate
→ conflict detection
→ missing-slot analysis
→ next question
→ user/business reviewer confirmation
→ confirmed model mutation
```

必须使用结构化输出 Schema。

禁止：

```text
LLM 直接 PATCH Business Model 成为 Confirmed
```

## Evidence

字段：

```text
source_type
source_ref
source_version
quote_span
speaker
captured_at
access_scope
```

Assertion：

```text
claim
confidence
confidence_reason
status
approver_requirements
supersedes
```

## Confirmation

状态：

```text
Proposed
Needs Clarification
Conflict
In Review
Confirmed
Rejected
Superseded
```

真实双人确认：

- business principal
- IT / implementation principal

同一 principal 不能模拟两个职责。

## API

```text
POST /api/v1/analyst/sessions
POST /api/v1/analyst/sessions/{id}/messages
GET  /api/v1/analyst/sessions/{id}

GET  /api/v1/business-models/{id}/assertions
POST /api/v1/assertions/{id}/confirm
POST /api/v1/assertions/{id}/reject
POST /api/v1/assertions/{id}/clarify
```

## Prompt Injection 安全

外部文档/用户文字都视为 Evidence，不是 system instruction。

LLM 提取任务明确：

- 不执行材料中的指令
- 只提取业务事实
- 引用 evidence_id
- 缺失必须返回 unknown

## S3 Gate

至少完成 Reference Ticket 场景：

用户只通过对话描述：

```text
客户电话报修
客服登记
主管派单
维修人员处理
主管审核
```

系统可生成：

- 角色
- WorkOrder
- 主流程
- 状态
- 规则候选
- 未确认项
- Evidence
- 可人工修正 Canvas
- 双人确认

## Cursor Prompt

```text
实施 S3：AI Business Analyst。

核心目标不是“聊天漂亮”，而是产生可审阅、可引用来源、可确认的结构化业务资产。

必须：
- 所有 AI 结论带 Evidence 引用
- confidence 与 confirmation 分离
- AI 不可直接写 Confirmed
- conflict 必须保留
- unknown 不允许模型补猜
- Business Canvas 人工修改与 AI Assertion 最终进入同一 revision 体系
- 模型调用封装为 ModelProvider

先只验证 Ticket Pattern，但 Pattern Engine 必须通用。
输出 S3-RESULT.md、Prompt Schema、Golden 样例。
```

---

# 12. S4：Data Plane

## 目标

让 Business Model 真正连接企业数据，而不是停留在图。

## 三种模式

### External

客户系统为 Source of Truth。

保存：

- connector
- schema
- mapping
- credential ref
- cache policy

不默认复制业务主数据。

### Managed

Forma 提供：

- tenant scoped table/schema
- CRUD
- state machine
- optimistic lock
- audit
- idempotency
- migration
- export

### Hybrid

按对象/字段指定 ownership。

禁止同一字段双主写入。

## Managed Runtime 隔离

MVP：

```text
一个 PostgreSQL Cluster
每 Tenant Dedicated Schema
```

高等级：

```text
Dedicated Database
```

平台控制表与业务交易表逻辑分开。

## Business Data Contract

```text
object_id
version
fields[]
types
required
enum
relations
sensitivity
ownership
source
refresh_policy
field_permission
```

## Mapping

必须支持：

- field rename
- enum mapping
- unit conversion
- timezone
- nullable
- API parameter
- output normalization

## API

```text
GET/POST /api/v1/data-contracts
POST /api/v1/data-contracts/{id}/versions
POST /api/v1/data-sources
POST /api/v1/data-mappings
POST /api/v1/data-mappings/{id}/validate

POST /api/v1/managed/{object}/query
POST /api/v1/managed/{object}
PATCH /api/v1/managed/{object}/{id}
```

Managed 通用接口内部仍必须经过对象 Schema / Policy，不允许任意表名/SQL。

## Reference Business

WorkOrder 使用 Managed：

- work orders

Equipment 使用 External mock REST：

- equipment service

构成真实 Hybrid。

## S4 Gate

- Tenant A WorkOrder 不可被 B 读取
- 状态机非法跳转拒绝
- 重复 idempotency key 不重复创建
- Mapping 未知 enum 报错
- Managed export 可用
- External source 失败显示明确 error contract

## Cursor Prompt

```text
实施 S4 Data Plane。

Reference Business：
- WorkOrder = Managed
- Equipment = External REST mock
形成 Hybrid。

必须：
- Data Contract 先于运行数据
- Managed 不允许 Agent 自由 SQL
- tenant schema 隔离必须后端测试
- 写操作幂等
- 状态转换校验
- Mapping 不得对未知枚举静默降级
- Knowledge / temp runtime / transaction data 生命周期分开

输出 S4-RESULT.md 与迁移策略。
```

---

# 13. S5：Capability Engine + Adapter SDK

## 目标

建立 Forma 的核心技术护城河：

> Agent 永远只看到 Business Capability，不知道实际数据来自哪里。

## Contract

```text
id
version
business_meaning
input_schema
output_schema
permission
side_effect
confirmation
timeout
preconditions
postconditions
error_contract
sla
compatibility
dependency_refs
test_suite_refs
```

## Side Effect

```text
Read
Write
ExternalSideEffect
HumanDecision
```

## Error Contract

固定：

```text
PermissionDenied
ValidationFailed
StateConflict
NotFound
RateLimited
TimeoutUnknownOutcome
Unavailable
HumanRejected
```

## Gateway

```text
Runtime
→ Capability Gateway
→ auth/policy
→ schema validate
→ confirmation
→ idempotency
→ adapter
→ result normalize
→ audit/trace
```

## Implementations

V1：

- Managed Runtime
- REST
- Human Task
- Database Read
- MCP
- Workflow

## Adapter SDK

```python
describe()
validate()
execute()
health()
reconcile()
cancel() # optional
```

Context：

```text
TenantContext
ActorContext
TraceContext
SecretResolver
Budget
IdempotencyContext
```

## REST Adapter

支持：

- base URL ref
- credential ref
- path mapping
- request transform
- response normalize
- timeout
- retry policy
- reconcile

SSR F 防护：

- allowlist
- private IP policy
- DNS rebinding protection
- egress restriction

## MCP Adapter

不得默认可信。

每个 Tool 映射到 Capability Contract。

## DB Adapter

只允许：

- pre-registered query
- parameterized query
- bounded result

不提供模型任意 SQL shell。

## S5 Gate

真正从 Agent 之外调用：

```text
work_order.create
```

能根据实现绑定执行 Managed Runtime。

`asset.lookup` 能执行 External REST。

同一 Capability 替换 Implementation 后上层 Contract 不变。

## Cursor Prompt

```text
实施 S5 Capability Engine。

这是 Forma 的核心边界。

禁止 Runtime/Agent 直接连接数据库或客户 REST。
所有工具调用必须经过 Capability Gateway。

实现：
Contract Registry
Gateway
Managed Adapter
REST Adapter
Human Task Adapter skeleton
DB Read Adapter
MCP Adapter skeleton
Adapter SDK

必须具备：
schema validation
permission hook
confirmation hook
timeout
idempotency
typed error
trace
audit
reconcile

输出 S5-RESULT.md 和 SDK README。
```

---

# 14. S6：Business Agent Center

## 目标

将原型 Agent CRUD 替换为真实资产服务。

## Agent Definition

六要素：

```text
Role
Context
Capability
Rule
Permission
Interaction
```

Knowledge 是引用。

Runtime 行为不存到 Agent Prompt 中重复实现。

## API

```text
GET    /api/v1/agents
POST   /api/v1/agents
GET    /api/v1/agents/{id}
PATCH  /api/v1/agents/{id}
POST   /api/v1/agents/{id}/copy
DELETE /api/v1/agents/{id}
POST   /api/v1/agents/{id}/restore

GET    /api/v1/agents/{id}/versions
POST   /api/v1/agents/{id}/versions
POST   /api/v1/agents/{id}/validate
POST   /api/v1/agents/{id}/test
```

## AI Generate Agent

根据：

- Confirmed Business Model
- Capability refs

AI 可建议：

- Role
- Interaction
- capability selection
- boundary

输出永远 Draft。

## Delete

删除前检查：

- Application refs
- Frozen snapshot refs
- Scheduled execution refs

不能只检查当前 UI Application。

## S6 Gate

前端 `/agents`：

- list
- create
- edit
- copy
- soft delete
- restore
- versions
- validate
- test

全部真实 API。

## Agent Test

不是只看自然语言。

检查：

- Capability selection
- boundary
- clarification
- permission
- write confirmation
- failure handling

## Cursor Prompt

```text
实施 S6 Business Agent。

必须保留 v1.2 Agent Center UX。

Agent 只维护业务层：
Role/Context/Capability/Rule/Permission/Interaction/Knowledge refs。

不要：
- 每个 Agent 自己实现 conversation loop
- 在 Agent Prompt 里硬编码 API
- 让 Agent 绕过 Capability Gateway

复制创建独立 Draft；
删除是软删除；
被任何有效 Application/Snapshot 引用时不得删除。

输出 S6-RESULT.md。
```

---

# 15. S7：Application Builder + Human Task

## 目标

真正让多个 Business Agent 组成一个最终交付 Application。

## Application Graph

序列化协议独立于 UI Canvas：

```text
nodes:
  agent / router / supervisor / human / merge
edges:
  trigger / route / handoff / next / fallback
policies:
  budget
  max_depth
  timeout
  context_whitelist
```

## 六种模式

### Router
只选择一个 Agent。

### Supervisor
拆任务、协调多个 Agent。

### Pipeline
顺序执行。

### Parallel
并发 + 聚合。

### Handoff
责任移交。

### Human-in-the-loop
持久化暂停。

## Shared Context

只允许声明白名单字段。

可信字段：

```text
tenant_id
actor_id
permissions
```

不能由模型覆盖。

## Human Task

数据：

```text
task_id
type
run_id
checkpoint_id
owner_group
assignee
deadline
priority
input_digest
resource_version
status
decision
reason
```

状态：

```text
Pending
Claimed
Approved
Rejected
Expired
Escalated
Resuming
Completed
Failed
```

批准不等于业务执行完成。

## S7 Gate

真实场景：

```text
用户创建高优先级工单
→ WorkOrder Agent
→ Human approval
→ run pause
→ 管理员批准
→ resume
→ create
→ completed
```

重复批准不能重复写业务数据。

## Cursor Prompt

```text
实施 S7 Application + Human Task。

Application 是最终交付单位，不是“超级 Prompt”。

先实现：
- Router
- Supervisor 基础
- Handoff
- Human-in-the-loop
Pipeline/Parallel 允许先完成契约与基础执行，再逐步增强。

Human Task 必须数据库持久化，可在服务重启后恢复。
批准和执行状态分开。
resume 必须幂等。
不同 Agent 共享 Context 采用字段白名单。

输出 S7-RESULT.md，附持久化恢复测试证据。
```

---

# 16. S8：Evaluation Engine

## 目标

建立“业务变更 → 测试失效 → 重新回归 → Gate”的真实闭环。

## Test Sources

Business Model 自动派生：

- role permission
- state transition
- normal flow
- exception
- invalid input
- duplicate request
- cross-tenant
- tool failure
- human task
- channel identity

Capability：

- contract
- error
- timeout
- unknown outcome

Agent：

- intent
- capability selection
- boundary
- clarification

Application：

- route
- handoff
- multi-agent result

## Evaluation 类型

```text
Contract Test
Integration Test
Behavior Evaluation
E2E Channel Test
Offline Regression
```

第一版优先确定性测试。

LLM-as-judge 后置。

## Snapshot Binding

Evaluation Run 必须绑定：

```text
business_revision
capability_versions
agent_revisions
application_revision
runtime_config_digest
model_config_digest
dataset_digest
```

任何相关变更：

```text
evaluation = stale
```

## API

```text
POST /api/v1/evaluation/suites/generate
POST /api/v1/evaluation/runs
GET  /api/v1/evaluation/runs/{id}
GET  /api/v1/evaluation/gate/{application_id}
```

## S8 Gate

修改一条业务规则后：

```text
旧 evaluation 自动失效
release gate blocked
```

重新运行：

```text
PASS
release gate ready
```

## Cursor Prompt

```text
实施 S8 Evaluation。

不要先做复杂 LLM judge。
第一版先把确定性 Contract / Permission / State / Route / Human Task 测试做扎实。

Evaluation 必须绑定完整资产 snapshot digest。
任何 Business/Capability/Agent/Application/Runtime/Channel 关键配置变化都应使对应结果 stale。

AI 生成测试只是候选，Golden assertion 需要人工确认字段。

输出 S8-RESULT.md、Evaluation JSON Schema、Gate Evidence。
```

---

# 17. S9：Runtime Adapter

## 目标

让 Forma Application 真正运行。

## Kernel 与 Runtime 边界

Forma Kernel 负责：

- Identity
- Behavior Policy
- Context scope
- Memory policy
- Tool authorization
- Budget
- Human checkpoint
- Audit
- Trace
- normalized streaming event

Runtime 负责：

- Agent execution
- model loop
- tool call mechanics
- state orchestration

## Runtime Adapter Contract

```text
run
stream
cancel
checkpoint
resume
invoke_tool
normalize_event
discover_capabilities
health
```

## 首个 Runtime

建议：

```text
LangGraphRuntimeAdapter
```

后续：

```text
DeepSeekHarnessRuntimeAdapter
EinoRuntimeAdapter
...
```

## Behavior Policy

优先级：

```text
Platform
> Organization
> Application
> Agent
```

下级只能收紧。

Platform Policy：

- no fabricated tool result
- permission denied behavior
- no-data behavior
- clarification
- write confirmation
- secret protection
- cross-tenant forbidden
- social/capability overview

## S9 Gate

从 Web 调用：

```text
用户：
帮我创建一个维修工单

Runtime：
clarify missing fields
Capability Gateway
Human confirmation if required
Managed Runtime write
stream result
```

Trace 完整。

## Cursor Prompt

```text
实施 S9 Runtime Adapter。

先接一个 Runtime，不要同时集成三个框架。

Runtime 不拥有业务权限。
所有 Tool 调用仍必须通过 Forma Capability Gateway。

实现统一事件：
run_started
message_delta
tool_requested
tool_started
tool_completed
human_waiting
run_completed
run_failed

确保 cancel/checkpoint/resume Contract 可扩展。
输出 S9-RESULT.md 和 Runtime Compatibility Matrix。
```

---

# 18. S10：Channel Gateway

## 目标

让同一 Agent Application 在多个用户入口工作。

## 优先顺序

### S10.1 Web / H5
必须先完成。

### S10.2 微信服务号
第二优先。

### S10.3 企业微信
### S10.4 飞书
### S10.5 钉钉
### S10.x Slack / Teams / App SDK / API Webhook

## Unified Message

```text
channel
tenant
channel_user_id
conversation_id
message_id
timestamp
text
attachments
reply_context
metadata
```

## Identity Chain

```text
channel_user_id
→ identity_binding
→ enterprise_actor_id
→ tenant membership
→ role/attribute
→ application entitlement
```

## 微信服务号

至少：

- callback verification
- signature
- duplicate message prevention
- OpenID
- UnionID optional
- OAuth H5
- AppID / Secret Vault refs
- text/image/file normalization
- service notification policy abstraction

禁止将 AppSecret 保存在前端。

## Web/H5

- SSE / WebSocket（依据 Runtime streaming）
- file upload
- auth
- reconnect
- run status
- Human Task redirect

## S10 Gate

同一个 Application：

- Web 正常
- 微信服务号正常（真实测试账号/沙箱/合法配置）
- 同一企业人员身份绑定后权限一致
- 匿名微信用户不能获得员工权限

## Cursor Prompt

```text
实施 S10 Channel Gateway。

严格按 Web/H5 → 微信服务号 → 企业微信/飞书 顺序。

Channel 只做：
message normalization
identity mapping
signature/auth
attachment
rate limit
delivery

不得把业务逻辑复制进 Channel Adapter。

Credential 只能存 Vault ref。
测试 callback replay、重复消息、身份过期、离职解绑。

输出 S10-RESULT.md，每个渠道标记：
Implemented / Verified / External-Gate / Not-Started。
```

---

# 19. S11：Release Engineering

## 目标

把原型 Release 页面变成真实版本和环境发布系统。

## Environment

```text
Dev
Test
Staging
Prod
```

## Freeze

Freeze 创建：

- immutable application snapshot
- dependency lock
- evaluation references
- digest
- migration plan refs
- runtime/channel compatibility

## Gate

至少检查：

- Business confirmed
- dependency complete
- Evaluation current & passed
- permission reviewed
- Runtime compatible
- Channel compatible
- migration reversible/acknowledged
- budget policy

## Canary

记录：

```text
percentage
start_at
observation_window
error_threshold
business_kpi_threshold
```

V1 可先支持 Application version routing。

## Rollback

注意：

> 回滚 Application Version ≠ 撤销已发生业务副作用。

UI 和后端都必须明确。

## S11 Gate

- 未评测不能冻结
- Freeze 后改资产不能修改原 Frozen，必须新 Draft
- Test → Staging → Prod
- Canary 10% → 50% → 100%
- threshold 失败可停止/回滚
- Production 版本依赖不可变

## Cursor Prompt

```text
实施 S11 Release Engineering。

前端现有 Release 页面保留视觉结构，替换为真实服务端状态。

Freeze 必须生成 immutable snapshot + digest。
Production 不解析 latest。
Evaluation 必须绑定同一 snapshot。
Rollback 不得声称撤销外部副作用。

输出 S11-RESULT.md、Release Manifest Schema、Rollback Boundary。
```

---

# 20. S12：Observability / Governance / Secret / Audit

## 目标

达到企业可诊断、可审计、可治理。

## Trace

统一：

```text
run_id
trace_id
tenant_id
actor_id
application_revision
agent_revision
capability_version
adapter_version
runtime_version
```

链路：

```text
Channel
→ Kernel
→ Agent
→ Capability
→ Adapter
→ Data/Human
```

## Metrics

- run success
- task completion
- first response latency
- total duration
- tool latency
- human wait time
- capability error
- token
- model cost
- business KPI

## Audit

必须区分：

```text
Debug Trace
Business Audit
```

Trace 可采样/清理。

关键业务 Audit 不随对话 TTL 清除。

## Permission

RBAC + ABAC：

- platform role
- tenant role
- object permission
- field permission
- capability action
- application entitlement

管理员不默认拥有所有业务数据读权限。

## Secret

抽象：

```text
SecretResolver
secret_ref
```

生产：

- HashiCorp Vault / Cloud KMS / enterprise secret provider 可插拔
- 不进入 Prompt
- 不进入 Browser
- 不进入 Package
- 日志脱敏

## Security Tests

- cross tenant
- IDOR
- SSRF
- prompt injection
- secret leak
- replay
- duplicate write
- authorization bypass
- tool parameter tampering

## Cursor Prompt

```text
实施 S12 Enterprise Governance。

必须：
- OpenTelemetry 全链路
- Debug Trace 与 Business Audit 分离
- Secret 仅引用
- RBAC + ABAC server-side
- capability-level authorization
- sensitive log redaction
- retention policy

不要把治理只做成 /governance 页面表单。
必须有服务端 enforcement tests。

输出 S12-RESULT.md 与 Threat Model。
```

---

# 21. S13：Application Package / Commercial Delivery

## 目标

形成真正可交付的 Agent Application 产品包。

## Package

```text
manifest
dependency-lock
business refs/snapshots
capability definitions
agent definitions
application graph
runtime config
channel config
policy
knowledge version refs
evaluation suite/report
migration plan
rollback plan
SBOM
digests
signature
```

Secret：

```text
仅 secret_ref
```

## 包类型

### Managed SaaS
发布到平台环境。

### Customer Dedicated
客户专属部署。

### Offline / Private
Docker images + package + license + upgrade bundle。

## License

V1 定义协议即可：

- tenant
- allowed application
- allowed agent count
- capabilities
- channels
- expiry
- quota
- signature

不要先做复杂计费中心。

## Upgrade

升级包必须：

- from version
- to version
- minimum platform version
- migration
- extension conflict
- rollback support
- breaking changes

## S13 Gate

- Package 可验证签名
- Secret 不在包中
- SBOM
- dependency lock
- Evaluation reference
- install validation
- upgrade dry-run
- rollback boundary

## Cursor Prompt

```text
实施 S13 Delivery。

目标是生产级 Application Package，而不是原型 JSON 导出。

必须：
- immutable manifest
- dependency lock
- digest/signature
- SBOM
- secret refs only
- evaluation evidence
- migration/rollback metadata
- compatibility check

先实现协议、生成、验证、导入 dry-run。
License 先做签名授权协议，不扩展复杂计费系统。

输出 S13-RESULT.md 与 Package Spec。
```

---

# 22. 前端原型到生产前端的迁移策略

不能“删掉原型重新写 UI”。

按照页面逐项替换 Store：

## Step 1

```text
lib/store.tsx
```

保留 demo provider，同时新增：

```text
lib/api/
lib/query/
lib/auth/
```

## Step 2

每完成一个 Stage：

```text
Business
Data
Capability
Agent
Application
...
```

对应页面切换真实 API。

## Step 3

开发环境提供：

```text
VITE_FORMA_DATA_MODE=mock|api
```

最终 Prod 强制：

```text
api
```

## Step 4

在所有真实页面增加：

- skeleton
- empty
- forbidden
- retry
- conflict
- stale revision
- server validation
- audit link

---

# 23. Figma / Cursor 协作规范

v1.2 已有 Design Tokens，不要用截图猜 UI。

Figma 文件：

```text
00 Foundations
01 Components
02 Product Flows
03 Screens
04 Handoff
```

Cursor 每个 UI Task：

1. 先运行当前页面。
2. 阅读 `design/tokens.json`。
3. 如果有 Figma node URL，读取对应节点。
4. 复用现有组件。
5. 不创建另一套颜色/spacing。
6. 完成后浏览器真实验收。
7. 更新 component map。

禁止：

- 截图转 UI 后覆盖现有组件体系
- 每页自行创建 Button/Input
- 将生产 Secret 上传到设计工具

---

# 24. 数据迁移与兼容策略

## 24.1 Asset Schema Version

所有导入导出：

```text
schema_version
```

旧版本必须：

- migrate
- reject with clear error
- never silently reinterpret

## 24.2 Business Model

Semantic Snapshot 永久保留。

Layout 可以独立迁移。

## 24.3 Capability Contract

兼容分类：

```text
compatible
conditionally-compatible
breaking
```

Breaking：

- required input 新增
- output semantic changed
- permission widened
- side effect changed
- error contract removed

## 24.4 Adapter

Adapter Version 与 Capability Version 独立。

## 24.5 Managed Business Runtime

Migration：

```text
plan
dry-run
backup
execute
verify
rollback window
```

---

# 25. 安全基线

## 必须从 S0 起具备

- no secret in repo
- no secret in frontend
- dependency scanning
- security headers
- request size limit
- file type/size limit
- path normalization
- structured error without stack leak
- log redaction

## S4 起

- tenant DB isolation
- row/schema boundary tests

## S5 起

- SSRF
- adapter allowlist
- idempotency
- tool permission

## S9 起

- prompt injection boundary
- tool result validation
- budget / loop limit

## S10 起

- webhook signature
- replay prevention
- user binding

## S12

完整 Threat Model。

---

# 26. 测试金字塔

## Unit

- domain invariants
- model diff
- state transition
- gate
- permission

## Contract

- JSON Schema
- Capability
- Adapter
- Runtime
- Channel

## Integration

- PostgreSQL
- Redis
- MinIO
- External mock system

## E2E

Playwright：

```text
business discovery
→ model edit
→ confirm
→ data mapping
→ capability
→ agent
→ application
→ evaluation
→ release
→ Web chat
```

## Security

独立 CI job。

## Regression

每个 Stage 不得破坏之前 Stage Gate。

---

# 27. CI/CD

Pull Request：

```text
frontend lint
frontend typecheck
frontend unit
backend lint
backend type
backend unit
migration check
contract test
security static scan
build
Playwright smoke
```

Main：

```text
integration
docker image
SBOM
image scan
```

Release：

```text
full regression
package
sign
staging deploy
```

---

# 28. Cursor 统一执行规范

每轮 Cursor 指令顶部必须带：

```text
你正在 Forma V1 项目中工作。

规则：
1. 先检查当前代码和上一个 RESULT，不直接修改。
2. 不改变已冻结产品 IA 和 v1.2 视觉语言，除非本任务明确要求。
3. 不把 mock 标记为 production。
4. 不虚构已验证的外部系统。
5. 不绕过测试解决失败。
6. 不删除已有领域测试。
7. 任何数据库变更必须 Alembic。
8. 任何权限必须 server-side。
9. 任何高风险写操作考虑幂等、审计和未知结果。
10. 只实施本 Stage 范围。
```

---

# 29. Cursor 每阶段结果回传格式

要求 Cursor 最终输出：

```markdown
# Sx RESULT

Status:
PASS | PASS_WITH_GATES | FAIL

## Implemented
- ...

## Architecture Decisions
- ...

## Files Changed
- ...

## Database Migrations
- ...

## API
- ...

## Frontend
- ...

## Tests
Command:
Result:

## Security Checks
- ...

## Manual Verification
- ...

## Remaining Mock
- ...

## Known Limitations
- ...

## External Gates
- ...

## Risks
- ...

## Next Stage Preconditions
- ...
```

**没有 Test Evidence 的 RESULT 不接受。**

---

# 30. 开发过程中 GPT ↔ Cursor 闭环建议

你当前可以采用固定循环：

```text
Forma Implementation Plan
        ↓
GPT 生成当前 Stage Cursor 指令
        ↓
Cursor 执行
        ↓
Cursor 输出 Sx-RESULT.md
        ↓
把 RESULT 给 GPT
        ↓
GPT 审查 Gate
        ↓
PASS → 下一 Stage
FAIL → 生成 Fix 指令
```

不要：

```text
一次把 S0-S13 全发给 Cursor 自动跑
```

原因：

- 上游架构错误会指数放大
- 企业级权限和数据隔离必须逐阶段确认
- Runtime / Channel 有外部 Gate
- 每个 Stage 会产生新的真实约束

---

# 31. 推荐 Milestones

## M0 — Engineering Ready

包含：

```text
S0
```

结果：

生产工程基线成立。

---

## M1 — Business Modeling Ready

包含：

```text
S1
S2
S3
```

结果：

真实用户可以：

```text
登录
→ 访谈
→ Evidence
→ Business Model
→ Visual Edit
→ Confirm
```

这是 Forma 第一核心价值闭环。

---

## M2 — Capability Ready

包含：

```text
S4
S5
```

结果：

```text
Business Model
→ Data Contract
→ Capability
→ Real execution
```

这是第二核心价值闭环。

---

## M3 — Agent Application Ready

包含：

```text
S6
S7
S8
S9
```

结果：

```text
Business Capability
→ Business Agent
→ Multi-Agent Application
→ Human Task
→ Evaluation
→ Runtime
```

至此可以做内部 Alpha。

---

## M4 — Customer Pilot Ready

包含：

```text
S10
S11
S12
```

结果：

- Web
- 微信服务号（或目标渠道）
- Release
- Audit
- Trace
- 企业权限

可以进入真实客户试点。

---

## M5 — Commercial Delivery Ready

包含：

```text
S13
```

结果：

应用可规范交付、升级和授权。

---

# 32. 风险清单

## R1：Business Analyst 看起来聪明，但业务模型不可靠

解决：

- Evidence
- Assertion
- Confirmation
- Unknown
- Conflict
- Golden Business cases

## R2：Managed Runtime 演化成 ERP

解决：

Managed 只实现：

- schema
- CRUD
- state
- audit
- event
- policy

复杂业务能力仍通过 Capability / Workflow。

## R3：Runtime 反向绑架 Forma 模型

解决：

所有 Runtime 只能通过 Adapter。

## R4：Capability 变成 Tool 别名

解决：

Capability 必须有业务语义、权限、Contract、Side Effect、Error、SLA、Dependency。

## R5：多 Agent 只靠 Supervisor Prompt

解决：

Application Graph 和 Handoff Contract 服务端化。

## R6：Human Task 不 Durable

解决：

数据库 checkpoint + idempotent resume。

## R7：发布 Gate 只在前端

解决：

Release Service 服务端强制。

## R8：渠道身份映射错误造成越权

解决：

Channel Identity → Enterprise Actor → Policy。

## R9：用户上传 API / OpenAPI 带 SSRF

解决：

Adapter Gateway egress controls。

## R10：Visual Editor 大图性能

V1 先给出业务模型节点合理上限和分页/子图能力；后续再评估专用图组件。

---

# 33. Definition of Done

Forma V1 不是“所有页面都能点”即完成。

必须至少满足：

## Business

- AI 生成业务模型有 Evidence
- 人工可修改
- 真实多人确认
- revision / diff / impact 正确

## Data

- External / Managed / Hybrid 可运行
- tenant isolation 有后端测试

## Capability

- Contract / Adapter / Gateway
- 权限 / 幂等 / Error Contract

## Agent

- CRUD / copy / version / test
- 不直连业务系统

## Application

- 多 Agent
- 至少 Router/Supervisor/Handoff/HITL 可运行

## Evaluation

- 资产变更后旧结果失效

## Runtime

- 一个真实 Runtime 完整运行

## Channel

- Web/H5
- 至少一个真实移动通讯渠道

## Release

- Freeze / Test / Staging / Prod / rollback

## Governance

- Auth
- Tenant
- Audit
- Secret
- Trace

## Delivery

- Package
- digest/signature
- dependency lock
- upgrade metadata

---

# 34. 推荐的第一批真实开发顺序

实际开始 Cursor 时，我建议不要机械一次写完 S0–S13 的所有指令。

立即执行：

```text
S0
```

通过后：

```text
S1
```

再：

```text
S2
```

到 S2 完成时先进行一次人工产品验收，因为：

> v1.2 Visual Business Model Editor 是当前已经确认的核心 UI 资产，服务端化后必须保证使用体验没有下降。

之后：

```text
S3 → S4 → S5
```

S5 完成后形成第一个重大 Gate：

> **“从业务到可执行 Capability” 是否真正成立。**

若成立，再进入：

```text
S6 → S7 → S8 → S9
```

完成内部 Alpha。

最后：

```text
S10 → S11 → S12 → S13
```

进入客户试点和商业交付。

---

# 35. 第一轮正式 Cursor 指令应该是什么

不要从“实现整个平台”开始。

第一条正式实施指令应当是：

> **FORMA-S0 — Production Engineering Baseline**

它只负责：

```text
工程目录
API skeleton
PostgreSQL
Redis
MinIO
Docker Compose
CI
Migration
Request Context
Logging
E2E smoke
保留 v1.2 前端
```

S0 PASS 之后，才进入平台业务功能开发。

---

# 36. 最终架构目标

```text
                         ┌────────────────────────┐
                         │    Channel Gateway     │
                         │ Web/微信/企微/飞书/... │
                         └───────────┬────────────┘
                                     │
                              Identity Mapping
                                     │
                         ┌───────────▼────────────┐
                         │ Platform Agent Kernel  │
                         │ Policy / Context / HITL│
                         └───────────┬────────────┘
                                     │
                         ┌───────────▼────────────┐
                         │    Runtime Adapter     │
                         │ LangGraph / DSH / ...  │
                         └───────────┬────────────┘
                                     │
                         ┌───────────▼────────────┐
                         │ Capability Gateway     │
                         └──────┬────┬────┬──────┘
                                │    │    │
                         REST/MCP   DB   Managed
                                │    │    │
                                └────┴────┘
                                     │
                             Human Task / Data
                                     │

┌──────────────────────────────────────────────────────────────┐
│                        CONTROL PLANE                         │
│                                                              │
│ Business Asset → Capability Asset → Agent Asset → Application│
│      │                │               │              │        │
│ Evidence          Contract         Business       Multi-Agent │
│ Confirmation      Mapping          Role           Graph       │
│ Version           Adapter          Rule           Channel     │
│ Diff/Impact       Dependency       Permission     Runtime     │
│                                                              │
│              Evaluation → Freeze → Release → Package         │
└──────────────────────────────────────────────────────────────┘
```

---

# 37. 最后一个项目原则

如果后续开发过程中出现这样的选择：

> “为了快速让 Agent 跑起来，是不是直接把 API、权限、流程写到 Prompt 里？”

答案统一是：

> **不允许。**

Forma 的长期价值恰恰来自：

```text
Business Model
        ↓
Business Data Contract
        ↓
Business Capability
        ↓
Business Agent
        ↓
Agent Application
        ↓
Runtime / Channel
```

业务资产与底层 Runtime、模型、客户系统解耦。

这也是所有 S0–S13 实施决策的最高优先级判断标准。

---

# Appendix A：每个 Stage 的复杂度参考

| Stage | 相对复杂度 | 外部依赖 |
|---|---:|---|
| S0 | 中 | 低 |
| S1 | 中高 | OIDC |
| S2 | 中高 | 低 |
| S3 | 高 | LLM |
| S4 | 高 | 客户数据/API |
| S5 | 高 | API/MCP |
| S6 | 中 | S5 |
| S7 | 很高 | Runtime/Human |
| S8 | 高 | S2–S7 |
| S9 | 很高 | Runtime |
| S10 | 很高 | 微信/企微/飞书审核与凭证 |
| S11 | 高 | Runtime/Infra |
| S12 | 高 | Security/Observability |
| S13 | 高 | Deployment/License |

---

# Appendix B：建议先冻结的 ADR

S0 建议创建：

```text
ADR-001 Modular Monolith First
ADR-002 Four-layer Asset Model
ADR-003 Semantic vs View Layout Separation
ADR-004 Capability Gateway Mandatory
ADR-005 Runtime Adapter Boundary
ADR-006 Data-in-Place First / Managed Optional
ADR-007 Immutable Release Snapshot
ADR-008 Server-side Tenant Enforcement
ADR-009 Durable Human Task
ADR-010 Evidence-first Business Modeling
```

这些 ADR 一旦批准，Cursor 后续不得擅自修改。

---

**文档结束。**
