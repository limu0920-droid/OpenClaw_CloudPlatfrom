<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import MarketingMobileMenu from './components/MarketingMobileMenu.vue'

const router = useRouter()
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
    <!-- 导航栏 -->
    <header class="header">
      <div class="container">
        <div class="header-content">
          <RouterLink class="logo" to="/">
            <div class="logo-icon">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none">
                <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <span>OpenClaw</span>
          </RouterLink>
          
          <!-- 桌面导航 -->
          <nav class="nav">
            <a href="#features" class="nav-link">产品特性</a>
            <a href="#advantages" class="nav-link">核心优势</a>
            <a href="#pricing" class="nav-link">价格</a>
            <a href="#docs" class="nav-link">文档</a>
            <a href="#case-studies" class="nav-link">客户案例</a>
          </nav>
          
          <div class="header-actions">
            <button class="btn-outline" type="button" @click="router.push('/login')">登录</button>
            <button class="btn-primary" type="button" @click="router.push('/login')">免费试用</button>
            <button class="mobile-menu-btn" type="button" @click="showMobileMenu = true">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <path d="M3 12H21" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                <path d="M3 6H21" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                <path d="M3 18H21" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </header>

    <main class="main">
      <!-- Hero 区域 -->
      <section class="hero-section">
        <div class="container">
          <div class="hero-content">
            <div class="hero-badge">企业级 AI 平台</div>
            <h1 class="hero-title">打造稳定可靠的<br>AI 应用管理平台</h1>
            <p class="hero-desc">
              OpenClaw 提供完整的一体化平台架构，帮助企业快速部署、管理和扩展 AI 应用，
              从实例管理到多渠道接入，从实时通信到企业级安全，为您的 AI 业务保驾护航。
            </p>
            <div class="hero-buttons">
              <button class="btn-large" type="button" @click="router.push('/login')">
                开始免费试用
              </button>
              <button class="btn-secondary" type="button" @click="scrollToSection('features')">
                了解更多
              </button>
            </div>
            <div class="hero-stats">
              <div class="stat-item">
                <span class="stat-number">500+</span>
                <span class="stat-label">企业客户</span>
              </div>
              <div class="stat-item">
                <span class="stat-number">99.99%</span>
                <span class="stat-label">可靠性</span>
              </div>
              <div class="stat-item">
                <span class="stat-number">24/7</span>
                <span class="stat-label">技术支持</span>
              </div>
            </div>
          </div>
          <div class="hero-visual">
            <div class="hero-image">
              <img src="https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=modern%20AI%20dashboard%20interface%20with%20instance%20management%2C%20metrics%20dashboard%2C%20and%20control%20panels&image_size=landscape_16_9" 
                   alt="OpenClaw 控制台界面" 
                   class="hero-img">
            </div>
          </div>
        </div>
      </section>

      <!-- 核心特性 -->
      <section id="features" class="features-section">
        <div class="container">
          <div class="section-head">
            <h2>核心特性</h2>
            <p>为企业级 AI 应用提供全方位的管理能力</p>
          </div>
          <div class="features-grid">
            <div class="feature-card">
              <div class="feature-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <h3>实例管理</h3>
              <p>轻松创建、配置、监控和管理 AI 实例，支持一键部署和自动扩缩容</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" stroke-width="2"/>
                  <path d="M12 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" stroke="currentColor" stroke-width="2"/>
                </svg>
              </div>
              <h3>多渠道接入</h3>
              <p>支持飞书、企微、钉钉、Slack 等多种渠道统一管理，一键集成</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" stroke-width="2"/>
                  <path d="M9 12l2 2 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <h3>实时通信</h3>
              <p>SSE 事件流支持实时消息传递，确保系统响应及时可靠</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M20 7h-4V4c0-1.1-.9-2-2-2h-4c-1.1 0-2 .9-2 2v3H4c-1.1 0-2 .9-2 2v11c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V9c0-1.1-.9-2-2-2z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <polyline points="16 17 12 13 8 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <h3>企业级安全</h3>
              <p>Keycloak/OIDC 认证，严格的权限管理，确保系统安全可靠</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" stroke="currentColor" stroke-width="2"/>
                  <polyline points="3.27 6.96 12 12.01 20.73 6.96" stroke="currentColor" stroke-width="2"/>
                  <line x1="12" y1="22.08" x2="12" y2="12" stroke="currentColor" stroke-width="2"/>
                </svg>
              </div>
              <h3>可扩展性</h3>
              <p>灵活的插件架构，支持自定义集成和功能扩展</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z" stroke="currentColor" stroke-width="2"/>
                  <path d="M12 8v4l3 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <h3>监控告警</h3>
              <p>实时监控系统状态，智能告警，确保服务稳定运行</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 核心优势 -->
      <section id="advantages" class="advantages-section">
        <div class="container">
          <div class="section-head">
            <h2>核心优势</h2>
            <p>为什么选择 OpenClaw</p>
          </div>
          <div class="advantages-grid">
            <div class="advantage-card">
              <div class="advantage-icon">
                <svg width="56" height="56" viewBox="0 0 24 24" fill="none">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" stroke-width="2"/>
                </svg>
              </div>
              <h3>一体化平台架构</h3>
              <p>前后端统一管理，单一后端入口，标准化部署流程，减少集成成本，提高开发效率</p>
            </div>
            <div class="advantage-card">
              <div class="advantage-icon">
                <svg width="56" height="56" viewBox="0 0 24 24" fill="none">
                  <path d="M20 7h-4V4c0-1.1-.9-2-2-2h-4c-1.1 0-2 .9-2 2v3H4c-1.1 0-2 .9-2 2v11c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V9c0-1.1-.9-2-2-2z" stroke="currentColor" stroke-width="2"/>
                </svg>
              </div>
              <h3>企业级功能完备</h3>
              <p>覆盖实例管理、配置、备份、任务、告警、审计全流程，满足企业级应用的各种需求</p>
            </div>
            <div class="advantage-card">
              <div class="advantage-icon">
                <svg width="56" height="56" viewBox="0 0 24 24" fill="none">
                  <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" stroke="currentColor" stroke-width="2"/>
                </svg>
              </div>
              <h3>严格的可靠性</h3>
              <p>严格运行模式确保依赖服务就绪，基于 PostgreSQL 的持久化存储，启动时自动验证，保证生产稳定性</p>
            </div>
            <div class="advantage-card">
              <div class="advantage-icon">
                <svg width="56" height="56" viewBox="0 0 24 24" fill="none">
                  <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z" stroke="currentColor" stroke-width="2"/>
                  <path d="M9 12l2 2 4-4" stroke="currentColor" stroke-width="2"/>
                </svg>
              </div>
              <h3>强大的集成能力</h3>
              <p>工作台桥接、多格式产物预览、对象存储集成、微信支付支持，满足复杂业务场景需求</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 价格方案 -->
      <section id="pricing" class="pricing-section">
        <div class="container">
          <div class="section-head">
            <h2>价格方案</h2>
            <p>选择适合您业务的方案</p>
          </div>
          <div class="pricing-grid">
            <div class="pricing-card">
              <div class="pricing-header">
                <h3>免费版</h3>
                <div class="price">
                  <span class="price-value">¥0</span>
                  <span class="price-unit">/月</span>
                </div>
              </div>
              <div class="pricing-features">
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>最多 3 个实例</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>基础监控</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>邮件支持</span>
                </div>
                <div class="feature-item disabled">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                    <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  </svg>
                  <span>高级集成</span>
                </div>
              </div>
              <button class="btn-outline" type="button" @click="router.push('/login')">
                开始使用
              </button>
            </div>
            <div class="pricing-card featured">
              <div class="featured-badge">推荐</div>
              <div class="pricing-header">
                <h3>企业版</h3>
                <div class="price">
                  <span class="price-value">¥1,999</span>
                  <span class="price-unit">/月</span>
                </div>
              </div>
              <div class="pricing-features">
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>无限实例</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>高级监控告警</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>24/7 技术支持</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>全功能集成</span>
                </div>
              </div>
              <button class="btn-primary" type="button" @click="router.push('/login')">
                立即购买
              </button>
            </div>
            <div class="pricing-card">
              <div class="pricing-header">
                <h3>定制版</h3>
                <div class="price">
                  <span class="price-value">联系销售</span>
                </div>
              </div>
              <div class="pricing-features">
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>完全定制</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>专属部署</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>专属客服</span>
                </div>
                <div class="feature-item">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <span>定制开发</span>
                </div>
              </div>
              <button class="btn-outline" type="button" @click="scrollToSection('contact')">
                联系我们
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- 客户案例 -->
      <section id="case-studies" class="case-studies-section">
        <div class="container">
          <div class="section-head">
            <h2>客户案例</h2>
            <p>看看其他企业如何使用 OpenClaw</p>
          </div>
          <div class="case-studies-grid">
            <div class="case-study-card">
              <div class="case-study-logo">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none">
                  <rect x="2" y="2" width="20" height="20" rx="4" stroke="currentColor" stroke-width="2"/>
                  <path d="M8 12h8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <path d="M8 16h4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <h3>某大型金融机构</h3>
              <p>通过 OpenClaw 平台，该金融机构实现了 AI 应用的快速部署和管理，提升了业务处理效率，降低了运维成本</p>
            </div>
            <div class="case-study-card">
              <div class="case-study-logo">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none">
                  <rect x="2" y="2" width="20" height="20" rx="4" stroke="currentColor" stroke-width="2"/>
                  <path d="M8 12h8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <path d="M8 16h4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <h3>某电商平台</h3>
              <p>利用 OpenClaw 的多渠道接入能力，该电商平台实现了客服系统的智能化升级，提升了客户满意度</p>
            </div>
            <div class="case-study-card">
              <div class="case-study-logo">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none">
                  <rect x="2" y="2" width="20" height="20" rx="4" stroke="currentColor" stroke-width="2"/>
                  <path d="M8 12h8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <path d="M8 16h4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <h3>某制造业企业</h3>
              <p>通过 OpenClaw 平台，该制造业企业实现了生产流程的智能化管理，提高了生产效率和产品质量</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 文档与资源 -->
      <section id="docs" class="docs-section">
        <div class="container">
          <div class="section-head">
            <h2>文档与资源</h2>
            <p>帮助您快速上手 OpenClaw</p>
          </div>
          <div class="docs-grid">
            <div class="doc-card">
              <div class="doc-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <polyline points="14 2 14 8 20 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <line x1="16" y1="13" x2="8" y2="13" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <line x1="16" y1="17" x2="8" y2="17" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <polyline points="10 9 9 9 8 9" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <h3>快速开始</h3>
              <p>帮助您快速部署和配置 OpenClaw 平台</p>
              <a href="#" class="doc-link">查看文档 →</a>
            </div>
            <div class="doc-card">
              <div class="doc-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
                  <path d="M12 16v-4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <path d="M12 8h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <h3>API 参考</h3>
              <p>详细的 API 文档和使用示例</p>
              <a href="#" class="doc-link">查看文档 →</a>
            </div>
            <div class="doc-card">
              <div class="doc-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                  <path d="M18 8h1a4 4 0 0 1 0 8h-1" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M2 8h16v9a4 4 0 0 1-4 4H6a4 4 0 0 1-4-4V8z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <line x1="6" y1="1" x2="6" y2="4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <line x1="10" y1="1" x2="10" y2="4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                  <line x1="14" y1="1" x2="14" y2="4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <h3>最佳实践</h3>
              <p>企业级部署和管理的最佳实践指南</p>
              <a href="#" class="doc-link">查看文档 →</a>
            </div>
          </div>
        </div>
      </section>

      <!-- 联系我们 -->
      <section id="contact" class="contact-section">
        <div class="container">
          <div class="contact-content">
            <div class="contact-info">
              <h2>联系我们</h2>
              <p>有任何问题或需求，欢迎随时联系我们</p>
              <div class="contact-details">
                <div class="contact-item">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" stroke="currentColor" stroke-width="2"/>
                    <polyline points="22,6 12,13 2,6" stroke="currentColor" stroke-width="2"/>
                  </svg>
                  <span>contact@openclaw.dev</span>
                </div>
                <div class="contact-item">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                    <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z" stroke="currentColor" stroke-width="2"/>
                  </svg>
                  <span>400-123-4567</span>
                </div>
              </div>
            </div>
            <div class="contact-form">
              <h3>发送消息</h3>
              <form>
                <div class="form-group">
                  <input type="text" placeholder="您的姓名" required>
                </div>
                <div class="form-group">
                  <input type="email" placeholder="您的邮箱" required>
                </div>
                <div class="form-group">
                  <textarea placeholder="您的消息" rows="4" required></textarea>
                </div>
                <button type="submit" class="btn-primary">发送消息</button>
              </form>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- 页脚 -->
    <footer class="footer">
      <div class="container">
        <div class="footer-content">
          <div class="footer-brand">
            <div class="logo">
              <div class="logo-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                  <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <span>OpenClaw</span>
            </div>
            <p>企业级 AI 平台管理解决方案</p>
          </div>
          <div class="footer-links">
            <div class="footer-col">
              <h4>产品</h4>
              <a href="#features">特性</a>
              <a href="#advantages">优势</a>
              <a href="#pricing">价格</a>
              <a href="#case-studies">案例</a>
            </div>
            <div class="footer-col">
              <h4>资源</h4>
              <a href="#docs">文档</a>
              <a href="#">API 参考</a>
              <a href="#">博客</a>
              <a href="#">社区</a>
            </div>
            <div class="footer-col">
              <h4>公司</h4>
              <a href="#">关于我们</a>
              <a href="#contact">联系我们</a>
              <a href="#">隐私政策</a>
              <a href="#">服务条款</a>
            </div>
          </div>
        </div>
        <div class="footer-bottom">
          <p>&copy; 2026 OpenClaw. 保留所有权利。</p>
        </div>
      </div>
    </footer>

    <MarketingMobileMenu :open="showMobileMenu" @close="showMobileMenu = false" />
  </div>
