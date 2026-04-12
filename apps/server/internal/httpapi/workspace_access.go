package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"openclaw/platformapi/internal/models"
)

type workspaceActor struct {
	UserID        int
	TenantID      int
	Name          string
	Email         string
	Role          string
	Authenticated bool
	GlobalAdmin   bool
}

var workspaceGlobalAdminRoles = map[string]struct{}{
	"platform_admin":   {},
	"platform_ops":     {},
	"platform_auditor": {},
	"super_admin":      {},
}

var workspaceAdminRoles = map[string]struct{}{
	"tenant_admin":     {},
	"tenant_operator":  {},
	"tenant_auditor":   {},
	"platform_admin":   {},
	"platform_ops":     {},
	"platform_auditor": {},
	"super_admin":      {},
}

func normalizeWorkspaceRole(role string) string {
	value := strings.TrimSpace(role)
	if value == "" {
		return "tenant_admin"
	}
	return value
}

func isWorkspaceGlobalAdminRole(role string) bool {
	_, ok := workspaceGlobalAdminRoles[normalizeWorkspaceRole(role)]
	return ok
}

func isWorkspaceAdminRole(role string) bool {
	_, ok := workspaceAdminRoles[normalizeWorkspaceRole(role)]
	return ok
}

func (a workspaceActor) identifier() string {
	if strings.TrimSpace(a.Email) != "" {
		return strings.TrimSpace(a.Email)
	}
	if strings.TrimSpace(a.Name) != "" {
		return strings.TrimSpace(a.Name)
	}
	if a.UserID > 0 {
		return fmt.Sprintf("user-%d", a.UserID)
	}
	return "unknown"
}

func (a workspaceActor) canUseAdminScope() bool {
	return isWorkspaceAdminRole(a.Role)
}

func (a workspaceActor) canAccessTenant(scope string, tenantID int) bool {
	if tenantID <= 0 {
		return false
	}
	if scope == "admin" && a.GlobalAdmin {
		return true
	}
	return a.TenantID > 0 && a.TenantID == tenantID
}

func (r *Router) resolveWorkspaceActor(req *http.Request, scope string) (workspaceActor, int, string) {
	if session, ok := r.readAuthSession(req); ok && session.Authenticated {
		actor := workspaceActor{
			UserID:        session.UserID,
			TenantID:      session.TenantID,
			Name:          strings.TrimSpace(session.Name),
			Email:         strings.TrimSpace(session.Email),
			Role:          normalizeWorkspaceRole(session.Role),
			Authenticated: true,
		}
		actor.GlobalAdmin = isWorkspaceGlobalAdminRole(actor.Role)
		if actor.UserID > 0 && (actor.TenantID <= 0 || actor.Name == "" || actor.Email == "") {
			if profile := r.findUserProfile(actor.UserID); profile != nil {
				if actor.TenantID <= 0 {
					actor.TenantID = profile.TenantID
				}
				if actor.Name == "" {
					actor.Name = profile.DisplayName
				}
				if actor.Email == "" {
					actor.Email = profile.Email
				}
			}
		}
		if scope == "admin" && !actor.canUseAdminScope() {
			return actor, http.StatusForbidden, "admin scope requires admin role"
		}
		if actor.TenantID <= 0 && !actor.GlobalAdmin {
			return actor, http.StatusForbidden, "tenant scope is missing"
		}
		return actor, 0, ""
	}

	if r.strictModeEnabled() {
		return workspaceActor{}, http.StatusUnauthorized, "authentication required"
	}

	role := "tenant_admin"
	actor := workspaceActor{
		TenantID: tenantFilterID(req, 1),
		Name:     fmt.Sprintf("dev-%s-user", scope),
		Role:     role,
	}
	actor.GlobalAdmin = isWorkspaceGlobalAdminRole(actor.Role)
	return actor, 0, ""
}

func (r *Router) canAccessWorkspaceInstance(actor workspaceActor, instance models.Instance, scope string) bool {
	return visibleInstance(instance) && actor.canAccessTenant(scope, instance.TenantID)
}

func (r *Router) canAccessWorkspaceSession(actor workspaceActor, session models.WorkspaceSession, instance models.Instance, scope string) bool {
	if session.InstanceID != instance.ID || session.TenantID != instance.TenantID {
		return false
	}
	return r.canAccessWorkspaceInstance(actor, instance, scope)
}

func (r *Router) canAccessWorkspaceArtifact(actor workspaceActor, artifact models.WorkspaceArtifact, session models.WorkspaceSession, instance models.Instance, scope string) bool {
	if artifact.SessionID != session.ID || artifact.InstanceID != instance.ID || artifact.TenantID != session.TenantID {
		return false
	}
	return r.canAccessWorkspaceSession(actor, session, instance, scope)
}

func (r *Router) appendWorkspaceAuditLocked(actor workspaceActor, tenantID int, instanceID int, action string, result string, metadata map[string]string) {
	if tenantID <= 0 {
		tenantID = actor.TenantID
	}
	if tenantID <= 0 {
		return
	}

	payload := cloneWorkspaceAuditMetadata(metadata)
	payload["scopeActorRole"] = normalizeWorkspaceRole(actor.Role)
	payload["scopeActor"] = actor.identifier()
	if actor.UserID > 0 {
		payload["scopeUserId"] = strconv.Itoa(actor.UserID)
	}

	r.data.Audits = append([]models.AuditEvent{{
		ID:        r.nextAuditID(),
		TenantID:  tenantID,
		Actor:     actor.identifier(),
		Action:    action,
		Target:    "instance",
		TargetID:  instanceID,
		Result:    result,
		CreatedAt: nowRFC3339(),
		Metadata:  payload,
	}}, r.data.Audits...)
}

func (r *Router) recordWorkspaceAudit(actor workspaceActor, tenantID int, instanceID int, action string, result string, metadata map[string]string) error {
	r.mu.Lock()
	r.appendWorkspaceAuditLocked(actor, tenantID, instanceID, action, result, metadata)
	r.mu.Unlock()
	return r.persistAllData()
}

func cloneWorkspaceAuditMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return map[string]string{}
	}
	copy := make(map[string]string, len(metadata))
	for key, value := range metadata {
		copy[key] = value
	}
	return copy
}

func validateExternalHTTPURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("url is invalid")
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("url must be absolute")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return parsed.String(), nil
	default:
		return "", fmt.Errorf("url scheme must be http or https")
	}
}

func normalizeWorkspaceArtifactTarget(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "preview") {
		return "preview"
	}
	return "source"
}

func (r *Router) findWorkspaceArtifact(id int) (models.WorkspaceArtifact, bool) {
	for _, item := range r.data.WorkspaceArtifacts {
		if item.ID == id {
			return item, true
		}
	}
	return models.WorkspaceArtifact{}, false
}

func pickWorkspaceArtifactURL(artifact models.WorkspaceArtifact, target string) (string, error) {
	if normalizeWorkspaceArtifactTarget(target) == "preview" && strings.TrimSpace(artifact.PreviewURL) != "" {
		return validateExternalHTTPURL(artifact.PreviewURL)
	}
	return validateExternalHTTPURL(artifact.SourceURL)
}
