param()

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$databaseUrl = "postgresql://platform:platform-dev-password@localhost:15432/platform?sslmode=disable"

Push-Location $scriptDir
try {
    .\up.ps1 -Wait

    Push-Location (Join-Path (Split-Path -Parent (Split-Path -Parent $scriptDir)) "apps\server")
    try {
        $env:DATABASE_URL = $databaseUrl
        if (-not $env:BOOTSTRAP_DATA_PATH) {
            throw "BOOTSTRAP_DATA_PATH is required for cmd/bootstrap"
        }
        go run ./cmd/migrate | Out-Null
        go run ./cmd/bootstrap | Out-Null
        $status = go run ./cmd/migrationstatus
    }
    finally {
        Pop-Location
    }

    [pscustomobject]@{
        MigrationStatus = ($status | Out-String).Trim()
    } | ConvertTo-Json -Depth 6
}
finally {
    try {
        .\down.ps1 -RemoveVolumes | Out-Null
    }
    catch {
    }
    Pop-Location
}

