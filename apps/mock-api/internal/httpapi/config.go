package httpapi

type ExternalConfig struct {
	KeycloakEnabled           bool
	KeycloakBaseURL           string
	KeycloakRealm             string
	KeycloakClientID          string
	KeycloakClientSecret      string
	KeycloakRedirectURL       string
	KeycloakLogoutRedirectURL string
	KeycloakMockUser          string
	KeycloakMockEmail         string
	KeycloakMockRole          string
	KeycloakSessionSecret     string
	KeycloakCookieName        string
	KeycloakCookieSecure      bool

	OpenSearchEnabled  bool
	OpenSearchURL      string
	OpenSearchIndex    string
	OpenSearchUsername string
	OpenSearchPassword string
}
