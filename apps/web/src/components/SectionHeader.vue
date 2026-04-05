<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
  actionLabel?: string
}>()
</script>

<template>
  <div class="section-head">
    <div>
      <div class="title">{{ title }}</div>
      <div v-if="subtitle" class="subtitle">{{ subtitle }}</div>
    </div>
    <button v-if="actionLabel" class="ghost">{{ actionLabel }}</button>
  </div>
</template>

<style scoped>
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.title {
  font-weight: 700;
  font-size: 16px;
}
.subtitle {
  color: var(--text-muted);
  font-size: 13px;
}
.ghost {
  padding: 8px 12px;
  border-radius: 10px;
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
  color: var(--text);
  cursor: pointer;
}
</style>
