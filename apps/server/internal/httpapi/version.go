package httpapi

import "net/http"

func (r *Router) handleVersionz(w http.ResponseWriter, req *http.Request) {
	response := map[string]any{
		"name":    r.serviceDisplayName(),
		"version": r.config.BuildVersion,
		"commit":  r.config.BuildCommit,
		"builtAt": r.config.BuildDate,
	}
	for key, value := range r.serviceMetadata() {
		response[key] = value
	}
	writeJSON(w, http.StatusOK, response)
}
