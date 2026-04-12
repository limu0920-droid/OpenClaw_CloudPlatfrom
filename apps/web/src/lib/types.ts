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
  instanceId?: string
  detailPath?: string
  workspacePath?: string
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

export interface SelfServiceStep {
  key: string
  title: string
  description: string
  status: 'completed' | 'ready' | 'pending'
  actionLabel: string
  actionPath: string
  result?: string
}

export interface UsageQuota {
  key: string
  label: string
  used: number
  limit: number
  unit: string
  percent: number
  status: 'healthy' | 'warning' | 'critical'
  detail: string
  usedText: string
  limitText: string
}

export interface SelfServiceReminder {
  key: string
  severity: 'info' | 'warning' | 'critical'
  title: string
  description: string
  actionLabel: string
  actionPath: string
  at?: string
}

export interface SelfServiceSession {
  id: string
  title: string
  instanceId: string
  instanceName: string
  sessionNo: string
  status: string
  updatedAt: string
  messageCount: number
  artifactCount: number
  workspacePath: string
}

export interface PortalArtifactCenterItem {
  id: string
  title: string
  kind: string
  sourceUrl: string
  previewUrl?: string
  archiveStatus?: string
  contentType?: string
  sizeBytes?: number
  sizeLabel?: string
  filename?: string
  createdAt: string
  updatedAt: string
  instanceId: string
  instanceName: string
  instanceStatus: string
  sessionId: string
  sessionNo?: string
  sessionTitle: string
  sessionStatus: string
  messageId?: string
  messagePreview?: string
  tenantId?: string
  tenantName?: string
  viewCount?: number
  downloadCount?: number
  workspacePath: string
  detailPath?: string
  lineageKey?: string
  version: number
  latestVersion: number
  parentArtifactId?: string
  isFavorite: boolean
  favoriteCount: number
  shareCount: number
  thumbnail: ArtifactThumbnail
  quality: ArtifactQualitySummary
  preview: WorkspaceArtifactPreview
}

export interface ArtifactThumbnail {
  mode: string
  url?: string
  label: string
  hint?: string
}

export interface ArtifactQualitySummary {
  status: string
  score: number
  inlinePreview: boolean
  previewMode: string
  strategy: string
  failureReason?: string
  lastViewedAt?: string
  lastDownloadedAt?: string
  viewCount: number
  downloadCount: number
}

export interface ArtifactFailureBucket {
  reason: string
  count: number
}

export interface ArtifactCenterStats {
  totalCount: number
  favoriteCount: number
  sharedCount: number
  versionedCount: number
  inlinePreviewCount: number
  fallbackCount: number
  recentViewedCount: number
  failureReasons: ArtifactFailureBucket[]
}

export interface ArtifactShare {
  id: string
  token: string
  shareUrl: string
  scope: string
  note?: string
  createdBy: string
  createdAt: string
  expiresAt?: string
  active: boolean
  useCount: number
  lastOpenedAt?: string
}

export interface ArtifactCenterResponse {
  items: PortalArtifactCenterItem[]
  recentViewed: PortalArtifactCenterItem[]
  stats: ArtifactCenterStats
}

export interface OEMBrand {
  id: string
  code: string
  name: string
  status: string
  logoUrl?: string
  faviconUrl?: string
  supportEmail?: string
  supportUrl?: string
  domains: string[]
}

export interface OEMTheme {
  brandId: string
  primaryColor?: string
  secondaryColor?: string
  accentColor?: string
  surfaceMode?: string
  fontFamily?: string
  radius?: string
}

export interface OEMFeatureFlags {
  brandId: string
  portalEnabled: boolean
  adminEnabled: boolean
  channelsEnabled: boolean
  ticketsEnabled: boolean
  purchaseEnabled: boolean
  runtimeControlEnabled: boolean
  auditEnabled: boolean
  ssoEnabled: boolean
}

export interface TenantBrandBinding {
  tenantId: string
  brandId: string
  bindingMode: string
  updatedAt?: string
}

export interface OEMConfig {
  brand: OEMBrand | null
  theme: OEMTheme | null
  features: OEMFeatureFlags | null
  binding: TenantBrandBinding | null
}

