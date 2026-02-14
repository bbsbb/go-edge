// Package middleware provides HTTP middleware for organization context extraction.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bbsbb/go-edge/core/domain"
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
)

type OrganizationLoader interface {
	LoadOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error)
}

const slugHeader = "X-Organization-Slug"

func extractSlug(r *http.Request) string {
	if slug := r.Header.Get(slugHeader); slug != "" {
		return slug
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return strings.Split(host, ".")[0]
}

type WithOrganizationConfig struct {
	SkipPaths []string
	Logger    *slog.Logger
	Loader    OrganizationLoader
}

// WithOrganization extracts organization from request subdomain and adds it to context.
func WithOrganization(cfg WithOrganizationConfig) func(http.Handler) http.Handler {
	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			slug := extractSlug(r)

			org, err := cfg.Loader.LoadOrganizationBySlug(ctx, slug)
			if err != nil {
				cfg.Logger.ErrorContext(ctx, "could not find organization", "slug", slug, "error", err)
				transporthttp.WriteError(w, r, domain.NewError(domain.CodeNotFound, "organization not found"), cfg.Logger)
				return
			}

			ctx = domain.ContextWithOrganization(ctx, org)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
