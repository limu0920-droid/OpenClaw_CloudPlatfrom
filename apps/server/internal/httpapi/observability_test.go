package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"openclaw/platformapi/internal/models"
)

func TestAdminInstanceDetailIncludesObservabilityContext(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/admin/instances/100", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected admin instance detail status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		WorkspaceSessions []struct {
			ID int `json:"id"`
		} `json:"workspaceSessions"`
		BridgeSummary struct {
			TraceCount   int `json:"traceCount"`
			RecentTraces []struct {
				TraceID string `json:"traceId"`
			} `json:"recentTraces"`
		} `json:"bridgeSummary"`
		RuntimeLogs []struct {
			WorkspacePath string `json:"workspacePath"`
			SessionID     int    `json:"sessionId"`
		} `json:"runtimeLogs"`
	}
	decodeResponse(t, recorder, &response)

	if len(response.WorkspaceSessions) == 0 {
		t.Fatal("expected related workspace sessions in admin instance detail")
	}
	if response.BridgeSummary.TraceCount == 0 || len(response.BridgeSummary.RecentTraces) == 0 {
		t.Fatalf("expected bridge summary traces, got %#v", response.BridgeSummary)
	}
	foundWorkspaceContext := false
	for _, item := range response.RuntimeLogs {
		if item.WorkspacePath != "" || item.SessionID > 0 {
			foundWorkspaceContext = true
			break
		}
	}
	if !foundWorkspaceContext {
		t.Fatalf("expected runtime logs with workspace context, got %#v", response.RuntimeLogs)
	}
}

func TestSearchLogsRespectsTenantIsolationAndReturnsWorkspaceContext(t *testing.T) {
	searchBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST search request, got %s", req.Method)
		}
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read search body: %v", err)
		}
		body := string(raw)

		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(body, "\"instanceId.keyword\":\"200\"") {
			_, _ = w.Write([]byte(`{"hits":{"total":{"value":0},"hits":[]}}`))
			return
		}
		payload := map[string]any{
			"hits": map[string]any{
				"total": map[string]any{"value": 1},
				"hits": []map[string]any{
					{
						"_id": "evt-1",
						"_source": map[string]any{
							"kind":          "workspace_event",
							"level":         "info",
							"title":         "Workspace Event",
							"message":       "artifact created",
							"instanceId":    "100",
							"sessionId":     "1",
							"messageId":     "2",
							"traceId":       "trace-demo-1",
							"source":        "workspace",
							"createdAt":     "2026-04-05T10:08:26Z",
							"instancePath":  "/admin/instances/100",
							"workspacePath": "/admin/instances/100/workspace?sessionId=1",
						},
					},
				},
			},
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode search payload: %v", err)
		}
	}))
	defer searchBackend.Close()

	router := newTestRouter(ExternalConfig{
		OpenSearchEnabled: true,
		OpenSearchURL:     searchBackend.URL,
		OpenSearchIndex:   "openclaw-logs",
	})

	tenantCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})
	platformCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Platform Admin",
		Email:         "platform@openclaw.local",
		Role:          "platform_admin",
	})

	tenantSearch := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/search/logs?scope=admin&kind=alert&instanceId=200", nil, tenantCookie)
	if tenantSearch.Code != http.StatusOK {
		t.Fatalf("expected tenant search status 200, got %d: %s", tenantSearch.Code, tenantSearch.Body.String())
	}
	var tenantResponse struct {
		Items []searchResult `json:"items"`
	}
	decodeResponse(t, tenantSearch, &tenantResponse)
	if len(tenantResponse.Items) != 0 {
		t.Fatalf("expected tenant search to hide cross-tenant logs, got %#v", tenantResponse.Items)
	}

	platformSearch := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/search/logs?scope=admin&kind=workspace_event&instanceId=100", nil, platformCookie)
	if platformSearch.Code != http.StatusOK {
		t.Fatalf("expected platform search status 200, got %d: %s", platformSearch.Code, platformSearch.Body.String())
	}
	var platformResponse struct {
		Items []searchResult `json:"items"`
	}
	decodeResponse(t, platformSearch, &platformResponse)
	if len(platformResponse.Items) == 0 {
		t.Fatal("expected workspace event logs for platform search")
	}
	if platformResponse.Items[0].WorkspacePath == "" || platformResponse.Items[0].SessionID == "" {
		t.Fatalf("expected search result workspace context, got %#v", platformResponse.Items[0])
	}
}

func TestWorkspaceSessionDetailSupportsAnchorMessageWindow(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	messages := make([]models.WorkspaceMessage, 0, 30)
	for index := 1; index <= 30; index++ {
		messages = append(messages, models.WorkspaceMessage{
			ID:         index,
			SessionID:  1,
			TenantID:   1,
			InstanceID: 100,
			Role:       "assistant",
			Status:     "delivered",
			Content:    fmt.Sprintf("message-%02d", index),
			TraceID:    fmt.Sprintf("trace-%02d", index/5),
			CreatedAt:  fmt.Sprintf("2026-04-09T03:%02d:00Z", index),
			UpdatedAt:  fmt.Sprintf("2026-04-09T03:%02d:00Z", index),
		})
	}
	router.data.WorkspaceMessages = messages

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/1?messageLimit=6&anchorMessageId=12", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected anchored workspace detail status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Messages []struct {
			ID int `json:"id"`
		} `json:"messages"`
		MessagesHasMore bool `json:"messagesHasMore"`
	}
	decodeResponse(t, recorder, &response)
	if len(response.Messages) != 6 {
		t.Fatalf("expected anchored message window size 6, got %#v", response.Messages)
	}
	foundAnchor := false
	for _, item := range response.Messages {
		if item.ID == 12 {
			foundAnchor = true
			break
		}
	}
	if !foundAnchor {
		t.Fatalf("expected anchored message 12 in response window, got %#v", response.Messages)
	}
	if !response.MessagesHasMore {
		t.Fatal("expected anchored response to indicate more messages exist")
	}
}
