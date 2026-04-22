-- +goose Up
ALTER TABLE quote_requests DROP COLUMN service_name;

-- +goose Down
ALTER TABLE quote_requests ADD COLUMN service_name TEXT NOT NULL DEFAULT '';