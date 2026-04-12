package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"openclaw/platformapi/internal/models"
)

func (r *Router) handlePortalArtifacts(w http.ResponseWriter, req *http.Request) {
	r.listArtifactCenter(w, req, "portal")
}

func (r *Router) handlePortalArtifactDetail(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/portal/artifacts/")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.getArtifactCenterDetail(w, req, artifactID, "portal")
}

func (r *Router) handleAdminArtifacts(w http.ResponseWriter, req *http.Request) {
	r.listArtifactCenter(w, req, "admin")
}

func (r *Router) handleAdminArtifactDetail(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/admin/artifacts/")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.getArtifactCenterDetail(w, req, artifactID, "admin")
}

func (r *Router) listArtifactCenter(w http.ResponseWriter, req *http.Request, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	q := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("q")))
	kind := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("kind")))
	instanceFilter, _ := strconv.Atoi(strings.TrimSpace(req.URL.Query().Get("instanceId")))

	r.mu.RLock()
	rawInstances := append([]models.Instance(nil), r.data.Instances...)
	sessions := append([]models.WorkspaceSession(nil), r.data.WorkspaceSessions...)
	artifacts := append([]models.WorkspaceArtifact(nil), r.data.WorkspaceArtifacts...)
	messages := append([]models.WorkspaceMessage(nil), r.data.WorkspaceMessages...)
	logs := append([]models.WorkspaceArtifactAccessLog(nil), r.data.WorkspaceArtifactLogs...)
	favorites := append([]models.WorkspaceArtifactFavorite(nil), r.data.WorkspaceArtifactFavorites...)
	shares := append([]models.WorkspaceArtifactShare(nil), r.data.WorkspaceArtifactShares...)
	r.mu.RUnlock()

	instances := r.resolveLiveInstances(rawInstances)
	instanceMap := make(map[int]models.Instance, len(instances))
	for _, instance := range instances {
		if r.canAccessWorkspaceInstance(actor, instance, scope) {
			instanceMap[instance.ID] = instance
		}
	}

	filteredSessions := make([]models.WorkspaceSession, 0, len(sessions))
	for _, session := range sessions {
		instance, ok := instanceMap[session.InstanceID]
		if !ok || !r.canAccessWorkspaceSession(actor, session, instance, scope) {
			continue
		}
		filteredSessions = append(filteredSessions, session)
	}

	filteredArtifacts := make([]models.WorkspaceArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if _, ok := instanceMap[artifact.InstanceID]; ok {
			filteredArtifacts = append(filteredArtifacts, artifact)
		}
	}

	items := r.buildArtifactCenterItems(req, actor, scope, filteredArtifacts, filteredSessions, instanceMap, messages, logs, favorites, shares)
	recentViewed := buildRecentViewedArtifacts(actor, scope, logs, items)
	matched := make([]portalArtifactCenterItem, 0, len(items))
	for _, item := range items {
		if kind != "" && strings.ToLower(item.Kind) != kind {
			continue
		}
		if instanceFilter > 0 && item.InstanceID != instanceFilter {
			continue
		}
		if q != "" {
			haystack := strings.ToLower(strings.Join([]string{
				item.Title,
				item.Kind,
				item.InstanceName,
				item.SessionTitle,
				item.SourceURL,
			}, " "))
			if !strings.Contains(haystack, q) {
				continue
			}
		}
		matched = append(matched, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":        matched,
		"recentViewed": recentViewed,
		"stats":        buildArtifactCenterStats(matched, recentViewed),
	})
}

