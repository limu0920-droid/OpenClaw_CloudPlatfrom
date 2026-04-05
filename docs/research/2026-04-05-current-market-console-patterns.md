# OpenClaw 当前市面对标补充

## 研究问题

- 在 2026 年 4 月这个时间点，主流云控制台、开发者平台和 AI Gateway 正在强化哪些控制台模式？
- 这些模式里，哪些适合直接补到当前 OpenClaw Portal/Admin 原型中？

## 结论先行

- 当前市面产品的共性已经非常明确：
  - 一级视图不仅展示资源数，还展示 `usage / spend / requests / logs / alerts`
  - 视图会明确区分 `team / project / environment` 或等价的作用域
  - 列表页之外，几乎都会提供实例或服务级详情页，详情页里再下钻监控、日志、配置、审计
  - 观测与控制正在合并，不再把监控、日志、成本分散在完全独立的子系统里
- 基于这些趋势，本轮已把 OpenClaw 原型补成：
  - `Portal` 首页新增“护栏与策略”信息
  - `Admin` 新增实例详情页，集中展示资源趋势、配置、访问入口、任务、告警、审计

## 市场事实

### 1. Vercel AI Gateway

来源：

- https://vercel.com/docs/ai-gateway/capabilities/observability

事实：

- 2026 年 2 月更新的文档明确把 `Usage`、`Requests` 放进 AI Gateway Overview。
- 指标支持 `Team level` 与 `Project level` 两种作用域。
- 使用侧重点已经扩展到：
  - 请求量
  - 模型分布
  - TTFT
  - 输入输出 token
  - Spend
  - 按项目与 API Key 分组的请求摘要
  - 可排序/导出的详细请求日志

推断：

- 这说明 AI 平台控制台的一级页面，不该只看“实例在不在”，而要看“谁在用、花了多少、请求是否健康”。

### 2. Railway Observability Dashboard

来源：

- https://docs.railway.com/observability
- https://docs.railway.com/guides/metrics
- https://docs.railway.com/guides/logs

事实：

- Railway 在 2026 年 3 月的官方文档中，把 Observability Dashboard 定义为一个可自定义的统一仪表板。
- 仪表板支持：
  - Metrics
  - Logs
  - Project Usage
- 仪表板明确按 `project environment` 作用域组织。
- 服务级 Metrics 强调：
  - CPU
  - Memory
  - Disk
  - Network
  - 多副本 Sum / Replica 视图
- Logs 既支持服务级，也支持跨服务 Log Explorer。

推断：

- 对 OpenClaw 来说，这验证了两点：
  - `Admin` 需要实例详情页，而不是只保留实例列表
  - 详情页里必须把资源趋势、任务和日志入口摆在一起

### 3. Vercel Logs / Activity / Audit

来源：

- https://vercel.com/docs/logs

事实：

- Vercel 官方文档把日志分成：
  - Build Logs
  - Runtime Logs
  - Activity Logs
  - Audit Logs
  - Log Drains
- Runtime logs 支持搜索、检查、分享。
- Audit logs 支持导出。

推断：

- OpenClaw 也不应该把“运行日志”和“操作审计”混成一张表。
- 目前原型已保留 Portal 日志页和 Admin 审计页分离，方向正确，后续只需要继续细化。

### 4. Supabase Reports

来源：

- https://supabase.com/docs/guides/telemetry/reports

事实：

- Supabase 当前把 Reports 做成项目级观测面，覆盖：
  - Database
  - Auth
  - Storage
  - Realtime
  - API systems
- 文档强调图表不仅用于“看数据”，而是用于 `self-debugging` 和优化。
- API Gateway 报表中重点关注：
  - Total Requests
  - Response Errors
  - Response Speed
  - Network Traffic

推断：

- 这说明 Portal 和 Admin 都应该包含“护栏型指标”，而不是只罗列业务对象。
- 本轮 Portal 首页补的“护栏与策略”区域，就是往这个方向走的第一步。

### 5. Cloudflare AI Gateway

来源：

- https://developers.cloudflare.com/ai-gateway/

事实：

- Cloudflare 把 AI Gateway 定义为“Observe and control your AI applications”。
- 页面把能力直接并列为：
  - Analytics
  - Logging
  - Caching
  - Rate limiting
  - Retry / Fallback

推断：

- AI 平台控制台已经不只是“看板”，而是“观测 + 控制 + 护栏”的一体化界面。
- OpenClaw 下一阶段应该把备份策略、配置护栏、限流、重试、回退等能力逐步变成平台级控制项。

### 6. 阿里云 ECS 实例监控

来源：

- https://help.aliyun.com/zh/ecs/user-guide/view-the-monitoring-information-of-an-ecs-instance

事实：

- 阿里云官方文档继续强调实例详情页中的监控页签，重点包括：
  - vCPU 使用率
  - 网络流量
  - 内存使用率
  - 时间范围切换
- 文档明确让用户从“实例列表 -> 实例详情 -> 监控页签”继续下钻。

推断：

- 对平台管理员来说，“实例详情页 + 资源趋势”依然是最稳妥、最通用的交互结构。
- 所以本轮把 Admin 实例详情页补出来，而不是继续往列表里堆字段。

## 已落地到当前原型的改动

### 已实现

- `Portal` 首页新增护栏视角
  - 体现作用域、风险提醒、恢复目标
- `Admin` 新增实例详情页
  - 资源趋势
  - 配置摘要
  - 访问入口
  - 任务与备份
  - 告警与审计
- `Portal` 核心操作已可用
  - 创建实例
  - 发布配置
  - 触发备份

### 暂未实现

- Team / Project / Environment 真实切换
- 日志分享与导出
- Spend / token / TTFT 等 AI 使用计量
- 多副本 Replica 级监控
- 真实 Log Drains / Audit Export

## 下一步建议

1. 给 `Admin` 增加真实的实例详情二级导航：
   - 概览
   - 监控
   - 日志
   - 配置
   - 审计
2. 给 `Portal` 增加 usage / cost 原型页
3. 把当前内存 Mock API 切到持久化数据库
4. 增加实例级日志和备份恢复流程
