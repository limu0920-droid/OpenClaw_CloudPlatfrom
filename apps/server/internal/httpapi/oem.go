package httpapi

import (
	"net/http"
	"strings"

	"openclaw/platformapi/internal/models"
)

type updateBrandRequest struct {
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	LogoURL      string   `json:"logoUrl"`
	FaviconURL   string   `json:"faviconUrl"`
	SupportEmail string   `json:"supportEmail"`
	SupportURL   string   `json:"supportUrl"`
	Domains      []string `json:"domains"`
}

type updateBrandThemeRequest struct {
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	AccentColor    string `json:"accentColor"`
	SurfaceMode    string `json:"surfaceMode"`
	FontFamily     string `json:"fontFamily"`
	Radius         string `json:"radius"`
}

type updateBrandFeaturesRequest struct {
	PortalEnabled         bool `json:"portalEnabled"`
	AdminEnabled          bool `json:"adminEnabled"`
	ChannelsEnabled       bool `json:"channelsEnabled"`
	TicketsEnabled        bool `json:"ticketsEnabled"`
	PurchaseEnabled       bool `json:"purchaseEnabled"`
	RuntimeControlEnabled bool `json:"runtimeControlEnabled"`
	AuditEnabled          bool `json:"auditEnabled"`
	SSOEnabled            bool `json:"ssoEnabled"`
}

type replaceBrandBindingsRequest struct {
	Bindings []struct {
		TenantID    int    `json:"tenantId"`
		BindingMode string `json:"bindingMode"`
	} `json:"bindings"`
}

func (r *Router) handleOEMConfig(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := r.resolveOEM(req)
	if len(result) == 0 {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleAdminOEMBrands(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]map[string]any, 0, len(r.data.Brands))
	for _, brand := range r.data.Brands {
		items = append(items, map[string]any{
			"brand":    brand,
			"theme":    r.findBrandTheme(brand.ID),
			"features": r.findBrandFeatures(brand.ID),
			"bindings": r.filterBrandBindings(brand.ID),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleAdminOEMBrandDetail(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	brandID, ok := parseTailID(req.URL.Path, "/api/v1/admin/oem/brands/")
	if !ok {
		http.Error(w, "invalid brand id", http.StatusBadRequest)
		return
	}
	brand := r.findBrand(brandID)
	if brand == nil {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"brand":    brand,
		"theme":    r.findBrandTheme(brandID),
		"features": r.findBrandFeatures(brandID),
		"bindings": r.filterBrandBindings(brandID),
	})
}

func (r *Router) handleAdminOEMBrandUpdate(w http.ResponseWriter, req *http.Request) {
	brandID, ok := parseTailID(req.URL.Path, "/api/v1/admin/oem/brands/")
	if !ok {
		http.Error(w, "invalid brand id", http.StatusBadRequest)
		return
	}

	var payload updateBrandRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	status := normalizeBrandStatus(payload.Status)
	if status == "" {
		http.Error(w, "status must be active, inactive or draft", http.StatusBadRequest)
		return
	}

	r.mu.Lock()
	index := r.findBrandIndex(brandID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	current := r.data.Brands[index]
	if strings.TrimSpace(payload.Name) != "" {
		current.Name = strings.TrimSpace(payload.Name)
	}
	current.Status = status
	current.LogoURL = strings.TrimSpace(payload.LogoURL)
	current.FaviconURL = strings.TrimSpace(payload.FaviconURL)
	current.SupportEmail = strings.TrimSpace(payload.SupportEmail)
	current.SupportURL = strings.TrimSpace(payload.SupportURL)
	current.Domains = normalizeBrandDomains(payload.Domains)
	current.UpdatedAt = nowRFC3339()
	r.data.Brands[index] = current
	r.mu.Unlock()

	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist oem brand failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"brand": current})
}

func (r *Router) handleAdminOEMBrandTheme(w http.ResponseWriter, req *http.Request) {
	brandID, ok := parseTailID(req.URL.Path, "/api/v1/admin/oem/brands/", "/theme")
	if !ok {
		http.Error(w, "invalid brand id", http.StatusBadRequest)
		return
	}

	var payload updateBrandThemeRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.findBrandThemeIndex(brandID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	r.data.BrandThemes[index] = models.OEMTheme{
		BrandID:        brandID,
		PrimaryColor:   payload.PrimaryColor,
		SecondaryColor: payload.SecondaryColor,
		AccentColor:    payload.AccentColor,
		SurfaceMode:    payload.SurfaceMode,
		FontFamily:     payload.FontFamily,
		Radius:         payload.Radius,
	}
	theme := r.data.BrandThemes[index]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist oem theme failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"theme": theme})
}

func (r *Router) handleAdminOEMBrandFeatures(w http.ResponseWriter, req *http.Request) {
	brandID, ok := parseTailID(req.URL.Path, "/api/v1/admin/oem/brands/", "/features")
	if !ok {
		http.Error(w, "invalid brand id", http.StatusBadRequest)
		return
	}

	var payload updateBrandFeaturesRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.findBrandFeaturesIndex(brandID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	r.data.BrandFeatures[index] = models.OEMFeatureFlags{
		BrandID:               brandID,
		PortalEnabled:         payload.PortalEnabled,
		AdminEnabled:          payload.AdminEnabled,
		ChannelsEnabled:       payload.ChannelsEnabled,
		TicketsEnabled:        payload.TicketsEnabled,
		PurchaseEnabled:       payload.PurchaseEnabled,
		RuntimeControlEnabled: payload.RuntimeControlEnabled,
		AuditEnabled:          payload.AuditEnabled,
		SSOEnabled:            payload.SSOEnabled,
	}
	features := r.data.BrandFeatures[index]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist oem features failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"features": features})
}

