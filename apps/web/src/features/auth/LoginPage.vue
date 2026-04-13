<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import { api } from '../../lib/api'
import { useBranding } from '../../lib/brand'
import { useAsyncData } from '../../lib/useAsyncData'

const router = useRouter()
const route = useRoute()
const { brand } = useBranding()
const loading = ref(false)
const error = ref('')

const { data: authConfig } = useAsyncData(() => api.getAuthConfig())
const { data: authSession } = useAsyncData(() => api.getAuthSession())

onMounted(() => {
  const code = route.query.code as string
  const state = route.query.state as string
  
  if (code) {
    handleAuthCallback(code, window.location.origin + '/login')
  }
})

async function handleAuthCallback(code: string, redirectUri: string) {
  loading.value = true
  error.value = ''
  try {
    await api.exchangeAuthCode(code, redirectUri)
    await router.push('/portal')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '登录失败，请重试'
  } finally {
    loading.value = false
  }
}

async function openKeycloak() {
  loading.value = true
  error.value = ''
  try {
    const response = await api.getKeycloakLoginURL(window.location.origin + '/login')
    window.location.href = response.url
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取登录链接失败'
    loading.value = false
  }
}

async function openWechat() {
  loading.value = true
  error.value = ''
  try {
    const response = await api.getWechatLoginURL(window.location.origin + '/login')
    window.location.href = response.url
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取微信登录链接失败'
    loading.value = false
  }
}

function goToPortal() {
  router.push('/portal')
}

function goToAdmin() {
  router.push('/admin')
}
</script>

<template>
  <div class="login-page">
    <section class="login-hero">
      <div class="hero-content">
        <div class="logo-section">
          <div class="logo-icon">
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
              <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <h1>{{ brand.name }}</h1>
        </div>
        <p class="hero-desc">
          一站式实例管理与协作平台，提供统一的多渠道接入、双控制台架构和原生 OpenClaw 实例运行环境。
        </p>
        <div class="hero-features">
          <div class="feature">
            <div class="feature-icon">🚀</div>
            <span>一键部署实例</span>
          </div>
          <div class="feature">
            <div class="feature-icon">🔐</div>
            <span>企业级安全</span>
          </div>
          <div class="feature">
            <div class="feature-icon">💬</div>
            <span>多渠道接入</span>
          </div>
        </div>
      </div>
    </section>

    <section class="login-panel">
      <div class="panel-header">
        <h2>登录</h2>
        <p>选择登录方式继续</p>
      </div>

      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <div class="login-methods">
        <button 
          class="login-btn keycloak-btn" 
          type="button" 
          @click="openKeycloak"
          :disabled="loading || !authConfig?.enabled"
        >
          <div class="btn-icon">🔑</div>
          <div class="btn-content">
            <strong>Keycloak 统一认证</strong>
            <span>企业 SSO 登录</span>
          </div>
        </button>

        <button 
          class="login-btn wechat-btn" 
          type="button" 
          @click="openWechat"
          :disabled="loading"
        >
          <div class="btn-icon">💚</div>
          <div class="btn-content">
            <strong>微信登录</strong>
            <span>扫码快速登录</span>
          </div>
        </button>
      </div>

      <div class="divider">
        <span>或者</span>
      </div>

      <div class="quick-access">
        <button 
          class="access-btn portal-btn" 
          type="button" 
          @click="goToPortal"
        >
          <div class="access-icon">🏢</div>
          <div class="access-content">
            <strong>进入 Portal</strong>
            <span>租户用户控制台</span>
          </div>
        </button>

        <button 
          class="access-btn admin-btn" 
          type="button" 
          @click="goToAdmin"
        >
          <div class="access-icon">⚙️</div>
          <div class="access-content">
            <strong>进入 Admin</strong>
            <span>平台管理控制台</span>
          </div>
        </button>
      </div>

      <div v-if="authSession" class="session-info">
        <div class="info-header">
          <span>当前会话</span>
        </div>
        <div class="info-content">
          <div class="info-item">
            <span class="label">认证状态</span>
            <span class="value">{{ authSession.authenticated ? '已认证' : '未认证' }}</span>
          </div>
          <div v-if="authSession.authenticated" class="info-item">
            <span class="label">用户</span>
            <span class="value">{{ authSession.user?.name || '-' }}</span>
          </div>
          <div v-if="authSession.authenticated" class="info-item">
            <span class="label">角色</span>
            <span class="value">{{ authSession.user?.role || '-' }}</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
}

