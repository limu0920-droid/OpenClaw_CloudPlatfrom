package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (r *Router) strictModeEnabled() bool {
	return r != nil && r.config.StrictMode
}

func (r *Router) strictModeError(message string) error {
	return errors.New(message)
}

func (r *Router) validateStartupDependencies() error {
	if !r.strictModeEnabled() {
		return nil
	}

	problems := make([]string, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if r.store == nil {
		problems = append(problems, "database store is unavailable in strict mode")
	} else if err := r.store.Ping(ctx); err != nil {
		problems = append(problems, fmt.Sprintf("database ping failed: %v", err))
	}

	if r.artifactStore == nil || !r.artifactStore.Enabled() {
		problems = append(problems, "object storage is unavailable in strict mode")
	} else if err := r.artifactStore.Ping(ctx); err != nil {
		problems = append(problems, fmt.Sprintf("object storage ping failed: %v", err))
	}

	if len(r.runtime.ListClusters()) == 0 {
		problems = append(problems, "runtime provider returned no reachable clusters")
	}

	if r.config.OpenSearchEnabled {
		if err := pingHTTP(ctx, strings.TrimRight(r.config.OpenSearchURL, "/")+"/_cluster/health"); err != nil {
			problems = append(problems, fmt.Sprintf("opensearch health check failed: %v", err))
		}
	}
	if r.traceSearchEnabled() {
		if err := pingHTTP(ctx, strings.TrimRight(r.config.TraceSearchURL, "/")+"/_cluster/health"); err != nil {
			problems = append(problems, fmt.Sprintf("trace search health check failed: %v", err))
		}
	}

	if r.config.KeycloakEnabled {
		discoveryURL := strings.TrimRight(r.config.KeycloakBaseURL, "/") + "/realms/" + url.PathEscape(r.config.KeycloakRealm) + "/.well-known/openid-configuration"
		if err := pingHTTP(ctx, discoveryURL); err != nil {
			problems = append(problems, fmt.Sprintf("keycloak discovery failed: %v", err))
		}
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}
