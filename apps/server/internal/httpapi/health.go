package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (r *Router) handleHealthz(w http.ResponseWriter, req *http.Request) {
	response, status := r.healthResponse(req.Context(), false)
	writeJSON(w, status, response)
}

func (r *Router) handleReadyz(w http.ResponseWriter, req *http.Request) {
	response, status := r.healthResponse(req.Context(), true)
	writeJSON(w, status, response)
}

func (r *Router) healthResponse(parent context.Context, deep bool) (map[string]any, int) {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()

	checks := map[string]any{}
	statusCode := http.StatusOK

	if deep {
		if r.store != nil {
			if err := r.store.Ping(ctx); err != nil {
				checks["database"] = map[string]any{"status": "error", "error": err.Error()}
				statusCode = http.StatusServiceUnavailable
			} else {
				checks["database"] = map[string]any{"status": "ok"}
			}
		} else {
			checks["database"] = map[string]any{"status": "disabled"}
		}

		if r.config.OpenSearchEnabled && r.config.OpenSearchURL != "" {
			if err := pingHTTP(ctx, r.config.OpenSearchURL+"/_cluster/health"); err != nil {
				checks["opensearch"] = map[string]any{"status": "error", "error": err.Error()}
				statusCode = http.StatusServiceUnavailable
			} else {
				checks["opensearch"] = map[string]any{"status": "ok"}
			}
		} else {
			checks["opensearch"] = map[string]any{"status": "disabled"}
		}

		if r.traceSearchEnabled() {
			if err := pingHTTP(ctx, strings.TrimRight(r.config.TraceSearchURL, "/")+"/_cluster/health"); err != nil {
				checks["traceSearch"] = map[string]any{"status": "error", "error": err.Error()}
				statusCode = http.StatusServiceUnavailable
			} else {
				checks["traceSearch"] = map[string]any{"status": "ok"}
			}
		} else {
			checks["traceSearch"] = map[string]any{"status": "disabled"}
		}
	}

	response := map[string]any{
		"status": "ok",
		"name":   r.serviceDisplayName(),
	}
	for key, value := range r.serviceMetadata() {
		response[key] = value
	}
	if statusCode != http.StatusOK {
		response["status"] = "degraded"
	}
	if deep {
		response["checks"] = checks
	}
	return response, statusCode
}

func pingHTTP(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
