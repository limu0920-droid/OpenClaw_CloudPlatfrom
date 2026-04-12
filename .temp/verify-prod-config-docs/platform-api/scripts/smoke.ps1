param(
    [string]$Namespace = "openclaw-system",
    [string]$Service = "openclaw-platform-api",
    [int]$LocalPort = 18080
)

$ErrorActionPreference = "Stop"

$job = Start-Job -ScriptBlock {
    param($ns, $svc, $port)
    kubectl port-forward "svc/$svc" "$port`:80" -n $ns
} -ArgumentList $Namespace, $Service, $LocalPort

try {
    Start-Sleep -Seconds 4
    $health = Invoke-WebRequest -UseBasicParsing -Uri ("http://127.0.0.1:{0}/healthz" -f $LocalPort) -TimeoutSec 10
    $ready = Invoke-WebRequest -UseBasicParsing -Uri ("http://127.0.0.1:{0}/readyz" -f $LocalPort) -TimeoutSec 10
    [pscustomobject]@{
        Healthz = $health.StatusCode
        Readyz  = $ready.StatusCode
    } | ConvertTo-Json -Depth 4
}
finally {
    Stop-Job $job -ErrorAction SilentlyContinue | Out-Null
    Remove-Job $job -Force -ErrorAction SilentlyContinue | Out-Null
}

