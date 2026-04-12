param(
    [string]$Overlay = "prod",
    [Parameter(Mandatory = $true)]
    [string]$Image
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $scriptDir
$targets = @(
    (Join-Path $root ("overlays\" + $Overlay + "\patch-deployment.yaml")),
    (Join-Path $root ("overlays\" + $Overlay + "\patch-bootstrap-job.yaml")),
    (Join-Path $root ("bootstrap\overlays\" + $Overlay + "\patch-bootstrap-job.yaml"))
)

$updated = $false
foreach ($target in $targets) {
    if (-not (Test-Path $target)) {
        continue
    }
    $content = Get-Content -Encoding utf8 $target -Raw
    $content = [regex]::Replace($content, 'image:\s*.+', "image: $Image")
    Set-Content -Encoding utf8 -Path $target -Value $content
    $updated = $true
}

if (-not $updated) {
    throw "no image patch files found for overlay: $Overlay"
}
