# Platform API Operations Runbook

## 目标

本手册用于当前 `platform-api` 包的部署、迁移、引导、发布、回滚和常见故障排查。

适用对象：

- 平台后端开发
- 运维 / SRE
- 交付与联调人员

## 当前交付范围

当前包已包含：

- `platform-api` 容器镜像
- `platform-api-bootstrap` 数据引导命令
- `platform-api-migrate` 数据迁移命令
- `.temp/dev-stack` 本地联调栈
- `deploy/k8s/platform-api` Kubernetes 清单
- CI / Release workflow 骨架

当前已做持久化的主要域：

- 租户、用户、认证标识
- 实例、访问入口、配置、运行态、凭据、备份、任务、审计
- 渠道、渠道活动
- 账户设置、钱包、账单
- 工单、告警
- OEM 品牌、主题、功能开关、租户绑定
- 订单、订阅、支付、退款、发票、支付回调事件

## 上线前准备

上线前至少确认：

- 生产数据库已创建，`DATABASE_URL` 可用。
- PostgreSQL、Redis、对象存储、OpenSearch 已可达。
- 生产密钥已经写入 Secret Manager，`ExternalSecret/openclaw-platform-api-secret` 可同步到目标集群。
- 生产 overlay 已替换为真实镜像 digest、域名、外部服务地址。
- Ingress TLS 证书 Secret 已存在。
- `AUTO_MIGRATE=false`、`AUTO_BOOTSTRAP=false`，由独立 Job 执行迁移和引导。
- 任何带 `DATABASE_URL` 的直接服务启动都不应依赖默认自动 bootstrap。

## 本地验证

启动依赖与容器化 API：

```powershell
cd .temp/dev-stack
.\up.ps1 -WithAPI -Build -Wait
.\smoke.ps1 -CheckAPI
.\smoke-persistence.ps1 -Build
```

停止环境：

```powershell
.\down.ps1 -RemoveVolumes
```

## 数据迁移

只执行迁移：

```powershell
$env:DATABASE_URL='postgresql://platform:password@host:5432/platform?sslmode=disable'
go run ./apps/platform-api/cmd/migrate
```

执行迁移并引导种子：

```powershell
$env:DATABASE_URL='postgresql://platform:password@host:5432/platform?sslmode=disable'
go run ./apps/platform-api/cmd/bootstrap
```

查看 migration 状态：

```powershell
$env:DATABASE_URL='postgresql://platform:password@host:5432/platform?sslmode=disable'
go run ./apps/platform-api/cmd/migrationstatus
```

容器镜像内等价命令：

```bash
/usr/local/bin/platform-api-migrate
/usr/local/bin/platform-api-bootstrap
```

## Kubernetes 发布步骤

1. 渲染并校验清单：

```powershell
.\deploy\k8s\platform-api\scripts\validate.ps1 -Overlay prod
```

2. 将固定镜像 digest 写入生产 overlay：

```powershell
.\deploy\k8s\platform-api\scripts\set-image.ps1 -Overlay prod -Image ghcr.io/openclaw/platform-api@sha256:...
```

3. 执行生产配置注入与严格预检：

```powershell
.\deploy\k8s\platform-api\scripts\set-config.ps1 -Overlay prod ...
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Strict
.\deploy\k8s\platform-api\scripts\preflight.ps1 -Overlay prod -Bootstrap -Strict
```

4. 先执行迁移 / 引导：

方式一：
使用 `bootstrap-job.yaml`

方式二：
单独执行

```powershell
kubectl apply -f rendered-bootstrap-job.yaml
kubectl logs job/openclaw-platform-api-bootstrap -n openclaw-system
```

在等待 Job 完成前，先确认：

```powershell
.\deploy\k8s\platform-api\scripts\verify-dependencies.ps1 `
  -Namespace openclaw-system `
  -DatabaseSecretName openclaw-platform-api-secret `
  -ExternalSecretName openclaw-platform-api-secret
```

5. 发布 Deployment：

```powershell
.\deploy\k8s\platform-api\scripts\apply.ps1 -Overlay prod
```

6. 校验就绪：

```powershell
kubectl rollout status deployment/openclaw-platform-api -n openclaw-system
kubectl get pods -n openclaw-system
kubectl port-forward svc/openclaw-platform-api 18080:80 -n openclaw-system
curl http://127.0.0.1:18080/readyz
```

## 发布后检查

至少确认：

- `/healthz` 返回 `200`
- `/readyz` 返回 `200`
- `schema_migration` 中包含最新版本
- `bootstrap` Job 成功
- API 日志里没有 migration/bootstrap error
- 实例列表、订单列表、渠道列表能正常返回

推荐检查：

- 新建一个测试订单
- 更新一次 OEM 主题
- 创建一张工单
- 重启 Deployment 后再次读取，确认数据仍在
- `verify-dependencies.ps1` 确认 `ExternalSecret` Ready 且目标 Secret 无占位符

## 回滚策略

镜像回滚：

```powershell
.\deploy\k8s\platform-api\scripts\set-image.ps1 -Overlay prod -Image ghcr.io/openclaw/platform-api@sha256:OLD_DIGEST
.\deploy\k8s\platform-api\scripts\apply.ps1 -Overlay prod
```

Deployment 回滚：

```powershell
kubectl rollout undo deployment/openclaw-platform-api -n openclaw-system
```

注意：

- 当前 migration 只支持前向迁移，没有自动 down migration。
- 如某次 migration 引入不可逆 schema 变更，需要数据库备份和人工回滚方案。

## 常见故障

### 1. API 容器启动即退出

优先检查：

- `DATABASE_URL` 是否正确
- `schema_migration` 是否损坏
- 最新 migration 是否幂等

命令：

```powershell
kubectl logs deploy/openclaw-platform-api -n openclaw-system
kubectl logs job/openclaw-platform-api-bootstrap -n openclaw-system
```

### 2. `/readyz` 失败

可能原因：

- PostgreSQL 不可达
- OpenSearch 不可达
- Secret / ConfigMap 配置错误

优先检查：

- `kubectl describe pod`
- 环境变量注入
- 网络策略是否阻断

### 3. 重启后数据丢失

优先确认：

- 是否真的走了 `DATABASE_URL`
- 是否启用了 `AUTO_BOOTSTRAP=true` 导致种子覆盖预期
- 相关接口是否已接持久化

建议先跑：

```powershell
go test ./apps/platform-api/...
.\.temp\dev-stack\smoke-persistence.ps1 -Build
```

### 4. K8s overlay 渲染失败

检查：

- patch 文件是否和 base 资源名称一致
- 是否误改了 `kind` / `metadata.name`

命令：

```powershell
kubectl kustomize deploy/k8s/platform-api/overlays/prod
kubectl apply --dry-run=client --validate=false -k deploy/k8s/platform-api/overlays/prod
```

## 当前已知限制

- 还没有正式的 down migration / schema rollback 体系
- 生产监控、告警规则、压测和容量基线还需继续补齐
- CI / Release 已支持 GitHub Environment + ExternalSecret 的生产配置注入，但仍需接入真实部署审批与容量基线



