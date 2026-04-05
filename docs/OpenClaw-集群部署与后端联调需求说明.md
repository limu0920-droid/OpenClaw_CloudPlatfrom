# OpenClaw 集群部署与后端联调需求说明

## 1. 文档目标

本文档用于明确 OpenClaw 在生产环境中的部署资源基线、集群化改造方向、多用户支持要求，以及与后端系统联调所需的接口与运维能力。本文档面向运维、后端、平台工程和实施人员。

## 2. 背景与范围

当前计划以 Kubernetes 作为生产环境主部署方式，Docker 仅用于本地验证、开发自测或兼容性排查，并满足以下目标：

- 单副本可稳定运行。
- 支持通过 Kubernetes 实现集群部署和横向扩缩。
- 支持多个用户同时使用。
- 需要与后端系统联调，提供 Web 访问获取、配置修改、自动创建、自动删除、备份设置等能力。

本文档覆盖部署与联调需求，不覆盖具体业务流程设计、计费模型和前端交互细节。

## 3. 部署资源要求

### 3.1 单副本最低部署要求

以下资源作为当前 Kubernetes 单副本运行基线：

- CPU Request：`500m`
- CPU Limit：`1 vCPU`
- 内存 Request：`1 GiB`
- 内存 Limit：`1.5 GiB`
- 临时存储：`10 GiB`

说明：

- `1.5 GiB` 内存上限可作为单副本基础运行配置，用于轻量使用场景。
- Kubernetes 生产集群通常不应将 Swap 作为容量设计前提，资源规划应以 `requests/limits` 和节点可分配资源为准。
- 该配置适合验证环境、小规模试运行或低并发场景。
- 不应直接将该配置视为多用户生产场景的充分配置。

### 3.2 单副本建议运行配置

为减少 OOM、Pod 频繁重启和高峰期响应不稳定，建议单副本生产配置不低于：

- CPU Request：`1 vCPU`
- CPU Limit：`2 vCPU`
- 内存 Request：`2 GiB`
- 内存 Limit：`4 GiB`
- 临时存储：`20 GiB` 以上
- 持久卷：`20 GiB` 以上

如需叠加浏览器访问、自动化调用、日志保留、备份生成或其他附加任务，建议提升到：

- CPU Limit：`4 vCPU`
- 内存 Limit：`6 GiB`
- 持久卷：按备份保留周期单独评估

## 4. 集群部署要求

### 4.1 总体原则

OpenClaw 集群部署应按无状态应用思路设计，避免将关键状态写入单个容器本地文件系统。集群节点可横向扩容，但用户状态、配置、审计和备份元数据应统一托管。

### 4.2 集群架构建议

建议部署形态如下：

- 入口层：使用 Ingress Controller 或 Gateway API 统一暴露访问入口。
- 应用层：OpenClaw 以 `Deployment` 多副本方式运行；如后续引入强状态组件，再单独评估 `StatefulSet`。
- 配置层：通过 `ConfigMap`、`Secret` 统一下发配置和敏感信息。
- 状态层：将配置、用户状态、任务元数据等持久化到外部存储。
- 弹性层：通过 `HPA`、Pod 反亲和或拓扑分布策略提升可用性。
- 备份层：使用独立备份存储，避免与运行容器强耦合。
- 监控层：接入日志、指标和告警系统。

### 4.3 集群能力要求

集群方案至少应满足以下要求：

- 支持 `Deployment` 多副本部署、滚动更新和版本回滚。
- 支持 `startupProbe`、`livenessProbe`、`readinessProbe`，以及异常实例自动摘除。
- 支持通过 `ConfigMap`、`Secret` 统一下发配置，避免每个节点手工修改。
- 支持基于 `Service + Ingress` 或 `Gateway API` 的统一域名访问和负载均衡。
- 支持按副本数或 `HPA` 自动扩缩容，并具备基础的副本分散能力。
- 支持日志、指标和事件集中采集。
- 支持 Pod 重建后业务配置和备份数据不丢失。

### 4.4 集群资源建议

若目标是给多个用户使用，建议从以下配置起步：

- OpenClaw 工作负载单副本建议：`requests: 1 vCPU / 2 GiB`，`limits: 2 vCPU / 4 GiB`
- 初始副本数建议：`2` 个
- 入口层建议：`1` 组负载均衡或网关
- `HPA` 建议起步：`minReplicas=2`、`maxReplicas=4`
- `PDB` 建议起步：`minAvailable=1`
- 备份存储建议：独立持久卷或对象存储

说明：

- 若仅按 `1.5 GiB` 为每个副本配置，在多用户、后台任务或集中操作时，内存余量偏小。
- 集群场景不建议简单复制最低配置，应按照并发量、操作频率和附加能力单独评估。

## 5. 多用户使用要求

面向多个用户提供服务时，系统应具备以下基础能力：

- 用户访问隔离，避免不同用户互相影响。
- 支持统一认证接入，或至少预留认证集成能力。
- 支持用户级或租户级权限控制。
- 支持操作审计，记录配置变更、创建、删除、备份等关键动作。
- 支持资源限流，避免单用户占满实例资源。

如后续存在租户隔离要求，建议优先评估以下两种模式：

- 单集群共享实例，基于逻辑权限隔离。
- 每个租户独立实例，由平台统一调度与管理。

若涉及高隔离、高审计或高可定制性，优先考虑“租户独立实例”模式。

## 6. 后端联调需求

### 6.1 联调目标

后端系统需要能够统一接管 OpenClaw 的访问地址管理、配置变更和生命周期操作，减少人工运维。

### 6.2 必需联调能力

后端联调至少需要覆盖以下能力：

- 获取实例 Web 访问地址。
- 获取实例运行状态。
- 获取当前生效配置。
- 修改实例配置。
- 创建新实例。
- 删除已有实例。
- 触发备份。
- 查询备份列表。
- 配置备份策略。
- 执行备份恢复。

### 6.3 建议接口能力清单

建议为平台层或运维服务层统一抽象以下接口能力：

- `createInstance`
- `deleteInstance`
- `getInstanceStatus`
- `getInstanceAccessInfo`
- `getInstanceConfig`
- `updateInstanceConfig`
- `backupInstance`
- `listBackups`
- `restoreBackup`
- `updateBackupPolicy`

如 OpenClaw 本身未直接提供完整管理接口，应在外层增加“平台控制服务”进行封装，由该服务负责调用 Kubernetes API、受控 Runtime Agent、配置中心和存储系统。

### 6.4 联调边界建议

为降低后端与底层容器平台的耦合度，建议联调边界如下：

- 业务后端不直接操作 Kubernetes API、容器运行时或集群凭证。
- 由平台服务对外暴露统一管理 API。
- 平台服务负责实例编排、配置发布、备份执行和状态回传。
- 后端系统只处理业务参数、权限校验和调用编排结果。

## 7. 配置管理要求

为支持集群和自动化运维，配置管理应满足以下要求：

- 配置应支持集中管理，不依赖手工进入容器修改。
- 配置变更应支持版本记录。
- 配置变更应支持审计追踪。
- 配置变更后应明确是否需要热加载或滚动重启。
- 敏感信息应与普通配置分离管理。

建议将配置拆分为以下几类：

- 基础运行配置
- 外部服务连接配置
- 用户访问与认证配置
- 备份与存储配置
- 日志与监控配置

## 8. 备份与恢复要求

### 8.1 备份范围

备份至少应覆盖以下内容：

- 实例配置
- 业务数据或运行状态数据
- 关键审计信息
- 与用户使用直接相关的持久化内容

### 8.2 备份能力

备份能力至少应满足：

- 支持手动触发备份。
- 支持定时备份。
- 支持查询备份记录。
- 支持按备份点恢复。
- 支持备份保留周期设置。
- 支持备份失败告警。

### 8.3 备份实现建议

建议采用以下策略：

- 备份文件写入独立持久存储，不写入临时容器层。
- 备份策略由平台统一维护。
- 恢复操作需具备权限校验和审计记录。
- 恢复前建议提供校验与预检查能力。

## 9. 运维与可观测性要求

生产部署至少需要具备以下运维能力：

- 健康检查
- 容器重启策略
- 统一日志采集
- 关键指标监控
- 告警通知
- 审计日志留存

建议重点监控以下指标：

