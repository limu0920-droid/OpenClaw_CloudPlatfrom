package httpapi

import (
	"strings"
	"testing"
)

func TestExternalConfigValidateStrictModeRequiresProductionDependencies(t *testing.T) {
	cfg := ExternalConfig{
		StrictMode:      true,
		RuntimeProvider: "mock",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected strict mode validation error")
	}
	message := err.Error()
	if !strings.Contains(message, "DATABASE_URL is required in strict mode") {
		t.Fatalf("expected database requirement error, got %q", message)
	}
	if !strings.Contains(message, "RUNTIME_PROVIDER must be a real runtime provider in strict mode") {
		t.Fatalf("expected runtime provider error, got %q", message)
	}
}

func TestExternalConfigValidateStrictModeAcceptsPersistentRuntimeConfig(t *testing.T) {
	cfg := ExternalConfig{
		StrictMode:                   true,
		CoreStore:                    &fakeCoreStore{},
		RuntimeProvider:              "kubectl",
		OpenSearchEnabled:            true,
		OpenSearchURL:                "http://opensearch.internal:9200",
		OpenSearchIndex:              "openclaw-logs",
		OpenSearchUsername:           "platform-api",
		OpenSearchPassword:           "real-password",
		ObjectStorageEndpoint:        "https://s3.openclaw.internal",
		ObjectStorageBucket:          "openclaw-prod",
		ObjectStorageAccessKey:       "object-ak",
		ObjectStorageSecretKey:       "object-sk",
		ArtifactPreviewPublicBaseURL: "https://preview.openclaw.internal",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected strict mode config to pass, got %v", err)
	}
}

func TestExternalConfigValidateStrictModeRejectsPlaceholderSecrets(t *testing.T) {
	cfg := ExternalConfig{
		StrictMode:                   true,
		CoreStore:                    &fakeCoreStore{},
		RuntimeProvider:              "kubectl",
		KeycloakEnabled:              true,
		KeycloakBaseURL:              "https://sso.openclaw.internal",
		KeycloakRealm:                "openclaw",
		KeycloakClientID:             "portal",
		KeycloakClientSecret:         "set-in-cluster",
		KeycloakRedirectURL:          "https://portal.openclaw.internal/login",
		KeycloakPostLoginRedirectURL: "https://portal.openclaw.internal/portal",
		KeycloakLogoutRedirectURL:    "https://portal.openclaw.internal/login",
		KeycloakSessionSecret:        "change-me-session",
		KeycloakCookieName:           "openclaw_auth",
		KeycloakFlowCookieName:       "openclaw_auth_flow",
		ObjectStorageEndpoint:        "https://s3.openclaw.internal",
		ObjectStorageBucket:          "openclaw-prod",
		ObjectStorageAccessKey:       "object-ak",
		ObjectStorageSecretKey:       "object-sk",
		ArtifactPreviewPublicBaseURL: "https://preview.openclaw.internal",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected placeholder secret validation error")
	}
	if !strings.Contains(err.Error(), "KEYCLOAK_* configuration is incomplete for strict mode") {
		t.Fatalf("expected keycloak strict validation error, got %q", err.Error())
	}
}
