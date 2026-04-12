package httpapi

import (
	"net/http"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
)

type channelConnectRequest struct {
	Method string `json:"method,omitempty"` // oauth, token, qrcode, webhook
	Token  string `json:"token,omitempty"`
}

func (r *Router) handleChannelConnect(w http.ResponseWriter, req *http.Request) {
	id, ok := parseTailID(req.URL.Path, "/api/v1/channels/", "/connect")
	if !ok {
		http.Error(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	var payload channelConnectRequest
	_ = decodeJSON(req, &payload) // tolerate empty body

	r.mu.Lock()

	ch, idx := r.findChannel(id)
	if idx < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	ch.Status = "connected"
	ch.Health.Status = "healthy"
	ch.Health.LastChecked = now
	ch.Health.LatencyMs = 180
	ch.LastError = ""
	ch.UpdatedAt = now
	if normalized := normalizeConnectMethod(payload.Method); normalized != "" {
		ch.ConnectMethod = normalized
	}
	if payload.Token != "" {
		ch.TokenMasked = maskToken(payload.Token)
		ch.CallbackSecret = maskToken(payload.Token)
	}
	if ch.WebhookURL == "" {
		ch.WebhookURL = "https://api.openclaw.local/channels/" + ch.Code + "/webhook"
	}
	r.data.Channels[idx] = ch

	r.data.Activities = append([]models.ChannelActivity{
		{
			ID:        r.nextActivityID(),
			ChannelID: ch.ID,
			Type:      "auth",
			Title:     "渠道已连接",
			Summary:   "连接方式：" + ch.ConnectMethod,
			CreatedAt: now,
		},
	}, r.data.Activities...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist channel connect failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"channel": ch})
}

func (r *Router) handleChannelDisconnect(w http.ResponseWriter, req *http.Request) {
	id, ok := parseTailID(req.URL.Path, "/api/v1/channels/", "/disconnect")
	if !ok {
		http.Error(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	ch, idx := r.findChannel(id)
	if idx < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	ch.Status = "disconnected"
	ch.Health.Status = "critical"
	ch.Health.LastChecked = now
	ch.Health.LatencyMs = 0
	ch.LastError = "已断开连接"
	ch.TokenMasked = ""
	ch.UpdatedAt = now
	r.data.Channels[idx] = ch

	r.data.Activities = append([]models.ChannelActivity{
		{
			ID:        r.nextActivityID(),
			ChannelID: ch.ID,
			Type:      "auth",
			Title:     "渠道已断开",
			Summary:   "人工断开连接",
			CreatedAt: now,
		},
	}, r.data.Activities...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist channel disconnect failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"channel": ch})
}

func (r *Router) handleChannelHealthCheck(w http.ResponseWriter, req *http.Request) {
	id, ok := parseTailID(req.URL.Path, "/api/v1/channels/", "/health")
	if !ok {
		http.Error(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	ch, idx := r.findChannel(id)
	if idx < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	// Simple simulation: flip degraded to healthy if connected, else keep warning
	if ch.Status == "connected" {
		ch.Health.Status = "healthy"
		ch.Health.LatencyMs = 180
		ch.LastError = ""
	} else {
		ch.Health.Status = "warning"
		ch.Health.LatencyMs = 0
		ch.LastError = "未连接，需完成授权或配置"
	}
	ch.Health.LastChecked = now
	ch.UpdatedAt = now
	r.data.Channels[idx] = ch

	r.data.Activities = append([]models.ChannelActivity{
		{
			ID:        r.nextActivityID(),
			ChannelID: ch.ID,
			Type:      "health",
			Title:     "健康检查",
			Summary:   "状态：" + ch.Health.Status,
			CreatedAt: now,
		},
	}, r.data.Activities...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist channel health failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"channel": ch})
}

func (r *Router) handleChannelActivities(w http.ResponseWriter, req *http.Request) {
	id, ok := parseTailID(req.URL.Path, "/api/v1/channels/", "/activities")
	if !ok {
		http.Error(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, idx := r.findChannel(id); idx < 0 {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": r.filterActivitiesByChannel(id, 50),
	})
}

func normalizeConnectMethod(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "qr":
		return "qrcode"
	case "":
		return ""
	default:
		return lower
	}
}
