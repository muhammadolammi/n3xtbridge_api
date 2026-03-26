-- +goose Up 
ALTER TABLE invoices
ADD status TEXT NOT NULL DEFAULT 'unpaid';

-- +goose Down
ALTER TABLE invoices
DROP COLUMN status;