- 内存使用率
- CPU 使用率
- 容器磁盘使用率
- 容器文件系统剩余空间
- 网络流入流出速率
- 容器重启次数
- OOM 次数
- 请求响应时间
- 错误率
- 备份成功率
- 配置变更成功率

建议图表至少支持以下统计维度：

- 按租户查看
- 按用户查看
- 按实例查看
- 按容器或 Pod 查看
- 按节点查看
- 按集群查看

## 10. 安全要求

为支撑多用户和平台联调，建议满足以下安全要求：

- 管理接口必须鉴权。
- 配置修改、删除、恢复等高风险动作必须审计。
- 敏感配置应加密存储或使用密钥管理服务。
- Web 访问入口应支持 HTTPS。
- 不允许将高权限容器管理能力直接暴露给业务侧。

## 11. 风险与限制

当前已知风险如下：

- `1.5 GiB RAM` 适合作为单实例低负载基线，不适合作为多用户集群统一标准。
- 集群部署后，内存压力不仅来自 OpenClaw 本体，还来自日志、代理、备份、调度和自动化任务。
- 若后续引入浏览器自动化、外部插件或更多后台任务，单实例资源需求会进一步上升。
- 若缺少统一配置中心和平台控制层，后端联调会直接耦合容器平台，维护成本较高。

## 12. 建议实施顺序

建议按以下阶段推进：

1. 先完成单命名空间或单租户的 Kubernetes 基础部署，按 `requests: 1 vCPU / 2 GiB`、`limits: 2 vCPU / 4 GiB` 验证可运行性。
2. 增加平台控制接口，打通创建、删除、查询、配置修改和备份操作。
3. 完成配置集中化和备份存储改造。
4. 引入双副本、`HPA`、`PDB` 和滚动发布策略，验证集群可用性。
5. 接入认证、审计、监控和告警，具备对外多用户服务能力。

## 13. 当前资源结论

当前可以将 `1.5 GiB` 内存上限视为 OpenClaw 单副本低负载验证的最低运行参考；若以 Kubernetes 作为生产主路径，建议以 `requests: 1 vCPU / 2 GiB`、`limits: 2 vCPU / 4 GiB` 作为单副本的规划起点，并在平台层补齐统一访问、配置管理、备份恢复、弹性扩缩和生命周期控制能力。

## 14. 前端需求补充

### 14.1 设计原则

前端不应只复用 OpenClaw 原生 Control UI，而应分成两层：

- OpenClaw 原生控制台层：用于直接管理网关、通道、技能、日志和调试能力。
- 平台门户层：用于面向最终用户、租户管理员、平台管理员和财务人员提供统一的实例、配置、计费和运维入口。

建议前端至少拆分为以下两个站点：

- 用户侧门户：面向租户管理员和普通用户。
- 平台管理后台：面向运维、客服、财务和平台管理员。

### 14.2 OpenClaw 原生能力基线

根据 OpenClaw 官方 Control UI 文档，当前原生控制台已覆盖以下能力，可作为平台前端设计的最低功能基线：

- Chat 对话与工具调用过程展示。
- 渠道状态查看、二维码登录、渠道级配置修改。
- 实例在线状态查看。
- 会话列表及会话级参数调整。
- Cron 任务列表、执行、启停和历史查看。
- Skills 启用、禁用、安装和密钥更新。
- 节点列表与节点能力查看。
- Exec 审批策略和白名单维护。
- 配置查看、修改、校验、应用和重启。
- 健康状态、模型快照、事件日志和手工 RPC 调试。
- 实时日志查看与导出。
- 包更新与重启。

说明：

- 以上能力适合作为“运维控制面”能力输入。
- 面向商业化、多租户和支付场景时，仍需要额外建设平台门户。

### 14.3 用户侧前端需求

参考 OpenClaw Control UI 能力，以及阿里云云市场和计算巢常见的 SaaS 控制台模式，用户侧门户建议提供以下模块：

- 登录与认证：支持账号密码、单点登录或企业身份接入。
- 服务工作台：展示当前租户已开通的实例、状态、到期时间、套餐和告警。
- 已购服务列表：参考阿里云“已购买的服务”模式，支持查看实例状态、进入详情、进入应用、升级、续费、转正。
- 实例详情页：展示实例名称、访问地址、后台地址、地域、版本、规格、到期时间、备份状态、最近操作记录。
- 访问入口页：提供前台访问地址、后台访问地址、免登跳转或一次性访问令牌。
- 配置中心：提供表单化配置编辑、只读配置查看、版本对比、回滚入口。
- 备份中心：支持创建备份、查看备份记录、恢复实例、查看恢复进度。
- 任务中心：展示创建、删除、重启、配置变更、备份恢复等异步任务进度。
- 账单与订单中心：展示订单、支付状态、订阅状态、续费记录、退款记录和发票状态。
- 通知中心：展示支付结果、实例异常、备份失败、到期提醒和版本升级提醒。
- 操作审计页：展示租户维度的关键操作日志。

### 14.4 平台管理后台需求

平台管理后台建议提供以下能力：

- 租户管理：创建租户、停用租户、套餐绑定、额度调整。
- 用户详情中心：查看用户基础信息、登录情况、所属租户、角色、订单、订阅和最近操作。
- 用户与角色管理：用户、角色、权限、菜单和数据范围控制。
- 集群与资源池管理：查看节点、集群健康、可用容量、实例分布和调度策略。
- 实例管理：按租户、地域、集群、状态统一检索实例，并支持批量运维。
- 用户实例映射视图：查看“用户 -> 实例 -> 集群 -> 节点 -> 容器 / Pod”的关联关系。
- 容器诊断中心：查看容器 ID、镜像版本、所在节点、启动时间、重启次数和最近事件。
- 远程终端：在权限校验、审批和审计通过后，允许管理员进入目标容器执行诊断命令。
- 资源图表中心：查看 CPU、内存、硬盘、网络、重启、OOM 等时间序列图表。
- 配置模板管理：为不同套餐、地域和渠道维护默认配置模板。
- 产品与套餐管理：维护产品、套餐、规格、售价、试用规则、升级规则和续费策略。
- 订单与支付管理：查看订单、支付单、退款单、对账结果和异常回调。
- 备份策略管理：按套餐或租户设置默认备份周期、保留天数和恢复权限。
- 运维审批中心：审批高风险动作，例如删除实例、恢复备份、变更生产配置。
- 工单与客服支持：查看用户问题、系统异常和人工处理记录。
- 监控与告警总览：查看实例健康、容量水位、支付异常和失败任务。

### 14.5 前端信息架构建议

建议前端菜单结构至少包含以下一级模块：

- 概览
- 用户中心
- 实例管理
- 容器与诊断
- 资源监控
- 访问入口
- 配置管理
- 任务中心
- 备份与恢复
- 用户与权限
- 订单与支付
- 审计与日志
- 告警与通知
- 平台设置

其中面向普通用户的菜单可以隐藏“平台设置”“集群管理”“套餐管理”等后台能力。

## 15. 后端需求补充

### 15.1 后端总体分层

后端建议采用控制平面与业务平面分离的模式：

- 控制平面：负责实例生命周期、配置管理、备份恢复、调度编排、支付和审计。
- 业务平面：负责 OpenClaw 实际运行、用户访问和网关处理。

### 15.2 核心后端服务拆分

建议后端至少拆分为以下服务或模块：

- API Gateway / BFF：统一对前端提供 API，并按角色聚合数据。
- IAM 服务：负责登录、认证、授权、租户隔离和访问令牌管理。
- 租户与用户服务：管理租户、用户、组织、角色、套餐关系。
- 实例编排服务：负责创建、启动、停止、重启、扩缩容、删除实例。
- 配置管理服务：负责配置模板、配置版本、灰度发布、变更审计和回滚。
- 访问接入服务：负责生成访问地址、免登跳转、域名绑定、SSL 和访问策略。
- 备份恢复服务：负责备份计划、备份执行、恢复流程和对象存储交互。
- 任务调度服务：负责异步任务、重试、超时控制和任务状态回传。
- 诊断接入服务：负责容器定位、远程终端、命令白名单、会话录制和退出回收。
- 监控聚合服务：负责整合容器、节点、实例和租户维度的监控指标，供前端图表调用。
- 计费与订单服务：负责产品、套餐、订单、订阅、账单和发票。
- 支付服务：负责支付下单、回调验签、退款和对账。
- 通知服务：负责短信、邮件、站内信和 Webhook。
- 审计服务：负责关键操作日志、登录日志和风控事件记录。

