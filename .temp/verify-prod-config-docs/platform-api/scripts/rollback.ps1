param(
    [string]$Namespace = "openclaw-system",
    [string]$Deployment = "openclaw-platform-api"
)

$ErrorActionPreference = "Stop"

kubectl rollout undo deployment/$Deployment -n $Namespace
kubectl rollout status deployment/$Deployment -n $Namespace --timeout=300s
kubectl get rs -n $Namespace -l "app.kubernetes.io/name=$Deployment"

