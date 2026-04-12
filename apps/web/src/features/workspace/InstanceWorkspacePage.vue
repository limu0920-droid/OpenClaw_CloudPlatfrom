<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { formatAccessEntryType } from '../../lib/access'
import { api } from '../../lib/api'
import { detectArtifactKind, getArtifactLabel, getDraftArtifactPreview } from '../../lib/artifacts'
import { useBranding } from '../../lib/brand'
import { getAdminAccess, getWorkspaceAccess } from '../../lib/workspace'
import {
  isWorkspaceSessionFavorite,
  isWorkspaceSessionPinned,
  loadWorkspaceSessionPreferences,
  sortWorkspaceSessionsByPreference,
  toggleWorkspaceSessionFavorite,
  toggleWorkspaceSessionPinned,
} from '../../lib/workspacePreferences'
import type {
  AdminInstanceDetail,
  PortalInstanceDetail,
  WorkspaceArtifact,
  WorkspaceArtifactPreview,
  WorkspaceMessage,
  WorkspaceSession,
  WorkspaceSessionEvent,
} from '../../lib/types'

type WorkspaceDetail = PortalInstanceDetail | AdminInstanceDetail
type WorkspaceEventStreamHandle = ReturnType<typeof api.subscribeWorkspaceSessionEvents>
type WorkspaceToolItem = {
  key: string
  messageId: string
  name: string
  status: string
  detail: string
  updatedAt: string
  order: number
}

type WorkspaceTraceSummary = {
  traceId: string
  latestAt: string
  order: number
  title: string
  preview: string
  messageCount: number
  artifactCount: number
  toolCount: number
  status: string
}

type WorkspaceTraceTimelineItem = {
  key: string
  traceId: string
  messageId?: string
  eventType: string
  category: string
  title: string
  detail: string
  status: string
  createdAt: string
}

const workspaceMessagePageSize = 40
const workspaceEventWindowSize = 180

const route = useRoute()
const router = useRouter()
const { brand } = useBranding()

const detail = ref<WorkspaceDetail | null>(null)
const loading = ref(true)
const error = ref('')
const frameKey = ref(0)
const artifactUrl = ref('')
const artifactInput = ref('')
const artifactPreviewInput = ref('')
const sessions = ref<WorkspaceSession[]>([])
const selectedSessionId = ref('')
const selectedArtifactId = ref('')
const artifacts = ref<WorkspaceArtifact[]>([])
const messages = ref<WorkspaceMessage[]>([])
const sessionEvents = ref<WorkspaceSessionEvent[]>([])
const sessionsLoading = ref(false)
const sessionSubmitting = ref(false)
const artifactSubmitting = ref(false)
const artifactPreviewLoading = ref(false)
const artifactPreview = ref<WorkspaceArtifactPreview | null>(null)
const artifactPreviewError = ref('')
const messageSubmitting = ref(false)
const messageInput = ref('')
const messageFeedback = ref('')
const sessionFeedback = ref('')
const bridgeHealth = ref<{ ok: boolean; target?: string; status?: number; error?: string; message?: string } | null>(null)
const sessionQuery = ref('')
const sessionStatus = ref('')
const adminGlobalSearch = ref(false)
const newSessionTitle = ref('')
const retryingMessageId = ref('')
const loadingOlderMessages = ref(false)
const streamStatus = ref<'idle' | 'connecting' | 'live' | 'reconnecting'>('idle')
const streamError = ref('')
const streamLastEventId = ref('')
const eventStream = shallowRef<WorkspaceEventStreamHandle | null>(null)
const sessionPreferences = ref(loadWorkspaceSessionPreferences('portal'))
const messagesHasMore = ref(false)
const eventsHasMore = ref(false)
const selectedTraceId = ref('')
const contextMessageId = ref('')
const contextTraceId = ref('')
const messageListRef = ref<HTMLElement | null>(null)
const shouldFollowMessageTail = ref(true)

const scope = computed<'portal' | 'admin'>(() => (route.path.startsWith('/admin') ? 'admin' : 'portal'))
const titlePrefix = computed(() => (scope.value === 'admin' ? '管理员' : '租户'))
const workspaceAccess = computed(() => (detail.value ? getWorkspaceAccess(detail.value.instance.access) : null))
const adminAccess = computed(() => (detail.value ? getAdminAccess(detail.value.instance.access) : null))
const accessEntries = computed(() => detail.value?.instance.access ?? [])
const activeSession = computed(() => sessions.value.find((item) => item.id === selectedSessionId.value) ?? null)
const orderedSessions = computed(() => sortWorkspaceSessionsByPreference(sessions.value, sessionPreferences.value))
const activeSessionPinned = computed(() =>
  activeSession.value ? isWorkspaceSessionPinned(sessionPreferences.value, activeSession.value.id) : false,
)
const activeSessionFavorite = computed(() =>
  activeSession.value ? isWorkspaceSessionFavorite(sessionPreferences.value, activeSession.value.id) : false,
)
const activeSessionArchived = computed(() => activeSession.value?.status === 'archived')
const selectedArtifact = computed(() => artifacts.value.find((item) => item.id === selectedArtifactId.value) ?? null)
const artifactKind = computed(() => detectArtifactKind(artifactUrl.value))
const artifactLabel = computed(() => getArtifactLabel(artifactKind.value))
const workspaceFrameUrl = computed(() => activeSession.value?.workspaceUrl || workspaceAccess.value?.url || '')
const draftArtifactPreview = computed(() =>
  getDraftArtifactPreview(artifactUrl.value, artifactKind.value, artifactPreviewInput.value),
)
const streamStatusText = computed(() => {
  switch (streamStatus.value) {
    case 'connecting':
      return '连接中'
    case 'live':
      return '实时中'
    case 'reconnecting':
      return '重连中'
    default:
      return '未连接'
  }
})
const messageLoadSummary = computed(() => {
  if (!activeSession.value) {
    return ''
  }
  const total = activeSession.value.messageCount ?? messages.value.length
  return `已加载 ${messages.value.length} / ${total} 条消息`
})
const reasoningByMessage = computed<Record<string, string>>(() => {
  const result: Record<string, string> = {}
  for (const event of sessionEvents.value) {
    if (event.eventType !== 'reasoning.delta') continue
    const messageId = event.messageId ?? event.payload.message?.id
    const delta = event.payload.reasoning ?? event.payload.delta
    if (!messageId || !delta) continue
    result[messageId] = `${result[messageId] ?? ''}${delta}`
  }
  return result
})
const toolsByMessage = computed<Record<string, WorkspaceToolItem[]>>(() => {
  const grouped = new Map<string, Map<string, WorkspaceToolItem>>()
  for (const event of sessionEvents.value) {
    if (!event.eventType.startsWith('tool.')) continue
    const messageId = event.messageId ?? event.payload.message?.id
    if (!messageId) continue
    const key = event.payload.toolCallId ?? `${messageId}:${event.payload.toolName ?? 'tool'}`
    const next: WorkspaceToolItem = {
      key,
      messageId,
      name: event.payload.toolName ?? '工具',
      status: event.payload.status ?? event.eventType.replace('tool.', ''),
      detail: event.payload.detail ?? '',
      updatedAt: event.createdAt,
      order: Number(event.id),
    }
    const messageTools = grouped.get(messageId) ?? new Map<string, WorkspaceToolItem>()
    messageTools.set(key, {
      ...(messageTools.get(key) ?? next),
      ...next,
    })
    grouped.set(messageId, messageTools)
  }
  const result: Record<string, WorkspaceToolItem[]> = {}
  grouped.forEach((value, key) => {
    result[key] = Array.from(value.values()).sort((left, right) => left.order - right.order)
  })
  return result
})
const traceSummaries = computed<WorkspaceTraceSummary[]>(() => {
  const traces = new Map<string, WorkspaceTraceSummary>()

  const ensureTrace = (traceId: string) => {
    const existing = traces.get(traceId)
    if (existing) {
      return existing
    }
    const created: WorkspaceTraceSummary = {
      traceId,
      latestAt: '',
      order: 0,
      title: `Trace ${traceId}`,
      preview: '',
      messageCount: 0,
      artifactCount: 0,
      toolCount: 0,
      status: 'active',
    }
    traces.set(traceId, created)
    return created
  }

  for (const message of messages.value) {
    if (!message.traceId) continue
    const trace = ensureTrace(message.traceId)
    trace.messageCount += 1
    if (Number(message.id) >= trace.order) {
      trace.order = Number(message.id)
      trace.latestAt = message.createdAt
    }
    if (!trace.preview && message.content.trim()) {
      trace.preview = message.content.trim()
    }
    if (message.role === 'assistant' && !trace.title.startsWith('龙虾')) {
      trace.title = `龙虾回复 · ${message.traceId}`
    }
    if (message.status === 'failed') {
      trace.status = 'failed'
    } else if (message.status === 'streaming' && trace.status !== 'failed') {
      trace.status = 'streaming'
    }
  }

  for (const event of sessionEvents.value) {
    const traceId = event.traceId || event.payload.message?.traceId
    if (!traceId) continue
    const trace = ensureTrace(traceId)
    if (Number(event.id) >= trace.order) {
      trace.order = Number(event.id)
      trace.latestAt = event.createdAt
    }
    if (event.eventType.startsWith('tool.')) {
      trace.toolCount += 1
    }
    if (event.eventType === 'artifact.created' || event.payload.artifact) {
      trace.artifactCount += 1
    }
    if (!trace.preview) {
      trace.preview =
        event.payload.detail ||
        event.payload.content ||
        event.payload.reasoning ||
        event.payload.artifact?.title ||
        event.payload.message?.content ||
        ''
    }
    if (event.eventType === 'message.failed' || event.payload.error || event.payload.errorMessage) {
      trace.status = 'failed'
    } else if ((event.eventType === 'message.completed' || event.eventType === 'tool.completed') && trace.status !== 'failed') {
      trace.status = 'completed'
    } else if (trace.status !== 'failed' && trace.status !== 'completed') {
      trace.status = event.payload.status || trace.status
    }
  }

  return Array.from(traces.values())
    .map((item) => ({
      ...item,
      preview: item.preview.length > 88 ? `${item.preview.slice(0, 88)}…` : item.preview,
    }))
    .sort((left, right) => right.order - left.order)
})
const activeTrace = computed(() => traceSummaries.value.find((item) => item.traceId === selectedTraceId.value) ?? null)
const activeTraceTimeline = computed<WorkspaceTraceTimelineItem[]>(() => {
  if (!selectedTraceId.value) {
    return []
  }

  const items = sessionEvents.value
    .filter((event) => (event.traceId || event.payload.message?.traceId) === selectedTraceId.value)
    .map((event) => ({
      key: event.id,
      traceId: selectedTraceId.value,
      messageId: event.messageId,
      eventType: event.eventType,
      category: getWorkspaceEventCategory(event.eventType),
      title: getWorkspaceEventTitle(event),
      detail: getWorkspaceEventDetail(event),
      status: event.payload.status || inferWorkspaceEventStatus(event),
      createdAt: event.createdAt,
    }))

  return items.sort((left, right) => Number(left.key) - Number(right.key))
})
const traceWindowNotice = computed(() =>
  eventsHasMore.value ? `当前只展示最近 ${sessionEvents.value.length} 个事件，历史会话请继续向上翻阅消息。` : '',
)

