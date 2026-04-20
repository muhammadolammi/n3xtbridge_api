-- +goose Up
ALTER TABLE quote_requests ADD COLUMN IF NOT EXISTS video_key TEXT NOT NULL DEFAULT ''; 

-- +goose Down
ALTER TABLE quote_requests DROP COLUMN IF EXISTS video_key;
