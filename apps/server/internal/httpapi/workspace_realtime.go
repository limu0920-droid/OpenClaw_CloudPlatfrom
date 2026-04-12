package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
)

const (
	workspaceRealtimeProtocolVersion = "openclaw-lobster-bridge/v2"
	workspaceEventRetryMillis        = 1500
	workspaceEventHeartbeatInterval  = 15 * time.Second
)

const (
	workspaceEventMessageCreated   = "message.created"
	workspaceEventMessageUpdated   = "message.updated"
	workspaceEventMessageDelta     = "message.delta"
	workspaceEventMessageCompleted = "message.completed"
	workspaceEventMessageFailed    = "message.failed"
	workspaceEventDispatchStatus   = "dispatch.status"
	workspaceEventArtifactCreated  = "artifact.created"
	workspaceEventReasoningDelta   = "reasoning.delta"
	workspaceEventToolStarted      = "tool.started"
	workspaceEventToolProgress     = "tool.progress"
	workspaceEventToolCompleted    = "tool.completed"
	workspaceEventToolFailed       = "tool.failed"
)

type workspaceEventPayload struct {
	Message      *models.WorkspaceMessage       `json:"message,omitempty"`
	Artifact     *models.WorkspaceArtifact      `json:"artifact,omitempty"`
	Dispatch     *workspaceBridgeDispatchResult `json:"dispatch,omitempty"`
	Delta        string                         `json:"delta,omitempty"`
	Content      string                         `json:"content,omitempty"`
	ToolCallID   string                         `json:"toolCallId,omitempty"`
	ToolName     string                         `json:"toolName,omitempty"`
	Status       string                         `json:"status,omitempty"`
	Detail       string                         `json:"detail,omitempty"`
	Reasoning    string                         `json:"reasoning,omitempty"`
	Error        string                         `json:"error,omitempty"`
	ErrorMessage string                         `json:"errorMessage,omitempty"`
}

