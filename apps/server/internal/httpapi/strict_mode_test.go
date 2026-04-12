package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

func TestHandleSearchLogsStrictModeRejectsMockFallback(t *testing.T) {
	router := newTestRouter(ExternalConfig{StrictMode: true})
	cookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})

	recorder := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/search/logs?q=test&scope=portal", nil, cookie)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "opensearch search backend is required") {
		t.Fatalf("expected strict search error, got %s", recorder.Body.String())
	}
}

func TestHandleAuthWechatURLStrictModeRejectsMockFallback(t *testing.T) {
	router := newTestRouter(ExternalConfig{StrictMode: true})

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/auth/wechat/url?redirect_uri=http://localhost/callback", nil)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "wechat login is not configured") {
		t.Fatalf("expected strict wechat auth error, got %s", recorder.Body.String())
	}
}

func TestHandlePortalOrderPayStrictModeRejectsMockGateway(t *testing.T) {
	router := newTestRouter(ExternalConfig{StrictMode: true})

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/orders/2/pay", map[string]any{
		"channel": "wechatpay",
		"payMode": "native",
	})
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "wechatpay is not configured") {
		t.Fatalf("expected strict payment gateway error, got %s", recorder.Body.String())
	}
}

func TestHandleInternalOrderQueryStrictModeRejectsMockGateway(t *testing.T) {
	router := newTestRouter(ExternalConfig{StrictMode: true})
	router.data.Orders[1].Status = "paying"
	router.data.Payments[1].Status = "paying"

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/internal/orders/2/query", nil)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "wechatpay gateway is required") {
		t.Fatalf("expected strict order query error, got %s", recorder.Body.String())
	}
}

func TestValidateStartupDependenciesStrictModeRequiresReachableRuntime(t *testing.T) {
	router := &Router{
		config:  ExternalConfig{StrictMode: true},
		store:   &fakeCoreStore{},
		runtime: &fakeEmptyRuntimeAdapter{},
	}

	err := router.validateStartupDependencies()
	if err == nil {
		t.Fatal("expected startup dependency error")
	}
	if !strings.Contains(err.Error(), "runtime provider returned no reachable clusters") {
		t.Fatalf("expected runtime dependency error, got %v", err)
	}
}

func TestNewRouterRejectsUnknownRuntimeProvider(t *testing.T) {
	handler, err := NewRouter(models.Data{}, ExternalConfig{RuntimeProvider: "unknown"})
	if err == nil {
		t.Fatal("expected router init error")
	}
	if handler != nil {
		t.Fatalf("expected nil handler on error, got %#v", handler)
	}
}

type fakeEmptyRuntimeAdapter struct{}

func (f *fakeEmptyRuntimeAdapter) ListClusters() []runtimeadapter.Cluster { return nil }
func (f *fakeEmptyRuntimeAdapter) GetCluster(id string) (runtimeadapter.Cluster, bool) {
	return runtimeadapter.Cluster{}, false
}
func (f *fakeEmptyRuntimeAdapter) ListNodes(clusterID string) []runtimeadapter.Node { return nil }
func (f *fakeEmptyRuntimeAdapter) ListNamespaces(clusterID string) []runtimeadapter.Namespace {
	return nil
}
func (f *fakeEmptyRuntimeAdapter) ListWorkloads(clusterID string, namespace string) []runtimeadapter.Workload {
	return nil
}
func (f *fakeEmptyRuntimeAdapter) GetWorkload(id string) (runtimeadapter.Workload, bool) {
	return runtimeadapter.Workload{}, false
}
func (f *fakeEmptyRuntimeAdapter) ListPods(workloadID string) []runtimeadapter.Pod { return nil }
func (f *fakeEmptyRuntimeAdapter) GetMetrics(workloadID string) (runtimeadapter.WorkloadMetrics, bool) {
	return runtimeadapter.WorkloadMetrics{}, false
}
func (f *fakeEmptyRuntimeAdapter) StartWorkload(id string) (runtimeadapter.ActionResult, bool) {
	return runtimeadapter.ActionResult{}, false
}
func (f *fakeEmptyRuntimeAdapter) StopWorkload(id string) (runtimeadapter.ActionResult, bool) {
	return runtimeadapter.ActionResult{}, false
}
func (f *fakeEmptyRuntimeAdapter) RestartWorkload(id string) (runtimeadapter.ActionResult, bool) {
	return runtimeadapter.ActionResult{}, false
}
func (f *fakeEmptyRuntimeAdapter) ScaleWorkload(id string, replicas int) (runtimeadapter.ActionResult, bool) {
	return runtimeadapter.ActionResult{}, false
}