### 15.3 后端必须支持的业务流程

后端应至少支持以下完整流程：

- 试用开通流程：用户下单试用后自动创建实例，并在试用到期时提醒、转正或释放。
- 正式购买流程：支付成功后自动创建实例、写入访问地址、发送通知。
- 配置变更流程：用户提交配置修改后进入校验、审批、执行、结果回传流程。
- 备份恢复流程：用户发起备份或恢复后，系统异步执行并记录快照和审计日志。
- 升级续费流程：支持升级套餐、变更规格、续费和自动续费。
- 删除释放流程：删除前执行权限校验、风险确认、可选备份和资源释放。
- 到期治理流程：到期提醒、宽限期、停服、数据保留和最终删除。

### 15.4 后端对外集成要求

后端需要与以下外部系统或基础设施联动：

- Kubernetes API 或受控 Runtime Agent
- 对象存储
- 关系型数据库
- Redis 或等价缓存
- 时序监控系统，例如 Prometheus、VictoriaMetrics 或云监控
- 日志检索系统，例如 Loki、Elasticsearch 或云日志服务
- 域名与证书服务
- 短信、邮件、Webhook 通知通道
- 支付渠道，例如支付宝、微信支付或 Stripe
- 监控与日志平台

### 15.5 高风险控制要求

后端需要对以下动作做额外控制：

- 删除实例
- 恢复备份
- 容器进入和远程命令执行
- 修改支付状态
- 调整租户套餐
- 生产环境配置发布
- 高权限免登访问

这些动作建议统一要求：

- 权限校验
- 二次确认
- 审批流或双人复核
- 审计日志
- 幂等控制

### 15.6 远程容器访问设计建议

如果管理员端需要进入用户对应 Pod 或查看容器诊断信息，建议统一抽象为“远程诊断会话”，而不是在前端直接暴露 Kubernetes 权限。

建议实现方式如下：

- 前端只调用平台后端的诊断接口，不直接持有 `kubeconfig`、`ServiceAccount Token` 或集群管理权限。
- 后端根据实例信息定位目标 `Namespace`、`Pod` 和 `Container`。
- 后端生成一次性诊断会话，并绑定操作人、实例、容器、开始时间和审批单号。
- 默认提供只读诊断能力，例如日志、环境信息、磁盘占用、进程列表和最近事件。
- 高权限命令执行需要额外审批，必要时限制为命令白名单。
- 终端输入输出应全量录制，并在会话结束后归档。

如后续需要兼容本地 Docker 环境，可在 `runtime-adapter` 内部将 `kubectl exec` 切换为 `docker exec` 或等价实现，前端交互与业务 API 不需要变化。

## 16. 数据库设计建议

### 16.1 数据存储选型

建议采用以下存储组合：

- 主数据库：`PostgreSQL`
- 缓存与分布式锁：`Redis`
- 备份文件与导出文件：对象存储
- 时序监控：`Prometheus`、`VictoriaMetrics` 或同类系统
- 日志检索：`Loki`、`Elasticsearch` 或同类系统

说明：

- 选择 PostgreSQL 的原因是关系建模稳定，并且可利用 `JSONB` 保存部分动态配置。
- 如果团队已有 MySQL 体系，也可以保留相同模型迁移到 MySQL，但动态配置和审计检索会相对受限。

### 16.2 业务域划分

数据库表建议按以下业务域划分：

- 身份与租户域
- 产品与套餐域
- 订单与支付域
- 实例与配置域
- 备份与恢复域
- 任务与审计域

### 16.3 核心表设计

建议至少包含以下核心表：

| 表名 | 作用 | 关键字段 |
| --- | --- | --- |
| `tenant` | 租户主表 | `id`, `name`, `status`, `plan_id`, `expired_at` |
| `user_account` | 用户账号 | `id`, `tenant_id`, `login_name`, `email`, `phone`, `status` |
| `app_role` | 角色定义 | `id`, `code`, `name`, `scope` |
| `user_role_rel` | 用户角色关系 | `user_id`, `role_id`, `tenant_id` |
| `product` | 产品定义 | `id`, `code`, `name`, `type`, `status` |
| `service_plan` | 套餐定义 | `id`, `product_id`, `code`, `name`, `resource_spec`, `trial_supported` |
| `plan_price` | 套餐价格 | `id`, `plan_id`, `billing_cycle`, `currency`, `amount`, `status` |
| `order_main` | 订单主表 | `id`, `tenant_id`, `order_no`, `source_platform`, `external_order_no`, `order_type`, `status`, `total_amount` |
| `order_item` | 订单明细 | `id`, `order_id`, `plan_id`, `quantity`, `unit_price`, `period_start`, `period_end` |
| `subscription` | 订阅关系 | `id`, `tenant_id`, `product_id`, `plan_id`, `status`, `renew_mode`, `expired_at` |
| `payment_transaction` | 支付流水 | `id`, `order_id`, `channel`, `trade_no`, `channel_order_no`, `amount`, `status`, `paid_at` |
| `refund_record` | 退款记录 | `id`, `order_id`, `payment_id`, `refund_no`, `amount`, `status` |
| `invoice_record` | 发票记录 | `id`, `tenant_id`, `order_id`, `invoice_type`, `status`, `amount` |
| `cluster` | 集群信息 | `id`, `name`, `region`, `status` |
| `cluster_node` | 节点信息 | `id`, `cluster_id`, `hostname`, `status`, `capacity_cpu`, `capacity_memory_mb` |
| `service_instance` | OpenClaw 实例主表 | `id`, `tenant_id`, `cluster_id`, `instance_code`, `status`, `plan_id`, `version` |
| `runtime_container` | 实例运行容器映射 | `id`, `instance_id`, `runtime_type`, `container_ref`, `node_id`, `status`, `resource_limit_json` |
| `instance_member` | 用户与实例关系 | `id`, `instance_id`, `user_id`, `member_role`, `status` |
| `instance_access` | 访问入口 | `id`, `instance_id`, `entry_type`, `url`, `domain`, `is_primary` |
| `instance_config` | 当前配置快照 | `instance_id`, `config_version`, `config_json`, `config_hash` |
| `instance_config_history` | 配置历史 | `id`, `instance_id`, `version`, `config_json`, `changed_by`, `changed_at` |
| `backup_record` | 备份记录 | `id`, `instance_id`, `backup_no`, `storage_uri`, `status`, `expired_at` |
| `restore_record` | 恢复记录 | `id`, `instance_id`, `backup_id`, `status`, `started_at`, `finished_at` |
| `operation_job` | 异步任务 | `id`, `job_type`, `target_type`, `target_id`, `status`, `payload_json` |
| `operation_log` | 业务操作日志 | `id`, `tenant_id`, `operator_id`, `action`, `target_type`, `target_id`, `result` |
| `terminal_session` | 远程终端会话审计 | `id`, `instance_id`, `container_ref`, `operator_id`, `approved_by`, `status`, `started_at`, `ended_at` |
| `approval_record` | 审批记录 | `id`, `approval_no`, `approval_type`, `target_type`, `target_id`, `applicant_id`, `status` |
| `notification_record` | 通知记录 | `id`, `tenant_id`, `notify_type`, `channel`, `status`, `template_code`, `sent_at` |
| `alert_record` | 告警事件 | `id`, `tenant_id`, `instance_id`, `metric_key`, `severity`, `status`, `triggered_at` |
| `audit_event` | 审计事件 | `id`, `tenant_id`, `event_type`, `risk_level`, `content_json`, `created_at` |

### 16.4 关键设计原则

数据库设计建议遵循以下原则：

- 所有业务主表都带 `tenant_id`，支持多租户隔离。
- 金额统一使用最小货币单位整数存储，例如“分”。
- 订单、支付、退款流水一旦生成，不允许直接覆盖更新，只允许追加状态流转。
- 配置采用“当前快照 + 历史版本”双表设计，便于回滚和审计。
- 任务表与业务表分离，避免长事务影响主业务写入。
- 删除类操作优先采用软删除或状态删除，便于审计和恢复。
- 所有外部回调都必须使用幂等键。
- 敏感字段不明文存储，密钥和令牌应存放在密钥管理系统。
- 高频监控指标不写入主业务库，图表数据应来自时序监控系统。
- 远程终端记录至少保存会话元数据，命令全文和终端录屏可归档到对象存储。

