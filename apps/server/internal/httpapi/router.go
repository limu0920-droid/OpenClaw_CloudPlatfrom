package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

type Router struct {
	mu              sync.RWMutex
	data            models.Data
	config          ExternalConfig
	runtime         runtimeadapter.Adapter
	store           corestore.CoreStore
	artifactStore   artifactObjectStore
	workspaceEvents *workspaceEventBroker
	metrics         *Metrics
}

func NewRouter(data models.Data, cfg ...ExternalConfig) (http.Handler, error) {
	router := &Router{
		data: data,
	}
	if len(cfg) > 0 {
		router.config = cfg[0]
	}
	router.store = normalizeCoreStore(router.config.CoreStore)
	artifactStore, err := newArtifactObjectStore(router.config)
	if err != nil {
		return nil, err
	}
	router.artifactStore = artifactStore
	runtime, err := runtimeadapter.NewAdapter(runtimeadapter.Config{
		StrictMode:      router.config.StrictMode,
		Provider:        router.config.RuntimeProvider,
		KubectlBinary:   router.config.RuntimeKubectlBinary,
		KubeContext:     router.config.RuntimeKubeContext,
		NamespacePrefix: router.config.RuntimeNamespacePrefix,
		WorkloadPrefix:  router.config.RuntimeWorkloadPrefix,
		Image:           router.config.RuntimeImage,
		ServiceType:     router.config.RuntimeServiceType,
		AccessHost:      router.config.RuntimeAccessHost,
		AccessScheme:    router.config.RuntimeAccessScheme,
		Port:            router.config.RuntimePort,
	})
	if err != nil {
		return nil, err
	}
	router.runtime = runtime
	router.metrics = NewMetrics()
	router.workspaceEvents = newWorkspaceEventBroker()
	if err := router.validateStartupDependencies(); err != nil {
		return nil, err
	}
	return router, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	observedWriter := newStatusCapturingResponseWriter(w)
	startedAt := time.Now()
	if r.metrics != nil && req.URL.Path != "/metrics" {
		r.metrics.inFlightRequests.Inc()
		defer func() {
			r.metrics.inFlightRequests.Dec()
			r.metrics.Observe(req, observedWriter.statusCode, time.Since(startedAt))
		}()
	}
	w = observedWriter

	setCORSHeaders(observedWriter)
	r.setServiceHeaders(observedWriter)

	if req.Method == http.MethodOptions {
		observedWriter.WriteHeader(http.StatusNoContent)
		return
	}

	path := req.URL.Path
	switch {
	case path == "/healthz" && req.Method == http.MethodGet:
		r.handleHealthz(w, req)
	case path == "/readyz" && req.Method == http.MethodGet:
		r.handleReadyz(observedWriter, req)
	case path == "/versionz" && req.Method == http.MethodGet:
		r.handleVersionz(observedWriter, req)
	case path == "/api/v1/docs/external" && req.Method == http.MethodGet:
		r.handleExternalDocsIndex(observedWriter, req)
	case path == "/api/v1/docs/external/openapi.yaml" && req.Method == http.MethodGet:
		r.handleExternalDocOpenAPI(observedWriter, req)
	case path == "/api/v1/docs/external/integration.md" && req.Method == http.MethodGet:
		r.handleExternalDocGuide(observedWriter, req)
	case path == "/metrics" && req.Method == http.MethodGet:
		r.metrics.Handler().ServeHTTP(observedWriter, req)
	case path == "/api/v1/bootstrap" && req.Method == http.MethodGet:
		r.handleBootstrap(observedWriter, req)
	case path == "/api/v1/i18n/config" && req.Method == http.MethodGet:
		r.handleI18nConfig(w, req)
	case path == "/api/v1/auth/providers" && req.Method == http.MethodGet:
		r.handleAuthProviders(w, req)
	case path == "/api/v1/auth/config" && req.Method == http.MethodGet:
		r.handleAuthConfig(w, req)
	case path == "/api/v1/auth/session" && req.Method == http.MethodGet:
		r.handleAuthSession(w, req)
	case path == "/api/v1/auth/me" && req.Method == http.MethodGet:
		r.handleAuthMe(w, req)
	case path == "/api/v1/auth/keycloak/url" && req.Method == http.MethodGet:
		r.handleAuthKeycloakURL(w, req)
	case path == "/api/v1/auth/wechat/url" && req.Method == http.MethodGet:
		r.handleAuthWechatURL(w, req)
	case path == "/api/v1/auth/callback" && req.Method == http.MethodGet:
		r.handleAuthCallback(w, req)
	case path == "/api/v1/auth/wechat/callback" && req.Method == http.MethodGet:
		r.handleAuthWechatCallback(w, req)
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
	case path == "/api/v1/search/logs/export.csv" && req.Method == http.MethodGet:
		r.handleSearchLogsExportCSV(w, req)
	case strings.HasPrefix(path, "/api/v1/search/traces/") && req.Method == http.MethodGet:
		r.handleSearchTrace(w, req)
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
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/pause") && req.Method == http.MethodPost:
		r.handleRuntimeWorkloadAction(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/restart") && req.Method == http.MethodPost:
		r.handleRuntimeWorkloadAction(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && strings.HasSuffix(path, "/scale") && req.Method == http.MethodPost:
		r.handleRuntimeWorkloadScale(w, req)
	case strings.HasPrefix(path, "/api/v1/runtime/workloads/") && req.Method == http.MethodGet:
		r.handleRuntimeWorkloadDetail(w, req)
	case path == "/api/v1/portal/overview" && req.Method == http.MethodGet:
		r.handlePortalOverview(w, req)
	case path == "/api/v1/portal/self-service" && req.Method == http.MethodGet:
		r.handlePortalSelfService(w, req)
	case path == "/api/v1/portal/artifacts" && req.Method == http.MethodGet:
		r.handlePortalArtifacts(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/artifacts/") && strings.HasSuffix(path, "/favorite") && (req.Method == http.MethodPost || req.Method == http.MethodDelete):
		r.handlePortalArtifactFavorite(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/artifacts/") && strings.HasSuffix(path, "/shares") && req.Method == http.MethodPost:
		r.handlePortalArtifactShares(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/artifact-shares/") && req.Method == http.MethodDelete:
		r.handlePortalArtifactShareDelete(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/artifacts/") && req.Method == http.MethodGet:
		r.handlePortalArtifactDetail(w, req)
	case path == "/api/v1/portal/instances" && req.Method == http.MethodGet:
		r.handlePortalInstances(w, req)
	case path == "/api/v1/portal/instances" && req.Method == http.MethodPost:
		r.handleCreatePortalInstance(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && req.Method == http.MethodDelete:
		r.handleDeletePortalInstance(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/config") && req.Method == http.MethodPatch:
		r.handleUpdatePortalInstanceConfig(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/backups") && req.Method == http.MethodPost:
		r.handleTriggerPortalBackup(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/runtime") && req.Method == http.MethodGet:
		r.handlePortalInstanceRuntime(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/workspace/sessions") && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handlePortalWorkspaceSessions(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/workspace/bridge-health") && req.Method == http.MethodGet:
		r.handlePortalWorkspaceBridgeHealth(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/start") && req.Method == http.MethodPost:
		r.handlePortalInstancePower(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/stop") && req.Method == http.MethodPost:
		r.handlePortalInstancePower(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/instances/") && strings.HasSuffix(path, "/pause") && req.Method == http.MethodPost:
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
	case path == "/api/v1/portal/orders" && req.Method == http.MethodGet:
		r.handlePortalOrders(w, req)
	case path == "/api/v1/portal/orders" && req.Method == http.MethodPost:
		r.handlePortalCreateOrder(w, req)
	case path == "/api/v1/portal/approvals" && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handlePortalApprovals(w, req)
	case path == "/api/v1/portal/profile" && req.Method == http.MethodGet:
		r.handlePortalProfile(w, req)
	case path == "/api/v1/portal/profile" && req.Method == http.MethodPatch:
		r.handlePortalUpdateProfile(w, req)
	case path == "/api/v1/portal/profile/password" && req.Method == http.MethodPost:
		r.handlePortalUpdatePassword(w, req)
	case path == "/api/v1/portal/account/settings" && req.Method == http.MethodGet:
		r.handlePortalAccountSettings(w, req)
	case path == "/api/v1/portal/account/settings" && req.Method == http.MethodPatch:
		r.handlePortalUpdateAccountSettings(w, req)
	case path == "/api/v1/portal/ops/report" && req.Method == http.MethodGet:
		r.handlePortalOpsReport(w, req)
	case path == "/api/v1/portal/ops/report/export.csv" && req.Method == http.MethodGet:
		r.handlePortalOpsReportExport(w, req)
	case path == "/api/v1/portal/account/wallet" && req.Method == http.MethodGet:
		r.handlePortalWallet(w, req)
	case path == "/api/v1/portal/billing/history" && req.Method == http.MethodGet:
		r.handlePortalBillingHistory(w, req)
	case path == "/api/v1/portal/subscriptions" && req.Method == http.MethodGet:
		r.handlePortalSubscriptions(w, req)
	case path == "/api/v1/portal/invoices" && req.Method == http.MethodGet:
		r.handlePortalInvoices(w, req)
	case path == "/api/v1/portal/invoices" && req.Method == http.MethodPost:
		r.handlePortalCreateInvoice(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/orders/") && strings.HasSuffix(path, "/pay") && req.Method == http.MethodPost:
		r.handlePortalOrderPay(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/orders/") && strings.HasSuffix(path, "/refunds") && req.Method == http.MethodPost:
		r.handlePortalOrderRefund(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/orders/") && req.Method == http.MethodGet:
		r.handlePortalOrderDetail(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/subscriptions/") && strings.HasSuffix(path, "/renew") && req.Method == http.MethodPost:
		r.handlePortalSubscriptionAction(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/subscriptions/") && strings.HasSuffix(path, "/upgrade") && req.Method == http.MethodPost:
		r.handlePortalSubscriptionAction(w, req)
	case path == "/api/v1/portal/tickets" && req.Method == http.MethodGet:
		r.handlePortalTickets(w, req)
	case path == "/api/v1/portal/tickets" && req.Method == http.MethodPost:
		r.handlePortalCreateTicket(w, req)
	case path == "/api/v1/platform/workspace/report" && req.Method == http.MethodPost:
		r.handleWorkspaceBridgeReport(w, req)
	case path == "/api/v1/portal/workspace/sessions" && req.Method == http.MethodGet:
		r.handlePortalWorkspaceSessionIndex(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/sessions/") && strings.HasSuffix(path, "/status") && req.Method == http.MethodPatch:
		r.handlePortalWorkspaceSessionStatus(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/sessions/") && strings.HasSuffix(path, "/artifacts") && req.Method == http.MethodPost:
		r.handlePortalWorkspaceArtifacts(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/messages/") && strings.HasSuffix(path, "/retry") && req.Method == http.MethodPost:
		r.handlePortalWorkspaceMessageRetry(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/artifacts/") && strings.HasSuffix(path, "/preview") && req.Method == http.MethodGet:
		r.handlePortalWorkspaceArtifactPreview(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/sessions/") && strings.HasSuffix(path, "/events") && req.Method == http.MethodGet:
		r.handlePortalWorkspaceEvents(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/sessions/") && strings.HasSuffix(path, "/messages/stream") && req.Method == http.MethodPost:
		r.handlePortalWorkspaceMessageStream(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/sessions/") && strings.HasSuffix(path, "/messages") && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handlePortalWorkspaceMessages(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/artifacts/") && strings.HasSuffix(path, "/preview-content") && (req.Method == http.MethodGet || req.Method == http.MethodHead):
		r.handlePortalWorkspaceArtifactPreviewContent(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/artifacts/") && strings.HasSuffix(path, "/download") && (req.Method == http.MethodGet || req.Method == http.MethodHead):
		r.handlePortalWorkspaceArtifactDownload(w, req)
	case strings.HasPrefix(path, "/api/v1/portal/workspace/sessions/") && req.Method == http.MethodGet:
		r.handlePortalWorkspaceSessionDetail(w, req)
	case path == "/api/v1/admin/overview" && req.Method == http.MethodGet:
		r.handleAdminOverview(w, req)
	case path == "/api/v1/admin/tenants" && req.Method == http.MethodGet:
		r.handleAdminTenants(w, req)
	case path == "/api/v1/admin/users" && req.Method == http.MethodGet:
		r.handleAdminUsers(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/accounts/") && strings.HasSuffix(path, "/portrait") && req.Method == http.MethodGet:
		r.handleAdminAccountPortrait(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/users/") && strings.HasSuffix(path, "/identities") && req.Method == http.MethodPost:
		r.handleAdminBindIdentity(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/users/") && strings.Contains(path, "/identities/") && strings.HasSuffix(path, "/primary") && req.Method == http.MethodPost:
		r.handleAdminSetPrimaryIdentity(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/users/") && strings.Contains(path, "/identities/") && strings.HasSuffix(path, "/status") && req.Method == http.MethodPatch:
		r.handleAdminIdentityStatus(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/users/") && strings.Contains(path, "/identities/") && strings.HasSuffix(path, "/unbind") && req.Method == http.MethodPost:
		r.handleAdminUnbindIdentity(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/users/") && strings.HasSuffix(path, "/status") && req.Method == http.MethodPatch:
		r.handleAdminUserStatus(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/users/") && req.Method == http.MethodPatch:
		r.handleAdminUpdateUser(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/users/") && req.Method == http.MethodGet:
		r.handleAdminUserDetail(w, req)
	case path == "/api/v1/admin/auth-identities" && req.Method == http.MethodGet:
		r.handleAdminAuthIdentities(w, req)
	case path == "/api/v1/admin/instances" && req.Method == http.MethodGet:
		r.handleAdminInstances(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && req.Method == http.MethodDelete:
		r.handleDeleteAdminInstance(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/runtime/start") && req.Method == http.MethodPost:
		r.handleAdminInstanceRuntimeAction(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/runtime/stop") && req.Method == http.MethodPost:
		r.handleAdminInstanceRuntimeAction(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/runtime/pause") && req.Method == http.MethodPost:
		r.handleAdminInstanceRuntimeAction(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/runtime/restart") && req.Method == http.MethodPost:
		r.handleAdminInstanceRuntimeAction(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/runtime/scale") && req.Method == http.MethodPost:
		r.handleAdminInstanceRuntimeScale(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/runtime") && req.Method == http.MethodGet:
		r.handleAdminInstanceRuntime(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/diagnostics") && req.Method == http.MethodGet:
		r.handleAdminInstanceDiagnostics(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/diagnostic-sessions") && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handleAdminDiagnosticSessions(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/workspace/sessions") && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handleAdminWorkspaceSessions(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/workspace/bridge-health") && req.Method == http.MethodGet:
		r.handleAdminWorkspaceBridgeHealth(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && req.Method == http.MethodGet:
		r.handleAdminInstanceDetail(w, req)
	case path == "/api/v1/admin/artifacts" && req.Method == http.MethodGet:
		r.handleAdminArtifacts(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/artifacts/") && strings.HasSuffix(path, "/favorite") && (req.Method == http.MethodPost || req.Method == http.MethodDelete):
		r.handleAdminArtifactFavorite(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/artifacts/") && strings.HasSuffix(path, "/shares") && req.Method == http.MethodPost:
		r.handleAdminArtifactShares(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/artifact-shares/") && req.Method == http.MethodDelete:
		r.handleAdminArtifactShareDelete(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/artifacts/") && req.Method == http.MethodGet:
		r.handleAdminArtifactDetail(w, req)
	case path == "/api/v1/admin/workspace/sessions" && req.Method == http.MethodGet:
		r.handleAdminWorkspaceSessionIndex(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/sessions/") && strings.HasSuffix(path, "/status") && req.Method == http.MethodPatch:
		r.handleAdminWorkspaceSessionStatus(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/sessions/") && strings.HasSuffix(path, "/artifacts") && req.Method == http.MethodPost:
		r.handleAdminWorkspaceArtifacts(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/messages/") && strings.HasSuffix(path, "/retry") && req.Method == http.MethodPost:
		r.handleAdminWorkspaceMessageRetry(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/artifacts/") && strings.HasSuffix(path, "/preview") && req.Method == http.MethodGet:
		r.handleAdminWorkspaceArtifactPreview(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/sessions/") && strings.HasSuffix(path, "/events") && req.Method == http.MethodGet:
		r.handleAdminWorkspaceEvents(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/sessions/") && strings.HasSuffix(path, "/messages/stream") && req.Method == http.MethodPost:
		r.handleAdminWorkspaceMessageStream(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/sessions/") && strings.HasSuffix(path, "/messages") && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handleAdminWorkspaceMessages(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/artifacts/") && strings.HasSuffix(path, "/preview-content") && (req.Method == http.MethodGet || req.Method == http.MethodHead):
		r.handleAdminWorkspaceArtifactPreviewContent(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/artifacts/") && strings.HasSuffix(path, "/download") && (req.Method == http.MethodGet || req.Method == http.MethodHead):
		r.handleAdminWorkspaceArtifactDownload(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/workspace/sessions/") && req.Method == http.MethodGet:
		r.handleAdminWorkspaceSessionDetail(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && strings.HasSuffix(path, "/commands") && req.Method == http.MethodPost:
		r.handleAdminDiagnosticSessionCommands(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && strings.HasSuffix(path, "/close") && req.Method == http.MethodPost:
		r.handleAdminDiagnosticSessionClose(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && strings.HasSuffix(path, "/record") && req.Method == http.MethodGet:
		r.handleAdminDiagnosticSessionRecord(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && req.Method == http.MethodGet:
		r.handleAdminDiagnosticSessionDetail(w, req)
	case path == "/api/v1/admin/jobs" && req.Method == http.MethodGet:
		r.handleAdminJobs(w, req)
	case path == "/api/v1/admin/alerts" && req.Method == http.MethodGet:
		r.handleAdminAlerts(w, req)
	case path == "/api/v1/admin/audit" && req.Method == http.MethodGet:
		r.handleAdminAudit(w, req)
	case path == "/api/v1/admin/approvals" && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handleAdminApprovals(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/approvals/") && strings.HasSuffix(path, "/approve") && req.Method == http.MethodPost:
		r.handleAdminApprovalApprove(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/approvals/") && strings.HasSuffix(path, "/reject") && req.Method == http.MethodPost:
		r.handleAdminApprovalReject(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/approvals/") && strings.HasSuffix(path, "/execute") && req.Method == http.MethodPost:
		r.handleAdminApprovalExecute(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/approvals/") && req.Method == http.MethodGet:
		r.handleAdminApprovalDetail(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/diagnostics") && req.Method == http.MethodGet:
		r.handleAdminInstanceDiagnostics(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/diagnostic-sessions") && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handleAdminDiagnosticSessions(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/instances/") && strings.HasSuffix(path, "/terminal-sessions") && (req.Method == http.MethodGet || req.Method == http.MethodPost):
		r.handleAdminDiagnosticSessions(w, req)
	case path == "/api/v1/admin/terminal-sessions" && req.Method == http.MethodGet:
		r.handleAdminTerminalSessions(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && strings.HasSuffix(path, "/commands") && req.Method == http.MethodPost:
		r.handleAdminDiagnosticSessionCommands(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && strings.HasSuffix(path, "/close") && req.Method == http.MethodPost:
		r.handleAdminDiagnosticSessionClose(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && strings.HasSuffix(path, "/record") && req.Method == http.MethodGet:
		r.handleAdminDiagnosticSessionRecord(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/diagnostic-sessions/") && req.Method == http.MethodGet:
		r.handleAdminDiagnosticSessionDetail(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/terminal-sessions/") && strings.HasSuffix(path, "/commands") && req.Method == http.MethodPost:
		r.handleAdminDiagnosticSessionCommands(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/terminal-sessions/") && strings.HasSuffix(path, "/close") && req.Method == http.MethodPost:
		r.handleAdminDiagnosticSessionClose(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/terminal-sessions/") && strings.HasSuffix(path, "/record") && req.Method == http.MethodGet:
		r.handleAdminDiagnosticSessionRecord(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/terminal-sessions/") && req.Method == http.MethodGet:
		r.handleAdminDiagnosticSessionDetail(w, req)
	case path == "/api/v1/admin/payments" && req.Method == http.MethodGet:
		r.handleAdminPayments(w, req)
	case path == "/api/v1/admin/payment-callback-events" && req.Method == http.MethodGet:
		r.handleAdminPaymentCallbackEvents(w, req)
	case path == "/api/v1/admin/refunds" && req.Method == http.MethodGet:
		r.handleAdminRefunds(w, req)
	case path == "/api/v1/admin/orders" && req.Method == http.MethodGet:
		r.handleAdminOrders(w, req)
	case path == "/api/v1/admin/subscriptions" && req.Method == http.MethodGet:
		r.handleAdminSubscriptions(w, req)
	case path == "/api/v1/admin/invoices" && req.Method == http.MethodGet:
		r.handleAdminInvoices(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/invoices/") && strings.HasSuffix(path, "/status") && req.Method == http.MethodPatch:
		r.handleAdminInvoiceStatus(w, req)
	case path == "/api/v1/admin/wallets" && req.Method == http.MethodGet:
		r.handleAdminWallets(w, req)
	case path == "/api/v1/admin/billing/history" && req.Method == http.MethodGet:
		r.handleAdminBillingHistory(w, req)
	case path == "/api/v1/internal/payments/reconcile/daily" && req.Method == http.MethodPost:
		r.handleInternalPaymentsReconcile(w, req)
	case path == "/api/v1/internal/payments/recover-missing-callback" && req.Method == http.MethodPost:
		r.handleInternalPaymentsRecover(w, req)
	case strings.HasPrefix(path, "/api/v1/internal/orders/") && strings.HasSuffix(path, "/close") && req.Method == http.MethodPost:
		r.handleInternalOrderClose(w, req)
	case strings.HasPrefix(path, "/api/v1/internal/orders/") && strings.HasSuffix(path, "/query") && req.Method == http.MethodPost:
		r.handleInternalOrderQuery(w, req)
	case path == "/api/v1/admin/oem/brands" && req.Method == http.MethodGet:
		r.handleAdminOEMBrands(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/oem/brands/") && strings.HasSuffix(path, "/theme") && req.Method == http.MethodPatch:
		r.handleAdminOEMBrandTheme(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/oem/brands/") && strings.HasSuffix(path, "/features") && req.Method == http.MethodPatch:
		r.handleAdminOEMBrandFeatures(w, req)
	case strings.HasPrefix(path, "/api/v1/admin/oem/brands/") && strings.HasSuffix(path, "/bindings") && req.Method == http.MethodPut:
		r.handleAdminOEMBrandBindings(w, req)
	case matchesTailIDPath(path, "/api/v1/admin/oem/brands/") && req.Method == http.MethodPatch:
		r.handleAdminOEMBrandUpdate(w, req)
	case matchesTailIDPath(path, "/api/v1/admin/oem/brands/") && req.Method == http.MethodGet:
		r.handleAdminOEMBrandDetail(w, req)
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
	case path == "/api/v1/callback/payment/wechatpay" && req.Method == http.MethodPost:
		r.handleWechatPayCallback(w, req)
	default:
		http.NotFound(observedWriter, req)
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
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
