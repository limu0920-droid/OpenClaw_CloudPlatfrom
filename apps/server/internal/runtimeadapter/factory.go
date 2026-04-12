package runtimeadapter

import (
	"fmt"
	"strings"
)

func NewAdapter(cfg Config) (Adapter, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = "kubectl"
	}

	if provider == "kubectl" {
		return NewKubectlAdapter(cfg), nil
	}

	return nil, fmt.Errorf("unsupported runtime provider %q", provider)
}
