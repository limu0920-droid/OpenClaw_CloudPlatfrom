param(
    [switch]$CheckIAM,
    [switch]$CheckAPI
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$envMap = @{}
$envFile = Join-Path $scriptDir ".env"
if (-not (Test-Path $envFile)) {
    $envFile = Join-Path $scriptDir ".env.example"
}

Get-Content -Encoding utf8 $envFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) {
        return
    }
    $parts = $line -split "=", 2
    if ($parts.Count -eq 2) {
        $envMap[$parts[0]] = $parts[1]
    }
}

function Test-HttpJson {
    param(
        [string]$Name,
        [string]$Uri
    )
    $response = Invoke-WebRequest -UseBasicParsing -Uri $Uri -TimeoutSec 10
    [pscustomobject]@{
        Name = $Name
        Uri = $Uri
        StatusCode = $response.StatusCode
    }
}

function Test-DockerExec {
    param(
        [string[]]$CommandArgs
    )
    Push-Location $scriptDir
    try {
        $dockerArgs = @("compose", "--env-file", ".env", "exec", "-T") + $CommandArgs
        & docker @dockerArgs
    }
    finally {
        Pop-Location
    }
}

function Resolve-ApiPort {
    if (-not [string]::IsNullOrWhiteSpace($envMap["PLATFORM_API_PORT"])) {
        return $envMap["PLATFORM_API_PORT"]
    }

    Push-Location $scriptDir
    try {
        $mapped = & docker compose --env-file .env port platform-api 8080 2>$null
        if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($mapped)) {
            return (($mapped | Select-Object -First 1).Trim() -split ":")[-1]
        }
    }
    finally {
        Pop-Location
    }
    return "18080"
}

$checks = @()
$checks += Test-HttpJson -Name "MinIO" -Uri ("http://localhost:{0}/minio/health/live" -f $envMap["MINIO_API_PORT"])
$checks += Test-HttpJson -Name "OpenSearch" -Uri ("http://localhost:{0}/_cluster/health" -f $envMap["OPENSEARCH_PORT"])

$tableCount = Test-DockerExec -CommandArgs @("postgres", "psql", "-U", $envMap["POSTGRES_USER"], "-d", $envMap["POSTGRES_DB"], "-t", "-c", "select count(*) from information_schema.tables where table_schema = 'platform';")
$redisPing = Test-DockerExec -CommandArgs @("redis", "redis-cli", "ping")

if ($CheckIAM) {
    $checks += Test-HttpJson -Name "Keycloak" -Uri ("http://localhost:{0}/realms/master/.well-known/openid-configuration" -f $envMap["KEYCLOAK_PORT"])
}
if ($CheckAPI) {
    $apiPort = Resolve-ApiPort
    $checks += Test-HttpJson -Name "PlatformAPI" -Uri ("http://localhost:{0}/readyz" -f $apiPort)
}

[pscustomobject]@{
    HttpChecks = $checks
    PlatformTableCount = ($tableCount | Out-String).Trim()
    RedisPing = ($redisPing | Out-String).Trim()
} | ConvertTo-Json -Depth 8
