param(
    [switch]$WithIAM,
    [switch]$WithSearchUI,
    [switch]$WithMail,
    [switch]$WithAPI,
    [switch]$Build,
    [switch]$Wait
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$envFile = Join-Path $scriptDir ".env"
$envExample = Join-Path $scriptDir ".env.example"
$apiEnvFile = Join-Path $scriptDir "platform-api.container.env"
$apiEnvExample = Join-Path $scriptDir "platform-api.container.env.example"

if (-not (Test-Path $envFile)) {
    Copy-Item -Path $envExample -Destination $envFile
}
if (-not (Test-Path $apiEnvFile)) {
    Copy-Item -Path $apiEnvExample -Destination $apiEnvFile
}

function Sync-EnvDefaults {
    param(
        [string]$TargetFile,
        [string]$ExampleFile
    )

    $targetLines = Get-Content -Encoding utf8 $TargetFile
    $existingKeys = @{}
    foreach ($line in $targetLines) {
        $trimmed = $line.Trim()
        if ($trimmed -eq "" -or $trimmed.StartsWith("#")) {
            continue
        }
        $parts = $trimmed -split "=", 2
        if ($parts.Count -eq 2) {
            $existingKeys[$parts[0]] = $true
        }
    }

    $missing = @()
    foreach ($line in Get-Content -Encoding utf8 $ExampleFile) {
        $trimmed = $line.Trim()
        if ($trimmed -eq "" -or $trimmed.StartsWith("#")) {
            continue
        }
        $parts = $trimmed -split "=", 2
        if ($parts.Count -eq 2 -and -not $existingKeys.ContainsKey($parts[0])) {
            $missing += $line
        }
    }

    if ($missing.Count -gt 0) {
        Add-Content -Encoding utf8 -Path $TargetFile -Value ""
        Add-Content -Encoding utf8 -Path $TargetFile -Value $missing
    }
}

Sync-EnvDefaults -TargetFile $envFile -ExampleFile $envExample
Sync-EnvDefaults -TargetFile $apiEnvFile -ExampleFile $apiEnvExample

$dataDirs = @(
    "data",
    "data/postgres",
    "data/redis",
    "data/minio",
    "data/opensearch",
    "data/keycloak-db"
)

foreach ($dir in $dataDirs) {
    $target = Join-Path $scriptDir $dir
    if (-not (Test-Path $target)) {
        New-Item -ItemType Directory -Path $target | Out-Null
    }
}

$args = @("--env-file", ".env")
if ($WithIAM) {
    $args += @("--profile", "iam")
}
if ($WithSearchUI) {
    $args += @("--profile", "search-ui")
}
if ($WithMail) {
    $args += @("--profile", "mail")
}
$args += @("up", "-d")
if ($Build) {
    $args += "--build"
}
if ($Wait) {
    $args += "--wait"
}

Push-Location $scriptDir
try {
    & docker compose @args

    if ($WithAPI) {
        $apiArgs = @("--env-file", ".env", "--profile", "api", "up", "-d")
        if ($Build) {
            $apiArgs += "--build"
        }
        $apiArgs += "platform-api-bootstrap"
        & docker compose @apiArgs

        $bootstrapDone = $false
        for ($i = 0; $i -lt 60; $i++) {
            $state = & docker compose --env-file .env ps --format json platform-api-bootstrap | ConvertFrom-Json
            if ($null -ne $state) {
                if ($state.ExitCode -eq 0) {
                    $bootstrapDone = $true
                    break
                }
                if ($state.State -eq "exited" -and $state.ExitCode -ne 0) {
                    throw "platform-api-bootstrap failed with exit code $($state.ExitCode)"
                }
            }
            Start-Sleep -Seconds 2
        }
        if (-not $bootstrapDone) {
            throw "platform-api-bootstrap did not complete in time"
        }

        $apiStartArgs = @("--env-file", ".env", "--profile", "api", "up", "-d")
        if ($Build) {
            $apiStartArgs += "--build"
        }
        if ($Wait) {
            $apiStartArgs += "--wait"
        }
        $apiStartArgs += "platform-api"
        & docker compose @apiStartArgs
    }
}
finally {
    Pop-Location
}
