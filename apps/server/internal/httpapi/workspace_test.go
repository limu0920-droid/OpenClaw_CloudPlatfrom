package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"openclaw/platformapi/internal/models"
)

func TestWorkspaceSessionLifecycle(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "龙虾发布对话",
		"workspaceUrl": "https://acme.example.com/workspace",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)
	if sessionResponse.Session.ID == 0 {
		t.Fatal("expected session id")
	}

	createArtifact := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/artifacts", map[string]any{
		"title":     "发布演示文稿",
		"kind":      "pptx",
		"sourceUrl": "https://files.acme.example.com/deck.pptx",
	})
	if createArtifact.Code != http.StatusCreated {
		t.Fatalf("expected artifact create status 201, got %d: %s", createArtifact.Code, createArtifact.Body.String())
	}

	getDetail := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID), nil)
	if getDetail.Code != http.StatusOK {
		t.Fatalf("expected session detail status 200, got %d: %s", getDetail.Code, getDetail.Body.String())
	}

	var detail struct {
		Session struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"session"`
		Artifacts []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
			Kind  string `json:"kind"`
		} `json:"artifacts"`
	}
	decodeResponse(t, getDetail, &detail)

	if detail.Session.Title != "龙虾发布对话" {
		t.Fatalf("expected session title persisted, got %q", detail.Session.Title)
	}
	if len(detail.Artifacts) == 0 {
		t.Fatal("expected workspace artifacts")
	}
	if detail.Artifacts[0].Kind != "pptx" {
		t.Fatalf("expected artifact kind pptx, got %q", detail.Artifacts[0].Kind)
	}
	if store.saveDataCalls < 2 {
		t.Fatalf("expected SaveData called for session and artifact, got %d", store.saveDataCalls)
	}
	if len(store.lastSavedData.WorkspaceSessions) == 0 {
		t.Fatal("expected workspace sessions included in persisted snapshot")
	}
	if len(store.lastSavedData.WorkspaceArtifacts) == 0 {
		t.Fatal("expected workspace artifacts included in persisted snapshot")
	}
}

func TestWorkspaceMessageDispatchLifecycle(t *testing.T) {
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/bridge/messages" {
			t.Fatalf("unexpected bridge path: %s", req.URL.Path)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true,"artifacts":[{"title":"龙虾生成的PPT","kind":"pptx","sourceUrl":"/generated/generated-deck.pptx","previewUrl":"/generated/generated-deck-preview.pdf"}]}`))
	}))
	defer bridge.Close()

	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			CoreStore:           store,
			WorkspaceBridgePath: "/bridge/messages",
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "桥接测试会话",
		"workspaceUrl": bridge.URL + "/workspace",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)

	createMessage := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/messages", map[string]any{
		"role":     "user",
		"content":  "帮我生成一份路演 PPT",
		"dispatch": true,
	})
	if createMessage.Code != http.StatusCreated {
		t.Fatalf("expected message create status 201, got %d: %s", createMessage.Code, createMessage.Body.String())
	}

	var response struct {
		Message struct {
			Status string `json:"status"`
		} `json:"message"`
		Dispatch struct {
			OK bool `json:"ok"`
		} `json:"dispatch"`
	}
	decodeResponse(t, createMessage, &response)

	if !response.Dispatch.OK {
		t.Fatal("expected dispatch ok")
	}
	if response.Message.Status != "sent" {
		t.Fatalf("expected message status sent, got %q", response.Message.Status)
	}
	if len(router.data.WorkspaceArtifacts) == 0 {
		t.Fatal("expected artifact created from bridge response")
	}
	if router.data.WorkspaceArtifacts[0].SourceURL != bridge.URL+"/generated/generated-deck.pptx" {
		t.Fatalf("unexpected artifact source url: %#v", router.data.WorkspaceArtifacts[0])
	}
	if router.data.WorkspaceArtifacts[0].PreviewURL != bridge.URL+"/generated/generated-deck-preview.pdf" {
		t.Fatalf("unexpected artifact preview url: %#v", router.data.WorkspaceArtifacts[0])
	}
}

func TestWorkspaceMessageDispatchPersistsAssistantReply(t *testing.T) {
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{
  "accepted": true,
  "assistant": {
    "role": "assistant",
    "status": "delivered",
    "content": "已经整理好路演大纲，接下来我会补齐 PPT 与官网结构。"
  }
}`))
	}))
	defer bridge.Close()

	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			CoreStore:           store,
			WorkspaceBridgePath: "/api/v1/platform/workspace/messages",
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "回复落库测试",
		"workspaceUrl": bridge.URL + "/workspace",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)

	createMessage := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/messages", map[string]any{
		"content":  "请先给我一个路演大纲。",
		"dispatch": true,
	})
	if createMessage.Code != http.StatusCreated {
		t.Fatalf("expected message create status 201, got %d: %s", createMessage.Code, createMessage.Body.String())
	}

	var response struct {
		Reply *struct {
			Role    string `json:"role"`
			Status  string `json:"status"`
			Content string `json:"content"`
		} `json:"reply"`
	}
	decodeResponse(t, createMessage, &response)

	if response.Reply == nil {
		t.Fatal("expected persisted assistant reply")
	}
	if response.Reply.Role != "assistant" || response.Reply.Status != "delivered" {
		t.Fatalf("unexpected reply metadata: %#v", response.Reply)
	}
	if !strings.Contains(response.Reply.Content, "路演大纲") {
		t.Fatalf("unexpected reply content: %#v", response.Reply)
	}
	if len(router.data.WorkspaceMessages) < 2 {
		t.Fatalf("expected reply message appended, got %#v", router.data.WorkspaceMessages)
	}
}

func TestWorkspaceMessageStreamEndpoint(t *testing.T) {
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{
  "accepted": true,
  "assistant": {
    "role": "assistant",
    "status": "delivered",
    "content": "第一部分先讲市场背景。第二部分再讲产品方案。"
  },
  "artifacts": [
    {
      "title": "路演目录",
      "kind": "docx",
      "sourceUrl": "https://files.acme.example.com/pitch-outline.docx"
    }
  ]
}`))
	}))
	defer bridge.Close()

	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			CoreStore:           store,
			WorkspaceBridgePath: "/api/v1/platform/workspace/messages",
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "流式测试会话",
		"workspaceUrl": bridge.URL + "/workspace",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)

	stream := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/messages/stream", map[string]any{
		"content": "请给我一个两段式路演结构。",
	})
	if stream.Code != http.StatusOK {
		t.Fatalf("expected stream status 200, got %d: %s", stream.Code, stream.Body.String())
	}
	if !strings.Contains(stream.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("expected event stream content type, got %q", stream.Header().Get("Content-Type"))
	}
	body := stream.Body.String()
	for _, marker := range []string{"event: message", "event: status", "event: chunk", "event: reply", "event: artifact", "event: done"} {
		if !strings.Contains(body, marker) {
			t.Fatalf("expected stream body to contain %q, got %s", marker, body)
		}
	}
	if len(router.data.WorkspaceArtifacts) == 0 {
		t.Fatal("expected stream dispatch artifact persisted")
	}
}