function currentWorkspaceQuery(overrides: Record<string, string | undefined> = {}) {
  return {
    artifact: typeof route.query.artifact === 'string' ? route.query.artifact : undefined,
    preview: typeof route.query.preview === 'string' ? route.query.preview : undefined,
    q: sessionQuery.value.trim() || undefined,
    status: sessionStatus.value || undefined,
    all: scope.value === 'admin' && adminGlobalSearch.value ? '1' : undefined,
    sessionId: selectedSessionId.value || undefined,
    messageId: contextMessageId.value || undefined,
    traceId: contextTraceId.value || undefined,
    ...overrides,
  }
}

async function syncWorkspaceQuery(overrides: Record<string, string | undefined> = {}) {
  await router.replace({ query: currentWorkspaceQuery(overrides) })
}

function eventCursorStorageKey(sessionId: string) {
  return `openclaw.workspace.events.${scope.value}.${sessionId}`
}

function readStoredEventCursor(sessionId: string) {
  if (typeof window === 'undefined') return ''
  try {
    return window.sessionStorage.getItem(eventCursorStorageKey(sessionId)) ?? ''
  } catch {
    return ''
  }
}

function writeStoredEventCursor(sessionId: string, eventId: string) {
  if (!eventId || typeof window === 'undefined') return
  try {
    window.sessionStorage.setItem(eventCursorStorageKey(sessionId), eventId)
  } catch {
    // ignore storage failures
  }
}

function sortMessages(items: WorkspaceMessage[]) {
  return [...items].sort((left, right) => Number(left.id) - Number(right.id))
}

function sortArtifacts(items: WorkspaceArtifact[]) {
  return [...items].sort((left, right) => Number(right.id) - Number(left.id))
}

function sortEvents(items: WorkspaceSessionEvent[]) {
  return [...items].sort((left, right) => Number(left.id) - Number(right.id))
}

function replaceSession(nextSession: WorkspaceSession) {
  const next = [...sessions.value]
  const index = next.findIndex((item) => item.id === nextSession.id)
  if (index >= 0) {
    next[index] = nextSession
  } else {
    next.unshift(nextSession)
  }
  sessions.value = next
}

function refreshSessionPreferences() {
  sessionPreferences.value = loadWorkspaceSessionPreferences(scope.value)
}

function sessionFlagLabels(session: WorkspaceSession) {
  return [
    isWorkspaceSessionPinned(sessionPreferences.value, session.id) ? '置顶' : '',
    isWorkspaceSessionFavorite(sessionPreferences.value, session.id) ? '收藏' : '',
    session.status === 'archived' ? '归档' : '',
  ].filter(Boolean)
}

function closeEventStream() {
  eventStream.value?.close()
  eventStream.value = null
}

function getMessageRoleLabel(role: string) {
  switch (role) {
    case 'assistant':
      return '龙虾'
    case 'system':
      return '系统'
    case 'note':
      return '备注'
    default:
      return '用户'
  }
}

function getMessageStatusLabel(status: string) {
  switch (status) {
    case 'streaming':
      return '流式中'
    case 'delivered':
      return '已完成'
    case 'failed':
      return '失败'
    case 'sent':
      return '已发送'
    default:
      return '已记录'
  }
}

function getToolStatusLabel(status: string) {
  switch (status) {
    case 'started':
      return '已启动'
    case 'completed':
    case 'success':
    case 'succeeded':
      return '已完成'
    case 'failed':
    case 'error':
      return '失败'
    default:
      return '进行中'
  }
}

function inferWorkspaceEventStatus(event: WorkspaceSessionEvent) {
  if (event.eventType === 'message.failed') return 'failed'
  if (event.eventType === 'message.completed') return 'completed'
  if (event.eventType === 'tool.completed') return 'completed'
  if (event.eventType === 'tool.failed') return 'failed'
  if (event.eventType === 'tool.started') return 'started'
  if (event.eventType === 'reasoning.delta') return 'streaming'
  return 'active'
}

