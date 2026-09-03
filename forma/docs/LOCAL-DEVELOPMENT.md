# Forma 本地开发启动指南

本文档是 Forma 本地开发的正式入口。目标：拉取仓库后，尽可能通过**一个命令**完成环境检查、中间件、Forma Migration、Backend、Frontend，并拿到可访问地址。

> **当前开发状态（重要）**  
> 当前处于 **S4-G6-F1 → 等待 G6-F2 Browser acceptance correction**。  
> **不是** S4 production complete，**不是** S5 ready。  
> 已知产品 UI 缺口（**不是启动失败**）：
>
> 1. Semantic Mapping **Edit & Confirm UI** 当前缺失  
> 2. Data Health Drift **fresh SchemaSnapshot mapping** 当前 UI 不完整  
>
> 这两个问题留给下一轮 **FORMA-S4-G6-F2**，不要在本地启动成功后误判为环境坏了。

---

## 1. 适用范围

- 适用：在本机查看 Forma 产品页面、做普通 CRUD、开发 Forma Frontend / Backend（基于 Coze Runtime）。
- 不适用：把 G6 acceptance harness、Playwright、一次性测试库包装成日常开发方案。
- 本仓库 Forma 代码位于 `coze-studio/`（Coze Studio fork）与 `forma/docs/`。

---

## 2. 当前架构

推荐本地拓扑（固定）：

```text
Browser
  ↓
Forma Frontend Dev Server
  http://localhost:3001
  /api/forma  → proxy
  ↓
Forma / Coze Backend（真实 opencoze main.go）
  http://localhost:8888
  ↓
MySQL / Redis / Elasticsearch / MinIO / Milvus / NSQ / etcd
（docker-compose-debug.yml --profile middleware）
```

说明：

- Frontend 包：`@forma/app`（`frontend/apps/forma`），`rsbuild dev`，端口 **3001**（本启动器要求占用失败时直接报错，不会静默改到 3002）。
- Backend：复用 Coze 正式入口（`scripts/setup/server.sh` / `main.go`）。Windows 因 milvus 原生编译限制，启动器通过 Docker 编译并运行**同一份** `opencoze`，**不是** `forma-live-harness`。
- Middleware：复用 `make middleware` 同一套 `docker/docker-compose-debug.yml --profile middleware`，不另起 Forma 专用 compose。

### 与 `make web` 的区别

| 路径 | 用途 |
|---|---|
| **Forma local launcher**（本文） | Forma 前端热更新 + 真实 Backend API，日常产品开发 |
| `make web` | Coze **完整 Docker 部署**（web+server 镜像），**不等于** Forma 热更新开发环境 |

---

## 3. 系统要求

| 组件 | 要求 |
|---|---|
| OS | Windows 11 + Docker Desktop，或 macOS / Ubuntu |
| Git | 已安装 |
| Docker + Compose | Daemon 可用 |
| Node.js | `rush.json`：`>=21`（建议 22 LTS） |
| Go | Linux/macOS 原生 Backend 需要 `go.mod` 中的版本（当前 `1.24.x`）；Windows 由 Docker 构建，可不装本机 Go |
| 磁盘 | 足够拉取 middleware 镜像与 Go 构建缓存 |
| 端口 | **3001**、**8888** 必须可用。Middleware 默认还需要 3306/6379/9200/9000/19530/4150/2379。若本机已占用 **3306**，启动器会自动启用 ports override，将 Coze MySQL 发布到 **13306**（容器内仍为 `mysql:3306`）。 |

不要指望启动脚本自动安装 Docker Desktop / Node / Go。

---

## 4. 第一次启动

### Windows（PowerShell）

若遇执行策略阻止脚本，仅对当前进程放开（不要永久 `Unrestricted`）：

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
```

```powershell
cd <repo>\coze-studio

.\scripts\forma\local\forma-local.ps1 doctor
.\scripts\forma\local\forma-local.ps1 start
```

### macOS / Linux（Bash）

```bash
cd <repo>/coze-studio

