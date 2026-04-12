package httpapi

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

const (
	diagnosticAccessModeReadonly         = "readonly"
	diagnosticAccessModeWhitelist        = "whitelist"
	diagnosticSessionStatusActive        = "active"
	diagnosticSessionStatusClosed        = "closed"
	diagnosticSessionStatusExpired       = "expired"
	diagnosticCommandStatusSucceeded     = "succeeded"
	diagnosticCommandStatusFailed        = "failed"
	diagnosticCommandStatusBlocked       = "blocked"
	diagnosticCommandStatusTimeout       = "timeout"
	diagnosticReadonlyTTLMinutes         = 15
	diagnosticWhitelistTTLMinutes        = 10
	diagnosticMaxActiveSessionsPerInst   = 3
	diagnosticCommandOutputMaxChars      = 12000
	diagnosticCommandErrorOutputMaxChars = 4000
)

var diagnosticExecutorRoles = map[string]struct{}{
	"tenant_admin":    {},
	"tenant_operator": {},
	"platform_admin":  {},
	"platform_ops":    {},
	"super_admin":     {},
}

type diagnosticSessionCreateRequest struct {
	PodName        string `json:"podName"`
	ContainerName  string `json:"containerName"`
	AccessMode     string `json:"accessMode"`
	ApprovalTicket string `json:"approvalTicket"`
	ApprovedBy     string `json:"approvedBy"`
	Reason         string `json:"reason"`
}

type diagnosticCommandExecuteRequest struct {
	CommandKey  string `json:"commandKey"`
	CommandText string `json:"commandText"`
}

type diagnosticSessionCloseRequest struct {
	Reason string `json:"reason"`
}

type diagnosticCommandSpec struct {
	Key         string
	Label       string
	Description string
	Command     []string
	Aliases     []string
}

func diagnosticCommandCatalog() []diagnosticCommandSpec {
	return []diagnosticCommandSpec{
		{
			Key:         "process.list",
			Label:       "进程列表",
			Description: "查看容器内主要进程与资源占用。",
			Command:     []string{"ps", "aux"},
			Aliases:     []string{"ps", "ps aux"},
		},
		{
			Key:         "disk.usage",
			Label:       "磁盘占用",
			Description: "查看主要挂载点的空间水位。",
			Command:     []string{"df", "-h"},
			Aliases:     []string{"df -h"},
		},
		{
			Key:         "memory.usage",
			Label:       "内存快照",
			Description: "查看容器当前内存用量。",
			Command:     []string{"free", "-m"},
			Aliases:     []string{"free -m"},
		},
		{
			Key:         "env.list",
			Label:       "环境变量",
			Description: "查看运行时环境变量清单。",
			Command:     []string{"printenv"},
			Aliases:     []string{"printenv", "env"},
		},
		{
			Key:         "network.sockets",
			Label:       "网络套接字",
			Description: "查看监听端口与已建立连接。",
			Command:     []string{"ss", "-tunlp"},
			Aliases:     []string{"ss -tunlp"},
		},
		{
			Key:         "kernel.info",
			Label:       "内核信息",
			Description: "查看内核与体系架构信息。",
			Command:     []string{"uname", "-a"},
			Aliases:     []string{"uname -a"},
		},
		{
			Key:         "os.release",
			Label:       "系统版本",
			Description: "查看容器基础镜像发行版信息。",
			Command:     []string{"cat", "/etc/os-release"},
			Aliases:     []string{"cat /etc/os-release"},
		},
		{
			Key:         "uptime",
			Label:       "运行时长",
			Description: "查看容器节点当前运行时长与负载。",
			Command:     []string{"uptime"},
			Aliases:     []string{"uptime"},
		},
	}
}

