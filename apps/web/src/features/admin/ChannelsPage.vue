<script setup lang="ts">
import { RouterLink } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: channels, loading, error } = useAsyncData(() => api.getChannels('admin'))
</script>

<template>
  <div class="card">
    <SectionHeader title="渠道接入中心" subtitle="查看所有渠道的连接状态与健康" />
    <div v-if="loading" class="state-card">正在加载渠道…</div>
    <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
    <div v-else class="table">
      <div class="head">
        <span>名称</span>
        <span>提供方</span>
        <span>状态</span>
        <span>授权方式</span>
        <span>24h 消息</span>
        <span>连接时间</span>
        <span>最近活动</span>
        <span>操作</span>
      </div>
      <div v-for="ch in channels" :key="ch.id" class="row">
        <span class="strong">{{ ch.name }}</span>
        <span class="muted">{{ ch.provider }}</span>
        <span class="pill">{{ ch.status }}</span>
        <span>{{ ch.authMode }}</span>
        <span>{{ ch.messages24h ?? 0 }}</span>
        <span class="muted">{{ ch.connectedAt || '—' }}</span>
        <span class="muted">{{ ch.lastActiveAt || '—' }}</span>
        <span><RouterLink :to="`/admin/channels/${ch.id}`">详情</RouterLink></span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.table {
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.head,
.row {
  display: grid;
  grid-template-columns: 1.2fr 0.9fr 0.9fr 0.9fr 0.8fr 1.1fr 1.1fr 0.7fr;
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
</style>
