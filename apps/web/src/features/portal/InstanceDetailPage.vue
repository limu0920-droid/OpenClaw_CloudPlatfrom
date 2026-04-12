<script setup lang="ts">
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/vue'
import { reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { formatAccessEntryType } from '../../lib/access'
import { api } from '../../lib/api'
import type { InstanceOperationsData, PlanOffer, PortalInstanceDetail, TriggerBackupPayload } from '../../lib/types'

const route = useRoute()
const router = useRouter()

const detail = ref<PortalInstanceDetail | null>(null)
const loading = ref(true)
const error = ref('')
const operations = ref<InstanceOperationsData | null>(null)
const plans = ref<PlanOffer[]>([])
const configSubmitting = ref(false)
const backupSubmitting = ref(false)
const powerSubmitting = ref('')
const purchaseSubmitting = ref('')
const feedback = reactive({
  config: '',
  backup: '',
  ops: '',
  purchase: '',
  error: '',
})

const configForm = reactive({
  model: '',
  allowedOrigins: '',
  backupPolicy: '',
})

const backupForm = reactive<TriggerBackupPayload>({
  type: 'manual',
  operator: 'portal-user',
})

const detailTabs = [
  { key: 'overview', label: '概览', subtitle: '入口 / 资源 / 运行动作' },
  { key: 'plans', label: '套餐任务', subtitle: '购买 / 续费 / 最近任务' },
  { key: 'config', label: '配置发布', subtitle: '模型 / 域名 / 策略' },
  { key: 'backup', label: '备份', subtitle: '触发备份 / 历史记录' },
]

async function load(id: string) {
  loading.value = true
  error.value = ''

  try {
    detail.value = await api.getPortalInstanceDetail(id)
    operations.value = await api.getPortalInstanceOperations(id)
    plans.value = await api.getPortalPlans()
    if (detail.value?.config) {
      configForm.model = detail.value.config.settings.model
      configForm.allowedOrigins = detail.value.config.settings.allowedOrigins
      configForm.backupPolicy = detail.value.config.settings.backupPolicy
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载实例失败'
  } finally {
    loading.value = false
  }
}

async function powerAction(action: 'start' | 'stop' | 'restart') {
  if (!detail.value) return
  powerSubmitting.value = action
  feedback.ops = ''
  feedback.error = ''
  try {
    await api.powerPortalInstance(String(detail.value.instance.id), action)
    feedback.ops = `实例${action === 'start' ? '启动' : action === 'stop' ? '停止' : '重启'}已完成`
    await load(String(detail.value.instance.id))
  } catch (err) {
    feedback.error = err instanceof Error ? err.message : '实例操作失败'
  } finally {
    powerSubmitting.value = ''
  }
}

async function purchasePlan(planCode: string, action: 'buy' | 'renew' | 'upgrade' = 'buy') {
  if (!detail.value) return
  purchaseSubmitting.value = `${planCode}-${action}`
  feedback.purchase = ''
  feedback.error = ''
  try {
    await api.createPurchase({
      planCode,
      instanceId: Number(detail.value.instance.id),
      action,
    })
    feedback.purchase = `${planCode} 套餐${action === 'renew' ? '续费' : action === 'upgrade' ? '升级' : '购买'}单已创建`
    await load(String(detail.value.instance.id))
  } catch (err) {
    feedback.error = err instanceof Error ? err.message : '创建订单失败'
  } finally {
    purchaseSubmitting.value = ''
  }
}

async function submitConfig() {
  if (!detail.value) return
  configSubmitting.value = true
  feedback.config = ''
  feedback.error = ''
  try {
    const response = await api.createApprovalRequest('portal', {
      approvalType: 'config_publish',
      targetType: 'instance',
      targetId: Number(detail.value.instance.id),
      instanceId: Number(detail.value.instance.id),
      riskLevel: 'high',
      reason: '申请发布实例配置变更。',
      comment: '由租户侧提交，待平台审批后执行。',
      metadata: {
        updatedBy: 'portal-user',
        model: configForm.model || 'gpt-4o',
        allowedOrigins: configForm.allowedOrigins || '',
        backupPolicy: configForm.backupPolicy || 'daily',
      },
    })
    feedback.config = `配置审批单已提交：${response.approval.approvalNo}`
  } catch (err) {
    feedback.error = err instanceof Error ? err.message : '提交配置审批失败'
  } finally {
    configSubmitting.value = false
  }
}

async function submitBackup() {
  if (!detail.value) return
  backupSubmitting.value = true
  feedback.backup = ''
  feedback.error = ''
  try {
    await api.triggerBackup(String(detail.value.instance.id), backupForm)
    feedback.backup = '备份已触发'
    await load(String(detail.value.instance.id))
  } catch (err) {
    feedback.error = err instanceof Error ? err.message : '触发备份失败'
  } finally {
    backupSubmitting.value = false
  }
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
  <el-card v-if="loading" shadow="never" class="state-card">正在加载实例详情…</el-card>
  <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
  <div v-else-if="detail" class="stack">
    <el-card shadow="never" class="summary-card">
      <div class="summary-head">
        <div>
          <div class="muted">实例</div>
          <div class="title">{{ detail.instance.name }}</div>
          <div class="muted">
            版本 {{ detail.instance.version }} · 计划 {{ detail.instance.plan }} · 区域 {{ detail.instance.region }}
          </div>
        </div>
        <el-tag round size="large" disable-transitions>{{ detail.instance.status }}</el-tag>
      </div>
      <div class="summary-chips">
        <span class="summary-chip">实例编码 {{ detail.instance.code }}</span>
        <span class="summary-chip">套餐 {{ detail.instance.plan }}</span>
        <span class="summary-chip">区域 {{ detail.instance.region }}</span>
        <span class="summary-chip">最近更新 {{ detail.instance.updatedAt }}</span>
      </div>
    </el-card>

    <TabGroup class="detail-tabs">
      <TabList class="detail-tab-list">
        <Tab v-for="item in detailTabs" :key="item.key" as="template" v-slot="{ selected }">
          <button :class="['detail-tab', selected ? 'selected' : '']" type="button">
            <span>{{ item.subtitle }}</span>
            <strong>{{ item.label }}</strong>
          </button>
        </Tab>
      </TabList>

      <TabPanels class="detail-tab-panels">
        <TabPanel class="detail-tab-panel">
          <div class="two-col">
            <el-card shadow="never">
              <SectionHeader title="访问入口" subtitle="来自实例访问配置" />
              <ul class="access">
                <li v-for="acc in detail.instance.access" :key="acc.url">
                  <div class="strong">{{ formatAccessEntryType(acc.entryType) }}</div>
                  <div class="muted">{{ acc.url }}</div>
                  <el-tag v-if="acc.isPrimary" round disable-transitions>主入口</el-tag>
                </li>
              </ul>
            </el-card>

            <el-card shadow="never">
              <SectionHeader title="资源与密码" subtitle="用户侧直接查看 CPU / 内存 / API 调用量 / 管理口令摘要" />
              <div v-if="operations?.runtime" class="runtime-grid">
                <div class="runtime-item">
                  <span class="muted">电源状态</span>
                  <strong>{{ operations.runtime.powerState }}</strong>
                </div>
                <div class="runtime-item">
                  <span class="muted">CPU</span>
                  <strong>{{ operations.runtime.cpuUsagePercent }}%</strong>
                </div>
                <div class="runtime-item">
                  <span class="muted">内存</span>
                  <strong>{{ operations.runtime.memoryUsagePercent }}%</strong>
                </div>
                <div class="runtime-item">
                  <span class="muted">磁盘</span>
                  <strong>{{ operations.runtime.diskUsagePercent }}%</strong>
                </div>
                <div class="runtime-item">
                  <span class="muted">24h API 请求</span>
                  <strong>{{ operations.runtime.apiRequests24h }}</strong>
                </div>
                <div class="runtime-item">
                  <span class="muted">24h Token</span>
                  <strong>{{ operations.runtime.apiTokens24h }}</strong>
                </div>
              </div>
              <div v-if="operations?.credentials" class="credential-box">
                <span class="muted">管理员账号</span>
                <strong>{{ operations.credentials.adminUser }}</strong>
                <span class="muted">
                  密码摘要 {{ operations.credentials.passwordMasked }} · 最近轮换 {{ operations.credentials.lastRotatedAt }}
                </span>
              </div>
              <div class="ops-actions">
                <el-button type="primary" plain @click="router.push(`/portal/instances/${detail.instance.id}/workspace`)">
                  网页版对话
                </el-button>
                <el-button type="primary" :loading="powerSubmitting === 'start'" @click="powerAction('start')">开启</el-button>
                <el-button plain :loading="powerSubmitting === 'stop'" @click="powerAction('stop')">关闭</el-button>
                <el-button plain :loading="powerSubmitting === 'restart'" @click="powerAction('restart')">重启</el-button>
                <el-button plain @click="router.push('/portal/tickets')">问题上报</el-button>
              </div>
              <el-alert v-if="feedback.error" :closable="false" show-icon type="error" :title="feedback.error" />
              <el-alert v-else-if="feedback.ops" :closable="false" show-icon type="success" :title="feedback.ops" />
            </el-card>
          </div>
        </TabPanel>

        <TabPanel class="detail-tab-panel">
          <el-card shadow="never" class="plans-card">
            <SectionHeader title="套餐与购买" subtitle="查看可购套餐，并为当前实例执行购买 / 续费 / 升级" />
            <div class="plans-grid">
              <article v-for="plan in plans" :key="plan.id" class="plan-card">
                <p class="eyebrow">{{ plan.highlight }}</p>
                <strong>{{ plan.name }}</strong>
                <p class="muted">¥{{ plan.monthlyPrice }} / 月 · {{ plan.cpu }} vCPU / {{ plan.memory }} / {{ plan.storage }}</p>
                <ul class="feature-list">
                  <li v-for="feature in plan.features" :key="feature">{{ feature }}</li>
                </ul>
                <div class="plan-actions">
                  <el-button type="primary" :loading="purchaseSubmitting === `${plan.code}-buy`" @click="purchasePlan(plan.code, 'buy')">
                    购买
                  </el-button>
                  <el-button plain :loading="purchaseSubmitting === `${plan.code}-renew`" @click="purchasePlan(plan.code, 'renew')">
                    续费
                  </el-button>
                </div>
              </article>
            </div>
            <el-alert v-if="feedback.error" :closable="false" show-icon type="error" :title="feedback.error" />
            <el-alert v-else-if="feedback.purchase" :closable="false" show-icon type="success" :title="feedback.purchase" />
          </el-card>

          <el-card shadow="never">
            <SectionHeader title="任务" subtitle="最近 3 条" />
            <ul class="tasks">
              <li v-for="job in detail.jobs.slice(0, 3)" :key="job.id">
                <div>
                  <div class="strong">{{ job.type }}</div>
                  <div class="muted">{{ job.target }}</div>
                </div>
                <el-tag round disable-transitions>{{ job.status }}</el-tag>
                <div class="muted">{{ job.startedAt }}</div>
              </li>
            </ul>
          </el-card>
        </TabPanel>

        <TabPanel class="detail-tab-panel">
          <el-card shadow="never" class="config-card">
            <SectionHeader title="配置发布" subtitle="编辑模型、域名与备份策略后提交审批" />
            <div class="config-grid">
              <label class="config-item">
                <span class="muted">模型</span>
                <el-input v-model="configForm.model" placeholder="例如 gpt-4o" />
              </label>
              <label class="config-item">
                <span class="muted">允许域名（逗号分隔）</span>
                <el-input v-model="configForm.allowedOrigins" placeholder="https://example.com,https://a.com" />
              </label>
              <label class="config-item">
                <span class="muted">备份策略</span>
                <el-input v-model="configForm.backupPolicy" placeholder="daily / weekly" />
              </label>
              <div class="config-item config-actions">
                <el-button type="primary" :loading="configSubmitting" @click="submitConfig">提交审批</el-button>
                <div class="muted small">
                  当前配置版本：{{ detail.config ? `v${detail.config.version}` : '暂无' }} · 最近发布：
                  {{ detail.config?.publishedAt || '—' }}
                </div>
              </div>
            </div>
            <el-alert v-if="feedback.error" :closable="false" show-icon type="error" :title="feedback.error" />
            <el-alert v-else-if="feedback.config" :closable="false" show-icon type="success" :title="feedback.config" />
          </el-card>
        </TabPanel>

        <TabPanel class="detail-tab-panel">
          <el-card shadow="never">
            <SectionHeader title="备份" subtitle="最近 3 个" actionLabel="查看更多" />
            <div class="backup-row">
              <label>
                <span class="muted">类型</span>
                <el-select v-model="backupForm.type">
                  <el-option label="manual" value="manual" />
                  <el-option label="scheduled" value="scheduled" />
                </el-select>
              </label>
              <label>
                <span class="muted">操作人</span>
                <el-input v-model="backupForm.operator" />
              </label>
              <el-button type="primary" :loading="backupSubmitting" @click="submitBackup">触发备份</el-button>
            </div>
            <el-alert v-if="feedback.error" :closable="false" show-icon type="error" :title="feedback.error" />
            <el-alert v-else-if="feedback.backup" :closable="false" show-icon type="success" :title="feedback.backup" />
            <el-table :data="detail.backups" class="surface-table">
              <el-table-column prop="name" label="名称" min-width="220" />
              <el-table-column label="状态" min-width="120">
                <template #default="{ row }">
                  <el-tag round disable-transitions>{{ row.status }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="size" label="大小" min-width="120" />
              <el-table-column prop="createdAt" label="时间" min-width="180" />
            </el-table>
          </el-card>
        </TabPanel>
      </TabPanels>
    </TabGroup>
  </div>
</template>

<style scoped>
.stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.state-card {
  padding: 24px;
  text-align: center;
}

.summary-card {
  overflow: hidden;
}

.summary-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.summary-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 16px;
}

.summary-chip {
  padding: 10px 14px;
  border-radius: 999px;
  background: var(--panel-muted);
  color: var(--text);
  font-size: 13px;
}

.title {
  font-size: 20px;
  font-weight: 700;
}

.detail-tabs {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.detail-tab-list {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.detail-tab {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  padding: 16px 18px;
  text-align: left;
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(255, 255, 255, 0.78);
}

.detail-tab span {
  font-size: 12px;
  color: var(--text-muted);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.detail-tab strong {
  font-size: 1.05rem;
  font-variation-settings: "wght" 650;
}

.detail-tab.selected {
  border-color: rgba(29, 107, 255, 0.24);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(239, 244, 255, 0.98));
  box-shadow: 0 18px 40px rgba(29, 107, 255, 0.08);
}

.detail-tab-panels,
.detail-tab-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.two-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.config-card,
.plans-card {
  padding: 14px;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.runtime-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.runtime-item,
.credential-box,
.plan-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.credential-box {
  margin-top: 10px;
}

.ops-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 12px;
}

.plans-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.feature-list {
  margin: 0;
  padding-left: 18px;
  color: var(--text-muted);
}

.plan-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.config-actions {
  gap: 10px;
}

.backup-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin: 12px 0;
  align-items: end;
}

.access,
.tasks {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 6px;
}

.access li,
.tasks li {
  display: grid;
  grid-template-columns: 1.2fr 1.2fr auto;
  gap: 10px;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

@media (max-width: 1024px) {
  .detail-tab-list,
  .two-col {
    grid-template-columns: 1fr;
  }

  .config-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .runtime-grid,
  .plans-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .config-grid {
    grid-template-columns: 1fr;
  }
}
</style>
