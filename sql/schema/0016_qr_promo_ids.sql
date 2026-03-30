-- +goose Up
ALTER TABLE quote_requests ADD COLUMN IF NOT EXISTS promo_ids TEXT[] NOT NULL DEFAULT '{}'; -- list of promos ids

-- +goose Down
ALTER TABLE quote_requests DROP COLUMN IF EXISTS promo_ids;
