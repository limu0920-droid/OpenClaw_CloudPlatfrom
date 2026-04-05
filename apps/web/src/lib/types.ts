export type InstanceStatus =
  | 'pending_payment'
  | 'provisioning'
  | 'running'
  | 'updating'
  | 'backing_up'
  | 'restoring'
  | 'failed'
  | 'stopped'
  | 'expired'
  | 'deleting'
  | 'deleted'

export type JobStatus =
  | 'accepted'
  | 'queued'
  | 'running'
  | 'verifying'
  | 'succeeded'
  | 'failed'
  | 'cancelled'

export type BackupStatus = 'pending' | 'running' | 'succeeded' | 'failed'

export type AlertSeverity = 'info' | 'warning' | 'critical'

export interface InstanceAccess {
  entryType: string
  url: string
  isPrimary?: boolean
  accessMode?: string
  domain?: string
}

export interface Instance {
  id: string
  code: string
  name: string
  status: InstanceStatus
  version: string
  plan: string
  region: string
  updatedAt: string
  expiresAt?: string
  access: InstanceAccess[]
  runtimeType?: string
  tenantName?: string
  clusterName?: string
  spec?: string
}

export interface Job {
  id: string
  type: string
  status: JobStatus
  target: string
  startedAt: string
  finishedAt?: string
  progress?: number
}

export interface Backup {
  id: string
  name: string
  status: BackupStatus
  size: string
  createdAt: string
}

export interface Alert {
  id: string
  title: string
  severity: AlertSeverity
  target: string
  time: string
}

export interface Tenant {
  id: string
  code: string
  name: string
  plan: string
  status: string
  expiresAt?: string
  createdAt?: string
  updatedAt?: string
}

export interface MetricCard {
  label: string
  value: string
  delta?: string
  tone?: 'neutral' | 'positive' | 'warning' | 'critical'
}

export interface PortalOverviewData {
  metrics: MetricCard[]
  quickLinks: Array<{ label: string; url: string }>
  jobs: Job[]
  alerts: Alert[]
  headline: string
  description: string
  primaryInstanceId?: string
}

export interface PortalInstanceDetail {
  instance: Instance
  backups: Backup[]
  jobs: Job[]
  alerts: Alert[]
  config: {
    version: number
    hash: string
    publishedAt: string
    updatedBy: string
    settings: {
      model: string
      allowedOrigins: string
      backupPolicy: string
    }
  } | null
}

export interface AdminOverviewData {
  metrics: MetricCard[]
  tasks: Job[]
  alerts: Alert[]
}

export interface AuditLog {
  id: string
  actor: string
  action: string
  target: string
  result: string
  time: string
  note?: string
}

export interface RuntimeLog {
  id: string
  timestamp: string
  level: 'info' | 'warning' | 'error'
  source: string
  message: string
}

export interface ApiTenant {
  id: number
  code: string
  name: string
  status: string
  plan: string
  expiredAt?: string
  createdAt: string
  updatedAt: string
}

export interface ApiCluster {
  id: number
  code: string
  name: string
  region: string
  status: string
  nodeCount: number
  createdAt: string
  updatedAt: string
}

export interface ApiInstance {
  id: number
  tenantId: number
  clusterId: number
  code: string
  name: string
  status: InstanceStatus
  version: string
  plan: string
  runtimeType: string
  region: string
  spec: Record<string, string>
  activatedAt?: string
  expiredAt?: string
  createdAt: string
  updatedAt: string
}

export interface ApiInstanceConfig {
  instanceId: number
  version: number
  hash: string
  publishedAt: string
  updatedBy: string
  settings: {
    model: string
    allowedOrigins: string
    backupPolicy: string
  }
}

export interface ApiBackup {
  id: number
  instanceId: number
  backupNo: string
  type: 'manual' | 'scheduled'
  status: BackupStatus
  sizeBytes: number
  startedAt: string
  finishedAt?: string
}

export interface ApiJob {
  id: number
  jobNo: string
  type: string
  targetType: string
  targetId: number
  status: JobStatus
  summary: string
  startedAt: string
  finishedAt?: string
}

export interface ApiAlert {
  id: number
  instanceId: number
  severity: AlertSeverity
  status: string
  metricKey: string
  summary: string
  triggeredAt: string
}

export interface ApiAuditEvent {
  id: number
  tenantId: number
  actor: string
  action: string
  target: string
  targetId: number
  result: string
  createdAt: string
  metadata?: Record<string, string>
}

export interface ApiPortalOverview {
  tenantId: number
  instanceTotal: number
  instanceRunning: number
  instanceAbnormal: number
  recentJobs: ApiJob[]
  recentAlerts: ApiAlert[]
  recentBackups: ApiBackup[]
  primaryInstanceId?: number | null
}

export interface ApiPortalInstanceListItem {
  instance: ApiInstance
  access: InstanceAccess[]
  config: ApiInstanceConfig | null
  backups: ApiBackup[]
}

export interface ApiPortalInstanceDetail {
  instance: ApiInstance
  access: InstanceAccess[]
  config: ApiInstanceConfig | null
  backups: ApiBackup[]
  jobs: ApiJob[]
  alerts: ApiAlert[]
}

export interface ResourceTrendPoint {
  label: string
  cpu: number
  memory: number
  requests: number
}

export interface AdminInstanceDetail {
  instance: Instance
  tenant: Tenant | null
  cluster: ApiCluster | null
  backups: Backup[]
  jobs: Job[]
  alerts: Alert[]
  audits: AuditLog[]
  runtimeLogs: RuntimeLog[]
  resourceTrend: ResourceTrendPoint[]
  config: {
    version: number
    hash: string
    publishedAt: string
    updatedBy: string
    settings: {
      model: string
      allowedOrigins: string
      backupPolicy: string
    }
  } | null
}

