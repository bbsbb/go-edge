-- name: FindOrderByID :one
SELECT * FROM app_sweetshop.orders WHERE id = $1;

-- name: CreateOrder :exec
INSERT INTO app_sweetshop.orders (id, organization_id, system_created_at, system_updated_at, status)
VALUES ($1, $2, $3, $4, $5);

-- name: CloseOrder :one
UPDATE app_sweetshop.orders
SET system_updated_at = $2, status = 'closed'
WHERE id = $1
RETURNING *;

-- name: CreateOrderItem :exec
INSERT INTO app_sweetshop.order_items (id, organization_id, order_id, product_id, system_created_at, quantity, price_cents)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListOrderItemsByOrderID :many
SELECT * FROM app_sweetshop.order_items WHERE order_id = $1 ORDER BY system_created_at;
