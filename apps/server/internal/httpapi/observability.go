package httpapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

type observabilityRuntimeLog struct {
	ID            string `json:"id"`
	Timestamp     string `json:"timestamp"`
	Level         string `json:"level"`
	Source        string `json:"source"`
	Message       string `json:"message"`
	SessionID     int    `json:"sessionId,omitempty"`
	MessageID     int    `json:"messageId,omitempty"`
	TraceID       string `json:"traceId,omitempty"`
	TraceStart    string `json:"traceStart,omitempty"`
	TraceEnd      string `json:"traceEnd,omitempty"`
	RootCause     string `json:"rootCause,omitempty"`
	TraceExplorer string `json:"traceExplorerUrl,omitempty"`
	InstancePath  string `json:"instancePath,omitempty"`
	WorkspacePath string `json:"workspacePath,omitempty"`
}

type observabilityTraceSummary struct {
	TraceID       string `json:"traceId"`
	SessionID     int    `json:"sessionId,omitempty"`
	SessionNo     string `json:"sessionNo,omitempty"`
	StartAt       string `json:"startAt,omitempty"`
	LatestAt      string `json:"latestAt"`
	EndAt         string `json:"endAt,omitempty"`
	AnchorMessage int    `json:"anchorMessageId,omitempty"`
	Status        string `json:"status"`
	Preview       string `json:"preview,omitempty"`
	RootCause     string `json:"rootCause,omitempty"`
	TraceExplorer string `json:"traceExplorerUrl,omitempty"`
	MessageCount  int    `json:"messageCount"`
	EventCount    int    `json:"eventCount"`
	ArtifactCount int    `json:"artifactCount"`
	ToolCount     int    `json:"toolCount"`
	SpanCount     int    `json:"spanCount"`
}

type observabilityBridgeSummary struct {
	TraceCount       int                         `json:"traceCount"`
	EventCount       int                         `json:"eventCount"`
	FailedTraceCount int                         `json:"failedTraceCount"`
	ArtifactCount    int                         `json:"artifactCount"`
	LastEventAt      string                      `json:"lastEventAt,omitempty"`
	TraceBackend     string                      `json:"traceBackend"`
	TraceEnabled     bool                        `json:"traceEnabled"`
	RecentTraces     []observabilityTraceSummary `json:"recentTraces"`
}

type observabilityAlertRecord struct {
	ID               int    `json:"id"`
	InstanceID       int    `json:"instanceId"`
	Severity         string `json:"severity"`
	Status           string `json:"status"`
	MetricKey        string `json:"metricKey"`
	Summary          string `json:"summary"`
	TriggeredAt      string `json:"triggeredAt"`
	SessionID        int    `json:"sessionId,omitempty"`
	MessageID        int    `json:"messageId,omitempty"`
	TraceID          string `json:"traceId,omitempty"`
	TraceStart       string `json:"traceStart,omitempty"`
	TraceEnd         string `json:"traceEnd,omitempty"`
	RootCause        string `json:"rootCause,omitempty"`
	DetailPath       string `json:"detailPath,omitempty"`
	WorkspacePath    string `json:"workspacePath,omitempty"`
	TraceExplorerURL string `json:"traceExplorerUrl,omitempty"`
}

