package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
)

type workspaceBridgeStreamPayload struct {
	Message      *workspaceBridgeMessage  `json:"message,omitempty"`
	Artifact     *workspaceBridgeArtifact `json:"artifact,omitempty"`
	Delta        string                   `json:"delta,omitempty"`
	Content      string                   `json:"content,omitempty"`
	Text         string                   `json:"text,omitempty"`
	Reasoning    string                   `json:"reasoning,omitempty"`
	ToolCallID   string                   `json:"toolCallId,omitempty"`
	ToolName     string                   `json:"toolName,omitempty"`
	Status       string                   `json:"status,omitempty"`
	Detail       string                   `json:"detail,omitempty"`
	Error        string                   `json:"error,omitempty"`
	ErrorMessage string                   `json:"errorMessage,omitempty"`
}

func (r *Router) consumeWorkspaceBridgeStream(session models.WorkspaceSession, assistant models.WorkspaceMessage, streamURL string, traceID string) {
	timeoutSeconds := r.config.WorkspaceBridgeTimeoutSecs * 6
	if timeoutSeconds < 60 {
		timeoutSeconds = 60
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		r.failWorkspaceStreamMessage(session, assistant.ID, traceID, err)
		return
	}
	req.Header.Set("Accept", "text/event-stream")
	r.applyWorkspaceBridgeHeaders(req, "")

	client := r.workspaceBridgeHTTPClient()
	if client.Timeout > 0 {
		clone := *client
		clone.Timeout = 0
		client = &clone
	}
	resp, err := client.Do(req)
	if err != nil {
		r.failWorkspaceStreamMessage(session, assistant.ID, traceID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		r.failWorkspaceStreamMessage(session, assistant.ID, traceID, fmt.Errorf("bridge stream returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))))
		return
	}

	if err := consumeWorkspaceSSE(resp.Body, func(eventType string, data string) error {
		return r.applyWorkspaceBridgeStreamEvent(session, assistant.ID, eventType, data, traceID)
	}); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
		r.failWorkspaceStreamMessage(session, assistant.ID, traceID, err)
		return
	}

	_, _, _ = session, assistant, traceID
	_ = r.syncWorkspaceSessionFromBridge(ctx, session)
}

func consumeWorkspaceSSE(reader io.Reader, handler func(eventType string, data string) error) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	eventType := ""
	dataLines := make([]string, 0, 4)
	flush := func() error {
		if len(dataLines) == 0 {
			eventType = ""
			return nil
		}
		currentEvent := strings.TrimSpace(eventType)
		if currentEvent == "" {
			currentEvent = "message"
		}
		payload := strings.Join(dataLines, "\n")
		eventType = ""
		dataLines = dataLines[:0]
		return handler(currentEvent, payload)
	}

	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if strings.HasPrefix(line, "\ufeff") {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return flush()
}

func (r *Router) applyWorkspaceBridgeStreamEvent(session models.WorkspaceSession, assistantMessageID int, eventType string, rawData string, traceID string) error {
	payload, plainText := decodeWorkspaceBridgeStreamPayload(rawData)
	normalizedEvent := strings.ToLower(strings.TrimSpace(eventType))
	if normalizedEvent == "" {
		normalizedEvent = "message"
	}

	r.mu.Lock()
	sessionIndex := r.findWorkspaceSessionIndex(session.ID)
	messageIndex := r.findWorkspaceMessageIndex(assistantMessageID)
	if sessionIndex < 0 || messageIndex < 0 {
		r.mu.Unlock()
		return fmt.Errorf("workspace stream target not found")
	}

	localSession := r.data.WorkspaceSessions[sessionIndex]
	currentMessage := r.data.WorkspaceMessages[messageIndex]
	mutation := corestore.WorkspaceMutation{}
	now := nowRFC3339()

	writeMessage := func(next models.WorkspaceMessage) {
		r.data.WorkspaceMessages[messageIndex] = next
		currentMessage = next
		mutation.Messages = append(mutation.Messages, next)
	}
	appendEvent := func(kind string, eventPayload workspaceEventPayload) error {
		event, err := r.appendWorkspaceMessageEventLocked(localSession, currentMessage.ID, kind, eventPayload, workspaceMessageOriginBridgeResponse, traceID)
		if err != nil {
			return err
		}
		mutation.Events = append(mutation.Events, event)
		if idx := r.findWorkspaceSessionIndex(localSession.ID); idx >= 0 {
			localSession = r.data.WorkspaceSessions[idx]
		}
		return nil
	}

	switch normalizedEvent {
	case "chunk", "message.delta":
		delta := firstNonEmpty(strings.TrimSpace(payload.Delta), strings.TrimSpace(payload.Content), strings.TrimSpace(payload.Text), strings.TrimSpace(plainText))
		if delta != "" {
			currentMessage.Status = "streaming"
			currentMessage.Origin = workspaceMessageOriginBridgeResponse
			currentMessage.TraceID = firstNonEmpty(currentMessage.TraceID, traceID)
			currentMessage.Content += delta
			currentMessage.UpdatedAt = now
			writeMessage(currentMessage)
			if err := appendEvent(workspaceEventMessageDelta, workspaceEventPayload{
				Message: &currentMessage,
				Delta:   delta,
				Content: currentMessage.Content,
				Status:  currentMessage.Status,
			}); err != nil {
				r.mu.Unlock()
				return err
			}
		}
	case "reply", "message.completed":
		replyMessage := workspaceBridgeMessage{
			Role:         "assistant",
			Status:       "delivered",
			Content:      firstNonEmpty(strings.TrimSpace(payload.Content), strings.TrimSpace(payload.Text), strings.TrimSpace(plainText)),
			ErrorMessage: strings.TrimSpace(payload.ErrorMessage),
		}
		if payload.Message != nil {
			replyMessage = *payload.Message
		}
		if normalized, ok := normalizeWorkspaceBridgeMessage(localSession, replyMessage, workspaceMessageOriginBridgeResponse, traceID); ok {
			normalized.ID = currentMessage.ID
			normalized.ParentMessageID = currentMessage.ParentMessageID
			mergeWorkspaceMessage(&currentMessage, normalized)
		}
		if currentMessage.Status != "failed" {
			currentMessage.Status = "delivered"
			currentMessage.DeliveredAt = defaultString(currentMessage.DeliveredAt, now)
		}
		currentMessage.Origin = workspaceMessageOriginBridgeResponse
		currentMessage.TraceID = firstNonEmpty(currentMessage.TraceID, traceID)
		currentMessage.UpdatedAt = now
		writeMessage(currentMessage)
		if err := appendEvent(workspaceEventMessageCompleted, workspaceEventPayload{
			Message:      &currentMessage,
			Content:      currentMessage.Content,
			ErrorMessage: currentMessage.ErrorMessage,
			Status:       currentMessage.Status,
		}); err != nil {
			r.mu.Unlock()
			return err
		}
	case "reasoning", "reasoning.delta":
		reasoning := firstNonEmpty(strings.TrimSpace(payload.Reasoning), strings.TrimSpace(payload.Delta), strings.TrimSpace(payload.Text), strings.TrimSpace(plainText))
		if reasoning != "" {
			if currentMessage.Status == "recorded" || currentMessage.Status == "sent" {
				currentMessage.Status = "streaming"
				currentMessage.UpdatedAt = now
				writeMessage(currentMessage)
			}
			if err := appendEvent(workspaceEventReasoningDelta, workspaceEventPayload{
				Message:   &currentMessage,
				Reasoning: reasoning,
				Status:    currentMessage.Status,
			}); err != nil {
				r.mu.Unlock()
				return err
			}
		}
	case "tool", "tool.started", "tool.progress", "tool.completed", "tool.failed":
		outEvent := normalizedEvent
		if outEvent == "tool" {
			switch strings.ToLower(strings.TrimSpace(payload.Status)) {
			case "started":
				outEvent = workspaceEventToolStarted
			case "completed", "succeeded", "success":
				outEvent = workspaceEventToolCompleted
			case "failed", "error":
				outEvent = workspaceEventToolFailed
			default:
				outEvent = workspaceEventToolProgress
			}
		}
		if currentMessage.Status == "recorded" || currentMessage.Status == "sent" {
			currentMessage.Status = "streaming"
			currentMessage.UpdatedAt = now
			writeMessage(currentMessage)
		}
		if err := appendEvent(outEvent, workspaceEventPayload{
			Message:    &currentMessage,
			ToolCallID: strings.TrimSpace(payload.ToolCallID),
			ToolName:   strings.TrimSpace(payload.ToolName),
			Status:     firstNonEmpty(strings.TrimSpace(payload.Status), strings.TrimPrefix(outEvent, "tool.")),
			Detail:     firstNonEmpty(strings.TrimSpace(payload.Detail), strings.TrimSpace(plainText)),
		}); err != nil {
			r.mu.Unlock()
			return err
		}
	case "artifact", "artifact.created":
		artifactPayload := workspaceBridgeArtifact{}
		if payload.Artifact != nil {
			artifactPayload = *payload.Artifact
		}
		if normalized, ok := normalizeWorkspaceBridgeArtifact(localSession, artifactPayload, workspaceArtifactOriginBridgeResponse, now); ok {
			if normalized.MessageID == 0 {
				normalized.MessageID = currentMessage.ID
			}
			if existingIndex := r.findWorkspaceArtifactByExternalIDLocked(localSession.ID, normalized.ExternalID); existingIndex < 0 &&
				r.findWorkspaceArtifactDuplicateLocked(localSession.ID, normalized.SourceURL, normalized.Title, normalized.CreatedAt) < 0 {
				normalized.ID = r.nextWorkspaceArtifactID()
				r.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{*normalized}, r.data.WorkspaceArtifacts...)
				mutation.Artifacts = append(mutation.Artifacts, *normalized)
				r.data.WorkspaceSessions[sessionIndex].LastArtifactAt = now
				localSession = r.data.WorkspaceSessions[sessionIndex]
				if err := appendEvent(workspaceEventArtifactCreated, workspaceEventPayload{
					Artifact: normalized,
				}); err != nil {
					r.mu.Unlock()
					return err
				}
			}
		}
	case "error", "message.failed":
		errorMessage := firstNonEmpty(strings.TrimSpace(payload.Error), strings.TrimSpace(payload.ErrorMessage), strings.TrimSpace(payload.Detail), strings.TrimSpace(plainText), "bridge stream failed")
		currentMessage.Status = "failed"
		currentMessage.Origin = workspaceMessageOriginBridgeResponse
		currentMessage.TraceID = firstNonEmpty(currentMessage.TraceID, traceID)
		currentMessage.ErrorCode = defaultString(currentMessage.ErrorCode, "bridge_stream_failed")
		currentMessage.ErrorMessage = errorMessage
		currentMessage.DeliveredAt = ""
		currentMessage.UpdatedAt = now
		writeMessage(currentMessage)
		if err := appendEvent(workspaceEventMessageFailed, workspaceEventPayload{
			Message:      &currentMessage,
			Error:        errorMessage,
			ErrorMessage: errorMessage,
			Status:       currentMessage.Status,
		}); err != nil {
			r.mu.Unlock()
			return err
		}
	case "done":
		if currentMessage.Status != "failed" {
			currentMessage.Status = "delivered"
			if strings.TrimSpace(currentMessage.Content) == "" {
				currentMessage.Content = "龙虾已完成处理，请查看右侧工作台或产物列表。"
			}
			currentMessage.DeliveredAt = defaultString(currentMessage.DeliveredAt, now)
			currentMessage.UpdatedAt = now
			writeMessage(currentMessage)
			if err := appendEvent(workspaceEventMessageCompleted, workspaceEventPayload{
				Message: &currentMessage,
				Content: currentMessage.Content,
				Status:  currentMessage.Status,
			}); err != nil {
				r.mu.Unlock()
				return err
			}
		}
	default:
		fallback := strings.TrimSpace(plainText)
		if fallback != "" {
			currentMessage.Status = "streaming"
			currentMessage.Origin = workspaceMessageOriginBridgeResponse
			currentMessage.TraceID = firstNonEmpty(currentMessage.TraceID, traceID)
			currentMessage.Content += fallback
			currentMessage.UpdatedAt = now
			writeMessage(currentMessage)
			if err := appendEvent(workspaceEventMessageDelta, workspaceEventPayload{
				Message: &currentMessage,
				Delta:   fallback,
				Content: currentMessage.Content,
				Status:  currentMessage.Status,
			}); err != nil {
				r.mu.Unlock()
				return err
			}
		}
	}

	if len(mutation.Events) > 0 {
		if idx := r.findWorkspaceSessionIndex(localSession.ID); idx >= 0 {
			mutation.Sessions = append(mutation.Sessions, r.data.WorkspaceSessions[idx])
		}
	}
	r.mu.Unlock()
	if len(mutation.Events) == 0 && len(mutation.Messages) == 0 && len(mutation.Artifacts) == 0 {
		return nil
	}
	return r.persistWorkspaceMutationAndPublish(mutation)
}

func (r *Router) failWorkspaceStreamMessage(session models.WorkspaceSession, assistantMessageID int, traceID string, err error) {
	_ = r.applyWorkspaceBridgeStreamEvent(session, assistantMessageID, workspaceEventMessageFailed, mustMarshalJSONString(map[string]any{
		"error": err.Error(),
	}), traceID)
}

func decodeWorkspaceBridgeStreamPayload(raw string) (workspaceBridgeStreamPayload, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return workspaceBridgeStreamPayload{}, ""
	}
	payload := workspaceBridgeStreamPayload{}
	if !json.Valid([]byte(trimmed)) {
		return payload, trimmed
	}
	if strings.HasPrefix(trimmed, "\"") {
		var plain string
		if err := json.Unmarshal([]byte(trimmed), &plain); err == nil {
			return payload, plain
		}
	}
	_ = json.Unmarshal([]byte(trimmed), &payload)
	return payload, ""
}

func mustMarshalJSONString(payload any) string {
	raw, err := json.Marshal(payload)
	if err != nil {
		return `{}`
	}
	return string(raw)
}
