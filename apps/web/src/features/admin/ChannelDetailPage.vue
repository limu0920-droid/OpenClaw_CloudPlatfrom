<script setup lang="ts">
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/vue'
import { useRoute } from 'vue-router'
import { computed, ref, watch } from 'vue'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { Channel } from '../../lib/types'

const route = useRoute()

const detail = ref<Channel | null>(null)
const loading = ref(true)
const error = ref('')
const checking = ref(false)

const channelTabs = [
  { key: 'overview', label: '基础信息', subtitle: '提供方 / 状态 / 指标' },
  { key: 'security', label: '回调与安全', subtitle: 'Webhook / 签名 / 备注' },
  { key: 'activity', label: '活动与入口', subtitle: '最近动作 / 用户入口' },
]

type TagTone = 'success' | 'warning' | 'danger' | 'info' | 'primary' | undefined

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
  <el-card v-if="loading" shadow="never" class="state-card">正在加载渠道详情…</el-card>
  <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
  <div v-else-if="detail" class="stack">
    <el-card shadow="never" class="hero">
      <div>
        <div class="eyebrow">Channel Detail (Admin)</div>
        <h2>{{ detail.name }}</h2>
        <p class="muted">
          负责统一渠道健康、回调与安全策略。当前状态：{{ detail.status }} · 授权方式：{{ detail.authMode }}。
        </p>
      </div>
      <div class="hero-meta">
        <el-tag :type="statusTone" round disable-transitions>{{ detail.status }}</el-tag>
        <el-button plain :loading="checking" @click="checkHealth">健康检查</el-button>
        <div class="muted">{{ detail.lastActiveAt || '—' }}</div>
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
            <SectionHeader title="基础信息" subtitle="连接、健康与入口指标" />
            <div class="info-grid">
              <div class="info-card">
                <span class="muted">提供方</span>
                <strong>{{ detail.provider }}</strong>
              </div>
              <div class="info-card">
                <span class="muted">状态</span>
                <el-tag :type="statusTone" round disable-transitions>{{ detail.status }}</el-tag>
              </div>
              <div class="info-card">
                <span class="muted">健康</span>
                <strong>{{ detail.health || 'unknown' }}</strong>
              </div>
              <div class="info-card">
                <span class="muted">接入时间</span>
                <strong>{{ detail.connectedAt || '—' }}</strong>
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
          <div class="detail-grid">
            <el-card shadow="never" class="panel">
              <SectionHeader title="Webhook / 回调" subtitle="回调地址与签名" />
              <div class="stack-list">
                <div class="stack-item">
                  <strong>Webhook URL</strong>
                  <span class="muted">{{ detail.webhookUrl || '未配置回调' }}</span>
                </div>
                <div v-if="detail.callbackSecret" class="stack-item">
                  <strong>签名密钥</strong>
                  <span class="muted">{{ detail.callbackSecret }}</span>
                </div>
              </div>
            </el-card>

            <el-card shadow="never" class="panel">
              <SectionHeader title="内部备注" subtitle="面向运维的补充说明" />
              <div class="stack-list">
                <div class="stack-item">
                  <strong>授权方式</strong>
                  <span class="muted">{{ detail.authMode }}</span>
                </div>
                <div class="stack-item">
                  <strong>备注</strong>
                  <span class="muted">{{ detail.notes || '暂无备注。' }}</span>
                </div>
              </div>
            </el-card>
          </div>
        </TabPanel>

        <TabPanel class="channel-tab-panel">
          <div class="detail-grid">
            <el-card shadow="never" class="panel">
              <SectionHeader title="最近活动" subtitle="用于运维排查与审计" />
              <div class="stack-list">
                <div v-for="act in detail.recentActivity" :key="act.id" class="stack-item">
                  <strong>{{ act.type }}</strong>
                  <span class="muted">{{ act.description }}</span>
                  <span class="stack-time">{{ act.time }}</span>
                </div>
                <div v-if="!detail.recentActivity?.length" class="muted">暂无活动。</div>
              </div>
            </el-card>

            <el-card shadow="never" class="panel">
              <SectionHeader title="入口与说明" subtitle="面向终端用户的入口与内部备注" />
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
  gap: 12px;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.16), rgba(14, 165, 233, 0.12));
}

.hero-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-end;
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
  border-color: rgba(59, 130, 246, 0.24);
  background: linear-gradient(180deg, rgba(20, 31, 48, 0.96), rgba(19, 28, 43, 0.98));
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.14);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
}

.panel {
  padding: 14px;
}

.channel-tab-panels,
.channel-tab-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
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

  .hero-meta {
    align-items: flex-start;
  }
}
</style>
