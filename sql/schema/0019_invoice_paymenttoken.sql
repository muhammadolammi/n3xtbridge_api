-- +goose Up 
ALTER TABLE invoices
ADD payment_token TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE invoices
DROP COLUMN payment_token;