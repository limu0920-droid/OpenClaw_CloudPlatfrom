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

function statusTagType(status: Channel['status']) {
  if (status === 'connected') return 'success'
  if (status === 'pending') return 'warning'
  if (status === 'degraded') return 'warning'
  return 'danger'
}

function openDocs() {
  window.open('/api/v1/docs/external/integration.md', '_blank', 'noopener,noreferrer')
}

function scrollToConnect() {
  document.getElementById('connect')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

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
    <el-card shadow="never" class="hero">
      <div>
        <div class="eyebrow">Channels</div>
        <h2>一键接入第三方聊天平台</h2>
        <p class="muted">
          支持飞书、企微、钉钉、Slack、Telegram、Discord、WhatsApp 等常见渠道。连接后可在门户侧直接承接会话入口。
        </p>
        <div class="hero-actions">
          <el-button round plain @click="openDocs">查看接入指南</el-button>
          <el-button round type="primary" @click="scrollToConnect">立即连接</el-button>
        </div>
      </div>
    </el-card>

    <el-card id="connect" shadow="never">
      <SectionHeader title="一键接入" subtitle="选择渠道并提交授权" />
      <div class="form-grid">
        <label>
          <span>渠道</span>
          <el-select v-model="connectForm.provider">
            <el-option v-for="opt in providerOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </label>
        <label>
          <span>授权方式</span>
          <el-select v-model="connectForm.authMode">
            <el-option label="OAuth" value="oauth" />
            <el-option label="扫码" value="qr" />
            <el-option label="Token" value="token" />
            <el-option label="Webhook" value="webhook" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </label>
        <label v-if="connectForm.authMode === 'oauth'">
          <span>回调地址</span>
          <el-input v-model="connectForm.redirectUri" placeholder="https://portal.example.com/callback" />
        </label>
        <label v-if="connectForm.authMode === 'token'">
          <span>Token / Bot Key</span>
          <el-input v-model="connectForm.token" placeholder="填写渠道 Token" />
        </label>
        <el-button type="primary" :loading="submitting" @click="connect">发起连接</el-button>
      </div>
      <el-alert v-if="connectError" :closable="false" show-icon type="error" :title="connectError" />
      <el-alert v-else-if="connectMsg" :closable="false" show-icon type="success" :title="connectMsg" />
    </el-card>

    <el-card shadow="never">
      <SectionHeader title="渠道列表" subtitle="连接状态、健康检查、最近活动" />
      <div v-if="loading" class="state-card">正在加载渠道…</div>
      <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
      <div v-else class="channel-grid">
        <el-card v-for="ch in filteredChannels" :key="ch.id" shadow="hover" class="channel-card">
          <RouterLink :to="`/portal/channels/${ch.id}`" class="channel-link">
            <div class="channel-head">
              <div>
                <div class="label">{{ ch.name }}</div>
                <div class="muted">{{ ch.provider }}</div>
              </div>
              <el-tag :type="statusTagType(ch.status)" round disable-transitions>{{ ch.status }}</el-tag>
            </div>
            <div class="muted">授权方式：{{ ch.authMode }}</div>
            <div class="muted">24h 消息：{{ ch.messages24h ?? 0 }} · 成功率：{{ Math.round((ch.successRate ?? 0) * 100) }}%</div>
            <div class="activity">
              <el-tag v-for="act in ch.recentActivity" :key="act.id" round effect="plain" disable-transitions>
                {{ act.description }}
              </el-tag>
            </div>
            <div v-if="ch.entrypoints?.length" class="entrance">
              <span class="muted">入口</span>
              <div class="links">
                <el-tag v-for="entry in ch.entrypoints" :key="entry.url" round disable-transitions>{{ entry.label }}</el-tag>
              </div>
            </div>
          </RouterLink>
        </el-card>
      </div>
    </el-card>
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
  align-items: end;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: var(--text-muted);
  font-size: 13px;
}

.channel-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 12px;
}

.channel-card {
  height: 100%;
}

.channel-link {
  display: grid;
  gap: 8px;
}

.channel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.label {
  font-weight: 700;
}

.activity {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.entrance .links {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.state-card {
  padding: 18px;
  border: 1px dashed var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel-muted);
  text-align: center;
}
</style>
