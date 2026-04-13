<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import ArchitectureDialog from './components/ArchitectureDialog.vue'
import ControlPlaneTabs from './components/ControlPlaneTabs.vue'
import MarketingMobileMenu from './components/MarketingMobileMenu.vue'
import QuickEntryPopover from './components/QuickEntryPopover.vue'

const router = useRouter()
const showArchitecture = ref(false)
const showMobileMenu = ref(false)

const valueProps = [
  {
    title: '能快',
    subtitle: '一键部署',
    description: '用平台壳层接住实例、渠道、IAM 与搜索配置，先把部署与入口做成稳定闭环。',
    icon: 'deploy'
  },
  {
    title: '够稳',
    subtitle: '原生体验',
    description: '保留 OpenClaw 的运行与调试能力，不在第一阶段重复造完整原生控制面。',
    icon: 'stable'
  },
  {
    title: '可控',
    subtitle: '企业级安全',
    description: '预留 Keycloak、审计检索、角色边界与运行状态可观测入口，便于后续接真环境。',
    icon: 'secure'
  }
]

const deliverySteps = [
  {
    number: '01',
    title: '部署平台运行层',
    description: '创建首个 OpenClaw 实例，建立基础运行环境'
  },
  {
    number: '02',
    title: '接入统一配置骨架',
    description: '统一身份、检索和渠道配置接线位就绪'
  },
  {
    number: '03',
    title: '分流两套控制台',
    description: '租户用户与平台运维分流到 Portal 和 Admin'
  }
]

const channelLogos = [
  { name: '飞书', color: '#2862ff' },
  { name: '企微', color: '#07c160' },
  { name: '钉钉', color: '#1890ff' },
  { name: 'Slack', color: '#4a154b' }
]
</script>

