package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOEMConfigResolvesForwardedHost(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oem/config", nil)
	req.Header.Set("X-Forwarded-Host", "ai.acme.example.com:8443")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected oem config status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Brand struct {
			Code string `json:"code"`
		} `json:"brand"`
	}
	decodeResponse(t, recorder, &response)

	if response.Brand.Code != "acme-oem" {
		t.Fatalf("expected forwarded host to resolve acme-oem, got %q", response.Brand.Code)
	}
}

func TestPortalOpsReportIncludesNotificationAndUsage(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/ops/report", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ops report status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Summary struct {
			InstanceCount     int    `json:"instanceCount"`
			SessionCount      int    `json:"sessionCount"`
			WalletBalance     int    `json:"walletBalance"`
			Currency          string `json:"currency"`
			BillingMonthCount int    `json:"billingMonthCount"`
		} `json:"summary"`
		NotificationChannels []struct {
			Key     string `json:"key"`
			Enabled bool   `json:"enabled"`
		} `json:"notificationChannels"`
		NotificationTemplates []struct {
			Key     string `json:"key"`
			Subject string `json:"subject"`
		} `json:"notificationTemplates"`
		MonthlyUsage []struct {
			Label string `json:"label"`
		} `json:"monthlyUsage"`
		Export struct {
			CSVPath string `json:"csvPath"`
		} `json:"export"`
	}
	decodeResponse(t, recorder, &response)

	if response.Summary.InstanceCount == 0 || response.Summary.SessionCount == 0 {
		t.Fatalf("expected ops summary counts, got %#v", response.Summary)
	}
	if response.Summary.Currency != "CNY" || response.Summary.WalletBalance == 0 {
		t.Fatalf("expected wallet summary, got %#v", response.Summary)
	}
	if response.Summary.BillingMonthCount == 0 || len(response.MonthlyUsage) == 0 {
		t.Fatalf("expected monthly usage data, got %#v", response.MonthlyUsage)
	}
	if len(response.NotificationChannels) < 3 {
		t.Fatalf("expected notification channel cards, got %#v", response.NotificationChannels)
	}
	if len(response.NotificationTemplates) < 3 {
		t.Fatalf("expected notification template previews, got %#v", response.NotificationTemplates)
	}
	if response.Export.CSVPath != "/api/v1/portal/ops/report/export.csv" {
		t.Fatalf("unexpected export path %q", response.Export.CSVPath)
	}
}

func TestPortalOpsReportExportCSV(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/ops/report/export.csv", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ops report export status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, "text/csv") {
		t.Fatalf("expected csv content type, got %q", got)
	}
	if got := recorder.Header().Get("Content-Disposition"); !strings.Contains(got, "openclaw-portal-ops-") {
		t.Fatalf("expected csv attachment filename, got %q", got)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "billingMonth") {
		t.Fatalf("expected csv header in export body, got %s", body)
	}
	if !strings.Contains(body, "2026-04") {
		t.Fatalf("expected billing month row in export body, got %s", body)
	}
}

func TestAdminOEMBrandThemePatchUsesThemeHandler(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodPatch, "/api/v1/admin/oem/brands/1/theme", map[string]any{
		"primaryColor":   "#112233",
		"secondaryColor": "#445566",
		"accentColor":    "#778899",
		"surfaceMode":    "light",
		"fontFamily":     "Noto Sans SC",
		"radius":         "14px",
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected theme patch status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Theme struct {
			BrandID        int    `json:"brandId"`
			PrimaryColor   string `json:"primaryColor"`
			SecondaryColor string `json:"secondaryColor"`
			AccentColor    string `json:"accentColor"`
			SurfaceMode    string `json:"surfaceMode"`
			FontFamily     string `json:"fontFamily"`
			Radius         string `json:"radius"`
		} `json:"theme"`
	}
	decodeResponse(t, recorder, &response)

	if response.Theme.BrandID != 1 {
		t.Fatalf("expected theme response for brand 1, got %#v", response.Theme)
	}
	if response.Theme.PrimaryColor != "#112233" || response.Theme.FontFamily != "Noto Sans SC" {
		t.Fatalf("expected theme payload persisted, got %#v", response.Theme)
	}
	if router.data.BrandThemes[0].PrimaryColor != "#112233" || router.data.BrandThemes[0].SurfaceMode != "light" {
		t.Fatalf("expected theme stored on brand theme slice, got %#v", router.data.BrandThemes[0])
	}
}

