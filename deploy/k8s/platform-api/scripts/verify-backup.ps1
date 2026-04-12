param(
    [Parameter(Mandatory = $true)]
    [string]$Path
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $Path)) {
    throw "Backup file not found: $Path"
}

$item = Get-Item $Path
if ($item.Length -le 0) {
    throw "Backup file is empty: $Path"
}

$preview = Get-Content -Encoding utf8 -Path $Path -TotalCount 20 | Out-String
if ($preview -notmatch "CREATE|INSERT|COPY|PostgreSQL database dump") {
    throw "Backup file does not look like a PostgreSQL dump: $Path"
}

Write-Output "backup verification passed: $Path"
