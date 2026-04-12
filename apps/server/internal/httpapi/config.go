package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"openclaw/platformapi/internal/corestore"
)

type ExternalConfig struct {
	ServiceName string
	StrictMode  bool

	BuildVersion string
	BuildCommit  string
	BuildDate    string

	CoreStore corestore.CoreStore

	BootstrapStrategy string
	HTTPClient        *http.Client

	RuntimeProvider        string
	RuntimeKubectlBinary   string
	RuntimeKubeContext     string
	RuntimeNamespacePrefix string
	RuntimeWorkloadPrefix  string
	RuntimeImage           string
	RuntimeServiceType     string
	RuntimeAccessHost      string
	RuntimeAccessScheme    string
	RuntimePort            int

	KeycloakEnabled              bool
	KeycloakBaseURL              string
	KeycloakRealm                string
	KeycloakClientID             string
	KeycloakClientSecret         string
	KeycloakRedirectURL          string
	KeycloakPostLoginRedirectURL string
	KeycloakLogoutRedirectURL    string
	KeycloakMockUser             string
	KeycloakMockEmail            string
	KeycloakMockRole             string
	KeycloakSessionSecret        string
	KeycloakCookieName           string
	KeycloakCookieSecure         bool
	KeycloakFlowCookieName       string

	OpenSearchEnabled  bool
	OpenSearchURL      string
	OpenSearchIndex    string
	OpenSearchUsername string
	OpenSearchPassword string

	TraceSearchEnabled       bool
	TraceSearchURL           string
	TraceSearchIndex         string
	TraceSearchUsername      string
	TraceSearchPassword      string
	TraceSearchPublicBaseURL string

	ObjectStorageEndpoint          string
	ObjectStorageBucket            string
	ObjectStorageAccessKey         string
	ObjectStorageSecretKey         string
	ObjectStorageRegion            string
	ObjectStorageForcePathStyle    bool
	ArtifactArchiveMaxBytes        int64
	ArtifactArchiveAllowPrivateURL bool

	WeChatPayEnabled         bool
	WeChatPayBaseURL         string
	WeChatPayMchID           string
	WeChatPayAppID           string
	WeChatPayClientSecret    string
	WeChatPayNotifyURL       string
	WeChatPayRefundNotifyURL string
	WeChatPaySerialNo        string
	WeChatPayPublicKeyID     string
	WeChatPayPublicKeyPEM    string
	WeChatPayPrivateKeyPEM   string
	WeChatPayAPIv3Key        string
	WeChatPayMode            string
	WeChatPaySubMchID        string
	WeChatPaySubAppID        string

	WeChatLoginEnabled     bool
	WeChatLoginOpenBaseURL string
	WeChatLoginAPIBaseURL  string
	WeChatLoginAppID       string
	WeChatLoginAppSecret   string
	WeChatLoginRedirectURL string
	WeChatLoginMockUser    string
	WeChatLoginMockOpenID  string
	WeChatLoginMockUnionID string

	WorkspaceBridgePath           string
	WorkspaceBridgeHealthPath     string
	WorkspaceBridgeHistoryPath    string
	WorkspaceBridgeStreamPath     string
	WorkspaceBridgeReportPath     string
	WorkspaceBridgeHeaderName     string
	WorkspaceBridgeToken          string
	WorkspaceBridgePublicBaseURL  string
	WorkspaceBridgeTimeoutSecs    int
	WorkspaceBridgeRetryCount     int
	WorkspaceBridgeRetryBackoffMs int
	WorkspaceBridgeHistorySync    bool

	ArtifactPreviewTimeoutSecs    int
	ArtifactPreviewHTMLMaxBytes   int
	ArtifactPreviewAllowPrivateIP bool
	ArtifactPreviewPublicBaseURL  string
	ArtifactPreviewAllowedHosts   string
}

