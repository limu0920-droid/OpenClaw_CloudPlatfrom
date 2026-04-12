export type ArtifactKind =
  | 'web'
  | 'pdf'
  | 'pptx'
  | 'docx'
  | 'xlsx'
  | 'image'
  | 'video'
  | 'audio'
  | 'text'
  | 'unknown'

export interface ArtifactDraftPreview {
  mode: 'html' | 'pdf' | 'image' | 'video' | 'audio' | 'text' | 'download'
  url?: string
  sandboxed?: boolean
  reason?: string
  note?: string
}

export function detectArtifactKind(rawUrl: string): ArtifactKind {
  const url = rawUrl.trim().toLowerCase()

  if (!url) return 'unknown'
  if (/\.(ppt|pptx)([?#].*)?$/.test(url)) return 'pptx'
  if (/\.(doc|docx)([?#].*)?$/.test(url)) return 'docx'
  if (/\.(xls|xlsx|csv|tsv)([?#].*)?$/.test(url)) return 'xlsx'
  if (/\.(pdf)([?#].*)?$/.test(url)) return 'pdf'
  if (/\.(png|jpg|jpeg|gif|webp|svg)([?#].*)?$/.test(url)) return 'image'
  if (/\.(mp4|webm|mov|m3u8)([?#].*)?$/.test(url)) return 'video'
  if (/\.(mp3|wav|ogg|m4a)([?#].*)?$/.test(url)) return 'audio'
  if (/\.(md|txt|json|xml|yaml|yml)([?#].*)?$/.test(url)) return 'text'
  if (/\.(html|htm)([?#].*)?$/.test(url)) return 'web'
  if (/^https?:\/\//.test(url)) return 'web'
  return 'unknown'
}

export function getArtifactLabel(kind: ArtifactKind) {
  switch (kind) {
    case 'web':
      return '网页'
    case 'pdf':
      return 'PDF'
    case 'pptx':
      return 'PPTX'
    case 'docx':
      return '文档'
    case 'xlsx':
      return '表格'
    case 'image':
      return '图片'
    case 'video':
      return '视频'
    case 'audio':
      return '音频'
    case 'text':
      return '文本'
    default:
      return '文件'
  }
}

export function getDraftArtifactPreview(sourceUrl: string, kind: ArtifactKind, previewUrl = ''): ArtifactDraftPreview {
  const source = sourceUrl.trim()
  const preview = previewUrl.trim()
  const target = preview || source

  if (!source) {
    return {
      mode: 'download',
      reason: '请先输入产物地址。',
    }
  }

  switch (kind) {
    case 'web':
      return {
        mode: 'html',
        url: target,
        sandboxed: true,
        note: '这是临时 URL 检查；正式上线预览以保存后的平台网关结果为准。',
      }
    case 'pdf':
      return {
        mode: 'pdf',
        url: target,
      }
    case 'pptx':
    case 'docx':
    case 'xlsx':
      switch (detectArtifactKind(preview)) {
        case 'pdf':
          return {
            mode: 'pdf',
            url: preview,
            note: '当前会按 PDF 衍生文件做临时预览；正式策略同样支持保存后的 PDF 预览地址。',
          }
        case 'web':
          return {
            mode: 'html',
            url: preview,
            sandboxed: true,
            note: '当前会按 HTML 衍生文件做临时预览；正式策略同样支持保存后的 HTML 沙箱预览。',
          }
      }
      return {
        mode: 'download',
        reason: 'Office 正式预览策略已切换为“预生成 PDF/HTML 或下载回退”，不再依赖第三方在线 Office 预览。',
      }
    case 'image':
      return {
        mode: 'image',
        url: target,
      }
    case 'video':
      return {
        mode: 'video',
        url: target,
      }
    case 'audio':
      return {
        mode: 'audio',
        url: target,
      }
    case 'text':
      return {
        mode: 'text',
        url: target,
      }
    default:
      return {
        mode: 'download',
        reason: '当前链接不支持平台内嵌预览，请保存后通过正式预览策略检查，或直接新窗口打开。',
      }
  }
}
