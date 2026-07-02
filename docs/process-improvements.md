# 项目流程梳理与改进方案

## 当前流程概览

项目是前后端分离架构：

- 前端：Vue 3 + Vite + TDesign，提供后台管理、项目、剧本、分镜、角色、素材和视频创作界面。
- 后端：Gin + GORM，提供 `/admin/v1` 管理 API。
- 异步任务：Asynq + Redis，负责剧本生成、角色/场景/道具提取、生图、视频生成、视频合并。
- 存储：MySQL 保存业务数据，`server/uploads` 保存上传与生成资源，FFmpeg 负责本地视频处理。

核心创作链路：

1. 创建短剧项目。
2. 创建或 AI 生成剧本。
3. 从剧本中提取角色、场景、道具。
4. 生成角色图、场景图、道具图。
5. 拆分分镜并生成帧提示词。
6. 生成分镜图片和视频。
7. 合并片段导出成片。

## 优先级 P0：先让部署和运行稳定

1. 拆分 API 与 Worker 进程
   - 状态：已完成。
   - 已新增 `worker` 子命令，只启动 Asynq 消费者。
   - Compose 已拆成 `server` 和 `worker` 两个服务，API 扩容不会重复消费任务。

2. 配置命名统一
   - 状态：已完成。
   - 已统一保留 `REDIS_ASYNC_DB`、`REDIS_MAIN_DB`、`REDIS_CACHE_DB`。
   - Asynq client/server 现在读取同一个 `redis.database_async` 配置。

3. 健康检查分层
   - 状态：已完成。
   - `/healthz` 用于基础存活检查。
   - `/readyz` 检查 MySQL、Redis、FFmpeg、`uploads` 和 `storage/logs` 可写性。

4. 密钥与敏感文件治理
   - `.env` 不入库，仅保留 `.env.example`。
   - 对已经泄漏过的 token 立即吊销并轮换。
   - 增加 pre-commit secret scan，例如 gitleaks。

## 优先级 P1：提升任务可观测性与可恢复性

1. 任务状态标准化
   - 状态：已完成基础版本。
   - `async_tasks` 已增加 `status_name`，统一为 `pending`、`running`、`succeeded`、`failed`、`cancelled`。
   - 每个任务记录 progress、payload、result、error、retry_count、started_at、finished_at。

2. 幂等与重试
   - 状态：已完成基础幂等。
   - `async_tasks` 已增加 `idempotency_key`，相同任务在 `pending/running` 状态下会复用原任务，避免重复点击造成重复投递。
   - 重试策略仍沿用 Asynq 当前配置，后续可按任务类型细化。

3. 任务日志
   - 状态：已完成基础版本。
   - 已新增 `async_task_events`，记录 queued、started、progress、succeeded、failed、cancelled。
   - `GET /admin/v1/tasks/:id` 返回任务和最近 20 条事件。

4. 任务取消
   - 状态：已完成基础版本。
   - 已新增 `POST /admin/v1/tasks/:id/cancel`。
   - 长轮询视频任务会感知状态变化；短任务取消后主要用于前端展示和后续调度控制。

5. 资源生成结果统一
   - 统一图片、视频、音频资源表，记录来源任务、URL、尺寸、时长、hash、provider。
   - 上传和 AI 生成资源走同一套管理逻辑。

## 优先级 P2：完善创作流程体验

1. 工作台状态机
   - 项目、章节、分镜分别维护清晰阶段，前端按阶段引导下一步。
   - 示例：剧本已就绪、角色待确认、分镜待生成、视频待合并。

2. 人工确认点
   - 角色/场景/道具提取后增加确认步骤，再批量生图。
   - 分镜生成后允许锁定关键镜头，后续重跑不覆盖已确认内容。

3. AI 配置按用途选择
   - 文本、生图、视频、提示词优化分别绑定默认 provider/model。
   - 在任务提交时保存当时使用的配置快照，便于复现结果。

4. 成本与限流
   - 每次任务预估调用次数和成本。
   - 给用户、项目、provider 配置并发和日限额。

## 优先级 P3：工程质量与协作

1. CI
   - 后端：`go test ./...`、`go vet ./...`。
   - 前端：`npm run build:type`、`npm run lint`、`npm run build`。
   - Docker：构建 server/web 镜像并做 smoke test。

2. API 契约
   - 补齐 Swagger 注解或引入 OpenAPI 文件。
   - 前端 API 类型从 OpenAPI 生成，减少字段漂移。

3. 数据库迁移
   - AutoMigrate 适合开发和初期部署。
   - 生产建议引入版本化迁移工具，所有 DDL 可审计、可回滚。

4. 测试分层
   - Service 层单元测试覆盖核心业务规则。
   - Controller 层集成测试覆盖登录、项目、任务提交。
   - Worker 用 fake provider 覆盖成功、失败、重试场景。

## 建议实施顺序

1. 完成 Docker 部署和 README 对齐。
2. 新增 `worker` 子命令并在 Compose 中拆分服务。
3. 增加 `/readyz` 与启动配置校验。
4. 统一异步任务状态表和任务日志。
5. 完善任务详情前端展示和取消按钮。
6. 完善前端工作台阶段引导与人工确认点。
7. 建立 CI、OpenAPI 和版本化迁移。
