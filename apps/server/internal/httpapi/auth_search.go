package httpapi

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type searchResult struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Level         string `json:"level"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	InstanceID    string `json:"instanceId,omitempty"`
	SessionID     string `json:"sessionId,omitempty"`
	MessageID     string `json:"messageId,omitempty"`
	TraceID       string `json:"traceId,omitempty"`
	TraceStart    string `json:"traceStart,omitempty"`
	TraceEnd      string `json:"traceEnd,omitempty"`
	RootCause     string `json:"rootCause,omitempty"`
	TraceExplorer string `json:"traceExplorerUrl,omitempty"`
	Source        string `json:"source"`
	CreatedAt     string `json:"createdAt"`
	InstancePath  string `json:"instancePath,omitempty"`
	WorkspacePath string `json:"workspacePath,omitempty"`
}

type searchLogsQuery struct {
	Query      string
	Kind       string
	InstanceID string
	TraceID    string
	Scope      string
	From       string
	To         string
	Page       int
	PageSize   int
}

type searchLogsResponse struct {
	Backend      string         `json:"backend"`
	TraceBackend string         `json:"traceBackend"`
	Page         int            `json:"page"`
	PageSize     int            `json:"pageSize"`
	Total        int            `json:"total"`
	HasMore      bool           `json:"hasMore"`
	ExportPath   string         `json:"exportPath"`
	Items        []searchResult `json:"items"`
}

func (r *Router) handleAuthConfig(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"provider":        "keycloak",
		"enabled":         r.config.KeycloakEnabled,
		"baseUrl":         r.config.KeycloakBaseURL,
		"realm":           r.config.KeycloakRealm,
		"clientId":        r.config.KeycloakClientID,
		"defaultRedirect": r.config.KeycloakRedirectURL,
		"mockUserEnabled": false,
		"providers": []map[string]any{
			{
				"provider": "keycloak",
				"enabled":  r.config.KeycloakEnabled,
				"mode":     "oidc",
			},
			{
				"provider": "wechat",
				"enabled":  r.config.WeChatLoginEnabled,
				"mode":     "website-qr-login",
			},
		},
	})
}

func (r *Router) handleAuthSession(w http.ResponseWriter, req *http.Request) {
	if session, ok := r.readAuthSession(req); ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": session.Authenticated,
			"provider":      session.Provider,
			"user": map[string]any{
				"name":    session.Name,
				"email":   session.Email,
				"role":    session.Role,
				"openId":  session.OpenID,
				"unionId": session.UnionID,
			},
		})
		return
	}

	if !r.config.KeycloakEnabled {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
			"provider":      "none",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": false,
		"provider":      "none",
	})
}

func (r *Router) handleAuthKeycloakURL(w http.ResponseWriter, req *http.Request) {
	if !r.config.KeycloakEnabled || r.config.KeycloakBaseURL == "" || r.config.KeycloakRealm == "" || r.config.KeycloakClientID == "" {
		http.Error(w, "keycloak not configured", http.StatusBadRequest)
		return
	}

	redirectURI := req.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = r.config.KeycloakRedirectURL
	}
	if redirectURI == "" {
		http.Error(w, "redirect_uri is required", http.StatusBadRequest)
		return
	}

	flow := authFlowState{
		Provider:    "keycloak",
		State:       generateAuthRandom(24),
		Nonce:       generateAuthRandom(16),
		RedirectURI: redirectURI,
		Next:        req.URL.Query().Get("next"),
		ExpiresAt:   time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339),
	}
	r.writeAuthFlow(w, flow)

	base := strings.TrimRight(r.config.KeycloakBaseURL, "/")
	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth?client_id=%s&response_type=code&scope=openid%%20profile%%20email&redirect_uri=%s&state=%s&nonce=%s",
		base,
		url.PathEscape(r.config.KeycloakRealm),
		url.QueryEscape(r.config.KeycloakClientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(flow.State),
		url.QueryEscape(flow.Nonce),
	)

	writeJSON(w, http.StatusOK, map[string]any{
		"url": authURL,
	})
}

func (r *Router) handleSearchConfig(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"provider": "opensearch",
		"enabled":  r.config.OpenSearchEnabled,
		"url":      r.config.OpenSearchURL,
		"index":    r.config.OpenSearchIndex,
		"export": map[string]any{
			"enabled": true,
		},
		"trace": map[string]any{
			"provider":      r.traceSearchProvider(),
			"enabled":       r.traceSearchEnabled(),
			"url":           r.config.TraceSearchURL,
			"index":         r.config.TraceSearchIndex,
			"publicBaseUrl": r.config.TraceSearchPublicBaseURL,
		},
	})
}

func (r *Router) handleSearchLogs(w http.ResponseWriter, req *http.Request) {
	query, err := parseSearchLogsQuery(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	scope := normalizeSearchScope(query.Scope)
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	if !r.config.OpenSearchEnabled || r.config.OpenSearchURL == "" || r.config.OpenSearchIndex == "" {
		http.Error(w, "opensearch search backend is required", http.StatusServiceUnavailable)
		return
	}

	result, err := r.searchLogsWithOpenSearch(query, actor, scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("opensearch search failed: %v", err), http.StatusBadGateway)
		return
	}
	result.ExportPath = buildSearchExportPath(query)
	result.TraceBackend = r.traceSearchProvider()
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleSearchLogsExportCSV(w http.ResponseWriter, req *http.Request) {
	query, err := parseSearchLogsQuery(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	scope := normalizeSearchScope(query.Scope)
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	query.Page = 1
	if query.PageSize < 1 || query.PageSize > 2000 {
		query.PageSize = 2000
	}

	if !r.config.OpenSearchEnabled || r.config.OpenSearchURL == "" || r.config.OpenSearchIndex == "" {
		http.Error(w, "opensearch search backend is required", http.StatusServiceUnavailable)
		return
	}

	result, err := r.searchLogsWithOpenSearch(query, actor, scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("opensearch search failed: %v", err), http.StatusBadGateway)
		return
	}
	items := result.Items
	if len(items) > query.PageSize {
		items = items[:query.PageSize]
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{"id", "kind", "level", "title", "message", "instanceId", "sessionId", "messageId", "traceId", "traceStart", "traceEnd", "rootCause", "source", "createdAt", "instancePath", "workspacePath", "traceExplorerUrl"})
	for _, item := range items {
		_ = writer.Write([]string{
			item.ID,
			item.Kind,
			item.Level,
			item.Title,
			item.Message,
			item.InstanceID,
			item.SessionID,
			item.MessageID,
			item.TraceID,
			item.TraceStart,
			item.TraceEnd,
			item.RootCause,
			item.Source,
			item.CreatedAt,
			item.InstancePath,
			item.WorkspacePath,
			item.TraceExplorer,
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("openclaw-search-%s.csv", time.Now().UTC().Format("20060102-150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	_, _ = w.Write(buffer.Bytes())
}

func (r *Router) handleSearchTrace(w http.ResponseWriter, req *http.Request) {
	traceID := strings.TrimSpace(strings.TrimPrefix(req.URL.Path, "/api/v1/search/traces/"))
	if traceID == "" {
		http.Error(w, "trace id is required", http.StatusBadRequest)
		return
	}
	instanceID := parseSearchInt(req.URL.Query().Get("instanceId"))
	if instanceID <= 0 {
		http.Error(w, "instanceId is required", http.StatusBadRequest)
		return
	}
	scope := normalizeSearchScope(req.URL.Query().Get("scope"))
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}
	instance, found := r.findInstance(instanceID)
	if !found || !r.canAccessWorkspaceInstance(actor, instance, scope) {
		http.NotFound(w, req)
		return
	}

	summary := r.traceSummaryValueForInstance(instanceID, traceID)
	writeJSON(w, http.StatusOK, map[string]any{
		"backend":     r.traceSearchProvider(),
		"traceId":     traceID,
		"instanceId":  instanceID,
		"sessionId":   summary.SessionID,
		"startAt":     summary.StartAt,
		"endAt":       summary.EndAt,
		"rootCause":   summary.RootCause,
		"explorerUrl": summary.TraceExplorer,
		"spans":       r.buildTraceSpans(instanceID, traceID),
	})
}

func (r *Router) searchLogsInMemory(query searchLogsQuery, actor workspaceActor, scope string) []searchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]searchResult, 0)
	traceIndexByInstance := make(map[int]map[string]observabilityTraceSummary)

	for _, audit := range r.data.Audits {
		instanceIDValue := intToStringIfPositive(audit.TargetID)
		sessionID := audit.Metadata["sessionId"]
		if audit.Target != "instance" {
			instanceIDValue = audit.Metadata["instanceId"]
		}
		parsedInstanceID, _ := strconv.Atoi(instanceIDValue)
		if !r.canAccessSearchRecord(actor, scope, audit.TenantID, parsedInstanceID) {
			continue
		}
		messageID := audit.Metadata["messageId"]
		traceID := audit.Metadata["traceId"]
		traceIndex := traceIndexByInstance[parsedInstanceID]
		if traceIndex == nil && parsedInstanceID > 0 {
			traceIndex = r.buildTraceSummaryIndexLocked(parsedInstanceID)
			traceIndexByInstance[parsedInstanceID] = traceIndex
		}
		item := searchResult{
			ID:            fmt.Sprintf("audit-%d", audit.ID),
			Kind:          "audit",
			Level:         auditResultToLogLevel(audit.Result),
			Title:         audit.Action,
			Message:       audit.Result,
			InstanceID:    instanceIDValue,
			SessionID:     sessionID,
			MessageID:     messageID,
			TraceID:       traceID,
			TraceStart:    traceStartAt(traceIndex, traceID),
			TraceEnd:      traceEndAt(traceIndex, traceID),
			RootCause:     traceRootCause(traceIndex, traceID),
			TraceExplorer: traceExplorerURL(traceIndex, traceID),
			Source:        audit.Actor,
			CreatedAt:     audit.CreatedAt,
			InstancePath:  buildScopedInstancePath(scope, parsedInstanceID),
			WorkspacePath: buildScopedWorkspacePathWithWindow(scope, parsedInstanceID, parseSearchInt(sessionID), parseSearchInt(messageID), traceID, traceStartAt(traceIndex, traceID), traceEndAt(traceIndex, traceID)),
		}
		if matchesSearch(item, query) {
			items = append(items, item)
		}
	}

	for _, ticket := range r.data.Tickets {
		if !r.canAccessSearchRecord(actor, scope, ticket.TenantID, ticket.InstanceID) {
			continue
		}
		item := searchResult{
			ID:           fmt.Sprintf("ticket-%d", ticket.ID),
			Kind:         "ticket",
			Level:        ticket.Severity,
			Title:        ticket.Title,
			Message:      ticket.Description,
			InstanceID:   intToStringIfPositive(ticket.InstanceID),
			Source:       ticket.Reporter,
			CreatedAt:    ticket.UpdatedAt,
			InstancePath: buildScopedInstancePath(scope, ticket.InstanceID),
		}
		if matchesSearch(item, query) {
			items = append(items, item)
		}
	}

	for _, activity := range r.data.Activities {
		if scope != "admin" || !actor.GlobalAdmin {
			continue
		}
		item := searchResult{
			ID:        fmt.Sprintf("channel-%d", activity.ID),
			Kind:      "channel",
			Level:     "info",
			Title:     activity.Title,
			Message:   activity.Summary,
			Source:    activity.Type,
			CreatedAt: activity.CreatedAt,
		}
		if matchesSearch(item, query) {
			items = append(items, item)
		}
	}

	for _, alert := range r.data.Alerts {
		instance, found := r.findInstance(alert.InstanceID)
		if !found || !r.canAccessSearchRecord(actor, scope, instance.TenantID, alert.InstanceID) {
			continue
		}
		record := r.buildObservabilityAlertLocked(alert, scope)
		item := searchResult{
			ID:            fmt.Sprintf("alert-%d", alert.ID),
			Kind:          "alert",
			Level:         severityToLogLevel(alert.Severity),
			Title:         alert.MetricKey,
			Message:       alert.Summary,
			InstanceID:    intToStringIfPositive(alert.InstanceID),
			SessionID:     intToStringIfPositive(record.SessionID),
			MessageID:     intToStringIfPositive(record.MessageID),
			TraceID:       record.TraceID,
			TraceStart:    record.TraceStart,
			TraceEnd:      record.TraceEnd,
			RootCause:     record.RootCause,
			TraceExplorer: record.TraceExplorerURL,
			Source:        "alert",
			CreatedAt:     alert.TriggeredAt,
			InstancePath:  buildScopedInstancePath(scope, alert.InstanceID),
			WorkspacePath: record.WorkspacePath,
		}
		if matchesSearch(item, query) {
			items = append(items, item)
		}
	}

	for _, event := range r.data.WorkspaceMessageEvents {
		if !r.canAccessSearchRecord(actor, scope, event.TenantID, event.InstanceID) {
			continue
		}
		traceIndex := traceIndexByInstance[event.InstanceID]
		if traceIndex == nil && event.InstanceID > 0 {
			traceIndex = r.buildTraceSummaryIndexLocked(event.InstanceID)
			traceIndexByInstance[event.InstanceID] = traceIndex
		}
		traceID := firstNonEmpty(event.TraceID, extractTraceIDFromEventPayload(event.PayloadJSON))
		item := searchResult{
			ID:            fmt.Sprintf("workspace-event-%d", event.ID),
			Kind:          "workspace_event",
			Level:         workspaceEventToLogLevel(event),
			Title:         event.EventType,
			Message:       summarizeWorkspaceEventPayload(event.PayloadJSON),
			InstanceID:    intToStringIfPositive(event.InstanceID),
			SessionID:     intToStringIfPositive(event.SessionID),
			MessageID:     intToStringIfPositive(event.MessageID),
			TraceID:       traceID,
			TraceStart:    traceStartAt(traceIndex, traceID),
			TraceEnd:      traceEndAt(traceIndex, traceID),
			RootCause:     traceRootCause(traceIndex, traceID),
			TraceExplorer: traceExplorerURL(traceIndex, traceID),
			Source:        firstNonEmpty(event.Origin, "workspace"),
			CreatedAt:     event.CreatedAt,
			InstancePath:  buildScopedInstancePath(scope, event.InstanceID),
			WorkspacePath: buildScopedWorkspacePathWithWindow(scope, event.InstanceID, event.SessionID, event.MessageID, traceID, traceStartAt(traceIndex, traceID), traceEndAt(traceIndex, traceID)),
		}
		if matchesSearch(item, query) {
			items = append(items, item)
		}
	}

	for _, record := range r.data.DiagnosticCommandRecords {
		instance, found := r.findInstance(record.InstanceID)
		if !found || !r.canAccessSearchRecord(actor, scope, instance.TenantID, record.InstanceID) {
			continue
		}
		item := searchResult{
			ID:            fmt.Sprintf("diagnostic-%d", record.ID),
			Kind:          "diagnostic",
			Level:         diagnosticStatusToLogLevel(record.Status),
			Title:         record.CommandKey,
			Message:       fmt.Sprintf("%s [%s]", record.CommandText, record.Status),
			InstanceID:    intToStringIfPositive(record.InstanceID),
			SessionID:     intToStringIfPositive(record.SessionID),
			Source:        "diagnostic",
			CreatedAt:     record.ExecutedAt,
			InstancePath:  buildScopedInstancePath(scope, record.InstanceID),
			WorkspacePath: buildScopedWorkspacePath(scope, record.InstanceID, record.SessionID, 0, ""),
		}
		if matchesSearch(item, query) {
			items = append(items, item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})

	return items
}

func matchesSearch(item searchResult, query searchLogsQuery) bool {
	if query.Kind != "" && item.Kind != query.Kind {
		return false
	}
	if query.InstanceID != "" && item.InstanceID != query.InstanceID {
		return false
	}
	if query.TraceID != "" && item.TraceID != query.TraceID {
		return false
	}
	if query.From != "" && item.CreatedAt < query.From {
		return false
	}
	if query.To != "" && item.CreatedAt > query.To {
		return false
	}
	if query.Query == "" {
		return true
	}

	needle := strings.ToLower(query.Query)
	haystack := strings.ToLower(item.Title + " " + item.Message + " " + item.Source + " " + item.RootCause + " " + item.TraceID)
	return strings.Contains(haystack, needle)
}

func (r *Router) searchLogsWithOpenSearch(query searchLogsQuery, actor workspaceActor, scope string) (searchLogsResponse, error) {
	body := map[string]any{
		"from": maxInt((query.Page-1)*query.PageSize, 0),
		"size": query.PageSize,
		"sort": []map[string]any{
			{"createdAt": map[string]any{"order": "desc"}},
			{"_id": map[string]any{"order": "desc"}},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{},
			},
		},
	}

	must := body["query"].(map[string]any)["bool"].(map[string]any)["must"].([]map[string]any)
	if query.Query != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  query.Query,
				"fields": []string{"title", "message", "source", "rootCause", "traceId"},
			},
		})
	}
	if query.Kind != "" {
		must = append(must, map[string]any{
			"term": map[string]any{"kind.keyword": query.Kind},
		})
	}
	if query.InstanceID != "" {
		must = append(must, map[string]any{
			"term": map[string]any{"instanceId.keyword": query.InstanceID},
		})
	}
	if query.TraceID != "" {
		must = append(must, map[string]any{
			"term": map[string]any{"traceId.keyword": query.TraceID},
		})
	}
	if query.From != "" || query.To != "" {
		rangeQuery := map[string]any{}
		if query.From != "" {
			rangeQuery["gte"] = query.From
		}
		if query.To != "" {
			rangeQuery["lte"] = query.To
		}
		must = append(must, map[string]any{
			"range": map[string]any{"createdAt": rangeQuery},
		})
	}
	if !actor.GlobalAdmin && actor.TenantID > 0 {
		must = append(must, map[string]any{
			"term": map[string]any{"tenantId": actor.TenantID},
		})
	}
	body["query"].(map[string]any)["bool"].(map[string]any)["must"] = must

	payload, err := json.Marshal(body)
	if err != nil {
		return searchLogsResponse{}, err
	}

	searchURL := strings.TrimRight(r.config.OpenSearchURL, "/") + "/" + strings.Trim(r.config.OpenSearchIndex, "/") + "/_search"
	request, err := http.NewRequest(http.MethodPost, searchURL, bytes.NewReader(payload))
	if err != nil {
		return searchLogsResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	if r.config.OpenSearchUsername != "" {
		request.SetBasicAuth(r.config.OpenSearchUsername, r.config.OpenSearchPassword)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return searchLogsResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		raw, _ := io.ReadAll(response.Body)
		return searchLogsResponse{}, fmt.Errorf("opensearch search failed: %s", string(raw))
	}

	var payloadResponse struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string `json:"_id"`
				Source struct {
					Kind          string `json:"kind"`
					Level         string `json:"level"`
					Title         string `json:"title"`
					Message       string `json:"message"`
					InstanceID    string `json:"instanceId"`
					SessionID     string `json:"sessionId"`
					MessageID     string `json:"messageId"`
					TraceID       string `json:"traceId"`
					TraceStart    string `json:"traceStart"`
					TraceEnd      string `json:"traceEnd"`
					RootCause     string `json:"rootCause"`
					TraceExplorer string `json:"traceExplorerUrl"`
					Source        string `json:"source"`
					CreatedAt     string `json:"createdAt"`
					InstancePath  string `json:"instancePath"`
					WorkspacePath string `json:"workspacePath"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(response.Body).Decode(&payloadResponse); err != nil {
		return searchLogsResponse{}, err
	}

	items := make([]searchResult, 0, len(payloadResponse.Hits.Hits))
	for _, hit := range payloadResponse.Hits.Hits {
		items = append(items, searchResult{
			ID:            hit.ID,
			Kind:          hit.Source.Kind,
			Level:         hit.Source.Level,
			Title:         hit.Source.Title,
			Message:       hit.Source.Message,
			InstanceID:    hit.Source.InstanceID,
			SessionID:     hit.Source.SessionID,
			MessageID:     hit.Source.MessageID,
			TraceID:       hit.Source.TraceID,
			TraceStart:    hit.Source.TraceStart,
			TraceEnd:      hit.Source.TraceEnd,
			RootCause:     hit.Source.RootCause,
			TraceExplorer: hit.Source.TraceExplorer,
			Source:        hit.Source.Source,
			CreatedAt:     hit.Source.CreatedAt,
			InstancePath:  hit.Source.InstancePath,
			WorkspacePath: hit.Source.WorkspacePath,
		})
	}

	return searchLogsResponse{
		Backend:      "opensearch",
		TraceBackend: r.traceSearchProvider(),
		Page:         query.Page,
		PageSize:     query.PageSize,
		Total:        payloadResponse.Hits.Total.Value,
		HasMore:      query.Page*query.PageSize < payloadResponse.Hits.Total.Value,
		Items:        items,
	}, nil
}

