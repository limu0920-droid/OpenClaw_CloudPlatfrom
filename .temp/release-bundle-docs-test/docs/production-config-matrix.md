# Platform API Production Config Matrix

本文档定义 `platform-api` 的正式生产配置矩阵。目标不是把真实密钥写进仓库，而是把“哪些参数必须存在、从哪里注入、按什么维度拆分、上线前如何验证”固定成可执行标准。

## 1. 维度定义

| 维度 | 典型值 | 说明 |
| --- | --- | --- |
| environment | `dev` / `staging` / `prod` | 发布环境边界 |
| region | `cn-east-1` / `cn-north-1` | 地域维度，决定域名、对象存储与集群邻近性 |
| cluster | `prod-cn-east-1-a` / `prod-cn-north-1-a` | Kubernetes 集群维度 |
| merchant | `default` / `service-provider` / `oem-a` | 支付商户或 OEM 结算主体维度 |

推荐 Secret Manager key prefix 规范：

```text
/openclaw/platform-api/{environment}/{region}/{cluster}/{merchant}
```

例如：

```text
/openclaw/platform-api/prod/cn-east-1/prod-cn-east-1-a/default
```

## 2. 配置来源分层

| 类别 | 载体 | 写入方式 | 说明 |
| --- | --- | --- | --- |
| 非敏感生产参数 | `ConfigMap` / `Ingress` patch | `set-config.ps1` | 域名、公开 URL、公开服务地址 |
| 敏感参数 | `ExternalSecret` -> runtime `Secret` | Secret Manager + `security/prod/externalsecret.yaml` | 数据库、对象存储、Keycloak、微信支付、微信登录、桥接 token |
| 固定运行参数 | `base/configmap.yaml` | 仓库默认值 + overlay patch | 路径模板、超时、重试、开关默认值 |

## 3. 非敏感参数矩阵

| 参数 | 注入位置 | 维度 | 严格模式 | 说明 |
| --- | --- | --- | --- | --- |
| `ApiHost` | `Ingress.host`、`tls.hosts` | environment + region + cluster | 必需 | 平台 API 主域名 |
| `TlsSecretName` | `Ingress.tls.secretName` | environment + cluster | 必需 | 对应集群中的 TLS Secret |
| `KEYCLOAK_BASE_URL` | `ConfigMap` | environment + region | Keycloak 开启时必需 | IAM 服务地址 |
| `KEYCLOAK_REDIRECT_URL` | `ConfigMap` | environment + region | Keycloak 开启时必需 | 登录页回调入口 |
| `KEYCLOAK_POST_LOGIN_REDIRECT_URL` | `ConfigMap` | environment + region | Keycloak 开启时必需 | 登录后默认跳转 |
| `KEYCLOAK_LOGOUT_REDIRECT_URL` | `ConfigMap` | environment + region | Keycloak 开启时必需 | 登出后跳转 |
| `OPENSEARCH_URL` | `ConfigMap` | environment + region | OpenSearch 开启时必需 | 日志检索集群地址 |
| `OBJECT_STORAGE_ENDPOINT` | `ConfigMap` | environment + region | strict 必需 | 对象存储公开或内网地址 |
| `WORKSPACE_BRIDGE_PUBLIC_BASE_URL` | `ConfigMap` | environment + region | 生产建议必需 | 供 Lobster 异步回调使用的公开基地址 |
| `ARTIFACT_PREVIEW_PUBLIC_BASE_URL` | `ConfigMap` | strict 必需 | strict 必需 | 预览与下载公共基地址 |
| `ARTIFACT_PREVIEW_ALLOWED_HOSTS` | `ConfigMap` | environment + region | 强烈建议 | 可信产物域名白名单，逗号分隔 |
| `SecretStoreKind` | `ExternalSecret.secretStoreRef.kind` | cluster | 必需 | `ClusterSecretStore` 或 `SecretStore` |
| `SecretStoreName` | `ExternalSecret.secretStoreRef.name` | cluster | 必需 | 集群内 ESO store 名称 |
| `SecretRefreshInterval` | `ExternalSecret.refreshInterval` | cluster | 必需 | 建议 `15m` 到 `1h` |
| `SecretKeyPrefix` | `ExternalSecret.data[*].remoteRef.key` | environment + region + cluster + merchant | 必需 | 指向该部署实例的远端 key 前缀 |

## 4. Secret Manager Key 矩阵

以下 key 通过 `security/prod/externalsecret.yaml` 注入。

### 4.1 核心上线必需项

| Key | 用途 | 说明 |
| --- | --- | --- |
| `DATABASE_URL` | PostgreSQL 持久化 | strict 必需 |
| `OBJECT_STORAGE_ACCESS_KEY` | 对象存储归档 | strict 必需 |
| `OBJECT_STORAGE_SECRET_KEY` | 对象存储归档 | strict 必需 |
| `KEYCLOAK_CLIENT_SECRET` | Keycloak OIDC | Keycloak 开启时必需 |
| `KEYCLOAK_SESSION_SECRET` | 签名 Cookie | 生产必需，必须每环境唯一 |
| `OPENSEARCH_USERNAME` | OpenSearch 认证 | OpenSearch 开启时必需 |
| `OPENSEARCH_PASSWORD` | OpenSearch 认证 | OpenSearch 开启时必需 |

