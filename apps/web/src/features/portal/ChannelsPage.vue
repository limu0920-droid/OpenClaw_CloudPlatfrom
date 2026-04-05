<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { Channel, ChannelConnectPayload, ChannelProvider } from '../../lib/types'
import { useAsyncData } from '../../lib/useAsyncData'

const providerOptions: Array<{ value: ChannelProvider; label: string }> = [
  { value: 'feishu', label: '飞书' },
  { value: 'wechat_work', label: '企业微信' },
  { value: 'dingtalk', label: '钉钉' },
  { value: 'slack', label: 'Slack' },
  { value: 'telegram', label: 'Telegram' },
  { value: 'discord', label: 'Discord' },
  { value: 'whatsapp', label: 'WhatsApp' },
  { value: 'custom', label: '自定义' },
]

const { data: channels, loading, error, reload } = useAsyncData(() => api.getChannels('portal'))

const connectForm = ref<ChannelConnectPayload>({
  provider: 'feishu',
  authMode: 'oauth',
  redirectUri: '',
  token: '',
})
const connectMsg = ref('')
const connectError = ref('')
const submitting = ref(false)

const filteredChannels = computed<Channel[]>(() => channels.value ?? [])

async function connect() {
  connectMsg.value = ''
  connectError.value = ''
  submitting.value = true
  try {
    await api.connectChannel(connectForm.value)
    connectMsg.value = '连接请求已提交，稍后刷新状态'
    await reload()
  } catch (err) {
    connectError.value = err instanceof Error ? err.message : '连接失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="stack">
    <div class="card hero">
      <div>
        <div class="eyebrow">Channels</div>
        <h2>一键接入第三方聊天平台</h2>
        <p class="muted">
          支持飞书、企微、钉钉、Slack、Telegram、Discord、WhatsApp 等常见渠道。连接后可在门户侧直接承接会话入口。
        </p>
        <div class="hero-actions">
          <a class="ghost" href="https://docs" target="_blank">查看接入指南</a>
          <a class="primary" href="#connect">立即连接</a>
        </div>
      </div>
    </div>

    <div id="connect" class="card">
      <SectionHeader title="一键接入" subtitle="选择渠道并提交授权" />
      <div class="form-grid">
        <label>
          <span>渠道</span>
          <select v-model="connectForm.provider">
            <option v-for="opt in providerOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
        </label>
        <label>
          <span>授权方式</span>
          <select v-model="connectForm.authMode">
            <option value="oauth">OAuth</option>
            <option value="qr">扫码</option>
            <option value="token">Token</option>
            <option value="webhook">Webhook</option>
            <option value="custom">自定义</option>
          </select>
        </label>
        <label v-if="connectForm.authMode === 'oauth'">
          <span>回调地址</span>
          <input v-model="connectForm.redirectUri" placeholder="https://portal.example.com/callback" />
        </label>
        <label v-if="connectForm.authMode === 'token'">
          <span>Token / Bot Key</span>
          <input v-model="connectForm.token" placeholder="填写渠道 Token" />
        </label>
        <button class="primary" :disabled="submitting" @click="connect">
          {{ submitting ? '连接中…' : '发起连接' }}
        </button>
      </div>
      <div class="feedback">
        <span v-if="connectError" class="error">{{ connectError }}</span>
        <span v-else-if="connectMsg" class="success">{{ connectMsg }}</span>
      </div>
    </div>

    <div class="card">
      <SectionHeader title="渠道列表" subtitle="连接状态、健康检查、最近活动" />
      <div v-if="loading" class="state-card">正在加载渠道…</div>
      <div v-else-if="error" class="state-card state-card--error">{{ error }}</div>
      <div v-else class="channel-grid">
        <RouterLink
          v-for="ch in filteredChannels"
          :key="ch.id"
          :to="`/portal/channels/${ch.id}`"
          class="channel-card"
        >
          <div class="channel-head">
            <div>
              <div class="label">{{ ch.name }}</div>
              <div class="muted">{{ ch.provider }}</div>
            </div>
            <span :class="['status', ch.status]">{{ ch.status }}</span>
          </div>
          <div class="muted">授权方式：{{ ch.authMode }}</div>
          <div class="muted">24h 消息：{{ ch.messages24h ?? 0 }} · 成功率：{{ Math.round((ch.successRate ?? 0) * 100) }}%</div>
          <div class="activity">
            <div v-for="act in ch.recentActivity" :key="act.id" class="pill pill-ghost">
              {{ act.description }}
            </div>
          </div>
          <div class="entrance" v-if="ch.entrypoints?.length">
            <span class="muted">入口</span>
            <div class="links">
              <span v-for="entry in ch.entrypoints" :key="entry.url" class="pill">{{ entry.label }}</span>
            </div>
          </div>
        </RouterLink>
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
  padding: 20px;
  background: linear-gradient(135deg, rgba(33, 96, 255, 0.18), rgba(51, 170, 255, 0.14));
}

.hero-actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
  align-items: center;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: var(--text-muted);
  font-size: 13px;
}

input,
select {
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--stroke);
  background: #fff;
}

button.primary {
  height: 44px;
}

.feedback {
  min-height: 20px;
  margin-top: 6px;
}

.error {
  color: #b91c1c;
}

.success {
  color: #15803d;
}

.channel-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 12px;
}

.channel-card {
  display: grid;
  gap: 8px;
  padding: 14px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel);
  box-shadow: var(--shadow-soft);
}

.channel-card:hover {
  border-color: var(--brand);
}

.channel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.label {
  font-weight: 700;
}

.status {
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  text-transform: lowercase;
}

.status.connected {
  background: rgba(52, 211, 153, 0.16);
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

.status.disconnected,
.status.error {
  background: rgba(248, 113, 113, 0.18);
  color: #b91c1c;
}

.activity {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.pill-ghost {
  background: var(--panel-muted);
}

.entrance .links {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
