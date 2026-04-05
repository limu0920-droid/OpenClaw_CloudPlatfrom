package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"openclaw/mockapi/internal/models"
)

type purchaseRequest struct {
	PlanCode   string `json:"planCode"`
	InstanceID int    `json:"instanceId,omitempty"`
	Action     string `json:"action"`
}

type createTicketRequest struct {
	InstanceID   int    `json:"instanceId,omitempty"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	Severity     string `json:"severity"`
	Description  string `json:"description"`
	Reporter     string `json:"reporter"`
}

type updateTicketStatusRequest struct {
	Status   string `json:"status"`
	Assignee string `json:"assignee"`
}

func (r *Router) handlePortalInstanceRuntime(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/runtime")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	instance, found := r.findInstance(instanceID)
	if !found {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"instance":    instance,
		"runtime":     r.findRuntime(instanceID),
		"credentials": r.findCredential(instanceID),
		"orders":      r.filterOrdersByInstance(instanceID),
	})
}

func (r *Router) handlePortalInstancePower(w http.ResponseWriter, req *http.Request) {
	action := "restart"
	switch {
	case strings.HasSuffix(req.URL.Path, "/start"):
		action = "start"
	case strings.HasSuffix(req.URL.Path, "/stop"):
		action = "stop"
	}

	var instanceID int
	var ok bool
	switch action {
	case "start":
		instanceID, ok = parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/start")
	case "stop":
		instanceID, ok = parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/stop")
	default:
		instanceID, ok = parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/restart")
	}
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	instanceIndex := r.findInstanceIndex(instanceID)
	runtimeIndex := r.findRuntimeIndex(instanceID)
	if instanceIndex < 0 || runtimeIndex < 0 {
		http.NotFound(w, req)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	jobID := r.nextJobID()
	auditID := r.nextAuditID()

	switch action {
	case "start":
		r.data.Instances[instanceIndex].Status = "running"
		r.data.Runtimes[runtimeIndex].PowerState = "running"
	case "stop":
		r.data.Instances[instanceIndex].Status = "stopped"
		r.data.Runtimes[runtimeIndex].PowerState = "stopped"
	default:
		r.data.Instances[instanceIndex].Status = "running"
		r.data.Runtimes[runtimeIndex].PowerState = "running"
	}
	r.data.Instances[instanceIndex].UpdatedAt = now
	r.data.Runtimes[runtimeIndex].LastSeenAt = now

	job := models.Job{
		ID:         jobID,
		JobNo:      fmt.Sprintf("job-%d", 3000+jobID),
		Type:       action + "Instance",
		TargetType: "instance",
		TargetID:   instanceID,
		Status:     "succeeded",
		Summary:    "实例电源操作完成",
		StartedAt:  now,
		FinishedAt: now,
	}

	audit := models.AuditEvent{
		ID:        auditID,
		TenantID:  r.data.Instances[instanceIndex].TenantID,
		Actor:     "portal-user",
		Action:    action + "Instance",
		Target:    "instance",
		TargetID:  instanceID,
		Result:    "success",
		CreatedAt: now,
	}

	r.data.Jobs = append([]models.Job{job}, r.data.Jobs...)
	r.data.Audits = append([]models.AuditEvent{audit}, r.data.Audits...)

	writeJSON(w, http.StatusOK, map[string]any{
		"instance": r.data.Instances[instanceIndex],
		"runtime":  r.data.Runtimes[runtimeIndex],
		"job":      job,
	})
}

func (r *Router) handlePortalPlans(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.PlanOffers})
}

func (r *Router) handlePortalPurchase(w http.ResponseWriter, req *http.Request) {
	var payload purchaseRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.PlanCode) == "" {
		http.Error(w, "planCode is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Action) == "" {
		payload.Action = "buy"
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	orderID := r.nextOrderID()
	tenantID := tenantFilterID(req, 1)
	offer := r.findPlanOffer(payload.PlanCode)
	if offer == nil {
		http.Error(w, "plan not found", http.StatusBadRequest)
		return
	}

	order := models.Order{
		ID:         orderID,
		TenantID:   tenantID,
		InstanceID: payload.InstanceID,
		PlanCode:   payload.PlanCode,
		Action:     payload.Action,
		Status:     "pending",
		Amount:     offer.MonthlyPrice,
		CreatedAt:  now,
	}
	r.data.Orders = append([]models.Order{order}, r.data.Orders...)

	writeJSON(w, http.StatusCreated, map[string]any{"order": order, "plan": offer})
}

func (r *Router) handlePortalTickets(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	writeJSON(w, http.StatusOK, map[string]any{"items": r.filterTicketsByTenant(tenantID)})
}

func (r *Router) handlePortalCreateTicket(w http.ResponseWriter, req *http.Request) {
	var payload createTicketRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Reporter) == "" {
		http.Error(w, "title and reporter are required", http.StatusBadRequest)
		return
	}

	if payload.Category == "" {
		payload.Category = "general"
	}
	if payload.Severity == "" {
		payload.Severity = "medium"
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	ticketID := r.nextTicketID()
	tenantID := tenantFilterID(req, 1)
	ticket := models.Ticket{
		ID:          ticketID,
		TicketNo:    fmt.Sprintf("TK-%s-%03d", time.Now().UTC().Format("20060102"), ticketID),
		TenantID:    tenantID,
		InstanceID:  payload.InstanceID,
		Title:       payload.Title,
		Category:    payload.Category,
		Severity:    payload.Severity,
		Status:      "open",
		Reporter:    payload.Reporter,
		Description: payload.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	r.data.Tickets = append([]models.Ticket{ticket}, r.data.Tickets...)
	writeJSON(w, http.StatusCreated, map[string]any{"ticket": ticket})
}

func (r *Router) handleAdminTickets(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Tickets})
}

func (r *Router) handleAdminUpdateTicketStatus(w http.ResponseWriter, req *http.Request) {
	ticketID, ok := parseTailID(req.URL.Path, "/api/v1/admin/tickets/", "/status")
	if !ok {
		http.Error(w, "invalid ticket id", http.StatusBadRequest)
		return
	}

	var payload updateTicketStatusRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	index := r.findTicketIndex(ticketID)
	if index < 0 {
		http.NotFound(w, req)
		return
	}

	if payload.Status != "" {
		r.data.Tickets[index].Status = payload.Status
	}
	if payload.Assignee != "" {
		r.data.Tickets[index].Assignee = payload.Assignee
	}
	r.data.Tickets[index].UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	writeJSON(w, http.StatusOK, map[string]any{"ticket": r.data.Tickets[index]})
}

func (r *Router) findRuntime(instanceID int) *models.InstanceRuntime {
	for _, item := range r.data.Runtimes {
		if item.InstanceID == instanceID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) findCredential(instanceID int) *models.InstanceCredential {
	for _, item := range r.data.Credentials {
		if item.InstanceID == instanceID {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) filterOrdersByInstance(instanceID int) []models.Order {
	out := make([]models.Order, 0)
	for _, item := range r.data.Orders {
		if item.InstanceID == instanceID {
			out = append(out, item)
		}
	}
	return out
}

func (r *Router) filterTicketsByTenant(tenantID int) []models.Ticket {
	out := make([]models.Ticket, 0)
	for _, item := range r.data.Tickets {
		if item.TenantID == tenantID {
			out = append(out, item)
		}
	}
	return out
}

func (r *Router) findPlanOffer(code string) *models.PlanOffer {
	for _, item := range r.data.PlanOffers {
		if item.Code == code {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *Router) findRuntimeIndex(instanceID int) int {
	for index, item := range r.data.Runtimes {
		if item.InstanceID == instanceID {
			return index
		}
	}
	return -1
}

func (r *Router) nextOrderID() int {
	maxID := 0
	for _, item := range r.data.Orders {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextTicketID() int {
	maxID := 0
	for _, item := range r.data.Tickets {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) findTicketIndex(id int) int {
	for index, item := range r.data.Tickets {
		if item.ID == id {
			return index
		}
	}
	return -1
}
