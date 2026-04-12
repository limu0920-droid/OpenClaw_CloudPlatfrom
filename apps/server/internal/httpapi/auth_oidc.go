package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/oidc"
)

type authSessionState struct {
	UserID           int    `json:"userId,omitempty"`
	TenantID         int    `json:"tenantId,omitempty"`
	Provider         string `json:"provider"`
	Authenticated    bool   `json:"authenticated"`
	Name             string `json:"name,omitempty"`
	Email            string `json:"email,omitempty"`
	Role             string `json:"role,omitempty"`
	Locale           string `json:"locale,omitempty"`
	Timezone         string `json:"timezone,omitempty"`
	OpenID           string `json:"openId,omitempty"`
	UnionID          string `json:"unionId,omitempty"`
	AccessToken      string `json:"accessToken,omitempty"`
	RefreshToken     string `json:"refreshToken,omitempty"`
	IDToken          string `json:"idToken,omitempty"`
	AccessExpiresAt  string `json:"accessExpiresAt,omitempty"`
	RefreshExpiresAt string `json:"refreshExpiresAt,omitempty"`
}

type authFlowState struct {
	Provider    string `json:"provider"`
	State       string `json:"state"`
	Nonce       string `json:"nonce"`
	RedirectURI string `json:"redirectUri"`
	Next        string `json:"next"`
	ExpiresAt   string `json:"expiresAt"`
}

type tokenExchangeRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirectUri"`
}

func (r *Router) handleAuthMe(w http.ResponseWriter, req *http.Request) {
	session, ok := r.readAuthSession(req)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
			"provider":      "none",
		})
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (r *Router) handleAuthToken(w http.ResponseWriter, req *http.Request) {
	var payload tokenExchangeRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Code) == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.RedirectURI) == "" {
		payload.RedirectURI = r.config.KeycloakRedirectURL
	}

	session, err := r.exchangeCodeToSession(req.Context(), payload.Code, payload.RedirectURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.writeAuthSession(w, session)
	writeJSON(w, http.StatusOK, session)
}

func (r *Router) handleAuthCallback(w http.ResponseWriter, req *http.Request) {
	code := strings.TrimSpace(req.URL.Query().Get("code"))
	if code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	flow, ok := r.readAuthFlow(req)
	if !ok {
		http.Error(w, "missing auth flow", http.StatusBadRequest)
		return
	}
	if state := strings.TrimSpace(req.URL.Query().Get("state")); state == "" || state != flow.State {
		http.Error(w, "invalid auth state", http.StatusBadRequest)
		return
	}

	redirectURI := flow.RedirectURI
	if redirectURI == "" {
		redirectURI = r.config.KeycloakRedirectURL
	}
	nextTarget := normalizeRedirectTarget(flow.Next, r.defaultPostLoginRedirect())

	session, err := r.exchangeCodeToSession(req.Context(), code, redirectURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.writeAuthSession(w, session)
	r.clearAuthFlow(w)

	payload := map[string]any{
		"authenticated": true,
		"provider":      session.Provider,
		"redirectTo":    nextTarget,
		"user": map[string]any{
			"name":  session.Name,
			"email": session.Email,
			"role":  session.Role,
		},
	}

	if shouldJSONResponse(req) {
		writeJSON(w, http.StatusOK, payload)
		return
	}

	http.Redirect(w, req, nextTarget, http.StatusFound)
}

func (r *Router) handleAuthRefresh(w http.ResponseWriter, req *http.Request) {
	session, ok := r.readAuthSession(req)
	if !ok {
		http.Error(w, "no active session", http.StatusUnauthorized)
		return
	}

	if !r.config.KeycloakEnabled || r.config.KeycloakBaseURL == "" || r.config.KeycloakRealm == "" || r.config.KeycloakClientID == "" {
		http.Error(w, "keycloak refresh is unavailable", http.StatusServiceUnavailable)
		return
	}

	client := r.newOIDCClient()
	token, err := client.RefreshToken(req.Context(), session.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	session.AccessToken = token.AccessToken
	session.RefreshToken = token.RefreshToken
	session.IDToken = token.IDToken
	session.AccessExpiresAt = time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339)
	session.RefreshExpiresAt = time.Now().UTC().Add(time.Duration(token.RefreshExpiresIn) * time.Second).Format(time.RFC3339)
	r.writeAuthSession(w, session)

	writeJSON(w, http.StatusOK, session)
}

func (r *Router) handleAuthLogout(w http.ResponseWriter, req *http.Request) {
	session, _ := r.readAuthSession(req)
	r.clearAuthSession(w)

	if session.Provider == "keycloak" && r.config.KeycloakEnabled {
		client := r.newOIDCClient()
		logoutURL, err := client.LogoutURL(context.Background(), session.IDToken, r.config.KeycloakLogoutRedirectURL)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"loggedOut": true,
				"provider":  "keycloak",
				"logoutUrl": logoutURL,
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"loggedOut": true,
		"provider":  session.Provider,
	})
}

