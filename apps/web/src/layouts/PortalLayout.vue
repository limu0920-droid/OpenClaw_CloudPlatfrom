<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const nav = [
  { label: '概览', to: '/portal' },
  { label: '实例', to: '/portal/instances' },
  { label: '渠道', to: '/portal/channels' },
  { label: '工单', to: '/portal/tickets' },
  { label: '任务', to: '/portal/jobs' },
  { label: '日志', to: '/portal/logs' },
]

const route = useRoute()
const active = computed(() => route.path)
</script>

<template>
  <div class="portal-shell">
    <aside class="portal-nav card">
      <div class="brand">
        <div class="brand-mark">OC</div>
        <div>
          <div class="brand-title">OpenClaw</div>
          <div class="brand-sub">Portal</div>
        </div>
      </div>
      <nav>
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          :class="['nav-item', active.startsWith(item.to) ? 'active' : '']"
        >
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
      <div class="nav-footer muted">
        <div>租户：Acme Corp</div>
        <div class="pill">试用 · 12 天剩余</div>
      </div>
    </aside>
    <main class="portal-main">
      <header class="portal-top card">
        <div>
          <div class="muted">OpenClaw 平台入口</div>
          <div class="page-title">租户控制台</div>
        </div>
        <div class="top-actions">
          <RouterLink class="ghost" to="/portal/logs">帮助与日志</RouterLink>
          <RouterLink class="primary" to="/portal/instances">实例列表</RouterLink>
        </div>
      </header>
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
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--panel);
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

.brand-title {
  font-weight: 700;
}

.brand-sub {
  color: var(--text-muted);
  font-size: 12px;
}

nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  padding: 10px 12px;
  border-radius: 12px;
  color: var(--text);
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.nav-item:hover {
  background: var(--panel-muted);
}

.nav-item.active {
  border-color: var(--brand);
  background: rgba(29, 107, 255, 0.08);
  color: var(--brand);
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
  padding: 16px 20px;
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

.ghost,
.primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 14px;
  border-radius: 999px;
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
  color: var(--text);
  cursor: pointer;
}

.primary {
  border-color: transparent;
  background: linear-gradient(120deg, var(--brand), var(--brand-strong));
  color: #fff;
  box-shadow: 0 8px 30px rgba(29, 107, 255, 0.35);
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
    flex-direction: row;
    gap: 10px;
    align-items: center;
  }
  nav {
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>
