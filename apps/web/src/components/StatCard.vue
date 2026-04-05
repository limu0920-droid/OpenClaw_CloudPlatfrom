<script setup lang="ts">
defineProps<{
  label: string
  value: string
  delta?: string
  tone?: 'neutral' | 'positive' | 'warning' | 'critical'
}>()

const toneClass = (tone?: string) => {
  if (tone === 'positive') return 'positive'
  if (tone === 'warning') return 'warning'
  if (tone === 'critical') return 'critical'
  return 'neutral'
}
</script>

<template>
  <div class="stat card" :class="toneClass(tone)">
    <div class="label">{{ label }}</div>
    <div class="value">{{ value }}</div>
    <div v-if="delta" class="delta">{{ delta }}</div>
  </div>
</template>

<style scoped>
.stat {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.label {
  color: var(--text-muted);
  font-size: 13px;
}
.value {
  font-size: 24px;
  font-weight: 700;
}
.delta {
  font-size: 12px;
  color: var(--text-muted);
}
.positive .delta {
  color: #16a34a;
}
.warning .delta {
  color: #f59e0b;
}
.critical .delta {
  color: #ef4444;
}
</style>
