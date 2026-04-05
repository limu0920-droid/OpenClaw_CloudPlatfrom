package httpapi

import (
	"net/http"

	"openclaw/mockapi/internal/models"
)

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

func (r *Router) handleOEMConfig(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	binding := r.findBrandBindingByTenant(tenantID)
	if binding == nil {
		http.NotFound(w, req)
		return
	}
	brand := r.findBrand(binding.BrandID)
	if brand == nil {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"brand":    brand,
		"theme":    r.findBrandTheme(binding.BrandID),
		"features": r.findBrandFeatures(binding.BrandID),
		"binding":  binding,
	})
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
	defer r.mu.Unlock()

	index := r.findBrandThemeIndex(brandID)
	if index < 0 {
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

	writeJSON(w, http.StatusOK, map[string]any{"theme": r.data.BrandThemes[index]})
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
	defer r.mu.Unlock()

	index := r.findBrandFeaturesIndex(brandID)
	if index < 0 {
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

	writeJSON(w, http.StatusOK, map[string]any{"features": r.data.BrandFeatures[index]})
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

func (r *Router) findBrandFeaturesIndex(brandID int) int {
	for index, item := range r.data.BrandFeatures {
		if item.BrandID == brandID {
			return index
		}
	}
	return -1
}
