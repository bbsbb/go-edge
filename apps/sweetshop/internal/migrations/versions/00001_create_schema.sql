-- +goose Up
CREATE SCHEMA IF NOT EXISTS app_sweetshop;

-- +goose Down
DROP SCHEMA IF EXISTS app_sweetshop CASCADE;
