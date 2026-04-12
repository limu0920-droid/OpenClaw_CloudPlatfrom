<script setup lang="ts">
import { onMounted, ref } from 'vue'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { OEMBrandRecord } from '../../lib/types'

const loading = ref(true)
const error = ref('')
const items = ref<OEMBrandRecord[]>([])

async function load() {
  loading.value = true
  error.value = ''
  try {
    items.value = await api.getAdminOEMBrands()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载 OEM 品牌失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="stack">
    <el-card shadow="never" class="panel">
      <SectionHeader
        title="OEM 品牌"
        subtitle="查看品牌、主题、功能开关和租户绑定的当前后端状态。"
      />
    </el-card>

    <el-card v-if="loading" shadow="never" class="panel">正在加载 OEM 品牌…</el-card>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />

    <div v-else class="grid">
      <el-card v-for="item in items" :key="item.brand.id" shadow="never" class="panel brand-card">
        <div class="brand-head">
          <div>
            <div class="title-row">
              <strong>{{ item.brand.name }}</strong>
              <el-tag size="small" round>{{ item.brand.code }}</el-tag>
              <el-tag size="small" round type="success">{{ item.brand.status }}</el-tag>
            </div>
            <p class="muted">{{ item.brand.supportEmail || '未配置支持邮箱' }}</p>
          </div>
          <div class="muted">{{ item.bindings.length }} 个租户绑定</div>
        </div>

        <div class="meta-grid">
          <div class="meta-block">
            <span class="meta-label">主题</span>
            <div class="meta-content">
              {{ item.theme?.primaryColor || '未配置主色' }} / {{ item.theme?.surfaceMode || '未配置模式' }}
            </div>
          </div>
          <div class="meta-block">
            <span class="meta-label">能力</span>
            <div class="meta-content">
              Portal {{ item.features?.portalEnabled ? '开' : '关' }}，
              Admin {{ item.features?.adminEnabled ? '开' : '关' }}，
              Audit {{ item.features?.auditEnabled ? '开' : '关' }}
            </div>
          </div>
          <div class="meta-block">
            <span class="meta-label">域名</span>
            <div class="meta-content">
              {{ item.brand.domains.length ? item.brand.domains.join(', ') : '未配置域名' }}
            </div>
          </div>
          <div class="meta-block">
            <span class="meta-label">绑定租户</span>
            <div class="meta-content">
              {{ item.bindings.length ? item.bindings.map((binding) => `${binding.tenantId}:${binding.bindingMode}`).join(' / ') : '未绑定租户' }}
            </div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.grid {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.panel {
  padding: 16px;
}

.brand-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.brand-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.title-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.meta-grid {
  display: grid;
  gap: 12px;
}

.meta-block {
  display: grid;
  gap: 4px;
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--stroke);
  background: var(--panel-muted);
}

.meta-label {
  color: var(--text-muted);
  font-size: 12px;
}

.meta-content {
  word-break: break-word;
}

@media (max-width: 1100px) {
  .grid {
    grid-template-columns: 1fr;
  }

  .brand-head {
    flex-direction: column;
  }
}
</style>