</template>

<style scoped>
/* 全局样式 */
.home-page {
  min-height: 100vh;
  background: #f8fafc;
  color: #0f172a;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* 容器 */
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1.5rem;
}

/* 导航栏 */
.header {
  background: white;
  border-bottom: 1px solid #e2e8f0;
  position: sticky;
  top: 0;
  z-index: 50;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 0;
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

.nav {
  display: flex;
  gap: 2rem;
  align-items: center;
}

.nav-link {
  color: #64748b;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.nav-link:hover {
  color: #1d6bff;
}

.header-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.mobile-menu-btn {
  display: none;
  background: none;
  border: none;
  color: #0f172a;
  cursor: pointer;
  padding: 0.5rem;
}

/* 按钮样式 */
.btn-outline {
  background: transparent;
  border: 2px solid #1d6bff;
  color: #1d6bff;
  padding: 0.75rem 1.5rem;
  border-radius: 0.75rem;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-outline:hover {
  background: #1d6bff;
  color: white;
  transform: translateY(-1px);
}

.btn-primary {
  background: linear-gradient(135deg, #1d6bff, #3b82f6);
  border: none;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 6px -1px rgba(29, 107, 255, 0.3);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px -3px rgba(29, 107, 255, 0.4);
}

.btn-large {
  background: linear-gradient(135deg, #1d6bff, #3b82f6);
  border: none;
  color: white;
  padding: 1rem 2rem;
  border-radius: 0.875rem;
  font-weight: 600;
  font-size: 1.125rem;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 6px -1px rgba(29, 107, 255, 0.3);
}

.btn-large:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px -3px rgba(29, 107, 255, 0.4);
}

.btn-secondary {
  background: white;
  border: 2px solid #e2e8f0;
  color: #475569;
  padding: 1rem 2rem;
  border-radius: 0.875rem;
  font-weight: 600;
  font-size: 1.125rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  border-color: #cbd5e1;
  transform: translateY(-1px);
}

/* Hero 区域 */
.hero-section {
  padding: 6rem 0 8rem;
  background: linear-gradient(135deg, #f8fafc, #e2e8f0);
}

.hero-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4rem;
  align-items: center;
}

.hero-badge {
  display: inline-block;
  background: #eff6ff;
  color: #1d4ed8;
  padding: 0.5rem 1rem;
  border-radius: 999px;
  font-weight: 600;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
}

.hero-title {
  font-size: 3.5rem;
  font-weight: 800;
  line-height: 1.1;
  margin-bottom: 1.5rem;
  letter-spacing: -0.02em;
  background: linear-gradient(135deg, #0f172a, #1d6bff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-desc {
  font-size: 1.25rem;
  color: #475569;
  line-height: 1.6;
  margin-bottom: 2rem;
  max-width: 500px;
}

.hero-buttons {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 2.5rem;
}

.hero-stats {
  display: flex;
  gap: 3rem;
}

.stat-item {
  text-align: center;
}

.stat-number {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: #1d6bff;
  margin-bottom: 0.25rem;
}

.stat-label {
  display: block;
  font-size: 0.875rem;
  color: #64748b;
}

.hero-image {
  border-radius: 1.5rem;
  overflow: hidden;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.hero-img {
  width: 100%;
  height: auto;
  border-radius: 1.5rem;
}

/* 通用部分样式 */
.section-head {
  text-align: center;
  margin-bottom: 4rem;
}

.section-head h2 {
  font-size: 2.5rem;
  font-weight: 800;
  margin-bottom: 1rem;
  letter-spacing: -0.02em;
  color: #0f172a;
}

.section-head p {
  font-size: 1.125rem;
  color: #64748b;
  max-width: 600px;
  margin: 0 auto;
}

/* 核心特性 */
.features-section {
  padding: 8rem 0;
  background: white;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

.feature-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2.5rem;
  text-align: center;
  transition: all 0.2s;
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  border-color: #cbd5e1;
}

.feature-icon {
  width: 4rem;
  height: 4rem;
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
  color: #1d6bff;
  display: grid;
  place-items: center;
  margin: 0 auto 1.5rem;
  border-radius: 1rem;
}

.feature-card h3 {
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #0f172a;
}

.feature-card p {
  color: #64748b;
  line-height: 1.6;
}

/* 核心优势 */
.advantages-section {
  padding: 8rem 0;
  background: #f8fafc;
}

.advantages-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 2.5rem;
}

.advantage-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 3rem;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.advantage-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  border-color: #cbd5e1;
}

.advantage-icon {
  width: 4.5rem;
  height: 4.5rem;
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
  color: #1d6bff;
  display: grid;
  place-items: center;
  margin-bottom: 1.5rem;
  border-radius: 1.25rem;
  flex-shrink: 0;
}

.advantage-card h3 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #0f172a;
}

.advantage-card p {
  color: #64748b;
  line-height: 1.6;
  flex: 1;
}

/* 价格方案 */
.pricing-section {
  padding: 8rem 0;
  background: white;
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

.pricing-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2.5rem;
  position: relative;
  transition: all 0.2s;
}

.pricing-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  border-color: #cbd5e1;
}

.pricing-card.featured {
  background: white;
  border: 2px solid #1d6bff;
  transform: scale(1.05);
}

.featured-badge {
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  background: #1d6bff;
  color: white;
  padding: 0.5rem 1.5rem;
  border-radius: 999px;
  font-weight: 600;
  font-size: 0.875rem;
}

.pricing-header {
  text-align: center;
  margin-bottom: 2rem;
}

.pricing-header h3 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #0f172a;
}

