package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/wechatpay"
)

type createOrderRequest struct {
	PlanCode   string `json:"planCode"`
	InstanceID int    `json:"instanceId,omitempty"`
	OrderType  string `json:"orderType"` // buy, renew, upgrade
}

type createPaymentRequest struct {
	Channel  string `json:"channel"` // wechatpay
	PayMode  string `json:"payMode"` // native, h5, jsapi
	OpenID   string `json:"openId,omitempty"`
	ClientIP string `json:"clientIp,omitempty"`
}

type refundPaymentRequest struct {
	Amount int    `json:"amount"`
	Reason string `json:"reason"`
}

type createInvoiceRequest struct {
	OrderID     int    `json:"orderId"`
	InvoiceType string `json:"invoiceType"`
	Title       string `json:"title"`
	TaxNo       string `json:"taxNo"`
	Email       string `json:"email"`
}

type updateInvoiceStatusRequest struct {
	Status string `json:"status"`
	PDFURL string `json:"pdfUrl"`
}

func (r *Router) handlePortalOrders(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	items := make([]models.Order, 0)
	for _, item := range r.data.Orders {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handlePortalSubscriptions(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	items := make([]models.Subscription, 0)
	for _, item := range r.data.Subscriptions {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handlePortalInvoices(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantID := tenantFilterID(req, 1)
	items := make([]models.InvoiceRecord, 0)
	for _, item := range r.data.Invoices {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handlePortalCreateInvoice(w http.ResponseWriter, req *http.Request) {
	var payload createInvoiceRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if payload.OrderID <= 0 || strings.TrimSpace(payload.Title) == "" {
		http.Error(w, "orderId and title are required", http.StatusBadRequest)
		return
	}
	if payload.InvoiceType == "" {
		payload.InvoiceType = "vat_normal"
	}

	r.mu.Lock()

	order, found := r.findOrder(payload.OrderID)
	if !found {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	invoiceID := r.nextInvoiceID()
	now := nowRFC3339()
	invoice := models.InvoiceRecord{
		ID:          invoiceID,
		TenantID:    order.TenantID,
		OrderID:     order.ID,
		InvoiceType: payload.InvoiceType,
		Status:      "pending",
		Amount:      order.Amount,
		Title:       payload.Title,
		TaxNo:       payload.TaxNo,
		Email:       payload.Email,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.data.Invoices = append([]models.InvoiceRecord{invoice}, r.data.Invoices...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist invoice failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"invoice": invoice})
}

func (r *Router) handlePortalCreateOrder(w http.ResponseWriter, req *http.Request) {
	var payload createOrderRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.PlanCode) == "" {
		http.Error(w, "planCode is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.OrderType) == "" {
		payload.OrderType = "buy"
	}

	r.mu.Lock()

	offer := r.findPlanOffer(payload.PlanCode)
	if offer == nil {
		r.mu.Unlock()
		http.Error(w, "plan not found", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	orderID := r.nextOrderID()
	tenantID := tenantFilterID(req, 1)
	order := models.Order{
		ID:         orderID,
		TenantID:   tenantID,
		InstanceID: payload.InstanceID,
		PlanCode:   payload.PlanCode,
		Action:     payload.OrderType,
		Status:     "pending",
		Amount:     offer.MonthlyPrice,
		Currency:   "CNY",
		OrderNo:    fmt.Sprintf("ORD-%s-%03d", time.Now().UTC().Format("20060102"), orderID),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.data.Orders = append([]models.Order{order}, r.data.Orders...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist order failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"order": order,
		"plan":  offer,
	})
}

func (r *Router) handlePortalOrderDetail(w http.ResponseWriter, req *http.Request) {
	orderID, ok := parseTailID(req.URL.Path, "/api/v1/portal/orders/")
	if !ok {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	order, found := r.findOrder(orderID)
	if !found {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"order":    order,
		"payments": r.filterPaymentsByOrder(orderID),
		"refunds":  r.filterRefundsByOrder(orderID),
	})
}

func (r *Router) handlePortalOrderPay(w http.ResponseWriter, req *http.Request) {
	orderID, ok := parseTailID(req.URL.Path, "/api/v1/portal/orders/", "/pay")
	if !ok {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	var payload createPaymentRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if payload.Channel == "" {
		payload.Channel = "wechatpay"
	}
	if payload.PayMode == "" {
		payload.PayMode = "native"
	}

	r.mu.Lock()

	order, index := r.findOrderIndex(orderID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	paymentID := r.nextPaymentID()

	paymentResp, err := r.createPaymentIntent(req.Context(), order, payload)
	if err != nil {
		r.mu.Unlock()
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	payment := models.PaymentTransaction{
		ID:             paymentID,
		OrderID:        order.ID,
		Channel:        payload.Channel,
		PayMode:        payload.PayMode,
		TradeNo:        fmt.Sprintf("PAY-%s-%03d", time.Now().UTC().Format("20060102"), paymentID),
		ChannelOrderNo: paymentResp.ChannelOrderNo,
		Amount:         order.Amount,
		Currency:       order.Currency,
		Status:         "paying",
		PayURL:         paymentResp.H5URL,
		CodeURL:        paymentResp.CodeURL,
		PrepayID:       paymentResp.PrepayID,
		AppID:          r.config.WeChatPayAppID,
		MchID:          r.config.WeChatPayMchID,
		CreatedAt:      now,
		UpdatedAt:      now,
		Raw:            paymentResp.Raw,
	}

	r.data.Orders[index].Status = "paying"
	r.data.Orders[index].UpdatedAt = now
	r.data.Payments = append([]models.PaymentTransaction{payment}, r.data.Payments...)
	responseOrder := r.data.Orders[index]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist payment intent failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"order":   responseOrder,
		"payment": payment,
	})
}

func (r *Router) handlePortalSubscriptionAction(w http.ResponseWriter, req *http.Request) {
	action := "upgrade"
	if strings.HasSuffix(req.URL.Path, "/renew") {
		action = "renew"
	}
	subscriptionID, ok := parseTailID(req.URL.Path, "/api/v1/portal/subscriptions/", "/"+action)
	if !ok {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	subscription, found := r.findSubscription(subscriptionID)
	if !found {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}
	offer := r.findPlanOffer(subscription.PlanCode)
	if offer == nil {
		r.mu.Unlock()
		http.Error(w, "plan not found", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	orderID := r.nextOrderID()
	order := models.Order{
		ID:         orderID,
		TenantID:   subscription.TenantID,
		InstanceID: subscription.InstanceID,
		PlanCode:   subscription.PlanCode,
		Action:     action,
		Status:     "pending",
		Amount:     offer.MonthlyPrice,
		Currency:   "CNY",
		OrderNo:    fmt.Sprintf("ORD-%s-%03d", time.Now().UTC().Format("20060102"), orderID),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.data.Orders = append([]models.Order{order}, r.data.Orders...)
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist subscription action failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"subscription": subscription,
		"order":        order,
	})
}

func (r *Router) handlePortalOrderRefund(w http.ResponseWriter, req *http.Request) {
	orderID, ok := parseTailID(req.URL.Path, "/api/v1/portal/orders/", "/refunds")
	if !ok {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	var payload refundPaymentRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	order, orderIndex := r.findOrderIndex(orderID)
	if orderIndex < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}
	payment, paymentIndex := r.findLatestPaymentByOrder(orderID)
	if paymentIndex < 0 {
		r.mu.Unlock()
		http.Error(w, "payment not found", http.StatusBadRequest)
		return
	}

	if payload.Amount <= 0 {
		payload.Amount = payment.Amount
	}
	if payload.Amount > payment.Amount {
		http.Error(w, "refund amount exceeds paid amount", http.StatusBadRequest)
		return
	}
	if payload.Reason == "" {
		payload.Reason = "用户申请退款"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	refundID := r.nextRefundID()
	refundNo := fmt.Sprintf("RF-%s-%03d", time.Now().UTC().Format("20060102"), refundID)
	resp, err := r.createRefund(req.Context(), order, payment, refundNo, payload)
	if err != nil {
		r.mu.Unlock()
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	refund := models.RefundRecord{
		ID:              refundID,
		OrderID:         orderID,
		PaymentID:       payment.ID,
		RefundNo:        refundNo,
		ChannelRefundNo: resp.ChannelRefundNo,
		Status:          mapRefundStatus(resp.Status),
		Amount:          payload.Amount,
		Reason:          payload.Reason,
		NotifyURL:       r.config.WeChatPayRefundNotifyURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	r.data.Refunds = append([]models.RefundRecord{refund}, r.data.Refunds...)
	switch refund.Status {
	case "success":
		r.applyRefundSuccess(r.data.Refunds[0], 0, defaultString(resp.ChannelRefundNo, refund.ChannelRefundNo), now)
	case "pending":
		r.data.Orders[orderIndex].Status = "refund_pending"
		r.data.Orders[orderIndex].UpdatedAt = now
	default:
		r.data.Payments[paymentIndex].UpdatedAt = now
	}
	responseRefund := r.data.Refunds[0]
	responseOrder := r.data.Orders[orderIndex]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist refund failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"refund": responseRefund,
		"order":  responseOrder,
	})
}

func (r *Router) handleAdminPayments(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Payments})
}

func (r *Router) handleAdminPaymentCallbackEvents(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.PaymentCallbackEvents})
}

func (r *Router) handleAdminRefunds(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Refunds})
}

func (r *Router) handleAdminOrders(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Orders})
}

func (r *Router) handleAdminSubscriptions(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Subscriptions})
}

func (r *Router) handleAdminInvoices(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": r.data.Invoices})
}

func (r *Router) handleAdminInvoiceStatus(w http.ResponseWriter, req *http.Request) {
	invoiceID, ok := parseTailID(req.URL.Path, "/api/v1/admin/invoices/", "/status")
	if !ok {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}

	var payload updateInvoiceStatusRequest
	if err := decodeJSON(req, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	r.mu.Lock()

	index := r.findInvoiceIndex(invoiceID)
	if index < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}

	if payload.Status != "" {
		r.data.Invoices[index].Status = payload.Status
	}
	if payload.PDFURL != "" {
		r.data.Invoices[index].PDFURL = payload.PDFURL
	}
	if payload.Status == "issued" && r.data.Invoices[index].InvoiceNo == "" {
		r.data.Invoices[index].InvoiceNo = fmt.Sprintf("INV-%s-%03d", time.Now().UTC().Format("20060102"), invoiceID)
	}
	r.data.Invoices[index].UpdatedAt = nowRFC3339()
	invoice := r.data.Invoices[index]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist invoice status failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"invoice": invoice})
}

func (r *Router) handleWechatPayCallback(w http.ResponseWriter, req *http.Request) {
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read callback body", http.StatusBadRequest)
		return
	}

	client, err := r.newWeChatPayClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	signatureStatus := "skipped"
	if err := client.VerifyNotification(req.Header, raw); err != nil && r.config.WeChatPayEnabled {
		r.mu.Lock()
		r.recordPaymentCallbackEvent(models.PaymentCallbackEvent{
			Channel:         "wechatpay",
			EventType:       "",
			SignatureStatus: "failed",
			DecryptStatus:   "skipped",
			ProcessStatus:   "failed",
			RequestSerial:   req.Header.Get("Wechatpay-Serial"),
			CreatedAt:       nowRFC3339(),
			RawBody:         string(raw),
		})
		r.mu.Unlock()
		_ = r.persistAllData()
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	} else if err == nil {
		signatureStatus = "verified"
	}

	var envelope wechatpay.NotificationEnvelope
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &envelope); err != nil {
			r.mu.Lock()
			r.recordPaymentCallbackEvent(models.PaymentCallbackEvent{
				Channel:         "wechatpay",
				EventType:       "",
				SignatureStatus: signatureStatus,
				DecryptStatus:   "skipped",
				ProcessStatus:   "failed",
				RequestSerial:   req.Header.Get("Wechatpay-Serial"),
				CreatedAt:       nowRFC3339(),
				RawBody:         string(raw),
			})
			r.mu.Unlock()
			_ = r.persistAllData()
			http.Error(w, "invalid callback payload", http.StatusBadRequest)
			return
		}
	}

	var transaction struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Amount        struct {
			Total int `json:"total"`
		} `json:"amount"`
	}

	var refund struct {
		OutRefundNo string `json:"out_refund_no"`
		RefundID    string `json:"refund_id"`
		Status      string `json:"refund_status"`
		SuccessTime string `json:"success_time"`
	}
	decryptStatus := "skipped"

	if envelope.Resource.Ciphertext != "" && strings.TrimSpace(r.config.WeChatPayAPIv3Key) != "" {
		if envelope.Resource.OriginalType == "refund" || strings.Contains(strings.ToLower(envelope.EventType), "refund") {
			if err := client.DecryptNotificationResource(envelope.Resource, &refund); err != nil {
				r.mu.Lock()
				r.recordPaymentCallbackEvent(models.PaymentCallbackEvent{
					Channel:         "wechatpay",
					EventType:       envelope.EventType,
					SignatureStatus: signatureStatus,
					DecryptStatus:   "failed",
					ProcessStatus:   "failed",
					RequestSerial:   req.Header.Get("Wechatpay-Serial"),
					CreatedAt:       nowRFC3339(),
					RawBody:         string(raw),
				})
				r.mu.Unlock()
				_ = r.persistAllData()
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			if err := client.DecryptNotificationResource(envelope.Resource, &transaction); err != nil {
				r.mu.Lock()
				r.recordPaymentCallbackEvent(models.PaymentCallbackEvent{
					Channel:         "wechatpay",
					EventType:       envelope.EventType,
					SignatureStatus: signatureStatus,
					DecryptStatus:   "failed",
					ProcessStatus:   "failed",
					RequestSerial:   req.Header.Get("Wechatpay-Serial"),
					CreatedAt:       nowRFC3339(),
					RawBody:         string(raw),
				})
				r.mu.Unlock()
				_ = r.persistAllData()
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		decryptStatus = "success"
	} else if len(raw) > 0 {
		_ = json.Unmarshal(raw, &transaction)
		_ = json.Unmarshal(raw, &refund)
	}

	r.mu.Lock()

	now := nowRFC3339()

	if transaction.OutTradeNo != "" {
		order, orderIndex := r.findOrderByOrderNo(transaction.OutTradeNo)
		if orderIndex < 0 {
			r.mu.Unlock()
			http.NotFound(w, req)
			return
		}

		payment, paymentIndex := r.findLatestPaymentByOrder(order.ID)
		if paymentIndex >= 0 {
			r.applyPaymentSuccess(order, orderIndex, payment, paymentIndex, defaultString(transaction.TransactionID, payment.ChannelOrderNo), defaultString(transaction.SuccessTime, now))
		}
		r.recordPaymentCallbackEvent(models.PaymentCallbackEvent{
			Channel:         "wechatpay",
			EventType:       defaultString(envelope.EventType, "TRANSACTION.SUCCESS"),
			OutTradeNo:      transaction.OutTradeNo,
			SignatureStatus: signatureStatus,
			DecryptStatus:   decryptStatus,
			ProcessStatus:   "succeeded",
			RequestSerial:   req.Header.Get("Wechatpay-Serial"),
			CreatedAt:       now,
			RawBody:         string(raw),
		})
		responseOrder := r.data.Orders[orderIndex]
		r.mu.Unlock()
		if err := r.persistAllData(); err != nil {
			http.Error(w, "persist payment callback failed", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"accepted": true,
			"type":     "transaction",
			"order":    responseOrder,
		})
		return
	}

	if refund.OutRefundNo != "" {
		refundRecord, refundIndex := r.findRefundByRefundNo(refund.OutRefundNo)
		if refundIndex < 0 {
			r.mu.Unlock()
			http.NotFound(w, req)
			return
		}

		r.applyRefundSuccess(refundRecord, refundIndex, defaultString(refund.RefundID, refundRecord.ChannelRefundNo), now)
		r.recordPaymentCallbackEvent(models.PaymentCallbackEvent{
			Channel:         "wechatpay",
			EventType:       defaultString(envelope.EventType, "REFUND.SUCCESS"),
			OutRefundNo:     refund.OutRefundNo,
			SignatureStatus: signatureStatus,
			DecryptStatus:   decryptStatus,
			ProcessStatus:   "succeeded",
			RequestSerial:   req.Header.Get("Wechatpay-Serial"),
			CreatedAt:       now,
			RawBody:         string(raw),
		})
		responseRefund := r.data.Refunds[refundIndex]
		r.mu.Unlock()
		if err := r.persistAllData(); err != nil {
			http.Error(w, "persist refund callback failed", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"accepted": true,
			"type":     "refund",
			"refund":   responseRefund,
		})
		return
	}
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist callback noop failed", http.StatusInternalServerError)
		return
	}

	if r.strictModeEnabled() {
		http.Error(w, "unsupported wechatpay callback payload", http.StatusBadRequest)
		return
	}
	http.Error(w, "unsupported wechatpay callback payload", http.StatusBadRequest)
}

func (r *Router) handleInternalOrderClose(w http.ResponseWriter, req *http.Request) {
	orderID, ok := parseTailID(req.URL.Path, "/api/v1/internal/orders/", "/close")
	if !ok {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	order, orderIndex := r.findOrderIndex(orderID)
	if orderIndex < 0 {
		r.mu.RUnlock()
		http.NotFound(w, req)
		return
	}
	payment, paymentIndex := r.findLatestPaymentByOrder(orderID)
	r.mu.RUnlock()

	if order.Status == "paid" || order.Status == "active" || order.Status == "refunded" {
		http.Error(w, "paid or refunded order cannot be closed", http.StatusConflict)
		return
	}

	if paymentIndex >= 0 && payment.Channel == "wechatpay" && r.config.WeChatPayEnabled {
		client, err := r.newWeChatPayClient()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if err := client.CloseOrder(req.Context(), order.OrderNo); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	} else if paymentIndex >= 0 && payment.Channel == "wechatpay" {
		http.Error(w, "wechatpay gateway is required", http.StatusServiceUnavailable)
		return
	}

	r.mu.Lock()

	order, orderIndex = r.findOrderIndex(orderID)
	if orderIndex < 0 {
		r.mu.Unlock()
		http.NotFound(w, req)
		return
	}
	if order.Status == "paid" || order.Status == "active" || order.Status == "refunded" {
		r.mu.Unlock()
		http.Error(w, "paid or refunded order cannot be closed", http.StatusConflict)
		return
	}

	now := nowRFC3339()
	r.applyOrderClosed(order, orderIndex, now)
	responseOrder := r.data.Orders[orderIndex]
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist order close failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": responseOrder})
}

func (r *Router) handleInternalOrderQuery(w http.ResponseWriter, req *http.Request) {
	orderID, ok := parseTailID(req.URL.Path, "/api/v1/internal/orders/", "/query")
	if !ok {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	order, orderIndex := r.findOrderIndex(orderID)
	if orderIndex < 0 {
		r.mu.RUnlock()
		http.NotFound(w, req)
		return
	}
	payment, paymentIndex := r.findLatestPaymentByOrder(orderID)
	r.mu.RUnlock()

	tradeState := "NOTPAY"
	source := "gateway"
	now := nowRFC3339()

	if paymentIndex >= 0 && payment.Channel == "wechatpay" && r.config.WeChatPayEnabled {
		client, err := r.newWeChatPayClient()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		result, err := client.QueryOrder(req.Context(), order.OrderNo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		tradeState = defaultString(result.TradeState, "NOTPAY")
		source = "gateway"

		r.mu.Lock()
		if currentOrder, currentOrderIndex := r.findOrderIndex(orderID); currentOrderIndex >= 0 {
			if currentPayment, currentPaymentIndex := r.findLatestPaymentByOrder(orderID); currentPaymentIndex >= 0 {
				switch tradeState {
				case "SUCCESS":
					if currentOrder.Status != "paid" && currentOrder.Status != "active" {
						r.applyPaymentSuccess(currentOrder, currentOrderIndex, currentPayment, currentPaymentIndex, defaultString(result.TransactionID, currentPayment.ChannelOrderNo), defaultString(result.SuccessTime, now))
					}
				case "CLOSED":
					r.applyOrderClosed(currentOrder, currentOrderIndex, now)
				case "PAYERROR":
					r.applyPaymentFailure(currentOrder, currentOrderIndex, currentPayment, currentPaymentIndex, now)
				}
			}
			order, _ = r.findOrder(orderID)
			payment, _ = r.findLatestPaymentByOrder(orderID)
		}
		r.mu.Unlock()
	} else if paymentIndex >= 0 && payment.Channel == "wechatpay" {
		http.Error(w, "wechatpay gateway is required", http.StatusServiceUnavailable)
		return
	}
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist order query failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"order":      order,
		"payment":    payment,
		"tradeState": tradeState,
		"checkedAt":  now,
		"source":     source,
	})
}

func (r *Router) handleInternalPaymentsReconcile(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()

	now := nowRFC3339()
	result := map[string]any{
		"checkedOrders":   len(r.data.Orders),
		"checkedPayments": len(r.data.Payments),
		"checkedRefunds":  len(r.data.Refunds),
		"callbackEvents":  len(r.data.PaymentCallbackEvents),
		"inconsistencies": []string{},
		"recovered":       []string{},
	}

	inconsistencies := make([]string, 0)
	recovered := make([]string, 0)
	for orderIndex, order := range r.data.Orders {
		payment, paymentIndex := r.findLatestPaymentByOrder(order.ID)
		if paymentIndex < 0 {
			continue
		}
		if payment.Status == "paid" && order.Status != "paid" && order.Status != "active" {
			inconsistencies = append(inconsistencies, fmt.Sprintf("order %s expected paid, actual %s", order.OrderNo, order.Status))
			r.applyPaymentSuccess(order, orderIndex, payment, paymentIndex, defaultString(payment.ChannelOrderNo, "reconcile-"+payment.TradeNo), now)
			recovered = append(recovered, order.OrderNo)
		}
		if payment.Status == "refunded" && order.Status != "refunded" {
			inconsistencies = append(inconsistencies, fmt.Sprintf("order %s expected refunded, actual %s", order.OrderNo, order.Status))
			r.data.Orders[orderIndex].Status = "refunded"
			r.data.Orders[orderIndex].UpdatedAt = now
			recovered = append(recovered, "refund:"+order.OrderNo)
		}
	}
	result["inconsistencies"] = inconsistencies
	result["recovered"] = recovered
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist reconcile failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"job": map[string]any{
			"type":      "payment_reconcile",
			"status":    "succeeded",
			"channel":   "wechatpay",
			"createdAt": now,
		},
		"result": result,
	})
}

