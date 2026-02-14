-- +goose Up
ALTER TABLE test_migrations_table ADD COLUMN description TEXT;

-- +goose Down
ALTER TABLE test_migrations_table DROP COLUMN description;
