package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExternalDocsIndex(t *testing.T) {
	router := newTestRouter(ExternalConfig{
		BuildVersion: "1.0.0",
		BuildCommit:  "abc123",
	})

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/docs/external", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected external docs index status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Service               string `json:"service"`
		Version               string `json:"version"`
		BridgeProtocolVersion string `json:"bridgeProtocolVersion"`
		Docs                  []struct {
			ID          string `json:"id"`
			URL         string `json:"url"`
			ContentType string `json:"contentType"`
		} `json:"docs"`
	}
	decodeResponse(t, recorder, &response)

	if response.Service != "openclaw-platform-api" {
		t.Fatalf("expected service openclaw-platform-api, got %q", response.Service)
	}
	if response.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %q", response.Version)
	}
	if response.BridgeProtocolVersion != workspaceBridgeProtocolVersion {
		t.Fatalf("expected bridge protocol version %q, got %q", workspaceBridgeProtocolVersion, response.BridgeProtocolVersion)
	}
	if len(response.Docs) != 2 {
		t.Fatalf("expected 2 external docs, got %#v", response.Docs)
	}
	if response.Docs[0].ID == "" || response.Docs[0].URL == "" || response.Docs[0].ContentType == "" {
		t.Fatalf("expected populated first doc descriptor, got %#v", response.Docs[0])
	}
}

func TestExternalDocsAssets(t *testing.T) {
	router := newTestRouter(ExternalConfig{})

	testCases := []struct {
		path        string
		contentType string
		contains    string
	}{
		{
			path:        "/api/v1/docs/external/openapi.yaml",
			contentType: "application/yaml",
			contains:    "openapi: 3.1.0",
		},
		{
			path:        "/api/v1/docs/external/integration.md",
			contentType: "text/markdown",
			contains:    "# OpenClaw Platform API External Integration Guide",
		},
	}

	for _, tc := range testCases {
		recorder := performRequest(t, router, http.MethodGet, tc.path, nil)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d: %s", tc.path, recorder.Code, recorder.Body.String())
		}
		if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, tc.contentType) {
			t.Fatalf("expected content type containing %q for %s, got %q", tc.contentType, tc.path, got)
		}
		if !strings.Contains(recorder.Body.String(), tc.contains) {
			t.Fatalf("expected %s to contain %q, got %s", tc.path, tc.contains, recorder.Body.String())
		}
	}
}

func TestExportExternalDocs(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "external-docs")
	if err := ExportExternalDocs(outputDir); err != nil {
		t.Fatalf("expected export external docs to succeed, got %v", err)
	}

	testCases := []struct {
		fileName string
		contains string
	}{
		{
			fileName: "openapi.yaml",
			contains: "openapi: 3.1.0",
		},
		{
			fileName: "integration-guide.md",
			contains: "# OpenClaw Platform API External Integration Guide",
		},
	}

	for _, tc := range testCases {
		body, err := os.ReadFile(filepath.Join(outputDir, tc.fileName))
		if err != nil {
			t.Fatalf("expected exported doc %q to exist: %v", tc.fileName, err)
		}
		if !strings.Contains(string(body), tc.contains) {
			t.Fatalf("expected exported doc %q to contain %q, got %s", tc.fileName, tc.contains, string(body))
		}
	}
}

func TestExportExternalDocsRequiresOutputDir(t *testing.T) {
	if err := ExportExternalDocs("   "); err == nil {
		t.Fatal("expected export external docs to reject an empty output dir")
	}
}
