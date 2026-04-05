# Keycloak + OpenSearch Follow-up Plan

## 当前已完成

### Keycloak

- 后端已提供：
  - `GET /api/v1/auth/config`
  - `GET /api/v1/auth/session`
  - `GET /api/v1/auth/keycloak/url`
- 前端登录页已可：
  - 展示 Keycloak 配置状态
  - 展示 mock/current user
  - 生成 Keycloak 登录跳转入口

### OpenSearch

- 后端已提供：
  - `GET /api/v1/search/config`
  - `GET /api/v1/search/logs`
- 当前搜索支持两种模式：
  - 配置了 OpenSearch 时尝试走 `_search`
  - 未配置时走内存 mock fallback
- Portal 日志页与 Admin 审计页已切到统一搜索入口

## 当前仍未实现

### Keycloak 真正接入

- OIDC code -> token 交换
- Access Token / Refresh Token 持久化
- Logout / silent refresh
- 用户角色与租户映射
- Keycloak Admin API 同步用户 / 角色 / 组

### OpenSearch 真正接入

- 索引模板
- 审计 / 渠道 / 工单 / 运行日志索引拆分
- ingest pipeline
- 查询 DSL 的更完整过滤项
- 分页、排序、聚合
- Dashboard / 可视化面板

## 推荐下一步

1. 先把 Keycloak 的 callback 与 token exchange 补齐
2. 再把 OpenSearch 的 index mapping 与真实写入链路补齐
3. 然后把 Portal/Admin 的日志搜索页接上真实字段过滤与分页
