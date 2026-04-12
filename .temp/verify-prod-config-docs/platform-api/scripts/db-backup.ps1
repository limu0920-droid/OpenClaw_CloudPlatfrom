param(
    [string]$Namespace = "openclaw-system",
    [string]$PostgresPod,
    [string]$Database = "platform",
    [string]$User = "platform",
    [string]$Output = ".\\platform-api-backup.sql"
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($PostgresPod)) {
    throw "PostgresPod is required"
}

kubectl exec -n $Namespace $PostgresPod -- sh -lc "PGPASSWORD=\$POSTGRES_PASSWORD pg_dump -U $User $Database" | Out-File -Encoding utf8 $Output
Write-Output "backup saved to $Output"