function getWorkspaceEventCategory(eventType: string) {
  if (eventType.startsWith('tool.')) return 'tool'
  if (eventType.startsWith('message.')) return 'message'
  if (eventType.startsWith('reasoning.')) return 'reasoning'
  if (eventType.startsWith('artifact.')) return 'artifact'
  if (eventType.startsWith('dispatch.')) return 'dispatch'
  return 'event'
}

function getWorkspaceEventTitle(event: WorkspaceSessionEvent) {
  if (event.eventType.startsWith('tool.')) {
    return `${event.payload.toolName || '工具'} · ${getToolStatusLabel(event.payload.status || event.eventType.replace('tool.', ''))}`
  }
  if (event.eventType === 'reasoning.delta') {
    return '推理增量'
  }
  if (event.eventType === 'artifact.created') {
    return `产物创建 · ${event.payload.artifact?.title || '未命名产物'}`
  }
  if (event.eventType === 'dispatch.status') {
    return '桥接派发状态'
  }
  if (event.eventType.startsWith('message.')) {
    return `消息事件 · ${event.payload.message ? getMessageRoleLabel(event.payload.message.role) : '会话'}`
  }
  return event.eventType
}

function getWorkspaceEventDetail(event: WorkspaceSessionEvent) {
  return (
    event.payload.detail ||
    event.payload.errorMessage ||
    event.payload.error ||
    event.payload.reasoning ||
    event.payload.delta ||
    event.payload.content ||
    event.payload.artifact?.title ||
    event.payload.message?.content ||
    ''
  )
}

async function copyText(value: string, successText: string, failureText: string) {
  if (!value.trim()) return
  try {
    await navigator.clipboard.writeText(value)
    messageFeedback.value = successText
  } catch {
    messageFeedback.value = failureText
  }
}

function rememberMessageListFollowState() {
  const list = messageListRef.value
  if (!list) return
  shouldFollowMessageTail.value = list.scrollHeight - list.clientHeight - list.scrollTop < 96
}

async function scrollMessagesToBottom() {
  await nextTick()
  const list = messageListRef.value
  if (!list) return
  list.scrollTop = list.scrollHeight
}

async function scrollMessageIntoView(messageId?: string) {
  if (!messageId) return
  await nextTick()
  const target = document.querySelector<HTMLElement>(`[data-message-id="${messageId}"]`)
  target?.scrollIntoView({ block: 'center', behavior: 'smooth' })
}

function ensureTraceSelection() {
  if (!traceSummaries.value.length) {
    selectedTraceId.value = ''
    return
  }
  if (!traceSummaries.value.some((item) => item.traceId === selectedTraceId.value)) {
    selectedTraceId.value = traceSummaries.value[0].traceId
  }
}

async function applyContextAnchor() {
  if (contextTraceId.value) {
    const matchedMessage = messages.value.find((item) => item.traceId === contextTraceId.value)
    if (matchedMessage && !contextMessageId.value) {
      contextMessageId.value = matchedMessage.id
    }
    selectedTraceId.value = contextTraceId.value
  } else {
    ensureTraceSelection()
  }

  if (contextMessageId.value && messages.value.some((item) => item.id === contextMessageId.value)) {
    await scrollMessageIntoView(contextMessageId.value)
  }
}

