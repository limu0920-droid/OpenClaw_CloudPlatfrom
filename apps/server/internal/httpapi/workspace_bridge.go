package httpapi

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
)

const (
	workspaceBridgeProtocolVersion       = "openclaw-lobster-bridge/v2"
	workspaceBridgeProtocolHeader        = "X-OpenClaw-Protocol-Version"
	workspaceBridgeRequestIDHeader       = "X-OpenClaw-Request-Id"
	defaultWorkspaceBridgeRetryCount     = 2
	defaultWorkspaceBridgeRetryBackoff   = 250 * time.Millisecond
	defaultWorkspaceBridgeHistoryPage    = 200
	defaultWorkspaceBridgeHistoryMaxPage = 5
)

const (
	workspaceMessageOriginPlatform        = "platform"
	workspaceMessageOriginBridgeResponse  = "bridge_response"
	workspaceMessageOriginBridgeHistory   = "bridge_history"
	workspaceMessageOriginBridgeReport    = "bridge_report"
	workspaceArtifactOriginManual         = "manual"
	workspaceArtifactOriginBridgeResponse = "bridge_response"
	workspaceArtifactOriginBridgeHistory  = "bridge_history"
	workspaceArtifactOriginBridgeReport   = "bridge_report"
)

type workspaceBridgeDispatchResult struct {
	OK              bool                      `json:"ok"`
	Target          string                    `json:"target"`
	RequestID       string                    `json:"requestId,omitempty"`
	TraceID         string                    `json:"traceId,omitempty"`
	ProtocolVersion string                    `json:"protocolVersion,omitempty"`
	Status          int                       `json:"status,omitempty"`
	Code            string                    `json:"code,omitempty"`
	Error           string                    `json:"error,omitempty"`
	Retryable       bool                      `json:"retryable,omitempty"`
	Attempt         int                       `json:"attempt,omitempty"`
	StreamURL       string                    `json:"streamUrl,omitempty"`
	Message         string                    `json:"message,omitempty"`
	Assistant       *workspaceBridgeMessage   `json:"assistant,omitempty"`
	Messages        []workspaceBridgeMessage  `json:"messages,omitempty"`
	Artifacts       []workspaceBridgeArtifact `json:"artifacts,omitempty"`
}

type workspaceBridgeSyncResult struct {
	OK              bool   `json:"ok"`
	Skipped         bool   `json:"skipped,omitempty"`
	Target          string `json:"target,omitempty"`
	RequestID       string `json:"requestId,omitempty"`
	ProtocolVersion string `json:"protocolVersion,omitempty"`
	Code            string `json:"code,omitempty"`
	Error           string `json:"error,omitempty"`
	MessagesSynced  int    `json:"messagesSynced,omitempty"`
	ArtifactsSynced int    `json:"artifactsSynced,omitempty"`
	LastSyncedAt    string `json:"lastSyncedAt,omitempty"`
}

type workspaceBridgeCapabilities struct {
	MessageDispatch bool `json:"messageDispatch"`
	HistoryPull     bool `json:"historyPull"`
	StreamSSE       bool `json:"streamSse"`
	ReportCallback  bool `json:"reportCallback"`
}

type workspaceBridgeHealthResult struct {
	OK              bool                        `json:"ok"`
	Target          string                      `json:"target"`
	RequestID       string                      `json:"requestId,omitempty"`
	ProtocolVersion string                      `json:"protocolVersion,omitempty"`
	Service         string                      `json:"service,omitempty"`
	Status          int                         `json:"status,omitempty"`
	Error           string                      `json:"error,omitempty"`
	Message         string                      `json:"message,omitempty"`
	Capabilities    workspaceBridgeCapabilities `json:"capabilities"`
}

type workspaceBridgeRequest struct {
	ProtocolVersion string                       `json:"protocolVersion"`
	RequestID       string                       `json:"requestId"`
	MessageID       int                          `json:"messageId"`
	SessionID       int                          `json:"sessionId"`
	SessionNo       string                       `json:"sessionNo"`
	TenantID        int                          `json:"tenantId"`
	InstanceID      int                          `json:"instanceId"`
	InstanceCode    string                       `json:"instanceCode"`
	InstanceName    string                       `json:"instanceName"`
	Role            string                       `json:"role"`
	Content         string                       `json:"content"`
	CreatedAt       string                       `json:"createdAt"`
	Delivery        *workspaceBridgeDelivery     `json:"delivery,omitempty"`
	Callback        *workspaceBridgeCallbackInfo `json:"callback,omitempty"`
}

type workspaceBridgeDelivery struct {
	TimeoutSeconds int  `json:"timeoutSeconds,omitempty"`
	RetryCount     int  `json:"retryCount,omitempty"`
	HistorySync    bool `json:"historySync,omitempty"`
	Stream         bool `json:"stream,omitempty"`
}

type workspaceBridgeCallbackInfo struct {
	ReportURL string `json:"reportUrl,omitempty"`
}