<template>
  <div class="home-shell">
    <header class="site-header">
      <RouterLink class="brand" to="/">
        <div class="brand-mark">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <span class="brand-text">openClaw</span>
      </RouterLink>
      <nav class="site-nav">
        <a href="#value" class="nav-link">价值主张</a>
        <a href="#control-planes" class="nav-link">双控制台</a>
        <a href="#channels" class="nav-link">渠道接入</a>
      </nav>
      <div class="site-actions">
        <QuickEntryPopover />
        <button class="nav-link-btn" type="button" @click="router.push('/login')">登录</button>
        <button class="nav-primary-btn" type="button" @click="router.push('/login')">立即开始</button>
      </div>
    </header>

    <main class="home-main">
      <section class="hero">
        <div class="hero-content">
          <div class="hero-badge">Phase 1 · 实例控制台先行</div>
          <h1 class="hero-title">能快 · 够稳 · 可控</h1>
          <p class="hero-description">
            用平台壳层接住实例、渠道、IAM 与搜索配置，同时保留 OpenClaw 原生调试能力。
            让部署先成为稳定闭环，原生控制面继续承接运维细节。
          </p>
          <div class="hero-actions">
            <button class="primary-btn" type="button" @click="router.push('/login')">进入控制台</button>
            <button class="secondary-btn" type="button" @click="showArchitecture = true">查看架构</button>
          </div>
        </div>
        <div class="hero-visual">
          <div class="hero-card hero-card-primary">
            <div class="card-header">
              <div class="card-badge">实例控制台</div>
              <button class="card-action" type="button">创建实例</button>
            </div>
            <div class="card-stats">
              <div class="stat-item">
                <span class="stat-label">运行中</span>
                <span class="stat-value">3</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">备份</span>
                <span class="stat-value">12</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">任务</span>
                <span class="stat-value">8</span>
              </div>
            </div>
            <div class="card-list">
              <div class="list-item">
                <div class="item-dot item-dot-green"></div>
                <div class="item-content">
                  <div class="item-title">生产环境实例</div>
                  <div class="item-meta">最后备份 2 小时前</div>
                </div>
                <button class="item-btn" type="button">访问</button>
              </div>
              <div class="list-item">
                <div class="item-dot item-dot-yellow"></div>
                <div class="item-content">
                  <div class="item-title">测试环境实例</div>
                  <div class="item-meta">任务执行中</div>
                </div>
                <button class="item-btn" type="button">访问</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="value" class="value-section">
        <div class="section-header">
          <h2 class="section-title">价值主张</h2>
          <p class="section-subtitle">官网先传达三件事</p>
        </div>
        <div class="value-grid">
          <div v-for="prop in valueProps" :key="prop.title" class="value-card">
            <div class="value-icon">
              <svg v-if="prop.icon === 'deploy'" width="32" height="32" viewBox="0 0 32 32" fill="none">
                <rect x="4" y="4" width="24" height="24" rx="6" stroke="currentColor" stroke-width="2"/>
                <path d="M16 10V22M10 16H22" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
              <svg v-else-if="prop.icon === 'stable'" width="32" height="32" viewBox="0 0 32 32" fill="none">
                <path d="M16 4L6 9V17C6 22 10 26 16 28C22 26 26 22 26 17V9L16 4Z" stroke="currentColor" stroke-width="2"/>
                <path d="M12 16L15 19L20 14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <svg v-else-if="prop.icon === 'secure'" width="32" height="32" viewBox="0 0 32 32" fill="none">
                <rect x="6" y="10" width="20" height="16" rx="2" stroke="currentColor" stroke-width="2"/>
                <path d="M10 10V8C10 5.23858 12.2386 3 15 3H17C19.7614 3 22 5.23858 22 8V10" stroke="currentColor" stroke-width="2"/>
                <circle cx="16" cy="18" r="2" fill="currentColor"/>
                <path d="M16 20V22" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </div>
            <div class="value-tag">{{ prop.subtitle }}</div>
            <h3 class="value-title">{{ prop.title }}</h3>
            <p class="value-description">{{ prop.description }}</p>
          </div>
        </div>
      </section>

      <section id="control-planes" class="control-planes-section">
        <div class="section-header">
          <h2 class="section-title">双控制台</h2>
          <p class="section-subtitle">租户入口和平台运维入口分开</p>
        </div>
        <ControlPlaneTabs />
      </section>

      <section id="channels" class="channels-section">
        <div class="section-header">
          <h2 class="section-title">统一接入</h2>
          <p class="section-subtitle">飞书、企微、钉钉、Slack 等入口集中配置</p>
        </div>
        <div class="channels-grid">
          <div v-for="channel in channelLogos" :key="channel.name" class="channel-card">
            <div class="channel-logo" :style="{ backgroundColor: channel.color + '20', color: channel.color }">
              {{ channel.name.charAt(0) }}
            </div>
            <span class="channel-name">{{ channel.name }}</span>
          </div>
        </div>
      </section>

      <section class="delivery-section">
        <div class="section-header">
          <h2 class="section-title">上线节奏</h2>
          <p class="section-subtitle">第一阶段不做大而全，先把实例控制与接入主链路跑通</p>
        </div>
        <div class="delivery-flow">
          <div v-for="(step, index) in deliverySteps" :key="step.number" class="delivery-step">
            <div class="step-number">{{ step.number }}</div>
            <div class="step-content">
              <h4>{{ step.title }}</h4>
              <p>{{ step.description }}</p>
            </div>
            <div v-if="index < deliverySteps.length - 1" class="step-arrow">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <path d="M5 12H19M19 12L12 5M19 12L12 19" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
          </div>
        </div>
      </section>

      <section class="cta-section">
        <div class="cta-card">
          <div class="cta-content">
            <h2>准备好开始了吗？</h2>
            <p>可信感亮色官网 + 登录入口 · 低学习成本 Portal 保持轻量 · 控制效率 Admin 深色控制面</p>
          </div>
          <div class="cta-actions">
            <button class="primary-btn" type="button" @click="router.push('/login')">立即开始</button>
          </div>
        </div>
      </section>

      <footer class="site-footer">
        <div class="footer-content">
          <div class="footer-brand">
            <h3>openClaw</h3>
            <p>实例控制台先行 · 统一渠道接入 · 双控制台职责清晰</p>
          </div>
          <div class="footer-links">
            <div class="footer-column">
              <h4>产品</h4>
              <a href="#value">价值主张</a>
              <a href="#control-planes">双控制台</a>
              <a href="#channels">渠道接入</a>
            </div>
            <div class="footer-column">
              <h4>资源</h4>
              <a href="#" @click.prevent="showArchitecture = true">架构文档</a>
              <a href="#">上线节奏</a>
            </div>
            <div class="footer-column">
              <h4>公司</h4>
              <a href="#">关于我们</a>
              <a href="#">联系我们</a>
            </div>
          </div>
        </div>
        <div class="footer-bottom">
          <p>&copy; 2024 openClaw. 后续扩展保留真实 IAM / Search / 渠道接线位。</p>
        </div>
      </footer>
    </main>

    <MarketingMobileMenu :open="showMobileMenu" @close="showMobileMenu = false" />
    <ArchitectureDialog :open="showArchitecture" @close="showArchitecture = false" />
  </div>
