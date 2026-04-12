param(
    [string]$Overlay = "prod",
    [switch]$Bootstrap,
    [switch]$Strict
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $scriptDir
$targetBase = "overlays"
if ($Bootstrap) {
    $targetBase = "bootstrap\overlays"
}
$target = Join-Path $root ($targetBase + "\" + $Overlay)

if (-not (Test-Path $target)) {
    throw "overlay not found: $target"
}

$rendered = (kubectl kustomize $target | Out-String)
if ($LASTEXITCODE -ne 0) {
    throw "kubectl kustomize failed for $target"
}

$errors = @()
$warnings = @()
$prodPlaceholderTokens = @(
    "replace-me",
    "change-me",
    "set-in-cluster",
    "example.com",
    "api.openclaw.local",
    "change-me-secret-store",
    "/change-me/"
)

if ($Overlay -eq "prod" -and $Strict) {
    foreach ($token in $prodPlaceholderTokens) {
        if ($rendered -match [regex]::Escape($token)) {
            $errors += "rendered manifest still contains placeholder token: $token"
        }
    }
} elseif ($Overlay -eq "prod") {
    foreach ($token in $prodPlaceholderTokens) {
        if ($rendered -match [regex]::Escape($token)) {
            $warnings += "rendered manifest still contains placeholder token: $token"
        }
    }
}

$imageMatches = [regex]::Matches($rendered, 'image:\s*(\S+)')
foreach ($match in $imageMatches) {
    $image = $match.Groups[1].Value
    if ($Overlay -eq "prod" -and $image -notmatch '@sha256:') {
        $errors += "prod overlay image is not pinned by digest: $image"
    }
}

if ($Overlay -eq "prod" -and -not $Bootstrap) {
    if ($rendered -notmatch 'PLATFORM_STRICT_MODE:\s*"true"') {
        $errors += "prod overlay must enable PLATFORM_STRICT_MODE=true"
    }
    if ($rendered -match 'RUNTIME_PROVIDER:\s*"mock"') {
        $errors += "prod overlay must not use RUNTIME_PROVIDER=mock"
    }
    if ($rendered -match 'AUTO_MIGRATE:\s*"true"') {
        $errors += "prod overlay must not enable AUTO_MIGRATE"
    }
    if ($rendered -match 'AUTO_BOOTSTRAP:\s*"true"') {
        $errors += "prod overlay must not enable AUTO_BOOTSTRAP"
    }
    if ($rendered -notmatch 'WORKSPACE_BRIDGE_PUBLIC_BASE_URL:\s*"?\S+' ) {
        $errors += "prod overlay must define WORKSPACE_BRIDGE_PUBLIC_BASE_URL"
    }
    if ($rendered -notmatch 'ARTIFACT_PREVIEW_PUBLIC_BASE_URL:\s*"?\S+' ) {
        $errors += "prod overlay must define ARTIFACT_PREVIEW_PUBLIC_BASE_URL"
    }
    if ($rendered -notmatch 'tls:') {
        $errors += "prod overlay ingress must define tls"
    }
}

if ($Overlay -eq "prod") {
    if ($rendered -notmatch '(?ms)^kind:\s+ExternalSecret\s+metadata:\s+name:\s+openclaw-platform-api-secret\b') {
        $errors += "prod overlay must define ExternalSecret/openclaw-platform-api-secret"
    }
    if ($rendered -match '(?ms)^kind:\s+Secret\s+metadata:\s+name:\s+openclaw-platform-api-secret\b') {
        $errors += "prod overlay must not render inline Secret/openclaw-platform-api-secret"
    }
    if ($rendered -notmatch 'secretStoreRef:') {
        $errors += "prod overlay must define ExternalSecret secretStoreRef"
    }
}

if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Error $_ }
    exit 1
}

if ($warnings.Count -gt 0) {
    $warnings | ForEach-Object { Write-Warning $_ }
}

$kind = "app"
if ($Bootstrap) {
    $kind = "bootstrap"
}
Write-Output "preflight passed for $kind overlay: $Overlay"
