package httpapi

import (
	"net/http"
	"sort"

	"openclaw/platformapi/internal/models"
)

func (r *Router) handleAdminAccountPortrait(w http.ResponseWriter, req *http.Request) {
	userID, ok := parseTailID(req.URL.Path, "/api/v1/admin/accounts/", "/portrait")
	if !ok {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	profile := r.findUserProfile(userID)
	if profile == nil {
		http.NotFound(w, req)
		return
	}

	tenantID := profile.TenantID
	identities := r.filterAuthIdentitiesByUser(userID)
	recentOrders := r.filterRecentOrdersByTenant(tenantID, 8)
	recentTickets := r.filterRecentTicketsByTenant(tenantID, 8)

	writeJSON(w, http.StatusOK, map[string]any{
		"profile":           profile,
		"accountSettings":   r.findAccountSettings(tenantID),
		"wallet":            r.findWallet(tenantID),
		"billingStatements": r.filterBillingStatements(tenantID),
		"identities":        identities,
		"recentOrders":      recentOrders,
		"recentTickets":     recentTickets,
		"summary": map[string]any{
			"identityCount":       len(identities),
			"activeIdentityCount": countActiveIdentities(identities),
			"recentOrderCount":    len(recentOrders),
			"recentTicketCount":   len(recentTickets),
		},
	})
}

func (r *Router) filterRecentOrdersByTenant(tenantID int, limit int) []models.Order {
	items := make([]models.Order, 0)
	for _, order := range r.data.Orders {
		if order.TenantID == tenantID {
			items = append(items, order)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func (r *Router) filterRecentTicketsByTenant(tenantID int, limit int) []models.Ticket {
	items := make([]models.Ticket, 0)
	for _, ticket := range r.data.Tickets {
		if ticket.TenantID == tenantID {
			items = append(items, ticket)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func countActiveIdentities(items []models.AuthIdentity) int {
	count := 0
	for _, item := range items {
		if item.Status == "" || item.Status == "active" {
			count++
		}
	}
	return count
}