### 16.5 索引与约束建议

建议重点增加以下索引或唯一约束：

- `order_main.order_no` 唯一索引
- `payment_transaction.trade_no` 唯一索引
- `subscription(tenant_id, product_id)` 组合索引
- `service_instance.instance_code` 唯一索引
- `service_instance(tenant_id, status)` 组合索引
- `backup_record(instance_id, created_at)` 组合索引
- `operation_job(status, created_at)` 组合索引
- `audit_event(tenant_id, created_at)` 组合索引

## 17. 支付设计建议

### 17.1 支付模式建议

如果该平台后续面向多个客户提供收费服务，建议至少支持以下计费模式：

- 免费试用
- 包月订阅
- 包年订阅
- 手动续费
- 自动续费
- 套餐升级
- 套餐降级
- 退款

如后续要按调用量、消息量、工具调用次数或存储量收费，可在后续版本增加按量计费模块，不建议在第一阶段同时引入。

### 17.2 推荐支付流程

推荐支付链路如下：

1. 用户选择产品、套餐和周期，生成待支付订单。
2. 支付服务向支付渠道创建交易单，并返回支付链接、二维码或收银台地址。
3. 用户支付成功后，支付渠道回调平台。
4. 平台完成回调验签、幂等校验和订单状态更新。
5. 订单状态变为已支付后，实例编排服务开始创建或升级实例。
6. 实例可用后，系统写入访问地址并通知用户。

### 17.3 续费、升级和到期策略

建议明确以下规则：

- 续费：支持到期前手动续费和自动续费。
- 升级：支持补差价立即生效，或下一计费周期生效，两种模式二选一。
- 降级：建议默认在下一计费周期生效，避免运行中缩容风险。
- 到期：进入宽限期后限制新配置变更与高风险操作。
- 停服：到期后可先停用访问，再保留数据一段时间。
- 删除：超过保留期后执行最终删除。

### 17.4 支付渠道与对账要求

若面向中国大陆客户，建议优先考虑：

- 支付宝
- 微信支付

若面向海外客户，可补充：

- Stripe

支付系统至少应具备以下能力：

- 支付回调验签
- 支付回调幂等
- 主动查询补单
- 日对账
- 退款
- 发票状态跟踪
- 支付失败和异常回调告警

### 17.5 支付与实例联动要求

支付状态不能只停留在财务层面，需要与实例状态联动：

- `待支付`：不创建正式实例，可选保留待开通记录。
- `已支付`：触发实例创建或套餐生效。
- `退款中`：限制高风险操作。
- `已退款`：根据业务规则执行停服、降级或释放。
- `已到期`：限制访问、进入宽限期或暂停服务。

### 17.6 商业化产品形态建议

结合阿里云计算巢和云市场常见模式，可优先设计以下商业化动作：

- 试用转正
- 升级套餐
- 续费
- 查看已购服务
- 查看应用前台地址和后台地址
- 免登跳转

说明：

- 这些能力并非 OpenClaw 原生提供，而是平台化售卖时常见的上层产品能力。
- 这里是基于阿里云公开文档抽象出的产品模式，不等于必须完全复制其交互样式。

### 17.7 第三方市场接入预留

如果后续计划接入阿里云云市场或其他第三方市场，支付和开通链路需要额外预留以下能力：

- 支持 `source_platform` 区分自有门户订单和第三方市场订单。
- 支持保存第三方订单号、实例号、授权码和回调请求号。
- 支持通过 SPI、Webhook 或授权码方式触发实例开通、续费、升级和停用。
- 支持外部平台回调后的幂等处理、审计和补偿任务。
- 支持平台对账时按外部订单号和内部实例号双向检索。

说明：

- 根据阿里云云市场公开文档，SaaS 商品在购买、续费等动作完成后，可通过 SPI 接口主动通知服务商执行实例生产或续费动作。
- 因此，支付系统不应只支持“前端下单 -> 自有支付回调”这一条链路，还应支持“第三方平台订单 -> 平台回调 -> 实例开通”这类渠道订单模式。

## 18. 建议的实施阶段

为控制复杂度，建议按三个阶段推进：

### 18.1 第一阶段

目标：

- 打通单实例和集群基础部署
- 完成实例创建、删除、配置修改、备份恢复
- 建成用户门户的实例管理和访问入口

本阶段可暂缓：

- 在线支付
- 自动续费
- 发票
- 复杂审批流

### 18.2 第二阶段

目标：

- 完成订单、订阅和支付闭环
- 支持试用转正、续费、升级
- 建成租户、角色、套餐和审计体系

### 18.3 第三阶段

目标：

- 完成多集群调度
- 接入更强的风控、审批和对账能力
- 支持多区域、多支付渠道和更复杂的计费模型

## 19. 参考依据

以下内容用于支撑本次文档中的产品和平台设计判断，访问日期为 `2026-03-17`：

- OpenClaw Control UI 官方文档：<https://docs.openclaw.ai/web/control-ui>
- OpenClaw Control UI 文档镜像页：<https://docs.openclaw.ai/control-ui>
- OpenClaw Kubernetes 部署文档：<https://docs.openclaw.ai/install/kubernetes>
- OpenClaw Docker 安装文档（本地验证参考）：<https://docs.openclaw.ai/install/docker>
- Kubernetes 资源管理文档：<https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/>
- 阿里云云市场“使用SaaS商品”：<https://help.aliyun.com/zh/marketplace/user-guide/use-saas-products/>
- 阿里云计算巢“软件服务的全生命周期管理”：<https://help.aliyun.com/zh/compute-nest/product-overview/product-mode>
- 阿里云计算巢“前期准备”：<https://help.aliyun.com/zh/compute-nest/getting-started/preparation>
- 阿里云计算巢 “SaaS Boost”：<https://help.aliyun.com/zh/compute-nest/deploy-and-use-saas-boost>
- Docker Engine 资源限制文档（兼容模式参考）：<https://docs.docker.com/engine/containers/resource_constraints/>

## 20. 总体结论

如果该项目的目标只是本地验证或单机自测，可继续使用 Docker 方式快速跑通；如果目标是做成面向多个用户的商业化平台，则应以 Kubernetes 作为生产基线，并按“OpenClaw 运行面 + 平台控制面 + 订单支付面”三层来设计。

面向平台化落地时，关键点不在于单纯复制 OpenClaw 原生界面，而在于补齐以下能力：

- 用户门户与平台管理后台分层
- 实例生命周期自动化
- 多租户、权限与审计
- 配置版本化和备份恢复
- 订单、订阅、支付、续费和退款闭环

建议后续优先按“实例管理先行、支付后补”的顺序推进，先把控制面和集群能力做稳，再接商业化闭环。

## 21. 页面原型清单

### 21.1 用户侧门户页面

建议用户侧门户至少包含以下页面：

| 页面 | 目标 | 关键模块 |
| --- | --- | --- |
| 登录页 | 用户认证与进入系统 | 账号密码、验证码、SSO 入口、租户选择 |
| 工作台 | 查看总体状态 | 实例数量、即将到期、待处理任务、告警、最近订单 |
| 已购服务列表 | 查看已开通实例 | 实例卡片、状态、版本、套餐、到期时间、快捷操作 |
| 实例详情页 | 查看单实例完整信息 | 基本信息、访问地址、运行状态、套餐信息、最近任务 |
| 配置管理页 | 修改实例配置 | 配置表单、版本对比、变更记录、应用配置按钮 |
| 备份与恢复页 | 管理备份 | 备份列表、创建备份、恢复入口、保留策略 |
| 任务中心 | 跟踪异步任务 | 创建、重启、配置发布、恢复、删除等任务状态 |
| 订单与订阅页 | 管理支付与订阅 | 订单列表、续费、升级、退款状态、发票申请 |
| 通知中心 | 接收关键提醒 | 到期提醒、支付成功、任务失败、备份异常 |
| 审计记录页 | 查询操作历史 | 谁在什么时候修改了什么 |

### 21.2 管理员端页面

建议管理员后台至少包含以下页面：

