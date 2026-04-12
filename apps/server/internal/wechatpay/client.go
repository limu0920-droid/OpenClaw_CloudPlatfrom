package wechatpay

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	StrictMode      bool
	Enabled         bool
	BaseURL         string
	MchID           string
	AppID           string
	ClientSecret    string
	NotifyURL       string
	RefundNotifyURL string
	SerialNo        string
	PublicKeyID     string
	PublicKeyPEM    string
	PrivateKeyPEM   string
	APIv3Key        string
	Mode            string
	SubMchID        string
	SubAppID        string
	HTTPClient      *http.Client
}

type Client struct {
	cfg        Config
	httpClient *http.Client
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type CreatePaymentRequest struct {
	Description string
	OutTradeNo  string
	Amount      int
	Currency    string
	Mode        string // native, h5, jsapi
	OpenID      string
	ClientIP    string
	Attach      string
}

type CreatePaymentResponse struct {
	Mode           string
	CodeURL        string
	H5URL          string
	PrepayID       string
	ChannelOrderNo string
	Raw            map[string]any
}

type RefundRequest struct {
	OutTradeNo   string
	OutRefundNo  string
	Reason       string
	Amount       int
	RefundAmount int
	Currency     string
}

type RefundResponse struct {
	RefundNo        string
	ChannelRefundNo string
	Status          string
	Raw             map[string]any
}

type QueryOrderResponse struct {
	OutTradeNo    string
	TransactionID string
	TradeState    string
	SuccessTime   string
	Raw           map[string]any
}

type NotificationResource struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
	OriginalType   string `json:"original_type"`
}

type NotificationEnvelope struct {
	ID           string               `json:"id"`
	CreateTime   string               `json:"create_time"`
	EventType    string               `json:"event_type"`
	ResourceType string               `json:"resource_type"`
	Summary      string               `json:"summary"`
	Resource     NotificationResource `json:"resource"`
}

func NewClient(cfg Config) (*Client, error) {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	client := &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}

	if strings.TrimSpace(cfg.PrivateKeyPEM) != "" {
		key, err := parsePrivateKey(cfg.PrivateKeyPEM)
		if err != nil {
			return nil, err
		}
		client.privateKey = key
	}
	if strings.TrimSpace(cfg.PublicKeyPEM) != "" {
		key, err := parsePublicKey(cfg.PublicKeyPEM)
		if err != nil {
			return nil, err
		}
		client.publicKey = key
	}

	return client, nil
}

func (c *Client) CreatePayment(ctx context.Context, req CreatePaymentRequest) (CreatePaymentResponse, error) {
	if !c.cfg.Enabled || c.privateKey == nil || c.cfg.MchID == "" {
		return CreatePaymentResponse{}, fmt.Errorf("wechatpay is not configured")
	}

	path, body := c.buildPaymentRequest(req)
	responseBody, err := c.doSignedJSON(ctx, http.MethodPost, path, body)
	if err != nil {
		return CreatePaymentResponse{}, err
	}

	var response map[string]any
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return CreatePaymentResponse{}, err
	}

	out := CreatePaymentResponse{
		Mode:           req.Mode,
		ChannelOrderNo: fmt.Sprintf("%v", response["transaction_id"]),
		Raw:            response,
	}
	if value, ok := response["code_url"].(string); ok {
		out.CodeURL = value
	}
	if value, ok := response["h5_url"].(string); ok {
		out.H5URL = value
	}
	if value, ok := response["prepay_id"].(string); ok {
		out.PrepayID = value
	}

	return out, nil
}

func (c *Client) CreateRefund(ctx context.Context, req RefundRequest) (RefundResponse, error) {
	if !c.cfg.Enabled || c.privateKey == nil || c.cfg.MchID == "" {
		return RefundResponse{}, fmt.Errorf("wechatpay is not configured")
	}

	body := map[string]any{
		"out_trade_no":  req.OutTradeNo,
		"out_refund_no": req.OutRefundNo,
		"reason":        req.Reason,
		"notify_url":    c.cfg.RefundNotifyURL,
		"amount": map[string]any{
			"refund":   req.RefundAmount,
			"total":    req.Amount,
			"currency": defaultString(req.Currency, "CNY"),
		},
	}
	responseBody, err := c.doSignedJSON(ctx, http.MethodPost, "/v3/refund/domestic/refunds", body)
	if err != nil {
		return RefundResponse{}, err
	}

	var response map[string]any
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return RefundResponse{}, err
	}

	return RefundResponse{
		RefundNo:        stringValue(response["out_refund_no"]),
		ChannelRefundNo: stringValue(response["refund_id"]),
		Status:          stringValue(response["status"]),
		Raw:             response,
	}, nil
}

func (c *Client) QueryOrder(ctx context.Context, outTradeNo string) (QueryOrderResponse, error) {
	if !c.cfg.Enabled || c.privateKey == nil || c.cfg.MchID == "" {
		return QueryOrderResponse{}, fmt.Errorf("wechatpay is not configured")
	}

	query := url.Values{}
	query.Set("mchid", c.cfg.MchID)
	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s?%s", url.PathEscape(outTradeNo), query.Encode())
	responseBody, err := c.doSignedRequest(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return QueryOrderResponse{}, err
	}

	var response map[string]any
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return QueryOrderResponse{}, err
	}

	return QueryOrderResponse{
		OutTradeNo:    stringValue(response["out_trade_no"]),
		TransactionID: stringValue(response["transaction_id"]),
		TradeState:    stringValue(response["trade_state"]),
		SuccessTime:   stringValue(response["success_time"]),
		Raw:           response,
	}, nil
}

