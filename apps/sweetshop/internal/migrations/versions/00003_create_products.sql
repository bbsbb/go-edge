-- +goose Up
CREATE TABLE IF NOT EXISTS app_sweetshop.products (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES app_sweetshop.organizations(id),
    system_created_at TIMESTAMPTZ NOT NULL DEFAULT TRANSACTION_TIMESTAMP(),
    system_updated_at TIMESTAMPTZ NOT NULL DEFAULT TRANSACTION_TIMESTAMP(),
    name TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('ice_cream', 'marshmallow')),
    price_cents INTEGER NOT NULL CHECK (price_cents > 0),
    UNIQUE (organization_id, name)
);

ALTER TABLE app_sweetshop.products ENABLE ROW LEVEL SECURITY;

CREATE POLICY organization_isolation_policy ON app_sweetshop.products
    USING (organization_id = current_setting('app_sweetshop.current_organization')::UUID);

-- +goose Down
DROP TABLE IF EXISTS app_sweetshop.products;
