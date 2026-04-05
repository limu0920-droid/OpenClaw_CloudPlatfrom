<script setup lang="ts">
import { RouterLink } from 'vue-router'

import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: instances, loading, error } = useAsyncData(() => api.getAdminInstances())
</script>

<template>
  <div class="card">
    <div class="title">实例（全局）</div>
    <div v-if="loading" class="state-card">正在读取全局实例…</div>
    <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
    <div v-else class="table">
      <div class="head">
        <span>名称</span>
        <span>租户</span>
        <span>状态</span>
        <span>版本</span>
        <span>区域</span>
        <span>计划</span>
        <span>更新时间</span>
        <span>操作</span>
      </div>
      <div v-for="inst in instances" :key="inst.id" class="row">
        <span class="strong">{{ inst.name }}</span>
        <span class="muted">{{ inst.tenantName || '—' }}</span>
        <span class="pill">{{ inst.status }}</span>
        <span>{{ inst.version }}</span>
        <span>{{ inst.region }}</span>
        <span>{{ inst.plan }}</span>
        <span class="muted">{{ inst.updatedAt }}</span>
        <span><RouterLink :to="`/admin/instances/${inst.id}`">详情</RouterLink></span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card {
  padding: 14px;
}

.state-card {
  padding: 18px 0;
  text-align: center;
}

.state-card--error {
  color: #fecaca;
}
.title {
  font-weight: 700;
  margin-bottom: 10px;
}
.table {
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
.head,
.row {
  display: grid;
  grid-template-columns: 1.3fr 1.1fr 0.9fr 0.9fr 1fr 1fr 1.1fr 0.7fr;
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
</style>
