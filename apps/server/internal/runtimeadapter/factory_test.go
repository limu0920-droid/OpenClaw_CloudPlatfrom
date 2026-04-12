package runtimeadapter

import "testing"

func TestNewAdapterRejectsUnknownProvider(t *testing.T) {
	adapter, err := NewAdapter(Config{Provider: "unknown"})
	if err == nil {
		t.Fatal("expected unknown provider error")
	}
	if adapter != nil {
		t.Fatalf("expected nil adapter on error, got %#v", adapter)
	}
}

func TestNewAdapterRejectsMockProvider(t *testing.T) {
	adapter, err := NewAdapter(Config{Provider: "mock"})
	if err == nil {
		t.Fatal("expected mock provider error")
	}
	if adapter != nil {
		t.Fatalf("expected nil adapter on error, got %#v", adapter)
	}
}

func TestNewAdapterUsesKubectlByDefault(t *testing.T) {
	adapter, err := NewAdapter(Config{})
	if err != nil {
		t.Fatalf("expected default kubectl adapter, got %v", err)
	}
	if adapter == nil {
		t.Fatal("expected adapter instance")
	}
}