func parseSearchLogsQuery(req *http.Request) (searchLogsQuery, error) {
	query := searchLogsQuery{
		Query:      strings.TrimSpace(req.URL.Query().Get("q")),
		Kind:       strings.TrimSpace(req.URL.Query().Get("kind")),
		InstanceID: strings.TrimSpace(req.URL.Query().Get("instanceId")),
		TraceID:    strings.TrimSpace(req.URL.Query().Get("traceId")),
		Scope:      normalizeSearchScope(req.URL.Query().Get("scope")),
		From:       strings.TrimSpace(req.URL.Query().Get("from")),
		To:         strings.TrimSpace(req.URL.Query().Get("to")),
		Page:       1,
		PageSize:   20,
	}
	if query.InstanceID != "" && parseSearchInt(query.InstanceID) <= 0 {
		return searchLogsQuery{}, fmt.Errorf("instanceId must be a positive integer")
	}
	if value := strings.TrimSpace(req.URL.Query().Get("page")); value != "" {
		query.Page = maxInt(parseSearchInt(value), 1)
	}
	if value := strings.TrimSpace(req.URL.Query().Get("pageSize")); value != "" {
		query.PageSize = clampInt(parseSearchInt(value), 1, 200)
	}
	if query.From != "" {
		if _, ok := parseRFC3339(query.From); !ok {
			return searchLogsQuery{}, fmt.Errorf("from must be RFC3339")
		}
	}
	if query.To != "" {
		if _, ok := parseRFC3339(query.To); !ok {
			return searchLogsQuery{}, fmt.Errorf("to must be RFC3339")
		}
	}
	if query.From != "" && query.To != "" && query.From > query.To {
		return searchLogsQuery{}, fmt.Errorf("from must be earlier than to")
	}
	return query, nil
}