func TestWorkspaceSessionDetailIncludesEvents(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/1", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected detail status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Events []struct {
			ID        int    `json:"id"`
			EventType string `json:"eventType"`
		} `json:"events"`
	}
	decodeResponse(t, recorder, &response)

	if len(response.Events) == 0 {
		t.Fatal("expected workspace events in detail payload")
	}
	if response.Events[0].EventType == "" {
		t.Fatalf("expected event type populated, got %#v", response.Events[0])
	}
}

func TestWorkspaceDispatchConsumesBridgeStream(t *testing.T) {
	mux := http.NewServeMux()
	bridge := httptest.NewServer(mux)
	defer bridge.Close()

	mux.HandleFunc("/bridge/messages", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(fmt.Sprintf(`{
  "accepted": true,
  "traceId": "trace-stream-1",
  "stream": { "url": "%s/bridge/stream" }
}`, bridge.URL)))
	})
	mux.HandleFunc("/bridge/stream", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if flusher, ok := w.(http.Flusher); ok {
			_, _ = fmt.Fprint(w, "event: reasoning.delta\n")
			_, _ = fmt.Fprint(w, "data: {\"reasoning\":\"先整理路演结构。\"}\n\n")
			_, _ = fmt.Fprint(w, "event: tool.started\n")
			_, _ = fmt.Fprint(w, "data: {\"toolCallId\":\"tool-1\",\"toolName\":\"deck-builder\",\"status\":\"started\",\"detail\":\"生成路演大纲\"}\n\n")
			_, _ = fmt.Fprint(w, "event: message.delta\n")
			_, _ = fmt.Fprint(w, "data: {\"delta\":\"第一部分先讲市场背景。\",\"content\":\"第一部分先讲市场背景。\",\"status\":\"streaming\"}\n\n")
			_, _ = fmt.Fprintf(w, "event: artifact.created\n")
			_, _ = fmt.Fprintf(w, "data: {\"artifact\":{\"id\":\"art-1\",\"title\":\"路演目录\",\"kind\":\"docx\",\"sourceUrl\":\"%s/files/pitch-outline.docx\"}}\n\n", bridge.URL)
			_, _ = fmt.Fprint(w, "event: message.completed\n")
			_, _ = fmt.Fprint(w, "data: {\"message\":{\"role\":\"assistant\",\"status\":\"delivered\",\"content\":\"第一部分先讲市场背景。第二部分再讲产品方案。\"}}\n\n")
			_, _ = fmt.Fprint(w, "event: done\n")
			_, _ = fmt.Fprint(w, "data: {}\n\n")
			flusher.Flush()
		}
	})

	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			CoreStore:           store,
			WorkspaceBridgePath: "/bridge/messages",
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "桥接流测试",
		"workspaceUrl": bridge.URL + "/workspace",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)

	createMessage := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/messages", map[string]any{
		"content":  "请给我一个两段式路演结构。",
		"dispatch": true,
	})
	if createMessage.Code != http.StatusCreated {
		t.Fatalf("expected message create status 201, got %d: %s", createMessage.Code, createMessage.Body.String())
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(router.data.WorkspaceArtifacts) > 0 {
			lastMessage := router.data.WorkspaceMessages[len(router.data.WorkspaceMessages)-1]
			if strings.Contains(lastMessage.Content, "第二部分再讲产品方案") && lastMessage.Status == "delivered" {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	lastMessage := router.data.WorkspaceMessages[len(router.data.WorkspaceMessages)-1]
	if !strings.Contains(lastMessage.Content, "第二部分再讲产品方案") {
		t.Fatalf("expected assistant message completed from stream, got %#v", lastMessage)
	}
	if lastMessage.Status != "delivered" {
		t.Fatalf("expected assistant message delivered, got %#v", lastMessage)
	}
	if len(router.data.WorkspaceArtifacts) == 0 {
		t.Fatal("expected artifact created from stream event")
	}

	eventTypes := make([]string, 0, len(router.data.WorkspaceMessageEvents))
	for _, item := range router.data.WorkspaceMessageEvents {
		if item.SessionID == sessionResponse.Session.ID {
			eventTypes = append(eventTypes, item.EventType)
		}
	}
	for _, expected := range []string{"reasoning.delta", "tool.started", "message.delta", "artifact.created", "message.completed"} {
		if !containsString(eventTypes, expected) {
			t.Fatalf("expected event %q in %#v", expected, eventTypes)
		}
	}
}

func TestWorkspaceAdminSessionIndexSearch(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}
	router.data.WorkspaceSessions = append([]models.WorkspaceSession{{
		ID:           20,
		SessionNo:    "WS-NOVA-020",
		TenantID:     2,
		InstanceID:   200,
		Title:        "Nova 上线检查",
		Status:       "active",
		WorkspaceURL: "https://nova.example.com/workspace",
		CreatedAt:    nowRFC3339(),
		UpdatedAt:    nowRFC3339(),
	}}, router.data.WorkspaceSessions...)
	router.data.WorkspaceMessages = append(router.data.WorkspaceMessages, models.WorkspaceMessage{
		ID:         20,
		SessionID:  20,
		TenantID:   2,
		InstanceID: 200,
		Role:       "user",
		Status:     "recorded",
		Content:    "排查 Nova 生产环境的上线问题",
		CreatedAt:  nowRFC3339(),
		UpdatedAt:  nowRFC3339(),
	})

	recorder := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/admin/workspace/sessions?q=nova&tenantId=2", nil,
		workspaceAuthCookie(t, router, authSessionState{
			Authenticated: true,
			UserID:        1,
			TenantID:      1,
			Name:          "Platform Admin",
			Email:         "platform@openclaw.local",
			Role:          "platform_admin",
		}),
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected session index status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Items []struct {
			SessionNo          string `json:"sessionNo"`
			TenantID           int    `json:"tenantId"`
			TenantName         string `json:"tenantName"`
			InstanceName       string `json:"instanceName"`
			LastMessagePreview string `json:"lastMessagePreview"`
		} `json:"items"`
	}
	decodeResponse(t, recorder, &response)

	if len(response.Items) != 1 {
		t.Fatalf("expected one filtered session, got %#v", response.Items)
	}
	if response.Items[0].TenantID != 2 || response.Items[0].TenantName == "" || response.Items[0].InstanceName == "" {
		t.Fatalf("unexpected session summary: %#v", response.Items[0])
	}
	if !strings.Contains(response.Items[0].LastMessagePreview, "Nova") && !strings.Contains(response.Items[0].LastMessagePreview, "nova") {
		t.Fatalf("expected last message preview included, got %#v", response.Items[0])
	}
}

