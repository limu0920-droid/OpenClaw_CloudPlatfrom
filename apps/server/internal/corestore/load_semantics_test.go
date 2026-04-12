package corestore

import "testing"

func TestNewEmptyDataInitializesEmptySlices(t *testing.T) {
	data := newEmptyData()

	if data.Tenants == nil || data.Instances == nil || data.PlanOffers == nil || data.Orders == nil {
		t.Fatalf("expected core slices initialized, got %#v", data)
	}
	if len(data.Tenants) != 0 || len(data.Instances) != 0 || len(data.PlanOffers) != 0 || len(data.Orders) != 0 {
		t.Fatalf("expected empty data set, got %#v", data)
	}
	if data.RuntimeBindings == nil || data.BrandBindings == nil || data.PaymentCallbackEvents == nil {
		t.Fatalf("expected auxiliary slices initialized, got %#v", data)
	}
}
