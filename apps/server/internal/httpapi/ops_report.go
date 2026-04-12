package httpapi

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
)

type portalOpsMonthlyUsage struct {
	Label        string `json:"label"`
	ChargeAmount int    `json:"chargeAmount"`
	PaidAmount   int    `json:"paidAmount"`
	Sessions     int    `json:"sessions"`
	Messages     int    `json:"messages"`
	Artifacts    int    `json:"artifacts"`
}

type portalNotificationChannel struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Enabled     bool   `json:"enabled"`
	Target      string `json:"target,omitempty"`
	Description string `json:"description"`
}

type portalNotificationTemplatePreview struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type portalOpsReportSummary struct {
	InstanceCount     int    `json:"instanceCount"`
	SessionCount      int    `json:"sessionCount"`
	MessageCount      int    `json:"messageCount"`
	ArtifactCount     int    `json:"artifactCount"`
	OrderCount        int    `json:"orderCount"`
	OpenTicketCount   int    `json:"openTicketCount"`
	StatementCount    int    `json:"statementCount"`
	WalletBalance     int    `json:"walletBalance"`
	Currency          string `json:"currency"`
	BillingMonthCount int    `json:"billingMonthCount"`
}

type portalOpsReportData struct {
	Tenant                *models.Tenant
	Brand                 *models.OEMBrand
	Settings              *models.AccountSettings
	Summary               portalOpsReportSummary
	NotificationChannels  []portalNotificationChannel
	NotificationTemplates []portalNotificationTemplatePreview
	MonthlyUsage          []portalOpsMonthlyUsage
}

func (r *Router) handlePortalOpsReport(w http.ResponseWriter, req *http.Request) {
	tenantID := tenantFilterID(req, 1)
	report, found := r.buildPortalOpsReport(tenantID)
	if !found {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": map[string]any{
			"id":     report.Tenant.ID,
			"name":   report.Tenant.Name,
			"plan":   report.Tenant.Plan,
			"status": report.Tenant.Status,
		},
		"brand": map[string]any{
			"name":         stringValue(report.Brand, func(item *models.OEMBrand) string { return item.Name }),
			"supportEmail": stringValue(report.Brand, func(item *models.OEMBrand) string { return item.SupportEmail }),
			"supportUrl":   stringValue(report.Brand, func(item *models.OEMBrand) string { return item.SupportURL }),
		},
		"settings":              report.Settings,
		"summary":               report.Summary,
		"notificationChannels":  report.NotificationChannels,
		"notificationTemplates": report.NotificationTemplates,
		"monthlyUsage":          report.MonthlyUsage,
		"export": map[string]any{
			"csvPath": "/api/v1/portal/ops/report/export.csv",
		},
	})
}