async function load(id: string) {
  loading.value = true
  error.value = ''
  sessionFeedback.value = ''
  refreshSessionPreferences()
  closeEventStream()
  streamStatus.value = 'idle'
  streamError.value = ''
  streamLastEventId.value = ''
  messagesHasMore.value = false
  eventsHasMore.value = false
  selectedTraceId.value = ''
  sessionQuery.value = typeof route.query.q === 'string' ? route.query.q : ''
  sessionStatus.value = typeof route.query.status === 'string' ? route.query.status : ''
  adminGlobalSearch.value = scope.value === 'admin' && route.query.all === '1'
  selectedSessionId.value = typeof route.query.sessionId === 'string' ? route.query.sessionId : ''
  contextMessageId.value = typeof route.query.messageId === 'string' ? route.query.messageId : ''
  contextTraceId.value = typeof route.query.traceId === 'string' ? route.query.traceId : ''

  try {
    detail.value =
      scope.value === 'admin' ? await api.getAdminInstanceDetail(id) : await api.getPortalInstanceDetail(id)
    await loadSessions(id)
    const bridge = await api.getWorkspaceBridgeHealth(id, scope.value)
    bridgeHealth.value = bridge.bridge
    const artifact = route.query.artifact
    if (typeof artifact === 'string' && artifact.trim()) {
      artifactUrl.value = artifact
      artifactInput.value = artifact
      selectedArtifactId.value = ''
      artifactPreview.value = null
      artifactPreviewError.value = ''
    }
    const preview = route.query.preview
    if (typeof preview === 'string' && preview.trim()) {
      artifactPreviewInput.value = preview
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载工作台失败'
  } finally {
    loading.value = false
  }
}

function refreshFrame() {
  if (!workspaceFrameUrl.value) return
  frameKey.value += 1
}

function openExternal() {
  if (!workspaceFrameUrl.value) return
  window.open(workspaceFrameUrl.value, '_blank', 'noopener,noreferrer')
}

function openAdmin() {
  if (!adminAccess.value?.url) return
  window.open(adminAccess.value.url, '_blank', 'noopener,noreferrer')
}

async function loadSessions(instanceId: string) {
  sessionsLoading.value = true
  try {
    sessions.value =
      scope.value === 'admin' && adminGlobalSearch.value
        ? await api.searchWorkspaceSessions('admin', {
            q: sessionQuery.value.trim() || undefined,
            status: sessionStatus.value || undefined,
            limit: 100,
          })
        : await api.getWorkspaceSessions(instanceId, scope.value, {
            q: sessionQuery.value.trim() || undefined,
            status: sessionStatus.value || undefined,
            limit: 100,
          })

    const targetSessionId = (typeof route.query.sessionId === 'string' && route.query.sessionId) || selectedSessionId.value
    if (targetSessionId) {
      const matched = sessions.value.find((item) => item.id === targetSessionId)
      if (matched) {
        if (scope.value === 'admin' && adminGlobalSearch.value && matched.instanceId !== String(instanceId)) {
          await router.replace({
            path: `/admin/instances/${matched.instanceId}/workspace`,
            query: currentWorkspaceQuery({ sessionId: matched.id, all: '1' }),
          })
          return
        }
        await loadSessionDetail(matched.id)
        return
      }
    }

    if (sessions.value.length > 0) {
      await loadSessionDetail(sessions.value[0].id)
      await syncWorkspaceQuery({ sessionId: sessions.value[0].id })
    } else {
      closeEventStream()
      selectedSessionId.value = ''
      artifacts.value = []
      messages.value = []
      sessionEvents.value = []
      messagesHasMore.value = false
      eventsHasMore.value = false
      selectedTraceId.value = ''
      streamStatus.value = 'idle'
      streamError.value = ''
      streamLastEventId.value = ''
    }
  } finally {
    sessionsLoading.value = false
  }
}

async function loadSessionDetail(sessionId: string) {
  const response = await api.getWorkspaceSessionDetail(sessionId, scope.value, {
    messageLimit: workspaceMessagePageSize,
    eventLimit: workspaceEventWindowSize,
    anchorMessageId: contextMessageId.value || undefined,
    anchorTraceId: contextTraceId.value || undefined,
  })
  applySessionDetail(response)
  await applyContextAnchor()
  if (selectedArtifactId.value) {
    const currentArtifact = response.artifacts.find((item) => item.id === selectedArtifactId.value)
    if (currentArtifact) {
      await loadArtifactPreview(currentArtifact.id)
      return
    }
    selectedArtifactId.value = ''
    artifactPreview.value = null
    artifactPreviewError.value = ''
  }
  if (!artifactUrl.value && response.artifacts.length > 0) {
    await openSavedArtifact(response.artifacts[0])
  }
}

async function loadOlderMessages() {
  if (!selectedSessionId.value || !messages.value.length || loadingOlderMessages.value) {
    return
  }

  const list = messageListRef.value
  const previousHeight = list?.scrollHeight ?? 0
  loadingOlderMessages.value = true
  try {
    const response = await api.getWorkspaceMessages(selectedSessionId.value, scope.value, {
      beforeId: messages.value[0]?.id,
      limit: workspaceMessagePageSize,
    })
    if (response.items.length) {
      const next = new Map<string, WorkspaceMessage>()
      for (const item of [...response.items, ...messages.value]) {
        next.set(item.id, item)
      }
      messages.value = sortMessages(Array.from(next.values()))
    }
    messagesHasMore.value = response.hasMore
    await nextTick()
    if (list) {
      list.scrollTop += list.scrollHeight - previousHeight
    }
  } catch (err) {
    messageFeedback.value = err instanceof Error ? err.message : '加载更早消息失败'
  } finally {
    loadingOlderMessages.value = false
  }
}

async function applySessionFilters() {
  if (!detail.value) return
  selectedSessionId.value = ''
  contextMessageId.value = ''
  contextTraceId.value = ''
  await syncWorkspaceQuery({ sessionId: undefined, messageId: undefined, traceId: undefined, q: sessionQuery.value.trim() || undefined, status: sessionStatus.value || undefined, all: adminGlobalSearch.value ? '1' : undefined })
  await loadSessions(detail.value.instance.id)
}

async function resetSessionFilters() {
  sessionQuery.value = ''
  sessionStatus.value = ''
  adminGlobalSearch.value = false
  await applySessionFilters()
}

async function openSession(session: WorkspaceSession) {
  if (scope.value === 'admin' && adminGlobalSearch.value && detail.value && session.instanceId !== detail.value.instance.id) {
    contextMessageId.value = ''
    contextTraceId.value = ''
    await router.push({
      path: `/admin/instances/${session.instanceId}/workspace`,
      query: currentWorkspaceQuery({ sessionId: session.id, messageId: undefined, traceId: undefined, all: '1' }),
    })
    return
  }
  contextMessageId.value = ''
  contextTraceId.value = ''
  await loadSessionDetail(session.id)
  await syncWorkspaceQuery({ sessionId: session.id, messageId: undefined, traceId: undefined })
}

async function createSession() {
  if (!detail.value) return
  sessionSubmitting.value = true
  try {
    const session = await api.createWorkspaceSession(
      detail.value.instance.id,
      {
        title: newSessionTitle.value.trim() || undefined,
        workspaceUrl: workspaceFrameUrl.value || undefined,
      },
      scope.value,
    )
    newSessionTitle.value = ''
    sessions.value = [session, ...sessions.value]
    contextMessageId.value = ''
    contextTraceId.value = ''
    await loadSessionDetail(session.id)
    await syncWorkspaceQuery({ sessionId: session.id, messageId: undefined, traceId: undefined })
  } catch (err) {
    error.value = err instanceof Error ? err.message : '创建会话失败'
  } finally {
    sessionSubmitting.value = false
  }
}

function toggleCurrentSessionPinned() {
  if (!activeSession.value) return
  const wasPinned = activeSessionPinned.value
  sessionPreferences.value = toggleWorkspaceSessionPinned(scope.value, activeSession.value.id)
  sessionFeedback.value = wasPinned ? '当前会话已取消置顶' : '当前会话已置顶'
}

function toggleCurrentSessionFavorite() {
  if (!activeSession.value) return
  const wasFavorite = activeSessionFavorite.value
  sessionPreferences.value = toggleWorkspaceSessionFavorite(scope.value, activeSession.value.id)
  sessionFeedback.value = wasFavorite ? '当前会话已取消收藏' : '当前会话已加入收藏'
}

async function toggleCurrentSessionArchived() {
  if (!activeSession.value) return

  const targetStatus = activeSessionArchived.value ? 'active' : 'archived'
  sessionFeedback.value = ''
  try {
    const updatedSession = await api.updateWorkspaceSessionStatus(activeSession.value.id, { status: targetStatus }, scope.value)
    replaceSession(updatedSession)
    sessionFeedback.value =
      targetStatus === 'archived'
        ? '当前会话已归档，消息发送与产物保存已锁定'
        : '当前会话已恢复为可写状态'
  } catch (err) {
    sessionFeedback.value = err instanceof Error ? err.message : '更新会话状态失败'
  }
}

function previewArtifact() {
  artifactUrl.value = artifactInput.value.trim()
  selectedArtifactId.value = ''
  artifactPreview.value = null
  artifactPreviewError.value = ''
}

async function copyMessage(message: WorkspaceMessage) {
  await copyText(message.content, '消息内容已复制到剪贴板', '复制失败，请检查浏览器剪贴板权限')
}

function focusTrace(traceId?: string) {
  if (!traceId) return
  selectedTraceId.value = traceId
  contextTraceId.value = traceId
  void syncWorkspaceQuery({ traceId, messageId: contextMessageId.value || undefined })
}

async function copyTraceId(traceId?: string) {
  if (!traceId) return
  await copyText(traceId, 'Trace ID 已复制到剪贴板', '复制 Trace ID 失败，请检查浏览器剪贴板权限')
}

void messageLoadSummary
void activeTrace
void traceWindowNotice
void rememberMessageListFollowState
void loadOlderMessages
void focusTrace
void copyTraceId

async function retryMessage(message: WorkspaceMessage) {
  if (activeSessionArchived.value) {
    messageFeedback.value = '当前会话已归档，不能重新派发消息'
    return
  }

  retryingMessageId.value = message.id
  messageFeedback.value = ''
  shouldFollowMessageTail.value = true
  try {
    const response = await api.retryWorkspaceMessage(message.id, scope.value)
    upsertMessage(response.message)
    if (response.reply) {
      upsertMessage(response.reply)
    }
    response.artifacts.forEach(prependArtifact)
    messageFeedback.value = response.dispatch.ok
      ? response.reply?.status === 'streaming'
        ? '消息已重新派发，正在接收流式回复'
        : '消息已重新派发到龙虾实例'
      : response.dispatch.error || '消息已重试记录，但重新派发失败'
    if (detail.value) {
      await loadSessions(detail.value.instance.id)
    }
  } catch (err) {
    messageFeedback.value = err instanceof Error ? err.message : '重新派发失败'
  } finally {
    retryingMessageId.value = ''
  }
}

function openArtifactExternal() {
  const target = artifactPreview.value?.downloadUrl || artifactPreview.value?.externalUrl || artifactUrl.value
  if (!target) return
  window.open(target, '_blank', 'noopener,noreferrer')
}

function openFormalPreview() {
  const target = artifactPreview.value?.previewUrl || artifactPreview.value?.downloadUrl
  if (!target) return
  window.open(target, '_blank', 'noopener,noreferrer')
}

function downloadArtifact() {
  if (!artifactPreview.value?.downloadUrl) return
  window.open(artifactPreview.value.downloadUrl, '_blank', 'noopener,noreferrer')
}

async function loadArtifactPreview(artifactId: string) {
  artifactPreviewLoading.value = true
  artifactPreviewError.value = ''
  try {
    artifactPreview.value = await api.getWorkspaceArtifactPreview(artifactId, scope.value)
  } catch (err) {
    artifactPreview.value = null
    artifactPreviewError.value = err instanceof Error ? err.message : '加载正式预览失败'
  } finally {
    artifactPreviewLoading.value = false
  }
}

async function saveArtifact() {
  if (!artifactUrl.value || !activeSession.value || activeSessionArchived.value) return
  artifactSubmitting.value = true
  try {
    const artifact = await api.createWorkspaceArtifact(
      activeSession.value.id,
      {
        title: `${artifactLabel.value} · ${new Date().toLocaleString('zh-CN')}`,
        kind: artifactKind.value,
        sourceUrl: artifactUrl.value,
        previewUrl: artifactPreviewInput.value.trim() || undefined,
      },
      scope.value,
    )
    artifacts.value = [artifact, ...artifacts.value]
    await openSavedArtifact(artifact)
    await loadSessions(detail.value?.instance.id || activeSession.value.instanceId)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存产物失败'
  } finally {
    artifactSubmitting.value = false
  }
}

async function openSavedArtifact(item: WorkspaceArtifact) {
  artifactUrl.value = item.sourceUrl
  artifactInput.value = item.sourceUrl
  artifactPreviewInput.value = item.previewUrl ?? ''
  selectedArtifactId.value = item.id
  await loadArtifactPreview(item.id)
}

function upsertMessage(message: WorkspaceMessage) {
  const next = [...messages.value]
  const index = next.findIndex((item) => item.id === message.id)
  if (index >= 0) {
    next[index] = message
  } else {
    next.push(message)
  }
  messages.value = sortMessages(next)
  ensureTraceSelection()
  if (shouldFollowMessageTail.value) {
    void scrollMessagesToBottom()
  }
}

function prependArtifact(artifact: WorkspaceArtifact) {
  const next = [...artifacts.value]
  const index = next.findIndex((item) => item.id === artifact.id)
  if (index >= 0) {
    next[index] = artifact
  } else {
    next.unshift(artifact)
  }
  artifacts.value = sortArtifacts(next)
}

function upsertSessionEvent(event: WorkspaceSessionEvent) {
  if (event.sessionId !== selectedSessionId.value) return
  if (!sessionEvents.value.some((item) => item.id === event.id)) {
    sessionEvents.value = sortEvents([...sessionEvents.value, event])
  }
  if (event.payload.message) {
    upsertMessage(event.payload.message)
  }
  if (event.payload.artifact) {
    prependArtifact(event.payload.artifact)
  }
  if (event.id) {
    streamLastEventId.value = event.id
    writeStoredEventCursor(event.sessionId, event.id)
  }
  if (event.eventType === 'dispatch.status' && event.payload.dispatch) {
    messageFeedback.value = event.payload.dispatch.ok
      ? '龙虾桥接已接受消息，正在持续回传实时事件'
      : event.payload.dispatch.error || '龙虾桥接发送失败'
  }
  if (event.eventType === 'message.failed') {
    messageFeedback.value = event.payload.errorMessage || event.payload.error || '龙虾处理失败'
  }
  ensureTraceSelection()
}

function connectEventStream(sessionId: string, afterId: string) {
  closeEventStream()
  streamStatus.value = afterId ? 'reconnecting' : 'connecting'
  streamError.value = ''
  eventStream.value = api.subscribeWorkspaceSessionEvents(sessionId, {
    scope: scope.value,
    afterId,
    onOpen: () => {
      streamStatus.value = 'live'
      streamError.value = ''
    },
    onClose: () => {
      if (selectedSessionId.value === sessionId) {
        streamStatus.value = 'reconnecting'
      }
    },
    onError: (err) => {
      if (selectedSessionId.value === sessionId) {
        streamStatus.value = 'reconnecting'
        streamError.value = err.message
      }
    },
    onEvent: (event) => {
      upsertSessionEvent(event)
    },
  })
}

function applySessionDetail(response: Awaited<ReturnType<typeof api.getWorkspaceSessionDetail>>) {
  selectedSessionId.value = response.session.id
  artifacts.value = sortArtifacts(response.artifacts)
  messages.value = sortMessages(response.messages)
  sessionEvents.value = sortEvents(response.events)
  messagesHasMore.value = response.messagesHasMore
  eventsHasMore.value = response.eventsHasMore
  shouldFollowMessageTail.value = true
  ensureTraceSelection()

  const responseCursor = response.events.at(-1)?.id ?? ''
  const storedCursor = readStoredEventCursor(response.session.id)
  const numericCursor = Math.max(Number(responseCursor || 0), Number(storedCursor || 0))
  streamLastEventId.value = numericCursor > 0 ? String(numericCursor) : ''
  if (streamLastEventId.value) {
    writeStoredEventCursor(response.session.id, streamLastEventId.value)
  }
  connectEventStream(response.session.id, streamLastEventId.value)
  void scrollMessagesToBottom()
}

async function sendMessageRecord() {
  if (!activeSession.value || activeSessionArchived.value || !messageInput.value.trim()) return
  const sessionId = activeSession.value.id
  const content = messageInput.value.trim()
  messageSubmitting.value = true
  messageFeedback.value = ''
  shouldFollowMessageTail.value = true

  try {
    const response = await api.createWorkspaceMessage(
      sessionId,
      {
        role: 'user',
        status: 'recorded',
        content,
        dispatch: true,
      },
      scope.value,
    )
    upsertMessage(response.message)
    if (response.reply) {
      upsertMessage(response.reply)
    }
    response.artifacts.forEach(prependArtifact)
    messageFeedback.value = response.dispatch.ok
      ? response.reply?.status === 'streaming'
        ? '消息已发送，正在接收龙虾流式回复'
        : '消息已发送到龙虾实例'
      : response.dispatch.error || '消息已记录，但发送到龙虾失败'
    messageInput.value = ''
    if (detail.value) {
      await loadSessions(detail.value.instance.id)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '记录消息失败'
  } finally {
    messageSubmitting.value = false
  }
}

watch(
  () => [String(route.params.id), scope.value, String(route.query.sessionId ?? ''), String(route.query.messageId ?? ''), String(route.query.traceId ?? ''), String(route.query.q ?? ''), String(route.query.status ?? ''), String(route.query.all ?? '')] as const,
  ([id]) => {
    void load(id)
  },
  { immediate: true },
)

watch(
  () => traceSummaries.value.map((item) => item.traceId).join(','),
  () => {
    ensureTraceSelection()
  },
)

onBeforeUnmount(() => {
  closeEventStream()
})
</script>

<template>
  <div class="workspace-shell">
    <el-card v-if="loading" shadow="never" class="state-card">正在打开网页版工作台…</el-card>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
    <template v-else-if="detail">
      <el-card shadow="never" class="workspace-hero">
        <div class="hero-copy">
          <div class="eyebrow">{{ titlePrefix }} Workspace</div>
          <h2>{{ detail.instance.name }}</h2>
          <p class="muted">通过 {{ brand.name }} Web 门户直接进入工作台，对话并查看它输出的网页、文档、PPTX、PDF 等结果。</p>
        </div>
        <div class="hero-metas">
          <span class="hero-chip">版本 {{ detail.instance.version }}</span>
          <span class="hero-chip">区域 {{ detail.instance.region }}</span>
          <span class="hero-chip">套餐 {{ detail.instance.plan }}</span>
          <span class="hero-chip">状态 {{ detail.instance.status }}</span>
        </div>
        <div class="hero-actions">
          <el-button type="primary" :disabled="!workspaceFrameUrl" @click="refreshFrame">刷新工作台</el-button>
          <el-button plain :disabled="!workspaceFrameUrl" @click="openExternal">新窗口打开</el-button>
          <el-button plain :disabled="!adminAccess" @click="openAdmin">打开后台入口</el-button>
          <el-button plain @click="router.back()">返回上一页</el-button>
        </div>
      </el-card>

      <section class="workspace-layout">
        <el-card shadow="never" class="workspace-column column-sessions">
          <SectionHeader title="会话" subtitle="平台侧工作台会话与桥接状态" />
          <el-alert
            v-if="bridgeHealth"
            :closable="false"
            show-icon
            :type="bridgeHealth.ok ? 'success' : 'warning'"
            :title="bridgeHealth.ok ? '龙虾桥接已就绪' : `龙虾桥接不可用：${bridgeHealth.error || '未返回详情'}`"
          />
          <div class="session-filters">
            <el-input
              v-model="sessionQuery"
              clearable
              placeholder="搜索会话标题、实例、消息摘要"
              @keyup.enter="applySessionFilters"
            />
            <el-select v-model="sessionStatus" placeholder="状态">
              <el-option label="全部状态" value="" />
              <el-option label="Active" value="active" />
              <el-option label="Archived" value="archived" />
            </el-select>
            <el-switch
              v-if="scope === 'admin'"
              v-model="adminGlobalSearch"
              active-text="全局检索"
              @change="applySessionFilters"
            />
            <div class="artifact-actions">
              <el-button plain @click="applySessionFilters">检索</el-button>
              <el-button text @click="resetSessionFilters">清空</el-button>
            </div>
          </div>
          <div class="session-actions">
            <el-input v-model="newSessionTitle" clearable placeholder="新会话标题，可选" />
            <el-button type="primary" :loading="sessionSubmitting" @click="createSession">新建会话</el-button>
          </div>
          <div v-if="activeSession" class="session-current-actions">
            <el-button plain size="small" @click="toggleCurrentSessionPinned">
              {{ activeSessionPinned ? '取消置顶' : '置顶当前会话' }}
            </el-button>
            <el-button plain size="small" @click="toggleCurrentSessionFavorite">
              {{ activeSessionFavorite ? '取消收藏' : '收藏当前会话' }}
            </el-button>
            <el-button plain size="small" type="warning" @click="toggleCurrentSessionArchived">
              {{ activeSessionArchived ? '恢复会话' : '归档会话' }}
            </el-button>
          </div>
          <el-alert v-if="sessionFeedback" :closable="false" show-icon type="info" :title="sessionFeedback" />
          <div v-if="sessionsLoading" class="muted">正在加载会话…</div>
          <div v-else-if="orderedSessions.length" class="session-list">
            <button
              v-for="session in orderedSessions"
              :key="session.id"
              type="button"
              :class="['session-item', { active: selectedSessionId === session.id }]"
              @click="openSession(session)"
            >
              <strong>{{ session.title }}</strong>
              <span class="muted">#{{ session.sessionNo }} · {{ session.updatedAt }}</span>
              <div v-if="sessionFlagLabels(session).length" class="session-flags">
                <span v-for="label in sessionFlagLabels(session)" :key="label" class="session-flag">{{ label }}</span>
              </div>
              <small class="muted">
                {{ session.messageCount ?? 0 }} 条消息 · {{ session.artifactCount ?? 0 }} 个产物
                <template v-if="scope === 'admin' && (session.tenantName || session.instanceName)">
                  · {{ session.tenantName || '—' }} / {{ session.instanceName || '—' }}
                </template>
              </small>
              <small v-if="session.lastMessagePreview" class="muted">{{ session.lastMessagePreview }}</small>
            </button>
          </div>
          <div v-else class="muted">当前实例还没有平台侧会话记录，先创建一个会话。</div>

          <SectionHeader title="访问入口" subtitle="当前实例可用的 Web 入口" />
          <div class="access-list">
            <div v-for="entry in accessEntries" :key="entry.url" class="access-item">
              <strong>{{ formatAccessEntryType(entry.entryType) }}</strong>
              <span class="muted">{{ entry.url }}</span>
              <small class="muted">mode: {{ entry.accessMode || '—' }}</small>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="workspace-column column-chat">
          <SectionHeader
            title="对话"
            :subtitle="activeSession ? `平台侧直连会话 · ${activeSession.sessionNo}` : '平台侧直连会话，流式展示回复与状态'"
          />
          <div class="message-panel">
            <div class="realtime-status">
              <span class="hero-chip">链路 {{ streamStatusText }}</span>
              <span v-if="activeSession?.protocolVersion" class="hero-chip">{{ activeSession.protocolVersion }}</span>
              <span v-if="streamLastEventId" class="hero-chip">游标 #{{ streamLastEventId }}</span>
              <span v-if="messageLoadSummary" class="hero-chip">{{ messageLoadSummary }}</span>
            </div>
            <el-alert
              v-if="activeSessionArchived"
              :closable="false"
              show-icon
              type="warning"
              title="当前会话已归档，消息发送与重新派发已锁定。"
            />
            <el-alert v-if="streamError" :closable="false" show-icon type="warning" :title="streamError" />
            <div v-if="messages.length" ref="messageListRef" class="message-list" @scroll="rememberMessageListFollowState">
              <div class="message-list-toolbar">
                <el-button
                  v-if="messagesHasMore"
                  text
                  size="small"
                  :loading="loadingOlderMessages"
                  @click="loadOlderMessages"
                >
                  加载更早的 {{ workspaceMessagePageSize }} 条消息
                </el-button>
                <span v-else class="muted">已经到达当前会话最早的已加载消息。</span>
              </div>
              <div
                v-for="message in messages"
                :key="message.id"
                :data-message-id="message.id"
                :class="[
                  'message-item',
                  `role-${message.role}`,
                  {
                    'message-item--trace': message.traceId && message.traceId === selectedTraceId,
                    'message-item--anchored': contextMessageId && message.id === contextMessageId,
                  },
                ]"
              >
                <div class="message-head">
                  <strong>{{ getMessageRoleLabel(message.role) }}</strong>
                  <div class="message-meta">
                    <span :class="['message-status', `status-${message.status}`]">{{ getMessageStatusLabel(message.status) }}</span>
                    <span class="muted">{{ message.createdAt }}</span>
                  </div>
                </div>
                <div v-if="message.traceId" class="message-trace-row">
                  <button type="button" class="trace-chip" @click="focusTrace(message.traceId)">
                    Trace {{ message.traceId }}
                  </button>
                  <el-button text size="small" @click="copyTraceId(message.traceId)">复制 Trace</el-button>
                </div>
                <p>{{ message.content || (message.status === 'streaming' ? '正在生成…' : '—') }}</p>
                <div v-if="message.errorMessage" class="message-error">{{ message.errorMessage }}</div>
                <div v-if="reasoningByMessage[message.id]" class="message-reasoning">
                  <strong>推理过程</strong>
                  <p>{{ reasoningByMessage[message.id] }}</p>
                </div>
                <div v-if="toolsByMessage[message.id]?.length" class="message-tools">
                  <div v-for="tool in toolsByMessage[message.id]" :key="tool.key" class="tool-item">
                    <div class="tool-head">
                      <strong>{{ tool.name }}</strong>
                      <span class="muted">{{ getToolStatusLabel(tool.status) }} · {{ tool.updatedAt }}</span>
                    </div>
                    <p v-if="tool.detail">{{ tool.detail }}</p>
                  </div>
                </div>
                <div class="message-item-actions">
                  <el-button text size="small" :disabled="!message.content.trim()" @click="copyMessage(message)">复制</el-button>
                  <el-button
                    v-if="message.role === 'user'"
                    text
                    size="small"
                    :loading="retryingMessageId === message.id"
                    :disabled="activeSessionArchived"
                    @click="retryMessage(message)"
                  >
                    重新派发
                  </el-button>
                </div>
              </div>
            </div>
            <div v-else class="muted">当前会话还没有平台侧消息记录。</div>
            <div v-if="traceSummaries.length" class="trace-panel">
              <SectionHeader title="Trace 视图" subtitle="按 trace 汇总工具、推理、产物和桥接状态" />
              <el-alert v-if="traceWindowNotice" :closable="false" show-icon type="info" :title="traceWindowNotice" />
              <div class="trace-summary-list">
                <button
                  v-for="trace in traceSummaries"
                  :key="trace.traceId"
                  type="button"
                  :class="['trace-summary', { active: selectedTraceId === trace.traceId }]"
                  @click="focusTrace(trace.traceId)"
                >
                  <strong>{{ trace.title }}</strong>
                  <span class="muted">{{ trace.latestAt }}</span>
                  <small class="muted">
                    {{ trace.messageCount }} 条消息 · {{ trace.toolCount }} 个工具事件 · {{ trace.artifactCount }} 个产物
                  </small>
                  <p v-if="trace.preview">{{ trace.preview }}</p>
                </button>
              </div>

              <div v-if="activeTrace" class="trace-timeline">
                <div class="trace-active-head">
                  <strong>{{ activeTrace.traceId }}</strong>
                  <span class="muted">状态 {{ activeTrace.status || 'active' }}</span>
                </div>
                <div v-if="activeTraceTimeline.length" class="trace-timeline-list">
                  <article v-for="item in activeTraceTimeline" :key="item.key" class="trace-timeline-item">
                    <div class="trace-timeline-head">
                      <strong>{{ item.title }}</strong>
                      <span class="muted">{{ item.createdAt }}</span>
                    </div>
                    <small class="muted">{{ item.category }} · {{ item.status || 'active' }}</small>
                    <p v-if="item.detail">{{ item.detail }}</p>
                  </article>
                </div>
                <div v-else class="muted">当前 trace 还没有时间线事件。</div>
              </div>
            </div>
            <div class="message-compose">
              <el-input
                v-model="messageInput"
                type="textarea"
                :autosize="{ minRows: 3, maxRows: 6 }"
                placeholder="在这里直接和龙虾对话。平台会保存消息记录，并通过会话级 SSE 事件展示文本、工具过程与推理状态。"
              />
              <div class="message-actions">
                <el-button
                  type="primary"
                  :disabled="!activeSession || activeSessionArchived || !messageInput.trim()"
                  :loading="messageSubmitting"
                  @click="sendMessageRecord"
                >
                  发送到龙虾
                </el-button>
              </div>
              <el-alert v-if="messageFeedback" :closable="false" show-icon type="info" :title="messageFeedback" />
            </div>
          </div>

          <iframe
            v-if="workspaceFrameUrl"
            :key="frameKey"
            :src="workspaceFrameUrl"
            class="workspace-frame"
            allow="clipboard-read; clipboard-write"
            referrerpolicy="strict-origin-when-cross-origin"
            :title="`${brand.name} Workspace`"
          />
          <div v-else class="artifact-fallback">
            <p class="muted">当前会话没有可用的嵌入式工作台入口，但平台侧会话、消息历史与产物管理仍可直接使用。</p>
          </div>
        </el-card>

        <el-card shadow="never" class="workspace-column column-artifacts">
          <SectionHeader title="产物" subtitle="保存、查看龙虾输出的网页 / 文档 / PPTX / PDF" />
          <div class="artifact-form">
            <div class="artifact-inputs">
              <el-input
                v-model="artifactInput"
                type="textarea"
                :autosize="{ minRows: 3, maxRows: 5 }"
                placeholder="把龙虾输出的网页、PDF、PPTX、DOCX、XLSX 链接粘贴到这里"
              />
              <el-input
                v-model="artifactPreviewInput"
                type="textarea"
                :autosize="{ minRows: 2, maxRows: 4 }"
                placeholder="可选：预览衍生地址。PPTX / DOCX / XLSX 的正式预览建议这里填写 PDF 或 HTML 版本 URL。"
              />
            </div>
            <el-alert
              :closable="false"
              show-icon
              type="info"
              title="正式策略：HTML 走平台沙箱代理；PPTX / DOCX / XLSX 优先使用 PDF 或 HTML 衍生预览，否则自动回退为下载。"
            />
            <div class="artifact-actions">
              <el-button type="primary" :disabled="!artifactInput.trim()" @click="previewArtifact">临时检查 URL</el-button>
              <el-button plain :disabled="!artifactUrl || !activeSession || activeSessionArchived" :loading="artifactSubmitting" @click="saveArtifact">
                保存到会话
              </el-button>
              <el-button plain :disabled="!artifactUrl" @click="openArtifactExternal">新窗口打开</el-button>
            </div>
          </div>

          <div v-if="artifactUrl" class="artifact-meta">
            <strong>{{ artifactLabel }}</strong>
            <span class="muted">{{ artifactUrl }}</span>
            <small v-if="artifactPreviewInput.trim()" class="muted">预览衍生地址：{{ artifactPreviewInput.trim() }}</small>
          </div>

          <div v-if="artifacts.length" class="saved-artifacts">
            <SectionHeader title="会话产物" subtitle="当前会话已归档产物" />
            <div class="artifact-list">
              <button
                v-for="item in artifacts"
                :key="item.id"
                type="button"
                :class="['artifact-item', { active: selectedArtifactId === item.id }]"
                @click="openSavedArtifact(item)"
              >
                <strong>{{ item.title }}</strong>
                <span class="muted">{{ item.kind }} · {{ item.updatedAt }}</span>
              </button>
            </div>
          </div>

          <div v-if="selectedArtifact" class="artifact-preview-card">
            <SectionHeader title="正式预览" subtitle="保存后由平台预览网关统一代理、隔离与回退" />
            <div v-if="artifactPreviewLoading" class="muted">正在加载正式预览…</div>
            <el-alert v-else-if="artifactPreviewError" :closable="false" show-icon type="error" :title="artifactPreviewError" />
            <template v-else-if="artifactPreview">
              <div class="artifact-preview-meta">
                <span class="hero-chip">策略 {{ artifactPreview.strategy }}</span>
                <span class="hero-chip">模式 {{ artifactPreview.mode }}</span>
                <span class="hero-chip">{{ artifactPreview.sandboxed ? '沙箱隔离' : '平台代理' }}</span>
              </div>
              <el-alert v-if="artifactPreview.note" :closable="false" show-icon type="info" :title="artifactPreview.note" />
              <div class="artifact-toolbar">
                <el-button v-if="artifactPreview.available && artifactPreview.previewUrl" type="primary" plain @click="openFormalPreview">新窗口打开正式预览</el-button>
                <el-button v-if="artifactPreview.downloadUrl" plain @click="downloadArtifact">下载源文件</el-button>
                <el-button v-if="artifactPreview.externalUrl" plain @click="openArtifactExternal">打开源地址</el-button>
              </div>
              <iframe
                v-if="artifactPreview.available && artifactPreview.mode === 'html' && artifactPreview.previewUrl"
                :src="artifactPreview.previewUrl"
                class="artifact-frame"
                title="Artifact Formal Preview"
                sandbox="allow-forms allow-scripts allow-modals allow-popups allow-downloads"
                referrerpolicy="no-referrer"
              />
              <iframe
                v-else-if="artifactPreview.available && (artifactPreview.mode === 'pdf' || artifactPreview.mode === 'text') && artifactPreview.previewUrl"
                :src="artifactPreview.previewUrl"
                class="artifact-frame"
                title="Artifact Formal Preview"
                referrerpolicy="no-referrer"
              />
              <img
                v-else-if="artifactPreview.available && artifactPreview.mode === 'image' && artifactPreview.previewUrl"
                :src="artifactPreview.previewUrl"
                class="artifact-image"
                alt="Artifact Formal Preview"
              />
              <video
                v-else-if="artifactPreview.available && artifactPreview.mode === 'video' && artifactPreview.previewUrl"
                :src="artifactPreview.previewUrl"
                class="artifact-media"
                controls
                playsinline
              />
              <audio
                v-else-if="artifactPreview.available && artifactPreview.mode === 'audio' && artifactPreview.previewUrl"
                :src="artifactPreview.previewUrl"
                class="artifact-audio"
                controls
              />
              <div v-else class="artifact-fallback">
                <p class="muted">{{ artifactPreview.failureReason || '平台当前无法对该产物做正式内嵌预览。' }}</p>
                <div class="artifact-actions">
                  <el-button v-if="artifactPreview.downloadUrl" type="primary" @click="downloadArtifact">下载源文件</el-button>
                  <el-button v-if="artifactPreview.externalUrl" plain @click="openArtifactExternal">打开源地址</el-button>
                </div>
              </div>
            </template>
          </div>

          <div v-else-if="artifactUrl" class="artifact-preview-card">
            <SectionHeader title="临时检查" subtitle="用于确认 URL 可用性，正式策略以保存后的平台网关结果为准" />
            <el-alert v-if="draftArtifactPreview.note" :closable="false" show-icon type="info" :title="draftArtifactPreview.note" />
            <iframe
              v-if="draftArtifactPreview.mode === 'html' && draftArtifactPreview.url"
              :src="draftArtifactPreview.url"
              class="artifact-frame"
              title="Artifact Draft Preview"
              sandbox="allow-forms allow-scripts allow-modals allow-popups allow-downloads"
              referrerpolicy="no-referrer"
            />
            <iframe
              v-else-if="(draftArtifactPreview.mode === 'pdf' || draftArtifactPreview.mode === 'text') && draftArtifactPreview.url"
              :src="draftArtifactPreview.url"
              class="artifact-frame"
              title="Artifact Draft Preview"
              referrerpolicy="strict-origin-when-cross-origin"
            />
            <img
              v-else-if="draftArtifactPreview.mode === 'image' && draftArtifactPreview.url"
              :src="draftArtifactPreview.url"
              class="artifact-image"
              alt="Artifact Draft Preview"
            />
            <video
              v-else-if="draftArtifactPreview.mode === 'video' && draftArtifactPreview.url"
              :src="draftArtifactPreview.url"
              class="artifact-media"
              controls
              playsinline
            />
            <audio
              v-else-if="draftArtifactPreview.mode === 'audio' && draftArtifactPreview.url"
              :src="draftArtifactPreview.url"
              class="artifact-audio"
              controls
            />
            <div v-else class="artifact-fallback">
              <p class="muted">{{ draftArtifactPreview.reason || '当前链接无法安全嵌入预览，请使用新窗口打开。' }}</p>
              <el-button type="primary" @click="openArtifactExternal">打开产物</el-button>
            </div>
          </div>
        </el-card>
      </section>
    </template>
  </div>
</template>

<style scoped>
.workspace-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.workspace-hero,
.workspace-column,
.state-card {
  padding: 18px;
}

.workspace-hero {
  display: grid;
  gap: 14px;
}

.workspace-hero h2 {
  margin: 8px 0;
  font-size: 2rem;
}

.hero-copy {
  display: grid;
  gap: 8px;
}

.hero-metas {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-chip {
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--panel-muted);
  color: var(--text-muted);
  font-size: 0.9rem;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.workspace-layout {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr) 360px;
  gap: 14px;
}

.workspace-column {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 72vh;
}

.column-chat {
  min-width: 0;
}

.session-list,
.artifact-list,
.access-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.session-filters,
.session-actions {
  display: grid;
  gap: 10px;
}

.session-current-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.session-item,
.artifact-item,
.access-item {
  display: grid;
  gap: 4px;
  text-align: left;
  padding: 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.session-item.active {
  border-color: rgba(29, 107, 255, 0.28);
  background: rgba(29, 107, 255, 0.08);
}

.session-flags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.session-flag {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: 999px;
  background: rgba(29, 107, 255, 0.12);
  color: #1d4ed8;
  font-size: 12px;
}

.artifact-item.active {
  border-color: rgba(22, 163, 74, 0.28);
  background: rgba(22, 163, 74, 0.08);
}

.message-panel {
  display: grid;
  gap: 12px;
}

.realtime-status {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 320px;
  overflow: auto;
}

.message-list-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-bottom: 4px;
}

.message-item {
  display: grid;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
  content-visibility: auto;
  contain-intrinsic-size: 220px;
}

.message-item p {
  margin: 0;
  white-space: pre-wrap;
}

.message-item-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.message-item--trace {
  box-shadow: inset 0 0 0 1px rgba(29, 107, 255, 0.18);
}

.message-item--anchored {
  box-shadow:
    inset 0 0 0 2px rgba(14, 165, 233, 0.4),
    0 12px 30px rgba(14, 165, 233, 0.08);
}

.message-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.message-trace-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.trace-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border: none;
  border-radius: 999px;
  background: rgba(29, 107, 255, 0.12);
  color: #1d4ed8;
  font-size: 0.82rem;
  cursor: pointer;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.message-status {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 0.82rem;
  background: rgba(148, 163, 184, 0.14);
  color: var(--text-muted);
}

.status-streaming {
  background: rgba(29, 107, 255, 0.12);
  color: #1d4ed8;
}

.status-delivered {
  background: rgba(22, 163, 74, 0.12);
  color: #15803d;
}

.status-failed {
  background: rgba(220, 38, 38, 0.12);
  color: #b91c1c;
}

.message-error,
.message-reasoning,
.message-tools {
  display: grid;
  gap: 6px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
}

.message-error {
  background: rgba(220, 38, 38, 0.08);
  color: #b91c1c;
}

.message-reasoning {
  background: rgba(245, 158, 11, 0.08);
}

.message-reasoning p,
.tool-item p {
  margin: 0;
  white-space: pre-wrap;
}

.message-tools {
  background: rgba(15, 23, 42, 0.04);
}

.tool-item {
  display: grid;
  gap: 6px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.65);
}

