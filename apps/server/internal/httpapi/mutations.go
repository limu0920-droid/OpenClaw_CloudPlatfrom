package httpapi

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

type createInstanceRequest struct {
	Name   string `json:"name"`
	Plan   string `json:"plan"`
	Region string `json:"region"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type updateConfigRequest struct {
	UpdatedBy string                `json:"updatedBy"`
	Settings  models.ConfigSettings `json:"settings"`
}

type triggerBackupRequest struct {
	Type     string `json:"type"`
	Operator string `json:"operator"`
}

func (r *Router) handleCreatePortalInstance(w http.ResponseWriter, req *http.Request) {
	var payload createInstanceRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.Name) == "" || strings.TrimSpace(payload.Plan) == "" || strings.TrimSpace(payload.Region) == "" {
		http.Error(w, "name, plan and region are required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.CPU) == "" {
		payload.CPU = "1"
	}
	if strings.TrimSpace(payload.Memory) == "" {
		payload.Memory = "2Gi"
	}

	r.mu.Lock()

	now := time.Now().UTC().Format(time.RFC3339)
	tenantID := tenantFilterID(req, 1)
	instanceID := r.nextInstanceID()
	jobID := r.nextJobID()
	auditID := r.nextAuditID()
	offer := r.findPlanOffer(payload.Plan)
	if offer == nil {
		r.mu.Unlock()
		http.Error(w, "plan not found", http.StatusBadRequest)
		return
	}
	cluster, ok := r.pickClusterByRegion(payload.Region)
	if !ok {
		r.mu.Unlock()
		http.Error(w, "cluster not found for region", http.StatusBadRequest)
		return
	}
	code := fmt.Sprintf("inst-%s-%d", slugify(payload.Name), instanceID)

	instance := models.Instance{
		ID:          instanceID,
		TenantID:    tenantID,
		ClusterID:   cluster.ID,
		Code:        code,
		Name:        payload.Name,
		Status:      "running",
		Version:     "1.6.3",
		Plan:        payload.Plan,
		RuntimeType: "kubernetes",
		Region:      payload.Region,
		Spec: map[string]string{
			"cpu":    payload.CPU,
			"memory": payload.Memory,
		},
		ActivatedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	accesses := []models.InstanceAccess{
		{
			InstanceID: instanceID,
			EntryType:  "web",
			URL:        fmt.Sprintf("https://%s.openclaw.local", code),
			Domain:     fmt.Sprintf("%s.openclaw.local", code),
			AccessMode: "sso",
			IsPrimary:  true,
		},
		{
			InstanceID: instanceID,
			EntryType:  "admin",
			URL:        fmt.Sprintf("https://%s-admin.openclaw.local", code),
			Domain:     fmt.Sprintf("%s-admin.openclaw.local", code),
			AccessMode: "password",
			IsPrimary:  false,
		},
	}

	var (
		binding  *models.RuntimeBinding
		runtime  *models.InstanceRuntime
		workload *runtimeadapter.Workload
	)

	if provisioner, ok := r.runtime.(runtimeadapter.ProvisioningAdapter); ok {
		runtimeNamespace := kubeNamespace(r.config.RuntimeNamespacePrefix, tenantID, payload.Name, instanceID)
		workloadName := kubeWorkloadName(r.config.RuntimeWorkloadPrefix, payload.Name, instanceID)
		workloadID := fmt.Sprintf("wl-%d-%d", instanceID, time.Now().UTC().UnixNano()%1000000000)

		result, err := provisioner.CreateWorkload(runtimeadapter.CreateWorkloadRequest{
			WorkloadID: workloadID,
			Namespace:  runtimeNamespace,
			Name:       workloadName,
			Image:      r.config.RuntimeImage,
			Replicas:   1,
			Port:       r.config.RuntimePort,
			CPU:        payload.CPU,
			Memory:     payload.Memory,
			Labels: map[string]string{
				"openclaw.io/tenant-id":   strconv.Itoa(tenantID),
				"openclaw.io/instance-id": strconv.Itoa(instanceID),
				"openclaw.io/plan":        payload.Plan,
			},
			Env: map[string]string{
				"OPENCLAW_INSTANCE_CODE": code,
				"OPENCLAW_INSTANCE_NAME": payload.Name,
			},
		})
		if err != nil {
			r.mu.Unlock()
			http.Error(w, fmt.Sprintf("runtime provision failed: %v", err), http.StatusBadGateway)
			return
		}

		workload = &result.Workload
		instance.Status = instanceStatusFromWorkload(result.Workload)
		accesses = buildAccessesFromRuntime(instanceID, result.AccessEndpoints)
		if len(accesses) == 0 {
			accesses = []models.InstanceAccess{
				{
					InstanceID: instanceID,
					EntryType:  "web",
					URL:        fmt.Sprintf("http://%s.%s.svc.cluster.local", workloadName, runtimeNamespace),
					Domain:     fmt.Sprintf("%s.%s.svc.cluster.local", workloadName, runtimeNamespace),
					AccessMode: "direct",
					IsPrimary:  true,
				},
			}
		}

		binding = &models.RuntimeBinding{
			InstanceID:   instanceID,
			ClusterID:    result.Workload.ClusterID,
			Namespace:    result.Workload.Namespace,
			WorkloadID:   result.Workload.ID,
			WorkloadName: result.Workload.Name,
		}

		var metrics *runtimeadapter.WorkloadMetrics
		if runtimeMetrics, ok := r.runtime.GetMetrics(result.Workload.ID); ok {
			metrics = &runtimeMetrics
		}
		runtimeSnapshot := buildInstanceRuntime(instanceID, instance.Spec, result.Workload, metrics)
		runtime = &runtimeSnapshot
		r.upsertCredentialLocked(models.InstanceCredential{
			InstanceID:     instanceID,
			AdminUser:      "admin",
			PasswordMasked: "setup-required",
			LastRotatedAt:  now,
			RequiresReset:  true,
		})
	}
	if runtime == nil {
		runtimeSnapshot := models.InstanceRuntime{
			InstanceID:         instanceID,
			PowerState:         "running",
			CPUUsagePercent:    0,
			MemoryUsagePercent: 0,
			DiskUsagePercent:   0,
			APIRequests24h:     0,
			APITokens24h:       0,
			LastSeenAt:         now,
		}
		runtime = &runtimeSnapshot
		r.upsertCredentialLocked(models.InstanceCredential{
			InstanceID:     instanceID,
			AdminUser:      "admin",
			PasswordMasked: "Init***00",
			LastRotatedAt:  now,
			RequiresReset:  true,
		})
	}

	config := models.InstanceConfig{
		InstanceID:  instanceID,
		Version:     1,
		Hash:        shortHash(instanceID, now),
		PublishedAt: now,
		UpdatedBy:   "system-bootstrap",
		Settings: models.ConfigSettings{
			Model:          "gpt-5",
			AllowedOrigins: accesses[0].URL,
			BackupPolicy:   "daily@02:00 保留 7 天",
		},
	}

	job := models.Job{
		ID:         jobID,
		JobNo:      fmt.Sprintf("job-%d", 3000+jobID),
		Type:       "createInstance",
		TargetType: "instance",
		TargetID:   instanceID,
		Status:     "succeeded",
		Summary:    "实例创建完成",
		StartedAt:  now,
		FinishedAt: now,
	}

	audit := models.AuditEvent{
		ID:        auditID,
		TenantID:  tenantID,
		Actor:     "portal-user",
		Action:    "createInstance",
		Target:    "instance",
		TargetID:  instanceID,
		Result:    "success",
		CreatedAt: now,
		Metadata: map[string]string{
			"plan":   payload.Plan,
			"region": payload.Region,
		},
	}

	r.data.Instances = append([]models.Instance{instance}, r.data.Instances...)
	r.data.Accesses = append(accesses, r.data.Accesses...)
	r.data.Configs = append([]models.InstanceConfig{config}, r.data.Configs...)
	r.data.Jobs = append([]models.Job{job}, r.data.Jobs...)
	r.data.Audits = append([]models.AuditEvent{audit}, r.data.Audits...)
	if binding != nil {
		r.upsertRuntimeBindingLocked(*binding)
	}
	if runtime != nil {
		r.upsertRuntimeLocked(*runtime)
	}
	state, shouldPersist := r.snapshotInstanceStateLocked(instanceID)
	r.mu.Unlock()
	if shouldPersist {
		if err := r.store.SaveInstanceState(state); err != nil {
			http.Error(w, fmt.Sprintf("persist instance state failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"instance": instance,
		"access":   accesses,
		"config":   config,
		"job":      job,
		"binding":  binding,
		"runtime":  runtime,
		"workload": workload,
	})
}

func (r *Router) handleUpdatePortalInstanceConfig(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/config")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	var payload updateConfigRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.UpdatedBy) == "" {
		payload.UpdatedBy = "portal-user"
	}
	approvalNo := strings.TrimSpace(req.URL.Query().Get("approvalNo"))
	if approvalNo == "" {
		approvalNo = strings.TrimSpace(req.Header.Get("X-OpenClaw-Approval-No"))
	}
	if approvalNo == "" {
		http.Error(w, "approval required", http.StatusConflict)
		return
	}
	if _, err := r.resolveApprovedApproval(approvalNo, approvalTypeConfigPublish, instanceID); err != nil {
		writeApprovalError(w, err)
		return
	}

	response, err := r.performConfigPublish(instanceID, payload)
	if err != nil {
		writeApprovalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) handleTriggerPortalBackup(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/portal/instances/", "/backups")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	var payload triggerBackupRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.Type) == "" {
		payload.Type = "manual"
	}
	if strings.TrimSpace(payload.Operator) == "" {
		payload.Operator = "portal-user"
	}

	r.mu.Lock()

	instanceIndex := r.findInstanceIndex(instanceID)
	if instanceIndex < 0 || !visibleInstance(r.data.Instances[instanceIndex]) {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	backupID := r.nextBackupID()
	jobID := r.nextJobID()
	auditID := r.nextAuditID()
	backup := models.BackupRecord{
		ID:         backupID,
		InstanceID: instanceID,
		BackupNo:   fmt.Sprintf("bk-%s-%03d", time.Now().UTC().Format("20060102-150405"), backupID),
		Type:       payload.Type,
		Status:     "succeeded",
		SizeBytes:  int64(256+backupID*8) * 1024 * 1024,
		StartedAt:  now,
		FinishedAt: now,
	}
	job := models.Job{
		ID:         jobID,
		JobNo:      fmt.Sprintf("job-%d", 3000+jobID),
		Type:       "backup",
		TargetType: "instance",
		TargetID:   instanceID,
		Status:     "succeeded",
		Summary:    "备份创建完成",
		StartedAt:  now,
		FinishedAt: now,
	}
	audit := models.AuditEvent{
		ID:        auditID,
		TenantID:  r.data.Instances[instanceIndex].TenantID,
		Actor:     payload.Operator,
		Action:    "backup",
		Target:    "instance",
		TargetID:  instanceID,
		Result:    "success",
		CreatedAt: now,
		Metadata: map[string]string{
			"backupNo": backup.BackupNo,
			"type":     payload.Type,
		},
	}

	r.data.Instances[instanceIndex].UpdatedAt = now
	r.data.Backups = append([]models.BackupRecord{backup}, r.data.Backups...)
	r.data.Jobs = append([]models.Job{job}, r.data.Jobs...)
	r.data.Audits = append([]models.AuditEvent{audit}, r.data.Audits...)
	state, shouldPersist := r.snapshotInstanceStateLocked(instanceID)
	r.mu.Unlock()
	if shouldPersist {
		if err := r.store.SaveInstanceState(state); err != nil {
			http.Error(w, fmt.Sprintf("persist backup state failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"backup": backup,
		"job":    job,
	})
}

func (r *Router) nextInstanceID() int {
	maxID := 0
	for _, item := range r.data.Instances {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextBackupID() int {
	maxID := 0
	for _, item := range r.data.Backups {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextJobID() int {
	maxID := 0
	for _, item := range r.data.Jobs {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextAuditID() int {
	maxID := 0
	for _, item := range r.data.Audits {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) pickClusterByRegion(region string) (models.Cluster, bool) {
	for _, cluster := range r.data.Clusters {
		if cluster.Region == region {
			return cluster, true
		}
	}
	return models.Cluster{}, false
}

func (r *Router) findInstanceIndex(id int) int {
	for index, instance := range r.data.Instances {
		if instance.ID == id {
			return index
		}
	}
	return -1
}

func (r *Router) findConfigIndex(instanceID int) int {
	for index, config := range r.data.Configs {
		if config.InstanceID == instanceID {
			return index
		}
	}
	return -1
}

func parseInstanceID(path string, prefix string, suffix string) (int, bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.TrimSuffix(trimmed, suffix)
	trimmed = strings.Trim(trimmed, "/")
	id, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, false
	}
	return id, true
}

func slugify(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	lower = strings.ReplaceAll(lower, " ", "-")
	matcher := regexp.MustCompile(`[^a-z0-9\-]+`)
	lower = matcher.ReplaceAllString(lower, "")
	if lower == "" {
		return "instance"
	}
	return lower
}

func shortHash(id int, timestamp string) string {
	return fmt.Sprintf("cfg-%d-%d", id, time.Now().UTC().Unix()%100000)
}

func parseTailID(path string, prefix string, suffix ...string) (int, bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	for _, s := range suffix {
		trimmed = strings.TrimSuffix(trimmed, s)
	}
	trimmed = strings.Trim(trimmed, "/")
	id, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, false
	}
	return id, true
}

func matchesTailIDPath(path string, prefix string, suffix ...string) bool {
	_, ok := parseTailID(path, prefix, suffix...)
	return ok
}

func maskToken(token string) string {
	if len(token) <= 4 {
		return "***"
	}
	return token[:2] + strings.Repeat("*", len(token)-4) + token[len(token)-2:]
}

func (r *Router) nextActivityID() int {
	maxID := 0
	for _, item := range r.data.Activities {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}