func (c *Client) CloseOrder(ctx context.Context, outTradeNo string) error {
	if !c.cfg.Enabled || c.privateKey == nil || c.cfg.MchID == "" {
		return fmt.Errorf("wechatpay is not configured")
	}

	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s/close", url.PathEscape(outTradeNo))
	_, err := c.doSignedJSON(ctx, http.MethodPost, path, map[string]any{
		"mchid": c.cfg.MchID,
	})
	return err
}

func (c *Client) buildPaymentRequest(req CreatePaymentRequest) (string, map[string]any) {
	body := map[string]any{
		"appid":        c.cfg.AppID,
		"mchid":        c.cfg.MchID,
		"description":  req.Description,
		"out_trade_no": req.OutTradeNo,
		"notify_url":   c.cfg.NotifyURL,
		"attach":       req.Attach,
		"amount": map[string]any{
			"total":    req.Amount,
			"currency": defaultString(req.Currency, "CNY"),
		},
	}

	switch req.Mode {
	case "jsapi":
		body["payer"] = map[string]any{"openid": req.OpenID}
		return "/v3/pay/transactions/jsapi", body
	case "h5":
		body["scene_info"] = map[string]any{
			"payer_client_ip": defaultString(req.ClientIP, "127.0.0.1"),
			"h5_info": map[string]any{
				"type": "Wap",
			},
		}
		return "/v3/pay/transactions/h5", body
	default:
		return "/v3/pay/transactions/native", body
	}
}

func (c *Client) doSignedJSON(ctx context.Context, method string, path string, body any) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return c.doSignedRequest(ctx, method, path, payload, "application/json")
}

func (c *Client) doSignedRequest(ctx context.Context, method string, path string, payload []byte, contentType string) ([]byte, error) {
	var bodyReader io.Reader
	if len(payload) > 0 {
		bodyReader = bytes.NewReader(payload)
	}

	url := strings.TrimRight(defaultString(c.cfg.BaseURL, "https://api.mch.weixin.qq.com"), "/") + path
	request, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	auth, err := c.authorization(method, path, string(payload))
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", auth)
	request.Header.Set("User-Agent", "openclaw-platform/0.1")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("wechatpay request failed: %s", string(raw))
	}
	return raw, nil
}

func (c *Client) authorization(method string, path string, body string) (string, error) {
	if c.privateKey == nil {
		return "", fmt.Errorf("wechatpay private key is not configured")
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce, err := randomNonce(32)
	if err != nil {
		return "", err
	}
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, path, timestamp, nonce, body)
	hashed := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		c.cfg.MchID,
		nonce,
		base64.StdEncoding.EncodeToString(signature),
		timestamp,
		c.cfg.SerialNo,
	), nil
}

func parsePrivateKey(content string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(content))
	if block == nil {
		return nil, fmt.Errorf("invalid private key pem")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func randomNonce(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer)[:size], nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprintf("%v", value)
}

func (c *Client) VerifyNotification(headers http.Header, body []byte) error {
	if c.publicKey == nil {
		if !c.cfg.Enabled {
			return nil
		}
		return fmt.Errorf("wechatpay public key is not configured")
	}

	timestamp := headers.Get("Wechatpay-Timestamp")
	nonce := headers.Get("Wechatpay-Nonce")
	signature := headers.Get("Wechatpay-Signature")
	serial := headers.Get("Wechatpay-Serial")
	if timestamp == "" || nonce == "" || signature == "" {
		return fmt.Errorf("wechatpay callback headers are incomplete")
	}
	if c.cfg.PublicKeyID != "" && serial != "" && c.cfg.PublicKeyID != serial {
		return fmt.Errorf("wechatpay serial mismatch")
	}

	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	hashed := sha256.Sum256([]byte(message))
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hashed[:], decoded)
}

func (c *Client) DecryptNotificationResource(resource NotificationResource, target any) error {
	if strings.TrimSpace(c.cfg.APIv3Key) == "" {
		return fmt.Errorf("wechatpay APIv3 key is not configured")
	}
	decodedCiphertext, err := base64.StdEncoding.DecodeString(resource.Ciphertext)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher([]byte(c.cfg.APIv3Key))
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	plaintext, err := gcm.Open(nil, []byte(resource.Nonce), decodedCiphertext, []byte(resource.AssociatedData))
	if err != nil {
		return err
	}
	return json.Unmarshal(plaintext, target)
}

func parsePublicKey(content string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(content))
	if block == nil {
		return nil, fmt.Errorf("invalid public key pem")
	}

	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		if key, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return key, nil
		}
	}

	if parsed, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if key, ok := parsed.(*rsa.PublicKey); ok {
			return key, nil
		}
	}

	return nil, fmt.Errorf("unsupported public key content")
}
