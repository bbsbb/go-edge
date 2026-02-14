package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	coredomain "github.com/bbsbb/go-edge/core/domain"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
	"github.com/bbsbb/go-edge/sweetshop/internal/infrastructure/persistence/sqlcgen"
)

// OrganizationRepo provides access to the organizations table.
// Organizations sit outside RLS — they are queried to establish tenant context,
// not the other way around.
type OrganizationRepo struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepo(pool *pgxpool.Pool) *OrganizationRepo {
	return &OrganizationRepo{pool: pool}
}

func (r *OrganizationRepo) conn(ctx context.Context) sqlcgen.DBTX {
	if tx := psqlfx.TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *OrganizationRepo) FindByID(ctx context.Context, id uuid.UUID) (*coredomain.Organization, error) {
	m, err := sqlcgen.New(r.conn(ctx)).FindOrganizationByID(ctx, id)
	if err != nil {
		return nil, psqlfx.TranslateError(err)
	}
	return organizationToDomain(m), nil
}

func (r *OrganizationRepo) FindBySlug(ctx context.Context, slug string) (*coredomain.Organization, error) {
	m, err := sqlcgen.New(r.conn(ctx)).FindOrganizationBySlug(ctx, slug)
	if err != nil {
		return nil, psqlfx.TranslateError(err)
	}
	return organizationToDomain(m), nil
}

func (r *OrganizationRepo) LoadOrganizationBySlug(ctx context.Context, slug string) (*coredomain.Organization, error) {
	return r.FindBySlug(ctx, slug)
}

func (r *OrganizationRepo) Create(ctx context.Context, org *coredomain.Organization) error {
	err := sqlcgen.New(r.conn(ctx)).CreateOrganization(ctx, organizationCreateParams(org, time.Now()))
	return psqlfx.TranslateError(err)
}
