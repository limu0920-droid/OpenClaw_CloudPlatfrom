param(
    [string]$WorkspaceBaseUrl = "http://127.0.0.1:8080",
    [string]$BridgePath = "/api/v1/platform/workspace/messages",
    [string]$BridgeHealthPath = "/api/v1/platform/workspace/health",
    [string]$BridgeToken = "",
    [switch]$Build
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

function Set-LocalBridgeEnv {
    param(
        [string]$TargetFile
    )

    $content = Get-Content -Path $TargetFile -Raw -Encoding utf8
    $content = [regex]::Replace($content, '^WORKSPACE_BRIDGE_PATH=.*$', "WORKSPACE_BRIDGE_PATH=$BridgePath", 'Multiline')
    $content = [regex]::Replace($content, '^WORKSPACE_BRIDGE_HEALTH_PATH=.*$', "WORKSPACE_BRIDGE_HEALTH_PATH=$BridgeHealthPath", 'Multiline')
    $content = [regex]::Replace($content, '^WORKSPACE_BRIDGE_TOKEN=.*$', "WORKSPACE_BRIDGE_TOKEN=$BridgeToken", 'Multiline')
    Set-Content -Path $TargetFile -Value $content -Encoding utf8
}

function Join-Url {
    param(
        [string]$BaseUrl,
        [string]$Path
    )

    return $BaseUrl.TrimEnd('/') + '/' + $Path.TrimStart('/')
}

Push-Location $scriptDir
try {
    if ($Build) {
        Set-ExecutionPolicy -Scope Process Bypass
        .\up.ps1 -WithAPI -Build -Wait
    } else {
        .\up.ps1 -WithAPI -Wait
    }

    Set-LocalBridgeEnv -TargetFile (Join-Path $scriptDir 'platform-api.container.env')

    $healthUrl = Join-Url -BaseUrl $WorkspaceBaseUrl -Path $BridgeHealthPath
    $healthResponse = Invoke-WebRequest -UseBasicParsing -Uri $healthUrl -TimeoutSec 10

    $session = Invoke-RestMethod -Method Post -Uri http://127.0.0.1:18080/api/v1/portal/instances/100/workspace/sessions -ContentType 'application/json' -Body (@{
        title = '本地龙虾桥接联调'
        workspaceUrl = $WorkspaceBaseUrl
    } | ConvertTo-Json)

    $messageHeaders = @{}
    if ($BridgeToken -ne '') {
        $messageHeaders['X-Platform-Bridge-Token'] = $BridgeToken
    }

    $message = Invoke-RestMethod -Method Post -Uri ("http://127.0.0.1:18080/api/v1/portal/workspace/sessions/{0}/messages" -f $session.session.id) -Headers $messageHeaders -ContentType 'application/json' -Body (@{
        role = 'user'
        content = '请返回一个网页或文档产物链接用于平台联调。'
        dispatch = $true
    } | ConvertTo-Json)

    $detail = Invoke-RestMethod -Uri ("http://127.0.0.1:18080/api/v1/portal/workspace/sessions/{0}" -f $session.session.id)

    [pscustomobject]@{
        BridgeHealthStatus = $healthResponse.StatusCode
        SessionId = $session.session.id
        DispatchOk = $message.dispatch.ok
        DispatchTarget = $message.dispatch.target
        DispatchError = $message.dispatch.error
        MessageStatus = $message.message.status
        ArtifactCount = @($detail.artifacts).Count
    } | ConvertTo-Json -Depth 8
}
finally {
    try {
        .\down.ps1 -RemoveVolumes | Out-Null
    }
    catch {
    }
    Pop-Location
}
