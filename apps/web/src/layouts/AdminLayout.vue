<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'

import { useBranding } from '../lib/brand'

const baseNav = [
  { label: '总览', to: '/admin', feature: 'adminEnabled' },
  { label: '租户', to: '/admin/tenants', feature: 'adminEnabled' },
  { label: '实例', to: '/admin/instances', feature: 'adminEnabled' },
  { label: '产物', to: '/admin/artifacts', feature: 'adminEnabled' },
  { label: '渠道', to: '/admin/channels', feature: 'channelsEnabled' },
  { label: '工单', to: '/admin/tickets', feature: 'ticketsEnabled' },
  { label: '任务', to: '/admin/jobs', feature: 'adminEnabled' },
  { label: 'OEM', to: '/admin/oem/brands', feature: 'adminEnabled' },
  { label: '审批', to: '/admin/approvals', feature: 'adminEnabled' },
  { label: '诊断', to: '/admin/diagnostics', feature: 'adminEnabled' },
  { label: '告警', to: '/admin/alerts', feature: 'adminEnabled' },
  { label: '审计', to: '/admin/audit', feature: 'auditEnabled' },
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
  () => nav.value.find((item) => (item.to === '/admin' ? route.path === item.to : route.path.startsWith(item.to)))?.to ?? '/admin',
)
</script>

<template>
  <div class="admin-shell">
    <el-card shadow="never" class="admin-nav">
      <div class="brand">
        <div class="brand-mark">
          <img v-if="brand.logoUrl" :src="brand.logoUrl" :alt="brand.name" class="brand-logo" />
          <span v-else>{{ brand.name.slice(0, 2).toUpperCase() }}</span>
        </div>
        <div>
          <div class="brand-title">{{ brand.name }}</div>
          <div class="brand-sub">Admin Control</div>
        </div>
      </div>
      <el-scrollbar class="nav-scroll">
        <el-menu :default-active="currentMenu" :router="true" class="admin-menu">
          <el-menu-item v-for="item in nav" :key="item.to" :index="item.to">
            {{ item.label }}
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
      <div class="nav-footer">
        <el-tag round effect="dark" disable-transitions>运行中：8 副本</el-tag>
        <div class="muted">{{ brand.supportEmail || '平台管理员' }}</div>
      </div>
    </el-card>
    <main class="admin-main">
      <el-card shadow="never" class="admin-top">
        <div>
          <div class="muted">{{ brand.name }} 控制面</div>
          <div class="page-title">运维与审计</div>
        </div>
        <div class="top-actions">
          <el-button round plain @click="router.push('/admin/approvals')">审批中心</el-button>
          <el-button round plain @click="router.push('/admin/diagnostics')">诊断中心</el-button>
          <el-button round type="primary" @click="router.push('/admin/alerts')">告警中心</el-button>
        </div>
      </el-card>
      <section class="admin-content">
        <RouterView />
      </section>
    </main>
  </div>
</template>

<style scoped>
.admin-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 18px;
  padding: 18px 18px 28px;
}

.admin-nav {
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
  padding-bottom: 8px;
  border-bottom: 1px solid var(--stroke);
}

.brand-mark {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  background: linear-gradient(135deg, #0ea5e9, var(--brand));
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: 700;
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

.admin-menu {
  border-right: none;
  background: transparent;
}

:deep(.admin-menu .el-menu-item) {
  height: 46px;
  margin-bottom: 6px;
  border-radius: 14px;
  color: var(--text);
  font-size: 14px;
}

:deep(.admin-menu .el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.06);
}

:deep(.admin-menu .el-menu-item.is-active) {
  background: rgba(59, 130, 246, 0.12);
  color: #e5edff;
  font-variation-settings: "wght" 620;
}

.nav-footer {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: var(--text-muted);
}

.admin-main {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.admin-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.85), rgba(17, 24, 39, 0.95));
}

.page-title {
  font-size: 20px;
  font-weight: 700;
  color: #e5e7eb;
}

.top-actions {
  display: flex;
  gap: 10px;
}

.admin-content {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

@media (max-width: 1100px) {
  .admin-shell {
    grid-template-columns: 1fr;
  }
  .admin-nav {
    position: static;
    min-height: auto;
  }
}
</style>

