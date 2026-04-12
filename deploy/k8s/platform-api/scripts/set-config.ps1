param(
    [string]$Overlay = "prod",
    [Parameter(Mandatory = $true)]
    [string]$ApiHost,
    [Parameter(Mandatory = $true)]
    [string]$TlsSecretName,
    [Parameter(Mandatory = $true)]
    [string]$KeycloakBaseUrl,
    [Parameter(Mandatory = $true)]
    [string]$PortalLoginUrl,
    [Parameter(Mandatory = $true)]
    [string]$PortalPostLoginUrl,
    [Parameter(Mandatory = $true)]
    [string]$PortalLogoutUrl,
    [Parameter(Mandatory = $true)]
    [string]$OpenSearchUrl,
    [Parameter(Mandatory = $true)]
    [string]$ObjectStorageEndpoint,
    [string]$WorkspaceBridgePublicBaseUrl = "",
    [string]$ArtifactPreviewPublicBaseUrl = "",
    [string]$ArtifactPreviewAllowedHosts = "",
    [string]$SecretStoreKind = "ClusterSecretStore",
    [Parameter(Mandatory = $true)]
    [string]$SecretStoreName,
    [string]$SecretRefreshInterval = "1h",
    [Parameter(Mandatory = $true)]
    [string]$SecretKeyPrefix
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $scriptDir

if ([string]::IsNullOrWhiteSpace($WorkspaceBridgePublicBaseUrl)) {
    $WorkspaceBridgePublicBaseUrl = "https://$ApiHost"
}
if ([string]::IsNullOrWhiteSpace($ArtifactPreviewPublicBaseUrl)) {
    $ArtifactPreviewPublicBaseUrl = "https://$ApiHost"
}
if ([string]::IsNullOrWhiteSpace($ArtifactPreviewAllowedHosts)) {
    $allowedHosts = New-Object System.Collections.Generic.List[string]
    $allowedHosts.Add($ApiHost) | Out-Null
    try {
        $storageHost = ([uri]$ObjectStorageEndpoint).Host
        if (-not [string]::IsNullOrWhiteSpace($storageHost)) {
            $allowedHosts.Add($storageHost) | Out-Null
        }
    }
    catch {
    }
    $ArtifactPreviewAllowedHosts = (($allowedHosts | Select-Object -Unique) -join ",")
}

$overlayDirs = @(
    (Join-Path $root ("overlays\" + $Overlay)),
    (Join-Path $root ("bootstrap\overlays\" + $Overlay))
)

foreach ($overlayDir in $overlayDirs) {
    if (-not (Test-Path $overlayDir)) {
        throw "overlay not found: $overlayDir"
    }

    $configPath = Join-Path $overlayDir "patch-configmap.yaml"
    $ingressPath = Join-Path $overlayDir "patch-ingress.yaml"

    if (Test-Path $configPath) {
        $config = Get-Content -Encoding utf8 $configPath -Raw
        $config = [regex]::Replace($config, 'KEYCLOAK_BASE_URL:\s*".*"', "KEYCLOAK_BASE_URL: `"$KeycloakBaseUrl`"")
        $config = [regex]::Replace($config, 'KEYCLOAK_REDIRECT_URL:\s*".*"', "KEYCLOAK_REDIRECT_URL: `"$PortalLoginUrl`"")
        $config = [regex]::Replace($config, 'KEYCLOAK_POST_LOGIN_REDIRECT_URL:\s*".*"', "KEYCLOAK_POST_LOGIN_REDIRECT_URL: `"$PortalPostLoginUrl`"")
        $config = [regex]::Replace($config, 'KEYCLOAK_LOGOUT_REDIRECT_URL:\s*".*"', "KEYCLOAK_LOGOUT_REDIRECT_URL: `"$PortalLogoutUrl`"")
        $config = [regex]::Replace($config, 'OPENSEARCH_URL:\s*".*"', "OPENSEARCH_URL: `"$OpenSearchUrl`"")
        $config = [regex]::Replace($config, 'OBJECT_STORAGE_ENDPOINT:\s*".*"', "OBJECT_STORAGE_ENDPOINT: `"$ObjectStorageEndpoint`"")
        $config = [regex]::Replace($config, 'WORKSPACE_BRIDGE_PUBLIC_BASE_URL:\s*".*"', "WORKSPACE_BRIDGE_PUBLIC_BASE_URL: `"$WorkspaceBridgePublicBaseUrl`"")
        $config = [regex]::Replace($config, 'ARTIFACT_PREVIEW_PUBLIC_BASE_URL:\s*".*"', "ARTIFACT_PREVIEW_PUBLIC_BASE_URL: `"$ArtifactPreviewPublicBaseUrl`"")
        $config = [regex]::Replace($config, 'ARTIFACT_PREVIEW_ALLOWED_HOSTS:\s*".*"', "ARTIFACT_PREVIEW_ALLOWED_HOSTS: `"$ArtifactPreviewAllowedHosts`"")
        Set-Content -Encoding utf8 -Path $configPath -Value $config
    }

    if ($overlayDir -like "*\overlays\$Overlay" -and (Test-Path $ingressPath)) {
        $ingress = Get-Content -Encoding utf8 $ingressPath -Raw
        $ingress = [regex]::Replace($ingress, '(?m)^(\s{8}-\s+)\S+\s*$', "`$1$ApiHost")
        $ingress = [regex]::Replace($ingress, '(?m)^(\s{4}-\s+host:\s+)\S+\s*$', "`$1$ApiHost")
        $ingress = [regex]::Replace($ingress, '(?m)^(\s{6}secretName:\s+)\S+\s*$', "`$1$TlsSecretName")
        Set-Content -Encoding utf8 -Path $ingressPath -Value $ingress
    }
}

$externalSecretPath = Join-Path $root "security\prod\externalsecret.yaml"
if (-not (Test-Path $externalSecretPath)) {
    throw "external secret template not found: $externalSecretPath"
}

$secretKeyPrefix = $SecretKeyPrefix.TrimEnd('/')
$externalSecret = Get-Content -Encoding utf8 $externalSecretPath -Raw
$externalSecret = [regex]::Replace($externalSecret, '(?m)^  refreshInterval:\s*.+$', "  refreshInterval: $SecretRefreshInterval")
$externalSecret = [regex]::Replace($externalSecret, '(?ms)(secretStoreRef:\s*\r?\n\s*kind:\s*).+?(\r?\n\s*name:\s*).+?(\r?\n)', "`$1$SecretStoreKind`$2$SecretStoreName`$3")

$secretKeys = @(
    "DATABASE_URL",
    "OBJECT_STORAGE_ACCESS_KEY",
    "OBJECT_STORAGE_SECRET_KEY",
    "KEYCLOAK_CLIENT_SECRET",
    "KEYCLOAK_SESSION_SECRET",
    "OPENSEARCH_USERNAME",
    "OPENSEARCH_PASSWORD",
    "WECHATPAY_MCH_ID",
    "WECHATPAY_APP_ID",
    "WECHATPAY_CLIENT_SECRET",
    "WECHATPAY_NOTIFY_URL",
    "WECHATPAY_REFUND_NOTIFY_URL",
    "WECHATPAY_SERIAL_NO",
    "WECHATPAY_PUBLIC_KEY_ID",
    "WECHATPAY_PUBLIC_KEY_PEM",
    "WECHATPAY_PRIVATE_KEY_PEM",
    "WECHATPAY_APIV3_KEY",
    "WECHATPAY_MODE",
    "WECHATPAY_SUB_MCH_ID",
    "WECHATPAY_SUB_APP_ID",
    "WECHAT_LOGIN_APP_ID",
    "WECHAT_LOGIN_APP_SECRET",
    "WECHAT_LOGIN_REDIRECT_URL",
    "WORKSPACE_BRIDGE_TOKEN"
)

foreach ($secretKey in $secretKeys) {
    $pattern = '(?m)^(\s*key:\s*").+/' + [regex]::Escape($secretKey) + '"\s*$'
    $replacement = "        key: `"$secretKeyPrefix/$secretKey`""
    $externalSecret = [regex]::Replace($externalSecret, $pattern, $replacement)
}

Set-Content -Encoding utf8 -Path $externalSecretPath -Value $externalSecret
