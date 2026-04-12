param(
    [string]$Namespace = "openclaw-system",
    [string]$PostgresPod,
    [string]$Database = "platform",
    [string]$User = "platform",
    [Parameter(Mandatory = $true)]
    [string]$Input
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($PostgresPod)) {
    throw "PostgresPod is required"
}
if (-not (Test-Path $Input)) {
    throw "Input file not found: $Input"
}

Get-Content -Raw -Encoding utf8 $Input | kubectl exec -i -n $Namespace $PostgresPod -- sh -lc "PGPASSWORD=\$POSTGRES_PASSWORD psql -U $User $Database"
Write-Output "restore completed from $Input"
