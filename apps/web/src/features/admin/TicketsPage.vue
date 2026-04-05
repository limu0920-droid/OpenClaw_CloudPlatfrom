<script setup lang="ts">
import { ref } from 'vue'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const changingId = ref('')
const { data: tickets, loading, error, reload } = useAsyncData(() => api.getAdminTickets())

async function updateStatus(id: string, status: string) {
  changingId.value = id
  try {
    await api.updateAdminTicketStatus(id, status, status === 'in_progress' ? 'ops-admin' : '')
    await reload()
  } finally {
    changingId.value = ''
  }
}
</script>

<template>
  <div class="card panel">
    <SectionHeader title="工单中心" subtitle="管理员视角处理问题、购买咨询和运行异常" />
    <div v-if="loading" class="state-card">正在读取工单…</div>
    <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
    <div v-else class="table">
      <div class="head">
        <span>单号</span>
        <span>标题</span>
        <span>分类</span>
        <span>优先级</span>
        <span>状态</span>
        <span>指派</span>
        <span>操作</span>
      </div>
      <div v-for="ticket in tickets" :key="ticket.id" class="row">
        <span class="strong">{{ ticket.ticketNo }}</span>
        <span>{{ ticket.title }}</span>
        <span>{{ ticket.category }}</span>
        <span>{{ ticket.severity }}</span>
        <span class="pill">{{ ticket.status }}</span>
        <span>{{ ticket.assignee || '—' }}</span>
        <div class="actions">
          <button class="ghost" :disabled="changingId === ticket.id" @click="updateStatus(ticket.id, 'in_progress')">接单</button>
          <button class="ghost" :disabled="changingId === ticket.id" @click="updateStatus(ticket.id, 'resolved')">解决</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel {
  padding: 14px;
}

.table {
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.head,
.row {
  display: grid;
  grid-template-columns: 1fr 1.6fr 0.8fr 0.8fr 0.8fr 0.8fr 1fr;
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

.actions {
  display: flex;
  gap: 8px;
}

.ghost {
  padding: 8px 10px;
  border-radius: 10px;
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
  color: var(--text);
}

@media (max-width: 1024px) {
  .head,
  .row {
    grid-template-columns: 1fr;
  }
}
</style>
