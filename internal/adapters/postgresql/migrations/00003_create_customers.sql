-- Migration 3: customers table for auth (email unique, password_hash for bcrypt).
-- orders.customer_id references customers.id logically; no FK to allow existing data.
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customers (
  id BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customers;
-- +goose StatementEnd
