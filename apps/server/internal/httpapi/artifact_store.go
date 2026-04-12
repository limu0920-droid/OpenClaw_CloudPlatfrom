package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type artifactObjectStore interface {
	Enabled() bool
	Ping(ctx context.Context) error
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, string, int64, error)
}

type minioArtifactStore struct {
	client   *minio.Client
	bucket   string
	region   string
	strict   bool
	initMu   sync.Mutex
	ready    bool
	disabled bool
}

func newArtifactObjectStore(cfg ExternalConfig) (artifactObjectStore, error) {
	endpoint := strings.TrimSpace(cfg.ObjectStorageEndpoint)
	bucket := strings.TrimSpace(cfg.ObjectStorageBucket)
	accessKey := strings.TrimSpace(cfg.ObjectStorageAccessKey)
	secretKey := strings.TrimSpace(cfg.ObjectStorageSecretKey)
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, nil
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse object storage endpoint: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("object storage endpoint must include scheme and host")
	}

	bucketLookup := minio.BucketLookupAuto
	if cfg.ObjectStorageForcePathStyle {
		bucketLookup = minio.BucketLookupPath
	}

	client, err := minio.New(parsed.Host, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       strings.EqualFold(parsed.Scheme, "https"),
		Region:       strings.TrimSpace(cfg.ObjectStorageRegion),
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return nil, fmt.Errorf("create object storage client: %w", err)
	}

	return &minioArtifactStore{
		client: client,
		bucket: bucket,
		region: strings.TrimSpace(cfg.ObjectStorageRegion),
		strict: cfg.StrictMode,
	}, nil
}

func (s *minioArtifactStore) Enabled() bool {
	return s != nil && s.client != nil && !s.disabled
}

func (s *minioArtifactStore) Ping(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("object storage bucket %q does not exist", s.bucket)
	}
	return nil
}

func (s *minioArtifactStore) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if !s.Enabled() {
		return errors.New("artifact object store is not configured")
	}
	if err := s.ensureBucket(ctx); err != nil {
		return err
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: defaultString(strings.TrimSpace(contentType), "application/octet-stream"),
	})
	return err
}

func (s *minioArtifactStore) Open(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	if !s.Enabled() {
		return nil, "", 0, errors.New("artifact object store is not configured")
	}
	reader, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", 0, err
	}
	info, err := reader.Stat()
	if err != nil {
		_ = reader.Close()
		return nil, "", 0, err
	}
	return reader, info.ContentType, info.Size, nil
}

func (s *minioArtifactStore) ensureBucket(ctx context.Context) error {
	if !s.Enabled() {
		return errors.New("artifact object store is not configured")
	}

	s.initMu.Lock()
	defer s.initMu.Unlock()
	if s.ready {
		return nil
	}

	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if s.strict {
			return fmt.Errorf("object storage bucket %q does not exist", s.bucket)
		}
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: s.region}); err != nil {
			exists, bucketErr := s.client.BucketExists(ctx, s.bucket)
			if bucketErr != nil || !exists {
				return err
			}
		}
	}
	s.ready = true
	return nil
}

type artifactArchiveResult struct {
	ArchiveStatus  string
	ContentType    string
	SizeBytes      int64
	StorageBucket  string
	StorageKey     string
	Filename       string
	ChecksumSHA256 string
}

func (r *Router) archiveArtifactFromSource(ctx context.Context, sourceURL string, sessionID int, tenantID int, instanceID int, title string, kind string) (artifactArchiveResult, error) {
	filename := artifactFilename(sourceURL, title, kind)
	if r.artifactStore == nil || !r.artifactStore.Enabled() {
		return artifactArchiveResult{
			ArchiveStatus: "not_configured",
			Filename:      filename,
		}, nil
	}

	if err := validateArtifactSourceURL(ctx, sourceURL, r.config.ArtifactArchiveAllowPrivateURL); err != nil {
		return artifactArchiveResult{}, err
	}

	maxBytes := r.config.ArtifactArchiveMaxBytes
	if maxBytes <= 0 {
		maxBytes = 52_428_800
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return artifactArchiveResult{}, err
	}

	resp, err := r.artifactHTTPClient().Do(req)
	if err != nil {
		return artifactArchiveResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return artifactArchiveResult{}, fmt.Errorf("archive source returned status %d", resp.StatusCode)
	}
	if resp.ContentLength > maxBytes && resp.ContentLength > 0 {
		return artifactArchiveResult{}, fmt.Errorf("artifact exceeds archive limit of %d bytes", maxBytes)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return artifactArchiveResult{}, err
	}
	if int64(len(body)) > maxBytes {
		return artifactArchiveResult{}, fmt.Errorf("artifact exceeds archive limit of %d bytes", maxBytes)
	}

	contentType := normalizeArtifactContentType(resp.Header.Get("Content-Type"), filename, kind)
	sum := sha256.Sum256(body)
	key := artifactStorageKey(tenantID, instanceID, sessionID, filename)

	if err := r.artifactStore.Put(ctx, key, bytes.NewReader(body), int64(len(body)), contentType); err != nil {
		return artifactArchiveResult{}, err
	}

	return artifactArchiveResult{
		ArchiveStatus:  "archived",
		ContentType:    contentType,
		SizeBytes:      int64(len(body)),
		StorageBucket:  r.config.ObjectStorageBucket,
		StorageKey:     key,
		Filename:       filename,
		ChecksumSHA256: hex.EncodeToString(sum[:]),
	}, nil
}