.tool-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.role-user {
  border-left: 3px solid rgba(29, 107, 255, 0.4);
}

.role-assistant {
  border-left: 3px solid rgba(22, 163, 74, 0.4);
}

.role-system,
.role-note {
  border-left: 3px solid rgba(148, 163, 184, 0.5);
}

.message-compose,
.artifact-form {
  display: grid;
  gap: 10px;
}

.trace-panel {
  display: grid;
  gap: 10px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: rgba(15, 23, 42, 0.03);
}

.trace-summary-list,
.trace-timeline-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.trace-summary,
.trace-timeline-item {
  display: grid;
  gap: 4px;
  padding: 12px;
  text-align: left;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.trace-summary.active {
  border-color: rgba(29, 107, 255, 0.28);
  background: rgba(29, 107, 255, 0.08);
}

.trace-summary p,
.trace-timeline-item p {
  margin: 0;
  white-space: pre-wrap;
}

.trace-timeline {
  display: grid;
  gap: 10px;
}

.trace-active-head,
.trace-timeline-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.artifact-inputs {
  display: grid;
  gap: 10px;
}

.message-actions {
  display: flex;
  justify-content: flex-end;
}

.artifact-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.artifact-meta {
  display: grid;
  gap: 4px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.artifact-preview-meta,
.artifact-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.workspace-frame {
  width: 100%;
  flex: 1;
  min-height: 56vh;
  border: none;
  border-radius: 18px;
  background: #fff;
}

.artifact-frame {
  width: 100%;
  min-height: 42vh;
  border: none;
  border-radius: 18px;
  background: #fff;
}

.artifact-image,
.artifact-media {
  width: 100%;
  border-radius: 18px;
  background: #000;
}

.artifact-audio {
  width: 100%;
}

.artifact-fallback {
  display: grid;
  gap: 10px;
  justify-items: center;
  padding: 28px;
}

.state-card {
  text-align: center;
}

@media (max-width: 1280px) {
  .workspace-layout {
    grid-template-columns: 1fr;
  }

  .workspace-column {
    min-height: auto;
  }
}

@media (max-width: 768px) {
  .workspace-frame {
    min-height: 64vh;
  }

  .hero-actions,
  .artifact-actions {
    flex-direction: column;
  }
}
</style>
