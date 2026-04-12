<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import { useBranding } from '../../lib/brand'
import type { ArtifactCenterDetail, ArtifactShare, PortalArtifactCenterItem } from '../../lib/types'

const route = useRoute()
const router = useRouter()
const { brand } = useBranding()

const detail = ref<ArtifactCenterDetail | null>(null)
const loading = ref(true)
const favoriteLoading = ref(false)
const shareSubmitting = ref(false)
const error = ref('')
const shareNote = ref('')
const shareDays = ref(7)
const compareVersionId = ref('')

const scope = computed<'portal' | 'admin'>(() => (route.path.startsWith('/admin') ? 'admin' : 'portal'))
const artifact = computed(() => detail.value?.artifact ?? null)
const compareTarget = computed<PortalArtifactCenterItem | null>(() => {
  const versions = detail.value?.versions ?? []
  if (!versions.length) return null
  if (compareVersionId.value) {
    return versions.find((item) => item.id === compareVersionId.value) ?? null
  }
  return versions.find((item) => item.id !== artifact.value?.id) ?? null
})

async function load(id: string) {
  loading.value = true
  error.value = ''
  try {
    detail.value = await api.getArtifactCenterDetail(
      id,
      scope.value,
      typeof route.query.share === 'string' ? route.query.share : undefined,
    )
    const fallbackCompare = detail.value.versions.find((item) => item.id !== detail.value?.artifact.id)
    compareVersionId.value = fallbackCompare?.id ?? ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载产物详情失败'
  } finally {
    loading.value = false
  }
}

function openUrl(target?: string) {
  if (!target) return
  window.open(target, '_blank', 'noopener,noreferrer')
}

function goWorkspace() {
  if (!artifact.value?.workspacePath) return
  router.push(artifact.value.workspacePath)
}

async function toggleFavorite() {
  if (!artifact.value) return
  const wasFavorite = artifact.value.isFavorite
  favoriteLoading.value = true
  try {
    await api.favoriteArtifact(artifact.value.id, !wasFavorite, scope.value)
    await load(artifact.value.id)
    ElMessage.success(wasFavorite ? '已取消收藏' : '已加入收藏')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '更新收藏失败')
  } finally {
    favoriteLoading.value = false
  }
}

async function createShare() {
  if (!artifact.value) return
  shareSubmitting.value = true
  try {
    const share = await api.createArtifactShare(
      artifact.value.id,
      {
        note: shareNote.value.trim() || undefined,
        expiresInDays: shareDays.value,
      },
      scope.value,
    )
    shareNote.value = ''
    await copyShare(share)
    await load(artifact.value.id)
    ElMessage.success('分享链接已创建并复制')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '创建分享失败')
  } finally {
    shareSubmitting.value = false
  }
}

async function revokeShare(share: ArtifactShare) {
  try {
    await api.revokeArtifactShare(share.id, scope.value)
    if (artifact.value) {
      await load(artifact.value.id)
    }
    ElMessage.success('分享链接已撤销')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '撤销分享失败')
  }
}

async function copyShare(share: ArtifactShare) {
  const url = share.shareUrl.startsWith('http') ? share.shareUrl : new URL(share.shareUrl, window.location.origin).toString()
  await navigator.clipboard.writeText(url)
}

function previewMode() {
  if (!detail.value?.preview.available || !detail.value.preview.previewUrl) return 'fallback'
  if (detail.value.preview.mode === 'image') return 'image'
  if (detail.value.preview.mode === 'video') return 'video'
  if (detail.value.preview.mode === 'audio') return 'audio'
  return 'frame'
}

function qualityTagType() {
  switch (artifact.value?.quality.status) {
    case 'healthy':
      return 'success'
    case 'blocked':
      return 'danger'
    default:
      return 'warning'
  }
}

watch(
  () => [String(route.params.id), scope.value, String(route.query.share ?? '')] as const,
  ([id]) => {
    void load(id)
  },
  { immediate: true },
)
</script>

