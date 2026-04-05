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
    <div class="card panel">
      <SectionHeader title="问题上报 / 工单" subtitle="提交运行问题、资源异常、渠道接入问题或购买咨询" />
      <div class="form-grid">
        <label>
          <span>标题</span>
          <input v-model="form.title" placeholder="例如：实例 CPU 异常升高" />
        </label>
        <label>
          <span>分类</span>
          <select v-model="form.category">
            <option value="general">general</option>
            <option value="backup">backup</option>
            <option value="performance">performance</option>
            <option value="billing">billing</option>
            <option value="channel">channel</option>
          </select>
        </label>
        <label>
          <span>优先级</span>
          <select v-model="form.severity">
            <option value="low">low</option>
            <option value="medium">medium</option>
            <option value="high">high</option>
          </select>
        </label>
      </div>
      <label class="block">
        <span>描述</span>
        <textarea v-model="form.description" placeholder="描述现象、影响范围、最近操作和期望结果" />
      </label>
      <div class="actions">
        <button class="primary" :disabled="submitting" @click="submitTicket">
          {{ submitting ? '提交中…' : '提交工单' }}
        </button>
        <span v-if="feedback" class="success">{{ feedback }}</span>
        <span v-if="feedbackError" class="error">{{ feedbackError }}</span>
      </div>
    </div>

    <div class="card panel">
      <SectionHeader title="我的工单" subtitle="查看当前租户下的问题处理状态" />
      <div v-if="loading" class="state-card">正在读取工单…</div>
      <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
      <div v-else class="table">
        <div class="head">
          <span>单号</span>
          <span>标题</span>
          <span>分类</span>
          <span>优先级</span>
          <span>状态</span>
          <span>更新时间</span>
        </div>
        <div v-for="ticket in tickets" :key="ticket.id" class="row">
          <span class="strong">{{ ticket.ticketNo }}</span>
          <span>{{ ticket.title }}</span>
          <span>{{ ticket.category }}</span>
          <span>{{ ticket.severity }}</span>
          <span class="pill">{{ ticket.status }}</span>
          <span class="muted">{{ ticket.updatedAt }}</span>
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

input,
select,
textarea {
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--stroke);
  background: #fff;
}

textarea {
  min-height: 120px;
  resize: vertical;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
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

.success {
  color: #15803d;
}

.error {
  color: #b91c1c;
}

.table {
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.head,
.row {
  display: grid;
  grid-template-columns: 1fr 1.6fr 0.8fr 0.8fr 0.8fr 1fr;
  gap: 10px;
  padding: 12px 14px;
  align-items: center;
}

.head {
  background: var(--panel-muted);
  color: var(--text-muted);
  font-weight: 600;
}

.row {
  border-top: 1px solid var(--stroke);
}

@media (max-width: 1024px) {
  .form-grid,
  .head,
  .row {
    grid-template-columns: 1fr;
  }
}
</style>
