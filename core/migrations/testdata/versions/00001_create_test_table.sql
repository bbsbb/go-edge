-- +goose Up
CREATE TABLE IF NOT EXISTS test_migrations_table (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS test_migrations_table;
