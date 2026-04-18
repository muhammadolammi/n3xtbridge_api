-- +goose Up
ALTER TABLE quote_requests ADD COLUMN IF NOT EXISTS vn_r2_key TEXT NOT NULL DEFAULT ''; 

-- +goose Down
ALTER TABLE quote_requests DROP COLUMN IF EXISTS vn_r2_key;
