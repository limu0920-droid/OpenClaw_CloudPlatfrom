# OpenClaw

OpenClaw 当前已经统一为一套前后端平台仓库：

- Web 前端入口：`apps/web`
- 唯一后端入口：`apps/server`

旧的 `apps/mock-api` 兼容目录已经移除；当前 `cmd/server` 也只允许真实数据库和真实 runtime provider 启动。

## 目录结构

- `apps/web`：Vue 3 + Vite 的 Portal / Admin / Workspace 前端
- `apps/server`：Go 平台后端，包含 API、持久化、桥接、预览、权限与运行时适配
- `deploy/k8s/platform-api`：Kubernetes 清单、发布脚本、预检与回滚脚本
- `docs/research`：调研记录
- `docs/plans`：实施计划
- `docs/runbooks`：上线、发布、回滚、运维手册
- `.temp/dev-stack`：本地联调栈与 smoke 脚本
- `perf/platform-api`：k6 压测 smoke 与基线记录

## 本地开发

后端：

```powershell
cd apps/server
go run ./cmd/server
```

前端：

```powershell
cd apps/web
npm install
npm run dev
```

默认本地代理关系：

```text
/api -> http://localhost:8080
```

## 验证命令

后端测试：

```powershell
cd apps/server
go test ./...
```

前端构建：

```powershell
cd apps/web
npm run build
```

生产清单预检：

先按 [deploy/k8s/platform-api/README.md](deploy/k8s/platform-api/README.md) 执行 `set-image.ps1` / `set-config.ps1` 注入目标环境配置；模板态 `prod` overlay 在 `-Strict` 下会按预期拦截占位符。

```powershell
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Strict
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Bootstrap -Strict
```

## 关键文档

- 平台后端说明：`apps/server/README.md`
- Web 前端说明：`apps/web/README.md`
- K8s 发布说明：`deploy/k8s/platform-api/README.md`
- 生产配置矩阵：`docs/runbooks/2026-04-09-platform-api-production-config-matrix.md`
- 对外 OpenAPI：`apps/server/internal/httpapi/externaldocs/openapi.yaml`
- 对外接入说明：`apps/server/internal/httpapi/externaldocs/integration-guide.md`
- 工作台桥接协议：`docs/龙虾工作台桥接协议.md`
- 上线阻塞项：`上线阻塞项.md`
- 已有能力清单：`已有功能.md`
- 待实现能力清单：`待实现功能.md`