export interface OEMBrandRecord {
  brand: OEMBrand
  theme: OEMTheme | null
  features: OEMFeatureFlags | null
  bindings: TenantBrandBinding[]
}

export interface UpdateOEMBrandPayload {
  name: string
  status: string
  logoUrl?: string
  faviconUrl?: string
  supportEmail?: string
  supportUrl?: string
  domains: string[]
}

export interface UpdateOEMThemePayload {
  primaryColor?: string
  secondaryColor?: string
  accentColor?: string
  surfaceMode?: string
  fontFamily?: string
  radius?: string
}

export interface UpdateOEMFeatureFlagsPayload {
  portalEnabled: boolean
  adminEnabled: boolean
  channelsEnabled: boolean
  ticketsEnabled: boolean
  purchaseEnabled: boolean
  runtimeControlEnabled: boolean
  auditEnabled: boolean
  ssoEnabled: boolean
}

export interface ReplaceTenantBrandBindingsPayload {
  bindings: Array<{
    tenantId: string
    bindingMode: string
  }>
}

export interface AccountSettings {
  tenantId: string
  primaryEmail: string
  billingEmail: string
  alertEmail: string
  preferredLocale: string
  secondaryLocale: string
  timezone: string
  emailVerified: boolean
  marketingOptIn: boolean
  notifyOnAlert: boolean
  notifyOnPayment: boolean
  notifyOnExpiry: boolean
  notifyChannelEmail: boolean
  notifyChannelWebhook: boolean
  notifyChannelInApp: boolean
  notificationWebhookUrl?: string
  portalHeadline?: string
  portalSubtitle?: string
  workspaceCallout?: string
  experimentBadge?: string
  updatedAt: string
}

export interface UserProfile {
  id: string
  tenantId: string
  loginName: string
  status: string
  lockReason?: string
  displayName: string
  email: string
  phone: string
  avatarUrl?: string
  locale: string
  timezone: string
  department?: string
  title?: string
  bio?: string
  passwordMasked: string
  updatedAt: string
}

export interface AuthIdentity {
  id: string
  userId: string
  tenantId: string
  provider: string
  isPrimary: boolean
  status: string
  statusReason?: string
  subject: string
  email?: string
  openId?: string
  unionId?: string
  externalName?: string
  lastLoginAt?: string
  updatedAt: string
}

export interface AccountPortrait {
  profile: UserProfile
  accountSettings: AccountSettings | null
  wallet: WalletBalance | null
  billingStatements: BillingStatement[]
  identities: AuthIdentity[]
  recentOrders: Order[]
  recentTickets: Ticket[]
  summary: {
    identityCount: number
    activeIdentityCount: number
    recentOrderCount: number
    recentTicketCount: number
  }
}

export interface PortalOpsMonthlyUsage {
  label: string
  chargeAmount: number
  paidAmount: number
  sessions: number
  messages: number
  artifacts: number
}

export interface PortalNotificationChannel {
  key: string
  label: string
  enabled: boolean
  target?: string
  description: string
}

export interface PortalNotificationTemplatePreview {
  key: string
  title: string
  subject: string
  body: string
}

export interface PortalOpsReport {
  tenant: {
    id: string
    name: string
    plan: string
    status: string
  }
  brand: {
    name?: string
    supportEmail?: string
    supportUrl?: string
  }
  settings: AccountSettings | null
  summary: {
    instanceCount: number
    sessionCount: number
    messageCount: number
    artifactCount: number
    orderCount: number
    openTicketCount: number
    statementCount: number
    walletBalance: number
    currency: string
    billingMonthCount: number
  }
  notificationChannels: PortalNotificationChannel[]
  notificationTemplates: PortalNotificationTemplatePreview[]
  monthlyUsage: PortalOpsMonthlyUsage[]
  export: {
    csvPath: string
  }
}

