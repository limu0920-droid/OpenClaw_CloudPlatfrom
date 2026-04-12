param(
    [string]$Overlay = "dev"
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $scriptDir
$target = Join-Path $root ("overlays\" + $Overlay)

if (-not (Test-Path $target)) {
    throw "overlay not found: $target"
}

$rendered = kubectl kustomize $target
if ($LASTEXITCODE -ne 0) {
    throw "kubectl kustomize failed for $target"
}
if ([string]::IsNullOrWhiteSpace(($rendered | Out-String).Trim())) {
    throw "rendered manifest is empty for overlay: $Overlay"
}

Write-Output "render validation passed for overlay: $Overlay"
