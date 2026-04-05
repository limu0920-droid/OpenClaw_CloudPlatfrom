# OpenClaw 开发加速开源栈

## 直接集成候选

### 1. Keycloak

- 用途：身份、SSO、OAuth2、OIDC、租户/角色体系
- 许可证：`Apache-2.0`
- 建议接法：
  - 平台统一身份层
  - Portal/Admin 登录
  - 后续第三方应用 OAuth / 企业 SSO

### 2. Headlamp

- 用途：Kubernetes 资源可视化与控制台嵌入
- 许可证：`Apache-2.0`
- 建议接法：
  - 管理员侧集群 / Namespace / Pod / Deployment 浏览
  - 不直接暴露 Headlamp 全部能力给普通用户
  - 以平台后端受控代理或嵌入式视图方式接入

### 3. Prometheus + Alertmanager

- 用途：实例 CPU / 内存 / API 调用量 / 告警
- 许可证：`Apache-2.0`
- 建议接法：
  - 采集 OpenClaw 实例与平台控制面指标
  - Portal 展示用户自己的资源数据
  - Admin 展示全局资源与告警

### 4. OpenSearch

- 用途：日志检索、审计查询、工单上下文搜索
- 许可证：`Apache-2.0`
- 建议接法：
  - 平台日志与审计索引
  - 渠道活动检索
  - 工单排障上下文串联

### 5. Chatwoot Community

- 用途：问题上报 / 工单 / 多渠道客服
- 许可证：`MIT`
- 建议接法：
  - 一期可只参考信息架构与 API
  - 二期可考虑把工单中心与渠道中心进一步联动

### 6. Appsmith Community

- 用途：内部运营台 / 快速搭后台
- 许可证：`Apache-2.0`
- 建议接法：
  - 用于内部运营工具，不直接替代外部门户
  - 可加速财务/客服/审计等内部模块

## 只参考不建议一期直接接入

### 1. Kubernetes Dashboard

- 原因：官方仓库已归档，不建议新项目作为主控制面

### 2. Zammad

- 原因：功能强，但 `AGPLv3`
- 建议：只有在你确定愿意承担 AGPL 合规要求时再接

### 3. ZITADEL v3+

- 原因：`AGPLv3`
- 建议：当前项目优先使用 `Keycloak`

### 4. Grafana

- 原因：核心项目已转 `AGPLv3`
- 建议：如无强依赖，优先以 `Prometheus + OpenSearch + 自研前端图表` 起步

## 推荐对接顺序

1. `Keycloak`
2. `Prometheus + Alertmanager`
3. `Headlamp`
4. `OpenSearch`
5. `Chatwoot` 或自研工单中心二选一
6. `Appsmith` 作为内部辅助工具

## 与当前项目的映射

- 已做原型：
  - 实例资源、启停、购买、工单、渠道接入中心
- 下一步建议先做真实接入：
  - `Keycloak` 登录
  - `Prometheus` 实例指标
  - `Headlamp` 管理员集群视图
  - `OpenSearch` 日志/审计查询
