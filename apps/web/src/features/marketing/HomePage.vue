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
</script>

<template>
  <div class="home-page">
    <header class="header">
      <RouterLink class="logo" to="/">
        <div class="logo-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <span>openClaw</span>
      </RouterLink>
      <nav class="nav">
        <a href="#features" class="nav-link">核心特性</a>
        <a href="#consoles" class="nav-link">双控制台</a>
        <a href="#channels" class="nav-link">统一接入</a>
        <a href="#steps" class="nav-link">上线节奏</a>
      </nav>
      <div class="header-actions">
        <QuickEntryPopover />
        <button class="btn-outline" type="button" @click="router.push('/login')">登录</button>
        <button class="btn-primary" type="button" @click="router.push('/login')">立即开始</button>
      </div>
    </header>

    <main class="main">
      <section class="hero">
        <div class="hero-content">
          <div class="phase-badge">Phase 1 · 实例控制台先行</div>
          <h1 class="hero-title">能快 · 够稳 · 可控</h1>
          <p class="hero-desc">
            用平台壳层接住实例、渠道、IAM 与搜索配置，同时保留 OpenClaw 原生调试能力。
          </p>
          <div class="hero-buttons">
            <button class="btn-large" type="button" @click="router.push('/login')">进入控制台</button>
            <button class="btn-text" type="button" @click="showArchitecture = true">查看架构 →</button>
          </div>
        </div>
        <div class="hero-visual">
          <div class="preview-card">
            <div class="preview-header">
              <span class="preview-tag">实例控制台</span>
              <button class="preview-btn" type="button">创建实例</button>
            </div>
            <div class="preview-stats">
              <div class="stat">
                <span class="stat-label">运行中</span>
                <span class="stat-value">3</span>
              </div>
              <div class="stat">
                <span class="stat-label">备份</span>
                <span class="stat-value">12</span>
              </div>
              <div class="stat">
                <span class="stat-label">任务</span>
                <span class="stat-value">8</span>
              </div>
            </div>
            <div class="preview-list">
              <div class="list-item">
                <div class="dot green"></div>
                <div class="item-info">
                  <div class="item-name">生产环境实例</div>
                  <div class="item-meta">最后备份 2 小时前</div>
                </div>
                <button class="item-btn" type="button">访问</button>
              </div>
              <div class="list-item">
                <div class="dot yellow"></div>
                <div class="item-info">
                  <div class="item-name">测试环境实例</div>
                  <div class="item-meta">任务执行中</div>
                </div>
                <button class="item-btn" type="button">访问</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="features" class="section features">
        <div class="section-head">
          <h2>核心特性</h2>
          <p>官网先传达三件事</p>
        </div>
        <div class="features-grid">
          <div class="feature-card">
            <div class="feature-icon">
              <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
                <rect x="4" y="4" width="24" height="24" rx="6" stroke="currentColor" stroke-width="2"/>
                <path d="M16 10V22M10 16H22" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </div>
            <div class="feature-tag">一键部署</div>
            <h3>能快</h3>
            <p>用平台壳层接住实例、渠道、IAM 与搜索配置，先把部署与入口做成稳定闭环。</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">
              <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
                <path d="M16 4L6 9V17C6 22 10 26 16 28C22 26 26 22 26 17V9L16 4Z" stroke="currentColor" stroke-width="2"/>
                <path d="M12 16L15 19L20 14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <div class="feature-tag">原生体验</div>
            <h3>够稳</h3>
            <p>保留 OpenClaw 的运行与调试能力，不在第一阶段重复造完整原生控制面。</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">
              <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
                <rect x="6" y="10" width="20" height="16" rx="2" stroke="currentColor" stroke-width="2"/>
                <path d="M10 10V8C10 5.23858 12.2386 3 15 3H17C19.7614 3 22 5.23858 22 8V10" stroke="currentColor" stroke-width="2"/>
                <circle cx="16" cy="18" r="2" fill="currentColor"/>
                <path d="M16 20V22" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </div>
            <div class="feature-tag">企业级安全</div>
            <h3>可控</h3>
            <p>预留 Keycloak、审计检索、角色边界与运行状态可观测入口，便于后续接真环境。</p>
          </div>
        </div>
      </section>

      <section id="consoles" class="section consoles">
        <div class="section-head">
          <h2>双控制台</h2>
          <p>租户入口和平台运维入口分开</p>
        </div>
        <ControlPlaneTabs />
      </section>

      <section id="channels" class="section channels">
        <div class="section-head">
          <h2>统一接入</h2>
          <p>飞书、企微、钉钉、Slack 等入口集中配置</p>
        </div>
        <div class="channels-list">
          <div class="channel-item">
            <div class="channel-logo" style="background: #eff6ff; color: #2862ff;">飞</div>
            <span>飞书</span>
          </div>
          <div class="channel-item">
            <div class="channel-logo" style="background: #f0fdf4; color: #07c160;">企</div>
            <span>企微</span>
          </div>
          <div class="channel-item">
            <div class="channel-logo" style="background: #eff6ff; color: #1890ff;">钉</div>
            <span>钉钉</span>
          </div>
          <div class="channel-item">
            <div class="channel-logo" style="background: #f3e8ff; color: #4a154b;">S</div>
            <span>Slack</span>
          </div>
        </div>
      </section>

      <section id="steps" class="section steps">
        <div class="section-head">
          <h2>上线节奏</h2>
          <p>第一阶段不做大而全，先把实例控制与接入主链路跑通</p>
        </div>
        <div class="steps-flow">
          <div class="step">
            <div class="step-num">01</div>
            <h4>部署平台运行层</h4>
            <p>创建首个 OpenClaw 实例，建立基础运行环境</p>
          </div>
          <div class="step-arrow">→</div>
          <div class="step">
            <div class="step-num">02</div>
            <h4>接入统一配置骨架</h4>
            <p>统一身份、检索和渠道配置接线位就绪</p>
          </div>
          <div class="step-arrow">→</div>
          <div class="step">
            <div class="step-num">03</div>
            <h4>分流两套控制台</h4>
            <p>租户用户与平台运维分流到 Portal 和 Admin</p>
          </div>
        </div>
      </section>
    </main>

    <footer class="footer">
      <div class="footer-content">
        <div class="footer-brand">
          <h3>openClaw</h3>
          <p>实例控制台先行 · 统一渠道接入 · 双控制台职责清晰</p>
        </div>
        <div class="footer-links">
          <div class="footer-col">
            <h4>产品</h4>
            <a href="#features">核心特性</a>
            <a href="#consoles">双控制台</a>
            <a href="#channels">渠道接入</a>
          </div>
          <div class="footer-col">
            <h4>资源</h4>
            <a href="#" @click.prevent="showArchitecture = true">架构文档</a>
            <a href="#steps">上线节奏</a>
          </div>
          <div class="footer-col">
            <h4>联系</h4>
            <a href="mailto:contact@openclaw.dev">contact@openclaw.dev</a>
          </div>
        </div>
      </div>
      <div class="footer-bottom">
        <p>&copy; 2024 openClaw. 后续扩展保留真实 IAM / Search / 渠道接线位。</p>
      </div>
    </footer>

    <MarketingMobileMenu :open="showMobileMenu" @close="showMobileMenu = false" />
    <ArchitectureDialog :open="showArchitecture" @close="showArchitecture = false" />
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #f8fafc;
  color: #0f172a;
}

