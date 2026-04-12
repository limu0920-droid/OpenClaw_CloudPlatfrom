package httpapi

import (
	"context"
	"fmt"
	"html"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
)

const (
	defaultArtifactPreviewTimeout  = 20 * time.Second
	defaultArtifactPreviewHTMLSize = 2 * 1024 * 1024
)

var (
	artifactPreviewHeadPattern   = regexp.MustCompile(`(?i)<head[^>]*>`)
	artifactPreviewMetaCSPRegexp = regexp.MustCompile(`(?is)<meta[^>]+http-equiv=["']?content-security-policy["']?[^>]*>`)
)

type workspaceArtifactPreviewDescriptor struct {
	Available     bool   `json:"available"`
	Mode          string `json:"mode"`
	Strategy      string `json:"strategy"`
	Sandboxed     bool   `json:"sandboxed"`
	Proxied       bool   `json:"proxied"`
	PreviewURL    string `json:"previewUrl,omitempty"`
	DownloadURL   string `json:"downloadUrl,omitempty"`
	ExternalURL   string `json:"externalUrl,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
	Note          string `json:"note,omitempty"`
}

type artifactPreviewPlan struct {
	mode            string
	strategy        string
	sandboxed       bool
	proxied         bool
	inlineTargetURL string
	downloadURL     string
	failureReason   string
	note            string
}

type workspaceArtifactValidationError struct {
	status  int
	message string
}

func (e workspaceArtifactValidationError) Error() string {
	return e.message
}

func (r *Router) handlePortalWorkspaceArtifactPreview(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/artifacts/", "/preview")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.getWorkspaceArtifactPreview(w, req, artifactID, "portal")
}

func (r *Router) handleAdminWorkspaceArtifactPreview(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/artifacts/", "/preview")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.getWorkspaceArtifactPreview(w, req, artifactID, "admin")
}

func (r *Router) handlePortalWorkspaceArtifactPreviewContent(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/artifacts/", "/preview-content")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.serveWorkspaceArtifactPreviewContent(w, req, artifactID, "portal")
}

func (r *Router) handleAdminWorkspaceArtifactPreviewContent(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/artifacts/", "/preview-content")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.serveWorkspaceArtifactPreviewContent(w, req, artifactID, "admin")
}

func (r *Router) handlePortalWorkspaceArtifactContent(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/artifacts/", "/content")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.serveWorkspaceArtifactContent(w, req, artifactID, "portal")
}

func (r *Router) handleAdminWorkspaceArtifactContent(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/artifacts/", "/content")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.serveWorkspaceArtifactContent(w, req, artifactID, "admin")
}

func (r *Router) handlePortalWorkspaceArtifactDownload(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/portal/workspace/artifacts/", "/download")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.serveWorkspaceArtifactDownload(w, req, artifactID, "portal")
}

func (r *Router) handleAdminWorkspaceArtifactDownload(w http.ResponseWriter, req *http.Request) {
	artifactID, ok := parseTailID(req.URL.Path, "/api/v1/admin/workspace/artifacts/", "/download")
	if !ok {
		http.Error(w, "invalid artifact id", http.StatusBadRequest)
		return
	}
	r.serveWorkspaceArtifactDownload(w, req, artifactID, "admin")
}

func (r *Router) getWorkspaceArtifactPreview(w http.ResponseWriter, req *http.Request, artifactID int, scope string) {
	actor, status, message := r.resolveWorkspaceActor(req, scope)
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	artifact, session, instance, found, allowed := r.findWorkspaceArtifactContext(actor, artifactID, scope)
	if !found {
		http.NotFound(w, req)
		return
	}
	if !allowed {
		_ = r.recordWorkspaceAudit(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.preview.describe", "denied", map[string]string{
			"scope":      scope,
			"artifactId": strconv.Itoa(artifact.ID),
			"sessionId":  strconv.Itoa(session.ID),
		})
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"artifact": artifact,
		"preview":  r.buildWorkspaceArtifactPreviewDescriptor(req, artifact, session, instance, scope),
	})
}

func (r *Router) serveWorkspaceArtifactContent(w http.ResponseWriter, req *http.Request, artifactID int, scope string) {
	actor, status, message := r.resolveWorkspaceActor(req, scope)
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	artifact, session, instance, found, allowed := r.findWorkspaceArtifactContext(actor, artifactID, scope)
	if !found {
		http.NotFound(w, req)
		return
	}
	if !allowed {
		_ = r.recordWorkspaceAudit(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.content", "denied", map[string]string{
			"scope":      scope,
			"artifactId": strconv.Itoa(artifact.ID),
			"sessionId":  strconv.Itoa(session.ID),
		})
		http.NotFound(w, req)
		return
	}

	disposition := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("disposition")))
	inline := disposition != "attachment"
	action := "view"
	mode := normalizeArtifactKind(artifact.Kind)
	if !inline {
		action = "download"
		mode = "download"
	}

	if strings.TrimSpace(artifact.StorageKey) != "" && strings.EqualFold(artifact.ArchiveStatus, "archived") {
		if err := r.recordWorkspaceArtifactAccess(actor, artifact, action, scope, req); err != nil {
			http.Error(w, "persist workspace artifact access failed", http.StatusInternalServerError)
			return
		}
		r.serveArchivedWorkspaceArtifact(w, req, artifact, inline, mode)
		return
	}

	targetURL, err := pickWorkspaceArtifactURL(artifact, req.URL.Query().Get("target"))
	if err != nil || strings.TrimSpace(targetURL) == "" {
		http.Error(w, "artifact content is not available", http.StatusConflict)
		return
	}
	if !r.isWorkspaceArtifactURLTrusted(session, instance, targetURL) {
		http.Error(w, "artifact content url is not trusted", http.StatusForbidden)
		return
	}
	if err := r.recordWorkspaceArtifactAccess(actor, artifact, action, scope, req); err != nil {
		http.Error(w, "persist workspace artifact access failed", http.StatusInternalServerError)
		return
	}
	r.proxyWorkspaceArtifactURL(w, req, artifact, targetURL, inline, mode)
}

func (r *Router) serveWorkspaceArtifactPreviewContent(w http.ResponseWriter, req *http.Request, artifactID int, scope string) {
	actor, status, message := r.resolveWorkspaceActor(req, scope)
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	artifact, session, instance, found, allowed := r.findWorkspaceArtifactContext(actor, artifactID, scope)
	if !found {
		http.NotFound(w, req)
		return
	}
	if !allowed {
		_ = r.recordWorkspaceAudit(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.preview", "denied", map[string]string{
			"scope":      scope,
			"artifactId": strconv.Itoa(artifact.ID),
			"sessionId":  strconv.Itoa(session.ID),
		})
		http.NotFound(w, req)
		return
	}

	plan := r.planWorkspaceArtifactPreview(artifact)
	if plan.mode == "download" || strings.TrimSpace(plan.inlineTargetURL) == "" {
		http.Error(w, defaultString(plan.failureReason, "artifact preview is not available"), http.StatusConflict)
		return
	}
	if strings.TrimSpace(artifact.StorageKey) != "" && strings.EqualFold(artifact.ArchiveStatus, "archived") {
		if err := r.recordWorkspaceArtifactAccess(actor, artifact, "view", scope, req); err != nil {
			http.Error(w, "persist workspace artifact access failed", http.StatusInternalServerError)
			return
		}
		r.serveArchivedWorkspaceArtifact(w, req, artifact, true, plan.mode)
		return
	}
	if !r.isWorkspaceArtifactURLTrusted(session, instance, plan.inlineTargetURL) {
		http.Error(w, "artifact preview url is not trusted", http.StatusForbidden)
		return
	}
	if err := r.recordWorkspaceArtifactAccess(actor, artifact, "view", scope, req); err != nil {
		http.Error(w, "persist workspace artifact access failed", http.StatusInternalServerError)
		return
	}

	r.proxyWorkspaceArtifactURL(w, req, artifact, plan.inlineTargetURL, true, plan.mode)
}

func (r *Router) serveWorkspaceArtifactDownload(w http.ResponseWriter, req *http.Request, artifactID int, scope string) {
	actor, status, message := r.resolveWorkspaceActor(req, scope)
	if status != 0 {
		http.Error(w, message, status)
		return
	}

	artifact, session, instance, found, allowed := r.findWorkspaceArtifactContext(actor, artifactID, scope)
	if !found {
		http.NotFound(w, req)
		return
	}
	if !allowed {
		_ = r.recordWorkspaceAudit(actor, artifact.TenantID, artifact.InstanceID, "workspace.artifact.download", "denied", map[string]string{
			"scope":      scope,
			"artifactId": strconv.Itoa(artifact.ID),
			"sessionId":  strconv.Itoa(session.ID),
		})
		http.NotFound(w, req)
		return
	}

	plan := r.planWorkspaceArtifactPreview(artifact)
	if strings.TrimSpace(plan.downloadURL) == "" {
		http.Error(w, "artifact source url is empty", http.StatusConflict)
		return
	}
	if strings.TrimSpace(artifact.StorageKey) != "" && strings.EqualFold(artifact.ArchiveStatus, "archived") {
		if err := r.recordWorkspaceArtifactAccess(actor, artifact, "download", scope, req); err != nil {
			http.Error(w, "persist workspace artifact access failed", http.StatusInternalServerError)
			return
		}
		r.serveArchivedWorkspaceArtifact(w, req, artifact, false, "download")
		return
	}
	if !r.isWorkspaceArtifactURLTrusted(session, instance, plan.downloadURL) {
		http.Error(w, "artifact download url is not trusted", http.StatusForbidden)
		return
	}
	if err := r.recordWorkspaceArtifactAccess(actor, artifact, "download", scope, req); err != nil {
		http.Error(w, "persist workspace artifact access failed", http.StatusInternalServerError)
		return
	}

	r.proxyWorkspaceArtifactURL(w, req, artifact, plan.downloadURL, false, "download")
}

func (r *Router) buildWorkspaceArtifactPreviewDescriptor(
	req *http.Request,
	artifact models.WorkspaceArtifact,
	session models.WorkspaceSession,
	instance models.Instance,
	scope string,
) workspaceArtifactPreviewDescriptor {
	plan := r.planWorkspaceArtifactPreview(artifact)
	basePath := fmt.Sprintf("/api/v1/%s/workspace/artifacts/%d", scope, artifact.ID)
	descriptor := workspaceArtifactPreviewDescriptor{
		Available:     plan.mode != "download" && strings.TrimSpace(plan.inlineTargetURL) != "",
		Mode:          plan.mode,
		Strategy:      plan.strategy,
		Sandboxed:     plan.sandboxed,
		Proxied:       defaultBool(plan.proxied, true),
		ExternalURL:   strings.TrimSpace(artifact.SourceURL),
		FailureReason: strings.TrimSpace(plan.failureReason),
		Note:          strings.TrimSpace(plan.note),
	}
	archived := strings.TrimSpace(artifact.StorageKey) != "" && strings.EqualFold(artifact.ArchiveStatus, "archived")

	if strings.TrimSpace(plan.downloadURL) != "" {
		if archived || r.isWorkspaceArtifactURLTrusted(session, instance, plan.downloadURL) {
			descriptor.DownloadURL = r.workspaceArtifactPublicURL(req, basePath+"/download")
		} else if descriptor.Note == "" {
			descriptor.Note = "平台下载代理已拒绝不受信任的产物地址。"
		}
	}

	if descriptor.Available {
		if archived || r.isWorkspaceArtifactURLTrusted(session, instance, plan.inlineTargetURL) {
			descriptor.PreviewURL = r.workspaceArtifactPublicURL(req, basePath+"/preview-content")
		} else {
			descriptor.Available = false
			descriptor.Mode = "download"
			descriptor.FailureReason = "预览网关拒绝不受信任的产物地址。"
		}
	}

	if descriptor.Mode == "" {
		descriptor.Mode = "download"
	}

	return descriptor
}

func (r *Router) planWorkspaceArtifactPreview(artifact models.WorkspaceArtifact) artifactPreviewPlan {
	sourceURL := strings.TrimSpace(artifact.SourceURL)
	previewURL := strings.TrimSpace(artifact.PreviewURL)
	kind := normalizeArtifactKind(artifact.Kind)
	if kind == "unknown" {
		kind = detectArtifactKindFromURL(sourceURL)
	}

	plan := artifactPreviewPlan{
		mode:        "download",
		strategy:    "download-proxy",
		proxied:     true,
		downloadURL: sourceURL,
	}

	switch kind {
	case "web":
		target := firstNonEmpty(previewURL, sourceURL)
		plan.mode = "html"
		plan.strategy = "html-sandbox-proxy"
		plan.sandboxed = true
		plan.inlineTargetURL = target
		plan.note = "HTML 产物通过平台预览网关代理，并在隔离沙箱中加载。"
	case "pdf":
		target := firstNonEmpty(previewURL, sourceURL)
		plan.mode = "pdf"
		plan.strategy = "pdf-inline-proxy"
		plan.inlineTargetURL = target
		plan.note = "PDF 通过平台代理直出，避免在前端暴露私有产物地址。"
	case "pptx", "docx", "xlsx":
		switch detectArtifactKindFromURL(previewURL) {
		case "pdf":
			plan.mode = "pdf"
			plan.strategy = "office-pdf-rendition"
			plan.inlineTargetURL = previewURL
			plan.note = "Office 正式策略优先使用实例侧生成的 PDF 预览衍生物。"
		case "web":
			plan.mode = "html"
			plan.strategy = "office-html-rendition"
			plan.sandboxed = true
			plan.inlineTargetURL = previewURL
			plan.note = "Office 正式策略支持实例侧生成的 HTML 预览衍生物。"
		case "image":
			plan.mode = "image"
			plan.strategy = "office-image-rendition"
			plan.inlineTargetURL = previewURL
			plan.note = "Office 正式策略支持实例侧生成的图片预览衍生物。"
		case "text":
			plan.mode = "text"
			plan.strategy = "office-text-rendition"
			plan.inlineTargetURL = previewURL
			plan.note = "Office 正式策略支持实例侧生成的文本预览衍生物。"
		default:
			plan.mode = "download"
			plan.strategy = "office-download-fallback"
			plan.note = "Office 正式策略为优先使用衍生预览，缺失时回退为平台受控下载。"
			plan.failureReason = "当前 Office 产物未提供可渲染的 PDF/HTML 预览衍生物，平台按正式策略回退为下载。"
		}
	case "image":
		target := firstNonEmpty(previewURL, sourceURL)
		plan.mode = "image"
		plan.strategy = "image-inline-proxy"
		plan.inlineTargetURL = target
	case "video":
		target := firstNonEmpty(previewURL, sourceURL)
		plan.mode = "video"
		plan.strategy = "video-inline-proxy"
		plan.inlineTargetURL = target
	case "audio":
		target := firstNonEmpty(previewURL, sourceURL)
		plan.mode = "audio"
		plan.strategy = "audio-inline-proxy"
		plan.inlineTargetURL = target
	case "text":
		target := firstNonEmpty(previewURL, sourceURL)
		plan.mode = "text"
		plan.strategy = "text-inline-proxy"
		plan.inlineTargetURL = target
	default:
		switch detectArtifactKindFromURL(firstNonEmpty(previewURL, sourceURL)) {
		case "pdf":
			plan.mode = "pdf"
			plan.strategy = "pdf-inline-proxy"
			plan.inlineTargetURL = firstNonEmpty(previewURL, sourceURL)
		case "image":
			plan.mode = "image"
			plan.strategy = "image-inline-proxy"
			plan.inlineTargetURL = firstNonEmpty(previewURL, sourceURL)
		case "video":
			plan.mode = "video"
			plan.strategy = "video-inline-proxy"
			plan.inlineTargetURL = firstNonEmpty(previewURL, sourceURL)
		case "audio":
			plan.mode = "audio"
			plan.strategy = "audio-inline-proxy"
			plan.inlineTargetURL = firstNonEmpty(previewURL, sourceURL)
		case "text":
			plan.mode = "text"
			plan.strategy = "text-inline-proxy"
			plan.inlineTargetURL = firstNonEmpty(previewURL, sourceURL)
		case "web":
			plan.mode = "html"
			plan.strategy = "html-sandbox-proxy"
			plan.sandboxed = true
			plan.inlineTargetURL = firstNonEmpty(previewURL, sourceURL)
			plan.note = "HTML 产物通过平台预览网关代理，并在隔离沙箱中加载。"
		default:
			plan.failureReason = "当前产物不支持平台内嵌预览，请下载或在新窗口打开。"
		}
	}

	if strings.TrimSpace(plan.downloadURL) == "" {
		plan.failureReason = defaultString(plan.failureReason, "产物源地址为空，无法预览或下载。")
	}

	return plan
}

func normalizeArtifactKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "web", "html":
		return "web"
	case "pdf":
		return "pdf"
	case "ppt", "pptx":
		return "pptx"
	case "doc", "docx":
		return "docx"
	case "xls", "xlsx", "csv", "tsv":
		return "xlsx"
	case "image":
		return "image"
	case "video":
		return "video"
	case "audio":
		return "audio"
	case "text", "md", "markdown", "txt", "json", "xml", "yaml", "yml":
		return "text"
	default:
		return "unknown"
	}
}

func detectArtifactKindFromURL(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if lower == "" {
		return "unknown"
	}
	if strings.HasSuffix(lower, ".ppt") || strings.Contains(lower, ".ppt?") || strings.Contains(lower, ".pptx") {
		return "pptx"
	}
	if strings.HasSuffix(lower, ".doc") || strings.Contains(lower, ".doc?") || strings.Contains(lower, ".docx") {
		return "docx"
	}
	if strings.Contains(lower, ".xls") || strings.Contains(lower, ".xlsx") || strings.Contains(lower, ".csv") || strings.Contains(lower, ".tsv") {
		return "xlsx"
	}
	if strings.Contains(lower, ".pdf") {
		return "pdf"
	}
	if strings.Contains(lower, ".png") || strings.Contains(lower, ".jpg") || strings.Contains(lower, ".jpeg") || strings.Contains(lower, ".gif") || strings.Contains(lower, ".webp") || strings.Contains(lower, ".svg") {
		return "image"
	}
	if strings.Contains(lower, ".mp4") || strings.Contains(lower, ".webm") || strings.Contains(lower, ".mov") || strings.Contains(lower, ".m3u8") {
		return "video"
	}
	if strings.Contains(lower, ".mp3") || strings.Contains(lower, ".wav") || strings.Contains(lower, ".ogg") || strings.Contains(lower, ".m4a") {
		return "audio"
	}
	if strings.Contains(lower, ".md") || strings.Contains(lower, ".txt") || strings.Contains(lower, ".json") || strings.Contains(lower, ".xml") || strings.Contains(lower, ".yaml") || strings.Contains(lower, ".yml") {
		return "text"
	}
	if strings.Contains(lower, ".html") || strings.Contains(lower, ".htm") {
		return "web"
	}
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return "web"
	}
	return "unknown"
}

func detectWorkspaceArtifactKind(raw string) string {
	return normalizeArtifactKind(detectArtifactKindFromURL(raw))
}

func defaultBool(value bool, fallback bool) bool {
	if value {
		return true
	}
	return fallback
}

func (r *Router) findWorkspaceArtifactContext(actor workspaceActor, id int, scope string) (models.WorkspaceArtifact, models.WorkspaceSession, models.Instance, bool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	artifact, ok := r.findWorkspaceArtifact(id)
	if !ok {
		return models.WorkspaceArtifact{}, models.WorkspaceSession{}, models.Instance{}, false, false
	}
	session, ok := r.findWorkspaceSession(artifact.SessionID)
	if !ok {
		return models.WorkspaceArtifact{}, models.WorkspaceSession{}, models.Instance{}, false, false
	}
	instance, found := r.findInstance(session.InstanceID)
	if !found {
		return models.WorkspaceArtifact{}, models.WorkspaceSession{}, models.Instance{}, false, false
	}
	if !r.canAccessWorkspaceArtifact(actor, artifact, session, instance, scope) {
		return artifact, session, instance, true, false
	}
	return artifact, session, instance, true, true
}

func (r *Router) workspaceArtifactPublicURL(req *http.Request, path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(r.config.ArtifactPreviewPublicBaseURL), "/")
	if baseURL == "" {
		return path
	}
	return baseURL + path
}

func (r *Router) artifactPreviewHTTPClient() *http.Client {
	if r.config.HTTPClient != nil {
		return r.config.HTTPClient
	}

	timeout := defaultArtifactPreviewTimeout
	if r.config.ArtifactPreviewTimeoutSecs > 0 {
		timeout = time.Duration(r.config.ArtifactPreviewTimeoutSecs) * time.Second
	}

	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return r.validateArtifactTargetURL(context.Background(), req.URL.String())
		},
	}
}

func (r *Router) artifactPreviewHTMLLimit() int64 {
	if r.config.ArtifactPreviewHTMLMaxBytes > 0 {
		return int64(r.config.ArtifactPreviewHTMLMaxBytes)
	}
	return defaultArtifactPreviewHTMLSize
}

func (r *Router) normalizeAndValidateWorkspaceArtifactURLs(
	session models.WorkspaceSession,
	instance models.Instance,
	sourceURL string,
	previewURL string,
) (string, string, error) {
	return r.normalizeAndValidateWorkspaceArtifactURLsWithAccess(
		session,
		r.workspaceArtifactAccessEntries(instance.ID),
		sourceURL,
		previewURL,
	)
}

func (r *Router) normalizeAndValidateWorkspaceArtifactURLsWithAccess(
	session models.WorkspaceSession,
	accessEntries []models.InstanceAccess,
	sourceURL string,
	previewURL string,
) (string, string, error) {
	resolvedSourceURL, err := r.normalizeWorkspaceArtifactURL(session, sourceURL)
	if err != nil {
		return "", "", workspaceArtifactValidationError{
			status:  http.StatusBadRequest,
			message: "sourceUrl 必须是合法的绝对地址，或基于工作台入口可解析的相对地址",
		}
	}
	if !r.isWorkspaceArtifactURLTrustedWithAccess(session, accessEntries, resolvedSourceURL) {
		return "", "", workspaceArtifactValidationError{
			status:  http.StatusForbidden,
			message: "sourceUrl 不在可信预览域名范围内",
		}
	}

	resolvedPreviewURL := ""
	if strings.TrimSpace(previewURL) != "" {
		resolvedPreviewURL, err = r.normalizeWorkspaceArtifactURL(session, previewURL)
		if err != nil {
			return "", "", workspaceArtifactValidationError{
				status:  http.StatusBadRequest,
				message: "previewUrl 必须是合法的绝对地址，或基于工作台入口可解析的相对地址",
			}
		}
		if !r.isWorkspaceArtifactURLTrustedWithAccess(session, accessEntries, resolvedPreviewURL) {
			return "", "", workspaceArtifactValidationError{
				status:  http.StatusForbidden,
				message: "previewUrl 不在可信预览域名范围内",
			}
		}
	}

	return resolvedSourceURL, resolvedPreviewURL, nil
}

func (r *Router) normalizeWorkspaceArtifactURL(session models.WorkspaceSession, raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "" || parsed.Host != "" {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return "", fmt.Errorf("仅支持 http/https 地址")
		}
		parsed.Fragment = ""
		return parsed.String(), nil
	}

	base, err := url.Parse(strings.TrimSpace(session.WorkspaceURL))
	if err != nil {
		return "", err
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return "", fmt.Errorf("workspace url is invalid")
	}
	resolved := base.ResolveReference(parsed)
	resolved.Fragment = ""
	return resolved.String(), nil
}

func (r *Router) workspaceArtifactAccessEntries(instanceID int) []models.InstanceAccess {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]models.InstanceAccess(nil), r.filterAccessByInstance(instanceID)...)
}

func (r *Router) workspaceArtifactTrustedHosts(session models.WorkspaceSession, accessEntries []models.InstanceAccess) []string {
	hosts := make([]string, 0, 8)
	appendHost := func(raw string) {
		host := canonicalArtifactHost(raw)
		if host == "" {
			return
		}
		for _, existing := range hosts {
			if existing == host {
				return
			}
		}
		hosts = append(hosts, host)
	}

	appendHost(session.WorkspaceURL)
	for _, entry := range accessEntries {
		appendHost(entry.URL)
	}
	for _, raw := range strings.Split(r.config.ArtifactPreviewAllowedHosts, ",") {
		appendHost(raw)
	}

	return hosts
}

func (r *Router) isWorkspaceArtifactURLTrusted(session models.WorkspaceSession, instance models.Instance, raw string) bool {
	return r.isWorkspaceArtifactURLTrustedWithAccess(session, r.workspaceArtifactAccessEntries(instance.ID), raw)
}

func (r *Router) isWorkspaceArtifactURLTrustedWithAccess(session models.WorkspaceSession, accessEntries []models.InstanceAccess, raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return false
	}
	for _, trusted := range r.workspaceArtifactTrustedHosts(session, accessEntries) {
		if host == trusted || strings.HasSuffix(host, "."+trusted) || sameArtifactBaseDomain(host, trusted) {
			return true
		}
	}
	return false
}

func canonicalArtifactHost(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return ""
		}
		return strings.ToLower(parsed.Hostname())
	}
	return strings.ToLower(strings.TrimPrefix(trimmed, "."))
}

func sameArtifactBaseDomain(left string, right string) bool {
	if left == "" || right == "" {
		return false
	}
	if net.ParseIP(left) != nil || net.ParseIP(right) != nil {
		return false
	}
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	if len(leftParts) < 2 || len(rightParts) < 2 {
		return false
	}
	return strings.Join(leftParts[len(leftParts)-2:], ".") == strings.Join(rightParts[len(rightParts)-2:], ".")
}

func (r *Router) validateArtifactTargetURL(ctx context.Context, raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("地址格式无效")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("仅支持 http/https 地址")
	}
	if parsed.Host == "" {
		return fmt.Errorf("地址缺少主机名")
	}
	if parsed.User != nil {
		return fmt.Errorf("不允许使用带账号口令的地址")
	}
	if r.config.ArtifactPreviewAllowPrivateIP {
		return nil
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("地址缺少主机名")
	}
	if strings.EqualFold(hostname, "localhost") {
		return fmt.Errorf("不允许代理 localhost 或私网地址")
	}

	if ip := net.ParseIP(hostname); ip != nil {
		if isPrivateArtifactIP(ip) {
			return fmt.Errorf("不允许代理 localhost 或私网地址")
		}
		return nil
	}

	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return fmt.Errorf("无法解析目标地址")
	}
	for _, addr := range addresses {
		if isPrivateArtifactIP(addr.IP) {
			return fmt.Errorf("不允许代理 localhost 或私网地址")
		}
	}
	return nil
}

func isPrivateArtifactIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsUnspecified()
}

func (r *Router) serveArchivedWorkspaceArtifact(w http.ResponseWriter, req *http.Request, artifact models.WorkspaceArtifact, inline bool, mode string) {
	if r.artifactStore == nil || !r.artifactStore.Enabled() {
		http.Error(w, "artifact archive is not configured", http.StatusServiceUnavailable)
		return
	}
	reader, contentType, size, err := r.artifactStore.Open(req.Context(), artifact.StorageKey)
	if err != nil {
		http.Error(w, "open archived artifact failed", http.StatusBadGateway)
		return
	}
	defer reader.Close()

	headers := w.Header()
	headers.Set("Cache-Control", "private, no-store, max-age=0")
	headers.Set("Pragma", "no-cache")
	headers.Set("Referrer-Policy", "no-referrer")
	headers.Set("X-Content-Type-Options", "nosniff")
	targetURL := "https://artifact.local/" + firstNonEmpty(strings.TrimSpace(artifact.Filename), path.Base(strings.TrimSpace(artifact.StorageKey)))
	setProxyDispositionHeader(headers, artifact, targetURL, inline)

	normalizedContentType := strings.TrimSpace(contentType)
	if normalizedContentType == "" {
		normalizedContentType = normalizeArtifactContentType("", artifact.Filename, artifact.Kind)
	}
	headers.Set("Content-Type", normalizedContentType)

	if inline && (mode == "web" || mode == "html") {
		body, err := io.ReadAll(io.LimitReader(reader, r.artifactPreviewHTMLLimit()+1))
		if err != nil {
			http.Error(w, "read archived html failed", http.StatusBadGateway)
			return
		}
		if int64(len(body)) > r.artifactPreviewHTMLLimit() {
			http.Error(w, "artifact html is too large to sandbox preview", http.StatusRequestEntityTooLarge)
			return
		}

		headers.Set("Content-Security-Policy", "sandbox allow-same-origin; default-src 'none'; img-src data: https: http:; media-src data: https: http:; font-src data: https: http:; style-src 'unsafe-inline' https: http:; frame-ancestors 'self'; form-action 'none'; base-uri 'none'; script-src 'none'")
		rewritten := rewriteArtifactPreviewHTML(body, targetURL)
		headers.Set("Content-Length", strconv.Itoa(len(rewritten)))
		w.WriteHeader(http.StatusOK)
		if req.Method != http.MethodHead {
			_, _ = w.Write(rewritten)
		}
		return
	}

	if size > 0 {
		headers.Set("Content-Length", strconv.FormatInt(size, 10))
	}
	w.WriteHeader(http.StatusOK)
	if req.Method != http.MethodHead {
		_, _ = io.Copy(w, reader)
	}
}

func (r *Router) proxyWorkspaceArtifactURL(w http.ResponseWriter, req *http.Request, artifact models.WorkspaceArtifact, targetURL string, inline bool, mode string) {
	if err := r.validateArtifactTargetURL(req.Context(), targetURL); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	upstreamReq, err := http.NewRequestWithContext(req.Context(), req.Method, targetURL, nil)
	if err != nil {
		http.Error(w, "build artifact proxy request failed", http.StatusBadGateway)
		return
	}
	copyRequestHeader(upstreamReq.Header, req.Header, "Range", "If-Range", "If-None-Match", "If-Modified-Since", "Accept")
	upstreamReq.Header.Set("User-Agent", "OpenClawArtifactPreview/1.0")

	resp, err := r.artifactPreviewHTTPClient().Do(upstreamReq)
	if err != nil {
		http.Error(w, "artifact proxy request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.Request != nil {
		if err := r.validateArtifactTargetURL(req.Context(), resp.Request.URL.String()); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	if resp.StatusCode >= 400 {
		http.Error(w, fmt.Sprintf("artifact upstream returned %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	headers := w.Header()
	headers.Set("Cache-Control", "private, no-store, max-age=0")
	headers.Set("Pragma", "no-cache")
	headers.Set("Referrer-Policy", "no-referrer")
	headers.Set("X-Content-Type-Options", "nosniff")
	setProxyDispositionHeader(headers, artifact, targetURL, inline)

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	mediaType := ""
	if contentType != "" {
		parsedType, _, err := mime.ParseMediaType(contentType)
		if err == nil {
			mediaType = strings.ToLower(parsedType)
			headers.Set("Content-Type", contentType)
		}
	}

	if inline && mode == "html" {
		body, err := io.ReadAll(io.LimitReader(resp.Body, r.artifactPreviewHTMLLimit()+1))
		if err != nil {
			http.Error(w, "read artifact html failed", http.StatusBadGateway)
			return
		}
		if int64(len(body)) > r.artifactPreviewHTMLLimit() {
			http.Error(w, "artifact html is too large to sandbox preview", http.StatusRequestEntityTooLarge)
			return
		}

		if mediaType == "" {
			headers.Set("Content-Type", "text/html; charset=utf-8")
		}
		headers.Set("Content-Security-Policy", "sandbox allow-scripts allow-forms allow-modals allow-popups allow-downloads; default-src https: http: data: blob:; img-src https: http: data: blob:; media-src https: http: data: blob:; font-src https: http: data: blob:; style-src 'unsafe-inline' https: http:; script-src 'unsafe-inline' 'unsafe-eval' https: http:; connect-src https: http: data: blob:")
		rewritten := rewriteArtifactPreviewHTML(body, targetURL)
		headers.Set("Content-Length", strconv.Itoa(len(rewritten)))
		w.WriteHeader(http.StatusOK)
		if req.Method != http.MethodHead {
			_, _ = w.Write(rewritten)
		}
		return
	}

	copyResponseHeader(headers, resp.Header, "Accept-Ranges", "Content-Range", "Content-Length", "ETag", "Last-Modified")
	w.WriteHeader(resp.StatusCode)
	if req.Method != http.MethodHead {
		_, _ = io.Copy(w, resp.Body)
	}
}

func rewriteArtifactPreviewHTML(body []byte, targetURL string) []byte {
	raw := artifactPreviewMetaCSPRegexp.ReplaceAllString(string(body), "")
	injection := fmt.Sprintf(`<base href="%s"><meta name="referrer" content="no-referrer">`, html.EscapeString(targetURL))

	if match := artifactPreviewHeadPattern.FindStringIndex(raw); match != nil {
		insertAt := match[1]
		raw = raw[:insertAt] + injection + raw[insertAt:]
		return []byte(raw)
	}

	return []byte("<head>" + injection + "</head>" + raw)
}

func copyRequestHeader(dst http.Header, src http.Header, keys ...string) {
	for _, key := range keys {
		if value := src.Get(key); value != "" {
			dst.Set(key, value)
		}
	}
}

func copyResponseHeader(dst http.Header, src http.Header, keys ...string) {
	for _, key := range keys {
		if value := src.Get(key); value != "" {
			dst.Set(key, value)
		}
	}
}

func setProxyDispositionHeader(headers http.Header, artifact models.WorkspaceArtifact, targetURL string, inline bool) {
	filename := artifactProxyFileName(artifact, targetURL)
	if inline {
		headers.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
		return
	}
	headers.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
}

func artifactProxyFileName(artifact models.WorkspaceArtifact, targetURL string) string {
	ext := ".bin"
	if parsed, err := url.Parse(targetURL); err == nil {
		if value := strings.TrimSpace(path.Ext(parsed.Path)); value != "" {
			ext = value
		}
	}

	title := strings.TrimSpace(artifact.Title)
	if title == "" {
		return fmt.Sprintf("artifact-%d%s", artifact.ID, ext)
	}

	safe := make([]rune, 0, len(title))
	for _, char := range title {
		switch {
		case char >= 'a' && char <= 'z':
			safe = append(safe, char)
		case char >= 'A' && char <= 'Z':
			safe = append(safe, char)
		case char >= '0' && char <= '9':
			safe = append(safe, char)
		case char == '-' || char == '_' || char == '.':
			safe = append(safe, char)
		case char == ' ':
			safe = append(safe, '-')
		}
	}
	if len(safe) == 0 {
		return fmt.Sprintf("artifact-%d%s", artifact.ID, ext)
	}
	name := strings.Trim(strings.TrimSpace(string(safe)), ".-")
	if name == "" {
		return fmt.Sprintf("artifact-%d%s", artifact.ID, ext)
	}
	if strings.HasSuffix(strings.ToLower(name), strings.ToLower(ext)) {
		return name
	}
	return name + ext
}
