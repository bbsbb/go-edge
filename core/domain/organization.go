// Package domain contains shared domain types.
package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrMissingOrganization = errors.New("domain: missing organization in context")

type Organization struct {
	ID   uuid.UUID
	Slug string
}

type organizationContextKey struct{}

func ContextWithOrganization(ctx context.Context, org *Organization) context.Context {
	return context.WithValue(ctx, organizationContextKey{}, org)
}

func OrganizationFromContext(ctx context.Context) (*Organization, error) {
	org, ok := ctx.Value(organizationContextKey{}).(*Organization)
	if !ok || org == nil {
		return nil, ErrMissingOrganization
	}
	return org, nil
}