### 4.2 支付与认证扩展项

| Key | merchant 维度 | 说明 |
| --- | --- | --- |
| `WECHATPAY_MCH_ID` | 必须按 merchant 拆分 | 直连商户或服务商商户号 |
| `WECHATPAY_APP_ID` | 必须按 merchant 拆分 | 支付 AppID |
| `WECHATPAY_CLIENT_SECRET` | 可按 merchant 拆分 | 兼容客户端扩展场景 |
| `WECHATPAY_NOTIFY_URL` | environment + region | 支付通知地址 |
| `WECHATPAY_REFUND_NOTIFY_URL` | environment + region | 退款通知地址 |
| `WECHATPAY_SERIAL_NO` | merchant | 商户证书序列号 |
| `WECHATPAY_PUBLIC_KEY_ID` | merchant | 微信支付公钥 ID |
| `WECHATPAY_PUBLIC_KEY_PEM` | merchant | 验签公钥 |
| `WECHATPAY_PRIVATE_KEY_PEM` | merchant | 商户私钥 |
| `WECHATPAY_APIV3_KEY` | merchant | API v3 解密密钥 |
| `WECHATPAY_MODE` | merchant | `merchant` 或 `service-provider` |
| `WECHATPAY_SUB_MCH_ID` | service-provider 专属 | 服务商模式子商户号 |
| `WECHATPAY_SUB_APP_ID` | service-provider 专属 | 服务商模式子 AppID |
| `WECHAT_LOGIN_APP_ID` | environment + merchant | 微信登录 AppID |
| `WECHAT_LOGIN_APP_SECRET` | environment + merchant | 微信登录 secret |
| `WECHAT_LOGIN_REDIRECT_URL` | environment + region | 微信登录回调地址 |
| `WORKSPACE_BRIDGE_TOKEN` | cluster + environment | Lobster 回调鉴权 token |

说明：

- 未启用微信支付、微信登录时，可在 Secret Manager 中保留空字符串，但 key 必须存在，避免 `ExternalSecret` Ready 状态因缺 key 失败。
- `WECHATPAY_MODE=service-provider` 时，`WECHATPAY_SUB_MCH_ID` / `WECHATPAY_SUB_APP_ID` 不能留空。

## 5. 运行开关与默认策略

| 参数 | 默认值 | 生产建议 |
| --- | --- | --- |
| `PLATFORM_STRICT_MODE` | `true` in prod overlay | 必须开启 |
| `RUNTIME_PROVIDER` | `kubectl` in prod overlay | 不允许 `mock` |
| `WORKSPACE_BRIDGE_HISTORY_SYNC` | `true` | 保持开启 |
| `WORKSPACE_BRIDGE_TIMEOUT_SECS` | `10` | 按上游 SLA 调整 |
| `WORKSPACE_BRIDGE_RETRY_COUNT` | `2` | 保持小而确定 |
| `ARTIFACT_PREVIEW_ALLOW_PRIVATE_IP` | `false` | 生产必须为 `false` |
| `ARTIFACT_PREVIEW_HTML_MAX_BYTES` | `2097152` | 只在评估预览上限时调整 |
| `WECHATPAY_ENABLED` | `false` | 商业化闭环落地后再启用 |
| `WECHAT_LOGIN_ENABLED` | `false` | 仅在微信登录真实联调后启用 |

## 6. 上线前校验矩阵

| 阶段 | 命令 | 必须验证 |
| --- | --- | --- |
| 渲染前 | `set-config.ps1` | `ApiHost`、Keycloak URL、OpenSearch URL、对象存储 endpoint、preview / bridge public URL 已写入 overlay |
| 静态预检 | `preflight.ps1 -Overlay prod -Strict` | digest 镜像、无模板值、`ARTIFACT_PREVIEW_PUBLIC_BASE_URL`、`WORKSPACE_BRIDGE_PUBLIC_BASE_URL`、TLS、ExternalSecret 存在 |
| bootstrap 预检 | `preflight.ps1 -Overlay prod -Bootstrap -Strict` | bootstrap overlay 与正式 overlay 一致 |
| 外部依赖校验 | `verify-dependencies.ps1` | `ExternalSecret` Ready、runtime Secret 已物化、核心 key 无占位符 |
| 发布后验证 | `verify-release.ps1` + `smoke.ps1` | strict mode、state backend、runtime provider、健康检查、基础接口可用 |
| 外部文档验证 | `GET /api/v1/docs/external` | OpenAPI 与接入说明可直接访问 |

## 7. 交付要求

正式交付包至少应包含：

- 渲染后的脱敏 manifest
- 本矩阵文档
- `apps/platform-api/internal/httpapi/externaldocs/openapi.yaml`
- `apps/platform-api/internal/httpapi/externaldocs/integration-guide.md`
- release checklist / operations runbook / rollback SOP

只要生产配置新增了新的环境变量或 Secret key，就必须同步更新：

1. `apps/platform-api/.env.example`
2. `deploy/k8s/platform-api/base/configmap.yaml` 或 `base/secret.yaml`
3. `deploy/k8s/platform-api/security/prod/externalsecret.yaml`
4. 本矩阵文档
