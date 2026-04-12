param(
    [string]$Overlay = "prod",
    [string]$BootstrapOverlay = "prod",
    [string]$OutputDir = ".\\release-bundle"
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $scriptDir)))

if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Path $OutputDir | Out-Null

$bootstrapManifest = Join-Path $OutputDir "platform-api-bootstrap-$BootstrapOverlay.yaml"
$appManifest = Join-Path $OutputDir "platform-api-$Overlay.yaml"
$bootstrapSanitized = Join-Path $OutputDir "platform-api-bootstrap-$BootstrapOverlay.sanitized.yaml"
$appSanitized = Join-Path $OutputDir "platform-api-$Overlay.sanitized.yaml"

$bootstrapRendered = kubectl kustomize (Join-Path (Split-Path -Parent $scriptDir) ("bootstrap\\overlays\\" + $BootstrapOverlay))
if ($LASTEXITCODE -ne 0) {
    throw "kubectl kustomize failed for bootstrap overlay: $BootstrapOverlay"
}
$bootstrapRendered | Set-Content -Encoding utf8 $bootstrapManifest

$appRendered = kubectl kustomize (Join-Path (Split-Path -Parent $scriptDir) ("overlays\\" + $Overlay))
if ($LASTEXITCODE -ne 0) {
    throw "kubectl kustomize failed for app overlay: $Overlay"
}
$appRendered | Set-Content -Encoding utf8 $appManifest

& (Join-Path $scriptDir "sanitize-manifest.ps1") -InputPath $bootstrapManifest -OutputPath $bootstrapSanitized
& (Join-Path $scriptDir "sanitize-manifest.ps1") -InputPath $appManifest -OutputPath $appSanitized

$docsOut = Join-Path $OutputDir "docs"
New-Item -ItemType Directory -Path $docsOut | Out-Null

Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\README.md") -Destination (Join-Path $docsOut "k8s-readme.md")
Copy-Item -Path (Join-Path $repoRoot "docs\\runbooks\\2026-04-07-platform-api-operations-runbook.md") -Destination (Join-Path $docsOut "operations-runbook.md")
Copy-Item -Path (Join-Path $repoRoot "docs\\runbooks\\2026-04-07-platform-api-release-checklist.md") -Destination (Join-Path $docsOut "release-checklist.md")
Copy-Item -Path (Join-Path $repoRoot "docs\\runbooks\\2026-04-07-platform-api-production-environment-template.md") -Destination (Join-Path $docsOut "production-environment-template.md")
Copy-Item -Path (Join-Path $repoRoot "docs\\runbooks\\2026-04-09-platform-api-production-config-matrix.md") -Destination (Join-Path $docsOut "production-config-matrix.md")
Copy-Item -Path (Join-Path $repoRoot "docs\\runbooks\\2026-04-07-platform-api-db-rollback-sop.md") -Destination (Join-Path $docsOut "db-rollback-sop.md")
Copy-Item -Path (Join-Path $repoRoot "docs\\runbooks\\2026-04-07-platform-api-release-approval-template.md") -Destination (Join-Path $docsOut "release-approval-template.md")
Copy-Item -Path (Join-Path $repoRoot "apps\\platform-api\\internal\\httpapi\\externaldocs\\openapi.yaml") -Destination (Join-Path $docsOut "external-openapi.yaml")
Copy-Item -Path (Join-Path $repoRoot "apps\\platform-api\\internal\\httpapi\\externaldocs\\integration-guide.md") -Destination (Join-Path $docsOut "external-integration-guide.md")

$obsOut = Join-Path $OutputDir "observability"
New-Item -ItemType Directory -Path $obsOut | Out-Null
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\observability\\README.md") -Destination (Join-Path $obsOut "README.md")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\observability\\servicemonitor.yaml") -Destination (Join-Path $obsOut "servicemonitor.yaml")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\observability\\prometheusrule.yaml") -Destination (Join-Path $obsOut "prometheusrule.yaml")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\observability\\grafana-dashboard.json") -Destination (Join-Path $obsOut "grafana-dashboard.json")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\observability\\grafana-dashboard-configmap.yaml") -Destination (Join-Path $obsOut "grafana-dashboard-configmap.yaml")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\observability\\alertmanagerconfig.example.yaml") -Destination (Join-Path $obsOut "alertmanagerconfig.example.yaml")

$securityOut = Join-Path $OutputDir "security"
New-Item -ItemType Directory -Path $securityOut | Out-Null
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\security\\externalsecret.example.yaml") -Destination (Join-Path $securityOut "externalsecret.example.yaml")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\security\\prod\\externalsecret.yaml") -Destination (Join-Path $securityOut "externalsecret.prod.yaml")

$scriptsOut = Join-Path $OutputDir "scripts"
New-Item -ItemType Directory -Path $scriptsOut | Out-Null
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\scripts\\smoke.ps1") -Destination (Join-Path $scriptsOut "smoke.ps1")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\scripts\\verify-release.ps1") -Destination (Join-Path $scriptsOut "verify-release.ps1")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\scripts\\verify-dependencies.ps1") -Destination (Join-Path $scriptsOut "verify-dependencies.ps1")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\scripts\\verify-observability.ps1") -Destination (Join-Path $scriptsOut "verify-observability.ps1")
Copy-Item -Path (Join-Path $repoRoot "deploy\\k8s\\platform-api\\scripts\\verify-backup.ps1") -Destination (Join-Path $scriptsOut "verify-backup.ps1")

$perfOut = Join-Path $OutputDir "perf"
New-Item -ItemType Directory -Path $perfOut | Out-Null
Copy-Item -Path (Join-Path $repoRoot "perf\\platform-api\\README.md") -Destination (Join-Path $perfOut "README.md")
Copy-Item -Path (Join-Path $repoRoot "perf\\platform-api\\k6-smoke.js") -Destination (Join-Path $perfOut "k6-smoke.js")
Copy-Item -Path (Join-Path $repoRoot "perf\\platform-api\\baseline-template.md") -Destination (Join-Path $perfOut "baseline-template.md")

Write-Output "release bundle created at $OutputDir"

