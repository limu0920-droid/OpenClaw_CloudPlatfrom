param(
    [Parameter(Mandatory = $true)]
    [string]$DatabaseUrl
)

$ErrorActionPreference = "Stop"
$env:DATABASE_URL = $DatabaseUrl

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $scriptDir))
$platformApiDir = Join-Path $repoRoot "apps\\platform-api"

Push-Location $platformApiDir
try {
    go run ./cmd/migrationstatus
}
finally {
    Pop-Location
}