.price {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.price-value {
  font-size: 2.5rem;
  font-weight: 800;
  color: #1d6bff;
}

.price-unit {
  font-size: 1rem;
  color: #64748b;
}

.pricing-features {
  margin-bottom: 2rem;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
  color: #475569;
}

.feature-item.disabled {
  color: #94a3b8;
}

.pricing-card button {
  width: 100%;
  margin-top: 1rem;
}

/* 客户案例 */
.case-studies-section {
  padding: 8rem 0;
  background: #f8fafc;
}

.case-studies-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

.case-study-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2.5rem;
  text-align: center;
  transition: all 0.2s;
}

.case-study-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  border-color: #cbd5e1;
}

.case-study-logo {
  width: 5rem;
  height: 5rem;
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
  color: #1d6bff;
  display: grid;
  place-items: center;
  margin: 0 auto 1.5rem;
  border-radius: 1.25rem;
}

.case-study-card h3 {
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #0f172a;
}

.case-study-card p {
  color: #64748b;
  line-height: 1.6;
}

/* 文档与资源 */
.docs-section {
  padding: 8rem 0;
  background: white;
}

.docs-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

.doc-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2.5rem;
  transition: all 0.2s;
}

.doc-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  border-color: #cbd5e1;
}

.doc-icon {
  width: 4rem;
  height: 4rem;
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
  color: #1d6bff;
  display: grid;
  place-items: center;
  margin-bottom: 1.5rem;
  border-radius: 1rem;
}

