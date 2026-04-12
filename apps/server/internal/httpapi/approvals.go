package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

const (
	approvalStatusPending        = "pending"
	approvalStatusApproved       = "approved"
	approvalStatusRejected       = "rejected"
	approvalStatusExecuting      = "executing"
	approvalStatusExecuted       = "executed"
	approvalStatusCancelled      = "cancelled"
	approvalStatusExpired        = "expired"
	approvalTypeDeleteInstance   = "delete_instance"
	approvalTypeRuntimeStop      = "runtime_stop"
	approvalTypeRuntimeRestart   = "runtime_restart"
	approvalTypeRuntimeScale     = "runtime_scale"
	approvalTypeConfigPublish    = "config_publish"
	approvalTypeDiagnosticAccess = "diagnostic_access"
)

type approvalCreateRequest struct {
	ApprovalType string            `json:"approvalType"`
	TargetType   string            `json:"targetType"`
	TargetID     int               `json:"targetId"`
	InstanceID   int               `json:"instanceId,omitempty"`
	RiskLevel    string            `json:"riskLevel"`
	Reason       string            `json:"reason"`
	Comment      string            `json:"comment,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type approvalDecisionRequest struct {
	Comment      string `json:"comment,omitempty"`
	RejectReason string `json:"rejectReason,omitempty"`
}

type approvalRequestError struct {
	StatusCode int
	Message    string
}

func (e approvalRequestError) Error() string {
	return e.Message
}

func (r *Router) handlePortalApprovals(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.listApprovals(w, req, "portal")
	case http.MethodPost:
		r.createApproval(w, req, "portal")
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleAdminApprovals(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.listApprovals(w, req, "admin")
	case http.MethodPost:
		r.createApproval(w, req, "admin")
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleAdminApprovalDetail(w http.ResponseWriter, req *http.Request) {
	approvalID, ok := parseTailID(req.URL.Path, "/api/v1/admin/approvals/")
	if !ok {
		http.Error(w, "invalid approval id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	r.expireApprovals()

	r.mu.RLock()
	record, index := r.findApprovalLocked(approvalID)
	if index < 0 || !r.canAccessApproval(actor, record, "admin") {
		r.mu.RUnlock()
		http.NotFound(w, req)
		return
	}
	response := r.buildApprovalDetailLocked(record)
	r.mu.RUnlock()
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) handleAdminApprovalApprove(w http.ResponseWriter, req *http.Request) {
	r.transitionApproval(w, req, approvalStatusApproved)
}

func (r *Router) handleAdminApprovalReject(w http.ResponseWriter, req *http.Request) {
	r.transitionApproval(w, req, approvalStatusRejected)
}

func (r *Router) handleAdminApprovalExecute(w http.ResponseWriter, req *http.Request) {
	approvalID, ok := parseTailID(req.URL.Path, "/api/v1/admin/approvals/", "/execute")
	if !ok {
		http.Error(w, "invalid approval id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	response, err := r.executeApproval(approvalID, actor)
	if err != nil {
		writeApprovalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) listApprovals(w http.ResponseWriter, req *http.Request, scope string) {
	actor, status, message := r.resolveWorkspaceActor(req, scope)
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	r.expireApprovals()

	statusFilter := strings.TrimSpace(req.URL.Query().Get("status"))
	typeFilter := strings.TrimSpace(req.URL.Query().Get("type"))
	instanceID := 0
	if raw := strings.TrimSpace(req.URL.Query().Get("instanceId")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			instanceID = parsed
		}
	}
	tenantID := 0
	if raw := strings.TrimSpace(req.URL.Query().Get("tenantId")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			tenantID = parsed
		}
	}

	r.mu.RLock()
	items := make([]map[string]any, 0)
	for _, item := range r.data.Approvals {
		if !r.canAccessApproval(actor, item, scope) {
			continue
		}
		if statusFilter != "" && item.Status != statusFilter {
			continue
		}
		if typeFilter != "" && item.ApprovalType != typeFilter {
			continue
		}
		if instanceID > 0 && item.InstanceID != instanceID {
			continue
		}
		if tenantID > 0 && item.TenantID != tenantID {
			continue
		}
		items = append(items, r.buildApprovalSummaryLocked(item))
	}
	r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) createApproval(w http.ResponseWriter, req *http.Request, scope string) {
	actor, status, message := r.resolveWorkspaceActor(req, scope)
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	var payload approvalCreateRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	record, err := r.createApprovalRecord(actor, scope, payload)
	if err != nil {
		writeApprovalError(w, err)
		return
	}

	r.mu.RLock()
	response := r.buildApprovalDetailLocked(record)
	r.mu.RUnlock()
	writeJSON(w, http.StatusCreated, response)
}

func (r *Router) transitionApproval(w http.ResponseWriter, req *http.Request, nextStatus string) {
	var suffix string
	switch nextStatus {
	case approvalStatusApproved:
		suffix = "/approve"
	case approvalStatusRejected:
		suffix = "/reject"
	default:
		http.NotFound(w, req)
		return
	}

	approvalID, ok := parseTailID(req.URL.Path, "/api/v1/admin/approvals/", suffix)
	if !ok {
		http.Error(w, "invalid approval id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	var payload approvalDecisionRequest
	if req.ContentLength > 0 {
		if err := decodeJSON(req, &payload); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
	}

	record, err := r.updateApprovalStatus(approvalID, actor, nextStatus, payload)
	if err != nil {
		writeApprovalError(w, err)
		return
	}

	r.mu.RLock()
	response := r.buildApprovalDetailLocked(record)
	r.mu.RUnlock()
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) createApprovalRecord(actor workspaceActor, scope string, payload approvalCreateRequest) (models.ApprovalRecord, error) {
	approvalType := normalizeApprovalType(payload.ApprovalType)
	if approvalType == "" {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusBadRequest, Message: "approvalType is required"}
	}
	if strings.TrimSpace(payload.TargetType) == "" {
		payload.TargetType = "instance"
	}
	instanceID := payload.InstanceID
	if instanceID <= 0 && payload.TargetType == "instance" {
		instanceID = payload.TargetID
	}
	if instanceID <= 0 {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusBadRequest, Message: "instanceId is required"}
	}
	if strings.TrimSpace(payload.Reason) == "" {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusBadRequest, Message: "reason is required"}
	}

	r.mu.Lock()

	instance, found := r.findInstance(instanceID)
	if !found || !r.canAccessWorkspaceInstance(actor, instance, scope) {
		r.mu.Unlock()
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusNotFound, Message: "instance not found"}
	}

	now := nowRFC3339()
	record := models.ApprovalRecord{
		ID:              r.nextApprovalIDLocked(),
		ApprovalNo:      r.nextApprovalNoLocked(),
		TenantID:        instance.TenantID,
		InstanceID:      instanceID,
		ApprovalType:    approvalType,
		TargetType:      payload.TargetType,
		TargetID:        positiveOrFallback(payload.TargetID, instanceID),
		ApplicantID:     r.resolveApprovalApplicantIDLocked(actor, instance.TenantID),
		ApplicantName:   actor.identifier(),
		Status:          approvalStatusPending,
		RiskLevel:       normalizeRiskLevel(payload.RiskLevel, approvalType),
		Reason:          strings.TrimSpace(payload.Reason),
		ApprovalComment: strings.TrimSpace(payload.Comment),
		ExpiredAt:       time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
		CreatedAt:       now,
		UpdatedAt:       now,
		Metadata:        cloneWorkspaceAuditMetadata(payload.Metadata),
	}
	record.Metadata["requestedScope"] = scope
	record.Metadata["requestedByRole"] = normalizeWorkspaceRole(actor.Role)
	record.Metadata["instanceName"] = instance.Name

	r.data.Approvals = append([]models.ApprovalRecord{record}, r.data.Approvals...)
	r.appendApprovalActionLocked(record.ID, actor, "submitted", strings.TrimSpace(payload.Comment), map[string]string{
		"status":       approvalStatusPending,
		"approvalType": record.ApprovalType,
	})
	r.appendWorkspaceAuditLocked(actor, record.TenantID, record.InstanceID, "approval.request", "success", map[string]string{
		"approvalId":   strconv.Itoa(record.ID),
		"approvalNo":   record.ApprovalNo,
		"approvalType": record.ApprovalType,
		"riskLevel":    record.RiskLevel,
	})
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusInternalServerError, Message: "persist approval failed"}
	}
	return record, nil
}

func (r *Router) updateApprovalStatus(approvalID int, actor workspaceActor, nextStatus string, payload approvalDecisionRequest) (models.ApprovalRecord, error) {
	r.expireApprovals()

	r.mu.Lock()
	record, index := r.findApprovalLocked(approvalID)
	if index < 0 || !r.canAccessApproval(actor, record, "admin") {
		r.mu.Unlock()
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusNotFound, Message: "approval not found"}
	}
	if record.Status != approvalStatusPending && !(nextStatus == approvalStatusRejected && record.Status == approvalStatusApproved) {
		r.mu.Unlock()
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusConflict, Message: "approval is not pending"}
	}

	now := nowRFC3339()
	switch nextStatus {
	case approvalStatusApproved:
		r.data.Approvals[index].Status = approvalStatusApproved
		r.data.Approvals[index].ApproverID = actor.UserID
		r.data.Approvals[index].ApproverName = actor.identifier()
		r.data.Approvals[index].ApprovedAt = now
		r.data.Approvals[index].ApprovalComment = firstNonEmpty(strings.TrimSpace(payload.Comment), r.data.Approvals[index].ApprovalComment)
		r.data.Approvals[index].UpdatedAt = now
		r.data.Approvals[index].Metadata["approverName"] = actor.identifier()
		r.appendApprovalActionLocked(approvalID, actor, "approved", strings.TrimSpace(payload.Comment), map[string]string{"status": approvalStatusApproved})
		r.appendWorkspaceAuditLocked(actor, record.TenantID, record.InstanceID, "approval.approve", "success", map[string]string{
			"approvalId": strconv.Itoa(record.ID),
			"approvalNo": record.ApprovalNo,
		})
	case approvalStatusRejected:
		r.data.Approvals[index].Status = approvalStatusRejected
		r.data.Approvals[index].ApproverID = actor.UserID
		r.data.Approvals[index].ApproverName = actor.identifier()
		r.data.Approvals[index].RejectReason = firstNonEmpty(strings.TrimSpace(payload.RejectReason), strings.TrimSpace(payload.Comment), r.data.Approvals[index].RejectReason)
		r.data.Approvals[index].ApprovalComment = firstNonEmpty(strings.TrimSpace(payload.Comment), r.data.Approvals[index].ApprovalComment)
		r.data.Approvals[index].UpdatedAt = now
		r.data.Approvals[index].Metadata["approverName"] = actor.identifier()
		r.appendApprovalActionLocked(approvalID, actor, "rejected", firstNonEmpty(strings.TrimSpace(payload.RejectReason), strings.TrimSpace(payload.Comment)), map[string]string{"status": approvalStatusRejected})
		r.appendWorkspaceAuditLocked(actor, record.TenantID, record.InstanceID, "approval.reject", "success", map[string]string{
			"approvalId": strconv.Itoa(record.ID),
			"approvalNo": record.ApprovalNo,
		})
	default:
		r.mu.Unlock()
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusBadRequest, Message: "unsupported approval transition"}
	}
	record = r.data.Approvals[index]
	r.mu.Unlock()

	if err := r.persistAllData(); err != nil {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusInternalServerError, Message: "persist approval transition failed"}
	}
	return record, nil
}

func (r *Router) executeApproval(approvalID int, actor workspaceActor) (map[string]any, error) {
	r.expireApprovals()

	r.mu.Lock()
	record, index := r.findApprovalLocked(approvalID)
	if index < 0 || !r.canAccessApproval(actor, record, "admin") {
		r.mu.Unlock()
		return nil, approvalRequestError{StatusCode: http.StatusNotFound, Message: "approval not found"}
	}
	if record.Status != approvalStatusApproved {
		r.mu.Unlock()
		return nil, approvalRequestError{StatusCode: http.StatusConflict, Message: "approval is not approved"}
	}
	now := nowRFC3339()
	r.data.Approvals[index].Status = approvalStatusExecuting
	r.data.Approvals[index].UpdatedAt = now
	record = r.data.Approvals[index]
	r.mu.Unlock()

	var (
		response map[string]any
		err      error
	)
	switch record.ApprovalType {
	case approvalTypeConfigPublish:
		response, err = r.executeConfigPublishApproval(record, actor)
	case approvalTypeDeleteInstance:
		response, err = r.executeDeleteInstanceApproval(record, actor)
	case approvalTypeRuntimeStop:
		response, err = r.executeRuntimeApproval(record, actor, "stop")
	case approvalTypeRuntimeRestart:
		response, err = r.executeRuntimeApproval(record, actor, "restart")
	case approvalTypeRuntimeScale:
		response, err = r.executeRuntimeApproval(record, actor, "scale")
	case approvalTypeDiagnosticAccess:
		response, err = r.executeDiagnosticApproval(record, actor)
	default:
		err = approvalRequestError{StatusCode: http.StatusBadRequest, Message: "unsupported approval type"}
	}
	if err != nil {
		r.mu.Lock()
		if _, idx := r.findApprovalLocked(approvalID); idx >= 0 {
			r.data.Approvals[idx].Status = approvalStatusApproved
			r.data.Approvals[idx].UpdatedAt = nowRFC3339()
		}
		r.mu.Unlock()
		_ = r.persistAllData()
		return nil, err
	}

	r.mu.Lock()
	if _, idx := r.findApprovalLocked(approvalID); idx >= 0 {
		now = nowRFC3339()
		r.data.Approvals[idx].Status = approvalStatusExecuted
		r.data.Approvals[idx].ExecutorID = actor.UserID
		r.data.Approvals[idx].ExecutorName = actor.identifier()
		r.data.Approvals[idx].ExecutedAt = now
		r.data.Approvals[idx].UpdatedAt = now
		r.data.Approvals[idx].Metadata["executorName"] = actor.identifier()
		record = r.data.Approvals[idx]
	}
	r.appendApprovalActionLocked(approvalID, actor, "executed", "", map[string]string{"status": approvalStatusExecuted})
	r.appendWorkspaceAuditLocked(actor, record.TenantID, record.InstanceID, "approval.execute", "success", map[string]string{
		"approvalId":   strconv.Itoa(record.ID),
		"approvalNo":   record.ApprovalNo,
		"approvalType": record.ApprovalType,
	})
	r.mu.Unlock()

	if err := r.persistAllData(); err != nil {
		return nil, approvalRequestError{StatusCode: http.StatusInternalServerError, Message: "persist approval execution failed"}
	}

	r.mu.RLock()
	response["approval"] = r.buildApprovalSummaryLocked(record)
	response["actions"] = r.findApprovalActionsLocked(record.ID)
	r.mu.RUnlock()
	return response, nil
}

func (r *Router) executeConfigPublishApproval(record models.ApprovalRecord, actor workspaceActor) (map[string]any, error) {
	payload := updateConfigRequest{
		UpdatedBy: firstNonEmpty(record.Metadata["updatedBy"], actor.identifier()),
		Settings: models.ConfigSettings{
			Model:          record.Metadata["model"],
			AllowedOrigins: record.Metadata["allowedOrigins"],
			BackupPolicy:   record.Metadata["backupPolicy"],
		},
	}
	if strings.TrimSpace(payload.Settings.Model) == "" {
		payload.Settings.Model = "gpt-5"
	}
	return r.performConfigPublish(record.InstanceID, payload)
}

func (r *Router) executeDeleteInstanceApproval(record models.ApprovalRecord, actor workspaceActor) (map[string]any, error) {
	return r.performDeleteInstance(record.InstanceID, actor.identifier())
}

func (r *Router) executeRuntimeApproval(record models.ApprovalRecord, actor workspaceActor, action string) (map[string]any, error) {
	replicas := 0
	if action == "scale" {
		if value, err := strconv.Atoi(strings.TrimSpace(record.Metadata["replicas"])); err == nil {
			replicas = value
		}
		if replicas <= 0 {
			return nil, approvalRequestError{StatusCode: http.StatusBadRequest, Message: "approval metadata replicas is required"}
		}
	}
	return r.performAdminRuntimeAction(record.InstanceID, action, replicas, actor.identifier())
}

func (r *Router) executeDiagnosticApproval(record models.ApprovalRecord, actor workspaceActor) (map[string]any, error) {
	payload := diagnosticSessionCreateRequest{
		PodName:        record.Metadata["podName"],
		ContainerName:  record.Metadata["containerName"],
		AccessMode:     firstNonEmpty(record.Metadata["accessMode"], diagnosticAccessModeWhitelist),
		ApprovalTicket: record.ApprovalNo,
		ApprovedBy:     firstNonEmpty(record.ApproverName, actor.identifier()),
		Reason:         record.Reason,
	}
	session, err := r.performDiagnosticSessionCreate(record.InstanceID, actor, payload)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"session": r.buildDiagnosticSessionSummary(session),
		"policy":  buildDiagnosticPolicyPayload(),
	}, nil
}

func (r *Router) performConfigPublish(instanceID int, payload updateConfigRequest) (map[string]any, error) {
	if strings.TrimSpace(payload.UpdatedBy) == "" {
		payload.UpdatedBy = "portal-user"
	}

	r.mu.Lock()
	instanceIndex := r.findInstanceIndex(instanceID)
	if instanceIndex < 0 || !visibleInstance(r.data.Instances[instanceIndex]) {
		r.mu.Unlock()
		return nil, approvalRequestError{StatusCode: http.StatusNotFound, Message: "instance not found"}
	}

	now := nowRFC3339()
	configIndex := r.findConfigIndex(instanceID)
	version := 1
	if configIndex >= 0 {
		version = r.data.Configs[configIndex].Version + 1
		r.data.Configs[configIndex].Version = version
		r.data.Configs[configIndex].Hash = shortHash(instanceID, now)
		r.data.Configs[configIndex].PublishedAt = now
		r.data.Configs[configIndex].UpdatedBy = payload.UpdatedBy
		r.data.Configs[configIndex].Settings = payload.Settings
	} else {
		r.data.Configs = append([]models.InstanceConfig{{
			InstanceID:  instanceID,
			Version:     version,
			Hash:        shortHash(instanceID, now),
			PublishedAt: now,
			UpdatedBy:   payload.UpdatedBy,
			Settings:    payload.Settings,
		}}, r.data.Configs...)
		configIndex = 0
	}

	r.data.Instances[instanceIndex].UpdatedAt = now
	jobID := r.nextJobID()
	auditID := r.nextAuditID()
	job := models.Job{
		ID:         jobID,
		JobNo:      fmt.Sprintf("job-%d", 3000+jobID),
		Type:       "updateConfig",
		TargetType: "instance",
		TargetID:   instanceID,
		Status:     "succeeded",
		Summary:    "配置发布完成",
		StartedAt:  now,
		FinishedAt: now,
	}
	audit := models.AuditEvent{
		ID:        auditID,
		TenantID:  r.data.Instances[instanceIndex].TenantID,
		Actor:     payload.UpdatedBy,
		Action:    "updateConfig",
		Target:    "instance",
		TargetID:  instanceID,
		Result:    "success",
		CreatedAt: now,
		Metadata: map[string]string{
			"version": strconv.Itoa(version),
			"model":   payload.Settings.Model,
		},
	}

	r.data.Jobs = append([]models.Job{job}, r.data.Jobs...)
	r.data.Audits = append([]models.AuditEvent{audit}, r.data.Audits...)
	state, shouldPersist := r.snapshotInstanceStateLocked(instanceID)
	responseConfig := r.data.Configs[configIndex]
	r.mu.Unlock()

	if shouldPersist {
		if err := r.store.SaveInstanceState(state); err != nil {
			return nil, approvalRequestError{StatusCode: http.StatusInternalServerError, Message: "persist config state failed"}
		}
	}

	return map[string]any{
		"config": responseConfig,
		"job":    job,
	}, nil
}

func (r *Router) performDeleteInstance(instanceID int, actor string) (map[string]any, error) {
	r.mu.RLock()
	instance, found := r.findInstance(instanceID)
	binding := r.findRuntimeBinding(instanceID)
	r.mu.RUnlock()
	if !found || isDeletedInstance(instance) {
		return nil, approvalRequestError{StatusCode: http.StatusNotFound, Message: "instance not found"}
	}

	var (
		result runtimeadapter.ActionResult
		err    error
	)
	if binding != nil {
		if deleter, ok := r.runtime.(runtimeadapter.DeletionAdapter); ok {
			result, err = deleter.DeleteWorkload(runtimeadapter.DeleteWorkloadRequest{
				WorkloadID:      binding.WorkloadID,
				Namespace:       binding.Namespace,
				Name:            binding.WorkloadName,
				DeleteNamespace: true,
			})
			if err != nil {
				return nil, approvalRequestError{StatusCode: http.StatusBadGateway, Message: fmt.Sprintf("runtime delete failed: %v", err)}
			}
		}
	}

	now := nowRFC3339()
	r.mu.Lock()
	instanceIndex := r.findInstanceIndex(instanceID)
	if instanceIndex < 0 {
		r.mu.Unlock()
		return nil, approvalRequestError{StatusCode: http.StatusNotFound, Message: "instance not found"}
	}

	r.data.Instances[instanceIndex].Status = "deleted"
	r.data.Instances[instanceIndex].UpdatedAt = now
	r.removeRuntimeBindingLocked(instanceID)
	r.removeRuntimeLocked(instanceID)
	r.removeCredentialLocked(instanceID)
	r.removeAccessesLocked(instanceID)

	jobID := r.nextJobID()
	auditID := r.nextAuditID()
	job := models.Job{
		ID:         jobID,
		JobNo:      fmt.Sprintf("job-%d", 3000+jobID),
		Type:       "deleteInstance",
		TargetType: "instance",
		TargetID:   instanceID,
		Status:     deleteResultStatus(result),
		Summary:    "实例删除指令已下发",
		StartedAt:  now,
		FinishedAt: now,
	}
	audit := models.AuditEvent{
		ID:        auditID,
		TenantID:  r.data.Instances[instanceIndex].TenantID,
		Actor:     actor,
		Action:    "deleteInstance",
		Target:    "instance",
		TargetID:  instanceID,
		Result:    deleteAuditResult(result),
		CreatedAt: now,
	}
	if binding != nil {
		audit.Metadata = map[string]string{
			"namespace":  binding.Namespace,
			"workloadId": binding.WorkloadID,
		}
	}

	r.data.Jobs = append([]models.Job{job}, r.data.Jobs...)
	r.data.Audits = append([]models.AuditEvent{audit}, r.data.Audits...)
	state, shouldPersist := r.snapshotInstanceStateLocked(instanceID)
	responseInstance := r.data.Instances[instanceIndex]
	r.mu.Unlock()

	if shouldPersist {
		if err := r.store.SaveInstanceState(state); err != nil {
			return nil, approvalRequestError{StatusCode: http.StatusInternalServerError, Message: "persist delete state failed"}
		}
	}

	return map[string]any{
		"instance": responseInstance,
		"binding":  binding,
		"job":      job,
		"result":   result,
	}, nil
}

func (r *Router) performAdminRuntimeAction(instanceID int, action string, replicas int, actor string) (map[string]any, error) {
	r.mu.RLock()
	binding := r.findRuntimeBinding(instanceID)
	instance, found := r.findInstance(instanceID)
	r.mu.RUnlock()
	if !found || !visibleInstance(instance) {
		return nil, approvalRequestError{StatusCode: http.StatusNotFound, Message: "instance not found"}
	}
	if binding == nil {
		return nil, approvalRequestError{StatusCode: http.StatusNotFound, Message: "runtime binding not found"}
	}

	var (
		result      runtimeadapter.ActionResult
		actionFound bool
	)
	switch action {
	case "start":
		result, actionFound = r.runtime.StartWorkload(binding.WorkloadID)
	case "stop":
		result, actionFound = r.runtime.StopWorkload(binding.WorkloadID)
	case "scale":
		result, actionFound = r.runtime.ScaleWorkload(binding.WorkloadID, replicas)
	default:
		result, actionFound = r.runtime.RestartWorkload(binding.WorkloadID)
	}
	if !actionFound {
		return nil, approvalRequestError{StatusCode: http.StatusNotFound, Message: "runtime action target not found"}
	}

	now := nowRFC3339()
	r.mu.Lock()
	instanceIndex := r.findInstanceIndex(instanceID)
	if instanceIndex >= 0 {
		if workload, ok := r.runtime.GetWorkload(binding.WorkloadID); ok {
			r.data.Instances[instanceIndex].Status = instanceStatusFromWorkload(workload)
			r.upsertRuntimeLocked(buildInstanceRuntime(instanceID, r.data.Instances[instanceIndex].Spec, workload, nil))
			r.data.Instances[instanceIndex].UpdatedAt = now
		}
	}
	jobID := r.nextJobID()
	auditID := r.nextAuditID()
	job := models.Job{
		ID:         jobID,
		JobNo:      fmt.Sprintf("job-%d", 3000+jobID),
		Type:       "runtime." + action,
		TargetType: "instance",
		TargetID:   instanceID,
		Status:     actionResultStatus(result, actionFound),
		Summary:    runtimeActionSummary(action, replicas),
		StartedAt:  now,
		FinishedAt: now,
	}
	auditMetadata := map[string]string{
		"workloadId": binding.WorkloadID,
		"namespace":  binding.Namespace,
	}
	if replicas > 0 {
		auditMetadata["replicas"] = strconv.Itoa(replicas)
	}
	audit := models.AuditEvent{
		ID:        auditID,
		TenantID:  instance.TenantID,
		Actor:     actor,
		Action:    "runtime." + action,
		Target:    "instance",
		TargetID:  instanceID,
		Result:    "success",
		CreatedAt: now,
		Metadata:  auditMetadata,
	}
	r.data.Jobs = append([]models.Job{job}, r.data.Jobs...)
	r.data.Audits = append([]models.AuditEvent{audit}, r.data.Audits...)
	state, shouldPersist := r.snapshotInstanceStateLocked(instanceID)
	r.mu.Unlock()

	if shouldPersist {
		if err := r.store.SaveInstanceState(state); err != nil {
			return nil, approvalRequestError{StatusCode: http.StatusInternalServerError, Message: "persist runtime action failed"}
		}
	}

	return map[string]any{
		"binding": binding,
		"result":  result,
		"job":     job,
	}, nil
}

func (r *Router) performDiagnosticSessionCreate(instanceID int, actor workspaceActor, payload diagnosticSessionCreateRequest) (models.DiagnosticSession, error) {
	if !canManageDiagnosticSession(actor) {
		return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusForbidden, Message: "diagnostic session creation requires operator role"}
	}

	state, found := r.loadLiveInstanceState(instanceID)
	if !found {
		return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusNotFound, Message: "instance not found"}
	}
	if !r.canAccessWorkspaceInstance(actor, state.Instance, "admin") {
		return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusNotFound, Message: "instance not found"}
	}
	if state.Binding == nil {
		return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusConflict, Message: "diagnostic runtime binding is unavailable"}
	}

	accessMode := normalizeDiagnosticAccessMode(payload.AccessMode)
	if accessMode == diagnosticAccessModeWhitelist {
		if strings.TrimSpace(payload.ApprovalTicket) == "" {
			return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusBadRequest, Message: "approvalTicket is required for whitelist diagnostic sessions"}
		}
		if _, err := r.resolveApprovedApproval(strings.TrimSpace(payload.ApprovalTicket), approvalTypeDiagnosticAccess, instanceID); err != nil {
			return models.DiagnosticSession{}, err
		}
	}

	pods := r.runtime.ListPods(state.Binding.WorkloadID)
	selectedPod, ok := pickDiagnosticPod(pods, payload.PodName)
	if !ok {
		return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusBadRequest, Message: "target pod is unavailable"}
	}

	now := nowRFC3339()
	expiresAt := time.Now().UTC().Add(diagnosticSessionTTL(accessMode)).Format(time.RFC3339)

	r.mu.Lock()
	expiredMutation := r.expireDiagnosticSessionsLocked()
	if activeCount := r.countActiveDiagnosticSessionsLocked(instanceID); activeCount >= diagnosticMaxActiveSessionsPerInst {
		r.mu.Unlock()
		if len(expiredMutation.Sessions) > 0 || len(expiredMutation.Commands) > 0 {
			_ = r.persistDiagnosticsMutation(expiredMutation)
		}
		return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusConflict, Message: "too many active diagnostic sessions for this instance"}
	}

	session := models.DiagnosticSession{
		ID:             r.nextDiagnosticSessionIDLocked(),
		SessionNo:      r.nextDiagnosticSessionNoLocked(),
		TenantID:       state.Instance.TenantID,
		InstanceID:     state.Instance.ID,
		ClusterID:      state.Binding.ClusterID,
		Namespace:      state.Binding.Namespace,
		WorkloadID:     state.Binding.WorkloadID,
		WorkloadName:   state.Binding.WorkloadName,
		PodName:        selectedPod.Name,
		ContainerName:  strings.TrimSpace(payload.ContainerName),
		AccessMode:     accessMode,
		Status:         diagnosticSessionStatusActive,
		ApprovalTicket: strings.TrimSpace(payload.ApprovalTicket),
		ApprovedBy:     strings.TrimSpace(payload.ApprovedBy),
		Operator:       actor.identifier(),
		OperatorUserID: actor.UserID,
		Reason:         strings.TrimSpace(payload.Reason),
		ExpiresAt:      expiresAt,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.data.DiagnosticSessions = append([]models.DiagnosticSession{session}, r.data.DiagnosticSessions...)
	r.appendWorkspaceAuditLocked(actor, session.TenantID, session.InstanceID, "diagnostic.session.open", "success", map[string]string{
		"sessionId":      strconv.Itoa(session.ID),
		"sessionNo":      session.SessionNo,
		"podName":        session.PodName,
		"workloadId":     session.WorkloadID,
		"accessMode":     session.AccessMode,
		"approvalTicket": session.ApprovalTicket,
		"containerName":  session.ContainerName,
		"expiresAt":      session.ExpiresAt,
	})
	r.mu.Unlock()

	if len(expiredMutation.Sessions) > 0 || len(expiredMutation.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expiredMutation)
	}
	if err := r.persistAllData(); err != nil {
		return models.DiagnosticSession{}, approvalRequestError{StatusCode: http.StatusInternalServerError, Message: "persist diagnostic session failed"}
	}

	return session, nil
}

func (r *Router) resolveApprovedApproval(approvalNo string, approvalType string, instanceID int) (models.ApprovalRecord, error) {
	r.expireApprovals()

	r.mu.RLock()
	defer r.mu.RUnlock()

	record, index := r.findApprovalByNoLocked(approvalNo)
	if index < 0 {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusNotFound, Message: "approval not found"}
	}
	if record.Status != approvalStatusApproved && record.Status != approvalStatusExecuting {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusConflict, Message: "approval is not approved"}
	}
	if record.ApprovalType != approvalType {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusConflict, Message: "approval type mismatch"}
	}
	if record.InstanceID != instanceID {
		return models.ApprovalRecord{}, approvalRequestError{StatusCode: http.StatusConflict, Message: "approval target mismatch"}
	}
	return record, nil
}

func (r *Router) expireApprovals() {
	r.mu.Lock()

	now := time.Now().UTC()
	changed := false
	for index, item := range r.data.Approvals {
		if item.Status != approvalStatusPending && item.Status != approvalStatusApproved {
			continue
		}
		expiresAt, ok := parseRFC3339(item.ExpiredAt)
		if !ok || now.Before(expiresAt) {
			continue
		}
		r.data.Approvals[index].Status = approvalStatusExpired
		r.data.Approvals[index].UpdatedAt = nowRFC3339()
		changed = true
	}
	r.mu.Unlock()
	if changed {
		_ = r.persistAllData()
	}
}

func (r *Router) canAccessApproval(actor workspaceActor, record models.ApprovalRecord, scope string) bool {
	return actor.canAccessTenant(scope, record.TenantID)
}

func (r *Router) buildApprovalSummaryLocked(record models.ApprovalRecord) map[string]any {
	return map[string]any{
		"id":              record.ID,
		"approvalNo":      record.ApprovalNo,
		"tenantId":        record.TenantID,
		"instanceId":      record.InstanceID,
		"approvalType":    record.ApprovalType,
		"targetType":      record.TargetType,
		"targetId":        record.TargetID,
		"applicantId":     record.ApplicantID,
		"applicantName":   record.ApplicantName,
		"approverId":      record.ApproverID,
		"approverName":    record.ApproverName,
		"executorId":      record.ExecutorID,
		"executorName":    record.ExecutorName,
		"status":          record.Status,
		"riskLevel":       record.RiskLevel,
		"reason":          record.Reason,
		"approvalComment": record.ApprovalComment,
		"rejectReason":    record.RejectReason,
		"approvedAt":      record.ApprovedAt,
		"executedAt":      record.ExecutedAt,
		"expiredAt":       record.ExpiredAt,
		"createdAt":       record.CreatedAt,
		"updatedAt":       record.UpdatedAt,
		"metadata":        record.Metadata,
	}
}

func (r *Router) buildApprovalDetailLocked(record models.ApprovalRecord) map[string]any {
	response := map[string]any{
		"approval": r.buildApprovalSummaryLocked(record),
		"actions":  r.findApprovalActionsLocked(record.ID),
	}
	if instance, found := r.findInstance(record.InstanceID); found {
		response["instance"] = instance
	}
	return response
}

func (r *Router) findApprovalActionsLocked(approvalID int) []models.ApprovalAction {
	items := make([]models.ApprovalAction, 0)
	for _, item := range r.data.ApprovalActions {
		if item.ApprovalID == approvalID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) appendApprovalActionLocked(approvalID int, actor workspaceActor, action string, comment string, metadata map[string]string) {
	r.data.ApprovalActions = append(r.data.ApprovalActions, models.ApprovalAction{
		ID:         r.nextApprovalActionIDLocked(),
		ApprovalID: approvalID,
		ActorID:    actor.UserID,
		ActorName:  actor.identifier(),
		Action:     action,
		Comment:    comment,
		CreatedAt:  nowRFC3339(),
		Metadata:   cloneWorkspaceAuditMetadata(metadata),
	})
}

func (r *Router) findApprovalLocked(id int) (models.ApprovalRecord, int) {
	for index, item := range r.data.Approvals {
		if item.ID == id {
			return item, index
		}
	}
	return models.ApprovalRecord{}, -1
}

func (r *Router) findApprovalByNoLocked(approvalNo string) (models.ApprovalRecord, int) {
	normalized := strings.TrimSpace(approvalNo)
	for index, item := range r.data.Approvals {
		if strings.TrimSpace(item.ApprovalNo) == normalized {
			return item, index
		}
	}
	return models.ApprovalRecord{}, -1
}

func (r *Router) nextApprovalIDLocked() int {
	maxID := 0
	for _, item := range r.data.Approvals {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextApprovalActionIDLocked() int {
	maxID := 0
	for _, item := range r.data.ApprovalActions {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextApprovalNoLocked() string {
	return fmt.Sprintf("APR-%s-%03d", time.Now().UTC().Format("20060102"), r.nextApprovalIDLocked())
}

func normalizeApprovalType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case approvalTypeDeleteInstance:
		return approvalTypeDeleteInstance
	case approvalTypeRuntimeStop, "runtime_pause":
		return approvalTypeRuntimeStop
	case approvalTypeRuntimeRestart:
		return approvalTypeRuntimeRestart
	case approvalTypeRuntimeScale:
		return approvalTypeRuntimeScale
	case approvalTypeConfigPublish:
		return approvalTypeConfigPublish
	case approvalTypeDiagnosticAccess:
		return approvalTypeDiagnosticAccess
	default:
		return ""
	}
}

func normalizeRiskLevel(value string, approvalType string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "medium", "high", "critical":
		return strings.ToLower(strings.TrimSpace(value))
	}
	switch approvalType {
	case approvalTypeDeleteInstance:
		return "critical"
	case approvalTypeRuntimeScale, approvalTypeRuntimeRestart, approvalTypeConfigPublish, approvalTypeDiagnosticAccess:
		return "high"
	default:
		return "medium"
	}
}

func runtimeActionSummary(action string, replicas int) string {
	switch action {
	case "stop":
		return "运行时停止指令已下发"
	case "scale":
		return fmt.Sprintf("运行时扩缩容到 %d 副本", replicas)
	default:
		return "运行时重启指令已下发"
	}
}

func (r *Router) resolveApprovalApplicantIDLocked(actor workspaceActor, tenantID int) int {
	if actor.UserID > 0 {
		return actor.UserID
	}
	for _, item := range r.data.Users {
		if item.TenantID == tenantID {
			return item.ID
		}
	}
	for _, item := range r.data.Users {
		if item.ID > 0 {
			return item.ID
		}
	}
	return 1
}

func positiveOrFallback(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func writeApprovalError(w http.ResponseWriter, err error) {
	var requestErr approvalRequestError
	if ok := errorAsApprovalRequest(err, &requestErr); ok {
		http.Error(w, requestErr.Message, requestErr.StatusCode)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func errorAsApprovalRequest(err error, target *approvalRequestError) bool {
	if err == nil || target == nil {
		return false
	}
	value, ok := err.(approvalRequestError)
	if !ok {
		return false
	}
	*target = value
	return true
}