func TestWorkspaceBridgeHealth(t *testing.T) {
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/bridge/health" {
			t.Fatalf("unexpected bridge health path: %s", req.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer bridge.Close()

	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			WorkspaceBridgeHealthPath: "/bridge/health",
		},
		runtime: newTestRuntimeAdapter(),
	}
	router.data.WorkspaceSessions = append([]models.WorkspaceSession{{
		ID:           99,
		SessionNo:    "WS-TEST",
		TenantID:     1,
		InstanceID:   100,
		Title:        "bridge health",
		Status:       "active",
		WorkspaceURL: bridge.URL + "/workspace",
		CreatedAt:    nowRFC3339(),
		UpdatedAt:    nowRFC3339(),
	}}, router.data.WorkspaceSessions...)

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/instances/100/workspace/bridge-health", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected bridge health status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"ok\": true") && !strings.Contains(recorder.Body.String(), "\"ok\":true") {
		t.Fatalf("expected ok response, got %s", recorder.Body.String())
	}
}

func TestWorkspaceArtifactPreviewOfficeFallback(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/2/preview", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected preview descriptor status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Preview struct {
			Available     bool   `json:"available"`
			Mode          string `json:"mode"`
			FailureReason string `json:"failureReason"`
			DownloadURL   string `json:"downloadUrl"`
		} `json:"preview"`
	}
	decodeResponse(t, recorder, &response)

	if response.Preview.Available {
		t.Fatal("expected office preview to fall back to download")
	}
	if response.Preview.Mode != "download" {
		t.Fatalf("expected mode download, got %q", response.Preview.Mode)
	}
	if !strings.Contains(response.Preview.FailureReason, "PDF/HTML") {
		t.Fatalf("expected Office fallback reason, got %q", response.Preview.FailureReason)
	}
	if response.Preview.DownloadURL == "" {
		t.Fatal("expected download url for office fallback")
	}
}

