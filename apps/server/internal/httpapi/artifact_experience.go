package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
)

type artifactThumbnailDescriptor struct {
	Mode  string `json:"mode"`
	URL   string `json:"url,omitempty"`
	Label string `json:"label"`
	Hint  string `json:"hint,omitempty"`
}

type artifactQualitySummary struct {
	Status           string `json:"status"`
	Score            int    `json:"score"`
	InlinePreview    bool   `json:"inlinePreview"`
	PreviewMode      string `json:"previewMode"`
	Strategy         string `json:"strategy"`
	FailureReason    string `json:"failureReason,omitempty"`
	LastViewedAt     string `json:"lastViewedAt,omitempty"`
	LastDownloadedAt string `json:"lastDownloadedAt,omitempty"`
	ViewCount        int    `json:"viewCount"`
	DownloadCount    int    `json:"downloadCount"`
}

type artifactFailureBucket struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type artifactCenterStats struct {
	TotalCount         int                     `json:"totalCount"`
	FavoriteCount      int                     `json:"favoriteCount"`
	SharedCount        int                     `json:"sharedCount"`
	VersionedCount     int                     `json:"versionedCount"`
	InlinePreviewCount int                     `json:"inlinePreviewCount"`
	FallbackCount      int                     `json:"fallbackCount"`
	RecentViewedCount  int                     `json:"recentViewedCount"`
	FailureReasons     []artifactFailureBucket `json:"failureReasons"`
}

