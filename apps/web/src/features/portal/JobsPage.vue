<script setup lang="ts">
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: jobs, loading, error } = useAsyncData(() => api.getPortalJobs())
</script>

<template>
  <div class="card">
    <div class="title">任务中心</div>
    <div v-if="loading" class="state-card">正在同步任务状态…</div>
    <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
    <div v-else class="table">
      <div class="head">
        <span>任务</span>
        <span>目标</span>
        <span>状态</span>
        <span>开始</span>
        <span>进度</span>
      </div>
      <div v-for="job in jobs" :key="job.id" class="row">
        <span class="strong">{{ job.type }}</span>
        <span>{{ job.target }}</span>
        <span class="pill">{{ job.status }}</span>
        <span class="muted">{{ job.startedAt }}</span>
        <span>{{ job.progress ? `${job.progress}%` : '—' }}</span>
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
  color: #b91c1c;
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
  grid-template-columns: 1.6fr 1.2fr 1fr 1fr 0.8fr;
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
