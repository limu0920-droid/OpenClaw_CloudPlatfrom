<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { PortalArtifactCenterItem } from '../../lib/types'

const router = useRouter()

const loading = ref(true)
const error = ref('')
const items = ref<PortalArtifactCenterItem[]>([])
const selectedId = ref('')
const query = ref('')
const kind = ref('')
const instanceId = ref('')

const selected = computed(() => items.value.find((item) => item.id === selectedId.value) ?? items.value[0] ?? null)
const kindOptions = computed(() => Array.from(new Set(items.value.map((item) => item.kind))).sort())
const instanceOptions = computed(() => {
  const map = new Map<string, string>()
  items.value.forEach((item) => {
    map.set(item.instanceId, item.instanceName)
  })
  return Array.from(map.entries()).map(([id, name]) => ({ id, name }))
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const response = await api.getPortalArtifactCenter({
      q: query.value.trim() || undefined,
      kind: kind.value || undefined,
      instanceId: instanceId.value || undefined,
    })
    items.value = response.items
    if (!items.value.find((item) => item.id === selectedId.value)) {
      selectedId.value = items.value[0]?.id ?? ''
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载产物中心失败'
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  query.value = ''
  kind.value = ''
  instanceId.value = ''
  void load()
}

function selectItem(id: string) {
  selectedId.value = id
}

function openWorkspace() {
  if (!selected.value) return
  router.push(selected.value.workspacePath)
}

function openExternal() {
  const target = selected.value?.preview.externalUrl || selected.value?.sourceUrl
  if (!target) return
  window.open(target, '_blank', 'noopener,noreferrer')
}

function openDownload() {
  const target = selected.value?.preview.downloadUrl || selected.value?.sourceUrl
  if (!target) return
  window.open(target, '_blank', 'noopener,noreferrer')
}

function previewMode(item: PortalArtifactCenterItem | null) {
  if (!item?.preview.available || !item.preview.previewUrl) {
    return 'none'
  }
  if (item.preview.mode === 'image') return 'image'
  if (item.preview.mode === 'video') return 'video'
  if (item.preview.mode === 'audio') return 'audio'
  return 'iframe'
}

onMounted(load)
</script>

<template>
  <div class="artifact-shell">
    <el-card shadow="never" class="toolbar-card">
      <SectionHeader title="产物中心" subtitle="按实例、会话与文件类型统一查看平台侧归档产物" />
      <div class="toolbar-grid">
        <el-input v-model="query" placeholder="搜索标题、实例、会话或来源地址" clearable @keyup.enter="load" />
        <el-select v-model="kind" placeholder="全部类型" clearable>
          <el-option v-for="item in kindOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="instanceId" placeholder="全部实例" clearable>
          <el-option v-for="item in instanceOptions" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
        <div class="toolbar-actions">
          <el-button type="primary" @click="load">检索</el-button>
          <el-button plain @click="clearFilters">清空</el-button>
        </div>
      </div>
    </el-card>

    <el-card v-if="loading" shadow="never" class="state-card">正在加载产物中心…</el-card>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
    <section v-else class="artifact-layout">
      <el-card shadow="never" class="artifact-list-card">
        <SectionHeader title="归档列表" subtitle="选择一份产物查看正式预览方案与回路入口" />
        <div v-if="items.length" class="artifact-list">
          <button
            v-for="item in items"
            :key="item.id"
            type="button"
            :class="['artifact-item', { active: selected?.id === item.id }]"
            @click="selectItem(item.id)"
          >
            <div class="artifact-item-head">
              <strong>{{ item.title }}</strong>
              <el-tag round disable-transitions>{{ item.kind }}</el-tag>
            </div>
            <div class="artifact-item-meta">
              <span>{{ item.instanceName }}</span>
              <span>{{ item.sessionTitle }}</span>
              <span>{{ item.updatedAt }}</span>
            </div>
            <div class="artifact-item-foot">
              <span>{{ item.archiveStatus || 'registered' }}</span>
              <span>{{ item.sizeLabel || '未记录大小' }}</span>
            </div>
          </button>
        </div>
        <div v-else class="empty-card">当前筛选条件下没有找到产物。</div>
      </el-card>

      <el-card shadow="never" class="artifact-preview-card">
        <SectionHeader title="预览与交付" subtitle="优先使用平台代理预览，失败时回退到下载或外部地址" />
        <template v-if="selected">
          <div class="selected-meta">
            <div class="meta-row">
              <span class="muted">实例</span>
              <strong>{{ selected.instanceName }}</strong>
            </div>
            <div class="meta-row">
              <span class="muted">会话</span>
              <strong>{{ selected.sessionTitle }}</strong>
            </div>
            <div class="meta-row">
              <span class="muted">归档状态</span>
              <strong>{{ selected.archiveStatus || 'registered' }}</strong>
            </div>
            <div class="meta-row">
              <span class="muted">文件大小</span>
              <strong>{{ selected.sizeLabel || '未记录' }}</strong>
            </div>
            <div class="meta-row">
              <span class="muted">预览策略</span>
              <strong>{{ selected.preview.strategy }}</strong>
            </div>
          </div>

          <div class="preview-actions">
            <el-button type="primary" @click="openWorkspace">回到工作台</el-button>
            <el-button plain :disabled="!selected.preview.downloadUrl && !selected.sourceUrl" @click="openDownload">下载 / 打开</el-button>
            <el-button plain :disabled="!selected.preview.externalUrl && !selected.sourceUrl" @click="openExternal">外部地址</el-button>
          </div>

          <div v-if="selected.preview.note" class="preview-note muted">{{ selected.preview.note }}</div>
          <div v-if="selected.preview.failureReason" class="preview-note preview-note--warn">{{ selected.preview.failureReason }}</div>

          <iframe
            v-if="previewMode(selected) === 'iframe'"
            :src="selected.preview.previewUrl"
            class="preview-frame"
            title="Artifact Preview"
            referrerpolicy="strict-origin-when-cross-origin"
          />
          <img v-else-if="previewMode(selected) === 'image'" :src="selected.preview.previewUrl" class="preview-image" alt="Artifact Preview" />
          <video v-else-if="previewMode(selected) === 'video'" :src="selected.preview.previewUrl" class="preview-media" controls playsinline />
          <audio v-else-if="previewMode(selected) === 'audio'" :src="selected.preview.previewUrl" class="preview-audio" controls />
          <div v-else class="empty-card">
            当前产物没有可嵌入的正式预览，使用上方按钮回到工作台或直接下载。
          </div>
        </template>
        <div v-else class="empty-card">请选择一份产物查看预览与交付信息。</div>
      </el-card>
    </section>
  </div>
</template>

<style scoped>
.artifact-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.toolbar-card,
.artifact-list-card,
.artifact-preview-card,
.state-card {
  padding: 18px;
}

.toolbar-grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr auto;
  gap: 10px;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
}

.artifact-layout {
  display: grid;
  grid-template-columns: 360px minmax(0, 1fr);
  gap: 14px;
}

.artifact-list,
.selected-meta {
  display: grid;
  gap: 10px;
}

.artifact-item,
.empty-card,
.selected-meta,
.preview-note {
  padding: 14px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.artifact-item {
  text-align: left;
  display: grid;
  gap: 8px;
}

.artifact-item.active {
  border-color: rgba(29, 107, 255, 0.28);
  background: rgba(29, 107, 255, 0.08);
}

.artifact-item-head,
.artifact-item-meta,
.artifact-item-foot,
.preview-actions,
.meta-row {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: center;
}

.artifact-item-meta,
.artifact-item-foot {
  flex-wrap: wrap;
  color: var(--text-muted);
  font-size: 13px;
}

.artifact-preview-card {
  display: grid;
  gap: 12px;
}

.preview-actions {
  flex-wrap: wrap;
}

.preview-note--warn {
  color: #b45309;
}

.preview-frame {
  width: 100%;
  min-height: 68vh;
  border: none;
  border-radius: 18px;
  background: #fff;
}

.preview-image,
.preview-media {
  width: 100%;
  border-radius: 18px;
  background: #000;
}

.preview-audio {
  width: 100%;
}

.empty-card,
.state-card {
  text-align: center;
  color: var(--text-muted);
}

@media (max-width: 1200px) {
  .artifact-layout,
  .toolbar-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .artifact-item-head,
  .artifact-item-meta,
  .artifact-item-foot,
  .preview-actions,
  .meta-row,
  .toolbar-actions {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