func (r *Router) buildInstanceResourceTrend(runtimeState *models.InstanceRuntime, metrics *runtimeadapter.WorkloadMetrics, alerts []models.Alert) []map[string]any {
	baseCPU := 0
	baseMemory := 0
	baseRequests := 0
	if runtimeState != nil {
		baseCPU = runtimeState.CPUUsagePercent
		baseMemory = runtimeState.MemoryUsagePercent
		baseRequests = maxInt(runtimeState.APIRequests24h/24, 120)
	}
	if metrics != nil {
		if metrics.RequestsPerMinute > 0 {
			baseRequests = maxInt(metrics.RequestsPerMinute*12, baseRequests)
		}
	}

	alertPressure := 0
	for _, item := range alerts {
		switch item.Severity {
		case "critical":
			alertPressure += 8
		case "warning":
			alertPressure += 4
		}
	}

	now := time.Now().UTC()
	offsets := []struct {
		hourDelta int
		cpuDelta  int
		memDelta  int
		reqDelta  int
	}{
		{-5, -18, -12, -320},
		{-4, -10, -6, -180},
		{-3, -4, -2, -60},
		{-2, 4, 5, 120},
		{-1, 7, 9, 260},
		{0, 0, 0, 0},
	}

	items := make([]map[string]any, 0, len(offsets))
	for _, item := range offsets {
		label := now.Add(time.Duration(item.hourDelta) * time.Hour).Format("15:04")
		cpu := clampInt(baseCPU+item.cpuDelta+alertPressure, 0, 100)
		memory := clampInt(baseMemory+item.memDelta+alertPressure/2, 0, 100)
		requests := maxInt(baseRequests+item.reqDelta+alertPressure*12, 0)
		items = append(items, map[string]any{
			"label":    label,
			"cpu":      cpu,
			"memory":   memory,
			"requests": requests,
		})
	}
	return items
}