| 页面 | 目标 | 关键模块 |
| --- | --- | --- |
| 平台总览 | 观察平台运行情况 | 租户数、实例数、支付成功率、告警数、资源水位 |
| 租户列表 | 管理租户 | 状态筛选、套餐、到期时间、实例数量、冻结操作 |
| 租户详情 | 查看租户完整画像 | 基本信息、用户列表、实例列表、订单订阅、审计记录 |
| 用户列表 | 管理用户 | 用户状态、登录时间、所属租户、角色、搜索筛选 |
| 用户详情 | 查看单用户信息 | 登录记录、角色、订阅、关联实例、最近操作 |
| 实例列表 | 管理全部实例 | 状态、集群、节点、套餐、租户、版本、批量操作 |
| 实例详情 | 进行单实例运维 | 基本信息、访问入口、配置、备份、任务、告警、审计 |
| 容器诊断中心 | 查看容器状态 | 容器 ID、镜像、节点、进程、重启、日志、事件 |
| 远程终端页 | 进入诊断会话 | 审批状态、命令终端、会话录制、自动断开 |
| 资源监控页 | 查看图表 | CPU、内存、硬盘、网络、OOM、节点容量趋势 |
| 集群与节点页 | 查看资源池 | 集群状态、节点列表、可用容量、调度分布 |
| 套餐与价格页 | 配置商业规则 | 套餐、规格、试用策略、售价、周期、续费规则 |
| 订单与支付页 | 管理财务动作 | 订单、支付单、退款单、对账结果、回调异常 |
| 告警中心 | 处理异常 | 告警列表、确认、关闭、关联实例、通知记录 |
| 审批中心 | 管控高风险动作 | 删除实例、恢复备份、终端进入、配置发布审批 |

### 21.3 核心详情页布局建议

为减少页面数量并提升运维效率，建议“实例详情页”采用标签页布局：

- `概览`：实例编码、租户、状态、版本、套餐、所在集群、节点、启动时间
- `访问入口`：前台地址、后台地址、域名、证书状态、免登链接
- `配置`：当前配置、草稿配置、配置版本、差异对比、发布记录
- `任务`：最近异步任务、执行耗时、失败原因、重试入口
- `备份`：备份记录、恢复入口、保留时间、恢复结果
- `监控`：CPU、内存、硬盘、网络、重启、OOM 时间序列图
- `容器`：容器或 Pod 列表、容器 ID、镜像、重启次数、最近事件
- `日志`：运行日志、错误日志、搜索与下载
- `审计`：实例维度的配置修改、终端进入、删除、恢复等操作日志

建议“用户详情页”至少包含以下标签页：

- `基本信息`
- `登录记录`
- `角色与权限`
- `订单与订阅`
- `关联实例`
- `操作审计`

### 21.4 资源图表清单

管理员端建议至少提供以下图表：

| 图表 | 统计对象 | 用途 |
| --- | --- | --- |
| CPU 使用率趋势 | 实例 / 容器 / 节点 | 判断是否存在持续高负载 |
| 内存使用率趋势 | 实例 / 容器 / 节点 | 判断是否接近 OOM |
| 磁盘已用空间趋势 | 实例 / 节点 | 判断日志或备份是否挤占空间 |
| 网络流入流出趋势 | 实例 / 节点 | 判断是否存在异常流量 |
| 容器重启次数趋势 | 实例 / 容器 | 判断运行稳定性 |
| OOM 事件趋势 | 实例 / 租户 | 判断内存配置是否不足 |
| 活跃实例数趋势 | 租户 / 平台 | 判断业务使用规模 |
| 创建与删除任务趋势 | 平台 | 判断资源变动频率 |
| 支付成功率趋势 | 平台 | 判断支付链路稳定性 |

### 21.5 管理员角色建议

建议至少拆分以下后台角色：

- 平台超级管理员：拥有全部权限
- 运维管理员：管理实例、配置、监控、终端与备份
- 客服管理员：查看租户、用户、订单和基础实例信息
- 财务管理员：查看订单、支付、退款、发票与对账
- 审计管理员：查看操作日志、审批记录和终端会话

## 22. 后端 API 草案

### 22.1 API 设计原则

建议后端 API 至少按以下边界拆分：

- `/api/portal/*`：用户门户接口
- `/api/admin/*`：管理员后台接口
- `/api/internal/*`：平台内部服务接口
- `/api/callback/*`：支付、第三方市场和异步回调接口

建议统一返回结构包含：

- `code`
- `message`
- `data`
- `requestId`

对于异步操作，建议统一返回：

- `jobId`
- `status`
- `estimatedSeconds`

### 22.2 认证与用户接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/portal/auth/login` | 用户登录 |
| `POST` | `/api/portal/auth/logout` | 用户退出 |
| `GET` | `/api/portal/auth/me` | 获取当前用户信息 |
| `GET` | `/api/admin/users` | 管理员查询用户列表 |
| `GET` | `/api/admin/users/{userId}` | 管理员查看用户详情 |
| `GET` | `/api/admin/users/{userId}/instances` | 查看用户关联实例 |
| `GET` | `/api/admin/users/{userId}/audit-events` | 查看用户操作记录 |
| `PATCH` | `/api/admin/users/{userId}/status` | 冻结或启用用户 |

### 22.3 租户与套餐接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/admin/tenants` | 查询租户列表 |
| `POST` | `/api/admin/tenants` | 创建租户 |
| `GET` | `/api/admin/tenants/{tenantId}` | 查看租户详情 |
| `PATCH` | `/api/admin/tenants/{tenantId}` | 更新租户资料 |
| `PATCH` | `/api/admin/tenants/{tenantId}/status` | 冻结或恢复租户 |
| `GET` | `/api/admin/plans` | 查询套餐列表 |
| `POST` | `/api/admin/plans` | 创建套餐 |
| `PATCH` | `/api/admin/plans/{planId}` | 更新套餐 |
| `GET` | `/api/admin/products` | 查询产品列表 |

### 22.4 实例生命周期接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/portal/instances` | 用户查询自己的实例 |
| `POST` | `/api/portal/instances` | 用户申请开通实例 |
| `GET` | `/api/portal/instances/{instanceId}` | 用户查看实例详情 |
| `GET` | `/api/portal/instances/{instanceId}/accesses` | 获取实例访问地址 |
| `POST` | `/api/admin/instances` | 管理员创建实例 |
| `GET` | `/api/admin/instances` | 管理员查询实例列表 |
| `GET` | `/api/admin/instances/{instanceId}` | 管理员查看实例详情 |
| `POST` | `/api/admin/instances/{instanceId}/start` | 启动实例 |
| `POST` | `/api/admin/instances/{instanceId}/stop` | 停止实例 |
| `POST` | `/api/admin/instances/{instanceId}/restart` | 重启实例 |
| `DELETE` | `/api/admin/instances/{instanceId}` | 删除实例 |
| `POST` | `/api/admin/instances/{instanceId}/scale` | 调整实例规格 |

### 22.5 配置与备份接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/admin/instances/{instanceId}/config` | 查看当前配置 |
| `PUT` | `/api/admin/instances/{instanceId}/config` | 提交配置变更 |
| `GET` | `/api/admin/instances/{instanceId}/config/history` | 查看配置历史 |
| `POST` | `/api/admin/instances/{instanceId}/config/validate` | 校验配置 |
| `POST` | `/api/admin/instances/{instanceId}/config/publish` | 发布配置 |
| `POST` | `/api/admin/instances/{instanceId}/backups` | 创建备份 |
| `GET` | `/api/admin/instances/{instanceId}/backups` | 查询备份列表 |
| `POST` | `/api/admin/instances/{instanceId}/backups/{backupId}/restore` | 执行恢复 |
| `PATCH` | `/api/admin/instances/{instanceId}/backup-policy` | 修改备份策略 |

### 22.6 容器诊断与远程终端接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/admin/instances/{instanceId}/containers` | 查询实例对应容器或 Pod |
| `GET` | `/api/admin/containers/{containerId}` | 查看容器详情 |
| `GET` | `/api/admin/containers/{containerId}/logs` | 查询容器日志 |
| `GET` | `/api/admin/containers/{containerId}/events` | 查询容器事件 |
| `POST` | `/api/admin/containers/{containerId}/terminal-sessions` | 创建远程终端会话 |
| `GET` | `/api/admin/terminal-sessions/{sessionId}` | 查询终端会话详情 |
| `POST` | `/api/admin/terminal-sessions/{sessionId}/close` | 主动关闭会话 |
| `GET` | `/api/admin/terminal-sessions/{sessionId}/record` | 获取会话录制信息 |

