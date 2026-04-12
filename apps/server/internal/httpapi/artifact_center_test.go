package httpapi

import (
	"net/http"
	"strconv"
	"testing"

	"openclaw/platformapi/internal/models"
)

func TestArtifactCenterListIncludesStatsFavoritesAndRecentViewed(t *testing.T) {
	router := newTestRouter(ExternalConfig{})
	router.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{{
		ID:            50,
		SessionID:     1,
		TenantID:      1,
		InstanceID:    100,
		MessageID:     2,
		Title:         "路演汇报",
		Kind:          "pptx",
		SourceURL:     "https://files.acme.example.com/artifacts/demo-deck-v2.pptx",
		PreviewURL:    "https://files.acme.example.com/artifacts/demo-deck-preview.pdf",
		ArchiveStatus: "archived",
		Filename:      "demo-deck-v2.pptx",
		CreatedAt:     "2026-04-05T10:12:00Z",
		UpdatedAt:     "2026-04-05T10:12:00Z",
	}}, router.data.WorkspaceArtifacts...)

	cookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Alice",
		Email:         "alice@acme.example.com",
		Role:          "tenant_admin",
	})

	favorite := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/portal/artifacts/50/favorite", nil, cookie)
	if favorite.Code != http.StatusOK {
		t.Fatalf("expected favorite status 200, got %d: %s", favorite.Code, favorite.Body.String())
	}

	share := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/portal/artifacts/50/shares", map[string]any{
		"note":          "share to marketing",
		"expiresInDays": 7,
	}, cookie)
	if share.Code != http.StatusCreated {
		t.Fatalf("expected share status 201, got %d: %s", share.Code, share.Body.String())
	}

	detail := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/portal/artifacts/50", nil, cookie)
	if detail.Code != http.StatusOK {
		t.Fatalf("expected artifact detail status 200, got %d: %s", detail.Code, detail.Body.String())
	}

	list := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/portal/artifacts", nil, cookie)
	if list.Code != http.StatusOK {
		t.Fatalf("expected artifact center status 200, got %d: %s", list.Code, list.Body.String())
	}

	var response struct {
		Items []struct {
			ID            int  `json:"id"`
			IsFavorite    bool `json:"isFavorite"`
			ShareCount    int  `json:"shareCount"`
			Version       int  `json:"version"`
			LatestVersion int  `json:"latestVersion"`
			Thumbnail     struct {
				Label string `json:"label"`
			} `json:"thumbnail"`
			Quality struct {
				Score         int  `json:"score"`
				InlinePreview bool `json:"inlinePreview"`
			} `json:"quality"`
		} `json:"items"`
		RecentViewed []struct {
			ID int `json:"id"`
		} `json:"recentViewed"`
		Stats struct {
			FavoriteCount     int `json:"favoriteCount"`
			SharedCount       int `json:"sharedCount"`
			VersionedCount    int `json:"versionedCount"`
			RecentViewedCount int `json:"recentViewedCount"`
		} `json:"stats"`
	}
	decodeResponse(t, list, &response)

	foundCurrentVersion := false
	currentIsFavorite := false
	currentShareCount := 0
	currentVersionNumber := 0
	currentLatestVersion := 0
	currentThumbnailLabel := ""
	currentQualityScore := 0
	for index := range response.Items {
		if response.Items[index].ID == 50 {
			foundCurrentVersion = true
			currentIsFavorite = response.Items[index].IsFavorite
			currentShareCount = response.Items[index].ShareCount
			currentVersionNumber = response.Items[index].Version
			currentLatestVersion = response.Items[index].LatestVersion
			currentThumbnailLabel = response.Items[index].Thumbnail.Label
			currentQualityScore = response.Items[index].Quality.Score
			break
		}
	}

	if !foundCurrentVersion {
		t.Fatal("expected current artifact version in center list")
	}
	if !currentIsFavorite {
		t.Fatalf("expected artifact favorite state, got %#v", response.Items)
	}
	if currentShareCount != 1 {
		t.Fatalf("expected one active share, got %#v", response.Items)
	}
	if currentVersionNumber != 2 || currentLatestVersion != 2 {
		t.Fatalf("expected version chain 2/2, got %#v", response.Items)
	}
	if currentThumbnailLabel == "" {
		t.Fatalf("expected thumbnail summary, got %#v", response.Items)
	}
	if currentQualityScore == 0 {
		t.Fatalf("expected quality score, got %#v", response.Items)
	}
	if len(response.RecentViewed) == 0 || response.RecentViewed[0].ID != 50 {
		t.Fatalf("expected recent viewed artifact 50, got %#v", response.RecentViewed)
	}
	if response.Stats.FavoriteCount != 1 || response.Stats.SharedCount != 1 || response.Stats.VersionedCount != 1 {
		t.Fatalf("unexpected artifact stats %#v", response.Stats)
	}
	if response.Stats.RecentViewedCount == 0 {
		t.Fatalf("expected recent viewed count in stats, got %#v", response.Stats)
	}
}

