param(
    [string]$Namespace = "openclaw-system",
    [string]$Service = "openclaw-platform-api",
    [int]$LocalPort = 18080,
    [string]$ExpectedVersion = "",
    [string]$ExpectedCommit = "",
    [string]$ExpectedStrictMode = "",
    [string]$ExpectedStateBackend = "",
    [string]$ExpectedRuntimeProvider = ""
)

$ErrorActionPreference = "Stop"

$job = Start-Job -ScriptBlock {
    param($ns, $svc, $port)
    kubectl port-forward "svc/$svc" "$port`:80" -n $ns
} -ArgumentList $Namespace, $Service, $LocalPort

try {
    Start-Sleep -Seconds 4

    $health = Invoke-RestMethod -Uri ("http://127.0.0.1:{0}/healthz" -f $LocalPort)
    $ready = Invoke-RestMethod -Uri ("http://127.0.0.1:{0}/readyz" -f $LocalPort)
    $version = Invoke-RestMethod -Uri ("http://127.0.0.1:{0}/versionz" -f $LocalPort)
    $bootstrap = Invoke-RestMethod -Uri ("http://127.0.0.1:{0}/api/v1/bootstrap" -f $LocalPort)

    if ($ExpectedVersion -ne "" -and $version.version -ne $ExpectedVersion) {
        throw "version mismatch: expected $ExpectedVersion, got $($version.version)"
    }
    if ($ExpectedCommit -ne "" -and $version.commit -ne $ExpectedCommit) {
        throw "commit mismatch: expected $ExpectedCommit, got $($version.commit)"
    }
    if ($ExpectedStrictMode -ne "" -and ([string]$version.strictMode).ToLower() -ne $ExpectedStrictMode.ToLower()) {
        throw "strictMode mismatch: expected $ExpectedStrictMode, got $($version.strictMode)"
    }
    if ($ExpectedStateBackend -ne "" -and $version.stateBackend -ne $ExpectedStateBackend) {
        throw "stateBackend mismatch: expected $ExpectedStateBackend, got $($version.stateBackend)"
    }
    if ($ExpectedRuntimeProvider -ne "" -and $version.runtimeProvider -ne $ExpectedRuntimeProvider) {
        throw "runtimeProvider mismatch: expected $ExpectedRuntimeProvider, got $($version.runtimeProvider)"
    }

    [pscustomobject]@{
        HealthStatus = $health.status
        ReadyStatus = $ready.status
        Version = $version.version
        Commit = $version.commit
        BuiltAt = $version.builtAt
        StrictMode = $version.strictMode
        StateBackend = $version.stateBackend
        RuntimeProvider = $version.runtimeProvider
        BootstrapMode = $bootstrap.app.mode
        BootstrapStateBackend = $bootstrap.app.stateBackend
        BootstrapRuntimeProvider = $bootstrap.app.runtimeProvider
    } | ConvertTo-Json -Depth 6
}
finally {
    Stop-Job $job -ErrorAction SilentlyContinue | Out-Null
    Remove-Job $job -Force -ErrorAction SilentlyContinue | Out-Null
}

