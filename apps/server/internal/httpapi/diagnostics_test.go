package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestAdminDiagnosticSessionLifecycle(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	adminCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})

	create := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/instances/100/diagnostic-sessions", map[string]any{
		"podName":    "openclaw-gateway-prod-7b79d6c767-rk8cz",
		"accessMode": "readonly",
		"reason":     "排查 CPU 波动",
	}, adminCookie)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected diagnostic session create status 201, got %d: %s", create.Code, create.Body.String())
	}

	var createResponse struct {
		Session struct {
			ID         int    `json:"id"`
			Status     string `json:"status"`
			AccessMode string `json:"accessMode"`
		} `json:"session"`
	}
	decodeResponse(t, create, &createResponse)
	if createResponse.Session.ID == 0 || createResponse.Session.Status != diagnosticSessionStatusActive || createResponse.Session.AccessMode != diagnosticAccessModeReadonly {
		t.Fatalf("unexpected create response: %#v", createResponse.Session)
	}

	execute := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/diagnostic-sessions/"+strconv.Itoa(createResponse.Session.ID)+"/commands", map[string]any{
		"commandKey": "process.list",
	}, adminCookie)
	if execute.Code != http.StatusOK {
		t.Fatalf("expected diagnostic command status 200, got %d: %s", execute.Code, execute.Body.String())
	}

	var executeResponse struct {
		Command struct {
			Status      string `json:"status"`
			CommandText string `json:"commandText"`
		} `json:"command"`
	}
	decodeResponse(t, execute, &executeResponse)
	if executeResponse.Command.Status != diagnosticCommandStatusSucceeded || executeResponse.Command.CommandText != "ps aux" {
		t.Fatalf("unexpected command response: %#v", executeResponse.Command)
	}

	detail := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/admin/diagnostic-sessions/"+strconv.Itoa(createResponse.Session.ID), nil, adminCookie)
	if detail.Code != http.StatusOK {
		t.Fatalf("expected diagnostic detail status 200, got %d: %s", detail.Code, detail.Body.String())
	}
	if !strings.Contains(detail.Body.String(), "ps aux") {
		t.Fatalf("expected diagnostic detail to include command record, got %s", detail.Body.String())
	}

	closeSession := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/diagnostic-sessions/"+strconv.Itoa(createResponse.Session.ID)+"/close", map[string]any{
		"reason": "operator_closed",
	}, adminCookie)
	if closeSession.Code != http.StatusOK {
		t.Fatalf("expected diagnostic close status 200, got %d: %s", closeSession.Code, closeSession.Body.String())
	}

	if len(router.data.DiagnosticSessions) == 0 || router.data.DiagnosticSessions[0].Status != diagnosticSessionStatusClosed {
		t.Fatalf("expected closed diagnostic session persisted, got %#v", router.data.DiagnosticSessions)
	}
	if len(router.data.DiagnosticCommandRecords) == 0 || router.data.DiagnosticCommandRecords[len(router.data.DiagnosticCommandRecords)-1].CommandText != "ps aux" {
		t.Fatalf("expected diagnostic command record persisted, got %#v", router.data.DiagnosticCommandRecords)
	}
	if store.saveDataCalls < 3 {
		t.Fatalf("expected diagnostic lifecycle to persist mutations, got %d saves", store.saveDataCalls)
	}
}

func TestAdminDiagnosticWhitelistRequiresApproval(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}

	adminCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})

	recorder := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/instances/100/diagnostic-sessions", map[string]any{
		"accessMode": "whitelist",
	}, adminCookie)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected whitelist approval status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestAdminDiagnosticWhitelistBlocksUnknownCommandAndRecordsIt(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	adminCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})

	create := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/instances/100/diagnostic-sessions", map[string]any{
		"accessMode":     "whitelist",
		"approvalTicket": "APR-20260409-004",
		"approvedBy":     "platform-ops",
	}, adminCookie)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected whitelist session create status 201, got %d: %s", create.Code, create.Body.String())
	}

	var createResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, create, &createResponse)

	execute := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/diagnostic-sessions/"+strconv.Itoa(createResponse.Session.ID)+"/commands", map[string]any{
		"commandText": "rm -rf /",
	}, adminCookie)
	if execute.Code != http.StatusForbidden {
		t.Fatalf("expected blocked whitelist command status 403, got %d: %s", execute.Code, execute.Body.String())
	}

	if len(router.data.DiagnosticCommandRecords) == 0 {
		t.Fatal("expected blocked diagnostic command recorded")
	}
	record := router.data.DiagnosticCommandRecords[len(router.data.DiagnosticCommandRecords)-1]
	if record.Status != diagnosticCommandStatusBlocked || !strings.Contains(record.ErrorOutput, "approved whitelist") {
		t.Fatalf("unexpected blocked command record: %#v", record)
	}
}

func TestAdminDiagnosticSessionTenantIsolation(t *testing.T) {
	router := &Router{
		data:    mockdata.Seed(),
		runtime: newTestRuntimeAdapter(),
	}

	platformCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Platform Admin",
		Email:         "platform@openclaw.local",
		Role:          "platform_admin",
	})

	create := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/instances/200/diagnostic-sessions", map[string]any{
		"accessMode": "readonly",
	}, platformCookie)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected platform diagnostic create status 201, got %d: %s", create.Code, create.Body.String())
	}

	var createResponse struct {
		Session struct {
			ID int `json:"id"`
		} `json:"session"`
	}
	decodeResponse(t, create, &createResponse)

	tenantCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})

	recorder := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/admin/diagnostic-sessions/"+strconv.Itoa(createResponse.Session.ID), nil, tenantCookie)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant diagnostic detail status 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
