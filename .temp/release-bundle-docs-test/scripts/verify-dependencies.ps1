param(
    [string]$Namespace = "openclaw-system",
    [string]$Service = "openclaw-platform-api",
    [string]$OpenSearchUrl = "",
    [string]$KeycloakBaseUrl = "",
    [string]$TlsSecretName = "",
    [string]$DatabaseSecretName = "openclaw-platform-api-secret",
    [string]$ExternalSecretName = "",
    [switch]$SkipDeploymentCheck,
    [int]$WaitSeconds = 0,
    [int]$PollIntervalSeconds = 5
)

$ErrorActionPreference = "Stop"

function New-CheckResult {
    param(
        [string]$Name,
        [string]$Status,
        [string]$Detail
    )

    return [pscustomobject]@{
        Name = $Name
        Status = $Status
        Detail = $Detail
    }
}

function Test-HttpStatus {
    param(
        [string]$Name,
        [string]$Uri
    )

    if ([string]::IsNullOrWhiteSpace($Uri)) {
        return New-CheckResult -Name $Name -Status "skipped" -Detail ""
    }

    try {
        $response = Invoke-WebRequest -UseBasicParsing -Uri $Uri -Method Get -TimeoutSec 10
        return New-CheckResult -Name $Name -Status "ok" -Detail ([string]$response.StatusCode)
    }
    catch {
        return New-CheckResult -Name $Name -Status "error" -Detail $_.Exception.Message
    }
}

function Test-K8sObject {
    param(
        [string]$Kind,
        [string]$Name,
        [string]$Namespace,
        [string]$DisplayName = $Kind
    )

    if ([string]::IsNullOrWhiteSpace($Name)) {
        return New-CheckResult -Name $DisplayName -Status "skipped" -Detail ""
    }

    try {
        $result = kubectl get $Kind $Name -n $Namespace -o name 2>$null
        if ([string]::IsNullOrWhiteSpace(($result | Out-String).Trim())) {
            return New-CheckResult -Name $DisplayName -Status "error" -Detail "$Kind/$Name not found"
        }
        return New-CheckResult -Name $DisplayName -Status "ok" -Detail (($result | Out-String).Trim())
    }
    catch {
        return New-CheckResult -Name $DisplayName -Status "error" -Detail $_.Exception.Message
    }
}

function Test-ExternalSecretReady {
    param(
        [string]$Name,
        [string]$Namespace
    )

    if ([string]::IsNullOrWhiteSpace($Name)) {
        return New-CheckResult -Name "externalsecret" -Status "skipped" -Detail ""
    }

    try {
        $raw = kubectl get externalsecret $Name -n $Namespace -o json 2>$null | Out-String
        if ([string]::IsNullOrWhiteSpace($raw)) {
            return New-CheckResult -Name "externalsecret" -Status "error" -Detail "externalsecret/$Name not found"
        }

        $resource = $raw | ConvertFrom-Json
        $conditions = @($resource.status.conditions)
        $readyCondition = $conditions | Where-Object { $_.type -eq "Ready" } | Select-Object -First 1
        if ($null -eq $readyCondition -or $readyCondition.status -ne "True") {
            $detail = "Ready condition missing"
            if ($null -ne $readyCondition) {
                $detail = "{0}: {1}" -f $readyCondition.reason, $readyCondition.message
            }
            return New-CheckResult -Name "externalsecret" -Status "error" -Detail $detail
        }

        return New-CheckResult -Name "externalsecret" -Status "ok" -Detail ($readyCondition.reason)
    }
    catch {
        return New-CheckResult -Name "externalsecret" -Status "error" -Detail $_.Exception.Message
    }
}

function Test-SecretMaterialization {
    param(
        [string]$Name,
        [string]$Namespace
    )

    if ([string]::IsNullOrWhiteSpace($Name)) {
        return New-CheckResult -Name "runtime-secret" -Status "skipped" -Detail ""
    }

    $requiredKeys = @(
        "DATABASE_URL",
        "OBJECT_STORAGE_ACCESS_KEY",
        "OBJECT_STORAGE_SECRET_KEY",
        "KEYCLOAK_CLIENT_SECRET",
        "KEYCLOAK_SESSION_SECRET",
        "OPENSEARCH_USERNAME",
        "OPENSEARCH_PASSWORD"
    )
    $placeholderTokens = @("replace-me", "change-me", "set-in-cluster", "example.com")

    try {
        $raw = kubectl get secret $Name -n $Namespace -o json 2>$null | Out-String
        if ([string]::IsNullOrWhiteSpace($raw)) {
            return New-CheckResult -Name "runtime-secret" -Status "error" -Detail "secret/$Name not found"
        }

        $resource = $raw | ConvertFrom-Json
        $data = $resource.data
        foreach ($secretKey in $requiredKeys) {
            $encoded = $data.$secretKey
            if ([string]::IsNullOrWhiteSpace($encoded)) {
                return New-CheckResult -Name "runtime-secret" -Status "error" -Detail "secret/$Name missing key $secretKey"
            }

            $decoded = [System.Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($encoded))
            if ([string]::IsNullOrWhiteSpace($decoded)) {
                return New-CheckResult -Name "runtime-secret" -Status "error" -Detail "secret/$Name key $secretKey is empty"
            }

            foreach ($token in $placeholderTokens) {
                if ($decoded -match [regex]::Escape($token)) {
                    return New-CheckResult -Name "runtime-secret" -Status "error" -Detail "secret/$Name key $secretKey still contains placeholder token $token"
                }
            }
        }

        return New-CheckResult -Name "runtime-secret" -Status "ok" -Detail ("{0} keys verified" -f $requiredKeys.Count)
    }
    catch {
        return New-CheckResult -Name "runtime-secret" -Status "error" -Detail $_.Exception.Message
    }
}

function Invoke-Checks {
    $results = @()

    if (-not $SkipDeploymentCheck -and -not [string]::IsNullOrWhiteSpace($Service)) {
        $results += Test-K8sObject -Kind "deployment" -Name $Service -Namespace $Namespace -DisplayName "deployment"
    }

    $results += Test-ExternalSecretReady -Name $ExternalSecretName -Namespace $Namespace
    $results += Test-SecretMaterialization -Name $DatabaseSecretName -Namespace $Namespace
    $results += Test-K8sObject -Kind "secret" -Name $TlsSecretName -Namespace $Namespace -DisplayName "tls-secret"
    $results += Test-HttpStatus -Name "OpenSearch" -Uri $(if ($OpenSearchUrl) { ($OpenSearchUrl.TrimEnd('/')) + "/_cluster/health" } else { "" })
    $results += Test-HttpStatus -Name "Keycloak" -Uri $(if ($KeycloakBaseUrl) { ($KeycloakBaseUrl.TrimEnd('/')) + "/realms/master/.well-known/openid-configuration" } else { "" })

    return ,$results
}

$deadline = (Get-Date).AddSeconds([Math]::Max($WaitSeconds, 0))
$checks = @()
$failed = @()

do {
    $checks = @(Invoke-Checks)
    $failed = @($checks | Where-Object { $_.Status -eq "error" })
    if ($failed.Count -eq 0 -or $WaitSeconds -le 0 -or (Get-Date) -ge $deadline) {
        break
    }
    Start-Sleep -Seconds ([Math]::Max($PollIntervalSeconds, 1))
} while ($true)

[pscustomobject]@{
    Namespace = $Namespace
    Service = $Service
    Checks = $checks
} | ConvertTo-Json -Depth 8

if ($failed.Count -gt 0) {
    exit 1
}
