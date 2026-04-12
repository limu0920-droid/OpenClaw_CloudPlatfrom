package httpapi

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPortalOrderRefundPreservesRefundNoForCallback(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v3/refund/domestic/refunds" {
			t.Fatalf("unexpected refund path: %s", req.URL.Path)
		}
		if req.Header.Get("Authorization") == "" {
			t.Fatal("expected Authorization header")
		}

		writeGatewayJSON(t, w, map[string]any{
			"out_refund_no": "RF-20260409-999",
			"refund_id":     "502000000009",
			"status":        "PROCESSING",
		})
	}))
	defer gateway.Close()

	router := newTestRouter(ExternalConfig{
		WeChatPayEnabled:       true,
		WeChatPayBaseURL:       gateway.URL,
		WeChatPayMchID:         "1900000109",
		WeChatPayAppID:         "wx-demo-app",
		WeChatPayPrivateKeyPEM: mustPrivateKeyPEM(t),
	})
	router.data.Orders[1].Status = "paid"
	router.data.Payments[1].Status = "paid"

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/orders/2/refunds", map[string]any{
		"amount": 100,
		"reason": "部分退款",
	})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected refund creation status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Refund struct {
			RefundNo string `json:"refundNo"`
			Status   string `json:"status"`
		} `json:"refund"`
		Order struct {
			Status string `json:"status"`
		} `json:"order"`
	}
	decodeResponse(t, recorder, &response)

	if response.Refund.RefundNo == "" {
		t.Fatal("expected refundNo in response")
	}
	if response.Refund.Status != "pending" {
		t.Fatalf("expected refund status pending before callback, got %q", response.Refund.Status)
	}
	if response.Order.Status != "refund_pending" {
		t.Fatalf("expected order status refund_pending, got %q", response.Order.Status)
	}
	if got := router.data.Payments[1].Status; got != "paid" {
		t.Fatalf("expected payment to remain paid before refund callback, got %q", got)
	}

	router.config.WeChatPayEnabled = false
	callback := performRequest(t, router, http.MethodPost, "/api/v1/callback/payment/wechatpay", map[string]any{
		"out_refund_no": response.Refund.RefundNo,
		"refund_id":     "502000000009",
		"refund_status": "SUCCESS",
	})
	if callback.Code != http.StatusOK {
		t.Fatalf("expected callback status 200, got %d: %s", callback.Code, callback.Body.String())
	}

	if got := router.data.Refunds[0].RefundNo; got != response.Refund.RefundNo {
		t.Fatalf("expected stored refundNo %q, got %q", response.Refund.RefundNo, got)
	}
	if got := router.data.Refunds[0].Status; got != "success" {
		t.Fatalf("expected refund status success after callback, got %q", got)
	}
	if got := router.data.Orders[1].Status; got != "refunded" {
		t.Fatalf("expected order status refunded after callback, got %q", got)
	}
	if got := router.data.Payments[1].Status; got != "refunded" {
		t.Fatalf("expected payment status refunded after callback, got %q", got)
	}
}

func TestHandleInternalOrderQueryUsesGatewayResult(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v3/pay/transactions/out-trade-no/ORD-20260405-002" {
			t.Fatalf("unexpected query path: %s", req.URL.Path)
		}
		if got := req.URL.Query().Get("mchid"); got != "1900000109" {
			t.Fatalf("expected mchid query param, got %q", got)
		}
		if req.Header.Get("Authorization") == "" {
			t.Fatal("expected Authorization header")
		}

		writeGatewayJSON(t, w, map[string]any{
			"out_trade_no":   "ORD-20260405-002",
			"transaction_id": "420000000000999",
			"trade_state":    "SUCCESS",
			"success_time":   "2026-04-06T10:00:00Z",
		})
	}))
	defer gateway.Close()

	router := newTestRouter(ExternalConfig{
		WeChatPayEnabled:       true,
		WeChatPayBaseURL:       gateway.URL,
		WeChatPayMchID:         "1900000109",
		WeChatPayAppID:         "wx-demo-app",
		WeChatPayPrivateKeyPEM: mustPrivateKeyPEM(t),
	})
	router.data.Orders[1].Status = "paying"
	router.data.Payments[1].Status = "paying"

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/internal/orders/2/query", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected query status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		TradeState string `json:"tradeState"`
		Source     string `json:"source"`
		Order      struct {
			Status string `json:"status"`
		} `json:"order"`
		Payment struct {
			Status         string `json:"status"`
			ChannelOrderNo string `json:"channelOrderNo"`
		} `json:"payment"`
	}
	decodeResponse(t, recorder, &response)

	if response.TradeState != "SUCCESS" {
		t.Fatalf("expected trade state SUCCESS, got %q", response.TradeState)
	}
	if response.Source != "gateway" {
		t.Fatalf("expected source gateway, got %q", response.Source)
	}
	if response.Order.Status != "paid" {
		t.Fatalf("expected response order status paid, got %q", response.Order.Status)
	}
	if response.Payment.Status != "paid" {
		t.Fatalf("expected response payment status paid, got %q", response.Payment.Status)
	}
	if response.Payment.ChannelOrderNo != "420000000000999" {
		t.Fatalf("expected channel order no from gateway, got %q", response.Payment.ChannelOrderNo)
	}
}

func TestHandleInternalOrderCloseUsesGateway(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v3/pay/transactions/out-trade-no/ORD-20260405-002/close" {
			t.Fatalf("unexpected close path: %s", req.URL.Path)
		}
		if req.Header.Get("Authorization") == "" {
			t.Fatal("expected Authorization header")
		}

		var payload map[string]any
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("decode close body: %v", err)
		}
		if payload["mchid"] != "1900000109" {
			t.Fatalf("expected close body mchid, got %#v", payload["mchid"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer gateway.Close()

	router := newTestRouter(ExternalConfig{
		WeChatPayEnabled:       true,
		WeChatPayBaseURL:       gateway.URL,
		WeChatPayMchID:         "1900000109",
		WeChatPayAppID:         "wx-demo-app",
		WeChatPayPrivateKeyPEM: mustPrivateKeyPEM(t),
	})
	router.data.Orders[1].Status = "paying"
	router.data.Payments[1].Status = "paying"

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/internal/orders/2/close", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected close status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if got := router.data.Orders[1].Status; got != "closed" {
		t.Fatalf("expected order status closed, got %q", got)
	}
	if got := router.data.Payments[1].Status; got != "closed" {
		t.Fatalf("expected payment status closed, got %q", got)
	}
}

func newTestRouter(cfg ExternalConfig) *Router {
	return &Router{
		data:    mockdata.Seed(),
		config:  cfg,
		runtime: newTestRuntimeAdapter(),
	}
}

func performRequest(t *testing.T, handler http.Handler, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	switch value := body.(type) {
	case nil:
		payload = nil
	case []byte:
		payload = value
	case string:
		payload = []byte(value)
	default:
		var err error
		payload, err = json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v\nbody: %s", err, recorder.Body.String())
	}
}

func mustPrivateKeyPEM(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: encoded,
	}))
}

func writeGatewayJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal gateway payload: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.WriteString(w, string(raw)); err != nil {
		t.Fatalf("write gateway payload: %v", err)
	}
}
