<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'
import type { CreateInstancePayload } from '../../lib/types'

const filter = ref('')
const submitLoading = ref(false)
const submitError = ref('')
const submitSuccess = ref('')

const form = ref<CreateInstancePayload>({
  name: '',
  plan: 'pro',
  region: 'cn-shanghai',
  cpu: '2',
  memory: '4Gi',
})

const { data: instances, loading, error, reload } = useAsyncData(() => api.getPortalInstances())
const filtered = computed(() =>
  (instances.value ?? []).filter((item) => item.name.toLowerCase().includes(filter.value.toLowerCase())),
)

async function handleCreate() {
  submitError.value = ''
  submitSuccess.value = ''
  submitLoading.value = true
  try {
    await api.createPortalInstance(form.value)
    submitSuccess.value = '实例创建请求已提交'
    await reload()
    form.value = { name: '', plan: 'pro', region: 'cn-shanghai', cpu: '2', memory: '4Gi' }
  } catch (err) {
    submitError.value = err instanceof Error ? err.message : '创建失败'
  } finally {
    submitLoading.value = false
  }
}
</script>

<template>
  <div class="card">
    <SectionHeader title="实例列表" subtitle="当前已通过 Go Mock API 拉取数据" />
    <div class="create-form">
      <div class="form-row">
        <label>
          <span>名称</span>
          <input v-model="form.name" placeholder="例如：生产环境实例" />
        </label>
        <label>
          <span>套餐</span>
          <select v-model="form.plan">
            <option value="pro">Pro</option>
            <option value="standard">Standard</option>
            <option value="trial">Trial</option>
          </select>
        </label>
        <label>
          <span>地域</span>
          <select v-model="form.region">
            <option value="cn-shanghai">cn-shanghai</option>
            <option value="cn-beijing">cn-beijing</option>
            <option value="ap-southeast-1">ap-southeast-1</option>
          </select>
        </label>
      </div>
      <div class="form-row">
        <label>
          <span>CPU (vCPU)</span>
          <input v-model="form.cpu" />
        </label>
        <label>
          <span>内存</span>
          <input v-model="form.memory" placeholder="例如 4Gi" />
        </label>
        <button class="primary" :disabled="submitLoading" @click="handleCreate">
          {{ submitLoading ? '创建中…' : '创建实例' }}
        </button>
      </div>
      <div class="form-feedback">
        <span v-if="submitError" class="error">{{ submitError }}</span>
        <span v-else-if="submitSuccess" class="success">{{ submitSuccess }}</span>
      </div>
    </div>
    <div class="toolbar">
      <input v-model="filter" placeholder="搜索实例..." />
      <RouterLink class="primary" to="/portal/jobs">查看变更任务</RouterLink>
    </div>
    <div v-if="loading" class="state-card">正在加载实例列表…</div>
    <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
    <div v-else class="table">
      <div class="head">
        <span>名称</span>
        <span>状态</span>
        <span>版本</span>
        <span>地域</span>
        <span>更新</span>
        <span>操作</span>
      </div>
      <div v-for="inst in filtered" :key="inst.id" class="row">
        <div class="name">
          <div class="strong">{{ inst.name }}</div>
          <div class="muted">{{ inst.plan }} · {{ inst.code }}</div>
        </div>
        <div><span class="pill">{{ inst.status }}</span></div>
        <div>{{ inst.version }}</div>
        <div>{{ inst.region }}</div>
        <div class="muted">{{ inst.updatedAt }}</div>
        <div class="actions">
          <RouterLink :to="`/portal/instances/${inst.id}`">详情</RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
}

.state-card {
  padding: 18px;
  border: 1px dashed var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel-muted);
  text-align: center;
}

.state-card--error {
  color: #b91c1c;
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 12px;
  padding: 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel-muted);
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 10px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: var(--text-muted);
}

input {
  flex: 1;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
}

.create-form input,
.create-form select {
  background: #fff;
}

.primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 14px;
  border-radius: 12px;
  border: 1px solid transparent;
  background: linear-gradient(120deg, var(--brand), var(--brand-strong));
  color: #fff;
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

.table {
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.head,
.row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 1.2fr 1fr;
  padding: 12px 14px;
  align-items: center;
}

.head {
  background: var(--panel-muted);
  color: var(--text-muted);
  font-size: 13px;
  font-weight: 600;
}

.row {
  border-top: 1px solid var(--stroke);
}

.name {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.actions {
  display: flex;
  gap: 10px;
}

@media (max-width: 1024px) {
  .head,
  .row {
    grid-template-columns: repeat(2, 1fr);
    row-gap: 8px;
  }
}
</style>
