-- +goose Up 
ALTER TABLE invoices
ADD quote_id UUID  UNIQUE REFERENCES quotes(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE invoices
DROP COLUMN quote_id;