type workspaceBridgeMessage struct {
	ID           string `json:"id,omitempty"`
	Role         string `json:"role"`
	Status       string `json:"status,omitempty"`
	Content      string `json:"content"`
	CreatedAt    string `json:"createdAt,omitempty"`
	TraceID      string `json:"traceId,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type workspaceBridgeArtifact struct {
	ID         string `json:"id,omitempty"`
	MessageID  string `json:"messageId,omitempty"`
	Title      string `json:"title"`
	Kind       string `json:"kind"`
	SourceURL  string `json:"sourceUrl"`
	PreviewURL string `json:"previewUrl,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

type workspaceBridgeStream struct {
	Mode string `json:"mode,omitempty"`
	URL  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type workspaceBridgeHistoryPayload struct {
	Messages  []workspaceBridgeMessage  `json:"messages,omitempty"`
	Artifacts []workspaceBridgeArtifact `json:"artifacts,omitempty"`
	Cursor    string                    `json:"cursor,omitempty"`
	HasMore   bool                      `json:"hasMore,omitempty"`
	SyncedAt  string                    `json:"syncedAt,omitempty"`
	RequestID string                    `json:"requestId,omitempty"`
	TraceID   string                    `json:"traceId,omitempty"`
	Code      string                    `json:"code,omitempty"`
	Message   string                    `json:"message,omitempty"`
	Retryable bool                      `json:"retryable,omitempty"`
}

type workspaceBridgeEnvelope struct {
	ProtocolVersion string                         `json:"protocolVersion,omitempty"`
	RequestID       string                         `json:"requestId,omitempty"`
	TraceID         string                         `json:"traceId,omitempty"`
	Accepted        *bool                          `json:"accepted,omitempty"`
	Code            string                         `json:"code,omitempty"`
	Message         string                         `json:"message,omitempty"`
	Assistant       *workspaceBridgeMessage        `json:"assistant,omitempty"`
	Retryable       bool                           `json:"retryable,omitempty"`
	Service         string                         `json:"service,omitempty"`
	Capabilities    workspaceBridgeCapabilities    `json:"capabilities"`
	Messages        []workspaceBridgeMessage       `json:"messages,omitempty"`
	Artifacts       []workspaceBridgeArtifact      `json:"artifacts,omitempty"`
	Stream          *workspaceBridgeStream         `json:"stream,omitempty"`
	Cursor          string                         `json:"cursor,omitempty"`
	HasMore         bool                           `json:"hasMore,omitempty"`
	SyncedAt        string                         `json:"syncedAt,omitempty"`
	History         *workspaceBridgeHistoryPayload `json:"history,omitempty"`
}

type workspaceBridgeReportRequest struct {
	ProtocolVersion string                    `json:"protocolVersion"`
	RequestID       string                    `json:"requestId"`
	SessionNo       string                    `json:"sessionNo"`
	TraceID         string                    `json:"traceId,omitempty"`
	SyncedAt        string                    `json:"syncedAt,omitempty"`
	Messages        []workspaceBridgeMessage  `json:"messages,omitempty"`
	Artifacts       []workspaceBridgeArtifact `json:"artifacts,omitempty"`
	Error           *workspaceBridgeErrorBody `json:"error,omitempty"`
}

type workspaceBridgeErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable,omitempty"`
}

func (r *Router) workspaceBridgePath() string {
	if value := strings.TrimSpace(r.config.WorkspaceBridgePath); value != "" {
		return value
	}
	return "/api/v1/platform/workspace/messages"
}

func (r *Router) workspaceBridgeHealthPath() string {
	if value := strings.TrimSpace(r.config.WorkspaceBridgeHealthPath); value != "" {
		return value
	}
	return "/api/v1/platform/workspace/health"
}

func (r *Router) workspaceBridgeHistoryPath() string {
	if value := strings.TrimSpace(r.config.WorkspaceBridgeHistoryPath); value != "" {
		return value
	}
	return "/api/v1/platform/workspace/sessions/{sessionNo}/history"
}

func (r *Router) workspaceBridgeStreamPath() string {
	if value := strings.TrimSpace(r.config.WorkspaceBridgeStreamPath); value != "" {
		return value
	}
	return "/api/v1/platform/workspace/sessions/{sessionNo}/stream"
}

func (r *Router) workspaceBridgeReportPath() string {
	if value := strings.TrimSpace(r.config.WorkspaceBridgeReportPath); value != "" {
		return value
	}
	return "/api/v1/platform/workspace/report"
}

func (r *Router) workspaceBridgeHeaderName() string {
	if value := strings.TrimSpace(r.config.WorkspaceBridgeHeaderName); value != "" {
		return value
	}
	return "X-Platform-Bridge-Token"
}

func (r *Router) workspaceBridgeHTTPClient() *http.Client {
	if r.config.HTTPClient != nil {
		return r.config.HTTPClient
	}

	timeout := 10 * time.Second
	if r.config.WorkspaceBridgeTimeoutSecs > 0 {
		timeout = time.Duration(r.config.WorkspaceBridgeTimeoutSecs) * time.Second
	}
	return &http.Client{Timeout: timeout}
}

func (r *Router) workspaceBridgeRetryCount() int {
	if r.config.WorkspaceBridgeRetryCount > 0 {
		return r.config.WorkspaceBridgeRetryCount
	}
	return defaultWorkspaceBridgeRetryCount
}

func (r *Router) workspaceBridgeRetryBackoff() time.Duration {
	if r.config.WorkspaceBridgeRetryBackoffMs > 0 {
		return time.Duration(r.config.WorkspaceBridgeRetryBackoffMs) * time.Millisecond
	}
	return defaultWorkspaceBridgeRetryBackoff
}

func (r *Router) workspaceBridgeHistorySyncEnabled() bool {
	return r.config.WorkspaceBridgeHistorySync
}

func (r *Router) workspaceBridgePublicReportURL() string {
	baseURL := strings.TrimRight(strings.TrimSpace(r.config.WorkspaceBridgePublicBaseURL), "/")
	if baseURL == "" {
		return ""
	}

	path := strings.TrimSpace(r.workspaceBridgeReportPath())
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return baseURL + path
}

func (r *Router) workspaceBridgeURL(workspaceURL string) (string, error) {
	return r.workspaceBridgeURLFromTemplate(workspaceURL, r.workspaceBridgePath(), nil, nil)
}

func (r *Router) workspaceBridgeHealthURL(workspaceURL string) (string, error) {
	return r.workspaceBridgeURLFromTemplate(workspaceURL, r.workspaceBridgeHealthPath(), nil, nil)
}

func (r *Router) workspaceBridgeHistoryURL(workspaceURL string, sessionNo string, query url.Values) (string, error) {
	return r.workspaceBridgeURLFromTemplate(workspaceURL, r.workspaceBridgeHistoryPath(), map[string]string{
		"{sessionNo}": url.PathEscape(sessionNo),
	}, query)
}

func (r *Router) workspaceBridgeStreamURL(workspaceURL string, sessionNo string, query url.Values) (string, error) {
	return r.workspaceBridgeURLFromTemplate(workspaceURL, r.workspaceBridgeStreamPath(), map[string]string{
		"{sessionNo}": url.PathEscape(sessionNo),
	}, query)
}

func (r *Router) workspaceBridgeURLFromTemplate(workspaceURL string, template string, replacements map[string]string, query url.Values) (string, error) {
	base, err := url.Parse(strings.TrimSpace(workspaceURL))
	if err != nil {
		return "", err
	}
	if base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("workspace url is invalid")
	}

	pathValue := strings.TrimSpace(template)
	for placeholder, value := range replacements {
		pathValue = strings.ReplaceAll(pathValue, placeholder, value)
	}

	if strings.HasPrefix(pathValue, "http://") || strings.HasPrefix(pathValue, "https://") {
		absolute, err := url.Parse(pathValue)
		if err != nil {
			return "", err
		}
		if len(query) > 0 {
			values := absolute.Query()
			for key, items := range query {
				for _, item := range items {
					values.Add(key, item)
				}
			}
			absolute.RawQuery = values.Encode()
		}
		return absolute.String(), nil
	}

	base.Path = pathValue
	base.RawPath = ""
	base.Fragment = ""
	if len(query) > 0 {
		base.RawQuery = query.Encode()
	} else {
		base.RawQuery = ""
	}
	return base.String(), nil
}

