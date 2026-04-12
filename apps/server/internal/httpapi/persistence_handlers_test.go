package httpapi

import (
	"context"
	"net/http"
	"testing"

	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
)

func TestHandlePortalPurchasePersistsOrder(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/purchases", map[string]any{
		"planCode": "trial",
		"action":   "buy",
	})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected purchase status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if store.saveDataCalls != 1 {
		t.Fatalf("expected SaveData called once, got %d", store.saveDataCalls)
	}
	if len(router.data.Orders) == 0 || router.data.Orders[0].OrderNo == "" {
		t.Fatalf("expected persisted order with orderNo, got %#v", router.data.Orders)
	}
}

func TestHandleWechatPayCallbackPersistsFailedEventOnInvalidPayload(t *testing.T) {
	store := &fakeCoreStore{}
	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{CoreStore: store},
		runtime: newTestRuntimeAdapter(),
		store:   store,
	}

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/callback/payment/wechatpay", "{")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid callback status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if store.saveDataCalls != 1 {
		t.Fatalf("expected SaveData called once for failed callback, got %d", store.saveDataCalls)
	}
	if len(router.data.PaymentCallbackEvents) == 0 {
		t.Fatal("expected payment callback event recorded")
	}
	if router.data.PaymentCallbackEvents[0].ProcessStatus != "failed" {
		t.Fatalf("expected failed callback event, got %#v", router.data.PaymentCallbackEvents[0])
	}
}

type fakeCoreStore struct {
	saveDataCalls int
	lastSavedData models.Data
}

func (f *fakeCoreStore) Migrate(ctx context.Context) error {
	return nil
}

func (f *fakeCoreStore) Bootstrap(seed models.Data) (models.Data, error) {
	return seed, nil
}

func (f *fakeCoreStore) Load(seed models.Data) (models.Data, error) {
	return seed, nil
}

func (f *fakeCoreStore) SaveInstanceState(state corestore.InstanceState) error {
	return nil
}

func (f *fakeCoreStore) SaveData(data models.Data) error {
	f.saveDataCalls++
	f.lastSavedData = data
	return nil
}

func (f *fakeCoreStore) SaveDiagnosticsMutation(mutation corestore.DiagnosticsMutation) error {
	f.saveDataCalls++
	return nil
}

func (f *fakeCoreStore) SaveWorkspaceMutation(mutation corestore.WorkspaceMutation) error {
	f.saveDataCalls++
	return nil
}

func (f *fakeCoreStore) Ping(ctx context.Context) error {
	return nil
}

func (f *fakeCoreStore) Close() error {
	return nil
}
