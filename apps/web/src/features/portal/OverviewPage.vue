<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import StatCard from '../../components/StatCard.vue'
import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'
import type { PortalArtifactCenterItem, SelfServiceReminder, SelfServiceStep, UsageQuota } from '../../lib/types'

const router = useRouter()
const { data: summary, loading, error, reload } = useAsyncData(() => api.getPortalSelfServiceSummary())

const readinessLabel = computed(() => (summary.value?.onboarding.isReadyForUsage ? '已可直接自助使用' : '仍有待完成引导'))

function openPath(path: string) {
  router.push(path)
}

function openExternalWorkspace() {
  const workspaceUrl = summary.value?.launchpad.workspaceUrl
  if (!workspaceUrl) {
    return
  }
  window.open(workspaceUrl, '_blank', 'noopener,noreferrer')
}

function openSupport() {
  const supportUrl = summary.value?.tenant.supportUrl
  if (!supportUrl) {
    return
  }
  window.open(supportUrl, '_blank', 'noopener,noreferrer')
}

function openArtifactSource(item: PortalArtifactCenterItem) {
  window.open(item.sourceUrl, '_blank', 'noopener,noreferrer')
}

function reminderType(reminder: SelfServiceReminder): 'danger' | 'warning' | 'info' {
  if (reminder.severity === 'critical') return 'danger'
  if (reminder.severity === 'warning') return 'warning'
  return 'info'
}

function stepType(step: SelfServiceStep) {
  if (step.status === 'completed') return 'success'
  if (step.status === 'ready') return 'warning'
  return 'info'
}

function quotaType(quota: UsageQuota) {
  if (quota.status === 'critical') return 'exception'
  if (quota.status === 'warning') return 'warning'
  return 'success'
}
</script>

