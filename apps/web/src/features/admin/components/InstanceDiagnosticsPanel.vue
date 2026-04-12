<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { api } from '../../../lib/api'
import type { AdminDiagnosticSessionDetail, AdminInstanceDiagnosticsSummary } from '../../../lib/types'

const props = defineProps<{
  instanceId: string
}>()

const summary = ref<AdminInstanceDiagnosticsSummary | null>(null)
const sessionDetail = ref<AdminDiagnosticSessionDetail | null>(null)
const loading = ref(true)
const sessionLoading = ref(false)
const creating = ref(false)
const runningCommand = ref('')
const closingSession = ref(false)
const error = ref('')
const actionError = ref('')
const selectedSessionId = ref('')
const manualCommand = ref('')

const createForm = ref({
  podName: '',
  containerName: '',
  accessMode: 'readonly' as 'readonly' | 'whitelist',
  approvalTicket: '',
  approvedBy: '',
  reason: '',
})

const availablePods = computed(() => summary.value?.pods ?? [])
const signals = computed(() => summary.value?.signals ?? [])
const sessions = computed(() => summary.value?.sessions ?? [])
const commandCatalog = computed(() => sessionDetail.value?.commandCatalog ?? summary.value?.policy.commandCatalog ?? [])
const selectedSession = computed(() => sessionDetail.value?.session ?? sessions.value.find((item) => item.id === selectedSessionId.value) ?? null)
const activeSessionCount = computed(() => sessions.value.filter((item) => item.status === 'active').length)
const manualAllowed = computed(() => selectedSession.value?.accessMode === 'whitelist')
const transcript = computed(() =>
  sessionDetail.value?.commands.length
    ? sessionDetail.value.commands
        .map((item) =>
          [`$ ${item.commandText}`, item.output, item.errorOutput, `[status=${item.status} exit=${item.exitCode} durationMs=${item.durationMs} at=${item.executedAt}]`]
            .filter(Boolean)
            .join('\n'),
        )
        .join('\n\n')
    : '',
)

async function loadSummary() {
  loading.value = true
  error.value = ''

  try {
    const response = await api.getAdminInstanceDiagnostics(props.instanceId)
    summary.value = response
    createForm.value.podName = createForm.value.podName || response.pods[0]?.name || ''

    const stillSelected = response.sessions.find((item) => item.id === selectedSessionId.value)
    selectedSessionId.value = stillSelected?.id || response.sessions[0]?.id || ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载诊断信息失败'
  } finally {
    loading.value = false
  }
}

async function loadSession(sessionId: string) {
  if (!sessionId) {
    sessionDetail.value = null
    return
  }

  sessionLoading.value = true
  actionError.value = ''

  try {
    sessionDetail.value = await api.getAdminDiagnosticSession(sessionId)
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '加载诊断会话失败'
  } finally {
    sessionLoading.value = false
  }
}

async function createSession() {
  creating.value = true
  actionError.value = ''

  try {
    const session = await api.createAdminDiagnosticSession(props.instanceId, {
      podName: createForm.value.podName || undefined,
      containerName: createForm.value.containerName || undefined,
      accessMode: createForm.value.accessMode,
      approvalTicket: createForm.value.accessMode === 'whitelist' ? createForm.value.approvalTicket || undefined : undefined,
      approvedBy: createForm.value.accessMode === 'whitelist' ? createForm.value.approvedBy || undefined : undefined,
      reason: createForm.value.reason || undefined,
    })

    selectedSessionId.value = session.id
    manualCommand.value = ''
    await loadSummary()
    await loadSession(session.id)
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '创建诊断会话失败'
  } finally {
    creating.value = false
  }
}

async function runPreset(commandKey: string) {
  if (!selectedSessionId.value) {
    return
  }

  runningCommand.value = commandKey
  actionError.value = ''

  try {
    await api.executeAdminDiagnosticCommand(selectedSessionId.value, { commandKey })
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '执行诊断命令失败'
  } finally {
    await loadSummary()
    await loadSession(selectedSessionId.value)
    runningCommand.value = ''
  }
}