</template>

<style scoped>
.home-shell {
  min-height: 100vh;
  background: #f8fafc;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 2rem;
  background: white;
  border-bottom: 1px solid #e2e8f0;
  position: sticky;
  top: 0;
  z-index: 50;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: #0f172a;
  font-weight: 700;
  font-size: 1.25rem;
}

.brand-mark {
  display: flex;
  align-items: center;
  color: #1d6bff;
}

.brand-text {
  font-variation-settings: 'wght' 700;
}

.site-nav {
  display: flex;
  gap: 2rem;
}

.nav-link {
  color: #64748b;
  text-decoration: none;
  font-size: 0.9375rem;
  transition: color 0.2s;
}

.nav-link:hover {
  color: #1d6bff;
}

.site-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.nav-link-btn {
  background: transparent;
  border: 1px solid #e2e8f0;
  color: #475569;
  padding: 0.625rem 1.25rem;
  border-radius: 0.75rem;
  cursor: pointer;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.nav-link-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.nav-primary-btn {
  background: linear-gradient(120deg, #1d6bff, #5c7bff 65%, #1ad1ff);
  border: none;
  color: white;
  padding: 0.625rem 1.5rem;
  border-radius: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.nav-primary-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(29, 107, 255, 0.35);
}

.home-main {
  color: #0f172a;
}

.hero {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 4rem;
  padding: 5rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
  align-items: center;
}

.hero-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1d4ed8;
  padding: 0.5rem 1rem;
  border-radius: 999px;
  font-size: 0.875rem;
  font-weight: 600;
  width: fit-content;
}

.hero-title {
  font-size: 3.5rem;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.02em;
  background: linear-gradient(120deg, #0f172a, #1d6bff 50%, #1ad1ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-description {
  font-size: 1.125rem;
  color: #475569;
  line-height: 1.7;
  max-width: 580px;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  margin-top: 0.5rem;
}

.primary-btn {
  background: linear-gradient(120deg, #1d6bff, #5c7bff 65%, #1ad1ff);
  border: none;
  color: white;
  padding: 1rem 2rem;
  border-radius: 0.875rem;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.primary-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(29, 107, 255, 0.4);
}

.secondary-btn {
  background: white;
  border: 1px solid #e2e8f0;
  color: #475569;
  padding: 1rem 2rem;
  border-radius: 0.875rem;
  font-weight: 500;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.secondary-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.hero-visual {
  position: relative;
}

.hero-card {
  background: white;
  border-radius: 1.5rem;
  border: 1px solid #e2e8f0;
  box-shadow: 0 20px 60px -20px rgba(15, 23, 42, 0.15);
  overflow: hidden;
}

.hero-card-primary {
  padding: 1.5rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.card-badge {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  padding: 0.375rem 0.875rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 600;
}

.card-action {
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1d4ed8;
  padding: 0.5rem 1rem;
  border-radius: 0.625rem;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.card-action:hover {
  background: #dbeafe;
}

.card-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem;
  background: #f8fafc;
  border-radius: 1rem;
}

.stat-label {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 0.25rem;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
}

.card-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 1rem;
  background: #f8fafc;
  border-radius: 1rem;
}

.item-dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.item-dot-green {
  background: #22c55e;
}

.item-dot-yellow {
  background: #eab308;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-weight: 600;
  font-size: 0.9375rem;
  margin-bottom: 0.125rem;
}

.item-meta {
  font-size: 0.8125rem;
  color: #64748b;
}

.item-btn {
  background: white;
  border: 1px solid #e2e8f0;
  color: #475569;
  padding: 0.5rem 0.875rem;
  border-radius: 0.625rem;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.item-btn:hover {
  background: #f1f5f9;
}

.section-header {
  text-align: center;
  margin-bottom: 3rem;
}

.section-title {
  font-size: 2.5rem;
  font-weight: 800;
  margin-bottom: 0.75rem;
  letter-spacing: -0.02em;
}

.section-subtitle {
  font-size: 1.125rem;
  color: #64748b;
}

.value-section,
.control-planes-section,
.channels-section,
.delivery-section {
  padding: 5rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.value-section {
  background: white;
}

.value-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 2rem;
}

.value-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2rem;
  text-align: center;
  transition: all 0.2s;
}

.value-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px -12px rgba(15, 23, 42, 0.15);
  border-color: #cbd5e1;
}

.value-icon {
  width: 4rem;
  height: 4rem;
  display: grid;
  place-items: center;
  margin: 0 auto 1.25rem;
  background: linear-gradient(135deg, #eff6ff, #e0f2fe);
  color: #1d6bff;
  border-radius: 1.25rem;
}

.value-tag {
  display: inline-block;
  background: white;
  border: 1px solid #e2e8f0;
  padding: 0.375rem 0.875rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #475569;
  margin-bottom: 1rem;
}

.value-card h3 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.value-description {
  color: #64748b;
  line-height: 1.7;
  font-size: 0.9375rem;
}

.control-planes-section {
  background: #f8fafc;
}

.channels-section {
  background: white;
}

.channels-grid {
  display: flex;
  justify-content: center;
  gap: 2rem;
  flex-wrap: wrap;
}

.channel-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.875rem;
  padding: 2rem 2.5rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  transition: all 0.2s;
}

.channel-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px -12px rgba(15, 23, 42, 0.15);
}

.channel-logo {
  width: 4rem;
  height: 4rem;
  display: grid;
  place-items: center;
  border-radius: 1.25rem;
  font-size: 1.5rem;
  font-weight: 700;
}

.channel-name {
  font-weight: 600;
  font-size: 1rem;
  color: #334155;
}

.delivery-section {
  background: #f8fafc;
}

.delivery-flow {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 2rem;
}

.delivery-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  flex: 1;
  max-width: 280px;
}

.step-number {
  width: 3.5rem;
  height: 3.5rem;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #1d6bff, #5c7bff);
  color: white;
  border-radius: 1rem;
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.step-content h4 {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.step-content p {
  color: #64748b;
  line-height: 1.6;
  font-size: 0.9375rem;
}

.step-arrow {
  display: none;
}

.cta-section {
  padding: 5rem 2rem;
  background: linear-gradient(135deg, #0f172a, #1e293b);
}

.cta-card {
  max-width: 1000px;
  margin: 0 auto;
  background: linear-gradient(120deg, #1d6bff, #5c7bff 65%, #1ad1ff);
  border-radius: 2rem;
  padding: 3rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 2rem;
}

.cta-content h2 {
  color: white;
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.cta-content p {
  color: rgba(255, 255, 255, 0.85);
  font-size: 1rem;
  line-height: 1.6;
  max-width: 600px;
}

.cta-actions .primary-btn {
  background: white;
  color: #1d6bff;
}

.cta-actions .primary-btn:hover {
  box-shadow: 0 8px 24px rgba(255, 255, 255, 0.3);
}

.site-footer {
  background: white;
  border-top: 1px solid #e2e8f0;
  padding: 3rem 2rem 1.5rem;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 3rem;
  margin-bottom: 2rem;
}

.footer-brand h3 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.footer-brand p {
  color: #64748b;
  line-height: 1.6;
}

.footer-column {
  display: flex;
  flex-direction: column;
}

.footer-column h4 {
  font-weight: 600;
  margin-bottom: 1rem;
  color: #334155;
}

.footer-column a {
  color: #64748b;
  text-decoration: none;
  margin-bottom: 0.5rem;
  transition: color 0.2s;
}

.footer-column a:hover {
  color: #1d6bff;
}

.footer-bottom {
  border-top: 1px solid #e2e8f0;
  padding-top: 1.5rem;
  text-align: center;
  color: #64748b;
  font-size: 0.9375rem;
}

@media (max-width: 980px) {
  .site-header {
    padding: 1rem;
  }
  
  .site-nav {
    display: none;
  }
  
  .hero {
    grid-template-columns: 1fr;
    gap: 2.5rem;
    padding: 3rem 1rem;
  }
  
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-actions {
    flex-direction: column;
  }
  
  .value-grid {
    grid-template-columns: 1fr;
  }
  
  .delivery-flow {
    flex-direction: column;
    align-items: center;
  }
  
  .delivery-step {
    max-width: 100%;
  }
  
  .step-arrow {
    display: block;
    transform: rotate(90deg);
    margin: 1rem 0;
  }
  
  .cta-card {
    flex-direction: column;
    text-align: center;
    padding: 2rem;
  }
  
  .footer-content {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
}
</style>