func TestWorkspaceArtifactPreviewOfficePDFRendition(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
		config: ExternalConfig{
			ArtifactPreviewAllowPrivateIP: true,
		},
	}
	for index := range router.data.WorkspaceArtifacts {
		if router.data.WorkspaceArtifacts[index].ID == 2 {
			router.data.WorkspaceArtifacts[index].PreviewURL = "https://files.acme.example.com/artifacts/demo-deck-preview.pdf"
		}
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/2/preview", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected preview descriptor status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Preview struct {
			Available   bool   `json:"available"`
			Mode        string `json:"mode"`
			Strategy    string `json:"strategy"`
			PreviewURL  string `json:"previewUrl"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"preview"`
	}
	decodeResponse(t, recorder, &response)

	if !response.Preview.Available {
		t.Fatal("expected office preview to be available when previewUrl is a PDF")
	}
	if response.Preview.Mode != "pdf" {
		t.Fatalf("expected mode pdf, got %q", response.Preview.Mode)
	}
	if response.Preview.Strategy != "office-pdf-rendition" {
		t.Fatalf("unexpected strategy: %q", response.Preview.Strategy)
	}
	if response.Preview.PreviewURL == "" || response.Preview.DownloadURL == "" {
		t.Fatalf("expected preview and download urls, got %#v", response.Preview)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func TestWorkspaceArtifactPreviewHTMLProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><head><title>Artifact</title></head><body><script>window.previewReady=true</script></body></html>`))
	}))
	defer upstream.Close()

	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			ArtifactPreviewAllowPrivateIP: true,
			ArtifactPreviewAllowedHosts:   "127.0.0.1",
		},
		runtime: newTestRuntimeAdapter(),
	}
	router.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{{
		ID:         99,
		SessionID:  1,
		TenantID:   1,
		InstanceID: 100,
		Title:      "HTML Preview",
		Kind:       "web",
		SourceURL:  upstream.URL + "/artifact.html",
		CreatedAt:  nowRFC3339(),
		UpdatedAt:  nowRFC3339(),
	}}, router.data.WorkspaceArtifacts...)

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/99/preview-content", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected preview content status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Header().Get("Content-Security-Policy"), "sandbox") {
		t.Fatalf("expected sandbox csp, got %q", recorder.Header().Get("Content-Security-Policy"))
	}
	if !strings.Contains(recorder.Body.String(), `<base href="`+upstream.URL+`/artifact.html">`) {
		t.Fatalf("expected base href injected, got %s", recorder.Body.String())
	}
}

func TestWorkspaceArtifactPreviewRejectsUntrustedHost(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}
	router.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{{
		ID:         100,
		SessionID:  1,
		TenantID:   1,
		InstanceID: 100,
		Title:      "Private HTML",
		Kind:       "web",
		SourceURL:  "http://127.0.0.1/private.html",
		CreatedAt:  nowRFC3339(),
		UpdatedAt:  nowRFC3339(),
	}}, router.data.WorkspaceArtifacts...)

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/100/preview", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected preview descriptor status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Preview struct {
			Available     bool   `json:"available"`
			Mode          string `json:"mode"`
			PreviewURL    string `json:"previewUrl"`
			DownloadURL   string `json:"downloadUrl"`
			FailureReason string `json:"failureReason"`
		} `json:"preview"`
	}
	decodeResponse(t, recorder, &response)

	if response.Preview.Available {
		t.Fatal("expected private host preview to be rejected")
	}
	if response.Preview.PreviewURL != "" {
		t.Fatalf("expected no preview url, got %q", response.Preview.PreviewURL)
	}
	if response.Preview.DownloadURL != "" {
		t.Fatalf("expected no download url for rejected private host, got %q", response.Preview.DownloadURL)
	}
	if !strings.Contains(response.Preview.FailureReason, "不受信任") {
		t.Fatalf("expected untrusted host failure reason, got %q", response.Preview.FailureReason)
	}
}

func TestWorkspaceArtifactCreateNormalizesRelativeURLs(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "相对地址会话",
		"workspaceUrl": "https://workspace.acme.example.com/app/session",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)

	createArtifact := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/artifacts", map[string]any{
		"title":      "相对地址产物",
		"kind":       "pptx",
		"sourceUrl":  "/exports/demo-deck.pptx",
		"previewUrl": "/exports/demo-deck-preview.html",
	})
	if createArtifact.Code != http.StatusCreated {
		t.Fatalf("expected artifact create status 201, got %d: %s", createArtifact.Code, createArtifact.Body.String())
	}

	var response struct {
		Artifact struct {
			SourceURL  string `json:"sourceUrl"`
			PreviewURL string `json:"previewUrl"`
		} `json:"artifact"`
	}
	decodeResponse(t, createArtifact, &response)

	if response.Artifact.SourceURL != "https://workspace.acme.example.com/exports/demo-deck.pptx" {
		t.Fatalf("unexpected normalized source url: %q", response.Artifact.SourceURL)
	}
	if response.Artifact.PreviewURL != "https://workspace.acme.example.com/exports/demo-deck-preview.html" {
		t.Fatalf("unexpected normalized preview url: %q", response.Artifact.PreviewURL)
	}
}

func TestWorkspaceArtifactCreateRejectsUntrustedSource(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "不可信域名会话",
		"workspaceUrl": "https://workspace.acme.example.com/app/session",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)

	createArtifact := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/artifacts", map[string]any{
		"title":     "危险外链",
		"kind":      "pdf",
		"sourceUrl": "https://evil.other-domain.example.net/report.pdf",
	})
	if createArtifact.Code != http.StatusForbidden {
		t.Fatalf("expected artifact create status 403, got %d: %s", createArtifact.Code, createArtifact.Body.String())
	}
}

func TestWorkspaceArtifactPreviewOfficeHTMLRendition(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}
	for index := range router.data.WorkspaceArtifacts {
		if router.data.WorkspaceArtifacts[index].ID == 2 {
			router.data.WorkspaceArtifacts[index].PreviewURL = "https://files.acme.example.com/artifacts/demo-deck-preview.html"
		}
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/2/preview", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected preview descriptor status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Preview struct {
			Available   bool   `json:"available"`
			Mode        string `json:"mode"`
			Strategy    string `json:"strategy"`
			Sandboxed   bool   `json:"sandboxed"`
			PreviewURL  string `json:"previewUrl"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"preview"`
	}
	decodeResponse(t, recorder, &response)

	if !response.Preview.Available {
		t.Fatal("expected office preview to be available when previewUrl is HTML")
	}
	if response.Preview.Mode != "html" {
		t.Fatalf("expected mode html, got %q", response.Preview.Mode)
	}
	if response.Preview.Strategy != "office-html-rendition" {
		t.Fatalf("unexpected strategy: %q", response.Preview.Strategy)
	}
	if !response.Preview.Sandboxed {
		t.Fatal("expected html rendition to be sandboxed")
	}
	if response.Preview.PreviewURL == "" || response.Preview.DownloadURL == "" {
		t.Fatalf("expected preview and download urls, got %#v", response.Preview)
	}
}

func TestWorkspaceRequiresAuthInStrictMode(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{StrictMode: true},
		runtime: newTestRuntimeAdapter(),
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/instances/100/workspace/sessions", nil)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected strict workspace auth status 401, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestWorkspacePortalTenantIsolationOnSessionDetail(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}
	router.data.WorkspaceSessions = append([]models.WorkspaceSession{{
		ID:           20,
		SessionNo:    "WS-NOVA-020",
		TenantID:     2,
		InstanceID:   200,
		Title:        "Nova 机密会话",
		Status:       "active",
		WorkspaceURL: "https://nova.example.com/workspace",
		CreatedAt:    nowRFC3339(),
		UpdatedAt:    nowRFC3339(),
	}}, router.data.WorkspaceSessions...)

	recorder := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/20", nil,
		workspaceAuthCookie(t, router, authSessionState{
			Authenticated: true,
			UserID:        1,
			TenantID:      1,
			Name:          "Acme Admin",
			Email:         "owner@acme.example.com",
			Role:          "tenant_admin",
		}),
	)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant session detail status 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestWorkspaceAdminScopeRespectsTenantAndPlatformAdmin(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}
	router.data.WorkspaceSessions = append([]models.WorkspaceSession{{
		ID:           21,
		SessionNo:    "WS-NOVA-021",
		TenantID:     2,
		InstanceID:   200,
		Title:        "Nova Admin Session",
		Status:       "active",
		WorkspaceURL: "https://nova.example.com/workspace",
		CreatedAt:    nowRFC3339(),
		UpdatedAt:    nowRFC3339(),
	}}, router.data.WorkspaceSessions...)

	tenantAdmin := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})
	recorder := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/admin/workspace/sessions/21", nil, tenantAdmin)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected tenant admin cross-tenant status 404, got %d: %s", recorder.Code, recorder.Body.String())
	}

	platformAdmin := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Platform Admin",
		Email:         "platform@openclaw.local",
		Role:          "platform_admin",
	})
	recorder = performRequestWithCookies(t, router, http.MethodGet, "/api/v1/admin/workspace/sessions/21", nil, platformAdmin)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected platform admin cross-tenant status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestWorkspaceArtifactDownloadTenantIsolationAndAudit(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("artifact-bytes"))
	}))
	defer upstream.Close()

	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			CoreStore:                     store,
			ArtifactPreviewAllowPrivateIP: true,
			ArtifactPreviewAllowedHosts:   "127.0.0.1",
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}
	router.data.WorkspaceSessions = append([]models.WorkspaceSession{{
		ID:           30,
		SessionNo:    "WS-ACME-030",
		TenantID:     1,
		InstanceID:   100,
		Title:        "Acme Download Session",
		Status:       "active",
		WorkspaceURL: upstream.URL + "/workspace",
		CreatedAt:    nowRFC3339(),
		UpdatedAt:    nowRFC3339(),
	}}, router.data.WorkspaceSessions...)
	router.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{{
		ID:         30,
		SessionID:  30,
		TenantID:   1,
		InstanceID: 100,
		Title:      "机密 PDF",
		Kind:       "pdf",
		SourceURL:  upstream.URL + "/artifact.pdf",
		CreatedAt:  nowRFC3339(),
		UpdatedAt:  nowRFC3339(),
	}}, router.data.WorkspaceArtifacts...)

	novaCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        2,
		TenantID:      2,
		Name:          "Nova Admin",
		Email:         "carol@nova.example.com",
		Role:          "tenant_admin",
	})
	recorder := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/30/download", nil, novaCookie)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant artifact download status 404, got %d: %s", recorder.Code, recorder.Body.String())
	}

	acmeCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})
	recorder = performRequestWithCookies(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/30/download", nil, acmeCookie)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected artifact download status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != "artifact-bytes" {
		t.Fatalf("unexpected artifact download body: %s", recorder.Body.String())
	}
	if store.saveDataCalls == 0 {
		t.Fatal("expected artifact download audit to persist")
	}
	if len(router.data.Audits) == 0 {
		t.Fatal("expected workspace audit event recorded")
	}
	if router.data.Audits[0].Action != "workspace.artifact.download" || router.data.Audits[0].Result != "success" {
		t.Fatalf("unexpected latest audit: %#v", router.data.Audits[0])
	}
}

func TestWorkspaceBridgeReportPersistsAsyncPayload(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			CoreStore:            store,
			WorkspaceBridgeToken: "bridge-secret",
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/platform/workspace/report", bytes.NewReader([]byte(`{
  "protocolVersion": "openclaw-lobster-bridge/v2",
  "requestId": "report-1",
  "sessionNo": "WS-20260405-001",
  "traceId": "trace-report-1",
  "messages": [
    {
      "id": "lob-msg-report-1",
      "role": "assistant",
      "status": "delivered",
      "content": "回调补发了一条异步回复。",
      "createdAt": "2026-04-05T10:10:00Z"
    }
  ],
  "artifacts": [
    {
      "id": "lob-art-report-1",
      "messageId": "lob-msg-report-1",
      "title": "回调补发文档",
      "kind": "docx",
      "sourceUrl": "https://files.acme.example.com/report-follow-up.docx",
      "previewUrl": "https://files.acme.example.com/report-follow-up.docx",
      "createdAt": "2026-04-05T10:10:00Z"
    }
  ]
}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform-Bridge-Token", "bridge-secret")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected report status 202, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if len(router.data.WorkspaceMessages) < 3 {
		t.Fatalf("expected async report message appended, got %#v", router.data.WorkspaceMessages)
	}
	if router.data.WorkspaceMessages[len(router.data.WorkspaceMessages)-1].ExternalID != "lob-msg-report-1" {
		t.Fatalf("expected external id persisted, got %#v", router.data.WorkspaceMessages[len(router.data.WorkspaceMessages)-1])
	}
	if router.data.WorkspaceArtifacts[0].ExternalID != "lob-art-report-1" || router.data.WorkspaceArtifacts[0].Origin != workspaceArtifactOriginBridgeReport {
		t.Fatalf("expected report artifact persisted, got %#v", router.data.WorkspaceArtifacts[0])
	}
	if store.saveDataCalls == 0 {
		t.Fatal("expected report payload persisted")
	}
}