func (r *Router) handleAdminInstanceDiagnostics(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseInstanceID(req.URL.Path, "/api/v1/admin/instances/", "/diagnostics")
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	expired := r.expireDiagnosticSessions()
	if len(expired.Sessions) > 0 || len(expired.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expired)
	}

	state, found := r.loadLiveInstanceState(instanceID)
	if !found {
		http.NotFound(w, req)
		return
	}
	if !r.canAccessWorkspaceInstance(actor, state.Instance, "admin") {
		http.NotFound(w, req)
		return
	}
	if state.Binding == nil {
		http.Error(w, "diagnostic runtime binding is unavailable", http.StatusConflict)
		return
	}

	pods := r.runtime.ListPods(state.Binding.WorkloadID)
	writeJSON(w, http.StatusOK, map[string]any{
		"instance": state.Instance,
		"binding":  state.Binding,
		"workload": state.Workload,
		"metrics":  state.Metrics,
		"pods":     buildDiagnosticPodItems(state.Workload, pods),
		"signals":  r.buildDiagnosticSignals(instanceID, pods),
		"policy":   buildDiagnosticPolicyPayload(),
		"sessions": r.buildDiagnosticSessionSummaries(instanceID),
	})
}

func (r *Router) handleAdminDiagnosticSessions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleAdminDiagnosticSessionIndex(w, req)
	case http.MethodPost:
		r.handleAdminDiagnosticSessionCreate(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleAdminTerminalSessions(w http.ResponseWriter, req *http.Request) {
	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	expired := r.expireDiagnosticSessions()
	if len(expired.Sessions) > 0 || len(expired.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expired)
	}

	statusFilter := strings.TrimSpace(req.URL.Query().Get("status"))
	instanceID := 0
	if raw := strings.TrimSpace(req.URL.Query().Get("instanceId")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			instanceID = parsed
		}
	}
	query := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("q")))

	r.mu.RLock()
	items := make([]map[string]any, 0)
	for _, session := range r.data.DiagnosticSessions {
		instance, found := r.findInstance(session.InstanceID)
		if !found || !r.canAccessWorkspaceInstance(actor, instance, "admin") {
			continue
		}
		if statusFilter != "" && session.Status != statusFilter {
			continue
		}
		if instanceID > 0 && session.InstanceID != instanceID {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(strings.Join([]string{
				session.SessionNo,
				session.PodName,
				session.Operator,
				session.WorkloadName,
				session.Reason,
			}, " "))
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		items = append(items, r.buildDiagnosticSessionSummary(session))
	}
	r.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func buildDiagnosticPolicyPayload() map[string]any {
	return map[string]any{
		"defaultAccessMode":            diagnosticAccessModeReadonly,
		"readonlyTtlMinutes":           diagnosticReadonlyTTLMinutes,
		"whitelistTtlMinutes":          diagnosticWhitelistTTLMinutes,
		"requiresApprovalForWhitelist": true,
		"maxActiveSessionsPerInstance": diagnosticMaxActiveSessionsPerInst,
		"commandCatalog":               buildDiagnosticCommandCatalogPayload(diagnosticAccessModeWhitelist),
	}
}

func buildDiagnosticCommandCatalogPayload(accessMode string) []map[string]any {
	catalog := diagnosticCommandCatalog()
	items := make([]map[string]any, 0, len(catalog))
	for _, item := range catalog {
		items = append(items, map[string]any{
			"key":           item.Key,
			"label":         item.Label,
			"description":   item.Description,
			"commandText":   strings.Join(item.Command, " "),
			"manualAllowed": accessMode == diagnosticAccessModeWhitelist,
		})
	}
	return items
}

func buildDiagnosticPodItems(workload *runtimeadapter.Workload, pods []runtimeadapter.Pod) []map[string]any {
	items := make([]map[string]any, 0, len(pods))
	for _, pod := range pods {
		items = append(items, map[string]any{
			"id":         pod.ID,
			"name":       pod.Name,
			"nodeName":   pod.NodeName,
			"status":     pod.Status,
			"restarts":   pod.Restarts,
			"startedAt":  pod.StartedAt,
			"workloadId": pod.WorkloadID,
			"image":      diagnosticWorkloadImage(workload),
		})
	}
	return items
}

func (r *Router) handleAdminDiagnosticSessionIndex(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseDiagnosticInstanceID(req.URL.Path)
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	r.mu.RLock()
	instance, found := r.findInstance(instanceID)
	r.mu.RUnlock()
	if !found || !r.canAccessWorkspaceInstance(actor, instance, "admin") {
		http.NotFound(w, req)
		return
	}

	expired := r.expireDiagnosticSessions()
	if len(expired.Sessions) > 0 || len(expired.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expired)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": r.buildDiagnosticSessionSummaries(instanceID),
	})
}

