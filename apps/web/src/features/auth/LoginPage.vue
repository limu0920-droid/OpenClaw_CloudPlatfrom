<script setup lang="ts">
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/vue'
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import { api } from '../../lib/api'
import { useBranding } from '../../lib/brand'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: authConfig } = useAsyncData(() => api.getAuthConfig())
const { data: authSession } = useAsyncData(() => api.getAuthSession())
const { brand, features } = useBranding()
const router = useRouter()

const entryModes = computed(() =>
  [
    features.value.portalEnabled
      ? {
          key: 'portal',
          label: 'Portal',
          subtitle: '租户用户',
          description: `适合 ${brand.value.name} 的业务、交付和实例使用方。围绕实例、渠道、任务和工单展开。`,
          actionLabel: '进入 Portal',
          action: () => router.push('/portal'),
        }
      : null,
    features.value.adminEnabled
      ? {
          key: 'admin',
          label: 'Admin',
          subtitle: '平台运维',
          description: `适合 ${brand.value.name} 的平台管理员和运维角色。聚焦租户、任务、告警与审计控制面。`,
          actionLabel: '进入 Admin',
          action: () => router.push('/admin'),
        }
      : null,
    {
      key: 'iam',
      label: 'IAM',
      subtitle: '统一认证',
      description: `当统一认证已启用时，使用 ${brand.value.name} 的身份入口跳转登录；未启用时该入口保持关闭。`,
      actionLabel: '使用统一登录',
      action: () => openKeycloak(),
    },
  ].filter(Boolean) as Array<{
    key: string
    label: string
    subtitle: string
    description: string
    actionLabel: string
    action: () => void
  }>,
)

async function openKeycloak() {
  const response = await api.getKeycloakLoginURL(window.location.origin + '/login')
  window.location.href = response.url
}
</script>

<template>
  <div class="login-shell">
    <section class="login-hero">
      <div class="eyebrow">OpenClaw Platform</div>
      <h1>通过 {{ brand.name }} 门户进入真实 IAM、实例编排和运营控制面。</h1>
      <p>
        这一版登录页参考了轻量企业 SaaS 的入口表达：左侧负责品牌与可信感，右侧负责明确的下一步动作。
      </p>
      <div class="login-hero__stats">
        <el-card shadow="never" class="metric-card">
          <strong>2 套壳层</strong>
          <span>Portal 与 Admin 分离</span>
        </el-card>
        <el-card shadow="never" class="metric-card">
          <strong>{{ brand.name }}</strong>
          <span>已接通实例、任务、告警接口</span>
        </el-card>
        <el-card shadow="never" class="metric-card">
          <strong>Keycloak</strong>
          <span>{{ authConfig?.enabled ? '已接入统一认证' : '当前未启用统一认证' }}</span>
        </el-card>
      </div>
    </section>

    <el-card shadow="never" class="login-panel">
      <div class="eyebrow">安全登录</div>
      <h2>选择你的入口</h2>
      <p class="muted">当前页面承接统一认证与控制台入口；未启用的认证提供方会保持关闭。</p>
      <TabGroup class="login-tabs">
        <TabList class="login-tab-list">
          <Tab v-for="mode in entryModes" :key="mode.key" as="template" v-slot="{ selected }">
            <button :class="['login-tab', selected ? 'selected' : '']" type="button">
              <span>{{ mode.subtitle }}</span>
              <strong>{{ mode.label }}</strong>
            </button>
          </Tab>
        </TabList>
        <TabPanels class="login-tab-panels">
          <TabPanel v-for="mode in entryModes" :key="mode.key" class="login-tab-panel">
            <div class="login-tab-panel__copy">
              <div class="eyebrow eyebrow-soft">{{ mode.subtitle }}</div>
              <h3>{{ mode.label }}</h3>
              <p>{{ mode.description }}</p>
            </div>
            <el-button
              round
              size="large"
              :type="mode.key === 'portal' ? 'primary' : 'default'"
              :plain="mode.key !== 'portal'"
              :disabled="mode.key === 'iam' && (!authConfig?.enabled || !features.ssoEnabled)"
              @click="mode.action"
            >
              {{ mode.key === 'iam' && (!authConfig?.enabled || !features.ssoEnabled) ? '统一认证未启用' : mode.actionLabel }}
            </el-button>
          </TabPanel>
        </TabPanels>
      </TabGroup>
      <div class="login-form">
        <label>
          <span>租户 / 邮箱</span>
          <el-input :model-value="authSession?.user?.email || 'acme@example.com'" readonly />
        </label>
        <label>
          <span>当前身份 / 模式</span>
          <el-input
            :model-value="
              authSession?.authenticated
                ? `provider: ${authSession.provider}\nuser: ${authSession.user?.name}\nrole: ${authSession.user?.role}`
                : `当前页面只做入口承接。\nPortal 面向租户用户，Admin 面向平台运维与审计角色。\n若配置了 Keycloak，可使用统一 IAM 登录。`
            "
            type="textarea"
            :autosize="{ minRows: 5, maxRows: 7 }"
            readonly
          />
        </label>
        <label v-if="authConfig">
          <span>Keycloak 配置</span>
          <el-input
            :model-value="`enabled: ${authConfig.enabled}\nrealm: ${authConfig.realm || '-'}\nclientId: ${authConfig.clientId || '-'}\nredirect: ${authConfig.defaultRedirect || '-'}`"
            type="textarea"
            :autosize="{ minRows: 4, maxRows: 6 }"
            readonly
          />
        </label>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.login-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  gap: 24px;
  padding: 32px;
}

