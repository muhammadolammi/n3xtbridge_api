-- +goose Up
ALTER TYPE quote_status ADD VALUE IF NOT EXISTS 'paid';

-- +goose Down
