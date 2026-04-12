param(
    [string]$KubeContext = "",
    [string]$RuntimeImage = "nginx:1.27-alpine",
    [switch]$SkipBootstrap,
    [switch]$KeepApiRunning
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
$platformApiDir = Join-Path $repoRoot "apps\server"
$baseEnvFile = Join-Path $scriptDir "platform-api.dev.env"
$baseEnvExample = Join-Path $scriptDir "platform-api.dev.env.example"
$smokeEnvFile = Join-Path $scriptDir "platform-api.k8s-smoke.env"
$stdoutLog = Join-Path $scriptDir "platform-api-k8s-smoke.out.log"
$stderrLog = Join-Path $scriptDir "platform-api-k8s-smoke.err.log"

function Read-EnvFile {
    param([string]$Path)

    $map = [ordered]@{}
    foreach ($line in Get-Content -Encoding utf8 $Path) {
        $trimmed = $line.Trim()
        if ($trimmed -eq "" -or $trimmed.StartsWith("#")) {
            continue
        }
        $parts = $trimmed -split "=", 2
        if ($parts.Count -eq 2) {
            $map[$parts[0]] = $parts[1]
        }
    }
    return $map
}

function Write-EnvFile {
    param(
        [string]$Path,
        [System.Collections.Specialized.OrderedDictionary]$Map
    )

    $lines = foreach ($entry in $Map.GetEnumerator()) {
        "{0}={1}" -f $entry.Key, $entry.Value
    }
    Set-Content -Encoding utf8 -Path $Path -Value $lines
}

function Wait-HttpReady {
    param(
        [string]$Uri,
        [int]$TimeoutSec = 90
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri $Uri -TimeoutSec 5
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 300) {
                return
            }
        }
        catch {
        }
        Start-Sleep -Seconds 2
    }

    $stdout = if (Test-Path $stdoutLog) { Get-Content -Encoding utf8 $stdoutLog -Tail 80 | Out-String } else { "" }
    $stderr = if (Test-Path $stderrLog) { Get-Content -Encoding utf8 $stderrLog -Tail 80 | Out-String } else { "" }
    throw "platform api was not ready at $Uri`nSTDOUT:`n$stdout`nSTDERR:`n$stderr"
}

function Wait-DeploymentReplicas {
    param(
        [string]$Namespace,
        [string]$Name,
        [int]$DesiredReplicas,
        [int]$TimeoutSec = 120
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        try {
            $json = kubectl get deployment $Name -n $Namespace -o json | ConvertFrom-Json
            $specReplicas = if ($null -ne $json.spec.replicas) { [int]$json.spec.replicas } else { 1 }
            $currentReplicas = if ($null -ne $json.status.replicas) { [int]$json.status.replicas } else { 0 }
            $readyReplicas = if ($null -ne $json.status.readyReplicas) { [int]$json.status.readyReplicas } else { 0 }

            if ($DesiredReplicas -eq 0) {
                if ($specReplicas -eq 0 -and $currentReplicas -eq 0 -and $readyReplicas -eq 0) {
                    return
                }
            }
            elseif ($specReplicas -eq $DesiredReplicas -and $currentReplicas -ge $DesiredReplicas) {
                return
            }
        }
        catch {
        }
        Start-Sleep -Seconds 2
    }

    throw "deployment $Namespace/$Name did not reach replicas=$DesiredReplicas in time"
}

function Stop-ProcessTree {
    param([int]$ProcessId)

    $children = Get-CimInstance Win32_Process -Filter ("ParentProcessId = {0}" -f $ProcessId) -ErrorAction SilentlyContinue
    foreach ($child in $children) {
        Stop-ProcessTree -ProcessId $child.ProcessId
    }

    Stop-Process -Id $ProcessId -Force -ErrorAction SilentlyContinue
}

function Wait-NamespaceDeleted {
    param(
        [string]$Namespace,
        [int]$TimeoutSec = 120
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        $result = & kubectl get namespace $Namespace --ignore-not-found -o name 2>$null
        if ([string]::IsNullOrWhiteSpace(($result | Out-String).Trim())) {
            return
        }
        Start-Sleep -Seconds 2
    }

    throw "namespace $Namespace was not deleted in time"
}

if (-not (Test-Path $baseEnvFile)) {
    Copy-Item -Path $baseEnvExample -Destination $baseEnvFile
}

$apiProcess = $null
$instanceId = 0
$namespace = ""