func (r *Router) handleAdminDiagnosticSessionCreate(w http.ResponseWriter, req *http.Request) {
	instanceID, ok := parseDiagnosticInstanceID(req.URL.Path)
	if !ok {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}
	if !canManageDiagnosticSession(actor) {
		http.Error(w, "diagnostic session creation requires operator role", http.StatusForbidden)
		return
	}

	var payload diagnosticSessionCreateRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	session, err := r.performDiagnosticSessionCreate(instanceID, actor, payload)
	if err != nil {
		writeApprovalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"session": r.buildDiagnosticSessionSummary(session),
		"policy":  buildDiagnosticPolicyPayload(),
	})
}

func (r *Router) handleAdminDiagnosticSessionDetail(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseDiagnosticSessionID(req.URL.Path)
	if !ok {
		http.Error(w, "invalid diagnostic session id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	session, commands, instance, found, allowed, expired := r.lookupDiagnosticSessionDetail(sessionID, actor)
	if len(expired.Sessions) > 0 || len(expired.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expired)
	}
	if !found || !allowed {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session":        r.buildDiagnosticSessionSummary(session),
		"commands":       commands,
		"record":         buildDiagnosticTranscript(commands),
		"commandCatalog": buildDiagnosticCommandCatalogPayload(session.AccessMode),
		"instance":       instance,
	})
}

