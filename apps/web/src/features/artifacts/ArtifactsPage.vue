<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { ArtifactCenterResponse, PortalArtifactCenterItem } from '../../lib/types'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const favoriteLoadingId = ref('')
const error = ref('')
const query = ref('')
const kind = ref('')
const response = ref<ArtifactCenterResponse>({
  items: [],
  recentViewed: [],
  stats: {
    totalCount: 0,
    favoriteCount: 0,
    sharedCount: 0,
    versionedCount: 0,
    inlinePreviewCount: 0,
    fallbackCount: 0,
    recentViewedCount: 0,
    failureReasons: [],
  },
})

const scope = computed<'portal' | 'admin'>(() => (route.path.startsWith('/admin') ? 'admin' : 'portal'))
const title = computed(() => (scope.value === 'admin' ? '平台产物中心' : '我的产物中心'))
const subtitle = computed(() =>
  scope.value === 'admin'
    ? '跨租户管理产物版本、分享链接、收藏状态与预览质量。'
    : '查看最近访问、版本链、收藏与正式预览状态，直接回到交付链路。',
)
const statItems = computed(() => [
  { label: '产物总数', value: response.value.stats.totalCount },
  { label: '我的收藏', value: response.value.stats.favoriteCount },
  { label: '分享链接', value: response.value.stats.sharedCount },
  { label: '版本链', value: response.value.stats.versionedCount },
  { label: '可内嵌预览', value: response.value.stats.inlinePreviewCount },
  { label: '最近查看', value: response.value.stats.recentViewedCount },
])

async function load() {
  loading.value = true
  error.value = ''
  try {
    response.value =
      scope.value === 'admin'
        ? await api.getAdminArtifactCenter({ q: query.value.trim() || undefined, kind: kind.value || undefined })
        : await api.getPortalArtifactCenter({ q: query.value.trim() || undefined, kind: kind.value || undefined })
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载产物中心失败'
  } finally {
    loading.value = false
  }
}

async function toggleFavorite(item: PortalArtifactCenterItem) {
  favoriteLoadingId.value = item.id
  try {
    await api.favoriteArtifact(item.id, !item.isFavorite, scope.value)
    await load()
    ElMessage.success(item.isFavorite ? '已取消收藏' : '已加入收藏')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '更新收藏状态失败')
  } finally {
    favoriteLoadingId.value = ''
  }
}

function openDetail(item: PortalArtifactCenterItem) {
  router.push(item.detailPath || `/${scope.value}/artifacts/${item.id}`)
}

function openWorkspace(item: PortalArtifactCenterItem) {
  router.push(item.workspacePath)
}

function qualityTagType(item: PortalArtifactCenterItem) {
  switch (item.quality.status) {
    case 'healthy':
      return 'success'
    case 'blocked':
      return 'danger'
    default:
      return 'warning'
  }
}

watch(
  () => scope.value,
  () => {
    void load()
  },
  { immediate: true },
)
</script>

