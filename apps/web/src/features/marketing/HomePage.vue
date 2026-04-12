<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import ArchitectureDialog from './components/ArchitectureDialog.vue'
import ControlPlaneTabs from './components/ControlPlaneTabs.vue'
import MarketingFaq from './components/MarketingFaq.vue'
import MarketingMobileMenu from './components/MarketingMobileMenu.vue'
import QuickEntryPopover from './components/QuickEntryPopover.vue'
import { useBranding } from '../../lib/brand'

const router = useRouter()
const showArchitecture = ref(false)
const showMobileMenu = ref(false)
const { brand } = useBranding()

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

function scrollTo(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
</script>

<template>
  <div class="home-shell">
    <header class="site-header">
      <RouterLink class="brand" to="/">
        <div class="brand-mark">
          <img v-if="brand.logoUrl" :src="brand.logoUrl" :alt="brand.name" class="brand-logo" />
          <span v-else>{{ brand.name.slice(0, 2).toUpperCase() }}</span>
        </div>
        <div>
          <div class="brand-title">{{ brand.name }}</div>
          <div class="brand-sub">Platform</div>
        </div>
      </RouterLink>
      <nav class="site-nav">
        <button type="button" @click="scrollTo('values')">价值主张</button>
        <button type="button" @click="scrollTo('planes')">控制台</button>
        <button type="button" @click="scrollTo('delivery')">上线节奏</button>
      </nav>
      <div class="site-actions">
        <QuickEntryPopover />
        <button class="nav-cta" type="button" @click="router.push('/login')">登录入口</button>
        <button class="mobile-nav-button" type="button" @click="showMobileMenu = true">菜单</button>
      </div>
    </header>

    <main class="home-main">
      <section class="hero">
        <div class="hero-copy">
          <div class="eyebrow">{{ brand.name }} 官网</div>
          <h1>把 {{ brand.name }} 的运行能力，整理成一套更适合企业落地的平台入口。</h1>
          <p class="hero-text">
            这次不只是把首页做漂亮，而是把官网、登录和双控制台的关系讲清楚。Headless UI 负责更定制的品牌交互，业务后台继续保持高效率组件化。
          </p>
          <div class="hero-actions">
            <button class="primary" type="button" @click="router.push('/login')">开始使用</button>
            <button class="ghost" type="button" @click="showArchitecture = true">查看架构预览</button>
            <button class="ghost ghost-soft" type="button" @click="scrollTo('planes')">查看控制台</button>
          </div>
          <div class="hero-footnote">
            <span class="pill">Portal / Admin 双壳</span>
            <span class="pill">Headless UI 交互层</span>
            <span class="pill">Element Plus 业务层</span>
          </div>
        </div>

        <div class="hero-stage card">
          <div class="stage-top">
            <div>
              <div class="stage-label">交付重点</div>
              <strong>实例控制台先行</strong>
            </div>
            <div class="stage-chip">Phase 1</div>
          </div>

          <div class="stage-grid">
            <article class="stage-card">
              <span>实例</span>
              <strong>生命周期</strong>
              <p>创建、配置、备份、恢复与任务跟踪。</p>
            </article>
            <article class="stage-card">
              <span>渠道</span>
              <strong>统一接入</strong>
              <p>飞书、企微、钉钉、Slack 等入口集中配置。</p>
            </article>
            <article class="stage-card">
              <span>品牌</span>
              <strong>更定制的交互层</strong>
              <p>用 Headless UI 做入口、弹层、切换面板，不再让官网停留在通用后台感。</p>
            </article>
            <article class="stage-card accent">
              <span>体验</span>
              <strong>原生能力不丢</strong>
              <p>平台壳层表达业务闭环，原生控制面继续承接调试与运维细节。</p>
            </article>
          </div>
        </div>
      </section>

      <section id="values" class="section">
        <div class="section-head">
          <div class="eyebrow">价值主张</div>
          <h2>官网先传达三件事：能快、够稳、可控。</h2>
        </div>
        <div class="value-grid">
          <article v-for="item in valueCards" :key="item.title" class="value-card card">
            <h3>{{ item.title }}</h3>
            <p>{{ item.text }}</p>
          </article>
        </div>
      </section>

      <section id="planes" class="section split">
        <div class="section-head compact">
          <div class="eyebrow">双控制台</div>
          <h2>租户入口和平台运维入口分开，并且用 Tab 交互把两套壳层的职责清楚展示出来。</h2>
        </div>
        <ControlPlaneTabs />
      </section>

      <section id="delivery" class="section delivery card">
        <div class="section-head compact">
          <div class="eyebrow">上线节奏</div>
          <h2>第一阶段不做大而全，先把实例控制与接入主链路跑通。</h2>
        </div>
        <div class="delivery-grid">
          <div class="delivery-column">
            <div class="delivery-kicker">交付路线</div>
            <ol class="step-list">
              <li v-for="step in steps" :key="step">{{ step }}</li>
            </ol>
          </div>
          <div class="delivery-column matrix">
            <div class="matrix-card">
              <span>可信感</span>
              <strong>亮色官网 + 登录入口</strong>
            </div>
            <div class="matrix-card">
              <span>低学习成本</span>
              <strong>Portal 保持轻量</strong>
            </div>
            <div class="matrix-card">
              <span>控制效率</span>
              <strong>Admin 深色控制面</strong>
            </div>
            <div class="matrix-card">
              <span>后续扩展</span>
              <strong>保留真实 IAM / Search / 渠道接线位</strong>
            </div>
          </div>
        </div>
      </section>

      <section class="section faq-section">
        <div class="section-head compact">
          <div class="eyebrow">设计说明</div>
          <h2>把 Headless UI 放在官网和品牌交互位，保留后台效率组件，这条边界更合理。</h2>
        </div>
        <MarketingFaq />
      </section>
    </main>

    <ArchitectureDialog :open="showArchitecture" @close="showArchitecture = false" />
    <MarketingMobileMenu :open="showMobileMenu" @close="showMobileMenu = false" />
  </div>
</template>

<style scoped>
.home-shell {
  min-height: 100vh;
  padding: 24px;
}

.site-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 18px 22px;
  margin: 0 auto 20px;
  max-width: 1280px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px);
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.08);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}

