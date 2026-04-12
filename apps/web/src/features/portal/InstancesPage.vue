<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'
import type { CreateInstancePayload } from '../../lib/types'

const filter = ref('')
const submitLoading = ref(false)
const submitError = ref('')
const submitSuccess = ref('')
const router = useRouter()

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
  <el-card shadow="never" class="surface-card">
    <SectionHeader title="实例列表" subtitle="当前通过平台 API 拉取实例清单" />
    <div class="create-form">
      <div class="form-row">
        <label>
          <span>名称</span>
          <el-input v-model="form.name" placeholder="例如：生产环境实例" />
        </label>
        <label>
          <span>套餐</span>
          <el-select v-model="form.plan">
            <el-option label="Pro" value="pro" />
            <el-option label="Standard" value="standard" />
            <el-option label="Trial" value="trial" />
          </el-select>
        </label>
        <label>
          <span>地域</span>
          <el-select v-model="form.region">
            <el-option label="cn-shanghai" value="cn-shanghai" />
            <el-option label="cn-beijing" value="cn-beijing" />
            <el-option label="ap-southeast-1" value="ap-southeast-1" />
          </el-select>
        </label>
      </div>
      <div class="form-row">
        <label>
          <span>CPU (vCPU)</span>
          <el-input v-model="form.cpu" />
        </label>
        <label>
          <span>内存</span>
          <el-input v-model="form.memory" placeholder="例如 4Gi" />
        </label>
        <el-button type="primary" :loading="submitLoading" @click="handleCreate">创建实例</el-button>
      </div>
      <el-alert v-if="submitError" :closable="false" show-icon type="error" :title="submitError" />
      <el-alert v-else-if="submitSuccess" :closable="false" show-icon type="success" :title="submitSuccess" />
    </div>

    <div class="toolbar">
      <el-input v-model="filter" placeholder="搜索实例..." clearable />
      <el-button plain @click="router.push('/portal/jobs')">查看变更任务</el-button>
    </div>

    <div v-if="loading" class="state-card">正在加载实例列表…</div>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
    <el-table v-else :data="filtered" class="surface-table">
      <el-table-column label="名称" min-width="240">
        <template #default="{ row }">
          <div class="name">
            <div class="strong">{{ row.name }}</div>
            <div class="muted">{{ row.plan }} · {{ row.code }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="状态" min-width="120">
        <template #default="{ row }">
          <el-tag round disable-transitions>{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" min-width="110" />
      <el-table-column prop="region" label="地域" min-width="130" />
      <el-table-column prop="updatedAt" label="更新" min-width="180" />
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <div class="row-actions">
            <el-button link type="primary" @click="router.push(`/portal/instances/${row.id}`)">详情</el-button>
            <el-button link type="primary" @click="router.push(`/portal/instances/${row.id}/workspace`)">对话</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<style scoped>
.surface-card {
  overflow: hidden;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  align-items: end;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: var(--text-muted);
}

.name {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.row-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

@media (max-width: 1024px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
