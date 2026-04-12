<script setup lang="ts">
import { reactive, ref } from 'vue'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const form = reactive({
  title: '',
  category: 'general',
  severity: 'medium',
  description: '',
  reporter: '陆明舟',
})

const submitting = ref(false)
const feedback = ref('')
const feedbackError = ref('')

const { data: tickets, loading, error, reload } = useAsyncData(() => api.getPortalTickets())

async function submitTicket() {
  feedback.value = ''
  feedbackError.value = ''
  submitting.value = true
  try {
    await api.createPortalTicket({
      title: form.title,
      category: form.category,
      severity: form.severity,
      description: form.description,
      reporter: form.reporter,
    })
    feedback.value = '问题已提交，平台管理员会在工单中心跟进。'
    form.title = ''
    form.description = ''
    await reload()
  } catch (err) {
    feedbackError.value = err instanceof Error ? err.message : '提交失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="stack">
    <el-card shadow="never" class="panel">
      <SectionHeader title="问题上报 / 工单" subtitle="提交运行问题、资源异常、渠道接入问题或购买咨询" />
      <div class="form-grid">
        <label>
          <span>标题</span>
          <el-input v-model="form.title" placeholder="例如：实例 CPU 异常升高" />
        </label>
        <label>
          <span>分类</span>
          <el-select v-model="form.category">
            <el-option label="general" value="general" />
            <el-option label="backup" value="backup" />
            <el-option label="performance" value="performance" />
            <el-option label="billing" value="billing" />
            <el-option label="channel" value="channel" />
          </el-select>
        </label>
        <label>
          <span>优先级</span>
          <el-select v-model="form.severity">
            <el-option label="low" value="low" />
            <el-option label="medium" value="medium" />
            <el-option label="high" value="high" />
          </el-select>
        </label>
      </div>
      <label class="block">
        <span>描述</span>
        <el-input
          v-model="form.description"
          type="textarea"
          :autosize="{ minRows: 5, maxRows: 8 }"
          placeholder="描述现象、影响范围、最近操作和期望结果"
        />
      </label>
      <div class="actions">
        <el-button type="primary" :loading="submitting" @click="submitTicket">提交工单</el-button>
      </div>
      <el-alert v-if="feedback" :closable="false" show-icon type="success" :title="feedback" />
      <el-alert v-if="feedbackError" :closable="false" show-icon type="error" :title="feedbackError" />
    </el-card>

    <el-card shadow="never" class="panel">
      <SectionHeader title="我的工单" subtitle="查看当前租户下的问题处理状态" />
      <div v-if="loading" class="state-card">正在读取工单…</div>
      <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
      <el-table v-else :data="tickets ?? []" class="surface-table">
        <el-table-column prop="ticketNo" label="单号" min-width="130" />
        <el-table-column prop="title" label="标题" min-width="220" />
        <el-table-column prop="category" label="分类" min-width="120" />
        <el-table-column prop="severity" label="优先级" min-width="120" />
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag round disable-transitions>{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" min-width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.panel {
  padding: 14px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 10px;
}

label span {
  color: var(--text-muted);
  font-size: 0.86rem;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
}

@media (max-width: 1024px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
