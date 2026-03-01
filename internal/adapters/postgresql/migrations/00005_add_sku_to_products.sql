-- Migration 5: Add sku (Stock Keeping Unit) to products.
-- Example migration: adds a column WITHOUT losing existing data.
-- Existing rows get sku = NULL. See docs/MIGRATION_EXAMPLE.md for full workflow.
--
-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN sku TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN sku;
-- +goose StatementEnd
