<script setup lang="ts">
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: tenants, loading, error } = useAsyncData(() => api.getAdminTenants())
</script>

<template>
  <div class="card">
    <div class="title">租户</div>
    <div v-if="loading" class="state-card">正在读取租户列表…</div>
    <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
    <div v-else class="table">
      <div class="head">
        <span>名称</span>
        <span>编码</span>
        <span>套餐</span>
        <span>状态</span>
        <span>到期</span>
      </div>
      <div v-for="t in tenants" :key="t.id" class="row">
        <span class="strong">{{ t.name }}</span>
        <span class="muted">{{ t.code }}</span>
        <span>{{ t.plan }}</span>
        <span class="pill">{{ t.status }}</span>
        <span class="muted">{{ t.expiresAt || '—' }}</span>
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
  grid-template-columns: 1.6fr 1fr 1fr 1fr 1fr;
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
