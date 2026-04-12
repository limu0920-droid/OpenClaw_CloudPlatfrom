<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import ArchitectureDialog from './components/ArchitectureDialog.vue'
import ControlPlaneTabs from './components/ControlPlaneTabs.vue'
import MarketingMobileMenu from './components/MarketingMobileMenu.vue'

const router = useRouter()
const showArchitecture = ref(false)
const showMobileMenu = ref(false)

const valueCards = [
  {
    title: '一键部署',
    text: '用平台壳层接住实例、渠道、IAM 与搜索配置，先把部署与入口做成稳定闭环。',
  },
  {
    title: '原生体验',
    text: '保留 OpenClaw 的运行与调试能力，不在第一阶段重复造完整原生控制面。',
  },
  {
    title: '企业级安全',
    text: '预留 Keycloak、审计检索、角色边界与运行状态可观测入口，便于后续接真环境。',
  },
]

const steps = [
  '部署平台运行层，创建首个 OpenClaw 实例。',
  '接入统一身份、检索和渠道配置骨架。',
  '将租户用户与平台运维分流到两套控制台。',
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
        <a href="#" class="nav-link">文档</a>
        <a href="#" class="nav-link">定价</a>
        <a href="#" class="nav-link">博客</a>
      </nav>
      <div class="site-actions">
        <button class="nav-link-btn" type="button" @click="router.push('/login')">登录</button>
        <button class="nav-primary-btn" type="button" @click="router.push('/login')">注册</button>
      </div>
    </header>

    <main class="home-main">
      <!-- Hero Section -->
      <section class="hero">
        <div class="hero-card hero-card-left">
          <div class="hero-badge">免费试用</div>
          <h1 class="hero-title">无限令牌，免费试用</h1>
          <p class="hero-description">
            体验 openClaw 的全部功能，享受无限令牌。无需信用卡。立即开始构建更智能的 AI 工作流程。
          </p>
          <div class="hero-actions">
            <button class="primary-btn" type="button" @click="router.push('/login')">开始免费试用</button>
            <button class="secondary-btn" type="button">查看演示</button>
          </div>
          <div class="hero-stats">
            <div class="stat">
              <span class="stat-value">10K+</span>
              <span class="stat-label">活跃用户</span>
            </div>
            <div class="stat">
              <span class="stat-value">99.9%</span>
              <span class="stat-label">可用性</span>
            </div>
            <div class="stat">
              <span class="stat-value">50+</span>
              <span class="stat-label">集成</span>
            </div>
          </div>
        </div>

        <div class="hero-card hero-card-right">
          <div class="deploy-icon">
            <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
              <rect x="4" y="4" width="40" height="40" rx="12" stroke="currentColor" stroke-width="2"/>
              <path d="M24 16V32M16 24H32" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <h2 class="deploy-title">一键部署</h2>
          <p class="deploy-description">
            秒级部署您的 AI 工作流程。无需配置，一键运行。
          </p>
          <button class="deploy-btn" type="button" @click="router.push('/login')">
            <span>立即部署</span>
          </button>
          <div class="deploy-features">
            <div class="feature-item">
              <span class="feature-dot"></span>
              <span>即时配置</span>
            </div>
            <div class="feature-item">
              <span class="feature-dot"></span>
              <span>自动扩缩</span>
            </div>
            <div class="feature-item">
              <span class="feature-dot"></span>
              <span>零维护</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Features Section -->
      <section class="features">
        <div class="feature-card">
          <div class="feature-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
              <rect x="3" y="3" width="18" height="18" rx="3" stroke="currentColor" stroke-width="2"/>
              <path d="M9 9H15M9 12H15M9 15H12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <h3 class="feature-title">应用场景</h3>
          <p class="feature-description">
            探索 openClaw 如何从内容生成到数据处理的各种应用场景提供动力。
          </p>
          <a href="#" class="feature-link">
            了解更多
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M2 7H12M12 7L8 3M12 7L8 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </a>
        </div>

        <div class="feature-card">
          <div class="feature-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
              <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <h3 class="feature-title">优势亮点</h3>
          <p class="feature-description">
            为什么选择 openClaw？卓越性能、企业级安全性和无与伦比的定价。
          </p>
          <a href="#" class="feature-link">
            了解更多
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M2 7H12M12 7L8 3M12 7L8 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </a>
        </div>

        <div class="feature-card feature-card-empty">
          <div class="feature-placeholder">
            <span>更多功能即将推出</span>
          </div>
        </div>
      </section>

      <!-- Value Props Section -->
      <section class="value-props">
        <div class="value-header">
          <h2 class="value-title">为开发者构建</h2>
          <p class="value-subtitle">
            针对开发者和企业团队设计的 AI 工作流程平台。
          </p>
        </div>
        <div class="value-cards">
          <div class="value-card" v-for="card in valueCards" :key="card.title">
            <h3>{{ card.title }}</h3>
            <p>{{ card.text }}</p>
          </div>
        </div>
      </section>

      <!-- How It Works Section -->
      <section class="how-it-works">
        <div class="comparison-header">
          <h2 class="comparison-title">三步开始使用 openClaw</h2>
          <p class="comparison-subtitle">
            无需复杂配置，几分钟内开始运行您的 AI 工作流程。
          </p>
        </div>
        <div class="comparison-content">
          <div class="workflow-step">
            <div class="step-number">01</div>
            <div class="step-content">
              <h4>部署平台运行层</h4>
              <p>连接您的数据源和 API</p>
            </div>
          </div>
          <div class="workflow-arrow">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
              <path d="M5 12H19M19 12L12 5M19 12L12 19" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <div class="workflow-step">
            <div class="step-number">02</div>
            <div class="step-content">
              <h4>配置工作流程</h4>
              <p>使用直观的界面定义您的 AI 工作流程</p>
            </div>
          </div>
          <div class="workflow-arrow">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
              <path d="M5 12H19M19 12L12 5M19 12L12 19" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <div class="workflow-step">
            <div class="step-number">03</div>
            <div class="step-content">
              <h4>监控与优化</h4>
              <p>实时监控工作流程性能并进行优化</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Use Cases Section -->
      <section class="use-cases">
        <div class="use-cases-header">
          <h2 class="use-cases-title">精选案例</h2>
          <p class="use-cases-subtitle">
            学习 openClaw 如何赋能各种业务场景。
          </p>
        </div>
        <div class="use-cases-grid">
          <div class="use-case">
            <div class="use-case-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <path d="M12 6V10M12 10L8 10M12 10L12 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h3>内容生成</h3>
            <p>自动化内容创作和个性化推荐</p>
          </div>
          <div class="use-case">
            <div class="use-case-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <rect x="3" y="3" width="18" height="18" rx="3" stroke="currentColor" stroke-width="2"/>
                <path d="M15 12H9M12 15V9" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </div>
            <h3>数据分析</h3>
            <p>智能分析和预测性洞察</p>
          </div>
          <div class="use-case">
            <div class="use-case-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <path d="M20 21V19C20 17.9391 19.5786 16.9217 18.8284 16.1716C18.0783 15.4214 17.0609 15 16 15H8C6.93913 15 5.92172 15.4214 5.17157 16.1716C4.42143 16.9217 4 17.9391 4 19V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <rect x="4" y="3" width="16" height="12" rx="3" stroke="currentColor" stroke-width="2"/>
              </svg>
            </div>
            <h3>客户支持</h3>
            <p>24/7 智能客服和自助服务</p>
          </div>
          <div class="use-case">
            <div class="use-case-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <rect x="6" y="6" width="12" height="12" rx="3" stroke="currentColor" stroke-width="2"/>
                <path d="M9 3V6M15 3V6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                <path d="M6 9H18M6 15H18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </div>
            <h3>文档处理</h3>
            <p>自动化文档处理和知识管理</p>
          </div>
        </div>
      </section>

      <!-- Architecture Section -->
      <section class="architecture">
        <div class="architecture-header">
          <h2 class="architecture-title">技术架构</h2>
          <p class="architecture-subtitle">
            基于现代技术栈构建的可扩展平台
          </p>
          <button class="architecture-btn" type="button" @click="showArchitecture = true">
            查看架构详情
          </button>
        </div>
        <div class="architecture-preview">
          <svg width="100%" height="200" viewBox="0 0 400 200" fill="none">
            <!-- Simplified architecture diagram -->
            <rect x="50" y="50" width="300" height="100" rx="10" fill="transparent" stroke="currentColor" stroke-width="2"/>
            <text x="200" y="110" text-anchor="middle" fill="currentColor" font-size="14">openClaw Platform</text>
            <path d="M0 100L50 100" stroke="currentColor" stroke-width="2"/>
            <path d="M350 100L400 100" stroke="currentColor" stroke-width="2"/>
            <circle cx="30" cy="100" r="8" fill="currentColor"/>
            <circle cx="370" cy="100" r="8" fill="currentColor"/>
          </svg>
        </div>
      </section>

      <!-- Footer Section -->
      <footer class="site-footer">
        <div class="footer-content">
          <div class="footer-brand">
            <h3>openClaw</h3>
            <p>为开发者和企业打造的智能 AI 工作流程平台</p>
          </div>
          <div class="footer-links">
            <div class="footer-column">
              <h4>产品</h4>
              <a href="#">功能特性</a>
              <a href="#">定价</a>
              <a href="#">文档</a>
            </div>
            <div class="footer-column">
              <h4>公司</h4>
              <a href="#">关于我们</a>
              <a href="#">联系我们</a>
              <a href="#">博客</a>
            </div>
            <div class="footer-column">
              <h4>支持</h4>
              <a href="#">帮助中心</a>
              <a href="#">社区论坛</a>
              <a href="#">状态</a>
            </div>
          </div>
        </div>
        <div class="footer-bottom">
          <p>&copy; 2024 openClaw. 保留所有权利。</p>
        </div>
      </footer>
    </main>

    <!-- Mobile Menu -->
    <MarketingMobileMenu v-model:show="showMobileMenu" />

    <!-- Architecture Dialog -->
    <ArchitectureDialog v-model:show="showArchitecture" />
  </div>
</template>

<style scoped>
.home-shell {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 2rem;
  color: white;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  color: white;
  font-weight: 600;
  font-size: 1.25rem;
}

.brand-mark {
  display: flex;
  align-items: center;
}

.site-nav {
  display: flex;
  gap: 2rem;
}

.nav-link {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  transition: color 0.2s;
}

.nav-link:hover {
  color: white;
}

.site-actions {
  display: flex;
  gap: 1rem;
}

.nav-link-btn {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
}

.nav-link-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.nav-primary-btn {
  background: white;
  border: none;
  color: #667eea;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.nav-primary-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.home-main {
  color: white;
}

.hero {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4rem;
  padding: 4rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.hero-card {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.hero-badge {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  width: fit-content;
}

.hero-title {
  font-size: 3rem;
  font-weight: 700;
  line-height: 1.2;
}

.hero-description {
  font-size: 1.125rem;
  opacity: 0.8;
  line-height: 1.6;
}

.hero-actions {
  display: flex;
  gap: 1rem;
}

.primary-btn {
  background: #10b981;
  border: none;
  color: white;
  padding: 1rem 2rem;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.primary-btn:hover {
  background: #059669;
  transform: translateY(-1px);
}

.secondary-btn {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 1rem 2rem;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
}

.secondary-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.hero-stats {
  display: flex;
  gap: 2rem;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
}

.stat-label {
  font-size: 0.875rem;
  opacity: 0.8;
}

.deploy-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
}

.deploy-title {
  font-size: 1.5rem;
  font-weight: 600;
}

.deploy-description {
  opacity: 0.8;
  line-height: 1.6;
}

.deploy-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.deploy-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.deploy-features {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.feature-dot {
  width: 0.5rem;
  height: 0.5rem;
  background: #10b981;
  border-radius: 50%;
}

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  padding: 4rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.feature-card {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 1rem;
  padding: 2rem;
}

.feature-card-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  border-style: dashed;
}

.feature-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 0.75rem;
  margin-bottom: 1rem;
}

.feature-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.feature-description {
  opacity: 0.8;
  line-height: 1.6;
  margin-bottom: 1rem;
}

.feature-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: white;
  text-decoration: none;
  font-weight: 600;
}

.feature-link:hover {
  text-decoration: underline;
}

.value-props {
  padding: 4rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.value-header {
  text-align: center;
  margin-bottom: 3rem;
}

.value-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.value-subtitle {
  font-size: 1.125rem;
  opacity: 0.8;
  max-width: 600px;
  margin: 0 auto;
}

.value-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.value-card {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 1rem;
  padding: 2rem;
}

.value-card h3 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.value-card p {
  opacity: 0.8;
  line-height: 1.6;
}

.how-it-works {
  padding: 4rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.comparison-header {
  text-align: center;
  margin-bottom: 3rem;
}

.comparison-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.comparison-subtitle {
  font-size: 1.125rem;
  opacity: 0.8;
  max-width: 600px;
  margin: 0 auto;
}

.comparison-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2rem;
}

.workflow-step {
  text-align: center;
  flex: 1;
  max-width: 200px;
}

.step-number {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  width: 3rem;
  height: 3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  margin: 0 auto 1rem;
}

.step-content h4 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.step-content p {
  opacity: 0.8;
  font-size: 0.875rem;
}

.workflow-arrow {
  display: flex;
  align-items: center;
}

.use-cases {
  padding: 4rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.use-cases-header {
  text-align: center;
  margin-bottom: 3rem;
}

.use-cases-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.use-cases-subtitle {
  font-size: 1.125rem;
  opacity: 0.8;
  max-width: 600px;
  margin: 0 auto;
}

.use-cases-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
}

.use-case {
  text-align: center;
  padding: 2rem;
}

.use-case-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
  margin: 0 auto 1rem;
}

.use-case h3 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.use-case p {
  opacity: 0.8;
  line-height: 1.6;
}

.architecture {
  padding: 4rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
}

.architecture-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.architecture-subtitle {
  font-size: 1.125rem;
  opacity: 0.8;
  margin-bottom: 2rem;
}

.architecture-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 1rem 2rem;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.architecture-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.architecture-preview {
  margin-top: 2rem;
}

.site-footer {
  background: rgba(0, 0, 0, 0.2);
  padding: 3rem 2rem 1rem;
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
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.footer-brand p {
  opacity: 0.8;
}

.footer-column {
  display: flex;
  flex-direction: column;
}

.footer-column h4 {
  font-weight: 600;
  margin-bottom: 1rem;
}

.footer-column a {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  margin-bottom: 0.5rem;
  transition: color 0.2s;
}

.footer-column a:hover {
  color: white;
}

.footer-bottom {
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 1rem;
  text-align: center;
  opacity: 0.8;
}

@media (max-width: 768px) {
  .site-header {
    padding: 1rem;
  }
  
  .site-nav {
    display: none;
  }
  
  .hero {
    grid-template-columns: 1fr;
    gap: 2rem;
    padding: 2rem 1rem;
  }
  
  .hero-title {
    font-size: 2rem;
  }
  
  .hero-actions {
    flex-direction: column;
  }
  
  .hero-stats {
    justify-content: center;
  }
  
  .comparison-content {
    flex-direction: column;
    gap: 1rem;
  }
  
  .footer-content {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
}
</style>