func TestAdminOEMBrandFeaturesPatchUsesFeatureHandler(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodPatch, "/api/v1/admin/oem/brands/2/features", map[string]any{
		"portalEnabled":         true,
		"adminEnabled":          true,
		"channelsEnabled":       false,
		"ticketsEnabled":        false,
		"purchaseEnabled":       true,
		"runtimeControlEnabled": true,
		"auditEnabled":          true,
		"ssoEnabled":            true,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected features patch status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Features struct {
			BrandID               int  `json:"brandId"`
			PortalEnabled         bool `json:"portalEnabled"`
			AdminEnabled          bool `json:"adminEnabled"`
			ChannelsEnabled       bool `json:"channelsEnabled"`
			TicketsEnabled        bool `json:"ticketsEnabled"`
			PurchaseEnabled       bool `json:"purchaseEnabled"`
			RuntimeControlEnabled bool `json:"runtimeControlEnabled"`
			AuditEnabled          bool `json:"auditEnabled"`
			SSOEnabled            bool `json:"ssoEnabled"`
		} `json:"features"`
	}
	decodeResponse(t, recorder, &response)

	if response.Features.BrandID != 2 {
		t.Fatalf("expected feature response for brand 2, got %#v", response.Features)
	}
	if !response.Features.AdminEnabled || response.Features.ChannelsEnabled {
		t.Fatalf("expected feature payload persisted, got %#v", response.Features)
	}
	if !router.data.BrandFeatures[1].RuntimeControlEnabled || router.data.BrandFeatures[1].ChannelsEnabled {
		t.Fatalf("expected feature flags stored on brand feature slice, got %#v", router.data.BrandFeatures[1])
	}
}

