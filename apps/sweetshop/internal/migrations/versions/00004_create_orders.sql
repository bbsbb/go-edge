-- +goose Up
CREATE TABLE IF NOT EXISTS app_sweetshop.orders (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES app_sweetshop.organizations(id),
    system_created_at TIMESTAMPTZ NOT NULL DEFAULT TRANSACTION_TIMESTAMP(),
    system_updated_at TIMESTAMPTZ NOT NULL DEFAULT TRANSACTION_TIMESTAMP(),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed'))
);

ALTER TABLE app_sweetshop.orders ENABLE ROW LEVEL SECURITY;

CREATE POLICY organization_isolation_policy ON app_sweetshop.orders
    USING (organization_id = current_setting('app_sweetshop.current_organization')::UUID);

CREATE TABLE IF NOT EXISTS app_sweetshop.order_items (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES app_sweetshop.organizations(id),
    order_id UUID NOT NULL REFERENCES app_sweetshop.orders(id),
    product_id UUID NOT NULL REFERENCES app_sweetshop.products(id),
    system_created_at TIMESTAMPTZ NOT NULL DEFAULT TRANSACTION_TIMESTAMP(),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price_cents INTEGER NOT NULL CHECK (price_cents > 0)
);

ALTER TABLE app_sweetshop.order_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY organization_isolation_policy ON app_sweetshop.order_items
    USING (organization_id = current_setting('app_sweetshop.current_organization')::UUID);

-- +goose Down
DROP TABLE IF EXISTS app_sweetshop.order_items;
DROP TABLE IF EXISTS app_sweetshop.orders;
