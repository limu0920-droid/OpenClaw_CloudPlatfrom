# Platform API K8s

该目录提供 `platform-api` 的基础 Kubernetes 清单与一个 `dev` overlay。
同时提供一个 `prod` overlay 入口，用于通过 GitHub Environment 变量和 `ExternalSecret` 注入真实生产配置。
数据库引导 Job 已拆分为单独的 `bootstrap/` kustomization，避免和应用 Deployment 同步发布时的初始化竞态。

## Render

```powershell
kubectl kustomize deploy/k8s/platform-api/overlays/dev
kubectl kustomize deploy/k8s/platform-api/overlays/prod
kubectl kustomize deploy/k8s/platform-api/bootstrap/overlays/dev
kubectl kustomize deploy/k8s/platform-api/bootstrap/overlays/prod
```

或：

```powershell
.\deploy\k8s\platform-api\scripts\render.ps1 -Overlay dev
.\deploy\k8s\platform-api\scripts\render.ps1 -Overlay prod
```

## Apply

```powershell
kubectl apply -k deploy/k8s/platform-api/overlays/dev
```

或：

```powershell
.\deploy\k8s\platform-api\scripts\validate.ps1 -Overlay dev
.\deploy\k8s\platform-api\scripts\apply.ps1 -Overlay dev
```

如需把固定镜像写入 overlay：

```powershell
.\deploy\k8s\platform-api\scripts\set-image.ps1 -Overlay prod -Image ghcr.io/openclaw/platform-api@sha256:...
```

如需把生产配置写入 overlay：

```powershell
.\deploy\k8s\platform-api\scripts\set-config.ps1 `
  -Overlay prod `
  -ApiHost "api.openclaw.example.com" `
  -TlsSecretName "openclaw-platform-api-tls" `
  -KeycloakBaseUrl "https://sso.openclaw.example.com" `
  -PortalLoginUrl "https://portal.openclaw.example.com/login" `
  -PortalPostLoginUrl "https://portal.openclaw.example.com/portal" `
  -PortalLogoutUrl "https://portal.openclaw.example.com/login" `
  -OpenSearchUrl "https://opensearch.openclaw.example.com" `
  -ObjectStorageEndpoint "https://minio.openclaw.example.com" `
  -WorkspaceBridgePublicBaseUrl "https://api.openclaw.example.com" `
  -ArtifactPreviewPublicBaseUrl "https://preview.openclaw.example.com" `
  -ArtifactPreviewAllowedHosts "minio.openclaw.example.com,api.openclaw.example.com" `
  -SecretStoreKind "ClusterSecretStore" `
  -SecretStoreName "openclaw-vault-prod" `
  -SecretRefreshInterval "15m" `
  -SecretKeyPrefix "/openclaw/platform-api/prod/cn-east-1/prod-cn-east-1-a/default"
```

发布前预检：

```powershell
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Bootstrap
```

如需对占位符也做严格拦截：

```powershell
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Strict
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Bootstrap -Strict
```

生产 overlay 默认要求：

- `PLATFORM_STRICT_MODE=true`
- `RUNTIME_PROVIDER` 不能为 `mock`
- 应用 Deployment 不允许启用 `AUTO_MIGRATE` / `AUTO_BOOTSTRAP`
- `prod` render 必须包含 `ExternalSecret/openclaw-platform-api-secret`
- `prod` render 必须包含 `WORKSPACE_BRIDGE_PUBLIC_BASE_URL`
- `prod` render 必须包含 `ARTIFACT_PREVIEW_PUBLIC_BASE_URL`
- `-Strict` 预检下不允许出现 `change-me` / `set-in-cluster` / `example.com` / `change-me-secret-store` 等模板占位符

查看滚动发布状态：

```powershell
.\deploy\k8s\platform-api\scripts\rollout-status.ps1
```

发布后做基础 smoke：

```powershell
.\deploy\k8s\platform-api\scripts\smoke.ps1
```

发布后核对版本与就绪状态：

```powershell
.\deploy\k8s\platform-api\scripts\verify-release.ps1 `
  -ExpectedVersion platform-api-v1.0.0 `
  -ExpectedCommit abcdef1234 `
  -ExpectedStrictMode true `
  -ExpectedStateBackend postgres `
  -ExpectedRuntimeProvider kubectl
```

发布前或发布后做外部依赖检查：

```powershell
.\deploy\k8s\platform-api\scripts\verify-dependencies.ps1 `
  -Namespace openclaw-system `
  -Service openclaw-platform-api `
  -DatabaseSecretName openclaw-platform-api-secret `
  -ExternalSecretName openclaw-platform-api-secret `
  -TlsSecretName openclaw-platform-api-tls `
  -OpenSearchUrl "https://opensearch.openclaw.example.com" `
  -KeycloakBaseUrl "https://sso.openclaw.example.com"
```

