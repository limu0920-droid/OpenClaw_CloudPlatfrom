package httpapi

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
)

const externalDocsUpdatedAt = "2026-04-09"

//go:embed externaldocs/openapi.yaml externaldocs/integration-guide.md
var externalDocsFS embed.FS

func (r *Router) handleExternalDocsIndex(w http.ResponseWriter, req *http.Request) {
	docs := []map[string]any{
		{
			"id":          "openapi",
			"title":       "OpenClaw Platform API External OpenAPI",
			"format":      "openapi+yaml",
			"contentType": "application/yaml; charset=utf-8",
			"path":        "/api/v1/docs/external/openapi.yaml",
			"url":         externalDocsAbsoluteURL(req, "/api/v1/docs/external/openapi.yaml"),
		},
		{
			"id":          "integration-guide",
			"title":       "OpenClaw Platform API External Integration Guide",
			"format":      "markdown",
			"contentType": "text/markdown; charset=utf-8",
			"path":        "/api/v1/docs/external/integration.md",
			"url":         externalDocsAbsoluteURL(req, "/api/v1/docs/external/integration.md"),
		},
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name":                  r.serviceDisplayName(),
		"service":               r.serviceName(),
		"version":               r.config.BuildVersion,
		"commit":                r.config.BuildCommit,
		"updatedAt":             externalDocsUpdatedAt,
		"bridgeProtocolVersion": workspaceBridgeProtocolVersion,
		"docs":                  docs,
	})
}

func (r *Router) handleExternalDocOpenAPI(w http.ResponseWriter, req *http.Request) {
	serveEmbeddedExternalDoc(w, "externaldocs/openapi.yaml", "application/yaml; charset=utf-8")
}

func (r *Router) handleExternalDocGuide(w http.ResponseWriter, req *http.Request) {
	serveEmbeddedExternalDoc(w, "externaldocs/integration-guide.md", "text/markdown; charset=utf-8")
}

func serveEmbeddedExternalDoc(w http.ResponseWriter, path string, contentType string) {
	body, err := externalDocsFS.ReadFile(path)
	if err != nil {
		http.Error(w, "external doc asset is unavailable", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func externalDocsAbsoluteURL(req *http.Request, path string) string {
	if req == nil {
		return path
	}

	host := strings.TrimSpace(req.Host)
	if host == "" {
		return path
	}

	scheme := "http"
	if forwardedProto := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
		scheme = strings.TrimSpace(strings.Split(forwardedProto, ",")[0])
	} else if req.TLS != nil {
		scheme = "https"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}