.login-hero,
.login-panel {
  padding: 32px;
}

.login-hero {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 18px;
  border-radius: 28px;
  background:
    radial-gradient(circle at top left, rgba(147, 197, 253, 0.35), transparent 28%),
    linear-gradient(135deg, #ffffff, #eaf2ff 68%, #dde9ff);
  border: 1px solid rgba(99, 102, 241, 0.16);
  box-shadow: var(--shadow-soft);
}

.login-hero h1 {
  margin: 0;
  max-width: 700px;
  font-size: clamp(2rem, 5vw, 3.5rem);
  line-height: 1.05;
}

.login-hero p {
  margin: 0;
  max-width: 640px;
  color: var(--text-muted);
  font-size: 1.04rem;
}

.login-hero__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.metric-card strong {
  display: block;
  font-size: 1.3rem;
}

.metric-card span {
  display: block;
  margin-top: 8px;
  color: var(--text-muted);
}

.login-panel {
  align-self: center;
}

.login-panel h2 {
  margin: 0;
  font-size: 2rem;
}

.login-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 20px;
}

.login-tabs {
  margin-top: 24px;
}

.login-tab-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.login-tab {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  padding: 16px 18px;
  text-align: left;
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(255, 255, 255, 0.72);
}

.login-tab span {
  font-size: 12px;
  color: var(--text-muted);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.login-tab strong {
  font-size: 1.08rem;
  font-variation-settings: "wght" 650;
}

.login-tab.selected {
  border-color: rgba(29, 107, 255, 0.24);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(239, 244, 255, 0.98));
  box-shadow: 0 18px 40px rgba(29, 107, 255, 0.08);
}

.login-tab-panels {
  margin-top: 12px;
}

.login-tab-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.94), rgba(243, 247, 255, 0.98));
}

.login-tab-panel__copy {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.login-tab-panel__copy h3 {
  margin: 0;
  font-size: 1.5rem;
}

.login-tab-panel__copy p {
  margin: 0;
  color: var(--text-muted);
  line-height: 1.75;
}

.eyebrow-soft {
  background: rgba(29, 107, 255, 0.06);
}

.login-form {
  display: grid;
  gap: 16px;
  margin-top: 24px;
}

.login-form label {
  display: grid;
  gap: 8px;
}

.login-form span {
  font-size: 0.9rem;
  color: var(--text-muted);
}

@media (max-width: 980px) {
  .login-shell {
    grid-template-columns: 1fr;
    padding: 18px;
  }

  .login-hero__stats {
    grid-template-columns: 1fr;
  }

  .login-tab-list {
    grid-template-columns: 1fr;
  }

  .login-tab-panel {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