一键执行发布流程：

```powershell
.\deploy\k8s\platform-api\scripts\release.ps1 -Overlay prod -Image ghcr.io/openclaw/platform-api@sha256:...
```

生产发布默认要求先做数据库备份。
如需指定备份目标：

```powershell
.\deploy\k8s\platform-api\scripts\release.ps1 `
  -Overlay prod `
  -Image ghcr.io/openclaw/platform-api@sha256:... `
  -BackupPostgresPod postgresql-0 `
  -BackupOutput .\platform-api-pre-release-backup.sql `
  -CheckDependencies `
  -DatabaseSecretName openclaw-platform-api-secret `
  -ExternalSecretName openclaw-platform-api-secret `
  -TlsSecretName openclaw-platform-api-tls `
  -OpenSearchUrl "https://opensearch.openclaw.example.com" `
  -KeycloakBaseUrl "https://sso.openclaw.example.com" `
  -ExpectedVersion platform-api-v1.0.0 `
  -ExpectedCommit abcdef1234 `
  -CheckObservability
```

回滚：

```powershell
.\deploy\k8s\platform-api\scripts\rollback.ps1
```

发布演练：

```powershell
.\deploy\k8s\platform-api\scripts\rehearsal.ps1 -Overlay prod -BootstrapOverlay prod
```

导出交付包：

```powershell
.\deploy\k8s\platform-api\scripts\package-release.ps1 -Overlay prod -BootstrapOverlay prod -OutputDir .\release-bundle
```

校验数据库备份：

```powershell
.\deploy\k8s\platform-api\scripts\verify-backup.ps1 -Path .\platform-api-pre-release-backup.sql
```

PR / 手工演练流程会额外通过 [platform-api-rehearsal.yml](../../../.github/workflows/platform-api-rehearsal.yml) 生成脱敏后的生产 manifest artifact。

查看数据库 migration 状态：

```powershell
.\deploy\k8s\platform-api\scripts\migration-status.ps1 -DatabaseUrl "postgresql://platform:***@host:5432/platform?sslmode=disable"
```

## Notes

- 基础清单假设 PostgreSQL、Redis、MinIO、OpenSearch 已经存在于集群内服务名：
  - `postgresql.openclaw-system.svc.cluster.local`
  - `redis.openclaw-system.svc.cluster.local`
  - `minio.openclaw-system.svc.cluster.local`
  - `opensearch.openclaw-system.svc.cluster.local`
- `bootstrap/overlays/*` 用于单独执行数据库引导，不再和 Deployment 打包在同一个 kustomization 里。
- `base/secret.yaml` 仅供 `dev` / 基础清单使用；`prod` overlay 会删除该资源并改用 `security/prod/externalsecret.yaml`。
- `security/externalsecret.example.yaml` 提供独立示例，`security/prod/externalsecret.yaml` 是发布时会被 `set-config.ps1` 更新的真实生产引用清单。
- `.github/workflows/platform-api-ci.yml` 提供了基础 CI：Go 测试、Docker 构建、Kustomize 渲染和 dry-run。
- `.github/workflows/platform-api-release.yml` 提供了基础发布骨架：推送镜像到 GHCR、签名 digest、把非敏感配置写入 `prod` overlay，并导出渲染后的生产清单。
- `.github/workflows/platform-api-loadtest.yml` 提供了基础容器化压测 smoke。
- `prod` overlay 的非敏感配置通过 `set-config.ps1` 写入；敏感配置通过 `ExternalSecret` 指向远端 secret manager key prefix，不再写入 manifest。
- `observability/` 目录提供了可选的 `ServiceMonitor`、`PrometheusRule`、Grafana dashboard 和 AlertmanagerConfig 模板。
- 可使用 `scripts/verify-observability.ps1` 做发布后的观测对象存在性检查。
- `package-release.ps1` 会导出脱敏 manifest 和核心 runbook，便于交付归档。
- `package-release.ps1` 现在也会导出 `production-config-matrix.md`、`external-openapi.yaml`、`external-integration-guide.md`。
- `package-release.ps1` 也会导出 smoke / release / dependency / observability 校验脚本，便于交付后核验。
- `docs/runbooks/2026-04-07-platform-api-production-environment-template.md` 列出了 release workflow 需要的 GitHub Environment 变量、ExternalSecret store 信息和远端 key 路径规范。
- `docs/runbooks/2026-04-09-platform-api-production-config-matrix.md` 固化了环境 / 集群 / 地域 / 商户维度下的正式参数矩阵。



