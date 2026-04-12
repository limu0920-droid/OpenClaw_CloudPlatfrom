<script setup lang="ts">
import { RouterView, useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'

import { useBranding } from '../lib/brand'

const baseNav = [
  { label: '自助中心', to: '/portal', feature: 'portalEnabled' },
  { label: '产物', to: '/portal/artifacts', feature: 'portalEnabled' },
  { label: '实例', to: '/portal/instances', feature: 'portalEnabled' },
  { label: '渠道', to: '/portal/channels', feature: 'channelsEnabled' },
  { label: '工单', to: '/portal/tickets', feature: 'ticketsEnabled' },
  { label: '运营', to: '/portal/operations', feature: 'portalEnabled' },
  { label: '任务', to: '/portal/jobs', feature: 'portalEnabled' },
  { label: '日志', to: '/portal/logs', feature: 'portalEnabled' },
] as const

const route = useRoute()
const router = useRouter()
const { brand, features } = useBranding()
const nav = computed(() =>
  baseNav.filter((item) => {
    const key = item.feature
    return key ? Boolean(features.value[key]) : true
  }),
)
const currentMenu = computed(
  () => nav.value.find((item) => (item.to === '/portal' ? route.path === item.to : route.path.startsWith(item.to)))?.to ?? '/portal',
)
</script>

<template>
  <div class="portal-shell">
    <el-card shadow="never" class="portal-nav">
      <div class="brand">
        <div class="brand-mark">
          <img v-if="brand.logoUrl" :src="brand.logoUrl" :alt="brand.name" class="brand-logo" />
          <span v-else>{{ brand.name.slice(0, 2).toUpperCase() }}</span>
        </div>
        <div>
          <div class="brand-title">{{ brand.name }}</div>
          <div class="brand-sub">{{ features.portalEnabled ? 'Portal' : 'Brand Workspace' }}</div>
        </div>
      </div>
      <el-scrollbar class="nav-scroll">
        <el-menu :default-active="currentMenu" :router="true" class="portal-menu">
          <el-menu-item v-for="item in nav" :key="item.to" :index="item.to">
            {{ item.label }}
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
      <div class="nav-footer muted">
        <div>{{ brand.supportEmail || 'Portal 自助使用视角' }}</div>
        <el-tag round disable-transitions>{{ brand.code || 'Self-Service' }}</el-tag>
      </div>
    </el-card>
    <main class="portal-main">
      <el-card shadow="never" class="portal-top">
        <div>
          <div class="muted">{{ brand.name }} 平台入口</div>
          <div class="page-title">{{ brand.name }} 租户控制台</div>
        </div>
        <div class="top-actions">
          <el-button round plain @click="router.push('/portal/artifacts')">产物中心</el-button>
          <el-button round plain @click="router.push('/portal/operations')">运营中心</el-button>
          <el-button round type="primary" @click="router.push('/portal/instances')">实例列表</el-button>
        </div>
      </el-card>
      <section class="portal-content">
        <RouterView />
      </section>
    </main>
  </div>
</template>

<style scoped>
.portal-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 18px;
  padding: 18px 18px 28px;
}

.portal-nav {
  position: sticky;
  top: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: calc(100vh - 36px);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--stroke);
}

.brand-mark {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  background: linear-gradient(135deg, var(--brand), var(--brand-strong));
  color: #fff;
  font-weight: 700;
  display: grid;
  place-items: center;
  letter-spacing: 0.5px;
}

.brand-logo {
  width: 26px;
  height: 26px;
  object-fit: contain;
}

.brand-title {
  font-weight: 700;
}

.brand-sub {
  color: var(--text-muted);
  font-size: 12px;
}

.nav-scroll {
  flex: 1;
}

.portal-menu {
  border-right: none;
  background: transparent;
}

:deep(.portal-menu .el-menu-item) {
  height: 46px;
  margin-bottom: 6px;
  border-radius: 14px;
  color: var(--text);
  font-size: 14px;
}

:deep(.portal-menu .el-menu-item:hover) {
  background: var(--panel-muted);
}

:deep(.portal-menu .el-menu-item.is-active) {
  background: rgba(29, 107, 255, 0.08);
  color: var(--brand);
  font-variation-settings: "wght" 620;
}

.nav-footer {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
}

.portal-main {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.portal-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 20px;
  font-weight: 700;
}

.top-actions {
  display: flex;
  gap: 10px;
}

.portal-content {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

@media (max-width: 1100px) {
  .portal-shell {
    grid-template-columns: 1fr;
  }
  .portal-nav {
    position: static;
    min-height: auto;
  }
}
</style>
