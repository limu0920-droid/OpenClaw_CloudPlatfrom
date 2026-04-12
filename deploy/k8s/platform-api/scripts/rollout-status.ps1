param(
    [string]$Namespace = "openclaw-system",
    [string]$Deployment = "openclaw-platform-api"
)

$ErrorActionPreference = "Stop"

kubectl rollout status deployment/$Deployment -n $Namespace
kubectl get pods -n $Namespace -l "app.kubernetes.io/name=$Deployment" -o wide