func (r *Router) handleAdminDiagnosticSessionCommands(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseDiagnosticSessionID(req.URL.Path, "/commands")
	if !ok {
		http.Error(w, "invalid diagnostic session id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}
	if !canManageDiagnosticSession(actor) {
		http.Error(w, "diagnostic command execution requires operator role", http.StatusForbidden)
		return
	}

	var payload diagnosticCommandExecuteRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	session, instance, found, allowed, expired := r.lookupDiagnosticSession(sessionID, actor)
	if len(expired.Sessions) > 0 || len(expired.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expired)
	}
	if !found || !allowed {
		http.NotFound(w, req)
		return
	}
	if !r.canAccessWorkspaceInstance(actor, instance, "admin") {
		http.NotFound(w, req)
		return
	}
	if session.Status != diagnosticSessionStatusActive {
		http.Error(w, "diagnostic session is not active", http.StatusConflict)
		return
	}

	commandSpec, command, commandText, blockedMessage, resolveStatus := resolveDiagnosticCommand(session.AccessMode, payload)
	if resolveStatus != 0 {
		record := r.recordDiagnosticCommandBlocked(session, commandSpec.Key, commandText, blockedMessage)
		writeJSON(w, resolveStatus, map[string]any{
			"command": record,
			"error":   blockedMessage,
		})
		return
	}

	execProvider, ok := r.runtime.(runtimeadapter.DiagnosticExecProvider)
	if !ok {
		http.Error(w, "runtime provider does not support diagnostic execution", http.StatusServiceUnavailable)
		return
	}

	result, execErr := execProvider.ExecuteDiagnosticCommand(runtimeadapter.DiagnosticExecRequest{
		WorkloadID:     session.WorkloadID,
		Namespace:      session.Namespace,
		PodName:        session.PodName,
		ContainerName:  session.ContainerName,
		Command:        command,
		TimeoutSeconds: 8,
	})
	if execErr != nil && result.PodName == "" && result.Namespace == "" {
		http.Error(w, execErr.Error(), http.StatusBadGateway)
		return
	}

	record := buildDiagnosticCommandRecordFromExec(session, commandSpec.Key, commandText, result, execErr)
	r.mu.Lock()
	record.ID = r.nextDiagnosticCommandIDLocked()
	sessionIndex := r.findDiagnosticSessionIndexLocked(session.ID)
	if sessionIndex >= 0 {
		r.data.DiagnosticSessions[sessionIndex].LastCommandAt = record.ExecutedAt
		r.data.DiagnosticSessions[sessionIndex].UpdatedAt = record.ExecutedAt
		session = r.data.DiagnosticSessions[sessionIndex]
	}
	r.data.DiagnosticCommandRecords = append(r.data.DiagnosticCommandRecords, record)
	r.mu.Unlock()

	if err := r.persistDiagnosticsMutation(corestore.DiagnosticsMutation{
		Sessions: []models.DiagnosticSession{session},
		Commands: []models.DiagnosticCommandRecord{record},
	}); err != nil {
		http.Error(w, "persist diagnostic command failed", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"session": r.buildDiagnosticSessionSummary(session),
		"command": record,
	}
	if execErr != nil {
		response["error"] = execErr.Error()
	}
	writeJSON(w, http.StatusOK, response)
}

func (r *Router) handleAdminDiagnosticSessionClose(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseDiagnosticSessionID(req.URL.Path, "/close")
	if !ok {
		http.Error(w, "invalid diagnostic session id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}
	if !canManageDiagnosticSession(actor) {
		http.Error(w, "diagnostic session close requires operator role", http.StatusForbidden)
		return
	}

	var payload diagnosticSessionCloseRequest
	if req.ContentLength > 0 {
		if err := decodeJSON(req, &payload); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
	}

	session, instance, found, allowed, expired := r.lookupDiagnosticSession(sessionID, actor)
	if len(expired.Sessions) > 0 || len(expired.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expired)
	}
	if !found || !allowed {
		http.NotFound(w, req)
		return
	}
	if !r.canAccessWorkspaceInstance(actor, instance, "admin") {
		http.NotFound(w, req)
		return
	}

	now := nowRFC3339()
	r.mu.Lock()
	sessionIndex := r.findDiagnosticSessionIndexLocked(sessionID)
	if sessionIndex < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}
	if r.data.DiagnosticSessions[sessionIndex].Status == diagnosticSessionStatusActive {
		r.data.DiagnosticSessions[sessionIndex].Status = diagnosticSessionStatusClosed
		r.data.DiagnosticSessions[sessionIndex].EndedAt = now
		r.data.DiagnosticSessions[sessionIndex].CloseReason = firstNonEmpty(strings.TrimSpace(payload.Reason), "operator_closed")
		r.data.DiagnosticSessions[sessionIndex].UpdatedAt = now
	}
	session = r.data.DiagnosticSessions[sessionIndex]
	r.appendWorkspaceAuditLocked(actor, session.TenantID, session.InstanceID, "diagnostic.session.close", "success", map[string]string{
		"sessionId":     strconv.Itoa(session.ID),
		"sessionNo":     session.SessionNo,
		"closeReason":   session.CloseReason,
		"lastCommandAt": session.LastCommandAt,
	})
	r.mu.Unlock()

	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist diagnostic session close failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session": r.buildDiagnosticSessionSummary(session),
	})
}

func (r *Router) handleAdminDiagnosticSessionRecord(w http.ResponseWriter, req *http.Request) {
	sessionID, ok := parseDiagnosticSessionID(req.URL.Path, "/record")
	if !ok {
		http.Error(w, "invalid diagnostic session id", http.StatusBadRequest)
		return
	}

	actor, status, message := r.resolveWorkspaceActor(req, "admin")
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	session, commands, _, found, allowed, expired := r.lookupDiagnosticSessionDetail(sessionID, actor)
	if len(expired.Sessions) > 0 || len(expired.Commands) > 0 {
		_ = r.persistDiagnosticsMutation(expired)
	}
	if !found || !allowed {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session":    r.buildDiagnosticSessionSummary(session),
		"items":      commands,
		"transcript": buildDiagnosticTranscript(commands),
	})
}

