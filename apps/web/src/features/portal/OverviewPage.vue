<script setup lang="ts">
import { RouterLink } from 'vue-router'

import StatCard from '../../components/StatCard.vue'
import SectionHeader from '../../components/SectionHeader.vue'
import ListCard from '../../components/ListCard.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: overview, loading, error } = useAsyncData(() => api.getPortalOverview())
</script>

<template>
  <div class="grid">
    <div v-if="loading" class="card state-card">正在同步租户概览…</div>
    <div v-else-if="error" class="card state-card state-card--error">{{ error }}</div>
    <template v-else-if="overview">
      <div class="card hero">
      <div>
        <div class="eyebrow">OpenClaw 控制台</div>
        <div class="headline">{{ overview.headline }}</div>
        <p class="muted">
          {{ overview.description }}
        </p>
        <div class="actions">
          <RouterLink class="primary" :to="overview.primaryInstanceId ? `/portal/instances/${overview.primaryInstanceId}` : '/portal/instances'">
            查看主实例
          </RouterLink>
          <RouterLink class="ghost" to="/portal/jobs">查看任务轨迹</RouterLink>
        </div>
      </div>
      <div class="quick-links">
        <RouterLink v-for="link in overview.quickLinks" :key="link.label" :to="link.url" class="q-link card">
          <div class="label">{{ link.label }}</div>
          <div class="muted">{{ link.url }}</div>
        </RouterLink>
      </div>
    </div>

    <div class="stat-grid">
      <StatCard v-for="m in overview.metrics" :key="m.label" v-bind="m" />
    </div>

    <div class="two-col">
      <div class="card">
        <SectionHeader title="任务中心" subtitle="最新 3 个任务" actionLabel="全部任务" />
        <ul class="tasks">
          <li v-for="job in overview.jobs" :key="job.id">
            <div>
              <div class="strong">{{ job.type }}</div>
              <div class="muted">目标：{{ job.target }}</div>
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
                  : '#22c55e',
          }))
        "
      />
    </div>

      <div class="guard-grid">
        <article class="card guard-card">
          <div class="eyebrow">Usage Scope</div>
          <strong>当前租户视角</strong>
          <p class="muted">所有指标默认按当前租户聚合，后续可继续下钻到实例与访问入口。</p>
        </article>
        <article class="card guard-card">
          <div class="eyebrow">Safety Rail</div>
          <strong>{{ overview.alerts.length }} 条近期风险提醒</strong>
          <p class="muted">先把异常实例、备份失败和配置漂移拦在 Portal 层，不必每次都进入 Admin。</p>
        </article>
        <article class="card guard-card">
          <div class="eyebrow">Recovery Goal</div>
          <strong>{{ overview.jobs.length }} 条近期任务轨迹</strong>
          <p class="muted">将备份、配置发布和实例创建放在同一条操作链里，便于追责与恢复。</p>
        </article>
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
  color: #b91c1c;
}

.hero {
  padding: 18px;
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 16px;
  background: linear-gradient(135deg, #ffffff, #eef3ff);
  border: 1px solid var(--stroke);
}

.eyebrow {
  color: var(--brand);
  font-weight: 700;
  font-size: 13px;
}

.headline {
  font-size: 24px;
  font-weight: 800;
  margin: 6px 0 4px;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.primary,
.ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 14px;
  border-radius: 12px;
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
  cursor: pointer;
}

.primary {
  background: linear-gradient(120deg, var(--brand), var(--brand-strong));
  color: #fff;
  border-color: transparent;
}

.quick-links {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 10px;
}

.q-link {
  padding: 12px;
  border: 1px solid var(--stroke);
}

.two-col {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 14px;
}

.guard-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.guard-card {
  padding: 16px;
}

.guard-card strong {
  display: block;
  margin-bottom: 8px;
  font-size: 1.15rem;
}

.card {
  padding: 14px;
}

.tasks {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 6px;
}

.tasks li {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 10px;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.strong {
  font-weight: 600;
}

@media (max-width: 1024px) {
  .hero {
    grid-template-columns: 1fr;
  }
  .two-col {
    grid-template-columns: 1fr;
  }
  .guard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
