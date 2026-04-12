package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"openclaw/platformapi/internal/models"
)

type updateAdminUserRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Locale      string `json:"locale"`
	Timezone    string `json:"timezone"`
	Department  string `json:"department"`
	Title       string `json:"title"`
	Bio         string `json:"bio"`
}

type bindIdentityRequest struct {
	Provider     string `json:"provider"`
	Subject      string `json:"subject"`
	Email        string `json:"email"`
	OpenID       string `json:"openId"`
	UnionID      string `json:"unionId"`
	ExternalName string `json:"externalName"`
	IsPrimary    bool   `json:"isPrimary"`
}

func (r *Router) handleAdminUsers(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider := strings.TrimSpace(req.URL.Query().Get("provider"))
	if provider == "" {
		writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Users})
		return
	}

	userIDs := make(map[int]bool)
	for _, identity := range r.data.AuthIdentities {
		if identity.Provider == provider {
			userIDs[identity.UserID] = true
		}
	}

	items := make([]models.UserProfile, 0)
	for _, user := range r.data.Users {
		if userIDs[user.ID] {
			items = append(items, user)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleAdminUserDetail(w http.ResponseWriter, req *http.Request) {
	userID, ok := parseTailID(req.URL.Path, "/api/v1/admin/users/")
	if !ok {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	profile := r.findUserProfile(userID)
	if profile == nil {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"profile":         profile,
		"accountSettings": r.findAccountSettings(profile.TenantID),
		"identities":      r.filterAuthIdentitiesByUser(userID),
	})
}

func (r *Router) handleAdminAuthIdentities(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider := strings.TrimSpace(req.URL.Query().Get("provider"))
	if provider == "" {
		writeJSON(w, http.StatusOK, map[string]any{"items": r.data.AuthIdentities})
		return
	}

	items := make([]models.AuthIdentity, 0)
	for _, item := range r.data.AuthIdentities {
		if item.Provider == provider {
			items = append(items, item)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleAdminUpdateUser(w http.ResponseWriter, req *http.Request) {
	userID, ok := parseTailID(req.URL.Path, "/api/v1/admin/users/")
	if !ok {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var payload updateAdminUserRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.findUserProfileIndex(userID)
	if index < 0 {
		http.NotFound(w, req)
		return
	}

	current := r.data.Users[index]
	if payload.DisplayName != "" {
		current.DisplayName = payload.DisplayName
	}
	if payload.Email != "" {
		current.Email = payload.Email
	}
	if payload.Phone != "" {
		current.Phone = payload.Phone
	}
	if payload.Locale != "" {
		current.Locale = payload.Locale
	}
	if payload.Timezone != "" {
		current.Timezone = payload.Timezone
	}
	if payload.Department != "" {
		current.Department = payload.Department
	}
	if payload.Title != "" {
		current.Title = payload.Title
	}
	if payload.Bio != "" {
		current.Bio = payload.Bio
	}
	current.UpdatedAt = nowRFC3339()
	r.data.Users[index] = current
	r.syncIdentityProfile(current)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist admin user failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"profile": current})
}

func (r *Router) handleAdminBindIdentity(w http.ResponseWriter, req *http.Request) {
	userID, ok := parseTailID(req.URL.Path, "/api/v1/admin/users/", "/identities")
	if !ok {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var payload bindIdentityRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if payload.Provider == "" {
		http.Error(w, "provider is required", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	profile := r.findUserProfile(userID)
	if profile == nil {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	identity := models.AuthIdentity{
		ID:           r.nextAuthIdentityID(),
		UserID:       profile.ID,
		TenantID:     profile.TenantID,
		Provider:     payload.Provider,
		IsPrimary:    payload.IsPrimary,
		Status:       "active",
		Subject:      defaultString(payload.Subject, payload.Email),
		Email:        defaultString(payload.Email, profile.Email),
		OpenID:       payload.OpenID,
		UnionID:      payload.UnionID,
		ExternalName: defaultString(payload.ExternalName, profile.DisplayName),
		LastLoginAt:  "",
		UpdatedAt:    nowRFC3339(),
	}
	if conflict := r.findConflictingIdentity(identity); conflict != nil {
		r.mu.Unlock()
		http.Error(w, "identity already bound to another user", http.StatusConflict)
		return
	}
	r.upsertAuthIdentity(identity)
	r.ensurePrimaryIdentity(profile.ID, identity.ID)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist identity bind failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"identity": identity})
}

func (r *Router) handleAdminSetPrimaryIdentity(w http.ResponseWriter, req *http.Request) {
	userID, identityID, ok := parseUserIdentityPrimaryPath(req.URL.Path)
	if !ok {
		http.Error(w, "invalid identity path", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	if !r.setPrimaryIdentity(userID, identityID) {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist primary identity failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"primaryIdentityId": identityID})
}

func (r *Router) handleAdminUnbindIdentity(w http.ResponseWriter, req *http.Request) {
	userID, identityID, ok := parseUserIdentityPath(req.URL.Path)
	if !ok {
		http.Error(w, "invalid identity path", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	if !r.canUnbindIdentity(userID, identityID) {
		r.mu.Unlock()
		http.Error(w, "cannot unbind the last or primary identity", http.StatusConflict)
		return
	}

	index := r.findAuthIdentityIndex(identityID, userID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}
	identity := r.data.AuthIdentities[index]
	r.data.AuthIdentities = append(r.data.AuthIdentities[:index], r.data.AuthIdentities[index+1:]...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist unbind identity failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"identity": identity, "unbound": true})
}

func (r *Router) findUserProfileIndex(userID int) int {
	for index, item := range r.data.Users {
		if item.ID == userID {
			return index
		}
	}
	return -1
}

func (r *Router) nextAuthIdentityID() int {
	maxID := 0
	for _, item := range r.data.AuthIdentities {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) findAuthIdentityIndex(identityID int, userID int) int {
	for index, item := range r.data.AuthIdentities {
		if item.ID == identityID && item.UserID == userID {
			return index
		}
	}
	return -1
}

func parseUserIdentityPath(path string) (int, int, bool) {
	trimmed := strings.TrimPrefix(path, "/api/v1/admin/users/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) != 4 || parts[1] != "identities" || parts[3] != "unbind" {
		return 0, 0, false
	}
	userID, err1 := strconv.Atoi(parts[0])
	identityID, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return userID, identityID, true
}

func parseUserIdentityPrimaryPath(path string) (int, int, bool) {
	trimmed := strings.TrimPrefix(path, "/api/v1/admin/users/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) != 4 || parts[1] != "identities" || parts[3] != "primary" {
		return 0, 0, false
	}
	userID, err1 := strconv.Atoi(parts[0])
	identityID, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return userID, identityID, true
}