func workspaceBridgeRequestID(session models.WorkspaceSession, message models.WorkspaceMessage) string {
	return fmt.Sprintf("ws-%s-%d", session.SessionNo, message.ID)
}

func (r *Router) applyWorkspaceBridgeHeaders(req *http.Request, requestID string) {
	req.Header.Set(workspaceBridgeProtocolHeader, workspaceBridgeProtocolVersion)
	if requestID != "" {
		req.Header.Set(workspaceBridgeRequestIDHeader, requestID)
	}
	if token := strings.TrimSpace(r.config.WorkspaceBridgeToken); token != "" {
		req.Header.Set(r.workspaceBridgeHeaderName(), token)
	}
}

func (r *Router) authorizeWorkspaceBridgeCallback(req *http.Request) bool {
	token := strings.TrimSpace(r.config.WorkspaceBridgeToken)
	if token == "" {
		return true
	}
	headerValue := strings.TrimSpace(req.Header.Get(r.workspaceBridgeHeaderName()))
	return subtle.ConstantTimeCompare([]byte(token), []byte(headerValue)) == 1
}

func (r *Router) dispatchWorkspaceMessage(ctx context.Context, session models.WorkspaceSession, instance models.Instance, message models.WorkspaceMessage) workspaceBridgeDispatchResult {
	if strings.TrimSpace(session.WorkspaceURL) == "" {
		return workspaceBridgeDispatchResult{
			OK:    false,
			Error: "workspace url is empty",
		}
	}

	target, err := r.workspaceBridgeURL(session.WorkspaceURL)
	if err != nil {
		return workspaceBridgeDispatchResult{
			OK:    false,
			Error: err.Error(),
		}
	}

	requestID := workspaceBridgeRequestID(session, message)
	requestPayload := workspaceBridgeRequest{
		ProtocolVersion: workspaceBridgeProtocolVersion,
		RequestID:       requestID,
		MessageID:       message.ID,
		SessionID:       session.ID,
		SessionNo:       session.SessionNo,
		TenantID:        session.TenantID,
		InstanceID:      session.InstanceID,
		InstanceCode:    instance.Code,
		InstanceName:    instance.Name,
		Role:            message.Role,
		Content:         message.Content,
		CreatedAt:       message.CreatedAt,
		Delivery: &workspaceBridgeDelivery{
			TimeoutSeconds: r.config.WorkspaceBridgeTimeoutSecs,
			RetryCount:     r.workspaceBridgeRetryCount(),
			HistorySync:    r.workspaceBridgeHistorySyncEnabled(),
			Stream:         true,
		},
	}
	if reportURL := r.workspaceBridgePublicReportURL(); reportURL != "" {
		requestPayload.Callback = &workspaceBridgeCallbackInfo{ReportURL: reportURL}
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return workspaceBridgeDispatchResult{
			OK:        false,
			Target:    target,
			RequestID: requestID,
			Error:     err.Error(),
		}
	}

	maxAttempts := r.workspaceBridgeRetryCount() + 1
	lastResult := workspaceBridgeDispatchResult{
		Target:          target,
		RequestID:       requestID,
		ProtocolVersion: workspaceBridgeProtocolVersion,
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result := r.dispatchWorkspaceMessageOnce(ctx, session.WorkspaceURL, target, requestID, body)
		result.Attempt = attempt
		if result.OK || !result.Retryable || attempt == maxAttempts {
			return result
		}
		lastResult = result
		timer := time.NewTimer(time.Duration(attempt) * r.workspaceBridgeRetryBackoff())
		select {
		case <-ctx.Done():
			timer.Stop()
			lastResult.Error = ctx.Err().Error()
			lastResult.Retryable = false
			return lastResult
		case <-timer.C:
		}
	}
	return lastResult
}

