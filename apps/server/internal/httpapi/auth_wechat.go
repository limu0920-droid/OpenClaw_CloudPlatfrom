package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"openclaw/platformapi/internal/wechatauth"
)

func (r *Router) handleAuthProviders(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": []map[string]any{
			{
				"provider": "keycloak",
				"enabled":  r.config.KeycloakEnabled,
				"mode":     "oidc",
			},
			{
				"provider": "wechat",
				"enabled":  r.config.WeChatLoginEnabled,
				"mode":     "website-qr-login",
			},
		},
	})
}

func (r *Router) handleAuthWechatURL(w http.ResponseWriter, req *http.Request) {
	redirectURI := req.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = r.config.WeChatLoginRedirectURL
	}
	if redirectURI == "" {
		http.Error(w, "redirect_uri is required", http.StatusBadRequest)
		return
	}

	flow := authFlowState{
		Provider:    "wechat",
		State:       generateAuthRandom(24),
		Nonce:       generateAuthRandom(16),
		RedirectURI: redirectURI,
		Next:        req.URL.Query().Get("next"),
		ExpiresAt:   time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339),
	}
	r.writeAuthFlow(w, flow)

	if !r.config.WeChatLoginEnabled || r.config.WeChatLoginAppID == "" {
		http.Error(w, "wechat login is not configured", http.StatusServiceUnavailable)
		return
	}

	client := r.newWeChatAuthClient()
	writeJSON(w, http.StatusOK, map[string]any{
		"url":      client.AuthorizeURL(flow.State, redirectURI),
		"provider": "wechat",
		"mode":     "website-qr-login",
		"state":    flow.State,
	})
}

func (r *Router) handleAuthWechatCallback(w http.ResponseWriter, req *http.Request) {
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
	if flow.Provider != "wechat" {
		http.Error(w, "unexpected auth provider", http.StatusBadRequest)
		return
	}
	if state := strings.TrimSpace(req.URL.Query().Get("state")); state == "" || state != flow.State {
		http.Error(w, "invalid auth state", http.StatusBadRequest)
		return
	}

	session, err := r.exchangeWechatCodeToSession(req.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.writeAuthSession(w, session)
	r.clearAuthFlow(w)

	nextTarget := normalizeRedirectTarget(flow.Next, r.defaultPostLoginRedirect())
	if shouldJSONResponse(req) {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": true,
			"provider":      "wechat",
			"redirectTo":    nextTarget,
			"user": map[string]any{
				"name":  session.Name,
				"email": session.Email,
				"role":  session.Role,
			},
		})
		return
	}
	http.Redirect(w, req, nextTarget, http.StatusFound)
}

func (r *Router) exchangeWechatCodeToSession(ctx context.Context, code string) (authSessionState, error) {
	if !r.config.WeChatLoginEnabled || r.config.WeChatLoginAppID == "" || r.config.WeChatLoginAppSecret == "" {
		return authSessionState{}, r.strictModeError("wechat login is not configured")
	}

	client := r.newWeChatAuthClient()
	token, err := client.ExchangeCode(ctx, code)
	if err != nil {
		return authSessionState{}, err
	}
	user, err := client.UserInfo(ctx, token.AccessToken, token.OpenID)
	if err != nil {
		return authSessionState{}, err
	}

	profile, identity := r.resolveUserFromIdentity("wechat", "", token.OpenID, token.UnionID, defaultString(token.UnionID, token.OpenID), defaultString(user.Nickname, "WeChat User"))
	session := authSessionState{
		Provider:         "wechat",
		Authenticated:    true,
		Name:             defaultString(user.Nickname, "WeChat User"),
		Email:            "",
		Role:             defaultString(r.config.KeycloakMockRole, "tenant_admin"),
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		AccessExpiresAt:  time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339),
		RefreshExpiresAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		OpenID:           token.OpenID,
		UnionID:          token.UnionID,
	}
	return r.buildAuthSessionFromProfile(profile, identity, session), nil
}

func (r *Router) newWeChatAuthClient() *wechatauth.Client {
	return wechatauth.NewClient(wechatauth.Config{
		OpenBaseURL: defaultString(r.config.WeChatLoginOpenBaseURL, "https://open.weixin.qq.com"),
		APIBaseURL:  defaultString(r.config.WeChatLoginAPIBaseURL, "https://api.weixin.qq.com"),
		AppID:       r.config.WeChatLoginAppID,
		AppSecret:   r.config.WeChatLoginAppSecret,
		RedirectURL: r.config.WeChatLoginRedirectURL,
	})
}
