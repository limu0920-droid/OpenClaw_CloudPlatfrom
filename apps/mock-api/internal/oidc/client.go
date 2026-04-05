package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HTTPClient   *http.Client
}

type Discovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

type UserInfo struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	Email             string `json:"email"`
}

type Client struct {
	cfg        Config
	discovery  Discovery
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}

	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (c *Client) Discover(ctx context.Context) (Discovery, error) {
	if c.discovery.TokenEndpoint != "" {
		return c.discovery, nil
	}

	base := strings.TrimRight(c.cfg.BaseURL, "/")
	discoveryURL := fmt.Sprintf("%s/realms/%s/.well-known/openid-configuration", base, url.PathEscape(c.cfg.Realm))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return Discovery{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.standardDiscovery(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return c.standardDiscovery(), nil
	}

	var d Discovery
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return Discovery{}, err
	}

	c.discovery = d
	return d, nil
}

func (c *Client) ExchangeCode(ctx context.Context, code string, redirectURI string) (TokenResponse, error) {
	d, err := c.Discover(ctx)
	if err != nil {
		return TokenResponse{}, err
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.cfg.ClientID)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	if c.cfg.ClientSecret != "" {
		form.Set("client_secret", c.cfg.ClientSecret)
	}

	return c.postTokenForm(ctx, d.TokenEndpoint, form)
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (TokenResponse, error) {
	d, err := c.Discover(ctx)
	if err != nil {
		return TokenResponse{}, err
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", c.cfg.ClientID)
	form.Set("refresh_token", refreshToken)
	if c.cfg.ClientSecret != "" {
		form.Set("client_secret", c.cfg.ClientSecret)
	}

	return c.postTokenForm(ctx, d.TokenEndpoint, form)
}

func (c *Client) UserInfo(ctx context.Context, accessToken string) (UserInfo, error) {
	d, err := c.Discover(ctx)
	if err != nil {
		return UserInfo{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.UserinfoEndpoint, nil)
	if err != nil {
		return UserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return UserInfo{}, fmt.Errorf("userinfo failed: %s", string(raw))
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return UserInfo{}, err
	}
	return info, nil
}

func (c *Client) LogoutURL(ctx context.Context, idTokenHint string, postLogoutRedirectURI string) (string, error) {
	d, err := c.Discover(ctx)
	if err != nil {
		return "", err
	}

	endpoint := d.EndSessionEndpoint
	if endpoint == "" {
		endpoint = c.standardDiscovery().EndSessionEndpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	if idTokenHint != "" {
		q.Set("id_token_hint", idTokenHint)
	}
	if postLogoutRedirectURI != "" {
		q.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}
	if c.cfg.ClientID != "" {
		q.Set("client_id", c.cfg.ClientID)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) postTokenForm(ctx context.Context, endpoint string, form url.Values) (TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return TokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return TokenResponse{}, fmt.Errorf("token exchange failed: %s", string(raw))
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return TokenResponse{}, err
	}
	return token, nil
}

func (c *Client) standardDiscovery() Discovery {
	base := strings.TrimRight(c.cfg.BaseURL, "/")
	realmBase := fmt.Sprintf("%s/realms/%s/protocol/openid-connect", base, url.PathEscape(c.cfg.Realm))
	return Discovery{
		AuthorizationEndpoint: realmBase + "/auth",
		TokenEndpoint:         realmBase + "/token",
		UserinfoEndpoint:      realmBase + "/userinfo",
		EndSessionEndpoint:    realmBase + "/logout",
	}
}
