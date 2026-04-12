param(
    [Parameter(Mandatory = $true)]
    [string]$Summary,
    [string]$Environment = "",
    [string]$ImageDigest = "",
    [string]$MigrationStatus = "",
    [string]$Operator = "",
    [string]$Output = ".\\perf\\platform-api\\platform-api-baseline-record.md"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $Summary)) {
    throw "Summary file not found: $Summary"
}

$json = Get-Content -Raw -Encoding utf8 $Summary | ConvertFrom-Json
$count = [double]$json.metrics.http_reqs.values.count
$failedRate = [double]$json.metrics.http_req_failed.values.rate
$avg = [double]$json.metrics.http_req_duration.values.avg
$p95 = [double]$json.metrics.http_req_duration.values.'p(95)'

$content = @"
# Platform API Load Baseline Record

## Context

- Environment: $Environment
- Image digest: $ImageDigest
- Database version / migration status: $MigrationStatus
- Date: $(Get-Date -Format s)
- Operator: $Operator

## Result Summary

- Total requests: $count
- Failed request rate: $failedRate
- Average duration: $avg ms
- p95 duration: $p95 ms

## Decision

- Pass / Fail:
- Notes:
"@

Set-Content -Encoding utf8 -Path $Output -Value $content
Write-Output "baseline record written to $Output"