func (r *Router) dispatchWorkspaceMessageOnce(ctx context.Context, workspaceURL string, target string, requestID string, body []byte) workspaceBridgeDispatchResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return workspaceBridgeDispatchResult{
			OK:        false,
			Target:    target,
			RequestID: requestID,
			Error:     err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/json")
	r.applyWorkspaceBridgeHeaders(req, requestID)

	resp, err := r.workspaceBridgeHTTPClient().Do(req)
	if err != nil {
		return workspaceBridgeDispatchResult{
			OK:        false,
			Target:    target,
			RequestID: requestID,
			Code:      "bridge_transport_error",
			Error:     err.Error(),
			Retryable: !errors.Is(err, context.Canceled),
		}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	envelope, rawText := parseWorkspaceBridgeEnvelope(raw)
	result := workspaceBridgeDispatchResult{
		Target:          target,
		RequestID:       defaultString(strings.TrimSpace(envelope.RequestID), requestID),
		TraceID:         strings.TrimSpace(envelope.TraceID),
		ProtocolVersion: defaultString(strings.TrimSpace(envelope.ProtocolVersion), workspaceBridgeProtocolVersion),
		Status:          resp.StatusCode,
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.OK = false
		result.Retryable = envelope.Retryable || workspaceBridgeRetryableStatus(resp.StatusCode)
		result.Code = defaultString(strings.TrimSpace(envelope.Code), workspaceBridgeCodeFromStatus(resp.StatusCode))
		result.Error = firstNonEmpty(strings.TrimSpace(envelope.Message), rawText, fmt.Sprintf("bridge returned %d", resp.StatusCode))
		return result
	}

	if envelope.Accepted != nil && !*envelope.Accepted {
		result.OK = false
		result.Retryable = envelope.Retryable
		result.Code = defaultString(strings.TrimSpace(envelope.Code), "bridge_rejected")
		result.Error = firstNonEmpty(strings.TrimSpace(envelope.Message), "bridge rejected request")
		return result
	}

	result.OK = true
	result.Code = defaultString(strings.TrimSpace(envelope.Code), "accepted")
	result.Message = strings.TrimSpace(envelope.Message)
	result.Messages = envelope.Messages
	if result.Assistant == nil && envelope.Assistant != nil {
		copy := *envelope.Assistant
		result.Assistant = &copy
		if len(result.Messages) == 0 {
			result.Messages = []workspaceBridgeMessage{copy}
		}
		if result.Message == "" && strings.TrimSpace(copy.Content) != "" {
			result.Message = strings.TrimSpace(copy.Content)
		}
	}
	result.Artifacts = envelope.Artifacts
	for _, item := range result.Messages {
		if item.Role == "assistant" || item.Role == "system" {
			copy := item
			result.Assistant = &copy
			if result.Message == "" {
				result.Message = strings.TrimSpace(item.Content)
			}
			if item.Role == "assistant" {
				break
			}
		}
	}
	if result.Message == "" {
		result.Message = rawText
	}
	if envelope.Stream != nil {
		result.StreamURL = resolveWorkspaceBridgeStreamURL(workspaceURL, *envelope.Stream)
	}
	return result
}

func (r *Router) checkWorkspaceBridge(ctx context.Context, workspaceURL string) workspaceBridgeHealthResult {
	if strings.TrimSpace(workspaceURL) == "" {
		return workspaceBridgeHealthResult{
			OK:    false,
			Error: "workspace url is empty",
		}
	}

	target, err := r.workspaceBridgeHealthURL(workspaceURL)
	if err != nil {
		return workspaceBridgeHealthResult{
			OK:     false,
			Target: "",
			Error:  err.Error(),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return workspaceBridgeHealthResult{
			OK:     false,
			Target: target,
			Error:  err.Error(),
		}
	}
	r.applyWorkspaceBridgeHeaders(req, "")

	resp, err := r.workspaceBridgeHTTPClient().Do(req)
	if err != nil {
		return workspaceBridgeHealthResult{
			OK:     false,
			Target: target,
			Error:  err.Error(),
		}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	envelope, rawText := parseWorkspaceBridgeEnvelope(raw)
	result := workspaceBridgeHealthResult{
		Target:          target,
		RequestID:       strings.TrimSpace(envelope.RequestID),
		ProtocolVersion: defaultString(strings.TrimSpace(envelope.ProtocolVersion), workspaceBridgeProtocolVersion),
		Service:         strings.TrimSpace(envelope.Service),
		Status:          resp.StatusCode,
		Message:         rawText,
		Capabilities:    envelope.Capabilities,
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		result.OK = false
		result.Error = firstNonEmpty(strings.TrimSpace(envelope.Message), rawText, fmt.Sprintf("bridge returned %d", resp.StatusCode))
		return result
	}
	result.OK = true
	if strings.TrimSpace(envelope.Message) != "" {
		result.Message = strings.TrimSpace(envelope.Message)
	}
	return result
}

func (r *Router) syncWorkspaceSessionFromBridge(ctx context.Context, session models.WorkspaceSession) workspaceBridgeSyncResult {
	if !r.workspaceBridgeHistorySyncEnabled() {
		return workspaceBridgeSyncResult{
			OK:      true,
			Skipped: true,
			Code:    "history_sync_disabled",
		}
	}
	if strings.TrimSpace(session.WorkspaceURL) == "" {
		return workspaceBridgeSyncResult{
			OK:      true,
			Skipped: true,
			Code:    "workspace_url_empty",
		}
	}

	r.mu.RLock()
	since := r.workspaceSessionSyncSinceLocked(session.ID)
	r.mu.RUnlock()

	result := workspaceBridgeSyncResult{
		OK:              true,
		ProtocolVersion: workspaceBridgeProtocolVersion,
	}
	cursor := ""
	for page := 0; page < defaultWorkspaceBridgeHistoryMaxPage; page++ {
		query := url.Values{}
		if page == 0 && since != "" {
			query.Set("since", since)
		}
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		query.Set("limit", fmt.Sprintf("%d", defaultWorkspaceBridgeHistoryPage))

		target, err := r.workspaceBridgeHistoryURL(session.WorkspaceURL, session.SessionNo, query)
		if err != nil {
			return workspaceBridgeSyncResult{
				OK:              false,
				ProtocolVersion: workspaceBridgeProtocolVersion,
				Code:            "history_url_invalid",
				Error:           err.Error(),
			}
		}
		result.Target = target

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return workspaceBridgeSyncResult{
				OK:              false,
				Target:          target,
				ProtocolVersion: workspaceBridgeProtocolVersion,
				Code:            "history_request_invalid",
				Error:           err.Error(),
			}
		}
		r.applyWorkspaceBridgeHeaders(req, "")

		resp, err := r.workspaceBridgeHTTPClient().Do(req)
		if err != nil {
			return workspaceBridgeSyncResult{
				OK:              false,
				Target:          target,
				ProtocolVersion: workspaceBridgeProtocolVersion,
				Code:            "history_transport_error",
				Error:           err.Error(),
			}
		}

		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		envelope, rawText := parseWorkspaceBridgeEnvelope(raw)
		result.RequestID = defaultString(strings.TrimSpace(envelope.RequestID), result.RequestID)
		result.ProtocolVersion = defaultString(strings.TrimSpace(envelope.ProtocolVersion), result.ProtocolVersion)
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
			result.OK = true
			result.Skipped = true
			result.Code = "history_not_supported"
			result.Error = firstNonEmpty(strings.TrimSpace(envelope.Message), rawText)
			return result
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			result.OK = false
			result.Code = defaultString(strings.TrimSpace(envelope.Code), workspaceBridgeCodeFromStatus(resp.StatusCode))
			result.Error = firstNonEmpty(strings.TrimSpace(envelope.Message), rawText, fmt.Sprintf("bridge returned %d", resp.StatusCode))
			return result
		}

		messages, artifacts, nextCursor, hasMore, syncedAt := workspaceBridgeHistoryData(envelope)
		originTraceID := strings.TrimSpace(envelope.TraceID)
		if envelope.History != nil && originTraceID == "" {
			originTraceID = strings.TrimSpace(envelope.History.TraceID)
		}
		addedMessages, addedArtifacts, err := r.persistWorkspaceBridgePayload(session, messages, artifacts, workspaceMessageOriginBridgeHistory, workspaceArtifactOriginBridgeHistory, originTraceID, syncedAt)
		if err != nil {
			return workspaceBridgeSyncResult{
				OK:              false,
				Target:          target,
				RequestID:       result.RequestID,
				ProtocolVersion: result.ProtocolVersion,
				Code:            "history_persist_failed",
				Error:           err.Error(),
			}
		}
		result.MessagesSynced += addedMessages
		result.ArtifactsSynced += addedArtifacts
		if syncedAt != "" {
			result.LastSyncedAt = syncedAt
		}
		if !hasMore || nextCursor == "" {
			break
		}
		cursor = nextCursor
	}
	return result
}

func (r *Router) proxyWorkspaceBridgeStream(ctx context.Context, workspaceURL string, sessionNo string, query url.Values, w http.ResponseWriter) (int, error) {
	target, err := r.workspaceBridgeStreamURL(workspaceURL, sessionNo, query)
	if err != nil {
		return http.StatusBadGateway, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return http.StatusBadGateway, err
	}
	r.applyWorkspaceBridgeHeaders(req, "")

	resp, err := r.workspaceBridgeHTTPClient().Do(req)
	if err != nil {
		return http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/event-stream; charset=utf-8"
	}
	w.Header().Set("Content-Type", contentType)
	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "" {
		w.Header().Set("Cache-Control", cacheControl)
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if connection := resp.Header.Get("Connection"); connection != "" {
		w.Header().Set("Connection", connection)
	}
	w.WriteHeader(resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		_, _ = w.Write(raw)
		return resp.StatusCode, nil
	}

	flusher, _ := w.(http.Flusher)
	buffer := make([]byte, 4*1024)
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
				return resp.StatusCode, writeErr
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if readErr == nil {
			continue
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		return resp.StatusCode, readErr
	}
	return resp.StatusCode, nil
}

func (r *Router) persistWorkspaceBridgePayload(session models.WorkspaceSession, messages []workspaceBridgeMessage, artifacts []workspaceBridgeArtifact, messageOrigin string, artifactOrigin string, traceID string, syncedAt string) (int, int, error) {
	now := nowRFC3339()
	normalizedSyncedAt := normalizeRFC3339(syncedAt, now)
	messagesAdded := 0
	artifactsAdded := 0
	changed := false

	r.mu.Lock()
	sessionIndex := r.findWorkspaceSessionIndex(session.ID)
	if sessionIndex < 0 {
		r.mu.Unlock()
		return 0, 0, fmt.Errorf("workspace session %d not found", session.ID)
	}

	localSession := r.data.WorkspaceSessions[sessionIndex]
	messageExternalMap := make(map[string]int)
	for _, item := range r.data.WorkspaceMessages {
		if item.SessionID == session.ID && strings.TrimSpace(item.ExternalID) != "" {
			messageExternalMap[strings.TrimSpace(item.ExternalID)] = item.ID
		}
	}

	for _, item := range messages {
		normalized, ok := normalizeWorkspaceBridgeMessage(localSession, item, messageOrigin, traceID)
		if !ok {
			continue
		}

		if existingIndex := r.findWorkspaceMessageByExternalIDLocked(session.ID, normalized.ExternalID); existingIndex >= 0 {
			if mergeWorkspaceMessage(&r.data.WorkspaceMessages[existingIndex], normalized) {
				changed = true
			}
			if normalized.ExternalID != "" {
				messageExternalMap[normalized.ExternalID] = r.data.WorkspaceMessages[existingIndex].ID
			}
			continue
		}
		if existingIndex := r.findWorkspaceMessageDuplicateLocked(session.ID, normalized.Role, normalized.Content, normalized.CreatedAt); existingIndex >= 0 {
			if mergeWorkspaceMessage(&r.data.WorkspaceMessages[existingIndex], normalized) {
				changed = true
			}
			if normalized.ExternalID != "" {
				messageExternalMap[normalized.ExternalID] = r.data.WorkspaceMessages[existingIndex].ID
			}
			continue
		}

		normalized.ID = r.nextWorkspaceMessageID()
		r.data.WorkspaceMessages = append(r.data.WorkspaceMessages, *normalized)
		if normalized.ExternalID != "" {
			messageExternalMap[normalized.ExternalID] = normalized.ID
		}
		messagesAdded++
		changed = true
	}

	for _, item := range artifacts {
		normalized, ok := normalizeWorkspaceBridgeArtifact(localSession, item, artifactOrigin, now)
		if !ok {
			continue
		}
		if normalized.MessageID == 0 && strings.TrimSpace(item.MessageID) != "" {
			normalized.MessageID = messageExternalMap[strings.TrimSpace(item.MessageID)]
		}

		if existingIndex := r.findWorkspaceArtifactByExternalIDLocked(session.ID, normalized.ExternalID); existingIndex >= 0 {
			if mergeWorkspaceArtifact(&r.data.WorkspaceArtifacts[existingIndex], normalized) {
				changed = true
			}
			continue
		}
		if existingIndex := r.findWorkspaceArtifactDuplicateLocked(session.ID, normalized.SourceURL, normalized.Title, normalized.CreatedAt); existingIndex >= 0 {
			if mergeWorkspaceArtifact(&r.data.WorkspaceArtifacts[existingIndex], normalized) {
				changed = true
			}
			continue
		}

		normalized.ID = r.nextWorkspaceArtifactID()
		r.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{*normalized}, r.data.WorkspaceArtifacts...)
		artifactsAdded++
		changed = true
	}

	if changed {
		r.data.WorkspaceSessions[sessionIndex].ProtocolVersion = workspaceBridgeProtocolVersion
		r.data.WorkspaceSessions[sessionIndex].LastSyncedAt = normalizedSyncedAt
		if artifactsAdded > 0 {
			r.data.WorkspaceSessions[sessionIndex].LastArtifactAt = now
		}
		r.data.WorkspaceSessions[sessionIndex].UpdatedAt = now
	}
	r.mu.Unlock()

	if !changed {
		return 0, 0, nil
	}
	if err := r.persistAllData(); err != nil {
		return 0, 0, err
	}
	return messagesAdded, artifactsAdded, nil
}

func (r *Router) workspaceSessionSyncSinceLocked(sessionID int) string {
	session, ok := r.findWorkspaceSession(sessionID)
	if !ok {
		return ""
	}
	values := []string{session.LastSyncedAt, session.LastArtifactAt, session.LastOpenedAt, session.UpdatedAt}
	for _, item := range r.data.WorkspaceMessages {
		if item.SessionID != sessionID {
			continue
		}
		values = append(values, item.CreatedAt, item.UpdatedAt, item.DeliveredAt)
	}
	for _, item := range r.data.WorkspaceArtifacts {
		if item.SessionID != sessionID {
			continue
		}
		values = append(values, item.CreatedAt, item.UpdatedAt)
	}
	return maxRFC3339(values...)
}

func (r *Router) findWorkspaceMessageByExternalIDLocked(sessionID int, externalID string) int {
	if strings.TrimSpace(externalID) == "" {
		return -1
	}
	for index, item := range r.data.WorkspaceMessages {
		if item.SessionID == sessionID && strings.TrimSpace(item.ExternalID) == strings.TrimSpace(externalID) {
			return index
		}
	}
	return -1
}

func (r *Router) findWorkspaceArtifactByExternalIDLocked(sessionID int, externalID string) int {
	if strings.TrimSpace(externalID) == "" {
		return -1
	}
	for index, item := range r.data.WorkspaceArtifacts {
		if item.SessionID == sessionID && strings.TrimSpace(item.ExternalID) == strings.TrimSpace(externalID) {
			return index
		}
	}
	return -1
}

func (r *Router) findWorkspaceMessageDuplicateLocked(sessionID int, role string, content string, createdAt string) int {
	normalizedRole := strings.TrimSpace(role)
	normalizedContent := strings.TrimSpace(content)
	normalizedCreatedAt := normalizeRFC3339(createdAt, createdAt)
	for index, item := range r.data.WorkspaceMessages {
		if item.SessionID != sessionID {
			continue
		}
		if strings.TrimSpace(item.Role) == normalizedRole && strings.TrimSpace(item.Content) == normalizedContent && normalizeRFC3339(item.CreatedAt, item.CreatedAt) == normalizedCreatedAt {
			return index
		}
	}
	return -1
}

func (r *Router) findWorkspaceArtifactDuplicateLocked(sessionID int, sourceURL string, title string, createdAt string) int {
	normalizedURL := strings.TrimSpace(sourceURL)
	normalizedTitle := strings.TrimSpace(title)
	normalizedCreatedAt := normalizeRFC3339(createdAt, createdAt)
	for index, item := range r.data.WorkspaceArtifacts {
		if item.SessionID != sessionID {
			continue
		}
		if strings.TrimSpace(item.SourceURL) == normalizedURL && strings.TrimSpace(item.Title) == normalizedTitle && normalizeRFC3339(item.CreatedAt, item.CreatedAt) == normalizedCreatedAt {
			return index
		}
	}
	return -1
}

func normalizeWorkspaceBridgeMessage(session models.WorkspaceSession, item workspaceBridgeMessage, origin string, traceID string) (*models.WorkspaceMessage, bool) {
	content := strings.TrimSpace(item.Content)
	errorCode := strings.TrimSpace(item.ErrorCode)
	errorMessage := strings.TrimSpace(item.ErrorMessage)
	if content == "" {
		content = firstNonEmpty(errorMessage, errorCode)
	}
	if content == "" {
		return nil, false
	}

	now := nowRFC3339()
	status := defaultString(strings.TrimSpace(item.Status), "delivered")
	deliveredAt := ""
	if status == "delivered" || status == "sent" {
		deliveredAt = normalizeRFC3339(item.CreatedAt, now)
	}
	if errorCode != "" {
		status = "failed"
		deliveredAt = ""
	}

	return &models.WorkspaceMessage{
		SessionID:       session.ID,
		TenantID:        session.TenantID,
		InstanceID:      session.InstanceID,
		Role:            defaultString(strings.TrimSpace(item.Role), "assistant"),
		Status:          status,
		ExternalID:      strings.TrimSpace(item.ID),
		Origin:          origin,
		TraceID:         firstNonEmpty(strings.TrimSpace(item.TraceID), traceID),
		ErrorCode:       errorCode,
		ErrorMessage:    errorMessage,
		DeliveryAttempt: 1,
		Content:         content,
		DeliveredAt:     deliveredAt,
		CreatedAt:       normalizeRFC3339(item.CreatedAt, now),
		UpdatedAt:       now,
	}, true
}

func normalizeWorkspaceBridgeArtifact(session models.WorkspaceSession, item workspaceBridgeArtifact, origin string, fallbackTime string) (*models.WorkspaceArtifact, bool) {
	sourceURL := strings.TrimSpace(item.SourceURL)
	if sourceURL == "" {
		return nil, false
	}
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = "龙虾产物"
	}
	previewURL := strings.TrimSpace(item.PreviewURL)
	if previewURL == "" {
		previewURL = sourceURL
	}
	return &models.WorkspaceArtifact{
		SessionID:  session.ID,
		TenantID:   session.TenantID,
		InstanceID: session.InstanceID,
		Title:      title,
		Kind:       defaultString(strings.TrimSpace(item.Kind), "unknown"),
		ExternalID: strings.TrimSpace(item.ID),
		Origin:     origin,
		SourceURL:  sourceURL,
		PreviewURL: previewURL,
		CreatedAt:  normalizeRFC3339(item.CreatedAt, fallbackTime),
		UpdatedAt:  fallbackTime,
	}, true
}

