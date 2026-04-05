<script setup lang="ts">
defineProps<{
  title: string
  items: Array<{ id: string; label: string; meta?: string; status?: string; accent?: string }>
}>()
</script>

<template>
  <div class="card list-card">
    <div class="title">{{ title }}</div>
    <ul>
      <li v-for="item in items" :key="item.id">
        <div class="label">
          <span class="dot" :style="{ background: item.accent || 'var(--brand)' }"></span>
          {{ item.label }}
        </div>
        <div class="meta muted">{{ item.meta }}</div>
        <div v-if="item.status" class="pill">{{ item.status }}</div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.list-card {
  padding: 14px;
}
.title {
  font-weight: 600;
  margin-bottom: 10px;
}
ul {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
li {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 10px;
  align-items: center;
}
.label {
  display: flex;
  gap: 8px;
  align-items: center;
  font-weight: 500;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.pill {
  justify-self: end;
  background: var(--panel-muted);
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
}
@media (max-width: 720px) {
  li {
    grid-template-columns: 1fr;
    align-items: flex-start;
  }
}
</style>
