# Dev Stack

这个目录用于本地联调，不是生产交付目录。

## Included Services

- `postgres`：加载 `docs/OpenClaw-数据库DDL初稿.sql`
- `redis`
- `minio` + `minio-init`
- `opensearch`
- `opensearch-dashboards`：可选 profile `search-ui`
- `mailpit`：可选 profile `mail`
- `keycloak-db` + `keycloak`：可选 profile `iam`

## First Run

```powershell
cd .temp/dev-stack
.\up.ps1 -Wait
.\smoke.ps1
```

如果需要 Keycloak / OpenSearch Dashboards / Mailpit：

```powershell
.\up.ps1 -WithIAM -WithSearchUI -WithMail -Wait
.\smoke.ps1 -CheckIAM
```

如果需要直接把 `platform-api` 也容器化启动：

```powershell
.\up.ps1 -WithAPI -Build -Wait
.\smoke.ps1 -CheckAPI
```

如果要验证“写入数据库后重启 API 仍保留”：

```powershell
.\smoke-persistence.ps1 -Build
```

如果要验证“宿主机 API + 本机 Kubernetes Runtime Adapter”：

```powershell
.\smoke-k8s-runtime.ps1
```

如果要验证“容器化 platform-api + 本地龙虾实例桥接”：

```powershell
.\smoke-workspace-bridge.ps1 -WorkspaceBaseUrl http://127.0.0.1:8080
```

如果本地龙虾实例桥接接口路径不同，可覆盖：

```powershell
.\smoke-workspace-bridge.ps1 `
  -WorkspaceBaseUrl http://127.0.0.1:8080 `
  -BridgePath /api/v1/platform/workspace/messages `
  -BridgeHealthPath /api/v1/platform/workspace/health
```

默认会：

- 启动本地 dev stack
- 执行数据库 bootstrap
- 以宿主机 `go run ./cmd/server` 启动平台 API
- 使用 `RUNTIME_PROVIDER=kubectl` 和当前 `kubectl` context
- 创建实例并验证对应 Deployment / Service / Namespace
- 验证实例启停和删除链路

如果要验证迁移 / 引导链路：

```powershell
.\smoke-migrations.ps1
```

## Run Platform API Against Local Dev Stack

```powershell
cd .temp/dev-stack
Copy-Item .\platform-api.dev.env.example .\platform-api.dev.env
.\bootstrap-db.ps1
.\start-platform-api.ps1
```

默认会把：

- `OPENSEARCH_URL` 指向 `http://localhost:19200`
- `RUNTIME_PROVIDER` 设为 `kubectl`
- `DATABASE_URL` / `REDIS_URL` / `OBJECT_STORAGE_*` 设到本地 docker 服务
- `BOOTSTRAP_DATA_PATH` 需要显式指向真实 bootstrap 数据文件

当前 `platform-api` 已经会把实例核心域、账户设置、渠道、OEM、工单、订单/支付等多个域写入 PostgreSQL。
当前容器化 `platform-api` 不再提供 mock runtime fallback；如需联调真实本地 Kubernetes，仍建议使用 [start-platform-api.ps1](./start-platform-api.ps1) 在宿主机启动。

## Migration

数据库迁移命令：

```powershell
$env:DATABASE_URL='postgresql://platform:platform-dev-password@localhost:15432/platform?sslmode=disable'
go run ../../apps/server/cmd/migrate
```

数据库引导命令：

```powershell
$env:DATABASE_URL='postgresql://platform:platform-dev-password@localhost:15432/platform?sslmode=disable'
$env:BOOTSTRAP_DATA_PATH='C:\path\to\bootstrap-data.json'
go run ../../apps/server/cmd/bootstrap
```

## Stop

```powershell
.\down.ps1
```

如需同时删卷：

```powershell
.\down.ps1 -RemoveVolumes
```


