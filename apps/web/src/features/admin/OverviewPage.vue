<script setup lang="ts">
import StatCard from '../../components/StatCard.vue'
import SectionHeader from '../../components/SectionHeader.vue'
import ListCard from '../../components/ListCard.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: overview, loading, error } = useAsyncData(() => api.getAdminOverview())
</script>

<template>
  <div class="grid">
    <div v-if="loading" class="card state-card">正在同步控制平面指标…</div>
    <div v-else-if="error" class="card state-card state-card--error">{{ error }}</div>
    <template v-else-if="overview">
      <div class="stat-grid">
        <StatCard v-for="m in overview.metrics" :key="m.label" v-bind="m" />
      </div>

      <div class="two-col">
        <div class="card">
          <SectionHeader title="任务队列" subtitle="全局任务" actionLabel="查看全部" />
          <ul class="tasks">
            <li v-for="job in overview.tasks" :key="job.id">
              <div>
                <div class="strong">{{ job.type }}</div>
                <div class="muted">目标 {{ job.target }}</div>
              </div>
              <div class="pill">{{ job.status }}</div>
              <div class="muted">{{ job.startedAt }}</div>
            </li>
          </ul>
        </div>

        <ListCard
          title="告警"
          :items="
            overview.alerts.map((a) => ({
              id: a.id,
              label: a.title,
              meta: `${a.target} · ${a.time}`,
              status: a.severity,
              accent:
                a.severity === 'critical'
                  ? '#ef4444'
                  : a.severity === 'warning'
                    ? '#f59e0b'
                    : '#22d3ee',
            }))
          "
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.state-card {
  padding: 24px;
  text-align: center;
}

.state-card--error {
  color: #fecaca;
}
.card {
  padding: 14px;
}
.two-col {
  display: grid;
  grid-template-columns: 1fr 0.9fr;
  gap: 14px;
}
.tasks {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.tasks li {
  display: grid;
  grid-template-columns: 1.4fr auto 1fr;
  gap: 10px;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}
@media (max-width: 1024px) {
  .two-col {
    grid-template-columns: 1fr;
  }
}
</style>
