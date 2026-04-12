param(
    [switch]$Build
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Push-Location $scriptDir
try {
    if ($Build) {
        Set-ExecutionPolicy -Scope Process Bypass
        .\up.ps1 -WithAPI -Build -Wait
    } else {
        .\up.ps1 -WithAPI -Wait
    }

    $order = Invoke-RestMethod -Method Post -Uri http://127.0.0.1:18080/api/v1/portal/orders -ContentType 'application/json' -Body (@{
        planCode = 'trial'
        orderType = 'buy'
        instanceId = 100
    } | ConvertTo-Json)

    $invoice = Invoke-RestMethod -Method Post -Uri http://127.0.0.1:18080/api/v1/portal/invoices -ContentType 'application/json' -Body (@{
        orderId = $order.order.id
        invoiceType = 'vat_normal'
        title = 'Persistence Smoke'
        taxNo = '91310000SMOKE'
        email = 'finance@smoke.example.com'
    } | ConvertTo-Json)

    $ticket = Invoke-RestMethod -Method Post -Uri http://127.0.0.1:18080/api/v1/portal/tickets -ContentType 'application/json' -Body (@{
        instanceId = 100
        title = 'Persistence Smoke Ticket'
        reporter = 'Smoke'
        category = 'ops'
        severity = 'medium'
        description = 'Smoke persistence test'
    } | ConvertTo-Json)

    $theme = Invoke-RestMethod -Method Patch -Uri http://127.0.0.1:18080/api/v1/admin/oem/brands/1/theme -ContentType 'application/json' -Body (@{
        primaryColor = '#121212'
        secondaryColor = '#343434'
        accentColor = '#ff7700'
        surfaceMode = 'light'
        fontFamily = 'MiSans VF'
        radius = '14px'
    } | ConvertTo-Json)

    docker compose --env-file .env restart platform-api | Out-Null
    Start-Sleep -Seconds 8

    $orders = Invoke-RestMethod -Uri http://127.0.0.1:18080/api/v1/admin/orders
    $invoices = Invoke-RestMethod -Uri http://127.0.0.1:18080/api/v1/admin/invoices
    $tickets = Invoke-RestMethod -Uri http://127.0.0.1:18080/api/v1/admin/tickets
    $brand = Invoke-RestMethod -Uri http://127.0.0.1:18080/api/v1/admin/oem/brands/1

    [pscustomobject]@{
        OrderPersisted = [bool](($orders.items | Where-Object { $_.id -eq $order.order.id }))
        InvoicePersisted = [bool](($invoices.items | Where-Object { $_.id -eq $invoice.invoice.id }))
        TicketPersisted = [bool](($tickets.items | Where-Object { $_.id -eq $ticket.ticket.id }))
        ThemePrimary = $brand.theme.primaryColor
    } | ConvertTo-Json -Depth 8
}
finally {
    try {
        .\down.ps1 -RemoveVolumes | Out-Null
    }
    catch {
    }
    Pop-Location
}
