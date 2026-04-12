package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"openclaw/platformapi/internal/buildinfo"
	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/httpapi"
	"openclaw/platformapi/internal/models"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if configured := strings.TrimSpace(os.Getenv("PLATFORM_STRICT_MODE")); configured != "" && !envBool("PLATFORM_STRICT_MODE", true) {
		log.Fatal("PLATFORM_STRICT_MODE=false is no longer supported; cmd/server now requires real backend dependencies")
	}

	autoMigrate := envBool("AUTO_MIGRATE", false)
	autoBootstrap := envBool("AUTO_BOOTSTRAP", false)
	if autoBootstrap {
		log.Fatal("AUTO_BOOTSTRAP is no longer supported in cmd/server; run go run ./cmd/bootstrap explicitly against the target database")
	}

	strictMode := true
	var dbStore *corestore.PostgresStore
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required; seeded mode has been removed from cmd/server")
	}

	var err error
	dbStore, err = corestore.Open(databaseURL)
	if err != nil {
		log.Fatalf("open database store: %v", err)
	}
	defer func() {
		if err := dbStore.Close(); err != nil {
			log.Printf("close database store: %v", err)
		}
	}()

	if autoMigrate {
		if err := dbStore.Migrate(context.Background()); err != nil {
			log.Fatalf("migrate database store: %v", err)
		}
	}

	data, err := dbStore.Load(models.Data{})
	if err != nil {
		log.Fatalf("initialize database store: %v", err)
	}

	runtimeProvider := normalizeRuntimeProvider(os.Getenv("RUNTIME_PROVIDER"))
	bootstrapStrategy := "database-load"
	stateBackend := "postgres"
	serviceMode := "persistent"

	cfg := httpapi.ExternalConfig{
		ServiceName:                    "openclaw-platform-api",
		StrictMode:                     strictMode,
		BuildVersion:                   buildinfo.Version,
		BuildCommit:                    buildinfo.Commit,
		BuildDate:                      buildinfo.BuildDate,
		CoreStore:                      dbStore,
		BootstrapStrategy:              bootstrapStrategy,
		RuntimeProvider:                runtimeProvider,
		RuntimeKubectlBinary:           os.Getenv("RUNTIME_KUBECTL_BINARY"),
		RuntimeKubeContext:             os.Getenv("RUNTIME_KUBE_CONTEXT"),
		RuntimeNamespacePrefix:         os.Getenv("RUNTIME_NAMESPACE_PREFIX"),
		RuntimeWorkloadPrefix:          os.Getenv("RUNTIME_WORKLOAD_PREFIX"),
		RuntimeImage:                   os.Getenv("RUNTIME_IMAGE"),
		RuntimeServiceType:             os.Getenv("RUNTIME_SERVICE_TYPE"),
		RuntimeAccessHost:              os.Getenv("RUNTIME_ACCESS_HOST"),
		RuntimeAccessScheme:            os.Getenv("RUNTIME_ACCESS_SCHEME"),
		RuntimePort:                    envInt("RUNTIME_PORT", 0),
		KeycloakEnabled:                os.Getenv("KEYCLOAK_ENABLED") == "true",
		KeycloakBaseURL:                os.Getenv("KEYCLOAK_BASE_URL"),
		KeycloakRealm:                  os.Getenv("KEYCLOAK_REALM"),
		KeycloakClientID:               os.Getenv("KEYCLOAK_CLIENT_ID"),
		KeycloakClientSecret:           os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		KeycloakRedirectURL:            os.Getenv("KEYCLOAK_REDIRECT_URL"),
		KeycloakPostLoginRedirectURL:   os.Getenv("KEYCLOAK_POST_LOGIN_REDIRECT_URL"),
		KeycloakLogoutRedirectURL:      os.Getenv("KEYCLOAK_LOGOUT_REDIRECT_URL"),
		KeycloakMockUser:               os.Getenv("KEYCLOAK_MOCK_USER"),
		KeycloakMockEmail:              os.Getenv("KEYCLOAK_MOCK_EMAIL"),
		KeycloakMockRole:               os.Getenv("KEYCLOAK_MOCK_ROLE"),
		KeycloakSessionSecret:          os.Getenv("KEYCLOAK_SESSION_SECRET"),
		KeycloakCookieName:             os.Getenv("KEYCLOAK_COOKIE_NAME"),
		KeycloakCookieSecure:           os.Getenv("KEYCLOAK_COOKIE_SECURE") == "true",
		KeycloakFlowCookieName:         os.Getenv("KEYCLOAK_FLOW_COOKIE_NAME"),
		OpenSearchEnabled:              os.Getenv("OPENSEARCH_ENABLED") == "true",
		OpenSearchURL:                  os.Getenv("OPENSEARCH_URL"),
		OpenSearchIndex:                os.Getenv("OPENSEARCH_INDEX"),
		OpenSearchUsername:             os.Getenv("OPENSEARCH_USERNAME"),
		OpenSearchPassword:             os.Getenv("OPENSEARCH_PASSWORD"),
		TraceSearchEnabled:             os.Getenv("TRACE_SEARCH_ENABLED") == "true",
		TraceSearchURL:                 os.Getenv("TRACE_SEARCH_URL"),
		TraceSearchIndex:               os.Getenv("TRACE_SEARCH_INDEX"),
		TraceSearchUsername:            os.Getenv("TRACE_SEARCH_USERNAME"),
		TraceSearchPassword:            os.Getenv("TRACE_SEARCH_PASSWORD"),
		TraceSearchPublicBaseURL:       os.Getenv("TRACE_SEARCH_PUBLIC_BASE_URL"),
		ObjectStorageEndpoint:          os.Getenv("OBJECT_STORAGE_ENDPOINT"),
		ObjectStorageBucket:            os.Getenv("OBJECT_STORAGE_BUCKET"),
		ObjectStorageAccessKey:         os.Getenv("OBJECT_STORAGE_ACCESS_KEY"),
		ObjectStorageSecretKey:         os.Getenv("OBJECT_STORAGE_SECRET_KEY"),
		ObjectStorageRegion:            os.Getenv("OBJECT_STORAGE_REGION"),
		ObjectStorageForcePathStyle:    envBool("OBJECT_STORAGE_FORCE_PATH_STYLE", false),
		ArtifactArchiveMaxBytes:        envInt64("ARTIFACT_ARCHIVE_MAX_BYTES", 52_428_800),
		ArtifactArchiveAllowPrivateURL: envBool("ARTIFACT_ARCHIVE_ALLOW_PRIVATE_URL", false),
		WeChatPayEnabled:               os.Getenv("WECHATPAY_ENABLED") == "true",
		WeChatPayBaseURL:               os.Getenv("WECHATPAY_BASE_URL"),
		WeChatPayMchID:                 os.Getenv("WECHATPAY_MCH_ID"),
		WeChatPayAppID:                 os.Getenv("WECHATPAY_APP_ID"),
		WeChatPayClientSecret:          os.Getenv("WECHATPAY_CLIENT_SECRET"),
		WeChatPayNotifyURL:             os.Getenv("WECHATPAY_NOTIFY_URL"),
		WeChatPayRefundNotifyURL:       os.Getenv("WECHATPAY_REFUND_NOTIFY_URL"),
		WeChatPaySerialNo:              os.Getenv("WECHATPAY_SERIAL_NO"),
		WeChatPayPublicKeyID:           os.Getenv("WECHATPAY_PUBLIC_KEY_ID"),
		WeChatPayPublicKeyPEM:          os.Getenv("WECHATPAY_PUBLIC_KEY_PEM"),
		WeChatPayPrivateKeyPEM:         os.Getenv("WECHATPAY_PRIVATE_KEY_PEM"),
		WeChatPayAPIv3Key:              os.Getenv("WECHATPAY_APIV3_KEY"),
		WeChatPayMode:                  os.Getenv("WECHATPAY_MODE"),
		WeChatPaySubMchID:              os.Getenv("WECHATPAY_SUB_MCH_ID"),
		WeChatPaySubAppID:              os.Getenv("WECHATPAY_SUB_APP_ID"),
		WeChatLoginEnabled:             os.Getenv("WECHAT_LOGIN_ENABLED") == "true",
		WeChatLoginOpenBaseURL:         os.Getenv("WECHAT_LOGIN_OPEN_BASE_URL"),
		WeChatLoginAPIBaseURL:          os.Getenv("WECHAT_LOGIN_API_BASE_URL"),
		WeChatLoginAppID:               os.Getenv("WECHAT_LOGIN_APP_ID"),
		WeChatLoginAppSecret:           os.Getenv("WECHAT_LOGIN_APP_SECRET"),
		WeChatLoginRedirectURL:         os.Getenv("WECHAT_LOGIN_REDIRECT_URL"),
		WeChatLoginMockUser:            os.Getenv("WECHAT_LOGIN_MOCK_USER"),
		WeChatLoginMockOpenID:          os.Getenv("WECHAT_LOGIN_MOCK_OPEN_ID"),
		WeChatLoginMockUnionID:         os.Getenv("WECHAT_LOGIN_MOCK_UNION_ID"),
		WorkspaceBridgePath:            os.Getenv("WORKSPACE_BRIDGE_PATH"),
		WorkspaceBridgeHealthPath:      os.Getenv("WORKSPACE_BRIDGE_HEALTH_PATH"),
		WorkspaceBridgeHistoryPath:     os.Getenv("WORKSPACE_BRIDGE_HISTORY_PATH"),
		WorkspaceBridgeStreamPath:      os.Getenv("WORKSPACE_BRIDGE_STREAM_PATH"),
		WorkspaceBridgeReportPath:      os.Getenv("WORKSPACE_BRIDGE_REPORT_PATH"),
		WorkspaceBridgeHeaderName:      os.Getenv("WORKSPACE_BRIDGE_HEADER_NAME"),
		WorkspaceBridgeToken:           os.Getenv("WORKSPACE_BRIDGE_TOKEN"),
		WorkspaceBridgePublicBaseURL:   os.Getenv("WORKSPACE_BRIDGE_PUBLIC_BASE_URL"),
		WorkspaceBridgeTimeoutSecs:     envInt("WORKSPACE_BRIDGE_TIMEOUT_SECS", 0),
		WorkspaceBridgeRetryCount:      envInt("WORKSPACE_BRIDGE_RETRY_COUNT", 0),
		WorkspaceBridgeRetryBackoffMs:  envInt("WORKSPACE_BRIDGE_RETRY_BACKOFF_MS", 0),
		WorkspaceBridgeHistorySync:     envBool("WORKSPACE_BRIDGE_HISTORY_SYNC", true),
		ArtifactPreviewTimeoutSecs:     envInt("ARTIFACT_PREVIEW_TIMEOUT_SECS", 0),
		ArtifactPreviewHTMLMaxBytes:    envInt("ARTIFACT_PREVIEW_HTML_MAX_BYTES", 0),
		ArtifactPreviewAllowPrivateIP:  envBool("ARTIFACT_PREVIEW_ALLOW_PRIVATE_IP", false),
		ArtifactPreviewPublicBaseURL:   os.Getenv("ARTIFACT_PREVIEW_PUBLIC_BASE_URL"),
		ArtifactPreviewAllowedHosts:    os.Getenv("ARTIFACT_PREVIEW_ALLOWED_HOSTS"),
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid platform api configuration: %v", err)
	}

	externalDocsOutputDir := strings.TrimSpace(os.Getenv("EXTERNAL_DOCS_OUTPUT_DIR"))
	if externalDocsOutputDir != "" {
		if err := httpapi.ExportExternalDocs(externalDocsOutputDir); err != nil {
			log.Fatalf("export external docs: %v", err)
		}
		log.Printf("exported external docs to %s", externalDocsOutputDir)
	}

	router, err := httpapi.NewRouter(data, cfg)
	if err != nil {
		log.Fatalf("initialize router: %v", err)
	}

	addr := ":" + port
	log.Printf(
		"openclaw platform api listening on %s (mode=%s strictMode=%t stateBackend=%s runtimeProvider=%s bootstrap=%s autoMigrate=%t)",
		addr,
		serviceMode,
		strictMode,
		stateBackend,
		runtimeProvider,
		bootstrapStrategy,
		autoMigrate,
	)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func normalizeRuntimeProvider(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return "kubectl"
	}
	return normalized
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}

func envInt64(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