.header {
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

.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: #0f172a;
  font-weight: 700;
  font-size: 1.25rem;
}

.logo-icon {
  color: #1d6bff;
  display: flex;
  align-items: center;
}

.nav {
  display: flex;
  gap: 1.5rem;
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

.header-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.btn-outline {
  background: transparent;
  border: 1px solid #e2e8f0;
  color: #475569;
  padding: 0.625rem 1.25rem;
  border-radius: 0.75rem;
  cursor: pointer;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.btn-outline:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.btn-primary {
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

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(29, 107, 255, 0.35);
}

.main {
  max-width: 1200px;
  margin: 0 auto;
}

.hero {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 4rem;
  padding: 5rem 2rem;
  align-items: center;
}

.hero-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.phase-badge {
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

.hero-desc {
  font-size: 1.125rem;
  color: #475569;
  line-height: 1.7;
  max-width: 580px;
}

.hero-buttons {
  display: flex;
  gap: 1rem;
  margin-top: 0.5rem;
}

.btn-large {
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

.btn-large:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(29, 107, 255, 0.4);
}

.btn-text {
  background: transparent;
  border: none;
  color: #475569;
  padding: 1rem 2rem;
  font-weight: 500;
  font-size: 1rem;
  cursor: pointer;
  transition: color 0.2s;
}

.btn-text:hover {
  color: #1d6bff;
}

.hero-visual {
  position: relative;
}

.preview-card {
  background: white;
  border-radius: 1.5rem;
  border: 1px solid #e2e8f0;
  box-shadow: 0 20px 60px -20px rgba(15, 23, 42, 0.15);
  padding: 1.5rem;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.preview-tag {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  padding: 0.375rem 0.875rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 600;
}

.preview-btn {
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

.preview-btn:hover {
  background: #dbeafe;
}

.preview-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat {
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

.preview-list {
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

.dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot.green {
  background: #22c55e;
}

.dot.yellow {
  background: #eab308;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
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

.section {
  padding: 5rem 2rem;
}

.section-head {
  text-align: center;
  margin-bottom: 3rem;
}

.section-head h2 {
  font-size: 2.5rem;
  font-weight: 800;
  margin-bottom: 0.75rem;
  letter-spacing: -0.02em;
}

.section-head p {
  font-size: 1.125rem;
  color: #64748b;
}

.features {
  background: white;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 2rem;
}

.feature-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2rem;
  text-align: center;
  transition: all 0.2s;
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px -12px rgba(15, 23, 42, 0.15);
  border-color: #cbd5e1;
}

.feature-icon {
  width: 4rem;
  height: 4rem;
  display: grid;
  place-items: center;
  margin: 0 auto 1.25rem;
  background: linear-gradient(135deg, #eff6ff, #e0f2fe);
  color: #1d6bff;
  border-radius: 1.25rem;
}

.feature-tag {
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

.feature-card h3 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.feature-card p {
  color: #64748b;
  line-height: 1.7;
  font-size: 0.9375rem;
}

.consoles {
  background: #f8fafc;
}

.channels {
  background: white;
}

.channels-list {
  display: flex;
  justify-content: center;
  gap: 2rem;
  flex-wrap: wrap;
}

.channel-item {
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

.channel-item:hover {
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

.channel-item span {
  font-weight: 600;
  font-size: 1rem;
  color: #334155;
}

.steps {
  background: #f8fafc;
}

.steps-flow {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  flex: 1;
  max-width: 280px;
}

.step-num {
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

.step h4 {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.step p {
  color: #64748b;
  line-height: 1.6;
  font-size: 0.9375rem;
}

.step-arrow {
  font-size: 1.5rem;
  color: #cbd5e1;
  font-weight: 700;
}

.footer {
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

.footer-col {
  display: flex;
  flex-direction: column;
}

.footer-col h4 {
  font-weight: 600;
  margin-bottom: 1rem;
  color: #334155;
}

.footer-col a {
  color: #64748b;
  text-decoration: none;
  margin-bottom: 0.5rem;
  transition: color 0.2s;
}

.footer-col a:hover {
  color: #1d6bff;
}

.footer-bottom {
  border-top: 1px solid #e2e8f0;
  padding-top: 1.5rem;
  text-align: center;
  color: #64748b;
  font-size: 0.9375rem;
  max-width: 1200px;
  margin: 0 auto;
}

@media (max-width: 980px) {
  .header {
    padding: 1rem;
  }

  .nav {
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

  .hero-buttons {
    flex-direction: column;
  }

  .features-grid {
    grid-template-columns: 1fr;
  }

  .steps-flow {
    flex-direction: column;
    gap: 1rem;
  }

  .step {
    max-width: 100%;
  }

  .step-arrow {
    transform: rotate(90deg);
  }

  .footer-content {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
}
</style>
