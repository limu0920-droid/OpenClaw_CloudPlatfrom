# Platform API Production Environment Template

## GitHub Environment

建议在 GitHub 中创建 `production` environment，并配置审批规则。

当前 `platform-api-release` workflow 不再接收明文业务密钥，职责拆分如下：

- GitHub Environment 变量：提供非敏感生产配置和 `ExternalSecret` 元数据。
- Secret Manager / Vault：保存数据库、对象存储、Keycloak、OpenSearch 的真实密钥。
- 集群内 External Secrets Operator：把远端密钥物化为 `openclaw-platform-api-secret`。

### Required Repository / Environment Variables

- `PLATFORM_API_API_HOST`
- `PLATFORM_API_TLS_SECRET_NAME`
- `PLATFORM_API_KEYCLOAK_BASE_URL`
- `PLATFORM_API_PORTAL_LOGIN_URL`
- `PLATFORM_API_PORTAL_POST_LOGIN_URL`
- `PLATFORM_API_PORTAL_LOGOUT_URL`
- `PLATFORM_API_OPENSEARCH_URL`
- `PLATFORM_API_OBJECT_STORAGE_ENDPOINT`
- `PLATFORM_API_WORKSPACE_BRIDGE_PUBLIC_BASE_URL`
- `PLATFORM_API_ARTIFACT_PREVIEW_PUBLIC_BASE_URL`
- `PLATFORM_API_ARTIFACT_PREVIEW_ALLOWED_HOSTS`
- `PLATFORM_API_SECRET_STORE_KIND`
- `PLATFORM_API_SECRET_STORE_NAME`
- `PLATFORM_API_SECRET_REFRESH_INTERVAL`
- `PLATFORM_API_SECRET_KEY_PREFIX`

### Optional Variables For Post-Release Verification

- `PLATFORM_API_DATABASE_SECRET_NAME`
- `PLATFORM_API_EXTERNAL_SECRET_NAME`

## Secret Manager Keys

默认 `PLATFORM_API_SECRET_KEY_PREFIX` 建议使用 `/openclaw/platform-api/prod/cn-east-1/prod-cn-east-1-a/default`，并在远端 Secret Manager 下准备以下键：

- `${PREFIX}/DATABASE_URL`
- `${PREFIX}/OBJECT_STORAGE_ACCESS_KEY`
- `${PREFIX}/OBJECT_STORAGE_SECRET_KEY`
- `${PREFIX}/KEYCLOAK_CLIENT_SECRET`
- `${PREFIX}/KEYCLOAK_SESSION_SECRET`
- `${PREFIX}/OPENSEARCH_USERNAME`
- `${PREFIX}/OPENSEARCH_PASSWORD`
- `${PREFIX}/WECHATPAY_MCH_ID`
- `${PREFIX}/WECHATPAY_APP_ID`
- `${PREFIX}/WECHATPAY_CLIENT_SECRET`
- `${PREFIX}/WECHATPAY_NOTIFY_URL`
- `${PREFIX}/WECHATPAY_REFUND_NOTIFY_URL`
- `${PREFIX}/WECHATPAY_SERIAL_NO`
- `${PREFIX}/WECHATPAY_PUBLIC_KEY_ID`
- `${PREFIX}/WECHATPAY_PUBLIC_KEY_PEM`
- `${PREFIX}/WECHATPAY_PRIVATE_KEY_PEM`
- `${PREFIX}/WECHATPAY_APIV3_KEY`
- `${PREFIX}/WECHATPAY_MODE`
- `${PREFIX}/WECHATPAY_SUB_MCH_ID`
- `${PREFIX}/WECHATPAY_SUB_APP_ID`
- `${PREFIX}/WECHAT_LOGIN_APP_ID`
- `${PREFIX}/WECHAT_LOGIN_APP_SECRET`
- `${PREFIX}/WECHAT_LOGIN_REDIRECT_URL`
- `${PREFIX}/WORKSPACE_BRIDGE_TOKEN`

其中 `${PREFIX}` 由 `PLATFORM_API_SECRET_KEY_PREFIX` 决定，发布前必须先完成这些 key 的写入。

如果当前环境尚未启用微信支付或微信登录，相关 key 仍建议在 Secret Manager 中预创建为空字符串，避免 `ExternalSecret` 因缺 key 而无法进入 Ready。

## Recommended Operational Settings

- 开启 `production` environment 审批
- 限制谁可以触发 tag 发布
- 保留 release artifact 与渲染产物
- 对 GHCR 镜像开启保留策略
- 对 `ClusterSecretStore` / `SecretStore` 配置变更启用审批

## Suggested Secret Ownership

- 数据库 URL：DBA / 平台运维
- 对象存储 AK/SK：平台运维
- Keycloak secret：IAM 管理员
- OpenSearch 用户密码：日志平台管理员
- SecretStore / Key Prefix：平台运维 / SRE

## Before Enabling Strict Release

必须确认：

- `prod` overlay 已固定 `PLATFORM_STRICT_MODE=true`
- `set-config.ps1` 注入的域名、URL、SecretStore 名称和 key prefix 都不是模板占位符
- `set-config.ps1` 注入了 `WORKSPACE_BRIDGE_PUBLIC_BASE_URL`、`ARTIFACT_PREVIEW_PUBLIC_BASE_URL`、`ARTIFACT_PREVIEW_ALLOWED_HOSTS`
- `security/prod/externalsecret.yaml` 已指向真实 `SecretStore` 与远端 key prefix
- 目标集群已安装 External Secrets Operator
- `set-image.ps1` 使用的是 digest，不是 tag
- `preflight.ps1 -Overlay prod -Strict` 在本地或预发环境通过
- `verify-dependencies.ps1` 能确认 `ExternalSecret` Ready 且 `openclaw-platform-api-secret` 已物化