type artifactShareSummary struct {
	ID           int    `json:"id"`
	Token        string `json:"token"`
	ShareURL     string `json:"shareUrl"`
	Scope        string `json:"scope"`
	Note         string `json:"note,omitempty"`
	CreatedBy    string `json:"createdBy"`
	CreatedAt    string `json:"createdAt"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
	Active       bool   `json:"active"`
	UseCount     int    `json:"useCount"`
	LastOpenedAt string `json:"lastOpenedAt,omitempty"`
}

type artifactVersionMeta struct {
	LineageKey       string
	Version          int
	LatestVersion    int
	ParentArtifactID int
}

type createArtifactShareRequest struct {
	Note          string `json:"note"`
	ExpiresAt     string `json:"expiresAt,omitempty"`
	ExpiresInDays int    `json:"expiresInDays,omitempty"`
}

func (r *Router) handlePortalArtifactFavorite(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseArtifactActionID(req.URL.Path, "/api/v1/portal/artifacts/", "/favorite")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.setArtifactFavorite(w, req, artifactID, "portal", req.Method == http.MethodPost)
}

func (r *Router) handleAdminArtifactFavorite(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseArtifactActionID(req.URL.Path, "/api/v1/admin/artifacts/", "/favorite")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.setArtifactFavorite(w, req, artifactID, "admin", req.Method == http.MethodPost)
}

func (r *Router) handlePortalArtifactShares(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseArtifactActionID(req.URL.Path, "/api/v1/portal/artifacts/", "/shares")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.createArtifactShare(w, req, artifactID, "portal")
}

func (r *Router) handleAdminArtifactShares(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseArtifactActionID(req.URL.Path, "/api/v1/admin/artifacts/", "/shares")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.createArtifactShare(w, req, artifactID, "admin")
}

func (r *Router) handlePortalArtifactShareDelete(w http.ResponseWriter, req *http.Request) {
	shareID, ok := parseTailID(req.URL.Path, "/api/v1/portal/artifact-shares/")
	if !ok {
		http.Error(w, "invalid share id", http.StatusBadRequest)
		return
	}
	r.revokeArtifactShare(w, req, shareID, "portal")
}

func (r *Router) handleAdminArtifactShareDelete(w http.ResponseWriter, req *http.Request) {
	shareID, ok := parseTailID(req.URL.Path, "/api/v1/admin/artifact-shares/")
	if !ok {
		http.Error(w, "invalid share id", http.StatusBadRequest)
		return
	}
	r.revokeArtifactShare(w, req, shareID, "admin")
}

func (r *Router) setArtifactFavorite(w http.ResponseWriter, req *http.Request, artifactID int, scope string, favorite bool) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	artifact, _, _, found, allowed := r.findWorkspaceArtifactContext(actor, artifactID, scope)
	if !found || !allowed {
		http.NotFound(w, req)
		return
	}

	if favorite {
		if err := r.addArtifactFavorite(actor, artifact, scope); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := r.removeArtifactFavorite(actor, artifact, scope); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	r.mu.RLock()
	favoriteCount := 0
	isFavorite := false
	for _, item := range r.data.WorkspaceArtifactFavorites {
		if item.ArtifactID != artifact.ID {
			continue
		}
		favoriteCount++
		if artifactFavoriteMatchesActor(item, actor) {
			isFavorite = true
		}
	}
	r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"artifactId":    artifact.ID,
		"isFavorite":    isFavorite,
		"favoriteCount": favoriteCount,
	})
}

func (r *Router) createArtifactShare(w http.ResponseWriter, req *http.Request, artifactID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	artifact, _, _, found, allowed := r.findWorkspaceArtifactContext(actor, artifactID, scope)
	if !found || !allowed {
		http.NotFound(w, req)
		return
	}

	var payload createArtifactShareRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	expiresAt, err := resolveArtifactShareExpiry(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	share, err := r.addArtifactShare(actor, artifact, scope, strings.TrimSpace(payload.Note), expiresAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"share": summarizeArtifactShare(scope, share),
	})
}

func (r *Router) revokeArtifactShare(w http.ResponseWriter, req *http.Request, shareID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	share, err := r.revokeArtifactShareByID(actor, shareID, scope)
	if err != nil {
		if errors.Is(err, errArtifactShareNotFound) {
			http.NotFound(w, req)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"share": summarizeArtifactShare(scope, share),
	})
}

func (r *Router) addArtifactFavorite(actor workspaceActor, artifact models.WorkspaceArtifact, scope string) error {
	now := nowRFC3339()

	r.mu.Lock()
	for _, item := range r.data.WorkspaceArtifactFavorites {
		if item.ArtifactID == artifact.ID && artifactFavoriteMatchesActor(item, actor) {
			r.mu.Unlock()
			return nil
		}
	}
	favorite := models.WorkspaceArtifactFavorite{
		ID:         r.nextWorkspaceArtifactFavoriteID(),
		ArtifactID: artifact.ID,
		SessionID:  artifact.SessionID,
		TenantID:   artifact.TenantID,
		InstanceID: artifact.InstanceID,
		UserID:     actor.UserID,
		Actor:      actor.identifier(),
		CreatedAt:  now,
	}
	r.data.WorkspaceArtifactFavorites = append([]models.WorkspaceArtifactFavorite{favorite}, r.data.WorkspaceArtifactFavorites...)
	r.appendWorkspaceAuditLocked(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.favorite", "success", map[string]string{
		"artifactId": strconv.Itoa(artifact.ID),
		"scope":      scope,
	})
	r.mu.Unlock()

	return r.persistAllData()
}

func (r *Router) removeArtifactFavorite(actor workspaceActor, artifact models.WorkspaceArtifact, scope string) error {
	r.mu.Lock()
	next := make([]models.WorkspaceArtifactFavorite, 0, len(r.data.WorkspaceArtifactFavorites))
	removed := false
	for _, item := range r.data.WorkspaceArtifactFavorites {
		if item.ArtifactID == artifact.ID && artifactFavoriteMatchesActor(item, actor) {
			removed = true
			continue
		}
		next = append(next, item)
	}
	r.data.WorkspaceArtifactFavorites = next
	if removed {
		r.appendWorkspaceAuditLocked(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.unfavorite", "success", map[string]string{
			"artifactId": strconv.Itoa(artifact.ID),
			"scope":      scope,
		})
	}
	r.mu.Unlock()

	if !removed {
		return nil
	}
	return r.persistAllData()
}

func (r *Router) addArtifactShare(actor workspaceActor, artifact models.WorkspaceArtifact, scope string, note string, expiresAt string) (models.WorkspaceArtifactShare, error) {
	now := nowRFC3339()
	token, err := randomArtifactShareToken()
	if err != nil {
		return models.WorkspaceArtifactShare{}, err
	}

	r.mu.Lock()
	share := models.WorkspaceArtifactShare{
		ID:              r.nextWorkspaceArtifactShareID(),
		ArtifactID:      artifact.ID,
		SessionID:       artifact.SessionID,
		TenantID:        artifact.TenantID,
		InstanceID:      artifact.InstanceID,
		Scope:           scope,
		Token:           token,
		Note:            note,
		CreatedBy:       actor.identifier(),
		CreatedByUserID: actor.UserID,
		UseCount:        0,
		ExpiresAt:       expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	r.data.WorkspaceArtifactShares = append([]models.WorkspaceArtifactShare{share}, r.data.WorkspaceArtifactShares...)
	r.appendWorkspaceAuditLocked(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.share.create", "success", map[string]string{
		"artifactId": strconv.Itoa(artifact.ID),
		"scope":      scope,
	})
	r.mu.Unlock()

	if err := r.persistAllData(); err != nil {
		return models.WorkspaceArtifactShare{}, err
	}
	return share, nil
}

var errArtifactShareNotFound = errors.New("artifact share not found")

func (r *Router) revokeArtifactShareByID(actor workspaceActor, shareID int, scope string) (models.WorkspaceArtifactShare, error) {
	now := nowRFC3339()

	r.mu.Lock()
	var updatedShare models.WorkspaceArtifactShare
	found := false
	for index, item := range r.data.WorkspaceArtifactShares {
		if item.ID != shareID || item.Scope != scope {
			continue
		}
		if !actor.canAccessTenant(scope, item.TenantID) {
			r.mu.Unlock()
			return models.WorkspaceArtifactShare{}, errArtifactShareNotFound
		}
		r.data.WorkspaceArtifactShares[index].RevokedAt = now
		r.data.WorkspaceArtifactShares[index].UpdatedAt = now
		updatedShare = r.data.WorkspaceArtifactShares[index]
		r.appendWorkspaceAuditLocked(actor, updatedShare.TenantID, updatedShare.InstanceID, "workspace.artifact.share.revoke", "success", map[string]string{
			"artifactId": strconv.Itoa(updatedShare.ArtifactID),
			"scope":      scope,
		})
		found = true
		break
	}
	r.mu.Unlock()

	if !found {
		return models.WorkspaceArtifactShare{}, errArtifactShareNotFound
	}
	if err := r.persistAllData(); err != nil {
		return models.WorkspaceArtifactShare{}, err
	}
	return updatedShare, nil
}

func (r *Router) markArtifactShareOpened(actor workspaceActor, artifact models.WorkspaceArtifact, scope string, token string) error {
	now := nowRFC3339()
	currentTime, _ := parseRFC3339(now)

	r.mu.Lock()
	for index, item := range r.data.WorkspaceArtifactShares {
		if item.Token != token || item.ArtifactID != artifact.ID || item.Scope != scope {
			continue
		}
		if !artifactShareIsActive(item, currentTime) {
			r.mu.Unlock()
			return errArtifactShareNotFound
		}
		r.data.WorkspaceArtifactShares[index].UseCount++
		r.data.WorkspaceArtifactShares[index].LastOpenedAt = now
		r.data.WorkspaceArtifactShares[index].UpdatedAt = now
		r.appendWorkspaceAuditLocked(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.share.open", "success", map[string]string{
			"artifactId": strconv.Itoa(artifact.ID),
			"scope":      scope,
		})
		r.mu.Unlock()
		return r.persistAllData()
	}
	r.mu.Unlock()
	return errArtifactShareNotFound
}

func (r *Router) buildArtifactCenterItems(
	req *http.Request,
	actor workspaceActor,
	scope string,
	artifacts []models.WorkspaceArtifact,
	sessions []models.WorkspaceSession,
	instanceMap map[int]models.Instance,
	messages []models.WorkspaceMessage,
	logs []models.WorkspaceArtifactAccessLog,
	favorites []models.WorkspaceArtifactFavorite,
	shares []models.WorkspaceArtifactShare,
) []portalArtifactCenterItem {
	sessionMap := make(map[int]models.WorkspaceSession, len(sessions))
	for _, session := range sessions {
		sessionMap[session.ID] = session
	}

	messageMap := make(map[int]models.WorkspaceMessage, len(messages))
	for _, item := range messages {
		messageMap[item.ID] = item
	}

	logsByArtifact := make(map[int][]models.WorkspaceArtifactAccessLog)
	for _, item := range logs {
		logsByArtifact[item.ArtifactID] = append(logsByArtifact[item.ArtifactID], item)
	}

	favoriteCountByArtifact := make(map[int]int)
	isFavoriteByArtifact := make(map[int]bool)
	for _, item := range favorites {
		favoriteCountByArtifact[item.ArtifactID]++
		if artifactFavoriteMatchesActor(item, actor) {
			isFavoriteByArtifact[item.ArtifactID] = true
		}
	}

	activeSharesByArtifact := make(map[int]int)
	now := time.Now().UTC()
	for _, item := range shares {
		if item.Scope != scope || !artifactShareIsActive(item, now) {
			continue
		}
		activeSharesByArtifact[item.ArtifactID]++
	}

	versionMetaByArtifact := buildArtifactVersionMetadata(artifacts)

	items := make([]portalArtifactCenterItem, 0, len(artifacts))
	for _, artifact := range artifacts {
		session, ok := sessionMap[artifact.SessionID]
		if !ok {
			continue
		}
		instance, ok := instanceMap[artifact.InstanceID]
		if !ok {
			continue
		}
		preview := r.buildWorkspaceArtifactPreviewDescriptor(req, artifact, session, instance, scope)
		quality := buildArtifactQualitySummary(preview, logsByArtifact[artifact.ID])
		versionMeta := versionMetaByArtifact[artifact.ID]
		messagePreview := ""
		if artifact.MessageID > 0 {
			if message, ok := messageMap[artifact.MessageID]; ok {
				messagePreview = compactArtifactPreviewText(message.Content)
			}
		}
		items = append(items, portalArtifactCenterItem{
			ID:               artifact.ID,
			Title:            artifact.Title,
			Kind:             artifact.Kind,
			SourceURL:        artifact.SourceURL,
			PreviewURL:       artifact.PreviewURL,
			ArchiveStatus:    artifact.ArchiveStatus,
			ContentType:      artifact.ContentType,
			SizeBytes:        artifact.SizeBytes,
			Filename:         artifact.Filename,
			CreatedAt:        artifact.CreatedAt,
			UpdatedAt:        artifact.UpdatedAt,
			InstanceID:       artifact.InstanceID,
			InstanceName:     instance.Name,
			InstanceStatus:   instance.Status,
			SessionID:        session.ID,
			SessionNo:        session.SessionNo,
			SessionTitle:     session.Title,
			SessionStatus:    session.Status,
			MessageID:        artifact.MessageID,
			MessagePreview:   messagePreview,
			TenantID:         artifact.TenantID,
			ViewCount:        quality.ViewCount,
			DownloadCount:    quality.DownloadCount,
			WorkspacePath:    workspaceScopePath(scope, artifact.InstanceID, firstNonEmpty(artifact.PreviewURL, artifact.SourceURL)),
			DetailPath:       fmt.Sprintf("/%s/artifacts/%d", scope, artifact.ID),
			LineageKey:       versionMeta.LineageKey,
			Version:          versionMeta.Version,
			LatestVersion:    versionMeta.LatestVersion,
			ParentArtifactID: versionMeta.ParentArtifactID,
			IsFavorite:       isFavoriteByArtifact[artifact.ID],
			FavoriteCount:    favoriteCountByArtifact[artifact.ID],
			ShareCount:       activeSharesByArtifact[artifact.ID],
			Thumbnail:        buildArtifactThumbnailDescriptor(artifact, preview),
			Quality:          quality,
			Preview:          preview,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return sortByTimeDesc(items[i].UpdatedAt, items[j].UpdatedAt)
	})

	return items
}

func buildArtifactCenterStats(items []portalArtifactCenterItem, recentViewed []portalArtifactCenterItem) artifactCenterStats {
	stats := artifactCenterStats{
		TotalCount:        len(items),
		RecentViewedCount: len(recentViewed),
	}
	failureBuckets := make(map[string]int)
	versionedChains := make(map[string]struct{})
	for _, item := range items {
		if item.IsFavorite {
			stats.FavoriteCount++
		}
		stats.SharedCount += item.ShareCount
		if item.LatestVersion > 1 && item.LineageKey != "" {
			versionedChains[item.LineageKey] = struct{}{}
		}
		if item.Quality.InlinePreview {
			stats.InlinePreviewCount++
		} else {
			stats.FallbackCount++
		}
		if strings.TrimSpace(item.Quality.FailureReason) != "" {
			failureBuckets[item.Quality.FailureReason]++
		}
	}
	for reason, count := range failureBuckets {
		stats.FailureReasons = append(stats.FailureReasons, artifactFailureBucket{
			Reason: reason,
			Count:  count,
		})
	}
	sort.Slice(stats.FailureReasons, func(i, j int) bool {
		if stats.FailureReasons[i].Count == stats.FailureReasons[j].Count {
			return stats.FailureReasons[i].Reason < stats.FailureReasons[j].Reason
		}
		return stats.FailureReasons[i].Count > stats.FailureReasons[j].Count
	})
	if len(stats.FailureReasons) > 4 {
		stats.FailureReasons = stats.FailureReasons[:4]
	}
	stats.VersionedCount = len(versionedChains)
	return stats
}

func buildRecentViewedArtifacts(actor workspaceActor, scope string, logs []models.WorkspaceArtifactAccessLog, items []portalArtifactCenterItem) []portalArtifactCenterItem {
	itemByID := make(map[int]portalArtifactCenterItem, len(items))
	for _, item := range items {
		itemByID[item.ID] = item
	}

	candidateLogs := append([]models.WorkspaceArtifactAccessLog(nil), logs...)
	sort.Slice(candidateLogs, func(i, j int) bool {
		return sortByTimeDesc(candidateLogs[i].CreatedAt, candidateLogs[j].CreatedAt)
	})

	result := make([]portalArtifactCenterItem, 0, 6)
	seen := make(map[int]struct{})
	identifier := actor.identifier()
	for _, log := range candidateLogs {
		if log.Scope != scope || log.Actor != identifier {
			continue
		}
		if _, ok := seen[log.ArtifactID]; ok {
			continue
		}
		item, ok := itemByID[log.ArtifactID]
		if !ok {
			continue
		}
		result = append(result, item)
		seen[log.ArtifactID] = struct{}{}
		if len(result) == 6 {
			break
		}
	}
	return result
}

func buildArtifactShareSummaries(scope string, artifactID int, shares []models.WorkspaceArtifactShare) []artifactShareSummary {
	items := make([]artifactShareSummary, 0)
	for _, item := range shares {
		if item.ArtifactID != artifactID || item.Scope != scope {
			continue
		}
		items = append(items, summarizeArtifactShare(scope, item))
	}
	sort.Slice(items, func(i, j int) bool {
		return sortByTimeDesc(items[i].CreatedAt, items[j].CreatedAt)
	})
	return items
}

func summarizeArtifactShare(scope string, item models.WorkspaceArtifactShare) artifactShareSummary {
	return artifactShareSummary{
		ID:           item.ID,
		Token:        item.Token,
		ShareURL:     fmt.Sprintf("/%s/artifacts/%d?share=%s", scope, item.ArtifactID, item.Token),
		Scope:        item.Scope,
		Note:         item.Note,
		CreatedBy:    item.CreatedBy,
		CreatedAt:    item.CreatedAt,
		ExpiresAt:    item.ExpiresAt,
		Active:       item.RevokedAt == "" && artifactShareIsActive(item, time.Now().UTC()),
		UseCount:     item.UseCount,
		LastOpenedAt: item.LastOpenedAt,
	}
}

func buildArtifactThumbnailDescriptor(item models.WorkspaceArtifact, preview workspaceArtifactPreviewDescriptor) artifactThumbnailDescriptor {
	label := strings.ToUpper(strings.TrimSpace(item.Kind))
	if label == "" {
		label = "FILE"
	}
	if preview.Available && preview.Mode == "image" && strings.TrimSpace(preview.PreviewURL) != "" {
		return artifactThumbnailDescriptor{
			Mode:  "image",
			URL:   preview.PreviewURL,
			Label: label,
			Hint:  firstNonEmpty(item.Filename, item.Title),
		}
	}
	return artifactThumbnailDescriptor{
		Mode:  "generated",
		Label: label,
		Hint:  firstNonEmpty(item.Filename, item.Title),
	}
}

func buildArtifactQualitySummary(preview workspaceArtifactPreviewDescriptor, logs []models.WorkspaceArtifactAccessLog) artifactQualitySummary {
	summary := artifactQualitySummary{
		InlinePreview: preview.Available,
		PreviewMode:   preview.Mode,
		Strategy:      preview.Strategy,
		FailureReason: preview.FailureReason,
	}
	for _, log := range logs {
		switch log.Action {
		case "download":
			summary.DownloadCount++
			if summary.LastDownloadedAt == "" || sortByTimeDesc(log.CreatedAt, summary.LastDownloadedAt) {
				summary.LastDownloadedAt = log.CreatedAt
			}
		case "detail", "view":
			summary.ViewCount++
			if summary.LastViewedAt == "" || sortByTimeDesc(log.CreatedAt, summary.LastViewedAt) {
				summary.LastViewedAt = log.CreatedAt
			}
		}
	}
	switch {
	case preview.Available:
		summary.Status = "healthy"
		summary.Score = 96
	case preview.DownloadURL != "":
		summary.Status = "warning"
		summary.Score = 68
	default:
		summary.Status = "blocked"
		summary.Score = 28
	}
	if strings.TrimSpace(summary.PreviewMode) == "" {
		summary.PreviewMode = "download"
	}
	return summary
}

func buildArtifactVersionMetadata(artifacts []models.WorkspaceArtifact) map[int]artifactVersionMeta {
	grouped := make(map[string][]models.WorkspaceArtifact)
	for _, item := range artifacts {
		grouped[normalizeArtifactVersionKey(item)] = append(grouped[normalizeArtifactVersionKey(item)], item)
	}

	result := make(map[int]artifactVersionMeta, len(artifacts))
	for key, group := range grouped {
		sort.Slice(group, func(i, j int) bool {
			left, leftOK := parseRFC3339(group[i].CreatedAt)
			right, rightOK := parseRFC3339(group[j].CreatedAt)
			switch {
			case leftOK && rightOK && left.Equal(right):
				return group[i].ID < group[j].ID
			case leftOK && rightOK:
				return left.Before(right)
			case leftOK:
				return true
			case rightOK:
				return false
			default:
				return group[i].ID < group[j].ID
			}
		})
		for index, item := range group {
			parentArtifactID := 0
			if index > 0 {
				parentArtifactID = group[index-1].ID
			}
			result[item.ID] = artifactVersionMeta{
				LineageKey:       key,
				Version:          index + 1,
				LatestVersion:    len(group),
				ParentArtifactID: parentArtifactID,
			}
		}
	}
	return result
}

func normalizeArtifactVersionKey(item models.WorkspaceArtifact) string {
	seed := normalizeArtifactVersionPart(firstNonEmpty(item.Title, artifactFilenameStem(item), item.SourceURL))
	return fmt.Sprintf("%d:%s:%s", item.SessionID, normalizeArtifactVersionPart(item.Kind), seed)
}

func normalizeArtifactVersionPart(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		".pptx", "",
		".ppt", "",
		".docx", "",
		".doc", "",
		".xlsx", "",
		".xls", "",
		".csv", "",
		".tsv", "",
		".pdf", "",
		".html", "",
		".htm", "",
	)
	normalized = replacer.Replace(normalized)
	normalized = strings.Join(strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '/' || r == '\\' || r == '_' || r == '-' || r == '.'
	}), "-")
	if normalized == "" {
		return "artifact"
	}
	return normalized
}

func artifactFilenameStem(item models.WorkspaceArtifact) string {
	for _, candidate := range []string{item.Filename, item.PreviewURL, item.SourceURL} {
		raw := strings.TrimSpace(candidate)
		if raw == "" {
			continue
		}
		base := path.Base(raw)
		if base == "." || base == "/" || base == "" {
			continue
		}
		ext := path.Ext(base)
		return strings.TrimSuffix(base, ext)
	}
	return ""
}

func compactArtifactPreviewText(value string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(value, "\n", " "))
	if len(trimmed) <= 88 {
		return trimmed
	}
	return trimmed[:88] + "..."
}

func artifactFavoriteMatchesActor(item models.WorkspaceArtifactFavorite, actor workspaceActor) bool {
	if actor.UserID > 0 && item.UserID > 0 {
		return actor.UserID == item.UserID
	}
	return strings.EqualFold(strings.TrimSpace(item.Actor), strings.TrimSpace(actor.identifier()))
}

func artifactShareIsActive(item models.WorkspaceArtifactShare, now time.Time) bool {
	if strings.TrimSpace(item.RevokedAt) != "" {
		return false
	}
	if strings.TrimSpace(item.ExpiresAt) == "" {
		return true
	}
	expiresAt, ok := parseRFC3339(item.ExpiresAt)
	if !ok {
		return false
	}
	return now.Before(expiresAt) || now.Equal(expiresAt)
}

func resolveArtifactShareExpiry(payload createArtifactShareRequest) (string, error) {
	if strings.TrimSpace(payload.ExpiresAt) != "" {
		parsed, ok := parseRFC3339(strings.TrimSpace(payload.ExpiresAt))
		if !ok {
			return "", errors.New("expiresAt must be RFC3339")
		}
		if parsed.Before(time.Now().UTC()) {
			return "", errors.New("expiresAt must be in the future")
		}
		return parsed.Format(time.RFC3339), nil
	}
	days := payload.ExpiresInDays
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		return "", errors.New("expiresInDays must be between 1 and 30")
	}
	return time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour).Format(time.RFC3339), nil
}

func randomArtifactShareToken() (string, error) {
	buffer := make([]byte, 10)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return "art-" + hex.EncodeToString(buffer), nil
}

func parseArtifactActionID(path string, prefix string, suffix string) (int, bool) {
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return 0, false
	}
	trimmed := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	value, err := strconv.Atoi(strings.Trim(trimmed, "/"))
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func (r *Router) nextWorkspaceArtifactFavoriteID() int {
	next := 1
	for _, item := range r.data.WorkspaceArtifactFavorites {
		if item.ID >= next {
			next = item.ID + 1
		}
	}
	return next
}

func (r *Router) nextWorkspaceArtifactShareID() int {
	next := 1
	for _, item := range r.data.WorkspaceArtifactShares {
		if item.ID >= next {
			next = item.ID + 1
		}
	}
	return next
}
