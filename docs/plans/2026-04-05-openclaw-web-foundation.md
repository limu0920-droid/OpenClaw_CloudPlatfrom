# OpenClaw Web Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 搭建 OpenClaw 平台第一阶段 Web 端基础工程，包含 Vue 双布局前端和 Go Mock API，并跑通 portal/admin 的核心主流程。

**Architecture:** 采用单仓结构，前端使用一个 Vue 应用承载 `/portal` 与 `/admin` 两套布局，复用组件、类型与状态枚举；后端使用 Go 输出统一 `/api/v1` Mock 数据接口，前端通过 HTTP 取数而不是直接写死页面数据。

**Tech Stack:** Vue 3, TypeScript, Vite, Vue Router, Pinia, Go 1.25, CSS Variables

---

### Task 1: Project Skeleton

**Files:**
- Create: `apps/web/*`
- Create: `apps/mock-api/*`
- Create: `package.json`
- Create: `README.md`

**Step 1: Create the frontend scaffold**

- 目录：
  - `apps/web/src/app`
  - `apps/web/src/components`
  - `apps/web/src/features`
  - `apps/web/src/lib`
  - `apps/web/src/router`
  - `apps/web/src/styles`

**Step 2: Create the mock API scaffold**

- 目录：
  - `apps/mock-api/cmd/server`
  - `apps/mock-api/internal/httpapi`
  - `apps/mock-api/internal/mockdata`
  - `apps/mock-api/internal/models`

**Step 3: Add workspace-level scripts**

Run:

```bash
npm run build --workspace apps/web
go test ./...
```

Expected:

- 前端可完成类型检查与构建
- Go 服务可通过基础编译与测试

**Step 4: Commit**

```bash
git add package.json README.md apps/web apps/mock-api
git commit -m "chore: scaffold openclaw web foundation"
```

### Task 2: Shared Domain Model And Mock Contracts

**Files:**
- Create: `apps/web/src/lib/types.ts`
- Create: `apps/web/src/lib/enums.ts`
- Create: `apps/web/src/lib/api.ts`
- Create: `apps/mock-api/internal/models/models.go`
- Create: `apps/mock-api/internal/mockdata/data.go`

**Step 1: Define shared enums**

- `InstanceStatus`
- `JobStatus`
- `BackupStatus`
- `AlertSeverity`
- `UserRole`

**Step 2: Define DTOs around the first-phase entities**

- `tenant`
- `user`
- `serviceInstance`
- `instanceAccess`
- `instanceConfig`
- `backupRecord`
- `operationJob`
- `alertRecord`
- `clusterSummary`

**Step 3: Implement mock repositories**

- 使用固定内存数据输出：
  - portal 概览
  - admin 概览
  - 实例列表
  - 实例详情
  - 任务列表
  - 告警列表

**Step 4: Verify**

Run:

```bash
go test ./apps/mock-api/...
npm run build --workspace apps/web
```

Expected:

- 类型与 JSON 结构一致
- 前端服务层能消费 Mock 响应

**Step 5: Commit**

```bash
git add apps/web/src/lib apps/mock-api/internal
git commit -m "feat: define openclaw mock domain contracts"
```

### Task 3: Portal Experience

**Files:**
- Create: `apps/web/src/features/portal/*`
- Create: `apps/web/src/components/portal/*`
- Modify: `apps/web/src/router/index.ts`
- Modify: `apps/web/src/app/App.vue`

**Step 1: Build the portal shell**

- 左侧导航 + 顶部租户栏
- 页面区支持响应式卡片布局

**Step 2: Build the key portal pages**

- `/portal/overview`
- `/portal/instances`
- `/portal/instances/:id`
- `/portal/jobs`
- `/portal/logs`

**Step 3: Connect data**

- 页面必须通过 `api.ts` 拉取数据
- 首页使用概览卡、最近任务、告警摘要、快捷入口
- 实例详情页使用 tab 组织“概览 / 入口 / 配置 / 备份”

**Step 4: Verify**

Run:

```bash
npm run build --workspace apps/web
```

Expected:

- Portal 路由可构建
- 无类型错误

**Step 5: Commit**

```bash
git add apps/web/src/features/portal apps/web/src/components/portal apps/web/src/router apps/web/src/app
git commit -m "feat: add openclaw portal experience"
```

### Task 4: Admin Experience

**Files:**
- Create: `apps/web/src/features/admin/*`
- Create: `apps/web/src/components/admin/*`
- Modify: `apps/web/src/router/index.ts`

**Step 1: Build the admin shell**

- 深色控制台外壳
- 固定侧边栏 + 顶部全局状态栏

**Step 2: Build the key admin pages**

- `/admin/overview`
- `/admin/tenants`
- `/admin/instances`
- `/admin/jobs`
- `/admin/alerts`
- `/admin/audit`

**Step 3: Build the step-flow cards**

- 用于表达任务、恢复、配置发布等状态流
- 采用可复用组件，不为单页面硬编码

**Step 4: Verify**

Run:

```bash
npm run build --workspace apps/web
```

Expected:

- Admin 路由与深色主题可正常构建
- Portal 与 Admin 主题共存不冲突

**Step 5: Commit**

```bash
git add apps/web/src/features/admin apps/web/src/components/admin apps/web/src/router
git commit -m "feat: add openclaw admin console"
```

### Task 5: Go Mock API And Frontend Integration

**Files:**
- Create: `apps/mock-api/cmd/server/main.go`
- Create: `apps/mock-api/internal/httpapi/router.go`
- Create: `apps/mock-api/internal/httpapi/handlers.go`
- Modify: `apps/web/vite.config.ts`
- Modify: `apps/web/src/lib/api.ts`
- Modify: `README.md`

**Step 1: Expose first-phase endpoints**

- `GET /api/v1/portal/overview`
- `GET /api/v1/portal/instances`
- `GET /api/v1/portal/instances/:id`
- `GET /api/v1/portal/jobs`
- `GET /api/v1/portal/logs`
- `GET /api/v1/admin/overview`
- `GET /api/v1/admin/tenants`
- `GET /api/v1/admin/instances`
- `GET /api/v1/admin/jobs`
- `GET /api/v1/admin/alerts`
- `GET /api/v1/admin/audit`

**Step 2: Configure frontend dev access**

- Vite 开发时代理到 Go 服务
- 生产构建允许读取 `VITE_API_BASE_URL`

**Step 3: End-to-end verification**

Run:

```bash
npm run build --workspace apps/web
go test ./apps/mock-api/...
```

Expected:

- Go 服务启动后，前端页面可用真实 Mock HTTP 取数
- 首批页面可展示完整信息架构与状态流

**Step 4: Commit**

```bash
git add apps/mock-api apps/web/vite.config.ts apps/web/src/lib/api.ts README.md
git commit -m "feat: wire vue frontend to go mock api"
```

Plan complete and saved to `docs/plans/2026-04-05-openclaw-web-foundation.md`. User already requested multi-agent execution in this session, so implementation proceeds with subagent-driven execution and review.
