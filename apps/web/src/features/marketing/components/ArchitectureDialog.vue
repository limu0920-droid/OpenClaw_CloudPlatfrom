<script setup lang="ts">
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { useRouter } from 'vue-router'

defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const router = useRouter()

function close() {
  emit('close')
}

function go(to: string) {
  close()
  void router.push(to)
}
</script>

<template>
  <TransitionRoot appear :show="open" as="template">
    <Dialog class="dialog-root" @close="close">
      <TransitionChild
        as="template"
        enter="backdrop-enter-active"
        enter-from="backdrop-enter-from"
        enter-to="backdrop-enter-to"
        leave="backdrop-leave-active"
        leave-from="backdrop-leave-from"
        leave-to="backdrop-leave-to"
      >
        <div class="dialog-backdrop" />
      </TransitionChild>

      <div class="dialog-wrap">
        <TransitionChild
          as="template"
          enter="panel-enter-active"
          enter-from="panel-enter-from"
          enter-to="panel-enter-to"
          leave="panel-leave-active"
          leave-from="panel-leave-from"
          leave-to="panel-leave-to"
        >
          <DialogPanel class="dialog-panel">
            <div class="dialog-top">
              <div>
                <div class="dialog-kicker">Headless UI 架构预览</div>
                <DialogTitle class="dialog-title">品牌入口、平台壳层和运行面分层表达</DialogTitle>
              </div>
              <button class="close-button" type="button" @click="close">关闭</button>
            </div>

            <div class="lane-grid">
              <article class="lane-card">
                <span>01</span>
                <strong>官网与登录</strong>
                <p>对外传达价值主张、体验入口和接入路径，让用户先理解产品再进入系统。</p>
              </article>
              <article class="lane-card">
                <span>02</span>
                <strong>Portal / Admin 壳层</strong>
                <p>把租户入口和平台运维控制面拆开，避免一个后台承担所有角色与流程。</p>
              </article>
              <article class="lane-card">
                <span>03</span>
                <strong>运行与集成平面</strong>
                <p>后续接入真实 IAM、搜索、runtime 与渠道时，前端结构无需重拆。</p>
              </article>
            </div>

            <div class="dialog-actions">
              <button class="secondary" type="button" @click="go('/login')">去登录页</button>
              <button class="primary" type="button" @click="go('/portal')">直接体验 Portal</button>
            </div>
          </DialogPanel>
        </TransitionChild>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<style scoped>
.dialog-root {
  position: fixed;
  inset: 0;
  z-index: 50;
}

.backdrop-enter-active,
.backdrop-leave-active {
  transition: opacity 0.2s ease;
}

.backdrop-enter-from,
.backdrop-leave-to {
  opacity: 0;
}

.backdrop-enter-to,
.backdrop-leave-from {
  opacity: 1;
}

.panel-enter-active,
.panel-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.panel-enter-from,
.panel-leave-to {
  opacity: 0;
  transform: translateY(16px) scale(0.96);
}

.panel-enter-to,
.panel-leave-from {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.dialog-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(10px);
}

.dialog-wrap {
  position: fixed;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 20px;
}

.dialog-panel {
  width: min(980px, 100%);
  padding: 24px;
  border-radius: 32px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    radial-gradient(circle at top left, rgba(29, 107, 255, 0.16), transparent 26%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.98), rgba(242, 246, 255, 0.98));
  box-shadow: 0 28px 80px rgba(15, 23, 42, 0.18);
}

.dialog-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.dialog-kicker {
  color: var(--brand);
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.dialog-title {
  margin-top: 8px;
  font-size: clamp(1.8rem, 3vw, 2.6rem);
  line-height: 1.1;
}

.close-button,
.secondary,
.primary {
  padding: 11px 16px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  font: inherit;
}

.close-button,
.secondary {
  background: rgba(255, 255, 255, 0.82);
}

.primary {
  border-color: transparent;
  background: linear-gradient(120deg, #0f6bff, #5c7bff 65%, #1ad1ff);
  color: #fff;
}

.lane-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 24px;
}

.lane-card {
  padding: 18px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(148, 163, 184, 0.14);
}

.lane-card span {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgba(29, 107, 255, 0.1);
  color: var(--brand);
  font-variation-settings: "wght" 620;
}

.lane-card strong {
  display: block;
  margin-top: 14px;
  font-size: 1.12rem;
}

.lane-card p {
  margin: 10px 0 0;
  line-height: 1.75;
  color: var(--text-muted);
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 24px;
}

@media (max-width: 900px) {
  .lane-grid {
    grid-template-columns: 1fr;
  }

  .dialog-top,
  .dialog-actions {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
