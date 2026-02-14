-- +goose Up
CREATE TABLE IF NOT EXISTS app_sweetshop.organizations (
    id UUID PRIMARY KEY,
    system_created_at TIMESTAMPTZ NOT NULL DEFAULT TRANSACTION_TIMESTAMP(),
    system_updated_at TIMESTAMPTZ NOT NULL DEFAULT TRANSACTION_TIMESTAMP(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE
);

-- No RLS on organizations: this table is queried to *establish* the tenant
-- context (e.g., resolve slug → org in middleware), before the RLS variable
-- is set. Only tenant-owned data (products, orders, items) sits behind RLS.

-- +goose Down
DROP TABLE IF EXISTS app_sweetshop.organizations;