func (r *Router) buildInstanceRuntimeLogsLocked(instanceID int, scope string) []observabilityRuntimeLog {
	items := make([]observabilityRuntimeLog, 0)
	traceIndex := r.buildTraceSummaryIndexLocked(instanceID)

	for _, alert := range r.filterAlertsByInstance(instanceID) {
		record := r.buildObservabilityAlertLocked(alert, scope)
		items = append(items, observabilityRuntimeLog{
			ID:            fmt.Sprintf("alert-%d", alert.ID),
			Timestamp:     alert.TriggeredAt,
			Level:         severityToLogLevel(alert.Severity),
			Source:        "alert",
			Message:       alert.Summary,
			SessionID:     record.SessionID,
			MessageID:     record.MessageID,
			TraceID:       record.TraceID,
			TraceStart:    record.TraceStart,
			TraceEnd:      record.TraceEnd,
			RootCause:     record.RootCause,
			TraceExplorer: record.TraceExplorerURL,
			InstancePath:  buildScopedInstancePath(scope, instanceID),
			WorkspacePath: record.WorkspacePath,
		})
	}

	for _, audit := range r.filterAuditsByInstance(instanceID) {
		sessionID, _ := strconv.Atoi(strings.TrimSpace(audit.Metadata["sessionId"]))
		messageID, _ := strconv.Atoi(strings.TrimSpace(audit.Metadata["messageId"]))
		traceID := strings.TrimSpace(audit.Metadata["traceId"])
		items = append(items, observabilityRuntimeLog{
			ID:            fmt.Sprintf("audit-%d", audit.ID),
			Timestamp:     audit.CreatedAt,
			Level:         auditResultToLogLevel(audit.Result),
			Source:        "audit",
			Message:       firstNonEmpty(audit.Metadata["summary"], audit.Action),
			SessionID:     sessionID,
			MessageID:     messageID,
			TraceID:       traceID,
			TraceStart:    traceStartAt(traceIndex, traceID),
			TraceEnd:      traceEndAt(traceIndex, traceID),
			RootCause:     traceRootCause(traceIndex, traceID),
			TraceExplorer: traceExplorerURL(traceIndex, traceID),
			InstancePath:  buildScopedInstancePath(scope, instanceID),
			WorkspacePath: buildScopedWorkspacePathWithWindow(scope, instanceID, sessionID, messageID, traceID, traceStartAt(traceIndex, traceID), traceEndAt(traceIndex, traceID)),
		})
	}

	for _, event := range r.filterWorkspaceMessageEventsByInstance(instanceID) {
		detail := summarizeWorkspaceEventPayload(event.PayloadJSON)
		traceID := firstNonEmpty(event.TraceID, extractTraceIDFromEventPayload(event.PayloadJSON))
		items = append(items, observabilityRuntimeLog{
			ID:            fmt.Sprintf("workspace-event-%d", event.ID),
			Timestamp:     event.CreatedAt,
			Level:         workspaceEventToLogLevel(event),
			Source:        firstNonEmpty(event.Origin, "workspace"),
			Message:       firstNonEmpty(detail, event.EventType),
			SessionID:     event.SessionID,
			MessageID:     event.MessageID,
			TraceID:       traceID,
			TraceStart:    traceStartAt(traceIndex, traceID),
			TraceEnd:      traceEndAt(traceIndex, traceID),
			RootCause:     traceRootCause(traceIndex, traceID),
			TraceExplorer: traceExplorerURL(traceIndex, traceID),
			InstancePath:  buildScopedInstancePath(scope, instanceID),
			WorkspacePath: buildScopedWorkspacePathWithWindow(scope, instanceID, event.SessionID, event.MessageID, traceID, traceStartAt(traceIndex, traceID), traceEndAt(traceIndex, traceID)),
		})
	}

	for _, record := range r.data.DiagnosticCommandRecords {
		if record.InstanceID != instanceID {
			continue
		}
		items = append(items, observabilityRuntimeLog{
			ID:            fmt.Sprintf("diagnostic-%d", record.ID),
			Timestamp:     record.ExecutedAt,
			Level:         diagnosticStatusToLogLevel(record.Status),
			Source:        "diagnostic",
			Message:       fmt.Sprintf("%s [%s]", record.CommandText, record.Status),
			SessionID:     record.SessionID,
			InstancePath:  buildScopedInstancePath(scope, instanceID),
			WorkspacePath: buildScopedWorkspacePath(scope, instanceID, record.SessionID, 0, ""),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp > items[j].Timestamp
	})
	if len(items) > 18 {
		return items[:18]
	}
	return items
}

func (r *Router) buildInstanceBridgeSummaryLocked(instanceID int) observabilityBridgeSummary {
	allTraces := r.buildInstanceTraceSummariesLocked(instanceID, 0)
	summary := observabilityBridgeSummary{
		RecentTraces: allTraces,
		TraceBackend: r.traceSearchProvider(),
		TraceEnabled: r.traceSearchEnabled(),
	}
	for _, item := range allTraces {
		summary.EventCount += item.EventCount
		summary.ArtifactCount += item.ArtifactCount
		if item.Status == "failed" {
			summary.FailedTraceCount++
		}
		summary.LastEventAt = maxRFC3339(summary.LastEventAt, item.LatestAt)
	}
	summary.TraceCount = len(allTraces)
	if len(summary.RecentTraces) > 6 {
		summary.RecentTraces = summary.RecentTraces[:6]
	}
	return summary
}

func (r *Router) buildInstanceTraceSummariesLocked(instanceID int, limit int) []observabilityTraceSummary {
	aggregated := make(map[string]*observabilityTraceSummary)
	sessionNoByID := make(map[int]string)
	for _, session := range r.data.WorkspaceSessions {
		if session.InstanceID == instanceID {
			sessionNoByID[session.ID] = session.SessionNo
		}
	}

	for _, message := range r.filterWorkspaceMessagesByInstance(instanceID) {
		traceID := strings.TrimSpace(message.TraceID)
		if traceID == "" {
			continue
		}
		item := ensureTraceSummary(aggregated, traceID)
		item.MessageCount++
		if item.SessionID == 0 {
			item.SessionID = message.SessionID
			item.SessionNo = sessionNoByID[message.SessionID]
		}
		item.StartAt = minRFC3339(item.StartAt, message.CreatedAt)
		item.LatestAt = maxRFC3339(item.LatestAt, message.CreatedAt)
		item.EndAt = maxRFC3339(item.EndAt, firstNonEmpty(message.DeliveredAt, message.UpdatedAt, message.CreatedAt))
		if message.ID > 0 && (item.AnchorMessage == 0 || item.EndAt == firstNonEmpty(message.DeliveredAt, message.UpdatedAt, message.CreatedAt)) {
			item.AnchorMessage = message.ID
		}
		if item.Preview == "" {
			item.Preview = trimPreviewText(message.Content, 84)
		}
		if item.RootCause == "" && (strings.EqualFold(message.Status, "failed") || strings.TrimSpace(message.ErrorMessage) != "" || strings.TrimSpace(message.ErrorCode) != "") {
			item.RootCause = firstNonEmpty(strings.TrimSpace(message.ErrorMessage), strings.TrimSpace(message.ErrorCode), trimPreviewText(message.Content, 84))
		}
		if message.Status == "failed" {
			item.Status = "failed"
		} else if message.Status == "streaming" && item.Status != "failed" {
			item.Status = "streaming"
		}
	}

	for _, event := range r.filterWorkspaceMessageEventsByInstance(instanceID) {
		traceID := strings.TrimSpace(event.TraceID)
		if traceID == "" {
			if payloadTraceID := extractTraceIDFromEventPayload(event.PayloadJSON); payloadTraceID != "" {
				traceID = payloadTraceID
			}
		}
		if traceID == "" {
			continue
		}
		item := ensureTraceSummary(aggregated, traceID)
		item.EventCount++
		if item.SessionID == 0 {
			item.SessionID = event.SessionID
			item.SessionNo = sessionNoByID[event.SessionID]
		}
		item.StartAt = minRFC3339(item.StartAt, event.CreatedAt)
		item.LatestAt = maxRFC3339(item.LatestAt, event.CreatedAt)
		item.EndAt = maxRFC3339(item.EndAt, event.CreatedAt)
		if event.MessageID > 0 && (item.AnchorMessage == 0 || item.EndAt == event.CreatedAt) {
			item.AnchorMessage = event.MessageID
		}
		if strings.HasPrefix(event.EventType, "tool.") {
			item.ToolCount++
		}
		if event.EventType == "artifact.created" {
			item.ArtifactCount++
		}
		if item.Preview == "" {
			item.Preview = summarizeWorkspaceEventPayload(event.PayloadJSON)
		}
		if item.RootCause == "" && workspaceEventToLogLevel(event) == "error" {
			item.RootCause = firstNonEmpty(summarizeWorkspaceEventPayload(event.PayloadJSON), event.EventType)
		}
		if workspaceEventToLogLevel(event) == "error" {
			item.Status = "failed"
		} else if (event.EventType == "message.completed" || event.EventType == "tool.completed") && item.Status != "failed" {
			item.Status = "completed"
		}
	}

	items := make([]observabilityTraceSummary, 0, len(aggregated))
	for _, item := range aggregated {
		item.TraceExplorer = buildTraceExplorerURL(r.config.TraceSearchPublicBaseURL, item.TraceID, item.StartAt, item.EndAt)
		if item.RootCause == "" {
			item.RootCause = item.Preview
		}
		if item.EndAt == "" {
			item.EndAt = item.LatestAt
		}
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LatestAt > items[j].LatestAt
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (r *Router) buildInstanceWorkspaceSessionsLocked(instanceID int, limit int) []workspaceSessionSummary {
	sessions := r.filterWorkspaceSessionsByInstance(instanceID)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt > sessions[j].UpdatedAt
	})
	items := make([]workspaceSessionSummary, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, r.buildWorkspaceSessionSummaryLocked(session))
		if limit > 0 && len(items) >= limit {
			break
		}
	}
	return items
}

