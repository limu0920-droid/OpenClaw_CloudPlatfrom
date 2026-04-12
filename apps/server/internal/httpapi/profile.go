package httpapi

import (
	"net/http"
	"strings"

	"openclaw/platformapi/internal/models"
)

type updateProfileRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	AvatarURL   string `json:"avatarUrl"`
	Locale      string `json:"locale"`
	Timezone    string `json:"timezone"`
	Department  string `json:"department"`
	Title       string `json:"title"`
	Bio         string `json:"bio"`
}

type updatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (r *Router) handlePortalProfile(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profile := r.resolveCurrentUserProfile(req)
	if profile == nil {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": profile})
}

func (r *Router) handlePortalUpdateProfile(w http.ResponseWriter, req *http.Request) {
	var payload updateProfileRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.resolveCurrentUserProfileIndex(req)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	current := r.data.Users[index]
	if strings.TrimSpace(payload.DisplayName) != "" {
		current.DisplayName = payload.DisplayName
	}
	if strings.TrimSpace(payload.Email) != "" {
		current.Email = payload.Email
	}
	if strings.TrimSpace(payload.Phone) != "" {
		current.Phone = payload.Phone
	}
	if strings.TrimSpace(payload.AvatarURL) != "" {
		current.AvatarURL = payload.AvatarURL
	}
	if strings.TrimSpace(payload.Locale) != "" {
		current.Locale = payload.Locale
	}
	if strings.TrimSpace(payload.Timezone) != "" {
		current.Timezone = payload.Timezone
	}
	if strings.TrimSpace(payload.Department) != "" {
		current.Department = payload.Department
	}
	if strings.TrimSpace(payload.Title) != "" {
		current.Title = payload.Title
	}
	if strings.TrimSpace(payload.Bio) != "" {
		current.Bio = payload.Bio
	}
	current.UpdatedAt = nowRFC3339()
	r.data.Users[index] = current
	r.syncIdentityProfile(current)
	if settingsIndex := r.findAccountSettingsIndex(current.TenantID); settingsIndex >= 0 {
		r.data.AccountSettings[settingsIndex].PrimaryEmail = current.Email
		r.data.AccountSettings[settingsIndex].PreferredLocale = current.Locale
		r.data.AccountSettings[settingsIndex].Timezone = current.Timezone
		r.data.AccountSettings[settingsIndex].UpdatedAt = current.UpdatedAt
	}
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist profile failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"profile": current})
}

func (r *Router) handlePortalUpdatePassword(w http.ResponseWriter, req *http.Request) {
	var payload updatePasswordRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.NewPassword) == "" {
		http.Error(w, "newPassword is required", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.resolveCurrentUserProfileIndex(req)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	r.data.Users[index].PasswordMasked = maskPassword(payload.NewPassword)
	r.data.Users[index].UpdatedAt = nowRFC3339()
	password := r.data.Users[index].PasswordMasked
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist password failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"updated":  true,
		"password": password,
	})
}

func (r *Router) resolveCurrentUserProfile(req *http.Request) *models.UserProfile {
	index := r.resolveCurrentUserProfileIndex(req)
	if index < 0 {
		return nil
	}
	copy := r.data.Users[index]
	return &copy
}

func (r *Router) resolveCurrentUserProfileIndex(req *http.Request) int {
	if session, ok := r.readAuthSession(req); ok && session.UserID > 0 {
		for index, item := range r.data.Users {
			if item.ID == session.UserID {
				return index
			}
		}
	}
	tenantID := tenantFilterID(req, 1)
	for index, item := range r.data.Users {
		if item.TenantID == tenantID {
			return index
		}
	}
	return -1
}

func maskPassword(raw string) string {
	if len(raw) <= 4 {
		return "***"
	}
	return raw[:2] + "!***" + raw[len(raw)-2:]
}