<template>
  <div class="artifact-page">
    <el-card shadow="never" class="hero-card">
      <div class="hero-copy">
        <div class="eyebrow">{{ scope === 'admin' ? 'Admin Artifacts' : 'Portal Artifacts' }}</div>
        <h2>{{ title }}</h2>
        <p class="muted">{{ subtitle }}</p>
      </div>
      <div class="hero-actions">
        <el-input v-model="query" clearable placeholder="搜索标题、实例、会话或来源 URL" @keyup.enter="load" />
        <el-select v-model="kind" placeholder="类型">
          <el-option label="全部类型" value="" />
          <el-option label="Web" value="web" />
          <el-option label="PDF" value="pdf" />
          <el-option label="PPTX" value="pptx" />
          <el-option label="DOCX" value="docx" />
          <el-option label="XLSX" value="xlsx" />
          <el-option label="Image" value="image" />
          <el-option label="Video" value="video" />
          <el-option label="Audio" value="audio" />
        </el-select>
        <el-button type="primary" @click="load">刷新列表</el-button>
      </div>
      <div class="stat-grid">
        <article v-for="stat in statItems" :key="stat.label" class="stat-pill">
          <span class="muted">{{ stat.label }}</span>
          <strong>{{ stat.value }}</strong>
        </article>
      </div>
    </el-card>

    <el-card v-if="loading" shadow="never" class="state-card">正在同步产物中心…</el-card>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
    <template v-else>
      <el-card v-if="response.recentViewed.length" shadow="never" class="panel-card">
        <SectionHeader title="最近查看" subtitle="从最近访问的产物直接回到版本、预览与工作台上下文。" />
        <div class="recent-grid">
          <button
            v-for="item in response.recentViewed"
            :key="item.id"
            type="button"
            class="recent-card"
            @click="openDetail(item)"
          >
            <div class="thumb-shell">
              <img v-if="item.thumbnail.url" :src="item.thumbnail.url" :alt="item.title" class="thumb-image" />
              <div v-else class="thumb-fallback">{{ item.thumbnail.label }}</div>
            </div>
            <div class="recent-copy">
              <strong>{{ item.title }}</strong>
              <span class="muted">v{{ item.version }} / {{ item.latestVersion }} · {{ item.instanceName }}</span>
              <span class="muted">{{ item.quality.lastViewedAt || item.updatedAt }}</span>
            </div>
          </button>
        </div>
      </el-card>

      <el-card v-if="response.stats.failureReasons.length" shadow="never" class="panel-card">
        <SectionHeader title="预览失败画像" subtitle="当前筛选结果里最常见的回退原因，便于优先治理高频问题。" />
        <div class="issue-grid">
          <article v-for="item in response.stats.failureReasons" :key="item.reason" class="issue-card">
            <strong>{{ item.reason }}</strong>
            <span class="muted">{{ item.count }} 个产物受影响</span>
          </article>
        </div>
      </el-card>

      <el-card shadow="never" class="panel-card">
        <SectionHeader title="归档产物" subtitle="支持缩略图、收藏、版本信息、分享统计和正式预览状态。" />
        <div v-if="response.items.length" class="artifact-grid">
          <article v-for="item in response.items" :key="item.id" class="artifact-card">
            <button type="button" class="card-main" @click="openDetail(item)">
              <div class="card-thumb">
                <img v-if="item.thumbnail.url" :src="item.thumbnail.url" :alt="item.title" class="thumb-image" />
                <div v-else class="thumb-fallback">
                  <span>{{ item.thumbnail.label }}</span>
                  <small>{{ item.thumbnail.hint }}</small>
                </div>
              </div>
              <div class="card-copy">
                <div class="card-head">
                  <strong>{{ item.title }}</strong>
                  <el-tag round disable-transitions :type="qualityTagType(item)">
                    {{ item.quality.inlinePreview ? '可预览' : '下载回退' }}
                  </el-tag>
                </div>
                <div class="card-meta muted">
                  <span>{{ item.instanceName }}</span>
                  <span>{{ item.sessionTitle }}</span>
                  <span>{{ item.kind.toUpperCase() }}</span>
                </div>
                <div class="card-meta">
                  <span>v{{ item.version }} / {{ item.latestVersion }}</span>
                  <span>收藏 {{ item.favoriteCount }}</span>
                  <span>分享 {{ item.shareCount }}</span>
                </div>
                <p class="muted card-text">{{ item.messagePreview || item.sourceUrl }}</p>
                <div class="card-foot">
                  <span class="muted">质量分 {{ item.quality.score }}</span>
                  <span class="muted">查看 {{ item.quality.viewCount }} · 下载 {{ item.quality.downloadCount }}</span>
                </div>
              </div>
            </button>
            <div class="card-actions">
              <el-button
                size="small"
                plain
                :loading="favoriteLoadingId === item.id"
                @click.stop="toggleFavorite(item)"
              >
                {{ item.isFavorite ? '取消收藏' : '收藏' }}
              </el-button>
              <el-button size="small" plain @click.stop="openWorkspace(item)">回到工作台</el-button>
              <el-button size="small" type="primary" @click.stop="openDetail(item)">查看详情</el-button>
            </div>
          </article>
        </div>
        <div v-else class="state-card">当前筛选条件下没有匹配的产物。</div>
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.artifact-page {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.hero-card,
.panel-card,
.state-card {
  padding: 18px;
}

.hero-card {
  display: grid;
  gap: 16px;
  background:
    radial-gradient(circle at top left, rgba(29, 107, 255, 0.16), transparent 28%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.98), rgba(239, 245, 255, 0.92));
}

.hero-copy {
  display: grid;
  gap: 8px;
}

.hero-copy h2 {
  margin: 0;
  font-size: 2rem;
}

.hero-actions {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 180px auto;
  gap: 10px;
}

.stat-grid,
.issue-grid,
.artifact-grid,
.recent-grid {
  display: grid;
  gap: 12px;
}

.stat-grid {
  grid-template-columns: repeat(6, minmax(0, 1fr));
}

.stat-pill,
.issue-card {
  display: grid;
  gap: 4px;
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.recent-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.recent-card,
.card-main {
  display: grid;
  gap: 12px;
  border: 0;
  background: transparent;
  padding: 0;
  text-align: left;
  color: inherit;
  cursor: pointer;
}

.recent-card {
  grid-template-columns: 88px minmax(0, 1fr);
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.recent-copy,
.card-copy {
  display: grid;
  gap: 6px;
}

.artifact-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.artifact-card {
  display: grid;
  gap: 12px;
  padding: 16px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--stroke);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(247, 250, 255, 0.95));
}

.card-thumb,
.thumb-shell {
  border-radius: var(--radius-md);
  overflow: hidden;
  min-height: 156px;
  background:
    linear-gradient(135deg, rgba(29, 107, 255, 0.12), rgba(18, 28, 45, 0.06)),
    var(--panel-muted);
}

.thumb-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.thumb-fallback {
  display: grid;
  place-items: center;
  gap: 8px;
  min-height: 156px;
  padding: 16px;
  color: var(--text-muted);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.thumb-fallback small {
  text-transform: none;
  letter-spacing: 0;
  text-align: center;
  font-weight: 500;
}

.card-head,
.card-meta,
.card-foot,
.card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: space-between;
}

.card-text {
  margin: 0;
  line-height: 1.6;
}

.card-actions {
  justify-content: flex-end;
}

.state-card {
  text-align: center;
}

@media (max-width: 1100px) {
  .hero-actions,
  .stat-grid,
  .recent-grid,
  .artifact-grid {
    grid-template-columns: 1fr;
  }
}
</style>