说明：

- 终端交互建议通过 WebSocket 或安全的流式通道承载。
- 创建终端会话前建议校验审批状态、操作人权限和实例状态。

### 22.7 监控与告警接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/admin/instances/{instanceId}/metrics` | 查询实例监控图表数据 |
| `GET` | `/api/admin/containers/{containerId}/metrics` | 查询容器监控图表数据 |
| `GET` | `/api/admin/nodes/{nodeId}/metrics` | 查询节点监控图表数据 |
| `GET` | `/api/admin/clusters/{clusterId}/metrics` | 查询集群监控图表数据 |
| `GET` | `/api/admin/alerts` | 查询告警列表 |
| `POST` | `/api/admin/alerts/{alertId}/ack` | 确认告警 |
| `POST` | `/api/admin/alerts/{alertId}/close` | 关闭告警 |

### 22.8 订单与支付接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/portal/orders` | 创建订单 |
| `GET` | `/api/portal/orders` | 查询当前租户订单列表 |
| `GET` | `/api/portal/orders/{orderId}` | 查询订单详情 |
| `POST` | `/api/portal/orders/{orderId}/pay` | 创建支付单 |
| `POST` | `/api/portal/subscriptions/{subscriptionId}/renew` | 发起续费 |
| `POST` | `/api/portal/subscriptions/{subscriptionId}/upgrade` | 发起升级 |
| `POST` | `/api/portal/orders/{orderId}/refunds` | 申请退款 |
| `GET` | `/api/admin/payments` | 管理员查询支付流水 |
| `GET` | `/api/admin/refunds` | 管理员查询退款记录 |
| `POST` | `/api/callback/payment/{channel}` | 支付渠道回调 |
| `POST` | `/api/callback/marketplace/{platform}` | 第三方市场开通或续费回调 |

### 22.9 任务与审计接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/portal/jobs` | 用户查询自己的异步任务 |
| `GET` | `/api/admin/jobs` | 管理员查询全部异步任务 |
| `GET` | `/api/admin/jobs/{jobId}` | 查询任务详情 |
| `POST` | `/api/admin/jobs/{jobId}/retry` | 重试失败任务 |
| `GET` | `/api/admin/audit-events` | 查询审计事件 |
| `GET` | `/api/admin/operation-logs` | 查询操作日志 |

### 22.10 典型异步任务类型

建议平台至少统一以下异步任务类型：

- `instance_create`
- `instance_start`
- `instance_stop`
- `instance_restart`
- `instance_delete`
- `instance_scale`
- `config_validate`
- `config_publish`
- `backup_create`
- `backup_restore`
- `terminal_session_open`
- `payment_reconcile`
- `marketplace_provision`

## 23. Kubernetes 部署与功能细化

### 23.1 部署模式建议

建议按以下三种模式区分部署方式：

| 模式 | 适用场景 | 说明 |
| --- | --- | --- |
| 本地验证模式 | 开发、自测、方案验证 | 可直接使用 OpenClaw 官方 Docker 安装方式，重点验证连通性和基础功能 |
| K8s 预生产模式 | 联调、压测、验收 | 在单集群或单命名空间中部署 `Deployment`、`Service`、`Ingress`、`ConfigMap`、`Secret` 和 `PVC`，验证平台控制面与运维链路 |
| K8s 生产模式 | 多租户、多个用户、持续运营 | 采用 Kubernetes 或等价编排平台，以多副本、高可用、审计和弹性扩缩容作为默认前提 |

说明：

- OpenClaw 官方 Docker 方案适合作为起点，但平台化、多用户、支付和审计场景仍需要补充控制平面。
- OpenClaw 官方 Kubernetes 清单文档明确说明其清单更偏“起步模板”，不能直接等同于生产就绪方案。

### 23.2 推荐工作负载拓扑

面向平台化部署，建议至少包含以下工作负载或服务：

| 组件 | 角色 | 是否必须 | 说明 |
| --- | --- | --- | --- |
| `openclaw-gateway` | OpenClaw 运行面 | 必须 | 提供实际网关能力、Control UI、通道与运行逻辑 |
| `platform-api` | 平台后端控制面 | 必须 | 管理租户、实例、配置、备份、支付、审批和审计 |
| `portal-web` | 用户前端门户 | 必须 | 面向客户和租户管理员 |
| `admin-web` | 管理后台 | 必须 | 面向运维、客服、财务和管理员 |
| `postgres` | 业务主库 | 必须 | 保存租户、实例、订单、支付、审计等数据 |
| `redis` | 缓存与锁 | 必须 | 用于缓存、分布式锁、任务协调和会话辅助 |
| `object-storage` | 备份与录制存储 | 必须 | 可用 MinIO 或云对象存储 |
| `job-worker` | 异步任务执行器 | 必须 | 处理创建、恢复、支付补单、通知等后台任务 |
| `monitoring` | 指标采集 | 建议 | 使用 Prometheus、VictoriaMetrics 或云监控 |
| `logging` | 日志检索 | 建议 | 使用 Loki、Elasticsearch 或云日志平台 |
| `grafana` | 图表展示 | 建议 | 展示容器、实例、节点和支付图表 |
| `ingress-gateway` | 统一入口 | 建议 | 提供 HTTPS、域名路由、访问控制和 SSO 集成 |

### 23.3 Kubernetes 运行要求

OpenClaw 相关工作负载在 Kubernetes 场景中建议满足以下要求：

- 核心运行面优先使用 `Deployment` 部署，并配置 `RollingUpdate` 策略。
- 必须提供 `startupProbe`、`livenessProbe` 和 `readinessProbe`。
- 配置文件必须通过 `ConfigMap`、`Secret` 或配置中心注入，禁止依赖容器内手工修改。
- 工作目录、备份目录和录屏目录必须挂到 `PVC` 或对象存储。
- 日志必须输出到标准输出，并同步进入集中日志系统。
- 高风险宿主机能力不得直接开放给前端；集群访问必须通过受控 ServiceAccount 和最小权限 `RBAC`。
- 如涉及外网访问或反向代理，需要显式处理绑定地址、允许来源和可信代理设置。

说明：

- 根据 OpenClaw 官方文档，默认更偏本地安全模式；如果需要通过反向代理或 HTTPS 对外暴露，需要额外处理 `gateway.bind`、`gateway.allowedOrigins`、`gateway.trustedProxies` 等相关设置。
- Control UI 与配对机制可以保留，但平台化场景下建议再叠加租户和管理员权限体系。

### 23.4 Kubernetes 资源与限制建议

建议以以下资源策略作为 Kubernetes 层的配置基线：

- `openclaw-gateway`：`requests: 1 vCPU / 2 GiB`，`limits: 2 vCPU / 4 GiB`
- `platform-api`：`requests: 500m / 1 GiB`，`limits: 2 vCPU / 2 GiB`
- `job-worker`：`requests: 500m / 1 GiB`，`limits: 1 vCPU / 2 GiB`
- `postgres`：`requests: 1 vCPU / 2 GiB`，`limits: 2 vCPU / 4 GiB`
- `redis`：`requests: 250m / 512 MiB`，`limits: 1 vCPU / 1 GiB`

实现注意事项：

- Kubernetes 集群通常不应依赖 Swap，应以 `requests/limits`、节点预留资源和 `ResourceQuota` 作为容量控制手段。
- 监控、日志和备份组件不建议与 OpenClaw 共用过小资源池，否则高峰期会互相挤占。
- 多实例部署时，应优先限制单实例资源上限，并结合命名空间配额、`LimitRange` 和节点容量做调度治理。

### 23.5 Kubernetes 安全边界

Kubernetes 相关高风险能力建议按以下原则实施：

- `portal-web` 和普通业务 API 不允许直接访问 Kubernetes API 或集群管理凭证。
- 只有 `platform-api` 或独立 `runtime-agent` 可以操作运行时，并且必须使用最小权限 `ServiceAccount`。
- 远程终端、日志下载、文件查看、重启实例等高风险能力必须经过权限校验和审计。
- 不建议默认开启 `privileged`、`hostNetwork`、`hostPID` 或大范围 `hostPath` 挂载。
- 如确需访问宿主机或低层运行时，必须通过受控 Agent、审批、命令限制和会话记录实现。

### 23.6 Kubernetes 运维功能清单

为满足管理员端运维，Kubernetes 相关能力至少应包括：