<template>
  <div class="artifact-detail">
    <el-card v-if="loading" shadow="never" class="state-card">正在加载产物详情…</el-card>
    <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
    <template v-else-if="detail && artifact">
      <el-card shadow="never" class="hero-card">
        <div class="hero-copy">
          <div class="eyebrow">{{ brand.name }} · {{ scope === 'admin' ? 'Admin Artifact' : 'Portal Artifact' }}</div>
          <h2>{{ artifact.title }}</h2>
          <p class="muted">
            {{ artifact.instanceName }} · {{ artifact.sessionTitle }} · {{ artifact.kind.toUpperCase() }} ·
            v{{ artifact.version }} / {{ artifact.latestVersion }}
          </p>
          <div class="hero-tags">
            <el-tag round disable-transitions :type="qualityTagType()">质量分 {{ artifact.quality.score }}</el-tag>
            <el-tag round disable-transitions>{{ artifact.archiveStatus || 'external' }}</el-tag>
            <el-tag round disable-transitions>收藏 {{ artifact.favoriteCount }}</el-tag>
            <el-tag round disable-transitions>分享 {{ artifact.shareCount }}</el-tag>
          </div>
        </div>
        <div class="hero-actions">
          <el-button v-if="detail.preview.previewUrl" type="primary" @click="openUrl(detail.preview.previewUrl)">打开正式预览</el-button>
          <el-button v-if="detail.preview.downloadUrl" plain @click="openUrl(detail.preview.downloadUrl)">下载源文件</el-button>
          <el-button plain @click="goWorkspace">回到工作台</el-button>
          <el-button plain :loading="favoriteLoading" @click="toggleFavorite">
            {{ artifact.isFavorite ? '取消收藏' : '加入收藏' }}
          </el-button>
          <el-button plain @click="router.back()">返回上一页</el-button>
        </div>
      </el-card>

      <section class="detail-grid">
        <el-card shadow="never" class="panel-card preview-card">
          <SectionHeader :title="`${brand.name} 正式预览`" subtitle="优先走平台预览策略，不可内嵌时回退到下载或外部地址。" />
          <div class="preview-meta">
            <span class="muted">策略 {{ detail.preview.strategy }}</span>
            <span class="muted">模式 {{ detail.preview.mode }}</span>
            <span class="muted">查看 {{ artifact.quality.viewCount }} · 下载 {{ artifact.quality.downloadCount }}</span>
          </div>
          <div v-if="previewMode() === 'frame'" class="preview-shell">
            <iframe
              :src="detail.preview.previewUrl"
              class="preview-frame"
              title="Artifact Preview"
              sandbox="allow-forms allow-scripts allow-modals allow-popups allow-downloads"
            />
          </div>
          <img
            v-else-if="previewMode() === 'image' && detail.preview.previewUrl"
            :src="detail.preview.previewUrl"
            :alt="artifact.title"
            class="preview-image"
          />
          <video
            v-else-if="previewMode() === 'video' && detail.preview.previewUrl"
            :src="detail.preview.previewUrl"
            class="preview-media"
            controls
            playsinline
          />
          <audio
            v-else-if="previewMode() === 'audio' && detail.preview.previewUrl"
            :src="detail.preview.previewUrl"
            class="preview-audio"
            controls
          />
          <div v-else class="fallback-card">
            <p class="muted">{{ artifact.quality.failureReason || detail.preview.failureReason || '当前产物没有可内嵌的正式预览。' }}</p>
            <div class="fallback-actions">
              <el-button v-if="detail.preview.downloadUrl" type="primary" @click="openUrl(detail.preview.downloadUrl)">下载</el-button>
              <el-button v-if="detail.preview.externalUrl" plain @click="openUrl(detail.preview.externalUrl)">打开源地址</el-button>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="panel-card">
          <SectionHeader title="产物质量" subtitle="预览成功率、最后访问时间和回退原因一屏可见。" />
          <div class="meta-grid">
            <div class="meta-item">
              <span class="muted">预览状态</span>
              <strong>{{ artifact.quality.inlinePreview ? '可内嵌预览' : '下载回退' }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">最后查看</span>
              <strong>{{ artifact.quality.lastViewedAt || '—' }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">最后下载</span>
              <strong>{{ artifact.quality.lastDownloadedAt || '—' }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">文件大小</span>
              <strong>{{ artifact.sizeLabel || '—' }}</strong>
            </div>
          </div>
          <div class="stack-item">
            <span class="muted">失败画像</span>
            <strong>{{ artifact.quality.failureReason || '当前产物预览链路正常。' }}</strong>
          </div>
          <div class="stack-item">
            <span class="muted">来源地址</span>
            <strong>{{ artifact.sourceUrl }}</strong>
          </div>
        </el-card>

        <el-card shadow="never" class="panel-card">
          <SectionHeader title="版本管理" subtitle="查看当前版本链，并和旧版本对比交付差异。" />
          <div class="version-toolbar">
            <span class="muted">当前版本 v{{ artifact.version }}</span>
            <el-select v-if="detail.versions.length > 1" v-model="compareVersionId" placeholder="选择对比版本">
              <el-option
                v-for="item in detail.versions.filter((item) => item.id !== artifact?.id)"
                :key="item.id"
                :label="`v${item.version} · ${item.updatedAt}`"
                :value="item.id"
              />
            </el-select>
          </div>
          <div class="version-list">
            <button
              v-for="item in detail.versions"
              :key="item.id"
              type="button"
              :class="['version-item', { active: item.id === artifact.id }]"
              @click="router.push(item.detailPath || `/${scope}/artifacts/${item.id}`)"
            >
              <strong>v{{ item.version }}</strong>
              <span class="muted">{{ item.updatedAt }}</span>
              <span class="muted">{{ item.sizeLabel || '未知大小' }}</span>
            </button>
          </div>
          <div v-if="compareTarget" class="compare-grid">
            <div class="meta-item">
              <span class="muted">对比版本</span>
              <strong>v{{ compareTarget.version }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">文件名变化</span>
              <strong>{{ compareTarget.filename === artifact.filename ? '无变化' : '已变化' }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">大小对比</span>
              <strong>{{ compareTarget.sizeLabel || '—' }} → {{ artifact.sizeLabel || '—' }}</strong>
            </div>
            <div class="meta-item">
              <span class="muted">预览策略</span>
              <strong>{{ compareTarget.preview.strategy }} → {{ artifact.preview.strategy }}</strong>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="panel-card">
          <SectionHeader title="分享链接" subtitle="生成带有效期的内部分享链接，可按需撤销。" />
          <div class="share-form">
            <el-input v-model="shareNote" maxlength="80" placeholder="补充分享备注，例如：交付给市场团队" />
            <el-select v-model="shareDays" placeholder="有效期">
              <el-option :value="3" label="3 天" />
              <el-option :value="7" label="7 天" />
              <el-option :value="14" label="14 天" />
              <el-option :value="30" label="30 天" />
            </el-select>
            <el-button type="primary" :loading="shareSubmitting" @click="createShare">创建分享</el-button>
          </div>
          <div v-if="detail.shares.length" class="share-list">
            <article v-for="share in detail.shares" :key="share.id" class="share-item">
              <div class="share-head">
                <strong>{{ share.note || share.shareUrl }}</strong>
                <el-tag round disable-transitions :type="share.active ? 'success' : 'info'">
                  {{ share.active ? '有效' : '已失效' }}
                </el-tag>
              </div>
              <div class="muted share-meta">
                <span>{{ share.createdBy }}</span>
                <span>{{ share.createdAt }}</span>
                <span>打开 {{ share.useCount }} 次</span>
                <span>到期 {{ share.expiresAt || '长期' }}</span>
              </div>
              <div class="share-actions">
                <el-button plain @click="copyShare(share)">复制链接</el-button>
                <el-button plain @click="openUrl(share.shareUrl)">打开分享页</el-button>
                <el-button v-if="share.active" type="danger" plain @click="revokeShare(share)">撤销</el-button>
              </div>
            </article>
          </div>
          <div v-else class="fallback-card">当前还没有分享链接。</div>
        </el-card>

        <el-card shadow="never" class="panel-card">
          <SectionHeader title="会话上下文" subtitle="保留交付来源，方便回到会话和关联消息。" />
          <div class="stack-item">
            <span class="muted">会话</span>
            <strong>{{ artifact.sessionTitle }}</strong>
            <small class="muted">#{{ artifact.sessionNo }} · {{ artifact.instanceName }}</small>
          </div>
          <div v-if="detail.message" class="stack-item">
            <span class="muted">关联消息</span>
            <strong>{{ detail.message.role }} · {{ detail.message.status }}</strong>
            <small class="muted">{{ detail.message.content }}</small>
          </div>
          <div class="context-actions">
            <el-button plain @click="goWorkspace">回到工作台</el-button>
            <el-button v-if="detail.preview.externalUrl" plain @click="openUrl(detail.preview.externalUrl)">打开源地址</el-button>
          </div>
        </el-card>

        <el-card shadow="never" class="panel-card span-two">
          <SectionHeader title="访问日志" subtitle="最近的详情查看、预览和下载记录。" />
          <div v-if="detail.accessLogs.length" class="log-list">
            <article v-for="log in detail.accessLogs" :key="log.id" class="log-item">
              <strong>{{ log.action }} · {{ log.scope }}</strong>
              <span class="muted">{{ log.actor }} · {{ log.createdAt }}</span>
              <small class="muted">{{ log.remoteAddr || '—' }} · {{ log.userAgent || '—' }}</small>
            </article>
          </div>
          <div v-else class="fallback-card">当前还没有访问日志。</div>
        </el-card>
      </section>
    </template>
  </div>
</template>

<style scoped>
.artifact-detail {
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
    linear-gradient(145deg, rgba(255, 255, 255, 0.98), rgba(241, 246, 255, 0.94));
}

.hero-copy,
.hero-tags,
.hero-actions,
.preview-meta,
.share-meta,
.share-actions,
.context-actions,
.version-toolbar,
.version-list,
.log-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-copy {
  display: grid;
  gap: 8px;
}

.hero-copy h2 {
  margin: 0;
  font-size: 2rem;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 14px;
}

.span-two {
  grid-column: span 2;
}

.preview-card {
  min-height: 540px;
}

.preview-shell,
.preview-frame,
.preview-image,
.preview-media {
  width: 100%;
  min-height: 420px;
  border: 0;
  border-radius: var(--radius-md);
}

.preview-frame {
  background: #fff;
}

.preview-image,
.preview-media {
  object-fit: contain;
  background: var(--panel-muted);
}

.preview-audio {
  width: 100%;
}

.meta-grid,
.compare-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.meta-item,
.stack-item,
.share-item,
.log-item,
.version-item,
.fallback-card {
  display: grid;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.version-item {
  border: 0;
  text-align: left;
  color: inherit;
  cursor: pointer;
  min-width: 120px;
}

.version-item.active {
  outline: 1px solid rgba(29, 107, 255, 0.2);
  background: rgba(29, 107, 255, 0.08);
}

.share-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 160px auto;
  gap: 10px;
}

.share-list,
.log-list {
  display: grid;
  gap: 10px;
}

.state-card {
  text-align: center;
}

@media (max-width: 1100px) {
  .detail-grid,
  .meta-grid,
  .compare-grid,
  .share-form {
    grid-template-columns: 1fr;
  }

  .span-two {
    grid-column: span 1;
  }
}
</style>