func (r *Router) handleInternalPaymentsRecover(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()

	now := nowRFC3339()
	recovered := make([]string, 0)
	for orderIndex, order := range r.data.Orders {
		payment, paymentIndex := r.findLatestPaymentByOrder(order.ID)
		if paymentIndex < 0 {
			continue
		}
		if payment.Status == "paying" && payment.ChannelOrderNo != "" && !r.hasPaymentCallbackForOrder(order.OrderNo) {
			r.applyPaymentSuccess(order, orderIndex, payment, paymentIndex, payment.ChannelOrderNo, now)
			r.recordPaymentCallbackEvent(models.PaymentCallbackEvent{
				Channel:         payment.Channel,
				EventType:       "TRANSACTION.SUCCESS",
				OutTradeNo:      order.OrderNo,
				SignatureStatus: "skipped",
				DecryptStatus:   "skipped",
				ProcessStatus:   "recovered",
				RequestSerial:   "recover-job",
				CreatedAt:       now,
				RawBody:         "{\"recovered\":true}",
			})
			recovered = append(recovered, order.OrderNo)
		}
	}
	r.mu.Unlock()
	if err := r.persistAllData(); err != nil {
		http.Error(w, "persist recover failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"job": map[string]any{
			"type":      "payment_recover_missing_callback",
			"status":    "succeeded",
			"channel":   "wechatpay",
			"createdAt": now,
		},
		"recovered": recovered,
	})
}

