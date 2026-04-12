package httpapi

import (
	"net/http"
	"sort"

	"openclaw/platformapi/internal/models"
)

type updateAccountSettingsRequest struct {
	PrimaryEmail           string `json:"primaryEmail"`
	BillingEmail           string `json:"billingEmail"`
	AlertEmail             string `json:"alertEmail"`
	PreferredLocale        string `json:"preferredLocale"`
	SecondaryLocale        string `json:"secondaryLocale"`
	Timezone               string `json:"timezone"`
	MarketingOptIn         bool   `json:"marketingOptIn"`
	NotifyOnAlert          bool   `json:"notifyOnAlert"`
	NotifyOnPayment        bool   `json:"notifyOnPayment"`
	NotifyOnExpiry         bool   `json:"notifyOnExpiry"`
	NotifyChannelEmail     bool   `json:"notifyChannelEmail"`
	NotifyChannelWebhook   bool   `json:"notifyChannelWebhook"`
	NotifyChannelInApp     bool   `json:"notifyChannelInApp"`
	NotificationWebhookURL string `json:"notificationWebhookUrl"`
	PortalHeadline         string `json:"portalHeadline"`
	PortalSubtitle         string `json:"portalSubtitle"`
	WorkspaceCallout       string `json:"workspaceCallout"`
	ExperimentBadge        string `json:"experimentBadge"`
}

func (r *Router) handlePortalAccountSettings(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	settings := r.findAccountSettings(tenantID)
	if settings == nil {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"settings": settings})
}

func (r *Router) handlePortalUpdateAccountSettings(w http.ResponseWriter, req *http.Request) {
	var payload updateAccountSettingsRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	tenantID := tenantFilterID(req, 1)
	index := r.findAccountSettingsIndex(tenantID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	current := r.data.AccountSettings[index]
	if payload.PrimaryEmail != "" {
		current.PrimaryEmail = payload.PrimaryEmail
	}
	if payload.BillingEmail != "" {
		current.BillingEmail = payload.BillingEmail
	}
	if payload.AlertEmail != "" {
		current.AlertEmail = payload.AlertEmail
	}
	if payload.PreferredLocale != "" {
		current.PreferredLocale = payload.PreferredLocale
	}
	if payload.SecondaryLocale != "" {
		current.SecondaryLocale = payload.SecondaryLocale
	}
	if payload.Timezone != "" {
		current.Timezone = payload.Timezone
	}
	current.MarketingOptIn = payload.MarketingOptIn
	current.NotifyOnAlert = payload.NotifyOnAlert
	current.NotifyOnPayment = payload.NotifyOnPayment
	current.NotifyOnExpiry = payload.NotifyOnExpiry
	current.NotifyChannelEmail = payload.NotifyChannelEmail
	current.NotifyChannelWebhook = payload.NotifyChannelWebhook
	current.NotifyChannelInApp = payload.NotifyChannelInApp
	current.NotificationWebhookURL = payload.NotificationWebhookURL
	current.PortalHeadline = payload.PortalHeadline
	current.PortalSubtitle = payload.PortalSubtitle
	current.WorkspaceCallout = payload.WorkspaceCallout
	current.ExperimentBadge = payload.ExperimentBadge
	current.UpdatedAt = nowRFC3339()
	r.data.AccountSettings[index] = current
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist account settings failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"settings": current})
}

func (r *Router) handlePortalWallet(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	wallet := r.findWallet(tenantID)
	if wallet == nil {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": wallet})
}

func (r *Router) handlePortalBillingHistory(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	writeJSON(w, http.StatusOK, map[string]any{"items": r.filterBillingStatements(tenantID)})
}

func (r *Router) handleAdminWallets(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Wallets})
}

func (r *Router) handleAdminBillingHistory(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.BillingStatements})
}

func (r *Router) handleI18nConfig(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	settings := r.findAccountSettings(tenantID)
	locale := "zh-CN"
	timezone := "Asia/Shanghai"
	secondary := "en-US"
	if settings != nil {
		if settings.PreferredLocale != "" {
			locale = settings.PreferredLocale
		}
		if settings.Timezone != "" {
			timezone = settings.Timezone
		}
		if settings.SecondaryLocale != "" {
			secondary = settings.SecondaryLocale
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"defaultLocale":   locale,
		"secondaryLocale": secondary,
		"timezone":        timezone,
		"supportedLocales": []map[string]any{
			{"code": "zh-CN", "label": "简体中文"},
			{"code": "en-US", "label": "English"},
		},
	})
}

func (r *Router) findAccountSettings(tenantID int) *models.AccountSettings {
	for _, item := range r.data.AccountSettings {
		if item.TenantID == tenantID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) findAccountSettingsIndex(tenantID int) int {
	for index, item := range r.data.AccountSettings {
		if item.TenantID == tenantID {
			return index
		}
	}
	return -1
}

func (r *Router) findWallet(tenantID int) *models.WalletBalance {
	for _, item := range r.data.Wallets {
		if item.TenantID == tenantID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) filterBillingStatements(tenantID int) []models.BillingStatement {
	out := make([]models.BillingStatement, 0)
	for _, item := range r.data.BillingStatements {
		if item.TenantID == tenantID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt > out[j].CreatedAt
	})
	return out
}