async function runManualCommand() {
  if (!selectedSessionId.value || !manualCommand.value.trim()) {
    return
  }

  runningCommand.value = '__manual__'
  actionError.value = ''

  try {
    await api.executeAdminDiagnosticCommand(selectedSessionId.value, {
      commandText: manualCommand.value.trim(),
    })
    manualCommand.value = ''
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '执行白名单命令失败'
  } finally {
    await loadSummary()
    await loadSession(selectedSessionId.value)
    runningCommand.value = ''
  }
}

async function closeSession() {
  if (!selectedSessionId.value) {
    return
  }

  closingSession.value = true
  actionError.value = ''

  try {
    await api.closeAdminDiagnosticSession(selectedSessionId.value, 'operator_closed')
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '关闭诊断会话失败'
  } finally {
    await loadSummary()
    await loadSession(selectedSessionId.value)
    closingSession.value = false
  }
}

watch(
  () => props.instanceId,
  async () => {
    selectedSessionId.value = ''
    sessionDetail.value = null
    createForm.value = {
      podName: '',
      containerName: '',
      accessMode: 'readonly',
      approvalTicket: '',
      approvedBy: '',
      reason: '',
    }
    manualCommand.value = ''
    await loadSummary()
  },
  { immediate: true },
)

watch(
  () => selectedSessionId.value,
  async (value, previous) => {
    if (!value || value === previous) {
      return
    }
    await loadSession(value)
  },
)
</script>