.brand-mark {
  width: 46px;
  height: 46px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 700;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.2), transparent),
    linear-gradient(135deg, #0f6bff, #5c7bff 65%, #1ad1ff);
  box-shadow: 0 14px 30px rgba(29, 107, 255, 0.32);
}

.brand-logo {
  width: 28px;
  height: 28px;
  object-fit: contain;
}

.brand-title {
  font-size: 1rem;
  font-variation-settings: "wght" 650;
}

.brand-sub {
  font-size: 12px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.18em;
}

.site-nav,
.site-actions {
  display: flex;
  align-items: center;
  gap: 14px;
}

.site-nav button,
.nav-cta {
  border: none;
  background: transparent;
  color: var(--text-muted);
  font: inherit;
}

.site-nav button:hover,
.nav-cta:hover {
  color: var(--text);
}

.nav-cta {
  padding: 10px 16px;
  border-radius: 999px;
  border: 1px solid rgba(29, 107, 255, 0.14);
  background: rgba(29, 107, 255, 0.08);
  color: var(--brand);
  font-variation-settings: "wght" 600;
}

.mobile-nav-button {
  display: none;
  padding: 10px 16px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.76);
  color: var(--text);
}

.home-main {
  display: flex;
  flex-direction: column;
  gap: 22px;
  max-width: 1280px;
  margin: 0 auto;
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(420px, 0.9fr);
  gap: 18px;
  align-items: stretch;
}

.hero-copy,
.hero-stage {
  position: relative;
  overflow: hidden;
  min-height: 560px;
  border-radius: 32px;
}

.hero-copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 44px;
  background:
    radial-gradient(circle at 20% 18%, rgba(29, 107, 255, 0.22), transparent 30%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.96), rgba(236, 243, 255, 0.92));
  border: 1px solid rgba(148, 163, 184, 0.18);
  box-shadow: 0 24px 64px rgba(15, 23, 42, 0.08);
}

