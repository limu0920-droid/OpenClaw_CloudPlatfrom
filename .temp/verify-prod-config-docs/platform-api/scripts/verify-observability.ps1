param(
    [string]$Namespace = "openclaw-system"
)

$ErrorActionPreference = "Stop"

$checks = [ordered]@{}

try {
    $checks.ServiceMonitor = (kubectl get servicemonitor openclaw-platform-api -n $Namespace -o name 2>$null)
} catch {
    $checks.ServiceMonitor = ""
}

try {
    $checks.PrometheusRule = (kubectl get prometheusrule openclaw-platform-api -n $Namespace -o name 2>$null)
} catch {
    $checks.PrometheusRule = ""
}

try {
    $checks.GrafanaDashboardConfigMap = (kubectl get configmap openclaw-platform-api-dashboard -n $Namespace -o name 2>$null)
} catch {
    $checks.GrafanaDashboardConfigMap = ""
}

[pscustomobject]$checks | ConvertTo-Json -Depth 4