func (r *Router) artifactHTTPClient() *http.Client {
	if r != nil && r.config.HTTPClient != nil {
		return r.config.HTTPClient
	}
	return &http.Client{Timeout: 20 * time.Second}
}

func validateArtifactSourceURL(ctx context.Context, rawURL string, allowPrivate bool) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("sourceUrl is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("sourceUrl must use http or https")
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return errors.New("sourceUrl host is required")
	}
	if allowPrivate {
		return nil
	}
	if isBlockedArtifactHost(host) {
		return errors.New("sourceUrl host is not allowed for archiving")
	}

	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("resolve sourceUrl host: %w", err)
	}
	if len(ips) == 0 {
		return errors.New("sourceUrl host resolved to no addresses")
	}
	for _, ip := range ips {
		if !ip.IsValid() || isPrivateArtifactAddr(ip.Unmap()) {
			return errors.New("sourceUrl resolves to a private or loopback address")
		}
	}
	return nil
}

func isBlockedArtifactHost(host string) bool {
	normalized := strings.ToLower(strings.TrimSpace(host))
	return normalized == "localhost" ||
		normalized == "0.0.0.0" ||
		normalized == "::1" ||
		strings.HasSuffix(normalized, ".local")
}

func isPrivateArtifactAddr(addr netip.Addr) bool {
	return addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsUnspecified()
}

func artifactFilename(sourceURL string, title string, kind string) string {
	parsed, err := url.Parse(strings.TrimSpace(sourceURL))
	if err == nil {
		name := path.Base(parsed.Path)
		if name != "" && name != "." && name != "/" {
			return name
		}
	}
	base := artifactStoreSlugify(defaultString(strings.TrimSpace(title), "artifact"))
	ext := artifactFileExtension(kind)
	if ext == "" {
		return base
	}
	return base + ext
}

func artifactFileExtension(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "web":
		return ".html"
	case "pdf":
		return ".pdf"
	case "pptx":
		return ".pptx"
	case "docx":
		return ".docx"
	case "xlsx":
		return ".xlsx"
	case "image":
		return ".png"
	case "video":
		return ".mp4"
	case "audio":
		return ".mp3"
	case "text":
		return ".txt"
	default:
		return ""
	}
}

func artifactStorageKey(tenantID int, instanceID int, sessionID int, filename string) string {
	stampedName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), artifactStoreSlugify(filename))
	return fmt.Sprintf("artifacts/tenant-%d/instance-%d/session-%d/%s", tenantID, instanceID, sessionID, stampedName)
}

func normalizeArtifactContentType(headerValue string, filename string, kind string) string {
	contentType := strings.TrimSpace(headerValue)
	if contentType != "" {
		return strings.TrimSpace(strings.Split(contentType, ";")[0])
	}
	if guessed := mime.TypeByExtension(strings.ToLower(path.Ext(filename))); guessed != "" {
		return guessed
	}
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "web":
		return "text/html; charset=utf-8"
	case "pdf":
		return "application/pdf"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "image":
		return "image/png"
	case "video":
		return "video/mp4"
	case "audio":
		return "audio/mpeg"
	case "text":
		return "text/plain; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

func artifactStoreSlugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "artifact"
	}
	var builder strings.Builder
	lastHyphen := false
	for _, r := range value {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			builder.WriteRune(r)
			lastHyphen = false
		case r == '.' || r == '_' || r == '-':
			builder.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen {
				builder.WriteRune('-')
				lastHyphen = true
			}
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "artifact"
	}
	return result
}
