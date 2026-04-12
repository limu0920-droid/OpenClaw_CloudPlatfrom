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

kubectl apply -k $target
