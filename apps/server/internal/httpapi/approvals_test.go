package httpapi

import (
	"net/http"
	"strconv"
	"testing"
)

func TestApprovalLifecycleExecutesConfigPublish(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	adminCookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Acme Admin",
		Email:         "owner@acme.example.com",
		Role:          "tenant_admin",
	})

	create := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/approvals", map[string]any{
		"approvalType": "config_publish",
		"targetType":   "instance",
		"targetId":     100,
		"instanceId":   100,
		"riskLevel":    "high",
		"reason":       "发布新模型配置",
		"metadata": map[string]string{
			"updatedBy":      "owner.acme",
			"model":          "gpt-5-mini",
			"allowedOrigins": "https://acme.example.com,https://console.acme.example.com",
			"backupPolicy":   "daily@03:00 保留 14 天",
		},
	}, adminCookie)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected approval create status 201, got %d: %s", create.Code, create.Body.String())
	}

	var createResponse struct {
		Approval struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
		} `json:"approval"`
	}
	decodeResponse(t, create, &createResponse)
	if createResponse.Approval.ID == 0 || createResponse.Approval.Status != approvalStatusPending {
		t.Fatalf("unexpected created approval: %#v", createResponse.Approval)
	}

	approve := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/approvals/"+strconv.Itoa(createResponse.Approval.ID)+"/approve", map[string]any{
		"comment": "审批通过",
	}, adminCookie)
	if approve.Code != http.StatusOK {
		t.Fatalf("expected approval approve status 200, got %d: %s", approve.Code, approve.Body.String())
	}

	execute := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/admin/approvals/"+strconv.Itoa(createResponse.Approval.ID)+"/execute", nil, adminCookie)
	if execute.Code != http.StatusOK {
		t.Fatalf("expected approval execute status 200, got %d: %s", execute.Code, execute.Body.String())
	}

	record, _ := router.findApprovalLocked(createResponse.Approval.ID)
	if record.Status != approvalStatusExecuted {
		t.Fatalf("expected approval status executed, got %#v", record)
	}

	config := router.findConfig(100)
	if config == nil || config.Settings.Model != "gpt-5-mini" {
		t.Fatalf("expected config updated from approval execution, got %#v", config)
	}
}

func TestAdminRuntimeRestartRequiresApproval(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/admin/instances/100/runtime/restart", nil)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected restart without approval status 409, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