.login-hero {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 3rem;
  background: 
    radial-gradient(circle at top left, rgba(29, 107, 255, 0.1), transparent 40%),
    radial-gradient(circle at bottom right, rgba(26, 209, 255, 0.1), transparent 40%),
    white;
}

.hero-content {
  max-width: 600px;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
}

.logo-icon {
  color: #1d6bff;
  display: flex;
  align-items: center;
}

.logo-section h1 {
  font-size: 2.5rem;
  font-weight: 800;
  margin: 0;
  background: linear-gradient(120deg, #0f172a, #1d6bff 50%, #1ad1ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-desc {
  font-size: 1.125rem;
  color: #475569;
  line-height: 1.7;
  margin: 0 0 2rem 0;
}

.hero-features {
  display: flex;
  gap: 2rem;
}

.feature {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.9375rem;
  color: #64748b;
}

.feature-icon {
  font-size: 1.5rem;
}

.login-panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 3rem;
  background: white;
}

.panel-header {
  margin-bottom: 2rem;
}

.panel-header h2 {
  font-size: 2rem;
  font-weight: 800;
  margin: 0 0 0.5rem 0;
  color: #0f172a;
}

.panel-header p {
  font-size: 0.9375rem;
  color: #64748b;
  margin: 0;
}

.error-message {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 1rem;
  border-radius: 0.75rem;
  margin-bottom: 1.5rem;
  font-size: 0.9375rem;
}

.login-methods {
  display: grid;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.login-btn {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: 1rem;
  border: 2px solid #e2e8f0;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.login-btn:hover:not(:disabled) {
  border-color: #1d6bff;
  background: #eff6ff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(29, 107, 255, 0.15);
}

.login-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  font-size: 1.75rem;
  flex-shrink: 0;
}

.btn-content {
  flex: 1;
}

.btn-content strong {
  display: block;
  font-size: 1rem;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 0.25rem;
}

.btn-content span {
  display: block;
  font-size: 0.875rem;
  color: #64748b;
}

.divider {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin: 1.5rem 0;
  color: #cbd5e1;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #e2e8f0;
}

.divider span {
  font-size: 0.875rem;
  color: #94a3b8;
}

.quick-access {
  display: grid;
  gap: 1rem;
  margin-bottom: 2rem;
}

.access-btn {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: 1rem;
  border: 2px solid #e2e8f0;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.access-btn:hover {
  border-color: #1d6bff;
  background: #eff6ff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(29, 107, 255, 0.15);
}

.access-icon {
  font-size: 1.75rem;
  flex-shrink: 0;
}

.access-content {
  flex: 1;
}

.access-content strong {
  display: block;
  font-size: 1rem;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 0.25rem;
}

.access-content span {
  display: block;
  font-size: 0.875rem;
  color: #64748b;
}

.session-info {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  padding: 1.25rem;
}

.info-header {
  font-size: 0.875rem;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 1rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-content {
  display: grid;
  gap: 0.75rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-item .label {
  font-size: 0.875rem;
  color: #64748b;
}

.info-item .value {
  font-size: 0.875rem;
  font-weight: 600;
  color: #0f172a;
}

@media (max-width: 980px) {
  .login-page {
    grid-template-columns: 1fr;
  }

  .login-hero {
    padding: 2rem;
  }

  .login-panel {
    padding: 2rem;
  }

  .logo-section h1 {
    font-size: 2rem;
  }

  .hero-features {
    flex-wrap: wrap;
  }
}
</style>
