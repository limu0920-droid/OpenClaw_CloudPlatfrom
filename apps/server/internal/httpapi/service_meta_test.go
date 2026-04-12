package httpapi

import (
	"net/http"
	"testing"
)

func TestServiceMetadataHealthzHeadersAndBody(t *testing.T) {
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			BuildVersion: "dev",
			BuildCommit:  "abc123",
			BuildDate:    "2026-04-08T00:00:00Z",
		},
		runtime: newTestRuntimeAdapter(),
	}

	recorder := performRequest(t, router, http.MethodGet, "/healthz", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected healthz status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("X-Mock-Source"); got != "" {
		t.Fatalf("expected X-Mock-Source removed, got %q", got)
	}
	if got := recorder.Header().Get("X-OpenClaw-Service"); got != "openclaw-platform-api" {
		t.Fatalf("expected X-OpenClaw-Service header, got %q", got)
	}
	if got := recorder.Header().Get("X-OpenClaw-Mode"); got != "seeded" {
		t.Fatalf("expected seeded mode header, got %q", got)
	}

	var response struct {
		Status            string `json:"status"`
		Name              string `json:"name"`
		Service           string `json:"service"`
		Mode              string `json:"mode"`
		StateBackend      string `json:"stateBackend"`
		RuntimeProvider   string `json:"runtimeProvider"`
		BootstrapStrategy string `json:"bootstrapStrategy"`
	}
	decodeResponse(t, recorder, &response)

	if response.Name != "OpenClaw Platform API" {
		t.Fatalf("expected healthz name OpenClaw Platform API, got %q", response.Name)
	}
	if response.Service != "openclaw-platform-api" {
		t.Fatalf("expected service name, got %q", response.Service)
	}
	if response.Mode != "seeded" {
		t.Fatalf("expected seeded mode, got %q", response.Mode)
	}
	if response.StateBackend != "memory" {
		t.Fatalf("expected memory state backend, got %q", response.StateBackend)
	}
	if response.RuntimeProvider != "kubectl" {
		t.Fatalf("expected kubectl runtime provider, got %q", response.RuntimeProvider)
	}
	if response.BootstrapStrategy != "memory-seed" {
		t.Fatalf("expected memory-seed bootstrap strategy, got %q", response.BootstrapStrategy)
	}
}

func TestServiceMetadataBootstrapIncludesPersistentMetadata(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			BuildVersion:      "1.2.3",
			BuildCommit:       "deadbeef",
			BuildDate:         "2026-04-08T01:02:03Z",
			CoreStore:         store,
			RuntimeProvider:   "kubectl",
			BootstrapStrategy: "database-load",
		},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/bootstrap", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected bootstrap status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("X-OpenClaw-State-Backend"); got != "postgres" {
		t.Fatalf("expected postgres state backend header, got %q", got)
	}
	if got := recorder.Header().Get("X-OpenClaw-Runtime-Provider"); got != "kubectl" {
		t.Fatalf("expected kubectl runtime provider header, got %q", got)
	}

	var response struct {
		App struct {
			Name              string `json:"name"`
			Service           string `json:"service"`
			Version           string `json:"version"`
			Commit            string `json:"commit"`
			BuiltAt           string `json:"builtAt"`
			Mode              string `json:"mode"`
			StateBackend      string `json:"stateBackend"`
			RuntimeProvider   string `json:"runtimeProvider"`
			BootstrapStrategy string `json:"bootstrapStrategy"`
		} `json:"app"`
	}
	decodeResponse(t, recorder, &response)

	if response.App.Name != "OpenClaw Platform API" {
		t.Fatalf("expected bootstrap app name, got %q", response.App.Name)
	}
	if response.App.Service != "openclaw-platform-api" {
		t.Fatalf("expected bootstrap service, got %q", response.App.Service)
	}
	if response.App.Version != "1.2.3" {
		t.Fatalf("expected bootstrap version 1.2.3, got %q", response.App.Version)
	}
	if response.App.Commit != "deadbeef" {
		t.Fatalf("expected bootstrap commit deadbeef, got %q", response.App.Commit)
	}
	if response.App.Mode != "persistent" {
		t.Fatalf("expected bootstrap mode persistent, got %q", response.App.Mode)
	}
	if response.App.StateBackend != "postgres" {
		t.Fatalf("expected bootstrap state backend postgres, got %q", response.App.StateBackend)
	}
	if response.App.RuntimeProvider != "kubectl" {
		t.Fatalf("expected bootstrap runtime provider kubectl, got %q", response.App.RuntimeProvider)
	}
	if response.App.BootstrapStrategy != "database-load" {
		t.Fatalf("expected bootstrap strategy database-load, got %q", response.App.BootstrapStrategy)
	}
}

func TestServiceMetadataVersionzIncludesMode(t *testing.T) {
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			BuildVersion:      "2.0.0",
			BuildCommit:       "cafebabe",
			BuildDate:         "2026-04-08T02:03:04Z",
			RuntimeProvider:   "kubectl",
			BootstrapStrategy: "memory-seed",
		},
		runtime: newTestRuntimeAdapter(),
	}

	recorder := performRequest(t, router, http.MethodGet, "/versionz", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected versionz status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Name              string `json:"name"`
		Service           string `json:"service"`
		Version           string `json:"version"`
		Commit            string `json:"commit"`
		BuiltAt           string `json:"builtAt"`
		Mode              string `json:"mode"`
		StateBackend      string `json:"stateBackend"`
		RuntimeProvider   string `json:"runtimeProvider"`
		BootstrapStrategy string `json:"bootstrapStrategy"`
	}
	decodeResponse(t, recorder, &response)

	if response.Name != "OpenClaw Platform API" {
		t.Fatalf("expected versionz name OpenClaw Platform API, got %q", response.Name)
	}
	if response.Service != "openclaw-platform-api" {
		t.Fatalf("expected versionz service openclaw-platform-api, got %q", response.Service)
	}
	if response.Version != "2.0.0" || response.Commit != "cafebabe" || response.BuiltAt != "2026-04-08T02:03:04Z" {
		t.Fatalf("unexpected build info: %#v", response)
	}
	if response.Mode != "seeded" || response.StateBackend != "memory" || response.RuntimeProvider != "kubectl" {
		t.Fatalf("unexpected service metadata: %#v", response)
	}
}
