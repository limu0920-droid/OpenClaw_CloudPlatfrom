package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

func TestPortalSelfServiceSummary(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/self-service", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected self-service status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Tenant struct {
			Name string `json:"name"`
		} `json:"tenant"`
		Launchpad struct {
			PrimaryInstanceID int    `json:"primaryInstanceId"`
			WorkspacePath     string `json:"workspacePath"`
		} `json:"launchpad"`
		Onboarding struct {
			CompletedCount int `json:"completedCount"`
			TotalCount     int `json:"totalCount"`
			Steps          []struct {
				Key    string `json:"key"`
				Status string `json:"status"`
			} `json:"steps"`
		} `json:"onboarding"`
		Quotas []struct {
			Key     string `json:"key"`
			Percent int    `json:"percent"`
		} `json:"quotas"`
		Reminders []struct {
			Severity string `json:"severity"`
			Title    string `json:"title"`
		} `json:"reminders"`
		RecentArtifacts []struct {
			Title        string `json:"title"`
			WorkspacePath string `json:"workspacePath"`
		} `json:"recentArtifacts"`
	} 
	decodeResponse(t, recorder, &response)

	if response.Tenant.Name != "Acme Studio" {
		t.Fatalf("expected tenant name Acme Studio, got %q", response.Tenant.Name)
	}
	if response.Launchpad.PrimaryInstanceID != 100 {
		t.Fatalf("expected primary instance 100, got %d", response.Launchpad.PrimaryInstanceID)
	}
	if !strings.Contains(response.Launchpad.WorkspacePath, "/portal/instances/100/workspace") {
		t.Fatalf("expected workspace path for instance 100, got %q", response.Launchpad.WorkspacePath)
	}
	if response.Onboarding.TotalCount == 0 || len(response.Onboarding.Steps) == 0 {
		t.Fatal("expected onboarding steps")
	}
	if response.Onboarding.CompletedCount == 0 {
		t.Fatal("expected at least one completed onboarding step")
	}
	if len(response.Quotas) < 3 {
		t.Fatalf("expected quota cards, got %d", len(response.Quotas))
	}
	if len(response.Reminders) == 0 {
		t.Fatal("expected reminders for current seed data")
	}
	if len(response.RecentArtifacts) == 0 {
		t.Fatal("expected recent artifacts")
	}
	if !strings.Contains(response.RecentArtifacts[0].WorkspacePath, "/portal/instances/100/workspace") {
		t.Fatalf("expected artifact workspace path, got %q", response.RecentArtifacts[0].WorkspacePath)
	}
}

func TestPortalArtifactCenterFilters(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/artifacts?kind=pptx&instanceId=100&q=%E8%B7%AF%E6%BC%94", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected artifact center status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Items []struct {
			Title        string `json:"title"`
			Kind         string `json:"kind"`
			InstanceID   int    `json:"instanceId"`
			InstanceName string `json:"instanceName"`
			SessionTitle string `json:"sessionTitle"`
			WorkspacePath string `json:"workspacePath"`
		} `json:"items"`
	}
	decodeResponse(t, recorder, &response)

	if len(response.Items) != 1 {
		t.Fatalf("expected exactly one filtered artifact, got %d", len(response.Items))
	}
	item := response.Items[0]
	if item.Kind != "pptx" {
		t.Fatalf("expected pptx artifact, got %q", item.Kind)
	}
	if item.InstanceID != 100 {
		t.Fatalf("expected instance 100, got %d", item.InstanceID)
	}
	if item.InstanceName != "Acme Prod" {
		t.Fatalf("expected instance Acme Prod, got %q", item.InstanceName)
	}
	if item.SessionTitle != "Acme 官网介绍页迭代" {
		t.Fatalf("unexpected session title %q", item.SessionTitle)
	}
	if !strings.Contains(item.WorkspacePath, "/portal/instances/100/workspace") {
		t.Fatalf("expected workspace path for artifact, got %q", item.WorkspacePath)
	}
}
