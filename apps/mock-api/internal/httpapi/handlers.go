package httpapi

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"openclaw/mockapi/internal/models"
)

func (r *Router) handlePortalOverview(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	instances := r.filterInstancesByTenant(tenantID)
	jobs := r.filterJobsByTenant(tenantID)
	alerts := r.filterAlertsByTenant(tenantID)
	backups := r.filterBackupsByTenant(tenantID)

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
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	instances := r.filterInstancesByTenant(tenantID)
	withAccess := make([]map[string]any, 0, len(instances))
	for _, inst := range instances {
		withAccess = append(withAccess, map[string]any{
			"instance": inst,
			"access":   r.filterAccessByInstance(inst.ID),
			"config":   r.findConfig(inst.ID),
			"backups":  limitBackups(r.filterBackupsByInstance(inst.ID), 3),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": withAccess})
}

func (r *Router) handlePortalInstanceDetail(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idStr := strings.TrimPrefix(req.URL.Path, "/api/v1/portal/instances/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}
	inst, ok := r.findInstance(id)
	if !ok {
		http.NotFound(w, req)
		return
	}
	res := map[string]any{
		"instance": inst,
		"access":   r.filterAccessByInstance(id),
		"config":   r.findConfig(id),
		"backups":  r.filterBackupsByInstance(id),
		"jobs":     r.filterJobsByInstance(id),
		"alerts":   r.filterAlertsByInstance(id),
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
	writeJSON(w, http.StatusOK, map[string]any{"items": r.filterAlertsByTenant(tenantID)})
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

	res := map[string]any{
		"tenantTotal":   len(r.data.Tenants),
		"instanceTotal": len(r.data.Instances),
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
	defer r.mu.RUnlock()

	items := make([]map[string]any, 0, len(r.data.Instances))
	for _, inst := range r.data.Instances {
		items = append(items, map[string]any{
			"instance": inst,
			"tenant":   r.findTenant(inst.TenantID),
			"cluster":  r.findCluster(inst.ClusterID),
			"access":   r.filterAccessByInstance(inst.ID),
			"config":   r.findConfig(inst.ID),
			"alerts":   r.filterAlertsByInstance(inst.ID),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleAdminInstanceDetail(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idStr := strings.TrimPrefix(req.URL.Path, "/api/v1/admin/instances/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	instance, ok := r.findInstance(id)
	if !ok {
		http.NotFound(w, req)
		return
	}

	tenant := r.findTenant(instance.TenantID)
	cluster := r.findCluster(instance.ClusterID)
	jobs := r.filterJobsByInstance(id)
	backups := r.filterBackupsByInstance(id)
	alerts := r.filterAlertsByInstance(id)

	writeJSON(w, http.StatusOK, map[string]any{
		"instance": instance,
		"tenant":   tenant,
		"cluster":  cluster,
		"access":   r.filterAccessByInstance(id),
		"config":   r.findConfig(id),
		"jobs":     jobs,
		"backups":  backups,
		"alerts":   alerts,
		"audits":   r.filterAuditsByInstance(id),
		"runtimeLogs": []map[string]any{
			{
				"id":        fmt.Sprintf("log-%d-1", id),
				"timestamp": "2026-04-05T08:12:00Z",
				"level":     "info",
				"source":    "gateway",
				"message":   "health probe passed and primary route is ready",
			},
			{
				"id":        fmt.Sprintf("log-%d-2", id),
				"timestamp": "2026-04-05T10:02:00Z",
				"level":     "warning",
				"source":    "runtime-adapter",
				"message":   "cpu usage crossed 80% for 4 minutes",
			},
			{
				"id":        fmt.Sprintf("log-%d-3", id),
				"timestamp": "2026-04-05T10:08:00Z",
				"level":     "info",
				"source":    "backup-service",
				"message":   "latest backup policy evaluated successfully",
			},
		},
		"resourceTrend": []map[string]any{
			{"label": "08:00", "cpu": 38, "memory": 44, "requests": 1200},
			{"label": "10:00", "cpu": 52, "memory": 49, "requests": 1860},
			{"label": "12:00", "cpu": 48, "memory": 55, "requests": 1740},
			{"label": "14:00", "cpu": 64, "memory": 58, "requests": 2210},
			{"label": "16:00", "cpu": 57, "memory": 62, "requests": 2050},
			{"label": "18:00", "cpu": 43, "memory": 53, "requests": 1480},
		},
	})
}

func (r *Router) handleAdminJobs(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Jobs})
}

func (r *Router) handleAdminAlerts(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Alerts})
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
		if inst.TenantID == tenantID {
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
