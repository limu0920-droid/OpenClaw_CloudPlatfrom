<script setup lang="ts">
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: alerts, loading, error } = useAsyncData(() => api.getAdminAlerts())
</script>

<template>
  <div class="card">
    <div class="title">告警</div>
    <div v-if="loading" class="state-card">正在同步告警…</div>
    <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
    <div v-else class="table">
      <div class="head">
        <span>告警</span>
        <span>目标</span>
        <span>级别</span>
        <span>时间</span>
      </div>
      <div v-for="alert in alerts" :key="alert.id" class="row">
        <span class="strong">{{ alert.title }}</span>
        <span>{{ alert.target }}</span>
        <span class="pill">{{ alert.severity }}</span>
        <span class="muted">{{ alert.time }}</span>
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
  grid-template-columns: 1.6fr 1.2fr 0.9fr 1fr;
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
