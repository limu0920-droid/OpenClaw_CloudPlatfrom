<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { computed, ref, watch } from 'vue'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { Channel } from '../../lib/types'

const route = useRoute()
const router = useRouter()

const detail = ref<Channel | null>(null)
const loading = ref(true)
const error = ref('')
const disconnecting = ref(false)
const checking = ref(false)

const healthTone = computed(() => {
  if (!detail.value?.health) return 'neutral'
  if (detail.value.health === 'critical') return 'critical'
  if (detail.value.health === 'warning') return 'warning'
  return 'good'
})

async function load(id: string) {
  loading.value = true
  error.value = ''
  try {
    detail.value = await api.getChannelDetail(id, 'portal')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载渠道失败'
  } finally {
    loading.value = false
  }
}

async function disconnect() {
  if (!detail.value) return
  disconnecting.value = true
  try {
    await api.disconnectChannel(detail.value.id)
    await load(detail.value.id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '断开失败'
  } finally {
    disconnecting.value = false
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
        <div class="eyebrow">Channel Detail</div>
        <h2>{{ detail.name }}</h2>
        <p class="muted">
          提供 {{ detail.provider }} 渠道的对话入口、Webhook 和健康检查。当前状态：{{ detail.status }}。
        </p>
      </div>
      <div class="hero-actions">
        <span :class="['status', detail.status]">{{ detail.status }}</span>
        <button class="ghost" @click="router.back()">返回</button>
        <button class="ghost" :disabled="checking" @click="checkHealth">
          {{ checking ? '检查中…' : '健康检查' }}
        </button>
        <button class="primary" :disabled="disconnecting" @click="disconnect">
          {{ disconnecting ? '断开中…' : '断开连接' }}
        </button>
      </div>
    </div>

    <div class="detail-grid">
      <div class="card panel">
        <SectionHeader title="连接信息" subtitle="授权方式与最近活动" />
        <div class="info-grid">
          <div>
            <span class="muted">授权方式</span>
            <strong>{{ detail.authMode }}</strong>
          </div>
          <div>
            <span class="muted">健康</span>
            <strong :class="['health', healthTone]">{{ detail.health || 'unknown' }}</strong>
          </div>
          <div>
            <span class="muted">接入时间</span>
            <strong>{{ detail.connectedAt || '—' }}</strong>
          </div>
          <div>
            <span class="muted">最近活跃</span>
            <strong>{{ detail.lastActiveAt || '—' }}</strong>
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
        <div class="activity stack-list">
          <div v-for="act in detail.recentActivity" :key="act.id" class="stack-item">
            <strong>{{ act.type }}</strong>
            <span class="muted">{{ act.description }} · {{ act.time }}</span>
          </div>
          <div v-if="!detail.recentActivity?.length" class="muted">暂无活动。</div>
        </div>
      </div>

      <div class="card panel">
        <SectionHeader title="Webhook / 回调" subtitle="用于事件推送与消息分发" />
        <p class="muted">{{ detail.webhookUrl || '未配置回调地址' }}</p>
        <p class="muted" v-if="detail.callbackSecret">签名密钥：{{ detail.callbackSecret }}</p>
      </div>

      <div class="card panel">
        <SectionHeader title="会话入口" subtitle="跳转到渠道原生界面" />
        <div class="stack-list">
          <div v-for="entry in detail.entrypoints" :key="entry.url" class="stack-item">
            <strong>{{ entry.label }}</strong>
            <span class="muted">{{ entry.url }}</span>
          </div>
          <div v-if="!detail.entrypoints?.length" class="muted">暂无入口。</div>
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
  align-items: flex-start;
  gap: 12px;
  background: linear-gradient(135deg, rgba(33, 96, 255, 0.18), rgba(51, 170, 255, 0.12));
}

.hero-actions {
  display: flex;
  gap: 8px;
  align-items: center;
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
  margin-bottom: 12px;
}

.health.good {
  color: #15803d;
}
.health.warning {
  color: #ca8a04;
}
.health.critical {
  color: #b91c1c;
}

.activity {
  margin-top: 8px;
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

.status {
  padding: 8px 12px;
  border-radius: 999px;
  font-size: 12px;
  text-transform: lowercase;
}

.status.connected {
  background: rgba(52, 211, 153, 0.18);
  color: #0f5132;
}

.status.pending {
  background: rgba(251, 191, 36, 0.18);
  color: #92400e;
}

.status.degraded {
  background: rgba(245, 158, 11, 0.18);
  color: #92400e;
}

.status.error,
.status.disconnected {
  background: rgba(248, 113, 113, 0.18);
  color: #b91c1c;
}

@media (max-width: 900px) {
  .hero {
    flex-direction: column;
  }
  .hero-actions {
    align-items: flex-start;
  }
}
</style>