func (r *Router) handlePortalOpsReportExport(w http.ResponseWriter, req *http.Request) {
	tenantID := tenantFilterID(req, 1)
	report, found := r.buildPortalOpsReport(tenantID)
	if !found {
		http.NotFound(w, req)
		return
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{
		"tenantId",
		"tenantName",
		"tenantPlan",
		"tenantStatus",
		"brandName",
		"billingMonth",
		"chargeAmount",
		"paidAmount",
		"sessions",
		"messages",
		"artifacts",
		"instanceCount",
		"sessionCount",
		"messageCount",
		"artifactCount",
		"orderCount",
		"openTicketCount",
		"statementCount",
		"walletBalance",
		"currency",
	})

	rows := report.MonthlyUsage
	if len(rows) == 0 {
		rows = []portalOpsMonthlyUsage{{}}
	}
	for _, item := range rows {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", report.Tenant.ID),
			report.Tenant.Name,
			report.Tenant.Plan,
			report.Tenant.Status,
			stringValue(report.Brand, func(brand *models.OEMBrand) string { return brand.Name }),
			item.Label,
			fmt.Sprintf("%d", item.ChargeAmount),
			fmt.Sprintf("%d", item.PaidAmount),
			fmt.Sprintf("%d", item.Sessions),
			fmt.Sprintf("%d", item.Messages),
			fmt.Sprintf("%d", item.Artifacts),
			fmt.Sprintf("%d", report.Summary.InstanceCount),
			fmt.Sprintf("%d", report.Summary.SessionCount),
			fmt.Sprintf("%d", report.Summary.MessageCount),
			fmt.Sprintf("%d", report.Summary.ArtifactCount),
			fmt.Sprintf("%d", report.Summary.OrderCount),
			fmt.Sprintf("%d", report.Summary.OpenTicketCount),
			fmt.Sprintf("%d", report.Summary.StatementCount),
			fmt.Sprintf("%d", report.Summary.WalletBalance),
			report.Summary.Currency,
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("openclaw-portal-ops-%s.csv", time.Now().UTC().Format("20060102-150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	_, _ = w.Write(buffer.Bytes())
}

func (r *Router) buildPortalOpsReport(tenantID int) (portalOpsReportData, bool) {
	r.mu.RLock()
	tenant := r.findTenant(tenantID)
	settings := r.findAccountSettings(tenantID)
	wallet := r.findWallet(tenantID)
	brandBinding := r.findBrandBindingByTenant(tenantID)
	var brand *models.OEMBrand
	if brandBinding != nil {
		brand = r.findBrand(brandBinding.BrandID)
	}
	rawInstances := append([]models.Instance(nil), r.filterInstancesByTenant(tenantID)...)
	sessions := append([]models.WorkspaceSession(nil), r.filterWorkspaceSessionsByTenant(tenantID)...)
	messages := append([]models.WorkspaceMessage(nil), r.filterWorkspaceMessagesByTenant(tenantID)...)
	artifacts := append([]models.WorkspaceArtifact(nil), r.filterWorkspaceArtifactsByTenant(tenantID)...)
	orders := append([]models.Order(nil), r.filterOrdersByTenant(tenantID)...)
	statements := append([]models.BillingStatement(nil), r.filterBillingStatements(tenantID)...)
	tickets := append([]models.Ticket(nil), r.filterTicketsByTenant(tenantID)...)
	r.mu.RUnlock()

	if tenant == nil {
		return portalOpsReportData{}, false
	}

	instances := r.resolveLiveInstances(rawInstances)
	monthlyUsage := buildPortalOpsMonthlyUsage(statements, sessions, messages, artifacts)
	currency := resolvePortalOpsCurrency(wallet, statements)

	report := portalOpsReportData{
		Tenant:   tenant,
		Brand:    brand,
		Settings: settings,
		Summary: portalOpsReportSummary{
			InstanceCount:     len(instances),
			SessionCount:      len(sessions),
			MessageCount:      len(messages),
			ArtifactCount:     len(artifacts),
			OrderCount:        len(orders),
			OpenTicketCount:   countOpenTickets(tickets),
			StatementCount:    len(statements),
			WalletBalance:     walletBalanceAmount(wallet),
			Currency:          currency,
			BillingMonthCount: len(monthlyUsage),
		},
		NotificationChannels: buildPortalNotificationChannels(settings),
		NotificationTemplates: buildPortalNotificationTemplates(
			tenant,
			brand,
			settings,
			currency,
			walletBalanceAmount(wallet),
		),
		MonthlyUsage: monthlyUsage,
	}
	return report, true
}

func buildPortalNotificationChannels(settings *models.AccountSettings) []portalNotificationChannel {
	if settings == nil {
		return []portalNotificationChannel{
			{
				Key:         "email",
				Label:       "邮件通知",
				Enabled:     false,
				Description: "账单、告警与系统变更通过邮件发送。",
			},
			{
				Key:         "webhook",
				Label:       "Webhook 通知",
				Enabled:     false,
				Description: "将支付、告警和生命周期事件推送到外部系统。",
			},
			{
				Key:         "in_app",
				Label:       "站内提醒",
				Enabled:     false,
				Description: "在 Portal 内显示待处理提醒与运营通知。",
			},
		}
	}

	return []portalNotificationChannel{
		{
			Key:         "email",
			Label:       "邮件通知",
			Enabled:     settings.NotifyChannelEmail,
			Target:      firstNonEmpty(settings.AlertEmail, settings.PrimaryEmail, settings.BillingEmail),
			Description: "账单、告警与系统变更通过邮件发送。",
		},
		{
			Key:         "webhook",
			Label:       "Webhook 通知",
			Enabled:     settings.NotifyChannelWebhook,
			Target:      strings.TrimSpace(settings.NotificationWebhookURL),
			Description: "将支付、告警和生命周期事件推送到外部系统。",
		},
		{
			Key:         "in_app",
			Label:       "站内提醒",
			Enabled:     settings.NotifyChannelInApp,
			Description: "在 Portal 内显示待处理提醒与运营通知。",
		},
	}
}

func buildPortalNotificationTemplates(
	tenant *models.Tenant,
	brand *models.OEMBrand,
	settings *models.AccountSettings,
	currency string,
	walletBalance int,
) []portalNotificationTemplatePreview {
	brandName := firstNonEmpty(stringValue(brand, func(item *models.OEMBrand) string { return item.Name }), tenant.Name, "OpenClaw")
	alertEmail := strings.TrimSpace(stringValue(settings, func(item *models.AccountSettings) string { return item.AlertEmail }))
	paymentEmail := strings.TrimSpace(stringValue(settings, func(item *models.AccountSettings) string { return item.BillingEmail }))
	if paymentEmail == "" {
		paymentEmail = strings.TrimSpace(stringValue(settings, func(item *models.AccountSettings) string { return item.PrimaryEmail }))
	}

	return []portalNotificationTemplatePreview{
		{
			Key:     "payment",
			Title:   "支付到账提醒",
			Subject: fmt.Sprintf("[%s] 账单支付结果通知", brandName),
			Body:    fmt.Sprintf("账单与支付结果会发送到 %s，当前钱包余额 %d %s。", firstNonEmpty(paymentEmail, "未配置账单邮箱"), walletBalance, currency),
		},
		{
			Key:     "expiry",
			Title:   "实例到期提醒",
			Subject: fmt.Sprintf("[%s] 订阅到期与续费提醒", brandName),
			Body:    "系统会在实例到期前推送续费提醒，并附带待处理订单与账单入口。",
		},
		{
			Key:     "alert",
			Title:   "运行告警提醒",
			Subject: fmt.Sprintf("[%s] 运行告警与运维事件通知", brandName),
			Body:    fmt.Sprintf("告警升级后会优先触达到 %s，并同步站内运营中心。", firstNonEmpty(alertEmail, "未配置告警邮箱")),
		},
	}
}

func buildPortalOpsMonthlyUsage(
	statements []models.BillingStatement,
	sessions []models.WorkspaceSession,
	messages []models.WorkspaceMessage,
	artifacts []models.WorkspaceArtifact,
) []portalOpsMonthlyUsage {
	usageByMonth := make(map[string]*portalOpsMonthlyUsage)

	ensure := func(label string) *portalOpsMonthlyUsage {
		if strings.TrimSpace(label) == "" {
			label = "unknown"
		}
		current, ok := usageByMonth[label]
		if ok {
			return current
		}
		current = &portalOpsMonthlyUsage{Label: label}
		usageByMonth[label] = current
		return current
	}

	for _, item := range statements {
		label := normalizeUsageMonth(item.BillingMonth, item.CreatedAt)
		entry := ensure(label)
		entry.ChargeAmount += item.ChargeAmount
		entry.PaidAmount += item.PaidAmount
	}
	for _, item := range sessions {
		ensure(normalizeUsageMonth("", item.CreatedAt)).Sessions++
	}
	for _, item := range messages {
		ensure(normalizeUsageMonth("", item.CreatedAt)).Messages++
	}
	for _, item := range artifacts {
		ensure(normalizeUsageMonth("", item.CreatedAt)).Artifacts++
	}

	items := make([]portalOpsMonthlyUsage, 0, len(usageByMonth))
	for _, item := range usageByMonth {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Label > items[j].Label
	})
	return items
}

func normalizeUsageMonth(value string, fallback string) string {
	label := strings.TrimSpace(value)
	if label != "" {
		return label
	}
	if parsed, ok := parseRFC3339(fallback); ok {
		return parsed.Format("2006-01")
	}
	return "unknown"
}

func resolvePortalOpsCurrency(wallet *models.WalletBalance, statements []models.BillingStatement) string {
	if wallet != nil && strings.TrimSpace(wallet.Currency) != "" {
		return wallet.Currency
	}
	for _, item := range statements {
		if strings.TrimSpace(item.Currency) != "" {
			return item.Currency
		}
	}
	return "CNY"
}

func walletBalanceAmount(wallet *models.WalletBalance) int {
	if wallet == nil {
		return 0
	}
	return wallet.AvailableAmount
}

func countOpenTickets(items []models.Ticket) int {
	count := 0
	for _, item := range items {
		switch item.Status {
		case "resolved", "closed":
			continue
		default:
			count++
		}
	}
	return count
}

func (r *Router) filterOrdersByTenant(tenantID int) []models.Order {
	items := make([]models.Order, 0)
	for _, item := range r.data.Orders {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items
}

func stringValue[T any](item *T, selector func(*T) string) string {
	if item == nil {
		return ""
	}
	return strings.TrimSpace(selector(item))
}
