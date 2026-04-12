<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useBranding } from '../../lib/brand'
import type { AccountSettings, PortalOpsReport } from '../../lib/types'

const { brand } = useBranding()

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const report = ref<PortalOpsReport | null>(null)
const form = reactive<AccountSettings>({
  tenantId: '',
  primaryEmail: '',
  billingEmail: '',
  alertEmail: '',
  preferredLocale: 'zh-CN',
  secondaryLocale: 'en-US',
  timezone: 'Asia/Shanghai',
  emailVerified: false,
  marketingOptIn: false,
  notifyOnAlert: true,
  notifyOnPayment: true,
  notifyOnExpiry: true,
  notifyChannelEmail: true,
  notifyChannelWebhook: false,
  notifyChannelInApp: true,
  notificationWebhookUrl: '',
  portalHeadline: '',
  portalSubtitle: '',
  workspaceCallout: '',
  experimentBadge: '',
  updatedAt: '',
})

function applyForm(settings: AccountSettings | null) {
  if (!settings) return
  Object.assign(form, settings)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    report.value = await api.getPortalOpsReport()
    applyForm(report.value.settings)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载运营中心失败'
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const settings = await api.updatePortalAccountSettings(form)
    applyForm(settings)
    await load()
    ElMessage.success('运营设置已更新')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '保存运营设置失败')
  } finally {
    saving.value = false
  }
}

function exportCSV() {
  if (!report.value?.export.csvPath) return
  window.open(report.value.export.csvPath, '_blank', 'noopener,noreferrer')
}

void load()
</script>

<template>
  <div class="operations-shell">
    <el-card v-if="loading" shadow="never" class="state-card">正在整理运营中心…</el-card>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
    <template v-else-if="report">
      <el-card shadow="never" class="hero-card">
        <div class="hero-copy">
          <div class="eyebrow">{{ brand.name }} Operations</div>
          <h2>{{ form.portalHeadline || `${brand.name} 运营中心` }}</h2>
          <p class="muted">{{ form.portalSubtitle || '统一查看通知渠道、引导文案、使用趋势与导出能力。' }}</p>
        </div>
        <div class="hero-actions">
          <el-button type="primary" @click="save" :loading="saving">保存运营设置</el-button>
          <el-button plain @click="exportCSV">导出 CSV</el-button>
        </div>
      </el-card>

      <section class="metric-grid">
        <el-card shadow="never" class="metric-card">
          <span class="muted">运行实例</span>
          <strong>{{ report.summary.instanceCount }}</strong>
        </el-card>
        <el-card shadow="never" class="metric-card">
          <span class="muted">平台会话</span>
          <strong>{{ report.summary.sessionCount }}</strong>
        </el-card>
        <el-card shadow="never" class="metric-card">
          <span class="muted">归档产物</span>
          <strong>{{ report.summary.artifactCount }}</strong>
        </el-card>
        <el-card shadow="never" class="metric-card">
          <span class="muted">开放工单</span>
          <strong>{{ report.summary.openTicketCount }}</strong>
        </el-card>
      </section>

      <section class="panel-grid">
        <el-card shadow="never" class="panel-card">
          <SectionHeader title="通知渠道" subtitle="按租户控制邮件、Webhook 与站内提醒开关" />
          <div class="form-grid">
            <el-switch v-model="form.notifyChannelEmail" active-text="邮件通知" />
            <el-switch v-model="form.notifyChannelWebhook" active-text="Webhook 通知" />
            <el-switch v-model="form.notifyChannelInApp" active-text="站内提醒" />
            <el-input v-model="form.primaryEmail" placeholder="主邮箱" />
            <el-input v-model="form.alertEmail" placeholder="告警邮箱" />
            <el-input v-model="form.billingEmail" placeholder="账单邮箱" />
            <el-input v-model="form.notificationWebhookUrl" placeholder="Webhook URL" />
          </div>
          <div class="channel-list">
            <article v-for="channel in report.notificationChannels" :key="channel.key" class="channel-item">
              <strong>{{ channel.label }}</strong>
              <span class="muted">{{ channel.description }}</span>
              <small class="muted">{{ channel.target || '未配置目标' }}</small>
            </article>
          </div>
        </el-card>

        <el-card shadow="never" class="panel-card">
          <SectionHeader title="引导文案" subtitle="为品牌门户配置 Portal 标题、副标题和实验标签" />
          <div class="form-grid">
            <el-input v-model="form.portalHeadline" placeholder="Portal 标题" />
            <el-input v-model="form.portalSubtitle" type="textarea" :autosize="{ minRows: 3, maxRows: 5 }" placeholder="Portal 副标题" />
            <el-input v-model="form.workspaceCallout" type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" placeholder="工作台引导语" />
            <el-input v-model="form.experimentBadge" placeholder="实验标签，例如 OEM Launch / Growth Beta" />
          </div>
        </el-card>
      </section>

      <section class="panel-grid">
        <el-card shadow="never" class="panel-card">
          <SectionHeader title="通知模板预览" subtitle="当前品牌下的支付、到期和运行提醒模板预览" />
          <div class="template-list">
            <article v-for="template in report.notificationTemplates" :key="template.key" class="template-item">
              <strong>{{ template.title }}</strong>
              <span class="muted">{{ template.subject }}</span>
              <p>{{ template.body }}</p>
            </article>
          </div>
        </el-card>

        <el-card shadow="never" class="panel-card">
          <SectionHeader title="月度趋势" subtitle="按账单月聚合 charge、sessions、messages 与 artifacts" />
          <el-table :data="report.monthlyUsage" stripe>
            <el-table-column prop="label" label="月份" min-width="120" />
            <el-table-column prop="chargeAmount" label="费用" min-width="100" />
            <el-table-column prop="sessions" label="会话" min-width="90" />
            <el-table-column prop="messages" label="消息" min-width="90" />
            <el-table-column prop="artifacts" label="产物" min-width="90" />
          </el-table>
        </el-card>
      </section>
    </template>
  </div>
</template>

<style scoped>
.operations-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.hero-card,
.panel-card,
.metric-card,
.state-card {
  padding: 18px;
}

.hero-card {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
}

.hero-copy {
  display: grid;
  gap: 8px;
}

.hero-copy h2 {
  margin: 0;
  font-size: 2rem;
}

.hero-actions,
.metric-grid,
.panel-grid,
.form-grid,
.channel-list,
.template-list {
  display: grid;
  gap: 12px;
}

.metric-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.metric-card strong {
  display: block;
  margin-top: 8px;
  font-size: 2rem;
}

.panel-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.form-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.channel-item,
.template-item {
  display: grid;
  gap: 6px;
  padding: 14px;
  border-radius: var(--radius-md);
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
}

.template-item p {
  margin: 0;
  white-space: pre-wrap;
}

@media (max-width: 1080px) {
  .metric-grid,
  .panel-grid,
  .form-grid {
    grid-template-columns: 1fr;
  }

  .hero-card {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
