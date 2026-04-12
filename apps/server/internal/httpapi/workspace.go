package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
)

type createWorkspaceSessionRequest struct {
	Title        string `json:"title"`
	WorkspaceURL string `json:"workspaceUrl"`
}

type updateWorkspaceSessionRequest struct {
	Status string `json:"status"`
}

type createWorkspaceArtifactRequest struct {
	Title      string `json:"title"`
	Kind       string `json:"kind"`
	SourceURL  string `json:"sourceUrl"`
	PreviewURL string `json:"previewUrl"`
	MessageID  int    `json:"messageId"`
}

type createWorkspaceMessageRequest struct {
	Role     string `json:"role"`
	Status   string `json:"status"`
	Content  string `json:"content"`
	Dispatch bool   `json:"dispatch"`
}

type workspaceSessionSummary struct {
	ID                 int    `json:"id"`
	SessionNo          string `json:"sessionNo"`
	TenantID           int    `json:"tenantId"`
	TenantName         string `json:"tenantName,omitempty"`
	InstanceID         int    `json:"instanceId"`
	InstanceName       string `json:"instanceName,omitempty"`
	Title              string `json:"title"`
	Status             string `json:"status"`
	WorkspaceURL       string `json:"workspaceUrl,omitempty"`
	ProtocolVersion    string `json:"protocolVersion,omitempty"`
	LastOpenedAt       string `json:"lastOpenedAt,omitempty"`
	LastArtifactAt     string `json:"lastArtifactAt,omitempty"`
	LastSyncedAt       string `json:"lastSyncedAt,omitempty"`
	LastMessageAt      string `json:"lastMessageAt,omitempty"`
	LastMessagePreview string `json:"lastMessagePreview,omitempty"`
	LastMessageRole    string `json:"lastMessageRole,omitempty"`
	LastMessageStatus  string `json:"lastMessageStatus,omitempty"`
	MessageCount       int    `json:"messageCount"`
	ArtifactCount      int    `json:"artifactCount"`
	HasArtifacts       bool   `json:"hasArtifacts"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type workspaceMessageDispatchOutcome struct {
	Message   models.WorkspaceMessage       `json:"message"`
	Reply     *models.WorkspaceMessage      `json:"reply,omitempty"`
	Dispatch  workspaceBridgeDispatchResult `json:"dispatch"`
	Artifacts []models.WorkspaceArtifact    `json:"artifacts,omitempty"`
}

type workspaceSessionFilters struct {
	InstanceID   int
	TenantID     int
	Query        string
	Status       string
	HasArtifacts *bool
	Limit        int
}

type workspaceMessageFilters struct {
	Query    string
	Role     string
	BeforeID int
	Limit    int
}

type workspaceSessionDetailWindow struct {
	MessageLimit    int
	EventLimit      int
	AnchorMessageID int
	AnchorTraceID   string
}

type workspaceRequestError struct {
	StatusCode int
	Message    string
}

func (e workspaceRequestError) Error() string {
	return e.Message
}

func (r *Router) handlePortalWorkspaceSessions(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/workspace/sessions")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.listWorkspaceSessions(w, req, instanceID, "portal")
	case http.MethodPost:
		r.createWorkspaceSession(w, req, instanceID, "portal")
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleAdminWorkspaceSessions(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/admin/instances/", "/workspace/sessions")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.listWorkspaceSessions(w, req, instanceID, "admin")
	case http.MethodPost:
		r.createWorkspaceSession(w, req, instanceID, "admin")
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handlePortalWorkspaceSessionIndex(w http.ResponseWriter, req *http.Request) {
	r.listWorkspaceSessionIndex(w, req, "portal")
}

func (r *Router) handleAdminWorkspaceSessionIndex(w http.ResponseWriter, req *http.Request) {
	r.listWorkspaceSessionIndex(w, req, "admin")
}

func (r *Router) handlePortalWorkspaceSessionDetail(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/sessions/")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.getWorkspaceSessionDetail(w, req, sessionID, "portal")
}

func (r *Router) handleAdminWorkspaceSessionDetail(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/sessions/")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.getWorkspaceSessionDetail(w, req, sessionID, "admin")
}

func (r *Router) handlePortalWorkspaceSessionStatus(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/sessions/", "/status")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.updateWorkspaceSessionStatus(w, req, sessionID, "portal")
}

func (r *Router) handleAdminWorkspaceSessionStatus(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/sessions/", "/status")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.updateWorkspaceSessionStatus(w, req, sessionID, "admin")
}

func (r *Router) handlePortalWorkspaceArtifacts(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/sessions/", "/artifacts")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.createWorkspaceArtifact(w, req, sessionID, "portal")
}

func (r *Router) handleAdminWorkspaceArtifacts(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/sessions/", "/artifacts")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.createWorkspaceArtifact(w, req, sessionID, "admin")
}

func (r *Router) handlePortalWorkspaceMessages(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/sessions/", "/messages")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.listWorkspaceMessages(w, req, sessionID, "portal")
	case http.MethodPost:
		r.createWorkspaceMessage(w, req, sessionID, "portal")
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleAdminWorkspaceMessages(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/sessions/", "/messages")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.listWorkspaceMessages(w, req, sessionID, "admin")
	case http.MethodPost:
		r.createWorkspaceMessage(w, req, sessionID, "admin")
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handlePortalWorkspaceMessageRetry(w http.ResponseWriter, req *http.Request) {
	messageID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/messages/", "/retry")
	if !ok {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}
	r.retryWorkspaceMessage(w, req, messageID, "portal")
}

func (r *Router) handleAdminWorkspaceMessageRetry(w http.ResponseWriter, req *http.Request) {
	messageID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/messages/", "/retry")
	if !ok {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}
	r.retryWorkspaceMessage(w, req, messageID, "admin")
}

func (r *Router) handlePortalWorkspaceMessageStream(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/sessions/", "/messages/stream")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.streamWorkspaceMessage(w, req, sessionID, "portal")
}

func (r *Router) handleAdminWorkspaceMessageStream(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/sessions/", "/messages/stream")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.streamWorkspaceMessage(w, req, sessionID, "admin")
}

func (r *Router) handlePortalWorkspaceBridgeHealth(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/workspace/bridge-health")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}
	r.checkWorkspaceBridgeHealth(w, req, instanceID, "portal")
}

func (r *Router) handleAdminWorkspaceBridgeHealth(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/admin/instances/", "/workspace/bridge-health")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}
	r.checkWorkspaceBridgeHealth(w, req, instanceID, "admin")
}

func (r *Router) handleWorkspaceBridgeReport(w http.ResponseWriter, req *http.Request) {
	if !r.authorizeWorkspaceBridgeCallback(req) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload workspaceBridgeReportRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.SessionNo) == "" {
		http.Error(w, "sessionNo is required", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	session, ok := r.findWorkspaceSessionByNo(strings.TrimSpace(payload.SessionNo))
	r.mu.RUnlock()
	if !ok {
		http.NotFound(w, req)
		return
	}

	messages := append([]workspaceBridgeMessage(nil), payload.Messages...)
	if payload.Error != nil {
		messages = append(messages, workspaceBridgeMessage{
			Role:         "system",
			Status:       "failed",
			Content:      firstNonEmpty(payload.Error.Message, payload.Error.Code),
			CreatedAt:    payload.SyncedAt,
			TraceID:      payload.TraceID,
			ErrorCode:    strings.TrimSpace(payload.Error.Code),
			ErrorMessage: strings.TrimSpace(payload.Error.Message),
		})
	}

	messagesSynced, artifactsSynced, err := r.persistWorkspaceBridgePayload(session, messages, payload.Artifacts, workspaceMessageOriginBridgeReport, workspaceArtifactOriginBridgeReport, payload.TraceID, payload.SyncedAt)
	if err != nil {
		http.Error(w, "persist workspace bridge report failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"accepted":        true,
		"requestId":       defaultString(strings.TrimSpace(payload.RequestID), strings.TrimSpace(req.Header.Get(workspaceBridgeRequestIDHeader))),
		"messagesSynced":  messagesSynced,
		"artifactsSynced": artifactsSynced,
	})
}

func (r *Router) listWorkspaceSessionIndex(w http.ResponseWriter, req *http.Request, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	filters, err := parseWorkspaceSessionFilters(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if scope == "portal" {
		filters.TenantID = actor.TenantID
	}

	r.mu.RLock()
	items := r.collectWorkspaceSessionSummariesLocked(actor, filters, scope)
	r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) listWorkspaceSessions(w http.ResponseWriter, req *http.Request, instanceID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	filters, err := parseWorkspaceSessionFilters(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filters.InstanceID = instanceID
	if scope == "portal" {
		filters.TenantID = actor.TenantID
	}

	r.mu.RLock()
	instance, found := r.findInstance(instanceID)
	if !found || !r.canAccessWorkspaceInstance(actor, instance, scope) {
		r.mu.RUnlock()
		http.NotFound(w, req)
		return
	}
	items := r.collectWorkspaceSessionSummariesLocked(actor, filters, scope)
	r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) createWorkspaceSession(w http.ResponseWriter, req *http.Request, instanceID int, scope string) {
	var payload createWorkspaceSessionRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	workspaceURL, err := validateExternalHTTPURL(payload.WorkspaceURL)
	if err != nil {
		http.Error(w, "workspaceUrl must be a valid absolute http/https URL", http.StatusBadRequest)
		return
	}

	r.mu.Lock()
	instance, found := r.findInstance(instanceID)
	if !found || !r.canAccessWorkspaceInstance(actor, instance, scope) {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	now := nowRFC3339()
	sessionID := r.nextWorkspaceSessionID()
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		title = fmt.Sprintf("%s 对话 %s", instance.Name, time.Now().UTC().Format("2006-01-02 15:04"))
	}
	if workspaceURL == "" {
		if entry := primaryWorkspaceAccess(r.filterAccessByInstance(instanceID)); entry != nil {
			workspaceURL = entry.URL
		}
	}
	workspaceURL, err = validateExternalHTTPURL(workspaceURL)
	if err != nil {
		r.mu.Unlock()
		http.Error(w, "workspaceUrl must be a valid absolute http/https URL", http.StatusBadRequest)
		return
	}

	session := models.WorkspaceSession{
		ID:              sessionID,
		SessionNo:       fmt.Sprintf("WS-%s-%03d", time.Now().UTC().Format("20060102"), sessionID),
		TenantID:        instance.TenantID,
		InstanceID:      instanceID,
		Title:           title,
		Status:          "active",
		WorkspaceURL:    workspaceURL,
		ProtocolVersion: workspaceBridgeProtocolVersion,
		LastOpenedAt:    now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	r.data.WorkspaceSessions = append([]models.WorkspaceSession{session}, r.data.WorkspaceSessions...)
	summary := r.buildWorkspaceSessionSummaryLocked(session)
	r.appendWorkspaceAuditLocked(actor, session.TenantID, session.InstanceID, "workspace.session.create", "success", map[string]string{
		"sessionId": strconv.Itoa(session.ID),
		"sessionNo": session.SessionNo,
	})
	r.mu.Unlock()

	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist workspace session failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"session": summary})
}

func (r *Router) checkWorkspaceBridgeHealth(w http.ResponseWriter, req *http.Request, instanceID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	r.mu.RLock()
	instance, found := r.findInstance(instanceID)
	if !found || !r.canAccessWorkspaceInstance(actor, instance, scope) {
		r.mu.RUnlock()
		http.NotFound(w, req)
		return
	}

	workspaceURL := ""
	if session := latestWorkspaceSessionForInstance(r.filterWorkspaceSessionsByInstance(instanceID)); session != nil {
		workspaceURL = session.WorkspaceURL
	}
	if workspaceURL == "" {
		if entry := primaryWorkspaceAccess(r.filterAccessByInstance(instanceID)); entry != nil {
			workspaceURL = entry.URL
		}
	}
	r.mu.RUnlock()

	result := r.checkWorkspaceBridge(req.Context(), workspaceURL)
	status := http.StatusOK
	if !result.OK {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]any{"bridge": result})
}

func (r *Router) getWorkspaceSessionDetail(w http.ResponseWriter, req *http.Request, sessionID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	window, err := parseWorkspaceSessionDetailWindow(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	session, instance, err := r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	if err != nil {
		r.mu.RUnlock()
		writeWorkspaceRequestError(w, err)
		return
	}
	r.mu.RUnlock()

	bridgeSync := r.syncWorkspaceSessionFromBridge(req.Context(), session)

	r.mu.RLock()
	session, instance, err = r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	if err != nil {
		r.mu.RUnlock()
		writeWorkspaceRequestError(w, err)
		return
	}
	summary := r.buildWorkspaceSessionSummaryLocked(session)
	artifacts := append([]models.WorkspaceArtifact(nil), r.filterWorkspaceArtifactsBySession(sessionID)...)
	allMessages := append([]models.WorkspaceMessage(nil), r.filterWorkspaceMessagesBySession(sessionID)...)
	allEvents := append([]models.WorkspaceMessageEvent(nil), r.filterWorkspaceMessageEventsBySession(sessionID)...)
	r.mu.RUnlock()

	messages, messagesHasMore := trimWorkspaceMessagesForDetail(allMessages, window)
	events, eventsHasMore := trimWorkspaceEventsForDetail(allEvents, window)

	if err := r.recordWorkspaceAudit(actor, session.TenantID, session.InstanceID, "workspace.session.read", "success", map[string]string{
		"sessionId":     strconv.Itoa(session.ID),
		"artifactCount": strconv.Itoa(len(artifacts)),
		"messageCount":  strconv.Itoa(len(messages)),
	}); err != nil {
		http.Error(w, "persist workspace audit failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session":         summary,
		"instance":        instance,
		"artifacts":       artifacts,
		"messages":        messages,
		"messagesHasMore": messagesHasMore,
		"events":          events,
		"eventsHasMore":   eventsHasMore,
		"bridgeSync":      bridgeSync,
	})
}

func (r *Router) updateWorkspaceSessionStatus(w http.ResponseWriter, req *http.Request, sessionID int, scope string) {
	var payload updateWorkspaceSessionRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	status, err := normalizeWorkspaceSessionStatus(payload.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	changed := false
	sessionMutation := models.WorkspaceSession{}
	summary := workspaceSessionSummary{}

	r.mu.Lock()
	session, _, accessErr := r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	if accessErr != nil {
		r.mu.Unlock()
		writeWorkspaceRequestError(w, accessErr)
		return
	}

	previousStatus := session.Status
	if index := r.findWorkspaceSessionIndex(sessionID); index >= 0 {
		if r.data.WorkspaceSessions[index].Status != status {
			changed = true
			r.data.WorkspaceSessions[index].Status = status
			r.data.WorkspaceSessions[index].UpdatedAt = nowRFC3339()
			sessionMutation = r.data.WorkspaceSessions[index]
			action := "workspace.session.activate"
			if status == "archived" {
				action = "workspace.session.archive"
			}
			r.appendWorkspaceAuditLocked(actor, session.TenantID, session.InstanceID, action, "success", map[string]string{
				"sessionId": strconv.Itoa(session.ID),
				"from":      previousStatus,
				"to":        status,
			})
		}
		summary = r.buildWorkspaceSessionSummaryLocked(r.data.WorkspaceSessions[index])
	} else {
		summary = r.buildWorkspaceSessionSummaryLocked(session)
	}
	r.mu.Unlock()

	if changed {
		if err := r.persistWorkspaceMutationAndPublish(corestore.WorkspaceMutation{
			Sessions: []models.WorkspaceSession{sessionMutation},
		}); err != nil {
			http.Error(w, "persist workspace session failed", http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"session": summary})
}

func (r *Router) createWorkspaceArtifact(w http.ResponseWriter, req *http.Request, sessionID int, scope string) {
	var payload createWorkspaceArtifactRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Title) == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	r.mu.RLock()
	session, instance, err := r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	if err != nil {
		r.mu.RUnlock()
		writeWorkspaceRequestError(w, err)
		return
	}
	if err := ensureWorkspaceSessionWritable(session); err != nil {
		r.mu.RUnlock()
		writeWorkspaceRequestError(w, err)
		return
	}

	sourceURL, previewURL, urlErr := r.normalizeAndValidateWorkspaceArtifactURLs(session, instance, payload.SourceURL, payload.PreviewURL)
	if urlErr != nil {
		r.mu.RUnlock()
		if validationErr, ok := urlErr.(workspaceArtifactValidationError); ok {
			http.Error(w, validationErr.message, validationErr.status)
			return
		}
		http.Error(w, urlErr.Error(), http.StatusBadRequest)
		return
	}
	r.mu.RUnlock()

	archiveResult, err := r.archiveArtifactFromSource(req.Context(), sourceURL, session.ID, session.TenantID, session.InstanceID, payload.Title, payload.Kind)
	if err != nil {
		_ = r.recordWorkspaceAudit(actor, session.TenantID, session.InstanceID, "workspace.artifact.create", "archive_failed", map[string]string{
			"sessionId": strconv.Itoa(session.ID),
			"kind":      strings.TrimSpace(payload.Kind),
			"error":     err.Error(),
		})
		http.Error(w, fmt.Sprintf("archive workspace artifact failed: %v", err), http.StatusBadGateway)
		return
	}

	r.mu.Lock()
	now := nowRFC3339()
	artifact := models.WorkspaceArtifact{
		ID:             r.nextWorkspaceArtifactID(),
		SessionID:      sessionID,
		MessageID:      payload.MessageID,
		TenantID:       session.TenantID,
		InstanceID:     session.InstanceID,
		Title:          strings.TrimSpace(payload.Title),
		Kind:           defaultString(strings.TrimSpace(payload.Kind), detectArtifactKindFromURL(sourceURL)),
		Origin:         workspaceArtifactOriginManual,
		SourceURL:      sourceURL,
		PreviewURL:     previewURL,
		ArchiveStatus:  defaultString(archiveResult.ArchiveStatus, "not_configured"),
		ContentType:    archiveResult.ContentType,
		SizeBytes:      archiveResult.SizeBytes,
		StorageBucket:  archiveResult.StorageBucket,
		StorageKey:     archiveResult.StorageKey,
		Filename:       archiveResult.Filename,
		ChecksumSHA256: archiveResult.ChecksumSHA256,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	r.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{artifact}, r.data.WorkspaceArtifacts...)
	event, eventErr := r.appendWorkspaceMessageEventLocked(session, artifact.MessageID, workspaceEventArtifactCreated, workspaceEventPayload{
		Artifact: &artifact,
	}, artifact.Origin, "")
	if eventErr != nil {
		r.mu.Unlock()
		http.Error(w, "encode workspace artifact event failed", http.StatusInternalServerError)
		return
	}
	sessionMutation := session
	if index := r.findWorkspaceSessionIndex(sessionID); index >= 0 {
		r.data.WorkspaceSessions[index].LastArtifactAt = now
		r.data.WorkspaceSessions[index].UpdatedAt = now
		sessionMutation = r.data.WorkspaceSessions[index]
	}
	r.appendWorkspaceAuditLocked(actor, session.TenantID, session.InstanceID, "workspace.artifact.create", "success", map[string]string{
		"artifactId":    strconv.Itoa(artifact.ID),
		"sessionId":     strconv.Itoa(session.ID),
		"kind":          artifact.Kind,
		"archiveStatus": artifact.ArchiveStatus,
	})
	r.mu.Unlock()

	if err := r.persistWorkspaceMutationAndPublish(corestore.WorkspaceMutation{
		Sessions:  []models.WorkspaceSession{sessionMutation},
		Artifacts: []models.WorkspaceArtifact{artifact},
		Events:    []models.WorkspaceMessageEvent{event},
	}); err != nil {
		http.Error(w, "persist workspace artifact failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"artifact": artifact})
}

func (r *Router) listWorkspaceMessages(w http.ResponseWriter, req *http.Request, sessionID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	filters, err := parseWorkspaceMessageFilters(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	_, _, accessErr := r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	if accessErr != nil {
		r.mu.RUnlock()
		writeWorkspaceRequestError(w, accessErr)
		return
	}
	items, hasMore := applyWorkspaceMessageFilters(r.filterWorkspaceMessagesBySession(sessionID), filters)
	session, _, sessionErr := r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	r.mu.RUnlock()
	if sessionErr != nil {
		writeWorkspaceRequestError(w, sessionErr)
		return
	}

	if err := r.recordWorkspaceAudit(actor, session.TenantID, session.InstanceID, "workspace.message.list", "success", map[string]string{
		"sessionId":    strconv.Itoa(session.ID),
		"messageCount": strconv.Itoa(len(items)),
	}); err != nil {
		http.Error(w, "persist workspace audit failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":   items,
		"hasMore": hasMore,
	})
}

func (r *Router) createWorkspaceMessage(w http.ResponseWriter, req *http.Request, sessionID int, scope string) {
	var payload createWorkspaceMessageRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	role, err := normalizeWorkspaceMessageRole(payload.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	status, err := normalizeWorkspaceMessageStatus(payload.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	if payload.Dispatch && role != "user" {
		http.Error(w, "dispatch only supports user messages", http.StatusBadRequest)
		return
	}

	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	session, instance, messageRecord, accessErr := r.recordWorkspaceMessage(actor, sessionID, role, status, payload.Content, scope)
	if accessErr != nil {
		writeWorkspaceRequestError(w, accessErr)
		return
	}

	response := map[string]any{
		"message": messageRecord,
	}
	if role == "user" && payload.Dispatch {
		outcome, err := r.finalizeWorkspaceUserDispatch(req, actor, session, instance, messageRecord)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response["message"] = outcome.Message
		response["dispatch"] = outcome.Dispatch
		if outcome.Reply != nil {
			response["reply"] = outcome.Reply
		}
		if len(outcome.Artifacts) > 0 {
			response["artifacts"] = outcome.Artifacts
		}
	}

	writeJSON(w, http.StatusCreated, response)
}

func (r *Router) retryWorkspaceMessage(w http.ResponseWriter, req *http.Request, messageID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	r.mu.RLock()
	sourceMessage, session, instance, accessErr := r.loadWorkspaceMessageContextLocked(actor, messageID, scope)
	r.mu.RUnlock()
	if accessErr != nil {
		writeWorkspaceRequestError(w, accessErr)
		return
	}
	if sourceMessage.Role != "user" {
		http.Error(w, "only user messages can be retried", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(sourceMessage.Content) == "" {
		http.Error(w, "message content is empty", http.StatusBadRequest)
		return
	}

	session, instance, retryMessage, retryErr := r.recordWorkspaceMessage(actor, session.ID, "user", "recorded", sourceMessage.Content, scope)
	if retryErr != nil {
		writeWorkspaceRequestError(w, retryErr)
		return
	}

	outcome, err := r.finalizeWorkspaceUserDispatch(req, actor, session, instance, retryMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.recordWorkspaceAudit(actor, session.TenantID, session.InstanceID, "workspace.message.retry", "success", map[string]string{
		"sessionId":       strconv.Itoa(session.ID),
		"sourceMessageId": strconv.Itoa(sourceMessage.ID),
		"retryMessageId":  strconv.Itoa(outcome.Message.ID),
		"artifactCount":   strconv.Itoa(len(outcome.Artifacts)),
	}); err != nil {
		http.Error(w, "persist workspace audit failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":   outcome.Message,
		"dispatch":  outcome.Dispatch,
		"reply":     outcome.Reply,
		"artifacts": outcome.Artifacts,
	})
}

func (r *Router) streamWorkspaceMessage(w http.ResponseWriter, req *http.Request, sessionID int, scope string) {
	var payload createWorkspaceMessageRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}

	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	session, instance, messageRecord, accessErr := r.recordWorkspaceMessage(actor, sessionID, "user", "recorded", payload.Content, scope)
	if accessErr != nil {
		writeWorkspaceRequestError(w, accessErr)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	_ = writeWorkspaceStreamEvent(w, flusher, "message", map[string]any{"message": messageRecord})

	outcome, err := r.finalizeWorkspaceUserDispatch(req, actor, session, instance, messageRecord)
	if err != nil {
		_ = writeWorkspaceStreamEvent(w, flusher, "error", map[string]any{"error": err.Error()})
		_ = writeWorkspaceStreamEvent(w, flusher, "done", map[string]any{"messageId": messageRecord.ID})
		return
	}

	_ = writeWorkspaceStreamEvent(w, flusher, "status", map[string]any{"dispatch": outcome.Dispatch})
	if outcome.Dispatch.Error != "" {
		_ = writeWorkspaceStreamEvent(w, flusher, "error", map[string]any{"error": outcome.Dispatch.Error})
	}
	if outcome.Reply != nil {
		for _, delta := range chunkWorkspaceReply(outcome.Reply.Content) {
			_ = writeWorkspaceStreamEvent(w, flusher, "chunk", map[string]any{"delta": delta})
		}
		_ = writeWorkspaceStreamEvent(w, flusher, "reply", map[string]any{"message": outcome.Reply})
	}
	for _, artifact := range outcome.Artifacts {
		_ = writeWorkspaceStreamEvent(w, flusher, "artifact", map[string]any{"artifact": artifact})
	}
	_ = writeWorkspaceStreamEvent(w, flusher, "done", map[string]any{"messageId": outcome.Message.ID})
}

func (r *Router) recordWorkspaceMessage(
	actor workspaceActor,
	sessionID int,
	role string,
	status string,
	content string,
	scope string,
) (models.WorkspaceSession, models.Instance, models.WorkspaceMessage, error) {
	r.mu.Lock()
	session, instance, accessErr := r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	if accessErr != nil {
		r.mu.Unlock()
		return models.WorkspaceSession{}, models.Instance{}, models.WorkspaceMessage{}, accessErr
	}
	if err := ensureWorkspaceSessionWritable(session); err != nil {
		r.mu.Unlock()
		return models.WorkspaceSession{}, models.Instance{}, models.WorkspaceMessage{}, err
	}

	now := nowRFC3339()
	message := models.WorkspaceMessage{
		ID:              r.nextWorkspaceMessageID(),
		SessionID:       sessionID,
		TenantID:        session.TenantID,
		InstanceID:      session.InstanceID,
		Role:            role,
		Status:          status,
		Origin:          workspaceMessageOriginPlatform,
		DeliveryAttempt: 0,
		Content:         strings.TrimSpace(content),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	r.data.WorkspaceMessages = append(r.data.WorkspaceMessages, message)
	event, eventErr := r.appendWorkspaceMessageEventLocked(session, message.ID, workspaceEventMessageCreated, workspaceEventPayload{
		Message: &message,
	}, message.Origin, "")
	if eventErr != nil {
		r.mu.Unlock()
		return models.WorkspaceSession{}, models.Instance{}, models.WorkspaceMessage{}, workspaceRequestError{
			StatusCode: http.StatusInternalServerError,
			Message:    "encode workspace message event failed",
		}
	}
	sessionMutation := session
	if index := r.findWorkspaceSessionIndex(sessionID); index >= 0 {
		r.data.WorkspaceSessions[index].LastOpenedAt = now
		r.data.WorkspaceSessions[index].UpdatedAt = now
		session = r.data.WorkspaceSessions[index]
		sessionMutation = r.data.WorkspaceSessions[index]
	}
	r.appendWorkspaceAuditLocked(actor, session.TenantID, session.InstanceID, "workspace.message.create", "success", map[string]string{
		"messageId": strconv.Itoa(message.ID),
		"sessionId": strconv.Itoa(session.ID),
		"role":      message.Role,
	})
	r.mu.Unlock()

	if err := r.persistWorkspaceMutationAndPublish(corestore.WorkspaceMutation{
		Sessions: []models.WorkspaceSession{sessionMutation},
		Messages: []models.WorkspaceMessage{message},
		Events:   []models.WorkspaceMessageEvent{event},
	}); err != nil {
		return models.WorkspaceSession{}, models.Instance{}, models.WorkspaceMessage{}, workspaceRequestError{
			StatusCode: http.StatusInternalServerError,
			Message:    "persist workspace message failed",
		}
	}
	return session, instance, message, nil
}

func (r *Router) finalizeWorkspaceUserDispatch(
	req *http.Request,
	actor workspaceActor,
	session models.WorkspaceSession,
	instance models.Instance,
	message models.WorkspaceMessage,
) (workspaceMessageDispatchOutcome, error) {
	dispatch := r.dispatchWorkspaceMessage(req.Context(), session, instance, message)
	artifacts, artifactErr := r.buildWorkspaceArtifactsFromDispatch(req.Context(), session, dispatch)
	if artifactErr != nil {
		dispatch.OK = false
		dispatch.Code = defaultString(strings.TrimSpace(dispatch.Code), "artifact_archive_failed")
		dispatch.Error = defaultString(strings.TrimSpace(dispatch.Error), artifactErr.Error())
	}

	r.mu.Lock()
	updatedMessage := message
	mutation := corestore.WorkspaceMutation{}
	if index := r.findWorkspaceMessageIndex(message.ID); index >= 0 {
		r.data.WorkspaceMessages[index].TraceID = firstNonEmpty(strings.TrimSpace(dispatch.TraceID), r.data.WorkspaceMessages[index].TraceID)
		r.data.WorkspaceMessages[index].DeliveryAttempt = dispatch.Attempt
		if dispatch.OK {
			r.data.WorkspaceMessages[index].Status = "sent"
			r.data.WorkspaceMessages[index].DeliveredAt = nowRFC3339()
			r.data.WorkspaceMessages[index].ErrorCode = ""
			r.data.WorkspaceMessages[index].ErrorMessage = ""
		} else {
			r.data.WorkspaceMessages[index].Status = "failed"
			r.data.WorkspaceMessages[index].ErrorCode = defaultString(strings.TrimSpace(dispatch.Code), "bridge_dispatch_failed")
			r.data.WorkspaceMessages[index].ErrorMessage = defaultString(strings.TrimSpace(dispatch.Error), "发送到龙虾实例失败")
			r.data.WorkspaceMessages[index].DeliveredAt = ""
		}
		r.data.WorkspaceMessages[index].UpdatedAt = nowRFC3339()
		updatedMessage = r.data.WorkspaceMessages[index]
		mutation.Messages = append(mutation.Messages, updatedMessage)
	}
	dispatchEvent, dispatchEventErr := r.appendWorkspaceMessageEventLocked(session, updatedMessage.ID, workspaceEventDispatchStatus, workspaceEventPayload{
		Message:  &updatedMessage,
		Dispatch: &dispatch,
		Status:   updatedMessage.Status,
	}, workspaceMessageOriginPlatform, dispatch.TraceID)
	if dispatchEventErr != nil {
		r.mu.Unlock()
		return workspaceMessageDispatchOutcome{}, dispatchEventErr
	}
	mutation.Events = append(mutation.Events, dispatchEvent)

	reply := r.buildWorkspaceDispatchReplyLocked(session, message, dispatch)
	var persistedReply *models.WorkspaceMessage
	if reply != nil {
		r.data.WorkspaceMessages = append(r.data.WorkspaceMessages, *reply)
		persistedReply = reply
		mutation.Messages = append(mutation.Messages, *reply)
		replyEventType := workspaceEventMessageCreated
		if persistedReply.Status == "failed" {
			replyEventType = workspaceEventMessageFailed
		} else if persistedReply.Status == "delivered" && strings.TrimSpace(persistedReply.Content) != "" {
			replyEventType = workspaceEventMessageCompleted
		}
		replyEvent, replyEventErr := r.appendWorkspaceMessageEventLocked(session, persistedReply.ID, replyEventType, workspaceEventPayload{
			Message:      persistedReply,
			Content:      persistedReply.Content,
			ErrorMessage: persistedReply.ErrorMessage,
			Status:       persistedReply.Status,
		}, persistedReply.Origin, persistedReply.TraceID)
		if replyEventErr != nil {
			r.mu.Unlock()
			return workspaceMessageDispatchOutcome{}, replyEventErr
		}
		mutation.Events = append(mutation.Events, replyEvent)
	}

	for index := range artifacts {
		artifacts[index].ID = r.nextWorkspaceArtifactID() + index
		if artifacts[index].MessageID == 0 {
			if persistedReply != nil {
				artifacts[index].MessageID = persistedReply.ID
			} else {
				artifacts[index].MessageID = message.ID
			}
		}
	}
	if len(artifacts) > 0 {
		r.data.WorkspaceArtifacts = append(artifacts, r.data.WorkspaceArtifacts...)
		mutation.Artifacts = append(mutation.Artifacts, artifacts...)
		for index := range artifacts {
			artifact := artifacts[index]
			artifactEvent, artifactEventErr := r.appendWorkspaceMessageEventLocked(session, artifact.MessageID, workspaceEventArtifactCreated, workspaceEventPayload{
				Artifact: &artifact,
			}, artifact.Origin, dispatch.TraceID)
			if artifactEventErr != nil {
				r.mu.Unlock()
				return workspaceMessageDispatchOutcome{}, artifactEventErr
			}
			mutation.Events = append(mutation.Events, artifactEvent)
		}
	}
	sessionMutation := session
	if index := r.findWorkspaceSessionIndex(session.ID); index >= 0 {
		now := nowRFC3339()
		r.data.WorkspaceSessions[index].ProtocolVersion = workspaceRealtimeProtocolVersion
		r.data.WorkspaceSessions[index].LastOpenedAt = now
		if len(artifacts) > 0 {
			r.data.WorkspaceSessions[index].LastArtifactAt = now
		}
		r.data.WorkspaceSessions[index].LastSyncedAt = now
		r.data.WorkspaceSessions[index].UpdatedAt = now
		sessionMutation = r.data.WorkspaceSessions[index]
	}
	mutation.Sessions = append(mutation.Sessions, sessionMutation)
	dispatchResult := "success"
	if !dispatch.OK {
		dispatchResult = "failed"
	}
	r.appendWorkspaceAuditLocked(actor, session.TenantID, session.InstanceID, "workspace.message.dispatch", dispatchResult, map[string]string{
		"messageId":     strconv.Itoa(message.ID),
		"sessionId":     strconv.Itoa(session.ID),
		"target":        dispatch.Target,
		"error":         dispatch.Error,
		"artifactCount": strconv.Itoa(len(artifacts)),
	})
	r.mu.Unlock()

	if err := r.persistWorkspaceMutationAndPublish(mutation); err != nil {
		return workspaceMessageDispatchOutcome{}, err
	}
	if dispatch.OK && dispatch.StreamURL != "" && persistedReply != nil {
		go r.consumeWorkspaceBridgeStream(sessionMutation, *persistedReply, dispatch.StreamURL, dispatch.TraceID)
	}
	return workspaceMessageDispatchOutcome{
		Message:   updatedMessage,
		Reply:     persistedReply,
		Dispatch:  dispatch,
		Artifacts: artifacts,
	}, nil
}

func (r *Router) buildWorkspaceDispatchReplyLocked(session models.WorkspaceSession, parent models.WorkspaceMessage, dispatch workspaceBridgeDispatchResult) *models.WorkspaceMessage {
	replyContent := ""
	replyRole := "assistant"
	replyStatus := "delivered"
	replyExternalID := ""
	replyErrorCode := ""
	replyErrorMessage := ""

	var preferredMessage *workspaceBridgeMessage
	for _, item := range dispatch.Messages {
		if strings.TrimSpace(item.Content) == "" && strings.TrimSpace(item.ErrorMessage) == "" && strings.TrimSpace(item.ErrorCode) == "" {
			continue
		}
		candidate := item
		if preferredMessage == nil {
			preferredMessage = &candidate
		}
		if strings.EqualFold(strings.TrimSpace(item.Role), "assistant") {
			preferredMessage = &candidate
			break
		}
	}

	if preferredMessage != nil {
		replyRole = defaultString(strings.TrimSpace(preferredMessage.Role), "assistant")
		replyStatus = defaultString(strings.TrimSpace(preferredMessage.Status), "delivered")
		replyExternalID = strings.TrimSpace(preferredMessage.ID)
		replyErrorCode = strings.TrimSpace(preferredMessage.ErrorCode)
		replyErrorMessage = strings.TrimSpace(preferredMessage.ErrorMessage)
		replyContent = strings.TrimSpace(preferredMessage.Content)
		if replyContent == "" {
			replyContent = firstNonEmpty(replyErrorMessage, replyErrorCode)
		}
	} else if dispatch.OK && strings.TrimSpace(dispatch.StreamURL) == "" {
		replyRole = "system"
		replyStatus = "recorded"
		if len(dispatch.Artifacts) > 0 {
			replyContent = fmt.Sprintf("龙虾已接收消息，并回传 %d 个产物。", len(dispatch.Artifacts))
		} else {
			replyContent = "龙虾已接收消息，当前桥接未返回平台侧文本内容，请继续查看工作台或等待后续产物。"
		}
	} else if dispatch.OK && strings.TrimSpace(dispatch.StreamURL) != "" {
		replyRole = "assistant"
		replyStatus = "streaming"
	} else {
		replyRole = "system"
		replyStatus = "failed"
		replyErrorCode = defaultString(strings.TrimSpace(dispatch.Code), "bridge_dispatch_failed")
		replyErrorMessage = defaultString(strings.TrimSpace(dispatch.Error), "发送到龙虾实例失败")
		replyContent = defaultString(strings.TrimSpace(dispatch.Error), "发送到龙虾实例失败")
	}

	replyContent = strings.TrimSpace(replyContent)
	if replyContent == "" && replyStatus != "streaming" {
		return nil
	}

	now := nowRFC3339()
	deliveredAt := now
	if replyStatus == "streaming" || replyStatus == "recorded" || replyStatus == "failed" {
		deliveredAt = ""
	}
	return &models.WorkspaceMessage{
		ID:              r.nextWorkspaceMessageID(),
		SessionID:       session.ID,
		TenantID:        session.TenantID,
		InstanceID:      session.InstanceID,
		ParentMessageID: parent.ID,
		Role:            replyRole,
		Status:          replyStatus,
		ExternalID:      replyExternalID,
		Origin:          workspaceMessageOriginBridgeResponse,
		TraceID:         dispatch.TraceID,
		ErrorCode:       replyErrorCode,
		ErrorMessage:    replyErrorMessage,
		DeliveryAttempt: dispatch.Attempt,
		Content:         replyContent,
		DeliveredAt:     deliveredAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (r *Router) buildWorkspaceArtifactsFromDispatch(ctx context.Context, session models.WorkspaceSession, dispatch workspaceBridgeDispatchResult) ([]models.WorkspaceArtifact, error) {
	if !dispatch.OK || len(dispatch.Artifacts) == 0 {
		return nil, nil
	}

	now := nowRFC3339()
	items := make([]models.WorkspaceArtifact, 0, len(dispatch.Artifacts))
	accessEntries := r.filterAccessByInstance(session.InstanceID)
	for _, item := range dispatch.Artifacts {
		sourceURL, previewURL, err := r.normalizeAndValidateWorkspaceArtifactURLsWithAccess(session, accessEntries, item.SourceURL, item.PreviewURL)
		if err != nil || sourceURL == "" {
			continue
		}
		kind := defaultString(strings.TrimSpace(item.Kind), detectArtifactKindFromURL(sourceURL))
		archiveResult, err := r.archiveArtifactFromSource(ctx, sourceURL, session.ID, session.TenantID, session.InstanceID, item.Title, kind)
		if err != nil {
			return nil, err
		}
		artifactID := r.nextWorkspaceArtifactID() + len(items)
		items = append(items, models.WorkspaceArtifact{
			ID:             artifactID,
			SessionID:      session.ID,
			TenantID:       session.TenantID,
			InstanceID:     session.InstanceID,
			Title:          defaultString(strings.TrimSpace(item.Title), fmt.Sprintf("龙虾产物 %d", artifactID)),
			Kind:           kind,
			ExternalID:     strings.TrimSpace(item.ID),
			Origin:         workspaceArtifactOriginBridgeResponse,
			SourceURL:      sourceURL,
			PreviewURL:     previewURL,
			ArchiveStatus:  defaultString(archiveResult.ArchiveStatus, "not_configured"),
			ContentType:    archiveResult.ContentType,
			SizeBytes:      archiveResult.SizeBytes,
			StorageBucket:  archiveResult.StorageBucket,
			StorageKey:     archiveResult.StorageKey,
			Filename:       archiveResult.Filename,
			ChecksumSHA256: archiveResult.ChecksumSHA256,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	return items, nil
}

func (r *Router) loadWorkspaceSessionContextLocked(actor workspaceActor, sessionID int, scope string) (models.WorkspaceSession, models.Instance, error) {
	session, ok := r.findWorkspaceSession(sessionID)
	if !ok {
		return models.WorkspaceSession{}, models.Instance{}, workspaceRequestError{
			StatusCode: http.StatusNotFound,
			Message:    "workspace session not found",
		}
	}
	instance, found := r.findInstance(session.InstanceID)
	if !found || !r.canAccessWorkspaceSession(actor, session, instance, scope) {
		return models.WorkspaceSession{}, models.Instance{}, workspaceRequestError{
			StatusCode: http.StatusNotFound,
			Message:    "workspace session not found",
		}
	}
	return session, instance, nil
}

func (r *Router) loadWorkspaceMessageContextLocked(actor workspaceActor, messageID int, scope string) (models.WorkspaceMessage, models.WorkspaceSession, models.Instance, error) {
	message, ok := r.findWorkspaceMessage(messageID)
	if !ok {
		return models.WorkspaceMessage{}, models.WorkspaceSession{}, models.Instance{}, workspaceRequestError{
			StatusCode: http.StatusNotFound,
			Message:    "workspace message not found",
		}
	}

	session, instance, err := r.loadWorkspaceSessionContextLocked(actor, message.SessionID, scope)
	if err != nil {
		return models.WorkspaceMessage{}, models.WorkspaceSession{}, models.Instance{}, err
	}
	if message.SessionID != session.ID || message.InstanceID != session.InstanceID || message.TenantID != session.TenantID {
		return models.WorkspaceMessage{}, models.WorkspaceSession{}, models.Instance{}, workspaceRequestError{
			StatusCode: http.StatusNotFound,
			Message:    "workspace message not found",
		}
	}
	return message, session, instance, nil
}

func (r *Router) collectWorkspaceSessionSummariesLocked(actor workspaceActor, filters workspaceSessionFilters, scope string) []workspaceSessionSummary {
	items := make([]workspaceSessionSummary, 0)
	query := strings.ToLower(strings.TrimSpace(filters.Query))
	for _, session := range r.data.WorkspaceSessions {
		instance, found := r.findInstance(session.InstanceID)
		if !found || !r.canAccessWorkspaceSession(actor, session, instance, scope) {
			continue
		}
		if filters.InstanceID > 0 && session.InstanceID != filters.InstanceID {
			continue
		}
		if filters.TenantID > 0 && session.TenantID != filters.TenantID {
			continue
		}
		if filters.Status != "" && session.Status != filters.Status {
			continue
		}

		summary := r.buildWorkspaceSessionSummaryLocked(session)
		if filters.HasArtifacts != nil && summary.HasArtifacts != *filters.HasArtifacts {
			continue
		}
		if query != "" && !workspaceSessionSummaryMatches(summary, query) {
			continue
		}
		items = append(items, summary)
		if len(items) >= filters.Limit {
			break
		}
	}
	return items
}

func (r *Router) buildWorkspaceSessionSummaryLocked(session models.WorkspaceSession) workspaceSessionSummary {
	summary := workspaceSessionSummary{
		ID:              session.ID,
		SessionNo:       session.SessionNo,
		TenantID:        session.TenantID,
		InstanceID:      session.InstanceID,
		Title:           session.Title,
		Status:          session.Status,
		WorkspaceURL:    session.WorkspaceURL,
		ProtocolVersion: session.ProtocolVersion,
		LastOpenedAt:    session.LastOpenedAt,
		LastArtifactAt:  session.LastArtifactAt,
		LastSyncedAt:    session.LastSyncedAt,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
	}
	if tenant := r.findTenant(session.TenantID); tenant != nil {
		summary.TenantName = tenant.Name
	}
	if instance, found := r.findInstance(session.InstanceID); found {
		summary.InstanceName = instance.Name
	}
	for _, item := range r.data.WorkspaceMessages {
		if item.SessionID != session.ID {
			continue
		}
		summary.MessageCount++
		if item.CreatedAt >= summary.LastMessageAt {
			summary.LastMessageAt = item.CreatedAt
			summary.LastMessagePreview = trimWorkspacePreview(item.Content)
			summary.LastMessageRole = item.Role
			summary.LastMessageStatus = item.Status
		}
	}
	for _, item := range r.data.WorkspaceArtifacts {
		if item.SessionID != session.ID {
			continue
		}
		summary.ArtifactCount++
	}
	summary.HasArtifacts = summary.ArtifactCount > 0
	return summary
}

func (r *Router) filterWorkspaceSessionsByInstance(instanceID int) []models.WorkspaceSession {
	items := make([]models.WorkspaceSession, 0)
	for _, item := range r.data.WorkspaceSessions {
		if item.InstanceID == instanceID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) filterWorkspaceMessagesByInstance(instanceID int) []models.WorkspaceMessage {
	items := make([]models.WorkspaceMessage, 0)
	for _, item := range r.data.WorkspaceMessages {
		if item.InstanceID == instanceID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) filterWorkspaceMessageEventsByInstance(instanceID int) []models.WorkspaceMessageEvent {
	items := make([]models.WorkspaceMessageEvent, 0)
	for _, item := range r.data.WorkspaceMessageEvents {
		if item.InstanceID == instanceID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) findWorkspaceSession(id int) (models.WorkspaceSession, bool) {
	for _, item := range r.data.WorkspaceSessions {
		if item.ID == id {
			return item, true
		}
	}
	return models.WorkspaceSession{}, false
}

func (r *Router) findWorkspaceSessionByNo(sessionNo string) (models.WorkspaceSession, bool) {
	normalized := strings.TrimSpace(sessionNo)
	for _, item := range r.data.WorkspaceSessions {
		if strings.TrimSpace(item.SessionNo) == normalized {
			return item, true
		}
	}
	return models.WorkspaceSession{}, false
}

func (r *Router) findWorkspaceSessionIndex(id int) int {
	for index, item := range r.data.WorkspaceSessions {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (r *Router) filterWorkspaceArtifactsBySession(sessionID int) []models.WorkspaceArtifact {
	items := make([]models.WorkspaceArtifact, 0)
	for _, item := range r.data.WorkspaceArtifacts {
		if item.SessionID == sessionID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) filterWorkspaceMessagesBySession(sessionID int) []models.WorkspaceMessage {
	items := make([]models.WorkspaceMessage, 0)
	for _, item := range r.data.WorkspaceMessages {
		if item.SessionID == sessionID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) nextWorkspaceSessionID() int {
	maxID := 0
	for _, item := range r.data.WorkspaceSessions {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextWorkspaceArtifactID() int {
	maxID := 0
	for _, item := range r.data.WorkspaceArtifacts {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextWorkspaceMessageID() int {
	maxID := 0
	for _, item := range r.data.WorkspaceMessages {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) findWorkspaceMessage(id int) (models.WorkspaceMessage, bool) {
	for _, item := range r.data.WorkspaceMessages {
		if item.ID == id {
			return item, true
		}
	}
	return models.WorkspaceMessage{}, false
}

func (r *Router) findWorkspaceMessageIndex(id int) int {
	for index, item := range r.data.WorkspaceMessages {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (r *Router) filterWorkspaceArtifactLogsByArtifact(artifactID int) []models.WorkspaceArtifactAccessLog {
	items := make([]models.WorkspaceArtifactAccessLog, 0)
	for _, item := range r.data.WorkspaceArtifactLogs {
		if item.ArtifactID == artifactID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) nextWorkspaceArtifactLogID() int {
	maxID := 0
	for _, item := range r.data.WorkspaceArtifactLogs {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) recordWorkspaceArtifactAccess(actor workspaceActor, artifact models.WorkspaceArtifact, action string, scope string, req *http.Request) error {
	now := nowRFC3339()

	r.mu.Lock()
	r.data.WorkspaceArtifactLogs = append([]models.WorkspaceArtifactAccessLog{{
		ID:         r.nextWorkspaceArtifactLogID(),
		ArtifactID: artifact.ID,
		SessionID:  artifact.SessionID,
		TenantID:   artifact.TenantID,
		InstanceID: artifact.InstanceID,
		Action:     action,
		Scope:      scope,
		Actor:      actor.identifier(),
		RemoteAddr: strings.TrimSpace(req.RemoteAddr),
		UserAgent:  strings.TrimSpace(req.UserAgent()),
		CreatedAt:  now,
	}}, r.data.WorkspaceArtifactLogs...)
	r.appendWorkspaceAuditLocked(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact."+action, "success", map[string]string{
		"artifactId": strconv.Itoa(artifact.ID),
		"sessionId":  strconv.Itoa(artifact.SessionID),
		"scope":      scope,
	})
	r.mu.Unlock()

	return r.persistAllData()
}

func primaryWorkspaceAccess(items []models.InstanceAccess) *models.InstanceAccess {
	for _, item := range items {
		if item.EntryType == "web" || item.EntryType == "h5" || item.EntryType == "portal" {
			copy := item
			return &copy
		}
	}
	if len(items) == 0 {
		return nil
	}
	copy := items[0]
	return &copy
}

func latestWorkspaceSessionForInstance(items []models.WorkspaceSession) *models.WorkspaceSession {
	if len(items) == 0 {
		return nil
	}
	copy := items[0]
	return &copy
}

func parseWorkspaceSessionFilters(req *http.Request) (workspaceSessionFilters, error) {
	query := req.URL.Query()
	filters := workspaceSessionFilters{
		Query:  strings.TrimSpace(query.Get("q")),
		Status: strings.TrimSpace(query.Get("status")),
		Limit:  50,
	}
	if value := strings.TrimSpace(query.Get("instanceId")); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id <= 0 {
			return workspaceSessionFilters{}, fmt.Errorf("instanceId must be a positive integer")
		}
		filters.InstanceID = id
	}
	if value := strings.TrimSpace(query.Get("tenantId")); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id <= 0 {
			return workspaceSessionFilters{}, fmt.Errorf("tenantId must be a positive integer")
		}
		filters.TenantID = id
	}
	if value := strings.TrimSpace(query.Get("hasArtifacts")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return workspaceSessionFilters{}, fmt.Errorf("hasArtifacts must be true or false")
		}
		filters.HasArtifacts = &parsed
	}
	if value := strings.TrimSpace(query.Get("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return workspaceSessionFilters{}, fmt.Errorf("limit must be a positive integer")
		}
		if limit > 200 {
			limit = 200
		}
		filters.Limit = limit
	}
	return filters, nil
}

func parseWorkspaceMessageFilters(req *http.Request) (workspaceMessageFilters, error) {
	query := req.URL.Query()
	filters := workspaceMessageFilters{
		Query: strings.TrimSpace(query.Get("q")),
		Role:  strings.TrimSpace(query.Get("role")),
		Limit: 200,
	}
	if value := strings.TrimSpace(query.Get("beforeId")); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id <= 0 {
			return workspaceMessageFilters{}, fmt.Errorf("beforeId must be a positive integer")
		}
		filters.BeforeID = id
	}
	if value := strings.TrimSpace(query.Get("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return workspaceMessageFilters{}, fmt.Errorf("limit must be a positive integer")
		}
		if limit > 500 {
			limit = 500
		}
		filters.Limit = limit
	}
	return filters, nil
}

func parseWorkspaceSessionDetailWindow(req *http.Request) (workspaceSessionDetailWindow, error) {
	query := req.URL.Query()
	window := workspaceSessionDetailWindow{
		MessageLimit: 50,
		EventLimit:   180,
	}
	if value := strings.TrimSpace(query.Get("messageLimit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return workspaceSessionDetailWindow{}, fmt.Errorf("messageLimit must be a positive integer")
		}
		if limit > 500 {
			limit = 500
		}
		window.MessageLimit = limit
	}
	if value := strings.TrimSpace(query.Get("eventLimit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return workspaceSessionDetailWindow{}, fmt.Errorf("eventLimit must be a positive integer")
		}
		if limit > 1000 {
			limit = 1000
		}
		window.EventLimit = limit
	}
	if value := strings.TrimSpace(query.Get("anchorMessageId")); value != "" {
		anchorID, err := strconv.Atoi(value)
		if err != nil || anchorID <= 0 {
			return workspaceSessionDetailWindow{}, fmt.Errorf("anchorMessageId must be a positive integer")
		}
		window.AnchorMessageID = anchorID
	}
	window.AnchorTraceID = strings.TrimSpace(query.Get("anchorTraceId"))
	return window, nil
}

func applyWorkspaceMessageFilters(items []models.WorkspaceMessage, filters workspaceMessageFilters) ([]models.WorkspaceMessage, bool) {
	query := strings.ToLower(strings.TrimSpace(filters.Query))
	role := strings.ToLower(strings.TrimSpace(filters.Role))
	filtered := make([]models.WorkspaceMessage, 0, len(items))
	for _, item := range items {
		if filters.BeforeID > 0 && item.ID >= filters.BeforeID {
			continue
		}
		if role != "" && strings.ToLower(item.Role) != role {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Content), query) {
			continue
		}
		filtered = append(filtered, item)
	}
	if filters.Limit <= 0 || len(filtered) <= filters.Limit {
		return filtered, false
	}
	return filtered[len(filtered)-filters.Limit:], true
}

func trimWorkspaceMessagesForDetail(items []models.WorkspaceMessage, window workspaceSessionDetailWindow) ([]models.WorkspaceMessage, bool) {
	limit := window.MessageLimit
	if limit <= 0 || len(items) <= limit {
		return items, false
	}
	if window.AnchorMessageID > 0 {
		for index, item := range items {
			if item.ID == window.AnchorMessageID {
				start := maxInt(0, index-limit/2)
				end := start + limit
				if end > len(items) {
					end = len(items)
					start = maxInt(0, end-limit)
				}
				return items[start:end], start > 0 || end < len(items)
			}
		}
	}
	if window.AnchorTraceID != "" {
		filtered := make([]models.WorkspaceMessage, 0)
		for _, item := range items {
			if item.TraceID == window.AnchorTraceID {
				filtered = append(filtered, item)
			}
		}
		if len(filtered) > 0 {
			if len(filtered) <= limit {
				return filtered, len(filtered) < len(items)
			}
			return filtered[len(filtered)-limit:], true
		}
	}
	return items[len(items)-limit:], true
}

func trimWorkspaceEventsForDetail(items []models.WorkspaceMessageEvent, window workspaceSessionDetailWindow) ([]models.WorkspaceMessageEvent, bool) {
	limit := window.EventLimit
	if limit <= 0 || len(items) <= limit {
		return items, false
	}
	if window.AnchorMessageID > 0 {
		for index, item := range items {
			if item.MessageID == window.AnchorMessageID {
				start := maxInt(0, index-limit/2)
				end := start + limit
				if end > len(items) {
					end = len(items)
					start = maxInt(0, end-limit)
				}
				return items[start:end], start > 0 || end < len(items)
			}
		}
	}
	if window.AnchorTraceID != "" {
		filtered := make([]models.WorkspaceMessageEvent, 0)
		for _, item := range items {
			traceID := firstNonEmpty(item.TraceID, extractTraceIDFromEventPayload(item.PayloadJSON))
			if traceID == window.AnchorTraceID {
				filtered = append(filtered, item)
			}
		}
		if len(filtered) > 0 {
			if len(filtered) <= limit {
				return filtered, len(filtered) < len(items)
			}
			return filtered[len(filtered)-limit:], true
		}
	}
	return items[len(items)-limit:], true
}

func normalizeWorkspaceMessageRole(raw string) (string, error) {
	role := strings.ToLower(strings.TrimSpace(raw))
	if role == "" {
		return "user", nil
	}
	switch role {
	case "user", "assistant", "system", "note":
		return role, nil
	default:
		return "", fmt.Errorf("role must be one of user, assistant, system, note")
	}
}

func normalizeWorkspaceMessageStatus(raw string) (string, error) {
	status := strings.ToLower(strings.TrimSpace(raw))
	if status == "" {
		return "recorded", nil
	}
	switch status {
	case "recorded", "sent", "streaming", "delivered", "failed":
		return status, nil
	default:
		return "", fmt.Errorf("status must be one of recorded, sent, streaming, delivered, failed")
	}
}

func normalizeWorkspaceSessionStatus(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "active":
		return "active", nil
	case "archived":
		return "archived", nil
	default:
		return "", fmt.Errorf("workspace session status must be active or archived")
	}
}

func ensureWorkspaceSessionWritable(session models.WorkspaceSession) error {
	if strings.EqualFold(strings.TrimSpace(session.Status), "archived") {
		return workspaceRequestError{
			StatusCode: http.StatusConflict,
			Message:    "workspace session is archived",
		}
	}
	return nil
}

func workspaceSessionSummaryMatches(item workspaceSessionSummary, query string) bool {
	for _, candidate := range []string{
		item.SessionNo,
		item.Title,
		item.TenantName,
		item.InstanceName,
		item.LastMessagePreview,
	} {
		if strings.Contains(strings.ToLower(candidate), query) {
			return true
		}
	}
	return false
}

func trimWorkspacePreview(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= 80 {
		return trimmed
	}
	return string(runes[:80]) + "…"
}

func chunkWorkspaceReply(content string) []string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}
	chunks := make([]string, 0)
	builder := strings.Builder{}
	runeCount := 0
	for _, item := range trimmed {
		builder.WriteRune(item)
		runeCount++
		if item == '\n' || strings.ContainsRune("，。！？；,.!?;", item) || runeCount >= 32 {
			chunks = append(chunks, builder.String())
			builder.Reset()
			runeCount = 0
		}
	}
	if builder.Len() > 0 {
		chunks = append(chunks, builder.String())
	}
	return chunks
}

func writeWorkspaceStreamEvent(w http.ResponseWriter, flusher http.Flusher, event string, payload any) error {
	data := "{}"
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		data = string(raw)
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func writeWorkspaceRequestError(w http.ResponseWriter, err error) {
	if reqErr, ok := err.(workspaceRequestError); ok {
		http.Error(w, reqErr.Message, reqErr.StatusCode)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