- 查看用户对应的实例列表。
- 查看实例对应的 `Namespace`、`Deployment`、`Pod` 和容器列表。
- 查看容器 ID、镜像、节点、启动时间、重启次数和状态。
- 查看容器 CPU、内存、磁盘、网络、OOM 和事件图表。
- 查看容器日志、Kubernetes Event 和发布记录。
- 一键重启实例、滚动重建副本或执行受控扩缩容。
- 发起远程终端会话。
- 调整实例资源规格。
- 触发备份、恢复和版本更新。

### 23.7 示例文件与说明

当前目录已补充一份平台级 Docker Compose 示例，用于本地验证或评审拓扑与资源边界：

- [OpenClaw-platform-docker-compose.example.yml](C:\Users\Administrator\Desktop\openclaw调研\OpenClaw-platform-docker-compose.example.yml)

说明：

- 该文件是评审草案，不是可直接投入生产的最终编排文件。
- 重点用于确认服务拆分、挂载、网络、资源限制和管理边界。
- 如果进入正式实施阶段，建议补充 `Helm Chart` 或 `Kustomize` 清单作为 Kubernetes 交付物。

## 24. 前端功能细化

### 24.1 前端应用拆分建议

建议前端至少拆分为以下两个主应用：

| 应用 | 面向对象 | 主要职责 |
| --- | --- | --- |
| `portal-web` | 租户管理员、普通用户 | 服务订阅、实例访问、配置操作、备份恢复、订单与支付 |
| `admin-web` | 运维、平台管理员、客服、财务 | 租户管理、实例管理、容器诊断、监控图表、审批、支付与对账 |

如后续平台运营复杂度继续上升，可再拆分为：

- `ops-console`：偏运维和监控
- `finance-console`：偏订单、支付、发票和对账

### 24.2 与 OpenClaw 原生 Control UI 的关系

前端建设建议遵循以下原则：

- 平台前端负责“租户、实例、配置、备份、支付、审计、监控”。
- OpenClaw 原生 Control UI 负责“高级运行调试、Chat 调试、技能调试、渠道和节点运维”。
- 第一阶段不建议完全重写 OpenClaw 原生 Control UI 的全部页面。
- 更合理的方式是：平台侧做统一入口和权限控制，必要时跳转或嵌入原生 Control UI 的高级能力。

### 24.3 用户侧功能矩阵

| 模块 | 普通用户 | 租户管理员 |
| --- | --- | --- |
| 查看已购服务 | 支持 | 支持 |
| 查看访问入口 | 支持 | 支持 |
| 查看运行状态 | 支持 | 支持 |
| 修改基础配置 | 限制 | 支持 |
| 发布配置 | 限制 | 支持 |
| 创建备份 | 限制 | 支持 |
| 恢复备份 | 不支持或审批 | 支持或审批 |
| 续费与升级 | 不支持 | 支持 |
| 查看订单和发票 | 限制 | 支持 |
| 查看审计记录 | 限制 | 支持 |

### 24.4 管理员侧功能矩阵

| 模块 | 运维管理员 | 客服管理员 | 财务管理员 | 审计管理员 |
| --- | --- | --- | --- | --- |
| 查看租户与用户 | 支持 | 支持 | 支持 | 支持 |
| 查看实例与容器 | 支持 | 部分支持 | 只读 | 只读 |
| 远程终端 | 支持且需审批 | 不支持 | 不支持 | 只读查看记录 |
| 查看监控图表 | 支持 | 部分支持 | 只读 | 只读 |
| 删除实例 | 支持且需审批 | 不支持 | 不支持 | 不支持 |
| 查看订单支付 | 支持 | 支持 | 支持 | 只读 |
| 手工改支付状态 | 不支持或极限权限 | 不支持 | 支持且需审批 | 审计 |
| 查看审计日志 | 支持 | 限制 | 限制 | 支持 |

### 24.5 关键交互要求

前端关键操作建议统一遵循以下交互约束：

- 创建、删除、发布配置、恢复备份、开终端等操作全部按异步任务处理。
- 前端不等待长任务直接完成，而是展示任务进度、阶段状态和失败原因。
- 所有高风险动作必须二次确认，并显示操作影响范围。
- 图表查询默认按时间范围和对象维度筛选。
- 终端页面必须展示操作人、实例、容器、审批单号和会话剩余时间。
- 日志和监控页面建议支持“从实例跳转到容器”和“从告警跳转到图表”。

### 24.6 前端非功能要求

建议前端至少满足以下要求：

- 支持基于角色的按钮级权限控制。
- 支持中文为主的界面文案。
- 支持对异步任务、告警和支付结果进行实时刷新。
- 支持按租户隔离缓存与数据访问范围。
- 管理员端支持大屏和桌面分辨率优先布局。
- 用户侧门户支持桌面优先并兼容移动端基础访问。

## 25. 后端功能细化

### 25.1 后端模块矩阵

| 模块 | 主要职责 | 关键能力 |
| --- | --- | --- |
| `iam-service` | 身份与权限 | 登录、角色、租户隔离、访问令牌 |
| `tenant-service` | 租户与用户管理 | 租户生命周期、用户管理、角色绑定 |
| `instance-service` | 实例生命周期 | 创建、删除、启动、停止、规格调整、到期治理 |
| `runtime-adapter` | Kubernetes / Docker 抽象 | 创建容器、列举容器、日志、终端、重启、资源信息 |
| `config-service` | 配置治理 | 模板、版本、校验、发布、回滚 |
| `backup-service` | 备份与恢复 | 备份生成、存储、恢复、保留策略 |
| `monitor-service` | 指标与告警 | 指标聚合、图表查询、告警生成与关闭 |
| `approval-service` | 高风险动作审批 | 删除、恢复、终端进入、支付状态修正审批 |
| `billing-service` | 商业规则 | 套餐、价格、订单、订阅、账单 |
| `payment-service` | 支付通道接入 | 下单、回调、验签、退款、对账 |
| `notification-service` | 通知 | 邮件、短信、站内信、Webhook |
| `audit-service` | 审计 | 操作日志、会话记录、风险事件 |

### 25.2 Runtime Adapter 设计建议

为避免后端与 Kubernetes API 或具体运行时强耦合，建议单独抽象运行时适配层，统一提供以下能力：

- `createInstance`
- `deleteInstance`
- `startInstance`
- `stopInstance`
- `restartInstance`
- `scaleInstance`
- `listRuntimeUnits`
- `getRuntimeUnitDetail`
- `fetchLogs`
- `fetchEvents`
- `openTerminalSession`
- `closeTerminalSession`
- `getResourceLimits`

说明：

- 在 Kubernetes 模式下，这一层优先调用 Kubernetes API 或受控 Runtime Agent。
- 在 Docker 模式下，这一层仅用于本地验证、兼容环境或应急排障。
- 上层业务服务不应感知底层是 Docker 还是 Kubernetes。

### 25.3 后端必须实现的实例状态机

建议至少定义以下实例状态：

- `pending_payment`
- `provisioning`
- `running`
- `updating`
- `backing_up`
- `restoring`
- `stopping`
- `stopped`
- `expired`
- `deleting`
- `deleted`
- `failed`

状态流转原则：

- 支付未完成前不进入正式运行态。
- 删除前必须经过终态检查、备份策略检查和审批流程。
- 备份、恢复、配置发布等操作进入专门中间态，避免并发冲突。

### 25.4 后端必须实现的订单状态机

建议至少定义以下订单状态：

- `pending`
- `paying`
- `paid`
- `provisioning`
- `active`
- `refund_pending`
- `refunded`
- `closed`
- `failed`

建议实例与订单建立明确联动：

- 订单 `paid` 后触发实例开通或升级。
- 订单 `refund_pending` 时限制高风险动作。
- 订单 `refunded` 后根据规则触发停服、降级或释放。

### 25.5 审批与审计闭环

后端应将以下动作纳入审批与审计闭环：

- 删除实例
- 恢复备份
- 打开远程终端
- 手工修改支付状态
- 发布高风险配置

闭环要求：

- 申请人、审批人、操作对象、审批结果全部留痕。
- 审批通过后产生一次性执行令牌。
- 执行动作后自动回写任务结果和审计日志。
- 超时未使用的审批令牌自动失效。

### 25.6 后端异步任务约束

所有长任务建议进入统一任务队列，并满足以下要求：