export interface ApiAdminInstanceDetail {
  instance: ApiInstance
  tenant: ApiTenant | null
  cluster: ApiCluster | null
  access: InstanceAccess[]
  config: ApiInstanceConfig | null
  backups: ApiBackup[]
  jobs: ApiJob[]
  alerts: ApiAlert[]
  audits: ApiAuditEvent[]
  runtimeLogs: RuntimeLog[]
  resourceTrend: ResourceTrendPoint[]
}

export interface CreateInstancePayload {
  name: string
  plan: string
  region: string
  cpu: string
  memory: string
}

export interface UpdateConfigPayload {
  updatedBy: string
  settings: {
    model: string
    allowedOrigins: string
    backupPolicy: string
  }
}

export interface TriggerBackupPayload {
  type: string
  operator: string
}

export interface InstanceRuntime {
  powerState: 'running' | 'stopped' | 'restarting'
  cpuUsagePercent: number
  memoryUsagePercent: number
  diskUsagePercent: number
  apiRequests24h: number
  apiTokens24h: number
  lastSeenAt: string
}

export interface InstanceCredential {
  adminUser: string
  passwordMasked: string
  lastRotatedAt: string
  requiresReset: boolean
}

export interface InstanceOperationsData {
  runtime: InstanceRuntime | null
  credentials: InstanceCredential | null
  orders: Order[]
}

export interface PlanOffer {
  id: string
  code: string
  name: string
  monthlyPrice: number
  cpu: string
  memory: string
  storage: string
  highlight: string
  features: string[]
}

export interface Order {
  id: string
  planCode: string
  action: string
  status: string
  amount: number
  createdAt: string
}

export interface Ticket {
  id: string
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
  instanceId?: string
}

export interface TicketCreatePayload {
  instanceId?: number
  title: string
  category: string
  severity: string
  description: string
  reporter: string
}

export interface PurchasePayload {
  planCode: string
  instanceId?: number
  action: string
}

export interface ApiAdminOverview {
  tenantTotal: number
  instanceTotal: number
  clusterTotal: number
  openAlerts: number
  recentJobs: ApiJob[]
  recentAlerts: ApiAlert[]
}

export interface ApiAdminInstanceListItem {
  instance: ApiInstance
  tenant: ApiTenant | null
  cluster: ApiCluster | null
  access: InstanceAccess[]
  config: ApiInstanceConfig | null
  alerts: ApiAlert[]
}

export type ChannelProvider =
  | 'feishu'
  | 'wechat_work'
  | 'dingtalk'
  | 'slack'
  | 'telegram'
  | 'discord'
  | 'whatsapp'
  | 'custom'

export type ChannelAuthMode = 'oauth' | 'qr' | 'token' | 'webhook' | 'custom'

export type ChannelStatus = 'connected' | 'pending' | 'disconnected' | 'degraded' | 'error'

export interface Channel {
  id: string
  code?: string
  name: string
  provider: ChannelProvider
  status: ChannelStatus
  authMode: ChannelAuthMode
  connectedAt?: string
  lastActiveAt?: string
  health?: 'good' | 'warning' | 'critical'
  webhookUrl?: string
  callbackSecret?: string
  recentActivity: Array<{
    id: string
    type: string
    description: string
    time: string
  }>
  entrypoints: Array<{
    label: string
    url: string
  }>
  notes?: string
  messages24h?: number
  usersActive24?: number
  successRate?: number
}

export interface ChannelConnectPayload {
  provider: ChannelProvider
  authMode: ChannelAuthMode
  redirectUri?: string
  token?: string
}

export interface ApiChannelHealth {
  lastChecked: string
  status: 'healthy' | 'warning' | 'critical'
  latencyMs: number
}

export interface ApiChannelStats {
  messages24h: number
  usersActive24: number
  successRate: number
}

export interface ApiChannel {
  id: number
  code: string
  name: string
  platform: string
  status: 'connected' | 'connecting' | 'disconnected' | 'degraded' | 'error'
  connectMethod: 'oauth' | 'webhook' | 'token' | 'qrcode'
  authUrl?: string
  webhookUrl?: string
  qrCodeUrl?: string
  tokenMasked?: string
  callbackSecret?: string
  health: ApiChannelHealth
  stats: ApiChannelStats
  lastError?: string
  settings?: Record<string, unknown>
  entryPoints?: Array<{
    label: string
    url: string
  }>
  notes?: string
  updatedAt: string
  createdAt: string
}

export interface ApiChannelActivity {
  id: number
  channelId: number
  type: string
  title: string
  summary: string
  createdAt: string
}

export interface ApiChannelListItem {
  channel: ApiChannel
  activities: ApiChannelActivity[]
}

export interface ApiChannelDetail {
  channel: ApiChannel
  activities: ApiChannelActivity[]
}

export interface AuthConfig {
  provider: 'keycloak'
  enabled: boolean
  baseUrl?: string
  realm?: string
  clientId?: string
  defaultRedirect?: string
  mockUserEnabled: boolean
}

export interface AuthSession {
  authenticated: boolean
  provider: 'keycloak' | 'mock'
  user?: {
    name: string
    email: string
    role: string
  }
}

export interface SearchConfig {
  provider: 'opensearch'
  enabled: boolean
  url?: string
  index?: string
}

export interface SearchLogItem {
  id: string
  kind: string
  level: string
  title: string
  message: string
  instanceId?: string
  source: string
  createdAt: string
}