func (r *Router) newWeChatPayClient() (*wechatpay.Client, error) {
	return wechatpay.NewClient(wechatpay.Config{
		StrictMode:      r.config.StrictMode,
		Enabled:         r.config.WeChatPayEnabled,
		BaseURL:         defaultString(r.config.WeChatPayBaseURL, "https://api.mch.weixin.qq.com"),
		MchID:           r.config.WeChatPayMchID,
		AppID:           r.config.WeChatPayAppID,
		ClientSecret:    r.config.WeChatPayClientSecret,
		NotifyURL:       defaultString(r.config.WeChatPayNotifyURL, "http://localhost:8080/api/v1/callback/payment/wechatpay"),
		RefundNotifyURL: defaultString(r.config.WeChatPayRefundNotifyURL, "http://localhost:8080/api/v1/callback/payment/wechatpay"),
		SerialNo:        r.config.WeChatPaySerialNo,
		PublicKeyID:     r.config.WeChatPayPublicKeyID,
		PublicKeyPEM:    r.config.WeChatPayPublicKeyPEM,
		PrivateKeyPEM:   r.config.WeChatPayPrivateKeyPEM,
		APIv3Key:        r.config.WeChatPayAPIv3Key,
		Mode:            defaultString(r.config.WeChatPayMode, "merchant"),
		SubMchID:        r.config.WeChatPaySubMchID,
		SubAppID:        r.config.WeChatPaySubAppID,
	})
}

