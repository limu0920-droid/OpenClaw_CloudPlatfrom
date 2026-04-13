import type {
  AccountPortrait,
  ApprovalActionRecord,
  ApprovalDetail,
  ApprovalSummary,
  AdminOverviewData,
  AdminInstanceDetail,
  AdminInstanceDiagnosticsSummary,
  AdminDiagnosticSessionDetail,
  AccountSettings,
  Alert,
  ArtifactCenterDetail,
  ArtifactCenterResponse,
  ArtifactShare,
  AuthIdentity,
  AuthConfig,
  AuthSession,
  BillingStatement,
  DiagnosticCommandCatalogItem,
  DiagnosticCommandRecord,
  DiagnosticPod,
  DiagnosticPolicy,
  DiagnosticSessionSummary,
  DiagnosticSignal,
  ApiAdminInstanceDetail,
  ApiAdminInstanceListItem,
  ApiAdminOverview,
  ApiAlert,
  ApiAuthIdentity,
  ApiAuditEvent,
  ApiBackup,
  ApiBillingStatement,
  ApiChannel,
  ApiChannelActivity,
  ApiChannelDetail,
  ApiChannelListItem,
  ApiCluster,
  ApiInvoiceRecord,
  ApiInstance,
  ApiJob,
  ApiOrder,
  ApiPaymentCallbackEvent,
  ApiPaymentTransaction,
  ApiPortalInstanceDetail,
  ApiPortalInstanceListItem,
  ApiPortalOverview,
  ApiRefundRecord,
  ApiSubscription,
  ApiTenant,
  ApiUserProfile,
  ApiWalletBalance,
  AuditLog,
  Backup,
  Channel,
  ChannelConnectPayload,
  ChannelProvider,
  CreateInstancePayload,
  InvoiceRecord,
  Instance,
  InstanceOperationsData,
  Job,
  OEMBrand,
  OEMBrandRecord,
  OEMConfig,
  OEMFeatureFlags,
  OEMTheme,
  Order,
  PaymentCallbackEvent,
  PaymentTransaction,
  PlanOffer,
  PortalNotificationChannel,
  PortalNotificationTemplatePreview,
  PortalInstanceDetail,
  PortalOpsMonthlyUsage,
  PortalOpsReport,
  PortalOverviewData,
  PortalArtifactCenterItem,
  PortalSelfServiceSummary,
  PurchasePayload,
  RuntimeBinding,
  RuntimeWorkload,
  RuntimeWorkloadMetrics,
  RefundRecord,
  SearchConfig,
  SearchLogItem,
  Subscription,
  Tenant,
  ReplaceTenantBrandBindingsPayload,
  Ticket,
  TicketCreatePayload,
  TriggerBackupPayload,
  UpdateOEMBrandPayload,
  UpdateOEMFeatureFlagsPayload,
  UpdateOEMThemePayload,
  UpdateConfigPayload,
  UserProfile,
  WalletBalance,
  WorkspaceArtifact,
  WorkspaceArtifactPreview,
  WorkspaceMessage,
  WorkspaceMessageDispatchResponse,
  WorkspaceSessionEvent,
  WorkspaceStreamEvent,
  WorkspaceSession,
} from './types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''
const API_PATH_PREFIX = normalizeApiPathPrefix(import.meta.env.VITE_API_PATH_PREFIX ?? '/api/v1')
const API_REQUEST_CREDENTIALS = normalizeRequestCredentials(
  import.meta.env.VITE_API_REQUEST_CREDENTIALS ?? 'same-origin',
)
const API_TOKEN_STORAGE_KEY = (import.meta.env.VITE_API_TOKEN_STORAGE_KEY ?? '').trim()

function normalizeApiPathPrefix(value: string) {
  const normalized = value.trim()

  if (!normalized || normalized === '/') {
    return ''
  }

  return normalized.startsWith('/') ? normalized.replace(/\/$/, '') : `/${normalized.replace(/\/$/, '')}`
}

function normalizeRequestCredentials(value: string): RequestCredentials {
  switch (value) {
    case 'include':
    case 'omit':
      return value
    default:
      return 'same-origin'
  }
}

function resolveApiPath(path: string) {
  if (/^https?:\/\//.test(path)) {
    return path
  }

  if (!API_PATH_PREFIX) {
    return path
  }

  if (path === '/api/v1') {
    return API_PATH_PREFIX
  }

  if (path.startsWith('/api/v1/')) {
    return `${API_PATH_PREFIX}${path.slice('/api/v1'.length)}`
  }

  return path
}

function resolveApiURL(path: string) {
  const resolvedPath = resolveApiPath(path)

  if (/^https?:\/\//.test(resolvedPath)) {
    return resolvedPath
  }

  return `${API_BASE_URL}${resolvedPath}`
}

function getStoredAccessToken() {
  if (!API_TOKEN_STORAGE_KEY || typeof window === 'undefined') {
    return ''
  }

  try {
    return window.localStorage.getItem(API_TOKEN_STORAGE_KEY) ?? window.sessionStorage.getItem(API_TOKEN_STORAGE_KEY) ?? ''
  } catch {
    return ''
  }
}

function createRequestHeaders(init?: RequestInit) {
  const headers = new Headers(init?.headers)

  if (!headers.has('Accept')) {
    headers.set('Accept', 'application/json')
  }

  const body = init?.body
  if (body && !(body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const token = getStoredAccessToken()
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  return headers
}

async function parseResponseBody(response: Response) {
  if (response.status === 204) {
    return undefined
  }

  const raw = await response.text()
  if (!raw) {
    return undefined
  }

  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json') || raw.startsWith('{') || raw.startsWith('[')) {
    try {
      return JSON.parse(raw) as unknown
    } catch {
      return raw
    }
  }

  return raw
}

function extractErrorMessage(payload: unknown, fallback: string) {
  if (typeof payload === 'string') {
    const normalized = payload.trim()
    return normalized || fallback
  }

  if (!payload || typeof payload !== 'object') {
    return fallback
  }

  const record = payload as Record<string, unknown>
  const directMessage = [record.message, record.error, record.detail].find(
    (value): value is string => typeof value === 'string' && value.trim().length > 0,
  )

  if (directMessage) {
    return directMessage
  }

  if (record.error && typeof record.error === 'object') {
    const nestedError = record.error as Record<string, unknown>
    if (typeof nestedError.message === 'string' && nestedError.message.trim()) {
      return nestedError.message
    }
  }

  if (Array.isArray(record.errors)) {
    const firstError = record.errors[0]
    if (typeof firstError === 'string' && firstError.trim()) {
      return firstError
    }
    if (firstError && typeof firstError === 'object') {
      const nestedError = firstError as Record<string, unknown>
      if (typeof nestedError.message === 'string' && nestedError.message.trim()) {
        return nestedError.message
      }
    }
  }

  return fallback
}

function unwrapResponsePayload<T>(payload: unknown) {
  if (!payload || typeof payload !== 'object') {
    return payload as T
  }

  const record = payload as Record<string, unknown>

  for (const key of ['data', 'result', 'payload']) {
    if (key in record) {
      return record[key] as T
    }
  }

  return payload as T
}

function appendQuery(path: string, params: Record<string, string | undefined>) {
  const query = new URLSearchParams()

  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      query.set(key, value)
    }
  })

  const suffix = query.toString()
  return suffix ? `${path}?${suffix}` : path
}

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

function formatOptionalDateTime(value?: string) {
  if (!value) {
    return undefined
  }

  const formatted = formatDateTime(value)
  return formatted === '—' ? undefined : formatted
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

function resolveMaybeApiURL(value?: string) {
  if (!value) {
    return undefined
  }
  return value.startsWith('/api/') ? resolveApiURL(value) : value
}

function normalizePreviewMode(mode?: string): WorkspaceArtifactPreview['mode'] {
  switch (mode) {
    case 'image':
    case 'video':
    case 'audio':
    case 'text':
    case 'download':
    case 'pdf':
      return mode
    case 'web':
    case 'sandbox':
    case 'iframe':
      return 'html'
    default:
      return 'download'
  }
}

async function request<T>(path: string, init?: RequestInit) {
  const response = await fetch(resolveApiURL(path), {
    credentials: API_REQUEST_CREDENTIALS,
    headers: createRequestHeaders(init),
    ...init,
  })

  const payload = await parseResponseBody(response)

  if (!response.ok) {
    throw new Error(extractErrorMessage(payload, `请求失败：${response.status}`))
  }

  return unwrapResponsePayload<T>(payload)
}

function parseSSEPayload(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return undefined
  }
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try {
      return JSON.parse(trimmed) as unknown
    } catch {
      return trimmed
    }
  }
  return trimmed
}

interface EventStreamFrame {
  event: string
  payload: unknown
  id?: string
}

function parseEventStreamFrames(rawBlock: string) {
  const block = rawBlock.replace(/\r/g, '')
  if (!block.trim()) {
    return null
  }

  let event = 'message'
  let id = ''
  const dataLines: string[] = []
  for (const line of block.split('\n')) {
    if (line.startsWith(':')) {
      continue
    }
    if (line.startsWith('event:')) {
      event = line.slice(6).trim() || 'message'
      continue
    }
    if (line.startsWith('id:')) {
      id = line.slice(3).trim()
      continue
    }
    if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trimStart())
    }
  }

  return {
    event,
    id: id || undefined,
    payload: parseSSEPayload(dataLines.join('\n')),
  } satisfies EventStreamFrame
}

async function readEventStream(
  response: Response,
  onEvent: (event: string, payload: unknown) => void | Promise<void>,
) {
  await readEventStreamFrames(response, async ({ event, payload }) => {
    await onEvent(event, payload)
  })
}

async function readEventStreamFrames(
  response: Response,
  onFrame: (frame: EventStreamFrame) => void | Promise<void>,
) {
  if (!response.body) {
    return
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  const flushBuffer = async (force = false) => {
    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const rawBlock = buffer.slice(0, boundary)
      buffer = buffer.slice(boundary + 2)
      const frame = parseEventStreamFrames(rawBlock)
      if (frame) {
        await onFrame(frame)
      }
      boundary = buffer.indexOf('\n\n')
    }

    if (force) {
      const frame = parseEventStreamFrames(buffer)
      if (frame) {
        await onFrame(frame)
      }
      buffer = ''
    }
  }

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      break
    }
    buffer += decoder.decode(value, { stream: true }).replace(/\r\n/g, '\n')
    await flushBuffer()
  }
  buffer += decoder.decode().replace(/\r\n/g, '\n')
  await flushBuffer(true)
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
    instanceId: String(alert.instanceId),
    detailPath: `/admin/instances/${alert.instanceId}?tab=monitoring`,
    workspacePath: `/admin/instances/${alert.instanceId}/workspace`,
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

function toUserProfile(item: ApiUserProfile): UserProfile {
  return {
    id: String(item.id),
    tenantId: String(item.tenantId),
    loginName: item.loginName,
    status: item.status,
    lockReason: item.lockReason,
    displayName: item.displayName,
    email: item.email,
    phone: item.phone,
    avatarUrl: item.avatarUrl,
    locale: item.locale,
    timezone: item.timezone,
    department: item.department,
    title: item.title,
    bio: item.bio,
    passwordMasked: item.passwordMasked,
    updatedAt: formatDateTime(item.updatedAt),
  }
}