.doc-card h3 {
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #0f172a;
}

.doc-card p {
  color: #64748b;
  line-height: 1.6;
  margin-bottom: 1.5rem;
}

.doc-link {
  color: #1d6bff;
  text-decoration: none;
  font-weight: 600;
  transition: color 0.2s;
  display: inline-block;
}

.doc-link:hover {
  color: #1e40af;
}

/* 联系我们 */
.contact-section {
  padding: 8rem 0;
  background: #f8fafc;
}

.contact-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4rem;
  align-items: start;
}

.contact-info h2 {
  font-size: 2.5rem;
  font-weight: 800;
  margin-bottom: 1rem;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.contact-info p {
  font-size: 1.125rem;
  color: #64748b;
  margin-bottom: 2rem;
}

.contact-details {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: #475569;
}

.contact-form {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2.5rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
}

.contact-form h3 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 1.5rem;
  color: #0f172a;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #1d6bff;
  box-shadow: 0 0 0 3px rgba(29, 107, 255, 0.1);
}

.form-group textarea {
  resize: vertical;
  min-height: 120px;
}

/* 页脚 */
.footer {
  background: white;
  border-top: 1px solid #e2e8f0;
  padding: 4rem 0 2rem;
}

.footer-content {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 3rem;
  margin-bottom: 2rem;
}