func (r *Router) createPaymentIntent(ctx context.Context, order models.Order, payload createPaymentRequest) (wechatpay.CreatePaymentResponse, error) {
	client, err := r.newWeChatPayClient()
	if err != nil {
		return wechatpay.CreatePaymentResponse{}, err
	}

	return client.CreatePayment(ctx, wechatpay.CreatePaymentRequest{
		Description: fmt.Sprintf("%s %s", order.Action, order.PlanCode),
		OutTradeNo:  order.OrderNo,
		Amount:      order.Amount,
		Currency:    order.Currency,
		Mode:        payload.PayMode,
		OpenID:      payload.OpenID,
		ClientIP:    payload.ClientIP,
		Attach:      fmt.Sprintf("tenant=%d", order.TenantID),
	})
}

func (r *Router) createRefund(ctx context.Context, order models.Order, payment models.PaymentTransaction, refundNo string, payload refundPaymentRequest) (wechatpay.RefundResponse, error) {
	client, err := r.newWeChatPayClient()
	if err != nil {
		return wechatpay.RefundResponse{}, err
	}

	return client.CreateRefund(ctx, wechatpay.RefundRequest{
		OutTradeNo:   order.OrderNo,
		OutRefundNo:  refundNo,
		Reason:       payload.Reason,
		Amount:       payment.Amount,
		RefundAmount: payload.Amount,
		Currency:     payment.Currency,
	})
}

