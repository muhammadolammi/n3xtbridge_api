-- +goose Up
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS promo_ids TEXT[] NOT NULL DEFAULT '{}'; -- list of promos ids

-- +goose Down
ALTER TABLE quotes DROP COLUMN IF EXISTS promo_ids;