chmod +x scripts/forma/local/forma-local.sh
./scripts/forma/local/forma-local.sh doctor
./scripts/forma/local/forma-local.sh start
```

第一次 `start` 会自动：

1. 如缺失则从 `docker/.env.debug.example` 创建 `docker/.env.debug`
2. 如缺失则创建 gitignored 的 `.forma-local/.env`（Forma 可选覆盖）
3. 启动 middleware
4. 对已有 `docker/atlas/forma/migrations/**` 执行 `atlas migrate apply`（可重复、不破坏已应用 revision）
5. 启动 Backend
6. 如需要则 `rush install`，再启动 `@forma/app` dev
7. 健康检查并打印访问地址

成功输出示例：

```text
Forma Local Development READY

Frontend:
  http://localhost:3001

Backend:
  http://localhost:8888

Pages:
  http://localhost:3001/business
  http://localhost:3001/analyst
  http://localhost:3001/data
  ...
```

可选 Make 薄封装（Linux/macOS；Windows 用户不依赖 Make）：

```bash
make forma-local-doctor
make forma-local
make forma-local-status
make forma-local-stop
```

---

## 5. 日常启动

```powershell
# Windows
cd <repo>\coze-studio
.\scripts\forma\local\forma-local.ps1 start
```

```bash
# macOS / Linux
cd <repo>/coze-studio
./scripts/forma/local/forma-local.sh start
```

默认 `stop` **不会**关掉 middleware，便于快速重启。

---

## 6. 查看状态

```powershell
.\scripts\forma\local\forma-local.ps1 status
```

```bash
./scripts/forma/local/forma-local.sh status
```

关注：Docker / MySQL / Redis / Backend / Forma API / Frontend。

---

## 7. 查看日志

```powershell
.\scripts\forma\local\forma-local.ps1 logs
.\scripts\forma\local\forma-local.ps1 logs backend
.\scripts\forma\local\forma-local.ps1 logs frontend
.\scripts\forma\local\forma-local.ps1 logs middleware
```

```bash
./scripts/forma/local/forma-local.sh logs
./scripts/forma/local/forma-local.sh logs backend
./scripts/forma/local/forma-local.sh logs frontend
./scripts/forma/local/forma-local.sh logs middleware
```

日志目录：`coze-studio/.forma-local/logs/`（已 gitignore）。输出会脱敏，不会打印 API Key / Cookie / `FORMA_SECRET_MASTER_KEY` 真值。

---

## 8. 停止 / 重启

```powershell
.\scripts\forma\local\forma-local.ps1 stop
.\scripts\forma\local\forma-local.ps1 stop --all
.\scripts\forma\local\forma-local.ps1 restart
```

```bash
./scripts/forma/local/forma-local.sh stop
./scripts/forma/local/forma-local.sh stop --all
./scripts/forma/local/forma-local.sh restart
```

- `stop`：只停 Frontend / Backend；middleware 继续跑  
- `stop --all`：再 `compose down`  
- `restart`：不清数据库  

PID 安全：只杀启动器记录的 PID / 容器名 `forma-local-backend`，不会 `killall node` / `killall go`。

---

## 9. 页面访问地址

| 页面 | URL |
|---|---|
| Home / Shell | http://localhost:3001/ |
| Business | http://localhost:3001/business |
| AI Analyst | http://localhost:3001/analyst |
| Data Plane | http://localhost:3001/data |
| Requirements | http://localhost:3001/data/requirements |
| Sources | http://localhost:3001/data/sources |
| Mappings | http://localhost:3001/data/mappings |
| Contracts | http://localhost:3001/data/contracts |
| Health | http://localhost:3001/data/health |

Backend 健康：`http://localhost:8888/api/forma/v1/health`

---

## 10. 测试账号 / 首次注册说明

Forma **复用 Coze Passport Session**，仓库内**没有**写死的 `admin/admin`。

1. 确认注册未关闭（`docker/.env.debug` 中 `DISABLE_USER_REGISTRATION` 为空或非 `true`）。
2. 用任意邮箱注册（示例，自行替换邮箱与密码）：

```powershell
# Windows PowerShell
Invoke-RestMethod -Method POST -Uri http://localhost:8888/api/passport/web/email/register/v2/ `
  -ContentType 'application/json' `
  -Body '{"email":"you@example.com","password":"YourLocalPass1!"}' `
  -SessionVariable s
```

```bash
# macOS / Linux
curl -c /tmp/forma-cookies.txt -X POST http://localhost:8888/api/passport/web/email/register/v2/ \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"YourLocalPass1!"}'
```

3. 若已注册则改走 login：`/api/passport/web/email/login/`。
4. 将响应中的 `session_key` Cookie 写入浏览器对 `localhost:3001` 的 Cookie（DevTools → Application → Cookies）。
5. 打开 Forma 页面；首次进入若无 Tenant，产品会尝试 `POST /api/forma/v1/bootstrap` 创建默认 Workspace（OWNER）。

未登录时页面仍可打开，Shell 会显示「未登录」横幅——这表示前端已起来，不是启动失败。

---

## 11. AI 功能如何开启

### Mode A — UI / Product Preview（默认）

- **不需要**真实 LLM API Key。
- 系统必须能启动与浏览；AI 按钮在无模型时可提示未配置。

### Mode B — Full AI Development

Forma **不**单独配置 Provider SDK Key（不要在 Forma 启动器里塞 `OPENAI_API_KEY` / `DEEPSEEK_API_KEY` 作为 Forma 私有配置）。

按 Coze 原生方式配置模型：

1. 从 `backend/conf/model/template/` 复制模板到 `backend/conf/model/*.yaml`（该目录已 gitignore）。
2. 填写 `id`、`meta.conn_config.api_key`、`meta.conn_config.model` 等。
3. 重启 Backend（`restart`）。

模型走 Coze / Eino Model Manager；Forma Analyst / Mapping Suggest 通过既有 ACL 使用。

---

## 12. DataSource Credential 配置

Credential 加密依赖环境变量：

`FORMA_SECRET_MASTER_KEY`

规则：

- 必须能解码为 **正好 32 字节**（base64 或 hex）。
- **禁止**在仓库硬编码默认 key；example 文件不放可用于 production 的固定 secret。
- 写入 gitignored：`coze-studio/.forma-local/.env`

生成：

```powershell
.\scripts\forma\local\forma-local.ps1 gen-secret
```

```bash
./scripts/forma/local/forma-local.sh gen-secret
```

将输出的一行写入 `.forma-local/.env` 后重启 Backend。

Preview 且不用 Credential：可不配置（doctor/start 会 **WARN**）。

---

## 13. 常见问题

**Q: 3001 被占用？**  
A: 启动器会 **FAIL** 并提示。请释放 3001，不要依赖 rsbuild 自动换端口（`strictPort=false` 会导致文档与真实端口不一致）。

**Q: 8888 被占用？**  
A: 常见是残留的 `forma-live-harness` 或其他 Backend。停掉占用进程后再 `start`。正式本地方案不是 harness。

**Q: Middleware 已有部分容器 Created 未 Up？**  
A: `start` 会执行 `docker compose ... --profile middleware up -d --wait`。

**Q: Forma 表不存在？**  
A: `start` 在 Backend 前会 apply `docker/atlas/forma/migrations`。重复 start 安全。

**Q: `make web` 能代替本文？**  
A: 不能当作 Forma 热更新开发环境。见第 2 节。

**Q: 是否会自动灌 Business / Requirement 测试数据？**  
A: **不会**。普通 `start` 只启平台。

---

## 14. Windows

- 入口：`.\scripts\forma\local\forma-local.ps1`
- 需要 Docker Desktop；**不要求 WSL**。
- Backend：Docker 构建 `backend/main.go` → 容器 `forma-local-backend` 接入 `coze-studio-debug_coze-network`。
- ExecutionPolicy：见第 4 节（仅 Process Bypass）。

---

## 15. macOS / Linux

- 入口：`./scripts/forma/local/forma-local.sh`
- 若本机有 Go + bash：优先走 `scripts/setup/server.sh`（与 `make server` 同源）。
- 否则回退到与 Windows 相同的 Docker 构建路径。

脚本根据自身路径解析 `COZE_STUDIO_ROOT`，可在任意 cwd 调用。

---

## 16. 数据清理

默认 **不**删除数据。

```powershell
.\scripts\forma\local\forma-local.ps1 reset
.\scripts\forma\local\forma-local.ps1 reset --data YES
```

```bash
./scripts/forma/local/forma-local.sh reset
./scripts/forma/local/forma-local.sh reset --data YES
```

`reset --data YES` 会 compose down 并删除 `docker/data`——**危险，需显式 YES**。  
`start` **绝不会**自动 wipe DB。

---

## 17. 安全注意事项

- 不要提交 `docker/.env.debug`、`.forma-local/.env`、真实模型 YAML、真实 master key。
- `status` / `logs` / `doctor` 只显示 `FORMA_SECRET_MASTER_KEY configured = YES|NO`，不打印真值。
- 不要把生产密钥放进 example 文件。
- Acceptance / G6 harness 与 Developer Runtime 分离；日常开发请用本文 launcher。

---

## 两种本地模式速查

| 模式 | 用途 | LLM | Secret Key |
|---|---|---|---|
| A Preview | 看页面、CRUD | 不需要 | Credential 不用则可省略（WARN） |
| B Full AI | Analyst / Suggest Mapping | Coze model YAML | Credential 功能需要合法 key |

---

## 相关路径

| 路径 | 说明 |
|---|---|
| `scripts/forma/local/forma-local.ps1` | Windows 入口 |
| `scripts/forma/local/forma-local.sh` | Unix 入口 |
| `scripts/forma/local/forma-local.mjs` | 共享逻辑 |
| `scripts/forma/local/.forma-local.env.example` | Forma 可选覆盖示例 |
| `docker/.env.debug.example` | Coze debug env 基础 |
| `docker/docker-compose-debug.yml` | Middleware |
| `docker/atlas/forma/migrations/` | Forma migrations |
| `.forma-local/` | 运行时 pid/logs/env（gitignore） |