func (r *Router) findOrder(id int) (models.Order, bool) {
	for _, item := range r.data.Orders {
		if item.ID == id {
			return item, true
		}
	}
	return models.Order{}, false
}

func (r *Router) findOrderIndex(id int) (models.Order, int) {
	for index, item := range r.data.Orders {
		if item.ID == id {
			return item, index
		}
	}
	return models.Order{}, -1
}

func (r *Router) findOrderByOrderNo(orderNo string) (models.Order, int) {
	for index, item := range r.data.Orders {
		if item.OrderNo == orderNo {
			return item, index
		}
	}
	return models.Order{}, -1
}

func (r *Router) filterPaymentsByOrder(orderID int) []models.PaymentTransaction {
	out := make([]models.PaymentTransaction, 0)
	for _, item := range r.data.Payments {
		if item.OrderID == orderID {
			out = append(out, item)
		}
	}
	return out
}

func (r *Router) findLatestPaymentByOrder(orderID int) (models.PaymentTransaction, int) {
	for index, item := range r.data.Payments {
		if item.OrderID == orderID {
			return item, index
		}
	}
	return models.PaymentTransaction{}, -1
}

func (r *Router) filterRefundsByOrder(orderID int) []models.RefundRecord {
	out := make([]models.RefundRecord, 0)
	for _, item := range r.data.Refunds {
		if item.OrderID == orderID {
			out = append(out, item)
		}
	}
	return out
}

