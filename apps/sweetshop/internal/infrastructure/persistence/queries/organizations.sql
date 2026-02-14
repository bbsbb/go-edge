-- name: FindOrganizationByID :one
SELECT * FROM app_sweetshop.organizations WHERE id = $1;

-- name: FindOrganizationBySlug :one
SELECT * FROM app_sweetshop.organizations WHERE slug = $1;

-- name: CreateOrganization :exec
INSERT INTO app_sweetshop.organizations (id, system_created_at, system_updated_at, name, slug)
VALUES ($1, $2, $3, $4, $5);
