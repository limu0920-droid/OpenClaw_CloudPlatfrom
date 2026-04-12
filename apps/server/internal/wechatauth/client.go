package wechatauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	OpenBaseURL string
	APIBaseURL  string
	AppID       string
	AppSecret   string
	RedirectURL string
	HTTPClient  *http.Client
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

type UserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
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

func (c *Client) AuthorizeURL(state string, redirectURI string) string {
	base := strings.TrimRight(defaultString(c.cfg.OpenBaseURL, "https://open.weixin.qq.com"), "/")
	redirect := redirectURI
	if redirect == "" {
		redirect = c.cfg.RedirectURL
	}

	query := url.Values{}
	query.Set("appid", c.cfg.AppID)
	query.Set("redirect_uri", redirect)
	query.Set("response_type", "code")
	query.Set("scope", "snsapi_login")
	query.Set("state", state)

	return fmt.Sprintf("%s/connect/qrconnect?%s#wechat_redirect", base, query.Encode())
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (TokenResponse, error) {
	base := strings.TrimRight(defaultString(c.cfg.APIBaseURL, "https://api.weixin.qq.com"), "/")
	query := url.Values{}
	query.Set("appid", c.cfg.AppID)
	query.Set("secret", c.cfg.AppSecret)
	query.Set("code", code)
	query.Set("grant_type", "authorization_code")

	url := fmt.Sprintf("%s/sns/oauth2/access_token?%s", base, query.Encode())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return TokenResponse{}, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return TokenResponse{}, err
	}
	defer response.Body.Close()

	var token TokenResponse
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return TokenResponse{}, err
	}
	return token, nil
}

func (c *Client) UserInfo(ctx context.Context, accessToken string, openID string) (UserInfo, error) {
	base := strings.TrimRight(defaultString(c.cfg.APIBaseURL, "https://api.weixin.qq.com"), "/")
	query := url.Values{}
	query.Set("access_token", accessToken)
	query.Set("openid", openID)
	query.Set("lang", "zh_CN")

	url := fmt.Sprintf("%s/sns/userinfo?%s", base, query.Encode())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return UserInfo{}, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return UserInfo{}, err
	}
	defer response.Body.Close()

	var user UserInfo
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return UserInfo{}, err
	}
	return user, nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
