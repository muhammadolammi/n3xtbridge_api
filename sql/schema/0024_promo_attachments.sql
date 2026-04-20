-- +goose Up
ALTER TABLE promotions ADD COLUMN IF NOT EXISTS attachments TEXT[] NOT NULL DEFAULT '{}'; -- list of promos ids

-- +goose Down
ALTER TABLE promotions DROP COLUMN IF EXISTS attachments;