func (r *Router) buildDiagnosticSignals(instanceID int, pods []runtimeadapter.Pod) []map[string]any {
	signals := make([]map[string]any, 0)
	for _, alert := range r.filterAlertsByInstance(instanceID) {
		signals = append(signals, map[string]any{
			"type":        "alert",
			"severity":    alert.Severity,
			"summary":     alert.Summary,
			"triggeredAt": alert.TriggeredAt,
		})
	}
	for _, pod := range pods {
		if pod.Restarts == 0 {
			continue
		}
		signals = append(signals, map[string]any{
			"type":     "pod_restart",
			"severity": "warning",
			"summary":  fmt.Sprintf("Pod %s 已重启 %d 次", pod.Name, pod.Restarts),
			"podName":  pod.Name,
			"restarts": pod.Restarts,
		})
	}
	if len(signals) == 0 {
		signals = append(signals, map[string]any{
			"type":     "info",
			"severity": "info",
			"summary":  "当前未发现新的诊断风险信号。",
		})
	}
	return signals
}

func (r *Router) buildDiagnosticSessionSummaries(instanceID int) []map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]map[string]any, 0)
	for _, session := range r.data.DiagnosticSessions {
		if session.InstanceID != instanceID {
			continue
		}
		items = append(items, r.buildDiagnosticSessionSummary(session))
	}
	return items
}

func (r *Router) buildDiagnosticSessionSummary(session models.DiagnosticSession) map[string]any {
	commandCount := 0
	lastCommandStatus := ""
	lastCommandText := ""
	for _, record := range r.data.DiagnosticCommandRecords {
		if record.SessionID != session.ID {
			continue
		}
		commandCount++
		lastCommandStatus = record.Status
		lastCommandText = record.CommandText
	}
	return map[string]any{
		"id":                session.ID,
		"sessionNo":         session.SessionNo,
		"tenantId":          session.TenantID,
		"instanceId":        session.InstanceID,
		"clusterId":         session.ClusterID,
		"namespace":         session.Namespace,
		"workloadId":        session.WorkloadID,
		"workloadName":      session.WorkloadName,
		"podName":           session.PodName,
		"containerName":     session.ContainerName,
		"accessMode":        session.AccessMode,
		"status":            session.Status,
		"approvalTicket":    session.ApprovalTicket,
		"approvedBy":        session.ApprovedBy,
		"operator":          session.Operator,
		"operatorUserId":    session.OperatorUserID,
		"reason":            session.Reason,
		"closeReason":       session.CloseReason,
		"expiresAt":         session.ExpiresAt,
		"lastCommandAt":     session.LastCommandAt,
		"startedAt":         session.StartedAt,
		"endedAt":           session.EndedAt,
		"createdAt":         session.CreatedAt,
		"updatedAt":         session.UpdatedAt,
		"commandCount":      commandCount,
		"lastCommandStatus": lastCommandStatus,
		"lastCommandText":   lastCommandText,
	}
}

func (r *Router) lookupDiagnosticSession(sessionID int, actor workspaceActor) (models.DiagnosticSession, models.Instance, bool, bool, corestore.DiagnosticsMutation) {
	r.mu.Lock()
	expired := r.expireDiagnosticSessionsLocked()
	session, found := r.findDiagnosticSessionLocked(sessionID)
	instance, instanceFound := r.findInstance(session.InstanceID)
	r.mu.Unlock()
	if !found || !instanceFound {
		return models.DiagnosticSession{}, models.Instance{}, false, false, expired
	}
	allowed := r.canAccessWorkspaceInstance(actor, instance, "admin")
	return session, instance, true, allowed, expired
}