func (r *Router) getArtifactCenterDetail(w http.ResponseWriter, req *http.Request, artifactID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	artifact, session, instance, found, allowed := r.findWorkspaceArtifactContext(actor, artifactID, scope)
	if !found {
		http.NotFound(w, req)
		return
	}
	if !allowed {
		http.NotFound(w, req)
		return
	}

	if err := r.recordWorkspaceArtifactAccess(actor, artifact, "detail", scope, req); err != nil {
		http.Error(w, "persist artifact detail access failed", http.StatusInternalServerError)
		return
	}
	if shareToken := strings.TrimSpace(req.URL.Query().Get("share")); shareToken != "" {
		if err := r.markArtifactShareOpened(actor, artifact, scope, shareToken); err != nil {
			http.Error(w, "invalid or expired artifact share", http.StatusForbidden)
			return
		}
	}

	r.mu.RLock()
	rawInstances := append([]models.Instance(nil), r.data.Instances...)
	allSessions := append([]models.WorkspaceSession(nil), r.data.WorkspaceSessions...)
	artifacts := append([]models.WorkspaceArtifact(nil), r.data.WorkspaceArtifacts...)
	messages := append([]models.WorkspaceMessage(nil), r.data.WorkspaceMessages...)
	accessLogs := append([]models.WorkspaceArtifactAccessLog(nil), r.filterWorkspaceArtifactLogsByArtifact(artifact.ID)...)
	allLogs := append([]models.WorkspaceArtifactAccessLog(nil), r.data.WorkspaceArtifactLogs...)
	favorites := append([]models.WorkspaceArtifactFavorite(nil), r.data.WorkspaceArtifactFavorites...)
	shares := append([]models.WorkspaceArtifactShare(nil), r.data.WorkspaceArtifactShares...)
	sessionSummary := r.buildWorkspaceSessionSummaryLocked(session)
	var messageRecord *models.WorkspaceMessage
	if artifact.MessageID > 0 {
		for _, item := range r.data.WorkspaceMessages {
			if item.ID == artifact.MessageID {
				copy := item
				messageRecord = &copy
				break
			}
		}
	}
	tenantName := ""
	if tenant := r.findTenant(artifact.TenantID); tenant != nil {
		tenantName = tenant.Name
	}
	r.mu.RUnlock()

	items := r.buildArtifactCenterItems(req, actor, scope, artifacts, []models.WorkspaceSession{session}, map[int]models.Instance{instance.ID: instance}, messages, allLogs, favorites, shares)
	var detailItem portalArtifactCenterItem
	for _, item := range items {
		if item.ID == artifact.ID {
			detailItem = item
			break
		}
	}
	detailItem.TenantName = tenantName
	versionItems := make([]portalArtifactCenterItem, 0)
	if detailItem.LineageKey != "" {
		resolvedInstances := r.resolveLiveInstances(rawInstances)
		instanceMap := make(map[int]models.Instance, len(resolvedInstances))
		for _, item := range resolvedInstances {
			if r.canAccessWorkspaceInstance(actor, item, scope) {
				instanceMap[item.ID] = item
			}
		}
		visibleSessions := make([]models.WorkspaceSession, 0, len(allSessions))
		for _, item := range allSessions {
			instance, ok := instanceMap[item.InstanceID]
			if !ok || !r.canAccessWorkspaceSession(actor, item, instance, scope) {
				continue
			}
			visibleSessions = append(visibleSessions, item)
		}
		fullItems := r.buildArtifactCenterItems(req, actor, scope, artifacts, visibleSessions, instanceMap, messages, allLogs, favorites, shares)
		for _, item := range fullItems {
			if item.LineageKey == detailItem.LineageKey {
				versionItems = append(versionItems, item)
			}
		}
		sort.Slice(versionItems, func(i, j int) bool {
			if versionItems[i].Version == versionItems[j].Version {
				return sortByTimeDesc(versionItems[i].UpdatedAt, versionItems[j].UpdatedAt)
			}
			return versionItems[i].Version > versionItems[j].Version
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"artifact":   detailItem,
		"preview":    detailItem.Preview,
		"session":    sessionSummary,
		"message":    messageRecord,
		"accessLogs": accessLogs,
		"versions":   versionItems,
		"shares":     buildArtifactShareSummaries(scope, artifact.ID, shares),
		"tenantName": tenantName,
	})
}

func workspaceScopePath(scope string, instanceID int, artifactURL string) string {
	base := "/portal"
	if scope == "admin" {
		base = "/admin"
	}
	path := fmt.Sprintf("%s/instances/%d/workspace", base, instanceID)
	if strings.TrimSpace(artifactURL) == "" {
		return path
	}
	return path + "?artifact=" + url.QueryEscape(strings.TrimSpace(artifactURL))
}