func (r *Router) buildTraceSummaryIndexLocked(instanceID int) map[string]observabilityTraceSummary {
	items := r.buildInstanceTraceSummariesLocked(instanceID, 0)
	index := make(map[string]observabilityTraceSummary, len(items))
	for _, item := range items {
		index[item.TraceID] = item
	}
	return index
}

func (r *Router) buildObservabilityAlertsLocked(alerts []models.Alert, scope string) []observabilityAlertRecord {
	items := make([]observabilityAlertRecord, 0, len(alerts))
	for _, alert := range alerts {
		items = append(items, r.buildObservabilityAlertLocked(alert, scope))
	}
	return items
}

func (r *Router) buildObservabilityAlertLocked(alert models.Alert, scope string) observabilityAlertRecord {
	record := observabilityAlertRecord{
		ID:          alert.ID,
		InstanceID:  alert.InstanceID,
		Severity:    alert.Severity,
		Status:      alert.Status,
		MetricKey:   alert.MetricKey,
		Summary:     alert.Summary,
		TriggeredAt: alert.TriggeredAt,
		DetailPath:  buildScopedInstancePath(scope, alert.InstanceID),
	}

	trace := r.matchAlertTraceLocked(alert)
	if trace.TraceID == "" {
		record.WorkspacePath = buildScopedWorkspacePath(scope, alert.InstanceID, 0, 0, "")
		return record
	}

	record.SessionID = trace.SessionID
	record.MessageID = trace.AnchorMessage
	record.TraceID = trace.TraceID
	record.TraceStart = trace.StartAt
	record.TraceEnd = trace.EndAt
	record.RootCause = trace.RootCause
	record.WorkspacePath = buildScopedWorkspacePathWithWindow(scope, alert.InstanceID, trace.SessionID, trace.AnchorMessage, trace.TraceID, trace.StartAt, trace.EndAt)
	record.TraceExplorerURL = trace.TraceExplorer
	return record
}