func (r *Router) lookupDiagnosticSessionDetail(sessionID int, actor workspaceActor) (models.DiagnosticSession, []models.DiagnosticCommandRecord, models.Instance, bool, bool, corestore.DiagnosticsMutation) {
	session, instance, found, allowed, expired := r.lookupDiagnosticSession(sessionID, actor)
	if !found {
		return models.DiagnosticSession{}, nil, models.Instance{}, false, false, expired
	}
	r.mu.RLock()
	commands := r.findDiagnosticCommandRecordsLocked(sessionID)
	r.mu.RUnlock()
	return session, commands, instance, true, allowed, expired
}

func (r *Router) recordDiagnosticCommandBlocked(session models.DiagnosticSession, commandKey string, commandText string, reason string) models.DiagnosticCommandRecord {
	now := nowRFC3339()
	record := models.DiagnosticCommandRecord{
		CommandKey:  commandKey,
		CommandText: firstNonEmpty(strings.TrimSpace(commandText), "<blocked>"),
		Status:      diagnosticCommandStatusBlocked,
		ExitCode:    126,
		ErrorOutput: reason,
		ExecutedAt:  now,
	}

	r.mu.Lock()
	record.ID = r.nextDiagnosticCommandIDLocked()
	record.SessionID = session.ID
	record.TenantID = session.TenantID
	record.InstanceID = session.InstanceID
	sessionIndex := r.findDiagnosticSessionIndexLocked(session.ID)
	if sessionIndex >= 0 {
		r.data.DiagnosticSessions[sessionIndex].LastCommandAt = now
		r.data.DiagnosticSessions[sessionIndex].UpdatedAt = now
		session = r.data.DiagnosticSessions[sessionIndex]
	}
	r.data.DiagnosticCommandRecords = append(r.data.DiagnosticCommandRecords, record)
	r.mu.Unlock()

	_ = r.persistDiagnosticsMutation(corestore.DiagnosticsMutation{
		Sessions: []models.DiagnosticSession{session},
		Commands: []models.DiagnosticCommandRecord{record},
	})
	return record
}

func buildDiagnosticCommandRecordFromExec(
	session models.DiagnosticSession,
	commandKey string,
	commandText string,
	result runtimeadapter.DiagnosticExecResult,
	execErr error,
) models.DiagnosticCommandRecord {
	executedAt := firstNonEmpty(result.FinishedAt, result.StartedAt, nowRFC3339())
	output, outputTruncated := trimDiagnosticText(result.Stdout, diagnosticCommandOutputMaxChars)
	errorOutput, errorTruncated := trimDiagnosticText(result.Stderr, diagnosticCommandErrorOutputMaxChars)
	status := diagnosticCommandStatusSucceeded
	if execErr != nil {
		if result.ExitCode == 124 {
			status = diagnosticCommandStatusTimeout
		} else {
			status = diagnosticCommandStatusFailed
		}
	}
	if result.ExitCode != 0 && execErr == nil {
		status = diagnosticCommandStatusFailed
	}
	if execErr != nil && errorOutput == "" {
		errorOutput = execErr.Error()
	}

	return models.DiagnosticCommandRecord{
		SessionID:       session.ID,
		TenantID:        session.TenantID,
		InstanceID:      session.InstanceID,
		CommandKey:      commandKey,
		CommandText:     commandText,
		Status:          status,
		ExitCode:        result.ExitCode,
		DurationMs:      result.DurationMs,
		Output:          output,
		ErrorOutput:     errorOutput,
		OutputTruncated: outputTruncated || errorTruncated || result.OutputTruncated,
		ExecutedAt:      executedAt,
	}
}

