# OpenClaw Platform

本仓库用于承接 OpenClaw 平台化研发调研与第一阶段 Web 端开发。

当前目录结构：

- `docs/research`：调研结论
- `docs/plans`：实施计划
- `apps/web`：Vue Web 前端
- `apps/mock-api`：Go Mock API

第一阶段目标：

- Portal：实例、访问入口、配置、备份、任务
- Admin：租户、实例、任务、告警、审计
- Mock API：提供前后端联调用的统一 HTTP 数据接口

当前已可用流程：

- `Portal / 实例列表`
  - 创建实例
  - 搜索实例
  - 跳转实例详情
- `Portal / 渠道接入中心`
  - 查看渠道列表
  - 一键发起连接
  - 查看渠道详情
  - 触发健康检查
  - 断开连接
- `Portal / 实例详情`
  - 查看访问入口
  - 查看 CPU / 内存 / 磁盘 / API 请求量 / Token 用量
  - 查看管理员账号与密码摘要
  - 开启 / 关闭 / 重启实例
  - 查看套餐并发起购买 / 续费
  - 发布配置
  - 触发备份
  - 查看最近任务与备份记录
- `Portal / 任务中心`
  - 查看任务执行状态
- `Admin`
  - 查看总览、租户、实例、任务、告警、审计
  - 从实例列表下钻到实例详情
  - 在实例详情中切换 `概览 / 监控 / 日志 / 配置 / 审计`
  - 查看渠道列表与渠道详情
  - 查看并处理工单

- `Keycloak / OpenSearch 骨架`
  - 登录页可读取 Keycloak 配置并生成登录地址
  - Portal / Admin 的日志与审计页支持统一搜索
  - 后端已提供 `auth/*` 与 `search/*` 接口

当前实现边界：

- Go Mock API 为进程内内存数据，重启服务后会重置
- 创建实例、配置发布、备份触发会即时成功，不模拟真实异步队列与审批
- 第三方聊天平台接入目前是 Mock 连接器，不包含真实 OAuth 回调、二维码扫码、Webhook 验签与消息收发
- Keycloak 目前是接入骨架，不包含真实 code -> token 交换
- OpenSearch 目前支持代理骨架与 mock fallback，未默认连接真实集群
- 未接入真实鉴权、数据库、Kubernetes、对象存储、支付与监控系统

常用命令：

```bash
cd apps/web
npm install
npm run dev:web
npm run build:web

cd apps/mock-api
go run ./cmd/server
go test ./...
```

联调方式：

1. 先在 `apps/mock-api` 启动 Go Mock API，默认监听 `http://localhost:8080`
2. 再在 `apps/web` 启动 `npm run dev`
3. Vite 已代理 `/api` 到 `http://localhost:8080`

如果前端需要显式指定 API 地址，可在 `apps/web/.env.example` 基础上创建本地环境变量文件。

后续开发缺口文档：

- [2026-04-05-openclaw-web-follow-up.md](C:/Users/Administrator/Desktop/openclaw调研/docs/plans/2026-04-05-openclaw-web-follow-up.md)
- [2026-04-05-channel-follow-up.md](C:/Users/Administrator/Desktop/openclaw调研/docs/plans/2026-04-05-channel-follow-up.md)
- [2026-04-05-oss-acceleration-stack.md](C:/Users/Administrator/Desktop/openclaw调研/docs/plans/2026-04-05-oss-acceleration-stack.md)
