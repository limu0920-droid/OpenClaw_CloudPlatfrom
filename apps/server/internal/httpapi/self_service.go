package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
)

type selfServiceQuotaPreset struct {
	Requests24h int
	Tokens24h   int
}

type selfServiceOverviewMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta,omitempty"`
	Tone  string `json:"tone,omitempty"`
}

type selfServiceStep struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ActionLabel string `json:"actionLabel"`
	ActionPath  string `json:"actionPath"`
	Result      string `json:"result,omitempty"`
}

type selfServiceQuota struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Used      float64 `json:"used"`
	Limit     float64 `json:"limit"`
	Unit      string  `json:"unit"`
	Percent   int     `json:"percent"`
	Status    string  `json:"status"`
	Detail    string  `json:"detail"`
	UsedText  string  `json:"usedText"`
	LimitText string  `json:"limitText"`
}

type selfServiceReminder struct {
	Key         string `json:"key"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ActionLabel string `json:"actionLabel"`
	ActionPath  string `json:"actionPath"`
	At          string `json:"at,omitempty"`
}

type selfServiceSessionItem struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	InstanceID    int    `json:"instanceId"`
	InstanceName  string `json:"instanceName"`
	SessionNo     string `json:"sessionNo"`
	Status        string `json:"status"`
	UpdatedAt     string `json:"updatedAt"`
	MessageCount  int    `json:"messageCount"`
	ArtifactCount int    `json:"artifactCount"`
	WorkspacePath string `json:"workspacePath"`
}

type portalArtifactCenterItem struct {
	ID               int                                `json:"id"`
	Title            string                             `json:"title"`
	Kind             string                             `json:"kind"`
	SourceURL        string                             `json:"sourceUrl"`
	PreviewURL       string                             `json:"previewUrl,omitempty"`
	ArchiveStatus    string                             `json:"archiveStatus"`
	ContentType      string                             `json:"contentType,omitempty"`
	SizeBytes        int64                              `json:"sizeBytes"`
	Filename         string                             `json:"filename,omitempty"`
	CreatedAt        string                             `json:"createdAt"`
	UpdatedAt        string                             `json:"updatedAt"`
	InstanceID       int                                `json:"instanceId"`
	InstanceName     string                             `json:"instanceName"`
	InstanceStatus   string                             `json:"instanceStatus"`
	SessionID        int                                `json:"sessionId"`
	SessionNo        string                             `json:"sessionNo"`
	SessionTitle     string                             `json:"sessionTitle"`
	SessionStatus    string                             `json:"sessionStatus"`
	MessageID        int                                `json:"messageId,omitempty"`
	MessagePreview   string                             `json:"messagePreview,omitempty"`
	TenantID         int                                `json:"tenantId"`
	TenantName       string                             `json:"tenantName,omitempty"`
	ViewCount        int                                `json:"viewCount"`
	DownloadCount    int                                `json:"downloadCount"`
	WorkspacePath    string                             `json:"workspacePath"`
	DetailPath       string                             `json:"detailPath,omitempty"`
	LineageKey       string                             `json:"lineageKey,omitempty"`
	Version          int                                `json:"version"`
	LatestVersion    int                                `json:"latestVersion"`
	ParentArtifactID int                                `json:"parentArtifactId,omitempty"`
	IsFavorite       bool                               `json:"isFavorite,omitempty"`
	FavoriteCount    int                                `json:"favoriteCount"`
	ShareCount       int                                `json:"shareCount"`
	Thumbnail        artifactThumbnailDescriptor        `json:"thumbnail"`
	Quality          artifactQualitySummary             `json:"quality"`
	Preview          workspaceArtifactPreviewDescriptor `json:"preview"`
}

func (r *Router) handlePortalSelfService(w http.ResponseWriter, req *http.Request) {
	tenantID := tenantFilterID(req, 1)

	r.mu.RLock()
	tenant := r.findTenant(tenantID)
	account := r.findAccountSettings(tenantID)
	brandBinding := r.findBrandBindingByTenant(tenantID)
	rawInstances := r.filterInstancesByTenant(tenantID)
	sessions := r.filterWorkspaceSessionsByTenant(tenantID)
	artifacts := r.filterWorkspaceArtifactsByTenant(tenantID)
	messages := r.filterWorkspaceMessagesByTenant(tenantID)
	subscriptions := r.filterSubscriptionsByTenant(tenantID)
	wallet := r.findWallet(tenantID)
	alerts := r.filterAlertsByTenant(tenantID)
	planOffers := append([]models.PlanOffer(nil), r.data.PlanOffers...)
	r.mu.RUnlock()

	if tenant == nil {
		http.NotFound(w, req)
		return
	}

	brand := &models.OEMBrand{}
	if brandBinding != nil {
		if resolved := r.findBrand(brandBinding.BrandID); resolved != nil {
			brand = resolved
		}
	}

	instances := r.resolveLiveInstances(rawInstances)
	instanceMap := make(map[int]models.Instance, len(instances))
	workspaceURLByInstance := make(map[int]string, len(instances))
	runtimeByInstance := make(map[int]*models.InstanceRuntime, len(instances))
	accessByInstance := make(map[int][]models.InstanceAccess, len(instances))
	for _, instance := range instances {
		instanceMap[instance.ID] = instance
		r.mu.RLock()
		accessByInstance[instance.ID] = append([]models.InstanceAccess(nil), r.filterAccessByInstance(instance.ID)...)
		runtimeByInstance[instance.ID] = r.findRuntime(instance.ID)
		r.mu.RUnlock()
		if entry := primaryWorkspaceAccess(accessByInstance[instance.ID]); entry != nil {
			workspaceURLByInstance[instance.ID] = entry.URL
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sortByTimeDesc(sessions[i].UpdatedAt, sessions[j].UpdatedAt)
	})
	sort.Slice(artifacts, func(i, j int) bool {
		return sortByTimeDesc(artifacts[i].UpdatedAt, artifacts[j].UpdatedAt)
	})

	primaryInstance := pickPrimarySelfServiceInstance(instances, workspaceURLByInstance)
	messageCountBySession := make(map[int]int)
	userMessageCountBySession := make(map[int]int)
	artifactCountBySession := make(map[int]int)
	for _, item := range messages {
		messageCountBySession[item.SessionID]++
		if item.Role == "user" {
			userMessageCountBySession[item.SessionID]++
		}
	}
	for _, item := range artifacts {
		artifactCountBySession[item.SessionID]++
	}

	recentSessions := make([]selfServiceSessionItem, 0, minInt(4, len(sessions)))
	for _, session := range sessions {
		instance, ok := instanceMap[session.InstanceID]
		if !ok {
			continue
		}
		recentSessions = append(recentSessions, selfServiceSessionItem{
			ID:            session.ID,
			Title:         session.Title,
			InstanceID:    session.InstanceID,
			InstanceName:  instance.Name,
			SessionNo:     session.SessionNo,
			Status:        session.Status,
			UpdatedAt:     session.UpdatedAt,
			MessageCount:  messageCountBySession[session.ID],
			ArtifactCount: artifactCountBySession[session.ID],
			WorkspacePath: portalWorkspacePath(session.InstanceID, ""),
		})
		if len(recentSessions) == 4 {
			break
		}
	}

	recentArtifacts := buildPortalArtifactItems(artifacts, sessions, instanceMap)
	if len(recentArtifacts) > 6 {
		recentArtifacts = recentArtifacts[:6]
	}

	onboarding := buildSelfServiceOnboarding(primaryInstance, account, sessions, recentArtifacts, userMessageCountBySession)
	quotas := buildSelfServiceQuotas(instances, runtimeByInstance, planOffers)
	reminders := buildSelfServiceReminders(primaryInstance, runtimeByInstance, subscriptions, wallet, alerts, instanceMap)

	runningCount := 0
	for _, instance := range instances {
		if instance.Status == "running" {
			runningCount++
		}
	}
	totalMessageCount := 0
	for _, count := range messageCountBySession {
		totalMessageCount += count
	}

	response := map[string]any{
		"tenant": map[string]any{
			"id":           tenant.ID,
			"code":         tenant.Code,
			"name":         tenant.Name,
			"plan":         tenant.Plan,
			"status":       tenant.Status,
			"expiredAt":    tenant.ExpiredAt,
			"supportEmail": strings.TrimSpace(brand.SupportEmail),
			"supportUrl":   strings.TrimSpace(brand.SupportURL),
		},
		"launchpad": map[string]any{
			"primaryInstanceId":   0,
			"primaryInstanceName": "",
			"workspacePath":       "/portal/instances",
			"artifactsPath":       "/portal/artifacts",
			"workspaceUrl":        "",
		},
		"experience": map[string]any{
			"portalHeadline":   buildPortalHeadline(tenant, account, brand),
			"portalSubtitle":   buildPortalSubtitle(account, brand),
			"workspaceCallout": buildWorkspaceCallout(account, brand),
			"experimentBadge":  buildExperimentBadge(account, brand),
		},
		"onboarding": onboarding,
		"metrics": []selfServiceOverviewMetric{
			{
				Label: "运行实例",
				Value: strconv.Itoa(runningCount),
				Delta: fmt.Sprintf("总实例 %d", len(instances)),
				Tone:  toneFromCount(runningCount, len(instances)),
			},
			{
				Label: "平台会话",
				Value: strconv.Itoa(len(sessions)),
				Delta: fmt.Sprintf("%d 条消息留痕", totalMessageCount),
				Tone:  "positive",
			},
			{
				Label: "已归档产物",
				Value: strconv.Itoa(len(recentArtifacts)),
				Delta: fmt.Sprintf("全量产物 %d 个", len(artifacts)),
				Tone:  "neutral",
			},
			{
				Label: "待处理提醒",
				Value: strconv.Itoa(len(reminders)),
				Delta: "到期、容量与告警统一收口",
				Tone:  toneFromReminderCount(len(reminders)),
			},
		},
		"quotas":          quotas,
		"reminders":       reminders,
		"recentSessions":  recentSessions,
		"recentArtifacts": recentArtifacts,
	}

	if primaryInstance != nil {
		response["launchpad"] = map[string]any{
			"primaryInstanceId":   primaryInstance.ID,
			"primaryInstanceName": primaryInstance.Name,
			"workspacePath":       portalWorkspacePath(primaryInstance.ID, ""),
			"artifactsPath":       "/portal/artifacts",
			"workspaceUrl":        workspaceURLByInstance[primaryInstance.ID],
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func (r *Router) resolveLiveInstances(raw []models.Instance) []models.Instance {
	instances := make([]models.Instance, 0, len(raw))
	for _, item := range raw {
		state, ok := r.loadLiveInstanceState(item.ID)
		if ok {
			instances = append(instances, state.Instance)
			continue
		}
		instances = append(instances, item)
	}
	return instances
}

func (r *Router) filterWorkspaceSessionsByTenant(tenantID int) []models.WorkspaceSession {
	items := make([]models.WorkspaceSession, 0)
	for _, item := range r.data.WorkspaceSessions {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) filterWorkspaceArtifactsByTenant(tenantID int) []models.WorkspaceArtifact {
	items := make([]models.WorkspaceArtifact, 0)
	for _, item := range r.data.WorkspaceArtifacts {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) filterWorkspaceMessagesByTenant(tenantID int) []models.WorkspaceMessage {
	items := make([]models.WorkspaceMessage, 0)
	for _, item := range r.data.WorkspaceMessages {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items
}

func (r *Router) filterSubscriptionsByTenant(tenantID int) []models.Subscription {
	items := make([]models.Subscription, 0)
	for _, item := range r.data.Subscriptions {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items
}

func pickPrimarySelfServiceInstance(instances []models.Instance, workspaceURLByInstance map[int]string) *models.Instance {
	if len(instances) == 0 {
		return nil
	}
	for _, item := range instances {
		if item.Status == "running" && strings.TrimSpace(workspaceURLByInstance[item.ID]) != "" {
			copy := item
			return &copy
		}
	}
	for _, item := range instances {
		if strings.TrimSpace(workspaceURLByInstance[item.ID]) != "" {
			copy := item
			return &copy
		}
	}
	copy := instances[0]
	return &copy
}

func buildSelfServiceOnboarding(
	primaryInstance *models.Instance,
	account *models.AccountSettings,
	sessions []models.WorkspaceSession,
	artifacts []portalArtifactCenterItem,
	userMessageCountBySession map[int]int,
) map[string]any {
	hasInstance := primaryInstance != nil
	hasSession := len(sessions) > 0
	hasUserMessage := false
	for _, count := range userMessageCountBySession {
		if count > 0 {
			hasUserMessage = true
			break
		}
	}
	hasArtifact := len(artifacts) > 0
	hasNotificationConfig := account != nil && account.NotifyOnAlert && account.NotifyOnExpiry

	steps := []selfServiceStep{
		{
			Key:         "instance",
			Title:       "确认实例已开通",
			Description: "进入实例列表核对实例状态、地域和套餐是否可直接使用。",
			Status:      ternaryStatus(hasInstance, "completed", "pending"),
			ActionLabel: "查看实例",
			ActionPath:  "/portal/instances",
			Result:      ternaryText(hasInstance, "已找到可用实例", ""),
		},
		{
			Key:         "workspace",
			Title:       "一键进入龙虾工作台",
			Description: "从 Portal 内直接进入实例工作台，不再依赖人工查找访问地址。",
			Status:      deriveReadyStatus(hasInstance, hasSession),
			ActionLabel: "进入工作台",
			ActionPath:  firstPath(hasInstance, portalWorkspacePath(instanceIDOrZero(primaryInstance), ""), "/portal/instances"),
			Result:      ternaryText(hasSession, "已建立平台侧工作台会话", ternaryText(hasInstance, "实例已具备入口，可直接进入", "")),
		},
		{
			Key:         "message",
			Title:       "完成首条平台侧对话",
			Description: "发送第一条问题并让平台保存消息留痕，后续可继续会话与审计。",
			Status:      deriveReadyStatus(hasSession, hasUserMessage),
			ActionLabel: "继续对话",
			ActionPath:  firstPath(hasSession || hasInstance, portalWorkspacePath(instanceIDOrZero(primaryInstance), ""), "/portal/instances"),
			Result:      ternaryText(hasUserMessage, "平台侧已记录用户消息", ternaryText(hasSession, "会话已就绪，等待首条消息", "")),
		},
		{
			Key:         "artifact",
			Title:       "查看平台产物中心",
			Description: "把网页、PPTX、PDF、文档和表格统一收口到平台侧，方便回看与交付。",
			Status:      deriveReadyStatus(hasUserMessage || hasSession, hasArtifact),
			ActionLabel: "打开产物中心",
			ActionPath:  "/portal/artifacts",
			Result:      ternaryText(hasArtifact, "产物已可在平台内统一回看", ternaryText(hasUserMessage || hasSession, "已有会话，可继续产出并归档", "")),
		},
		{
			Key:         "reminder",
			Title:       "确认到期与告警提醒",
			Description: "启用到期、容量和异常提醒，避免实例或配额在无通知状态下失效。",
			Status:      ternaryStatus(hasNotificationConfig, "completed", "ready"),
			ActionLabel: "查看提醒",
			ActionPath:  "/portal",
			Result:      ternaryText(hasNotificationConfig, "提醒通知已启用", "Portal 已显示到期与容量提醒，建议同步通知策略"),
		},
	}

	completedCount := 0
	for _, item := range steps {
		if item.Status == "completed" {
			completedCount++
		}
	}

	return map[string]any{
		"showGuide":       completedCount < len(steps),
		"completedCount":  completedCount,
		"totalCount":      len(steps),
		"steps":           steps,
		"isReadyForUsage": hasInstance && hasSession && hasUserMessage && hasArtifact,
	}
}

func buildSelfServiceQuotas(
	instances []models.Instance,
	runtimeByInstance map[int]*models.InstanceRuntime,
	planOffers []models.PlanOffer,
) []selfServiceQuota {
	storageByPlan := make(map[string]float64, len(planOffers))
	for _, offer := range planOffers {
		storageByPlan[offer.Code] = parseStorageGi(offer.Storage)
	}

	var requestsUsed float64
	var requestsLimit float64
	var tokensUsed float64
	var tokensLimit float64
	var storageUsed float64
	var storageLimit float64

	for _, instance := range instances {
		preset := quotaPresetForPlan(instance.Plan)
		requestsLimit += float64(preset.Requests24h)
		tokensLimit += float64(preset.Tokens24h)
		storageLimitGi := storageByPlan[instance.Plan]
		if storageLimitGi == 0 {
			storageLimitGi = 20
		}
		storageLimit += storageLimitGi

		runtime := runtimeByInstance[instance.ID]
		if runtime == nil {
			continue
		}
		requestsUsed += float64(runtime.APIRequests24h)
		tokensUsed += float64(runtime.APITokens24h)
		storageUsed += storageLimitGi * float64(runtime.DiskUsagePercent) / 100
	}

	return []selfServiceQuota{
		buildSelfServiceQuota("requests", "24h 请求量", requestsUsed, requestsLimit, "req", "实例请求总量与套餐上限"),
		buildSelfServiceQuota("tokens", "24h Token", tokensUsed, tokensLimit, "token", "模型调用消耗与套餐上限"),
		buildSelfServiceQuota("storage", "归档容量", storageUsed, storageLimit, "GiB", "按实例磁盘利用率折算归档容量"),
	}
}

func buildSelfServiceQuota(key string, label string, used float64, limit float64, unit string, detail string) selfServiceQuota {
	percent := 0
	if limit > 0 {
		percent = int(used * 100 / limit)
		if percent > 100 {
			percent = 100
		}
	}
	status := "healthy"
	if percent >= 90 {
		status = "critical"
	} else if percent >= 75 {
		status = "warning"
	}
	return selfServiceQuota{
		Key:       key,
		Label:     label,
		Used:      used,
		Limit:     limit,
		Unit:      unit,
		Percent:   percent,
		Status:    status,
		Detail:    detail,
		UsedText:  formatQuotaValue(used, unit),
		LimitText: formatQuotaValue(limit, unit),
	}
}

func buildSelfServiceReminders(
	primaryInstance *models.Instance,
	runtimeByInstance map[int]*models.InstanceRuntime,
	subscriptions []models.Subscription,
	wallet *models.WalletBalance,
	alerts []models.Alert,
	instanceMap map[int]models.Instance,
) []selfServiceReminder {
	now := time.Now().UTC()
	items := make([]selfServiceReminder, 0)

	for _, subscription := range subscriptions {
		checkpoint := firstNonEmpty(subscription.ExpiredAt, subscription.CurrentPeriodEnd)
		if checkpoint == "" {
			continue
		}
		deadline, ok := parseRFC3339(checkpoint)
		if !ok {
			continue
		}
		instanceName := fmt.Sprintf("实例 #%d", subscription.InstanceID)
		if instance, found := instanceMap[subscription.InstanceID]; found {
			instanceName = instance.Name
		}
		switch {
		case !deadline.After(now):
			items = append(items, selfServiceReminder{
				Key:         fmt.Sprintf("subscription-expired-%d", subscription.ID),
				Severity:    "critical",
				Title:       fmt.Sprintf("%s 订阅已到期", instanceName),
				Description: fmt.Sprintf("订阅已于 %s 到期，建议立即续费并核对实例可用性。", checkpoint),
				ActionLabel: "处理续费",
				ActionPath:  portalInstancePath(subscription.InstanceID),
				At:          checkpoint,
			})
		case deadline.Sub(now) <= 7*24*time.Hour:
			items = append(items, selfServiceReminder{
				Key:         fmt.Sprintf("subscription-expiring-%d", subscription.ID),
				Severity:    "warning",
				Title:       fmt.Sprintf("%s 即将到期", instanceName),
				Description: fmt.Sprintf("订阅将在 %s 到期，建议提前续费避免服务中断。", checkpoint),
				ActionLabel: "查看实例",
				ActionPath:  portalInstancePath(subscription.InstanceID),
				At:          checkpoint,
			})
		}
	}

	for instanceID, runtime := range runtimeByInstance {
		if runtime == nil {
			continue
		}
		instance := instanceMap[instanceID]
		if runtime.DiskUsagePercent >= 80 {
			items = append(items, selfServiceReminder{
				Key:         fmt.Sprintf("disk-%d", instanceID),
				Severity:    "critical",
				Title:       fmt.Sprintf("%s 磁盘使用率过高", instance.Name),
				Description: fmt.Sprintf("磁盘使用率已到 %d%%，建议清理旧产物或升级套餐容量。", runtime.DiskUsagePercent),
				ActionLabel: "查看产物",
				ActionPath:  "/portal/artifacts",
				At:          runtime.LastSeenAt,
			})
		}
		if runtime.MemoryUsagePercent >= 75 {
			items = append(items, selfServiceReminder{
				Key:         fmt.Sprintf("memory-%d", instanceID),
				Severity:    "warning",
				Title:       fmt.Sprintf("%s 内存接近上限", instance.Name),
				Description: fmt.Sprintf("内存使用率为 %d%%，建议观察模型负载与并发情况。", runtime.MemoryUsagePercent),
				ActionLabel: "查看实例",
				ActionPath:  portalInstancePath(instanceID),
				At:          runtime.LastSeenAt,
			})
		}
	}

	openAlertCount := 0
	for _, alert := range alerts {
		if alert.Status == "open" {
			openAlertCount++
		}
	}
	if openAlertCount > 0 && primaryInstance != nil {
		items = append(items, selfServiceReminder{
			Key:         "alerts-open",
			Severity:    "warning",
			Title:       "近期存在未关闭告警",
			Description: fmt.Sprintf("当前租户还有 %d 条打开告警，建议在继续自助使用前先确认实例健康。", openAlertCount),
			ActionLabel: "查看日志",
			ActionPath:  "/portal/logs",
		})
	}

	if wallet != nil && wallet.AvailableAmount <= 500 {
		items = append(items, selfServiceReminder{
			Key:         "wallet-low",
			Severity:    "info",
			Title:       "账户余额偏低",
			Description: fmt.Sprintf("当前可用余额仅剩 %d %s，建议预留续费和临时扩容空间。", wallet.AvailableAmount, wallet.Currency),
			ActionLabel: "查看实例",
			ActionPath:  firstPath(primaryInstance != nil, portalInstancePath(instanceIDOrZero(primaryInstance)), "/portal/instances"),
			At:          wallet.UpdatedAt,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		left := severityRank(items[i].Severity)
		right := severityRank(items[j].Severity)
		if left != right {
			return left > right
		}
		return sortByTimeDesc(items[i].At, items[j].At)
	})

	if len(items) > 5 {
		return items[:5]
	}
	return items
}

func buildPortalArtifactItems(
	artifacts []models.WorkspaceArtifact,
	sessions []models.WorkspaceSession,
	instanceMap map[int]models.Instance,
) []portalArtifactCenterItem {
	sessionMap := make(map[int]models.WorkspaceSession, len(sessions))
	for _, session := range sessions {
		sessionMap[session.ID] = session
	}

	items := make([]portalArtifactCenterItem, 0, len(artifacts))
	for _, artifact := range artifacts {
		session, ok := sessionMap[artifact.SessionID]
		if !ok {
			continue
		}
		instance, ok := instanceMap[artifact.InstanceID]
		if !ok {
			continue
		}
		items = append(items, portalArtifactCenterItem{
			ID:             artifact.ID,
			Title:          artifact.Title,
			Kind:           artifact.Kind,
			SourceURL:      artifact.SourceURL,
			PreviewURL:     artifact.PreviewURL,
			ArchiveStatus:  artifact.ArchiveStatus,
			ContentType:    artifact.ContentType,
			SizeBytes:      artifact.SizeBytes,
			Filename:       artifact.Filename,
			CreatedAt:      artifact.CreatedAt,
			UpdatedAt:      artifact.UpdatedAt,
			InstanceID:     artifact.InstanceID,
			InstanceName:   instance.Name,
			InstanceStatus: instance.Status,
			SessionID:      session.ID,
			SessionNo:      session.SessionNo,
			SessionTitle:   session.Title,
			SessionStatus:  session.Status,
			MessageID:      artifact.MessageID,
			TenantID:       artifact.TenantID,
			WorkspacePath:  portalWorkspacePath(artifact.InstanceID, firstNonEmpty(artifact.PreviewURL, artifact.SourceURL)),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return sortByTimeDesc(items[i].UpdatedAt, items[j].UpdatedAt)
	})

	return items
}

func quotaPresetForPlan(plan string) selfServiceQuotaPreset {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "trial":
		return selfServiceQuotaPreset{Requests24h: 10000, Tokens24h: 500000}
	case "standard":
		return selfServiceQuotaPreset{Requests24h: 30000, Tokens24h: 1500000}
	default:
		return selfServiceQuotaPreset{Requests24h: 80000, Tokens24h: 4000000}
	}
}

func parseStorageGi(value string) float64 {
	raw := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.HasSuffix(raw, "gib"):
		parsed, _ := strconv.ParseFloat(strings.TrimSuffix(raw, "gib"), 64)
		return parsed
	case strings.HasSuffix(raw, "gi"):
		parsed, _ := strconv.ParseFloat(strings.TrimSuffix(raw, "gi"), 64)
		return parsed
	case strings.HasSuffix(raw, "gb"):
		parsed, _ := strconv.ParseFloat(strings.TrimSuffix(raw, "gb"), 64)
		return parsed
	default:
		parsed, _ := strconv.ParseFloat(raw, 64)
		return parsed
	}
}

func formatQuotaValue(value float64, unit string) string {
	switch unit {
	case "GiB":
		return fmt.Sprintf("%.1f %s", value, unit)
	case "token":
		return fmt.Sprintf("%.0f %s", value, unit)
	default:
		return fmt.Sprintf("%.0f %s", value, unit)
	}
}

func portalWorkspacePath(instanceID int, artifactURL string) string {
	path := fmt.Sprintf("/portal/instances/%d/workspace", instanceID)
	if strings.TrimSpace(artifactURL) == "" {
		return path
	}
	return path + "?artifact=" + url.QueryEscape(artifactURL)
}

func portalInstancePath(instanceID int) string {
	return fmt.Sprintf("/portal/instances/%d", instanceID)
}

func instanceIDOrZero(instance *models.Instance) int {
	if instance == nil {
		return 0
	}
	return instance.ID
}

func buildPortalHeadline(tenant *models.Tenant, account *models.AccountSettings, brand *models.OEMBrand) string {
	if account != nil && strings.TrimSpace(account.PortalHeadline) != "" {
		return strings.TrimSpace(account.PortalHeadline)
	}
	if brand != nil && strings.TrimSpace(brand.Name) != "" {
		return fmt.Sprintf("%s 自助控制台", strings.TrimSpace(brand.Name))
	}
	if tenant != nil && strings.TrimSpace(tenant.Name) != "" {
		return fmt.Sprintf("%s 自助控制台", strings.TrimSpace(tenant.Name))
	}
	return "Portal Self-Service Loop"
}

func buildPortalSubtitle(account *models.AccountSettings, brand *models.OEMBrand) string {
	if account != nil && strings.TrimSpace(account.PortalSubtitle) != "" {
		return strings.TrimSpace(account.PortalSubtitle)
	}
	if brand != nil && strings.TrimSpace(brand.SupportURL) != "" {
		return fmt.Sprintf("通过 %s 品牌门户统一进入工作台、产物、账单与运营视图。", strings.TrimSpace(brand.Name))
	}
	return "从首登引导、进入工作台、查看产物，到配额监控和到期提醒，当前入口已把用户自助链路收口到同一页面。"
}

func buildWorkspaceCallout(account *models.AccountSettings, brand *models.OEMBrand) string {
	if account != nil && strings.TrimSpace(account.WorkspaceCallout) != "" {
		return strings.TrimSpace(account.WorkspaceCallout)
	}
	if brand != nil && strings.TrimSpace(brand.Name) != "" {
		return fmt.Sprintf("通过 %s 直接进入品牌化工作台与交付物。", strings.TrimSpace(brand.Name))
	}
	return "一键进入工作台并继续最近的交付上下文。"
}

func buildExperimentBadge(account *models.AccountSettings, brand *models.OEMBrand) string {
	if account != nil && strings.TrimSpace(account.ExperimentBadge) != "" {
		return strings.TrimSpace(account.ExperimentBadge)
	}
	if brand != nil && !strings.EqualFold(strings.TrimSpace(brand.Code), "openclaw") {
		return "OEM Launch"
	}
	return "Self-Service"
}

func severityRank(value string) int {
	switch value {
	case "critical":
		return 3
	case "warning":
		return 2
	default:
		return 1
	}
}

func sortByTimeDesc(left string, right string) bool {
	leftTime, leftOK := parseRFC3339(left)
	rightTime, rightOK := parseRFC3339(right)
	switch {
	case leftOK && rightOK:
		return leftTime.After(rightTime)
	case leftOK:
		return true
	case rightOK:
		return false
	default:
		return left > right
	}
}

func toneFromCount(running int, total int) string {
	if total == 0 {
		return "warning"
	}
	if running == total {
		return "positive"
	}
	if running == 0 {
		return "critical"
	}
	return "warning"
}

func toneFromReminderCount(count int) string {
	switch {
	case count == 0:
		return "positive"
	case count >= 3:
		return "critical"
	default:
		return "warning"
	}
}

func ternaryStatus(condition bool, yes string, no string) string {
	if condition {
		return yes
	}
	return no
}

func deriveReadyStatus(prerequisite bool, completed bool) string {
	if completed {
		return "completed"
	}
	if prerequisite {
		return "ready"
	}
	return "pending"
}

func ternaryText(condition bool, yes string, no string) string {
	if condition {
		return yes
	}
	return no
}

func firstPath(condition bool, yes string, no string) string {
	if condition && strings.TrimSpace(yes) != "" {
		return yes
	}
	return no
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