- 支持幂等提交。
- 支持重试与最大重试次数。
- 支持失败原因落库。
- 支持阶段化进度回传。
- 支持任务取消和超时回收。

建议至少定义以下任务阶段：

- `accepted`
- `queued`
- `running`
- `verifying`
- `succeeded`
- `failed`
- `cancelled`

### 25.7 后端对外能力边界

后端设计建议严格控制以下边界：

- 用户前端永远不直接访问 Docker 或 Kubernetes。
- 管理后台只访问平台 API，不直接操作容器运行时。
- 只有运行时适配层或受控 Agent 可以触达 Docker Socket、Kubernetes API 或宿主机能力。
- 支付回调、第三方市场回调和审批执行必须通过单独入口并做验签和幂等处理。

### 25.8 实施优先级建议

如果需要按审查顺序逐步实现，建议优先级如下：

1. `instance-service + runtime-adapter + config-service`
2. `backup-service + monitor-service + audit-service`
3. `portal-web + admin-web` 的实例和监控主流程
4. `billing-service + payment-service + notification-service`
5. `approval-service + marketplace callback` 等扩展能力

## 26. 联调与验收标准

### 26.1 部署与运行验收

建议至少满足以下验收项：

- 能基于统一配置在新环境完成一次完整实例开通，不依赖手工进入容器修改关键参数。
- 能以 `2` 个副本运行平台控制面核心服务，并完成一次滚动发布验证。
- 实例创建后可自动分配访问地址，用户侧门户和管理后台都能正确展示。
- 健康检查、日志采集、指标采集、告警链路均可正常工作。
- 备份文件可写入独立持久化存储，并能在新实例或原实例上完成一次恢复演练。
- 删除、恢复、进入远程终端等高风险动作必须经过权限校验和审计记录。

### 26.2 后端 API 联调验收

后端联调阶段建议至少完成以下接口验证：

- `createInstance` 可返回实例标识、任务标识和初始状态。
- `getInstanceStatus` 可查询实例状态、运行位置、版本和最近一次任务结果。
- `getInstanceAccessInfo` 可返回用户前台地址、后台地址和访问方式。
- `getInstanceConfig` 与 `updateInstanceConfig` 可完成一次读取、修改、发布和结果回传。
- `backupInstance`、`listBackups`、`restoreBackup` 可完成一次端到端备份恢复闭环。
- `deleteInstance` 可完成删除前校验、可选备份、资源释放和状态回写。

接口行为建议同时满足：

- 所有长任务返回统一 `taskId`。
- 支持幂等键，避免重复创建实例或重复扣费。
- 失败时返回明确错误码、错误信息和失败阶段。
- 关键任务支持回调、轮询或事件通知三选一，且需在项目内统一。

### 26.3 支付与订单联动验收

如项目进入商业化交付，建议补充以下验收项：

- 用户支付成功后，订单状态、订阅状态和实例状态能按预期联动。
- 升级、续费、退款、到期停服、宽限期释放等场景都有明确状态变化。
- 支付回调具备验签、幂等、防重放和异常补偿机制。
- 对账结果可落库并支持异常订单人工复核。

### 26.4 安全与审计验收

安全相关能力建议至少通过以下验证：

- 普通用户无法调用管理员 API，也无法直接接触 Docker 或 Kubernetes 凭证。
- 管理员高风险动作必须经过审批流、双重确认或白名单校验。
- 远程终端会话能记录操作人、目标实例、开始结束时间和命令执行痕迹。
- 审计日志可按租户、用户、实例、操作类型和时间范围检索。

## 27. 交付物清单建议

### 27.1 部署交付物

建议最终交付至少包含以下内容：

- Kubernetes 交付清单：`Namespace`、`Deployment`、`Service`、`Ingress` 或 `Gateway`、`ConfigMap`、`Secret`、`PVC`、`HPA`、`PDB`、`NetworkPolicy`
- `Helm Chart` 或 `Kustomize` 目录，用于不同环境参数化部署
- 环境变量模板和密钥注入说明
- 基础监控、日志、备份存储的接入说明
- 回滚、扩容、故障切换和恢复演练说明

### 27.2 后端交付物

后端建议输出以下标准化交付物：

- OpenAPI 或等价接口文档
- 异步任务模型与状态机说明
- 错误码清单与失败补偿策略
- 权限模型、角色矩阵和审批流说明
- 支付回调、通知回调和市场回调接入规范

### 27.3 数据与运维交付物

建议同步交付以下数据与运维材料：

- 最终版数据库 DDL 与迁移脚本
- 初始化数据脚本，例如角色、套餐、配置模板
- 备份保留策略、恢复流程和数据保留周期说明
- 监控面板、告警规则和运维 Runbook
- 联调测试用例和上线检查清单

## 28. 与当前目录文件的对应关系

### 28.1 Docker Compose 示例文件定位

当前目录中的 `OpenClaw-platform-docker-compose.example.yml` 可视为本地验证和架构评审样例，已覆盖以下主要组件：

- `reverse-proxy`：统一入口代理
- `portal-web`：用户侧门户前端
- `admin-web`：管理员后台前端
- `platform-api`：平台控制面 API
- `job-worker`：异步任务执行器
- `runtime-agent`：运行时适配层或受控 Agent
- `openclaw-gateway`：OpenClaw 实际运行实例
- `postgres`、`redis`、`minio`：基础数据与对象存储
- `prometheus`、`loki`、`grafana`：监控与日志基础设施

该文件适合作为：

- 本地演示拓扑
- 架构评审输入
- 联调前的最小可运行环境

该文件暂不等同于生产交付物。若进入正式交付阶段，建议补齐 Kubernetes 清单并以 `Helm Chart` 或 `Kustomize` 作为主交付形式。

### 28.2 数据库 DDL 初稿对应关系

当前目录中的 `OpenClaw-数据库DDL初稿.sql` 已覆盖平台化所需的大部分核心业务域：

- 产品与套餐域：`product`、`service_plan`、`plan_price`
- 租户与用户域：`tenant`、`user_account`、`app_role`、`user_role_rel`
- 集群与实例域：`cluster`、`cluster_node`、`service_instance`、`runtime_container`、`instance_member`、`instance_access`
- 配置与变更域：`instance_config`、`instance_config_history`
- 备份与恢复域：`backup_record`、`restore_record`
- 订单与支付域：`order_main`、`order_item`、`subscription`、`payment_transaction`、`refund_record`、`invoice_record`
- 运维与审计域：`operation_job`、`operation_log`、`terminal_session`、`approval_record`、`notification_record`、`alert_record`、`audit_event`

这份 DDL 可作为平台数据库设计初稿，但正式实施时仍建议补充以下动作：

- 按版本输出迁移脚本，而不是长期只维护一份初稿 SQL。
- 明确枚举值、状态机和字段字典，避免前后端各自维护不同状态。
- 为核心查询路径补齐索引、唯一约束、外键级联策略和归档策略。
- 明确 `runtime_type` 在 Docker 与 Kubernetes 两种模式下的取值规范和对象映射方式。

### 28.3 文档、DDL 与部署样例的关系

建议将当前三个文件理解为三个层次：

- 本文档：定义目标能力、边界、验收标准和实施方向。
- `OpenClaw-数据库DDL初稿.sql`：定义平台控制面的数据模型初稿。
- `OpenClaw-platform-docker-compose.example.yml`：定义本地验证与评审环境的组件拓扑样例。

后续若进入研发阶段，建议以本文档为总需求基线，以 DDL 和部署清单作为可执行交付物持续迭代。

## 29. 待确认问题清单

在正式排期前，建议至少确认以下问题：

- 多租户模式最终采用“共享实例逻辑隔离”还是“租户独立实例”。
- OpenClaw 原生可直接复用的管理接口范围，以及哪些能力必须由平台额外封装。
- 备份范围是否包含会话数据、技能配置、日志、上传文件和工作区数据。
- 支付渠道范围是否只包含国内渠道，还是需要同时支持 Stripe 等海外渠道。
- 是否需要对接企业单点登录、企业微信、钉钉或 LDAP。
- 数据保留周期、删除后保留天数、审计留存时间和合规要求由谁定义。
- 生产环境可接受的可用性目标、RTO、RPO 和最大恢复时间是否已有约束。
- 远程终端是否默认关闭，仅在审批后临时开启，是否需要全量录屏或命令级留痕。
