package httpapi

import (
	"net/http"
	"strings"

	"openclaw/platformapi/internal/runtimeadapter"
)

type scaleWorkloadRequest struct {
	Replicas int `json:"replicas"`
}

func (r *Router) handleRuntimeClusters(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": r.runtime.ListClusters(),
	})
}

func (r *Router) handleRuntimeClusterDetail(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/api/v1/runtime/clusters/")
	if strings.Contains(id, "/") || id == "" {
		http.NotFound(w, req)
		return
	}

	cluster, ok := r.runtime.GetCluster(id)
	if !ok {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"cluster":    cluster,
		"nodes":      r.runtime.ListNodes(id),
		"namespaces": r.runtime.ListNamespaces(id),
	})
}

func (r *Router) handleRuntimeClusterNodes(w http.ResponseWriter, req *http.Request) {
	id, ok := parseRuntimeTail(req.URL.Path, "/api/v1/runtime/clusters/", "/nodes")
	if !ok {
		http.Error(w, "invalid cluster id", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": r.runtime.ListNodes(id)})
}

func (r *Router) handleRuntimeClusterNamespaces(w http.ResponseWriter, req *http.Request) {
	id, ok := parseRuntimeTail(req.URL.Path, "/api/v1/runtime/clusters/", "/namespaces")
	if !ok {
		http.Error(w, "invalid cluster id", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": r.runtime.ListNamespaces(id)})
}

func (r *Router) handleRuntimeWorkloads(w http.ResponseWriter, req *http.Request) {
	clusterID := req.URL.Query().Get("clusterId")
	namespace := req.URL.Query().Get("namespace")
	writeJSON(w, http.StatusOK, map[string]any{
		"items": r.runtime.ListWorkloads(clusterID, namespace),
	})
}

func (r *Router) handleRuntimeWorkloadDetail(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/api/v1/runtime/workloads/")
	if strings.Contains(id, "/") || id == "" {
		http.NotFound(w, req)
		return
	}

	workload, ok := r.runtime.GetWorkload(id)
	if !ok {
		http.NotFound(w, req)
		return
	}

	metrics, _ := r.runtime.GetMetrics(id)
	response := map[string]any{
		"workload": workload,
		"pods":     r.runtime.ListPods(id),
		"metrics":  metrics,
	}
	if accessProvider, ok := r.runtime.(runtimeadapter.AccessInfoProvider); ok {
		response["accessEndpoints"] = accessProvider.GetWorkloadAccess(id)
	}
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) handleRuntimeWorkloadPods(w http.ResponseWriter, req *http.Request) {
	id, ok := parseRuntimeTail(req.URL.Path, "/api/v1/runtime/workloads/", "/pods")
	if !ok {
		http.Error(w, "invalid workload id", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": r.runtime.ListPods(id)})
}

func (r *Router) handleRuntimeWorkloadMetrics(w http.ResponseWriter, req *http.Request) {
	id, ok := parseRuntimeTail(req.URL.Path, "/api/v1/runtime/workloads/", "/metrics")
	if !ok {
		http.Error(w, "invalid workload id", http.StatusBadRequest)
		return
	}
	metrics, found := r.runtime.GetMetrics(id)
	if !found {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"metrics": metrics})
}

func (r *Router) handleRuntimeWorkloadAction(w http.ResponseWriter, req *http.Request) {
	action := "restart"
	switch {
	case strings.HasSuffix(req.URL.Path, "/start"):
		action = "start"
	case strings.HasSuffix(req.URL.Path, "/stop"), strings.HasSuffix(req.URL.Path, "/pause"):
		action = "stop"
	}

	var (
		id string
		ok bool
	)
	switch action {
	case "start":
		id, ok = parseRuntimeTail(req.URL.Path, "/api/v1/runtime/workloads/", "/start")
	case "stop":
		id, ok = parseRuntimeActionID(req.URL.Path, "stop")
	default:
		id, ok = parseRuntimeTail(req.URL.Path, "/api/v1/runtime/workloads/", "/restart")
	}
	if !ok {
		http.Error(w, "invalid workload id", http.StatusBadRequest)
		return
	}

	var (
		result runtimeadapter.ActionResult
		found  bool
	)
	switch action {
	case "start":
		result, found = r.runtime.StartWorkload(id)
	case "stop":
		result, found = r.runtime.StopWorkload(id)
	default:
		result, found = r.runtime.RestartWorkload(id)
	}
	if !found {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}

func (r *Router) handleRuntimeWorkloadScale(w http.ResponseWriter, req *http.Request) {
	id, ok := parseRuntimeTail(req.URL.Path, "/api/v1/runtime/workloads/", "/scale")
	if !ok {
		http.Error(w, "invalid workload id", http.StatusBadRequest)
		return
	}

	var payload scaleWorkloadRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	result, found := r.runtime.ScaleWorkload(id, payload.Replicas)
	if !found {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}

func parseRuntimeTail(path string, prefix string, suffix string) (string, bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.TrimSuffix(trimmed, suffix)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return "", false
	}
	return trimmed, true
}

func parseRuntimeActionID(path string, action string) (string, bool) {
	if action != "stop" {
		return "", false
	}
	if strings.HasSuffix(path, "/pause") {
		return parseRuntimeTail(path, "/api/v1/runtime/workloads/", "/pause")
	}
	return parseRuntimeTail(path, "/api/v1/runtime/workloads/", "/stop")
}
