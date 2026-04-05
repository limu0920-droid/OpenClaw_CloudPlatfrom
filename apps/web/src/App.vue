<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const themeClass = computed(() => (route.meta.theme === 'admin' ? 'theme-admin' : 'theme-portal'))
</script>

<template>
  <div :class="['app-root', themeClass]">
    <RouterView />
  </div>
</template>
