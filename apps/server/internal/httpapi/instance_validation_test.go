package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"openclaw/platformapi/internal/models"
)

func TestHandleCreatePortalInstanceRequiresExistingPlan(t *testing.T) {
	router := newTestRouter(ExternalConfig{})
	router.data.PlanOffers = []models.PlanOffer{}

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances", map[string]any{
		"name":   "生产实例",
		"plan":   "pro",
		"region": "cn-shanghai",
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "plan not found") {
		t.Fatalf("expected plan validation error, got %s", recorder.Body.String())
	}
}

func TestHandleCreatePortalInstanceRequiresExistingClusterForRegion(t *testing.T) {
	router := newTestRouter(ExternalConfig{})
	router.data.Clusters = []models.Cluster{}

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances", map[string]any{
		"name":   "生产实例",
		"plan":   "pro",
		"region": "cn-shanghai",
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "cluster not found for region") {
		t.Fatalf("expected cluster validation error, got %s", recorder.Body.String())
	}
}
