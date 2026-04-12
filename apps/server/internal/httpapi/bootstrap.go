package httpapi

import (
	"net/http"
)

func (r *Router) handleBootstrap(w http.ResponseWriter, req *http.Request) {
	authConfig := map[string]any{
		"provider":        "keycloak",
		"enabled":         r.config.KeycloakEnabled,
		"baseUrl":         r.config.KeycloakBaseURL,
		"realm":           r.config.KeycloakRealm,
		"clientId":        r.config.KeycloakClientID,
		"defaultRedirect": r.config.KeycloakRedirectURL,
	}

	searchConfig := map[string]any{
		"provider": "opensearch",
		"enabled":  r.config.OpenSearchEnabled,
		"url":      r.config.OpenSearchURL,
		"index":    r.config.OpenSearchIndex,
		"export": map[string]any{
			"enabled": true,
		},
		"trace": map[string]any{
			"provider":      r.traceSearchProvider(),
			"enabled":       r.traceSearchEnabled(),
			"url":           r.config.TraceSearchURL,
			"index":         r.config.TraceSearchIndex,
			"publicBaseUrl": r.config.TraceSearchPublicBaseURL,
		},
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	oem := r.resolveOEM(req)
	clusterCount := len(r.runtime.ListClusters())
	workloadCount := len(r.runtime.ListWorkloads("", ""))

	writeJSON(w, http.StatusOK, map[string]any{
		"app": map[string]any{
			"name":              r.serviceDisplayName(),
			"service":           r.serviceName(),
			"version":           r.config.BuildVersion,
			"commit":            r.config.BuildCommit,
			"builtAt":           r.config.BuildDate,
			"mode":              r.serviceMode(),
			"stateBackend":      r.stateBackend(),
			"runtimeProvider":   r.runtimeProviderName(),
			"bootstrapStrategy": r.bootstrapStrategy(),
		},
		"auth":   authConfig,
		"search": searchConfig,
		"oem":    oem,
		"runtime": map[string]any{
			"clusters":  clusterCount,
			"workloads": workloadCount,
		},
		"features": map[string]any{
			"instances": true,
			"channels":  true,
			"tickets":   true,
			"runtime":   true,
			"oem":       true,
			"sso":       r.config.KeycloakEnabled,
			"search":    true,
		},
	})
}
