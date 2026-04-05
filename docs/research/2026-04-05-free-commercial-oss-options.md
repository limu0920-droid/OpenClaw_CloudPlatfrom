# 免费商用开源候选

## 结论先行

当前最适合这个项目优先接入的免费商用开源栈：

- `Keycloak`
- `Headlamp`
- `Prometheus + Alertmanager`
- `OpenSearch`
- `Chatwoot Community`
- `Appsmith Community`

需要谨慎评估许可证风险的候选：

- `Zammad`
- `ZITADEL v3+`
- `Grafana`

## 推荐栈

### 1. Keycloak

- 用途：SSO、OAuth2、OIDC、角色、租户身份
- 许可证：`Apache-2.0`
- 官网：https://www.keycloak.org/

### 2. Headlamp

- 用途：Kubernetes 可视化与控制面
- 许可证：`Apache-2.0`
- 官网：https://headlamp.dev/

### 3. Prometheus + Alertmanager

- 用途：资源监控、告警、API 调用量
- 许可证：`Apache-2.0`
- 官网：https://prometheus.io/

### 4. OpenSearch

- 用途：日志、审计、检索
- 许可证：`Apache-2.0`
- 官网：https://opensearch.org/

### 5. Chatwoot Community

- 用途：工单 / 客服 / 问题上报
- 许可证：`MIT`
- 官方仓库：https://github.com/chatwoot/chatwoot

### 6. Appsmith Community

- 用途：内部运营后台 / 工具台
- 许可证：`Apache-2.0`
- 官方仓库：https://github.com/appsmithorg/appsmith

## 风险项

### Zammad

- 工单能力强
- 但许可证是 `AGPLv3`

### ZITADEL v3+

- 身份能力强
- 但 `v3` 起是 `AGPLv3`

### Grafana

- 生态成熟
- 但核心项目已转 `AGPLv3`

## 当前建议

1. 身份直接接 `Keycloak`
2. K8s 控制面优先接 `Headlamp`
3. 监控优先接 `Prometheus + Alertmanager`
4. 检索优先接 `OpenSearch`
5. 工单先自研轻量流程，再评估是否引入 `Chatwoot`