Push-Location $scriptDir
try {
    .\up.ps1 -Wait

    if (-not $SkipBootstrap) {
        .\bootstrap-db.ps1
    }

    $envMap = Read-EnvFile -Path $baseEnvFile
    $envMap["PLATFORM_STRICT_MODE"] = "true"
    $envMap["RUNTIME_PROVIDER"] = "kubectl"
    $envMap["RUNTIME_IMAGE"] = $RuntimeImage
    if (-not [string]::IsNullOrWhiteSpace($KubeContext)) {
        $envMap["RUNTIME_KUBE_CONTEXT"] = $KubeContext
    }
    Write-EnvFile -Path $smokeEnvFile -Map $envMap

    Remove-Item -Force $stdoutLog, $stderrLog -ErrorAction SilentlyContinue
    $apiProcess = Start-Process -FilePath "powershell" -ArgumentList @(
        "-NoProfile",
        "-ExecutionPolicy", "Bypass",
        "-File", (Join-Path $scriptDir "start-platform-api.ps1"),
        "-EnvFile", ".\platform-api.k8s-smoke.env"
    ) -WorkingDirectory $scriptDir -PassThru -RedirectStandardOutput $stdoutLog -RedirectStandardError $stderrLog

    $apiPort = if ($envMap.Contains("PORT")) { $envMap["PORT"] } else { "18080" }
    Wait-HttpReady -Uri ("http://127.0.0.1:{0}/readyz" -f $apiPort)

    $bootstrap = Invoke-RestMethod -Uri ("http://127.0.0.1:{0}/api/v1/bootstrap" -f $apiPort)
    if ($bootstrap.app.runtimeProvider -ne "kubectl") {
        throw "expected runtimeProvider=kubectl, got $($bootstrap.app.runtimeProvider)"
    }
    if ($bootstrap.app.stateBackend -ne "postgres") {
        throw "expected stateBackend=postgres, got $($bootstrap.app.stateBackend)"
    }

    $suffix = Get-Date -Format "HHmmss"
    $created = Invoke-RestMethod -Method Post -Uri ("http://127.0.0.1:{0}/api/v1/portal/instances" -f $apiPort) -ContentType "application/json" -Body (@{
        name   = "k8s-smoke-$suffix"
        plan   = "trial"
        region = "cn-shanghai"
        cpu    = "1"
        memory = "1Gi"
    } | ConvertTo-Json)

    $instanceId = [int]$created.instance.id
    $namespace = [string]$created.binding.namespace
    $deploymentName = [string]$created.binding.workloadName
    $workloadId = [string]$created.binding.workloadId

    if ($instanceId -le 0 -or [string]::IsNullOrWhiteSpace($namespace) -or [string]::IsNullOrWhiteSpace($deploymentName)) {
        throw "instance creation did not return runtime binding: $(ConvertTo-Json $created -Depth 8)"
    }

    Wait-DeploymentReplicas -Namespace $namespace -Name $deploymentName -DesiredReplicas 1
    $service = kubectl get service $deploymentName -n $namespace -o json | ConvertFrom-Json
    $pod = kubectl get pod -n $namespace -l ("openclaw.io/workload-id={0}" -f $workloadId) -o json | ConvertFrom-Json
    $runtime = Invoke-RestMethod -Uri ("http://127.0.0.1:{0}/api/v1/portal/instances/{1}/runtime" -f $apiPort, $instanceId)
    if ($runtime.binding.workloadId -ne $workloadId) {
        throw "runtime endpoint workloadId mismatch: expected $workloadId got $($runtime.binding.workloadId)"
    }

    $stopped = Invoke-RestMethod -Method Post -Uri ("http://127.0.0.1:{0}/api/v1/portal/instances/{1}/stop" -f $apiPort, $instanceId)
    Wait-DeploymentReplicas -Namespace $namespace -Name $deploymentName -DesiredReplicas 0

    $started = Invoke-RestMethod -Method Post -Uri ("http://127.0.0.1:{0}/api/v1/portal/instances/{1}/start" -f $apiPort, $instanceId)
    Wait-DeploymentReplicas -Namespace $namespace -Name $deploymentName -DesiredReplicas 1

    $deleted = Invoke-RestMethod -Method Delete -Uri ("http://127.0.0.1:{0}/api/v1/portal/instances/{1}" -f $apiPort, $instanceId)
    Wait-NamespaceDeleted -Namespace $namespace

    [pscustomobject]@{
        Bootstrap = [pscustomobject]@{
            Mode = $bootstrap.app.mode
            StateBackend = $bootstrap.app.stateBackend
            RuntimeProvider = $bootstrap.app.runtimeProvider
        }
        Instance = [pscustomobject]@{
            Id = $instanceId
            Namespace = $namespace
            WorkloadId = $workloadId
            Deployment = $deploymentName
        }
        Service = [pscustomobject]@{
            Name = $service.metadata.name
            Type = $service.spec.type
            NodePort = if ($service.spec.ports.Count -gt 0) { $service.spec.ports[0].nodePort } else { $null }
        }
        Pod = [pscustomobject]@{
            Name = if ($pod.items.Count -gt 0) { $pod.items[0].metadata.name } else { $null }
            Phase = if ($pod.items.Count -gt 0) { $pod.items[0].status.phase } else { $null }
            WaitingReason = if ($pod.items.Count -gt 0 -and $pod.items[0].status.containerStatuses.Count -gt 0) { $pod.items[0].status.containerStatuses[0].state.waiting.reason } else { $null }
        }
        Actions = [pscustomobject]@{
            StopStatus = $stopped.job.status
            StartStatus = $started.job.status
            DeleteStatus = $deleted.job.status
        }
    } | ConvertTo-Json -Depth 8
}
finally {
    if (-not [string]::IsNullOrWhiteSpace($namespace)) {
        try {
            kubectl delete namespace $namespace --wait=false --ignore-not-found=true | Out-Null
        }
        catch {
        }
    }
    if ($apiProcess -and -not $KeepApiRunning) {
        try {
            Stop-ProcessTree -ProcessId $apiProcess.Id
        }
        catch {
        }
    }
    Pop-Location
}