func TestWorkspaceSessionDetailSyncsBridgeHistory(t *testing.T) {
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if !strings.Contains(req.URL.Path, "/api/v1/platform/workspace/sessions/WS-HISTORY-001/history") {
			t.Fatalf("unexpected history path: %s", req.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
  "protocolVersion": "openclaw-lobster-bridge/v2",
  "requestId": "history-1",
  "messages": [
    {
      "id": "lob-msg-history-1",
      "role": "assistant",
      "status": "delivered",
      "content": "这里是拉历史补回的一条消息。",
      "createdAt": "2026-04-05T11:01:00Z"
    }
  ],
  "artifacts": [
    {
      "id": "lob-art-history-1",
      "messageId": "lob-msg-history-1",
      "title": "历史回补网页",
      "kind": "web",
      "sourceUrl": "https://acme.example.com/history-page",
      "previewUrl": "https://acme.example.com/history-page",
      "createdAt": "2026-04-05T11:01:10Z"
    }
  ],
  "syncedAt": "2026-04-05T11:01:15Z"
}`))
	}))
	defer bridge.Close()

	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			CoreStore:                  store,
			WorkspaceBridgeHistorySync: true,
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}
	router.data.WorkspaceSessions = append([]models.WorkspaceSession{{
		ID:              40,
		SessionNo:       "WS-HISTORY-001",
		TenantID:        1,
		InstanceID:      100,
		Title:           "历史同步测试",
		Status:          "active",
		WorkspaceURL:    bridge.URL + "/workspace",
		ProtocolVersion: workspaceBridgeProtocolVersion,
		CreatedAt:       nowRFC3339(),
		UpdatedAt:       nowRFC3339(),
	}}, router.data.WorkspaceSessions...)

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/40", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected session detail status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Session struct {
			ProtocolVersion string `json:"protocolVersion"`
			LastSyncedAt    string `json:"lastSyncedAt"`
		} `json:"session"`
		BridgeSync struct {
			OK              bool `json:"ok"`
			MessagesSynced  int  `json:"messagesSynced"`
			ArtifactsSynced int  `json:"artifactsSynced"`
		} `json:"bridgeSync"`
		Messages  []models.WorkspaceMessage  `json:"messages"`
		Artifacts []models.WorkspaceArtifact `json:"artifacts"`
	}
	decodeResponse(t, recorder, &response)

	if !response.BridgeSync.OK || response.BridgeSync.MessagesSynced == 0 || response.BridgeSync.ArtifactsSynced == 0 {
		t.Fatalf("expected bridge history sync to append data, got %#v", response.BridgeSync)
	}
	if response.Session.ProtocolVersion != workspaceBridgeProtocolVersion || response.Session.LastSyncedAt == "" {
		t.Fatalf("expected protocol version and sync timestamp, got %#v", response.Session)
	}
	if len(response.Messages) == 0 || response.Messages[len(response.Messages)-1].ExternalID != "lob-msg-history-1" {
		t.Fatalf("expected history message persisted, got %#v", response.Messages)
	}
	if len(response.Artifacts) == 0 || response.Artifacts[0].ExternalID != "lob-art-history-1" {
		t.Fatalf("expected history artifact persisted, got %#v", response.Artifacts)
	}
}

func TestWorkspaceDispatchRetriesRetryableBridgeErrors(t *testing.T) {
	attempts := 0
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"code":"bridge_unavailable","message":"upstream busy","retryable":true}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true}`))
	}))
	defer bridge.Close()

	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			WorkspaceBridgePath: "/api/v1/platform/workspace/messages",
		},
		runtime: newTestRuntimeAdapter(),
	}

	createSession := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/workspace/sessions", map[string]any{
		"title":        "重试测试",
		"workspaceUrl": bridge.URL + "/workspace",
	})
	if createSession.Code != http.StatusCreated {
		t.Fatalf("expected session create status 201, got %d: %s", createSession.Code, createSession.Body.String())
	}

	var sessionResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, createSession, &sessionResponse)

	createMessage := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/"+strconv.Itoa(sessionResponse.Session.ID)+"/messages", map[string]any{
		"content":  "请在重试后发送成功。",
		"dispatch": true,
	})
	if createMessage.Code != http.StatusCreated {
		t.Fatalf("expected message create status 201, got %d: %s", createMessage.Code, createMessage.Body.String())
	}
	if attempts != 2 {
		t.Fatalf("expected retry to hit bridge twice, got %d attempts", attempts)
	}

	var response struct {
		Message struct {
			Status          string `json:"status"`
			DeliveryAttempt int    `json:"deliveryAttempt"`
		} `json:"message"`
		Dispatch struct {
			OK      bool `json:"ok"`
			Attempt int  `json:"attempt"`
		} `json:"dispatch"`
	}
	decodeResponse(t, createMessage, &response)

	if !response.Dispatch.OK || response.Dispatch.Attempt != 2 {
		t.Fatalf("expected successful retry dispatch, got %#v", response.Dispatch)
	}
	if response.Message.Status != "sent" || response.Message.DeliveryAttempt != 2 {
		t.Fatalf("expected delivery attempt persisted on message, got %#v", response.Message)
	}
}