func buildDiagnosticTranscript(commands []models.DiagnosticCommandRecord) string {
	if len(commands) == 0 {
		return ""
	}
	blocks := make([]string, 0, len(commands))
	for _, item := range commands {
		block := "$ " + item.CommandText
		if item.Output != "" {
			block += "\n" + item.Output
		}
		if item.ErrorOutput != "" {
			block += "\n" + item.ErrorOutput
		}
		block += fmt.Sprintf("\n[status=%s exit=%d durationMs=%d at=%s]", item.Status, item.ExitCode, item.DurationMs, item.ExecutedAt)
		blocks = append(blocks, block)
	}
	return strings.Join(blocks, "\n\n")
}

func resolveDiagnosticCommand(accessMode string, payload diagnosticCommandExecuteRequest) (diagnosticCommandSpec, []string, string, string, int) {
	if strings.TrimSpace(payload.CommandKey) != "" {
		spec, ok := findDiagnosticCommandByKey(strings.TrimSpace(payload.CommandKey))
		if !ok {
			return diagnosticCommandSpec{}, nil, "", "commandKey is not supported", http.StatusBadRequest
		}
		return spec, append([]string(nil), spec.Command...), strings.Join(spec.Command, " "), "", 0
	}
	if accessMode != diagnosticAccessModeWhitelist {
		return diagnosticCommandSpec{}, nil, strings.TrimSpace(payload.CommandText), "readonly diagnostic sessions only accept commandKey", http.StatusForbidden
	}
	commandText := normalizeDiagnosticCommandText(payload.CommandText)
	if commandText == "" {
		return diagnosticCommandSpec{}, nil, "", "commandKey or commandText is required", http.StatusBadRequest
	}
	spec, ok := findDiagnosticCommandByAlias(commandText)
	if !ok {
		return diagnosticCommandSpec{}, nil, commandText, "commandText is outside the approved whitelist", http.StatusForbidden
	}
	return spec, append([]string(nil), spec.Command...), strings.Join(spec.Command, " "), "", 0
}

func findDiagnosticCommandByKey(key string) (diagnosticCommandSpec, bool) {
	for _, item := range diagnosticCommandCatalog() {
		if item.Key == key {
			return item, true
		}
	}
	return diagnosticCommandSpec{}, false
}

func findDiagnosticCommandByAlias(commandText string) (diagnosticCommandSpec, bool) {
	for _, item := range diagnosticCommandCatalog() {
		if strings.EqualFold(strings.Join(item.Command, " "), commandText) {
			return item, true
		}
		for _, alias := range item.Aliases {
			if strings.EqualFold(alias, commandText) {
				return item, true
			}
		}
	}
	return diagnosticCommandSpec{}, false
}

func normalizeDiagnosticAccessMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case diagnosticAccessModeWhitelist:
		return diagnosticAccessModeWhitelist
	default:
		return diagnosticAccessModeReadonly
	}
}

func parseDiagnosticInstanceID(path string) (int, bool) {
	if id, ok := parseInstanceID(path, "/api/v1/admin/instances/", "/diagnostic-sessions"); ok {
		return id, true
	}
	return parseInstanceID(path, "/api/v1/admin/instances/", "/terminal-sessions")
}

func parseDiagnosticSessionID(path string, suffix ...string) (int, bool) {
	if id, ok := parseTailID(path, "/api/v1/admin/diagnostic-sessions/", suffix...); ok {
		return id, true
	}
	return parseTailID(path, "/api/v1/admin/terminal-sessions/", suffix...)
}

func normalizeDiagnosticCommandText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func diagnosticSessionTTL(accessMode string) time.Duration {
	if accessMode == diagnosticAccessModeWhitelist {
		return diagnosticWhitelistTTLMinutes * time.Minute
	}
	return diagnosticReadonlyTTLMinutes * time.Minute
}