func (r *Router) handleAdminOEMBrandBindings(w http.ResponseWriter, req *http.Request) {
	brandID, ok := parseTailID(req.URL.Path, "/api/v1/admin/oem/brands/", "/bindings")
	if !ok {
		http.Error(w, "invalid brand id", http.StatusBadRequest)
		return
	}

	var payload replaceBrandBindingsRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()
	if r.findBrandIndex(brandID) < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	incomingTenantIDs := make(map[int]struct{}, len(payload.Bindings))
	for _, item := range payload.Bindings {
		if item.TenantID <= 0 || r.findTenant(item.TenantID) == nil {
			r.mu.Unlock()
			http.Error(w, "binding tenantId is invalid", http.StatusBadRequest)
			return
		}
		if _, exists := incomingTenantIDs[item.TenantID]; exists {
			r.mu.Unlock()
			http.Error(w, "binding tenantId must be unique", http.StatusBadRequest)
			return
		}
		incomingTenantIDs[item.TenantID] = struct{}{}
	}

	nextBindings := make([]models.TenantBrandBinding, 0, len(r.data.BrandBindings)+len(payload.Bindings))
	for _, item := range r.data.BrandBindings {
		if item.BrandID == brandID {
			continue
		}
		if _, exists := incomingTenantIDs[item.TenantID]; exists {
			continue
		}
		nextBindings = append(nextBindings, item)
	}
	now := nowRFC3339()
	for _, item := range payload.Bindings {
		nextBindings = append(nextBindings, models.TenantBrandBinding{
			TenantID:    item.TenantID,
			BrandID:     brandID,
			BindingMode: normalizeBindingMode(item.BindingMode),
			UpdatedAt:   now,
		})
	}
	r.data.BrandBindings = nextBindings
	bindings := r.filterBrandBindings(brandID)
	r.mu.Unlock()

	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist oem bindings failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"bindings": bindings})
}

func (r *Router) findBrand(id int) *models.OEMBrand {
	for _, item := range r.data.Brands {
		if item.ID == id {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) findBrandTheme(brandID int) *models.OEMTheme {
	for _, item := range r.data.BrandThemes {
		if item.BrandID == brandID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) findBrandFeatures(brandID int) *models.OEMFeatureFlags {
	for _, item := range r.data.BrandFeatures {
		if item.BrandID == brandID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) findBrandBindingByTenant(tenantID int) *models.TenantBrandBinding {
	for _, item := range r.data.BrandBindings {
		if item.TenantID == tenantID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) filterBrandBindings(brandID int) []models.TenantBrandBinding {
	out := make([]models.TenantBrandBinding, 0)
	for _, item := range r.data.BrandBindings {
		if item.BrandID == brandID {
			out = append(out, item)
		}
	}
	return out
}

func (r *Router) findBrandThemeIndex(brandID int) int {
	for index, item := range r.data.BrandThemes {
		if item.BrandID == brandID {
			return index
		}
	}
	return -1
}

func (r *Router) findBrandIndex(brandID int) int {
	for index, item := range r.data.Brands {
		if item.ID == brandID {
			return index
		}
	}
	return -1
}

func (r *Router) findBrandFeaturesIndex(brandID int) int {
	for index, item := range r.data.BrandFeatures {
		if item.BrandID == brandID {
			return index
		}
	}
	return -1
}

func normalizeBrandDomains(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizeBrandStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "active":
		return "active"
	case "inactive":
		return "inactive"
	case "draft":
		return "draft"
	default:
		return ""
	}
}

func normalizeBindingMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "shared":
		return "shared"
	default:
		return "dedicated"
	}
}
