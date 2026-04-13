<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import ArchitectureDialog from './components/ArchitectureDialog.vue'
import MarketingMobileMenu from './components/MarketingMobileMenu.vue'

const router = useRouter()
const showArchitecture = ref(false)
const showMobileMenu = ref(false)

const scrollToSection = (id: string) => {
  const element = document.getElementById(id)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth' })
  }
}
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
        <span>OpenClaw</span>
      </RouterLink>
      <div class="header-spacer"></div>
      <div class="header-actions">
        <button class="btn-outline" type="button" @click="router.push('/login')">登录</button>
        <button class="btn-primary" type="button" @click="router.push('/login')">注册</button>
      </div>
    </header>

    <main class="main">
      <section class="promo-section">
        <div class="promo-cards">
          <div class="promo-card primary">
            <h3>无限 Token</h3>
            <p>免费试用</p>
          </div>
          <div class="promo-card secondary">
            <h3>精英人式一键部署</h3>
            <p>即刻体验</p>
          </div>
        </div>
      </section>

      <section class="info-section">
        <div class="info-cards">
          <div class="info-card" @click="router.push('/login')">
            <h3>使用方法</h3>
            <div class="arrow">→</div>
          </div>
          <div class="info-card" @click="scrollToSection('advantages')">
            <h3>优势</h3>
            <div class="arrow">→</div>
          </div>
          <div class="info-card" @click="scrollToSection('compare')">
            <h3>对比</h3>
            <div class="arrow">→</div>
          </div>
        </div>
      </section>

      <section id="compare" class="compare-section">
        <div class="compare-container">
          <div class="compare-left">
            <h2>对比：流程点</h2>
            <div class="compare-content">
              <div class="compare-item">
                <div class="checkmark">✓</div>
                <span>完整的一体化平台架构</span>
              </div>
              <div class="compare-item">
                <div class="checkmark">✓</div>
                <span>企业级功能完备性</span>
              </div>
              <div class="compare-item">
                <div class="checkmark">✓</div>
                <span>严格的生产可靠性</span>
              </div>
              <div class="compare-item">
                <div class="checkmark">✓</div>
                <span>强大的集成与扩展能力</span>
              </div>
              <div class="compare-item">
                <div class="checkmark">✓</div>
                <span>实时通信与事件处理</span>
              </div>
            </div>
          </div>
          <div class="compare-right">
            <h3>最新作用</h3>
            <div class="latest-features">
              <div class="feature-tag">实时通信</div>
              <div class="feature-tag">SSE事件流</div>
              <div class="feature-tag">异步回调</div>
              <div class="feature-tag">企业安全</div>
              <div class="feature-tag">多渠道接入</div>
              <div class="feature-tag">容器编排</div>
            </div>
          </div>
        </div>
      </section>

      <section id="advantages" class="advantages-section">
        <div class="section-head">
          <h2>核心优势</h2>
          <p>OpenClaw 商业化项目的核心竞争力</p>
        </div>
        <div class="advantages-grid">
          <div class="advantage-card">
            <div class="advantage-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h3>完整的一体化平台架构</h3>
            <p>前后端统一：Vue 3 + Vite 前端与 Go 后端在同一代码仓库中管理，单一后端入口，标准化 Kubernetes 部署</p>
          </div>
          <div class="advantage-card">
            <div class="advantage-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" stroke-width="2"/>
                <path d="M12 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" stroke="currentColor" stroke-width="2"/>
              </svg>
            </div>
            <h3>企业级功能完备性</h3>
            <p>实例、配置、备份、任务、告警、审计全流程覆盖，多渠道接入，Keycloak/OIDC 认证，OpenSearch 检索</p>
          </div>
          <div class="advantage-card">
            <div class="advantage-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h3>严格的生产可靠性</h3>
            <p>PLATFORM_STRICT_MODE 确保所有依赖就绪，基于 PostgreSQL 的唯一数据源，启动时自动校验，无降级回退</p>
          </div>
          <div class="advantage-card">
            <div class="advantage-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" stroke="currentColor" stroke-width="2"/>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96" stroke="currentColor" stroke-width="2"/>
                <line x1="12" y1="22.08" x2="12" y2="12" stroke="currentColor" stroke-width="2"/>
              </svg>
            </div>
            <h3>强大的集成与扩展能力</h3>
            <p>工作台桥接，多格式产物预览（web/html、pdf、office），对象存储集成，微信支付 API v3 完整支持</p>
          </div>
          <div class="advantage-card">
            <div class="advantage-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <path d="M22 12h-4l-3 9L9 3l-3 9H2" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h3>实时通信与事件处理</h3>
            <p>SSE 事件流支持实时事件传递和历史记录回放，异步回调平台侧消息桥接，标准化事件格式</p>
          </div>
          <div class="advantage-card">
            <div class="advantage-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <path d="M12 15v2" stroke="currentColor" stroke-width="2"/>
                <path d="M12 9v2" stroke="currentColor" stroke-width="2"/>
                <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
                <circle cx="12" cy="6" r="1" fill="currentColor"/>
                <circle cx="12" cy="18" r="1" fill="currentColor"/>
              </svg>
            </div>
            <h3>完善的运维与监控</h3>
            <p>健康检查 /healthz、/versionz 端点，详细的生产配置矩阵，完整的发布流程，k6 压测脚本</p>
          </div>
        </div>
      </section>

      <section class="contact-section">
        <h2>联系我们</h2>
        <div class="contact-links">
          <a href="mailto:contact@openclaw.dev" class="contact-link">邮件</a>
          <a href="#" class="contact-link">GitHub</a>
          <a href="#" class="contact-link">文档</a>
          <a href="#" class="contact-link">社区</a>
          <a href="#" class="contact-link">支持</a>
        </div>
      </section>
    </main>

    <footer class="footer">
      <div class="footer-bottom">
        <p>&copy; 2024 OpenClaw. 架构完整性 · 生产可靠性 · 功能完备性</p>
      </div>
    </footer>

    <MarketingMobileMenu :open="showMobileMenu" @close="showMobileMenu = false" />
    <ArchitectureDialog :open="showArchitecture" @close="showArchitecture = false" />
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #fef3c7 0%, #f8fafc 100%);
  color: #0f172a;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 2rem;
  background: rgba(255, 255, 255, 0.95);
  border-bottom: 2px solid #0f172a;
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
  font-weight: 800;
  font-size: 1.5rem;
  letter-spacing: -0.02em;
}

