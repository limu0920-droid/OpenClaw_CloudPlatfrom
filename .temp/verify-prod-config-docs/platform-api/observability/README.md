# Observability Assets

此目录包含可选监控资产：

- `servicemonitor.yaml`
- `prometheusrule.yaml`
- `grafana-dashboard.json`
- `grafana-dashboard-configmap.yaml`
- `alertmanagerconfig.example.yaml`

说明：

- `ServiceMonitor` / `PrometheusRule` / `AlertmanagerConfig` 适用于 Prometheus Operator。
- `grafana-dashboard-configmap.yaml` 适用于 Grafana sidecar 自动加载 dashboard 的部署方式。
- 这些文件默认不纳入 base kustomization，需要按实际集群能力单独应用。
