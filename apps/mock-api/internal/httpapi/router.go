package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"openclaw/mockapi/internal/models"
	"openclaw/mockapi/internal/runtimeadapter"
)

type Router struct {
	mu      sync.RWMutex
	data    models.Data
	config  ExternalConfig
	runtime runtimeadapter.Adapter
}

func NewRouter(data models.Data, cfg ...ExternalConfig) http.Handler {
	router := &Router{
		data:    data,
		runtime: runtimeadapter.NewMockAdapter(),
	}
	if len(cfg) > 0 {
		router.config = cfg[0]
	}
	return router
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	setCORSHeaders(w)

	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := req.URL.Path
	switch {
	case path == "/healthz" && req.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case path == "/api/v1/auth/config" && req.Method == http.MethodGet:
		r.handleAuthConfig(w, req)
	case path == "/api/v1/auth/session" && req.Method == http.MethodGet:
		r.handleAuthSession(w, req)
	case path == "/api/v1/auth/me" && req.Method == http.MethodGet:
		r.handleAuthMe(w, req)
	case path == "/api/v1/auth/keycloak/url" && req.Method == http.MethodGet:
		r.handleAuthKeycloakURL(w, req)
	case path == "/api/v1/auth/callback" && req.Method == http.MethodGet:
		r.handleAuthCallback(w, req)
	case path == "/api/v1/auth/token" && req.Method == http.MethodPost:
		r.handleAuthToken(w, req)
	case path == "/api/v1/auth/refresh" && req.Method == http.MethodPost:
		r.handleAuthRefresh(w, req)
	case path == "/api/v1/auth/logout" && req.Method == http.MethodPost:
		r.handleAuthLogout(w, req)
	case path == "/api/v1/oem/config" && req.Method == http.MethodGet:
		r.handleOEMConfig(w, req)
	case path == "/api/v1/search/config" && req.Method == http.MethodGet:
		r.handleSearchConfig(w, req)
	case path == "/api/v1/search/logs" && req.Method == http.MethodGet:
		r.handleSearchLogs(w, req)
	case path == "/api/v1/runtime/clusters" && req.Method == http.MethodGet:
		r.handleRuntimeClusters(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/clusters/") && strings.HasSuffix(path, "/nodes") && req.Method == http.MethodGet:
		r.handleRuntimeClusterNodes(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/clusters/") && strings.HasSuffix(path, "/namespaces") && req.Method == http.MethodGet:
		r.handleRuntimeClusterNamespaces(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/clusters/") && req.Method == http.MethodGet:
		r.handleRuntimeClusterDetail(w, req)
	case path == "/api/v1/runtime/workloads" && req.Method == http.MethodGet:
		r.handleRuntimeWorkloads(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/pods") && req.Method == http.MethodGet:
		r.handleRuntimeWorkloadPods(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/metrics") && req.Method == http.MethodGet:
		r.handleRuntimeWorkloadMetrics(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/start") && req.Method == http.MethodPost:
		r.handleRuntimeWorkloadAction(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/stop") && req.Method == http.MethodPost:
		r.handleRuntimeWorkloadAction(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/restart") && req.Method == http.MethodPost:
		r.handleRuntimeWorkloadAction(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/scale") && req.Method == http.MethodPost:
		r.handleRuntimeWorkloadScale(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && req.Method == http.MethodGet:
		r.handleRuntimeWorkloadDetail(w, req)
	case path == "/api/v1/portal/overview" && req.Method == http.MethodGet:
		r.handlePortalOverview(w, req)
	case path == "/api/v1/portal/instances" && req.Method == http.MethodGet:
		r.handlePortalInstances(w, req)
	case path == "/api/v1/portal/instances" && req.Method == http.MethodPost:
		r.handleCreatePortalInstance(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/config") && req.Method == http.MethodPatch:
		r.handleUpdatePortalInstanceConfig(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/backups") && req.Method == http.MethodPost:
		r.handleTriggerPortalBackup(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/runtime") && req.Method == http.MethodGet:
		r.handlePortalInstanceRuntime(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/start") && req.Method == http.MethodPost:
		r.handlePortalInstancePower(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/stop") && req.Method == http.MethodPost:
		r.handlePortalInstancePower(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/restart") && req.Method == http.MethodPost:
		r.handlePortalInstancePower(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && req.Method == http.MethodGet:
		r.handlePortalInstanceDetail(w, req)
	case path == "/api/v1/portal/jobs" && req.Method == http.MethodGet:
		r.handlePortalJobs(w, req)
	case path == "/api/v1/portal/alerts" && req.Method == http.MethodGet:
		r.handlePortalAlerts(w, req)
	case path == "/api/v1/portal/logs" && req.Method == http.MethodGet:
		r.handlePortalLogs(w, req)
	case path == "/api/v1/portal/plans" && req.Method == http.MethodGet:
		r.handlePortalPlans(w, req)
	case path == "/api/v1/portal/purchases" && req.Method == http.MethodPost:
		r.handlePortalPurchase(w, req)
	case path == "/api/v1/portal/tickets" && req.Method == http.MethodGet:
		r.handlePortalTickets(w, req)
	case path == "/api/v1/portal/tickets" && req.Method == http.MethodPost:
		r.handlePortalCreateTicket(w, req)
	case path == "/api/v1/admin/overview" && req.Method == http.MethodGet:
		r.handleAdminOverview(w, req)
	case path == "/api/v1/admin/tenants" && req.Method == http.MethodGet:
		r.handleAdminTenants(w, req)
	case path == "/api/v1/admin/instances" && req.Method == http.MethodGet:
		r.handleAdminInstances(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && req.Method == http.MethodGet:
		r.handleAdminInstanceDetail(w, req)
	case path == "/api/v1/admin/jobs" && req.Method == http.MethodGet:
		r.handleAdminJobs(w, req)
	case path == "/api/v1/admin/alerts" && req.Method == http.MethodGet:
		r.handleAdminAlerts(w, req)
	case path == "/api/v1/admin/audit" && req.Method == http.MethodGet:
		r.handleAdminAudit(w, req)
	case path == "/api/v1/admin/oem/brands" && req.Method == http.MethodGet:
		r.handleAdminOEMBrands(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/oem/brands/") && req.Method == http.MethodGet:
		r.handleAdminOEMBrandDetail(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/oem/brands/") && strings.HasSuffix(path, "/theme") && req.Method == http.MethodPatch:
		r.handleAdminOEMBrandTheme(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/oem/brands/") && strings.HasSuffix(path, "/features") && req.Method == http.MethodPatch:
		r.handleAdminOEMBrandFeatures(w, req)
	case path == "/api/v1/admin/tickets" && req.Method == http.MethodGet:
		r.handleAdminTickets(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/tickets/") && strings.HasSuffix(path, "/status") && req.Method == http.MethodPatch:
		r.handleAdminUpdateTicketStatus(w, req)
	case path == "/api/v1/channels" && req.Method == http.MethodGet:
		r.handleChannelList(w, req)
	case strings.HasPrefix(path, "/api/v1/channels/") && strings.HasSuffix(path, "/activities") && req.Method == http.MethodGet:
		r.handleChannelActivities(w, req)
	case strings.HasPrefix(path, "/api/v1/channels/") && !strings.Contains(strings.TrimPrefix(path, "/api/v1/channels/"), "/") && req.Method == http.MethodGet:
		r.handleChannelDetail(w, req)
	case strings.HasPrefix(path, "/api/v1/channels/") && strings.HasSuffix(path, "/connect") && req.Method == http.MethodPost:
		r.handleChannelConnect(w, req)
	case strings.HasPrefix(path, "/api/v1/channels/") && strings.HasSuffix(path, "/disconnect") && req.Method == http.MethodPost:
		r.handleChannelDisconnect(w, req)
	case strings.HasPrefix(path, "/api/v1/channels/") && strings.HasSuffix(path, "/health") && req.Method == http.MethodPost:
		r.handleChannelHealthCheck(w, req)
	default:
		http.NotFound(w, req)
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func tenantFilterID(req *http.Request, defaultTenantID int) int {
	tenantParam := req.URL.Query().Get("tenantId")
	if tenantParam == "" {
		return defaultTenantID
	}
	if id, err := strconv.Atoi(tenantParam); err == nil {
		return id
	}
	return defaultTenantID
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Mock-Source", "openclaw")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func decodeJSON(req *http.Request, target any) error {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
