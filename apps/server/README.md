# Platform API

本目录是 OpenClaw 平台控制面的 Go 后端。

当前已包含：

- 实例、配置、备份、任务、告警、审计
- 渠道接入中心
- 工单与购买
- Keycloak / OIDC 接入骨架
- OpenSearch 检索骨架
- Kubernetes Runtime Adapter 骨架
- OEM / 品牌配置域
- `DATABASE_URL` 真实持久化启动，实例核心域会写入 PostgreSQL

`cmd/server` 现在只支持真实后端运行形态：

- `persistent`：必须配置 `DATABASE_URL`，服务从 PostgreSQL 加载并持久化核心业务状态
- `PLATFORM_STRICT_MODE=true`：默认开启，要求数据库、真实 runtime provider 和已启用的第三方配置全部就绪

注意：

- `cmd/server` 已移除 `seeded` 启动路径，数据库是唯一数据源
- 如果数据库尚未执行 bootstrap，列表接口会返回空集合，实例创建也会因为缺少套餐或集群基础数据而失败
- 如需初始化数据库，请显式执行 `cmd/bootstrap`；`cmd/server` 不再支持 `AUTO_BOOTSTRAP=true`
- `strict` 模式下，服务会在启动时拒绝以下配置：无 `DATABASE_URL`、无可用 runtime provider、`KEYCLOAK_MOCK_*`、`WECHAT_LOGIN_MOCK_*`
- `strict` 模式下，服务启动时还会校验数据库连通、runtime provider 可达；如启用了 OpenSearch / Keycloak，还会校验对应服务可达
- 搜索、微信登录、支付等能力不再回退到内存或 mock fallback；未配置就直接报错
- 平台侧工作台会话已支持冻结版龙虾桥接协议：`openclaw-lobster-bridge/v2`
- 平台侧消息桥接、历史补偿、异步回调、事件流均可通过 `WORKSPACE_BRIDGE_*` 配置
- 平台侧产物正式预览已支持：`ARTIFACT_PREVIEW_*` + `OBJECT_STORAGE_*`

## Artifact Preview

正式预览策略：

- `web` / `html`：走平台 `artifact preview gateway`，并在独立 iframe sandbox 中隔离加载
- `pdf`：走平台代理直出
- `pptx` / `docx` / `xlsx`：优先使用 `previewUrl` 指向的可信衍生预览文件；推荐 PDF，也支持实例侧生成的 HTML / 图片 / 文本衍生物；缺失时自动回退到下载
- 已归档到对象存储的产物优先从归档副本提供预览与下载

生产要求：

- `PLATFORM_STRICT_MODE=true` 时必须配置 `ARTIFACT_PREVIEW_PUBLIC_BASE_URL`
- `ARTIFACT_PREVIEW_ALLOWED_HOSTS` 只允许填写平台信任的对象存储 / CDN / 工作台域名
- `ARTIFACT_PREVIEW_ALLOW_PRIVATE_IP` 只允许在本地联调时打开
- 如启用归档副本，必须配置完整的 `OBJECT_STORAGE_*`

## Run

```bash
go run ./cmd/server
```

如需只做数据库引导：

```bash
DATABASE_URL=postgresql://... BOOTSTRAP_DATA_PATH=./bootstrap-data.json go run ./cmd/bootstrap
```

如需查看 migration 状态：

```bash
DATABASE_URL=postgresql://... go run ./cmd/migrationstatus
```

服务进程默认不会自动执行 migration / bootstrap。
如需显式执行 migration：

```bash
AUTO_MIGRATE=true
go run ./cmd/server
```

如需在服务启动时把对外 OpenAPI 和接入说明额外写到某个目录：

```bash
EXTERNAL_DOCS_OUTPUT_DIR=./external-docs
go run ./cmd/server
```

默认端口：

```text
http://localhost:8080
```

## Key Endpoints

- `GET /healthz`
- `GET /versionz`
- `GET /api/v1/docs/external`
- `GET /api/v1/docs/external/openapi.yaml`
- `GET /api/v1/docs/external/integration.md`
- `GET /api/v1/bootstrap`
- `GET /api/v1/auth/config`
- `GET /api/v1/search/logs`
- `GET /api/v1/runtime/clusters`
- `GET /api/v1/oem/config`
- `POST /api/v1/platform/workspace/report`
- `GET /api/v1/portal/workspace/sessions/:id/events`
- `GET /api/v1/portal/workspace/artifacts/:id/preview`
- `GET /api/v1/portal/workspace/artifacts/:id/preview-content`
- `GET /api/v1/portal/workspace/artifacts/:id/download`

## Env

参考：

- `apps/server/.env.example`
- `.temp/dev-stack/README.md`
- `docs/龙虾工作台桥接协议.md`
- `apps/server/internal/httpapi/externaldocs/openapi.yaml`
- `apps/server/internal/httpapi/externaldocs/integration-guide.md`
- `docs/runbooks/2026-04-09-platform-api-production-config-matrix.md`

