package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

type purchaseRequest struct {
	PlanCode   string `json:"planCode"`
	InstanceID int    `json:"instanceId,omitempty"`
	Action     string `json:"action"`
}

type createTicketRequest struct {
	InstanceID  int    `json:"instanceId,omitempty"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Reporter    string `json:"reporter"`
}

type updateTicketStatusRequest struct {
	Status   string `json:"status"`
	Assignee string `json:"assignee"`
}

func (r *Router) handlePortalInstanceRuntime(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/runtime")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	state, found := r.loadLiveInstanceState(instanceID)
	r.mu.RLock()
	credential := r.findCredential(instanceID)
	orders := r.filterOrdersByInstance(instanceID)
	r.mu.RUnlock()
	if !found {
		http.NotFound(w, req)
		return
	}

	response := map[string]any{
		"instance":    state.Instance,
		"runtime":     state.Runtime,
		"credentials": credential,
		"orders":      orders,
	}
	if state.Binding != nil {
		response["binding"] = state.Binding
	}
	if state.Workload != nil {
		response["workload"] = state.Workload
		response["pods"] = r.runtime.ListPods(state.Binding.WorkloadID)
	}
	if state.Metrics != nil {
		response["metrics"] = *state.Metrics
	}
	if state.Binding != nil {
		if accessProvider, ok := r.runtime.(runtimeadapter.AccessInfoProvider); ok {
			response["accessEndpoints"] = accessProvider.GetWorkloadAccess(state.Binding.WorkloadID)
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) handlePortalInstancePower(w http.ResponseWriter, req *http.Request) {
	action := "restart"
	switch {
	case strings.HasSuffix(req.URL.Path, "/start"):
		action = "start"
	case strings.HasSuffix(req.URL.Path, "/stop"), strings.HasSuffix(req.URL.Path, "/pause"):
		action = "stop"
	}

	var instanceID int
	var ok bool
	switch action {
	case "start":
		instanceID, ok = parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/start")
	case "stop":
		instanceID, ok = parsePortalPowerInstanceID(req.URL.Path, "stop")
	default:
		instanceID, ok = parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/restart")
	}
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	instance, found := r.findInstance(instanceID)
	binding := r.findRuntimeBinding(instanceID)
	r.mu.RUnlock()
	if !found || !visibleInstance(instance) {
		http.NotFound(w, req)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var (
		result          runtimeadapter.ActionResult
		workload        *runtimeadapter.Workload
		runtimeSnapshot *models.InstanceRuntime
		actionFound     bool
	)
	if binding != nil {
		switch action {
		case "start":
			result, actionFound = r.runtime.StartWorkload(binding.WorkloadID)
		case "stop":
			result, actionFound = r.runtime.StopWorkload(binding.WorkloadID)
		default:
			result, actionFound = r.runtime.RestartWorkload(binding.WorkloadID)
		}
		if actionFound {
			if latestWorkload, ok := r.runtime.GetWorkload(binding.WorkloadID); ok {
				workload = &latestWorkload
				if runtimeMetrics, ok := r.runtime.GetMetrics(binding.WorkloadID); ok {
					snapshot := buildInstanceRuntime(instanceID, instance.Spec, latestWorkload, &runtimeMetrics)
					runtimeSnapshot = &snapshot
				} else {
					snapshot := buildInstanceRuntime(instanceID, instance.Spec, latestWorkload, nil)
					runtimeSnapshot = &snapshot
				}
			}
		}
	}

	r.mu.Lock()

	instanceIndex := r.findInstanceIndex(instanceID)
	runtimeIndex := r.findRuntimeIndex(instanceID)
	if instanceIndex < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	jobID := r.nextJobID()
	auditID := r.nextAuditID()

	if actionFound && workload != nil {
		r.data.Instances[instanceIndex].Status = instanceStatusFromWorkload(*workload)
		r.upsertRuntimeLocked(*runtimeSnapshot)
	} else {
		if runtimeIndex < 0 {
			r.upsertRuntimeLocked(models.InstanceRuntime{
				InstanceID: instanceID,
				LastSeenAt: now,
			})
			runtimeIndex = r.findRuntimeIndex(instanceID)
		}
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
		r.data.Runtimes[runtimeIndex].LastSeenAt = now
	}
	r.data.Instances[instanceIndex].UpdatedAt = now

	job := models.Job{
		ID:         jobID,
		JobNo:      fmt.Sprintf("job-%d", 3000+jobID),
		Type:       action + "Instance",
		TargetType: "instance",
		TargetID:   instanceID,
		Status:     actionResultStatus(result, actionFound),
		Summary:    actionResultSummary(action, actionFound),
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
	state, shouldPersist := r.snapshotInstanceStateLocked(instanceID)
	responseInstance := r.data.Instances[instanceIndex]
	responseRuntime := r.findRuntime(instanceID)
	r.mu.Unlock()
	if shouldPersist {
		if err := r.store.SaveInstanceState(state); err != nil {
			http.Error(w, fmt.Sprintf("persist power state failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"instance": responseInstance,
		"runtime":  responseRuntime,
		"job":      job,
		"binding":  binding,
		"result":   result,
		"workload": workload,
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

	now := time.Now().UTC().Format(time.RFC3339)
	orderID := r.nextOrderID()
	tenantID := tenantFilterID(req, 1)
	offer := r.findPlanOffer(payload.PlanCode)
	if offer == nil {
		r.mu.Unlock()
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
		Currency:   "CNY",
		OrderNo:    fmt.Sprintf("ORD-%s-%03d", time.Now().UTC().Format("20060102"), orderID),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.data.Orders = append([]models.Order{order}, r.data.Orders...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist purchase failed", http.StatusInternalServerError)
		return
	}

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

	reporter := strings.TrimSpace(payload.Reporter)
	if reporter == "" {
		if profile := r.resolveCurrentUserProfile(req); profile != nil {
			reporter = firstNonEmpty(strings.TrimSpace(profile.DisplayName), strings.TrimSpace(profile.Email), profile.LoginName)
		}
	}

	if strings.TrimSpace(payload.Title) == "" || reporter == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	if payload.Category == "" {
		payload.Category = "general"
	}
	if payload.Severity == "" {
		payload.Severity = "medium"
	}

	r.mu.Lock()

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
		Reporter:    reporter,
		Description: payload.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	r.data.Tickets = append([]models.Ticket{ticket}, r.data.Tickets...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist ticket failed", http.StatusInternalServerError)
		return
	}
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

	index := r.findTicketIndex(ticketID)
	if index < 0 {
		r.mu.Unlock()
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
	ticket := r.data.Tickets[index]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist ticket status failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ticket": ticket})
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

func parsePortalPowerInstanceID(path string, action string) (int, bool) {
	if action != "stop" {
		return 0, false
	}
	if strings.HasSuffix(path, "/pause") {
		return parseInstanceID(path, "/api/v1/portal/instances/", "/pause")
	}
	return parseInstanceID(path, "/api/v1/portal/instances/", "/stop")
}

func actionResultStatus(result runtimeadapter.ActionResult, found bool) string {
	if !found || strings.TrimSpace(result.Status) == "" {
		return "succeeded"
	}
	return strings.ToLower(result.Status)
}

func actionResultSummary(action string, found bool) string {
	if !found {
		return "实例电源操作完成"
	}
	switch action {
	case "start":
		return "实例启动指令已下发"
	case "stop":
		return "实例暂停指令已下发"
	default:
		return "实例重启指令已下发"
	}
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
