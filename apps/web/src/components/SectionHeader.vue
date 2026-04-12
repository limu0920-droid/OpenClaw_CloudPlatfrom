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
    <el-button v-if="actionLabel" round plain>{{ actionLabel }}</el-button>
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
</style>