func (r *Router) matchAlertTraceLocked(alert models.Alert) observabilityTraceSummary {
	traces := r.buildInstanceTraceSummariesLocked(alert.InstanceID, 0)
	if len(traces) == 0 {
		return observabilityTraceSummary{}
	}

	alertTime, ok := parseRFC3339(alert.TriggeredAt)
	if !ok {
		return traces[0]
	}

	var matched observabilityTraceSummary
	matchedDiff := 1 << 30
	for _, trace := range traces {
		traceTime, ok := parseRFC3339(firstNonEmpty(trace.EndAt, trace.LatestAt, trace.StartAt))
		if !ok {
			continue
		}
		diff := absMinutes(alertTime.Sub(traceTime))
		if diff > 180 {
			continue
		}
		score := diff
		if trace.Status == "failed" {
			score -= 15
		}
		if strings.Contains(strings.ToLower(firstNonEmpty(trace.RootCause, trace.Preview)), strings.ToLower(alert.MetricKey)) {
			score -= 5
		}
		if matched.TraceID == "" || score < matchedDiff {
			matched = trace
			matchedDiff = score
		}
	}
	if matched.TraceID != "" {
		return matched
	}
	return traces[0]
}

func ensureTraceSummary(items map[string]*observabilityTraceSummary, traceID string) *observabilityTraceSummary {
	if item, ok := items[traceID]; ok {
		return item
	}
	item := &observabilityTraceSummary{
		TraceID: traceID,
		Status:  "active",
	}
	items[traceID] = item
	return item
}

func summarizeWorkspaceEventPayload(payloadJSON string) string {
	if strings.TrimSpace(payloadJSON) == "" {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return trimPreviewText(payloadJSON, 96)
	}
	for _, key := range []string{"detail", "errorMessage", "error", "reasoning", "delta", "content"} {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return trimPreviewText(value, 96)
		}
	}
	if artifact, ok := payload["artifact"].(map[string]any); ok {
		if title, ok := artifact["title"].(string); ok {
			return trimPreviewText(title, 96)
		}
	}
	if message, ok := payload["message"].(map[string]any); ok {
		if content, ok := message["content"].(string); ok {
			return trimPreviewText(content, 96)
		}
	}
	return trimPreviewText(payloadJSON, 96)
}

func extractTraceIDFromEventPayload(payloadJSON string) string {
	if strings.TrimSpace(payloadJSON) == "" {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return ""
	}
	if value, ok := payload["traceId"].(string); ok {
		return strings.TrimSpace(value)
	}
	if message, ok := payload["message"].(map[string]any); ok {
		if value, ok := message["traceId"].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func severityToLogLevel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical":
		return "error"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}

func auditResultToLogLevel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "failed", "error":
		return "error"
	case "running", "pending":
		return "warning"
	default:
		return "info"
	}
}

func diagnosticStatusToLogLevel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "failed", "timeout":
		return "error"
	case "blocked":
		return "warning"
	default:
		return "info"
	}
}

func workspaceEventToLogLevel(event models.WorkspaceMessageEvent) string {
	if strings.Contains(event.EventType, "failed") || strings.Contains(strings.ToLower(event.PayloadJSON), "error") {
		return "error"
	}
	if strings.HasPrefix(event.EventType, "tool.") || strings.HasPrefix(event.EventType, "reasoning.") {
		return "warning"
	}
	return "info"
}