func mergeWorkspaceMessage(current *models.WorkspaceMessage, next *models.WorkspaceMessage) bool {
	if current == nil || next == nil {
		return false
	}
	changed := false
	if current.ExternalID == "" && next.ExternalID != "" {
		current.ExternalID = next.ExternalID
		changed = true
	}
	if current.Origin == "" && next.Origin != "" {
		current.Origin = next.Origin
		changed = true
	}
	if current.TraceID == "" && next.TraceID != "" {
		current.TraceID = next.TraceID
		changed = true
	}
	if current.ErrorCode == "" && next.ErrorCode != "" {
		current.ErrorCode = next.ErrorCode
		changed = true
	}
	if current.ErrorMessage == "" && next.ErrorMessage != "" {
		current.ErrorMessage = next.ErrorMessage
		changed = true
	}
	if current.Status != next.Status && next.Status != "" {
		current.Status = next.Status
		changed = true
	}
	if current.DeliveredAt == "" && next.DeliveredAt != "" {
		current.DeliveredAt = next.DeliveredAt
		changed = true
	}
	if current.DeliveryAttempt < next.DeliveryAttempt {
		current.DeliveryAttempt = next.DeliveryAttempt
		changed = true
	}
	if current.Content != next.Content && next.Content != "" {
		current.Content = next.Content
		changed = true
	}
	if changed {
		current.UpdatedAt = nowRFC3339()
	}
	return changed
}

