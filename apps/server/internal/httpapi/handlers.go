package httpapi

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"openclaw/platformapi/internal/models"
)

func (r *Router) handlePortalOverview(w http.ResponseWriter, req *http.Request) {
	tenantID := tenantFilterID(req, 1)
	r.mu.RLock()
	rawInstances := r.filterInstancesByTenant(tenantID)
	jobs := r.filterJobsByTenant(tenantID)
	alerts := r.filterAlertsByTenant(tenantID)
	backups := r.filterBackupsByTenant(tenantID)
	r.mu.RUnlock()

	instances := make([]models.Instance, 0, len(rawInstances))
	for _, inst := range rawInstances {
		if state, ok := r.loadLiveInstanceState(inst.ID); ok {
			instances = append(instances, state.Instance)
			continue
		}
		instances = append(instances, inst)
	}

	res := map[string]any{
		"tenantId":          tenantID,
		"instanceTotal":     len(instances),
		"instanceRunning":   countStatus(instances, "running"),
		"instanceAbnormal":  countNotStatus(instances, "running"),
		"recentJobs":        limitJobs(jobs, 5),
		"recentAlerts":      limitAlerts(alerts, 5),
		"recentBackups":     limitBackups(backups, 5),
		"primaryInstanceId": primaryInstanceID(instances),
	}
	writeJSON(w, http.StatusOK, res)
}

func (r *Router) handlePortalInstances(w http.ResponseWriter, req *http.Request) {
	tenantID := tenantFilterID(req, 1)
	r.mu.RLock()
	instances := r.filterInstancesByTenant(tenantID)
	r.mu.RUnlock()

	withAccess := make([]map[string]any, 0, len(instances))
	for _, inst := range instances {
		liveInstance := inst
		if state, ok := r.loadLiveInstanceState(inst.ID); ok {
			liveInstance = state.Instance
		}

		r.mu.RLock()
		withAccess = append(withAccess, map[string]any{
			"instance": liveInstance,
			"access":   r.filterAccessByInstance(liveInstance.ID),
			"config":   r.findConfig(liveInstance.ID),
			"backups":  limitBackups(r.filterBackupsByInstance(liveInstance.ID), 3),
		})
		r.mu.RUnlock()
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": withAccess})
}

func (r *Router) handlePortalInstanceDetail(w http.ResponseWriter, req *http.Request) {
	idStr := strings.TrimPrefix(req.URL.Path, "/api/v1/portal/instances/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}
	state, ok := r.loadLiveInstanceState(id)
	if !ok {
		http.NotFound(w, req)
		return
	}

	r.mu.RLock()
	res := map[string]any{
		"instance": state.Instance,
		"access":   r.filterAccessByInstance(id),
		"config":   r.findConfig(id),
		"backups":  r.filterBackupsByInstance(id),
		"jobs":     r.filterJobsByInstance(id),
		"alerts":   r.buildObservabilityAlertsLocked(r.filterAlertsByInstance(id), "portal"),
	}
	r.mu.RUnlock()
	if state.Runtime != nil {
		res["runtime"] = state.Runtime
	}
	if state.Binding != nil {
		res["binding"] = state.Binding
	}
	writeJSON(w, http.StatusOK, res)
}

func (r *Router) handlePortalJobs(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	writeJSON(w, http.StatusOK, map[string]any{"items": r.filterJobsByTenant(tenantID)})
}

func (r *Router) handlePortalAlerts(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	writeJSON(w, http.StatusOK, map[string]any{"items": r.buildObservabilityAlertsLocked(r.filterAlertsByTenant(tenantID), "portal")})
}

func (r *Router) handlePortalLogs(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	logs := make([]models.AuditEvent, 0)
	for _, a := range r.data.Audits {
		if a.TenantID == tenantID {
			logs = append(logs, a)
		}
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt > logs[j].CreatedAt
	})
	writeJSON(w, http.StatusOK, map[string]any{"items": logs})
}

func (r *Router) handleAdminOverview(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instanceTotal := 0
	for _, item := range r.data.Instances {
		if visibleInstance(item) {
			instanceTotal++
		}
	}

	res := map[string]any{
		"tenantTotal":   len(r.data.Tenants),
		"instanceTotal": instanceTotal,
		"clusterTotal":  len(r.data.Clusters),
		"openAlerts":    countOpenAlerts(r.data.Alerts),
		"recentJobs":    limitJobs(r.data.Jobs, 8),
		"recentAlerts":  limitAlerts(r.data.Alerts, 8),
	}
	writeJSON(w, http.StatusOK, res)
}

func (r *Router) handleAdminTenants(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Tenants})
}