.logo-icon {
  color: #1d6bff;
  display: flex;
  align-items: center;
}

.header-spacer {
  flex: 1;
}

.header-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.btn-outline {
  background: transparent;
  border: 2px solid #0f172a;
  color: #0f172a;
  padding: 0.75rem 1.5rem;
  border-radius: 0;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 700;
  transition: all 0.2s;
}

.btn-outline:hover {
  background: #0f172a;
  color: white;
}

.btn-primary {
  background: #0f172a;
  border: 2px solid #0f172a;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 0;
  font-weight: 700;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: #1e293b;
  border-color: #1e293b;
  transform: translateY(-2px);
}

.main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.promo-section {
  padding: 4rem 0;
}

.promo-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
}

.promo-card {
  border: 3px solid #0f172a;
  padding: 3rem;
  background: white;
  box-shadow: 8px 8px 0 0 #0f172a;
  transition: all 0.2s;
}

.promo-card:hover {
  transform: translate(-4px, -4px);
  box-shadow: 12px 12px 0 0 #0f172a;
}

.promo-card.primary {
  background: linear-gradient(135deg, #fef3c7, #fde68a);
}

.promo-card h3 {
  font-size: 2rem;
  font-weight: 800;
  margin-bottom: 0.75rem;
  letter-spacing: -0.02em;
}

.promo-card p {
  font-size: 1.25rem;
  color: #475569;
  font-weight: 600;
}

.info-section {
  padding: 2rem 0 4rem;
}

.info-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

.info-card {
  border: 3px solid #0f172a;
  padding: 2rem;
  background: white;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: all 0.2s;
}

.info-card:hover {
  background: #0f172a;
  color: white;
}

.info-card h3 {
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.arrow {
  font-size: 2rem;
  font-weight: 800;
}

.compare-section {
  padding: 4rem 0;
  background: white;
  border: 3px solid #0f172a;
  margin: 2rem 0;
}

.compare-container {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 3rem;
  padding: 3rem;
}

.compare-left h2 {
  font-size: 2rem;
  font-weight: 800;
  margin-bottom: 2rem;
  letter-spacing: -0.02em;
}

.compare-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.compare-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 1.125rem;
  font-weight: 600;
}

.checkmark {
  width: 2rem;
  height: 2rem;
  background: #22c55e;
  color: white;
  display: grid;
  place-items: center;
  font-weight: 800;
  flex-shrink: 0;
}

.compare-right {
  border-left: 3px solid #0f172a;
  padding-left: 3rem;
}

.compare-right h3 {
  font-size: 1.5rem;
  font-weight: 800;
  margin-bottom: 1.5rem;
  letter-spacing: -0.02em;
}

.latest-features {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.feature-tag {
  display: inline-block;
  background: #0f172a;
  color: white;
  padding: 0.5rem 1rem;
  font-weight: 700;
  width: fit-content;
}

.advantages-section {
  padding: 5rem 0;
}

.section-head {
  text-align: center;
  margin-bottom: 4rem;
}

.section-head h2 {
  font-size: 3rem;
  font-weight: 800;
  margin-bottom: 1rem;
  letter-spacing: -0.02em;
}

.section-head p {
  font-size: 1.25rem;
  color: #475569;
  font-weight: 600;
}

.advantages-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 2rem;
}

.advantage-card {
  border: 3px solid #0f172a;
  background: white;
  padding: 2.5rem;
  transition: all 0.2s;
}

.advantage-card:hover {
  transform: translate(-4px, -4px);
  box-shadow: 8px 8px 0 0 #0f172a;
}

.advantage-icon {
  width: 4rem;
  height: 4rem;
  background: linear-gradient(135deg, #fef3c7, #fde68a);
  border: 2px solid #0f172a;
  display: grid;
  place-items: center;
  margin-bottom: 1.5rem;
  color: #0f172a;
}

.advantage-card h3 {
  font-size: 1.375rem;
  font-weight: 800;
  margin-bottom: 1rem;
  letter-spacing: -0.02em;
}

.advantage-card p {
  color: #475569;
  line-height: 1.7;
  font-weight: 500;
}

.contact-section {
  padding: 4rem 0;
  text-align: center;
  border-top: 3px solid #0f172a;
}

.contact-section h2 {
  font-size: 2.5rem;
  font-weight: 800;
  margin-bottom: 2rem;
  letter-spacing: -0.02em;
}

.contact-links {
  display: flex;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.contact-link {
  display: inline-block;
  border: 2px solid #0f172a;
  background: white;
  color: #0f172a;
  padding: 1rem 2rem;
  text-decoration: none;
  font-weight: 700;
  transition: all 0.2s;
}

.contact-link:hover {
  background: #0f172a;
  color: white;
}

.footer {
  background: white;
  border-top: 3px solid #0f172a;
  padding: 2rem;
  margin-top: 4rem;
}

.footer-bottom {
  text-align: center;
  color: #475569;
  font-weight: 600;
}

@media (max-width: 980px) {
  .header {
    padding: 1rem;
  }

  .promo-cards {
    grid-template-columns: 1fr;
    gap: 2rem;
  }

  .info-cards {
    grid-template-columns: 1fr;
  }

  .compare-container {
    grid-template-columns: 1fr;
  }

  .compare-right {
    border-left: none;
    border-top: 3px solid #0f172a;
    padding-left: 0;
    padding-top: 2rem;
  }

  .advantages-grid {
    grid-template-columns: 1fr;
  }

  .section-head h2 {
    font-size: 2rem;
  }
}
</style>