export interface PortalSelfServiceSummary {
  tenant: {
    id: string
    code: string
    name: string
    plan: string
    status: string
    expiredAt?: string
    supportEmail?: string
    supportUrl?: string
  }
  launchpad: {
    primaryInstanceId?: string
    primaryInstanceName?: string
    workspacePath: string
    artifactsPath: string
    workspaceUrl?: string
  }
  experience: {
    portalHeadline: string
    portalSubtitle: string
    workspaceCallout: string
    experimentBadge: string
  }
  onboarding: {
    showGuide: boolean
    completedCount: number
    totalCount: number
    isReadyForUsage: boolean
    steps: SelfServiceStep[]
  }
  metrics: MetricCard[]
  quotas: UsageQuota[]
  reminders: SelfServiceReminder[]
  recentSessions: SelfServiceSession[]
  recentArtifacts: PortalArtifactCenterItem[]
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
  sessionId?: string
  messageId?: string
  traceId?: string
  instancePath?: string
  workspacePath?: string
}

export interface ApiUserProfile {
  id: number
  tenantId: number
  loginName: string
  status: string
  lockReason?: string
  displayName: string
  email: string
  phone: string
  avatarUrl?: string
  locale: string
  timezone: string
  department?: string
  title?: string
  bio?: string
  passwordMasked: string
  updatedAt: string
}