func mergeWorkspaceArtifact(current *models.WorkspaceArtifact, next *models.WorkspaceArtifact) bool {
	if current == nil || next == nil {
		return false
	}
	changed := false
	if current.ExternalID == "" && next.ExternalID != "" {
		current.ExternalID = next.ExternalID
		changed = true
	}
	if current.Origin == "" && next.Origin != "" {
		current.Origin = next.Origin
		changed = true
	}
	if current.MessageID == 0 && next.MessageID > 0 {
		current.MessageID = next.MessageID
		changed = true
	}
	if current.Title != next.Title && next.Title != "" {
		current.Title = next.Title
		changed = true
	}
	if current.Kind != next.Kind && next.Kind != "" {
		current.Kind = next.Kind
		changed = true
	}
	if current.SourceURL != next.SourceURL && next.SourceURL != "" {
		current.SourceURL = next.SourceURL
		changed = true
	}
	if current.PreviewURL == "" && next.PreviewURL != "" {
		current.PreviewURL = next.PreviewURL
		changed = true
	}
	if changed {
		current.UpdatedAt = nowRFC3339()
	}
	return changed
}

func workspaceBridgeHistoryData(envelope workspaceBridgeEnvelope) ([]workspaceBridgeMessage, []workspaceBridgeArtifact, string, bool, string) {
	messages := envelope.Messages
	artifacts := envelope.Artifacts
	cursor := envelope.Cursor
	hasMore := envelope.HasMore
	syncedAt := envelope.SyncedAt
	if envelope.History != nil {
		if len(messages) == 0 {
			messages = envelope.History.Messages
		}
		if len(artifacts) == 0 {
			artifacts = envelope.History.Artifacts
		}
		if cursor == "" {
			cursor = envelope.History.Cursor
		}
		if !hasMore {
			hasMore = envelope.History.HasMore
		}
		if syncedAt == "" {
			syncedAt = envelope.History.SyncedAt
		}
	}
	return messages, artifacts, cursor, hasMore, normalizeRFC3339(syncedAt, nowRFC3339())
}

