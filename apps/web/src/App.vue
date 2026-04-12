<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import { computed, onMounted, watchEffect } from 'vue'

import { applyBrandingForRoute, ensureBrandingLoaded } from './lib/brand'

const route = useRoute()
const themeClass = computed(() => {
  if (route.meta.theme === 'admin') return 'theme-admin'
  if (route.meta.theme === 'marketing') return 'theme-marketing'
  return 'theme-portal'
})

onMounted(() => {
  void ensureBrandingLoaded()
})

watchEffect(() => {
  applyBrandingForRoute(route.path, String(route.meta.theme ?? 'portal'))
})
</script>

<template>
  <div :class="['app-root', themeClass]">
    <RouterView />
  </div>
</template>