func (r *Router) exchangeCodeToSession(ctx context.Context, code string, redirectURI string) (authSessionState, error) {
	if !r.config.KeycloakEnabled || r.config.KeycloakBaseURL == "" || r.config.KeycloakRealm == "" || r.config.KeycloakClientID == "" {
		return authSessionState{}, r.strictModeError("keycloak is not configured")
	}

	client := r.newOIDCClient()
	token, err := client.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return authSessionState{}, err
	}
	userInfo, err := client.UserInfo(ctx, token.AccessToken)
	if err != nil {
		return authSessionState{}, err
	}

	profile, identity := r.resolveUserFromIdentity("keycloak", userInfo.Email, "", "", defaultString(userInfo.Sub, userInfo.Email), defaultString(userInfo.Name, userInfo.PreferredUsername))
	session := authSessionState{
		Provider:         "keycloak",
		Authenticated:    true,
		Name:             defaultString(userInfo.Name, userInfo.PreferredUsername),
		Email:            userInfo.Email,
		Role:             defaultString(r.config.KeycloakMockRole, "tenant_admin"),
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		IDToken:          token.IDToken,
		AccessExpiresAt:  time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339),
		RefreshExpiresAt: time.Now().UTC().Add(time.Duration(token.RefreshExpiresIn) * time.Second).Format(time.RFC3339),
	}

	return r.buildAuthSessionFromProfile(profile, identity, session), nil
}

func (r *Router) newOIDCClient() *oidc.Client {
	return oidc.NewClient(oidc.Config{
		BaseURL:      r.config.KeycloakBaseURL,
		Realm:        r.config.KeycloakRealm,
		ClientID:     r.config.KeycloakClientID,
		ClientSecret: r.config.KeycloakClientSecret,
		RedirectURL:  r.config.KeycloakRedirectURL,
	})
}

func (r *Router) authCookieName() string {
	if strings.TrimSpace(r.config.KeycloakCookieName) != "" {
		return r.config.KeycloakCookieName
	}
	return "openclaw_auth"
}

func (r *Router) authFlowCookieName() string {
	if strings.TrimSpace(r.config.KeycloakFlowCookieName) != "" {
		return r.config.KeycloakFlowCookieName
	}
	return "openclaw_auth_flow"
}

func (r *Router) defaultPostLoginRedirect() string {
	if strings.TrimSpace(r.config.KeycloakPostLoginRedirectURL) != "" {
		return r.config.KeycloakPostLoginRedirectURL
	}
	return "/portal"
}

func (r *Router) sessionSecret() string {
	if strings.TrimSpace(r.config.KeycloakSessionSecret) != "" {
		return r.config.KeycloakSessionSecret
	}
	return "dev-session-secret-change-me"
}

func (r *Router) writeAuthFlow(w http.ResponseWriter, flow authFlowState) {
	r.writeSignedCookie(w, r.authFlowCookieName(), flow, 10*60)
}

func (r *Router) readAuthFlow(req *http.Request) (authFlowState, bool) {
	var flow authFlowState
	if !r.readSignedCookie(req, r.authFlowCookieName(), &flow) {
		return authFlowState{}, false
	}
	return flow, true
}

func (r *Router) clearAuthFlow(w http.ResponseWriter) {
	r.clearSignedCookie(w, r.authFlowCookieName())
}

func (r *Router) writeAuthSession(w http.ResponseWriter, session authSessionState) {
	r.writeSignedCookie(w, r.authCookieName(), session, 0)
}

func (r *Router) writeSignedCookie(w http.ResponseWriter, name string, value any, maxAge int) {
	payload, _ := json.Marshal(value)
	raw := base64.RawURLEncoding.EncodeToString(payload)
	signature := r.signSession(raw)
	cookieValue := raw + "." + signature

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.config.KeycloakCookieSecure,
		MaxAge:   maxAge,
	})
}

func (r *Router) readAuthSession(req *http.Request) (authSessionState, bool) {
	var session authSessionState
	if !r.readSignedCookie(req, r.authCookieName(), &session) {
		return authSessionState{}, false
	}
	return session, true
}

func (r *Router) readSignedCookie(req *http.Request, name string, target any) bool {
	cookie, err := req.Cookie(name)
	if err != nil || cookie.Value == "" {
		return false
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return false
	}
	if !hmac.Equal([]byte(parts[1]), []byte(r.signSession(parts[0]))) {
		return false
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return false
	}
	return true
}

func (r *Router) clearAuthSession(w http.ResponseWriter) {
	r.clearSignedCookie(w, r.authCookieName())
}

func (r *Router) clearSignedCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.config.KeycloakCookieSecure,
		MaxAge:   -1,
	})
}

func (r *Router) signSession(raw string) string {
	h := hmac.New(sha256.New, []byte(r.sessionSecret()))
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

func generateAuthRandom(size int) string {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return base64.RawURLEncoding.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func shouldJSONResponse(req *http.Request) bool {
	if req.URL.Query().Get("mode") == "json" {
		return true
	}
	return strings.Contains(req.Header.Get("Accept"), "application/json")
}

func normalizeRedirectTarget(candidate string, fallback string) string {
	target := strings.TrimSpace(candidate)
	if target == "" {
		target = fallback
	}
	if target == "" {
		return "/portal"
	}
	return target
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (r *Router) buildAuthSessionFromProfile(profile *models.UserProfile, identity *models.AuthIdentity, session authSessionState) authSessionState {
	if identity != nil {
		session.UserID = identity.UserID
		session.TenantID = identity.TenantID
	}
	if profile != nil {
		session.UserID = profile.ID
		session.TenantID = profile.TenantID
		session.Name = defaultString(profile.DisplayName, session.Name)
		session.Email = defaultString(profile.Email, session.Email)
		session.Locale = profile.Locale
		session.Timezone = profile.Timezone
	}
	return session
}
