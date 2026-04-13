param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$SummaryOutput = ".\\perf\\platform-api\\platform-api-k6-stability-summary.json"
)

$ErrorActionPreference = "Stop"

$env:BASE_URL = $BaseUrl
k6 run --summary-export $SummaryOutput .\perf\platform-api\k6-stability.js
Write-Output "k6 stability test summary exported to $SummaryOutput"
