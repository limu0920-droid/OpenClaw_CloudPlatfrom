package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultServiceName        = "openclaw-platform-api"
	defaultServiceDisplayName = "OpenClaw Platform API"
)

func (r *Router) serviceName() string {
	if value := strings.TrimSpace(r.config.ServiceName); value != "" {
		return value
	}
	return defaultServiceName
}

func (r *Router) serviceDisplayName() string {
	return defaultServiceDisplayName
}

func (r *Router) serviceMode() string {
	if r.store != nil {
		return "persistent"
	}
	return "seeded"
}

func (r *Router) stateBackend() string {
	if r.store != nil {
		return "postgres"
	}
	return "memory"
}

func (r *Router) runtimeProviderName() string {
	if value := strings.ToLower(strings.TrimSpace(r.config.RuntimeProvider)); value != "" {
		return value
	}
	return "kubectl"
}

func (r *Router) bootstrapStrategy() string {
	if value := strings.ToLower(strings.TrimSpace(r.config.BootstrapStrategy)); value != "" {
		return value
	}
	if r.store != nil {
		return "database-load"
	}
	return "memory-seed"
}

func (r *Router) serviceMetadata() map[string]any {
	return map[string]any{
		"service":           r.serviceName(),
		"mode":              r.serviceMode(),
		"strictMode":        r.config.StrictMode,
		"stateBackend":      r.stateBackend(),
		"runtimeProvider":   r.runtimeProviderName(),
		"bootstrapStrategy": r.bootstrapStrategy(),
	}
}

func (r *Router) setServiceHeaders(w interface{ Header() http.Header }) {
	headers := w.Header()
	headers.Set("X-OpenClaw-Service", r.serviceName())
	headers.Set("X-OpenClaw-Mode", r.serviceMode())
	headers.Set("X-OpenClaw-Strict-Mode", strings.ToLower(strconv.FormatBool(r.config.StrictMode)))
	headers.Set("X-OpenClaw-State-Backend", r.stateBackend())
	headers.Set("X-OpenClaw-Runtime-Provider", r.runtimeProviderName())
	headers.Set("X-OpenClaw-Bootstrap-Strategy", r.bootstrapStrategy())
}
