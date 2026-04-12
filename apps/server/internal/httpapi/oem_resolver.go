package httpapi

import (
	"net/http"
	"net/url"
	"strings"

	"openclaw/platformapi/internal/models"
)

func (r *Router) resolveOEM(req *http.Request) map[string]any {
	domain := normalizeOEMRequestHost(req)

	if domain != "" {
		if brand := r.findBrandByDomain(domain); brand != nil {
			return map[string]any{
				"brand":    brand,
				"theme":    r.findBrandTheme(brand.ID),
				"features": r.findBrandFeatures(brand.ID),
				"binding":  r.findBrandBindingByBrand(brand.ID),
			}
		}
	}

	tenantID := tenantFilterID(req, 1)
	binding := r.findBrandBindingByTenant(tenantID)
	if binding == nil {
		return map[string]any{}
	}
	brand := r.findBrand(binding.BrandID)
	if brand == nil {
		return map[string]any{}
	}

	return map[string]any{
		"brand":    brand,
		"theme":    r.findBrandTheme(binding.BrandID),
		"features": r.findBrandFeatures(binding.BrandID),
		"binding":  binding,
	}
}

func normalizeOEMRequestHost(req *http.Request) string {
	if req == nil {
		return ""
	}

	raw := strings.TrimSpace(req.URL.Query().Get("domain"))
	if raw == "" {
		forwardedHost := strings.TrimSpace(req.Header.Get("X-Forwarded-Host"))
		if forwardedHost != "" {
			raw = strings.TrimSpace(strings.Split(forwardedHost, ",")[0])
		}
	}
	if raw == "" {
		raw = strings.TrimSpace(req.Host)
	}
	if raw == "" {
		return ""
	}

	candidate := raw
	if !strings.Contains(candidate, "://") {
		candidate = "//" + candidate
	}
	parsed, err := url.Parse(candidate)
	if err == nil && strings.TrimSpace(parsed.Hostname()) != "" {
		return strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	}
	return strings.ToLower(strings.TrimSpace(raw))
}

func (r *Router) findBrandByDomain(domain string) *models.OEMBrand {
	normalized := strings.TrimSpace(strings.ToLower(domain))
	for _, brand := range r.data.Brands {
		if !strings.EqualFold(strings.TrimSpace(brand.Status), "active") {
			continue
		}
		for _, item := range brand.Domains {
			value := strings.ToLower(strings.TrimSpace(item))
			switch {
			case value == normalized:
				copy := brand
				return &copy
			case strings.HasPrefix(value, "*.") && strings.HasSuffix(normalized, strings.TrimPrefix(value, "*")):
				copy := brand
				return &copy
			}
		}
	}
	return nil
}

func (r *Router) findBrandBindingByBrand(brandID int) *models.TenantBrandBinding {
	for _, item := range r.data.BrandBindings {
		if item.BrandID == brandID {
			copy := item
			return &copy
		}
	}
	return nil
}
