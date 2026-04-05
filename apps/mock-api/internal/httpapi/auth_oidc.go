package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"openclaw/mockapi/internal/oidc"
)

type authSessionState struct {
	Provider         string `json:"provider"`
	Authenticated    bool   `json:"authenticated"`
	Name             string `json:"name,omitempty"`
	Email            string `json:"email,omitempty"`
	Role             string `json:"role,omitempty"`
	AccessToken      string `json:"accessToken,omitempty"`
	RefreshToken     string `json:"refreshToken,omitempty"`
	IDToken          string `json:"idToken,omitempty"`
	AccessExpiresAt  string `json:"accessExpiresAt,omitempty"`
	RefreshExpiresAt string `json:"refreshExpiresAt,omitempty"`
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
			"provider":      "mock",
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

	redirectURI := req.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = r.config.KeycloakRedirectURL
	}

	session, err := r.exchangeCodeToSession(req.Context(), code, redirectURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.writeAuthSession(w, session)

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"provider":      session.Provider,
		"redirectTo":    redirectURI,
		"user": map[string]any{
			"name":  session.Name,
			"email": session.Email,
			"role":  session.Role,
		},
	})
}

func (r *Router) handleAuthRefresh(w http.ResponseWriter, req *http.Request) {
	session, ok := r.readAuthSession(req)
	if !ok {
		http.Error(w, "no active session", http.StatusUnauthorized)
		return
	}

	if !r.config.KeycloakEnabled || r.config.KeycloakBaseURL == "" || r.config.KeycloakRealm == "" || r.config.KeycloakClientID == "" {
		session.AccessExpiresAt = time.Now().UTC().Add(30 * time.Minute).Format(time.RFC3339)
		r.writeAuthSession(w, session)
		writeJSON(w, http.StatusOK, session)
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
		return authSessionState{
			Provider:         "mock",
			Authenticated:    true,
			Name:             defaultString(r.config.KeycloakMockUser, "Mock Keycloak User"),
			Email:            defaultString(r.config.KeycloakMockEmail, "mock.keycloak@example.com"),
			Role:             defaultString(r.config.KeycloakMockRole, "tenant_admin"),
			AccessToken:      "mock-access-token",
			RefreshToken:     "mock-refresh-token",
			IDToken:          "mock-id-token",
			AccessExpiresAt:  time.Now().UTC().Add(30 * time.Minute).Format(time.RFC3339),
			RefreshExpiresAt: time.Now().UTC().Add(12 * time.Hour).Format(time.RFC3339),
		}, nil
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

	return authSessionState{
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
	}, nil
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

func (r *Router) sessionSecret() string {
	if strings.TrimSpace(r.config.KeycloakSessionSecret) != "" {
		return r.config.KeycloakSessionSecret
	}
	return "dev-session-secret-change-me"
}

func (r *Router) writeAuthSession(w http.ResponseWriter, session authSessionState) {
	payload, _ := json.Marshal(session)
	raw := base64.RawURLEncoding.EncodeToString(payload)
	signature := r.signSession(raw)
	value := raw + "." + signature

	http.SetCookie(w, &http.Cookie{
		Name:     r.authCookieName(),
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.config.KeycloakCookieSecure,
	})
}

func (r *Router) readAuthSession(req *http.Request) (authSessionState, bool) {
	cookie, err := req.Cookie(r.authCookieName())
	if err != nil || cookie.Value == "" {
		return authSessionState{}, false
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return authSessionState{}, false
	}
	if !hmac.Equal([]byte(parts[1]), []byte(r.signSession(parts[0]))) {
		return authSessionState{}, false
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return authSessionState{}, false
	}

	var session authSessionState
	if err := json.Unmarshal(raw, &session); err != nil {
		return authSessionState{}, false
	}
	return session, true
}

func (r *Router) clearAuthSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     r.authCookieName(),
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

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
