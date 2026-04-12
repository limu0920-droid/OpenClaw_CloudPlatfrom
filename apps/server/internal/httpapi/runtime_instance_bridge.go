package httpapi

import (
	"net/http"
	"strings"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

type instanceRuntimeScaleRequest struct {
	Replicas int `json:"replicas"`
}

func (r *Router) handleAdminInstanceRuntime(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/admin/instances/", "/runtime")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	state, found := r.loadLiveInstanceState(instanceID)
	if !found || state.Binding == nil {
		http.NotFound(w, req)
		return
	}

	response := map[string]any{
		"instance": state.Instance,
		"binding":  state.Binding,
		"runtime":  state.Runtime,
	}
	if state.Workload != nil {
		response["workload"] = state.Workload
		response["pods"] = r.runtime.ListPods(state.Binding.WorkloadID)
	}
	if state.Metrics != nil {
		response["metrics"] = *state.Metrics
	}
	if accessProvider, ok := r.runtime.(runtimeadapter.AccessInfoProvider); ok {
		response["accessEndpoints"] = accessProvider.GetWorkloadAccess(state.Binding.WorkloadID)
	}

	writeJSON(w, http.StatusOK, response)
}

func (r *Router) handleAdminInstanceRuntimeAction(w http.ResponseWriter, req *http.Request) {
	action := "restart"
	switch {
	case hasRuntimeAction(req.URL.Path, "/runtime/start"):
		action = "start"
	case hasRuntimeAction(req.URL.Path, "/runtime/stop"), hasRuntimeAction(req.URL.Path, "/runtime/pause"):
		action = "stop"
	}

	var instanceID int
	var ok bool
	switch action {
	case "start":
		instanceID, ok = parseInstanceID(req.URL.Path, "/api/v1/admin/instances/", "/runtime/start")
	case "stop":
		instanceID, ok = parseAdminRuntimeStopID(req.URL.Path)
	default:
		instanceID, ok = parseInstanceID(req.URL.Path, "/api/v1/admin/instances/", "/runtime/restart")
	}
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}
	if action != "start" {
		approvalNo := strings.TrimSpace(req.URL.Query().Get("approvalNo"))
		if approvalNo == "" {
			approvalNo = strings.TrimSpace(req.Header.Get("X-OpenClaw-Approval-No"))
		}
		if approvalNo == "" {
			http.Error(w, "approval required", http.StatusConflict)
			return
		}
		requiredType := approvalTypeRuntimeRestart
		if action == "stop" {
			requiredType = approvalTypeRuntimeStop
		}
		if _, err := r.resolveApprovedApproval(approvalNo, requiredType, instanceID); err != nil {
			writeApprovalError(w, err)
			return
		}
	}

	response, err := r.performAdminRuntimeAction(instanceID, action, 0, "admin-user")
	if err != nil {
		writeApprovalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) handleAdminInstanceRuntimeScale(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/admin/instances/", "/runtime/scale")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	var payload instanceRuntimeScaleRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	approvalNo := strings.TrimSpace(req.URL.Query().Get("approvalNo"))
	if approvalNo == "" {
		approvalNo = strings.TrimSpace(req.Header.Get("X-OpenClaw-Approval-No"))
	}
	if approvalNo == "" {
		http.Error(w, "approval required", http.StatusConflict)
		return
	}
	if _, err := r.resolveApprovedApproval(approvalNo, approvalTypeRuntimeScale, instanceID); err != nil {
		writeApprovalError(w, err)
		return
	}

	response, err := r.performAdminRuntimeAction(instanceID, "scale", payload.Replicas, "admin-user")
	if err != nil {
		writeApprovalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) findRuntimeBinding(instanceID int) *models.RuntimeBinding {
	for _, item := range r.data.RuntimeBindings {
		if item.InstanceID == instanceID {
			copy := item
			return &copy
		}
	}
	return nil
}

func hasRuntimeAction(path string, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}

func parseAdminRuntimeStopID(path string) (int, bool) {
	if hasRuntimeAction(path, "/runtime/pause") {
		return parseInstanceID(path, "/api/v1/admin/instances/", "/runtime/pause")
	}
	return parseInstanceID(path, "/api/v1/admin/instances/", "/runtime/stop")
}