func (r *Router) handleAdminInstances(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	instances := append([]models.Instance(nil), r.data.Instances...)
	r.mu.RUnlock()
	items := make([]map[string]any, 0, len(instances))
	for _, inst := range instances {
		if !visibleInstance(inst) {
			continue
		}
		liveInstance := inst
		if state, ok := r.loadLiveInstanceState(inst.ID); ok {
			liveInstance = state.Instance
		}

		r.mu.RLock()
		items = append(items, map[string]any{
			"instance": liveInstance,
			"tenant":   r.findTenant(liveInstance.TenantID),
			"cluster":  r.findCluster(liveInstance.ClusterID),
			"access":   r.filterAccessByInstance(liveInstance.ID),
			"config":   r.findConfig(liveInstance.ID),
			"alerts":   r.filterAlertsByInstance(liveInstance.ID),
		})
		r.mu.RUnlock()
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleAdminInstanceDetail(w http.ResponseWriter, req *http.Request) {
	idStr := strings.TrimPrefix(req.URL.Path, "/api/v1/admin/instances/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	state, ok := r.loadLiveInstanceState(id)
	if !ok {
		http.NotFound(w, req)
		return
	}

	r.mu.RLock()
	instance := state.Instance
	tenant := r.findTenant(instance.TenantID)
	cluster := r.findCluster(instance.ClusterID)
	jobs := r.filterJobsByInstance(id)
	backups := r.filterBackupsByInstance(id)
	alerts := r.filterAlertsByInstance(id)
	workspaceSessions := r.buildInstanceWorkspaceSessionsLocked(id, 6)
	bridgeSummary := r.buildInstanceBridgeSummaryLocked(id)
	runtimeLogs := r.buildInstanceRuntimeLogsLocked(id, "admin")
	res := map[string]any{
		"instance":          instance,
		"tenant":            tenant,
		"cluster":           cluster,
		"access":            r.filterAccessByInstance(id),
		"config":            r.findConfig(id),
		"jobs":              jobs,
		"backups":           backups,
		"alerts":            r.buildObservabilityAlertsLocked(alerts, "admin"),
		"audits":            r.filterAuditsByInstance(id),
		"runtimeLogs":       runtimeLogs,
		"resourceTrend":     r.buildInstanceResourceTrend(state.Runtime, state.Metrics, alerts),
		"workspaceSessions": workspaceSessions,
		"bridgeSummary":     bridgeSummary,
	}
	r.mu.RUnlock()
	if state.Runtime != nil {
		res["runtime"] = state.Runtime
	}
	if state.Binding != nil {
		res["binding"] = state.Binding
	}
	writeJSON(w, http.StatusOK, res)
}

func (r *Router) handleAdminJobs(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Jobs})
}

func (r *Router) handleAdminAlerts(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{"items": r.buildObservabilityAlertsLocked(r.data.Alerts, "admin")})
}

func (r *Router) handleAdminAudit(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Audits})
}

// Channels