<template>
  <div class="self-service-shell">
    <el-card v-if="loading" shadow="never" class="state-card">正在整理自助使用闭环…</el-card>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error">
      <template #default>
        <el-button text type="primary" @click="reload">重试</el-button>
      </template>
    </el-alert>
    <template v-else-if="summary">
      <el-card shadow="never" class="hero-card">
        <div class="hero-copy">
          <div class="eyebrow">Portal Self-Service Loop</div>
          <h2>{{ summary.experience.portalHeadline }}</h2>
          <p class="muted">
            {{ summary.experience.portalSubtitle }}
          </p>
          <div class="hero-actions">
            <el-button type="primary" @click="openPath(summary.launchpad.workspacePath)">一键进入工作台</el-button>
            <el-button plain @click="openPath(summary.launchpad.artifactsPath)">打开产物中心</el-button>
            <el-button plain :disabled="!summary.launchpad.workspaceUrl" @click="openExternalWorkspace">外部工作台</el-button>
            <el-button plain :disabled="!summary.tenant.supportUrl" @click="openSupport">支持入口</el-button>
          </div>
        </div>
        <div class="hero-side">
          <div class="hero-status">
            <span class="status-pill" :class="summary.onboarding.isReadyForUsage ? 'ok' : 'pending'">{{ readinessLabel }}</span>
            <span class="status-pill pending">{{ summary.experience.experimentBadge }}</span>
            <span class="muted">已完成 {{ summary.onboarding.completedCount }}/{{ summary.onboarding.totalCount }} 项引导</span>
          </div>
          <div class="hero-meta">
            <div class="meta-item">
              <span class="muted">套餐</span>
              <strong>{{ summary.tenant.plan }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">租户状态</span>
              <strong>{{ summary.tenant.status }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">到期时间</span>
              <strong>{{ summary.tenant.expiredAt || '按实例订阅判断' }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">支持邮箱</span>
              <strong>{{ summary.tenant.supportEmail || '—' }}</strong>
            </div>
          </div>
        </div>
      </el-card>

      <section class="stat-grid">
        <StatCard v-for="metric in summary.metrics" :key="metric.label" v-bind="metric" />
      </section>

      <el-card v-if="summary.onboarding.showGuide" shadow="never" class="guide-card">
        <SectionHeader title="首登引导" subtitle="把开通后的第一次使用动作压缩成可执行检查单" />
        <div class="guide-grid">
          <article v-for="step in summary.onboarding.steps" :key="step.key" class="guide-item">
            <div class="guide-head">
              <strong>{{ step.title }}</strong>
              <el-tag round disable-transitions :type="stepType(step)">{{ step.status }}</el-tag>
            </div>
            <p class="muted">{{ step.description }}</p>
            <div v-if="step.result" class="guide-result">{{ step.result }}</div>
            <el-button class="guide-action" plain @click="openPath(step.actionPath)">{{ step.actionLabel }}</el-button>
          </article>
        </div>
      </el-card>

      <el-card v-else shadow="never" class="guide-card guide-card--done">
        <SectionHeader title="自助闭环" subtitle="当前租户已经完成首登引导，后续可直接复用这条链路" />
        <div class="guide-done">
          <strong>平台侧对话、产物中心、提醒与配额视图均已可直接使用。</strong>
          <span class="muted">建议继续通过产物中心与工作台留痕做日常交付。</span>
        </div>
      </el-card>

      <section class="quota-grid">
        <el-card v-for="quota in summary.quotas" :key="quota.key" shadow="never" class="quota-card">
          <div class="quota-head">
            <div>
              <div class="quota-title">{{ quota.label }}</div>
              <div class="muted">{{ quota.detail }}</div>
            </div>
            <el-tag round disable-transitions :type="quota.status === 'critical' ? 'danger' : quota.status === 'warning' ? 'warning' : 'success'">
              {{ quota.percent }}%
            </el-tag>
          </div>
          <div class="quota-value">{{ quota.usedText }} / {{ quota.limitText }}</div>
          <el-progress :percentage="quota.percent" :status="quotaType(quota)" :stroke-width="10" />
        </el-card>
      </section>

      <section class="main-grid">
        <el-card shadow="never" class="panel-card">
          <SectionHeader title="提醒中心" subtitle="到期、容量与运行风险统一在 Portal 侧提示" />
          <div v-if="summary.reminders.length" class="reminder-list">
            <article v-for="reminder in summary.reminders" :key="reminder.key" class="reminder-item">
              <div class="reminder-head">
                <strong>{{ reminder.title }}</strong>
                <el-tag round disable-transitions :type="reminderType(reminder)">{{ reminder.severity }}</el-tag>
              </div>
              <p class="muted">{{ reminder.description }}</p>
              <div class="reminder-foot">
                <span class="muted">{{ reminder.at || '即时提醒' }}</span>
                <el-button text type="primary" @click="openPath(reminder.actionPath)">{{ reminder.actionLabel }}</el-button>
              </div>
            </article>
          </div>
          <div v-else class="empty-card">当前没有待处理提醒，可以直接继续使用工作台和产物中心。</div>
        </el-card>

        <el-card shadow="never" class="panel-card">
          <SectionHeader title="最近会话" subtitle="继续之前的平台侧会话，而不是重新找入口" />
          <div v-if="summary.recentSessions.length" class="session-list">
            <article v-for="session in summary.recentSessions" :key="session.id" class="session-item">
              <div class="session-head">
                <strong>{{ session.title }}</strong>
                <span class="muted">{{ session.updatedAt }}</span>
              </div>
              <div class="session-meta">
                <span>{{ session.instanceName }}</span>
                <span>{{ session.messageCount }} 条消息</span>
                <span>{{ session.artifactCount }} 个产物</span>
              </div>
              <el-button plain @click="openPath(session.workspacePath)">继续会话</el-button>
            </article>
          </div>
          <div v-else class="empty-card">当前还没有平台侧会话，建议从上方入口直接开启一次对话。</div>
        </el-card>
      </section>

      <el-card shadow="never" class="panel-card">
        <SectionHeader title="最近产物" subtitle="最近归档到平台侧的网页、文档、PPTX、PDF 和表格" />
        <div v-if="summary.recentArtifacts.length" class="artifact-grid">
          <article v-for="artifact in summary.recentArtifacts" :key="artifact.id" class="artifact-card">
            <div class="artifact-top">
              <strong>{{ artifact.title }}</strong>
              <el-tag round disable-transitions>{{ artifact.kind }}</el-tag>
            </div>
            <div class="muted artifact-lines">
              <span>{{ artifact.instanceName }}</span>
              <span>{{ artifact.sessionTitle }}</span>
              <span>{{ artifact.updatedAt }}</span>
            </div>
            <div class="artifact-actions">
              <el-button plain @click="openPath(artifact.workspacePath)">回到工作台</el-button>
              <el-button text type="primary" @click="openArtifactSource(artifact)">打开源地址</el-button>
            </div>
          </article>
        </div>
        <div v-else class="empty-card">当前还没有归档产物，发送一次消息并保存龙虾输出后会出现在这里。</div>
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.self-service-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.state-card,
.panel-card,
.hero-card,
.guide-card,
.quota-card {
  padding: 18px;
}

.hero-card {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 18px;
  background:
    radial-gradient(circle at top left, rgba(29, 107, 255, 0.18), transparent 32%),
    linear-gradient(140deg, rgba(255, 255, 255, 0.98), rgba(237, 243, 255, 0.96));
}

.eyebrow {
  color: var(--brand);
  font-weight: 700;
  font-size: 13px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.hero-copy {
  display: grid;
  gap: 12px;
}

.hero-copy h2 {
  margin: 0;
  font-size: 2rem;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-side {
  display: grid;
  gap: 14px;
}

.hero-status {
  display: grid;
  gap: 8px;
  align-content: start;
}

.status-pill {
  display: inline-flex;
  width: fit-content;
  padding: 8px 12px;
  border-radius: 999px;
  font-weight: 700;
}

.status-pill.ok {
  background: rgba(22, 163, 74, 0.12);
  color: #15803d;
}

.status-pill.pending {
  background: rgba(217, 119, 6, 0.12);
  color: #b45309;
}

.hero-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.meta-item,
.guide-item,
.quota-card,
.reminder-item,
.session-item,
.artifact-card,
.empty-card,
.guide-done {
  padding: 14px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.guide-grid,
.quota-grid,
.artifact-grid {
  display: grid;
  gap: 12px;
}

.guide-grid {
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.guide-item,
.guide-done {
  display: grid;
  gap: 10px;
}

.guide-head,
.quota-head,
.reminder-head,
.session-head,
.artifact-top {
  display: flex;
  justify-content: space-between;
  align-items: start;
  gap: 10px;
}

.guide-result {
  color: var(--text);
  font-weight: 600;
}

.guide-action {
  justify-self: start;
}

.guide-card--done {
  background: linear-gradient(180deg, rgba(240, 253, 244, 0.92), rgba(255, 255, 255, 0.98));
}

.quota-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.quota-card {
  display: grid;
  gap: 14px;
}

.quota-title {
  font-weight: 700;
  margin-bottom: 4px;
}

.quota-value {
  font-size: 1.1rem;
  font-weight: 700;
}

.main-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.reminder-list,
.session-list {
  display: grid;
  gap: 10px;
}

.reminder-item,
.session-item,
.artifact-card {
  display: grid;
  gap: 10px;
}

.reminder-foot,
.artifact-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.session-meta,
.artifact-lines {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 14px;
  color: var(--text-muted);
  font-size: 13px;
}

.artifact-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.empty-card {
  text-align: center;
  color: var(--text-muted);
}

@media (max-width: 1200px) {
  .hero-card,
  .main-grid,
  .quota-grid,
  .artifact-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .hero-meta {
    grid-template-columns: 1fr;
  }

  .hero-actions,
  .artifact-actions,
  .reminder-foot {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