.footer-brand .logo {
  margin-bottom: 1rem;
}

.footer-brand p {
  color: #64748b;
  line-height: 1.6;
}

.footer-col h4 {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 1.5rem;
  color: #0f172a;
}

.footer-col a {
  display: block;
  color: #64748b;
  text-decoration: none;
  margin-bottom: 0.75rem;
  transition: color 0.2s;
}

.footer-col a:hover {
  color: #1d6bff;
}

.footer-bottom {
  border-top: 1px solid #e2e8f0;
  padding-top: 2rem;
  text-align: center;
  color: #64748b;
  font-size: 0.875rem;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .hero-content {
    grid-template-columns: 1fr;
    text-align: center;
  }
  
  .hero-stats {
    justify-content: center;
  }
  
  .features-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .advantages-grid {
    grid-template-columns: 1fr;
  }
  
  .pricing-grid {
    grid-template-columns: 1fr;
  }
  
  .pricing-card.featured {
    transform: none;
  }
  
  .case-studies-grid {
    grid-template-columns: 1fr;
  }
  
  .docs-grid {
    grid-template-columns: 1fr;
  }
  
  .contact-content {
    grid-template-columns: 1fr;
  }
  
  .footer-content {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
  
  .nav {
    display: none;
  }
  
  .mobile-menu-btn {
    display: block;
  }
}

@media (max-width: 768px) {
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-buttons {
    flex-direction: column;
    align-items: center;
  }
  
  .features-grid {
    grid-template-columns: 1fr;
  }
  
  .section-head h2 {
    font-size: 2rem;
  }
  
  .container {
    padding: 0 1rem;
  }
  
  .header-content {
    padding: 1rem 0;
  }
  
  .hero-section,
  .features-section,
  .advantages-section,
  .pricing-section,
  .case-studies-section,
  .docs-section,
  .contact-section {
    padding: 4rem 0;
  }
}
</style>