func TestWorkspaceMessageListPaginatesLatestWindow(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}

	for id := 3; id <= 10; id++ {
		router.data.WorkspaceMessages = append(router.data.WorkspaceMessages, models.WorkspaceMessage{
			ID:         id,
			SessionID:  1,
			TenantID:   1,
			InstanceID: 100,
			Role:       "user",
			Status:     "recorded",
			Content:    fmt.Sprintf("message-%02d", id),
			CreatedAt:  fmt.Sprintf("2026-04-05T10:%02d:00Z", id),
			UpdatedAt:  fmt.Sprintf("2026-04-05T10:%02d:00Z", id),
		})
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/1/messages?limit=3", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected message list status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Items []struct {
			ID int `json:"id"`
		} `json:"items"`
		HasMore bool `json:"hasMore"`
	}
	decodeResponse(t, recorder, &response)
	if len(response.Items) != 3 || response.Items[0].ID != 8 || response.Items[2].ID != 10 || !response.HasMore {
		t.Fatalf("unexpected latest message window: %#v", response)
	}

	recorder = performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/1/messages?limit=3&beforeId=8", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected historical message list status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	decodeResponse(t, recorder, &response)
	if len(response.Items) != 3 || response.Items[0].ID != 5 || response.Items[2].ID != 7 || !response.HasMore {
		t.Fatalf("unexpected historical message window: %#v", response)
	}
}

