param(
    [Parameter(Mandatory = $true)]
    [string]$InputPath,
    [Parameter(Mandatory = $true)]
    [string]$OutputPath
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $InputPath)) {
    throw "Input file not found: $InputPath"
}

$lines = Get-Content -Encoding utf8 $InputPath
$inSecret = $false
$inStringData = $false
$sanitized = New-Object System.Collections.Generic.List[string]

foreach ($line in $lines) {
    if ($line -match '^kind:\s+Secret\s*$') {
        $inSecret = $true
        $inStringData = $false
        $sanitized.Add($line)
        continue
    }
    if ($inSecret -and $line -match '^---\s*$') {
        $inSecret = $false
        $inStringData = $false
        $sanitized.Add($line)
        continue
    }
    if ($inSecret -and $line -match '^\s*stringData:\s*$') {
        $inStringData = $true
        $sanitized.Add($line)
        continue
    }
    if ($inStringData -and $line -match '^\s+[A-Za-z0-9_]+\s*:\s*') {
        $prefix = ($line -replace '^(\s*[A-Za-z0-9_]+\s*:\s*).+$', '$1')
        $sanitized.Add($prefix + '"REDACTED"')
        continue
    }
    if ($inStringData -and $line -notmatch '^\s+') {
        $inStringData = $false
    }
    $sanitized.Add($line)
}

Set-Content -Encoding utf8 -Path $OutputPath -Value $sanitized
