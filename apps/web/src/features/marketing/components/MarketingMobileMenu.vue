<script setup lang="ts">
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { useRouter } from 'vue-router'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const router = useRouter()

const links = [
  { label: '价值主张', id: 'values' },
  { label: '控制台', id: 'planes' },
  { label: '上线节奏', id: 'delivery' },
]

function close() {
  emit('close')
}

function scrollTo(id: string) {
  close()
  requestAnimationFrame(() => {
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  })
}

function go(to: string) {
  close()
  void router.push(to)
}
</script>

<template>
  <TransitionRoot appear :show="props.open" as="template">
    <Dialog class="mobile-menu-root" @close="close">
      <TransitionChild
        as="template"
        enter="mobile-fade-enter-active"
        enter-from="mobile-fade-enter-from"
        enter-to="mobile-fade-enter-to"
        leave="mobile-fade-leave-active"
        leave-from="mobile-fade-leave-from"
        leave-to="mobile-fade-leave-to"
      >
        <div class="mobile-menu-backdrop" />
      </TransitionChild>

      <div class="mobile-menu-wrap">
        <TransitionChild
          as="template"
          enter="mobile-panel-enter-active"
          enter-from="mobile-panel-enter-from"
          enter-to="mobile-panel-enter-to"
          leave="mobile-panel-leave-active"
          leave-from="mobile-panel-leave-from"
          leave-to="mobile-panel-leave-to"
        >
          <DialogPanel class="mobile-menu-panel">
            <div class="mobile-menu-head">
              <div>
                <div class="mobile-menu-kicker">导航</div>
                <DialogTitle class="mobile-menu-title">OpenClaw Platform</DialogTitle>
              </div>
              <button class="mobile-close" type="button" @click="close">关闭</button>
            </div>

            <div class="mobile-links">
              <button v-for="item in links" :key="item.id" class="mobile-link" type="button" @click="scrollTo(item.id)">
                <strong>{{ item.label }}</strong>
                <span>跳转到官网对应章节</span>
              </button>
            </div>

            <div class="mobile-actions">
              <button class="mobile-action secondary" type="button" @click="go('/login')">登录入口</button>
              <button class="mobile-action secondary" type="button" @click="go('/portal')">进入 Portal</button>
              <button class="mobile-action primary" type="button" @click="go('/admin')">进入 Admin</button>
            </div>
          </DialogPanel>
        </TransitionChild>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<style scoped>
.mobile-menu-root {
  position: fixed;
  inset: 0;
  z-index: 60;
}

.mobile-fade-enter-active,
.mobile-fade-leave-active {
  transition: opacity 0.18s ease;
}

.mobile-fade-enter-from,
.mobile-fade-leave-to {
  opacity: 0;
}

.mobile-fade-enter-to,
.mobile-fade-leave-from {
  opacity: 1;
}

.mobile-panel-enter-active,
.mobile-panel-leave-active {
  transition:
    opacity 0.22s ease,
    transform 0.22s ease;
}

.mobile-panel-enter-from,
.mobile-panel-leave-to {
  opacity: 0;
  transform: translateX(24px);
}

.mobile-panel-enter-to,
.mobile-panel-leave-from {
  opacity: 1;
  transform: translateX(0);
}

.mobile-menu-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(10px);
}

.mobile-menu-wrap {
  position: fixed;
  inset: 0;
  display: flex;
  justify-content: flex-end;
}

.mobile-menu-panel {
  width: min(420px, 100%);
  height: 100%;
  padding: 20px;
  background:
    radial-gradient(circle at top left, rgba(29, 107, 255, 0.16), transparent 24%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(242, 246, 255, 0.98));
  box-shadow: -18px 0 52px rgba(15, 23, 42, 0.12);
}

.mobile-menu-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.mobile-menu-kicker {
  color: var(--brand);
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.mobile-menu-title {
  margin-top: 8px;
  font-size: 1.8rem;
}

.mobile-close,
.mobile-action,
.mobile-link {
  font: inherit;
}

.mobile-close {
  padding: 10px 14px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.76);
}

.mobile-links {
  display: grid;
  gap: 10px;
  margin-top: 28px;
}

.mobile-link {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px 18px;
  text-align: left;
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(255, 255, 255, 0.78);
}

.mobile-link strong {
  font-size: 1rem;
  font-variation-settings: "wght" 640;
}

.mobile-link span {
  color: var(--text-muted);
}

.mobile-actions {
  display: grid;
  gap: 10px;
  margin-top: 28px;
}

.mobile-action {
  padding: 13px 18px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.mobile-action.secondary {
  background: rgba(255, 255, 255, 0.78);
}

.mobile-action.primary {
  border-color: transparent;
  background: linear-gradient(120deg, #0f6bff, #5c7bff 65%, #1ad1ff);
  color: #fff;
}
</style>
