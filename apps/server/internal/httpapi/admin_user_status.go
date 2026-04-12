package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

type updateAdminUserStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type updateAdminIdentityStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func (r *Router) handleAdminUserStatus(w http.ResponseWriter, req *http.Request) {
	userID, ok := parseTailID(req.URL.Path, "/api/v1/admin/users/", "/status")
	if !ok {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var payload updateAdminUserStatusRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	status := strings.TrimSpace(strings.ToLower(payload.Status))
	switch status {
	case "active", "disabled", "locked":
	default:
		http.Error(w, "status must be active, disabled, or locked", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.findUserProfileIndex(userID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	user := r.data.Users[index]
	user.Status = status
	if status == "active" {
		user.LockReason = ""
	} else {
		user.LockReason = strings.TrimSpace(payload.Reason)
	}
	user.UpdatedAt = nowRFC3339()
	r.data.Users[index] = user
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist user status failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"profile": user,
		"status": map[string]any{
			"userId":    user.ID,
			"status":    user.Status,
			"reason":    user.LockReason,
			"updatedAt": user.UpdatedAt,
		},
	})
}

func (r *Router) handleAdminIdentityStatus(w http.ResponseWriter, req *http.Request) {
	userID, identityID, ok := parseUserIdentityStatusPath(req.URL.Path)
	if !ok {
		http.Error(w, "invalid identity path", http.StatusBadRequest)
		return
	}

	var payload updateAdminIdentityStatusRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	status := strings.TrimSpace(strings.ToLower(payload.Status))
	switch status {
	case "active", "disabled":
	default:
		http.Error(w, "status must be active or disabled", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.findAuthIdentityIndex(identityID, userID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	if status == "disabled" && !r.canDisableIdentity(userID, identityID) {
		r.mu.Unlock()
		http.Error(w, "cannot disable the last or primary active identity", http.StatusConflict)
		return
	}

	r.data.AuthIdentities[index].Status = status
	if status == "active" {
		r.data.AuthIdentities[index].StatusReason = ""
	} else {
		r.data.AuthIdentities[index].StatusReason = strings.TrimSpace(payload.Reason)
	}
	r.data.AuthIdentities[index].UpdatedAt = nowRFC3339()
	r.ensurePrimaryIdentity(userID, 0)
	identity := r.data.AuthIdentities[index]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist identity status failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"identity": identity,
	})
}

func parseUserIdentityStatusPath(path string) (int, int, bool) {
	trimmed := strings.TrimPrefix(path, "/api/v1/admin/users/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) != 4 || parts[1] != "identities" || parts[3] != "status" {
		return 0, 0, false
	}
	userID, err1 := strconv.Atoi(parts[0])
	identityID, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return userID, identityID, true
}