function toAuthIdentity(item: ApiAuthIdentity): AuthIdentity {
  return {
    id: String(item.id),
    userId: String(item.userId),
    tenantId: String(item.tenantId),
    provider: item.provider,
    isPrimary: Boolean(item.isPrimary),
    status: item.status,
    statusReason: item.statusReason,
    subject: item.subject,
    email: item.email,
    openId: item.openId,
    unionId: item.unionId,
    externalName: item.externalName,
    lastLoginAt: formatOptionalDateTime(item.lastLoginAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

function toOrder(item: ApiOrder): Order {
  return {
    id: String(item.id),
    tenantId: String(item.tenantId),
    instanceId: item.instanceId ? String(item.instanceId) : undefined,
    planCode: item.planCode,
    action: item.action,
    status: item.status,
    amount: item.amount,
    currency: item.currency,
    orderNo: item.orderNo,
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

// @ts-ignore: unused function
function toSubscription(item: ApiSubscription): Subscription {
  return {
    id: String(item.id),
    subscriptionNo: item.subscriptionNo,
    tenantId: String(item.tenantId),
    instanceId: item.instanceId ? String(item.instanceId) : undefined,
    productCode: item.productCode,
    planCode: item.planCode,
    status: item.status,
    renewMode: item.renewMode,
    currentPeriodStart: formatDateTime(item.currentPeriodStart),
    currentPeriodEnd: formatDateTime(item.currentPeriodEnd),
    expiredAt: formatOptionalDateTime(item.expiredAt),
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

// @ts-ignore: unused function
function toPaymentTransaction(item: ApiPaymentTransaction): PaymentTransaction {
  return {
    id: String(item.id),
    orderId: String(item.orderId),
    channel: item.channel,
    payMode: item.payMode,
    tradeNo: item.tradeNo,
    channelOrderNo: item.channelOrderNo,
    amount: item.amount,
    currency: item.currency,
    status: item.status,
    payUrl: item.payUrl,
    codeUrl: item.codeUrl,
    prepayId: item.prepayId,
    appId: item.appId,
    mchId: item.mchId,
    paidAt: formatOptionalDateTime(item.paidAt),
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
    raw: item.raw,
  }
}

// @ts-ignore: unused function
function toRefundRecord(item: ApiRefundRecord): RefundRecord {
  return {
    id: String(item.id),
    orderId: String(item.orderId),
    paymentId: String(item.paymentId),
    refundNo: item.refundNo,
    channelRefundNo: item.channelRefundNo,
    status: item.status,
    amount: item.amount,
    reason: item.reason,
    notifyUrl: item.notifyUrl,
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

// @ts-ignore: unused function
function toInvoiceRecord(item: ApiInvoiceRecord): InvoiceRecord {
  return {
    id: String(item.id),
    tenantId: String(item.tenantId),
    orderId: String(item.orderId),
    invoiceType: item.invoiceType,
    status: item.status,
    amount: item.amount,
    title: item.title,
    taxNo: item.taxNo,
    email: item.email,
    invoiceNo: item.invoiceNo,
    pdfUrl: item.pdfUrl,
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

// @ts-ignore: unused function
function toPaymentCallbackEvent(item: ApiPaymentCallbackEvent): PaymentCallbackEvent {
  return {
    id: String(item.id),
    channel: item.channel,
    eventType: item.eventType,
    outTradeNo: item.outTradeNo,
    outRefundNo: item.outRefundNo,
    signatureStatus: item.signatureStatus,
    decryptStatus: item.decryptStatus,
    processStatus: item.processStatus,
    requestSerial: item.requestSerial,
    createdAt: formatDateTime(item.createdAt),
    rawBody: item.rawBody,
  }
}

function toWalletBalance(item: ApiWalletBalance): WalletBalance {
  return {
    tenantId: String(item.tenantId),
    currency: item.currency,
    availableAmount: item.availableAmount,
    frozenAmount: item.frozenAmount,
    creditLimit: item.creditLimit,
    autoRecharge: Boolean(item.autoRecharge),
    lastSettlementAt: formatOptionalDateTime(item.lastSettlementAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

function toBillingStatement(item: ApiBillingStatement): BillingStatement {
  return {
    id: String(item.id),
    tenantId: String(item.tenantId),
    statementNo: item.statementNo,
    billingMonth: item.billingMonth,
    status: item.status,
    currency: item.currency,
    openingBalance: item.openingBalance,
    chargeAmount: item.chargeAmount,
    refundAmount: item.refundAmount,
    closingBalance: item.closingBalance,
    paidAmount: item.paidAmount,
    dueAt: formatDateTime(item.dueAt),
    createdAt: formatDateTime(item.createdAt),
  }
}

// @ts-ignore: unused function
function toAccountPortrait(item: {
  profile: ApiUserProfile
  accountSettings?: Parameters<typeof toAccountSettings>[0] | null
  wallet?: ApiWalletBalance | null
  billingStatements?: ApiBillingStatement[]
  identities?: ApiAuthIdentity[]
  recentOrders?: ApiOrder[]
  recentTickets?: Array<{
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
  }>
  summary?: {
    identityCount: number
    activeIdentityCount: number
    recentOrderCount: number
    recentTicketCount: number
  }
}): AccountPortrait {
  return {
    profile: toUserProfile(item.profile),
    accountSettings: item.accountSettings ? toAccountSettings(item.accountSettings) : null,
    wallet: item.wallet ? toWalletBalance(item.wallet) : null,
    billingStatements: (item.billingStatements ?? []).map(toBillingStatement),
    identities: (item.identities ?? []).map(toAuthIdentity),
    recentOrders: (item.recentOrders ?? []).map(toOrder),
    recentTickets: (item.recentTickets ?? []).map(toTicket),
    summary: {
      identityCount: item.summary?.identityCount ?? 0,
      activeIdentityCount: item.summary?.activeIdentityCount ?? 0,
      recentOrderCount: item.summary?.recentOrderCount ?? 0,
      recentTicketCount: item.summary?.recentTicketCount ?? 0,
    },
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

function toWorkspaceSession(item: {
  id: number
  sessionNo: string
  tenantId: number
  instanceId: number
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
}): WorkspaceSession {
  return {
    id: String(item.id),
    sessionNo: item.sessionNo,
    tenantId: String(item.tenantId),
    instanceId: String(item.instanceId),
    tenantName: item.tenantName,
    instanceName: item.instanceName,
    title: item.title,
    status: item.status,
    workspaceUrl: item.workspaceUrl,
    protocolVersion: item.protocolVersion,
    lastOpenedAt: formatDateTime(item.lastOpenedAt),
    lastArtifactAt: formatDateTime(item.lastArtifactAt),
    lastSyncedAt: formatDateTime(item.lastSyncedAt),
    lastMessageAt: formatDateTime(item.lastMessageAt),
    lastMessagePreview: item.lastMessagePreview,
    lastMessageRole: item.lastMessageRole,
    lastMessageStatus: item.lastMessageStatus,
    messageCount: item.messageCount ?? 0,
    artifactCount: item.artifactCount ?? 0,
    hasArtifacts: item.hasArtifacts ?? (item.artifactCount ?? 0) > 0,
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

function toWorkspaceArtifact(item: {
  id: number
  sessionId: number
  tenantId: number
  instanceId: number
  messageId?: number
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
}): WorkspaceArtifact {
  return {
    id: String(item.id),
    sessionId: String(item.sessionId),
    tenantId: String(item.tenantId),
    instanceId: String(item.instanceId),
    messageId: item.messageId ? String(item.messageId) : undefined,
    title: item.title,
    kind: item.kind,
    origin: item.origin,
    sourceUrl: item.sourceUrl,
    previewUrl: item.previewUrl,
    archiveStatus: item.archiveStatus,
    contentType: item.contentType,
    sizeBytes: item.sizeBytes,
    filename: item.filename,
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

function toWorkspaceMessage(item: {
  id: number
  sessionId: number
  tenantId: number
  instanceId: number
  parentMessageId?: number
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
}): WorkspaceMessage {
  return {
    id: String(item.id),
    sessionId: String(item.sessionId),
    tenantId: String(item.tenantId),
    instanceId: String(item.instanceId),
    parentMessageId: item.parentMessageId ? String(item.parentMessageId) : undefined,
    role: item.role,
    status: item.status,
    origin: item.origin,
    traceId: item.traceId,
    errorCode: item.errorCode,
    errorMessage: item.errorMessage,
    deliveryAttempt: item.deliveryAttempt,
    content: item.content,
    deliveredAt: formatDateTime(item.deliveredAt),
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
  }
}

function toWorkspaceMessageDispatchResponse(response: {
  message: Parameters<typeof toWorkspaceMessage>[0]
  dispatch?: {
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
  reply?: Parameters<typeof toWorkspaceMessage>[0]
  artifacts?: Array<Parameters<typeof toWorkspaceArtifact>[0]>
}): WorkspaceMessageDispatchResponse {
  return {
    message: toWorkspaceMessage(response.message),
    dispatch: response.dispatch ?? { ok: false },
    reply: response.reply ? toWorkspaceMessage(response.reply) : undefined,
    artifacts: response.artifacts?.map(toWorkspaceArtifact) ?? [],
  }
}

function toWorkspaceSessionEvent(item: {
  id: number
  sessionId: number
  messageId?: number
  tenantId: number
  instanceId: number
  eventType: string
  origin?: string
  traceId?: string
  payload?: Record<string, unknown>
  createdAt: string
}): WorkspaceSessionEvent {
  const payloadRecord = item.payload ?? {}
  const message =
    payloadRecord.message && typeof payloadRecord.message === 'object'
      ? toWorkspaceMessage(payloadRecord.message as Parameters<typeof toWorkspaceMessage>[0])
      : undefined
  const artifact =
    payloadRecord.artifact && typeof payloadRecord.artifact === 'object'
      ? toWorkspaceArtifact(payloadRecord.artifact as Parameters<typeof toWorkspaceArtifact>[0])
      : undefined
  const dispatch =
    payloadRecord.dispatch && typeof payloadRecord.dispatch === 'object'
      ? (() => {
          const record = payloadRecord.dispatch as Record<string, unknown>
          return {
            ok: Boolean(record.ok),
            target: typeof record.target === 'string' ? record.target : undefined,
            status: typeof record.status === 'number' ? record.status : undefined,
            code: typeof record.code === 'string' ? record.code : undefined,
            error: typeof record.error === 'string' ? record.error : undefined,
            traceId: typeof record.traceId === 'string' ? record.traceId : undefined,
            streamUrl: typeof record.streamUrl === 'string' ? record.streamUrl : undefined,
          }
        })()
      : undefined
  return {
    id: String(item.id),
    sessionId: String(item.sessionId),
    messageId: item.messageId ? String(item.messageId) : undefined,
    tenantId: String(item.tenantId),
    instanceId: String(item.instanceId),
    eventType: item.eventType,
    origin: item.origin,
    traceId: item.traceId,
    payload: {
      message,
      artifact,
      dispatch,
      delta: typeof payloadRecord.delta === 'string' ? payloadRecord.delta : undefined,
      content: typeof payloadRecord.content === 'string' ? payloadRecord.content : undefined,
      toolCallId: typeof payloadRecord.toolCallId === 'string' ? payloadRecord.toolCallId : undefined,
      toolName: typeof payloadRecord.toolName === 'string' ? payloadRecord.toolName : undefined,
      status: typeof payloadRecord.status === 'string' ? payloadRecord.status : undefined,
      detail: typeof payloadRecord.detail === 'string' ? payloadRecord.detail : undefined,
      reasoning: typeof payloadRecord.reasoning === 'string' ? payloadRecord.reasoning : undefined,
      error: typeof payloadRecord.error === 'string' ? payloadRecord.error : undefined,
      errorMessage: typeof payloadRecord.errorMessage === 'string' ? payloadRecord.errorMessage : undefined,
    },
    createdAt: formatDateTime(item.createdAt),
  }
}

function toArtifactPreviewDescriptor(item: {
  available?: boolean
  mode?: string
  strategy?: string
  sandboxed?: boolean
  proxied?: boolean
  previewUrl?: string
  downloadUrl?: string
  externalUrl?: string
  failureReason?: string
  note?: string
}): WorkspaceArtifactPreview {
  return {
    available: Boolean(item.available),
    mode: normalizePreviewMode(item.mode),
    strategy: item.strategy ?? 'download',
    sandboxed: Boolean(item.sandboxed),
    proxied: Boolean(item.proxied),
    previewUrl: resolveMaybeApiURL(item.previewUrl),
    downloadUrl: resolveMaybeApiURL(item.downloadUrl),
    externalUrl: item.externalUrl,
    failureReason: item.failureReason,
    note: item.note,
  }
}

function toArtifactAccessLog(item: {
  id: number
  artifactId: number
  sessionId: number
  tenantId: number
  instanceId: number
  action: string
  scope: string
  actor: string
  remoteAddr?: string
  userAgent?: string
  createdAt: string
}) {
  return {
    id: String(item.id),
    artifactId: String(item.artifactId),
    sessionId: String(item.sessionId),
    tenantId: String(item.tenantId),
    instanceId: String(item.instanceId),
    action: item.action,
    scope: item.scope,
    actor: item.actor,
    remoteAddr: item.remoteAddr,
    userAgent: item.userAgent,
    createdAt: formatDateTime(item.createdAt),
  }
}

function toDiagnosticCommandCatalogItem(item: {
  key: string
  label: string
  description: string
  commandText: string
  manualAllowed: boolean
}): DiagnosticCommandCatalogItem {
  return {
    key: item.key,
    label: item.label,
    description: item.description,
    commandText: item.commandText,
    manualAllowed: item.manualAllowed,
  }
}

function toDiagnosticPolicy(item: {
  defaultAccessMode: 'readonly' | 'whitelist'
  readonlyTtlMinutes: number
  whitelistTtlMinutes: number
  requiresApprovalForWhitelist: boolean
  maxActiveSessionsPerInstance: number
  commandCatalog: Array<{
    key: string
    label: string
    description: string
    commandText: string
    manualAllowed: boolean
  }>
}): DiagnosticPolicy {
  return {
    defaultAccessMode: item.defaultAccessMode,
    readonlyTtlMinutes: item.readonlyTtlMinutes,
    whitelistTtlMinutes: item.whitelistTtlMinutes,
    requiresApprovalForWhitelist: item.requiresApprovalForWhitelist,
    maxActiveSessionsPerInstance: item.maxActiveSessionsPerInstance,
    commandCatalog: item.commandCatalog.map(toDiagnosticCommandCatalogItem),
  }
}

function toDiagnosticSessionSummary(item: {
  id: number
  sessionNo: string
  tenantId: number
  instanceId: number
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
  operatorUserId?: number
  reason?: string
  closeReason?: string
  expiresAt?: string
  lastCommandAt?: string
  startedAt?: string
  endedAt?: string
  createdAt: string
  updatedAt: string
  commandCount?: number
  lastCommandStatus?: string
  lastCommandText?: string
}): DiagnosticSessionSummary {
  return {
    id: String(item.id),
    sessionNo: item.sessionNo,
    tenantId: String(item.tenantId),
    instanceId: String(item.instanceId),
    clusterId: item.clusterId,
    namespace: item.namespace,
    workloadId: item.workloadId,
    workloadName: item.workloadName,
    podName: item.podName,
    containerName: item.containerName,
    accessMode: item.accessMode,
    status: item.status,
    approvalTicket: item.approvalTicket,
    approvedBy: item.approvedBy,
    operator: item.operator,
    operatorUserId: item.operatorUserId ? String(item.operatorUserId) : undefined,
    reason: item.reason,
    closeReason: item.closeReason,
    expiresAt: formatOptionalDateTime(item.expiresAt),
    lastCommandAt: formatOptionalDateTime(item.lastCommandAt),
    startedAt: formatOptionalDateTime(item.startedAt),
    endedAt: formatOptionalDateTime(item.endedAt),
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
    commandCount: item.commandCount ?? 0,
    lastCommandStatus: item.lastCommandStatus,
    lastCommandText: item.lastCommandText,
  }
}

function toDiagnosticCommandRecord(item: {
  id: number
  sessionId: number
  tenantId: number
  instanceId: number
  commandKey?: string
  commandText: string
  status: string
  exitCode: number
  durationMs: number
  output?: string
  errorOutput?: string
  outputTruncated?: boolean
  executedAt: string
}): DiagnosticCommandRecord {
  return {
    id: String(item.id),
    sessionId: String(item.sessionId),
    tenantId: String(item.tenantId),
    instanceId: String(item.instanceId),
    commandKey: item.commandKey,
    commandText: item.commandText,
    status: item.status,
    exitCode: item.exitCode,
    durationMs: item.durationMs,
    output: item.output,
    errorOutput: item.errorOutput,
    outputTruncated: item.outputTruncated,
    executedAt: formatDateTime(item.executedAt),
  }
}

function toDiagnosticPod(item: {
  id: string
  name: string
  nodeName: string
  status: string
  restarts: number
  startedAt?: string
  workloadId?: string
  image?: string
}): DiagnosticPod {
  return {
    id: item.id,
    name: item.name,
    nodeName: item.nodeName,
    status: item.status,
    restarts: item.restarts,
    startedAt: formatOptionalDateTime(item.startedAt),
    workloadId: item.workloadId,
    image: item.image,
  }
}

function toDiagnosticSignal(item: {
  type: string
  severity: string
  summary: string
  triggeredAt?: string
  podName?: string
  restarts?: number
}): DiagnosticSignal {
  return {
    type: item.type,
    severity: item.severity,
    summary: item.summary,
    triggeredAt: formatOptionalDateTime(item.triggeredAt),
    podName: item.podName,
    restarts: item.restarts,
  }
}

function toApprovalActionRecord(item: {
  id: number
  approvalId: number
  actorName: string
  action: string
  comment?: string
  createdAt: string
  metadata?: Record<string, string>
}): ApprovalActionRecord {
  return {
    id: String(item.id),
    approvalId: String(item.approvalId),
    actorName: item.actorName,
    action: item.action,
    comment: item.comment,
    createdAt: formatDateTime(item.createdAt),
    metadata: item.metadata,
  }
}

function toApprovalSummary(item: {
  id: number
  approvalNo: string
  tenantId: number
  instanceId?: number
  approvalType: string
  targetType: string
  targetId?: number
  applicantId: number
  applicantName?: string
  approverId?: number
  approverName?: string
  executorId?: number
  executorName?: string
  status: ApprovalSummary['status']
  riskLevel: ApprovalSummary['riskLevel']
  reason?: string
  approvalComment?: string
  rejectReason?: string
  approvedAt?: string
  executedAt?: string
  expiredAt?: string
  createdAt: string
  updatedAt: string
  metadata?: Record<string, string>
}): ApprovalSummary {
  return {
    id: String(item.id),
    approvalNo: item.approvalNo,
    tenantId: String(item.tenantId),
    instanceId: item.instanceId ? String(item.instanceId) : undefined,
    approvalType: item.approvalType,
    targetType: item.targetType,
    targetId: item.targetId ? String(item.targetId) : undefined,
    applicantId: String(item.applicantId),
    applicantName: item.applicantName,
    approverId: item.approverId ? String(item.approverId) : undefined,
    approverName: item.approverName,
    executorId: item.executorId ? String(item.executorId) : undefined,
    executorName: item.executorName,
    status: item.status,
    riskLevel: item.riskLevel,
    reason: item.reason,
    approvalComment: item.approvalComment,
    rejectReason: item.rejectReason,
    approvedAt: formatOptionalDateTime(item.approvedAt),
    executedAt: formatOptionalDateTime(item.executedAt),
    expiredAt: formatOptionalDateTime(item.expiredAt),
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
    metadata: item.metadata,
  }
}

function toApprovalDetail(
  item: {
    approval: Parameters<typeof toApprovalSummary>[0]
    actions: Parameters<typeof toApprovalActionRecord>[0][]
    instance?: ApiInstance
  },
): ApprovalDetail {
  return {
    approval: toApprovalSummary(item.approval),
    actions: item.actions.map(toApprovalActionRecord),
    instance: item.instance ? toInstance(item.instance, []) : undefined,
  }
}

function toRuntimeBinding(item?: {
  clusterId: string
  namespace: string
  workloadId: string
  workloadName: string
} | null): RuntimeBinding | null {
  if (!item) {
    return null
  }
  return {
    clusterId: item.clusterId,
    namespace: item.namespace,
    workloadId: item.workloadId,
    workloadName: item.workloadName,
  }
}

function toRuntimeWorkload(item?: {
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
} | null): RuntimeWorkload | null {
  if (!item) {
    return null
  }
  return {
    id: item.id,
    clusterId: item.clusterId,
    namespace: item.namespace,
    name: item.name,
    kind: item.kind,
    image: item.image,
    status: item.status,
    desired: item.desired,
    ready: item.ready,
    available: item.available,
    lastActionAt: formatDateTime(item.lastActionAt),
  }
}

function toRuntimeWorkloadMetrics(item?: {
  workloadId: string
  cpuUsageMilli: number
  memoryUsageMB: number
  networkRxKB: number
  networkTxKB: number
  errorRatePercent: number
  requestsPerMinute: number
} | null): RuntimeWorkloadMetrics | null {
  if (!item) {
    return null
  }
  return {
    workloadId: item.workloadId,
    cpuUsageMilli: item.cpuUsageMilli,
    memoryUsageMB: item.memoryUsageMB,
    networkRxKB: item.networkRxKB,
    networkTxKB: item.networkTxKB,
    errorRatePercent: item.errorRatePercent,
    requestsPerMinute: item.requestsPerMinute,
  }
}

function toArtifactThumbnail(item: {
  mode?: string
  url?: string
  label?: string
  hint?: string
}) {
  return {
    mode: item.mode ?? 'generated',
    url: resolveMaybeApiURL(item.url),
    label: item.label ?? 'FILE',
    hint: item.hint,
  }
}

function toArtifactQualitySummary(item?: {
  status?: string
  score?: number
  inlinePreview?: boolean
  previewMode?: string
  strategy?: string
  failureReason?: string
  lastViewedAt?: string
  lastDownloadedAt?: string
  viewCount?: number
  downloadCount?: number
}) {
  return {
    status: item?.status ?? 'warning',
    score: item?.score ?? 0,
    inlinePreview: Boolean(item?.inlinePreview),
    previewMode: item?.previewMode ?? 'download',
    strategy: item?.strategy ?? 'download',
    failureReason: item?.failureReason,
    lastViewedAt: formatDateTime(item?.lastViewedAt),
    lastDownloadedAt: formatDateTime(item?.lastDownloadedAt),
    viewCount: item?.viewCount ?? 0,
    downloadCount: item?.downloadCount ?? 0,
  }
}

function toArtifactShare(item: {
  id: number
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
}): ArtifactShare {
  return {
    id: String(item.id),
    token: item.token,
    shareUrl: item.shareUrl,
    scope: item.scope,
    note: item.note,
    createdBy: item.createdBy,
    createdAt: formatDateTime(item.createdAt),
    expiresAt: formatDateTime(item.expiresAt),
    active: Boolean(item.active),
    useCount: item.useCount,
    lastOpenedAt: formatDateTime(item.lastOpenedAt),
  }
}

function toPortalArtifactCenterItem(item: {
  id: number
  title: string
  kind: string
  sourceUrl: string
  previewUrl?: string
  archiveStatus?: string
  contentType?: string
  sizeBytes?: number
  filename?: string
  createdAt: string
  updatedAt: string
  instanceId: number
  instanceName: string
  instanceStatus: string
  sessionId: number
  sessionNo?: string
  sessionTitle: string
  sessionStatus: string
  messageId?: number
  messagePreview?: string
  tenantId?: number
  tenantName?: string
  viewCount?: number
  downloadCount?: number
  workspacePath: string
  detailPath?: string
  lineageKey?: string
  version?: number
  latestVersion?: number
  parentArtifactId?: number
  isFavorite?: boolean
  favoriteCount?: number
  shareCount?: number
  thumbnail?: {
    mode?: string
    url?: string
    label?: string
    hint?: string
  }
  quality?: {
    status?: string
    score?: number
    inlinePreview?: boolean
    previewMode?: string
    strategy?: string
    failureReason?: string
    lastViewedAt?: string
    lastDownloadedAt?: string
    viewCount?: number
    downloadCount?: number
  }
  preview?: {
    available?: boolean
    mode?: string
    strategy?: string
    sandboxed?: boolean
    proxied?: boolean
    previewUrl?: string
    downloadUrl?: string
    externalUrl?: string
    failureReason?: string
    note?: string
  }
}): PortalArtifactCenterItem {
  return {
    id: String(item.id),
    title: item.title,
    kind: item.kind,
    sourceUrl: item.sourceUrl,
    previewUrl: item.previewUrl,
    archiveStatus: item.archiveStatus,
    contentType: item.contentType,
    sizeBytes: item.sizeBytes,
    sizeLabel: item.sizeBytes ? formatBytes(item.sizeBytes) : undefined,
    filename: item.filename,
    createdAt: formatDateTime(item.createdAt),
    updatedAt: formatDateTime(item.updatedAt),
    instanceId: String(item.instanceId),
    instanceName: item.instanceName,
    instanceStatus: item.instanceStatus,
    sessionId: String(item.sessionId),
    sessionNo: item.sessionNo,
    sessionTitle: item.sessionTitle,
    sessionStatus: item.sessionStatus,
    messageId: item.messageId ? String(item.messageId) : undefined,
    messagePreview: item.messagePreview,
    tenantId: item.tenantId ? String(item.tenantId) : undefined,
    tenantName: item.tenantName,
    viewCount: item.viewCount,
    downloadCount: item.downloadCount,
    workspacePath: item.workspacePath,
    detailPath: item.detailPath,
    lineageKey: item.lineageKey,
    version: item.version ?? 1,
    latestVersion: item.latestVersion ?? item.version ?? 1,
    parentArtifactId: item.parentArtifactId ? String(item.parentArtifactId) : undefined,
    isFavorite: Boolean(item.isFavorite),
    favoriteCount: item.favoriteCount ?? 0,
    shareCount: item.shareCount ?? 0,
    thumbnail: toArtifactThumbnail(item.thumbnail ?? {
      mode: 'generated',
      label: item.kind?.toUpperCase() || 'FILE',
      hint: item.filename || item.title,
    }),
    quality: toArtifactQualitySummary(item.quality ?? {
      status: item.preview?.available ? 'healthy' : 'warning',
      score: item.preview?.available ? 96 : 68,
      inlinePreview: item.preview?.available,
      previewMode: item.preview?.mode ?? 'download',
      strategy: item.preview?.strategy ?? 'download',
      failureReason: item.preview?.failureReason,
      viewCount: item.viewCount,
      downloadCount: item.downloadCount,
    }),
    preview: toArtifactPreviewDescriptor(item.preview ?? {
      available: Boolean(item.previewUrl),
      mode: 'download',
      strategy: 'download',
      externalUrl: item.previewUrl || item.sourceUrl,
    }),
  }
}

function toOEMBrand(item: {
  id: number
  code: string
  name: string
  status: string
  logoUrl?: string
  faviconUrl?: string
  supportEmail?: string
  supportUrl?: string
  domains?: string[]
}): OEMBrand {
  return {
    id: String(item.id),
    code: item.code,
    name: item.name,
    status: item.status,
    logoUrl: item.logoUrl,
    faviconUrl: item.faviconUrl,
    supportEmail: item.supportEmail,
    supportUrl: item.supportUrl,
    domains: item.domains ?? [],
  }
}

function toOEMTheme(item?: {
  brandId: number
  primaryColor?: string
  secondaryColor?: string
  accentColor?: string
  surfaceMode?: string
  fontFamily?: string
  radius?: string
} | null): OEMTheme | null {
  if (!item) return null
  return {
    brandId: String(item.brandId),
    primaryColor: item.primaryColor,
    secondaryColor: item.secondaryColor,
    accentColor: item.accentColor,
    surfaceMode: item.surfaceMode,
    fontFamily: item.fontFamily,
    radius: item.radius,
  }
}

function toOEMFeatureFlags(item?: {
  brandId: number
  portalEnabled: boolean
  adminEnabled: boolean
  channelsEnabled: boolean
  ticketsEnabled: boolean
  purchaseEnabled: boolean
  runtimeControlEnabled: boolean
  auditEnabled: boolean
  ssoEnabled: boolean
} | null): OEMFeatureFlags | null {
  if (!item) return null
  return {
    brandId: String(item.brandId),
    portalEnabled: Boolean(item.portalEnabled),
    adminEnabled: Boolean(item.adminEnabled),
    channelsEnabled: Boolean(item.channelsEnabled),
    ticketsEnabled: Boolean(item.ticketsEnabled),
    purchaseEnabled: Boolean(item.purchaseEnabled),
    runtimeControlEnabled: Boolean(item.runtimeControlEnabled),
    auditEnabled: Boolean(item.auditEnabled),
    ssoEnabled: Boolean(item.ssoEnabled),
  }
}

function toTenantBrandBinding(item?: {
  tenantId: number
  brandId: number
  bindingMode: string
  updatedAt?: string
} | null) {
  if (!item) return null
  return {
    tenantId: String(item.tenantId),
    brandId: String(item.brandId),
    bindingMode: item.bindingMode,
    updatedAt: formatOptionalDateTime(item.updatedAt),
  }
}

function toOEMBrandRecord(response: {
  brand: Parameters<typeof toOEMBrand>[0]
  theme?: Parameters<typeof toOEMTheme>[0]
  features?: Parameters<typeof toOEMFeatureFlags>[0]
  bindings?: Array<{
    tenantId: number
    brandId: number
    bindingMode: string
    updatedAt?: string
  }>
}): OEMBrandRecord {
  return {
    brand: toOEMBrand(response.brand),
    theme: toOEMTheme(response.theme),
    features: toOEMFeatureFlags(response.features),
    bindings: (response.bindings ?? [])
      .map((item) => toTenantBrandBinding(item))
      .filter((item): item is NonNullable<typeof item> => Boolean(item)),
  }
}

function toOEMConfig(response: {
  brand?: Parameters<typeof toOEMBrand>[0]
  theme?: Parameters<typeof toOEMTheme>[0]
  features?: Parameters<typeof toOEMFeatureFlags>[0]
  binding?: {
    tenantId: number
    brandId: number
    bindingMode: string
    updatedAt?: string
  } | null
}): OEMConfig {
  return {
    brand: response.brand ? toOEMBrand(response.brand) : null,
    theme: response.theme ? toOEMTheme(response.theme) : null,
    features: response.features ? toOEMFeatureFlags(response.features) : null,
    binding: toTenantBrandBinding(response.binding),
  }
}

function toAccountSettings(item: {
  tenantId: number
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
  notifyChannelEmail?: boolean
  notifyChannelWebhook?: boolean
  notifyChannelInApp?: boolean
  notificationWebhookUrl?: string
  portalHeadline?: string
  portalSubtitle?: string
  workspaceCallout?: string
  experimentBadge?: string
  updatedAt: string
}): AccountSettings {
  return {
    tenantId: String(item.tenantId),
    primaryEmail: item.primaryEmail,
    billingEmail: item.billingEmail,
    alertEmail: item.alertEmail,
    preferredLocale: item.preferredLocale,
    secondaryLocale: item.secondaryLocale,
    timezone: item.timezone,
    emailVerified: Boolean(item.emailVerified),
    marketingOptIn: Boolean(item.marketingOptIn),
    notifyOnAlert: Boolean(item.notifyOnAlert),
    notifyOnPayment: Boolean(item.notifyOnPayment),
    notifyOnExpiry: Boolean(item.notifyOnExpiry),
    notifyChannelEmail: item.notifyChannelEmail ?? true,
    notifyChannelWebhook: item.notifyChannelWebhook ?? false,
    notifyChannelInApp: item.notifyChannelInApp ?? true,
    notificationWebhookUrl: item.notificationWebhookUrl,
    portalHeadline: item.portalHeadline,
    portalSubtitle: item.portalSubtitle,
    workspaceCallout: item.workspaceCallout,
    experimentBadge: item.experimentBadge,
    updatedAt: formatDateTime(item.updatedAt),
  }
}

function toPortalNotificationChannel(item: {
  key: string
  label: string
  enabled: boolean
  target?: string
  description: string
}): PortalNotificationChannel {
  return {
    key: item.key,
    label: item.label,
    enabled: Boolean(item.enabled),
    target: item.target,
    description: item.description,
  }
}

function toPortalNotificationTemplatePreview(item: {
  key: string
  title: string
  subject: string
  body: string
}): PortalNotificationTemplatePreview {
  return {
    key: item.key,
    title: item.title,
    subject: item.subject,
    body: item.body,
  }
}

function toPortalOpsMonthlyUsage(item: {
  label: string
  chargeAmount: number
  paidAmount: number
  sessions: number
  messages: number
  artifacts: number
}): PortalOpsMonthlyUsage {
  return {
    label: item.label,
    chargeAmount: item.chargeAmount,
    paidAmount: item.paidAmount,
    sessions: item.sessions,
    messages: item.messages,
    artifacts: item.artifacts,
  }
}

export const api = {
  async getOEMConfig(domain?: string): Promise<OEMConfig> {
    const path = appendQuery('/api/v1/oem/config', {
      domain: domain?.trim() || undefined,
    })
    const response = await request<{
      brand?: {
        id: number
        code: string
        name: string
        status: string
        logoUrl?: string
        faviconUrl?: string
        supportEmail?: string
        supportUrl?: string
        domains?: string[]
      }
      theme?: {
        brandId: number
        primaryColor?: string
        secondaryColor?: string
        accentColor?: string
        surfaceMode?: string
        fontFamily?: string
        radius?: string
      }
      features?: {
        brandId: number
        portalEnabled: boolean
        adminEnabled: boolean
        channelsEnabled: boolean
        ticketsEnabled: boolean
        purchaseEnabled: boolean
        runtimeControlEnabled: boolean
        auditEnabled: boolean
        ssoEnabled: boolean
      }
      binding?: {
        tenantId: number
        brandId: number
        bindingMode: string
        updatedAt?: string
      } | null
    }>(path)
    return toOEMConfig(response)
  },

  async getAdminOEMBrands(): Promise<OEMBrandRecord[]> {
    const response = await request<{
      items: Array<{
        brand: Parameters<typeof toOEMBrand>[0]
        theme?: Parameters<typeof toOEMTheme>[0]
        features?: Parameters<typeof toOEMFeatureFlags>[0]
        bindings?: Array<{
          tenantId: number
          brandId: number
          bindingMode: string
          updatedAt?: string
        }>
      }>
    }>('/api/v1/admin/oem/brands')
    return response.items.map(toOEMBrandRecord)
  },

  async getAdminOEMBrandDetail(id: string): Promise<OEMBrandRecord> {
    const response = await request<{
      brand: Parameters<typeof toOEMBrand>[0]
      theme?: Parameters<typeof toOEMTheme>[0]
      features?: Parameters<typeof toOEMFeatureFlags>[0]
      bindings?: Array<{
        tenantId: number
        brandId: number
        bindingMode: string
        updatedAt?: string
      }>
    }>(`/api/v1/admin/oem/brands/${id}`)
    return toOEMBrandRecord(response)
  },

  async updateAdminOEMBrand(id: string, payload: UpdateOEMBrandPayload): Promise<OEMBrand> {
    const response = await request<{
      brand: Parameters<typeof toOEMBrand>[0]
    }>(`/api/v1/admin/oem/brands/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
    return toOEMBrand(response.brand)
  },

  async updateAdminOEMBrandTheme(id: string, payload: UpdateOEMThemePayload): Promise<OEMTheme | null> {
    const response = await request<{
      theme?: Parameters<typeof toOEMTheme>[0]
    }>(`/api/v1/admin/oem/brands/${id}/theme`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
    return toOEMTheme(response.theme)
  },

  async updateAdminOEMBrandFeatures(id: string, payload: UpdateOEMFeatureFlagsPayload): Promise<OEMFeatureFlags | null> {
    const response = await request<{
      features?: Parameters<typeof toOEMFeatureFlags>[0]
    }>(`/api/v1/admin/oem/brands/${id}/features`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
    return toOEMFeatureFlags(response.features)
  },

  async replaceAdminOEMBrandBindings(
    id: string,
    payload: ReplaceTenantBrandBindingsPayload,
  ): Promise<OEMBrandRecord['bindings']> {
    const response = await request<{
      bindings?: Array<{
        tenantId: number
        brandId: number
        bindingMode: string
        updatedAt?: string
      }>
    }>(`/api/v1/admin/oem/brands/${id}/bindings`, {
      method: 'PUT',
      body: JSON.stringify({
        bindings: payload.bindings.map((item) => ({
          tenantId: Number(item.tenantId),
          bindingMode: item.bindingMode,
        })),
      }),
    })
    return (response.bindings ?? [])
      .map((item) => toTenantBrandBinding(item))
      .filter((item): item is NonNullable<typeof item> => Boolean(item))
  },

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

  async searchLogs(query: { q?: string; kind?: string; instanceId?: string }, scope: 'portal' | 'admin' = 'portal') {
    const params = new URLSearchParams()
    if (query.q) params.set('q', query.q)
    if (query.kind) params.set('kind', query.kind)
    if (query.instanceId) params.set('instanceId', query.instanceId)
    params.set('scope', scope)
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
      description: '当前界面已接入平台 API，后续可继续映射到真实控制面与业务数据。',
      primaryInstanceId: overview.primaryInstanceId ? String(overview.primaryInstanceId) : undefined,
    }
  },

  async getPortalSelfServiceSummary(): Promise<PortalSelfServiceSummary> {
    const response = await request<{
      tenant: {
        id: number
        code: string
        name: string
        plan: string
        status: string
        expiredAt?: string
        supportEmail?: string
        supportUrl?: string
      }
      launchpad: {
        primaryInstanceId?: number
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
        steps: Array<{
          key: string
          title: string
          description: string
          status: 'completed' | 'ready' | 'pending'
          actionLabel: string
          actionPath: string
          result?: string
        }>
      }
      metrics: Array<{
        label: string
        value: string
        delta?: string
        tone?: 'neutral' | 'positive' | 'warning' | 'critical'
      }>
      quotas: Array<{
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
      }>
      reminders: Array<{
        key: string
        severity: 'info' | 'warning' | 'critical'
        title: string
        description: string
        actionLabel: string
        actionPath: string
        at?: string
      }>
      recentSessions: Array<{
        id: number
        title: string
        instanceId: number
        instanceName: string
        sessionNo: string
        status: string
        updatedAt: string
        messageCount: number
        artifactCount: number
        workspacePath: string
      }>
      recentArtifacts: Array<{
        id: number
        title: string
        kind: string
        sourceUrl: string
        previewUrl?: string
        createdAt: string
        updatedAt: string
        instanceId: number
        instanceName: string
        instanceStatus: string
        sessionId: number
        sessionTitle: string
        sessionStatus: string
        workspacePath: string
      }>
    }>('/api/v1/portal/self-service')

    return {
      tenant: {
        id: String(response.tenant.id),
        code: response.tenant.code,
        name: response.tenant.name,
        plan: response.tenant.plan,
        status: response.tenant.status,
        expiredAt: formatDateTime(response.tenant.expiredAt),
        supportEmail: response.tenant.supportEmail,
        supportUrl: response.tenant.supportUrl,
      },
      launchpad: {
        primaryInstanceId: response.launchpad.primaryInstanceId ? String(response.launchpad.primaryInstanceId) : undefined,
        primaryInstanceName: response.launchpad.primaryInstanceName,
        workspacePath: response.launchpad.workspacePath,
        artifactsPath: response.launchpad.artifactsPath,
        workspaceUrl: response.launchpad.workspaceUrl,
      },
      experience: response.experience,
      onboarding: response.onboarding,
      metrics: response.metrics,
      quotas: response.quotas,
      reminders: response.reminders.map((item) => ({
        ...item,
        at: formatDateTime(item.at),
      })),
      recentSessions: response.recentSessions.map((item) => ({
        id: String(item.id),
        title: item.title,
        instanceId: String(item.instanceId),
        instanceName: item.instanceName,
        sessionNo: item.sessionNo,
        status: item.status,
        updatedAt: formatDateTime(item.updatedAt),
        messageCount: item.messageCount,
        artifactCount: item.artifactCount,
        workspacePath: item.workspacePath,
      })),
      recentArtifacts: response.recentArtifacts.map((item) =>
        toPortalArtifactCenterItem({
          ...item,
          preview: {
            available: Boolean(item.previewUrl),
            mode: 'download',
            strategy: 'download',
            externalUrl: item.previewUrl || item.sourceUrl,
          },
        }),
      ),
    }
  },

  async getPortalAccountSettings(): Promise<AccountSettings> {
    const response = await request<{ settings: {
      tenantId: number
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
      notifyChannelEmail?: boolean
      notifyChannelWebhook?: boolean
      notifyChannelInApp?: boolean
      notificationWebhookUrl?: string
      portalHeadline?: string
      portalSubtitle?: string
      workspaceCallout?: string
      experimentBadge?: string
      updatedAt: string
    } }>('/api/v1/portal/account/settings')
    return toAccountSettings(response.settings)
  },

  async updatePortalAccountSettings(payload: Partial<AccountSettings>): Promise<AccountSettings> {
    const response = await request<{ settings: {
      tenantId: number
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
      notifyChannelEmail?: boolean
      notifyChannelWebhook?: boolean
      notifyChannelInApp?: boolean
      notificationWebhookUrl?: string
      portalHeadline?: string
      portalSubtitle?: string
      workspaceCallout?: string
      experimentBadge?: string
      updatedAt: string
    } }>('/api/v1/portal/account/settings', {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
    return toAccountSettings(response.settings)
  },

  async getPortalOpsReport(): Promise<PortalOpsReport> {
    const response = await request<{
      tenant: {
        id: number
        name: string
        plan: string
        status: string
      }
      brand: {
        name?: string
        supportEmail?: string
        supportUrl?: string
      }
      settings: {
        tenantId: number
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
        notifyChannelEmail?: boolean
        notifyChannelWebhook?: boolean
        notifyChannelInApp?: boolean
        notificationWebhookUrl?: string
        portalHeadline?: string
        portalSubtitle?: string
        workspaceCallout?: string
        experimentBadge?: string
        updatedAt: string
      } | null
      summary: PortalOpsReport['summary']
      notificationChannels: Array<{
        key: string
        label: string
        enabled: boolean
        target?: string
        description: string
      }>
      notificationTemplates: Array<{
        key: string
        title: string
        subject: string
        body: string
      }>
      monthlyUsage: Array<{
        label: string
        chargeAmount: number
        paidAmount: number
        sessions: number
        messages: number
        artifacts: number
      }>
      export: {
        csvPath: string
      }
    }>('/api/v1/portal/ops/report')
    return {
      tenant: {
        id: String(response.tenant.id),
        name: response.tenant.name,
        plan: response.tenant.plan,
        status: response.tenant.status,
      },
      brand: response.brand,
      settings: response.settings ? toAccountSettings(response.settings) : null,
      summary: response.summary,
      notificationChannels: response.notificationChannels.map(toPortalNotificationChannel),
      notificationTemplates: response.notificationTemplates.map(toPortalNotificationTemplatePreview),
      monthlyUsage: response.monthlyUsage.map(toPortalOpsMonthlyUsage),
      export: response.export,
    }
  },

  async getPortalArtifactCenter(filters?: { q?: string; kind?: string; instanceId?: string }): Promise<ArtifactCenterResponse> {
    const response = await request<{
      items: Array<{
      id: number
      title: string
      kind: string
      sourceUrl: string
      previewUrl?: string
      archiveStatus?: string
      contentType?: string
      sizeBytes?: number
      filename?: string
      createdAt: string
      updatedAt: string
      instanceId: number
      instanceName: string
      instanceStatus: string
      sessionId: number
      sessionNo?: string
      sessionTitle: string
      sessionStatus: string
      messageId?: number
      messagePreview?: string
      tenantId?: number
      tenantName?: string
      viewCount?: number
      downloadCount?: number
      workspacePath: string
      detailPath?: string
      lineageKey?: string
      version?: number
      latestVersion?: number
      parentArtifactId?: number
      isFavorite?: boolean
      favoriteCount?: number
      shareCount?: number
      thumbnail?: {
        mode?: string
        url?: string
        label?: string
        hint?: string
      }
      quality?: {
        status?: string
        score?: number
        inlinePreview?: boolean
        previewMode?: string
        strategy?: string
        failureReason?: string
        lastViewedAt?: string
        lastDownloadedAt?: string
        viewCount?: number
        downloadCount?: number
      }
      preview: {
        available?: boolean
        mode?: string
        strategy?: string
        sandboxed?: boolean
        proxied?: boolean
        previewUrl?: string
        downloadUrl?: string
        externalUrl?: string
        failureReason?: string
        note?: string
      }
    }>
      recentViewed: Array<{
        id: number
        title: string
        kind: string
        sourceUrl: string
        previewUrl?: string
        archiveStatus?: string
        contentType?: string
        sizeBytes?: number
        filename?: string
        createdAt: string
        updatedAt: string
        instanceId: number
        instanceName: string
        instanceStatus: string
        sessionId: number
        sessionNo?: string
        sessionTitle: string
        sessionStatus: string
        messageId?: number
        messagePreview?: string
        tenantId?: number
        tenantName?: string
        viewCount?: number
        downloadCount?: number
        workspacePath: string
        detailPath?: string
        lineageKey?: string
        version?: number
        latestVersion?: number
        parentArtifactId?: number
        isFavorite?: boolean
        favoriteCount?: number
        shareCount?: number
        thumbnail?: {
          mode?: string
          url?: string
          label?: string
          hint?: string
        }
        quality?: {
          status?: string
          score?: number
          inlinePreview?: boolean
          previewMode?: string
          strategy?: string
          failureReason?: string
          lastViewedAt?: string
          lastDownloadedAt?: string
          viewCount?: number
          downloadCount?: number
        }
        preview: {
          available?: boolean
          mode?: string
          strategy?: string
          sandboxed?: boolean
          proxied?: boolean
          previewUrl?: string
          downloadUrl?: string
          externalUrl?: string
          failureReason?: string
          note?: string
        }
      }>
      stats: {
        totalCount: number
        favoriteCount: number
        sharedCount: number
        versionedCount: number
        inlinePreviewCount: number
        fallbackCount: number
        recentViewedCount: number
        failureReasons: Array<{
          reason: string
          count: number
        }>
      }
    }>(appendQuery('/api/v1/portal/artifacts', {
      q: filters?.q,
      kind: filters?.kind,
      instanceId: filters?.instanceId,
    }))
    return {
      items: response.items.map(toPortalArtifactCenterItem),
      recentViewed: response.recentViewed.map(toPortalArtifactCenterItem),
      stats: response.stats,
    }
  },

  async getAdminArtifactCenter(filters?: { q?: string; kind?: string; instanceId?: string }): Promise<ArtifactCenterResponse> {
    const response = await request<{
      items: Array<{
      id: number
      title: string
      kind: string
      sourceUrl: string
      previewUrl?: string
      archiveStatus?: string
      contentType?: string
      sizeBytes?: number
      filename?: string
      createdAt: string
      updatedAt: string
      instanceId: number
      instanceName: string
      instanceStatus: string
      sessionId: number
      sessionNo?: string
      sessionTitle: string
      sessionStatus: string
      messageId?: number
      messagePreview?: string
      tenantId?: number
      tenantName?: string
      viewCount?: number
      downloadCount?: number
      workspacePath: string
      detailPath?: string
      lineageKey?: string
      version?: number
      latestVersion?: number
      parentArtifactId?: number
      isFavorite?: boolean
      favoriteCount?: number
      shareCount?: number
      thumbnail?: {
        mode?: string
        url?: string
        label?: string
        hint?: string
      }
      quality?: {
        status?: string
        score?: number
        inlinePreview?: boolean
        previewMode?: string
        strategy?: string
        failureReason?: string
        lastViewedAt?: string
        lastDownloadedAt?: string
        viewCount?: number
        downloadCount?: number
      }
      preview: {
        available?: boolean
        mode?: string
        strategy?: string
        sandboxed?: boolean
        proxied?: boolean
        previewUrl?: string
        downloadUrl?: string
        externalUrl?: string
        failureReason?: string
        note?: string
      }
    }>
      recentViewed: Array<{
        id: number
        title: string
        kind: string
        sourceUrl: string
        previewUrl?: string
        archiveStatus?: string
        contentType?: string
        sizeBytes?: number
        filename?: string
        createdAt: string
        updatedAt: string
        instanceId: number
        instanceName: string
        instanceStatus: string
        sessionId: number
        sessionNo?: string
        sessionTitle: string
        sessionStatus: string
        messageId?: number
        messagePreview?: string
        tenantId?: number
        tenantName?: string
        viewCount?: number
        downloadCount?: number
        workspacePath: string
        detailPath?: string
        lineageKey?: string
        version?: number
        latestVersion?: number
        parentArtifactId?: number
        isFavorite?: boolean
        favoriteCount?: number
        shareCount?: number
        thumbnail?: {
          mode?: string
          url?: string
          label?: string
          hint?: string
        }
        quality?: {
          status?: string
          score?: number
          inlinePreview?: boolean
          previewMode?: string
          strategy?: string
          failureReason?: string
          lastViewedAt?: string
          lastDownloadedAt?: string
          viewCount?: number
          downloadCount?: number
        }
        preview: {
          available?: boolean
          mode?: string
          strategy?: string
          sandboxed?: boolean
          proxied?: boolean
          previewUrl?: string
          downloadUrl?: string
          externalUrl?: string
          failureReason?: string
          note?: string
        }
      }>
      stats: {
        totalCount: number
        favoriteCount: number
        sharedCount: number
        versionedCount: number
        inlinePreviewCount: number
        fallbackCount: number
        recentViewedCount: number
        failureReasons: Array<{
          reason: string
          count: number
        }>
      }
    }>(appendQuery('/api/v1/admin/artifacts', {
      q: filters?.q,
      kind: filters?.kind,
      instanceId: filters?.instanceId,
    }))
    return {
      items: response.items.map(toPortalArtifactCenterItem),
      recentViewed: response.recentViewed.map(toPortalArtifactCenterItem),
      stats: response.stats,
    }
  },

  async getArtifactCenterDetail(id: string, scope: 'portal' | 'admin' = 'portal', shareToken?: string): Promise<ArtifactCenterDetail> {
    const basePath = scope === 'admin' ? `/api/v1/admin/artifacts/${id}` : `/api/v1/portal/artifacts/${id}`
    const path = appendQuery(basePath, { share: shareToken })
    const response = await request<{
      artifact: {
        id: number
        title: string
        kind: string
        sourceUrl: string
        previewUrl?: string
        archiveStatus?: string
        contentType?: string
        sizeBytes?: number
        filename?: string
        createdAt: string
        updatedAt: string
        instanceId: number
        instanceName: string
        instanceStatus: string
        sessionId: number
        sessionNo?: string
        sessionTitle: string
        sessionStatus: string
        messageId?: number
        messagePreview?: string
        tenantId?: number
        tenantName?: string
        viewCount?: number
        downloadCount?: number
        workspacePath: string
        detailPath?: string
        lineageKey?: string
        version?: number
        latestVersion?: number
        parentArtifactId?: number
        isFavorite?: boolean
        favoriteCount?: number
        shareCount?: number
        thumbnail?: {
          mode?: string
          url?: string
          label?: string
          hint?: string
        }
        quality?: {
          status?: string
          score?: number
          inlinePreview?: boolean
          previewMode?: string
          strategy?: string
          failureReason?: string
          lastViewedAt?: string
          lastDownloadedAt?: string
          viewCount?: number
          downloadCount?: number
        }
      }
      preview: {
        available?: boolean
        mode?: string
        strategy?: string
        sandboxed?: boolean
        proxied?: boolean
        previewUrl?: string
        downloadUrl?: string
        externalUrl?: string
        failureReason?: string
        note?: string
      }
      session?: {
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
      }
      message?: {
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        parentMessageId?: number
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
      } | null
      accessLogs?: Array<{
        id: number
        artifactId: number
        sessionId: number
        tenantId: number
        instanceId: number
        action: string
        scope: string
        actor: string
        remoteAddr?: string
        userAgent?: string
        createdAt: string
      }>
      versions?: Array<{
        id: number
        title: string
        kind: string
        sourceUrl: string
        previewUrl?: string
        archiveStatus?: string
        contentType?: string
        sizeBytes?: number
        filename?: string
        createdAt: string
        updatedAt: string
        instanceId: number
        instanceName: string
        instanceStatus: string
        sessionId: number
        sessionNo?: string
        sessionTitle: string
        sessionStatus: string
        messageId?: number
        messagePreview?: string
        tenantId?: number
        tenantName?: string
        viewCount?: number
        downloadCount?: number
        workspacePath: string
        detailPath?: string
        lineageKey?: string
        version?: number
        latestVersion?: number
        parentArtifactId?: number
        isFavorite?: boolean
        favoriteCount?: number
        shareCount?: number
        thumbnail?: {
          mode?: string
          url?: string
          label?: string
          hint?: string
        }
        quality?: {
          status?: string
          score?: number
          inlinePreview?: boolean
          previewMode?: string
          strategy?: string
          failureReason?: string
          lastViewedAt?: string
          lastDownloadedAt?: string
          viewCount?: number
          downloadCount?: number
        }
        preview: {
          available?: boolean
          mode?: string
          strategy?: string
          sandboxed?: boolean
          proxied?: boolean
          previewUrl?: string
          downloadUrl?: string
          externalUrl?: string
          failureReason?: string
          note?: string
        }
      }>
      shares?: Array<{
        id: number
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
      }>
    }>(path)
    return {
      artifact: toPortalArtifactCenterItem({
        ...response.artifact,
        preview: response.preview,
      }),
      preview: toArtifactPreviewDescriptor(response.preview),
      session: response.session ? toWorkspaceSession(response.session) : undefined,
      message: response.message ? toWorkspaceMessage(response.message) : null,
      accessLogs: (response.accessLogs ?? []).map(toArtifactAccessLog),
      versions: (response.versions ?? []).map(toPortalArtifactCenterItem),
      shares: (response.shares ?? []).map(toArtifactShare),
    }
  },

  async favoriteArtifact(id: string, favorite: boolean, scope: 'portal' | 'admin' = 'portal') {
    const path = scope === 'admin' ? `/api/v1/admin/artifacts/${id}/favorite` : `/api/v1/portal/artifacts/${id}/favorite`
    return request<{ artifactId: number; isFavorite: boolean; favoriteCount: number }>(path, {
      method: favorite ? 'POST' : 'DELETE',
    })
  },

  async createArtifactShare(
    id: string,
    payload: { note?: string; expiresAt?: string; expiresInDays?: number },
    scope: 'portal' | 'admin' = 'portal',
  ) {
    const path = scope === 'admin' ? `/api/v1/admin/artifacts/${id}/shares` : `/api/v1/portal/artifacts/${id}/shares`
    const response = await request<{ share: {
      id: number
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
    } }>(path, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    return toArtifactShare(response.share)
  },

  async revokeArtifactShare(id: string, scope: 'portal' | 'admin' = 'portal') {
    const path = scope === 'admin' ? `/api/v1/admin/artifact-shares/${id}` : `/api/v1/portal/artifact-shares/${id}`
    const response = await request<{ share: {
      id: number
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
    } }>(path, {
      method: 'DELETE',
    })
    return toArtifactShare(response.share)
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
        sessionId: item.sessionId ? String(item.sessionId) : undefined,
        messageId: item.messageId ? String(item.messageId) : undefined,
        traceId: item.traceId,
        instancePath: item.instancePath,
        workspacePath: item.workspacePath,
      })),
      resourceTrend: response.resourceTrend,
      workspaceSessions: response.workspaceSessions.map(toWorkspaceSession),
      bridgeSummary: {
        traceCount: response.bridgeSummary.traceCount,
        eventCount: response.bridgeSummary.eventCount,
        failedTraceCount: response.bridgeSummary.failedTraceCount,
        artifactCount: response.bridgeSummary.artifactCount,
        lastEventAt: formatOptionalDateTime(response.bridgeSummary.lastEventAt),
        recentTraces: response.bridgeSummary.recentTraces.map((item) => ({
          traceId: item.traceId,
          sessionId: item.sessionId ? String(item.sessionId) : undefined,
          sessionNo: item.sessionNo,
          latestAt: formatDateTime(item.latestAt),
          status: item.status,
          preview: item.preview,
          messageCount: item.messageCount,
          eventCount: item.eventCount,
          artifactCount: item.artifactCount,
          toolCount: item.toolCount,
        })),
      },
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

  async getAdminInstanceDiagnostics(id: string): Promise<AdminInstanceDiagnosticsSummary> {
    const response = await request<{
      binding?: {
        clusterId: string
        namespace: string
        workloadId: string
        workloadName: string
      } | null
      workload?: {
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
      } | null
      metrics?: {
        workloadId: string
        cpuUsageMilli: number
        memoryUsageMB: number
        networkRxKB: number
        networkTxKB: number
        errorRatePercent: number
        requestsPerMinute: number
      } | null
      pods: Array<{
        id: string
        name: string
        nodeName: string
        status: string
        restarts: number
        startedAt?: string
        workloadId?: string
        image?: string
      }>
      signals: Array<{
        type: string
        severity: string
        summary: string
        triggeredAt?: string
        podName?: string
        restarts?: number
      }>
      policy: {
        defaultAccessMode: 'readonly' | 'whitelist'
        readonlyTtlMinutes: number
        whitelistTtlMinutes: number
        requiresApprovalForWhitelist: boolean
        maxActiveSessionsPerInstance: number
        commandCatalog: Array<{
          key: string
          label: string
          description: string
          commandText: string
          manualAllowed: boolean
        }>
      }
      sessions: Array<{
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
        operatorUserId?: number
        reason?: string
        closeReason?: string
        expiresAt?: string
        lastCommandAt?: string
        startedAt?: string
        endedAt?: string
        createdAt: string
        updatedAt: string
        commandCount?: number
        lastCommandStatus?: string
        lastCommandText?: string
      }>
    }>(`/api/v1/admin/instances/${id}/diagnostics`)

    return {
      binding: toRuntimeBinding(response.binding ?? null),
      workload: toRuntimeWorkload(response.workload ?? null),
      metrics: toRuntimeWorkloadMetrics(response.metrics ?? null),
      pods: response.pods.map(toDiagnosticPod),
      signals: response.signals.map(toDiagnosticSignal),
      policy: toDiagnosticPolicy(response.policy),
      sessions: response.sessions.map(toDiagnosticSessionSummary),
    }
  },

  async createAdminDiagnosticSession(
    instanceId: string,
    payload: {
      podName?: string
      containerName?: string
      accessMode?: 'readonly' | 'whitelist'
      approvalTicket?: string
      approvedBy?: string
      reason?: string
    },
  ) {
    const response = await request<{
      session: {
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
        operatorUserId?: number
        reason?: string
        closeReason?: string
        expiresAt?: string
        lastCommandAt?: string
        startedAt?: string
        endedAt?: string
        createdAt: string
        updatedAt: string
        commandCount?: number
        lastCommandStatus?: string
        lastCommandText?: string
      }
    }>(`/api/v1/admin/instances/${instanceId}/diagnostic-sessions`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })

    return toDiagnosticSessionSummary(response.session)
  },

  async getAdminDiagnosticSession(id: string): Promise<AdminDiagnosticSessionDetail> {
    const response = await request<{
      session: {
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
        operatorUserId?: number
        reason?: string
        closeReason?: string
        expiresAt?: string
        lastCommandAt?: string
        startedAt?: string
        endedAt?: string
        createdAt: string
        updatedAt: string
        commandCount?: number
        lastCommandStatus?: string
        lastCommandText?: string
      }
      commands: Array<{
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        commandKey?: string
        commandText: string
        status: string
        exitCode: number
        durationMs: number
        output?: string
        errorOutput?: string
        outputTruncated?: boolean
        executedAt: string
      }>
      record: string
      commandCatalog: Array<{
        key: string
        label: string
        description: string
        commandText: string
        manualAllowed: boolean
      }>
    }>(`/api/v1/admin/diagnostic-sessions/${id}`)

    return {
      session: toDiagnosticSessionSummary(response.session),
      commands: response.commands.map(toDiagnosticCommandRecord),
      record: response.record,
      commandCatalog: response.commandCatalog.map(toDiagnosticCommandCatalogItem),
    }
  },

  async executeAdminDiagnosticCommand(
    sessionId: string,
    payload: { commandKey?: string; commandText?: string },
  ) {
    const response = await request<{
      session: {
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
        operatorUserId?: number
        reason?: string
        closeReason?: string
        expiresAt?: string
        lastCommandAt?: string
        startedAt?: string
        endedAt?: string
        createdAt: string
        updatedAt: string
        commandCount?: number
        lastCommandStatus?: string
        lastCommandText?: string
      }
      command: {
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
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
    }>(`/api/v1/admin/diagnostic-sessions/${sessionId}/commands`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })

    return {
      session: toDiagnosticSessionSummary(response.session),
      command: toDiagnosticCommandRecord(response.command),
    }
  },

  async closeAdminDiagnosticSession(sessionId: string, reason?: string) {
    const response = await request<{
      session: {
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
        operatorUserId?: number
        reason?: string
        closeReason?: string
        expiresAt?: string
        lastCommandAt?: string
        startedAt?: string
        endedAt?: string
        createdAt: string
        updatedAt: string
        commandCount?: number
        lastCommandStatus?: string
        lastCommandText?: string
      }
    }>(`/api/v1/admin/diagnostic-sessions/${sessionId}/close`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    })

    return toDiagnosticSessionSummary(response.session)
  },

  async getAdminTerminalSessions(query: { status?: string; instanceId?: string; q?: string } = {}) {
    const suffix = appendQuery('/api/v1/admin/terminal-sessions', query)
    const response = await request<{
      items: Array<{
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
        operatorUserId?: number
        reason?: string
        closeReason?: string
        expiresAt?: string
        lastCommandAt?: string
        startedAt?: string
        endedAt?: string
        createdAt: string
        updatedAt: string
        commandCount?: number
        lastCommandStatus?: string
        lastCommandText?: string
      }>
    }>(suffix)

    return response.items.map(toDiagnosticSessionSummary)
  },

  async getApprovals(scope: 'portal' | 'admin', query: { status?: string; type?: string; instanceId?: string; tenantId?: string } = {}) {
    const suffix = appendQuery(`/api/v1/${scope}/approvals`, query)
    const response = await request<{ items: Parameters<typeof toApprovalSummary>[0][] }>(suffix)
    return response.items.map(toApprovalSummary)
  },

  async createApprovalRequest(
    scope: 'portal' | 'admin',
    payload: {
      approvalType: string
      targetType?: string
      targetId: number
      instanceId?: number
      riskLevel?: 'medium' | 'high' | 'critical'
      reason: string
      comment?: string
      metadata?: Record<string, string>
    },
  ) {
    const response = await request<{
      approval: Parameters<typeof toApprovalSummary>[0]
      actions: Parameters<typeof toApprovalActionRecord>[0][]
      instance?: ApiInstance
    }>(`/api/v1/${scope}/approvals`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    return toApprovalDetail(response)
  },

  async getAdminApprovalDetail(id: string): Promise<ApprovalDetail> {
    const response = await request<{
      approval: Parameters<typeof toApprovalSummary>[0]
      actions: Parameters<typeof toApprovalActionRecord>[0][]
      instance?: ApiInstance
    }>(`/api/v1/admin/approvals/${id}`)
    return toApprovalDetail(response)
  },

  async approveAdminApproval(id: string, comment?: string) {
    const response = await request<{
      approval: Parameters<typeof toApprovalSummary>[0]
      actions: Parameters<typeof toApprovalActionRecord>[0][]
      instance?: ApiInstance
    }>(`/api/v1/admin/approvals/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify({ comment }),
    })
    return toApprovalDetail(response)
  },

  async rejectAdminApproval(id: string, rejectReason: string, comment?: string) {
    const response = await request<{
      approval: Parameters<typeof toApprovalSummary>[0]
      actions: Parameters<typeof toApprovalActionRecord>[0][]
      instance?: ApiInstance
    }>(`/api/v1/admin/approvals/${id}/reject`, {
      method: 'POST',
      body: JSON.stringify({ rejectReason, comment }),
    })
    return toApprovalDetail(response)
  },

  async executeAdminApproval(id: string) {
    const response = await request<{
      approval: Parameters<typeof toApprovalSummary>[0]
      actions: Parameters<typeof toApprovalActionRecord>[0][]
      instance?: ApiInstance
      job?: ApiJob
      session?: {
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
        operatorUserId?: number
        reason?: string
        closeReason?: string
        expiresAt?: string
        lastCommandAt?: string
        startedAt?: string
        endedAt?: string
        createdAt: string
        updatedAt: string
        commandCount?: number
        lastCommandStatus?: string
        lastCommandText?: string
      }
    }>(`/api/v1/admin/approvals/${id}/execute`, {
      method: 'POST',
    })

    return {
      approval: toApprovalSummary(response.approval),
      actions: response.actions.map(toApprovalActionRecord),
      instance: response.instance ? toInstance(response.instance, []) : undefined,
      job: response.job ? toJob(response.job) : undefined,
      session: response.session ? toDiagnosticSessionSummary(response.session) : undefined,
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
      orders: response.orders.map((item: any) => toOrder(item as ApiOrder)),
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

  async getChannels(scope: 'portal' | 'admin' = 'portal'): Promise<Channel[]> {
    const response = await request<{ items: ApiChannelListItem[] }>(appendQuery('/api/v1/channels', { scope }))
    return response.items.map((item) => toChannel(item.channel, item.activities))
  },

  async getChannelDetail(id: string, scope: 'portal' | 'admin' = 'portal'): Promise<Channel> {
    const response = await request<ApiChannelDetail>(appendQuery(`/api/v1/channels/${id}`, { scope }))
    return toChannel(response.channel, response.activities)
  },

  async connectChannel(payload: ChannelConnectPayload) {
    const targetId = payload.channelId || (await this.getChannels('portal')).find((item) => item.provider === payload.provider)?.id
    if (!targetId) {
      throw new Error('后端尚未返回该渠道资源，请先确认 provider 已在平台侧注册')
    }

    return request(`/api/v1/channels/${targetId}/connect`, {
      method: 'POST',
      body: JSON.stringify({
        provider: payload.provider,
        method: payload.authMode,
        redirectUri: payload.redirectUri,
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

  async getWorkspaceSessions(
    instanceId: string,
    scope: 'portal' | 'admin' = 'portal',
    filters?: { q?: string; status?: string; limit?: number },
  ) {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/instances/${instanceId}/workspace/sessions`
        : `/api/v1/portal/instances/${instanceId}/workspace/sessions`
    const response = await request<{ items: Array<{
      id: number
      sessionNo: string
      tenantId: number
      instanceId: number
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
    }> }>(
      appendQuery(path, {
        q: filters?.q,
        status: filters?.status,
        limit: filters?.limit ? String(filters.limit) : undefined,
      }),
    )
    return response.items.map(toWorkspaceSession)
  },

  async searchWorkspaceSessions(
    scope: 'portal' | 'admin' = 'portal',
    filters?: { q?: string; status?: string; tenantId?: string; instanceId?: string; hasArtifacts?: boolean; limit?: number },
  ) {
    const path =
      scope === 'admin'
        ? '/api/v1/admin/workspace/sessions'
        : '/api/v1/portal/workspace/sessions'
    const response = await request<{ items: Array<{
      id: number
      sessionNo: string
      tenantId: number
      instanceId: number
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
    }> }>(
      appendQuery(path, {
        q: filters?.q,
        status: filters?.status,
        tenantId: filters?.tenantId,
        instanceId: filters?.instanceId,
        hasArtifacts:
          typeof filters?.hasArtifacts === 'boolean' ? String(filters.hasArtifacts) : undefined,
        limit: filters?.limit ? String(filters.limit) : undefined,
      }),
    )
    return response.items.map(toWorkspaceSession)
  },

  async createWorkspaceSession(instanceId: string, payload: { title?: string; workspaceUrl?: string }, scope: 'portal' | 'admin' = 'portal') {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/instances/${instanceId}/workspace/sessions`
        : `/api/v1/portal/instances/${instanceId}/workspace/sessions`
    const response = await request<{ session: {
      id: number
      sessionNo: string
      tenantId: number
      instanceId: number
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
    } }>(path, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    return toWorkspaceSession(response.session)
  },

  async updateWorkspaceSessionStatus(
    sessionId: string,
    payload: { status: 'active' | 'archived' },
    scope: 'portal' | 'admin' = 'portal',
  ) {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/sessions/${sessionId}/status`
        : `/api/v1/portal/workspace/sessions/${sessionId}/status`
    const response = await request<{ session: {
      id: number
      sessionNo: string
      tenantId: number
      instanceId: number
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
    } }>(path, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
    return toWorkspaceSession(response.session)
  },

  async getWorkspaceSessionDetail(
    sessionId: string,
    scope: 'portal' | 'admin' = 'portal',
    options?: { messageLimit?: number; eventLimit?: number; anchorMessageId?: string; anchorTraceId?: string },
  ) {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/sessions/${sessionId}`
        : `/api/v1/portal/workspace/sessions/${sessionId}`
    const response = await request<{
      session: {
        id: number
        sessionNo: string
        tenantId: number
        instanceId: number
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
      artifacts: Array<{
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        messageId?: number
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
      }>
      messages: Array<{
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        parentMessageId?: number
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
      }>
      messagesHasMore?: boolean
      events?: Array<{
        id: number
        sessionId: number
        messageId?: number
        tenantId: number
        instanceId: number
        eventType: string
        origin?: string
        traceId?: string
        payload?: Record<string, unknown>
        createdAt: string
      }>
      eventsHasMore?: boolean
    }>(
      appendQuery(path, {
        messageLimit: options?.messageLimit ? String(options.messageLimit) : undefined,
        eventLimit: options?.eventLimit ? String(options.eventLimit) : undefined,
        anchorMessageId: options?.anchorMessageId,
        anchorTraceId: options?.anchorTraceId,
      }),
    )
    return {
      session: toWorkspaceSession(response.session),
      artifacts: response.artifacts.map(toWorkspaceArtifact),
      messages: response.messages.map(toWorkspaceMessage),
      events: response.events?.map(toWorkspaceSessionEvent) ?? [],
      messagesHasMore: Boolean(response.messagesHasMore),
      eventsHasMore: Boolean(response.eventsHasMore),
    }
  },

  async createWorkspaceArtifact(
    sessionId: string,
    payload: { title: string; kind: string; sourceUrl: string; previewUrl?: string },
    scope: 'portal' | 'admin' = 'portal',
  ) {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/sessions/${sessionId}/artifacts`
        : `/api/v1/portal/workspace/sessions/${sessionId}/artifacts`
    const response = await request<{ artifact: {
      id: number
      sessionId: number
      tenantId: number
      instanceId: number
      messageId?: number
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
    } }>(path, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    return toWorkspaceArtifact(response.artifact)
  },

  async getWorkspaceArtifactPreview(artifactId: string, scope: 'portal' | 'admin' = 'portal') {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/artifacts/${artifactId}/preview`
        : `/api/v1/portal/workspace/artifacts/${artifactId}/preview`
    const response = await request<{
      preview: {
        available: boolean
        mode: 'html' | 'pdf' | 'image' | 'video' | 'audio' | 'text' | 'download'
        strategy: string
        sandboxed: boolean
        proxied: boolean
        previewUrl?: string
        downloadUrl?: string
        externalUrl?: string
        failureReason?: string
        note?: string
      }
    }>(path)
    return response.preview as WorkspaceArtifactPreview
  },

  async getWorkspaceMessages(
    sessionId: string,
    scope: 'portal' | 'admin' = 'portal',
    filters?: { q?: string; role?: string; beforeId?: string; limit?: number },
  ) {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/sessions/${sessionId}/messages`
        : `/api/v1/portal/workspace/sessions/${sessionId}/messages`
    const response = await request<{ items: Array<{
      id: number
      sessionId: number
      tenantId: number
      instanceId: number
      parentMessageId?: number
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
    }>
      hasMore?: boolean
    }>(
      appendQuery(path, {
        q: filters?.q,
        role: filters?.role,
        beforeId: filters?.beforeId,
        limit: filters?.limit ? String(filters.limit) : undefined,
      }),
    )
    return {
      items: response.items.map(toWorkspaceMessage),
      hasMore: Boolean(response.hasMore),
    }
  },

  async createWorkspaceMessage(
    sessionId: string,
    payload: { role?: string; status?: string; content: string; dispatch?: boolean },
    scope: 'portal' | 'admin' = 'portal',
  ) {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/sessions/${sessionId}/messages`
        : `/api/v1/portal/workspace/sessions/${sessionId}/messages`
    const response = await request<{
      message: {
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        parentMessageId?: number
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
      dispatch?: {
        ok: boolean
        target?: string
        traceId?: string
        error?: string
        message?: string
      }
      reply?: {
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        parentMessageId?: number
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
      artifacts?: Array<{
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        messageId?: number
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
      }>
    }>(path, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    return toWorkspaceMessageDispatchResponse(response)
  },

  async retryWorkspaceMessage(messageId: string, scope: 'portal' | 'admin' = 'portal') {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/messages/${messageId}/retry`
        : `/api/v1/portal/workspace/messages/${messageId}/retry`
    const response = await request<{
      message: {
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        parentMessageId?: number
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
      dispatch?: {
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
      reply?: {
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        parentMessageId?: number
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
      artifacts?: Array<{
        id: number
        sessionId: number
        tenantId: number
        instanceId: number
        messageId?: number
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
      }>
    }>(path, {
      method: 'POST',
    })
    return toWorkspaceMessageDispatchResponse(response)
  },

  async streamWorkspaceMessage(
    sessionId: string,
    payload: { content: string },
    scope: 'portal' | 'admin' = 'portal',
    onEvent?: (event: WorkspaceStreamEvent) => void | Promise<void>,
  ) {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/sessions/${sessionId}/messages/stream`
        : `/api/v1/portal/workspace/sessions/${sessionId}/messages/stream`

    const response = await fetch(resolveApiURL(path), {
      method: 'POST',
      credentials: API_REQUEST_CREDENTIALS,
      headers: createRequestHeaders({
        headers: {
          Accept: 'text/event-stream',
          'Content-Type': 'application/json',
        },
      }),
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      const errorPayload = await parseResponseBody(response)
      throw new Error(extractErrorMessage(errorPayload, `请求失败：${response.status}`))
    }

    await readEventStream(response, async (eventName, payloadData) => {
      if (!onEvent) {
        return
      }

      const payloadRecord = payloadData && typeof payloadData === 'object'
        ? (payloadData as Record<string, unknown>)
        : {}

      switch (eventName) {
        case 'message':
          if (payloadRecord.message && typeof payloadRecord.message === 'object') {
            await onEvent({ type: 'message', message: toWorkspaceMessage(payloadRecord.message as Parameters<typeof toWorkspaceMessage>[0]) })
          }
          break
        case 'status':
          if (payloadRecord.dispatch && typeof payloadRecord.dispatch === 'object') {
            const dispatch = payloadRecord.dispatch as Record<string, unknown>
            await onEvent({
              type: 'status',
              dispatch: {
                ok: Boolean(dispatch.ok),
                target: typeof dispatch.target === 'string' ? dispatch.target : undefined,
                error: typeof dispatch.error === 'string' ? dispatch.error : undefined,
                message: typeof dispatch.message === 'string' ? dispatch.message : undefined,
                traceId: typeof dispatch.traceId === 'string' ? dispatch.traceId : undefined,
              },
            })
          }
          break
        case 'chunk':
          await onEvent({
            type: 'chunk',
            delta:
              typeof payloadRecord.delta === 'string'
                ? payloadRecord.delta
                : typeof payloadData === 'string'
                  ? payloadData
                  : '',
          })
          break
        case 'reply':
          if (payloadRecord.message && typeof payloadRecord.message === 'object') {
            await onEvent({ type: 'reply', message: toWorkspaceMessage(payloadRecord.message as Parameters<typeof toWorkspaceMessage>[0]) })
          }
          break
        case 'artifact':
          if (payloadRecord.artifact && typeof payloadRecord.artifact === 'object') {
            await onEvent({ type: 'artifact', artifact: toWorkspaceArtifact(payloadRecord.artifact as Parameters<typeof toWorkspaceArtifact>[0]) })
          }
          break
        case 'error':
          await onEvent({
            type: 'error',
            error:
              typeof payloadRecord.error === 'string'
                ? payloadRecord.error
                : typeof payloadData === 'string'
                  ? payloadData
                  : '流式回复失败',
          })
          break
        case 'done':
          await onEvent({
            type: 'done',
            messageId: typeof payloadRecord.messageId === 'number' ? String(payloadRecord.messageId) : undefined,
          })
          break
      }
    })
  },

  subscribeWorkspaceSessionEvents(
    sessionId: string,
    options: {
      scope?: 'portal' | 'admin'
      afterId?: string
      reconnectDelayMs?: number
      onOpen?: () => void | Promise<void>
      onEvent?: (event: WorkspaceSessionEvent) => void | Promise<void>
      onError?: (error: Error) => void | Promise<void>
      onClose?: () => void | Promise<void>
    } = {},
  ) {
    const scope = options.scope ?? 'portal'
    const path =
      scope === 'admin'
        ? `/api/v1/admin/workspace/sessions/${sessionId}/events`
        : `/api/v1/portal/workspace/sessions/${sessionId}/events`
    const controller = new AbortController()
    let closed = false
    let lastEventId = options.afterId ?? ''

    const run = async () => {
      while (!closed) {
        try {
          const response = await fetch(
            resolveApiURL(appendQuery(path, { after: lastEventId || undefined })),
            {
              method: 'GET',
              credentials: API_REQUEST_CREDENTIALS,
              headers: createRequestHeaders({
                headers: {
                  Accept: 'text/event-stream',
                },
              }),
              signal: controller.signal,
            },
          )

          if (!response.ok) {
            const errorPayload = await parseResponseBody(response)
            throw new Error(extractErrorMessage(errorPayload, `请求失败：${response.status}`))
          }

          await options.onOpen?.()
          await readEventStreamFrames(response, async (frame) => {
            if (!frame.payload || typeof frame.payload !== 'object') {
              return
            }
            if (frame.id) {
              lastEventId = frame.id
            }
            await options.onEvent?.(
              toWorkspaceSessionEvent(frame.payload as Parameters<typeof toWorkspaceSessionEvent>[0]),
            )
          })
          await options.onClose?.()
        } catch (error) {
          if (closed || controller.signal.aborted) {
            break
          }
          await options.onError?.(error instanceof Error ? error : new Error('会话事件流已中断'))
        }

        if (closed || controller.signal.aborted) {
          break
        }

        await new Promise((resolve) => window.setTimeout(resolve, options.reconnectDelayMs ?? 1500))
      }
    }

    void run()

    return {
      close() {
        closed = true
        controller.abort()
      },
      getLastEventId() {
        return lastEventId
      },
    }
  },

  async getWorkspaceBridgeHealth(instanceId: string, scope: 'portal' | 'admin' = 'portal') {
    const path =
      scope === 'admin'
        ? `/api/v1/admin/instances/${instanceId}/workspace/bridge-health`
        : `/api/v1/portal/instances/${instanceId}/workspace/bridge-health`
    return request<{
      bridge: {
        ok: boolean
        target?: string
        status?: number
        error?: string
        message?: string
      }
    }>(path)
  },
}