.hero-copy::after {
  content: "";
  position: absolute;
  right: -10%;
  bottom: -18%;
  width: 240px;
  height: 240px;
  border-radius: 40px;
  background: linear-gradient(135deg, rgba(29, 107, 255, 0.24), rgba(34, 197, 94, 0.18));
  transform: rotate(24deg);
  filter: blur(6px);
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(29, 107, 255, 0.08);
  color: var(--brand);
  font-size: 12px;
  font-variation-settings: "wght" 620;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.hero-copy h1 {
  margin: 18px 0 16px;
  max-width: 10ch;
  font-size: clamp(3rem, 7vw, 5.2rem);
  line-height: 0.98;
  font-variation-settings: "wght" 680;
}

.hero-text {
  max-width: 640px;
  margin: 0;
  font-size: 1.08rem;
  line-height: 1.75;
  color: var(--text-muted);
}

.hero-actions,
.hero-footnote {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero-actions {
  margin-top: 26px;
}

.hero-footnote {
  margin-top: 18px;
}

.primary,
.ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 148px;
  padding: 13px 18px;
  border-radius: 999px;
  border: 1px solid transparent;
  font: inherit;
}

.primary {
  color: #fff;
  background: linear-gradient(120deg, #0f6bff, #5c7bff 65%, #1ad1ff);
  box-shadow: 0 16px 34px rgba(29, 107, 255, 0.28);
}

.ghost {
  border-color: rgba(148, 163, 184, 0.24);
  background: rgba(255, 255, 255, 0.72);
  color: var(--text);
}

.ghost-soft {
  background: rgba(244, 247, 255, 0.96);
}

.pill {
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.hero-stage {
  padding: 24px;
  background:
    radial-gradient(circle at top right, rgba(34, 197, 94, 0.16), transparent 18%),
    linear-gradient(165deg, #0d1728, #12213a 55%, #102847);
  border: 1px solid rgba(96, 165, 250, 0.18);
  color: #e5edf9;
}

.hero-stage::before {
  content: "";
  position: absolute;
  inset: 14px;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.12);
  pointer-events: none;
}

.stage-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.stage-label {
  color: rgba(226, 232, 240, 0.72);
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.stage-chip {
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(59, 130, 246, 0.16);
  color: #c7ddff;
  font-size: 12px;
}

.stage-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 28px;
}

.stage-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 18px;
  min-height: 150px;
  border-radius: 24px;
  background: rgba(8, 15, 29, 0.52);
  border: 1px solid rgba(148, 163, 184, 0.12);
  backdrop-filter: blur(10px);
}

.stage-card span,
.delivery-kicker {
  color: rgba(191, 219, 254, 0.76);
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.stage-card strong,
.value-card h3 {
  font-size: 1.35rem;
  font-variation-settings: "wght" 650;
}

.stage-card p,
.value-card p {
  margin: 0;
  color: rgba(226, 232, 240, 0.78);
  line-height: 1.7;
}

.stage-card.accent {
  background: linear-gradient(160deg, rgba(29, 107, 255, 0.24), rgba(12, 17, 26, 0.8));
}

.section {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.section-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
}

.section-head h2 {
  max-width: 920px;
  margin: 0;
  font-size: clamp(1.9rem, 4vw, 3rem);
  line-height: 1.08;
}

.section-head.compact {
  align-items: flex-start;
  justify-content: flex-start;
}

.value-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.value-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 22px;
}

.value-card p {
  color: var(--text-muted);
}

.delivery {
  padding: 24px;
}

.delivery-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  gap: 18px;
  align-items: stretch;
}

.delivery-column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.step-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0;
  padding-left: 22px;
}

.step-list li {
  color: var(--text);
  line-height: 1.7;
}

.matrix {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.matrix-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 18px;
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(29, 107, 255, 0.08), rgba(255, 255, 255, 0.96));
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.matrix-card span {
  color: var(--text-muted);
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.matrix-card strong {
  font-size: 1.05rem;
}

.faq-section {
  padding-bottom: 12px;
}

@media (max-width: 1100px) {
  .hero,
  .value-grid,
  .delivery-grid,
  .matrix {
    grid-template-columns: 1fr;
  }

  .site-header {
    position: static;
    flex-direction: column;
    align-items: flex-start;
    border-radius: 28px;
  }

  .site-nav {
    display: none;
  }

  .hero-stage,
  .hero-copy {
    min-height: auto;
  }
}

@media (max-width: 760px) {
  .home-shell {
    padding: 16px;
  }

  .site-header,
  .hero-copy,
  .hero-stage,
  .delivery {
    padding: 18px;
  }

.site-actions {
  width: 100%;
  flex-direction: column;
  align-items: stretch;
}

.site-actions :deep(.entry-popover),
.nav-cta {
  display: none;
}

  .mobile-nav-button {
    display: inline-flex;
    justify-content: center;
  }

  .hero-copy h1 {
    max-width: none;
    font-size: 2.5rem;
  }

  .stage-grid,
  .matrix {
    grid-template-columns: 1fr;
  }
}
</style>