func pickDiagnosticPod(pods []runtimeadapter.Pod, requested string) (runtimeadapter.Pod, bool) {
	if len(pods) == 0 {
		return runtimeadapter.Pod{}, false
	}
	requested = strings.TrimSpace(requested)
	if requested != "" {
		for _, pod := range pods {
			if pod.Name == requested {
				return pod, true
			}
		}
		return runtimeadapter.Pod{}, false
	}
	for _, pod := range pods {
		if strings.EqualFold(pod.Status, "running") {
			return pod, true
		}
	}
	return pods[0], true
}

func diagnosticWorkloadImage(workload *runtimeadapter.Workload) string {
	if workload == nil {
		return ""
	}
	return workload.Image
}

func canManageDiagnosticSession(actor workspaceActor) bool {
	_, ok := diagnosticExecutorRoles[normalizeWorkspaceRole(actor.Role)]
	return ok
}

func trimDiagnosticText(value string, limit int) (string, bool) {
	if limit <= 0 || len(value) <= limit {
		return value, false
	}
	if limit <= 3 {
		return value[:limit], true
	}
	return value[:limit-3] + "...", true
}

func (r *Router) expireDiagnosticSessions() corestore.DiagnosticsMutation {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.expireDiagnosticSessionsLocked()
}

func (r *Router) expireDiagnosticSessionsLocked() corestore.DiagnosticsMutation {
	now := nowRFC3339()
	currentTime := time.Now().UTC()
	mutation := corestore.DiagnosticsMutation{
		Sessions: []models.DiagnosticSession{},
	}
	for index, item := range r.data.DiagnosticSessions {
		if item.Status != diagnosticSessionStatusActive {
			continue
		}
		expiresAt, ok := parseRFC3339(item.ExpiresAt)
		if !ok || currentTime.Before(expiresAt) {
			continue
		}
		r.data.DiagnosticSessions[index].Status = diagnosticSessionStatusExpired
		if r.data.DiagnosticSessions[index].EndedAt == "" {
			r.data.DiagnosticSessions[index].EndedAt = expiresAt.Format(time.RFC3339)
		}
		if strings.TrimSpace(r.data.DiagnosticSessions[index].CloseReason) == "" {
			r.data.DiagnosticSessions[index].CloseReason = "ttl_expired"
		}
		r.data.DiagnosticSessions[index].UpdatedAt = now
		mutation.Sessions = append(mutation.Sessions, r.data.DiagnosticSessions[index])
	}
	return mutation
}

func (r *Router) countActiveDiagnosticSessionsLocked(instanceID int) int {
	count := 0
	for _, item := range r.data.DiagnosticSessions {
		if item.InstanceID == instanceID && item.Status == diagnosticSessionStatusActive {
			count++
		}
	}
	return count
}

func (r *Router) findDiagnosticSessionLocked(id int) (models.DiagnosticSession, bool) {
	for _, item := range r.data.DiagnosticSessions {
		if item.ID == id {
			return item, true
		}
	}
	return models.DiagnosticSession{}, false
}

func (r *Router) findDiagnosticSessionIndexLocked(id int) int {
	for index, item := range r.data.DiagnosticSessions {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (r *Router) findDiagnosticCommandRecordsLocked(sessionID int) []models.DiagnosticCommandRecord {
	items := make([]models.DiagnosticCommandRecord, 0)
	for _, item := range r.data.DiagnosticCommandRecords {
		if item.SessionID == sessionID {
			items = append(items, item)
		}
	}
	slices.SortFunc(items, func(a models.DiagnosticCommandRecord, b models.DiagnosticCommandRecord) int {
		return strings.Compare(a.ExecutedAt, b.ExecutedAt)
	})
	return items
}

func (r *Router) nextDiagnosticSessionIDLocked() int {
	maxID := 0
	for _, item := range r.data.DiagnosticSessions {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextDiagnosticCommandIDLocked() int {
	maxID := 0
	for _, item := range r.data.DiagnosticCommandRecords {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextDiagnosticSessionNoLocked() string {
	return fmt.Sprintf("DG-%s-%03d", time.Now().UTC().Format("20060102"), r.nextDiagnosticSessionIDLocked())
}
