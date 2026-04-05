package main

import (
	"log"
	"net/http"
	"os"

	"openclaw/mockapi/internal/httpapi"
	"openclaw/mockapi/internal/mockdata"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	data := mockdata.Seed()
	router := httpapi.NewRouter(data, httpapi.ExternalConfig{
		KeycloakEnabled:           os.Getenv("KEYCLOAK_ENABLED") == "true",
		KeycloakBaseURL:           os.Getenv("KEYCLOAK_BASE_URL"),
		KeycloakRealm:             os.Getenv("KEYCLOAK_REALM"),
		KeycloakClientID:          os.Getenv("KEYCLOAK_CLIENT_ID"),
		KeycloakClientSecret:      os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		KeycloakRedirectURL:       os.Getenv("KEYCLOAK_REDIRECT_URL"),
		KeycloakLogoutRedirectURL: os.Getenv("KEYCLOAK_LOGOUT_REDIRECT_URL"),
		KeycloakMockUser:          os.Getenv("KEYCLOAK_MOCK_USER"),
		KeycloakMockEmail:         os.Getenv("KEYCLOAK_MOCK_EMAIL"),
		KeycloakMockRole:          os.Getenv("KEYCLOAK_MOCK_ROLE"),
		KeycloakSessionSecret:     os.Getenv("KEYCLOAK_SESSION_SECRET"),
		KeycloakCookieName:        os.Getenv("KEYCLOAK_COOKIE_NAME"),
		KeycloakCookieSecure:      os.Getenv("KEYCLOAK_COOKIE_SECURE") == "true",
		OpenSearchEnabled:         os.Getenv("OPENSEARCH_ENABLED") == "true",
		OpenSearchURL:             os.Getenv("OPENSEARCH_URL"),
		OpenSearchIndex:           os.Getenv("OPENSEARCH_INDEX"),
		OpenSearchUsername:        os.Getenv("OPENSEARCH_USERNAME"),
		OpenSearchPassword:        os.Getenv("OPENSEARCH_PASSWORD"),
	})

	addr := ":" + port
	log.Printf("mock api listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
