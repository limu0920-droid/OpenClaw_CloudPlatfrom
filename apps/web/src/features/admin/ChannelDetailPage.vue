<script setup lang="ts">
import { useRoute } from 'vue-router'
import { ref, watch } from 'vue'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { Channel } from '../../lib/types'

const route = useRoute()

const detail = ref<Channel | null>(null)
const loading = ref(true)
const error = ref('')
const checking = ref(false)

async function load(id: string) {
  loading.value = true
  error.value = ''
  try {
    detail.value = await api.getChannelDetail(id, 'admin')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载渠道失败'
  } finally {
    loading.value = false
  }
}

async function checkHealth() {
  if (!detail.value) return
  checking.value = true
  try {
    await api.checkChannelHealth(detail.value.id)
    await load(detail.value.id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '健康检查失败'
  } finally {
    checking.value = false
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
  <div v-if="loading" class="card state-card">正在加载渠道详情…</div>
  <div v-else-if="error" class="card state-card state-card--error">{{ error }}</div>
  <div v-else-if="detail" class="stack">
    <div class="card hero">
      <div>
        <div class="eyebrow">Channel Detail (Admin)</div>
        <h2>{{ detail.name }}</h2>
        <p class="muted">
          负责统一渠道健康、回调与安全策略。当前状态：{{ detail.status }} · 授权方式：{{ detail.authMode }}。
        </p>
      </div>
      <div class="hero-meta">
        <div class="pill">{{ detail.status }}</div>
        <button class="ghost" :disabled="checking" @click="checkHealth">
          {{ checking ? '检查中…' : '健康检查' }}
        </button>
        <div class="muted">{{ detail.lastActiveAt || '—' }}</div>
      </div>
    </div>

    <div class="detail-grid">
      <div class="card panel">
        <SectionHeader title="基础信息" subtitle="连接、健康与入口" />
        <div class="info-grid">
          <div>
            <span class="muted">提供方</span>
            <strong>{{ detail.provider }}</strong>
          </div>
          <div>
            <span class="muted">状态</span>
            <strong>{{ detail.status }}</strong>
          </div>
          <div>
            <span class="muted">健康</span>
            <strong>{{ detail.health || 'unknown' }}</strong>
          </div>
          <div>
            <span class="muted">接入时间</span>
            <strong>{{ detail.connectedAt || '—' }}</strong>
          </div>
          <div>
            <span class="muted">24h 消息</span>
            <strong>{{ detail.messages24h ?? 0 }}</strong>
          </div>
          <div>
            <span class="muted">成功率</span>
            <strong>{{ Math.round((detail.successRate ?? 0) * 100) }}%</strong>
          </div>
        </div>
      </div>

      <div class="card panel">
        <SectionHeader title="Webhook / 回调" subtitle="回调地址与签名" />
        <p class="muted">{{ detail.webhookUrl || '未配置回调' }}</p>
        <p class="muted" v-if="detail.callbackSecret">签名密钥：{{ detail.callbackSecret }}</p>
      </div>

      <div class="card panel">
        <SectionHeader title="最近活动" subtitle="用于运维排查与审计" />
        <div class="stack-list">
          <div v-for="act in detail.recentActivity" :key="act.id" class="stack-item">
            <strong>{{ act.type }}</strong>
            <span class="muted">{{ act.description }} · {{ act.time }}</span>
          </div>
          <div v-if="!detail.recentActivity?.length" class="muted">暂无活动。</div>
        </div>
      </div>

      <div class="card panel span-two">
        <SectionHeader title="入口与说明" subtitle="面向终端用户的入口与内部备注" />
        <div class="stack-list">
          <div v-for="entry in detail.entrypoints" :key="entry.url" class="stack-item">
            <strong>{{ entry.label }}</strong>
            <span class="muted">{{ entry.url }}</span>
          </div>
          <div v-if="!detail.entrypoints?.length" class="muted">暂无入口。</div>
          <p class="muted" v-if="detail.notes">备注：{{ detail.notes }}</p>
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

.hero {
  padding: 18px;
  display: flex;
  justify-content: space-between;
  gap: 12px;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.16), rgba(14, 165, 233, 0.12));
}

.hero-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-items: flex-end;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
}

.panel {
  padding: 14px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.stack-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stack-item {
  padding: 10px 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.span-two {
  grid-column: span 2;
}

@media (max-width: 900px) {
  .hero {
    flex-direction: column;
  }
  .hero-meta {
    align-items: flex-start;
  }
}
</style>
