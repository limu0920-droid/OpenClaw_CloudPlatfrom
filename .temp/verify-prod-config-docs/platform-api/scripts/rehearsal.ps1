param(
    [string]$Overlay = "prod",
    [string]$BootstrapOverlay = "prod"
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $scriptDir

& (Join-Path $scriptDir "preflight.ps1") -Overlay $Overlay
& (Join-Path $scriptDir "validate.ps1") -Overlay $Overlay
kubectl apply --dry-run=client -k (Join-Path $root ("bootstrap\\overlays\\" + $BootstrapOverlay))
kubectl kustomize (Join-Path $root ("bootstrap\\overlays\\" + $BootstrapOverlay)) | Out-File -Encoding utf8 "platform-api-bootstrap-$BootstrapOverlay.rehearsal.yaml"
kubectl kustomize (Join-Path $root ("overlays\\" + $Overlay)) | Out-File -Encoding utf8 "platform-api-$Overlay.rehearsal.yaml"

Write-Output "rehearsal completed"