func TestAdminOEMBrandNestedRoutes(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	themeBody, err := json.Marshal(map[string]any{
		"primaryColor":   "#123456",
		"secondaryColor": "#654321",
		"accentColor":    "#fedcba",
		"surfaceMode":    "dark",
		"fontFamily":     "Source Han Sans",
		"radius":         "24px",
	})
	if err != nil {
		t.Fatalf("marshal theme body: %v", err)
	}

	themeReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/oem/brands/2/theme", bytes.NewReader(themeBody))
	themeReq.Header.Set("Content-Type", "application/json")
	themeRecorder := httptest.NewRecorder()
	router.ServeHTTP(themeRecorder, themeReq)
	if themeRecorder.Code != http.StatusOK {
		t.Fatalf("expected theme update status 200, got %d: %s", themeRecorder.Code, themeRecorder.Body.String())
	}

	var themeResponse struct {
		Theme struct {
			BrandID        int    `json:"brandId"`
			PrimaryColor   string `json:"primaryColor"`
			SecondaryColor string `json:"secondaryColor"`
			AccentColor    string `json:"accentColor"`
			SurfaceMode    string `json:"surfaceMode"`
			FontFamily     string `json:"fontFamily"`
			Radius         string `json:"radius"`
		} `json:"theme"`
	}
	decodeResponse(t, themeRecorder, &themeResponse)
	if themeResponse.Theme.BrandID != 2 {
		t.Fatalf("expected theme brand id 2, got %d", themeResponse.Theme.BrandID)
	}
	if themeResponse.Theme.PrimaryColor != "#123456" || themeResponse.Theme.SurfaceMode != "dark" {
		t.Fatalf("unexpected theme response %#v", themeResponse.Theme)
	}

	featureBody, err := json.Marshal(map[string]any{
		"portalEnabled":         true,
		"adminEnabled":          true,
		"channelsEnabled":       false,
		"ticketsEnabled":        false,
		"purchaseEnabled":       true,
		"runtimeControlEnabled": true,
		"auditEnabled":          true,
		"ssoEnabled":            false,
	})
	if err != nil {
		t.Fatalf("marshal features body: %v", err)
	}

	featureReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/oem/brands/2/features", bytes.NewReader(featureBody))
	featureReq.Header.Set("Content-Type", "application/json")
	featureRecorder := httptest.NewRecorder()
	router.ServeHTTP(featureRecorder, featureReq)
	if featureRecorder.Code != http.StatusOK {
		t.Fatalf("expected features update status 200, got %d: %s", featureRecorder.Code, featureRecorder.Body.String())
	}

	var featureResponse struct {
		Features struct {
			BrandID               int  `json:"brandId"`
			PortalEnabled         bool `json:"portalEnabled"`
			AdminEnabled          bool `json:"adminEnabled"`
			ChannelsEnabled       bool `json:"channelsEnabled"`
			TicketsEnabled        bool `json:"ticketsEnabled"`
			PurchaseEnabled       bool `json:"purchaseEnabled"`
			RuntimeControlEnabled bool `json:"runtimeControlEnabled"`
			AuditEnabled          bool `json:"auditEnabled"`
			SSOEnabled            bool `json:"ssoEnabled"`
		} `json:"features"`
	}
	decodeResponse(t, featureRecorder, &featureResponse)
	if featureResponse.Features.BrandID != 2 {
		t.Fatalf("expected features brand id 2, got %d", featureResponse.Features.BrandID)
	}
	if !featureResponse.Features.AdminEnabled || featureResponse.Features.ChannelsEnabled || featureResponse.Features.SSOEnabled {
		t.Fatalf("unexpected features response %#v", featureResponse.Features)
	}

	bindingRecorder := performRequest(t, router, http.MethodPut, "/api/v1/admin/oem/brands/2/bindings", map[string]any{
		"bindings": []map[string]any{
			{
				"tenantId":    2,
				"bindingMode": "shared",
			},
		},
	})
	if bindingRecorder.Code != http.StatusOK {
		t.Fatalf("expected bindings update status 200, got %d: %s", bindingRecorder.Code, bindingRecorder.Body.String())
	}

	var bindingResponse struct {
		Bindings []struct {
			TenantID    int    `json:"tenantId"`
			BrandID     int    `json:"brandId"`
			BindingMode string `json:"bindingMode"`
		} `json:"bindings"`
	}
	decodeResponse(t, bindingRecorder, &bindingResponse)
	if len(bindingResponse.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %#v", bindingResponse.Bindings)
	}
	if bindingResponse.Bindings[0].TenantID != 2 || bindingResponse.Bindings[0].BrandID != 2 || bindingResponse.Bindings[0].BindingMode != "shared" {
		t.Fatalf("unexpected binding response %#v", bindingResponse.Bindings[0])
	}

	resolvedRecorder := performRequest(t, router, http.MethodGet, "/api/v1/oem/config?tenantId=2", nil)
	if resolvedRecorder.Code != http.StatusOK {
		t.Fatalf("expected tenant 2 oem config status 200, got %d: %s", resolvedRecorder.Code, resolvedRecorder.Body.String())
	}

	var resolvedResponse struct {
		Brand struct {
			ID   int    `json:"id"`
			Code string `json:"code"`
		} `json:"brand"`
		Binding struct {
			TenantID int `json:"tenantId"`
			BrandID  int `json:"brandId"`
		} `json:"binding"`
	}
	decodeResponse(t, resolvedRecorder, &resolvedResponse)
	if resolvedResponse.Brand.ID != 2 || resolvedResponse.Brand.Code != "acme-oem" {
		t.Fatalf("expected tenant 2 to resolve acme-oem, got %#v", resolvedResponse.Brand)
	}
	if resolvedResponse.Binding.TenantID != 2 || resolvedResponse.Binding.BrandID != 2 {
		t.Fatalf("unexpected tenant 2 binding %#v", resolvedResponse.Binding)
	}
}
