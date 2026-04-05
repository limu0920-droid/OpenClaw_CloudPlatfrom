import type {
  AdminOverviewData,
  AdminInstanceDetail,
  Alert,
  AuthConfig,
  AuthSession,
  ApiAdminInstanceDetail,
  ApiAdminInstanceListItem,
  ApiAdminOverview,
  ApiAlert,
  ApiAuditEvent,
  ApiBackup,
  ApiChannel,
  ApiChannelActivity,
  ApiChannelDetail,
  ApiChannelListItem,
  ApiCluster,
  ApiInstance,
  ApiJob,
  ApiPortalInstanceDetail,
  ApiPortalInstanceListItem,
  ApiPortalOverview,
  ApiTenant,
  AuditLog,
  Backup,
  Channel,
  ChannelConnectPayload,
  ChannelProvider,
  CreateInstancePayload,
  Instance,
  InstanceOperationsData,
  Job,
  Order,
  PlanOffer,
  PortalInstanceDetail,
  PortalOverviewData,
  PurchasePayload,
  SearchConfig,
  SearchLogItem,
  Tenant,
  Ticket,
  TicketCreatePayload,
  TriggerBackupPayload,
  UpdateConfigPayload,
} from './types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

function formatDateTime(value?: string) {
  if (!value) {
    return '—'
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatBytes(bytes: number) {
  if (!bytes) {
    return '—'
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let index = 0

  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index += 1
  }

  return `${value.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

async function request<T>(path: string, init?: RequestInit) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  if (!response.ok) {
    throw new Error(`请求失败：${response.status}`)
  }

  return (await response.json()) as T
}

function getInstanceSpec(spec: Record<string, string>) {
  const cpu = spec.cpu ?? '-'
  const memory = spec.memory ?? '-'
  return `${cpu} vCPU / ${memory}`
}

function toInstance(
  item: ApiInstance,
  access: Instance['access'],
  tenant?: ApiTenant | null,
  cluster?: ApiCluster | null,
): Instance {
  return {
    id: String(item.id),
    code: item.code,
    name: item.name,
    status: item.status,
    version: item.version,
    plan: item.plan,
    region: item.region,
    updatedAt: formatDateTime(item.updatedAt),
    expiresAt: formatDateTime(item.expiredAt),
    access,
    runtimeType: item.runtimeType,
    tenantName: tenant?.name,
    clusterName: cluster?.name,
    spec: getInstanceSpec(item.spec),
  }
}

function toJob(job: ApiJob, instanceName?: string): Job {
  return {
    id: String(job.id),
    type: job.type,
    status: job.status,
    target: instanceName ?? `实例 #${job.targetId}`,
    startedAt: formatDateTime(job.startedAt),
    finishedAt: formatDateTime(job.finishedAt),
    progress:
      job.status === 'running'
        ? 66
        : job.status === 'verifying'
          ? 92
          : job.status === 'succeeded'
            ? 100
            : undefined,
  }
}

function toAlert(alert: ApiAlert, instanceName?: string): Alert {
  return {
    id: String(alert.id),
    title: alert.summary,
    severity: alert.severity,
    target: instanceName ?? `实例 #${alert.instanceId}`,
    time: formatDateTime(alert.triggeredAt),
  }
}

function toBackup(backup: ApiBackup): Backup {
  return {
    id: String(backup.id),
    name: backup.backupNo,
    status: backup.status,
    size: formatBytes(backup.sizeBytes),
    createdAt: formatDateTime(backup.finishedAt || backup.startedAt),
  }
}

function toTenant(tenant: ApiTenant): Tenant {
  return {
    id: String(tenant.id),
    code: tenant.code,
    name: tenant.name,
    plan: tenant.plan,
    status: tenant.status,
    expiresAt: formatDateTime(tenant.expiredAt),
    createdAt: formatDateTime(tenant.createdAt),
    updatedAt: formatDateTime(tenant.updatedAt),
  }
}

function toOrder(item: {
  id: number
  planCode: string
  action: string
  status: string
  amount: number
  createdAt: string
}): Order {
  return {
    id: String(item.id),
    planCode: item.planCode,
    action: item.action,
    status: item.status,
    amount: item.amount,
    createdAt: formatDateTime(item.createdAt),
  }
}

function toPlanOffer(item: {
  id: number
  code: string
  name: string
  monthlyPrice: number
  cpu: string
  memory: string
  storage: string
  highlight: string
  features: string[]
}): PlanOffer {
  return {
    id: String(item.id),
    code: item.code,
    name: item.name,
    monthlyPrice: item.monthlyPrice,
    cpu: item.cpu,
    memory: item.memory,
    storage: item.storage,
    highlight: item.highlight,
    features: item.features,
  }
}

function toTicket(item: {
  id: number
  ticketNo: string
  title: string
  category: string
  severity: string
  status: string
  reporter: string
  assignee?: string
  description: string
  createdAt: string
  updatedAt: string
  instanceId?: number
}): Ticket {
  return {
    id: String(item.id),
    ticketNo: item.ticketNo,
    title: item.title,
    category: item.category,
    severity: item.severity,
    status: item.status,
    reporter: item.reporter,
    assignee: item.assignee,
    description: item.description,
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
    instanceId: item.instanceId ? String(item.instanceId) : undefined,
  }
}

function mapChannelProvider(platform: string): ChannelProvider {
  switch (platform) {
    case 'wecom':
      return 'wechat_work'
    default:
      return platform as ChannelProvider
  }
}

function mapChannelStatus(status: ApiChannel['status']): Channel['status'] {
  switch (status) {
    case 'connecting':
      return 'pending'
    default:
      return status
  }
}

function mapChannelAuthMode(method: ApiChannel['connectMethod']): Channel['authMode'] {
  switch (method) {
    case 'qrcode':
      return 'qr'
    default:
      return method as Channel['authMode']
  }
}

function toChannelActivity(activity: ApiChannelActivity) {
  return {
    id: String(activity.id),
    type: activity.type,
    description: `${activity.title} · ${activity.summary}`,
    time: formatDateTime(activity.createdAt),
  }
}

function toChannel(channel: ApiChannel, activities: ApiChannelActivity[]): Channel {
  return {
    id: String(channel.id),
    code: channel.code,
    name: channel.name,
    provider: mapChannelProvider(channel.platform),
    status: mapChannelStatus(channel.status),
    authMode: mapChannelAuthMode(channel.connectMethod),
    connectedAt: formatDateTime(channel.createdAt),
    lastActiveAt: formatDateTime(channel.updatedAt),
    health:
      channel.health.status === 'healthy'
        ? 'good'
        : channel.health.status === 'warning'
          ? 'warning'
          : 'critical',
    webhookUrl: channel.webhookUrl,
    callbackSecret: channel.callbackSecret,
    recentActivity: activities.map(toChannelActivity),
    entrypoints: (channel.entryPoints ?? []).map((item) => ({
      label: item.label,
      url: item.url,
    })),
    notes: channel.notes ?? channel.lastError,
    messages24h: channel.stats.messages24h,
    usersActive24: channel.stats.usersActive24,
    successRate: channel.stats.successRate,
  }
}

export const api = {
  async getAuthConfig(): Promise<AuthConfig> {
    return request<AuthConfig>('/api/v1/auth/config')
  },

  async getAuthSession(): Promise<AuthSession> {
    return request<AuthSession>('/api/v1/auth/session')
  },

  async getKeycloakLoginURL(redirectURI?: string) {
    const suffix = redirectURI ? `?redirect_uri=${encodeURIComponent(redirectURI)}` : ''
    return request<{ url: string }>(`/api/v1/auth/keycloak/url${suffix}`)
  },

  async getSearchConfig(): Promise<SearchConfig> {
    return request<SearchConfig>('/api/v1/search/config')
  },

  async searchLogs(query: { q?: string; kind?: string; instanceId?: string }) {
    const params = new URLSearchParams()
    if (query.q) params.set('q', query.q)
    if (query.kind) params.set('kind', query.kind)
    if (query.instanceId) params.set('instanceId', query.instanceId)
    const suffix = params.toString() ? `?${params.toString()}` : ''
    return request<{ backend: string; items: SearchLogItem[] }>(`/api/v1/search/logs${suffix}`)
  },

  async getPortalOverview(): Promise<PortalOverviewData> {
    const overview = await request<ApiPortalOverview>('/api/v1/portal/overview')

    return {
      metrics: [
        {
          label: '运行实例',
          value: String(overview.instanceRunning),
          delta: `总实例 ${overview.instanceTotal}`,
          tone: 'positive',
        },
        {
          label: '异常实例',
          value: String(overview.instanceAbnormal),
          delta: `${overview.recentAlerts.length} 条近端告警`,
          tone: overview.instanceAbnormal > 0 ? 'warning' : 'neutral',
        },
        {
          label: '近端备份',
          value: String(overview.recentBackups.length),
          delta: '按实例最近窗口统计',
          tone: 'neutral',
        },
        {
          label: '活跃任务',
          value: String(
            overview.recentJobs.filter((job) => job.status === 'running' || job.status === 'verifying')
              .length,
          ),
          delta: `${overview.recentJobs.length} 条任务轨迹`,
          tone: 'critical',
        },
      ],
      quickLinks: [
        {
          label: '进入实例列表',
          url: '/portal/instances',
        },
        {
          label: '查看任务中心',
          url: '/portal/jobs',
        },
        {
          label: '追踪操作日志',
          url: '/portal/logs',
        },
        {
          label: '主实例详情',
          url: overview.primaryInstanceId ? `/portal/instances/${overview.primaryInstanceId}` : '/portal/instances',
        },
      ],
      jobs: overview.recentJobs.map((job) => toJob(job)),
      alerts: overview.recentAlerts.map((alert) => toAlert(alert)),
      headline: '一站式管理你的实例、配置与备份',
      description:
        '当前界面已接入 Go Mock API，后续替换真实控制面接口时可复用相同的数据形态与路由结构。',
      primaryInstanceId: overview.primaryInstanceId ? String(overview.primaryInstanceId) : undefined,
    }
  },

  async getPortalInstances(): Promise<Instance[]> {
    const response = await request<{ items: ApiPortalInstanceListItem[] }>('/api/v1/portal/instances')
    return response.items.map((item) => toInstance(item.instance, item.access))
  },

  async getPortalInstanceDetail(id: string): Promise<PortalInstanceDetail> {
    const response = await request<ApiPortalInstanceDetail>(`/api/v1/portal/instances/${id}`)
    const instanceName = response.instance.name

    return {
      instance: toInstance(response.instance, response.access),
      backups: response.backups.map(toBackup),
      jobs: response.jobs.map((job) => toJob(job, instanceName)),
      alerts: response.alerts.map((alert) => toAlert(alert, instanceName)),
      config: response.config
        ? {
            version: response.config.version,
            hash: response.config.hash,
            publishedAt: formatDateTime(response.config.publishedAt),
            updatedBy: response.config.updatedBy,
            settings: {
              model: response.config.settings.model,
              allowedOrigins: response.config.settings.allowedOrigins,
              backupPolicy: response.config.settings.backupPolicy,
            },
          }
        : null,
    }
  },

  async getPortalJobs(): Promise<Job[]> {
    const response = await request<{ items: ApiJob[] }>('/api/v1/portal/jobs')
    return response.items.map((job) => toJob(job))
  },

  async getPortalLogs(): Promise<AuditLog[]> {
    const response = await request<{ items: ApiAuditEvent[] }>('/api/v1/portal/logs')
    return response.items.map((item) => ({
      id: String(item.id),
      actor: item.actor,
      action: item.action,
      target: `${item.target} #${item.targetId}`,
      result: item.result,
      time: formatDateTime(item.createdAt),
      note: item.metadata ? Object.entries(item.metadata).map(([key, value]) => `${key}: ${value}`).join(' · ') : undefined,
    }))
  },

  async getAdminOverview(): Promise<AdminOverviewData> {
    const overview = await request<ApiAdminOverview>('/api/v1/admin/overview')

    return {
      metrics: [
        {
          label: '租户',
          value: String(overview.tenantTotal),
          delta: '首批租户样例',
          tone: 'neutral',
        },
        {
          label: '实例',
          value: String(overview.instanceTotal),
          delta: '跨集群统一视图',
          tone: 'positive',
        },
        {
          label: '集群',
          value: String(overview.clusterTotal),
          delta: '运行时资源池',
          tone: 'neutral',
        },
        {
          label: '打开告警',
          value: String(overview.openAlerts),
          delta: `${overview.recentAlerts.length} 条最近告警`,
          tone: overview.openAlerts > 0 ? 'critical' : 'neutral',
        },
      ],
      tasks: overview.recentJobs.map((job) => toJob(job)),
      alerts: overview.recentAlerts.map((alert) => toAlert(alert)),
    }
  },

  async getAdminTenants(): Promise<Tenant[]> {
    const response = await request<{ items: ApiTenant[] }>('/api/v1/admin/tenants')
    return response.items.map(toTenant)
  },

  async getAdminInstances(): Promise<Instance[]> {
    const response = await request<{ items: ApiAdminInstanceListItem[] }>('/api/v1/admin/instances')
    return response.items.map((item) => toInstance(item.instance, item.access, item.tenant, item.cluster))
  },

  async getAdminInstanceDetail(id: string): Promise<AdminInstanceDetail> {
    const response = await request<ApiAdminInstanceDetail>(`/api/v1/admin/instances/${id}`)
    const instanceName = response.instance.name

    return {
      instance: toInstance(response.instance, response.access, response.tenant, response.cluster),
      tenant: response.tenant ? toTenant(response.tenant) : null,
      cluster: response.cluster,
      backups: response.backups.map(toBackup),
      jobs: response.jobs.map((job) => toJob(job, instanceName)),
      alerts: response.alerts.map((alert) => toAlert(alert, instanceName)),
      audits: response.audits.map((item) => ({
        id: String(item.id),
        actor: item.actor,
        action: item.action,
        target: `${item.target} #${item.targetId}`,
        result: item.result,
        time: formatDateTime(item.createdAt),
        note: item.metadata
          ? Object.entries(item.metadata)
              .map(([key, value]) => `${key}: ${value}`)
              .join(' · ')
          : undefined,
      })),
      runtimeLogs: response.runtimeLogs.map((item) => ({
        id: item.id,
        timestamp: formatDateTime(item.timestamp),
        level: item.level,
        source: item.source,
        message: item.message,
      })),
      resourceTrend: response.resourceTrend,
      config: response.config
        ? {
            version: response.config.version,
            hash: response.config.hash,
            publishedAt: formatDateTime(response.config.publishedAt),
            updatedBy: response.config.updatedBy,
            settings: {
              model: response.config.settings.model,
              allowedOrigins: response.config.settings.allowedOrigins,
              backupPolicy: response.config.settings.backupPolicy,
            },
          }
        : null,
    }
  },

  async getAdminJobs(): Promise<Job[]> {
    const response = await request<{ items: ApiJob[] }>('/api/v1/admin/jobs')
    return response.items.map((job) => toJob(job))
  },

  async getAdminAlerts(): Promise<Alert[]> {
    const response = await request<{ items: ApiAlert[] }>('/api/v1/admin/alerts')
    return response.items.map((alert) => toAlert(alert))
  },

  async getAdminAudit(): Promise<AuditLog[]> {
    const response = await request<{ items: ApiAuditEvent[] }>('/api/v1/admin/audit')
    return response.items.map((item) => ({
      id: String(item.id),
      actor: item.actor,
      action: item.action,
      target: `${item.target} #${item.targetId}`,
      result: item.result,
      time: formatDateTime(item.createdAt),
      note: item.metadata ? Object.entries(item.metadata).map(([key, value]) => `${key}: ${value}`).join(' · ') : undefined,
    }))
  },

  async createPortalInstance(payload: CreateInstancePayload) {
    return request('/api/v1/portal/instances', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async updateInstanceConfig(id: string, payload: UpdateConfigPayload) {
    return request(`/api/v1/portal/instances/${id}/config`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
  },

  async triggerBackup(id: string, payload: TriggerBackupPayload) {
    return request(`/api/v1/portal/instances/${id}/backups`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async getPortalInstanceOperations(id: string): Promise<InstanceOperationsData> {
    const response = await request<{
      runtime: {
        powerState: 'running' | 'stopped' | 'restarting'
        cpuUsagePercent: number
        memoryUsagePercent: number
        diskUsagePercent: number
        apiRequests24h: number
        apiTokens24h: number
        lastSeenAt: string
      } | null
      credentials: {
        adminUser: string
        passwordMasked: string
        lastRotatedAt: string
        requiresReset: boolean
      } | null
      orders: Array<{
        id: number
        planCode: string
        action: string
        status: string
        amount: number
        createdAt: string
      }>
    }>(`/api/v1/portal/instances/${id}/runtime`)

    return {
      runtime: response.runtime
        ? {
            ...response.runtime,
            lastSeenAt: formatDateTime(response.runtime.lastSeenAt),
          }
        : null,
      credentials: response.credentials
        ? {
            ...response.credentials,
            lastRotatedAt: formatDateTime(response.credentials.lastRotatedAt),
          }
        : null,
      orders: response.orders.map(toOrder),
    }
  },

  async powerPortalInstance(id: string, action: 'start' | 'stop' | 'restart') {
    return request(`/api/v1/portal/instances/${id}/${action}`, {
      method: 'POST',
    })
  },

  async getPortalPlans(): Promise<PlanOffer[]> {
    const response = await request<{ items: Array<{
      id: number
      code: string
      name: string
      monthlyPrice: number
      cpu: string
      memory: string
      storage: string
      highlight: string
      features: string[]
    }> }>('/api/v1/portal/plans')
    return response.items.map(toPlanOffer)
  },

  async createPurchase(payload: PurchasePayload) {
    return request('/api/v1/portal/purchases', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async getPortalTickets(): Promise<Ticket[]> {
    const response = await request<{ items: Array<{
      id: number
      ticketNo: string
      title: string
      category: string
      severity: string
      status: string
      reporter: string
      assignee?: string
      description: string
      createdAt: string
      updatedAt: string
      instanceId?: number
    }> }>('/api/v1/portal/tickets')
    return response.items.map(toTicket)
  },

  async createPortalTicket(payload: TicketCreatePayload) {
    return request('/api/v1/portal/tickets', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async getAdminTickets(): Promise<Ticket[]> {
    const response = await request<{ items: Array<{
      id: number
      ticketNo: string
      title: string
      category: string
      severity: string
      status: string
      reporter: string
      assignee?: string
      description: string
      createdAt: string
      updatedAt: string
      instanceId?: number
    }> }>('/api/v1/admin/tickets')
    return response.items.map(toTicket)
  },

  async updateAdminTicketStatus(id: string, status: string, assignee = '') {
    return request(`/api/v1/admin/tickets/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, assignee }),
    })
  },

  async getChannels(_scope: 'portal' | 'admin' = 'portal'): Promise<Channel[]> {
    const response = await request<{ items: ApiChannelListItem[] }>('/api/v1/channels')
    return response.items.map((item) => toChannel(item.channel, item.activities))
  },

  async getChannelDetail(id: string, _scope: 'portal' | 'admin' = 'portal'): Promise<Channel> {
    const response = await request<ApiChannelDetail>(`/api/v1/channels/${id}`)
    return toChannel(response.channel, response.activities)
  },

  async connectChannel(payload: ChannelConnectPayload) {
    const channels = await this.getChannels('portal')
    const target = channels.find((item) => item.provider === payload.provider)
    if (!target) {
      throw new Error('当前原型尚未为该渠道准备连接器，请先在后端 seed 中启用此 provider')
    }

    return request(`/api/v1/channels/${target.id}/connect`, {
      method: 'POST',
      body: JSON.stringify({
        method: payload.authMode,
        token: payload.token,
      }),
    })
  },

  async disconnectChannel(id: string) {
    return request(`/api/v1/channels/${id}/disconnect`, { method: 'POST' })
  },

  async checkChannelHealth(id: string) {
    return request(`/api/v1/channels/${id}/health`, { method: 'POST' })
  },

  async getChannelActivities(id: string) {
    const response = await request<{ items: ApiChannelActivity[] }>(`/api/v1/channels/${id}/activities`)
    return response.items.map(toChannelActivity)
  },
}