func parseWorkspaceBridgeEnvelope(raw []byte) (workspaceBridgeEnvelope, string) {
	envelope := workspaceBridgeEnvelope{}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return envelope, ""
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return workspaceBridgeEnvelope{}, trimmed
	}
	if envelope.History != nil {
		if len(envelope.Messages) == 0 {
			envelope.Messages = envelope.History.Messages
		}
		if len(envelope.Artifacts) == 0 {
			envelope.Artifacts = envelope.History.Artifacts
		}
		if envelope.Cursor == "" {
			envelope.Cursor = envelope.History.Cursor
		}
		if !envelope.HasMore {
			envelope.HasMore = envelope.History.HasMore
		}
		if envelope.SyncedAt == "" {
			envelope.SyncedAt = envelope.History.SyncedAt
		}
		if envelope.RequestID == "" {
			envelope.RequestID = envelope.History.RequestID
		}
		if envelope.TraceID == "" {
			envelope.TraceID = envelope.History.TraceID
		}
		if envelope.Code == "" {
			envelope.Code = envelope.History.Code
		}
		if envelope.Message == "" {
			envelope.Message = envelope.History.Message
		}
		if !envelope.Retryable {
			envelope.Retryable = envelope.History.Retryable
		}
	}
	return envelope, trimmed
}

