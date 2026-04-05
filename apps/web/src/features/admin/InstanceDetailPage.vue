<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { AdminInstanceDetail } from '../../lib/types'

type DetailTab = 'overview' | 'monitoring' | 'logs' | 'config' | 'audit'

const route = useRoute()
const router = useRouter()

const detail = ref<AdminInstanceDetail | null>(null)
const loading = ref(true)
const error = ref('')

const tabs: Array<{ key: DetailTab; label: string; description: string }> = [
  { key: 'overview', label: '概览', description: '实例基础信息、入口与近期动作' },
  { key: 'monitoring', label: '监控', description: '资源趋势与近期告警' },
  { key: 'logs', label: '日志', description: '运行日志与实例级事件线索' },
  { key: 'config', label: '配置', description: '当前生效配置与访问策略' },
  { key: 'audit', label: '审计', description: '任务、审计事件与操作痕迹' },
]

const activeTab = computed<DetailTab>(() => {
  const tab = route.query.tab
  if (tab === 'monitoring' || tab === 'logs' || tab === 'config' || tab === 'audit') {
    return tab
  }
  return 'overview'
})

const latestPoint = computed(() => detail.value?.resourceTrend.at(-1))

async function load(id: string) {
  loading.value = true
  error.value = ''

  try {
    detail.value = await api.getAdminInstanceDetail(id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载实例详情失败'
  } finally {
    loading.value = false
  }
}

function setTab(tab: DetailTab) {
  void router.replace({
    query: {
      ...route.query,
      tab,
    },
  })
}

watch(
  () => String(route.params.id),
  (id) => {
    void load(id)
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="loading" class="card state-card">正在同步实例控制面详情…</div>
  <div v-else-if="error" class="card state-card state-card--error">{{ error }}</div>
  <div v-else-if="detail" class="admin-detail">
    <section class="card hero">
      <div>
        <div class="eyebrow">Admin Instance Detail</div>
        <h2>{{ detail.instance.name }}</h2>
        <p class="muted">
          {{ detail.tenant?.name || '未绑定租户' }} · {{ detail.cluster?.name || '未分配集群' }} ·
          {{ detail.instance.region }} · {{ detail.instance.spec || '—' }}
        </p>
      </div>
      <div class="hero-meta">
        <div class="pill">{{ detail.instance.status }}</div>
        <div class="muted">运行时：{{ detail.instance.runtimeType || '—' }}</div>
        <div class="muted">版本：{{ detail.instance.version }}</div>
      </div>
    </section>

    <section class="tab-strip card">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        :class="['tab-chip', { 'tab-chip--active': activeTab === tab.key }]"
        @click="setTab(tab.key)"
      >
        <strong>{{ tab.label }}</strong>
        <span>{{ tab.description }}</span>
      </button>
    </section>

    <template v-if="activeTab === 'overview'">
      <section class="stat-grid">
        <article class="card mini-stat">
          <span class="muted">CPU 负载</span>
          <strong>{{ latestPoint?.cpu ?? 0 }}%</strong>
          <small>最近窗口</small>
        </article>
        <article class="card mini-stat">
          <span class="muted">内存水位</span>
          <strong>{{ latestPoint?.memory ?? 0 }}%</strong>
          <small>最近窗口</small>
        </article>
        <article class="card mini-stat">
          <span class="muted">请求量</span>
          <strong>{{ latestPoint?.requests ?? 0 }}</strong>
          <small>最近窗口</small>
        </article>
        <article class="card mini-stat">
          <span class="muted">打开告警</span>
          <strong>{{ detail.alerts.filter((item) => item.severity === 'critical' || item.severity === 'warning').length }}</strong>
          <small>需重点关注</small>
        </article>
      </section>

      <section class="detail-grid">
        <div class="card panel">
          <SectionHeader title="访问入口" subtitle="面向租户与管理员的访问落点" />
          <div class="stack-list">
            <div v-for="entry in detail.instance.access" :key="entry.url" class="stack-item">
              <strong>{{ entry.entryType }}</strong>
              <span class="muted">{{ entry.url }}</span>
            </div>
          </div>
        </div>

        <div class="card panel">
          <SectionHeader title="实例归属" subtitle="租户、集群与区域落点" />
          <div class="config-grid">
            <div class="config-item">
              <span class="muted">租户</span>
              <strong>{{ detail.tenant?.name || '—' }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">集群</span>
              <strong>{{ detail.cluster?.name || '—' }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">区域</span>
              <strong>{{ detail.instance.region }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">资源规格</span>
              <strong>{{ detail.instance.spec || '—' }}</strong>
            </div>
          </div>
        </div>

        <div class="card panel span-two">
          <SectionHeader title="近期动作" subtitle="创建、发布与备份动作摘要" />
          <div class="overview-activity">
            <div class="stack-list">
              <div v-for="job in detail.jobs.slice(0, 4)" :key="job.id" class="stack-item">
                <strong>{{ job.type }}</strong>
                <span class="muted">{{ job.status }} · {{ job.startedAt }}</span>
              </div>
            </div>
            <div class="stack-list">
              <div v-for="backup in detail.backups.slice(0, 3)" :key="backup.id" class="stack-item">
                <strong>{{ backup.name }}</strong>
                <span class="muted">{{ backup.status }} · {{ backup.size }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="activeTab === 'monitoring'">
      <section class="detail-grid">
        <div class="card panel span-two">
          <SectionHeader title="资源趋势" subtitle="CPU、内存与请求量按时间窗口查看" />
          <div class="trend-list">
            <div v-for="point in detail.resourceTrend" :key="point.label" class="trend-row">
              <span>{{ point.label }}</span>
              <div class="trend-bars">
                <div class="trend-bar">
                  <i :style="{ width: `${point.cpu}%` }" />
                </div>
                <div class="trend-bar trend-bar--memory">
                  <i :style="{ width: `${point.memory}%` }" />
                </div>
              </div>
              <span class="muted">{{ point.requests }} req</span>
            </div>
          </div>
        </div>

        <div class="card panel">
          <SectionHeader title="近期告警" subtitle="实例级风险提醒" />
          <div class="stack-list">
            <div v-for="alert in detail.alerts" :key="alert.id" class="stack-item">
              <strong>{{ alert.title }}</strong>
              <span class="muted">{{ alert.severity }} · {{ alert.time }}</span>
            </div>
          </div>
        </div>

        <div class="card panel">
          <SectionHeader title="监控摘要" subtitle="运维视角的快速判断" />
          <div class="config-grid">
            <div class="config-item">
              <span class="muted">最近 CPU 峰值</span>
              <strong>{{ Math.max(...detail.resourceTrend.map((point) => point.cpu)) }}%</strong>
            </div>
            <div class="config-item">
              <span class="muted">最近内存峰值</span>
              <strong>{{ Math.max(...detail.resourceTrend.map((point) => point.memory)) }}%</strong>
            </div>
            <div class="config-item">
              <span class="muted">最近请求峰值</span>
              <strong>{{ Math.max(...detail.resourceTrend.map((point) => point.requests)) }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">告警数</span>
              <strong>{{ detail.alerts.length }}</strong>
            </div>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="activeTab === 'logs'">
      <section class="detail-grid">
        <div class="card panel span-two">
          <SectionHeader title="运行日志" subtitle="当前以实例级 Mock 日志模拟近期关键事件" />
          <div class="log-list">
            <div v-for="log in detail.runtimeLogs" :key="log.id" class="log-row">
              <span :class="['log-level', `log-level--${log.level}`]">{{ log.level }}</span>
              <strong>{{ log.source }}</strong>
              <span class="muted">{{ log.timestamp }}</span>
              <p>{{ log.message }}</p>
            </div>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="activeTab === 'config'">
      <section class="detail-grid">
        <div class="card panel">
          <SectionHeader title="当前配置" subtitle="当前生效版本与关键护栏" />
          <div v-if="detail.config" class="config-grid">
            <div class="config-item">
              <span class="muted">版本</span>
              <strong>v{{ detail.config.version }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">发布人</span>
              <strong>{{ detail.config.updatedBy }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">模型</span>
              <strong>{{ detail.config.settings.model }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">Hash</span>
              <strong>{{ detail.config.hash }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">允许来源</span>
              <strong>{{ detail.config.settings.allowedOrigins }}</strong>
            </div>
            <div class="config-item">
              <span class="muted">备份策略</span>
              <strong>{{ detail.config.settings.backupPolicy }}</strong>
            </div>
          </div>
          <div v-else class="muted">暂无配置记录。</div>
        </div>

        <div class="card panel">
          <SectionHeader title="访问策略" subtitle="入口与访问方式" />
          <div class="stack-list">
            <div v-for="entry in detail.instance.access" :key="entry.url" class="stack-item">
              <strong>{{ entry.entryType }}</strong>
              <span class="muted">{{ entry.url }}</span>
              <small class="muted">mode: {{ entry.accessMode || '—' }}</small>
            </div>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="activeTab === 'audit'">
      <section class="detail-grid">
        <div class="card panel">
          <SectionHeader title="审计事件" subtitle="高风险与关键操作留痕" />
          <div class="stack-list">
            <div v-for="audit in detail.audits" :key="audit.id" class="stack-item">
              <strong>{{ audit.actor }} · {{ audit.action }}</strong>
              <span class="muted">{{ audit.result }} · {{ audit.time }}</span>
              <small v-if="audit.note" class="muted">{{ audit.note }}</small>
            </div>
          </div>
        </div>

        <div class="card panel">
          <SectionHeader title="任务轨迹" subtitle="异步动作的结果视图" />
          <div class="stack-list">
            <div v-for="job in detail.jobs" :key="job.id" class="stack-item">
              <strong>{{ job.type }}</strong>
              <span class="muted">{{ job.status }} · {{ job.startedAt }}</span>
              <small v-if="job.progress" class="muted">progress: {{ job.progress }}%</small>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.admin-detail {
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

.hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
  background: linear-gradient(135deg, rgba(30, 64, 175, 0.28), rgba(14, 165, 233, 0.12));
}

.hero h2 {
  margin: 6px 0;
  font-size: 1.8rem;
}

.hero-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.tab-strip {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 10px;
  padding: 10px;
}

.tab-chip {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px;
  border-radius: 16px;
  border: 1px solid transparent;
  background: rgba(255, 255, 255, 0.03);
  color: var(--text);
  text-align: left;
}

.tab-chip span {
  color: var(--text-muted);
  font-size: 0.82rem;
}

.tab-chip--active {
  border-color: rgba(96, 165, 250, 0.45);
  background: rgba(59, 130, 246, 0.14);
}

.mini-stat {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.mini-stat strong {
  font-size: 1.9rem;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.panel {
  padding: 14px;
}

.span-two {
  grid-column: span 2;
}

.trend-list,
.stack-list,
.log-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.trend-row,
.stack-item,
.log-row {
  display: grid;
  gap: 10px;
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
}

.trend-row {
  grid-template-columns: 64px 1fr 88px;
  align-items: center;
}

.trend-bars {
  display: grid;
  gap: 8px;
}

.trend-bar {
  overflow: hidden;
  height: 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.06);
}

.trend-bar i {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
}

.trend-bar--memory i {
  background: linear-gradient(90deg, #22d3ee, #38bdf8);
}

.config-grid,
.overview-activity {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.log-row {
  grid-template-columns: 74px 120px 120px 1fr;
  align-items: start;
}

.log-row p {
  margin: 0;
}

.log-level {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
}

.log-level--info {
  background: rgba(59, 130, 246, 0.14);
}

.log-level--warning {
  background: rgba(245, 158, 11, 0.18);
}

.log-level--error {
  background: rgba(239, 68, 68, 0.18);
}

@media (max-width: 1180px) {
  .tab-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 1024px) {
  .detail-grid,
  .config-grid,
  .overview-activity {
    grid-template-columns: 1fr;
  }

  .span-two {
    grid-column: span 1;
  }

  .hero {
    flex-direction: column;
  }

  .hero-meta {
    align-items: flex-start;
  }

  .log-row,
  .trend-row {
    grid-template-columns: 1fr;
  }
}
</style>
