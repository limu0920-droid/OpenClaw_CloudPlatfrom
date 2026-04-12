$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
$platformApiDir = Join-Path $repoRoot "apps\\server"
$envFile = Join-Path $scriptDir "platform-api.dev.env"
$envExample = Join-Path $scriptDir "platform-api.dev.env.example"

if (-not (Test-Path $envFile)) {
    Copy-Item -Path $envExample -Destination $envFile
}

Get-Content -Encoding utf8 $envFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) {
        return
    }
    $parts = $line -split "=", 2
    if ($parts.Count -eq 2) {
        [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1], "Process")
    }
}

Push-Location $platformApiDir
try {
    if (-not $env:BOOTSTRAP_DATA_PATH) {
        throw "BOOTSTRAP_DATA_PATH is required for cmd/bootstrap"
    }
    go run ./cmd/bootstrap
}
finally {
    Pop-Location
}
