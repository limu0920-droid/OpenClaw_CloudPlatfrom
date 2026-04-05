<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const nav = [
  { label: '总览', to: '/admin' },
  { label: '租户', to: '/admin/tenants' },
  { label: '实例', to: '/admin/instances' },
  { label: '渠道', to: '/admin/channels' },
  { label: '工单', to: '/admin/tickets' },
  { label: '任务', to: '/admin/jobs' },
  { label: '告警', to: '/admin/alerts' },
  { label: '审计', to: '/admin/audit' },
]

const route = useRoute()
const active = computed(() => route.path)
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-nav card">
      <div class="brand">
        <div class="brand-mark">OC</div>
        <div>
          <div class="brand-title">OpenClaw</div>
          <div class="brand-sub">Admin Control</div>
        </div>
      </div>
      <nav>
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          :class="['nav-item', active.startsWith(item.to) ? 'active' : '']"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
      <div class="nav-footer">
        <div class="pill">运行中：8 副本</div>
        <div class="muted">平台管理员</div>
      </div>
    </aside>
    <main class="admin-main">
      <header class="admin-top card">
        <div>
          <div class="muted">平台控制面</div>
          <div class="page-title">运维与审计</div>
        </div>
        <div class="top-actions">
          <RouterLink class="ghost" to="/admin/alerts">告警中心</RouterLink>
          <RouterLink class="primary" to="/admin/jobs">发布维护窗口</RouterLink>
        </div>
      </header>
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
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: var(--panel);
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
  background: rgba(255, 255, 255, 0.06);
}

.nav-item.active {
  border-color: var(--brand);
  background: rgba(59, 130, 246, 0.12);
  color: #e5edff;
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
  padding: 16px 20px;
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
  background: linear-gradient(120deg, #1e40af, var(--brand));
  color: #e5edff;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.35);
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
    flex-direction: row;
    flex-wrap: wrap;
    gap: 10px;
  }
  nav {
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>