func resolveWorkspaceBridgeStreamURL(workspaceURL string, stream workspaceBridgeStream) string {
	if strings.TrimSpace(stream.URL) != "" {
		return strings.TrimSpace(stream.URL)
	}
	if strings.TrimSpace(stream.Path) == "" {
		return ""
	}
	base, err := url.Parse(strings.TrimSpace(workspaceURL))
	if err != nil {
		return ""
	}
	if base.Scheme == "" || base.Host == "" {
		return ""
	}
	base.Path = strings.TrimSpace(stream.Path)
	base.RawPath = ""
	base.RawQuery = ""
	base.Fragment = ""
	return base.String()
}

func workspaceBridgeRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func workspaceBridgeCodeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bridge_bad_request"
	case http.StatusUnauthorized:
		return "bridge_unauthorized"
	case http.StatusForbidden:
		return "bridge_forbidden"
	case http.StatusNotFound:
		return "bridge_not_found"
	case http.StatusConflict:
		return "bridge_conflict"
	case http.StatusUnprocessableEntity:
		return "bridge_unprocessable_entity"
	case http.StatusRequestTimeout:
		return "bridge_timeout"
	case http.StatusTooManyRequests:
		return "bridge_rate_limited"
	case http.StatusBadGateway:
		return "bridge_bad_gateway"
	case http.StatusServiceUnavailable:
		return "bridge_unavailable"
	case http.StatusGatewayTimeout:
		return "bridge_gateway_timeout"
	default:
		return "bridge_error"
	}
}
