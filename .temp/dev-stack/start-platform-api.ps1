param(
    [string]$EnvFile = ".\\platform-api.dev.env"
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$targetEnvFile = Join-Path $scriptDir $EnvFile
$exampleEnvFile = Join-Path $scriptDir "platform-api.dev.env.example"

if (-not (Test-Path $targetEnvFile)) {
    Copy-Item -Path $exampleEnvFile -Destination $targetEnvFile
}

Get-Content -Encoding utf8 $targetEnvFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) {
        return
    }
    $parts = $line -split "=", 2
    if ($parts.Count -eq 2) {
        [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1], "Process")
    }
}

$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
$platformApiDir = Join-Path $repoRoot "apps\\server"

Push-Location $platformApiDir
try {
    if ($env:DATABASE_URL) {
        if (-not $env:BOOTSTRAP_DATA_PATH) {
            throw "BOOTSTRAP_DATA_PATH is required before running cmd/bootstrap"
        }
        go run ./cmd/bootstrap
    }
    go run ./cmd/server
}
finally {
    Pop-Location
}
