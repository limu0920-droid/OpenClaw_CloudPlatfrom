param(
    [switch]$RemoveVolumes
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$envFile = ".env"
if (-not (Test-Path (Join-Path $scriptDir $envFile))) {
    $envFile = ".env.example"
}

$args = @(
    "--env-file", $envFile,
    "--profile", "api",
    "--profile", "iam",
    "--profile", "search-ui",
    "--profile", "mail",
    "down",
    "--remove-orphans"
)
if ($RemoveVolumes) {
    $args += "--volumes"
}

Push-Location $scriptDir
try {
    & docker compose @args
}
finally {
    Pop-Location
}
