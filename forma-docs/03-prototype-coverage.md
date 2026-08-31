# 原型覆盖范围与演示指南

## 交付性质

本项目是可运行、可点击的 React + TypeScript 高保真**前端原型**。有本地状态、跨页依赖和业务条件校验，不是生产 Agent 平台。默认数据保存在浏览器 localStorage，不连接真实模型、数据库、第三方渠道、客户系统或 Figma。

## 需求覆盖矩阵

| 需求 | 原型实现 | 生产待实现 |
|---|---|---|
| 企业 Business-to-Agent 定位 | 总览、资产链、业务场景 | 客户业务实测与商业验证 |
| 四层核心资产 | 业务/能力/Agent/应用独立页面 | 资产服务、完整多记录 Registry |
| Data Plane | 模式切换、字段校验、知识/TTL/迁移页面 | 持久数据、事务、迁移、备份与租户隔离 |
| Analyst / Pattern / Evidence | 10 Patterns、访谈保存、规则候选、双人模拟确认、Diff/Impact | 真实 LLM 抽取、证据存储、图算法、多人身份 |
| Capability | 6 类实现、契约详情、依赖、OpenAPI JSON 解析 | 真实 Adapter、SDK 包、契约执行与版本管理 |
| Agent Center | 搜索、新建、编辑、复制、软删除、恢复、版本标记、模拟测试、JSON 导入导出 | 完整历史快照、批量操作、AI 生成、API 与权限 |
| Application Builder | 多 Agent 选择、6 种模式切换、节点选择、共享字段/知识、降级配置 | 可执行图、任意拖拽连线、并行调度与状态恢复 |
| Platform Kernel | 分层说明、候选 Runtime 与 Policy 矩阵 | Runtime 执行、记忆、流、工具授权 |
| Human Task Center | 队列、详情、理由、批准/拒绝/升级、审计 | 真实人员、通知、SLA 计时、checkpoint 恢复 |
| Evaluation | 预置用例生成、Mock 选项、模拟回归、报告导出 | 自动测试生成服务、真实模型评测与回归引擎 |
| 发布控制 | 业务 Gate、冻结、逐级晋升、Canary 比例、Prod 模拟与回滚 | CI/CD、真实流量、健康观察、服务端授权 |
| Channel Gateway | 9 类渠道、身份映射、Vault 引用验证 | OAuth、平台审核、回调与验签、实际消息 |
| Runtime Adapter | 候选选择、兼容要求、未认证标识 | 各框架实现与兼容认证；明确 DeepSeek Harness |
| 可观测性 | Trace 选择、搜索、执行瀑布、成本与 KPI 示例 | 遥测采集、告警、真实计费与脱敏 |
| 企业治理 | 权限矩阵、Secret 示例、区域/Retention 演练、操作审计 | 服务端策略、真实 Vault、驻留、不可篡改审计 |
| 商业交付 | Gate 约束的 Manifest 导出、授权、部署/升级说明 | 签名 ZIP/镜像、SBOM、许可证校验、客户升级 |
| 设计工程化 | Tokens JSON、组件、16 主路由、交互状态、Figma/Cursor 方案 | 实际 Figma 文件、节点映射、Code Connect 发布 |

## 演示 A：从规则共识到交付包

1. 打开 `/business` → 证据与确认。分别点击两个“模拟确认”。生产必须是两个真实账号，原型允许一人演示。
2. 在 `/agents` 编辑“工单 Agent”，修改业务说明或能力，再保存。保存后当前评测失效。
3. 在 `/applications` 管理 Agent 组合；切换 Supervisor / Pipeline / Parallel 等模式，观察画布变化。保留至少一个 Agent。
4. 在 `/evaluation` 点击“生成业务测试”，保持 Mock 正常，运行回归。当前快照应通过。
5. 在 `/releases` 点击“冻结版本”，依次晋级 Test、Staging，然后确认模拟发布到 Prod。调整 Canary 比例。
6. 在 `/delivery` 导出 Application Package Manifest。它是本地 JSON，包含依赖与评测，不是生产可部署包。
7. 回到 `/releases` 执行回滚，稳定版本恢复为 v1.3.0。

注意：在步骤 4 之后修改规则、Agent、组合、映射、Runtime 或渠道，会清除旧评测与冻结，需要重新回归。

## 演示 B：失败与保护

- 未完成双人确认时运行回归，Gate 被阻断。
- Mock 选择“注入契约破坏”，回归失败。切回正常后重跑。
- 尝试删除被当前应用引用的 Agent，系统拒绝并给出解除引用的路径。
- 创建 Agent 不填写名称/Role/Capability，保存校验失败。
- 导入未知 Capability ID 或非 JSON 文件，显示错误，不新增资产。
- 尝试为应用移除全部 Agent，发布 Gate 被阻断。
- 修改冻结后的资产，冻结被撤销，发布回到 Dev。

## 演示 C：人工与运营

- `/applications` → 冲突与降级 → 模拟超时与人工接管，进入 `/human`。
- 输入理由后批准/拒绝或升级；切换“全部任务”查看结果。
- `/observability` 选择不同 Trace，查看完成、等待审批、超时降级的说明。
- `/governance` → 审计日志，查看原型操作记录；可导出 JSON。
- 用“重置演示数据”恢复初始状态，会清除本地编辑，已下载文件不受影响。

## 工程限制

16 个一级路由可以直接访问，子视图用页内 Tabs；当前不提供每个 Tab 的独立 URL。业务模型与应用重点展示一个园区场景；总览的其他应用与部分统计为展示数据。Node 点击用于配置选择，不支持自由拖拽。版本标记与差异是演示数据，不是通用版本引擎。

部分辅助治理配置仅生成演练记录，页面有明确提示；未提供真实数据删除、部署、授权计费或 Secret 轮换。暂停、错误与成功的展示用于说明产品行为，不能作为真实系统验收证据。

浏览器视觉、点击自动化和响应式截图验收尚未执行；本轮验证范围为 TypeScript、生产构建、路由 HTTP 响应及关键领域规则测试。建议下一阶段用真实浏览器按上述三个脚本完成用户验收。