func paginateSearchResults(items []searchResult, page int, pageSize int) []searchResult {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []searchResult{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func buildSearchExportPath(query searchLogsQuery) string {
	params := url.Values{}
	if query.Query != "" {
		params.Set("q", query.Query)
	}
	if query.Kind != "" {
		params.Set("kind", query.Kind)
	}
	if query.InstanceID != "" {
		params.Set("instanceId", query.InstanceID)
	}
	if query.TraceID != "" {
		params.Set("traceId", query.TraceID)
	}
	if query.From != "" {
		params.Set("from", query.From)
	}
	if query.To != "" {
		params.Set("to", query.To)
	}
	params.Set("scope", query.Scope)
	params.Set("pageSize", "2000")
	return "/api/v1/search/logs/export.csv?" + params.Encode()
}

func (r *Router) traceSummaryValueForInstance(instanceID int, traceID string) observabilityTraceSummary {
	if instanceID <= 0 || strings.TrimSpace(traceID) == "" {
		return observabilityTraceSummary{}
	}
	traceIndex := r.buildTraceSummaryIndexLocked(instanceID)
	return traceSummaryValue(traceIndex, traceID)
}

func (r *Router) buildTraceSpans(instanceID int, traceID string) []map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]map[string]any, 0)
	for _, message := range r.filterWorkspaceMessagesByInstance(instanceID) {
		if strings.TrimSpace(message.TraceID) != strings.TrimSpace(traceID) {
			continue
		}
		endAt := firstNonEmpty(message.DeliveredAt, message.UpdatedAt, message.CreatedAt)
		items = append(items, map[string]any{
			"spanId":       fmt.Sprintf("msg-%d", message.ID),
			"parentSpanId": "",
			"name":         fmt.Sprintf("workspace.message.%s", message.Role),
			"kind":         "internal",
			"status":       message.Status,
			"serviceName":  firstNonEmpty(message.Origin, "workspace"),
			"startTime":    message.CreatedAt,
			"endTime":      endAt,
			"durationMs":   durationBetweenRFC3339(message.CreatedAt, endAt),
			"detail":       trimPreviewText(firstNonEmpty(message.ErrorMessage, message.Content), 120),
		})
	}
	for _, event := range r.filterWorkspaceMessageEventsByInstance(instanceID) {
		eventTraceID := firstNonEmpty(event.TraceID, extractTraceIDFromEventPayload(event.PayloadJSON))
		if strings.TrimSpace(eventTraceID) != strings.TrimSpace(traceID) {
			continue
		}
		items = append(items, map[string]any{
			"spanId":       fmt.Sprintf("evt-%d", event.ID),
			"parentSpanId": firstNonEmpty(intToStringIfPositive(event.MessageID), ""),
			"name":         event.EventType,
			"kind":         "internal",
			"status":       workspaceEventToLogLevel(event),
			"serviceName":  firstNonEmpty(event.Origin, "workspace"),
			"startTime":    event.CreatedAt,
			"endTime":      event.CreatedAt,
			"durationMs":   0,
			"detail":       summarizeWorkspaceEventPayload(event.PayloadJSON),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		left, _ := items[i]["startTime"].(string)
		right, _ := items[j]["startTime"].(string)
		return left < right
	})
	return items
}

func durationBetweenRFC3339(start string, end string) int {
	startTime, ok := parseRFC3339(start)
	if !ok {
		return 0
	}
	endTime, ok := parseRFC3339(end)
	if !ok {
		return 0
	}
	if endTime.Before(startTime) {
		return 0
	}
	return int(endTime.Sub(startTime) / time.Millisecond)
}

func normalizeSearchScope(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "admin") {
		return "admin"
	}
	return "portal"
}

func (r *Router) canAccessSearchRecord(actor workspaceActor, scope string, tenantID int, instanceID int) bool {
	if tenantID > 0 {
		return actor.canAccessTenant(scope, tenantID)
	}
	if instanceID > 0 {
		instance, found := r.findInstance(instanceID)
		if !found {
			return false
		}
		return r.canAccessWorkspaceInstance(actor, instance, scope)
	}
	return scope == "admin" && actor.GlobalAdmin
}

func parseSearchInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}

func intToStringIfPositive(value int) string {
	if value <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", value)
}