func TestArtifactDetailShareLifecycleAndVersions(t *testing.T) {
	router := newTestRouter(ExternalConfig{})
	router.data.WorkspaceArtifacts = append([]models.WorkspaceArtifact{{
		ID:            50,
		SessionID:     1,
		TenantID:      1,
		InstanceID:    100,
		MessageID:     2,
		Title:         "路演汇报",
		Kind:          "pptx",
		SourceURL:     "https://files.acme.example.com/artifacts/demo-deck-v2.pptx",
		PreviewURL:    "https://files.acme.example.com/artifacts/demo-deck-preview.pdf",
		ArchiveStatus: "archived",
		Filename:      "demo-deck-v2.pptx",
		CreatedAt:     "2026-04-05T10:12:00Z",
		UpdatedAt:     "2026-04-05T10:12:00Z",
	}}, router.data.WorkspaceArtifacts...)

	cookie := workspaceAuthCookie(t, router, authSessionState{
		Authenticated: true,
		UserID:        1,
		TenantID:      1,
		Name:          "Alice",
		Email:         "alice@acme.example.com",
		Role:          "tenant_admin",
	})

	share := performRequestWithCookies(t, router, http.MethodPost, "/api/v1/portal/artifacts/50/shares", map[string]any{
		"note":          "share to reviewer",
		"expiresInDays": 7,
	}, cookie)
	if share.Code != http.StatusCreated {
		t.Fatalf("expected share status 201, got %d: %s", share.Code, share.Body.String())
	}

	var shareResponse struct {
		Share struct {
			ID    int    `json:"id"`
			Token string `json:"token"`
		} `json:"share"`
	}
	decodeResponse(t, share, &shareResponse)

	detail := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/portal/artifacts/50?share="+shareResponse.Share.Token, nil, cookie)
	if detail.Code != http.StatusOK {
		t.Fatalf("expected detail status 200, got %d: %s", detail.Code, detail.Body.String())
	}

	var detailResponse struct {
		Artifact struct {
			ID int `json:"id"`
		} `json:"artifact"`
		Versions []struct {
			ID      int `json:"id"`
			Version int `json:"version"`
		} `json:"versions"`
		Shares []struct {
			ID       int  `json:"id"`
			Active   bool `json:"active"`
			UseCount int  `json:"useCount"`
		} `json:"shares"`
	}
	decodeResponse(t, detail, &detailResponse)

	if detailResponse.Artifact.ID != 50 {
		t.Fatalf("expected artifact 50, got %#v", detailResponse.Artifact)
	}
	if len(detailResponse.Versions) < 2 {
		t.Fatalf("expected version timeline, got %#v", detailResponse.Versions)
	}
	if len(detailResponse.Shares) != 1 || detailResponse.Shares[0].UseCount != 1 || !detailResponse.Shares[0].Active {
		t.Fatalf("expected active share with one open, got %#v", detailResponse.Shares)
	}

	revoke := performRequestWithCookies(t, router, http.MethodDelete, "/api/v1/portal/artifact-shares/"+
		strconv.Itoa(shareResponse.Share.ID), nil, cookie)
	if revoke.Code != http.StatusOK {
		t.Fatalf("expected revoke status 200, got %d: %s", revoke.Code, revoke.Body.String())
	}

	detailAfterRevoke := performRequestWithCookies(t, router, http.MethodGet, "/api/v1/portal/artifacts/50", nil, cookie)
	if detailAfterRevoke.Code != http.StatusOK {
		t.Fatalf("expected detail after revoke status 200, got %d: %s", detailAfterRevoke.Code, detailAfterRevoke.Body.String())
	}

	decodeResponse(t, detailAfterRevoke, &detailResponse)
	if len(detailResponse.Shares) != 1 || detailResponse.Shares[0].Active {
		t.Fatalf("expected revoked share in detail response, got %#v", detailResponse.Shares)
	}
}
