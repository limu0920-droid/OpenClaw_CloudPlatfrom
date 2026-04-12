package httpapi

import (
	"net/http"
	"strings"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

func (r *Router) handleDeletePortalInstance(w http.ResponseWriter, req *http.Request) {
	r.handleDeleteInstance(w, req, "/api/v1/portal/instances/", "portal-user")
}

func (r *Router) handleDeleteAdminInstance(w http.ResponseWriter, req *http.Request) {
	r.handleDeleteInstance(w, req, "/api/v1/admin/instances/", "admin-user")
}

func (r *Router) handleDeleteInstance(w http.ResponseWriter, req *http.Request, prefix string, actor string) {
	instanceID, ok := parseTailID(req.URL.Path, prefix)
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
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
	if _, err := r.resolveApprovedApproval(approvalNo, approvalTypeDeleteInstance, instanceID); err != nil {
		writeApprovalError(w, err)
		return
	}

	response, err := r.performDeleteInstance(instanceID, actor)
	if err != nil {
		writeApprovalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func isDeletedInstance(instance models.Instance) bool {
	return strings.EqualFold(strings.TrimSpace(instance.Status), "deleted")
}

func deleteResultStatus(result runtimeadapter.ActionResult) string {
	if strings.TrimSpace(result.Status) == "" {
		return "accepted"
	}
	return strings.ToLower(strings.TrimSpace(result.Status))
}

func deleteAuditResult(result runtimeadapter.ActionResult) string {
	if strings.TrimSpace(result.Status) == "" {
		return "success"
	}
	return strings.ToLower(strings.TrimSpace(result.Status))
}
