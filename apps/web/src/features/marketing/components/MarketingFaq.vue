<script setup lang="ts">
import { Disclosure, DisclosureButton, DisclosurePanel } from '@headlessui/vue'

const faqs = [
  {
    question: '为什么还要引入 Headless UI，而不是继续堆现成组件？',
    answer: '因为官网和品牌表达更需要可控的视觉语言。Headless UI 给出的是可访问性的交互骨架，不会把视觉直接锁死在通用后台风格上。',
  },
  {
    question: 'Portal 和 Admin 之后会拆成两个应用吗？',
    answer: '第一阶段不急着拆。当前先共用一个 Vue 应用和双布局，等真实鉴权、租户边界和部署方式稳定后再决定是否拆分。',
  },
  {
    question: '这套官网和控制台结构能否继续接真实后端？',
    answer: '可以。当前前端已经按实例、工单、审计、渠道等实体拆分，接真实 IAM、Search 和 runtime 时只需要替换接口和流程，不需要重做信息架构。',
  },
]
</script>

<template>
  <div class="faq-list">
    <Disclosure v-for="item in faqs" :key="item.question" v-slot="{ open }" as="div" class="faq-item">
      <DisclosureButton :class="['faq-trigger', open ? 'is-open' : '']">
        <span>{{ item.question }}</span>
        <span class="faq-mark">{{ open ? '-' : '+' }}</span>
      </DisclosureButton>
      <DisclosurePanel class="faq-panel">
        <p>{{ item.answer }}</p>
      </DisclosurePanel>
    </Disclosure>
  </div>
</template>

<style scoped>
.faq-list {
  display: grid;
  gap: 12px;
}

.faq-item {
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(255, 255, 255, 0.82);
  overflow: hidden;
}

.faq-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px 20px;
  text-align: left;
  background: transparent;
}

.faq-trigger span:first-child {
  font-size: 1rem;
  font-variation-settings: "wght" 620;
}

.faq-trigger.is-open {
  background: rgba(29, 107, 255, 0.06);
}

.faq-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 999px;
  background: rgba(29, 107, 255, 0.08);
  color: var(--brand);
  font-size: 1rem;
}

.faq-panel {
  padding: 0 20px 18px;
}

.faq-panel p {
  margin: 0;
  line-height: 1.8;
  color: var(--text-muted);
}
</style>
