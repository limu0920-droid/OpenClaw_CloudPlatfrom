<script setup lang="ts">
import { Popover, PopoverButton, PopoverPanel } from '@headlessui/vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const entries = [
  {
    label: '登录入口',
    description: '统一进入登录页，后续继续承接真实 IAM。',
    to: '/login',
  },
  {
    label: 'Portal',
    description: '租户用户的实例、渠道、备份与工单入口。',
    to: '/portal',
  },
  {
    label: 'Admin',
    description: '平台运维的租户、任务、告警与审计控制面。',
    to: '/admin',
  },
]

function navigate(to: string, close: () => void) {
  close()
  void router.push(to)
}
</script>

<template>
  <Popover class="entry-popover" v-slot="{ open, close }">
    <PopoverButton :class="['entry-trigger', open ? 'is-open' : '']">体验入口</PopoverButton>
    <transition name="entry-fade">
      <PopoverPanel class="entry-panel">
        <div class="entry-grid">
          <button
            v-for="entry in entries"
            :key="entry.to"
            class="entry-item"
            type="button"
            @click="navigate(entry.to, close)"
          >
            <strong>{{ entry.label }}</strong>
            <span>{{ entry.description }}</span>
          </button>
        </div>
      </PopoverPanel>
    </transition>
  </Popover>
</template>

<style scoped>
.entry-popover {
  position: relative;
}

.entry-trigger {
  min-width: 116px;
  padding: 10px 16px;
  border-radius: 999px;
  border: 1px solid rgba(29, 107, 255, 0.14);
  background: rgba(29, 107, 255, 0.08);
  color: var(--brand);
  font: inherit;
  font-variation-settings: "wght" 600;
}

.entry-trigger.is-open {
  background: rgba(29, 107, 255, 0.14);
}

.entry-panel {
  position: absolute;
  top: calc(100% + 12px);
  right: 0;
  width: min(360px, calc(100vw - 40px));
  padding: 12px;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 24px 56px rgba(15, 23, 42, 0.14);
  backdrop-filter: blur(20px);
}

.entry-grid {
  display: grid;
  gap: 10px;
}

.entry-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  text-align: left;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(239, 244, 255, 0.96));
}

.entry-item strong {
  font-size: 1rem;
  font-variation-settings: "wght" 640;
}

.entry-item span {
  color: var(--text-muted);
  line-height: 1.6;
}

.entry-item:hover {
  border-color: rgba(29, 107, 255, 0.22);
  transform: translateY(-1px);
}

.entry-fade-enter-active,
.entry-fade-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.entry-fade-enter-from,
.entry-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