export interface ApiAuthIdentity {
  id: number
  userId: number
  tenantId: number
  provider: string
  isPrimary: boolean
  status: string
  statusReason?: string
  subject: string
  email?: string
  openId?: string
  unionId?: string
  externalName?: string
  lastLoginAt?: string
  updatedAt: string
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

export interface ApiOrder {
  id: number
  tenantId: number
  instanceId?: number
  planCode: string
  action: string
  status: string
  amount: number
  currency: string
  orderNo: string
  createdAt: string
  updatedAt: string
}

export interface ApiSubscription {
  id: number
  subscriptionNo: string
  tenantId: number
  instanceId?: number
  productCode: string
  planCode: string
  status: string
  renewMode: string
  currentPeriodStart: string
  currentPeriodEnd: string
  expiredAt?: string
  createdAt: string
  updatedAt: string
}

export interface ApiPaymentTransaction {
  id: number
  orderId: number
  channel: string
  payMode: string
  tradeNo: string
  channelOrderNo?: string
  amount: number
  currency: string
  status: string
  payUrl?: string
  codeUrl?: string
  prepayId?: string
  appId?: string
  mchId?: string
  paidAt?: string
  createdAt: string
  updatedAt: string
  raw?: Record<string, unknown>
}

export interface ApiRefundRecord {
  id: number
  orderId: number
  paymentId: number
  refundNo: string
  channelRefundNo?: string
  status: string
  amount: number
  reason: string
  notifyUrl?: string
  createdAt: string
  updatedAt: string
}

export interface ApiInvoiceRecord {
  id: number
  tenantId: number
  orderId: number
  invoiceType: string
  status: string
  amount: number
  title: string
  taxNo?: string
  email?: string
  invoiceNo?: string
  pdfUrl?: string
  createdAt: string
  updatedAt: string
}

export interface ApiPaymentCallbackEvent {
  id: number
  channel: string
  eventType: string
  outTradeNo?: string
  outRefundNo?: string
  signatureStatus: string
  decryptStatus: string
  processStatus: string
  requestSerial?: string
  createdAt: string
  rawBody: string
}

export interface ApiWalletBalance {
  tenantId: number
  currency: string
  availableAmount: number
  frozenAmount: number
  creditLimit: number
  autoRecharge: boolean
  lastSettlementAt?: string
  updatedAt: string
}

export interface ApiBillingStatement {
  id: number
  tenantId: number
  statementNo: string
  billingMonth: string
  status: string
  currency: string
  openingBalance: number
  chargeAmount: number
  refundAmount: number
  closingBalance: number
  paidAmount: number
  dueAt: string
  createdAt: string
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

export interface RuntimeBinding {
  clusterId: string
  namespace: string
  workloadId: string
  workloadName: string
}

export interface RuntimeWorkload {
  id: string
  clusterId: string
  namespace: string
  name: string
  kind: string
  image: string
  status: string
  desired: number
  ready: number
  available: number
  lastActionAt: string
}

export interface RuntimeWorkloadMetrics {
  workloadId: string
  cpuUsageMilli: number
  memoryUsageMB: number
  networkRxKB: number
  networkTxKB: number
  errorRatePercent: number
  requestsPerMinute: number
}

export interface DiagnosticPod {
  id: string
  name: string
  nodeName: string
  status: string
  restarts: number
  startedAt?: string
  workloadId?: string
  image?: string
}

export interface DiagnosticSignal {
  type: string
  severity: string
  summary: string
  triggeredAt?: string
  podName?: string
  restarts?: number
}

export interface DiagnosticCommandCatalogItem {
  key: string
  label: string
  description: string
  commandText: string
  manualAllowed: boolean
}

export interface DiagnosticPolicy {
  defaultAccessMode: 'readonly' | 'whitelist'
  readonlyTtlMinutes: number
  whitelistTtlMinutes: number
  requiresApprovalForWhitelist: boolean
  maxActiveSessionsPerInstance: number
  commandCatalog: DiagnosticCommandCatalogItem[]
}

export interface DiagnosticSessionSummary {
  id: string
  sessionNo: string
  tenantId: string
  instanceId: string
  clusterId?: string
  namespace?: string
  workloadId?: string
  workloadName?: string
  podName: string
  containerName?: string
  accessMode: 'readonly' | 'whitelist'
  status: string
  approvalTicket?: string
  approvedBy?: string
  operator: string
  operatorUserId?: string
  reason?: string
  closeReason?: string
  expiresAt?: string
  lastCommandAt?: string
  startedAt?: string
  endedAt?: string
  createdAt: string
  updatedAt: string
  commandCount: number
  lastCommandStatus?: string
  lastCommandText?: string
}

export interface DiagnosticCommandRecord {
  id: string
  sessionId: string
  tenantId: string
  instanceId: string
  commandKey?: string
  commandText: string
  status: string
  exitCode: number
  durationMs: number
  output?: string
  errorOutput?: string
  outputTruncated?: boolean
  executedAt: string
}

export interface AdminInstanceDiagnosticsSummary {
  binding: RuntimeBinding | null
  workload: RuntimeWorkload | null
  metrics: RuntimeWorkloadMetrics | null
  pods: DiagnosticPod[]
  signals: DiagnosticSignal[]
  policy: DiagnosticPolicy
  sessions: DiagnosticSessionSummary[]
}

export interface AdminDiagnosticSessionDetail {
  session: DiagnosticSessionSummary
  commands: DiagnosticCommandRecord[]
  record: string
  commandCatalog: DiagnosticCommandCatalogItem[]
}

export type ApprovalStatus =
  | 'pending'
  | 'approved'
  | 'rejected'
  | 'executing'
  | 'executed'
  | 'cancelled'
  | 'expired'

export type ApprovalRiskLevel = 'medium' | 'high' | 'critical'

export interface ApprovalActionRecord {
  id: string
  approvalId: string
  actorName: string
  action: string
  comment?: string
  createdAt: string
  metadata?: Record<string, string>
}

export interface ApprovalSummary {
  id: string
  approvalNo: string
  tenantId: string
  instanceId?: string
  approvalType: string
  targetType: string
  targetId?: string
  applicantId: string
  applicantName?: string
  approverId?: string
  approverName?: string
  executorId?: string
  executorName?: string
  status: ApprovalStatus
  riskLevel: ApprovalRiskLevel
  reason?: string
  approvalComment?: string
  rejectReason?: string
  approvedAt?: string
  executedAt?: string
  expiredAt?: string
  createdAt: string
  updatedAt: string
  metadata?: Record<string, string>
}

export interface ApprovalDetail {
  approval: ApprovalSummary
  actions: ApprovalActionRecord[]
  instance?: Instance
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
  workspaceSessions: WorkspaceSession[]
  bridgeSummary: {
    traceCount: number
    eventCount: number
    failedTraceCount: number
    artifactCount: number
    lastEventAt?: string
    recentTraces: Array<{
      traceId: string
      sessionId?: string
      sessionNo?: string
      latestAt: string
      status: string
      preview?: string
      messageCount: number
      eventCount: number
      artifactCount: number
      toolCount: number
    }>
  }
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
  runtimeLogs: Array<{
    id: string
    timestamp: string
    level: 'info' | 'warning' | 'error'
    source: string
    message: string
    sessionId?: number
    messageId?: number
    traceId?: string
    instancePath?: string
    workspacePath?: string
  }>
  resourceTrend: ResourceTrendPoint[]
  workspaceSessions: Array<{
    id: number
    sessionNo: string
    tenantId: number
    tenantName?: string
    instanceId: number
    instanceName?: string
    title: string
    status: string
    workspaceUrl?: string
    protocolVersion?: string
    lastOpenedAt?: string
    lastArtifactAt?: string
    lastSyncedAt?: string
    lastMessageAt?: string
    lastMessagePreview?: string
    lastMessageRole?: string
    lastMessageStatus?: string
    messageCount?: number
    artifactCount?: number
    hasArtifacts?: boolean
    createdAt: string
    updatedAt: string
  }>
  bridgeSummary: {
    traceCount: number
    eventCount: number
    failedTraceCount: number
    artifactCount: number
    lastEventAt?: string
    recentTraces: Array<{
      traceId: string
      sessionId?: number
      sessionNo?: string
      latestAt: string
      status: string
      preview?: string
      messageCount: number
      eventCount: number
      artifactCount: number
      toolCount: number
    }>
  }
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
  tenantId: string
  instanceId?: string
  planCode: string
  action: string
  status: string
  amount: number
  currency: string
  orderNo: string
  createdAt: string
  updatedAt: string
}

export interface Subscription {
  id: string
  subscriptionNo: string
  tenantId: string
  instanceId?: string
  productCode: string
  planCode: string
  status: string
  renewMode: string
  currentPeriodStart: string
  currentPeriodEnd: string
  expiredAt?: string
  createdAt: string
  updatedAt: string
}

export interface PaymentTransaction {
  id: string
  orderId: string
  channel: string
  payMode: string
  tradeNo: string
  channelOrderNo?: string
  amount: number
  currency: string
  status: string
  payUrl?: string
  codeUrl?: string
  prepayId?: string
  appId?: string
  mchId?: string
  paidAt?: string
  createdAt: string
  updatedAt: string
  raw?: Record<string, unknown>
}

export interface RefundRecord {
  id: string
  orderId: string
  paymentId: string
  refundNo: string
  channelRefundNo?: string
  status: string
  amount: number
  reason: string
  notifyUrl?: string
  createdAt: string
  updatedAt: string
}

export interface InvoiceRecord {
  id: string
  tenantId: string
  orderId: string
  invoiceType: string
  status: string
  amount: number
  title: string
  taxNo?: string
  email?: string
  invoiceNo?: string
  pdfUrl?: string
  createdAt: string
  updatedAt: string
}

export interface PaymentCallbackEvent {
  id: string
  channel: string
  eventType: string
  outTradeNo?: string
  outRefundNo?: string
  signatureStatus: string
  decryptStatus: string
  processStatus: string
  requestSerial?: string
  createdAt: string
  rawBody: string
}

export interface WalletBalance {
  tenantId: string
  currency: string
  availableAmount: number
  frozenAmount: number
  creditLimit: number
  autoRecharge: boolean
  lastSettlementAt?: string
  updatedAt: string
}

export interface BillingStatement {
  id: string
  tenantId: string
  statementNo: string
  billingMonth: string
  status: string
  currency: string
  openingBalance: number
  chargeAmount: number
  refundAmount: number
  closingBalance: number
  paidAmount: number
  dueAt: string
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

export type WorkspaceScope = 'portal' | 'admin'

export interface WorkspaceSession {
  id: string
  sessionNo: string
  tenantId: string
  instanceId: string
  tenantName?: string
  instanceName?: string
  title: string
  status: string
  workspaceUrl?: string
  protocolVersion?: string
  lastOpenedAt?: string
  lastArtifactAt?: string
  lastSyncedAt?: string
  lastMessageAt?: string
  lastMessagePreview?: string
  lastMessageRole?: string
  lastMessageStatus?: string
  messageCount?: number
  artifactCount?: number
  hasArtifacts?: boolean
  createdAt: string
  updatedAt: string
}

export interface WorkspaceArtifact {
  id: string
  sessionId: string
  tenantId: string
  instanceId: string
  messageId?: string
  title: string
  kind: string
  origin?: string
  sourceUrl: string
  previewUrl?: string
  archiveStatus?: string
  contentType?: string
  sizeBytes?: number
  filename?: string
  createdAt: string
  updatedAt: string
}

export interface WorkspaceArtifactPreview {
  available: boolean
  mode: string
  strategy: string
  sandboxed: boolean
  proxied: boolean
  previewUrl?: string
  downloadUrl?: string
  externalUrl?: string
  failureReason?: string
  note?: string
}

export interface WorkspaceMessage {
  id: string
  sessionId: string
  tenantId: string
  instanceId: string
  parentMessageId?: string
  role: string
  status: string
  origin?: string
  traceId?: string
  errorCode?: string
  errorMessage?: string
  deliveryAttempt?: number
  content: string
  deliveredAt?: string
  createdAt: string
  updatedAt: string
}

export interface WorkspaceMessageDispatch {
  ok: boolean
  target?: string
  status?: number
  code?: string
  error?: string
  message?: string
  traceId?: string
  attempt?: number
  streamUrl?: string
}

export interface WorkspaceMessageDispatchResponse {
  message: WorkspaceMessage
  dispatch: WorkspaceMessageDispatch
  reply?: WorkspaceMessage
  artifacts: WorkspaceArtifact[]
}

export interface WorkspaceSessionEventPayload {
  message?: WorkspaceMessage
  artifact?: WorkspaceArtifact
  dispatch?: WorkspaceMessageDispatch
  delta?: string
  content?: string
  toolCallId?: string
  toolName?: string
  status?: string
  detail?: string
  reasoning?: string
  error?: string
  errorMessage?: string
}

export interface WorkspaceSessionEvent {
  id: string
  sessionId: string
  messageId?: string
  tenantId: string
  instanceId: string
  eventType: string
  origin?: string
  traceId?: string
  payload: WorkspaceSessionEventPayload
  createdAt: string
}

export type ArtifactCenterItem = PortalArtifactCenterItem

export interface ArtifactAccessLog {
  id: string
  artifactId: string
  sessionId: string
  tenantId: string
  instanceId: string
  action: string
  scope: string
  actor: string
  remoteAddr?: string
  userAgent?: string
  createdAt: string
}

export interface ArtifactCenterDetail {
  artifact: ArtifactCenterItem
  preview: WorkspaceArtifactPreview
  session?: WorkspaceSession
  message?: WorkspaceMessage | null
  accessLogs: ArtifactAccessLog[]
  versions: PortalArtifactCenterItem[]
  shares: ArtifactShare[]
}

export type WorkspaceStreamEvent =
  | { type: 'message'; message: WorkspaceMessage }
  | { type: 'status'; dispatch: WorkspaceMessageDispatch }
  | { type: 'chunk'; delta: string }
  | { type: 'reply'; message: WorkspaceMessage }
  | { type: 'artifact'; artifact: WorkspaceArtifact }
  | { type: 'error'; error: string }
  | { type: 'done'; messageId?: string }

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
  channelId?: string
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
  provider: string
  enabled: boolean
  baseUrl?: string
  realm?: string
  clientId?: string
  defaultRedirect?: string
  mockUserEnabled?: boolean
  providers?: Array<{
    provider: string
    enabled: boolean
    mode?: string
  }>
}

export interface AuthSession {
  authenticated: boolean
  provider: string
  user?: {
    name: string
    email: string
    role: string
    openId?: string
    unionId?: string
  }
}

export interface SearchConfig {
  provider: string
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
  sessionId?: string
  messageId?: string
  traceId?: string
  source: string
  createdAt: string
  instancePath?: string
  workspacePath?: string
}
