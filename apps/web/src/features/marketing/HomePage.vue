<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import ArchitectureDialog from './components/ArchitectureDialog.vue'
import MarketingMobileMenu from './components/MarketingMobileMenu.vue'
import QuickEntryPopover from './components/QuickEntryPopover.vue'

const router = useRouter()
const showArchitecture = ref(false)
const showMobileMenu = ref(false)

const features = [
  {
    title: '无限 token',
    subtitle: '免费试用',
    description: '畅享无限 token，免费试用体验所有功能'
  },
  {
    title: '按收入一键部署',
    subtitle: '快速启动',
    description: '根据收入情况，一键部署适合的方案'
  },
  {
    title: '使用友好',
    subtitle: '易于操作',
    description: '直观的用户界面，让操作变得简单高效'
  },
  {
    title: '优势',
    subtitle: '核心竞争力',
    description: '提供行业领先的解决方案和服务'
  }
]

const comparisonItems = [
  {
    title: '对比',
    subtitle: '流程',
    description: '详细对比不同方案的实施流程和效果'
  },
  {
    title: '详细使用',
    subtitle: '指南',
    description: '提供完整的使用指南和最佳实践'
  }
]

const contactInfo = [
  {
    title: '联系我们',
    items: [
      '邮箱：contact@openclaw.dev',
      '电话：400-123-4567',
      '地址：北京市朝阳区',
      '微信：OpenClawOfficial',
      'QQ：123456789'
    ]
  }
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
        <RouterLink to="/" class="nav-link">首页</RouterLink>
        <a href="#features" class="nav-link">功能</a>
        <a href="#comparison" class="nav-link">对比</a>
        <a href="#contact" class="nav-link">联系我们</a>
      </nav>
      <div class="site-actions">
        <button class="nav-link-btn" type="button" @click="router.push('/login')">登录</button>
        <button class="nav-primary-btn" type="button" @click="router.push('/register')">注册</button>
      </div>
    </header>

    <main class="home-main">
      <section class="hero">
        <div class="hero-content">
          <h1 class="hero-title">欢迎使用 openClaw</h1>
          <p class="hero-description">
            提供强大的 AI 服务和解决方案，助力企业数字化转型
          </p>
          <div class="hero-actions">
            <button class="primary-btn" type="button" @click="router.push('/register')">立即开始</button>
            <button class="secondary-btn" type="button" @click="showArchitecture = true">了解更多</button>
          </div>
        </div>
      </section>

      <section id="features" class="features-section">
        <div class="features-grid">
          <div v-for="(feature, index) in features" :key="index" class="feature-card">
            <h3 class="feature-title">{{ feature.title }}</h3>
            <div class="feature-subtitle">{{ feature.subtitle }}</div>
            <p class="feature-description">{{ feature.description }}</p>
          </div>
        </div>
      </section>

      <section id="comparison" class="comparison-section">
        <div class="comparison-grid">
          <div v-for="(item, index) in comparisonItems" :key="index" class="comparison-card">
            <h3 class="comparison-title">{{ item.title }}</h3>
            <div class="comparison-subtitle">{{ item.subtitle }}</div>
            <p class="comparison-description">{{ item.description }}</p>
          </div>
        </div>
      </section>

      <section id="contact" class="contact-section">
        <div v-for="(contact, index) in contactInfo" :key="index" class="contact-card">
          <h3 class="contact-title">{{ contact.title }}</h3>
          <div class="contact-items">
            <div v-for="(item, itemIndex) in contact.items" :key="itemIndex" class="contact-item">
              {{ item }}
            </div>
          </div>
        </div>
      </section>

      <footer class="site-footer">
        <div class="footer-content">
          <p class="copyright">&copy; 2024 openClaw. 保留所有权利。</p>
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
  padding: 5rem 2rem;
  background: linear-gradient(135deg, #f8fafc, #eef2ff);
  text-align: center;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.hero-title {
  font-size: 3rem;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.02em;
  background: linear-gradient(120deg, #0f172a, #1d6bff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-description {
  font-size: 1.125rem;
  color: #475569;
  line-height: 1.7;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  margin-top: 1rem;
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

.features-section {
  padding: 5rem 2rem;
  background: white;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.feature-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2rem;
  transition: all 0.2s;
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px -12px rgba(15, 23, 42, 0.15);
  border-color: #cbd5e1;
}

.feature-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  color: #0f172a;
}

.feature-subtitle {
  display: inline-block;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1d4ed8;
  padding: 0.375rem 0.875rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.feature-description {
  color: #64748b;
  line-height: 1.7;
  font-size: 0.9375rem;
}

.comparison-section {
  padding: 5rem 2rem;
  background: #f8fafc;
}

.comparison-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.comparison-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2.5rem;
  transition: all 0.2s;
}

.comparison-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px -12px rgba(15, 23, 42, 0.15);
  border-color: #cbd5e1;
}

.comparison-title {
  font-size: 1.75rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  color: #0f172a;
}

.comparison-subtitle {
  display: inline-block;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  padding: 0.375rem 0.875rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
}

.comparison-description {
  color: #64748b;
  line-height: 1.7;
  font-size: 1rem;
}

.contact-section {
  padding: 5rem 2rem;
  background: white;
}

.contact-card {
  max-width: 800px;
  margin: 0 auto;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  padding: 2.5rem;
  text-align: center;
}

.contact-title {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 2rem;
  color: #0f172a;
}

.contact-items {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.contact-item {
  color: #64748b;
  font-size: 0.9375rem;
  line-height: 1.6;
}

.site-footer {
  background: #0f172a;
  color: white;
  padding: 2rem;
  text-align: center;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
}

.copyright {
  font-size: 0.9375rem;
  color: #94a3b8;
}

@media (max-width: 980px) {
  .site-header {
    padding: 1rem;
  }
  
  .site-nav {
    display: none;
  }
  
  .hero {
    padding: 3rem 1rem;
  }
  
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-actions {
    flex-direction: column;
    align-items: center;
  }
  
  .features-section,
  .comparison-section,
  .contact-section {
    padding: 3rem 1rem;
  }
  
  .features-grid,
  .comparison-grid {
    grid-template-columns: 1fr;
  }
  
  .contact-items {
    grid-template-columns: 1fr;
  }
}
</style>