func TestWorkspaceSessionDetailSupportsRecentWindow(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}

	for id := 3; id <= 8; id++ {
		router.data.WorkspaceMessages = append(router.data.WorkspaceMessages, models.WorkspaceMessage{
			ID:         id,
			SessionID:  1,
			TenantID:   1,
			InstanceID: 100,
			Role:       "assistant",
			Status:     "delivered",
			Content:    fmt.Sprintf("window-message-%02d", id),
			CreatedAt:  fmt.Sprintf("2026-04-05T11:%02d:00Z", id),
			UpdatedAt:  fmt.Sprintf("2026-04-05T11:%02d:00Z", id),
		})
		router.data.WorkspaceMessageEvents = append(router.data.WorkspaceMessageEvents, models.WorkspaceMessageEvent{
			ID:          8 + id,
			SessionID:   1,
			MessageID:   id,
			TenantID:    1,
			InstanceID:  100,
			EventType:   workspaceEventMessageCompleted,
			PayloadJSON: fmt.Sprintf(`{"message":{"id":%d,"sessionId":1,"tenantId":1,"instanceId":100,"role":"assistant","status":"delivered","content":"window-message-%02d","createdAt":"2026-04-05T11:%02d:00Z","updatedAt":"2026-04-05T11:%02d:00Z"}}`, id, id, id, id),
			CreatedAt:   fmt.Sprintf("2026-04-05T11:%02d:00Z", id),
		})
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/sessions/1?messageLimit=2&eventLimit=3", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected session detail status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Messages []struct {
			ID int `json:"id"`
		} `json:"messages"`
		MessagesHasMore bool `json:"messagesHasMore"`
		Events          []struct {
			ID int `json:"id"`
		} `json:"events"`
		EventsHasMore bool `json:"eventsHasMore"`
	}
	decodeResponse(t, recorder, &response)
	if len(response.Messages) != 2 || response.Messages[0].ID != 7 || response.Messages[1].ID != 8 || !response.MessagesHasMore {
		t.Fatalf("unexpected recent messages window: %#v", response)
	}
	if len(response.Events) != 3 || response.Events[0].ID != 14 || response.Events[2].ID != 16 || !response.EventsHasMore {
		t.Fatalf("unexpected recent events window: %#v", response)
	}
}

func TestWorkspaceSessionArchiveLifecycle(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	archive := performRequest(t, router, http.MethodPatch, "/api/v1/portal/workspace/sessions/1/status", map[string]any{
		"status": "archived",
	})
	if archive.Code != http.StatusOK {
		t.Fatalf("expected archive status 200, got %d: %s", archive.Code, archive.Body.String())
	}

	var archiveResponse struct {
		Session struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
		} `json:"session"`
	}
	decodeResponse(t, archive, &archiveResponse)

	if archiveResponse.Session.ID != 1 || archiveResponse.Session.Status != "archived" {
		t.Fatalf("expected archived session summary, got %#v", archiveResponse.Session)
	}
	if router.data.WorkspaceSessions[0].Status != "archived" {
		t.Fatalf("expected session status persisted, got %#v", router.data.WorkspaceSessions[0])
	}

	createMessage := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/1/messages", map[string]any{
		"content": "归档后不应允许继续发送。",
	})
	if createMessage.Code != http.StatusConflict {
		t.Fatalf("expected archived session write blocked with 409, got %d: %s", createMessage.Code, createMessage.Body.String())
	}

	restore := performRequest(t, router, http.MethodPatch, "/api/v1/portal/workspace/sessions/1/status", map[string]any{
		"status": "active",
	})
	if restore.Code != http.StatusOK {
		t.Fatalf("expected restore status 200, got %d: %s", restore.Code, restore.Body.String())
	}
}

