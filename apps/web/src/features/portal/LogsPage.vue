<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

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

async function search() {
  loading.value = true
  error.value = ''
  try {
    searchResult.value = await api.searchLogs(filters)
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
    <div class="card panel">
      <div class="title">日志与审计检索</div>
      <p class="muted">
        当前使用 {{ searchConfig?.enabled ? 'OpenSearch 配置骨架' : 'Mock 搜索后端' }}。后续接入真实 OpenSearch 后，可直接复用本页查询表单。
      </p>
      <div class="filters">
        <input v-model="filters.q" placeholder="搜索标题、消息、来源" />
        <select v-model="filters.kind">
          <option value="">全部类型</option>
          <option value="audit">audit</option>
          <option value="ticket">ticket</option>
          <option value="channel">channel</option>
        </select>
        <input v-model="filters.instanceId" placeholder="实例 ID，如 100" />
        <button class="primary" @click="search">搜索</button>
      </div>
    </div>

    <div class="card panel">
      <div v-if="loading" class="state-card">正在检索日志…</div>
      <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
      <div v-else class="table">
        <div class="head">
          <span>类型</span>
          <span>标题</span>
          <span>来源</span>
          <span>消息</span>
          <span>时间</span>
        </div>
        <div v-for="item in searchResult?.items" :key="item.id" class="row">
          <span class="pill">{{ item.kind }}</span>
          <span class="strong">{{ item.title }}</span>
          <span>{{ item.source }}</span>
          <span class="muted">{{ item.message }}</span>
          <span class="muted">{{ item.createdAt }}</span>
        </div>
      </div>
    </div>
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

input,
select {
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--stroke);
  background: #fff;
}

.primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 14px;
  border-radius: 12px;
  border: 1px solid transparent;
  background: linear-gradient(120deg, var(--brand), var(--brand-strong));
  color: #fff;
}

.table {
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.head,
.row {
  display: grid;
  grid-template-columns: 0.8fr 1.2fr 1fr 1.8fr 1fr;
  gap: 10px;
  padding: 12px 14px;
  align-items: center;
}

.head {
  background: var(--panel-muted);
  color: var(--text-muted);
  font-weight: 600;
}

.row {
  border-top: 1px solid var(--stroke);
}

@media (max-width: 1024px) {
  .filters,
  .head,
  .row {
    grid-template-columns: 1fr;
  }
}
</style>