func (cfg ExternalConfig) Validate() error {
	if !cfg.StrictMode {
		return nil
	}

	problems := make([]string, 0)
	if !coreStoreConfigured(cfg.CoreStore) {
		problems = append(problems, "DATABASE_URL is required in strict mode")
	}

	runtimeProvider := strings.ToLower(strings.TrimSpace(cfg.RuntimeProvider))
	if runtimeProvider == "" || runtimeProvider == "mock" {
		problems = append(problems, "RUNTIME_PROVIDER must be a real runtime provider in strict mode")
	}

	if strings.TrimSpace(cfg.KeycloakMockUser) != "" || strings.TrimSpace(cfg.KeycloakMockEmail) != "" {
		problems = append(problems, "KEYCLOAK_MOCK_* must be empty in strict mode")
	}
	if strings.TrimSpace(cfg.WeChatLoginMockUser) != "" || strings.TrimSpace(cfg.WeChatLoginMockOpenID) != "" || strings.TrimSpace(cfg.WeChatLoginMockUnionID) != "" {
		problems = append(problems, "WECHAT_LOGIN_MOCK_* must be empty in strict mode")
	}

	if cfg.KeycloakEnabled {
		if !allStrictValuesSet(
			cfg.KeycloakBaseURL,
			cfg.KeycloakRealm,
			cfg.KeycloakClientID,
			cfg.KeycloakClientSecret,
			cfg.KeycloakRedirectURL,
			cfg.KeycloakPostLoginRedirectURL,
			cfg.KeycloakLogoutRedirectURL,
			cfg.KeycloakSessionSecret,
			cfg.KeycloakCookieName,
			cfg.KeycloakFlowCookieName,
		) {
			problems = append(problems, "KEYCLOAK_* configuration is incomplete for strict mode")
		}
	}

	if cfg.WeChatLoginEnabled {
		if !allStrictValuesSet(cfg.WeChatLoginAppID, cfg.WeChatLoginAppSecret, cfg.WeChatLoginRedirectURL) {
			problems = append(problems, "WECHAT_LOGIN_* configuration is incomplete for strict mode")
		}
	}

	if cfg.OpenSearchEnabled {
		if !allStrictValuesSet(cfg.OpenSearchURL, cfg.OpenSearchIndex, cfg.OpenSearchUsername, cfg.OpenSearchPassword) {
			problems = append(problems, "OPENSEARCH_* configuration is incomplete for strict mode")
		}
	}
	if cfg.TraceSearchEnabled {
		if !allStrictValuesSet(cfg.TraceSearchURL, cfg.TraceSearchIndex) {
			problems = append(problems, "TRACE_SEARCH_* configuration is incomplete")
		}
	}

	if strings.TrimSpace(cfg.ObjectStorageEndpoint) != "" || strings.TrimSpace(cfg.ObjectStorageBucket) != "" || strings.TrimSpace(cfg.ObjectStorageAccessKey) != "" || strings.TrimSpace(cfg.ObjectStorageSecretKey) != "" {
		if !allStrictValuesSet(cfg.ObjectStorageEndpoint, cfg.ObjectStorageBucket, cfg.ObjectStorageAccessKey, cfg.ObjectStorageSecretKey) {
			problems = append(problems, "OBJECT_STORAGE_* configuration is incomplete")
		}
	}

	if cfg.StrictMode {
		if !allStrictValuesSet(cfg.ObjectStorageEndpoint, cfg.ObjectStorageBucket, cfg.ObjectStorageAccessKey, cfg.ObjectStorageSecretKey) {
			problems = append(problems, "OBJECT_STORAGE_* configuration is required in strict mode")
		}
	}

	if cfg.WeChatPayEnabled {
		if !allStrictValuesSet(
			cfg.WeChatPayMchID,
			cfg.WeChatPayAppID,
			cfg.WeChatPayPrivateKeyPEM,
			cfg.WeChatPayNotifyURL,
			cfg.WeChatPayRefundNotifyURL,
			cfg.WeChatPayAPIv3Key,
		) {
			problems = append(problems, "WECHATPAY_* configuration is incomplete for strict mode")
		}
	}

	if value := strings.TrimSpace(cfg.ArtifactPreviewPublicBaseURL); value != "" {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			problems = append(problems, "ARTIFACT_PREVIEW_PUBLIC_BASE_URL must be a valid absolute URL")
		}
	}
	if value := strings.TrimSpace(cfg.TraceSearchPublicBaseURL); value != "" {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			problems = append(problems, "TRACE_SEARCH_PUBLIC_BASE_URL must be a valid absolute URL")
		}
	}

	if cfg.StrictMode {
		if strings.TrimSpace(cfg.ArtifactPreviewPublicBaseURL) == "" {
			problems = append(problems, "ARTIFACT_PREVIEW_PUBLIC_BASE_URL is required in strict mode")
		}
		if cfg.ArtifactPreviewAllowPrivateIP {
			problems = append(problems, "ARTIFACT_PREVIEW_ALLOW_PRIVATE_IP must be false in strict mode")
		}
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func coreStoreConfigured(store corestore.CoreStore) bool {
	if store == nil {
		return false
	}

	value := reflect.ValueOf(store)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}

func normalizeCoreStore(store corestore.CoreStore) corestore.CoreStore {
	if !coreStoreConfigured(store) {
		return nil
	}
	return store
}

func allStrictValuesSet(values ...string) bool {
	for _, value := range values {
		if !strictValueSet(value) {
			return false
		}
	}
	return true
}

func strictValueSet(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	normalized := strings.ToLower(trimmed)
	for _, token := range []string{"replace-me", "change-me", "set-in-cluster", "example.com"} {
		if strings.Contains(normalized, token) {
			return false
		}
	}
	return true
}
