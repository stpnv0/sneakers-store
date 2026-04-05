-- +goose Up
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_url TEXT;

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS payment_url;
