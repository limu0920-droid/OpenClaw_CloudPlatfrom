param(
    [Parameter(Mandatory = $true)]
    [string]$Summary,
    [double]$MaxFailedRate = 0.01,
    [double]$MaxP95Ms = 1000
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $Summary)) {
    throw "Summary file not found: $Summary"
}

$json = Get-Content -Raw -Encoding utf8 $Summary | ConvertFrom-Json

$failedRate = [double]$json.metrics.http_req_failed.values.rate
$p95 = [double]$json.metrics.http_req_duration.values.'p(95)'
$avg = [double]$json.metrics.http_req_duration.values.avg
$count = [double]$json.metrics.http_reqs.values.count

$errors = @()
if ($failedRate -gt $MaxFailedRate) {
    $errors += "http_req_failed rate $failedRate exceeds threshold $MaxFailedRate"
}
if ($p95 -gt $MaxP95Ms) {
    $errors += "http_req_duration p95 $p95 exceeds threshold $MaxP95Ms ms"
}

[pscustomobject]@{
    Requests     = $count
    FailedRate   = $failedRate
    AvgMs        = $avg
    P95Ms        = $p95
    MaxFailedRate = $MaxFailedRate
    MaxP95Ms      = $MaxP95Ms
} | ConvertTo-Json -Depth 6

if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Error $_ }
    exit 1
}
