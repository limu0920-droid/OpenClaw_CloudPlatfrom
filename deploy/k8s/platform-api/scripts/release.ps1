param(
    [string]$Overlay = "prod",
    [string]$Namespace = "openclaw-system",
    [string]$Deployment = "openclaw-platform-api",
    [string]$BootstrapOverlay = "prod",
    [string]$Image = "",
    [string]$BackupPostgresPod = "",
    [string]$BackupOutput = "",
    [switch]$SkipBackup,
    [switch]$CheckObservability,
    [switch]$CheckDependencies,
    [string]$TlsSecretName = "",
    [string]$DatabaseSecretName = "openclaw-platform-api-secret",
    [string]$ExternalSecretName = "openclaw-platform-api-secret",
    [string]$OpenSearchUrl = "",
    [string]$KeycloakBaseUrl = "",
    [string]$ExpectedVersion = "",
    [string]$ExpectedCommit = "",
    [string]$ExpectedStrictMode = "true",
    [string]$ExpectedStateBackend = "postgres",
    [string]$ExpectedRuntimeProvider = "kubectl"
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

if ($Image -ne "") {
    & (Join-Path $scriptDir "set-image.ps1") -Overlay $Overlay -Image $Image
}

& (Join-Path $scriptDir "preflight.ps1") -Overlay $Overlay -Strict
& (Join-Path $scriptDir "preflight.ps1") -Overlay $BootstrapOverlay -Bootstrap -Strict
& (Join-Path $scriptDir "validate.ps1") -Overlay $Overlay
if (-not $SkipBackup -and $BackupPostgresPod -eq "") {
    throw "BackupPostgresPod is required for release unless -SkipBackup is explicitly provided"
}
if ($BackupPostgresPod -ne "") {
    if ($BackupOutput -eq "") {
        $BackupOutput = ".\platform-api-pre-release-backup.sql"
    }
    & (Join-Path $scriptDir "db-backup.ps1") -Namespace $Namespace -PostgresPod $BackupPostgresPod -Output $BackupOutput
    & (Join-Path $scriptDir "verify-backup.ps1") -Path $BackupOutput
}
kubectl apply --dry-run=client --validate=false -k (Join-Path (Split-Path -Parent $scriptDir) ("bootstrap\\overlays\\" + $BootstrapOverlay)) | Out-Null
kubectl apply -k (Join-Path (Split-Path -Parent $scriptDir) ("bootstrap\\overlays\\" + $BootstrapOverlay))
& (Join-Path $scriptDir "verify-dependencies.ps1") `
    -Namespace $Namespace `
    -DatabaseSecretName $DatabaseSecretName `
    -ExternalSecretName $ExternalSecretName `
    -SkipDeploymentCheck `
    -WaitSeconds 180
kubectl wait --for=condition=complete job/openclaw-platform-api-bootstrap -n $Namespace --timeout=300s
& (Join-Path $scriptDir "apply.ps1") -Overlay $Overlay

kubectl rollout status deployment/$Deployment -n $Namespace --timeout=300s

& (Join-Path $scriptDir "smoke.ps1") -Namespace $Namespace -Service $Deployment
& (Join-Path $scriptDir "verify-release.ps1") `
    -Namespace $Namespace `
    -Service $Deployment `
    -ExpectedVersion $ExpectedVersion `
    -ExpectedCommit $ExpectedCommit `
    -ExpectedStrictMode $ExpectedStrictMode `
    -ExpectedStateBackend $ExpectedStateBackend `
    -ExpectedRuntimeProvider $ExpectedRuntimeProvider

if ($CheckDependencies -or $TlsSecretName -ne "" -or $OpenSearchUrl -ne "" -or $KeycloakBaseUrl -ne "" -or $DatabaseSecretName -ne "" -or $ExternalSecretName -ne "") {
    & (Join-Path $scriptDir "verify-dependencies.ps1") `
        -Namespace $Namespace `
        -Service $Deployment `
        -DatabaseSecretName $DatabaseSecretName `
        -ExternalSecretName $ExternalSecretName `
        -TlsSecretName $TlsSecretName `
        -OpenSearchUrl $OpenSearchUrl `
        -KeycloakBaseUrl $KeycloakBaseUrl
}

if ($CheckObservability) {
    & (Join-Path $scriptDir "verify-observability.ps1") -Namespace $Namespace
}

