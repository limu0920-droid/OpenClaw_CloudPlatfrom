package httpapi

import (
	"strings"

	"openclaw/platformapi/internal/models"
)

func (r *Router) resolveUserFromIdentity(provider string, email string, openID string, unionID string, subject string, externalName string) (*models.UserProfile, *models.AuthIdentity) {
	if identity := r.findAuthIdentity(provider, email, openID, unionID, subject); identity != nil {
		if profile := r.findUserProfile(identity.UserID); profile != nil {
			return profile, identity
		}
	}

	if email != "" {
		if profile := r.findUserByEmail(email); profile != nil {
			return profile, nil
		}
	}

	return nil, nil
}

func (r *Router) findAuthIdentity(provider string, email string, openID string, unionID string, subject string) *models.AuthIdentity {
	for _, item := range r.data.AuthIdentities {
		if item.Provider != provider {
			continue
		}
		if subject != "" && item.Subject == subject {
			copy := item
			return &copy
		}
		if email != "" && strings.EqualFold(item.Email, email) {
			copy := item
			return &copy
		}
		if openID != "" && item.OpenID == openID {
			copy := item
			return &copy
		}
		if unionID != "" && item.UnionID == unionID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) upsertAuthIdentity(identity models.AuthIdentity) {
	for index, item := range r.data.AuthIdentities {
		if item.Provider == identity.Provider && ((identity.Subject != "" && item.Subject == identity.Subject) || (identity.UnionID != "" && item.UnionID == identity.UnionID) || (identity.OpenID != "" && item.OpenID == identity.OpenID)) {
			r.data.AuthIdentities[index] = identity
			return
		}
	}
	r.data.AuthIdentities = append([]models.AuthIdentity{identity}, r.data.AuthIdentities...)
}

func (r *Router) findUserProfile(userID int) *models.UserProfile {
	for _, item := range r.data.Users {
		if item.ID == userID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) findUserByEmail(email string) *models.UserProfile {
	for _, item := range r.data.Users {
		if strings.EqualFold(item.Email, email) {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) filterAuthIdentitiesByUser(userID int) []models.AuthIdentity {
	out := make([]models.AuthIdentity, 0)
	for _, item := range r.data.AuthIdentities {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out
}

func (r *Router) syncIdentityProfile(user models.UserProfile) {
	for index, item := range r.data.AuthIdentities {
		if item.UserID != user.ID {
			continue
		}
		r.data.AuthIdentities[index].Email = user.Email
		r.data.AuthIdentities[index].ExternalName = user.DisplayName
		r.data.AuthIdentities[index].UpdatedAt = nowRFC3339()
	}
}

func (r *Router) findConflictingIdentity(candidate models.AuthIdentity) *models.AuthIdentity {
	for _, item := range r.data.AuthIdentities {
		if item.Provider != candidate.Provider {
			continue
		}
		if item.UserID == candidate.UserID {
			continue
		}
		if candidate.Subject != "" && item.Subject == candidate.Subject {
			copy := item
			return &copy
		}
		if candidate.UnionID != "" && item.UnionID == candidate.UnionID {
			copy := item
			return &copy
		}
		if candidate.OpenID != "" && item.OpenID == candidate.OpenID {
			copy := item
			return &copy
		}
		if candidate.Email != "" && strings.EqualFold(item.Email, candidate.Email) {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) ensurePrimaryIdentity(userID int, identityID int) {
	hasPrimary := false
	for _, item := range r.data.AuthIdentities {
		if item.UserID == userID && item.IsPrimary {
			hasPrimary = true
			break
		}
	}
	if hasPrimary {
		return
	}
	for index, item := range r.data.AuthIdentities {
		if item.UserID == userID && (identityID == 0 || item.ID == identityID) {
			r.data.AuthIdentities[index].IsPrimary = true
			r.data.AuthIdentities[index].UpdatedAt = nowRFC3339()
			return
		}
	}
}

func (r *Router) setPrimaryIdentity(userID int, identityID int) bool {
	found := false
	for index, item := range r.data.AuthIdentities {
		if item.UserID != userID {
			continue
		}
		r.data.AuthIdentities[index].IsPrimary = item.ID == identityID
		r.data.AuthIdentities[index].UpdatedAt = nowRFC3339()
		if item.ID == identityID {
			found = true
		}
	}
	return found
}

func (r *Router) canUnbindIdentity(userID int, identityID int) bool {
	count := 0
	for _, item := range r.data.AuthIdentities {
		if item.UserID == userID && item.Status == "active" {
			count++
		}
	}
	if count <= 1 {
		return false
	}
	for _, item := range r.data.AuthIdentities {
		if item.UserID == userID && item.ID == identityID && item.IsPrimary {
			return false
		}
	}
	return true
}

func (r *Router) canDisableIdentity(userID int, identityID int) bool {
	count := 0
	for _, item := range r.data.AuthIdentities {
		if item.UserID == userID && item.Status == "active" {
			count++
		}
	}
	if count <= 1 {
		return false
	}
	for _, item := range r.data.AuthIdentities {
		if item.UserID == userID && item.ID == identityID && item.IsPrimary {
			return false
		}
	}
	return true
}
