<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { AdminDiagnosticSessionDetail, DiagnosticSessionSummary } from '../../lib/types'

const filters = reactive({
  status: '',
  q: '',
})

const sessions = ref<DiagnosticSessionSummary[]>([])
const selectedId = ref('')
const detail = ref<AdminDiagnosticSessionDetail | null>(null)
const loading = ref(true)
const detailLoading = ref(false)
const actionError = ref('')

const selectedSession = computed(() => detail.value?.session ?? sessions.value.find((item) => item.id === selectedId.value) ?? null)

async function loadSessions() {
  loading.value = true
  actionError.value = ''
  try {
    sessions.value = await api.getAdminTerminalSessions({
      status: filters.status || undefined,
      q: filters.q || undefined,
    })
    if (!selectedId.value && sessions.value.length) {
      selectedId.value = sessions.value[0].id
    }
    if (selectedId.value) {
      await loadDetail(selectedId.value)
    }
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '加载诊断会话失败'
  } finally {
    loading.value = false
  }
}

async function loadDetail(id: string) {
  detailLoading.value = true
  try {
    detail.value = await api.getAdminDiagnosticSession(id)
    selectedId.value = id
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '加载诊断会话详情失败'
  } finally {
    detailLoading.value = false
  }
}

onMounted(async () => {
  await loadSessions()
})
</script>

<template>
  <div class="diagnostics-shell">
    <section class="card panel">
      <SectionHeader title="终端与诊断会话" subtitle="跨实例查看只读/白名单诊断会话与命令轨迹" />
      <div class="filters">
        <el-select v-model="filters.status" placeholder="全部状态" clearable>
          <el-option label="active" value="active" />
          <el-option label="closed" value="closed" />
          <el-option label="expired" value="expired" />
        </el-select>
        <el-input v-model="filters.q" placeholder="搜索会话号、Pod、操作人" clearable />
        <el-button type="primary" @click="loadSessions">刷新</el-button>
      </div>
      <div v-if="loading" class="state-card">正在同步诊断会话…</div>
      <div v-else class="session-list">
        <button
          v-for="item in sessions"
          :key="item.id"
          type="button"
          :class="['session-item', { 'session-item--active': item.id === selectedId }]"
          @click="loadDetail(item.id)"
        >
          <div class="session-item__head">
            <strong>{{ item.sessionNo }}</strong>
            <span class="pill">{{ item.status }}</span>
          </div>
          <div>{{ item.podName }} · {{ item.accessMode }} · 实例 #{{ item.instanceId }}</div>
          <small class="muted">{{ item.operator }} · {{ item.updatedAt }}</small>
        </button>
      </div>
    </section>

    <section class="card panel detail-panel">
      <SectionHeader title="会话详情" subtitle="录制、命令审计与实例回跳" />
      <div v-if="detailLoading" class="state-card">正在加载会话详情…</div>
      <div v-else-if="selectedSession" class="detail-stack">
        <div class="detail-meta">
          <div>
            <div class="eyebrow">{{ selectedSession.accessMode }}</div>
            <h3>{{ selectedSession.sessionNo }}</h3>
            <p class="muted">
              实例 #{{ selectedSession.instanceId }} · {{ selectedSession.podName }} · {{ selectedSession.status }}
            </p>
          </div>
          <div class="link-row">
            <RouterLink :to="`/admin/instances/${selectedSession.instanceId}?tab=diagnostics`">实例诊断页</RouterLink>
            <RouterLink :to="`/admin/instances/${selectedSession.instanceId}?tab=monitoring`">实例监控页</RouterLink>
            <RouterLink :to="`/admin/instances/${selectedSession.instanceId}/workspace`">工作台</RouterLink>
          </div>
        </div>

        <div class="meta-grid">
          <div class="meta-card">
            <span class="muted">命名空间</span>
            <strong>{{ selectedSession.namespace || '—' }}</strong>
          </div>
          <div class="meta-card">
            <span class="muted">工作负载</span>
            <strong>{{ selectedSession.workloadName || '—' }}</strong>
          </div>
          <div class="meta-card">
            <span class="muted">审批单号</span>
            <strong>{{ selectedSession.approvalTicket || '—' }}</strong>
          </div>
          <div class="meta-card">
            <span class="muted">最近命令</span>
            <strong>{{ selectedSession.lastCommandText || '—' }}</strong>
          </div>
        </div>

        <div class="transcript-card">
          <strong>录制文本</strong>
          <pre>{{ detail?.record || '当前会话还没有命令记录。' }}</pre>
        </div>

        <div class="history-list">
          <div v-for="item in detail?.commands ?? []" :key="item.id" class="history-item">
            <strong>{{ item.commandText }}</strong>
            <span class="muted">{{ item.status }} · exit={{ item.exitCode }} · {{ item.executedAt }}</span>
            <small v-if="item.errorOutput" class="muted">{{ item.errorOutput }}</small>
          </div>
        </div>
      </div>
      <div v-else class="state-card">选择一条诊断会话后查看详情。</div>
      <el-alert v-if="actionError" :closable="false" show-icon type="error" :title="actionError" />
    </section>
  </div>
</template>

<style scoped>
.diagnostics-shell {
  display: grid;
  grid-template-columns: 0.92fr 1.08fr;
  gap: 14px;
}

.panel {
  padding: 16px;
}

.filters {
  display: grid;
  grid-template-columns: 0.8fr 1.2fr auto;
  gap: 10px;
  margin: 14px 0;
}

.state-card {
  padding: 24px;
  text-align: center;
  border: 1px dashed var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel-muted);
}

.session-list,
.history-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.session-item,
.meta-card,
.transcript-card,
.history-item {
  padding: 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
  text-align: left;
}

.session-item--active {
  border-color: rgba(59, 130, 246, 0.55);
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.2);
}

.session-item__head,
.detail-meta {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.detail-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.link-row {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.transcript-card pre {
  margin: 10px 0 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Cascadia Code', 'JetBrains Mono', monospace;
  font-size: 13px;
}

.history-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

@media (max-width: 1180px) {
  .diagnostics-shell {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 860px) {
  .filters,
  .meta-grid {
    grid-template-columns: 1fr;
  }

  .detail-meta {
    flex-direction: column;
  }
}
</style>
