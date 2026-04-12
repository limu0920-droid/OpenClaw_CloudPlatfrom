<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { api } from '../../lib/api'
import type { SearchConfig, SearchLogItem } from '../../lib/types'

const filters = reactive({
  q: '',
  kind: '',
  instanceId: '',
})

const searchConfig = ref<SearchConfig | null>(null)
const searchResult = ref<{ backend: string; items: SearchLogItem[] } | null>(null)
const loading = ref(true)
const error = ref('')

const backendLabel = computed(() => {
  const backend = searchResult.value?.backend || searchConfig.value?.provider

  if (backend === 'opensearch') {
    return 'OpenSearch'
  }

  if (backend === 'mock') {
    return '平台内置检索后端'
  }

  return backend || '平台检索后端'
})

function resolveInstancePath(row: SearchLogItem) {
  return row.instancePath || (row.instanceId ? `/portal/instances/${row.instanceId}` : '')
}

function resolveWorkspacePath(row: SearchLogItem) {
  const query = new URLSearchParams()
  if (row.sessionId) query.set('sessionId', row.sessionId)
  if (row.messageId) query.set('messageId', row.messageId)
  if (row.traceId) query.set('traceId', row.traceId)
  const suffix = query.toString()
  return row.workspacePath || (row.instanceId ? `/portal/instances/${row.instanceId}/workspace${suffix ? `?${suffix}` : ''}` : '')
}

async function search() {
  loading.value = true
  error.value = ''
  try {
    searchResult.value = await api.searchLogs(filters, 'portal')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '检索失败'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  searchConfig.value = await api.getSearchConfig()
  await search()
})
</script>

<template>
  <div class="stack">
    <el-card shadow="never" class="panel">
      <div class="title">日志与审计检索</div>
      <p class="muted">
        当前接入 {{ backendLabel }}。查询表单会直接复用后端当前检索配置。
      </p>
      <div class="filters">
        <el-input v-model="filters.q" placeholder="搜索标题、消息、来源" clearable />
        <el-select v-model="filters.kind" placeholder="全部类型" clearable>
          <el-option label="audit" value="audit" />
          <el-option label="alert" value="alert" />
          <el-option label="workspace_event" value="workspace_event" />
          <el-option label="diagnostic" value="diagnostic" />
          <el-option label="ticket" value="ticket" />
          <el-option label="channel" value="channel" />
        </el-select>
        <el-input v-model="filters.instanceId" placeholder="实例 ID，如 100" clearable />
        <el-button type="primary" @click="search">搜索</el-button>
      </div>
    </el-card>

    <el-card shadow="never" class="panel">
      <div v-if="loading" class="state-card">正在检索日志…</div>
      <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
      <el-table v-else :data="searchResult?.items ?? []" class="surface-table">
        <el-table-column label="类型" min-width="120">
          <template #default="{ row }">
            <el-tag round disable-transitions>{{ row.kind }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="source" label="来源" min-width="140" />
        <el-table-column prop="message" label="消息" min-width="260" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="时间" min-width="180" />
        <el-table-column label="上下文" min-width="180">
          <template #default="{ row }">
            <div class="jump-links">
              <RouterLink v-if="resolveInstancePath(row)" :to="resolveInstancePath(row)">实例</RouterLink>
              <RouterLink v-if="resolveWorkspacePath(row)" :to="resolveWorkspacePath(row)">工作台</RouterLink>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.panel {
  padding: 14px;
}

.title {
  font-weight: 700;
  margin-bottom: 8px;
}

.filters {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr auto;
  gap: 10px;
  margin-top: 12px;
}

.state-card {
  padding: 18px;
  border: 1px dashed var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel-muted);
  text-align: center;
}

.jump-links {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

@media (max-width: 1024px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