func (r *Router) handleChannelList(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]map[string]any, 0, len(r.data.Channels))
	for _, channel := range r.data.Channels {
		items = append(items, map[string]any{
			"channel":    channel,
			"activities": r.filterActivitiesByChannel(channel.ID, 5),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleChannelDetail(w http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, "/activities") {
		r.handleChannelActivities(w, req)
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := parseTailID(req.URL.Path, "/api/v1/channels/")
	if !ok {
		http.Error(w, "invalid channel id", http.StatusBadRequest)
		return
	}
	ch, idx := r.findChannel(id)
	if idx < 0 {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"channel":    ch,
		"activities": r.filterActivitiesByChannel(id, 10),
	})
}

func (r *Router) filterInstancesByTenant(tenantID int) []models.Instance {
	out := make([]models.Instance, 0)
	for _, inst := range r.data.Instances {
		if inst.TenantID == tenantID && visibleInstance(inst) {
			out = append(out, inst)
		}
	}
	return out
}

func (r *Router) filterJobsByTenant(tenantID int) []models.Job {
	out := make([]models.Job, 0)
	for _, inst := range r.data.Instances {
		if inst.TenantID != tenantID {
			continue
		}
		out = append(out, r.filterJobsByInstance(inst.ID)...)
	}
	return out
}

func (r *Router) filterJobsByInstance(instanceID int) []models.Job {
	out := make([]models.Job, 0)
	for _, job := range r.data.Jobs {
		if job.TargetType == "instance" && job.TargetID == instanceID {
			out = append(out, job)
		}
	}
	return out
}

func (r *Router) filterAlertsByTenant(tenantID int) []models.Alert {
	out := make([]models.Alert, 0)
	for _, inst := range r.data.Instances {
		if inst.TenantID != tenantID {
			continue
		}
		out = append(out, r.filterAlertsByInstance(inst.ID)...)
	}
	return out
}

func (r *Router) filterAlertsByInstance(instanceID int) []models.Alert {
	out := make([]models.Alert, 0)
	for _, alert := range r.data.Alerts {
		if alert.InstanceID == instanceID {
			out = append(out, alert)
		}
	}
	return out
}

func (r *Router) filterBackupsByTenant(tenantID int) []models.BackupRecord {
	out := make([]models.BackupRecord, 0)
	for _, inst := range r.data.Instances {
		if inst.TenantID != tenantID {
			continue
		}
		out = append(out, r.filterBackupsByInstance(inst.ID)...)
	}
	return out
}

func (r *Router) filterBackupsByInstance(instanceID int) []models.BackupRecord {
	out := make([]models.BackupRecord, 0)
	for _, bk := range r.data.Backups {
		if bk.InstanceID == instanceID {
			out = append(out, bk)
		}
	}
	return out
}

func (r *Router) filterAuditsByInstance(instanceID int) []models.AuditEvent {
	out := make([]models.AuditEvent, 0)
	for _, audit := range r.data.Audits {
		if audit.Target == "instance" && audit.TargetID == instanceID {
			out = append(out, audit)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt > out[j].CreatedAt
	})
	return out
}

func (r *Router) filterActivitiesByChannel(channelID int, limit int) []models.ChannelActivity {
	out := make([]models.ChannelActivity, 0)
	for _, act := range r.data.Activities {
		if act.ChannelID == channelID {
			out = append(out, act)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt > out[j].CreatedAt
	})
	if len(out) > limit {
		return out[:limit]
	}
	return out
}

func (r *Router) filterAccessByInstance(instanceID int) []models.InstanceAccess {
	out := make([]models.InstanceAccess, 0)
	for _, a := range r.data.Accesses {
		if a.InstanceID == instanceID {
			out = append(out, a)
		}
	}
	return out
}

func (r *Router) findConfig(instanceID int) *models.InstanceConfig {
	for _, cfg := range r.data.Configs {
		if cfg.InstanceID == instanceID {
			c := cfg
			return &c
		}
	}
	return nil
}

func (r *Router) findInstance(id int) (models.Instance, bool) {
	for _, inst := range r.data.Instances {
		if inst.ID == id {
			return inst, true
		}
	}
	return models.Instance{}, false
}

func (r *Router) findTenant(id int) *models.Tenant {
	for _, t := range r.data.Tenants {
		if t.ID == id {
			tt := t
			return &tt
		}
	}
	return nil
}

func (r *Router) findCluster(id int) *models.Cluster {
	for _, c := range r.data.Clusters {
		if c.ID == id {
			cc := c
			return &cc
		}
	}
	return nil
}

func (r *Router) findChannel(id int) (models.Channel, int) {
	for idx, ch := range r.data.Channels {
		if ch.ID == id {
			return ch, idx
		}
	}
	return models.Channel{}, -1
}

func countStatus(instances []models.Instance, status string) int {
	n := 0
	for _, inst := range instances {
		if inst.Status == status {
			n++
		}
	}
	return n
}

func countNotStatus(instances []models.Instance, status string) int {
	n := 0
	for _, inst := range instances {
		if inst.Status != status {
			n++
		}
	}
	return n
}

func primaryInstanceID(instances []models.Instance) *int {
	if len(instances) == 0 {
		return nil
	}
	id := instances[0].ID
	return &id
}

func limitJobs(jobs []models.Job, n int) []models.Job {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].StartedAt > jobs[j].StartedAt
	})
	if len(jobs) <= n {
		return jobs
	}
	return jobs[:n]
}

func limitAlerts(alerts []models.Alert, n int) []models.Alert {
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].TriggeredAt > alerts[j].TriggeredAt
	})
	if len(alerts) <= n {
		return alerts
	}
	return alerts[:n]
}

func limitBackups(backups []models.BackupRecord, n int) []models.BackupRecord {
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].StartedAt > backups[j].StartedAt
	})
	if len(backups) <= n {
		return backups
	}
	return backups[:n]
}

func countOpenAlerts(alerts []models.Alert) int {
	n := 0
	for _, a := range alerts {
		if a.Status == "open" {
			n++
		}
	}
	return n
}