type workspaceEventEnvelope struct {
	ID         int             `json:"id"`
	SessionID  int             `json:"sessionId"`
	MessageID  int             `json:"messageId,omitempty"`
	TenantID   int             `json:"tenantId"`
	InstanceID int             `json:"instanceId"`
	EventType  string          `json:"eventType"`
	Origin     string          `json:"origin,omitempty"`
	TraceID    string          `json:"traceId,omitempty"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  string          `json:"createdAt"`
}

type workspaceEventBroker struct {
	mu          sync.Mutex
	subscribers map[int]map[chan models.WorkspaceMessageEvent]struct{}
}

func newWorkspaceEventBroker() *workspaceEventBroker {
	return &workspaceEventBroker{
		subscribers: make(map[int]map[chan models.WorkspaceMessageEvent]struct{}),
	}
}

func (b *workspaceEventBroker) subscribe(sessionID int) (<-chan models.WorkspaceMessageEvent, func()) {
	ch := make(chan models.WorkspaceMessageEvent, 256)
	b.mu.Lock()
	if b.subscribers == nil {
		b.subscribers = make(map[int]map[chan models.WorkspaceMessageEvent]struct{})
	}
	if _, ok := b.subscribers[sessionID]; !ok {
		b.subscribers[sessionID] = make(map[chan models.WorkspaceMessageEvent]struct{})
	}
	b.subscribers[sessionID][ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		sessionSubscribers, ok := b.subscribers[sessionID]
		if !ok {
			return
		}
		if _, exists := sessionSubscribers[ch]; exists {
			delete(sessionSubscribers, ch)
			close(ch)
		}
		if len(sessionSubscribers) == 0 {
			delete(b.subscribers, sessionID)
		}
	}
	return ch, unsubscribe
}

func (b *workspaceEventBroker) publish(event models.WorkspaceMessageEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	sessionSubscribers, ok := b.subscribers[event.SessionID]
	if !ok {
		return
	}
	for ch := range sessionSubscribers {
		select {
		case ch <- event:
		default:
			delete(sessionSubscribers, ch)
			close(ch)
		}
	}
	if len(sessionSubscribers) == 0 {
		delete(b.subscribers, event.SessionID)
	}
}

func (r *Router) workspaceEventBrokerInstance() *workspaceEventBroker {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.workspaceEvents == nil {
		r.workspaceEvents = newWorkspaceEventBroker()
	}
	return r.workspaceEvents
}

func (r *Router) nextWorkspaceMessageEventIDLocked() int {
	maxID := 0
	for _, item := range r.data.WorkspaceMessageEvents {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) filterWorkspaceMessageEventsBySession(sessionID int) []models.WorkspaceMessageEvent {
	items := make([]models.WorkspaceMessageEvent, 0)
	for _, item := range r.data.WorkspaceMessageEvents {
		if item.SessionID == sessionID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) workspaceMessageEventsAfterLocked(sessionID int, afterID int) []models.WorkspaceMessageEvent {
	items := make([]models.WorkspaceMessageEvent, 0)
	for _, item := range r.data.WorkspaceMessageEvents {
		if item.SessionID == sessionID && item.ID > afterID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) appendWorkspaceMessageEventLocked(session models.WorkspaceSession, messageID int, eventType string, payload any, origin string, traceID string) (models.WorkspaceMessageEvent, error) {
	now := nowRFC3339()
	payloadJSON := "{}"
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return models.WorkspaceMessageEvent{}, err
		}
		payloadJSON = string(raw)
	}
	event := models.WorkspaceMessageEvent{
		ID:          r.nextWorkspaceMessageEventIDLocked(),
		SessionID:   session.ID,
		MessageID:   messageID,
		TenantID:    session.TenantID,
		InstanceID:  session.InstanceID,
		EventType:   eventType,
		Origin:      defaultString(strings.TrimSpace(origin), workspaceMessageOriginPlatform),
		TraceID:     strings.TrimSpace(traceID),
		PayloadJSON: payloadJSON,
		CreatedAt:   now,
	}
	r.data.WorkspaceMessageEvents = append(r.data.WorkspaceMessageEvents, event)
	if index := r.findWorkspaceSessionIndex(session.ID); index >= 0 {
		r.data.WorkspaceSessions[index].ProtocolVersion = workspaceRealtimeProtocolVersion
		r.data.WorkspaceSessions[index].LastSyncedAt = now
		r.data.WorkspaceSessions[index].UpdatedAt = now
		session = r.data.WorkspaceSessions[index]
	}
	return event, nil
}

func (r *Router) persistWorkspaceMutationAndPublish(mutation corestore.WorkspaceMutation) error {
	if err := r.persistWorkspaceMutation(mutation); err != nil {
		return err
	}
	broker := r.workspaceEventBrokerInstance()
	for _, event := range mutation.Events {
		broker.publish(event)
	}
	return nil
}

func (r *Router) handlePortalWorkspaceEvents(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/sessions/", "/events")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.streamWorkspaceSessionEvents(w, req, sessionID, "portal")
}

func (r *Router) handleAdminWorkspaceEvents(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/sessions/", "/events")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	r.streamWorkspaceSessionEvents(w, req, sessionID, "admin")
}

func (r *Router) streamWorkspaceSessionEvents(w http.ResponseWriter, req *http.Request, sessionID int, scope string) {
	actor, statusCode, message := r.resolveWorkspaceActor(req, scope)
	if statusCode != 0 {
		http.Error(w, message, statusCode)
		return
	}

	afterID := parseWorkspaceEventCursor(req)

	r.mu.RLock()
	_, _, accessErr := r.loadWorkspaceSessionContextLocked(actor, sessionID, scope)
	if accessErr != nil {
		r.mu.RUnlock()
		writeWorkspaceRequestError(w, accessErr)
		return
	}
	backlog := append([]models.WorkspaceMessageEvent(nil), r.workspaceMessageEventsAfterLocked(sessionID, afterID)...)
	r.mu.RUnlock()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusInternalServerError)
		return
	}

	broker := r.workspaceEventBrokerInstance()
	events, unsubscribe := broker.subscribe(sessionID)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "retry: %d\n\n", workspaceEventRetryMillis); err != nil {
		return
	}
	flusher.Flush()

	lastSentID := afterID
	for _, item := range backlog {
		if err := writeWorkspaceRealtimeEvent(w, flusher, item); err != nil {
			return
		}
		lastSentID = item.ID
	}

	heartbeat := time.NewTicker(workspaceEventHeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-req.Context().Done():
			return
		case item, ok := <-events:
			if !ok {
				return
			}
			if item.ID <= lastSentID {
				continue
			}
			if err := writeWorkspaceRealtimeEvent(w, flusher, item); err != nil {
				return
			}
			lastSentID = item.ID
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func parseWorkspaceEventCursor(req *http.Request) int {
	queryValue := strings.TrimSpace(req.URL.Query().Get("after"))
	if queryValue == "" {
		queryValue = strings.TrimSpace(req.Header.Get("Last-Event-ID"))
	}
	if queryValue == "" {
		return 0
	}
	value, err := strconv.Atoi(queryValue)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func writeWorkspaceRealtimeEvent(w http.ResponseWriter, flusher http.Flusher, item models.WorkspaceMessageEvent) error {
	payload := json.RawMessage([]byte("{}"))
	if strings.TrimSpace(item.PayloadJSON) != "" {
		payload = json.RawMessage(item.PayloadJSON)
	}
	raw, err := json.Marshal(workspaceEventEnvelope{
		ID:         item.ID,
		SessionID:  item.SessionID,
		MessageID:  item.MessageID,
		TenantID:   item.TenantID,
		InstanceID: item.InstanceID,
		EventType:  item.EventType,
		Origin:     item.Origin,
		TraceID:    item.TraceID,
		Payload:    payload,
		CreatedAt:  item.CreatedAt,
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "id: %d\n", item.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", item.EventType); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(raw)); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}
