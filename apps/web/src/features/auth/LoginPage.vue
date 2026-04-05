<script setup lang="ts">
import { RouterLink } from 'vue-router'

import { api } from '../../lib/api'
import { useAsyncData } from '../../lib/useAsyncData'

const { data: authConfig } = useAsyncData(() => api.getAuthConfig())
const { data: authSession } = useAsyncData(() => api.getAuthSession())

async function openKeycloak() {
  const response = await api.getKeycloakLoginURL(window.location.origin + '/login')
  window.location.href = response.url
}
</script>

<template>
  <div class="login-shell">
    <section class="login-hero">
      <div class="eyebrow">OpenClaw Platform</div>
      <h1>先进入平台门户，再逐步接入真实 IAM 与实例编排。</h1>
      <p>
        这一版登录页参考了轻量企业 SaaS 的入口表达：左侧负责品牌与可信感，右侧负责明确的下一步动作。
      </p>
      <div class="login-hero__stats">
        <div class="card">
          <strong>2 套壳层</strong>
          <span>Portal 与 Admin 分离</span>
        </div>
        <div class="card">
          <strong>Go Mock API</strong>
          <span>已接通实例、任务、告警接口</span>
        </div>
        <div class="card">
          <strong>Keycloak</strong>
          <span>{{ authConfig?.enabled ? '已配置接入骨架' : '当前未启用，仍可使用本地演示入口' }}</span>
        </div>
      </div>
    </section>

    <section class="login-panel card">
      <div class="eyebrow">安全登录</div>
      <h2>选择你的入口</h2>
      <p class="muted">当前已接入 Keycloak 配置骨架。未启用时仍可直接进入 Portal / Admin 演示模式。</p>
      <div class="login-actions">
        <RouterLink class="primary" to="/portal">进入 Portal</RouterLink>
        <RouterLink class="ghost" to="/admin">进入 Admin</RouterLink>
        <button v-if="authConfig?.enabled" class="ghost" type="button" @click="openKeycloak">使用 Keycloak 登录</button>
      </div>
      <div class="login-form">
        <label>
          <span>租户 / 邮箱</span>
          <input :value="authSession?.user?.email || 'acme@example.com'" readonly />
        </label>
        <label>
          <span>当前身份 / 模式</span>
          <textarea readonly>{{
            authSession?.authenticated
              ? `provider: ${authSession.provider}\nuser: ${authSession.user?.name}\nrole: ${authSession.user?.role}`
              : `当前页面只做入口承接。\nPortal 面向租户用户，Admin 面向平台运维与审计角色。\n若配置了 Keycloak，可使用统一 IAM 登录。`
          }}</textarea>
        </label>
        <label v-if="authConfig">
          <span>Keycloak 配置</span>
          <textarea readonly>{{
            `enabled: ${authConfig.enabled}\nrealm: ${authConfig.realm || '-'}\nclientId: ${authConfig.clientId || '-'}\nredirect: ${authConfig.defaultRedirect || '-'}`
          }}</textarea>
        </label>
      </div>
    </section>
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

.login-hero__stats .card {
  padding: 18px;
}

.login-hero__stats strong {
  display: block;
  font-size: 1.3rem;
}

.login-hero__stats span {
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
  gap: 12px;
  margin-top: 20px;
}

.primary,
.ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 150px;
  padding: 12px 18px;
  border-radius: 999px;
  border: 1px solid var(--stroke);
}

.primary {
  color: #fff;
  background: linear-gradient(120deg, var(--brand), var(--brand-strong));
  border-color: transparent;
}

.ghost {
  background: var(--panel-muted);
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

.login-form input,
.login-form textarea {
  width: 100%;
  padding: 14px 16px;
  border-radius: 14px;
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
  color: var(--text);
}

.login-form textarea {
  min-height: 128px;
  resize: none;
}

@media (max-width: 980px) {
  .login-shell {
    grid-template-columns: 1fr;
    padding: 18px;
  }

  .login-hero__stats {
    grid-template-columns: 1fr;
  }
}
</style>
