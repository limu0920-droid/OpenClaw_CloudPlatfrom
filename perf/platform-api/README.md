# Platform API Load Test

## Local

先启动容器化开发栈：

```powershell
cd .temp/dev-stack
.\up.ps1 -WithAPI -Build -Wait
```

再执行 k6：

```powershell
k6 run perf/platform-api/k6-smoke.js
```

可指定目标地址：

```powershell
$env:BASE_URL='http://127.0.0.1:18080'
k6 run perf/platform-api/k6-smoke.js
```

或使用包装脚本导出 summary：

```powershell
.\perf\platform-api\run-k6.ps1 -BaseUrl http://127.0.0.1:18080 -SummaryOutput .\perf\platform-api\platform-api-k6-summary.json
```

校验 summary 是否满足基线：

```powershell
.\perf\platform-api\compare-k6.ps1 -Summary .\perf\platform-api\platform-api-k6-summary.json
```

将结果写成基线记录：

```powershell
.\perf\platform-api\record-baseline.ps1 `
  -Summary .\perf\platform-api\platform-api-k6-summary.json `
  -Environment dev-stack `
  -ImageDigest ghcr.io/openclaw/platform-api@sha256:... `
  -MigrationStatus "0001,0002,0003" `
  -Operator your-name `
  -Output .\perf\platform-api\platform-api-baseline-record.md
```

## Baseline

当前默认阈值：

- `http_req_failed < 1%`
- `p95 < 1000ms`

此脚本用于发布前的轻量 smoke/load baseline，不替代正式容量压测。
结果记录模板见 [baseline-template.md](./baseline-template.md)。
CI 会通过 [platform-api-loadtest.yml](../../.github/workflows/platform-api-loadtest.yml) 自动执行一次容器化 k6 smoke。