func TestWorkspaceMessageRetryDispatch(t *testing.T) {
	bridgeCalls := 0
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bridgeCalls++
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{
  "accepted": true,
  "assistant": {
    "role": "assistant",
    "status": "delivered",
    "content": "已经按原消息重新派发，并补齐新的交付结果。"
  }
}`))
	}))
	defer bridge.Close()

	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			WorkspaceBridgePath: "/api/v1/platform/workspace/messages",
		},
		runtime: newTestRuntimeAdapter(),
	}
	router.data.WorkspaceSessions[0].WorkspaceURL = bridge.URL + "/workspace"

	retry := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/messages/1/retry", nil)
	if retry.Code != http.StatusCreated {
		t.Fatalf("expected retry status 201, got %d: %s", retry.Code, retry.Body.String())
	}
	if bridgeCalls != 1 {
		t.Fatalf("expected retry to dispatch once, got %d", bridgeCalls)
	}

	var response struct {
		Message struct {
			ID      int    `json:"id"`
			Status  string `json:"status"`
			Content string `json:"content"`
		} `json:"message"`
		Dispatch struct {
			OK bool `json:"ok"`
		} `json:"dispatch"`
		Reply *struct {
			Content string `json:"content"`
		} `json:"reply"`
	}
	decodeResponse(t, retry, &response)

	if !response.Dispatch.OK {
		t.Fatalf("expected retry dispatch ok, got %#v", response.Dispatch)
	}
	if response.Message.ID == 1 || response.Message.Status != "sent" {
		t.Fatalf("expected new retried message persisted as sent, got %#v", response.Message)
	}
	if response.Message.Content != "帮我整理一版适合客户演示的官网文案。" {
		t.Fatalf("expected retried content copied from source message, got %#v", response.Message)
	}
	if response.Reply == nil || !strings.Contains(response.Reply.Content, "重新派发") {
		t.Fatalf("expected assistant reply from retry dispatch, got %#v", response.Reply)
	}
	if len(router.data.WorkspaceMessages) != 4 {
		t.Fatalf("expected retried user message plus assistant reply appended, got %#v", router.data.WorkspaceMessages)
	}

	retried, ok := router.findWorkspaceMessage(response.Message.ID)
	if !ok || retried.Status != "sent" {
		t.Fatalf("expected retried message stored with sent status, got %#v", retried)
	}
}

func TestWorkspaceMessageRetryRejectsNonUserMessage(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}

	retry := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/messages/2/retry", nil)
	if retry.Code != http.StatusBadRequest {
		t.Fatalf("expected non-user retry rejected with 400, got %d: %s", retry.Code, retry.Body.String())
	}
}

func performRequestWithCookies(t *testing.T, handler http.Handler, method string, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	switch value := body.(type) {
	case nil:
		payload = nil
	case []byte:
		payload = value
	case string:
		payload = []byte(value)
	default:
		var err error
		payload, err = json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func workspaceAuthCookie(t *testing.T, router *Router, session authSessionState) *http.Cookie {
	t.Helper()

	recorder := httptest.NewRecorder()
	router.writeAuthSession(recorder, session)
	result := recorder.Result()
	defer result.Body.Close()
	cookies := result.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected auth cookie")
	}
	return cookies[0]
}

func TestWorkspaceArtifactCreateArchivesIntoObjectStore(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-demo"))
	}))
	defer source.Close()

	store := &fakeArtifactObjectStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			ArtifactArchiveAllowPrivateURL: true,
			ObjectStorageBucket:            "openclaw-test",
		},
		runtime:       newTestRuntimeAdapter(),
		artifactStore: store,
	}
	router.data.WorkspaceSessions[0].WorkspaceURL = source.URL

	create := performRequest(t, router, http.MethodPost, "/api/v1/portal/workspace/sessions/1/artifacts", map[string]any{
		"title":     "归档测试 PDF",
		"kind":      "pdf",
		"sourceUrl": source.URL + "/artifact.pdf",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d: %s", create.Code, create.Body.String())
	}

	if len(router.data.WorkspaceArtifacts) == 0 {
		t.Fatal("expected archived artifact")
	}
	artifact := router.data.WorkspaceArtifacts[0]
	if artifact.ArchiveStatus != "archived" {
		t.Fatalf("expected archive status archived, got %#v", artifact)
	}
	if artifact.StorageKey == "" || artifact.Filename == "" {
		t.Fatalf("expected storage key and filename, got %#v", artifact)
	}
	if got := store.objects[artifact.StorageKey]; string(got.body) != "%PDF-demo" {
		t.Fatalf("unexpected archived body: %q", string(got.body))
	}
}

func TestWorkspaceArtifactDownloadRecordsAccessLog(t *testing.T) {
	store := &fakeArtifactObjectStore{
		objects: map[string]fakeArtifactObject{
			"artifacts/test/demo.pdf": {
				body:        []byte("%PDF-archived"),
				contentType: "application/pdf",
			},
		},
	}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			ObjectStorageBucket: "openclaw-test",
		},
		runtime:       newTestRuntimeAdapter(),
		artifactStore: store,
	}
	router.data.WorkspaceArtifacts[0].ArchiveStatus = "archived"
	router.data.WorkspaceArtifacts[0].StorageKey = "artifacts/test/demo.pdf"
	router.data.WorkspaceArtifacts[0].Filename = "demo.pdf"
	router.data.WorkspaceArtifacts[0].ContentType = "application/pdf"

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/workspace/artifacts/1/download", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected download status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "%PDF-archived") {
		t.Fatalf("expected archived body, got %s", recorder.Body.String())
	}
	if len(router.data.WorkspaceArtifactLogs) != 1 {
		t.Fatalf("expected one access log, got %d", len(router.data.WorkspaceArtifactLogs))
	}
	if router.data.WorkspaceArtifactLogs[0].Action != "download" {
		t.Fatalf("expected download action, got %#v", router.data.WorkspaceArtifactLogs[0])
	}
}

type fakeArtifactObject struct {
	body        []byte
	contentType string
}

type fakeArtifactObjectStore struct {
	objects map[string]fakeArtifactObject
}

func (f *fakeArtifactObjectStore) Enabled() bool {
	return true
}

func (f *fakeArtifactObjectStore) Ping(ctx context.Context) error {
	return nil
}

func (f *fakeArtifactObjectStore) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if f.objects == nil {
		f.objects = make(map[string]fakeArtifactObject)
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	f.objects[key] = fakeArtifactObject{body: data, contentType: contentType}
	return nil
}

func (f *fakeArtifactObjectStore) Open(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	item, ok := f.objects[key]
	if !ok {
		return nil, "", 0, io.EOF
	}
	return io.NopCloser(strings.NewReader(string(item.body))), item.contentType, int64(len(item.body)), nil
}
