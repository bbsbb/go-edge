-- name: FindProductByID :one
SELECT * FROM app_sweetshop.products WHERE id = $1;

-- name: ListProducts :many
SELECT * FROM app_sweetshop.products ORDER BY name;

-- name: CreateProduct :exec
INSERT INTO app_sweetshop.products (id, organization_id, system_created_at, system_updated_at, name, category, price_cents)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateProduct :execrows
UPDATE app_sweetshop.products
SET system_updated_at = $2, name = $3, category = $4, price_cents = $5
WHERE id = $1;

-- name: DeleteProduct :execrows
DELETE FROM app_sweetshop.products WHERE id = $1;
