package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type searchResult struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Level      string `json:"level"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	InstanceID string `json:"instanceId,omitempty"`
	Source     string `json:"source"`
	CreatedAt  string `json:"createdAt"`
}

func (r *Router) handleAuthConfig(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"provider":         "keycloak",
		"enabled":          r.config.KeycloakEnabled,
		"baseUrl":          r.config.KeycloakBaseURL,
		"realm":            r.config.KeycloakRealm,
		"clientId":         r.config.KeycloakClientID,
		"defaultRedirect":  r.config.KeycloakRedirectURL,
		"mockUserEnabled":  r.config.KeycloakMockUser != "",
	})
}

func (r *Router) handleAuthSession(w http.ResponseWriter, req *http.Request) {
	if !r.config.KeycloakEnabled && r.config.KeycloakMockUser == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
			"provider":      "mock",
		})
		return
	}

	name := r.config.KeycloakMockUser
	if name == "" {
		name = "Keycloak Demo User"
	}
	email := r.config.KeycloakMockEmail
	if email == "" {
		email = "keycloak.demo@example.com"
	}
	role := r.config.KeycloakMockRole
	if role == "" {
		role = "tenant_admin"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"provider":      "keycloak",
		"user": map[string]any{
			"name":  name,
			"email": email,
			"role":  role,
		},
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

	base := strings.TrimRight(r.config.KeycloakBaseURL, "/")
	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth?client_id=%s&response_type=code&scope=openid%%20profile%%20email&redirect_uri=%s",
		base,
		url.PathEscape(r.config.KeycloakRealm),
		url.QueryEscape(r.config.KeycloakClientID),
		url.QueryEscape(redirectURI),
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
	})
}

func (r *Router) handleSearchLogs(w http.ResponseWriter, req *http.Request) {
	query := strings.TrimSpace(req.URL.Query().Get("q"))
	kind := strings.TrimSpace(req.URL.Query().Get("kind"))
	instanceID := strings.TrimSpace(req.URL.Query().Get("instanceId"))

	if r.config.OpenSearchEnabled && r.config.OpenSearchURL != "" && r.config.OpenSearchIndex != "" {
		if items, err := r.searchLogsWithOpenSearch(query, kind, instanceID); err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"backend": "opensearch",
				"items":   items,
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"backend": "mock",
		"items":   r.searchLogsInMemory(query, kind, instanceID),
	})
}

func (r *Router) searchLogsInMemory(query string, kind string, instanceID string) []searchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]searchResult, 0)

	for _, audit := range r.data.Audits {
		item := searchResult{
			ID:         fmt.Sprintf("audit-%d", audit.ID),
			Kind:       "audit",
			Level:      "info",
			Title:      audit.Action,
			Message:    audit.Result,
			InstanceID: intToStringIfPositive(audit.TargetID),
			Source:     audit.Actor,
			CreatedAt:  audit.CreatedAt,
		}
		if matchesSearch(item, query, kind, instanceID) {
			items = append(items, item)
		}
	}

	for _, ticket := range r.data.Tickets {
		item := searchResult{
			ID:         fmt.Sprintf("ticket-%d", ticket.ID),
			Kind:       "ticket",
			Level:      ticket.Severity,
			Title:      ticket.Title,
			Message:    ticket.Description,
			InstanceID: intToStringIfPositive(ticket.InstanceID),
			Source:     ticket.Reporter,
			CreatedAt:  ticket.UpdatedAt,
		}
		if matchesSearch(item, query, kind, instanceID) {
			items = append(items, item)
		}
	}

	for _, activity := range r.data.Activities {
		item := searchResult{
			ID:         fmt.Sprintf("channel-%d", activity.ID),
			Kind:       "channel",
			Level:      "info",
			Title:      activity.Title,
			Message:    activity.Summary,
			Source:     activity.Type,
			CreatedAt:  activity.CreatedAt,
		}
		if matchesSearch(item, query, kind, instanceID) {
			items = append(items, item)
		}
	}

	return items
}

func matchesSearch(item searchResult, query string, kind string, instanceID string) bool {
	if kind != "" && item.Kind != kind {
		return false
	}
	if instanceID != "" && item.InstanceID != instanceID {
		return false
	}
	if query == "" {
		return true
	}

	needle := strings.ToLower(query)
	haystack := strings.ToLower(item.Title + " " + item.Message + " " + item.Source)
	return strings.Contains(haystack, needle)
}

func (r *Router) searchLogsWithOpenSearch(query string, kind string, instanceID string) ([]searchResult, error) {
	body := map[string]any{
		"size": 50,
		"sort": []map[string]any{
			{"createdAt": map[string]any{"order": "desc"}},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{},
			},
		},
	}

	must := body["query"].(map[string]any)["bool"].(map[string]any)["must"].([]map[string]any)
	if query != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"title", "message", "source"},
			},
		})
	}
	if kind != "" {
		must = append(must, map[string]any{
			"term": map[string]any{"kind.keyword": kind},
		})
	}
	if instanceID != "" {
		must = append(must, map[string]any{
			"term": map[string]any{"instanceId.keyword": instanceID},
		})
	}
	body["query"].(map[string]any)["bool"].(map[string]any)["must"] = must

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	searchURL := strings.TrimRight(r.config.OpenSearchURL, "/") + "/" + strings.Trim(r.config.OpenSearchIndex, "/") + "/_search"
	request, err := http.NewRequest(http.MethodPost, searchURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if r.config.OpenSearchUsername != "" {
		request.SetBasicAuth(r.config.OpenSearchUsername, r.config.OpenSearchPassword)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		raw, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("opensearch search failed: %s", string(raw))
	}

	var payloadResponse struct {
		Hits struct {
			Hits []struct {
				ID     string `json:"_id"`
				Source struct {
					Kind       string `json:"kind"`
					Level      string `json:"level"`
					Title      string `json:"title"`
					Message    string `json:"message"`
					InstanceID string `json:"instanceId"`
					Source     string `json:"source"`
					CreatedAt  string `json:"createdAt"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(response.Body).Decode(&payloadResponse); err != nil {
		return nil, err
	}

	items := make([]searchResult, 0, len(payloadResponse.Hits.Hits))
	for _, hit := range payloadResponse.Hits.Hits {
		items = append(items, searchResult{
			ID:         hit.ID,
			Kind:       hit.Source.Kind,
			Level:      hit.Source.Level,
			Title:      hit.Source.Title,
			Message:    hit.Source.Message,
			InstanceID: hit.Source.InstanceID,
			Source:     hit.Source.Source,
			CreatedAt:  hit.Source.CreatedAt,
		})
	}

	return items, nil
}

func intToStringIfPositive(value int) string {
	if value <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", value)
}