func (r *Router) findSubscription(id int) (models.Subscription, bool) {
	for _, item := range r.data.Subscriptions {
		if item.ID == id {
			return item, true
		}
	}
	return models.Subscription{}, false
}

func (r *Router) findSubscriptionIndexByInstance(instanceID int) int {
	for index, item := range r.data.Subscriptions {
		if item.InstanceID == instanceID {
			return index
		}
	}
	return -1
}

func (r *Router) nextPaymentID() int {
	maxID := 0
	for _, item := range r.data.Payments {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextRefundID() int {
	maxID := 0
	for _, item := range r.data.Refunds {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) nextInvoiceID() int {
	maxID := 0
	for _, item := range r.data.Invoices {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (r *Router) findInvoiceIndex(id int) int {
	for index, item := range r.data.Invoices {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (r *Router) findInvoiceIndexByOrder(orderID int) int {
	for index, item := range r.data.Invoices {
		if item.OrderID == orderID {
			return index
		}
	}
	return -1
}

func (r *Router) findRefundByRefundNo(refundNo string) (models.RefundRecord, int) {
	for index, item := range r.data.Refunds {
		if item.RefundNo == refundNo {
			return item, index
		}
	}
	return models.RefundRecord{}, -1
}

func (r *Router) recordPaymentCallbackEvent(event models.PaymentCallbackEvent) {
	event.ID = r.nextPaymentCallbackEventID()
	r.data.PaymentCallbackEvents = append([]models.PaymentCallbackEvent{event}, r.data.PaymentCallbackEvents...)
}

func (r *Router) nextPaymentCallbackEventID() int {
	maxID := 0
	for _, item := range r.data.PaymentCallbackEvents {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func extendOneMonth(current string) string {
	if current == "" {
		return nowRFC3339()
	}
	if parsed, err := time.Parse(time.RFC3339, current); err == nil {
		return parsed.AddDate(0, 1, 0).Format(time.RFC3339)
	}
	return current
}

func (r *Router) mockTradeState(order models.Order, payment models.PaymentTransaction) string {
	switch payment.Status {
	case "paid", "refunded":
		return "SUCCESS"
	case "closed":
		return "CLOSED"
	case "failed":
		return "PAYERROR"
	case "created":
		return "NOTPAY"
	case "paying":
		if payment.ChannelOrderNo != "" || strings.HasPrefix(payment.TradeNo, "PAY-") {
			return "USERPAYING"
		}
		return "NOTPAY"
	default:
		if order.Status == "paid" || order.Status == "active" {
			return "SUCCESS"
		}
		return "NOTPAY"
	}
}

func (r *Router) applyPaymentSuccess(order models.Order, orderIndex int, payment models.PaymentTransaction, paymentIndex int, channelOrderNo string, now string) {
	r.data.Orders[orderIndex].Status = "paid"
	r.data.Orders[orderIndex].UpdatedAt = now

	r.data.Payments[paymentIndex].Status = "paid"
	r.data.Payments[paymentIndex].ChannelOrderNo = defaultString(channelOrderNo, payment.ChannelOrderNo)
	r.data.Payments[paymentIndex].PaidAt = now
	r.data.Payments[paymentIndex].UpdatedAt = now

	if subscriptionIndex := r.findSubscriptionIndexByInstance(order.InstanceID); subscriptionIndex >= 0 {
		r.data.Subscriptions[subscriptionIndex].Status = "active"
		r.data.Subscriptions[subscriptionIndex].UpdatedAt = now
		r.data.Subscriptions[subscriptionIndex].CurrentPeriodEnd = extendOneMonth(r.data.Subscriptions[subscriptionIndex].CurrentPeriodEnd)
	}

	if instanceIndex := r.findInstanceIndex(order.InstanceID); instanceIndex >= 0 {
		switch order.Action {
		case "buy", "renew":
			r.data.Instances[instanceIndex].Status = "running"
		case "upgrade":
			r.data.Instances[instanceIndex].Status = "updating"
			r.data.Instances[instanceIndex].Plan = order.PlanCode
		default:
			r.data.Instances[instanceIndex].Status = "running"
		}
		r.data.Instances[instanceIndex].UpdatedAt = now
	}

	if walletIndex := r.findWalletIndex(order.TenantID); walletIndex >= 0 {
		if r.data.Wallets[walletIndex].FrozenAmount >= order.Amount {
			r.data.Wallets[walletIndex].FrozenAmount -= order.Amount
		}
		r.data.Wallets[walletIndex].UpdatedAt = now
	}

	if statementIndex := r.findLatestBillingStatementIndex(order.TenantID); statementIndex >= 0 {
		r.data.BillingStatements[statementIndex].PaidAmount += order.Amount
		r.data.BillingStatements[statementIndex].ClosingBalance += order.Amount
	}
}

func (r *Router) hasPaymentCallbackForOrder(orderNo string) bool {
	for _, item := range r.data.PaymentCallbackEvents {
		if item.OutTradeNo == orderNo && item.ProcessStatus == "succeeded" {
			return true
		}
	}
	return false
}

func (r *Router) findWalletIndex(tenantID int) int {
	for index, item := range r.data.Wallets {
		if item.TenantID == tenantID {
			return index
		}
	}
	return -1
}

func (r *Router) findLatestBillingStatementIndex(tenantID int) int {
	for index := len(r.data.BillingStatements) - 1; index >= 0; index-- {
		if r.data.BillingStatements[index].TenantID == tenantID {
			return index
		}
	}
	return -1
}

func (r *Router) applyRefundSuccess(refund models.RefundRecord, refundIndex int, channelRefundNo string, updatedAt string) {
	r.data.Refunds[refundIndex].Status = "success"
	r.data.Refunds[refundIndex].ChannelRefundNo = channelRefundNo
	r.data.Refunds[refundIndex].UpdatedAt = updatedAt

	if _, paymentIndex := r.findLatestPaymentByOrder(refund.OrderID); paymentIndex >= 0 {
		r.data.Payments[paymentIndex].Status = "refunded"
		r.data.Payments[paymentIndex].UpdatedAt = updatedAt
		if order, orderIndex := r.findOrderIndex(refund.OrderID); orderIndex >= 0 {
			r.data.Orders[orderIndex].Status = "refunded"
			r.data.Orders[orderIndex].UpdatedAt = updatedAt
			if instanceIndex := r.findInstanceIndex(order.InstanceID); instanceIndex >= 0 {
				r.data.Instances[instanceIndex].Status = "stopped"
				r.data.Instances[instanceIndex].UpdatedAt = updatedAt
			}
			if invoiceIndex := r.findInvoiceIndexByOrder(order.ID); invoiceIndex >= 0 {
				r.data.Invoices[invoiceIndex].Status = "reversal_pending"
				r.data.Invoices[invoiceIndex].UpdatedAt = updatedAt
			}
			if walletIndex := r.findWalletIndex(order.TenantID); walletIndex >= 0 {
				r.data.Wallets[walletIndex].AvailableAmount += refund.Amount
				r.data.Wallets[walletIndex].UpdatedAt = updatedAt
			}
			if statementIndex := r.findLatestBillingStatementIndex(order.TenantID); statementIndex >= 0 {
				r.data.BillingStatements[statementIndex].RefundAmount += refund.Amount
				r.data.BillingStatements[statementIndex].ClosingBalance += refund.Amount
			}
		}
	}
}

func ageExceeds(timestamp string, duration time.Duration) bool {
	if parsed, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return time.Since(parsed) > duration
	}
	return false
}

func mapRefundStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESS":
		return "success"
	case "ABNORMAL", "CLOSED", "FAILED":
		return "failed"
	case "", "PROCESSING", "CHANGE":
		return "pending"
	default:
		return strings.ToLower(status)
	}
}

func (r *Router) applyOrderClosed(order models.Order, orderIndex int, updatedAt string) {
	r.data.Orders[orderIndex].Status = "closed"
	r.data.Orders[orderIndex].UpdatedAt = updatedAt

	if payment, paymentIndex := r.findLatestPaymentByOrder(order.ID); paymentIndex >= 0 {
		if payment.Status != "paid" && payment.Status != "refunded" {
			r.data.Payments[paymentIndex].Status = "closed"
			r.data.Payments[paymentIndex].UpdatedAt = updatedAt
		}
	}
}

func (r *Router) applyPaymentFailure(order models.Order, orderIndex int, payment models.PaymentTransaction, paymentIndex int, updatedAt string) {
	r.data.Payments[paymentIndex].Status = "failed"
	r.data.Payments[paymentIndex].UpdatedAt = updatedAt

	if order.Status == "paying" {
		r.data.Orders[orderIndex].Status = "pending"
		r.data.Orders[orderIndex].UpdatedAt = updatedAt
	}
}
