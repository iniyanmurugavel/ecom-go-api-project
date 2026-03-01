-- Migration 4: Add foreign key from order_items.product_id to products.id.
-- +goose Up
-- +goose StatementBegin
ALTER TABLE order_items ADD CONSTRAINT fk_product FOREIGN KEY (product_id) REFERENCES products(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_product;
-- +goose StatementEnd
