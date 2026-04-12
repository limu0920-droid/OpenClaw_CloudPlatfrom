# Platform API Release Checklist

## 发布前

- 已在目标分支执行 `go test ./...`
- 已执行 `.temp/dev-stack/smoke-persistence.ps1 -Build`
- 已执行 `.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Strict`
- 已执行 `.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Bootstrap -Strict`
- 已生成并记录生产镜像 digest
- 已使用 `set-image.ps1` 将 digest 写入 `prod` overlay
- 已确认 GitHub Environment 中的域名、外部服务地址、SecretStore 元数据都是正式值
- 已确认 Secret Manager 中的数据库、对象存储、OpenSearch、Keycloak 密钥均已写入远端 key prefix
- 已确认 `ARTIFACT_PREVIEW_PUBLIC_BASE_URL` 指向正式 preview 子域名
- 已确认 `ARTIFACT_PREVIEW_ALLOWED_HOSTS` 只包含受信任的工作台 / 对象存储 / CDN 域名
- 已确认 TLS Secret 已存在

## 发布中

- 已执行数据库备份，且备份文件可读
- 已确认 `ExternalSecret/openclaw-platform-api-secret` Ready
- 已确认 `openclaw-platform-api-secret` 已物化且无占位符
- 先执行 `platform-api-migrate`
- 再执行 `platform-api-bootstrap`
- 再发布 `Deployment`
- 等待 `rollout status` 成功
- 执行 `smoke.ps1`
- 执行 `verify-release.ps1`
- 执行 `verify-dependencies.ps1`

## 发布后

- `/readyz` 返回 `200`
- `/versionz` 返回预期版本 / commit / builtAt
- `schema_migration` 包含最新版本
- Prometheus 可抓取 `/metrics`
- `verify-observability.ps1` 返回目标对象存在
- 关键 API：
  - `/api/v1/bootstrap`
  - `/api/v1/portal/instances`
  - `/api/v1/admin/orders`
- 关键预览链路：
  - `/api/v1/portal/workspace/artifacts/:id/preview`
  - `/api/v1/portal/workspace/artifacts/:id/preview-content`
  - `/api/v1/portal/workspace/artifacts/:id/download`
- 已验证至少 1 个 HTML 产物和 1 个 PDF/Office 衍生产物可在平台内正式预览
- 业务写入后重启 Pod 再读取，确认数据仍保留

## 回滚

- 回滚镜像 digest
- `kubectl rollout undo deployment/openclaw-platform-api -n openclaw-system`
- 如涉及 schema 问题，按数据库备份方案人工恢复


