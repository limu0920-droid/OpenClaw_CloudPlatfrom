<script setup lang="ts">
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/vue'
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

const channelTabs = [
  { key: 'overview', label: '概览', subtitle: '授权 / 健康 / 指标' },
  { key: 'activity', label: '活动', subtitle: '最近动作 / 连接轨迹' },
  { key: 'entrypoints', label: '回调与入口', subtitle: 'Webhook / 会话入口' },
]

type TagTone = 'success' | 'warning' | 'danger' | 'info' | 'primary' | undefined

const healthTone = computed<TagTone>(() => {
  if (!detail.value?.health) return undefined
  if (detail.value.health === 'critical') return 'danger'
  if (detail.value.health === 'warning') return 'warning'
  return 'success'
})

const statusTone = computed<TagTone>(() => {
  if (!detail.value?.status) return undefined
  if (detail.value.status === 'connected') return 'success'
  if (detail.value.status === 'pending' || detail.value.status === 'degraded') return 'warning'
  return 'danger'
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
  <el-card v-if="loading" shadow="never" class="state-card">正在加载渠道详情…</el-card>
  <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
  <div v-else-if="detail" class="stack">
    <el-card shadow="never" class="hero">
      <div>
        <div class="eyebrow">Channel Detail</div>
        <h2>{{ detail.name }}</h2>
        <p class="muted">
          提供 {{ detail.provider }} 渠道的对话入口、Webhook 和健康检查。当前状态：{{ detail.status }}。
        </p>
      </div>
      <div class="hero-actions">
        <el-tag :type="statusTone" round disable-transitions>{{ detail.status }}</el-tag>
        <el-button plain @click="router.back()">返回</el-button>
        <el-button plain :loading="checking" @click="checkHealth">健康检查</el-button>
        <el-button type="primary" :loading="disconnecting" @click="disconnect">断开连接</el-button>
      </div>
    </el-card>

    <TabGroup class="channel-tabs">
      <TabList class="channel-tab-list">
        <Tab v-for="item in channelTabs" :key="item.key" as="template" v-slot="{ selected }">
          <button :class="['channel-tab', selected ? 'selected' : '']" type="button">
            <span>{{ item.subtitle }}</span>
            <strong>{{ item.label }}</strong>
          </button>
        </Tab>
      </TabList>

      <TabPanels class="channel-tab-panels">
        <TabPanel class="channel-tab-panel">
          <el-card shadow="never" class="panel">
            <SectionHeader title="连接信息" subtitle="授权方式、健康与最近活跃" />
            <div class="info-grid">
              <div class="info-card">
                <span class="muted">授权方式</span>
                <strong>{{ detail.authMode }}</strong>
              </div>
              <div class="info-card">
                <span class="muted">健康</span>
                <el-tag :type="healthTone" round disable-transitions>{{ detail.health || 'unknown' }}</el-tag>
              </div>
              <div class="info-card">
                <span class="muted">接入时间</span>
                <strong>{{ detail.connectedAt || '—' }}</strong>
              </div>
              <div class="info-card">
                <span class="muted">最近活跃</span>
                <strong>{{ detail.lastActiveAt || '—' }}</strong>
              </div>
              <div class="info-card">
                <span class="muted">24h 消息</span>
                <strong>{{ detail.messages24h ?? 0 }}</strong>
              </div>
              <div class="info-card">
                <span class="muted">成功率</span>
                <strong>{{ Math.round((detail.successRate ?? 0) * 100) }}%</strong>
              </div>
            </div>
          </el-card>
        </TabPanel>

        <TabPanel class="channel-tab-panel">
          <el-card shadow="never" class="panel">
            <SectionHeader title="最近活动" subtitle="用于排查连接和消息同步问题" />
            <div class="stack-list">
              <div v-for="act in detail.recentActivity" :key="act.id" class="stack-item">
                <strong>{{ act.type }}</strong>
                <span class="muted">{{ act.description }}</span>
                <span class="stack-time">{{ act.time }}</span>
              </div>
              <div v-if="!detail.recentActivity?.length" class="muted">暂无活动。</div>
            </div>
          </el-card>
        </TabPanel>

        <TabPanel class="channel-tab-panel">
          <div class="detail-grid">
            <el-card shadow="never" class="panel">
              <SectionHeader title="Webhook / 回调" subtitle="用于事件推送与消息分发" />
              <div class="stack-list">
                <div class="stack-item">
                  <strong>Webhook URL</strong>
                  <span class="muted">{{ detail.webhookUrl || '未配置回调地址' }}</span>
                </div>
                <div v-if="detail.callbackSecret" class="stack-item">
                  <strong>签名密钥</strong>
                  <span class="muted">{{ detail.callbackSecret }}</span>
                </div>
              </div>
            </el-card>

            <el-card shadow="never" class="panel">
              <SectionHeader title="会话入口" subtitle="跳转到渠道原生界面" />
              <div class="stack-list">
                <div v-for="entry in detail.entrypoints" :key="entry.url" class="stack-item">
                  <strong>{{ entry.label }}</strong>
                  <span class="muted">{{ entry.url }}</span>
                </div>
                <div v-if="!detail.entrypoints?.length" class="muted">暂无入口。</div>
              </div>
            </el-card>
          </div>
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

.hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  background: linear-gradient(135deg, rgba(33, 96, 255, 0.18), rgba(51, 170, 255, 0.12));
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.channel-tabs {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.channel-tab-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.channel-tab {
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

.channel-tab span {
  font-size: 12px;
  color: var(--text-muted);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.channel-tab strong {
  font-size: 1.05rem;
  font-variation-settings: "wght" 650;
}

.channel-tab.selected {
  border-color: rgba(29, 107, 255, 0.24);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(239, 244, 255, 0.98));
  box-shadow: 0 18px 40px rgba(29, 107, 255, 0.08);
}

.channel-tab-panels,
.channel-tab-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
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

.info-card,
.stack-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.stack-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stack-time {
  color: var(--text-muted);
  font-size: 12px;
}

@media (max-width: 900px) {
  .hero,
  .channel-tab-list,
  .info-grid {
    grid-template-columns: 1fr;
  }

  .hero {
    flex-direction: column;
  }
}
</style>