func buildScopedInstancePath(scope string, instanceID int) string {
	if instanceID <= 0 {
		return ""
	}
	if scope == "admin" {
		return fmt.Sprintf("/admin/instances/%d?tab=monitoring", instanceID)
	}
	return fmt.Sprintf("/portal/instances/%d", instanceID)
}

func buildScopedWorkspacePath(scope string, instanceID int, sessionID int, messageID int, traceID string) string {
	return buildScopedWorkspacePathWithWindow(scope, instanceID, sessionID, messageID, traceID, "", "")
}

func buildScopedWorkspacePathWithWindow(scope string, instanceID int, sessionID int, messageID int, traceID string, traceStart string, traceEnd string) string {
	if instanceID <= 0 {
		return ""
	}
	base := fmt.Sprintf("/portal/instances/%d/workspace", instanceID)
	if scope == "admin" {
		base = fmt.Sprintf("/admin/instances/%d/workspace", instanceID)
	}
	query := url.Values{}
	if sessionID > 0 {
		query.Set("sessionId", strconv.Itoa(sessionID))
	}
	if messageID > 0 {
		query.Set("messageId", strconv.Itoa(messageID))
	}
	if strings.TrimSpace(traceID) != "" {
		query.Set("traceId", strings.TrimSpace(traceID))
	}
	if strings.TrimSpace(traceStart) != "" {
		query.Set("traceStart", strings.TrimSpace(traceStart))
	}
	if strings.TrimSpace(traceEnd) != "" {
		query.Set("traceEnd", strings.TrimSpace(traceEnd))
	}
	if encoded := query.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

func buildTraceExplorerURL(base string, traceID string, traceStart string, traceEnd string) string {
	base = strings.TrimSpace(base)
	traceID = strings.TrimSpace(traceID)
	if base == "" || traceID == "" {
		return ""
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return ""
	}
	query := parsed.Query()
	query.Set("traceId", traceID)
	if strings.TrimSpace(traceStart) != "" {
		query.Set("from", strings.TrimSpace(traceStart))
	}
	if strings.TrimSpace(traceEnd) != "" {
		query.Set("to", strings.TrimSpace(traceEnd))
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func (r *Router) traceSearchEnabled() bool {
	return r != nil && r.config.TraceSearchEnabled && strings.TrimSpace(r.config.TraceSearchURL) != "" && strings.TrimSpace(r.config.TraceSearchIndex) != ""
}

func (r *Router) traceSearchProvider() string {
	if r.traceSearchEnabled() {
		return "opensearch"
	}
	return "workspace"
}

func traceStartAt(index map[string]observabilityTraceSummary, traceID string) string {
	return traceSummaryValue(index, traceID).StartAt
}

func traceEndAt(index map[string]observabilityTraceSummary, traceID string) string {
	return traceSummaryValue(index, traceID).EndAt
}

func traceRootCause(index map[string]observabilityTraceSummary, traceID string) string {
	return traceSummaryValue(index, traceID).RootCause
}

func traceExplorerURL(index map[string]observabilityTraceSummary, traceID string) string {
	return traceSummaryValue(index, traceID).TraceExplorer
}

func traceSummaryValue(index map[string]observabilityTraceSummary, traceID string) observabilityTraceSummary {
	if index == nil {
		return observabilityTraceSummary{}
	}
	value, ok := index[strings.TrimSpace(traceID)]
	if !ok {
		return observabilityTraceSummary{}
	}
	return value
}

func trimPreviewText(value string, limit int) string {
	trimmed := strings.TrimSpace(value)
	if limit <= 0 || len(trimmed) <= limit {
		return trimmed
	}
	if limit <= 3 {
		return trimmed[:limit]
	}
	return trimmed[:limit-3] + "..."
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func minRFC3339(current string, candidate string) string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return current
	}
	if strings.TrimSpace(current) == "" || candidate < current {
		return candidate
	}
	return current
}

func absMinutes(value time.Duration) int {
	if value < 0 {
		value = -value
	}
	return int(value / time.Minute)
}