<template>
  <div v-if="loading" class="card state-card">正在准备诊断面板…</div>
  <div v-else-if="error" class="card state-card state-card--error">{{ error }}</div>
  <div v-else-if="summary" class="diagnostics-shell">
    <section class="stat-grid">
      <article class="card mini-stat">
        <span class="muted">工作负载</span>
        <strong>{{ summary.workload?.name || summary.binding?.workloadName || '—' }}</strong>
        <small>{{ summary.binding?.namespace || '未绑定命名空间' }}</small>
      </article>
      <article class="card mini-stat">
        <span class="muted">Pod 数量</span>
        <strong>{{ summary.pods.length }}</strong>
        <small>{{ summary.pods.map((item) => item.status).join(' / ') || '—' }}</small>
      </article>
      <article class="card mini-stat">
        <span class="muted">活跃会话</span>
        <strong>{{ activeSessionCount }}</strong>
        <small>上限 {{ summary.policy.maxActiveSessionsPerInstance }}</small>
      </article>
      <article class="card mini-stat">
        <span class="muted">CPU / 内存</span>
        <strong>{{ summary.metrics?.cpuUsageMilli ?? 0 }}m / {{ summary.metrics?.memoryUsageMB ?? 0 }}MB</strong>
        <small>{{ summary.metrics?.requestsPerMinute ?? 0 }} rpm</small>
      </article>
    </section>

    <section class="detail-grid">
      <div class="card panel">
        <h3>诊断入口</h3>
        <div class="form-grid">
          <label>
            <span>Pod</span>
            <el-select v-model="createForm.podName" placeholder="选择 Pod">
              <el-option v-for="pod in availablePods" :key="pod.name" :label="`${pod.name} · ${pod.status}`" :value="pod.name" />
            </el-select>
          </label>
          <label>
            <span>模式</span>
            <el-select v-model="createForm.accessMode">
              <el-option label="只读诊断" value="readonly" />
              <el-option label="白名单命令" value="whitelist" />
            </el-select>
          </label>
          <label>
            <span>容器名</span>
            <el-input v-model="createForm.containerName" placeholder="默认主容器" />
          </label>
          <label>
            <span>诊断原因</span>
            <el-input v-model="createForm.reason" placeholder="例如：排查 CPU 波动与重启" />
          </label>
          <label v-if="createForm.accessMode === 'whitelist'">
            <span>审批单号</span>
            <el-input v-model="createForm.approvalTicket" placeholder="必填" />
          </label>
          <label v-if="createForm.accessMode === 'whitelist'">
            <span>批准人</span>
            <el-input v-model="createForm.approvedBy" placeholder="例如：platform-ops" />
          </label>
        </div>
        <div class="panel-actions">
          <el-button type="primary" :loading="creating" @click="createSession">创建诊断会话</el-button>
          <span class="muted">
            只读 {{ summary.policy.readonlyTtlMinutes }} 分钟，白名单 {{ summary.policy.whitelistTtlMinutes }} 分钟。
          </span>
        </div>
      </div>

      <div class="card panel">
        <h3>风险信号</h3>
        <div class="signal-list">
          <article v-for="signal in signals" :key="`${signal.type}-${signal.summary}`" class="signal-item">
            <strong>{{ signal.summary }}</strong>
            <span class="muted">{{ signal.severity }} · {{ signal.triggeredAt || signal.podName || '实时聚合' }}</span>
          </article>
        </div>
      </div>
    </section>

    <section class="detail-grid">
      <div class="card panel">
        <h3>会话列表</h3>
        <div v-if="sessions.length" class="session-list">
          <button
            v-for="session in sessions"
            :key="session.id"
            type="button"
            :class="['session-item', { 'session-item--active': selectedSessionId === session.id }]"
            @click="selectedSessionId = session.id"
          >
            <strong>{{ session.sessionNo }}</strong>
            <span>{{ session.podName }} · {{ session.accessMode }} · {{ session.status }}</span>
            <small>{{ session.commandCount }} 条命令 · {{ session.lastCommandAt || session.updatedAt }}</small>
          </button>
        </div>
        <div v-else class="empty-card">当前还没有诊断会话，建议先对目标 Pod 发起只读诊断。</div>
      </div>

      <div class="card panel terminal-panel">
        <div class="terminal-head">
          <div>
            <h3>会话录制</h3>
            <p class="muted">
              {{ selectedSession?.sessionNo || '未选中会话' }}
              <template v-if="selectedSession"> · {{ selectedSession.podName }} · {{ selectedSession.accessMode }}</template>
            </p>
          </div>
          <el-button v-if="selectedSession && selectedSession.status === 'active'" plain :loading="closingSession" @click="closeSession">
            关闭会话
          </el-button>
        </div>

        <div v-if="actionError" class="inline-error">{{ actionError }}</div>

        <div class="command-toolbar">
          <el-button
            v-for="item in commandCatalog"
            :key="item.key"
            plain
            size="small"
            :loading="runningCommand === item.key"
            :disabled="!selectedSession || selectedSession.status !== 'active'"
            @click="runPreset(item.key)"
          >
            {{ item.label }}
          </el-button>
        </div>

        <div v-if="manualAllowed" class="manual-box">
          <el-input
            v-model="manualCommand"
            placeholder="输入白名单命令，例如：df -h"
            @keyup.enter="runManualCommand"
          />
          <el-button
            type="primary"
            plain
            :loading="runningCommand === '__manual__'"
            :disabled="!manualCommand.trim() || !selectedSession || selectedSession.status !== 'active'"
            @click="runManualCommand"
          >
            执行
          </el-button>
        </div>

        <div v-if="sessionLoading" class="terminal-window terminal-window--loading">正在同步诊断录制…</div>
        <pre v-else class="terminal-window">{{ transcript || '当前会话还没有执行记录。' }}</pre>
      </div>
    </section>
  </div>
</template>

<style scoped>
.diagnostics-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.state-card {
  padding: 24px;
  text-align: center;
}

.state-card--error,
.inline-error {
  color: #fecaca;
}

.mini-stat {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.mini-stat strong {
  font-size: 1.35rem;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.panel {
  padding: 14px;
}

.panel h3 {
  margin: 0 0 12px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.panel-actions,
.terminal-head,
.manual-box,
.command-toolbar {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.panel-actions {
  margin-top: 14px;
}

.signal-list,
.session-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.signal-item,
.session-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
}

.session-item {
  text-align: left;
  color: var(--text);
}

.session-item--active {
  border-color: rgba(59, 130, 246, 0.5);
  background: rgba(59, 130, 246, 0.12);
}

.terminal-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.terminal-window {
  min-height: 360px;
  margin: 0;
  padding: 16px;
  overflow: auto;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: #06121f;
  color: #dbeafe;
  font-size: 0.82rem;
  line-height: 1.55;
  white-space: pre-wrap;
}

.terminal-window--loading,
.empty-card {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
}

.empty-card {
  min-height: 120px;
  border-radius: var(--radius-md);
  border: 1px dashed var(--stroke);
}

@media (max-width: 1080px) {
  .detail-grid,
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
