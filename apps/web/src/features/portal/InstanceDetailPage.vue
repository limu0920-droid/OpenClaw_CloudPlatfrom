<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { InstanceOperationsData, PlanOffer, PortalInstanceDetail, TriggerBackupPayload, UpdateConfigPayload } from '../../lib/types'

const route = useRoute()

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
  purchaseSubmitting.value = planCode
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
    await api.updateInstanceConfig(String(detail.value.instance.id), {
      updatedBy: 'portal-user',
      settings: {
        model: configForm.model || 'gpt-4o',
        allowedOrigins: configForm.allowedOrigins || '',
        backupPolicy: configForm.backupPolicy || 'daily',
      },
    } as UpdateConfigPayload)
    feedback.config = '配置发布已提交'
    await load(String(detail.value.instance.id))
  } catch (err) {
    feedback.error = err instanceof Error ? err.message : '发布配置失败'
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
  <div v-if="loading" class="card state-card">正在加载实例详情…</div>
  <div v-else-if="error" class="card state-card state-card--error">{{ error }}</div>
  <div v-else-if="detail" class="stack">
    <div class="card head">
      <div>
        <div class="muted">实例</div>
        <div class="title">{{ detail.instance.name }}</div>
        <div class="muted">
          版本 {{ detail.instance.version }} · 计划 {{ detail.instance.plan }} · 区域 {{ detail.instance.region }}
        </div>
      </div>
      <div class="pill">{{ detail.instance.status }}</div>
    </div>

    <div class="two-col">
      <div class="card">
        <SectionHeader title="访问入口" subtitle="来自实例访问配置" />
        <ul class="access">
          <li v-for="acc in detail.instance.access" :key="acc.url">
            <div class="strong">{{ acc.entryType }}</div>
            <div class="muted">{{ acc.url }}</div>
            <div v-if="acc.isPrimary" class="pill">主入口</div>
          </li>
        </ul>
      </div>
      <div class="card">
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
          <span class="muted">密码摘要 {{ operations.credentials.passwordMasked }} · 最近轮换 {{ operations.credentials.lastRotatedAt }}</span>
        </div>
        <div class="ops-actions">
          <button class="primary" :disabled="powerSubmitting === 'start'" @click="powerAction('start')">
            {{ powerSubmitting === 'start' ? '启动中…' : '开启' }}
          </button>
          <button class="ghost" :disabled="powerSubmitting === 'stop'" @click="powerAction('stop')">
            {{ powerSubmitting === 'stop' ? '停止中…' : '关闭' }}
          </button>
          <button class="ghost" :disabled="powerSubmitting === 'restart'" @click="powerAction('restart')">
            {{ powerSubmitting === 'restart' ? '重启中…' : '重启' }}
          </button>
          <RouterLink class="ghost" to="/portal/tickets">问题上报</RouterLink>
        </div>
        <div class="form-feedback">
          <span v-if="feedback.error" class="error">{{ feedback.error }}</span>
          <span v-else-if="feedback.ops" class="success">{{ feedback.ops }}</span>
        </div>
      </div>
    </div>

    <div class="card plans-card">
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
            <button class="primary" :disabled="purchaseSubmitting === plan.code" @click="purchasePlan(plan.code, 'buy')">
              {{ purchaseSubmitting === plan.code ? '处理中…' : '购买' }}
            </button>
            <button class="ghost" :disabled="purchaseSubmitting === `${plan.code}-renew`" @click="purchasePlan(plan.code, 'renew')">续费</button>
          </div>
        </article>
      </div>
      <div class="form-feedback">
        <span v-if="feedback.error" class="error">{{ feedback.error }}</span>
        <span v-else-if="feedback.purchase" class="success">{{ feedback.purchase }}</span>
      </div>
    </div>

    <div class="two-col">
      <div class="card">
        <SectionHeader title="任务" subtitle="最近 3 条" />
        <ul class="tasks">
          <li v-for="job in detail.jobs.slice(0, 3)" :key="job.id">
            <div>
              <div class="strong">{{ job.type }}</div>
              <div class="muted">{{ job.target }}</div>
            </div>
            <div class="pill">{{ job.status }}</div>
            <div class="muted">{{ job.startedAt }}</div>
          </li>
        </ul>
      </div>
    </div>

    <div class="card config-card">
      <SectionHeader title="配置发布" subtitle="编辑模型、域名与备份策略后提交发布" />
      <div class="config-grid">
        <label class="config-item">
          <span class="muted">模型</span>
          <input v-model="configForm.model" placeholder="例如 gpt-4o" />
        </label>
        <label class="config-item">
          <span class="muted">允许域名（逗号分隔）</span>
          <input v-model="configForm.allowedOrigins" placeholder="https://example.com,https://a.com" />
        </label>
        <label class="config-item">
          <span class="muted">备份策略</span>
          <input v-model="configForm.backupPolicy" placeholder="daily / weekly" />
        </label>
        <div class="config-item config-actions">
          <button class="primary" :disabled="configSubmitting" @click="submitConfig">
            {{ configSubmitting ? '发布中…' : '发布配置' }}
          </button>
          <div class="muted small">
            当前配置版本：{{ detail.config ? `v${detail.config.version}` : '暂无' }} · 最近发布：
            {{ detail.config?.publishedAt || '—' }}
          </div>
        </div>
      </div>
      <div class="form-feedback">
        <span v-if="feedback.error" class="error">{{ feedback.error }}</span>
        <span v-else-if="feedback.config" class="success">{{ feedback.config }}</span>
      </div>
    </div>

    <div class="card">
      <SectionHeader title="备份" subtitle="最近 3 个" actionLabel="查看更多" />
      <div class="backup-row">
        <label>
          <span class="muted">类型</span>
          <select v-model="backupForm.type">
            <option value="manual">manual</option>
            <option value="scheduled">scheduled</option>
          </select>
        </label>
        <label>
          <span class="muted">操作人</span>
          <input v-model="backupForm.operator" />
        </label>
        <button class="primary" :disabled="backupSubmitting" @click="submitBackup">
          {{ backupSubmitting ? '触发中…' : '触发备份' }}
        </button>
      </div>
      <div class="form-feedback">
        <span v-if="feedback.error" class="error">{{ feedback.error }}</span>
        <span v-else-if="feedback.backup" class="success">{{ feedback.backup }}</span>
      </div>
      <div class="table">
        <div class="head">
          <span>名称</span>
          <span>状态</span>
          <span>大小</span>
          <span>时间</span>
        </div>
        <div v-for="bk in detail.backups" :key="bk.id" class="row">
          <span class="strong">{{ bk.name }}</span>
          <span class="pill">{{ bk.status }}</span>
          <span>{{ bk.size }}</span>
          <span class="muted">{{ bk.createdAt }}</span>
        </div>
      </div>
    </div>
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

.state-card--error {
  color: #b91c1c;
}

.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
}
.title {
  font-size: 20px;
  font-weight: 700;
}
.two-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.config-card {
  padding: 14px;
}

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

.config-item input,
.backup-row input,
.backup-row select {
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--stroke);
}

.config-actions {
  gap: 10px;
}

.backup-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin: 12px 0;
}

.form-feedback {
  min-height: 20px;
  font-size: 13px;
}

.error {
  color: #b91c1c;
}

.success {
  color: #15803d;
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
.table {
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
.head,
.row {
  display: grid;
  grid-template-columns: 1.4fr 1fr 1fr 1.4fr;
  padding: 12px 14px;
}
.head {
  background: var(--panel-muted);
  color: var(--text-muted);
  font-weight: 600;
}
.row {
  border-top: 1px solid var(--stroke);
  align-items: center;
}
@media (max-width: 1024px) {
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
</style>
